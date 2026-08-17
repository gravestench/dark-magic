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
        test.case("exclusive_group_replaces_state_and_exact_stat_source", {
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
                        state_id = "frozenarmor",
                        source_id = "skill:alice:40",
                        duration = 100,
                        policy = "refresh_same_source",
                        stat = "defense",
                        stat_operation = "percent",
                        stat_value = 30,
                        stat_order = 300,
                        exclusive_group = "state-group:1",
                    },
                })
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local target = ecs.query({ all = { "d2legacy.player.combat_stats" } })[1]
                ecs.create({
                    ["d2legacy.state.request"] = {
                        operation = "apply",
                        target = target,
                        state_id = "shiverarmor",
                        source_id = "skill:alice:50",
                        duration = 200,
                        policy = "refresh_same_source",
                        stat = "defense",
                        stat_operation = "percent",
                        stat_value = 45,
                        stat_order = 300,
                        exclusive_group = "state-group:1",
                    },
                })
            end),
            test.step(1),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local instances = ecs.query({ all = { "d2legacy.state.instance" } })
                local sources = ecs.query({ all = { "d2legacy.stat.source" } })
                test.assert(#instances == 1 and #sources == 1, [=[#instances == 1 and #sources == 1]=])
                test.expect(ecs.get(instances[1], "d2legacy.state.instance"):get("state_id")):equals("shiverarmor")
                test.expect(ecs.get(sources[1], "d2legacy.stat.source"):get("source_id")):equals("skill:alice:50")
                local replaced = false
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.state.event" } })) do
                    local event = ecs.get(entity, "d2legacy.state.event")
                    if event:get("reason") == "exclusive_group_replaced" then
                        replaced = event:get("state_id") == "frozenarmor"
                    end
                end
                test.assert(replaced, [=[replaced]=])
            end),
        }),
        test.case("one_state_owns_multiple_provenance_preserving_stat_sources", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local target = ecs.create({
                    ["d2legacy.player.combat_stats"] = {
                        base_attack_rating = 100,
                        base_defense = 0,
                        attack_rating = 100,
                        defense = 0,
                    },
                })
                local owner = "skill:alice:52"
                ecs.create({
                    ["d2legacy.state.request"] = {
                        operation = "apply",
                        target = target,
                        state_id = "enchant",
                        source_id = owner,
                        duration = 3,
                        policy = "refresh_same_source",
                    },
                })
                for _, source in ipairs({
                    { stat = "firemindam", operation = "add", value = 2048 },
                    { stat = "firemaxdam", operation = "add", value = 2560 },
                    { stat = "item_tohit_percent", operation = "percent", value = 20 },
                }) do
                    ecs.create({
                        ["d2legacy.state.stat_request"] = {
                            target = target,
                            owner_source_id = owner,
                            source_id = owner .. ":" .. source.stat,
                            stat = source.stat,
                            operation = source.operation,
                            value = source.value,
                            order = 300,
                        },
                    })
                end
            end),
            test.step(1),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local sources = ecs.query({ all = { "d2legacy.stat.source" } })
                test.expect(#sources):equals(3)
                local seen = {}
                for _, entity in ipairs(sources) do
                    local source = ecs.get(entity, "d2legacy.stat.source")
                    test.expect(source:get("owner_source_id")):equals("skill:alice:52")
                    seen[source:get("stat")] = source:get("value")
                end
                test.expect(seen.firemindam):equals(2048)
                test.expect(seen.firemaxdam):equals(2560)
                test.expect(seen.item_tohit_percent):equals(20)
            end),
            test.step(3),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(#ecs.query({ all = { "d2legacy.state.instance" } })):equals(0)
                test.expect(#ecs.query({ all = { "d2legacy.stat.source" } })):equals(0)
            end),
        }),
        test.case("ranked_exclusive_state_rejects_lower_and_accepts_higher_priority", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local target = ecs.create({
                    ["d2legacy.combat.defense"] = {
                        base_physical_resist = 0,
                        base_fire_resist = 0,
                    },
                })
                local function request(state, source, priority, value)
                    ecs.create({
                        ["d2legacy.state.request"] = {
                            operation = "apply",
                            target = target,
                            state_id = state,
                            source_id = source,
                            duration = 100,
                            policy = "refresh_same_source",
                            exclusive_group = "state-group:curse",
                            replacement_priority = priority,
                            reject_lower_priority = true,
                        },
                    })
                    ecs.create({
                        ["d2legacy.state.stat_request"] = {
                            target = target,
                            owner_source_id = source,
                            source_id = source .. ":physical_resist",
                            stat = "physical_resist",
                            operation = "add",
                            value = value,
                            order = 300,
                        },
                    })
                end
                request("strongcurse", "skill:alice:strong", 3, -50)
                request("same_tick_weak", "skill:bob:same-tick-weak", 2, -100)
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local target = ecs.query({ all = { "d2legacy.combat.defense" } })[1]
                ecs.create({
                    ["d2legacy.state.request"] = {
                        operation = "apply",
                        target = target,
                        state_id = "weakcurse",
                        source_id = "skill:bob:weak",
                        duration = 100,
                        policy = "refresh_same_source",
                        exclusive_group = "state-group:curse",
                        replacement_priority = 2,
                        reject_lower_priority = true,
                    },
                })
                ecs.create({
                    ["d2legacy.state.stat_request"] = {
                        target = target,
                        owner_source_id = "skill:bob:weak",
                        source_id = "skill:bob:weak:physical_resist",
                        stat = "physical_resist",
                        operation = "add",
                        value = -100,
                        order = 300,
                    },
                })
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local instance =
                    ecs.get(ecs.query({ all = { "d2legacy.state.instance" } })[1], "d2legacy.state.instance")
                test.expect(instance:get("state_id")):equals("strongcurse")
                local target = ecs.query({ all = { "d2legacy.combat.defense" } })[1]
                ecs.create({
                    ["d2legacy.state.request"] = {
                        operation = "apply",
                        target = target,
                        state_id = "strongestcurse",
                        source_id = "skill:bob:strongest",
                        duration = 100,
                        policy = "refresh_same_source",
                        exclusive_group = "state-group:curse",
                        replacement_priority = 4,
                        reject_lower_priority = true,
                    },
                })
                ecs.create({
                    ["d2legacy.state.stat_request"] = {
                        target = target,
                        owner_source_id = "skill:bob:strongest",
                        source_id = "skill:bob:strongest:physical_resist",
                        stat = "physical_resist",
                        operation = "add",
                        value = -75,
                        order = 300,
                    },
                })
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local instances = ecs.query({ all = { "d2legacy.state.instance" } })
                local sources = ecs.query({ all = { "d2legacy.stat.source" } })
                test.expect(#instances):equals(1)
                test.expect(#sources):equals(1)
                test.expect(ecs.get(instances[1], "d2legacy.state.instance"):get("state_id")):equals("strongestcurse")
                test.expect(ecs.get(sources[1], "d2legacy.stat.source"):get("value")):equals(-75)
                local rejected = false
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.state.event" } })) do
                    local event = ecs.get(entity, "d2legacy.state.event")
                    rejected = rejected or event:get("reason") == "lower_replacement_priority"
                end
                test.assert(rejected, [=[lower-priority exclusive state is explicitly rejected]=])
            end),
        }),
        test.case("cold_velocity_orders_with_skill_armor_and_item_frw_then_expires", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = require("d2legacy.tests.support.fixtures").player_entry(),
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local target = test.entities_with("d2legacy.player.identity")[1]
                for _, source in ipairs({
                    { source_id = "skill:velocity", stat = "velocitypercent", value = 20, order = 100 },
                    { source_id = "test:armor-velocity", stat = "velocitypercent", value = -10, order = 200 },
                    { source_id = "test:item-frw", stat = "item_fastermovevelocity", value = 100, order = 200 },
                }) do
                    ecs.create({
                        ["d2legacy.stat.source"] = {
                            target = target,
                            source_id = source.source_id,
                            stat = source.stat,
                            operation = "add",
                            value = source.value,
                            order = source.order,
                        },
                    })
                end
                ecs.create({
                    ["d2legacy.state.request"] = {
                        operation = "apply",
                        target = target,
                        state_id = "cold",
                        source_id = "state:cold:test",
                        duration = 4,
                        policy = "refresh_same_source",
                        stat = "velocitypercent",
                        stat_operation = "add",
                        stat_value = -50,
                        stat_order = 300,
                    },
                })
            end),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local movement =
                    ecs.get(test.entities_with("d2legacy.player.identity")[1], "d2legacy.player.movement_stats")
                test.expect(movement:get("velocitypercent")):equals(-40)
                test.expect(movement:get("item_fastermovevelocity")):equals(100)
                local rules = require("d2legacy.movement_rules/v1")
                test.expect(rules.animation_rate("Amazon", false, -40, 100)):equals(255)
            end),
            test.restore_checkpoint(),
            test.step(4),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local target = test.entities_with("d2legacy.player.identity")[1]
                local movement = ecs.get(target, "d2legacy.player.movement_stats")
                test.expect(movement:get("velocitypercent")):equals(10)
                test.expect(require("d2legacy.movement_rules/v1").animation_rate("Amazon", false, 10, 100)):equals(362)
                for _, entity in ipairs(test.entities_with("d2legacy.stat.source")) do
                    test.assert(
                        ecs.get(entity, "d2legacy.stat.source"):get("source_id") ~= "state:cold:test",
                        [=[ecs.get(entity, "d2legacy.stat.source"):get("source_id") ~= "state:cold:test"]=]
                    )
                end
            end),
        }),
    },
})
