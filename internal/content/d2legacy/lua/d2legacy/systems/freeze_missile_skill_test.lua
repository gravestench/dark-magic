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
        mana = id == 45 and "12" or "1",
        lvlmana = id == 45 and "1" or "0",
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

local ice_blast = skill(45, "Ice Blast", "iceblast", "iceblast", 16, 24)
ice_blast.EDmgSymPerCalc = "(skill('Ice Bolt'.blvl)+skill('Blizzard'.blvl)+skill('Frozen Orb'.blvl))*par8"
ice_blast.Param8 = "8"
ice_blast.ELen = "75"
ice_blast.ELevLen1 = "5"
ice_blast.ELevLen2 = "5"
ice_blast.ELevLen3 = "5"
ice_blast.ELenSymPerCalc = "(skill('Glacial Spike'.blvl))*par7"
ice_blast.Param7 = "10"

local function monster(life)
    life = life or 20000 * 256
    return fixtures.monster_spawn({
        spawn_id = "freeze-target",
        x = 3,
        y = 0,
        definition = {
            id = "fallen",
            base_id = "fallen",
            graphics_id = "fallen",
            name_key = "Fallen",
            ai = "fallen",
            token = "FA",
            weapon_class = "HTH",
            components = {},
            life_min = life,
            life_max = life,
            level = 1,
            defense = 0,
            attack_rating = 0,
            physical_min = 0,
            physical_max = 0,
            experience = 0,
            treasure_class = "",
            collider_radius = 0.5,
            select_radius = 0.5,
            velocity = 2,
            think_interval = 1,
            aggro_radius = 20,
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
            skill(39, "Ice Bolt", "icebolt", "icebolt", 1, 1),
            ice_blast,
            skill(55, "Glacial Spike", "glacialspike", "", 1, 1),
            skill(59, "Blizzard", "blizzard", "blizzard", 1, 1),
            skill(64, "Frozen Orb", "frozenorb", "frozenorb", 1, 1),
        },
        ["data/global/excel/skilldesc.txt"] = descriptions,
        ["data/global/excel/Missiles.txt"] = {
            {
                Missile = "iceblast",
                Skill = "Ice Blast",
                pSrvDoFunc = "1",
                pSrvDmgFunc = "4",
                CollideType = "3",
                CollideKill = "1",
                Vel = "12",
                Range = "50",
                Size = "1",
                CelFile = "IceBlast",
                ExplosionMissile = "freezingarrowexp1",
                HitSound = "sorceress_iceblast_impact_1",
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
        test.case("straight_freeze_family_snapshots_damage_duration_and_control_state", {
            test.submit_system(fixtures.command("system.player.enter", fixtures.player_entry({ x = 0, y = 0 }), {
                tick = 1,
                sequence = 1,
            })),
            test.submit_system(fixtures.command("system.monster.spawn", monster(), {
                tick = 1,
                sequence = 2,
            })),
            test.step(1),
            test.submit(fixtures.command("player.assign_skills", { left = 45 }, {
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
            test.step(8),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local events = ecs.query({ all = { "d2legacy.combat.event" } })
                test.assert(#events == 1, [=[one direct cold result is emitted]=])
                local event = ecs.get(events[1], "d2legacy.combat.event")
                test.expect(event:get("damage_channel")):equals("cold")
                test.assert(
                    event:get("rolled_damage_raw") >= 2539 and event:get("rolled_damage_raw") <= 3809,
                    [=[three 8% hard-point synergies modify the generic damage range]=]
                )
                local instances = ecs.query({ all = { "d2legacy.state.instance" } })
                test.assert(#instances == 1, [=[one freeze state is materialized]=])
                local instance = ecs.get(instances[1], "d2legacy.state.instance")
                test.expect(instance:get("state_id")):equals("freeze")
                test.expect(instance:get("source_id")):equals("skill:player:alice:45:on-hit")
                test.expect(instance:get("expires_tick") - instance:get("applied_tick")):equals(82)
                test.assert(instance:get("action_disabled"), [=[freeze suppresses monster action]=])
                local effects = ecs.query({ all = { "d2legacy.missile.effect" } })
                test.assert(#effects == 1, [=[the direct hit owns one presentation-only freeze effect]=])
                test.expect(ecs.get(effects[1], "d2legacy.missile.effect"):get("missile_id"))
                    :equals("freezingarrowexp1")
            end),
            test.expect_checkpoint_parity(1),
            test.step(82),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.assert(#ecs.query({ all = { "d2legacy.state.instance" } }) == 0, [=[freeze expires exactly once]=])
                test.assert(
                    #ecs.query({ all = { "d2legacy.combat.event" } }) == 1,
                    [=[expiration cannot replay damage]=]
                )
            end),
        }),
        test.case("lethal_contact_does_not_materialize_a_control_state", {
            test.submit_system(fixtures.command("system.player.enter", fixtures.player_entry({ x = 0, y = 0 }), {
                tick = 1,
                sequence = 1,
            })),
            test.submit_system(fixtures.command("system.monster.spawn", monster(256), {
                tick = 1,
                sequence = 2,
            })),
            test.step(1),
            test.submit(fixtures.command("player.assign_skills", { left = 45 }, {
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
            test.step(8),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local events = ecs.query({ all = { "d2legacy.combat.event" } })
                test.assert(#events == 1, [=[the lethal contact still emits one damage result]=])
                test.expect(ecs.get(events[1], "d2legacy.combat.event"):get("kind")):equals("unit_died")
                test.assert(
                    #ecs.query({ all = { "d2legacy.state.instance" } }) == 0,
                    [=[a dead target receives no control state]=]
                )
            end),
        }),
    },
})
