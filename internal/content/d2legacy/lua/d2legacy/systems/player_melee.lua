-- Own the complete point-and-click Attack action.
--
-- Repeated intent for the same skill and target preserves progress. A new
-- target replaces the old action. Targetless Shift-Attack swings in place and
-- lets melee resolution choose an enemy already within weapon reach.

local ecs = require("engine.ecs/v1")
local direction = require("d2legacy.policy.direction")
local melee = require("d2legacy.policy.melee")
local M = {}

local function controlled(entities, player)
    for _, entity in ipairs(entities) do
        local control = ecs.get(entity, "d2legacy.world.player_control")
        if control and control:get("player") == player then return entity end
    end
end

local function selected(entities, wanted)
    for _, entity in ipairs(entities) do
        local selectable = ecs.get(entity, "d2legacy.world.selectable")
        if selectable and selectable:get("id") == wanted then return entity end
    end
end

local function stop(entity, mode)
    local velocity, animation = ecs.get(entity, "d2legacy.world.velocity"), ecs.get(entity, "d2legacy.player.animation")
    if velocity then velocity:set("x", 0); velocity:set("y", 0) end
    if animation then animation:set("mode", mode or "NU") end
end

local function start_swing(context, attacker, target_id, dx, dy, structural)
    stop(attacker, "A1")
    local facing = ecs.get(attacker, "d2legacy.world.facing")
    if facing and (dx ~= 0 or dy ~= 0) then
        facing:set("direction", direction.quantize(dx, dy, facing:get("directions")))
    end
    -- These reviewed fallback ticks keep simulation independent of renderer
    -- frames. Typed AnimData timing will replace the defaults through mod data.
    structural:set(attacker, "d2legacy.combat.attack_animation", {
        skill_id=0,target_id=target_id,start_tick=context.tick,
        impact_tick=context.tick+3,complete_tick=context.tick+8,impact_fired=false,
    })
end

function M.register()
    ecs.system({id="d2legacy.combat.accept_player_melee",phase="pre_simulation",
        query={any={"d2legacy.skill.cast_event","d2legacy.world.player_control"}},
        read={"d2legacy.skill.cast_event","d2legacy.world.player_control",
            "d2legacy.combat.attack_approach","d2legacy.combat.attack_animation"},
        write={"d2legacy.combat.attack_approach","d2legacy.combat.attack_animation"},
        update=function(_, entities, structural)
            for _, event_entity in ipairs(entities) do
                local event = ecs.get(event_entity, "d2legacy.skill.cast_event")
                if event and event:get("behavior") == "basic.melee" then
                    local player = controlled(entities, event:get("player"))
                    if player then
                        local target_id = event:get("target_id")
                        local approach = ecs.get(player, "d2legacy.combat.attack_approach")
                        local attack = ecs.get(player, "d2legacy.combat.attack_animation")
                        local same = (approach and approach:get("target_id") == target_id)
                            or (attack and attack:get("target_id") == target_id)
                        if not same then
                            structural:remove(player, "d2legacy.combat.attack_animation")
                            structural:set(player, "d2legacy.combat.attack_approach", {
                                skill_id=0,target_id=target_id,request_tick=event:get("tick"),
                                target_x=event:get("target_x"),target_y=event:get("target_y"),
                            })
                        end
                    end
                    structural:destroy(event_entity)
                end
            end
        end})

    ecs.system({id="d2legacy.combat.player_melee_approach",phase="pre_simulation",
        query={any={"d2legacy.combat.attack_approach","d2legacy.world.selectable"}},
        read={"d2legacy.combat.attack_approach","d2legacy.world.selectable","d2legacy.world.position",
            "d2legacy.world.location","d2legacy.world.collider","d2legacy.combat.melee_profile"},
        write={"d2legacy.combat.attack_approach","d2legacy.combat.attack_animation",
            "d2legacy.world.velocity","d2legacy.world.facing","d2legacy.player.animation"},
        update=function(context, entities, structural)
            for _, attacker in ipairs(entities) do
                local approach, profile = ecs.get(attacker, "d2legacy.combat.attack_approach"), ecs.get(attacker, "d2legacy.combat.melee_profile")
                if approach and profile then
                    local target_id = approach:get("target_id")
                    local target = target_id ~= "" and selected(entities, target_id) or nil
                    local position = ecs.get(attacker, "d2legacy.world.position")
                    if target_id == "" then
                        structural:remove(attacker, "d2legacy.combat.attack_approach")
                        start_swing(context, attacker, "", approach:get("target_x")-position:get("x"), approach:get("target_y")-position:get("y"), structural)
                    elseif not target then
                        stop(attacker); structural:remove(attacker, "d2legacy.combat.attack_approach")
                    else
                        local target_position = ecs.get(target, "d2legacy.world.position")
                        local dx, dy = target_position:get("x")-position:get("x"), target_position:get("y")-position:get("y")
                        local length = math.sqrt(dx*dx+dy*dy)
                        local range = melee.reach(
                            profile:get("range"),
                            ecs.get(attacker, "d2legacy.world.collider"):get("radius"),
                            ecs.get(target, "d2legacy.world.collider"):get("radius"))
                        if length <= range then
                            structural:remove(attacker, "d2legacy.combat.attack_approach")
                            start_swing(context, attacker, target_id, dx, dy, structural)
                        elseif length > 0 then
                            local velocity = ecs.get(attacker, "d2legacy.world.velocity")
                            velocity:set("x", dx/length*10); velocity:set("y", dy/length*10)
                            local animation = ecs.get(attacker, "d2legacy.player.animation")
                            if animation then animation:set("mode", "WL") end
                        end
                    end
                end
            end
        end})

    ecs.system({id="d2legacy.combat.player_melee_animation",phase="pre_simulation",
        query={all={"d2legacy.combat.attack_animation"}},
        read={"d2legacy.combat.attack_animation","d2legacy.world.velocity",
            "d2legacy.player.animation"},
        write={"d2legacy.combat.attack_animation","d2legacy.combat.basic_attack_request",
            "d2legacy.world.velocity","d2legacy.player.animation"},
        update=function(context, entities, structural)
            for _, attacker in ipairs(entities) do
                local attack = ecs.get(attacker, "d2legacy.combat.attack_animation")
                if not attack:get("impact_fired") and context.tick >= attack:get("impact_tick") then
                    attack:set("impact_fired", true)
                    structural:set(attacker, "d2legacy.combat.basic_attack_request", {target_id=attack:get("target_id"),request_tick=context.tick})
                end
                if context.tick >= attack:get("complete_tick") then
                    stop(attacker); structural:remove(attacker, "d2legacy.combat.attack_animation")
                end
            end
        end})
end

return M
