-- Commit the first common player-death state without inventing consequences.
--
-- The durable player remains the same ECS entity. This system only freezes
-- live actions, records immediate/ultimate attribution, and emits a semantic
-- event. Corpse equipment, gold, experience, respawn, and permanent Hardcore
-- transitions remain independent consumers gated on owned 1.14d evidence.

local ecs = require("engine.ecs/v1")
local attribution = require("d2legacy.owned_units.attribution")
local player_motion = require("d2legacy.gameplay.player_motion")
local game_rules = require("d2legacy.policy.game_rules")
local M = {}

local function selectable_by_id(entities, wanted)
    for _, entity in ipairs(entities) do
        local selected = ecs.get(entity, "d2legacy.world.selectable")
        if selected and selected:get("id") == wanted then
            return entity
        end
    end
    return nil
end

local function collect_lethal_results(entities, structural)
    local results = {}
    local seen = {}
    for _, entity in ipairs(entities) do
        local event = ecs.get(entity, "d2legacy.combat.event")
        if event then
            structural:set(entity, "d2legacy.combat.player_death_observed", {})
            local target_id = event:get("target_id")
            if event:get("kind") == "unit_died" and not seen[target_id] then
                seen[target_id] = true
                results[#results + 1] = { target_id = target_id, killer_id = event:get("attacker_id") }
            end
        end
    end
    return results
end

local function stop_actions(player, context, structural)
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
    local animation = ecs.get(player, "d2legacy.player.animation")
    if animation then
        animation:set("mode", "DT")
        animation:set("start_tick", context.tick)
    end

    for _, component in ipairs({
        "d2legacy.skill.cast_request",
        "d2legacy.skill.cast",
        "d2legacy.skill.cast_action",
        "d2legacy.combat.attack_approach",
        "d2legacy.combat.attack_animation",
        "d2legacy.combat.basic_attack_request",
        "d2legacy.world.forced_motion_request",
        "d2legacy.world.forced_motion",
    }) do
        structural:remove(player, component)
    end
end

local function commit(context, entities, structural, player, killer_id)
    local identity = ecs.get(player, "d2legacy.player.identity")
    local vitals = ecs.get(player, "d2legacy.player.vitals")
    if not identity or not vitals or vitals:get("health") > 0 or ecs.get(player, "d2legacy.player.death") then
        return
    end

    local credited_id = attribution.resolve(entities, killer_id).ultimate_owner_id
    local hardcore = game_rules.get().hardcore
    local values = {
        tick = context.tick,
        killer_id = killer_id,
        credited_id = credited_id,
        hardcore = hardcore,
        stage = "dead",
        consequences_pending = true,
    }
    structural:set(player, "d2legacy.player.death", values)
    stop_actions(player, context, structural)
    structural:create({
        ["d2legacy.player.death_event"] = {
            kind = "player_died",
            tick = context.tick,
            player_id = "player:" .. identity:get("player"),
            killer_id = killer_id,
            credited_id = credited_id,
            hardcore = hardcore,
            consequences_pending = true,
        },
    })
end

local function update(context, entities, structural)
    local lethal_results = collect_lethal_results(entities, structural)
    for _, result in ipairs(lethal_results) do
        local player = selectable_by_id(entities, result.target_id)
        if player and ecs.get(player, "d2legacy.player.identity") then
            commit(context, entities, structural, player, result.killer_id)
        end
    end
end

function M.register()
    ecs.system({
        id = "d2legacy.player.death",
        phase = "effects",
        after = { "d2legacy.monster.death" },
        query = {
            any = {
                "d2legacy.combat.event",
                "d2legacy.player.identity",
                "d2legacy.world.selectable",
                "d2legacy.owned_unit",
            },
            none = { "d2legacy.world.inactive", "d2legacy.combat.player_death_observed" },
        },
        read = {
            "d2legacy.combat.event",
            "d2legacy.combat.player_death_observed",
            "d2legacy.player.identity",
            "d2legacy.player.vitals",
            "d2legacy.player.death",
            "d2legacy.world.selectable",
            "d2legacy.owned_unit",
            "d2legacy.player.motion",
            "d2legacy.world.velocity",
            "d2legacy.player.movement_mode",
            "d2legacy.player.animation",
        },
        write = {
            "d2legacy.combat.player_death_observed",
            "d2legacy.player.death",
            "d2legacy.player.death_event",
            "d2legacy.player.motion",
            "d2legacy.world.velocity",
            "d2legacy.player.movement_mode",
            "d2legacy.player.animation",
            "d2legacy.skill.cast_request",
            "d2legacy.skill.cast",
            "d2legacy.skill.cast_action",
            "d2legacy.combat.attack_approach",
            "d2legacy.combat.attack_animation",
            "d2legacy.combat.basic_attack_request",
            "d2legacy.world.forced_motion_request",
            "d2legacy.world.forced_motion",
        },
        update = update,
    })
end

return M
