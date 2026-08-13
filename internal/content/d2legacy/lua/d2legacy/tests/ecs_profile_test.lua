local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "ecs",
    tier = "fast",
    cases = {
        test.case("loads_production_components_without_authority_commands", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local entity = ecs.create({ ["d2legacy.world.position"] = { x = 3, y = 4 } })
                test.expect(ecs.get(entity, "d2legacy.world.position"):get("x"), "position x"):equals(3)
                local has_commands = pcall(require, "engine.authority_command/v1")
                test.assert(
                    not has_commands,
                    "ECS profile unexpectedly exposed authority commands",
                    [=[not has_commands, "ECS profile unexpectedly exposed authority commands"]=]
                )
            end),
        }),
    },
})
