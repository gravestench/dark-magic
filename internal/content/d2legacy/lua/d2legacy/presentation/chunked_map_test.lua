local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("reuses_warm_nodes_when_camera_returns", {
            test.run(function()
                local created, destroyed = 0, 0
                local nodes = {}
                test.mock_module("engine.render/v1", {
                    create = function(_, parent)
                        created = created + 1
                        local node = { visible = true, alive = true }
                        function node:set_z() end
                        function node:set_position() end
                        function node:set_visible(value)
                            self.visible = value
                        end
                        function node:exists()
                            return self.alive
                        end
                        function node:destroy()
                            self.alive = false
                            destroyed = destroyed + 1
                        end
                        function node:set_world_tile(_, _, index)
                            nodes[index + 1] = self
                            return index * 1000, 0, 100, 100
                        end
                        return node
                    end,
                    preload = function()
                        return {}
                    end,
                    preload_status = function()
                        return { done = true, failed = 0 }
                    end,
                    preload_release = function() end,
                    world_tiles = function()
                        return {
                            width = 2000,
                            height = 600,
                            bucket_size = 512,
                            draws = {
                                {
                                    index = 0,
                                    x = 0,
                                    y = 0,
                                    width = 100,
                                    height = 100,
                                    depth = 0,
                                },
                                {
                                    index = 1,
                                    x = 1000,
                                    y = 0,
                                    width = 100,
                                    height = 100,
                                    depth = 0,
                                },
                            },
                            buckets = { ["0:0"] = { 1 }, ["1:0"] = { 1 }, ["2:0"] = { 2 } },
                        }
                    end,
                }, {
                    "create",
                    "preload",
                    "preload_status",
                    "preload_release",
                    "world_tiles",
                })
                local chunks = require("d2legacy.presentation.chunked_map")
                local state = chunks.create({}, { world = {}, palette = "" }, {
                    viewport_width = 100,
                    viewport_height = 100,
                    margin = 0,
                    resident_limit = 2,
                    admit_per_frame = 32,
                })
                chunks.update(state, 0, 0, 50, 50)
                chunks.update(state, 0, 0, 50, 50)
                test.assert(nodes[1] and nodes[1].visible, [=[nodes[1] and nodes[1].visible]=])
                chunks.update(state, 1000, 0, 50, 50)
                chunks.update(state, 1000, 0, 50, 50)
                test.assert(
                    nodes[1] and not nodes[1].visible and nodes[2] and nodes[2].visible,
                    [=[nodes[1] and not nodes[1].visible and nodes[2] and nodes[2].visible]=]
                )
                local before = created
                chunks.update(state, 0, 0, 50, 50)
                test.assert(
                    created == before and nodes[1].visible and not nodes[2].visible,
                    [=[created == before and nodes[1].visible and not nodes[2].visible]=]
                )
                test.assert(destroyed == 0, [=[destroyed == 0]=])
            end),
        }),
    },
})
