local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local amplify = {
    Id = "66",
    skill = "Amplify Damage",
    skilldesc = "amplify damage",
    srvstfunc = "",
    srvdofunc = "30",
    aurafilter = "3",
    auratargetstate = "amplifydamage",
    auralencalc = "ln34",
    aurarangecalc = "ln12",
    aurastat1 = "damageresist",
    aurastatcalc1 = "-par5",
    range = "none",
    LineOfSight = "4",
    mana = "4",
    lvlmana = "0",
    minmana = "1",
    manashift = "8",
    interrupt = "1",
    Param1 = "3",
    Param2 = "1",
    Param3 = "200",
    Param4 = "75",
    Param5 = "100",
    InGame = "1",
}

local weaken = {
    Id = "72",
    skill = "Weaken",
    skilldesc = "weaken",
    srvstfunc = "",
    srvdofunc = "30",
    aurafilter = "3",
    auratargetstate = "weaken",
    auralencalc = "ln34",
    aurarangecalc = "ln12",
    aurastat1 = "damagepercent",
    aurastatcalc1 = "-par5",
    range = "none",
    LineOfSight = "4",
    mana = "4",
    lvlmana = "0",
    minmana = "1",
    manashift = "8",
    interrupt = "1",
    Param1 = "9",
    Param2 = "1",
    Param3 = "350",
    Param4 = "60",
    Param5 = "33",
    InGame = "1",
}

local function monster(id, x, base_physical_resist)
    local ecs = require("engine.ecs/v1")
    return ecs.create({
        ["d2legacy.monster.stats"] = {
            level = 1,
            health = 100 * 256,
            max_health = 100 * 256,
            defense = 0,
            attack_rating = 100,
        },
        ["d2legacy.combat.defense"] = {
            base_physical_resist = base_physical_resist,
            base_fire_resist = 0,
            physical_resist = base_physical_resist,
            fire_resist = 0,
            max_fire_resist = 75,
            physical_reduction_raw = 0,
        },
        ["d2legacy.world.position"] = { x = x, y = 12 },
        ["d2legacy.world.location"] = { act = 1, level_id = 1 },
        ["d2legacy.world.collider"] = { radius = 0.5 },
        ["d2legacy.world.selectable"] = { id = id, kind = "hostile", label = id, radius = 0.5 },
    })
end

return test.suite({
    profile = "authority",
    tier = "integration",
    covers = { "internal/game/skill", "internal/game/state", "internal/game/combat" },
    records = {
        ["data/global/excel/skills.txt"] = { amplify, weaken },
        ["data/global/excel/skilldesc.txt"] = {
            { skilldesc = "amplify damage", ListRow = "1", IconCel = "0" },
            { skilldesc = "weaken", ListRow = "2", IconCel = "1" },
        },
        ["data/global/excel/states.txt"] = {
            { state = "amplifydamage", id = "9", curse = "1" },
            { state = "weaken", id = "19", curse = "1" },
        },
    },
    cases = {
        test.case("point_area_curse_applies_ordered_timed_resistance_sources", {
            test.submit_system(
                fixtures.command("system.player.enter", fixtures.player_entry({ mana = 100, max_mana = 100 }), {
                    tick = 1,
                    sequence = 1,
                })
            ),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                ecs.create({
                    ["d2legacy.player.learned_skill"] = {
                        owner = player,
                        skill_id = 66,
                        level = 2,
                        list_row = 1,
                        left_allowed = false,
                        right_allowed = true,
                    },
                })
                monster("monster:normal", 12, 0)
                monster("monster:immune", 16, 100)
                monster("monster:outside", 20, 0)
            end),
            test.submit(fixtures.command("player.assign_skills", { right = 66 }, {
                tick = 2,
                sequence = 1,
                player = "alice",
            })),
            test.step(1),
            test.submit(fixtures.command("player.use_skill", { side = "right", target_x = 12, target_y = 12 }, {
                tick = 3,
                sequence = 2,
                player = "alice",
            })),
            test.step(3),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local by_id = {}
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.world.selectable" } })) do
                    by_id[ecs.get(entity, "d2legacy.world.selectable"):get("id")] = entity
                end
                test.expect(#ecs.query({ all = { "d2legacy.state.instance" } })):equals(2)
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.state.instance" } })) do
                    local state = ecs.get(entity, "d2legacy.state.instance")
                    test.expect(state:get("state_id")):equals("amplifydamage")
                    test.expect(state:get("expires_tick") - state:get("applied_tick")):equals(275)
                    test.expect(state:get("replacement_priority")):equals(2)
                    test.assert(state:get("reject_lower_priority"))
                end
                test.expect(ecs.get(by_id["monster:normal"], "d2legacy.combat.defense"):get("physical_resist"))
                    :equals(-100)
                test.expect(ecs.get(by_id["monster:immune"], "d2legacy.combat.defense"):get("physical_resist"))
                    :equals(80)
                test.expect(ecs.get(by_id["monster:outside"], "d2legacy.combat.defense"):get("physical_resist"))
                    :equals(0)
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                test.expect(ecs.get(player, "d2legacy.player.vitals"):get("mana_raw")):equals(96 * 256)
                local event =
                    ecs.get(ecs.query({ all = { "d2legacy.skill.cast_event" } })[1], "d2legacy.skill.cast_event")
                test.expect(event:get("behavior")):equals("state.point-area-curse")
                test.expect(event:get("target_x")):equals(12)
                test.expect(event:get("target_y")):equals(12)
            end),
            test.expect_checkpoint_parity(1),
        }),
        test.case("same_family_publishes_outgoing_damage_percentage", {
            test.submit_system(
                fixtures.command("system.player.enter", fixtures.player_entry({ mana = 20, max_mana = 20 }), {
                    tick = 1,
                    sequence = 1,
                })
            ),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                ecs.create({
                    ["d2legacy.player.learned_skill"] = {
                        owner = player,
                        skill_id = 72,
                        level = 1,
                        list_row = 2,
                        left_allowed = false,
                        right_allowed = true,
                    },
                })
                monster("monster:weakened", 12, 0)
            end),
            test.submit(fixtures.command("player.assign_skills", { right = 72 }, {
                tick = 2,
                sequence = 1,
                player = "alice",
            })),
            test.step(1),
            test.submit(fixtures.command("player.use_skill", { side = "right", target_x = 12, target_y = 12 }, {
                tick = 3,
                sequence = 2,
                player = "alice",
            })),
            test.step(3),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local state = ecs.get(ecs.query({ all = { "d2legacy.state.instance" } })[1], "d2legacy.state.instance")
                test.expect(state:get("state_id")):equals("weaken")
                test.expect(state:get("expires_tick") - state:get("applied_tick")):equals(350)
                local sources = ecs.query({ all = { "d2legacy.stat.source" } })
                test.expect(#sources):equals(1)
                local source = ecs.get(sources[1], "d2legacy.stat.source")
                test.expect(source:get("stat")):equals("damagepercent")
                test.expect(source:get("operation")):equals("percent")
                test.expect(source:get("value")):equals(-33)
            end),
            test.expect_checkpoint_parity(1),
        }),
    },
})
