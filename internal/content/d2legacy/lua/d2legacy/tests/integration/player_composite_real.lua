local render = require("engine.render/v1")
local adapter = require("d2legacy.gameplay.player_composite")
for _, token in ipairs({ "AM", "SO", "NE", "PA", "BA", "AI", "DZ" }) do
    -- TN is the compact frontend/Realm-roster idle used by character lists;
    -- the remaining modes are live world presentation.
    for _, mode in ipairs({ "TN", "NU", "WL", "RN" }) do
        local ok, err = pcall(function()
            local composite = adapter.unarmed({
                token = token,
                mode = mode,
                weapon_class = "HTH",
                palette = "data/global/Palette/units/pal.dat",
                direction = 3,
            })
            assert(composite.direction == 14 and composite.dcc_direction == 3)
            if token == "NE" then
                assert(composite.components.S1 and composite.components.S2)
            end
            local node = render.create("world")
            assert(
                node:set_cof_animation(
                    composite.cof,
                    composite.palette,
                    composite.direction,
                    composite.components,
                    "loop",
                    composite.rate
                ) > 0
            )
            node:destroy()
        end)
        assert(ok, token .. mode .. ": " .. tostring(err))
    end
end
for _, mode in ipairs({ "NU", "WL", "RN" }) do
    local composite = adapter.resolve({
        token = "AM",
        mode = mode,
        weapon_class = "HTH",
        palette = "data/global/Palette/units/pal.dat",
        direction = 3,
    }, {
        active_weapon_set = 0,
        items = {
            {
                container = "equipment",
                slot = "rarm",
                weapon_set = 0,
                weapon_class = "1HS",
                composite = { RH = "SSD" },
            },
        },
    })
    local node = render.create("world")
    assert(
        node:set_cof_animation(
            composite.cof,
            composite.palette,
            composite.direction,
            composite.components,
            "loop",
            composite.rate
        ) > 0
    )
    node:destroy()
end
