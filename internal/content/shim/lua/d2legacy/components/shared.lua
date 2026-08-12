-- Engine-facing component contracts used by the first gameplay slices.
-- These are facts, not rules. Declaring them here lets a renderer-free server
-- boot d2legacy without relying on the client world builder to run first.

local ecs = require("dm.ecs/v1")
local M = {}

local function component(name, fields) ecs.component({name=name,fields=fields}) end

function M.register()
    component("dm.player.identity", {
        {name="character_id",type="string"},{name="player",type="string"},
        {name="name",type="string"},{name="class",type="string"},
    })
    component("dm.player.vitals", {
        {name="health",type="i64"},{name="max_health",type="i64"},
        {name="mana",type="i64"},{name="max_mana",type="i64"},
        {name="mana_raw",type="i64"},{name="max_mana_raw",type="i64"},
    })
    component("dm.player.learned_skill", {
        {name="owner",type="entity"},{name="skill_id",type="i64"},
        {name="level",type="i64"},{name="list_row",type="i64"},
        {name="left_allowed",type="bool"},{name="right_allowed",type="bool"},
    })
    component("dm.player.skill_assignment", {
        {name="left",type="i64"},{name="right",type="i64"},
    })
    component("dm.world.position", {{name="x",type="f64"},{name="y",type="f64"}})
    component("dm.world.location", {{name="act",type="i64"},{name="level_id",type="i64"}})
    component("dm.world.collider", {{name="radius",type="f64"}})
    component("dm.world.selectable", {
        {name="id",type="string"},{name="kind",type="string"},{name="label",type="string"},
        {name="owner",type="string"},{name="radius",type="f64"},{name="priority",type="i64"},
    })
    component("dm.monster.stats", {
        {name="level",type="i64"},{name="health",type="i64"},{name="max_health",type="i64"},
        {name="defense",type="i64"},{name="attack_rating",type="i64"},
        {name="physical_min",type="i64"},{name="physical_max",type="i64"},
        {name="experience",type="i64"},
    })
end

return M
