local test = require("d2legacy.tests/v1")

local function skill(id, name, missile, channel)
    return {
        Id = tostring(id),
        skill = name,
        srvmissile = missile,
        EType = channel,
        interrupt = "1",
        srvstfunc = "",
        srvdofunc = "",
        mana = "5",
        manashift = "7",
        EMin = "3",
        EMax = "6",
        HitShift = "8",
    }
end

local function missile(name, skill_name, velocity, knockback)
    return {
        Missile = name,
        Skill = skill_name,
        pSrvDoFunc = "1",
        CollideType = "3",
        CollideKill = "1",
        Vel = tostring(velocity),
        Range = "40",
        Size = "2",
        KnockBack = knockback and tostring(knockback) or "",
        CelFile = name,
    }
end

local fire_bolt = skill(36, "Fire Bolt", "firebolt", "fire")
fire_bolt.EDmgSymPerCalc = "(skill('Fire Ball'.blvl)+skill('Meteor'.blvl))*par8"
fire_bolt.Param8 = "16"
local fire_ball = skill(47, "Fire Ball", "fireball", "fire")
fire_ball.EDmgSymPerCalc = "(skill('Fire Bolt'.blvl)+skill('Meteor'.blvl))*par8"
fire_ball.Param8 = "14"
local ice_blast = skill(45, "Ice Blast", "iceblast", "cold")
ice_blast.EDmgSymPerCalc = "(skill('Ice Bolt'.blvl)+skill('Blizzard'.blvl)+skill('Frozen Orb'.blvl))*par8"
ice_blast.Param8 = "8"
ice_blast.ELen = "75"
ice_blast.ELevLen1 = "5"
ice_blast.ELevLen2 = "5"
ice_blast.ELevLen3 = "5"
ice_blast.ELenSymPerCalc = "(skill('Glacial Spike'.blvl))*par7"
ice_blast.Param7 = "10"
local glacial_spike = skill(55, "Glacial Spike", "glacialspike", "cold")
glacial_spike.EMin = "32"
glacial_spike.EMax = "48"
glacial_spike.HitShift = "7"
glacial_spike.EDmgSymPerCalc = "(skill('Ice Bolt'.blvl)+skill('Ice Blast'.blvl)+skill('Frozen Orb'.blvl))*par8"
glacial_spike.Param8 = "5"
glacial_spike.aurarangecalc = "ln12"
glacial_spike.Param1 = "4"
glacial_spike.Param2 = "0"
glacial_spike.Param3 = "50"
glacial_spike.Param4 = "3"
glacial_spike.Param7 = "3"
glacial_spike.auralencalc = "ln34 * (100 + skill('Blizzard'.blvl) * par7) / 100"
local fixture_burst = skill(901, "Fixture Burst", "fixtureburst", "magic")

local function impact_missile(name, skill_name, explosion, radius)
    local result = missile(name, skill_name, 20)
    result.pSrvHitFunc = "1"
    result.sHitPar1 = tostring(radius)
    result.ExplosionMissile = explosion
    return result
end

local function explosion(name, lifetime)
    return {
        Missile = name,
        Explosion = "1",
        Range = tostring(lifetime),
        CelFile = name,
        NumDirections = "1",
        AnimSpeed = "16",
        LoopAnim = "0",
    }
end

return test.suite({
    name = "straight missile skill definitions",
    profile = "module",
    tier = "fast",
    records = {
        ["data/global/excel/skills.txt"] = {
            fire_bolt,
            fire_ball,
            ice_blast,
            glacial_spike,
            { Id = "39", skill = "Ice Bolt" },
            { Id = "56", skill = "Meteor" },
            { Id = "59", skill = "Blizzard" },
            { Id = "64", skill = "Frozen Orb" },
            -- Synthetic authored data proves family reuse; it is not shipped
            -- gameplay coverage or a claim about another Diablo II skill.
            skill(900, "Fixture Bolt", "fixturebolt", "magic"),
            fixture_burst,
        },
        ["data/global/excel/Missiles.txt"] = {
            missile("firebolt", "Fire Bolt", 20),
            missile("fixturebolt", "Fixture Bolt", 25, 75),
            impact_missile("fireball", "Fire Ball", "fireexplosion", 4),
            impact_missile("fixtureburst", "Fixture Burst", "fixtureexplosion", 6),
            (function()
                local result = missile("iceblast", "Ice Blast", 12)
                result.pSrvDmgFunc = "4"
                result.ExplosionMissile = "freezingarrowexp1"
                return result
            end)(),
            (function()
                local result = missile("glacialspike", "Glacial Spike", 16)
                result.pSrvHitFunc = "13"
                result.EType = "frze"
                result.HitFlags = "2"
                result.CltHitSubMissile1 = "freezingarrowexp1"
                return result
            end)(),
            explosion("fireexplosion", 16),
            explosion("fixtureexplosion", 12),
            explosion("freezingarrowexp1", 16),
        },
    },
    cases = {
        test.case("decodes_multiple_configurations_without_skill_branches", function(t)
            t:run(function()
                local definitions = require("d2legacy.data.missile_skills").load({ 36, 900 })
                test.expect(definitions[36].behavior):equals("missile.straight")
                test.expect(definitions[36].missile_id):equals("firebolt")
                test.expect(definitions[36].damage_channel):equals("fire")
                test.expect(definitions[36].knockback_value):equals(0)
                test.expect(definitions[36].damage_synergy_skill_ids[1]):equals(47)
                test.expect(definitions[36].damage_synergy_skill_ids[2]):equals(56)
                test.expect(definitions[36].damage_synergy_percent_per_level):equals(16)
                test.expect(definitions[900].behavior):equals("missile.straight")
                test.expect(definitions[900].missile_id):equals("fixturebolt")
                test.expect(definitions[900].damage_channel):equals("magic")
                test.expect(definitions[900].knockback_value):equals(75)
                test.expect(definitions[900].speed_per_tick):equals(1)
            end)
        end),
        test.case("decodes_multiple_area_impact_configurations_without_skill_branches", function(t)
            t:run(function()
                local definitions =
                    require("d2legacy.data.missile_skills").load({ 47, 901 }, "missile.straight-impact-area")
                test.expect(definitions[47].trajectory):equals("straight")
                test.expect(definitions[47].impact_radius):equals(4)
                test.expect(definitions[47].impact_missile_id):equals("fireexplosion")
                test.expect(definitions[47].impact_lifetime_ticks):equals(16)
                test.expect(definitions[47].damage_synergy_skill_ids[1]):equals(36)
                test.expect(definitions[47].damage_synergy_skill_ids[2]):equals(56)
                test.expect(definitions[47].damage_synergy_percent_per_level):equals(14)
                test.expect(definitions[901].impact_radius):equals(6)
                test.expect(definitions[901].impact_missile_id):equals("fixtureexplosion")
                test.expect(definitions[901].impact_lifetime_ticks):equals(12)
            end)
        end),
        test.case("decodes_record_driven_freeze_state_without_a_skill_branch", function(t)
            t:run(function()
                local definition = require("d2legacy.data.missile_skills").load({ 45 }, "missile.straight-freeze")[45]
                test.expect(definition.trajectory):equals("straight")
                test.expect(definition.on_hit_state_id):equals("freeze")
                test.expect(definition.on_hit_state_duration_policy):equals("monster_cold")
                test.expect(definition.effect_duration_base):equals(75)
                test.expect(definition.effect_duration_per_level):equals(5)
                test.expect(definition.effect_duration_synergy_skill_ids[1]):equals(55)
                test.expect(definition.effect_duration_synergy_percent_per_level):equals(10)
                test.expect(definition.damage_synergy_skill_ids[1]):equals(39)
                test.expect(definition.damage_synergy_skill_ids[2]):equals(59)
                test.expect(definition.damage_synergy_skill_ids[3]):equals(64)
                test.expect(definition.damage_synergy_percent_per_level):equals(8)
                test.expect(definition.impact_missile_id):equals("freezingarrowexp1")
            end)
        end),
        test.case("composes_area_impact_and_freeze_without_a_skill_branch", function(t)
            t:run(function()
                local definition =
                    require("d2legacy.data.missile_skills").load({ 55 }, "missile.straight-impact-area-freeze")[55]
                test.expect(definition.trajectory):equals("straight")
                test.expect(definition.impact_radius):equals(4)
                test.expect(definition.impact_radius_per_level):equals(0)
                test.expect(definition.on_hit_state_id):equals("freeze")
                test.expect(definition.effect_duration_base):equals(50)
                test.expect(definition.effect_duration_per_level):equals(3)
                test.expect(definition.effect_duration_synergy_skill_ids[1]):equals(59)
                test.expect(definition.effect_duration_synergy_percent_per_level):equals(3)
                test.expect(definition.damage_synergy_skill_ids[1]):equals(39)
                test.expect(definition.damage_synergy_skill_ids[2]):equals(45)
                test.expect(definition.damage_synergy_skill_ids[3]):equals(64)
                test.expect(definition.damage_synergy_percent_per_level):equals(5)
                test.expect(definition.impact_missile_id):equals("freezingarrowexp1")
            end)
        end),
    },
})
