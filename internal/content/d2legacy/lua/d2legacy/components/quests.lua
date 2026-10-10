-- Durable quest, reward, and difficulty facts owned by d2legacy.

local ecs=require("engine.ecs/v1")
local M={}

function M.register()
    ecs.component({name="d2legacy.quest.progress",fields={
        {name="quest_id",type="i64"},{name="stage",type="i64"},
        {name="completed",type="bool"},{name="rewarded",type="bool"},
    }})
    ecs.component({name="d2legacy.quest.event",fields={
        {name="kind",type="string"},{name="tick",type="i64"},
        {name="player",type="string"},{name="quest_id",type="i64"},
        {name="npc",type="string"},{name="reward",type="string"},
    }})
    -- Character progression only. The game's difficulty is read from the
    -- immutable game rules, never copied onto players.
    ecs.component({name="d2legacy.player.difficulty",fields={
        {name="highest_completed",type="i64"},
    }})
end

return M
