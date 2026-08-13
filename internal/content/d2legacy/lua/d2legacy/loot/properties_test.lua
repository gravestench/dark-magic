local test = require("d2legacy.tests/v1")

local records = {
    ["data/global/excel/itemratio.txt"] = {
        {
            Version = "100",
            Uber = "0",
            ["Class Specific"] = "0",
            Unique = "1000000000",
            Set = "1000000000",
            Rare = "1000000000",
            Magic = "1000000000",
            HiQuality = "1000000000",
            Normal = "1000000000",
        },
    },
    ["data/global/excel/magicprefix.txt"] = {
        {
            Name = "Strong",
            spawnable = "1",
            frequency = "1",
            level = "1",
            version = "100",
            itype1 = "swor",
            mod1code = "dmg",
            mod1min = "7",
            mod1max = "7",
        },
    },
    ["data/global/excel/properties.txt"] = {
        { code = "dmg", func1 = "1", stat1 = "item_maxdamage" },
    },
}

local function roll_prefixes(affixes, sword, quality)
    for _ = 1, 16 do
        local prefixes = affixes.roll(sword, quality, 1, 100)
        if #prefixes == 1 then
            return prefixes
        end
    end
    error("Strong prefix was not selected in 16 deterministic rolls")
end

local function quality_affix_and_property_vectors()
    local quality = require("d2legacy.loot.quality")
    local affixes = require("d2legacy.loot.affixes")
    local properties = require("d2legacy.loot.properties")
    local sword = {
        type = "swor",
        type2 = "",
        level = 1,
        uber = false,
        class_specific = false,
    }
    local rolled = quality.roll(sword, {
        version = 100,
        monster_level = 1,
        magic_find = 0,
    }, {
        unique = 0,
        set = 0,
        rare = 0,
        magic = 1024,
    })
    test.assert(rolled == "magic", [=[rolled == "magic"]=])

    local prefixes = roll_prefixes(affixes, sword, rolled)
    test.assert(prefixes[1].name == "Strong", [=[prefixes[1].name == "Strong"]=])
    test.assert(prefixes[1].modifiers[1].value == 7, [=[prefixes[1].modifiers[1].value == 7]=])

    local stats, effects, unsupported = properties.apply(prefixes, {})
    test.assert(#stats == 1, [=[#stats == 1]=])
    test.assert(stats[1].code == "item_maxdamage", [=[stats[1].code == "item_maxdamage"]=])
    test.assert(stats[1].value == 7 and stats[1].fn == 1, [=[stats[1].value == 7 and stats[1].fn == 1]=])
    test.assert(#effects == 0 and #unsupported == 0, [=[#effects == 0 and #unsupported == 0]=])
end

return test.suite({
    profile = "module",
    tier = "fast",
    covers = { "internal/game/loot" },
    records = records,
    cases = {
        test.case("quality_affix_and_property_vectors", {
            { run = quality_affix_and_property_vectors },
        }),
    },
})
