local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local function elemental_skill(id, name, description, missile, minimum, maximum)
    return {
        Id = tostring(id),
        skill = name,
        skilldesc = description,
        leftskill = "1",
        general = "1",
        passive = "0",
        srvmissile = missile,
        EType = "fire",
        interrupt = "1",
        srvstfunc = "",
        srvdofunc = "",
        mana = id == 47 and "10" or "5",
        lvlmana = id == 47 and "1" or "0",
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

local fire_bolt = elemental_skill(36, "Fire Bolt", "firebolt", "firebolt", 6, 12)
fire_bolt.EDmgSymPerCalc = "(skill('Fire Ball'.blvl)+skill('Meteor'.blvl))*par8"
fire_bolt.Param8 = "16"

local fire_ball = elemental_skill(47, "Fire Ball", "fireball", "fireball", 12, 28)
fire_ball.EDmgSymPerCalc = "(skill('Fire Bolt'.blvl)+skill('Meteor'.blvl))*par8"
fire_ball.Param8 = "14"

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

return test.suite({
    profile = "authority",
    tier = "integration",
    covers = { "internal/game/combat", "internal/game/skill", "internal/game/missile" },
    records = {
        ["data/global/excel/skills.txt"] = {
            fire_bolt,
            fire_ball,
            {
                Id = "56",
                skill = "Meteor",
                skilldesc = "meteor",
                leftskill = "1",
                general = "1",
                passive = "0",
            },
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
            {
                Missile = "fireball",
                Skill = "Fire Ball",
                pSrvDoFunc = "1",
                pSrvHitFunc = "1",
                sHitPar1 = "4",
                CollideType = "3",
                CollideKill = "1",
                Vel = "20",
                Range = "50",
                Size = "1",
                CelFile = "Fireball",
                ExplosionMissile = "explodingarrowexp",
                HitSound = "sorceress_fireball_impact_1",
            },
            {
                Missile = "explodingarrowexp",
                Explosion = "1",
                Range = "16",
                Size = "1",
                CelFile = "ExpArrowExplode",
                NumDirections = "1",
                AnimSpeed = "16",
                LoopAnim = "0",
                TravelSound = "explosion_medium_1",
            },
        },
    },
    cases = {
        test.case("straight_impact_family_applies_one_area_result_and_visual_effect", {
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
            test.submit(fixtures.command("player.assign_skills", { left = 47 }, {
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
            test.step(4),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local events = ecs.query({ all = { "d2legacy.combat.event" } })
                test.assert(#events == 2, [=[the impact damages exactly the two targets inside its radius]=])
                local targets = {}
                for _, entity in ipairs(events) do
                    local event = ecs.get(entity, "d2legacy.combat.event")
                    targets[event:get("target_id")] = true
                    test.expect(event:get("damage_channel")):equals("fire")
                    test.assert(
                        event:get("rolled_damage_raw") >= 1966 and event:get("rolled_damage_raw") <= 4587,
                        [=[Fire Ball uses its level bands and 28% hard-point synergy snapshot]=]
                    )
                end
                test.assert(targets["monster:center"] and targets["monster:near"] and not targets["monster:outside"])
                test.assert(
                    #ecs.query({ all = { "d2legacy.missile.projectile" } }) == 0,
                    [=[the contact projectile is destroyed]=]
                )
                local effects = ecs.query({ all = { "d2legacy.missile.effect" } })
                test.assert(#effects == 1, [=[one non-damaging impact effect is materialized]=])
                local effect = ecs.get(effects[1], "d2legacy.missile.effect")
                test.expect(effect:get("missile_id")):equals("explodingarrowexp")
                test.expect(effect:get("remaining_ticks")):equals(16)
                local snapshots = require("d2legacy.gameplay.world").missile_snapshots()
                test.assert(#snapshots == 1, [=[the presentation projection includes only the impact effect]=])
                test.expect(snapshots[1].missile_id):equals("explodingarrowexp")
                local recipe = require("d2legacy.gameplay.missile_presentation").resolve(snapshots[1])
                test.expect(recipe.path):equals("data/global/missiles/ExpArrowExplode.dcc")
            end),
            test.expect_checkpoint_parity(1),
            test.step(15),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.assert(
                    #ecs.query({ all = { "d2legacy.missile.effect" } }) == 0,
                    [=[the record-authored impact effect expires]=]
                )
                test.assert(
                    #ecs.query({ all = { "d2legacy.combat.event" } }) == 2,
                    [=[presentation lifetime cannot reapply area damage]=]
                )
            end),
        }),
    },
})
