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
-- presentation layer later projects subtiles to pixels through engine.world/v1.

local ecs = require("engine.ecs/v1")
local components = require("d2legacy.gameplay.components.init")
local systems = require("d2legacy.gameplay.systems.init")

local M = {}

-- Presentation may run with only the schemas supplied by the active authority
-- composition. Missing optional event families mean "no cue", not a fatal
-- scene error. ecs.get still reports real failures for required components in
-- the focused snapshot functions below.
local function optional_component(entity, name)
    local ok, component = pcall(ecs.get, entity, name)
    if not ok then return nil end
    return component
end

function M.create(width, height, collision, player)
    -- world.lua is the composition root: it chooses which data and behavior
    -- make up this playable world, while the imported modules explain details.
    components.register()
    systems.register(collision)

    -- Default logical player ID used by the local single-player d2legacy mod.
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
            "d2legacy.player.identity", "d2legacy.player.appearance", "d2legacy.player.animation",
            "d2legacy.world.position", "d2legacy.world.velocity",
            "d2legacy.world.player_control", "d2legacy.world.bounds", "d2legacy.world.collider",
        },
    })

    for _, entity in ipairs(entities) do
        local control = ecs.get(entity, "d2legacy.world.player_control")

        if control:get("player") == state.player then
            state.hero = entity

            -- Camera IS a Lua-created ECS entity because camera-follow is
            -- presentation state, not a second authoritative player.
            state.camera = ecs.create({
                -- Snapshot copies current position values instead of aliasing the
                -- hero's live component storage.
                ["d2legacy.world.position"] = ecs.get(entity, "d2legacy.world.position"):snapshot(),
                ["d2legacy.world.camera_follow"] = { target = entity },
            })
            return true
        end
    end

    return false
end

function M.position(entity)
    local position = assert(ecs.get(entity, "d2legacy.world.position"))
    -- Lua returns multiple values naturally, so callers can write `x, y = ...`.
    return position:get("x"), position:get("y")
end

function M.set_collision(state, collision)
	state.collision = collision
	systems.set_collision(collision)
end

function M.composite_snapshot(entity)
    local appearance = assert(ecs.get(entity, "d2legacy.player.appearance"))
    local animation = assert(ecs.get(entity, "d2legacy.player.animation"))
    local facing = assert(ecs.get(entity, "d2legacy.world.facing"))
    local snapshot = appearance:snapshot()
    snapshot.direction = facing:get("direction")
    snapshot.mode = animation:get("mode")
    return snapshot
end

-- Copy every live monster into ordinary Lua values. Retained presentation code
-- may keep these tables for a frame; it never receives writable ECS component
-- handles. Entity ID is included only as a stable presentation-node key.
function M.monster_snapshots()
    local result = {}
    local entities = ecs.query({ all = {
        "d2legacy.monster.identity", "d2legacy.monster.appearance", "d2legacy.world.position",
        "d2legacy.world.velocity", "d2legacy.world.facing", "d2legacy.world.location",
    }})
    for _, entity in ipairs(entities) do
        local identity = ecs.get(entity, "d2legacy.monster.identity"):snapshot()
        local appearance = ecs.get(entity, "d2legacy.monster.appearance"):snapshot()
        local position = ecs.get(entity, "d2legacy.world.position")
        local velocity = ecs.get(entity, "d2legacy.world.velocity")
        local facing = ecs.get(entity, "d2legacy.world.facing")
        local location = ecs.get(entity, "d2legacy.world.location")
        local ai = ecs.get(entity, "d2legacy.monster.ai")
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
            direction = facing:get("direction"),
            act = location:get("act"), level_id = location:get("level_id"),
        }
    end
    return result
end

-- Semantic events are also copied. Consumers can remember entity_id to avoid
-- presenting one durable event more than once; observation itself is read-only.
function M.semantic_cues(observed)
    observed = observed or {}
    local result = {}
    local entities, known = {}, {}
    for _, component in ipairs({"d2legacy.monster.death_event", "d2legacy.missile.event",
        "d2legacy.combat.melee_event", "d2legacy.combat.event"}) do
        local ok, matches = pcall(ecs.query, {all = {component}})
        if ok then
            for _, entity in ipairs(matches) do
                local id = entity:id()
                if not observed[id] and not known[id] then
                    entities[#entities + 1] = entity
                    known[id] = true
                end
            end
        end
    end
    for _, entity in ipairs(entities) do
        local kind, values
        local death = optional_component(entity, "d2legacy.monster.death_event")
        if death then kind, values = "monster_death", death:snapshot() end
        local missile = optional_component(entity, "d2legacy.missile.event")
        if missile then kind, values = "missile", missile:snapshot() end
        local melee = optional_component(entity, "d2legacy.combat.melee_event")
        if melee then kind, values = "combat", melee:snapshot() end
        local combat = optional_component(entity, "d2legacy.combat.event")
        if combat then kind, values = "combat", combat:snapshot() end
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
    -- Projectile schemas are d2legacy authority facts. Presentation copies
    -- either supported shape without deciding how the projectile behaves.
    for _, component in ipairs({ "d2legacy.missile.instance", "d2legacy.missile.projectile" }) do
        local ok, entities = pcall(ecs.query, { all = {
            component, "d2legacy.world.position", "d2legacy.world.location",
        } })
        if ok then
            for _, entity in ipairs(entities) do
                local snapshot = ecs.get(entity, component):snapshot()
                local position = ecs.get(entity, "d2legacy.world.position")
                local location = ecs.get(entity, "d2legacy.world.location")
                snapshot.entity_id = entity:id()
                snapshot.owner_entity = nil
                snapshot.x, snapshot.y = position:get("x"), position:get("y")
                snapshot.act, snapshot.level_id = location:get("act"), location:get("level_id")
                result[#result + 1] = snapshot
            end
        end
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

    local progress = assert(ecs.get(entity, "d2legacy.player.progress"))
    local vitals = assert(ecs.get(entity, "d2legacy.player.vitals"))
    local movement_mode = assert(ecs.get(entity, "d2legacy.player.movement_mode"))
    local skills = assert(ecs.get(entity, "d2legacy.player.skill_assignment"))
    local belt = assert(ecs.get(entity, "d2legacy.player.belt"))

    local belt_slots = {}
    for slot = 1, 16 do belt_slots[slot] = belt:get("slot_" .. slot) end

    local learned_skills = {}
    for _, learned in ipairs(ecs.query({ all = { "d2legacy.player.learned_skill" } })) do
        local skill = ecs.get(learned, "d2legacy.player.learned_skill")
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
