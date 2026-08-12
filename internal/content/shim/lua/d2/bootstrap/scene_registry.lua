-- Scene registry: the shim's phone book.
--
-- A scene is just a Lua table with callbacks such as create/update/destroy.
-- The engine does not need to know filenames like `d2.screens.title`.
-- Instead, mods register friendly IDs such as `title` and later navigate to
-- those IDs through `engine.scene/v1`.
--
-- This file also demonstrates a very useful modding technique: decorate a
-- scene with shared behavior (cursor ownership, overlay routing) instead of
-- copying that behavior into every screen.

-- Engine capability: owns the scene catalog and navigation stack.
local scenes = require("engine.scene/v1")
-- Ordinary shim module: adds software-cursor behavior around scene callbacks.
local cursor = require("d2.ui.cursor")
-- Ordinary shim module: adds the common gameplay-overlay keyboard rules.
local routing = require("d2.bootstrap.overlay_routing")
-- Registers placeholder/shell interfaces that do not yet need custom Lua.
local shells = require("d2.bootstrap.shell_registry")

-- Modules normally return a table containing the functions they export.
local registry = {}

-- Small helpers are valuable when they name an idea. Every scene needs the same
-- manifest cursor definition and palette table, so callers should not repeat
-- those arguments over and over.
local function with_cursor(definition, manifest, options)
    -- `cursor.wrap` returns the same scene definition after decorating some of
    -- its lifecycle callbacks. It does not create a second native scene type.
    return cursor.wrap(definition, manifest.cursor, manifest.palettes, options)
end

local function register_screens(manifest)
    -- This table is DATA. Each key becomes a public scene ID; `module` tells us
    -- which ordinary Lua module supplies that scene definition.
    --
    -- Keeping this mapping together makes the available screen surface easy for
    -- a mod author to discover without reading engine code.
    local screens = {
        -- Loading screens hide the normal pointer because the player cannot use
        -- it there.
        loading={module="d2.screens.loading", cursor={hidden=true}},
        title={module="d2.screens.title"},
        main_menu={module="d2.screens.main_menu"},
        character_select={module="d2.screens.character_select"},
        character_create={module="d2.screens.character_create"},
        game_world={module="d2.screens.game_world"},
        game_loading={module="d2.screens.game_loading", cursor={hidden=true}},
        tcpip={module="d2.screens.tcpip"},
        credits={module="d2.screens.credits"},
        -- Development labs are real scenes too. That is useful: they exercise
        -- the exact same modding APIs as shipping presentation.
        font_lab={module="d2.screens.font_lab"},
        ui_lab={module="d2.screens.ui_lab"},
		composite_lab={module="d2.screens.composite_lab"},
		monster_lab={module="d2.screens.monster_lab"},
		missile_lab={module="d2.screens.missile_lab"},
		combat_lab={module="d2.screens.combat_lab"},
		dt1_lab={module="d2.screens.dt1_lab"},
		ds1_lab={module="d2.screens.ds1_lab"},
		mapgen_lab={module="d2.screens.mapgen_lab"},
		warp_lab={module="d2.screens.warp_lab"},
    }

    -- `pairs` is appropriate here because registration order is not meaningful.
    -- For each row:
    --   1. require the scene's Lua table
    --   2. wrap it with cursor behavior
    --   3. publish it under its friendly scene name
    for name, item in pairs(screens) do
        scenes.register(name, with_cursor(require(item.module), manifest, item.cursor))
    end

    -- The movie browser needs one extra rule. While choosing a movie the cursor
    -- is useful, but while a video owns the whole screen it should disappear.
    -- `visible_when` is a callback passed into the generic cursor decorator.
    local cinematics = require("d2.screens.cinematics")
    scenes.register("cinematics", with_cursor(cinematics, manifest, {
        -- A nil playback handle means no movie is currently active.
        visible_when=function(scene) return scene.playback == nil end,
    }))
end

local function register_overlays(manifest)
    -- Gameplay overlays need more policy than root screens. Besides the module,
    -- this table describes:
    --   slot   = which overlay lane it occupies
    --   view   = where the world should frame the player while it is visible
    --   passes = whether useful routed input may continue below it
    --
    -- Notice that the overlay's own file does NOT need to know all this global
    -- routing policy. That separation keeps individual panels easier to read.
    local overlays = {
        inventory={module="d2.overlays.inventory",slot="right",view="left",passes=true},
        character={module="d2.overlays.character",slot="left",view="right",passes=true},
        skills={module="d2.overlays.skills",slot="right",view="left",passes=true},
        automap={module="d2.overlays.automap",slot="full",view="center",passes=true},
        options={module="d2.overlays.options",slot="full",view="center",passes=false},
        pause={module="d2.overlays.pause",slot="full",view="center",passes=false},
        help={module="d2.overlays.help",slot="full",view="center",passes=true},
        quests={module="d2.overlays.quests",slot="left",view="right",passes=true},
        party={module="d2.overlays.party",slot="full",view="center",passes=true},
    }

    for name, item in pairs(overlays) do
        -- First decorate the overlay with common open/close routing...
        local overlay = routing.wrap(require(item.module), name, item.slot, item.view, item.passes)
        -- ...then decorate it with common cursor ownership and register it.
        scenes.register(name, with_cursor(overlay, manifest))
    end
end

local function register_simple_overlays(manifest)
    -- These files already return a fully configured overlay or are simple enough
    -- not to need the shared route metadata above. `ipairs` preserves the list's
    -- written order, although registration itself does not depend on it.
    for _, name in ipairs({"stash", "cube", "hireling", "vendor", "waypoint", "death"}) do
        -- String concatenation (`..`) lets one short loop load six predictable
        -- module names without six near-identical lines.
        local overlay = require("d2.overlays." .. name)
        scenes.register(name, with_cursor(overlay, manifest))
    end
end

function registry.register_all(manifest)
    -- One public entry point keeps boot.lua simple and gives future mods one
    -- obvious place to add another first-party scene category.
    register_screens(manifest)
    register_overlays(manifest)
    register_simple_overlays(manifest)
    shells.register_all(manifest)
end

return registry
