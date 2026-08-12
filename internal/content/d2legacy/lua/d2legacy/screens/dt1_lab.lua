-- DT1 Lab lays every tile in one file into a scrollable gallery. Codec work
-- stays in Go; this scene only arranges retained image and text nodes.

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local text = require("d2legacy.ui.text")
local tooltip = require("d2legacy.ui.tooltip")
local fuzzy_picker = require("d2legacy.ui.fuzzy_picker")
local vfs = require("engine.vfs/v1")

local palettes = {
    "data/global/palette/ACT1/pal.dat", "data/global/palette/ACT2/pal.dat",
    "data/global/palette/ACT3/pal.dat", "data/global/palette/ACT4/pal.dat",
    "data/global/palette/ACT5/pal.dat",
}
local preview = {left=40, top=95, right=760, bottom=525}
local cell = {width=240, height=180, image_width=210, image_height=150}
local zoom_step = 0.05
local tiles_per_frame = 2
local lab = {}

-- Paul Siramy and Riiablo call this field "orientation". The old codec name
-- "type" is retained in the API, but these names explain what each value does
-- when a DS1 cell asks the DT1 collection for artwork.
local orientation_names = {
    [0]="floor", [1]="left wall", [2]="right wall",
    [3]="right half of north corner", [4]="left half of north corner",
    [5]="left wall end", [6]="right wall end", [7]="south corner",
    [8]="left wall with door", [9]="right wall with door",
    [10]="special I", [11]="special II", [12]="pillar/object",
    [13]="shadow", [14]="tree/object", [15]="roof",
    [16]="lower left wall", [17]="lower right wall",
    [18]="lower north corner", [19]="lower south corner",
}

-- Equal lookup keys receive equal quiet background colors. This makes records
-- which are rarity-weighted alternatives visible as a group without labels.
local group_colors = {
    {38, 61, 78}, {67, 48, 75}, {65, 61, 35},
    {37, 69, 54}, {76, 45, 39}, {48, 50, 77},
}

local function text_backdrop(root, top, height)
    local node = render.create("hud", root)
    node:fill_rect(800, height, 0, 0, 0, 128)
    node:set_position(400, top + height / 2)
    return node
end

local function label(root, value, y, style)
    local node = render.create("hud", root)
    local _, height = text.set(node, style or "font_lab_caption", value, 760, "center")
    node:set_position(400, y + height / 2)
    return node
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
    if normalized:match("/expansion/") then return 5 end
    local matched = normalized:match("/act([1-5])/")
    return matched and tonumber(matched) or nil
end

local function quantized_zoom(value)
    local snapped = math.floor(value / zoom_step + 0.5) * zoom_step
    return math.max(zoom_step, math.min(3, snapped))
end

function lab:infer_palette()
    local act = act_from_path(self.path)
    if not act then return end
    self.palette_index, self.palette = act, palettes[act]
end

function lab:destroy_tiles()
    for _, node in ipairs(self.tile_nodes) do
        if node:exists() then node:destroy() end
    end
    self.tile_nodes = {}
    self.tile_metadata = {}
    self.tile_groups = {}
    self.group_count = 0
    if self.tile_tooltip then self.tile_tooltip:set_visible(false) end
end

-- The gallery coordinate system is centered around zero. That means resetting
-- pan to zero always puts the center of the complete group in the viewport,
-- independent of its row and column count.
function lab:tile_position(index)
    local column = index % self.columns
    local row = math.floor(index / self.columns)
    return (column - (self.columns - 1) / 2) * cell.width,
        (row - (self.rows - 1) / 2) * cell.height
end

function lab:position_gallery()
    self.gallery:set_scale(self.zoom, self.zoom)
    self.gallery:set_position(400 + self.pan_x, (preview.top + preview.bottom) / 2 + self.pan_y)
end

function lab:center_gallery()
    self.pan_x, self.pan_y = 0, 0
    self:position_gallery()
end

function lab:center_tile(index)
    local x, y = self:tile_position(index)
    self.pan_x, self.pan_y = -x * self.zoom, -y * self.zoom
    self:position_gallery()
end

function lab:set_zoom(value, anchor_x, anchor_y, continuous)
    local next_zoom = continuous and math.max(0.02, math.min(3, value)) or quantized_zoom(value)
    if next_zoom == self.zoom then return false end
    anchor_x, anchor_y = anchor_x or 400, anchor_y or ((preview.top + preview.bottom) / 2)
    local center_x = 400 + self.pan_x
    local center_y = (preview.top + preview.bottom) / 2 + self.pan_y
    local local_x, local_y = (anchor_x - center_x) / self.zoom, (anchor_y - center_y) / self.zoom
    self.pan_x = anchor_x - 400 - local_x * next_zoom
    self.pan_y = anchor_y - (preview.top + preview.bottom) / 2 - local_y * next_zoom
    self.zoom = next_zoom
    self:position_gallery()
    return true
end

function lab:fit_gallery()
    local width = math.max(cell.width, self.columns * cell.width)
    local height = math.max(cell.height, self.rows * cell.height)
    self.zoom = math.max(0.05, math.min(1, 700 / width, 410 / height))
    self:center_gallery()
end

function lab:update_status()
    if self.total == 0 then return end
    local view = self.view_mode == "focus" and ("TILE #" .. self.index) or "GRID"
    text.set(self.status, "font_lab_color", string.format(
        "[blue]%s   [white]%d tiles   [green]ACT%d   [gold]%s   [white]zoom %.3fx",
        file_name(self.path), self.total, self.palette_index, view, self.zoom), 760, "center")
end

function lab:set_view(mode)
    self.view_mode = mode
    if mode == "focus" then
        self.zoom = 1
        self:center_tile(self.index)
    else
        self:fit_gallery()
    end
    self:update_status()
end

function lab:toggle_view()
    self:set_view(self.view_mode == "gallery" and "focus" or "gallery")
end

-- Start an incremental rebuild. Decoding every tile in one Lua update would
-- make a large DT1 hitch or exceed the scene deadline, so update() admits only
-- a couple of cells per frame.
function lab:start_gallery()
    self:destroy_tiles()
    self.total, self.build_index = 0, 0
    self.columns, self.rows = 1, 1
    self.building = self.path ~= ""
    if not self.building then
        text.set(self.status, "font_lab_color", "[gold]NO DT1 SELECTED", 760, "center")
		text.set(self.source, "font_lab_color", "[white]No mounted DT1 assets were found", 760, "center")
        return
    end
    text.set(self.status, "font_lab_color", "[gold]LOADING [blue]" .. file_name(self.path), 760, "center")
    text.set(self.source, "font_lab_color", "[white]" .. self.path, 760, "center")
end

function lab:add_tile(index)
    local image_node = render.create("hud", self.gallery)
    local ok, width, height, metadata = pcall(function()
        return image_node:set_dt1(self.path, self.palette, index, "composite")
    end)
    if not ok then
        image_node:destroy()
        self.building = false
        text.set(self.status, "font_lab_color", "[red]DT1 ERROR", 760, "center")
        text.set(self.source, "font_lab_color", "[white]tile " .. index .. ": " .. tostring(width), 760, "center")
        return false
    end

    if self.total == 0 then
        self.total = metadata.total
        -- A 720x430 viewport is wider than it is tall, so bias the square-root
        -- grid toward extra columns instead of producing a very tall strip.
        self.columns = math.max(1, math.ceil(math.sqrt(self.total * 1.6)))
        self.rows = math.max(1, math.ceil(self.total / self.columns))
        self.index = self.index % math.max(1, self.total)
    end

    local x, y = self:tile_position(index)
    local orientation = metadata.orientation or metadata.type
    local main_index = metadata.main_index or metadata.style
    local sub_index = metadata.sub_index or metadata.sequence
    local key = string.format("%d/%d/%d", orientation, main_index, sub_index)
    local group = self.tile_groups[key]
    if not group then
        self.group_count = self.group_count + 1
        group = {id=self.group_count, members={}}
        self.tile_groups[key] = group
    end
    table.insert(group.members, index)

    local color = group_colors[((group.id - 1) % #group_colors) + 1]
    local group_backdrop = render.create("hud", self.gallery)
    group_backdrop:fill_rect(cell.width - 8, cell.height - 8, color[1], color[2], color[3], 72)
    group_backdrop:set_position(x, y)
    group_backdrop:set_z(-2)
    table.insert(self.tile_nodes, group_backdrop)

    local scale = math.min(1.5, cell.image_width / math.max(1, width), cell.image_height / math.max(1, height))
    image_node:set_scale(scale, scale)
    image_node:set_position(x, y - 18)
    table.insert(self.tile_nodes, image_node)

    -- The selected tile receives a quiet gold backing. It remains centered in
    -- exactly the same cell as every other image rather than changing layout.
    if index == self.index then
        local selection = render.create("hud", self.gallery)
        selection:fill_rect(cell.width - 8, cell.height - 8, 105, 76, 18, 96)
        selection:set_position(x, y)
        selection:set_z(-1)
        table.insert(self.tile_nodes, selection)
    end

    self.tile_metadata[index + 1] = {
        index=index, orientation=orientation, main_index=main_index, sub_index=sub_index,
        direction=metadata.direction, rarity=metadata.rarity, blocks=metadata.blocks,
        width=metadata.tile_width, height=metadata.tile_height, group=group,
    }
    return true
end

function lab:tooltip_value(index)
    local item = self.tile_metadata[index + 1]
    local variant_index = 1
    for position, member in ipairs(item.group.members) do
        if member == index then variant_index = position; break end
    end
    return string.format(
        "[gold]tile #%d\n[white]orientation/type: %d - %s\nmain/style: %d\nsub/sequence: %d\n[blue]variant %d of %d [white](rarity %d)\ndirection: %d   blocks: %d   source: %dx%d",
        index, item.orientation, orientation_names[item.orientation] or "unknown",
        item.main_index, item.sub_index, variant_index, #item.group.members, item.rarity,
        item.direction, item.blocks, item.width, item.height)
end

function lab:hovered_tile(pointer_x, pointer_y)
    if self.total == 0 or self.zoom == 0 then return nil end
    if pointer_x < preview.left or pointer_x > preview.right
        or pointer_y < preview.top or pointer_y > preview.bottom then return nil end
    local local_x = (pointer_x - (400 + self.pan_x)) / self.zoom
    local local_y = (pointer_y - ((preview.top + preview.bottom) / 2 + self.pan_y)) / self.zoom
    local column = math.floor((local_x + (self.columns - 1) * cell.width / 2 + cell.width / 2) / cell.width)
    local row = math.floor((local_y + (self.rows - 1) * cell.height / 2 + cell.height / 2) / cell.height)
    if column < 0 or column >= self.columns or row < 0 or row >= self.rows then return nil end
    local index = row * self.columns + column
    if index >= self.total or not self.tile_metadata[index + 1] then return nil end
    return index
end

function lab:update_tooltip(pointer_x, pointer_y)
    local index = not self.dragging and self:hovered_tile(pointer_x, pointer_y) or nil
    if index == nil then
        self.hovered_index = nil
        self.tile_tooltip:set_visible(false)
        return
    end
    if self.hovered_index ~= index then
        self.hovered_index = index
        self.tile_tooltip:set_text(self:tooltip_value(index))
    end
    self.tile_tooltip:set_position(pointer_x + 16, pointer_y + 18)
    self.tile_tooltip:set_visible(true)
end

function lab:build_some_tiles()
    for _ = 1, tiles_per_frame do
        if not self.building then return end
        if not self:add_tile(self.build_index) then return end
        self.build_index = self.build_index + 1
        if self.build_index >= self.total then
            self.building = false
            -- Force a hovered tooltip to refresh now that every variant count
            -- is final rather than showing the partial count from assembly.
            self.hovered_index = nil
            -- Grid view is the stable default. Focusing one readable tile is an
            -- explicit view change, never a surprising final loading phase.
            self:set_view(self.view_mode)
            return
        end
    end
    text.set(self.status, "font_lab_color", string.format(
        "[gold]LAYING OUT [blue]%s   [white]%d / %d", file_name(self.path), self.build_index, self.total), 760, "center")
end

function lab:random_asset()
    if #self.assets == 0 then return end
    self.random_state = (self.random_state * 48271) % 2147483647
    self.path = self.assets[(self.random_state % #self.assets) + 1]
    self:infer_palette()
    self.index = 0
    self:start_gallery()
end

function lab:create()
    local dev = require("engine.dev/v1")
    self.root = render.create("hud")
    self.gallery = render.create("hud", self.root)
    self.gallery:set_clip(preview.left, preview.top, preview.right - preview.left, preview.bottom - preview.top)
    self.top_text_backdrop = text_backdrop(self.root, 0, 94)
    self.bottom_text_backdrop = text_backdrop(self.root, 525, 75)
    self.title = label(self.root, "DT1 TILE GALLERY", 18, "font_lab_heading")
    self.status = label(self.root, "", 64, "font_lab_color")
    self.source = label(self.root, "", 535, "font_lab_caption")
	self.help = label(self.root, "F: find asset   Enter: random   Tab: grid/tile   Arrows/drag: pan   Scroll/Home/End: zoom   Space: fit", 565)
    -- This tooltip is a sibling of the transformed gallery, so its text stays
    -- at ordinary screen scale regardless of gallery pan or zoom.
    self.tile_tooltip = tooltip.create(self.root, "", 0, 0, {
        style="dt1_lab_tooltip", max_width=310, origin_x="left", origin_y="top", alpha=192,
    })

	self.path = ""
	self.palette = palettes[1]
    self.palette_index = index_of(palettes, self.palette)
    self.palette = palettes[self.palette_index]
    self:infer_palette()
	self.index = 0
    self.assets = vfs.list("data/global/tiles", ".dt1") or {}
	self.picker = fuzzy_picker.create(self.root, {title="SELECT DT1", items=self.assets, on_select=function(path)
		self.path, self.index = path, 0
		self:infer_palette()
		self:start_gallery()
	end})
    self.random_state = dev.seed()
    self.tile_nodes = {}
    self.pan_x, self.pan_y, self.zoom = 0, 0, 1
    self.view_mode = "gallery"
    self.dragging, self.drag_x, self.drag_y = false, 0, 0
    self.high_resolution_scroll_frames = 0
	self:random_asset()
end

function lab:update()
	if self.picker:update() then return end
	if input.pressed("search") then self.picker:show(); return end
    if self.building then self:build_some_tiles() end
    if input.pressed("tab") and self.total > 0 then self:toggle_view(); return end
    if input.pressed("confirm") then self:random_asset(); return end
    if input.pressed("page_up") then
        self.palette_index = ((self.palette_index - 2) % #palettes) + 1
        self.palette = palettes[self.palette_index]
        self:start_gallery()
        return
    end
    if input.pressed("page_down") then
        self.palette_index = (self.palette_index % #palettes) + 1
        self.palette = palettes[self.palette_index]
        self:start_gallery()
        return
    end

    local moved = false
    if input.pressed("left") then self.pan_x = self.pan_x - 24; moved = true end
    if input.pressed("right") then self.pan_x = self.pan_x + 24; moved = true end
    if input.pressed("up") then self.pan_y = self.pan_y - 24; moved = true end
    if input.pressed("down") then self.pan_y = self.pan_y + 24; moved = true end
    if input.pressed("home") then moved = self:set_zoom(self.zoom - zoom_step) or moved end
    if input.pressed("end") then moved = self:set_zoom(self.zoom + zoom_step) or moved end
    if input.pressed("space") then self.view_mode = "gallery"; self:fit_gallery(); self:update_status(); moved = true end

    local pointer_x, pointer_y = input.cursor()
    local pointer_inside = pointer_x >= preview.left and pointer_x <= preview.right
        and pointer_y >= preview.top and pointer_y <= preview.bottom
    local _, scroll_y = input.scroll()
    if scroll_y ~= 0 and pointer_inside then
        local fractional = math.abs(scroll_y - math.floor(scroll_y + 0.5)) > 0.0001
        if fractional then self.high_resolution_scroll_frames = 8 end
        if self.high_resolution_scroll_frames > 0 then
            moved = self:set_zoom(self.zoom * math.exp(scroll_y * 0.12), pointer_x, pointer_y, true) or moved
        else
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
    if moved then self:position_gallery(); self:update_status() end
    self:update_tooltip(pointer_x, pointer_y)
end

return lab
