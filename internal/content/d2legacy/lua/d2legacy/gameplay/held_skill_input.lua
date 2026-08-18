-- Convert rendered pointer samples into one skill request per authoritative
-- action. A physical click commonly remains down for several presentation
-- frames; those frames must not become a queued second cast. Holding still
-- repeats after authority reports the preceding action complete.

local M = {}

function M.update(state, pressed, down, action_active)
    state = state or {}
    if not down then
        state.submitted = false
        state.observed_active = false
        return false, state
    end
    if action_active then
        state.observed_active = true
    elseif state.observed_active then
        state.observed_active = false
        state.submitted = false
    end
    return pressed or (not action_active and not state.submitted), state
end

function M.submitted(state)
    state.submitted = true
end

return M
