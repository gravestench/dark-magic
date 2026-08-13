local test = require("d2legacy.tests/v1")
local stats = require("d2legacy.policy.stat_resolution")

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("flat_sources_precede_the_combined_percentage_phase", function(t)
            t:run(function()
                local sources = {
                    { id = "late-flat", order = 30, operation = "add", value = 25 },
                    { id = "first-percent", order = 10, operation = "percent", value = 20 },
                    { id = "early-flat", order = 5, operation = "add", value = 75 },
                    { id = "second-percent", order = 20, operation = "percent", value = 10 },
                }
                test.expect(stats.resolve(100, sources), "resolved stat"):equals(260)
            end)
        end),
        test.case("integer_truncation_occurs_after_percentage_application", function(t)
            t:run(function()
                local sources = {
                    { id = "percent", order = 1, operation = "percent", value = 33 },
                }
                test.expect(stats.resolve(10, sources), "truncated stat"):equals(13)
                test.expect(stats.local_value(41, 50), "local armor defense"):equals(61)
            end)
        end),
    },
})
