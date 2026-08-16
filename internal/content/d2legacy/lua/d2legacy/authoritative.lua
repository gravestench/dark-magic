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
local leave_player = require("d2legacy.commands.leave_player")
local world_transition = require("d2legacy.systems.world_transition")
local timed_state = require("d2legacy.systems.timed_state")
local item_components = require("d2legacy.components.items")
local equipment = require("d2legacy.systems.equipment")
local item_bootstrap = require("d2legacy.items.bootstrap")
local item_commands = require("d2legacy.commands.items")
local interaction_commands = require("d2legacy.commands.interactions")
local population = require("d2legacy.bootstrap.population")
local move_player = require("d2legacy.commands.move_player")
local owned_units = require("d2legacy.commands.owned_units")
local owned_unit_component = require("d2legacy.components.owned_unit")
local facing = require("d2legacy.systems.facing")
local state_skill_data = require("d2legacy.data.state_skills")
local quest_components = require("d2legacy.components.quests")
local quest_commands = require("d2legacy.commands.quests")
local derived_stats = require("d2legacy.systems.derived_stats")
local progression_data = require("d2legacy.data.progression")
local progression = require("d2legacy.systems.progression")
local owned_unit_lifecycle = require("d2legacy.systems.owned_unit_lifecycle")
local movement = require("d2legacy.gameplay.systems.movement")
local game_rules = require("d2legacy.policy.game_rules")

local M = {
    id = "d2legacy.authoritative",
}

function M.start()
    game_rules.initialize()
    shared_components.register()
    components.register()
    melee_components.register()
    item_components.register()
    owned_unit_component.register()
    quest_components.register()
    item_bootstrap.load()

    -- Record interpretation happens once during composition. Systems receive a
    -- small immutable definition instead of repeatedly parsing legacy strings.
    M.fire_bolt = fire_bolt_data.load()
    M.progression = progression_data.load()
    M.state_skills = state_skill_data.load()
    cast_command.register(M.state_skills)
    cast_system.register(M.fire_bolt)
    fire_bolt_system.register(M.fire_bolt)
    projectile_system.register()
    melee_system.register()
    monster_ai.register()
    player_melee.register()
    spawn_monster.register()
    monster_death.register()
    progression.register(M.progression)
    enter_player.register()
    leave_player.register()
    world_transition.register()
    timed_state.register()
    equipment.register()
    derived_stats.register()
    item_commands.register()
    interaction_commands.register()
    quest_commands.register()
    population.register()
    move_player.register()
    movement.register()
    owned_units.register()
    owned_unit_lifecycle.register()
    facing.register()
end

function M.stop()
    M.fire_bolt = nil
    M.state_skills = nil
    M.progression = nil
end

return M
