-- Combat Lab is the production game world with extra observation.
--
-- It deliberately delegates the complete scene lifecycle instead of copying
-- map or actor code. Therefore collision, A* pathing, sparse tile residency,
-- culling, depth ordering, camera rules, pointer targeting, player/monster
-- composites, missiles, death, loot cues, and HUD input are exactly the same
-- code paths exercised by ordinary gameplay.

local game_world = require("d2legacy.screens.game_world")
local diagnostics = require("d2legacy.dev.combat_lab.diagnostics")

local lab = {}

function lab:create()
    game_world.create(self)
    self.combat_diagnostics = diagnostics.create(self)
end

function lab:update(elapsed, focused, input_allowed, world_view)
    game_world.update(self, elapsed, focused, input_allowed, world_view)
    diagnostics.update(self.combat_diagnostics, self, elapsed)
end

function lab:destroy()
    diagnostics.destroy(self.combat_diagnostics)
    game_world.destroy(self)
end

return lab
