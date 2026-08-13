-- Resolve one derived integer stat from its durable base and named sources.
--
-- Diablo keeps flat and percentage contributions distinct. Stable sorting is
-- still important for diagnostics and replay evidence even though additions
-- within one phase are commutative. Percentages apply once, after every flat
-- contribution, and integer truncation happens at the phase boundary.

local M = {}

local function ordered(sources)
    local result = {}
    for _, source in ipairs(sources or {}) do
        result[#result + 1] = source
    end
    table.sort(result, function(left, right)
        if left.order ~= right.order then
            return left.order < right.order
        end
        return left.id < right.id
    end)
    return result
end

function M.resolve(base, sources)
    local flat = base
    local percent = 0
    for _, source in ipairs(ordered(sources)) do
        local operation = source.operation or "add"
        if operation == "add" then
            flat = flat + source.value
        elseif operation == "percent" then
            percent = percent + source.value
        else
            error("unsupported stat operation " .. tostring(operation))
        end
    end
    return math.floor(flat * (100 + percent) / 100)
end

function M.local_value(base, percent)
    return math.floor(base * (100 + percent) / 100)
end

return M
