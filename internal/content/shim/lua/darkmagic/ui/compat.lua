-- Recovered Diablo II UI presentation facts shared by shim screens.
--
-- Values here are behavioral observations cross-checked against the reference
-- engine audit and original assets. This is intentionally data, not a port of
-- any reference project's widget/rendering implementation.
local M = {}

-- Legacy D2 draw-mode semantics observed in OpenD2's renderer. Several numeric
-- modes collapse to ordinary alpha blending; mode 3 is the logo/fire screen
-- blend used by the frontend. Mode 4 and 6 are retained as named facts for
-- future nodes that need those exact GPU factors.
M.draw_modes = {
    [0] = "alpha",
    [1] = "alpha",
    [2] = "alpha",
    [3] = "screen",
    [4] = "d2-dst-color",
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

return M
