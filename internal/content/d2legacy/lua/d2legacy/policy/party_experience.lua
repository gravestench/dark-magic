-- Monster-kill party XP recipients. Membership is owned by party.lua; this
-- policy only collects a death-time recipient snapshot and raw XP shares.
local ecs = require("engine.ecs/v1")
local party = require("d2legacy.policy.party")

local M = {}

local function truncate(value)
    if value < 0 then
        return math.ceil(value)
    end
    return math.floor(value)
end

local function players_by_id(entities)
    local players = {}
    for _, entity in ipairs(entities) do
        local identity = ecs.get(entity, "d2legacy.player.identity")
        if identity then
            players[identity:get("player")] = entity
        end
    end
    return players
end

local function admitted(entity, level_id, victim_x, victim_y)
    local identity = ecs.get(entity, "d2legacy.player.identity")
    local progress = ecs.get(entity, "d2legacy.player.progress")
    local vitals = ecs.get(entity, "d2legacy.player.vitals")
    local location = ecs.get(entity, "d2legacy.world.location")
    local position = ecs.get(entity, "d2legacy.world.position")
    if
        not identity
        or not progress
        or not vitals
        or not location
        or not position
        or ecs.get(entity, "d2legacy.player.death")
        or vitals:get("health") <= 0
        or location:get("level_id") ~= level_id
    then
        return nil
    end
    local x = position:get("x")
    local y = position:get("y")
    if victim_x == 0 or x == 0 then
        return nil
    end
    local dx = x - victim_x
    local dy = y - victim_y
    if dx * dx + dy * dy > 6400 then
        return nil
    end
    return { entity = entity, level = progress:get("level") }
end

-- Coordinates use authoritative f64 world units provisionally. The original
-- dynamic getter reads integer coordinates; Q03 and PXL-R04b need runtime
-- probes before this distance/quotient can be called 1.14d parity.
function M.shares(entities, credited_player, victim_position, raw_experience)
    local identity = ecs.get(credited_player, "d2legacy.player.identity")
    if not identity then
        return nil
    end
    local value = party.snapshot()
    local party_id = value.membership[identity:get("player")]
    if not party_id then
        return nil
    end
    local credited_location = ecs.get(credited_player, "d2legacy.world.location")
    local credited_position = ecs.get(credited_player, "d2legacy.world.position")
    if not victim_position or not credited_location or not credited_position then
        return nil
    end

    local record = value.parties[party_id]
    if not record then
        return {}
    end
    local players = players_by_id(entities)
    local recipients = {}
    local level_sum = 0
    local victim_x = victim_position:get("x")
    local victim_y = victim_position:get("y")
    local level_id = credited_location:get("level_id")
    for _, member_id in ipairs(record.members) do
        local member = players[member_id]
        local entry = member and admitted(member, level_id, victim_x, victim_y)
        if entry then
            assert(#recipients < 8, "party experience exceeds eight admitted members")
            recipients[#recipients + 1] = entry
            level_sum = level_sum + entry.level
        end
    end

    if #recipients == 0 or level_sum <= 0 then
        return {}
    end
    if #recipients == 1 then
        return { { entity = credited_player, raw = raw_experience } }
    end

    local pool = raw_experience + truncate((#recipients - 1) * raw_experience * 89 / 256)
    local quotient = pool / level_sum
    local result = {}
    for _, entry in ipairs(recipients) do
        result[#result + 1] = { entity = entry.entity, raw = truncate(quotient * entry.level) }
    end
    return result
end

return M
