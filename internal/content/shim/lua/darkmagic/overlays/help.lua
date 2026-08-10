-- Canonical 800x600 in-game help overlay recovered from OpenDiablo2.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local locale = require("dm.locale/v1")
local controls = require("darkmagic.ui.controls")
local button = require("darkmagic.ui.button")
local text = require("darkmagic.ui.text")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.help

local function frame(root, definition, index, x, y)
    local node = render.create("modal", root)
    local width, height = node:set_dc6(definition.sheet, manifest.palettes[definition.palette], 0, index)
    node:set_position(x + width / 2, y + height / 2)
    return width, height
end

return {
    blocks_update_below = true,

    enter = function(self)
        self.root = render.create("modal")
        self.controls = controls.new()
        if not render.assets_available() then return end

        -- The source sheet is irregular: the first two frames form the left
        -- edge, frames 2-5 form the top, and frame 6 is the lower-right edge.
        local border = screen.border
        local w0, h0 = frame(self.root, border, 0, 0, 0)
        frame(self.root, border, 1, 0, h0)
        local x = w0
        local w2, h2 = frame(self.root, border, 2, x, 0)
        x = x + w2 - border.middle_overlap
        local w3 = select(1, frame(self.root, border, 3, x, 0))
        x = x + w3
        local w4 = select(1, frame(self.root, border, 4, x, 0))
        x = x + w4
        frame(self.root, border, 5, x, 0)
        frame(self.root, border, 6, x, h2)

        text.create(self.root, "panel_heading", assert(locale.text("darkmagic.help.title")), 363, 2, 260)
        local bullets = {"run", "items", "stand", "map", "menu", "chat", "skills", "help"}
        for index, key in ipairs(bullets) do
            local y = 59 + (index - 1) * 20
            local dot = render.create("modal", self.root)
            local dw, dh = dot:set_dc6(screen.yellow_bullet, manifest.palettes.sky, 0, 0)
            dot:set_position(88 + dw / 2, y + 4 + dh / 2)
            text.create(self.root, "formal_large", assert(locale.text("darkmagic.help." .. key)), 100, y, 600, "left")
        end

        local callouts = {
            {"Life Orb",65,451},{"New Stats",222,355},{"Stamina Bar",315,450},
            {"Run/Walk Toggle",264,480},{"Experience Bar",370,476},{"Mini-Panel",450,371},
            {"Belt",535,490},{"New Skill",578,355},{"Mana Orb",745,451},
            {"Left Mouse Button Skill",135,382},{"Right Mouse Button Skill",675,381},
        }
        for _, label in ipairs(callouts) do
            text.create(self.root, "formal_large", label[1], label[2], label[3], 180)
        end

        local close = screen.close
        button.create(self.root, self.controls, "close", close, assert(locale.text(close.label)), {
            layer="modal", show_label=false, sound=manifest.sounds.button,
            tooltip=assert(locale.text(close.label)), on_activate=function() scenes.pop() end,
        })
    end,

    update = function(self)
        self.controls:update()
        if input.pressed("help") or input.pressed("cancel") then scenes.pop() end
    end,
}
