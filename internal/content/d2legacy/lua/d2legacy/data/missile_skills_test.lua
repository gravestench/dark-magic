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

local function missile(name, skill_name, velocity)
    return {
        Missile = name,
        Skill = skill_name,
        pSrvDoFunc = "1",
        CollideType = "3",
        CollideKill = "1",
        Vel = tostring(velocity),
        Range = "40",
        Size = "2",
        CelFile = name,
    }
end

return test.suite({
    name = "straight missile skill definitions",
    profile = "module",
    tier = "fast",
    records = {
        ["data/global/excel/skills.txt"] = {
            skill(36, "Fire Bolt", "firebolt", "fire"),
            -- Synthetic authored data proves family reuse; it is not shipped
            -- gameplay coverage or a claim about another Diablo II skill.
            skill(900, "Fixture Bolt", "fixturebolt", "magic"),
        },
        ["data/global/excel/Missiles.txt"] = {
            missile("firebolt", "Fire Bolt", 20),
            missile("fixturebolt", "Fixture Bolt", 25),
        },
    },
    cases = {
        test.case("decodes_multiple_configurations_without_skill_branches", function(t)
            t:run(function()
                local definitions = require("d2legacy.data.missile_skills").load({ 36, 900 })
                test.expect(definitions[36].behavior):equals("missile.straight")
                test.expect(definitions[36].missile_id):equals("firebolt")
                test.expect(definitions[36].damage_channel):equals("fire")
                test.expect(definitions[900].behavior):equals("missile.straight")
                test.expect(definitions[900].missile_id):equals("fixturebolt")
                test.expect(definitions[900].damage_channel):equals("magic")
                test.expect(definitions[900].speed_per_tick):equals(1)
            end)
        end),
    },
})
