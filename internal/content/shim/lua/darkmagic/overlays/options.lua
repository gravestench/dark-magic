-- Diablo II in-game options overlay.
--
-- This is the same recovered Escape-menu hierarchy used by pause.lua, entered
-- directly at OPTIONS for Dark Magic's convenience hotkey. PREVIOUS MENU from
-- this root closes the standalone overlay instead of manufacturing a pause
-- scene underneath it.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local settings = require("dm.settings/v1")
local compat = require("darkmagic.ui.compat")
local escape_menu = require("darkmagic.ui.escape_menu")

local recovered = compat.ingame.escape_menu

return {
    blocks_update_below = recovered.simulation.pauses_single_player,

    create = function(self)
        self.root = render.create("modal")
        self.root:set_position(recovered.center.x, recovered.center.y)
        self.root:fill_rect(
            800,
            600,
            recovered.dim.red,
            recovered.dim.green,
            recovered.dim.blue,
            recovered.dim.alpha
        )
        self.menu = escape_menu.new(self.root, {
            start_layout = "options",
            on_close = function()
                scenes.pop()
            end,
        })
    end,

    update = function(self, elapsed)
        self.menu:update(elapsed)
        if input.pressed("options") or input.pressed("cancel") then
            scenes.pop()
        end
    end,

    destroy = function()
        local status = settings.status()
        if status.dirty and status.path ~= "" then settings.save() end
    end,
}
