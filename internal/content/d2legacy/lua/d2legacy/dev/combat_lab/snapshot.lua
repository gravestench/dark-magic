-- Copy combat facts out of ECS for Combat Lab diagnostics.
--
-- A returned table is ordinary Lua data. Diagnostics may format and remember
-- it, but they never retain writable components or influence simulation.

local ecs
local M = {}

local function optional(entity, name)
    local ok, component = pcall(ecs.get, entity, name)
    if not ok or not component then return nil end
    return component:snapshot()
end

local function first(query)
    local ok, entities = pcall(ecs.query, query)
    if not ok then return nil end
    return entities[1]
end

local function copy_player()
    local entity = first({all={"d2legacy.world.player_control", "d2legacy.world.position", "d2legacy.world.location"}})
    if not entity then return nil end
    local result = {
        entity_id = entity:id(),
        position = optional(entity, "d2legacy.world.position"),
        location = optional(entity, "d2legacy.world.location"),
        velocity = optional(entity, "d2legacy.world.velocity"),
        collider = optional(entity, "d2legacy.world.collider"),
        profile = optional(entity, "d2legacy.combat.melee_profile"),
        animation = optional(entity, "d2legacy.player.animation"),
        approach = optional(entity, "d2legacy.combat.attack_approach"),
        attack = optional(entity, "d2legacy.combat.attack_animation"),
        vitals = optional(entity, "d2legacy.player.vitals"),
    }
    local selectable = optional(entity, "d2legacy.world.selectable")
    result.id = selectable and selectable.id or "local-player"
    return result
end

local function copy_monsters(level_id)
    local result = {}
    local ok, entities = pcall(ecs.query, {all={
        "d2legacy.monster.identity", "d2legacy.monster.stats", "d2legacy.world.position",
        "d2legacy.world.location", "d2legacy.world.selectable",
    }})
    if not ok then return result end
    for _, entity in ipairs(entities) do
        local location = optional(entity, "d2legacy.world.location")
        if location and location.level_id == level_id then
            local identity = optional(entity, "d2legacy.monster.identity") or {}
            local selectable = optional(entity, "d2legacy.world.selectable") or {}
            local label = selectable.label or ""
            -- Several recovered name fields still contain unresolved `$...`
            -- keys. A typed monster ID is more useful in a diagnostic than
            -- printing that implementation detail as though it were a name.
            if label == "" or string.sub(label, 1, 1) == "$" then
                label = identity.definition_id or identity.base_id or "HOSTILE"
            end
            result[#result + 1] = {
                entity_id = entity:id(),
                id = selectable.id or identity.spawn_id or tostring(entity:id()),
                label = label,
                position = optional(entity, "d2legacy.world.position"),
                velocity = optional(entity, "d2legacy.world.velocity"),
                collider = optional(entity, "d2legacy.world.collider"),
                stats = optional(entity, "d2legacy.monster.stats"),
                ai = optional(entity, "d2legacy.monster.ai"),
                appearance = optional(entity, "d2legacy.monster.appearance"),
                death = optional(entity, "d2legacy.monster.death"),
            }
        end
    end
    table.sort(result, function(left, right) return left.id < right.id end)
    return result
end

local function copy_events()
    local result = {}
    for _, component in ipairs({"d2legacy.combat.event", "d2legacy.combat.melee_event",
        "d2legacy.combat.event"}) do
        local ok, entities = pcall(ecs.query, {all={component}})
        if ok then
            for _, entity in ipairs(entities) do
                local event = optional(entity, component)
                if event then
                    event.entity_id = entity:id()
                    result[#result + 1] = event
                end
            end
        end
    end
    table.sort(result, function(left, right)
        if left.tick ~= right.tick then return left.tick < right.tick end
        return left.entity_id < right.entity_id
    end)
    return result
end

function M.read(level_id)
    -- Scene modules are cataloged by frontend-only hosts too. Acquire gameplay
    -- capability only after the activated Combat Lab actually asks for facts.
    ecs = ecs or require("engine.ecs/v1")
    local events = copy_events()
    local latest_tick = events[#events] and events[#events].tick or 0
    return {
        tick = latest_tick,
        player = copy_player(),
        monsters = copy_monsters(level_id),
        events = events,
    }
end

return M
