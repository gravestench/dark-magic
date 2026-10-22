local test = require("d2legacy.tests/v1")

local function actor(id, player, health, level)
    local components = {
        ["d2legacy.world.selectable"] = { id = id, kind = player and "player" or "hostile", label = id },
        ["d2legacy.world.location"] = { act = 1, level_id = level or 2 },
    }
    if player then
        components["d2legacy.player.identity"] = { player = id:sub(8), name = id, class = "Amazon" }
        components["d2legacy.player.vitals"] = { health = health, max_health = health }
    else
        components["d2legacy.monster.stats"] = { health = health, max_health = health }
    end
    return require("engine.ecs/v1").create(components)
end

local function poison(target, rate, length, attacker, tick)
    require("engine.ecs/v1").create({
        ["d2legacy.state.instance"] = {
            target = target,
            state_id = "poison",
            source_id = "poison:test:" .. target:id(),
            source_attacker_id = attacker or "monster:attacker",
            applied_tick = tick or 0,
            expires_tick = (tick or 0) + length,
            policy = "refresh_same_source",
            stat = "hpregen",
            stat_operation = "add",
            stat_value = -rate,
        },
    })
end

local function vitals(id)
    local ecs = require("engine.ecs/v1")
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.world.selectable" } })) do
        if ecs.get(entity, "d2legacy.world.selectable"):get("id") == id then
            return ecs.get(entity, "d2legacy.player.vitals")
        end
    end
end

return test.suite({
    profile = "authority",
    tier = "fast",
    cases = {
        test.case("player_spends_raw_poison_rate_without_losing_fractional_units", {
            test.run(function()
                poison(actor("player:poison", true, 2), 128, 3)
            end),
            test.step(1),
            test.run(function()
                test.expect(vitals("player:poison"):get("health")):equals(2)
                test.expect(vitals("player:poison"):get("poison_remainder_raw")):equals(128)
            end),
            test.step(1),
            test.run(function()
                test.expect(vitals("player:poison"):get("health")):equals(1)
                test.expect(vitals("player:poison"):get("poison_remainder_raw")):equals(0)
            end),
            test.step(1),
            test.run(function()
                test.expect(vitals("player:poison"):get("health")):equals(1)
                test.expect(vitals("player:poison"):get("poison_remainder_raw")):equals(128)
            end),
            test.step(1),
            test.run(function()
                test.expect(vitals("player:poison"):get("health")):equals(1)
                test.expect(vitals("player:poison"):get("poison_remainder_raw")):equals(128)
            end),
        }),
        test.case("one_raw_unit_accumulates_and_monster_health_is_raw", {
            test.run(function()
                poison(actor("player:slow", true, 2), 1, 3)
                poison(actor("monster:raw", false, 512), 128, 3)
            end),
            test.step(3),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(vitals("player:slow"):get("health")):equals(2)
                test.expect(vitals("player:slow"):get("poison_remainder_raw")):equals(3)
                local monster = ecs.query({ all = { "d2legacy.monster.stats" } })[1]
                test.expect(ecs.get(monster, "d2legacy.monster.stats"):get("health")):equals(128)
            end),
        }),
        test.case("town_and_inactive_or_dead_targets_do_not_lose_life", {
            test.run(function()
                poison(actor("player:town", true, 2, 1), 128, 2)
                local inactive = actor("monster:inactive", false, 512)
                require("engine.ecs/v1").set(inactive, "d2legacy.world.inactive", {})
                poison(inactive, 128, 2)
                poison(actor("monster:dead", false, 0), 128, 2)
            end),
            test.step(3),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(vitals("player:town"):get("health")):equals(2)
                test.expect(vitals("player:town"):get("poison_remainder_raw")):equals(0)
                for _, monster in ipairs(ecs.query({ all = { "d2legacy.monster.stats" } })) do
                    local health = ecs.get(monster, "d2legacy.monster.stats"):get("health")
                    test.assert(health == 0 or health == 512, "inactive or dead monster lost health")
                end
            end),
        }),
        test.case("lethal_poison_emits_one_attributed_player_death", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = require("d2legacy.tests.support.fixtures").player_entry({ health = 1, max_health = 1 }),
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = test.entities_with("d2legacy.player.identity")[1]
                ecs.get(player, "d2legacy.world.location"):set("level_id", 2)
                poison(player, 256, 3, "monster:attacker", 1)
            end),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = test.entities_with("d2legacy.player.identity")[1]
                test.expect(ecs.get(player, "d2legacy.player.vitals"):get("health")):equals(0)
                test.expect(ecs.get(player, "d2legacy.player.death"):get("killer_id")):equals("monster:attacker")
                local count = 0
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.combat.event" } })) do
                    local event = ecs.get(entity, "d2legacy.combat.event")
                    if event:get("source_kind") == "poison" and event:get("kind") == "unit_died" then
                        count = count + 1
                        test.expect(event:get("attacker_id")):equals("monster:attacker")
                    end
                end
                test.expect(count):equals(1)
            end),
        }),
        test.case("lethal_poison_reaches_monster_death_owner_with_original_attacker", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local monster = actor("monster:poison-victim", false, 256)
                ecs.set(monster, "d2legacy.monster.identity", {
                    spawn_id = "poison-victim",
                    definition_id = "fallen",
                    seed = "test",
                    treasure_class = "",
                })
                poison(monster, 256, 3, "player:killer")
            end),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local monster = ecs.query({ all = { "d2legacy.monster.identity" } })[1]
                local death = ecs.get(monster, "d2legacy.monster.death")
                test.assert(death ~= nil, "monster death owner did not commit poison death")
                test.expect(death:get("killer_id")):equals("player:killer")
                local count = 0
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.combat.event" } })) do
                    local event = ecs.get(entity, "d2legacy.combat.event")
                    if event:get("source_kind") == "poison" and event:get("kind") == "unit_died" then
                        count = count + 1
                    end
                end
                test.expect(count):equals(1)
            end),
        }),
    },
})
