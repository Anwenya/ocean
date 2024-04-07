local key = KEYS[1]
local curToken = ARGV[1]
local duration = ARGV[2]
-- 不存在返回false
local oldToken = redis.call('get', key)

-- 续约时必须持有锁
if oldToken == curToken then
    -- 执行成功返回 (integer) 1
    return redis.call('expire', key, duration)
else
    return 0
end