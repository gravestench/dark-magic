local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("uses_joined_monstats2_pieces", {
            test.run(function()
                test.mock_module("engine.render/v1", {
                    cof_info = function(path)
                        test.assert(
                            path == "data/global/monsters/FA/COF/FAWLHTH.cof",
                            [=[path == "data/global/monsters/FA/COF/FAWLHTH.cof"]=]
                        )
                        return {
                            directions = 8,
                            frames = 8,
                            events = {},
                            layers = {
                                { type = "HD", weapon_class = "hth" },
                                { type = "TR", weapon_class = "hth" },
                                { type = "SH", weapon_class = "hth" },
                            },
                        }
                    end,
                    asset_exists = function(path)
                        return path ~= "data/global/monsters/FA/SH/FASHLITWLHTH.dcc"
                    end,
                    animdata_info = function()
                        return nil
                    end,
                }, { "cof_info", "asset_exists", "animdata_info" })
                local adapter = require("d2legacy.gameplay.monster_composite")
                local direction = require("d2legacy.policy.direction")
                test.assert(direction.quantize(1, 0, 8) == 3, [=[direction.quantize(1, 0, 8) == 3]=])
                test.assert(direction.quantize(1, 1, 8) == 4, [=[direction.quantize(1, 1, 8) == 4]=])
                local composite = adapter.resolve({
                    token = "FA",
                    mode = "WL",
                    weapon_class = "HTH",
                    components = "HD=LIT,TR=MED",
                    direction = 3,
                })
                test.assert(composite.direction == 7, [=[composite.direction == 7]=])
                test.assert(
                    composite.components.HD == "data/global/monsters/FA/HD/FAHDLITWLHTH.dcc",
                    [=[composite.components.HD == "data/global/monsters/FA/HD/FAHDLITWLHTH.dcc"]=]
                )
                test.assert(
                    composite.components.TR == "data/global/monsters/FA/TR/FATRMEDWLHTH.dcc",
                    [=[composite.components.TR == "data/global/monsters/FA/TR/FATRMEDWLHTH.dcc"]=]
                )
                test.assert(composite.components.SH == nil, [=[composite.components.SH == nil]=])
                test.assert(
                    composite.rate == 128 and composite.frames == 8,
                    [=[composite.rate == 128 and composite.frames == 8]=]
                )
            end),
        }),
    },
})
