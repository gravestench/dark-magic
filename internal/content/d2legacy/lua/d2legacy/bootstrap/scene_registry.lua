-- Scene registry: the d2legacy mod's phone book.
--
-- A scene is just a Lua table with callbacks such as create/update/destroy.
-- The engine does not need to know filenames like `d2legacy.screens.title`.
-- Instead, mods register friendly IDs such as `title` and later navigate to
-- those IDs through `engine.scene/v1`.
--
-- This file also demonstrates a very useful modding technique: decorate a
-- scene with shared behavior (cursor ownership, overlay routing) instead of
-- copying that behavior into every screen.

-- Engine capability: owns the scene catalog and navigation stack.
local scenes = require("engine.scene/v1")
-- Ordinary d2legacy module: adds software-cursor behavior around scene callbacks.
local cursor = require("d2legacy.ui.cursor")
-- Ordinary d2legacy module: adds the common gameplay-overlay keyboard rules.
local routing = require("d2legacy.bootstrap.overlay_routing")
-- Registers placeholder/shell interfaces that do not yet need custom Lua.
local shells = require("d2legacy.bootstrap.shell_registry")

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
        loading={module="d2legacy.screens.loading", cursor={hidden=true}},
        title={module="d2legacy.screens.title"},
        main_menu={module="d2legacy.screens.main_menu"},
        character_select={module="d2legacy.screens.character_select", character_mode="local"},
        character_create={module="d2legacy.screens.character_create", character_mode="local"},
        game_world={module="d2legacy.screens.game_world"},
        game_loading={module="d2legacy.screens.game_loading", cursor={hidden=true}},
        tcpip={module="d2legacy.screens.tcpip"},
        realm_connecting={module="d2legacy.screens.realm_connecting"},
        realm_gateway={module="d2legacy.screens.realm_gateway"},
        realm_login={module="d2legacy.screens.realm_login"},
        realm_signup={module="d2legacy.screens.realm_signup"},
        realm_recovery={module="d2legacy.screens.realm_recovery"},
        -- Realm and local play instantiate the same character presentation.
        -- Only the roster authority adapter differs.
        realm_characters={module="d2legacy.screens.character_select", character_mode="realm"},
        realm_create={module="d2legacy.screens.character_create", character_mode="realm"},
        realm_lobby={module="d2legacy.screens.realm_lobby"},
        realm_game_create={module="d2legacy.screens.realm_game_create"},
        credits={module="d2legacy.screens.credits"},
        -- Development labs are real scenes too. That is useful: they exercise
        -- the exact same modding APIs as shipping presentation.
        font_lab={module="d2legacy.screens.font_lab"},
        ui_lab={module="d2legacy.screens.ui_lab"},
		composite_lab={module="d2legacy.screens.composite_lab"},
		monster_lab={module="d2legacy.screens.monster_lab"},
		missile_lab={module="d2legacy.screens.missile_lab"},
		combat_lab={module="d2legacy.screens.combat_lab"},
		spell_lab={module="d2legacy.screens.spell_lab"},
		dt1_lab={module="d2legacy.screens.dt1_lab"},
		ds1_lab={module="d2legacy.screens.ds1_lab"},
		mapgen_lab={module="d2legacy.screens.mapgen_lab"},
		warp_lab={module="d2legacy.screens.warp_lab"},
    }

    -- `pairs` is appropriate here because registration order is not meaningful.
    -- For each row:
    --   1. require the scene's Lua table
    --   2. wrap it with cursor behavior
    --   3. publish it under its friendly scene name
    for name, item in pairs(screens) do
        local definition = require(item.module)
        if item.character_mode then
            definition = definition.for_mode(item.character_mode)
        end
        scenes.register(name, with_cursor(definition, manifest, item.cursor))
    end

    -- The movie browser needs one extra rule. While choosing a movie the cursor
    -- is useful, but while a video owns the whole screen it should disappear.
    -- `visible_when` is a callback passed into the generic cursor decorator.
    local cinematics = require("d2legacy.screens.cinematics")
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
        inventory={module="d2legacy.overlays.inventory",slot="right",view="left",passes=true},
        character={module="d2legacy.overlays.character",slot="left",view="right",passes=true},
        skills={module="d2legacy.overlays.skills",slot="right",view="left",passes=true},
        automap={module="d2legacy.overlays.automap",slot="full",view="center",passes=true},
        options={module="d2legacy.overlays.options",slot="full",view="center",passes=false},
        pause={module="d2legacy.overlays.pause",slot="full",view="center",passes=false},
        help={module="d2legacy.overlays.help",slot="full",view="center",passes=true},
        quests={module="d2legacy.overlays.quests",slot="left",view="right",passes=true},
        party={module="d2legacy.overlays.party",slot="full",view="center",passes=true},
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
        local overlay = require("d2legacy.overlays." .. name)
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
