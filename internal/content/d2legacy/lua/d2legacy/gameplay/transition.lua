-- One relocation transaction shared by seams, warps, waypoints, and objects.

local ecs = require("engine.ecs/v1")
local M = {}

local function finite(value)
    return type(value) == "number" and value == value and value ~= math.huge and value ~= -math.huge
end

function M.validate(destination)
    assert(type(destination) == "table", "world transition destination is required")
    for _, name in ipairs({ "level_id", "x", "y", "width", "height" }) do
        assert(finite(destination[name]), "world transition " .. name .. " must be finite")
    end
    assert(destination.width > 0 and destination.height > 0, "world transition bounds must be positive")
    assert(destination.x >= 0 and destination.x < destination.width, "world transition x is outside bounds")
    assert(destination.y >= 0 and destination.y < destination.height, "world transition y is outside bounds")
end

function M.apply(entity, destination)
    M.validate(destination)
    local location = assert(ecs.get(entity, "d2legacy.world.location"))
    local position = assert(ecs.get(entity, "d2legacy.world.position"))
    local bounds = assert(ecs.get(entity, "d2legacy.world.bounds"))
    local velocity = assert(ecs.get(entity, "d2legacy.world.velocity"))
    location:set("level_id", destination.level_id)
    position:set("x", destination.x)
    position:set("y", destination.y)
    bounds:set("width", destination.width)
    bounds:set("height", destination.height)
    velocity:set("x", 0)
    velocity:set("y", 0)
end

return M
