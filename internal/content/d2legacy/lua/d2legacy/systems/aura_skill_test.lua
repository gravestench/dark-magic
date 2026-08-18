local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local function entry(player, class, x)
    return fixtures.player_entry({
        character_id = "hero-" .. player,
        player = player,
        name = player,
        class = class,
        x = x,
        y = 12,
        defense = 100,
        vitality = class == "Paladin" and 25 or 20,
    })
end

local function aura_effects()
    local ecs = require("engine.ecs/v1")
    return ecs.query({ all = { "d2legacy.skill.aura_effect", "d2legacy.stat.source" } })
end

return test.suite({
    profile = "authority",
    tier = "fast",
    records = {
        ["data/global/excel/charstats.txt"] = {
            {
                class = "Paladin",
                StartSkill = "Fixture Might",
                WalkVelocity = "6",
                RunVelocity = "9",
                stamina = "89",
                vit = "25",
                RunDrain = "20",
                StaminaPerLevel = "4",
                StaminaPerVitality = "4",
            },
            {
                class = "Amazon",
                StartSkill = "Fixture Idle",
                WalkVelocity = "6",
                RunVelocity = "9",
                stamina = "84",
                vit = "20",
                RunDrain = "20",
                StaminaPerLevel = "4",
                StaminaPerVitality = "4",
            },
        },
        ["data/global/excel/skills.txt"] = {
            {
                Id = "98",
                skill = "Fixture Might",
                skilldesc = "might",
                charclass = "pal",
                srvstfunc = "",
                srvdofunc = "65",
                aura = "1",
                immediate = "1",
                leftskill = "",
                range = "none",
                InGame = "1",
                aurafilter = "73731",
                aurarangecalc = "ln12",
                aurastate = "might",
                auratargetstate = "might",
                aurastat1 = "damagepercent",
                aurastatcalc1 = "ln34",
                mana = "0",
                lvlmana = "0",
                Param1 = "16",
                Param2 = "2",
                Param3 = "40",
                Param4 = "10",
                perdelay = "50",
            },
            {
                Id = "997",
                skill = "Fixture Idle",
                skilldesc = "idle",
                general = "1",
                leftskill = "1",
                passive = "0",
            },
            {
                Id = "104",
                skill = "Fixture Defiance",
                skilldesc = "defiance",
                charclass = "pal",
                srvstfunc = "",
                srvdofunc = "65",
                aura = "1",
                immediate = "1",
                leftskill = "",
                range = "none",
                InGame = "1",
                aurafilter = "73731",
                aurarangecalc = "ln12",
                aurastate = "defiance",
                auratargetstate = "defiance",
                aurastat1 = "skill_armor_percent",
                aurastatcalc1 = "ln34",
                mana = "0",
                lvlmana = "0",
                Param1 = "16",
                Param2 = "2",
                Param3 = "70",
                Param4 = "10",
                perdelay = "50",
            },
        },
        ["data/global/excel/skilldesc.txt"] = {
            { skilldesc = "might", ListRow = "1", IconCel = "4" },
            { skilldesc = "idle", ListRow = "2", IconCel = "0" },
            { skilldesc = "defiance", ListRow = "3", IconCel = "16" },
        },
        ["data/global/excel/states.txt"] = {
            { state = "might", id = "33", aura = "1", stat = "damagepercent" },
            { state = "defiance", id = "37", aura = "1", stat = "skill_armor_percent" },
        },
    },
    cases = {
        test.case("right_selected_aura_tracks_party_range_without_casting", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = entry("alice", "Paladin", 10),
            }),
            test.submit_system({
                tick = 1,
                sequence = 2,
                kind = "system.player.enter",
                payload = entry("bob", "Amazon", 14),
            }),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local effects = aura_effects()
                test.assert(#effects == 1, [=[#effects == 1]=])
                local effect = ecs.get(effects[1], "d2legacy.skill.aura_effect")
                local source = ecs.get(effects[1], "d2legacy.stat.source")
                test.assert(
                    effect:get("source_id") == "aura:alice:98"
                        and effect:get("skill_level") == 1
                        and effect:get("state_id") == "might"
                        and source:get("value") == 40
                        and source:get("operation") == "percent",
                    [=[solo aura owns one self relationship and source]=]
                )
                local alice
                for _, player in ipairs(ecs.query({ all = { "d2legacy.player.identity" } })) do
                    if ecs.get(player, "d2legacy.player.identity"):get("player") == "alice" then
                        alice = player
                    end
                end
                local emitter = ecs.get(alice, "d2legacy.skill.aura_emitter")
                test.assert(emitter:get("radius") == 16 and emitter:get("value") == 40)
                local assignment = ecs.get(alice, "d2legacy.player.skill_assignment")
                local original_left = assignment:get("left")
                local ok, err = pcall(require("d2legacy.commands.cast").apply_assignment, {
                    player = "alice",
                    payload = { left = 98 },
                })
                test.assert(not ok and tostring(err):find("skill is not allowed on left", 1, true))
                test.expect(assignment:get("left")):equals(original_left)
            end),
            test.submit({
                tick = 3,
                sequence = 1,
                player = "alice",
                kind = "party.invite",
                payload = { target = "bob" },
            }),
            test.step(1),
            test.submit({
                tick = 4,
                sequence = 1,
                player = "bob",
                kind = "party.accept",
                payload = { inviter = "alice" },
            }),
            test.step(2),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local effects = aura_effects()
                test.assert(#effects == 2, [=[#effects == 2]=])
                for _, entity in ipairs(effects) do
                    test.expect(ecs.get(entity, "d2legacy.stat.source"):get("value")):equals(40)
                end
                local applied = 0
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.state.event" } })) do
                    if ecs.get(entity, "d2legacy.state.event"):get("kind") == "state_applied" then
                        applied = applied + 1
                    end
                end
                test.assert(applied == 2, [=[applied == 2]=])
            end),
            test.submit({
                tick = 6,
                sequence = 1,
                player = "alice",
                kind = "player.use_skill",
                payload = { side = "right" },
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice
                for _, player in ipairs(ecs.query({ all = { "d2legacy.player.identity" } })) do
                    if ecs.get(player, "d2legacy.player.identity"):get("player") == "alice" then
                        alice = player
                    end
                end
                test.assert(ecs.get(alice, "d2legacy.skill.cast") == nil, [=[a selected aura does not cast]=])
                test.expect(ecs.get(alice, "d2legacy.player.vitals"):get("mana")):equals(20)
            end),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                for _, player in ipairs(ecs.query({ all = { "d2legacy.player.identity" } })) do
                    if ecs.get(player, "d2legacy.player.identity"):get("player") == "bob" then
                        ecs.get(player, "d2legacy.world.position"):set("x", 40)
                    end
                end
            end),
            test.step(2),
            test.run(function()
                test.assert(#aura_effects() == 1, [=[#aura_effects() == 1]=])
            end),
            test.submit({
                tick = 9,
                sequence = 2,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 997 },
            }),
            test.step(2),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.assert(#aura_effects() == 0, [=[#aura_effects() == 0]=])
                test.assert(
                    #ecs.query({ all = { "d2legacy.skill.aura_emitter" } }) == 0,
                    [=[#ecs.query({ all = { "d2legacy.skill.aura_emitter" } }) == 0]=]
                )
            end),
        }),
        test.case("overlapping_same_state_auras_choose_one_deterministic_strongest_source", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = entry("alice", "Paladin", 10),
            }),
            test.submit_system({
                tick = 1,
                sequence = 2,
                kind = "system.player.enter",
                payload = entry("carol", "Paladin", 12),
            }),
            test.step(2),
            test.submit({
                tick = 3,
                sequence = 1,
                player = "alice",
                kind = "party.invite",
                payload = { target = "carol" },
            }),
            test.step(1),
            test.submit({
                tick = 4,
                sequence = 1,
                player = "carol",
                kind = "party.accept",
                payload = { inviter = "alice" },
            }),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local effects = aura_effects()
                test.assert(#effects == 2, [=[same-state auras produce one relationship per target]=])
                for _, entity in ipairs(effects) do
                    local source = ecs.get(entity, "d2legacy.stat.source")
                    test.expect(source:get("source_id")):equals("aura:alice:98")
                    test.expect(source:get("value")):equals(40)
                end
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.player.learned_skill" } })) do
                    local learned = ecs.get(entity, "d2legacy.player.learned_skill")
                    local owner = learned:get("owner")
                    local identity = ecs.get(owner, "d2legacy.player.identity")
                    if identity:get("player") == "carol" and learned:get("skill_id") == 98 then
                        learned:set("level", 2)
                    end
                end
            end),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local effects = aura_effects()
                test.assert(#effects == 2, [=[stronger same-state aura still produces one relationship per target]=])
                for _, entity in ipairs(effects) do
                    local effect = ecs.get(entity, "d2legacy.skill.aura_effect")
                    local source = ecs.get(entity, "d2legacy.stat.source")
                    test.expect(effect:get("source_id")):equals("aura:carol:98")
                    test.expect(effect:get("skill_level")):equals(2)
                    test.expect(source:get("value")):equals(50)
                end
            end),
        }),
        test.case("different_selected_aura_states_stack_on_each_party_member", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = entry("alice", "Paladin", 10),
            }),
            test.submit_system({
                tick = 1,
                sequence = 2,
                kind = "system.player.enter",
                payload = entry("carol", "Paladin", 12),
            }),
            test.step(2),
            test.submit({
                tick = 3,
                sequence = 1,
                player = "alice",
                kind = "party.invite",
                payload = { target = "carol" },
            }),
            test.step(1),
            test.submit({
                tick = 4,
                sequence = 1,
                player = "carol",
                kind = "party.accept",
                payload = { inviter = "alice" },
            }),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                for _, player in ipairs(ecs.query({ all = { "d2legacy.player.identity" } })) do
                    if ecs.get(player, "d2legacy.player.identity"):get("player") == "carol" then
                        ecs.create({
                            ["d2legacy.player.learned_skill"] = {
                                owner = player,
                                skill_id = 104,
                                level = 1,
                                list_row = 3,
                                left_allowed = false,
                                right_allowed = true,
                            },
                        })
                    end
                end
            end),
            test.submit({
                tick = 6,
                sequence = 2,
                player = "carol",
                kind = "player.assign_skills",
                payload = { right = 104 },
            }),
            test.step(3),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local effects = aura_effects()
                test.assert(#effects == 4, [=[two distinct aura states produce two relationships per target]=])
                local counts = { might = 0, defiance = 0 }
                for _, entity in ipairs(effects) do
                    local effect = ecs.get(entity, "d2legacy.skill.aura_effect")
                    local source = ecs.get(entity, "d2legacy.stat.source")
                    local state = effect:get("state_id")
                    counts[state] = counts[state] + 1
                    if state == "might" then
                        test.assert(
                            source:get("stat") == "damagepercent" and source:get("value") == 40,
                            [=[Might retains its damage source]=]
                        )
                    else
                        test.assert(
                            state == "defiance"
                                and source:get("stat") == "defense"
                                and source:get("operation") == "percent"
                                and source:get("value") == 70,
                            [=[Defiance owns a generic defense percentage source]=]
                        )
                    end
                end
                test.assert(counts.might == 2 and counts.defiance == 2, [=[both states affect both party members]=])
                for _, player in ipairs(ecs.query({ all = { "d2legacy.player.combat_stats" } })) do
                    test.expect(ecs.get(player, "d2legacy.player.combat_stats"):get("defense")):equals(178)
                end
            end),
        }),
    },
})
