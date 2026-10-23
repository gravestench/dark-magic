local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local weapon = {
    id = "target-type-weapon",
    code = "ssd",
    width = 1,
    height = 1,
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
    attack_rating = 0,
    attack_rate = 0,
    stat_modifiers = {
        {
            source_id = "demon-ar",
            source_kind = "affix",
            stat = "item_demon_tohit",
            operation = "add",
            value = 50,
            order = 10,
        },
        {
            source_id = "undead-ar",
            source_kind = "affix",
            stat = "item_undead_tohit",
            operation = "add",
            value = 40,
            order = 11,
        },
    },
}

local function move_item(tick, sequence, destination)
    return {
        tick = tick,
        sequence = sequence,
        player = "alice",
        kind = "item.move",
        payload = { item_id = weapon.id, destination = destination },
    }
end

local function expect_sources(demon, undead)
    local ecs = require("engine.ecs/v1")
    local resolve = require("d2legacy.gameplay.stat_sources").resolve
    local player = test.only_entity_with("d2legacy.player.identity")
    local entities = test.entities_with("d2legacy.stat.source")
    local function expect_source(stat, source_id, order, value)
        test.expect(resolve(entities, player, stat, 0)):equals(value)
        local found
        for _, entity in ipairs(entities) do
            local source = ecs.get(entity, "d2legacy.stat.source")
            if
                source:get("source_id")
                == "equipment:modifier:" .. stat .. ":" .. weapon.id .. ":affix:" .. order .. ":" .. source_id
            then
                found = source:get("value")
            end
        end
        if value == 0 then
            test.expect(found):is_nil()
        else
            test.expect(found):equals(value)
        end
    end
    expect_source("item_demon_tohit", "demon-ar", 10, demon)
    expect_source("item_undead_tohit", "undead-ar", 11, undead)
    test.expect(resolve(entities, player, "item_ignoretargetac", 0)):equals(0)
    test.expect(resolve(entities, player, "item_fractionaltargetac", 0)):equals(0)
end

return test.suite({
    profile = "authority",
    tier = "integration",
    seed = 429,
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
            items = { weapon },
        },
    },
    cases = {
        test.case("target_type_attack_rating_sources_track_active_weapon", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry(),
            }),
            test.step(1),
            test.submit(move_item(2, 1, { container = "held" })),
            test.step(1),
            test.submit(move_item(3, 2, { container = "equipment", slot = "rarm", weapon_set = 0 })),
            test.step(2),
            test.run(function()
                expect_sources(50, 40)
            end),
            test.submit({ tick = 6, sequence = 3, player = "alice", kind = "item.weapon_set", payload = { set = 1 } }),
            test.step(2),
            test.run(function()
                expect_sources(0, 0)
            end),
            test.submit({ tick = 7, sequence = 4, player = "alice", kind = "item.weapon_set", payload = { set = 0 } }),
            test.step(2),
            test.run(function()
                expect_sources(50, 40)
            end),
            test.submit(move_item(9, 5, { container = "inventory", x = 2, y = 0 })),
            test.step(2),
            test.run(function()
                expect_sources(0, 0)
            end),
        }),
    },
})
