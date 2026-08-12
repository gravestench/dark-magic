-- Resolve impact requests emitted by reusable approach/animation mechanisms.
--
-- Movement and animation merely say that an impact moment happened. This
-- authoritative system chooses an in-range target, rolls Diablo's hit and
-- damage formulas, mutates health, and emits a factual result event. Keeping
-- that distinction lets presentation animate without becoming game authority.

local ecs = require("engine.ecs/v1")
local policy = require("d2legacy.policy.melee")
local mitigation = require("d2legacy.policy.mitigation")
local M = {}

local function identity(entity)
    local selectable = ecs.get(entity, "d2legacy.world.selectable")
    return selectable and selectable:get("id") or "entity:" .. entity:id()
end

local function target_for(attacker, wanted, candidates)
    local ap = ecs.get(attacker, "d2legacy.world.position")
    local al = ecs.get(attacker, "d2legacy.world.location")
    local ac = ecs.get(attacker, "d2legacy.world.collider")
    local profile = ecs.get(attacker, "d2legacy.combat.melee_profile")
    local best, best_distance, best_id = nil, math.huge, ""
    -- A named target wins when it remains valid. Shift-attacking supplies no
    -- target, so the nearest hostile inside melee reach is selected instead.
    for _, candidate in ipairs(candidates) do
        if candidate:id() ~= attacker:id() then
            local selectable = ecs.get(candidate, "d2legacy.world.selectable")
            local collider = ecs.get(candidate, "d2legacy.world.collider")
            local id, kind = selectable:get("id"), selectable:get("kind")
            if (wanted ~= "" and id == wanted) or (wanted == "" and kind == "hostile") then
                local cp, cl = ecs.get(candidate, "d2legacy.world.position"), ecs.get(candidate, "d2legacy.world.location")
                local dx, dy = cp:get("x")-ap:get("x"), cp:get("y")-ap:get("y")
                local distance = math.sqrt(dx*dx+dy*dy)
                if cl:get("act") == al:get("act") and cl:get("level_id") == al:get("level_id")
                    and distance <= policy.reach(
                        profile:get("range"), ac:get("radius"), collider:get("radius"))
                    and (distance < best_distance or distance == best_distance and id < best_id) then
                    best, best_distance, best_id = candidate, distance, id
                end
            end
        end
    end
    return best, best_id
end

local function hurt(target, damage)
    damage = mitigation.apply(damage, "physical", ecs.get(target, "d2legacy.combat.defense"))
    local monster = ecs.get(target, "d2legacy.monster.stats")
    if monster then
        local remaining = math.max(monster:get("health") - damage, 0)
        monster:set("health", remaining)
        return remaining, damage
    end
    local player = assert(ecs.get(target, "d2legacy.player.vitals"), "melee target has no health")
    local whole = math.floor(damage / 256)
    local remaining = math.max(player:get("health") - whole, 0)
    player:set("health", remaining)
    return remaining * 256, damage
end

local function combat_values(entity)
    local monster = ecs.get(entity, "d2legacy.monster.stats")
    if monster then return monster:get("level"), monster:get("attack_rating"), monster:get("defense") end
    local progress = assert(ecs.get(entity,"d2legacy.player.progress"), "player has no progress")
    local stats = assert(ecs.get(entity,"d2legacy.player.combat_stats"), "player has no combat stats")
    return progress:get("level"), stats:get("attack_rating"), stats:get("defense")
end

local function event(structural, values)
    structural:create({["d2legacy.combat.melee_event"] = values})
end

function M.register()
    ecs.system({id="d2legacy.combat.basic_melee", phase="combat",
        -- One broad deterministic query supplies both attackers and candidate
        -- targets. Systems may not perform a hidden second ECS query while a
        -- tick is running.
        query={any={"d2legacy.combat.basic_attack_request","d2legacy.world.selectable"}},
        read={"d2legacy.combat.basic_attack_request","d2legacy.world.selectable","d2legacy.world.position",
            "d2legacy.world.location","d2legacy.world.collider","d2legacy.combat.melee_profile","d2legacy.monster.stats","d2legacy.player.vitals",
            "d2legacy.player.progress","d2legacy.player.combat_stats","d2legacy.combat.defense"},
        write={"d2legacy.combat.basic_attack_request","d2legacy.monster.stats","d2legacy.player.vitals",
            "d2legacy.combat.melee_event"},
        update=function(context, entities, structural)
            for _, attacker in ipairs(entities) do
                local request = ecs.get(attacker, "d2legacy.combat.basic_attack_request")
                local profile = ecs.get(attacker, "d2legacy.combat.melee_profile")
                if request and profile then
                    assert(request:get("request_tick") <= context.tick, "future melee request")
                    local target, target_id = target_for(attacker, request:get("target_id"), entities)
                    structural:remove(attacker, "d2legacy.combat.basic_attack_request")
                    local base = {kind="hit_resolved",tick=context.tick,attacker_id=identity(attacker),
                        target_id=target_id,hit=false,damage_raw=0,remaining_health_raw=0}
                    if target then
                        local attacker_level, attack_rating = combat_values(attacker)
                        local defender_level, _, defense = combat_values(target)
                        if policy.hits(attacker_level, defender_level, attack_rating, defense) then
                            local damage = policy.damage(
                                profile:get("physical_min"),
                                profile:get("physical_max")
                            )
                            base.hit = true
                            base.remaining_health_raw, base.damage_raw = hurt(target, damage)
                        end
                    end
                    event(structural, base)
                end
            end
        end})
end

return M
