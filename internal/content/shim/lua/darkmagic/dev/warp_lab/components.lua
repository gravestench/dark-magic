-- Data used only by Warp Lab's tiny authoritative fixture.
--
-- The names say what each fact means. Presentation never stores a destination
-- node pointer or moves the player itself; it adds an intent and observes the
-- fixed-tick system changing ordinary ECS state.

local ecs = require("dm.ecs/v1")
local M = {}

local function field(name, kind) return {name=name, type=kind} end

function M.register()
    -- Re-registering an identical schema is safe and keeps this lab independent
    -- from whichever gameplay scene happened to run first.
    ecs.component({name="dm.world.position", version=1, fields={field("x","f64"),field("y","f64")}})
    ecs.component({name="dm.lab.warp.portal", version=1, fields={
        field("id","string"), field("pair","entity"), field("label","string"), field("radius","f64"),
    }})
    ecs.component({name="dm.lab.warp.intent", version=1, fields={field("portal","entity")}})
    ecs.component({name="dm.lab.warp.move_intent", version=1, fields={
        field("x", "f64"), field("y", "f64"),
    }})
    ecs.component({name="dm.lab.warp.actor", version=1, fields={field("speed","f64")}})
    ecs.component({name="dm.lab.warp.state", version=1, fields={
        field("event","string"), field("warp_count","i64"),
    }})
end

return M
