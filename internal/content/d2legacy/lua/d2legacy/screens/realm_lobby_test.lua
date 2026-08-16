local test = require("d2legacy.tests/v1")

local function load_screen(initial_state, assets_available)
    local state = initial_state
    local pressed = false
    local calls = {
        buttons = {},
        messages = {},
        roster_updates = {},
        scenes = {},
        logout_count = 0,
        refresh_count = 0,
        selected_games = {},
        joined_games = {},
        nodes = {},
    }

    test.mock_module("engine.render/v1", {
        assets_available = function()
            return assets_available == true
        end,
        create = function(_, parent)
            local node = { parent = parent }
            function node:fill_rect(...)
                self.fill = { ... }
            end
            function node:set_position(...)
                self.position = { ... }
            end
            function node:set_z(value)
                self.z = value
            end
            function node:set_dc6_combined(path)
                self.sheet = path
                if path:find("waitingroom", 1, true) then
                    return 800, 600
                end
                return 373, 373
            end
            calls.nodes[#calls.nodes + 1] = node
            return node
        end,
    }, { "assets_available", "create" })
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
        send_message = function(message)
            calls.messages[#calls.messages + 1] = message
            return true
        end,
        select_game = function(reference)
            calls.selected_games[#calls.selected_games + 1] = reference
            return true
        end,
        join_game = function(reference, password)
            calls.joined_games[#calls.joined_games + 1] = { reference, password }
            return true
        end,
        refresh = function()
            calls.refresh_count = calls.refresh_count + 1
            return true
        end,
        logout = function()
            calls.logout_count = calls.logout_count + 1
            state.phase = "logging_out"
            return true
        end,
    }, { "status", "send_message", "select_game", "join_game", "refresh", "logout" })
    test.mock_module("d2legacy.ui.controls", {
        new = function()
            return { update = function() end }
        end,
    }, { "new" })
    test.mock_module("d2legacy.ui.list", {
        create = function(_, _, _, _, _, options)
            calls.game_list_options = options
            return {
                set_items = function(_, items)
                    calls.games = items
                end,
            }
        end,
    }, { "create" })
    test.mock_module("d2legacy.ui.text_field", {
        create = function(_, _, id, definition)
            local field = { value = definition.value or "" }
            function field:set_value(value)
                self.value = value
            end
            calls.fields = calls.fields or {}
            calls.fields[id] = field
            return field
        end,
    }, { "create" })
    test.mock_module("d2legacy.ui.button", {
        create = function(_, _, id, definition, _, options)
            calls.buttons[id] = { activate = options.on_activate, definition = definition, enabled = options.enabled }
        end,
    }, { "create" })
    test.mock_module("d2legacy.realm.common", {
        manifest = {
            palettes = { act1 = "act1.dat" },
            screens = {
                realm_lobby = {
                    background = { sheet = "waitingroombkgd.dc6", palette = "act1" },
                    join_panel = { sheet = "joingamebckg.dc6", palette = "act1" },
                },
            },
        },
        label = function(_, value)
            return { text = value }
        end,
        set_label = function(node, value)
            node.text = value
        end,
        error = function(state)
            return state.error or "REALM REQUEST FAILED"
        end,
        screen_definition = function(_, kind, x, y)
            local widths = {
                top_button = 119,
                bottom_button = 79,
                chat_button = 80,
                panel_cancel_button = 96,
                panel_game_button = 172,
            }
            return { x = x, y = y, width = widths[kind], height = kind:find("panel", 1, true) and 32 or 20 }
        end,
    }, { "label", "set_label", "screen_definition", "error" })
    test.mock_module("d2legacy.realm.lobby_roster", {
        create = function(root)
            calls.roster_root = root
            return {}
        end,
        update = function(_, members)
            calls.roster_updates[#calls.roster_updates + 1] = members
        end,
    }, { "create", "update" })

    return require("d2legacy.screens.realm_lobby"), calls, function(value)
        pressed = value
    end
end

return test.suite({
    name = "Realm lobby flow",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("projects_only_live_channel_members_into_the_visible_roster", function(t)
            t:run(function()
                local live = { member_id = "live-session", character = { name = "OnlineHero" } }
                local state = {
                    phase = "lobby",
                    channel = { members = { live } },
                    characters = { { character = { name = "OfflineAccountCharacter" } } },
                    selected = { character = { name = "SelectedButNotPresent" } },
                    games = {},
                    events = {},
                }
                local screen, calls = load_screen(state)
                local scene = {}
                screen.create(scene)
                screen.update(scene, 0)

                test.expect(calls.roster_updates):has_length(1)
                test.expect(calls.roster_updates[1]):deep_equals({ live })
                test.expect(calls.buttons.nav_create.definition.x):equals(532)
                test.expect(calls.buttons.nav_join.definition.x):equals(653)
                test.expect(calls.buttons.nav_channel.enabled):equals(false)
                test.expect(calls.buttons.nav_ladder.enabled):equals(false)
                test.expect(calls.buttons.nav_quit.definition.x):equals(697)
            end)
        end),
        test.case("keeps_lobby_children_on_a_stationary_scene_root", function(t)
            t:run(function()
                local state = { phase = "lobby", channel = { members = {} }, games = {}, events = {} }
                local screen, calls = load_screen(state, true)
                local scene = {}
                screen.create(scene)

                test.expect(calls.nodes[1].position):equals(nil)
                test.expect(calls.nodes[2].parent):equals(calls.nodes[1])
                test.expect(calls.nodes[2].position):deep_equals({ 400, 300 })
                test.expect(calls.nodes[3].parent):equals(calls.nodes[1])
                test.expect(calls.roster_root):equals(calls.nodes[1])
            end)
        end),
        test.case("logs_out_before_leaving_the_realm_lobby", function(t)
            t:run(function()
                local state = { phase = "lobby", channel = { members = {} }, games = {}, events = {} }
                local screen, calls, set_pressed = load_screen(state)
                local scene = {}
                screen.create(scene)

                set_pressed(true)
                screen.update(scene, 0)
                test.expect(calls.logout_count):equals(1)
                test.expect(calls.scenes):has_length(0)
                test.expect(scene.status.text):equals("LOGGING OUT")

                set_pressed(false)
                state.phase = "login"
                screen.update(scene, 0)
                test.expect(calls.scenes):deep_equals({ "main_menu" })
            end)
        end),
        test.case("supports_selected_and_manual_named_game_join", function(t)
            t:run(function()
                local game = { game_id = "opaque-id", name = "Trist Run", players = 1, maximum_players = 8 }
                local state = { phase = "lobby", channel = { members = {} }, games = { game }, events = {} }
                local screen, calls = load_screen(state)
                local scene = {}
                screen.create(scene)

                calls.game_list_options.on_select(game)
                test.expect(scene.game_name.value):equals("Trist Run")
                screen.update(scene, 0)
                test.expect(calls.selected_games):deep_equals({ "opaque-id" })

                scene.game_password.value = "secret"
                calls.buttons.join_game.activate()
                screen.update(scene, 0)
                test.expect(calls.joined_games):deep_equals({ { "Trist Run", "secret" } })

                scene.game_name.value = "Hidden Friends Game"
                calls.buttons.join_game.activate()
                screen.update(scene, 0)
                test.expect(calls.joined_games[2]):deep_equals({ "Hidden Friends Game", "secret" })
            end)
        end),
        test.case("recovers_live_lobby_state_after_a_failed_game_pane", function(t)
            t:run(function()
                local state = {
                    phase = "error",
                    error = "game service unavailable",
                    channel = { members = {} },
                    games = {},
                    events = {},
                }
                local screen, calls = load_screen(state)
                local scene = {}
                screen.create(scene)
                screen.update(scene, 0)

                test.expect(calls.refresh_count):equals(1)
                screen.update(scene, 0)
                test.expect(calls.refresh_count):equals(1)
            end)
        end),
    },
})
