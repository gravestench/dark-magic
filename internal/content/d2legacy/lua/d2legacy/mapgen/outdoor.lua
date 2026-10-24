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

-- key names a route cell without depending on Lua table iteration order.
local function key(x, y) return x .. ":" .. y end

-- Plan a player-sized branch using only semantic passability. The flood is
-- seeded in canonical route order and uses fixed neighbor order, so neither
-- table iteration nor the earlier outdoor streams choose the cave placement.
local function den_approaches(zone)
    local blocked, reached, queue = {}, {}, {}
    local width, height = zone.bounds.width, zone.bounds.height
    for _, tile in ipairs(zone.structures) do
        if not tile.passable then blocked[key(tile.x, tile.y)] = true end
    end
    -- clear requires a three-tile footprint inside bounds and away from blocked structures.
    local function clear(x, y)
        if x < 1 or y < 1 or x >= width - 1 or y >= height - 1 then return false end
        for dy = -1, 1 do
            for dx = -1, 1 do
                if blocked[key(x + dx, y + dy)] then return false end
            end
        end
        return true
    end
    for _, tile in ipairs(zone.paths) do
        local name = key(tile.x, tile.y)
        if clear(tile.x, tile.y) then
            reached[name] = {x = tile.x, y = tile.y}
            queue[#queue + 1] = reached[name]
        end
    end
    local index = 1
    while queue[index] do
        local source = queue[index]
        for _, delta in ipairs({{1,0}, {-1,0}, {0,1}, {0,-1}}) do
            local x, y = source.x + delta[1], source.y + delta[2]
            local name = key(x, y)
            if not reached[name] and clear(x, y) then
                reached[name] = {x = x, y = y, parent = source}
                queue[#queue + 1] = reached[name]
            end
        end
        index = index + 1
    end
    return reached
end

-- den_candidates reserves a cave cell and an approach branch outside earlier authored routes.
local function den_candidates(zone)
    local reached, candidates = den_approaches(zone), {}
    for y = CELL, zone.bounds.height - 2 * CELL, CELL do
        for x = CELL, zone.bounds.width - 2 * CELL, CELL do
            local eligible = true
            for _, stamp in ipairs(zone.stamps) do
                if stamp.overlay and x < stamp.x + stamp.width and stamp.x < x + CELL
                    and y < stamp.y + stamp.height and stamp.y < y + CELL then eligible = false end
            end
            for _, tile in ipairs(zone.paths) do
                if tile.x >= x and tile.x < x + CELL and tile.y >= y and tile.y < y + CELL then eligible = false end
            end
            local approach, path = reached[key(x + CELL / 2, y + CELL + 1)], {}
            if not approach then eligible = false end
            while eligible and approach do
                -- The three-tile branch must not cut through the authored
                -- cave's footprint, including at turns and its route join.
                if approach.x + 1 >= x and approach.x - 1 < x + CELL
                    and approach.y + 1 >= y and approach.y - 1 < y + CELL then eligible = false end
                path[#path + 1] = {x = approach.x, y = approach.y}
                approach = approach.parent
            end
            if eligible then candidates[#candidates + 1] = {x = x, y = y, path = path} end
        end
    end
    return candidates
end

-- V4 alone opts into the Den. Existing fill, structures, and border warps are
-- copied from the original admitted recipe without changing any prior draw.
-- Placement and approach geometry are provisional Dark Magic policy.
function M.generate_with_den(level_id, seed, town_exit, difficulty)
    local zone = M.generate(level_id, seed, town_exit, difficulty)
    local moor = assert(data.level(2))
    local den = assert(data.level(8), "d2legacy mapgen: Den level is absent")
    assert(data.integer(moor.Vis3, -1) == 8 and data.integer(moor.Warp3, -1) == 0
        and data.integer(den.Vis0, -1) == 2 and data.integer(den.Warp0, -1) == 4,
        "d2legacy mapgen: Blood Moor and Den warp relationship is invalid")
    local preset = assert(data.preset_by_definition(51), "d2legacy mapgen: cave entrance LvlPrest is absent")
    assert(data.integer(preset.SizeX, 0) == CELL and data.integer(preset.SizeY, 0) == CELL,
        "d2legacy mapgen: cave entrance must be 8x8")
    local variants = assets.preset_files(preset)
    local count = data.integer(preset.Files, 0)
    assert(count >= 2 and count <= 6 and #variants == count,
        "d2legacy mapgen: cave entrance variant is absent")
    local level_type = assert(data.level_type(2))
    local tiles = assets.masked_tiles(level_type, data.integer(preset.Dt1Mask, 0))
    assert(#tiles > 0, "d2legacy mapgen: cave entrance DT1 resources are absent")
    local candidates = den_candidates(zone)
    assert(#candidates > 0, "d2legacy mapgen: no reserved player-sized Den approach")
    local selected = candidates[draw(seed, "d2legacy.mapgen.den.entrance_placement", #candidates) + 1]
    -- Owned 1.14d File2 carries 10/6/24 for Vis3; File1 carries 10/5/21
    -- for another relationship. Materialization must verify the full identity.
    local eligible = {2}
    local variant = eligible[draw(seed, "d2legacy.mapgen.den.entrance_variant", #eligible) + 1] - 1
    local reserved = selected.y / CELL * (zone.bounds.width / CELL) + selected.x / CELL + 1
    assert(zone.stamps[reserved].id == reserved and not zone.stamps[reserved].overlay,
        "d2legacy mapgen: Den must replace its reserved fill cell")
    -- A separate overlay retains base fill fences on absent authored layers,
    -- splitting the cave's walkable interior. Replace this one fill instead.
    zone.stamps[reserved] = {
        id = reserved, preset_def = 51, role = "den-entrance",
        x = selected.x, y = selected.y, width = CELL, height = CELL,
        ds1_path = assets.path(variants[variant + 1]), tile_paths = tiles,
        variant = variant, populate = preset.Populate == "1", logical_walls = preset.Logicals == "1",
    }
    zone.warps[#zone.warps + 1] = {
        id = 3, role = "den-entrance", direction = "south", destination_level = 8,
        -- Recipe placeholder only: decoded full-identity special tiles own
        -- the eventual interaction anchor, not this coarse cell coordinate.
        x = selected.x + CELL / 2, y = selected.y + CELL - 1,
    }
    local paths = {}
    for _, tile in ipairs(zone.paths) do paths[key(tile.x, tile.y)] = {x = tile.x, y = tile.y} end
    for _, tile in ipairs(selected.path) do
        for dy = -1, 1 do
            for dx = -1, 1 do
                local x, y = tile.x + dx, tile.y + dy
                paths[key(x, y)] = {x = x, y = y}
            end
        end
    end
    zone.paths = {}
    for _, tile in pairs(paths) do zone.paths[#zone.paths + 1] = tile end
    zone.paths = dirt_path.apply(zone.paths)
    zone.trace[#zone.trace + 1] = "definition 51 Den entrance and player-sized branch use independent Den streams; "
        .. "placement and approach are provisional Dark Magic policy"
    return assert(worldgen.admit(zone))
end

return M
