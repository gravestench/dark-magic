local test = require("d2legacy.tests/v1")
local action_rate = require("d2legacy.policy.action_rate")

local timing = {
    name = "AMA1HTH",
    frames = 13,
    speed = 256,
    frame_scale = 256,
    events = { { kind = "attack", frame = 8 } },
}

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("uses_integer_eias_and_fixed_point_animdata_boundaries", function(t)
            t:run(function()
                test.expect(action_rate.effective_item_rate(20)):equals(17)
                test.expect(action_rate.effective_item_rate(40)):equals(30)
                local schedule = action_rate.schedule(timing, {
                    attack_rate = 100,
                    item_fasterattackrate = 40,
                })
                test.expect(schedule.rate_percent):equals(130)
                test.expect(schedule.animation_speed):equals(332)
                test.expect(schedule.attack_delay):equals(7)
                test.expect(schedule.complete_delay):equals(11)
            end)
        end),
        test.case("applies_weapon_speed_dual_average_sequence_and_rate_caps", function(t)
            t:run(function()
                test.expect(action_rate.rate_percent({ attack_rate = 80 })):equals(80)
                test.expect(action_rate.rate_percent({
                    attack_rate = 130,
                    primary_weapon_attack_rate = 30,
                    secondary_weapon_attack_rate = -20,
                    dual_wield = true,
                })):equals(105)
                test.expect(action_rate.rate_percent({ attack_rate = 20, sequence = true })):equals(15)
                test.expect(action_rate.rate_percent({ attack_rate = 300 })):equals(175)
            end)
        end),
    },
})
