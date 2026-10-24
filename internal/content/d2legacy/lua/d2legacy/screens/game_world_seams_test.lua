local test = require("d2legacy.tests/v1")

-- Keep the real scene, layout and sparse tile adapter; replace only engine IO
-- and unrelated character/HUD content. Each map has a tile on either side of
-- the join so canvas-only rendering cannot satisfy the visibility assertions.
local function fixture()
    local f = {
        level = 1,
        nodes = {},
        hero = {
            x = 940,
            y = 500,
            id = function()
                return 1
            end,
        },
        camera = { x = 920, y = 500 },
    }
    local function node(_, parent)
        local n = { parent = parent, x = 0, y = 0, visible = true, alive = true }
        function n:set_position(x, y)
            self.x, self.y = x, y
        end
        function n:set_z(z)
            self.z = z
        end
        function n:set_visible(v)
            self.visible = v
        end
        function n:exists()
            return self.alive
        end
        function n:destroy()
            self.alive = false
        end
        function n:set_world_tile(map, _, index)
            self.level, self.width, self.height = map.id, 160, 160
            return map.id == 1 and 900 or 0, 440, 160, 160
        end
        function n:set_dc6(path)
            self.item = path
            return 10, 10, -5, -5
        end
        function n:set_tint() end
        function n:set_cof_animation() end
        function n:set_scale() end
        f.nodes[#f.nodes + 1] = n
        return n
    end
    test.mock_module("engine.render/v1", {
        create = node,
        assets_available = function()
            return true
        end,
        preload = function()
            return {}
        end,
        preload_release = function() end,
        preload_status = function()
            return { done = true, failed = 0 }
        end,
        world_tiles = function(map)
            return {
                width = 1000,
                height = 1000,
                chunks = {
                    { index = 0, x = map.id == 1 and 900 or 0, y = 440, width = 160, height = 160, depth = 510 },
                },
            }
        end,
    })
    for _, name in ipairs({
        "engine.scene/v1",
        "engine.vfs/v1",
        "engine.audio/v1",
        "d2legacy.audio",
        "d2legacy.save/v1",
        "d2legacy.ui.game_hud",
        "d2legacy.gameplay.monster_composite",
        "d2legacy.gameplay.missile_presentation",
        "d2legacy.gameplay.cast_presentation",
        "d2legacy.gameplay.state_overlay_presentation",
        "d2legacy.data.skill",
        "d2legacy.ui.tooltip",
        "d2legacy.ui.text",
    }) do
        test.mock_module(name, {})
    end
    test.mock_module("d2legacy.gameplay.state_overlay_presentation", {
        resolve = function()
            return nil
        end,
    })
    local items = {}
    for id = 1, 2 do
        items[id] = {
            id = function()
                return id
            end,
            level = id,
        }
    end
    test.mock_module("engine.ecs/v1", {
        query = function()
            return f.items and items or {}
        end,
        get = function(entity, name)
            local values = name == "d2legacy.world.location" and { level_id = entity.level }
                or name == "d2legacy.item.presentation" and {
                    world_dc6 = "item" .. entity.level,
                    world_animated = false,
                }
                or { x = entity.level == 1 and 950 or 20, y = 500 }
            return {
                get = function(_, key)
                    return values[key]
                end,
            }
        end,
    })
    test.mock_module("engine.input/v1", {
        cursor = function()
            return 200, 330
        end,
        pressed = function(action)
            return f.click and action == "pointer_primary"
        end,
        down = function()
            return false
        end,
        released = function()
            return false
        end,
    })
    test.mock_module("engine.network/v1", {
        status = function()
            return {}
        end,
    })
    test.mock_module("engine.data/v1", {
        load_manifest = function()
            return {
                resolution = { width = 800, height = 600 },
                palettes = { units = "units" },
                screens = { game_world = { map = {}, hero = { screen_x = 400, screen_y = 330, scale = 1 } } },
            }
        end,
    })
    test.mock_module("d2legacy.gameplay.player_composite", {
        resolve = function()
            return { mode = "idle", key = "hero", components = {} }
        end,
        new_playback = function()
            return { mode = "idle", seconds = 0 }
        end,
        advance = function()
            return {}
        end,
        preload_request = function()
            return {}
        end,
    })
    test.mock_module("d2legacy.gameplay.aura_overlay_cycle", {
        select = function()
            return {}, {}
        end,
    })
    local seams = {
        { source_level = 1, destination_level = 2, source_x = 1000, source_y = 500, arrival_x = 0, arrival_y = 500 },
        { source_level = 2, destination_level = 1, source_x = 0, source_y = 500, arrival_x = 1000, arrival_y = 500 },
    }
    seams[#seams + 1] =
        { source_level = 2, destination_level = 3, source_x = 1000, source_y = 500, arrival_x = 0, arrival_y = 500 }
    seams[#seams + 1] =
        { source_level = 3, destination_level = 2, source_x = 0, source_y = 500, arrival_x = 1000, arrival_y = 500 }
    test.mock_module("d2legacy.gameplay.seam_geometry", {
        configured_seams = function()
            return seams
        end,
        matching_seam = function(_, level, x, y)
            for _, seam in ipairs(seams) do
                if
                    seam.source_level == level
                    and math.abs(x - seam.source_x) <= 70
                    and math.abs(y - seam.source_y) <= 70
                then
                    return seam
                end
            end
        end,
        pointer_target = function(_, _, _, x, y)
            return x, y
        end,
    })
    local maps = {}
    for id = 1, 4 do
        maps[id] = {
            id = id,
            selectable_at = function()
                return nil
            end,
            canvas = function()
                return 1000, 1000
            end,
            dimensions = function()
                return { width_subtiles = 1000, height_subtiles = 1000 }
            end,
            subtile_to_pixel = function(_, x, y)
                return x, y
            end,
            entity_depth = function(_, _, y)
                return y
            end,
            screen_to_subtile = function(_, x, y, cx, cy, tx, ty)
                return x - tx + cx, y - ty + cy
            end,
        }
    end
    local function empty()
        return {}
    end
    local gw = {
        bind = function()
            return true
        end,
        set_collision = function() end,
        snap_camera = function()
            f.camera.x, f.camera.y = f.hero.x, f.hero.y
        end,
        position = function(p)
            return p.x, p.y
        end,
        refresh_camera_settings = function() end,
        player_snapshots = empty,
        monster_snapshots = empty,
        npc_snapshots = empty,
        warp_snapshots = empty,
        state_snapshots = empty,
        missile_snapshots = empty,
        semantic_cues = empty,
        composite_snapshot = empty,
        skill_action_active = function()
            return false
        end,
        destroy = function() end,
    }
    local scene = require("d2legacy.screens.game_world")
    local s = {
        root = node(),
        gameplay_world = gw,
        gameplay = { player = "local-player", hero = f.hero, camera = f.camera },
        world_capability = {
            current_level = function()
                return f.level
            end,
            current = function()
                return maps[f.level], { level_id = f.level }
            end,
            presentation = function()
                return {
                    active_level = f.level,
                    maps = {
                        { level_id = 1, map = maps[1] },
                        { level_id = 2, map = maps[2] },
                        { level_id = 3, map = maps[3] },
                        { level_id = 4, map = maps[4] },
                    },
                }
            end,
        },
        player = {
            request_move = function(x, y)
                f.move = { x, y }
            end,
            movement_pending = function()
                return false
            end,
        },
    }
    function f:frame(dt, view)
        scene.update(s, dt or 0, true, true, view)
    end
    function f:cross(level, x, y)
        self.level, self.hero.x, self.hero.y = level, x, y
    end
    function f:close()
        scene.destroy(s)
    end
    f.scene = s
    return f
end

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("warp_between_outdoor_levels_does_not_apply_seam_carry", function(t)
            t:run(function()
                local f = fixture()
                f.hero.x = 400
                f:frame()
                f:frame()
                f:cross(2, 0, 500)
                f:frame(1 / 60)
                test.expect(f.scene.hero.x):equals(1000)
                test.expect(f.scene.hero.parent.x):equals(-600)
                f:close()
            end)
        end),
        test.case("cold_plains_neighbor_tiles_and_bidirectional_visual_continuity", function(t)
            t:run(function()
                local f = fixture()
                f:cross(2, 940, 500)
                f:frame()
                f:frame()
                local neighbor
                for _, n in ipairs(f.nodes) do
                    if n.level == 3 and n.alive then
                        neighbor = n
                    end
                end
                test.expect(neighbor ~= nil and neighbor.visible):is_true()
                local left = neighbor.x + neighbor.parent.x - 80
                test.expect(left >= 0 and left < 800):is_true()
                for _, crossing in ipairs({ { 3, 0, 500 }, { 2, 1000, 500 } }) do
                    local root = f.scene.hero.parent
                    local rx, hx = root.x, root.x + f.scene.hero.x
                    f:cross(crossing[1], crossing[2], crossing[3])
                    f:frame(1 / 60)
                    test.expect(root.x):equals(rx)
                    test.expect(root.x + f.scene.hero.x):equals(hx)
                    test.expect(neighbor.alive):is_true()
                    f:frame(0.04)
                end
                f:close()
            end)
        end),
        test.case("only_active_ground_items_share_global_tile_coordinates", function(t)
            t:run(function()
                local f = fixture()
                f.items = true
                f:frame()
                f:frame()
                local old
                for _, n in ipairs(f.nodes) do
                    if n.item then
                        test.expect(n.item):equals("item1")
                        old = n
                    end
                end
                test.expect(old.x):equals(950)
                test.expect(old.parent == f.scene.hero.parent):is_true()
                f:cross(2, 0, 500)
                f:frame()
                test.expect(old.alive):is_false()
                local item
                for _, n in ipairs(f.nodes) do
                    if n.item and n.alive then
                        item = n
                    end
                end
                test.expect(item.item):equals("item2")
                test.expect(item.x):equals(1020)
                test.expect(item.z):equals(f.scene.hero.z)
                f:close()
                test.expect(item.alive):is_false()
            end)
        end),
        test.case("centered_crossing_and_union_outer_clamp_keep_floor_anchor", function(t)
            t:run(function()
                local f = fixture()
                f.hero.x = 1000
                f:frame()
                f:frame()
                local root = f.scene.hero.parent
                test.expect(root.x):equals(-600)
                f:cross(2, 0, 500)
                f:frame(1 / 60)
                test.expect(root.x):equals(-600)
                test.expect(f.scene.hero.x + root.x):equals(400)
                f:cross(3, 0, 500)
                f:frame()
                f.hero.x, f.camera.x = 990, 990
                f:frame(1)
                test.expect(root.x):equals(-2200)
                f:close()
            end)
        end),
        test.case("town_tile_wins_equal_depth_overlap_independently_of_crossing", function(t)
            t:run(function()
                local f = fixture()
                f:frame()
                f:frame()
                local tiles = {}
                for _, n in ipairs(f.nodes) do
                    if n.level then
                        tiles[n.level] = n
                    end
                end
                test.expect(tiles[1].z > tiles[2].z):is_true()
                f:cross(2, 0, 500)
                f:frame()
                f:frame(1)
                test.expect(tiles[1].z > tiles[2].z):is_true()
                test.expect(f.scene.hero.z < tiles[2].z):is_true()
                f:close()
            end)
        end),
        test.case("disconnected_warp_terrain_is_not_drawn_or_eased", function(t)
            t:run(function()
                local f = fixture()
                f:frame()
                f:frame()
                for _, n in ipairs(f.nodes) do
                    test.expect(n.level == 4 and n.alive):is_false()
                end
                local old_root = f.scene.hero.parent
                f:cross(4, 500, 500)
                f:frame()
                f:frame()
                test.expect(old_root.alive):is_false()
                test.expect(f.scene.hero.x + f.scene.hero.parent.x):equals(400)
                f:close()
            end)
        end),
        test.case("neighbor_tiles_are_visible_before_crossing_and_survive_it", function(t)
            t:run(function()
                local f = fixture()
                f:frame()
                f:frame()
                local neighbor
                for _, n in ipairs(f.nodes) do
                    if n.level == 2 and n.alive then
                        neighbor = n
                    end
                end
                test.expect(neighbor ~= nil):is_true()
                test.expect(neighbor.visible):is_true()
                local left = neighbor.x + neighbor.parent.x - neighbor.width / 2
                test.expect(left >= 0 and left < 800):is_true()
                test.expect(neighbor.parent == f.scene.hero.parent):is_true()
                test.expect(f.scene.hero.z < neighbor.z):is_true()
                f:cross(2, 0, 500)
                f:frame()
                test.expect(neighbor.alive):is_true()
                f:close()
                test.expect(neighbor.alive):is_false()
                test.expect(neighbor.parent.alive):is_false()
            end)
        end),
        test.case("off_center_crossings_and_rapid_reversal_keep_first_frame_positions", function(t)
            t:run(function()
                local f = fixture()
                f:frame()
                f:frame()
                for _, crossing in ipairs({ { 2, 0, 500 }, { 1, 1000, 500 }, { 2, 0, 500 } }) do
                    local root = f.scene.hero.parent
                    local rx, ry = root.x, root.y
                    local hx, hy = f.scene.hero.x + rx, f.scene.hero.y + ry
                    f:cross(crossing[1], crossing[2], crossing[3])
                    f:frame(1 / 60)
                    root = f.scene.hero.parent
                    test.expect(root.x):equals(rx)
                    test.expect(root.y):equals(ry)
                    test.expect(f.scene.hero.x + root.x):equals(hx)
                    test.expect(f.scene.hero.y + root.y):equals(hy)
                    f:frame(0.04)
                end
                f:frame(1)
                test.expect(f.scene.hero.x):equals(1000)
                f:close()
            end)
        end),
        test.case("pointer_uses_visible_floor_during_side_panel_settle", function(t)
            t:run(function()
                local f = fixture()
                f:frame(0, "left")
                f:frame(0, "left")
                f:cross(2, 0, 500)
                f.click = true
                f:frame(1 / 60, "left")
                -- Before crossing camera was at global 940. Panel anchor is 200.
                test.expect(f.move):deep_equals({ -60, 500 })
                f:close()
            end)
        end),
    },
})
