-- Presentation entry point for a new mod.
--
-- Keep this definition small. Once the mod has presentation, require a
-- namespace-owned bootstrap module here, register scenes in start(), and let
-- those scene modules own their retained resources.

local mod = require("mod_template")

return {
    id = "mod_template.boot",
    api = 1,

    start = function(self)
        self.mod_id = mod.id
    end,

    stop = function(self)
        self.mod_id = nil
    end,
}
