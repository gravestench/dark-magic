local test = require("d2legacy.tests/v1")

local fixtures = require("d2legacy.tests.support.fixtures")
local monster = {
    spawn_id = "fallen",
    seed = 1,
    x = 4,
    y = 0,
    act = 1,
    level_id = 1,
    definition = {
        id = "fallen",
        base_id = "fallen",
        graphics_id = "fallen",
        name_key = "Fallen",
        ai = "fallen",
        token = "FA",
        weapon_class = "HTH",
        components = {},
        life_min = 4096,
        life_max = 4096,
        level = 1,
        defense = 0,
        attack_rating = 0,
        physical_min = 0,
        physical_max = 0,
        experience = 0,
        treasure_class = "",
        collider_radius = 0.5,
        select_radius = 0.5,
        velocity = 0,
        think_interval = 100,
        aggro_radius = 0,
        attack_range = 1,
    },
}
return test.suite({
    profile = "authority",
    tier = "fast",
    tests = {
        cast_runs_headlessly_through_lua = {
            {
                submit_system = {
                    tick = 1,
                    sequence = 1,
                    kind = "system.player.enter",
                    payload = fixtures.fire_bolt_entry,
                },
            },
            {
                submit_system = {
                    tick = 1,
                    sequence = 2,
                    kind = "system.monster.spawn",
                    payload = monster,
                },
            },
            { step = 1 },
            {
                submit = {
                    tick = 2,
                    sequence = 1,
                    player = "alice",
                    kind = "player.use_skill",
                    payload = { side = "left", target_x = 8, target_y = 0 },
                },
            },
            { step = 6 },
            {
                run = function()
                    local ecs = require("engine.ecs/v1")
                    local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                    local vitals = ecs.get(player, "d2legacy.player.vitals")
                    assert(vitals:get("mana_raw") == 1920 and vitals:get("mana") == 7)
                    local monsters = ecs.query({ all = { "d2legacy.monster.stats" } })
                    assert(#monsters == 1)
                    local health = ecs.get(monsters[1], "d2legacy.monster.stats"):get("health")
                    assert(health < 4096 and health >= 2560)
                    assert(#ecs.query({ all = { "d2legacy.missile.projectile" } }) == 0)
                    local events = ecs.query({ all = { "d2legacy.combat.event" } })
                    assert(#events == 1)
                    local event = ecs.get(events[1], "d2legacy.combat.event")
                    assert(event:get("target_id") == "monster:fallen" and event:get("damage_channel") == "fire")
                end,
            },
        },
    },
})
