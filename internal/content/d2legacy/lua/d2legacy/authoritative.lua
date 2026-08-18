-- First-party Diablo II gameplay composition root.
--
-- This file deliberately contains no formulas or mutations. Its only job is to
-- show which small modules make up the mod and register them in readable order.

local components = require("d2legacy.components.skill_actions")
local shared_components = require("d2legacy.components.shared")
local cast_action_data = require("d2legacy.data.cast_actions")
local missile_skill_data = require("d2legacy.data.missile_skills")
local point_movement_skill_data = require("d2legacy.data.point_movement_skills")
local radial_missile_skill_data = require("d2legacy.data.radial_missile_skills")
local melee_skill_data = require("d2legacy.data.melee_skills")
local skill_behavior_coverage = require("d2legacy.data.skill_behavior_coverage")
local cast_command = require("d2legacy.commands.cast")
local cast_system = require("d2legacy.systems.cast")
local missile_skill_system = require("d2legacy.systems.missile_skill")
local point_movement_skill_system = require("d2legacy.systems.point_movement_skill")
local melee_skill_system = require("d2legacy.systems.melee_skill")
local state_skill_system = require("d2legacy.systems.state_skill")
local reactive_state = require("d2legacy.systems.reactive_state")
local projectile_system = require("d2legacy.systems.projectile")
local melee_components = require("d2legacy.components.melee")
local melee_system = require("d2legacy.systems.melee")
local reflected_damage = require("d2legacy.systems.reflected_damage")
local monster_ai = require("d2legacy.systems.monster_ai")
local player_melee = require("d2legacy.systems.player_melee")
local spawn_monster = require("d2legacy.commands.spawn_monster")
local monster_death = require("d2legacy.systems.monster_death")
local player_death = require("d2legacy.systems.player_death")
local enter_player = require("d2legacy.commands.enter_player")
local leave_player = require("d2legacy.commands.leave_player")
local world_transition = require("d2legacy.systems.world_transition")
local world_warp = require("d2legacy.systems.world_warp")
local world_object = require("d2legacy.systems.world_object")
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
local targeted_state_skill_data = require("d2legacy.data.targeted_state_skills")
local targeted_state_skill_system = require("d2legacy.systems.targeted_state_skill")
local area_curse_skill_data = require("d2legacy.data.area_curse_skills")
local area_curse_skill_system = require("d2legacy.systems.area_curse_skill")
local aura_skill_data = require("d2legacy.data.aura_skills")
local periodic_aura_skill_data = require("d2legacy.data.periodic_aura_skills")
local corpse_aura_skill_data = require("d2legacy.data.corpse_aura_skills")
local corpse_summon_skill_data = require("d2legacy.data.corpse_summon_skills")
local corpse_summon_skill_system = require("d2legacy.systems.corpse_summon_skill")
local golem_summon_skill_data = require("d2legacy.data.golem_summon_skills")
local golem_summon_skill_system = require("d2legacy.systems.golem_summon_skill")
local summon_effects = require("d2legacy.systems.summon_effects")
local aura_skill_system = require("d2legacy.systems.aura_skill")
local aura_pulse_system = require("d2legacy.systems.aura_pulse")
local learned_passive_skill_system = require("d2legacy.systems.learned_passive_skill")
local quest_components = require("d2legacy.components.quests")
local quest_commands = require("d2legacy.commands.quests")
local derived_stats = require("d2legacy.systems.derived_stats")
local stamina = require("d2legacy.systems.stamina")
local mana = require("d2legacy.systems.mana")
local player_motion = require("d2legacy.systems.player_motion")
local progression_data = require("d2legacy.data.progression")
local progression = require("d2legacy.systems.progression")
local maximum_stamina = require("d2legacy.systems.maximum_stamina")
local environment = require("d2legacy.systems.environment")
local owned_unit_lifecycle = require("d2legacy.systems.owned_unit_lifecycle")
local movement = require("d2legacy.gameplay.systems.movement")
local forced_motion = require("d2legacy.gameplay.systems.forced_motion")
local game_rules = require("d2legacy.policy.game_rules")
local player_count = require("d2legacy.policy.player_count")
local player_count_commands = require("d2legacy.commands.player_count")
local party = require("d2legacy.policy.party")
local party_commands = require("d2legacy.commands.party")
local party_projection = require("d2legacy.systems.party_projection")

local M = {
    id = "d2legacy.authoritative",
}

function M.start()
    game_rules.initialize()
    player_count.initialize()
    party.initialize()
    shared_components.register()
    components.register()
    melee_components.register()
    item_components.register()
    owned_unit_component.register()
    quest_components.register()
    item_bootstrap.load()

    -- Record interpretation happens once during composition. Systems receive a
    -- small immutable definition instead of repeatedly parsing legacy strings.
    -- Coverage is explicit and target-locked in one reviewed manifest; sharing
    -- record function IDs never enables another skill by resemblance.
    M.skill_behavior_coverage = skill_behavior_coverage.load()
    M.missile_skills = missile_skill_data.load(M.skill_behavior_coverage.by_family["missile.straight"] or {})
    M.impact_missile_skills = missile_skill_data.load(
        M.skill_behavior_coverage.by_family["missile.straight-impact-area"] or {},
        "missile.straight-impact-area"
    )
    M.freeze_missile_skills = missile_skill_data.load(
        M.skill_behavior_coverage.by_family["missile.straight-freeze"] or {},
        "missile.straight-freeze"
    )
    M.area_freeze_missile_skills = missile_skill_data.load(
        M.skill_behavior_coverage.by_family["missile.straight-impact-area-freeze"] or {},
        "missile.straight-impact-area-freeze"
    )
    M.radial_missile_skills =
        radial_missile_skill_data.load(M.skill_behavior_coverage.by_family["missile.radial"] or {})
    M.melee_skills = melee_skill_data.load(M.skill_behavior_coverage.by_family["action.melee"] or {})
    M.point_movement_skills =
        point_movement_skill_data.load(M.skill_behavior_coverage.by_family["movement.point-relocate"] or {})
    M.progression = progression_data.load()
    M.state_skills = state_skill_data.load(M.skill_behavior_coverage.by_family["state.self-timed"] or {})
    M.targeted_state_skills =
        targeted_state_skill_data.load(M.skill_behavior_coverage.by_family["state.targeted-timed"] or {})
    M.area_curse_skills =
        area_curse_skill_data.load(M.skill_behavior_coverage.by_family["state.point-area-curse"] or {})
    M.aura_skills = aura_skill_data.load(M.skill_behavior_coverage.by_family["aura.selected-party-stat"] or {})
    M.periodic_aura_skills, M.periodic_aura_state_policy =
        periodic_aura_skill_data.load(M.skill_behavior_coverage.by_family["aura.selected-party-periodic"] or {})
    for skill_id, definition in pairs(M.periodic_aura_skills) do
        assert(not M.aura_skills[skill_id], "skill has multiple selected-aura behavior declarations")
        M.aura_skills[skill_id] = definition
    end
    M.corpse_aura_skills =
        corpse_aura_skill_data.load(M.skill_behavior_coverage.by_family["aura.selected-corpse-periodic"] or {})
    for skill_id, definition in pairs(M.corpse_aura_skills) do
        assert(not M.aura_skills[skill_id], "skill has multiple selected-aura behavior declarations")
        M.aura_skills[skill_id] = definition
    end
    M.corpse_summon_skills =
        corpse_summon_skill_data.load(M.skill_behavior_coverage.by_family["summon.targeted-corpse"] or {})
    M.golem_summon_skills = golem_summon_skill_data.load(M.skill_behavior_coverage.by_family["summon.golem"] or {})
    M.cast_skills = {}
    for skill_id, definition in pairs(M.missile_skills) do
        M.cast_skills[skill_id] = definition
    end
    for skill_id, definition in pairs(M.radial_missile_skills) do
        assert(not M.cast_skills[skill_id], "skill has multiple behavior-family declarations")
        M.cast_skills[skill_id] = definition
    end
    for skill_id, definition in pairs(M.impact_missile_skills) do
        assert(not M.cast_skills[skill_id], "skill has multiple behavior-family declarations")
        M.cast_skills[skill_id] = definition
    end
    for skill_id, definition in pairs(M.freeze_missile_skills) do
        assert(not M.cast_skills[skill_id], "skill has multiple behavior-family declarations")
        M.cast_skills[skill_id] = definition
    end
    for skill_id, definition in pairs(M.area_freeze_missile_skills) do
        assert(not M.cast_skills[skill_id], "skill has multiple behavior-family declarations")
        M.cast_skills[skill_id] = definition
    end
    for skill_id, definition in pairs(M.state_skills) do
        assert(not M.cast_skills[skill_id], "skill has multiple behavior-family declarations")
        M.cast_skills[skill_id] = definition
    end
    for skill_id, definition in pairs(M.melee_skills) do
        assert(not M.cast_skills[skill_id], "skill has multiple behavior-family declarations")
        M.cast_skills[skill_id] = definition
    end
    for skill_id, definition in pairs(M.point_movement_skills) do
        assert(not M.cast_skills[skill_id], "skill has multiple behavior-family declarations")
        M.cast_skills[skill_id] = definition
    end
    for skill_id, definition in pairs(M.targeted_state_skills) do
        assert(not M.cast_skills[skill_id], "skill has multiple behavior-family declarations")
        M.cast_skills[skill_id] = definition
    end
    for skill_id, definition in pairs(M.area_curse_skills) do
        assert(not M.cast_skills[skill_id], "skill has multiple behavior-family declarations")
        M.cast_skills[skill_id] = definition
    end
    for skill_id, definition in pairs(M.aura_skills) do
        assert(not M.cast_skills[skill_id], "skill has multiple behavior-family declarations")
        M.cast_skills[skill_id] = definition
    end
    for skill_id, definition in pairs(M.corpse_summon_skills) do
        assert(not M.cast_skills[skill_id], "skill has multiple behavior-family declarations")
        M.cast_skills[skill_id] = definition
    end
    for skill_id, definition in pairs(M.golem_summon_skills) do
        assert(not M.cast_skills[skill_id], "skill has multiple behavior-family declarations")
        M.cast_skills[skill_id] = definition
    end
    M.cast_actions = cast_action_data.load(M.cast_skills)
    cast_command.register(M.cast_skills)
    cast_system.register(M.cast_skills, M.cast_actions)
    melee_skill_system.register(M.melee_skills)
    missile_skill_system.register(M.cast_skills)
    point_movement_skill_system.register(M.point_movement_skills)
    state_skill_system.register(M.state_skills)
    targeted_state_skill_system.register(M.targeted_state_skills)
    area_curse_skill_system.register(M.area_curse_skills)
    aura_skill_system.register(M.aura_skills)
    aura_pulse_system.register(M.periodic_aura_state_policy)
    corpse_summon_skill_system.register(M.corpse_summon_skills)
    golem_summon_skill_system.register(M.golem_summon_skills)
    summon_effects.register()
    learned_passive_skill_system.register(M.aura_skills)
    projectile_system.register()
    melee_system.register()
    monster_ai.register()
    player_melee.register()
    spawn_monster.register()
    reflected_damage.register()
    monster_death.register()
    player_death.register()
    progression.register(M.progression)
    environment.register()
    maximum_stamina.register()
    enter_player.register()
    leave_player.register()
    world_transition.register()
    reactive_state.register(M.state_skills)
    timed_state.register()
    equipment.register()
    derived_stats.register()
    mana.register()
    player_motion.register()
    stamina.register()
    item_commands.register()
    interaction_commands.register()
    world_warp.register()
    world_object.register()
    quest_commands.register()
    population.register()
    party_commands.register()
    player_count_commands.register()
    party_projection.register()
    move_player.register()
    forced_motion.register()
    movement.register()
    owned_units.register()
    owned_unit_lifecycle.register()
    facing.register()
end

function M.stop()
    M.missile_skills = nil
    M.radial_missile_skills = nil
    M.impact_missile_skills = nil
    M.freeze_missile_skills = nil
    M.area_freeze_missile_skills = nil
    M.melee_skills = nil
    M.point_movement_skills = nil
    M.state_skills = nil
    M.targeted_state_skills = nil
    M.area_curse_skills = nil
    M.aura_skills = nil
    M.periodic_aura_skills = nil
    M.periodic_aura_state_policy = nil
    M.corpse_summon_skills = nil
    M.golem_summon_skills = nil
    M.cast_skills = nil
    M.cast_actions = nil
    M.skill_behavior_coverage = nil
    M.progression = nil
end

return M
