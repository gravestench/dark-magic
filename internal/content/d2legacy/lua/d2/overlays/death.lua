-- Diablo II player-death PRESENTATION overlay.
--
-- This file is intentionally NOT a death/respawn system. It draws the message
-- and reacts to presentation input. Whether a character is dead, whether they
-- may respawn, where they respawn, hardcore permanence, save/network changes,
-- etc. belong to authoritative game/session code.
--
-- That boundary is important for mods: a panel saying "You have died" should
-- never secretly become the authority that decides the player is dead.

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local locale = require("engine.locale/v1")
local text = require("d2.ui.text")
local data = require("engine.data/v1")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2.presentation/v1"))
local screen = assert(manifest.screens.death)

return {
    -- Current shell behavior blocks the world below while the death surface owns
    -- the interaction. The real authoritative death transition can refine this.
    blocks_update_below = true,

    enter = function(self)
        -- HUD layer keeps the world visually available beneath transparent text.
        self.root = render.create("hud")

        -- Headless tests can still exercise scene lifecycle without bitmap assets.
        if not render.assets_available() then return end

        -- Font30 matches the tall display lettering in the original death
        -- message. The hardcore epitaph uses the visibly smaller Font16.
        -- Both use Sky/Pal.pl2's red text slot so glyph shading and outlines
        -- remain palette-authored instead of becoming flat RGB modulation.
        text.create(self.root, "death_primary", assert(locale.text("d2.death.died")), screen.x, screen.died_y, screen.width)
        text.create(self.root, "death_primary", assert(locale.text("d2.death.continue")), screen.x, screen.continue_y, screen.width)
        text.create(self.root, "death_secondary", assert(locale.text("d2.death.hardcore")), screen.x, screen.hardcore_y, screen.width)
    end,

    update = function()
        if input.pressed("cancel") or input.pressed("confirm") then
            -- This is presentation-shell behavior only. A bound game session
            -- will replace this simple pop with its authoritative death/continue
            -- command or transition. Lua does not mutate player life state here.
            scenes.pop()
        end
    end,
}
