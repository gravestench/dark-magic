local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local function difficulty_row(name, penalty)
    return {
        Name = name,
        ResistPenalty = penalty,
        MonsterFreezeDivisor = "1",
        MonsterColdDivisor = "1",
        DeathExpPenalty = "0",
        LifeStealDivisor = "1",
        ManaStealDivisor = "1",
    }
end

return test.suite({
    name = "character panel resistances",
    profile = "authority",
    tier = "fast",
    initial_data = {
        ["d2legacy.game_rules"] = {
            target = "lod-1.14d",
            expansion = true,
            difficulty = 2,
        },
    },
    -- Owned 1.14d ResistPenalty values: Normal 0, Nightmare -40, Hell -100.
    records = {
        ["data/global/excel/difficultylevels.txt"] = {
            difficulty_row("Normal", "0"),
            difficulty_row("Nightmare", "-40"),
            difficulty_row("Hell", "-100"),
        },
    },
    cases = {
        test.case("values_are_limited_to_each_maximum", function(t)
            t:run(function()
                local resistances = require("d2legacy.ui.character_resistances")
                local shown = resistances.displayed({
                    fire_resist = 90,
                    max_fire_resist = 75,
                    cold_resist = -100,
                    max_cold_resist = 75,
                    lightning_resist = 80,
                    max_lightning_resist = 85,
                })
                test.expect(shown.fire):equals(75)
                test.expect(shown.cold):equals(-100)
                test.expect(shown.lightning):equals(80)
                test.expect(shown.poison):equals(nil)
            end)
        end),
        test.case("connected_mirror_displays_current_owner_values", function(t)
            t:run(function()
                local ecs = require("engine.ecs/v1")
                local resistances = require("d2legacy.ui.character_resistances")
                local owner = ecs.create({
                    ["d2legacy.player.identity"] = { player = "connected-owner" },
                    ["d2legacy.presentation.character_resistances"] = {
                        fire_resist = 90,
                        cold_resist = -100,
                        lightning_resist = 31,
                        max_fire_resist = 75,
                        max_cold_resist = 80,
                        max_lightning_resist = 85,
                    },
                })
                ecs.create({
                    ["d2legacy.player.identity"] = { player = "peer" },
                    ["d2legacy.presentation.character_resistances"] = {
                        fire_resist = 999,
                        cold_resist = 999,
                        lightning_resist = 999,
                        max_fire_resist = 999,
                        max_cold_resist = 999,
                        max_lightning_resist = 999,
                    },
                })
                test.expect(resistances.for_player("connected-owner"))
                    :deep_equals({ fire = 75, cold = -100, lightning = 31 })
                local projected = ecs.get(owner, "d2legacy.presentation.character_resistances")
                projected:set("fire_resist", -40)
                projected:set("max_fire_resist", 95)
                test.expect(resistances.for_player("connected-owner"))
                    :deep_equals({ fire = -40, cold = -100, lightning = 31 })
            end)
        end),
        test.case("hell_penalty_reaches_the_panel_for_the_admitted_player", {
            test.submit_system(fixtures.command("system.player.enter", fixtures.player_entry(), {
                tick = 1,
                sequence = 1,
            })),
            test.step(1),
            test.run(function()
                local resistances = require("d2legacy.ui.character_resistances")
                test.expect(resistances.for_player("alice")):deep_equals({ fire = -100, cold = -100, lightning = -100 })
                test.expect(resistances.for_player("bob")):equals(nil)
            end),
        }),
    },
})
