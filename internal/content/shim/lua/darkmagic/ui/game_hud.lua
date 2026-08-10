-- Authored in-game HUD composition for the supported 800x600 profile.
--
-- The manifest owns asset/frame/placement facts. This module owns only the
-- disposable render-node lifecycle and value-driven clipping behavior.
local render = require("dm.render/v1")
local scenes = require("dm.scene/v1")
local locale = require("dm.locale/v1")
local controls = require("darkmagic.ui.controls")
local tooltip = require("darkmagic.ui.tooltip")

local M = {}

local function dc6_at(root, sheet, palette, frame, x, y)
    local node = render.create("hud", root)
    local width, height = node:set_dc6(sheet, palette, 0, frame)
    node:set_position(x + width / 2, y + height / 2)
    return node, width, height
end

local function dc6_bottom(root, sheet, palette, frame, x, bottom)
    local node = render.create("hud", root)
    local width, height = node:set_dc6(sheet, palette, 0, frame)
    node:set_position(x + width / 2, bottom - height / 2)
    return node
end

local function add_globe(root, definition, sheet, overlap_sheet, palette)
    local liquid = dc6_at(root, sheet, palette, definition.frame, definition.x, definition.y)
    dc6_at(
        root,
        overlap_sheet,
        palette,
        definition.overlap_frame,
        definition.overlap_x,
        definition.overlap_y
    )
    return liquid
end

local function ratio(value, maximum)
    if type(value) ~= "number" or type(maximum) ~= "number" or maximum <= 0 then return 0 end
    return math.max(0, math.min(1, value / maximum))
end

local function clip_globe(node, definition, fill)
    local visible_height = math.floor(definition.height * fill + 0.5)
    node:set_visible(visible_height > 0)
    node:set_clip(
        definition.x,
        definition.y + definition.height - visible_height,
        definition.width,
        math.max(1, visible_height)
    )
end

local function add_status_control(hud, id, definition, label, tip)
    hud.controls:add({
        id = id,
        label = label,
        x = definition.x,
        y = definition.y,
        width = definition.width,
        height = definition.height,
        focusable = false,
        on_state = function(_, state) tip:set_visible(state == "hover") end,
    })
end

function M.create(root, definition, palettes, commands)
    commands = commands or {}
    local hud = { root = render.create("hud", root), controls = controls.new(), running = false, menu_open = false, definition = definition }
    local palette = palettes[definition.palette]

    for _, part in ipairs(definition.panel_parts) do
        dc6_bottom(hud.root, definition.panel_sheet, palette, part.frame, part.x, part.bottom)
    end

    local globes = definition.globes
    hud.health_liquid = add_globe(hud.root, globes.health, globes.sheet, globes.overlap_sheet, palette)
    hud.mana_liquid = add_globe(hud.root, globes.mana, globes.sheet, globes.overlap_sheet, palette)

    hud.bars = {}
    for name, bar in pairs({ stamina = definition.stamina, experience = definition.experience }) do
        local node = render.create("hud", hud.root)
        hud.bars[name] = { node = node, definition = bar, pixels = -1 }
    end

    hud.tips = {
        health = tooltip.create(hud.root, "", globes.health.x + globes.health.width / 2, globes.health.y, {}),
        mana = tooltip.create(hud.root, "", globes.mana.x + globes.mana.width / 2, globes.mana.y, {}),
        stamina = tooltip.create(hud.root, "", definition.stamina.x + definition.stamina.width / 2, definition.stamina.y, {}),
        experience = tooltip.create(hud.root, "", definition.experience.x + definition.experience.width / 2, definition.experience.y, {}),
    }
    add_status_control(hud, "health", globes.health, assert(locale.text("darkmagic.hud.health")), hud.tips.health)
    add_status_control(hud, "mana", globes.mana, assert(locale.text("darkmagic.hud.mana")), hud.tips.mana)
    add_status_control(hud, "stamina", definition.stamina, assert(locale.text("darkmagic.hud.stamina")), hud.tips.stamina)
    add_status_control(hud, "experience", definition.experience, assert(locale.text("darkmagic.hud.experience")), hud.tips.experience)

    local skills = definition.skills
    hud.skills = {}
    for _, skill in ipairs({
        { side = "left", placement = skills.left },
        { side = "right", placement = skills.right },
    }) do
        local side = skill.side
        local placement = skill.placement
        local node = dc6_at(hud.root, skills.sheet, palette, skills.frame, placement.x, placement.y)
        local tip = tooltip.create(hud.root, "", placement.x + skills.width / 2, placement.y, {})
        hud.skills[side] = { node = node, tip = tip, id = -1 }
        hud.controls:add({
            id = side .. "_skill",
            label = side .. " skill",
            x = placement.x,
            y = placement.y,
            width = skills.width,
            height = skills.height,
            focusable = false,
            on_state = function(_, state) tip:set_visible(state == "hover") end,
        })
    end

    local run = definition.run
    local run_node = dc6_at(hud.root, run.sheet, palette, run.walk_frame, run.x, run.y)
    local run_tip = tooltip.create(hud.root, assert(locale.text("darkmagic.hud.walk")), run.x + run.width / 2, run.y, {})
    hud.run_node = run_node
    hud.run_tip = run_tip
    hud.palette = palette
    hud.controls:add({
        id = "run",
        label = "Run/Walk",
        x = run.x,
        y = run.y,
        width = run.width,
        height = run.height,
        on_activate = function()
            if commands.request_running then commands.request_running(not hud.running) end
        end,
        on_state = function(_, state) run_tip:set_visible(state == "hover") end,
    })

    local minipanel = definition.minipanel
    local minipanel_node = dc6_at(hud.root, minipanel.sheet, palette, 0, minipanel.x, minipanel.y)
    minipanel_node:set_visible(false)
    for index, button in ipairs(minipanel.buttons) do
        local button_definition = button
        local node = dc6_at(
            hud.root,
            minipanel.button_sheet,
            palette,
            button_definition.frame,
            minipanel.button_x + (index - 1) * minipanel.button_step,
            minipanel.button_y
        )
        node:set_visible(false)
        local control = hud.controls:add({
            id = "minipanel_" .. button_definition.id,
            label = assert(locale.text(button_definition.label)),
            x = minipanel.button_x + (index - 1) * minipanel.button_step,
            y = minipanel.button_y,
            width = minipanel.button_width,
            height = minipanel.button_height,
            enabled = button_definition.enabled,
            visible = false,
            on_activate = function()
                if button_definition.scene then
                    scenes.push(button_definition.scene)
                end
            end,
        })
        control.node = node
    end

    local menu = definition.menu
    local menu_node = dc6_at(hud.root, menu.sheet, palette, menu.closed_frame, menu.x, menu.y)
    hud.controls:add({
        id = "minipanel_toggle",
        label = "Open/Close Mini-panel",
        x = menu.x,
        y = menu.y,
        width = menu.width,
        height = menu.height,
        on_activate = function()
            hud.menu_open = not hud.menu_open
            menu_node:set_dc6(menu.sheet, palette, 0, hud.menu_open and menu.open_frame or menu.closed_frame)
            minipanel_node:set_visible(hud.menu_open)
            for _, button in ipairs(minipanel.buttons) do
                local control = hud.controls:get("minipanel_" .. button.id)
                control.visible = hud.menu_open
                control.node:set_visible(hud.menu_open)
            end
        end,
    })
    M.snapshot(hud, nil)
    return hud
end

local function update_bar(bar, fill)
    local pixels = math.floor(bar.definition.width * fill + 0.5)
    if pixels == bar.pixels then return end
    bar.pixels = pixels
    local color = bar.definition.color
    bar.node:fill_rect(math.max(1, pixels), bar.definition.height, color.red, color.green, color.blue, color.alpha)
    bar.node:set_visible(pixels > 0)
    bar.node:set_position(bar.definition.x + pixels / 2, bar.definition.y + bar.definition.height / 2)
end

local function update_skill(hud, side, detail, skill_id)
    local well = hud.skills[side]
    if well.id == skill_id then return end
    well.id = skill_id
    detail = detail or {
        id = skill_id,
        icon = hud.definition.skills.frame,
        sheet = hud.definition.skills.sheet,
    }
    well.node:set_dc6(detail.sheet, hud.palette, 0, detail.icon)
    local name = detail.name_key and locale.text(detail.name_key) or nil
    local short = detail.short_key and locale.text(detail.short_key) or nil
    name = name or string.format("%s SKILL %d", string.upper(side), skill_id)
    well.tip:set_text(short and short ~= name and name .. "\n" .. short or name)
end

function M.snapshot(hud, stats)
    stats = stats or {}
    local health, max_health = stats.health or 0, stats.max_health or 0
    local mana, max_mana = stats.mana or 0, stats.max_mana or 0
    local stamina, max_stamina = stats.stamina or 0, stats.max_stamina or 0
    local experience, next_experience = stats.experience or 0, stats.next_level_experience or 0
    local running = stats.running == true
    update_skill(hud, "left", stats.left_skill_detail, stats.left_skill or 0)
    update_skill(hud, "right", stats.right_skill_detail, stats.right_skill or 0)
    if hud.running ~= running then
        hud.running = running
        hud.run_node:set_dc6(hud.definition.run.sheet, hud.palette, 0, running and hud.definition.run.run_frame or hud.definition.run.walk_frame)
        hud.run_tip:set_text(assert(locale.text(running and "darkmagic.hud.run" or "darkmagic.hud.walk")))
    end
    local health_fill, mana_fill = ratio(health, max_health), ratio(mana, max_mana)
    if hud.health_fill ~= health_fill then
        hud.health_fill = health_fill
        clip_globe(hud.health_liquid, hud.definition.globes.health, health_fill)
    end
    if hud.mana_fill ~= mana_fill then
        hud.mana_fill = mana_fill
        clip_globe(hud.mana_liquid, hud.definition.globes.mana, mana_fill)
    end
    update_bar(hud.bars.stamina, ratio(stamina, max_stamina))
    update_bar(hud.bars.experience, ratio(experience, next_experience))
    hud.tips.health:set_text(string.format("%s: %d / %d", assert(locale.text("darkmagic.hud.health")), health, max_health))
    hud.tips.mana:set_text(string.format("%s: %d / %d", assert(locale.text("darkmagic.hud.mana")), mana, max_mana))
    hud.tips.stamina:set_text(string.format("%s: %d / %d", assert(locale.text("darkmagic.hud.stamina")), stamina, max_stamina))
    hud.tips.experience:set_text(string.format("%s: %d / %d", assert(locale.text("darkmagic.hud.experience")), experience, next_experience))
end

function M.update(hud, stats)
    M.snapshot(hud, stats)
    hud.controls:update()
end

return M
