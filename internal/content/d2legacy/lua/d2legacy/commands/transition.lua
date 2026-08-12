-- Apply a trusted transition recipe produced by the active world adapter.
--
-- The engine/map adapter discovers collision-space seam geometry. This module
-- owns Diablo's rule that a nearby player changes level, appears at the paired
-- arrival point, stops moving, and adopts the destination world bounds.

local commands = require("engine.authority_command/v1")
local ecs = require("engine.ecs/v1")
local M = {}
local trigger_radius = 2

local function finite(value)
    return type(value)=="number" and value==value and value~=math.huge and value~=-math.huge
end

function M.validate(command)
    local p=command.payload
    assert(type(p)=="table", "transition payload must be a table")
    assert((p.source_level==1 or p.source_level==2) and (p.destination_level==1 or p.destination_level==2)
        and p.source_level~=p.destination_level, "transition levels are invalid")
    for _, name in ipairs({"source_x","source_y","arrival_x","arrival_y","world_width","world_height"}) do
        assert(finite(p[name]), "transition "..name.." must be finite")
    end
    assert(p.world_width>0 and p.world_height>0 and p.arrival_x>=0 and p.arrival_y>=0
        and p.arrival_x<p.world_width and p.arrival_y<p.world_height, "transition destination is invalid")
end

function M.apply(command)
    local p=command.payload
    for _, entity in ipairs(ecs.query({all={"d2legacy.world.player_control","d2legacy.world.location",
            "d2legacy.world.position","d2legacy.world.bounds","d2legacy.world.velocity"}})) do
        local control=ecs.get(entity,"d2legacy.world.player_control")
        if control:get("player")==command.player then
            local location=ecs.get(entity,"d2legacy.world.location")
            local position=ecs.get(entity,"d2legacy.world.position")
            assert(location:get("level_id")==p.source_level, "player is not in transition source level")
            local dx,dy=position:get("x")-p.source_x,position:get("y")-p.source_y
            assert(math.sqrt(dx*dx+dy*dy)<=trigger_radius, "player is outside the authored seam")
            location:set("level_id",p.destination_level)
            position:set("x",p.arrival_x); position:set("y",p.arrival_y)
            local bounds=ecs.get(entity,"d2legacy.world.bounds")
            bounds:set("width",p.world_width); bounds:set("height",p.world_height)
            local velocity=ecs.get(entity,"d2legacy.world.velocity")
            velocity:set("x",0); velocity:set("y",0)
            return
        end
    end
    error("transition player is unavailable")
end

function M.register()
    commands.register({kind="system.world.transition",authorities={"system","administrator"},
        validate=M.validate,apply=M.apply})
end
return M
