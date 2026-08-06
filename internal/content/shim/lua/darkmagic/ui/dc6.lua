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

return M
