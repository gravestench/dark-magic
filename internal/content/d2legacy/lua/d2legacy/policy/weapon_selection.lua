-- Select the hand requested by Skills.txt's weapsel field.
--
-- Basic Attack uses the default selector (right, then left). Sequence skills
-- may alternate explicitly with selector 3; equipping two weapons alone never
-- changes Basic Attack into an alternating skill.

local M = {}

function M.can_dual_wield(class)
    class = string.lower(class or "")
    return class == "barbarian" or class == "assassin"
end

function M.select(selector, primary, secondary, sequence_frame)
    if selector == 4 then
        return "unarmed"
    end
    if selector == 1 then
        return secondary ~= "" and secondary or primary
    end
    if selector == 3 and secondary ~= "" and sequence_frame % 2 == 1 then
        return secondary
    end
    return primary ~= "" and primary or secondary ~= "" and secondary or "unarmed"
end

return M
