local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("loads_policy_without_authority_or_ecs", {
            test.run(function()
                test.assert(
                    type(require("d2legacy.policy.mitigation").apply) == "function",
                    "policy module did not load"
                )
                local has_ecs = pcall(require, "engine.ecs/v1")
                test.assert(
                    not has_ecs,
                    "module profile unexpectedly exposed ECS",
                    [=[not has_ecs, "module profile unexpectedly exposed ECS"]=]
                )
                local has_commands = pcall(require, "engine.authority_command/v1")
                test.assert(
                    not has_commands,
                    "module profile unexpectedly exposed authority commands",
                    [=[not has_commands, "module profile unexpectedly exposed authority commands"]=]
                )
            end),
        }),
    },
})
