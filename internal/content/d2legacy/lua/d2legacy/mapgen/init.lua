-- Public d2legacy map-generation API. Each strategy lives in its own module so
-- algorithms stay approachable instead of accumulating in a god-file.

local M = {preset = require("d2legacy.mapgen.preset").generate}

-- These two temporary delegates are deleted as their Lua strategy ports land.
function M.maze(...) return require("d2legacy.mapgen.native/v1").maze(...) end
function M.outdoor(...) return require("d2legacy.mapgen.native/v1").outdoor(...) end

return M
