local data = require("engine.data/v1")
local render = require("engine.render/v1")
local assets = require("ds1editor.ui.assets")
local text = require("ds1editor.ui.text")

local theme = assert(data.load("darkmagic/ds1-editor/ui/theme.json"))
local metrics = assert(theme.composition)
local composition = {}

local panel_names = {
    top_left="panel_top_left", top="panel_top", top_right="panel_top_right",
    left="panel_left", fill="panel_fill", right="panel_right",
    bottom_left="panel_bottom_left", bottom="panel_bottom", bottom_right="panel_bottom_right",
}

local grand_names = {
    top_left="grand_top_left", top="grand_top", top_right="grand_top_right",
    left="grand_left", fill="grand_fill", right="grand_right",
    bottom_left="grand_bottom_left", bottom="grand_bottom", bottom_right="grand_bottom_right",
}

local recess_names = {
    top_left="recess_top_left", top="recess_top", top_right="recess_top_right",
    left="recess_left", fill="recess_fill", right="recess_right",
    bottom_left="recess_bottom_left", bottom="recess_bottom", bottom_right="recess_bottom_right",
}

local function seeded(seed, name)
    local value = seed or 0
    for index = 1, #name do
        value = (value * 33 + string.byte(name, index)) % 2147483647
    end
    return value
end

local function fixed(node, name, variant)
    local path, palette, frame = assets.definition("composition", name, variant)
    node:set_scale(1, 1)
    return node:set_dc6(path, palette, 0, frame)
end

local function tiled(node, name, width, height, seed)
    local path, palette, frames = assets.variants("composition", name)
    node:set_scale(1, 1)
    if #frames > 1 then
        return node:set_dc6_tiled_variants(path, palette, 0, frames, width, height, seeded(seed, name))
    end
    return node:set_dc6_tiled(path, palette, 0, frames[1], width, height)
end

local function methods(control)
    function control:set_visible(visible) self.root:set_visible(visible) end
    function control:exists() return self.root ~= nil and self.root:exists() end
    function control:contains(x, y)
        return x >= self.left and x < self.left + self.width
            and y >= self.top and y < self.top + self.height
    end
    function control:destroy() self.root:destroy() end
    return control
end

local function size(name)
    local _, _, _, width, height = assets.definition("composition", name)
    return width, height
end

-- Build one native-pixel frame. Corners never scale; edge and field sprites
-- repeat into the remaining extent. This is the only structural primitive used
-- by the editor shell, viewport, inspector, status bar, and recessed galleries.
local function frame(parent, options, names, minimum)
    options = options or {}
    local control = methods({
        left=options.left or 0,
        top=options.top or 0,
        width=assert(options.width, "frame width is required"),
        height=assert(options.height, "frame height is required"),
        seed=options.seed or 0,
        fill_enabled=options.fill ~= false,
        nodes={},
    })
    control.root = render.create("hud", parent)
    control.root:set_z(options.z or 0)

    for _, key in ipairs({
        "fill", "top", "bottom", "left", "right",
        "top_left", "top_right", "bottom_left", "bottom_right",
    }) do
        control.nodes[key] = render.create("hud", control.root)
        control.nodes[key]:set_z(key == "fill" and 0 or 1)
    end

    function control:set_bounds(left, top, width, height)
        self.left, self.top = left, top
        self.width, self.height = math.max(minimum, width), math.max(minimum, height)
        self.root:set_position(left + self.width / 2, top + self.height / 2)

        local left_width = select(1, size(names.left))
        local right_width = select(1, size(names.right))
        local _, top_height = size(names.top)
        local _, bottom_height = size(names.bottom)
        local top_left_width, top_left_height = size(names.top_left)
        local top_right_width, top_right_height = size(names.top_right)
        local bottom_left_width, bottom_left_height = size(names.bottom_left)
        local bottom_right_width, bottom_right_height = size(names.bottom_right)

        local center_width = math.max(1, self.width - math.max(top_left_width, bottom_left_width)
            - math.max(top_right_width, bottom_right_width))
        local center_height = math.max(1, self.height - math.max(top_left_height, top_right_height)
            - math.max(bottom_left_height, bottom_right_height))

        local center_left = -self.width / 2 + math.max(top_left_width, bottom_left_width)
        local center_top = -self.height / 2 + math.max(top_left_height, top_right_height)

        self.nodes.fill:set_visible(self.fill_enabled)
        if self.fill_enabled then
            tiled(self.nodes.fill, names.fill, center_width, center_height, self.seed + 5)
            self.nodes.fill:set_position(center_left + center_width / 2, center_top + center_height / 2)
        end

        tiled(self.nodes.top, names.top, center_width, top_height, self.seed + 2)
        self.nodes.top:set_position(center_left + center_width / 2, -self.height / 2 + top_height / 2)
        tiled(self.nodes.bottom, names.bottom, center_width, bottom_height, self.seed + 8)
        self.nodes.bottom:set_position(center_left + center_width / 2, self.height / 2 - bottom_height / 2)
        tiled(self.nodes.left, names.left, left_width, center_height, self.seed + 4)
        self.nodes.left:set_position(-self.width / 2 + left_width / 2, center_top + center_height / 2)
        tiled(self.nodes.right, names.right, right_width, center_height, self.seed + 6)
        self.nodes.right:set_position(self.width / 2 - right_width / 2, center_top + center_height / 2)

        fixed(self.nodes.top_left, names.top_left, self.seed + 1)
        self.nodes.top_left:set_position(-self.width / 2 + top_left_width / 2,
            -self.height / 2 + top_left_height / 2)
        fixed(self.nodes.top_right, names.top_right, self.seed + 3)
        self.nodes.top_right:set_position(self.width / 2 - top_right_width / 2,
            -self.height / 2 + top_right_height / 2)
        fixed(self.nodes.bottom_left, names.bottom_left, self.seed + 7)
        self.nodes.bottom_left:set_position(-self.width / 2 + bottom_left_width / 2,
            self.height / 2 - bottom_left_height / 2)
        fixed(self.nodes.bottom_right, names.bottom_right, self.seed + 9)
        self.nodes.bottom_right:set_position(self.width / 2 - bottom_right_width / 2,
            self.height / 2 - bottom_right_height / 2)
    end

    function control:set_state(state)
        local red, green, blue = 255, 255, 255
        if state == "selected" then red, green, blue = 155, 235, 255 end
        for key, node in pairs(self.nodes) do
            if key ~= "fill" then node:set_tint(red, green, blue) end
        end
    end

    control:set_bounds(control.left, control.top, control.width, control.height)
    return control
end

function composition.frame(parent, options)
    return frame(parent, options, panel_names, 40)
end

function composition.chrome(parent, options)
    options = options or {}
    options.fill = false
    return frame(parent, options, panel_names, 40)
end

function composition.grand_frame(parent, options)
    return frame(parent, options, grand_names, 128)
end

function composition.grand_chrome(parent, options)
    options = options or {}
    options.fill = false
    return frame(parent, options, grand_names, 128)
end

function composition.recess(parent, options)
    return frame(parent, options, recess_names, 32)
end

-- Horizontal bands terminate against the shell rails, so they intentionally
-- omit side rails and corner caps. Using a complete frame here would stack
-- ornaments and produce the heavy, doubled chrome the mockup avoids.
function composition.band(parent, options)
    options = options or {}
    local band = methods({
        left=options.left or 0,
        top=options.top or 0,
        width=assert(options.width, "band width is required"),
        height=assert(options.height, "band height is required"),
        seed=options.seed or 0,
    })
    band.root = render.create("hud", parent)
    band.root:set_z(options.z or 0)
    band.fill = render.create("hud", band.root)
    band.top_edge = render.create("hud", band.root)
    band.bottom_edge = render.create("hud", band.root)
    band.fill:set_z(0)
    band.top_edge:set_z(1)
    band.bottom_edge:set_z(1)

    function band:set_bounds(left, top, width, height)
        self.left, self.top, self.width, self.height = left, top, width, height
        local _, top_height = size("panel_top")
        local _, bottom_height = size("panel_bottom")

        tiled(self.fill, "panel_fill", width, height, self.seed)
        self.fill:set_position(0, 0)
        self.top_edge:set_visible(options.top_edge == true)
        if options.top_edge == true then
            tiled(self.top_edge, "panel_top", width, top_height, self.seed + 1)
            self.top_edge:set_position(0, -height / 2 + top_height / 2)
        end
        self.bottom_edge:set_visible(options.bottom_edge ~= false)
        if options.bottom_edge ~= false then
            tiled(self.bottom_edge, "panel_bottom", width, bottom_height, self.seed + 2)
            self.bottom_edge:set_position(0, height / 2 - bottom_height / 2)
        end
        self.root:set_position(left + width / 2, top + height / 2)
    end

    band:set_bounds(band.left, band.top, band.width, band.height)
    return band
end

-- The mockup's left authoring rail is open into the outer shell. It owns one
-- dark stone field and one inward border, with no duplicate corner caps.
function composition.tool_tray(parent, options)
    options = options or {}
    local tray = methods({
        left=options.left or 0,
        top=options.top or 0,
        width=assert(options.width, "tool tray width is required"),
        height=assert(options.height, "tool tray height is required"),
        seed=options.seed or 0,
    })
    tray.root = render.create("hud", parent)
    tray.root:set_z(options.z or 0)
    tray.fill = render.create("hud", tray.root)
    tray.rail = render.create("hud", tray.root)
    tray.fill:set_z(0); tray.rail:set_z(1)

    function tray:set_bounds(left, top, width, height)
        self.left, self.top, self.width, self.height = left, top, width, height
        local rail_width = select(1, size("panel_right"))
        tiled(self.fill, "panel_fill", width, height, self.seed)
        tiled(self.rail, "panel_right", rail_width, height, self.seed + 1)
        self.fill:set_position(0, 0)
        self.rail:set_position(width / 2 - rail_width / 2, 0)
        self.root:set_position(left + width / 2, top + height / 2)
    end

    tray:set_bounds(tray.left, tray.top, tray.width, tray.height)
    return tray
end

-- Assemble the exact nested regions seen in the approved mockup. The screen
-- owns content inside these regions; this component owns every structural rail
-- so independently authored panels cannot overlap or leave dangling joints.
function composition.workspace(parent, options)
    options = options or {}
    local width, height = assert(options.width), assert(options.height)
    local canvas = assert(options.canvas)
    local inspector_left, inspector_right = assert(options.inspector_left), assert(options.inspector_right)
    local workspace = {
        shell=composition.grand_chrome(parent, {
            left=4, top=4, width=width - 8, height=height - 8, z=30, seed=11,
        }),
        title=composition.band(parent, {
            left=64, top=8, width=width - 128, height=48, z=10, seed=13,
        }),
        commands=composition.band(parent, {
            left=80, top=56, width=width - 88, height=80, z=10, seed=17,
        }),
        viewport=composition.chrome(parent, {
            left=canvas.left - 4, top=canvas.top - 4,
            width=canvas.right - canvas.left + 8,
            height=canvas.bottom - canvas.top + 8,
            z=12, seed=21,
        }),
        tools=composition.tool_tray(parent, {
            left=8, top=canvas.top, width=80, height=canvas.bottom - canvas.top,
            z=10, seed=31,
        }),
        inspector=composition.frame(parent, {
            left=inspector_left, top=canvas.top,
            width=inspector_right - inspector_left,
            height=canvas.bottom - canvas.top,
            z=10, seed=37,
        }),
        status=composition.frame(parent, {
            left=8, top=height - 68, width=width - 16, height=60,
            z=10, seed=41,
        }),
    }
    return workspace
end

local strip_skins = {
    button={idle="button_idle", hover="button_hover", pressed="button_pressed",
        selected="button_hover", disabled="button_idle"},
    tab={idle="tab", hover="tab", pressed="tab", selected="tab", tab="tab"},
    dropdown={idle="dropdown_idle", hover="dropdown_hover", pressed="dropdown_hover",
        selected="dropdown_hover"},
    section={idle="section", selected="section"},
    well={idle="well", hover="well", selected="well"},
}

-- Variable-width controls are three-piece strips. Their caps retain authored
-- dimensions and only the center repeats, so state changes never alter bounds.
function composition.strip(parent, options)
    options = options or {}
    local skin = options.skin or "button"
    local control = methods({
        left=options.left or 0,
        top=options.top or 0,
        width=assert(options.width, "strip width is required"),
        height=0,
        state=options.state or "idle",
        skin=skin,
        seed=options.seed or 0,
        nodes={},
    })
    control.root = render.create("hud", parent)
    control.root:set_z(options.z or 0)
    for index = 1, 3 do
        control.nodes[index] = render.create("hud", control.root)
        control.nodes[index]:set_z(1)
    end

    function control:render_skin()
        local states = assert(strip_skins[self.skin], "unknown strip skin: " .. tostring(self.skin))
        local prefix = states[self.state] or states.idle
        local left_name, center_name, right_name = prefix .. "_left", prefix .. "_center", prefix .. "_right"
        local left_width, authored_height = size(left_name)
        local right_width = select(1, size(right_name))
        self.height = authored_height
        self.width = math.max(self.width, left_width + right_width + 1)
        local center_width = self.width - left_width - right_width

        fixed(self.nodes[1], left_name, self.seed + 1)
        tiled(self.nodes[2], center_name, center_width, authored_height, self.seed + 2)
        fixed(self.nodes[3], right_name, self.seed + 3)
        self.nodes[1]:set_position(-self.width / 2 + left_width / 2, 0)
        self.nodes[2]:set_position((left_width - right_width) / 2, 0)
        self.nodes[3]:set_position(self.width / 2 - right_width / 2, 0)

        local red, green, blue = 255, 255, 255
        if self.skin == "tab" and self.state == "selected" then red, green, blue = 150, 230, 255 end
        if self.state == "disabled" then red, green, blue = 120, 120, 120 end
        for _, node in ipairs(self.nodes) do node:set_tint(red, green, blue) end
    end

    function control:set_bounds(left, top, width)
        self.left, self.top, self.width = left, top, width
        self:render_skin()
        self.root:set_position(left + self.width / 2, top + self.height / 2)
    end

    function control:set_state(state)
        self.state = state or "idle"
        self:render_skin()
    end

    control:set_bounds(control.left, control.top, control.width)
    return control
end

local function measured(options)
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

function composition.button(parent, options)
    options = options or {}
    local content = measured(options)
    local inset_left = options.padding_left or options.padding_x or metrics.button_inset_left
    local inset_right = options.padding_right or options.padding_x or metrics.button_inset_right
    local width = math.max(options.width or 0, metrics.button_min_width,
        content.width + inset_left + inset_right)
    local button = composition.strip(parent, {
        left=options.left or 0, top=options.top or 0, width=width,
        state=options.state or "idle", skin=options.skin or "button",
        z=options.z or 0, seed=options.seed,
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
        local label_x = content_x + content.icon_width + content.gap
            - content.width / 2 + content.text_width / 2
        button.label:set_position(label_x, button.state == "pressed" and 1 or 0)
        text.set(button.label, content.style, content.label, 0, "center")
    end
    function button:set_state(state)
        original_state(self, state)
        position_content()
    end
    function button:set_label(value)
        options.label = tostring(value or "")
        content = measured(options)
        self.content = content
        if not options.width then
            self:set_bounds(self.left, self.top,
                math.max(metrics.button_min_width, content.width + inset_left + inset_right))
        end
        position_content()
    end
    position_content()
    return button
end

function composition.dropdown(parent, options)
    options = options or {}
    options.skin = "dropdown"
    options.padding_right = math.max(options.padding_right or 0, 42)
    return composition.button(parent, options)
end

local icon_states = {
    idle="icon_idle", hover="icon_hover", pressed="icon_pressed",
    selected="icon_selected", disabled="icon_disabled", tab="icon_idle",
}

function composition.icon_button(parent, options)
    options = options or {}
    local state = options.state or "idle"
    local plate_name = assert(icon_states[state])
    local width, height = size(plate_name)
    local control = methods({
        left=options.left or 0, top=options.top or 0,
        width=width, height=height, state=state,
    })
    control.root = render.create("hud", parent)
    control.root:set_position(control.left + width / 2, control.top + height / 2)
    control.root:set_z(options.z or 0)
    control.plate = render.create("hud", control.root)
    control.plate:set_z(1)
    fixed(control.plate, plate_name, options.seed)
    control.icon = assets.create(control.root, assert(options.icon_sheet),
        assert(options.icon_name), 0, 0, 3)

    function control:set_state(next_state)
        self.state = next_state or "idle"
        fixed(self.plate, assert(icon_states[self.state]), options.seed)
        self.icon:set_position(0, self.state == "pressed" and 1 or 0)
    end
    return control
end

function composition.well(parent, options)
    options = options or {}
    return composition.strip(parent, {
        left=options.left, top=options.top, width=options.width,
        skin="well", state=options.state or "idle", z=options.z, seed=options.seed,
    })
end

function composition.section(parent, options)
    options = options or {}
    return composition.strip(parent, {
        left=options.left, top=options.top, width=options.width,
        skin="section", state=options.state or "idle", z=options.z, seed=options.seed,
    })
end

local function binary(parent, options, kind)
    options = options or {}
    local label_value = tostring(options.label or "")
    local style = options.text_style or "editor_small"
    local text_width, text_height = text.measure(style, label_value, 0, "left")
    local size_value = 24
    local gap = label_value ~= "" and (options.content_gap or metrics.content_gap) or 0
    local control = methods({
        left=options.left or 0, top=options.top or 0,
        width=size_value + gap + text_width, height=math.max(size_value, text_height),
        checked=options.checked == true, state=options.state or "idle",
    })
    control.root = render.create("hud", parent)
    control.root:set_position(control.left + control.width / 2, control.top + control.height / 2)
    control.root:set_z(options.z or 0)
    control.plate = render.create("hud", control.root)
    control.plate:set_position(-control.width / 2 + size_value / 2, 0)
    control.plate:set_z(1)
    control.label = render.create("hud", control.root)
    control.label:set_position(-control.width / 2 + size_value + gap, 0)
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

function composition.checkbox(parent, options) return binary(parent, options, "checkbox") end
function composition.radio(parent, options) return binary(parent, options, "radio") end

return composition
