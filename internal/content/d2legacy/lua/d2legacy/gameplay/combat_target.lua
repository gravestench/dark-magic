-- Normalize current unit-target legality for authoritative combat systems.
--
-- The command's target ID is only a request. Every consumer re-resolves it
-- against current ECS state, alignment, life, location, footprint range, and
-- the collision map before beginning or resolving a melee action.

local ecs = require("engine.ecs/v1")
local collision = require("d2legacy.gameplay.collision")
local melee = require("d2legacy.policy.melee")
local M = {}

local function alive(entity, kind)
    if kind == "hostile" or kind == "friendly" then
        local stats = ecs.get(entity, "d2legacy.monster.stats")
        return stats and stats:get("health") > 0
    end
    if kind == "player" then
        local vitals = ecs.get(entity, "d2legacy.player.vitals")
        return vitals and vitals:get("health") > 0
    end
    return false
end

local function opponents(attacker_kind, target_kind)
    return (attacker_kind == "player" and target_kind == "hostile")
        or (attacker_kind == "hostile" and target_kind == "player")
        or (attacker_kind == "friendly" and target_kind == "hostile")
        or (attacker_kind == "hostile" and target_kind == "friendly")
end

function M.unit(attacker, target)
    if not target or attacker:id() == target:id() then
        return nil
    end
    local attacker_selectable = ecs.get(attacker, "d2legacy.world.selectable")
    local target_selectable = ecs.get(target, "d2legacy.world.selectable")
    local attacker_position = ecs.get(attacker, "d2legacy.world.position")
    local target_position = ecs.get(target, "d2legacy.world.position")
    local attacker_location = ecs.get(attacker, "d2legacy.world.location")
    local target_location = ecs.get(target, "d2legacy.world.location")
    local attacker_collider = ecs.get(attacker, "d2legacy.world.collider")
    local target_collider = ecs.get(target, "d2legacy.world.collider")
    if
        not attacker_selectable
        or not target_selectable
        or not attacker_position
        or not target_position
        or not attacker_location
        or not target_location
        or not attacker_collider
        or not target_collider
        or not opponents(attacker_selectable:get("kind"), target_selectable:get("kind"))
        or not alive(attacker, attacker_selectable:get("kind"))
        or not alive(target, target_selectable:get("kind"))
        or attacker_location:get("act") ~= target_location:get("act")
        or attacker_location:get("level_id") ~= target_location:get("level_id")
    then
        return nil
    end
    local dx = target_position:get("x") - attacker_position:get("x")
    local dy = target_position:get("y") - attacker_position:get("y")
    return {
        id = target_selectable:get("id"),
        dx = dx,
        dy = dy,
        distance = math.sqrt(dx * dx + dy * dy),
        attacker_x = attacker_position:get("x"),
        attacker_y = attacker_position:get("y"),
        target_x = target_position:get("x"),
        target_y = target_position:get("y"),
        level_id = attacker_location:get("level_id"),
        attacker_radius = attacker_collider:get("radius"),
        target_radius = target_collider:get("radius"),
    }
end

function M.named(attacker, wanted, candidates)
    for _, candidate in ipairs(candidates) do
        local selectable = ecs.get(candidate, "d2legacy.world.selectable")
        if selectable and selectable:get("id") == wanted then
            local facts = M.unit(attacker, candidate)
            if facts then
                return candidate, facts
            end
            return nil
        end
    end
end

function M.in_melee_range(facts, attack_range)
    return M.within_reach(facts, attack_range)
        and collision.melee_clear(facts.level_id, facts.attacker_x, facts.attacker_y, facts.target_x, facts.target_y)
end

function M.within_reach(facts, attack_range)
    return facts.distance <= melee.reach(attack_range, facts.attacker_radius, facts.target_radius)
end

function M.select_melee(attacker, wanted, attack_range, candidates)
    local best, best_facts
    for _, candidate in ipairs(candidates) do
        local selectable = ecs.get(candidate, "d2legacy.world.selectable")
        if selectable and (wanted == "" or selectable:get("id") == wanted) then
            local facts = M.unit(attacker, candidate)
            if
                facts
                and M.in_melee_range(facts, attack_range)
                and (
                    not best_facts
                    or facts.distance < best_facts.distance
                    or facts.distance == best_facts.distance and facts.id < best_facts.id
                )
            then
                best, best_facts = candidate, facts
            end
        end
    end
    return best, best_facts and best_facts.id or ""
end

return M
