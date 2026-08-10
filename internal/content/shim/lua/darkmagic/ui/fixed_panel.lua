-- Shared fixed-resolution panel shell for original 800x600 game overlays.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local locale = require("dm.locale/v1")
local controls = require("darkmagic.ui.controls")
local button = require("darkmagic.ui.button")
local text = require("darkmagic.ui.text")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local M = {}

function M.overlay(definition)
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
            local close = {
                sheet="data/global/ui/PANEL/buysellbtn.DC6", palette="sky",
                up_frame=10, down_frame=11,
                x=definition.x + definition.close_x,
                y=definition.y + height - definition.close_y - 32,
                width=32, height=32, label=definition.close_label,
            }
            button.create(self.root, self.controls, "close", close, assert(locale.text(close.label)), {
                layer="modal", show_label=false, sound=manifest.sounds.button,
                tooltip=assert(locale.text(close.label)), on_activate=function() scenes.pop() end,
            })
        end,
        update = function(self)
            self.controls:update()
            if input.pressed("cancel") then scenes.pop() end
        end,
    }
end

return M
