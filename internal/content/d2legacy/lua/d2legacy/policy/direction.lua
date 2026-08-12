-- Quantize a continuous world-space vector into d2legacy's readable player
-- direction space. The first eight IDs are the historical cardinal/diagonal
-- values; IDs 8-15 fill the half-angle sectors used by 16-direction COFs.

local M = {}

local player16 = {
    {id=0,  x=0,                   y=1},                  -- south
    {id=8,  x=0.3826834323650898,  y=0.9238795325112867}, -- south-southeast
    {id=4,  x=0.7071067811865476,  y=0.7071067811865476}, -- southeast
    {id=15, x=0.9238795325112867,  y=0.3826834323650898}, -- east-southeast
    {id=3,  x=1,                   y=0},                  -- east
    {id=14, x=0.9238795325112867,  y=-0.3826834323650898},
    {id=7,  x=0.7071067811865476,  y=-0.7071067811865476},
    {id=13, x=0.3826834323650898,  y=-0.9238795325112867},
    {id=2,  x=0,                   y=-1},                 -- north
    {id=12, x=-0.3826834323650898, y=-0.9238795325112867},
    {id=6,  x=-0.7071067811865476, y=-0.7071067811865476},
    {id=11, x=-0.9238795325112867, y=-0.3826834323650898},
    {id=1,  x=-1,                  y=0},                  -- west
    {id=10, x=-0.9238795325112867, y=0.3826834323650898},
    {id=5,  x=-0.7071067811865476, y=0.7071067811865476},
    {id=9,  x=-0.3826834323650898, y=0.9238795325112867},
}

local actor8 = {
    player16[1], player16[3], player16[5], player16[7],
    player16[9], player16[11], player16[13], player16[15],
}

function M.quantize(x, y, count)
    if x == 0 and y == 0 then return 0 end
    local directions = assert(({[8]=actor8,[16]=player16})[count],
        "unsupported facing direction count")
    local best_id, best_dot = 0, -math.huge
    for _, direction in ipairs(directions) do
        local dot = x * direction.x + y * direction.y
        if dot > best_dot then
            best_id, best_dot = direction.id, dot
        end
    end
    return best_id
end

function M.player(x, y) return M.quantize(x, y, 16) end

return M
