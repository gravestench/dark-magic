local test = require("d2legacy.tests/v1")

local function den_records()
    local presets = {}
    for definition = 83, 86 do
        presets[#presets + 1] = {
            Def = tostring(definition),
            SizeX = "24",
            SizeY = "24",
            Files = "1",
            File1 = "Act1/Caves/previous.ds1",
            Dt1Mask = "1",
        }
    end
    return {
        ["data/global/excel/Levels.txt"] = {
            { Id = "2", Vis3 = "8", Warp3 = "0" },
            { Id = "8", Act = "0", DrlgType = "1", LevelType = "3", Vis0 = "2", Warp0 = "4" },
        },
        ["data/global/excel/LvlMaze.txt"] = {
            {
                Level = "8",
                Rooms = "1",
                ["Rooms(N)"] = "1",
                ["Rooms(H)"] = "1",
                SizeX = "24",
                SizeY = "24",
                Merge = "500",
            },
        },
        ["data/global/excel/LvlTypes.txt"] = { {}, {}, {}, { ["File 1"] = "Act1/Caves/floor.dt1" } },
        ["data/global/excel/LvlPrest.txt"] = presets,
    }
end

local function load_den(rows)
    test.mock_module("engine.records/v1", {
        load = function(path)
            return rows[path] or {}
        end,
    }, { "load" })
    test.unload_module("d2legacy.mapgen.maze")
    test.unload_module("d2legacy.mapgen.records")
    return require("d2legacy.mapgen.maze")
end

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("TestDenOfEvilOneRoomRecipe", {
            test.run(function()
                local maze = load_den(den_records())
                for difficulty = 0, 2 do
                    local zone = maze.generate(8, 42, difficulty)
                    test.assert(zone.request.level_id == 8, "Den level identity")
                    test.assert(
                        #zone.rooms == 1 and #zone.stamps == 1 and #(zone.links or {}) == 0,
                        "Den retains its single authored room"
                    )
                    test.assert(zone.bounds.width == 24 and zone.bounds.height == 24, "24x24 Den bounds")
                    local stamp = zone.stamps[1]
                    test.assert(
                        stamp.role == "previous-level" and stamp.width == 24 and stamp.height == 24,
                        "single previous-level chamber"
                    )
                    test.assert(
                        stamp.preset_def >= 83 and stamp.preset_def <= 86 and (stamp.variant or 0) == 0,
                        "authored previous-level definition and variant"
                    )
                    test.assert(
                        #zone.warps == 1 and zone.warps[1].role == "den-exit" and zone.warps[1].destination_level == 2,
                        "reciprocal Blood Moor exit"
                    )
                    test.assert(zone.checksum == maze.generate(8, 42, difficulty).checksum, "stable Den checksum")
                end
            end),
        }),
        test.case("TestDenOfEvilRejectsInvalidAuthoredVariants", {
            test.run(function()
                for _, mutation in ipairs({
                    function(rows)
                        rows["data/global/excel/LvlPrest.txt"] = {}
                    end,
                    function(rows)
                        rows["data/global/excel/LvlPrest.txt"][1].File1 = ""
                    end,
                    function(rows)
                        rows["data/global/excel/LvlPrest.txt"][1].Files = "2"
                    end,
                    function(rows)
                        rows["data/global/excel/LvlPrest.txt"][1].SizeX = "8"
                    end,
                    function(rows)
                        rows["data/global/excel/Levels.txt"][2].Vis0 = "3"
                    end,
                }) do
                    local rows = den_records()
                    mutation(rows)
                    local maze = load_den(rows)
                    local ok = pcall(maze.generate, 8, 42, 0)
                    test.assert(not ok, "invalid Den source records must fail before admission")
                end
            end),
        }),
        test.case("owns_act_one_cave_maze_policy", {
            test.run(function()
                local prest = {}
                for definition = 53, 67 do
                    prest[#prest + 1] = {
                        Def = tostring(definition),
                        SizeX = "24",
                        SizeY = "24",
                        Files = "1",
                        File1 = "Act1/Caves/room.ds1",
                        Dt1Mask = "1",
                    }
                end
                for definition = 83, 90 do
                    prest[#prest + 1] = {
                        Def = tostring(definition),
                        SizeX = "24",
                        SizeY = "24",
                        Files = "1",
                        File1 = "Act1/Caves/special.ds1",
                        Dt1Mask = "1",
                    }
                end
                local rows = {
                    ["data/global/excel/Levels.txt"] = {
                        { Id = "9", Act = "0", DrlgType = "1", LevelType = "3" },
                    },
                    ["data/global/excel/LvlMaze.txt"] = {
                        {
                            Level = "9",
                            Rooms = "12",
                            ["Rooms(N)"] = "14",
                            ["Rooms(H)"] = "16",
                            SizeX = "24",
                            SizeY = "24",
                            Merge = "500",
                        },
                    },
                    ["data/global/excel/LvlTypes.txt"] = {
                        {},
                        {},
                        {},
                        { ["File 1"] = "Act1/Caves/floor.dt1" },
                    },
                    ["data/global/excel/LvlPrest.txt"] = prest,
                }
                test.mock_module("engine.records/v1", {
                    load = function(path)
                        return rows[path] or {}
                    end,
                }, { "load" })
                test.unload_module("d2legacy.mapgen.maze")
                local maze = require("d2legacy.mapgen.maze")
                local zone = maze.generate(9, 42, 0)
                test.assert(
                    zone.kind == "maze" and #zone.rooms == 12 and #zone.stamps == 12 and #zone.links >= 11,
                    [=[zone.kind == "maze" and #zone.rooms == 12 and #zone.stamps == 12 and #zone.links >= 11]=]
                )
                local previous, next_level = 0, 0
                local occupied = {}
                for _, stamp in ipairs(zone.stamps) do
                    if stamp.role == "previous-level" then
                        previous = previous + 1
                    end
                    if stamp.role == "next-level" then
                        next_level = next_level + 1
                    end
                    local key = stamp.x .. ":" .. stamp.y
                    test.assert(not occupied[key], [=[not occupied[key]]=])
                    occupied[key] = true
                end
                test.assert(previous == 1 and next_level == 1, [=[previous == 1 and next_level == 1]=])
                test.assert(
                    zone.checksum == maze.generate(9, 42, 0).checksum,
                    [=[zone.checksum == maze.generate(9, 42, 0).checksum]=]
                )
                test.assert(#maze.generate(9, 42, 1).rooms == 14, [=[#maze.generate(9, 42, 1).rooms == 14]=])
            end),
        }),
    },
})
