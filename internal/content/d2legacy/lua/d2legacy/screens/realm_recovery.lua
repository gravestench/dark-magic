-- Password recovery starts in the client but completes in the browser from the
-- emailed (or development-log) bearer link.

local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local realm = require("engine.realm/v1")
local controls = require("d2legacy.ui.controls")
local text_field = require("d2legacy.ui.text_field")
local common = require("d2legacy.realm.common")
local account_ui = require("d2legacy.realm.account_ui")

local layout = {
    heading_y = 240,
    instruction_y = 265,
    email_y = 310,
    status_y = 360,
    send_y = 400,
    back_y = 440,
}

return {
    create = function(self)
        self.root, self.logo = account_ui.create_root()
        self.controls = controls.new()
        common.label(self.root, "RESET REALM PASSWORD", 400, layout.heading_y, 440, "realm_heading", "center")
        common.label(self.root, "WE WILL SEND A BROWSER PASSWORD-RESET LINK", 400, layout.instruction_y, 520, "formal_small", "center")
        self.status = common.label(self.root, "", 400, layout.status_y, 540, "network_status", "center")
        self.email = text_field.create(self.root, self.controls, "email", account_ui.text_field({
            x = 264,
            y = layout.email_y,
            max_length = 320,
        }), "EMAIL", { text_style = "text_field_value" })
        account_ui.add_button(self.root, self.controls, "send", layout.send_y, "SEND RESET LINK", function()
            if self.email.value == "" then
                common.set_label(self.status, "EMAIL IS REQUIRED", 540, "network_error", "center")
            elseif realm.recover_password(self.email.value) then
                common.set_label(self.status, "SENDING RESET LINK", 540, "network_status", "center")
            end
        end)
        account_ui.add_button(self.root, self.controls, "back", layout.back_y, "BACK TO LOG IN", function()
            scenes.replace("realm_login")
        end)
    end,
    update = function(self, elapsed)
        account_ui.update_logo(self.logo, elapsed)
        self.controls:update(elapsed)
        local state = realm.status()
        if state.phase == "recovery_sent" then
            common.set_label(self.status, "CHECK YOUR EMAIL OR REALM DEVELOPMENT LOG", 540, "network_status", "center")
        elseif state.phase == "error" then
            common.set_label(self.status, common.error(state), 540, "network_error", "center")
        end
        if input.pressed("cancel") then
            scenes.replace("realm_login")
        end
    end,
}
