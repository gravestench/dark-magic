local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("placement_slot_drives_stock_purchase_and_authority_quote", {
            test.run(function()
                local controls, drawn, texts = {}, {}, {}
                local bought, popped, page = nil, 0, "trade"
                local cancel, closes = false, 0
                local function node()
                    return {
                        set_position = function() end,
                        set_visible = function() end,
                        set_dc6 = function(_, path)
                            drawn[path] = true
                            return 30, 30
                        end,
                        set_dc6_combined = function()
                            return 320, 420
                        end,
                    }
                end
                test.mock_module("engine.render/v1", {
                    create = node,
                    assets_available = function()
                        return true
                    end,
                })
                test.mock_module("engine.input/v1", {
                    cursor = function()
                        return 45, 65
                    end,
                    pressed = function()
                        return cancel
                    end,
                })
                test.mock_module("engine.scene/v1", {
                    pop = function()
                        popped = popped + 1
                    end,
                })
                test.mock_module("engine.locale/v1", {
                    text = function(value)
                        return value
                    end,
                })
                test.mock_module("engine.data/v1", {
                    load_manifest = function()
                        return {
                            palettes = { units = "units", sky = "sky" },
                            screens = {
                                vendor = {
                                    sheet = "panel",
                                    x = 0,
                                    y = 0,
                                    palette = "sky",
                                    item_grid = { records = "inventory", record_class = "Monster" },
                                    tabs = {
                                        sheet = "tabs",
                                        categories = { "armo", "weap", "weap", "misc" },
                                        labels = { "a", "b", "c", "d" },
                                    },
                                    close = { x = 0, bottom_inset = 0, label = "close" },
                                },
                            },
                        }
                    end,
                })
                test.mock_module("engine.records/v1", {
                    load = function()
                        return {
                            {
                                class = "Monster",
                                gridLeft = "0",
                                invLeft = "0",
                                gridTop = "50",
                                gridBoxWidth = "30",
                                gridBoxHeight = "30",
                                gridX = "2",
                                gridY = "1",
                            },
                        }
                    end,
                })
                test.mock_module("d2legacy.ui.controls", {
                    new = function()
                        return {
                            add = function(_, value)
                                controls[value.id] = value
                            end,
                            update = function() end,
                        }
                    end,
                })
                test.mock_module("d2legacy.ui.button", { create = function() end })
                test.mock_module("d2legacy.ui.text", {
                    create = node,
                    set_font = function()
                        return 20, 10
                    end,
                    set = function(_, _, value)
                        texts[#texts + 1] = value
                    end,
                })
                test.mock_module("d2legacy.items.api", {
                    snapshot = function()
                        return {
                            carried_gold = 870,
                            stashed_gold = 9999,
                            items = {
                                {
                                    id = "offer",
                                    code = "tsc",
                                    container = "vendor",
                                    slot = "misc",
                                    page = 0,
                                    x = 1,
                                    y = 0,
                                    width = 1,
                                    height = 1,
                                    inventory_dc6 = "scroll",
                                    vendor = "akara",
                                    buy_price = 100,
                                },
                            },
                        }
                    end,
                    buy_to_inventory = function(id, vendor)
                        bought = { id, vendor }
                    end,
                })
                test.mock_module("d2legacy.interactions.api", {
                    snapshot = function()
                        return { active = true, vendor = "Akara", categories = { "misc" } }
                    end,
                    dialogue = function()
                        return { page = page }
                    end,
                    close = function()
                        closes = closes + 1
                    end,
                })
                local overlay = require("d2legacy.overlays.vendor")
                local instance = {}
                overlay.enter(instance)
                test.expect(drawn.scroll):is_true()
                test.expect(texts[#texts - 1]):equals("870")
                test.expect(texts[#texts]):equals("d2legacy.vendor.price 100")
                controls.vendor_1_0.on_activate()
                test.expect(bought):deep_equals({ "offer", "Akara" })
                cancel = true
                overlay.update(instance)
                overlay.update(instance)
                test.expect(closes):equals(1)
                test.expect(popped):equals(0)
                page = ""
                overlay.update(instance)
                test.expect(popped):equals(1)
            end),
        }),
    },
})
