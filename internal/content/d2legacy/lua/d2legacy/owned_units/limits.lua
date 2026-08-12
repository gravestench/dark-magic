-- Decide which older owned units must become inactive when a new one appears.
--
-- This module deliberately knows nothing about ECS.  It receives plain Lua
-- tables, sorts them deterministically, and returns the victims.  That makes
-- the rather fiddly Diablo summon-limit rule easy to read and test by itself.

local M = {}

local function ordered(candidates, newest_first)
    local copy = {}
    for index, candidate in ipairs(candidates) do copy[index] = candidate end
    table.sort(copy, function(left, right)
        if left.created_tick == right.created_tick then
            return left.entity:id() < right.entity:id()
        end
        if newest_first then return left.created_tick > right.created_tick end
        return left.created_tick < right.created_tick
    end)
    return copy
end

function M.victims(candidates, category)
    local same_category, same_group = {}, {}
    for _, candidate in ipairs(candidates) do
        if candidate.active then
            if candidate.category == category.id then
                same_category[#same_category + 1] = candidate
            elseif category.group > 0 and candidate.group == category.group then
                -- PetType groups are mutually exclusive even when each member
                -- has room under its own category maximum.
                same_group[#same_group + 1] = candidate
            end
        end
    end

    local needed = math.max(0, #same_category + 1 - category.base_max)
    if needed == 0 and #same_group == 0 then return {} end
    assert(category.replacement ~= "reject", "owned-unit category limit reached")

    local result = {}
    for _, candidate in ipairs(same_group) do result[#result + 1] = candidate end
    same_category = ordered(same_category, category.replacement == "replace_newest")
    for index = 1, needed do result[#result + 1] = same_category[index] end
    table.sort(result, function(left, right)
        return left.entity:id() < right.entity:id()
    end)
    return result
end

return M
