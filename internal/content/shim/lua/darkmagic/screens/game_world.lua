-- Minimal interactive game-world orchestration scene.
--
-- Lua owns input and presentation flow while dm.simulation/v1 owns persistent,
-- deterministic gameplay state. The placeholder hero makes that separation
-- visible until the composite character renderer is connected.
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

        -- VFS provenance and optional asset checks are examples of querying
        -- capabilities without receiving direct filesystem/native ownership.
        self.content_source = assert(vfs.source("boot.lua"))
        self.town_music_available = audio.exists("data/global/music/Act1/town1.wav")
    end,

    update = function(self, elapsed, focused)
        -- Transparent overlays may allow updates below them, but only the top
        -- scene receives input focus.
        if not focused then
            return
        end
        local speed = 160 * elapsed
        local dx, dy = 0, 0
        if input.down("left") then
            dx = dx - speed
        end
        if input.down("right") then
            dx = dx + speed
        end
        if input.down("up") then
            dy = dy - speed
        end
        if input.down("down") then
            dy = dy + speed
        end
        if dx ~= 0 or dy ~= 0 then
            simulation.move_hero(dx, dy)
        end

        local state = simulation.state()
        self.hero:set_position(state.hero_x, state.hero_y)

        -- Panels are scene overlays rather than long-lived engine services.
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
