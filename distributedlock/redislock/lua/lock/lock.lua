local key = KEYS[1]
local curToken = ARGV[1]
local duration = ARGV[2]
local oldToken = redis.call('get', key)
-- 只有锁未被任何人持有才能加锁成功 被自己持有也会加锁失败
if oldToken == false then
    -- OK
    return redis.call('setex', key, duration, curToken)
else
    return 0
end