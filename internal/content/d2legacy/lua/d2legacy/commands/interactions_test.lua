local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local function command(tick, kind, payload, player, sequence)
    return test.submit({
        tick = tick,
        sequence = sequence or tick - 1,
        player = player or "alice",
        kind = kind,
        payload = payload,
    })
end

local function choose(tick, choice, revision, target, player, sequence)
    return command(tick, "interaction.choose", {
        target = target or "npc:akara",
        choice = choice,
        revision = revision,
    }, player, sequence)
end

local function context(owner)
    return assert(require("d2legacy.interactions.context").for_owner(owner or "alice"))
end

local function player()
    local ecs = require("engine.ecs/v1")
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.player.identity" } })) do
        if ecs.get(entity, "d2legacy.player.identity"):get("player") == "alice" then
            return entity
        end
    end
end

local function enter()
    return test.submit_system({
        tick = 1,
        sequence = 1,
        kind = "system.player.enter",
        payload = fixtures.player_entry(),
    })
end

return test.suite({
    profile = "authority",
    tier = "fast",
    initial_data = {
        ["d2legacy.interactions"] = {
            owner = "alice",
            targets = {
                {
                    id = "npc:akara",
                    npc = "Akara",
                    x = 12,
                    y = 12,
                    radius = 4,
                    act = 1,
                    level_id = 1,
                    selectable = true,
                },
                { id = "npc:other", npc = "Other", x = 12, y = 12, radius = 4 },
            },
        },
    },
    cases = {
        test.case("dialogue_choices_restore_and_close_reopen_without_retaining_text", {
            enter(),
            test.step(1),
            command(2, "interaction.open", { target = "npc:akara" }),
            test.step(1),
            choose(3, "talk", 1),
            test.step(1),
            test.run(function()
                test.expect(context():get("dialogue_page")):equals("topics")
                test.expect(context():get("dialogue_revision")):equals(2)
            end),
            test.restore_checkpoint(),
            choose(4, "gossip", 2, nil, nil, 1),
            test.step(1),
            test.run(function()
                test.expect(context():get("dialogue_page")):equals("text")
                test.expect(context():get("dialogue_line")):equals("akara_act1_gossip_8")
            end),
            choose(5, "back", 3, nil, nil, 2),
            test.step(1),
            test.run(function()
                test.expect(context():get("dialogue_page")):equals("menu")
                test.expect(context():get("dialogue_line")):equals("")
            end),
            choose(6, "cancel", 4, nil, nil, 3),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(ecs.get(context():get("target"), "d2legacy.interaction.null_target") ~= nil):is_true()
            end),
            command(7, "interaction.open", { target = "npc:akara" }, nil, 4),
            test.step(1),
            test.run(function()
                test.expect(context():get("dialogue_page")):equals("menu")
                test.expect(context():get("dialogue_line")):equals("")
                test.expect(context():get("dialogue_revision")):equals(6)
                test.expect(#require("engine.ecs/v1").query({ all = { "d2legacy.interaction.null_target" } })):equals(0)
            end),
        }),
        test.case("stale_wrong_target_unavailable_and_other_owner_choices_are_no_ops", {
            enter(),
            test.submit_system({
                tick = 1,
                sequence = 2,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ player = "bob", character_id = "bob" }),
            }),
            test.step(1),
            command(2, "interaction.open", { target = "npc:akara" }),
            test.step(1),
            choose(3, "gossip", 1),
            test.step(1),
            choose(4, "talk", 0),
            test.step(1),
            choose(5, "talk", 1, "npc:other"),
            test.step(1),
            choose(6, "talk", 1, "npc:akara", "bob", 1),
            test.step(1),
            choose(7, "trade", 1, nil, nil, 5),
            test.step(1),
            test.run(function()
                test.expect(context():get("dialogue_page")):equals("menu")
                test.expect(context():get("dialogue_revision")):equals(1)
                test.expect(context("bob"):get("dialogue_page")):equals("")
                local ok = pcall(require("d2legacy.commands.interactions").validate_choice, {
                    authority = "player",
                    player = "bob",
                    payload = { owner = "alice", target = "npc:akara", choice = "talk", revision = 1 },
                })
                test.expect(ok):is_false()
            end),
        }),
        test.case("choices_recheck_range_level_and_target_residency", {
            enter(),
            test.step(1),
            command(2, "interaction.open", { target = "npc:akara" }),
            test.step(1),
            test.run(function()
                require("engine.ecs/v1").get(player(), "d2legacy.world.position"):set("x", 90)
            end),
            choose(3, "talk", 1),
            test.step(1),
            test.run(function()
                test.expect(context():get("dialogue_revision")):equals(2)
                test.expect(context():get("dialogue_page")):equals("")
                local ecs = require("engine.ecs/v1")
                ecs.get(player(), "d2legacy.world.position"):set("x", 10)
            end),
            test.step(1),
            test.run(function()
                test.expect(context():get("dialogue_page")):equals("")
            end),
            command(5, "interaction.open", { target = "npc:akara" }, nil, 3),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                ecs.get(player(), "d2legacy.world.location"):set("level_id", 2)
            end),
            choose(6, "talk", 3, nil, nil, 4),
            test.step(1),
            test.run(function()
                test.expect(context():get("dialogue_revision")):equals(4)
                local ecs = require("engine.ecs/v1")
                ecs.get(player(), "d2legacy.world.location"):set("level_id", 1)
            end),
            command(7, "interaction.open", { target = "npc:akara" }, nil, 5),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                ecs.set(context():get("target"), "d2legacy.world.inactive", {})
            end),
            choose(8, "talk", 5, nil, nil, 6),
            test.step(1),
            test.run(function()
                test.expect(context():get("dialogue_revision")):equals(6)
                test.expect(context():get("dialogue_page")):equals("")
                test.expect(require("d2legacy.interactions.context").active("alice")):is_nil()
            end),
        }),
    },
})
