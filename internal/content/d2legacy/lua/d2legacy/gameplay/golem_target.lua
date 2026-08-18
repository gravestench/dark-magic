-- Resolve point and item targets for the record-driven golem family. The same
-- function is called before mana payment and again at the effect tick.

local ecs = require("engine.ecs/v1")
local M = {}

local function local_percentages(item, entities)
    local result = { defense = 0, damagepercent = 0 }
    for _, entity in ipairs(entities) do
        local modifier = ecs.get(entity, "d2legacy.item.stat_modifier")
        if modifier and modifier:get("item"):id() == item:id() and modifier:get("operation") == "local_percent" then
            local stat = modifier:get("stat")
            if result[stat] == nil then
                return nil, "item_modifier_unsupported"
            end
            result[stat] = result[stat] + modifier:get("value")
        end
    end
    return result
end

local function type_set(encoded)
    local result = {}
    for value in string.gmatch(encoded or "", "[^,]+") do
        result[value] = true
    end
    return result
end

local function item_is_admitted(identity, definition)
    local flags = identity:get("material_flags")
    local required = definition.item_required_material_flag or 0
    if required > 0 and math.floor(flags / required) % 2 ~= 1 then
        return false, "item_material_ineligible"
    end
    local types = type_set(identity:get("item_types"))
    local allowed = false
    for _, kind in ipairs(definition.item_allowed_types or {}) do
        allowed = allowed or types[kind]
    end
    if not allowed then
        return false, "item_type_ineligible"
    end
    for _, kind in ipairs(definition.item_excluded_types or {}) do
        if types[kind] then
            return false, "item_type_excluded"
        end
    end
    return true
end

function M.resolve(caster, target_x, target_y, target_id, definition, entities)
    local location = assert(ecs.get(caster, "d2legacy.world.location"))
    if not definition.requires_item_target then
        if type(target_x) ~= "number" or type(target_y) ~= "number" then
            return nil, "point_required"
        end
        return { x = target_x, y = target_y }
    end
    for _, entity in ipairs(entities) do
        local identity = ecs.get(entity, "d2legacy.item.identity")
        if identity and identity:get("id") == target_id then
            if ecs.get(entity, "d2legacy.world.inactive") then
                return nil, "item_unavailable"
            end
            local placement = ecs.get(entity, "d2legacy.item.placement")
            local position = ecs.get(entity, "d2legacy.world.position")
            local item_location = ecs.get(entity, "d2legacy.world.location")
            if not placement or placement:get("container") ~= "world" or not position or not item_location then
                return nil, "item_not_on_ground"
            end
            if
                item_location:get("act") ~= location:get("act")
                or item_location:get("level_id") ~= location:get("level_id")
            then
                return nil, "item_wrong_level"
            end
            local admitted, reason = item_is_admitted(identity, definition)
            if not admitted then
                return nil, reason
            end
            local melee = ecs.get(entity, "d2legacy.item.melee")
            local armor = ecs.get(entity, "d2legacy.item.armor")
            local local_percent, modifier_reason = local_percentages(entity, entities)
            if not local_percent then
                return nil, modifier_reason
            end
            local weapon_minimum_raw = melee and melee:get("physical_min") or 0
            local weapon_maximum_raw = melee and melee:get("physical_max") or 0
            weapon_minimum_raw = math.floor(weapon_minimum_raw * (100 + local_percent.damagepercent) / 100)
            weapon_maximum_raw = math.floor(weapon_maximum_raw * (100 + local_percent.damagepercent) / 100)
            local item_defense = armor and armor:get("defense") or 0
            if local_percent.defense ~= 0 and armor and armor:get("base_defense_max") > 0 then
                -- Enhanced Defense rerolls an armor base to its authored
                -- maximum plus one before applying the local percentage.
                item_defense = armor:get("base_defense_max") + 1
            end
            item_defense = math.floor(item_defense * (100 + local_percent.defense) / 100)
            return {
                x = position:get("x"),
                y = position:get("y"),
                entity = entity,
                item_id = identity:get("id"),
                item_code = identity:get("code"),
                item_types = identity:get("item_types"),
                identified = identity:get("identified"),
                weapon_minimum_raw = weapon_minimum_raw,
                weapon_maximum_raw = weapon_maximum_raw,
                item_defense = item_defense,
            }
        end
    end
    return nil, "item_missing"
end

return M
