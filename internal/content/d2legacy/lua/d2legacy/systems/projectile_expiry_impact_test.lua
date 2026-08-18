local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local function install_ground_impact_projectile()
    local ecs = require("engine.ecs/v1")
    local spawn = require("d2legacy.gameplay.projectile_spawn")
    local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
    local definition = {
        skill_id = 251,
        lifetime_ticks = 1,
        collision_radius = 0.25,
        destroy_on_contact = true,
        impact_on_expiry = true,
        next_hit_delay = 0,
        impact_radius = 4,
        impact_missile_id = "fixture-impact",
        impact_dcc = "fixture-impact.dcc",
        impact_palette = "fixture-pal.dat",
        impact_lifetime_ticks = 3,
        impact_directions = 1,
        impact_frames_per_second = 25,
        impact_loop = false,
        impact_transparency_mode = 0,
        impact_sound = "",
        minimum_damage_raw = 256,
        maximum_damage_raw = 256,
        minimum_damage_per_level_raw = { 0, 0, 0, 0, 0 },
        maximum_damage_per_level_raw = { 0, 0, 0, 0, 0 },
        damage_channel = "fire",
        missile_id = "fixture-air",
        dcc = "fixture-air.dcc",
        palette = "fixture-pal.dat",
        travel_sound = "",
        hit_sound = "",
        directions = 1,
        frames_per_second = 25,
        loop = false,
        transparency_mode = 0,
        offset_x = 0,
        offset_y = 0,
        offset_z = 0,
    }
    ecs.create(spawn.components(player, definition, {
        cast_id = "ground-arrival",
        projectile_id = "ground-arrival",
        target_x = 0.8,
        target_y = 0,
        velocity_x = 0.8,
        velocity_y = 0,
        skill_level = 1,
    }))
end

return test.suite({
    name = "projectile expiry impact",
    profile = "authority",
    tier = "integration",
    cases = {
        test.case("arrival_without_unit_contact_applies_area_payload_once", {
            test.submit_system(fixtures.command("system.player.enter", fixtures.player_entry({ x = 0, y = 0 }), {
                tick = 1,
                sequence = 1,
            })),
            test.submit_system(fixtures.command(
                "system.monster.spawn",
                fixtures.monster_spawn({
                    spawn_id = "near-ground-impact",
                    x = 0.8,
                    y = 3,
                }),
                {
                    tick = 1,
                    sequence = 2,
                }
            )),
            test.step(1),
            test.run(install_ground_impact_projectile),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local events = ecs.query({ all = { "d2legacy.combat.event" } })
                test.expect(#events):equals(1)
                local event = ecs.get(events[1], "d2legacy.combat.event")
                test.expect(event:get("target_id")):equals("monster:near-ground-impact")
                test.expect(event:get("damage_channel")):equals("fire")
                test.expect(event:get("damage_raw")):equals(256)
                test.expect(#ecs.query({ all = { "d2legacy.missile.projectile" } })):equals(0)
                test.expect(#ecs.query({ all = { "d2legacy.missile.effect" } })):equals(1)
            end),
            test.expect_checkpoint_parity(1),
            test.step(3),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(#ecs.query({ all = { "d2legacy.missile.effect" } })):equals(0)
                test.expect(#ecs.query({ all = { "d2legacy.combat.event" } })):equals(1)
            end),
        }),
    },
})
