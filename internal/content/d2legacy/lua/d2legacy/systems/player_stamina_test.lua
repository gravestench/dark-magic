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
        test.case("exhaustion_preserves_a_close_waypoint_velocity_cap", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ x = 10, y = 12, level_id = 2 }),
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                ecs.get(player(), "d2legacy.player.vitals"):set("stamina_raw", 1)
            end),
            test.submit({
                tick = 2,
                sequence = 1,
                player = "alice",
                kind = "player.move",
                payload = { x = 0, y = 0, running = true, target = { x = 10.25, y = 12 } },
            }),
            test.step(4),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local entity = player()
                local position = ecs.get(entity, "d2legacy.world.position")
                local motion = ecs.get(entity, "d2legacy.player.motion")
                test.expect(position:get("x")):equals(10.25)
                test.expect(position:get("y")):equals(12)
                test.expect(motion:get("running")):equals(false)
            end),
            test.expect_checkpoint_parity(1),
        }),
        test.case("wilderness_running_drains_fixed_point_stamina_and_falls_back_to_walk", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({
                    level_id = 2,
                    stamina = 1,
                    max_stamina = 1,
                    passive_stat_sources = {
                        { id = "test-minimum-maximum", stat = "maxstamina", value = -83 },
                    },
                }),
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
                test.assert(math.abs(velocity:get("x") - 9.375) < 0.000000001, "walk fallback rate")
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
                test.assert(math.abs(velocity:get("x") - 9.375) < 0.000000001, "zero stamina moves at walk rate")
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
        test.case("act_ii_town_run_uses_town_recovery", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ level_id = 40, stamina = 10, max_stamina = 84 }),
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                ecs.get(player(), "d2legacy.player.vitals"):set("stamina_raw", 2560)
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
                test.expect(ecs.get(player(), "d2legacy.player.vitals"):get("stamina_raw")):equals(2602)
            end),
        }),
        test.case("former_false_positive_town_level_drains_running_stamina", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ level_id = 45, stamina = 10, max_stamina = 84 }),
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                ecs.get(player(), "d2legacy.player.vitals"):set("stamina_raw", 2560)
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
                test.expect(ecs.get(player(), "d2legacy.player.vitals"):get("stamina_raw")):equals(2520)
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
                test.assert(
                    math.abs(velocity:get("x") - 11.71875) < 0.000000001,
                    "30 item FRW becomes 25 effective FRW"
                )
            end),
        }),
    },
})
