local test = require("d2legacy.tests/v1")

local fixtures = require("d2legacy.tests.support.fixtures")
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
                Param1 = "3",
            },
        },
        ["data/global/excel/skilldesc.txt"] = {
            { skilldesc = "firebolt", ListRow = "0", IconCel = "0" },
            { skilldesc = "frozenarmor", ListRow = "1", IconCel = "1" },
        },
    },
    cases = {
        test.case("targetless_non_damage_skill_applies_self_state", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.amazon_entry,
            }),
            test.step(1),
            test.submit({
                tick = 2,
                sequence = 1,
                player = "alice",
                kind = "player.use_skill",
                payload = { side = "left" },
            }),
            test.step(1),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local instances = ecs.query({ all = { "d2legacy.state.instance" } })
                test.assert(#instances == 1, [=[#instances == 1]=])
                local state = ecs.get(instances[1], "d2legacy.state.instance")
                test.assert(
                    state:get("target"):id() == player:id()
                        and state:get("state_id") == "frozen_armor"
                        and state:get("source_id") == "skill:alice:40"
                )
                local events = ecs.query({ all = { "d2legacy.skill.cast_event" } })
                test.assert(
                    #events == 1 and ecs.get(events[1], "d2legacy.skill.cast_event"):get("behavior") == "state.self"
                )
            end),
        }),
    },
})
