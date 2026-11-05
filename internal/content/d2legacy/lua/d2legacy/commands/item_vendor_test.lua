local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local function submit(tick, sequence, kind, payload, player)
    return test.submit({ tick = tick, sequence = sequence, player = player or "alice", kind = kind, payload = payload })
end

local function buy(tick, sequence, code, player)
    return submit(
        tick,
        sequence,
        "item.vendor_buy",
        { vendor = "akara", item_id = "akara-offer:" .. code, to_inventory = true },
        player
    )
end

local function setup(name, steps, level)
    local actions = {
        test.submit_system({
            tick = 1,
            sequence = 1,
            kind = "system.player.enter",
            payload = fixtures.player_entry({ level = level or 1 }),
        }),
        test.step(1),
        submit(2, 1, "interaction.open", { target = "npc:akara" }),
        test.step(1),
        submit(3, 2, "interaction.choose", { target = "npc:akara", choice = "trade", revision = 1 }),
        test.step(1),
    }
    for _, action in ipairs(steps) do
        actions[#actions + 1] = action
    end
    return test.case(name, actions)
end

local function state(gold, bought)
    local find = require("d2legacy.items.find")
    local ecs = require("engine.ecs/v1")
    local layout, value = find.layout("alice")
    test.expect(value:get("carried_gold")):equals(gold)
    test.expect(#ecs.query({ all = { "d2legacy.vendor.offer" } })):equals(2)
    test.expect(#ecs.query({ all = { "d2legacy.item.identity" } })):equals(2 + bought)
    return layout, value
end

local function quest_state(id, completed, rewarded)
    return test.run(function()
        local ecs = require("engine.ecs/v1")
        local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
        ecs.set(player, "d2legacy.quest.progress", {
            quest_id = id,
            stage = 1,
            completed = completed,
            rewarded = rewarded,
        })
    end)
end

local suite = {
    profile = "authority",
    tier = "integration",
    records = {
        ["data/global/excel/npc.txt"] = { { npc = "akara", ["sell mult"] = "1024" } },
        ["data/global/excel/misc.txt"] = {
            {
                code = "tsc",
                cost = "100",
                spawnable = "1",
                PermStoreItem = "1",
                AkaraMax = "17",
                stackable = "0",
                invwidth = "1",
                invheight = "1",
                invfile = "synthetic-scroll",
            },
            {
                code = "hp1",
                cost = "30",
                spawnable = "1",
                PermStoreItem = "1",
                AkaraMax = "1",
                stackable = "0",
                invwidth = "1",
                invheight = "1",
                invfile = "synthetic-potion",
            },
        },
    },
    initial_data = {
        ["d2legacy.items"] = {
            owner = "alice",
            inventory_width = 2,
            inventory_height = 2,
            vendor_width = 10,
            vendor_height = 10,
            carried_gold = 1000,
            items = {},
        },
        ["d2legacy.interactions"] = {
            owner = "alice",
            targets = {
                {
                    id = "npc:akara",
                    npc = "Akara",
                    vendor = "Akara",
                    categories = "misc",
                    x = 12,
                    y = 12,
                    radius = 4,
                    act = 1,
                    level_id = 1,
                    selectable = true,
                },
            },
        },
    },
    cases = {
        setup("den_keeps_quotes_debits_supply_and_reward_across_checkpoint", {
            quest_state(1, false, false),
            buy(4, 3, "hp1"),
            test.step(1),
            quest_state(1, true, false),
            buy(5, 4, "tsc"),
            test.step(1),
            submit(6, 5, "quest.claim_reward", {}),
            test.step(1),
            buy(7, 6, "hp1"),
            test.step(1),
            test.run(function()
                state(840, 3)
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                test.expect(ecs.get(player, "d2legacy.quest.progress"):get("rewarded")):is_true()
                test.expect(ecs.get(player, "d2legacy.player.progress"):get("unspent_skill_points")):equals(1)
                test.expect(require("d2legacy.items.vendor_offers").supported("alice")):is_true()
                require("d2legacy.commands.quests").complete({ payload = { player = "alice", quest_id = 1 } })
                test.expect(ecs.get(player, "d2legacy.quest.progress"):get("rewarded")):is_true()
                test.expect(#ecs.query({ all = { "d2legacy.quest.event" } })):equals(1)
            end),
            test.expect_checkpoint_parity(1),
            test.restore_checkpoint(),
            submit(9, 1, "quest.claim_reward", {}),
            test.step(1),
            buy(10, 2, "hp1"),
            test.step(1),
            test.run(function()
                state(810, 4)
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                test.expect(ecs.get(player, "d2legacy.player.progress"):get("unspent_skill_points")):equals(1)
            end),
        }),
        test.case("valid_empty_stock_allows_trade_refusal_and_close", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry(),
            }),
            test.step(1),
            test.run(function()
                local original = require("engine.records/v1")
                test.mock_module("engine.records/v1", {
                    load = function(path)
                        if path == "data/global/excel/misc.txt" then
                            return {}
                        end
                        return original.load(path)
                    end,
                })
            end),
            submit(2, 1, "interaction.open", { target = "npc:akara" }),
            test.step(1),
            submit(3, 2, "interaction.choose", { target = "npc:akara", choice = "trade", revision = 1 }),
            test.step(1),
            buy(4, 3, "tsc"),
            test.step(1),
            submit(5, 4, "interaction.close", {}),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local _, layout = require("d2legacy.items.find").layout("alice")
                test.expect(layout:get("carried_gold")):equals(1000)
                test.expect(#ecs.query({ all = { "d2legacy.item.identity" } })):equals(0)
                test.expect(require("d2legacy.interactions.context").for_owner("alice"):get("dialogue_page")):equals("")
            end),
        }),
        setup("higher_level_purchases_keep_distinct_items_and_restore_supply", {
            buy(4, 3, "tsc"),
            buy(4, 4, "tsc"),
            test.step(1),
            test.run(function()
                local layout = state(800, 2)
                local find = require("d2legacy.items.find")
                local ecs = require("engine.ecs/v1")
                local a = find.item(layout, "akara-purchase:tsc:1")
                local b = find.item(layout, "akara-purchase:tsc:2")
                local player = ecs.query({ all = { "d2legacy.player.progress" } })[1]
                test.expect(ecs.get(player, "d2legacy.player.progress"):get("level")):equals(25)
                test.expect(a ~= nil and b ~= nil):is_true()
                test.expect(ecs.get(a, "d2legacy.item.placement"):get("x")):equals(0)
                test.expect(ecs.get(b, "d2legacy.item.placement"):get("x")):equals(1)
                local snapshot = require("d2legacy.items.snapshot").for_owner("alice")
                test.expect(snapshot.items[2].buy_price):equals(100)
                local context = require("d2legacy.interactions.context").for_owner("alice")
                test.expect(context:get("trade_available")):is_true()
            end),
            test.expect_checkpoint_parity(1),
            test.restore_checkpoint(),
            buy(6, 1, "hp1"),
            test.step(1),
            submit(7, 2, "interaction.close", {}),
            test.step(1),
            buy(8, 3, "tsc"),
            test.step(1),
            test.run(function()
                state(770, 3)
            end),
            submit(9, 4, "interaction.open", { target = "npc:akara" }),
            test.step(1),
            submit(10, 5, "interaction.choose", { target = "npc:akara", choice = "trade", revision = 4 }),
            test.step(1),
            buy(11, 6, "tsc"),
            test.step(1),
            test.run(function()
                state(670, 4)
            end),
        }, 25),
        setup("full_inventory_refusal_keeps_offer_and_wallet_and_can_close", {
            buy(4, 3, "tsc"),
            test.step(1),
            buy(5, 4, "tsc"),
            test.step(1),
            buy(6, 5, "tsc"),
            test.step(1),
            buy(7, 6, "tsc"),
            test.step(1),
            buy(8, 7, "tsc"),
            test.step(1),
            submit(9, 8, "interaction.close", {}),
            test.step(1),
            test.run(function()
                state(600, 4)
                test.expect(require("d2legacy.interactions.context").for_owner("alice"):get("dialogue_page")):equals("")
            end),
        }),
        setup("insufficient_gold_and_foreign_item_requests_are_safe", {
            test.run(function()
                local _, layout = require("d2legacy.items.find").layout("alice")
                layout:set("carried_gold", 99)
            end),
            buy(4, 3, "tsc"),
            test.step(1),
            buy(5, 1, "tsc", "bob"),
            test.step(1),
            buy(6, 4, "hp1"),
            test.step(1),
            test.run(function()
                state(69, 1)
            end),
        }),
        setup("out_of_range_context_refuses_without_debit", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                ecs.get(player, "d2legacy.world.position"):set("x", 90)
            end),
            buy(4, 3, "tsc"),
            test.step(1),
            test.run(function()
                state(1000, 0)
            end),
        }),
        setup("held_item_refuses_and_later_inventory_purchase_succeeds", {
            buy(4, 3, "tsc"),
            test.step(1),
            submit(5, 4, "item.move", { item_id = "akara-purchase:tsc:1", destination = { container = "held" } }),
            test.step(1),
            buy(6, 5, "hp1"),
            test.step(1),
            test.run(function()
                state(900, 1)
            end),
            submit(
                7,
                6,
                "item.move",
                { item_id = "akara-purchase:tsc:1", destination = { container = "inventory", x = 0, y = 0 } }
            ),
            test.step(1),
            buy(8, 7, "hp1"),
            test.step(1),
            test.run(function()
                state(870, 2)
            end),
        }),
        setup("level_change_keeps_permanent_offers_available", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                ecs.get(ecs.query({ all = { "d2legacy.player.progress" } })[1], "d2legacy.player.progress")
                    :set("level", 2)
            end),
            buy(4, 3, "tsc"),
            test.step(1),
            buy(5, 4, "hp1"),
            test.step(1),
            test.run(function()
                state(870, 2)
            end),
        }),
        setup("wrong_vendor_and_missing_offer_refuse_without_losing_session", {
            submit(4, 3, "item.vendor_buy", { vendor = "charsi", item_id = "akara-offer:tsc", to_inventory = true }),
            test.step(1),
            submit(5, 4, "item.vendor_buy", { vendor = "akara", item_id = "missing", to_inventory = true }),
            test.step(1),
            buy(6, 5, "hp1"),
            test.step(1),
            test.run(function()
                state(970, 1)
            end),
        }),
        setup("reduced_price_sources_refuse_the_undiscounted_offer", {
            quest_state(1, true, true),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                ecs.create({
                    ["d2legacy.stat.source"] = {
                        target = player,
                        source_id = "discount",
                        stat = "item_reducedprices",
                        operation = "add",
                        value = 10,
                    },
                })
            end),
            buy(4, 3, "tsc"),
            test.step(1),
            test.run(function()
                state(1000, 0)
            end),
        }),
        setup("dead_player_cannot_buy_from_a_retained_trade_context", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                ecs.get(player, "d2legacy.player.vitals"):set("health", 0)
            end),
            buy(4, 3, "tsc"),
            test.step(1),
            test.run(function()
                state(1000, 0)
            end),
        }),
    },
}

for _, id in ipairs({ 4, 999 }) do
    for _, rewarded in ipairs({ false, true }) do
        suite.cases[#suite.cases + 1] = setup("completed_quest_refused_" .. id .. "_" .. tostring(rewarded), {
            quest_state(id, true, rewarded),
            buy(4, 3, "tsc"),
            test.step(1),
            submit(5, 4, "interaction.open", { target = "npc:akara" }),
            test.step(1),
            test.run(function()
                state(1000, 0)
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                test.expect(ecs.get(player, "d2legacy.quest.progress"):get("quest_id")):equals(id)
                local context = require("d2legacy.interactions.context").for_owner("alice")
                test.expect(context:get("trade_available")):is_false()
                test.expect(context:get("dialogue_page")):equals("menu")
            end),
        })
    end
end

return test.suite(suite)
