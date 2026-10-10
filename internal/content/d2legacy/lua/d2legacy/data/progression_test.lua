local test = require("d2legacy.tests/v1")

local function row(level, amazon)
    return {
        Level = level,
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
            row("MaxLvl", "99"),
            row("0", "0"),
            row("1", "500"),
            row("2", "1500"),
        },
    },
    cases = {
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
