-- Shared profile-selected shell for fixed gameplay panels.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local locale = require("dm.locale/v1")
local controls = require("darkmagic.ui.controls")
local button = require("darkmagic.ui.button")
local text = require("darkmagic.ui.text")
local item_grid = require("darkmagic.ui.item_grid")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local M = {}

local function number(record, key)
    return assert(tonumber(record[key]), "Inventory.txt field is not numeric: " .. key)
end

local function inventory_record(grid)
    local records = require("dm.records/v1")
    for _, row in ipairs(assert(records.load(grid.records))) do
        if row.class == grid.record_class then return row end
    end
    error("Inventory.txt has no record for " .. grid.record_class)
end

function M.overlay(id, slot)
    local definition = assert(manifest.screens[id], "missing presentation screen: " .. id)
    return {
        blocks_update_below = true,
        enter = function(self)
            self.root = render.create("modal")
            self.controls = controls.new()
            if not render.assets_available() then return end
            local palette = manifest.palettes[definition.palette or "sky"]
            self.panel = render.create("modal", self.root)
            local width, height = self.panel:set_dc6_combined(definition.sheet, palette, 0, 0)
            self.panel:set_position(definition.x + width / 2, definition.y + height / 2)

            for _, label in ipairs(definition.labels or {}) do
                text.create(self.root, label.style or "disabled", assert(locale.text(label.key)),
                    definition.x + label.x, definition.y + label.y, label.width or width - 20, label.align)
            end
            if definition.item_grid then
                local grid = definition.item_grid
                local record = inventory_record(grid)
                self.item_grid = item_grid.create(self.root, self.controls, {
                    container = grid.container,
                    columns = number(record, "gridX"),
                    rows = number(record, "gridY"),
                    -- Inventory.txt coordinates are legacy screen coordinates,
                    -- not offsets from invLeft/invTop. The selected profile has
                    -- already moved the whole panel, so add that panel offset.
                    left = definition.x + number(record, "gridLeft"),
                    top = definition.y + number(record, "gridTop"),
                    cell_width = number(record, "gridBoxWidth"),
                    cell_height = number(record, "gridBoxHeight"),
                    palette = manifest.palettes.units,
                })
                self.__darkmagic_item_held = self.item_grid.held
            end
            local close = {
                sheet="data/global/ui/PANEL/buysellbtn.DC6", palette="sky",
                up_frame=10, down_frame=11,
                x=definition.x + definition.close.x,
                y=definition.y + height - definition.close.bottom_inset - 32,
                width=32, height=32, label=definition.close.label,
            }
            button.create(self.root, self.controls, "close", close, assert(locale.text(close.label)), {
                layer="modal", show_label=false, sound=manifest.sounds.button,
                tooltip=assert(locale.text(close.label)), on_activate=function() scenes.toggle_overlay(id, slot) end,
            })
        end,
        update = function(self)
            -- Controls may close this overlay immediately. Never touch its item
            -- nodes after dispatching a button that can destroy their scope.
            if self.item_grid then
                item_grid.update(self.item_grid)
                self.__darkmagic_item_held = self.item_grid.held
            end
            self.controls:update()
            if input.pressed("cancel") then scenes.pop() end
        end,
    }
end

return M
