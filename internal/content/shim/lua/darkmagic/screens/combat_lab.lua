-- Combat Lab drives the same authoritative melee transaction as gameplay.
-- This file draws and observes; native combat decides hit, damage, and death.

-- Loaded only when this scene is opened. Frontend-only hosts deliberately do
-- not install gameplay capabilities, but must still register every scene.
local ecs
local input = require("dm.input/v1")
local render = require("dm.render/v1")
local fuzzy_picker = require("darkmagic.ui.fuzzy_picker")
local text = require("darkmagic.ui.text")

local lab = {}
local fixed_one = 256
local attacker_id, target_id = "combat-lab:attacker", "combat-lab:target"
local presets = {
    {name="UNARMED 1-2 VS 10 HP", minimum=1, maximum=2, health=10},
    {name="HEAVY 4-8 VS 30 HP", minimum=4, maximum=8, health=30},
    {name="LETHAL 10 VS 5 HP", minimum=10, maximum=10, health=5},
}

local function field(name, kind)
    return {name=name, type=kind}
end

local function label(root, value, x, y, width, align, style)
    local node = render.create("hud", root)
    text.set(node, style or "font_lab_caption", value, width or 760, align or "center")
    node:set_position(x or 400, y)
    return node
end

local function register_schemas()
    -- Identical registration is harmless and documents every borrowed field.
    ecs.component({name="dm.world.selectable", version=1, fields={
        field("id", "string"), field("kind", "string"),
        field("label", "string"), field("owner", "string"),
        field("radius", "f64"), field("priority", "i64"),
    }})
    ecs.component({name="dm.world.position", version=1, fields={
        field("x", "f64"), field("y", "f64"),
    }})
    ecs.component({name="dm.world.location", version=1, fields={
        field("act", "i64"), field("level_id", "i64"),
    }})
    ecs.component({name="dm.combat.melee_profile", version=1, fields={
        field("range", "f64"), field("physical_min", "i64"),
        field("physical_max", "i64"),
    }})
    ecs.component({name="dm.combat.basic_attack_request", version=1, fields={
        field("target_id", "string"), field("request_tick", "i64"),
    }})
    ecs.component({name="dm.monster.stats", version=1, fields={
        field("level", "i64"), field("health", "i64"),
        field("max_health", "i64"), field("defense", "i64"),
        field("attack_rating", "i64"), field("physical_min", "i64"),
        field("physical_max", "i64"), field("experience", "i64"),
    }})
end

local function selectable(id, kind, name)
    return {id=id,kind=kind,label=name,owner="combat-lab",radius=1,priority=1}
end

function lab:create_fixture(preset)
    self.attacker = ecs.create({
        ["dm.world.selectable"]=selectable(attacker_id,"player","LAB HERO"),
        ["dm.world.position"]={x=0,y=0}, ["dm.world.location"]={act=1,level_id=1},
        ["dm.combat.melee_profile"]={range=2,physical_min=preset.minimum*fixed_one,physical_max=preset.maximum*fixed_one},
    })
    self.target = ecs.create({
        ["dm.world.selectable"]=selectable(target_id,"hostile","LAB HOSTILE"),
        ["dm.world.position"]={x=1,y=0}, ["dm.world.location"]={act=1,level_id=1},
        ["dm.monster.stats"]={level=1,health=preset.health*fixed_one,max_health=preset.health*fixed_one,defense=0,attack_rating=0,physical_min=0,physical_max=0,experience=0},
    })
end

function lab:destroy_fixture()
    if self.attacker and self.attacker:exists() then ecs.destroy(self.attacker) end
    if self.target and self.target:exists() then ecs.destroy(self.target) end
    self.attacker, self.target = nil, nil
end

function lab:clear_events()
    for _, entity in ipairs(ecs.query({all={"dm.combat.event"}})) do
        local event = ecs.get(entity,"dm.combat.event")
        if event and event:get("attacker_id") == attacker_id then ecs.destroy(entity) end
    end
    self.seen, self.lines = {}, {}
end

function lab:reset(index)
    self:destroy_fixture(); self:clear_events()
    self.preset_index = index or self.preset_index or 1
    self:create_fixture(presets[self.preset_index]); self:refresh()
end

function lab:attack()
    local stats = self.target and self.target:exists() and ecs.get(self.target,"dm.monster.stats") or nil
    if not stats or stats:get("health") <= 0 then self:reset(); return end
    -- Tick zero means ready now. Combat consumes this on the next fixed tick.
    ecs.set(self.attacker,"dm.combat.basic_attack_request",{target_id=target_id,request_tick=0})
end

local function event_line(event)
    local kind, tick = event:get("kind"), event:get("tick")
    if kind == "hit_resolved" then return string.format("[white]%d[/]  HIT RESOLVED  %s",tick,event:get("hit") and "[green]HIT[/]" or "[red]MISS[/]") end
    if kind == "damage_applied" then return string.format("[white]%d[/]  [gold]%g PHYSICAL[/]  [blue]%g HP LEFT[/]",tick,event:get("damage")/fixed_one,event:get("remaining_health")/fixed_one) end
    if kind == "unit_died" then return string.format("[white]%d[/]  [red]UNIT DIED[/]",tick) end
    return string.format("[white]%d[/]  %s",tick,kind:upper())
end

function lab:observe_events()
    for _, entity in ipairs(ecs.query({all={"dm.combat.event"}})) do
        local key, event = tostring(entity:id()), ecs.get(entity,"dm.combat.event")
        if event and event:get("attacker_id") == attacker_id and not self.seen[key] then
            self.seen[key] = true; self.lines[#self.lines+1] = event_line(event)
            while #self.lines > 8 do table.remove(self.lines,1) end
        end
    end
end

function lab:refresh()
    local preset, health = presets[self.preset_index], 0
    if self.target and self.target:exists() then health=ecs.get(self.target,"dm.monster.stats"):get("health")/fixed_one end
    local summary = string.format(
        "[gold]%s[/]   DAMAGE [white]%d-%d[/]   TARGET [green]%g/%d HP[/]",
        preset.name, preset.minimum, preset.maximum, health, preset.health)
    text.set(self.summary, "font_lab_color", summary, 760, "center")
    for index, node in ipairs(self.event_nodes) do
        local line_index = #self.lines - #self.event_nodes + index
        text.set(node, "font_lab_color", self.lines[line_index] or "", 680, "left")
    end
end

function lab:create_labels()
    label(self.root,"AUTHORITATIVE COMBAT LAB",400,36,760,"center","font_lab_heading")
    self.summary=label(self.root,"",400,92,760)
    label(self.root,"ATTACKER",220,180,240,"center","font_lab_heading")
    label(self.root,"LAB HERO",220,225,240)
    label(self.root,"TARGET",580,180,240,"center","font_lab_heading")
    label(self.root,"LAB HOSTILE",580,225,240)
    label(self.root,"SEMANTIC EVENT STREAM",400,300,680,"center","font_lab_heading")
    self.event_nodes={}
    for index=1,8 do
        self.event_nodes[index]=label(self.root,"",400,330+index*24,680,"left")
    end
    label(self.root,"Enter: attack/reset dead target   R: reset   F: choose fixture",400,570,760)
end

function lab:create_picker()
    local names={}; for _,preset in ipairs(presets) do names[#names+1]=preset.name end
    self.picker=fuzzy_picker.create(self.root,{title="SELECT COMBAT FIXTURE",items=names,on_select=function(value)
        for index,preset in ipairs(presets) do if preset.name==value then self:reset(index); return end end
    end})
end

function lab:create()
    ecs = require("dm.ecs/v1")
    register_schemas()
    self.root=render.create("hud")
    self.seen,self.lines,self.preset_index={},{},1
    self:create_labels()
    self:create_picker()
    self:create_fixture(presets[1]); self:refresh()
end

function lab:update()
    if self.picker:update() then return end
    if input.pressed("search") then self.picker:show(); return end
    if input.pressed("confirm") then self:attack() end
    if input.pressed("toggle_run") then self:reset() end
    self:observe_events(); self:refresh()
end

function lab:destroy() self:destroy_fixture(); self:clear_events() end

return lab
