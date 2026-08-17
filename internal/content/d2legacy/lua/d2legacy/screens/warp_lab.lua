-- Warp Lab is the production game world with transition diagnostics.
--
-- The direct-start fixture enters just inside the authored Rogue Encampment /
-- Blood Moor seam. Everything that can change gameplay state is therefore the
-- ordinary local Session: authenticated player.move commands, collision-aware
-- pathfinding, velocity integration, facing and animation systems, level
-- transition authority, camera following, and world residency. This wrapper
-- owns only read-only labels and a short presentation mask after a crossing.

local game_world = require("d2legacy.screens.game_world")
local diagnostics = require("d2legacy.dev.warp_lab.diagnostics")

local lab = {}

function lab:create()
    game_world.create(self)
    self.warp_diagnostics = diagnostics.create(self)
end

function lab:update(elapsed, focused, input_allowed, world_view)
    game_world.update(self, elapsed, focused, input_allowed, world_view)
    diagnostics.update(self.warp_diagnostics, self, elapsed)
end

function lab:destroy()
    diagnostics.destroy(self.warp_diagnostics)
    game_world.destroy(self)
end

return lab
