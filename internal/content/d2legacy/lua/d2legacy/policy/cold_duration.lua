-- Expansion 1.14d difficulty scaling for monster cold/freeze lengths.

local M = {}

local divisors = {
    [0] = 1,
    [1] = 2,
    [2] = 4,
}

function M.monster_frames(frames, difficulty)
    assert(type(frames) == "number" and frames > 0 and frames == math.floor(frames), "cold frames must be positive")
    local divisor = assert(divisors[difficulty], "unsupported difficulty")
    return math.max(math.floor(frames / divisor), 1)
end

return M
