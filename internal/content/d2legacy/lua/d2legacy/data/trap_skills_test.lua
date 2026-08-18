local test = require("d2legacy.tests/v1")

local function copy(values)
    local result = {}
    for key, value in pairs(values) do
        result[key] = value
    end
    return result
end

local function skill(id, name, values)
    local row = {
        Id = tostring(id),
        skill = name,
        skilldesc = string.lower(name),
        charclass = "ass",
        InGame = "1",
        range = "rng",
        mana = "6",
        lvlmana = "1",
        minmana = "1",
        manashift = "8",
        HitShift = "8",
        EType = "fire",
        EMin = "1",
        EMax = "2",
        EMinLev1 = "1",
        EMinLev2 = "2",
        EMinLev3 = "3",
        EMinLev4 = "4",
        EMinLev5 = "5",
        EMaxLev1 = "2",
        EMaxLev2 = "3",
        EMaxLev3 = "4",
        EMaxLev4 = "5",
        EMaxLev5 = "6",
        EDmgSymPerCalc = "",
    }
    for key, value in pairs(values or {}) do
        row[key] = value
    end
    return row
end

local function missile(name, values)
    local row = {
        Missile = name,
        pSrvDoFunc = "1",
        Range = "25",
        Vel = "25",
        Size = "2",
        CollideKill = "1",
        CelFile = "fixture",
        NumDirections = "1",
        AnimSpeed = "16",
        LoopAnim = "",
        Trans = "0",
    }
    for key, value in pairs(values or {}) do
        row[key] = value
    end
    return row
end

local fire_blast_synergy = table.concat({
    "(skill('Shock Web'.blvl)+skill('Charged Bolt Sentry'.blvl)",
    "+skill('Wake of Fire Sentry'.blvl)+skill('Lightning Sentry'.blvl)",
    "+skill('Wake of Inferno'.blvl)+skill('Death Sentry'.blvl))*par8",
})

local skills = {
    skill(251, "Fire Blast", {
        lob = "1",
        srvmissile = "fire-air",
        Param1 = "4",
        EDmgSymPerCalc = fire_blast_synergy,
        Param8 = "9",
    }),
    skill(256, "Shock Web", {
        srvdofunc = "43",
        srvmissilea = "web-air",
        prgcalc1 = "par1+skill('Charged Bolt Sentry'.blvl)/3",
        Param1 = "5",
        Param2 = "4",
        delay = "8",
    }),
    skill(257, "Blade Sentinel", {
        srvdofunc = "44",
        summon = "blade-sentinel",
        pettype = "trap",
        srvmissilea = "blade-sentinel-missile",
        SrcDam = "96",
        Param1 = "100",
        Param2 = "10",
        petmax = "5",
        delay = "6",
    }),
    skill(261, "Charged Bolt Sentry", {
        srvdofunc = "45",
        summon = "charged-sentry",
        pettype = "trap",
        sumskill1 = "Charged Helper",
        Param1 = "5",
        petmax = "5",
    }),
    skill(262, "Wake of Fire Sentry", {
        srvdofunc = "45",
        summon = "wake-fire-sentry",
        pettype = "trap",
        sumskill1 = "Wake Fire Helper",
        Param1 = "5",
        petmax = "5",
    }),
    skill(266, "Blade Fury", {
        srvstfunc = "26",
        srvdofunc = "48",
        ["repeat"] = "1",
        usemanaondo = "1",
        srvmissilea = "blade-fury-missile",
        SrcDam = "96",
        Param3 = "6",
        Param4 = "3",
    }),
    skill(271, "Lightning Sentry", {
        srvdofunc = "45",
        summon = "lightning-sentry",
        pettype = "trap",
        sumskill1 = "Lightning Helper",
        Param1 = "10",
        petmax = "5",
    }),
    skill(272, "Wake of Inferno", {
        srvdofunc = "45",
        summon = "inferno-sentry",
        pettype = "trap",
        sumskill1 = "Inferno Helper",
        Param1 = "10",
        petmax = "5",
        Param7 = "7",
        Param8 = "10",
        EDmgSymPerCalc = "skill('Wake of Fire Sentry'.blvl)*par8+skill('Death Sentry'.blvl)*par7",
    }),
    skill(276, "Death Sentry", {
        srvdofunc = "45",
        summon = "death-sentry",
        pettype = "trap",
        sumskill1 = "Corpse Helper",
        sumskill2 = "Death Lightning Helper",
        Param1 = "5",
        petmax = "5",
    }),
    skill(277, "Blade Shield", {
        srvstfunc = "28",
        srvdofunc = "54",
        periodic = "1",
        aurastate = "blade-shield",
        srvmissilea = "blade-shield-attachment",
        SrcDam = "64",
        Param1 = "500",
        Param2 = "25",
        Param3 = "25",
        Param4 = "5",
        range = "none",
    }),
    skill(
        900,
        "Charged Helper",
        { srvmissile = "charged-shot", calc4 = "par1+skill('Fire Blast'.blvl)/3", Param1 = "3", Param2 = "1" }
    ),
    skill(901, "Wake Fire Helper", { srvmissile = "wake-fire-shot", Param1 = "1", Param2 = "0" }),
    skill(
        902,
        "Lightning Helper",
        { srvmissile = "lightning-shot", calc4 = "skill('Death Sentry'.blvl)/3", Param1 = "1", Param2 = "0" }
    ),
    skill(903, "Inferno Helper", { srvmissile = "inferno-shot", Param1 = "1", Param2 = "0" }),
    skill(
        904,
        "Corpse Helper",
        { TargetCorpse = "1", Param1 = "40", Param2 = "80", Param3 = "6", Param4 = "1", calc3 = "50" }
    ),
    skill(905, "Death Lightning Helper", { srvmissile = "death-shot", Param1 = "1", Param2 = "0" }),
}

local descriptions = {}
for _, row in ipairs(skills) do
    descriptions[#descriptions + 1] = {
        skilldesc = row.skilldesc,
        ["str name"] = "fixture-name",
        dsc3texta1 = "Sksyn",
        dsc3texta2 = "fixture-synergy",
    }
end

local missiles = {
    missile("fire-air", { HitSubMissile1 = "fire-ground" }),
    missile("fire-ground", { ExplosionMissile = "fire-impact" }),
    missile("fire-impact"),
    missile("web-air", { HitSubMissile1 = "web-ground" }),
    missile("web-ground", { Vel = "0", CollideKill = "" }),
    missile("blade-sentinel-missile"),
    missile("blade-fury-missile"),
    missile("charged-shot"),
    missile("wake-fire-shot"),
    missile("lightning-shot"),
    missile("inferno-shot"),
    missile("death-shot"),
    missile("blade-shield-attachment", { CltSubMissile1 = "blade-shield-overlay" }),
}

return test.suite({
    name = "complete Assassin trap definitions",
    profile = "module",
    tier = "fast",
    records = {
        ["data/global/excel/skills.txt"] = skills,
        ["data/global/excel/Missiles.txt"] = missiles,
        ["data/global/excel/PetType.txt"] = { { ["pet type"] = "trap", group = "1", range = "1" } },
        ["data/global/excel/MonStats.txt"] = {
            { Id = "charged-sentry", MonStatsEx = "charged-gfx", AI = "AssassinSentry", aip3 = "20", aip4 = "15" },
            { Id = "wake-fire-sentry", MonStatsEx = "wake-gfx", AI = "AssassinSentry", aip3 = "20", aip4 = "15" },
            { Id = "lightning-sentry", MonStatsEx = "lightning-gfx", AI = "AssassinSentry", aip3 = "20", aip4 = "15" },
            { Id = "inferno-sentry", MonStatsEx = "inferno-gfx", AI = "AssassinSentry", aip3 = "20", aip4 = "15" },
            { Id = "death-sentry", MonStatsEx = "death-gfx", AI = "DeathSentry", aip3 = "20", aip4 = "15" },
        },
        ["data/global/excel/MonStats2.txt"] = {
            { Id = "charged-gfx" },
            { Id = "wake-gfx" },
            { Id = "lightning-gfx" },
            { Id = "inferno-gfx" },
            { Id = "death-gfx" },
        },
        ["data/global/excel/SkillDesc.txt"] = descriptions,
    },
    cases = {
        test.case("decodes_all_ten_exact_ids_by_six_record_shapes", function(t)
            t:run(function()
                local ids = { 251, 256, 257, 261, 262, 266, 271, 272, 276, 277 }
                local definitions = require("d2legacy.data.trap_skills").load(ids)
                local shapes = {
                    [251] = "lobbed_payload",
                    [256] = "persistent_field",
                    [257] = "returning_weapon_patrol",
                    [261] = "stationary_sentry",
                    [262] = "stationary_sentry",
                    [266] = "repeat_weapon_missile",
                    [271] = "stationary_sentry",
                    [272] = "stationary_sentry",
                    [276] = "stationary_sentry",
                    [277] = "periodic_weapon_state",
                }
                for _, id in ipairs(ids) do
                    test.expect(definitions[id].skill_id):equals(id)
                    test.expect(definitions[id].shape):equals(shapes[id])
                end
                test.expect(definitions[251].field_count_base):equals(nil)
                test.expect(definitions[256].field_count_synergy_skill_id):equals(261)
                test.expect(definitions[257].weapon_fraction):equals(96)
                test.expect(definitions[266].minimum_start_mana_raw):equals(6 * 256)
                test.expect(definitions[276].operation):equals("corpse_or_projectile")
                test.expect(definitions[277].requires_point_target):equals(false)
            end)
        end),
        test.case("preserves_heterogeneous_damage_synergy_rates", function(t)
            t:run(function()
                local definition = require("d2legacy.data.trap_skills").load({ 272 })[272]
                test.expect(#definition.damage_synergy_terms):equals(2)
                test.expect(definition.damage_synergy_terms[1].skill_id):equals(262)
                test.expect(definition.damage_synergy_terms[1].percent):equals(10)
                test.expect(definition.damage_synergy_terms[2].skill_id):equals(276)
                test.expect(definition.damage_synergy_terms[2].percent):equals(7)
            end)
        end),
    },
})
