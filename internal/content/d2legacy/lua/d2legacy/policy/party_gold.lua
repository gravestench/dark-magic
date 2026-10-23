-- Supplied-input party gold split from GPS-R01/R02. This reports original
-- add/replace requests, not proof that the original stat writer stored them.
local M = {}

local U32 = 4294967296
local I32_SIGN = 2147483648

local function signed32(value)
    local wrapped = value % U32
    if wrapped >= I32_SIGN then
        return wrapped - U32
    end
    return wrapped
end

local function truncate(value)
    if value < 0 then
        return math.ceil(value)
    end
    return math.floor(value)
end

local function capacity(gold, requested, level, safe_production)
    local cap = signed32(level * 10000)
    local accepted
    local excess
    if safe_production then
        -- A vendor may already have put this layout above the provisional
        -- level cap. Never debit that balance when collecting a new pile.
        accepted = math.max(0, math.min(requested, cap - gold))
        excess = requested - accepted
    else
        local proposed = signed32(gold + requested)
        accepted = requested
        excess = 0
        if proposed > cap then
            accepted = signed32(cap - gold)
            excess = signed32(requested - accepted)
        end
    end

    local next_balance = signed32(gold + accepted)
    local kind = "add"
    local amount = accepted
    if not safe_production and (next_balance < 0 or next_balance > cap) then
        kind = "replace"
        amount = 0
        next_balance = 0
    end
    if safe_production then
        next_balance = gold + accepted
    end
    return { accepted = accepted, kind = kind, amount = amount, next_balance = next_balance }, excess
end

function M.allocate(amount, picker_id, members, safe_production)
    local picker
    local count = 0
    for _, member in ipairs(members) do
        if member.id == picker_id then
            assert(not picker, "duplicate party gold picker")
            picker = member
        end
        if member.counts_for_divisor then
            count = count + 1
        end
    end
    assert(picker, "party gold picker is absent")

    local requests = {}
    local remaining = amount
    if count > 1 then
        local share = truncate(amount / count)
        for _, member in ipairs(members) do
            if member.id ~= picker_id and member.can_receive then
                local request = capacity(member.gold, share, member.level, safe_production)
                request.id = member.id
                requests[#requests + 1] = request
                remaining = signed32(remaining - request.accepted)
            end
        end
    end

    local picker_request, excess = capacity(picker.gold, remaining, picker.level, safe_production)
    picker_request.id = picker_id
    requests[#requests + 1] = picker_request
    return { requests = requests, excess = excess, count = count }
end

return M
