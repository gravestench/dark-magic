local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local function player()
    local ecs = require("engine.ecs/v1")
    return ecs.query({ all = { "d2legacy.player.stamina_progression" } })[1]
end

return test.suite({
    profile = "authority",
    tier = "fast",
    covers = { "internal/game/player" },
    cases = {
        test.case("level_vitality_and_named_sources_resolve_in_record_units", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({
                    level = 10,
                    vitality = 29,
                    stamina = 0,
                    max_stamina = 84,
                    passive_stat_sources = test.array({
                        { id = "vitality", stat = "vitality", value = 5 },
                        { id = "flat", stat = "maxstamina", value = 10 },
                        { id = "active-skill", stat = "skill_staminapercent", value = 25 },
                        { id = "passive-skill", stat = "skill_passive_staminapercent", value = 10 },
                        { id = "per-level", stat = "item_stamina_perlevel", value = 8 },
                    }),
                }),
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local entity = player()
                local basis = ecs.get(entity, "d2legacy.player.stamina_progression")
                local vitals = ecs.get(entity, "d2legacy.player.vitals")
                test.expect(basis:get("vitality")):equals(34)
                test.expect(vitals:get("max_stamina_raw")):equals(39997)
                test.expect(vitals:get("max_stamina")):equals(156)
                test.expect(vitals:get("stamina_raw")):equals(156)
            end),
            test.expect_checkpoint_parity(1),
        }),
        test.case("max_source_activation_rescales_positive_current_stamina", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ stamina = 42, max_stamina = 84 }),
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local entity = player()
                local vitals = ecs.get(entity, "d2legacy.player.vitals")
                vitals:set("stamina", 42)
                vitals:set("stamina_raw", 42 * 256)
                ecs.get(entity, "d2legacy.player.animation"):set("mode", "RN")
                ecs.create({
                    ["d2legacy.stat.source"] = {
                        target = entity,
                        source_id = "test:maxstamina",
                        stat = "maxstamina",
                        operation = "add",
                        value = 10,
                        order = 1,
                    },
                })
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local vitals = ecs.get(player(), "d2legacy.player.vitals")
                test.expect(vitals:get("max_stamina_raw")):equals(94 * 256)
                test.expect(vitals:get("stamina_raw")):equals(47 * 256)
            end),
        }),
        test.case("packed_stamina_bytime_uses_the_checkpointed_act_cycle", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({
                    stamina = 0,
                    max_stamina = 84,
                    passive_stat_sources = test.array({
                        {
                            id = "stamina-bytime",
                            stat = "item_stamina_bytime",
                            value = 4 * ((-10 + 256) + (20 + 256) * 1024),
                        },
                    }),
                }),
            }),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local vitals = ecs.get(player(), "d2legacy.player.vitals")
                test.expect(vitals:get("max_stamina_raw")):equals(104 * 256)
            end),
            test.expect_checkpoint_parity(1),
        }),
        test.case("level_up_refills_the_new_derived_maximum", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ experience = 5, stamina = 10, max_stamina = 84 }),
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local entity = player()
                test.expect(ecs.get(entity, "d2legacy.player.progress"):get("level")):equals(2)
                local vitals = ecs.get(entity, "d2legacy.player.vitals")
                test.expect(vitals:get("max_stamina_raw")):equals(85 * 256)
                test.expect(vitals:get("stamina_raw")):equals(85 * 256)
            end),
        }),
    },
})
