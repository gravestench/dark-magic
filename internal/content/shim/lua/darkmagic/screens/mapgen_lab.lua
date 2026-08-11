-- Pointer-driven inspection for renderer-independent generated zones. This lab
-- stops at the recipe boundary on purpose: DS1 Lab owns expensive asset
-- materialization, while this scene stays responsive during generation work.

local controls = require("darkmagic.ui.controls")
local label_button = require("darkmagic.ui.label_button")
local render = require("dm.render/v1")
local text = require("darkmagic.ui.text")

local lab = {}
local normal = "ui_lab_control_idle"
local hovered = "ui_lab_control_hover"
local pressed = "ui_lab_control_active"

local function label(root, value, x, y, width, style, align)
    local node = render.create("hud", root)
    local _, height = text.set(node, style or "font_lab_color", value, width, align or "left")
    node:set_position(x + width / 2, y + height / 2)
    return node
end

local function add_pointer_button(self, id, x, caption, action)
    local control = label_button.create(self.root, self.controls, {
        id=id, x=x, y=68, width=120, height=32,
    }, caption, {
        normal_style=normal, hover_style=hovered, pressed_style=pressed,
        on_activate=action,
    })
    control.focusable = false
end

function lab:generate()
    local mapgen = require("dm.mapgen/v1")
    local ok, zone = pcall(mapgen.preset, self.level_id, self.seed, self.difficulty)
    if not ok then
        text.set(self.status, "font_lab_color", "[red]GENERATION ERROR: [white]" .. tostring(zone), 720, "center")
        return
    end
    self.zone = zone
    local stamp = assert(zone.stamps[1], "preset zone has no stamp")
    text.set(self.status, "font_lab_color", string.format(
        "[white]seed %d  [blue]level %d  [gold]preset %d  [green]%s",
        self.seed, self.level_id, stamp.preset_def, zone.checksum:sub(1, 16)), 720, "center")
    local _, recipe_height = text.set(self.recipe, "font_lab_color", string.format(
        "[gold]DS1\n[white]%s\n\n[gold]DT1 MASK RESULT (%d files)\n[white]%s",
        stamp.ds1, #stamp.dt1, table.concat(stamp.dt1, "\n")), 700, "left")
    self.recipe:set_position(400, 155 + recipe_height / 2)
    local _, trace_height = text.set(self.trace, "font_lab_color", "[gold]GENERATION TRACE\n[white]" .. table.concat(zone.trace, "\n"), 700, "left")
    self.trace:set_position(400, 390 + trace_height / 2)
end

function lab:create()
    self.root = render.create("hud")
    local backdrop = render.create("hud", self.root)
    backdrop:fill_rect(760, 560, 0, 0, 0, 192)
    backdrop:set_position(400, 300)
    self.title = label(self.root, "MAP GENERATION LAB", 40, 38, 720, "font_lab_heading", "center")
    self.status = label(self.root, "", 40, 125, 720, "font_lab_color", "center")
    self.recipe = label(self.root, "", 50, 155, 700, "font_lab_color", "left")
    self.trace = label(self.root, "", 50, 390, 700, "font_lab_color", "left")
    label(self.root, "Pointer controls choose the seed. DS1 Lab materializes the emitted recipe.", 40, 565, 720, "font_lab_caption", "center")
    self.level_id, self.difficulty = 38, 0 -- compact Act I Tristram preset
    self.seed = require("dm.dev/v1").seed()
    self.controls = controls.new()
    add_pointer_button(self, "previous_seed", 48, "< SEED", function()
        self.seed = math.max(0, self.seed - 1)
        self:generate()
    end)
    add_pointer_button(self, "next_seed", 632, "SEED >", function()
        self.seed = self.seed + 1
        self:generate()
    end)
    self.controls.focus = nil
    self:generate()
end

function lab:update()
    self.controls:update()
end

return lab
