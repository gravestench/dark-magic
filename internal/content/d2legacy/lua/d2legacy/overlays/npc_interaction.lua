-- One presentation of the owner's authoritative conversation. Controls submit
-- semantic choices; page and speech identity arrive through the interaction view.
local render = require("engine.render/v1")
local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local data = require("engine.data/v1")
local locale = require("engine.locale/v1")
local controls = require("d2legacy.ui.controls")
local text = require("d2legacy.ui.text")
local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2legacy.presentation/v1"))

local function rectangle(parent, x, y, width, height, red, green, blue, alpha)
    local node = render.create("modal", parent)
    node:fill_rect(width, height, red, green, blue, alpha)
    node:set_position(x + width / 2, y + height / 2)
end

local function panel(parent, x, y, width, height)
    rectangle(parent, x, y, width, height, 0, 0, 0, 190)
    rectangle(parent, x, y, width, 1, 125, 111, 76, 255)
    rectangle(parent, x, y + height - 1, width, 1, 125, 111, 76, 255)
    rectangle(parent, x, y, 1, height, 125, 111, 76, 255)
    rectangle(parent, x + width - 1, y, 1, height, 125, 111, 76, 255)
end

local function font(node, value, width, color)
    return text.set_font(node, "font16", value, width, "center", { text_color = color or "white" })
end

local function submit(self, choice)
    if self.pending_revision == self.snapshot.revision then return end
    self.pending_revision = self.snapshot.revision
    self.interaction.choose(self.snapshot, choice)
end

local function refresh(self, snapshot)
    self.snapshot = snapshot
    self.pending_revision = nil
    if self.content then self.content:destroy() end
    self.content = render.create("modal", self.root)
    self.controls = controls.new()
    if not render.assets_available() then return end

    local title = snapshot.page == "topics" and locale.text("d2legacy.npc.talk") or locale.text(snapshot.npc)
    local width, height = 120, 28 + #snapshot.choices * 22
    local center_x, top = manifest.resolution.width / 2 + 64, 100
    if snapshot.page == "text" then
        local speech = assert(require("d2legacy.quest_catalog/v1").dialog(snapshot.line), "dialogue speech is unavailable")
        width, center_x, top = 420, manifest.resolution.width / 2, 8
        local body = render.create("modal", self.content)
        local _, body_height = text.set_font(body, "font16", speech.text, width - 24, "left", { text_color = "white" })
        height = math.max(130, body_height + 42)
        panel(self.content, center_x - width / 2, top, width, height)
        body:set_position(center_x, top + 12 + body_height / 2)
        body:set_z(1)
    else
        panel(self.content, center_x - width / 2, top, width, height)
        local heading = render.create("modal", self.content)
        heading:set_z(1)
        local _, heading_height = font(heading, title, width - 16, "gold")
        heading:set_position(center_x, top + 8 + heading_height / 2)
    end

    for index, choice in ipairs(snapshot.choices) do
        local row_top = snapshot.page == "text" and top + height - 28 or top + 26 + (index - 1) * 22
        local label = locale.text(choice.label)
        local node = render.create("modal", self.content)
        node:set_z(1)
        local _, label_height = font(node, label, width - 16)
        node:set_position(center_x, row_top + label_height / 2)
        self.controls:add({
            id = choice.id, label = label,
            x = center_x - width / 2 + 4, y = row_top, width = width - 8, height = 22,
            on_activate = function() submit(self, choice.id) end,
            on_state = function(_, state)
                font(node, label, width - 16, state == "normal" and "white" or "blue")
            end,
        })
    end
end

return {
    blocks_update_below = false,
    world_view = "center",
    enter = function(self)
        self.root = render.create("modal")
        self.interaction = require("d2legacy.interactions.api")
    end,
    update = function(self, _, focused, input_allowed)
        local snapshot = self.interaction.dialogue()
        if not snapshot.active then
            if focused ~= false then scenes.pop() end
            return
        end
        if not self.snapshot or snapshot.revision ~= self.snapshot.revision or snapshot.target_id ~= self.snapshot.target_id then
            refresh(self, snapshot)
        end
        if input_allowed == false then return end
        if input.pressed("cancel") then
            submit(self, "back")
        elseif self.pending_revision ~= snapshot.revision then
            self.controls:update()
        end
    end,
}
