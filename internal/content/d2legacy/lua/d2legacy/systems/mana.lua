-- Advance authoritative Diablo II mana at the 25 Hz simulation cadence.
--
-- Life/mana/stamina use 8.8 fixed-point values. CharStats supplies the base
-- full-regeneration divisor; named sources contribute percentage and flat
-- recovery. Suppression remains a separate ECS relationship so a resource
-- consumer does not need to know which skill or state currently owns it.

local ecs = require("engine.ecs/v1")
local M = {}

local function signed32(value)
    local wrapped = value % 4294967296
    if wrapped >= 2147483648 then
        return wrapped - 4294967296
    end
    return wrapped
end

local function trunc(value)
    return math.modf(value)
end

-- PMR calls NRF-R05 with the percentage factor first and divisor 100.
-- For signed-32 raw maximum mana, the wide-product branches cannot be reached.
local function mul_div_100(factor, base)
    if base > 0x100000 then
        return signed32(trunc(base / 100) * factor)
    end
    if factor > 0x10000 then
        return signed32(trunc(factor / 100) * base)
    end
    return trunc(signed32(factor * base) / 100)
end

local function suppressed_targets(entities)
    local result = {}
    for _, entity in ipairs(entities) do
        local suppression = ecs.get(entity, "d2legacy.resource.mana_regen_suppression")
        if suppression then
            result[suppression:get("target"):id()] = true
        end
    end
    return result
end

local function update(entity, suppressed)
    local vitals = ecs.get(entity, "d2legacy.player.vitals")
    local stats = ecs.get(entity, "d2legacy.player.resource_stats")
    local current = vitals:get("mana_raw")
    local maximum = vitals:get("max_mana_raw")
    local recovery = signed32(stats:get("manarecovery"))
    local frames = stats:get("mana_regen_frames")
    if not suppressed and frames > 0 then
        local base = math.floor(maximum / (25 * frames))
        if base < 1 then
            base = 1
        end
        recovery = signed32(recovery + mul_div_100(signed32(100 + stats:get("manarecoverybonus")), base))
    end
    local room = signed32(maximum - current)
    if recovery > room then
        recovery = room
    end
    local lower = signed32(-current)
    if recovery < lower then
        recovery = lower
    end
    current = current + recovery
    vitals:set("mana_raw", current)
    vitals:set("mana", math.floor(current / 256))
    vitals:set("max_mana", math.floor(maximum / 256))
end

function M.register()
    ecs.system({
        id = "d2legacy.player.mana",
        phase = "pre_simulation",
        after = { "d2legacy.combat.derived_stats", "d2legacy.skill.aura_periodic_effect" },
        query = {
            any = {
                "d2legacy.player.resource_stats",
                "d2legacy.resource.mana_regen_suppression",
            },
            none = { "d2legacy.world.inactive" },
        },
        read = {
            "d2legacy.player.resource_stats",
            "d2legacy.resource.mana_regen_suppression",
            "d2legacy.player.death",
            "d2legacy.world.inactive",
        },
        write = { "d2legacy.player.vitals" },
        update = function(_, entities)
            local suppressed = suppressed_targets(entities)
            for _, entity in ipairs(entities) do
                if
                    ecs.get(entity, "d2legacy.player.resource_stats")
                    and ecs.get(entity, "d2legacy.player.vitals")
                    and not ecs.get(entity, "d2legacy.player.death")
                then
                    update(entity, suppressed[entity:id()] == true)
                end
            end
        end,
    })
end

return M
