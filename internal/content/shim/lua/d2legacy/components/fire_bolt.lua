-- Data shapes used by the first migrated skill.
--
-- A component is only a named box of values. It does not contain behavior.
-- Keeping schemas here lets readers inspect saved state without reading the
-- cast, movement, collision, or damage rules.

local ecs = require("dm.ecs/v1")

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
        version = 1,
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
        version = 1,
        fields = {
            { name = "kind", type = "string" },
            { name = "tick", type = "i64" },
            { name = "attacker_id", type = "string" },
            { name = "target_id", type = "string" },
            { name = "damage_channel", type = "string" },
            { name = "damage_raw", type = "i64" },
            { name = "remaining_health_raw", type = "i64" },
        },
    })
end

return M
