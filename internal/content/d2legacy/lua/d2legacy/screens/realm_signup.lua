-- Realm account creation remains separate from login. Even development
-- auto-verification only verifies email; the player returns and logs in.

local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local realm = require("engine.realm/v1")
local controls = require("d2legacy.ui.controls")
local text_field = require("d2legacy.ui.text_field")
local common = require("d2legacy.realm.common")
local account_ui = require("d2legacy.realm.account_ui")

-- Four account fields must fit between the frontend logo and footer actions.
-- These 800x600 coordinates retain a consistent label/field rhythm without
-- intruding into either region.
local layout = {
    heading_y = 240,
    name_y = 275,
    email_y = 330,
    password_y = 385,
    confirm_y = 440,
    status_y = 470,
    create_y = 490,
    back_y = 530,
}

return {
    create = function(self)
        self.root, self.logo = account_ui.create_root()
        self.controls = controls.new()
        common.label(self.root, "CREATE REALM ACCOUNT", 400, layout.heading_y, 420, "realm_heading", "center")
        self.status = common.label(self.root, "", 400, layout.status_y, 540, "network_error", "center")
        self.name = text_field.create(self.root, self.controls, "account_name", account_ui.text_field({
            x = 264,
            y = layout.name_y,
            max_length = 32,
        }), "ACCOUNT NAME", { text_style = "text_field_value" })
        self.email = text_field.create(self.root, self.controls, "email", account_ui.text_field({
            x = 264,
            y = layout.email_y,
            max_length = 320,
        }), "EMAIL", { text_style = "text_field_value" })
        self.password = text_field.create(self.root, self.controls, "password", account_ui.text_field({
            x = 264,
            y = layout.password_y,
            max_length = 72,
            mask = true,
        }), "PASSWORD", { text_style = "text_field_value" })
        self.confirm = text_field.create(self.root, self.controls, "confirm_password", account_ui.text_field({
            x = 264,
            y = layout.confirm_y,
            max_length = 72,
            mask = true,
        }), "CONFIRM PASSWORD", { text_style = "text_field_value" })
        account_ui.add_button(self.root, self.controls, "create", layout.create_y, "CREATE ACCOUNT", function()
            if #self.password.value < 8 then
                common.set_label(self.status, "PASSWORD MUST BE AT LEAST 8 CHARACTERS", 540, "network_error", "center")
            elseif self.password.value ~= self.confirm.value then
                common.set_label(self.status, "PASSWORDS DO NOT MATCH", 540, "network_error", "center")
            elseif realm.signup(self.name.value, self.email.value, self.password.value) then
                account_ui.clear_secret(self.password)
                account_ui.clear_secret(self.confirm)
                common.set_label(self.status, "CREATING ACCOUNT", 540, "network_status", "center")
            end
        end)
        account_ui.add_button(self.root, self.controls, "back", layout.back_y, "BACK TO LOG IN", function()
            account_ui.clear_secret(self.password)
            account_ui.clear_secret(self.confirm)
            scenes.replace("realm_login")
        end)
    end,
    update = function(self, elapsed)
        account_ui.update_logo(self.logo, elapsed)
        self.controls:update(elapsed)
        local state = realm.status()
        if state.phase == "verification_required" then
            common.set_label(self.status, "CHECK YOUR EMAIL, THEN LOG IN", 540, "network_status", "center")
        elseif state.phase == "error" then
            common.set_label(self.status, common.error(state), 540, "network_error", "center")
        end
        if input.pressed("cancel") then
            account_ui.clear_secret(self.password)
            account_ui.clear_secret(self.confirm)
            scenes.replace("realm_login")
        end
    end,
}
