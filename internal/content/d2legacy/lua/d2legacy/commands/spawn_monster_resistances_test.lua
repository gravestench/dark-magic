local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

return test.suite({
    name = "monster spawn resistances",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("spawn_stores_all_six_resistances_only_in_authoritative_defense", function(t)
            t:run(function()
                test.mock_module("engine.authority_command/v1", { register = function() end }, { "register" })
                test.mock_module("engine.ecs/v1", {
                    query = function()
                        return {}
                    end,
                    get = function()
                        return nil
                    end,
                    create = function() end,
                }, { "query", "get", "create" })
                test.mock_module("engine.authority_random/v1", {
                    integer = function()
                        return 0
                    end,
                }, { "integer" })
                test.mock_module("d2legacy.policy.player_count", {
                    monster_spawn = function()
                        return { effective_player_count = 1, life_bonus_percent = 0, experience_bonus_percent = 0 }
                    end,
                }, { "monster_spawn" })
                local spawn = fixtures.monster_spawn()
                local values = {
                    physical_resist = 10,
                    magic_resist = 20,
                    fire_resist = 30,
                    lightning_resist = 40,
                    cold_resist = 50,
                    poison_resist = 100,
                }
                for name, value in pairs(values) do
                    spawn.definition[name] = value
                end
                local components =
                    require("d2legacy.commands.spawn_monster").components({ tick = 7, payload = spawn }, 1)
                local defense = components["d2legacy.combat.defense"]
                test.expect(defense ~= nil):equals(true)
                for name, value in pairs(values) do
                    test.expect(defense["base_" .. name]):equals(value)
                    test.expect(defense[name]):equals(value)
                end
                test.expect(components["d2legacy.presentation.character_resistances"]):equals(nil)
            end)
        end),
    },
})
