-- A bounded subset of Akara's permanent Normal stock, not random shop generation.
-- D2MOO 1.10f separates eligible PermStoreItem rows from randomized vendor stock.
-- Source pin 1fe7444dc2cfb6bf1c2934b8901e5699a4c36913:
-- SUNITPROXY_FillIGlobaltemCacheRecordForNpc and ITEMS_CalculateTransactionCost.
local records = require("engine.records/v1")
local M = {}

local function integer(row, key)
    local value = tonumber(row[key] or "")
    assert(value and value >= 0 and value % 1 == 0, "invalid vendor field: " .. key)
    return value
end

function M.load()
    local multiplier
    for _, row in ipairs(records.load("data/global/excel/npc.txt")) do
        if string.lower(row.npc or "") == "akara" then
            multiplier = integer(row, "sell mult")
            break
        end
    end
    if not multiplier then
        return {}
    end
    local result = {}
    for _, row in ipairs(records.load("data/global/excel/misc.txt")) do
        if
            (row.code == "tsc" or row.code == "hp1")
            and row.spawnable == "1"
            and row.PermStoreItem == "1"
            and ((tonumber(row.AkaraMax) or 0) > 0 or (tonumber(row.AkaraMagicMax) or 0) > 0)
        then
            assert(row.stackable == "0", "vendor consumable must be nonstackable")
            local width, height = integer(row, "invwidth"), integer(row, "invheight")
            assert(width == 1 and height == 1, "unsupported vendor consumable footprint")
            assert(type(row.invfile) == "string" and row.invfile ~= "", "vendor consumable art is required")
            local cost = integer(row, "cost")
            result[#result + 1] = {
                code = row.code,
                width = width,
                height = height,
                base_cost = cost,
                price = math.max(1, math.floor(cost * multiplier / 1024)),
                inventory_dc6 = "data/global/items/" .. row.invfile .. ".dc6",
                belt_eligible = row.code == "hp1",
            }
        end
    end
    table.sort(result, function(a, b)
        return a.code < b.code
    end)
    return result
end

return M
