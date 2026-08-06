-- TCP/IP multiplayer intent screen.
--
-- Network hosting and joining are not implemented here. This example shows how
-- a screen collects validated user intent without exposing networking internals
-- to Lua or storing native objects in the scene table.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local locale = require("dm.locale/v1")
local controls = require("darkmagic.ui.controls")
local button = require("darkmagic.ui.button")
local cursor = require("darkmagic.ui.cursor")
local dialog = require("darkmagic.ui.dialog")
local dc6 = require("darkmagic.ui.dc6")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen, font = manifest.screens.tcpip, manifest.fonts.exocet10

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

        local function add(id, activate)
            local definition = screen.controls[id]
            button.create(self.root, self.controls, id, definition, assert(locale.text(definition.label)), {
                layer = "hud",
                on_activate = activate,
            })
        end

        add("host", function()
            self.intent = { mode = "host" }
        end)
        add("join", function()
            self.dialog = dialog.text_entry(
                self.root,
                screen.dialog,
                font,
                manifest.palettes[screen.dialog.palette],
                manifest.palettes[font.palette],
                assert(locale.text("darkmagic.tcpip.address")),
                "",
                function(address)
                    if address ~= "" then
                        self.intent = { mode = "join", address = address }
                    end
                end
            )
        end)
        self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)
    end,
    update = function(self)
        self.cursor:update()
        if self.dialog and self.dialog.open then
            self.dialog:update()
            if input.pressed("cancel") then
                self.dialog:close()
            end
            return
        end
        self.controls:update()
        if input.pressed("cancel") then
            scenes.replace("main_menu")
        end
    end,
}
