local test = require("d2legacy.tests/v1")

local fixtures = require("d2legacy.tests.support.fixtures")

local function experience_row(level, amazon)
    return {
        Level = level,
        Amazon = amazon,
        Sorceress = amazon,
        Necromancer = amazon,
        Paladin = amazon,
        Barbarian = amazon,
        Druid = amazon,
        Assassin = amazon,
    }
end

local function enter(level, experience)
    return test.submit_system(
        fixtures.command(
            "system.player.enter",
            fixtures.player_entry({ level = level, experience = experience }),
            { tick = 1, sequence = 1 }
        )
    )
end

-- Kill alice with a killer whose credit resolves through ownership, if any.
local function kill(owner_kind)
    return test.run(function()
        local ecs = require("engine.ecs/v1")
        local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
        ecs.get(player, "d2legacy.player.vitals"):set("health", 0)
        local killer = { id = "monster:fallen", kind = "hostile", label = "Fallen", owner = "", radius = 0.5 }
        local components = { ["d2legacy.world.selectable"] = killer }
        if owner_kind then
            local owner = ecs.create({
                ["d2legacy.world.selectable"] = {
                    id = "owner:bob",
                    kind = owner_kind,
                    label = "Bob",
                    owner = "bob",
                    radius = 0.75,
                },
            })
            killer.kind = "friendly"
            components["d2legacy.owned_unit"] = {
                owner = owner,
                owner_id = "owner:bob",
                ultimate_owner_id = "owner:bob",
                category = "wolf",
                group = 1,
                limit = 1,
                replacement = "replace_oldest",
                created_tick = 1,
                active = true,
            }
        end
        ecs.create(components)
        ecs.create({
            ["d2legacy.combat.event"] = {
                kind = "unit_died",
                tick = 1,
                attacker_id = "monster:fallen",
                target_id = "player:alice",
                source_kind = "melee",
                damage_channel = "physical",
            },
        })
    end)
end

local function expect_loss(experience, lost)
    return test.run(function()
        local ecs = require("engine.ecs/v1")
        local player = ecs.query({ all = { "d2legacy.player.death" } })[1]
        test.expect(ecs.get(player, "d2legacy.player.progress"):get("experience")):equals(experience)
        test.expect(ecs.get(player, "d2legacy.player.death"):get("experience_lost")):equals(lost)
        local event = ecs.query({ all = { "d2legacy.player.death_event" } })[1]
        test.expect(ecs.get(event, "d2legacy.player.death_event"):get("experience_lost")):equals(lost)
    end)
end

return test.suite({
    profile = "authority",
    tier = "fast",
    covers = { "internal/game/player" },
    initial_data = {
        ["d2legacy.game_rules"] = {
            target = "lod-1.14d",
            expansion = true,
            difficulty = 2,
        },
    },
    records = {
        ["data/global/excel/experience.txt"] = {
            experience_row("0", "0"),
            experience_row("1", "500"),
            experience_row("2", "1500"),
            experience_row("3", "3750"),
        },
    },
    cases = {
        -- Level 3 spans 1500..3750; Hell removes 10% of 2250.
        test.case("monster_kill_costs_a_percentage_of_the_level_span", {
            enter(3, 2000),
            test.step(1),
            kill(nil),
            test.step(1),
            expect_loss(1775, 225),
        }),
        test.case("loss_never_drops_below_the_start_of_the_level", {
            enter(3, 1600),
            test.step(1),
            kill(nil),
            test.step(1),
            expect_loss(1501, 100),
        }),
        test.case("player_owned_killer_costs_nothing", {
            enter(3, 2000),
            test.step(1),
            kill("player"),
            test.step(1),
            expect_loss(2000, 0),
        }),
        test.case("first_level_costs_nothing", {
            enter(1, 300),
            test.step(1),
            kill(nil),
            test.step(1),
            expect_loss(300, 0),
        }),
    },
})
