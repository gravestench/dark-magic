local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("coalesces_composed_combat_facts_by_entity", function(t)
            t:run(function()
                local result_entity = {
                    id = function()
                        return 41
                    end,
                }
                local animation_entity = {
                    id = function()
                        return 42
                    end,
                }
                local values = {
                    ["d2legacy.combat.attack_result"] = {
                        tick = 9,
                        attacker_id = "player:alice",
                        target_id = "monster:fallen",
                        outcome = "hit",
                    },
                    ["d2legacy.combat.event"] = {
                        tick = 9,
                        damage_channel = "physical",
                        damage_raw = 256,
                        remaining_health_raw = 768,
                    },
                    ["d2legacy.combat.damage_bundle"] = {
                        physical_rolled_raw = 512,
                        physical_mitigated_raw = 256,
                    },
                    ["d2legacy.combat.melee_event"] = { tick = 9, hit = true },
                    ["d2legacy.combat.attack_animation_event"] = {
                        tick = 8,
                        kind = "attack_impact",
                    },
                }
                test.mock_module("engine.ecs/v1", {
                    query = function(specification)
                        local component = specification.all[1]
                        if component == "d2legacy.combat.attack_animation_event" then
                            return { animation_entity }
                        end
                        if values[component] then
                            return { result_entity }
                        end
                        return {}
                    end,
                    get = function(entity, component)
                        local snapshot = entity == result_entity and values[component]
                            or entity == animation_entity and values["d2legacy.combat.attack_animation_event"]
                        if not snapshot then
                            return nil
                        end
                        return {
                            snapshot = function()
                                return snapshot
                            end,
                        }
                    end,
                }, { "query", "get" })

                local snapshot = require("d2legacy.dev.combat_lab.snapshot").read(2)
                test.expect(#snapshot.events):equals(2)
                local composed = snapshot.events[2]
                test.expect(composed.entity_id):equals(41)
                test.expect(composed.tick):equals(9)
                test.expect(composed.attack.outcome):equals("hit")
                test.expect(composed.damage.damage_raw):equals(256)
                test.expect(composed.bundle.physical_rolled_raw):equals(512)
                test.assert(composed.melee.hit, [=[composed.melee.hit]=])
            end)
        end),
    },
})
