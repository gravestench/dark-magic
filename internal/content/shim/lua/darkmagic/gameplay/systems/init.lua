-- Register the world systems in their intended execution order.
--
-- Phases still provide the real scheduler ordering. Listing the registrations
-- here is useful documentation and prevents world.lua becoming a behavior dump.

local movement = require("darkmagic.gameplay.systems.movement")
local camera_follow = require("darkmagic.gameplay.systems.camera_follow")

local M = {}

function M.register(collision)
    movement.register(collision)
    camera_follow.register()
end

return M
