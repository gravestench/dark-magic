-- Deterministic room topology for Diablo II's maze DRLG family.
--
-- Cells use integer grid coordinates. Edges are canonicalized so generation,
-- checksums, and tests never depend on Lua table traversal order.

local draw = require("engine.deterministic/v1").integer
local M = {}

local directions = {{-1, 0}, {1, 0}, {0, 1}, {0, -1}}

local function key(x, y)
	-- Lua numbers may preserve a negative-zero representation after arithmetic.
	-- Spatially, -0 and 0 are the same cell, so normalize before making a key.
	if x == 0 then x = 0 end
	if y == 0 then y = 0 end
	return x .. ":" .. y
end

local function before(a, b)
    return a.y < b.y or a.y == b.y and a.x < b.x
end

local function edge(ax, ay, bx, by)
	if by < ay or by == ay and bx < ax then
		return {
			ax = bx, ay = by, bx = ax, by = ay,
			key = key(bx, by) .. ">" .. key(ax, ay),
		}
	end
    return {
		ax = ax, ay = ay, bx = bx, by = by,
		key = key(ax, ay) .. ">" .. key(bx, by),
    }
end

local function degrees(edges)
    local result = {}
    for _, link in ipairs(edges) do
        result[key(link.ax, link.ay)] = (result[key(link.ax, link.ay)] or 0) + 1
        result[key(link.bx, link.by)] = (result[key(link.bx, link.by)] or 0) + 1
    end
    return result
end

local function grow(seed, count)
    local cells = {{x = 0, y = 0}}
    local links = {}
    local occupied = {[key(0, 0)] = true}
    while #cells < count do
        local frontier = {}
        for _, from in ipairs(cells) do
            for _, delta in ipairs(directions) do
                local to = {x = from.x + delta[1], y = from.y + delta[2]}
                if not occupied[key(to.x, to.y)] then
					frontier[#frontier + 1] = {ax = from.x, ay = from.y, bx = to.x, by = to.y}
                end
            end
        end
        local ordinal = #cells - 1
        local picked = frontier[draw(seed, "d2legacy.mapgen.maze.grow", #frontier, ordinal) + 1]
		occupied[key(picked.bx, picked.by)] = true
		cells[#cells + 1] = {x = picked.bx, y = picked.by}
		links[#links + 1] = edge(picked.ax, picked.ay, picked.bx, picked.by)
    end
    return cells, links, occupied
end

local function merge(seed, cells, links, occupied, chance)
    local existing = {}
    for _, link in ipairs(links) do existing[link.key] = true end
    local room_degrees = degrees(links)
    chance = math.max(0, math.min(1000, chance))
    local ordinal = 0
    for _, cell in ipairs(cells) do
        for _, delta in ipairs({{1, 0}, {0, 1}}) do
            local other = {x = cell.x + delta[1], y = cell.y + delta[2]}
			local candidate = edge(cell.x, cell.y, other.x, other.y)
            if occupied[key(other.x, other.y)] and not existing[candidate.key]
                and (room_degrees[key(cell.x, cell.y)] or 0) > 1
                and (room_degrees[key(other.x, other.y)] or 0) > 1
                and draw(seed, "d2legacy.mapgen.maze.merge", 1000, ordinal) < chance then
                existing[candidate.key] = true
                links[#links + 1] = candidate
            end
            ordinal = ordinal + 1
        end
    end
end

local function distances(start, links)
    local adjacent = {}
    local function join(a, b)
        local name = key(a.x, a.y)
        adjacent[name] = adjacent[name] or {}
        adjacent[name][#adjacent[name] + 1] = b
    end
    for _, link in ipairs(links) do
        join({x = link.ax, y = link.ay}, {x = link.bx, y = link.by})
        join({x = link.bx, y = link.by}, {x = link.ax, y = link.ay})
    end
    local result = {[key(start.x, start.y)] = 0}
    local queue = {start}
    local cursor = 1
    while cursor <= #queue do
        local current = queue[cursor]
        cursor = cursor + 1
        for _, next_cell in ipairs(adjacent[key(current.x, current.y)] or {}) do
            local name = key(next_cell.x, next_cell.y)
            if result[name] == nil then
                result[name] = result[key(current.x, current.y)] + 1
                queue[#queue + 1] = next_cell
            end
        end
    end
    return result
end

local function assign_special_roles(seed, cells, links)
    if #cells == 1 then return {[key(0, 0)] = "previous-level"} end
    local room_degrees = degrees(links)
    local leaves = {}
    for _, cell in ipairs(cells) do
        if room_degrees[key(cell.x, cell.y)] == 1 then leaves[#leaves + 1] = cell end
    end
    table.sort(leaves, before)
    assert(#leaves >= 2, "d2legacy mapgen: maze has fewer than two endpoint rooms")
	local entrance = leaves[draw(seed, "d2legacy.mapgen.maze.entrance", #leaves) + 1]
	local entrance_key = key(entrance.x, entrance.y)
	local distance = distances(entrance, links)
	local exit = nil
	for _, candidate in ipairs(leaves) do
		local candidate_key = key(candidate.x, candidate.y)
		if candidate_key ~= entrance_key and (exit == nil
			or (distance[candidate_key] or -1) > (distance[key(exit.x, exit.y)] or -1)) then
            exit = candidate
        end
    end
    return {
        [key(entrance.x, entrance.y)] = "previous-level",
        [key(exit.x, exit.y)] = "next-level",
    }
end

function M.generate(seed, count, merge_chance)
    assert(count > 0, "d2legacy mapgen: maze room count must be positive")
    local cells, links, occupied = grow(seed, count)
    merge(seed, cells, links, occupied, merge_chance)
    return cells, links, assign_special_roles(seed, cells, links)
end

M.key = key

return M
