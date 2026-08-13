local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("builds_development_items_from_raw_records", {
            test.run(function()
                local rows = {
                    ["data/global/excel/weapons.txt"] = {
                        {
                            code = "ssd",
                            invwidth = "1",
                            invheight = "3",
                            cost = "466",
                            invfile = "invssd",
                            flippyfile = "flpssd",
                            component = "RH",
                            alternategfx = "SSD",
                            wclass = "1hs",
                            rangeadder = "1",
                            mindam = "2",
                            maxdam = "7",
                        },
                    },
                    ["data/global/excel/armor.txt"] = {
                        {
                            code = "cap",
                            invwidth = "2",
                            invheight = "2",
                            cost = "64",
                            invfile = "invcap",
                            flippyfile = "flpcap",
                            component = "0",
                            alternategfx = "CAP",
                        },
                    },
                    ["data/global/excel/misc.txt"] = {
                        {
                            code = "hp1",
                            invwidth = "1",
                            invheight = "1",
                            cost = "30",
                            invfile = "invhp1",
                            flippyfile = "flphp1",
                        },
                        {
                            code = "mp1",
                            invwidth = "1",
                            invheight = "1",
                            cost = "30",
                            invfile = "invmp1",
                            flippyfile = "flpmp1",
                        },
                    },
                    ["data/global/excel/Npc.txt"] = {
                        {
                            npc = "Akara",
                            ["buy mult"] = "2048",
                            ["sell mult"] = "1024",
                            ["max buy"] = "5000",
                        },
                    },
                }
                test.mock_module("engine.records/v1", {
                    load = function(path)
                        return rows[path] or {}
                    end,
                }, { "load" })
                test.unload_module("d2legacy.items.development_fixtures")
                local fixtures = require("d2legacy.items.development_fixtures").build(true)
                test.assert(
                    fixtures.owner == "local-player"
                        and fixtures.inventory_width == 10
                        and fixtures.trade_terms.Akara.buy_multiplier == 2048
                )
                local by_id = {}
                for _, item in ipairs(fixtures.items) do
                    by_id[item.id] = item
                end
                local sword = test.assert(by_id["fixture-short-sword"], [=[by_id["fixture-short-sword"]]=])
                test.assert(
                    sword.width == 1 and sword.height == 3 and sword.container == "inventory",
                    [=[sword.width == 1 and sword.height == 3 and sword.container == "inventory"]=]
                )
                test.assert(
                    sword.body_slots == "rarm,larm" and sword.composite == "RH=SSD",
                    [=[sword.body_slots == "rarm,larm" and sword.composite == "RH=SSD"]=]
                )
                test.assert(
                    sword.weapon_class == "1HS" and sword.melee_range == 2,
                    [=[sword.weapon_class == "1HS" and sword.melee_range == 2]=]
                )
                test.assert(
                    sword.physical_min == 512 and sword.physical_max == 1792,
                    [=[sword.physical_min == 512 and sword.physical_max == 1792]=]
                )
                test.assert(
                    by_id["fixture-vendor-short-sword"].slot == "weap"
                        and by_id["fixture-hireling-cap"].slot == "head"
                        and by_id["fixture-mp1"].container == "belt"
                )
            end),
        }),
    },
})
