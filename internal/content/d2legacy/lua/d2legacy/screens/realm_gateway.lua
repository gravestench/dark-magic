-- Gateway selection is deliberately separate from identity association. The
-- chosen address is durable client configuration and the connection starts
-- only when the player subsequently presses Realm.

local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local realm = require("engine.realm/v1")
local controls = require("d2legacy.ui.controls")
local text_field = require("d2legacy.ui.text_field")
local button = require("d2legacy.ui.button")
local common = require("d2legacy.realm.common")

return {
    create = function(self)
        self.root = common.frontend_root()
        common.popup(self.root, 230, 170)
        common.label(self.root, "SELECT GATEWAY", 400, 205, 300, "realm_heading", "center")
        common.label(self.root, "Enter the realm host name or IP address.", 400, 240, 300, "formal_small", "center")
        self.status = common.label(self.root, "", 400, 300, 250, "network_error", "center")
        self.controls = controls.new()
        local status = realm.status()
        self.gateway = text_field.create(self.root, self.controls, "gateway", {
            x=291, y=267, width=218, height=26, kind="ip",
            value=status.gateway or "127.0.0.1", max_length=255, palette="units",
        }, "GATEWAY", { label_style="panel_label" })
        button.create(self.root, self.controls, "ok", common.button_definition("ok_button", 288, 345), "OK", {
            normal_style="dialog_button_normal",
            hover_style="dialog_button_hover",
            on_activate=function()
                if realm.set_gateway(self.gateway.value) then
                    scenes.replace("main_menu")
                else
                    common.set_label(self.status, "INVALID GATEWAY", 320, "network_error", "center")
                end
            end,
        })
        button.create(self.root, self.controls, "cancel", common.button_definition("cancel_button", 416, 345), "CANCEL", {
            normal_style="dialog_button_normal",
            hover_style="dialog_button_hover",
            on_activate=function() scenes.replace("main_menu") end,
        })
    end,

    update = function(self, elapsed)
        self.controls:update(elapsed)
        if input.pressed("cancel") then scenes.replace("main_menu") end
    end,
}
