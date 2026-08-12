local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "policy",
    tier = "fast",
    tests = {
        owns_preset_generation_policy = {
            {
                run = function()
                    local rows = {
                        ["data/global/excel/Levels.txt"] = {
                            {
                                Id = "1",
                                Act = "0",
                                DrlgType = "2",
                                LevelType = "1",
                                SizeX = "40",
                                SizeY = "30",
                                ["SizeX(N)"] = "50",
                                ["SizeY(N)"] = "35",
                            },
                        },
                        ["data/global/excel/LvlPrest.txt"] = {
                            {
                                Def = "7",
                                LevelId = "1",
                                Files = "0",
                                File1 = "Act1\\Town\\TownN1.ds1",
                                File2 = "Act1\\Town\\TownE1.ds1",
                                File3 = "Act1\\Town\\TownS1.ds1",
                                File4 = "Act1\\Town\\TownW1.ds1",
                                Dt1Mask = "5",
                                Populate = "1",
                                Logicals = "1",
                            },
                        },
                        ["data/global/excel/LvlTypes.txt"] = {
                            {},
                            {
                                ["File 1"] = "Act1\\Town\\floor.dt1",
                                ["File 2"] = "ignored.dt1",
                                ["File 3"] = "Act1\\Town\\wall.dt1",
                            },
                        },
                    }
                    package.loaded["engine.records/v1"] = {
                        load = function(path)
                            return rows[path] or {}
                        end,
                    }
                    package.loaded["d2legacy.mapgen.preset"] = nil
                    local preset = require("d2legacy.mapgen.preset")
                    local first = preset.generate(1, 11, 0)
                    local again = preset.generate(1, 11, 0)
                    assert(first.checksum == again.checksum)
                    assert(first.kind == "preset" and first.bounds.width == 40 and first.bounds.height == 30)
                    assert(first.stamps[1].preset_def == 7 and first.stamps[1].populate)
                    assert(
                        #first.stamps[1].tile_paths == 2
                            and first.stamps[1].tile_paths[1] == "data/global/tiles/Act1/Town/floor.dt1"
                    )
                    assert(first.stamps[1].role:match("^act1%-town:exit%-"))
                end,
            },
        },
    },
})
