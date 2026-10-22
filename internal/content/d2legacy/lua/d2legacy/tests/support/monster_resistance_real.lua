-- Exercise the production MonStats loader and spawn materializer with rows
-- supplied by the owned-MPQ Go test.

local M = {}

function M.check(id, expected)
    local definition = require("d2legacy.data.monster").load(id)
    local command = {
        tick = 1,
        payload = {
            spawn_id = id,
            seed = 1,
            x = 0,
            y = 0,
            act = 1,
            level_id = 1,
            definition = definition,
        },
    }
    local components = require("d2legacy.commands.spawn_monster").components(command, 1)
    local defense = components["d2legacy.combat.defense"]
    assert(defense, "spawned monster lacks authoritative defense")
    for name, value in pairs(expected) do
        assert(defense[name] == value, "spawned " .. name .. " differs from owned MonStats")
    end
end

return M
