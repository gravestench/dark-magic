local test = require("d2legacy.tests/v1")

local fixtures = require("d2legacy.tests.support.fixtures")
local monster = fixtures.monster_spawn()
return test.suite({
    profile = "authority",
    tier = "integration",
    covers = { "internal/game/combat", "internal/game/skill", "internal/game/missile" },
    cases = {
        test.case("configured_straight_missile_runs_headlessly", {
            test.submit_system(fixtures.command("system.player.enter", fixtures.straight_missile_entry, {
                tick = 1,
                sequence = 1,
            })),
            test.submit_system(fixtures.command("system.monster.spawn", monster, {
                tick = 1,
                sequence = 2,
            })),
            test.step(1),
            test.submit(fixtures.command("player.use_skill", {
                side = "left",
                target_x = 8,
                target_y = 0,
            }, {
                tick = 2,
                sequence = 1,
                player = "alice",
            })),
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
        test.case("underfunded_cast_has_no_effect_and_preserves_mana", {
            test.submit_system(fixtures.command(
                "system.player.enter",
                fixtures.player_entry({
                    level = 99,
                    x = 0,
                    y = 0,
                    mana = 2,
                    max_mana = 10,
                    skills = {
                        {
                            id = 36,
                            level = 1,
                            list_row = 0,
                            left_allowed = true,
                            right_allowed = true,
                        },
                    },
                }),
                {
                    tick = 1,
                    sequence = 1,
                }
            )),
            test.submit_system(fixtures.command("system.monster.spawn", monster, {
                tick = 1,
                sequence = 2,
            })),
            test.step(1),
            test.submit(fixtures.command("player.use_skill", {
                side = "left",
                target_x = 8,
                target_y = 0,
            }, {
                tick = 2,
                sequence = 1,
                player = "alice",
            })),
            test.step(4),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local vitals = ecs.get(player, "d2legacy.player.vitals")
                test.assert(
                    vitals:get("mana_raw") == 512 and vitals:get("mana") == 2,
                    [=[vitals:get("mana_raw") == 512 and vitals:get("mana") == 2]=]
                )
                test.assert(
                    ecs.get(player, "d2legacy.skill.cast_request") == nil
                        and ecs.get(player, "d2legacy.skill.cast") == nil,
                    [=[cast request is consumed without starting a cast]=]
                )
                test.assert(
                    #ecs.query({ all = { "d2legacy.missile.projectile" } }) == 0,
                    [=[#ecs.query({ all = { "d2legacy.missile.projectile" } }) == 0]=]
                )
                test.assert(
                    #ecs.query({ all = { "d2legacy.skill.cast_event" } }) == 0,
                    [=[#ecs.query({ all = { "d2legacy.skill.cast_event" } }) == 0]=]
                )
                local monsters = ecs.query({ all = { "d2legacy.monster.stats" } })
                test.assert(#monsters == 1, [=[#monsters == 1]=])
                test.assert(
                    ecs.get(monsters[1], "d2legacy.monster.stats"):get("health") == 4096,
                    [=[monster health remains 4096]=]
                )
            end),
        }),
    },
})
