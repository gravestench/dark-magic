local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local palette = "data/global/Palette/units/pal.dat"

local function install_render_fixture()
    test.mock_module("engine.render/v1", {
        cof_info = function(path)
            test.assert(
                path == "data/global/chars/AM/COF/AMWLHTH.cof"
                    or path == "data/global/chars/AM/COF/AMWL1HS.cof"
                    or path == "data/global/chars/AM/COF/AMRNHTH.cof"
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
            test.assert(
                key == "AMWLHTH" or key == "AMWL1HS" or key == "AMRNHTH",
                [=[key == "AMWLHTH" or key == "AMWL1HS" or key == "AMRNHTH"]=]
            )
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
        class = "Amazon",
        velocitypercent = 0,
        item_fastermovevelocity = 0,
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
    test.assert(composite.rate == 213 and composite.frames == 8, [=[composite.rate == 213 and composite.frames == 8]=])
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

local function assert_network_synchronization(adapter)
    local playback = adapter.new_playback({ mode = "WL" })
    local animation = {
        rate = 256,
        frames = 4,
        events = { [2] = 1, [3] = 3 },
        mode = "WL",
    }
    local initial = adapter.synchronize(playback, animation, 0.09)
    test.assert(playback.frame == 3 and #initial == 0, [=[playback.frame == 3 and #initial == 0]=])
    local crossed = adapter.synchronize(playback, animation, 0.13)
    test.assert(playback.frame == 4 and #crossed == 0, [=[playback.frame == 4 and #crossed == 0]=])
    adapter.synchronize(playback, animation, 0.17)
    test.assert(playback.frame == 1, [=[playback.frame == 1]=])
    adapter.synchronize(playback, animation, 0.01)
    test.assert(playback.frame == 1, [=[playback.frame == 1]=])
end

local function resolves_cof_layer_weapon_classes()
    install_render_fixture()
    local adapter = require("d2legacy.gameplay.player_composite")
    assert_unarmed_recipe(adapter)
    assert_equipped_recipe(adapter)
    assert_playback_events(adapter)
    assert_network_synchronization(adapter)
end

local function movement_stats_not_integrated_distance_drive_animation_playback()
    install_render_fixture()
    local adapter = require("d2legacy.gameplay.player_composite")
    local ordinary = player_recipe(3)
    ordinary.velocity_x = 6
    ordinary.item_fastermovevelocity = 0
    local faster = player_recipe(3)
    faster.velocity_x = 12
    faster.item_fastermovevelocity = 100

    local ordinary_composite = adapter.unarmed(ordinary)
    local faster_composite = adapter.unarmed(faster)
    test.expect(ordinary_composite.rate):equals(213)
    test.expect(faster_composite.rate):equals(340)

    local ordinary_playback = adapter.new_playback(ordinary_composite)
    local faster_playback = adapter.new_playback(faster_composite)
    adapter.advance(ordinary_playback, ordinary_composite, 0.1)
    adapter.advance(faster_playback, faster_composite, 0.1)
    test.assert(
        ordinary_playback.frame ~= faster_playback.frame,
        [=[ordinary_playback.frame ~= faster_playback.frame]=]
    )
    test.expect(ordinary_playback.seconds):equals(faster_playback.seconds)

    local displaced = player_recipe(3)
    displaced.velocity_x = 999
    test.expect(adapter.unarmed(displaced).rate):equals(ordinary_composite.rate)

    local running = player_recipe(3)
    running.mode = "RN"
    test.expect(adapter.unarmed(running).rate):equals(151)

    local retained = adapter.new_playback(ordinary_composite)
    adapter.synchronize(retained, ordinary_composite, 0.02)
    local previous_frame_seconds = 256 / (ordinary_composite.rate * 25)
    local faster_frame_seconds = 256 / (faster_composite.rate * 25)
    local expected_phase = 0.02 / previous_frame_seconds + 0.001 / faster_frame_seconds
    adapter.synchronize(retained, faster_composite, 0.021)
    test.assert(
        math.abs(retained.remainder / faster_frame_seconds - expected_phase) < 0.0000001,
        [=[math.abs(retained.remainder / faster_frame_seconds - expected_phase) < 0.0000001]=]
    )
    test.assert(math.abs(retained.seconds - expected_phase * faster_frame_seconds) < 0.0000001)
    test.expect(retained.authority_seconds):equals(0.021)
end

return test.suite({
    profile = "module",
    tier = "fast",
    covers = { "internal/game/player" },
    cases = {
        test.case("resolves_cof_layer_weapon_classes", {
            { run = resolves_cof_layer_weapon_classes },
        }),
        test.case("movement_stats_not_integrated_distance_drive_animation_playback", {
            { run = movement_stats_not_integrated_distance_drive_animation_playback },
        }),
    },
})
