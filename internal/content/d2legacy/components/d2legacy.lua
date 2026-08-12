-- Managed entry point for the canonical first-party gameplay mod.
-- All rules remain in lua/d2legacy; discovery sees this tiny lifecycle wrapper.

local authoritative = require("d2legacy.authoritative")

return {
    id = "d2legacy.authoritative",
    api = 1,
    start = function() authoritative.start() end,
    stop = function() authoritative.stop() end,
}
