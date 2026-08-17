local test = require("d2legacy.tests/v1")

return test.suite({
    name = "player-count authority commands",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("validated_commands_set_and_clear_the_gameplay_override", function(t)
            t:run(function()
                local registered = {}
                local calls = {}
                test.mock_module("engine.authority_command/v1", {
                    register = function(definition)
                        registered[definition.kind] = definition
                    end,
                }, { "register" })
                test.mock_module("d2legacy.policy.player_count", {
                    set_override = function(value)
                        calls.set = value
                    end,
                    clear_override = function()
                        calls.cleared = true
                    end,
                }, { "set_override", "clear_override" })

                local commands = require("d2legacy.commands.player_count")
                commands.register()
                local override = registered["game.player_count.override"]
                local follow = registered["game.player_count.follow_population"]
                test.assert(override ~= nil and follow ~= nil, "player-count commands were not registered")
                test.expect(override.authorities):deep_equals({ "system", "administrator" })

                override.validate({ payload = { count = 8 } })
                override.apply({ payload = { count = 8 } })
                test.expect(calls.set):equals(8)
                local valid = pcall(override.validate, { payload = { count = 9 } })
                test.expect(valid):is_false()

                follow.validate({ payload = {} })
                follow.apply({ payload = {} })
                test.expect(calls.cleared):is_true()
            end)
        end),
    },
})
