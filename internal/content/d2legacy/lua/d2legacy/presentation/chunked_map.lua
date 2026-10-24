-- Sparse DS1 presentation adapter.
--
-- Think of a DS1 as a very large sheet of paper. Uploading that whole sheet to
-- the GPU is wasteful when the window can see only a small rectangle. This
-- helper keeps the paper's coordinates, but creates retained render nodes only
-- for pictures near the camera. Runtime worlds borrow one texture for every
-- unique DT1 graphic and place it many times. DS1 Lab intentionally keeps the
-- older 512x512 composed chunks because inspecting composition is its job.
--
-- This module owns pictures, never gameplay facts. Callers still use
-- engine.world/v1 for collision, objects, projection, and authoritative positions.

local render = require("engine.render/v1")
local camera_bounds = require("d2legacy.presentation.camera_bounds")

local chunked_map = {}

local function residency_margin(recipe, options)
    local requested = options.margin or 96
    if not recipe.world then return requested end
    -- A side panel moves the camera anchor by one viewport quarter in a single
    -- update. Keep that future strip plus one legacy 160px tile resident before
    -- the panel opens or closes; otherwise the parent transform moves first and
    -- exposes black while the newly visible tile jobs catch up one frame later.
    local panel_shift = (options.viewport_width or 800) / 4
    return math.max(requested, panel_shift + 160)
end

-- Overlay world_view describes camera framing and pointer authority, not an
-- opaque render clip. Legacy panels contain transparent holes and decorative
-- edges, so the world must stay resident behind the whole physical viewport.
-- Otherwise a left/right panel reveals a hard tile boundary; two panels used
-- to produce world_view="none" and freeze residency on both sides.
local function render_viewport(width, height)
    return 0, 0, width, height
end

local function clear_nodes(state)
    for index, node in pairs(state.nodes) do
        if node:exists() then node:destroy() end
        state.nodes[index] = nil
    end
end

local function queue_tile(state, draw)
    local key = draw.index + 1
    state.pending[key] = render.preload({{
        kind = "world_tile",
        world = state.recipe.world,
        palette = state.recipe.palette,
        chunk_index = draw.index,
    }})
end

local function canvas_origin(state)
    if state.shared_root then return state.offset_x, state.offset_y end
    return -state.set.width / 2, -state.set.height / 2
end

local function admit_chunk(state, chunk, key)
    local node = render.create("world", state.root)
    local x, y, width, height
    if state.recipe.world then
        x, y, width, height = node:set_world_tile(
            state.recipe.world, state.recipe.palette, chunk.index
        )
    else
        x, y, width, height = node:set_ds1_chunk(
            state.recipe.ds1, state.recipe.dt1, state.recipe.palette,
            chunk.index, state.chunk_size
        )
    end
    local origin_x, origin_y = canvas_origin(state)
    node:set_position(x + origin_x + width / 2, y + origin_y + height / 2)
    local depth = chunk.depth
    -- Floor/shadow and roof bands remain global passes. Upper walls share
    -- the comparable entity band across all neighboring maps.
    if state.shared_root and depth > -1000000 and depth < 1000000 then
        depth = depth + state.depth_offset
    end
    if state.shared_root then depth = depth * state.depth_scale + state.depth_priority end
    node:set_z(depth)
    state.nodes[key] = node
    state.last_visible[key] = state.revision
end

local function trim_residency(state, visible_keys)
    local resident, hidden = 0, {}
    for key, node in pairs(state.nodes) do
        resident = resident + 1
        if not visible_keys[key] then
            hidden[#hidden + 1] = {key=key,seen=state.last_visible[key] or 0,node=node}
        end
    end
    if resident <= state.resident_limit then return end
    table.sort(hidden, function(left, right)
        if left.seen == right.seen then return left.key < right.key end
        return left.seen < right.seen
    end)
    for _, entry in ipairs(hidden) do
        if resident <= state.resident_limit then break end
        entry.node:destroy()
        state.nodes[entry.key] = nil
        state.last_visible[entry.key] = nil
        resident = resident - 1
    end
end

local function finish_preload(state)
    if not state.job then return end
    local status = render.preload_status(state.job)
    if not status or not status.done then return end
    render.preload_release(state.job)
    state.job = nil
    if status.failed > 0 then
        state.error = tostring(status.errors[1] or "DS1 chunk preload failed")
        return
    end
    local ok, result = pcall(function()
        if state.recipe.world then
            return render.world_tiles(state.recipe.world, state.recipe.palette)
        end
        return render.ds1_chunks(state.recipe.ds1, state.recipe.dt1, state.recipe.palette, state.chunk_size)
    end)
    if not ok then
        state.error = tostring(result)
        return
    end
    state.set = result
end

-- Convert the unobscured screen rectangle back into map-canvas pixels, then
-- inspect only the spatial buckets it crosses. A town may contain thousands of
-- placements, while an 800x600 camera usually touches only a handful of 512px
-- buckets. Draws spanning bucket edges are deduplicated here.
local function nearby_entries(state, left, top, right, bottom)
    if not state.set.draws or not state.set.buckets then return state.set.chunks end
    local size = state.set.bucket_size
    local origin_x, origin_y = canvas_origin(state)
    local map_left = left - state.margin - state.root_x - origin_x
    local map_top = top - state.margin - state.root_y - origin_y
    local map_right = right + state.margin - state.root_x - origin_x
    local map_bottom = bottom + state.margin - state.root_y - origin_y
    local first_column, last_column = math.floor(map_left / size), math.floor(map_right / size)
    local first_row, last_row = math.floor(map_top / size), math.floor(map_bottom / size)
    local result, seen = state.nearby_scratch, state.nearby_seen
    for index = #result, 1, -1 do result[index] = nil end
    for index in pairs(seen) do seen[index] = nil end
    for row = first_row, last_row do
        for column = first_column, last_column do
            local indexes = state.set.buckets[string.format("%d:%d", column, row)]
            if indexes then
                for _, index in ipairs(indexes) do
                    if not seen[index] then
                        seen[index] = true
                        result[#result + 1] = state.set.draws[index]
                    end
                end
            end
        end
    end
    return result
end

local function refresh_nodes(state)
    if not state.set then return end
    state.revision = state.revision + 1
    local left, top, right, bottom = render_viewport(state.viewport_width, state.viewport_height)
    local admitted, pending_count, visible_keys = 0, 0, state.visible_scratch
    for _ in pairs(state.pending) do pending_count = pending_count + 1 end
    for key in pairs(visible_keys) do visible_keys[key] = nil end
    local entries = nearby_entries(state, left, top, right, bottom)
    for _, chunk in ipairs(entries) do
        -- Chunk coordinates begin at the map's top-left. Render children are
        -- centered around their parent in default mode; shared mode instead
        -- translates directly into the scene's common presentation pixels.
        local origin_x, origin_y = canvas_origin(state)
        local local_left = chunk.x + origin_x
        local local_top = chunk.y + origin_y
        local screen_left = state.root_x + local_left
        local screen_top = state.root_y + local_top
        local visible = screen_left + chunk.width >= left - state.margin
            and screen_left <= right + state.margin
            and screen_top + chunk.height >= top - state.margin
            and screen_top <= bottom + state.margin
        local key = chunk.index + 1
        if visible then visible_keys[key] = true end
        if visible and state.nodes[key] then
            state.nodes[key]:set_visible(true)
            state.last_visible[key] = state.revision
        end
        if visible and not state.nodes[key] and state.pending[key] then
            local status = render.preload_status(state.pending[key])
            if status and status.done then
                if status.failed > 0 then
                    render.preload_release(state.pending[key])
                    state.pending[key] = nil
                    state.failed[key] = true
                    state.error = tostring(status.errors[1] or "world chunk preload failed")
                elseif admitted < state.admit_per_frame then
                    -- This is now a cache hit: raster work happened on a bounded
                    -- preload worker, never inside the scene update deadline.
                    render.preload_release(state.pending[key])
                    state.pending[key] = nil
                    admit_chunk(state, chunk, key)
                    admitted = admitted + 1
                end
            end
        elseif visible and not state.nodes[key] and not state.pending[key] and not state.failed[key]
            and admitted < state.admit_per_frame
            and (not state.shared_root or pending_count < state.admit_per_frame) then
            if state.recipe.world then
                -- Decode each unique visible DT1 graphic on a bounded worker;
                -- repeated placements join the same cache flight.
                queue_tile(state, chunk)
                pending_count = pending_count + 1
            else
                admit_chunk(state, chunk, key)
            end
            admitted = admitted + 1
        end
    end
    -- Keep recently visited placements as hidden retained nodes. Panning back
    -- then toggles visibility instead of rebuilding render resources in a
    -- burst. The bounded LRU below still prevents an explored large world from
    -- retaining every placement forever.
    for key, node in pairs(state.nodes) do
        if not visible_keys[key] then
            node:set_visible(false)
        end
    end
    trim_residency(state, visible_keys)
    for key, job in pairs(state.pending) do
        if not visible_keys[key] then
            local status = render.preload_status(job)
            if status and status.done then
                render.preload_release(job)
                state.pending[key] = nil
            end
        end
    end
end

-- Construction queues CPU decoding and immediately returns. Nothing cold is
-- decoded or uploaded on the scene's update thread.
function chunked_map.create(parent, recipe, options)
    options = options or {}
    local state = {
        recipe = assert(recipe, "chunked map recipe is required"),
        root = options.shared_root or render.create("world", parent),
        shared_root = options.shared_root,
        offset_x = options.offset_x or 0,
        offset_y = options.offset_y or 0,
        depth_offset = options.depth_offset or 0,
        depth_scale = options.depth_scale or 1,
        depth_priority = options.depth_priority or 0,
        nodes = {},
        last_visible = {},
        pending = {},
        failed = {},
        nearby_scratch = {},
        nearby_seen = {},
        visible_scratch = {},
        chunk_size = options.chunk_size or 512,
        margin = residency_margin(recipe, options),
        -- Tile placements are much smaller than 512px composed chunks. Admit a
        -- useful screenful quickly; the renderer's byte budget still governs
        -- actual native uploads and prevents one frame from monopolizing them.
        admit_per_frame = options.admit_per_frame or (recipe.world and 32 or 2),
        resident_limit = options.resident_limit or (recipe.world and 512 or 32),
        revision = 0,
        viewport_width = options.viewport_width or 800,
        viewport_height = options.viewport_height or 600,
        canvas_width = options.canvas_width,
        canvas_height = options.canvas_height,
        root_x = 0,
        root_y = 0,
    }
    if not state.shared_root then state.root:set_z(options.z or 0) end
    local request = {
        kind = recipe.world and "world_tiles" or "ds1_chunks",
        path = recipe.ds1,
        tiles = recipe.dt1,
        world = recipe.world,
        palette = recipe.palette,
        chunk_size = state.chunk_size,
    }
    state.job = render.preload({request})
    return state
end

-- Camera and target are already in presentation pixels. This formula has no
-- remembered "first camera" baseline: the same authoritative position always
-- produces the same map placement, including after scene re-entry.
function chunked_map.update(state, camera_x, camera_y, target_x, target_y, world_view)
    finish_preload(state)
    if state.shared_root then
        -- The scene owns the common camera transform. Every adapter culls
        -- against that same transform, without clamping to its own canvas.
        state.root_x, state.root_y = target_x - camera_x, target_y - camera_y
        if not state.set then return false, state.error end
        refresh_nodes(state)
        return true, nil, target_x, target_y
    end
    local width = state.set and state.set.width or state.canvas_width
    local height = state.set and state.set.height or state.canvas_height
    if not width or not height then return false, state.error end
    local wanted_root_x = target_x + width / 2 - camera_x
    local wanted_root_y = target_y + height / 2 - camera_y
    -- Overlay anchoring may move the desired camera by a quarter screen. Near
    -- a finite zone edge that would reveal black even when every nearby tile is
    -- resident. Clamp the map's center so its authored canvas continues to
    -- cover the whole physical viewport behind translucent panel artwork.
    state.root_x = camera_bounds.clamp_center(wanted_root_x, width, state.viewport_width)
    state.root_y = camera_bounds.clamp_center(wanted_root_y, height, state.viewport_height)
    state.root:set_position(state.root_x, state.root_y)
    if not state.set then return false, state.error end
    refresh_nodes(state)
    -- Reverse projection must use the anchor the clamped camera actually drew,
    -- or pointer targets drift away from the visible floor near zone edges.
    local effective_target_x = camera_bounds.anchor_for_center(state.root_x, width, camera_x)
    local effective_target_y = camera_bounds.anchor_for_center(state.root_y, height, camera_y)
    return true, nil, effective_target_x, effective_target_y
end

function chunked_map.destroy(state)
    if not state then return end
    clear_nodes(state)
    -- Active jobs cannot be forgotten by engine.render/v1. The shared native
    -- preloader retains their records; scene teardown does not release them.
    -- Release completed jobs here; unfinished-job cleanup is tracked by #642.
    if state.job then
        local status = render.preload_status(state.job)
        if status and status.done then
            render.preload_release(state.job)
            state.job = nil
        end
    end
    for key, job in pairs(state.pending) do
        local status = render.preload_status(job)
        if status and status.done then
            render.preload_release(job)
            state.pending[key] = nil
        end
    end
    if not state.shared_root and state.root:exists() then state.root:destroy() end
end

return chunked_map
