local test = require("d2legacy.tests/v1")

local function fixture()
    for _, name in ipairs({
        "engine.input/v1",
        "engine.scene/v1",
        "engine.vfs/v1",
        "d2legacy.save/v1",
        "d2legacy.ui.game_hud",
        "d2legacy.gameplay.player_composite",
        "d2legacy.gameplay.monster_composite",
        "d2legacy.gameplay.cast_presentation",
        "d2legacy.gameplay.held_skill_input",
        "d2legacy.gameplay.state_overlay_presentation",
        "d2legacy.gameplay.interaction_approach",
        "d2legacy.data.skill",
        "d2legacy.ui.tooltip",
        "d2legacy.ui.text",
    }) do
        test.mock_module(name, {})
    end
    local function node()
        return {
            set_visible = function() end,
            set_position = function() end,
            set_z = function() end,
            destroy = function() end,
            exists = function()
                return true
            end,
        }
    end
    test.mock_module("engine.render/v1", {
        assets_available = function()
            return true
        end,
        create = node,
    })
    test.mock_module("engine.data/v1", {
        load_manifest = function()
            return { resolution = { width = 800, height = 600 }, screens = { game_world = { map = {} } } }
        end,
    })
    test.mock_module("engine.network/v1", {
        status = function()
            return { player_id = "local-player" }
        end,
    })
    test.mock_module("d2legacy.gameplay.missile_presentation", {
        resolve = function()
            return nil
        end,
    })
    test.mock_module("d2legacy.gameplay.aura_overlay_cycle", {
        select = function()
            return {}, {}
        end,
    })
    test.mock_module("d2legacy.presentation.chunked_map", {
        destroy = function() end,
        create = function()
            return { root = node() }
        end,
    })
    test.mock_module("engine.records/v1", {
        load = function()
            return {
                { Sound = "travel", FileName = "loop.wav", Loop = "1", ["Volume Min"] = "255" },
                { Sound = "once", FileName = "once.wav", Loop = "0", ["Volume Min"] = "255" },
            }
        end,
    })
    local played = {}
    test.mock_module("engine.audio/v1", {
        exists = function()
            return true
        end,
        play = function(path, options)
            local sound = { path = path, loop = options.loop, stopped = false }
            function sound:stop()
                self.stopped = true
            end
            played[#played + 1] = sound
            return sound
        end,
    })
    local live, cues = {}, {}
    local level = 1
    local map_world = {
        subtile_to_pixel = function()
            return 0, 0
        end,
        entity_depth = function()
            return 0
        end,
        dimensions = function()
            return { width_subtiles = 100, height_subtiles = 100 }
        end,
        canvas = function()
            return 800, 600
        end,
    }
    local function empty()
        return {}
    end
    local scene = {
        gameplay = { player = "local-player" },
        world_level_id = 1,
        world_canvas_width = 800,
        world_canvas_height = 600,
        map = { root = node() },
        world = {
            subtile_to_pixel = function()
                return 0, 0
            end,
            entity_depth = function()
                return 0
            end,
        },
        world_capability = {
            current_level = function()
                return level
            end,
            current = function()
                return map_world, { level_id = level }
            end,
        },
        gameplay_world = {
            bind = function() end,
            set_collision = function() end,
            snap_camera = function() end,
            destroy = function() end,
            player_snapshots = empty,
            monster_snapshots = empty,
            warp_snapshots = empty,
            state_snapshots = empty,
            missile_snapshots = function()
                return live
            end,
            semantic_cues = function()
                return cues
            end,
        },
    }
    local world = require("d2legacy.screens.game_world")
    return {
        played = played,
        level = function(value)
            level = value
        end,
        frame = function(snapshots, events)
            live, cues = snapshots, events or {}
            world.update(scene, 1 / 60, true, false, false)
        end,
        close = function()
            world.destroy(scene)
        end,
    }
end

local function missile(id)
    return {
        entity_id = id,
        sound_owner_id = "shot:" .. id,
        level_id = 1,
        x = 0,
        y = 0,
        offset_x = 0,
        offset_y = 0,
        offset_z = 0,
    }
end

local function cue(id, sound, owner)
    return {
        entity_id = id,
        cue_type = "effect",
        kind = "missile_spawn",
        sound = sound or "travel",
        sound_owner_id = owner,
        tick = id,
    }
end

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("stops_each_travel_loop_when_its_projectile_disappears", {
            test.run(function()
                local f = fixture()
                f.frame({ missile(1), missile(2) }, { cue(101, "travel", "shot:1"), cue(102, "travel", "shot:2") })
                test.expect(#f.played):equals(2)
                f.frame({ missile(1), missile(2) }, { cue(101, "travel", "shot:1"), cue(102, "travel", "shot:2") })
                test.expect(#f.played):equals(2)
                f.frame({ missile(2) })
                test.assert(f.played[1].stopped, "removed projectile's travel loop is still playing")
                test.assert(not f.played[2].stopped, "remaining projectile lost its travel loop")
                f.frame({})
                test.assert(f.played[2].stopped, "expired projectile's travel loop is still playing")
                f.close()
            end),
        }),
        test.case("keeps_one_shots_and_stops_loops_on_scene_exit", {
            test.run(function()
                local f = fixture()
                f.frame({ missile(1) }, { cue(101, "travel", "shot:1"), cue(102, "once", "shot:2") })
                f.close()
                test.assert(f.played[1].stopped, "scene exit retained its travel loop")
                test.assert(not f.played[2].stopped, "one-shot sound was cut short by projectile cleanup")
            end),
        }),
        test.case("stops_old_level_loops_before_replacing_the_world", {
            test.run(function()
                local f = fixture()
                f.frame({ missile(1) }, { cue(101, "travel", "shot:1") })
                f.level(2)
                f.frame({ missile(1) })
                test.assert(f.played[1].stopped, "previous level retained its travel loop")
                f.close()
            end),
        }),
        test.case("does_not_leave_an_unowned_or_already_expired_loop", {
            test.run(function()
                local f = fixture()
                f.frame({}, { cue(101, "travel", "shot:1"), cue(102, "travel"), cue(103, "once") })
                for _, sound in ipairs(f.played) do
                    test.assert(not sound.loop or sound.stopped, "travel loop has no living owner")
                end
                test.assert(f.played[#f.played].path == "once.wav" and not f.played[#f.played].stopped)
                f.close()
            end),
        }),
    },
})
