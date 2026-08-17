-- Authoritative expansion party relationships and same-level membership queries.
--
-- High-confidence recovered server structure establishes explicit party IDs,
-- invitations, membership, leave-time dissolution below two members, and
-- living same-level iteration. A 1.14d probe still owns target-version UI event
-- timing; this module implements no earlier-version compatibility branch.

local ecs = require("engine.ecs/v1")
local state = require("engine.authority_state/v1")
local game_rules = require("d2legacy.policy.game_rules")

local M = {}
local STATE_ID = "d2legacy.party"
local STATE_SCHEMA = "d2legacy.party/v1"

local function present(value)
    return type(value) == "string" and value:match("%S") ~= nil
end

local function player_entity(player_id, entities)
    for _, entity in ipairs(entities or ecs.query({ all = { "d2legacy.player.identity" } })) do
        local identity = ecs.get(entity, "d2legacy.player.identity")
        if identity and identity:get("player") == player_id then
            return entity
        end
    end
    return nil
end

local function current()
    return assert(state.read(STATE_ID), "party state is not initialized")
end

local function commit(value)
    value.revision = value.revision + 1
    state.replace(STATE_ID, STATE_SCHEMA, value)
end

local function invitation(value, inviter, target)
    local received = value.invites[target]
    return received and received[inviter] or nil
end

local function remove_invitations_for(value, player_id)
    local changed = value.invites[player_id] ~= nil
    value.invites[player_id] = nil
    for target, received in pairs(value.invites) do
        if received[player_id] then
            received[player_id] = nil
            changed = true
        end
        if next(received) == nil then
            value.invites[target] = nil
        end
    end
    return changed
end

local function remove_member(value, player_id)
    local party_id = value.membership[player_id]
    if not party_id then
        return false
    end
    local party = assert(value.parties[party_id], "party membership references an unknown party")
    local removed = false
    for index, member in ipairs(party.members) do
        if member == player_id then
            table.remove(party.members, index)
            removed = true
            break
        end
    end
    assert(removed, "party membership is absent from its party")
    value.membership[player_id] = nil

    -- High-confidence recovered structure destroys the party when only one
    -- member remains and clears that final player's identity during teardown.
    if #party.members <= 1 then
        for _, member in ipairs(party.members) do
            value.membership[member] = nil
        end
        value.parties[party_id] = nil
    end
    return true
end

function M.initialize()
    state.register(STATE_ID, STATE_SCHEMA, {
        schema = STATE_SCHEMA,
        revision = 0,
        next_party_id = 1,
        parties = {},
        membership = {},
        invites = {},
    })
end

function M.snapshot()
    return current()
end

function M.invite(inviter, target, tick)
    assert(present(inviter) and present(target) and inviter ~= target, "party invitation players are invalid")
    assert(player_entity(inviter), "party inviter is not present")
    assert(player_entity(target), "party invitation target is not present")

    local value = current()
    assert(not value.membership[target], "party invitation target is already in a party")
    assert(not invitation(value, inviter, target), "party invitation already exists")
    local received = value.invites[target] or {}
    received[inviter] = { inviter = inviter, target = target, created_tick = tick }
    value.invites[target] = received
    commit(value)
end

function M.cancel(inviter, target)
    assert(present(inviter) and present(target) and inviter ~= target, "party invitation players are invalid")
    local value = current()
    assert(invitation(value, inviter, target), "party invitation does not exist")
    local received = value.invites[target]
    received[inviter] = nil
    if next(received) == nil then
        value.invites[target] = nil
    end
    commit(value)
end

function M.accept(target, inviter)
    assert(present(target) and present(inviter) and target ~= inviter, "party acceptance players are invalid")
    assert(player_entity(target), "party acceptance player is not present")
    assert(player_entity(inviter), "party inviter is not present")

    local value = current()
    assert(invitation(value, inviter, target), "party invitation does not exist")
    assert(not value.membership[target], "party acceptance player is already in a party")

    local party_id = value.membership[inviter]
    local party
    if party_id then
        party = assert(value.parties[party_id], "party membership references an unknown party")
    else
        party_id = "party:" .. value.next_party_id
        value.next_party_id = value.next_party_id + 1
        party = { id = party_id, members = { inviter } }
        value.parties[party_id] = party
        value.membership[inviter] = party_id
    end
    assert(#party.members < game_rules.get().maximum_players, "party is full")
    table.insert(party.members, target)
    value.membership[target] = party_id
    value.invites[target] = nil
    commit(value)
end

function M.leave(player_id)
    assert(present(player_id), "party player is required")
    local value = current()
    assert(value.membership[player_id], "player is not in a party")
    remove_member(value, player_id)
    remove_invitations_for(value, player_id)
    commit(value)
end

-- Game departure is intentionally idempotent. A transport disconnect does not
-- call it during the reconnect grace period, so membership survives reconnect.
function M.depart(player_id)
    assert(present(player_id), "party player is required")
    local value = current()
    local changed = remove_member(value, player_id)
    if remove_invitations_for(value, player_id) then
        changed = true
    end
    if changed then
        commit(value)
    end
end

function M.living_members_in_same_level(player_id, entities)
    assert(present(player_id), "party player is required")
    local source = assert(player_entity(player_id, entities), "party player is not present")
    local source_location = assert(ecs.get(source, "d2legacy.world.location"))
    local value = current()
    local party_id = value.membership[player_id]
    local candidates = party_id
            and assert(value.parties[party_id], "party membership references an unknown party").members
        or { player_id }
    local result = {}
    for _, member in ipairs(candidates) do
        local entity = player_entity(member, entities)
        if entity then
            local location = ecs.get(entity, "d2legacy.world.location")
            local vitals = ecs.get(entity, "d2legacy.player.vitals")
            if
                location
                and vitals
                and location:get("level_id") == source_location:get("level_id")
                and vitals:get("health") > 0
            then
                table.insert(result, member)
            end
        end
    end
    return result
end

function M.additional_living_members_in_same_level(player_id, entities)
    local members = M.living_members_in_same_level(player_id, entities)
    for index, member in ipairs(members) do
        if member == player_id then
            table.remove(members, index)
            break
        end
    end
    return #members
end

return M
