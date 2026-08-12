-- Monster Lab browses coherent MonStats + MonStats2 presentation recipes.
-- It uses the same adapter as the game world, so a lab success is meaningful.

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local text = require("d2legacy.ui.text")
local composite = require("d2legacy.gameplay.monster_composite")
local fuzzy_picker = require("d2legacy.ui.fuzzy_picker")

local modes = {"NU", "WL", "A1", "A2", "GH", "DT", "DD", "RN"}
local lab = {}

local function label(root, value, y, style)
    local node = render.create("hud", root)
    local _, height = text.set(node, style or "font_lab_caption", value, 760, "center")
    node:set_position(400, y + height / 2)
    return node
end

function lab:create()
    local data = require("d2legacy.data.monster")
    self.root = render.create("hud")
    self.actor = render.create("hud", self.root)
    self.actor:set_position(400, 335); self.actor:set_scale(2, 2)
    self.title = label(self.root, "MONSTER ANIMATION LAB", 18, "font_lab_heading")
    self.status = label(self.root, "", 68)
	self.help = label(self.root, "F: find monster   Left/Right: direction   Up/Down: mode   PgUp/PgDn: browse   Enter: random", 548)
    self.detail = label(self.root, "", 574)
    self.records = data.all(0)
	local choices = {}
	for _, record in ipairs(self.records) do choices[#choices + 1] = record.id end
	self.picker = fuzzy_picker.create(self.root, {title="SELECT MONSTER", items=choices, on_select=function(value)
		for index, record in ipairs(self.records) do
			if record.id == value then self.index, self.direction, self.dirty = index, 0, true; return end
		end
	end})
    self.index, self.mode_index, self.direction, self.random_counter = 1, 1, 0, 0
    self.dirty = true
end

function lab:randomize()
    if #self.records == 0 then return end
    self.random_counter = self.random_counter + 1
    self.index = ((self.random_counter * 1103515245 + 12345) % #self.records) + 1
    self.mode_index = ((self.random_counter - 1) % #modes) + 1
    self.direction = (self.random_counter * 3) % 8
    self.dirty = true
end

function lab:rebuild()
    local record = self.records[self.index]
    if not record then
        self.actor:set_visible(false); text.set(self.status, "font_lab_color", "[red]NO MONSTER RECIPES[/]", 760, "center"); return
    end
    local snapshot = {token=record.token, mode=modes[self.mode_index], weapon_class=record.weapon_class,
        components=record.components, direction=self.direction}
    local ok, recipe = pcall(composite.resolve, snapshot)
    if ok and recipe then
        ok, recipe = pcall(function()
            self.actor:set_cof_animation(recipe.cof, recipe.palette, recipe.direction, recipe.components, recipe.mode == "DT" and "once" or "loop", recipe.rate)
            return recipe
        end)
    end
    self.actor:set_visible(ok and recipe ~= nil)
    if ok and recipe then
        text.set(self.status, "font_lab_color", string.format("[gold]%s[/]  [white]%s[/]  mode [blue]%s[/]  direction [green]%d[/]", record.id, record.token, snapshot.mode, self.direction), 760, "center")
        text.set(self.detail, "font_lab_color", "[white]" .. recipe.cof .. "[/]", 760, "center")
    else
        text.set(self.status, "font_lab_color", "[red]UNRESOLVED[/]  " .. record.id .. " / " .. snapshot.mode, 760, "center")
        text.set(self.detail, "font_lab_color", "[white]" .. tostring(recipe) .. "[/]", 760, "center")
    end
    self.dirty = false
end

function lab:update()
	if self.picker:update() then return end
	if input.pressed("search") then self.picker:show(); return end
    if input.pressed("left") then self.direction=(self.direction+7)%8;self.dirty=true end
    if input.pressed("right") then self.direction=(self.direction+1)%8;self.dirty=true end
    if input.pressed("up") then self.mode_index=((self.mode_index-2)%#modes)+1;self.dirty=true end
    if input.pressed("down") then self.mode_index=(self.mode_index%#modes)+1;self.dirty=true end
    if #self.records > 0 and input.pressed("page_up") then self.index=((self.index-2)%#self.records)+1;self.dirty=true end
    if #self.records > 0 and input.pressed("page_down") then self.index=(self.index%#self.records)+1;self.dirty=true end
    if input.pressed("confirm") then self:randomize() end
    if self.dirty then self:rebuild() end
end

return lab
