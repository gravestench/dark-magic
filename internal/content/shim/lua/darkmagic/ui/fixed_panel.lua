-- Shared profile-selected factory for fixed gameplay panels.
--
-- Stash, Cube, hireling equipment, vendor, and waypoint panels can share a lot:
-- draw one authored background, add some labels, optionally attach an item grid
-- or named item slots, then add a close button.
--
-- This file turns that repeated pattern into ONE factory. Feature modules such as
-- `overlays/stash.lua` can therefore be tiny adapters instead of copy/paste.

local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local locale = require("dm.locale/v1")
local controls = require("darkmagic.ui.controls")
local button = require("darkmagic.ui.button")
local text = require("darkmagic.ui.text")
local item_grid = require("darkmagic.ui.item_grid")
local item_slots = require("darkmagic.ui.item_slots")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local M = {}

-- Records loaded from Diablo II text tables are strings. Convert one required
-- numeric field and fail at the exact field name if the source data is invalid.
local function number(record, key)
    return assert(tonumber(record[key]), "Inventory.txt field is not numeric: " .. key)
end

local function inventory_record(grid)
    -- Records are loaded lazily. A headless harness that never enters this
    -- asset-backed item path therefore does not need dm.records/v1 installed.
    local records = require("dm.records/v1")

    for _, row in ipairs(assert(records.load(grid.records))) do
        if row.class == grid.record_class then return row end
    end

    error("Inventory.txt has no record for " .. grid.record_class)
end

-- Build and return one complete scene definition from a manifest screen ID.
function M.overlay(id, slot)
    -- The manifest supplies all feature-specific presentation facts.
    local definition = assert(manifest.screens[id], "missing presentation screen: " .. id)

    return {
        -- These fixed panels are modal with respect to lower scene update unless
        -- another decorator explicitly changes that policy.
        blocks_update_below = true,

        enter = function(self)
            -- All retained nodes created below inherit this scene-owned root.
            self.root = render.create("modal")
            self.controls = controls.new()

            -- Keep the Lua scene bootable in test environments with no game art.
            if not render.assets_available() then return end

            local palette = manifest.palettes[definition.palette or "sky"]

            -- `set_dc6_combined` assembles a multi-frame DC6 page into one retained
            -- surface and returns its actual decoded size.
            self.panel = render.create("modal", self.root)
            local width, height = self.panel:set_dc6_combined(definition.sheet, palette, 0, 0)

            -- Manifest x/y are the panel TOP-LEFT. Retained position is CENTER.
            self.panel:set_position(definition.x + width / 2, definition.y + height / 2)

            -- Optional labels are data-driven. The loop does not know whether a
            -- particular panel has zero labels or ten.
            for _, label in ipairs(definition.labels or {}) do
                text.create(
                    self.root,
                    label.style or "disabled",
                    assert(locale.text(label.key)),
                    definition.x + label.x,
                    definition.y + label.y,
                    label.width or width - 20,
                    label.align
                )
            end

            if definition.item_grid then
                local grid = definition.item_grid
                local record = inventory_record(grid)

                -- Convert Inventory.txt authority/geometry facts into the small
                -- definition expected by the generic item-grid presenter.
                self.item_grid = item_grid.create(self.root, self.controls, {
                    container = grid.container,
                    columns = number(record, "gridX"),
                    rows = number(record, "gridY"),

                    -- These coordinates are a subtle legacy gotcha. gridLeft /
                    -- gridTop are old SCREEN coordinates, not offsets from
                    -- invLeft/invTop. The current presentation profile already
                    -- moved the panel, so add that selected panel offset here.
                    left = definition.x + number(record, "gridLeft"),
                    top = definition.y + number(record, "gridTop"),
                    cell_width = number(record, "gridBoxWidth"),
                    cell_height = number(record, "gridBoxHeight"),
                    palette = manifest.palettes.units,
                })

                -- Cursor decorator reads this tiny copied flag. If an item is
                -- held, the item image itself is the pointer and the hand hides.
                self.__darkmagic_item_held = self.item_grid.held
            end

            if definition.item_slots then
                local slots = definition.item_slots
                local record = inventory_record(slots)
                local geometry = {}

                -- Inventory.txt names equipment-well coordinates with field
                -- prefixes such as HeadLeft/HeadTop/etc. Convert those records to
                -- a semantic body-location -> rectangle table for item_slots.lua.
                for _, slot_definition in ipairs(slots.slots) do
                    geometry[slot_definition.body_loc] = {
                        x = definition.x + number(record, slot_definition.prefix .. "Left"),
                        y = definition.y + number(record, slot_definition.prefix .. "Top"),
                        width = number(record, slot_definition.prefix .. "Width"),
                        height = number(record, slot_definition.prefix .. "Height"),
                        placeholder = slot_definition.placeholder,
                    }
                end

                self.item_slots = item_slots.create(self.root, self.controls, {
                    container = slots.container,
                    slots = geometry,
                    palette = manifest.palettes.units,
                })
                self.__darkmagic_item_held = self.item_slots.held
            end

            -- Build the tiny X button definition from common art plus this
            -- panel's authored placement facts.
            local close = {
                sheet="data/global/ui/PANEL/buysellbtn.DC6",
                palette="sky",
                up_frame=10,
                down_frame=11,
                x=definition.x + definition.close.x,
                -- `bottom_inset` is measured upward from the decoded panel bottom.
                y=definition.y + height - definition.close.bottom_inset - 32,
                width=32,
                height=32,
                label=definition.close.label,
            }

            button.create(self.root, self.controls, "close", close, assert(locale.text(close.label)), {
                layer="modal",
                show_label=false,
                sound=manifest.sounds.button,
                tooltip=assert(locale.text(close.label)),
                -- Use the same ID/slot pair that opened this generated overlay.
                on_activate=function() scenes.toggle_overlay(id, slot) end,
            })
        end,

        update = function(self)
            -- ORDER MATTERS HERE.
            --
            -- A control callback may close this scene synchronously. Closing the
            -- scope destroys the render handles it owns. Therefore refresh item
            -- presentation BEFORE dispatching controls; never touch those nodes
            -- afterward if a close button might have run.
            if self.item_grid then
                item_grid.update(self.item_grid)
                self.__darkmagic_item_held = self.item_grid.held
            end

            if self.item_slots then
                item_slots.update(self.item_slots)
                self.__darkmagic_item_held = self.item_slots.held
            end

            self.controls:update()

            -- Generic stack close path for Escape/Cancel.
            if input.pressed("cancel") then scenes.pop() end
        end,
    }
end

return M
