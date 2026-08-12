-- Blood Moor edge warps and primary semantic route.
--
-- The route is expressed in world tiles, independently of DT1 artwork. That
-- lets collision, population, and connectivity tests run without decoding a
-- single map asset.

local draw = require("engine.deterministic/v1").integer
local M = {}

local CELL = 8
local opposites = {north = "south", east = "west", south = "north", west = "east"}

local function point_key(x, y) return x .. ":" .. y end

local function edge_warp(width, height, town_direction)
    local direction = assert(opposites[town_direction], "d2legacy mapgen: town exit must be cardinal")
    local x, y = math.floor(width / 2), math.floor(height / 2)
    if direction == "north" then y = 0
    elseif direction == "east" then x = width - 1
    elseif direction == "south" then y = height - 1
    else x = 0 end
    return {id = 1, role = "town-entry", direction = direction, x = x, y = y, destination_level = 1}
end

local function exit_warp(width, height, entry_direction)
    local direction = opposites[entry_direction]
    local warp = edge_warp(width, height, opposites[direction])
    warp.id, warp.role, warp.direction, warp.destination_level = 2, "next-level-exit", direction, 3
    return warp
end

local function coarse_route(seed, columns, rows, entry_direction)
    local horizontal = entry_direction == "west" or entry_direction == "east"
    local length = horizontal and columns or rows
    local cross_size = horizontal and rows or columns
    local forward = entry_direction == "west" or entry_direction == "north"
    local cross, target = math.floor(cross_size / 2), math.floor(cross_size / 2)
    local result, ordered = {}, {}
    for step = 0, length - 1 do
        local axis = forward and step or length - step - 1
        local x, y = cross, axis
        if horizontal then x, y = axis, cross end
        result[point_key(x, y)] = true
        ordered[#ordered + 1] = {x = x, y = y}
        local remaining = length - step - 1
        if remaining > 0 then
            local delta = draw(seed, "d2legacy.mapgen.outdoor.route", 3, step) - 1
            local next_cross = math.max(0, math.min(cross_size - 1, cross + delta))
            if next_cross - target > remaining - 1 then next_cross = target + remaining - 1 end
            if target - next_cross > remaining - 1 then next_cross = target - remaining + 1 end
            cross = next_cross
        end
    end
    return result, ordered
end

local function raster(from, to, output)
    local x, y = from.x, from.y
    local dx, dy = math.abs(to.x - x), -math.abs(to.y - y)
    local sx, sy = x < to.x and 1 or -1, y < to.y and 1 or -1
    local error_value = dx + dy
    while true do
        output[point_key(x, y)] = {x = x, y = y}
        if x == to.x and y == to.y then return end
        local twice = 2 * error_value
        if twice >= dy then error_value, x = error_value + dy, x + sx end
        if twice <= dx then error_value, y = error_value + dx, y + sy end
    end
end

local function tile_path(ordered, entry, exit)
    local points = {{x = entry.x, y = entry.y}}
    for _, cell in ipairs(ordered) do
        points[#points + 1] = {x = cell.x * CELL + CELL / 2, y = cell.y * CELL + CELL / 2}
    end
    points[#points + 1] = {x = exit.x, y = exit.y}
    local set = {}
    for index = 2, #points do raster(points[index - 1], points[index], set) end
    local result = {}
    for _, tile in pairs(set) do result[#result + 1] = tile end
    return result
end

function M.plan(seed, width, height, town_direction)
    assert(width % CELL == 0 and height % CELL == 0, "d2legacy mapgen: outdoor size must use an 8-tile grid")
    local entry = edge_warp(width, height, town_direction)
    local exit = exit_warp(width, height, entry.direction)
    local cells, ordered = coarse_route(seed, width / CELL, height / CELL, entry.direction)
    return {
        entry = entry, exit = exit, cells = cells, ordered = ordered,
        path = tile_path(ordered, entry, exit), cell_size = CELL,
    }
end

M.opposite = function(direction) return opposites[direction] end

return M
