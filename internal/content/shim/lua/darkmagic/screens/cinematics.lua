local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local locale = require("dm.locale/v1")
local vfs = require("dm.vfs/v1")
local video = require("dm.video/v1")
local controls = require("darkmagic.ui.controls")
local cursor = require("darkmagic.ui.cursor")
local dc6 = require("darkmagic.ui.dc6")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen, font = manifest.screens.cinematics, manifest.fonts.exocet10

return {
    create = function(self)
        self.root = render.create("hud")
        self.background = dc6.frontend_background(self.root, "hud", screen.background,
            manifest.palettes[screen.palette], manifest.layouts.frontend_tiles)
        self.controls = controls.new()
        self.entries = {}
        for index, definition in ipairs(screen.entries) do
            local label_text = assert(locale.text(definition.label))
            local y = screen.list.y + (index - 1) * screen.list.row_height
            local control = {
                id = definition.id,
                label = label_text,
                x = screen.list.x,
                y = y,
                width = screen.list.width,
                height = screen.list.row_height,
                enabled = vfs.source(definition.path) ~= nil,
                on_activate = function()
                    if not video.available() then return end
                    local ok, playback = pcall(video.play, definition.path)
                    if ok then self.playback = playback end
                end,
            }
            if render.assets_available() then
                local label = render.create("hud", self.root)
                local function draw(state)
                    local focused = state == "focused" or state == "hover"
                    label:set_text(font.table, font.sheet, manifest.palettes[font.palette], label_text, {
                        red = focused and 235 or 165,
                        green = focused and 205 or 145,
                        blue = focused and 125 or 90,
                        max_width = screen.list.width,
                        align = "center",
                    })
                end
                draw("normal")
                label:set_position(screen.list.x + screen.list.width / 2, y + screen.list.row_height / 2)
                control.on_state = function(_, state) draw(state) end
            end
            self.entries[#self.entries + 1] = self.controls:add(control)
        end
        self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)
    end,
    update = function(self)
        if self.playback then
            if input.pressed("confirm") or input.pressed("cancel") then
                self.playback:stop()
                self.playback = nil
                return
            end
            local status = self.playback:status()
            if status.state ~= "playing" then self.playback = nil end
            return
        end
        self.controls:update()
        self.cursor:update()
        if input.pressed("cancel") then scenes.replace("main_menu") end
    end,
}
