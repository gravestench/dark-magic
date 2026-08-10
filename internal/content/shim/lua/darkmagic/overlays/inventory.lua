-- Character inventory overlay.
--
-- Inventory.txt owns panel, grid, and equipment geometry. The presentation
-- manifest owns only the DC6 quadrant order and expansion-specific corrections.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local saves = require("dm.save/v1")
local data = require("dm.data/v1")
local locale = require("dm.locale/v1")
local controls = require("darkmagic.ui.controls")
local button = require("darkmagic.ui.button")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.inventory
local offset_x, offset_y = screen.offset_x or 0, screen.offset_y or 0

local function number(record, key)
    return assert(tonumber(record[key]), "Inventory.txt field is not numeric: " .. key)
end

local function inventory_record(class)
    -- Records are required lazily because headless presentation harnesses may
    -- intentionally omit asset/record capabilities and never enter this path.
    local records = require("dm.records/v1")
    local rows = assert(records.load(screen.records))
    local wanted = class .. screen.record_suffix
    for _, row in ipairs(rows) do
        if row.class == wanted then
            return row
        end
    end
    error("Inventory.txt has no expansion record for " .. class)
end

local function dc6_at(root, sheet, palette, frame, x, y)
    local node = render.create("modal", root)
    local width, height = node:set_dc6(sheet, palette, 0, frame)
    node:set_position(x + width / 2, y + height / 2)
    return node, width, height
end

local function item_at(snapshot, column, row)
    for _, item in ipairs(snapshot.items) do
        if item.container == "inventory"
            and column >= item.x and column < item.x + item.width
            and row >= item.y and row < item.y + item.height then
            return item
        end
    end
end

local function held_item(snapshot)
    for _, item in ipairs(snapshot.items) do
        if item.container == "held" then return item end
    end
end

local function equipment_item(snapshot, body_loc)
    for _, item in ipairs(snapshot.items) do
        if item.container == "equipment" and item.slot == body_loc then return item end
    end
end

local function refresh_items(self)
    local snapshot = assert(self.items.snapshot())
    local cursor_x, cursor_y = input.cursor()
    self.item_snapshot = snapshot
    self.__darkmagic_item_held = held_item(snapshot) ~= nil
    for _, item in ipairs(snapshot.items) do
        local drawing = self.item_nodes[item.id]
        if drawing == nil and item.inventory_dc6 ~= "" then
            local node = render.create("modal", self.root)
            local width, height = node:set_dc6(item.inventory_dc6, manifest.palettes.units, 0, 0)
            drawing = { node = node, width = width, height = height }
            self.item_nodes[item.id] = drawing
        end
        if drawing ~= nil then
            local equipment = item.container == "equipment" and self.equipment_slots[item.slot] or nil
            local visible = item.container == "inventory" or item.container == "held" or equipment ~= nil
            drawing.node:set_visible(visible)
            if item.container == "inventory" then
                drawing.node:set_position(
                    self.grid_left + item.x * self.cell_width + drawing.width / 2,
                    self.grid_top + item.y * self.cell_height + drawing.height / 2
                )
            elseif item.container == "held" then
                -- The held container is authoritative even though its picture
                -- follows the local pointer. Reconnecting does not lose it.
                drawing.node:set_position(cursor_x, cursor_y)
            elseif equipment ~= nil then
                -- Inventory.txt gives the whole well. Diablo II centers the
                -- item's front-facing DC6 inside it instead of stretching it.
                drawing.node:set_position(
                    equipment.x + equipment.width / 2,
                    equipment.y + equipment.height / 2
                )
            end
        end
    end
end


local function activate_equipment(self, body_loc)
    local held = held_item(self.item_snapshot)
    if held ~= nil then
        self.items.move(held.id, { container = "equipment", slot = body_loc }, true)
        return
    end
    local item = equipment_item(self.item_snapshot, body_loc)
    if item ~= nil then self.items.move(item.id, { container = "held" }) end
end

local function activate_cell(self, column, row)
    local held = held_item(self.item_snapshot)
    if held ~= nil then
        self.items.move(held.id, { container = "inventory", x = column, y = row }, true)
        return
    end
    local item = item_at(self.item_snapshot, column, row)
    if item ~= nil then self.items.move(item.id, { container = "held" }) end
end

return {
    blocks_update_below = true,

    enter = function(self)
        self.root = render.create("modal")
        self.controls = controls.new()
        if not render.assets_available() then
            return
        end

        -- Item authority is a gameplay capability. Presentation-only test
        -- harnesses do not install it, so ask for it only after we know this
        -- overlay can actually build its asset-backed UI.
        self.items = require("dm.items/v1")

        local selected = assert(saves.selected(), "inventory requires a selected character")
        self.record = inventory_record(selected.class)
        local panel = screen.panel
        local palette = manifest.palettes[panel.palette]
        local origin_x = number(self.record, "invLeft") + panel.origin_x_correction
        local origin_y = panel.origin_y_correction
        local positions = {
            { x = origin_x, y = origin_y },
            { x = origin_x + 256, y = origin_y },
            { x = origin_x + 256, y = origin_y + 256 },
            { x = origin_x, y = origin_y + 256 },
        }
        for index, frame in ipairs(panel.frames) do
            local position = positions[index]
            dc6_at(self.root, panel.sheet, palette, frame, position.x, position.y)
        end

        -- Equipment hit regions come directly from the selected class record.
        self.equipment_slots = {}
        for _, slot in ipairs(screen.slots) do
            local body_loc = assert(slot.body_loc, "inventory slot requires body_loc")
            local geometry = {
                x = number(self.record, slot.prefix .. "Left"),
                y = number(self.record, slot.prefix .. "Top"),
                width = number(self.record, slot.prefix .. "Width"),
                height = number(self.record, slot.prefix .. "Height"),
            }
            self.equipment_slots[body_loc] = geometry
            self.controls:add({
                id = "equipment_" .. slot.id,
                label = assert(locale.text(slot.label)),
                x = geometry.x,
                y = geometry.y,
                width = geometry.width,
                height = geometry.height,
                on_activate = function() activate_equipment(self, body_loc) end,
            })
        end

        -- The item grid is registered cell-by-cell for deterministic focus,
        -- pointer hit testing, and future authoritative item placement.
        local columns = number(self.record, "gridX")
        local rows = number(self.record, "gridY")
        self.grid_left = number(self.record, "gridLeft")
        self.grid_top = number(self.record, "gridTop")
        self.cell_width = number(self.record, "gridBoxWidth")
        self.cell_height = number(self.record, "gridBoxHeight")
        self.item_nodes = {}
        for row = 0, rows - 1 do
            for column = 0, columns - 1 do
                -- Each callback gets its own coordinates. Think of these as
                -- little labels glued to the cell instead of shared loop pens.
                local cell_column, cell_row = column, row
                self.controls:add({
                    id = string.format("inventory_%d_%d", column, row),
                    label = string.format("Inventory %d, %d", column + 1, row + 1),
                    x = self.grid_left + column * self.cell_width,
                    y = self.grid_top + row * self.cell_height,
                    width = self.cell_width,
                    height = self.cell_height,
                    on_activate = function() activate_cell(self, cell_column, cell_row) end,
                })
            end
        end

        -- Item positions are authoritative grid cells. The panel converts those
        -- cells to pixels and draws the separate front-facing inventory artwork.
        refresh_items(self)

        local close = screen.close
        local close_placement = {
            sheet=close.sheet, palette=close.palette, up_frame=close.up_frame, down_frame=close.down_frame,
            x=close.x + offset_x, y=close.y + offset_y, width=close.width, height=close.height, label=close.label,
        }
        button.create(self.root, self.controls, "close", close_placement, assert(locale.text(close.label)), {
            layer = "modal",
            show_label = false,
            sound = manifest.sounds.button,
            tooltip = assert(locale.text(close.label)),
            on_activate = function()
                scenes.toggle_overlay("inventory", "right")
            end,
        })
    end,

    update = function(self)
        -- Refresh before controls: a close-button activation may synchronously
        -- destroy this overlay and every render node it owns.
        if self.items ~= nil then refresh_items(self) end
        self.controls:update()
        if input.pressed("inventory") or input.pressed("cancel") then
            scenes.pop()
        end
    end,
}
