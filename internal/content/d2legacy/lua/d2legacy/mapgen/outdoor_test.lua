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
    test.assert(
        zone.kind == "outdoor" and #zone.rooms == 100 and #zone.links == 180 and #zone.stamps >= 120,
        [=[zone.kind == "outdoor" and #zone.rooms == 100 and #zone.links == 180 and #zone.stamps >= 120]=]
    )
    test.assert(
        zone.checksum == outdoor.generate(2, 42, direction, 0).checksum,
        [=[zone.checksum == outdoor.generate(2, 42, direction, 0).checksum]=]
    )
    local path = {}
    for _, tile in ipairs(zone.paths) do
        path[tile.x .. ":" .. tile.y] = true
    end
    test.assert(
        path[zone.warps[1].x .. ":" .. zone.warps[1].y] and path[zone.warps[2].x .. ":" .. zone.warps[2].y],
        [=[path[zone.warps[1].x .. ":" .. zone.warps[1].y] and path[zone.warps[2].x .. ":" .. zone.warps[2].y]]=]
    )
    local bridge, passable = 0, 0
    for _, tile in ipairs(zone.structures) do
        if path[tile.x .. ":" .. tile.y] then
            test.assert(tile.passable, [=[tile.passable]=])
        end
        if tile.kind == "bridge" then
            bridge = bridge + 1
            if tile.passable then
                passable = passable + 1
            end
        end
    end
    test.assert(bridge == 64 and passable == 48, [=[bridge == 64 and passable == 48]=])
end

return test.suite({
    profile = "module",
    tier = "fast",
    covers = { "internal/mod/d2legacy/mapgen" },
    disable_execution_budget = true,
    cases = {
        test.case("north_recipe_and_structures", {
            test.run(function()
                verify("north")
            end),
        }),
        test.case("east_recipe_and_structures", {
            test.run(function()
                verify("east")
            end),
        }),
        test.case("south_recipe_and_structures", {
            test.run(function()
                verify("south")
            end),
        }),
        test.case("west_recipe_and_structures", {
            test.run(function()
                verify("west")
            end),
        }),
    },
})
