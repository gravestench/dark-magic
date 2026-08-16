-- Resolve the compact roster portrait for one durable character record.
--
-- Imported saves may already carry an exact COF/DCC snapshot. Fresh characters
-- have no equipment snapshot yet, so their class supplies the canonical naked
-- town-idle composite. Neither path uses the oversized frontend class actors.

local player_composite = require("d2legacy.gameplay.player_composite")

local class_tokens = {
    Amazon = "AM",
    Sorceress = "SO",
    Necromancer = "NE",
    Paladin = "PA",
    Barbarian = "BA",
    Assassin = "AI",
    Druid = "DZ",
}

local composite = {}

function composite.recipe(character)
    assert(character, "character is required")
    local appearance = character.appearance
    if appearance and appearance.cof and appearance.cof ~= "" then
        return {
            cof = appearance.cof,
            palette = appearance.palette,
            direction = appearance.direction or 0,
            components = appearance.components or {},
            rate = appearance.rate,
        }
    end

    return player_composite.recipe({
        token = assert(class_tokens[character.class], "unknown character class: " .. tostring(character.class)),
        mode = "TN",
        palette = "data/global/Palette/units/pal.dat",
        direction = 0,
        direction_space = "semantic",
        weapon_class = "HTH",
    }, {})
end

return composite
