local test = require("d2legacy.tests/v1")

return test.suite({
    name = "purpose-specific player counts",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("nodrop_keeps_actual_effective_and_nearby_counts_distinct", function(t)
            t:run(function()
                test.mock_module("d2legacy.policy.game_rules", {
                    get = function()
                        return { maximum_players = 8 }
                    end,
                    effective_player_count = function(actual)
                        return math.max(actual, 3)
                    end,
                }, { "get", "effective_player_count" })

                local policy = require("d2legacy.policy.player_count")
                local context = policy.no_drop(2, 1, 2)
                test.expect(context.game_player_count):equals(2)
                test.expect(context.effective_player_count):equals(3)
                test.expect(context.nearby_party_member_count):equals(1)
                test.expect(context.monster_player_count):equals(2)
                test.expect(context.no_drop_player_count):equals(2)

                local scaled = policy.monster_spawn(2, true)
                test.expect(scaled.effective_player_count):equals(3)
                test.expect(scaled.life_bonus_percent):equals(100)
                test.expect(scaled.experience_bonus_percent):equals(100)

                local friendly = policy.monster_spawn(2, false)
                test.expect(friendly.effective_player_count):equals(1)
                test.expect(friendly.life_bonus_percent):equals(0)
            end)
        end),
    },
})
