-- DS1 Map Editor is a focused authoring screen. Lua owns familiar desktop
-- interaction (canvas, tool rail, pickers, feedback); engine.map_editor/v1
-- owns source fidelity, undo history, preview materialization, and saving.

-- This capability is required when the scene opens, rather than while the
-- boot registry merely enumerates screens. Headless/older hosts can therefore
-- keep their existing scene catalog without being forced to expose editing.
local editor
local display
local input = require("engine.input/v1")
local render = require("engine.render/v1")
local text = require("ds1editor.ui.text")
local vfs = require("engine.vfs/v1")
local fuzzy_picker = require("ds1editor.ui.fuzzy_picker")
local ui_assets = require("ds1editor.ui.assets")
local ui = require("ds1editor.ui.composition")

local map_editor = {}

local palettes = {
    {label="ACT I", path="data/global/palette/ACT1/pal.dat"},
    {label="ACT II", path="data/global/palette/ACT2/pal.dat"},
    {label="ACT III", path="data/global/palette/ACT3/pal.dat"},
    {label="ACT IV", path="data/global/palette/ACT4/pal.dat"},
    {label="ACT V", path="data/global/palette/ACT5/pal.dat"},
}

local layout = {
    canvas_top=144,
    toolbar_top=64,
    toolbar_bottom=128,
    document_meta_top=84,
    tool_top=160,
    layer_top=160,
    mode_top=232,
    auto_top=288,
    tileset_heading=352,
    tileset_rows=382,
    physical_heading=510,
    tile_top=542,
}

local canvas = {left=66, top=layout.canvas_top, right=554, bottom=564}
local canvas_center_x = (canvas.left + canvas.right) / 2
local canvas_center_y = (canvas.top + canvas.bottom) / 2
local zoom_step = 0.05
local map_chunk_size = 384

local tools = {
    {id="pan", label="PAN", hint="Drag the map; hold Space for temporary pan"},
    {id="paint", label="PAINT", hint="Paint selected logical tile"},
    {id="pick", label="PICK", hint="Sample a tile from the map"},
    {id="erase", label="ERASE", hint="Clear the active layer"},
}

local layers = {
    {kind="floor", label="FLOOR"},
    {kind="wall", label="WALL"},
    {kind="shadow", label="SHADOW"},
}

local visibility_layers = {
    {kind="floor", sheet="authoring", icon="floor"},
    {kind="wall", sheet="authoring", icon="wall"},
    {kind="shadow", sheet="authoring", icon="shadow"},
    {kind="collision", sheet="utility", icon="collision"},
}

-- Test pointer coordinates against one half-open UI rectangle.
local function in_rect(x, y, rect)
    return x >= rect.left and x < rect.right and y >= rect.top and y < rect.bottom
end

-- Constrain numeric state before it reaches scroll, zoom, or palette indices.
local function clamp(value, low, high)
    return math.max(low, math.min(high, value))
end

-- Present only the authored filename in compact document chrome.
local function file_name(path)
    return tostring(path or ""):match("([^/]+)$") or tostring(path or "")
end

-- Derive a sibling-safe output directory from a mounted VFS path.
local function parent_path(path)
    return tostring(path or ""):match("^(.*)/[^/]+$") or ""
end

-- Snap keyboard zoom to stable increments while preserving wheel precision.
local function quantized_zoom(value)
    return clamp(math.floor(value / zoom_step + 0.5) * zoom_step, zoom_step, 3)
end

-- Read the live native drawable with a safe minimum for early startup frames.
local function surface_size()
    if display then
        local width, height = display.size()
        if width and height and width >= 800 and height >= 600 then return width, height end
    end
    return 800, 600
end

-- The map owns flexible space; controls retain a readable minimum width.
-- Native-resolution mode supplies these dimensions directly from the drawable
-- surface, not from a scaled 800×600 render target.
local function layout_canvas(width, height)
    local inspector_width = clamp(math.floor(width * 0.25), 360, 408)
    local left, top = 96, layout.canvas_top
    local right = math.max(left + 360, width - inspector_width - 16)
    return {left=left, top=top, right=right, bottom=height - 72}, right + 8, width - 8
end

-- Create a solid retained rectangle for canvas and content-card backgrounds.
local function box(parent, rect, color, z)
    local node = render.create("hud", parent)
    node:fill_rect(rect.right - rect.left, rect.bottom - rect.top, color[1], color[2], color[3], color[4])
    node:set_position((rect.left + rect.right) / 2, (rect.top + rect.bottom) / 2)
    node:set_z(z or 0)
    return node
end

-- Apply measured bitmap text and re-anchor the node from its top-left layout
-- coordinate. Dynamic multi-line labels must be repositioned after set_text:
-- changing the texture height while retaining its old center makes it grow
-- upward into neighboring content.
local function set_label(node, value, x, y, width, alignment, style)
    local actual_width = width or 200
    local _, height = text.set(node, style or "font_lab_caption", value, actual_width, alignment or "left")
    node:set_position(x + actual_width / 2, y + height / 2)
    return height
end

-- Create measured bitmap text using top-left UI coordinates.
local function label(parent, value, x, y, width, alignment, style, z)
    local node = render.create("hud", parent)
    node:set_z(z or 0)
    set_label(node, value, x, y, width, alignment, style)
    return node
end

-- Build an icon-only control on the mockup's fixed 64px authored plate.
local function icon_button(parent, rect, sheet, name, z, icon_x)
    local control = ui.icon_button(parent, {
        left=rect.left,
        top=rect.top,
        icon_sheet=sheet,
        icon_name=name,
        state="idle",
        z=z or 20,
        compact=(rect.right - rect.left) < 48,
    })
    if icon_x then control.icon:set_position(icon_x, 0) end
    return control
end

-- Publish one status message without rebuilding the surrounding status frame.
function map_editor:set_message(value, kind)
    self.message, self.message_kind = value or "", kind or "white"
    text.set(
        self.status,
        "font_lab_caption",
        "[" .. self.message_kind .. "]" .. self.message .. "[/]",
        math.max(240, self.surface_width - 820),
        "left"
    )
end

function map_editor:refresh_status_detail()
    if not self.status_detail then return end
    local point = self.selected_cell and self.selected_cell.point
    local tile = point and string.format("%d, %d", point.x, point.y) or "--, --"
    local layer = layers[self.layer_index] and layers[self.layer_index].label or "FLOOR"
    text.set(
        self.status_detail,
        "font_lab_caption",
        string.format("TILE: %s    LAYER: %s    ZOOM: %d%%", tile, layer, math.floor(self.zoom * 100 + 0.5)),
        560,
        "right"
    )
end

-- Synchronize visible control state with the authoritative editor session.
function map_editor:refresh_chrome()
    text.set(self.title, "font_lab_color", "DS1 MAP EDITOR", self.surface_width - 32, "center")
    text.set(self.subtitle, "font_lab_caption", "", 1, "right")
    text.set(self.document_meta, "font_lab_caption", "", 1, "left")

    for index, tool in ipairs(tools) do
        local selected = self.tool == tool.id
        self.tool_boxes[index]:set_state(selected and "selected" or "idle")
    end
    self:refresh_status_detail()
    if self.grid_button then
        self.grid_button:set_state(self.grid_visible and "selected" or "idle")
    end
    if self.auto_button then
        self.auto_button:set_state(self.auto_draw and "selected" or "idle")
        self.auto_button:set_label(self.auto_draw and "AUTO: ON" or "AUTO: OFF")
    end
    local visibility = self.inspector_mode == "selected" and self.selected_visibility or self.map_visibility
    for index, layer in ipairs(visibility_layers) do
        self.layer_boxes[index]:set_state(visibility[layer.kind] and "selected" or "idle")
    end
    local brush = self.brush
    if self.brush_meta and not self.selected_compact then
        set_label(self.brush_meta, string.format(
            "[gold]BRUSH[/]\n[white]%02d / %02d / %02d[/]\n[grey]type / style / sequence[/]",
            brush.orientation or 0, brush.style or 0, brush.sequence or 0
        ), self.inspector_left + 24, layout.tileset_heading + 4,
            self.selected_preview_rect.left - self.inspector_left - 36, "left", "font_lab_caption")
    end
end

-- Switch inspector subtrees while retaining both panels for cheap toggles.
function map_editor:set_inspector_mode(mode)
    self.inspector_mode = mode
    if self.library_root then self.library_root:set_visible(mode == "library") end
    if self.selected_root then self.selected_root:set_visible(mode == "selected") end
    for index, value in ipairs({"library", "selected"}) do
        local selected = mode == value
        self.mode_boxes[index]:set_state(selected and "selected" or "tab")
    end
    if mode == "selected" then self:refresh_selected_preview() end
    self:refresh_chrome()
end

-- Toggle learned transition selection independently from the active brush.
function map_editor:toggle_auto_draw()
    self.auto_draw = not self.auto_draw
    self:refresh_chrome()
    local message = self.auto_draw
        and "Auto Draw enabled: compatible learned tile transitions will be preferred."
        or "Auto Draw disabled: paint the selected logical tile exactly."
    self:set_message(message, "white")
end

-- Choose a display palette without touching authored DS1 data or undo history.
-- Missing act palettes resolve to Act I, while explicit menu choices survive later map opens.
function map_editor:set_palette(index, explicit, suppress_refresh)
    local fallback = palettes[1]
    local selected = palettes[clamp(index or 1, 1, #palettes)] or fallback
    if not render.asset_exists(selected.path) then selected, index = fallback, 1 end
    local changed = self.palette ~= selected.path
    self.palette, self.palette_index = selected.path, index
    if explicit then self.palette_override = true end
    if self.palette_button then self.palette_button:set_label("PALETTE: " .. selected.label) end
    for item, button in ipairs(self.palette_options or {}) do
        button:set_state(item == index and "selected" or "idle")
    end
    if changed and self.summary and not suppress_refresh then self:queue_preview(false) end
end

-- Show or hide the measured palette drop-down buttons as one interaction group.
function map_editor:toggle_palette_menu()
    self.palette_menu_open = not self.palette_menu_open
    for _, button in ipairs(self.palette_options or {}) do
        button:set_visible(self.palette_menu_open)
    end
    self.palette_button:set_state(self.palette_menu_open and "pressed" or "idle")
end

-- Apply the current pan and zoom once at the retained map root.
function map_editor:position_map()
    self.map_root:set_position(canvas_center_x + self.pan_x, canvas_center_y + self.pan_y)
    self.map_root:set_scale(self.zoom, self.zoom)
    if self.map_visibility and self.map_visibility.collision then self:refresh_collision_nodes() end
    self:refresh_status_detail()
end

-- Change zoom while preserving the map-local point currently under the mouse cursor.
function map_editor:zoom_at(pointer_x, pointer_y, factor)
    local previous = self.zoom
    local next_zoom = clamp(previous * factor, 0.03, 3)
    if next_zoom == previous then return end
    local local_x = (pointer_x - (canvas_center_x + self.pan_x)) / previous
    local local_y = (pointer_y - (canvas_center_y + self.pan_y)) / previous
    self.zoom = next_zoom
    self.pan_x = pointer_x - canvas_center_x - local_x * next_zoom
    self.pan_y = pointer_y - canvas_center_y - local_y * next_zoom
    self:position_map()
end

-- Release every retained node and preload handle owned by the current preview.
function map_editor:clear_preview()
    if self.preview_job then
        render.preload_release(self.preview_job)
        self.preview_job = nil
    end
    for _, retained in pairs(self.chunk_nodes) do
        if retained.node:exists() then retained.node:destroy() end
    end
    self.chunk_nodes = {}
    self.visible_signature, self.visible_pending = nil, true
    for _, node in pairs(self.collision_nodes or {}) do
        if node:exists() then node:destroy() end
    end
    self.collision_nodes = {}
    self:clear_grid()
    self:clear_selection()
    self:clear_hover()
end

-- Recreate retained chrome on resize while preserving the authoritative Go
-- document and the user's unsaved edit state. Reopening from the mount here
-- would be a data-loss bug, so the rebuilt scene asks the existing session for
-- its current preview instead.
function map_editor:relayout_if_needed()
    local width, height = surface_size()
    if width == self.surface_width and height == self.surface_height then return false end
    local saved = {
        path=self.path, save_path=self.save_path, summary=self.summary,
        palette=self.palette, palette_index=self.palette_index, palette_override=self.palette_override,
        zoom=self.zoom, pan_x=self.pan_x, pan_y=self.pan_y,
        tool=self.tool, layer_index=self.layer_index, brush=self.brush,
        auto_draw=self.auto_draw, inspector_mode=self.inspector_mode, selected_cell=self.selected_cell,
        map_visibility=self.map_visibility, selected_visibility=self.selected_visibility,
    }
    self:destroy()
    self:create()
    if saved.summary then
        self.path, self.save_path, self.summary = saved.path, saved.save_path, editor.summary()
        self.palette, self.palette_index = saved.palette, saved.palette_index
        self.palette_override = saved.palette_override
        self.palette_button:set_label("PALETTE: " .. palettes[self.palette_index].label)
        self.zoom, self.pan_x, self.pan_y = saved.zoom, saved.pan_x, saved.pan_y
        self.tool, self.layer_index, self.brush = saved.tool, saved.layer_index, saved.brush
        self.auto_draw, self.selected_cell = saved.auto_draw, saved.selected_cell
        self.map_visibility, self.selected_visibility = saved.map_visibility, saved.selected_visibility
        self:set_inspector_mode(saved.inspector_mode or "library")
        self:refresh_chrome()
        self:refresh_tilesets(self.brush.source_path)
        self:refresh_selected_view()
        self:queue_preview(true)
        self:set_message("Workspace resized. Your unsaved edits are still active.", "white")
    end
    return true
end

-- Fit the complete chunk bounds inside the canvas without upscaling them.
function map_editor:fit()
    if not self.chunk_set then return end
    self.zoom = clamp(math.min(1, (canvas.right - canvas.left - 18) / self.chunk_set.width,
        (canvas.bottom - canvas.top - 18) / self.chunk_set.height), 0.05, 1)
    self.pan_x, self.pan_y = 0, 0
    self:position_map()
end

-- Queue only the lightweight placement index. DT1 pixels remain lazy and are
-- decoded when one visible composite chunk first needs them.
function map_editor:queue_preview(fit, dirty_points)
    if not self.summary then return end
    if self.preview_job then
        render.preload_release(self.preview_job)
        self.preview_job = nil
    end
    local ok, world = pcall(editor.preview, dirty_points)
    if not ok then
        self:set_message(tostring(world), "red")
        return
    end
    self.preview_world = world
    self.pending_fit = fit == true
    self.pending_dirty = dirty_points
    self.preview_job = render.preload({{kind="world_tile_index", world=world, palette=self.palette}})
    self:set_message(dirty_points and "Updating changed map tiles…" or "Indexing map tiles…", "gold")
end

local function map_chunk_key(column, row)
    return tostring(column) .. ":" .. tostring(row)
end

local function dirty_cell_key(x, y)
    return tostring(x) .. ":" .. tostring(y)
end

-- Mark only spatial chunks that can contain graphics owned by edited cells.
-- The generous upward reach covers tall wall and roof sprites while retaining
-- every unrelated chunk and its GPU texture.
function map_editor:invalidate_dirty_chunks(points)
    if not self.preview_world or not self.chunk_set then return end
    for _, point in ipairs(points or {}) do
        local pixel_x, pixel_y = self.preview_world:subtile_to_pixel(point.x * 5, point.y * 5)
        local first_column = math.floor((pixel_x - 192) / map_chunk_size)
        local last_column = math.floor((pixel_x + 192) / map_chunk_size)
        local first_row = math.floor((pixel_y - 640) / map_chunk_size)
        local last_row = math.floor((pixel_y + 160) / map_chunk_size)
        for row = first_row, last_row do
            for column = first_column, last_column do
                local retained = self.chunk_nodes[map_chunk_key(column, row)]
                if retained then retained.dirty = true end
            end
        end
    end
    self.visible_pending = true
end

function map_editor:invalidate_all_chunks()
    for _, retained in pairs(self.chunk_nodes) do retained.dirty = true end
    self.visible_signature, self.visible_pending = nil, true
end

-- Publish completed background preview work and stream its visible chunks.
function map_editor:finish_preview()
    if self.preview_job then
        local status = render.preload_status(self.preview_job)
        if not status or not status.done then return end
        render.preload_release(self.preview_job)
        self.preview_job = nil
        if status.failed and status.failed > 0 then
            if self.palette ~= palettes[1].path then
                self:set_palette(1, false, true)
                self:queue_preview(false)
                self:set_message("Selected palette failed; retrying with Act I.", "gold")
                return
            end
            self:set_message(tostring(status.errors[1] or "Could not render this map"), "red")
            return
        end
        local ok, chunk_set = pcall(render.world_tile_metadata, self.preview_world, self.palette)
        if not ok then
            self:set_message(tostring(chunk_set), "red")
            return
        end
        local incremental = self.pending_dirty and self.chunk_set
        if not incremental then
            self:clear_preview()
        end
        self.chunk_set = chunk_set
        if incremental then self:invalidate_dirty_chunks(self.pending_dirty) end
        self.visible_pending = true
        if self.pending_fit then self:fit() else self:position_map() end
        self:refresh_collision_nodes(self.pending_dirty)
        if self.selected_cell then self:refresh_selected_preview() end
        self.pending_fit = false
        self.pending_dirty = nil
        self:rebuild_grid()
        if self.selected_cell then self:highlight_selection(self.selected_cell.point) end
        self:set_message("Streaming visible map tiles…", "gold")
    end

    if not self.chunk_set then return end
    self:refresh_visible_chunks()
end

-- Destroy committed selection feedback before selecting another map cell.
function map_editor:clear_selection()
    for _, node in ipairs(self.selection_nodes or {}) do
        if node:exists() then node:destroy() end
    end
    self.selection_nodes, self.hover_nodes = {}, {}
end

-- Draw one isometric DS1 cell outline into a caller-owned retained-node list.
local function outline_tile(self, point, color, thickness, alpha, z, destination)
    if not point or not self.preview_world or not self.chunk_set then return end
    local corners = {
        {self.preview_world:subtile_to_pixel(point.x * 5, point.y * 5)},
        {self.preview_world:subtile_to_pixel((point.x + 1) * 5, point.y * 5)},
        {self.preview_world:subtile_to_pixel((point.x + 1) * 5, (point.y + 1) * 5)},
        {self.preview_world:subtile_to_pixel(point.x * 5, (point.y + 1) * 5)},
    }
    for index = 1, 4 do
        local first, second = corners[index], corners[index % 4 + 1]
        local dx, dy = second[1] - first[1], second[2] - first[2]
        local length = math.max(1, math.floor(math.sqrt(dx * dx + dy * dy) + 0.5))
        local node = render.create("hud", self.map_root)
        node:fill_rect(length, thickness, color[1], color[2], color[3], alpha)
        node:set_position((first[1] + second[1]) / 2 - self.chunk_set.width / 2,
            (first[2] + second[2]) / 2 - self.chunk_set.height / 2)
        node:set_rotation(math.deg(math.atan(dy / dx)))
        node:set_z(z)
        destination[#destination + 1] = node
    end
end

-- Highlight the committed PICK selection independently from transient pointer hover.
function map_editor:highlight_selection(point)
    self:clear_selection()
    self.selected_point = point
    if not point then return end
    local layer = layers[self.layer_index]
    local colors = {floor={84, 201, 141}, wall={241, 190, 77}, shadow={174, 126, 222}}
    outline_tile(self, point, colors[layer.kind], 2, 224, 9200, self.selection_nodes)
end

-- Destroy transient eyedropper feedback without disturbing the committed selection outline.
function map_editor:clear_hover()
    for _, node in ipairs(self.hover_nodes or {}) do
        if node:exists() then node:destroy() end
    end
    self.hover_nodes, self.hover_key = {}, nil
end

-- Return the active floor, wall, or shadow record from one inspected cell.
local function active_cell_value(cell, kind)
    local values = kind == "floor" and cell.floors or (kind == "wall" and cell.walls or cell.shadows)
    return values and values[1] or nil
end

-- Preview exactly one eyedropper target. Auto Draw affects painting, not the
-- sampling footprint, so PICK never implies a multi-cell selection.
function map_editor:hover_pick(point)
    local layer = layers[self.layer_index]
    local key = point and table.concat({point.x, point.y, layer.kind}, ":") or ""
    if key == self.hover_key then return end
    self:clear_hover()
    self.hover_key = key
    if not point then return end

    local ok, cell = pcall(editor.cell, point.x, point.y)
    local value = ok and active_cell_value(cell, layer.kind) or nil
    if not value or (value.properties or 0) == 0 then return end

    local colors = {floor={84, 201, 141}, wall={241, 190, 77}, shadow={174, 126, 222}}
    outline_tile(self, point, colors[layer.kind], 2, 196, 9190, self.hover_nodes)
end

-- Release every retained debug-grid segment as one toggleable group.
function map_editor:clear_grid()
    for _, node in pairs(self.grid_nodes or {}) do
        if node:exists() then node:destroy() end
    end
    self.grid_nodes = {}
end

-- Draw shared isometric tile boundaries, rather than four edges per cell. The
-- result is a light diamond lattice that makes empty authored DS1 cells easy
-- to find without obscuring the art or allocating a node per tile.
function map_editor:rebuild_grid()
    self:clear_grid()
    if not self.grid_visible or not self.summary or not self.preview_world or not self.chunk_set then return end
    -- Retain one thin segment using map-local coordinates and shared styling.
    local function line(x1, y1, x2, y2)
        local dx, dy = x2 - x1, y2 - y1
        local length = math.max(1, math.floor(math.sqrt(dx * dx + dy * dy) + 0.5))
        local node = render.create("hud", self.map_root)
        node:fill_rect(length, 1, 119, 156, 186, 54)
        node:set_position((x1 + x2) / 2 - self.chunk_set.width / 2, (y1 + y2) / 2 - self.chunk_set.height / 2)
        -- The two projected boundary families always have a non-zero X span.
        -- One-argument atan keeps this compatible with the Lua 5.1 math API.
        node:set_rotation(math.deg(math.atan(dy / dx)))
        node:set_z(9000)
        self.grid_nodes[#self.grid_nodes + 1] = node
    end
    -- Five subtiles form one authored DS1 map tile. Constant-X and constant-Y
    -- subtile paths are the two diagonal families of the isometric lattice.
    for x = 0, self.summary.width do
        local x1, y1 = self.preview_world:subtile_to_pixel(x * 5, 0)
        local x2, y2 = self.preview_world:subtile_to_pixel(x * 5, self.summary.height * 5)
        line(x1, y1, x2, y2)
    end
    for y = 0, self.summary.height do
        local x1, y1 = self.preview_world:subtile_to_pixel(0, y * 5)
        local x2, y2 = self.preview_world:subtile_to_pixel(self.summary.width * 5, y * 5)
        line(x1, y1, x2, y2)
    end
end

-- Toggle the subtle empty-cell grid without rebuilding map chunks.
function map_editor:toggle_grid()
    self.grid_visible = not self.grid_visible
    self:rebuild_grid()
    self:refresh_chrome()
    local message = self.grid_visible
        and "DS1 grid shown (F4 toggles it)."
        or "DS1 grid hidden (F4 toggles it)."
    self:set_message(message, "white")
end

-- Keep collision diagnostics in one retained node per authored cell. Edits
-- replace only returned dirty cells; toggling or panning merely changes node visibility.
function map_editor:refresh_collision_nodes(points)
    self.collision_nodes = self.collision_nodes or {}
    if not self.map_visibility or not self.map_visibility.collision
        or not self.preview_world or not self.chunk_set or not self.summary then
        for _, node in pairs(self.collision_nodes) do node:set_visible(false) end
        return
    end

    local targets = {}
    if points then
        for _, point in ipairs(points) do targets[dirty_cell_key(point.x, point.y)] = point end
    else
        local map_left = (canvas.left - (canvas_center_x + self.pan_x)) / self.zoom + self.chunk_set.width / 2
        local map_top = (canvas.top - (canvas_center_y + self.pan_y)) / self.zoom + self.chunk_set.height / 2
        local map_right = (canvas.right - (canvas_center_x + self.pan_x)) / self.zoom + self.chunk_set.width / 2
        local map_bottom = (canvas.bottom - (canvas_center_y + self.pan_y)) / self.zoom + self.chunk_set.height / 2
        for y = 0, self.summary.height - 1 do
            for x = 0, self.summary.width - 1 do
                local pixel_x, pixel_y = self.preview_world:subtile_to_pixel(x * 5, y * 5)
                if pixel_x >= map_left - 160 and pixel_x <= map_right + 160
                    and pixel_y >= map_top - 80 and pixel_y <= map_bottom + 80 then
                    targets[dirty_cell_key(x, y)] = {x=x, y=y}
                end
            end
        end
    end

    for key, point in pairs(targets) do
        local node = self.collision_nodes[key]
        if not node then
            node = render.create("hud", self.map_root)
            node:set_z(9050)
            self.collision_nodes[key] = node
        end
        local left, top, width, height = node:set_world_collision_region(
            self.preview_world, point.x * 5, point.y * 5, (point.x + 1) * 5, (point.y + 1) * 5
        )
        node:set_position(left - self.chunk_set.width / 2 + width / 2,
            top - self.chunk_set.height / 2 + height / 2)
        node:set_visible(true)
    end
    if not points then
        for key, node in pairs(self.collision_nodes) do
            if not targets[key] then node:set_visible(false) end
        end
    end
end

-- Render the map as a small grid of flattened spatial textures. Chunk nodes are
-- retained while visible; a paint stroke updates only chunks intersecting its
-- edited cells, and panning inside the same chunk range performs no work.
function map_editor:refresh_visible_chunks()
    if not self.chunk_set then return end
    local map_left = (canvas.left - (canvas_center_x + self.pan_x)) / self.zoom + self.chunk_set.width / 2
    local map_top = (canvas.top - (canvas_center_y + self.pan_y)) / self.zoom + self.chunk_set.height / 2
    local map_right = (canvas.right - (canvas_center_x + self.pan_x)) / self.zoom + self.chunk_set.width / 2
    local map_bottom = (canvas.bottom - (canvas_center_y + self.pan_y)) / self.zoom + self.chunk_set.height / 2
    local first_column = math.max(0, math.floor(map_left / map_chunk_size))
    local last_column = math.min(math.ceil(self.chunk_set.width / map_chunk_size) - 1,
        math.floor(map_right / map_chunk_size))
    local first_row = math.max(0, math.floor(map_top / map_chunk_size))
    local last_row = math.min(math.ceil(self.chunk_set.height / map_chunk_size) - 1,
        math.floor(map_bottom / map_chunk_size))
    local signature = table.concat({first_column, first_row, last_column, last_row}, ":")
    if self.visible_signature == signature and not self.visible_pending then return end

    local visible, admitted, remaining = {}, 0, 0
    local chunk_publish_budget = 4
    for row = first_row, last_row do
        for column = first_column, last_column do
            local key = map_chunk_key(column, row)
            visible[key] = true
            local retained = self.chunk_nodes[key]
            if not retained then
                retained = {
                    node=render.create("hud", self.map_root),
                    column=column,
                    row=row,
                    dirty=true,
                }
                retained.node:set_z(0)
                self.chunk_nodes[key] = retained
            end

            if retained.dirty and admitted < chunk_publish_budget then
                local left, top = column * map_chunk_size, row * map_chunk_size
                local right = math.min(left + map_chunk_size, self.chunk_set.width)
                local bottom = math.min(top + map_chunk_size, self.chunk_set.height)
                local ok, x, y, width, height = pcall(function()
                    return retained.node:set_world_composite_region(
                        self.preview_world,
                        self.palette,
                        left,
                        top,
                        right,
                        bottom,
                        self.map_visibility.floor,
                        self.map_visibility.wall,
                        self.map_visibility.shadow
                    )
                end)
                if ok then
                    retained.node:set_position(
                        x - self.chunk_set.width / 2 + width / 2,
                        y - self.chunk_set.height / 2 + height / 2
                    )
                    retained.dirty = false
                    admitted = admitted + 1
                else
                    self:set_message(tostring(x), "red")
                    retained.node:destroy()
                    self.chunk_nodes[key] = nil
                end
            elseif retained.dirty then
                remaining = remaining + 1
            end
        end
    end

    for key, retained in pairs(self.chunk_nodes) do
        if not visible[key] then
            retained.node:destroy()
            self.chunk_nodes[key] = nil
        end
    end

    if remaining > 0 then
        self:set_message("Publishing " .. tostring(remaining) .. " visible map chunks…", "gold")
    elseif next(self.chunk_nodes) then
        self:set_message("Ready. Paint, pick, erase, or pan — every stroke is undoable.", "green")
    end
    self.visible_signature, self.visible_pending = signature, remaining > 0
end

-- Open one mounted DS1 and refresh all derived palette, tileset, and preview state.
function map_editor:open(path)
    local ok, summary = pcall(editor.open, path)
    if not ok then
        self:set_message(tostring(summary), "red")
        return
    end
    self.path, self.summary = path, summary
    self.selected_cell = nil
    self:clear_selection()
    if not self.palette_override then
        self:set_palette(clamp(summary.act or 1, 1, #palettes), false, true)
    end
    self.save_path = path
    self:refresh_chrome()
    self:refresh_tilesets()
    self:queue_preview(true)
end

-- Re-read the inspected record after edits, undo, or redo changes its snapshot.
function map_editor:reload_selected_cell()
    local selected = self.selected_cell
    if not selected then return end
    local ok, cell = pcall(editor.cell, selected.point.x, selected.point.y)
    if not ok then self.selected_cell = nil; self:clear_selection(); return end
    local values = selected.kind == "floor" and cell.floors or (selected.kind == "wall" and cell.walls or cell.shadows)
    local value = values[selected.layer + 1]
    if not value or (value.properties or 0) == 0 then
        self.selected_cell = nil
        self:clear_selection()
    else
        selected.value = value
        self:select_tile(value)
        self:highlight_selection(selected.point)
    end
    self:refresh_selected_view()
end

-- Rebuild the DT1 list for the active layer and preserve a meaningful selection.
function map_editor:refresh_tilesets(preferred_path)
    if not self.summary then return end
    local layer = layers[self.layer_index]
    local ok, sets = pcall(editor.tilesets, layer.kind)
    if not ok then self:set_message(tostring(sets), "red"); return end
    self.tilesets, self.tileset_index = sets, nil
    for index, value in ipairs(sets) do
        if value.path == preferred_path then self.tileset_index = index; break end
    end
    self.tileset_index = self.tileset_index or (#sets > 0 and 1 or nil)
    self.tileset_scroll = clamp((self.tileset_index or 1) - 1, 0, math.max(0, #sets - #self.tileset_rows))
    self:refresh_tileset_rows()
    self:refresh_tileset_tiles()
end

-- Rebind the small visible tileset row pool after selection or scrolling.
function map_editor:refresh_tileset_rows()
    for row, control in ipairs(self.tileset_rows) do
        local index = self.tileset_scroll + row
        local value = self.tilesets[index]
        control.box:set_visible(value ~= nil)
        control.label:set_visible(value ~= nil)
        if value then
            local selected = index == self.tileset_index
            control.box:set_state(selected and "selected" or "idle")
            local format = selected and "[gold]%s[/]  [grey]%d[/]" or "%s  [grey]%d[/]"
            text.set(control.label, "font_lab_caption", string.format(format, value.label, value.count),
                self.inspector_right - self.inspector_left - 46, "left")
        end
    end
end

-- Select a DT1 source and expose its physical tile records.
function map_editor:select_tileset(index)
    if not self.tilesets[index] then return end
    self.tileset_index = index
    self:refresh_tileset_rows()
    self:refresh_tileset_tiles()
end

-- Load physical tile metadata for the selected DT1 and restore a sampled index.
function map_editor:refresh_tileset_tiles(preferred_source_index)
    self.tileset_tiles = {}
    if self.tileset_index then
        local layer = layers[self.layer_index]
        local ok, values = pcall(editor.tileset_tiles, self.tilesets[self.tileset_index].path, layer.kind)
        if ok then self.tileset_tiles = values else self:set_message(tostring(values), "red") end
    end
    self.tile_scroll = 0
    if preferred_source_index then
        for index, value in ipairs(self.tileset_tiles) do
            if value.source_index == preferred_source_index then
                self.tile_scroll = math.floor((index - 1) / self.tile_columns) * self.tile_columns
                break
            end
        end
    end
    self:refresh_tile_grid()
end

-- Rebind visible physical-tile cards without recreating their retained nodes.
function map_editor:refresh_tile_grid()
    for slot, control in ipairs(self.tile_cells) do
        local choice = self.tileset_tiles[self.tile_scroll + slot]
        control.choice = choice
        control.box:set_visible(choice ~= nil)
        control.preview:set_visible(choice ~= nil)
        control.label:set_visible(choice ~= nil)
        if choice then
            local selected = self.brush.source_path == choice.source_path
                and self.brush.source_index == choice.source_index
            control.box:set_state(selected and "selected" or "idle")
            local ok, width, height = pcall(function()
                return control.preview:set_dt1(choice.source_path, self.palette, choice.source_index, "composite")
            end)
            if ok then
                local scale = math.min(
                    0.7,
                    control.preview_width / math.max(1, width),
                    control.preview_height / math.max(1, height)
                )
                control.preview:set_scale(scale, scale)
            end
            local caption = string.format(
                "#%d  %02d/%02d/%02d",
                choice.source_index,
                choice.orientation,
                choice.style,
                choice.sequence
            )
            if selected then caption = "[gold]" .. caption .. "[/]" end
            text.set(control.label, "font_lab_caption", caption, control.width - 4, "center")
        end
    end
end

-- Align both inspector lists with a tile sampled from the map.
function map_editor:sync_tileset_selection(choice)
    if not choice or not choice.source_path then return end
    local found
    for index, value in ipairs(self.tilesets or {}) do
        if value.path == choice.source_path then found = index; break end
    end
    if not found then
        self:refresh_tilesets(choice.source_path)
        found = self.tileset_index
    else
        self.tileset_index = found
        self.tileset_scroll = clamp(found - 1, 0, math.max(0, #self.tilesets - #self.tileset_rows))
        self:refresh_tileset_rows()
        self:refresh_tileset_tiles(choice.source_index)
    end
end

-- Populate and open the keyboard-searchable logical tile picker.
function map_editor:show_tile_picker()
    if not self.summary then return end
    local layer = layers[self.layer_index]
    local ok, choices = pcall(editor.tiles, layer.kind)
    if not ok then
        self:set_message(tostring(choices), "red")
        return
    end
    self.tile_by_label = {}
    self.tile_picker.items = {}
    for _, choice in ipairs(choices) do
        local label = choice.label
        self.tile_by_label[label] = choice
        self.tile_picker.items[#self.tile_picker.items + 1] = label
    end
    if #self.tile_picker.items == 0 then
        self:set_message("No compatible DT1 identities were declared by this map.", "gold")
        return
    end
    self.tile_picker:show()
end

-- Adopt one catalog choice as the active physical brush.
function map_editor:select_tile(choice)
    if not choice then return end
    self.brush = {
        orientation=choice.orientation or 0,
        style=choice.style or 0,
        sequence=choice.sequence or 0,
        properties=choice.properties or 0,
        source_path=choice.source_path,
        source_index=choice.source_index,
        variants=choice.variants or 0,
    }
    if self.brush.source_path then
        local ok, width, height = pcall(function()
            return self.tile_preview:set_dt1(self.brush.source_path, self.palette, self.brush.source_index, "composite")
        end)
        if ok then
            local bounds = self.selected_preview_rect
            local scale = math.min(
                1,
                (bounds.right - bounds.left) / math.max(1, width),
                (bounds.bottom - bounds.top) / math.max(1, height)
            )
            self.tile_preview:set_scale(scale, scale)
        end
    end
    self:refresh_chrome()
    self:sync_tileset_selection(choice)
    self:refresh_tile_grid()
    self:refresh_selected_view()
    local selected = string.format(
        "%02d / %02d / %02d",
        self.brush.orientation,
        self.brush.style,
        self.brush.sequence
    )
    self:set_message("Selected tile " .. selected, "green")
end

-- Inspect the active layer at one cell and synchronize the brush to its DT1 tile.
function map_editor:sample_at(point)
    local ok, cell = pcall(editor.cell, point.x, point.y)
    if not ok then self:set_message(tostring(cell), "red"); return end
    local layer = layers[self.layer_index]
    local values = layer.kind == "floor" and cell.floors or (layer.kind == "wall" and cell.walls or cell.shadows)
    local value = values[1]
    if not value or (value.properties or 0) == 0 then
        self:set_message("That cell is empty on the active layer.", "gold")
        return
    end
    self.selected_cell = {point=point, kind=layer.kind, layer=0, value=value}
    self:select_tile({
        orientation=value.orientation or (layer.kind == "shadow" and 13 or 0),
        style=value.style, sequence=value.sequence, properties=value.properties,
        source_path=value.source_path, source_index=value.source_index, rarity=value.rarity,
    })
    self:highlight_selection(point)
    self:set_inspector_mode("selected")
    self:refresh_selected_view()
    self:set_message("Sampled tile from " .. point.x .. ", " .. point.y, "green")
end

function map_editor:clear_selected_preview_nodes()
    for _, node in ipairs(self.selected_preview_nodes or {}) do
        if node:exists() then node:destroy() end
    end
    self.selected_preview_nodes = {}
end

-- Compose the sampled cell from independently toggleable floor, wall, shadow,
-- and collision nodes. The ordinary brush preview remains available before a cell is sampled.
function map_editor:refresh_selected_preview()
    if not self.tile_preview then return end
    self:clear_selected_preview_nodes()
    if self.selected_compact then
        self.tile_preview:set_visible(false)
        return
    end
    local selected = self.selected_cell
    if not selected then
        local active_kind = layers[self.layer_index].kind
        self.tile_preview:set_visible(self.selected_visibility[active_kind])
        return
    end

    self.tile_preview:set_visible(false)
    local ok, cell = pcall(editor.cell, selected.point.x, selected.point.y)
    if not ok then return end
    local bounds = self.selected_preview_rect
    local baseline = (bounds.bottom - bounds.top) / 2
    local available_width = bounds.right - bounds.left
    local available_height = bounds.bottom - bounds.top
    local groups = {
        {kind="floor", values=cell.floors, z=2},
        {kind="wall", values=cell.walls, z=3},
        {kind="shadow", values=cell.shadows, z=4},
    }
    for _, group in ipairs(groups) do
        if self.selected_visibility[group.kind] then
            for _, value in ipairs(group.values or {}) do
                if value.source_path and value.source_index and (value.properties or 0) ~= 0 then
                    local node = render.create("hud", self.selected_preview_root)
                    local loaded, width, height = pcall(function()
                        return node:set_dt1(value.source_path, self.palette, value.source_index, "composite")
                    end)
                    if loaded then
                        local scale = math.min(
                            1,
                            available_width / math.max(1, width),
                            available_height / math.max(1, height)
                        )
                        node:set_scale(scale, scale)
                        node:set_position(0, baseline - height * scale / 2)
                        node:set_clip(bounds.left, bounds.top, available_width, available_height)
                        node:set_z(group.z)
                        self.selected_preview_nodes[#self.selected_preview_nodes + 1] = node
                    else
                        node:destroy()
                    end
                end
            end
        end
    end
    if self.selected_visibility.collision and self.preview_world then
        local node = render.create("hud", self.selected_preview_root)
        local point = selected.point
        local _, _, width, height = node:set_world_collision_region(
            self.preview_world, point.x * 5, point.y * 5, (point.x + 1) * 5, (point.y + 1) * 5
        )
        local scale = math.min(
            1,
            available_width / math.max(1, width),
            available_height / math.max(1, height)
        )
        node:set_scale(scale, scale)
        node:set_position(0, baseline - height * scale / 2)
        node:set_clip(bounds.left, bounds.top, available_width, available_height)
        node:set_z(8)
        self.selected_preview_nodes[#self.selected_preview_nodes + 1] = node
    end
end

-- Render metadata and edit controls for the committed sampled record.
function map_editor:refresh_selected_view()
    if not self.selected_meta then return end
    self:refresh_selected_preview()
    local selected, value = self.selected_cell, self.selected_cell and self.selected_cell.value
    if not selected or not value then
        local prompt = self.selected_compact
            and "[grey]PICK a non-empty tile to inspect it.[/]"
            or "[grey]Use PICK and click a non-empty floor, wall, or shadow record to inspect and edit it.[/]"
        set_label(
            self.selected_meta,
            prompt,
            self.selected_meta_left,
            self.selected_meta_top,
            self.selected_meta_width,
            "left",
            "font_lab_caption"
        )
        for _, control in ipairs(self.edit_controls or {}) do
            control.box:set_visible(false)
            control.label:set_visible(false)
        end
        for _, node in pairs(self.edit_value_labels or {}) do node:set_visible(false) end
        if self.hidden_control then self.hidden_control:set_visible(false) end
        return
    end
    local orientation = value.orientation or (selected.kind == "shadow" and 13 or 0)
    local metadata
    if self.selected_compact then
        metadata = string.format(
            "[gold]%d,%d %s[/] [white]%02d/%02d/%02d[/]  %s #%d",
            selected.point.x, selected.point.y, string.upper(selected.kind), orientation, value.style or 0,
            value.sequence or 0, file_name(value.source_path), value.source_index or -1
        )
    else
        metadata = string.format(
            "[gold]CELL %d, %d  %s[/]   [white]%02d/%02d/%02d[/]   RAW 0x%08X\n"
                .. "[gold]DT1[/] %s #%d   [grey]%dx%d  R%d D%d H%d[/]\n"
                .. "[grey]DT1 FLAGS  walk %d  LOS %d  jump %d  player %d  light %d[/]",
            selected.point.x, selected.point.y, string.upper(selected.kind), orientation, value.style or 0,
            value.sequence or 0, value.properties or 0, file_name(value.source_path), value.source_index or -1,
            value.width or 0, math.abs(value.height or 0), value.rarity or 0, value.direction or 0,
            value.roof_height or 0, value.blocked_walk or 0, value.blocked_los or 0, value.blocked_jump or 0,
            value.blocked_player or 0, value.blocked_light or 0
        )
    end
    set_label(self.selected_meta, metadata, self.selected_meta_left, self.selected_meta_top,
        self.selected_meta_width, "left", "font_lab_caption")
    for _, control in ipairs(self.edit_controls or {}) do
        control.disabled = control.field == "orientation" and selected.kind ~= "wall"
        control.box:set_visible(true)
        control.label:set_visible(true)
        control.box:set_state(control.disabled and "disabled" or "idle")
    end
    local field_values = {
        orientation=orientation,
        style=value.style or 0,
        sequence=value.sequence or 0,
        prop1=value.prop1 or ((value.properties or 0) % 256),
        unknown1=value.unknown1 or (math.floor((value.properties or 0) / 16384) % 64),
        unknown2=value.unknown2 or (math.floor((value.properties or 0) / 67108864) % 32),
    }
    for field, node in pairs(self.edit_value_labels or {}) do
        node:set_visible(true)
        text.set(node, "font_lab_caption", string.format("%02d", field_values[field] or 0), 30, "center")
    end
    if self.hidden_control then
        self.hidden_control:set_visible(true)
        self.hidden_control:set_label(value.hidden and "HIDDEN ON" or "HIDDEN OFF")
        self.hidden_control:set_state(value.hidden and "selected" or "idle")
    end
end

-- Apply one bounded metadata edit to the selected floor, wall, or shadow record.
function map_editor:edit_selected(field, delta)
    local selected = self.selected_cell
    if not selected then return end
    local value = selected.value
    local orientation = value.orientation or (selected.kind == "shadow" and 13 or 0)
    local style, sequence = value.style or 0, value.sequence or 0
    local properties = value.properties or 0
    if field == "orientation" and selected.kind == "wall" then orientation = clamp(orientation + delta, 0, 255)
    elseif field == "style" then style = clamp(style + delta, 0, 63)
    elseif field == "sequence" then sequence = clamp(sequence + delta, 0, 63)
    elseif field == "prop1" then
        local prop1 = clamp(properties % 256 + delta, 0, 255)
        properties = properties - properties % 256 + prop1
    elseif field == "unknown1" then
        local current = math.floor(properties / 16384) % 64
        local next_value = clamp(current + delta, 0, 63)
        properties = properties + (next_value - current) * 16384
    elseif field == "unknown2" then
        local current = math.floor(properties / 67108864) % 32
        local next_value = clamp(current + delta, 0, 31)
        properties = properties + (next_value - current) * 67108864
    elseif field == "hidden" then
        properties = properties >= 2147483648 and properties - 2147483648 or properties + 2147483648
    else return end
    local brush = {orientation=orientation, style=style, sequence=sequence, properties=properties}
    local ok, result = pcall(editor.begin_stroke, selected.kind, selected.layer, brush)
    if ok then ok, result = pcall(editor.paint, selected.point.x, selected.point.y) end
    local changed, points
    if ok then ok, changed, points = pcall(editor.end_stroke) end
    if not ok then self:set_message(tostring(result or changed), "red"); return end
    if changed then
        local cell = editor.cell(selected.point.x, selected.point.y)
        local values = selected.kind == "floor" and cell.floors
            or (selected.kind == "wall" and cell.walls or cell.shadows)
        selected.value = values[selected.layer + 1]
        self.summary = editor.summary()
        self:select_tile(selected.value)
        self:highlight_selection(selected.point)
        self:refresh_chrome()
        self:queue_preview(false, points)
        self:set_message("Updated selected tile metadata. Undo is available.", "green")
    end
end

-- Convert a native pointer position into a valid DS1 cell coordinate.
function map_editor:point_at(pointer_x, pointer_y)
    if not self.preview_world or not self.chunk_set or not self.summary then return nil end
    local pixel_x = (pointer_x - (canvas_center_x + self.pan_x)) / self.zoom + self.chunk_set.width / 2
    local pixel_y = (pointer_y - (canvas_center_y + self.pan_y)) / self.zoom + self.chunk_set.height / 2
    local subtile_x, subtile_y = self.preview_world:pixel_to_subtile(pixel_x, pixel_y)
    local x, y = math.floor(subtile_x / 5), math.floor(subtile_y / 5)
    if x < 0 or y < 0 or x >= self.summary.width or y >= self.summary.height then return nil end
    return {x=x, y=y}
end

-- Start one undoable stroke and immediately apply its first cell.
function map_editor:begin_paint(point)
    local layer = layers[self.layer_index]
    local brush = self.tool == "erase" and {empty=true} or self.brush
    local ok, result = pcall(editor.begin_stroke, layer.kind, 0, brush, self.auto_draw)
    if not ok then self:set_message(tostring(result), "red"); return false end
    self.stroke_active, self.stroke_seen = true, {}
    self:paint_at(point)
    return true
end

-- Add a distinct DS1 cell to the active stroke without refreshing the preview yet.
function map_editor:paint_at(point)
    if not self.stroke_active or not point then return end
    local key = point.x .. ":" .. point.y
    if self.stroke_seen[key] then return end
    self.stroke_seen[key] = true
    local ok, result = pcall(editor.paint, point.x, point.y)
    if not ok then self:set_message(tostring(result), "red"); self.stroke_active = false end
end

-- Commit the stroke once and queue only its affected preview chunks.
function map_editor:end_paint()
    if not self.stroke_active then return end
    self.stroke_active = false
    local ok, changed, points = pcall(editor.end_stroke)
    if not ok then self:set_message(tostring(changed), "red"); return end
    if changed then
        self.summary = editor.summary()
        self:refresh_chrome()
        self:queue_preview(false, points)
    end
end

-- Save a copy through protected editor storage, never back into mounted MPQs.
function map_editor:save()
    if not self.summary then return end
    local ok, destination = pcall(editor.save, self.save_path)
    if not ok then
        self:set_message(tostring(destination), "red")
        return
    end
    self.summary = editor.summary()
    self:refresh_chrome()
    self:set_message("Saved " .. destination, "green")
end

-- Restore the previous document revision and invalidate its changed cells.
function map_editor:undo()
    if not self.summary then return end
    local changed, points = editor.undo()
    if changed then
        self.summary = editor.summary()
        self:reload_selected_cell()
        self:refresh_chrome()
        self:queue_preview(false, points)
    end
end

-- Reapply the next document revision and invalidate its changed cells.
function map_editor:redo()
    if not self.summary then return end
    local changed, points = editor.redo()
    if changed then
        self.summary = editor.summary()
        self:reload_selected_cell()
        self:refresh_chrome()
        self:queue_preview(false, points)
    end
end

-- Route one primary click through modal chrome before it can reach the map.
function map_editor:click_chrome(pointer_x, pointer_y)
    if self.palette_menu_open then
        for index, button in ipairs(self.palette_options) do
            if button:contains(pointer_x, pointer_y) then
                self:set_palette(index, true)
                self:toggle_palette_menu()
                return true
            end
        end
    end
    if self.palette_button:contains(pointer_x, pointer_y) then
        self:toggle_palette_menu()
        return true
    elseif self.palette_menu_open then
        self:toggle_palette_menu()
    end
    local actions = {
        function() self.asset_picker:show() end,
        function() self:save() end,
        function() self:undo() end,
        function() self:redo() end,
        function() self:set_inspector_mode("library") end,
        function() self:toggle_grid() end,
    }
    for index, action in ipairs(actions) do
        if self.top_buttons[index]:contains(pointer_x, pointer_y) then action(); return true end
    end
    for index, button in ipairs(self.mode_boxes or {}) do
        if button:contains(pointer_x, pointer_y) then
            self:set_inspector_mode(index == 1 and "library" or "selected")
            if index == 2 then self:refresh_selected_view() end
            return true
        end
    end
    if self.auto_button:contains(pointer_x, pointer_y) then
        self:toggle_auto_draw(); return true
    end
    for index, tool in ipairs(tools) do
        if self.tool_boxes[index]:contains(pointer_x, pointer_y) then
            if self.tool == "pick" and tool.id ~= "pick" then self:clear_hover() end
            self.tool = tool.id
            self:refresh_chrome()
            self:set_message(tool.hint, "white")
            return true
        end
    end
    for index, layer in ipairs(visibility_layers) do
        if self.layer_boxes[index]:contains(pointer_x, pointer_y) then
            local visibility = self.inspector_mode == "selected" and self.selected_visibility or self.map_visibility
            visibility[layer.kind] = not visibility[layer.kind]
            if index <= #layers and visibility[layer.kind] then
                self.layer_index = index
                self:refresh_tilesets()
            end
            self:refresh_chrome()
            if self.inspector_mode == "selected" then
                self:refresh_selected_preview()
            else
                self:invalidate_all_chunks()
                self:refresh_visible_chunks()
                self:refresh_collision_nodes()
            end
            return true
        end
    end
    if self.inspector_mode == "library" then
        for row, control in ipairs(self.tileset_rows) do
            local top = layout.tileset_rows + (row - 1) * 32
            local rect = {left=self.inspector_left + 18, top=top, right=self.inspector_right - 18, bottom=top + 30}
            if control.box:exists() and in_rect(pointer_x, pointer_y, rect) then
                self:select_tileset(self.tileset_scroll + row); return true
            end
        end
        for _, control in ipairs(self.tile_cells) do
            if control.choice and in_rect(pointer_x, pointer_y, control.rect) then
                self:select_tile(control.choice); return true
            end
        end
    elseif self.inspector_mode == "selected" then
        for _, control in ipairs(self.edit_controls) do
            if not control.disabled and in_rect(pointer_x, pointer_y, control.rect) then
                self:edit_selected(control.field, control.delta); return true
            end
        end
    end
    return false
end

-- Construct the complete responsive editor scene and its retained control pools.
function map_editor:create()
	editor = require("engine.map_editor/v1")
	local available, capability = pcall(require, "engine.display/v1")
	display = available and capability or nil
	self.surface_width, self.surface_height = surface_size()
	canvas, self.inspector_left, self.inspector_right = layout_canvas(self.surface_width, self.surface_height)
	canvas_center_x = (canvas.left + canvas.right) / 2
    canvas_center_y = (canvas.top + canvas.bottom) / 2
    self.root = render.create("hud")
    box(self.root, {left=0,top=0,right=self.surface_width,bottom=self.surface_height}, {11, 16, 24, 255}, -20)
    box(self.root, {left=canvas.left,top=canvas.top,right=canvas.right,bottom=canvas.bottom}, {6, 10, 15, 255}, -15)
    self.workspace = ui.workspace(self.root, {
        width=self.surface_width,
        height=self.surface_height,
        canvas=canvas,
        inspector_left=self.inspector_left,
        inspector_right=self.inspector_right,
    })

    self.map_root = render.create("hud", self.root)
    -- The canvas background is at -15. Keep map art immediately above it but
    -- beneath inspector chrome; the old -10000 ordering hid every DS1 chunk.
    self.map_root:set_z(-10)
    self.map_root:set_clip(canvas.left, canvas.top, canvas.right - canvas.left, canvas.bottom - canvas.top)
    self.map_root:set_visible(true)
    self.chunk_nodes = {}
    self.title = label(
        self.root,
        "",
        96,
        18,
        self.surface_width - 192,
        "center",
        "font_lab_color",
        40
    )
    self.subtitle = label(
        self.root,
        "",
        self.surface_width - 292,
        layout.document_meta_top,
        260,
        "right",
        "font_lab_caption",
        20
    )
    self.document_meta = label(
        self.root,
        "",
        560,
        layout.document_meta_top,
        math.max(1, self.surface_width - 876),
        "left",
        "font_lab_caption",
        20
    )
    self.status = label(
        self.root,
        "",
        80,
        self.surface_height - 49,
        math.max(240, self.surface_width - 820),
        "left",
        "font_lab_caption",
        20
    )
    self.status_detail = label(
        self.root,
        "",
        self.surface_width - 680,
        self.surface_height - 49,
        560,
        "right",
        "font_lab_caption",
        20
    )

    self.top_buttons = {}
    for index, value in ipairs({"OPEN", "SAVE", "UNDO", "REDO", "TILES", "GRID"}) do
        local left = 88 + (index - 1) * 72
        local rect = {
            left=left,
            top=layout.toolbar_top,
            right=left + 64,
            bottom=layout.toolbar_bottom,
        }
        local top_icon_names = {"open", "save", "undo", "redo", "inspector", "grid"}
        local sheet = index == 5 and "authoring" or "utility"
        local icon_name = index == 5 and "stamp" or top_icon_names[index]
        self.top_buttons[index] = icon_button(self.root, rect, sheet, icon_name, 20)
        if value == "GRID" then
            self.grid_button = self.top_buttons[index]
        end
    end

    self.palette_button = ui.dropdown(self.root, {
        left=616,
        top=layout.toolbar_top + 8,
        label="PALETTE: ACT I",
        text_style="font_lab_caption",
        state="idle",
        z=30,
    })
    self.palette_options = {}
    for index, palette in ipairs(palettes) do
        local button = ui.button(self.root, {
            left=616,
            top=layout.toolbar_bottom + 4 + (index - 1) * 48,
            label=palette.label,
            text_style="font_lab_caption",
            width=self.palette_button.width,
            state=index == 1 and "selected" or "idle",
            z=50,
        })
        button:set_visible(false)
        self.palette_options[index] = button
    end

    self.tool_boxes = {}
    for index, tool in ipairs(tools) do
        local top = layout.tool_top + (index - 1) * 72
        local rect = {left=16, top=top, right=80, bottom=top + 64}
        self.tool_boxes[index] = icon_button(self.root, rect, "authoring", tool.id, 20)
    end

    self.layer_boxes = {}
    for index, layer in ipairs(visibility_layers) do
        local left = self.inspector_left + 16 + (index - 1) * 72
        local rect = {left=left, top=layout.layer_top, right=left + 64, bottom=layout.layer_top + 64}
        self.layer_boxes[index] = icon_button(self.root, rect, layer.sheet, layer.icon, 20)
    end
    self.mode_boxes = {}
    local mode_left = self.inspector_left + 12
    for index, value in ipairs({"LIBRARY", "SELECTED"}) do
        local button = ui.button(self.root, {
            left=mode_left,
            top=layout.mode_top,
            label=value,
            text_style="font_lab_caption",
            skin="tab",
            state="tab",
            z=20,
        })
        self.mode_boxes[index] = button
        mode_left = mode_left + button.width + 4
    end
    local auto_rect = {
        left=self.inspector_left + 12,
        top=layout.auto_top,
        right=self.inspector_left + 188,
        bottom=layout.auto_top + 48,
    }
    self.auto_button = ui.button(self.root, {
        left=auto_rect.left, top=auto_rect.top, width=auto_rect.right - auto_rect.left,
        label="AUTO: OFF", text_style="font_lab_caption", icon_sheet="utility",
        icon_name="auto_draw", state="idle", z=20,
    })

    self.library_root = render.create("hud", self.root)
    self.library_root:set_z(20)
    ui.section(self.library_root, {
        left=self.inspector_left + 12,
        top=layout.tileset_heading - 8,
        width=self.inspector_right - self.inspector_left - 24,
        z=0,
        seed=101,
    })
    label(
        self.library_root,
        "[gold]DT1 TILESETS[/]",
        self.inspector_left + 18,
        layout.tileset_heading,
        self.inspector_right - self.inspector_left - 36,
        "left",
        "font_lab_caption",
        1
    )
    self.tileset_rows = {}
    for row = 1, 4 do
        local top = layout.tileset_rows + (row - 1) * 32
        local rect = {left=self.inspector_left + 18, top=top, right=self.inspector_right - 18, bottom=top + 30}
        self.tileset_rows[row] = {
            box=ui.well(self.library_root, {
                left=rect.left, top=rect.top, width=rect.right - rect.left,
                height=rect.bottom - rect.top, z=1, seed=row,
            }),
            label=label(
                self.library_root,
                "",
                rect.left + 5,
                rect.top + 6,
                rect.right - rect.left - 10,
                "left",
                "font_lab_caption",
                2
            ),
        }
    end
    ui.section(self.library_root, {
        left=self.inspector_left + 12,
        top=layout.physical_heading - 8,
        width=self.inspector_right - self.inspector_left - 24,
        z=0,
        seed=102,
    })
    label(
        self.library_root,
        "[gold]PHYSICAL TILES[/]  [grey]wheel to browse[/]",
        self.inspector_left + 18,
        layout.physical_heading,
        self.inspector_right - self.inspector_left - 36,
        "left",
        "font_lab_caption",
        1
    )
    self.tile_columns = self.inspector_right - self.inspector_left >= 300 and 3 or 2
    local tile_gap, tile_top = 8, layout.tile_top
    local tile_width = math.floor(
        (self.inspector_right - self.inspector_left - 36 - tile_gap * (self.tile_columns - 1))
        / self.tile_columns
    )
    local tile_height = 96
    local tile_rows = math.max(1, math.floor((canvas.bottom - tile_top - 8) / (tile_height + tile_gap)))
    self.tile_cells = {}
    for slot = 1, tile_rows * self.tile_columns do
        local column, row = (slot - 1) % self.tile_columns, math.floor((slot - 1) / self.tile_columns)
        local left = self.inspector_left + 18 + column * (tile_width + tile_gap)
        local top = tile_top + row * (tile_height + tile_gap)
        local rect = {left=left, top=top, right=left + tile_width, bottom=top + tile_height}
        -- Give every thumbnail its own positioned root. DT1 images are local
        -- to the card instead of inheriting a screen-space coordinate through
        -- a shared list root, and an absolute clip keeps tall wall art inside
        -- the recessed card.
        local preview_root = render.create("hud", self.library_root)
        preview_root:set_position((rect.left + rect.right) / 2, rect.top + 36)
        preview_root:set_z(2)
        local preview = render.create("hud", preview_root)
        preview:set_position(0, 0)
        preview:set_clip(rect.left + 4, rect.top + 4, tile_width - 8, tile_height - 22)
        preview:set_z(2)
        self.tile_cells[slot] = {
            box=ui.recess(self.library_root, {
                left=rect.left, top=rect.top, width=rect.right - rect.left,
                height=rect.bottom - rect.top, z=1, seed=slot + 20,
            }),
            preview=preview,
            label=label(
                self.library_root,
                "",
                rect.left + 2,
                rect.bottom - 14,
                tile_width - 4,
                "center",
                "font_lab_caption",
                3
            ),
            preview_root=preview_root,
            preview_width=tile_width - 8,
            preview_height=tile_height - 22,
            rect=rect, width=tile_width, height=tile_height,
        }
    end

    self.selected_root = render.create("hud", self.root)
    self.selected_root:set_z(20)
    local selected_left = self.inspector_left + 18
    local selected_width = self.inspector_right - self.inspector_left - 36
    local selected_top = layout.tileset_heading - 4
    local selected_room = canvas.bottom - selected_top
    self.selected_compact = selected_room < 300
    local hero_height = self.selected_compact and 0 or 72
    local preview_left = self.inspector_left + math.floor((self.inspector_right - self.inspector_left) * 0.52)
    self.selected_preview_rect = {
        left=preview_left + 4,
        top=selected_top + 4,
        right=self.inspector_right - 22,
        bottom=selected_top + math.max(32, hero_height) - 4,
    }
    self.selected_hero = ui.recess(self.selected_root, {
        left=self.inspector_left + 12,
        top=selected_top,
        width=self.inspector_right - self.inspector_left - 24,
        height=math.max(32, hero_height),
        z=1,
        seed=103,
    })
    self.selected_hero:set_visible(not self.selected_compact)
    self.selected_preview_root = render.create("hud", self.selected_root)
    self.selected_preview_root:set_position(
        (self.selected_preview_rect.left + self.selected_preview_rect.right) / 2,
        (self.selected_preview_rect.top + self.selected_preview_rect.bottom) / 2
    )
    self.selected_preview_root:set_z(2)
    self.selected_preview_root:set_visible(not self.selected_compact)
    self.tile_preview = render.create("hud", self.selected_preview_root)
    self.tile_preview:set_position(0, 0)
    self.tile_preview:set_clip(
        self.selected_preview_rect.left,
        self.selected_preview_rect.top,
        self.selected_preview_rect.right - self.selected_preview_rect.left,
        self.selected_preview_rect.bottom - self.selected_preview_rect.top
    )
    self.tile_preview:set_z(2)
    self.selected_meta_left = selected_left
    self.selected_meta_top = self.selected_compact and selected_top or selected_top + hero_height + 4
    self.selected_meta_width = selected_width
    self.selected_meta = label(
        self.selected_root,
        "",
        self.selected_meta_left,
        self.selected_meta_top,
        self.selected_meta_width,
        "left",
        "font_lab_caption",
        2
    )
    local edit_columns = self.selected_compact and 3 or 2
    local edit_rows_count = math.ceil(6 / edit_columns)
    local edit_height = 32 + edit_rows_count * 36 + 36
    local edit_top = canvas.bottom - edit_height
    ui.section(self.selected_root, {
        left=self.inspector_left + 12,
        top=edit_top,
        width=self.inspector_right - self.inspector_left - 24,
        z=1,
        seed=104,
    })
    label(
        self.selected_root,
        "[gold]AUTHORED DS1 FIELDS[/]",
        self.inspector_left + 18,
        edit_top + 8,
        self.inspector_right - self.inspector_left - 36,
        "left",
        "font_lab_caption",
        2
    )
    self.edit_controls = {}
    self.edit_value_labels = {}
    local edit_fields = {
        {"TYPE", "orientation"}, {"STYLE", "style"}, {"SEQ", "sequence"},
        {"PROP1", "prop1"}, {"AUX14", "unknown1"}, {"AUX26", "unknown2"},
    }
    local field_gap = 6
    local field_width = math.floor((selected_width - field_gap * (edit_columns - 1)) / edit_columns)
    for index, value in ipairs(edit_fields) do
        local column = (index - 1) % edit_columns
        local row = math.floor((index - 1) / edit_columns)
        local left = selected_left + column * (field_width + field_gap)
        local top = edit_top + 32 + row * 36
        label(self.selected_root, value[1], left, top + 8, 54, "left", "font_lab_caption", 2)
        self.edit_value_labels[value[2]] = label(
            self.selected_root, "00", left + 52, top + 8, 30, "center", "font_lab_caption", 2
        )
        for _, action in ipairs({{-1, "-"}, {1, "+"}}) do
            local button_left = left + field_width - (action[1] == -1 and 72 or 36)
            local rect = {left=button_left, top=top, right=button_left + 34, bottom=top + 32}
            local button = ui.button(self.selected_root, {
                left=rect.left, top=rect.top, width=rect.right - rect.left,
                label=action[2], text_style="font_lab_caption", skin="well",
                min_width=33, padding_x=4, state="idle", z=2,
            })
            self.edit_controls[#self.edit_controls + 1] = {
                field=value[2], delta=action[1], rect=rect,
                box=button,
                label=button.label,
            }
        end
    end
    local hidden_rect = {
        left=selected_left,
        top=edit_top + 32 + edit_rows_count * 36 + 2,
        right=self.inspector_right - 18,
        bottom=edit_top + 32 + edit_rows_count * 36 + 34,
    }
    local hidden_button = ui.button(self.selected_root, {
        left=hidden_rect.left, top=hidden_rect.top,
        width=hidden_rect.right - hidden_rect.left,
        label="HIDDEN OFF", text_style="font_lab_caption", skin="well", state="idle", z=2,
    })
    self.hidden_control = hidden_button
    self.edit_controls[#self.edit_controls + 1] = {
        field="hidden", delta=1, rect=hidden_rect,
        box=hidden_button,
        label=hidden_button.label,
    }
    -- Kept for the compact brush summary used by refresh_chrome.
    self.brush_meta = label(
        self.selected_root,
        "",
        selected_left + 6,
        selected_top + 8,
        preview_left - selected_left - 12,
        "left",
        "font_lab_caption",
        2
    )
    self.brush_meta:set_visible(not self.selected_compact)

    self.assets = vfs.list("data/global/tiles", ".ds1") or {}
    self.asset_picker = fuzzy_picker.create(self.root, {
        title="OPEN DS1 MAP",
        width=self.surface_width,
        height=self.surface_height,
        items=self.assets,
        on_select=function(path) self:open(path) end,
    })
    self.tile_by_label = {}
    self.tile_picker = fuzzy_picker.create(self.root, {
        title="SELECT LOGICAL TILE",
        width=self.surface_width,
        height=self.surface_height,
        items={},
        on_select=function(value) self:select_tile(self.tile_by_label[value]) end,
    })

    self.path, self.summary, self.chunk_set, self.preview_world = "", nil, nil, nil
    self.visible_signature, self.visible_pending = nil, true
    self.palette, self.palette_index = palettes[1].path, 1
    self.palette_override, self.palette_menu_open = false, false
    self.zoom, self.pan_x, self.pan_y = 1, 0, 0
    self.tool, self.layer_index = "pan", 1
    self.brush = {orientation=0, style=0, sequence=0, properties=0}
    self.tilesets, self.tileset_tiles, self.tileset_scroll, self.tile_scroll = {}, {}, 0, 0
    self.inspector_mode, self.auto_draw = "library", false
    self.map_visibility = {floor=true, wall=true, shadow=true, collision=false}
    self.selected_visibility = {floor=true, wall=true, shadow=true, collision=false}
    self.grid_visible, self.grid_nodes = true, {}
    self.selection_nodes, self.selected_preview_nodes, self.collision_nodes = {}, {}, {}
    self.dragging, self.panning, self.stroke_active = false, false, false
    self:position_map()
    self:set_message("Ready", "white")
    self:set_inspector_mode("library")
    self:refresh_selected_view()
    if map_editor.initial_path and map_editor.initial_path ~= "" then
        self:open(map_editor.initial_path)
    end
    self:refresh_chrome()
end

-- Advance background work and route one frame of keyboard, wheel, and pointer input.
function map_editor:update()
    if self:relayout_if_needed() then return end
    self:finish_preview()
    if self.asset_picker:update() or self.tile_picker:update() then return end
    if input.pressed("search") then self.asset_picker:show(); return end
    if input.pressed("save") then self:save(); return end
    if input.pressed("undo") then self:undo(); return end
    if input.pressed("redo") then self:redo(); return end
    if input.pressed("debug_map_tiles") then self:toggle_grid(); return end
    if input.pressed("home") then self.zoom = quantized_zoom(self.zoom - zoom_step); self:position_map() end
    if input.pressed("end") then self.zoom = quantized_zoom(self.zoom + zoom_step); self:position_map() end
    local keyboard_pan = false
    if input.pressed("left") then self.pan_x = self.pan_x - 32; keyboard_pan = true end
    if input.pressed("right") then self.pan_x = self.pan_x + 32; keyboard_pan = true end
    if input.pressed("up") then self.pan_y = self.pan_y - 32; keyboard_pan = true end
    if input.pressed("down") then self.pan_y = self.pan_y + 32; keyboard_pan = true end
    if keyboard_pan then self:position_map() end

    local pointer_x, pointer_y = input.cursor()
    local pointer_in_canvas = in_rect(pointer_x, pointer_y, canvas)
    local temporary_pan = input.down("space")
    local pointer_in_inspector = pointer_x >= self.inspector_left and pointer_x < self.inspector_right
        and pointer_y >= layout.tileset_heading and pointer_y < canvas.bottom
    if self.tool == "pick" and not temporary_pan and pointer_in_canvas and self.summary then
        self:hover_pick(self:point_at(pointer_x, pointer_y))
    else
        self:clear_hover()
    end
    local _, scroll_y = input.scroll()
    if pointer_in_canvas and scroll_y ~= 0 then
        self:zoom_at(pointer_x, pointer_y, math.exp(scroll_y * 0.12))
    end
    if pointer_in_inspector and self.inspector_mode == "library" and scroll_y ~= 0 then
        if pointer_y < layout.physical_heading then
            local maximum = math.max(0, #self.tilesets - #self.tileset_rows)
            self.tileset_scroll = clamp(self.tileset_scroll - scroll_y, 0, maximum)
            self.tileset_scroll = math.floor(self.tileset_scroll)
            self:refresh_tileset_rows()
        else
            local maximum = math.max(0, #self.tileset_tiles - #self.tile_cells)
            self.tile_scroll = clamp(self.tile_scroll - scroll_y * self.tile_columns, 0, maximum)
            self.tile_scroll = math.floor(self.tile_scroll / self.tile_columns) * self.tile_columns
            self:refresh_tile_grid()
        end
    end

    if input.pressed("pointer_primary") then
        if self:click_chrome(pointer_x, pointer_y) then return end
        if pointer_in_canvas and self.summary then
            local point = self:point_at(pointer_x, pointer_y)
            if self.tool == "pan" or temporary_pan then
                self.dragging, self.drag_x, self.drag_y = true, pointer_x, pointer_y
            elseif self.tool == "pick" then
                if point then self:sample_at(point) end
            elseif point then
                self:begin_paint(point)
            end
        end
    end
    if input.pressed("pointer_secondary") and pointer_in_canvas then
        self.panning, self.drag_x, self.drag_y = true, pointer_x, pointer_y
    end
    if (self.dragging and input.down("pointer_primary")) or (self.panning and input.down("pointer_secondary")) then
        local delta_x, delta_y = pointer_x - self.drag_x, pointer_y - self.drag_y
        if delta_x ~= 0 or delta_y ~= 0 then
            self.pan_x, self.pan_y = self.pan_x + delta_x, self.pan_y + delta_y
            self.drag_x, self.drag_y = pointer_x, pointer_y
            self:position_map()
        end
    elseif self.stroke_active and input.down("pointer_primary") then
        self:paint_at(self:point_at(pointer_x, pointer_y))
    end
    if input.released("pointer_primary") then
        self.dragging = false
        self:end_paint()
    end
    if input.released("pointer_secondary") then self.panning = false end
end

-- Release the editor scene without closing the authoritative Go document session.
function map_editor:destroy()
    self:clear_preview()
    if self.map_root and self.map_root:exists() then self.map_root:destroy() end
    if self.root and self.root:exists() then self.root:destroy() end
end

return map_editor
