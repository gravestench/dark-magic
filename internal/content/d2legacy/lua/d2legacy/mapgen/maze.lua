-- Authoritative Act I cave maze assembly.
--
-- Topology, authored-room selection, and engine recipe admission are kept in
-- separate modules. This file is the short policy coordinator which reads the
-- legacy rows and turns cells into positioned DS1 stamps.

local assets = require("d2legacy.mapgen.assets")
local caves = require("d2legacy.mapgen.cave_rooms")
local data = require("d2legacy.mapgen.records")
local draw = require("engine.deterministic/v1").integer
local topology = require("d2legacy.mapgen.maze_topology")
local worldgen = require("engine.worldgen/v1")
local M = {}

local function room_count(rules, difficulty)
    local column = ({[0] = "Rooms", [1] = "Rooms(N)", [2] = "Rooms(H)"})[difficulty]
    assert(column, "d2legacy mapgen: difficulty must be 0, 1, or 2")
    return data.integer(rules[column], 0)
end

local function sorted_cells(cells)
    table.sort(cells, function(a, b) return a.y < b.y or a.y == b.y and a.x < b.x end)
    return cells
end

function M.generate(level_id, seed, difficulty)
    level_id = math.floor(assert(tonumber(level_id), "d2legacy mapgen: level ID is required"))
    seed = math.floor(assert(tonumber(seed), "d2legacy mapgen: seed is required"))
    difficulty = math.floor(tonumber(difficulty) or 0)
    local level = assert(data.level(level_id), "d2legacy mapgen: level is absent from Levels")
    assert(data.integer(level.DrlgType, -1) == 1, "d2legacy mapgen: level is not maze DRLG type 1")
    assert(data.integer(level.LevelType, -1) == 3, "d2legacy mapgen: current maze strategy supports Act I caves")
    local rules = assert(data.maze_for_level(level_id), "d2legacy mapgen: level has no LvlMaze rules")
    local width, height = data.integer(rules.SizeX, 0), data.integer(rules.SizeY, 0)
    assert(width > 0 and height > 0, "d2legacy mapgen: maze room dimensions are invalid")

    local cells, edges, roles = topology.generate(seed, room_count(rules, difficulty), data.integer(rules.Merge, 0))
	sorted_cells(cells)
    local ids, min_x, min_y, max_x, max_y = {}, cells[1].x, cells[1].y, cells[1].x, cells[1].y
    for index, cell in ipairs(cells) do
        ids[topology.key(cell.x, cell.y)] = index
        min_x, min_y = math.min(min_x, cell.x), math.min(min_y, cell.y)
        max_x, max_y = math.max(max_x, cell.x), math.max(max_y, cell.y)
    end

    local stamps, rooms = {}, {}
    local level_type = assert(data.level_type(3), "d2legacy mapgen: cave LvlTypes row is absent")
    for _, cell in ipairs(cells) do
        local name = topology.key(cell.x, cell.y)
		local id, role, mask = ids[name], roles[name] or "", caves.connection_mask(cell, edges)
		if mask == 0 and #cells > 1 then
			local rendered = {}
			for _, edge in ipairs(edges) do
				rendered[#rendered + 1] = string.format("%s,%s>%s,%s", edge.ax, edge.ay, edge.bx, edge.by)
			end
			error("d2legacy mapgen: disconnected cave cell " .. cell.x .. "," .. cell.y .. " across " .. table.concat(rendered, ";"))
		end
        local definition = caves.preset_definition(role, mask)
		local preset = assert(data.preset_by_definition(definition),
			"d2legacy mapgen: cave LvlPrest definition " .. definition .. " is absent (mask " .. mask .. ", role " .. role .. ")")
        assert(data.integer(preset.SizeX, 0) == width and data.integer(preset.SizeY, 0) == height,
            "d2legacy mapgen: cave room dimensions do not match LvlMaze")
        local variants = assets.preset_files(preset)
        assert(#variants > 0, "d2legacy mapgen: cave preset has no DS1 variants")
        local variant = draw(seed, "d2legacy.mapgen.maze.room_variant", #variants, id)
        local x, y = (cell.x - min_x) * width, (cell.y - min_y) * height
        stamps[#stamps + 1] = {
            id = id, preset_def = definition, role = role,
            x = x, y = y, width = width, height = height,
            ds1_path = assets.path(variants[variant + 1]),
            tile_paths = assets.masked_tiles(level_type, data.integer(preset.Dt1Mask, 0)),
            variant = variant, populate = preset.Populate == "1", logical_walls = preset.Logicals == "1",
        }
        rooms[#rooms + 1] = {id = id, x = x, y = y, width = width, height = height, stamp_id = id}
    end

    local links = {}
    for _, edge in ipairs(edges) do
		links[#links + 1] = {from = ids[topology.key(edge.ax, edge.ay)], to = ids[topology.key(edge.bx, edge.by)]}
    end
    return assert(worldgen.admit({
        request = {version = 1, seed = seed, act = data.integer(level.Act, 0) + 1, level_id = level_id, difficulty = difficulty},
        kind = "maze", bounds = {x = 0, y = 0, width = (max_x - min_x + 1) * width, height = (max_y - min_y + 1) * height},
        stamps = stamps, rooms = rooms, links = links,
        trace = {
            string.format("LvlMaze[%d] requested %d rooms of %dx%d", level_id, #rooms, width, height),
            string.format("topology created %d canonical room links", #links),
            "Act I Cave chamber definitions selected by exact W/E/S/N masks",
            "distinct leaf chambers assigned previous-level and next-level roles",
        },
    }))
end

return M
