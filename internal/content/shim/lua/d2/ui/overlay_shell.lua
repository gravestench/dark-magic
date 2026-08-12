-- Independently bootable shell for state-dependent in-game interfaces.
--
-- A shell is a TEMPORARY/GENERIC presentation for a feature whose real gameplay
-- or session capability is not wired yet. It is useful because the rest of the
-- scene/navigation system can already open, position, test, and close the panel.
--
-- Most importantly, this file does NOT fake missing gameplay authority. It draws
-- a status message and leaves unavailable operations unavailable until the real
-- engine capability exists.

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local data = require("engine.data/v1")
local locale = require("engine.locale/v1")
local controls = require("d2.ui.controls")
local button = require("d2.ui.button")
local text = require("d2.ui.text")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2.presentation/v1"))
local M = {}

-- Factory: turn one plain definition table into a full scene-definition table.
function M.new(definition)
    return {
        -- Omitted means true here. Only explicit false lets lower scenes continue.
        blocks_update_below = definition.blocks_update_below ~= false,
        -- Omitted means false. Only explicit true passes routed input below.
        passes_input_below = definition.passes_input_below == true,
        world_view = definition.world_view,

        enter = function(self)
            self.root = render.create(definition.layer or "modal")
            self.controls = controls.new()

            -- Default geometry makes even a very small shell definition useful.
            local x, y = definition.x or 160, definition.y or 100
            local width, height = definition.width or 480, definition.height or 360

            if definition.sheet and render.assets_available() then
                -- Use authored multi-frame panel art when supplied.
                self.panel = render.create(definition.layer or "modal", self.root)
                width, height = self.panel:set_dc6_combined(
                    definition.sheet,
                    manifest.palettes[definition.palette or "sky"],
                    definition.direction or 0,
                    definition.page or 0)
                self.panel:set_position(x + width / 2, y + height / 2)
            else
                -- Otherwise a simple rectangle keeps the shell visible in labs,
                -- incomplete content packs, or early feature development.
                self.panel = render.create(definition.layer or "modal", self.root)
                self.panel:fill_rect(width, height, 18, 14, 10, definition.alpha or 235)
                self.panel:set_position(x + width / 2, y + height / 2)
            end

            -- Headless tests still exercise lifecycle/control structure. They do
            -- not need proprietary bitmap fonts/button art below this point.
            if not render.assets_available() then return end

            text.create(self.root, "panel_heading", assert(locale.text(definition.title)),
                x + width / 2, y + 24, width - 40)

            -- Explicitly tell the player/developer this is a shell rather than
            -- pretending an unavailable gameplay operation succeeded.
            text.create(self.root, "disabled", assert(locale.text(definition.status or "d2.overlay.shell_status")),
                x + width / 2, y + height / 2, width - 60)

            local close = {
                sheet="data/global/ui/PANEL/buysellbtn.DC6", palette="sky",
                up_frame=10, down_frame=11, x=x + width - 48, y=y + height - 48,
                width=32, height=32, label="d2.overlay.close",
            }

            button.create(self.root, self.controls, "close", close, assert(locale.text(close.label)), {
                layer=definition.layer or "modal", show_label=false, sound=manifest.sounds.button,
                tooltip=assert(locale.text(close.label)), on_activate=function() scenes.pop() end,
            })
        end,

        update = function(self)
            self.controls:update()
            -- Standard Escape/Cancel stack close path.
            if input.pressed("cancel") then scenes.pop() end
        end,
    }
end

return M
