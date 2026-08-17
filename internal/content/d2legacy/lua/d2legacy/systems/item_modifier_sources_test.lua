local test = require("d2legacy.tests/v1")

-- Shared test fixtures keep item and player data in one place.
local fixtures = require("d2legacy.tests.support.fixtures")

-- A weapon with two modifier sources: one affix and one socket.
-- The exact source IDs are what this test verifies.
local weapon = {
    id = "weapon",
    code = "ssd",
    width = 1.0,
    height = 1.0,
    body_slots = "rarm,larm",
    belt_eligible = false,
    base_cost = 100,
    container = "inventory",
    x = 2,
    y = 0,
    slot = "",
    belt_slot = 0,
    weapon_set = 0,
    page = 0,
    melee_range = 3,
    physical_min = 512,
    physical_max = 1024,
    melee_weapon_class = "1HS",
    attack_rating = 900,
    attack_rate = 20,
    stat_modifiers = {
        {
            source_id = "precision",
            source_kind = "affix",
            stat = "attack_rating",
            operation = "add",
            value = 75,
            order = 10,
        },
        {
            source_id = "socket-rune",
            source_kind = "socket",
            stat = "attack_rating",
            operation = "add",
            value = 25,
            order = 20,
        },
        {
            source_id = "alacrity",
            source_kind = "affix",
            stat = "item_fasterattackrate",
            operation = "add",
            value = 40,
            order = 40,
        },
        {
            source_id = "mastery",
            source_kind = "attribute",
            stat = "attack_rating",
            operation = "percent",
            value = 20,
            order = 30,
        },
    },
}

-- Armor with an affix modifier and a socket modifier.
-- This proves the equipment system preserves source provenance
-- for defensive stats too, not just weapon offense.
local armor = {
    id = "armor",
    code = "cap",
    width = 1.0,
    height = 1.0,
    body_slots = "head",
    belt_eligible = false,
    base_cost = 100,
    container = "inventory",
    x = 4,
    y = 0,
    slot = "",
    belt_slot = 0,
    weapon_set = 0,
    page = 0,
    melee_range = 0,
    physical_min = 0,
    physical_max = 0,
    melee_weapon_class = "",
    defense = 40,
    base_defense_max = 40,
    stat_modifiers = {
        {
            source_id = "sturdy",
            source_kind = "affix",
            stat = "defense",
            operation = "add",
            value = 7,
            order = 10,
        },
        {
            source_id = "socket-jewel",
            source_kind = "socket",
            stat = "defense",
            operation = "add",
            value = 3,
            order = 20,
        },
        {
            source_id = "enhanced-defense",
            source_kind = "affix",
            stat = "defense",
            operation = "local_percent",
            value = 50,
            order = 30,
        },
        {
            source_id = "armor-mastery",
            source_kind = "attribute",
            stat = "defense",
            operation = "percent",
            value = 20,
            order = 40,
        },
    },
}

-- Build the initial item database for Alice.
-- Both items start in inventory so the test can equip them deliberately.
local function initial_items()
    return {
        owner = "alice",
        inventory_width = 10,
        inventory_height = 4,
        stash_width = 6,
        stash_height = 8,
        cube_width = 3,
        cube_height = 4,
        belt_capacity = 4,
        active_weapon_set = 0,
        vendor_width = 10,
        vendor_height = 10,
        carried_gold = 1000,
        stashed_gold = 0,
        items = { weapon, armor },
    }
end

-- Collect every active stat source into a map keyed by source ID.
-- This helper avoids repeating the same ECS query and assertion pattern.
local function collect_sources()
    local ecs = require("engine.ecs/v1")
    local sources = {}
    for _, entity in ipairs(test.entities_with("d2legacy.stat.source")) do
        local source = ecs.get(entity, "d2legacy.stat.source")
        sources[source:get("source_id")] = source:get("value")
    end
    return sources
end

-- Assert that a source is present with an exact value,
-- or assert that it has been removed when expected is nil.
local function expect_source(sources, source_id, expected)
    local actual = sources[source_id]
    if expected == nil then
        test.expect(actual, source_id):is_nil()
    else
        test.expect(actual, source_id):equals(expected)
    end
end

-- After equipping the weapon and armor, all base and modifier sources should exist.
local function assert_equipped_sources(sources)
    expect_source(sources, "equipment:attack:weapon", 900)
    expect_source(sources, "equipment:attackrate:weapon", 20)
    expect_source(sources, "equipment:modifier:item_fasterattackrate:weapon:affix:40:alacrity", 40)
    expect_source(sources, "equipment:defense:armor", 61)
    expect_source(sources, "equipment:modifier:attack_rating:weapon:affix:10:precision", 75)
    expect_source(sources, "equipment:modifier:attack_rating:weapon:socket:20:socket-rune", 25)
    expect_source(sources, "equipment:modifier:attack_rating:weapon:attribute:30:mastery", 20)
    expect_source(sources, "equipment:modifier:defense:armor:affix:10:sturdy", 7)
    expect_source(sources, "equipment:modifier:defense:armor:socket:20:socket-jewel", 3)
    expect_source(sources, "equipment:modifier:defense:armor:attribute:40:armor-mastery", 20)
end

-- After unequipping the weapon only, the weapon's sources should vanish
-- while the armor's sources should remain untouched.
local function assert_unequipped_weapon_sources(sources)
    expect_source(sources, "equipment:attack:weapon", nil)
    expect_source(sources, "equipment:attackrate:weapon", nil)
    expect_source(sources, "equipment:modifier:item_fasterattackrate:weapon:affix:40:alacrity", nil)
    expect_source(sources, "equipment:modifier:attack_rating:weapon:affix:10:precision", nil)
    expect_source(sources, "equipment:modifier:attack_rating:weapon:socket:20:socket-rune", nil)
    expect_source(sources, "equipment:modifier:attack_rating:weapon:attribute:30:mastery", nil)
    expect_source(sources, "equipment:defense:armor", 61)
    expect_source(sources, "equipment:modifier:defense:armor:affix:10:sturdy", 7)
    expect_source(sources, "equipment:modifier:defense:armor:socket:20:socket-jewel", 3)
    expect_source(sources, "equipment:modifier:defense:armor:attribute:40:armor-mastery", 20)
end

-- After unequipping the armor, its sources should vanish too.
local function assert_unequipped_armor_sources(sources)
    expect_source(sources, "equipment:defense:armor", nil)
    expect_source(sources, "equipment:modifier:defense:armor:affix:10:sturdy", nil)
    expect_source(sources, "equipment:modifier:defense:armor:socket:20:socket-jewel", nil)
    expect_source(sources, "equipment:modifier:defense:armor:attribute:40:armor-mastery", nil)
end

local function assert_resolved_stats()
    local ecs = require("engine.ecs/v1")
    local players = test.entities_with("d2legacy.player.combat_stats")
    local stats = ecs.get(players[1], "d2legacy.player.combat_stats")
    local rates = ecs.get(players[1], "d2legacy.combat.action_rate")
    -- AR: (5 * (20 - 7)) + 900 + 75 + 25, then +20%.
    test.expect(stats:get("attack_rating")):equals(1278)
    -- Defense: Dex/4 + enhanced armor 61 + flat 10, then +20%.
    test.expect(stats:get("defense")):equals(91)
    test.expect(rates:get("attack_rate")):equals(120)
    test.expect(rates:get("item_fasterattackrate")):equals(40)
end

-- Move an item from one container to another.
-- Breaking this into a helper keeps each harness action short and readable.
local function move_item(tick, sequence, item_id, destination)
    return {
        tick = tick,
        sequence = sequence,
        player = "alice",
        kind = "item.move",
        payload = {
            item_id = item_id,
            destination = destination,
        },
    }
end

return test.suite({
    profile = "authority",
    tier = "fast",
    initial_data = {
        ["d2legacy.items"] = initial_items(),
    },
    cases = {
        test.case("equipped_modifier_sources_preserve_provenance_and_remove_on_unequip", {
            -- Create Alice as a fresh Amazon so equipment rules can project stats.
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry(),
            }),
            test.step(1),
            -- Pick up the weapon so it can be equipped into a hand slot.
            test.submit(move_item(2, 1, "weapon", { container = "held" })),
            test.step(1),
            -- Equip the weapon into the right-arm slot.
            test.submit(move_item(3, 2, "weapon", { container = "equipment", slot = "rarm", weapon_set = 0 })),
            test.step(1),
            -- Equip the armor into the head slot.
            test.submit(move_item(4, 3, "armor", { container = "equipment", slot = "head", weapon_set = 0 })),
            test.step(2),
            test.run(function()
                assert_equipped_sources(collect_sources())
                assert_resolved_stats()
            end),
            -- Unequip the weapon back into inventory.
            test.submit(move_item(6, 4, "weapon", { container = "inventory", x = 2, y = 0 })),
            test.step(2),
            test.run(function()
                assert_unequipped_weapon_sources(collect_sources())
            end),
            -- Unequip the armor back into inventory.
            test.submit(move_item(8, 5, "armor", { container = "inventory", x = 4, y = 0 })),
            test.step(2),
            test.run(function()
                assert_unequipped_armor_sources(collect_sources())
            end),
        }),
    },
})
