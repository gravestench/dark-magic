-- Supplied-seed draw policy from NRF-R01/R02, verified statically in 1.13c.
-- The caller owns the pair; this module never initializes or retains a seed.
local integer = require("engine.integer/v1")

local M = {}

local function checked_integer(value, minimum, maximum, domain)
    if
        type(value) ~= "number"
        or value ~= value
        or value < minimum
        or value > maximum
        or math.floor(value) ~= value
    then
        error("expected a finite exact " .. domain .. " integer", 3)
    end
    return value
end

local function checked_seed(seed)
    if type(seed) ~= "table" then
        error("expected a supplied seed pair", 3)
    end
    checked_integer(seed.lo, 0, 4294967295, "u32 seed.lo")
    checked_integer(seed.hi, 0, 4294967295, "u32 seed.hi")
end

local function advance(seed)
    -- Keep the wide product in the native capability as two exact words.
    seed.lo, seed.hi = integer.mul_add_u32(seed.lo, 0x6AC690C5, seed.hi)
end

function M.step(seed)
    checked_seed(seed)
    advance(seed)
end

function M.rand(seed, n)
    checked_seed(seed)
    checked_integer(n, -2147483648, 2147483647, "i32 modulus")
    if n <= 0 then
        return 0
    end
    advance(seed)
    if integer.band_u32(n, n - 1) == 0 then
        return integer.band_u32(seed.lo, n - 1)
    end
    return seed.lo % n
end

function M.rand_mask(seed, m)
    checked_seed(seed)
    checked_integer(m, 0, 4294967295, "u32 mask")
    advance(seed)
    -- m - 1 is still exact in Lua, including the wrapping mask at m = 0.
    return integer.band_u32(seed.lo, (m - 1) % 4294967296)
end

-- R10b extracts unsigned integer parts from supplied 16.16 path words.
function M.path_integer_position(path)
    if type(path) ~= "table" then
        error("expected a supplied path", 2)
    end
    checked_integer(path.x_bits, 0, 4294967295, "u32 path.x_bits")
    checked_integer(path.y_bits, 0, 4294967295, "u32 path.y_bits")
    return integer.high_u16(path.x_bits), integer.high_u16(path.y_bits)
end

return M
