-- Import trusted item facts into authoritative ECS state once.
--
-- The host decodes files and durable saves into plain value trees. This module
-- decides which Diablo components those facts become. After startup, commands
-- and systems consult ECS only; initial_data is never a mutable side channel.

local ecs = require("engine.ecs/v1")
local development_fixtures = require("d2legacy.items.development_fixtures")
local M = {}

local function layout_exists(owner)
    for _, entity in ipairs(ecs.query({
        all = { "d2legacy.items.layout" },
    })) do
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
    ecs.create({
        ["d2legacy.item.identity"] = {
            owner = layout,
            id = item.id,
            code = item.code,
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
            weapon_class = item.melee_weapon_class or "",
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
    return ecs.create({
        ["d2legacy.interaction.target"] = {
            id = target.id,
            npc = target.npc,
            vendor = target.vendor or "",
            categories = target.categories or "",
            services = target.services or "",
            x = target.x,
            y = target.y,
            radius = target.radius,
        },
    })
end

local function create_interactions(data, default_owner)
    local targets = {}
    for _, target in ipairs(data.targets or {}) do
        targets[target.id] = create_target(target)
    end

    -- An empty entity is a stable, serializable null-object reference. It lets
    -- the context component keep a required entity field while no panel is up.
    local no_target = ecs.create()
    local initial = targets[data.initial_target or ""] or no_target
    ecs.create({
        ["d2legacy.interaction.context"] = {
            owner = data.owner or default_owner,
            target = initial,
        },
    })
end

function M.load()
    local available, initial = pcall(require, "engine.initial_data/v1")
    if not available then return end

    local fixture_config = initial.get("d2legacy.development_items") or {}
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
	if not data and fixture_config.create_empty_containers == true then
		data = {
			owner = "local-player",
			inventory_width = 10, inventory_height = 4,
			stash_width = 6, stash_height = 8,
			cube_width = 3, cube_height = 4,
			belt_capacity = 4,
			vendor_width = 10, vendor_height = 10,
			active_weapon_set = 0,
			carried_gold = 0, stashed_gold = 0,
			items = {},
		}
    end
    if not data or not data.owner then return end

    -- A reconstructed runtime registers schemas and runs this composition root
    -- before checkpoint participant state is attached. Its ECS snapshot already
    -- contains the durable layout, items, vendor terms, service rules, and
    -- interaction context. Never import immutable creation facts a second time.
    if layout_exists(data.owner) then return end

    local layout = create_layout(data)
    for _, item in ipairs(data.items or {}) do
        create_item(layout, item)
    end
    create_vendor_terms(data.trade_terms)
    create_service_rules(data.service_rules)

    local interactions = initial.get("d2legacy.interactions") or {}
    create_interactions(interactions, data.owner)
end

return M
