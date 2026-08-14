local test = require("d2legacy.tests/v1")

return test.suite({
    name = "monster catalog indexing",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("legacy_tables_are_loaded_once_per_vm", function(t)
            t:run(function()
                local loads = {}
                test.mock_module("engine.records/v1", {
                    load = function(path)
                        loads[path] = (loads[path] or 0) + 1
                        if string.find(path, "monstats2", 1, true) then
                            return { { Id = "fallen", SizeX = "2", SizeY = "2", MeleeRng = "1" } }
                        end
                        if string.find(path, "monlvl", 1, true) then
                            return { { Level = "1", HP = "100", AC = "100", TH = "100", DM = "100", XP = "100" } }
                        end
                        return {
                            {
                                Id = "fallen",
                                BaseId = "fallen",
                                MonStatsEx = "fallen",
                                enabled = "1",
                                isSpawn = "1",
                                npc = "0",
                                Level = "1",
                                Rarity = "1",
                                MinGrp = "1",
                                MaxGrp = "1",
                            },
                        }
                    end,
                }, { "load" })

                local monster = require("d2legacy.data.monster")
                monster.load("fallen", 0)
                monster.load("fallen", 0)

                test.expect(loads["data/global/excel/monstats.txt"]):equals(1)
                test.expect(loads["data/global/excel/monstats2.txt"]):equals(1)
                test.expect(loads["data/global/excel/monlvl.txt"]):equals(1)
            end)
        end),
    },
})
