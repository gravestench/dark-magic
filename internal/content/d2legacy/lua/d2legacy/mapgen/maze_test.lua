local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "policy",
    tier = "fast",
    tests = {
        owns_act_one_cave_maze_policy = {
            {
                run = function()
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
                    package.loaded["engine.records/v1"] = {
                        load = function(path)
                            return rows[path] or {}
                        end,
                    }
                    package.loaded["d2legacy.mapgen.maze"] = nil
                    local maze = require("d2legacy.mapgen.maze")
                    local zone = maze.generate(9, 42, 0)
                    assert(zone.kind == "maze" and #zone.rooms == 12 and #zone.stamps == 12 and #zone.links >= 11)
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
                        assert(not occupied[key])
                        occupied[key] = true
                    end
                    assert(previous == 1 and next_level == 1)
                    assert(zone.checksum == maze.generate(9, 42, 0).checksum)
                    assert(#maze.generate(9, 42, 1).rooms == 14)
                end,
            },
        },
    },
})
