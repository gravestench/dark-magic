-- Retained UI panel/container helper.
--
-- A panel is NOT another scene. It is just a convenient group of render nodes
-- inside the current scene. Grouping gives a composite widget one place to:
--   * create child nodes;
--   * optionally draw a backdrop;
--   * apply the same clip rectangle;
--   * hide/show the group;
--   * move its root through retained hierarchy behavior.
--
-- This keeps simple UI composition in Lua instead of inventing a second native
-- GUI hierarchy beside Dark Magic's retained renderer.

local render = require("engine.render/v1")
local data = require("engine.data/v1")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2legacy.presentation/v1"))
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
        -- This retained root gives all panel children one common parent/owner.
        root = render.create(layer, parent),
        -- `nodes` tracks presentation pieces that should share visibility/clip.
        nodes = {},
        x = x,
        y = y,
        width = width,
        height = height,
        visible = true,
        -- Only literal true enables clipping; nil/false means no panel clip.
        clip = definition.clip == true,
    }

    if definition.sheet and render.assets_available() then
        -- Prefer authored DC6 backing when the panel definition names one.
        result.background = render.create(layer, result.root)
        local palette = manifest.palettes[definition.palette or "sky"]

        -- A custom/mod panel might reference optional art. `pcall` lets the panel
        -- fall back to a simple rectangle instead of making the entire lab fail.
        local ok = pcall(
            result.background.set_dc6,
            result.background,
            definition.sheet,
            palette,
            0,
            definition.frame or 0
        )

        if not ok then
            result.background:fill_rect(width, height, 20, 16, 12, options.alpha or 235)
        end

        result.background:set_position(x + width / 2, y + height / 2)
    elseif render.assets_available() or options.force_background then
        -- No authored sheet: optionally make a plain diagnostic/custom backdrop.
        result.background = render.create(layer, result.root)
        local color = options.background_color or {20, 16, 12, options.alpha or 235}
        result.background:fill_rect(width, height, color[1], color[2], color[3], color[4] or 255)
        result.background:set_position(x + width / 2, y + height / 2)
    end

    if result.background then result.nodes[#result.nodes + 1] = result.background end

    -- Register an existing child node with this panel's group behavior.
    function result:track(node)
        if node then
            self.nodes[#self.nodes + 1] = node

            if self.clip then
                -- Clip coordinates are logical screen-space facts from this panel.
                node:set_clip(self.x, self.y, self.width, self.height)
            end
        end
        return node
    end

    -- Convenience: create a retained node beneath panel.root and track it in one call.
    function result:node(node_layer)
        return self:track(render.create(node_layer or layer, self.root))
    end

    function result:set_visible(visible)
        self.visible = visible == true
        -- Visibility is applied to every tracked presentation piece. The Lua
        -- objects remain alive so showing the panel again does not rebuild them.
        for _, node in ipairs(self.nodes) do node:set_visible(self.visible) end
    end

    function result:set_z(z)
        -- Root z handles hierarchy order; background receives it explicitly too
        -- because it is the visual backing expected at the panel's base depth.
        self.root:set_z(z)
        if self.background then self.background:set_z(z) end
    end

    return result
end

return M
