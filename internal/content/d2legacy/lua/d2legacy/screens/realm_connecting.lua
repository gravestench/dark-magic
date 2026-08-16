-- Realm connection progress is its own frontend state. Once transport and
-- compatibility checks pass, the client proceeds to explicit account login.

local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local realm = require("engine.realm/v1")
local controls = require("d2legacy.ui.controls")
local button = require("d2legacy.ui.button")
local common = require("d2legacy.realm.common")

local phase_labels = {
    checking_versions = "CHECKING VERSIONS",
}

return {
    create = function(self)
        self.root = common.frontend_root()
        common.popup(self.root, 230, 170)
        common.label(self.root, "CONNECTING TO REALM", 400, 200, 290, "realm_heading", "center")
        self.progress = common.label(self.root, "CHECKING VERSIONS", 400, 265, 280, "network_status", "center")
        self.dots = common.label(self.root, "", 400, 295, 120, "network_status", "center")
        self.error = common.label(self.root, "", 400, 305, 280, "network_error", "center")
        self.controls = controls.new()
        button.create(
            self.root,
            self.controls,
            "cancel",
            common.button_definition("cancel_button", 352, 345),
            "CANCEL",
            {
                normal_style = "dialog_button_normal",
                hover_style = "dialog_button_hover",
                on_activate = function()
                    realm.cancel()
                    scenes.replace("main_menu")
                end,
            }
        )
        self.elapsed = 0
        if not realm.connect("") then
            common.set_label(self.error, "INVALID GATEWAY", 250, "network_error", "center")
        end
    end,

    update = function(self, elapsed)
        self.controls:update(elapsed)
        self.elapsed = self.elapsed + (elapsed or 0)
        common.set_label(self.dots, string.rep(".", math.floor(self.elapsed * 2) % 4), 120, "network_status", "center")

        -- Cancellation wins over a transport completion observed in the same
        -- frame; only one scene transition may be emitted from this update.
        if input.pressed("cancel") then
            realm.cancel()
            scenes.replace("main_menu")
            return
        end

        local status = realm.status()
        if status.phase == "login" then
            scenes.replace("realm_login")
        elseif status.phase == "error" then
            if self.progress then
                self.progress:set_visible(false)
            end
            if self.dots then
                self.dots:set_visible(false)
            end
            common.set_label(self.error, common.error(status), 270, "network_error", "center")
        else
            common.set_label(self.progress, phase_labels[status.phase] or "CONNECTING", 240, "network_status", "center")
        end

    end,
}
