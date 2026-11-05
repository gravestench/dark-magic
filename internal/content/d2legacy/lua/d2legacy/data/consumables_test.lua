local test = require("d2legacy.tests/v1")
return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("all_classes_and_critical_threshold_use_raw_amount", {
            test.run(function()
                local data = require("d2legacy.data.consumables")
                local definition = { amount_raw = 7680 }
                local zero = {
                    integer = function()
                        error("zero Vitality must not roll")
                    end,
                }
                for class, expected in pairs({
                    Amazon = 11520,
                    Paladin = 11520,
                    Assassin = 11520,
                    Barbarian = 15360,
                    Sorceress = 7680,
                    Necromancer = 7680,
                    Druid = 7680,
                }) do
                    test.expect(data.amount(definition, class, 0, zero)):equals(expected)
                end
                local function rolls(percent)
                    return {
                        integer = function(stream, bound)
                            if stream == "d2legacy.potion.critical" then
                                test.expect(bound):equals(100)
                                return percent
                            end
                            test.expect(bound):equals(20)
                            return 19
                        end,
                    }
                end
                test.expect(data.amount(definition, "Barbarian", 20, rolls(8))):equals(30720)
                test.expect(data.amount(definition, "Barbarian", 20, rolls(9))):equals(15360)
            end),
        }),
        test.case("unsupported_formula_state_and_family_refuse", {
            test.run(function()
                local row = {
                    code = "hp1",
                    pSpell = "3",
                    stat1 = "hpregen",
                    state = "healthpot",
                    calc1 = "30",
                    len = "175",
                }
                test.mock_module("engine.records/v1", {
                    load = function()
                        return { row }
                    end,
                })
                local data = require("d2legacy.data.consumables")
                test.expect(data.load("hp1")):deep_equals({ amount_raw = 7680, duration = 175 })
                test.expect(data.load("hp2")):is_nil()
                for key, value in pairs({
                    calc1 = "lvl*10",
                    len = "0",
                    pSpell = "5",
                    state = "other",
                    stat2 = "mana",
                }) do
                    local old = row[key]
                    row[key] = value
                    test.expect(data.load("hp1")):is_nil()
                    row[key] = old
                end
            end),
        }),
    },
})
