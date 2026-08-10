-- Diablo II player-death presentation. The world remains visible beneath this
-- transparent overlay; death/respawn authority belongs to the game session.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local locale = require("dm.locale/v1")
local text = require("darkmagic.ui.text")
local data = require("dm.data/v1")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = assert(manifest.screens.death)

return {
    blocks_update_below = true,

    enter = function(self)
        self.root = render.create("hud")
        if not render.assets_available() then return end

        -- Font30 matches the tall display lettering in the original death
        -- message. The hardcore epitaph uses the visibly smaller Font16.
        -- Both use Sky/Pal.pl2's red text slot so glyph shading and outlines
        -- remain palette-authored instead of becoming flat RGB modulation.
        text.create(self.root, "death_primary", assert(locale.text("darkmagic.death.died")), screen.x, screen.died_y, screen.width)
        text.create(self.root, "death_primary", assert(locale.text("darkmagic.death.continue")), screen.x, screen.continue_y, screen.width)
        text.create(self.root, "death_secondary", assert(locale.text("darkmagic.death.hardcore")), screen.x, screen.hardcore_y, screen.width)
    end,

    update = function()
        if input.pressed("cancel") or input.pressed("confirm") then
            -- This is presentation-shell behavior only. A bound game session
            -- will replace this pop with its authoritative death transition.
            scenes.pop()
        end
    end,
}
