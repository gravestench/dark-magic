-- Recovered Diablo II UI presentation facts shared by shim screens.
--
-- Values here are behavioral observations cross-checked against the reference
-- engine audit and original assets. This is intentionally data, not a port of
-- any reference project's widget/rendering implementation.
local M = {}

-- Legacy D2 draw-mode semantics observed in OpenD2's renderer. Several numeric
-- modes collapse to ordinary alpha blending. Raylib's predefined multiply mode
-- is the exact blend equation used by D2 draw mode 4:
-- GL_DST_COLOR, GL_ONE_MINUS_SRC_ALPHA. Mode 3 is kept as the custom frontend
-- screen blend, while mode 6 remains a named future compatibility target.
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
    assert(result ~= nil, "unknown Diablo II draw mode: " .. tostring(value))
    return result
end

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
}

-- Original 800x600 frontend placement recovered from OpenD2 and corroborated
-- by Riiablo's authored asset choices. Gaps on the main menu intentionally
-- preserve the original Battle.net/gateway rows even when Dark Magic does not
-- expose those controls.
M.frontend = {
    main_menu = {
        controls = {
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
            -- Small Exocet, deliberately quieter than the authored buttons.
            style = "button_normal",
        },
    },
    character_create = {
        -- OpenD2 supplies top coordinates for these labels. Dark Magic's text
        -- helper converts the recovered top edge to retained-node center space.
        heading = { x = 400, y = 25, width = 800 },
        class_name = { x = 400, y = 75, width = 800 },
        description = { x = 400, y = 105, width = 800 },

        -- OpenD2's animated renderer adds 400 to the supplied frontend Y and
        -- then applies the DC6 offset. Dark Magic's normalized frame top has a
        -- +1 term, so the equivalent anchor is sourceY + 399.
        campfire = { anchor = { x = 375, y = 319 }, draw_mode = 3 },
        idle_back_frames_per_second = 8,
        idle_front_frames_per_second = 25,
        transition_frames_per_second = 25,

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

        -- Static buttons are always present. The dynamic form appears only
        -- while a class is selected.
        controls = {
            exit = { x = 35, y = 535, sheet = "data/global/ui/FrontEnd/MediumSelButtonBlank.dc6", up_frames = {0}, down_frames = {1} },
            ok = { x = 630, y = 535, sheet = "data/global/ui/FrontEnd/MediumSelButtonBlank.dc6", up_frames = {0}, down_frames = {1} },
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
}

-- Compatibility facts override the manifest without mutating the loaded
-- manifest table shared by other Lua modules. Behavior/locale/targets remain
-- in the manifest; only recovered presentation facts are merged here.
function M.screen_control(screen_id, control_id, fallback)
    local result = {}
    for key, value in pairs(assert(fallback, "fallback control is required")) do
        result[key] = value
    end
    local screen = M.frontend[screen_id]
    local override = screen and screen.controls and screen.controls[control_id]
    if override then
        for key, value in pairs(override) do result[key] = value end
    end
    return result
end

return M
