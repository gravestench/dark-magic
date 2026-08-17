-- Admit semantic forced-movement requests before the movement phase.
--
-- Combat decides whether an effect exists and supplies its effect-specific
-- distance/speed policy. Movement owns the spatial transaction: direction is
-- derived away from the source, progress is checkpointed, and collision later
-- decides whether the request completes, stops short, or is fully blocked.

local ecs = require("engine.ecs/v1")
local entity_identity = require("d2legacy.policy.entity_identity")

local M = {}

local function emit_rejection(context, entity, request, position, outcome, structural)
    structural:create({
        ["d2legacy.world.forced_motion_event"] = {
            kind = request:get("kind"),
            outcome = outcome,
            tick = context.tick,
            target_id = entity_identity.semantic_id(entity),
            start_x = position:get("x"),
            start_y = position:get("y"),
            end_x = position:get("x"),
            end_y = position:get("y"),
            requested_distance = request:get("distance"),
            applied_distance = 0,
        },
    })
end

local function emit_replacement(context, entity, active, position, structural)
    structural:create({
        ["d2legacy.world.forced_motion_event"] = {
            kind = active:get("kind"),
            outcome = "replaced",
            tick = context.tick,
            target_id = entity_identity.semantic_id(entity),
            start_x = active:get("start_x"),
            start_y = active:get("start_y"),
            end_x = position:get("x"),
            end_y = position:get("y"),
            requested_distance = active:get("requested_distance"),
            applied_distance = active:get("applied_distance"),
        },
    })
end

local function admit(context, entity, request, position, active, structural)
    assert(request:get("request_tick") <= context.tick, "future forced-motion request")
    local distance, speed = request:get("distance"), request:get("speed")
    local dx = position:get("x") - request:get("source_x")
    local dy = position:get("y") - request:get("source_y")
    local major = math.max(math.abs(dx), math.abs(dy))
    if distance <= 0 or speed <= 0 or major == 0 then
        emit_rejection(context, entity, request, position, "invalid", structural)
        structural:remove(entity, "d2legacy.world.forced_motion_request")
        return
    end

    -- Diablo's knockback destination extends the source-to-target ray by a
    -- major-axis distance. Keeping distance in the request leaves exact target
    -- policy outside this generic spatial mechanism.
    local scale = distance / major
    if active then
        -- A target can own only one spatial transaction. Preserve the existing
        -- deterministic replacement behavior as an explicit semantic outcome
        -- before the newer admitted request becomes authoritative.
        emit_replacement(context, entity, active, position, structural)
    end
    structural:set(entity, "d2legacy.world.forced_motion", {
        kind = request:get("kind"),
        target_x = position:get("x") + dx * scale,
        target_y = position:get("y") + dy * scale,
        speed = speed,
        start_x = position:get("x"),
        start_y = position:get("y"),
        requested_distance = distance,
        applied_distance = 0,
        start_tick = context.tick,
    })
    structural:remove(entity, "d2legacy.world.forced_motion_request")
end

function M.register()
    ecs.system({
        id = "d2legacy.world.admit_forced_motion",
        phase = "pre_simulation",
        query = {
            all = {
                "d2legacy.world.forced_motion_request",
                "d2legacy.world.position",
                "d2legacy.world.velocity",
                "d2legacy.world.bounds",
                "d2legacy.world.collider",
                "d2legacy.world.occupancy",
            },
        },
        read = {
            "d2legacy.world.forced_motion_request",
            "d2legacy.world.position",
            "d2legacy.world.velocity",
            "d2legacy.world.bounds",
            "d2legacy.world.collider",
            "d2legacy.world.occupancy",
            "d2legacy.world.forced_motion",
            "d2legacy.world.selectable",
        },
        write = {
            "d2legacy.world.forced_motion_request",
            "d2legacy.world.forced_motion",
            "d2legacy.world.forced_motion_event",
        },
        update = function(context, entities, structural)
            for _, entity in ipairs(entities) do
                admit(
                    context,
                    entity,
                    ecs.get(entity, "d2legacy.world.forced_motion_request"),
                    ecs.get(entity, "d2legacy.world.position"),
                    ecs.get(entity, "d2legacy.world.forced_motion"),
                    structural
                )
            end
        end,
    })
end

return M
