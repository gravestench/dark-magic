local test = require("d2legacy.tests/v1")

local fixtures = require("d2legacy.tests.support.fixtures")

local function enter(health)
    return test.submit_system(
        fixtures.command(
            "system.player.enter",
            fixtures.player_entry({ health = health, max_health = 50, mana = 0, max_mana = 20 }),
            { tick = 1, sequence = 1 }
        )
    )
end

-- Hit a monster with the given drain; physical and lethal cap in whole points.
local function hit(drain, physical, damage, remaining, life, mana)
    return test.run(function()
        local ecs = require("engine.ecs/v1")
        local leech = require("d2legacy.policy.leech")
        local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
        ecs.create({
            ["d2legacy.stat.source"] = {
                target = player,
                source_id = "ring:life",
                stat = "lifedrainmindam",
                value = 30,
            },
        })
        ecs.create({
            ["d2legacy.stat.source"] = {
                target = player,
                source_id = "ring:mana",
                stat = "manadrainmindam",
                value = 15,
            },
        })
        local monster = ecs.create({
            ["d2legacy.monster.stats"] = {
                level = 1,
                health = remaining * 256,
                max_health = 500 * 256,
                drain_percent = drain,
            },
        })
        local result = {
            mitigated = { physical = physical * 256 },
            damage_raw = damage * 256,
            remaining_health_raw = remaining * 256,
        }
        local sources = ecs.query({ all = { "d2legacy.stat.source" } })
        local restored_life, restored_mana = leech.apply(player, monster, result, sources, ecs)
        test.expect(restored_life):equals(life * 256)
        test.expect(restored_mana):equals(mana)
    end)
end

local function expect_vitals(health, mana_raw)
    return test.run(function()
        local ecs = require("engine.ecs/v1")
        local vitals = ecs.get(ecs.query({ all = { "d2legacy.player.identity" } })[1], "d2legacy.player.vitals")
        test.expect(vitals:get("health")):equals(health)
        test.expect(vitals:get("mana_raw")):equals(mana_raw)
        test.expect(vitals:get("mana")):equals(math.floor(mana_raw / 256))
    end)
end

return test.suite({
    profile = "authority",
    tier = "fast",
    initial_data = {
        ["d2legacy.game_rules"] = {
            target = "lod-1.14d",
            expansion = true,
            difficulty = 2,
        },
    },
    cases = {
        -- Hell divides by 3: 30% life -> 640/6400 and 15% mana -> 320/6400 of
        -- 200 physical, then halved by drain: 10 life and 1280 raw mana.
        test.case("hell_divisors_and_monster_drain_scale_leech", {
            enter(10),
            test.step(1),
            hit(50, 200, 200, 100, 10, 1280),
            expect_vitals(20, 1280),
        }),
        test.case("life_leech_stops_at_maximum_life", {
            enter(45),
            test.step(1),
            hit(50, 200, 200, 100, 5, 1280),
            expect_vitals(50, 1280),
        }),
        test.case("monsters_without_drain_cannot_be_leeched", {
            enter(10),
            test.step(1),
            hit(0, 200, 200, 100, 0, 0),
            expect_vitals(10, 0),
        }),
        -- Only the 50 life the monster had left counts: 2 life, 320 raw mana.
        test.case("physical_damage_is_capped_at_the_defender_life", {
            enter(10),
            test.step(1),
            hit(50, 200, 50, 0, 2, 320),
            expect_vitals(12, 320),
        }),
        test.case("non_physical_hits_do_not_leech", {
            enter(10),
            test.step(1),
            hit(100, 0, 200, 100, 0, 0),
            expect_vitals(10, 0),
        }),
    },
})
