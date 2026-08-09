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

-- OpenD2 preserves the original 800x600 expansion menu geometry, including
-- the gaps occupied by the closed-Battle.net/gateway controls. Riiablo also
-- identifies 3WideButtonBlank.dc6 as the large frontend button artwork. Dark
-- Magic currently exposes the supported subset of controls, but positions that
-- subset at the authored locations instead of compacting the menu vertically.
M.screens = {
    main_menu = {
        controls = {
            single_player = {
                x = 265, y = 290, width = 272, height = 35,
                sheet = "data/global/ui/FrontEnd/3WideButtonBlank.dc6",
                palette = "units",
                up_frames = {0, 1}, down_frames = {2, 3}, disabled_frames = {4, 5},
                text_offset = 0,
            },
            multiplayer = {
                x = 265, y = 400, width = 272, height = 35,
                sheet = "data/global/ui/FrontEnd/3WideButtonBlank.dc6",
                palette = "units",
                up_frames = {0, 1}, down_frames = {2, 3}, disabled_frames = {4, 5},
                text_offset = 0,
            },
            credits = {
                x = 265, y = 495, width = 135, height = 35,
                sheet = "data/global/ui/FrontEnd/MediumButtonBlank.dc6",
                palette = "units",
                up_frames = {0}, down_frames = {1},
                text_offset = 0,
            },
            cinematics = {
                x = 410, y = 495, width = 135, height = 35,
                sheet = "data/global/ui/FrontEnd/MediumButtonBlank.dc6",
                palette = "units",
                up_frames = {0}, down_frames = {1},
                text_offset = 0,
            },
            exit = {
                x = 265, y = 535, width = 272, height = 35,
                sheet = "data/global/ui/FrontEnd/3WideButtonBlank.dc6",
                palette = "units",
                up_frames = {0, 1}, down_frames = {2, 3}, disabled_frames = {4, 5},
                text_offset = 0,
            },
        },
    },
}

-- Return a copy so compatibility facts can override recovered/legacy values
-- without mutating the loaded manifest table shared by other Lua modules.
function M.screen_control(screen_id, control_id, fallback)
    local result = {}
    for key, value in pairs(assert(fallback, "fallback control is required")) do
        result[key] = value
    end
    local screen = M.screens[screen_id]
    local override = screen and screen.controls and screen.controls[control_id]
    if override then
        for key, value in pairs(override) do result[key] = value end
    end
    return result
end

return M
