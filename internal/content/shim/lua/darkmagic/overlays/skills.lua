-- Class-specific expansion skill-tree shell at canonical desktop geometry.
--
-- This is intentionally still a PRESENTATION SHELL. It proves that the correct
-- class-specific panel art, scene lifecycle, close control, and routing exist;
-- skill allocation/gameplay authority should be added through a capability, not
-- by letting this panel directly invent or mutate learned skills.

local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local saves = require("dm.save/v1")
local data = require("dm.data/v1")
local locale = require("dm.locale/v1")
local controls = require("darkmagic.ui.controls")
local button = require("darkmagic.ui.button")
local text = require("darkmagic.ui.text")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.skills

-- Selected character class determines which original skill-tree background file
-- is used. Keeping the map as data is clearer than a seven-branch if statement.
local backgrounds = {
    Amazon="skltree_a_back", Sorceress="skltree_s_back", Necromancer="skltree_n_back",
    Paladin="skltree_p_back", Barbarian="skltree_b_back", Druid="skltree_d_back",
    Assassin="skltree_i_back",
}

return {
    blocks_update_below = true,

    enter = function(self)
        self.root = render.create("modal")
        self.controls = controls.new()
        if not render.assets_available() then return end

        local character = assert(saves.selected(), "skill tree requires a selected character")
        local background = assert(backgrounds[character.class], "unsupported skill-tree class: " .. tostring(character.class))

        self.panel = render.create("modal", self.root)
        local width, height = self.panel:set_dc6_combined(
            -- String concatenation constructs the class-specific asset path from
            -- the small safe lookup name above.
            "data/global/ui/SPELLS/" .. background .. ".dc6", manifest.palettes.sky, 0, 0)

        local x, y = screen.x, screen.y
        self.panel:set_position(x + width / 2, y + height / 2)

        -- Be explicit that the interaction body is unfinished rather than
        -- pretending presentation owns skill allocation rules.
        text.create(self.root, "disabled", assert(locale.text("darkmagic.skills.unavailable")),
            x + width / 2, y + height - screen.unavailable_bottom, width - 40)

        local close = {
            sheet="data/global/ui/PANEL/buysellbtn.DC6", palette="sky",
            up_frame=10, down_frame=11, x=x + width - screen.close_inset, y=y + height - screen.close_inset,
            width=32, height=32, label="darkmagic.skills.close",
        }

        button.create(self.root, self.controls, "close", close, assert(locale.text(close.label)), {
            layer="modal", show_label=false, sound=manifest.sounds.button,
            tooltip=assert(locale.text(close.label)), on_activate=function() scenes.toggle_overlay("skills", "right") end,
        })
    end,

    update = function(self)
        self.controls:update()
        if input.pressed("skills") or input.pressed("cancel") then scenes.pop() end
    end,
}
