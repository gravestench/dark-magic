-- Missile Lab browses the typed Missiles.txt presentation catalog.

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local text = require("d2legacy.ui.text")
local adapter = require("d2legacy.gameplay.missile_presentation")
local fuzzy_picker = require("d2legacy.ui.fuzzy_picker")

local lab = {}

local function label(root, value, y, style)
    local node = render.create("hud", root)
    local _, height = text.set(node, style or "font_lab_caption", value, 760, "center")
    node:set_position(400, y + height / 2)
    return node
end

function lab:create()
    local data = require("d2legacy.data.missile")
    self.root = render.create("hud")
    self.actor = render.create("hud", self.root)
    self.actor:set_position(400, 320); self.actor:set_scale(3, 3)
    self.title = label(self.root, "MISSILE ANIMATION LAB", 18, "font_lab_heading")
    self.status = label(self.root, "", 68)
	self.help = label(self.root, "F: find missile   Left/Right: direction   Page Up/Down: browse   Enter: random", 548)
    self.detail = label(self.root, "", 574)
    self.records = data.all()
	local choices = {}
	for _, record in ipairs(self.records) do choices[#choices + 1] = record.id end
	self.picker = fuzzy_picker.create(self.root, {title="SELECT MISSILE", items=choices, on_select=function(value)
		for index, record in ipairs(self.records) do
			if record.id == value then self.index, self.direction, self.dirty = index, 0, true; return end
		end
	end})
    self.index, self.direction, self.random_counter, self.dirty = 1, 0, 0, true
end

function lab:randomize()
    if #self.records == 0 then return end
    self.random_counter = self.random_counter + 1
    self.index = ((self.random_counter * 1664525 + 1013904223) % #self.records) + 1
    local directions = math.max(1, self.records[self.index].directions)
    self.direction = (self.random_counter * 5) % directions
    self.dirty = true
end

function lab:rebuild()
    local record = self.records[self.index]
    if not record then
        self.actor:set_visible(false); text.set(self.status, "font_lab_color", "[red]NO MISSILE RECIPES[/]", 760, "center"); return
    end
    record.logical_direction = self.direction
    record.velocity_x, record.velocity_y = 0, 0
    local recipe = adapter.resolve(record)
    local ok, result = false, "missing DCC"
    if recipe then
        ok, result = pcall(self.actor.set_dcc_animation, self.actor, recipe.path, recipe.palette,
            recipe.direction, recipe.frames_per_second, recipe.loop)
    end
    self.actor:set_visible(ok)
    if ok then
        text.set(self.status, "font_lab_color", string.format("[gold]%s[/]  direction [green]%d[/] / %d  [blue]%d fps[/]", record.id, self.direction, record.directions, recipe.frames_per_second), 760, "center")
        text.set(self.detail, "font_lab_color", string.format("[white]%s[/]  travel %s  hit %s", recipe.path, record.travel_sound, record.hit_sound), 760, "center")
    else
        text.set(self.status, "font_lab_color", "[red]UNRESOLVED[/]  " .. record.id, 760, "center")
        text.set(self.detail, "font_lab_color", "[white]" .. tostring(result) .. "[/]", 760, "center")
    end
    self.dirty = false
end

function lab:update()
	if self.picker:update() then return end
	if input.pressed("search") then self.picker:show(); return end
    local directions = self.records[self.index] and math.max(1, self.records[self.index].directions) or 1
    if input.pressed("left") then self.direction=(self.direction+directions-1)%directions;self.dirty=true end
    if input.pressed("right") then self.direction=(self.direction+1)%directions;self.dirty=true end
    if #self.records > 0 and input.pressed("page_up") then self.index=((self.index-2)%#self.records)+1;self.direction=0;self.dirty=true end
    if #self.records > 0 and input.pressed("page_down") then self.index=(self.index%#self.records)+1;self.direction=0;self.dirty=true end
    if input.pressed("confirm") then self:randomize() end
    if self.dirty then self:rebuild() end
end

return lab
