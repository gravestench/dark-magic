-- Authoritative Diablo II preset-level generation.
--
-- This module owns DRLG type 2, difficulty sizes, DT1 masks, authored variant
-- selection, and Rogue Encampment exit classification. The engine merely
-- validates and canonicalizes the completed recipe.

local assets = require("d2legacy.mapgen.assets")
local data = require("d2legacy.mapgen.records")
local draw = require("engine.deterministic/v1").integer
local worldgen = require("engine.worldgen/v1")
local M = {}

local function size_for_difficulty(level, difficulty)
    local suffix = ({[0] = "", [1] = "(N)", [2] = "(H)"})[difficulty]
    assert(suffix ~= nil, "d2legacy mapgen: difficulty must be 0, 1, or 2")
    return data.integer(level["SizeX" .. suffix], 0), data.integer(level["SizeY" .. suffix], 0)
end

local function town_role(ds1_path)
    local name = assets.basename(ds1_path):lower()
    local directions = {townn = "north", towne = "east", towns = "south", townw = "west"}
    for marker, direction in pairs(directions) do
        if name:sub(1, #marker) == marker then return "act1-town:exit-" .. direction end
    end
    return "act1-town"
end

function M.generate(level_id, seed, difficulty)
    level_id = math.floor(assert(tonumber(level_id), "d2legacy mapgen: level ID is required"))
    seed = math.floor(assert(tonumber(seed), "d2legacy mapgen: seed is required"))
    difficulty = math.floor(tonumber(difficulty) or 0)
    local level = assert(data.level(level_id), "d2legacy mapgen: level is absent from Levels")
    assert(data.integer(level.DrlgType, -1) == 2, "d2legacy mapgen: level is not preset DRLG type 2")
    local preset = assert(data.preset_for_level(level_id), "d2legacy mapgen: level has no LvlPrest definition")
    local variants = assets.preset_files(preset)
    assert(#variants > 0, "d2legacy mapgen: preset has no DS1 variants")
    local variant = draw(seed, "d2legacy.mapgen.preset.variant", #variants)

    local width, height = data.integer(preset.SizeX, 0), data.integer(preset.SizeY, 0)
    if width <= 0 or height <= 0 then width, height = size_for_difficulty(level, difficulty) end
    assert(width > 0 and height > 0, "d2legacy mapgen: preset has no positive dimensions")

    local level_type_id = data.integer(level.LevelType, -1)
    local level_type = assert(data.level_type(level_type_id), "d2legacy mapgen: LvlTypes row is absent")
    local tiles = assets.masked_tiles(level_type, data.integer(preset.Dt1Mask, 0))
    assert(#tiles > 0, "d2legacy mapgen: DT1 mask selected no files")
    local act = data.integer(level.Act, 0) + 1
    local ds1_path = assets.path(variants[variant + 1])
    local role = act == 1 and level_id == 1 and town_role(ds1_path) or ""

    return assert(worldgen.admit({
        request = {version = 1, seed = seed, act = act, level_id = level_id, difficulty = difficulty},
        kind = "preset",
        bounds = {x = 0, y = 0, width = width, height = height},
        stamps = {{
            id = 1, preset_def = data.integer(preset.Def, 0), role = role,
            x = 0, y = 0, width = width, height = height,
            ds1_path = ds1_path, tile_paths = tiles, variant = variant,
            populate = preset.Populate == "1", logical_walls = preset.Logicals == "1",
        }},
        rooms = {{id = 1, x = 0, y = 0, width = width, height = height, stamp_id = 1}},
        trace = {
            string.format("Levels[%d] selected preset DRLG type 2", level_id),
            string.format("LvlPrest[%s] selected variant %d of %d", preset.Def, variant + 1, #variants),
            string.format("LvlTypes[%d] mask %s selected %d DT1 files", level_type_id, preset.Dt1Mask, #tiles),
        },
    }))
end

return M
