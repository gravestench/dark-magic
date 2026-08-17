local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

return test.suite({
    name = "monster spawn player scaling",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("pins_effective_count_and_scales_life_and_experience", function(t)
            t:run(function()
                local created
                test.mock_module("engine.authority_command/v1", { register = function() end }, { "register" })
                test.mock_module("engine.authority_random/v1", {
                    integer = function(stream, maximum)
                        test.expect(stream):equals("d2legacy.monster.spawn.life")
                        test.expect(maximum):equals(1)
                        return 0
                    end,
                }, { "integer" })
                test.mock_module("engine.ecs/v1", {
                    query = function(specification)
                        if specification.all[1] == "d2legacy.player.identity" then
                            return { 1, 2 }
                        end
                        return {}
                    end,
                    get = function()
                        return nil
                    end,
                    create = function(components)
                        created = components
                    end,
                }, { "query", "get", "create" })
                test.mock_module("d2legacy.policy.player_count", {
                    monster_spawn = function(actual, evil)
                        test.expect(actual):equals(2)
                        test.expect(evil):equals(true)
                        return {
                            effective_player_count = 2,
                            life_bonus_percent = 50,
                            experience_bonus_percent = 50,
                        }
                    end,
                }, { "monster_spawn" })

                local spawn = fixtures.monster_spawn()
                spawn.definition.evil = true
                spawn.definition.experience = 100
                require("d2legacy.commands.spawn_monster").materialize({ tick = 7, payload = spawn })

                local stats = created["d2legacy.monster.stats"]
                test.expect(stats.spawn_player_count):equals(2)
                test.expect(stats.health):equals(24 * 256)
                test.expect(stats.max_health):equals(24 * 256)
                test.expect(stats.experience):equals(150)
            end)
        end),
    },
})
