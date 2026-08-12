local test = require("d2legacy.tests/v1")

local function verify(direction)
    local prest = {}
    for _, definition in ipairs({ 17, 26, 27, 28, 29, 30, 35 }) do
        local row = {
            Def = tostring(definition),
            SizeX = "8",
            SizeY = "8",
            Files = "1",
            File1 = "Act1/Outdoors/fill" .. definition .. ".ds1",
            Dt1Mask = "1",
            Populate = "1",
        }
        if definition >= 26 and definition <= 28 then
            row.Files = "4"
            for variant = 2, 4 do
                row["File" .. variant] = "Act1/Outdoors/structure" .. definition .. "-" .. variant .. ".ds1"
            end
        end
        prest[#prest + 1] = row
    end
    local rows = {
        ["data/global/excel/Levels.txt"] = {
            { Id = "2", Act = "0", DrlgType = "3", LevelType = "2", SizeX = "80", SizeY = "80" },
        },
        ["data/global/excel/LvlTypes.txt"] = {
            {},
            {},
            { ["File 1"] = "Act1/Outdoors/Outdoor1.dt1" },
        },
        ["data/global/excel/LvlPrest.txt"] = prest,
    }
    package.loaded["engine.records/v1"] = {
        load = function(path)
            return rows[path] or {}
        end,
    }
    package.loaded["d2legacy.mapgen.outdoor"] = nil
    local outdoor = require("d2legacy.mapgen.outdoor")
    local zone = outdoor.generate(2, 42, direction, 0)
    assert(zone.kind == "outdoor" and #zone.rooms == 100 and #zone.links == 180 and #zone.stamps >= 120)
    assert(zone.checksum == outdoor.generate(2, 42, direction, 0).checksum)
    local path = {}
    for _, tile in ipairs(zone.paths) do
        path[tile.x .. ":" .. tile.y] = true
    end
    assert(path[zone.warps[1].x .. ":" .. zone.warps[1].y] and path[zone.warps[2].x .. ":" .. zone.warps[2].y])
    local bridge, passable = 0, 0
    for _, tile in ipairs(zone.structures) do
        if path[tile.x .. ":" .. tile.y] then
            assert(tile.passable)
        end
        if tile.kind == "bridge" then
            bridge = bridge + 1
            if tile.passable then
                passable = passable + 1
            end
        end
    end
    assert(bridge == 64 and passable == 48)
end

return test.suite({
    profile = "policy",
    tier = "fast",
    disable_execution_budget = true,
    tests = {
        north_recipe_and_structures = {
            {
                run = function()
                    verify("north")
                end,
            },
        },
        east_recipe_and_structures = {
            {
                run = function()
                    verify("east")
                end,
            },
        },
        south_recipe_and_structures = {
            {
                run = function()
                    verify("south")
                end,
            },
        },
        west_recipe_and_structures = {
            {
                run = function()
                    verify("west")
                end,
            },
        },
    },
})
