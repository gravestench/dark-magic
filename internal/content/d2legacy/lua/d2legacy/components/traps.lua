-- Checkpointed facts shared by deployed devices and periodic weapon effects.
--
-- These components intentionally do not identify Assassin skills by name.
-- Exact Skills.txt rows admit content during composition; the runtime systems
-- filter and schedule entities by reusable behavior facts.

local ecs = require("engine.ecs/v1")
local M = {}

function M.register()
    -- Empty capability used by ordinary movement and monster-AI queries. A
    -- deployed device can remain a normal owned monster for life, targeting,
    -- presentation, and kill credit without accidentally acquiring chase AI.
    ecs.component({ name = "d2legacy.world.stationary", version = 1, fields = {} })

    ecs.component({
        name = "d2legacy.trap.returning_weapon",
        version = 1,
        fields = {
            { name = "owner", type = "entity" },
            { name = "target_x", type = "f64" },
            { name = "target_y", type = "f64" },
            { name = "speed_per_tick", type = "f64" },
            { name = "outbound", type = "bool" },
            { name = "expires_tick", type = "i64" },
        },
    })

    ecs.component({
        name = "d2legacy.trap.autonomous",
        version = 1,
        fields = {
            { name = "owner", type = "entity" },
            { name = "owner_id", type = "string" },
            { name = "skill_id", type = "i64" },
            { name = "skill_level", type = "i64" },
            { name = "shots_remaining", type = "i64" },
            { name = "next_fire_tick", type = "i64" },
            { name = "fire_interval", type = "i64" },
            { name = "target_radius", type = "f64" },
            { name = "cast_serial", type = "i64" },
        },
    })

    -- A selected timed state can own a periodic weapon transaction without
    -- making the state system understand radius queries or weapon damage.
    ecs.component({
        name = "d2legacy.combat.periodic_weapon",
        version = 1,
        fields = {
            { name = "owner", type = "entity" },
            { name = "owner_id", type = "string" },
            { name = "skill_id", type = "i64" },
            { name = "skill_level", type = "i64" },
            { name = "source_id", type = "string" },
            { name = "radius", type = "f64" },
            { name = "weapon_fraction", type = "i64" },
            { name = "next_tick", type = "i64" },
            { name = "period_ticks", type = "i64" },
            { name = "expires_tick", type = "i64" },
        },
    })
end

return M
