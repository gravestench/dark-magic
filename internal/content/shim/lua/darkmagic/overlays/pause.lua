-- Diablo II in-game Escape menu.
--
-- blocks_update_below models the single-player behavior recovered from
-- OpenDiablo2: opening Escape pauses world simulation. Multiplayer eventually
-- needs session-aware blocking because the original keeps the networked world
-- advancing while this UI has focus.
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
        -- Menu definitions use absolute 800x600 viewport coordinates. Keep the
        -- owning root at the origin so retained rendering and absolute input
        -- hit regions share one coordinate space; only center the backdrop.
        self.backdrop = render.create("modal", self.root)
        self.backdrop:set_position(recovered.center.x, recovered.center.y)
        self.backdrop:fill_rect(
            800,
            600,
            recovered.dim.red,
            recovered.dim.green,
            recovered.dim.blue,
            recovered.dim.alpha
        )
        self.menu = escape_menu.new(self.root, {
            start_layout = "main",
            on_close = function()
                scenes.pop()
            end,
            on_save_exit = function()
                -- Navigation requests are deferred and applied in order. Pop the
                -- overlay first, then replace the exposed gameplay scene so the
                -- result is a clean frontend stack rather than world+menu.
                scenes.pop()
                scenes.replace("main_menu")
            end,
        })
    end,

    update = function(self, elapsed)
        self.menu:update(elapsed)
        if input.pressed("pause") or input.pressed("cancel") then
            scenes.pop()
        end
    end,

    destroy = function()
        local status = settings.status()
        if status.dirty and status.path ~= "" then settings.save() end
    end,
}
