local test = require("d2legacy.tests/v1")

local function outdoor_records()
    local prest = {}
    for _, definition in ipairs({ 17, 26, 27, 28, 29, 30, 35, 51 }) do
        local row = {
            Def = tostring(definition),
            SizeX = "8",
            SizeY = "8",
            Files = "1",
            File1 = "Act1/Outdoors/fill" .. definition .. ".ds1",
            Dt1Mask = "1",
            Populate = "1",
        }
        if definition >= 26 and definition <= 28 or definition == 51 then
            row.Files = "4"
            for variant = 2, 4 do
                row["File" .. variant] = "Act1/Outdoors/structure" .. definition .. "-" .. variant .. ".ds1"
            end
        end
        prest[#prest + 1] = row
    end
    local rows = {
        ["data/global/excel/Levels.txt"] = {
            {
                Id = "2",
                Act = "0",
                DrlgType = "3",
                LevelType = "2",
                SizeX = "80",
                SizeY = "80",
                ["SizeX(N)"] = "88",
                ["SizeY(N)"] = "88",
                ["SizeX(H)"] = "96",
                ["SizeY(H)"] = "96",
                Vis3 = "8",
                Warp3 = "0",
            },
            { Id = "8", Vis0 = "2", Warp0 = "4" },
        },
        ["data/global/excel/LvlTypes.txt"] = {
            {},
            {},
            { ["File 1"] = "Act1/Outdoors/Outdoor1.dt1" },
        },
        ["data/global/excel/LvlPrest.txt"] = prest,
    }
    return rows
end

local function load_outdoor(rows)
    test.mock_module("engine.records/v1", {
        load = function(path)
            return rows[path] or {}
        end,
    }, { "load" })
    test.unload_module("d2legacy.mapgen.outdoor")
    test.unload_module("d2legacy.mapgen.records")
    test.unload_module("d2legacy.mapgen.outdoor_structures")
    return require("d2legacy.mapgen.outdoor")
end

local function verify(direction)
    local outdoor = load_outdoor(outdoor_records())
    local zone = outdoor.generate(2, 42, direction, 0)
    test.assert(
        zone.kind == "outdoor" and #zone.rooms == 100 and #zone.links == 180 and #zone.stamps >= 120,
        [=[zone.kind == "outdoor" and #zone.rooms == 100 and #zone.links == 180 and #zone.stamps >= 120]=]
    )
    test.assert(
        zone.checksum == outdoor.generate(2, 42, direction, 0).checksum,
        [=[zone.checksum == outdoor.generate(2, 42, direction, 0).checksum]=]
    )
    local path = {}
    for _, tile in ipairs(zone.paths) do
        path[tile.x .. ":" .. tile.y] = true
    end
    test.assert(
        path[zone.warps[1].x .. ":" .. zone.warps[1].y] and path[zone.warps[2].x .. ":" .. zone.warps[2].y],
        [=[path[zone.warps[1].x .. ":" .. zone.warps[1].y] and path[zone.warps[2].x .. ":" .. zone.warps[2].y]]=]
    )
    local bridge, passable = 0, 0
    for _, tile in ipairs(zone.structures) do
        if path[tile.x .. ":" .. tile.y] then
            test.assert(tile.passable, [=[tile.passable]=])
        end
        if tile.kind == "bridge" then
            bridge = bridge + 1
            if tile.passable then
                passable = passable + 1
            end
        end
    end
    test.assert(bridge == 64 and passable == 48, [=[bridge == 64 and passable == 48]=])
end

local function equal(first, second)
    if type(first) ~= "table" or type(second) ~= "table" then
        return first == second
    end
    for key, value in pairs(first) do
        if not equal(value, second[key]) then
            return false
        end
    end
    for key in pairs(second) do
        if first[key] == nil then
            return false
        end
    end
    return true
end

-- Missing stamps, shifted earlier draw streams, blocked approaches, or a cave
-- placed over either outdoor join must fail at the admitted recipe boundary.
local function verify_den()
    local outdoor = load_outdoor(outdoor_records())
    for difficulty = 0, 2 do
        for _, direction in ipairs({ "north", "east", "south", "west" }) do
            for _, seed in ipairs({ 0, 42, 73 }) do
                local legacy = outdoor.generate(2, seed, direction, difficulty)
                local zone = outdoor.generate_with_den(2, seed, direction, difficulty)
                test.assert(
                    zone.checksum == outdoor.generate_with_den(2, seed, direction, difficulty).checksum,
                    "Den placement and variant replay deterministically"
                )
                test.assert(
                    outdoor.generate(2, seed, direction, difficulty).checksum == legacy.checksum,
                    "legacy generator remains unchanged after v4 generation"
                )
                local cave, count = nil, 0
                for _, stamp in ipairs(zone.stamps) do
                    if stamp.preset_def == 51 then
                        cave, count = stamp, count + 1
                    end
                end
                test.assert(
                    count == 1 and cave.width == 8 and cave.height == 8 and cave.overlay,
                    "one authored 8x8 cave overlay"
                )
                test.assert(
                    cave.ds1_path:match("%.ds1$") and #cave.tile_paths == 1,
                    "cave joins declared DS1 and DT1 resources"
                )
                test.assert(#zone.stamps == #legacy.stamps + 1, "one additional stamp only")
                for index, stamp in ipairs(legacy.stamps) do
                    test.assert(equal(stamp, zone.stamps[index]), "earlier fill and structural draw streams unchanged")
                end
                test.assert(
                    equal(zone.rooms, legacy.rooms) and equal(zone.links, legacy.links),
                    "outdoor topology unchanged"
                )
                test.assert(equal(zone.structures, legacy.structures), "river and cliff crossings unchanged")
                test.assert(#zone.warps == 3, "one cave warp added")
                test.assert(equal(zone.warps[1], legacy.warps[1]), "town join unchanged")
                test.assert(equal(zone.warps[2], legacy.warps[2]), "Cold Plains join unchanged")
                local warp = zone.warps[3]
                test.assert(warp.role == "den-entrance" and warp.destination_level == 8, "unique semantic Den entrance")
                test.assert(
                    cave.x >= 8
                        and cave.y >= 8
                        and cave.x + 8 <= zone.bounds.width - 8
                        and cave.y + 8 <= zone.bounds.height - 8,
                    "cave leaves both outdoor edge joins reserved"
                )
                for _, stamp in ipairs(legacy.stamps) do
                    if stamp.overlay then
                        test.assert(
                            cave.x + 8 <= stamp.x
                                or stamp.x + stamp.width <= cave.x
                                or cave.y + 8 <= stamp.y
                                or stamp.y + stamp.height <= cave.y,
                            "cave does not overlap river or cliff artwork"
                        )
                    end
                end
                local paths, old_paths, blocked = {}, {}, {}
                for _, tile in ipairs(zone.paths) do
                    paths[tile.x .. ":" .. tile.y] = true
                end
                for _, tile in ipairs(legacy.paths) do
                    old_paths[tile.x .. ":" .. tile.y] = true
                    test.assert(paths[tile.x .. ":" .. tile.y], "established outdoor path retained")
                    test.assert(
                        tile.x < cave.x or tile.x >= cave.x + 8 or tile.y < cave.y or tile.y >= cave.y + 8,
                        "cave does not overwrite the established route"
                    )
                end
                for _, tile in ipairs(zone.structures) do
                    if not tile.passable then
                        blocked[tile.x .. ":" .. tile.y] = true
                    end
                end
                -- A three-tile-wide corridor must connect the authored cell's
                -- southern approach to a pre-existing route cell. One tile is
                -- five subtiles, wider than the current player diameter.
                local queue, visited, connected = { { x = cave.x + 4, y = cave.y + 9 } }, {}, false
                for index = 1, zone.bounds.width * zone.bounds.height do
                    local tile = queue[index]
                    if not tile then
                        break
                    end
                    if old_paths[tile.x .. ":" .. tile.y] then
                        connected = true
                        break
                    end
                    for _, delta in ipairs({ { 1, 0 }, { -1, 0 }, { 0, 1 }, { 0, -1 } }) do
                        local x, y = tile.x + delta[1], tile.y + delta[2]
                        local key, clear = x .. ":" .. y, true
                        for dy = -1, 1 do
                            for dx = -1, 1 do
                                local neighbor = (x + dx) .. ":" .. (y + dy)
                                if not paths[neighbor] or blocked[neighbor] then
                                    clear = false
                                end
                            end
                        end
                        if clear and not visited[key] then
                            visited[key] = true
                            queue[#queue + 1] = { x = x, y = y }
                        end
                    end
                end
                test.assert(connected, "player-sized cave approach connects to the established route")
            end
        end
    end
end

return test.suite({
    profile = "module",
    tier = "fast",
    covers = { "internal/mod/d2legacy/mapgen" },
    disable_execution_budget = true,
    cases = {
        test.case("TestBloodMoorDenEntranceRecipe", { test.run(verify_den) }),
        test.case("north_recipe_and_structures", {
            test.run(function()
                verify("north")
            end),
        }),
        test.case("east_recipe_and_structures", {
            test.run(function()
                verify("east")
            end),
        }),
        test.case("south_recipe_and_structures", {
            test.run(function()
                verify("south")
            end),
        }),
        test.case("west_recipe_and_structures", {
            test.run(function()
                verify("west")
            end),
        }),
    },
})
