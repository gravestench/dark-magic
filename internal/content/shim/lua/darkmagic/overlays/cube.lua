-- Horadric Cube overlay.
--
-- This file is intentionally tiny because the Cube currently fits the reusable
-- fixed-panel pattern. Tiny adapter modules are useful living documentation:
-- they show that a mod author should compose an existing helper instead of
-- copying hundreds of lines just to change one panel ID.

-- Ordinary Lua helper from the shim.
local fixed = require("darkmagic.ui.fixed_panel")

-- Build the scene from the `cube` presentation manifest entry and place it in
-- the left gameplay-overlay slot. The helper will construct the render nodes,
-- item grid, close button, input manager, and lifecycle callbacks.
return fixed.overlay("cube", "left")
