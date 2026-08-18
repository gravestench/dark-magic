local test = require("d2legacy.tests/v1")

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

local mage = {
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
}

local revive = {
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
}

return test.suite({
    name = "targeted corpse summon definitions",
    profile = "module",
    tier = "fast",
    records = {
        ["data/global/excel/skills.txt"] = {
            { Id = "69", skill = "Fixture Mastery", Param1 = "8", Param2 = "2", Param3 = "5", Param4 = "10" },
            raise,
            mage,
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
            revive,
        },
        ["data/global/excel/PetType.txt"] = {
            { ["pet type"] = "skeleton", group = "0", basemax = "0", unsummon = "1", warp = "1", range = "" },
            { ["pet type"] = "skeletonmage", group = "0", basemax = "0", unsummon = "1", warp = "1" },
            { ["pet type"] = "revive", group = "0", basemax = "0", unsummon = "1", warp = "1" },
        },
    },
    cases = {
        test.case("joins_pet_category_and_modifier_skills_without_an_id_branch", function(t)
            t:run(function()
                local definitions = require("d2legacy.data.corpse_summon_skills").load({ 70, 80, 95 })
                local definition = definitions[70]
                test.expect(definition.behavior):equals("summon.targeted-corpse")
                test.expect(definition.summon_monster_id):equals("fixturepet")
                test.expect(definition.category):equals("skeleton")
                test.assert(definition.requires_corpse_target)
                test.assert(definition.unsummon and definition.warp_with_owner)
                test.expect(definition.mana_cost_raw):equals(6 * 256)
                test.expect(definition.mastery_skill_id):equals(69)
                test.expect(definition.mastery_life_flat_per_level):equals(8)
                test.expect(definition.mastery_damage_flat_per_level):equals(2)
                test.expect(definition.summon_resist.skill_id):equals(89)
                test.expect(definition.summon_resist.minimum):equals(20)
                test.expect(definition.summon_resist.maximum):equals(75)
                local progression = require("d2legacy.policy.skill_progression")
                local summon = require("d2legacy.policy.summon")
                local limits = { 1, 2, 3, 3, 3, 4, 4, 4, 5, 5, 5, 6, 6, 6, 7, 7, 7, 8, 8, 8 }
                for level = 1, 20 do
                    test.expect(progression.mana_cost(definition, level)):equals((5 + level) * 256)
                    test.expect(summon.pet_limit(definition, level)):equals(limits[level])
                end
                test.expect(summon.pet_level(1, 1)):equals(1)
                test.expect(summon.pet_level(1, 20)):equals(16)
                test.expect(summon.pet_level(20, 10)):equals(10)
                local mage_definition = definitions[80]
                test.expect(mage_definition.materialization):equals("fixed_pet")
                test.expect(mage_definition.granted_skill_name):equals("FixtureMageMissile")
                test.expect(summon.granted_skill_level(mage_definition, 20, 3)):equals(12)
                test.expect(mage_definition.defense_per_combined_level):equals(10)
                test.expect(mage_definition.damage_percent_per_level_after_three):equals(0)
                local revive_definition = definitions[95]
                test.expect(revive_definition.materialization):equals("revived_corpse")
                test.assert(revive_definition.requires_revivable_corpse)
                test.expect(summon.pet_limit(revive_definition, 20)):equals(20)
                test.expect(summon.duration_ticks(revive_definition, 1)):equals(4500)
                test.expect(revive_definition.life_percent_base):equals(200)
                test.expect(revive_definition.mastery_life_percent_per_level):equals(5)
                test.expect(revive_definition.mastery_damage_percent_per_level):equals(10)
                test.expect(revive_definition.velocity_percent):equals(50)
            end)
        end),
    },
})
