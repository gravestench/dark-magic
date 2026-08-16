-- Shared authored Realm lobby composition.
--
-- Waiting room tabs are not separate frontend pages: the 800x600 waiting-room
-- background stays fixed while one 373x373 game pane is swapped on the right.
-- Keeping that geometry here prevents Create and Join from drifting apart.

local render = require("engine.render/v1")
local button = require("d2legacy.ui.button")
local common = require("d2legacy.realm.common")

local M = {}

local panel_x = 414
local panel_y = 72

function M.create(panel_kind)
    local root = render.create("hud")
    local backdrop = render.create("hud", root)

    if render.assets_available() then
        local background = assert(common.manifest.screens.realm_lobby.background)
        backdrop:set_dc6_combined(
            background.sheet,
            assert(common.manifest.palettes[background.palette]),
            0,
            0
        )
        backdrop:set_position(400, 300)
        backdrop:set_z(-100)

        local definition = assert(common.manifest.screens.realm_lobby[panel_kind])
        local panel = render.create("hud", root)
        local width, height = panel:set_dc6_combined(
            definition.sheet,
            assert(common.manifest.palettes[definition.palette]),
            0,
            0
        )
        panel:set_position(panel_x + width / 2, panel_y + height / 2)
        return root, backdrop, panel
    end

    backdrop:fill_rect(800, 600, 8, 8, 8, 255)
    backdrop:set_position(400, 300)
    backdrop:set_z(-100)
    return root, backdrop, nil
end

function M.button(kind, x, y)
    return common.screen_definition("realm_lobby", kind, x, y)
end

function M.chat_lines(events)
    local first = math.max(1, #(events or {}) - 11)
    local lines = {}
    for index = first, #(events or {}) do
        local event = events[index]
        local sender = event.sender and event.sender.character and event.sender.character.name or "Realm"
        if event.kind == "message" then
            lines[#lines + 1] = string.format("<%s> %s", sender, event.text or "")
        elseif event.kind == "member_joined" then
            lines[#lines + 1] = sender .. " has joined the channel."
        elseif event.kind == "member_left" then
            lines[#lines + 1] = sender .. " has left the channel."
        end
    end
    return table.concat(lines, "\n")
end

-- The right-side navigation and disabled placeholders remain visible while
-- either game tab is open. Callers provide semantics; this helper owns only the
-- authored geometry and consistent presentation.
function M.add_navigation(root, manager, handlers)
    handlers = handlers or {}
    local function add(id, kind, x, y, label, handler, enabled)
        button.create(root, manager, id, M.button(kind, x, y), label, {
            enabled=enabled,
            normal_style="realm_lobby_button",
            hover_style="realm_lobby_button",
            disabled_style="realm_lobby_button_disabled",
            on_activate=handler,
        })
    end
    add("nav_create", "top_button", 532, 450, "CREATE", handlers.create, handlers.create ~= false)
    add("nav_join", "top_button", 653, 450, "JOIN", handlers.join, handlers.join ~= false)
    add("nav_channel", "bottom_button", 535, 477, "CHANNEL", nil, false)
    add("nav_ladder", "bottom_button", 616, 477, "LADDER", nil, false)
    add("nav_quit", "bottom_button", 697, 477, "QUIT", handlers.quit, handlers.quit ~= false)
end

return M
