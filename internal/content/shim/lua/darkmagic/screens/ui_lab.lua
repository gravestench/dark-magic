-- Interactive Diablo II UI/widget laboratory.
--
-- ui_lab exercises the same retained render nodes and Lua control hooks used by
-- shipping screens. It is intentionally not a second GUI implementation.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local controls = require("darkmagic.ui.controls")
local button = require("darkmagic.ui.button")
local checkbox = require("darkmagic.ui.checkbox")
local text_field = require("darkmagic.ui.text_field")
local text = require("darkmagic.ui.text")
local cursor = require("darkmagic.ui.cursor")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local ui_lab = {}

local function copy_at(source, x, y)
    local result = {}
    for key, value in pairs(source) do
        result[key] = value
    end
    result.x = x
    result.y = y
    return result
end

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
end

function ui_lab.create(self)
    self.root = render.create("hud")
    self.background = box(self.root, 0, 0, manifest.resolution.width, manifest.resolution.height, 18, 15, 13, 255)
    put(self.root, "font_lab_heading", "UI LAB", 40, 18, 720, "center")
    put(self.root, "font_lab_caption", "Shipping Lua hooks + recovered Diablo II widget behavior", 40, 62, 720, "center")
    put(self.root, "font_lab_caption", "Arrows/hover move focus   Enter/click activates   Type in text field   Esc returns", 40, 548, 720, "center")

    self.controls = controls.new()

    -- BUTTON + TOOLTIP: same verified WideButtonBlank mapping as main_menu.
    local authored = copy_at(manifest.screens.main_menu.controls.single_player, 70, 120)
    self.primary = button.create(self.root, self.controls, "authored_button", authored, "AUTHORED BUTTON", {
        tooltip = "Up art stays visible on hover/focus; down art is only pressed; label moves -2,+2",
        on_activate = function()
            self.activation:set_visible(true)
        end,
    })

    -- DISABLED BUTTON: verifies focus exclusion and semantic disabled text.
    local disabled = copy_at(manifest.screens.main_menu.controls.credits, 420, 120)
    self.disabled = button.create(self.root, self.controls, "disabled_button", disabled, "DISABLED", { enabled = false })

    put(self.root, "font_lab_caption", "CHECKBOX (clickbox.dc6 frames 0/1)", 70, 205, 300, "left")
    self.checkbox = checkbox.create(self.root, self.controls, "checkbox", {
        x = 70, y = 238, checked = true, palette = "fechar",
    }, "Expansion character")

    put(self.root, "font_lab_caption", "TEXT FIELD (textbox.dc6 + Formal12)", 70, 285, 300, "left")
    self.text_field = text_field.create(self.root, self.controls, "text_field", {
        x = 70, y = 320, kind = "name", value = "DarkMagic", max_length = 15, palette = "fechar",
    }, "Character Name")

    put(self.root, "font_lab_caption", "SCROLLBAR / SLIDER (control-manager primitive)", 70, 395, 330, "left")
    make_scrollbar(self, {
        id = "scrollbar", label = "Value", x = 70, y = 430, width = 272, height = 18,
        min = 0, max = 100, step = 10, value = 40, thumb_width = 20,
    })

    self.activation = put(self.root, "font_lab_caption", "button callback fired", 520, 168, 200, "center")
    self.activation:set_visible(false)

    put(self.root, "font_lab_caption", "Button, checkbox, textbox and scrollbar all register through darkmagic.ui.controls.", 420, 245, 320, "left")
    put(self.root, "font_lab_caption", "Tooltip and bitmap text are rendered through the same shipping Lua helpers.", 420, 305, 320, "left")
    put(self.root, "font_lab_caption", "This scene therefore catches broken render, input, focus and Lua callback hooks together.", 420, 365, 320, "left")

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
