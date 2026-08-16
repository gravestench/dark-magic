local test = require("d2legacy.tests/v1")

local function load_screen(gateway_accepted)
    local calls = {
        buttons = {},
        gateways = {},
        labels = {},
        scenes = {},
    }

    test.mock_module("engine.input/v1", {
        pressed = function()
            return false
        end,
    }, { "pressed" })
    test.mock_module("engine.scene/v1", {
        replace = function(scene)
            calls.scenes[#calls.scenes + 1] = scene
        end,
    }, { "replace" })
    test.mock_module("engine.realm/v1", {
        status = function()
            return { gateway = "realm.example.test" }
        end,
        set_gateway = function(gateway)
            calls.gateways[#calls.gateways + 1] = gateway
            return gateway_accepted
        end,
    }, { "status", "set_gateway" })
    test.mock_module("d2legacy.ui.controls", {
        new = function()
            return { update = function() end }
        end,
    }, { "new" })
    test.mock_module("d2legacy.ui.text_field", {
        create = function(_, _, _, definition)
            calls.field = definition
            return { value = definition.value }
        end,
    }, { "create" })
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
        popup = function(_, x, y, kind)
            calls.popup = { x = x, y = y, kind = kind }
        end,
        label = function(_, value)
            local node = { text = value }
            calls.labels[#calls.labels + 1] = node
            return node
        end,
        set_label = function(node, value)
            node.text = value
        end,
        button_definition = function(kind, x, y)
            return {
                x = x,
                y = y,
                width = 96,
                height = 32,
                sheet = "data/global/ui/FrontEnd/CancelButtonBlank.dc6",
                palette = "units",
                up_frame = 0,
                down_frame = 1,
            }
        end,
    }, { "frontend_root", "popup", "label", "set_label", "button_definition" })

    return require("d2legacy.screens.realm_gateway"), calls
end

return test.suite({
    name = "Realm gateway screen",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("uses_the_units_popup_and_authored_compact_actions", function(t)
            t:run(function()
                local screen, calls = load_screen(true)
                local scene = {}
                screen.create(scene)

                test.expect(calls.popup):deep_equals({ x = 230, y = 170 })
                test.expect(calls.field.value):equals("realm.example.test")
                test.expect(calls.field.palette):equals("units")
                test.expect(calls.buttons.ok.definition.x):equals(288)
                test.expect(calls.buttons.ok.definition.y):equals(345)
                test.expect(calls.buttons.ok.definition.palette):equals("units")
                test.expect(calls.buttons.ok.definition.sheet):equals("data/global/ui/FrontEnd/CancelButtonBlank.dc6")
                test.expect(calls.buttons.cancel.definition.x):equals(416)
                test.expect(calls.buttons.cancel.definition.sheet)
                    :equals("data/global/ui/FrontEnd/CancelButtonBlank.dc6")

                scene.gateway.value = "new-realm.example.test"
                calls.buttons.ok.options.on_activate()
                test.expect(calls.gateways):deep_equals({ "new-realm.example.test" })
                test.expect(calls.scenes):deep_equals({ "main_menu" })
            end)
        end),
        test.case("keeps_an_invalid_gateway_visible_for_correction", function(t)
            t:run(function()
                local screen, calls = load_screen(false)
                local scene = {}
                screen.create(scene)

                calls.buttons.ok.options.on_activate()
                test.expect(calls.scenes):has_length(0)
                test.expect(scene.status.text):equals("INVALID GATEWAY")

                calls.buttons.cancel.options.on_activate()
                test.expect(calls.scenes):deep_equals({ "main_menu" })
            end)
        end),
    },
})
