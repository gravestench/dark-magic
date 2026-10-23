local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local function player()
    local ecs = require("engine.ecs/v1")
    return ecs.query({ all = { "d2legacy.player.identity" } })[1]
end

local function mana_tick_case(name, inputs)
    return test.case(name, {
        test.submit_system({
            tick = 1,
            sequence = 1,
            kind = "system.player.enter",
            payload = fixtures.player_entry({
                character_id = "hero-mana-case",
                player = "mana-case",
                name = "mana-case",
                class = inputs.class or "Amazon",
                mana = 0,
                max_mana = 300,
            }),
        }),
        test.step(2),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            local owner = player()
            local vitals = ecs.get(owner, "d2legacy.player.vitals")
            vitals:set("mana_raw", inputs.current)
            vitals:set("max_mana_raw", inputs.maximum)
            if inputs.bonus then
                ecs.create({
                    ["d2legacy.stat.source"] = {
                        target = owner,
                        source_id = "fixture:mana-percent",
                        stat = "manarecoverybonus",
                        operation = "add",
                        value = inputs.bonus,
                        order = 1,
                    },
                })
            end
            if inputs.flat then
                ecs.create({
                    ["d2legacy.stat.source"] = {
                        target = owner,
                        source_id = "fixture:mana-flat",
                        stat = "manarecovery",
                        operation = "add",
                        value = inputs.flat,
                        order = 2,
                    },
                })
            end
            if inputs.suppressed then
                ecs.create({
                    ["d2legacy.resource.mana_regen_suppression"] = {
                        target = owner,
                        source_id = "fixture:mana-suppression",
                    },
                })
            end
        end),
        test.step(1),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            test.expect(ecs.get(player(), "d2legacy.player.vitals"):get("mana_raw")):equals(inputs.expected)
        end),
    })
end

return test.suite({
    profile = "authority",
    tier = "fast",
    records = {
        ["data/global/excel/charstats.txt"] = {
            {
                class = "Amazon",
                StartSkill = "Fixture Idle",
                ToHitFactor = "5",
                WalkVelocity = "6",
                RunVelocity = "9",
                stamina = "84",
                vit = "20",
                RunDrain = "20",
                StaminaPerLevel = "4",
                StaminaPerVitality = "4",
                ManaRegen = "120",
            },
            {
                class = "Sorceress",
                StartSkill = "Fixture Idle",
                ToHitFactor = "5",
                WalkVelocity = "6",
                RunVelocity = "9",
                stamina = "84",
                vit = "20",
                RunDrain = "20",
                StaminaPerLevel = "4",
                StaminaPerVitality = "4",
                ManaRegen = "0",
            },
        },
        ["data/global/excel/skills.txt"] = {
            {
                Id = "997",
                skill = "Fixture Idle",
                skilldesc = "idle",
                general = "1",
                leftskill = "1",
                passive = "0",
            },
        },
        ["data/global/excel/skilldesc.txt"] = {
            { skilldesc = "idle", ListRow = "0", IconCel = "0" },
        },
    },
    cases = {
        test.case("regenerates_fixed_point_mana_from_class_and_named_stat_facts", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({
                    character_id = "hero-alice",
                    player = "alice",
                    name = "alice",
                    class = "Amazon",
                    mana = 0,
                    max_mana = 300,
                }),
            }),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local owner = player()
                local vitals = ecs.get(owner, "d2legacy.player.vitals")
                vitals:set("mana_raw", 0)
                vitals:set("mana", 0)
                test.expect(ecs.get(owner, "d2legacy.player.resource_stats"):get("mana_regen_frames")):equals(120)
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local owner = player()
                test.expect(ecs.get(owner, "d2legacy.player.vitals"):get("mana_raw")):equals(25)
                ecs.create({
                    ["d2legacy.stat.source"] = {
                        target = owner,
                        source_id = "fixture:mana-bonus",
                        stat = "manarecoverybonus",
                        operation = "add",
                        value = 300,
                        order = 1,
                    },
                })
                ecs.create({
                    ["d2legacy.stat.source"] = {
                        target = owner,
                        source_id = "fixture:flat-mana",
                        stat = "manarecovery",
                        operation = "add",
                        value = 7,
                        order = 2,
                    },
                })
                local vitals = ecs.get(owner, "d2legacy.player.vitals")
                vitals:set("mana_raw", 0)
                vitals:set("mana", 0)
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local owner = player()
                test.expect(ecs.get(owner, "d2legacy.player.vitals"):get("mana_raw")):equals(107)
                ecs.create({
                    ["d2legacy.resource.mana_regen_suppression"] = {
                        target = owner,
                        source_id = "fixture:suppressed",
                    },
                })
                local vitals = ecs.get(owner, "d2legacy.player.vitals")
                vitals:set("mana_raw", 0)
                vitals:set("mana", 0)
            end),
            test.step(1),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local owner = player()
                local vitals = ecs.get(owner, "d2legacy.player.vitals")
                test.expect(vitals:get("mana_raw")):equals(7)
                local suppression = ecs.query({ all = { "d2legacy.resource.mana_regen_suppression" } })[1]
                ecs.destroy(suppression)
                vitals:set("mana_raw", vitals:get("max_mana_raw") - 3)
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local vitals = ecs.get(player(), "d2legacy.player.vitals")
                test.expect(vitals:get("mana_raw")):equals(vitals:get("max_mana_raw"))
                test.expect(vitals:get("mana")):equals(300)
            end),
        }),
        mana_tick_case("flat_recovery_wraps_before_lower_mana_clamp", {
            current = 100,
            maximum = 25600,
            flat = 2147483647,
            expected = 0,
        }),
        mana_tick_case("percent_factor_wraps_before_mana_multiply", {
            current = 0,
            maximum = 25600,
            bonus = 2147483647,
            expected = 7,
        }),
        mana_tick_case("mana_recovery_caps_at_maximum", {
            current = 25599,
            maximum = 25600,
            expected = 25600,
        }),
        mana_tick_case("negative_mana_recovery_stops_at_zero", {
            current = 1,
            maximum = 25600,
            bonus = -130,
            expected = 0,
        }),
        mana_tick_case("suppression_keeps_flat_mana_recovery", {
            current = 1000,
            maximum = 25600,
            bonus = 999,
            flat = 5,
            suppressed = true,
            expected = 1005,
        }),
        mana_tick_case("zero_charstats_mana_regen_uses_fallback", {
            class = "Sorceress",
            current = 0,
            maximum = 25600,
            expected = 3,
        }),
        test.case("mana_percent_uses_original_divide_first_operand_order", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({
                    character_id = "hero-order",
                    player = "order",
                    name = "order",
                    class = "Amazon",
                    mana = 0,
                    max_mana = 300,
                }),
            }),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local owner = player()
                local vitals = ecs.get(owner, "d2legacy.player.vitals")
                vitals:set("mana_raw", 0)
                vitals:set("max_mana_raw", 30000000)
                ecs.create({
                    ["d2legacy.stat.source"] = {
                        target = owner,
                        source_id = "fixture:large-mana-percent",
                        stat = "manarecoverybonus",
                        operation = "add",
                        value = 65437,
                        order = 1,
                    },
                })
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(ecs.get(player(), "d2legacy.player.vitals"):get("mana_raw")):equals(6550000)
            end),
        }),
    },
})
