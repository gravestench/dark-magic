-- Authored Act I dirt-path DT1 sequence selection.
--
-- Bits describe neighboring route cells in this order: NE, E, SE, N, S, NW,
-- W, SW. The table is the verified legacy 1.10f outdoor rule. Keeping both the
-- vocabulary and lookup here means generic world materialization sees only the
-- opaque floor identity selected by d2legacy.

local M = {}

local sequence = {
    0,0,16,16,0,0,16,16,14,14,6,19,14,14,6,19, 15,15,5,5,15,15,21,21,8,8,10,38,8,8,40,20,
    0,0,16,16,0,0,16,16,14,14,6,19,14,14,6,19, 15,15,5,5,15,15,21,21,8,8,10,38,8,8,40,20,
    13,13,7,7,13,13,13,7,4,4,11,37,4,4,11,43, 3,3,12,12,3,3,39,39,9,9,2,43,9,9,44,26,
    13,13,7,7,13,13,13,7,23,23,41,17,23,23,41,17, 3,3,12,12,3,3,39,39,42,42,46,42,42,42,33,31,
    0,0,16,16,0,0,16,16,14,14,6,19,14,14,6,19, 15,15,5,5,15,15,21,21,8,8,10,38,8,8,35,20,
    0,0,16,16,0,0,16,16,14,14,6,19,14,14,6,19, 15,15,5,5,15,15,21,21,8,8,10,38,8,8,40,20,
    13,13,7,7,13,13,13,7,4,4,11,37,4,4,11,37, 18,18,35,35,18,18,22,22,36,36,45,34,36,36,28,29,
    13,13,7,7,13,13,13,7,23,23,41,17,23,23,41,17, 18,18,35,35,18,18,22,22,24,24,25,32,24,24,30,1,
}

local neighbors = {
    {1,-1,7},{1,0,6},{1,1,5},{0,-1,4},
    {0,1,3},{-1,-1,2},{-1,0,1},{-1,1,0},
}

local function key(x, y) return x .. ":" .. y end

function M.apply(path)
    local occupied = {}
    for _, tile in ipairs(path) do occupied[key(tile.x, tile.y)] = true end
    for _, tile in ipairs(path) do
        local mask = 0
        for _, neighbor in ipairs(neighbors) do
            if occupied[key(tile.x + neighbor[1], tile.y + neighbor[2])] then
                mask = mask + 2 ^ neighbor[3]
            end
        end
        local selected = sequence[mask + 1]
        if selected and selected ~= 0 then
            tile.main_index, tile.sub_index = 0, selected
        end
    end
    return path
end

return M
