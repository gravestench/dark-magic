local M = {}

local valid_profiles = {
    authority = true,
    module = true,
    ecs = true,
}

local valid_tiers = {
    fast = true,
    integration = true,
    real_assets = true,
    stress = true,
}

M.generators = {}

function M.mock_module(name, implementation, required_functions)
    assert(type(name) == "string" and name ~= "", "mock module name must be a non-empty string")
    assert(type(implementation) == "table", "mock module implementation must be a table")
    for _, field in ipairs(required_functions or {}) do
        assert(type(implementation[field]) == "function", name .. " mock requires function " .. field)
    end
    setmetatable(implementation, {
        __index = function(_, field)
            error(name .. " mock does not implement " .. tostring(field), 2)
        end,
    })
    package.loaded[name] = implementation
    return implementation
end

function M.unload_module(name)
    package.loaded[name] = nil
end

function M.generators.integer(minimum, maximum)
    assert(type(minimum) == "number" and type(maximum) == "number" and minimum <= maximum, "invalid integer range")
    return function(seed)
        local value = (seed * 1103515245 + 12345) % 2147483648
        return minimum + (value % (maximum - minimum + 1))
    end
end

function M.generators.one_of(values)
    assert(type(values) == "table" and #values > 0, "one_of requires at least one value")
    local index = M.generators.integer(1, #values)
    return function(seed)
        return values[index(seed)]
    end
end

function M.generators.map(generator, transform)
    local valid = type(generator) == "function" and type(transform) == "function"
    assert(valid, "map requires generator and transform functions")
    return function(seed)
        return transform(generator(seed), seed)
    end
end

function M.generators.tuple(...)
    local generators = { ... }
    assert(#generators > 0, "tuple requires at least one generator")
    return function(seed)
        local result = {}
        for index, generator in ipairs(generators) do
            result[index] = generator(seed + (index - 1) * 104729)
        end
        return result
    end
end

local array_metatable = { __d2legacy_test_array = true }

function M.array(values)
    return setmetatable(values or {}, array_metatable)
end

local function fail(message, level)
    error(message, (level or 1) + 1)
end

function M.assert(condition, message)
    if not condition then
        fail(message or "expectation failed", 2)
    end
    return condition
end

local function describe(value)
    if type(value) == "string" then
        return string.format("%q", value)
    end
    return tostring(value)
end

local function deep_difference(actual, expected, path, seen)
    if type(actual) ~= type(expected) then
        return string.format(
            "%s: got %s (%s), want %s (%s)",
            path,
            describe(actual),
            type(actual),
            describe(expected),
            type(expected)
        )
    end
    if type(actual) ~= "table" then
        if actual ~= expected then
            return string.format("%s: got %s, want %s", path, describe(actual), describe(expected))
        end
        return nil
    end
    if actual == expected then
        return nil
    end
    seen[actual] = seen[actual] or {}
    if seen[actual][expected] then
        return nil
    end
    seen[actual][expected] = true
    for key, expected_value in pairs(expected) do
        local difference = deep_difference(actual[key], expected_value, path .. "[" .. describe(key) .. "]", seen)
        if difference then
            return difference
        end
    end
    for key in pairs(actual) do
        if expected[key] == nil then
            return string.format("%s[%s]: unexpected value %s", path, describe(key), describe(actual[key]))
        end
    end
    return nil
end

function M.expect(actual, label)
    local subject = label or "value"
    local assertion = {}

    function assertion:equals(expected)
        if actual ~= expected then
            fail(string.format("%s: got %s, want %s", subject, describe(actual), describe(expected)), 2)
        end
        return self
    end

    function assertion:deep_equals(expected)
        local difference = deep_difference(actual, expected, subject, {})
        if difference then
            fail(difference, 2)
        end
        return self
    end

    function assertion:is_true()
        if actual ~= true then
            fail(string.format("%s: got %s, want true", subject, describe(actual)), 2)
        end
        return self
    end

    function assertion:is_false()
        if actual ~= false then
            fail(string.format("%s: got %s, want false", subject, describe(actual)), 2)
        end
        return self
    end

    function assertion:is_nil()
        if actual ~= nil then
            fail(string.format("%s: got %s, want nil", subject, describe(actual)), 2)
        end
        return self
    end

    function assertion:has_length(expected)
        local ok, length = pcall(function()
            return #actual
        end)
        if not ok or length ~= expected then
            fail(string.format("%s length: got %s, want %d", subject, ok and length or "unavailable", expected), 2)
        end
        return self
    end

    function assertion:is_near(expected, tolerance)
        if type(actual) ~= "number" or math.abs(actual - expected) > tolerance then
            fail(string.format("%s: got %s, want %s +/- %s", subject, describe(actual), expected, tolerance), 2)
        end
        return self
    end

    return assertion
end

function M.entities_with(...)
    local ecs = require("engine.ecs/v1")
    return ecs.query({ all = { ... } })
end

function M.only_entity_with(...)
    local entities = M.entities_with(...)
    M.expect(entities, "matching entities"):has_length(1)
    return entities[1]
end

function M.events(component)
    local ecs = require("engine.ecs/v1")
    local entities = ecs.query({ all = { component } })
    local result = {}
    for _, entity in ipairs(entities) do
        result[#result + 1] = ecs.get(entity, component)
    end
    return result
end

local function action_builder()
    local actions = {}
    local builder = {}

    function builder:run(callback)
        assert(type(callback) == "function", "run callback must be a function")
        actions[#actions + 1] = { run = callback }
        return self
    end

    builder.arrange = builder.run
    builder.check = builder.run

    function builder:step(count)
        actions[#actions + 1] = { step = count or 1 }
        return self
    end

    function builder:update(milliseconds)
        actions[#actions + 1] = { engine_update_ms = milliseconds }
        return self
    end

    function builder:command(kind, payload, options)
        options = options or {}
        actions[#actions + 1] = {
            submit = {
                tick = options.tick or 0,
                sequence = options.sequence or 0,
                player = options.player,
                kind = kind,
                payload = payload,
            },
        }
        return self
    end

    function builder:system_command(kind, payload, options)
        options = options or {}
        actions[#actions + 1] = {
            submit_system = {
                tick = options.tick or 0,
                sequence = options.sequence or 0,
                player = options.player,
                kind = kind,
                payload = payload,
            },
        }
        return self
    end

    function builder:restore_checkpoint()
        actions[#actions + 1] = { checkpoint_restore = true }
        return self
    end

    function builder:expect_checkpoint_parity(steps)
        actions[#actions + 1] = { checkpoint_parity_steps = steps or 1 }
        return self
    end

    return builder, actions
end

function M.case(name, define)
    assert(type(name) == "string" and name ~= "", "case name must be a non-empty string")
    if type(define) == "table" then
        assert(#define > 0, "case " .. name .. " must contain at least one action")
        return { name = name, actions = define }
    end
    assert(type(define) == "function", "case definition must be a function or action array")
    local builder, actions = action_builder()
    define(builder)
    assert(#actions > 0, "case " .. name .. " must contain at least one action")
    return { name = name, actions = actions }
end

function M.run(callback)
    assert(type(callback) == "function", "run callback must be a function")
    return { run = callback }
end

function M.step(count)
    return { step = count or 1 }
end

function M.update(milliseconds)
    return { engine_update_ms = milliseconds }
end

function M.submit(command)
    assert(type(command) == "table", "submitted command must be a table")
    return { submit = command }
end

function M.submit_system(command)
    assert(type(command) == "table", "submitted system command must be a table")
    return { submit_system = command }
end

function M.restore_checkpoint()
    return { checkpoint_restore = true }
end

function M.expect_checkpoint_parity(steps)
    return { checkpoint_parity_steps = steps or 1 }
end

function M.property(name, options, define)
    options = options or {}
    local cases = {}
    local seeds = options.seeds
    if seeds == nil then
        seeds = {}
        for seed = 1, options.samples or 4 do
            seeds[#seeds + 1] = seed
        end
    end
    for _, seed in ipairs(seeds) do
        cases[#cases + 1] = M.case(name .. "_seed_" .. seed, function(test)
            local value = options.generator and options.generator(seed) or seed
            define(test, value, seed)
        end)
    end
    return cases
end

local function add_case(tests, case)
    assert(type(case) == "table" and type(case.name) == "string", "suite cases must be created with test.case")
    assert(tests[case.name] == nil, "duplicate case " .. case.name)
    tests[case.name] = case.actions
end

function M.suite(options)
    options = options or {}
    local profile = options.profile or "authority"
    local tier = options.tier or "fast"
    assert(valid_profiles[profile], "unknown test profile " .. tostring(profile))
    assert(valid_tiers[tier], "unknown test tier " .. tostring(tier))

    assert(options.tests == nil, "tests maps are retired; declare cases with test.case")
    local tests = {}
    for _, case in ipairs(options.cases or {}) do
        if case.name then
            add_case(tests, case)
        else
            for _, generated in ipairs(case) do
                add_case(tests, generated)
            end
        end
    end

    return {
        api_version = 1,
        name = options.name,
        profile = profile,
        tier = tier,
        covers = options.covers or {},
        seed = options.seed,
        initial_data = options.initial_data,
        records = options.records,
        disable_execution_budget = options.disable_execution_budget,
        tests = tests,
    }
end

return M
