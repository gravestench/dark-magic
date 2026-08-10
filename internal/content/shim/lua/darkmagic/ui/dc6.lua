-- Higher-level helpers for composing Diablo II DC6 presentation assets.
--
-- DC6 is a legacy sprite format. The engine capability knows how to decode it;
-- THIS Lua file knows how the shim wants to arrange decoded pictures on screen.
-- That distinction is important for modding:
--
--   binary decoding / checked native resources -> engine capability
--   layout / common anchors / composition      -> ordinary Lua helper
--
-- A mod author therefore gets useful high-level behavior without needing to
-- understand the DC6 byte stream or own GPU resources.

local render = require("dm.render/v1")

local M = {}

-- Diablo II's 800x600 front-end backgrounds are stored as a grid of DC6 frames
-- rather than one giant image. `layout.columns` and `layout.rows` contain the
-- width/height of each tile. The final column/row are smaller than 256 pixels.
function M.frontend_background(parent, layer, path, palette, layout)
    -- Keep every node in an array. The caller may want to hold onto the pieces,
    -- and the scene's resource scope will own their native lifetime.
    local nodes = {}

    -- Headless tests may deliberately run without proprietary game assets.
    -- Returning an empty list preserves the same control-flow shape.
    if not render.assets_available() then
        return nodes
    end

    local widths = layout.columns
    local heights = layout.rows
    local frame = layout.first_frame
    local y = 0

    -- Walk rows from top to bottom.
    for row = 1, #heights do
        local x = 0

        -- Then walk each tile in that row from left to right.
        for column = 1, #widths do
            -- Create one retained image node beneath the caller's parent.
            local node = render.create(layer, parent)

            -- Decode/select exactly one DC6 frame. The node now knows its actual
            -- pixel content, while Lua still owns only the checked handle.
            node:set_dc6(path, palette, layout.direction, frame)

            -- Our layout cursor `(x,y)` is the TILE'S TOP-LEFT. Retained nodes
            -- are positioned by CENTER, so add half the tile width/height.
            node:set_position(x + widths[column] / 2, y + heights[row] / 2)

            -- Put the background well behind ordinary foreground UI.
            node:set_z(-100)

            -- Append this piece so the caller can retain the set.
            nodes[#nodes + 1] = node

            -- Advance the top-left cursor and source frame for the next tile.
            x = x + widths[column]
            frame = frame + 1
        end

        -- Start the next row below the one just completed.
        y = y + heights[row]
    end

    return nodes
end

-- DC6 frames can be cropped differently from one another. Their stored offsets
-- tell us where the cropped rectangle belongs relative to a common animation
-- anchor. This helper converts that old anchor-space rule to retained-node
-- center positioning.
function M.anchored_frame(node, path, palette, anchor_x, anchor_y, frame)
    if not render.assets_available() then
        return
    end

    -- The renderer returns both the decoded dimensions and authored offsets.
    local width, height, offset_x, offset_y = node:set_dc6(path, palette, 0, frame)

    -- Start from the COMMON logical anchor, apply the frame's authored top-left
    -- offset, then add half its decoded dimensions because node position=centre.
    node:set_position(
        anchor_x + offset_x + width / 2,
        anchor_y + offset_y + height / 2
    )
end

-- Same idea as anchored_frame, except the node contains a whole animation.
function M.anchored_animation(node, path, palette, anchor_x, anchor_y, frames_per_second, loop, anchor_mode)
    if not render.assets_available() then
        return 0
    end

    -- `set_dc6_animation` normalizes the animation and returns:
    -- frame count, canvas width/height, and its offset from the common anchor.
    local frames, width, height, offset_x, offset_y = node:set_dc6_animation(
        path,
        palette,
        0,
        frames_per_second,
        -- Defaults keep the common call site pleasantly small.
        loop or "loop",
        anchor_mode or "offsets"
    )

    node:set_position(
        anchor_x + offset_x + width / 2,
        anchor_y + offset_y + height / 2
    )

    -- Returning the count is useful for timing one-shot transitions.
    return frames
end

-- Load several independently cropped animations into ONE shared anchor-space
-- rectangle. This is the key to stable multi-layer composites such as a black
-- logo layer plus a flame layer: if each layer chose its own canvas, their
-- centers could shift as frames change.
function M.anchored_composite(nodes, paths, palette, anchor_x, anchor_y, frames_per_second, loop)
    if not render.assets_available() then
        return 0
    end

    -- These four values will become the UNION of every layer's animation bounds.
    local min_x, min_y, max_x, max_y

    for index, path in ipairs(paths) do
        -- `index` is not needed for the bounds calculation, but ipairs also gives
        -- us deterministic array order matching the nodes/paths relationship.
        local x1, y1, x2, y2 = render.dc6_animation_bounds(path, palette, 0, "offsets")

        -- This compact Lua idiom means:
        -- if min_x already exists, keep the smaller value; otherwise use x1.
        min_x = min_x and math.min(min_x, x1) or x1
        min_y = min_y and math.min(min_y, y1) or y1
        max_x = max_x and math.max(max_x, x2) or x2
        max_y = max_y and math.max(max_y, y2) or y2
    end

    -- Track the largest layer frame count as the composite's useful count.
    local count = 0

    for index, node in ipairs(nodes) do
        -- Every layer receives the SAME explicit union rectangle. The renderer
        -- can therefore normalize them into canvases that line up pixel-for-pixel.
        local layer_count = node:set_dc6_animation(
            paths[index],
            palette,
            0,
            frames_per_second,
            loop or "loop",
            "offsets",
            min_x,
            min_y,
            max_x,
            max_y
        )

        -- Because every layer shares the same union box, every layer also gets
        -- exactly the same retained-node center.
        node:set_position(
            anchor_x + min_x + (max_x - min_x) / 2,
            anchor_y + min_y + (max_y - min_y) / 2
        )

        count = math.max(count, layer_count)
    end

    return count
end

-- Managed animation nodes normally advance themselves. A composite often wants
-- ONE master clock instead, so pause every layer first.
function M.pause_animations(nodes)
    for _, node in pairs(nodes) do
        node:animation_pause()
    end
end

-- Seek every layer to the same time value. This is what prevents independently
-- loaded logo/effect layers from slowly drifting out of sync.
function M.synchronize_animations(nodes, position)
    for _, node in pairs(nodes) do
        node:animation_seek(position)
    end
end

return M
