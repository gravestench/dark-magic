local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local spawn = fixtures.monster_spawn()
spawn.definition.life_min = 100 * 256
spawn.definition.life_max = 100 * 256
spawn.definition.physical_resist = 25
spawn.definition.magic_resist = 50
spawn.definition.fire_resist = 80
spawn.definition.lightning_resist = 100
spawn.definition.cold_resist = -25
spawn.definition.poison_resist = 100

return test.suite({
    name = "spawned monster damage mitigation",
    profile = "authority",
    tier = "fast",
    cases = {
        test.case("spawned_resistances_control_all_six_damage_channels", {
            test.submit_system(fixtures.command("system.monster.spawn", spawn, { tick = 1, sequence = 1 })),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local damage = require("d2legacy.policy.damage")
                local monster = ecs.query({ all = { "d2legacy.monster.identity" } })[1]
                local defense = ecs.get(monster, "d2legacy.combat.defense")
                test.expect(defense:get("magic_resist")):equals(50)
                test.expect(defense:get("poison_resist")):equals(100)
                local result = damage.resolve(monster, {
                    physical = 1000,
                    magic = 1000,
                    fire = 1000,
                    lightning = 1000,
                    cold = 1000,
                    poison = 1000,
                }, ecs)
                test.expect(result.mitigated.physical):equals(750)
                test.expect(result.mitigated.magic):equals(500)
                test.expect(result.mitigated.fire):equals(200)
                test.expect(result.mitigated.lightning):equals(0)
                test.expect(result.mitigated.cold):equals(1250)
                test.expect(result.mitigated.poison):equals(0)
                test.expect(result.damage_raw):equals(2700)
                test.expect(result.remaining_health_raw):equals(100 * 256 - 2700)
            end),
        }),
    },
})
