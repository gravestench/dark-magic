-- Player-selectable cinematic browser.
--
-- Availability is derived from the layered VFS, so missing localized movies
-- disable only their corresponding controls instead of failing the whole scene.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local locale = require("dm.locale/v1")
local vfs = require("dm.vfs/v1")
local video = require("dm.video/v1")
local controls = require("darkmagic.ui.controls")
local label_button = require("darkmagic.ui.label_button")
local cursor = require("darkmagic.ui.cursor")
local dc6 = require("darkmagic.ui.dc6")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.cinematics

return {
    create = function(self)
        self.root = render.create("hud")
        self.background = dc6.frontend_background(
            self.root,
            "hud",
            screen.background,
            manifest.palettes[screen.palette],
            manifest.layouts.frontend_tiles
        )
        self.controls = controls.new()
        self.entries = {}
        for index, definition in ipairs(screen.entries) do
            local label_text = assert(locale.text(definition.label))
            local y = screen.list.y + (index - 1) * screen.list.row_height
            local control_definition = {
                id = definition.id,
                x = screen.list.x,
                y = y,
                width = screen.list.width,
                height = screen.list.row_height,
            }
            local control = label_button.create(self.root, self.controls, control_definition, label_text, {
                enabled = vfs.source(definition.path) ~= nil,
                on_activate = function()
                    if not video.available() then
                        return
                    end
                    local ok, playback = pcall(video.play, definition.path)
                    if ok then
                        self.playback = playback
                    end
                end,
            })
            self.entries[#self.entries + 1] = control
        end
        self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)
    end,
    update = function(self)
        if self.playback then
            if input.pressed("skip") then
                self.playback:stop()
                self.playback = nil
                return
            end
            local status = self.playback:status()
            if status.state ~= "playing" then
                self.playback = nil
            end
            return
        end
        self.controls:update()
        self.cursor:update()
        if input.pressed("cancel") then
            scenes.replace("main_menu")
        end
    end,
}
