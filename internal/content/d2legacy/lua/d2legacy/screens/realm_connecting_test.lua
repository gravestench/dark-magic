local test = require("d2legacy.tests/v1")

local function load_screen()
    local status = { phase = "checking_versions" }
    local pressed = false
    local calls = {
        buttons = {},
        connects = {},
        labels = {},
        scenes = {},
        cancel_count = 0,
    }

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
        connect = function(endpoint)
            calls.connects[#calls.connects + 1] = endpoint
            return true
        end,
        status = function()
            return status
        end,
        cancel = function()
            calls.cancel_count = calls.cancel_count + 1
        end,
    }, { "connect", "status", "cancel" })
    test.mock_module("d2legacy.ui.controls", {
        new = function()
            return { update = function() end }
        end,
    }, { "new" })
    test.mock_module("d2legacy.ui.button", {
        create = function(_, _, id, definition, label, options)
            calls.buttons[id] = {
                definition = definition,
                label = label,
                options = options,
            }
        end,
    }, { "create" })
    test.mock_module("d2legacy.realm.common", {
        frontend_root = function()
            return {}
        end,
        popup = function() end,
        button_definition = function(kind, x, y)
            test.expect(kind):equals("cancel_button")
            return {
                x = x,
                y = y,
                sheet = "data/global/ui/FrontEnd/CancelButtonBlank.dc6",
                palette = "units",
                width = 96,
                height = 32,
                up_frame = 0,
                down_frame = 1,
            }
        end,
        label = function(_, value)
            local node = {
                text = value,
                visible = true,
                set_visible = function(self, visible)
                    self.visible = visible
                end,
            }
            calls.labels[#calls.labels + 1] = node
            return node
        end,
        set_label = function(node, value)
            node.text = value
        end,
        error = function(realm_status)
            return realm_status.error or "UNABLE TO CONTACT REALM"
        end,
    }, { "frontend_root", "popup", "button_definition", "label", "set_label", "error" })

    return require("d2legacy.screens.realm_connecting"), calls, status, function(value)
        pressed = value
    end
end

return test.suite({
    name = "Realm connection screen",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("uses_the_authored_compact_cancel_button", function(t)
            t:run(function()
                local screen, calls = load_screen()
                local scene = {}
                screen.create(scene)

                test.expect(calls.connects):deep_equals({ "" })
                test.expect(calls.buttons.cancel.label):equals("CANCEL")
                test.expect(calls.buttons.cancel.definition):deep_equals({
                    x = 352,
                    y = 345,
                    width = 96,
                    height = 32,
                    sheet = "data/global/ui/FrontEnd/CancelButtonBlank.dc6",
                    palette = "units",
                    up_frame = 0,
                    down_frame = 1,
                })

                calls.buttons.cancel.options.on_activate()
                test.expect(calls.cancel_count):equals(1)
                test.expect(calls.scenes):deep_equals({ "main_menu" })
            end)
        end),
        test.case("reports_connection_errors_and_opens_login_only_when_ready", function(t)
            t:run(function()
                local screen, calls, status, set_pressed = load_screen()
                local scene = {}
                screen.create(scene)

                status.phase = "error"
                status.error = "REALM CONNECTION TIMED OUT"
                screen.update(scene, 0.5)
                test.expect(scene.progress.visible):equals(false)
                test.expect(scene.dots.visible):equals(false)
                test.expect(scene.error.text):equals("REALM CONNECTION TIMED OUT")
                test.expect(calls.scenes):has_length(0)

                status.phase = "login"
                screen.update(scene, 0)
                test.expect(calls.scenes):deep_equals({ "realm_login" })

                set_pressed(true)
                screen.update(scene, 0)
                test.expect(calls.cancel_count):equals(1)
                test.expect(calls.scenes):deep_equals({ "realm_login", "main_menu" })
            end)
        end),
    },
})
