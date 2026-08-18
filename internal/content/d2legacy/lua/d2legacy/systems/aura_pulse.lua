-- Apply checkpointed direct effects for selected-aura target relationships.
--
-- Aura selection, party/radius eligibility, state arbitration, and visuals stay
-- in the shared selected-aura system. This consumer only executes a due pulse
-- and advances its durable schedule, so a repeated step cannot apply it twice.

local ecs = require("engine.ecs/v1")
local progression = require("d2legacy.policy.skill_progression")
local resources = require("d2legacy.policy.resources")
local M = {}

local function pulse_targets(entities, emitter)
    local result = {}
    local seen = {}
    for _, entity in ipairs(entities) do
        local effect = ecs.get(entity, "d2legacy.skill.aura_effect")
        if effect and effect:get("emitter"):id() == emitter:id() then
            local target = effect:get("target")
            local vitals = ecs.get(target, "d2legacy.player.vitals")
            if
                vitals
                and vitals:get("health") > 0
                and not ecs.get(target, "d2legacy.player.death")
                and not ecs.get(target, "d2legacy.world.inactive")
                and not seen[target:id()]
            then
                seen[target:id()] = true
                result[#result + 1] = target
            end
        end
    end
    table.sort(result, function(left, right)
        local left_identity = ecs.get(left, "d2legacy.player.identity")
        local right_identity = ecs.get(right, "d2legacy.player.identity")
        return left_identity:get("player") < right_identity:get("player")
    end)
    return result
end

local function heal_life(entities, emitter, pulse)
    local owner_vitals = ecs.get(emitter, "d2legacy.player.vitals")
    if not owner_vitals then
        return
    end
    local cost = pulse:get("mana_cost_raw")
    local available = resources.mana_raw(owner_vitals)
    if available < cost then
        return
    end
    local targets = pulse_targets(entities, emitter)
    local changed = false
    for _, target in ipairs(targets) do
        local vitals = ecs.get(target, "d2legacy.player.vitals")
        local health = vitals:get("health")
        local healed = math.min(health + pulse:get("value"), vitals:get("max_health"))
        if healed ~= health then
            vitals:set("health", healed)
            changed = true
        end
    end
    if changed then
        assert(resources.spend_mana(owner_vitals, cost))
    end
end

function M.register()
    ecs.system({
        id = "d2legacy.skill.aura_periodic_effect",
        phase = "pre_simulation",
        after = { "d2legacy.skill.selected_party_aura" },
        query = {
            any = {
                "d2legacy.skill.aura_pulse",
                "d2legacy.skill.aura_effect",
                "d2legacy.player.vitals",
            },
        },
        read = {
            "d2legacy.skill.aura_effect",
            "d2legacy.player.identity",
            "d2legacy.player.death",
            "d2legacy.world.inactive",
        },
        write = {
            "d2legacy.skill.aura_pulse",
            "d2legacy.player.vitals",
        },
        update = function(context, entities)
            for _, emitter in ipairs(entities) do
                local pulse = ecs.get(emitter, "d2legacy.skill.aura_pulse")
                if pulse and context.tick >= pulse:get("next_tick") then
                    pulse:set(
                        "next_tick",
                        progression.next_periodic_tick(context.tick, pulse:get("period_ticks"))
                    )
                    if pulse:get("kind") == "heal_life" then
                        heal_life(entities, emitter, pulse)
                    else
                        error("unsupported aura pulse kind " .. pulse:get("kind"))
                    end
                end
            end
        end,
    })
end

return M
