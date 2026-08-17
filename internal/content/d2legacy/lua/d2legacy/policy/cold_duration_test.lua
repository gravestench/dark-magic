local test = require("d2legacy.tests/v1")

return test.suite({
    name = "cold duration policy",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("scales_monster_frames_by_expansion_difficulty", function(t)
            t:run(function()
                local policy = require("d2legacy.policy.cold_duration")
                test.expect(policy.monster_frames(30, 0)):equals(30)
                test.expect(policy.monster_frames(30, 1)):equals(15)
                test.expect(policy.monster_frames(30, 2)):equals(7)
            end)
        end),
    },
})
