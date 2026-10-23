-- Supported text-only conversations. Speech IDs resolve through the recovered
-- speech/localization join; quest availability and gossip selection are separate policy.
local M = {}

local conversations = {
    akara = { gossip = "akara_act1_gossip_8" },
}

function M.supports(npc)
    return conversations[string.lower(npc or "")] ~= nil
end

function M.choices(npc, page, trade)
    if not M.supports(npc) then return {} end
    if page == "topics" then
        return {
            { id = "gossip", label = "d2legacy.npc.gossip" },
            { id = "back", label = "cancel" },
        }
    elseif page == "text" then
        return { { id = "back", label = "cancel" } }
    end
    local choices = {
        { id = "talk", label = "d2legacy.npc.talk" },
        { id = "cancel", label = "cancel" },
    }
    if trade then table.insert(choices, 2, { id = "trade", label = "d2legacy.npc.trade" }) end
    return choices
end

function M.select(npc, page, choice)
    local definition = conversations[string.lower(npc or "")]
    if not definition then return nil end
    if page == "menu" and choice == "talk" then
        return "topics", ""
    elseif page == "topics" and choice == "gossip" then
        return "text", definition.gossip
    elseif (page == "topics" or page == "text") and choice == "back" then
        return "menu", ""
    elseif page == "menu" and (choice == "cancel" or choice == "back") then
        return "closed", ""
    end
    return nil
end

return M
