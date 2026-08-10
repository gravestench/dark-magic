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

local M = {}

local function define_components()
    -- Components declare stable schemas. A schema says which fields exist and
    -- what type each field stores. The ECS capability owns the actual storage.

    ecs.component({
        name = "dm.player.identity",
        version = 1,
        fields = {
            { name = "character_id", type = "string" },
            { name = "player", type = "string" },
            { name = "name", type = "string" },
            { name = "class", type = "string" },
        },
    })

    ecs.component({
        name = "dm.player.progress",
        version = 1,
        fields = {
            { name = "level", type = "i64" },
            { name = "experience", type = "i64" },
        },
    })

    ecs.component({
        name = "dm.player.vitals",
        version = 1,
        fields = {
            { name = "health", type = "i64" },
            { name = "max_health", type = "i64" },
            { name = "mana", type = "i64" },
            { name = "max_mana", type = "i64" },
        },
    })

    ecs.component({
        name = "dm.player.movement_mode",
        version = 1,
        -- One-field component can still use the same schema shape.
        fields = {{ name = "running", type = "bool" }},
    })

    ecs.component({
        name = "dm.player.skill_assignment",
        version = 1,
        fields = {
            { name = "left", type = "i64" },
            { name = "right", type = "i64" },
        },
    })

    -- Belt capacity is fixed at a maximum of 16 slots in this presentation
    -- schema. Build the repeated slot field declarations with a loop instead of
    -- hand-writing sixteen almost-identical lines.
    local belt_fields = {{ name = "capacity", type = "i64" }}
    for slot = 1, 16 do
        table.insert(belt_fields, { name = "slot_" .. slot, type = "string" })
    end
    ecs.component({ name = "dm.player.belt", version = 1, fields = belt_fields })

    -- Learned skills are separate entities owned by a player entity. That lets a
    -- query enumerate the learned set without putting an arbitrary array inside
    -- one mutable component blob.
    ecs.component({
        name = "dm.player.learned_skill",
        version = 1,
        fields = {
            { name = "owner", type = "entity" },
            { name = "skill_id", type = "i64" },
            { name = "level", type = "i64" },
            { name = "list_row", type = "i64" },
            { name = "left_allowed", type = "bool" },
            { name = "right_allowed", type = "bool" },
        },
    })

    ecs.component({
        name = "dm.world.position",
        fields = {
            { name = "x", type = "f64" },
            { name = "y", type = "f64" },
        },
    })

    ecs.component({
        name = "dm.world.velocity",
        fields = {
            { name = "x", type = "f64" },
            { name = "y", type = "f64" },
        },
    })

    -- Which logical player controls this entity. This is separate from identity
    -- so control/session ownership can be queried independently.
    ecs.component({
        name = "dm.world.player_control",
        fields = {{ name = "player", type = "string" }},
    })

    ecs.component({
        name = "dm.world.bounds",
        fields = {
            { name = "width", type = "f64" },
            { name = "height", type = "f64" },
        },
    })

    -- Camera-follow entity stores only a reference to the entity it follows.
    ecs.component({
        name = "dm.world.camera_follow",
        fields = {{ name = "target", type = "entity" }},
    })
end

local function clamp(value, minimum, maximum)
    -- Common clamp idiom: first cap at maximum, then floor at minimum.
    return math.max(minimum, math.min(maximum, value))
end

function M.create(width, height, collision, player)
    define_components()

    -- Default logical player ID used by the local single-player shim.
    player = player or "local-player"

    ecs.system({
        id = "darkmagic.world.integrate",
        phase = "movement",

        -- Query says which entities participate: they must have ALL 3 components.
        query = { all = { "dm.world.position", "dm.world.velocity", "dm.world.bounds" } },

        -- Explicit read/write declarations document and constrain system access.
        read = { "dm.world.velocity", "dm.world.bounds" },
        write = { "dm.world.position" },

        update = function(context, entities)
            for _, entity in ipairs(entities) do
                local position = ecs.get(entity, "dm.world.position")
                local velocity = ecs.get(entity, "dm.world.velocity")
                local bounds = ecs.get(entity, "dm.world.bounds")

                -- Component handles expose typed get/set methods; Lua does not
                -- receive a mutable Go struct pointer.
                local x, y = position:get("x"), position:get("y")

                -- Euler integration: new position = old + velocity * elapsed time.
                -- Then clamp to declared world bounds.
                local next_x = clamp(x + velocity:get("x") * context.delta_seconds, 0, bounds:get("width"))

                -- Collision lives in DS1 subtile cells, so floor continuous
                -- coordinates only at the collision query boundary.
                if not collision or not collision:blocked(math.floor(next_x), math.floor(y)) then
                    x = next_x
                end

                -- Resolve Y after X. Notice the Y collision check uses the possibly
                -- accepted new X, which gives simple axis-separated sliding.
                local next_y = clamp(y + velocity:get("y") * context.delta_seconds, 0, bounds:get("height"))
                if not collision or not collision:blocked(math.floor(x), math.floor(next_y)) then
                    y = next_y
                end

                position:set("x", x)
                position:set("y", y)
            end
        end,
    })

    ecs.system({
        id = "darkmagic.world.camera_follow",
        phase = "presentation",
        query = { all = { "dm.world.position", "dm.world.camera_follow" } },
        read = { "dm.world.camera_follow" },
        write = { "dm.world.position" },

        -- First argument would be system context, but this rule does not need it.
        update = function(_, entities)
            for _, entity in ipairs(entities) do
                local follow = ecs.get(entity, "dm.world.camera_follow")
                local target = ecs.get(follow:get("target"), "dm.world.position")
                local position = ecs.get(entity, "dm.world.position")

                -- Camera entity simply copies the target's authoritative world position.
                position:set("x", target:get("x"))
                position:set("y", target:get("y"))
            end
        end,
    })

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
            "dm.player.identity", "dm.world.position", "dm.world.velocity",
            "dm.world.player_control", "dm.world.bounds",
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
