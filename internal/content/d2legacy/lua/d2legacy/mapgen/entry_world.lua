-- Compose the first playable d2legacy world without leaking level identities
-- into the generic host. Rogue Encampment and Blood Moor are mod choices; the
-- Go adapter only transports the two admitted recipes returned here.

local preset = require("d2legacy.mapgen.preset")
local outdoor = require("d2legacy.mapgen.outdoor")
local M = {}

function M.town(seed, difficulty)
    return preset.generate(1, seed, difficulty)
end

function M.wilderness(seed, difficulty, town)
    local stamp = assert(town.stamps and town.stamps[1], "d2legacy entry town has no stamp")
    local direction = assert(stamp.role:match("^act1%-town:exit%-(%a+)$"),
        "d2legacy entry town has no cardinal exit")
    return outdoor.generate(2, seed, direction, difficulty)
end

return M
