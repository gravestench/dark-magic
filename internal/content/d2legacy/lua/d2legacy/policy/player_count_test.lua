local test = require("d2legacy.tests/v1")

return test.suite({
    name = "purpose-specific player counts",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("population_is_default_and_override_is_independent_from_capacity", function(t)
            t:run(function()
                local stored
                test.mock_module("d2legacy.policy.game_rules", {
                    get = function()
                        return { maximum_players = 2 }
                    end,
                }, { "get" })
                test.mock_module("engine.authority_state/v1", {
                    register = function(id, schema, value)
                        test.expect(id):equals("d2legacy.player_count")
                        test.expect(schema):equals("d2legacy.player_count/v1")
                        stored = value
                    end,
                    read = function()
                        return stored
                    end,
                    replace = function(id, schema, value)
                        test.expect(id):equals("d2legacy.player_count")
                        test.expect(schema):equals("d2legacy.player_count/v1")
                        stored = value
                    end,
                }, { "register", "read", "replace" })

                local policy = require("d2legacy.policy.player_count")
                policy.initialize()
                local context = policy.no_drop(2, 1, 2)
                test.expect(context.game_player_count):equals(2)
                test.expect(context.effective_player_count):equals(2)
                test.expect(context.nearby_party_member_count):equals(1)
                test.expect(context.monster_player_count):equals(2)
                test.expect(context.no_drop_player_count):equals(2)

                policy.set_override(8)
                test.expect(policy.snapshot().override):equals(8)
                local scaled = policy.monster_spawn(2, true)
                test.expect(scaled.effective_player_count):equals(8)
                test.expect(scaled.life_bonus_percent):equals(350)
                test.expect(scaled.experience_bonus_percent):equals(350)

                local friendly = policy.monster_spawn(2, false)
                test.expect(friendly.effective_player_count):equals(1)
                test.expect(friendly.life_bonus_percent):equals(0)

                policy.set_override(1)
                local forced_low = policy.no_drop(2, 1, 8)
                test.expect(forced_low.game_player_count):equals(2)
                test.expect(forced_low.effective_player_count):equals(1)
                test.expect(forced_low.nearby_party_member_count):equals(1)
                test.expect(forced_low.no_drop_player_count):equals(1)

                policy.clear_override()
                test.expect(policy.snapshot().override):is_nil()
                test.expect(policy.effective(2)):equals(2)
            end)
        end),
    },
})
