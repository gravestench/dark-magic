local test = require("d2legacy.tests/v1")

-- records supplies a reciprocal synthetic five-zone recipe without owned game files.
local function records()
    local presets = {}
    for _, definition in ipairs({ 17, 26, 27, 28, 29, 30, 35, 51, 83, 84, 85, 86 }) do
        local size = definition >= 83 and "24" or "8"
        local row = {
            Def = tostring(definition),
            LevelId = "0",
            Outdoors = "1",
            SizeX = size,
            SizeY = size,
            Files = "4",
            Dt1Mask = "1",
        }
        for variant = 1, 4 do
            row["File" .. variant] = "Act1/fixture" .. definition .. "-" .. variant .. ".ds1"
        end
        presets[#presets + 1] = row
    end
    presets[#presets + 1] = {
        Def = "1",
        LevelId = "1",
        SizeX = "40",
        SizeY = "40",
        Files = "1",
        File1 = "Act1/Town/townE1.ds1",
        Dt1Mask = "1",
    }
    local levels = { { Id = "1", Act = "0", DrlgType = "2", LevelType = "1" } }
    for id = 2, 4 do
        levels[#levels + 1] = {
            Id = tostring(id),
            Act = "0",
            DrlgType = "3",
            LevelType = "2",
            SizeX = "80",
            SizeY = "80",
            ["SizeX(N)"] = "80",
            ["SizeY(N)"] = "80",
            ["SizeX(H)"] = "80",
            ["SizeY(H)"] = "80",
            Vis3 = "8",
            Warp3 = "0",
        }
    end
    levels[#levels + 1] = { Id = "8", Act = "0", DrlgType = "1", LevelType = "3", Vis0 = "2", Warp0 = "4" }
    return {
        ["data/global/excel/Levels.txt"] = levels,
        ["data/global/excel/LvlPrest.txt"] = presets,
        ["data/global/excel/LvlTypes.txt"] = {
            {},
            { ["File 1"] = "Act1/Town/town.dt1" },
            { ["File 1"] = "Act1/Outdoors/outdoor.dt1" },
            { ["File 1"] = "Act1/Caves/cave.dt1" },
        },
        ["data/global/excel/LvlMaze.txt"] = {
            { Level = "8", SizeX = "24", SizeY = "24", Rooms = "1", ["Rooms(N)"] = "1", ["Rooms(H)"] = "1" },
        },
    }
end

-- equal detects recipe drift through nested zone and seam tables.
local function equal(first, second)
    if type(first) ~= "table" or type(second) ~= "table" then
        return first == second
    end
    for key, value in pairs(first) do
        if not equal(value, second[key]) then
            return false
        end
    end
    for key in pairs(second) do
        if first[key] == nil then
            return false
        end
    end
    return true
end

-- load_entry confines mock records to each v4 recipe case.
local function load_entry(rows)
    test.mock_module("engine.records/v1", {
        load = function(path)
            return rows[path] or {}
        end,
    }, { "load" })
    for _, name in ipairs({
        "entry_world",
        "records",
        "preset",
        "outdoor",
        "outdoor_structures",
        "cold_plains",
        "stony_field",
        "maze",
    }) do
        test.unload_module("d2legacy.mapgen." .. name)
    end
    return require("d2legacy.mapgen.entry_world")
end

return test.suite({
    profile = "module",
    tier = "fast",
    disable_execution_budget = true,
    cases = {
        test.case("TestEntryWorldV4Recipe", {
            test.run(function()
                local entry = load_entry(records())
                for difficulty = 0, 2 do
                    local town = entry.town(42, difficulty)
                    local moor = entry.wilderness(42, difficulty, town)
                    local plains = entry.cold_plains_v3(42, difficulty, moor)
                    local stony = entry.stony_field(42, difficulty, plains)
                    local original_seams = {
                        entry.seam(town, moor),
                        entry.generated_seam(moor, plains),
                        entry.stony_field_seam(plains, stony),
                    }
                    local world = entry.v4(42, difficulty)
                    test.assert(#world.zones == 5, "exactly five v4 zones")
                    for index, id in ipairs({ 1, 2, 3, 4, 8 }) do
                        test.assert(world.zones[index].request.level_id == id, "canonical nonconsecutive level set")
                    end
                    test.assert(equal(world.seams, original_seams), "all three seamless outdoor joins unchanged")
                    test.assert(
                        world.zones[1].checksum == town.checksum
                            and world.zones[3].checksum == plains.checksum
                            and world.zones[4].checksum == stony.checksum,
                        "earlier non-Blood-Moor recipes unchanged"
                    )
                    local cave = world.cave_warp
                    test.assert(
                        cave.first_level == 2
                            and cave.second_level == 8
                            and cave.first_warp_role == "den-entrance"
                            and cave.second_warp_role == "den-exit",
                        "reciprocal semantic cave link"
                    )
                    test.assert(
                        cave.first_zone_warp_id == 3
                            and cave.second_zone_warp_id == 1
                            and cave.first_vis_slot == 3
                            and cave.second_vis_slot == 0
                            and cave.first_lvlwarp_id == 0
                            and cave.second_lvlwarp_id == 4,
                        "semantic endpoint IDs are distinct from Levels slots and LvlWarp IDs"
                    )
                    test.assert(equal(world, entry.v4(42, difficulty)), "whole v4 recipe replays deterministically")
                    test.assert(
                        entry.wilderness(42, difficulty, town).checksum == moor.checksum,
                        "legacy wilderness remains unchanged"
                    )
                end
            end),
        }),
        test.case("TestEntryWorldV4RejectsAmbiguousCavePolicy", {
            test.run(function()
                local entry = load_entry(records())
                for index, mutate in ipairs({
                    function(moor)
                        moor.warps = {}
                    end,
                    function(moor)
                        moor.warps[2] = moor.warps[1]
                    end,
                    function(moor)
                        moor.warps[1].destination_level = 9
                    end,
                    function(_, den)
                        den.warps[1].destination_level = 3
                    end,
                    function(_, den)
                        den.request.level_id = 9
                    end,
                }) do
                    local first = {
                        request = { level_id = 2 },
                        warps = { { id = 3, role = "den-entrance", direction = "south", destination_level = 8 } },
                    }
                    local second = {
                        request = { level_id = 8 },
                        warps = { { id = 1, role = "den-exit", direction = "north", destination_level = 2 } },
                    }
                    mutate(first, second)
                    test.assert(
                        not pcall(entry.den_warp, first, second),
                        "malformed cave endpoints fail closed: " .. index
                    )
                end
            end),
        }),
    },
})
