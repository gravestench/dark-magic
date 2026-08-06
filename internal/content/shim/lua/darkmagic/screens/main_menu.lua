local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local locale = require("dm.locale/v1")
local audio = require("dm.audio/v1")
local dc6 = require("darkmagic.ui.dc6")
local controls = require("darkmagic.ui.controls")
local app = require("dm.app/v1")
local cursor = require("darkmagic.ui.cursor")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.main_menu
local logo = screen.logo
local units_palette = manifest.palettes[logo.palette]

return {
    enter = function(self)
        self.root = render.create("hud")
        self.background = dc6.frontend_background(
            self.root,
            "hud",
            screen.background,
            manifest.palettes[screen.palette],
            manifest.layouts.frontend_tiles
        )
        if render.assets_available() then
            self.logo = {
                black_left = render.create("hud", self.root),
                black_right = render.create("hud", self.root),
                fire_left = render.create("hud", self.root),
                fire_right = render.create("hud", self.root),
            }
            self.logo.fire_left:set_blend(logo.fire_blend)
            self.logo.fire_right:set_blend(logo.fire_blend)
            self:configure_logo()
        end
        self:configure_controls()
        self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)
    end,

    configure_logo = function(self)
        local function animate(node, path)
            dc6.anchored_frame(node, path, units_palette, logo.anchor.x, logo.anchor.y, 0)
            node:set_dc6_animation(path, units_palette, 0, logo.frames_per_second, "loop")
        end
        animate(self.logo.black_left, logo.black_left)
        animate(self.logo.black_right, logo.black_right)
        animate(self.logo.fire_left, logo.fire_left)
        animate(self.logo.fire_right, logo.fire_right)
    end,

    configure_controls = function(self)
        self.controls = controls.new()
        local function add_control(id, definition)
          local control = {
            id = id,
            label = assert(locale.text(definition.label)),
            x = definition.x, y = definition.y,
            width = definition.width, height = definition.height,
            on_activate = function()
                if audio.exists(manifest.sounds.select) then audio.play(manifest.sounds.select) end
                if definition.action == "exit" then app.request_exit()
                else scenes.replace(definition.target or "character_select") end
            end,
        }
        if render.assets_available() then
            local palette = manifest.palettes[definition.palette]
            local pieces = {}
            for index = 1, #definition.up_frames do pieces[index] = render.create("hud", self.root) end
            local label = render.create("hud", self.root)
            local function draw_frames(frames)
                for index, node in ipairs(pieces) do node:set_dc6(definition.sheet, palette, 0, frames[index]) end
            end
            draw_frames(definition.up_frames)
            if #pieces == 1 then pieces[1]:set_position(definition.x + definition.width / 2, definition.y + definition.height / 2)
            else
                pieces[1]:set_position(definition.x + 128, definition.y + definition.height / 2)
                pieces[2]:set_position(definition.x + 264, definition.y + definition.height / 2)
            end
            local font = manifest.fonts.exocet10
            label:set_text(font.table, font.sheet, manifest.palettes[font.palette], control.label, {
                red = 210, green = 180, blue = 110, max_width = definition.width, align = "center"
            })
            label:set_position(definition.x + definition.width / 2, definition.y + definition.height / 2)
            control.on_state = function(_, state)
                if state == "focused" or state == "hover" then draw_frames(definition.down_frames)
                else draw_frames(definition.up_frames) end
            end
        end
          self.controls:add(control)
        end
        for _, id in ipairs({"single_player", "multiplayer", "credits", "cinematics", "exit"}) do add_control(id, screen.controls[id]) end
    end,

    update = function(self, elapsed)
        if self.controls then self.controls:update() end
        if self.cursor then self.cursor:update() end
        if input.pressed("cancel") then scenes.replace("title") end
    end,
}
