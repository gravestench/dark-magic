local test = require("d2legacy.tests/v1")

local function stats(id, spawn)
    local row = {
        Id = id,
        BaseId = "base",
        NameStr = "name",
        AI = "ai",
        Align = "1",
        Code = "xx",
        MonStatsEx = "graphics",
        Velocity = "6",
        aidel = "4",
        aidist = "33",
        boss = "0",
        primeevil = "0",
        MinGrp = "2",
        MaxGrp = "4",
        Rarity = "3",
        enabled = "1",
        isSpawn = spawn,
        npc = "0",
        noRatio = "0",
    }
    for difficulty, suffix in ipairs({ "", "(N)", "(H)" }) do
        for index, name in ipairs({ "minHP", "maxHP", "AC", "A1TH", "A1MinD", "A1MaxD", "Exp", "Drain" }) do
            row[name .. suffix] = tostring(10 * difficulty + index)
        end
        for index, name in ipairs({ "ResDm", "ResMa", "ResFi", "ResLi", "ResCo", "ResPo" }) do
            row[name .. suffix] = tostring(40 * index + difficulty - 50)
        end
        row["Level" .. suffix] = tostring(difficulty)
        row["TreasureClass1" .. suffix] = "reward" .. difficulty
    end
    return row
end

local graphics = {
    Id = "graphics",
    SizeX = "3",
    SizeY = "4",
    OverlayHeight = "2",
    BaseW = "hth",
    MeleeRng = "3",
    mKB = "1",
    small = "1",
    large = "0",
    corpseSel = "1",
    revive = "1",
}
local component_names =
    { "HD", "TR", "LG", "RA", "LA", "RH", "LH", "SH", "S1", "S2", "S3", "S4", "S5", "S6", "S7", "S8" }
for _, name in ipairs(component_names) do
    graphics[name], graphics[name .. "v"] = "1", "lit"
end
local levels = {}
for level = 1, 5 do
    local row = { Level = tostring(level) }
    for _, suffix in ipairs({ "", "(N)", "(H)" }) do
        for index, name in ipairs({ "HP", "AC", "TH", "DM", "XP", "L-AC", "L-TH" }) do
            row[name .. suffix] = tostring(100 * level + index)
        end
    end
    levels[#levels + 1] = row
end
local npc = stats("npc", "0")
npc.npc = "1"
local ordinary = stats("ordinary", "1")
ordinary.demon = "1"
ordinary.hUndead = "1"
local pet = stats("pet", "0")
pet.lUndead = "1"
local neutral = stats("neutral", "1")

return test.suite({
    name = "monster catalog indexing",
    profile = "module",
    tier = "fast",
    records = {
        ["data/global/excel/monstats.txt"] = { ordinary, pet, neutral, npc },
        ["data/global/excel/monstats2.txt"] = { graphics },
        ["data/global/excel/monlvl.txt"] = levels,
    },
    cases = {
        test.case("authored_type_flags_survive_population_summon_and_revive_definitions", function(t)
            t:run(function()
                test.mock_module("d2legacy.policy.game_rules", {
                    difficulty = function()
                        return 0
                    end,
                }, { "difficulty" })
                test.unload_module("d2legacy.data.monster")
                local monster = require("d2legacy.data.monster")
                local loaded = monster.load("ordinary")
                test.expect(loaded.demon):equals(true)
                test.expect(loaded.low_undead):equals(false)
                test.expect(loaded.high_undead):equals(true)
                local summoned = monster.summon("pet", 2)
                test.expect(summoned.demon):equals(false)
                test.expect(summoned.low_undead):equals(true)
                test.expect(summoned.high_undead):equals(false)
                local revived = monster.revive("ordinary", 2)
                test.expect(revived.demon):equals(true)
                test.expect(revived.low_undead):equals(false)
                test.expect(revived.high_undead):equals(true)
                local missing = monster.load("neutral")
                test.expect(missing.demon):equals(false)
                test.expect(missing.low_undead):equals(false)
                test.expect(missing.high_undead):equals(false)
            end)
        end),
        test.case("selected_columns_preserve_full_consumer_results", function(t)
            t:run(function()
                local records = require("engine.records/v1")
                local function snapshot(difficulty, selected)
                    test.mock_module("engine.records/v1", {
                        load = function(path, columns)
                            if selected then
                                return records.load(path, columns)
                            end
                            return records.load(path)
                        end,
                    }, { "load" })
                    test.mock_module("d2legacy.policy.game_rules", {
                        difficulty = function()
                            return difficulty
                        end,
                    }, { "difficulty" })
                    test.unload_module("d2legacy.data.monster")
                    local monster = require("d2legacy.data.monster")
                    return {
                        ordinary = monster.load("ordinary"),
                        area = monster.load("ordinary", { MonLvl2Ex = "4", MonLvl3Ex = "5" }),
                        summoned = monster.summon("pet", 5),
                        revived = monster.revive("ordinary", 4),
                        npc = monster.npc_appearance("npc"),
                        all = monster.all(),
                    }
                end
                for value = 0, 2 do
                    local full = snapshot(value, false)
                    local projected = snapshot(value, true)
                    test.expect(projected):deep_equals(full)
                    for index, name in ipairs({
                        "physical_resist",
                        "magic_resist",
                        "fire_resist",
                        "lightning_resist",
                        "cold_resist",
                        "poison_resist",
                    }) do
                        test.expect(projected.ordinary[name]):equals(40 * index + value - 49)
                    end
                    projected.ordinary.components.HD = "changed"
                    test.expect(full.ordinary.components.HD):equals("LIT")
                    test.expect(snapshot(value, true)):deep_equals(full)
                end
            end)
        end),
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
                test.mock_module("d2legacy.policy.game_rules", {
                    difficulty = function()
                        return 0
                    end,
                }, { "difficulty" })
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
