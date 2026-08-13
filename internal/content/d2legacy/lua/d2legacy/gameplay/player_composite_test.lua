local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local palette = "data/global/Palette/units/pal.dat"

local function install_render_fixture()
    test.mock_module("engine.render/v1", {
        cof_info = function(path)
            test.assert(
                path == "data/global/chars/AM/COF/AMWLHTH.cof" or path == "data/global/chars/AM/COF/AMWL1HS.cof"
            )
            return {
                directions = 16,
                layers = {
                    { type = "HD", weapon_class = "1ht" },
                    { type = "RA", weapon_class = "hth" },
                    { type = "RH", weapon_class = "hth" },
                },
            }
        end,
        asset_exists = function(path)
            return path == "data/global/chars/AM/HD/AMHDLITWL1HT.dcc"
                or path == "data/global/chars/AM/RA/AMRALITWLHTH.dcc"
        end,
        animdata_info = function(key)
            test.assert(key == "AMWLHTH" or key == "AMWL1HS", [=[key == "AMWLHTH" or key == "AMWL1HS"]=])
            return { speed = 333, frames = 8, events = {} }
        end,
    }, { "cof_info", "asset_exists", "animdata_info" })
end

local function player_recipe(direction)
    return {
        token = "AM",
        mode = "WL",
        weapon_class = "HTH",
        palette = palette,
        direction = direction,
    }
end

local function assert_unarmed_recipe(adapter)
    local composite = adapter.unarmed(player_recipe(3))
    test.assert(
        string.sub(composite.key, 1, 12) == "AM:WL:HTH:14",
        [=[string.sub(composite.key, 1, 12) == "AM:WL:HTH:14"]=]
    )
    test.assert(
        composite.direction == 14 and composite.dcc_direction == 3,
        [=[composite.direction == 14 and composite.dcc_direction == 3]=]
    )
    test.assert(
        composite.components.HD == "data/global/chars/AM/HD/AMHDLITWL1HT.dcc",
        [=[composite.components.HD == "data/global/chars/AM/HD/AMHDLITWL1HT.dcc"]=]
    )
    test.assert(
        composite.components.RA == "data/global/chars/AM/RA/AMRALITWLHTH.dcc",
        [=[composite.components.RA == "data/global/chars/AM/RA/AMRALITWLHTH.dcc"]=]
    )
    test.assert(composite.components.RH == nil, [=[composite.components.RH == nil]=])
    test.assert(composite.rate == 333 and composite.frames == 8, [=[composite.rate == 333 and composite.frames == 8]=])
    test.assert(
        adapter.unarmed(player_recipe(15)).direction == 15,
        [=[adapter.unarmed(player_recipe(15)).direction == 15]=]
    )
end

local function assert_equipped_recipe(adapter)
    local weapon = fixtures.item({
        container = "equipment",
        slot = "rarm",
        weapon_set = 0,
        weapon_class = "1hs",
        composite = { RH = "ssd" },
    })
    local equipment = fixtures.equipment({ items = { weapon } })
    local equipped = adapter.resolve(player_recipe(3), equipment)
    test.assert(
        equipped.cof == "data/global/chars/AM/COF/AMWL1HS.cof",
        [=[equipped.cof == "data/global/chars/AM/COF/AMWL1HS.cof"]=]
    )
    test.assert(
        equipped.components.RH == "data/global/chars/AM/RH/AMRHSSDWLHTH.dcc",
        [=[equipped.components.RH == "data/global/chars/AM/RH/AMRHSSDWLHTH.dcc"]=]
    )
end

local function assert_playback_events(adapter)
    local playback = adapter.new_playback({ mode = "A" })
    local animation = {
        rate = 256,
        frames = 4,
        events = { [2] = 1, [3] = 3 },
        mode = "A",
    }
    local crossed = adapter.advance(playback, animation, 0.09)
    test.assert(playback.frame == 3, [=[playback.frame == 3]=])
    test.assert(#crossed == 2, [=[#crossed == 2]=])
    test.assert(crossed[1].event == 1 and crossed[2].event == 3, [=[crossed[1].event == 1 and crossed[2].event == 3]=])
end

local function resolves_cof_layer_weapon_classes()
    install_render_fixture()
    local adapter = require("d2legacy.gameplay.player_composite")
    assert_unarmed_recipe(adapter)
    assert_equipped_recipe(adapter)
    assert_playback_events(adapter)
end

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("resolves_cof_layer_weapon_classes", {
            { run = resolves_cof_layer_weapon_classes },
        }),
    },
})
