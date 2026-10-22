-- Apply record-authored state tint to retained actor render nodes.
local M = {}

function M.apply(actors, snapshots, resolve, level_id)
    local green = {}
    for _, snapshot in ipairs(snapshots) do
        if snapshot.level_id == level_id then
            local state = resolve(snapshot.state_id)
            if state and state.green then
                green[tostring(snapshot.target_entity_id)] = true
            end
        end
    end
    for key, node in pairs(actors) do
        if green[key] then
            node:set_tint(128, 255, 128)
        else
            node:set_tint(255, 255, 255)
        end
    end
end

return M
