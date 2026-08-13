local test = require("d2legacy.tests/v1")

local function assert_connected(plan)
    for index = 2, #plan.ordered do
        local previous = plan.ordered[index - 1]
        local current = plan.ordered[index]
        test.assert(math.abs(previous.x - current.x) <= 1, [=[math.abs(previous.x - current.x) <= 1]=])
        test.assert(math.abs(previous.y - current.y) <= 1, [=[math.abs(previous.y - current.y) <= 1]=])
    end
end

local function blood_moor_edge_and_route_policy()
    local route = require("d2legacy.mapgen.outdoor_route")
    for _, direction in ipairs({ "north", "east", "south", "west" }) do
        local plan = route.plan(42, 80, 80, direction)
        test.assert(#plan.ordered == 10, [=[#plan.ordered == 10]=])
        test.assert(#plan.path > 10, [=[#plan.path > 10]=])
        test.assert(
            plan.entry.direction == route.opposite(direction),
            [=[plan.entry.direction == route.opposite(direction)]=]
        )
        test.assert(plan.exit.direction == direction, [=[plan.exit.direction == direction]=])
        test.assert(plan.path[1] and plan.cells, [=[plan.path[1] and plan.cells]=])
        assert_connected(plan)
    end

    local east = route.plan(42, 80, 80, "east")
    test.assert(east.entry.x == 0 and east.entry.y == 40, [=[east.entry.x == 0 and east.entry.y == 40]=])
    test.assert(east.exit.x == 79 and east.exit.y == 40, [=[east.exit.x == 79 and east.exit.y == 40]=])
end

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("blood_moor_edge_and_route_policy", {
            { run = blood_moor_edge_and_route_policy },
        }),
    },
})
