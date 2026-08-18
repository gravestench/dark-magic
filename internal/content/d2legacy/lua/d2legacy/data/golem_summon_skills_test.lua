local test = require("d2legacy.tests/v1")

local function merge(base, values)
    local result = {}
    for key, value in pairs(base) do
        result[key] = value
    end
    for key, value in pairs(values) do
        result[key] = value
    end
    return result
end

local base = {
    charclass = "nec",
    srvdofunc = "56",
    pettype = "golem",
    petmax = "1",
    summode = "S1",
    range = "none",
    leftskill = "",
    InGame = "1",
    InTown = "1",
    mana = "15",
    lvlmana = "3",
    minmana = "1",
    manashift = "8",
    passivestat1 = "velocitypercent",
    passivecalc1 = "skill('Golem Mastery'.dm34)",
    passivecalc2 = "skill('Golem Mastery'.ln56)+skill('Clay Golem'.blvl)*skill('Clay Golem'.par8)",
    passivecalc3 = "(skill('FireGolem'.blvl)*skill('FireGolem'.par8))",
    passivecalc4 = "skill('IronGolem'.blvl)*skill('IronGolem'.par8)",
    calc1 = "skill('Golem Mastery'.ln12) + (skill('BloodGolem'.blvl)*skill('BloodGolem'.par8))",
    Param1 = "0",
    Param2 = "0",
    Param3 = "0",
    Param4 = "0",
    Param5 = "0",
    Param6 = "0",
    Param8 = "0",
}

local clay_life_formula = table.concat({
    "(100+(par1 * (lvl - 1)))*(100+skill('Golem Mastery'.ln12)",
    " + (skill('BloodGolem'.blvl)*skill('BloodGolem'.par8)))/100-100",
})

local clay = merge(base, {
    Id = "75",
    skill = "Clay Golem",
    skilldesc = "clay golem",
    summon = "ClayGolem",
    aurastat1 = "item_slow",
    aurastatcalc1 = "dm34",
    auraevent1 = "damagedinmelee",
    auraeventfunc1 = "27",
    passivecalc2 = "skill('Golem Mastery'.ln56) + (lvl*par8)",
    passivecalc3 = "par2 * (lvl - 1) + (skill('FireGolem'.blvl)*skill('FireGolem'.par8))",
    calc1 = clay_life_formula,
    Param1 = "35",
    Param2 = "35",
    Param3 = "0",
    Param4 = "75",
    Param8 = "20",
})

local blood = merge(base, {
    Id = "85",
    skill = "BloodGolem",
    skilldesc = "bloodgolem",
    summon = "BloodGolem",
    mana = "25",
    lvlmana = "4",
    auraevent1 = "domeleedamage",
    auraeventfunc1 = "23",
    auraevent2 = "damagedinmelee",
    auraeventfunc2 = "26",
    auraevent3 = "damagedbymissile",
    auraeventfunc3 = "26",
    passivecalc3 = "par4 * (lvl - 1) + (skill('FireGolem'.blvl)*skill('FireGolem'.par8))",
    Param1 = "75",
    Param2 = "150",
    Param3 = "30",
    Param4 = "35",
    Param5 = "0",
    Param6 = "25",
    Param8 = "5",
})

local iron = merge(base, {
    Id = "90",
    skill = "IronGolem",
    skilldesc = "irongolem",
    summon = "IronGolem",
    srvstfunc = "20",
    srvdofunc = "57",
    TargetItem = "1",
    stsuccessonly = "1",
    mana = "35",
    lvlmana = "0",
    aurastat1 = "thorns_percent",
    aurastatcalc1 = "ln12",
    aurastat2 = "fade",
    aurastatcalc2 = "16",
    passivecalc3 = "lvl*par8",
    passivecalc4 = "(skill('FireGolem'.blvl)*skill('FireGolem'.par8))",
    Param1 = "150",
    Param2 = "15",
    Param8 = "35",
})

local fire = merge(base, {
    Id = "94",
    skill = "FireGolem",
    skilldesc = "firegolem",
    summon = "FireGolem",
    mana = "50",
    lvlmana = "10",
    aurastat1 = "fireresist",
    aurastatcalc1 = "100 - dm12",
    aurastat2 = "item_absorbfire_percent",
    aurastatcalc2 = "dm12",
    aurastat3 = "firemindam",
    aurastatcalc3 = "edmn",
    aurastat4 = "firemaxdam",
    aurastatcalc4 = "edmx",
    sumskill1 = "holy fire",
    sumsk1calc = "min(ln56,30)",
    Param1 = "25",
    Param2 = "100",
    Param3 = "100",
    Param4 = "35",
    Param5 = "8",
    Param6 = "1",
    Param8 = "6",
    EMin = "10",
    EMax = "27",
    EMinLev1 = "9",
    EMinLev2 = "10",
    EMinLev3 = "11",
    EMinLev4 = "12",
    EMinLev5 = "13",
    EMaxLev1 = "10",
    EMaxLev2 = "11",
    EMaxLev3 = "12",
    EMaxLev4 = "13",
    EMaxLev5 = "14",
})

local function description(name, modifiers)
    local row = { skilldesc = name, dsc3texta1 = "Sksyn", dsc3texta2 = "mastery", dsc3texta3 = "resist" }
    for index = 1, 3 do
        row["dsc3texta" .. (index + 3)] = modifiers[index]
    end
    return row
end

return test.suite({
    name = "complete golem summon definitions",
    profile = "module",
    tier = "fast",
    records = {
        ["data/global/excel/skills.txt"] = {
            clay,
            {
                Id = "79",
                skill = "Golem Mastery",
                Param1 = "20",
                Param2 = "20",
                Param3 = "0",
                Param4 = "40",
                Param5 = "25",
                Param6 = "25",
            },
            blood,
            {
                Id = "89",
                skill = "Summon Resist",
                passivestat1 = "passive_summon_resist",
                passivecalc1 = "dm12",
                Param1 = "20",
                Param2 = "75",
            },
            iron,
            fire,
            {
                Id = "102",
                skill = "holy fire",
                srvdofunc = "66",
                aura = "1",
                periodic = "",
                aurafilter = "42883",
                EType = "fire",
                Param1 = "6",
                Param2 = "1",
                perdelay = "50",
                EMin = "2",
                EMax = "6",
                EMinLev1 = "1",
                EMinLev2 = "2",
                EMinLev3 = "3",
                EMinLev4 = "5",
                EMinLev5 = "7",
                EMaxLev1 = "1",
                EMaxLev2 = "2",
                EMaxLev3 = "3",
                EMaxLev4 = "5",
                EMaxLev5 = "7",
            },
        },
        ["data/global/excel/PetType.txt"] = {
            { ["pet type"] = "golem", group = "0", basemax = "0", unsummon = "1", warp = "1", range = "" },
        },
        ["data/global/excel/SkillDesc.txt"] = {
            description("clay golem", { "blood", "iron", "fire" }),
            description("bloodgolem", { "clay", "iron", "fire" }),
            description("irongolem", { "clay", "blood", "fire" }),
            description("firegolem", { "clay", "blood", "iron" }),
        },
    },
    cases = {
        test.case("decodes_all_four_members_and_their_cross_family_effects_together", function(t)
            t:run(function()
                local definitions = require("d2legacy.data.golem_summon_skills").load({ 75, 85, 90, 94 })
                test.expect(definitions[75].slow_maximum):equals(75)
                test.expect(definitions[75].clay_attack_per_hard_level):equals(20)
                test.expect(definitions[85].life_steal_minimum):equals(75)
                test.expect(definitions[85].owner_share_percent):equals(30)
                test.assert(definitions[90].requires_item_target and definitions[90].success_only)
                test.expect(definitions[90].thorns_base):equals(150)
                test.expect(definitions[90].iron_defense_per_hard_level):equals(35)
                test.expect(definitions[90].defense_level_source):equals("effective_self")
                test.expect(definitions[75].defense_level_source):equals("hard_modifier")
                test.expect(definitions[94].fire_absorb_maximum):equals(100)
                test.expect(definitions[94].granted_skill_name):equals("holy fire")
                test.expect(definitions[94].granted_aura_period_ticks):equals(50)
                test.expect(definitions[94].granted_aura_radius_base):equals(6)
                test.expect(definitions[94].fire_damage_per_hard_level):equals(6)
                for _, id in ipairs({ 75, 85, 90, 94 }) do
                    local definition = definitions[id]
                    test.expect(definition.behavior):equals("summon.golem")
                    test.expect(definition.category):equals("golem")
                    test.expect(definition.mastery_skill_id):equals(79)
                    test.expect(definition.summon_resist_skill_id):equals(89)
                    test.expect(require("d2legacy.policy.summon").pet_limit(definition, 20)):equals(1)
                end
            end)
        end),
    },
})
