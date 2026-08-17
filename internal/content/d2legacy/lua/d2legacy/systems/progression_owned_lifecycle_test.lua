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

local function create_wolf(inactive)
    local ecs = require("engine.ecs/v1")
    local components = {
        ["d2legacy.world.selectable"] = {
            id = "monster:wolf",
            kind = "friendly",
            label = "Wolf",
            owner = "alice",
            radius = 1,
            priority = 1,
        },
    }
    if inactive then
        components["d2legacy.world.inactive"] = {}
    end
    ecs.create(components)
end

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
                create_wolf(false)
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
        test.case("inactive_owned_unit_is_filtered_until_reactivation", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.amazon_level_up_entry,
            }),
            test.step(1),
            test.run(function()
                create_wolf(true)
            end),
            test.submit_system({
                tick = 2,
                sequence = 2,
                kind = "system.owned_unit.attach",
                payload = attach,
            }),
            test.step(1),
            test.expect_checkpoint_parity(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local units = ecs.query({
                    all = { "d2legacy.owned_unit", "d2legacy.world.inactive" },
                })
                test.assert(#units == 1, [=[#units == 1]=])
                ecs.remove(units[1], "d2legacy.world.inactive")
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.assert(
                    #ecs.query({ all = { "d2legacy.owned_unit" } }) == 0,
                    [=[#ecs.query({ all = { "d2legacy.owned_unit" } }) == 0]=]
                )
            end),
        }),
    },
})
