-- Engine-facing component contracts used by the first gameplay slices.
-- These are facts, not rules. Declaring them here lets a renderer-free server
-- boot d2legacy without relying on the client world builder to run first.

local ecs = require("engine.ecs/v1")
local M = {}

local function component(name, fields) ecs.component({name=name,fields=fields}) end

function M.register()
    component("d2legacy.player.identity", {
        {name="character_id",type="string"},{name="player",type="string"},
        {name="name",type="string"},{name="class",type="string"},
    })
    component("d2legacy.player.vitals", {
        {name="health",type="i64"},{name="max_health",type="i64"},
        {name="mana",type="i64"},{name="max_mana",type="i64"},
        {name="mana_raw",type="i64"},{name="max_mana_raw",type="i64"},
    })
    component("d2legacy.player.learned_skill", {
        {name="owner",type="entity"},{name="skill_id",type="i64"},
        {name="level",type="i64"},{name="list_row",type="i64"},
        {name="left_allowed",type="bool"},{name="right_allowed",type="bool"},
    })
    component("d2legacy.player.skill_assignment", {
        {name="left",type="i64"},{name="right",type="i64"},
    })
    component("d2legacy.world.position", {{name="x",type="f64"},{name="y",type="f64"}})
    component("d2legacy.world.velocity", {{name="x",type="f64"},{name="y",type="f64"}})
    component("d2legacy.world.player_control", {{name="player",type="string"}})
    component("d2legacy.player.animation", {{name="direction",type="i64"},{name="mode",type="string"}})
    component("d2legacy.world.location", {{name="act",type="i64"},{name="level_id",type="i64"}})
    component("d2legacy.world.collider", {{name="radius",type="f64"}})
    component("d2legacy.world.selectable", {
        {name="id",type="string"},{name="kind",type="string"},{name="label",type="string"},
        {name="owner",type="string"},{name="radius",type="f64"},{name="priority",type="i64"},
    })
    component("d2legacy.monster.stats", {
        {name="level",type="i64"},{name="health",type="i64"},{name="max_health",type="i64"},
        {name="defense",type="i64"},{name="attack_rating",type="i64"},
        {name="physical_min",type="i64"},{name="physical_max",type="i64"},
        {name="experience",type="i64"},
    })
    component("d2legacy.monster.ai", {
        {name="behavior",type="string"},{name="state",type="string"},
        {name="target_id",type="string"},{name="next_think_tick",type="i64"},
        {name="think_interval",type="i64"},{name="aggro_radius",type="f64"},
        {name="attack_range",type="f64"},{name="speed",type="f64"},
    })
    component("d2legacy.player.progress", {
        {name="level",type="i64"},{name="experience",type="i64"},
    })
end

return M
