-- Authoritative facts used by d2legacy melee resolution.
--
-- These schemas are owned here, alongside the Lua commands and systems which
-- interpret them. Go's ECS stores the values but knows nothing about attacks,
-- impact frames, melee range, hit results, or Diablo skill events.

local ecs = require("engine.ecs/v1")
local M = {}

function M.register()
    ecs.component({
        name = "d2legacy.combat.basic_attack_request",
        fields = {
            { name = "target_id", type = "string" },
            { name = "request_tick", type = "i64" },
            { name = "hand", type = "string" },
        },
    })
    ecs.component({
        name = "d2legacy.combat.melee_profile",
        fields = {
            { name = "range", type = "f64" },
            { name = "physical_min", type = "i64" },
            { name = "physical_max", type = "i64" },
            { name = "primary_attack_rating", type = "i64" },
            { name = "primary_weapon_attack_rate", type = "i64" },
            { name = "primary_hand", type = "string" },
            { name = "secondary_range", type = "f64" },
            { name = "secondary_physical_min", type = "i64" },
            { name = "secondary_physical_max", type = "i64" },
            { name = "secondary_attack_rating", type = "i64" },
            { name = "secondary_weapon_attack_rate", type = "i64" },
            { name = "secondary_hand", type = "string" },
            { name = "dual_wield", type = "bool" },
        },
    })
    ecs.component({
        name = "d2legacy.skill.cast_event",
        fields = {
            { name = "kind", type = "string" },
            { name = "tick", type = "i64" },
            { name = "player", type = "string" },
            { name = "skill_id", type = "i64" },
            { name = "skill_level", type = "i64" },
            { name = "behavior", type = "string" },
            { name = "animation_mode", type = "string" },
            { name = "target_x", type = "f64" },
            { name = "target_y", type = "f64" },
            { name = "target_id", type = "string" },
            { name = "reason", type = "string" },
            { name = "weapon_selection", type = "i64" },
        },
    })
    ecs.component({
        name = "d2legacy.combat.attack_approach",
        fields = {
            { name = "skill_id", type = "i64" },
            { name = "target_id", type = "string" },
            { name = "request_tick", type = "i64" },
            { name = "target_x", type = "f64" },
            { name = "target_y", type = "f64" },
            { name = "weapon_selection", type = "i64" },
            { name = "animation_mode", type = "string" },
        },
    })
    ecs.component({
        name = "d2legacy.combat.attack_animation",
        fields = {
            { name = "skill_id", type = "i64" },
            { name = "target_id", type = "string" },
            { name = "start_tick", type = "i64" },
            { name = "impact_tick", type = "i64" },
            { name = "complete_tick", type = "i64" },
            { name = "impact_fired", type = "bool" },
            { name = "hand", type = "string" },
            { name = "weapon_selection", type = "i64" },
            { name = "animation_mode", type = "string" },
        },
    })
    ecs.component({
        name = "d2legacy.combat.attack_animation_event",
        fields = {
            { name = "kind", type = "string" },
            { name = "tick", type = "i64" },
            { name = "attacker_id", type = "string" },
            { name = "target_id", type = "string" },
            { name = "skill_id", type = "i64" },
            { name = "hand", type = "string" },
        },
    })
    ecs.component({
        name = "d2legacy.combat.melee_event",
        version = 2,
        fields = {
            { name = "kind", type = "string" },
            { name = "tick", type = "i64" },
            { name = "attacker_id", type = "string" },
            { name = "target_id", type = "string" },
            { name = "hit", type = "bool" },
            { name = "damage_raw", type = "i64" },
            { name = "remaining_health_raw", type = "i64" },
            { name = "hand", type = "string" },
            { name = "attack_rating", type = "i64" },
            { name = "defense", type = "i64" },
            { name = "hit_chance", type = "i64" },
            { name = "outcome", type = "string" },
            { name = "defender_effects_processed", type = "bool", default = false },
        },
    })
end

return M
