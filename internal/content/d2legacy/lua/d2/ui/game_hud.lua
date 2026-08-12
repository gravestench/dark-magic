-- Authored in-game HUD composition for the supported 800x600 profile.
--
-- This is another capstone file. The HUD does NOT own health, mana, learned
-- skills, item containers, or movement mode. It receives VALUE snapshots and a
-- small table of command functions from game_world.lua, then turns those facts
-- into retained presentation and player intent.
--
-- A useful mental model:
--
--   authoritative snapshot ---> HUD pictures/tooltips
--   HUD click -------------> command function ---> authority
--
-- Never reverse those arrows by treating a render node as gameplay state.

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local locale = require("engine.locale/v1")
local controls = require("d2.ui.controls")
local tooltip = require("d2.ui.tooltip")
local skill_selector = require("d2.ui.skill_selector")

local M = {}

-- Search a copied item snapshot for the authoritative held container.
local function held_item(snapshot)
    if snapshot == nil then return nil end
    for _, item in ipairs(snapshot.items) do
        if item.container == "held" then return item end
    end
end

local function belt_item(snapshot, slot)
    if snapshot == nil then return nil end
    for _, item in ipairs(snapshot.items) do
        if item.container == "belt" and item.belt_slot == slot then return item end
    end
end

-- Convert one belt-cell click into item MOVE INTENT.
local function activate_belt_slot(hud, slot)
    local held = held_item(hud.item_snapshot)

    if held ~= nil then
        -- Ask authority to place held item in belt slot; `true` enables legal
        -- atomic swap behavior without implementing that swap in presentation.
        hud.commands.move_item(held.id, { container = "belt", belt_slot = slot }, true)
        return
    end

    local item = belt_item(hud.item_snapshot, slot)
    if item ~= nil then hud.commands.move_item(item.id, { container = "held" }) end
end

local function refresh_hud_items(hud)
    -- HUD can be created without item commands in presentation-only contexts.
    if hud.commands.item_snapshot == nil then return end

    local snapshot = assert(hud.commands.item_snapshot())
    local belt = hud.definition.belt
    local cursor_x, cursor_y = input.cursor()

    hud.item_snapshot = snapshot
    hud.item_held = held_item(snapshot) ~= nil

    for _, item in ipairs(snapshot.items) do
        -- One item can need separate retained nodes for separate visual roles.
        -- Reuse a stable little drawing record by authoritative item ID.
        local drawing = hud.item_nodes[item.id] or {}
        hud.item_nodes[item.id] = drawing

        if item.inventory_dc6 ~= "" and item.container == "belt" and drawing.belt_node == nil then
            drawing.belt_node = render.create("hud", hud.root)
            drawing.width, drawing.height = drawing.belt_node:set_dc6(item.inventory_dc6, hud.item_palette, 0, 0)
        end

        if item.inventory_dc6 ~= "" and item.container == "held" and drawing.held_node == nil then
            -- Held item is cursor-layer presentation so it stays above normal HUD.
            drawing.held_node = render.create("cursor")
            drawing.width, drawing.height = drawing.held_node:set_dc6(item.inventory_dc6, hud.item_palette, 0, 0)
            drawing.held_node:set_z(999)
        end

        if drawing.belt_node ~= nil then drawing.belt_node:set_visible(item.container == "belt") end
        if drawing.held_node ~= nil then drawing.held_node:set_visible(item.container == "held") end

        if drawing.width ~= nil then
            if item.container == "belt" then
                -- Belt slot is a single zero-based number. `% columns` gives its
                -- column; integer division via floor gives its row.
                local column = item.belt_slot % belt.columns
                local row = math.floor(item.belt_slot / belt.columns)

                drawing.belt_node:set_position(
                    belt.x + column * belt.cell_width + belt.cell_width / 2,
                    -- Extra belt rows open UPWARD, so row increases by subtracting Y.
                    belt.y - row * belt.cell_height + belt.cell_height / 2
                )
            elseif item.container == "held" then
                drawing.held_node:set_position(cursor_x, cursor_y)
            end
        end
    end
end

-- Tiny top-left -> retained-center helper for one DC6 frame.
local function dc6_at(root, sheet, palette, frame, x, y)
    local node = render.create("hud", root)
    local width, height = node:set_dc6(sheet, palette, 0, frame)
    node:set_position(x + width / 2, y + height / 2)
    return node, width, height
end

-- Some HUD pieces are authored relative to a shared BOTTOM edge instead of top.
local function dc6_bottom(root, sheet, palette, frame, x, bottom)
    local node = render.create("hud", root)
    local width, height = node:set_dc6(sheet, palette, 0, frame)
    node:set_position(x + width / 2, bottom - height / 2)
    return node
end

-- A globe is liquid art plus an overlap/frame decoration drawn above it.
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
    -- Return only liquid because that is the part later clipped by current value.
    return liquid
end

local function ratio(value, maximum)
    -- Defensive presentation conversion: malformed/missing/zero max displays empty.
    if type(value) ~= "number" or type(maximum) ~= "number" or maximum <= 0 then return 0 end
    return math.max(0, math.min(1, value / maximum))
end

local function clip_globe(node, definition, fill)
    -- Convert normalized fill into a whole pixel height, rounding to nearest pixel.
    local visible_height = math.floor(definition.height * fill + 0.5)
    node:set_visible(visible_height > 0)

    -- Liquid empties from TOP downward: clip rectangle begins at bottom minus the
    -- visible amount. The underlying liquid texture itself does not need rebuilding.
    node:set_clip(
        definition.x,
        definition.y + definition.height - visible_height,
        definition.width,
        math.max(1, visible_height)
    )
end

-- Status regions (globes/bars) are not keyboard focus targets, but they still
-- become controls so pointer hover can drive accessible tooltips consistently.
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

    -- The HUD owns only its presentation/control state and the function table
    -- supplied by its caller. It does not store a mutable player object.
    local hud = { root = render.create("hud", root), controls = controls.new(), running = false, menu_open = false, definition = definition, commands = commands, item_nodes = {} }
    local palette = palettes[definition.palette]
    hud.item_palette = palettes.units

    -- Compose bottom HUD bar from authored pieces aligned to common bottom edges.
    for _, part in ipairs(definition.panel_parts) do
        dc6_bottom(hud.root, definition.panel_sheet, palette, part.frame, part.x, part.bottom)
    end

    local globes = definition.globes
    hud.health_liquid = add_globe(hud.root, globes.health, globes.sheet, globes.overlap_sheet, palette)
    hud.mana_liquid = add_globe(hud.root, globes.mana, globes.sheet, globes.overlap_sheet, palette)

    -- Stamina/experience use generated rectangular fills rather than clipping globe art.
    hud.bars = {}
    for name, bar in pairs({ stamina = definition.stamina, experience = definition.experience }) do
        local node = render.create("hud", hud.root)
        -- pixels=-1 guarantees the first update_bar call cannot think nothing changed.
        hud.bars[name] = { node = node, definition = bar, pixels = -1 }
    end

    -- Tooltips start empty and get real values in M.snapshot below.
    hud.tips = {
        health = tooltip.create(hud.root, "", globes.health.x + globes.health.width / 2, globes.health.y, {}),
        mana = tooltip.create(hud.root, "", globes.mana.x + globes.mana.width / 2, globes.mana.y, {}),
        stamina = tooltip.create(hud.root, "", definition.stamina.x + definition.stamina.width / 2, definition.stamina.y, {}),
        experience = tooltip.create(hud.root, "", definition.experience.x + definition.experience.width / 2, definition.experience.y, {}),
    }

    add_status_control(hud, "health", globes.health, assert(locale.text("d2.hud.health")), hud.tips.health)
    add_status_control(hud, "mana", globes.mana, assert(locale.text("d2.hud.mana")), hud.tips.mana)
    add_status_control(hud, "stamina", definition.stamina, assert(locale.text("d2.hud.stamina")), hud.tips.stamina)
    add_status_control(hud, "experience", definition.experience, assert(locale.text("d2.hud.experience")), hud.tips.experience)

    -- BELT ------------------------------------------------------------------
    local belt = definition.belt
    hud.belt = { rows = {}, expanded = false, hovered = false, capacity = 4 }

    -- Bottom row is part of regular HUD; create additional rows 2..N hidden above it.
    for row = 2, belt.rows do
        local node = dc6_at(hud.root, belt.sheet, palette, 0, belt.x - 1, belt.y - (row - 1) * belt.cell_height)
        node:set_visible(false)
        hud.belt.rows[row] = node
    end

    local belt_control
    local function refresh_belt()
        -- Belt visually expands only while hovered and only if authority says
        -- capacity exceeds the always-visible bottom-row column count.
        local expanded = hud.belt.hovered == true and hud.belt.capacity > belt.columns
        local visible_rows = math.ceil(hud.belt.capacity / belt.columns)
        hud.belt.expanded = expanded

        for row = 2, belt.rows do
            hud.belt.rows[row]:set_visible(expanded and row <= visible_rows)
        end

        -- IMPORTANT: the control's hit rectangle changes with visual expansion.
        -- This is plain Lua data; controls.Manager reads it on the next hit test.
        belt_control.y = expanded and belt.y - (visible_rows - 1) * belt.cell_height or belt.y
        belt_control.height = expanded and visible_rows * belt.cell_height or belt.cell_height
    end

    belt_control = hud.controls:add({
        id = "belt",
        label = "Belt",
        x = belt.x,
        y = belt.y,
        width = belt.columns * belt.cell_width,
        height = belt.cell_height,
        focusable = false,
        on_state = function(_, state)
            hud.belt.hovered = state == "hover" or state == "pressed"
            refresh_belt()
        end,
    })

    -- Expose refresh so snapshots can update capacity and immediately recompute rows.
    hud.belt.refresh = refresh_belt

    if commands.item_snapshot and commands.move_item then
        -- The equipped belt currently exposes four bottom-row interactive cells.
        -- Authority already supports larger capacity; future expanded-cell
        -- controls can grow from exactly the same command/snapshot pattern.
        for slot = 0, belt.columns - 1 do
            -- Per-iteration local makes activation closure remember this slot.
            local belt_slot = slot
            hud.controls:add({
                id = "belt_" .. slot,
                label = "Belt slot " .. (slot + 1),
                x = belt.x + slot * belt.cell_width,
                y = belt.y,
                width = belt.cell_width,
                height = belt.cell_height,
                focusable = false,
                on_activate = function() activate_belt_slot(hud, belt_slot) end,
            })
        end
        refresh_hud_items(hud)
    end

    -- SKILL WELLS ------------------------------------------------------------
    local skills = definition.skills
    hud.skills = {}

    -- Assignment function is required because the selector must submit intent;
    -- it should never directly rewrite a local `selected skill` variable as authority.
    hud.skill_selector = skill_selector.create(hud.root, skills, palette, hud.controls, assert(commands.assign_skill))

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
            on_activate = function() skill_selector.toggle(hud.skill_selector, side) end,
            on_state = function(_, state) tip:set_visible(state == "hover") end,
        })
    end

    -- RUN/WALK ---------------------------------------------------------------
    local run = definition.run
    local run_node = dc6_at(hud.root, run.sheet, palette, run.walk_frame, run.x, run.y)
    local run_tip = tooltip.create(hud.root, assert(locale.text("d2.hud.walk")), run.x + run.width / 2, run.y, {})
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
            -- Submit desired next running state. The visual updates only after a
            -- later authoritative snapshot reports the new state.
            if commands.request_running then commands.request_running(not hud.running) end
        end,
        on_state = function(_, state) run_tip:set_visible(state == "hover") end,
    })

    -- MINI-PANEL -------------------------------------------------------------
    local minipanel = definition.minipanel
    local minipanel_node = dc6_at(hud.root, minipanel.sheet, palette, 0, minipanel.x, minipanel.y)
    minipanel_node:set_visible(false)

    for index, button in ipairs(minipanel.buttons) do
        -- Local alias prevents the loop's meaning from getting lost inside callbacks.
        local button_definition = button
        local button_x = minipanel.button_x + (index - 1) * minipanel.button_step
        local label = assert(locale.text(button_definition.label))

        local node = dc6_at(
            hud.root,
            minipanel.button_sheet,
            palette,
            button_definition.frame,
            button_x,
            minipanel.button_y
        )
        node:set_visible(false)

        local tip = tooltip.create(hud.root, label, button_x + minipanel.button_width / 2, minipanel.button_y, {})

        local control = hud.controls:add({
            id = "minipanel_" .. button_definition.id,
            label = label,
            x = button_x,
            y = minipanel.button_y,
            width = minipanel.button_width,
            height = minipanel.button_height,
            enabled = button_definition.enabled,
            visible = false,

            on_activate = function()
                if button_definition.scene then
                    scenes.toggle_overlay(button_definition.scene, assert(button_definition.slot))
                end
            end,

            on_state = function(_, state)
                -- Pressed art is authored in next adjacent frame.
                node:set_dc6(minipanel.button_sheet, palette, 0, state == "pressed" and button_definition.frame + 1 or button_definition.frame)
                tip:set_visible(state == "hover" or state == "focused" or state == "pressed")
            end,
        })

        -- Store visuals directly on control so menu toggle can reveal/hide them.
        control.node = node
        control.tip = tip
    end

    local menu = definition.menu
    local menu_node = dc6_at(hud.root, menu.sheet, palette, menu.closed_frame, menu.x, menu.y)
    local menu_tip = tooltip.create(hud.root, assert(locale.text("d2.minipanel.open")), menu.x + menu.width / 2, menu.y, {})

    local function update_menu_frame(pressed)
        local frame = hud.menu_open and menu.open_frame or menu.closed_frame
        menu_node:set_dc6(menu.sheet, palette, 0, pressed and frame + 1 or frame)
    end

    hud.controls:add({
        id = "minipanel_toggle",
        label = "Open/Close Mini-panel",
        x = menu.x,
        y = menu.y,
        width = menu.width,
        height = menu.height,

        on_activate = function()
            hud.menu_open = not hud.menu_open
            update_menu_frame(false)

            -- Tooltip wording follows resulting state.
            menu_tip:set_text(assert(locale.text(hud.menu_open and "d2.minipanel.close" or "d2.minipanel.open")))
            minipanel_node:set_visible(hud.menu_open)

            for _, button in ipairs(minipanel.buttons) do
                local control = hud.controls:get("minipanel_" .. button.id)
                -- Keep input visibility and render visibility synchronized.
                control.visible = hud.menu_open
                control.node:set_visible(hud.menu_open)
                if not hud.menu_open then control.tip:set_visible(false) end
            end
        end,

        on_state = function(_, state)
            update_menu_frame(state == "pressed")
            menu_tip:set_visible(state == "hover" or state == "focused" or state == "pressed")
        end,
    })

    -- Initialize all value-driven visuals with an empty snapshot.
    M.snapshot(hud, nil)
    return hud
end

local function update_bar(bar, fill)
    local pixels = math.floor(bar.definition.width * fill + 0.5)

    -- Retained optimization: do nothing when the visible integer width is unchanged.
    if pixels == bar.pixels then return end
    bar.pixels = pixels

    local color = bar.definition.color
    bar.node:fill_rect(math.max(1, pixels), bar.definition.height, color.red, color.green, color.blue, color.alpha)
    bar.node:set_visible(pixels > 0)

    -- Bar grows from left edge, so center shifts with current pixel width.
    bar.node:set_position(bar.definition.x + pixels / 2, bar.definition.y + bar.definition.height / 2)
end

local function update_skill(hud, side, detail, skill_id)
    local well = hud.skills[side]

    -- Skill art/tooltip only needs rebuilding when assigned ID changes.
    if well.id == skill_id then return end
    well.id = skill_id

    -- Fallback display record when immutable game-data detail is unavailable.
    detail = detail or {
        id = skill_id,
        icon = hud.definition.skills.frame,
        sheet = hud.definition.skills.sheet,
    }

    well.node:set_dc6(detail.sheet, hud.palette, 0, detail.icon)

    local name = detail.name_key and locale.text(detail.name_key) or nil
    local short = detail.short_key and locale.text(detail.short_key) or nil
    name = name or string.format("%s SKILL %d", string.upper(side), skill_id)

    -- If short description exists and differs from name, show two lines.
    well.tip:set_text(short and short ~= name and name .. "\n" .. short or name)
end

-- Apply one value-only gameplay snapshot to retained HUD presentation.
function M.snapshot(hud, stats)
    stats = stats or {}

    local health, max_health = stats.health or 0, stats.max_health or 0
    local mana, max_mana = stats.mana or 0, stats.max_mana or 0
    local stamina, max_stamina = stats.stamina or 0, stats.max_stamina or 0
    local experience, next_experience = stats.experience or 0, stats.next_level_experience or 0
    local running = stats.running == true

    -- Belt authority supports 4..16 capacity; clamp presentation into supported range.
    hud.belt.capacity = math.max(4, math.min(16, stats.belt_capacity or 4))
    hud.belt.refresh()

    update_skill(hud, "left", stats.left_skill_detail, stats.left_skill or 0)
    update_skill(hud, "right", stats.right_skill_detail, stats.right_skill or 0)
    skill_selector.set_skills(hud.skill_selector, stats.learned_skills or {})

    if hud.running ~= running then
        hud.running = running

        -- Visual state follows authoritative snapshot, not the click that requested it.
        hud.run_node:set_dc6(hud.definition.run.sheet, hud.palette, 0, running and hud.definition.run.run_frame or hud.definition.run.walk_frame)
        hud.run_tip:set_text(assert(locale.text(running and "d2.hud.run" or "d2.hud.walk")))
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

    -- Tooltips are presentation strings derived from snapshot values.
    hud.tips.health:set_text(string.format("%s: %d / %d", assert(locale.text("d2.hud.health")), health, max_health))
    hud.tips.mana:set_text(string.format("%s: %d / %d", assert(locale.text("d2.hud.mana")), mana, max_mana))
    hud.tips.stamina:set_text(string.format("%s: %d / %d", assert(locale.text("d2.hud.stamina")), stamina, max_stamina))
    hud.tips.experience:set_text(string.format("%s: %d / %d", assert(locale.text("d2.hud.experience")), experience, next_experience))
end

function M.update(hud, stats)
    -- Update order: apply gameplay values, refresh item snapshot/art, then dispatch
    -- current-frame UI input against the resulting presentation.
    M.snapshot(hud, stats)
    refresh_hud_items(hud)
    hud.controls:update()
end

-- World scene calls this when overlay routing removes pointer ownership. Held
-- item authority does not change; only this particular cursor-layer picture hides.
function M.set_item_cursor_visible(hud, visible)
    local held = held_item(hud.item_snapshot)

    for id, drawing in pairs(hud.item_nodes) do
        if drawing.held_node ~= nil then
            drawing.held_node:set_visible(visible == true and held ~= nil and held.id == id)
        end
    end
end

return M