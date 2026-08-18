-- Materialize one trusted durable character as a live Diablo II player.
--
-- Go owns save I/O, authentication, and command admission. The meaning of a
-- Diablo player entity belongs here: its components, initial mouse skills,
-- belt shape, collision size, and legacy composite appearance are mod policy.

local commands = require("engine.authority_command/v1")
local ecs = require("engine.ecs/v1")
local skills = require("d2legacy.data.skill")
local player_stats = require("d2legacy.data.player_stats")
local item_bootstrap = require("d2legacy.items.bootstrap")
local game_rules = require("d2legacy.policy.game_rules")
local movement_rules = require("d2legacy.movement_rules/v1")
local M = {}

local initial_available, initial = pcall(require, "engine.initial_data/v1")

local function finite(value)
    return type(value) == "number" and value == value and value ~= math.huge and value ~= -math.huge
end

local function present(value)
    return type(value) == "string" and value:match("%S") ~= nil
end

local function already_entered(player_id)
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.player.identity" } })) do
        if ecs.get(entity, "d2legacy.player.identity"):get("player") == player_id then
            return true
        end
    end
    return false
end

function M.validate(command)
    local p = command.payload
    assert(type(p) == "table", "player entry payload must be a table")
    for _, field in ipairs({ "character_id", "player", "name", "class" }) do
        assert(present(p[field]), "player entry " .. field .. " is required")
    end
    assert(p.level >= 1 and p.health >= 0 and p.max_health >= p.health, "player entry progression or health is invalid")
    assert(p.mana >= 0 and p.max_mana >= p.mana, "player entry mana is invalid")
    assert(p.stamina >= 0 and p.max_stamina >= p.stamina, "player entry stamina is invalid")
    local _, _, _, _, starting_vitality = movement_rules.class_facts(p.class)
    assert(p.vitality >= starting_vitality, "player entry vitality is below the class starting value")
    assert(
        finite(p.x) and finite(p.y) and finite(p.world_width) and finite(p.world_height),
        "player entry world values must be finite"
    )
    assert(
        p.world_width > 0
            and p.world_height > 0
            and p.x >= 0
            and p.y >= 0
            and p.x < p.world_width
            and p.y < p.world_height,
        "player entry is outside the world"
    )
    assert(p.act >= 1 and p.act <= 5 and p.level_id > 0, "player entry location is invalid")
end

local class_tokens = {
    amazon = "AM",
    sorceress = "SO",
    necromancer = "NE",
    paladin = "PA",
    barbarian = "BA",
    assassin = "AI",
    druid = "DZ",
}

local function initial_skills(skills, preferred)
    local left, right, left_chosen, right_chosen = 0, 0, false, false
    for _, skill in ipairs(skills or {}) do
        assert(skill.id >= 0 and skill.level >= 1 and skill.list_row >= 0, "learned skill is invalid")
        assert(skill.left_allowed or skill.right_allowed, "learned skill is not assignable")
        if not left_chosen and skill.left_allowed then
            left, left_chosen = skill.id, true
        end
        if not right_chosen and skill.right_allowed then
            right, right_chosen = skill.id, true
        end
    end
    preferred = preferred or {}
    for _, skill in ipairs(skills or {}) do
        if preferred.left == skill.id then
            assert(skill.left_allowed, "preferred left skill is not left-assignable")
            left, left_chosen = skill.id, true
        end
        if preferred.right == skill.id then
            assert(skill.right_allowed, "preferred right skill is not right-assignable")
            right, right_chosen = skill.id, true
        end
    end
    return left, right
end

local function configured_skills(class)
    local learned = skills.starting_for_class(class)
    local config = initial_available and initial.get("d2legacy.development_skills") or nil
    if type(config) ~= "table" or config.enabled ~= true then
        return learned, {}
    end
    local by_id = {}
    if config.replace ~= true then
        for _, skill in ipairs(learned) do
            by_id[skill.id] = skill
        end
    end
    local configured_ids = config.skill_ids or {}
    if config.all_implemented == true then
        configured_ids = {}
        local coverage = require("d2legacy.data.skill_behavior_coverage").load()
        for skill_id in pairs(coverage.by_id) do
            configured_ids[#configured_ids + 1] = skill_id
        end
        table.sort(configured_ids)
    end
    for _, skill in ipairs(skills.learned_for_ids(configured_ids, config.skill_level or 1)) do
        by_id[skill.id] = skill
    end
    learned = {}
    for _, skill in pairs(by_id) do
        learned[#learned + 1] = skill
    end
    table.sort(learned, function(left, right)
        return left.id < right.id
    end)
    return learned, { left = config.left, right = config.right }
end

local function create_passive_sources(player, sources)
    local seen = {}
    for index, source in ipairs(sources or {}) do
        assert(
            source.stat == "attack_rating"
                or source.stat == "defense"
                or source.stat == "attackrate"
                or source.stat == "item_fasterattackrate"
                or source.stat == "velocitypercent"
                or source.stat == "item_fastermovevelocity"
                or source.stat == "staminarecoverybonus"
                or source.stat == "manarecoverybonus"
                or source.stat == "manarecovery"
                or source.stat == "item_staminadrainpct"
                or source.stat == "vitality"
                or source.stat == "maxstamina"
                or source.stat == "skill_staminapercent"
                or source.stat == "skill_passive_staminapercent"
                or source.stat == "item_stamina_perlevel"
                or source.stat == "item_stamina_bytime",
            "unsupported passive combat stat"
        )
        assert(
            source.operation == nil or source.operation == "add" or source.operation == "percent",
            "unsupported passive stat operation"
        )
        if
            source.stat == "vitality"
            or source.stat == "maxstamina"
            or source.stat == "skill_staminapercent"
            or source.stat == "skill_passive_staminapercent"
            or source.stat == "item_stamina_perlevel"
            or source.stat == "item_stamina_bytime"
        then
            assert(source.operation == nil or source.operation == "add", "stamina operand must be additive")
        end
        assert(
            type(source.value) == "number" and source.value == math.floor(source.value),
            "passive stat value must be an integer"
        )
        local source_id = "passive:" .. tostring(source.id or index)
        assert(not seen[source_id], "duplicate passive stat source " .. source_id)
        seen[source_id] = true
        ecs.create({
            ["d2legacy.stat.source"] = {
                target = player,
                source_id = source_id,
                stat = source.stat,
                operation = source.operation or "add",
                value = source.value,
                order = source.order or index,
            },
        })
    end
end

function M.apply(command)
    local p = command.payload
    assert(not already_entered(p.player), "player already entered")
    assert(#ecs.query({ all = { "d2legacy.player.identity" } }) < game_rules.get().maximum_players, "game is full")
    local difficulty = game_rules.difficulty()
    assert(
        p.difficulty == nil or p.difficulty == difficulty,
        "player entry difficulty differs from immutable game rules"
    )
    local learned, preferred_skills = configured_skills(p.class)
    local _, run_drain = movement_rules.class_facts(p.class)
    local left, right = initial_skills(learned, preferred_skills)
    local player = ecs.create({
        ["d2legacy.player.identity"] = {
            character_id = p.character_id,
            player = p.player,
            name = p.name,
            class = p.class,
        },
        ["d2legacy.player.progress"] = {
            level = p.level,
            experience = p.experience,
            unspent_skill_points = p.unspent_skill_points or 0,
        },
        ["d2legacy.player.difficulty"] = {
            current = difficulty,
            highest_completed = p.highest_completed_difficulty or -1,
        },
        ["d2legacy.player.combat_stats"] = {
            base_attack_rating = player_stats.base_attack_rating(p.class, p.dexterity, p.base_attack_rating),
            base_defense = player_stats.base_defense(p.dexterity, p.defense),
            attack_rating = player_stats.base_attack_rating(p.class, p.dexterity, p.base_attack_rating),
            defense = player_stats.base_defense(p.dexterity, p.defense),
        },
        ["d2legacy.combat.action_rate"] = {
            base_attack_rate = 100,
            attack_rate = 100,
            item_fasterattackrate = 0,
        },
        ["d2legacy.combat.defense"] = {
            base_physical_resist = p.physical_resistance or 0,
            base_fire_resist = p.fire_resistance or 0,
            base_cold_resist = p.cold_resistance or 0,
            base_lightning_resist = p.lightning_resistance or 0,
            physical_resist = p.physical_resistance or 0,
            fire_resist = p.fire_resistance or 0,
            cold_resist = p.cold_resistance or 0,
            lightning_resist = p.lightning_resistance or 0,
            max_fire_resist = 75,
            max_cold_resist = 75,
            max_lightning_resist = 75,
            physical_reduction_raw = 0,
        },
        ["d2legacy.player.vitals"] = {
            health = p.health,
            max_health = p.max_health,
            mana = p.mana,
            max_mana = p.max_mana,
            mana_raw = p.mana * 256,
            max_mana_raw = p.max_mana * 256,
            stamina = p.stamina,
            max_stamina = p.max_stamina,
            stamina_raw = p.stamina * 256,
            max_stamina_raw = p.max_stamina * 256,
        },
        ["d2legacy.player.resource_stats"] = {
            mana_regen_frames = player_stats.mana_regen_frames(p.class),
            manarecoverybonus = 0,
            manarecovery = 0,
        },
        ["d2legacy.player.stamina_progression"] = {
            base_vitality = p.vitality,
            vitality = p.vitality,
            last_level = p.level,
        },
        ["d2legacy.player.movement_stats"] = {
            run_drain = run_drain,
            velocitypercent = 0,
            item_fastermovevelocity = 0,
            staminarecoverybonus = 0,
            item_staminadrainpct = 0,
            armor_run_drain = 0,
        },
        -- D2 synthesizes an unarmed profile when no weapon contributes one.
        ["d2legacy.combat.melee_profile"] = {
            range = 2,
            physical_min = 256,
            physical_max = 512,
            primary_hand = "unarmed",
            primary_weapon_attack_rate = 0,
            secondary_hand = "",
            secondary_weapon_attack_rate = 0,
            dual_wield = false,
        },
        ["d2legacy.player.appearance"] = {
            cof = p.cof or "",
            token = assert(class_tokens[string.lower(p.class)], "unknown player class"),
            palette = present(p.palette) and p.palette or "data/global/Palette/units/pal.dat",
            weapon_class = "HTH",
        },
        ["d2legacy.player.animation"] = { direction = p.direction or 0, mode = "NU", start_tick = command.tick },
        ["d2legacy.world.position"] = { x = p.x, y = p.y },
        ["d2legacy.world.velocity"] = { x = 0, y = 0 },
        ["d2legacy.world.facing"] = { direction = p.direction or 0, directions = 16 },
        ["d2legacy.player.movement_mode"] = { running = false },
        ["d2legacy.player.motion"] = {
            owner = "none",
            kind = "idle",
            x = 0,
            y = 0,
            target_x = p.x,
            target_y = p.y,
            has_target = false,
            running = false,
            active = false,
        },
        ["d2legacy.player.skill_assignment"] = { left = left, right = right },
        ["d2legacy.player.skill_intent"] = { side = "", skill_id = 0, target_x = p.x, target_y = p.y, target_id = "" },
        ["d2legacy.player.belt"] = { capacity = 4 },
        ["d2legacy.world.player_control"] = { player = p.player },
        ["d2legacy.world.bounds"] = { width = p.world_width, height = p.world_height },
        ["d2legacy.world.location"] = { act = p.act, level_id = p.level_id },
        -- A player occupies two subtiles across in legacy D2: radius one.
        ["d2legacy.world.collider"] = { radius = 1 },
        ["d2legacy.world.occupancy"] = { blocks_movement = true },
        ["d2legacy.world.selectable"] = {
            id = "player:" .. p.player,
            kind = "player",
            label = p.name,
            owner = p.player,
            radius = 0.75,
            priority = 10,
        },
    })
    item_bootstrap.ensure_player(p.player)
    create_passive_sources(player, p.passive_stat_sources)
    for _, skill in ipairs(learned) do
        ecs.create({
            ["d2legacy.player.learned_skill"] = {
                owner = player,
                skill_id = skill.id,
                level = skill.level,
                list_row = skill.list_row,
                left_allowed = skill.left_allowed,
                right_allowed = skill.right_allowed,
            },
            ["d2legacy.player.skill_hard_level"] = { level = skill.hard_level or skill.level },
        })
    end
end

function M.register()
    commands.register({
        kind = "system.player.enter",
        authorities = { "system", "administrator" },
        validate = M.validate,
        apply = M.apply,
    })
end

return M
