-- Rebuild Expansion 1.14d maximum stamina from progression and active sources.
--
-- The legacy tables mix whole points, quarters, eighths, and 8.8 values. Go's
-- shared movement rule owns that arithmetic; this system owns ECS source
-- aggregation and the event-sensitive current-resource transition.

local ecs = require("engine.ecs/v1")
local movement_rules = require("d2legacy.movement_rules/v1")
local M = {}

local stamina_stats = {
    vitality = true,
    maxstamina = true,
    skill_staminapercent = true,
    skill_passive_staminapercent = true,
    item_stamina_perlevel = true,
    item_stamina_bytime = true,
}

local function totals_for(entities, target, base_time)
    local totals = {}
    for _, entity in ipairs(entities) do
        local source = ecs.get(entity, "d2legacy.stat.source")
        if source and source:get("target"):id() == target:id() and stamina_stats[source:get("stat")] then
            local operation = source:get("operation") ~= "" and source:get("operation") or "add"
            assert(operation == "add", "stamina ItemStatCost operands must be additive source stats")
            local stat = source:get("stat")
            if stat == "item_stamina_bytime" then
                totals[stat] = (totals[stat] or 0) + movement_rules.by_time_adjustment(source:get("value"), base_time)
            else
                totals[stat] = (totals[stat] or 0) + source:get("value")
            end
        end
    end
    return totals
end

local function base_time_for(entities, act)
    for _, entity in ipairs(entities) do
        local environment = ecs.get(entity, "d2legacy.world.environment")
        if environment and environment:get("act") == act then
            local rate = environment:get("time_rate")
            return rate > 0 and math.floor(environment:get("ticks") / rate) or 0
        end
    end
    return 0
end

function M.register()
    ecs.system({
        id = "d2legacy.player.maximum_stamina",
        phase = "effects",
        after = { "d2legacy.player.progression" },
        query = {
            any = {
                "d2legacy.player.stamina_progression",
                "d2legacy.stat.source",
                "d2legacy.world.environment",
            },
        },
        read = {
            "d2legacy.player.identity",
            "d2legacy.player.progress",
            "d2legacy.player.stamina_progression",
            "d2legacy.player.vitals",
            "d2legacy.world.location",
            "d2legacy.world.environment",
            "d2legacy.stat.source",
        },
        write = {
            "d2legacy.player.stamina_progression",
            "d2legacy.player.vitals",
        },
        update = function(_, entities)
            for _, player in ipairs(entities) do
                local identity = ecs.get(player, "d2legacy.player.identity")
                local progress = ecs.get(player, "d2legacy.player.progress")
                local basis = ecs.get(player, "d2legacy.player.stamina_progression")
                local vitals = ecs.get(player, "d2legacy.player.vitals")
                if identity and progress and basis and vitals then
                    local location = ecs.get(player, "d2legacy.world.location")
                    local totals = totals_for(entities, player, base_time_for(entities, location:get("act")))
                    local level = progress:get("level")
                    local base_vitality = basis:get("base_vitality")
                    local bonus_vitality = totals.vitality or 0
                    local maximum = movement_rules.maximum_stamina(
                        identity:get("class"),
                        level,
                        base_vitality,
                        bonus_vitality,
                        totals.maxstamina or 0,
                        totals.skill_staminapercent or 0,
                        totals.skill_passive_staminapercent or 0,
                        totals.item_stamina_perlevel or 0,
                        totals.item_stamina_bytime or 0
                    )
                    local previous = vitals:get("max_stamina_raw")
                    local current = vitals:get("stamina_raw")
                    if level > basis:get("last_level") then
                        -- PLAYERSTATS_LevelUp fills stamina after increasing max.
                        current = maximum
                    elseif maximum ~= previous then
                        current = movement_rules.rescale_stamina(current, previous, maximum)
                    end
                    basis:set("vitality", base_vitality + bonus_vitality)
                    basis:set("last_level", level)
                    vitals:set("stamina_raw", current)
                    vitals:set("max_stamina_raw", maximum)
                    vitals:set("stamina", math.floor(current / 256))
                    vitals:set("max_stamina", math.floor(maximum / 256))
                end
            end
        end,
    })
end

return M
