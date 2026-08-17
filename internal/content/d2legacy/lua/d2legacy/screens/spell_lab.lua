-- Spell Lab is the production game world with a small usage legend.
--
-- Learned skills are supplied by the development composition root, but every
-- assignment, cast, target, mana payment, missile, state, damage result, and
-- presentation entity runs through the ordinary authoritative gameplay path.

local game_world = require("d2legacy.screens.game_world")

local lab = {}

function lab:create()
    game_world.create(self)
    self.spell_diagnostics_module = require("d2legacy.dev.spell_lab.diagnostics")
    self.spell_diagnostics = self.spell_diagnostics_module.create(self)
end

function lab:update(elapsed, focused, input_allowed, world_view)
    game_world.update(self, elapsed, focused, input_allowed, world_view)
    self.spell_diagnostics_module.update(self.spell_diagnostics)
end

function lab:destroy()
    self.spell_diagnostics_module.destroy(self.spell_diagnostics)
    game_world.destroy(self)
end

return lab
