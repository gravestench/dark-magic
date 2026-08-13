-- Admit a player's request to use a skill.
--
-- Go has already authenticated the player, tick, and sequence when these
-- functions run. Validation below is Diablo policy: the payload must describe
-- one finite target and a positive learned-skill level.

local commands = require("engine.authority_command/v1")
local ecs = require("engine.ecs/v1")
local skills = require("d2legacy.data.skill")

local M = {}
local state_skills = {}

local function finite(value)
    return type(value) == "number" and value == value and value ~= math.huge and value ~= -math.huge
end

local function find_player(player_id)
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.player.identity" } })) do
        local identity = ecs.get(entity, "d2legacy.player.identity")
        if identity:get("player") == player_id then
            return entity
        end
    end
    return nil
end

local function learned_skill(player, skill_id)
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.player.learned_skill" } })) do
        local learned = ecs.get(entity, "d2legacy.player.learned_skill")
        if learned:get("owner"):id() == player:id() and learned:get("skill_id") == skill_id then
            return learned
        end
    end
    return nil
end

function M.validate_assignment(command)
    local payload = command.payload
    assert(type(payload) == "table", "assignment payload must be a table")
    assert(payload.left ~= nil or payload.right ~= nil, "assignment is empty")
    for _, side in ipairs({ "left", "right" }) do
        local skill_id = payload[side]
        assert(
            skill_id == nil or type(skill_id) == "number" and skill_id >= 0,
            side .. " skill must be a non-negative number"
        )
    end
end

function M.apply_assignment(command)
    local player = assert(find_player(command.player), "assignment player does not exist")
    local assignment = assert(ecs.get(player, "d2legacy.player.skill_assignment"))
    for _, side in ipairs({ "left", "right" }) do
        local skill_id = command.payload[side]
        if skill_id ~= nil then
            local learned = assert(learned_skill(player, skill_id), "skill is not learned")
            assert(learned:get(side .. "_allowed"), "skill is not allowed on " .. side)
            assignment:set(side, skill_id)
        end
    end
end

function M.validate(command)
    local payload = command.payload
    assert(type(payload) == "table", "cast payload must be a table")
    assert(payload.side == "left" or payload.side == "right", "cast side must be left or right")
    assert(payload.target_x == nil or finite(payload.target_x), "cast target X must be finite")
    assert(payload.target_y == nil or finite(payload.target_y), "cast target Y must be finite")
    assert(payload.target_id == nil or type(payload.target_id) == "string", "target ID must be a string")
end

function M.apply(command)
    local player = assert(find_player(command.player), "cast player does not exist")
    local payload = command.payload
    local assignments =
        assert(ecs.get(player, "d2legacy.player.skill_assignment"), "cast player has no skill assignments")
    local skill_id = assignments:get(payload.side)
    local state_definition = state_skills[skill_id]
    if state_definition then
        local learned = assert(learned_skill(player, skill_id), "skill is not learned")
        ecs.create({
            ["d2legacy.state.request"] = {
                operation = "apply",
                target = player,
                state_id = state_definition.state_id,
                source_id = "skill:" .. command.player .. ":" .. skill_id,
                duration = state_definition.duration,
                policy = "refresh_same_source",
            },
        })
        ecs.create({
            ["d2legacy.skill.cast_event"] = {
                kind = "skill_effect",
                tick = command.tick,
                player = command.player,
                skill_id = skill_id,
                skill_level = learned:get("level"),
                behavior = "state.self",
                target_x = 0,
                target_y = 0,
                target_id = "",
                reason = "",
            },
        })
        return
    end
    local values = {
        player = command.player,
        skill_id = skill_id,
        -- The lifecycle resolves the authoritative learned level. Input never
        -- gets to claim a level supplied by the client.
        skill_level = 0,
        target_x = payload.target_x,
        target_y = payload.target_y,
        target_id = payload.target_id or "",
        request_tick = command.tick,
    }
    if skill_id == 36 then
        assert(finite(payload.target_x) and finite(payload.target_y), "missile cast target must be finite")
        ecs.set(player, "d2legacy.skill.cast_request", values)
        return
    end

    -- The approach/animation adapter consumes this semantic fact. Cast
    -- admission and Diablo policy no longer cross into the old Go lifecycle.
    assert(skill_id == 0, "assigned skill has not migrated to d2legacy")
    local learned = assert(learned_skill(player, skill_id), "skill is not learned")
    ecs.create({
        ["d2legacy.skill.cast_event"] = {
            kind = "skill_effect",
            tick = command.tick,
            player = command.player,
            skill_id = skill_id,
            skill_level = learned:get("level"),
            behavior = "basic.melee",
            weapon_selection = skills.weapon_selection(skill_id),
            target_x = payload.target_x,
            target_y = payload.target_y,
            target_id = payload.target_id or "",
            reason = "",
        },
    })
end

function M.register(definitions)
    state_skills = definitions or {}
    commands.register({
        kind = "player.assign_skills",
        authorities = { "player" },
        validate = M.validate_assignment,
        apply = M.apply_assignment,
    })
    commands.register({
        kind = "player.use_skill",
        authorities = { "player" },
        validate = M.validate,
        apply = M.apply,
    })
end

return M
