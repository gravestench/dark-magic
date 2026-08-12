-- Authored Diablo II bitmap button composition.
--
-- This module sits BETWEEN `controls.lua` and the renderer:
--
--   controls.lua decides: normal / hover / focused / pressed / disabled
--   button.lua decides:   which DC6 frames, text position, blend, sound, tooltip
--
-- There is no native Button object hiding underneath. The actual interactive
-- control is a plain Lua table registered with the shared Manager.
--
-- Historical note: this code models OBSERVED Diablo II presentation behavior;
-- it does not copy a reference engine's widget implementation.

local render = require("engine.render/v1")
local audio = require("engine.audio/v1")
local data = require("engine.data/v1")
local text = require("d2.ui.text")
local tooltips = require("d2.ui.tooltip")
local compat = require("d2.ui.compat")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2.presentation/v1"))

local button = {}

-- Button art sometimes stores one visual state in ONE frame and sometimes in
-- several side-by-side frames. This helper normalizes both manifest shapes into
-- an array so the rest of the function can use one code path.
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
    -- Optional configuration is always turned into a table first so every later
    -- `options.foo` read is safe.
    options = options or {}

    local layer = options.layer or "hud"
    local z = options.z or definition.z or 0

    -- Diablo II does not recolor button text for hover/pressed state the way a
    -- modern web UI often would. It mostly changes the BUTTON ART and nudges the
    -- label while depressed. These style defaults preserve that behavior.
    local normal_style = options.normal_style or definition.normal_style or "button_hover"
    local hover_style = options.hover_style or definition.hover_style or normal_style
    local disabled_style = options.disabled_style or definition.disabled_style or normal_style

    -- Compatibility data stores old D2 draw-mode numbers. `compat.draw_mode`
    -- translates that legacy number to a renderer blend name.
    local text_blend = options.text_blend or definition.text_blend
        or compat.draw_mode(compat.widgets.button.text_draw_mode)

    -- Label motion while depressed is tiny but visible. A caller can override
    -- it, then manifest data, then recovered compatibility defaults win in turn.
    local pressed_dx = options.pressed_dx or definition.pressed_dx or compat.widgets.button.pressed_dx
    local pressed_dy = options.pressed_dy or definition.pressed_dy or compat.widgets.button.pressed_dy
    local base_text_offset = options.text_offset or definition.text_offset or 0

    -- Normalize every state into frame arrays.
    local up_frames = assert(frames(definition, "up_frames", "up_frame"), "button up frame is required")
    local down_frames = assert(frames(definition, "down_frames", "down_frame"), "button down frame is required")
    local disabled_frames = frames(definition, "disabled_frames", "disabled_frame")

    -- A multi-piece button needs exactly one down/disabled piece for every up
    -- piece, otherwise pieces would be missing or misaligned when state changes.
    assert(#up_frames == #down_frames, "button state frame counts must match")
    if disabled_frames then
        assert(#up_frames == #disabled_frames, "button disabled frame counts must match")
    end

    -- Manifest dimensions are a headless fallback. When real art is available,
    -- we replace them below with the ACTUAL decoded dimensions.
    local button_width = assert(definition.width, "button width is required")
    local button_height = assert(definition.height, "button height is required")

    -- These start as no-op functions so the interaction logic below works in a
    -- headless test even when no proprietary art is mounted.
    local draw = function() end
    local draw_label = function() end
    local help

    if render.assets_available() then
        local pieces = {}
        local palette = assert(manifest.palettes[definition.palette], "unknown button palette")

        -- Allocate one retained render node per side-by-side frame piece.
        for index = 1, #up_frames do
            pieces[index] = render.create(layer, root)
            pieces[index]:set_z(z)
        end

        draw = function(selected_frames)
            -- `left` is a top-left placement cursor walking across each piece.
            local left = definition.x
            local decoded_width = 0
            local decoded_height = 0

            for index, node in ipairs(pieces) do
                local width, height = node:set_dc6(
                    definition.sheet,
                    palette,
                    0,
                    assert(selected_frames[index])
                )

                -- Piece art is positioned from its top-left source placement to
                -- a retained-node center by adding half decoded width/height.
                node:set_position(left + width / 2, definition.y + height / 2)

                left = left + width
                decoded_width = decoded_width + width
                decoded_height = math.max(decoded_height, height)
            end

            -- Interaction geometry follows the decoded AUTHORED art rather than
            -- a second guessed rectangle that could slowly drift out of sync.
            button_width = decoded_width > 0 and decoded_width or button_width
            button_height = decoded_height > 0 and decoded_height or button_height
        end

        -- Decode normal art BEFORE registering the control so hitbox, label
        -- centering, and tooltip anchor all use real dimensions from the start.
        draw(up_frames)

        if options.show_label ~= false then
            local label_node = render.create(layer, root)
            label_node:set_z(z + 1)
            label_node:set_blend(text_blend)

            draw_label = function(style, dx, dy)
                text.set(label_node, style, label, button_width, "center")
                label_node:set_position(
                    definition.x + button_width / 2 + (dx or 0),
                    definition.y + button_height / 2 + base_text_offset + (dy or 0)
                )
            end
        end

        if options.tooltip then
            help = tooltips.create(
                root,
                options.tooltip,
                definition.x + button_width / 2,
                definition.y - (options.tooltip_gap or 2),
                {
                    layer = options.tooltip_layer or "modal",
                    style = options.tooltip_style or "tooltip",
                }
            )
        end
    end

    -- This is the ACTUAL interactive control consumed by controls.lua.
    local control = {
        id = id,
        label = label,
        x = definition.x,
        y = definition.y,
        width = button_width,
        height = button_height,
        enabled = options.enabled,
        scope = options.scope or definition.scope,

        on_activate = function(current)
            -- Choose sound from the most specific override down through defaults.
            local sound = options.sound or definition.sound
                or compat.widgets.button.click_sound or manifest.sounds.button or manifest.sounds.select

            if sound and audio.exists(sound) then
                -- Scene replacement can tear down scene-scoped audio immediately.
                -- A click should still finish audibly after it navigates away, so
                -- use a persistent named group. A later click replaces the same
                -- group instead of allowing thousands of old one-shots to pile up.
                audio.play_persistent(sound, {
                    bus = "ui",
                    group = options.sound_group or definition.sound_group or "ui_button_click",
                })
            end

            -- The visual helper does not know what the button MEANS. Delegate the
            -- semantic action back to the caller that created it.
            if options.on_activate then
                options.on_activate(current)
            end
        end,

        on_state = function(_, state)
            local pressed = state == "pressed"
            local highlighted = state == "focused" or state == "hover"

            -- Disabled art is optional. Without it, the normal/up image remains.
            if state == "disabled" and disabled_frames then
                draw(disabled_frames)
            else
                -- Compact conditional: pressed -> down frames, otherwise up.
                draw(pressed and down_frames or up_frames)
            end

            if state == "disabled" then
                draw_label(disabled_style, 0, 0)
            elseif pressed then
                -- Depressed label uses hover treatment plus original D2 pixel nudge.
                draw_label(hover_style, pressed_dx, pressed_dy)
            else
                draw_label(highlighted and hover_style or normal_style, 0, 0)
            end

            if help then
                -- Tooltip is intentionally pointer-hover only here.
                help:set_visible(state == "hover")
            end
        end,
    }

    -- Expose the helper-created tooltip for screens that need to inspect it.
    control.tooltip = help

    manager:add(control)

    -- Manager:add establishes the initial STATE string but intentionally does
    -- not invoke presentation callbacks. Draw once now so disabled/non-normal
    -- initial controls look right before the first update tick.
    control.on_state(control, control.state)

    return control
end

return button
