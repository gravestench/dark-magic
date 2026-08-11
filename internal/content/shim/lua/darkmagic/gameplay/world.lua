-- ECS-owned world motion used by the game-world scene.
--
-- If ECS is new to you, think of it like LEGO:
--
--   entity     = one numbered LEGO model
--   component  = one kind of data brick attached to a model
--   system     = a rule that finds models with certain bricks and updates them
--
-- For example, something with Position + Velocity + Bounds can be moved by the
-- movement system without that system needing to know whether the entity is a
-- hero, monster, camera, or something else.
--
-- IMPORTANT AUTHORITY RULE: the real player entity is admitted by the fixed-tick
-- session. This Lua helper BINDS to it; it does not invent a second player because
-- a presentation scene happened to load early.
--
-- World coordinates here are continuous DS1 SUBTILES, not screen pixels. The
-- presentation layer later projects subtiles to pixels through dm.world/v1.

local ecs = require("dm.ecs/v1")
local components = require("darkmagic.gameplay.components.init")
local systems = require("darkmagic.gameplay.systems.init")

local M = {}

function M.create(width, height, collision, player)
    -- world.lua is the composition root: it chooses which data and behavior
    -- make up this playable world, while the imported modules explain details.
    components.register()
    systems.register(collision)

    -- Default logical player ID used by the local single-player shim.
    player = player or "local-player"

    -- `state` is presentation-side binding state, not a second simulation world.
    local state = { player = player }
    M.bind(state)
    return state
end

-- Bind presentation to a player entity admitted by the authoritative session.
--
-- This may return false briefly when a scene transition happens before the next
-- fixed simulation tick admits the selected player. That is normal. The caller
-- can try again next update. We NEVER manufacture a second player as a shortcut.
function M.bind(state)
    -- Already bound and still alive: nothing to do.
    if state.hero and state.hero:exists() then return true end

    -- Find entities that have the complete set needed by world presentation.
    local entities = ecs.query({
        all = {
            "dm.player.identity", "dm.player.appearance", "dm.player.animation",
            "dm.world.position", "dm.world.velocity",
            "dm.world.player_control", "dm.world.bounds", "dm.world.collider",
        },
    })

    for _, entity in ipairs(entities) do
        local control = ecs.get(entity, "dm.world.player_control")

        if control:get("player") == state.player then
            state.hero = entity

            -- Camera IS a Lua-created ECS entity because camera-follow is
            -- presentation state, not a second authoritative player.
            state.camera = ecs.create({
                -- Snapshot copies current position values instead of aliasing the
                -- hero's live component storage.
                ["dm.world.position"] = ecs.get(entity, "dm.world.position"):snapshot(),
                ["dm.world.camera_follow"] = { target = entity },
            })
            return true
        end
    end

    return false
end

function M.position(entity)
    local position = assert(ecs.get(entity, "dm.world.position"))
    -- Lua returns multiple values naturally, so callers can write `x, y = ...`.
    return position:get("x"), position:get("y")
end

function M.set_collision(state, collision)
	state.collision = collision
	systems.set_collision(collision)
end

function M.composite_snapshot(entity)
    local appearance = assert(ecs.get(entity, "dm.player.appearance"))
    local animation = assert(ecs.get(entity, "dm.player.animation"))
    local snapshot = appearance:snapshot()
    snapshot.direction = animation:get("direction")
    snapshot.mode = animation:get("mode")
    return snapshot
end

-- Copy every live monster into ordinary Lua values. Retained presentation code
-- may keep these tables for a frame; it never receives writable ECS component
-- handles. Entity ID is included only as a stable presentation-node key.
function M.monster_snapshots()
    local result = {}
    local entities = ecs.query({ all = {
        "dm.monster.identity", "dm.monster.appearance", "dm.world.position",
        "dm.world.velocity", "dm.world.location",
    }})
    for _, entity in ipairs(entities) do
        local identity = ecs.get(entity, "dm.monster.identity"):snapshot()
        local appearance = ecs.get(entity, "dm.monster.appearance"):snapshot()
        local position = ecs.get(entity, "dm.world.position")
        local velocity = ecs.get(entity, "dm.world.velocity")
        local location = ecs.get(entity, "dm.world.location")
        local ai = ecs.get(entity, "dm.monster.ai")
        local mode = appearance.mode
        if mode ~= "DT" and ai and ai:get("state") == "attack" then
            mode = "A1"
        elseif mode ~= "DT" and (velocity:get("x") ~= 0 or velocity:get("y") ~= 0) then
            mode = "WL"
        end
        result[#result + 1] = {
            entity_id = entity:id(), spawn_id = identity.spawn_id,
            token = appearance.token, mode = mode,
            weapon_class = appearance.weapon_class, components = appearance.components,
            name_key = appearance.name_key, death_sound = appearance.death_sound,
            x = position:get("x"), y = position:get("y"),
            velocity_x = velocity:get("x"), velocity_y = velocity:get("y"),
            act = location:get("act"), level_id = location:get("level_id"),
        }
    end
    return result
end

-- Semantic events are also copied. Consumers can remember entity_id to avoid
-- presenting one durable event more than once; observation itself is read-only.
function M.semantic_cues()
    local result = {}
    local entities, known = {}, {}
    for _, component in ipairs({"dm.monster.death_event", "dm.missile.event"}) do
        local ok, matches = pcall(ecs.query, {all = {component}})
        if ok then
            for _, entity in ipairs(matches) do
                if not known[entity:id()] then entities[#entities + 1] = entity; known[entity:id()] = true end
            end
        end
    end
    for _, entity in ipairs(entities) do
        local kind, values
        local death = ecs.get(entity, "dm.monster.death_event")
        if death then kind, values = "monster_death", death:snapshot() end
        local missile = ecs.get(entity, "dm.missile.event")
        if missile then kind, values = "missile", missile:snapshot() end
        if values then
            -- Dynamic entity fields are checked ECS handles. Collapse the
            -- missile reference to a number before this table crosses into
            -- presentation, just like the event entity itself below.
            if values.missile then
                values.missile_id = values.missile:id()
                values.missile = nil
            end
            values.entity_id = entity:id()
            values.cue_type = kind
            result[#result + 1] = values
        end
    end
    return result
end

-- Copy live projectile instances separately from their semantic spawn/hit
-- events. A renderer follows these values; it never predicts or advances them.
function M.missile_snapshots()
    local result = {}
    local ok, entities = pcall(ecs.query, { all = {
        "dm.missile.instance", "dm.world.position", "dm.world.location",
    } })
    if not ok then return result end
    for _, entity in ipairs(entities) do
        local snapshot = ecs.get(entity, "dm.missile.instance"):snapshot()
        local position = ecs.get(entity, "dm.world.position")
        local location = ecs.get(entity, "dm.world.location")
        snapshot.entity_id = entity:id()
        snapshot.owner_entity = nil
        snapshot.x, snapshot.y = position:get("x"), position:get("y")
        snapshot.act, snapshot.level_id = location:get("act"), location:get("level_id")
        result[#result + 1] = snapshot
    end
    return result
end

-- Build one VALUE-ONLY HUD snapshot.
--
-- Live simulation facts come from ECS. Save metadata fills only fields whose
-- runtime simulation schema is not authoritative yet. This prevents presentation
-- from accidentally preferring stale saved health/mana over live values.
function M.hud_snapshot(entity, saved)
    saved = saved or {}

    local progress = assert(ecs.get(entity, "dm.player.progress"))
    local vitals = assert(ecs.get(entity, "dm.player.vitals"))
    local movement_mode = assert(ecs.get(entity, "dm.player.movement_mode"))
    local skills = assert(ecs.get(entity, "dm.player.skill_assignment"))
    local belt = assert(ecs.get(entity, "dm.player.belt"))

    local belt_slots = {}
    for slot = 1, 16 do belt_slots[slot] = belt:get("slot_" .. slot) end

    local learned_skills = {}
    for _, learned in ipairs(ecs.query({ all = { "dm.player.learned_skill" } })) do
        local skill = ecs.get(learned, "dm.player.learned_skill")
        if skill:get("owner") == entity then
            -- Copy the component into ordinary values for the HUD. The HUD should
            -- not hold/write the live learned-skill component.
            learned_skills[#learned_skills + 1] = skill:snapshot()
        end
    end

    return {
        health = vitals:get("health"),
        max_health = vitals:get("max_health"),
        mana = vitals:get("mana"),
        max_mana = vitals:get("max_mana"),
        experience = progress:get("experience"),
        next_level_experience = saved.next_level_experience or 0,
        stamina = saved.stamina or 0,
        max_stamina = saved.max_stamina or 0,
        running = movement_mode:get("running"),
        left_skill = skills:get("left"),
        right_skill = skills:get("right"),
        belt_capacity = belt:get("capacity"),
        belt_slots = belt_slots,
        learned_skills = learned_skills,
    }
end

function M.destroy(state)
    if not state then return end

    -- Hero belongs to the authoritative session, so DO NOT destroy it here.
    -- The presentation-only camera entity was created above, so this helper owns
    -- exactly that entity and may tear it down when the scene disappears.
    if state.camera and state.camera:exists() then ecs.destroy(state.camera) end
end

return M
