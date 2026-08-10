-- Character inventory overlay.
--
-- This is one of the BEST files to study after controls.lua.
-- It joins several parts of the modding API without letting presentation steal
-- ownership from gameplay:
--
--   Inventory.txt        -> authoritative/recovered geometry facts
--   presentation manifest -> panel frame order + profile corrections
--   dm.items/v1 snapshot  -> which item is actually where right now
--   dm.items/v1 move      -> player MOVE INTENT sent back to authority
--   controls.lua          -> clickable cells/equipment wells
--   render nodes          -> disposable pictures of that state
--
-- The big rule is: **drawing an item is not owning an item**.
-- Lua may position its picture under the cursor, but the item is only "held"
-- because authoritative state says its container is `held`.

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

-- Some presentation profiles need small corrections to recovered 800x600 facts.
local offset_x, offset_y = screen.offset_x or 0, screen.offset_y or 0

-- Diablo II TXT table fields arrive as strings. Convert required numeric geometry
-- once and fail at the exact field name when malformed.
local function number(record, key)
    return assert(tonumber(record[key]), "Inventory.txt field is not numeric: " .. key)
end

local function inventory_record(class)
    -- `dm.records/v1` is required lazily because headless presentation harnesses
    -- may intentionally omit recovered-record capabilities and never enter this
    -- real-asset inventory path.
    local records = require("dm.records/v1")
    local rows = assert(records.load(screen.records))

    -- Expansion inventory records use a class name plus a manifest suffix.
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
    -- x/y describe top-left placement; retained node uses image center.
    node:set_position(x + width / 2, y + height / 2)
    return node, width, height
end

-- Find whichever inventory item footprint covers one logical grid cell.
local function item_at(snapshot, column, row)
    for _, item in ipairs(snapshot.items) do
        if item.container == "inventory"
            and column >= item.x and column < item.x + item.width
            and row >= item.y and row < item.y + item.height then
            return item
        end
    end
end

-- `held` is a real authoritative container, not just visual mouse state.
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
    -- Every frame begins from a fresh value snapshot of authoritative item state.
    local snapshot = assert(self.items.snapshot())
    local cursor_x, cursor_y = input.cursor()

    -- Click callbacks use exactly this snapshot until the next refresh.
    self.item_snapshot = snapshot

    -- Cursor decorator reads this copied fact to decide whether the normal hand
    -- should hide while held item art acts as the pointer.
    self.__darkmagic_item_held = held_item(snapshot) ~= nil

    for _, item in ipairs(snapshot.items) do
        local drawing = self.item_nodes[item.id]

        if drawing == nil and item.inventory_dc6 ~= "" then
            -- Lazily create each item's presentation node once. Stable item ID is
            -- the cache key; movement later reuses the same node.
            local node = render.create("modal", self.root)
            local width, height = node:set_dc6(item.inventory_dc6, manifest.palettes.units, 0, 0)
            drawing = { node = node, width = width, height = height }
            self.item_nodes[item.id] = drawing
        end

        if drawing ~= nil then
            -- Compact conditional: only equipment items have named well geometry.
            local equipment = item.container == "equipment" and self.equipment_slots[item.slot] or nil

            -- This inventory overlay displays backpack items, held items, and the
            -- equipment slots it knows how to position. Other containers remain hidden.
            local visible = item.container == "inventory" or item.container == "held" or equipment ~= nil
            drawing.node:set_visible(visible)

            if item.container == "inventory" then
                -- Convert logical item grid origin into pixel top-left, then to
                -- retained image center. Large item footprints need no special
                -- visual case because the DC6 itself supplies its pixel dimensions.
                drawing.node:set_position(
                    self.grid_left + item.x * self.cell_width + drawing.width / 2,
                    self.grid_top + item.y * self.cell_height + drawing.height / 2
                )
            elseif item.container == "held" then
                -- Authority says it is held; only its PICTURE follows local pointer.
                -- Because held state is authoritative, reconnecting does not lose it.
                drawing.node:set_position(cursor_x, cursor_y)
            elseif equipment ~= nil then
                -- Inventory.txt gives the whole equipment well. Diablo II centers
                -- front-facing inventory art inside it rather than stretching it.
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
        -- Submit intent: "try to put held item in this equipment body location."
        -- `true` asks authority for the allowed atomic swap behavior if occupied.
        self.items.move(held.id, { container = "equipment", slot = body_loc }, true)
        return
    end

    -- Empty hand: clicking occupied equipment requests moving that item to held.
    local item = equipment_item(self.item_snapshot, body_loc)
    if item ~= nil then self.items.move(item.id, { container = "held" }) end
end

local function activate_cell(self, column, row)
    local held = held_item(self.item_snapshot)

    if held ~= nil then
        -- Submit intended destination cell; fit/overlap/swap legality is NOT
        -- calculated here. The authority receives the request and decides.
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

        -- Headless tests can still instantiate lifecycle/control state without art.
        if not render.assets_available() then
            return
        end

        -- Item authority is gameplay capability. Require it only in the real
        -- asset-backed path where this panel will actually interact with items.
        self.items = require("dm.items/v1")

        -- Selected character chooses the correct Inventory.txt class geometry.
        local selected = assert(saves.selected(), "inventory requires a selected character")
        self.record = inventory_record(selected.class)

        local panel = screen.panel
        local palette = manifest.palettes[panel.palette]

        -- Inventory.txt's invLeft contributes to original panel position; the
        -- manifest owns known expansion-specific correction terms.
        local origin_x = number(self.record, "invLeft") + panel.origin_x_correction
        local origin_y = panel.origin_y_correction

        -- The inventory panel is four 256-ish quadrant frames. This table names
        -- their exact top-left placement before the generic dc6_at conversion.
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

        -- Equipment hit regions come DIRECTLY from selected class Inventory.txt.
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
                -- `body_loc` is a local created for this loop iteration, so the
                -- callback remembers the correct slot later.
                on_activate = function() activate_equipment(self, body_loc) end,
            })
        end

        -- Backpack grid dimensions/coordinates also come from Inventory.txt.
        local columns = number(self.record, "gridX")
        local rows = number(self.record, "gridY")
        self.grid_left = number(self.record, "gridLeft")
        self.grid_top = number(self.record, "gridTop")
        self.cell_width = number(self.record, "gridBoxWidth")
        self.cell_height = number(self.record, "gridBoxHeight")
        self.item_nodes = {}

        -- Register each cell as an ordinary control. Zero-based columns/rows
        -- match authoritative item placement coordinates.
        for row = 0, rows - 1 do
            for column = 0, columns - 1 do
                -- Closure lesson: create per-iteration copies before building the
                -- callback. Think of these as labels glued permanently to this cell.
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

        -- Draw current authoritative snapshot immediately before first update.
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
        -- ORDER IS A LIFETIME RULE, NOT STYLE.
        --
        -- A close-button callback can synchronously close this scene and destroy
        -- every render handle owned by its scope. Refresh items BEFORE controls;
        -- do not touch item nodes after control dispatch if close may have run.
        if self.items ~= nil then refresh_items(self) end
        self.controls:update()

        if input.pressed("inventory") or input.pressed("cancel") then
            scenes.pop()
        end
    end,
}
