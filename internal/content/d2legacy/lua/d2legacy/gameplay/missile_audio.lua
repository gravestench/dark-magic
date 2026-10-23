local audio = require("d2legacy.audio")
local M = {}

function M.play(voices, cue)
    local sound, loop = audio.play_record(cue.sound, cue.tick or 0)
    if not loop then
        return
    end
    local owner = cue.sound_owner_id
    if not owner or owner == "" then
        -- Old checkpoints and peers cannot give a travel loop a bounded owner.
        sound:stop()
        return
    end
    if voices[owner] then
        voices[owner]:stop()
    end
    voices[owner] = sound
end

function M.reconcile(voices, snapshots, level_id)
    if not voices or not next(voices) then
        return
    end
    local live = {}
    for _, snapshot in ipairs(snapshots) do
        if snapshot.level_id == level_id and snapshot.sound_owner_id then
            live[snapshot.sound_owner_id] = true
        end
    end
    for owner, sound in pairs(voices) do
        if not live[owner] then
            sound:stop()
            voices[owner] = nil
        end
    end
end

function M.close(voices)
    for owner, sound in pairs(voices or {}) do
        sound:stop()
        voices[owner] = nil
    end
end

return M
