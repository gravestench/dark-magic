local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local function armor(id, name, event, event_function, other_a, other_b, values)
    return {
        Id = tostring(id),
        skill = name,
        skilldesc = string.lower(string.gsub(name, " ", "")),
        charclass = "sor",
        leftskill = "1",
        general = "0",
        passive = "0",
        srvstfunc = "",
        srvdofunc = "18",
        srvmissilea = values.missile or "",
        cltoverlaya = values.reaction_overlay or "",
        aurastate = values.state,
        auralencalc = "ln34+(skill('" .. other_a .. "'.blvl)+skill('" .. other_b .. "'.blvl))*par7",
        aurastat1 = "skill_armor_percent",
        aurastatcalc1 = "ln12",
        auraevent1 = event,
        auraeventfunc1 = event_function,
        calc1 = values.calc1 or "",
        mana = tostring(values.mana),
        manashift = "8",
        Param1 = tostring(values.defense),
        Param2 = tostring(values.defense_per_level),
        Param3 = tostring(values.duration),
        Param4 = tostring(values.duration_per_level),
        Param5 = tostring(values.reaction_duration or 0),
        Param6 = tostring(values.reaction_duration_per_level or 0),
        Param7 = "250",
        Param8 = tostring(values.synergy_percent),
        EType = values.element or "",
        HitShift = "7",
        EMin = tostring(values.minimum or 0),
        EMax = tostring(values.maximum or 0),
        EMinLev1 = tostring(values.minimum_gains and values.minimum_gains[1] or 0),
        EMinLev2 = tostring(values.minimum_gains and values.minimum_gains[2] or 0),
        EMinLev3 = tostring(values.minimum_gains and values.minimum_gains[3] or 0),
        EMinLev4 = tostring(values.minimum_gains and values.minimum_gains[4] or 0),
        EMinLev5 = tostring(values.minimum_gains and values.minimum_gains[5] or 0),
        EMaxLev1 = tostring(values.maximum_gains and values.maximum_gains[1] or 0),
        EMaxLev2 = tostring(values.maximum_gains and values.maximum_gains[2] or 0),
        EMaxLev3 = tostring(values.maximum_gains and values.maximum_gains[3] or 0),
        EMaxLev4 = tostring(values.maximum_gains and values.maximum_gains[4] or 0),
        EMaxLev5 = tostring(values.maximum_gains and values.maximum_gains[5] or 0),
        EDmgSymPerCalc = values.damage_synergy or "",
        ELen = tostring(values.cold_length or 0),
        ELevLen1 = "0",
        ELevLen2 = "25",
        ELevLen3 = tostring(values.cold_length_tier_three or 25),
        ELevLen4 = "0",
        ELevLen5 = "0",
    }
end

local skills = {
    armor(40, "Frozen Armor", "damagedinmelee", "2", "Shiver Armor", "Chilling Armor", {
        state = "frozenarmor",
        mana = 7,
        defense = 30,
        defense_per_level = 5,
        duration = 3000,
        duration_per_level = 300,
        reaction_duration = 30,
        reaction_duration_per_level = 3,
        synergy_percent = 5,
        calc1 = "ln56*(100+((skill('Shiver Armor'.blvl)+skill('Chilling Armor'.blvl))*par8))/100",
        reaction_overlay = "frozenarmor_hit",
    }),
    armor(50, "Shiver Armor", "attackedinmelee", "3", "Frozen Armor", "Chilling Armor", {
        state = "shiverarmor",
        mana = 11,
        defense = 45,
        defense_per_level = 6,
        duration = 3000,
        duration_per_level = 300,
        synergy_percent = 9,
        element = "cold",
        minimum = 12,
        maximum = 16,
        minimum_gains = { 4, 6, 8, 10, 12 },
        maximum_gains = { 5, 7, 9, 11, 13 },
        damage_synergy = "(skill('Frozen Armor'.blvl)+skill('Chilling Armor'.blvl))*par8",
        cold_length = 100,
        cold_length_tier_three = 50,
        reaction_overlay = "shiverarmor_hit",
    }),
    armor(60, "Chilling Armor", "hitbymissile", "1", "Frozen Armor", "Shiver Armor", {
        state = "chillingarmor",
        missile = "chillingarmorbolt",
        mana = 17,
        defense = 45,
        defense_per_level = 5,
        duration = 3600,
        duration_per_level = 150,
        synergy_percent = 7,
        element = "cold",
        minimum = 8,
        maximum = 12,
        minimum_gains = { 2, 4, 6, 8, 10 },
        maximum_gains = { 3, 5, 7, 9, 11 },
        damage_synergy = "(skill('Frozen Armor'.blvl)+skill('Shiver Armor'.blvl))*par8",
        cold_length = 100,
        reaction_overlay = "chillingarmor_hit",
    }),
}

local function unit(id, x, cold_resist)
    local ecs = require("engine.ecs/v1")
    return ecs.create({
        ["d2legacy.monster.stats"] = { health = 10000, max_health = 10000 },
        ["d2legacy.combat.defense"] = {
            base_cold_resist = cold_resist or 0,
            cold_resist = cold_resist or 0,
            max_cold_resist = 75,
        },
        ["d2legacy.world.position"] = { x = x, y = 12 },
        ["d2legacy.world.location"] = { act = 1, level_id = 1 },
        ["d2legacy.world.collider"] = { radius = 0.5 },
        ["d2legacy.world.selectable"] = { id = id, kind = "hostile", label = id, radius = 0.5 },
    })
end

local function cast(skill_id, tick, sequence)
    sequence = sequence or 1
    return {
        fixtures.command(
            "player.assign_skills",
            { left = skill_id },
            { tick = tick, sequence = sequence, player = "alice" }
        ),
        test.step(1),
        fixtures.command(
            "player.use_skill",
            { side = "left" },
            { tick = tick + 1, sequence = sequence + 1, player = "alice" }
        ),
        test.step(3),
    }
end

return test.suite({
    profile = "authority",
    tier = "fast",
    initial_data = {
        ["d2legacy.development_skills"] = {
            enabled = true,
            replace = true,
            skill_ids = { 40, 50, 60 },
            skill_level = 3,
            left = 40,
        },
    },
    records = {
        ["data/global/excel/charstats.txt"] = {
            {
                class = "Amazon",
                WalkVelocity = "6",
                RunVelocity = "9",
                stamina = "84",
                vit = "20",
                RunDrain = "20",
                StaminaPerLevel = "4",
                StaminaPerVitality = "4",
            },
        },
        ["data/global/excel/skills.txt"] = skills,
        ["data/global/excel/skilldesc.txt"] = {
            { skilldesc = "frozenarmor", ListRow = "0", IconCel = "0" },
            { skilldesc = "shiverarmor", ListRow = "1", IconCel = "1" },
            { skilldesc = "chillingarmor", ListRow = "2", IconCel = "2" },
        },
        ["data/global/excel/states.txt"] = {
            { state = "freeze", id = "1", group = "0" },
            { state = "cold", id = "2", group = "0" },
            { state = "frozenarmor", id = "10", group = "1" },
            { state = "shiverarmor", id = "88", group = "1" },
            { state = "chillingarmor", id = "20", group = "1" },
        },
        ["data/global/excel/Missiles.txt"] = {
            {
                Missile = "chillingarmorbolt",
                Skill = "Chilling Armor",
                pSrvDoFunc = "1",
                CollideType = "3",
                CollideKill = "1",
                Vel = "18",
                Range = "25",
                Size = "1",
                CelFile = "IceBolt",
                NumDirections = "16",
                AnimSpeed = "16",
                LoopAnim = "1",
                Trans = "1",
                TravelSound = "sorceress_icebolt_1",
                HitSound = "impact_cold_1",
            },
        },
    },
    cases = {
        test.case("recasting_refreshes_and_another_armor_replaces_the_exclusive_state", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ mana = 100, max_mana = 100 }),
            }),
            test.step(1),
            test.submit(cast(40, 2)[1]),
            cast(40, 2)[2],
            test.submit(cast(40, 2)[3]),
            cast(40, 2)[4],
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local states = ecs.query({ all = { "d2legacy.state.instance" } })
                test.assert(#states == 1, [=[the first cast creates one armor state]=])
                test.expect(ecs.get(states[1], "d2legacy.state.instance"):get("state_id")):equals("frozenarmor")
            end),
            test.submit(cast(40, 6, 3)[1]),
            cast(40, 6, 3)[2],
            test.submit(cast(40, 6, 3)[3]),
            cast(40, 6, 3)[4],
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local states = ecs.query({ all = { "d2legacy.state.instance" } })
                test.assert(#states == 1, [=[recasting refreshes rather than stacking]=])
                test.expect(ecs.get(states[1], "d2legacy.state.instance"):get("applied_tick")):equals(4)
                test.expect(ecs.get(states[1], "d2legacy.state.instance"):get("expires_tick")):equals(5108)
            end),
            test.submit(cast(50, 10, 5)[1]),
            cast(50, 10, 5)[2],
            test.submit(cast(50, 10, 5)[3]),
            cast(50, 10, 5)[4],
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local states = ecs.query({ all = { "d2legacy.state.instance" } })
                test.assert(#states == 1, [=[the shared States group retains only the selected armor]=])
                local state = ecs.get(states[1], "d2legacy.state.instance")
                test.expect(state:get("state_id")):equals("shiverarmor")
                test.expect(state:get("expires_tick") - state:get("applied_tick")):equals(5100)
                local sources = ecs.query({ all = { "d2legacy.stat.source" } })
                test.assert(#sources == 1, [=[replacement keeps one defense modifier source]=])
                test.expect(ecs.get(sources[1], "d2legacy.stat.source"):get("value")):equals(57)
                local replaced = false
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.state.event" } })) do
                    local event = ecs.get(entity, "d2legacy.state.event")
                    replaced = replaced or event:get("reason") == "exclusive_group_replaced"
                end
                test.assert(replaced, [=[replacement is an explicit deterministic state event]=])
            end),
        }),
        test.case("frozen_armor_requires_damage_and_converts_player_freeze_to_chill", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ mana = 100, max_mana = 100 }),
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.player.learned_skill" } })) do
                    local learned = ecs.get(entity, "d2legacy.player.learned_skill")
                    if learned:get("skill_id") == 50 or learned:get("skill_id") == 60 then
                        ecs.get(entity, "d2legacy.player.skill_hard_level"):set("level", 1)
                    end
                end
            end),
            test.submit(cast(40, 2)[1]),
            cast(40, 2)[2],
            test.submit(cast(40, 2)[3]),
            cast(40, 2)[4],
            test.run(function()
                local ecs = require("engine.ecs/v1")
                unit("monster:frozen", 13, 0)
                ecs.create({
                    ["d2legacy.combat.melee_event"] = {
                        kind = "hit_resolved",
                        tick = 6,
                        attacker_id = "monster:frozen",
                        target_id = "player:alice",
                        hit = false,
                        outcome = "miss",
                    },
                })
            end),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local states = ecs.query({ all = { "d2legacy.state.instance" } })
                test.assert(#states == 1, [=[a miss does not freeze]=])
                local armor_state = ecs.get(states[1], "d2legacy.state.instance")
                test.expect(armor_state:get("expires_tick") - armor_state:get("applied_tick")):equals(4100)
                ecs.create({
                    ["d2legacy.combat.melee_event"] = {
                        kind = "hit_resolved",
                        tick = 8,
                        attacker_id = "monster:frozen",
                        target_id = "player:alice",
                        hit = true,
                        damage_raw = 256,
                        outcome = "hit",
                    },
                })
            end),
            test.step(2),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local found
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.state.instance" } })) do
                    local state = ecs.get(entity, "d2legacy.state.instance")
                    if state:get("state_id") == "freeze" then
                        found = state
                    end
                end
                test.assert(found ~= nil and found:get("reaction") == "", [=[the successful hit freezes the monster]=])
                test.expect(found:get("expires_tick") - found:get("applied_tick")):equals(39)
                local reaction_cues = ecs.query({ all = { "d2legacy.presentation.effect_cue" } })
                test.assert(#reaction_cues == 1, [=[one successful melee reaction emits one presentation cue]=])
                test.expect(ecs.get(reaction_cues[1], "d2legacy.presentation.effect_cue"):get("overlay_id"))
                    :equals("frozenarmor_hit")
                ecs.create({
                    ["d2legacy.player.identity"] = {
                        character_id = "bob",
                        player = "bob",
                        name = "Bob",
                        class = "Amazon",
                    },
                    ["d2legacy.player.vitals"] = { health = 50, max_health = 50 },
                    ["d2legacy.player.movement_stats"] = {},
                    ["d2legacy.combat.defense"] = { cold_resist = 0, max_cold_resist = 75 },
                    ["d2legacy.world.position"] = { x = 14, y = 12 },
                    ["d2legacy.world.location"] = { act = 1, level_id = 1 },
                    ["d2legacy.world.collider"] = { radius = 1 },
                    ["d2legacy.world.selectable"] = {
                        id = "player:bob",
                        kind = "player",
                        label = "Bob",
                        radius = 0.75,
                    },
                })
                ecs.create({
                    ["d2legacy.combat.melee_event"] = {
                        kind = "hit_resolved",
                        tick = 10,
                        attacker_id = "player:bob",
                        target_id = "player:alice",
                        hit = true,
                        damage_raw = 256,
                        outcome = "hit",
                    },
                })
            end),
            test.step(2),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local bob
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.player.identity" } })) do
                    if ecs.get(entity, "d2legacy.player.identity"):get("player") == "bob" then
                        bob = entity
                    end
                end
                local chilled
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.state.instance" } })) do
                    local state = ecs.get(entity, "d2legacy.state.instance")
                    if state:get("target"):id() == bob:id() and state:get("state_id") == "cold" then
                        chilled = state
                    end
                end
                test.assert(chilled ~= nil and not chilled:get("action_disabled"), [=[PvP converts freeze to chill]=])
                test.expect(chilled:get("expires_tick") - chilled:get("applied_tick")):equals(39)
                local boss = unit("monster:boss", 15, 0)
                ecs.set(boss, "d2legacy.monster.freeze_immune", {})
                ecs.create({
                    ["d2legacy.combat.melee_event"] = {
                        kind = "hit_resolved",
                        tick = 12,
                        attacker_id = "monster:boss",
                        target_id = "player:alice",
                        hit = true,
                        damage_raw = 256,
                        outcome = "hit",
                    },
                })
            end),
            test.step(2),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local boss
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.world.selectable" } })) do
                    if ecs.get(entity, "d2legacy.world.selectable"):get("id") == "monster:boss" then
                        boss = entity
                    end
                end
                local chilled
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.state.instance" } })) do
                    local state = ecs.get(entity, "d2legacy.state.instance")
                    if state:get("target"):id() == boss:id() and state:get("state_id") == "cold" then
                        chilled = state
                    end
                    test.assert(
                        state:get("target"):id() ~= boss:id() or state:get("state_id") ~= "freeze",
                        [=[a freeze-immune monster never receives the hard-freeze state]=]
                    )
                end
                test.assert(
                    chilled ~= nil and not chilled:get("action_disabled"),
                    [=[boss-class monsters are chilled]=]
                )
                test.expect(chilled:get("expires_tick") - chilled:get("applied_tick")):equals(39)
            end),
        }),
        test.case("shiver_armor_reacts_to_misses_and_respects_cold_immunity", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ mana = 100, max_mana = 100 }),
            }),
            test.step(1),
            test.submit(cast(50, 2)[1]),
            cast(50, 2)[2],
            test.submit(cast(50, 2)[3]),
            cast(50, 2)[4],
            test.run(function()
                local ecs = require("engine.ecs/v1")
                unit("monster:normal", 13, 0)
                unit("monster:immune", 15, 100)
                for index, id in ipairs({ "monster:normal", "monster:immune" }) do
                    ecs.create({
                        ["d2legacy.combat.melee_event"] = {
                            kind = "hit_resolved",
                            tick = 6,
                            attacker_id = id,
                            target_id = "player:alice",
                            hit = false,
                            outcome = "miss",
                            hand = tostring(index),
                        },
                    })
                end
            end),
            test.step(2),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local by_id = {}
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.world.selectable" } })) do
                    by_id[ecs.get(entity, "d2legacy.world.selectable"):get("id")] = entity
                end
                test.assert(
                    ecs.get(by_id["monster:normal"], "d2legacy.monster.stats"):get("health") < 10000,
                    [=[a melee miss still receives Shiver Armor damage]=]
                )
                test.expect(ecs.get(by_id["monster:immune"], "d2legacy.monster.stats"):get("health")):equals(10000)
                local cold_targets = {}
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.state.instance" } })) do
                    local state = ecs.get(entity, "d2legacy.state.instance")
                    if state:get("state_id") == "cold" then
                        cold_targets[state:get("target"):id()] = state
                    end
                end
                test.assert(cold_targets[by_id["monster:normal"]:id()] ~= nil, [=[non-immune attacker is chilled]=])
                test.assert(cold_targets[by_id["monster:immune"]:id()] == nil, [=[cold immunity suppresses chill]=])
                test.expect(
                    cold_targets[by_id["monster:normal"]:id()]:get("expires_tick")
                        - cold_targets[by_id["monster:normal"]:id()]:get("applied_tick")
                ):equals(100)
            end),
        }),
        test.case("chilling_armor_returns_an_ordinary_record_authored_projectile", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ mana = 100, max_mana = 100 }),
            }),
            test.step(1),
            test.submit(cast(60, 2)[1]),
            cast(60, 2)[2],
            test.submit(cast(60, 2)[3]),
            cast(60, 2)[4],
            test.run(function()
                local ecs = require("engine.ecs/v1")
                unit("monster:ranged", 15, 0)
                ecs.create({
                    ["d2legacy.combat.event"] = {
                        kind = "damage_applied",
                        tick = 6,
                        attacker_id = "monster:ranged",
                        target_id = "player:alice",
                        source_kind = "missile",
                        damage_channel = "physical",
                        damage_raw = 256,
                    },
                })
            end),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local projectiles = ecs.query({ all = { "d2legacy.missile.projectile" } })
                test.assert(#projectiles == 1, [=[one ranged hit creates one return projectile]=])
                local projectile = ecs.get(projectiles[1], "d2legacy.missile.projectile")
                test.expect(projectile:get("missile_id")):equals("chillingarmorbolt")
                test.expect(projectile:get("directions")):equals(16)
                test.expect(projectile:get("transparency_mode")):equals(1)
                local cues = ecs.query({ all = { "d2legacy.presentation.effect_cue" } })
                test.assert(#cues == 2, [=[the ranged reaction emits its overlay and missile travel sound]=])
                local saw_overlay = false
                local saw_travel_sound = false
                for _, entity in ipairs(cues) do
                    local cue = ecs.get(entity, "d2legacy.presentation.effect_cue")
                    saw_overlay = saw_overlay or cue:get("overlay_id") == "chillingarmor_hit"
                    saw_travel_sound = saw_travel_sound or cue:get("sound") == "sorceress_icebolt_1"
                end
                test.assert(saw_overlay, [=[the ranged reaction exposes its authored overlay]=])
                test.assert(saw_travel_sound, [=[the return missile exposes its authored travel sound]=])
            end),
            test.step(8),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local ranged
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.world.selectable" } })) do
                    if ecs.get(entity, "d2legacy.world.selectable"):get("id") == "monster:ranged" then
                        ranged = entity
                    end
                end
                test.assert(
                    ecs.get(ranged, "d2legacy.monster.stats"):get("health") < 10000,
                    [=[the shared projectile contact path applies Chilling Armor damage]=]
                )
                local chilled = false
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.state.instance" } })) do
                    local state = ecs.get(entity, "d2legacy.state.instance")
                    chilled = chilled or state:get("state_id") == "cold" and state:get("target"):id() == ranged:id()
                end
                test.assert(chilled, [=[the return projectile applies its authored cold length]=])
                local hit_sound = false
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.presentation.effect_cue" } })) do
                    hit_sound = hit_sound
                        or ecs.get(entity, "d2legacy.presentation.effect_cue"):get("sound") == "impact_cold_1"
                end
                test.assert(hit_sound, [=[projectile contact emits its authored hit sound cue]=])
            end),
        }),
    },
})
