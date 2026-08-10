-- Expansion quest log using recovered quest hierarchy and canonical panel geometry.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local locale = require("dm.locale/v1")
local controls = require("darkmagic.ui.controls")
local button = require("darkmagic.ui.button")
local text = require("darkmagic.ui.text")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.quests

local function dc6(root, sheet, palette, frame, x, y)
    local node = render.create("modal", root)
    local w, h = node:set_dc6(sheet, palette, 0, frame)
    node:set_position(x + w / 2, y + h / 2)
    return node
end

local function translated(key, fallback)
    return locale.text(key) or fallback or key
end

return {
    blocks_update_below = true,
    enter = function(self)
        self.root = render.create("modal")
        self.controls = controls.new()
        if not render.assets_available() then return end
        local panel, palette = screen.panel, manifest.palettes[screen.panel.palette]
        dc6(self.root, panel.sheet, palette, panel.frames[1], panel.x, panel.y)
        dc6(self.root, panel.sheet, palette, panel.frames[2], panel.x + 256, panel.y)
        dc6(self.root, panel.sheet, palette, panel.frames[3], panel.x, panel.y + 256)
        dc6(self.root, panel.sheet, palette, panel.frames[4], panel.x + 256, panel.y + 256)

        for act = 0, 4 do
            dc6(self.root, screen.tabs.sheet, palette, act * 2, screen.tabs.x + act * screen.tabs.step, screen.tabs.y)
        end
        -- Some presentation-only harnesses deliberately omit recovered-data
        -- capabilities and never enter this asset-backed path.
        local catalog = require("dm.quest_catalog/v1")
        local quests = catalog.quests(0)
        for index, socket in ipairs(screen.sockets.positions) do
            dc6(self.root, screen.sockets.sheet, palette, 0, socket.x, socket.y)
            local quest = quests[index]
            if quest then
                text.create(self.root, "formal_small", translated(quest.title_string_key, quest.name), socket.x + 39, socket.y + 57, 90)
            end
        end
        text.create(self.root, "disabled", assert(locale.text("darkmagic.quests.unavailable")), 240, 317, 280)
        local close = screen.close
        button.create(self.root, self.controls, "close", close, assert(locale.text(close.label)), {
            layer="modal", show_label=false, sound=manifest.sounds.button,
            tooltip=assert(locale.text(close.label)), on_activate=function() scenes.pop() end,
        })
    end,
    update = function(self)
        self.controls:update()
        if input.pressed("quests") or input.pressed("cancel") then scenes.pop() end
    end,
}
