-- Shared keyboard/controller routing for gameplay overlays.
--
-- This file is small, but it demonstrates one of the cleverest patterns in the
-- shim: **wrapping** an existing scene definition to add policy without copying
-- or subclassing it.
--
-- Imagine an inventory overlay already knows how to draw itself. We still want
-- every gameplay overlay to understand hotkeys such as Inventory, Skills, Help,
-- and Cancel. Instead of teaching the same hotkey rules to nine files, this
-- module adds those rules around each overlay's own `update` callback.

-- Versioned input capability: asks what player actions happened this frame.
local input = require("dm.input/v1")
-- Versioned scene capability: owns overlay open/close/replacement behavior.
local scenes = require("dm.scene/v1")

local routing = {}

-- These are the overlays a player can toggle with ordinary game actions.
--
-- `action` is an input action name, NOT a physical key. A keyboard, controller,
-- or remapped binding can all produce the same action.
--
-- `slot` says which overlay lane the panel owns. Left and right can coexist;
-- `full` is a whole-screen/modal lane.
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
    -- `ipairs` walks this array in order. The first matching action wins, which
    -- means one frame cannot accidentally open several overlays at once.
    for _, route in ipairs(routes) do
        if input.pressed(route.action) then
            -- `toggle_overlay` means: open it if absent, close it if it is the
            -- active overlay in that slot, or replace the current occupant.
            scenes.toggle_overlay(route.id, route.slot)
            -- Returning true tells the wrapper that navigation happened, so the
            -- old overlay's own update must NOT continue afterward.
            return true
        end
    end

    if input.pressed("cancel") then
        -- Cancel toggles the *current* overlay, which is a generic way to say
        -- "close me" without each overlay hard-coding its own filename/module.
        scenes.toggle_overlay(current_id, current_slot)
        return true
    end

    -- No shared navigation consumed this frame.
    return false
end

-- `wrap` adds the shared overlay rules around one overlay's own update function.
-- Think of wrapping paper: the gift inside stays the same, but another behavior
-- now surrounds it.
function routing.wrap(definition, id, slot, world_view, passes_input)
    if passes_input then
        -- Some panels (inventory, character sheet, etc.) let the world continue
        -- updating and receive a carefully routed subset of input below them.
        definition.passes_input_below = true
        definition.blocks_update_below = false
    end

    -- `or` is a compact default: use the caller's value when present, otherwise
    -- use "center". The world scene reads this to frame the hero around panels.
    definition.world_view = world_view or "center"

    -- Save the overlay's original callback BEFORE replacing it. In Lua,
    -- functions are ordinary values, so a local variable can hold one.
    local update_content = definition.update

    -- Replace the callback with a new function that performs shared routing
    -- first and then delegates to the original overlay update.
    definition.update = function(self, ...)
        -- `...` means "all remaining arguments". The wrapper does not need to
        -- know every scene-update parameter to pass them through unchanged.
        if handle_route(id, slot) then return end

        -- Some tiny overlays have no update callback of their own, so guard the
        -- call rather than trying to invoke nil.
        if update_content then update_content(self, ...) end
    end

    -- Returning the same table makes decorators easy to chain:
    -- routing.wrap(...) -> cursor.wrap(...) -> scenes.register(...)
    return definition
end

return routing
