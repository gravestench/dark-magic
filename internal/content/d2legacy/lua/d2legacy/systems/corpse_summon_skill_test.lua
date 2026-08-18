local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local raise = {
    Id = "70",
    skill = "Fixture Raise",
    charclass = "nec",
    srvstfunc = "15",
    srvdofunc = "31",
    TargetCorpse = "1",
    SelectProc = "2",
    range = "none",
    leftskill = "",
    InGame = "1",
    InTown = "",
    summon = "fixturepet",
    pettype = "skeleton",
    summode = "S1",
    mana = "6",
    lvlmana = "1",
    minmana = "1",
    manashift = "8",
    petmax = "(lvl < 4) ?lvl:(2+lvl/3)",
    calc1 = "(lvl < 4) ? 0 : (par2 * (lvl - 3))",
    aurastat1 = "damagepercent",
    aurastatcalc1 = "((lvl < 4) ? 0 : ((lvl-3)*par3))",
    aurastat2 = "tohit",
    aurastatcalc2 = "(lvl+skill('Fixture Mastery'.lvl))*par4",
    aurastat3 = "armorclass",
    aurastatcalc3 = "(lvl+skill('Fixture Mastery'.lvl))*par5",
    passivestat1 = "maxhp",
    passivecalc1 = "skill('Fixture Mastery'.lvl) * skill('Fixture Mastery'.par1) * 256",
    passivestat2 = "item_normaldamage",
    passivecalc2 = "skill('Fixture Mastery'.lvl) * skill('Fixture Mastery'.par2) + edmn",
    Param2 = "50",
    Param3 = "7",
    Param4 = "15",
    Param5 = "15",
}

local function learn()
    local ecs = require("engine.ecs/v1")
    local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
    for row, skill in ipairs({
        { id = 69, level = 3 },
        { id = 70, level = 20 },
        { id = 80, level = 20 },
        { id = 89, level = 2 },
        { id = 95, level = 20 },
    }) do
        ecs.create({
            ["d2legacy.player.learned_skill"] = {
                owner = player,
                skill_id = skill.id,
                level = skill.level,
                list_row = row,
                left_allowed = false,
                right_allowed = true,
            },
        })
    end
end

local function corpse(id, level_id, revivable)
    local ecs = require("engine.ecs/v1")
    local components = {
        ["d2legacy.monster.identity"] = {
            spawn_id = id,
            definition_id = revivable and "fixturecorpse" or "fallen",
            seed = id,
        },
        ["d2legacy.monster.stats"] = { level = 1, health = 0, max_health = 10 * 256 },
        ["d2legacy.monster.death"] = { tick = 1, active = false, corpse_usable = true },
        ["d2legacy.monster.corpse_selectable"] = {},
        ["d2legacy.world.position"] = { x = 12, y = 12 },
        ["d2legacy.world.location"] = { act = 1, level_id = level_id or 2 },
        ["d2legacy.world.selectable"] = { id = "monster:" .. id, kind = "corpse", label = id, priority = 5 },
    }
    if revivable then
        components["d2legacy.monster.revivable"] = {}
    end
    return ecs.create(components)
end

local function existing_skeletons(count)
    local ecs = require("engine.ecs/v1")
    local owner = ecs.query({ all = { "d2legacy.player.identity" } })[1]
    for index = 1, count do
        ecs.create({
            ["d2legacy.world.selectable"] = {
                id = "monster:old-" .. index,
                kind = "friendly",
                label = "old",
                priority = 20,
            },
            ["d2legacy.owned_unit"] = {
                owner = owner,
                owner_id = "player:alice",
                ultimate_owner_id = "player:alice",
                category = "skeleton",
                group = 0,
                limit = count,
                replacement = "replace_oldest",
                created_tick = index,
                active = true,
            },
        })
    end
end

local function lethal_result_from_owned_summon()
    local ecs = require("engine.ecs/v1")
    local summon = ecs.query({ all = { "d2legacy.owned_unit", "d2legacy.world.selectable" } })[1]
    local summon_id = ecs.get(summon, "d2legacy.world.selectable"):get("id")
    ecs.create({
        ["d2legacy.monster.identity"] = {
            spawn_id = "victim",
            definition_id = "victim",
            seed = "victim",
            treasure_class = "",
        },
        ["d2legacy.monster.stats"] = {
            level = 1,
            spawn_player_count = 1,
            health = 0,
            max_health = 10 * 256,
            defense = 0,
            attack_rating = 0,
            physical_min = 0,
            physical_max = 0,
            experience = 10,
        },
        ["d2legacy.monster.appearance"] = { mode = "NU" },
        ["d2legacy.world.position"] = { x = 13, y = 12 },
        ["d2legacy.world.velocity"] = { x = 0, y = 0 },
        ["d2legacy.world.location"] = { act = 1, level_id = 2 },
        ["d2legacy.world.collider"] = { radius = 1 },
        ["d2legacy.world.selectable"] = { id = "monster:victim", kind = "hostile", label = "victim" },
    })
    ecs.create({
        ["d2legacy.combat.event"] = {
            kind = "unit_died",
            tick = 6,
            attacker_id = summon_id,
            target_id = "monster:victim",
            source_kind = "melee",
            damage_channel = "physical",
            remaining_health_raw = 0,
        },
    })
end

local function alive_hostile()
    local ecs = require("engine.ecs/v1")
    local summon = ecs.query({ all = { "d2legacy.owned_unit", "d2legacy.monster.ai" } })[1]
    ecs.get(summon, "d2legacy.monster.ai"):set("next_think_tick", 0)
    ecs.create({
        ["d2legacy.monster.identity"] = { spawn_id = "ai-target", definition_id = "target", seed = "target" },
        ["d2legacy.monster.stats"] = {
            level = 1,
            spawn_player_count = 1,
            health = 100 * 256,
            max_health = 100 * 256,
            defense = 0,
            attack_rating = 0,
            physical_min = 0,
            physical_max = 0,
            experience = 0,
        },
        ["d2legacy.combat.melee_profile"] = { range = 1, physical_min = 0, physical_max = 0 },
        ["d2legacy.world.position"] = { x = 20, y = 12 },
        ["d2legacy.world.velocity"] = { x = 0, y = 0 },
        ["d2legacy.world.location"] = { act = 1, level_id = 2 },
        ["d2legacy.world.collider"] = { radius = 1 },
        ["d2legacy.world.selectable"] = { id = "monster:ai-target", kind = "hostile", label = "target" },
    })
end

return test.suite({
    profile = "authority",
    tier = "integration",
    records = {
        ["data/global/excel/skills.txt"] = {
            { Id = "69", skill = "Fixture Mastery", Param1 = "8", Param2 = "2", Param3 = "5", Param4 = "10" },
            raise,
            {
                Id = "80",
                skill = "Fixture Mage",
                charclass = "nec",
                srvstfunc = "15",
                srvdofunc = "31",
                TargetCorpse = "1",
                SelectProc = "2",
                range = "none",
                leftskill = "",
                InGame = "1",
                InTown = "",
                summon = "fixturemage",
                pettype = "skeletonmage",
                summode = "S1",
                mana = "8",
                lvlmana = "1",
                minmana = "1",
                manashift = "8",
                petmax = "(lvl < 4) ?lvl:(2+lvl/3)",
                calc1 = "(lvl < 4) ? 0 : (par2 * (lvl - 3))",
                aurastat1 = "armorclass",
                aurastatcalc1 = "(lvl+skill('Fixture Mastery'.lvl))*par5",
                passivestat1 = "maxhp",
                passivecalc1 = "skill('Fixture Mastery'.lvl) * skill('Fixture Mastery'.par1) * 256",
                sumskill1 = "FixtureMageMissile",
                sumsk1calc = "skill('Fixture Mastery'.lvl) + ((lvl < 4)?0:((lvl-2)/2))",
                Param2 = "7",
                Param5 = "10",
            },
            {
                Id = "89",
                skill = "Fixture Resist",
                charclass = "nec",
                passive = "1",
                passivestat1 = "passive_summon_resist",
                passivecalc1 = "dm12",
                Param1 = "20",
                Param2 = "75",
            },
            {
                Id = "95",
                skill = "Fixture Revive",
                charclass = "nec",
                srvstfunc = "21",
                srvdofunc = "58",
                TargetCorpse = "1",
                SelectProc = "3",
                range = "none",
                leftskill = "",
                InGame = "1",
                InTown = "",
                summon = "",
                pettype = "revive",
                summode = "NU",
                mana = "45",
                lvlmana = "0",
                minmana = "1",
                manashift = "8",
                petmax = "lvl",
                calc1 = "par1+skill('Fixture Mastery'.lvl) * skill('Fixture Mastery'.par3)",
                calc2 = "ln34",
                aurastat1 = "damagepercent",
                aurastatcalc1 = "skill('Fixture Mastery'.lvl) * skill('Fixture Mastery'.par4)",
                passivestat1 = "velocitypercent",
                passivecalc1 = "par5",
                Param1 = "200",
                Param3 = "4500",
                Param4 = "0",
                Param5 = "50",
            },
        },
        ["data/global/excel/PetType.txt"] = {
            { ["pet type"] = "skeleton", group = "0", basemax = "0", unsummon = "1", warp = "1" },
            { ["pet type"] = "skeletonmage", group = "0", basemax = "0", unsummon = "1", warp = "1" },
            { ["pet type"] = "revive", group = "0", basemax = "0", unsummon = "1", warp = "1" },
        },
        ["data/global/excel/states.txt"] = test.array(),
        ["data/global/excel/monstats.txt"] = {
            {
                Id = "fixturepet",
                BaseId = "fixturepet",
                MonStatsEx = "fixturepet",
                NameStr = "FixturePet",
                AI = "NecroPet",
                Align = "1",
                Code = "SK",
                enabled = "1",
                npc = "",
                noRatio = "1",
                minHP = "21",
                maxHP = "21",
                AC = "5",
                A1TH = "5",
                A1MinD = "1",
                A1MaxD = "2",
                Velocity = "13",
                aidel = "15",
                aidist = "35",
            },
            {
                Id = "fixturemage",
                BaseId = "fixturemage",
                MonStatsEx = "fixturemage",
                NameStr = "FixtureMage",
                AI = "NecroPet",
                Align = "1",
                Code = "SM",
                enabled = "1",
                npc = "",
                noRatio = "1",
                minHP = "25",
                maxHP = "25",
                AC = "7",
                A1TH = "0",
                A1MinD = "0",
                A1MaxD = "0",
                Velocity = "12",
                aidel = "15",
                aidist = "35",
            },
            {
                Id = "fixturecorpse",
                BaseId = "fixturecorpse",
                MonStatsEx = "fixturecorpse",
                NameStr = "FixtureCorpse",
                AI = "Fallen",
                Align = "0",
                Code = "FC",
                enabled = "1",
                npc = "",
                noRatio = "1",
                minHP = "10",
                maxHP = "10",
                AC = "4",
                A1TH = "6",
                A1MinD = "2",
                A1MaxD = "4",
                Velocity = "10",
                aidel = "10",
                aidist = "35",
            },
        },
        ["data/global/excel/monstats2.txt"] = {
            {
                Id = "fixturepet",
                BaseW = "1hs",
                SizeX = "2",
                SizeY = "2",
                MeleeRng = "1",
                OverlayHeight = "2",
                mKB = "1",
                HD = "1",
                HDv = "lit",
            },
            {
                Id = "fixturemage",
                BaseW = "hth",
                SizeX = "2",
                SizeY = "2",
                MeleeRng = "1",
                OverlayHeight = "2",
                mKB = "1",
                HD = "1",
                HDv = "lit",
            },
            {
                Id = "fixturecorpse",
                BaseW = "hth",
                SizeX = "2",
                SizeY = "2",
                MeleeRng = "1",
                OverlayHeight = "1",
                mKB = "1",
                corpseSel = "1",
                revive = "1",
                HD = "1",
                HDv = "lit",
            },
        },
        ["data/global/excel/monlvl.txt"] = {
            { Level = "1", ["L-AC"] = "10", ["L-TH"] = "20" },
            { Level = "10", ["L-AC"] = "100", ["L-TH"] = "200" },
        },
    },
    cases = {
        test.case("a_full_pet_category_replaces_the_deterministic_oldest_unit", {
            test.submit_system(fixtures.command(
                "system.player.enter",
                fixtures.player_entry({
                    level = 10,
                    mana = 100,
                    max_mana = 100,
                    level_id = 2,
                }),
                { tick = 1, sequence = 1 }
            )),
            test.step(1),
            test.run(function()
                learn()
                existing_skeletons(8)
                corpse("replacement")
            end),
            test.submit(
                fixtures.command("player.assign_skills", { right = 70 }, { tick = 2, sequence = 1, player = "alice" })
            ),
            test.step(1),
            test.submit(
                fixtures.command(
                    "player.use_skill",
                    { side = "right", target_id = "monster:replacement" },
                    { tick = 3, sequence = 2, player = "alice" }
                )
            ),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local units = ecs.query({ all = { "d2legacy.owned_unit", "d2legacy.world.selectable" } })
                test.expect(#units):equals(8)
                local found_oldest, found_new = false, false
                for _, unit in ipairs(units) do
                    local id = ecs.get(unit, "d2legacy.world.selectable"):get("id")
                    found_oldest = found_oldest or id == "monster:old-1"
                    found_new = found_new or id:match("^monster:summon:alice:70:") ~= nil
                end
                test.assert(not found_oldest, "oldest category member survived replacement")
                test.assert(found_new, "replacement summon is missing")
            end),
            test.expect_checkpoint_parity(1),
        }),
        test.case("town_rejects_a_valid_corpse_before_mana_payment", {
            test.submit_system(fixtures.command(
                "system.player.enter",
                fixtures.player_entry({
                    level = 10,
                    mana = 100,
                    max_mana = 100,
                    level_id = 1,
                }),
                { tick = 1, sequence = 1 }
            )),
            test.step(1),
            test.run(function()
                learn()
                corpse("town-body", 1)
            end),
            test.submit(
                fixtures.command("player.assign_skills", { right = 70 }, { tick = 2, sequence = 1, player = "alice" })
            ),
            test.step(1),
            test.submit(
                fixtures.command(
                    "player.use_skill",
                    { side = "right", target_id = "monster:town-body" },
                    { tick = 3, sequence = 2, player = "alice" }
                )
            ),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.vitals" } })[1]
                test.expect(ecs.get(player, "d2legacy.player.vitals"):get("mana_raw")):equals(100 * 256)
                test.expect(#ecs.query({ all = { "d2legacy.skill.cast" } })):equals(0)
                test.expect(#ecs.query({ all = { "d2legacy.owned_unit" } })):equals(0)
            end),
        }),
        test.case("corpse_is_revalidated_at_effect_time_after_mana_was_committed", {
            test.submit_system(fixtures.command(
                "system.player.enter",
                fixtures.player_entry({
                    level = 10,
                    mana = 100,
                    max_mana = 100,
                    level_id = 2,
                }),
                { tick = 1, sequence = 1 }
            )),
            test.step(1),
            test.run(function()
                learn()
                corpse("stale")
            end),
            test.submit(
                fixtures.command("player.assign_skills", { right = 70 }, { tick = 2, sequence = 1, player = "alice" })
            ),
            test.step(1),
            test.submit(
                fixtures.command(
                    "player.use_skill",
                    { side = "right", target_id = "monster:stale" },
                    { tick = 3, sequence = 2, player = "alice" }
                )
            ),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local body = ecs.query({ all = { "d2legacy.monster.death" } })[1]
                ecs.get(body, "d2legacy.monster.death"):set("corpse_usable", false)
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(#ecs.query({ all = { "d2legacy.owned_unit" } })):equals(0)
                local events = ecs.query({ all = { "d2legacy.skill.summon_event" } })
                test.expect(#events):equals(1)
                local event = ecs.get(events[1], "d2legacy.skill.summon_event")
                test.expect(event:get("outcome")):equals("invalidated")
                test.expect(event:get("reason")):equals("corpse_target_unavailable")
                local player = ecs.query({ all = { "d2legacy.player.vitals" } })[1]
                test.assert(ecs.get(player, "d2legacy.player.vitals"):get("mana_raw") < 100 * 256)
            end),
            test.expect_checkpoint_parity(1),
        }),
        test.case("underfunded_or_unusable_corpse_requests_do_not_start_or_spend", {
            test.submit_system(fixtures.command(
                "system.player.enter",
                fixtures.player_entry({
                    level = 10,
                    mana = 20,
                    max_mana = 20,
                    level_id = 2,
                }),
                { tick = 1, sequence = 1 }
            )),
            test.step(1),
            test.run(function()
                learn()
                corpse("body")
            end),
            test.submit(
                fixtures.command("player.assign_skills", { right = 70 }, { tick = 2, sequence = 1, player = "alice" })
            ),
            test.step(1),
            test.submit(
                fixtures.command(
                    "player.use_skill",
                    { side = "right", target_id = "monster:body" },
                    { tick = 3, sequence = 2, player = "alice" }
                )
            ),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.vitals" } })[1]
                test.expect(ecs.get(player, "d2legacy.player.vitals"):get("mana_raw")):equals(20 * 256)
                test.expect(#ecs.query({ all = { "d2legacy.skill.cast" } })):equals(0)
                test.expect(#ecs.query({ all = { "d2legacy.owned_unit" } })):equals(0)
                local body = ecs.query({ all = { "d2legacy.monster.death" } })[1]
                ecs.get(body, "d2legacy.monster.death"):set("corpse_usable", false)
                ecs.get(player, "d2legacy.player.vitals"):set("mana_raw", 100 * 256)
                ecs.get(player, "d2legacy.player.vitals"):set("mana", 100)
                ecs.get(player, "d2legacy.player.vitals"):set("max_mana_raw", 100 * 256)
                ecs.get(player, "d2legacy.player.vitals"):set("max_mana", 100)
            end),
            test.submit(
                fixtures.command(
                    "player.use_skill",
                    { side = "right", target_id = "monster:body" },
                    { tick = 5, sequence = 3, player = "alice" }
                )
            ),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.vitals" } })[1]
                test.expect(ecs.get(player, "d2legacy.player.vitals"):get("mana_raw")):equals(100 * 256)
                test.expect(#ecs.query({ all = { "d2legacy.owned_unit" } })):equals(0)
            end),
        }),
        test.case("skeletal_mage_uses_the_same_corpse_transaction_with_its_authored_skill_grant", {
            test.submit_system(fixtures.command(
                "system.player.enter",
                fixtures.player_entry({
                    level = 10,
                    mana = 100,
                    max_mana = 100,
                    level_id = 2,
                }),
                { tick = 1, sequence = 1 }
            )),
            test.step(1),
            test.run(function()
                learn()
                corpse("mage-body")
            end),
            test.submit(
                fixtures.command("player.assign_skills", { right = 80 }, { tick = 2, sequence = 1, player = "alice" })
            ),
            test.step(1),
            test.submit(
                fixtures.command(
                    "player.use_skill",
                    { side = "right", target_id = "monster:mage-body" },
                    { tick = 3, sequence = 2, player = "alice" }
                )
            ),
            test.step(3),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local summon = ecs.query({ all = { "d2legacy.owned_unit", "d2legacy.monster.granted_skill" } })[1]
                test.assert(summon ~= nil, "skeletal mage was not materialized")
                local identity = ecs.get(summon, "d2legacy.monster.identity")
                local stats = ecs.get(summon, "d2legacy.monster.stats")
                local relation = ecs.get(summon, "d2legacy.owned_unit")
                local granted = ecs.get(summon, "d2legacy.monster.granted_skill")
                test.expect(identity:get("definition_id")):equals("fixturemage")
                test.expect(relation:get("category")):equals("skeletonmage")
                test.expect(relation:get("limit")):equals(8)
                test.expect(stats:get("max_health")):equals(78 * 256)
                test.expect(stats:get("defense")):equals(337)
                test.expect(granted:get("skill")):equals("FixtureMageMissile")
                test.expect(granted:get("level")):equals(12)
            end),
            test.expect_checkpoint_parity(1),
        }),
        test.case("revive_rejects_a_usable_but_non_revivable_corpse_before_payment", {
            test.submit_system(
                fixtures.command(
                    "system.player.enter",
                    fixtures.player_entry({ level = 10, mana = 100, max_mana = 100, level_id = 2 }),
                    { tick = 1, sequence = 1 }
                )
            ),
            test.step(1),
            test.run(function()
                learn()
                corpse("forbidden-body")
            end),
            test.submit(
                fixtures.command("player.assign_skills", { right = 95 }, { tick = 2, sequence = 1, player = "alice" })
            ),
            test.step(1),
            test.submit(
                fixtures.command(
                    "player.use_skill",
                    { side = "right", target_id = "monster:forbidden-body" },
                    { tick = 3, sequence = 2, player = "alice" }
                )
            ),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.vitals" } })[1]
                test.expect(ecs.get(player, "d2legacy.player.vitals"):get("mana_raw")):equals(100 * 256)
                test.expect(#ecs.query({ all = { "d2legacy.skill.cast" } })):equals(0)
                test.expect(#ecs.query({ all = { "d2legacy.owned_unit" } })):equals(0)
            end),
        }),
        test.case("revive_preserves_the_corpse_monster_type_and_adds_a_timed_owned_lifetime", {
            test.submit_system(
                fixtures.command(
                    "system.player.enter",
                    fixtures.player_entry({ level = 10, mana = 100, max_mana = 100, level_id = 2 }),
                    { tick = 1, sequence = 1 }
                )
            ),
            test.step(1),
            test.run(function()
                learn()
                corpse("revive-body", 2, true)
            end),
            test.submit(
                fixtures.command("player.assign_skills", { right = 95 }, { tick = 2, sequence = 1, player = "alice" })
            ),
            test.step(1),
            test.submit(
                fixtures.command(
                    "player.use_skill",
                    { side = "right", target_id = "monster:revive-body" },
                    { tick = 3, sequence = 2, player = "alice" }
                )
            ),
            test.step(3),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local summon = ecs.query({ all = { "d2legacy.owned_unit", "d2legacy.monster.stats" } })[1]
                test.assert(summon ~= nil, "revived monster was not materialized")
                local identity = ecs.get(summon, "d2legacy.monster.identity")
                local stats = ecs.get(summon, "d2legacy.monster.stats")
                local relation = ecs.get(summon, "d2legacy.owned_unit")
                local brain = ecs.get(summon, "d2legacy.monster.ai")
                local selected = ecs.get(summon, "d2legacy.world.selectable")
                test.expect(identity:get("definition_id")):equals("fixturecorpse")
                test.expect(selected:get("kind")):equals("friendly")
                test.expect(relation:get("category")):equals("revive")
                test.expect(relation:get("limit")):equals(20)
                test.expect(relation:get("expires_tick") - relation:get("created_tick")):equals(4500)
                test.expect(stats:get("max_health")):equals(31 * 256)
                test.expect(stats:get("physical_min")):equals(665)
                test.expect(stats:get("physical_max")):equals(1331)
                test.expect(brain:get("speed")):equals(15)
                test.assert(ecs.get(summon, "d2legacy.monster.revivable") == nil)
                test.expect(#ecs.query({ all = { "d2legacy.monster.death" } })):equals(0)
            end),
            test.expect_checkpoint_parity(1),
        }),
        test.case("valid_corpse_becomes_owned_friendly_monster_with_snapshotted_modifiers", {
            test.submit_system(fixtures.command(
                "system.player.enter",
                fixtures.player_entry({
                    level = 10,
                    mana = 100,
                    max_mana = 100,
                    level_id = 2,
                }),
                { tick = 1, sequence = 1 }
            )),
            test.step(1),
            test.run(function()
                learn()
                corpse("body")
            end),
            test.submit(
                fixtures.command("player.assign_skills", { right = 70 }, { tick = 2, sequence = 1, player = "alice" })
            ),
            test.step(1),
            test.submit(
                fixtures.command(
                    "player.use_skill",
                    { side = "right", target_id = "monster:body" },
                    { tick = 3, sequence = 2, player = "alice" }
                )
            ),
            test.step(3),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local summons = ecs.query({ all = { "d2legacy.owned_unit", "d2legacy.monster.stats" } })
                test.expect(#summons):equals(1)
                local summon = summons[1]
                local selected = ecs.get(summon, "d2legacy.world.selectable")
                local stats = ecs.get(summon, "d2legacy.monster.stats")
                local relation = ecs.get(summon, "d2legacy.owned_unit")
                local defense = ecs.get(summon, "d2legacy.combat.defense")
                test.expect(selected:get("kind")):equals("friendly")
                test.expect(relation:get("owner_id")):equals("player:alice")
                test.expect(relation:get("limit")):equals(8)
                test.expect(stats:get("max_health")):equals(223 * 256)
                test.expect(stats:get("attack_rating")):equals(550)
                test.expect(stats:get("defense")):equals(450)
                test.expect(defense:get("fire_resist")):equals(34)
                test.expect(#ecs.query({ all = { "d2legacy.monster.death" } })):equals(0)
                local events = ecs.query({ all = { "d2legacy.skill.summon_event" } })
                test.expect(ecs.get(events[1], "d2legacy.skill.summon_event"):get("outcome")):equals("summoned")
                local player = ecs.query({ all = { "d2legacy.player.vitals" } })[1]
                test.expect(ecs.get(player, "d2legacy.player.vitals"):get("mana_raw")):equals(75 * 256)
            end),
            test.expect_checkpoint_parity(2),
            test.run(function()
                alive_hostile()
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local summon = ecs.query({ all = { "d2legacy.owned_unit", "d2legacy.monster.ai" } })[1]
                local brain = ecs.get(summon, "d2legacy.monster.ai")
                test.expect(brain:get("target_id")):equals("monster:ai-target")
                test.expect(brain:get("state")):equals("chase")
            end),
            test.run(function()
                lethal_result_from_owned_summon()
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local victim
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.monster.death" } })) do
                    local identity = ecs.get(entity, "d2legacy.monster.identity")
                    if identity and identity:get("spawn_id") == "victim" then
                        victim = entity
                    end
                end
                test.assert(victim ~= nil, "owned summon kill did not reach monster death")
                test.expect(ecs.get(victim, "d2legacy.monster.death"):get("credited_id")):equals("player:alice")
                local player = ecs.query({ all = { "d2legacy.player.progress" } })[1]
                test.expect(ecs.get(player, "d2legacy.player.progress"):get("experience")):equals(10)
            end),
            test.submit_system(
                fixtures.command("system.player.leave", { player = "alice" }, { tick = 10, sequence = 1 })
            ),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(#ecs.query({ all = { "d2legacy.owned_unit" } })):equals(0)
                test.expect(#ecs.query({ all = { "d2legacy.player.identity" } })):equals(0)
            end),
        }),
    },
})
