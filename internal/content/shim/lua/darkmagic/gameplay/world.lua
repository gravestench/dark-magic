-- ECS-owned world motion used by the game-world scene.
--
-- Coordinates are continuous DS1 subtiles. Presentation projects them through
-- dm.world/v1, so collision and rendering never share ambiguous pixel values.
local ecs = require("dm.ecs/v1")

local M = {}

local function define_components()
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
    ecs.component({
        name = "dm.world.camera_follow",
        fields = {{ name = "target", type = "entity" }},
    })
end

local function clamp(value, minimum, maximum)
    return math.max(minimum, math.min(maximum, value))
end

function M.create(width, height, collision, player)
    define_components()
    player = player or "local-player"

    ecs.system({
        id = "darkmagic.world.integrate",
        phase = "movement",
        query = { all = { "dm.world.position", "dm.world.velocity", "dm.world.bounds" } },
        read = { "dm.world.velocity", "dm.world.bounds" },
        write = { "dm.world.position" },
        update = function(context, entities)
            for _, entity in ipairs(entities) do
                local position = ecs.get(entity, "dm.world.position")
                local velocity = ecs.get(entity, "dm.world.velocity")
                local bounds = ecs.get(entity, "dm.world.bounds")
                local x, y = position:get("x"), position:get("y")
                local next_x = clamp(x + velocity:get("x") * context.delta_seconds, 0, bounds:get("width"))
                if not collision or not collision:blocked(math.floor(next_x), math.floor(y)) then
                    x = next_x
                end
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
        update = function(_, entities)
            for _, entity in ipairs(entities) do
                local follow = ecs.get(entity, "dm.world.camera_follow")
                local target = ecs.get(follow:get("target"), "dm.world.position")
                local position = ecs.get(entity, "dm.world.position")
                position:set("x", target:get("x"))
                position:set("y", target:get("y"))
            end
        end,
    })
    local state = { player = player }
    M.bind(state)
    return state
end

-- Bind presentation to a player entity admitted by the authoritative session.
-- This may return false briefly when a scene transition wins the race with the
-- next fixed simulation tick; it never manufactures a second player entity.
function M.bind(state)
    if state.hero and state.hero:exists() then return true end
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
            state.camera = ecs.create({
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
    return position:get("x"), position:get("y")
end

-- Build one value-only HUD snapshot. Live vitals and progression always come
-- from the authoritative ECS entity; save metadata fills only fields that the
-- current simulation schema does not own yet.
function M.hud_snapshot(entity, saved)
    saved = saved or {}
    local progress = assert(ecs.get(entity, "dm.player.progress"))
    local vitals = assert(ecs.get(entity, "dm.player.vitals"))
    return {
        health = vitals:get("health"),
        max_health = vitals:get("max_health"),
        mana = vitals:get("mana"),
        max_mana = vitals:get("max_mana"),
        experience = progress:get("experience"),
        next_level_experience = saved.next_level_experience or 0,
        stamina = saved.stamina or 0,
        max_stamina = saved.max_stamina or 0,
    }
end

function M.destroy(state)
    if not state then return end
    if state.camera and state.camera:exists() then ecs.destroy(state.camera) end
end

return M
