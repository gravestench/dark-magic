local test = require("d2legacy.tests/v1")
local first_source_id

local function target(health, id)
    return require("engine.ecs/v1").create({
        ["d2legacy.world.selectable"] = { id = id or "monster:poison-target", kind = "hostile", label = "target" },
        ["d2legacy.monster.stats"] = { health = health, max_health = 4096 },
    })
end

local function hit(rate, length, attacker, target_id)
    require("engine.ecs/v1").create({
        ["d2legacy.combat.event"] = {
            kind = "damage_applied",
            tick = 1,
            attacker_id = attacker or "player:one",
            target_id = target_id or "monster:poison-target",
            source_kind = "melee",
            damage_channel = "poison",
            damage_raw = 0,
            remaining_health_raw = 4096,
        },
        ["d2legacy.combat.damage_bundle"] = {
            poison_rolled_raw = rate,
            poison_mitigated_raw = rate,
            poison_length_frames = length,
        },
    })
end

local function active()
    local ecs = require("engine.ecs/v1")
    local instances = ecs.query({ all = { "d2legacy.state.instance" } })
    test.expect(#instances):equals(1)
    return ecs.get(instances[1], "d2legacy.state.instance")
end

return test.suite({
    profile = "authority",
    tier = "fast",
    cases = {
        test.case("poison_only_hit_applies_attributed_rate_for_its_duration_once", {
            test.run(function()
                target(4096)
                hit(100, 5)
            end),
            test.step(1),
            test.run(function()
                local instance = active()
                test.expect(instance:get("state_id")):equals("poison")
                test.expect(instance:get("stat")):equals("hpregen")
                test.expect(instance:get("stat_value")):equals(-100)
                test.expect(instance:get("expires_tick")):equals(6)
                test.expect(instance:get("source_attacker_id")):equals("player:one")
                local ecs = require("engine.ecs/v1")
                local source = ecs.query({ all = { "d2legacy.stat.source" } })[1]
                test.expect(ecs.get(source, "d2legacy.stat.source"):get("value")):equals(-100)
            end),
            test.step(1),
            test.run(function()
                test.expect(#require("engine.ecs/v1").query({ all = { "d2legacy.state.instance" } })):equals(1)
            end),
        }),
        test.case("immune_zero_length_and_dead_targets_do_not_acquire_poison", {
            test.run(function()
                target(0, "monster:dead")
                target(4096, "monster:immune")
                target(4096, "monster:zero-length")
                hit(100, 5, "player:one", "monster:dead")
                hit(0, 5, "player:one", "monster:immune")
                hit(100, 0, "player:one", "monster:zero-length")
            end),
            test.step(1),
            test.run(function()
                test.expect(#require("engine.ecs/v1").query({ all = { "d2legacy.state.instance" } })):equals(0)
            end),
        }),
        test.case("weaker_hit_is_rejected_and_equal_hit_resets_expiry_from_same_attacker", {
            test.run(function()
                target(4096)
                hit(100, 5)
            end),
            test.step(1),
            test.run(function()
                hit(90, 10, "player:two")
            end),
            test.step(1),
            test.run(function()
                local previous = active()
                test.expect(previous:get("stat_value")):equals(-100)
                test.expect(previous:get("expires_tick")):equals(6)
                first_source_id = previous:get("source_id")
                hit(100, 2)
            end),
            test.step(1),
            test.run(function()
                local refreshed = active()
                test.expect(refreshed:get("expires_tick")):equals(5)
                test.expect(refreshed:get("source_attacker_id")):equals("player:one")
                test.expect(refreshed:get("applied_tick")):equals(3)
                test.assert(
                    refreshed:get("source_id") ~= first_source_id,
                    "a repeat hit has a new application identity"
                )
            end),
        }),
        test.case("same_frame_rank_selects_strongest_and_does_not_replay", {
            test.run(function()
                target(4096)
                hit(90, 10, "player:one")
                hit(120, 4, "player:two")
            end),
            test.step(1),
            test.run(function()
                local winner = active()
                test.expect(winner:get("stat_value")):equals(-120)
                test.expect(winner:get("source_attacker_id")):equals("player:two")
                test.expect(winner:get("expires_tick")):equals(5)
            end),
            test.step(1),
            test.run(function()
                local winner = active()
                test.expect(winner:get("stat_value")):equals(-120)
                test.expect(winner:get("expires_tick")):equals(5)
                test.expect(#require("engine.ecs/v1").query({ all = { "d2legacy.state.event" } })):equals(1)
            end),
        }),
    },
})
