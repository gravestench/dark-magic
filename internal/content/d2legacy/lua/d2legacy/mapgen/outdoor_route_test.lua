local test = require("d2legacy.tests/v1")

local function assert_connected(plan)
    for index = 2, #plan.ordered do
        local previous = plan.ordered[index - 1]
        local current = plan.ordered[index]
        assert(math.abs(previous.x - current.x) <= 1)
        assert(math.abs(previous.y - current.y) <= 1)
    end
end

local function blood_moor_edge_and_route_policy()
    local route = require("d2legacy.mapgen.outdoor_route")
    for _, direction in ipairs({ "north", "east", "south", "west" }) do
        local plan = route.plan(42, 80, 80, direction)
        assert(#plan.ordered == 10)
        assert(#plan.path > 10)
        assert(plan.entry.direction == route.opposite(direction))
        assert(plan.exit.direction == direction)
        assert(plan.path[1] and plan.cells)
        assert_connected(plan)
    end

    local east = route.plan(42, 80, 80, "east")
    assert(east.entry.x == 0 and east.entry.y == 40)
    assert(east.exit.x == 79 and east.exit.y == 40)
end

return test.suite({
    profile = "policy",
    tier = "fast",
    tests = {
        blood_moor_edge_and_route_policy = {
            { run = blood_moor_edge_and_route_policy },
        },
    },
})
