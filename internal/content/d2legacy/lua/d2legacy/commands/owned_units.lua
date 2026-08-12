-- Attach Diablo ownership and lifetime policy to an existing world entity.
--
-- Stable selectable IDs cross the command boundary.  Lua resolves them to ECS
-- entities, applies category/group limits, and records the relationship.  The
-- engine only supplies generic commands and ECS storage; it does not know what
-- a summon, pet, trap, minion, or hireling means.

local authority = require("engine.authority_command/v1")
local ecs = require("engine.ecs/v1")
local limits = require("d2legacy.owned_units.limits")
local M = {}

local replacements = {reject=true,replace_oldest=true,replace_newest=true}

local function entity_by_id(wanted)
    for _, entity in ipairs(ecs.query({all={"d2legacy.world.selectable"}})) do
        if ecs.get(entity, "d2legacy.world.selectable"):get("id") == wanted then
            return entity
        end
    end
    return nil
end

local function category_from(payload)
    local value = payload.category
    assert(type(value) == "table", "owned-unit category is required")
    assert(type(value.id) == "string" and value.id ~= "", "category ID is required")
    assert(type(value.group) == "number" and value.group >= 0, "category group must be non-negative")
    assert(type(value.base_max) == "number" and value.base_max >= 1, "category maximum must be positive")
    assert(replacements[value.replacement], "invalid owned-unit replacement policy")
    return value
end

local function candidates(owner)
    local result = {}
    for _, entity in ipairs(ecs.query({all={"d2legacy.owned_unit"}})) do
        local relation = ecs.get(entity, "d2legacy.owned_unit")
        if relation:get("owner"):id() == owner:id() then
            result[#result + 1] = {
                entity=entity, category=relation:get("category"),
                group=relation:get("group"), active=relation:get("active"),
                created_tick=relation:get("created_tick"),
            }
        end
    end
    return result
end

function M.validate(command)
    local payload = command.payload
    assert(type(payload) == "table", "owned-unit payload is required")
    assert(type(payload.unit_id) == "string" and payload.unit_id ~= "", "unit ID is required")
    assert(type(payload.owner_id) == "string" and payload.owner_id ~= "", "owner ID is required")
    assert(type(payload.ultimate_owner_id) == "string" and payload.ultimate_owner_id ~= "",
        "ultimate owner ID is required")
    category_from(payload)
end

function M.apply(command)
    local payload, category = command.payload, category_from(command.payload)
    local unit, owner = entity_by_id(payload.unit_id), entity_by_id(payload.owner_id)
    assert(unit and owner and unit:id() ~= owner:id(), "distinct live unit and owner are required")
    assert(not ecs.get(unit, "d2legacy.owned_unit"), "unit already has an owner")

    local victims = limits.victims(candidates(owner), category)
    ecs.set(unit, "d2legacy.owned_unit", {
        owner=owner, owner_id=payload.owner_id,
        ultimate_owner_id=payload.ultimate_owner_id,
        category=category.id, group=category.group, limit=category.base_max,
        replacement=category.replacement, created_tick=command.tick,
        expires_tick=payload.expires_tick or 0, durable_id=payload.durable_id or "",
        durable=category.durable or false, unsummon=category.unsummon or false,
        warp_with_owner=category.warp_with_owner or false,
        range_limited=category.range_limited or false, active=true,
        survives_owner_death=category.survives_owner_death or false,
    })
    for _, victim in ipairs(victims) do
        ecs.get(victim.entity, "d2legacy.owned_unit"):set("active", false)
    end
end

function M.register()
    authority.register({
        kind="system.owned_unit.attach",
        authorities={"system","administrator"},
        validate=M.validate,
        apply=M.apply,
    })
end

return M
