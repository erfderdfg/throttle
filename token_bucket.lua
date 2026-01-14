local key      = KEYS[1]
local rate     = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now      = tonumber(ARGV[3])
local cost     = tonumber(ARGV[4])

local state      = redis.call('HMGET', key, 'tokens', 'last_refill')
local tokens     = tonumber(state[1])
local last_refill = tonumber(state[2])

if tokens == nil then
    tokens     = capacity
    last_refill = now
end

local elapsed      = now - last_refill
if elapsed < 0 then elapsed = 0 end
local tokens_to_add = elapsed * rate
tokens = tokens + tokens_to_add

-- TODO: burst cap + allow/deny
return {0, 0, "0", "0"}
