local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local realm = require("engine.realm/v1")
local controls = require("d2legacy.ui.controls")
local list = require("d2legacy.ui.list")
local text_field = require("d2legacy.ui.text_field")
local button = require("d2legacy.ui.button")
local common = require("d2legacy.realm.common")
local lobby_roster = require("d2legacy.realm.lobby_roster")
local lobby_ui = require("d2legacy.realm.lobby_ui")

local function lobby_button(name, x, y)
    return lobby_ui.button(name, x, y)
end

local function begin_logout(scene)
    if scene.leaving then return end
    scene.leaving = realm.logout()
    if scene.leaving then
        common.set_label(scene.status, "LOGGING OUT", 330, "network_status", "center")
    else
        common.set_label(scene.status, "REALM IS BUSY", 330, "network_error", "center")
    end
end

local function queue_game_join(scene)
    local reference = scene.game_name.value
    if reference == "" then reference = scene.selected_game end
    if reference then
        scene.pending_join = { reference=reference, password=scene.game_password.value }
    end
end

local function clear_game_selection(scene)
    scene.selected_game = nil
    scene.pending_game_detail = nil
    scene.game_name:set_value("")
    scene.game_password:set_value("")
    realm.select_game("")
end

return {
    create = function(self)
        self.root, self.backdrop, self.game_panel = lobby_ui.create("join_panel")
        self.controls = controls.new()
        self.chat = common.label(self.root, "", 195, 255, 340, "realm_lobby_text", "left")
        self.members = lobby_roster.create(self.root)
        self.status = common.label(self.root, "", 590, 560, 330, "realm_lobby_error", "center")
        self.game_name = text_field.create(self.root, self.controls, "game_name", {
            x=426, y=124, width=160, height=20, value="", max_length=32, background=false,
        }, "", { text_style="realm_lobby_text" })
        self.game_password = text_field.create(self.root, self.controls, "game_password", {
            x=600, y=124, width=160, height=20, value="", max_length=64, mask=true, background=false,
        }, "", { text_style="realm_lobby_text" })
        self.games = list.create(self.root, self.controls, "games", {
            x=426, y=194, width=160, row_height=22, page_size=8,
        }, {}, {
            item_label=function(item) return string.format("%s   %d/%d",item.name,item.players,item.maximum_players) end,
            on_select=function(item)
                self.selected_game = item.game_id
                self.game_name:set_value(item.name)
                self.pending_game_detail = item.game_id
            end,
            normal_style="realm_lobby_gold", hover_style="realm_lobby_text", selected_style="realm_lobby_text",
        })
        self.game_detail = common.label(self.root, "", 684, 215, 160, "realm_lobby_text", "center")
        self.game_players = common.label(self.root, "", 684, 255, 160, "realm_lobby_text", "left")
        self.message = text_field.create(self.root, self.controls, "message", {
            x=35, y=415, width=340, height=28, value="", max_length=255, palette="units",
        }, "", { text_style="realm_lobby_text" })
        button.create(self.root, self.controls, "send", lobby_button("chat_button", 35, 450), "SEND", {
            normal_style="realm_lobby_button", hover_style="realm_lobby_button",
            on_activate=function()
                if self.message.value ~= "" and realm.send_message(self.message.value) then self.message.value = "" end
            end,
        })
        lobby_ui.add_navigation(self.root, self.controls, {
            create=function() scenes.replace("realm_game_create") end,
            join=function() self.controls:set_focus("game_name") end,
            quit=function() begin_logout(self) end,
        })
        button.create(self.root, self.controls, "cancel_game", lobby_button("panel_cancel_button", 432, 399), "CANCEL", {
            normal_style="realm_lobby_button", hover_style="realm_lobby_button",
            on_activate=function() clear_game_selection(self) end,
        })
        button.create(self.root, self.controls, "join_game", lobby_button("panel_game_button", 593, 399), "JOIN GAME", {
            normal_style="realm_lobby_button", hover_style="realm_lobby_button",
            on_activate=function() queue_game_join(self) end,
        })
        self.last_refresh = 0
    end,
    update = function(self, elapsed)
        self.controls:update(elapsed)
        self.last_refresh = self.last_refresh + (elapsed or 0)
        local state = realm.status()
        self.games:set_items(state.games or {})
        local selected = (state.selected_game or {}).entry or {}
        local detail = ""
        if selected.game_id then
            detail = string.format("%s\n%s\n%d/%d PLAYERS", selected.name or "", string.upper(selected.difficulty or "normal"), selected.players or 0, selected.maximum_players or 8)
        end
        common.set_label(self.game_detail, detail, 160, "realm_lobby_text", "center")
        local player_lines = {}
        for _, player in ipairs((state.selected_game or {}).players or {}) do
            player_lines[#player_lines + 1] = string.format("%s  LEVEL %d %s", player.name or "", player.level or 1, string.upper(player.class or ""))
        end
        common.set_label(self.game_players, table.concat(player_lines, "\n"), 160, "realm_lobby_text", "left")
        common.set_label(self.chat, lobby_ui.chat_lines(state.events or {}), 340, "realm_lobby_text", "left")
        -- Channel membership is the Realm's live-presence contract. Do not
        -- project the account roster or selected character into this strip.
        lobby_roster.update(self.members, (state.channel or {}).members or {})
        if state.phase == "error" then
            common.set_label(self.status, common.error(state), 330, "realm_lobby_error", "center")
        elseif state.phase == "game_connecting" or state.phase == "game_connected" then
            scenes.replace("game_loading")
            return
        end
        -- Polling and foreground actions share one native request lane. Keep
        -- user intent until that lane is idle instead of losing a click that
        -- happened during the short periodic refresh.
        if state.phase == "lobby" and self.pending_game_detail then
            if realm.select_game(self.pending_game_detail) then self.pending_game_detail = nil end
        elseif state.phase == "lobby" and self.pending_join then
            local request = self.pending_join
            if realm.join_game(request.reference, request.password) then self.pending_join = nil end
        end
        -- Returning from a canceled/failed game pane intentionally lands in a
        -- neutral native phase. One refresh re-establishes live channel state;
        -- without this recovery the screen rendered but could not act again.
        if (state.phase == "ready" or state.phase == "error") and not self.leaving and not self.recovering then
            self.recovering = realm.refresh()
        elseif state.phase == "lobby" then
            self.recovering = nil
        end
        if self.last_refresh >= 1 and state.phase == "lobby" then self.last_refresh=0; realm.refresh() end
        if self.leaving and state.phase == "login" then
            scenes.replace("main_menu")
            return
        end
        if input.pressed("cancel") and not self.leaving then
            begin_logout(self)
        end
    end,
}
