-- Presentation-facing item API for the first-party mod.
--
-- Reads come from Lua-owned ECS state. Writes still travel through the generic
-- fixed-tick request mailbox, so a screen cannot mutate authority directly.

local requests=require("engine.items/v1")
local snapshots=require("d2legacy.items.snapshot")
local M={api=1}

function M.snapshot() return snapshots.local_player() end
M.move=requests.move
M.select_weapon_set=requests.select_weapon_set
M.sell_held=requests.sell_held
M.buy_to_held=requests.buy_to_held
M.complete_service=requests.complete_service
return M
