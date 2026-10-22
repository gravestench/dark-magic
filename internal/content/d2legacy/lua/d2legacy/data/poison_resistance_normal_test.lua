local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

return test.suite({
    name = "normal player poison resistance",
    profile = "authority",
    tier = "fast",
    initial_data = {
        ["d2legacy.game_rules"] = {
            target = "lod-1.14d",
            expansion = true,
            difficulty = 0,
        },
    },
    records = {
        ["data/global/excel/difficultylevels.txt"] = {
            {
                Name = "Normal",
                ResistPenalty = "0",
                MonsterFreezeDivisor = "1",
                MonsterColdDivisor = "1",
                DeathExpPenalty = "0",
                LifeStealDivisor = "1",
                ManaStealDivisor = "1",
            },
        },
    },
    cases = {
        test.case("fresh_player_has_zero_poison_resistance_in_authority_and_panel", {
            test.submit_system(fixtures.command("system.player.enter", fixtures.player_entry(), {
                tick = 1,
                sequence = 1,
            })),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local resistances = require("d2legacy.ui.character_resistances")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local defense = ecs.get(player, "d2legacy.combat.defense")
                test.expect(defense:get("poison_resist")):equals(0)
                test.expect(defense:get("max_poison_resist")):equals(75)
                test.expect(resistances.for_player("alice").poison):equals(0)
            end),
        }),
    },
})
