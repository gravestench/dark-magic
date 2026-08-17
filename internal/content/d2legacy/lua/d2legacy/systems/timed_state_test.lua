local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "authority",
    tier = "fast",
    covers = { "internal/game/state" },
    cases = {
        test.case("applied_state_expires_at_the_authored_tick", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local target = ecs.create({
                    ["d2legacy.player.combat_stats"] = {
                        base_attack_rating = 0,
                        base_defense = 100,
                        attack_rating = 0,
                        defense = 100,
                    },
                })
                ecs.create({
                    ["d2legacy.state.request"] = {
                        operation = "apply",
                        target = target,
                        state_id = "poison",
                        source_id = "monster:fallen",
                        duration = 2,
                        policy = "refresh_same_source",
                        stat = "defense",
                        stat_operation = "percent",
                        stat_value = 30,
                        stat_order = 300,
                    },
                })
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.assert(
                    #ecs.query({ all = { "d2legacy.state.instance" } }) == 1,
                    [=[#ecs.query({ all = { "d2legacy.state.instance" } }) == 1]=]
                )
            end),
            test.restore_checkpoint(),
            test.step(3),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.assert(
                    #ecs.query({ all = { "d2legacy.state.instance" } }) == 0,
                    [=[#ecs.query({ all = { "d2legacy.state.instance" } }) == 0]=]
                )
                test.assert(
                    #ecs.query({ all = { "d2legacy.state.event" } }) == 2,
                    [=[#ecs.query({ all = { "d2legacy.state.event" } }) == 2]=]
                )
                test.assert(
                    #ecs.query({ all = { "d2legacy.stat.source" } }) == 0,
                    [=[#ecs.query({ all = { "d2legacy.stat.source" } }) == 0]=]
                )
                local target = ecs.query({ all = { "d2legacy.player.combat_stats" } })[1]
                test.expect(ecs.get(target, "d2legacy.player.combat_stats"):get("defense")):equals(100)
            end),
        }),
    },
})
