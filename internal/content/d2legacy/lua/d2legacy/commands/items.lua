-- Item-command composition root.
--
-- Keeping this tiny makes the full policy inventory visible without mixing
-- unrelated transactions into one large file.

local movement = require("d2legacy.commands.item_movement")
local vendor = require("d2legacy.commands.item_vendor")
local services = require("d2legacy.commands.item_services")
local M = {}

function M.register()
    movement.register()
    vendor.register()
    services.register()
end

return M
