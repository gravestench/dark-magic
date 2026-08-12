-- Player-selectable cinematic browser.
--
-- This scene demonstrates graceful OPTIONAL CONTENT. The layered VFS decides
-- whether each movie path exists. A missing localized movie disables only that
-- row; it does not crash the entire browser.
--
-- It also demonstrates modal-ish state inside ONE scene: while a movie playback
-- handle exists, menu controls/cursor stop updating and video input owns the flow.

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local data = require("engine.data/v1")
local locale = require("engine.locale/v1")
local vfs = require("engine.vfs/v1")
local video = require("engine.video/v1")
local controls = require("d2.ui.controls")
local label_button = require("d2.ui.label_button")
local cursor = require("d2.ui.cursor")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2.presentation/v1"))
local screen = manifest.screens.cinematics

return {
    create = function(self)
        self.root = render.create("hud")
        self.background = render.create("hud", self.root)

        if render.assets_available() then
            local width, height = self.background:set_dc6_combined(
                screen.background,
                manifest.palettes[screen.palette],
                0,
                0
            )
            self.background:set_position(screen.x + width / 2, screen.y + height / 2)
            self.background:set_z(-100)
        end

        self.controls = controls.new()
        self.entries = {}

        for index, definition in ipairs(screen.entries) do
            local label_text = assert(locale.text(definition.label))

            -- Each row's Y follows the familiar start + index*step pattern.
            local y = screen.list.y + (index - 1) * screen.list.row_height
            local control_definition = {
                id = definition.id,
                x = screen.list.x,
                y = y,
                width = screen.list.width,
                height = screen.list.row_height,
            }

            local control = label_button.create(self.root, self.controls, control_definition, label_text, {
                -- VFS source lookup is a capability query. Lua does not inspect
                -- host directories/MPQ internals to decide whether a movie exists.
                enabled = vfs.source(definition.path) ~= nil,

                on_activate = function()
                    if not video.available() then
                        return
                    end

                    -- Movie start can fail for optional codec/content reasons. A
                    -- protected call prevents one unavailable row from killing UI.
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
            -- While playback exists, the menu below is deliberately NOT updated.
            if input.pressed("skip") then
                self.playback:stop()
                self.playback = nil
                return
            end

            local status = self.playback:status()
            if status.state ~= "playing" then
                -- Dropping our checked handle returns this scene to browser mode.
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
