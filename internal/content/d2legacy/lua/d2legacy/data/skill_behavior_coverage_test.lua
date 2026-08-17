local test = require("d2legacy.tests/v1")

return test.suite({
    name = "skill behavior coverage declarations",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("loads_exact_target_locked_family_ids", function(t)
            t:run(function()
                local coverage = require("d2legacy.data.skill_behavior_coverage").load()
                test.expect(coverage.by_family["missile.straight"][1]):equals(36)
                test.expect(coverage.by_family["state.self-timed"][1]):equals(40)
                test.expect(coverage.by_id[36].evidence_status):equals("owned-target-records-verified")
                test.expect(coverage.by_id[39] == nil):is_true()
            end)
        end),
    },
})
