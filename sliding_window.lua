local key    = KEYS[1]
local limit  = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now    = tonumber(ARGV[3])
local cost   = tonumber(ARGV[4])
local nonce  = ARGV[5]

-- TODO: implement sliding window
return {0, 0, "0", "0"}
