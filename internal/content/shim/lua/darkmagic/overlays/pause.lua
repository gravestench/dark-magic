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
local data = require("dm.data/v1")
local compat = require("darkmagic.ui.compat")
local escape_menu = require("darkmagic.ui.escape_menu")

local recovered = compat.ingame.escape_menu
local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local viewport = manifest.resolution
local center = { x = viewport.width / 2, y = viewport.height / 2 }

return {
    blocks_update_below = recovered.simulation.pauses_single_player,

    create = function(self)
        self.root = render.create("modal")
        -- Keep the owning root at the origin so retained rendering and input
        -- hit regions share one coordinate space; center the authored menu in
        -- the selected presentation profile's logical viewport.
        self.backdrop = render.create("modal", self.root)
        self.backdrop:set_position(center.x, center.y)
        self.backdrop:fill_rect(
            viewport.width,
            viewport.height,
            recovered.dim.red,
            recovered.dim.green,
            recovered.dim.blue,
            recovered.dim.alpha
        )
        self.menu = escape_menu.new(self.root, {
            center = center,
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
