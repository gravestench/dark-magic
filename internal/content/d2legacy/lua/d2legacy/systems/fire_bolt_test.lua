local test = require("d2legacy.tests/v1")

local fixtures = require("d2legacy.tests.support.fixtures")
local monster = fixtures.monster_spawn()
return test.suite({
    profile = "authority",
    tier = "fast",
    covers = { "internal/game/combat", "internal/game/skill", "internal/game/missile" },
    cases = {
        test.case("cast_runs_headlessly_through_lua", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.fire_bolt_entry,
            }),
            test.submit_system({
                tick = 1,
                sequence = 2,
                kind = "system.monster.spawn",
                payload = monster,
            }),
            test.step(1),
            test.submit({
                tick = 2,
                sequence = 1,
                player = "alice",
                kind = "player.use_skill",
                payload = { side = "left", target_x = 8, target_y = 0 },
            }),
            test.step(6),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local vitals = ecs.get(player, "d2legacy.player.vitals")
                test.assert(
                    vitals:get("mana_raw") == 1920 and vitals:get("mana") == 7,
                    [=[vitals:get("mana_raw") == 1920 and vitals:get("mana") == 7]=]
                )
                local monsters = ecs.query({ all = { "d2legacy.monster.stats" } })
                test.assert(#monsters == 1, [=[#monsters == 1]=])
                local health = ecs.get(monsters[1], "d2legacy.monster.stats"):get("health")
                test.assert(health < 4096 and health >= 2560, [=[health < 4096 and health >= 2560]=])
                test.assert(
                    #ecs.query({ all = { "d2legacy.missile.projectile" } }) == 0,
                    [=[#ecs.query({ all = { "d2legacy.missile.projectile" } }) == 0]=]
                )
                local events = ecs.query({ all = { "d2legacy.combat.event" } })
                test.assert(#events == 1, [=[#events == 1]=])
                local event = ecs.get(events[1], "d2legacy.combat.event")
                test.assert(
                    event:get("target_id") == "monster:fallen" and event:get("damage_channel") == "fire",
                    [=[event:get("target_id") == "monster:fallen" and event:get("damage_channel") == "fire"]=]
                )
            end),
        }),
    },
})
