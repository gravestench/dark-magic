-- Waypoint overlay adapter.
--
-- This panel currently needs the same generic fixed-panel behavior as several
-- other interfaces. Keeping this file tiny makes the extension point obvious:
-- when waypoint-specific behavior is needed later, this module can grow without
-- changing every place that already opens the scene ID `waypoint`.

local fixed = require("d2legacy.ui.fixed_panel")

-- Build and export the left-slot fixed panel described by the waypoint manifest.
return fixed.overlay("waypoint", "left")
