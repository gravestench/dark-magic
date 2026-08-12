-- Fixed-profile party panel shell.
--
-- Network/party roster state belongs to the engine/session. This Lua currently
-- draws the correct panel shape and selected hero heading, then clearly marks the
-- missing roster interaction rather than manufacturing fake multiplayer state.

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local saves = require("d2legacy.save/v1")
local data = require("engine.data/v1")
local locale = require("engine.locale/v1")
local controls = require("d2legacy.ui.controls")
local button = require("d2legacy.ui.button")
local text = require("d2legacy.ui.text")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2legacy.presentation/v1"))
local screen = manifest.screens.party
local offset_x, offset_y = screen.offset_x or 0, screen.offset_y or 0

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

        -- Four authored quadrants make the full party panel.
        panel_frame(self.root, panel, panel.frames[1], panel.x + offset_x, panel.y + offset_y)
        panel_frame(self.root, panel, panel.frames[2], panel.x + offset_x + 256, panel.y + offset_y)
        panel_frame(self.root, panel, panel.frames[3], panel.x + offset_x, panel.y + offset_y + 256)
        panel_frame(self.root, panel, panel.frames[4], panel.x + offset_x + 256, panel.y + offset_y + 256)

        -- For now only show selected local character as a heading. A real party
        -- roster should come from a network/session snapshot capability later.
        local hero = saves.selected()
        text.create(self.root, "panel_heading", hero and hero.name or "", screen.heading.x + offset_x, screen.heading.y + offset_y, screen.heading.width)

        text.create(self.root, "disabled", assert(locale.text("d2legacy.party.unavailable")), screen.unavailable.x + offset_x, screen.unavailable.y + offset_y, screen.unavailable.width)

        local close = screen.close
        local close_placement = {
            sheet=close.sheet, palette=close.palette, up_frame=close.up_frame, down_frame=close.down_frame,
            x=close.x + offset_x, y=close.y + offset_y, width=close.width, height=close.height, label=close.label,
        }
        button.create(self.root, self.controls, "close", close_placement, assert(locale.text(close.label)), {
            layer="modal", show_label=false, sound=manifest.sounds.button,
            tooltip=assert(locale.text(close.label)), on_activate=function() scenes.toggle_overlay("party", "full") end,
        })
    end,

    update = function(self)
        self.controls:update()
        if input.pressed("party") or input.pressed("cancel") then scenes.pop() end
    end,
}
