-- Composite Lab exercises the real legacy adapter and retained renderer.
-- Nothing here knows how to decode or draw a COF/DCC. It only changes a recipe,
-- asks the shared player-composite module to resolve it, and reports failures.

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local text = require("d2legacy.ui.text")
local composite = require("d2legacy.gameplay.player_composite")
local fuzzy_picker = require("d2legacy.ui.fuzzy_picker")

local tokens = {"AM", "SO", "NE", "PA", "BA", "DZ", "AI"}
local modes = {"NU", "WL", "RN"}

-- Curated recipes are intentionally boring and known-coherent. "Random" must
-- be useful for regression hunting, not produce arbitrary filenames that never
-- existed in Diablo II. More verified recipes can be added here as coverage grows.
local recipes = {
    {weapon="HTH", appearance={}},
    {token="AM", weapon="1HS", appearance={RH="SSD"}},
}

local function label(root, value, y, style)
    local node = render.create("hud", root)
    local _, height = text.set(node, style or "font_lab_caption", value, 760, "center")
    node:set_position(400, y + height / 2)
    return node
end

local function index_of(values, wanted)
    for index, value in ipairs(values) do if value == wanted then return index end end
    return 1
end

local lab = {}

function lab:choose_random()
    -- A deterministic counter makes captures reproducible while still letting a
    -- human cycle through varied, coherent combinations with Enter.
    self.random_counter = (self.random_counter or 0) + 1
    local recipe = recipes[((self.random_counter - 1) % #recipes) + 1]
    self.token_index = recipe.token and index_of(tokens, recipe.token) or ((self.random_counter - 1) % #tokens) + 1
    self.mode_index = ((self.random_counter - 1) % #modes) + 1
    self.weapon = recipe.weapon
    self.appearance = {}
    for key, value in pairs(recipe.appearance) do self.appearance[key] = value end
    self.dirty = true
end

function lab:create()
    -- Load the development capability only when this lab is actually opened.
    -- Shipping/test boots may register the scene without granting dev options.
    self.root = render.create("hud")
    self.actor = render.create("hud", self.root)
    self.actor:set_position(400, 345)
    self.actor:set_scale(2, 2)
    self.title = label(self.root, "COMPOSITE ANIMATION LAB", 18, "font_lab_heading")
    self.status = label(self.root, "", 68)
	self.help = label(self.root, "F: find recipe   Left/Right: direction   Up/Down: mode   PgUp/PgDn: class   Enter: random", 548)
    self.detail = label(self.root, "", 574)

	self.token_index = index_of(tokens, "AM")
	self.mode_index = index_of(modes, "NU")
	self.weapon = "HTH"
    -- Lab-facing direction is deliberately a plain spatial index. Asset-facing
    -- direction codes stay hidden behind the lookup below.
	self.direction = 0
	self.appearance = {}
	self.frame = -1
    self.playing, self.dirty = self.frame < 0, true
    self.playback_seconds = 0
	local choices = {}
	for _, token in ipairs(tokens) do
		for _, mode in ipairs(modes) do choices[#choices + 1] = token .. " " .. mode .. " HTH" end
	end
	choices[#choices + 1] = "AM WL 1HS"
	self.picker = fuzzy_picker.create(self.root, {title="SELECT COMPOSITE RECIPE", items=choices, on_select=function(value)
		local token, mode, weapon = value:match("^(%S+)%s+(%S+)%s+(%S+)$")
		self.token_index, self.mode_index, self.weapon = index_of(tokens, token), index_of(modes, mode), weapon
		self.appearance = weapon == "1HS" and {RH="SSD"} or {}
		self.dirty = true
	end})
end

function lab:rebuild()
    local authority = {token=tokens[self.token_index], mode=modes[self.mode_index], weapon_class=self.weapon,
        direction=self.direction, direction_space="logical", palette="data/global/palette/ACT1/pal.dat"}
    local ok, resolved = pcall(composite.recipe, authority, self.appearance, self.weapon)
    if ok then
        ok, resolved = pcall(function()
            self.actor:set_cof_animation(resolved.cof, resolved.palette, resolved.direction,
                resolved.components, "loop", resolved.rate, self.playback_seconds)
            return resolved
        end)
    end
    if ok then
        self.actor:set_visible(true)
        self.resolved = resolved
        if not self.playing then
            self.frame = math.max(0, math.min(resolved.frames - 1, self.frame))
            self.playback_seconds = self.frame * 256 / (resolved.rate * 25)
            self.actor:animation_pause()
            self.actor:animation_seek(self.playback_seconds)
        end
        text.set(self.status, "font_lab_color", string.format("[white]%s  %s  %s  logical/COF %d  DCC %d  rate %d  frames %d%s",
            authority.token, authority.mode, self.weapon, self.direction, resolved.dcc_direction, resolved.rate, resolved.frames,
            self.playing and "" or ("  showing " .. self.frame)), 760, "center")
        text.set(self.detail, "font_lab_color", "[white]" .. resolved.cof, 760, "center")
    else
        self.actor:set_visible(false)
        text.set(self.status, "font_lab_color", "[red]UNRESOLVED: " .. tostring(resolved), 760, "center")
    end
    self.dirty = false
end

function lab:update(elapsed)
	if self.picker:update() then return end
	if input.pressed("search") then self.picker:show(); return end
    -- Facing changes select another direction resource, not another animation.
    -- Keep one lab-owned clock so rebuilding the retained animation can seek to
    -- the same point instead of visibly restarting at frame zero.
    if self.playing then self.playback_seconds = self.playback_seconds + (elapsed or 0) end
    if input.pressed("left") then self.direction = (self.direction + 15) % 16; self.dirty = true end
    if input.pressed("right") then self.direction = (self.direction + 1) % 16; self.dirty = true end
    if input.pressed("up") then self.mode_index = ((self.mode_index - 2) % #modes) + 1; self.dirty = true end
    if input.pressed("down") then self.mode_index = (self.mode_index % #modes) + 1; self.dirty = true end
    if input.pressed("page_up") then self.token_index = ((self.token_index - 2) % #tokens) + 1; self.weapon="HTH"; self.appearance={}; self.dirty=true end
    if input.pressed("page_down") then self.token_index = (self.token_index % #tokens) + 1; self.weapon="HTH"; self.appearance={}; self.dirty=true end
    if input.pressed("confirm") then self:choose_random() end
    if input.pressed("space") then
        self.playing = not self.playing
        if self.playing then self.frame = -1; self.actor:animation_resume()
        else self.frame = 0; self.dirty = true end
    end
    if input.pressed("home") or input.pressed("end") then
        self.playing = false
        local count = self.resolved and self.resolved.frames or 1
        self.frame = ((math.max(0, self.frame) + (input.pressed("end") and 1 or -1)) % count)
        self.dirty = true
    end
    if self.dirty then self:rebuild() end
end

return lab
