-- Copy a followed entity's world position into a presentation camera entity.
--
-- The camera is disposable presentation state. The target is authoritative
-- world state. This one-way copy keeps camera concerns out of the player entity.

local ecs = require("engine.ecs/v1")

local M = {}

local function curve(strategy, value, parameter)
    value = math.max(0, math.min(1, value))
    if strategy == "linear" then return value end
    if strategy == "quad_in" then return value * value end
    if strategy == "quad_out" then return 1 - (1-value) * (1-value) end
    if strategy == "quad_in_out" then
        if value < 0.5 then return 2 * value * value end
        return 1 - ((-2 * value + 2)^2) / 2
    end
    if strategy == "cubic_in" then return value^3 end
    if strategy == "cubic_out" then return 1 - (1-value)^3 end
    if strategy == "cubic_in_out" then
        if value < 0.5 then return 4 * value^3 end
        return 1 - ((-2 * value + 2)^3) / 2
    end
    if strategy == "exponential_out" then
        if value == 1 then return 1 end
        local exponent = parameter ~= 0 and parameter or 10
        return 1 - 2^(-exponent * value)
    end
    if strategy == "back_out" then
        local overshoot = parameter ~= 0 and parameter or 1.70158
        local shifted = value - 1
        return 1 + (overshoot + 1) * shifted^3 + overshoot * shifted^2
    end
    error("unknown camera follow strategy " .. tostring(strategy))
end

local function follow(entity, elapsed)
    local relationship = ecs.get(entity, "d2legacy.world.camera_follow")
    local target_position = ecs.get(relationship:get("target"), "d2legacy.world.position")
    local camera_position = ecs.get(entity, "d2legacy.world.position")
    local target_x, target_y = target_position:get("x"), target_position:get("y")
    local strategy, duration = relationship:get("strategy"), relationship:get("duration")
    if strategy == "instant" or duration <= 0 then
        camera_position:set("x", target_x); camera_position:set("y", target_y)
        relationship:set("origin_x", target_x); relationship:set("origin_y", target_y)
        relationship:set("destination_x", target_x); relationship:set("destination_y", target_y)
        relationship:set("elapsed", 0)
        return
    end
    if target_x ~= relationship:get("destination_x") or target_y ~= relationship:get("destination_y") then
        relationship:set("origin_x", camera_position:get("x"))
        relationship:set("origin_y", camera_position:get("y"))
        relationship:set("destination_x", target_x)
        relationship:set("destination_y", target_y)
        relationship:set("elapsed", 0)
    end
    local time = math.min(relationship:get("elapsed") + math.max(elapsed, 0), duration)
    relationship:set("elapsed", time)
    local amount = curve(strategy, time / duration, relationship:get("param_1"))
    camera_position:set("x", relationship:get("origin_x") + (relationship:get("destination_x") - relationship:get("origin_x")) * amount)
    camera_position:set("y", relationship:get("origin_y") + (relationship:get("destination_y") - relationship:get("origin_y")) * amount)
end

function M.register()
    ecs.system({
        id = "d2legacy.world.camera_follow",
        phase = "presentation",
        query = { all = { "d2legacy.world.position", "d2legacy.world.camera_follow" } },
        read = { "d2legacy.world.camera_follow" },
        write = { "d2legacy.world.position", "d2legacy.world.camera_follow" },
        update = function(context, entities)
            for _, entity in ipairs(entities) do follow(entity, context.delta_seconds) end
        end,
    })
end

return M
