local test = require("d2legacy.tests/v1")

-- Reusable fixtures keep test data in one place and explain what each value means.
local fixtures = require("d2legacy.tests.support.fixtures")
local active_barrier, active_target

-- Spawn a stationary hostile monster exactly where the player can target it.
-- This helper explains each component so the test case itself stays readable.
local function spawn_event_target_monster(overrides)
    overrides = overrides or {}
    local ecs = require("engine.ecs/v1")
    return ecs.create({
        -- A stable monster identity lets us select it deterministically.
        ["d2legacy.monster.identity"] = {
            spawn_id = "event-target",
            definition_id = "target",
            base_id = "",
            graphics_id = "",
            seed = "1",
            treasure_class = "",
        },
        -- Baseline combat stats so the test focuses on animation lifecycle,
        -- not damage formulas or resistances.
        ["d2legacy.monster.stats"] = {
            level = 1,
            health = 4096,
            max_health = 4096,
            defense = 0,
            attack_rating = 0,
            physical_min = 0,
            physical_max = 0,
            experience = 0,
        },
        -- Position the monster at a fixed world coordinate the player will target.
        ["d2legacy.world.position"] = { x = overrides.x or 12, y = overrides.y or 12 },
        -- Act 1, level 1 matches the default player entry fixture.
        ["d2legacy.world.location"] = { act = 1, level_id = overrides.level_id or 1 },
        -- A small collider makes targeting feel natural without movement logic.
        ["d2legacy.world.collider"] = { radius = 1 },
        -- Selectable and hostile so player targeting can find this monster.
        ["d2legacy.world.selectable"] = {
            id = overrides.id or "monster:event-target",
            kind = overrides.kind or "hostile",
            label = "Target",
            owner = "",
            radius = 1,
            priority = 1,
        },
    })
end

local function collision_switch()
    local state = { blocked = false }
    require("d2legacy.gameplay.collision").set({
        barrier_clear = function()
            return not state.blocked
        end,
        integrate_velocity = function(_, x, y, velocity_x, velocity_y, _, _, _, elapsed)
            return x + velocity_x * elapsed, y + velocity_y * elapsed
        end,
    })
    return state
end

local function submit_attack(target_id)
    return test.submit({
        tick = 2,
        sequence = 1,
        player = "alice",
        kind = "player.use_skill",
        payload = {
            side = "left",
            target_x = 12,
            target_y = 12,
            target_id = target_id or "monster:event-target",
        },
    })
end

local function player_entity()
    return require("engine.ecs/v1").query({ all = { "d2legacy.player.identity" } })[1]
end

local function resolved_melee_events()
    return require("engine.ecs/v1").query({ all = { "d2legacy.combat.melee_event" } })
end

-- Collect all attack animation lifecycle events emitted during the scenario.
-- The test asserts their order, not just their existence.
local function collect_attack_events()
    local ecs = require("engine.ecs/v1")
    local events = ecs.query({
        all = { "d2legacy.combat.attack_animation_event" },
    })
    return events
end

-- Assert the three expected lifecycle events appeared for the right participants
-- and in the correct temporal order: start, impact, completion.
local function assert_attack_lifecycle_events(events, expected_impact_delay, expected_complete_delay)
    test.assert(#events == 3, [=[#events == 3]=])

    -- Map each event kind to its tick so we can prove ordering.
    local kinds = {}
    local ticks = {}
    for _, entity in ipairs(events) do
        local event = require("engine.ecs/v1").get(entity, "d2legacy.combat.attack_animation_event")

        -- Every lifecycle event should name the same attacker and target.
        test.assert(
            event:get("attacker_id") == "player:alice"
                and event:get("target_id") == "monster:event-target"
                and event:get("skill_id") == 0
        )

        kinds[event:get("kind")] = true
        ticks[event:get("kind")] = event:get("tick")
    end

    -- The animation should have started, hit, and finished.
    test.assert(kinds.attack_started, [=[kinds.attack_started]=])
    test.assert(kinds.attack_impact, [=[kinds.attack_impact]=])
    test.assert(kinds.attack_completed, [=[kinds.attack_completed]=])

    -- Impact must happen after start and before completion.
    test.assert(
        ticks.attack_started < ticks.attack_impact and ticks.attack_impact < ticks.attack_completed,
        [=[ticks.attack_started < ticks.attack_impact and ticks.attack_impact < ticks.attack_completed]=]
    )
    -- The synthetic AMA1HTH record has speed 128, its attack marker on frame
    -- three, and eight total frames: ceil(3*256/128)=6 ticks to impact and
    -- ceil(8*256/128)=16 ticks to completion.
    test.expect(ticks.attack_impact - ticks.attack_started):equals(expected_impact_delay or 6)
    test.expect(ticks.attack_completed - ticks.attack_started):equals(expected_complete_delay or 16)
end

return test.suite({
    profile = "authority",
    tier = "integration",
    covers = { "internal/game/action", "internal/game/skill" },
    records = {
        -- Minimal skill records so the test can issue a left-skill attack.
        -- The exact IDs match the default player entry fixture's skill slot.
        ["data/global/excel/skills.txt"] = {
            {
                Id = "0",
                skill = "Attack",
                skilldesc = "attack",
                leftskill = "1",
                general = "1",
                passive = "0",
                srvstfunc = "1",
                srvdofunc = "1",
                cltstfunc = "1",
                cltdofunc = "1",
                range = "both",
                itypea1 = "weap",
                anim = "A1",
                UseAttackRate = "1",
                TargetableOnly = "1",
                SearchEnemyXY = "1",
                AttackNoMana = "1",
                interrupt = "1",
                InGame = "1",
                mana = "0",
                minmana = "0",
                SrcDam = "128",
                weapsel = "0",
            },
            {
                Id = "36",
                skill = "Fire Bolt",
                skilldesc = "firebolt",
                leftskill = "1",
                general = "0",
                passive = "0",
                etype = "fire",
                interrupt = "1",
                srvmissile = "firebolt",
                srvdofunc = "",
                srvstfunc = "",
                mana = "5",
                manashift = "7",
                emin = "3",
                emax = "6",
                HitShift = "8",
            },
        },
        ["data/global/excel/skilldesc.txt"] = {
            { skilldesc = "attack", ListRow = "0", IconCel = "0" },
            { skilldesc = "firebolt", ListRow = "1", IconCel = "1" },
        },
    },
    cases = {
        -- Using the basic left-skill attack should emit a full animation lifecycle.
        -- This proves the player melee system admits the attack, schedules impact,
        -- and completes the animation without extra player input.
        test.case("emits_animation_lifecycle_events", {
            -- Create the player first so equipment and combat systems can query
            -- authoritative player state during the attack.
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.amazon_entry,
            }),
            test.step(1),
            -- Place a hostile monster in range so the attack can target it.
            -- Without this target, the attack would have no valid opponent.
            test.run(function()
                spawn_event_target_monster()
            end),
            -- Issue a left-skill attack toward the monster's exact position.
            -- The payload uses target_x/target_y so the runtime can path to range
            -- before admitting the attack animation.
            submit_attack(),
            -- Advance through enough ticks for approach, attack, and completion.
            test.step(19),
            test.run(function()
                local events = collect_attack_events()
                assert_attack_lifecycle_events(events)
                local ecs = require("engine.ecs/v1")
                local player = player_entity()
                local vitals = ecs.get(player, "d2legacy.player.vitals")
                test.assert(vitals:get("mana") == 20, [=[zero-cost Attack preserves mana]=])
                local resolved = resolved_melee_events()
                test.expect(#resolved):equals(1)
                local melee = ecs.get(resolved[1], "d2legacy.combat.melee_event")
                test.assert(melee:get("hit"), [=[deterministic fixture resolves the basic attack as a hit]=])
                local damage = ecs.get(resolved[1], "d2legacy.combat.event")
                test.assert(damage ~= nil, [=[successful melee composes a shared damage event on the same entity]=])
                test.assert(
                    damage:get("source_kind") == "melee"
                        and damage:get("damage_channel") == "physical"
                        and damage:get("rolled_damage_raw") >= damage:get("damage_raw")
                        and damage:get("damage_raw") == melee:get("damage_raw")
                        and damage:get("remaining_health_raw") == melee:get("remaining_health_raw"),
                    [=[melee and generic consumers observe one ordered damage result]=]
                )
                local bundle = ecs.get(resolved[1], "d2legacy.combat.damage_bundle")
                test.assert(
                    bundle:get("physical_rolled_raw") == damage:get("rolled_damage_raw")
                        and bundle:get("physical_mitigated_raw") == damage:get("damage_raw")
                        and bundle:get("fire_rolled_raw") == 0,
                    [=[melee result preserves physical separately from other channels]=]
                )
                test.assert(
                    ecs.get(resolved[1], "d2legacy.combat.death_observed") ~= nil,
                    [=[generic death consumer marks the composed result exactly once]=]
                )
            end),
        }),
        test.case("uses_resolved_attack_rate_for_the_shared_animdata_schedule", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({
                    passive_stat_sources = {
                        {
                            id = "test-ias",
                            stat = "item_fasterattackrate",
                            operation = "add",
                            value = 40,
                            order = 1,
                        },
                    },
                }),
            }),
            test.step(1),
            test.run(function()
                spawn_event_target_monster()
            end),
            submit_attack(),
            test.step(16),
            test.run(function()
                -- Synthetic AMA1HTH: EIAS(40)=30, effective speed is
                -- floor(128*130/100)=166, so frame 3 lands on tick 5 and the
                -- eight-frame action completes on tick 13.
                assert_attack_lifecycle_events(collect_attack_events(), 5, 13)
            end),
        }),
        test.case("rejects_non_opponent_targets", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.amazon_entry,
            }),
            test.step(1),
            test.run(function()
                spawn_event_target_monster({ kind = "friendly" })
            end),
            submit_attack(),
            test.step(4),
            test.run(function()
                test.expect(#collect_attack_events()):equals(0)
                local ecs = require("engine.ecs/v1")
                test.assert(
                    ecs.get(player_entity(), "d2legacy.combat.attack_approach") == nil,
                    [=[invalid alignment does not leave an approach]=]
                )
            end),
        }),
        test.case("rejects_targets_in_another_level", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.amazon_entry,
            }),
            test.step(1),
            test.run(function()
                spawn_event_target_monster({ level_id = 2 })
            end),
            submit_attack(),
            test.step(4),
            test.run(function()
                test.expect(#collect_attack_events()):equals(0)
                local ecs = require("engine.ecs/v1")
                test.assert(
                    ecs.get(player_entity(), "d2legacy.combat.attack_approach") == nil,
                    [=[other-level target does not leave an approach]=]
                )
            end),
        }),
        test.case("rejects_melee_barrier_before_animation", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.amazon_entry,
            }),
            test.step(1),
            test.run(function()
                local barrier = collision_switch()
                barrier.blocked = true
                spawn_event_target_monster()
            end),
            submit_attack(),
            test.step(4),
            test.run(function()
                test.expect(#collect_attack_events()):equals(0)
                local ecs = require("engine.ecs/v1")
                test.assert(
                    ecs.get(player_entity(), "d2legacy.combat.attack_animation") == nil,
                    [=[barrier-blocked Attack starts no animation]=]
                )
            end),
        }),
        test.case("revalidates_melee_barrier_at_impact", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.amazon_entry,
            }),
            test.step(1),
            test.run(function()
                active_barrier = collision_switch()
                spawn_event_target_monster()
            end),
            submit_attack(),
            test.step(3),
            test.run(function()
                test.expect(#collect_attack_events()):equals(1)
                active_barrier.blocked = true
            end),
            test.step(7),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local events = resolved_melee_events()
                test.expect(#events):equals(1)
                local event = ecs.get(events[1], "d2legacy.combat.melee_event")
                test.expect(event:get("target_id")):equals("")
                test.assert(not event:get("hit"), [=[barrier appearing before impact prevents the hit]=])
            end),
        }),
        test.case("revalidates_range_at_impact", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.amazon_entry,
            }),
            test.step(1),
            test.run(function()
                active_target = spawn_event_target_monster()
            end),
            submit_attack(),
            test.step(3),
            test.run(function()
                test.expect(#collect_attack_events()):equals(1)
                require("engine.ecs/v1").get(active_target, "d2legacy.world.position"):set("x", 30)
            end),
            test.step(7),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local events = resolved_melee_events()
                test.expect(#events):equals(1)
                local event = ecs.get(events[1], "d2legacy.combat.melee_event")
                test.expect(event:get("target_id")):equals("")
                test.assert(not event:get("hit"), [=[target leaving range before impact prevents the hit]=])
            end),
        }),
        test.case("revalidates_living_target_at_impact", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.amazon_entry,
            }),
            test.step(1),
            test.run(function()
                active_target = spawn_event_target_monster()
            end),
            submit_attack(),
            test.step(3),
            test.run(function()
                test.expect(#collect_attack_events()):equals(1)
                require("engine.ecs/v1").get(active_target, "d2legacy.monster.stats"):set("health", 0)
            end),
            test.step(7),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local events = resolved_melee_events()
                test.expect(#events):equals(1)
                local event = ecs.get(events[1], "d2legacy.combat.melee_event")
                test.expect(event:get("target_id")):equals("")
                test.assert(not event:get("hit"), [=[dead target cannot be hit at impact]=])
            end),
        }),
        test.case("attack_approach_owns_motion_until_explicit_locomotion_replaces_it", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.amazon_entry,
            }),
            test.step(1),
            test.run(function()
                collision_switch()
                spawn_event_target_monster({ x = 30, y = 12 })
            end),
            submit_attack(),
            test.step(3),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = player_entity()
                local motion = ecs.get(player, "d2legacy.player.motion")
                local velocity = ecs.get(player, "d2legacy.world.velocity")
                test.expect(motion:get("owner")):equals("attack_approach")
                test.expect(motion:get("kind")):equals("target")
                test.assert(motion:get("active"), [=[motion:get("active")]=])
                test.assert(
                    math.abs(velocity:get("x") - 6) < 0.000000001,
                    "attack approach bypassed the Amazon's pinned walk rate"
                )
            end),
            test.restore_checkpoint(),
            test.submit({
                tick = 5,
                sequence = 1,
                player = "alice",
                kind = "player.move",
                payload = { x = -1, y = 0, running = false },
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = player_entity()
                local motion = ecs.get(player, "d2legacy.player.motion")
                local velocity = ecs.get(player, "d2legacy.world.velocity")
                test.expect(motion:get("owner")):equals("locomotion")
                test.expect(motion:get("kind")):equals("direction")
                test.assert(velocity:get("x") < 0, "explicit locomotion did not replace attack approach")
                test.assert(
                    ecs.get(player, "d2legacy.combat.attack_approach") == nil,
                    "explicit locomotion retained attack approach ownership"
                )
            end),
        }),
    },
})
