local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local endpoints = {
    {
        id = "static-warp:1:3",
        pair_id = "static-warp:8:1",
        static = true,
        label = "DEN OF EVIL",
        level_id = 1,
        x = 12,
        y = 12,
        radius = 4,
        destination_level = 8,
        destination_x = 20,
        destination_y = 20,
        destination_width = 120,
        destination_height = 120,
    },
    {
        id = "static-warp:8:1",
        pair_id = "static-warp:1:3",
        static = true,
        label = "BLOOD MOOR",
        level_id = 8,
        x = 22,
        y = 20,
        radius = 4,
        destination_level = 1,
        destination_x = 10,
        destination_y = 12,
        destination_width = 100,
        destination_height = 80,
    },
}

-- enter admits the player before static interaction scenarios begin.
local function enter()
    return {
        test.submit_system({ tick = 1, sequence = 1, kind = "system.player.enter", payload = fixtures.amazon_entry }),
        test.step(1),
    }
end

-- open_at submits an ordinary interaction command and advances the authoritative tick.
local function open_at(tick, id)
    return {
        test.submit({
            tick = tick,
            sequence = tick - 1,
            player = "alice",
            kind = "interaction.open",
            payload = { target = id },
        }),
        test.step(1),
    }
end

-- player_at checks the committed location and optional transition kind after a command.
local function player_at(level, x, y, kind)
    return test.run(function()
        local ecs = require("engine.ecs/v1")
        local entity = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
        local location = ecs.get(entity, "d2legacy.world.location")
        local position = ecs.get(entity, "d2legacy.world.position")
        test.expect(location:get("level_id")):equals(level)
        test.expect(position:get("x")):equals(x)
        test.expect(position:get("y")):equals(y)
        if kind then
            test.expect(location:get("transition_kind")):equals(kind)
        end
    end)
end

-- pair_entities resolves current selectable IDs so mutation cases alter live ECS state.
local function pair_entities()
    local ecs = require("engine.ecs/v1")
    local by_id = {}
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.world.warp", "d2legacy.world.selectable" } })) do
        by_id[ecs.get(entity, "d2legacy.world.selectable"):get("id")] = entity
    end
    return ecs, by_id
end

-- joined preserves action order across setup, mutation, click, and assertion phases.
local function joined(...)
    local result = {}
    for _, actions in ipairs({ ... }) do
        for _, action in ipairs(actions) do
            result[#result + 1] = action
        end
    end
    return result
end

-- reject_after_mutation proves a stale reciprocal pair cannot relocate the player.
local function reject_after_mutation(change)
    return joined(enter(), {
        test.run(function()
            local ecs, by_id = pair_entities()
            change(ecs, by_id)
        end),
    }, open_at(2, "static-warp:1:3"), { player_at(1, 10, 12) })
end

return test.suite({
    profile = "authority",
    tier = "integration",
    covers = { "internal/game/transition" },
    initial_data = { ["d2legacy.world_warps"] = { endpoints = endpoints } },
    cases = {
        test.case(
            "tokenless_static_pair_is_selectable_and_warps_both_directions",
            joined(
                enter(),
                {
                    test.run(function()
                        local ecs, by_id = pair_entities()
                        test.expect(#ecs.query({ all = { "d2legacy.world.warp" } })):equals(2)
                        test.assert(by_id["static-warp:1:3"] and by_id["static-warp:8:1"], "static pair absent")
                        test.assert(
                            not ecs.get(by_id["static-warp:1:3"], "d2legacy.world.warp_appearance"),
                            "cave has portal sprite"
                        )
                        test.expect(#require("d2legacy.gameplay.world").warp_snapshots()):equals(0)
                    end),
                },
                open_at(2, "static-warp:1:3"),
                { player_at(8, 20, 20, "warp") },
                open_at(3, "static-warp:8:1"),
                { player_at(1, 10, 12, "warp") }
            )
        ),
        test.case(
            "removed_reciprocal_endpoint_rejects_operation",
            reject_after_mutation(function(ecs, pair)
                ecs.remove(pair["static-warp:8:1"], "d2legacy.world.warp")
            end)
        ),
        test.case(
            "inactive_reciprocal_endpoint_rejects_operation",
            reject_after_mutation(function(ecs, pair)
                ecs.set(pair["static-warp:8:1"], "d2legacy.world.inactive", {})
            end)
        ),
        test.case(
            "stale_clicked_endpoint_rejects_operation",
            reject_after_mutation(function(ecs, pair)
                ecs.remove(pair["static-warp:1:3"], "d2legacy.world.warp")
            end)
        ),
        test.case(
            "changed_reciprocal_identity_rejects_operation",
            reject_after_mutation(function(ecs, pair)
                ecs.get(pair["static-warp:8:1"], "d2legacy.world.warp"):set("pair_id", "another")
            end)
        ),
        test.case(
            "changed_reciprocal_level_rejects_operation",
            reject_after_mutation(function(ecs, pair)
                ecs.get(pair["static-warp:8:1"], "d2legacy.world.location"):set("level_id", 3)
            end)
        ),
        test.case(
            "changed_reciprocal_destination_rejects_operation",
            reject_after_mutation(function(ecs, pair)
                ecs.get(pair["static-warp:8:1"], "d2legacy.world.warp"):set("destination_level", 3)
            end)
        ),
        test.case(
            "wrong_level_static_request_does_not_relocate",
            joined(enter(), {
                test.run(function()
                    local ecs = require("engine.ecs/v1")
                    local entity = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
                    ecs.get(entity, "d2legacy.world.location"):set("level_id", 3)
                end),
            }, open_at(2, "static-warp:1:3"), { player_at(3, 10, 12) })
        ),
        test.case(
            "far_static_request_does_not_relocate",
            joined(enter(), {
                test.run(function()
                    local ecs = require("engine.ecs/v1")
                    local entity = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
                    ecs.get(entity, "d2legacy.world.position"):set("x", 80)
                    ecs.get(entity, "d2legacy.world.position"):set("y", 80)
                end),
            }, open_at(2, "static-warp:1:3"), { player_at(1, 80, 80) })
        ),
    },
})
