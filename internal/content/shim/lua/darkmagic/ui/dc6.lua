local render = require("dm.render/v1")

local M = {}

-- Diablo II's 800x600 front-end backgrounds are stored as four columns by
-- three rows. The final column/row are clipped to 32/88 pixels.
function M.frontend_background(parent, layer, path, palette, layout)
    local nodes = {}
    if not render.assets_available() then return nodes end
    local widths = layout.columns
    local heights = layout.rows
    local frame = layout.first_frame
    local y = 0
    for row = 1, #heights do
        local x = 0
        for column = 1, #widths do
            local node = render.create(layer, parent)
            node:set_dc6(path, palette, layout.direction, frame)
            node:set_position(x + widths[column] / 2, y + heights[row] / 2)
            node:set_z(-100)
            nodes[#nodes + 1] = node
            x = x + widths[column]
            frame = frame + 1
        end
        y = y + heights[row]
    end
    return nodes
end

-- DC6 offsets describe placement relative to a common animation anchor.
function M.anchored_frame(node, path, palette, anchor_x, anchor_y, frame)
    if not render.assets_available() then return end
    local width, height, offset_x, offset_y = node:set_dc6(path, palette, 0, frame)
    node:set_position(
        anchor_x + offset_x + width / 2,
        anchor_y - offset_y + height / 2
    )
end

function M.anchored_animation(node, path, palette, anchor_x, anchor_y, frames_per_second, loop, anchor_mode)
    if not render.assets_available() then return 0 end
    local frames, width, height, offset_x, offset_y = node:set_dc6_animation(
        path, palette, 0, frames_per_second, loop or "loop", anchor_mode or "offsets")
    node:set_position(
        anchor_x + offset_x + width / 2,
        anchor_y - offset_y + height / 2
    )
    return frames
end

-- Load independently cropped layers into the same anchor-space canvas. This
-- keeps a composite immobile while preserving each DC6 frame's authored offset.
function M.anchored_composite(nodes, paths, palette, anchor_x, anchor_y, frames_per_second, loop)
    if not render.assets_available() then return 0 end
    local min_x, min_y, max_x, max_y
    for index, path in ipairs(paths) do
        local x1, y1, x2, y2 = render.dc6_animation_bounds(path, palette, 0)
        min_x = min_x and math.min(min_x, x1) or x1
        min_y = min_y and math.min(min_y, y1) or y1
        max_x = max_x and math.max(max_x, x2) or x2
        max_y = max_y and math.max(max_y, y2) or y2
    end
    local count = 0
    for index, node in ipairs(nodes) do
        count = node:set_dc6_animation(paths[index], palette, 0, frames_per_second,
            loop or "loop", "offsets", min_x, min_y, max_x, max_y)
        node:set_position(anchor_x + min_x + (max_x - min_x) / 2,
            anchor_y + min_y + (max_y - min_y) / 2)
    end
    return count
end

function M.pause_animations(nodes)
    for _, node in pairs(nodes) do
        node:animation_pause()
    end
end

function M.synchronize_animations(nodes, position)
    for _, node in pairs(nodes) do
        node:animation_seek(position)
    end
end

return M
