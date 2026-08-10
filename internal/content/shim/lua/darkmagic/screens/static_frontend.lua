-- Factory for simple manifest-backed frontend screens.
--
-- A FUNCTION can be the value returned by a Lua module. That is what happens at
-- the bottom of this file: `require("darkmagic.screens.static_frontend")` gives
-- the caller a factory function, and calling that function with a screen name
-- produces a complete scene-definition table.
--
-- This avoids copy/pasting the same background/cursor/cancel lifecycle into
-- every simple frontend screen. Screens with unique behavior (credits, movies,
-- TCP/IP) still use their own modules.

local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local dc6 = require("darkmagic.ui.dc6")
local cursor = require("darkmagic.ui.cursor")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))

-- Return the factory itself.
return function(name)
    -- Validate that the requested manifest screen exists when the factory is called.
    local screen = assert(manifest.screens[name])

    -- And return an ordinary scene definition for the scene registry to register.
    return {
        create = function(self)
            self.root = render.create("hud")

            -- Reuse the DC6 helper that assembles the 800x600 tiled frontend background.
            self.background = dc6.frontend_background(
                self.root,
                "hud",
                screen.background,
                manifest.palettes[screen.palette],
                manifest.layouts.frontend_tiles
            )

            self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)
        end,

        update = function(self)
            self.cursor:update()

            -- Simple frontend surfaces all share the same Escape/Cancel destination.
            if input.pressed("cancel") then
                scenes.replace("main_menu")
            end
        end,
    }
end
