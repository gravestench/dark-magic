-- Presentation placement only: maps retain their own gameplay coordinates.
local M = {}
local PIXEL_TOLERANCE = 0.5
-- world.TileDepth advances 100 per tile diagonal, projected as 40 pixels.
local DEPTH_PER_PIXEL = 100 / 40

local function finite(value)
    return type(value) == "number" and value == value and value ~= math.huge and value ~= -math.huge
end

local function project(map, x, y)
    assert(finite(x) and finite(y), "outdoor seam coordinates must be finite")
    local px, py = map:subtile_to_pixel(x, y)
    assert(finite(px) and finite(py), "outdoor seam projection must be finite")
    return px, py
end

local function aligned(x, y, expected_x, expected_y)
    return math.abs(x - expected_x) <= PIXEL_TOLERANCE and math.abs(y - expected_y) <= PIXEL_TOLERANCE
end

function M.layout(snapshot, directed_seams)
    assert(type(snapshot) == "table" and type(snapshot.maps) == "table", "outdoor maps are required")
    assert(type(directed_seams) == "table", "outdoor seams are required")
    local levels, ids, edges = {}, {}, {}
    for _, entry in ipairs(snapshot.maps) do
        local id = entry.level_id
        assert(finite(id) and id > 0 and id == math.floor(id), "outdoor level ID must be a positive integer")
        assert(not levels[id], "duplicate outdoor level")
        assert(entry.map, "outdoor map is required")
        local width, height = entry.map:canvas()
        assert(finite(width) and finite(height) and width > 0 and height > 0, "outdoor canvas must be positive")
        levels[id] = {
            map = entry.map,
            recipe = { world = entry.map, ds1 = entry.ds1, dt1 = entry.dt1 },
            width = width,
            height = height,
        }
        ids[#ids + 1] = id
        edges[id] = {}
    end
    assert(#ids > 0 and levels[snapshot.active_level], "active outdoor level must be prepared")
    table.sort(ids)

    for _, seam in ipairs(directed_seams) do
        assert(type(seam) == "table", "outdoor seam must be a table")
        local source, destination = seam.source_level, seam.destination_level
        assert(finite(source) and finite(destination) and source ~= destination, "invalid outdoor seam levels")
        -- A prepared two-zone snapshot can still publish seams to a third,
        -- unavailable zone. Only edges whose two maps exist participate.
        if levels[source] and levels[destination] then
            assert(not edges[source][destination], "duplicate directed outdoor seam")
            local sx, sy = project(levels[source].map, seam.source_x, seam.source_y)
            local ax, ay = project(levels[destination].map, seam.arrival_x, seam.arrival_y)
            edges[source][destination] = { x = sx - ax, y = sy - ay }
        end
    end
    for _, source in ipairs(ids) do
        for destination in pairs(edges[source]) do
            assert(edges[destination][source], "outdoor seam must have a reciprocal edge")
        end
    end

    -- Anchor to a stable level, never to the changing authoritative level.
    local anchor = levels[ids[1]]
    anchor.offset_x, anchor.offset_y = 0, 0
    local queue, next_index = { ids[1] }, 1
    while next_index <= #queue do
        local source = queue[next_index]
        next_index = next_index + 1
        for _, destination in ipairs(ids) do
            local edge = edges[source][destination]
            if edge then
                local target = levels[destination]
                local x = levels[source].offset_x + edge.x
                local y = levels[source].offset_y + edge.y
                if target.offset_x == nil then
                    target.offset_x, target.offset_y = x, y
                    queue[#queue + 1] = destination
                else
                    assert(
                        aligned(x, y, target.offset_x, target.offset_y),
                        "outdoor seam alignment residual exceeds tolerance"
                    )
                end
            end
        end
    end

    local bounds = { min_x = math.huge, min_y = math.huge, max_x = -math.huge, max_y = -math.huge }
    for _, id in ipairs(ids) do
        local level = levels[id]
        assert(level.offset_x ~= nil, "outdoor maps must be connected by prepared seams")
        level.depth_offset = level.offset_y * DEPTH_PER_PIXEL
        bounds.min_x = math.min(bounds.min_x, level.offset_x)
        bounds.min_y = math.min(bounds.min_y, level.offset_y)
        bounds.max_x = math.max(bounds.max_x, level.offset_x + level.width)
        bounds.max_y = math.max(bounds.max_y, level.offset_y + level.height)
    end
    bounds.width, bounds.height = bounds.max_x - bounds.min_x, bounds.max_y - bounds.min_y
    return { levels = levels, bounds = bounds }
end

function M.pixel(level, map_pixel_x, map_pixel_y, layout)
    local placement = assert(layout.levels[level], "outdoor level is not prepared")
    return map_pixel_x + placement.offset_x, map_pixel_y + placement.offset_y
end

function M.depth(level, local_depth, layout)
    local placement = assert(layout.levels[level], "outdoor level is not prepared")
    -- Lower walls, floor, shadow and roofs use fixed pass bands. Only the
    -- comparable middle pass (upper walls and entities) follows map placement.
    if local_depth <= -1000000 or local_depth >= 1000000 then
        return local_depth
    end
    return local_depth + placement.depth_offset
end

return M
