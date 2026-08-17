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
                            return {
                                {
                                    Id = "fallen",
                                    SizeX = "2",
                                    SizeY = "2",
                                    MeleeRng = "1",
                                    mKB = "1",
                                    small = "1",
                                },
                            }
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
                test.mock_module("d2legacy.policy.game_rules", {
                    difficulty = function()
                        return 2
                    end,
                }, { "difficulty" })

                local monster = require("d2legacy.data.monster")
                local definition = monster.load("fallen")
                monster.load("fallen")

                test.expect(definition.knockback_mode):equals(true)
                test.expect(definition.knockback_size):equals("small")

                test.expect(loads["data/global/excel/monstats.txt"]):equals(1)
                test.expect(loads["data/global/excel/monstats2.txt"]):equals(1)
                test.expect(loads["data/global/excel/monlvl.txt"]):equals(1)
            end)
        end),
        test.case("difficulty_is_owned_by_immutable_game_rules", function(t)
            t:run(function()
                test.mock_module("engine.records/v1", {
                    load = function(path)
                        if string.find(path, "monstats2", 1, true) then
                            return { { Id = "fallen", SizeX = "2", SizeY = "2", MeleeRng = "1" } }
                        end
                        if string.find(path, "monlvl", 1, true) then
                            return {
                                {
                                    Level = "1",
                                    ["HP(H)"] = "200",
                                    ["AC(H)"] = "200",
                                    ["TH(H)"] = "200",
                                    ["DM(H)"] = "200",
                                    ["XP(H)"] = "200",
                                },
                            }
                        end
                        return {
                            {
                                Id = "fallen",
                                BaseId = "fallen",
                                MonStatsEx = "fallen",
                                enabled = "1",
                                isSpawn = "1",
                                npc = "0",
                                noRatio = "1",
                                Level = "1",
                                ["minHP(H)"] = "20",
                                ["maxHP(H)"] = "30",
                                ["AC(H)"] = "40",
                                ["A1TH(H)"] = "50",
                                ["A1MinD(H)"] = "6",
                                ["A1MaxD(H)"] = "7",
                                ["Exp(H)"] = "80",
                                ["TreasureClass1(H)"] = "Act 5 (H) H2H C",
                            },
                        }
                    end,
                }, { "load" })
                test.mock_module("d2legacy.policy.game_rules", {
                    difficulty = function()
                        return 2
                    end,
                }, { "difficulty" })

                local monster = require("d2legacy.data.monster")
                local definition = monster.load("fallen")
                test.expect(definition.life_min):equals(20 * 256)
                test.expect(definition.life_max):equals(30 * 256)
                test.expect(definition.experience):equals(80)
                test.expect(definition.treasure_class):equals("Act 5 (H) H2H C")

                local ok, message = pcall(monster.load, "fallen", 0)
                test.expect(ok):equals(false)
                test.assert(
                    string.find(message, "immutable game rules", 1, true) ~= nil,
                    [=[string.find(message, "immutable game rules", 1, true) ~= nil]=]
                )
            end)
        end),
    },
})
