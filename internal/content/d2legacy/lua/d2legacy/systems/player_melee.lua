-- Own the complete point-and-click Attack action.
--
-- Repeated intent for the same skill and target preserves progress. A new
-- target replaces the old action. Targetless Shift-Attack swings in place and
-- lets melee resolution choose an enemy already within weapon reach.

local ecs = require("engine.ecs/v1")
local animdata = require("engine.animdata/v1")
local action_rate = require("d2legacy.policy.action_rate")
local combat_target = require("d2legacy.gameplay.combat_target")
local player_motion = require("d2legacy.gameplay.player_motion")
local direction = require("d2legacy.policy.direction")
local weapon_selection = require("d2legacy.policy.weapon_selection")
local M = {}

local function controlled(entities, player)
    for _, entity in ipairs(entities) do
        local control = ecs.get(entity, "d2legacy.world.player_control")
        if control and control:get("player") == player then
            return entity
        end
    end
end

local function stop(entity, mode, tick)
    player_motion.stop(entity)
    local animation = ecs.get(entity, "d2legacy.player.animation")
    if animation then
        local next_mode = mode or "NU"
        if animation:get("mode") ~= next_mode then
            animation:set("mode", next_mode)
            animation:set("start_tick", tick)
        end
    end
end

local function identity(entity)
    local selectable = ecs.get(entity, "d2legacy.world.selectable")
    return selectable and selectable:get("id") or tostring(entity:id())
end

local function animation_event(structural, attacker, attack, kind, tick)
    local target_id, skill_id
    if type(attack) == "table" then
        target_id, skill_id = attack.target_id, attack.skill_id
    else
        target_id, skill_id = attack:get("target_id"), attack:get("skill_id")
    end
    structural:create({
        ["d2legacy.combat.attack_animation_event"] = {
            kind = kind,
            tick = tick,
            attacker_id = identity(attacker),
            target_id = target_id,
            skill_id = skill_id,
            hand = type(attack) == "table" and attack.hand or attack:get("hand"),
        },
    })
end

local function attack_hand(attacker, selector, sequence_frame)
    local profile = assert(ecs.get(attacker, "d2legacy.combat.melee_profile"), "player has no melee profile")
    return weapon_selection.select(
        selector,
        profile:get("primary_hand"),
        profile:get("dual_wield") and profile:get("secondary_hand") or "",
        sequence_frame
    )
end

local function action_timing(attacker, mode)
    local appearance = assert(ecs.get(attacker, "d2legacy.player.appearance"), "player has no composite appearance")
    local key = string.upper(appearance:get("token") .. mode .. appearance:get("weapon_class"))
    local timing, err = animdata.record(key)
    assert(timing, err or ("missing authoritative AnimData record " .. key))
    local rate = assert(ecs.get(attacker, "d2legacy.combat.action_rate"), "player has no action-rate facts")
    local profile = assert(ecs.get(attacker, "d2legacy.combat.melee_profile"), "player has no melee profile")
    local schedule = action_rate.schedule(timing, {
        attack_rate = rate:get("attack_rate"),
        item_fasterattackrate = rate:get("item_fasterattackrate"),
        primary_weapon_attack_rate = profile:get("primary_weapon_attack_rate"),
        secondary_weapon_attack_rate = profile:get("secondary_weapon_attack_rate"),
        dual_wield = profile:get("dual_wield"),
        sequence = false,
    })
    return schedule.attack_delay, schedule.complete_delay
end

local function start_swing(context, attacker, skill_id, target_id, dx, dy, selector, mode, structural)
    stop(attacker, mode, context.tick)
    local facing = ecs.get(attacker, "d2legacy.world.facing")
    if facing and (dx ~= 0 or dy ~= 0) then
        facing:set("direction", direction.quantize(dx, dy, facing:get("directions")))
    end
    local impact_delay, complete_delay = action_timing(attacker, mode)
    local attack = {
        skill_id = skill_id,
        target_id = target_id,
        start_tick = context.tick,
        impact_tick = context.tick + impact_delay,
        complete_tick = context.tick + complete_delay,
        impact_fired = false,
        weapon_selection = selector,
        hand = attack_hand(attacker, selector, context.tick),
        animation_mode = mode,
    }
    structural:set(attacker, "d2legacy.combat.attack_animation", attack)
    animation_event(structural, attacker, attack, "attack_started", context.tick)
end

function M.register()
    ecs.system({
        id = "d2legacy.combat.accept_player_melee",
        phase = "pre_simulation",
        after = { "d2legacy.skill.emit_melee_action" },
        query = { any = { "d2legacy.skill.cast_event", "d2legacy.world.player_control" } },
        read = {
            "d2legacy.skill.cast_event",
            "d2legacy.world.player_control",
            "d2legacy.combat.attack_approach",
            "d2legacy.combat.attack_animation",
        },
        write = { "d2legacy.combat.attack_approach", "d2legacy.combat.attack_animation" },
        update = function(_, entities, structural)
            for _, event_entity in ipairs(entities) do
                local event = ecs.get(event_entity, "d2legacy.skill.cast_event")
                if event and event:get("behavior") == "action.melee" then
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
                                skill_id = event:get("skill_id"),
                                target_id = target_id,
                                request_tick = event:get("tick"),
                                target_x = event:get("target_x"),
                                target_y = event:get("target_y"),
                                weapon_selection = event:get("weapon_selection"),
                                animation_mode = event:get("animation_mode"),
                            })
                        end
                    end
                    structural:destroy(event_entity)
                end
            end
        end,
    })

    ecs.system({
        id = "d2legacy.combat.player_melee_approach",
        phase = "pre_simulation",
        query = {
            any = { "d2legacy.combat.attack_approach", "d2legacy.world.selectable" },
            none = { "d2legacy.world.inactive" },
        },
        read = {
            "d2legacy.combat.attack_approach",
            "d2legacy.world.selectable",
            "d2legacy.world.position",
            "d2legacy.world.location",
            "d2legacy.world.collider",
            "d2legacy.combat.melee_profile",
            "d2legacy.monster.stats",
            "d2legacy.player.vitals",
            "d2legacy.player.appearance",
            "d2legacy.combat.action_rate",
            "d2legacy.player.motion",
        },
        write = {
            "d2legacy.combat.attack_approach",
            "d2legacy.combat.attack_animation",
            "d2legacy.player.motion",
            "d2legacy.world.facing",
            "d2legacy.player.animation",
            "d2legacy.combat.attack_animation_event",
            "d2legacy.combat.melee_profile",
        },
        update = function(context, entities, structural)
            for _, attacker in ipairs(entities) do
                local approach, profile =
                    ecs.get(attacker, "d2legacy.combat.attack_approach"),
                    ecs.get(attacker, "d2legacy.combat.melee_profile")
                if approach and profile then
                    local target_id = approach:get("target_id")
                    local target, facts = nil, nil
                    if target_id ~= "" then
                        target, facts = combat_target.named(attacker, target_id, entities)
                    end
                    local position = ecs.get(attacker, "d2legacy.world.position")
                    if target_id == "" then
                        structural:remove(attacker, "d2legacy.combat.attack_approach")
                        start_swing(
                            context,
                            attacker,
                            approach:get("skill_id"),
                            "",
                            approach:get("target_x") - position:get("x"),
                            approach:get("target_y") - position:get("y"),
                            approach:get("weapon_selection"),
                            approach:get("animation_mode"),
                            structural
                        )
                    elseif not target then
                        stop(attacker, nil, context.tick)
                        structural:remove(attacker, "d2legacy.combat.attack_approach")
                    else
                        local dx, dy, length = facts.dx, facts.dy, facts.distance
                        if combat_target.in_melee_range(facts, profile:get("range")) then
                            structural:remove(attacker, "d2legacy.combat.attack_approach")
                            start_swing(
                                context,
                                attacker,
                                approach:get("skill_id"),
                                target_id,
                                dx,
                                dy,
                                approach:get("weapon_selection"),
                                approach:get("animation_mode"),
                                structural
                            )
                        elseif combat_target.within_reach(facts, profile:get("range")) then
                            -- In footprint range but separated by a melee
                            -- barrier. The current direct-motion approach has
                            -- no route ownership, so reject instead of swinging
                            -- through it or leaving a permanently stuck action.
                            stop(attacker, nil, context.tick)
                            structural:remove(attacker, "d2legacy.combat.attack_approach")
                        elseif length > 0 then
                            player_motion.approach(attacker, facts.target_x, facts.target_y)
                        end
                    end
                end
            end
        end,
    })

    ecs.system({
        id = "d2legacy.combat.player_melee_animation",
        phase = "pre_simulation",
        query = { all = { "d2legacy.combat.attack_animation" } },
        read = {
            "d2legacy.combat.attack_animation",
            "d2legacy.player.motion",
            "d2legacy.player.animation",
            "d2legacy.world.selectable",
        },
        write = {
            "d2legacy.combat.attack_animation",
            "d2legacy.combat.basic_attack_request",
            "d2legacy.player.motion",
            "d2legacy.player.animation",
            "d2legacy.combat.attack_animation_event",
        },
        update = function(context, entities, structural)
            for _, attacker in ipairs(entities) do
                local attack = ecs.get(attacker, "d2legacy.combat.attack_animation")
                if not attack:get("impact_fired") and context.tick >= attack:get("impact_tick") then
                    attack:set("impact_fired", true)
                    animation_event(structural, attacker, attack, "attack_impact", context.tick)
                    structural:set(attacker, "d2legacy.combat.basic_attack_request", {
                        target_id = attack:get("target_id"),
                        request_tick = context.tick,
                        hand = attack:get("hand"),
                    })
                end
                if context.tick >= attack:get("complete_tick") then
                    animation_event(structural, attacker, attack, "attack_completed", context.tick)
                    stop(attacker, nil, context.tick)
                    structural:remove(attacker, "d2legacy.combat.attack_animation")
                end
            end
        end,
    })
end

return M
