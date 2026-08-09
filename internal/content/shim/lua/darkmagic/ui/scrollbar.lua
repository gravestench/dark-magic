-- Authored Diablo II text/layout scrollbar.
--
-- OpenDiablo2 recovered the TextSlid.dc6 frame meanings used by the original
-- text scrollbar: hollow arrows 8/9, filled arrows 10/11, inner gutter 13,
-- and slider fill 14. The values live in darkmagic.ui.compat; this module owns
-- only Dark Magic's independent retained rendering and input behavior.
local render = require("dm.render/v1")
local data = require("dm.data/v1")
local text = require("darkmagic.ui.text")
local compat = require("darkmagic.ui.compat")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local M = {}
local recovered = compat.widgets.text_scrollbar

local defaults = {
    sheet = recovered.sheet,
    palette = recovered.palette,
    part_width = recovered.part_width,
    part_height = recovered.part_height,
    frame_down_hollow = recovered.down_hollow_frame,
    frame_up_hollow = recovered.up_hollow_frame,
    frame_down_filled = recovered.down_filled_frame,
    frame_up_filled = recovered.up_filled_frame,
    frame_gutter = recovered.gutter_frame,
    frame_thumb = recovered.thumb_frame,
}

local function dc6_or_rect(node, sheet, palette, frame, fallback_width, fallback_height, color)
    if render.assets_available() then
        local ok, width, height = pcall(node.set_dc6, node, sheet, palette, 0, frame)
        if ok then return width, height end
    end
    node:fill_rect(fallback_width, fallback_height, color[1], color[2], color[3], color[4] or 255)
    return fallback_width, fallback_height
end

function M.create(root, manager, id, definition, label, options)
    definition = definition or {}
    options = options or {}
    local x = assert(definition.x, "scrollbar x is required")
    local y = assert(definition.y, "scrollbar y is required")
    local height = assert(definition.height, "scrollbar height is required")
    local sheet = definition.sheet or defaults.sheet
    local palette = manifest.palettes[definition.palette or defaults.palette]
    local part_width = definition.width or defaults.part_width
    local part_height = definition.part_height or defaults.part_height
    local gutter_top = y + part_height
    local gutter_height = math.max(part_height, height - part_height * 2)
    local thumb_height = math.max(part_height, math.min(gutter_height, definition.thumb_height or part_height * 2))
    local track_travel = math.max(1, gutter_height - thumb_height)

    local up_node = render.create(options.layer or "hud", root)
    local down_node = render.create(options.layer or "hud", root)
    local gutter_nodes = {}
    local thumb_nodes = {}
    local label_node, value_node

    local _, up_height = dc6_or_rect(up_node, sheet, palette, defaults.frame_up_filled, part_width, part_height, {100, 82, 54, 255})
    up_node:set_position(x + part_width / 2, y + up_height / 2)
    local _, down_height = dc6_or_rect(down_node, sheet, palette, defaults.frame_down_filled, part_width, part_height, {100, 82, 54, 255})
    down_node:set_position(x + part_width / 2, y + height - down_height / 2)

    local gutter_step = math.max(1, part_height - 1)
    local gutter_parts = math.ceil(gutter_height / gutter_step)
    for index = 1, gutter_parts do
        local node = render.create(options.layer or "hud", root)
        dc6_or_rect(node, sheet, palette, defaults.frame_gutter, part_width, part_height, {45, 37, 28, 255})
        node:set_position(x + part_width / 2, gutter_top + part_height / 2 + (index - 1) * gutter_step)
        gutter_nodes[#gutter_nodes + 1] = node
    end

    local thumb_parts = math.max(1, math.ceil(thumb_height / gutter_step))
    for index = 1, thumb_parts do
        local node = render.create(options.layer or "hud", root)
        dc6_or_rect(node, sheet, palette, defaults.frame_thumb, part_width, part_height, {164, 137, 83, 255})
        thumb_nodes[#thumb_nodes + 1] = node
    end

    if label and label ~= "" then
        label_node = render.create(options.layer or "hud", root)
        local label_width = options.label_width or 160
        local _, label_height = text.set(label_node, options.label_style or "font_lab_caption", label, label_width, "left")
        label_node:set_position(x - 8 - label_width / 2, y + label_height / 2)
    end
    if options.show_value then
        value_node = render.create(options.layer or "hud", root)
    end

    local function normalized(control)
        local range = control.max - control.min
        if range <= 0 then return 0 end
        return (control.value - control.min) / range
    end

    local function refresh_thumb(control)
        local top = gutter_top + track_travel * normalized(control)
        for index, node in ipairs(thumb_nodes) do
            node:set_position(x + part_width / 2, top + part_height / 2 + (index - 1) * gutter_step)
        end
        if value_node then
            local formatted = options.format_value and options.format_value(control.value) or tostring(control.value)
            local value_width = options.value_width or 52
            local _, value_height = text.set(value_node, options.value_style or "font_lab_caption", formatted, value_width, "right")
            value_node:set_position(x - 8 - value_width / 2, y + height - value_height / 2)
        end
    end

    local changed = options.on_change
    local control = manager:add_scrollbar({
        id = id,
        label = label or id,
        x = x,
        y = gutter_top,
        width = part_width,
        height = gutter_height,
        min = definition.min,
        max = definition.max,
        step = definition.step,
        value = definition.value,
        orientation = "vertical",
        enabled = options.enabled,
        scope = options.scope or definition.scope,
        pointer_to_value = function(current, _, py)
            local fraction = (py - gutter_top - thumb_height / 2) / track_travel
            fraction = math.max(0, math.min(1, fraction))
            return current.min + (current.max - current.min) * fraction
        end,
        on_change = function(current, value)
            refresh_thumb(current)
            if changed then changed(current, value) end
        end,
        on_state = function(current)
            refresh_thumb(current)
        end,
    })

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

    local function arrow_control(suffix, ay, frame_filled, frame_hollow, node, delta)
        local arrow = manager:add({
            id = id .. "_" .. suffix,
            label = (label or id) .. " " .. suffix,
            role = "button",
            x = x,
            y = ay,
            width = part_width,
            height = part_height,
            scope = options.scope or definition.scope,
            hit_priority = options.hit_priority or 10000,
            on_activate = function()
                control:set_value(control.value + delta * control.step)
            end,
            on_state = function(_, state)
                local frame = state == "pressed" and frame_hollow or frame_filled
                dc6_or_rect(node, sheet, palette, frame, part_width, part_height,
                    state == "pressed" and {190, 163, 105, 255} or {100, 82, 54, 255})
            end,
        })
        return arrow
    end

    control.up = arrow_control("up", y, defaults.frame_up_filled, defaults.frame_up_hollow, up_node, -1)
    control.down = arrow_control("down", y + height - part_height, defaults.frame_down_filled, defaults.frame_down_hollow, down_node, 1)
    control.up_node = up_node
    control.down_node = down_node
    control.gutter_nodes = gutter_nodes
    control.thumb_nodes = thumb_nodes
    control.label_node = label_node
    control.value_node = value_node
    refresh_thumb(control)
    return control
end

M.defaults = defaults
return M
