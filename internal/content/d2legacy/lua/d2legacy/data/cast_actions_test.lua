local test = require("d2legacy.tests/v1")

return test.suite({
    name = "shared cast action definitions",
    profile = "module",
    tier = "fast",
    records = {
        ["data/global/excel/skills.txt"] = {
            {
                Id = "47",
                anim = "SC",
                seqtrans = "SC",
                seqnum = "",
                stsound = "sorceress_cast_fire",
                dosound = "",
                castoverlay = "fire_cast_2",
                cltmissile = "fireball",
            },
            { Id = "0", anim = "A1", seqtrans = "A1", seqnum = "0", UseAttackRate = "1" },
        },
    },
    cases = {
        test.case("joins_action_and_presentation_fields_without_skill_branches", function(t)
            t:run(function()
                local actions = require("d2legacy.data.cast_actions").load({
                    [47] = { behavior = "missile.straight-impact-area" },
                    [0] = { behavior = "action.melee" },
                })
                test.expect(actions[47].animation_mode):equals("SC")
                test.expect(actions[47].sequence_transition):equals("SC")
                test.expect(actions[47].start_sound):equals("sorceress_cast_fire")
                test.expect(actions[47].cast_overlay):equals("fire_cast_2")
                test.expect(actions[47].client_missile):equals("fireball")
                test.expect(actions[0].animation_mode):equals("A1")
                test.assert(actions[0].use_attack_rate)
            end)
        end),
    },
})
