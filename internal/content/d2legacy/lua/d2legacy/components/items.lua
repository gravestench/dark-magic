-- Declare the authoritative ECS vocabulary for items and interactions.
--
-- Components contain facts only. They do not decide whether an item fits,
-- what a vendor charges, or whether an NPC can be reached. Those decisions
-- live in command and system modules where they can be read as game rules.

local ecs = require("engine.ecs/v1")
local M = {}

local function register(name, fields)
    ecs.component({ name = name, fields = fields })
end

local function register_layout()
    register("d2legacy.items.layout", {
        { name = "owner", type = "string" },
        { name = "inventory_width", type = "i64" },
        { name = "inventory_height", type = "i64" },
        { name = "stash_width", type = "i64" },
        { name = "stash_height", type = "i64" },
        { name = "cube_width", type = "i64" },
        { name = "cube_height", type = "i64" },
        { name = "belt_capacity", type = "i64" },
        { name = "active_weapon_set", type = "i64" },
        { name = "vendor_width", type = "i64" },
        { name = "vendor_height", type = "i64" },
        { name = "carried_gold", type = "i64" },
        { name = "stashed_gold", type = "i64" },
    })
end

local function register_item_identity()
    register("d2legacy.item.identity", {
        { name = "owner", type = "entity" },
        { name = "id", type = "string" },
        { name = "code", type = "string" },
        { name = "width", type = "i64" },
        { name = "height", type = "i64" },
        { name = "body_slots", type = "string" },
        { name = "belt_eligible", type = "bool" },
        { name = "base_cost", type = "i64" },
        { name = "applied_services", type = "string" },
    })
end

local function register_item_location()
    register("d2legacy.item.placement", {
        { name = "container", type = "string" },
        { name = "x", type = "i64" },
        { name = "y", type = "i64" },
        { name = "slot", type = "string" },
        { name = "belt_slot", type = "i64" },
        { name = "weapon_set", type = "i64" },
        { name = "page", type = "i64" },
    })
end

local function register_item_views()
    register("d2legacy.item.presentation", {
        { name = "inventory_dc6", type = "string" },
        { name = "world_dc6", type = "string" },
        { name = "world_animated", type = "bool" },
        { name = "composite", type = "string" },
        { name = "weapon_class", type = "string" },
    })
    register("d2legacy.item.melee", {
        { name = "range", type = "f64" },
        { name = "physical_min", type = "i64" },
        { name = "physical_max", type = "i64" },
        { name = "attack_rating", type = "i64" },
        { name = "attack_rate", type = "i64" },
        { name = "weapon_class", type = "string" },
    })
    register("d2legacy.item.armor", {
        { name = "defense", type = "i64" },
        { name = "base_defense_max", type = "i64" },
        { name = "speed_penalty", type = "i64" },
    })
    -- Generated properties remain immutable item facts until location policy
    -- activates them. Keep the original source identity/kind and ordering so a
    -- later ItemStatCost-derived resolver can add percentage/op semantics
    -- without reconstructing provenance from an already-flattened total.
    register("d2legacy.item.stat_modifier", {
        { name = "item", type = "entity" },
        { name = "source_id", type = "string" },
        { name = "source_kind", type = "string" },
        { name = "stat", type = "string" },
        { name = "operation", type = "string" },
        { name = "value", type = "i64" },
        { name = "order", type = "i64" },
    })
end

local function register_vendor_terms()
    register("d2legacy.vendor.terms", {
        { name = "vendor", type = "string" },
        { name = "buy_multiplier", type = "i64" },
        { name = "sell_multiplier", type = "i64" },
        { name = "max_buy", type = "i64" },
    })
    register("d2legacy.item.service_rule", {
        { name = "id", type = "string" },
        { name = "target_slot", type = "string" },
        { name = "consume_slots", type = "string" },
        { name = "gold_cost", type = "i64" },
    })
end

local function register_interactions()
    register("d2legacy.interaction.target", {
        { name = "id", type = "string" },
        { name = "npc", type = "string" },
        { name = "vendor", type = "string" },
        { name = "categories", type = "string" },
        { name = "services", type = "string" },
        { name = "x", type = "f64" },
        { name = "y", type = "f64" },
        { name = "radius", type = "f64" },
    })
    register("d2legacy.interaction.context", {
        { name = "owner", type = "string" },
        { name = "target", type = "entity" },
    })
    -- Each context owns one explicit null-object while no interaction is open.
    -- Marking it makes lifecycle cleanup exact; untyped empty entities are
    -- otherwise indistinguishable and leak across close/disconnect cycles.
    register("d2legacy.interaction.null_target", {})
end

function M.register()
    register_layout()
    register_item_identity()
    register_item_location()
    register_item_views()
    register_vendor_terms()
    register_interactions()
end

return M
