local test = require("d2legacy.tests/v1")

-- Baseline items carry modifier provenance with stable source identity.
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
    stat_modifiers = {
        { source_id = "precision", source_kind = "affix", stat = "attack_rating", operation = "add", value = 75, order = 10 },
        { source_id = "socket-rune", source_kind = "socket", stat = "attack_rating", operation = "add", value = 25, order = 20 },
    },
}

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
    stat_modifiers = {
        { source_id = "sturdy", source_kind = "affix", stat = "defense", operation = "add", value = 7, order = 10 },
        { source_id = "socket-jewel", source_kind = "socket", stat = "defense", operation = "add", value = 3, order = 20 },
    },
}

local function player_entry()
    return {
        character_id = "hero",
        player = "alice",
        name = "Hero",
        class = "Amazon",
        level = 1,
        experience = 0,
        dexterity = 20,
        defense = 0,
        health = 50,
        max_health = 50,
        mana = 20,
        max_mana = 20,
        expansion = true,
        hardcore = false,
        cof = "",
        palette = "units",
        direction = 0,
        mode = "NU",
        x = 0,
        y = 0,
        world_width = 100,
        world_height = 100,
        act = 1,
        level_id = 1,
        skills = {},
    }
end

local function expect_source_value(sources, source_id, expected)
    local actual = sources[source_id]
    if expected == nil then
        test.expect(actual, source_id):is_nil()
    else
        test.expect(actual, source_id):equals(expected)
    end
end

local function assert_equipped_sources(sources)
    expect_source_value(sources, "equipment:attack:weapon", 900)
    expect_source_value(sources, "equipment:defense:armor", 40)
    expect_source_value(sources, "equipment:modifier:attack_rating:weapon:affix:10:precision", 75)
    expect_source_value(sources, "equipment:modifier:attack_rating:weapon:socket:20:socket-rune", 25)
    expect_source_value(sources, "equipment:modifier:defense:armor:affix:10:sturdy", 7)
    expect_source_value(sources, "equipment:modifier:defense:armor:socket:20:socket-jewel", 3)
end

local function assert_unequipped_weapon_sources(sources)
    expect_source_value(sources, "equipment:attack:weapon", nil)
    expect_source_value(sources, "equipment:modifier:attack_rating:weapon:affix:10:precision", nil)
    expect_source_value(sources, "equipment:modifier:attack_rating:weapon:socket:20:socket-rune", nil)
    expect_source_value(sources, "equipment:defense:armor", 40)
    expect_source_value(sources, "equipment:modifier:defense:armor:affix:10:sturdy", 7)
    expect_source_value(sources, "equipment:modifier:defense:armor:socket:20:socket-jewel", 3)
end

local function assert_unequipped_armor_sources(sources)
    expect_source_value(sources, "equipment:defense:armor", nil)
    expect_source_value(sources, "equipment:modifier:defense:armor:affix:10:sturdy", nil)
    expect_source_value(sources, "equipment:modifier:defense:armor:socket:20:socket-jewel", nil)
end

local function stat_sources_by_id()
    local ecs = require("engine.ecs/v1")
    local result = {}
    for _, entity in ipairs(test.entities_with("d2legacy.stat.source")) do
        local source = ecs.get(entity, "d2legacy.stat.source")
        result[source:get("source_id")] = source:get("value")
    end
    return result
end

return test.suite({
    profile = "authority",
    tier = "fast",
    initial_data = {
        ["d2legacy.items"] = {
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
        },
    },
    tests = {
        equipped_modifier_sources_preserve_provenance_and_remove_on_unequip = {
            {
                submit_system = {
                    tick = 1,
                    sequence = 1,
                    kind = "system.player.enter",
                    payload = player_entry(),
                },
            },
            { step = 1 },
            {
                submit = {
                    tick = 2,
                    sequence = 1,
                    player = "alice",
                    kind = "item.move",
                    payload = { item_id = "weapon", destination = { container = "held" } },
                },
            },
            { step = 1 },
            {
                submit = {
                    tick = 3,
                    sequence = 2,
                    player = "alice",
                    kind = "item.move",
                    payload = { item_id = "weapon", place_held = true, destination = { container = "equipment", slot = "rarm", weapon_set = 0 } },
                },
            },
            { step = 1 },
            {
                submit = {
                    tick = 4,
                    sequence = 3,
                    player = "alice",
                    kind = "item.move",
                    payload = { item_id = "armor", destination = { container = "equipment", slot = "head", weapon_set = 0 } },
                },
            },
            { step = 1 },
            {
                run = function()
                    assert_equipped_sources(stat_sources_by_id())
                end,
            },
            {
                submit = {
                    tick = 5,
                    sequence = 4,
                    player = "alice",
                    kind = "item.move",
                    payload = { item_id = "weapon", destination = { container = "inventory", x = 2, y = 0 } },
                },
            },
            { step = 2 },
            {
                run = function()
                    assert_unequipped_weapon_sources(stat_sources_by_id())
                end,
            },
            {
                submit = {
                    tick = 7,
                    sequence = 5,
                    player = "alice",
                    kind = "item.move",
                    payload = { item_id = "armor", destination = { container = "inventory", x = 4, y = 0 } },
                },
            },
            { step = 2 },
            {
                run = function()
                    assert_unequipped_armor_sources(stat_sources_by_id())
                end,
            },
        },
    },
})
