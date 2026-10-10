local test = require("d2legacy.tests/v1")

local fixtures = require("d2legacy.tests.support.fixtures")

-- Enter alice into the Hell game with the given character progression.
local function enter(highest_completed)
    local entry = fixtures.player_entry({ highest_completed_difficulty = highest_completed })
    return pcall(require("d2legacy.commands.enter_player").apply, { tick = 1, payload = entry })
end

local function players()
    local ecs = require("engine.ecs/v1")
    return ecs.query({ all = { "d2legacy.player.identity" } })
end

return test.suite({
    profile = "authority",
    tier = "fast",
    initial_data = {
        ["d2legacy.game_rules"] = {
            target = "lod-1.14d",
            expansion = true,
            difficulty = 2,
        },
    },
    cases = {
        test.case("a_character_that_completed_nightmare_may_join_hell", {
            test.run(function()
                local ok, err = enter(1)
                test.assert(ok, tostring(err))
                local difficulty = require("engine.ecs/v1").get(players()[1], "d2legacy.player.difficulty")
                test.expect(difficulty:get("highest_completed")):equals(1)
            end),
        }),
        test.case("a_character_that_only_completed_normal_cannot_join_hell", {
            test.run(function()
                local ok, err = enter(0)
                test.assert(not ok, "entry into a locked difficulty was admitted")
                test.assert(
                    string.find(tostring(err), "has not unlocked", 1, true) ~= nil,
                    "unexpected rejection: " .. tostring(err)
                )
                test.expect(#players()):equals(0)
            end),
        }),
        -- Hosts that do not yet supply durable progression are not blocked.
        test.case("unknown_progression_is_not_checked", {
            test.run(function()
                local ok, err = enter(nil)
                test.assert(ok, tostring(err))
                local difficulty = require("engine.ecs/v1").get(players()[1], "d2legacy.player.difficulty")
                test.expect(difficulty:get("highest_completed")):equals(-1)
            end),
        }),
    },
})
