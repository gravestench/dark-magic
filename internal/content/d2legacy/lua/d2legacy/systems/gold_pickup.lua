-- Collect a trusted world pile through the ordinary interaction operation.
-- Recipient selection follows GPS-R01/R02; persistence of leftover gold on
-- this same pile is a provisional Dark Magic integration choice.
local ecs = require("engine.ecs/v1")
local operations = require("d2legacy.interactions.operations")
local party = require("d2legacy.policy.party")
local split = require("d2legacy.policy.party_gold")

local M = {}
local MAX_GOLD = 2147483647

local function valid_level(value)
    return type(value) == "number" and value % 1 == 0 and value >= 1 and value <= 99
end

local function valid_gold(value)
    return type(value) == "number" and value % 1 == 0 and value >= 0 and value <= MAX_GOLD
end

local function player_entities()
    local result = {}
    local duplicates = {}
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.player.identity" } })) do
        local identity = ecs.get(entity, "d2legacy.player.identity")
        local id = identity:get("player")
        if result[id] then
            duplicates[id] = true
        else
            result[id] = entity
        end
    end
    return result, duplicates
end

local function layouts()
    local result = {}
    local duplicates = {}
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.items.layout" } })) do
        local layout = ecs.get(entity, "d2legacy.items.layout")
        local id = layout:get("owner")
        if result[id] then
            duplicates[id] = true
        else
            result[id] = layout
        end
    end
    return result, duplicates
end

local function resident(entity, act, level_id)
    local location = ecs.get(entity, "d2legacy.world.location")
    return location and location:get("act") == act and location:get("level_id") == level_id
end

local function member_snapshot(id, entity, layout, act, level_id)
    if not entity or not resident(entity, act, level_id) then
        return nil
    end
    local vitals = ecs.get(entity, "d2legacy.player.vitals")
    local living = vitals and vitals:get("health") > 0 and not ecs.get(entity, "d2legacy.player.death")
    local progress = ecs.get(entity, "d2legacy.player.progress")
    local level = progress and progress:get("level")
    local gold = layout and layout:get("carried_gold")
    local valid = valid_level(level) and valid_gold(gold)
    return {
        id = id,
        counts_for_divisor = not not living,
        can_receive = not not valid,
        level = valid and level or 1,
        gold = valid and gold or 0,
    }
end

local function pickup(owner, target, pile)
    local amount = pile:get("amount")
    if ecs.get(target, "d2legacy.world.inactive") or not valid_gold(amount) or amount == 0 then
        return
    end
    local target_location = ecs.get(target, "d2legacy.world.location")
    local target_position = ecs.get(target, "d2legacy.world.position")
    local interaction = ecs.get(target, "d2legacy.interaction.target")
    if not target_location or not target_position or not interaction then
        return
    end

    local players, duplicate_players = player_entities()
    local balances, duplicate_layouts = layouts()
    local picker = players[owner]
    if not picker or duplicate_players[owner] or duplicate_layouts[owner] then
        return
    end
    local act, level_id = target_location:get("act"), target_location:get("level_id")
    local picker_member = member_snapshot(owner, picker, balances[owner], act, level_id)
    if not picker_member or not picker_member.counts_for_divisor or not picker_member.can_receive then
        return
    end
    local position = ecs.get(picker, "d2legacy.world.position")
    if not position then
        return
    end
    local dx = position:get("x") - target_position:get("x")
    local dy = position:get("y") - target_position:get("y")
    local radius = interaction:get("radius")
    if radius <= 0 or dx * dx + dy * dy > radius * radius then
        return
    end

    local members = {}
    local state = party.snapshot()
    local party_id = state.membership[owner]
    local record = party_id and state.parties[party_id]
    if record then
        for _, id in ipairs(record.members) do
            local member
            if id == owner then
                member = picker_member
            else
                member = member_snapshot(id, players[id], balances[id], act, level_id)
                if member and (duplicate_players[id] or duplicate_layouts[id]) then
                    member.can_receive = false
                end
            end
            if member then
                members[#members + 1] = member
            end
        end
    end
    local found_picker = false
    for _, member in ipairs(members) do
        if member.id == owner then
            found_picker = true
        end
    end
    if not found_picker then
        members = { picker_member }
    end

    local allocation = split.allocate(amount, owner, members, true)
    for _, request in ipairs(allocation.requests) do
        balances[request.id]:set("carried_gold", request.next_balance)
    end
    pile:set("amount", allocation.excess)
    if allocation.excess == 0 then
        ecs.set(target, "d2legacy.world.inactive", {})
    end
end

function M.register()
    operations.register("d2legacy.gold.pile", pickup)
end

return M
