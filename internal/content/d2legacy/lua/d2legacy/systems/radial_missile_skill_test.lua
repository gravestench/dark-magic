local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local nova_skill = {
    Id = "48",
    skill = "Nova",
    skilldesc = "nova",
    leftskill = "1",
    -- The narrow fixture marks Nova generally available so the current
    -- record-driven entry bootstrap learns it without importing save data.
    general = "1",
    passive = "0",
    srvstfunc = "",
    srvdofunc = "22",
    cltstfunc = "",
    cltdofunc = "25",
    anim = "SC",
    range = "none",
    interrupt = "1",
    srvmissilea = "nova",
    srvmissileb = "nova",
    srvmissilec = "nova",
    EType = "ltng",
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
    Param1 = "12",
    Param2 = "4",
}

local nova_missile = {
    Missile = "nova",
    Skill = "Nova",
    pSrvDoFunc = "1",
    CollideType = "3",
    CollideKill = "",
    LastCollide = "1",
    NextHit = "1",
    NextDelay = "4",
    Vel = "24",
    Range = "13",
    Size = "1",
    CelFile = "ElectricNova",
    AnimSpeed = "16",
    NumDirections = "16",
    LoopAnim = "0",
    TravelSound = "sorceress_nova",
}

local player = fixtures.player_entry({
    x = 0,
    y = 0,
    mana = 20,
    max_mana = 20,
    skills = {
        {
            id = 48,
            level = 1,
            list_row = 0,
            left_allowed = true,
            right_allowed = true,
        },
    },
})

local monster = fixtures.monster_spawn({
    x = 3,
    y = 0,
    definition = {
        id = "fallen",
        base_id = "fallen",
        graphics_id = "fallen",
        name_key = "Fallen",
        ai = "fallen",
        token = "FA",
        weapon_class = "HTH",
        components = {},
        life_min = 20000 * 256,
        life_max = 20000 * 256,
        level = 1,
        defense = 0,
        attack_rating = 0,
        physical_min = 0,
        physical_max = 0,
        experience = 0,
        treasure_class = "",
        collider_radius = 0.5,
        select_radius = 0.5,
        velocity = 0,
        think_interval = 100,
        aggro_radius = 0,
        attack_range = 1,
    },
})

return test.suite({
    profile = "authority",
    tier = "integration",
    covers = { "internal/game/combat", "internal/game/skill", "internal/game/missile" },
    records = {
        ["data/global/excel/skills.txt"] = { nova_skill },
        ["data/global/excel/skilldesc.txt"] = {
            { skilldesc = "nova", ListRow = "0", IconCel = "0" },
        },
        ["data/global/excel/Missiles.txt"] = { nova_missile },
    },
    cases = {
        test.case("configured_radial_family_runs_headlessly", {
            test.submit_system(fixtures.command("system.player.enter", player, {
                tick = 1,
                sequence = 1,
            })),
            test.submit_system(fixtures.command("system.monster.spawn", monster, {
                tick = 1,
                sequence = 2,
            })),
            test.step(1),
            test.run(function()
                local authoritative = require("d2legacy.authoritative")
                local ecs = require("engine.ecs/v1")
                test.assert(authoritative.radial_missile_skills[48] ~= nil, [=[Nova radial definition is composed]=])
                local entered = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                test.expect(ecs.get(entered, "d2legacy.player.skill_assignment"):get("left")):equals(48)
            end),
            test.submit(fixtures.command("player.use_skill", {
                side = "left",
            }, {
                tick = 2,
                sequence = 1,
                player = "alice",
            })),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local projectiles = ecs.query({ all = { "d2legacy.missile.projectile" } })
                test.assert(#projectiles == 12, [=[#projectiles == 12]=])
                local cast_id
                local resident_ids = {}
                for _, entity in ipairs(projectiles) do
                    local projectile = ecs.get(entity, "d2legacy.missile.projectile")
                    cast_id = cast_id or projectile:get("cast_id")
                    test.expect(projectile:get("cast_id")):equals(cast_id)
                    test.expect(projectile:get("damage_channel")):equals("lightning")
                    test.expect(projectile:get("destroy_on_contact")):equals(false)
                    test.expect(projectile:get("next_hit_delay")):equals(4)
                    local resident = ecs.get(entity, "d2legacy.world.room_resident")
                    if resident then
                        test.assert(
                            not resident_ids[resident:get("resident_id")],
                            [=[radial residents have unique IDs]=]
                        )
                        resident_ids[resident:get("resident_id")] = true
                    end
                end
                local entered = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local vitals = ecs.get(entered, "d2legacy.player.vitals")
                test.assert(
                    vitals:get("mana_raw") == 5 * 256 and vitals:get("mana") == 5,
                    [=[Nova spends its level-one authored mana cost exactly once]=]
                )
            end),
            test.expect_checkpoint_parity(1),
            test.step(3),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local events = ecs.query({ all = { "d2legacy.combat.event" } })
                test.assert(#events == 1, [=[one cast damages one target at most once during the lock]=])
                local event = ecs.get(events[1], "d2legacy.combat.event")
                test.assert(
                    event:get("attacker_id") == "player:alice"
                        and event:get("target_id") == "monster:fallen"
                        and event:get("source_kind") == "missile"
                        and event:get("damage_channel") == "lightning"
                        and event:get("rolled_damage_raw") >= 256
                        and event:get("rolled_damage_raw") <= 20 * 256,
                    [=[radial contact emits one shared lightning damage result]=]
                )
                test.assert(
                    #ecs.query({ all = { "d2legacy.missile.contact_lock" } }) == 1,
                    [=[the cast-target repeat-contact lock is checkpointed as an ECS entity]=]
                )
            end),
            test.step(10),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.assert(
                    #ecs.query({ all = { "d2legacy.missile.projectile" } }) == 0,
                    [=[radial projectiles expire from their record-authored lifetime]=]
                )
                test.assert(
                    #ecs.query({ all = { "d2legacy.missile.contact_lock" } }) == 0,
                    [=[expired contact locks leave no simulation entities]=]
                )
            end),
        }),
    },
})
