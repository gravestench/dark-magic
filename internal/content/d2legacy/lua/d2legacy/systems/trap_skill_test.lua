local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local death_sentry = {
    Id = "276",
    skill = "Death Sentry",
    skilldesc = "death sentry",
    charclass = "ass",
    InGame = "1",
    range = "rng",
    mana = "20",
    lvlmana = "1",
    minmana = "1",
    manashift = "8",
    HitShift = "8",
    EType = "ltng",
    EMin = "1",
    EMax = "2",
    EMinLev1 = "1",
    EMinLev2 = "1",
    EMinLev3 = "1",
    EMinLev4 = "1",
    EMinLev5 = "1",
    EMaxLev1 = "1",
    EMaxLev2 = "1",
    EMaxLev3 = "1",
    EMaxLev4 = "1",
    EMaxLev5 = "1",
    srvdofunc = "45",
    summon = "death-sentry-fixture",
    pettype = "trap",
    petmax = "5",
    sumskill1 = "Corpse Explosion Fixture",
    sumskill2 = "Lightning Fixture",
    Param1 = "5",
}

local function install_transaction()
    local ecs = require("engine.ecs/v1")
    local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
    ecs.create({
        ["d2legacy.world.position"] = { x = 0, y = 0 },
        ["d2legacy.world.location"] = { act = 1, level_id = 1 },
        ["d2legacy.trap.autonomous"] = {
            owner = player,
            owner_id = "player:alice",
            skill_id = 276,
            skill_level = 1,
            shots_remaining = 1,
            next_fire_tick = 2,
            fire_interval = 10,
            target_radius = 6,
            cast_serial = 0,
        },
    })
    ecs.create({
        ["d2legacy.monster.identity"] = { spawn_id = "corpse", definition_id = "fallen", seed = "corpse" },
        ["d2legacy.monster.stats"] = { level = 1, health = 0, max_health = 10 * 256 },
        ["d2legacy.monster.death"] = { tick = 1, active = false, corpse_usable = true },
        ["d2legacy.world.position"] = { x = 5, y = 0 },
        ["d2legacy.world.location"] = { act = 1, level_id = 1 },
        ["d2legacy.world.selectable"] = { id = "monster:corpse", kind = "corpse", label = "corpse" },
    })
    ecs.create({
        ["d2legacy.monster.identity"] = { spawn_id = "victim", definition_id = "fallen", seed = "victim" },
        ["d2legacy.monster.stats"] = { level = 1, health = 100 * 256, max_health = 100 * 256 },
        ["d2legacy.world.position"] = { x = 8, y = 0 },
        ["d2legacy.world.location"] = { act = 1, level_id = 1 },
        ["d2legacy.world.selectable"] = { id = "monster:victim", kind = "hostile", label = "victim" },
    })
end

local function install_returning_weapon()
    local ecs = require("engine.ecs/v1")
    local spawn = require("d2legacy.gameplay.projectile_spawn")
    local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
    local definition = {
        skill_id = 257,
        lifetime_ticks = 10,
        collision_radius = 0.25,
        destroy_on_contact = false,
        next_hit_delay = 2,
        impact_radius = 0,
        minimum_damage_raw = 256,
        maximum_damage_raw = 256,
        minimum_damage_per_level_raw = { 0, 0, 0, 0, 0 },
        maximum_damage_per_level_raw = { 0, 0, 0, 0, 0 },
        damage_channel = "physical",
        missile_id = "blade-sentinel-fixture",
        dcc = "fixture.dcc",
        palette = "fixture-pal.dat",
        travel_sound = "",
        hit_sound = "",
        directions = 1,
        frames_per_second = 25,
        loop = true,
        transparency_mode = 0,
        offset_x = 0,
        offset_y = 0,
        offset_z = 0,
    }
    local components = spawn.components(player, definition, {
        cast_id = "returning-weapon",
        projectile_id = "returning-weapon",
        target_x = 2,
        target_y = 0,
        velocity_x = 0,
        velocity_y = 0,
        skill_level = 1,
    })
    components["d2legacy.trap.returning_weapon"] = {
        owner = player,
        target_x = 2,
        target_y = 0,
        speed_per_tick = 1,
        outbound = true,
        expires_tick = 20,
    }
    ecs.create(components)
end

return test.suite({
    name = "Assassin trap transactions",
    profile = "authority",
    tier = "integration",
    covers = { "internal/game/trap" },
    records = {
        ["data/global/excel/skills.txt"] = {
            death_sentry,
            {
                Id = "904",
                skill = "Corpse Explosion Fixture",
                TargetCorpse = "1",
                Param1 = "40",
                Param2 = "80",
                Param3 = "12",
                Param4 = "0",
                calc3 = "50",
            },
            {
                Id = "905",
                skill = "Lightning Fixture",
                srvmissile = "death-shot",
                Param1 = "1",
                Param2 = "0",
            },
        },
        ["data/global/excel/Missiles.txt"] = {
            {
                Missile = "death-shot",
                pSrvDoFunc = "1",
                Range = "25",
                Vel = "25",
                Size = "1",
                CollideKill = "1",
                CelFile = "fixture",
                NumDirections = "1",
                AnimSpeed = "16",
            },
        },
        ["data/global/excel/PetType.txt"] = { { ["pet type"] = "trap", group = "1", range = "1" } },
        ["data/global/excel/MonStats.txt"] = {
            {
                Id = "death-sentry-fixture",
                MonStatsEx = "death-sentry-gfx",
                AI = "DeathSentry",
                aip3 = "6",
                aip4 = "10",
            },
        },
        ["data/global/excel/MonStats2.txt"] = { { Id = "death-sentry-gfx" } },
        ["data/global/excel/SkillDesc.txt"] = {
            { skilldesc = "death sentry", ["str name"] = "fixture-death-sentry" },
        },
    },
    cases = {
        test.case("death_sentry_centers_ordered_corpse_damage_on_the_consumed_corpse", {
            test.submit_system(fixtures.command("system.player.enter", fixtures.player_entry(), {
                tick = 1,
                sequence = 1,
            })),
            test.step(1),
            test.run(install_transaction),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(#ecs.query({ all = { "d2legacy.monster.death" } })):equals(0)
                local events = ecs.query({ all = { "d2legacy.combat.event" } })
                test.expect(#events):equals(1)
                local event = ecs.get(events[1], "d2legacy.combat.event")
                test.expect(event:get("target_id")):equals("monster:victim")
                test.expect(event:get("source_kind")):equals("trap_corpse")
                test.expect(event:get("damage_channel")):equals("mixed")
            end),
            test.expect_checkpoint_parity(1),
        }),
        test.case("blade_sentinel_patrols_to_the_target_and_returns_to_its_owner", {
            test.submit_system(fixtures.command("system.player.enter", fixtures.player_entry({ x = 0, y = 0 }), {
                tick = 1,
                sequence = 1,
            })),
            test.step(1),
            test.run(install_returning_weapon),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local entity = ecs.query({ all = { "d2legacy.trap.returning_weapon" } })[1]
                test.expect(ecs.get(entity, "d2legacy.world.position"):get("x")):equals(1)
                test.expect(ecs.get(entity, "d2legacy.trap.returning_weapon"):get("outbound")):equals(true)
            end),
            test.expect_checkpoint_parity(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local entity = ecs.query({ all = { "d2legacy.trap.returning_weapon" } })[1]
                test.expect(ecs.get(entity, "d2legacy.world.position"):get("x")):equals(0)
                test.expect(ecs.get(entity, "d2legacy.trap.returning_weapon"):get("outbound")):equals(false)
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(#ecs.query({ all = { "d2legacy.trap.returning_weapon" } })):equals(0)
            end),
        }),
    },
})
