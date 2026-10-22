-- Observe typed combat results and admit a poison rate to timed-state policy.
local ecs = require("engine.ecs/v1")
local M = {}

local function living(target)
    local monster = ecs.get(target, "d2legacy.monster.stats")
    if monster then
        return monster:get("health") > 0
    end
    local player = ecs.get(target, "d2legacy.player.vitals")
    return player and player:get("health") > 0
end

function M.register()
    ecs.system({
        id = "d2legacy.combat.apply_poison",
        phase = "effects",
        after = { "d2legacy.state.react_to_combat" },
        query = {
            any = { "d2legacy.combat.event", "d2legacy.world.selectable" },
            none = { "d2legacy.world.inactive" },
        },
        read = {
            "d2legacy.combat.event",
            "d2legacy.combat.damage_bundle",
            "d2legacy.combat.poison_observed",
            "d2legacy.world.selectable",
            "d2legacy.player.vitals",
            "d2legacy.monster.stats",
        },
        write = { "d2legacy.combat.poison_observed", "d2legacy.state.request" },
        update = function(_, entities, structural)
            local targets = {}
            for _, entity in ipairs(entities) do
                local selectable = ecs.get(entity, "d2legacy.world.selectable")
                if selectable then
                    targets[selectable:get("id")] = entity
                end
            end
            for _, entity in ipairs(entities) do
                local event = ecs.get(entity, "d2legacy.combat.event")
                local bundle = ecs.get(entity, "d2legacy.combat.damage_bundle")
                if event and bundle and not ecs.get(entity, "d2legacy.combat.poison_observed") then
                    structural:set(entity, "d2legacy.combat.poison_observed", {})
                    local rate = bundle:get("poison_mitigated_raw")
                    local length = bundle:get("poison_length_frames")
                    local target = targets[event:get("target_id")]
                    if
                        event:get("kind") == "damage_applied"
                        and rate > 0
                        and length > 0
                        and target
                        and living(target)
                    then
                        structural:create({
                            ["d2legacy.state.request"] = {
                                operation = "apply",
                                target = target,
                                state_id = "poison",
                                source_id = "poison:event:" .. entity:id(),
                                source_attacker_id = event:get("attacker_id"),
                                duration = length,
                                policy = "refresh_same_source",
                                stat = "hpregen",
                                stat_operation = "add",
                                stat_value = -rate,
                                stat_order = 300,
                                exclusive_group = "poison",
                                replacement_priority = rate,
                                reject_lower_priority = true,
                            },
                        })
                    end
                end
            end
        end,
    })
end

return M
