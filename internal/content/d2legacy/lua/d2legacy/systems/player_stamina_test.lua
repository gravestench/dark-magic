local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local function player()
    local ecs = require("engine.ecs/v1")
    return ecs.query({ all = { "d2legacy.player.identity" } })[1]
end

return test.suite({
    profile = "authority",
    tier = "fast",
    covers = { "internal/game/player" },
    cases = {
        test.case("wilderness_running_drains_fixed_point_stamina_and_falls_back_to_walk", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ level_id = 2, stamina = 1, max_stamina = 1 }),
            }),
            test.step(1),
            test.submit({
                tick = 2,
                sequence = 1,
                player = "alice",
                kind = "player.move",
                payload = { x = 1, y = 0, running = true },
            }),
            test.step(7),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local entity = player()
                local vitals = ecs.get(entity, "d2legacy.player.vitals")
                local mode = ecs.get(entity, "d2legacy.player.movement_mode")
                local animation = ecs.get(entity, "d2legacy.player.animation")
                local velocity = ecs.get(entity, "d2legacy.world.velocity")
                test.expect(vitals:get("stamina_raw")):equals(0)
                test.expect(vitals:get("stamina")):equals(0)
                test.expect(mode:get("running")):equals(false)
                test.expect(animation:get("mode")):equals("WL")
                test.assert(math.abs(velocity:get("x") - 6) < 0.000000001, [=[walk fallback velocity is six]=])
            end),
        }),
        test.case("zero_stamina_rejects_run_mode_without_rejecting_movement", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ level_id = 2, stamina = 0, max_stamina = 0 }),
            }),
            test.step(1),
            test.submit({
                tick = 2,
                sequence = 1,
                player = "alice",
                kind = "player.move",
                payload = { x = 1, y = 0, running = true },
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local entity = player()
                local mode = ecs.get(entity, "d2legacy.player.movement_mode")
                local velocity = ecs.get(entity, "d2legacy.world.velocity")
                test.expect(mode:get("running")):equals(false)
                test.assert(math.abs(velocity:get("x") - 6) < 0.000000001, [=[zero stamina moves at walk rate]=])
            end),
        }),
        test.case("idle_and_town_motion_recover_at_full_and_half_rates", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ level_id = 1, stamina = 10, max_stamina = 84 }),
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local vitals = ecs.get(player(), "d2legacy.player.vitals")
                test.expect(vitals:get("stamina_raw")):equals(10 * 256 + 84)
            end),
            test.submit({
                tick = 2,
                sequence = 1,
                player = "alice",
                kind = "player.move",
                payload = { x = 1, y = 0, running = true },
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local vitals = ecs.get(player(), "d2legacy.player.vitals")
                test.expect(vitals:get("stamina_raw")):equals(10 * 256 + 84 + 42)
            end),
        }),
        test.case("item_frw_uses_the_generic_stat_source_and_diminishing_return_rule", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({
                    level_id = 2,
                    passive_stat_sources = {
                        { id = "frw", stat = "item_fastermovevelocity", value = 30 },
                    },
                }),
            }),
            test.step(1),
            test.submit({
                tick = 2,
                sequence = 1,
                player = "alice",
                kind = "player.move",
                payload = { x = 1, y = 0, running = false },
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local entity = player()
                local stats = ecs.get(entity, "d2legacy.player.movement_stats")
                local velocity = ecs.get(entity, "d2legacy.world.velocity")
                test.expect(stats:get("item_fastermovevelocity")):equals(30)
                test.assert(math.abs(velocity:get("x") - 7.5) < 0.000000001, [=[30 item FRW becomes 25 effective FRW]=])
            end),
        }),
    },
})
