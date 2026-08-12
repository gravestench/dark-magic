local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "presentation",
    tier = "fast",
    tests = {
        uses_joined_monstats2_pieces = {
            {
                run = function()
                    package.loaded["engine.render/v1"] = {
                        cof_info = function(path)
                            assert(path == "data/global/monsters/FA/COF/FAWLHTH.cof")
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
                    }
                    local adapter = require("d2legacy.gameplay.monster_composite")
                    local direction = require("d2legacy.policy.direction")
                    assert(direction.quantize(1, 0, 8) == 3)
                    assert(direction.quantize(1, 1, 8) == 4)
                    local composite = adapter.resolve({
                        token = "FA",
                        mode = "WL",
                        weapon_class = "HTH",
                        components = "HD=LIT,TR=MED",
                        direction = 3,
                    })
                    assert(composite.direction == 7)
                    assert(composite.components.HD == "data/global/monsters/FA/HD/FAHDLITWLHTH.dcc")
                    assert(composite.components.TR == "data/global/monsters/FA/TR/FATRMEDWLHTH.dcc")
                    assert(composite.components.SH == nil)
                    assert(composite.rate == 128 and composite.frames == 8)
                end,
            },
        },
    },
})
