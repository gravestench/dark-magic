-- Only hp1's numeric timed-healing record is admitted by this slice.
local records = require("engine.records/v1")
local M = {}

function M.load(code)
    if code ~= "hp1" then
        return nil
    end
    for _, row in ipairs(records.load("data/global/excel/misc.txt")) do
        if row.code == code then
            local amount, duration = tonumber(row.calc1), tonumber(row.len)
            if
                row.pSpell == "3"
                and row.stat1 == "hpregen"
                and row.state == "healthpot"
                and (row.stat2 or "") == ""
                and (row.stat3 or "") == ""
                and amount
                and amount > 0
                and amount <= 65535
                and amount == math.floor(amount)
                and duration
                and duration >= 1
                and duration <= 65535
                and duration == math.floor(duration)
            then
                return { amount_raw = amount * 256, duration = duration }
            end
            return nil
        end
    end
end

function M.amount(definition, class, vitality, random)
    local amount = definition.amount_raw
    if class == "Amazon" or class == "Paladin" or class == "Assassin" then
        amount = amount + math.floor(amount / 2)
    elseif class == "Barbarian" then
        amount = amount * 2
    end
    -- Engine streams reproduce the recovered distribution, not native unit seeds.
    if vitality > 0 then
        local roll = random.integer("d2legacy.potion.critical", 100)
        local threshold = math.floor(random.integer("d2legacy.potion.vitality", vitality) / 2)
        if roll < threshold then
            amount = amount * 2
        end
    end
    return amount
end

return M
