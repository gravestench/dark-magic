-- Authored Diablo II checkbox composition.
--
-- Checkbox behavior is split cleanly:
--   controls.lua   toggles `checked` and dispatches on_change
--   checkbox.lua   chooses DC6 art and lays out the label
--
-- This is a useful pattern for mods: interaction semantics and presentation do
-- not have to be welded together.

local render = require("engine.render/v1")
local data = require("engine.data/v1")
local text = require("d2.ui.text")
local compat = require("d2.ui.compat")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2.presentation/v1"))
local M = {}

function M.create(root, manager, id, definition, label, options)
    -- Both arguments are optional extension/configuration tables.
    definition = definition or {}
    options = options or {}

    -- Recovered/default checkbox art facts live in compat, not scattered through widgets.
    local defaults = compat.widgets.checkbox

    local x = assert(definition.x, "checkbox x is required")
    local y = assert(definition.y, "checkbox y is required")
    local width = definition.width or defaults.width
    local height = definition.height or defaults.height
    local sheet = definition.sheet or defaults.sheet

    -- Most frontend checkboxes use the Fechar palette unless overridden.
    local palette_name = definition.palette or options.palette or "fechar"
    local palette = assert(manifest.palettes[palette_name], "unknown checkbox palette")

    local unchecked = definition.unchecked_frame or defaults.unchecked_frame
    local checked = definition.checked_frame or defaults.checked_frame
    local node
    local label_node

    local function draw(current)
        if not node then return end

        -- Compact conditional: choose checked frame when true, otherwise unchecked.
        local w, h = node:set_dc6(sheet, palette, 0, current.checked and checked or unchecked)

        -- Source x/y are top-left; retained node position is image center.
        node:set_position(x + w / 2, y + h / 2)
    end

    if render.assets_available() then
        node = render.create(options.layer or "hud", root)

        if label and label ~= "" then
            label_node = render.create(options.layer or "hud", root)
            local label_width = definition.label_width or 240
            local _, label_height = text.set(
                label_node,
                options.label_style or defaults.label_style,
                label,
                label_width,
                "left"
            )

            -- Place label to the RIGHT of the checkbox with the recovered gap.
            label_node:set_position(
                x + width + defaults.label_gap + label_width / 2,
                y + label_height / 2
            )
        end
    end

    -- Save the caller's semantic change callback before wrapping it with redraw logic.
    local changed = options.on_change

    -- `add_checkbox` supplies generic toggle behavior. We only provide geometry,
    -- initial value, and callbacks for presentation/meaning.
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
            -- Redraw from the new checked state first...
            draw(current)
            -- ...then tell the caller about the semantic value change.
            if changed then changed(current, value) end
        end,

        -- Focus/hover state does not change D2 checkbox art, but state callbacks
        -- still redraw so visual state remains synchronized after manager changes.
        on_state = function(current)
            draw(current)
        end,
    })

    -- Manager registration does not automatically call presentation callbacks.
    draw(control)

    -- Expose child render handles for composite screens that need visibility/z control.
    control.node = node
    control.label_node = label_node
    return control
end

return M
