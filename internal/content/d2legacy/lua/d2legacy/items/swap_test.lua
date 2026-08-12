local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "authority",
    tier = "fast",
    initial_data = {
        ["d2legacy.items"] = {
            owner = "alice",
            inventory_width = 4,
            inventory_height = 4,
            belt_capacity = 4,
            items = {
                { id = "held", code = "ssd", width = 2, height = 1, container = "held" },
                {
                    id = "placed",
                    code = "cap",
                    width = 2,
                    height = 1,
                    container = "inventory",
                    x = 1,
                    y = 1,
                },
            },
        },
        ["d2legacy.interactions"] = { owner = "alice" },
    },
    tests = {
        held_item_swaps_with_occupied_destination = {
            {
                submit = {
                    tick = 1,
                    sequence = 1,
                    player = "alice",
                    kind = "item.move",
                    payload = {
                        item_id = "held",
                        place_held = true,
                        destination = { container = "inventory", x = 1, y = 1 },
                    },
                },
            },
            { step = 1 },
            {
                run = function()
                    local ecs = require("engine.ecs/v1")
                    local found = {}
                    for _, entity in
                        ipairs(ecs.query({
                            all = { "d2legacy.item.identity", "d2legacy.item.placement" },
                        }))
                    do
                        found[ecs.get(entity, "d2legacy.item.identity"):get("id")] =
                            ecs.get(entity, "d2legacy.item.placement"):get("container")
                    end
                    assert(found.held == "inventory" and found.placed == "held")
                end,
            },
        },
    },
})
