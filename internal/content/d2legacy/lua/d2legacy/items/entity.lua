-- Build authoritative item entities from decoded durable facts.

local ecs = require("engine.ecs/v1")
local records = require("engine.records/v1")
local world = require("d2legacy.items.world")
local M = {}
local item_types_by_code

local function item_type_index()
    if item_types_by_code then
        return item_types_by_code
    end
    item_types_by_code = {}
    local available, rows = pcall(records.load, "data/global/excel/ItemTypes.txt")
    rows = available and rows or {}
    for _, row in ipairs(rows) do
        local code = row.Code or row.code
        if code and code ~= "" then
            item_types_by_code[code] = row
        end
    end
    return item_types_by_code
end

local function item_type_closure(item)
    local index, found, visiting = item_type_index(), {}, {}
    local function visit(code)
        if not code or code == "" or found[code] then
            return
        end
        assert(not visiting[code], "cyclic ItemTypes inheritance at " .. code)
        visiting[code] = true
        found[code] = true
        local row = index[code]
        if row then
            visit(row.Equiv1 or row.equiv1)
            visit(row.Equiv2 or row.equiv2)
        end
        visiting[code] = nil
    end
    visit(item.type)
    visit(item.type2)
    local result = {}
    for code in pairs(found) do
        result[#result + 1] = code
    end
    table.sort(result)
    return table.concat(result, ",")
end

function M.components(layout, item)
    local components = {
        ["d2legacy.item.identity"] = {
            owner = layout,
            id = item.id,
            code = item.code,
            item_types = item_type_closure(item),
            material_flags = item.material_flags or 0,
            identified = item.identified ~= false,
            width = item.width,
            height = item.height,
            body_slots = item.body_slots or "",
            belt_eligible = item.belt_eligible or false,
            base_cost = item.base_cost or 0,
            applied_services = item.applied_services or "",
        },
        ["d2legacy.item.placement"] = {
            container = item.container or "world",
            x = item.x or 0,
            y = item.y or 0,
            slot = item.slot or "",
            belt_slot = item.belt_slot or 0,
            weapon_set = item.weapon_set or 0,
            page = item.page or 0,
        },
        ["d2legacy.item.presentation"] = {
            inventory_dc6 = item.inventory_dc6 or "",
            world_dc6 = item.world_dc6 or "",
            world_animated = item.world_animated or false,
            composite = item.composite or "",
            weapon_class = item.weapon_class or "",
        },
        ["d2legacy.item.melee"] = {
            range = item.melee_range or 0,
            physical_min = item.physical_min or 0,
            physical_max = item.physical_max or 0,
            attack_rating = item.attack_rating or 0,
            attack_rate = item.attack_rate or 0,
            weapon_class = item.melee_weapon_class or "",
        },
        ["d2legacy.item.armor"] = {
            defense = item.defense or 0,
            base_defense_max = item.base_defense_max or 0,
            speed_penalty = item.speed_penalty or 0,
        },
    }
    if item.generation then
        components["d2legacy.item.generation"] = item.generation
    end
    if item.origin then
        components["d2legacy.item.origin"] = { key = item.origin }
    end
    local layout_value = assert(ecs.get(layout, "d2legacy.items.layout"))
    for name, values in pairs(world.initial_components(layout_value:get("owner"), item)) do
        components[name] = values
    end
    return components
end

function M.create(layout, item)
    local item_entity = ecs.create(M.components(layout, item))

    -- Property generation and stat activation are separate layers. Import the
    -- already-normalized modifiers with their stable source identity intact;
    -- equipment policy decides later whether each source is active.
    local modifier_ids = {}
    for index, modifier in ipairs(item.stat_modifiers or {}) do
        local source_id = modifier.source_id or tostring(index)
        local source_kind = modifier.source_kind or "item"
        local order = modifier.order or index
        local identity = source_kind .. ":" .. tostring(order) .. ":" .. source_id
        assert(not modifier_ids[identity], "duplicate item stat modifier identity " .. identity)
        modifier_ids[identity] = true
        ecs.create({
            ["d2legacy.item.stat_modifier"] = {
                item = item_entity,
                source_id = source_id,
                source_kind = source_kind,
                stat = modifier.stat or "",
                operation = modifier.operation or "add",
                value = modifier.value or 0,
                order = order,
            },
        })
    end
    return item_entity
end

function M.reset()
    item_types_by_code = nil
end

return M
