local test = require("d2legacy.tests/v1")

return test.suite({
    name = "skill behavior coverage declarations",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("loads_exact_target_locked_family_ids", function(t)
            t:run(function()
                local coverage = require("d2legacy.data.skill_behavior_coverage").load()
                test.expect(coverage.by_family["action.melee"][1]):equals(0)
                test.expect(coverage.by_family["missile.straight"][1]):equals(36)
                test.expect(coverage.by_family["missile.straight-impact-area"][1]):equals(47)
                test.expect(coverage.by_family["missile.straight-impact-area-freeze"][1]):equals(55)
                test.expect(coverage.by_family["missile.straight-freeze"][1]):equals(45)
                test.expect(coverage.by_family["movement.point-relocate"][1]):equals(54)
                test.expect(coverage.by_family["state.self-timed"][1]):equals(40)
                test.expect(coverage.by_family["state.targeted-timed"][1]):equals(52)
                test.expect(coverage.by_family["state.point-area-curse"][1]):equals(66)
                test.expect(coverage.by_family["state.point-area-curse"][2]):equals(72)
                test.expect(coverage.by_family["aura.selected-party-stat"][1]):equals(98)
                test.expect(coverage.by_family["aura.selected-party-stat"][2]):equals(100)
                test.expect(coverage.by_family["aura.selected-party-stat"][3]):equals(104)
                test.expect(coverage.by_family["aura.selected-party-stat"][4]):equals(108)
                test.expect(coverage.by_id[36].evidence_status)
                    :equals("owned-target-records-and-localized-synergy-partial")
                test.expect(coverage.by_id[39] == nil):is_true()
            end)
        end),
    },
})
