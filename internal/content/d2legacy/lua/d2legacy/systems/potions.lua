-- One checkpointed hp1 rate per player; no item identity survives consumption.
local ecs = require("engine.ecs/v1")
local resources = require("d2legacy.policy.resources")
local M = {}

function M.apply(player, amount, duration)
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.player.potion" } })) do
        local effect = ecs.get(entity, "d2legacy.player.potion")
        if effect:get("owner"):id() == player:id() then
            local remaining = effect:get("remaining_ticks")
            effect:set("rate_raw", math.floor((amount + remaining * effect:get("rate_raw")) / (remaining + duration)))
            effect:set("remaining_ticks", remaining + duration)
            return
        end
    end
    ecs.create({
        ["d2legacy.player.potion"] = {
            owner = player,
            rate_raw = math.floor(amount / duration),
            remaining_ticks = duration,
        },
    })
end

function M.register()
    ecs.system({
        id = "d2legacy.player.potions",
        phase = "pre_simulation",
        after = { "d2legacy.combat.poison_tick" },
        query = { all = { "d2legacy.player.potion" } },
        read = {
            "d2legacy.player.potion",
            "d2legacy.player.vitals",
            "d2legacy.player.death",
            "d2legacy.world.inactive",
        },
        write = { "d2legacy.player.potion", "d2legacy.player.vitals" },
        update = function(_, entities, structural)
            for _, entity in ipairs(entities) do
                local effect = ecs.get(entity, "d2legacy.player.potion")
                local player = effect:get("owner")
                local vitals = player:exists() and ecs.get(player, "d2legacy.player.vitals")
                if not vitals or ecs.get(player, "d2legacy.player.death") or vitals:get("health") <= 0 then
                    structural:destroy(entity)
                elseif not ecs.get(player, "d2legacy.world.inactive") then
                    local rate = effect:get("rate_raw")
                    local room = (vitals:get("max_health") - vitals:get("health")) * 256
                        + vitals:get("poison_remainder_raw")
                    resources.restore_health(vitals, rate)
                    local remaining = effect:get("remaining_ticks") - 1
                    effect:set("remaining_ticks", remaining)
                    -- Recovered regen removes healthpot on overshoot, not equality.
                    if remaining == 0 or rate > room then
                        structural:destroy(entity)
                    end
                end
            end
        end,
    })
end

return M
