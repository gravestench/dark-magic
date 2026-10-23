local test = require("d2legacy.tests/v1")

local function ground_art_lifecycle()
    local visible, nodes, releases, animations = true, {}, 0, 0
    local function component(values)
        return {
            get = function(_, name)
                return values[name]
            end,
        }
    end
    local entity = {
        id = function()
            return 1
        end,
    }
    local values = {
        ["d2legacy.world.location"] = component({ level_id = 8 }),
        ["d2legacy.item.presentation"] = component({ world_dc6 = "sword.dc6", world_animated = true }),
        ["d2legacy.world.position"] = component({ x = 12, y = 14 }),
    }
    test.mock_module("engine.ecs/v1", {
        query = function()
            return visible and { entity } or {}
        end,
        get = function(_, name)
            return values[name]
        end,
    }, { "query", "get" })
    test.mock_module("engine.render/v1", {
        create = function()
            local node = {}
            function node:set_visible(value)
                self.visible = value
            end
            function node:set_dc6_animation(_, _, _, _, loop)
                test.expect(loop):equals("once")
                animations = animations + 1
                return 8, 10, 20, -4, -18
            end
            function node:set_position(x, y)
                self.x, self.y = x, y
            end
            function node:set_z(value)
                self.z = value
            end
            function node:destroy()
                self.destroyed = true
            end
            nodes[#nodes + 1] = node
            return node
        end,
        preload = function()
            return 1
        end,
        preload_status = function()
            return { done = true, failed = 0 }
        end,
        preload_release = function()
            releases = releases + 1
        end,
    }, { "create", "preload", "preload_status", "preload_release" })
    local world = {
        canvas = function()
            return 800, 600
        end,
        subtile_to_pixel = function(_, x, y)
            return x * 10, y * 10
        end,
        entity_depth = function(_, x, y)
            return x + y
        end,
    }
    test.unload_module("d2legacy.gameplay.ground_items")
    local view = require("d2legacy.gameplay.ground_items").new({}, world, 8, "units.pal")
    view:update()
    view:update()
    test.expect(#nodes):equals(1)
    test.expect(animations):equals(1)
    test.expect(releases):equals(1)
    test.expect(nodes[1].x):equals(120 - 400 + 1)
    test.expect(nodes[1].y):equals(140 - 300 - 8)
    visible = false
    view:update()
    test.assert(nodes[1].destroyed, "pickup removes retained ground art")
    visible = true
    view:update()
    view:destroy()
    test.assert(nodes[2].destroyed, "world teardown removes retained ground art")
end

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("ground_animation_anchor_pickup_and_world_teardown", { test.run(ground_art_lifecycle) }),
    },
})
