local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local function player()
    local ecs = require("engine.ecs/v1")
    return ecs.query({ all = { "d2legacy.player.identity" } })[1]
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
    },
})
