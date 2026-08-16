-- Shared presentation pieces for explicit Realm account screens.

local data = require("engine.data/v1")
local button = require("d2legacy.ui.button")
local common = require("d2legacy.realm.common")
local frontend_logo = require("d2legacy.ui.frontend_logo")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2legacy.presentation/v1"))
local M = {}

-- Realm account entry uses textbox2.dc6. The source is a multipart frontend
-- control rendered with the units palette; decoding a single frame with the
-- fechar palette produces the bright cyan/magenta corruption seen when these
-- screens were first wired.
function M.text_field(definition)
    local result = {
        width = 272,
        height = 32,
        palette = "units",
        combined = true,
    }
    for key, value in pairs(definition or {}) do
        result[key] = value
    end
    return result
end

local function wide_button(y)
    return {
        x = 264,
        y = y,
        width = 272,
        height = 35,
        sheet = "data/global/ui/FrontEnd/WideButtonBlank.dc6",
        palette = "units",
        up_frames = { 0, 1 },
        down_frames = { 2, 3 },
        text_offset = 1,
    }
end

function M.create_root()
    local root = common.frontend_root()
    return root, frontend_logo.create(root, manifest.screens.main_menu.logo, manifest.palettes, "hud")
end

function M.update_logo(logo, elapsed)
    frontend_logo.update(logo, elapsed)
end

function M.add_button(root, controls, id, y, label, on_activate)
    return button.create(root, controls, id, wide_button(y), label, {
        normal_style = "realm_wide_button_normal",
        hover_style = "realm_wide_button_hover",
        disabled_style = "disabled",
        on_activate = on_activate,
    })
end

function M.clear_secret(field)
    if not field then
        return
    end
    field.value = ""
    field.cursor = 0
    if field.on_change then
        field.on_change(field, "")
    end
end

return M
