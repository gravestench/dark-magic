-- Explicit Realm login. Endpoint trust is prepared by realm_connecting; this
-- screen always requires account credentials and never persists the password.

local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local realm = require("engine.realm/v1")
local controls = require("d2legacy.ui.controls")
local text_field = require("d2legacy.ui.text_field")
local common = require("d2legacy.realm.common")
local account_ui = require("d2legacy.realm.account_ui")

-- Recovered 800x600 frontend coordinates. The logo owns the upper third; the
-- account form begins below the expansion subtitle and ends above the footer.
local layout = {
    instruction_y = 240,
    name_y = 300,
    password_y = 365,
    status_y = 415,
    login_y = 440,
    create_y = 480,
    recover_y = 520,
}

return {
    create = function(self)
        self.root, self.logo = account_ui.create_root()
        self.controls = controls.new()
        common.label(self.root, "ENTER YOUR ACCOUNT NAME AND PASSWORD", 400, layout.instruction_y, 440, "formal_small", "center")
        self.status = common.label(self.root, "", 400, layout.status_y, 520, "network_error", "center")

        self.name = text_field.create(self.root, self.controls, "account_name", account_ui.text_field({
            x = 264,
            y = layout.name_y,
            value = "",
            max_length = 32,
        }), "ACCOUNT NAME", { label_style = "character_create_option", text_style = "text_field_value" })
        self.password = text_field.create(self.root, self.controls, "password", account_ui.text_field({
            x = 264,
            y = layout.password_y,
            value = "",
            max_length = 72,
            mask = true,
        }), "PASSWORD", { label_style = "character_create_option", text_style = "text_field_value" })

        account_ui.add_button(self.root, self.controls, "login", layout.login_y, "LOG IN", function()
            if self.name.value == "" or self.password.value == "" then
                common.set_label(self.status, "ACCOUNT NAME AND PASSWORD ARE REQUIRED", 520, "network_error", "center")
                return
            end
            if realm.login(self.name.value, self.password.value) then
                account_ui.clear_secret(self.password)
                common.set_label(self.status, "LOGGING IN", 520, "network_status", "center")
            else
                common.set_label(self.status, "REALM IS BUSY", 520, "network_error", "center")
            end
        end)
        account_ui.add_button(self.root, self.controls, "create", layout.create_y, "CREATE ACCOUNT", function()
            account_ui.clear_secret(self.password)
            scenes.replace("realm_signup")
        end)
        account_ui.add_button(self.root, self.controls, "recover", layout.recover_y, "FORGOT PASSWORD", function()
            account_ui.clear_secret(self.password)
            scenes.replace("realm_recovery")
        end)
    end,

    update = function(self, elapsed)
        account_ui.update_logo(self.logo, elapsed)
        self.controls:update(elapsed)
        local state = realm.status()
        if state.phase == "characters" then
            account_ui.clear_secret(self.password)
            scenes.replace(#(state.characters or {}) == 0 and "realm_create" or "realm_characters")
        elseif state.phase == "error" then
            account_ui.clear_secret(self.password)
            common.set_label(self.status, common.error(state), 520, "network_error", "center")
        elseif state.phase == "authenticating_account" or state.phase == "loading_characters" then
            common.set_label(self.status, "LOGGING IN", 520, "network_status", "center")
        end
        if input.pressed("cancel") then
            account_ui.clear_secret(self.password)
            realm.cancel()
            scenes.replace("main_menu")
        end
    end,
}
