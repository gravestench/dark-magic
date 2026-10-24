-- Stony Field structural zone policy. The fill choice, edge placement, and
-- route are deterministic Dark Magic policy; original 1.13c/1.14d placement
-- and seed call order remain unverified.

local assets = require("d2legacy.mapgen.assets")
local data = require("d2legacy.mapgen.records")
local draw = require("engine.deterministic/v1").integer
local dirt_path = require("d2legacy.mapgen.act1_dirt_path")
local worldgen = require("engine.worldgen/v1")
local M = {}

local CELL = 8
local opposite = {north = "south", east = "west", south = "north", west = "east"}
local fill_definitions = {29, 30}

-- Reject fractional identities before shared parsers can round them.
local function integer(value)
    local parsed = tonumber(value)
    assert(parsed and parsed > -math.huge and parsed < math.huge and parsed == math.floor(parsed),
        "d2legacy mapgen: Stony Field numeric fields must be finite integers")
    return parsed
end

-- Keep each difficulty's dimensions aligned to the fixed-size DS1 fill grid.
local function selected_size(level, difficulty)
    local suffix = ({[0] = "", [1] = "(N)", [2] = "(H)"})[difficulty]
    assert(suffix, "d2legacy mapgen: Stony Field difficulty must be 0, 1, or 2")
    local width = integer(level["SizeX" .. suffix])
    local height = integer(level["SizeY" .. suffix])
    assert(width > 0 and height > 0 and width % CELL == 0 and height % CELL == 0,
        "d2legacy mapgen: Stony Field dimensions must use an 8-tile grid")
    return width, height
end

-- Check the original table field before normalization can turn an empty or
-- directory-only value into an apparently usable tile-root path.
local function valid_asset_reference(value, extension)
    if type(value) ~= "string" then return false end
    local normalized = value:match("^%s*(.-)%s*$"):gsub("\\", "/")
    if normalized == "" or normalized:sub(1, 1) == "/"
        or normalized:find(":", 1, true) or normalized:find("//", 1, true)
        or not normalized:lower():match("%." .. extension .. "$") then return false end
    for segment in normalized:gmatch("[^/]+") do
        if segment == "." or segment == ".." then return false end
    end
    return true
end

-- The shared asset helpers omit blank slots, so verify every declared DS1
-- variant and every mask-selected DT1 before building an admissible recipe.
local function validate_fill_references(row, level_type)
    local declared = tonumber(row.Files) or 0
    assert(declared >= 1 and declared <= 6 and declared == math.floor(declared),
        "d2legacy mapgen: Stony Field stone-fill preset has invalid DS1 variant count")
    for index = 1, declared do
        assert(valid_asset_reference(row["File" .. index], "ds1"),
            "d2legacy mapgen: Stony Field stone-fill preset has invalid DS1 reference")
    end

    local mask = tonumber(row.Dt1Mask) or 0
    assert(mask > 0 and mask < 2 ^ 32 and mask == math.floor(mask),
        "d2legacy mapgen: Stony Field stone-fill preset has invalid DT1 mask")
    for index = 1, 32 do
        local bit = 2 ^ (index - 1)
        if math.floor(mask / bit) % 2 == 1 then
            local reference = level_type["File " .. index] or level_type["File" .. index]
            assert(valid_asset_reference(reference, "dt1"),
                "d2legacy mapgen: Stony Field stone-fill preset has invalid DT1 reference")
        end
    end
    return mask
end

-- Inspected 1.14d LvlPrest definitions 29/30 are generic outdoor fill.
-- The runtime records may come from another version or a synthetic fixture.
local function fill_candidates(level_type)
    local candidates = {}
    for _, definition in ipairs(fill_definitions) do
        local row = assert(data.preset_by_definition(definition),
            "d2legacy mapgen: Stony Field stone-fill preset is absent")
        assert(integer(row.Def) == definition and integer(row.LevelId) == 0 and integer(row.Outdoors) == 1
            and integer(row.SizeX) == CELL and integer(row.SizeY) == CELL,
            "d2legacy mapgen: Stony Field stone-fill preset is not a generic 8x8 outdoor piece")
        local mask = validate_fill_references(row, level_type)
        local variants = assets.preset_files(row)
        local tile_paths = assets.masked_tiles(level_type, mask)
        assert(#variants > 0 and #tile_paths > 0,
            "d2legacy mapgen: Stony Field stone-fill preset lacks DS1 or DT1 references")
        candidates[#candidates + 1] = {row = row, variants = variants, tile_paths = tile_paths}
    end
    return candidates
end

-- This warp record is a semantic outdoor border endpoint. A future seam
-- supplies the source direction; it does not describe a discrete portal.
local function entry_warp(width, height, direction)
    local x, y = math.floor(width / 2), math.floor(height / 2)
    if direction == "north" then y = 0
    elseif direction == "east" then x = width - 1
    elseif direction == "south" then y = height - 1
    else x = 0 end
    return {id = 1, role = "cold-plains-entry", direction = direction,
        x = x, y = y, destination_level = 3}
end

-- A semantic route reserves a connected corridor through provisional fill.
local function route_to_interior(entry, width, height)
    local path, x, y = {}, entry.x, entry.y
    local goal_x, goal_y = math.floor(width / 2), math.floor(height / 2)
    while true do
        path[#path + 1] = {x = x, y = y}
        if x == goal_x and y == goal_y then break end
        if x < goal_x then x = x + 1
        elseif x > goal_x then x = x - 1
        elseif y < goal_y then y = y + 1
        else y = y - 1 end
    end
    return dirt_path.apply(path)
end

-- Admit a canonical structural recipe using named, Stony Field-only streams.
function M.generate(level_id, seed, source_direction, difficulty)
    level_id, seed = integer(level_id), integer(seed)
    difficulty = integer(difficulty == nil and 0 or difficulty)
    assert(level_id == 4, "d2legacy mapgen: Stony Field strategy requires level 4")
    local direction = assert(opposite[source_direction],
        "d2legacy mapgen: Cold Plains exit direction must be cardinal")
    local level = assert(data.level(level_id), "d2legacy mapgen: Stony Field is absent from Levels")
    assert(integer(level.Act) == 0 and integer(level.DrlgType) == 3
        and integer(level.LevelType) == 2,
        "d2legacy mapgen: Stony Field level rules do not match")
    local width, height = selected_size(level, difficulty)
    local level_type = assert(data.level_type(2), "d2legacy mapgen: Act I outdoor LvlTypes row is absent")
    local candidates = fill_candidates(level_type)
    local columns = width / CELL
    local rows = height / CELL
    local stamps, rooms, links = {}, {}, {}
    for y = 0, rows - 1 do
        for x = 0, columns - 1 do
            local id = y * columns + x + 1
            local choice = candidates[draw(seed, "d2legacy.mapgen.stony_field.fill", #candidates, id) + 1]
            local variant = draw(seed, "d2legacy.mapgen.stony_field.fill_variant", #choice.variants, id)
            local tile_x, tile_y = x * CELL, y * CELL
            stamps[#stamps + 1] = {
                id = id, preset_def = integer(choice.row.Def), role = "stony-field-fill",
                x = tile_x, y = tile_y, width = CELL, height = CELL,
                ds1_path = assets.path(choice.variants[variant + 1]), tile_paths = choice.tile_paths,
                variant = variant, populate = choice.row.Populate == "1",
                logical_walls = choice.row.Logicals == "1",
            }
            rooms[#rooms + 1] = {id = id, x = tile_x, y = tile_y,
                width = CELL, height = CELL, stamp_id = id}
            if x > 0 then links[#links + 1] = {from = id - 1, to = id} end
            if y > 0 then links[#links + 1] = {from = id - columns, to = id} end
        end
    end
    local entry = entry_warp(width, height, direction)
    return assert(worldgen.admit({
        request = {version = 1, seed = seed, act = 1, level_id = level_id, difficulty = difficulty},
        kind = "outdoor", bounds = {x = 0, y = 0, width = width, height = height},
        stamps = stamps, rooms = rooms, links = #links > 0 and links or nil, warps = {entry},
        paths = route_to_interior(entry, width, height),
        trace = {
            "Policy evidence: inspected 1.14d Levels level 4 is Act I outdoor DRLG 3, type 2, 80x80 in all modes",
            "Policy evidence: inspected 1.14d LvlTypes row 2 names Act I wilderness DT1 libraries",
            "Policy evidence: inspected 1.14d LvlPrest defs 29 and 30 qualify as generic 8x8 outdoor stone "
                .. "fill with three DS1 variants each",
            "Runtime records are supplied independently; the inspected evidence does not establish their version",
            string.format("Cold Plains %s source joins the Stony Field %s semantic outdoor border endpoint",
                source_direction, direction),
            "fill arrangement, entry placement, and interior route are provisional deterministic Dark Magic "
                .. "structural policy",
            "original layout and seed call order remain unverified",
        },
    }))
end

return M
