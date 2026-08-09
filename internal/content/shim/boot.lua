local render = require("dm.render/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local cursor = require("darkmagic.ui.cursor")
local loading = require("darkmagic.screens.loading")
local title = require("darkmagic.screens.title")
local main_menu = require("darkmagic.screens.main_menu")
local character_select = require("darkmagic.screens.character_select")
local character_create = require("darkmagic.screens.character_create")
local game_world = require("darkmagic.screens.game_world")
local game_loading = require("darkmagic.screens.game_loading")
local inventory = require("darkmagic.overlays.inventory")
local character = require("darkmagic.overlays.character")
local skills = require("darkmagic.overlays.skills")
local automap = require("darkmagic.overlays.automap")
local options = require("darkmagic.overlays.options")
local pause = require("darkmagic.overlays.pause")
local tcpip = require("darkmagic.screens.tcpip")
local credits = require("darkmagic.screens.credits")
local cinematics = require("darkmagic.screens.cinematics")
local font_lab = require("darkmagic.screens.font_lab")
local ui_lab = require("darkmagic.screens.ui_lab")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))

return {
    id = "darkmagic.boot",
    api = 1,

    start = function(self)
        -- The bootstrap owns a transition-layer root. Scene content will attach
        -- below this root as the Lua-authored shell is brought online.
        self.root = render.create("transition")

        -- The Diablo software pointer is a shell-wide invariant. Every focused
        -- screen/overlay receives one even if the screen itself never explicitly
        -- creates a cursor. Startup cinematics and the game loading screen hide
        -- it. The cinematics browser keeps it visible while choosing a movie and
        -- hides it only while an act/epilogue video is actually playing.
        local function with_cursor(definition, options)
            return cursor.wrap(definition, manifest.cursor, manifest.palettes, options)
        end

        scenes.register("loading", with_cursor(loading, { hidden = true }))
        scenes.register("title", with_cursor(title))
        scenes.register("main_menu", with_cursor(main_menu))
        scenes.register("character_select", with_cursor(character_select))
        scenes.register("character_create", with_cursor(character_create))
        scenes.register("game_world", with_cursor(game_world))
        scenes.register("game_loading", with_cursor(game_loading, { hidden = true }))
        scenes.register("tcpip", with_cursor(tcpip))
        scenes.register("credits", with_cursor(credits))
        scenes.register("cinematics", with_cursor(cinematics, {
            visible_when = function(scene)
                return scene.playback == nil
            end,
        }))
        scenes.register("font_lab", with_cursor(font_lab))
        scenes.register("ui_lab", with_cursor(ui_lab))
        scenes.register("inventory", with_cursor(inventory))
        scenes.register("character", with_cursor(character))
        scenes.register("skills", with_cursor(skills))
        scenes.register("automap", with_cursor(automap))
        scenes.register("options", with_cursor(options))
        scenes.register("pause", with_cursor(pause))
        scenes.replace("loading")
    end,

    stop = function(self)
        self.root = nil
    end,
}
