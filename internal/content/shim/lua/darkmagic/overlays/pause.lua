-- Diablo II in-game Escape/pause menu scene.
--
-- Notice the division of labor:
--   pause.lua                = scene lifecycle + what close/save-exit mean
--   darkmagic.ui.escape_menu = reusable menu pages/controls/pentagram visuals
--   dm.settings/v1           = actual settings values/persistence
--   dm.scene/v1              = actual navigation stack
--
-- This lets the exact same menu widget be reused by options.lua with different
-- scene-level behavior.

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
    -- Original single-player Escape behavior pauses the world below. Multiplayer
    -- eventually needs session-aware policy because network simulation continues.
    blocks_update_below = recovered.simulation.pauses_single_player,

    create = function(self)
        self.root = render.create("modal")

        -- Keep the root at origin so visual coordinates and control hitboxes use
        -- the same screen coordinate system. Center the BACKDROP itself instead.
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

            -- The reusable widget reports intent through callbacks. This scene
            -- decides that "close" means pop the overlay stack entry.
            on_close = function()
                scenes.pop()
            end,

            on_save_exit = function()
                -- Scene navigation is queued/applied in request order. Pop the
                -- menu first, then replace the exposed gameplay root. Reversing
                -- those requests could leave a strange world/menu stack shape.
                scenes.pop()
                scenes.replace("main_menu")
            end,
        })
    end,

    update = function(self, elapsed)
        self.menu:update(elapsed)

        -- Both Pause and universal Cancel close the menu when pressed here.
        if input.pressed("pause") or input.pressed("cancel") then
            scenes.pop()
        end
    end,

    destroy = function()
        -- Settings capability owns the file and serialization. This scene merely
        -- requests persistence if menu interaction made the settings dirty.
        local status = settings.status()
        if status.dirty and status.path ~= "" then settings.save() end
    end,
}
