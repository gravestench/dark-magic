-- Automap overlay example.
--
-- This is a small, friendly place to learn what an OVERLAY really is: an
-- overlay is just another scene sitting above the root game-world scene.
--
-- `blocks_update_below = false` is the important line. It says the world may
-- keep updating while this transparent HUD presentation exists. Input routing
-- is handled separately by the scene/overlay system, so "keeps updating" does
-- not automatically mean "owns the same input."

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local data = require("engine.data/v1")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2legacy.presentation/v1"))
local screen = assert(manifest.screens.automap)

return {
    -- Lower scenes continue their update lifecycle under this overlay.
    blocks_update_below = false,

    create = function(self)
        -- The root handle is owned by this overlay's scene scope. When the scene
        -- is popped, native render resources are reclaimed automatically.
        self.root = render.create("hud")

        -- Manifest x/y describe where this overlay's rectangle belongs.
        self.root:set_position(screen.x, screen.y)

        -- This is currently a simple translucent placeholder. A real automap can
        -- replace the fill with record-driven line/icon nodes while keeping the
        -- same scene lifecycle and routing behavior.
        self.root:fill_rect(screen.width, screen.height,
            screen.fill.red, screen.fill.green, screen.fill.blue, screen.fill.alpha)
    end,

    update = function(self)
        -- A useful overlay convention: the same logical action that opened the
        -- panel closes it again. Cancel is the universal escape path.
        if input.pressed("automap") or input.pressed("cancel") then
            -- `full` is the slot this overlay occupies; toggling that same ID/slot
            -- asks the scene manager to close it.
            scenes.toggle_overlay("automap", "full")
        end
    end,
}
