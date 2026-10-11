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
