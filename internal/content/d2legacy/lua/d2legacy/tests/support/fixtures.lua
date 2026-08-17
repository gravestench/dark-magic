-- Semantic fixtures describe game concepts while the harness owns JSON encoding.
-- Defaults live in one place so test data evolves with production commands.

local M = {}
local test = require("d2legacy.tests/v1")

local function copy(value)
    if type(value) ~= "table" then
        return value
    end
    local result = {}
    for key, item in pairs(value) do
        result[key] = copy(item)
    end
    return result
end

function M.player_entry(overrides)
    local entry = {
        character_id = "hero",
        player = "alice",
        name = "Hero",
        class = "Amazon",
        level = 1,
        experience = 0,
        dexterity = 20,
        defense = 0,
        health = 50,
        max_health = 50,
        mana = 20,
        max_mana = 20,
        expansion = true,
        hardcore = false,
        cof = "",
        palette = "units",
        direction = 0,
        mode = "NU",
        x = 10,
        y = 12,
        world_width = 100,
        world_height = 80,
        act = 1,
        level_id = 1,
        skills = test.array(),
    }
    for key, value in pairs(overrides or {}) do
        entry[key] = copy(value)
    end
    return entry
end

function M.monster_spawn(overrides)
    local spawn = {
        spawn_id = "fallen",
        seed = 1,
        x = 4,
        y = 0,
        act = 1,
        level_id = 1,
        definition = {
            id = "fallen",
            base_id = "fallen",
            graphics_id = "fallen",
            name_key = "Fallen",
            ai = "fallen",
            token = "FA",
            weapon_class = "HTH",
            components = {},
            life_min = 4096,
            life_max = 4096,
            level = 1,
            defense = 0,
            attack_rating = 0,
            physical_min = 0,
            physical_max = 0,
            experience = 0,
            treasure_class = "",
            collider_radius = 0.5,
            select_radius = 0.5,
            velocity = 0,
            think_interval = 100,
            aggro_radius = 0,
            attack_range = 1,
        },
    }
    for key, value in pairs(overrides or {}) do
        spawn[key] = copy(value)
    end
    return spawn
end

function M.item(overrides)
    local item = {
        id = "item-1",
        code = "ssd",
        name = "Short Sword",
        container = "inventory",
        x = 0,
        y = 0,
        width = 1,
        height = 3,
        quantity = 1,
        identified = true,
        ethereal = false,
    }
    for key, value in pairs(overrides or {}) do
        item[key] = copy(value)
    end
    return item
end

function M.equipment(overrides)
    local equipment = {
        active_weapon_set = 0,
        items = test.array(),
    }
    for key, value in pairs(overrides or {}) do
        equipment[key] = copy(value)
    end
    return equipment
end

function M.owned_unit(overrides)
    local unit = {
        owner = "alice",
        category = "summon",
        group = "default",
        entity = 1,
        created_tick = 0,
    }
    for key, value in pairs(overrides or {}) do
        unit[key] = copy(value)
    end
    return unit
end

function M.command(kind, payload, overrides)
    local command = {
        tick = 0,
        sequence = 0,
        kind = kind,
        payload = copy(payload or {}),
    }
    for key, value in pairs(overrides or {}) do
        command[key] = copy(value)
    end
    return command
end

M.amazon_entry = M.player_entry()
M.amazon_level_up_entry = M.player_entry({ experience = 5 })
M.straight_missile_entry = M.player_entry({
    level = 99,
    x = 0,
    y = 0,
    mana = 10,
    max_mana = 10,
    skills = {
        {
            id = 36,
            level = 1,
            list_row = 0,
            left_allowed = true,
            right_allowed = true,
        },
    },
})

return M
