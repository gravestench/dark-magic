local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local function fire_arrow()
    local ecs = require("engine.ecs/v1")
    local spawn = require("d2legacy.gameplay.projectile_spawn")
    local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
    ecs.create({
        ["d2legacy.stat.source"] = {
            target = player,
            source_id = "test:ring-life",
            stat = "lifedrainmindam",
            value = 30,
        },
    })
    local monster = ecs.query({ all = { "d2legacy.monster.stats" } })[1]
    ecs.get(monster, "d2legacy.monster.stats"):set("drain_percent", 100)
    ecs.create(spawn.components(player, {
        skill_id = 0,
        lifetime_ticks = 5,
        collision_radius = 0.25,
        destroy_on_contact = true,
        impact_on_expiry = false,
        next_hit_delay = 0,
        impact_radius = 0,
        impact_missile_id = "",
        impact_dcc = "",
        impact_palette = "",
        impact_lifetime_ticks = 0,
        impact_directions = 1,
        impact_frames_per_second = 1,
        impact_loop = false,
        impact_transparency_mode = 0,
        impact_sound = "",
        minimum_damage_raw = 100 * 256,
        maximum_damage_raw = 100 * 256,
        minimum_damage_per_level_raw = { 0, 0, 0, 0, 0 },
        maximum_damage_per_level_raw = { 0, 0, 0, 0, 0 },
        damage_channel = "physical",
        missile_id = "fixture-arrow",
        dcc = "fixture-arrow.dcc",
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
    }, {
        cast_id = "arrow",
        projectile_id = "arrow",
        target_x = 4,
        target_y = 0,
        velocity_x = 0.8,
        velocity_y = 0,
        skill_level = 1,
    }))
end

return test.suite({
    name = "projectile leech",
    profile = "authority",
    tier = "integration",
    cases = {
        -- Normal has no steal divisor. The 16-life fixture monster caps the
        -- 100-damage arrow, so 30% of 16 returns 4 whole life.
        test.case("player_missile_hits_leech_from_the_owner_stats", {
            test.submit_system(
                fixtures.command(
                    "system.player.enter",
                    fixtures.player_entry({ x = 0, y = 0, health = 10, max_health = 50 }),
                    { tick = 1, sequence = 1 }
                )
            ),
            test.submit_system(
                fixtures.command(
                    "system.monster.spawn",
                    fixtures.monster_spawn({ spawn_id = "target", x = 0.8, y = 0 }),
                    { tick = 1, sequence = 2 }
                )
            ),
            test.step(1),
            test.run(fire_arrow),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local events = ecs.query({ all = { "d2legacy.combat.event" } })
                test.expect(#events):equals(1)
                test.expect(ecs.get(events[1], "d2legacy.combat.event"):get("damage_channel")):equals("physical")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                test.expect(ecs.get(player, "d2legacy.player.vitals"):get("health")):equals(14)
                local overlays = 0
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.presentation.effect_cue" } })) do
                    local cue = ecs.get(entity, "d2legacy.presentation.effect_cue")
                    if cue:get("kind") == "leech" then
                        overlays = overlays + 1
                        test.expect(cue:get("overlay_id")):equals("steallife")
                    end
                end
                test.expect(overlays):equals(1)
            end),
            test.expect_checkpoint_parity(1),
        }),
    },
})
