local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "policy",
    tier = "fast",
    tests = {
        hit_chance_preserves_recovered_integer_order = {
            {
                run = function()
                    local melee = require("d2legacy.policy.melee")
                    local direction = require("d2legacy.policy.direction")
                    assert(direction.player(0, 1) == 0 and direction.player(-1, 0) == 1)
                    assert(direction.player(0, -1) == 2 and direction.player(1, 0) == 3)
                    assert(direction.player(1, 1) == 4 and direction.player(-1, 1) == 5)
                    assert(direction.player(-1, -1) == 6 and direction.player(1, -1) == 7)
                    assert(direction.player(0.4, 1) == 8 and direction.player(-0.4, 1) == 9)
                    assert(direction.player(-1, 0.4) == 10 and direction.player(-1, -0.4) == 11)
                    assert(direction.player(-0.4, -1) == 12 and direction.player(0.4, -1) == 13)
                    assert(direction.player(1, -0.4) == 14 and direction.player(1, 0.4) == 15)
                    assert(melee.reach(1, 1, 1) == 3 and melee.reach(2, 0.75, 1.25) == 4)
                    assert(melee.hit_chance(10, 10, 100, 100) == 50)
                    assert(melee.hit_chance(7, 11, 101, 100) == 38)
                    assert(melee.hit_chance(1, 99, 1, 10000) == 5)
                    assert(melee.hit_chance(99, 1, 10000, 1) == 95)
                    assert(melee.hit_chance(10, 10, 0, 0) == 95)
                    assert(melee.hit_chance(10, 10, 50, -50) == 95)
                    assert(melee.hit_chance(10, 10, -50, 50) == 5)
                end,
            },
        },
    },
})
