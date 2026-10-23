local test = require("d2legacy.tests/v1")
return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("belt_right_click_uses_exact_item_and_hides_retired_picture", {
            test.run(function()
                local x, y, clicked = 21, 41, true
                test.mock_module("engine.input/v1", {
                    cursor = function()
                        return x, y
                    end,
                    pressed = function(action)
                        return clicked and action == "pointer_secondary"
                    end,
                })
                for _, name in ipairs({
                    "engine.render/v1",
                    "engine.scene/v1",
                    "engine.locale/v1",
                    "d2legacy.ui.controls",
                    "d2legacy.ui.tooltip",
                    "d2legacy.ui.skill_selector",
                }) do
                    test.mock_module(name, {})
                end
                local picture = {
                    visible = true,
                    set_visible = function(self, value)
                        self.visible = value
                    end,
                }
                local snapshot = {
                    items = { { id = "belt-hp1", code = "hp1", container = "belt", belt_slot = 0, inventory_dc6 = "" } },
                }
                local used
                local hud = {
                    commands = {
                        item_snapshot = function()
                            return snapshot
                        end,
                        use_item = function(id)
                            used = id
                        end,
                    },
                    definition = { belt = { x = 20, y = 40, columns = 4, cell_width = 10, cell_height = 10 } },
                    item_nodes = { ["belt-hp1"] = { belt_node = picture } },
                    controls = { update = function() end },
                }
                local module = require("d2legacy.ui.game_hud")
                -- Globe/skill drawing is unrelated to this item-input case.
                module.snapshot = function() end
                module.update(hud, {})
                test.expect(used):equals("belt-hp1")
                test.assert(picture.visible)
                used, clicked, snapshot.items = nil, false, {}
                module.update(hud, {})
                test.assert(not picture.visible)
                test.expect(used):is_nil()
                snapshot.items = {
                    { id = "unsupported", code = "mp1", container = "belt", belt_slot = 0, inventory_dc6 = "" },
                }
                clicked = true
                module.update(hud, {})
                test.expect(used):is_nil()
                snapshot.items[1].code = "hp1"
                snapshot.items[2] = { id = "held", container = "held", inventory_dc6 = "" }
                module.update(hud, {})
                test.expect(used):is_nil()
                snapshot.items[2], x = nil, 19
                module.update(hud, {})
                test.expect(used):is_nil()
            end),
        }),
    },
})
