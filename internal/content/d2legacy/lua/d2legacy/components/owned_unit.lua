-- Durable relationship shared by summons, pets, traps, minions, and hirelings.
--
-- An owned unit remains an ordinary entity with its own AI, combat facts, and
-- presentation. This component records only who owns it and which Diablo
-- lifetime policy applies. Explicit ownership lets combat credit the player
-- behind a minion without guessing from proximity or appearance.

local ecs = require("engine.ecs/v1")
local M = {}

function M.register()
    ecs.component({name="d2legacy.owned_unit", fields={
        {name="owner",type="entity"},{name="owner_id",type="string"},
        {name="ultimate_owner_id",type="string"},{name="category",type="string"},
        {name="group",type="i64"},{name="limit",type="i64"},
        {name="replacement",type="string"},{name="created_tick",type="i64"},
        {name="expires_tick",type="i64"},{name="durable_id",type="string"},
        {name="durable",type="bool"},{name="unsummon",type="bool"},
        {name="warp_with_owner",type="bool"},{name="range_limited",type="bool"},
        {name="active",type="bool"},{name="survives_owner_death",type="bool"},
    }})
end

return M
