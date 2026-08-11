-- DT1 Lab browses real tiles through the same lazy decoder and retained renderer
-- used by DS1 presentation. It intentionally contains no codec logic.

local render = require("dm.render/v1")
local input = require("dm.input/v1")
local text = require("darkmagic.ui.text")

local views = {"composite", "floor", "wall"}
local lab = {}

local function label(root, value, y, style)
    local node = render.create("hud", root)
    local _, height = text.set(node, style or "font_lab_caption", value, 760, "center")
    node:set_position(400, y + height / 2)
    return node
end

local function view_index(wanted)
    for index, value in ipairs(views) do if value == wanted then return index end end
    return 1
end

function lab:create()
    local dev = require("dm.dev/v1")
    self.root = render.create("hud")
    self.tile_node = render.create("hud", self.root)
    self.tile_node:set_position(400, 315)
    self.title = label(self.root, "DT1 TILE LAB", 18, "font_lab_heading")
    self.status = label(self.root, "", 64, "font_lab_color")
    self.detail = label(self.root, "", 510, "font_lab_color")
    self.help = label(self.root, "Left/Right: tile   Home/End: -/+10   Up/Down: composite/floor/wall", 560)
    self.path = tostring(dev.option("dt1_path") or "")
    self.palette = tostring(dev.option("dt1_palette") or "")
    self.index = math.max(0, tonumber(dev.option("dt1_tile")) or 0)
    self.view_index = view_index(tostring(dev.option("dt1_view") or "composite"))
    self.total, self.dirty = 0, true
end

function lab:rebuild()
    if self.path == "" then
        self.tile_node:set_visible(false)
        text.set(self.status, "font_lab_color", "[gold]NO DT1 SELECTED", 760, "center")
        text.set(self.detail, "font_lab_color", "[white]Pass --dt1-path and optionally --dt1-palette/--dt1-tile/--dt1-view", 760, "center")
        self.dirty = false
        return
    end
    local view = views[self.view_index]
    local ok, width, height, metadata = pcall(function()
        return self.tile_node:set_dt1(self.path, self.palette, self.index, view)
    end)
    if ok then
        self.total = metadata.total
        if self.total > 0 then self.index = self.index % self.total end
        local scale = math.min(3, 380 / math.max(1, width), 360 / math.max(1, height))
        self.tile_node:set_scale(scale, scale)
        self.tile_node:set_position(400 - width * scale / 2, 105 + (370 - height * scale) / 2)
        self.tile_node:set_visible(true)
        text.set(self.status, "font_lab_color", string.format("[white]tile %d / %d   [blue]%s   [white]%dx%d", self.index, self.total - 1, view, width, height), 760, "center")
        text.set(self.detail, "font_lab_color", string.format("[white]type/style/sequence %d/%d/%d   dir %d   rarity %d   blocks %d   source %dx%d   roof %d", metadata.type, metadata.style, metadata.sequence, metadata.direction, metadata.rarity, metadata.blocks, metadata.tile_width, metadata.tile_height, metadata.roof_height), 760, "center")
    else
        self.tile_node:set_visible(false)
        text.set(self.status, "font_lab_color", "[red]DT1 ERROR", 760, "center")
        text.set(self.detail, "font_lab_color", "[white]" .. tostring(width), 760, "center")
    end
    self.dirty = false
end

function lab:update()
    local delta = 0
    if input.pressed("left") then delta = -1 end
    if input.pressed("right") then delta = 1 end
    if input.pressed("home") then delta = -10 end
    if input.pressed("end") then delta = 10 end
    if delta ~= 0 then
        local total = math.max(1, self.total)
        self.index = (self.index + delta) % total
        self.dirty = true
    end
    if input.pressed("up") then self.view_index = ((self.view_index - 2) % #views) + 1; self.dirty = true end
    if input.pressed("down") then self.view_index = (self.view_index % #views) + 1; self.dirty = true end
    if self.dirty then self:rebuild() end
end

return lab
