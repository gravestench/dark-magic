local test = require("d2legacy.tests/v1")

local fixtures = require("d2legacy.tests.support.fixtures")
local player_entity_id
local result_entity_id

return test.suite({
    profile = "authority",
    tier = "fast",
    covers = { "internal/game/combat", "internal/game/player" },
    cases = {
        test.case("player_death_is_one_checkpointed_entity_state_transition", {
            test.submit_system(fixtures.command(
                "system.player.enter",
                fixtures.player_entry({
                    health = 3,
                    max_health = 3,
                    skills = fixtures.straight_missile_entry.skills,
                }),
                {
                    tick = 1,
                    sequence = 1,
                }
            )),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local damage = require("d2legacy.policy.damage")
                local bundle = require("d2legacy.policy.damage_bundle")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local player_id = player:id()
                local owner = ecs.create({
                    ["d2legacy.world.selectable"] = {
                        id = "player:bob",
                        kind = "player",
                        label = "Bob",
                        owner = "bob",
                        radius = 0.75,
                        priority = 10,
                    },
                })
                ecs.create({
                    ["d2legacy.world.selectable"] = {
                        id = "monster:wolf",
                        kind = "friendly",
                        label = "Wolf",
                        owner = "bob",
                        radius = 0.5,
                        priority = 1,
                    },
                    ["d2legacy.owned_unit"] = {
                        owner = owner,
                        owner_id = "player:bob",
                        ultimate_owner_id = "player:bob",
                        category = "wolf",
                        group = 1,
                        limit = 1,
                        replacement = "replace_oldest",
                        created_tick = 1,
                        active = true,
                    },
                })

                local motion = ecs.get(player, "d2legacy.player.motion")
                motion:set("owner", "locomotion")
                motion:set("kind", "direction")
                motion:set("x", 1)
                motion:set("active", true)
                ecs.get(player, "d2legacy.world.velocity"):set("x", 9)
                ecs.get(player, "d2legacy.player.movement_mode"):set("running", true)
                ecs.set(player, "d2legacy.skill.cast_request", {
                    player = "alice",
                    skill_id = 36,
                    skill_level = 1,
                    target_x = 5,
                    target_y = 5,
                    target_id = "",
                    request_tick = 1,
                })
                ecs.set(player, "d2legacy.combat.attack_approach", {
                    skill_id = 0,
                    target_id = "monster:wolf",
                    request_tick = 1,
                    target_x = 5,
                    target_y = 5,
                    weapon_selection = 0,
                    animation_mode = "A1",
                })

                local result = damage.resolve(player, bundle.single("physical", 3 * 256), ecs)
                test.assert(result.lethal, "test damage did not reach lethal player health")
                local event = ecs.create({
                    ["d2legacy.combat.event"] = {
                        kind = "unit_died",
                        tick = 1,
                        attacker_id = "monster:wolf",
                        target_id = "player:alice",
                        source_kind = "melee",
                        damage_channel = "physical",
                        rolled_damage_raw = result.rolled_damage_raw,
                        damage_raw = result.damage_raw,
                        remaining_health_raw = result.remaining_health_raw,
                    },
                })
                player_entity_id = player_id
                result_entity_id = event:id()
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local players = ecs.query({ all = { "d2legacy.player.identity", "d2legacy.player.death" } })
                test.assert(
                    #players == 1 and players[1]:id() == player_entity_id,
                    "death replaced the durable player entity"
                )
                local player = players[1]
                local death = ecs.get(player, "d2legacy.player.death")
                test.assert(
                    death:get("stage") == "dead"
                        and death:get("killer_id") == "monster:wolf"
                        and death:get("credited_id") == "player:bob"
                        and death:get("hardcore") == false
                        and death:get("consequences_pending"),
                    "player death state lost attribution or unresolved consequence facts"
                )
                local motion = ecs.get(player, "d2legacy.player.motion")
                local velocity = ecs.get(player, "d2legacy.world.velocity")
                local animation = ecs.get(player, "d2legacy.player.animation")
                test.assert(not motion:get("active") and motion:get("owner") == "none", "death left locomotion active")
                test.assert(velocity:get("x") == 0 and velocity:get("y") == 0, "death left execution velocity active")
                test.assert(
                    animation:get("mode") == "DT" and animation:get("start_tick") == 2,
                    "death mode was not committed"
                )
                test.assert(
                    not ecs.get(player, "d2legacy.skill.cast_request"),
                    "death retained an admitted cast request"
                )
                test.assert(not ecs.get(player, "d2legacy.combat.attack_approach"), "death retained melee approach")
                local events = ecs.query({ all = { "d2legacy.player.death_event" } })
                test.assert(#events == 1, "player death did not emit exactly one semantic event")
                local event = ecs.get(events[1], "d2legacy.player.death_event")
                test.assert(
                    event:get("kind") == "player_died"
                        and event:get("player_id") == "player:alice"
                        and event:get("credited_id") == "player:bob"
                        and event:get("consequences_pending"),
                    "semantic player-death event disagrees with state"
                )
                local found_result = false
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.combat.event" } })) do
                    if entity:id() == result_entity_id then
                        found_result = true
                        test.assert(
                            ecs.get(entity, "d2legacy.combat.player_death_observed") ~= nil,
                            "player-death consumer did not independently mark the common result"
                        )
                    end
                end
                test.assert(found_result, "common lethal result disappeared before independent consumers completed")
            end),
            test.submit({
                tick = 3,
                sequence = 1,
                player = "alice",
                kind = "player.move",
                payload = { x = 1, y = 0, running = true },
            }),
            test.submit({
                tick = 3,
                sequence = 2,
                player = "alice",
                kind = "player.use_skill",
                payload = { side = "left", target_x = 10, target_y = 10 },
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.death" } })[1]
                test.assert(player:id() == player_entity_id, "dead command handling changed player identity")
                test.assert(
                    not ecs.get(player, "d2legacy.player.motion"):get("active"),
                    "dead player accepted movement"
                )
                test.assert(not ecs.get(player, "d2legacy.skill.cast_request"), "dead player accepted a skill cast")
                test.assert(
                    #ecs.query({ all = { "d2legacy.player.death_event" } }) == 1,
                    "death committed more than once"
                )
                player_entity_id = nil
                result_entity_id = nil
            end),
            test.expect_checkpoint_parity(1),
        }),
    },
})
