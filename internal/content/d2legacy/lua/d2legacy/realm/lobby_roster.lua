-- Live Realm-channel character strip.
--
-- This is intentionally fed only by ChannelView.members. Account characters,
-- the locally selected character, and historical chat senders are not presence
-- and must never appear in the online roster.

local render = require("engine.render/v1")
local common = require("d2legacy.realm.common")
local character_composite = require("d2legacy.characters.composite")

local M = {}

local maximum_members = 8
local slot_width = 96
local first_x = 16
local actor_top = 468
local actor_height = 90
local actor_bottom = actor_top + actor_height
local label_y = 570

local function fit_scale(width, height)
    width = math.max(1, tonumber(width) or 1)
    height = math.max(1, tonumber(height) or 1)
    return math.min((slot_width - 6) / width, actor_height / height)
end

local function presence_key(member)
    if not member or not member.character then return nil end
    local character = member.character
    return table.concat({
        tostring(member.member_id or ""),
        tostring(character.character_id or character.id or ""),
        tostring(character.class or ""),
        tostring(character.level or ""),
    }, ":")
end

function M.create(root)
    local roster = { slots = {} }
    for index = 1, maximum_members do
        local left = first_x + (index - 1) * slot_width
        local slot = { left = left }
        if render.assets_available() then
            slot.actor = render.create("hud", root)
            slot.actor:set_clip(left, 468, slot_width, 100)
            slot.label = common.label(root, "", left + slot_width / 2, label_y, slot_width - 4,
                "realm_lobby_text", "center")
            slot.actor:set_visible(false)
            if slot.label then slot.label:set_visible(false) end
        end
        roster.slots[index] = slot
    end
    return roster
end

function M.update(roster, members)
    members = members or {}
    for index, slot in ipairs(roster.slots) do
        local member = members[index]
        local character = member and member.character
        local visible = character ~= nil

        if slot.actor then slot.actor:set_visible(visible) end
        if slot.label then slot.label:set_visible(visible) end

        if visible then
            local key = presence_key(member)
            if slot.presence_key ~= key and slot.actor then
                local recipe = character_composite.recipe(character)
                local _, _, width, height, _, origin_y = slot.actor:set_cof_animation(
                    recipe.cof,
                    recipe.palette,
                    recipe.direction,
                    recipe.components,
                    "loop",
                    recipe.rate
                )
                local scale = fit_scale(width, height)
                slot.actor:set_scale(scale, scale)
                -- COF nodes are positioned at their logical ground contact,
                -- not their canvas center. This puts the scaled canvas at the
                -- top of the authored roster viewport for every class/equipment
                -- combination while keeping its feet in the correct place.
                -- Bottom alignment is authoritative. Width-limited composites
                -- can be shorter than the viewport and should stand on the same
                -- baseline as every other player, immediately above the name.
                local ground_y = actor_bottom - (height - (tonumber(origin_y) or height)) * scale
                slot.actor:set_position(slot.left + slot_width / 2, ground_y)
            end
            slot.presence_key = key

            local title = tostring(character.title or "")
            local caption = "[white]" .. tostring(character.name or "")
            if title ~= "" then caption = caption .. "\n[green]" .. title end
            common.set_label(slot.label, caption, slot_width - 4, "character_select_metadata", "center")
        else
            slot.presence_key = nil
        end
    end
end

return M
