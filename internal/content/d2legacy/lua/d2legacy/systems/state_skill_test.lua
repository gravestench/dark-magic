local test = require("d2legacy.tests/v1")

local fixtures = require("d2legacy.tests.support.fixtures")
local player_entry = fixtures.player_entry({ defense = 100 })
return test.suite({
    profile = "authority",
    tier = "fast",
    records = {
        ["data/global/excel/charstats.txt"] = {
            { class = "Amazon", StartSkill = "Frozen Armor" },
        },
        ["data/global/excel/skills.txt"] = {
            {
                Id = "36",
                skill = "Fire Bolt",
                srvmissile = "firebolt",
                skilldesc = "firebolt",
                leftskill = "1",
                general = "0",
                passive = "0",
                etype = "fire",
                interrupt = "1",
                srvstfunc = "",
                srvdofunc = "",
                mana = "5",
                manashift = "7",
                emin = "3",
                emax = "6",
                HitShift = "8",
            },
            {
                Id = "40",
                skill = "Frozen Armor",
                skilldesc = "frozenarmor",
                leftskill = "1",
                general = "0",
                passive = "0",
                srvstfunc = "",
                srvdofunc = "18",
                aurastate = "frozenarmor",
                auralencalc = "ln34+(skill('Shiver Armor'.blvl)+skill('Chilling Armor'.blvl))*par7",
                aurastat1 = "skill_armor_percent",
                aurastatcalc1 = "ln12",
                auraevent1 = "damagedinmelee",
                auraeventfunc1 = "2",
                calc1 = "ln56*(100+((skill('Shiver Armor'.blvl)+skill('Chilling Armor'.blvl))*par8))/100",
                mana = "7",
                manashift = "8",
                Param1 = "30",
                Param2 = "5",
                Param3 = "3000",
                Param4 = "300",
                Param5 = "30",
                Param6 = "3",
                Param7 = "250",
                Param8 = "5",
            },
            {
                Id = "50",
                skill = "Shiver Armor",
                skilldesc = "shiverarmor",
                leftskill = "1",
                general = "1",
                passive = "0",
            },
            {
                Id = "60",
                skill = "Chilling Armor",
                skilldesc = "chillingarmor",
                leftskill = "1",
                general = "1",
                passive = "0",
            },
        },
        ["data/global/excel/skilldesc.txt"] = {
            { skilldesc = "firebolt", ListRow = "0", IconCel = "0" },
            { skilldesc = "frozenarmor", ListRow = "1", IconCel = "1" },
            { skilldesc = "shiverarmor", ListRow = "2", IconCel = "2" },
            { skilldesc = "chillingarmor", ListRow = "3", IconCel = "3" },
        },
        ["data/global/excel/states.txt"] = {
            { state = "frozenarmor", id = "10", group = "1" },
            { state = "freeze", id = "1", group = "0" },
        },
    },
    cases = {
        test.case("targetless_non_damage_skill_applies_self_state", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = player_entry,
            }),
            test.step(1),
            test.submit({
                tick = 2,
                sequence = 1,
                player = "alice",
                kind = "player.use_skill",
                payload = { side = "left" },
            }),
            test.step(3),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local instances = ecs.query({ all = { "d2legacy.state.instance" } })
                test.assert(#instances == 1, [=[#instances == 1]=])
                local state = ecs.get(instances[1], "d2legacy.state.instance")
                test.assert(
                    state:get("target"):id() == player:id()
                        and state:get("state_id") == "frozenarmor"
                        and state:get("source_id") == "skill:alice:40"
                        and state:get("expires_tick") == 3503
                )
                local sources = ecs.query({ all = { "d2legacy.stat.source" } })
                test.assert(#sources == 1, [=[#sources == 1]=])
                local source = ecs.get(sources[1], "d2legacy.stat.source")
                test.assert(
                    source:get("stat") == "defense"
                        and source:get("operation") == "percent"
                        and source:get("value") == 30
                )
                local combat = ecs.get(player, "d2legacy.player.combat_stats")
                local vitals = ecs.get(player, "d2legacy.player.vitals")
                test.assert(combat:get("defense") == 136, [=[combat:get("defense") == 136]=])
                test.assert(vitals:get("mana_raw") == 3328 and vitals:get("mana") == 13)
                local events = ecs.query({ all = { "d2legacy.skill.cast_event" } })
                test.assert(
                    #events == 1
                        and ecs.get(events[1], "d2legacy.skill.cast_event"):get("behavior")
                            == "state.self-timed-stat"
                )
            end),
            test.submit({
                tick = 5,
                sequence = 1,
                player = "alice",
                kind = "player.use_skill",
                payload = { side = "left" },
            }),
            test.step(3),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local instances = ecs.query({ all = { "d2legacy.state.instance" } })
                local sources = ecs.query({ all = { "d2legacy.stat.source" } })
                test.assert(#instances == 1 and #sources == 1, [=[#instances == 1 and #sources == 1]=])
                test.expect(ecs.get(instances[1], "d2legacy.state.instance"):get("expires_tick")):equals(3506)
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local vitals = ecs.get(player, "d2legacy.player.vitals")
                test.assert(vitals:get("mana_raw") == 1536 and vitals:get("mana") == 6)
                test.assert(
                    #ecs.query({ all = { "d2legacy.state.event" } }) == 2,
                    [=[#ecs.query({ all = { "d2legacy.state.event" } }) == 2]=]
                )
            end),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local monster = ecs.create({
                    ["d2legacy.monster.stats"] = {},
                    ["d2legacy.monster.ai"] = {
                        behavior = "fixture",
                        state = "attack",
                        next_think_tick = 9999,
                        think_interval = 100,
                        aggro_radius = 0,
                        attack_range = 1,
                        speed = 5,
                    },
                    ["d2legacy.world.velocity"] = { x = 5, y = 0 },
                    ["d2legacy.world.selectable"] = {
                        id = "monster:armor-attacker",
                        kind = "hostile",
                        label = "Armor Attacker",
                    },
                })
                ecs.create({
                    ["d2legacy.combat.melee_event"] = {
                        kind = "hit_resolved",
                        tick = 8,
                        attacker_id = "monster:armor-attacker",
                        target_id = "player:alice",
                        hit = true,
                        damage_raw = 256,
                        remaining_health_raw = 49 * 256,
                        hand = "hth",
                    },
                })
                test.assert(monster ~= nil, [=[monster ~= nil]=])
            end),
            test.step(2),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local monster
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.world.selectable" } })) do
                    if ecs.get(entity, "d2legacy.world.selectable"):get("id") == "monster:armor-attacker" then
                        monster = entity
                    end
                end
                test.assert(monster ~= nil, [=[monster ~= nil]=])
                local freeze
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.state.instance" } })) do
                    local instance = ecs.get(entity, "d2legacy.state.instance")
                    if instance:get("state_id") == "freeze" then
                        freeze = instance
                    end
                end
                test.assert(freeze ~= nil, [=[freeze ~= nil]=])
                test.assert(
                    freeze:get("target"):id() == monster:id()
                        and freeze:get("expires_tick") - freeze:get("applied_tick") == 33
                        and freeze:get("action_disabled"),
                    [=[freeze targets the attacker for the target-derived duration]=]
                )
                local velocity = ecs.get(monster, "d2legacy.world.velocity")
                test.assert(velocity:get("x") == 0 and velocity:get("y") == 0)
                test.expect(ecs.get(monster, "d2legacy.monster.ai"):get("state")):equals("disabled")
                local events = ecs.query({ all = { "d2legacy.combat.melee_event" } })
                test.assert(#events == 1, [=[#events == 1]=])
                test.assert(ecs.get(events[1], "d2legacy.combat.melee_event"):get("defender_effects_processed"))
            end),
        }),
    },
})
