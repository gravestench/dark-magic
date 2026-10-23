local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local weapon = {
    id = "fractional-weapon",
    code = "ssd",
    width = 1,
    height = 1,
    body_slots = "rarm,larm",
    belt_eligible = false,
    base_cost = 100,
    container = "inventory",
    x = 2,
    y = 0,
    slot = "",
    belt_slot = 0,
    weapon_set = 0,
    page = 0,
    melee_range = 3,
    physical_min = 512,
    physical_max = 1024,
    melee_weapon_class = "1HS",
    attack_rating = 0,
    attack_rate = 0,
    stat_modifiers = {
        {
            source_id = "fractional-defense",
            source_kind = "affix",
            stat = "item_fractionaltargetac",
            operation = "add",
            value = 50,
            order = 10,
        },
    },
}

local function initial_items()
    return {
        owner = "alice",
        inventory_width = 10,
        inventory_height = 4,
        stash_width = 6,
        stash_height = 8,
        cube_width = 3,
        cube_height = 4,
        belt_capacity = 4,
        active_weapon_set = 0,
        vendor_width = 10,
        vendor_height = 10,
        carried_gold = 1000,
        stashed_gold = 0,
        items = { weapon },
    }
end

local function move_item(tick, sequence, destination)
    return {
        tick = tick,
        sequence = sequence,
        player = "alice",
        kind = "item.move",
        payload = { item_id = "fractional-weapon", destination = destination },
    }
end

local function expect_fractional_source(value)
    local ecs = require("engine.ecs/v1")
    local player = test.only_entity_with("d2legacy.player.identity")
    local sources = test.entities_with("d2legacy.stat.source")
    test.expect(require("d2legacy.gameplay.stat_sources").resolve(sources, player, "item_fractionaltargetac", 0))
        :equals(value)
    local found
    for _, entity in ipairs(sources) do
        local source = ecs.get(entity, "d2legacy.stat.source")
        if
            source:get("source_id")
            == "equipment:modifier:item_fractionaltargetac:fractional-weapon:affix:10:fractional-defense"
        then
            found = source:get("value")
        end
    end
    if value == 0 then
        test.expect(found):is_nil()
    else
        test.expect(found):equals(value)
    end
end

local function spawn_target(level, id, classification)
    local ecs = require("engine.ecs/v1")
    local target = ecs.create({
        ["d2legacy.monster.stats"] = {
            level = level,
            health = 100000,
            max_health = 100000,
            defense = 500,
            attack_rating = 300,
            physical_min = 0,
            physical_max = 0,
            experience = 0,
        },
        ["d2legacy.world.position"] = { x = 12, y = 12 },
        ["d2legacy.world.location"] = { act = 1, level_id = 1 },
        ["d2legacy.world.collider"] = { radius = 1 },
        ["d2legacy.world.selectable"] = {
            id = id,
            kind = "hostile",
            label = "Target",
            owner = "",
            radius = 1,
            priority = 1,
        },
    })
    if classification == "boss" then
        ecs.set(target, "d2legacy.monster.boss", {})
    elseif classification == "hireling" or classification == "inactive_hireling" then
        -- A hostile target reaches melee selection; the owned-unit relation
        -- supplies only the protected-target classification under test.
        ecs.set(target, "d2legacy.owned_unit", {
            owner_id = "player:alice",
            ultimate_owner_id = "player:alice",
            category = "hireable",
            group = 0,
            limit = 1,
            replacement = "replace_oldest",
            created_tick = 1,
            active = classification == "hireling",
        })
    end
    return target
end

local function request_player_melee(tick, target_id)
    require("engine.ecs/v1").set(
        test.only_entity_with("d2legacy.player.identity"),
        "d2legacy.combat.basic_attack_request",
        {
            target_id = target_id,
            request_tick = tick,
            hand = "rarm",
        }
    )
end

local function latest_melee(count)
    local ecs = require("engine.ecs/v1")
    local events = test.entities_with("d2legacy.combat.melee_event")
    test.expect(#events):equals(count)
    return ecs.get(events[#events], "d2legacy.combat.melee_event"),
        ecs.get(events[#events], "d2legacy.combat.attack_result")
end

local function expect_melee(count, attack_rating, defense, chance, hit)
    local melee, result = latest_melee(count)
    test.expect(melee:get("attack_rating")):equals(attack_rating)
    test.expect(melee:get("defense")):equals(defense)
    test.expect(melee:get("hit_chance")):equals(chance)
    test.expect(result:get("attack_rating")):equals(attack_rating)
    test.expect(result:get("defense")):equals(defense)
    test.expect(result:get("hit_chance")):equals(chance)
    test.expect(result:get("outcome")):equals(melee:get("hit") and "hit" or "miss")
    if hit ~= nil then
        test.expect(melee:get("hit")):equals(hit)
        test.expect(result:get("outcome")):equals(hit and "hit" or "miss")
    end
end

local function player_case(name, value, classification, defense, chance, hit)
    return test.case(name, {
        test.submit_system({
            tick = 1,
            sequence = 1,
            kind = "system.player.enter",
            payload = fixtures.player_entry({
                level = 20,
                base_attack_rating = 235,
                passive_stat_sources = {
                    { id = "fractional-defense", stat = "item_fractionaltargetac", operation = "add", value = value },
                },
            }),
        }),
        test.step(1),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            local player = test.only_entity_with("d2legacy.player.identity")
            local target = spawn_target(20, "monster:fractional-target", classification)
            test.expect(ecs.get(player, "d2legacy.player.progress"):get("level")):equals(20)
            test.expect(ecs.get(player, "d2legacy.player.combat_stats"):get("attack_rating")):equals(300)
            test.expect(ecs.get(target, "d2legacy.monster.stats"):get("level")):equals(20)
            request_player_melee(2, "monster:fractional-target")
        end),
        test.step(1),
        test.run(function()
            expect_melee(1, 300, defense, chance, hit)
        end),
    })
end

return test.suite({
    profile = "authority",
    tier = "integration",
    seed = 36,
    initial_data = { ["d2legacy.items"] = initial_items() },
    cases = {
        player_case("ordinary_monster_reduction_changes_chance_and_draw", 50, nil, 250, 54, true),
        player_case("boss_halves_fractional_target_defense", 50, "boss", 375, 44, false),
        player_case("active_hireling_halves_fractional_target_defense", 50, "hireling", 375, 44, false),
        player_case("inactive_hireling_marker_does_not_halve_reduction", 50, "inactive_hireling", 250, 54, true),
        player_case("zero_fractional_target_defense_keeps_full_defense", 0, nil, 500, 37, false),
        player_case("negative_fractional_target_defense_keeps_full_defense", -50, nil, 500, 37, false),
        test.case("monster_attacker_does_not_apply_fractional_target_defense", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ level = 20, defense = 495 }),
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = test.only_entity_with("d2legacy.player.identity")
                local monster = spawn_target(20, "monster:fractional-target")
                test.expect(ecs.get(player, "d2legacy.player.combat_stats"):get("defense")):equals(500)
                ecs.set(monster, "d2legacy.combat.melee_profile", {
                    range = 3,
                    physical_min = 0,
                    physical_max = 0,
                    primary_attack_rating = 300,
                    primary_hand = "rarm",
                    dual_wield = false,
                })
                ecs.create({
                    ["d2legacy.stat.source"] = {
                        target = monster,
                        source_id = "monster:fractional-defense",
                        stat = "item_fractionaltargetac",
                        operation = "add",
                        value = 50,
                        order = 1,
                    },
                })
                ecs.set(monster, "d2legacy.combat.basic_attack_request", {
                    target_id = "player:alice",
                    request_tick = 2,
                    hand = "rarm",
                })
            end),
            test.step(1),
            test.run(function()
                expect_melee(1, 300, 500, 37, false)
            end),
        }),
        test.case("fractional_target_defense_source_tracks_active_weapon", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry(),
            }),
            test.step(1),
            test.submit(move_item(2, 1, { container = "held" })),
            test.step(1),
            test.submit(move_item(3, 2, { container = "equipment", slot = "rarm", weapon_set = 0 })),
            test.step(2),
            test.run(function()
                expect_fractional_source(50)
                spawn_target(1, "monster:fractional-lifecycle")
                request_player_melee(5, "monster:fractional-lifecycle")
            end),
            test.step(1),
            test.run(function()
                expect_melee(1, 65, 250, 20)
            end),
            test.submit({
                tick = 6,
                sequence = 3,
                player = "alice",
                kind = "item.weapon_set",
                payload = { set = 1 },
            }),
            test.step(2),
            test.run(function()
                expect_fractional_source(0)
                request_player_melee(8, "monster:fractional-lifecycle")
            end),
            test.step(1),
            test.run(function()
                expect_melee(2, 65, 500, 11)
            end),
            test.submit({
                tick = 9,
                sequence = 4,
                player = "alice",
                kind = "item.weapon_set",
                payload = { set = 0 },
            }),
            test.step(2),
            test.run(function()
                expect_fractional_source(50)
                request_player_melee(11, "monster:fractional-lifecycle")
            end),
            test.step(1),
            test.run(function()
                expect_melee(3, 65, 250, 20)
            end),
            test.submit(move_item(12, 5, { container = "inventory", x = 2, y = 0 })),
            test.step(2),
            test.run(function()
                expect_fractional_source(0)
                request_player_melee(14, "monster:fractional-lifecycle")
            end),
            test.step(1),
            test.run(function()
                expect_melee(4, 65, 500, 11)
            end),
        }),
    },
})
