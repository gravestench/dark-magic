local test = require("d2legacy.tests/v1")

return test.suite({
    name = "Realm numeric selector",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("clamps_arrow_changes_to_the_authored_range", function(t)
            t:run(function()
                local activations = {}
                test.mock_module("engine.render/v1", {
                    assets_available = function()
                        return false
                    end,
                }, { "assets_available" })
                test.mock_module("d2legacy.ui.text", { set = function() end }, { "set" })
                test.mock_module("d2legacy.ui.button", {
                    create = function(_, _, id, definition, _, options)
                        activations[id] = options.on_activate
                        return { definition = definition }
                    end,
                }, { "create" })

                local selector = require("d2legacy.realm.number_selector").create({}, {}, "players", {
                    x = 10,
                    y = 20,
                    width = 30,
                    height = 32,
                    arrow_x = 45,
                    value = 2,
                    minimum = 1,
                    maximum = 3,
                })
                activations.players_up()
                activations.players_up()
                test.expect(selector.value):equals(3)
                activations.players_down()
                activations.players_down()
                activations.players_down()
                test.expect(selector.value):equals(1)
                test.expect(selector.up.definition.up_frame):equals(0)
                test.expect(selector.down.definition.up_frame):equals(2)
            end)
        end),
    },
})
