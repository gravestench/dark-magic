local test = require("d2legacy.tests/v1")

local function dependencies(options)
    options = options or {}
    local calls = { scenes = {}, selected = {}, created = {}, joined = {} }
    local realm_state = options.realm_state or { phase = "characters", characters = {} }

    test.mock_module("engine.scene/v1", {
        replace = function(name)
            calls.scenes[#calls.scenes + 1] = name
        end,
    }, { "replace" })
    test.mock_module("d2legacy.save/v1", {
        characters = function()
            return options.local_characters or {}
        end,
        select = function(id)
            calls.selected[#calls.selected + 1] = id
            return true
        end,
        create_named = function(name, class, expansion, hardcore)
            calls.created[#calls.created + 1] = { name, class, expansion, hardcore }
            return "local-created"
        end,
        delete = function()
            return true
        end,
    }, { "characters", "select", "create_named", "delete" })
    test.mock_module("d2legacy.network.flow", {
        start_selected = function()
            return true
        end,
        cancel_destination = function(fallback)
            return fallback
        end,
    }, { "start_selected", "cancel_destination" })
    test.mock_module("engine.realm/v1", {
        status = function()
            return realm_state
        end,
        select_character = function(id)
            calls.selected[#calls.selected + 1] = id
            return true
        end,
        create_character = function(name, class, expansion, hardcore)
            calls.created[#calls.created + 1] = { name, class, expansion, hardcore }
            return true
        end,
        delete_character = function(id)
            calls.deleted = id
            return true
        end,
        join_channel = function(name)
            calls.joined[#calls.joined + 1] = name
            return true
        end,
        cancel = function() end,
    }, { "status", "select_character", "create_character", "delete_character", "join_channel", "cancel" })
    test.mock_module("d2legacy.realm.common", {
        error = function(state)
            return state.error or "REALM REQUEST FAILED"
        end,
    }, { "error" })

    return require("d2legacy.characters.roster"), calls, realm_state
end

return test.suite({
    name = "shared local and Realm character roster adapters",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("local_selection_uses_profile_then_enters_game", function(t)
            t:run(function()
                local rosters, calls = dependencies({
                    local_characters = { { id = "local-hero", name = "Hero", class = "Amazon" } },
                })
                local authority = rosters.for_mode("local")

                test.expect(authority.characters()):has_length(1)
                authority.activate({}, "local-hero")

                test.expect(calls.selected):deep_equals({ "local-hero" })
                test.expect(calls.scenes):deep_equals({ "game_loading" })
            end)
        end),
        test.case("realm_roster_projects_the_same_character_shape", function(t)
            t:run(function()
                local rosters = dependencies({
                    realm_state = {
                        phase = "characters",
                        characters = {
                            { revision = 4, character = { id = "realm-hero", name = "Hero", class = "Druid" } },
                        },
                    },
                })
                local authority = rosters.for_mode("realm")

                test.expect(authority.characters()):deep_equals({
                    { id = "realm-hero", name = "Hero", class = "Druid" },
                })
                test.expect(authority.create_scene):equals("realm_create")
                test.expect(authority.select_scene):equals("realm_characters")
            end)
        end),
        test.case("realm_selection_joins_chat_after_server_confirmation", function(t)
            t:run(function()
                local state = {
                    phase = "characters",
                    characters = {
                        { character = { id = "realm-hero", name = "Hero", class = "Barbarian" } },
                    },
                }
                local rosters, calls = dependencies({ realm_state = state })
                local authority = rosters.for_mode("realm")
                local scene = {}

                authority.activate(scene, "realm-hero")
                state.phase = "character_selected"
                state.selected = { character = state.characters[1].character }
                authority.update(scene)
                authority.update(scene)

                test.expect(calls.selected):deep_equals({ "realm-hero" })
                test.expect(calls.joined):deep_equals({ "Diablo II" })

                state.phase = "lobby"
                authority.update(scene)
                test.expect(calls.scenes):deep_equals({ "realm_lobby" })
            end)
        end),
        test.case("realm_creation_submits_the_shared_form_values", function(t)
            t:run(function()
                local rosters, calls = dependencies()
                local authority = rosters.for_mode("realm")
                local accepted = authority.create({}, "NewHero", "Assassin", true, false)

                test.expect(accepted):is_true()
                test.expect(calls.created):deep_equals({ { "NewHero", "Assassin", true, false } })
            end)
        end),
        test.case("realm_deletion_refreshes_the_shared_roster", function(t)
            t:run(function()
                local state = { phase = "characters", characters = {} }
                local rosters, calls = dependencies({ realm_state = state })
                local authority = rosters.for_mode("realm")
                local scene = {}

                local accepted, err, immediate = authority.delete(scene, "realm-hero")
                test.expect(accepted):is_true()
                test.expect(err):is_nil()
                test.expect(immediate):is_false()
                test.expect(calls.deleted):equals("realm-hero")
                test.expect(authority.update(scene)):is_true()
                test.expect(authority.update(scene)):is_false()
            end)
        end),
    },
})
