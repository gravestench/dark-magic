-- Saved facts used by basic melee resolution.
--
-- The request/profile schemas already exist in today's transitional Go
-- approach and animation mechanisms. Registering the same shapes here makes
-- their ownership visible and lets restored snapshots validate them.

local ecs = require("engine.ecs/v1")
local M = {}

function M.register()
    ecs.component({name="d2legacy.combat.basic_attack_request", fields={
        {name="target_id",type="string"},{name="request_tick",type="i64"},
    }})
    ecs.component({name="d2legacy.combat.melee_profile", fields={
        {name="range",type="f64"},{name="physical_min",type="i64"},
        {name="physical_max",type="i64"},
    }})
    ecs.component({name="d2legacy.skill.cast_event", fields={
        {name="kind",type="string"},{name="tick",type="i64"},
        {name="player",type="string"},{name="skill_id",type="i64"},
        {name="skill_level",type="i64"},{name="behavior",type="string"},
        {name="target_x",type="f64"},{name="target_y",type="f64"},
        {name="target_id",type="string"},{name="reason",type="string"},
    }})
    ecs.component({name="d2legacy.combat.attack_approach", fields={
        {name="skill_id",type="i64"},{name="target_id",type="string"},
        {name="request_tick",type="i64"},{name="target_x",type="f64"},
        {name="target_y",type="f64"},
    }})
    ecs.component({name="d2legacy.combat.attack_animation", fields={
        {name="skill_id",type="i64"},{name="target_id",type="string"},
        {name="start_tick",type="i64"},{name="impact_tick",type="i64"},
        {name="complete_tick",type="i64"},{name="impact_fired",type="bool"},
    }})
    ecs.component({name="d2legacy.combat.melee_event", fields={
        {name="kind",type="string"},{name="tick",type="i64"},
        {name="attacker_id",type="string"},{name="target_id",type="string"},
        {name="hit",type="bool"},{name="damage_raw",type="i64"},
        {name="remaining_health_raw",type="i64"},
    }})
end

return M
