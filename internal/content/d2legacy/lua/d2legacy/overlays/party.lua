-- Fixed-profile party panel backed by the owner-scoped semantic projection.
-- The panel observes d2legacy.party/v1 through a derived component and never
-- treats its roster rows as authoritative membership.

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local saves = require("d2legacy.save/v1")
local data = require("engine.data/v1")
local locale = require("engine.locale/v1")
local controls = require("d2legacy.ui.controls")
local button = require("d2legacy.ui.button")
local text = require("d2legacy.ui.text")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2legacy.presentation/v1"))
local screen = manifest.screens.party
local offset_x, offset_y = screen.offset_x or 0, screen.offset_y or 0

local function panel_frame(root, panel, frame_index, x, y)
    local node = render.create("modal", root)
    local w, h = node:set_dc6(panel.sheet, manifest.palettes[panel.palette], 0, frame_index)
    node:set_position(x + w / 2, y + h / 2)
end

local function selected_party_view()
    -- Frontend-only hosts register the scene catalog before a gameplay ECS
    -- exists, so acquire this capability only when the in-game panel opens.
    local ecs = require("engine.ecs/v1")
    local selected = saves.selected()
    local fallback
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.player.identity", "d2legacy.player.party_view" } })) do
        local identity = ecs.get(entity, "d2legacy.player.identity")
        local view = ecs.get(entity, "d2legacy.player.party_view")
        fallback = fallback or { identity = identity, view = view }
        if selected and identity:get("character_id") == selected.id then
            return identity, view
        end
    end
    return fallback and fallback.identity, fallback and fallback.view
end

local relationship_keys = {
    self = "d2legacy.party.relationship.self",
    party = "d2legacy.party.relationship.party",
    invited_you = "d2legacy.party.relationship.invited_you",
    invited = "d2legacy.party.relationship.invited",
    unavailable = "d2legacy.party.relationship.unavailable",
    available = "d2legacy.party.relationship.available",
}

return {
    blocks_update_below = true,

    enter = function(self)
        self.root = render.create("modal")
        self.controls = controls.new()
        if not render.assets_available() then return end

        local panel = screen.panel

        -- Four authored quadrants make the full party panel.
        panel_frame(self.root, panel, panel.frames[1], panel.x + offset_x, panel.y + offset_y)
        panel_frame(self.root, panel, panel.frames[2], panel.x + offset_x + 256, panel.y + offset_y)
        panel_frame(self.root, panel, panel.frames[3], panel.x + offset_x, panel.y + offset_y + 256)
        panel_frame(self.root, panel, panel.frames[4], panel.x + offset_x + 256, panel.y + offset_y + 256)

        local identity, view = selected_party_view()
        text.create(self.root, "panel_heading", identity and identity:get("name") or "", screen.heading.x + offset_x, screen.heading.y + offset_y, screen.heading.width)

        if not view or view:get("schema_version") ~= 1 or view:get("roster_count") < 1 then
            text.create(self.root, "disabled", assert(locale.text("d2legacy.party.unavailable")), screen.unavailable.x + offset_x, screen.unavailable.y + offset_y, screen.unavailable.width)
        else
            for slot = 1, view:get("roster_count") do
                local relationship = view:get("relationship_" .. slot)
                local status = assert(locale.text(assert(relationship_keys[relationship], "unknown party relationship")))
                local row = string.format(
                    "%s  %s  %s %d  %s",
                    view:get("name_" .. slot),
                    view:get("class_" .. slot),
                    assert(locale.text("d2legacy.party.level")),
                    view:get("level_" .. slot),
                    status
                )
                text.create(
                    self.root,
                    "panel_label",
                    row,
                    screen.roster.x + offset_x,
                    screen.roster.y + offset_y + (slot - 1) * screen.roster.row_height,
                    screen.roster.width,
                    "left"
                )
            end
        end

        local close = screen.close
        local close_placement = {
            sheet=close.sheet, palette=close.palette, up_frame=close.up_frame, down_frame=close.down_frame,
            x=close.x + offset_x, y=close.y + offset_y, width=close.width, height=close.height, label=close.label,
        }
        button.create(self.root, self.controls, "close", close_placement, assert(locale.text(close.label)), {
            layer="modal", show_label=false, sound=manifest.sounds.button,
            tooltip=assert(locale.text(close.label)), on_activate=function() scenes.toggle_overlay("party", "full") end,
        })
    end,

    update = function(self)
        self.controls:update()
        if input.pressed("party") or input.pressed("cancel") then scenes.pop() end
    end,
}
