local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("shared_tiles_cover_seam_and_interleave_with_entity", {
            test.run(function()
                local alive, jobs, released = {}, {}, {}
                local draws = {
                    { index = 0, x = 0, y = 0, width = 100, height = 100, depth = 10 },
                    { index = 1, x = 1000, y = 0, width = 100, height = 100, depth = -2000000 },
                    { index = 2, x = 2000, y = 0, width = 100, height = 100, depth = 2000000 },
                }
                local function node(parent)
                    local n = { parent = parent, alive = true }
                    function n:set_position(x, y)
                        self.x, self.y = x, y
                    end
                    function n:set_z(z)
                        self.z = z
                    end
                    function n:set_visible(v)
                        self.visible = v
                    end
                    function n:exists()
                        return self.alive
                    end
                    function n:destroy()
                        self.alive = false
                    end
                    function n:set_world_tile(world, _, index)
                        self.world, self.index = world, index
                        local d = draws[index + 1]
                        return d.x, d.y, d.width, d.height
                    end
                    alive[#alive + 1] = n
                    return n
                end
                test.mock_module("engine.render/v1", {
                    create = function(_, parent)
                        return node(parent)
                    end,
                    preload = function(request)
                        local job = { request = request[1], done = true }
                        jobs[#jobs + 1] = job
                        return job
                    end,
                    preload_status = function(job)
                        return { done = job.done, failed = 0 }
                    end,
                    preload_release = function(job)
                        if not job.done then
                            return false
                        end
                        released[job] = true
                        return true
                    end,
                    world_tiles = function()
                        return {
                            width = 3000,
                            height = 100,
                            bucket_size = 100,
                            draws = draws,
                            buckets = {
                                ["0:0"] = { 1 },
                                ["10:0"] = { 2 },
                                ["20:0"] = { 3 },
                            },
                        }
                    end,
                }, { "create", "preload", "preload_status", "preload_release", "world_tiles" })
                local chunks = require("d2legacy.presentation.chunked_map")
                local root = node(nil)
                root:set_position(-2900, 0)
                root:set_z(99)
                local entity = node(root)
                entity:set_z(63)
                local source_world, destination_world = {}, {}
                local function adapter(world, offset, depth)
                    return chunks.create({}, { world = world, palette = "" }, {
                        shared_root = root,
                        offset_x = offset,
                        offset_y = 0,
                        depth_offset = depth,
                        depth_scale = 4,
                        depth_priority = world == source_world and 3 or 2,
                        viewport_width = 100,
                        viewport_height = 100,
                        resident_limit = 1,
                        admit_per_frame = 1,
                    })
                end
                local source = adapter(source_world, 2900, 0)
                local destination = adapter(destination_world, 3000, 10)
                -- Camera is still inside source canvas; destination starts at
                -- the right viewport edge and must preload before crossing.
                chunks.update(source, 2950, 50, 50, 50)
                chunks.update(destination, 2950, 50, 50, 50)
                test.assert(destination.pending[1] ~= nil, "neighbor tile queued before crossing")
                chunks.update(source, 2950, 50, 50, 50)
                local ok, err, ax, ay = chunks.update(destination, 2950, 50, 50, 50)
                test.assert(ok and not err and ax == 50 and ay == 50, "shared camera anchor is not clamped")
                local a, b = source.nodes[1], destination.nodes[1]
                test.assert(
                    a and b and a.parent == root and b.parent == root and entity.parent == root,
                    "tiles and entity are directly sortable siblings"
                )
                test.assert(a.x == 2950 and b.x == 3050 and b.y == 50, "actual tile rectangles use global offsets")
                test.assert(
                    a.x - 50 - 2900 == 0 and b.x - 50 - 2900 == 100,
                    "actual source and destination tile edges meet at viewport seam"
                )
                test.assert(a.z == 43 and entity.z == 63 and b.z == 82, "cross-map wall depths interleave with entity")
                test.assert(root.x == -2900 and root.y == 0 and root.z == 99, "adapter preserves scene root transform")
                root:set_position(-2950, 0)
                chunks.update(source, 3000, 50, 50, 50)
                chunks.update(destination, 3000, 50, 50, 50)
                test.assert(
                    a.x + 50 + root.x == 50
                        and b.x - 50 + root.x == 50
                        and a.x - 50 + root.x <= 0
                        and b.x + 50 + root.x >= 100,
                    "actual neighboring tile rectangles cover both halves of seam viewport"
                )
                -- Leave an unfinished tile job behind, then pan farther away.
                chunks.update(destination, 3950, 50, 50, 50)
                local pending = destination.pending[2]
                test.assert(pending ~= nil, "new visible tile queues")
                pending.done = false
                chunks.update(destination, 4950, 50, 50, 50)
                test.assert(
                    not released[pending] and destination.pending[2] == pending and not destination.pending[3],
                    "unfinished offscreen job remains owned and prevents unbounded new work"
                )
                pending.done = true
                chunks.update(destination, 4950, 50, 50, 50)
                test.assert(released[pending], "completed offscreen job releases")
                chunks.update(destination, 4950, 50, 50, 50)
                chunks.update(destination, 4950, 50, 50, 50)
                test.assert(
                    not b.alive and destination.nodes[3].z == 8000002,
                    "residency bounded and roof band preserved"
                )
                chunks.update(destination, 3950, 50, 50, 50)
                chunks.update(destination, 3950, 50, 50, 50)
                test.assert(destination.nodes[2].z == -7999998, "floor band preserved")
                chunks.update(destination, 4950, 50, 50, 50)
                chunks.destroy(destination)
                test.assert(root.alive and a.alive and entity.alive, "destroy leaves sibling adapter and root alive")
                for _, job in ipairs(jobs) do
                    test.assert(released[job], "all owned jobs released")
                end
                chunks.destroy(source)
                test.assert(root.alive and not a.alive, "last adapter still does not own shared root")
            end),
        }),
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
