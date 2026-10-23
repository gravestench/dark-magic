local test = require("d2legacy.tests/v1")

-- Mirrors owned 1.14d Experience.txt: ExpRatio 1024 through level 69, 976 at
-- 70, 5 at 99, a shift of 10, and a level cap of 99.
local function rules()
    local ratios = {}
    for level = 0, 69 do
        ratios[level] = 1024
    end
    ratios[70] = 976
    ratios[99] = 5
    return { ratio_shift = 10, ratios = ratios, max_levels = { Amazon = 99 } }
end

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("level_difference_tables_and_ratio_scale_kill_experience", {
            test.run(function()
                local xp = require("d2legacy.policy.experience_gain")
                local r = rules()
                -- Same level, and up to 5 levels apart, keeps everything.
                test.expect(xp.gain(r, "Amazon", 10, 10, 1000, 0)):equals(1000)
                test.expect(xp.gain(r, "Amazon", 15, 10, 1000, 0)):equals(1000)
                -- Killer 6 and 10+ levels above: 207/256 and 13/256.
                test.expect(xp.gain(r, "Amazon", 16, 10, 1000, 0)):equals(808)
                test.expect(xp.gain(r, "Amazon", 40, 10, 1000, 0)):equals(50)
                -- Monster 6 levels above a killer below 25: 225/256.
                test.expect(xp.gain(r, "Amazon", 10, 16, 1000, 0)):equals(878)
                -- From killer level 25, a higher monster gives killer/monster.
                test.expect(xp.gain(r, "Amazon", 30, 40, 1000, 0)):equals(750)
                -- Level 70 scales by 976/1024; level 99 gains nothing.
                test.expect(xp.gain(r, "Amazon", 70, 70, 1000, 0)):equals(953)
                test.expect(xp.gain(r, "Amazon", 99, 99, 1000, 0)):equals(0)
                -- item_addexperience applies after the ratio.
                test.expect(xp.gain(r, "Amazon", 10, 10, 1000, 10)):equals(1100)
                -- PXL-R05a returns one for nonpositive raw party shares
                -- before checking the recipient's class level cap.
                test.expect(xp.gain(r, "Amazon", 10, 10, 0, 10)):equals(1)
                test.expect(xp.gain(r, "Amazon", 99, 10, -1, 0)):equals(1)
            end),
        }),
        test.case("recovered_overflow_branches_are_kept", {
            test.run(function()
                local xp = require("d2legacy.policy.experience_gain")
                -- Large values divide before multiplying.
                test.expect(xp.percentage(2000000, 30, 40)):equals(1500000)
                test.expect(xp.percentage(1000, 30, 40)):equals(750)
                test.expect(xp.percentage(1000, 30, 0)):equals(0)
                -- Above INT_MAX >> 10 the ratio applies after the shift:
                -- 976 * (3000000 >> 10) rather than 3000000 * 976 >> 10.
                test.expect(xp.gain(rules(), "Amazon", 70, 70, 3000000, 0)):equals(976 * 2929)
            end),
        }),
        test.case("total_experience_stops_at_the_level_cap_threshold", {
            test.run(function()
                local xp = require("d2legacy.policy.experience_gain")
                local thresholds = { Amazon = {} }
                thresholds.Amazon[99] = 5000
                test.expect(xp.capped_total(rules(), thresholds, "Amazon", 4900, 300)):equals(5000)
                test.expect(xp.capped_total(rules(), thresholds, "Amazon", 100, 300)):equals(400)
                test.expect(xp.capped_total({ max_levels = {} }, {}, "Amazon", 100, 300)):equals(400)
            end),
        }),
    },
})
