local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local function submit(tick, sequence, kind, payload)
    return test.submit({ tick = tick, sequence = sequence, player = "alice", kind = kind, payload = payload })
end

local function sell(tick, sequence, vendor, category)
    return submit(tick, sequence, "item.vendor_sell", {
        vendor = vendor or "akara",
        category = category or "misc",
        item_id = "sale",
    })
end

local function buy(tick, sequence, vendor, item_id)
    return submit(tick, sequence, "item.vendor_buy", {
        vendor = vendor or "akara",
        item_id = item_id or "stock",
    })
end

local function state(gold, sale_container, stock_container)
    local ecs = require("engine.ecs/v1")
    local find = require("d2legacy.items.find")
    local layout_entity, layout = find.layout("alice")
    local sale = find.item(layout_entity, "sale")
    local stock = find.item(layout_entity, "stock")
    test.expect(sale ~= nil and stock ~= nil):is_true()
    test.expect(layout:get("carried_gold")):equals(gold)
    test.expect(#ecs.query({ all = { "d2legacy.item.identity" } })):equals(3)
    local offers = ecs.query({ all = { "d2legacy.vendor.offer" } })
    test.expect(#offers):equals(1)
    test.expect(ecs.get(offers[1], "d2legacy.vendor.offer"):get("purchases")):equals(0)
    test.expect(ecs.get(sale, "d2legacy.item.placement"):get("container")):equals(sale_container)
    test.expect(ecs.get(sale, "d2legacy.item.placement"):get("slot"))
        :equals(sale_container == "vendor" and "misc" or "")
    test.expect(ecs.get(stock, "d2legacy.item.placement"):get("container")):equals(stock_container)
    test.expect(ecs.get(stock, "d2legacy.item.placement"):get("slot"))
        :equals(stock_container == "vendor" and "misc" or "")
    if gold == 1000 then
        for _, entity in ipairs({ sale, stock }) do
            local placed = ecs.get(entity, "d2legacy.item.placement")
            test.expect(placed:get("x")):equals(0)
            test.expect(placed:get("y")):equals(0)
            test.expect(placed:get("page")):equals(0)
        end
    end
end

local function opened(name, steps)
    local actions = {
        test.submit_system({ tick = 1, sequence = 1, kind = "system.player.enter", payload = fixtures.player_entry() }),
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

local function refused_then_valid(name, change, invalid, restore, revision)
    return opened(name, {
        test.run(change),
        invalid,
        test.step(1),
        test.run(function()
            state(1000, "held", "vendor")
        end),
        test.run(restore),
        submit(5, 4, "interaction.open", { target = "npc:akara" }),
        test.step(1),
        test.run(function()
            local context = require("d2legacy.interactions.context").for_owner("alice")
            test.expect(context:get("dialogue_revision")):equals(revision)
        end),
        submit(6, 5, "interaction.choose", { target = "npc:akara", choice = "trade", revision = revision }),
        test.step(1),
        test.run(function()
            local context = require("d2legacy.interactions.context").for_owner("alice")
            test.expect(context:get("dialogue_page")):equals("trade")
        end),
        sell(7, 6),
        test.step(1),
        test.run(function()
            state(1060, "vendor", "vendor")
        end),
    })
end

local function player_entity()
    return require("engine.ecs/v1").query({ all = { "d2legacy.player.progress" } })[1]
end

local function target_entity()
    local ecs = require("engine.ecs/v1")
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.interaction.target" } })) do
        if ecs.get(entity, "d2legacy.interaction.target"):get("id") == "npc:akara" then
            return entity
        end
    end
    error("missing Akara target")
end

return test.suite({
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
            trade_terms = { akara = { buy_multiplier = 1024, sell_multiplier = 1024 } },
            items = {
                {
                    id = "stock",
                    code = "hp1",
                    width = 1,
                    height = 1,
                    base_cost = 40,
                    container = "vendor",
                    slot = "misc",
                },
                {
                    id = "sale",
                    code = "hp1",
                    width = 1,
                    height = 1,
                    base_cost = 60,
                    container = "held",
                },
            },
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
        opened("matching_trade_allows_ordinary_sell_and_buy", {
            sell(4, 3, "AKARA"),
            test.step(1),
            test.run(function()
                state(1060, "vendor", "vendor")
            end),
            buy(5, 4, "Akara"),
            test.step(1),
            test.run(function()
                state(1020, "vendor", "held")
            end),
            test.run(function()
                local ok, problem = pcall(require("d2legacy.commands.item_vendor").apply_sell, {
                    authority = "player",
                    player = "alice",
                    payload = { item_id = "sale", vendor = "akara", category = "misc" },
                })
                test.expect(ok):is_false()
                test.expect(string.find(problem, "item is not held", 1, true) ~= nil):is_true()
                state(1020, "vendor", "held")
            end),
            test.expect_checkpoint_parity(1),
        }),
        opened("closed_page_refuses_and_reopen_restores_trade", {
            submit(4, 3, "interaction.close", {}),
            test.step(1),
            sell(5, 4),
            buy(5, 5),
            test.step(1),
            test.run(function()
                state(1000, "held", "vendor")
            end),
            submit(6, 6, "interaction.open", { target = "npc:akara" }),
            test.step(1),
            submit(7, 7, "interaction.choose", { target = "npc:akara", choice = "trade", revision = 4 }),
            test.step(1),
            sell(8, 8),
            test.step(1),
            test.run(function()
                state(1060, "vendor", "vendor")
            end),
        }),
        opened("mismatched_and_unknown_vendor_refuse_without_crashing", {
            sell(4, 3, "charsi"),
            buy(4, 4, "charsi"),
            sell(4, 5, "unknown"),
            buy(4, 6, "unknown"),
            test.step(1),
            test.run(function()
                state(1000, "held", "vendor")
            end),
            test.expect_checkpoint_parity(1),
            sell(6, 1),
            test.step(1),
            test.run(function()
                state(1060, "vendor", "vendor")
            end),
        }),
        refused_then_valid(
            "out_of_range_refuses",
            function()
                require("engine.ecs/v1").get(player_entity(), "d2legacy.world.position"):set("x", 90)
            end,
            sell(4, 3),
            function()
                require("engine.ecs/v1").get(player_entity(), "d2legacy.world.position"):set("x", 10)
            end,
            4
        ),
        refused_then_valid(
            "missing_admitted_player_refuses",
            function()
                require("engine.ecs/v1").remove(player_entity(), "d2legacy.player.identity")
            end,
            sell(4, 3),
            function()
                require("engine.ecs/v1").set(player_entity(), "d2legacy.player.identity", {
                    character_id = "hero",
                    player = "alice",
                    name = "Hero",
                    class = "Amazon",
                })
            end,
            3
        ),
        refused_then_valid(
            "missing_player_position_refuses",
            function()
                require("engine.ecs/v1").remove(player_entity(), "d2legacy.world.position")
            end,
            sell(4, 3),
            function()
                require("engine.ecs/v1").set(player_entity(), "d2legacy.world.position", { x = 10, y = 12 })
            end,
            4
        ),
        opened("missing_player_location_refuses_without_tick_system_failure", {
            test.run(function()
                -- The stamina system also requires this component at tick boundaries.
                -- Exercise the apply-time guard while the fact is absent, then restore it.
                local ecs = require("engine.ecs/v1")
                local entity = player_entity()
                ecs.remove(entity, "d2legacy.world.location")
                require("d2legacy.commands.item_vendor").apply_sell({
                    authority = "player",
                    player = "alice",
                    payload = { item_id = "sale", vendor = "akara", category = "misc" },
                })
                state(1000, "held", "vendor")
                ecs.set(entity, "d2legacy.world.location", { act = 1, level_id = 1 })
            end),
            sell(4, 3),
            test.step(1),
            test.run(function()
                state(1060, "vendor", "vendor")
            end),
        }),
        refused_then_valid(
            "missing_target_location_refuses",
            function()
                require("engine.ecs/v1").remove(target_entity(), "d2legacy.world.location")
            end,
            sell(4, 3),
            function()
                require("engine.ecs/v1").set(target_entity(), "d2legacy.world.location", { act = 1, level_id = 1 })
            end,
            3
        ),
        refused_then_valid(
            "different_level_refuses",
            function()
                require("engine.ecs/v1").get(player_entity(), "d2legacy.world.location"):set("level_id", 2)
            end,
            sell(4, 3),
            function()
                require("engine.ecs/v1").get(player_entity(), "d2legacy.world.location"):set("level_id", 1)
            end,
            4
        ),
        refused_then_valid(
            "disallowed_category_refuses",
            function() end,
            sell(4, 3, "akara", "armor"),
            function() end,
            3
        ),
        opened("buy_uses_authoritative_vendor_category", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local layout = require("d2legacy.items.find").layout("alice")
                local stock = require("d2legacy.items.find").item(layout, "stock")
                test.expect(stock ~= nil):is_true()
                ecs.get(stock, "d2legacy.item.placement"):set("slot", "armor")
            end),
            buy(4, 3),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local layout = require("d2legacy.items.find").layout("alice")
                local stock = require("d2legacy.items.find").item(layout, "stock")
                test.expect(stock ~= nil):is_true()
                test.expect(ecs.get(stock, "d2legacy.item.placement"):get("container")):equals("vendor")
                test.expect(ecs.get(stock, "d2legacy.item.placement"):get("slot")):equals("armor")
                local _, wallet = require("d2legacy.items.find").layout("alice")
                test.expect(wallet:get("carried_gold")):equals(1000)
                ecs.get(stock, "d2legacy.item.placement"):set("slot", "misc")
            end),
            sell(5, 4),
            test.step(1),
            buy(6, 5),
            test.step(1),
            test.run(function()
                state(1020, "vendor", "held")
            end),
        }),
        opened("missing_ordinary_buy_is_a_no_op", {
            buy(4, 3, "akara", "missing"),
            test.step(1),
            test.run(function()
                state(1000, "held", "vendor")
            end),
            sell(5, 4),
            test.step(1),
            test.run(function()
                state(1060, "vendor", "vendor")
            end),
        }),
    },
})
