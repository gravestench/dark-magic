-- Shared cast, missile, and combat-event state for migrated skill families.
--
-- Components contain checkpointed facts, never skill-specific behavior. A
-- definition selected by skill_id tells the lifecycle and missile systems how
-- to interpret these generic action records.

local ecs = require("engine.ecs/v1")

local M = {}

function M.register()
    ecs.component({
        name = "d2legacy.skill.cast_request",
        version = 1,
        fields = {
            { name = "player", type = "string" },
            { name = "skill_id", type = "i64" },
            { name = "skill_level", type = "i64" },
            { name = "target_x", type = "f64" },
            { name = "target_y", type = "f64" },
            { name = "target_id", type = "string" },
            { name = "request_tick", type = "i64" },
        },
    })

    ecs.component({
        name = "d2legacy.skill.cast",
        version = 3,
        fields = {
            { name = "skill_id", type = "i64" },
            { name = "skill_level", type = "i64" },
            { name = "target_x", type = "f64" },
            { name = "target_y", type = "f64" },
            { name = "target_id", type = "string" },
            { name = "effect_tick", type = "i64" },
            { name = "complete_tick", type = "i64" },
            { name = "elemental_damage_percent", type = "i64" },
            { name = "effect_duration_ticks", type = "i64" },
            { name = "effect_emitted", type = "bool", default = false },
        },
    })

    ecs.component({
        name = "d2legacy.missile.projectile",
        version = 5,
        fields = {
            { name = "owner_id", type = "string" },
            { name = "cast_id", type = "string" },
            { name = "target_x", type = "f64" },
            { name = "target_y", type = "f64" },
            { name = "velocity_x", type = "f64" },
            { name = "velocity_y", type = "f64" },
            { name = "previous_x", type = "f64" },
            { name = "previous_y", type = "f64" },
            { name = "remaining_ticks", type = "i64" },
            { name = "collision_radius", type = "f64" },
            { name = "destroy_on_contact", type = "bool" },
            { name = "next_hit_delay", type = "i64" },
            { name = "impact_radius", type = "f64" },
            { name = "impact_missile_id", type = "string" },
            { name = "impact_dcc", type = "string" },
            { name = "impact_palette", type = "string" },
            { name = "impact_lifetime_ticks", type = "i64" },
            { name = "impact_directions", type = "i64" },
            { name = "impact_frames_per_second", type = "i64" },
            { name = "impact_loop", type = "bool" },
            { name = "impact_sound", type = "string" },
            { name = "on_hit_state_id", type = "string" },
            { name = "on_hit_state_source_id", type = "string" },
            { name = "on_hit_state_duration", type = "i64" },
            { name = "on_hit_state_duration_policy", type = "string" },
            { name = "on_hit_state_action_disabled", type = "bool" },
            { name = "on_hit_state_exclusive_group", type = "string" },
            { name = "knockback_value", type = "i64" },
            { name = "minimum_damage_raw", type = "i64" },
            { name = "maximum_damage_raw", type = "i64" },
            { name = "damage_channel", type = "string" },
            { name = "missile_id", type = "string" },
            { name = "dcc", type = "string" },
            { name = "palette", type = "string" },
            { name = "travel_sound", type = "string" },
            { name = "hit_sound", type = "string" },
            { name = "directions", type = "i64" },
            { name = "frames_per_second", type = "i64" },
            { name = "loop", type = "bool" },
            { name = "offset_x", type = "f64" },
            { name = "offset_y", type = "f64" },
            { name = "offset_z", type = "f64" },
        },
    })

    -- Presentation-bearing aftermath is its own short-lived ECS entity. It
    -- has no damage capability, so rendering an impact cannot reapply policy.
    ecs.component({
        name = "d2legacy.missile.effect",
        version = 1,
        fields = {
            { name = "owner_id", type = "string" },
            { name = "cast_id", type = "string" },
            { name = "remaining_ticks", type = "i64" },
            { name = "missile_id", type = "string" },
            { name = "dcc", type = "string" },
            { name = "palette", type = "string" },
            { name = "travel_sound", type = "string" },
            { name = "hit_sound", type = "string" },
            { name = "directions", type = "i64" },
            { name = "logical_direction", type = "i64" },
            { name = "frames_per_second", type = "i64" },
            { name = "loop", type = "bool" },
            { name = "offset_x", type = "f64" },
            { name = "offset_y", type = "f64" },
            { name = "offset_z", type = "f64" },
        },
    })

    -- One cast/target contact lock is its own ECS entity. Radial rays share a
    -- cast ID, so overlap policy composes without a Nova-specific hit path.
    ecs.component({
        name = "d2legacy.missile.contact_lock",
        version = 1,
        fields = {
            { name = "cast_id", type = "string" },
            { name = "target_id", type = "string" },
            { name = "expires_tick", type = "i64" },
        },
    })

    ecs.component({
        name = "d2legacy.combat.event",
        version = 2,
        fields = {
            { name = "kind", type = "string" },
            { name = "tick", type = "i64" },
            { name = "attacker_id", type = "string" },
            { name = "target_id", type = "string" },
            { name = "source_kind", type = "string" },
            { name = "damage_channel", type = "string" },
            { name = "rolled_damage_raw", type = "i64" },
            { name = "damage_raw", type = "i64" },
            { name = "remaining_health_raw", type = "i64" },
        },
    })

    -- Generic attack resolution is distinct from damage commitment. Current
    -- melee produces hit/miss/invalidated outcomes; verified block/avoidance
    -- families can extend this vocabulary without inventing damage events.
    ecs.component({
        name = "d2legacy.combat.attack_result",
        version = 1,
        fields = {
            { name = "tick", type = "i64" },
            { name = "attacker_id", type = "string" },
            { name = "target_id", type = "string" },
            { name = "source_kind", type = "string" },
            { name = "outcome", type = "string" },
            { name = "attack_rating", type = "i64" },
            { name = "defense", type = "i64" },
            { name = "hit_chance", type = "i64" },
        },
    })

    ecs.component({
        name = "d2legacy.combat.damage_bundle",
        version = 1,
        fields = {
            { name = "physical_rolled_raw", type = "i64" },
            { name = "physical_mitigated_raw", type = "i64" },
            { name = "fire_rolled_raw", type = "i64" },
            { name = "fire_mitigated_raw", type = "i64" },
            { name = "lightning_rolled_raw", type = "i64" },
            { name = "lightning_mitigated_raw", type = "i64" },
            { name = "cold_rolled_raw", type = "i64" },
            { name = "cold_mitigated_raw", type = "i64" },
            { name = "magic_rolled_raw", type = "i64" },
            { name = "magic_mitigated_raw", type = "i64" },
            { name = "poison_rolled_raw", type = "i64" },
            { name = "poison_mitigated_raw", type = "i64" },
        },
    })

    -- Empty consumer marker: death attribution has observed this generic
    -- combat result. Other effect systems may compose their own independent
    -- markers without centralizing all proc policy in one event dispatcher.
    ecs.component({
        name = "d2legacy.combat.death_observed",
        version = 1,
        fields = {},
    })

    -- Empty independent consumer marker. Player-death policy can observe the
    -- same generic result as monster death without either consumer retiring or
    -- mutating the other's view of the event.
    ecs.component({
        name = "d2legacy.combat.player_death_observed",
        version = 1,
        fields = {},
    })
end

return M
