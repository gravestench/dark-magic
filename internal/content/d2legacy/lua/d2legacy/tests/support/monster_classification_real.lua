-- Exercise the production loader and spawn materializer with owned MonStats
-- rows supplied by the local-only Go test.

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
    local flags = assert(components["d2legacy.monster.type_flags"], "spawned monster lacks type flags")
    for name, value in pairs(expected) do
        assert(definition[name] == value, "loaded " .. id .. " " .. name .. " differs from owned MonStats")
        assert(flags[name] == value, "spawned " .. id .. " " .. name .. " differs from owned MonStats")
    end
end

return M
