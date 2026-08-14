-- TCP/IP multiplayer INTENT screen.
--
-- This scene deliberately does NOT own sockets, listeners, peer connections, or
-- networking threads. It only collects what the player wants:
--
--   { mode = "host" }
--   { mode = "join", address = "..." }
--
-- A future networking/session capability can consume that intent. This is a
-- useful general modding pattern: presentation gathers choices; engine authority
-- owns the dangerous/native operation that eventually acts on those choices.

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local data = require("engine.data/v1")
local locale = require("engine.locale/v1")
local controls = require("d2legacy.ui.controls")
local button = require("d2legacy.ui.button")
local cursor = require("d2legacy.ui.cursor")
local dialog = require("d2legacy.ui.dialog")
local dc6 = require("d2legacy.ui.dc6")
local text = require("d2legacy.ui.text")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2legacy.presentation/v1"))

-- Lua can assign several locals from one comma-separated expression.
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
        self.status = render.create("hud", self.root)
        self.status_value = ""
        self.status_is_error = false

        local function show_status(value, is_error)
            is_error = is_error == true
            if value == self.status_value and is_error == self.status_is_error then
                return
            end
            self.status_value = value
            self.status_is_error = is_error
            if render.assets_available() then
                text.set(
                    self.status,
                    is_error and screen.status.error_style or screen.status.style,
                    value,
                    screen.status.width,
                    "center"
                )
            end
            -- Status belongs below the popup bounds so validation and
            -- asynchronous connection failures remain visible while it is open.
            self.status:set_position(screen.status.x, screen.status.y)
        end

        self.show_status = show_status

        -- Local helper removes repeated button boilerplate while allowing each
        -- call to provide a completely different semantic activation function.
        local function add(id, activate)
            local definition = screen.controls[id]
            button.create(self.root, self.controls, id, definition, assert(locale.text(definition.label)), {
                layer = "hud",
                on_activate = activate,
            })
        end

        add("host", function()
            show_status("", false)
            local network = require("engine.network/v1")
            if network.host() then
                scenes.replace("character_select")
            else
                local status = network.status()
                show_status(status.error or "UNABLE TO HOST GAME", true)
            end
        end)

        add("join", function()
            show_status("", false)
            -- Joining needs one extra piece of data, so open the reusable modal
            -- text-entry dialog instead of putting networking logic in the button.
            self.dialog = dialog.text_entry(
                self.root,
                screen.dialog,
                font,
                manifest.palettes[screen.dialog.palette],
                manifest.palettes[font.palette],
                assert(locale.text("d2legacy.tcpip.address")),
                "",
                function(address)
                    -- Returning nil/non-false allows dialog.text_entry to close.
                    -- Empty input simply leaves intent unchanged.
                    if address ~= "" then
                        local network = require("engine.network/v1")
                        if not network.join(address) then
                            local status = network.status()
                            show_status(status.error or "UNABLE TO JOIN GAME", true)
                            return false
                        end
                        scenes.replace("character_select")
                    end
                end
            )
        end)

        self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)
    end,

    update = function(self)
        self.cursor:update()

        local ok, network = pcall(require, "engine.network/v1")
        if ok then
            local status = network.status()
            if status.phase == "starting" then
                self.show_status("STARTING LISTEN SERVER...", false)
            elseif status.phase == "connected" then
                self.show_status("LISTENING ON " .. tostring(status.address), false)
            elseif status.phase == "failed" then
                self.show_status(status.error or "NETWORK OPERATION FAILED", true)
            end
        end

        if self.dialog and self.dialog.open then
            -- Modal owns input while open. Do NOT update underlying screen controls.
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
