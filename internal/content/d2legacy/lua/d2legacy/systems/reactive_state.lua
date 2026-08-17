-- Convert factual melee hits into state reactions declared by active states.

local ecs = require("engine.ecs/v1")
local cold_duration = require("d2legacy.policy.cold_duration")
local game_rules = require("d2legacy.policy.game_rules")
local M = {}

local function selectable_index(entities)
    local result = {}
    for _, entity in ipairs(entities) do
        local selectable = ecs.get(entity, "d2legacy.world.selectable")
        if selectable then
            result[selectable:get("id")] = entity
        end
    end
    return result
end

local function reactions_for(entities, defender)
    local result = {}
    for _, entity in ipairs(entities) do
        local instance = ecs.get(entity, "d2legacy.state.instance")
        if
            instance
            and instance:get("target"):id() == defender:id()
            and instance:get("on_melee_hit_state_id") ~= ""
            and instance:get("on_melee_hit_duration") > 0
        then
            result[#result + 1] = instance
        end
    end
    return result
end

function M.register()
    ecs.system({
        id = "d2legacy.state.react_to_melee_hit",
        phase = "effects",
        after = { "d2legacy.monster.death" },
        query = {
            any = {
                "d2legacy.combat.melee_event",
                "d2legacy.state.instance",
                "d2legacy.world.selectable",
            },
            none = { "d2legacy.world.inactive" },
        },
        read = {
            "d2legacy.combat.melee_event",
            "d2legacy.state.instance",
            "d2legacy.world.selectable",
            "d2legacy.monster.stats",
        },
        write = {
            "d2legacy.combat.melee_event",
            "d2legacy.state.request",
        },
        update = function(_, entities, structural)
            local by_id = selectable_index(entities)
            for _, event_entity in ipairs(entities) do
                local event = ecs.get(event_entity, "d2legacy.combat.melee_event")
                if event and not event:get("defender_effects_processed") then
                    event:set("defender_effects_processed", true)
                    local defender = event:get("hit") and by_id[event:get("target_id")] or nil
                    local attacker = defender and by_id[event:get("attacker_id")] or nil
                    -- PvP uses chill rather than freeze and remains a separate
                    -- hostility/target-policy slice. Do not apply the monster
                    -- reaction to a player attacker by resemblance.
                    if attacker and ecs.get(attacker, "d2legacy.monster.stats") then
                        for _, reaction in ipairs(reactions_for(entities, defender)) do
                            structural:create({
                                ["d2legacy.state.request"] = {
                                    operation = "apply",
                                    target = attacker,
                                    state_id = reaction:get("on_melee_hit_state_id"),
                                    source_id = reaction:get("source_id") .. ":melee-response",
                                    duration = cold_duration.monster_frames(
                                        reaction:get("on_melee_hit_duration"),
                                        game_rules.difficulty()
                                    ),
                                    policy = "refresh_same_source",
                                    action_disabled = reaction:get("on_melee_hit_disables_action"),
                                },
                            })
                        end
                    end
                end
            end
        end,
    })
end

return M
