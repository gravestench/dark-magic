local test = require("d2legacy.tests/v1")

return test.suite({
    name = "monster MonStats resistances",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("each_difficulty_materializes_all_six_resistances", function(t)
            t:run(function()
                local difficulty = 0
                local row = {
                    Id = "fallen",
                    MonStatsEx = "fallen",
                    enabled = "1",
                    isSpawn = "1",
                    noRatio = "1",
                    Level = "1",
                    ["Level(N)"] = "1",
                    ["Level(H)"] = "1",
                    minHP = "16",
                    maxHP = "16",
                    ["minHP(N)"] = "16",
                    ["maxHP(N)"] = "16",
                    ["minHP(H)"] = "16",
                    ["maxHP(H)"] = "16",
                    ResDm = "10",
                    ResMa = "20",
                    ResFi = "30",
                    ResLi = "40",
                    ResCo = "50",
                    ResPo = "60",
                    ["ResDm(N)"] = "11",
                    ["ResMa(N)"] = "21",
                    ["ResFi(N)"] = "31",
                    ["ResLi(N)"] = "41",
                    ["ResCo(N)"] = "51",
                    ["ResPo(N)"] = "61",
                    ["ResDm(H)"] = "12",
                    ["ResMa(H)"] = "22",
                    ["ResFi(H)"] = "32",
                    ["ResLi(H)"] = "42",
                    ["ResCo(H)"] = "52",
                    ["ResPo(H)"] = "102",
                }
                test.mock_module("engine.records/v1", {
                    load = function(path)
                        if string.find(path, "monstats2", 1, true) then
                            return { { Id = "fallen", SizeX = "2", SizeY = "2", MeleeRng = "1" } }
                        end
                        if string.find(path, "monlvl", 1, true) then
                            return { { Level = "1" } }
                        end
                        return { row }
                    end,
                }, { "load" })
                test.mock_module("d2legacy.policy.game_rules", {
                    difficulty = function()
                        return difficulty
                    end,
                }, { "difficulty" })
                local monster = require("d2legacy.data.monster")
                local names = {
                    "physical_resist",
                    "magic_resist",
                    "fire_resist",
                    "lightning_resist",
                    "cold_resist",
                    "poison_resist",
                }
                local expected = {
                    { 10, 20, 30, 40, 50, 60 },
                    { 11, 21, 31, 41, 51, 61 },
                    { 12, 22, 32, 42, 52, 102 },
                }
                for index = 1, 3 do
                    difficulty = index - 1
                    local definition = monster.load("fallen")
                    for channel, name in ipairs(names) do
                        test.expect(definition[name]):equals(expected[index][channel])
                    end
                end
            end)
        end),
    },
})
