local test = require("d2legacy.tests/v1")

local function load_composite()
    local resolved
    test.mock_module("d2legacy.gameplay.player_composite", {
        recipe = function(authority, appearance)
            resolved = { authority = authority, appearance = appearance }
            return { cof = authority.token .. authority.mode .. authority.weapon_class, components = {} }
        end,
    }, { "recipe" })
    return require("d2legacy.characters.composite"), function()
        return resolved
    end
end

return test.suite({
    name = "character roster composite resolution",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("saved_composite_is_used_without_frontend_actor_fallback", function(t)
            t:run(function()
                local composite, resolved = load_composite()
                local appearance = {
                    cof = "saved-character.cof",
                    palette = "data/global/Palette/units/pal.dat",
                    direction = 3,
                    components = { HD = "head.dcc" },
                }

                test.expect(composite.recipe({ class = "Assassin", appearance = appearance })):deep_equals({
                    cof = appearance.cof,
                    palette = appearance.palette,
                    direction = 3,
                    components = { HD = "head.dcc" },
                })
                test.expect(resolved()):is_nil()
            end)
        end),
        test.case("fresh_character_uses_class_default_town_composite", function(t)
            t:run(function()
                local composite, resolved = load_composite()
                test.expect(composite.recipe({ class = "Barbarian" }).cof):equals("BATNHTH")
                test.expect(resolved().authority):deep_equals({
                    token = "BA",
                    mode = "TN",
                    palette = "data/global/Palette/units/pal.dat",
                    direction = 0,
                    direction_space = "semantic",
                    weapon_class = "HTH",
                })
                test.expect(resolved().appearance):deep_equals({})
            end)
        end),
    },
})
