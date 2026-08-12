-- Choose a nearby player, chase it, and request melee attacks.
--
-- This is Diablo policy, so it belongs here rather than in the generic ECS.
-- Collision-aware movement remains an engine mechanism: this system only sets
-- desired velocity, and the movement phase validates the resulting position.

local ecs = require("engine.ecs/v1")
local melee = require("d2legacy.policy.melee")
local M = {}

local function distance(a, b)
    local dx, dy = b:get("x") - a:get("x"), b:get("y") - a:get("y")
    return math.sqrt(dx * dx + dy * dy), dx, dy
end

local function player_targets(entities)
    local result = {}
    for _, entity in ipairs(entities) do
        local selectable = ecs.get(entity, "d2legacy.world.selectable")
        if selectable and selectable:get("kind") == "player" then
            table.insert(result, entity)
        end
    end
    table.sort(result, function(a, b)
        return ecs.get(a, "d2legacy.world.selectable"):get("id")
            < ecs.get(b, "d2legacy.world.selectable"):get("id")
    end)
    return result
end

local function same_zone(a, b)
    local left, right = ecs.get(a, "d2legacy.world.location"), ecs.get(b, "d2legacy.world.location")
    return left and right and left:get("act") == right:get("act")
        and left:get("level_id") == right:get("level_id")
end

local function choose(monster, brain, targets)
    local origin, remembered = ecs.get(monster, "d2legacy.world.position"), brain:get("target_id")
    local best, best_distance = nil, math.huge
    for _, target in ipairs(targets) do
        if same_zone(monster, target) then
            local current = distance(origin, ecs.get(target, "d2legacy.world.position"))
            local id = ecs.get(target, "d2legacy.world.selectable"):get("id")
            if current <= brain:get("aggro_radius") and (id == remembered or current < best_distance) then
                best, best_distance = target, current
                if id == remembered then return best end
            end
        end
    end
    return best
end

local function stop(velocity)
    velocity:set("x", 0)
    velocity:set("y", 0)
end

function M.register()
    ecs.system({id="d2legacy.monster.basic_ai",phase="intent",
        query={any={"d2legacy.monster.ai","d2legacy.world.selectable"}},
        read={"d2legacy.world.selectable","d2legacy.world.position","d2legacy.world.location",
            "d2legacy.world.collider","d2legacy.monster.ai"},
        write={"d2legacy.monster.ai","d2legacy.world.velocity","d2legacy.combat.basic_attack_request"},
        update=function(context, entities, structural)
            local targets = player_targets(entities)
            for _, monster in ipairs(entities) do
                local brain = ecs.get(monster, "d2legacy.monster.ai")
                local velocity = ecs.get(monster, "d2legacy.world.velocity")
                if brain and velocity and brain:get("next_think_tick") <= context.tick then
                    brain:set("next_think_tick", context.tick + brain:get("think_interval"))
                    local target = choose(monster, brain, targets)
                    if not target then
                        brain:set("state", "idle"); brain:set("target_id", ""); stop(velocity)
                        structural:remove(monster, "d2legacy.combat.basic_attack_request")
                    else
                        local selected = ecs.get(target, "d2legacy.world.selectable")
                        local id = selected:get("id")
                        local range = melee.reach(
                            brain:get("attack_range"),
                            ecs.get(monster, "d2legacy.world.collider"):get("radius"),
                            ecs.get(target, "d2legacy.world.collider"):get("radius"))
                        local length, dx, dy = distance(ecs.get(monster, "d2legacy.world.position"), ecs.get(target, "d2legacy.world.position"))
                        brain:set("target_id", id)
                        if length <= range then
                            brain:set("state", "attack"); stop(velocity)
                            structural:set(monster, "d2legacy.combat.basic_attack_request", {target_id=id,request_tick=context.tick})
                        else
                            brain:set("state", "chase")
                            local speed = brain:get("speed")
                            velocity:set("x", dx / length * speed); velocity:set("y", dy / length * speed)
                            structural:remove(monster, "d2legacy.combat.basic_attack_request")
                        end
                    end
                end
            end
        end})
end

return M
