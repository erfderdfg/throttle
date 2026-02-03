local key    = KEYS[1]
local limit  = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now    = tonumber(ARGV[3])
local cost   = tonumber(ARGV[4])
local nonce  = ARGV[5]

local window_start = now - window
redis.call('ZREMRANGEBYSCORE', key, '-inf', window_start)

local count = redis.call('ZCARD', key)

local allowed     = 0
local remaining   = 0
local retry_after = 0.0
local reset_time  = 0.0

if count + cost <= limit then
	local member = tostring(now) .. ':' .. nonce
	redis.call('ZADD', key, now, member)
	allowed   = 1
	remaining = limit - count - cost
	reset_time = now
else
	local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
	if #oldest >= 2 then
		local oldest_time = tonumber(oldest[2])
		retry_after = (oldest_time + window) - now
		if retry_after < 0 then retry_after = 0 end
	end
	reset_time = now + retry_after
end

return {allowed, remaining, tostring(retry_after), tostring(reset_time)}
