local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    covers = { "internal/mod/d2legacy/mapgen" },
    cases = {
        test.case("caches_and_indexes_immutable_record_tables", {
            test.run(function()
                local loads = {}
                local rows = {
                    ["data/global/excel/Levels.txt"] = { { Id = "2" } },
                    ["data/global/excel/LvlPrest.txt"] = {
                        { Def = "29", LevelId = "2" },
                        { Def = "30", LevelId = "3" },
                    },
                    ["data/global/excel/LvlTypes.txt"] = { {}, {}, { File1 = "outdoor.dt1" } },
                }
                test.mock_module("engine.records/v1", {
                    load = function(path)
                        loads[path] = (loads[path] or 0) + 1
                        return rows[path] or {}
                    end,
                }, { "load" })
                test.unload_module("d2legacy.mapgen.records")
                local records = require("d2legacy.mapgen.records")

                for _ = 1, 100 do
                    test.assert(records.preset_by_definition(29).Def == "29")
                    test.assert(records.preset_for_level(2).LevelId == "2")
                end
                test.assert(records.level(2).Id == "2")
                test.assert(records.level_type(2).File1 == "outdoor.dt1")
                test.assert(loads["data/global/excel/LvlPrest.txt"] == 1)
                test.assert(loads["data/global/excel/Levels.txt"] == 1)
                test.assert(loads["data/global/excel/LvlTypes.txt"] == 1)
            end),
        }),
    },
})
