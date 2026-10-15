local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

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
        cltmissilea = "blade-fury-visual",
        SrcDam = "96",
        startmana = "6",
        mana = "8",
        manashift = "5",
        Param3 = "3",
        Param4 = "5",
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
    skill(900, "Charged Helper", {
        srvdofunc = "17",
        srvmissile = "charged-shot",
        calc4 = "par1+skill('Fire Blast'.blvl)/3",
        Param1 = "3",
        Param2 = "1",
    }),
    skill(901, "Wake Fire Helper", { srvdofunc = "17", srvmissile = "wake-fire-shot", Param1 = "1", Param2 = "0" }),
    skill(902, "Lightning Helper", {
        srvdofunc = "19",
        srvmissile = "lightning-shot",
        calc4 = "skill('Death Sentry'.blvl)/3",
        Param1 = "1",
        Param2 = "0",
    }),
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
    missile("blade-sentinel-missile", { NextHit = "1", NextDelay = "25", CollideKill = "" }),
    missile("blade-fury-missile", { Vel = "22", Range = "40" }),
    missile("blade-fury-visual", { Vel = "22", Range = "33" }),
    missile("charged-shot", { Vel = "12", Range = "98" }),
    missile("wake-fire-shot"),
    missile("lightning-shot", { Vel = "30", Range = "98" }),
    missile("inferno-shot"),
    missile("death-shot"),
    missile("blade-shield-attachment", { CltSubMissile1 = "blade-shield-overlay" }),
}

local function fury_start_case(mana, expected)
    return test.case("blade_fury_start_at_" .. mana .. "_mana", {
        test.submit_system(fixtures.command(
            "system.player.enter",
            fixtures.player_entry({
                x = 0,
                y = 0,
                mana = mana,
                max_mana = mana,
            }),
            { tick = 1, sequence = 1 }
        )),
        test.step(1),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
            ecs.get(player, "d2legacy.player.resource_stats"):set("mana_regen_frames", 0)
            ecs.create({
                ["d2legacy.player.learned_skill"] = {
                    owner = player,
                    skill_id = 266,
                    level = 1,
                    list_row = 1,
                    left_allowed = true,
                    right_allowed = true,
                },
            })
        end),
        test.submit(fixtures.command("player.assign_skills", { left = 266 }, {
            tick = 2,
            sequence = 1,
            player = "alice",
        })),
        test.submit(fixtures.command("player.use_skill", { side = "left", target_x = 8, target_y = 0 }, {
            tick = 2,
            sequence = 2,
            player = "alice",
        })),
        test.step(1),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
            local started = ecs.get(player, "d2legacy.skill.cast") ~= nil
            local remaining = ecs.get(player, "d2legacy.player.vitals"):get("mana_raw")
            test.assert(
                started == expected,
                string.format(
                    "Blade Fury: initial mana=%d, cast started=%s, remaining mana=%g; expected started=%s",
                    mana,
                    tostring(started),
                    remaining / 256,
                    tostring(expected)
                )
            )
            test.expect(remaining):equals((mana - (expected and 1 or 0)) * 256)
        end),
    })
end

local function sentinel_contact_case(name, options)
    options = options or {}
    local function create_target()
        local ecs = require("engine.ecs/v1")
        ecs.create({
            ["d2legacy.monster.stats"] = { level = 1, health = 100 * 256, max_health = 100 * 256 },
            ["d2legacy.world.position"] = { x = 1, y = 0 },
            ["d2legacy.world.location"] = { act = 1, level_id = 1 },
            ["d2legacy.world.selectable"] = { id = "monster:sentinel-target", kind = "hostile", radius = 0.5 },
        })
    end
    local function spawn_sentinel(cast_id, next_hit)
        local ecs = require("engine.ecs/v1")
        local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
        local definition = require("d2legacy.data.trap_skills").load({ 257 })[257]
        if next_hit == false then
            definition.next_hit = false
            definition.destroy_on_contact = true
        end
        if options.zero_delay then
            definition.next_hit_delay = 0
            definition.destroy_on_contact = true
        end
        ecs.create(require("d2legacy.gameplay.projectile_spawn").components(player, definition, {
            cast_id = cast_id,
            velocity_x = 1,
            velocity_y = 0,
            skill_level = 1,
            minimum_damage_raw = 256,
            maximum_damage_raw = 256,
        }))
    end
    return test.case("blade_sentinel_contact_" .. name, {
        test.submit_system(fixtures.command("system.player.enter", fixtures.player_entry({ x = 0, y = 0 }), {
            tick = 1,
            sequence = 1,
        })),
        test.step(1),
        test.run(function()
            create_target()
            spawn_sentinel("sentinel-cast-1", options.first_next_hit)
            if options.same_tick then
                spawn_sentinel("sentinel-cast-2")
            end
        end),
        test.step(1),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            local target = ecs.query({ all = { "d2legacy.monster.stats" } })[1]
            test.expect(ecs.get(target, "d2legacy.monster.stats"):get("health")):equals(99 * 256)
            test.expect(#ecs.query({ all = { "d2legacy.missile.contact_lock" } }))
                :equals(options.first_next_hit == false and 0 or 1)
        end),
        test.expect_checkpoint_parity(options.wait_ticks or 1),
        test.run(function()
            if options.replace_target then
                local ecs = require("engine.ecs/v1")
                ecs.destroy(ecs.query({ all = { "d2legacy.monster.stats" } })[1])
                create_target()
            end
            spawn_sentinel(options.same_cast and "sentinel-cast-1" or "sentinel-cast-2", options.second_next_hit)
        end),
        test.step(1),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            local target = ecs.query({ all = { "d2legacy.monster.stats" } })[1]
            local health = ecs.get(target, "d2legacy.monster.stats"):get("health") / 256
            local expected = options.expected_health or 99
            test.assert(
                health == expected,
                string.format(
                    "Blade Sentinel %s: target health=%g; expected=%d (each contact deals 1)",
                    name,
                    health,
                    expected
                )
            )
        end),
    })
end

return test.suite({
    name = "complete Assassin trap definitions",
    profile = "authority",
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
        fury_start_case(2, false),
        fury_start_case(4, false),
        fury_start_case(6, true),
        sentinel_contact_case("zero_delay_blocks_only_the_contact_tick", {
            same_tick = true,
            zero_delay = true,
            expected_health = 97,
        }),
        sentinel_contact_case("same_cast", { same_cast = true }),
        sentinel_contact_case("independent_casts"),
        sentinel_contact_case("replacement_target_with_same_public_id", { replace_target = true }),
        sentinel_contact_case("independent_casts_same_tick", { same_tick = true }),
        sentinel_contact_case("before_cooldown_expiry", { wait_ticks = 23 }),
        sentinel_contact_case("at_cooldown_expiry", { wait_ticks = 24, expected_health = 98 }),
        sentinel_contact_case("ordinary_missile_ignores_cooldown", { second_next_hit = false, expected_health = 98 }),
        sentinel_contact_case(
            "ordinary_missile_does_not_arm_cooldown",
            { first_next_hit = false, expected_health = 98 }
        ),
        test.case("uses_blade_fury_server_missile_instead_of_client_visual", function(t)
            t:run(function()
                local definition = require("d2legacy.data.trap_skills").load({ 266 })[266]
                test.expect(definition.missile_id):equals("blade-fury-missile")
                test.expect(definition.speed_per_tick):equals(67584 / 65536)
                test.expect(definition.lifetime_ticks):equals(40)
                test.expect(definition.complete_delay):equals(6)
            end)
        end),
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
                test.expect(definitions[261].speed_per_tick):equals(36864 / 65536)
                test.expect(definitions[271].speed_per_tick):equals(92160 / 65536)
                test.expect(definitions[256].speed_per_tick):equals(0)
                test.expect(definitions[261].lifetime_ticks):equals(77)
                test.expect(definitions[262].lifetime_ticks):equals(25)
                test.expect(definitions[271].lifetime_ticks):equals(98)
                test.expect(definitions[251].field_count_base):equals(nil)
                test.expect(definitions[256].field_count_synergy_skill_id):equals(261)
                test.expect(definitions[257].weapon_fraction):equals(96)
                test.expect(definitions[257].next_hit):equals(true)
                test.expect(definitions[257].next_hit_delay):equals(25)
                test.expect(definitions[266].next_hit):equals(false)
                test.expect(definitions[266].minimum_start_mana_raw):equals(6 * 256)
                test.expect(definitions[276].operation):equals("corpse_or_projectile")
                test.expect(definitions[277].requires_point_target):equals(false)
            end)
        end),
        test.case("charged_helper_lifetime_reaches_projectile_expiry_and_checkpoint", {
            test.submit_system(fixtures.command("system.player.enter", fixtures.player_entry(), {
                tick = 1,
                sequence = 1,
            })),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local definition = require("d2legacy.data.trap_skills").load({ 261 })[261]
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local spawn = require("d2legacy.gameplay.projectile_spawn")
                ecs.create(spawn.components(player, definition, {
                    cast_id = "charged-lifetime",
                    velocity_x = 0,
                    velocity_y = 0,
                    skill_level = 1,
                }))
                local projectile = ecs.query({ all = { "d2legacy.missile.projectile" } })[1]
                test.expect(ecs.get(projectile, "d2legacy.missile.projectile"):get("remaining_ticks")):equals(77)
            end),
            test.step(76),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local projectiles = ecs.query({ all = { "d2legacy.missile.projectile" } })
                test.expect(#projectiles):equals(1)
                test.expect(ecs.get(projectiles[1], "d2legacy.missile.projectile"):get("remaining_ticks")):equals(1)
            end),
            test.expect_checkpoint_parity(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.expect(#ecs.query({ all = { "d2legacy.missile.projectile" } })):equals(0)
            end),
        }),
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
