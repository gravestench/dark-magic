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

local sentry_monstats = {
    {
        Id = "death-sentry-fixture",
        BaseId = "death-sentry-fixture",
        MonStatsEx = "death-sentry-gfx",
        AI = "DeathSentry",
        Align = "1",
        Code = "DS",
        enabled = "1",
        boss = "1",
        primeevil = "1",
        noRatio = "1",
        minHP = "20",
        maxHP = "20",
        AC = "17",
        ResFi = "23",
        A1TH = "0",
        A1MinD = "0",
        A1MaxD = "0",
        aip3 = "6",
        aip4 = "10",
    },
}
local sentry_monstats2 = {
    { Id = "death-sentry-gfx", SizeX = "2", SizeY = "2", MeleeRng = "1" },
}

-- Health 202 with 53/50 percent bounds yields 107/101 and must consume no RNG draw.
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
        ["d2legacy.monster.stats"] = { level = 1, health = 0, max_health = 202 },
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

return test.suite({
    name = "Death Sentry reversed corpse range",
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
                Param1 = "53",
                Param2 = "50",
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
        ["data/global/excel/MonStats.txt"] = sentry_monstats,
        ["data/global/excel/MonStats2.txt"] = sentry_monstats2,
        ["data/global/excel/monstats.txt"] = sentry_monstats,
        ["data/global/excel/monstats2.txt"] = sentry_monstats2,
        ["data/global/excel/monlvl.txt"] = {
            { Level = "1", ["L-AC"] = "10", ["L-TH"] = "20" },
        },
        ["data/global/excel/SkillDesc.txt"] = {
            { skilldesc = "death sentry", ["str name"] = "fixture-death-sentry" },
        },
    },
    cases = {
        test.case("death_sentry_reversed_corpse_range_skips_random_draw", {
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
                local bundle = ecs.get(events[1], "d2legacy.combat.damage_bundle")
                test.expect(bundle:get("physical_rolled_raw")):equals(53)
                test.expect(bundle:get("fire_rolled_raw")):equals(53)
                test.expect(event:get("rolled_damage_raw")):equals(106)
            end),
            test.expect_checkpoint_parity(1),
            test.run(function()
                local random = require("engine.authority_random/v1")
                -- A reversed corpse range leaves 132 as the stream's first value.
                test.expect(random.integer("d2legacy.combat.damage.roll", 1000)):equals(132)
            end),
        }),
    },
})
