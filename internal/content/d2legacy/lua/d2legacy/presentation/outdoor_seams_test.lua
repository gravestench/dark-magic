local test = require("d2legacy.tests/v1")

local function equal(actual, expected)
    test.expect(actual):equals(expected)
end

-- Synthetic maps use the native map projection: 16 pixels horizontally and
-- 8 vertically per subtile. Origins differ with the map's tile height.
local function entry(level, origin, width, height)
    local map = {}
    function map:subtile_to_pixel(x, y)
        return origin + (x - y) * 16, 168 + (x + y) * 8
    end
    function map:canvas()
        return width, height
    end
    return { level_id = level, map = map, ds1 = "zone" .. level, dt1 = { "tiles" .. level } }
end

local function fixture()
    local snapshot = {
        active_level = 2,
        maps = { entry(1, 320, 960, 640), entry(2, 480, 1280, 800), entry(3, 640, 1600, 960) },
    }
    local seams = {
        { source_level = 1, destination_level = 2, source_x = 20, source_y = 10, arrival_x = 0, arrival_y = 10 },
        { source_level = 2, destination_level = 1, source_x = 0, source_y = 10, arrival_x = 20, arrival_y = 10 },
        { source_level = 2, destination_level = 3, source_x = 30, source_y = 20, arrival_x = 0, arrival_y = 0 },
        { source_level = 3, destination_level = 2, source_x = 0, source_y = 0, arrival_x = 30, arrival_y = 20 },
    }
    return snapshot, seams
end

local function rejected(snapshot, seams)
    local helper = require("d2legacy.presentation.outdoor_seams")
    local ok = pcall(helper.layout, snapshot, seams)
    test.assert(not ok, "invalid presentation layout must be rejected")
end

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("aligns_both_prepared_seams_and_unions_canvases", {
            test.run(function()
                local helper = require("d2legacy.presentation.outdoor_seams")
                local snapshot, seams = fixture()
                local layout = helper.layout(snapshot, seams)
                local x, y = helper.pixel(2, 320, 248, layout)
                equal(x, 480)
                equal(y, 408)
                x, y = helper.pixel(3, 640, 168, layout)
                equal(x, 800)
                equal(y, 728)
                equal(layout.bounds.min_x, 0)
                equal(layout.bounds.min_y, 0)
                equal(layout.bounds.max_x, 1760)
                equal(layout.bounds.max_y, 1520)
                equal(layout.bounds.width, 1760)
                equal(layout.bounds.height, 1520)
                equal(layout.levels[2].map, snapshot.maps[2].map)
                equal(layout.levels[2].recipe.ds1, "zone2")
                equal(layout.levels[2].recipe.dt1[1], "tiles2")
                equal(layout.levels[2].recipe.world, snapshot.maps[2].map)
            end),
        }),
        test.case("fourth_zone_aligns_and_interleaves_translated_art_in_either_active_map", {
            test.run(function()
                local helper = require("d2legacy.presentation.outdoor_seams")
                local snapshot, seams = fixture()
                snapshot.maps[4] = entry(4, 800, 1920, 1120)
                seams[#seams + 1] = {
                    source_level = 3,
                    destination_level = 4,
                    source_x = 40,
                    source_y = 30,
                    arrival_x = 0,
                    arrival_y = 0,
                }
                seams[#seams + 1] = {
                    source_level = 4,
                    destination_level = 3,
                    source_x = 0,
                    source_y = 0,
                    arrival_x = 40,
                    arrival_y = 30,
                }
                for _, active in ipairs({ 3, 4 }) do
                    snapshot.active_level = active
                    local layout = helper.layout(snapshot, seams)
                    local x, y = helper.pixel(4, 800, 168, layout)
                    equal(x, 960)
                    equal(y, 1288)
                    x, y = helper.pixel(3, 800, 728, layout)
                    equal(x, 960)
                    equal(y, 1288)
                    equal(layout.bounds.max_x, 2080)
                    equal(layout.bounds.max_y, 2240)
                    -- A Plains wall at 1900 and Stony wall at 500 share
                    -- world depth 3300; a Stony entity at 505 is between
                    -- those walls and a Plains wall at 1910.
                    equal(helper.depth(3, 1900, layout), 3300)
                    equal(helper.depth(4, 500, layout), 3300)
                    equal(helper.depth(4, 505, layout), 3305)
                    equal(helper.depth(3, 1910, layout), 3310)
                    for _, depth in ipairs({ -1000002, -1000001, -1000000, 1000000 }) do
                        equal(helper.depth(4, depth, layout), depth)
                    end
                end
            end),
        }),
        test.case("placement_is_stable_across_map_order_and_active_level", {
            test.run(function()
                local helper = require("d2legacy.presentation.outdoor_seams")
                local snapshot, seams = fixture()
                snapshot.maps = { snapshot.maps[3], snapshot.maps[1], snapshot.maps[2] }
                snapshot.active_level = 3
                local layout = helper.layout(snapshot, { seams[4], seams[2], seams[3], seams[1] })
                equal(layout.levels[1].offset_x, 0)
                equal(layout.levels[2].offset_x, 160)
                equal(layout.levels[2].offset_y, 160)
                equal(layout.levels[3].offset_y, 560)
            end),
        }),
        test.case("two_zone_snapshot_ignores_unprepared_neighbors", {
            test.run(function()
                local helper = require("d2legacy.presentation.outdoor_seams")
                local snapshot, seams = fixture()
                snapshot.maps[3] = nil
                local layout = helper.layout(snapshot, seams)
                equal(layout.levels[2].offset_y, 160)
                equal(layout.bounds.max_x, 1440)
                equal(layout.bounds.max_y, 960)
                equal(layout.levels[3], nil)
            end),
        }),
        test.case("rejects_reverse_alignment_residual", {
            test.run(function()
                local snapshot, seams = fixture()
                seams[2].arrival_x = 21
                rejected(snapshot, seams)
            end),
        }),
        test.case("allows_subpixel_reverse_alignment_residual", {
            test.run(function()
                local helper = require("d2legacy.presentation.outdoor_seams")
                local snapshot, seams = fixture()
                seams[2].arrival_x = 20.01
                local layout = helper.layout(snapshot, seams)
                equal(layout.levels[2].offset_x, 160)
            end),
        }),
        test.case("rejects_missing_reverse_and_disconnected_maps", {
            test.run(function()
                local snapshot, seams = fixture()
                rejected(snapshot, { seams[1], seams[3], seams[4] })
                rejected(snapshot, { seams[1], seams[2] })
            end),
        }),
        test.case("rejects_malformed_geometry_and_snapshot", {
            test.run(function()
                local snapshot, seams = fixture()
                seams[1].source_x = 0 / 0
                rejected(snapshot, seams)
                snapshot, seams = fixture()
                snapshot.maps[2].level_id = 1
                rejected(snapshot, seams)
                snapshot, seams = fixture()
                snapshot.active_level = 9
                rejected(snapshot, seams)
                snapshot, seams = fixture()
                snapshot.maps[2].map.canvas = function()
                    return -1, 100
                end
                rejected(snapshot, seams)
            end),
        }),
        test.case("neighboring_upper_walls_interleave_with_entity_and_keep_pass_bands", {
            test.run(function()
                local helper = require("d2legacy.presentation.outdoor_seams")
                local snapshot, seams = fixture()
                local layout = helper.layout(snapshot, seams)
                -- Town wall depth 900, neighbor wall 500 + translated 400;
                -- neighbor entity at 505 is between that wall and town's 1000.
                equal(helper.depth(2, 500, layout), 900)
                equal(helper.depth(1, 900, layout), 900)
                equal(helper.depth(2, 505, layout), 905)
                test.assert(helper.depth(1, 900, layout) < helper.depth(2, 505, layout))
                test.assert(helper.depth(2, 505, layout) < helper.depth(1, 1000, layout))
                for _, depth in ipairs({ -1000002, -1000001, -1000000, 1000000 }) do
                    equal(helper.depth(3, depth, layout), depth)
                end
            end),
        }),
    },
})
