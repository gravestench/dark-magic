local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")
-- The player prefix occupies seven bytes in the public player identity.
local owner = string.rep("p", 121)
local item_id = string.rep("s", 127) .. "a"

return test.suite({
    profile = "authority",
    tier = "integration",
    initial_data = {
        ["d2legacy.items"] = {
            owner = owner,
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
            carried_gold = 900,
            stashed_gold = 0,
            items = {
                { id = item_id, code = "tsc", width = 1, height = 1, container = "inventory", x = 0, y = 0 },
                {
                    id = string.rep("s", 127) .. "b",
                    code = "tsc",
                    width = 1,
                    height = 1,
                    container = "inventory",
                    x = 1,
                    y = 0,
                },
            },
        },
        ["d2legacy.interactions"] = { owner = owner, targets = {} },
        ["d2legacy.world_spawns"] = { ["1"] = { x = 10, y = 12, width = 100, height = 100 } },
    },
    cases = {
        test.case("long_item_and_owner_labels_create_a_pair", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ player = owner, level_id = 2, x = 20, y = 20 }),
            }),
            test.step(1),
            test.run(function()
                local collision = require("d2legacy.gameplay.collision")
                local map = {
                    blocked_position = function()
                        return false
                    end,
                    line_clear = function()
                        return true
                    end,
                }
                collision.set(1, map)
                collision.set(2, map)
            end),
            test.submit({ tick = 2, sequence = 1, player = owner, kind = "item.use", payload = { item_id = item_id } }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(#ecs.query({ all = { "d2legacy.world.town_portal" } })):equals(2)
            end),
        }),
    },
})
