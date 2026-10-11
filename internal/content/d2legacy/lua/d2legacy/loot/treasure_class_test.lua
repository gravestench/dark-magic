local test = require("d2legacy.tests/v1")

local records = {
    ["data/global/excel/treasureclassex.txt"] = {
        {
            ["Treasure Class"] = "root",
            Picks = "-1",
            Item1 = "child",
            Prob1 = "1",
            Unique = "200",
        },
        {
            ["Treasure Class"] = "child",
            Picks = "-1",
            Item1 = "rin",
            Prob1 = "1",
            Unique = "100",
            Set = "300",
        },
        { ["Treasure Class"] = "equipment", Picks = "-1", Item1 = "weap3", Prob1 = "1", Magic = "400" },
        { ["Treasure Class"] = "empty-family", Picks = "-2", Item1 = "weap96", Prob1 = "2" },
        { ["Treasure Class"] = "empty-family-nodrop", Picks = "2", NoDrop = "5", Item1 = "weap96", Prob1 = "1" },
    },
    ["data/global/excel/ItemTypes.txt"] = {
        { Code = "weap", TreasureClass = "1" },
        { Code = "mele", Equiv1 = "weap" },
        { Code = "swor", Equiv1 = "absent", Equiv2 = "mele", Rarity = "2" },
        { Code = "axe", Equiv1 = "mele", Rarity = "0" },
        { Code = "armo", TreasureClass = "1" },
        { Code = "helm", Equiv1 = "armo", Rarity = "1" },
        { Code = "mpot", Equiv1 = "weap", TreasureClass = "1" },
    },
    ["data/global/excel/weapons.txt"] = {
        { code = "ssd", type = "swor", level = "3", spawnable = "1" },
        { code = "hax", type = "axe", level = "4", spawnable = "1" },
        { code = "axe", type = "axe", level = "6", spawnable = "1" },
        { code = "high", type = "axe", level = "7", spawnable = "1" },
        { code = "offtype", type = "absent", level = "1", spawnable = "1" },
        { code = "disabled", type = "swor", level = "1", spawnable = "0" },
        { code = "quest", type = "swor", level = "1", spawnable = "1", quest = "1" },
        { code = "zero", type = "swor", level = "0", spawnable = "1" },
    },
    ["data/global/excel/armor.txt"] = {
        { code = "cap", type = "helm", level = "1", spawnable = "1" },
    },
    ["data/global/excel/misc.txt"] = {
        { code = "gpl", type = "mpot", level = "1", spawnable = "1" },
    },
}

local function implicit_item_families_resolve_to_eligible_base_codes()
    local treasure = require("d2legacy.loot.treasure_class")
    local context = { no_drop_player_count = 1 }
    local drops = treasure.roll("equipment", context)
    test.expect(#drops):equals(1)
    test.expect(drops[1].code):equals("ssd")
    test.expect(table.concat(drops[1].path, " > ")):equals("equipment > weap3")
    test.expect(drops[1].quality.magic):equals(400)
    local code = treasure.roll("weap6", context)[1].code
    test.expect(code == "hax" or code == "axe"):equals(true)
    test.expect(treasure.roll("armo3", context)[1].code):equals("cap")
    test.expect(treasure.roll("mpot3", context)[1].code):equals("gpl")
    test.expect(#treasure.roll("weap96", context)):equals(0)
end

local function empty_item_families_drop_nothing()
    local treasure = require("d2legacy.loot.treasure_class")
    test.expect(#treasure.roll("empty-family", { no_drop_player_count = 1 })):equals(0)
    test.expect(#treasure.roll("empty-family-nodrop", { no_drop_player_count = 3 })):equals(0)
end

local function generated_classes_keep_weights_order_and_catalog()
    local loads, roll = {}, 0
    test.mock_module("engine.records/v1", {
        load = function(path)
            loads[path] = (loads[path] or 0) + 1
            return records[path]
        end,
    })
    test.mock_module("engine.authority_random/v1", {
        integer = function(_, total)
            test.expect(total):equals(3)
            return roll
        end,
    })
    local rows = records["data/global/excel/weapons.txt"]
    rows[#rows + 1] = { code = "secondary", type = "helm", type2 = "swor", level = "3", spawnable = "1" }
    local treasure = require("d2legacy.loot.treasure_class")
    for index, code in ipairs({ "ssd", "ssd", "secondary" }) do
        roll = index - 1
        test.expect(treasure.roll("weap3", {})[1].code):equals(code)
    end
    for _, path in ipairs({ "treasureclassex.txt", "ItemTypes.txt", "weapons.txt", "armor.txt", "misc.txt" }) do
        test.expect(loads["data/global/excel/" .. path]):equals(1)
    end
end

local function explicit_classes_override_generated_names()
    test.mock_module("engine.records/v1", {
        load = function(path)
            test.expect(path):equals("data/global/excel/treasureclassex.txt")
            return {
                { ["Treasure Class"] = "weap3", Picks = "-1", Item1 = "root", Prob1 = "1" },
                records[path][1],
                records[path][2],
            }
        end,
    })
    local drop = require("d2legacy.loot.treasure_class").roll("weap3", {})[1]
    test.expect(drop.code):equals("rin")
    test.expect(table.concat(drop.path, " > ")):equals("weap3 > root > child")
    test.expect(drop.quality.unique):equals(200)
end

local function invalid_families_and_inheritance_fail()
    local types = records["data/global/excel/ItemTypes.txt"]
    types[#types + 1] = { Code = "cycle", Equiv1 = "cycle" }
    records["data/global/excel/weapons.txt"][1].type = "cycle"
    test.mock_module("engine.records/v1", {
        load = function(path)
            return records[path]
        end,
    })
    local treasure = require("d2legacy.loot.treasure_class")
    for _, name in ipairs({ "weap0", "weap4", "weap99", "missing3", "swor3" }) do
        local ok, message = pcall(treasure.roll, name, {})
        test.expect(ok):equals(false)
        test.expect(string.find(message, "unknown treasure class", 1, true) ~= nil):equals(true)
    end
    local ok, message = pcall(treasure.roll, "weap3", {})
    test.expect(ok):equals(false)
    test.expect(string.find(message, "cyclic ItemTypes inheritance at cycle", 1, true) ~= nil):equals(true)
end

local function missing_source_tables_fail_without_caching_partial_rows()
    local missing = { path = "ItemTypes.txt" }
    test.mock_module("engine.records/v1", {
        load = function(path)
            if path == "data/global/excel/" .. missing.path then
                return nil, "missing source table"
            end
            return records[path]
        end,
    })
    local treasure = require("d2legacy.loot.treasure_class")
    for _, path in ipairs({ "ItemTypes.txt", "armor.txt", "misc.txt" }) do
        missing.path = path
        local ok, message = pcall(treasure.roll, "weap3", {})
        test.expect(ok):equals(false)
        test.expect(string.find(message, "missing source table", 1, true) ~= nil):equals(true)
    end
    missing.path = ""
    test.expect(treasure.roll("weap3", {})[1].code):equals("ssd")
end

local function nested_classes_keep_strongest_quality_modifiers()
    local treasure = require("d2legacy.loot.treasure_class")
    local drops = treasure.roll("root", { no_drop_player_count = 1 })

    test.assert(#drops == 1, [=[#drops == 1]=])
    test.assert(drops[1].code == "rin", [=[drops[1].code == "rin"]=])
    test.assert(drops[1].quality.unique == 200, [=[drops[1].quality.unique == 200]=])
    test.assert(drops[1].quality.set == 300, [=[drops[1].quality.set == 300]=])
    test.assert(
        require("d2legacy.loot.generate").encode({}) == "[]",
        [=[require("d2legacy.loot.generate").encode({}) == "[]"]=]
    )
end

local function nodrop_uses_distinct_effective_and_nearby_counts()
    local treasure = require("d2legacy.loot.treasure_class")
    local function adjusted(count)
        return treasure.adjusted_no_drop(100, 100, {
            no_drop_player_count = count,
        })
    end

    test.assert(adjusted(1) == 100, [=[adjusted(1) == 100]=])
    test.assert(adjusted(2) == 33, [=[adjusted(2) == 33]=])
    test.assert(adjusted(3) == 14, [=[adjusted(3) == 14]=])
    test.assert(adjusted(4) == 6, [=[adjusted(4) == 6]=])
end

return test.suite({
    profile = "module",
    tier = "fast",
    records = records,
    cases = {
        test.case(
            "missing_source_tables_fail_without_caching_partial_rows",
            { { run = missing_source_tables_fail_without_caching_partial_rows } }
        ),
        test.case("invalid_families_and_inheritance_fail", { { run = invalid_families_and_inheritance_fail } }),
        test.case("explicit_classes_override_generated_names", { { run = explicit_classes_override_generated_names } }),
        test.case(
            "generated_classes_keep_weights_order_and_catalog",
            { { run = generated_classes_keep_weights_order_and_catalog } }
        ),
        test.case("implicit_item_families_resolve_to_eligible_base_codes", {
            { run = implicit_item_families_resolve_to_eligible_base_codes },
        }),
        test.case("empty_item_families_drop_nothing", { { run = empty_item_families_drop_nothing } }),
        test.case("nested_classes_keep_strongest_quality_modifiers", {
            { run = nested_classes_keep_strongest_quality_modifiers },
        }),
        test.case("nodrop_uses_distinct_effective_and_nearby_counts", {
            { run = nodrop_uses_distinct_effective_and_nearby_counts },
        }),
    },
})
