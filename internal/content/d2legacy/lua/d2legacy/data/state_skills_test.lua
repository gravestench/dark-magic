local test = require("d2legacy.tests/v1")

local function skill(id, name, synergy_a, synergy_b)
    return {
        Id = tostring(id),
        skill = name,
        srvstfunc = "",
        srvdofunc = "18",
        aurastate = "fixturearmor" .. tostring(id),
        auralencalc = "ln34+(skill('" .. synergy_a .. "'.blvl)+skill('" .. synergy_b .. "'.blvl))*par7",
        aurastat1 = "skill_armor_percent",
        aurastatcalc1 = "ln12",
        auraevent1 = "damagedinmelee",
        auraeventfunc1 = "2",
        calc1 = "ln56*(100+((skill('" .. synergy_a .. "'.blvl)+skill('" .. synergy_b .. "'.blvl))*par8))/100",
        mana = "7",
        manashift = "8",
        Param1 = "30",
        Param2 = "5",
        Param3 = "3000",
        Param4 = "300",
        Param5 = "30",
        Param6 = "3",
        Param7 = "250",
        Param8 = "5",
    }
end

local function cold_reaction(id, name, event, event_function, synergy_a, synergy_b, missile)
    return {
        Id = tostring(id),
        skill = name,
        srvstfunc = "",
        srvdofunc = "18",
        srvmissilea = missile or "",
        aurastate = "fixturearmor" .. tostring(id),
        auralencalc = "ln34+(skill('" .. synergy_a .. "'.blvl)+skill('" .. synergy_b .. "'.blvl))*par7",
        aurastat1 = "skill_armor_percent",
        aurastatcalc1 = "ln12",
        auraevent1 = event,
        auraeventfunc1 = event_function,
        mana = "11",
        manashift = "8",
        Param1 = "45",
        Param2 = "6",
        Param3 = "3000",
        Param4 = "300",
        Param7 = "250",
        Param8 = "9",
        EType = "cold",
        HitShift = "7",
        EMin = "12",
        EMax = "16",
        EMinLev1 = "4",
        EMinLev2 = "6",
        EMinLev3 = "8",
        EMinLev4 = "10",
        EMinLev5 = "12",
        EMaxLev1 = "5",
        EMaxLev2 = "7",
        EMaxLev3 = "9",
        EMaxLev4 = "11",
        EMaxLev5 = "13",
        EDmgSymPerCalc = "(skill('" .. synergy_a .. "'.blvl)+skill('" .. synergy_b .. "'.blvl))*par8",
        ELen = "100",
        ELevLen1 = "0",
        ELevLen2 = "25",
        ELevLen3 = "50",
        ELevLen4 = "0",
        ELevLen5 = "0",
    }
end

return test.suite({
    name = "timed self-state skill definitions",
    profile = "module",
    tier = "fast",
    records = {
        ["data/global/excel/skills.txt"] = {
            skill(40, "Fixture Armor", "Shiver Armor", "Chilling Armor"),
            skill(900, "Second Fixture Armor", "Shiver Armor", "Chilling Armor"),
            cold_reaction(50, "Shiver Armor", "attackedinmelee", "3", "Fixture Armor", "Chilling Armor"),
            cold_reaction(
                60,
                "Chilling Armor",
                "hitbymissile",
                "1",
                "Fixture Armor",
                "Shiver Armor",
                "chillingarmorbolt"
            ),
        },
        ["data/global/excel/states.txt"] = {
            { state = "fixturearmor40", id = "10", group = "1" },
            { state = "fixturearmor900", id = "11", group = "1" },
            { state = "fixturearmor50", id = "12", group = "1" },
            { state = "fixturearmor60", id = "13", group = "1" },
            { state = "freeze", id = "1", group = "0" },
            { state = "cold", id = "14", group = "0" },
        },
        ["data/global/excel/Missiles.txt"] = {
            {
                Missile = "chillingarmorbolt",
                Skill = "Chilling Armor",
                pSrvDoFunc = "1",
                CollideType = "3",
                CollideKill = "1",
                Vel = "18",
                Range = "25",
                Size = "1",
                CelFile = "IceBolt",
                NumDirections = "16",
                AnimSpeed = "16",
                LoopAnim = "1",
                Trans = "1",
                TravelSound = "sorceress_icebolt_1",
                HitSound = "impact_cold_1",
            },
        },
    },
    cases = {
        test.case("decodes_multiple_configurations_without_skill_branches", function(t)
            t:run(function()
                local definitions = require("d2legacy.data.state_skills").load({ 40, 50, 60, 900 })
                test.expect(definitions[40].behavior):equals("state.self-timed-reactive")
                test.expect(definitions[40].state_id):equals("fixturearmor40")
                test.expect(definitions[40].mana_cost_raw):equals(1792)
                test.expect(definitions[40].duration_base):equals(3000)
                test.expect(definitions[40].duration_synergy_skill_ids[1]):equals(50)
                test.expect(definitions[40].exclusive_group):equals("state-group:1")
                test.expect(definitions[40].reaction):equals("melee_damage_freeze")
                test.expect(definitions[40].reaction_state_id):equals("freeze")
                test.expect(definitions[40].reaction_chill_state_id):equals("cold")
                test.expect(definitions[40].reaction_duration_base):equals(30)
                test.expect(definitions[40].reaction_duration_synergy_skill_ids[2]):equals(60)
                test.expect(definitions[900].state_id):equals("fixturearmor900")
                test.expect(definitions[900].stat):equals("defense")
                test.expect(definitions[50].reaction):equals("melee_attack_cold")
                test.expect(definitions[50].minimum_damage_raw):equals(1536)
                test.expect(definitions[50].minimum_damage_per_level_raw[3]):equals(1024)
                test.expect(definitions[50].damage_synergy_skill_ids[1]):equals(40)
                test.expect(definitions[50].reaction_duration_per_level[2]):equals(25)
                test.expect(definitions[60].reaction):equals("missile_hit_return")
                test.expect(definitions[60].reaction_missile.missile_id):equals("chillingarmorbolt")
                test.expect(definitions[60].reaction_missile.directions):equals(16)
                test.expect(definitions[60].reaction_missile.transparency_mode):equals(1)
            end)
        end),
    },
})
