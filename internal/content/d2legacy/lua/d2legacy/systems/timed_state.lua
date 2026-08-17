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

local function stat_source(entities,target,source_id)
    for _,entity in ipairs(entities) do
        local source=ecs.get(entity,"d2legacy.stat.source")
        if source and source:get("target"):id()==target:id() and source:get("source_id")==source_id then
            return entity
        end
    end
    return nil
end

local function reconcile_stat(structural,entities,request)
    local existing=stat_source(entities,request:get("target"),request:get("source_id"))
    if request:get("stat")=="" then
        if existing then structural:destroy(existing) end
        return
    end
    assert(request:get("stat_operation")=="add" or request:get("stat_operation")=="percent",
        "invalid timed-state stat operation")
    local values={target=request:get("target"),source_id=request:get("source_id"),stat=request:get("stat"),
        operation=request:get("stat_operation"),value=request:get("stat_value"),order=request:get("stat_order")}
    if existing then
        local source=ecs.get(existing,"d2legacy.stat.source")
        source:set("target",values.target)
        source:set("source_id",values.source_id)
        source:set("stat",values.stat)
        source:set("operation",values.operation)
        source:set("value",values.value)
        source:set("order",values.order)
    else
        structural:create({["d2legacy.stat.source"]=values})
    end
end

local function remove_stat(structural,entities,target,source_id)
    local existing=stat_source(entities,target,source_id)
    if existing then structural:destroy(existing) end
end

function M.register()
    ecs.system({id="d2legacy.state.timed_instances",phase="effects",
        query={any={"d2legacy.state.request","d2legacy.state.instance","d2legacy.stat.source"}},
        read={"d2legacy.state.request","d2legacy.state.instance","d2legacy.stat.source"},
        write={"d2legacy.state.request","d2legacy.state.instance","d2legacy.state.event","d2legacy.stat.source"},
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
                            local instance=ecs.get(match,"d2legacy.state.instance")
                            instance:set("expires_tick",expires)
                            instance:set("stat",request:get("stat"))
                            instance:set("stat_operation",request:get("stat_operation"))
                            instance:set("stat_value",request:get("stat_value"))
                            instance:set("stat_order",request:get("stat_order"))
                            touched[match:id()]=true
                            emit(structural,"state_refreshed",context.tick,request,expires,"refresh")
                        else
                            structural:create({["d2legacy.state.instance"]={target=request:get("target"),
                                state_id=request:get("state_id"),source_id=request:get("source_id"),
                                applied_tick=context.tick,expires_tick=expires,policy=refresh_policy,
                                stat=request:get("stat"),stat_operation=request:get("stat_operation"),
                                stat_value=request:get("stat_value"),stat_order=request:get("stat_order")}})
                            emit(structural,"state_applied",context.tick,request,expires,"apply")
                        end
                        reconcile_stat(structural,entities,request)
                    elseif operation=="remove" then
                        if match then
                            local expires=ecs.get(match,"d2legacy.state.instance"):get("expires_tick")
                            touched[match:id()]=true; structural:destroy(match)
                            remove_stat(structural,entities,request:get("target"),request:get("source_id"))
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
                    remove_stat(structural,entities,instance:get("target"),instance:get("source_id"))
                    structural:destroy(entity)
                end
            end
        end})
end
return M
