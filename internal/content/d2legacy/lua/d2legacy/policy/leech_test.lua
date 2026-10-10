local test = require("d2legacy.tests/v1")

local fixtures = require("d2legacy.tests.support.fixtures")

local function enter(health, mana)
    return test.submit_system(
        fixtures.command(
            "system.player.enter",
            fixtures.player_entry({ health = health, max_health = 50, mana = mana or 0, max_mana = 20 }),
            { tick = 1, sequence = 1 }
        )
    )
end

-- Records the effect cues a leech emits instead of creating entities.
local function recorder()
    local cues = {}
    return cues,
        {
            create = function(_, components)
                cues[#cues + 1] = components["d2legacy.presentation.effect_cue"]
            end,
        }
end

local function source(ecs, target, stat, value)
    ecs.create({
        ["d2legacy.stat.source"] = { target = target, source_id = "test:" .. stat, stat = stat, value = value },
    })
end

local function monster(ecs, health, drain)
    return ecs.create({
        ["d2legacy.monster.stats"] = {
            level = 1,
            health = health * 256,
            max_health = 500 * 256,
            drain_percent = drain,
        },
    })
end

local function result(physical, damage, remaining)
    return {
        mitigated = { physical = physical * 256 },
        damage_raw = damage * 256,
        remaining_health_raw = remaining * 256,
    }
end

-- A player with 30% life and 15% mana leech hits a monster with the given drain.
local function player_hit(drain, physical, damage, remaining, life, mana, overlay)
    return test.run(function()
        local ecs = require("engine.ecs/v1")
        local leech = require("d2legacy.policy.leech")
        local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
        source(ecs, player, "lifedrainmindam", 30)
        source(ecs, player, "manadrainmindam", 15)
        local target = monster(ecs, remaining, drain)
        local sources = ecs.query({ all = { "d2legacy.stat.source" } })
        local cues, structural = recorder()
        local restored_life, restored_mana =
            leech.apply(player, target, result(physical, damage, remaining), sources, ecs, structural, 2)
        test.expect(restored_life):equals(life * 256)
        test.expect(restored_mana):equals(mana)
        if overlay then
            test.expect(#cues):equals(1)
            test.assert(cues[1].target:id() == player:id(), "leech overlay is not on the attacker")
            test.assert(
                cues[1].overlay_id == "steallife" or cues[1].overlay_id == "stealmana",
                "leech overlay is not a steal overlay"
            )
        else
            test.expect(#cues):equals(0)
        end
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
            player_hit(50, 200, 200, 100, 10, 1280, true),
            expect_vitals(20, 1280),
        }),
        test.case("life_leech_stops_at_maximum_life", {
            enter(45),
            test.step(1),
            player_hit(50, 200, 200, 100, 5, 1280, true),
            expect_vitals(50, 1280),
        }),
        test.case("monsters_without_drain_cannot_be_leeched", {
            enter(10),
            test.step(1),
            player_hit(0, 200, 200, 100, 0, 0, false),
            expect_vitals(10, 0),
        }),
        -- Only the 50 life the monster had left counts: 2 life, 320 raw mana.
        test.case("physical_damage_is_capped_at_the_defender_life", {
            enter(10),
            test.step(1),
            player_hit(50, 200, 50, 0, 2, 320, true),
            expect_vitals(12, 320),
        }),
        test.case("non_physical_hits_do_not_leech", {
            enter(10),
            test.step(1),
            player_hit(100, 0, 200, 100, 0, 0, false),
            expect_vitals(10, 0),
        }),
        test.case("a_single_leech_stat_selects_its_overlay", {
            enter(10),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local leech = require("d2legacy.policy.leech")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                source(ecs, player, "manadrainmindam", 15)
                local cues, structural = recorder()
                leech.apply(
                    player,
                    monster(ecs, 100, 100),
                    result(200, 200, 100),
                    ecs.query({ all = { "d2legacy.stat.source" } }),
                    ecs,
                    structural,
                    2
                )
                test.expect(#cues):equals(1)
                test.expect(cues[1].overlay_id):equals("stealmana")
                test.expect(cues[1].kind):equals("leech")
            end),
        }),
        -- Hirelings skip the difficulty divisor: 30% of 200 is 60, halved to 30.
        test.case("hirelings_leech_a_percentage_without_the_divisor", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local leech = require("d2legacy.policy.leech")
                local hireling = monster(ecs, 100, 100)
                ecs.set(hireling, "d2legacy.owned_unit", {
                    owner_id = "player:alice",
                    ultimate_owner_id = "player:alice",
                    category = "hireable",
                    group = 0,
                    limit = 1,
                    replacement = "replace_oldest",
                    created_tick = 1,
                    active = true,
                })
                source(ecs, hireling, "lifedrainmindam", 30)
                local _, structural = recorder()
                local life = leech.apply(
                    hireling,
                    monster(ecs, 100, 50),
                    result(200, 200, 100),
                    ecs.query({ all = { "d2legacy.stat.source" } }),
                    ecs,
                    structural,
                    2
                )
                test.expect(life):equals(30 * 256)
                test.expect(ecs.get(hireling, "d2legacy.monster.stats"):get("health")):equals(130 * 256)
            end),
        }),
        -- A fixed 2 life and 4 mana roll is added to itself at the default
        -- skill damage share; mana is taken from the player and heals the monster.
        test.case("monsters_leech_a_flat_amount_and_drain_player_mana", {
            enter(40, 10),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local leech = require("d2legacy.policy.leech")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local attacker = monster(ecs, 100, 100)
                source(ecs, attacker, "lifedrainmindam", 2)
                source(ecs, attacker, "lifedrainmaxdam", 2)
                source(ecs, attacker, "manadrainmindam", 4)
                source(ecs, attacker, "manadrainmaxdam", 4)
                local cues, structural = recorder()
                local sources = ecs.query({ all = { "d2legacy.stat.source" } })
                local life = leech.apply(attacker, player, result(200, 200, 100), sources, ecs, structural, 2)
                test.expect(life):equals(1024 + 2048)
                test.expect(ecs.get(attacker, "d2legacy.monster.stats"):get("health")):equals(100 * 256 + 3072)
                local vitals = ecs.get(player, "d2legacy.player.vitals")
                test.expect(vitals:get("mana_raw")):equals(512)
                test.expect(vitals:get("mana")):equals(2)
                test.expect(#cues):equals(1)
                test.expect(cues[1].overlay_id):equals("steallife")
                local missile =
                    leech.apply(attacker, player, result(200, 200, 100), sources, ecs, structural, 2, "missile")
                test.expect(missile):equals(0)
            end),
        }),
    },
})
