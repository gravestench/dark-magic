-- Authored Diablo II bitmap button composition.
--
-- A button definition supplies one or more DC6 frames for its normal and
-- highlighted states. Frames are laid out using their decoded widths rather
-- than magic split coordinates, so both one-piece and segmented controls use
-- the same component. Interaction remains in the shared control manager.
local render = require("dm.render/v1")
local audio = require("dm.audio/v1")
local data = require("dm.data/v1")
local text = require("darkmagic.ui.text")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))

local button = {}

-- Create, render, and register an authored button. Options may override the
-- semantic text styles, render layer, text offset, or activation callback.
function button.create(root, manager, id, definition, label, options)
    options = options or {}
    local layer = options.layer or "hud"
    local normal_style = options.normal_style or definition.normal_style or "button_normal"
    local hover_style = options.hover_style or definition.hover_style or "button_hover"
    local disabled_style = options.disabled_style or definition.disabled_style or "disabled"
    local text_offset = options.text_offset or definition.text_offset or 0
    local draw = function() end
    local draw_label = function() end

    if render.assets_available() then
        local pieces = {}
        local palette = assert(manifest.palettes[definition.palette], "unknown button palette")
        for index = 1, #definition.up_frames do
            pieces[index] = render.create(layer, root)
        end
        draw = function(frames)
            local left = definition.x
            for index, node in ipairs(pieces) do
                local width, height = node:set_dc6(definition.sheet, palette, 0, assert(frames[index]))
                node:set_position(left + width / 2, definition.y + height / 2)
                left = left + width
            end
        end

        local label_node = render.create(layer, root)
        draw_label = function(style)
            text.set(label_node, style, label, definition.width, "center")
            label_node:set_position(
                definition.x + definition.width / 2,
                definition.y + definition.height / 2 + text_offset
            )
        end
        draw(definition.up_frames)
        draw_label(normal_style)
    end

    local control = {
        id = id,
        label = label,
        x = definition.x,
        y = definition.y,
        width = definition.width,
        height = definition.height,
        enabled = options.enabled,
        on_activate = function(current)
            local sound = options.sound or manifest.sounds.select
            if sound and audio.exists(sound) then
                audio.play(sound, { bus = "ui" })
            end
            if options.on_activate then
                options.on_activate(current)
            end
        end,
        on_state = function(_, state)
            local highlighted = state == "focused" or state == "hover"
            draw(highlighted and definition.down_frames or definition.up_frames)
            if state == "disabled" then
                draw_label(disabled_style)
            else
                draw_label(highlighted and hover_style or normal_style)
            end
        end,
    }

    manager:add(control)
    return control
end

return button
