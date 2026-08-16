local test = require("d2legacy.tests/v1")

local function load_screen(name, initial_state)
    local state = initial_state or { phase = "login" }
    local calls = {
        scenes = {},
        login = {},
        signup = {},
        recovery = {},
        buttons = {},
        button_y = {},
        labels = {},
        fields = {},
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
            return state
        end,
        login = function(account_name, password)
            calls.login[#calls.login + 1] = { account_name, password }
            return true
        end,
        signup = function(account_name, email, password)
            calls.signup[#calls.signup + 1] = { account_name, email, password }
            return true
        end,
        recover_password = function(email)
            calls.recovery[#calls.recovery + 1] = email
            return true
        end,
        cancel = function() end,
    }, { "status", "login", "signup", "recover_password", "cancel" })
    test.mock_module("d2legacy.ui.controls", {
        new = function()
            return {
                update = function() end,
            }
        end,
    }, { "new" })
    test.mock_module("d2legacy.ui.text_field", {
        create = function(_, _, id, options, label, presentation)
            calls.fields[id] = {
                definition = options,
                label = label,
                presentation = presentation,
            }
            return {
                id = id,
                value = options.value or "",
                cursor = 0,
            }
        end,
    }, { "create" })
    test.mock_module("d2legacy.realm.common", {
        label = function(_, text)
            local label = { text = text }
            calls.labels[#calls.labels + 1] = label
            return label
        end,
        set_label = function(label, text)
            label.text = text
        end,
        error = function(realm_state)
            return realm_state.error or "REALM REQUEST FAILED"
        end,
    }, { "label", "set_label", "error" })
    test.mock_module("d2legacy.realm.account_ui", {
        create_root = function()
            return {}, {}
        end,
        update_logo = function() end,
        add_button = function(_, _, id, y, _, on_activate)
            calls.buttons[id] = on_activate
            calls.button_y[id] = y
        end,
        clear_secret = function(field)
            field.value = ""
            field.cursor = 0
        end,
        text_field = function(definition)
            local result = {
                width = 272,
                height = 32,
                palette = "units",
                combined = true,
            }
            for key, value in pairs(definition or {}) do
                result[key] = value
            end
            return result
        end,
    }, { "create_root", "update_logo", "add_button", "clear_secret", "text_field" })

    return require("d2legacy.screens." .. name), calls, state
end

return test.suite({
    name = "Realm account screen flow",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("login_requires_an_explicit_credential_submission", function(t)
            t:run(function()
                local screen, calls, state = load_screen("realm_login")
                local scene = {}
                screen.create(scene)

                test.expect(calls.fields.account_name.definition.palette):equals("units")
                test.expect(calls.fields.account_name.definition.combined):equals(true)
                test.expect(calls.fields.account_name.definition.y):equals(300)
                test.expect(calls.fields.password.definition.palette):equals("units")
                test.expect(calls.fields.password.definition.combined):equals(true)
                test.expect(calls.fields.password.definition.y):equals(365)

                scene.name.value = "wanderer"
                scene.password.value = "secret phrase"
                calls.buttons.login()

                test.expect(calls.login):deep_equals({ { "wanderer", "secret phrase" } })
                test.expect(scene.password.value):equals("")
                test.expect(calls.scenes):has_length(0)

                state.phase = "characters"
                state.characters = {}
                screen.update(scene, 0)
                test.expect(calls.scenes):deep_equals({ "realm_create" })
            end)
        end),
        test.case("successful_login_opens_an_existing_character_roster", function(t)
            t:run(function()
                local screen, calls, state = load_screen("realm_login")
                local scene = {}
                screen.create(scene)

                state.phase = "characters"
                state.characters = { { character = { id = "realm-hero" } } }
                screen.update(scene, 0)

                test.expect(calls.scenes):deep_equals({ "realm_characters" })
            end)
        end),
        test.case("signup_verifies_email_without_logging_the_client_in", function(t)
            t:run(function()
                local screen, calls, state = load_screen("realm_signup")
                local scene = {}
                screen.create(scene)

                test.expect(calls.fields.account_name.definition.y):equals(275)
                test.expect(calls.fields.email.definition.y):equals(330)
                test.expect(calls.fields.password.definition.y):equals(385)
                test.expect(calls.fields.confirm_password.definition.y):equals(440)
                test.expect(calls.fields.confirm_password.definition.palette):equals("units")
                test.expect(calls.fields.confirm_password.definition.combined):equals(true)
                test.expect(calls.button_y.create):equals(490)
                test.expect(calls.button_y.back):equals(530)

                scene.name.value = "wanderer"
                scene.email.value = "wanderer@example.test"
                scene.password.value = "secret phrase"
                scene.confirm.value = "secret phrase"
                calls.buttons.create()

                test.expect(calls.signup):deep_equals({
                    { "wanderer", "wanderer@example.test", "secret phrase" },
                })
                test.expect(calls.login):has_length(0)
                test.expect(scene.password.value):equals("")
                test.expect(scene.confirm.value):equals("")

                state.phase = "verification_required"
                screen.update(scene, 0)
                test.expect(calls.scenes):has_length(0)
            end)
        end),
        test.case("password_recovery_starts_in_client_and_stays_out_of_login", function(t)
            t:run(function()
                local screen, calls, state = load_screen("realm_recovery")
                local scene = {}
                screen.create(scene)

                test.expect(calls.fields.email.definition.y):equals(310)
                test.expect(calls.fields.email.definition.palette):equals("units")
                test.expect(calls.fields.email.definition.combined):equals(true)
                test.expect(calls.button_y.send):equals(400)
                test.expect(calls.button_y.back):equals(440)

                scene.email.value = "wanderer@example.test"
                calls.buttons.send()

                test.expect(calls.recovery):deep_equals({ "wanderer@example.test" })
                test.expect(calls.login):has_length(0)

                state.phase = "recovery_sent"
                screen.update(scene, 0)
                test.expect(calls.scenes):has_length(0)
            end)
        end),
    },
})
