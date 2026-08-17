-- Materialize friendly-target timed multi-stat effects from admitted casts.

local ecs = require("engine.ecs/v1")
local friendly_target = require("d2legacy.gameplay.friendly_target")
local progression = require("d2legacy.policy.skill_progression")
local M = {}

local function learned_levels(entities)
    local result = {}
    for _, entity in ipairs(entities) do
        local learned = ecs.get(entity, "d2legacy.player.learned_skill")
        if learned then
            local owner = learned:get("owner"):id()
            result[owner] = result[owner] or {}
            result[owner][learned:get("skill_id")] = learned:get("level")
        end
    end
    return result
end

local function damage_percent(definition, levels)
    local total = 0
    for _, skill_id in ipairs(definition.damage_synergy_skill_ids) do
        total = total + (levels[skill_id] or 0)
    end
    return total * definition.damage_synergy_percent_per_level
end

local function target_id(target)
    local selectable = ecs.get(target, "d2legacy.world.selectable")
    return selectable and selectable:get("id") or ("entity:" .. target:id())
end

local function emit_source(structural, target, owner_source_id, stat, operation, value)
    structural:create({
        ["d2legacy.state.stat_request"] = {
            target = target,
            owner_source_id = owner_source_id,
            source_id = owner_source_id .. ":" .. stat,
            stat = stat,
            operation = operation,
            value = value,
            order = 300,
        },
    })
end

function M.register(definitions)
    ecs.system({
        id = "d2legacy.state.apply_targeted_skill",
        phase = "pre_simulation",
        after = { "d2legacy.skill.cast_lifecycle" },
        query = {
            any = {
                "d2legacy.skill.cast",
                "d2legacy.player.learned_skill",
                "d2legacy.world.selectable",
            },
        },
        read = {
            "d2legacy.skill.cast",
            "d2legacy.player.learned_skill",
            "d2legacy.player.identity",
            "d2legacy.player.vitals",
            "d2legacy.player.death",
            "d2legacy.monster.stats",
            "d2legacy.world.selectable",
            "d2legacy.world.location",
            "d2legacy.world.inactive",
        },
        write = {
            "d2legacy.skill.cast",
            "d2legacy.state.request",
            "d2legacy.state.stat_request",
            "d2legacy.skill.cast_event",
        },
        update = function(context, entities, structural)
            local levels_by_owner = learned_levels(entities)
            for _, caster in ipairs(entities) do
                local cast = ecs.get(caster, "d2legacy.skill.cast")
                local definition = cast and definitions[cast:get("skill_id")] or nil
                if
                    definition
                    and definition.behavior == "state.targeted-timed-stats"
                    and not cast:get("effect_emitted")
                    and context.tick >= cast:get("effect_tick")
                then
                    local identity = assert(ecs.get(caster, "d2legacy.player.identity"))
                    local target = assert(friendly_target.named(caster, cast:get("target_id"), entities, definition))
                    local level = cast:get("skill_level")
                    local levels = levels_by_owner[caster:id()] or {}
                    local minimum, maximum =
                        progression.damage_range(definition, level, damage_percent(definition, levels))
                    local owner_source_id = "skill:" .. identity:get("player") .. ":" .. cast:get("skill_id")
                    structural:create({
                        ["d2legacy.state.request"] = {
                            operation = "apply",
                            target = target,
                            state_id = definition.state_id,
                            source_id = owner_source_id,
                            duration = progression.linear(
                                definition.duration_base,
                                definition.duration_per_level,
                                level
                            ),
                            policy = "refresh_same_source",
                        },
                    })
                    emit_source(structural, target, owner_source_id, "firemindam", "add", minimum)
                    emit_source(structural, target, owner_source_id, "firemaxdam", "add", maximum)
                    emit_source(
                        structural,
                        target,
                        owner_source_id,
                        "item_tohit_percent",
                        "percent",
                        progression.linear(
                            definition.attack_rating_percent_base,
                            definition.attack_rating_percent_per_level,
                            level
                        )
                    )
                    structural:create({
                        ["d2legacy.skill.cast_event"] = {
                            kind = "skill_effect",
                            tick = context.tick,
                            player = identity:get("player"),
                            skill_id = cast:get("skill_id"),
                            skill_level = level,
                            behavior = definition.behavior,
                            target_x = 0,
                            target_y = 0,
                            target_id = target_id(target),
                            reason = "",
                        },
                    })
                    cast:set("effect_emitted", true)
                end
            end
        end,
    })
end

return M
