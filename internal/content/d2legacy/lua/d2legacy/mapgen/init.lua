-- Public d2legacy map-generation API. Each strategy lives in its own module so
-- algorithms stay approachable instead of accumulating in a god-file.

local M = {
    preset = require("d2legacy.mapgen.preset").generate,
    maze = require("d2legacy.mapgen.maze").generate,
}

-- These two temporary delegates are deleted as their Lua strategy ports land.
function M.outdoor(...) return require("d2legacy.mapgen.native/v1").outdoor(...) end

return M
