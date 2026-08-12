-- Hireling equipment overlay.
--
-- Like stash.lua and cube.lua, this is an adapter around the reusable
-- fixed-panel helper. The important lesson is that a feature-specific module
-- does not need feature-specific machinery when shared machinery already fits.

local fixed = require("d2.ui.fixed_panel")

-- The manifest entry named `hireling` supplies the art/layout/item-slot facts.
-- The helper supplies behavior. Returning the result exports the completed scene
-- definition to the registry.
return fixed.overlay("hireling", "left")
