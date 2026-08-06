local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local vfs = require("dm.vfs/v1")
local audio = require("dm.audio/v1")
local simulation = require("dm.simulation/v1")

return {
    create = function(self)
        self.root = render.create("world")
        self.hero = render.create("world", self.root)
        self.hero:fill_rect(24, 32, 180, 40, 30, 255)
        self.content_source = assert(vfs.source("boot.lua"))
        self.town_music_available = audio.exists("data/global/music/Act1/town1.wav")
    end,

    update = function(self, elapsed, focused)
        if not focused then return end
        local speed = 160 * elapsed
        local dx, dy = 0, 0
        if input.down("left") then dx = dx - speed end
        if input.down("right") then dx = dx + speed end
        if input.down("up") then dy = dy - speed end
        if input.down("down") then dy = dy + speed end
        if dx ~= 0 or dy ~= 0 then simulation.move_hero(dx, dy) end
        local state = simulation.state()
        self.hero:set_position(state.hero_x, state.hero_y)
        if input.pressed("inventory") then
            scenes.push("inventory")
        elseif input.pressed("character") then
            scenes.push("character")
        elseif input.pressed("skills") then
            scenes.push("skills")
        elseif input.pressed("automap") then
            scenes.push("automap")
        elseif input.pressed("options") then
            scenes.push("options")
        elseif input.pressed("pause") or input.pressed("cancel") then
            scenes.push("pause")
        end
    end,
}
