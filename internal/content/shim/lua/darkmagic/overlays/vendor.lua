-- Vendor panel adapter.
--
-- The current vendor presentation uses the shared fixed-panel shell. Actual
-- buying/selling authority is intentionally NOT implemented by mutating items
-- here; transactions belong to gameplay/session capabilities.

local fixed = require("darkmagic.ui.fixed_panel")

-- Reuse the generic panel composition for the `vendor` manifest entry. This is
-- presentation wiring only, which keeps future vendor transaction rules out of
-- the UI layer.
return fixed.overlay("vendor", "left")
