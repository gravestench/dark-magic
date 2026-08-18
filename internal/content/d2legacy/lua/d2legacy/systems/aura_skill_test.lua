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
        base_attack_rating = 100,
        vitality = class == "Paladin" and 25 or 20,
    })
end

local function aura_effects()
    local ecs = require("engine.ecs/v1")
    return ecs.query({ all = { "d2legacy.skill.aura_effect" } })
end

local function aura_source(effect_entity, stat)
    local ecs = require("engine.ecs/v1")
    local effect = ecs.get(effect_entity, "d2legacy.skill.aura_effect")
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.stat.source" } })) do
        local source = ecs.get(entity, "d2legacy.stat.source")
        if
            source:get("target"):id() == effect:get("target"):id()
            and source:get("owner_source_id") == effect:get("source_id")
            and (not stat or source:get("stat") == stat)
        then
            return source
        end
    end
    return nil
end

local function player_named(name)
    local ecs = require("engine.ecs/v1")
    for _, player in ipairs(ecs.query({ all = { "d2legacy.player.identity" } })) do
        if ecs.get(player, "d2legacy.player.identity"):get("player") == name then
            return player
        end
    end
    return nil
end

local function reflected_damage_attacker(id, health_raw)
    local ecs = require("engine.ecs/v1")
    id = id or "monster:reflect-attacker"
    health_raw = health_raw or 100 * 256
    return ecs.create({
        ["d2legacy.monster.identity"] = {
            spawn_id = id,
            definition_id = "reflection-test",
            base_id = "",
            graphics_id = "",
            seed = "1",
            treasure_class = "",
        },
        ["d2legacy.monster.stats"] = {
            level = 1,
            health = health_raw,
            max_health = health_raw,
            defense = 0,
            attack_rating = 100,
        },
        ["d2legacy.combat.defense"] = {
            base_physical_resist = 50,
            base_fire_resist = 0,
            base_cold_resist = 0,
            base_lightning_resist = 0,
            physical_resist = 50,
            fire_resist = 0,
            cold_resist = 0,
            lightning_resist = 0,
            max_fire_resist = 75,
            max_cold_resist = 75,
            max_lightning_resist = 75,
            physical_reduction_raw = 0,
        },
        ["d2legacy.world.selectable"] = {
            id = id,
            kind = "hostile",
            label = "Reflection attacker",
            owner = "",
            radius = 0.5,
            priority = 1,
        },
    })
end

local function emit_melee_damage(attacker_id, rolled_physical_raw)
    local ecs = require("engine.ecs/v1")
    local damage = require("d2legacy.policy.damage")
    local bundle = require("d2legacy.policy.damage_bundle")
    local defender = player_named("alice")
    test.assert(defender ~= nil, "fixture Paladin is missing")
    local resolved = damage.resolve(defender, bundle.single("physical", rolled_physical_raw), ecs)
    return ecs.create({
        ["d2legacy.combat.melee_event"] = {
            kind = "hit_resolved",
            tick = 7,
            attacker_id = attacker_id,
            target_id = "player:alice",
            hit = true,
            damage_raw = resolved.damage_raw,
            remaining_health_raw = resolved.remaining_health_raw,
            hand = "rarm",
            outcome = "hit",
        },
        ["d2legacy.combat.attack_result"] = {
            tick = 7,
            attacker_id = attacker_id,
            target_id = "player:alice",
            source_kind = "melee",
            outcome = "hit",
        },
        ["d2legacy.combat.event"] = {
            kind = resolved.lethal and "unit_died" or "damage_applied",
            tick = 7,
            attacker_id = attacker_id,
            target_id = "player:alice",
            source_kind = "melee",
            damage_channel = resolved.channel,
            rolled_damage_raw = resolved.rolled_damage_raw,
            damage_raw = resolved.damage_raw,
            remaining_health_raw = resolved.remaining_health_raw,
        },
        ["d2legacy.combat.damage_bundle"] = bundle.stage_component(resolved.rolled, resolved.mitigated),
    })
end

local function emit_non_reflecting_damage_boundaries()
    local ecs = require("engine.ecs/v1")
    local bundle = require("d2legacy.policy.damage_bundle")
    ecs.create({
        ["d2legacy.combat.melee_event"] = {
            kind = "hit_resolved",
            attacker_id = "monster:reflect-attacker",
            target_id = "player:alice",
            hit = false,
            outcome = "miss",
        },
    })
    local fire = bundle.normalize({ fire = 5 * 256 })
    ecs.create({
        ["d2legacy.combat.melee_event"] = {
            kind = "hit_resolved",
            attacker_id = "monster:reflect-attacker",
            target_id = "player:alice",
            hit = true,
            damage_raw = 5 * 256,
            outcome = "hit",
        },
        ["d2legacy.combat.event"] = {
            kind = "damage_applied",
            attacker_id = "monster:reflect-attacker",
            target_id = "player:alice",
            source_kind = "melee",
            damage_channel = "fire",
            rolled_damage_raw = 5 * 256,
            damage_raw = 5 * 256,
        },
        ["d2legacy.combat.damage_bundle"] = bundle.stage_component(fire, fire),
    })
    ecs.create({
        ["d2legacy.combat.event"] = {
            kind = "damage_applied",
            attacker_id = "monster:reflect-attacker",
            target_id = "player:alice",
            source_kind = "missile",
            damage_channel = "physical",
            rolled_damage_raw = 5 * 256,
            damage_raw = 5 * 256,
        },
    })
end

return test.suite({
    profile = "authority",
    tier = "fast",
    records = {
        ["data/global/excel/ItemStatCost.txt"] = {
            { Stat = "skill_staminapercent", op = "1", ["op stat1"] = "maxstamina" },
        },
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
                Id = "103",
                skill = "Fixture Thorns",
                skilldesc = "thorns",
                charclass = "pal",
                srvstfunc = "",
                srvdofunc = "65",
                aura = "1",
                immediate = "",
                leftskill = "",
                range = "none",
                InGame = "1",
                aurafilter = "73731",
                aurarangecalc = "ln12",
                aurastate = "thorns",
                auratargetstate = "thorns",
                aurastat1 = "thorns_percent",
                aurastatcalc1 = "ln34",
                mana = "0",
                lvlmana = "0",
                Param1 = "16",
                Param2 = "2",
                Param3 = "250",
                Param4 = "40",
                perdelay = "50",
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
            {
                Id = "100",
                skill = "Fixture Resist Fire",
                skilldesc = "resist fire",
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
                aurastate = "resistfire",
                auratargetstate = "resistfire",
                aurastat1 = "fireresist",
                aurastatcalc1 = "dm34",
                aurastat2 = "maxfireresist",
                aurastatcalc2 = "skill('Fixture Resist Fire'.blvl)",
                passivestate = "passive_resistfire",
                passivestat1 = "maxfireresist",
                passivecalc1 = "skill('Fixture Resist Fire'.blvl)/2",
                mana = "0",
                lvlmana = "0",
                Param1 = "16",
                Param2 = "2",
                Param3 = "35",
                Param4 = "150",
                perdelay = "50",
            },
            {
                Id = "105",
                skill = "Fixture Resist Cold",
                skilldesc = "resist cold",
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
                aurastate = "resistcold",
                auratargetstate = "resistcold",
                aurastat1 = "coldresist",
                aurastatcalc1 = "dm34",
                aurastat2 = "maxcoldresist",
                aurastatcalc2 = "skill('Fixture Resist Cold'.blvl)",
                passivestate = "passive_resistcold",
                passivestat1 = "maxcoldresist",
                passivecalc1 = "skill('Fixture Resist Cold'.blvl)/2",
                mana = "0",
                lvlmana = "0",
                Param1 = "16",
                Param2 = "2",
                Param3 = "35",
                Param4 = "150",
                perdelay = "50",
            },
            {
                Id = "110",
                skill = "Fixture Resist Lightning",
                skilldesc = "resist lightning",
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
                aurastate = "resistlight",
                auratargetstate = "resistlight",
                aurastat1 = "lightresist",
                aurastatcalc1 = "dm34",
                aurastat2 = "maxlightresist",
                aurastatcalc2 = "skill('Fixture Resist Lightning'.blvl)",
                passivestate = "passive_resistltng",
                passivestat1 = "maxlightresist",
                passivecalc1 = "skill('Fixture Resist Lightning'.blvl)/2",
                mana = "0",
                lvlmana = "0",
                Param1 = "16",
                Param2 = "2",
                Param3 = "35",
                Param4 = "150",
                perdelay = "50",
            },
            {
                Id = "115",
                skill = "Fixture Vigor",
                skilldesc = "vigor",
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
                aurastate = "stamina",
                auratargetstate = "stamina",
                aurastat1 = "staminarecoverybonus",
                aurastatcalc1 = "ln34",
                aurastat2 = "skill_staminapercent",
                aurastatcalc2 = "ln34",
                aurastat3 = "velocitypercent",
                aurastatcalc3 = "dm56",
                mana = "0",
                lvlmana = "0",
                Param1 = "16",
                Param2 = "3",
                Param3 = "50",
                Param4 = "25",
                Param5 = "7",
                Param6 = "50",
                perdelay = "50",
            },
            {
                Id = "125",
                skill = "Fixture Salvation",
                skilldesc = "salvation",
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
                aurastate = "resistall",
                auratargetstate = "resistall",
                aurastat1 = "fireresist",
                aurastatcalc1 = "dm34",
                aurastat2 = "coldresist",
                aurastatcalc2 = "dm34",
                aurastat3 = "lightresist",
                aurastatcalc3 = "dm34",
                mana = "0",
                lvlmana = "0",
                Param1 = "16",
                Param2 = "2",
                Param3 = "50",
                Param4 = "120",
                perdelay = "50",
            },
            {
                Id = "108",
                skill = "Fixture Blessed Aim",
                skilldesc = "blessed aim",
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
                aurastate = "blessedaim",
                auratargetstate = "blessedaim",
                aurastat1 = "item_tohit_percent",
                aurastatcalc1 = "ln34",
                passivestate = "penetrate",
                passivestat1 = "item_tohit_percent",
                passivecalc1 = "skill('Fixture Blessed Aim'.blvl) * par8",
                mana = "0",
                lvlmana = "0",
                Param1 = "16",
                Param2 = "2",
                Param3 = "75",
                Param4 = "15",
                Param8 = "5",
                perdelay = "50",
            },
        },
        ["data/global/excel/skilldesc.txt"] = {
            { skilldesc = "might", ListRow = "1", IconCel = "4" },
            { skilldesc = "thorns", ListRow = "10", IconCel = "14" },
            { skilldesc = "idle", ListRow = "2", IconCel = "0" },
            { skilldesc = "defiance", ListRow = "3", IconCel = "16" },
            { skilldesc = "blessed aim", ListRow = "4", IconCel = "24" },
            { skilldesc = "resist fire", ListRow = "5", IconCel = "8" },
            { skilldesc = "resist cold", ListRow = "6", IconCel = "18" },
            { skilldesc = "resist lightning", ListRow = "7", IconCel = "28" },
            { skilldesc = "salvation", ListRow = "8", IconCel = "58" },
            { skilldesc = "vigor", ListRow = "9", IconCel = "26" },
        },
        ["data/global/excel/states.txt"] = {
            { state = "might", id = "33", aura = "1", stat = "damagepercent" },
            { state = "thorns", id = "36", aura = "1", stat = "" },
            { state = "defiance", id = "37", aura = "1", stat = "skill_armor_percent" },
            { state = "blessedaim", id = "40", aura = "1", stat = "item_tohit_percent" },
            { state = "penetrate", id = "67" },
            { state = "resistfire", id = "3", aura = "1", stat = "fireresist" },
            { state = "passive_resistfire", id = "181" },
            { state = "resistcold", id = "4", aura = "1", stat = "coldresist" },
            { state = "passive_resistcold", id = "182" },
            { state = "resistlight", id = "5", aura = "1", stat = "lightresist" },
            { state = "passive_resistltng", id = "183" },
            { state = "resistall", id = "8", aura = "1", stat = "lightresist" },
            { state = "stamina", id = "41", aura = "1", stat = "maxstamina" },
        },
    },
    cases = {
        test.case("thorns_reflects_committed_melee_physical_damage_through_attacker_mitigation", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = entry("alice", "Paladin", 10),
            }),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice = player_named("alice")
                test.assert(alice ~= nil, "fixture Paladin is missing")
                ecs.create({
                    ["d2legacy.player.learned_skill"] = {
                        owner = alice,
                        skill_id = 103,
                        level = 3,
                        list_row = 10,
                        left_allowed = false,
                        right_allowed = true,
                    },
                })
                ecs.get(alice, "d2legacy.combat.defense"):set("base_physical_resist", 50)
                ecs.get(alice, "d2legacy.combat.defense"):set("physical_resist", 50)
                reflected_damage_attacker()
            end),
            test.submit({
                tick = 4,
                sequence = 1,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 103 },
            }),
            test.step(3),
            test.restore_checkpoint(),
            test.run(function()
                local effects = aura_effects()
                test.expect(#effects):equals(1)
                local source = aura_source(effects[1], "thorns_percent")
                test.expect(source:get("operation")):equals("add")
                test.expect(source:get("value")):equals(330)
                emit_non_reflecting_damage_boundaries()
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local monster = ecs.query({ all = { "d2legacy.monster.stats" } })[1]
                test.expect(ecs.get(monster, "d2legacy.monster.stats"):get("health")):equals(100 * 256)
                test.expect(#ecs.query({ all = { "d2legacy.combat.reflection_observed" } })):equals(2)
                test.expect(#ecs.query({ all = { "d2legacy.combat.event", "d2legacy.combat.melee_event" } })):equals(1)
            end),
            test.run(function()
                emit_melee_damage("monster:reflect-attacker", 20 * 256)
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local monsters = ecs.query({ all = { "d2legacy.monster.stats" } })
                test.expect(#monsters):equals(1)
                -- The defender first mitigates 20 to 10. Level-three Thorns
                -- reflects 330% (33), then the attacker's 50% resistance
                -- reduces that second physical transaction to 16.5.
                test.expect(ecs.get(monsters[1], "d2legacy.monster.stats"):get("health")):equals(21376)
                local reflected
                local observed = 0
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.combat.event" } })) do
                    local event = ecs.get(entity, "d2legacy.combat.event")
                    if event:get("source_kind") == "melee_reflection" then
                        reflected = event
                    end
                    if ecs.get(entity, "d2legacy.combat.reflection_observed") then
                        observed = observed + 1
                    end
                end
                test.assert(reflected ~= nil, "successful physical melee damage did not produce reflection")
                test.expect(reflected:get("rolled_damage_raw")):equals(33 * 256)
                test.expect(reflected:get("damage_raw")):equals(4224)
                test.expect(observed):equals(2)
            end),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local monster = ecs.query({ all = { "d2legacy.monster.stats" } })[1]
                test.expect(ecs.get(monster, "d2legacy.monster.stats"):get("health")):equals(21376)
            end),
            test.run(function()
                reflected_damage_attacker("monster:lethal-reflect-attacker", 8 * 256)
                emit_melee_damage("monster:lethal-reflect-attacker", 10 * 256)
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local lethal
                for _, monster in ipairs(ecs.query({ all = { "d2legacy.monster.identity" } })) do
                    local identity = ecs.get(monster, "d2legacy.monster.identity")
                    if identity:get("spawn_id") == "monster:lethal-reflect-attacker" then
                        lethal = monster
                    end
                end
                test.assert(lethal ~= nil, "lethal reflection attacker is missing")
                test.expect(ecs.get(lethal, "d2legacy.monster.stats"):get("health")):equals(0)
                local death = ecs.get(lethal, "d2legacy.monster.death")
                test.assert(death ~= nil, "lethal reflected damage bypassed the shared monster death consumer")
                test.expect(death:get("killer_id")):equals("player:alice")
                test.expect(death:get("credited_id")):equals("player:alice")
            end),
            test.submit({
                tick = 11,
                sequence = 1,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 98 },
            }),
            test.step(3),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(#ecs.query({ all = { "d2legacy.stat.source" } })):equals(1)
                emit_melee_damage("monster:reflect-attacker", 2 * 256)
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local monster = ecs.query({ all = { "d2legacy.monster.stats" } })[1]
                test.expect(ecs.get(monster, "d2legacy.monster.stats"):get("health")):equals(21376)
            end),
        }),
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
                local source = aura_source(effects[1])
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
                    test.expect(aura_source(entity):get("value")):equals(40)
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
                    local source = aura_source(entity)
                    test.expect(source:get("owner_source_id")):equals("aura:alice:98")
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
                    local source = aura_source(entity)
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
                    local source = aura_source(entity)
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
        test.case("learned_hard_point_passive_yields_to_the_selected_active_aura", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = entry("alice", "Paladin", 10),
            }),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice
                for _, player in ipairs(ecs.query({ all = { "d2legacy.player.identity" } })) do
                    if ecs.get(player, "d2legacy.player.identity"):get("player") == "alice" then
                        alice = player
                    end
                end
                ecs.create({
                    ["d2legacy.player.learned_skill"] = {
                        owner = alice,
                        skill_id = 108,
                        level = 2,
                        list_row = 4,
                        left_allowed = false,
                        right_allowed = true,
                    },
                })
            end),
            test.step(2),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local passive
                for _, entity in
                    ipairs(ecs.query({ all = { "d2legacy.player.learned_skill", "d2legacy.stat.source" } }))
                do
                    local learned = ecs.get(entity, "d2legacy.player.learned_skill")
                    if learned:get("skill_id") == 108 then
                        passive = ecs.get(entity, "d2legacy.stat.source")
                    end
                end
                test.assert(passive ~= nil, [=[Blessed Aim hard points own a composed passive source]=])
                test.expect(passive:get("stat")):equals("item_tohit_percent")
                test.expect(passive:get("value")):equals(10)
                for _, player in ipairs(ecs.query({ all = { "d2legacy.player.combat_stats" } })) do
                    test.expect(ecs.get(player, "d2legacy.player.combat_stats"):get("attack_rating")):equals(181)
                end
            end),
            test.submit({
                tick = 6,
                sequence = 1,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 108 },
            }),
            test.step(3),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local passive_count = 0
                for _, entity in
                    ipairs(ecs.query({ all = { "d2legacy.player.learned_skill", "d2legacy.stat.source" } }))
                do
                    if ecs.get(entity, "d2legacy.player.learned_skill"):get("skill_id") == 108 then
                        passive_count = passive_count + 1
                    end
                end
                test.expect(passive_count):equals(0)
                local effects = aura_effects()
                test.expect(#effects):equals(1)
                local source = aura_source(effects[1])
                test.expect(source:get("owner_source_id")):equals("aura:alice:108")
                test.expect(source:get("value")):equals(90)
                for _, player in ipairs(ecs.query({ all = { "d2legacy.player.combat_stats" } })) do
                    test.expect(ecs.get(player, "d2legacy.player.combat_stats"):get("attack_rating")):equals(313)
                end
            end),
            test.submit({
                tick = 8,
                sequence = 1,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 98 },
            }),
            test.step(3),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local passive_count = 0
                for _, entity in
                    ipairs(ecs.query({ all = { "d2legacy.player.learned_skill", "d2legacy.stat.source" } }))
                do
                    if ecs.get(entity, "d2legacy.player.learned_skill"):get("skill_id") == 108 then
                        passive_count = passive_count + 1
                        test.expect(ecs.get(entity, "d2legacy.stat.source"):get("value")):equals(10)
                    end
                end
                test.expect(passive_count):equals(1)
            end),
        }),
        test.case("multi_stat_resistance_aura_replaces_only_its_inactive_hard_point_passive", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = entry("alice", "Paladin", 10),
            }),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                for _, player in ipairs(ecs.query({ all = { "d2legacy.player.identity" } })) do
                    if ecs.get(player, "d2legacy.player.identity"):get("player") == "alice" then
                        ecs.create({
                            ["d2legacy.player.learned_skill"] = {
                                owner = player,
                                skill_id = 100,
                                level = 3,
                                list_row = 5,
                                left_allowed = false,
                                right_allowed = true,
                            },
                        })
                    end
                end
            end),
            test.step(2),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local passive
                for _, entity in
                    ipairs(ecs.query({ all = { "d2legacy.player.learned_skill", "d2legacy.stat.source" } }))
                do
                    if ecs.get(entity, "d2legacy.player.learned_skill"):get("skill_id") == 100 then
                        passive = ecs.get(entity, "d2legacy.stat.source")
                    end
                end
                test.expect(passive:get("stat")):equals("max_fire_resist")
                test.expect(passive:get("value")):equals(1)
                for _, player in ipairs(ecs.query({ all = { "d2legacy.combat.defense" } })) do
                    local defense = ecs.get(player, "d2legacy.combat.defense")
                    test.expect(defense:get("fire_resist")):equals(0)
                    test.expect(defense:get("max_fire_resist")):equals(76)
                end
            end),
            test.submit({
                tick = 6,
                sequence = 1,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 100 },
            }),
            test.step(3),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local effects = aura_effects()
                test.expect(#effects):equals(1)
                local fire = aura_source(effects[1], "fire_resist")
                local maximum = aura_source(effects[1], "max_fire_resist")
                test.expect(fire:get("value")):equals(76)
                test.expect(maximum:get("value")):equals(3)
                test.expect(#ecs.query({ all = { "d2legacy.stat.source" } })):equals(2)
                for _, player in ipairs(ecs.query({ all = { "d2legacy.combat.defense" } })) do
                    local defense = ecs.get(player, "d2legacy.combat.defense")
                    test.expect(defense:get("fire_resist")):equals(76)
                    test.expect(defense:get("max_fire_resist")):equals(78)
                end
            end),
            test.submit({
                tick = 8,
                sequence = 1,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 98 },
            }),
            test.step(3),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(#aura_effects()):equals(1)
                local aura_sources = 0
                local passive_sources = 0
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.stat.source" } })) do
                    local source = ecs.get(entity, "d2legacy.stat.source")
                    if source:get("owner_source_id") == "aura:alice:98" then
                        aura_sources = aura_sources + 1
                    elseif source:get("source_id"):find("skill%-passive:.*:100") then
                        passive_sources = passive_sources + 1
                        test.expect(source:get("value")):equals(1)
                    end
                end
                test.expect(aura_sources):equals(1)
                test.expect(passive_sources):equals(1)
            end),
        }),
        test.case("cold_resistance_reuses_multi_stat_aura_and_rational_passive", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = entry("alice", "Paladin", 10),
            }),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                for _, player in ipairs(ecs.query({ all = { "d2legacy.player.identity" } })) do
                    if ecs.get(player, "d2legacy.player.identity"):get("player") == "alice" then
                        ecs.create({
                            ["d2legacy.player.learned_skill"] = {
                                owner = player,
                                skill_id = 105,
                                level = 3,
                                list_row = 6,
                                left_allowed = false,
                                right_allowed = true,
                            },
                        })
                    end
                end
            end),
            test.step(2),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local passive
                for _, entity in
                    ipairs(ecs.query({ all = { "d2legacy.player.learned_skill", "d2legacy.stat.source" } }))
                do
                    if ecs.get(entity, "d2legacy.player.learned_skill"):get("skill_id") == 105 then
                        passive = ecs.get(entity, "d2legacy.stat.source")
                    end
                end
                test.expect(passive:get("stat")):equals("max_cold_resist")
                test.expect(passive:get("value")):equals(1)
                for _, player in ipairs(ecs.query({ all = { "d2legacy.combat.defense" } })) do
                    local defense = ecs.get(player, "d2legacy.combat.defense")
                    test.expect(defense:get("cold_resist")):equals(0)
                    test.expect(defense:get("max_cold_resist")):equals(76)
                end
            end),
            test.submit({
                tick = 6,
                sequence = 1,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 105 },
            }),
            test.step(3),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local effects = aura_effects()
                test.expect(#effects):equals(1)
                test.expect(aura_source(effects[1], "cold_resist"):get("value")):equals(76)
                test.expect(aura_source(effects[1], "max_cold_resist"):get("value")):equals(3)
                local passive_count = 0
                for _, entity in
                    ipairs(ecs.query({ all = { "d2legacy.player.learned_skill", "d2legacy.stat.source" } }))
                do
                    if ecs.get(entity, "d2legacy.player.learned_skill"):get("skill_id") == 105 then
                        passive_count = passive_count + 1
                    end
                end
                test.expect(passive_count):equals(0)
                local mitigation = require("d2legacy.policy.mitigation")
                for _, player in ipairs(ecs.query({ all = { "d2legacy.combat.defense" } })) do
                    local defense = ecs.get(player, "d2legacy.combat.defense")
                    test.expect(defense:get("cold_resist")):equals(76)
                    test.expect(defense:get("max_cold_resist")):equals(78)
                    test.expect(mitigation.apply(1000, "cold", defense)):equals(240)
                end
            end),
        }),
        test.case("lightning_resistance_reuses_multi_stat_aura_and_rational_passive", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = entry("alice", "Paladin", 10),
            }),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                for _, player in ipairs(ecs.query({ all = { "d2legacy.player.identity" } })) do
                    if ecs.get(player, "d2legacy.player.identity"):get("player") == "alice" then
                        ecs.create({
                            ["d2legacy.player.learned_skill"] = {
                                owner = player,
                                skill_id = 110,
                                level = 3,
                                list_row = 7,
                                left_allowed = false,
                                right_allowed = true,
                            },
                        })
                    end
                end
            end),
            test.step(2),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local passive
                for _, entity in
                    ipairs(ecs.query({ all = { "d2legacy.player.learned_skill", "d2legacy.stat.source" } }))
                do
                    if ecs.get(entity, "d2legacy.player.learned_skill"):get("skill_id") == 110 then
                        passive = ecs.get(entity, "d2legacy.stat.source")
                    end
                end
                test.expect(passive:get("stat")):equals("max_lightning_resist")
                test.expect(passive:get("value")):equals(1)
            end),
            test.submit({
                tick = 6,
                sequence = 1,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 110 },
            }),
            test.step(3),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local effects = aura_effects()
                test.expect(#effects):equals(1)
                test.expect(aura_source(effects[1], "lightning_resist"):get("value")):equals(76)
                test.expect(aura_source(effects[1], "max_lightning_resist"):get("value")):equals(3)
                local passive_count = 0
                for _, entity in
                    ipairs(ecs.query({ all = { "d2legacy.player.learned_skill", "d2legacy.stat.source" } }))
                do
                    if ecs.get(entity, "d2legacy.player.learned_skill"):get("skill_id") == 110 then
                        passive_count = passive_count + 1
                    end
                end
                test.expect(passive_count):equals(0)
                local mitigation = require("d2legacy.policy.mitigation")
                for _, player in ipairs(ecs.query({ all = { "d2legacy.combat.defense" } })) do
                    local defense = ecs.get(player, "d2legacy.combat.defense")
                    test.expect(defense:get("lightning_resist")):equals(76)
                    test.expect(defense:get("max_lightning_resist")):equals(78)
                    test.expect(mitigation.apply(1000, "lightning", defense)):equals(240)
                end
            end),
        }),
        test.case("salvation_owns_three_elemental_sources_without_a_passive", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = entry("alice", "Paladin", 10),
            }),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                for _, player in ipairs(ecs.query({ all = { "d2legacy.player.identity" } })) do
                    if ecs.get(player, "d2legacy.player.identity"):get("player") == "alice" then
                        ecs.create({
                            ["d2legacy.player.learned_skill"] = {
                                owner = player,
                                skill_id = 125,
                                level = 3,
                                list_row = 8,
                                left_allowed = false,
                                right_allowed = true,
                            },
                        })
                    end
                end
            end),
            test.submit({
                tick = 4,
                sequence = 1,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 125 },
            }),
            test.step(3),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local effects = aura_effects()
                test.expect(#effects):equals(1)
                test.expect(aura_source(effects[1], "fire_resist"):get("value")):equals(75)
                test.expect(aura_source(effects[1], "cold_resist"):get("value")):equals(75)
                test.expect(aura_source(effects[1], "lightning_resist"):get("value")):equals(75)
                test.expect(#ecs.query({ all = { "d2legacy.stat.source" } })):equals(3)
                local mitigation = require("d2legacy.policy.mitigation")
                for _, player in ipairs(ecs.query({ all = { "d2legacy.combat.defense" } })) do
                    local defense = ecs.get(player, "d2legacy.combat.defense")
                    test.expect(defense:get("fire_resist")):equals(75)
                    test.expect(defense:get("cold_resist")):equals(75)
                    test.expect(defense:get("lightning_resist")):equals(75)
                    test.expect(defense:get("max_fire_resist")):equals(75)
                    test.expect(defense:get("max_cold_resist")):equals(75)
                    test.expect(defense:get("max_lightning_resist")):equals(75)
                    test.expect(mitigation.apply(1000, "fire", defense)):equals(250)
                    test.expect(mitigation.apply(1000, "cold", defense)):equals(250)
                    test.expect(mitigation.apply(1000, "lightning", defense)):equals(250)
                end
            end),
            test.submit({
                tick = 7,
                sequence = 1,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 98 },
            }),
            test.step(3),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(#ecs.query({ all = { "d2legacy.stat.source" } })):equals(1)
                for _, player in ipairs(ecs.query({ all = { "d2legacy.combat.defense" } })) do
                    local defense = ecs.get(player, "d2legacy.combat.defense")
                    test.expect(defense:get("fire_resist")):equals(0)
                    test.expect(defense:get("cold_resist")):equals(0)
                    test.expect(defense:get("lightning_resist")):equals(0)
                end
            end),
        }),
        test.case("vigor_sources_drive_movement_recovery_and_maximum_stamina", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = entry("alice", "Paladin", 10),
            }),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                for _, player in ipairs(ecs.query({ all = { "d2legacy.player.identity" } })) do
                    if ecs.get(player, "d2legacy.player.identity"):get("player") == "alice" then
                        ecs.create({
                            ["d2legacy.player.learned_skill"] = {
                                owner = player,
                                skill_id = 115,
                                level = 3,
                                list_row = 9,
                                left_allowed = false,
                                right_allowed = true,
                            },
                        })
                    end
                end
            end),
            test.submit({
                tick = 4,
                sequence = 1,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 115 },
            }),
            test.step(3),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local effects = aura_effects()
                test.expect(#effects):equals(1)
                test.expect(aura_source(effects[1], "staminarecoverybonus"):get("value")):equals(100)
                test.expect(aura_source(effects[1], "skill_staminapercent"):get("value")):equals(100)
                test.expect(aura_source(effects[1], "velocitypercent"):get("value")):equals(22)
                local players = ecs.query({ all = { "d2legacy.player.movement_stats", "d2legacy.player.vitals" } })
                test.expect(#players):equals(1)
                local stats = ecs.get(players[1], "d2legacy.player.movement_stats")
                local vitals = ecs.get(players[1], "d2legacy.player.vitals")
                test.expect(stats:get("staminarecoverybonus")):equals(100)
                test.expect(stats:get("velocitypercent")):equals(22)
                test.expect(vitals:get("max_stamina_raw")):equals(178 * 256)
                vitals:set("stamina_raw", 89 * 256)
                vitals:set("stamina", 89)
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                for _, player in ipairs(ecs.query({ all = { "d2legacy.player.vitals" } })) do
                    test.expect(ecs.get(player, "d2legacy.player.vitals"):get("stamina_raw")):equals(89 * 256 + 356)
                end
            end),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player_motion = require("d2legacy.gameplay.player_motion")
                for _, player in ipairs(ecs.query({ all = { "d2legacy.player.identity" } })) do
                    player_motion.locomotion(player, { x = 1, y = 0, running = false })
                end
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local moving_players = ecs.query({ all = { "d2legacy.player.identity", "d2legacy.world.velocity" } })
                for _, player in ipairs(moving_players) do
                    local velocity = ecs.get(player, "d2legacy.world.velocity")
                    test.assert(
                        math.abs(velocity:get("x") - 7.32) < 0.000000001,
                        "Vigor walk velocity = " .. tostring(velocity:get("x"))
                    )
                end
            end),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                for _, player in ipairs(ecs.query({ all = { "d2legacy.player.skill_assignment" } })) do
                    ecs.get(player, "d2legacy.player.skill_assignment"):set("right", 98)
                end
            end),
            test.step(3),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                for _, player in ipairs(ecs.query({ all = { "d2legacy.player.movement_stats" } })) do
                    local stats = ecs.get(player, "d2legacy.player.movement_stats")
                    local vitals = ecs.get(player, "d2legacy.player.vitals")
                    test.expect(stats:get("staminarecoverybonus")):equals(0)
                    test.expect(stats:get("velocitypercent")):equals(0)
                    test.expect(vitals:get("max_stamina_raw")):equals(89 * 256)
                end
            end),
        }),
    },
})
