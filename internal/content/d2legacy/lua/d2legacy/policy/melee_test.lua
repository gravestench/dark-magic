local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    covers = { "internal/game/stats" },
    cases = {
        test.case("hit_chance_preserves_recovered_integer_order", {
            test.run(function()
                local melee = require("d2legacy.policy.melee")
                local direction = require("d2legacy.policy.direction")
                test.assert(
                    direction.player(0, 1) == 0 and direction.player(-1, 0) == 1,
                    [=[direction.player(0, 1) == 0 and direction.player(-1, 0) == 1]=]
                )
                test.assert(
                    direction.player(0, -1) == 2 and direction.player(1, 0) == 3,
                    [=[direction.player(0, -1) == 2 and direction.player(1, 0) == 3]=]
                )
                test.assert(
                    direction.player(1, 1) == 4 and direction.player(-1, 1) == 5,
                    [=[direction.player(1, 1) == 4 and direction.player(-1, 1) == 5]=]
                )
                test.assert(
                    direction.player(-1, -1) == 6 and direction.player(1, -1) == 7,
                    [=[direction.player(-1, -1) == 6 and direction.player(1, -1) == 7]=]
                )
                test.assert(
                    direction.player(0.4, 1) == 8 and direction.player(-0.4, 1) == 9,
                    [=[direction.player(0.4, 1) == 8 and direction.player(-0.4, 1) == 9]=]
                )
                test.assert(
                    direction.player(-1, 0.4) == 10 and direction.player(-1, -0.4) == 11,
                    [=[direction.player(-1, 0.4) == 10 and direction.player(-1, -0.4) == 11]=]
                )
                test.assert(
                    direction.player(-0.4, -1) == 12 and direction.player(0.4, -1) == 13,
                    [=[direction.player(-0.4, -1) == 12 and direction.player(0.4, -1) == 13]=]
                )
                test.assert(
                    direction.player(1, -0.4) == 14 and direction.player(1, 0.4) == 15,
                    [=[direction.player(1, -0.4) == 14 and direction.player(1, 0.4) == 15]=]
                )
                test.assert(
                    melee.reach(1, 1, 1) == 3 and melee.reach(2, 0.75, 1.25) == 4,
                    [=[melee.reach(1, 1, 1) == 3 and melee.reach(2, 0.75, 1.25) == 4]=]
                )
                test.assert(melee.hit_chance(10, 10, 100, 100) == 50, [=[melee.hit_chance(10, 10, 100, 100) == 50]=])
                test.assert(melee.hit_chance(7, 11, 101, 100) == 38, [=[melee.hit_chance(7, 11, 101, 100) == 38]=])
                test.assert(melee.hit_chance(1, 99, 1, 10000) == 5, [=[melee.hit_chance(1, 99, 1, 10000) == 5]=])
                test.assert(melee.hit_chance(99, 1, 10000, 1) == 95, [=[melee.hit_chance(99, 1, 10000, 1) == 95]=])
                test.assert(melee.hit_chance(10, 10, 0, 0) == 95, [=[melee.hit_chance(10, 10, 0, 0) == 95]=])
                test.assert(melee.hit_chance(10, 10, 50, -50) == 95, [=[melee.hit_chance(10, 10, 50, -50) == 95]=])
                test.assert(melee.hit_chance(10, 10, -50, 50) == 5, [=[melee.hit_chance(10, 10, -50, 50) == 5]=])
            end),
        }),
    },
})
