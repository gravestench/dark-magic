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
                mana = "7",
                manashift = "8",
                Param1 = "30",
                Param2 = "5",
                Param3 = "3000",
                Param4 = "300",
                Param7 = "250",
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
        }),
    },
})
