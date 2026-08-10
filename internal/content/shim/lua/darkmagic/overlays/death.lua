-- Diablo II player-death presentation. The world remains visible beneath this
-- transparent overlay; death/respawn authority belongs to the game session.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local locale = require("dm.locale/v1")
local text = require("darkmagic.ui.text")

return {
    blocks_update_below = true,

    enter = function(self)
        self.root = render.create("hud")
        if not render.assets_available() then return end

        -- Font30 matches the tall display lettering in the original death
        -- message. The hardcore epitaph uses the visibly smaller Font16.
        -- Both use Sky/Pal.pl2's red text slot so glyph shading and outlines
        -- remain palette-authored instead of becoming flat RGB modulation.
        text.create(self.root, "death_primary", assert(locale.text("darkmagic.death.died")), 400, 145, 760)
        text.create(self.root, "death_primary", assert(locale.text("darkmagic.death.continue")), 400, 195, 760)
        text.create(self.root, "death_secondary", assert(locale.text("darkmagic.death.hardcore")), 400, 255, 760)
    end,

    update = function()
        if input.pressed("cancel") or input.pressed("confirm") then
            -- This is presentation-shell behavior only. A bound game session
            -- will replace this pop with its authoritative death transition.
            scenes.pop()
        end
    end,
}
