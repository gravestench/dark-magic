-- Turn one admitted skill request into a definition-driven timed cast.
--
-- Mana is stored in 8.8 fixed-point units: 256 means one visible mana point.
-- The cost is paid exactly once when a cast starts. An underfunded request is
-- rejected before the action exists and preserves the available mana. The
-- system then remembers effect and completion ticks so replay does not depend
-- on animation frames.

local ecs = require("engine.ecs/v1")
local animdata = require("engine.animdata/v1")
local direction = require("d2legacy.policy.direction")
local player_motion = require("d2legacy.gameplay.player_motion")
local skill_progression = require("d2legacy.policy.skill_progression")

local M = {}

local function learned_levels(entities)
    local levels = {}
    for _, entity in ipairs(entities) do
        local learned = ecs.get(entity, "d2legacy.player.learned_skill")
        if learned then
            local owner = learned:get("owner"):id()
            levels[owner] = levels[owner] or {}
            levels[owner][learned:get("skill_id")] = learned:get("level")
        end
    end
    return levels
end

local function elemental_damage_percent(definition, levels)
    local total_levels = 0
    for _, skill_id in ipairs(definition.damage_synergy_skill_ids or {}) do
        total_levels = total_levels + (levels[skill_id] or 0)
    end
    return total_levels * (definition.damage_synergy_percent_per_level or 0)
end

local function effect_duration(definition, level, levels)
    local base =
        skill_progression.linear(definition.effect_duration_base or 0, definition.effect_duration_per_level or 0, level)
    local synergy_levels = 0
    for _, skill_id in ipairs(definition.effect_duration_synergy_skill_ids or {}) do
        synergy_levels = synergy_levels + (levels[skill_id] or 0)
    end
    local multiplier = 100 + synergy_levels * (definition.effect_duration_synergy_percent_per_level or 0)
    return math.floor(base * multiplier / 100)
end

local function cast_timing(player, definition, action)
    if
        definition.behavior == "action.melee"
        or not action
        or action.animation_mode == ""
        or action.sequence_transition == ""
    then
        return definition.effect_delay, definition.complete_delay, ""
    end
    assert(action.animation_mode ~= "SQ", "sequence cast timing is not implemented")
    local appearance = assert(ecs.get(player, "d2legacy.player.appearance"), "cast actor has no appearance")
    local key = string.upper(appearance:get("token") .. action.animation_mode .. appearance:get("weapon_class"))
    local timing, err = animdata.record(key)
    assert(timing, err or ("missing authoritative AnimData record " .. key))
    local effect_delay
    for _, event in ipairs(timing.events or {}) do
        if event.kind ~= "sound" and (not effect_delay or event.delay < effect_delay) then
            effect_delay = event.delay
        end
    end
    assert(effect_delay, "AnimData record " .. key .. " has no cast action event")
    assert(effect_delay < timing.complete_delay, "cast action event must precede completion for " .. key)
    return effect_delay, timing.complete_delay, action.animation_mode
end

local function begin_action(context, player, request, mode, commands)
    if mode == "" then
        return
    end
    player_motion.stop(player)
    local velocity = ecs.get(player, "d2legacy.world.velocity")
    if velocity then
        velocity:set("x", 0)
        velocity:set("y", 0)
    end
    local movement_mode = ecs.get(player, "d2legacy.player.movement_mode")
    if movement_mode then
        movement_mode:set("running", false)
    end
    local facing = ecs.get(player, "d2legacy.world.facing")
    local position = ecs.get(player, "d2legacy.world.position")
    if facing and position then
        local dx = request:get("target_x") - position:get("x")
        local dy = request:get("target_y") - position:get("y")
        if dx ~= 0 or dy ~= 0 then
            facing:set("direction", direction.quantize(dx, dy, facing:get("directions")))
        end
    end
    local animation = ecs.get(player, "d2legacy.player.animation")
    if animation then
        animation:set("mode", mode)
        animation:set("start_tick", context.tick)
    end
    commands:set(player, "d2legacy.skill.cast_action", {})
end

local function finish_action(context, player, commands)
    if not ecs.get(player, "d2legacy.skill.cast_action") then
        return
    end
    local animation = ecs.get(player, "d2legacy.player.animation")
    if animation then
        animation:set("mode", "NU")
        animation:set("start_tick", context.tick)
    end
    commands:remove(player, "d2legacy.skill.cast_action")
end

local function begin_cast(context, player, request, definitions, actions, levels, commands)
    local vitals = ecs.get(player, "d2legacy.player.vitals")
    local available = vitals:get("mana_raw")
    if available == 0 then
        available = vitals:get("mana") * 256
    end

    local player_levels = levels[player:id()] or {}
    local known_level = player_levels[request:get("skill_id")] or 0
    local definition = definitions[request:get("skill_id")]
    local valid = request:get("request_tick") <= context.tick
        and definition ~= nil
        and known_level > 0
        and available >= skill_progression.mana_cost(definition, known_level)

    if valid then
        local effect_delay, complete_delay, animation_mode =
            cast_timing(player, definition, actions[request:get("skill_id")])
        local remaining = available - skill_progression.mana_cost(definition, known_level)
        vitals:set("mana_raw", remaining)
        vitals:set("mana", math.floor(remaining / 256))
        commands:set(player, "d2legacy.skill.cast", {
            skill_id = request:get("skill_id"),
            skill_level = known_level,
            target_x = request:get("target_x"),
            target_y = request:get("target_y"),
            target_id = request:get("target_id"),
            effect_tick = context.tick + effect_delay,
            complete_tick = context.tick + complete_delay,
            elemental_damage_percent = elemental_damage_percent(definition, player_levels),
            effect_duration_ticks = effect_duration(definition, known_level, player_levels),
            effect_emitted = false,
            effect_cue_emitted = false,
        })
        commands:create({
            ["d2legacy.skill.cast_cue"] = {
                kind = "cast_started",
                tick = context.tick,
                effect_tick = context.tick + effect_delay,
                caster = player,
                player = request:get("player"),
                skill_id = request:get("skill_id"),
                target_x = request:get("target_x"),
                target_y = request:get("target_y"),
                target_id = request:get("target_id"),
            },
        })
        begin_action(context, player, request, animation_mode, commands)
    end
    commands:remove(player, "d2legacy.skill.cast_request")
end

function M.register(definitions, actions)
    actions = actions or {}
    ecs.system({
        id = "d2legacy.skill.cast_lifecycle",
        phase = "pre_simulation",
        query = {
            any = {
                "d2legacy.skill.cast_request",
                "d2legacy.skill.cast",
                "d2legacy.player.learned_skill",
            },
            none = { "d2legacy.player.death" },
        },
        read = {
            "d2legacy.skill.cast_request",
            "d2legacy.skill.cast",
            "d2legacy.player.learned_skill",
            "d2legacy.player.vitals",
            "d2legacy.player.death",
            "d2legacy.player.appearance",
            "d2legacy.player.motion",
            "d2legacy.world.velocity",
            "d2legacy.player.movement_mode",
            "d2legacy.player.animation",
            "d2legacy.world.facing",
            "d2legacy.world.position",
            "d2legacy.skill.cast_action",
            "d2legacy.skill.cast_cue",
        },
        write = {
            "d2legacy.skill.cast_request",
            "d2legacy.skill.cast",
            "d2legacy.player.vitals",
            "d2legacy.player.motion",
            "d2legacy.world.velocity",
            "d2legacy.player.movement_mode",
            "d2legacy.player.animation",
            "d2legacy.world.facing",
            "d2legacy.skill.cast_action",
            "d2legacy.skill.cast_cue",
        },
        update = function(context, entities, structural)
            local levels = learned_levels(entities)
            for _, player in ipairs(entities) do
                local request = ecs.get(player, "d2legacy.skill.cast_request")
                local cast = ecs.get(player, "d2legacy.skill.cast")
                if request and not cast then
                    begin_cast(context, player, request, definitions, actions, levels, structural)
                elseif cast and context.tick >= cast:get("complete_tick") then
                    finish_action(context, player, structural)
                    structural:remove(player, "d2legacy.skill.cast")
                end
            end
        end,
    })

    ecs.system({
        id = "d2legacy.skill.cast_effect_cue",
        phase = "effects",
        query = {
            all = { "d2legacy.skill.cast", "d2legacy.player.identity" },
            none = { "d2legacy.player.death" },
        },
        read = { "d2legacy.skill.cast", "d2legacy.player.identity", "d2legacy.player.death" },
        write = { "d2legacy.skill.cast", "d2legacy.skill.cast_cue" },
        update = function(context, entities, structural)
            for _, player in ipairs(entities) do
                local cast = ecs.get(player, "d2legacy.skill.cast")
                if cast:get("effect_emitted") and not cast:get("effect_cue_emitted") then
                    cast:set("effect_cue_emitted", true)
                    local identity = ecs.get(player, "d2legacy.player.identity")
                    structural:create({
                        ["d2legacy.skill.cast_cue"] = {
                            kind = "cast_effect",
                            tick = context.tick,
                            effect_tick = cast:get("effect_tick"),
                            caster = player,
                            player = identity:get("player"),
                            skill_id = cast:get("skill_id"),
                            target_x = cast:get("target_x"),
                            target_y = cast:get("target_y"),
                            target_id = cast:get("target_id"),
                        },
                    })
                end
            end
        end,
    })
end

return M
