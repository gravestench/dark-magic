local test = require("d2legacy.tests/v1")

local function row(level, amazon, ratio)
    return {
        Level = level,
        ExpRatio = ratio,
        Amazon = amazon,
        Sorceress = amazon,
        Necromancer = amazon,
        Paladin = amazon,
        Barbarian = amazon,
        Druid = amazon,
        Assassin = amazon,
    }
end

return test.suite({
    profile = "module",
    tier = "fast",
    -- Mirrors the owned 1.14d layout: a MaxLvl row precedes the cumulative
    -- per-level thresholds.
    records = {
        ["data/global/excel/experience.txt"] = {
            row("MaxLvl", "99", "10"),
            row("0", "0", "1024"),
            row("1", "500", "1024"),
            row("2", "1500", "976"),
        },
    },
    cases = {
        test.case("max_level_row_supplies_the_cap_and_ratio_shift", {
            test.run(function()
                local progression = require("d2legacy.data.progression")
                local rules = progression.experience_rules()
                test.expect(rules.ratio_shift):equals(10)
                test.expect(rules.max_levels.Amazon):equals(99)
                test.expect(rules.max_levels.Assassin):equals(99)
                test.expect(rules.ratios[0]):equals(1024)
                test.expect(rules.ratios[2]):equals(976)
            end),
        }),
        test.case("level_cap_row_is_not_a_threshold", {
            test.run(function()
                local progression = require("d2legacy.data.progression")
                local thresholds = progression.load()
                test.expect(#thresholds.Amazon):equals(3)
                test.expect(thresholds.Amazon[1]):equals(0)
                test.expect(progression.level_for(thresholds, "Amazon", 0, 1)):equals(1)
                test.expect(progression.level_for(thresholds, "Amazon", 499, 1)):equals(1)
                test.expect(progression.level_for(thresholds, "Amazon", 500, 1)):equals(2)
                test.expect(progression.level_for(thresholds, "Amazon", 1500, 1)):equals(3)
            end),
        }),
    },
})
