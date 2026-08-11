-- Sparse DS1 presentation adapter.
--
-- Think of a DS1 as a very large sheet of paper. Uploading that whole sheet to
-- the GPU is wasteful when the window can see only a small rectangle. This
-- helper keeps the paper's coordinates, but creates retained render nodes only
-- for the little 512x512 squares near the camera.
--
-- This module owns pictures, never gameplay facts. Callers still use
-- dm.world/v1 for collision, objects, projection, and authoritative positions.

local render = require("dm.render/v1")

local chunked_map = {}

local function viewport_for(view, width, height)
    if view == "left" then return 0, 0, width / 2, height end
    if view == "right" then return width / 2, 0, width, height end
    if view == "none" then return nil end
    return 0, 0, width, height
end

local function clear_nodes(state)
    for index, node in pairs(state.nodes) do
        if node:exists() then node:destroy() end
        state.nodes[index] = nil
    end
end

local function finish_preload(state)
    if not state.job then return end
    local status = render.preload_status(state.job)
    if not status or not status.done then return end
    state.job = nil
    if status.failed > 0 then
        state.error = tostring(status.errors[1] or "DS1 chunk preload failed")
        return
    end
    local ok, result = pcall(function()
        return render.ds1_chunks(
            state.recipe.ds1,
            state.recipe.dt1,
            state.recipe.palette,
            state.chunk_size
        )
    end)
    if not ok then
        state.error = tostring(result)
        return
    end
    state.set = result
end

local function refresh_nodes(state, world_view)
    if not state.set or world_view == "none" then return end
    local left, top, right, bottom = viewport_for(
        world_view, state.viewport_width, state.viewport_height
    )
    local admitted = 0
    for _, chunk in ipairs(state.set.chunks) do
        -- Chunk coordinates begin at the map's top-left. Render children are
        -- centered around their parent, so translate once into centered space.
        local local_left = chunk.x - state.set.width / 2
        local local_top = chunk.y - state.set.height / 2
        local screen_left = state.root_x + local_left
        local screen_top = state.root_y + local_top
        local visible = screen_left + chunk.width >= left - state.margin
            and screen_left <= right + state.margin
            and screen_top + chunk.height >= top - state.margin
            and screen_top <= bottom + state.margin
        local key = chunk.index + 1
        if visible and not state.nodes[key] and admitted < state.admit_per_frame then
            local node = render.create("world", state.root)
            node:set_ds1_chunk(
                state.recipe.ds1,
                state.recipe.dt1,
                state.recipe.palette,
                chunk.index,
                state.chunk_size
            )
            node:set_position(
                local_left + chunk.width / 2,
                local_top + chunk.height / 2
            )
            node:set_z(chunk.depth)
            state.nodes[key] = node
            admitted = admitted + 1
        elseif not visible and state.nodes[key] then
            state.nodes[key]:destroy()
            state.nodes[key] = nil
        end
    end
end

-- Construction queues CPU decoding and immediately returns. Nothing cold is
-- decoded or uploaded on the scene's update thread.
function chunked_map.create(parent, recipe, options)
    options = options or {}
    local state = {
        recipe = assert(recipe, "chunked map recipe is required"),
        root = render.create("world", parent),
        nodes = {},
        chunk_size = options.chunk_size or 512,
        margin = options.margin or 96,
        admit_per_frame = options.admit_per_frame or 2,
        viewport_width = options.viewport_width or 800,
        viewport_height = options.viewport_height or 600,
        canvas_width = options.canvas_width,
        canvas_height = options.canvas_height,
        root_x = 0,
        root_y = 0,
    }
    state.root:set_z(options.z or 0)
    state.job = render.preload({{
        kind = "ds1_chunks",
        path = recipe.ds1,
        tiles = recipe.dt1,
        palette = recipe.palette,
        chunk_size = state.chunk_size,
    }})
    return state
end

-- Camera and target are already in presentation pixels. This formula has no
-- remembered "first camera" baseline: the same authoritative position always
-- produces the same map placement, including after scene re-entry.
function chunked_map.update(state, camera_x, camera_y, target_x, target_y, world_view)
    finish_preload(state)
    local width = state.set and state.set.width or state.canvas_width
    local height = state.set and state.set.height or state.canvas_height
    if not width or not height then return false, state.error end
    state.root_x = target_x + width / 2 - camera_x
    state.root_y = target_y + height / 2 - camera_y
    state.root:set_position(state.root_x, state.root_y)
    if not state.set then return false, state.error end
    refresh_nodes(state, world_view or "center")
    return true, nil
end

function chunked_map.destroy(state)
    if not state then return end
    clear_nodes(state)
    if state.root:exists() then state.root:destroy() end
end

return chunked_map
