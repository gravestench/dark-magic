local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    records = {
        ["data/global/excel/skills.txt"] = {
            {
                Id = "0",
                skill = "Attack",
                skilldesc = "attack",
                general = "1",
                passive = "0",
                leftskill = "1",
            },
            {
                Id = "36",
                skill = "Fire Bolt",
                skilldesc = "firebolt",
                general = "0",
                passive = "0",
                leftskill = "1",
            },
            {
                Id = "37",
                skill = "Warmth",
                skilldesc = "warmth",
                general = "0",
                passive = "1",
                leftskill = "0",
            },
        },
        ["data/global/excel/skilldesc.txt"] = {
            { skilldesc = "attack", ListRow = "0", IconCel = "0" },
            { skilldesc = "firebolt", ListRow = "1", IconCel = "1" },
            { skilldesc = "warmth", ListRow = "2", IconCel = "2" },
        },
        ["data/global/excel/charstats.txt"] = {
            {
                class = "Sorceress",
                StartSkill = "Fire Bolt",
                WalkVelocity = "6",
                RunVelocity = "9",
                stamina = "74",
                vit = "10",
                RunDrain = "20",
                StaminaPerLevel = "4",
                StaminaPerVitality = "4",
            },
        },
    },
    cases = {
        test.case("resolves_explicit_fixture_ids_through_owned_skill_records", {
            test.run(function()
                local skill = require("d2legacy.data.skill")
                local learned = skill.learned_for_ids({ 36, 0 }, 20)
                test.expect(#learned):equals(2)
                test.expect(learned[1].id):equals(0)
                test.expect(learned[1].level):equals(20)
                test.assert(learned[1].left_allowed and learned[1].right_allowed)
                test.expect(learned[2].id):equals(36)
                test.expect(learned[2].level):equals(20)
                test.assert(learned[2].left_allowed and learned[2].right_allowed)
            end),
        }),
    },
})
