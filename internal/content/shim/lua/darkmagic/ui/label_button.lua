-- Text-only Diablo menu control with semantic state colors.
local render = require("dm.render/v1")
local audio = require("dm.audio/v1")
local data = require("dm.data/v1")
local text = require("darkmagic.ui.text")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local label_button = {}

function label_button.create(root, manager, definition, label, options)
    options = options or {}
    local normal_style = options.normal_style or "label_button_normal"
    local hover_style = options.hover_style or "label_button_hover"
    local disabled_style = options.disabled_style or "disabled"
    local label_node

    local function draw(style)
        if not label_node then
            return
        end
        text.set(label_node, style, label, definition.width, definition.align or "center")
        label_node:set_position(
            definition.x + definition.width / 2,
            definition.y + definition.height / 2 + (definition.text_offset or 0)
        )
    end

    if render.assets_available() then
        label_node = render.create(options.layer or "hud", root)
        draw(normal_style)
    end

    local control = {
        id = assert(definition.id),
        label = label,
        x = definition.x,
        y = definition.y,
        width = definition.width,
        height = definition.height,
        enabled = options.enabled,
        scope = options.scope or definition.scope,
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
            if state == "disabled" then
                draw(disabled_style)
            elseif state == "hover" or state == "focused" or state == "pressed" then
                draw(hover_style)
            else
                draw(normal_style)
            end
        end,
    }
    control.visual = label_node
    manager:add(control)
    return control
end

return label_button
