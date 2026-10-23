local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local function submit(tick, kind, payload)
    return test.submit({ tick = tick, sequence = tick - 1, player = "alice", kind = kind, payload = payload or {} })
end

local function player()
    return require("engine.ecs/v1").query({ all = { "d2legacy.player.identity" } })[1]
end

local function pending()
    return test.run(function()
        require("engine.ecs/v1").set(player(), "d2legacy.quest.progress", {
            quest_id = 1,
            stage = 5,
            completed = true,
            rewarded = false,
        })
    end)
end

local cases = {}
for _, scenario in ipairs({ "early", "wrong_npc", "closed", "range", "level", "inactive", "dead", "other_quest" }) do
    cases[#cases + 1] = test.case("refuses_" .. scenario, {
        test.submit_system({ tick = 1, sequence = 1, kind = "system.player.enter", payload = fixtures.player_entry() }),
        test.step(1),
        submit(2, "interaction.open", { target = scenario == "wrong_npc" and "npc:kashya" or "npc:akara" }),
        test.step(1),
        pending(),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            if scenario == "early" then
                ecs.get(player(), "d2legacy.quest.progress"):set("completed", false)
            elseif scenario == "other_quest" then
                ecs.get(player(), "d2legacy.quest.progress"):set("quest_id", 4)
            elseif scenario == "range" then
                ecs.get(player(), "d2legacy.world.position"):set("x", 90)
            elseif scenario == "level" then
                ecs.get(player(), "d2legacy.world.location"):set("level_id", 8)
            elseif scenario == "inactive" then
                local context = require("d2legacy.interactions.context").for_owner("alice")
                ecs.set(context:get("target"), "d2legacy.world.inactive", {})
            elseif scenario == "dead" then
                ecs.get(player(), "d2legacy.player.vitals"):set("health", 0)
            elseif scenario == "closed" then
                require("d2legacy.commands.interactions").close({ player = "alice", payload = {} })
            end
        end),
        submit(3, "quest.claim_reward"),
        test.step(1),
        submit(4, "quest.claim_reward"),
        test.step(1),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            test.expect(ecs.get(player(), "d2legacy.player.progress"):get("unspent_skill_points")):equals(0)
            test.expect(ecs.get(player(), "d2legacy.quest.progress"):get("rewarded")):is_false()
            test.expect(#ecs.query({ all = { "d2legacy.quest.event" } })):equals(0)
        end),
    })
end

return test.suite({
    profile = "authority",
    tier = "integration",
    initial_data = {
        ["d2legacy.interactions"] = {
            owner = "alice",
            targets = {
                {
                    id = "npc:akara",
                    npc = "akara",
                    x = 10,
                    y = 12,
                    radius = 3,
                    act = 1,
                    level_id = 1,
                    selectable = true,
                },
                {
                    id = "npc:kashya",
                    npc = "kashya",
                    x = 10,
                    y = 12,
                    radius = 3,
                    act = 1,
                    level_id = 1,
                    selectable = true,
                },
            },
        },
    },
    cases = cases,
})
