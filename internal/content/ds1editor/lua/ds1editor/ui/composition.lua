local data = require("engine.data/v1")
local render = require("engine.render/v1")
local assets = require("ds1editor.ui.assets")
local text = require("ds1editor.ui.text")

local theme = assert(data.load("darkmagic/ds1-editor/ui/theme.json"))
local metrics = assert(theme.composition)
local composition = {}

local panel_parts = {
    {name="panel_top_left", horizontal="left", vertical="top"},
    {name="panel_top", horizontal="stretch", vertical="top"},
    {name="panel_top_right", horizontal="right", vertical="top"},
    {name="panel_left", horizontal="left", vertical="stretch"},
    {name="panel_fill", horizontal="stretch", vertical="stretch"},
    {name="panel_right", horizontal="right", vertical="stretch"},
    {name="panel_bottom_left", horizontal="left", vertical="bottom"},
    {name="panel_bottom", horizontal="stretch", vertical="bottom"},
    {name="panel_bottom_right", horizontal="right", vertical="bottom"},
}

local grand_parts = {
    {name="grand_top_left", horizontal="left", vertical="top"},
    {name="grand_top", horizontal="stretch", vertical="top"},
    {name="grand_top_right", horizontal="right", vertical="top"},
    {name="grand_left", horizontal="left", vertical="stretch"},
    {name="grand_fill", horizontal="stretch", vertical="stretch"},
    {name="grand_right", horizontal="right", vertical="stretch"},
    {name="grand_bottom_left", horizontal="left", vertical="bottom"},
    {name="grand_bottom", horizontal="stretch", vertical="bottom"},
    {name="grand_bottom_right", horizontal="right", vertical="bottom"},
}

local recess_parts = {
    {name="recess_top_left", horizontal="left", vertical="top"},
    {name="recess_top", horizontal="stretch", vertical="top"},
    {name="recess_top_right", horizontal="right", vertical="top"},
    {name="recess_left", horizontal="left", vertical="stretch"},
    {name="recess_fill", horizontal="stretch", vertical="stretch"},
    {name="recess_right", horizontal="right", vertical="stretch"},
    {name="recess_bottom_left", horizontal="left", vertical="bottom"},
    {name="recess_bottom", horizontal="stretch", vertical="bottom"},
    {name="recess_bottom_right", horizontal="right", vertical="bottom"},
}

local function seeded(seed, name)
    local result = seed or 0
    for index = 1, #name do result = (result * 33 + string.byte(name, index)) % 2147483647 end
    return result
end

local function tiled(node, name, width, height, seed)
    local path, palette, frames = assets.variants("composition", name)
    if #frames > 1 then
        return node:set_dc6_tiled_variants(path, palette, 0, frames, width, height, seeded(seed, name))
    end
    return node:set_dc6_tiled(path, palette, 0, frames[1], width, height)
end

local function fixed(node, name, variant)
    local path, palette, frame = assets.definition("composition", name, variant)
    return node:set_dc6(path, palette, 0, frame)
end

local function axis_geometry(mode, extent, size, leading, trailing)
    if mode == "left" or mode == "top" then return -extent / 2 + size / 2, size end
    if mode == "right" or mode == "bottom" then return extent / 2 - size / 2, size end
    return 0, math.max(1, extent - leading - trailing)
end

local function control_methods(control)
    function control:set_visible(visible) self.root:set_visible(visible) end
    function control:exists() return self.root ~= nil and self.root:exists() end
    function control:contains(x, y)
        return x >= self.left and x < self.left + self.width and y >= self.top and y < self.top + self.height
    end
    function control:destroy() self.root:destroy() end
    return control
end

local function nine_slice(parent, options, definitions, minimum_size)
    options = options or {}
    local frame = control_methods({
        left=options.left or 0,
        top=options.top or 0,
        width=assert(options.width, "composition frame width is required"),
        height=assert(options.height, "composition frame height is required"),
        parts={},
        seed=options.seed or 0,
        fill=options.fill ~= false,
    })
    frame.root = render.create("hud", parent)
    frame.root:set_z(options.z or 0)
    for index, definition in ipairs(definitions) do
        local node = render.create("hud", frame.root)
        node:set_z(index == 5 and 1 or 2)
        frame.parts[index] = {node=node, definition=definition}
    end

    function frame:set_bounds(left, top, width, height)
        self.left, self.top = left, top
        self.width, self.height = math.max(minimum_size, width), math.max(minimum_size, height)
        self.root:set_position(left + self.width / 2, top + self.height / 2)
        for index, part in ipairs(self.parts) do
            local definition = part.definition
            if index == 5 and not self.fill then
                part.node:set_visible(false)
            else
                part.node:set_visible(true)
                local _, _, _, part_width, part_height = assets.definition("composition", definition.name)
                local horizontal_inset = part_width
                local vertical_inset = part_height
                if definition.horizontal == "stretch" then
                    local _, _, _, leading_width = assets.definition("composition", definitions[index - 1].name)
                    local _, _, _, trailing_width = assets.definition("composition", definitions[index + 1].name)
                    horizontal_inset = math.max(leading_width, trailing_width)
                end
                if definition.vertical == "stretch" then
                    local _, _, _, _, leading_height = assets.definition("composition", definitions[index - 3].name)
                    local _, _, _, _, trailing_height = assets.definition("composition", definitions[index + 3].name)
                    vertical_inset = math.max(leading_height, trailing_height)
                end
                local x, width_value = axis_geometry(
                    definition.horizontal, self.width, part_width, horizontal_inset, horizontal_inset
                )
                local y, height_value = axis_geometry(
                    definition.vertical, self.height, part_height, vertical_inset, vertical_inset
                )
                if definition.horizontal == "stretch" or definition.vertical == "stretch" then
                    tiled(part.node, definition.name, width_value, height_value, self.seed + index)
                else
                    fixed(part.node, definition.name, seeded(self.seed + index, definition.name))
                end
                part.node:set_position(x, y)
            end
        end
    end

    function frame:set_state(state)
        local selected = state == "selected"
        for index, part in ipairs(self.parts) do
            if index ~= 5 then
                part.node:set_tint(selected and 150 or 255, selected and 230 or 255, 255)
            end
        end
    end
    frame:set_bounds(frame.left, frame.top, frame.width, frame.height)
    return frame
end

-- Structural panels retain compact ornamental corners and use deterministic
-- authored variants across every long rail and stone field.
function composition.frame(parent, options)
    return nine_slice(parent, options, panel_parts, 40)
end

function composition.chrome(parent, options)
    options = options or {}
    options.fill = false
    return nine_slice(parent, options, panel_parts, 40)
end

function composition.grand_frame(parent, options)
    return nine_slice(parent, options, grand_parts, 128)
end

function composition.grand_chrome(parent, options)
    options = options or {}
    options.fill = false
    return nine_slice(parent, options, grand_parts, 128)
end

-- A tool tray shares the shell's outer edge, so it needs a stone field and a
-- single inner rail—not a second complete frame with dangling top/bottom caps.
function composition.tool_tray(parent, options)
    options = options or {}
    local tray = control_methods({
        left=options.left or 0,
        top=options.top or 0,
        width=assert(options.width, "tool tray width is required"),
        height=assert(options.height, "tool tray height is required"),
        seed=options.seed or 0,
    })
    tray.root = render.create("hud", parent)
    tray.root:set_z(options.z or 0)
    tray.fill = render.create("hud", tray.root)
    tray.fill:set_z(1)
    tray.rail = render.create("hud", tray.root)
    tray.rail:set_z(2)

    function tray:set_bounds(left, top, width, height)
        self.left, self.top, self.width, self.height = left, top, width, height
        local _, _, _, rail_width = assets.definition("composition", "panel_right")
        tiled(self.fill, "panel_fill", width, height, self.seed)
        tiled(self.rail, "panel_right", rail_width, height, self.seed + 1)
        self.fill:set_position(0, 0)
        self.rail:set_position(width / 2 - rail_width / 2, 0)
        self.root:set_position(left + width / 2, top + height / 2)
    end

    tray:set_bounds(tray.left, tray.top, tray.width, tray.height)
    return tray
end

-- Preview cells need a true four-sided recess. A horizontal three-piece strip
-- cannot be made taller without repeating its top and bottom rails as bars.
function composition.recess(parent, options)
    return nine_slice(parent, options, recess_parts, 32)
end

local strip_skins = {
    button={idle="button_idle", hover="button_hover", pressed="button_pressed",
        selected="button_hover", tab="tab"},
    tab={idle="tab", hover="tab", pressed="button_pressed", selected="button_hover", tab="tab"},
    dropdown={idle="dropdown_idle", hover="dropdown_hover", pressed="dropdown_hover",
        selected="dropdown_hover"},
    section={idle="section"},
    well={idle="well"},
}

-- Three-piece strips preserve purpose-built end caps while only the quiet
-- center field repeats. This is used for text buttons, tabs, rows, and headers.
function composition.strip(parent, options)
    options = options or {}
    local skin = options.skin or "button"
    local strip = control_methods({
        left=options.left or 0,
        top=options.top or 0,
        width=assert(options.width, "composition strip width is required"),
        height=options.height,
        state=options.state or "idle",
        skin=skin,
        seed=options.seed or 0,
        nodes={},
    })
    strip.root = render.create("hud", parent)
    strip.root:set_z(options.z or 0)
    for index = 1, 3 do
        strip.nodes[index] = render.create("hud", strip.root)
        strip.nodes[index]:set_z(1)
    end

    function strip:render_skin()
        local states = assert(strip_skins[self.skin], "unknown strip skin: " .. tostring(self.skin))
        local prefix = states[self.state] or states.idle
        local left_name, center_name, right_name = prefix .. "_left", prefix .. "_center", prefix .. "_right"
        local _, _, _, left_width, authored_height = assets.definition("composition", left_name)
        local _, _, _, right_width = assets.definition("composition", right_name)
        self.height = self.height or authored_height
        self.width = math.max(self.width, left_width + right_width + 1)
        local center_width = self.width - left_width - right_width
        if self.height == authored_height then
            fixed(self.nodes[1], left_name)
            fixed(self.nodes[3], right_name)
        else
            tiled(self.nodes[1], left_name, left_width, self.height, self.seed + 1)
            tiled(self.nodes[3], right_name, right_width, self.height, self.seed + 3)
        end
        tiled(self.nodes[2], center_name, center_width, self.height, self.seed)
        self.nodes[1]:set_position(-self.width / 2 + left_width / 2, 0)
        self.nodes[2]:set_position((left_width - right_width) / 2, 0)
        self.nodes[3]:set_position(self.width / 2 - right_width / 2, 0)
    end

    function strip:set_bounds(left, top, width)
        self.left, self.top, self.width = left, top, width
        self:render_skin()
        self.root:set_position(left + self.width / 2, top + self.height / 2)
    end
    function strip:set_state(state)
        self.state = state or "idle"
        self:render_skin()
    end
    strip:set_bounds(strip.left, strip.top, strip.width)
    return strip
end

local function measured_content(options)
    local style = options.text_style or "editor_small"
    local label = tostring(options.label or "")
    local text_width, text_height = text.measure(style, label, 0, "center")
    local icon_width, icon_height = 0, 0
    if options.icon_sheet and options.icon_name then
        local _, _, _, width, height = assets.definition(options.icon_sheet, options.icon_name)
        icon_width, icon_height = width, height
    end
    local gap = label ~= "" and icon_width > 0 and (options.content_gap or metrics.content_gap) or 0
    return {
        style=style, label=label, text_width=text_width, text_height=text_height,
        icon_width=icon_width, icon_height=icon_height, gap=gap,
        width=text_width + icon_width + gap,
        height=math.max(text_height, icon_height),
    }
end

-- Text controls size from exact bitmap bounds plus asymmetric content insets.
-- Icon+label controls use the same measurement path, so neither can collide with caps.
function composition.button(parent, options)
    options = options or {}
    local content = measured_content(options)
    local inset_left = options.padding_left or options.padding_x or metrics.button_inset_left
    local inset_right = options.padding_right or options.padding_x or metrics.button_inset_right
    local inset_top = options.padding_top or options.padding_y or metrics.button_inset_top
    local inset_bottom = options.padding_bottom or options.padding_y or metrics.button_inset_bottom
    local width = math.max(
        options.width or 0,
        metrics.button_min_width,
        content.width + inset_left + inset_right
    )
    local height = math.max(metrics.button_height, content.height + inset_top + inset_bottom)
    local button = composition.strip(parent, {
        left=options.left or 0, top=options.top or 0, width=width, height=height,
        state=options.state or "idle", skin=options.skin or "button", z=options.z or 0,
        seed=options.seed,
    })
    button.content = content
    button.label = render.create("hud", button.root)
    button.label:set_z(4)
    if options.icon_sheet and options.icon_name then
        button.icon = assets.create(button.root, options.icon_sheet, options.icon_name, 0, 0, 4)
    end

    local original_state = button.set_state
    local function position_content()
        local content_x = (inset_left - inset_right) / 2
        if button.icon then
            button.icon:set_position(content_x - content.width / 2 + content.icon_width / 2, 0)
        end
        local label_x = content_x + content.icon_width + content.gap - content.width / 2 + content.text_width / 2
        button.label:set_position(label_x, button.state == "pressed" and 1 or 0)
        text.set(button.label, content.style, content.label, 0, "center")
    end
    function button:set_state(state)
        original_state(self, state)
        position_content()
    end
    function button:set_label(value)
        options.label = tostring(value or "")
        content = measured_content(options)
        self.content = content
        if not options.width then
            width = math.max(metrics.button_min_width, content.width + inset_left + inset_right)
            self:set_bounds(self.left, self.top, width)
        end
        position_content()
    end
    position_content()
    return button
end

-- Dropdowns share text measurement with buttons but reserve a purpose-built
-- arrow cap instead of placing a glyph over a generic skin.
function composition.dropdown(parent, options)
    options = options or {}
    options.skin = "dropdown"
    options.padding_right = math.max(options.padding_right or 0, 42)
    return composition.button(parent, options)
end

local function binary_control(parent, options, kind)
    options = options or {}
    local label_value = tostring(options.label or "")
    local style = options.text_style or "editor_small"
    local text_width, text_height = text.measure(style, label_value, 0, "left")
    local size = 24
    local gap = label_value ~= "" and (options.content_gap or metrics.content_gap) or 0
    local control = control_methods({
        left=options.left or 0,
        top=options.top or 0,
        width=size + gap + text_width,
        height=math.max(size, text_height),
        checked=options.checked == true,
        state=options.state or "idle",
    })
    control.root = render.create("hud", parent)
    control.root:set_position(control.left + control.width / 2, control.top + control.height / 2)
    control.root:set_z(options.z or 0)
    control.plate = render.create("hud", control.root)
    control.plate:set_position(-control.width / 2 + size / 2, 0)
    control.plate:set_z(1)
    control.label = render.create("hud", control.root)
    control.label:set_position(-control.width / 2 + size + gap, 0)
    control.label:set_z(2)
    text.set(control.label, style, label_value, text_width, "left")

    local function frame_name()
        if control.state == "hover" then return kind .. "_hover" end
        return kind .. (control.checked and "_on" or "_off")
    end
    function control:set_state(state)
        self.state = state or "idle"
        fixed(self.plate, frame_name(), options.seed)
    end
    function control:set_checked(checked)
        self.checked = checked == true
        fixed(self.plate, frame_name(), options.seed)
    end
    control:set_state(control.state)
    return control
end

function composition.checkbox(parent, options)
    return binary_control(parent, options, "checkbox")
end

function composition.radio(parent, options)
    return binary_control(parent, options, "radio")
end

local icon_states = {
    idle="icon_idle", hover="icon_hover", pressed="icon_pressed",
    selected="icon_selected", disabled="icon_disabled", tab="icon_idle",
}

-- Fixed icon plates retain an authored 48px silhouette. The 28px icon sits in
-- a measured quiet area with ten native pixels of margin on every side.
function composition.icon_button(parent, options)
    options = options or {}
    local state = options.state or "idle"
    local plate_name = options.compact and "tool_idle" or assert(icon_states[state])
    local _, _, _, width, height = assets.definition("composition", plate_name)
    local control = control_methods({
        left=options.left or 0, top=options.top or 0, width=width, height=height,
        state=state,
    })
    control.root = render.create("hud", parent)
    control.root:set_position(control.left + width / 2, control.top + height / 2)
    control.root:set_z(options.z or 0)
    control.plate = render.create("hud", control.root)
    control.plate:set_z(1)
    fixed(control.plate, plate_name, options.seed)
    control.icon, control.icon_width, control.icon_height = assets.create(
        control.root, assert(options.icon_sheet), assert(options.icon_name), 0, 0, 3
    )
    function control:set_state(next_state)
        self.state = next_state or "idle"
        fixed(self.plate, options.compact and "tool_idle" or assert(icon_states[self.state]), options.seed)
        self.icon:set_position(0, self.state == "pressed" and 1 or 0)
    end
    return control
end

-- Recessed rows and preview fields use dedicated caps and a quieter variable center.
function composition.well(parent, options)
    options = options or {}
    return composition.strip(parent, {
        left=options.left, top=options.top, width=options.width, height=options.height,
        skin="well", state="idle", z=options.z, seed=options.seed,
    })
end

function composition.section(parent, options)
    options = options or {}
    return composition.strip(parent, {
        left=options.left, top=options.top, width=options.width,
        skin="section", state="idle", z=options.z, seed=options.seed,
    })
end

return composition
