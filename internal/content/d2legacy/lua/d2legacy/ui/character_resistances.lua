-- Value-only resistance figures for the character panel.
--
-- Authority owns the resistances: derived_stats applies the difficulty penalty
-- and stat sources to d2legacy.combat.defense. Connected clients receive only
-- the current display inputs in a presentation mirror. This module never
-- writes ECS state.

local M = {}

-- Elements in panel order. Poison is not modeled yet, so it has no value.
M.elements = { "fire", "cold", "lightning" }

-- The panel shows the effective resistance limited to its maximum. The limit is
-- inferred from the original game's behavior, not from recovered client code.
function M.displayed(defense)
    local values = {}
    for _, element in ipairs(M.elements) do
        local value = defense[element .. "_resist"]
        local maximum = defense["max_" .. element .. "_resist"]
        values[element] = math.min(value, maximum)
    end
    return values
end

-- Prefer local authority when present; connected clients read the owner-only
-- presentation mirror. Missing values leave the panel boxes empty.
function M.for_player(player_id)
    -- Required lazily: the overlay is registered at boot, before frontend
    -- scenes have an ECS capability.
    local ecs = require("engine.ecs/v1")
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.player.identity", "d2legacy.combat.defense" } })) do
        if ecs.get(entity, "d2legacy.player.identity"):get("player") == player_id then
            return M.displayed(ecs.get(entity, "d2legacy.combat.defense"):snapshot())
        end
    end
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.player.identity", "d2legacy.presentation.character_resistances" } })) do
        if ecs.get(entity, "d2legacy.player.identity"):get("player") == player_id then
            return M.displayed(ecs.get(entity, "d2legacy.presentation.character_resistances"):snapshot())
        end
    end
    return nil
end

-- Offline play admits the local character as "local-player"; a hosted or
-- connected session reports its own player ID through the network status.
function M.local_player()
    local available, network = pcall(require, "engine.network/v1")
    if available then
        local status = network.status()
        if status.player_id and status.player_id ~= "" then
            return M.for_player(status.player_id)
        end
    end
    return M.for_player("local-player")
end

return M
