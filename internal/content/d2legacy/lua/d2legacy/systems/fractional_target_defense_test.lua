local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local weapon = {
    id = "fractional-weapon",
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
            source_id = "fractional-defense",
            source_kind = "affix",
            stat = "item_fractionaltargetac",
            operation = "add",
            value = 50,
            order = 10,
        },
    },
}

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
        items = { weapon },
    }
end

local function move_item(tick, sequence, destination)
    return {
        tick = tick,
        sequence = sequence,
        player = "alice",
        kind = "item.move",
        payload = { item_id = "fractional-weapon", destination = destination },
    }
end

local function expect_fractional_source(value)
    local ecs = require("engine.ecs/v1")
    local player = test.only_entity_with("d2legacy.player.identity")
    local sources = test.entities_with("d2legacy.stat.source")
    test.expect(require("d2legacy.gameplay.stat_sources").resolve(sources, player, "item_fractionaltargetac", 0))
        :equals(value)
    local found
    for _, entity in ipairs(sources) do
        local source = ecs.get(entity, "d2legacy.stat.source")
        if
            source:get("source_id")
            == "equipment:modifier:item_fractionaltargetac:fractional-weapon:affix:10:fractional-defense"
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

return test.suite({
    profile = "authority",
    tier = "integration",
    seed = 36,
    initial_data = { ["d2legacy.items"] = initial_items() },
    cases = {
        test.case("fractional_target_defense_source_tracks_active_weapon", {
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
                expect_fractional_source(50)
            end),
            test.submit({
                tick = 5,
                sequence = 3,
                player = "alice",
                kind = "item.weapon_set",
                payload = { set = 1 },
            }),
            test.step(2),
            test.run(function()
                expect_fractional_source(0)
            end),
            test.submit({
                tick = 7,
                sequence = 4,
                player = "alice",
                kind = "item.weapon_set",
                payload = { set = 0 },
            }),
            test.step(2),
            test.run(function()
                expect_fractional_source(50)
            end),
            test.submit(move_item(9, 5, { container = "inventory", x = 2, y = 0 })),
            test.step(2),
            test.run(function()
                expect_fractional_source(0)
            end),
        }),
    },
})
