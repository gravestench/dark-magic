-- Small, reusable pieces of spatial world state.
--
-- Coordinates are continuous DS1 subtiles here. Pixel projection belongs to
-- presentation. Keeping Position generic lets the same movement rules later
-- operate on players, monsters, missiles, and other bounded world entities.

local ecs = require("dm.ecs/v1")

local M = {}

function M.register()
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
        name = "dm.world.bounds",
        fields = {
            { name = "width", type = "f64" },
            { name = "height", type = "f64" },
        },
    })

    ecs.component({
        name = "dm.world.collider",
        fields = {
            -- Radius is measured in DS1 subtiles around the entity's center.
            -- Keeping it separate from map bounds avoids one overloaded brick.
            { name = "radius", type = "f64" },
        },
    })

    ecs.component({
        name = "dm.world.player_control",
        fields = {
            -- Logical session player ID, not a native input-device handle.
            { name = "player", type = "string" },
        },
    })

    ecs.component({
        name = "dm.world.camera_follow",
        fields = {
            -- Entity references remain generation-checked ECS handles.
            { name = "target", type = "entity" },
        },
    })
end

return M
