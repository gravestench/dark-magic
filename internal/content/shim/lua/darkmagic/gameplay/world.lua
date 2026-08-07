-- ECS-owned world motion used by the game-world scene.
--
-- Coordinates are continuous DS1 subtiles. Presentation projects them through
-- dm.world/v1, so collision and rendering never share ambiguous pixel values.
local ecs = require("dm.ecs/v1")

local M = {}

local function define_components()
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
    local hero = ecs.create({
        ["dm.world.position"] = { x = width / 2, y = height / 2 },
        ["dm.world.velocity"] = {},
        ["dm.world.player_control"] = { player = player },
        ["dm.world.bounds"] = { width = width, height = height },
    })
    local camera = ecs.create({
        ["dm.world.position"] = { x = width / 2, y = height / 2 },
        ["dm.world.camera_follow"] = { target = hero },
    })

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
    return { hero = hero, camera = camera }
end

function M.position(entity)
    local position = assert(ecs.get(entity, "dm.world.position"))
    return position:get("x"), position:get("y")
end

function M.destroy(state)
    if not state then return end
    if state.camera and state.camera:exists() then ecs.destroy(state.camera) end
    if state.hero and state.hero:exists() then ecs.destroy(state.hero) end
end

return M
