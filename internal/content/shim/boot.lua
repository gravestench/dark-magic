local render = require("dm.render/v1")
local scenes = require("dm.scene/v1")
local loading = require("darkmagic.screens.loading")
local main_menu = require("darkmagic.screens.main_menu")
local game_world = require("darkmagic.screens.game_world")
local inventory = require("darkmagic.overlays.inventory")

return {
    id = "darkmagic.boot",
    api = 1,

    start = function(self)
        -- The bootstrap owns a transition-layer root. Scene content will attach
        -- below this root as the Lua-authored shell is brought online.
        self.root = render.create("transition")

        scenes.register("loading", loading)
        scenes.register("main_menu", main_menu)
        scenes.register("game_world", game_world)
        scenes.register("inventory", inventory)
        scenes.replace("loading")
    end,

    stop = function(self)
        self.root = nil
    end,
}
