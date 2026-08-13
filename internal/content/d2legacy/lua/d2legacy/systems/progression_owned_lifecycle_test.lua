local test = require("d2legacy.tests/v1")

local fixtures = require("d2legacy.tests.support.fixtures")
local attach = {
    unit_id = "monster:wolf",
    owner_id = "player:alice",
    ultimate_owner_id = "player:alice",
    expires_tick = 3,
    category = {
        id = "wolf",
        group = 1,
        base_max = 1,
        replacement = "replace_oldest",
    },
}
return test.suite({
    profile = "authority",
    tier = "fast",
    cases = {
        test.case("level_and_owned_unit_expiration_restore_identically", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.amazon_level_up_entry,
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                ecs.create({
                    ["d2legacy.world.selectable"] = {
                        id = "monster:wolf",
                        kind = "friendly",
                        label = "Wolf",
                        owner = "alice",
                        radius = 1,
                        priority = 1,
                    },
                })
            end),
            test.submit_system({
                tick = 2,
                sequence = 2,
                kind = "system.owned_unit.attach",
                payload = attach,
            }),
            test.step(1),
            test.expect_checkpoint_parity(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.progress" } })[1]
                test.assert(
                    ecs.get(player, "d2legacy.player.progress"):get("level") == 2,
                    [=[ecs.get(player, "d2legacy.player.progress"):get("level") == 2]=]
                )
                test.assert(
                    #ecs.query({ all = { "d2legacy.owned_unit" } }) == 0,
                    [=[#ecs.query({ all = { "d2legacy.owned_unit" } }) == 0]=]
                )
            end),
        }),
    },
})
