调度一般指的是由g0按照特定的策略找到下一个可执行的g的过程，
本笔记记录的调度指的是调度器p实现的从一个g切换到另一个g的过程。

## 主动调度

通过主动调用runtime.Gosched()方法来让出执行权。

```go

// 1
// runtime/proc.go:336
func Gosched() {
    // g栈切换到g0栈并调用fn(g) 其中g是进行调用的程序
    // 也就是会将g切换到g0 之后调用gosched_m(g)
    mcall(gosched_m)
}

// 2
// runtime/proc.go:3764
func gosched_m(gp *g) {
    goschedImpl(gp) 
}

// 3
// runtime/proc.go:3748
func goschedImpl(gp *g) {
    status := readgstatus(gp)
    ...
    // 切换当前g的状态为_Grunnable
    casgstatus(gp, _Grunning, _Grunnable)
    
    // 将当前g和m进行解绑 本身就是软绑定
    dropg()
    
    // 将当前g添加到全局队列 需要锁
    lock(&sched.lock)
    globrunqput(gp)
    unlock(&sched.lock)
    
    // 新一轮调度 找下一个可以执行的g
    schedule()
}

```

## 被动调度

因当前不满足某种执行条件，g可能会陷入阻塞态无法被调度，
如channel、mutex、netpoll等相关内容会涉及到大量的阻塞/唤醒操作，
这两个操作的具体实现对应到runtime的gopark、goready两个函数，
且需要由上层应用自己维护待唤醒的g。

阻塞

```go

// 1
// runtime/proc.go:381
func gopark(unlockf func(*g, unsafe.Pointer) bool, lock unsafe.Pointer, reason waitReason, traceReason traceBlockReason, traceskip int) {
    // 当前g绑定的m的locks++  之后返回g的m
    mp := acquirem()
    // 拿到m正在执行的g 也就是当前g
    gp := mp.curg
    // g状态校验
    status := readgstatus(gp)
    if status != _Grunning && status != _Gscanrunning {
        throw("gopark: bad g status")
    }
    // 用于将参数从gopark携带到park_m
    mp.waitlock = lock
    // 如果这个函数返回false 该g会恢复执行
    mp.waitunlockf = unlockf
    // 记录等待的原因
    gp.waitreason = reason
    mp.waitTraceBlockReason = traceReason
    mp.waitTraceSkip = traceskip
    // 当前g绑定的m的locks--
    releasem(mp)
    
    // 切换
    mcall(park_m)
}


// 2
// runtime/proc.go:3721
func park_m(gp *g) {
    // 拿到当前的m
    mp := getg().m
    
    // 状态切换
    casgstatus(gp, _Grunning, _Gwaiting)
    // 解绑
    dropg()
    
    // 如果这个函数返回false 该g会恢复执行
    if fn := mp.waitunlockf; fn != nil {
        ok := fn(gp, mp.waitlock)
        mp.waitunlockf = nil
        mp.waitlock = nil
        if !ok {
            casgstatus(gp, _Gwaiting, _Grunnable)
            execute(gp, true) // Schedule it back, never returns.
        }
    }
    // 新一轮调度 找下一个可以执行的g
    schedule()
}
```

唤醒，将 g 从阻塞态中恢复，重新进入等待执行的状态。
```go

// 1
// runtime/proc.go:407
func goready(gp *g, traceskip int) {
    systemstack(func() {
        ready(gp, traceskip, true)
    })
}

// 2
// runtime/proc.go:890
func ready(gp *g, traceskip int, next bool) {
    status := readgstatus(gp)
    // ...
    // 将 g 的状态从阻塞态改为可执行的状态
    casgstatus(gp, _Gwaiting, _Grunnable)
    runqput(mp.p.ptr(), gp, next)
    // ...
}

// 3
// runtime/proc.go:6199
func runqput(pp *p, gp *g, next bool) {
    if randomizeScheduler && next && fastrandn(2) == 0 {
        next = false
    }
    
    // 直接将当前gp替换为pp.runnext
    // 再判断oldnext存不存在,
    // 不存在返回即可
    // 如果存在则让oldnext作为gp执行后续的操作
    if next {
    retryNext:
        oldnext := pp.runnext
        if !pp.runnext.cas(oldnext, guintptr(unsafe.Pointer(gp))) {
            goto retryNext
        }
        if oldnext == 0 {
            return
        }
        gp = oldnext.ptr()
    }
    
    
    retry:
    // 如果本地队列未满 将gp放到最后 接着返回
    h := atomic.LoadAcq(&pp.runqhead)
    t := pp.runqtail
    if t-h < uint32(len(pp.runq)) {
        pp.runq[t%uint32(len(pp.runq))].set(gp)
        atomic.StoreRel(&pp.runqtail, t+1)
        return
    }
    // 如果队列满了则入全局队列并且会连带g一起将一半的元素转移到全局队列
    if runqputslow(pp, gp, h, t) {
        return
    }
    
    goto retry
}

```

## 抢占调度

如果 g 执行系统调用超过指定的时长，且全局的 p 资源比较紧缺，就会将 p 抢占出来用于其他 g 的调度。
等 g 完成系统调用后，会重新进入可执行队列中等待被调度。

因为发起系统调用时需要打破用户态的边界进入内核态，此时 **m 也会因系统调用而陷入僵直，无法主动完成抢占调度的行为。**

因此，在 Go 进程会有一个全局监控协程 monitor g 的存在，这个 g 会越过 p 直接与一个 m 进行绑定，
不断轮询对所有 p 的执行状况进行监控。 倘若发现满足抢占调度的条件，则会从第三方的角度手干预，主动发起该动作。

```go

// 1
// runtime/proc.go:144
func main() {
    mp := getg().m
    
    if GOARCH != "wasm" {
        systemstack(func() {
            newm(sysmon, nil, -1)
        })
    }
}


// 2
// runtime/proc.go:5515
func sysmon() {
    // ...
    // retake P's blocked in syscalls
    // and preempt long running G's
    if retake(now) != 0 {
        idle = 0
    } else {
        idle++
    }
    // ...  
}



// 3
// runtime/proc.go:5672
func retake(now int64) uint32 {
    n := 0
    lock(&allpLock)
    // 遍历全局队列 检查所有p
    for i := 0; i < len(allp); i++ {
        pp := allp[i]
        if pp == nil {
            continue
        }
        pd := &pp.sysmontick
        s := pp.status
        // ...
        if s == _Psyscall {
            // ...
            // 当前 p 没有要执行的 g 而且存在 自旋/空闲的 p 并且系统调用未超时  则跳过
            // 简单理解为如果满足下列条件就需要抢占
            // 1、执行系统调用超过 10 ms
            // 2、p 本地队列有等待执行的 g 或者 当前没有空闲的 p 和 m
            if runqempty(pp) && sched.nmspinning.Load()+sched.npidle.Load() > 0 && pd.syscallwhen+10*1000*1000 > now {
                continue
            }
            unlock(&allpLock)
                
            // 抢占调度
            // 将当前 p 的状态更新为 idle
            if atomic.Cas(&pp.status, s, _Pidle) {
                n++
                pp.syscalltick++
                handoffp(pp)
            }
            
            incidlelocked(1)
            lock(&allpLock)
        }
    }
    unlock(&allpLock)
    return uint32(n)
}

// 4
// runtime/proc.go:2659
func handoffp(pp *p) {
        
    // 本地队列还有 g
    if !runqempty(pp) || sched.runqsize != 0 {
        startm(pp, false, false)
        return
    }
    
    // 全局没有空闲的 p 和 m
    if sched.nmspinning.Load()+sched.npidle.Load() == 0 && sched.nmspinning.CompareAndSwap(0, 1) {
        sched.needspinning.Store(0)
        startm(pp, true, false)
        return
    }
    
    // 全局队列有 g
    if sched.runqsize != 0 {
        unlock(&sched.lock)
        startm(pp, false, false)
        return
    }
    
    // 当前仅有一个p在运行且有网络调用
    if sched.npidle.Load() == gomaxprocs-1 && sched.lastpoll.Load() != 0 {
        unlock(&sched.lock)
        startm(pp, false, false)
        return
    }
}


// 5 先尝试获取已有的空闲的 m，若不存在，则会创建一个新的 m
// runtime/proc.go:2563
func startm(pp *p, spinning, lockheld bool) {
    // ...
    nmp := mget()
    if nmp == nil {
        // ...
        id := mReserveID()
        unlock(&sched.lock)
        
        var fn func()
        if spinning {
            fn = mspinning
        }
        newm(fn, pp, id)
    }
    // ...
}
```

## 正常执行完成

当 g 执行完成时，会先执行 mcall 方法切换至 g0，然后调用 goexit0 方法。

```go

// 1
// runtime/proc.go:3850
func goexit1() {
    mcall(goexit0)  
}

// 2
// runtime/proc.go:3861
func goexit0(gp *g) {
    mp := getg().m
    pp := mp.p.ptr()
    // 标记g死亡
    casgstatus(gp, _Grunning, _Gdead)
    // ...
    // 解绑g和m
    dropg()
    // ...
    // 新一轮调度
    schedule()
}

```
