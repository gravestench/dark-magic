local render = require("dm.render/v1")
local scenes = require("dm.scene/v1")
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

return {
    id = "darkmagic.boot",
    api = 1,

    start = function(self)
        -- The bootstrap owns a transition-layer root. Scene content will attach
        -- below this root as the Lua-authored shell is brought online.
        self.root = render.create("transition")

        scenes.register("loading", loading)
        scenes.register("title", title)
        scenes.register("main_menu", main_menu)
        scenes.register("character_select", character_select)
        scenes.register("character_create", character_create)
        scenes.register("game_world", game_world)
        scenes.register("game_loading", game_loading)
        scenes.register("tcpip", tcpip)
        scenes.register("credits", credits)
        scenes.register("cinematics", cinematics)
        scenes.register("font_lab", font_lab)
        scenes.register("inventory", inventory)
        scenes.register("character", character)
        scenes.register("skills", skills)
        scenes.register("automap", automap)
        scenes.register("options", options)
        scenes.register("pause", pause)
        scenes.replace("loading")
    end,

    stop = function(self)
        self.root = nil
    end,
}
