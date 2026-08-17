-- Semantic ownership of authoritative player locomotion.
--
-- Route planning is a client concern. Commands deliver a direction or the next
-- admitted waypoint; combat delivers an approach target. Both become one
-- explicit authority fact before the resolver writes raw execution velocity.

local ecs = require("engine.ecs/v1")
local movement_rules = require("d2legacy.movement_rules/v1")
local M = {}

local function state(entity)
    return assert(ecs.get(entity, "d2legacy.player.motion"), "player has no motion state")
end

function M.locomotion(entity, payload)
    local motion = state(entity)
    local target = payload.target
    motion:set("owner", "locomotion")
    motion:set("kind", target and "waypoint" or "direction")
    motion:set("x", payload.x)
    motion:set("y", payload.y)
    motion:set("target_x", target and target.x or 0)
    motion:set("target_y", target and target.y or 0)
    motion:set("has_target", target ~= nil)
    motion:set("running", payload.running)
    motion:set("active", target ~= nil or payload.x ~= 0 or payload.y ~= 0)
end

function M.approach(entity, target_x, target_y)
    local motion = state(entity)
    motion:set("owner", "attack_approach")
    motion:set("kind", "target")
    motion:set("x", 0)
    motion:set("y", 0)
    motion:set("target_x", target_x)
    motion:set("target_y", target_y)
    motion:set("has_target", true)
    motion:set("running", false)
    motion:set("active", true)
end

function M.stop(entity)
    local motion = ecs.get(entity, "d2legacy.player.motion")
    if not motion then
        return
    end
    motion:set("owner", "none")
    motion:set("kind", "idle")
    motion:set("x", 0)
    motion:set("y", 0)
    motion:set("has_target", false)
    motion:set("running", false)
    motion:set("active", false)
end

-- Exhaustion is an execution correction inside the same motion boundary. The
-- stamina system decides when it occurs; this helper owns how semantic run
-- intent, velocity, mode, and animation change together on that tick.
function M.force_walk(entity, tick)
    local motion = state(entity)
    if motion:get("owner") == "locomotion" then
        motion:set("running", false)
    end
    local identity = assert(ecs.get(entity, "d2legacy.player.identity"))
    local stats = assert(ecs.get(entity, "d2legacy.player.movement_stats"))
    local velocity = assert(ecs.get(entity, "d2legacy.world.velocity"))
    local walk =
        movement_rules.rates(identity:get("class"), stats:get("velocitypercent"), stats:get("item_fastermovevelocity"))
    local x, y = velocity:get("x"), velocity:get("y")
    local magnitude = math.sqrt(x * x + y * y)
    if magnitude > 0 then
        velocity:set("x", x / magnitude * walk)
        velocity:set("y", y / magnitude * walk)
    end
    local mode = ecs.get(entity, "d2legacy.player.movement_mode")
    if mode then
        mode:set("running", false)
    end
    local animation = ecs.get(entity, "d2legacy.player.animation")
    if animation and animation:get("mode") == "RN" then
        animation:set("mode", "WL")
        animation:set("start_tick", tick)
    end
end

return M
