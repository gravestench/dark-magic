-- Small, reusable pieces of spatial world state.
--
-- Coordinates are continuous DS1 subtiles here. Pixel projection belongs to
-- presentation. Keeping Position generic lets the same movement rules later
-- operate on players, monsters, missiles, and other bounded world entities.

local ecs = require("engine.ecs/v1")

local M = {}

function M.register()
    ecs.component({
        name = "d2legacy.world.position",
        fields = {
            { name = "x", type = "f64" },
            { name = "y", type = "f64" },
        },
    })

    ecs.component({
        name = "d2legacy.world.velocity",
        fields = {
            { name = "x", type = "f64" },
            { name = "y", type = "f64" },
        },
    })

    ecs.component({
        name = "d2legacy.world.bounds",
        fields = {
            { name = "width", type = "f64" },
            { name = "height", type = "f64" },
        },
    })

    ecs.component({
        name = "d2legacy.world.collider",
        fields = {
            -- Radius is measured in DS1 subtiles around the entity's center.
            -- Keeping it separate from map bounds avoids one overloaded brick.
            { name = "radius", type = "f64" },
        },
    })

    ecs.component({
        name = "d2legacy.world.forced_motion_request",
        fields = {
            { name = "kind", type = "string" },
            { name = "source_x", type = "f64" },
            { name = "source_y", type = "f64" },
            { name = "distance", type = "f64" },
            { name = "speed", type = "f64" },
            { name = "request_tick", type = "i64" },
        },
    })

    ecs.component({
        name = "d2legacy.world.forced_motion",
        fields = {
            { name = "kind", type = "string" },
            { name = "target_x", type = "f64" },
            { name = "target_y", type = "f64" },
            { name = "speed", type = "f64" },
            { name = "start_x", type = "f64" },
            { name = "start_y", type = "f64" },
            { name = "requested_distance", type = "f64" },
            { name = "applied_distance", type = "f64", default = 0 },
            { name = "start_tick", type = "i64" },
        },
    })

    ecs.component({
        name = "d2legacy.world.forced_motion_event",
        fields = {
            { name = "kind", type = "string" },
            { name = "outcome", type = "string" },
            { name = "tick", type = "i64" },
            { name = "target_id", type = "string" },
            { name = "start_x", type = "f64" },
            { name = "start_y", type = "f64" },
            { name = "end_x", type = "f64" },
            { name = "end_y", type = "f64" },
            { name = "requested_distance", type = "f64" },
            { name = "applied_distance", type = "f64" },
        },
    })

    ecs.component({
        name = "d2legacy.world.player_control",
        fields = {
            -- Logical session player ID, not a native input-device handle.
            { name = "player", type = "string" },
        },
    })

    ecs.component({
        name = "d2legacy.world.camera_follow",
        fields = {
            -- Entity references remain generation-checked ECS handles.
            { name = "target", type = "entity" },
            { name = "strategy", type = "string" },
            { name = "duration", type = "f64" },
            { name = "param_1", type = "f64" },
            { name = "param_2", type = "f64" },
            { name = "param_3", type = "f64" },
            { name = "origin_x", type = "f64" },
            { name = "origin_y", type = "f64" },
            { name = "destination_x", type = "f64" },
            { name = "destination_y", type = "f64" },
            { name = "elapsed", type = "f64" },
        },
    })
end

return M
