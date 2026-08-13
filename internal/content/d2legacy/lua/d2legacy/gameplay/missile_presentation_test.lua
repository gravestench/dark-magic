local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("resolves_a_copied_recipe", {
            test.run(function()
                local adapter = require("d2legacy.gameplay.missile_presentation")
                local recipe = adapter.resolve({
                    dcc = "data/global/missiles/firebolt.dcc",
                    palette = "",
                    velocity_x = 1,
                    velocity_y = 0,
                    directions = 8,
                    frames_per_second = 0,
                    loop = true,
                })
                test.assert(
                    recipe.path == "data/global/missiles/firebolt.dcc",
                    [=[recipe.path == "data/global/missiles/firebolt.dcc"]=]
                )
                test.assert(
                    recipe.palette == "data/global/palette/units/pal.dat",
                    [=[recipe.palette == "data/global/palette/units/pal.dat"]=]
                )
                test.assert(
                    recipe.direction == 3 and recipe.frames_per_second == 25 and recipe.loop == "loop",
                    [=[recipe.direction == 3 and recipe.frames_per_second == 25 and recipe.loop == "loop"]=]
                )
                test.assert(
                    adapter.resolve({ dcc = "", velocity_x = 0, velocity_y = 0, directions = 1 }) == nil,
                    [=[adapter.resolve({ dcc = "", velocity_x = 0, velocity_y = 0, directions = 1 }) == nil]=]
                )
            end),
        }),
    },
})
