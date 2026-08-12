local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "presentation",
    tier = "fast",
    tests = {
        defaults_instant_and_supports_parameterized_easing = {
            {
                run = function()
                    local ecs = require("engine.ecs/v1")
                    ecs.component({
                        name = "d2legacy.world.camera_follow",
                        fields = {
                            { name = "target", type = "entity" },
                            { name = "strategy", type = "string" },
                            { name = "duration", type = "f64" },
                            { name = "param_1", type = "f64" },
                            { name = "param_2", type = "f64" },
                            { name = "param_3", type = "f64" },
                            { name = "origin_x", type = "f64" },
                            { name = "origin_y", type = "f64" },
                            { name = "destination_x", type = "f64" },
                            { name = "destination_y", type = "f64" },
                            { name = "elapsed", type = "f64" },
                        },
                    })
                    require("d2legacy.gameplay.systems.camera_follow").register()
                    target = ecs.create({ ["d2legacy.world.position"] = { x = 0, y = 0 } })
                    instant = ecs.create({
                        ["d2legacy.world.position"] = { x = 0, y = 0 },
                        ["d2legacy.world.camera_follow"] = {
                            target = target,
                            strategy = "instant",
                            duration = 0,
                            param_1 = 0,
                            param_2 = 0,
                            param_3 = 0,
                            origin_x = 0,
                            origin_y = 0,
                            destination_x = 0,
                            destination_y = 0,
                            elapsed = 0,
                        },
                    })
                    eased = ecs.create({
                        ["d2legacy.world.position"] = { x = 0, y = 0 },
                        ["d2legacy.world.camera_follow"] = {
                            target = target,
                            strategy = "cubic_out",
                            duration = 1,
                            param_1 = 0,
                            param_2 = 0,
                            param_3 = 0,
                            origin_x = 0,
                            origin_y = 0,
                            destination_x = 0,
                            destination_y = 0,
                            elapsed = 0,
                        },
                    })
                    ecs.get(target, "d2legacy.world.position"):set("x", 10)
                end,
            },
            { engine_update_ms = 500 },
            {
                run = function()
                    local ecs = require("engine.ecs/v1")
                    assert(ecs.get(instant, "d2legacy.world.position"):get("x") == 10)
                    assert(ecs.get(eased, "d2legacy.world.position"):get("x") == 8.75)
                end,
            },
            { engine_update_ms = 500 },
            {
                run = function()
                    local ecs = require("engine.ecs/v1")
                    assert(ecs.get(eased, "d2legacy.world.position"):get("x") == 10)
                end,
            },
        },
    },
})
