local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local teleport = {
    Id = "54",
    skill = "Teleport",
    skilldesc = "teleport",
    srvstfunc = "",
    srvdofunc = "27",
    cltstfunc = "",
    cltdofunc = "",
    anim = "SC",
    range = "none",
    warp = "1",
    interrupt = "1",
    leftskill = "",
    general = "",
    InGame = "1",
    mana = "24",
    lvlmana = "-1",
    manashift = "8",
    minmana = "1",
}

local function collision(line_clear)
    return {
        blocked_position = function()
            return false
        end,
        line_clear = function()
            return line_clear
        end,
    }
end

-- Teleport traces and forced-motion integration over open ground.
local function open_collision()
    return {
        blocked_position = function()
            return false
        end,
        line_clear = function()
            return true
        end,
        integrate_velocity = function(_, x, y, velocity_x, velocity_y, _, _, _, elapsed)
            return x + velocity_x * elapsed, y + velocity_y * elapsed
        end,
    }
end

local function learn_teleport()
    local ecs = require("engine.ecs/v1")
    local players = ecs.query({ all = { "d2legacy.player.identity" } })
    test.assert(#players == 1, [=[one player exists]=])
    ecs.create({
        ["d2legacy.player.learned_skill"] = {
            owner = players[1],
            skill_id = 54,
            level = 1,
            list_row = 0,
            left_allowed = false,
            right_allowed = true,
        },
    })
end

local function entry(level_id)
    return fixtures.player_entry({
        x = 2,
        y = 2,
        level_id = level_id,
        mana = 50,
        max_mana = 50,
    })
end

return test.suite({
    profile = "authority",
    tier = "integration",
    covers = { "internal/game/skill", "internal/game/world" },
    records = {
        ["data/global/excel/skills.txt"] = { teleport },
        ["data/global/excel/skilldesc.txt"] = {
            { skilldesc = "teleport", ListRow = "0", IconCel = "0" },
        },
        ["data/global/excel/Levels.txt"] = {
            { Id = "0", Teleport = "0" },
            { Id = "2", Teleport = "1" },
            { Id = "73", Teleport = "2" },
        },
    },
    cases = {
        test.case("point_relocation_is_atomic_and_uses_signed_mana_progression", {
            test.run(function()
                require("d2legacy.gameplay.systems.movement").set_collision(2, collision(false))
            end),
            test.submit_system(fixtures.command("system.player.enter", entry(2), { tick = 1, sequence = 1 })),
            test.step(1),
            test.run(learn_teleport),
            test.submit(fixtures.command("player.assign_skills", { right = 54 }, {
                tick = 2,
                sequence = 1,
                player = "alice",
            })),
            test.step(1),
            test.submit(fixtures.command("player.use_skill", {
                side = "right",
                target_x = 8,
                target_y = 9,
            }, {
                tick = 3,
                sequence = 2,
                player = "alice",
            })),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local position = ecs.get(player, "d2legacy.world.position")
                test.expect(position:get("x")):equals(8)
                test.expect(position:get("y")):equals(9)
                test.expect(ecs.get(player, "d2legacy.player.vitals"):get("mana_raw")):equals(26 * 256)
                local events = ecs.query({ all = { "d2legacy.world.relocation_event" } })
                test.assert(#events == 1, [=[one relocation fact is emitted]=])
                local event = ecs.get(events[1], "d2legacy.world.relocation_event")
                test.expect(event:get("outcome")):equals("completed")
                test.expect(event:get("source_id")):equals("skill:player:alice:54")
            end),
            test.expect_checkpoint_parity(1),
        }),
        -- Teleport already ends an active knockback. It must record where the
        -- knockback stopped, before the relocation moves the player.
        test.case("completed_relocation_interrupts_active_forced_motion", {
            test.run(function()
                require("d2legacy.gameplay.systems.movement").set_collision(2, open_collision())
            end),
            test.submit_system(fixtures.command("system.player.enter", entry(2), { tick = 1, sequence = 1 })),
            test.step(1),
            test.run(learn_teleport),
            test.submit(fixtures.command("player.assign_skills", { right = 54 }, {
                tick = 2,
                sequence = 1,
                player = "alice",
            })),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                ecs.set(player, "d2legacy.world.forced_motion_request", {
                    kind = "knockback",
                    source_x = 0,
                    source_y = 2,
                    distance = 20,
                    speed = 25,
                    request_tick = 2,
                })
            end),
            test.submit(fixtures.command("player.use_skill", {
                side = "right",
                target_x = 8,
                target_y = 9,
            }, {
                tick = 3,
                sequence = 2,
                player = "alice",
            })),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local position = ecs.get(player, "d2legacy.world.position")
                test.expect(position:get("x")):equals(8)
                test.expect(position:get("y")):equals(9)
                local relocations = ecs.query({ all = { "d2legacy.world.relocation_event" } })
                test.assert(#relocations == 1, [=[one relocation fact is emitted]=])
                test.expect(ecs.get(relocations[1], "d2legacy.world.relocation_event"):get("outcome"))
                    :equals("completed")
                test.assert(not ecs.get(player, "d2legacy.world.forced_motion"), [=[relocation kept forced motion]=])
                local events = ecs.query({ all = { "d2legacy.world.forced_motion_event" } })
                test.expect(#events):equals(1)
                local event = ecs.get(events[1], "d2legacy.world.forced_motion_event")
                test.expect(event:get("outcome")):equals("interrupted")
                test.expect(event:get("target_id")):equals("player:alice")
                test.expect(event:get("start_x")):equals(2)
                test.expect(event:get("end_y")):equals(2)
                -- Admitted on tick 3 and moved one subtile; Teleport completes
                -- in tick 4's pre-simulation, before any further movement.
                test.expect(event:get("applied_distance")):equals(1)
                test.expect(event:get("end_x")):equals(3)
            end),
            test.expect_checkpoint_parity(1),
        }),
        test.case("limited_level_policy_requires_a_clear_trace", {
            test.run(function()
                require("d2legacy.gameplay.systems.movement").set_collision(73, collision(false))
            end),
            test.submit_system(fixtures.command("system.player.enter", entry(73), { tick = 1, sequence = 1 })),
            test.step(1),
            test.run(learn_teleport),
            test.submit(fixtures.command("player.assign_skills", { right = 54 }, {
                tick = 2,
                sequence = 1,
                player = "alice",
            })),
            test.step(1),
            test.submit(fixtures.command("player.use_skill", {
                side = "right",
                target_x = 8,
                target_y = 9,
            }, {
                tick = 3,
                sequence = 2,
                player = "alice",
            })),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local position = ecs.get(player, "d2legacy.world.position")
                test.expect(position:get("x")):equals(2)
                test.expect(position:get("y")):equals(2)
                local event = ecs.get(
                    ecs.query({ all = { "d2legacy.world.relocation_event" } })[1],
                    "d2legacy.world.relocation_event"
                )
                test.expect(event:get("outcome")):equals("line-blocked")
            end),
        }),
    },
})
