local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "authority",
    tier = "fast",
    cases = {
        test.case("boots_headlessly_without_renderer_or_client_startup", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.assert(
                    type(require("d2legacy.authoritative")) == "table",
                    [=[type(require("d2legacy.authoritative")) == "table"]=]
                )
                test.assert(
                    type(ecs.query({ all = { "d2legacy.player.identity" } })) == "table",
                    [=[type(ecs.query({ all = { "d2legacy.player.identity" } })) == "table"]=]
                )
                local ok = pcall(require, "engine.render/v1")
                test.assert(
                    not ok,
                    "headless authority unexpectedly installed the renderer",
                    [=[not ok, "headless authority unexpectedly installed the renderer"]=]
                )
            end),
        }),
    },
})
