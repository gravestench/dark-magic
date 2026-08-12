-- Reusable retained-mode value slider.
--
-- controls.lua owns RANGE INTERACTION: focus, arrow-key steps, pointer capture,
-- clamping, and change callbacks. This helper owns only the VISUAL GEOMETRY:
-- track, active fill, thumb, optional label/value, and pointer-to-value mapping.
--
-- The same interaction can therefore be drawn with real Diablo II DC6 art or
-- with simple colored rectangles in diagnostics/dev tools.

local render = require("engine.render/v1")
local text = require("d2legacy.ui.text")

local M = {}

local function fill(node, width, height, color)
    -- Lua arrays are 1-based, so RGBA lives at color[1]..color[4].
    -- Missing alpha defaults to fully opaque.
    node:fill_rect(width, height, color[1], color[2], color[3], color[4] or 255)
end

function M.create(root, manager, id, definition, label, options)
    definition = definition or {}
    options = options or {}

    local x = assert(definition.x, "slider x is required")
    local y = assert(definition.y, "slider y is required")
    local width = assert(definition.width, "slider width is required")
    local height = assert(definition.height, "slider height is required")
    local orientation = definition.orientation or "horizontal"

    -- Default thumb thickness matches the short axis of the control.
    local thumb_size = definition.thumb_size or (orientation == "horizontal" and height or width)
    local track_thickness = definition.track_thickness or 4

    -- Fallback rectangle colors are intentionally ordinary Lua data.
    local track_color = options.track_color or {58, 49, 38, 255}
    local fill_color = options.fill_color or {139, 111, 66, 255}
    local thumb_color = options.thumb_color or {190, 163, 105, 255}
    local pressed_color = options.pressed_color or {225, 205, 150, 255}

    local track, active, thumb, label_node, value_node

    -- "Authored" means the caller provided real track/thumb DC6 sheets AND
    -- assets are mounted. Otherwise use the diagnostic rectangle presentation.
    local authored = render.assets_available() and options.track_sheet and options.thumb_sheet

    if render.assets_available() then
        track = render.create(options.layer or "hud", root)
        thumb = render.create(options.layer or "hud", root)

        if authored then
            -- Some D2 slider tracks are split across DC6 frames. set_dc6_strip
            -- combines that authored strip rather than pretending one frame is enough.
            track:set_dc6_strip(options.track_sheet, assert(options.palette), 0)
            thumb:set_dc6(options.thumb_sheet, assert(options.palette), 0, options.thumb_frame or 0)
            track:set_position(x + width / 2, y + height / 2)
        else
            -- Diagnostic version has a separate "filled/active" portion of track.
            active = render.create(options.layer or "hud", root)
        end

        if not authored and orientation == "horizontal" then
            fill(track, width, track_thickness, track_color)
            track:set_position(x + width / 2, y + height / 2)
        elseif not authored then
            fill(track, track_thickness, height, track_color)
            track:set_position(x + width / 2, y + height / 2)
        end

        if options.show_label ~= false and label and label ~= "" then
            label_node = render.create(options.layer or "hud", root)
            local label_width = options.label_width or width
            local _, label_height = text.set(
                label_node,
                options.label_style or "font_lab_caption",
                label,
                label_width,
                "left"
            )
            -- Put the label just above the track.
            label_node:set_position(x + label_width / 2, y - 6 - label_height / 2)
        end

        if options.show_value ~= false then
            value_node = render.create(options.layer or "hud", root)
        end
    end

    local function normalized(control)
        local range = control.max - control.min
        if range <= 0 then return 0 end
        -- Convert arbitrary [min,max] value into a simple 0..1 fraction.
        return (control.value - control.min) / range
    end

    local function refresh(control)
        -- Headless logic still has a control but no presentation nodes.
        if not thumb then return end

        local fraction = normalized(control)
        local pressed = control.state == "pressed"
        local color = pressed and pressed_color or thumb_color

        if orientation == "horizontal" then
            -- The THUMB CENTER cannot travel the entire width or half the thumb
            -- would hang outside each end. Subtract thumb_size to get its travel.
            local travel = math.max(0, width - thumb_size)
            local center = x + thumb_size / 2 + travel * fraction

            if active then
                -- Diagnostic filled portion reaches from left edge to thumb center.
                local active_width = math.max(1, center - x)
                fill(active, active_width, track_thickness, fill_color)
                active:set_position(x + active_width / 2, y + height / 2)
            end

            if not authored then fill(thumb, thumb_size, height, color) end
            thumb:set_position(center, y + height / 2)
        else
            local travel = math.max(0, height - thumb_size)
            local center = y + thumb_size / 2 + travel * fraction
            local active_height = math.max(1, center - y)

            if active then
                fill(active, track_thickness, active_height, fill_color)
                active:set_position(x + width / 2, y + active_height / 2)
            end

            if not authored then fill(thumb, width, thumb_size, color) end
            thumb:set_position(x + width / 2, center)
        end

        if value_node then
            -- A caller may supply friendly formatting (percent/dB/etc.).
            local value = options.format_value and options.format_value(control.value) or tostring(control.value)
            local value_width = options.value_width or 64
            text.set(value_node, options.value_style or "font_lab_caption", value, value_width, "right")
            value_node:set_position(x + width + value_width / 2 + 8, y + height / 2)
        end
    end

    -- Save semantic callback while we insert our presentation refresh.
    local changed = options.on_change

    local control = manager:add_slider({
        id = id,
        label = label or id,
        x = x,
        y = y,
        width = width,
        height = height,
        min = definition.min,
        max = definition.max,
        step = definition.step,
        value = definition.value,
        orientation = orientation,
        enabled = options.enabled,
        scope = options.scope or definition.scope,

        -- Tell the generic manager how this slider's POINTER geometry maps back
        -- to a numeric value. The manager still owns clamping and step snapping.
        pointer_to_value = function(current, px, py)
            local fraction

            if orientation == "horizontal" then
                local travel = math.max(1, width - thumb_size)
                fraction = (px - x - thumb_size / 2) / travel
            else
                local travel = math.max(1, height - thumb_size)
                fraction = (py - y - thumb_size / 2) / travel
            end

            fraction = math.max(0, math.min(1, fraction))
            return current.min + (current.max - current.min) * fraction
        end,

        on_change = function(current, value)
            refresh(current)
            if changed then changed(current, value) end
        end,

        -- Pressed state changes thumb appearance in diagnostic mode.
        on_state = function(current)
            refresh(current)
        end,
    })

    -- Programmatic setter for screens that want to change the slider value
    -- without synthesizing input. It mirrors the manager's clamp/snap behavior.
    function control:set_value(value)
        value = math.max(self.min, math.min(self.max, value))

        if self.step and self.step > 0 then
            value = self.min + math.floor(((value - self.min) / self.step) + 0.5) * self.step
            value = math.min(self.max, value)
        end

        if value ~= self.value then
            self.value = value
            if self.on_change then self.on_change(self, value) end
        end
    end

    -- Expose presentation pieces so larger composite menus can hide/show them.
    control.track_node = track
    control.active_node = active
    control.thumb_node = thumb
    control.label_node = label_node
    control.value_node = value_node

    refresh(control)
    return control
end

return M
