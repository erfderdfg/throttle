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
	-- TODO: record and return
	allowed   = 1
	remaining = limit - count - cost
	reset_time = now
else
	allowed = 0
end

return {allowed, remaining, tostring(retry_after), tostring(reset_time)}
