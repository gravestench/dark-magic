local scenes = require("dm.scene/v1")
local cursor = require("darkmagic.ui.cursor")
local routing = require("darkmagic.bootstrap.overlay_routing")
local shells = require("darkmagic.bootstrap.shell_registry")

local registry = {}

local function with_cursor(definition, manifest, options)
    return cursor.wrap(definition, manifest.cursor, manifest.palettes, options)
end

local function register_screens(manifest)
    local screens = {
        loading={module="darkmagic.screens.loading", cursor={hidden=true}},
        title={module="darkmagic.screens.title"},
        main_menu={module="darkmagic.screens.main_menu"},
        character_select={module="darkmagic.screens.character_select"},
        character_create={module="darkmagic.screens.character_create"},
        game_world={module="darkmagic.screens.game_world"},
        game_loading={module="darkmagic.screens.game_loading", cursor={hidden=true}},
        tcpip={module="darkmagic.screens.tcpip"},
        credits={module="darkmagic.screens.credits"},
        font_lab={module="darkmagic.screens.font_lab"},
        ui_lab={module="darkmagic.screens.ui_lab"},
    }
    for name, item in pairs(screens) do
        scenes.register(name, with_cursor(require(item.module), manifest, item.cursor))
    end

    -- The movie browser needs a special rule: show the pointer while choosing,
    -- hide it while a movie owns the whole screen.
    local cinematics = require("darkmagic.screens.cinematics")
    scenes.register("cinematics", with_cursor(cinematics, manifest, {
        visible_when=function(scene) return scene.playback == nil end,
    }))
end

local function register_overlays(manifest)
    local overlays = {
        inventory={module="darkmagic.overlays.inventory",slot="right",view="left",passes=true},
        character={module="darkmagic.overlays.character",slot="left",view="right",passes=true},
        skills={module="darkmagic.overlays.skills",slot="right",view="left",passes=true},
        automap={module="darkmagic.overlays.automap",slot="full",view="center",passes=true},
        options={module="darkmagic.overlays.options",slot="full",view="center",passes=false},
        pause={module="darkmagic.overlays.pause",slot="full",view="center",passes=false},
        help={module="darkmagic.overlays.help",slot="full",view="center",passes=true},
        quests={module="darkmagic.overlays.quests",slot="left",view="right",passes=true},
        party={module="darkmagic.overlays.party",slot="full",view="center",passes=true},
    }
    for name, item in pairs(overlays) do
        local overlay = routing.wrap(require(item.module), name, item.slot, item.view, item.passes)
        scenes.register(name, with_cursor(overlay, manifest))
    end
end

local function register_simple_overlays(manifest)
    for _, name in ipairs({"stash", "cube", "hireling", "vendor", "waypoint", "death"}) do
        local overlay = require("darkmagic.overlays." .. name)
        scenes.register(name, with_cursor(overlay, manifest))
    end
end

function registry.register_all(manifest)
    register_screens(manifest)
    register_overlays(manifest)
    register_simple_overlays(manifest)
    shells.register_all(manifest)
end

return registry
