-- ECS-owned world motion used by the game-world scene.
--
-- Coordinates remain presentation pixels until the DS1 isometric transform is
-- admitted explicitly. Collision must not mix these values with subtile indices.
local ecs = require("dm.ecs/v1")
local input = require("dm.input/v1")

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
    ecs.component({ name = "dm.world.player_control", fields = {} })
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

function M.create(width, height)
    define_components()
    local hero = ecs.create({
        ["dm.world.position"] = { x = width / 2, y = height / 2 },
        ["dm.world.velocity"] = {},
        ["dm.world.player_control"] = {},
        ["dm.world.bounds"] = { width = width, height = height },
    })
    local camera = ecs.create({
        ["dm.world.position"] = { x = width / 2, y = height / 2 },
        ["dm.world.camera_follow"] = { target = hero },
    })

    ecs.system({
        id = "darkmagic.world.player_intent",
        phase = "input",
        query = { all = { "dm.world.player_control", "dm.world.velocity" } },
        write = { "dm.world.velocity" },
        update = function(_, entities)
            for _, entity in ipairs(entities) do
                local velocity = ecs.get(entity, "dm.world.velocity")
                local x, y = 0, 0
                if input.down("left") then x = x - 160 end
                if input.down("right") then x = x + 160 end
                if input.down("up") then y = y - 160 end
                if input.down("down") then y = y + 160 end
                velocity:set("x", x)
                velocity:set("y", y)
            end
        end,
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
                position:set("x", clamp(position:get("x") + velocity:get("x") * context.delta_seconds, 0, bounds:get("width")))
                position:set("y", clamp(position:get("y") + velocity:get("y") * context.delta_seconds, 0, bounds:get("height")))
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
