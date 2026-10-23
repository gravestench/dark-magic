local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

return test.suite({
    name = "monster spawn type flags",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("spawn_publishes_authored_demon_and_undead_flags", function(t)
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
                spawn.definition.demon = true
                spawn.definition.low_undead = false
                spawn.definition.high_undead = true
                local materializer = require("d2legacy.commands.spawn_monster")
                local components = materializer.components({ tick = 1, payload = spawn }, 1)
                local flags = components["d2legacy.monster.type_flags"]
                test.expect(flags.demon):equals(true)
                test.expect(flags.low_undead):equals(false)
                test.expect(flags.high_undead):equals(true)
                spawn.definition.demon = nil
                spawn.definition.low_undead = nil
                spawn.definition.high_undead = nil
                local absent = materializer.components({ tick = 2, payload = spawn }, 1)
                test.expect(absent["d2legacy.monster.type_flags"].demon):equals(false)
                test.expect(absent["d2legacy.monster.type_flags"].low_undead):equals(false)
                test.expect(absent["d2legacy.monster.type_flags"].high_undead):equals(false)
            end)
        end),
    },
})
