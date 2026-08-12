-- Resolve impact requests emitted by reusable approach/animation mechanisms.

local ecs = require("dm.ecs/v1")
local policy = require("d2legacy.policy.melee")
local M = {}

local function identity(entity)
    local selectable = ecs.get(entity, "dm.world.selectable")
    return selectable and selectable:get("id") or "entity:" .. entity:id()
end

local function target_for(attacker, wanted, candidates)
    local ap = ecs.get(attacker, "dm.world.position")
    local al = ecs.get(attacker, "dm.world.location")
    local profile = ecs.get(attacker, "dm.combat.melee_profile")
    local best, best_distance, best_id = nil, math.huge, ""
    for _, candidate in ipairs(candidates) do
        if candidate:id() ~= attacker:id() then
            local selectable = ecs.get(candidate, "dm.world.selectable")
            local id, kind = selectable:get("id"), selectable:get("kind")
            if (wanted ~= "" and id == wanted) or (wanted == "" and kind == "hostile") then
                local cp, cl = ecs.get(candidate, "dm.world.position"), ecs.get(candidate, "dm.world.location")
                local dx, dy = cp:get("x")-ap:get("x"), cp:get("y")-ap:get("y")
                local distance = math.sqrt(dx*dx+dy*dy)
                if cl:get("act") == al:get("act") and cl:get("level_id") == al:get("level_id")
                    and distance <= profile:get("range") + selectable:get("radius")
                    and (distance < best_distance or distance == best_distance and id < best_id) then
                    best, best_distance, best_id = candidate, distance, id
                end
            end
        end
    end
    return best, best_id
end

local function hurt(target, damage)
    local monster = ecs.get(target, "dm.monster.stats")
    if monster then
        local remaining = math.max(monster:get("health") - damage, 0)
        monster:set("health", remaining)
        return remaining
    end
    local player = assert(ecs.get(target, "dm.player.vitals"), "melee target has no health")
    local whole = math.floor(damage / 256)
    local remaining = math.max(player:get("health") - whole, 0)
    player:set("health", remaining)
    return remaining * 256
end

local function event(structural, values)
    structural:create({["d2legacy.combat.melee_event"] = values})
end

function M.register()
    ecs.system({id="d2legacy.combat.basic_melee", phase="combat",
        -- One broad deterministic query supplies both attackers and candidate
        -- targets. Systems may not perform a hidden second ECS query while a
        -- tick is running.
        query={any={"dm.combat.basic_attack_request","dm.world.selectable"}},
        read={"dm.combat.basic_attack_request","dm.world.selectable","dm.world.position",
            "dm.world.location","dm.combat.melee_profile","dm.monster.stats","dm.player.vitals"},
        write={"dm.combat.basic_attack_request","dm.monster.stats","dm.player.vitals",
            "d2legacy.combat.melee_event"},
        update=function(context, entities, structural)
            for _, attacker in ipairs(entities) do
                local request = ecs.get(attacker, "dm.combat.basic_attack_request")
                local profile = ecs.get(attacker, "dm.combat.melee_profile")
                if request and profile then
                    assert(request:get("request_tick") <= context.tick, "future melee request")
                    local target, target_id = target_for(attacker, request:get("target_id"), entities)
                    structural:remove(attacker, "dm.combat.basic_attack_request")
                    local base = {kind="hit_resolved",tick=context.tick,attacker_id=identity(attacker),
                        target_id=target_id,hit=false,damage_raw=0,remaining_health_raw=0}
                    if target and policy.hits() then
                        local damage = policy.damage(profile:get("physical_min"), profile:get("physical_max"))
                        base.hit, base.damage_raw = true, damage
                        base.remaining_health_raw = hurt(target, damage)
                    end
                    event(structural, base)
                end
            end
        end})
end

return M
