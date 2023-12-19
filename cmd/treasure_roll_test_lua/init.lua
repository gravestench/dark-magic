function sprintf(format, ...)
    local args = {...}
    local numArgs = select("#", ...)
    local index = 1
    local result = format:gsub("%%.", function(subst)
        local specifier = subst:sub(-1)
        if specifier == "s" then
            local arg = args[index]
            index = index + 1
            return tostring(arg)
        elseif specifier == "d" or specifier == "i" or specifier == "f" then
            local arg = args[index]
            index = index + 1
            return string.format("%" .. subst, arg)
        end
        return subst
    end)
    return result
end

function randomIntn(n)
    -- Generate a random number between 0 and (n-1)
    local randomNumber = math.random(n)
    return randomNumber
end

function round(num)
    local intPart = num / 1
    local fracPart = num - intPart
    if fracPart >= 0.5 then
        return intPart + 1
    end
    return intPart
end

function filterPredicate(inputTable, predicateFunc)
    local filteredTable = {}

    for _, obj in ipairs(inputTable) do
        if predicateFunc(obj) then
            table.insert(filteredTable, obj)
        end
    end

    return filteredTable
end

function filterEmptyStrings(array)
    local filteredArray = {}
    for i = 1, #array do
        local value = array[i]
        if value ~= "" then
            table.insert(filteredArray, i, value)
        end
    end
    return filteredArray
end

function monsterIdEquals(monster, id)
   return monster.BaseId == id
end

function randomPick(array)
    return array[randomIntn(#array)]
end

-- Function to pick an object based on probability from separate arrays
function pickObjectByProbability(drops, probs)
    assert(#drops == #probs, "The number of drops and probabilities must be the same.")

    local sumOfProbabilities = 0

    -- Calculate the sum of probabilities
    for i = 1, #probs do
        sumOfProbabilities = sumOfProbabilities + probs[i]
    end

    -- Generate a random number between 0 and the sum total
    local randomValue = math.random() * sumOfProbabilities

    -- Pick the object based on the rolled value
    local pickedObject = nil
    local currentSum = 0
    for i = 1, #probs do
        currentSum = currentSum + probs[i]
        if randomValue <= currentSum then
            pickedObject = drops[i]
            break
        end
    end

    return randomValue, pickedObject
end

function randomLoot(monsterID, lvlMon, lvlChar, lvlArea)
    print("generating random loot: \n\tMonster ID: " .. monsterID .. "\n\tMonster Level: " .. lvlMon .. "\n\tCharacter Level: " .. lvlChar .. "\n\tArea Level: " .. lvlArea)

    -- get the monster record
    local isMonster = function(monster)
        return monsterIdEquals(monster, monsterID)
    end

    -- filter all records with the same base ID
    monsterRecords = filterPredicate(records.MonsterStats, isMonster)
    if (#monsterRecords < 1) then
        print("could not find monster by id: " .. monsterID)
        return
    end

    print("matching monster records: " .. #monsterRecords)

    -- pick one of the records
    chosenMonsterRecord = randomPick(monsterRecords)
    print("chosen monster:'" .. chosenMonsterRecord.Id .. "'")

    -- get the treasure record
    local matchTreasure = function(t)
        local r = chosenMonsterRecord
        return t.TreasureClass == r.TreasureClass1
    end

    treasureRecords = filterPredicate(records.Treasures, matchTreasure)
    if #treasureRecords < 1 then
        print("could not find treasure record")
        return
    end

    print("matching treasure records: " .. #treasureRecords)

    chosenTreasure = randomPick(treasureRecords)
    print("chosen treasure:'" .. chosenTreasure.TreasureClass .. "'")
    print("Number of picks: " .. chosenTreasure.Picks)

    local t = chosenTreasure
    drops = {t.Item1, t.Item2, t.Item3, t.Item4, t.Item5, t.Item6, t.Item7, t.Item8}
    probs = {t.Prob1, t.Prob2, t.Prob3, t.Prob4, t.Prob5, t.Prob6, t.Prob7, t.Prob8}

    sum = 0
    for i, j in ipairs(drops) do
        sum = sum + probs[i]
    end

    print("sum of all probabilities: " .. sum)

    -- print the drops and their probabilities
    for i, j in ipairs(drops) do
        if (probs[i]+0) < 1 then goto continue end
        chance = round(probs[i]/sum*1000)
        chance = chance / 10
        print("Drop/probability info #" .. i ..": '".. j .."', ".. probs[i] .. " (" .. chance .."%%)")
        ::continue::
    end

    picks = {}
    for _ = 1, chosenTreasure.Picks do
        roll, picked = pickObjectByProbability(drops, probs)
        print("rolled value: " .. roll .. ", picked: '" .. picked .."'")
        table.insert(picks, picked)
    end

    return picks
end

for _ = 1, 3 do
    randomMonsterRecord = records.MonsterStats[randomIntn(#records.MonsterStats)]
    randomLoot(randomMonsterRecord.BaseId, 50, 50, 50)
end