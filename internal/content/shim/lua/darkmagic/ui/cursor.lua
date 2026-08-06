local render = require("dm.render/v1")
local input = require("dm.input/v1")

local M = {}

function M.new(parent, definition, palettes)
    local cursor = { node = nil, width = 0, height = 0, definition = definition }
    if render.assets_available() then
        cursor.node = render.create("cursor", parent)
        cursor.node:set_z(1000)
        cursor.width, cursor.height = cursor.node:set_dc6(definition.sheet,
            palettes[definition.palette], definition.direction, definition.frame)
    end
    function cursor:update()
        if not self.node then return end
        local x, y = input.cursor()
        self.node:set_position(x - definition.hotspot.x + self.width / 2,
            y - definition.hotspot.y + self.height / 2)
    end
    cursor:update()
    return cursor
end

return M
