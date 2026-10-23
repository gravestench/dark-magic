-- Import trusted item facts into authoritative ECS state once.
--
-- The host decodes files and durable saves into plain value trees. This module
-- decides which Diablo components those facts become. After startup, commands
-- and systems consult ECS only; initial_data is never a mutable side channel.

local ecs = require("engine.ecs/v1")
local development_fixtures = require("d2legacy.items.development_fixtures")
local item_entity = require("d2legacy.items.entity")
local M = {}
local interaction_data = {}

-- Cain and ambient actors need their own quest or behavior contracts.
local initial_town_npcs = { akara = true, charsi = true, gheed = true, kashya = true, warriv1 = true }

local function layout_exists(owner)
    for _, entity in
        ipairs(ecs.query({
            all = { "d2legacy.items.layout" },
        }))
    do
        if ecs.get(entity, "d2legacy.items.layout"):get("owner") == owner then
            return true
        end
    end
    return false
end

local function create_layout(data)
    return ecs.create({
        ["d2legacy.items.layout"] = {
            owner = data.owner,
            inventory_width = data.inventory_width or 0,
            inventory_height = data.inventory_height or 0,
            stash_width = data.stash_width or 0,
            stash_height = data.stash_height or 0,
            cube_width = data.cube_width or 0,
            cube_height = data.cube_height or 0,
            belt_capacity = data.belt_capacity or 0,
            active_weapon_set = data.active_weapon_set or 0,
            vendor_width = data.vendor_width or 0,
            vendor_height = data.vendor_height or 0,
            carried_gold = data.carried_gold or 0,
            stashed_gold = data.stashed_gold or 0,
        },
    })
end

local function create_vendor_terms(trade_terms)
    for vendor, terms in pairs(trade_terms or {}) do
        ecs.create({
            ["d2legacy.vendor.terms"] = {
                vendor = string.lower(vendor),
                buy_multiplier = terms.buy_multiplier or 0,
                sell_multiplier = terms.sell_multiplier or 0,
                max_buy = terms.max_buy or 0,
            },
        })
    end
end

local function create_service_rules(service_rules)
    for _, rule in ipairs(service_rules or {}) do
        ecs.create({
            ["d2legacy.item.service_rule"] = {
                id = string.lower(rule.id),
                target_slot = rule.target_slot,
                consume_slots = rule.consume_slots or "",
                gold_cost = rule.gold_cost or 0,
            },
        })
    end
end

local function create_target(target)
    if target.gold_amount ~= nil then
        local amount = target.gold_amount
        local function finite(value)
            return type(value) == "number" and value == value and value ~= math.huge and value ~= -math.huge
        end
        assert(type(amount) == "number" and amount % 1 == 0 and amount >= 1 and amount <= 2147483647,
            "gold pile amount must be a positive signed-32-bit integer")
        assert(target.selectable == true, "gold pile must be selectable")
        assert(finite(target.act) and target.act % 1 == 0 and target.act > 0, "gold pile act is required")
        assert(finite(target.level_id) and target.level_id % 1 == 0 and target.level_id > 0,
            "gold pile level is required")
        assert(finite(target.x) and finite(target.y), "gold pile position is required")
        assert(finite(target.radius) and target.radius > 0, "gold pile radius is required")
    end
    local wanted = string.lower(target.id)
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.interaction.target" } })) do
        if ecs.get(entity, "d2legacy.interaction.target"):get("id") == wanted then
            return entity
        end
    end
    local npc_class = target.class or (wanted == "act1-akara" and "akara")
    local appearance
    if target.level_id == 1 and initial_town_npcs[npc_class] then
        appearance = require("d2legacy.data.monster").npc_appearance(npc_class)
        appearance.mode, appearance.death_sound = "NU", ""
    end
    local label = appearance and appearance.name_key or target.npc
    local components = {
        ["d2legacy.interaction.target"] = {
            id = wanted,
            npc = label,
            vendor = target.vendor or "",
            categories = target.categories or "",
            services = target.services or "",
            x = target.x,
            y = target.y,
            radius = target.radius,
        },
    }
    if target.gold_amount ~= nil then
        components["d2legacy.gold.pile"] = { amount = target.gold_amount }
    end
    if target.selectable or appearance then
        local act = appearance and 1 or target.act
        assert(type(act) == "number" and act > 0, "selectable act is required")
        assert(type(target.level_id) == "number" and target.level_id > 0, "selectable level is required")
        components["d2legacy.world.position"] = { x = target.x, y = target.y }
        components["d2legacy.world.location"] = { act = act, level_id = target.level_id }
        components["d2legacy.world.selectable"] = {
            id = wanted, kind = "dynamic-object", label = label,
            radius = target.radius, priority = 20,
        }
        if appearance then
            components["d2legacy.monster.appearance"] = appearance
        end
    end
    if target.room_id ~= nil then
        assert(type(target.resident_id) == "string" and target.resident_id ~= "", "resident ID is required")
        assert(type(target.level_id) == "number" and target.level_id > 0, "resident level is required")
        components["d2legacy.world.room_resident"] = {
            id = target.resident_id,
            level_id = target.level_id,
            room_id = tostring(target.room_id),
        }
    end
    if target.object_state then
        local state = target.object_state
        assert(type(state) == "table", "object state must be a table")
        components["d2legacy.object.state"] = {
            id = target.id,
            definition_id = state.definition_id or "",
            mode = state.mode or "",
            used = state.used or false,
            locked = state.locked or false,
            disabled = state.disabled or false,
            seed = state.seed or 0,
            revision = state.revision or 0,
        }
        if state.once_result_mode then
            assert(
                type(state.once_result_mode) == "string" and state.once_result_mode ~= "",
                "one-shot object result mode is required"
            )
            components["d2legacy.object.once_operation"] = {
                result_mode = state.once_result_mode,
            }
        end
    end
    local entity = ecs.create(components)
    local action_ids = {}
    for index, action in ipairs(target.pending_actions or {}) do
        local action_id = action.id
        assert(type(action_id) == "string" and action_id ~= "", "object pending-action ID is required")
        assert(not action_ids[action_id], "duplicate object pending-action ID " .. action_id)
        action_ids[action_id] = true
        local action_components = {
            ["d2legacy.object.pending_action"] = {
                id = action_id,
                target = entity,
                kind = action.kind or "",
                due_tick = action.due_tick or 0,
                sequence = action.sequence or index,
                active = action.active ~= false,
            },
        }
        if target.room_id ~= nil then
            action_components["d2legacy.world.room_resident"] = {
                id = "object-action:" .. target.id .. ":" .. action_id,
                level_id = target.level_id,
                room_id = tostring(target.room_id),
            }
        end
        ecs.create(action_components)
    end
    return entity
end

local function create_interactions(data, default_owner)
    local targets = {}
    for _, target in ipairs(data.targets or {}) do
        targets[target.id] = create_target(target)
    end

    -- An empty entity is a stable, serializable null-object reference. It lets
    -- the context component keep a required entity field while no panel is up.
    local no_target = ecs.create({ ["d2legacy.interaction.null_target"] = {} })
    local initial = targets[data.initial_target or ""] or no_target
    ecs.create({
        ["d2legacy.interaction.context"] = {
            owner = default_owner or data.owner,
            target = initial,
        },
    })
end

local function interaction_context_exists(owner)
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.interaction.context" } })) do
        if ecs.get(entity, "d2legacy.interaction.context"):get("owner") == owner then
            return true
        end
    end
    return false
end

local function empty_layout(owner)
    return {
        owner = owner,
        inventory_width = 10,
        inventory_height = 4,
        stash_width = 6,
        stash_height = 8,
        cube_width = 3,
        cube_height = 4,
        belt_capacity = 4,
        vendor_width = 10,
        vendor_height = 10,
        active_weapon_set = 0,
        carried_gold = 0,
        stashed_gold = 0,
        items = {},
    }
end

-- Player-owned containers are admission state, not process startup state.
-- This keeps listen/realm authorities from inventing one "local-player"
-- inventory before they know which authenticated members will enter.
function M.ensure_player(owner)
    assert(type(owner) == "string" and owner:match("%S"), "player owner is required")
    if not layout_exists(owner) then
        create_layout(empty_layout(owner))
    end
    if not interaction_context_exists(owner) then
        create_interactions(interaction_data, owner)
    end
end

local function ensure_world_source()
    local sources = ecs.query({ all = { "d2legacy.items.world_source" } })
    assert(#sources <= 1, "multiple world item source layouts")
    if #sources == 1 then return end
    local layout = create_layout({ owner = "", inventory_width = 0, inventory_height = 0,
        stash_width = 0, stash_height = 0, cube_width = 0, cube_height = 0,
        belt_capacity = 0, vendor_width = 0, vendor_height = 0 })
    ecs.set(layout, "d2legacy.items.world_source", {})
end

function M.load()
    item_entity.reset()
    local available, initial = pcall(require, "engine.initial_data/v1")
    if not available then
        ensure_world_source()
        return
    end

    local fixture_config = initial.get("d2legacy.development_items") or {}
    interaction_data = initial.get("d2legacy.interactions") or {}
    -- This session source marks completed bootstrap, including a checkpoint where a fixture owner departed.
    -- Reconstructing that checkpoint must not recreate the departed owner's initial possessions.
    if #ecs.query({ all = { "d2legacy.items.world_source" } }) > 0 then
        ensure_world_source()
        return
    end
    -- Tests, save importers, and servers may supply already-decoded durable
    -- item facts. The development catalog is only a fallback explicitly
    -- requested by the interactive client.
    local data = initial.get("d2legacy.items")
    if not data and fixture_config.enabled == true then
        data = development_fixtures.build(true)
    end
    -- A newly created character owns empty containers before any durable item
    -- import exists. Their dimensions and gold policy are Diablo rules, so the
    -- first-party mod supplies them instead of asking the generic host to know
    -- what an inventory, stash, cube, belt, or vendor page looks like.
    -- create_empty_containers is consumed by player admission through
    -- ensure_player. Startup cannot assign per-player ownership safely.
    if not data or not data.owner then
        ensure_world_source()
        return
    end

    assert(type(data.owner) == "string" and data.owner:match("%S"), "item bootstrap player owner is required")

    -- A reconstructed runtime registers schemas and runs this composition root
    -- before checkpoint participant state is attached. Its ECS snapshot already
    -- contains the durable layout, items, vendor terms, service rules, and
    -- interaction context. Never import immutable creation facts a second time.
    if layout_exists(data.owner) then
        ensure_world_source()
        return
    end

    local layout = create_layout(data)
    for _, item in ipairs(data.items or {}) do
        item_entity.create(layout, item)
    end
    create_vendor_terms(data.trade_terms)
    create_service_rules(data.service_rules)

    create_interactions(interaction_data, data.owner)
    ensure_world_source()
end

return M
