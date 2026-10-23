local test = require("d2legacy.tests/v1")

local function row(code, cost)
    return {
        code = code,
        cost = cost,
        spawnable = "1",
        PermStoreItem = "1",
        AkaraMax = "1",
        stackable = "0",
        invwidth = "1",
        invheight = "1",
        invfile = "synthetic",
    }
end

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("eligible_permanent_subset_has_record_price", {
            test.run(function()
                local excluded = row("hp1", "30")
                excluded.PermStoreItem = "0"
                test.mock_module("engine.records/v1", {
                    load = function(path)
                        if path:find("npc.txt", 1, true) then
                            return { { npc = "akara", ["sell mult"] = "1088" } }
                        end
                        return { row("tsc", "100"), excluded, row("other", "1") }
                    end,
                })
                local result = require("d2legacy.data.vendor_consumables").load()
                test.expect(#result):equals(1)
                test.expect(result[1].code):equals("tsc")
                test.expect(result[1].price):equals(106)
            end),
        }),
        test.case("valid_empty_stock_is_not_malformed_stock", {
            test.run(function()
                local misc = row("hp1", "30")
                misc.AkaraMax = "0"
                test.mock_module("engine.records/v1", {
                    load = function(path)
                        if path:find("npc.txt", 1, true) then
                            return { { npc = "akara", ["sell mult"] = "1024" } }
                        end
                        return { misc }
                    end,
                })
                local data = require("d2legacy.data.vendor_consumables")
                test.expect(#data.load()):equals(0)
                misc.AkaraMax, misc.cost = "1", "invalid"
                test.expect(pcall(data.load)):is_false()
            end),
        }),
    },
})
