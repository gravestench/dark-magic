-- Authoritative Blood Moor recipe assembly.

local assets = require("d2legacy.mapgen.assets")
local data = require("d2legacy.mapgen.records")
local draw = require("engine.deterministic/v1").integer
local route_policy = require("d2legacy.mapgen.outdoor_route")
local structure_policy = require("d2legacy.mapgen.outdoor_structures")
local dirt_path = require("d2legacy.mapgen.act1_dirt_path")
local worldgen = require("engine.worldgen/v1")
local M = {}
local CELL = 8

local function level_size(level, difficulty)
    local suffix = ({[0] = "", [1] = "(N)", [2] = "(H)"})[difficulty]
    assert(suffix, "d2legacy mapgen: difficulty must be 0, 1, or 2")
    return data.integer(level["SizeX" .. suffix], 0), data.integer(level["SizeY" .. suffix], 0)
end

local function fill_stamp(seed, id, x, y, route_cells, level_type)
    local definitions = {29, 30, 35}
    local definition = definitions[draw(seed, "d2legacy.mapgen.outdoor.fill", #definitions, id) + 1]
    local preset = assert(data.preset_by_definition(definition), "d2legacy mapgen: Blood Moor fill LvlPrest is absent")
    assert(data.integer(preset.SizeX, 0) == CELL and data.integer(preset.SizeY, 0) == CELL,
        "d2legacy mapgen: Blood Moor fill must be 8x8")
    local variants = assets.preset_files(preset)
    local variant = draw(seed, "d2legacy.mapgen.outdoor.fill_variant", #variants, id)
    local cell_key = math.floor(x / CELL) .. ":" .. math.floor(y / CELL)
    return {
        id = id, preset_def = definition,
        role = route_cells[cell_key] and "blood-moor-route" or "blood-moor-fill",
        x = x, y = y, width = CELL, height = CELL,
        ds1_path = assets.path(variants[variant + 1]),
        tile_paths = assets.masked_tiles(level_type, data.integer(preset.Dt1Mask, 0)),
        variant = variant, populate = preset.Populate == "1", logical_walls = preset.Logicals == "1",
    }
end

function M.generate(level_id, seed, town_exit, difficulty)
    level_id, seed = math.floor(assert(tonumber(level_id))), math.floor(assert(tonumber(seed)))
    difficulty = math.floor(tonumber(difficulty) or 0)
    assert(level_id == 2, "d2legacy mapgen: current outdoor strategy supports Blood Moor level 2")
    assert(({north=true,east=true,south=true,west=true})[town_exit], "d2legacy mapgen: town exit must be cardinal")
    local level = assert(data.level(level_id), "d2legacy mapgen: Blood Moor is absent from Levels")
    assert(data.integer(level.Act, -1) == 0 and data.integer(level.DrlgType, -1) == 3
        and data.integer(level.LevelType, -1) == 2, "d2legacy mapgen: Blood Moor level rules do not match")
    local width, height = level_size(level, difficulty)
    assert(width > 0 and height > 0 and width % CELL == 0 and height % CELL == 0,
        "d2legacy mapgen: Blood Moor dimensions are not an 8-tile grid")
    local columns, rows = width / CELL, height / CELL
    local route = route_policy.plan(seed, width, height, town_exit)
    local level_type = assert(data.level_type(2), "d2legacy mapgen: Blood Moor LvlTypes row is absent")
    local structures = structure_policy.plan(seed, columns, rows, route.entry.direction, route.path, level_type)

    local stamps, rooms, links = {}, {}, {}
    for y = 0, rows - 1 do
        for x = 0, columns - 1 do
            local id = y * columns + x + 1
            local stamp = fill_stamp(seed, id, x * CELL, y * CELL, route.cells, level_type)
            stamps[#stamps + 1] = stamp
            rooms[#rooms + 1] = {id = id, x = stamp.x, y = stamp.y, width = CELL, height = CELL, stamp_id = id}
            if x > 0 then links[#links + 1] = {from = id - 1, to = id} end
            if y > 0 then links[#links + 1] = {from = id - columns, to = id} end
        end
    end
    for _, stamp in ipairs(structures.stamps) do stamps[#stamps + 1] = stamp end
    return assert(worldgen.admit({
        request = {version = 1, seed = seed, act = 1, level_id = level_id, difficulty = difficulty},
        kind = "outdoor", bounds = {x = 0, y = 0, width = width, height = height},
        stamps = stamps, rooms = rooms, links = links,
        warps = {route.entry, route.exit}, paths = dirt_path.apply(structures.path), structures = structures.tiles,
        trace = {
            string.format("Levels[%d] selected Act I outdoor strategy on a %dx%d coarse grid", level_id, columns, rows),
            "authored 8x8 Blood Moor fill presets selected by independent cell keys",
            string.format("Rogue Encampment joins the %s Blood Moor edge", route_policy.opposite(town_exit)),
            string.format("a deterministic %d-cell route joins town to the opposite next-level edge", #route.ordered),
            "a continuous river and cliff ridge preserve explicit passable crossings on the primary route",
        },
    }))
end

return M
