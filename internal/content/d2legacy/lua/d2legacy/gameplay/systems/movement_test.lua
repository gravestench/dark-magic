local test = require("d2legacy.tests/v1")

local function unit(x, velocity, blocks, request)
    local ecs = require("engine.ecs/v1")
    local components = {
        ["d2legacy.world.position"] = { x = x, y = 10 },
        ["d2legacy.world.velocity"] = { x = velocity, y = 0 },
        ["d2legacy.world.bounds"] = { width = 100, height = 100 },
        ["d2legacy.world.collider"] = { radius = 1 },
        ["d2legacy.world.occupancy"] = { blocks_movement = blocks },
        ["d2legacy.world.location"] = { act = 1, level_id = 2 },
    }
    if request then
        components["d2legacy.world.forced_motion_request"] = request
    end
    return ecs.create(components)
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

local function forced_event()
    local ecs = require("engine.ecs/v1")
    local events = ecs.query({ all = { "d2legacy.world.forced_motion_event" } })
    test.expect(#events):equals(1)
    return ecs.get(events[1], "d2legacy.world.forced_motion_event")
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
        test.case("semantic_forced_motion_completes_over_multiple_ticks", {
            test.run(function()
                unit(4, 0, true, {
                    kind = "knockback",
                    source_x = 2,
                    source_y = 10,
                    distance = 2,
                    speed = 25,
                    request_tick = 0,
                })
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local movers = ecs.query({ all = { "d2legacy.world.forced_motion" } })
                test.expect(positions()[1]):equals(5)
                test.expect(#movers):equals(1)
                test.expect(ecs.get(movers[1], "d2legacy.world.forced_motion"):get("applied_distance")):equals(1)
            end),
            test.expect_checkpoint_parity(1),
            test.run(function()
                test.expect(positions()[1]):equals(6)
                local event = forced_event()
                test.expect(event:get("kind")):equals("knockback")
                test.expect(event:get("outcome")):equals("completed")
                test.expect(event:get("applied_distance")):equals(2)
            end),
        }),
        test.case("forced_motion_reports_a_fully_blocked_footprint", {
            test.run(function()
                unit(2, 0, true, {
                    kind = "knockback",
                    source_x = 0,
                    source_y = 10,
                    distance = 3,
                    speed = 25,
                    request_tick = 0,
                })
                unit(4, 0, true)
            end),
            test.step(1),
            test.run(function()
                local actual = positions()
                test.expect(actual[1]):equals(2)
                test.expect(actual[2]):equals(4)
                local event = forced_event()
                test.expect(event:get("outcome")):equals("blocked")
                test.expect(event:get("applied_distance")):equals(0)
            end),
        }),
        test.case("forced_motion_uses_static_collision_authority", {
            test.run(function()
                require("d2legacy.gameplay.systems.movement").set_collision(2, {
                    integrate_velocity = function(_, x, y)
                        return x, y
                    end,
                })
                unit(2, 0, true, {
                    kind = "knockback",
                    source_x = 0,
                    source_y = 10,
                    distance = 3,
                    speed = 25,
                    request_tick = 0,
                })
            end),
            test.step(1),
            test.run(function()
                test.expect(positions()[1]):equals(2)
                test.expect(forced_event():get("outcome")):equals("blocked")
            end),
        }),
        test.case("forced_motion_rejects_an_undefined_direction", {
            test.run(function()
                unit(2, 0, true, {
                    kind = "knockback",
                    source_x = 2,
                    source_y = 10,
                    distance = 3,
                    speed = 25,
                    request_tick = 0,
                })
            end),
            test.step(1),
            test.run(function()
                test.expect(positions()[1]):equals(2)
                test.expect(forced_event():get("outcome")):equals("invalid")
            end),
        }),
        test.case("forced_motion_reports_partial_progress_at_contact", {
            test.run(function()
                unit(2, 0, true, {
                    kind = "knockback",
                    source_x = 0,
                    source_y = 10,
                    distance = 4,
                    speed = 25,
                    request_tick = 0,
                })
                unit(5, 0, true)
            end),
            test.step(2),
            test.run(function()
                local actual = positions()
                test.expect(actual[1]):equals(3)
                test.expect(actual[2]):equals(5)
                local event = forced_event()
                test.expect(event:get("outcome")):equals("partial")
                test.expect(event:get("applied_distance")):equals(1)
            end),
            test.expect_checkpoint_parity(2),
        }),
    },
})
