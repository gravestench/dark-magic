-- Recovered Diablo II UI presentation facts shared by d2legacy screens.
--
-- THIS FILE IS MOSTLY DATA ON PURPOSE.
--
-- When reverse-engineering an old game's presentation, it is very tempting to
-- scatter magic numbers such as `265`, frame `14`, or draw mode `4` directly
-- through widget code. Dark Magic keeps cross-checked historical facts here so a
-- reader can tell the difference between:
--
--   "Diablo II appears to have used this value"
-- and
--   "Dark Magic chose this Lua implementation."
--
-- That distinction matters for clean-room work and for mods. A mod may replace
-- presentation facts without having to replace the reusable control/widget logic.
--
-- Values here are behavioral observations cross-checked against the reference
-- engine audit and original assets. This is intentionally data, not a port of
-- any reference project's widget/rendering implementation.

local M = {}

-- DRAW MODES ---------------------------------------------------------------
-- Old D2 rendering APIs identify blend behavior with small numeric mode IDs.
-- Screens should not have to remember what "4" means, so translate those legacy
-- numbers once into descriptive Dark Magic renderer blend names.
--
-- Several observed numeric modes collapse to ordinary alpha blending. Raylib's
-- predefined multiply mode is the exact blend equation used by D2 draw mode 4:
-- GL_DST_COLOR, GL_ONE_MINUS_SRC_ALPHA. Mode 3 is the custom frontend screen
-- blend, while mode 6 remains a named future compatibility target.
M.draw_modes = {
    [0] = "alpha",
    [1] = "alpha",
    [2] = "alpha",
    [3] = "screen",
    [4] = "multiply",
    [5] = "alpha",
    [6] = "d2-src-color-add",
    [7] = "alpha",
}

function M.draw_mode(value)
    local result = M.draw_modes[value]
    -- Fail loudly for an unknown recovered number instead of silently guessing.
    assert(result ~= nil, "unknown Diablo II draw mode: " .. tostring(value))
    return result
end

-- SHARED WIDGET FACTS ------------------------------------------------------
-- These are recovered/verified facts used by reusable widget implementations.
-- Keeping them here means checkbox.lua, scrollbar.lua, etc. can talk in semantic
-- names such as `checked_frame` instead of unexplained numeric literals.
M.widgets = {
    checkbox = {
        sheet = "data/global/ui/FrontEnd/clickbox.dc6",
        width = 15,
        height = 16,
        unchecked_frame = 0,
        checked_frame = 1,
        label_gap = 6,
        label_style = "character_create_option",
    },
    text_field = {
        name_sheet = "data/global/ui/FrontEnd/textbox.dc6",
        generic_sheet = "data/global/ui/FrontEnd/textbox2.dc6",
        ip_sheet = "data/global/ui/FrontEnd/IPAddressBox.dc6",
        background_y = 4,
        text_x = 6,
        text_y = -2,
        label_y = -15,
        text_style = "frontend_legal",
        label_style = "character_create_option",
    },
    button = {
        pressed_dx = -2,
        pressed_dy = 2,
        text_draw_mode = 4,
        disabled_alpha = 0.5,
        click_sound = "data/global/sfx/cursor/button.wav",
    },
    -- OpenDiablo2 recovered these TextSlid.dc6 frame meanings while rebuilding
    -- the original text/layout scrollbar. Keep the facts here so the visual
    -- widget does not own reverse-engineered magic numbers.
    text_scrollbar = {
        sheet = "data/global/ui/MENU/TextSlid.dc6",
        palette = "sky",
        part_width = 12,
        part_height = 13,
        down_hollow_frame = 8,
        up_hollow_frame = 9,
        down_filled_frame = 10,
        up_filled_frame = 11,
        gutter_frame = 13,
        thumb_frame = 14,
    },
    -- Verified in the shipped archive and exercised by the in-game sound and
    -- music settings. OptBarC's 255px body and 35px right cap form a 290px
    -- track; OptSkull is the independently positioned 28px thumb.
    option_slider = {
        track_sheet = "data/global/ui/Widgets/OptBarC.dc6",
        thumb_sheet = "data/global/ui/Widgets/OptSkull.dc6",
        palette = "units",
        width = 290,
        height = 37,
        thumb_size = 28,
        confidence = "verified",
    },
}

-- FRONTEND FACTS -----------------------------------------------------------
-- Original 800x600 placement recovered from OpenD2 and corroborated by Riiablo's
-- authored asset choices. These tables intentionally look like configuration,
-- because that is exactly what they are.
M.frontend = {
    main_menu = {
        controls = {
            -- Each row says: top-left position, source sheet, and which spatial
            -- frames represent up/down/disabled states. Multi-frame buttons are
            -- arrays because their art is assembled side by side.
            single_player = { x = 265, y = 290, sheet = "data/global/ui/FrontEnd/3WideButtonBlank.dc6", up_frames = {0, 1}, down_frames = {2, 3}, disabled_frames = {4, 5} },
            multiplayer = { x = 265, y = 400, sheet = "data/global/ui/FrontEnd/3WideButtonBlank.dc6", up_frames = {0, 1}, down_frames = {2, 3}, disabled_frames = {4, 5} },
            credits = { x = 265, y = 495, sheet = "data/global/ui/FrontEnd/MediumButtonBlank.dc6", up_frames = {0}, down_frames = {1} },
            cinematics = { x = 410, y = 495, sheet = "data/global/ui/FrontEnd/MediumButtonBlank.dc6", up_frames = {0}, down_frames = {1} },
            exit = { x = 265, y = 535, sheet = "data/global/ui/FrontEnd/3WideButtonBlank.dc6", up_frames = {0, 1}, down_frames = {2, 3}, disabled_frames = {4, 5} },
        },
        disclaimer = {
            x = 400,
            y = 590,
            width = 760,
            align = "center",
            -- Gold Formal 10 is the small legal/disclaimer treatment used by
            -- the frontend. The existing credits style is exactly that font,
            -- palette transform, and gold text color.
            style = "credits",
        },
    },
    character_create = {
        -- These are TOP-based text boxes. d2.ui.text converts them to
        -- retained center coordinates after measuring the actual bitmap text.
        heading = { x = 400, y = 25, width = 800 },
        class_name = { x = 400, y = 75, width = 800 },
        description = { x = 400, y = 105, width = 800 },

        -- OpenD2's animated renderer adds 400 to supplied frontend Y and then
        -- applies the DC6 offset. Dark Magic's normalized frame top has a +1
        -- term, so the equivalent common anchor is sourceY + 399.
        campfire = { anchor = { x = 375, y = 319 }, draw_mode = 3 },
        idle_back_frames_per_second = 8,
        idle_front_frames_per_second = 25,
        transition_frames_per_second = 25,

        -- Overlapping class actors need deterministic front-to-back hit/render order.
        draw_order = {
            Barbarian = 20,
            Necromancer = 30,
            Paladin = 40,
            Assassin = 50,
            Sorceress = 60,
            Amazon = 70,
            Druid = 80,
        },
        stage = {
            Amazon = { anchor = { x = 100, y = 329 } },
            Sorceress = { anchor = { x = 626, y = 329 }, overlay_draw_mode = 3 },
            Necromancer = { anchor = { x = 301, y = 329 }, overlay_draw_mode = 3 },
            Paladin = { anchor = { x = 520, y = 329 } },
            Barbarian = { anchor = { x = 400, y = 329 } },
            Druid = { anchor = { x = 720, y = 349 } },
            Assassin = { anchor = { x = 232, y = 349 } },
        },

        -- Static buttons are always present. The original scene uses PAL_FECHAR,
        -- and MediumSelButtonBlank is authored for that palette.
        controls = {
            exit = { x = 35, y = 535, sheet = "data/global/ui/FrontEnd/MediumSelButtonBlank.dc6", palette = "fechar", up_frames = {0}, down_frames = {1} },
            ok = { x = 630, y = 535, sheet = "data/global/ui/FrontEnd/MediumSelButtonBlank.dc6", palette = "fechar", up_frames = {0}, down_frames = {1} },
        },
        form = {
            x = 320,
            y = 490,
            name = { x = 320, y = 490, kind = "name", max_length = 15 },
            expansion = { x = 320, y = 525 },
            hardcore = { x = 320, y = 545 },
            minimum_name_length = 2,
        },
    },
    game_loading = {
        -- Both paths are present and byte-identical in the verified asset
        -- fixture. Some installations/archive stacks expose only one usable
        -- copy, so loading tries both before surfacing a decode error.
        sheets = {
            "data/global/ui/Loading/loadingscreen.dc6",
            "data/local/ui/loadingscreen.dc6",
        },
    },
}

-- IN-GAME ESCAPE MENU FACTS -----------------------------------------------
-- This large nested table is the DATA MODEL consumed by escape_menu.lua.
-- `target` means "navigate to another menu layout"; `action` means "report a
-- semantic action to owning scene"; `values` means cycle through a fixed list;
-- `range` means bind a numeric slider to a real setting.
--
-- Riiablo confirms localized DC6 label art, modal dim, select sound, paired
-- pents, and reversed left pent; OpenDiablo2 supplies keyboard ordering, default
-- focus, submenu structure, and options vocabulary. Dark Magic implements these
-- facts with its own retained controls.
M.ingame = {
    escape_menu = {
        confidence = "high",
        -- Source strings are provenance notes for future maintainers, not runtime imports.
        sources = {
            "OpenDiablo2/OpenDiablo2:d2game/d2player/escape_menu.go",
            "OpenDiablo2/OpenDiablo2:d2common/d2resource/resource_paths.go",
            "collinsmith/riiablo:core/src/main/java/com/riiablo/screen/panel/EscapePanel.java",
        },
        dim = { red = 0, green = 0, blue = 0, alpha = 128 },
        center = { x = 400, y = 300 },
        menu_width = 500,
        row_height = 54,
        label_gutter = 10,
        label_side_gap = 24,
        palette = "units",
        select_sound = "data/global/sfx/cursor/select.wav",
        pentagram = {
            sheet = "data/global/ui/CURSOR/pentspin.DC6",
            width = 54,
            height = 54,
            frames_per_second = 15,
            left_reversed = true,
            right_reversed = false,
        },
        simulation = {
            -- These are compatibility/policy facts, not something escape_menu.lua
            -- infers from graphics.
            pauses_single_player = true,
            pauses_multiplayer = false,
        },
        layouts = {
            main = {
                default_focus = "return_to_game",
                font = "font42",
                rows = {
                    { id = "options", label = "OPTIONS", sheet = "data/local/ui/eng/options.dc6", target = "options" },
                    { id = "save_exit", label = "SAVE AND EXIT GAME", sheet = "data/local/ui/eng/exit.dc6", action = "save_exit" },
                    { id = "return_to_game", label = "RETURN TO GAME", sheet = "data/local/ui/eng/returntogame.dc6", action = "close" },
                },
            },
            options = {
                default_focus = "previous_menu",
                font = "font42",
                rows = {
                    { id = "sound_options", label = "SOUND OPTIONS", sheet = "data/local/ui/eng/soundoptions.dc6", target = "sound" },
                    { id = "video_options", label = "VIDEO OPTIONS", sheet = "data/local/ui/eng/videoOptions.dc6", target = "video" },
                    { id = "automap_options", label = "AUTOMAP OPTIONS", sheet = "data/local/ui/eng/automapOptions.dc6", target = "automap" },
                    { id = "configure_controls", label = "CONFIGURE CONTROLS", sheet = "data/local/ui/eng/cfgOptions.dc6", target = "controls" },
                    { id = "previous_menu", label = "PREVIOUS MENU", sheet = "data/local/ui/eng/previous.dc6", target = "main" },
                },
            },
            sound = {
                title = "SOUND OPTIONS",
                default_focus = "previous_menu",
                font = "font30",
                -- OpenDiablo2 supplies this hierarchy and the OptBarC/OptSkull
                -- paths. Direct archive inspection supplies frame dimensions;
                -- Riiablo and AbyssEngine independently validate normalized
                -- 0..1 music/effects ranges. OpenD2 corroborates panel/input
                -- ordering but has no implemented options surface.
                rows = {
                    -- `range` rows become slider controls tied to named settings.
                    { id = "sound_volume", label = "SOUND", sheet = "data/local/ui/eng/sound.dc6", range = { setting = "sound_volume", min = 0, max = 1, step = 0.05 } },
                    { id = "music_volume", label = "MUSIC", sheet = "data/local/ui/eng/music.dc6", range = { setting = "music_volume", min = 0, max = 1, step = 0.05 } },
                    -- `unavailable=true` intentionally leaves recovered rows visible
                    -- but disabled until Dark Magic exposes the required capability.
                    { id = "3d_bias", label = "3D BIAS", sheet = "data/local/ui/eng/3dbias.dc6", unavailable = true },
                    { id = "hardware_acceleration", label = "HARDWARE ACCELERATION", values = { "ON", "OFF" }, unavailable = true },
                    { id = "environmental_effects", label = "ENVIRONMENTAL EFFECTS", values = { "ON", "OFF" }, unavailable = true },
                    { id = "npc_speech", label = "NPC SPEECH", sheet = "data/local/ui/eng/npcspeech.dc6", values = { "AUDIO AND TEXT", "AUDIO ONLY", "TEXT ONLY" }, unavailable = true },
                    { id = "previous_menu", label = "PREVIOUS MENU", sheet = "data/local/ui/eng/previous.dc6", target = "options" },
                },
            },
            video = {
                title = "VIDEO OPTIONS",
                default_focus = "previous_menu",
                font = "font30",
                rows = {
                    { id = "resolution", label = "VIDEO RESOLUTION", sheet = "data/local/ui/eng/resolution.dc6", values = { "800X600", "1024X768" }, unavailable = true },
                    { id = "lighting_quality", label = "LIGHTING QUALITY", sheet = "data/local/ui/eng/lightquality.dc6", values = { "LOW", "HIGH" }, unavailable = true },
                    { id = "blended_shadows", label = "BLENDED SHADOWS", sheet = "data/local/ui/eng/blendshadow.dc6", values = { "ON", "OFF" }, unavailable = true },
                    { id = "perspective", label = "PERSPECTIVE", sheet = "data/local/ui/eng/prespective.dc6", values = { "ON", "OFF" }, unavailable = true },
                    { id = "gamma", label = "GAMMA", sheet = "data/local/ui/eng/gamma.dc6", unavailable = true },
                    { id = "contrast", label = "CONTRAST", sheet = "data/local/ui/eng/contrast.dc6", unavailable = true },
                    { id = "previous_menu", label = "PREVIOUS MENU", sheet = "data/local/ui/eng/previous.dc6", target = "options" },
                },
            },
            automap = {
                title = "AUTOMAP OPTIONS",
                default_focus = "previous_menu",
                font = "font30",
                rows = {
                    { id = "automap_size", label = "AUTOMAP SIZE", sheet = "data/local/ui/eng/automapmode.dc6", values = { "FULL SCREEN", "MINI MAP" }, unavailable = true },
                    { id = "automap_fade", label = "FADE", sheet = "data/local/ui/eng/automapfade.dc6", values = { "YES", "NO" }, unavailable = true },
                    { id = "automap_center", label = "CENTER WHEN CLEARED", sheet = "data/local/ui/eng/automapcenter.dc6", values = { "YES", "NO" }, unavailable = true },
                    { id = "automap_party", label = "SHOW PARTY", values = { "YES", "NO" }, unavailable = true },
                    { id = "automap_names", label = "SHOW NAMES", sheet = "data/local/ui/eng/automappartynames.dc6", values = { "YES", "NO" }, unavailable = true },
                    { id = "previous_menu", label = "PREVIOUS MENU", sheet = "data/local/ui/eng/previous.dc6", target = "options" },
                },
            },
            controls = {
                title = "CONFIGURE CONTROLS",
                default_focus = "previous_menu",
                font = "font30",
                -- OpenDiablo2 delegates the editable binding list to a separate
                -- key-binding menu. Dark Magic records that boundary here until
                -- an engine settings/keymap write capability is exposed to Lua.
                rows = {
                    { id = "previous_menu", label = "PREVIOUS MENU", sheet = "data/local/ui/eng/previous.dc6", target = "options" },
                },
            },
        },
        option_assets = {
            -- Reuse the shared recovered slider facts rather than repeating the
            -- same paths/dimensions in a second data section.
            range_track = M.widgets.option_slider.track_sheet,
            range_thumb = M.widgets.option_slider.thumb_sheet,
            range_width = M.widgets.option_slider.width,
            range_height = M.widgets.option_slider.height,
            range_thumb_size = M.widgets.option_slider.thumb_size,
            on = "data/local/ui/eng/smallon.dc6",
            off = "data/local/ui/eng/smalloff.dc6",
            yes = "data/local/ui/eng/smallyes.dc6",
            no = "data/local/ui/eng/smallno.dc6",
            full = "data/local/ui/eng/full.dc6",
            mini = "data/local/ui/eng/mini.dc6",
            resolution_640 = "data/local/ui/eng/640x480.dc6",
            resolution_800 = "data/local/ui/eng/800x800.dc6",
            slider = "data/global/ui/widgets/optbarc.dc6",
            slider_skull = "data/global/ui/widgets/optskull.dc6",
        },
    },
}

-- MERGING COMPAT FACTS WITH MOD MANIFEST BEHAVIOR -------------------------
--
-- The manifest still owns semantic behavior/localization/targets. This helper
-- makes a fresh copy of that fallback definition, then overlays ONLY recovered
-- presentation facts from this catalog. It never mutates the shared manifest.
function M.screen_control(screen_id, control_id, fallback)
    local result = {}

    -- Shallow-copy caller's manifest definition first.
    for key, value in pairs(assert(fallback, "fallback control is required")) do
        result[key] = value
    end

    -- These chained `and` operations safely walk optional nested tables:
    -- if screen is nil, override becomes nil without indexing through nil.
    local screen = M.frontend[screen_id]
    local override = screen and screen.controls and screen.controls[control_id]

    if override then
        -- Recovered presentation facts win when present.
        for key, value in pairs(override) do result[key] = value end
    end

    return result
end

return M
