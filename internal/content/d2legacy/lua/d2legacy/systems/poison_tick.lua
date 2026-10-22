-- Spend an admitted poison rate once per active frame.
local ecs = require("engine.ecs/v1")
local movement_rules = require("d2legacy.movement_rules/v1")
local M = {}

local function debit_player(vitals, rate)
    local before = vitals:get("health") * 256 - vitals:get("poison_remainder_raw")
    local after = math.max(before - rate, 0)
    local health = math.floor((after + 255) / 256)
    vitals:set("health", health)
    vitals:set("poison_remainder_raw", health * 256 - after)
    return before - after, after
end

local function debit_monster(stats, rate)
    local before = stats:get("health")
    local after = math.max(before - rate, 0)
    stats:set("health", after)
    return before - after, after
end

local function emit_death(structural, context, target_id, attacker_id, rate, spent)
    structural:create({
        ["d2legacy.combat.event"] = {
            kind = "unit_died",
            tick = context.tick,
            attacker_id = attacker_id,
            target_id = target_id,
            source_kind = "poison",
            damage_channel = "poison",
            rolled_damage_raw = rate,
            damage_raw = spent,
            remaining_health_raw = 0,
        },
    })
end

function M.register()
    ecs.system({
        id = "d2legacy.combat.poison_tick",
        phase = "pre_simulation",
        query = { all = { "d2legacy.state.instance" } },
        read = {
            "d2legacy.state.instance",
            "d2legacy.world.inactive",
            "d2legacy.world.location",
            "d2legacy.world.selectable",
            "d2legacy.player.death",
            "d2legacy.monster.death",
            "d2legacy.player.vitals",
            "d2legacy.monster.stats",
        },
        write = { "d2legacy.player.vitals", "d2legacy.monster.stats", "d2legacy.combat.event" },
        update = function(context, entities, structural)
            for _, entity in ipairs(entities) do
                local state = ecs.get(entity, "d2legacy.state.instance")
                if
                    state:get("state_id") == "poison"
                    and state:get("applied_tick") < context.tick
                    and context.tick <= state:get("expires_tick")
                then
                    local target = state:get("target")
                    local location = ecs.get(target, "d2legacy.world.location")
                    local in_town = location and movement_rules.is_town(location:get("level_id"))
                    if
                        not in_town
                        and not ecs.get(target, "d2legacy.world.inactive")
                        and not ecs.get(target, "d2legacy.player.death")
                        and not ecs.get(target, "d2legacy.monster.death")
                    then
                        local rate = math.max(-state:get("stat_value"), 0)
                        local player = ecs.get(target, "d2legacy.player.vitals")
                        local monster = ecs.get(target, "d2legacy.monster.stats")
                        local selectable = ecs.get(target, "d2legacy.world.selectable")
                        if rate > 0 and selectable then
                            local spent, remaining
                            if player and player:get("health") > 0 then
                                spent, remaining = debit_player(player, rate)
                            elseif monster and monster:get("health") > 0 then
                                spent, remaining = debit_monster(monster, rate)
                            end
                            if spent and remaining == 0 then
                                emit_death(
                                    structural,
                                    context,
                                    selectable:get("id"),
                                    state:get("source_attacker_id"),
                                    rate,
                                    spent
                                )
                            end
                        end
                    end
                end
            end
        end,
    })
end

return M
