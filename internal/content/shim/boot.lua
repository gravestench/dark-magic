local render = require("dm.render/v1")
local scenes = require("dm.scene/v1")
local loading = require("darkmagic.screens.loading")
local title = require("darkmagic.screens.title")
local main_menu = require("darkmagic.screens.main_menu")
local character_select = require("darkmagic.screens.character_select")
local game_world = require("darkmagic.screens.game_world")
local inventory = require("darkmagic.overlays.inventory")
local character = require("darkmagic.overlays.character")
local skills = require("darkmagic.overlays.skills")
local automap = require("darkmagic.overlays.automap")
local options = require("darkmagic.overlays.options")
local pause = require("darkmagic.overlays.pause")
local static_frontend = require("darkmagic.screens.static_frontend")

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
        scenes.register("game_world", game_world)
        scenes.register("tcpip", static_frontend("tcpip"))
        scenes.register("credits", static_frontend("credits"))
        scenes.register("cinematics", static_frontend("cinematics"))
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
