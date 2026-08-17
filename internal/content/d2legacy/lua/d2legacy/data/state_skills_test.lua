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
        mana = "7",
        manashift = "8",
        Param1 = "30",
        Param2 = "5",
        Param3 = "3000",
        Param4 = "300",
        Param7 = "250",
    }
end

return test.suite({
    name = "timed self-state skill definitions",
    profile = "module",
    tier = "fast",
    records = {
        ["data/global/excel/skills.txt"] = {
            skill(40, "Fixture Armor", "Synergy One", "Synergy Two"),
            skill(900, "Second Fixture Armor", "Synergy One", "Synergy Two"),
            { Id = "50", skill = "Synergy One" },
            { Id = "60", skill = "Synergy Two" },
        },
    },
    cases = {
        test.case("decodes_multiple_configurations_without_skill_branches", function(t)
            t:run(function()
                local definitions = require("d2legacy.data.state_skills").load({ 40, 900 })
                test.expect(definitions[40].behavior):equals("state.self-timed-stat")
                test.expect(definitions[40].state_id):equals("fixturearmor40")
                test.expect(definitions[40].mana_cost_raw):equals(1792)
                test.expect(definitions[40].duration_base):equals(3000)
                test.expect(definitions[40].duration_synergy_skill_ids[1]):equals(50)
                test.expect(definitions[900].state_id):equals("fixturearmor900")
                test.expect(definitions[900].stat):equals("defense")
            end)
        end),
    },
})
