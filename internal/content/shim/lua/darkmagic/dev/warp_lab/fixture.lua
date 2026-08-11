-- Construct and own Warp Lab's three authoritative entities.

local ecs = require("dm.ecs/v1")
local components = require("darkmagic.dev.warp_lab.components")
local system = require("darkmagic.dev.warp_lab.system")
local M = {}

function M.create(a,b,spawn)
    components.register()
    system.register()
    local fixture = {entities = {}}
    local function portal_position(point)
        local entity=ecs.create({
            ["dm.world.position"]={x=point.x,y=point.y},
        })
        fixture.entities[#fixture.entities + 1] = entity
        return entity
    end
    fixture.portal_a = portal_position(a)
    fixture.portal_b = portal_position(b)
    ecs.set(fixture.portal_a, "dm.lab.warp.portal", {
        id = "warp-lab:a", pair = fixture.portal_b,
        label = "WESTERN PORTAL", radius = 1.5,
    })
    ecs.set(fixture.portal_b, "dm.lab.warp.portal", {
        id = "warp-lab:b", pair = fixture.portal_a,
        label = "EASTERN PORTAL", radius = 1.5,
    })
    fixture.player = ecs.create({
        ["dm.world.position"] = {x = spawn.x, y = spawn.y},
        ["dm.lab.warp.actor"] = {speed = 12},
        ["dm.lab.warp.state"] = {event = "ready: click a portal", warp_count = 0},
    })
    fixture.entities[#fixture.entities + 1] = fixture.player
    return fixture
end

function M.intent(fixture,portal_id)
    local portal = portal_id == "warp-lab:a" and fixture.portal_a or fixture.portal_b
    ecs.set(fixture.player, "dm.lab.warp.intent", {portal = portal})
    ecs.get(fixture.player, "dm.lab.warp.state"):set("event", "interaction intent: " .. portal_id)
end

function M.destroy(fixture)
    if not fixture then return end
    for _, entity in ipairs(fixture.entities) do
        if entity:exists() then ecs.destroy(entity) end
    end
end

return M
