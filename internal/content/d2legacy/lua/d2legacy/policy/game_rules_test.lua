local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "authority",
    tier = "fast",
    initial_data = {
        ["d2legacy.game_rules"] = {
            target = "lod-1.14d",
            expansion = true,
            difficulty = 1,
            hardcore = true,
            ladder = false,
            player_count = 2,
            maximum_players = 8,
        },
        ["engine.game_data_generation_id"] = "sha256:test-generation",
    },
    cases = {
        test.case("rules_are_validated_copied_and_checkpointed", {
            test.run(function()
                local rules = require("d2legacy.policy.game_rules")
                local value = rules.get()
                test.assert(value.target == "lod-1.14d", [=[value.target == "lod-1.14d"]=])
                test.assert(
                    value.expansion and value.difficulty == 1,
                    [=[value.expansion and value.difficulty == 1]=]
                )
                test.assert(
                    value.hardcore and value.player_count == 2,
                    [=[value.hardcore and value.player_count == 2]=]
                )
                value.difficulty = 2
                test.assert(rules.difficulty() == 1, [=[rules.difficulty() == 1]=])
                local state = require("engine.authority_state/v1").read("d2legacy.game_rules")
                test.assert(state.difficulty == 1, [=[state.difficulty == 1]=])
                test.assert(state.game_data_generation_id == "sha256:test-generation",
                    [=[state.game_data_generation_id == "sha256:test-generation"]=])
            end),
        }),
    },
})
