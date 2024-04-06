package os

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// GetCurrentProcessAndGoroutineIDStr 进程+协程 id
func GetCurrentProcessAndGoroutineIDStr() string {
	pid := GetCurrentProcessID()
	goroutineID := GetCurrentGoroutineID()
	return fmt.Sprintf("%d_%s", pid, goroutineID)
}

// GetCurrentGoroutineID 获得协程ID
func GetCurrentGoroutineID() string {
	buf := make([]byte, 128)
	runtime.Stack(buf, false)
	stackInfo := string(buf)
	return strings.TrimSpace(strings.Split(strings.Split(stackInfo, "[running]")[0], "goroutine")[1])
}

// GetCurrentProcessID 获得进程ID
func GetCurrentProcessID() int {
	return os.Getpid()
}
