-- Resolve Warp Lab interaction intents on the deterministic ECS clock.
--
-- One small system demonstrates the whole transaction in readable order:
-- find the requested portal, walk toward it, enter at interaction range, find
-- its declared pair, then publish the new authoritative position atomically.

local ecs = require("dm.ecs/v1")
local M = {}

local function move_toward(position, target, speed, elapsed)
    local dx = target.x - position:get("x")
    local dy = target.y - position:get("y")
    local distance = math.sqrt(dx * dx + dy * dy)
    if distance == 0 then return 0 end
    local step = math.min(distance, speed * elapsed)
    position:set("x", position:get("x") + dx / distance * step)
    position:set("y", position:get("y") + dy / distance * step)
    return distance - step
end

function M.register()
    ecs.system({
        id="darkmagic.lab.warp.resolve_intent", phase="movement",
        query={all={"dm.lab.warp.actor","dm.lab.warp.intent","dm.lab.warp.state","dm.world.position"}},
        read={"dm.lab.warp.actor","dm.lab.warp.intent","dm.lab.warp.portal"},
        write={"dm.lab.warp.state","dm.world.position","dm.lab.warp.intent"},
        update=function(context,entities,commands)
            for _, entity in ipairs(entities) do
                local actor = ecs.get(entity, "dm.lab.warp.actor")
                local intent = ecs.get(entity, "dm.lab.warp.intent")
                local position = ecs.get(entity, "dm.world.position")
                local state = ecs.get(entity, "dm.lab.warp.state")
                local target_entity = intent:get("portal")
                local target_portal = ecs.get(target_entity, "dm.lab.warp.portal")
                local target_position = ecs.get(target_entity, "dm.world.position")
                if not target_portal or not target_position then
                    state:set("event", "intent rejected: unknown portal")
                    commands:remove(entity, "dm.lab.warp.intent")
                else
                    state:set("event", "walking to " .. target_portal:get("label"))
                    local remaining = move_toward(position, {
                        x = target_position:get("x"),
                        y = target_position:get("y"),
                    }, actor:get("speed"), context.delta_seconds)
                    if remaining <= target_portal:get("radius") then
                        local destination_entity = target_portal:get("pair")
                        local destination_portal = ecs.get(destination_entity, "dm.lab.warp.portal")
                        local destination_position = ecs.get(destination_entity, "dm.world.position")
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
                        commands:remove(entity, "dm.lab.warp.intent")
                    end
                end
            end
        end,
    })
end

return M
