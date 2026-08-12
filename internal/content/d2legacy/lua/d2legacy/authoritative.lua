-- First-party Diablo II gameplay composition root.
--
-- This file deliberately contains no formulas or mutations. Its only job is to
-- show which small modules make up the mod and register them in readable order.

local components = require("d2legacy.components.fire_bolt")
local shared_components = require("d2legacy.components.shared")
local fire_bolt_data = require("d2legacy.data.fire_bolt")
local cast_command = require("d2legacy.commands.cast")
local cast_system = require("d2legacy.systems.cast")
local fire_bolt_system = require("d2legacy.systems.fire_bolt")
local projectile_system = require("d2legacy.systems.projectile")
local melee_components = require("d2legacy.components.melee")
local melee_system = require("d2legacy.systems.melee")
local monster_ai = require("d2legacy.systems.monster_ai")
local player_melee = require("d2legacy.systems.player_melee")
local spawn_monster = require("d2legacy.commands.spawn_monster")
local monster_death = require("d2legacy.systems.monster_death")
local enter_player = require("d2legacy.commands.enter_player")
local transition = require("d2legacy.commands.transition")
local timed_state = require("d2legacy.systems.timed_state")

local M = {
    id = "d2legacy.authoritative",
}

function M.start()
    shared_components.register()
    components.register()
    melee_components.register()

    -- Record interpretation happens once during composition. Systems receive a
    -- small immutable definition instead of repeatedly parsing legacy strings.
    M.fire_bolt = fire_bolt_data.load()
    cast_command.register()
    cast_system.register(M.fire_bolt)
    fire_bolt_system.register(M.fire_bolt)
    projectile_system.register()
    melee_system.register()
    monster_ai.register()
    player_melee.register()
    spawn_monster.register()
    monster_death.register()
    enter_player.register()
    transition.register()
    timed_state.register()
end

function M.stop()
    M.fire_bolt = nil
end

return M
