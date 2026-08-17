-- Apply, refresh, replace, remove, and expire source-tagged state instances.

local ecs = require("engine.ecs/v1")
local M = {}
local refresh_policy = "refresh_same_source"

local function same(instance, request)
    return instance:get("target"):id() == request:get("target"):id()
        and instance:get("state_id") == request:get("state_id")
        and instance:get("source_id") == request:get("source_id")
end

local function group_conflict(instance, request)
    local group = request:get("exclusive_group")
    return group ~= ""
        and instance:get("target"):id() == request:get("target"):id()
        and instance:get("exclusive_group") == group
        and not same(instance, request)
end

local function request_key(request)
    local prefix = tostring(request:get("target"):id()) .. ":"
    if request:get("exclusive_group") ~= "" then
        return prefix .. "group:" .. request:get("exclusive_group")
    end
    return prefix .. "source:" .. request:get("state_id") .. ":" .. request:get("source_id")
end

local function emit_request(structural, kind, tick, request, expires, reason)
    structural:create({
        ["d2legacy.state.event"] = {
            kind = kind,
            tick = tick,
            target = request:get("target"),
            state_id = request:get("state_id"),
            source_id = request:get("source_id"),
            expires_tick = expires,
            reason = reason,
        },
    })
end

local function emit_instance(structural, kind, tick, instance, reason)
    structural:create({
        ["d2legacy.state.event"] = {
            kind = kind,
            tick = tick,
            target = instance:get("target"),
            state_id = instance:get("state_id"),
            source_id = instance:get("source_id"),
            expires_tick = instance:get("expires_tick"),
            reason = reason,
        },
    })
end

local function source_owner(source)
    local owner = source:get("owner_source_id")
    return owner ~= "" and owner or source:get("source_id")
end

local function stat_source(entities, target, owner_source_id, stat)
    for _, entity in ipairs(entities) do
        local source = ecs.get(entity, "d2legacy.stat.source")
        if
            source
            and source:get("target"):id() == target:id()
            and source_owner(source) == owner_source_id
            and source:get("stat") == stat
        then
            return entity
        end
    end
    return nil
end

local function reconcile_stat(structural, entities, values)
    local existing = stat_source(entities, values.target, values.owner_source_id, values.stat)
    assert(values.operation == "add" or values.operation == "percent", "invalid timed-state stat operation")
    if existing then
        local source = ecs.get(existing, "d2legacy.stat.source")
        source:set("target", values.target)
        source:set("source_id", values.source_id)
        source:set("owner_source_id", values.owner_source_id)
        source:set("stat", values.stat)
        source:set("operation", values.operation)
        source:set("value", values.value)
        source:set("order", values.order)
    else
        structural:create({ ["d2legacy.stat.source"] = values })
    end
end

local function remove_stats(structural, entities, target, owner_source_id, except)
    for _, entity in ipairs(entities) do
        local source = ecs.get(entity, "d2legacy.stat.source")
        if
            source
            and source:get("target"):id() == target:id()
            and source_owner(source) == owner_source_id
            and not (except and except[source:get("stat")])
        then
            structural:destroy(entity)
        end
    end
end

local function stat_requests(entities, target, owner_source_id)
    local result = {}
    for _, entity in ipairs(entities) do
        local request = ecs.get(entity, "d2legacy.state.stat_request")
        if
            request
            and request:get("target"):id() == target:id()
            and request:get("owner_source_id") == owner_source_id
        then
            result[request:get("stat")] = {
                target = target,
                owner_source_id = owner_source_id,
                source_id = request:get("source_id"),
                stat = request:get("stat"),
                operation = request:get("operation"),
                value = request:get("value"),
                order = request:get("order"),
            }
        end
    end
    return result
end

local function reconcile_stats(structural, entities, request)
    local target = request:get("target")
    local owner_source_id = request:get("source_id")
    local desired = stat_requests(entities, target, owner_source_id)
    if request:get("stat") ~= "" then
        desired[request:get("stat")] = {
            target = target,
            owner_source_id = owner_source_id,
            source_id = owner_source_id,
            stat = request:get("stat"),
            operation = request:get("stat_operation"),
            value = request:get("stat_value"),
            order = request:get("stat_order"),
        }
    end
    local kept = {}
    for stat, value in pairs(desired) do
        assert(stat ~= "" and value.owner_source_id == owner_source_id, "invalid timed-state stat request")
        kept[stat] = true
        reconcile_stat(structural, entities, value)
    end
    remove_stats(structural, entities, target, owner_source_id, kept)
end

local function update_instance(instance, request, expires)
    instance:set("expires_tick", expires)
    for _, field in ipairs({
        "stat",
        "stat_operation",
        "stat_value",
        "stat_order",
        "exclusive_group",
        "on_melee_hit_state_id",
        "on_melee_hit_duration",
        "on_melee_hit_disables_action",
        "action_disabled",
    }) do
        instance:set(field, request:get(field))
    end
end

local function instance_values(request, tick, expires)
    return {
        target = request:get("target"),
        state_id = request:get("state_id"),
        source_id = request:get("source_id"),
        applied_tick = tick,
        expires_tick = expires,
        policy = refresh_policy,
        stat = request:get("stat"),
        stat_operation = request:get("stat_operation"),
        stat_value = request:get("stat_value"),
        stat_order = request:get("stat_order"),
        exclusive_group = request:get("exclusive_group"),
        on_melee_hit_state_id = request:get("on_melee_hit_state_id"),
        on_melee_hit_duration = request:get("on_melee_hit_duration"),
        on_melee_hit_disables_action = request:get("on_melee_hit_disables_action"),
        action_disabled = request:get("action_disabled"),
    }
end

function M.register()
    ecs.system({
        id = "d2legacy.state.timed_instances",
        phase = "effects",
        after = { "d2legacy.state.react_to_melee_hit" },
        query = {
            any = {
                "d2legacy.state.request",
                "d2legacy.state.stat_request",
                "d2legacy.state.instance",
                "d2legacy.stat.source",
            },
        },
        read = {
            "d2legacy.state.request",
            "d2legacy.state.stat_request",
            "d2legacy.state.instance",
            "d2legacy.stat.source",
        },
        write = {
            "d2legacy.state.request",
            "d2legacy.state.stat_request",
            "d2legacy.state.instance",
            "d2legacy.state.event",
            "d2legacy.stat.source",
        },
        update = function(context, entities, structural)
            local touched = {}
            local last_request = {}
            for _, entity in ipairs(entities) do
                local request = ecs.get(entity, "d2legacy.state.request")
                if request then
                    last_request[request_key(request)] = entity:id()
                end
            end
            for _, request_entity in ipairs(entities) do
                local request = ecs.get(request_entity, "d2legacy.state.request")
                if request and last_request[request_key(request)] == request_entity:id() then
                    assert(
                        request:get("state_id") ~= ""
                            and request:get("source_id") ~= ""
                            and request:get("policy") == refresh_policy,
                        "invalid timed-state request"
                    )
                    local match
                    for _, candidate in ipairs(entities) do
                        local instance = ecs.get(candidate, "d2legacy.state.instance")
                        if instance and same(instance, request) then
                            match = candidate
                            break
                        end
                    end
                    local operation = request:get("operation")
                    if operation == "apply" then
                        assert(request:get("duration") > 0, "state duration must be positive")
                        local expires = context.tick + request:get("duration")
                        for _, candidate in ipairs(entities) do
                            local instance = ecs.get(candidate, "d2legacy.state.instance")
                            if instance and group_conflict(instance, request) then
                                touched[candidate:id()] = true
                                remove_stats(structural, entities, instance:get("target"), instance:get("source_id"))
                                emit_instance(
                                    structural,
                                    "state_removed",
                                    context.tick,
                                    instance,
                                    "exclusive_group_replaced"
                                )
                                structural:destroy(candidate)
                            end
                        end
                        if match then
                            update_instance(ecs.get(match, "d2legacy.state.instance"), request, expires)
                            touched[match:id()] = true
                            emit_request(structural, "state_refreshed", context.tick, request, expires, "refresh")
                        else
                            structural:create({
                                ["d2legacy.state.instance"] = instance_values(request, context.tick, expires),
                            })
                            emit_request(structural, "state_applied", context.tick, request, expires, "apply")
                        end
                        reconcile_stats(structural, entities, request)
                    elseif operation == "remove" then
                        if match then
                            local instance = ecs.get(match, "d2legacy.state.instance")
                            touched[match:id()] = true
                            remove_stats(structural, entities, request:get("target"), request:get("source_id"))
                            emit_request(
                                structural,
                                "state_removed",
                                context.tick,
                                request,
                                instance:get("expires_tick"),
                                "explicit"
                            )
                            structural:destroy(match)
                        end
                    else
                        error("unsupported timed-state operation")
                    end
                    structural:destroy(request_entity)
                elseif request then
                    -- Stable entity order makes the last same-tick request for
                    -- one state/source or exclusive group authoritative.
                    structural:destroy(request_entity)
                end
            end
            for _, entity in ipairs(entities) do
                if ecs.get(entity, "d2legacy.state.stat_request") then
                    structural:destroy(entity)
                end
            end
            for _, entity in ipairs(entities) do
                local instance = ecs.get(entity, "d2legacy.state.instance")
                if instance and not touched[entity:id()] and context.tick >= instance:get("expires_tick") then
                    emit_instance(structural, "state_removed", context.tick, instance, "expired")
                    remove_stats(structural, entities, instance:get("target"), instance:get("source_id"))
                    structural:destroy(entity)
                end
            end
        end,
    })
end

return M
