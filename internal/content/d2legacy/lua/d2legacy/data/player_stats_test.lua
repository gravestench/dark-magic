local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    records = {
        ["data/global/excel/charstats.txt"] = {
            { class = "Amazon", ToHitFactor = "5" },
        },
    },
    cases = {
        test.case("uses_class_factor_and_legacy_dexterity_formulas", {
            test.run(function()
                local stats = require("d2legacy.data.player_stats")
                test.expect(stats.base_attack_rating("Amazon", 20, 3)):equals(73)
                test.expect(stats.base_defense(23, 40)):equals(45)
            end),
        }),
    },
})
