local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("consumed_item_picture_disappears_while_inventory_stays_open", {
            test.run(function()
                local clicked = true
                test.mock_module("engine.render/v1", {}, {})
                test.mock_module("engine.input/v1", {
                    cursor = function()
                        return 0.5, 0.5
                    end,
                    pressed = function(action)
                        return clicked and action == "pointer_secondary"
                    end,
                }, { "cursor", "pressed" })
                test.mock_module("engine.scene/v1", {}, {})
                test.mock_module("d2legacy.save/v1", {}, {})
                test.mock_module("engine.locale/v1", {}, {})
                test.mock_module("d2legacy.ui.controls", {}, {})
                test.mock_module("d2legacy.ui.button", {}, {})
                test.mock_module("engine.data/v1", {
                    load_manifest = function()
                        return { screens = { inventory = {} }, palettes = {} }
                    end,
                }, { "load_manifest" })
                local function picture()
                    return {
                        visible = true,
                        set_visible = function(self, visible)
                            self.visible = visible
                        end,
                        set_position = function() end,
                    }
                end
                local scroll, potion = picture(), picture()
                local bought = {
                    id = "bought",
                    code = "tsc",
                    container = "inventory",
                    inventory_dc6 = "",
                    x = 0,
                    y = 0,
                    width = 1,
                    height = 1,
                }
                local kept = {
                    id = "potion",
                    code = "hp1",
                    container = "inventory",
                    inventory_dc6 = "",
                    x = 1,
                    y = 0,
                    width = 1,
                    height = 1,
                }
                local snapshot = { items = { bought, kept }, active_weapon_set = 0 }
                local used
                local instance = {
                    items = {
                        snapshot = function()
                            return snapshot
                        end,
                        use = function(id)
                            used = id
                            snapshot.items = { kept }
                        end,
                    },
                    controls = { update = function() end },
                    item_nodes = {
                        bought = { node = scroll, width = 1, height = 1 },
                        potion = { node = potion, width = 1, height = 1 },
                    },
                    grid_left = 0,
                    grid_top = 0,
                    cell_width = 1,
                    cell_height = 1,
                    equipment_slots = {},
                }
                local overlay = require("d2legacy.overlays.inventory")
                overlay.update(instance)
                test.expect(used):equals("bought")
                test.assert(scroll.visible and potion.visible)
                clicked = false
                overlay.update(instance)
                test.assert(not scroll.visible, "consumed scroll picture remains visible")
                test.assert(potion.visible, "retained potion picture disappeared")
            end),
        }),
    },
})
