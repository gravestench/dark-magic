local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

return test.suite({
    name = "record velocity through projectile movement",
    profile = "authority",
    tier = "integration",
    records = {
        ["data/global/excel/Missiles.txt"] = {
            {
                Missile = "firebolt",
                Skill = "Fire Bolt",
                pSrvDoFunc = "1",
                CollideType = "3",
                CollideKill = "1",
                Vel = "30",
                Range = "10",
                Size = "2",
                CelFile = "firebolt",
            },
        },
    },
    cases = {
        test.case("axial_motion_matches_owned_sentry_steps_and_record_lifetime", {
            test.submit_system(fixtures.command("system.player.enter", fixtures.player_entry({ x = 0, y = 0 }), {
                tick = 1,
                sequence = 1,
            })),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local source = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local definition = require("d2legacy.data.missile_skills").load({ 36 })[36]
                local spawn = require("d2legacy.gameplay.projectile_spawn")
                ecs.create(spawn.components(source, definition, {
                    cast_id = "native-speed",
                    projectile_id = "native-speed",
                    velocity_x = definition.speed_per_tick,
                    velocity_y = 0,
                    skill_level = 1,
                }))
            end),
            test.step(9),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local projectiles = ecs.query({ all = { "d2legacy.missile.projectile" } })
                test.expect(#projectiles):equals(1)
                local position = ecs.get(projectiles[1], "d2legacy.world.position")
                -- Native Vel 30 axial increments are 92160 in 16.16 coordinates.
                test.expect(position:get("x")):equals(9 * 92160 / 65536)
                test.expect(position:get("y")):equals(0)
                local projectile = ecs.get(projectiles[1], "d2legacy.missile.projectile")
                test.expect(projectile:get("remaining_ticks")):equals(1)
                test.expect(projectile:get("owner_id")):equals("player:alice")
            end),
            test.expect_checkpoint_parity(1),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(#ecs.query({ all = { "d2legacy.missile.projectile" } })):equals(0)
            end),
        }),
    },
})
