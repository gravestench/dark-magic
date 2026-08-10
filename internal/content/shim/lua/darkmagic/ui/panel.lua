-- Retained UI panel/container helper.
--
-- Panels group presentation nodes, own an optional backdrop, and provide one
-- visibility/clip boundary for composite widgets without introducing another
-- scene or GUI hierarchy.
local render = require("dm.render/v1")
local data = require("dm.data/v1")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local M = {}

function M.create(parent, definition, options)
    definition = definition or {}
    options = options or {}
    local layer = options.layer or "hud"
    local x = assert(definition.x, "panel x is required")
    local y = assert(definition.y, "panel y is required")
    local width = assert(definition.width, "panel width is required")
    local height = assert(definition.height, "panel height is required")

    local result = {
        root = render.create(layer, parent),
        nodes = {},
        x = x,
        y = y,
        width = width,
        height = height,
        visible = true,
        clip = definition.clip == true,
    }

    if definition.sheet and render.assets_available() then
        result.background = render.create(layer, result.root)
        local palette = manifest.palettes[definition.palette or "sky"]
        local ok = pcall(result.background.set_dc6, result.background, definition.sheet, palette, 0, definition.frame or 0)
        if not ok then
            result.background:fill_rect(width, height, 20, 16, 12, options.alpha or 235)
        end
        result.background:set_position(x + width / 2, y + height / 2)
    elseif render.assets_available() or options.force_background then
        result.background = render.create(layer, result.root)
        local color = options.background_color or {20, 16, 12, options.alpha or 235}
        result.background:fill_rect(width, height, color[1], color[2], color[3], color[4] or 255)
        result.background:set_position(x + width / 2, y + height / 2)
    end

    if result.background then result.nodes[#result.nodes + 1] = result.background end

    function result:track(node)
        if node then
            self.nodes[#self.nodes + 1] = node
            if self.clip then node:set_clip(self.x, self.y, self.width, self.height) end
        end
        return node
    end

    function result:node(node_layer)
        return self:track(render.create(node_layer or layer, self.root))
    end

    function result:set_visible(visible)
        self.visible = visible == true
        for _, node in ipairs(self.nodes) do node:set_visible(self.visible) end
    end

    function result:set_z(z)
        self.root:set_z(z)
        if self.background then self.background:set_z(z) end
    end

    return result
end

return M
