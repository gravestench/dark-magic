local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local function enter(health, class, vitality)
    return test.submit_system({
        tick = 1,
        sequence = 1,
        kind = "system.player.enter",
        payload = fixtures.player_entry({
            health = health or 1,
            max_health = 200,
            class = class or "Amazon",
            vitality = vitality or 20,
        }),
    })
end

local function use(tick, sequence, id, player)
    return test.submit({
        tick = tick,
        sequence = sequence,
        player = player or "alice",
        kind = "item.use",
        payload = { item_id = id or "first" },
    })
end

local function player()
    local ecs = require("engine.ecs/v1")
    return ecs.query({ all = { "d2legacy.player.identity" } })[1]
end

local function health()
    local vitals = require("engine.ecs/v1").get(player(), "d2legacy.player.vitals")
    return vitals:get("health") * 256 - vitals:get("poison_remainder_raw")
end

local function effects()
    return require("engine.ecs/v1").query({ all = { "d2legacy.player.potion" } })
end

local function item(id)
    local find = require("d2legacy.items.find")
    return find.item(find.layout("alice"), id)
end

local function expect_supply(first, second, gold)
    test.expect(item("first") ~= nil):equals(first)
    test.expect(item("second") ~= nil):equals(second)
    local _, layout = require("d2legacy.items.find").layout("alice")
    test.expect(layout:get("carried_gold")):equals(gold or 1000)
end

local function purchase_vector()
    -- Two fixed holdouts: the rounding fixture and the mounted hp1 record.
    for _, row in ipairs(require("engine.records/v1").load("data/global/excel/misc.txt")) do
        if row.code == "hp1" then
            local vectors = {
                ["175"] = { first = 38465, final = 49775 },
                ["192"] = { first = 38460, final = 49920 },
            }
            return test.assert(vectors[row.len], "unqualified hp1 duration")
        end
    end
    error("hp1 record missing")
end

return test.suite({
    profile = "authority",
    tier = "integration",
    records = {
        ["data/global/excel/npc.txt"] = { { npc = "akara", ["sell mult"] = "1024" } },
        ["data/global/excel/misc.txt"] = {
            {
                code = "hp1",
                pSpell = "3",
                stat1 = "hpregen",
                state = "healthpot",
                calc1 = "30",
                len = "175",
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
            inventory_width = 10,
            inventory_height = 4,
            stash_width = 6,
            stash_height = 8,
            belt_capacity = 4,
            vendor_width = 10,
            vendor_height = 10,
            carried_gold = 1000,
            items = {
                { id = "first", code = "hp1", width = 1, height = 1, container = "inventory", x = 0, y = 0 },
                { id = "second", code = "hp1", width = 1, height = 1, container = "belt", belt_slot = 0 },
                { id = "stored", code = "hp1", width = 1, height = 1, container = "stash", x = 0, y = 0 },
                { id = "unsupported", code = "mp1", width = 1, height = 1, container = "inventory", x = 1, y = 0 },
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
        test.case("record_rate_not_instant_or_exact_budget_and_duplicate_preserves_spare", {
            enter(),
            test.step(1),
            use(2, 1),
            use(2, 2),
            test.step(1),
            test.run(function()
                test.expect(health()):equals(321) -- 1 life + floor(45*256/175), not 45 life immediately.
                expect_supply(false, true)
                test.expect(#effects()):equals(1)
            end),
            test.step(73),
            test.expect_checkpoint_parity(50),
            test.restore_checkpoint(),
            test.step(51),
            test.run(function()
                test.expect(health()):equals(11631) -- 175 frames at 65 raw, including the admission frame.
                test.expect(#effects()):equals(0)
                expect_supply(false, true)
            end),
            use(177, 1),
            test.step(1),
            test.run(function()
                test.expect(health()):equals(11631)
            end),
        }),
        test.case("overlap_extends_duration_at_weighted_rate_not_double_speed", {
            enter(),
            test.step(1),
            use(2, 1),
            test.step(25),
            use(27, 2, "second"),
            test.step(1),
            test.run(function()
                test.expect(health()):equals(1946) -- 26 frames at 65 raw.
                test.expect(#effects()):equals(1)
                local effect = require("engine.ecs/v1").get(effects()[1], "d2legacy.player.potion")
                test.expect(effect:get("remaining_ticks")):equals(324)
                expect_supply(false, false)
            end),
            test.expect_checkpoint_parity(20),
            test.restore_checkpoint(),
            test.step(304),
            test.run(function()
                test.expect(health()):equals(23006) -- 350 total frames, no remainder redistribution.
                test.expect(#effects()):equals(0)
            end),
        }),
        test.case("same_tick_distinct_drinks_retire_both_but_keep_one_rate", {
            enter(),
            test.step(1),
            use(2, 1),
            use(2, 2, "second"),
            test.step(1),
            test.run(function()
                test.expect(health()):equals(321)
                expect_supply(false, false)
                test.expect(#effects()):equals(1)
            end),
            test.expect_checkpoint_parity(3),
        }),
        test.case("wrong_owner_stash_missing_unsupported_and_held_refuse_without_healing", {
            enter(),
            test.step(1),
            use(2, 1, "first", "bob"),
            use(2, 1, "stored"),
            use(2, 2, "unsupported"),
            use(2, 3, "missing"),
            test.step(1),
            test.run(function()
                test.expect(health()):equals(256)
                test.expect(#effects()):equals(0)
                expect_supply(true, true)
                test.assert(item("stored") and item("unsupported"))
            end),
            test.submit({
                tick = 3,
                sequence = 4,
                player = "alice",
                kind = "item.move",
                payload = { item_id = "unsupported", destination = { container = "held" } },
            }),
            test.step(1),
            use(4, 5),
            test.step(1),
            test.run(function()
                test.expect(health()):equals(256)
                expect_supply(true, true)
                test.expect(#effects()):equals(0)
            end),
        }),
        test.case("dead_player_and_active_effect_cannot_resurrect", {
            enter(),
            test.step(1),
            use(2, 1),
            test.step(1),
            test.run(function()
                -- A declared lethal strike uses the production damage owner, not a health write.
                local result = require("d2legacy.policy.damage").resolve(
                    player(),
                    require("d2legacy.policy.damage_bundle").single("physical", 1000 * 256),
                    require("engine.ecs/v1")
                )
                test.assert(result.lethal)
            end),
            use(3, 2, "second"),
            test.step(1),
            test.run(function()
                test.expect(health()):equals(0)
                test.expect(#effects()):equals(0)
                expect_supply(false, true)
            end),
            test.expect_checkpoint_parity(2),
        }),
        test.case("ceiling_discards_effect_and_does_not_bank_later_healing", {
            enter(199),
            test.step(1),
            use(2, 1),
            test.step(4),
            test.run(function()
                test.expect(health()):equals(51200)
                test.expect(#effects()):equals(0)
                expect_supply(false, true)
                require("d2legacy.policy.damage").resolve(
                    player(),
                    require("d2legacy.policy.damage_bundle").single("physical", 5 * 256),
                    require("engine.ecs/v1")
                )
            end),
            test.step(10),
            test.run(function()
                test.expect(health()):equals(49920)
            end),
        }),
        test.case("purchased_identity_heals_combat_damage_and_retires_once", {
            enter(200),
            test.step(1),
            test.submit({
                tick = 2,
                sequence = 1,
                player = "alice",
                kind = "interaction.open",
                payload = { target = "npc:akara" },
            }),
            test.step(1),
            test.submit({
                tick = 3,
                sequence = 2,
                player = "alice",
                kind = "interaction.choose",
                payload = { target = "npc:akara", choice = "trade", revision = 1 },
            }),
            test.step(1),
            test.submit({
                tick = 4,
                sequence = 3,
                player = "alice",
                kind = "item.vendor_buy",
                payload = { vendor = "akara", item_id = "akara-offer:hp1", to_inventory = true },
            }),
            test.step(1),
            test.run(function()
                test.assert(item("akara-purchase:hp1:1"))
                local result = require("d2legacy.policy.damage").resolve(
                    player(),
                    require("d2legacy.policy.damage_bundle").single("physical", 50 * 256),
                    require("engine.ecs/v1")
                )
                test.expect(result.damage_raw):equals(12800)
                test.expect(health()):equals(38400)
            end),
            use(5, 4, "akara-purchase:hp1:1"),
            use(5, 5, "akara-purchase:hp1:1"),
            test.step(1),
            test.run(function()
                test.assert(not item("akara-purchase:hp1:1"))
                test.expect(health()):equals(purchase_vector().first)
                expect_supply(true, true, 970)
            end),
            test.expect_checkpoint_parity(20),
            test.restore_checkpoint(),
            test.step(171),
            test.run(function()
                test.expect(health()):equals(purchase_vector().final)
                test.expect(#effects()):equals(0)
                expect_supply(true, true, 970)
            end),
        }),
        test.case("leave_removes_active_effect_without_dangling_checkpoint_references", {
            enter(),
            test.step(1),
            use(2, 1),
            test.step(1),
            test.submit_system({ tick = 3, sequence = 2, kind = "system.player.leave", payload = { player = "alice" } }),
            test.step(1),
            test.restore_checkpoint(),
            test.step(1),
            test.run(function()
                test.expect(#effects()):equals(0)
            end),
        }),
        test.case("poison_and_potion_share_fractional_health_without_rounding_each_tick", {
            enter(),
            test.step(1),
            test.run(function()
                require("engine.ecs/v1").create({
                    ["d2legacy.state.instance"] = {
                        target = player(),
                        state_id = "poison",
                        source_id = "test:poison",
                        applied_tick = 1,
                        expires_tick = 11,
                        policy = "refresh_same_source",
                        stat = "hpregen",
                        stat_value = -64,
                    },
                })
                require("engine.ecs/v1").get(player(), "d2legacy.world.location"):set("level_id", 2)
            end),
            use(2, 1),
            test.step(10),
            test.run(function()
                test.expect(health()):equals(266)
            end),
            test.expect_checkpoint_parity(2),
        }),
        test.case("seeded_high_vitality_critical_doubles_rate_and_checkpoint_continues", {
            enter(1, "Amazon", 10000),
            test.step(1),
            use(2, 1),
            test.step(1),
            test.run(function()
                test.expect(health()):equals(387)
            end),
            test.expect_checkpoint_parity(10),
            test.restore_checkpoint(),
            test.run(function()
                test.expect(health()):equals(1697)
            end),
        }),
        test.case("critical_overlap_weights_unspent_rate_and_discards_division_remainder", {
            enter(),
            test.step(1),
            use(2, 1),
            test.step(25),
            test.run(function()
                require("engine.ecs/v1").create({
                    ["d2legacy.stat.source"] = {
                        target = player(),
                        source_id = "test:vitality",
                        stat = "vitality",
                        operation = "add",
                        value = 9980,
                    },
                })
            end),
            use(27, 2, "second"),
            test.step(1),
            test.run(function()
                test.expect(health()):equals(1981)
                test.expect(#effects()):equals(1)
            end),
            test.expect_checkpoint_parity(20),
            test.restore_checkpoint(),
            test.step(304),
            test.run(function()
                test.expect(health()):equals(34381)
                test.expect(#effects()):equals(0)
            end),
        }),
        test.case("inactive_and_death_marked_player_preserve_the_selected_item", {
            enter(),
            test.step(1),
            test.run(function()
                require("engine.ecs/v1").set(player(), "d2legacy.world.inactive", {})
            end),
            use(2, 1),
            test.step(1),
            test.run(function()
                test.expect(health()):equals(256)
                expect_supply(true, true)
                local ecs = require("engine.ecs/v1")
                ecs.remove(player(), "d2legacy.world.inactive")
                ecs.set(player(), "d2legacy.player.death", { stage = "dead" })
            end),
            use(3, 2),
            test.step(1),
            test.run(function()
                test.expect(health()):equals(256)
                expect_supply(true, true)
                test.expect(#effects()):equals(0)
            end),
        }),
    },
})
