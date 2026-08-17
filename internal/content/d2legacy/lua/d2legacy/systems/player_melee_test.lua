local test = require("d2legacy.tests/v1")

-- Reusable fixtures keep test data in one place and explain what each value means.
local fixtures = require("d2legacy.tests.support.fixtures")

-- Spawn a stationary hostile monster exactly where the player can target it.
-- This helper explains each component so the test case itself stays readable.
local function spawn_event_target_monster()
    local ecs = require("engine.ecs/v1")
    ecs.create({
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
        ["d2legacy.world.position"] = { x = 12, y = 12 },
        -- Act 1, level 1 matches the default player entry fixture.
        ["d2legacy.world.location"] = { act = 1, level_id = 1 },
        -- A small collider makes targeting feel natural without movement logic.
        ["d2legacy.world.collider"] = { radius = 1 },
        -- Selectable and hostile so player targeting can find this monster.
        ["d2legacy.world.selectable"] = {
            id = "monster:event-target",
            kind = "hostile",
            label = "Target",
            owner = "",
            radius = 1,
            priority = 1,
        },
    })
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
local function assert_attack_lifecycle_events(events)
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
            test.submit({
                tick = 2,
                sequence = 1,
                player = "alice",
                kind = "player.use_skill",
                payload = {
                    side = "left",
                    target_x = 12,
                    target_y = 12,
                    target_id = "monster:event-target",
                },
            }),
            -- Advance through enough ticks for approach, attack, and completion.
            test.step(11),
            test.run(function()
                local events = collect_attack_events()
                assert_attack_lifecycle_events(events)
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local vitals = ecs.get(player, "d2legacy.player.vitals")
                test.assert(vitals:get("mana") == 20, [=[zero-cost Attack preserves mana]=])
            end),
        }),
    },
})
