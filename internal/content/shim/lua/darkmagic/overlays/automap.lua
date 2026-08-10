-- Automap overlay example.
--
-- Overlays are ordinary scenes pushed above a root screen. Setting
-- blocks_update_below to false lets the game simulation continue while this
-- transparent HUD layer is visible.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = assert(manifest.screens.automap)

return {
    blocks_update_below = false,

    -- Scene-owned render handles are released automatically when this overlay
    -- is popped, so mods do not need a matching destroy callback.
    create = function(self)
        self.root = render.create("hud")
        self.root:set_position(screen.x, screen.y)
        -- This placeholder demonstrates a translucent, screen-centered layer.
        -- A complete automap can replace it with record-driven map geometry.
        self.root:fill_rect(screen.width, screen.height,
            screen.fill.red, screen.fill.green, screen.fill.blue, screen.fill.alpha)
    end,

    update = function(self)
        -- The same action that opened an overlay should conventionally close it;
        -- cancel provides a consistent keyboard/controller escape path.
        if input.pressed("automap") or input.pressed("cancel") then
            scenes.toggle_overlay("automap", "full")
        end
    end,
}
