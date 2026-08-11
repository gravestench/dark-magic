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
local zoom_step = 0.05
local preview = {left=40, top=95, right=760, bottom=525}

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
    local normalized = tostring(path or ""):lower()
    -- Lord of Destruction stores Act V map art beneath "Expansion" rather
    -- than an "Act5" directory, but it still uses the Act V display palette.
    if normalized:match("/expansion/") then return 5 end
    local matched = normalized:match("/act([1-5])/")
    if not matched then return nil end
    return tonumber(matched)
end

local function quantized_zoom(value)
    local snapped = math.floor(value / zoom_step + 0.5) * zoom_step
    return math.max(zoom_step, math.min(4, snapped))
end

local function quantized_fit(value)
    local snapped = math.floor(value / zoom_step) * zoom_step
    return math.max(zoom_step, math.min(4, snapped))
end

function lab:infer_palette()
    local act = act_from_path(self.path)
    if not act then return end
    self.palette_index = act
    self.palette = palettes[act]
end

-- DS1 composition can decode many DT1 files. Queue that CPU work on the
-- bounded asset workers so a cold random map cannot consume the Lua update
-- deadline. The old complete preview remains visible until the new one is
-- ready, which also avoids presenting a partially composed map.
function lab:queue_preview()
    if self.path == "" or #self.tiles == 0 then
        self.pending_job = nil
        self.dirty = true
        return
    end
    self.pending_job = render.preload({{
        kind = "ds1",
        path = self.path,
        tiles = self.tiles,
        palette = self.palette,
    }})
    self.dirty = false
    text.set(self.status, "font_lab_color", "[gold]LOADING [blue]" .. file_name(self.path), 760, "center")
    text.set(self.detail, "font_lab_color", "[white]" .. self.path, 760, "center")
end

function lab:random_asset()
    if #self.assets == 0 then return end
    self.random_state = (self.random_state * 48271) % 2147483647
    self.path = self.assets[(self.random_state % #self.assets) + 1]
    self:infer_palette()
    local directory = self.path:match("^(.*)/[^/]+$") or "data/global/tiles"
    self.tiles = vfs.list(directory, ".dt1") or {}
    self.width, self.height = nil, nil
    self:queue_preview()
end

function lab:create()
    local dev = require("dm.dev/v1")
    self.root = render.create("hud")
    self.map_node = render.create("hud", self.root)
    self.map_node:set_visible(false)
    self.title = label(self.root, "DS1 MAP LAB", 18, "font_lab_heading")
    self.status = label(self.root, "", 62, "font_lab_color")
    self.detail = label(self.root, "", 535, "font_lab_color")
    self.help = label(self.root, "Arrows/drag: pan   Scroll/Home/End: zoom 0.05   Page Up/Down: palette   Space: fit   Enter: random DS1", 565)
    self.path = tostring(dev.option("ds1_path") or "")
    self.tiles = split_paths(dev.option("ds1_tiles"))
    self.palette = tostring(dev.option("ds1_palette") or "")
    self.palette_index = index_of(palettes, self.palette)
    self.palette = palettes[self.palette_index]
    self:infer_palette()
    self.assets = vfs.list("data/global/tiles", ".ds1") or {}
    self.random_state = dev.seed()
	self.pan_x, self.pan_y, self.zoom, self.dirty = 0, 0, 1, false
	self.dragging, self.drag_x, self.drag_y = false, 0, 0
	self.high_resolution_scroll_frames = 0
    self:queue_preview()
end

function lab:fit()
    -- Fit always rounds downward so snapping never makes the map overflow the
    -- viewport it was supposed to fit inside.
    self.zoom = quantized_fit(math.min(1, 720 / math.max(1, self.width), 430 / math.max(1, self.height)))
    self.pan_x, self.pan_y = 0, 0
end

function lab:position_map()
    self.map_node:set_scale(self.zoom, self.zoom)
    -- The retained renderer anchors images at their center. Pan that center
    -- around the middle of the map preview instead of applying top-left layout
    -- math a second time.
    self.map_node:set_position(400 + self.pan_x, 95 + 430 / 2 + self.pan_y)
end

function lab:set_zoom(value, anchor_x, anchor_y, continuous)
    local next_zoom = continuous and math.max(0.01, math.min(4, value)) or quantized_zoom(value)
    if next_zoom == self.zoom then return false end
    anchor_x, anchor_y = anchor_x or 400, anchor_y or (95 + 430 / 2)
    local center_x, center_y = 400 + self.pan_x, 95 + 430 / 2 + self.pan_y
    local local_x = (anchor_x - center_x) / self.zoom
    local local_y = (anchor_y - center_y) / self.zoom
    self.pan_x = anchor_x - 400 - local_x * next_zoom
    self.pan_y = anchor_y - (95 + 430 / 2) - local_y * next_zoom
    self.zoom = next_zoom
    return true
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
        text.set(self.status, "font_lab_color", string.format("[blue]%s   [white]%dx%d   %d DT1 source%s   zoom %.3fx   [green]ACT%d", file_name(self.path), width, height, #self.tiles, #self.tiles == 1 and "" or "s", self.zoom, self.palette_index), 760, "center")
        text.set(self.detail, "font_lab_color", "[white]" .. self.path, 760, "center")
    else
        self.map_node:set_visible(false)
        text.set(self.status, "font_lab_color", "[red]DS1 ERROR", 760, "center")
        text.set(self.detail, "font_lab_color", "[white]" .. tostring(width), 760, "center")
    end
    self.dirty = false
end

function lab:update()
    if self.pending_job then
        local status = render.preload_status(self.pending_job)
        if not status or not status.done then return end
        self.pending_job = nil
        if status.failed > 0 then
            self.map_node:set_visible(false)
            text.set(self.status, "font_lab_color", "[red]DS1 ERROR", 760, "center")
            text.set(self.detail, "font_lab_color", "[white]" .. tostring(status.errors[1] or "preview preload failed"), 760, "center")
            return
        end
        -- set_ds1 now performs a cheap cache lookup and retained-node update.
        self.dirty = true
    end
    if self.dirty then self:rebuild() end
    if input.pressed("confirm") then self:random_asset(); return end
    if input.pressed("page_up") then self.palette_index = ((self.palette_index - 2) % #palettes) + 1; self.palette = palettes[self.palette_index]; self:queue_preview(); return end
    if input.pressed("page_down") then self.palette_index = (self.palette_index % #palettes) + 1; self.palette = palettes[self.palette_index]; self:queue_preview(); return end
    if not self.width then return end
    local moved = false
    if input.pressed("left") then self.pan_x = self.pan_x - 24; moved = true end
    if input.pressed("right") then self.pan_x = self.pan_x + 24; moved = true end
    if input.pressed("up") then self.pan_y = self.pan_y - 24; moved = true end
    if input.pressed("down") then self.pan_y = self.pan_y + 24; moved = true end
    if input.pressed("home") then moved = self:set_zoom(self.zoom - zoom_step) or moved end
    if input.pressed("end") then moved = self:set_zoom(self.zoom + zoom_step) or moved end
    if input.pressed("space") then self:fit(); moved = true end

    local pointer_x, pointer_y = input.cursor()
    local pointer_inside = pointer_x >= preview.left and pointer_x <= preview.right
        and pointer_y >= preview.top and pointer_y <= preview.bottom
    local _, scroll_y = input.scroll()
    if scroll_y ~= 0 and pointer_inside then
        local fractional = math.abs(scroll_y - math.floor(scroll_y + 0.5)) > 0.0001
        if fractional then self.high_resolution_scroll_frames = 8 end
        if self.high_resolution_scroll_frames > 0 then
            -- Trackpads are an analog gesture. Multiplicative zoom feels even at
            -- every scale and deliberately bypasses the discrete 0.05 grid.
            moved = self:set_zoom(self.zoom * math.exp(scroll_y * 0.12), pointer_x, pointer_y, true) or moved
        else
            -- A notched wheel remains predictable: one notch, one 0.05 step.
            moved = self:set_zoom(self.zoom + scroll_y * zoom_step, pointer_x, pointer_y, false) or moved
        end
    elseif self.high_resolution_scroll_frames > 0 then
        self.high_resolution_scroll_frames = self.high_resolution_scroll_frames - 1
    end
    if input.pressed("pointer_primary") and pointer_inside then
        self.dragging, self.drag_x, self.drag_y = true, pointer_x, pointer_y
    elseif input.released("pointer_primary") then
        self.dragging = false
    end
    if self.dragging and input.down("pointer_primary") then
        local delta_x, delta_y = pointer_x - self.drag_x, pointer_y - self.drag_y
        if delta_x ~= 0 or delta_y ~= 0 then
            self.pan_x, self.pan_y = self.pan_x + delta_x, self.pan_y + delta_y
            self.drag_x, self.drag_y, moved = pointer_x, pointer_y, true
        end
    end
    if moved then
        self:position_map()
        text.set(self.status, "font_lab_color", string.format("[blue]%s   [white]%dx%d   %d DT1 source%s   zoom %.3fx   [green]ACT%d", file_name(self.path), self.width, self.height, #self.tiles, #self.tiles == 1 and "" or "s", self.zoom, self.palette_index), 760, "center")
    end
end

return lab
