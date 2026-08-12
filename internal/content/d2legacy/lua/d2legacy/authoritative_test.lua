local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "authority",
    tier = "fast",
    tests = {
        boots_headlessly_without_renderer_or_client_startup = {
            {
                run = function()
                    local ecs = require("engine.ecs/v1")
                    assert(type(require("d2legacy.authoritative")) == "table")
                    assert(type(ecs.query({ all = { "d2legacy.player.identity" } })) == "table")
                    local ok = pcall(require, "engine.render/v1")
                    assert(not ok, "headless authority unexpectedly installed the renderer")
                end,
            },
        },
    },
})
