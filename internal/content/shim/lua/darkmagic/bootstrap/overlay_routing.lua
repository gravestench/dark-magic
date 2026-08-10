local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")

local routing = {}

-- These are the overlays a player can toggle with ordinary game actions. The
-- slot says which part of the screen the overlay owns.
local routes = {
    { action="inventory", id="inventory", slot="right" },
    { action="character", id="character", slot="left" },
    { action="skills", id="skills", slot="right" },
    { action="automap", id="automap", slot="full" },
    { action="help", id="help", slot="full" },
    { action="quests", id="quests", slot="left" },
    { action="party", id="party", slot="full" },
    { action="options", id="options", slot="full" },
}

local function handle_route(current_id, current_slot)
    for _, route in ipairs(routes) do
        if input.pressed(route.action) then
            scenes.toggle_overlay(route.id, route.slot)
            return true
        end
    end
    if input.pressed("cancel") then
        scenes.toggle_overlay(current_id, current_slot)
        return true
    end
    return false
end

-- wrap adds the shared overlay rules around one overlay's own update function.
-- It is wrapping paper: the gift inside stays unchanged.
function routing.wrap(definition, id, slot, world_view, passes_input)
    if passes_input then
        definition.passes_input_below = true
        definition.blocks_update_below = false
    end
    definition.world_view = world_view or "center"
    local update_content = definition.update
    definition.update = function(self, ...)
        if handle_route(id, slot) then return end
        if update_content then update_content(self, ...) end
    end
    return definition
end

return routing
