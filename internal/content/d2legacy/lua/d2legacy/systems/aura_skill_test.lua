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

local function pulse_effects(owner)
    local ecs = require("engine.ecs/v1")
    local result = {}
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.skill.aura_pulse_effect" } })) do
        local effect = ecs.get(entity, "d2legacy.skill.aura_pulse_effect")
        if not owner or effect:get("emitter"):id() == owner:id() then
            result[#result + 1] = effect
        end
    end
    table.sort(result, function(left, right)
        return left:get("order") < right:get("order")
    end)
    return result
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

local function timed_state(target, state_id)
    local ecs = require("engine.ecs/v1")
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.state.instance" } })) do
        local instance = ecs.get(entity, "d2legacy.state.instance")
        if instance:get("target"):id() == target:id() and instance:get("state_id") == state_id then
            return instance
        end
    end
    return nil
end

local function learned_skill(owner, skill_id)
    local ecs = require("engine.ecs/v1")
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.player.learned_skill" } })) do
        local learned = ecs.get(entity, "d2legacy.player.learned_skill")
        if learned:get("owner"):id() == owner:id() and learned:get("skill_id") == skill_id then
            return entity
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

local function corpse(id, x, level_id, selectable, inactive)
    local ecs = require("engine.ecs/v1")
    local components = {
        ["d2legacy.monster.identity"] = {
            spawn_id = id,
            definition_id = "corpse-fixture",
            base_id = "corpse-fixture",
            graphics_id = "corpse-fixture",
            seed = id,
            treasure_class = "",
        },
        ["d2legacy.monster.death"] = {
            tick = 1,
            active = false,
            corpse_usable = true,
        },
        ["d2legacy.world.position"] = { x = x, y = 12 },
        ["d2legacy.world.location"] = { act = 1, level_id = level_id or 2 },
    }
    if selectable ~= false then
        components["d2legacy.monster.corpse_selectable"] = {}
    end
    if inactive then
        components["d2legacy.world.inactive"] = {}
    end
    return ecs.create(components)
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
                Id = "99",
                skill = "Prayer",
                skilldesc = "prayer",
                charclass = "pal",
                srvstfunc = "",
                srvdofunc = "65",
                aura = "1",
                immediate = "",
                leftskill = "",
                range = "none",
                InGame = "1",
                InTown = "1",
                aurafilter = "73731",
                aurarangecalc = "ln12",
                aurastate = "prayer",
                auratargetstate = "prayer",
                aurastat1 = "hitpoints",
                aurastatcalc1 = "edns",
                mana = "16",
                lvlmana = "3",
                minmana = "1",
                manashift = "4",
                Param1 = "16",
                Param2 = "2",
                perdelay = "50",
                HitShift = "8",
                EMin = "2",
                EMinLev1 = "1",
                EMinLev2 = "1",
                EMinLev3 = "2",
                EMinLev4 = "2",
                EMinLev5 = "3",
            },
            {
                Id = "109",
                skill = "Cleansing",
                skilldesc = "cleansing",
                charclass = "pal",
                srvstfunc = "",
                srvdofunc = "65",
                aura = "1",
                immediate = "",
                leftskill = "",
                range = "none",
                InGame = "1",
                InTown = "1",
                aurafilter = "73731",
                aurarangecalc = "ln12",
                aurastate = "cleansing",
                auratargetstate = "cleansing",
                aurastat1 = "item_poisonlengthresist",
                aurastatcalc1 = "100-dm34",
                aurastat2 = "hitpoints",
                aurastatcalc2 = "skill('Prayer'.edns)",
                mana = "0",
                lvlmana = "0",
                minmana = "0",
                manashift = "8",
                Param1 = "16",
                Param2 = "2",
                Param3 = "30",
                Param4 = "90",
                perdelay = "50",
                HitShift = "8",
            },
            {
                Id = "120",
                skill = "Meditation",
                skilldesc = "meditation",
                charclass = "pal",
                srvstfunc = "",
                srvdofunc = "65",
                aura = "1",
                immediate = "",
                leftskill = "",
                range = "none",
                InGame = "1",
                InTown = "1",
                aurafilter = "73729",
                aurarangecalc = "ln12",
                aurastate = "meditation",
                auratargetstate = "meditation",
                aurastat1 = "manarecoverybonus",
                aurastatcalc1 = "ln34",
                aurastat2 = "hitpoints",
                aurastatcalc2 = "skill('Prayer'.edns)",
                mana = "0",
                lvlmana = "0",
                minmana = "0",
                manashift = "8",
                Param1 = "16",
                Param2 = "2",
                Param3 = "300",
                Param4 = "25",
                perdelay = "50",
                HitShift = "8",
            },
            {
                Id = "124",
                skill = "Fixture Redemption",
                skilldesc = "redemption",
                charclass = "pal",
                srvstfunc = "",
                srvdofunc = "82",
                aura = "1",
                immediate = "",
                leftskill = "",
                range = "none",
                InGame = "1",
                InTown = "1",
                aurafilter = "4354",
                aurarangecalc = "ln12",
                aurastate = "redemption",
                auratargetstate = "",
                calc1 = "dm34",
                calc2 = "ln56",
                calc3 = "ln56",
                mana = "0",
                lvlmana = "0",
                minmana = "0",
                manashift = "8",
                Param1 = "16",
                Param2 = "0",
                Param3 = "10",
                Param4 = "100",
                Param5 = "25",
                Param6 = "5",
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
            { skilldesc = "prayer", ListRow = "3", IconCel = "6" },
            { skilldesc = "cleansing", ListRow = "3", IconCel = "38" },
            { skilldesc = "meditation", ListRow = "3", IconCel = "48" },
            { skilldesc = "redemption", ListRow = "3", IconCel = "54" },
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
            { state = "prayer", id = "34", aura = "1", stat = "" },
            { state = "cleansing", id = "45", aura = "1", stat = "" },
            { state = "meditation", id = "48", aura = "1", stat = "" },
            { state = "redemption", id = "50", aura = "1", stat = "" },
            { state = "poison", id = "2" },
            { state = "amplifydamage", id = "9", curse = "1", curable = "1" },
            { state = "battlecry", id = "89", curse = "1", curable = "" },
            { state = "shrine_armor", id = "128", curse = "1", curable = "" },
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
        test.case("cleansing_composes_duration_scaling_with_the_owners_learned_prayer_effect", {
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
                payload = entry("bob", "Amazon", 12),
            }),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice = player_named("alice")
                local bob = player_named("bob")
                for _, skill in ipairs({ { id = 99, level = 3 }, { id = 109, level = 3 } }) do
                    ecs.create({
                        ["d2legacy.player.learned_skill"] = {
                            owner = alice,
                            skill_id = skill.id,
                            level = skill.level,
                            list_row = 3,
                            left_allowed = false,
                            right_allowed = true,
                        },
                    })
                end
                local alice_vitals = ecs.get(alice, "d2legacy.player.vitals")
                alice_vitals:set("max_health", 50)
                alice_vitals:set("health", 45)
                alice_vitals:set("mana_raw", 0)
                alice_vitals:set("mana", 0)
                local bob_vitals = ecs.get(bob, "d2legacy.player.vitals")
                bob_vitals:set("max_health", 50)
                bob_vitals:set("health", 40)
                local function add_state(target, state_id, expires)
                    ecs.create({
                        ["d2legacy.state.instance"] = {
                            target = target,
                            state_id = state_id,
                            source_id = "fixture:" .. target:id() .. ":" .. state_id,
                            applied_tick = 2,
                            expires_tick = expires,
                            policy = "refresh_same_source",
                        },
                    })
                end
                for _, target in ipairs({ alice, bob }) do
                    add_state(target, "poison", 251)
                    add_state(target, "amplifydamage", 151)
                    add_state(target, "battlecry", 151)
                    add_state(target, "shrine_armor", 151)
                end
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
            test.submit({
                tick = 4,
                sequence = 2,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 109 },
            }),
            test.step(3),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice = player_named("alice")
                local pulse = ecs.get(alice, "d2legacy.skill.aura_pulse")
                test.assert(pulse ~= nil, "Cleansing did not compose its pulse schedule")
                test.expect(pulse:get("mana_cost_raw")):equals(0)
                test.expect(pulse:get("next_tick")):equals(51)
                local effects = pulse_effects(alice)
                test.expect(#effects):equals(2)
                test.expect(effects[1]:get("kind")):equals("scale_remaining_timed_state")
                test.expect(effects[1]:get("value")):equals(49)
                test.expect(effects[2]:get("kind")):equals("heal_life")
                test.expect(effects[2]:get("value")):equals(4)
                test.expect(#aura_effects()):equals(2)
            end),
            test.step(45),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice = player_named("alice")
                local bob = player_named("bob")
                test.expect(ecs.get(alice, "d2legacy.player.vitals"):get("health")):equals(49)
                test.expect(ecs.get(bob, "d2legacy.player.vitals"):get("health")):equals(44)
                test.expect(timed_state(alice, "poison"):get("expires_tick")):equals(149)
                test.expect(timed_state(bob, "poison"):get("expires_tick")):equals(149)
                test.expect(timed_state(alice, "amplifydamage"):get("expires_tick")):equals(100)
                test.expect(timed_state(bob, "amplifydamage"):get("expires_tick")):equals(100)
                test.expect(timed_state(alice, "battlecry"):get("expires_tick")):equals(151)
                test.expect(timed_state(bob, "battlecry"):get("expires_tick")):equals(151)
                test.expect(timed_state(alice, "shrine_armor"):get("expires_tick")):equals(100)
                test.expect(timed_state(bob, "shrine_armor"):get("expires_tick")):equals(100)
                ecs.get(alice, "d2legacy.player.vitals"):set("health", 30)
                ecs.get(bob, "d2legacy.player.vitals"):set("health", 30)
                ecs.destroy(assert(learned_skill(alice, 99), "fixture Prayer skill is missing"))
            end),
            test.step(50),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice = player_named("alice")
                local bob = player_named("bob")
                test.expect(ecs.get(alice, "d2legacy.player.vitals"):get("health")):equals(30)
                test.expect(ecs.get(bob, "d2legacy.player.vitals"):get("health")):equals(30)
                test.expect(timed_state(alice, "poison"):get("expires_tick")):equals(124)
                test.expect(timed_state(bob, "poison"):get("expires_tick")):equals(124)
                test.assert(timed_state(alice, "amplifydamage") == nil)
                test.assert(timed_state(bob, "amplifydamage") == nil)
                test.expect(timed_state(alice, "battlecry"):get("expires_tick")):equals(151)
                test.assert(timed_state(alice, "shrine_armor") == nil)
                test.expect(pulse_effects(alice)[2]:get("value")):equals(0)
            end),
            test.submit({
                tick = 102,
                sequence = 1,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 98 },
            }),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice = player_named("alice")
                test.assert(ecs.get(alice, "d2legacy.skill.aura_pulse") == nil)
                test.expect(#pulse_effects(alice)):equals(0)
            end),
        }),
        test.case("meditation_composes_mana_recovery_with_the_owners_learned_prayer_effect", {
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
                payload = entry("bob", "Amazon", 12),
            }),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice = player_named("alice")
                local bob = player_named("bob")
                for _, skill in ipairs({ { id = 99, level = 3 }, { id = 120, level = 3 } }) do
                    ecs.create({
                        ["d2legacy.player.learned_skill"] = {
                            owner = alice,
                            skill_id = skill.id,
                            level = skill.level,
                            list_row = 3,
                            left_allowed = false,
                            right_allowed = true,
                        },
                    })
                end
                for _, target in ipairs({ alice, bob }) do
                    local vitals = ecs.get(target, "d2legacy.player.vitals")
                    vitals:set("max_health", 50)
                    vitals:set("health", 40)
                    vitals:set("max_mana_raw", 300 * 256)
                    vitals:set("max_mana", 300)
                    vitals:set("mana_raw", 0)
                    vitals:set("mana", 0)
                    ecs.get(target, "d2legacy.player.resource_stats"):set("mana_regen_frames", 120)
                end
                reflected_damage_attacker("monster:meditation-filter", 100 * 256)
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
            test.submit({
                tick = 4,
                sequence = 2,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 120 },
            }),
            test.step(3),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice = player_named("alice")
                local pulse = ecs.get(alice, "d2legacy.skill.aura_pulse")
                test.assert(pulse ~= nil, "Meditation did not compose its pulse schedule")
                test.expect(pulse:get("mana_cost_raw")):equals(0)
                test.expect(pulse:get("next_tick")):equals(51)
                local effects = pulse_effects(alice)
                test.expect(#effects):equals(1)
                test.expect(effects[1]:get("kind")):equals("heal_life")
                test.expect(effects[1]:get("value")):equals(4)
                test.expect(#aura_effects()):equals(2)
                for _, effect_entity in ipairs(aura_effects()) do
                    test.expect(aura_source(effect_entity, "manarecoverybonus"):get("value")):equals(350)
                    local target = ecs.get(effect_entity, "d2legacy.skill.aura_effect"):get("target")
                    test.assert(ecs.get(target, "d2legacy.player.identity") ~= nil, "filter 73729 admitted a monster")
                    test.expect(ecs.get(target, "d2legacy.player.resource_stats"):get("manarecoverybonus")):equals(350)
                    local vitals = ecs.get(target, "d2legacy.player.vitals")
                    vitals:set("mana_raw", 0)
                    vitals:set("mana", 0)
                end
                test.expect(#ecs.query({ all = { "d2legacy.resource.mana_regen_suppression" } })):equals(0)
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(ecs.get(player_named("alice"), "d2legacy.player.vitals"):get("mana_raw")):equals(112)
                test.expect(ecs.get(player_named("bob"), "d2legacy.player.vitals"):get("mana_raw")):equals(112)
            end),
            test.step(44),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice = player_named("alice")
                local bob = player_named("bob")
                test.expect(ecs.get(alice, "d2legacy.player.vitals"):get("health")):equals(44)
                test.expect(ecs.get(bob, "d2legacy.player.vitals"):get("health")):equals(44)
                ecs.destroy(assert(learned_skill(alice, 99), "fixture Prayer skill is missing"))
            end),
            test.step(50),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice = player_named("alice")
                local bob = player_named("bob")
                test.expect(ecs.get(alice, "d2legacy.player.vitals"):get("health")):equals(44)
                test.expect(ecs.get(bob, "d2legacy.player.vitals"):get("health")):equals(44)
                test.expect(pulse_effects(alice)[1]:get("value")):equals(0)
            end),
            test.submit({
                tick = 102,
                sequence = 1,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 98 },
            }),
            test.step(3),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice = player_named("alice")
                test.assert(ecs.get(alice, "d2legacy.skill.aura_pulse") == nil)
                test.expect(#pulse_effects(alice)):equals(0)
                test.expect(#ecs.query({ all = { "d2legacy.resource.mana_regen_suppression" } })):equals(0)
                test.expect(ecs.get(alice, "d2legacy.player.resource_stats"):get("manarecoverybonus")):equals(0)
            end),
        }),
        test.case("corpse_periodic_aura_uses_stable_eligibility_and_consumes_each_success_once", {
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
                ecs.get(alice, "d2legacy.world.location"):set("level_id", 2)
                local definition = require("d2legacy.authoritative").aura_skills[124]
                definition.pulse.chance_minimum = 0
                definition.pulse.chance_maximum = 0
                ecs.create({
                    ["d2legacy.player.learned_skill"] = {
                        owner = alice,
                        skill_id = 124,
                        level = 3,
                        list_row = 3,
                        left_allowed = false,
                        right_allowed = true,
                    },
                })
                local vitals = ecs.get(alice, "d2legacy.player.vitals")
                vitals:set("health", 10)
                vitals:set("max_health", 200)
                vitals:set("mana_raw", 0)
                vitals:set("mana", 0)
                vitals:set("max_mana_raw", 200 * 256)
                vitals:set("max_mana", 200)
                ecs.get(alice, "d2legacy.player.resource_stats"):set("mana_regen_frames", 0)
                corpse("eligible-b", 13, 2, true, false)
                corpse("eligible-a", 12, 2, true, false)
                corpse("missing-capability", 11, 2, false, false)
                corpse("out-of-range", 40, 2, true, false)
                corpse("other-level", 12, 3, true, false)
                corpse("inactive", 12, 2, true, true)
            end),
            test.submit({
                tick = 3,
                sequence = 1,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 124 },
            }),
            test.step(3),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice = player_named("alice")
                local pulse = ecs.get(alice, "d2legacy.skill.aura_pulse")
                test.expect(pulse:get("target_policy")):equals("eligible_corpses")
                test.expect(pulse:get("radius")):equals(16)
                test.expect(pulse:get("chance")):equals(0)
                test.expect(pulse:get("next_tick")):equals(51)
                test.expect(pulse:get("mana_cost_raw")):equals(0)
                local effects = pulse_effects(alice)
                test.expect(#effects):equals(3)
                test.expect(effects[1]:get("kind")):equals("restore_owner_life")
                test.expect(effects[1]:get("value")):equals(35 * 256)
                test.expect(effects[2]:get("kind")):equals("restore_owner_mana")
                test.expect(effects[2]:get("value")):equals(35 * 256)
                test.expect(effects[3]:get("kind")):equals("consume_corpse")
                test.expect(#aura_effects()):equals(1)
                local owner_state = ecs.get(aura_effects()[1], "d2legacy.skill.aura_effect")
                test.expect(owner_state:get("target"):id()):equals(alice:id())
                test.expect(owner_state:get("state_id")):equals("redemption")
            end),
            test.step(46),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.monster.death" } })) do
                    test.expect(ecs.get(entity, "d2legacy.monster.death"):get("corpse_usable")):equals(true)
                end
                test.expect(#ecs.query({ all = { "d2legacy.skill.aura_pulse_event" } })):equals(0)
                local definition = require("d2legacy.authoritative").aura_skills[124]
                definition.pulse.chance_minimum = 100
                definition.pulse.chance_maximum = 100
            end),
            test.step(50),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice = player_named("alice")
                local pulse = ecs.get(alice, "d2legacy.skill.aura_pulse")
                test.expect(pulse:get("chance")):equals(100)
                test.expect(pulse:get("next_tick")):equals(151)
                local vitals = ecs.get(alice, "d2legacy.player.vitals")
                test.expect(vitals:get("health")):equals(80)
                test.expect(vitals:get("mana_raw")):equals(70 * 256)
                local consumed = {}
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.monster.death" } })) do
                    local id = ecs.get(entity, "d2legacy.monster.identity"):get("spawn_id")
                    consumed[id] = not ecs.get(entity, "d2legacy.monster.death"):get("corpse_usable")
                end
                test.assert(consumed["eligible-a"] and consumed["eligible-b"], "eligible corpses were not consumed")
                test.assert(
                    not consumed["missing-capability"]
                        and not consumed["out-of-range"]
                        and not consumed["other-level"]
                        and not consumed["inactive"],
                    "corpse target filtering changed"
                )
                local events = ecs.query({ all = { "d2legacy.skill.aura_pulse_event" } })
                test.expect(#events):equals(2)
                local first = ecs.get(events[1], "d2legacy.skill.aura_pulse_event")
                local second = ecs.get(events[2], "d2legacy.skill.aura_pulse_event")
                local ordered = {
                    ecs.get(first:get("target"), "d2legacy.monster.identity"):get("spawn_id"),
                    ecs.get(second:get("target"), "d2legacy.monster.identity"):get("spawn_id"),
                }
                test.expect(table.concat(ordered, ",")):equals("eligible-a,eligible-b")
                test.expect(first:get("life_delta_raw")):equals(35 * 256)
                test.expect(first:get("mana_delta_raw")):equals(35 * 256)
                test.expect(second:get("life_delta_raw")):equals(35 * 256)
                test.expect(second:get("mana_delta_raw")):equals(35 * 256)
                vitals:set("health", vitals:get("max_health"))
                vitals:set("mana_raw", vitals:get("max_mana_raw"))
                vitals:set("mana", vitals:get("max_mana"))
            end),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(#ecs.query({ all = { "d2legacy.skill.aura_pulse_event" } })):equals(2)
                local definition = require("d2legacy.authoritative").aura_skills[124]
                definition.pulse.chance_minimum = 100
                definition.pulse.chance_maximum = 100
                local vitals = ecs.get(player_named("alice"), "d2legacy.player.vitals")
                vitals:set("health", vitals:get("max_health"))
                vitals:set("mana_raw", vitals:get("max_mana_raw"))
                vitals:set("mana", vitals:get("max_mana"))
                corpse("full-resource", 12, 2, true, false)
            end),
            test.step(50),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local full
                for _, entity in ipairs(ecs.query({ all = { "d2legacy.monster.identity" } })) do
                    if ecs.get(entity, "d2legacy.monster.identity"):get("spawn_id") == "full-resource" then
                        full = entity
                    end
                end
                test.assert(full ~= nil, "full-resource corpse fixture is missing")
                test.expect(ecs.get(full, "d2legacy.monster.death"):get("corpse_usable")):equals(false)
                local events = ecs.query({ all = { "d2legacy.skill.aura_pulse_event" } })
                test.expect(#events):equals(3)
                local final = ecs.get(events[3], "d2legacy.skill.aura_pulse_event")
                test.expect(final:get("life_delta_raw")):equals(0)
                test.expect(final:get("mana_delta_raw")):equals(0)
            end),
            test.step(1),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(#ecs.query({ all = { "d2legacy.skill.aura_pulse_event" } })):equals(3)
            end),
        }),
        test.case("corpse_periodic_aura_does_no_corpse_work_in_town_or_after_deselection", {
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
                ecs.get(alice, "d2legacy.world.location"):set("level_id", 1)
                local definition = require("d2legacy.authoritative").aura_skills[124]
                definition.pulse.chance_minimum = 100
                definition.pulse.chance_maximum = 100
                ecs.create({
                    ["d2legacy.player.learned_skill"] = {
                        owner = alice,
                        skill_id = 124,
                        level = 20,
                        list_row = 3,
                        left_allowed = false,
                        right_allowed = true,
                    },
                })
                corpse("town-corpse", 12, 1, true, false)
            end),
            test.submit({
                tick = 3,
                sequence = 1,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 124 },
            }),
            test.step(3),
            test.step(46),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local corpse_entity = ecs.query({ all = { "d2legacy.monster.death" } })[1]
                test.expect(ecs.get(corpse_entity, "d2legacy.monster.death"):get("corpse_usable")):equals(true)
            end),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice = player_named("alice")
                ecs.get(alice, "d2legacy.world.location"):set("level_id", 2)
                local corpse_entity = ecs.query({ all = { "d2legacy.monster.death" } })[1]
                ecs.get(corpse_entity, "d2legacy.world.location"):set("level_id", 2)
            end),
            test.submit({
                tick = 52,
                sequence = 2,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 98 },
            }),
            test.step(50),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local corpse_entity = ecs.query({ all = { "d2legacy.monster.death" } })[1]
                test.expect(ecs.get(corpse_entity, "d2legacy.monster.death"):get("corpse_usable")):equals(true)
                test.assert(ecs.get(player_named("alice"), "d2legacy.skill.aura_pulse") == nil)
            end),
        }),
        test.case("periodic_aura_heals_only_when_the_full_pulse_cost_is_funded_and_useful", {
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
                payload = entry("bob", "Amazon", 12),
            }),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice = player_named("alice")
                local bob = player_named("bob")
                ecs.create({
                    ["d2legacy.player.learned_skill"] = {
                        owner = alice,
                        skill_id = 99,
                        level = 3,
                        list_row = 3,
                        left_allowed = false,
                        right_allowed = true,
                    },
                })
                local alice_vitals = ecs.get(alice, "d2legacy.player.vitals")
                alice_vitals:set("max_health", 50)
                alice_vitals:set("health", 48)
                alice_vitals:set("mana_raw", 351)
                alice_vitals:set("mana", 1)
                local bob_vitals = ecs.get(bob, "d2legacy.player.vitals")
                bob_vitals:set("max_health", 50)
                bob_vitals:set("health", 45)
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
            test.submit({
                tick = 4,
                sequence = 2,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 99 },
            }),
            test.step(3),
            test.restore_checkpoint(),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice = player_named("alice")
                local pulse = ecs.get(alice, "d2legacy.skill.aura_pulse")
                test.assert(pulse ~= nil, "Prayer did not compose its periodic effect facts")
                local effects = pulse_effects(alice)
                test.expect(#effects):equals(1)
                test.expect(effects[1]:get("kind")):equals("heal_life")
                test.expect(effects[1]:get("value")):equals(4)
                test.expect(pulse:get("mana_cost_raw")):equals(352)
                test.expect(pulse:get("period_ticks")):equals(50)
                test.expect(pulse:get("next_tick")):equals(51)
                test.expect(#aura_effects()):equals(2)
                test.expect(#ecs.query({ all = { "d2legacy.stat.source" } })):equals(0)
            end),
            test.step(45),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice_vitals = ecs.get(player_named("alice"), "d2legacy.player.vitals")
                local bob_vitals = ecs.get(player_named("bob"), "d2legacy.player.vitals")
                test.expect(alice_vitals:get("health")):equals(48)
                test.expect(bob_vitals:get("health")):equals(45)
                test.expect(alice_vitals:get("mana_raw")):equals(351)
                test.expect(ecs.get(player_named("alice"), "d2legacy.skill.aura_pulse"):get("next_tick")):equals(101)
                test.expect(#ecs.query({ all = { "d2legacy.resource.mana_regen_suppression" } })):equals(0)
                alice_vitals:set("mana_raw", 352)
                alice_vitals:set("mana", 1)
            end),
            test.step(50),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice_vitals = ecs.get(player_named("alice"), "d2legacy.player.vitals")
                local bob_vitals = ecs.get(player_named("bob"), "d2legacy.player.vitals")
                test.expect(alice_vitals:get("health")):equals(50)
                test.expect(bob_vitals:get("health")):equals(49)
                test.expect(alice_vitals:get("mana_raw")):equals(0)
                test.expect(alice_vitals:get("mana")):equals(0)
                local suppressions = ecs.query({ all = { "d2legacy.resource.mana_regen_suppression" } })
                test.expect(#suppressions):equals(1)
                local suppression = ecs.get(suppressions[1], "d2legacy.resource.mana_regen_suppression")
                test.expect(suppression:get("target"):id()):equals(player_named("alice"):id())
                test.expect(suppression:get("source_id")):equals("aura:alice:99")
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice_vitals = ecs.get(player_named("alice"), "d2legacy.player.vitals")
                local bob_vitals = ecs.get(player_named("bob"), "d2legacy.player.vitals")
                test.expect(alice_vitals:get("health")):equals(50)
                test.expect(bob_vitals:get("health")):equals(49)
                alice_vitals:set("mana_raw", 352)
                alice_vitals:set("mana", 1)
                bob_vitals:set("health", 50)
            end),
            test.step(49),
            test.restore_checkpoint(),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice = player_named("alice")
                test.expect(ecs.get(alice, "d2legacy.player.vitals"):get("mana_raw")):equals(352)
                test.expect(ecs.get(alice, "d2legacy.skill.aura_pulse"):get("next_tick")):equals(201)
                test.expect(#ecs.query({ all = { "d2legacy.resource.mana_regen_suppression" } })):equals(0)
            end),
            test.submit({
                tick = 153,
                sequence = 1,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 98 },
            }),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice = player_named("alice")
                test.assert(ecs.get(alice, "d2legacy.skill.aura_pulse") == nil)
                test.expect(#pulse_effects(alice)):equals(0)
                test.expect(#ecs.query({ all = { "d2legacy.resource.mana_regen_suppression" } })):equals(0)
            end),
        }),
        test.case("deselecting_a_paid_periodic_aura_removes_its_mana_regen_suppression", {
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
                ecs.create({
                    ["d2legacy.player.learned_skill"] = {
                        owner = alice,
                        skill_id = 99,
                        level = 1,
                        list_row = 3,
                        left_allowed = false,
                        right_allowed = true,
                    },
                })
                local vitals = ecs.get(alice, "d2legacy.player.vitals")
                vitals:set("max_health", 50)
                vitals:set("health", 40)
                vitals:set("mana_raw", 1000)
                vitals:set("mana", 3)
            end),
            test.submit({
                tick = 3,
                sequence = 1,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 99 },
            }),
            test.step(49),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(ecs.get(player_named("alice"), "d2legacy.player.vitals"):get("health")):equals(42)
                test.expect(#ecs.query({ all = { "d2legacy.resource.mana_regen_suppression" } })):equals(1)
            end),
            test.submit({
                tick = 52,
                sequence = 2,
                player = "alice",
                kind = "player.assign_skills",
                payload = { right = 98 },
            }),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local alice = player_named("alice")
                test.assert(ecs.get(alice, "d2legacy.skill.aura_pulse") == nil)
                test.expect(#ecs.query({ all = { "d2legacy.resource.mana_regen_suppression" } })):equals(0)
            end),
        }),
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
