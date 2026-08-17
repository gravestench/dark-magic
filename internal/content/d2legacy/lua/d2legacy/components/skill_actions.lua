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
        version = 1,
        fields = {
            { name = "skill_id", type = "i64" },
            { name = "skill_level", type = "i64" },
            { name = "target_x", type = "f64" },
            { name = "target_y", type = "f64" },
            { name = "target_id", type = "string" },
            { name = "effect_tick", type = "i64" },
            { name = "complete_tick", type = "i64" },
            { name = "effect_emitted", type = "bool", default = false },
        },
    })

    ecs.component({
        name = "d2legacy.missile.projectile",
        version = 2,
        fields = {
            { name = "owner_id", type = "string" },
            { name = "target_x", type = "f64" },
            { name = "target_y", type = "f64" },
            { name = "velocity_x", type = "f64" },
            { name = "velocity_y", type = "f64" },
            { name = "previous_x", type = "f64" },
            { name = "previous_y", type = "f64" },
            { name = "remaining_ticks", type = "i64" },
            { name = "collision_radius", type = "f64" },
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
end

return M
