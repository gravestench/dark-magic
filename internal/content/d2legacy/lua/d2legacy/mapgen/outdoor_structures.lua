-- Blood Moor river, bridge, and cliff policy.
--
-- Structural cells carry authoritative passability. Authored overlay stamps are
-- emitted separately so a headless server can reason about crossings without
-- decoding their DS1/DT1 presentation.

local assets = require("d2legacy.mapgen.assets")
local data = require("d2legacy.mapgen.records")
local draw = require("engine.deterministic/v1").integer
local M = {}
local CELL = 8

local function key(x, y) return x .. ":" .. y end

local function raster(from, to, output)
    local x, y = from.x, from.y
    local dx, dy = math.abs(to.x - x), -math.abs(to.y - y)
    local sx, sy = x < to.x and 1 or -1, y < to.y and 1 or -1
    local error_value = dx + dy
    while true do
        output[key(x, y)] = {x = x, y = y}
        if x == to.x and y == to.y then return end
        local twice = 2 * error_value
        if twice >= dy then error_value, x = error_value + dy, x + sx end
        if twice <= dx then error_value, y = error_value + dx, y + sy end
    end
end

local function path_through_bridge(path, river_x, bridge_y)
    local river_width = 2 * CELL
    local left = {x = river_x - 1, y = bridge_y}
    local right = {x = river_x + river_width, y = bridge_y}
    local left_source, right_source = left, right
    local left_distance, right_distance
    local result = {}
    for _, tile in ipairs(path) do
        if tile.x < river_x or tile.x >= river_x + river_width then result[key(tile.x, tile.y)] = tile end
        if tile.x == left.x and (left_distance == nil or math.abs(tile.y - bridge_y) < left_distance) then
            left_source, left_distance = tile, math.abs(tile.y - bridge_y)
        end
        if tile.x == right.x and (right_distance == nil or math.abs(tile.y - bridge_y) < right_distance) then
            right_source, right_distance = tile, math.abs(tile.y - bridge_y)
        end
    end
    raster(left_source, left, result)
    raster(left, right, result)
    raster(right, right_source, result)
    local connected = {}
    for _, tile in pairs(result) do connected[#connected + 1] = tile end
    return connected
end

local function closest_path_x(path, wall_y)
    local best_x, distance
    for _, tile in ipairs(path) do
        local candidate = math.abs(tile.y - wall_y)
        if distance == nil or candidate < distance or candidate == distance and tile.x < best_x then
            best_x, distance = tile.x, candidate
        end
    end
    return best_x
end

local function river_plan(seed, columns, rows, entry_direction, path)
    local crosses = entry_direction == "west" or entry_direction == "east"
    local plan = {column = math.floor(columns / 2) - 1, row = math.floor(rows / 2), crosses_route = crosses}
    if not crosses then
        plan.column = columns - 2
        plan.row = draw(seed, "d2legacy.mapgen.outdoor.bridge_row", rows)
        return plan
    end
    for _, tile in ipairs(path) do
        if tile.x >= plan.column * CELL and tile.x < (plan.column + 2) * CELL then
            plan.row = math.floor(tile.y / CELL)
            break
        end
    end
    return plan
end

local function cliff_plan(seed, rows, path)
    local row = math.floor(rows / 4)
    if draw(seed, "d2legacy.mapgen.outdoor.cliff_side", 2) == 1 then row = rows - 3 end
    local wall_y = row * CELL + 5
    return {row = row, opening_column = math.floor(closest_path_x(path, wall_y) / CELL)}
end

local function semantic_tiles(columns, rows, river, cliff)
    local result = {}
    for y = 0, rows * CELL - 1 do
        for x = river.column * CELL, (river.column + 2) * CELL - 1 do
            local kind, passable = "river", false
            local local_y, local_x = y % CELL, x - river.column * CELL
            if math.floor(y / CELL) == river.row and local_y >= 2 and local_y <= 5 then
                kind = "bridge"
                passable = local_y == 3 or local_y == 4
                    or (local_y == 2 or local_y == 5) and (local_x < 4 or local_x >= 12)
            end
            result[#result + 1] = {x = x, y = y, kind = kind, passable = passable}
        end
    end
    local wall_y = cliff.row * CELL + 5
    for x = 0, columns * CELL - 1 do
        local column = math.floor(x / CELL)
        if column ~= cliff.opening_column and column ~= river.column and column ~= river.column + 1 then
            result[#result + 1] = {x = x, y = wall_y, kind = "cliff", passable = false}
        end
    end
    return result
end

local function overlay_stamp(level_type, preset, variant, id, x, y, role)
    local variants = assets.preset_files(preset)
    assert(variants[variant + 1], "d2legacy mapgen: structural preset lacks required variant")
    return {
        id = id, preset_def = data.integer(preset.Def, 0), role = role or "blood-moor-structure",
        x = x, y = y, width = data.integer(preset.SizeX, 0), height = data.integer(preset.SizeY, 0),
        ds1_path = assets.path(variants[variant + 1]),
        tile_paths = assets.masked_tiles(level_type, data.integer(preset.Dt1Mask, 0)),
        variant = variant, logical_walls = preset.Logicals == "1", overlay = true,
    }
end

local function authored_stamps(columns, rows, level_type, river, cliff)
    local stamps = {}
    for row = 0, rows - 1 do
        for side = 0, 1 do
            local definition, variant = 26 + side, 3
            if row == river.row then definition, variant = 28, side == 0 and 0 or 2 end
            local preset = assert(data.preset_by_definition(definition), "d2legacy mapgen: structural LvlPrest is absent")
            stamps[#stamps + 1] = overlay_stamp(level_type, preset, variant,
                columns * rows + row * 2 + side + 1, (river.column + side) * CELL, row * CELL)
        end
    end
    for column = 0, columns - 1 do
        if column ~= cliff.opening_column and column ~= river.column and column ~= river.column + 1 then
            local preset = assert(data.preset_by_definition(17), "d2legacy mapgen: straight-cliff LvlPrest is absent")
            stamps[#stamps + 1] = overlay_stamp(level_type, preset, 0,
                columns * rows + rows * 2 + column + 1, column * CELL, cliff.row * CELL, "blood-moor-cliff")
        end
    end
    return stamps
end

function M.plan(seed, columns, rows, entry_direction, path, level_type)
    local river = river_plan(seed, columns, rows, entry_direction, path)
	if river.crosses_route then
		path = path_through_bridge(path, river.column * CELL, river.row * CELL + CELL / 2)
	end
    local cliff = cliff_plan(seed, rows, path)
    return {
        river = river,
        cliff = cliff,
        tiles = semantic_tiles(columns, rows, river, cliff),
        stamps = authored_stamps(columns, rows, level_type, river, cliff),
		path = path,
    }
end

return M
