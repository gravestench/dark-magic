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

return {
    blocks_update_below = true,

    enter = function(self)
        self.root = render.create("modal")
        self.controls = controls.new()
        if not render.assets_available() then
            return
        end

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
        for _, slot in ipairs(screen.slots) do
            self.controls:add({
                id = "equipment_" .. slot.id,
                label = assert(locale.text(slot.label)),
                x = number(self.record, slot.prefix .. "Left"),
                y = number(self.record, slot.prefix .. "Top"),
                width = number(self.record, slot.prefix .. "Width"),
                height = number(self.record, slot.prefix .. "Height"),
            })
        end

        -- The item grid is registered cell-by-cell for deterministic focus,
        -- pointer hit testing, and future authoritative item placement.
        local columns = number(self.record, "gridX")
        local rows = number(self.record, "gridY")
        local left = number(self.record, "gridLeft")
        local top = number(self.record, "gridTop")
        local cell_width = number(self.record, "gridBoxWidth")
        local cell_height = number(self.record, "gridBoxHeight")
        for row = 0, rows - 1 do
            for column = 0, columns - 1 do
                self.controls:add({
                    id = string.format("inventory_%d_%d", column, row),
                    label = string.format("Inventory %d, %d", column + 1, row + 1),
                    x = left + column * cell_width,
                    y = top + row * cell_height,
                    width = cell_width,
                    height = cell_height,
                })
            end
        end

        local close = screen.close
        button.create(self.root, self.controls, "close", close, assert(locale.text(close.label)), {
            layer = "modal",
            show_label = false,
            sound = manifest.sounds.button,
            tooltip = assert(locale.text(close.label)),
            on_activate = function()
                scenes.pop()
            end,
        })
    end,

    update = function(self)
        self.controls:update()
        if input.pressed("inventory") or input.pressed("cancel") then
            scenes.pop()
        end
    end,
}
