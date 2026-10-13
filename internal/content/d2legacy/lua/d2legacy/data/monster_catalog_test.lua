local test = require("d2legacy.tests/v1")

return test.suite({
    name = "monster catalog indexing",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("npc_appearance_rejects_invalid_overlay_height", function(t)
            t:run(function()
                test.mock_module("engine.records/v1", {
                    load = function(path)
                        if string.find(path, "monstats2", 1, true) then
                            return { { Id = "akara", OverlayHeight = "5" } }
                        end
                        if string.find(path, "monlvl", 1, true) then
                            return { { Level = "1" } }
                        end
                        return { { Id = "akara", MonStatsEx = "akara", npc = "1" } }
                    end,
                }, { "load" })
                local ok, message = pcall(require("d2legacy.data.monster").npc_appearance, "akara")
                test.expect(ok):equals(false)
                test.expect(string.find(message, "OverlayHeight must be in 0..4", 1, true) ~= nil):equals(true)
            end)
        end),
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
                                    OverlayHeight = "3",
                                    mKB = "1",
                                    small = "1",
                                    corpseSel = "1",
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
                test.expect(definition.overlay_height):equals(3)
                test.expect(definition.corpse_selectable):equals(true)

                test.expect(loads["data/global/excel/monstats.txt"]):equals(1)
                test.expect(loads["data/global/excel/monstats2.txt"]):equals(1)
                test.expect(loads["data/global/excel/monlvl.txt"]):equals(1)
            end)
        end),
        test.case("expansion_hell_monsters_take_the_area_level", function(t)
            t:run(function()
                local function stats(id, boss)
                    return {
                        Id = id,
                        BaseId = id,
                        MonStatsEx = id,
                        enabled = "1",
                        isSpawn = "1",
                        npc = "0",
                        boss = boss and "1" or "",
                        Level = "1",
                        ["Level(H)"] = boss and "75" or "67",
                        ["minHP(H)"] = "100",
                        ["maxHP(H)"] = "100",
                        ["AC(H)"] = "100",
                        ["A1TH(H)"] = "100",
                        ["A1MinD(H)"] = "100",
                        ["A1MaxD(H)"] = "100",
                        ["Exp(H)"] = "100",
                    }
                end
                local function level(value, life)
                    return {
                        Level = value,
                        ["HP(H)"] = life,
                        ["AC(H)"] = "1",
                        ["TH(H)"] = "1",
                        ["DM(H)"] = "1",
                        ["XP(H)"] = life,
                    }
                end
                test.mock_module("engine.records/v1", {
                    load = function(path)
                        if string.find(path, "monstats2", 1, true) then
                            return {
                                { Id = "fallen", SizeX = "2", SizeY = "2", MeleeRng = "1" },
                                { Id = "lord", SizeX = "2", SizeY = "2", MeleeRng = "1" },
                            }
                        end
                        if string.find(path, "monlvl", 1, true) then
                            return { level("67", "1000"), level("68", "2000"), level("75", "3000") }
                        end
                        return { stats("fallen", false), stats("lord", true) }
                    end,
                }, { "load" })
                test.mock_module("d2legacy.policy.game_rules", {
                    difficulty = function()
                        return 2
                    end,
                }, { "difficulty" })

                local monster = require("d2legacy.data.monster")
                local area = { Id = "3", MonLvl3Ex = "68" }
                -- An ordinary monster takes the area's Hell level and its MonLvl row.
                local fallen = monster.load("fallen", area)
                test.expect(fallen.level):equals(68)
                test.expect(fallen.life_min):equals(2000 * 256)
                test.expect(fallen.experience):equals(2000)
                -- Without an area it keeps its own MonStats Level(H).
                test.expect(monster.load("fallen").level):equals(67)
                -- Bosses ignore the area level.
                local lord = monster.load("lord", area)
                test.expect(lord.level):equals(75)
                test.expect(lord.life_min):equals(3000 * 256)
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
