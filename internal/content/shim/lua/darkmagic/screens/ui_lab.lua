-- Interactive Diablo II UI/widget laboratory.
--
-- ui_lab exercises the same retained render nodes and Lua control hooks used by
-- shipping screens. It is intentionally not a second GUI implementation: each
-- sample is built from darkmagic.ui.controls and darkmagic.ui.button so input,
-- focus, text entry, authored DC6 states, tooltip behavior, and accessibility
-- can be inspected in one deterministic development scene.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local controls = require("darkmagic.ui.controls")
local button = require("darkmagic.ui.button")
local text = require("darkmagic.ui.text")
local cursor = require("darkmagic.ui.cursor")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local lab = manifest.screens.ui_lab

local ui_lab = {}

local function put(root, style, value, left, top, width, alignment)
    local node = render.create("hud", root)
    local _, height = text.set(node, style, value, width, alignment or "left")
    node:set_position(left + width / 2, top + height / 2)
    return node
end

local function box(root, x, y, width, height, r, g, b, a)
    local node = render.create("hud", root)
    node:fill_rect(width, height, r, g, b, a or 255)
    node:set_position(x + width / 2, y + height / 2)
    return node
end

local function make_checkbox(self, definition)
    local outer = box(self.root, definition.x, definition.y, definition.height, definition.height, 104, 88, 57, 255)
    local inner = box(self.root, definition.x + 3, definition.y + 3, definition.height - 6, definition.height - 6, 18, 15, 13, 255)
    local mark = box(self.root, definition.x + 6, definition.y + 6, definition.height - 12, definition.height - 12, 197, 164, 92, 255)
    put(self.root, "font_lab_caption", definition.label, definition.x + definition.height + 10, definition.y + 2, 220, "left")

    local function refresh(control, state)
        mark:set_visible(control.checked)
        local active = state == "hover" or state == "focused" or state == "pressed"
        outer:set_visible(true)
        inner:set_visible(true)
        if active then
            mark:set_visible(true)
        elseif not control.checked then
            mark:set_visible(false)
        end
    end

    local control = self.controls:add_checkbox({
        id = definition.id,
        label = definition.label,
        x = definition.x,
        y = definition.y,
        width = definition.width,
        height = definition.height,
        checked = definition.checked,
        on_change = function(current)
            refresh(current, current.state)
        end,
        on_state = refresh,
    })
    refresh(control, "normal")
end

local function make_text_field(self, definition)
    box(self.root, definition.x, definition.y, definition.width, definition.height, 104, 88, 57, 255)
    box(self.root, definition.x + 2, definition.y + 2, definition.width - 4, definition.height - 4, 12, 10, 9, 255)
    local value = put(self.root, "font_lab_caption", definition.value, definition.x + 8, definition.y + 6, definition.width - 16, "left")
    local control = self.controls:add_text_field({
        id = definition.id,
        label = definition.label,
        x = definition.x,
        y = definition.y,
        width = definition.width,
        height = definition.height,
        value = definition.value,
        max_length = definition.max_length,
        on_change = function(current)
            text.set(value, "font_lab_caption", current.value .. "_", definition.width - 16, "left")
        end,
        on_state = function(current, state)
            local suffix = state == "focused" and "_" or ""
            text.set(value, "font_lab_caption", current.value .. suffix, definition.width - 16, "left")
        end,
    })
    return control
end

local function make_scrollbar(self, definition)
    box(self.root, definition.x, definition.y, definition.width, definition.height, 48, 42, 34, 255)
    local thumb_width = definition.thumb_width or 18
    local thumb = box(self.root, definition.x, definition.y, thumb_width, definition.height, 151, 126, 75, 255)
    local label = put(self.root, "font_lab_caption", "", definition.x, definition.y + definition.height + 8, definition.width, "center")

    local function refresh(control)
        local span = math.max(1, definition.width - thumb_width)
        local range = math.max(1, control.max - control.min)
        local fraction = (control.value - control.min) / range
        thumb:set_position(definition.x + thumb_width / 2 + span * fraction, definition.y + definition.height / 2)
        text.set(label, "font_lab_caption", string.format("value: %.0f", control.value), definition.width, "center")
    end

    local control = self.controls:add_scrollbar({
        id = definition.id,
        label = definition.label,
        x = definition.x,
        y = definition.y,
        width = definition.width,
        height = definition.height,
        min = definition.min,
        max = definition.max,
        step = definition.step,
        value = definition.value,
        on_change = refresh,
        on_state = function(current) refresh(current) end,
    })
    refresh(control)
    return control
end

function ui_lab.create(self)
    self.root = render.create("hud")
    self.background = box(self.root, 0, 0, manifest.resolution.width, manifest.resolution.height, 18, 15, 13, 255)
    put(self.root, "font_lab_heading", "UI LAB", 40, 18, 720, "center")
    put(self.root, "font_lab_caption", "Shipping Lua hooks + recovered Diablo II widget behavior", 40, 62, 720, "center")
    put(self.root, "font_lab_caption", "Arrows/hover move focus   Enter/click activates   Type in text field   Esc returns", 40, 548, 720, "center")

    self.controls = controls.new()

    local authored = lab.authored_button
    self.primary = button.create(self.root, self.controls, "authored_button", authored, "AUTHORED BUTTON", {
        tooltip = "DC6 up/down states; down art only while pressed; label shifts -2,+2",
        on_activate = function(current)
            current.lab_count = (current.lab_count or 0) + 1
            self.activation:set_visible(true)
        end,
    })

    local disabled = lab.disabled_button
    self.disabled = button.create(self.root, self.controls, "disabled_button", disabled, "DISABLED", {
        enabled = false,
        tooltip = "Disabled controls remain inspectable but cannot receive focus",
    })

    put(self.root, "font_lab_caption", "CHECKBOX", 70, 210, 180, "left")
    make_checkbox(self, lab.checkbox)

    put(self.root, "font_lab_caption", "TEXT FIELD", 70, 290, 180, "left")
    make_text_field(self, lab.text_field)

    put(self.root, "font_lab_caption", "SCROLLBAR / SLIDER", 70, 380, 220, "left")
    make_scrollbar(self, lab.scrollbar)

    self.activation = put(self.root, "font_lab_caption", "activated", 520, 168, 200, "center")
    self.activation:set_visible(false)

    put(self.root, "font_lab_caption", "Focus order and hit testing are owned by darkmagic.ui.controls.", 420, 250, 310, "left")
    put(self.root, "font_lab_caption", "The authored button uses the same DC6 component used by main_menu/tcpip screens.", 420, 300, 310, "left")
    put(self.root, "font_lab_caption", "This scene intentionally keeps state and rendering in Lua to verify mod hooks.", 420, 350, 310, "left")

    self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)
end

function ui_lab.update(self)
    self.controls:update()
    self.cursor:update()
    if input.pressed("cancel") then
        scenes.replace("main_menu")
    end
end

return ui_lab
