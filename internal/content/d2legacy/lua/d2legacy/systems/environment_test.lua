local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local function environment()
    local ecs = require("engine.ecs/v1")
    return ecs.get(test.entities_with("d2legacy.world.environment")[1], "d2legacy.world.environment")
end

return test.suite({
    profile = "authority",
    tier = "fast",
    covers = { "internal/game/player" },
    cases = {
        test.case("normal_act_cycle_starts_at_noon_and_is_checkpointed", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ act = 1, level_id = 1 }),
            }),
            test.step(1),
            test.run(function()
                local value = environment()
                test.expect(value:get("act")):equals(1)
                test.expect(value:get("cycle_index")):equals(2)
                test.expect(value:get("period_of_day")):equals(0)
                test.expect(value:get("ticks")):equals(0)
                test.expect(value:get("time_rate")):equals(128)
                value:set("ticks", 160 * 128)
            end),
            test.step(1),
            test.run(function()
                local value = environment()
                test.expect(value:get("cycle_index")):equals(3)
                test.expect(value:get("period_of_day")):equals(1)
                test.expect(value:get("ticks")):equals(160 * 128)
            end),
            test.expect_checkpoint_parity(1),
        }),
        test.case("act_three_night_advances_eleven_environment_ticks", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ act = 3, level_id = 75 }),
            }),
            test.step(1),
            test.run(function()
                local value = environment()
                value:set("cycle_index", 5)
                value:set("period_of_day", 2)
                value:set("ticks", 200 * 128)
            end),
            test.step(1),
            test.run(function()
                test.expect(environment():get("ticks")):equals(200 * 128 + 11)
            end),
        }),
        test.case("act_four_uses_the_recovered_sixteen_tick_step", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ act = 4, level_id = 103 }),
            }),
            test.step(2),
            test.run(function()
                local value = environment()
                test.expect(value:get("act")):equals(4)
                test.expect(value:get("ticks")):equals(16)
            end),
        }),
    },
})
