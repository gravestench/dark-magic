-- Build the optional single-player development inventory from D2 records.
--
-- These fixtures make UI and combat labs useful, but every interpretation is
-- still real d2legacy policy: item footprint, body slots, melee range, 8.8
-- damage units, graphics, vendor category, and container placement. The Go
-- host only says whether development fixtures are enabled.

local records = require("engine.records/v1")

local M = {}

local component_tokens = {
    "HD", "TR", "LG", "RA", "LA", "RH", "LH", "SH",
    "S1", "S2", "S3", "S4", "S5", "S6", "S7", "S8",
}

local function integer(value, fallback)
    local parsed = tonumber(value)
    if not parsed then
        return fallback or 0
    end
    return math.floor(parsed)
end

local function by_code(path)
    local result = {}
    for _, row in ipairs(records.load(path)) do
        if row.code and row.code ~= "" then
            result[row.code] = row
        end
    end
    return result
end

local function item_asset(name)
    if not name or name == "" then
        return ""
    end
    return "data/global/items/" .. name .. ".dc6"
end

local function composite(component, appearance)
    appearance = string.upper(appearance or "")
    if appearance == "" then
        return ""
    end
    component = string.upper(component or "")
    for _, token in ipairs(component_tokens) do
        if component == token then
            return token .. "=" .. appearance
        end
    end
    local index = tonumber(component)
    local token = index and component_tokens[math.floor(index) + 1]
    return token and token .. "=" .. appearance or ""
end

local function common(id, row)
    return {
        id = id,
        code = row.code,
        type = row.type or "",
        type2 = row.type2 or "",
        material_flags = integer(row.bitfield1),
        identified = true,
        width = integer(row.invwidth),
        height = integer(row.invheight),
        base_cost = integer(row.cost),
        inventory_dc6 = item_asset(row.invfile),
        world_dc6 = item_asset(row.flippyfile),
        world_animated = true,
    }
end

local function append(items, item, placement)
    for key, value in pairs(placement) do
        item[key] = value
    end
    items[#items + 1] = item
end

local function weapon_items(items, weapons)
    local row = weapons.ssd
    if not row then
        return
    end
    local inventory = common("fixture-short-sword", row)
    inventory.body_slots = "rarm,larm"
    inventory.composite = composite(row.component, row.alternategfx)
    inventory.weapon_class = string.upper(row.wclass or "")
    inventory.melee_range = 1 + integer(row.rangeadder)
    inventory.physical_min = integer(row.mindam) * 256
    inventory.physical_max = integer(row.maxdam) * 256
    -- Weapons.txt speed is a penalty: the runtime attackrate contribution has
    -- the opposite sign (Phase Blade -30 => attackrate +30).
    inventory.attack_rate = -integer(row.speed)
    inventory.melee_weapon_class = inventory.weapon_class
    append(items, inventory, {container = "inventory"})

    local vendor = {}
    for key, value in pairs(inventory) do vendor[key] = value end
    vendor.id = "fixture-vendor-short-sword"
    append(items, vendor, {container = "vendor", slot = "weap"})
end

local function armor_items(items, armor)
    local row = armor.cap
    if not row then
        return
    end
    local hireling = common("fixture-hireling-cap", row)
    hireling.body_slots = "head"
    hireling.defense = integer(row.minac)
    hireling.speed_penalty = integer(row.speed)
    hireling.composite = composite(row.component, row.alternategfx)
    append(items, hireling, {container = "hireling", slot = "head"})

    local vendor = {}
    for key, value in pairs(hireling) do vendor[key] = value end
    vendor.id = "fixture-vendor-cap"
    append(items, vendor, {container = "vendor", slot = "armo"})
end

local function misc_item(items, misc, id, code, placement, belt_eligible)
    local row = misc[code]
    if not row then
        return
    end
    local item = common(id, row)
    item.belt_eligible = belt_eligible or false
    append(items, item, placement)
end

local function trade_terms()
    local result = {}
    for _, row in ipairs(records.load("data/global/excel/Npc.txt")) do
        if row.npc and row.npc ~= "" then
            result[row.npc] = {
                buy_multiplier = integer(row["buy mult"]),
                sell_multiplier = integer(row["sell mult"]),
                max_buy = integer(row["max buy"]),
            }
        end
    end
    return result
end

function M.build(enabled)
    local data = {
        owner = "local-player",
        belt_capacity = 4,
        active_weapon_set = 0,
        vendor_width = 10,
        vendor_height = 10,
        carried_gold = 10000,
        stashed_gold = 0,
        inventory_width = 10,
        inventory_height = 4,
        stash_width = 6,
        stash_height = 8,
        cube_width = 3,
        cube_height = 4,
        items = {},
        trade_terms = trade_terms(),
    }
    if not enabled then
        return data
    end

    weapon_items(data.items, by_code("data/global/excel/weapons.txt"))
    armor_items(data.items, by_code("data/global/excel/armor.txt"))
    local misc = by_code("data/global/excel/misc.txt")
    misc_item(data.items, misc, "fixture-hp1", "hp1", {container = "inventory", x = 2}, true)
    misc_item(data.items, misc, "fixture-mp1", "mp1", {container = "belt", belt_slot = 0}, true)
    misc_item(data.items, misc, "fixture-vendor-hp1", "hp1", {container = "vendor", slot = "misc"}, true)
    misc_item(data.items, misc, "fixture-rvs", "rvs", {container = "stash"}, true)
    misc_item(data.items, misc, "fixture-tsc", "tsc", {container = "cube"}, false)
    table.sort(data.items, function(left, right) return left.id < right.id end)
    return data
end

return M
