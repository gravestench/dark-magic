local test = require("d2legacy.tests/v1")

local function melee_skill(id, selector)
    return {
        Id = tostring(id),
        skill = "fixture " .. id,
        srvstfunc = "1",
        srvdofunc = "1",
        cltstfunc = "1",
        cltdofunc = "1",
        range = "both",
        itypea1 = "weap",
        anim = "A1",
        UseAttackRate = "1",
        TargetableOnly = "1",
        SearchEnemyXY = "1",
        AttackNoMana = "1",
        leftskill = "1",
        interrupt = "1",
        general = "1",
        InGame = "1",
        mana = "0",
        minmana = "0",
        SrcDam = "128",
        weapsel = tostring(selector),
    }
end

return test.suite({
    name = "melee action skill definitions",
    profile = "module",
    tier = "fast",
    records = {
        ["data/global/excel/skills.txt"] = {
            melee_skill(0, 0),
            melee_skill(500, 1),
        },
    },
    cases = {
        test.case("decodes_multiple_exact_configurations_without_identity_branches", function(t)
            t:run(function()
                local definitions = require("d2legacy.data.melee_skills").load({ 0, 500 })
                test.expect(definitions[0].behavior):equals("action.melee")
                test.expect(definitions[0].mana_cost_raw):equals(0)
                test.expect(definitions[500].weapon_selection):equals(1)
            end)
        end),
    },
})
