local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local function skill(id, name, description, missile, minimum, maximum)
    return {
        Id = tostring(id),
        skill = name,
        skilldesc = description,
        leftskill = "1",
        general = "1",
        passive = "0",
        srvmissile = missile,
        EType = "cold",
        interrupt = "1",
        srvstfunc = "",
        srvdofunc = "",
        mana = id == 55 and "20" or "1",
        lvlmana = id == 55 and "1" or "0",
        minmana = "1",
        manashift = "7",
        EMin = tostring(minimum),
        EMax = tostring(maximum),
        EMinLev1 = "0",
        EMinLev2 = "0",
        EMinLev3 = "0",
        EMinLev4 = "0",
        EMinLev5 = "0",
        EMaxLev1 = "0",
        EMaxLev2 = "0",
        EMaxLev3 = "0",
        EMaxLev4 = "0",
        EMaxLev5 = "0",
        HitShift = "7",
    }
end

local glacial_spike = skill(55, "Glacial Spike", "glacialspike", "glacialspike", 32, 48)
glacial_spike.EDmgSymPerCalc = "(skill('Ice Bolt'.blvl)+skill('Ice Blast'.blvl)+skill('Frozen Orb'.blvl))*par8"
glacial_spike.Param8 = "5"
glacial_spike.aurarangecalc = "ln12"
glacial_spike.Param1 = "4"
glacial_spike.Param2 = "0"
glacial_spike.Param3 = "50"
glacial_spike.Param4 = "3"
glacial_spike.Param7 = "3"
glacial_spike.auralencalc = "ln34 * (100 + skill('Blizzard'.blvl) * par7) / 100"

local function monster(id, x, y)
    return fixtures.monster_spawn({
        spawn_id = id,
        x = x,
        y = y,
        definition = {
            id = "fallen",
            base_id = "fallen",
            graphics_id = "fallen",
            name_key = "Fallen",
            ai = "fallen",
            token = "FA",
            weapon_class = "HTH",
            components = {},
            life_min = 20000 * 256,
            life_max = 20000 * 256,
            level = 1,
            defense = 0,
            attack_rating = 0,
            physical_min = 0,
            physical_max = 0,
            experience = 0,
            treasure_class = "",
            collider_radius = 0.5,
            select_radius = 0.5,
            velocity = 0,
            think_interval = 100,
            aggro_radius = 0,
            attack_range = 1,
        },
    })
end

local descriptions = {
    { skilldesc = "icebolt", ListRow = "0", IconCel = "0" },
    { skilldesc = "iceblast", ListRow = "1", IconCel = "1" },
    { skilldesc = "glacialspike", ListRow = "2", IconCel = "2" },
    { skilldesc = "blizzard", ListRow = "3", IconCel = "3" },
    { skilldesc = "frozenorb", ListRow = "4", IconCel = "4" },
}

return test.suite({
    profile = "authority",
    tier = "integration",
    covers = { "internal/game/combat", "internal/game/skill", "internal/game/missile", "internal/game/state" },
    records = {
        ["data/global/excel/skills.txt"] = {
            skill(39, "Ice Bolt", "icebolt", "", 1, 1),
            skill(45, "Ice Blast", "iceblast", "", 1, 1),
            glacial_spike,
            skill(59, "Blizzard", "blizzard", "", 1, 1),
            skill(64, "Frozen Orb", "frozenorb", "", 1, 1),
        },
        ["data/global/excel/skilldesc.txt"] = descriptions,
        ["data/global/excel/Missiles.txt"] = {
            {
                Missile = "glacialspike",
                Skill = "Glacial Spike",
                pSrvDoFunc = "1",
                pSrvHitFunc = "13",
                EType = "frze",
                HitFlags = "2",
                CollideType = "3",
                CollideKill = "1",
                Vel = "16",
                Range = "40",
                Size = "1",
                CelFile = "GlacialSpike",
                CltHitSubMissile1 = "freezingarrowexp1",
                HitSound = "sorceress_glacialspike_impact_1",
            },
            {
                Missile = "freezingarrowexp1",
                Explosion = "1",
                Range = "16",
                Size = "1",
                CelFile = "FreezeExplodeCenter",
                NumDirections = "1",
                AnimSpeed = "16",
                LoopAnim = "0",
            },
        },
    },
    cases = {
        test.case("area_freeze_composes_shared_impact_damage_and_timed_state", {
            test.submit_system(fixtures.command("system.player.enter", fixtures.player_entry({ x = 0, y = 0 }), {
                tick = 1,
                sequence = 1,
            })),
            test.submit_system(fixtures.command("system.monster.spawn", monster("center", 3, 0), {
                tick = 1,
                sequence = 2,
            })),
            test.submit_system(fixtures.command("system.monster.spawn", monster("near", 3, 3.5), {
                tick = 1,
                sequence = 3,
            })),
            test.submit_system(fixtures.command("system.monster.spawn", monster("outside", 3, 6), {
                tick = 1,
                sequence = 4,
            })),
            test.step(1),
            test.submit(fixtures.command("player.assign_skills", { left = 55 }, {
                tick = 2,
                sequence = 1,
                player = "alice",
            })),
            test.step(1),
            test.submit(fixtures.command("player.use_skill", {
                side = "left",
                target_x = 10,
                target_y = 0,
            }, {
                tick = 3,
                sequence = 2,
                player = "alice",
            })),
            test.step(5),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local events = ecs.query({ all = { "d2legacy.combat.event" } })
                test.assert(#events == 2, [=[exactly the two in-radius targets receive cold damage]=])
                local targets = {}
                for _, entity in ipairs(events) do
                    local event = ecs.get(entity, "d2legacy.combat.event")
                    targets[event:get("target_id")] = true
                    test.expect(event:get("damage_channel")):equals("cold")
                    test.assert(
                        event:get("rolled_damage_raw") >= 4710 and event:get("rolled_damage_raw") <= 7065,
                        [=[three 5% hard-point synergies modify the shared damage range]=]
                    )
                end
                test.assert(targets["monster:center"] and targets["monster:near"] and not targets["monster:outside"])

                local instances = ecs.query({ all = { "d2legacy.state.instance" } })
                test.assert(#instances == 2, [=[each nonlethal in-radius result receives the shared freeze state]=])
                for _, entity in ipairs(instances) do
                    local instance = ecs.get(entity, "d2legacy.state.instance")
                    test.expect(instance:get("state_id")):equals("freeze")
                    test.expect(instance:get("source_id")):equals("skill:player:alice:55:on-hit")
                    test.expect(instance:get("expires_tick") - instance:get("applied_tick")):equals(51)
                    test.assert(instance:get("action_disabled"), [=[freeze suppresses monster action]=])
                end

                local effects = ecs.query({ all = { "d2legacy.missile.effect" } })
                test.assert(#effects == 1, [=[the impact owns one presentation-only center effect]=])
                test.expect(ecs.get(effects[1], "d2legacy.missile.effect"):get("missile_id"))
                    :equals("freezingarrowexp1")
            end),
            test.expect_checkpoint_parity(1),
            test.step(51),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.assert(#ecs.query({ all = { "d2legacy.state.instance" } }) == 0, [=[both freezes expire]=])
                test.assert(
                    #ecs.query({ all = { "d2legacy.combat.event" } }) == 2,
                    [=[state and presentation lifetime cannot replay area damage]=]
                )
            end),
        }),
    },
})
