-- Authored Create Game tab inside the Realm waiting room.

local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local realm = require("engine.realm/v1")
local controls = require("d2legacy.ui.controls")
local text_field = require("d2legacy.ui.text_field")
local button = require("d2legacy.ui.button")
local checkbox = require("d2legacy.ui.checkbox")
local text = require("d2legacy.ui.text")
local common = require("d2legacy.realm.common")
local lobby_ui = require("d2legacy.realm.lobby_ui")
local lobby_roster = require("d2legacy.realm.lobby_roster")
local number_selector = require("d2legacy.realm.number_selector")

local function selected_rules()
    local selected = ((realm.status() or {}).selected or {}).character or {}
    return selected.expansion ~= false, selected.hardcore == true
end

local function return_to_lobby()
    -- Cancel owns any in-flight allocation request. realm_lobby recovers the
    -- controller's resulting ready/error state with a fresh directory request.
    realm.cancel()
    scenes.replace("realm_lobby")
end

return {
    create = function(self)
        self.root, self.backdrop, self.game_panel = lobby_ui.create("create_panel")
        self.controls = controls.new()
        self.chat = common.label(self.root, "", 195, 255, 340, "realm_lobby_text", "left")
        self.members = lobby_roster.create(self.root)

        text.create(self.root, "realm_lobby_heading", "CREATE GAME", 600, 76, 340, "center", "hud")
        text.create(self.root, "realm_lobby_label", "GAME NAME", 518, 113, 180, "left", "hud")
        text.create(self.root, "realm_lobby_label", "PASSWORD (OPTIONAL)", 518, 168, 180, "left", "hud")
        text.create(self.root, "realm_lobby_label", "GAME DESCRIPTION (OPTIONAL)", 596, 223, 336, "left", "hud")

        self.name = text_field.create(self.root, self.controls, "name", {
            x = 428,
            y = 137,
            width = 180,
            height = 20,
            value = "",
            max_length = 32,
            background = false,
        }, "", { text_style = "realm_lobby_text" })
        self.password = text_field.create(self.root, self.controls, "password", {
            x = 428,
            y = 191,
            width = 180,
            height = 20,
            value = "",
            max_length = 64,
            mask = true,
            background = false,
        }, "", { text_style = "realm_lobby_text" })
        self.description = text_field.create(self.root, self.controls, "description", {
            x = 428,
            y = 245,
            width = 336,
            height = 20,
            value = "",
            max_length = 96,
            background = false,
        }, "", { text_style = "realm_lobby_text" })

        text.create(self.root, "realm_lobby_label", "MAX. NUMBER OF PLAYERS", 535, 288, 210, "left", "hud")
        text.create(self.root, "realm_lobby_label", "CHARACTER DIFFERENCE", 535, 323, 210, "left", "hud")
        self.maximum_players = number_selector.create(self.root, self.controls, "maximum_players", {
            x = 648,
            y = 280,
            width = 30,
            height = 32,
            arrow_x = 684,
            value = 4,
            minimum = 1,
            maximum = 8,
        })
        self.character_difference = number_selector.create(self.root, self.controls, "character_difference", {
            x = 648,
            y = 315,
            width = 30,
            height = 32,
            arrow_x = 684,
            value = 4,
            minimum = 1,
            maximum = 99,
        })
        self.character_difference_enabled = true
        self.character_difference_toggle = checkbox.create(
            self.root,
            self.controls,
            "character_difference_enabled",
            {
                x = 714,
                y = 323,
                width = 15,
                height = 16,
                checked = true,
                sheet = "data/global/ui/FrontEnd/clickbox.dc6",
                palette = "fechar",
            },
            "",
            {
                layer = "hud",
                on_change = function(_, checked)
                    self.character_difference_enabled = checked
                end,
            }
        )
        self.character_difference_toggle.focusable = false
        self.status = common.label(self.root, "", 600, 360, 330, "realm_lobby_error", "center")

        self.message = text_field.create(self.root, self.controls, "message", {
            x = 35,
            y = 415,
            width = 340,
            height = 28,
            value = "",
            max_length = 255,
            palette = "units",
        }, "", { text_style = "realm_lobby_text" })
        button.create(self.root, self.controls, "send", lobby_ui.button("chat_button", 35, 450), "SEND", {
            normal_style = "realm_lobby_button",
            hover_style = "realm_lobby_button",
            on_activate = function()
                if self.message.value ~= "" and realm.send_message(self.message.value) then
                    self.message:set_value("")
                end
            end,
        })
        lobby_ui.add_navigation(self.root, self.controls, {
            create = false,
            join = function()
                scenes.replace("realm_lobby")
            end,
            quit = function()
                if realm.logout() then
                    self.leaving = true
                end
            end,
        })

        button.create(self.root, self.controls, "cancel", lobby_ui.button("panel_cancel_button", 432, 399), "CANCEL", {
            normal_style = "realm_lobby_button",
            hover_style = "realm_lobby_button",
            on_activate = return_to_lobby,
        })
        button.create(
            self.root,
            self.controls,
            "create",
            lobby_ui.button("panel_game_button", 593, 399),
            "CREATE GAME",
            {
                normal_style = "realm_lobby_button",
                hover_style = "realm_lobby_button",
                on_activate = function()
                    if self.name.value == "" then
                        common.set_label(self.status, "ENTER A GAME NAME", 330, "realm_lobby_error", "center")
                        return
                    end
                    local use_expansion, use_hardcore = selected_rules()
                    self.pending_create = {
                        name = self.name.value,
                        password = self.password.value,
                        description = self.description.value,
                        difficulty = "normal",
                        maximum_players = self.maximum_players.value,
                        character_difference = self.character_difference_enabled and self.character_difference.value
                            or 0,
                        expansion = use_expansion,
                        hardcore = use_hardcore,
                    }
                end,
            }
        )
        self.last_refresh = 0
        self.controls:set_focus("name")
    end,

    update = function(self, elapsed)
        self.controls:update(elapsed)
        self.last_refresh = self.last_refresh + (elapsed or 0)
        local state = realm.status()
        common.set_label(self.chat, lobby_ui.chat_lines(state.events or {}), 340, "realm_lobby_text", "left")
        lobby_roster.update(self.members, (state.channel or {}).members or {})

        if state.phase == "game_connecting" or state.phase == "game_connected" then
            scenes.replace("game_loading")
            return
        elseif state.phase == "error" then
            common.set_label(self.status, common.error(state), 330, "realm_lobby_error", "center")
        elseif state.phase == "creating_game" then
            common.set_label(self.status, "CREATING GAME", 330, "network_status", "center")
        end
        if self.leaving and state.phase == "login" then
            scenes.replace("main_menu")
            return
        end

        -- Periodic presence refresh and the create click share one native lane.
        -- Preserve the click until that lane becomes available.
        if self.pending_create and (state.phase == "lobby" or state.phase == "error") then
            if realm.create_game(self.pending_create) then
                self.pending_create = nil
            end
        elseif self.last_refresh >= 1 and state.phase == "lobby" then
            self.last_refresh = 0
            realm.refresh()
        end

        if input.pressed("cancel") then
            return_to_lobby()
        end
    end,
}
