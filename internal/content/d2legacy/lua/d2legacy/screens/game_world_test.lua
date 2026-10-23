local test = require("d2legacy.tests/v1")

local function fixture(frame)
    frame = frame or {}
    for _, name in ipairs({
        "engine.input/v1",
        "engine.vfs/v1",
        "engine.audio/v1",
        "d2legacy.audio",
        "d2legacy.save/v1",
        "d2legacy.ui.game_hud",
        "d2legacy.gameplay.player_composite",
        "d2legacy.gameplay.monster_composite",
        "d2legacy.gameplay.missile_presentation",
        "d2legacy.gameplay.cast_presentation",
        "d2legacy.gameplay.state_overlay_presentation",
        "d2legacy.gameplay.aura_overlay_cycle",
        "d2legacy.data.skill",
        "d2legacy.presentation.chunked_map",
        "d2legacy.ui.tooltip",
        "d2legacy.ui.text",
    }) do
        test.mock_module(name, {})
    end
    test.mock_module("engine.input/v1", {
        cursor = function()
            return 14, 12
        end,
        pressed = function(action)
            return action == "pointer_primary" and frame.pressed == true
        end,
        down = function(action)
            return action == "pointer_primary" and frame.down == true
        end,
        released = function()
            return false
        end,
    })
    test.mock_module("engine.render/v1", {
        assets_available = function()
            return false
        end,
    })
    test.mock_module("engine.data/v1", {
        load_manifest = function()
            return {
                resolution = { width = 800, height = 600 },
                screens = { game_world = { hero = { screen_x = 400, screen_y = 330 } } },
            }
        end,
    })
    test.mock_module("engine.network/v1", {
        status = function()
            return { player_id = "local-player" }
        end,
    })
    local calls = { visible = true }
    test.mock_module("engine.scene/v1", {
        push = function(name)
            calls.pushed = name
        end,
    })
    local scene = require("d2legacy.screens.game_world")
    local instance = {
        gameplay = { player = "local-player" },
        gameplay_world = { bind = function() end },
        world_hover_tip = {
            set_visible = function(_, value)
                calls.visible = value
            end,
            set_text = function(_, value)
                calls.label = value
            end,
            set_position = function() end,
        },
        interaction = {
            dialogue = function()
                return { active = true }
            end,
        },
    }
    return scene, instance, calls
end

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("existing_item_hover_and_approach_emit_one_exact_pickup_intent", function(t)
            t:run(function()
                local frame = {}
                local scene, instance, calls = fixture(frame)
                local submitted = {}
                test.mock_module("engine.command_intent/v1", {
                    submit = function(kind, payload)
                        submitted[#submitted + 1] = { kind = kind, payload = payload }
                    end,
                })
                test.mock_module("d2legacy.items.snapshot", {})
                instance.items = require("d2legacy.items.api")
                instance.interaction = {
                    dialogue = function()
                        return { active = false }
                    end,
                }
                instance.gameplay.hero, instance.gameplay.camera = {}, {}
                instance.gameplay_world = {
                    bind = function()
                        return true
                    end,
                    position = function()
                        return 12, 12
                    end,
                    refresh_camera_settings = function() end,
                    skill_action_active = function()
                        return false
                    end,
                }
                local item = {
                    id = "item:1:" .. string.rep("a", 64),
                    kind = "item",
                    label = "vps",
                    owner = "",
                    radius = 3.5,
                    x = 14,
                    y = 12,
                }
                instance.targeting = {
                    selectable_at = function()
                        return item
                    end,
                }
                instance.world = {
                    screen_to_subtile = function()
                        return 14, 12
                    end,
                    subtile_to_pixel = function(_, x, y)
                        return x, y
                    end,
                    entity_depth = function()
                        return 1
                    end,
                    line_clear = function()
                        return true
                    end,
                }
                instance.hero = { set_z = function() end, set_position = function() end }
                instance.world_canvas_width, instance.world_canvas_height = 800, 600
                instance.player = {
                    request_move = function(x, y, radius)
                        calls.approach = { x, y, radius }
                        frame.pending = true
                    end,
                    movement_pending = function()
                        return frame.pending == true
                    end,
                    request_skill = function()
                        error("item input requested a skill")
                    end,
                }
                scene.update(instance, 0, true, true)
                test.expect(calls.label):equals("vps")
                test.expect(#submitted):equals(0)
                frame.pressed, frame.down = true, true
                scene.update(instance, 0, true, true)
                test.expect(calls.approach):deep_equals({ 14, 12, 3.5 })
                test.expect(#submitted):equals(0)
                frame.pressed, frame.down, frame.pending = false, false, false
                scene.update(instance, 0, true, true)
                test.expect(submitted):deep_equals({ { kind = "item.pickup", payload = { target_id = item.id } } })
                test.expect(instance.pending_interaction):is_nil()
                scene.update(instance, 0, true, true)
                test.expect(#submitted):equals(1)
            end)
        end),
        test.case("dialogue_open_hides_the_previous_world_hover_label", function(t)
            t:run(function()
                local scene, instance, calls = fixture()
                scene.update(instance, 0, true, true)
                test.expect(calls.visible):is_false()
                test.expect(calls.pushed):equals("npc_interaction")
            end)
        end),
        test.case("covered_world_hides_the_previous_world_hover_label", function(t)
            t:run(function()
                local scene, instance, calls = fixture()
                scene.update(instance, 0, false, false)
                test.expect(calls.visible):is_false()
                test.expect(calls.pushed):is_nil()
            end)
        end),
    },
})
