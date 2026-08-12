-- Small Diablo-policy decisions for the first melee migration.
-- The fixed chance intentionally preserves the old scaffold until the reviewed
-- attack-rating/defense inputs are connected to the verified 5..95% formula.

local random = require("dm.authority_random/v1")
local M = { temporary_hit_chance = 75 }

function M.hits()
    return random.integer("d2legacy.combat.basic_melee.hit", 100)
        < M.temporary_hit_chance
end

function M.damage(minimum, maximum)
    assert(minimum >= 0 and maximum >= minimum, "invalid melee damage range")
    return minimum + random.integer(
        "d2legacy.combat.basic_melee.damage", maximum - minimum + 1)
end

return M
