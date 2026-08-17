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

return test.suite({
    name = "straight missile skill definitions",
    profile = "module",
    tier = "fast",
    records = {
        ["data/global/excel/skills.txt"] = {
            fire_bolt,
            { Id = "47", skill = "Fire Ball" },
            { Id = "56", skill = "Meteor" },
            -- Synthetic authored data proves family reuse; it is not shipped
            -- gameplay coverage or a claim about another Diablo II skill.
            skill(900, "Fixture Bolt", "fixturebolt", "magic"),
        },
        ["data/global/excel/Missiles.txt"] = {
            missile("firebolt", "Fire Bolt", 20),
            missile("fixturebolt", "Fixture Bolt", 25, 75),
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
    },
})
