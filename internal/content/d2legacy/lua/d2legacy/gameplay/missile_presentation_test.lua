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
                    transparency_mode = 1,
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
                    recipe.direction == 3
                        and recipe.frames_per_second == 25
                        and recipe.loop == "loop"
                        and recipe.blend == "screen",
                    [=[the copied recipe includes direction, timing, looping, and record-authored luminosity]=]
                )
                test.assert(
                    adapter.direction(1, 0, 4) == 0
                        and adapter.direction(1, 0, 8) == 3
                        and adapter.direction(1, 0, 16) == 3
                        and adapter.direction(1, 0, 32) == 3,
                    [=[east-facing velocity resolves through every owned DCC direction count]=]
                )
                test.assert(
                    adapter.direction(math.cos(math.pi / 8), math.sin(math.pi / 8), 16) == 15,
                    [=[16-way missiles preserve the half-angle between east and southeast]=]
                )
                test.assert(
                    adapter.direction(math.cos(math.pi / 16), math.sin(math.pi / 16), 32) == 30,
                    [=[32-way missiles preserve the quarter-angle adjacent to east]=]
                )
                test.expect(adapter.direction(1, 1, 16), "southeast world diagonal"):equals(4)
                test.expect(adapter.direction(0, 1, 16), "positive Y world axis"):equals(0)
                test.expect(adapter.direction(-1, 0, 16), "negative X world axis"):equals(1)
                test.expect(adapter.direction(0, -1, 16), "negative Y world axis"):equals(2)
                test.assert(
                    adapter.resolve({ dcc = "", velocity_x = 0, velocity_y = 0, directions = 1 }) == nil,
                    [=[adapter.resolve({ dcc = "", velocity_x = 0, velocity_y = 0, directions = 1 }) == nil]=]
                )
            end),
        }),
    },
})
