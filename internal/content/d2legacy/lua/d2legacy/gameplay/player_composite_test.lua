local test = require("d2legacy.tests/v1")

local palette = "data/global/Palette/units/pal.dat"

local function install_render_fixture()
    package.loaded["engine.render/v1"] = {
        cof_info = function(path)
            assert(path == "data/global/chars/AM/COF/AMWLHTH.cof" or path == "data/global/chars/AM/COF/AMWL1HS.cof")
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
            assert(key == "AMWLHTH" or key == "AMWL1HS")
            return { speed = 333, frames = 8, events = {} }
        end,
    }
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
    assert(string.sub(composite.key, 1, 12) == "AM:WL:HTH:14")
    assert(composite.direction == 14 and composite.dcc_direction == 3)
    assert(composite.components.HD == "data/global/chars/AM/HD/AMHDLITWL1HT.dcc")
    assert(composite.components.RA == "data/global/chars/AM/RA/AMRALITWLHTH.dcc")
    assert(composite.components.RH == nil)
    assert(composite.rate == 333 and composite.frames == 8)
    assert(adapter.unarmed(player_recipe(15)).direction == 15)
end

local function assert_equipped_recipe(adapter)
    local equipment = {
        active_weapon_set = 0,
        items = {
            {
                container = "equipment",
                slot = "rarm",
                weapon_set = 0,
                weapon_class = "1hs",
                composite = { RH = "ssd" },
            },
        },
    }
    local equipped = adapter.resolve(player_recipe(3), equipment)
    assert(equipped.cof == "data/global/chars/AM/COF/AMWL1HS.cof")
    assert(equipped.components.RH == "data/global/chars/AM/RH/AMRHSSDWLHTH.dcc")
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
    assert(playback.frame == 3)
    assert(#crossed == 2)
    assert(crossed[1].event == 1 and crossed[2].event == 3)
end

local function resolves_cof_layer_weapon_classes()
    install_render_fixture()
    local adapter = require("d2legacy.gameplay.player_composite")
    assert_unarmed_recipe(adapter)
    assert_equipped_recipe(adapter)
    assert_playback_events(adapter)
end

return test.suite({
    profile = "presentation",
    tier = "fast",
    tests = {
        resolves_cof_layer_weapon_classes = {
            { run = resolves_cof_layer_weapon_classes },
        },
    },
})
