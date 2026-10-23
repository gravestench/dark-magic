local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local weapon = {
    id = "target-type-weapon",
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
            source_id = "demon-ar",
            source_kind = "affix",
            stat = "item_demon_tohit",
            operation = "add",
            value = 50,
            order = 10,
        },
        {
            source_id = "undead-ar",
            source_kind = "affix",
            stat = "item_undead_tohit",
            operation = "add",
            value = 40,
            order = 11,
        },
    },
}

local offhand = {
    id = "offhand-weapon",
    code = "ssd",
    width = 1,
    height = 1,
    body_slots = "rarm,larm",
    belt_eligible = false,
    base_cost = 100,
    container = "inventory",
    x = 3,
    y = 0,
    slot = "",
    belt_slot = 0,
    weapon_set = 0,
    page = 0,
    melee_range = 3,
    physical_min = 512,
    physical_max = 1024,
    melee_weapon_class = "1HS",
    attack_rating = 20,
    attack_rate = 0,
}

local function move_item(tick, sequence, destination)
    return {
        tick = tick,
        sequence = sequence,
        player = "alice",
        kind = "item.move",
        payload = { item_id = weapon.id, destination = destination },
    }
end

local function expect_sources(demon, undead)
    local ecs = require("engine.ecs/v1")
    local resolve = require("d2legacy.gameplay.stat_sources").resolve
    local player = test.only_entity_with("d2legacy.player.identity")
    local entities = test.entities_with("d2legacy.stat.source")
    local function expect_source(stat, source_id, order, value)
        test.expect(resolve(entities, player, stat, 0)):equals(value)
        local found
        for _, entity in ipairs(entities) do
            local source = ecs.get(entity, "d2legacy.stat.source")
            if
                source:get("source_id")
                == "equipment:modifier:" .. stat .. ":" .. weapon.id .. ":affix:" .. order .. ":" .. source_id
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
    expect_source("item_demon_tohit", "demon-ar", 10, demon)
    expect_source("item_undead_tohit", "undead-ar", 11, undead)
    test.expect(resolve(entities, player, "item_ignoretargetac", 0)):equals(0)
    test.expect(resolve(entities, player, "item_fractionaltargetac", 0)):equals(0)
end

local function target_case(name, flags, demon_bonus, undead_bonus, expected_ar, expected_chance, expected_hit)
    return test.case(name, {
        test.submit_system({
            tick = 1,
            sequence = 1,
            kind = "system.player.enter",
            payload = fixtures.player_entry({
                level = 20,
                base_attack_rating = 235,
                passive_stat_sources = {
                    { id = "demon-ar", stat = "item_demon_tohit", operation = "add", value = demon_bonus },
                    { id = "undead-ar", stat = "item_undead_tohit", operation = "add", value = undead_bonus },
                },
            }),
        }),
        test.step(1),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            local player = test.only_entity_with("d2legacy.player.identity")
            test.expect(ecs.get(player, "d2legacy.player.combat_stats"):get("attack_rating")):equals(300)
            ecs.create({
                ["d2legacy.stat.source"] = {
                    target = player,
                    source_id = "test:percent-attack-rating",
                    stat = "item_tohit_percent",
                    operation = "percent",
                    value = 20,
                    order = 300,
                },
            })
            local target = ecs.create({
                ["d2legacy.monster.stats"] = {
                    level = 20,
                    health = 100000,
                    max_health = 100000,
                    defense = 400,
                    attack_rating = 300,
                    physical_min = 0,
                    physical_max = 0,
                    experience = 0,
                },
                ["d2legacy.world.position"] = { x = 12, y = 12 },
                ["d2legacy.world.location"] = { act = 1, level_id = 1 },
                ["d2legacy.world.collider"] = { radius = 1 },
                ["d2legacy.world.selectable"] = {
                    id = "monster:target-type",
                    kind = "hostile",
                    label = "Target",
                    owner = "",
                    radius = 1,
                    priority = 1,
                },
            })
            if flags then
                ecs.set(target, "d2legacy.monster.type_flags", flags)
            end
            ecs.set(player, "d2legacy.combat.basic_attack_request", {
                target_id = "monster:target-type",
                request_tick = 2,
                hand = "rarm",
            })
        end),
        test.step(1),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            local events = test.entities_with("d2legacy.combat.melee_event")
            test.expect(#events):equals(1)
            local melee = ecs.get(events[1], "d2legacy.combat.melee_event")
            local result = ecs.get(events[1], "d2legacy.combat.attack_result")
            for _, record in ipairs({ melee, result }) do
                test.expect(record:get("attack_rating")):equals(expected_ar)
                test.expect(record:get("defense")):equals(400)
                test.expect(record:get("hit_chance")):equals(expected_chance)
            end
            test.expect(melee:get("hit")):equals(expected_hit)
            test.expect(result:get("outcome")):equals(expected_hit and "hit" or "miss")
        end),
    })
end

local function hand_case(name, flags, hand, expected_ar)
    return test.case(name, {
        test.submit_system({
            tick = 1,
            sequence = 1,
            kind = "system.player.enter",
            payload = fixtures.player_entry({ class = "Barbarian", level = 20, base_attack_rating = 235 }),
        }),
        test.step(1),
        test.submit(move_item(2, 1, { container = "held" })),
        test.step(1),
        test.submit(move_item(3, 2, { container = "equipment", slot = "rarm", weapon_set = 0 })),
        test.step(1),
        test.submit({
            tick = 4,
            sequence = 3,
            player = "alice",
            kind = "item.move",
            payload = { item_id = offhand.id, destination = { container = "held" } },
        }),
        test.step(1),
        test.submit({
            tick = 5,
            sequence = 4,
            player = "alice",
            kind = "item.move",
            payload = {
                item_id = offhand.id,
                destination = { container = "equipment", slot = "larm", weapon_set = 0 },
            },
        }),
        test.step(2),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            local player = test.only_entity_with("d2legacy.player.identity")
            local profile = ecs.get(player, "d2legacy.combat.melee_profile")
            test.expect(profile:get("dual_wield")):equals(true)
            test.expect(profile:get("primary_attack_rating")):equals(0)
            test.expect(profile:get("secondary_attack_rating")):equals(20)
            ecs.create({
                ["d2legacy.stat.source"] = {
                    target = player,
                    source_id = "test:percent-attack-rating",
                    stat = "item_tohit_percent",
                    operation = "percent",
                    value = 20,
                    order = 300,
                },
            })
            local target = ecs.create({
                ["d2legacy.monster.stats"] = {
                    level = 20,
                    health = 100000,
                    max_health = 100000,
                    defense = 400,
                    attack_rating = 300,
                    physical_min = 0,
                    physical_max = 0,
                    experience = 0,
                },
                ["d2legacy.world.position"] = { x = 12, y = 12 },
                ["d2legacy.world.location"] = { act = 1, level_id = 1 },
                ["d2legacy.world.collider"] = { radius = 1 },
                ["d2legacy.world.selectable"] = {
                    id = "monster:hand-target",
                    kind = "hostile",
                    label = "Target",
                    owner = "",
                    radius = 1,
                    priority = 1,
                },
            })
            ecs.set(target, "d2legacy.monster.type_flags", flags)
            ecs.set(player, "d2legacy.combat.basic_attack_request", {
                target_id = "monster:hand-target",
                request_tick = 7,
                hand = hand,
            })
        end),
        test.step(1),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            local events = test.entities_with("d2legacy.combat.melee_event")
            test.expect(#events):equals(1)
            local melee = ecs.get(events[1], "d2legacy.combat.melee_event")
            local result = ecs.get(events[1], "d2legacy.combat.attack_result")
            test.expect(melee:get("attack_rating")):equals(expected_ar)
            test.expect(result:get("attack_rating")):equals(expected_ar)
        end),
    })
end

local function request_lifecycle_melee(tick)
    local ecs = require("engine.ecs/v1")
    ecs.set(test.only_entity_with("d2legacy.player.identity"), "d2legacy.combat.basic_attack_request", {
        target_id = "monster:lifecycle-target",
        request_tick = tick,
        hand = "rarm",
    })
end

local function expect_lifecycle_melee(count, attack_rating, chance)
    local ecs = require("engine.ecs/v1")
    local events = test.entities_with("d2legacy.combat.melee_event")
    test.expect(#events):equals(count)
    local melee = ecs.get(events[#events], "d2legacy.combat.melee_event")
    local result = ecs.get(events[#events], "d2legacy.combat.attack_result")
    for _, record in ipairs({ melee, result }) do
        test.expect(record:get("attack_rating")):equals(attack_rating)
        test.expect(record:get("hit_chance")):equals(chance)
    end
end

return test.suite({
    profile = "authority",
    tier = "integration",
    seed = 429,
    records = {
        ["data/global/excel/charstats.txt"] = {
            {
                class = "Amazon",
                StartSkill = "Fire Bolt",
                WalkVelocity = "6",
                RunVelocity = "9",
                vit = "20",
                stamina = "84",
                RunDrain = "20",
                StaminaPerLevel = "4",
                StaminaPerVitality = "4",
            },
            {
                class = "Barbarian",
                StartSkill = "Fire Bolt",
                WalkVelocity = "6",
                RunVelocity = "9",
                vit = "20",
                stamina = "84",
                RunDrain = "20",
                StaminaPerLevel = "4",
                StaminaPerVitality = "4",
            },
        },
    },
    initial_data = {
        ["d2legacy.items"] = {
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
            items = { weapon, offhand },
        },
    },
    cases = {
        target_case("demon_bonus_precedes_percent_and_changes_seeded_hit", { demon = true }, 50, 0, 420, 51, true),
        target_case("nonmatching_monster_does_not_get_target_bonus", {}, 50, 40, 360, 47, false),
        target_case("low_undead_uses_undead_bonus", { low_undead = true }, 50, 40, 408, 50, false),
        target_case("high_undead_uses_undead_bonus", { high_undead = true }, 50, 40, 408, 50, false),
        target_case(
            "demon_and_both_undead_bits_add_once",
            { demon = true, low_undead = true, high_undead = true },
            50,
            40,
            468,
            53,
            true
        ),
        target_case("negative_target_bonus_reduces_attack_rating", { demon = true }, -50, 0, 300, 42, false),
        target_case("missing_type_fact_defaults_to_no_bonus", nil, 50, 40, 360, 47, false),
        hand_case("primary_hand_keeps_percent_order_without_target_match", {}, "rarm", 360),
        hand_case("secondary_hand_delta_follows_percent_without_target_match", {}, "larm", 380),
        hand_case("primary_hand_keeps_percent_order_with_demon", { demon = true }, "rarm", 420),
        hand_case("secondary_hand_delta_follows_percent_with_demon", { demon = true }, "larm", 440),
        test.case("monster_attacker_ignores_target_type_attack_rating_sources", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ level = 20, defense = 395 }),
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = test.only_entity_with("d2legacy.player.identity")
                ecs.set(player, "d2legacy.monster.type_flags", { demon = true, low_undead = true })
                local monster = ecs.create({
                    ["d2legacy.monster.stats"] = {
                        level = 20,
                        health = 100000,
                        max_health = 100000,
                        defense = 0,
                        attack_rating = 300,
                        physical_min = 0,
                        physical_max = 0,
                        experience = 0,
                    },
                    ["d2legacy.world.position"] = { x = 12, y = 12 },
                    ["d2legacy.world.location"] = { act = 1, level_id = 1 },
                    ["d2legacy.world.collider"] = { radius = 1 },
                    ["d2legacy.world.selectable"] = {
                        id = "monster:attacker",
                        kind = "hostile",
                        label = "Attacker",
                        owner = "",
                        radius = 1,
                        priority = 1,
                    },
                    ["d2legacy.combat.melee_profile"] = {
                        range = 3,
                        physical_min = 0,
                        physical_max = 0,
                        primary_attack_rating = 300,
                        primary_hand = "rarm",
                        dual_wield = false,
                    },
                })
                for _, stat in ipairs({ "item_demon_tohit", "item_undead_tohit" }) do
                    ecs.create({
                        ["d2legacy.stat.source"] = {
                            target = monster,
                            source_id = "test:monster:" .. stat,
                            stat = stat,
                            operation = "add",
                            value = 50,
                            order = 1,
                        },
                    })
                end
                ecs.set(monster, "d2legacy.combat.basic_attack_request", {
                    target_id = "player:alice",
                    request_tick = 2,
                    hand = "rarm",
                })
            end),
            test.step(1),
            test.run(function()
                expect_lifecycle_melee(1, 300, 42)
            end),
        }),
        test.case("target_type_attack_rating_sources_track_active_weapon", {
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
                expect_sources(50, 40)
                local ecs = require("engine.ecs/v1")
                local player = test.only_entity_with("d2legacy.player.identity")
                test.expect(ecs.get(player, "d2legacy.player.combat_stats"):get("attack_rating")):equals(65)
                ecs.create({
                    ["d2legacy.monster.stats"] = {
                        level = 1,
                        health = 100000,
                        max_health = 100000,
                        defense = 400,
                        attack_rating = 300,
                        physical_min = 0,
                        physical_max = 0,
                        experience = 0,
                    },
                    ["d2legacy.monster.type_flags"] = { demon = true },
                    ["d2legacy.world.position"] = { x = 12, y = 12 },
                    ["d2legacy.world.location"] = { act = 1, level_id = 1 },
                    ["d2legacy.world.collider"] = { radius = 1 },
                    ["d2legacy.world.selectable"] = {
                        id = "monster:lifecycle-target",
                        kind = "hostile",
                        label = "Target",
                        owner = "",
                        radius = 1,
                        priority = 1,
                    },
                })
                request_lifecycle_melee(5)
            end),
            test.step(1),
            test.run(function()
                expect_lifecycle_melee(1, 115, 22)
            end),
            test.submit({ tick = 6, sequence = 3, player = "alice", kind = "item.weapon_set", payload = { set = 1 } }),
            test.step(2),
            test.run(function()
                expect_sources(0, 0)
                request_lifecycle_melee(8)
            end),
            test.step(1),
            test.run(function()
                expect_lifecycle_melee(2, 65, 13)
            end),
            test.submit({ tick = 9, sequence = 4, player = "alice", kind = "item.weapon_set", payload = { set = 0 } }),
            test.step(2),
            test.run(function()
                expect_sources(50, 40)
                request_lifecycle_melee(11)
            end),
            test.step(1),
            test.run(function()
                expect_lifecycle_melee(3, 115, 22)
            end),
            test.submit(move_item(12, 5, { container = "inventory", x = 2, y = 0 })),
            test.step(2),
            test.run(function()
                expect_sources(0, 0)
                request_lifecycle_melee(14)
            end),
            test.step(1),
            test.run(function()
                expect_lifecycle_melee(4, 65, 13)
            end),
        }),
    },
})
