local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local enchant = {
    Id = "52",
    skill = "Enchant",
    skilldesc = "enchant",
    srvstfunc = "",
    srvdofunc = "25",
    range = "none",
    TargetPet = "1",
    TargetAlly = "1",
    leftskill = "",
    general = "",
    InGame = "1",
    mana = "25",
    lvlmana = "1",
    manashift = "8",
    minmana = "1",
    aurastate = "enchant",
    auralencalc = "ln12",
    aurastat1 = "firemindam",
    aurastatcalc1 = "enma",
    aurastat2 = "firemaxdam",
    aurastatcalc2 = "exma",
    aurastat3 = "item_tohit_percent",
    aurastatcalc3 = "toht",
    EType = "fire",
    HitShift = "7",
    EMin = "16",
    EMax = "20",
    EMinLev1 = "3",
    EMinLev2 = "7",
    EMinLev3 = "11",
    EMinLev4 = "15",
    EMinLev5 = "19",
    EMaxLev1 = "5",
    EMaxLev2 = "9",
    EMaxLev3 = "13",
    EMaxLev4 = "17",
    EMaxLev5 = "21",
    EDmgSymPerCalc = "(skill('Warmth'.blvl))*par8",
    Param1 = "3600",
    Param2 = "600",
    Param8 = "9",
    ToHit = "20",
    LevToHit = "9",
    ToHitCalc = "",
}

local function unit(id, kind)
    local ecs = require("engine.ecs/v1")
    return ecs.create({
        ["d2legacy.monster.stats"] = {
            level = 1,
            health = 100 * 256,
            max_health = 100 * 256,
            defense = 0,
            attack_rating = 100,
        },
        ["d2legacy.world.position"] = { x = 11, y = 12 },
        ["d2legacy.world.location"] = { act = 1, level_id = 1 },
        ["d2legacy.world.collider"] = { radius = 0.5 },
        ["d2legacy.world.selectable"] = { id = id, kind = kind, label = id },
    })
end

local function entry()
    return fixtures.player_entry({
        mana = 100,
        max_mana = 100,
    })
end

local function learn_skills()
    local ecs = require("engine.ecs/v1")
    local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
    for _, skill in ipairs({
        { id = 37, level = 3, row = 0, right = false },
        { id = 52, level = 2, row = 1, right = true },
    }) do
        ecs.create({
            ["d2legacy.player.learned_skill"] = {
                owner = player,
                skill_id = skill.id,
                level = skill.level,
                list_row = skill.row,
                left_allowed = false,
                right_allowed = skill.right,
            },
        })
    end
end

return test.suite({
    profile = "authority",
    tier = "integration",
    covers = { "internal/game/skill", "internal/game/state", "internal/game/combat" },
    records = {
        ["data/global/excel/skills.txt"] = { { Id = "37", skill = "Warmth" }, enchant },
        ["data/global/excel/skilldesc.txt"] = {
            { skilldesc = "enchant", ListRow = "1", IconCel = "1" },
        },
        ["data/global/excel/states.txt"] = { { state = "enchant", id = "16", group = "" } },
    },
    cases = {
        test.case("friendly_target_receives_one_timed_state_and_three_stat_sources", {
            test.submit_system(fixtures.command("system.player.enter", entry(), { tick = 1, sequence = 1 })),
            test.step(1),
            test.run(function()
                learn_skills()
                unit("monster:ally", "friendly")
                unit("monster:hostile", "hostile")
            end),
            test.submit(fixtures.command("player.assign_skills", { right = 52 }, {
                tick = 2,
                sequence = 1,
                player = "alice",
            })),
            test.step(1),
            test.submit(fixtures.command("player.use_skill", { side = "right", target_id = "monster:ally" }, {
                tick = 3,
                sequence = 2,
                player = "alice",
            })),
            test.step(3),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local instances = ecs.query({ all = { "d2legacy.state.instance" } })
                test.expect(#instances):equals(1)
                local instance = ecs.get(instances[1], "d2legacy.state.instance")
                test.expect(instance:get("state_id")):equals("enchant")
                test.expect(instance:get("expires_tick") - instance:get("applied_tick")):equals(4200)
                local values = {}
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.stat.source" } })) do
                    local source = ecs.get(entity, "d2legacy.stat.source")
                    if source:get("owner_source_id") == "skill:alice:52" then
                        values[source:get("stat")] = source:get("value")
                        test.assert(source:get("target"):id() == instance:get("target"):id())
                    end
                end
                test.expect(values.firemindam):equals(3088)
                test.expect(values.firemaxdam):equals(4064)
                test.expect(values.item_tohit_percent):equals(29)
                local events = ecs.query({ all = { "d2legacy.skill.cast_event" } })
                test.expect(ecs.get(events[1], "d2legacy.skill.cast_event"):get("target_id")):equals("monster:ally")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                test.expect(ecs.get(player, "d2legacy.player.vitals"):get("mana_raw")):equals(74 * 256)
            end),
            test.submit(fixtures.command("player.use_skill", { side = "right", target_id = "monster:hostile" }, {
                tick = 6,
                sequence = 3,
                player = "alice",
            })),
            test.step(3),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local self_enchanted = false
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.state.instance" } })) do
                    local state = ecs.get(entity, "d2legacy.state.instance")
                    self_enchanted = self_enchanted
                        or state:get("state_id") == "enchant" and state:get("target"):id() == player:id()
                end
                test.assert(self_enchanted, [=[invalid hostile target falls back to the caster]=])
                test.expect(ecs.get(player, "d2legacy.player.vitals"):get("mana_raw")):equals(48 * 256)
            end),
            test.expect_checkpoint_parity(1),
        }),
    },
})
