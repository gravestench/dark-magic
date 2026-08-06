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

        -- Both buttons use the same interaction contract but retain independent
        -- manifest geometry, localized labels, and DC6 frame definitions.
        local function add(id, activate)
            local definition = screen.controls[id]
            local label_text = assert(locale.text(definition.label))
            local control = {
                id = id,
                label = label_text,
                x = definition.x,
                y = definition.y,
                width = definition.width,
                height = definition.height,
                on_activate = activate,
            }
            if render.assets_available() then
                local left = render.create("hud", self.root)
                local right = render.create("hud", self.root)
                local label = render.create("hud", self.root)
                local palette = manifest.palettes[definition.palette]

                local function draw_frames(frames)
                    left:set_dc6(definition.sheet, palette, 0, frames[1])
                    right:set_dc6(definition.sheet, palette, 0, frames[2])
                end

                draw_frames(definition.up_frames)
                left:set_position(definition.x + 128, definition.y + definition.height / 2)
                right:set_position(definition.x + 264, definition.y + definition.height / 2)
                label:set_text(
                    font.table,
                    font.sheet,
                    manifest.palettes[font.palette],
                    label_text,
                    {
                        red = 210,
                        green = 180,
                        blue = 110,
                        max_width = definition.width,
                        align = "center",
                    }
                )
                label:set_position(definition.x + definition.width / 2, definition.y + definition.height / 2)
                control.on_state = function(_, state)
                    if state == "focused" or state == "hover" then
                        draw_frames(definition.down_frames)
                    else
                        draw_frames(definition.up_frames)
                    end
                end
            end
            self.controls:add(control)
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
