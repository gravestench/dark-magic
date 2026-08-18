local test = require("d2legacy.tests/v1")

return test.suite({
    name = "record-driven cast presentation",
    profile = "module",
    tier = "fast",
    records = {
        ["data/global/excel/skills.txt"] = {
            {
                Id = "47",
                stsound = "sorceress_cast_fire",
                dosound = "fire_release",
                castoverlay = "fire_cast_2",
            },
        },
        ["data/global/excel/Overlay.txt"] = {
            {
                overlay = "fire_cast_2",
                Filename = "FireCast2",
                Frames = "16",
                AnimRate = "16",
                NumDirections = "1",
                Trans = "3",
            },
        },
    },
    cases = {
        test.case("follows_skill_overlay_and_sound_records", function(t)
            t:run(function()
                local recipe = require("d2legacy.gameplay.cast_presentation").resolve(47)
                test.expect(recipe.start_sound):equals("sorceress_cast_fire")
                test.expect(recipe.effect_sound):equals("fire_release")
                test.expect(recipe.overlay.path):equals("data/global/overlays/FireCast2.dcc")
                test.expect(recipe.overlay.frames):equals(16)
            end)
        end),
    },
})
