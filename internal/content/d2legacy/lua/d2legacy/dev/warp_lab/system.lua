-- Resolve Warp Lab interaction intents on the deterministic ECS clock.
--
-- One small system demonstrates the whole transaction in readable order:
-- find the requested portal, walk toward it, enter at interaction range, find
-- its declared pair, then publish the new authoritative position atomically.

local ecs = require("engine.ecs/v1")
local M = {}

local function logical_direction(dx, dy)
    local x = dx == 0 and 0 or (dx > 0 and 1 or -1)
    local y = dy == 0 and 0 or (dy > 0 and 1 or -1)
    if x == 0 and y == 1 then return 0 end
    if x == -1 and y == 0 then return 1 end
    if x == 0 and y == -1 then return 2 end
    if x == 1 and y == 0 then return 3 end
    if x == 1 and y == 1 then return 4 end
    if x == -1 and y == 1 then return 5 end
    if x == -1 and y == -1 then return 6 end
    return 7
end

local function move_toward(position, actor, target, elapsed)
    local dx = target.x - position:get("x")
    local dy = target.y - position:get("y")
    local distance = math.sqrt(dx * dx + dy * dy)
    if distance == 0 then return 0 end
    actor:set("direction", logical_direction(dx, dy))
    local step = math.min(distance, actor:get("speed") * elapsed)
    position:set("x", position:get("x") + dx / distance * step)
    position:set("y", position:get("y") + dy / distance * step)
    return distance - step
end

-- Routes are tiny immutable strings stored inside the intent component. This
-- keeps the deterministic waypoints in authoritative ECS state instead of a
-- presentation-owned Lua table that would disappear during replay or reload.
local function route_waypoint(route, wanted)
    local index = 0
    for x, y in route:gmatch("([^,;]+),([^;]+)") do
        index = index + 1
        if index == wanted then return {x=tonumber(x), y=tonumber(y)} end
    end
    return nil
end

local function follow_route(position, actor, intent, fallback, elapsed)
    local index = intent:get("waypoint")
    local target = route_waypoint(intent:get("route"), index) or fallback
    local remaining = move_toward(position, actor, target, elapsed)
    if remaining <= 0.1 and route_waypoint(intent:get("route"), index + 1) then
        intent:set("waypoint", index + 1)
        return math.huge
    end
    return remaining
end

function M.register()
    ecs.system({
        id = "d2legacy.lab.warp.resolve_move", phase = "movement",
        query = {all = {"d2legacy.lab.warp.actor", "d2legacy.lab.warp.move_intent", "d2legacy.world.position"}},
        read = {"d2legacy.lab.warp.actor", "d2legacy.lab.warp.move_intent"},
        write = {"d2legacy.world.position", "d2legacy.lab.warp.actor", "d2legacy.lab.warp.move_intent"},
        update = function(context, entities, commands)
            for _, entity in ipairs(entities) do
                local actor = ecs.get(entity, "d2legacy.lab.warp.actor")
                local intent = ecs.get(entity, "d2legacy.lab.warp.move_intent")
                local position = ecs.get(entity, "d2legacy.world.position")
                local remaining = follow_route(position, actor, intent, {
                    x = intent:get("x"), y = intent:get("y"),
                }, context.delta_seconds)
                if remaining <= 0.1 then
                    commands:remove(entity, "d2legacy.lab.warp.move_intent")
                end
            end
        end,
    })
    ecs.system({
        id="d2legacy.lab.warp.resolve_intent", phase="movement",
        query={all={"d2legacy.lab.warp.actor","d2legacy.lab.warp.intent","d2legacy.lab.warp.state","d2legacy.world.position"}},
        read={"d2legacy.lab.warp.actor","d2legacy.lab.warp.intent","d2legacy.lab.warp.portal"},
        write={"d2legacy.lab.warp.state","d2legacy.world.position","d2legacy.lab.warp.actor","d2legacy.lab.warp.intent"},
        update=function(context,entities,commands)
            for _, entity in ipairs(entities) do
                local actor = ecs.get(entity, "d2legacy.lab.warp.actor")
                local intent = ecs.get(entity, "d2legacy.lab.warp.intent")
                local position = ecs.get(entity, "d2legacy.world.position")
                local state = ecs.get(entity, "d2legacy.lab.warp.state")
                local target_entity = intent:get("portal")
                local target_portal = ecs.get(target_entity, "d2legacy.lab.warp.portal")
                local target_position = ecs.get(target_entity, "d2legacy.world.position")
                if not target_portal or not target_position then
                    state:set("event", "intent rejected: unknown portal")
                    commands:remove(entity, "d2legacy.lab.warp.intent")
                else
                    state:set("event", "walking to " .. target_portal:get("label"))
                    follow_route(position, actor, intent, {
                        x = target_position:get("x"),
                        y = target_position:get("y"),
                    }, context.delta_seconds)
                    local dx = position:get("x") - target_position:get("x")
                    local dy = position:get("y") - target_position:get("y")
                    if math.sqrt(dx * dx + dy * dy) <= target_portal:get("radius") then
                        local destination_entity = target_portal:get("pair")
                        local destination_portal = ecs.get(destination_entity, "d2legacy.lab.warp.portal")
                        local destination_position = ecs.get(destination_entity, "d2legacy.world.position")
                        if not destination_portal or not destination_position then
                            state:set("event", "intent rejected: unpaired portal")
                        else
                            -- Arrive just beside the paired portal so another
                            -- click is required to travel back.
                            position:set("x", destination_position:get("x") + 2)
                            position:set("y", destination_position:get("y") + 2)
                            state:set("warp_count", state:get("warp_count") + 1)
                            state:set("event", "arrived through " .. destination_portal:get("label"))
                        end
                        commands:remove(entity, "d2legacy.lab.warp.intent")
                    end
                end
            end
        end,
    })
end

return M
