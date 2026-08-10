-- Independently bootable shell for state-dependent in-game interfaces.
-- It deliberately owns presentation only; unavailable game/session operations
-- stay disabled until an authoritative capability is bound.
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

function M.new(definition)
    return {
        blocks_update_below = definition.blocks_update_below ~= false,
        enter = function(self)
            self.root = render.create(definition.layer or "modal")
            self.controls = controls.new()
            local x, y = definition.x or 160, definition.y or 100
            local width, height = definition.width or 480, definition.height or 360

            if definition.sheet and render.assets_available() then
                self.panel = render.create(definition.layer or "modal", self.root)
                width, height = self.panel:set_dc6_combined(
                    definition.sheet,
                    manifest.palettes[definition.palette or "sky"],
                    definition.direction or 0,
                    definition.page or 0)
                self.panel:set_position(x + width / 2, y + height / 2)
            else
                self.panel = render.create(definition.layer or "modal", self.root)
                self.panel:fill_rect(width, height, 18, 14, 10, definition.alpha or 235)
                self.panel:set_position(x + width / 2, y + height / 2)
            end

            if not render.assets_available() then return end
            text.create(self.root, "panel_heading", assert(locale.text(definition.title)),
                x + width / 2, y + 24, width - 40)
            text.create(self.root, "disabled", assert(locale.text(definition.status or "darkmagic.overlay.shell_status")),
                x + width / 2, y + height / 2, width - 60)

            local close = {
                sheet="data/global/ui/PANEL/buysellbtn.DC6", palette="sky",
                up_frame=10, down_frame=11, x=x + width - 48, y=y + height - 48,
                width=32, height=32, label="darkmagic.overlay.close",
            }
            button.create(self.root, self.controls, "close", close, assert(locale.text(close.label)), {
                layer=definition.layer or "modal", show_label=false, sound=manifest.sounds.button,
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
