-- Exercise the shipped overlay and decoded item art after modal-node reuse.
local render = require("engine.render/v1")
local test = require("d2legacy.tests/v1")
local initial = require("engine.initial_data/v1")

return {
    id = "test.inventory.render",
    start = function(self)
        if initial.get("test.reuse_nodes") then
            local prior = render.create("modal")
            for _ = 1, 64 do
                render.create("modal", prior)
            end
            prior:destroy()
        end
        test.mock_module("engine.input/v1", {
            cursor = function()
                return 700, 500
            end,
            pressed = function()
                return false
            end,
            down = function()
                return false
            end,
        })
        test.mock_module("engine.audio/v1", {})
        test.mock_module("engine.scene/v1", {})
        test.mock_module("engine.locale/v1", {
            text = function(value)
                return value
            end,
        })
        test.mock_module("d2legacy.save/v1", {
            selected = function()
                return { class = "Amazon" }
            end,
        })
        test.mock_module("d2legacy.items.api", {
            snapshot = function()
                return {
                    active_weapon_set = 0,
                    items = {
                        {
                            id = "owned-potion",
                            code = "vps",
                            container = "inventory",
                            x = 0,
                            y = 0,
                            width = 1,
                            height = 1,
                            inventory_dc6 = "data/global/items/invwps.dc6",
                        },
                    },
                }
            end,
        })
        self.inventory = {}
        require("d2legacy.overlays.inventory").enter(self.inventory)
    end,
}
