local test = require("d2legacy.tests/v1")

local function skill(id, name, missile, channel, count, count_per_level)
    return {
        Id = tostring(id),
        skill = name,
        srvstfunc = "",
        srvdofunc = "22",
        cltstfunc = "",
        cltdofunc = "25",
        anim = "SC",
        range = "none",
        interrupt = "1",
        srvmissilea = missile,
        srvmissileb = missile,
        srvmissilec = missile,
        EType = channel,
        HitShift = "8",
        mana = "15",
        lvlmana = "1",
        minmana = "1",
        manashift = "8",
        EMin = "1",
        EMax = "20",
        EMinLev1 = "6",
        EMinLev2 = "7",
        EMinLev3 = "8",
        EMinLev4 = "9",
        EMinLev5 = "10",
        EMaxLev1 = "8",
        EMaxLev2 = "9",
        EMaxLev3 = "10",
        EMaxLev4 = "11",
        EMaxLev5 = "12",
        Param1 = tostring(count),
        Param2 = tostring(count_per_level),
    }
end

local function missile(id, skill_name, velocity, lifetime)
    return {
        Missile = id,
        Skill = skill_name,
        pSrvDoFunc = "1",
        CollideType = "3",
        CollideKill = "",
        LastCollide = "1",
        NextHit = "1",
        NextDelay = "4",
        Vel = tostring(velocity),
        Range = tostring(lifetime),
        Size = "1",
        CelFile = id,
        AnimSpeed = "16",
        NumDirections = "16",
        LoopAnim = "0",
    }
end

return test.suite({
    name = "radial missile skill definitions",
    profile = "module",
    tier = "fast",
    records = {
        ["data/global/excel/skills.txt"] = {
            skill(48, "Nova", "nova", "ltng", 12, 4),
            -- Synthetic authored data proves family reuse; it is not shipped
            -- gameplay coverage or a claim about another Diablo II skill.
            skill(900, "Fixture Ring", "fixturering", "cold", 8, 2),
        },
        ["data/global/excel/Missiles.txt"] = {
            missile("nova", "Nova", 24, 13),
            missile("fixturering", "Fixture Ring", 25, 20),
        },
    },
    cases = {
        test.case("decodes_multiple_configurations_without_skill_branches", function(t)
            t:run(function()
                local definitions = require("d2legacy.data.radial_missile_skills").load({ 48, 900 })
                local nova = definitions[48]
                test.expect(nova.behavior):equals("missile.radial")
                test.expect(nova.missile_id):equals("nova")
                test.expect(nova.damage_channel):equals("lightning")
                test.expect(nova.missile_count_base):equals(12)
                test.expect(nova.missile_count_per_level):equals(4)
                test.expect(nova.minimum_damage_per_level_raw[5]):equals(10 * 256)
                test.expect(nova.maximum_damage_per_level_raw[5]):equals(12 * 256)
                test.expect(nova.next_hit_delay):equals(4)
                test.expect(nova.destroy_on_contact):equals(false)

                local fixture = definitions[900]
                test.expect(fixture.behavior):equals("missile.radial")
                test.expect(fixture.missile_id):equals("fixturering")
                test.expect(fixture.damage_channel):equals("cold")
                test.expect(fixture.missile_count_base):equals(8)
                test.expect(fixture.missile_count_per_level):equals(2)
                test.expect(fixture.speed_per_tick):equals(1)
                test.expect(fixture.lifetime_ticks):equals(20)
            end)
        end),
    },
})
