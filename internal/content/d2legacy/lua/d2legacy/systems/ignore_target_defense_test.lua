local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local function player_entry(ignore_target_ac, defense)
    return fixtures.player_entry({
        level = 20,
        -- The synthetic Amazon record has ToHitFactor 0: 235 + 5 * (20 - 7) = 300.
        base_attack_rating = 235,
        defense = defense or 0,
        passive_stat_sources = {
            { id = "ignore-defense", stat = "item_ignoretargetac", operation = "add", value = ignore_target_ac },
        },
    })
end

local function spawn_target(classification)
    local ecs = require("engine.ecs/v1")
    local monster = ecs.create({
        ["d2legacy.monster.stats"] = {
            level = 20,
            health = 4096,
            max_health = 4096,
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
            id = "monster:target",
            kind = "hostile",
            label = "Target",
            owner = "",
            radius = 1,
            priority = 1,
        },
    })
    if classification == "boss" then
        ecs.set(monster, "d2legacy.monster.boss", {})
    elseif classification == "hireling" then
        -- A hostile classification reaches current targeting. The active
        -- hireling relation supplies only the HIT-R02c exclusion predicate.
        ecs.set(monster, "d2legacy.owned_unit", {
            owner_id = "player:alice",
            ultimate_owner_id = "player:alice",
            category = "hireable",
            group = 0,
            limit = 1,
            replacement = "replace_oldest",
            created_tick = 1,
            active = true,
        })
    end
    return monster
end

local function result()
    local ecs = require("engine.ecs/v1")
    local events = test.entities_with("d2legacy.combat.melee_event")
    test.expect(#events):equals(1)
    return ecs.get(events[1], "d2legacy.combat.melee_event"), ecs.get(events[1], "d2legacy.combat.attack_result")
end

local function player_case(name, ignore_target_ac, classification, defense, chance, hit)
    return test.case(name, {
        test.submit_system({
            tick = 1,
            sequence = 1,
            kind = "system.player.enter",
            payload = player_entry(ignore_target_ac),
        }),
        test.step(1),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            local player = test.only_entity_with("d2legacy.player.identity")
            local target = spawn_target(classification)
            test.expect(ecs.get(player, "d2legacy.player.progress"):get("level")):equals(20)
            test.expect(ecs.get(player, "d2legacy.player.combat_stats"):get("attack_rating")):equals(300)
            test.expect(ecs.get(target, "d2legacy.monster.stats"):get("level")):equals(20)
            ecs.set(player, "d2legacy.combat.basic_attack_request", {
                target_id = "monster:target",
                request_tick = 2,
                hand = "rarm",
            })
        end),
        test.step(1),
        test.run(function()
            local melee, attack = result()
            test.expect(melee:get("attack_rating")):equals(300)
            test.expect(melee:get("defense")):equals(defense)
            test.expect(melee:get("hit_chance")):equals(chance)
            test.expect(melee:get("hit")):equals(hit)
            test.expect(attack:get("defense")):equals(defense)
            test.expect(attack:get("hit_chance")):equals(chance)
            test.expect(attack:get("outcome")):equals(hit and "hit" or "miss")
        end),
    })
end

return test.suite({
    profile = "authority",
    tier = "integration",
    seed = 36,
    cases = {
        player_case("ordinary_monster_ignores_defense_for_chance_and_draw", 1, nil, 0, 95, true),
        player_case("zero_stat_keeps_monster_defense", 0, nil, 500, 37, false),
        player_case("boss_keeps_monster_defense", 1, "boss", 500, 37, false),
        player_case("active_hireling_classification_keeps_monster_defense", 1, "hireling", 500, 37, false),
        test.case("monster_attacker_with_stat_does_not_ignore_player_defense", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = player_entry(0, 495),
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = test.only_entity_with("d2legacy.player.identity")
                local monster = spawn_target()
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
                        source_id = "monster:ignore-defense",
                        stat = "item_ignoretargetac",
                        operation = "add",
                        value = 1,
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
                local melee, attack = result()
                test.expect(melee:get("attack_rating")):equals(300)
                test.expect(melee:get("defense")):equals(500)
                test.expect(melee:get("hit_chance")):equals(37)
                test.expect(melee:get("hit")):equals(false)
                test.expect(attack:get("defense")):equals(500)
                test.expect(attack:get("outcome")):equals("miss")
            end),
        }),
    },
})
