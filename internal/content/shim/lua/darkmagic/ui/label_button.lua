-- Text-only Diablo menu control with semantic state colors.
--
-- This is a simpler cousin of button.lua. There is no DC6 button background;
-- the text itself is the visual. Interaction still comes from the SAME shared
-- controls.Manager, which is why keyboard, pointer, focus, and accessibility
-- behavior remain consistent with bitmap buttons.

local render = require("dm.render/v1")
local audio = require("dm.audio/v1")
local data = require("dm.data/v1")
local text = require("darkmagic.ui.text")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local label_button = {}

function label_button.create(root, manager, definition, label, options)
    options = options or {}

    -- Semantic style names keep exact font/palette details out of this widget.
    local normal_style = options.normal_style or "label_button_normal"
    local hover_style = options.hover_style or "label_button_hover"
    local pressed_style = options.pressed_style or hover_style
    local disabled_style = options.disabled_style or "disabled"
    local label_node

    local function draw(style)
        -- Headless tests have no render node, but the control still exists.
        if not label_node then
            return
        end

        text.set(label_node, style, label, definition.width, definition.align or "center")

        -- Definition x/y are top-left control coordinates. Retained nodes use
        -- their center, so add half width/height. Optional text_offset nudges the
        -- baseline without changing the control's hit rectangle.
        label_node:set_position(
            definition.x + definition.width / 2,
            definition.y + definition.height / 2 + (definition.text_offset or 0)
        )
    end

    if render.assets_available() then
        label_node = render.create(options.layer or "hud", root)
        -- Draw once immediately so the button never appears blank before update.
        draw(normal_style)
    end

    -- Plain control table consumed by controls.lua.
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
            -- Meaning belongs to the caller. This visual helper only forwards activation.
            if options.on_activate then
                options.on_activate(current)
            end
        end,

        on_state = function(_, state)
            -- controls.lua gives us one semantic state; choose the matching text style.
            if state == "disabled" then
                draw(disabled_style)
            elseif state == "pressed" then
                draw(pressed_style)
            elseif state == "hover" or state == "focused" then
                draw(hover_style)
            else
                draw(normal_style)
            end
        end,
    }

    -- Expose the retained label for composite widgets that need visibility/z control.
    control.visual = label_node
    manager:add(control)
    return control
end

return label_button
