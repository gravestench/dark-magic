local test = require("d2legacy.tests/v1")

local function unit(x, velocity, blocks)
    local ecs = require("engine.ecs/v1")
    return ecs.create({
        ["d2legacy.world.position"] = { x = x, y = 10 },
        ["d2legacy.world.velocity"] = { x = velocity, y = 0 },
        ["d2legacy.world.bounds"] = { width = 100, height = 100 },
        ["d2legacy.world.collider"] = { radius = 1 },
        ["d2legacy.world.occupancy"] = { blocks_movement = blocks },
        ["d2legacy.world.location"] = { act = 1, level_id = 2 },
    })
end

local function positions()
    local ecs = require("engine.ecs/v1")
    local result = {}
    for _, entity in ipairs(test.entities_with("d2legacy.world.occupancy")) do
        result[#result + 1] = ecs.get(entity, "d2legacy.world.position"):get("x")
    end
    table.sort(result)
    return result
end

return test.suite({
    profile = "authority",
    tier = "fast",
    covers = { "internal/game/world" },
    cases = {
        test.case("same_tick_contenders_cannot_enter_an_occupied_footprint", {
            test.run(function()
                unit(2, 25, true)
                unit(4, -25, true)
            end),
            test.step(1),
            test.run(function()
                local actual = positions()
                test.expect(actual[1]):equals(2)
                test.expect(actual[2]):equals(4)
            end),
            test.expect_checkpoint_parity(1),
        }),
        test.case("nonblocking_unit_policy_is_distinct_from_collider_radius", {
            test.run(function()
                unit(2, 25, true)
                unit(4, 0, false)
            end),
            test.step(1),
            test.run(function()
                local actual = positions()
                test.expect(actual[1]):equals(3)
                test.expect(actual[2]):equals(4)
            end),
        }),
        test.case("unit_can_separate_from_an_existing_spawn_overlap", {
            test.run(function()
                unit(2, 25, true)
                unit(2, 0, true)
            end),
            test.step(1),
            test.run(function()
                local actual = positions()
                test.expect(actual[1]):equals(2)
                test.expect(actual[2]):equals(3)
            end),
            test.expect_checkpoint_parity(1),
        }),
    },
})
