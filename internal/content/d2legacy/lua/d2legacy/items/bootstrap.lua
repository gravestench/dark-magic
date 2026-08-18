-- Import trusted item facts into authoritative ECS state once.
--
-- The host decodes files and durable saves into plain value trees. This module
-- decides which Diablo components those facts become. After startup, commands
-- and systems consult ECS only; initial_data is never a mutable side channel.

local ecs = require("engine.ecs/v1")
local records = require("engine.records/v1")
local development_fixtures = require("d2legacy.items.development_fixtures")
local world = require("d2legacy.items.world")
local M = {}
local interaction_data = {}
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
        if not code or code == "" or found[code] then return end
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
    for code in pairs(found) do result[#result + 1] = code end
    table.sort(result)
    return table.concat(result, ",")
end

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

local function create_item(layout, item)
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
    local layout_value = assert(ecs.get(layout, "d2legacy.items.layout"))
    for name, values in pairs(world.initial_components(layout_value:get("owner"), item)) do
        components[name] = values
    end
    local item_entity = ecs.create(components)

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
    local wanted = string.lower(target.id)
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.interaction.target" } })) do
        if ecs.get(entity, "d2legacy.interaction.target"):get("id") == wanted then
            return entity
        end
    end
    local components = {
        ["d2legacy.interaction.target"] = {
            id = wanted,
            npc = target.npc,
            vendor = target.vendor or "",
            categories = target.categories or "",
            services = target.services or "",
            x = target.x,
            y = target.y,
            radius = target.radius,
        },
    }
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

function M.load()
    item_types_by_code = nil
    local available, initial = pcall(require, "engine.initial_data/v1")
    if not available then
        return
    end

    local fixture_config = initial.get("d2legacy.development_items") or {}
    interaction_data = initial.get("d2legacy.interactions") or {}
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
        return
    end

    -- A reconstructed runtime registers schemas and runs this composition root
    -- before checkpoint participant state is attached. Its ECS snapshot already
    -- contains the durable layout, items, vendor terms, service rules, and
    -- interaction context. Never import immutable creation facts a second time.
    if layout_exists(data.owner) then
        return
    end

    local layout = create_layout(data)
    for _, item in ipairs(data.items or {}) do
        create_item(layout, item)
    end
    create_vendor_terms(data.trade_terms)
    create_service_rules(data.service_rules)

    create_interactions(interaction_data, data.owner)
end

return M
