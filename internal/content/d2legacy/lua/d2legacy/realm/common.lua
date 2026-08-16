local render = require("engine.render/v1")
local data = require("engine.data/v1")
local dc6 = require("d2legacy.ui.dc6")
local text = require("d2legacy.ui.text")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2legacy.presentation/v1"))
local M = { manifest = manifest }

function M.frontend_root()
    local root = render.create("hud")
    local screen = manifest.screens.main_menu
    dc6.frontend_background(
        root,
        "hud",
        screen.background,
        manifest.palettes[screen.palette],
        manifest.layouts.frontend_tiles
    )
    return root
end

function M.panel(root, x, y, width, height)
    local panel = render.create("hud", root)
    panel:fill_rect(width, height, 9, 8, 7, 235)
    panel:set_position(x + width / 2, y + height / 2)
    return panel
end

function M.popup(root, x, y, kind)
    local definition = assert(manifest.screens.realm_common[kind or "popup"])
    local panel = render.create("hud", root)
    local width, height = definition.width, definition.height
    if render.assets_available() then
        width, height = panel:set_dc6_combined(definition.sheet, assert(manifest.palettes[definition.palette]), 0, 0)
    else
        panel:fill_rect(width, height, 9, 8, 7, 235)
    end
    panel:set_position(x + width / 2, y + height / 2)
    return panel, width, height
end

-- Return a scene-local copy so callers can position authored button art
-- without mutating the shared presentation manifest.
function M.screen_definition(screen, kind, x, y)
    local source = assert(assert(manifest.screens[screen])[kind])
    local definition = {}
    for key, value in pairs(source) do
        definition[key] = value
    end
    definition.x = x
    definition.y = y
    return definition
end

function M.button_definition(kind, x, y)
    return M.screen_definition("realm_common", kind, x, y)
end

function M.label(root, value, x, y, width, style, align)
    if not render.assets_available() then
        return nil
    end
    local node = render.create("hud", root)
    text.set(node, style or "dialog_text", value or "", width, align or "left")
    node:set_position(x, y)
    return node
end

function M.set_label(node, value, width, style, align)
    if node then
        text.set(node, style or "dialog_text", value or "", width, align or "left")
    end
end

function M.error(status)
    if status.phase ~= "error" then
        return ""
    end
    local detail = string.lower(tostring(status.error or ""))
    if detail:find("already_exists", 1, true) then
        return "THAT NAME IS ALREADY IN USE"
    end
    if detail:find("character_online", 1, true) or detail:find("already online", 1, true) then
        return "THAT CHARACTER IS ALREADY ONLINE"
    end
    if detail:find("already exists", 1, true) then
        return "THAT ACCOUNT NAME OR EMAIL IS ALREADY IN USE"
    end
    if detail:find("not verified", 1, true) then
        return "VERIFY YOUR EMAIL BEFORE LOGGING IN"
    end
    if detail:find("account credentials", 1, true) then
        return "ACCOUNT NAME OR PASSWORD IS INCORRECT"
    end
    if detail:find("unauthorized", 1, true) then
        return "ACCOUNT NAME OR PASSWORD IS INCORRECT"
    end
    if detail:find("invalid_input", 1, true) then
        return "INVALID REALM REQUEST"
    end
    if detail:find("capacity", 1, true) then
        return "THE GAME OR REALM IS CURRENTLY FULL"
    end
    if detail:find("not_found", 1, true) then
        return "THAT GAME DOES NOT EXIST"
    end
    if detail:find("invalid_password", 1, true) then
        return "THE GAME PASSWORD IS INCORRECT"
    end
    if detail:find("level_restricted", 1, true) then
        return "YOUR CHARACTER LEVEL IS OUTSIDE THIS GAME'S RANGE"
    end
    if detail:find("recipe differs", 1, true)
        or (detail:find("package", 1, true) and detail:find("differs", 1, true))
        or (detail:find("package", 1, true) and detail:find("digest", 1, true)) then
        return "GAME CONTENT DOES NOT MATCH THE SERVER"
    end
    if detail:find("unavailable", 1, true) then
        return "THE GAME SERVER IS UNAVAILABLE"
    end
    if detail:find("network trust", 1, true) or detail:find("identity", 1, true) then
        return "REALM IDENTITY COULD NOT BE VERIFIED"
    end
    if detail:find("timeout", 1, true) or detail:find("deadline", 1, true) then
        return "REALM CONNECTION TIMED OUT"
    end
    if detail:find("incompatible", 1, true) and detail:find("version", 1, true) then
        return "INCOMPATIBLE REALM VERSION"
    end
    return "UNABLE TO CONTACT REALM"
end

return M
