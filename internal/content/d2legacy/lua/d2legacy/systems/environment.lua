-- Advance the checkpointed per-act day cycle used by ItemStatCost op 6.

local ecs = require("engine.ecs/v1")
local M = {}

local normal = {
    { begin = 320, period = 3 },
    { begin = 340, period = 3 },
    { begin = 0, period = 0 },
    { begin = 160, period = 1 },
    { begin = 180, period = 1 },
    { begin = 200, period = 2 },
}

local act_four = {
    { begin = 340, period = 3 },
    { begin = 350, period = 3 },
    { begin = 0, period = 0 },
    { begin = 180, period = 1 },
    { begin = 190, period = 1 },
    { begin = 200, period = 2 },
}

local eclipse = {
    { begin = 300, period = 3 },
    { begin = 0, period = 0 },
    { begin = 60, period = 1 },
    { begin = 120, period = 2 },
    { begin = 180, period = 2 },
    { begin = 240, period = 2 },
}

local function cycle_for(index, act, eclipsed)
    if act == 4 then
        return act_four[index + 1]
    end
    if eclipsed then
        return eclipse[index + 1]
    end
    return normal[index + 1]
end

local function advance(environment)
    local act = environment:get("act")
    local eclipsed = environment:get("eclipse")
    local period = environment:get("period_of_day")
    local rate = environment:get("time_rate")
    local ticks = environment:get("ticks") + 1
    if not eclipsed then
        if act == 4 then
            ticks = ticks + 15
        elseif period == 2 then
            ticks = ticks + 1
            if act == 3 then
                ticks = ticks + 9
            end
        end
    end
    if ticks >= 360 * rate then
        ticks = 0
    end
    local next_index = (environment:get("cycle_index") + 1) % 6
    local next_cycle = cycle_for(next_index, act, eclipsed)
    if ticks > rate * next_cycle.begin then
        environment:set("cycle_index", next_index)
        -- The runtime uses the normal cycle here for non-eclipse acts, even
        -- though Act IV uses its own boundary table above.
        local entered = eclipsed and eclipse[next_index + 1] or normal[next_index + 1]
        environment:set("period_of_day", entered.period)
        ticks = rate * entered.begin
    end
    environment:set("ticks", ticks)
end

function M.register()
    ecs.system({
        id = "d2legacy.world.environment_cycle",
        phase = "pre_simulation",
        query = {
            any = { "d2legacy.world.environment", "d2legacy.world.location" },
            none = { "d2legacy.world.inactive" },
        },
        read = { "d2legacy.world.environment", "d2legacy.world.location" },
        write = { "d2legacy.world.environment" },
        update = function(_, entities, structural)
            local environments = {}
            local acts = {}
            for _, entity in ipairs(entities) do
                local environment = ecs.get(entity, "d2legacy.world.environment")
                if environment then
                    environments[environment:get("act")] = true
                    advance(environment)
                end
                local location = ecs.get(entity, "d2legacy.world.location")
                if location then
                    acts[location:get("act")] = true
                end
            end
            for act = 1, 5 do
                if acts[act] and not environments[act] then
                    structural:create({
                        ["d2legacy.world.environment"] = {
                            act = act,
                            cycle_index = 2,
                            period_of_day = 0,
                            ticks = 0,
                            time_rate = 128,
                            eclipse = false,
                        },
                    })
                end
            end
        end,
    })
end

return M
