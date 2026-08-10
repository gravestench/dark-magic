-- Authored Diablo II checkbox composition.
local render = require("dm.render/v1")
local data = require("dm.data/v1")
local text = require("darkmagic.ui.text")
local compat = require("darkmagic.ui.compat")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local M = {}

function M.create(root, manager, id, definition, label, options)
    definition = definition or {}
    options = options or {}
    local defaults = compat.widgets.checkbox
    local x = assert(definition.x, "checkbox x is required")
    local y = assert(definition.y, "checkbox y is required")
    local width = definition.width or defaults.width
    local height = definition.height or defaults.height
    local sheet = definition.sheet or defaults.sheet
    local palette_name = definition.palette or options.palette or "fechar"
    local palette = assert(manifest.palettes[palette_name], "unknown checkbox palette")
    local unchecked = definition.unchecked_frame or defaults.unchecked_frame
    local checked = definition.checked_frame or defaults.checked_frame
    local node
    local label_node

    local function draw(current)
        if not node then return end
        local w, h = node:set_dc6(sheet, palette, 0, current.checked and checked or unchecked)
        node:set_position(x + w / 2, y + h / 2)
    end

    if render.assets_available() then
        node = render.create(options.layer or "hud", root)
        if label and label ~= "" then
            label_node = render.create(options.layer or "hud", root)
            local label_width = definition.label_width or 240
            local _, label_height = text.set(label_node, options.label_style or defaults.label_style, label, label_width, "left")
            label_node:set_position(x + width + defaults.label_gap + label_width / 2, y + label_height / 2)
        end
    end

    local changed = options.on_change
    local control = manager:add_checkbox({
        id = id,
        label = label or id,
        x = x,
        y = y,
        width = width,
        height = height,
        checked = definition.checked == true,
        enabled = options.enabled,
        scope = options.scope or definition.scope,
        on_change = function(current, value)
            draw(current)
            if changed then changed(current, value) end
        end,
        on_state = function(current)
            draw(current)
        end,
    })
    draw(control)
    control.node = node
    control.label_node = label_node
    return control
end

return M
