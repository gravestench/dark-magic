local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local fire_bolt = {
    Id = "36",
    skill = "Fire Bolt",
    skilldesc = "firebolt",
    leftskill = "1",
    general = "0",
    passive = "0",
    srvmissile = "firebolt",
    EType = "fire",
    interrupt = "1",
    srvstfunc = "",
    srvdofunc = "",
    mana = "5",
    lvlmana = "0",
    minmana = "1",
    manashift = "7",
    EMin = "6",
    EMax = "12",
    EMinLev1 = "3",
    EMinLev2 = "4",
    EMinLev3 = "8",
    EMinLev4 = "18",
    EMinLev5 = "54",
    EMaxLev1 = "3",
    EMaxLev2 = "6",
    EMaxLev3 = "10",
    EMaxLev4 = "20",
    EMaxLev5 = "56",
    HitShift = "7",
    EDmgSymPerCalc = "(skill('Fire Ball'.blvl)+skill('Meteor'.blvl))*par8",
    Param8 = "16",
}

local function supporting_skill(id, name, description)
    return {
        Id = tostring(id),
        skill = name,
        skilldesc = description,
        leftskill = "1",
        general = "1",
        passive = "0",
    }
end

return test.suite({
    profile = "authority",
    tier = "integration",
    covers = { "internal/game/skill", "internal/game/missile" },
    records = {
        ["data/global/excel/skills.txt"] = {
            fire_bolt,
            supporting_skill(47, "Fire Ball", "fireball"),
            supporting_skill(56, "Meteor", "meteor"),
        },
        ["data/global/excel/skilldesc.txt"] = {
            { skilldesc = "firebolt", ListRow = "0", IconCel = "0" },
            { skilldesc = "fireball", ListRow = "1", IconCel = "1" },
            { skilldesc = "meteor", ListRow = "2", IconCel = "2" },
        },
        ["data/global/excel/Missiles.txt"] = {
            {
                Missile = "firebolt",
                Skill = "Fire Bolt",
                pSrvDoFunc = "1",
                CollideType = "3",
                CollideKill = "1",
                Vel = "20",
                Range = "40",
                Size = "2",
                CelFile = "firebolt",
            },
        },
    },
    cases = {
        test.case("hard_point_synergies_modify_the_generic_damage_snapshot", {
            test.submit_system(fixtures.command("system.player.enter", fixtures.player_entry({ x = 0, y = 0 }), {
                tick = 1,
                sequence = 1,
            })),
            test.step(1),
            test.submit(fixtures.command("player.use_skill", {
                side = "left",
                target_x = 30,
                target_y = 0,
            }, {
                tick = 2,
                sequence = 1,
                player = "alice",
            })),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local cast = ecs.get(player, "d2legacy.skill.cast")
                test.expect(cast:get("skill_id")):equals(36)
                test.expect(cast:get("elemental_damage_percent")):equals(32)
                local projectiles = ecs.query({ all = { "d2legacy.missile.projectile" } })
                test.assert(#projectiles == 1, [=[#projectiles == 1]=])
                local projectile = ecs.get(projectiles[1], "d2legacy.missile.projectile")
                test.expect(projectile:get("minimum_damage_raw")):equals(1013)
                test.expect(projectile:get("maximum_damage_raw")):equals(2027)
                local vitals = ecs.get(player, "d2legacy.player.vitals")
                test.expect(vitals:get("mana_raw")):equals(4480)
            end),
            test.expect_checkpoint_parity(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local projectiles = ecs.query({ all = { "d2legacy.missile.projectile" } })
                test.assert(#projectiles == 1, [=[#projectiles == 1]=])
                local projectile = ecs.get(projectiles[1], "d2legacy.missile.projectile")
                test.expect(projectile:get("minimum_damage_raw")):equals(1013)
                test.expect(projectile:get("maximum_damage_raw")):equals(2027)
            end),
        }),
    },
})
