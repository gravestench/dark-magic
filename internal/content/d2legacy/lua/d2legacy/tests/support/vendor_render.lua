-- Real retained rendering with controlled item facts. Released scene nodes make
-- allocator reuse independent of camera residency or the size of a gossip page.
local render = require("engine.render/v1")
local test = require("d2legacy.tests/v1")
local initial = require("engine.initial_data/v1")

return {
    id = "test.vendor.render",
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
                return 136, 141
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
                local labels = {
                    ["d2legacy.vendor.armor"] = "ARMOR",
                    ["d2legacy.vendor.weapons"] = "WEAPONS",
                    ["d2legacy.vendor.misc"] = "MISC",
                    ["d2legacy.vendor.carried_gold"] = "Carried gold",
                    ["d2legacy.vendor.price"] = "Price:",
                }
                return labels[value] or value
            end,
        })
        test.mock_module("d2legacy.items.api", {
            snapshot = function()
                return {
                    carried_gold = 1000,
                    items = {
                        {
                            id = "offer-hp1",
                            container = "vendor",
                            slot = "misc",
                            page = 0,
                            x = 0,
                            y = 0,
                            width = 1,
                            height = 1,
                            inventory_dc6 = "data/global/items/invhp1.dc6",
                            vendor = "akara",
                            buy_price = 30,
                        },
                        {
                            id = "offer-tsc",
                            container = "vendor",
                            slot = "misc",
                            page = 0,
                            x = 1,
                            y = 0,
                            width = 1,
                            height = 1,
                            inventory_dc6 = "data/global/items/invbsc.dc6",
                            vendor = "akara",
                            buy_price = 100,
                        },
                    },
                }
            end,
        })
        test.mock_module("d2legacy.interactions.api", {
            snapshot = function()
                return { active = true, vendor = "Akara", categories = { "misc" } }
            end,
            dialogue = function()
                return { page = "trade" }
            end,
        })
        self.vendor = {}
        require("d2legacy.overlays.vendor").enter(self.vendor)
    end,
}
