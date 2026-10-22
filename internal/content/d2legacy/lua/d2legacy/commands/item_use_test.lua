local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")
local initial_items = {
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
    carried_gold = 870,
    stashed_gold = 0,
    items = {
        { id = "bought", code = "tsc", width = 1, height = 1, container = "inventory", x = 0, y = 0 },
        { id = "spare", code = "tsc", width = 1, height = 1, container = "inventory", x = 1, y = 0 },
        { id = "potion", code = "hp1", width = 1, height = 1, container = "inventory", x = 2, y = 0 },
    },
}

local function setup(blocked)
    local collision = require("d2legacy.gameplay.collision")
    local map = {
        blocked_position = function()
            return blocked
        end,
        line_clear = function()
            return true
        end,
    }
    collision.set(1, map)
    collision.set(2, map)
end

local function enter(level)
    return test.submit_system({
        tick = 1,
        sequence = 1,
        kind = "system.player.enter",
        payload = fixtures.player_entry({ level_id = level or 2, x = 20, y = 20 }),
    })
end

local function use(tick, sequence, id, player)
    return test.submit({
        tick = tick,
        sequence = sequence,
        player = player or "alice",
        kind = "item.use",
        payload = { item_id = id or "bought" },
    })
end

local function interact(tick, side, sequence)
    return test.submit({
        tick = tick,
        sequence = sequence or 1,
        player = "alice",
        kind = "interaction.open",
        payload = { at = true, x = side == "field" and 20 or 10, y = side == "field" and 20 or 12 },
    })
end

local function assertions(items, portals, level)
    local ecs = require("engine.ecs/v1")
    test.expect(#ecs.query({ all = { "d2legacy.item.identity" } })):equals(items)
    test.expect(#ecs.query({ all = { "d2legacy.world.town_portal" } })):equals(portals)
    local entity = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
    test.expect(ecs.get(entity, "d2legacy.world.location"):get("level_id")):equals(level)
    local _, layout = require("d2legacy.items.find").layout("alice")
    test.expect(layout:get("carried_gold")):equals(870)
end

local function field_residency(inactive)
    local ecs = require("engine.ecs/v1")
    local state = require("engine.authority_state/v1")
    local room = state.read("d2legacy.population.plan").rooms[1]
    test.expect(#room.inactive_residents):equals(inactive and 1 or 0)
    if inactive then
        local endpoints = ecs.query({ all = { "d2legacy.world.town_portal", "d2legacy.world.inactive" } })
        test.expect(#endpoints):equals(1)
        test.expect(room.inactive_residents[1].id):equals(ecs.get(endpoints[1], "d2legacy.world.selectable"):get("id"))
        test.expect(ecs.get(endpoints[1], "d2legacy.world.room_resident"):get("room_id")):equals("field")
    end
end

return test.suite({
    profile = "authority",
    tier = "integration",
    records = { ["data/global/excel/levels.txt"] = { { Id = "2", NumMon = "0" } } },
    initial_data = {
        ["d2legacy.items"] = initial_items,
        ["d2legacy.interactions"] = { owner = "alice", targets = {} },
        ["d2legacy.world_spawns"] = { ["1"] = { x = 10, y = 12, width = 100, height = 80 } },
    },
    cases = {
        test.case("exact_scroll_consumption_pair_round_trip_and_duplicate_refusal", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                player = "population",
                kind = "system.population.bootstrap",
                payload = {
                    act = 1,
                    level_id = 2,
                    links = {},
                    rooms = { { id = "field", populate = false, x = 0, y = 0, width = 100, height = 80, points = {} } },
                },
            }),
            enter(),
            test.step(1),
            test.run(function()
                setup(false)
                local ecs = require("engine.ecs/v1")
                local find = require("d2legacy.items.find")
                local layout = find.layout("alice")
                for _, id in ipairs({ "bought", "spare" }) do
                    ecs.create({
                        ["d2legacy.item.stat_modifier"] = {
                            item = find.item(layout, id),
                            source_id = id,
                            source_kind = "item",
                            stat = "strength",
                            operation = "add",
                            value = 1,
                            order = 1,
                        },
                    })
                end
            end),
            use(2, 1),
            use(2, 2),
            test.step(1),
            test.run(function()
                assertions(2, 2, 2)
                local find = require("d2legacy.items.find")
                local layout = find.layout("alice")
                test.assert(not find.item(layout, "bought"))
                test.assert(find.item(layout, "spare") and find.item(layout, "potion"))
                local ecs = require("engine.ecs/v1")
                local modifiers = ecs.query({ all = { "d2legacy.item.stat_modifier" } })
                test.expect(#modifiers):equals(1)
                test.expect(ecs.get(modifiers[1], "d2legacy.item.stat_modifier"):get("source_id")):equals("spare")
            end),
            test.expect_checkpoint_parity(2),
            test.restore_checkpoint(),
            interact(5, "field"),
            test.step(1),
            test.run(function()
                assertions(2, 2, 1)
                field_residency(true)
            end),
            test.restore_checkpoint(),
            interact(6, "town"),
            test.step(1),
            test.run(function()
                assertions(2, 0, 2)
                field_residency(false)
                local ecs = require("engine.ecs/v1")
                local entity = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
                test.expect(ecs.get(entity, "d2legacy.world.position"):get("x")):equals(20)
                test.expect(ecs.get(entity, "d2legacy.world.position"):get("y")):equals(20)
            end),
            interact(7, "town", 2),
            use(7, 3),
            test.step(1),
            test.run(function()
                assertions(2, 0, 2)
            end),
        }),
        test.case("town_cast_missing_and_wrong_item_preserve_resources", {
            enter(1),
            test.step(1),
            test.run(function()
                setup(false)
            end),
            use(2, 1),
            use(2, 2, "missing"),
            use(2, 3, "potion"),
            test.step(1),
            test.run(function()
                assertions(3, 0, 1)
            end),
        }),
        test.case("unsafe_cast_keeps_scroll_and_client_can_continue", {
            enter(),
            test.step(1),
            test.run(function()
                setup(true)
            end),
            use(2, 1),
            test.step(1),
            test.run(function()
                assertions(3, 0, 2)
                setup(false)
            end),
            use(3, 2),
            test.step(1),
            test.run(function()
                assertions(2, 2, 2)
            end),
        }),
        test.case("another_owner_cannot_consume_the_scroll", {
            enter(),
            test.step(1),
            test.run(function()
                setup(false)
            end),
            use(2, 1, "bought", "bob"),
            test.step(1),
            test.run(function()
                assertions(3, 0, 2)
            end),
        }),
        test.case("held_item_blocks_scroll_use", {
            enter(),
            test.step(1),
            test.run(function()
                setup(false)
                local ecs = require("engine.ecs/v1")
                local find = require("d2legacy.items.find")
                local layout = find.layout("alice")
                local item = find.item(layout, "potion")
                ecs.get(item, "d2legacy.item.placement"):set("container", "held")
            end),
            use(2, 1),
            test.step(1),
            test.run(function()
                assertions(3, 0, 2)
            end),
        }),
        test.case("missing_pair_refuses_traversal", {
            enter(),
            test.step(1),
            test.run(function()
                setup(false)
            end),
            use(2, 1),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.world.town_portal" } })) do
                    if ecs.get(entity, "d2legacy.world.town_portal"):get("return_endpoint") then
                        ecs.destroy(entity)
                    end
                end
            end),
            interact(3, "field", 2),
            test.step(1),
            test.run(function()
                assertions(2, 1, 2)
            end),
        }),
        test.case("dead_player_cannot_consume_a_scroll", {
            enter(),
            test.step(1),
            test.run(function()
                setup(false)
                local ecs = require("engine.ecs/v1")
                local entity = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
                ecs.get(entity, "d2legacy.player.vitals"):set("health", 0)
            end),
            use(2, 1),
            test.step(1),
            test.run(function()
                assertions(3, 0, 2)
            end),
        }),
        test.case("remote_portal_refuses_and_recast_replaces_pair", {
            enter(),
            test.step(1),
            test.run(function()
                setup(false)
            end),
            use(2, 1),
            test.step(1),
            interact(3, "town", 2),
            test.step(1),
            test.run(function()
                assertions(2, 2, 2)
                setup(true)
            end),
            use(4, 3, "spare"),
            test.step(1),
            test.run(function()
                assertions(2, 2, 2)
                local ecs = require("engine.ecs/v1")
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.world.town_portal" } })) do
                    test.assert(not ecs.get(entity, "d2legacy.world.town_portal"):get("expired"))
                end
                setup(false)
            end),
            use(5, 4, "spare"),
            test.step(1),
            test.run(function()
                assertions(1, 2, 2)
            end),
            use(6, 5),
            test.step(1),
            test.run(function()
                assertions(1, 2, 2)
            end),
        }),
    },
})
