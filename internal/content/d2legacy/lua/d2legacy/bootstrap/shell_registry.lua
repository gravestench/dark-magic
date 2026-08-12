-- Registry for simple placeholder/shell interfaces.
--
-- Not every Diablo II panel needs a custom implementation on day one. A shell
-- gives a scene a name, rectangle, lifecycle, cursor, and close behavior while
-- the real feature is still being built. This is useful for engine development
-- and for mods: you can reserve the interaction surface first, then replace the
-- shell with a richer module later without changing every caller.

local scenes = require("engine.scene/v1")
local cursor = require("d2legacy.ui.cursor")
local overlay_shell = require("d2legacy.ui.overlay_shell")
local routing = require("d2legacy.bootstrap.overlay_routing")

local registry = {}

-- Shells are deliberately DATA-DRIVEN. The table below is easier to compare
-- with recovered Diablo II geometry than sixteen separate nearly-identical Lua
-- files would be.
--
-- Common fields:
--   title               localization key shown by the generic shell
--   x/y/width/height    logical 800x600 placement
--   sheet/palette       optional authored DC6 background
--   blocks_update_below whether the scene under this shell keeps updating
--   passes_input_below  whether routed input may continue below
--   world_view          how the world frames around the overlay
--   layer               retained render layer (modal/hud/etc.)
--   slot                overlay-routing lane when one is needed
local definitions = {
    quick_skills={title="engine.shell.quick_skills",x=470,y=220,width=250,height=270},
    belt={title="engine.shell.belt",x=250,y=430,width=300,height=100,blocks_update_below=false,layer="hud"},
    messages={title="engine.shell.messages",x=120,y=100,width=560,height=380,blocks_update_below=false,passes_input_below=true,world_view="center",layer="hud",slot="full"},
    move_gold={title="engine.shell.move_gold",sheet="data/global/ui/MENU/dialogbackground.DC6",x=270,y=175},
    npc_interaction={title="engine.shell.npc_interaction",x=250,y=180,width=300,height=260},
    npc_dialogue={title="engine.shell.npc_dialogue",x=100,y=390,width=600,height=130,blocks_update_below=false},
    item_tooltip={title="engine.shell.item_tooltip",x=250,y=140,width=300,height=320,blocks_update_below=false},
    ground_items={title="engine.shell.ground_items",x=170,y=120,width=460,height=340,blocks_update_below=false,layer="hud"},
    confirmation_dialog={title="engine.shell.confirmation_dialog",sheet="data/global/ui/FrontEnd/PopUpOkCancel.dc6",palette="fechar",x=270,y=175},
    area_transition={title="engine.shell.area_transition",x=100,y=180,width=600,height=220},
    player_trade={title="engine.shell.player_trade",x=80,y=64,width=640,height=432},
    gambling={title="engine.shell.gambling",sheet="data/global/ui/PANEL/buysell.dc6",x=80,y=64},
    npc_services={title="engine.shell.npc_services",x=200,y=160,width=400,height=300},
    hireling_hire={title="engine.shell.hireling_hire",x=160,y=100,width=480,height=380},
    chat={title="engine.shell.chat",x=80,y=430,width=640,height=100,blocks_update_below=false},
    overhead_labels={title="engine.shell.overhead_labels",x=120,y=100,width=560,height=380,blocks_update_below=false,layer="hud"},
}

function registry.register_all(manifest)
    -- Registration order does not matter, so `pairs` is the simple fit here.
    for name, definition in pairs(definitions) do
        -- Convert one data record into an actual scene definition using a
        -- reusable factory. This is composition: data + generic behavior.
        local scene = overlay_shell.new(definition)

        if definition.slot then
            -- Only shells participating in the gameplay overlay lanes need the
            -- shared hotkey/cancel routing decorator.
            scene = routing.wrap(scene, name, definition.slot, definition.world_view, definition.passes_input_below)
        end

        -- Every shell still gets the normal software-cursor ownership decorator,
        -- then is published under its friendly scene ID.
        scenes.register(name, cursor.wrap(scene, manifest.cursor, manifest.palettes))
    end
end

return registry
