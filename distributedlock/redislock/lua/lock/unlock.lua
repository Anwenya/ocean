local key = KEYS[1]
local curToken = ARGV[1]
local oldToken = redis.call('get', key)

-- 解锁时必须拥有锁
if curToken == oldToken then
    -- del成功返回 (integer) 1
    return redis.call('del', key)
else
    return 0
end