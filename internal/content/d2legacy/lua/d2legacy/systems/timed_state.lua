-- Apply, refresh, remove, and expire source-tagged Diablo state instances.
--
-- A state is not merely a timer. Its exact source matters: poison from two
-- monsters, an aura, and a shrine must be removable independently. Requests
-- are short-lived facts; durable instances and emitted events are checkpointed
-- ECS state like every other authoritative system output.

local ecs=require("engine.ecs/v1")
local M={}
local refresh_policy="refresh_same_source"

local function same(instance, request)
    return instance:get("target"):id()==request:get("target"):id()
        and instance:get("state_id")==request:get("state_id")
        and instance:get("source_id")==request:get("source_id")
end

local function emit(structural,kind,tick,request,expires,reason)
    structural:create({["d2legacy.state.event"]={kind=kind,tick=tick,target=request:get("target"),
        state_id=request:get("state_id"),source_id=request:get("source_id"),
        expires_tick=expires,reason=reason}})
end

function M.register()
    ecs.system({id="d2legacy.state.timed_instances",phase="effects",
        query={any={"d2legacy.state.request","d2legacy.state.instance"}},
        read={"d2legacy.state.request","d2legacy.state.instance"},
        write={"d2legacy.state.request","d2legacy.state.instance","d2legacy.state.event"},
        update=function(context,entities,structural)
            -- Stable entity order gives repeated same-tick requests a stable
            -- last-writer order without relying on Lua table iteration.
            local touched={}
            for _,request_entity in ipairs(entities) do
                local request=ecs.get(request_entity,"d2legacy.state.request")
                if request then
                    assert(request:get("state_id")~="" and request:get("source_id")~=""
                        and request:get("policy")==refresh_policy,"invalid timed-state request")
                    local match=nil
                    for _,candidate in ipairs(entities) do
                        local instance=ecs.get(candidate,"d2legacy.state.instance")
                        if instance and same(instance,request) then match=candidate; break end
                    end
                    local operation=request:get("operation")
                    if operation=="apply" then
                        assert(request:get("duration")>0,"state duration must be positive")
                        local expires=context.tick+request:get("duration")
                        if match then
                            ecs.get(match,"d2legacy.state.instance"):set("expires_tick",expires)
                            touched[match:id()]=true
                            emit(structural,"state_refreshed",context.tick,request,expires,"refresh")
                        else
                            structural:create({["d2legacy.state.instance"]={target=request:get("target"),
                                state_id=request:get("state_id"),source_id=request:get("source_id"),
                                applied_tick=context.tick,expires_tick=expires,policy=refresh_policy}})
                            emit(structural,"state_applied",context.tick,request,expires,"apply")
                        end
                    elseif operation=="remove" then
                        if match then
                            local expires=ecs.get(match,"d2legacy.state.instance"):get("expires_tick")
                            touched[match:id()]=true; structural:destroy(match)
                            emit(structural,"state_removed",context.tick,request,expires,"explicit")
                        end
                    else error("unsupported timed-state operation") end
                    structural:destroy(request_entity)
                end
            end
            for _,entity in ipairs(entities) do
                local instance=ecs.get(entity,"d2legacy.state.instance")
                if instance and not touched[entity:id()] and context.tick>=instance:get("expires_tick") then
                    structural:create({["d2legacy.state.event"]={kind="state_removed",tick=context.tick,
                        target=instance:get("target"),state_id=instance:get("state_id"),
                        source_id=instance:get("source_id"),expires_tick=instance:get("expires_tick"),reason="expired"}})
                    structural:destroy(entity)
                end
            end
        end})
end
return M
