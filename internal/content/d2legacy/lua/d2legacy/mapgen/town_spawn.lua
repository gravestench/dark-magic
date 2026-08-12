-- Choose the Rogue Encampment entry marker from raw decoded DS1 object facts.
-- Object type 2 / ID 2 is the authored campfire relationship recovered from
-- legacy data. The host remains responsible only for collision-space search.

local M = {}

function M.choose(objects)
    for _, object in ipairs(objects or {}) do
        if object.type == 2 and object.id == 2 then
            return {x = object.x, y = object.y, search_inset = 4}
        end
    end
    return nil
end

return M
