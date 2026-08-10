local render = require("dm.render/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local input = require("dm.input/v1")
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
local help = require("darkmagic.overlays.help")
local quests = require("darkmagic.overlays.quests")
local party = require("darkmagic.overlays.party")
local stash = require("darkmagic.overlays.stash")
local cube = require("darkmagic.overlays.cube")
local hireling = require("darkmagic.overlays.hireling")
local vendor = require("darkmagic.overlays.vendor")
local waypoint = require("darkmagic.overlays.waypoint")
local death = require("darkmagic.overlays.death")
local overlay_shell = require("darkmagic.ui.overlay_shell")
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

        local overlay_routes = {
            { action="inventory", id="inventory", slot="right" },
            { action="character", id="character", slot="left" },
            { action="skills", id="skills", slot="right" },
            { action="automap", id="automap", slot="full" },
            { action="help", id="help", slot="full" },
            { action="quests", id="quests", slot="left" },
            { action="party", id="party", slot="full" },
            { action="options", id="options", slot="full" },
        }

        local function route_overlay_input(current_id, current_slot)
            for _, route in ipairs(overlay_routes) do
                if input.pressed(route.action) then
                    scenes.toggle_overlay(route.id, route.slot)
                    return true
                end
            end
            if input.pressed("cancel") then
                scenes.toggle_overlay(current_id, current_slot)
                return true
            end
            return false
        end

        local function slotted_overlay(definition, id, slot, world_view, passes_input)
            if passes_input then
                definition.passes_input_below = true
                definition.blocks_update_below = false
            end
            definition.world_view = world_view or "center"
            local original_update = definition.update
            definition.update = function(self, ...)
                if route_overlay_input(id, slot) then return end
                if original_update then original_update(self, ...) end
            end
            return definition
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
        scenes.register("inventory", with_cursor(slotted_overlay(inventory, "inventory", "right", "left", true)))
        scenes.register("character", with_cursor(slotted_overlay(character, "character", "left", "right", true)))
        scenes.register("skills", with_cursor(slotted_overlay(skills, "skills", "right", "left", true)))
        scenes.register("automap", with_cursor(slotted_overlay(automap, "automap", "full", "center", true)))
        scenes.register("options", with_cursor(slotted_overlay(options, "options", "full", "center", false)))
        scenes.register("pause", with_cursor(slotted_overlay(pause, "pause", "full", "center", false)))
        scenes.register("help", with_cursor(slotted_overlay(help, "help", "full", "center", true)))
        scenes.register("quests", with_cursor(slotted_overlay(quests, "quests", "left", "right", true)))
        scenes.register("party", with_cursor(slotted_overlay(party, "party", "full", "center", true)))
        scenes.register("stash", with_cursor(stash))
        scenes.register("cube", with_cursor(cube))
        scenes.register("hireling", with_cursor(hireling))
        scenes.register("vendor", with_cursor(vendor))
        scenes.register("waypoint", with_cursor(waypoint))
        scenes.register("death", with_cursor(death))
        local shells = {
            quick_skills={title="darkmagic.shell.quick_skills",x=470,y=220,width=250,height=270},
            belt={title="darkmagic.shell.belt",x=250,y=430,width=300,height=100,blocks_update_below=false,layer="hud"},
            messages={title="darkmagic.shell.messages",x=120,y=100,width=560,height=380,blocks_update_below=false,passes_input_below=true,world_view="center",layer="hud",slot="full"},
            move_gold={title="darkmagic.shell.move_gold",sheet="data/global/ui/MENU/dialogbackground.DC6",x=270,y=175},
            npc_interaction={title="darkmagic.shell.npc_interaction",x=250,y=180,width=300,height=260},
            npc_dialogue={title="darkmagic.shell.npc_dialogue",x=100,y=390,width=600,height=130,blocks_update_below=false},
            item_tooltip={title="darkmagic.shell.item_tooltip",x=250,y=140,width=300,height=320,blocks_update_below=false},
            ground_items={title="darkmagic.shell.ground_items",x=170,y=120,width=460,height=340,blocks_update_below=false,layer="hud"},
            confirmation_dialog={title="darkmagic.shell.confirmation_dialog",sheet="data/global/ui/FrontEnd/PopUpOkCancel.dc6",palette="fechar",x=270,y=175},
            area_transition={title="darkmagic.shell.area_transition",x=100,y=180,width=600,height=220},
            player_trade={title="darkmagic.shell.player_trade",x=80,y=64,width=640,height=432},
            gambling={title="darkmagic.shell.gambling",sheet="data/global/ui/PANEL/buysell.dc6",x=80,y=64},
            npc_services={title="darkmagic.shell.npc_services",x=200,y=160,width=400,height=300},
            hireling_hire={title="darkmagic.shell.hireling_hire",x=160,y=100,width=480,height=380},
            chat={title="darkmagic.shell.chat",x=80,y=430,width=640,height=100,blocks_update_below=false},
            overhead_labels={title="darkmagic.shell.overhead_labels",x=120,y=100,width=560,height=380,blocks_update_below=false,layer="hud"},
        }
        for name, definition in pairs(shells) do
            local scene = overlay_shell.new(definition)
            if definition.slot then
                scene = slotted_overlay(scene, name, definition.slot, definition.world_view, definition.passes_input_below)
            end
            scenes.register(name, with_cursor(scene))
        end
        scenes.replace("loading")
    end,

    stop = function(self)
        self.root = nil
    end,
}
