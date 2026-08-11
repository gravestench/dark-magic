-- Pointer-driven inspection for renderer-independent generated zones. This lab
-- stops at the recipe boundary on purpose: DS1 Lab owns expensive asset
-- materialization, while this scene stays responsive during generation work.

local controls = require("darkmagic.ui.controls")
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
    local node = label(self.root, caption, x, 72, 120, normal, "center")
    self.controls:add({
        id=id, label=caption, x=x, y=68, width=120, height=32, focusable=false,
        on_activate=action,
        on_state=function(_, state)
            local style = state == "pressed" and pressed
                or ((state == "hover" or state == "focused") and hovered or normal)
            text.set(node, style, caption, 120, "center")
        end,
    })
end

function lab:generate()
    local mapgen = require("dm.mapgen/v1")
    local ok, zone = pcall(mapgen.maze, self.level_id, self.seed, self.difficulty)
    if not ok then
        text.set(self.status, "font_lab_color", "[red]GENERATION ERROR: [white]" .. tostring(zone), 720, "center")
        return
    end
    self.zone = zone
    text.set(self.status, "font_lab_color", string.format(
        "[white]seed %d  [blue]level %d  [gold]%d rooms / %d links  [green]%s",
        self.seed, self.level_id, #zone.rooms, #zone.links, zone.checksum:sub(1, 16)), 720, "center")
    self:draw_topology(zone)
    local stamp = assert(zone.stamps[1], "maze zone has no chamber stamp")
    local _, recipe_height = text.set(self.recipe, "font_lab_color", string.format(
        "[gold]ROOM 1 RECIPE  [white]preset %d  %s  (%d DT1 files)",
        stamp.preset_def, stamp.ds1, #stamp.dt1), 700, "left")
    self.recipe:set_position(400, 375 + recipe_height / 2)
    local _, trace_height = text.set(self.trace, "font_lab_color", "[gold]GENERATION TRACE\n[white]" .. table.concat(zone.trace, "\n"), 700, "left")
    self.trace:set_position(400, 390 + trace_height / 2)
end

function lab:draw_topology(zone)
    for _, node in ipairs(self.topology_nodes or {}) do
        if node:exists() then node:destroy() end
    end
    self.topology_nodes = {}
    local scale = math.min(640 / math.max(1, zone.width), 190 / math.max(1, zone.height))
    local left, top = 400 - zone.width * scale / 2, 155
    local by_id = {}
    for _, room in ipairs(zone.rooms) do by_id[room.id] = room end
    local function rectangle(width, height, x, y, r, g, b, a)
        local node = render.create("hud", self.root)
        node:fill_rect(math.max(2, width), math.max(2, height), r, g, b, a)
        node:set_position(x, y)
        table.insert(self.topology_nodes, node)
    end
    for _, link in ipairs(zone.links) do
        local a, b = by_id[link.from], by_id[link.to]
        local ax = left + (a.x + a.width / 2) * scale
        local ay = top + (a.y + a.height / 2) * scale
        local bx = left + (b.x + b.width / 2) * scale
        local by = top + (b.y + b.height / 2) * scale
        rectangle(math.abs(bx - ax) + 4, math.abs(by - ay) + 4, (ax + bx) / 2, (ay + by) / 2, 96, 96, 160, 220)
    end
    for _, room in ipairs(zone.rooms) do
        rectangle(room.width * scale - 3, room.height * scale - 3,
            left + (room.x + room.width / 2) * scale,
            top + (room.y + room.height / 2) * scale, 70, 95, 130, 240)
    end
end

function lab:create()
    self.root = render.create("hud")
    local backdrop = render.create("hud", self.root)
    backdrop:fill_rect(760, 560, 0, 0, 0, 192)
    backdrop:set_position(400, 300)
    self.title = label(self.root, "MAP GENERATION LAB", 40, 38, 720, "font_lab_heading", "center")
    self.status = label(self.root, "", 40, 125, 720, "font_lab_color", "center")
    self.recipe = label(self.root, "", 50, 375, 700, "font_lab_color", "left")
    self.trace = label(self.root, "", 50, 420, 700, "font_lab_color", "left")
    label(self.root, "Pointer controls choose the seed. DS1 Lab materializes the emitted recipe.", 40, 565, 720, "font_lab_caption", "center")
    self.level_id, self.difficulty = 9, 0 -- Act I Cave Level 1 maze
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
