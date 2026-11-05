local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")
local restored_drops

local function drop_identities()
    local ecs, result = require("engine.ecs/v1"), {}
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.item.identity" } })) do
        local identity = ecs.get(entity, "d2legacy.item.identity")
        result[tostring(entity:id())] = {
            id = identity:get("id"),
            owner = identity:get("owner"):id(),
            generation = ecs.get(entity, "d2legacy.item.generation"):snapshot(),
            placement = ecs.get(entity, "d2legacy.item.placement"):snapshot(),
        }
    end
    return result
end

local function expect_drop()
    local ecs = require("engine.ecs/v1")
    local entities = ecs.query({ all = { "d2legacy.item.identity" } })
    test.expect(#entities):equals(1)
    local entity = entities[1]
    local item = ecs.get(entity, "d2legacy.item.identity")
    test.expect(item:get("code")):equals("ssd")
    test.expect(ecs.get(entity, "d2legacy.item.placement"):get("container")):equals("world")
    test.expect(ecs.get(entity, "d2legacy.world.location"):get("level_id")):equals(1)
    test.expect(ecs.get(entity, "d2legacy.world.position"):get("x")):equals(12)
    test.expect(ecs.get(entity, "d2legacy.item.melee"):get("physical_min")):equals(512)
    test.expect(ecs.get(entity, "d2legacy.item.presentation"):get("world_dc6")):equals("data/global/items/flpssd.dc6")
    local generation = ecs.get(entity, "d2legacy.item.generation")
    test.expect(generation:get("level")):equals(1)
    local death = ecs.query({ all = { "d2legacy.monster.death" } })[1]
    test.expect(generation:get("facts")):equals(ecs.get(death, "d2legacy.monster.death"):get("drops"))
end

return test.suite({
    profile = "authority",
    tier = "fast",
    covers = { "internal/game/item" },
    initial_data = {
        ["d2legacy.items"] = { owner = "alice", inventory_width = 10, inventory_height = 4, items = test.array() },
    },
    records = {
        ["data/global/excel/treasureclassex.txt"] = {
            { ["Treasure Class"] = "sword", Picks = "-1", Item1 = "ssd", Prob1 = "1" },
            { ["Treasure Class"] = "double", Picks = "-2", Item1 = "ssd", Prob1 = "2" },
            { ["Treasure Class"] = "empty-family", Picks = "-1", Item1 = "weap96", Prob1 = "1" },
            { ["Treasure Class"] = "unresolved-base", Picks = "-1", Item1 = "missing-base", Prob1 = "1" },
        },
        ["data/global/excel/ItemTypes.txt"] = {
            { Code = "weap", TreasureClass = "1" },
            { Code = "swor", Equiv1 = "weap", Rarity = "1" },
        },
        ["data/global/excel/weapons.txt"] = {
            {
                code = "ssd",
                type = "swor",
                invwidth = "1",
                invheight = "3",
                mindam = "2",
                maxdam = "7",
                rangeadder = "0",
                invfile = "invssd",
                flippyfile = "flpssd",
                wclass = "1hs",
                level = "1",
                levelreq = "1",
            },
        },
        ["data/global/excel/itemratio.txt"] = {
            {
                Version = "1",
                Uber = "0",
                ["Class Specific"] = "0",
                Unique = "9999",
                Set = "9999",
                Rare = "9999",
                Magic = "9999",
                HiQuality = "9999",
                Normal = "1",
            },
        },
    },
    cases = {
        test.case("unresolved_positive_roll_is_retained_but_not_a_ground_item", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.monster.spawn",
                payload = fixtures.monster_spawn({ x = 12, y = 12 }),
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local monster = ecs.query({ all = { "d2legacy.monster.identity" } })[1]
                ecs.get(monster, "d2legacy.monster.identity"):set("treasure_class", "unresolved-base")
                ecs.get(monster, "d2legacy.monster.stats"):set("health", 0)
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local death = ecs.query({ all = { "d2legacy.monster.death" } })[1]
                local drops = ecs.get(death, "d2legacy.monster.death"):get("drops")
                test.expect(drops)
                    :equals('[{"code":"missing-base","level":1,"path":"unresolved-base","quality":"unresolved"}]')
                test.expect(#ecs.query({ all = { "d2legacy.item.identity" } })):equals(0)
            end),
            test.expect_checkpoint_parity(2),
        }),
        test.case("empty_generated_family_survives_death_and_checkpoint_without_a_fallback", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.monster.spawn",
                payload = fixtures.monster_spawn({ x = 12, y = 12 }),
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local monster = ecs.query({ all = { "d2legacy.monster.identity" } })[1]
                ecs.get(monster, "d2legacy.monster.identity"):set("treasure_class", "empty-family")
                ecs.get(monster, "d2legacy.monster.stats"):set("health", 0)
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local death = ecs.query({ all = { "d2legacy.monster.death" } })[1]
                test.expect(ecs.get(death, "d2legacy.monster.death"):get("drops")):equals("[]")
                test.expect(#ecs.query({ all = { "d2legacy.item.identity" } })):equals(0)
            end),
            test.expect_checkpoint_parity(2),
            test.restore_checkpoint(),
            test.step(2),
            test.run(function()
                test.expect(#require("engine.ecs/v1").query({ all = { "d2legacy.item.identity" } })):equals(0)
            end),
        }),
        test.case("no_admitted_player_still_uses_one_session_source", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.monster.spawn",
                payload = fixtures.monster_spawn({ x = 12, y = 12 }),
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(#ecs.query({ all = { "d2legacy.player.identity" } })):equals(0)
                local monster = ecs.query({ all = { "d2legacy.monster.identity" } })[1]
                ecs.get(monster, "d2legacy.monster.identity"):set("treasure_class", "sword")
                ecs.get(monster, "d2legacy.monster.stats"):set("health", 0)
            end),
            test.step(1),
            test.run(expect_drop),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local source = ecs.query({ all = { "d2legacy.items.world_source" } })
                test.expect(#source):equals(1)
                local item = ecs.query({ all = { "d2legacy.item.identity" } })[1]
                test.expect(ecs.get(item, "d2legacy.item.identity"):get("owner"):id()):equals(source[1]:id())
            end),
            test.step(2),
            test.expect_checkpoint_parity(2),
        }),
        test.case("multi_drop_deaths_keep_distinct_ids_after_restore", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.monster.spawn",
                payload = fixtures.monster_spawn({ spawn_id = "first", x = 12, y = 12 }),
            }),
            test.submit_system({
                tick = 1,
                sequence = 2,
                kind = "system.monster.spawn",
                payload = fixtures.monster_spawn({ spawn_id = "second", x = 14, y = 12 }),
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local monsters = ecs.query({ all = { "d2legacy.monster.identity" } })
                test.expect(#monsters):equals(2)
                for _, monster in ipairs(monsters) do
                    ecs.get(monster, "d2legacy.monster.identity"):set("treasure_class", "double")
                    ecs.get(monster, "d2legacy.monster.stats"):set("health", 0)
                end
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local items = ecs.query({ all = { "d2legacy.item.identity" } })
                test.expect(#items):equals(4)
                local ids = {}
                for _, item in ipairs(items) do
                    local id = ecs.get(item, "d2legacy.item.identity"):get("id")
                    test.assert(not ids[id], "generated drops share identity")
                    ids[id] = true
                end
                restored_drops = drop_identities()
            end),
            test.expect_checkpoint_parity(2),
            test.restore_checkpoint(),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(#ecs.query({ all = { "d2legacy.item.identity" } })):equals(4)
                test.expect(#ecs.query({ all = { "d2legacy.items.world_source" } })):equals(1)
                test.expect(drop_identities()):deep_equals(restored_drops)
            end),
        }),
        test.case("death_materializes_one_ground_drop_without_reroll", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry(),
            }),
            test.submit_system({
                tick = 1,
                sequence = 2,
                kind = "system.monster.spawn",
                payload = fixtures.monster_spawn({ x = 12, y = 12 }),
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local monster = ecs.query({ all = { "d2legacy.monster.identity" } })[1]
                ecs.get(monster, "d2legacy.monster.identity"):set("treasure_class", "sword")
                ecs.get(monster, "d2legacy.monster.stats"):set("health", 0)
            end),
            test.step(1),
            test.run(expect_drop),
            test.step(3),
            test.run(expect_drop),
            test.expect_checkpoint_parity(2),
            test.restore_checkpoint(),
            test.run(expect_drop),
        }),
    },
})
