local test = require("d2legacy.tests/v1")

local function load_screen(initial_state)
    local state = initial_state
    local pressed = false
    local calls = { buttons = {}, fields = {}, created = {}, scenes = {}, cancel_count = 0, refresh_count = 0 }

    test.mock_module("engine.input/v1", {
        pressed = function(action)
            return action == "cancel" and pressed
        end,
    }, { "pressed" })
    test.mock_module("engine.scene/v1", {
        replace = function(scene)
            calls.scenes[#calls.scenes + 1] = scene
        end,
    }, { "replace" })
    test.mock_module("engine.realm/v1", {
        status = function()
            return state
        end,
        create_game = function(options)
            calls.created[#calls.created + 1] = options
            state.phase = "creating_game"
            return true
        end,
        refresh = function()
            calls.refresh_count = calls.refresh_count + 1
            return true
        end,
        cancel = function()
            calls.cancel_count = calls.cancel_count + 1
            state.phase = "ready"
        end,
        send_message = function()
            return true
        end,
        logout = function()
            return true
        end,
    }, { "status", "create_game", "refresh", "cancel", "send_message", "logout" })
    test.mock_module("d2legacy.ui.controls", {
        new = function()
            return {
                update = function() end,
                set_focus = function(_, id)
                    calls.focus = id
                end,
            }
        end,
    }, { "new" })
    test.mock_module("d2legacy.ui.text_field", {
        create = function(_, _, id, definition)
            local field = { value = definition.value or "", definition = definition }
            function field:set_value(value)
                self.value = value
            end
            calls.fields[id] = field
            return field
        end,
    }, { "create" })
    test.mock_module("d2legacy.ui.button", {
        create = function(_, _, id, definition, label, options)
            calls.buttons[id] = { definition = definition, label = label, activate = options.on_activate }
        end,
    }, { "create" })
    test.mock_module("d2legacy.ui.checkbox", {
        create = function(_, _, id, definition, _, options)
            local control = { checked = definition.checked == true }
            control.activate = function()
                control.checked = not control.checked
                options.on_change(control, control.checked)
            end
            calls.buttons[id] = control
            return control
        end,
    }, { "create" })
    test.mock_module("d2legacy.ui.text", {
        create = function(_, style, value, x, y, width, align)
            calls.labels = calls.labels or {}
            calls.labels[#calls.labels + 1] = {
                style = style,
                value = value,
                x = x,
                y = y,
                width = width,
                align = align,
            }
            return {}
        end,
    }, { "create" })
    test.mock_module("d2legacy.realm.common", {
        label = function(_, value)
            return { text = value }
        end,
        set_label = function(node, value)
            node.text = value
        end,
        error = function(realm_state)
            return "FRIENDLY: " .. tostring(realm_state.error)
        end,
    }, { "label", "set_label", "error" })
    test.mock_module("d2legacy.realm.lobby_ui", {
        create = function(kind)
            calls.panel = kind
            return {}, {}, {}
        end,
        button = function(kind, x, y)
            return { x = x, y = y, width = kind == "panel_game_button" and 172 or 96, height = 32 }
        end,
        chat_lines = function()
            return ""
        end,
        add_navigation = function() end,
    }, { "create", "button", "chat_lines", "add_navigation" })
    test.mock_module("d2legacy.realm.lobby_roster", {
        create = function()
            return {}
        end,
        update = function() end,
    }, { "create", "update" })
    test.mock_module("d2legacy.realm.number_selector", {
        create = function(_, _, id, definition)
            local selector = { value = definition.value, definition = definition }
            calls.fields[id] = selector
            return selector
        end,
    }, { "create" })

    return require("d2legacy.screens.realm_game_create"), calls, function(value)
        pressed = value
    end
end

return test.suite({
    name = "Realm game creation screen",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("uses_the_authored_lobby_pane_and_selected_character_rules", function(t)
            t:run(function()
                local state = {
                    phase = "lobby",
                    selected = { character = { expansion = false, hardcore = true } },
                }
                local screen, calls = load_screen(state)
                local scene = {}
                screen.create(scene)

                test.expect(calls.panel):equals("create_panel")
                test.expect(calls.focus):equals("name")
                test.expect(calls.fields.name.definition):deep_equals({
                    x = 428,
                    y = 137,
                    width = 180,
                    height = 20,
                    value = "",
                    max_length = 32,
                    background = false,
                })
                test.expect(calls.buttons.create.definition.x):equals(593)
                test.expect(calls.labels[2]):deep_equals({
                    style = "realm_lobby_label",
                    value = "GAME NAME",
                    x = 518,
                    y = 113,
                    width = 180,
                    align = "left",
                })

                scene.name.value = "Trist Run"
                scene.password.value = "secret"
                scene.description.value = "Act I friends"
                scene.maximum_players.value = 6
                calls.buttons.create.activate()
                screen.update(scene, 0)

                test.expect(calls.created):has_length(1)
                test.expect(calls.created[1]):deep_equals({
                    name = "Trist Run",
                    password = "secret",
                    description = "Act I friends",
                    difficulty = "normal",
                    maximum_players = 6,
                    character_difference = 4,
                    expansion = false,
                    hardcore = true,
                })
            end)
        end),
        test.case("can_disable_the_character_difference_rule", function(t)
            t:run(function()
                local state = { phase = "lobby", selected = { character = {} } }
                local screen, calls = load_screen(state)
                local scene = {}
                screen.create(scene)
                scene.name.value = "Open Levels"
                calls.buttons.character_difference_enabled.activate()
                calls.buttons.create.activate()
                screen.update(scene, 0)

                test.expect(calls.created[1].character_difference):equals(0)
            end)
        end),
        test.case("preserves_a_create_click_while_the_refresh_lane_is_busy", function(t)
            t:run(function()
                local state = { phase = "refreshing", selected = { character = {} } }
                local screen, calls = load_screen(state)
                local scene = {}
                screen.create(scene)
                scene.name.value = "Delayed Game"
                calls.buttons.create.activate()

                screen.update(scene, 0)
                test.expect(calls.created):has_length(0)
                state.phase = "lobby"
                screen.update(scene, 0)
                test.expect(calls.created[1].name):equals("Delayed Game")
            end)
        end),
        test.case("shows_mapped_failures_and_cancels_back_to_the_lobby", function(t)
            t:run(function()
                local state = { phase = "error", error = "game service unavailable" }
                local screen, calls, set_pressed = load_screen(state)
                local scene = {}
                screen.create(scene)
                screen.update(scene, 0)
                test.expect(scene.status.text):equals("FRIENDLY: game service unavailable")

                set_pressed(true)
                screen.update(scene, 0)
                test.expect(calls.cancel_count):equals(1)
                test.expect(calls.scenes):deep_equals({ "realm_lobby" })
            end)
        end),
    },
})
