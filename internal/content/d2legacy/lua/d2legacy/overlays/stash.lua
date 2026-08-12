-- Stash overlay: a deliberately tiny example of composition.
--
-- There is no stash-specific drawing loop here because stash currently uses the
-- same fixed-panel behavior as several other item panels. Reusing that helper is
-- the point: a mod file can be two executable lines when the shared abstraction
-- already expresses the feature.

-- Load an ordinary d2legacy Lua helper (not an engine capability).
local fixed = require("d2legacy.ui.fixed_panel")

-- Ask the helper to build the scene definition for the manifest entry named
-- "stash". The second argument says this panel occupies the LEFT overlay slot.
-- `return` makes that generated scene table the value other code receives from
-- `require("d2legacy.overlays.stash")`.
return fixed.overlay("stash", "left")
