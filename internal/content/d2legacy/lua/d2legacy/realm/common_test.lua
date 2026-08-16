local test = require("d2legacy.tests/v1")

test.mock_module("engine.render/v1", {}, {})
test.mock_module("engine.data/v1", {
    load_manifest = function()
        return {}
    end,
}, { "load_manifest" })
test.mock_module("d2legacy.ui.dc6", {}, {})
test.mock_module("d2legacy.ui.text", {}, {})

local common = require("d2legacy.realm.common")

return test.suite({
    name = "Realm common presentation",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("explains_duplicate_character_presence", function(t)
            t:run(function()
                test.expect(common.error({ phase = "error", error = "realm request failed: character_online" }))
                    :equals("THAT CHARACTER IS ALREADY ONLINE")
            end)
        end),
        test.case("explains_game_content_mismatch", function(t)
            t:run(function()
                test.expect(common.error({
                    phase = "error",
                    error = "network recipe differs from the locally composed deterministic runtime",
                })):equals("GAME CONTENT DOES NOT MATCH THE SERVER")
            end)
        end),
    },
})
