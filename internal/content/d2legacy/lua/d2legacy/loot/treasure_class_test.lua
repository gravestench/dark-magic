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
    local drops = treasure.roll("root")

    assert(#drops == 1)
    assert(drops[1].code == "rin")
    assert(drops[1].quality.unique == 200)
    assert(drops[1].quality.set == 300)
    assert(require("d2legacy.loot.generate").encode({}) == "[]")
end

return test.suite({
    profile = "policy",
    tier = "fast",
    records = records,
    tests = {
        nested_classes_keep_strongest_quality_modifiers = {
            { run = nested_classes_keep_strongest_quality_modifiers },
        },
    },
})
