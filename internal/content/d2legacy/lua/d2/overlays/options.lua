-- Diablo II in-game options overlay.
--
-- This file demonstrates REUSING a complicated widget in a different scene.
-- `d2.ui.escape_menu` can represent the whole Escape hierarchy. pause.lua
-- starts that widget at `main`; this convenience overlay starts the SAME widget
-- directly at `options`.
--
-- Because there is no pause scene underneath this standalone overlay, choosing
-- PREVIOUS MENU at the options root calls `on_close` and pops this overlay.

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local settings = require("engine.settings/v1")
local data = require("engine.data/v1")
local compat = require("d2.ui.compat")
local escape_menu = require("d2.ui.escape_menu")

-- Compatibility catalog contains recovered original menu facts. The manifest
-- contains the selected Dark Magic presentation profile/resolution.
local recovered = compat.ingame.escape_menu
local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2.presentation/v1"))
local viewport = manifest.resolution

-- Retained nodes use center positions, so calculate the logical viewport center once.
local center = { x = viewport.width / 2, y = viewport.height / 2 }

return {
    -- This policy comes from recovered simulation behavior rather than being
    -- invented by the visual widget. Multiplayer will eventually need session-aware policy.
    blocks_update_below = recovered.simulation.pauses_single_player,

    create = function(self)
        self.root = render.create("modal")

        -- Keep root at the screen origin. Moving the whole root would shift
        -- visuals while controls still use screen-relative hit rectangles.
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

        -- Construct the reusable Escape-menu object, but begin at OPTIONS.
        self.menu = escape_menu.new(self.root, {
            center = center,
            start_layout = "options",
            -- The reusable menu does not own scene navigation. The scene gives it
            -- this callback describing what "close" means in this context.
            on_close = function()
                scenes.pop()
            end,
        })
    end,

    update = function(self, elapsed)
        -- Pentagram animation and control input need elapsed time/update each frame.
        self.menu:update(elapsed)

        -- The dedicated Options action toggles the convenience overlay too.
        if input.pressed("options") or input.pressed("cancel") then
            scenes.pop()
        end
    end,

    destroy = function()
        -- Settings capability owns persistence. Lua asks for status and requests
        -- a save only when there are dirty settings AND a configured save path.
        local status = settings.status()
        if status.dirty and status.path ~= "" then settings.save() end
    end,
}
