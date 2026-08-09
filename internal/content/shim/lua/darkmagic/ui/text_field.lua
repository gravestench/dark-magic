-- Authored Diablo II text-entry composition.
local render = require("dm.render/v1")
local data = require("dm.data/v1")
local text = require("darkmagic.ui.text")
local compat = require("darkmagic.ui.compat")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local M = {}

local function sheet_for(defaults, kind)
    if kind == "name" then return defaults.name_sheet end
    if kind == "ip" then return defaults.ip_sheet end
    return defaults.generic_sheet
end

function M.create(root, manager, id, definition, label, options)
    definition = definition or {}
    options = options or {}
    local defaults = compat.widgets.text_field
    local x = assert(definition.x, "text field x is required")
    local y = assert(definition.y, "text field y is required")
    local palette_name = definition.palette or options.palette or "fechar"
    local palette = assert(manifest.palettes[palette_name], "unknown text field palette")
    local sheet = definition.sheet or sheet_for(defaults, definition.kind)
    local background
    local value_node
    local label_node
    local width = definition.width
    local height = definition.height

    if render.assets_available() then
        background = render.create(options.layer or "hud", root)
        local decoded_width, decoded_height = background:set_dc6(sheet, palette, 0, definition.frame or 0)
        width = width or decoded_width
        height = height or decoded_height
        background:set_position(x + decoded_width / 2, y + defaults.background_y + decoded_height / 2)

        value_node = render.create(options.layer or "hud", root)
        if label and label ~= "" then
            label_node = render.create(options.layer or "hud", root)
            local label_width = definition.label_width or width
            local _, label_height = text.set(label_node, options.label_style or defaults.label_style, label, label_width, "left")
            label_node:set_position(x + label_width / 2, y + defaults.label_y + label_height / 2)
        end
    end

    width = width or 272
    height = height or 32

    local function draw_value(current)
        if not value_node then return end
        local suffix = current.state == "focused" and "_" or ""
        local value_width = math.max(1, width - defaults.text_x - 4)
        local _, text_height = text.set(value_node, options.text_style or defaults.text_style, current.value .. suffix, value_width, "left")
        value_node:set_position(x + defaults.text_x + value_width / 2, y + defaults.text_y + text_height / 2)
    end

    local changed = options.on_change
    local control = manager:add_text_field({
        id = id,
        label = label or id,
        x = x,
        y = y,
        width = width,
        height = height,
        value = definition.value or "",
        max_length = definition.max_length,
        filter = definition.filter,
        enabled = options.enabled,
        scope = options.scope or definition.scope,
        on_change = function(current, value)
            draw_value(current)
            if changed then changed(current, value) end
        end,
        on_state = function(current)
            draw_value(current)
        end,
    })
    draw_value(control)
    control.background_node = background
    control.value_node = value_node
    control.label_node = label_node
    return control
end

return M
