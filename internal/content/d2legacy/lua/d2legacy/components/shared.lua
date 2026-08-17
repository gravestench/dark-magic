-- Engine-facing component contracts used by the first gameplay slices.
-- These are facts, not rules. Declaring them here lets a renderer-free server
-- boot d2legacy without relying on the client world builder to run first.

local ecs = require("engine.ecs/v1")
local M = {}

local function component(name, fields)
    ecs.component({ name = name, fields = fields })
end

function M.register()
    component("d2legacy.player.identity", {
        { name = "character_id", type = "string" },
        { name = "player", type = "string" },
        { name = "name", type = "string" },
        { name = "class", type = "string" },
    })
    component("d2legacy.player.vitals", {
        { name = "health", type = "i64" },
        { name = "max_health", type = "i64" },
        { name = "mana", type = "i64" },
        { name = "max_mana", type = "i64" },
        { name = "mana_raw", type = "i64" },
        { name = "max_mana_raw", type = "i64" },
    })
    component("d2legacy.player.learned_skill", {
        { name = "owner", type = "entity" },
        { name = "skill_id", type = "i64" },
        { name = "level", type = "i64" },
        { name = "list_row", type = "i64" },
        { name = "left_allowed", type = "bool" },
        { name = "right_allowed", type = "bool" },
    })
    component("d2legacy.player.skill_assignment", {
        { name = "left", type = "i64" },
        { name = "right", type = "i64" },
    })
    component("d2legacy.player.skill_intent", {
        { name = "side", type = "string" },
        { name = "skill_id", type = "i64" },
        { name = "target_x", type = "f64" },
        { name = "target_y", type = "f64" },
        { name = "target_id", type = "string" },
    })
    component("d2legacy.player.movement_mode", { { name = "running", type = "bool" } })
    component("d2legacy.player.appearance", {
        { name = "cof", type = "string" },
        { name = "token", type = "string" },
        { name = "palette", type = "string" },
        { name = "weapon_class", type = "string" },
    })
    local belt = { { name = "capacity", type = "i64" } }
    for slot = 1, 16 do
        belt[#belt + 1] = { name = "slot_" .. slot, type = "string" }
    end
    component("d2legacy.player.belt", belt)
    component("d2legacy.world.position", { { name = "x", type = "f64" }, { name = "y", type = "f64" } })
    component("d2legacy.world.velocity", { { name = "x", type = "f64" }, { name = "y", type = "f64" } })
    component("d2legacy.world.facing", {
        { name = "direction", type = "i64" },
        { name = "directions", type = "i64" },
    })
    -- Empty opt-in marker consumed by the engine's generic velocity integrator.
    component("engine.world.velocity_mover", {})
    component("d2legacy.world.player_control", { { name = "player", type = "string" } })
    component("d2legacy.player.animation", {
        { name = "direction", type = "i64" },
        { name = "mode", type = "string" },
        { name = "start_tick", type = "i64" },
    })
    component("d2legacy.presentation.animation_clock", { { name = "seconds", type = "f64" } })
    component("d2legacy.world.bounds", { { name = "width", type = "f64" }, { name = "height", type = "f64" } })
    component("d2legacy.world.location", { { name = "act", type = "i64" }, { name = "level_id", type = "i64" } })
    component("d2legacy.world.collider", { { name = "radius", type = "f64" } })
    component("d2legacy.world.selectable", {
        { name = "id", type = "string" },
        { name = "kind", type = "string" },
        { name = "label", type = "string" },
        { name = "owner", type = "string" },
        { name = "radius", type = "f64" },
        { name = "priority", type = "i64" },
    })
    component("d2legacy.monster.stats", {
        { name = "level", type = "i64" },
        { name = "spawn_player_count", type = "i64" },
        { name = "health", type = "i64" },
        { name = "max_health", type = "i64" },
        { name = "defense", type = "i64" },
        { name = "attack_rating", type = "i64" },
        { name = "physical_min", type = "i64" },
        { name = "physical_max", type = "i64" },
        { name = "experience", type = "i64" },
    })
    component("d2legacy.monster.ai", {
        { name = "behavior", type = "string" },
        { name = "state", type = "string" },
        { name = "target_id", type = "string" },
        { name = "next_think_tick", type = "i64" },
        { name = "think_interval", type = "i64" },
        { name = "aggro_radius", type = "f64" },
        { name = "attack_range", type = "f64" },
        { name = "speed", type = "f64" },
    })
    component("d2legacy.monster.identity", {
        { name = "spawn_id", type = "string" },
        { name = "definition_id", type = "string" },
        { name = "base_id", type = "string" },
        { name = "graphics_id", type = "string" },
        { name = "seed", type = "string" },
        { name = "treasure_class", type = "string" },
    })
    component("d2legacy.population.room_resident", {
        { name = "room_id", type = "string" },
    })
    component("d2legacy.monster.appearance", {
        { name = "token", type = "string" },
        { name = "mode", type = "string" },
        { name = "weapon_class", type = "string" },
        { name = "name_key", type = "string" },
        { name = "components", type = "string" },
        { name = "death_sound", type = "string" },
    })
    component("d2legacy.player.progress", {
        { name = "level", type = "i64" },
        { name = "experience", type = "i64" },
        { name = "unspent_skill_points", type = "i64" },
    })
    -- Owner-scoped, derived presentation data. Party commands and reward
    -- policy never read this component; d2legacy.party/v1 remains authoritative.
    local party_view = {
        { name = "schema_version", type = "i64" },
        { name = "revision", type = "i64" },
        { name = "party_id", type = "string" },
        { name = "roster_count", type = "i64" },
    }
    for slot = 1, 8 do
        party_view[#party_view + 1] = { name = "player_" .. slot, type = "string" }
        party_view[#party_view + 1] = { name = "name_" .. slot, type = "string" }
        party_view[#party_view + 1] = { name = "class_" .. slot, type = "string" }
        party_view[#party_view + 1] = { name = "level_" .. slot, type = "i64" }
        party_view[#party_view + 1] = { name = "relationship_" .. slot, type = "string" }
    end
    component("d2legacy.player.party_view", party_view)
    component("d2legacy.player.combat_stats", {
        { name = "base_attack_rating", type = "i64" },
        { name = "base_defense", type = "i64" },
        { name = "attack_rating", type = "i64" },
        { name = "defense", type = "i64" },
    })
    component("d2legacy.combat.defense", {
        { name = "base_physical_resist", type = "i64" },
        { name = "base_fire_resist", type = "i64" },
        { name = "physical_resist", type = "i64" },
        { name = "fire_resist", type = "i64" },
        { name = "max_fire_resist", type = "i64" },
        { name = "physical_reduction_raw", type = "i64" },
    })
    component("d2legacy.stat.source", {
        { name = "target", type = "entity" },
        { name = "source_id", type = "string" },
        { name = "stat", type = "string" },
        { name = "operation", type = "string" },
        { name = "value", type = "i64" },
        { name = "order", type = "i64" },
    })
    component("d2legacy.monster.death", {
        { name = "tick", type = "i64" },
        { name = "killer_id", type = "string" },
        { name = "credited_id", type = "string" },
        { name = "xp", type = "i64" },
        { name = "loot_seed", type = "string" },
        { name = "treasure_class", type = "string" },
        { name = "drops", type = "string" },
        { name = "game_player_count", type = "i64" },
        { name = "effective_player_count", type = "i64" },
        { name = "nearby_party_member_count", type = "i64" },
        { name = "monster_player_count", type = "i64" },
        { name = "no_drop_player_count", type = "i64" },
        { name = "active", type = "bool" },
        { name = "corpse_usable", type = "bool" },
    })
    component("d2legacy.monster.death_event", {
        { name = "kind", type = "string" },
        { name = "tick", type = "i64" },
        { name = "monster_id", type = "string" },
        { name = "killer_id", type = "string" },
        { name = "credited_id", type = "string" },
        { name = "xp", type = "i64" },
        { name = "loot_seed", type = "string" },
        { name = "treasure_class", type = "string" },
        { name = "drops", type = "string" },
        { name = "game_player_count", type = "i64" },
        { name = "effective_player_count", type = "i64" },
        { name = "nearby_party_member_count", type = "i64" },
        { name = "monster_player_count", type = "i64" },
        { name = "no_drop_player_count", type = "i64" },
    })
    component("d2legacy.state.request", {
        { name = "operation", type = "string" },
        { name = "target", type = "entity" },
        { name = "state_id", type = "string" },
        { name = "source_id", type = "string" },
        { name = "duration", type = "i64" },
        { name = "policy", type = "string" },
        { name = "stat", type = "string" },
        { name = "stat_operation", type = "string" },
        { name = "stat_value", type = "i64" },
        { name = "stat_order", type = "i64" },
    })
    component("d2legacy.state.instance", {
        { name = "target", type = "entity" },
        { name = "state_id", type = "string" },
        { name = "source_id", type = "string" },
        { name = "applied_tick", type = "i64" },
        { name = "expires_tick", type = "i64" },
        { name = "policy", type = "string" },
        { name = "stat", type = "string" },
        { name = "stat_operation", type = "string" },
        { name = "stat_value", type = "i64" },
        { name = "stat_order", type = "i64" },
    })
    component("d2legacy.state.event", {
        { name = "kind", type = "string" },
        { name = "tick", type = "i64" },
        { name = "target", type = "entity" },
        { name = "state_id", type = "string" },
        { name = "source_id", type = "string" },
        { name = "expires_tick", type = "i64" },
        { name = "reason", type = "string" },
    })
end

return M
