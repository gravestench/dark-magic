local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")
local ground_item, player, barrier, before, modifier_id
local retired_entity

local function state()
    local ecs = require("engine.ecs/v1")
    local identity = ecs.get(ground_item, "d2legacy.item.identity")
    return {
        owner = identity:get("owner"):id(),
        id = identity:get("id"),
        code = identity:get("code"),
        placement = ecs.get(ground_item, "d2legacy.item.placement"):snapshot(),
        generation = ecs.get(ground_item, "d2legacy.item.generation"):snapshot(),
        selectable = ecs.get(ground_item, "d2legacy.world.selectable") ~= nil,
    }
end

local function prepare()
    local ecs = require("engine.ecs/v1")
    player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
    local layout = ecs.create({
        ["d2legacy.items.layout"] = {
            owner = "ground-source",
            inventory_width = 2,
            inventory_height = 3,
            stash_width = 0,
            stash_height = 0,
            cube_width = 0,
            cube_height = 0,
            belt_capacity = 0,
            active_weapon_set = 0,
            vendor_width = 0,
            vendor_height = 0,
            carried_gold = 0,
            stashed_gold = 0,
        },
    })
    ground_item = require("d2legacy.items.entity").create(layout, {
        id = "sword",
        code = "ssd",
        width = 1,
        height = 3,
        container = "world",
        x = 12,
        y = 12,
        act = 1,
        level_id = 1,
        generation = { quality = "normal", level = 1, facts = "original-rolled-facts" },
    })
    barrier = { clear = true }
    require("d2legacy.gameplay.collision").set(1, {
        line_clear = function()
            return barrier.clear
        end,
        integrate_velocity = function(_, x, y)
            return x, y
        end,
    })
    before = state()
end

local function pickup(tick, sequence, owner, target)
    return test.submit({
        tick = tick or 2,
        sequence = sequence or 1,
        player = owner or "alice",
        kind = "item.pickup",
        payload = { target_id = target or "item:1" },
    })
end

local function entered()
    return test.submit_system({
        tick = 1,
        sequence = 1,
        kind = "system.player.enter",
        payload = fixtures.player_entry(),
    })
end

local function refused_case(name, change, target)
    return test.case(name, {
        entered(),
        test.step(1),
        test.run(prepare),
        test.run(change),
        test.run(function()
            if
                name == "admitted_zero_health_actor_without_death_refused"
                or name == "admitted_inactive_actor_refused"
                or name == "admitted_committed_dead_actor_refused"
            then
                local world = require("d2legacy.items.world")
                local ecs = require("engine.ecs/v1")
                test.assert(not world.can_pick_up(player, ground_item), "direct actor bypassed shared pickup guard")
                test.assert(
                    not world.within_reach(player, ecs.get(ground_item, "d2legacy.world.location"), 12, 12),
                    "direct actor bypassed shared reach guard"
                )
            end
        end),
        pickup(2, 1, "alice", target),
        test.step(1),
        test.run(function()
            test.expect(state()):deep_equals(before)
        end),
    })
end

local cases = {
    test.case("retired_selector_cannot_alias_a_reallocated_entity_after_restore", {
        entered(),
        test.step(1),
        test.run(prepare),
        test.run(function()
            retired_entity = ground_item:id()
            local ecs = require("engine.ecs/v1")
            test.expect(ecs.get(ground_item, "d2legacy.world.selectable"):get("id")):equals("item:1")
            ecs.destroy(ground_item)
        end),
        test.step(1),
        test.expect_checkpoint_parity(1),
        test.restore_checkpoint(),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            ground_item =
                require("d2legacy.items.entity").create(require("d2legacy.items.find").layout("ground-source"), {
                    id = "sword",
                    code = "ssd",
                    width = 1,
                    height = 3,
                    container = "world",
                    x = 12,
                    y = 12,
                    act = 1,
                    level_id = 1,
                    generation = { quality = "normal", level = 1, facts = "new-rolled-facts" },
                })
            test.expect(ground_item:id()):equals(retired_entity)
            test.expect(ecs.get(ground_item, "d2legacy.world.selectable"):get("id")):equals("item:2")
            require("d2legacy.gameplay.collision").set(1, {
                line_clear = function()
                    return true
                end,
                integrate_velocity = function(_, x, y)
                    return x, y
                end,
            })
            before = state()
        end),
        pickup(4, 1, "alice", "item:1"),
        test.step(1),
        test.run(function()
            test.expect(state()):deep_equals(before)
        end),
        pickup(5, 2, "alice", "item:2"),
        test.step(1),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            test.expect(ecs.get(ground_item, "d2legacy.item.placement"):get("container")):equals("inventory")
            test.expect(state().generation):deep_equals(before.generation)
        end),
    }),
    test.case("departed_killer_cannot_delete_session_item_or_modifier", {
        entered(),
        test.submit_system({
            tick = 1,
            sequence = 2,
            kind = "system.player.enter",
            payload = fixtures.player_entry({ player = "bob", character_id = "bob", name = "Bob", x = 14, y = 12 }),
        }),
        test.step(1),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            local source = ecs.query({ all = { "d2legacy.items.world_source" } })[1]
            ground_item = require("d2legacy.items.entity").create(source, {
                id = "sword",
                code = "ssd",
                width = 1,
                height = 3,
                container = "world",
                x = 12,
                y = 12,
                act = 1,
                level_id = 1,
                generation = { quality = "normal", level = 1, facts = "original-rolled-facts" },
                stat_modifiers = {
                    {
                        source_id = "retained",
                        source_kind = "item",
                        stat = "strength",
                        operation = "add",
                        value = 7,
                        order = 1,
                    },
                },
            })
            modifier_id = ecs.query({ all = { "d2legacy.item.stat_modifier" } })[1]:id()
            require("d2legacy.gameplay.collision").set(1, {
                blocked_position = function()
                    return false
                end,
                line_clear = function()
                    return true
                end,
                integrate_velocity = function(_, x, y)
                    return x, y
                end,
            })
            before = state()
        end),
        test.submit_system({ tick = 2, sequence = 3, kind = "system.player.leave", payload = { player = "alice" } }),
        test.step(1),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            test.expect(state()):deep_equals(before)
            test.expect(#ecs.query({ all = { "d2legacy.items.world_source" } })):equals(1)
            test.expect(ecs.query({ all = { "d2legacy.item.stat_modifier" } })[1]:id()):equals(modifier_id)
        end),
        test.expect_checkpoint_parity(2),
        test.restore_checkpoint(),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            ground_item = ecs.query({ all = { "d2legacy.item.identity" } })[1]
            require("d2legacy.gameplay.collision").set(1, {
                blocked_position = function()
                    return false
                end,
                line_clear = function()
                    return true
                end,
                integrate_velocity = function(_, x, y)
                    return x, y
                end,
            })
            test.assert(not require("d2legacy.items.find").layout("alice"), "restore recreated departed fixture layout")
            for _, entity in ipairs(ecs.query({ all = { "d2legacy.interaction.context" } })) do
                test.assert(
                    ecs.get(entity, "d2legacy.interaction.context"):get("owner") ~= "alice",
                    "restore recreated departed fixture context"
                )
            end
            test.expect(#ecs.query({ all = { "d2legacy.item.identity" } })):equals(1)
        end),
        test.submit_system({
            tick = 5,
            sequence = 1,
            kind = "system.player.enter",
            payload = fixtures.player_entry({ player = "carol", character_id = "carol", name = "Carol" }),
        }),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            local world = require("d2legacy.items.world")
            local actor = world.player("bob")
            test.assert(actor, "restored Bob is not healthy active actor")
            test.assert(world.can_pick_up(actor, ground_item), "restored Bob cannot reach item")
            test.expect(ecs.get(ground_item, "d2legacy.world.selectable"):get("id")):equals("item:1")
        end),
        pickup(5, 1, "bob", "item:1"),
        test.step(1),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            local identity = ecs.get(ground_item, "d2legacy.item.identity")
            test.expect(ecs.get(identity:get("owner"), "d2legacy.items.layout"):get("owner")):equals("bob")
            test.expect(ecs.get(ground_item, "d2legacy.item.placement"):get("container")):equals("inventory")
            test.expect(ecs.query({ all = { "d2legacy.item.stat_modifier" } })[1]:id()):equals(modifier_id)
            test.expect(state().generation):deep_equals(before.generation)
            test.expect(#ecs.query({ all = { "d2legacy.items.world_source" } })):equals(1)
            local configured = false
            for _, entity in ipairs(ecs.query({ all = { "d2legacy.interaction.context" } })) do
                local context = ecs.get(entity, "d2legacy.interaction.context")
                if context:get("owner") == "carol" then
                    configured = ecs.get(context:get("target"), "d2legacy.interaction.target"):get("id")
                        == "restored-npc"
                end
            end
            test.assert(configured, "fresh admission lost restored interaction configuration")
        end),
    }),
    test.case("admitted_pickup_transfers_to_destination_owner_once", {
        entered(),
        test.step(1),
        test.run(prepare),
        pickup(),
        test.step(1),
        pickup(3, 2),
        test.step(1),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            local identity = ecs.get(ground_item, "d2legacy.item.identity")
            test.expect(ecs.get(identity:get("owner"), "d2legacy.items.layout"):get("owner")):equals("alice")
            test.expect(ecs.get(ground_item, "d2legacy.item.placement"):get("container")):equals("inventory")
            test.expect(state().generation):deep_equals(before.generation)
            test.expect(state().id):equals(before.id)
            test.assert(not ecs.get(ground_item, "d2legacy.world.selectable"), "picked item remained selectable")
            test.expect(#ecs.query({ all = { "d2legacy.item.identity" } })):equals(1)
        end),
    }),
    refused_case("admitted_zero_health_actor_without_death_refused", function()
        local ecs = require("engine.ecs/v1")
        ecs.get(player, "d2legacy.player.vitals"):set("health", 0)
        test.assert(not ecs.get(player, "d2legacy.player.death"), "zero-health control already committed death")
    end),
    refused_case("admitted_inactive_actor_refused", function()
        require("engine.ecs/v1").set(player, "d2legacy.world.inactive", {})
    end),
    refused_case("admitted_committed_dead_actor_refused", function()
        require("engine.ecs/v1").set(player, "d2legacy.player.death", {
            tick = 1,
            killer_id = "",
            credited_id = "",
            hardcore = false,
            stage = "dead",
            experience_lost = 0,
            consequences_pending = true,
        })
    end),
    refused_case("admitted_inactive_target_refused", function()
        require("engine.ecs/v1").set(ground_item, "d2legacy.world.inactive", {})
    end),
    refused_case("admitted_distant_target_refused", function()
        require("engine.ecs/v1").get(ground_item, "d2legacy.world.position"):set("x", 30)
    end),
    refused_case("admitted_other_level_refused", function()
        require("engine.ecs/v1").get(ground_item, "d2legacy.world.location"):set("level_id", 2)
    end),
    refused_case("admitted_other_act_refused", function()
        require("engine.ecs/v1").get(ground_item, "d2legacy.world.location"):set("act", 2)
    end),
    refused_case("admitted_blocked_target_refused", function()
        barrier.clear = false
    end),
    refused_case("admitted_full_destination_inventory_refused", function()
        local layout = require("d2legacy.items.find").layout("alice")
        require("d2legacy.items.entity").create(
            layout,
            { id = "occupied", code = "ssd", width = 1, height = 3, container = "inventory", x = 0, y = 0 }
        )
    end),
    refused_case("admitted_held_item_refuses_ground_pickup", function()
        local layout = require("d2legacy.items.find").layout("alice")
        require("d2legacy.items.entity").create(
            layout,
            { id = "cursor", code = "ssd", width = 1, height = 3, container = "held" }
        )
    end),
    refused_case("admitted_stale_target_refused", function() end, "item:ground-source:missing"),
    test.case("same_tick_admitted_race_transfers_once", {
        entered(),
        test.submit_system({
            tick = 1,
            sequence = 2,
            kind = "system.player.enter",
            payload = fixtures.player_entry({ player = "bob", name = "Bob", character_id = "bob" }),
        }),
        test.step(1),
        test.run(prepare),
        pickup(2, 1, "alice"),
        pickup(2, 1, "bob"),
        test.step(1),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            local identity = ecs.get(ground_item, "d2legacy.item.identity")
            test.expect(ecs.get(identity:get("owner"), "d2legacy.items.layout"):get("owner")):equals("alice")
            test.expect(#ecs.query({ all = { "d2legacy.item.identity" } })):equals(1)
            test.expect(state().generation):deep_equals(before.generation)
        end),
        test.expect_checkpoint_parity(2),
    }),
}

for _, condition in ipairs({ "distant", "blocked", "zero_health", "inactive" }) do
    cases[#cases + 1] = test.case("admitted_legacy_move_refuses_" .. condition, {
        entered(),
        test.step(1),
        test.run(prepare),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            ecs.get(ground_item, "d2legacy.item.identity"):set("owner", require("d2legacy.items.find").layout("alice"))
            before = state()
            if condition == "distant" then
                ecs.get(ground_item, "d2legacy.world.position"):set("x", 30)
            elseif condition == "blocked" then
                barrier.clear = false
            elseif condition == "zero_health" then
                ecs.get(player, "d2legacy.player.vitals"):set("health", 0)
            else
                ecs.set(player, "d2legacy.world.inactive", {})
            end
        end),
        test.submit({
            tick = 2,
            sequence = 1,
            player = "alice",
            kind = "item.move",
            payload = { item_id = "sword", destination = { container = "held" } },
        }),
        test.step(1),
        test.run(function()
            test.expect(state()):deep_equals(before)
        end),
    })
end

return test.suite({
    profile = "authority",
    tier = "fast",
    covers = { "internal/game/item" },
    initial_data = {
        ["d2legacy.interactions"] = {
            initial_target = "restored-npc",
            targets = { { id = "restored-npc", npc = "Akara", x = 10, y = 10, radius = 3.5 } },
        },
        ["d2legacy.items"] = { owner = "alice", inventory_width = 1, inventory_height = 3, items = test.array() },
    },
    cases = cases,
})
