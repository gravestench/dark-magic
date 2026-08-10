-- Fixed 800x600 party panel. Network roster state remains engine-owned.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local saves = require("dm.save/v1")
local data = require("dm.data/v1")
local locale = require("dm.locale/v1")
local controls = require("darkmagic.ui.controls")
local button = require("darkmagic.ui.button")
local text = require("darkmagic.ui.text")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.party

local function panel_frame(root, panel, frame_index, x, y)
    local node = render.create("modal", root)
    local w, h = node:set_dc6(panel.sheet, manifest.palettes[panel.palette], 0, frame_index)
    node:set_position(x + w / 2, y + h / 2)
end

return {
    blocks_update_below = true,
    enter = function(self)
        self.root = render.create("modal")
        self.controls = controls.new()
        if not render.assets_available() then return end
        local panel = screen.panel
        panel_frame(self.root, panel, panel.frames[1], panel.x, panel.y)
        panel_frame(self.root, panel, panel.frames[2], panel.x + 256, panel.y)
        panel_frame(self.root, panel, panel.frames[3], panel.x, panel.y + 256)
        panel_frame(self.root, panel, panel.frames[4], panel.x + 256, panel.y + 256)
        local hero = saves.selected()
        text.create(self.root, "panel_heading", hero and hero.name or "", 180, 80, 180)
        text.create(self.root, "disabled", assert(locale.text("darkmagic.party.unavailable")), 240, 145, 280)
        local close = screen.close
        button.create(self.root, self.controls, "close", close, assert(locale.text(close.label)), {
            layer="modal", show_label=false, sound=manifest.sounds.button,
            tooltip=assert(locale.text(close.label)), on_activate=function() scenes.pop() end,
        })
    end,
    update = function(self)
        self.controls:update()
        if input.pressed("party") or input.pressed("cancel") then scenes.pop() end
    end,
}
