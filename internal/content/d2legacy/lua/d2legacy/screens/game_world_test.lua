local test = require("d2legacy.tests/v1")

local function fixture()
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
        "d2legacy.gameplay.held_skill_input",
        "d2legacy.gameplay.state_overlay_presentation",
        "d2legacy.gameplay.aura_overlay_cycle",
        "d2legacy.gameplay.interaction_approach",
        "d2legacy.data.skill",
        "d2legacy.presentation.chunked_map",
        "d2legacy.ui.tooltip",
        "d2legacy.ui.text",
    }) do
        test.mock_module(name, {})
    end
    test.mock_module("engine.render/v1", {
        assets_available = function()
            return false
        end,
    })
    test.mock_module("engine.data/v1", {
        load_manifest = function()
            return { screens = { game_world = {} } }
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
