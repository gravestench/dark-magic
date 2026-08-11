-- DS1 Lab composes a real map stamp from its declared DT1 dependencies. The
-- image is presentation only; game-world collision remains dm.world authority.

local render = require("dm.render/v1")
local input = require("dm.input/v1")
local text = require("darkmagic.ui.text")
local vfs = require("dm.vfs/v1")

local lab = {}
local palettes = {
    "data/global/palette/ACT1/pal.pl2", "data/global/palette/ACT2/pal.pl2",
    "data/global/palette/ACT3/pal.pl2", "data/global/palette/ACT4/pal.pl2",
    "data/global/palette/ACT5/pal.pl2",
}

local function label(root, value, y, style)
    local node = render.create("hud", root)
    local _, height = text.set(node, style or "font_lab_caption", value, 760, "center")
    node:set_position(400, y + height / 2)
    return node
end

local function split_paths(value)
    local result = {}
    for path in tostring(value or ""):gmatch("[^,]+") do
        path = path:match("^%s*(.-)%s*$")
        if path ~= "" then table.insert(result, path) end
    end
    return result
end

local function index_of(values, wanted)
    wanted = tostring(wanted or ""):lower()
    for index, value in ipairs(values) do if value:lower() == wanted then return index end end
    return 1
end

local function file_name(path)
    return tostring(path or ""):match("([^/]+)$") or tostring(path or "")
end

local function act_from_path(path)
    return tonumber(tostring(path or ""):lower():match("/act([1-5])/"))
end

function lab:infer_palette()
    local act = act_from_path(self.path)
    if not act then return end
    self.palette_index = act
    self.palette = palettes[act]
end

function lab:random_asset()
    if #self.assets == 0 then return end
    self.random_state = (self.random_state * 48271) % 2147483647
    self.path = self.assets[(self.random_state % #self.assets) + 1]
    self:infer_palette()
    local directory = self.path:match("^(.*)/[^/]+$") or "data/global/tiles"
    self.tiles = vfs.list(directory, ".dt1") or {}
    self.width, self.height = nil, nil
    self.dirty = true
end

function lab:create()
    local dev = require("dm.dev/v1")
    self.root = render.create("hud")
    self.map_node = render.create("hud", self.root)
    self.title = label(self.root, "DS1 MAP LAB", 18, "font_lab_heading")
    self.status = label(self.root, "", 62, "font_lab_color")
    self.detail = label(self.root, "", 535, "font_lab_color")
    self.help = label(self.root, "Arrows: pan   Home/End: zoom   Page Up/Down: palette   Space: fit   Enter: random DS1", 565)
    self.path = tostring(dev.option("ds1_path") or "")
    self.tiles = split_paths(dev.option("ds1_tiles"))
    self.palette = tostring(dev.option("ds1_palette") or "")
    self.palette_index = index_of(palettes, self.palette)
    self.palette = palettes[self.palette_index]
    self:infer_palette()
    self.assets = vfs.list("data/global/tiles", ".ds1") or {}
    self.random_state = 1
    self.pan_x, self.pan_y, self.zoom, self.dirty = 0, 0, 1, true
end

function lab:fit()
    self.zoom = math.min(1, 720 / math.max(1, self.width), 430 / math.max(1, self.height))
    self.pan_x, self.pan_y = 0, 0
end

function lab:position_map()
    self.map_node:set_scale(self.zoom, self.zoom)
    -- The retained renderer anchors images at their center. Pan that center
    -- around the middle of the map preview instead of applying top-left layout
    -- math a second time.
    self.map_node:set_position(400 + self.pan_x, 95 + 430 / 2 + self.pan_y)
end

function lab:rebuild()
    if self.path == "" or #self.tiles == 0 then
        self.map_node:set_visible(false)
        text.set(self.status, "font_lab_color", "[gold]NO DS1 RECIPE SELECTED", 760, "center")
        text.set(self.detail, "font_lab_color", "[white]Pass --ds1-path and comma-separated --ds1-tiles (plus the act palette if needed)", 760, "center")
        self.dirty = false
        return
    end
    local ok, width, height = pcall(function()
        return self.map_node:set_ds1(self.path, self.tiles, self.palette)
    end)
    if ok then
        self.width, self.height = width, height
        self:fit()
        self:position_map()
        self.map_node:set_visible(true)
        text.set(self.status, "font_lab_color", string.format("[blue]%s   [white]%dx%d   %d DT1 source%s   zoom %.2fx   [green]ACT%d", file_name(self.path), width, height, #self.tiles, #self.tiles == 1 and "" or "s", self.zoom, self.palette_index), 760, "center")
        text.set(self.detail, "font_lab_color", "[white]" .. self.path, 760, "center")
    else
        self.map_node:set_visible(false)
        text.set(self.status, "font_lab_color", "[red]DS1 ERROR", 760, "center")
        text.set(self.detail, "font_lab_color", "[white]" .. tostring(width), 760, "center")
    end
    self.dirty = false
end

function lab:update()
    if self.dirty then self:rebuild() end
    if input.pressed("confirm") then self:random_asset(); return end
    if input.pressed("page_up") then self.palette_index = ((self.palette_index - 2) % #palettes) + 1; self.palette = palettes[self.palette_index]; self.dirty = true; return end
    if input.pressed("page_down") then self.palette_index = (self.palette_index % #palettes) + 1; self.palette = palettes[self.palette_index]; self.dirty = true; return end
    if not self.width then return end
    local moved = false
    if input.pressed("left") then self.pan_x = self.pan_x - 24; moved = true end
    if input.pressed("right") then self.pan_x = self.pan_x + 24; moved = true end
    if input.pressed("up") then self.pan_y = self.pan_y - 24; moved = true end
    if input.pressed("down") then self.pan_y = self.pan_y + 24; moved = true end
    if input.pressed("home") then self.zoom = math.max(0.05, self.zoom / 1.25); moved = true end
    if input.pressed("end") then self.zoom = math.min(4, self.zoom * 1.25); moved = true end
    if input.pressed("space") then self:fit(); moved = true end
    if moved then
        self:position_map()
        text.set(self.status, "font_lab_color", string.format("[blue]%s   [white]%dx%d   %d DT1 source%s   zoom %.2fx   [green]ACT%d", file_name(self.path), self.width, self.height, #self.tiles, #self.tiles == 1 and "" or "s", self.zoom, self.palette_index), 760, "center")
    end
end

return lab
