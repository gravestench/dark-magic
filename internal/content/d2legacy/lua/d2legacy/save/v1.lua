-- d2legacy character creation and roster policy.
--
-- The engine store only persists opaque records. This module owns the legacy
-- class set, character-name grammar, storage-ID convention, and creation
-- defaults. Keeping those decisions here lets another mod use the same durable
-- store without inheriting Diablo II rules.

local store = require("d2legacy.save_store/v1")

local save = {}

local canonical_classes = {
    amazon = "Amazon",
    sorceress = "Sorceress",
    necromancer = "Necromancer",
    paladin = "Paladin",
    barbarian = "Barbarian",
    assassin = "Assassin",
    druid = "Druid",
}

local function trim(value)
    return tostring(value or ""):match("^%s*(.-)%s*$")
end

local function canonical_class(value)
    return canonical_classes[string.lower(trim(value))]
end

local function valid_name(name)
    if #name < 2 or #name > 15 then
        return false, "character name must contain 2 to 15 characters"
    end
    if not name:match("^[A-Za-z][A-Za-z'-]*[A-Za-z]$") then
        return false, "character name may contain ASCII letters and single internal hyphens or apostrophes"
    end
    if name:find("[-'][-']") then
        return false, "character name may not contain adjacent punctuation"
    end
    return true
end

local function storage_id(class, name)
    local id = string.lower(class .. "-" .. name)
    id = id:gsub("[ '%-]+", "-"):gsub("[^a-z0-9%-]", "")
    return id:gsub("^%-+", ""):gsub("%-+$", "")
end

local function name_is_available(name)
    local wanted = string.lower(name)
    for _, existing in ipairs(store.characters()) do
        if string.lower(existing.name) == wanted then
            return false
        end
    end
    return true
end

function save.create_named(name, class, expansion, hardcore)
    name = trim(name)
    class = canonical_class(class)
    local ok, reason = valid_name(name)
    if not ok then return nil, reason end
    if not class then return nil, "unsupported character class" end
    if not name_is_available(name) then return nil, "character name already exists" end

    return store.create_selected({
        id = storage_id(class, name),
        name = name,
        class = class,
        level = 1,
        expansion = expansion ~= false,
        hardcore = hardcore == true,
    })
end

-- Imported saves already carry an identity. They still pass through the same
-- d2legacy validation instead of bypassing policy in the Go storage adapter.
function save.create(id, name, class, expansion, hardcore, level)
    name = trim(name)
    class = canonical_class(class)
    local ok, reason = valid_name(name)
    if not ok then return nil, reason end
    if not class then return nil, "unsupported character class" end
    if trim(id) == "" then return nil, "character ID is required" end
    if not name_is_available(name) then return nil, "character name already exists" end
    return store.create({
        id = trim(id), name = name, class = class,
        level = math.max(1, math.floor(tonumber(level) or 1)),
        expansion = expansion ~= false, hardcore = hardcore == true,
    })
end

save.characters = store.characters
save.select = store.select
save.delete = store.delete
save.selected = store.selected

return save
