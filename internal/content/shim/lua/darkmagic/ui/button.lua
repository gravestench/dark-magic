-- Authored Diablo II bitmap button composition.
--
-- This module intentionally models recovered presentation behavior rather than
-- any reference engine's widget implementation. Diablo II buttons keep their
-- normal artwork while merely focused/hovered, switch to the down artwork only
-- while pressed, and offset their label -2,+2 while depressed.
local render = require("dm.render/v1")
local audio = require("dm.audio/v1")
local data = require("dm.data/v1")
local text = require("darkmagic.ui.text")
local tooltips = require("darkmagic.ui.tooltip")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))

local button = {}

local function frames(definition, plural, singular)
    if definition[plural] then
        return definition[plural]
    end
    if definition[singular] ~= nil then
        return { definition[singular] }
    end
    return nil
end

function button.create(root, manager, id, definition, label, options)
    options = options or {}
    local layer = options.layer or "hud"
    local normal_style = options.normal_style or definition.normal_style or "button_normal"
    local hover_style = options.hover_style or definition.hover_style or "button_hover"
    local disabled_style = options.disabled_style or definition.disabled_style or "disabled"
    local pressed_dx = options.pressed_dx or definition.pressed_dx or -2
    local pressed_dy = options.pressed_dy or definition.pressed_dy or 2
    local base_text_offset = options.text_offset or definition.text_offset or 0
    local up_frames = assert(frames(definition, "up_frames", "up_frame"), "button up frame is required")
    local down_frames = assert(frames(definition, "down_frames", "down_frame"), "button down frame is required")
    local disabled_frames = frames(definition, "disabled_frames", "disabled_frame")
    assert(#up_frames == #down_frames, "button state frame counts must match")
    if disabled_frames then
        assert(#up_frames == #disabled_frames, "button disabled frame counts must match")
    end

    local draw = function() end
    local draw_label = function() end
    local help

    if render.assets_available() then
        local pieces = {}
        local palette = assert(manifest.palettes[definition.palette], "unknown button palette")
        for index = 1, #up_frames do
            pieces[index] = render.create(layer, root)
        end
        draw = function(selected_frames)
            local left = definition.x
            for index, node in ipairs(pieces) do
                local width, height = node:set_dc6(definition.sheet, palette, 0, assert(selected_frames[index]))
                node:set_position(left + width / 2, definition.y + height / 2)
                left = left + width
            end
        end

        if options.show_label ~= false then
            local label_node = render.create(layer, root)
            draw_label = function(style, dx, dy)
                text.set(label_node, style, label, definition.width, "center")
                label_node:set_position(
                    definition.x + definition.width / 2 + (dx or 0),
                    definition.y + definition.height / 2 + base_text_offset + (dy or 0)
                )
            end
        end
        if options.tooltip then
            help = tooltips.create(
                root,
                options.tooltip,
                definition.x + definition.width / 2,
                definition.y - (options.tooltip_gap or 2),
                { layer = options.tooltip_layer or "modal", style = options.tooltip_style or "tooltip" }
            )
        end
    end

    local control = {
        id = id,
        label = label,
        x = definition.x,
        y = definition.y,
        width = definition.width,
        height = definition.height,
        enabled = options.enabled,
        scope = options.scope or definition.scope,
        on_activate = function(current)
            local sound = options.sound or manifest.sounds.button or manifest.sounds.select
            if sound and audio.exists(sound) then
                audio.play(sound, { bus = "ui" })
            end
            if options.on_activate then
                options.on_activate(current)
            end
        end,
        on_state = function(_, state)
            local pressed = state == "pressed"
            local highlighted = state == "focused" or state == "hover"
            if state == "disabled" and disabled_frames then
                draw(disabled_frames)
            else
                draw(pressed and down_frames or up_frames)
            end
            if state == "disabled" then
                draw_label(disabled_style, 0, 0)
            elseif pressed then
                draw_label(hover_style, pressed_dx, pressed_dy)
            else
                draw_label(highlighted and hover_style or normal_style, 0, 0)
            end
            if help then
                help:set_visible(state == "hover")
            end
        end,
    }
    control.tooltip = help

    manager:add(control)
    -- Manager:add establishes the initial state but intentionally does not call
    -- presentation callbacks. Render it once here so disabled controls and
    -- other non-normal initial states are correct before the first input tick.
    control.on_state(control, control.state)
    return control
end

return button
