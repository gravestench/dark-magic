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
    },
}

local function nested_classes_keep_strongest_quality_modifiers()
    local treasure = require("d2legacy.loot.treasure_class")
    local drops = treasure.roll("root", { effective_player_count = 1, nearby_party_member_count = 0 })

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
    local function adjusted(effective, nearby)
        return treasure.adjusted_no_drop(100, 100, {
            effective_player_count = effective,
            nearby_party_member_count = nearby,
        })
    end

    test.assert(adjusted(1, 0) == 100, [=[adjusted(1, 0) == 100]=])
    test.assert(adjusted(2, 0) == 100, [=[adjusted(2, 0) == 100]=])
    test.assert(adjusted(3, 0) == 33, [=[adjusted(3, 0) == 33]=])
    test.assert(adjusted(2, 1) == 33, [=[adjusted(2, 1) == 33]=])
    test.assert(adjusted(8, 0) == 6, [=[adjusted(8, 0) == 6]=])
end

return test.suite({
    profile = "module",
    tier = "fast",
    records = records,
    cases = {
        test.case("nested_classes_keep_strongest_quality_modifiers", {
            { run = nested_classes_keep_strongest_quality_modifiers },
        }),
        test.case("nodrop_uses_distinct_effective_and_nearby_counts", {
            { run = nodrop_uses_distinct_effective_and_nearby_counts },
        }),
    },
})
