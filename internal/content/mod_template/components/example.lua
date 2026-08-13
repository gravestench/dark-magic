-- Optional independently managed component.
-- Delete this file when the new mod needs only its boot component.

return {
    id = "mod_template.example",
    api = 1,

    start = function(self)
        self.started = true
    end,

    stop = function(self)
        self.started = false
    end,
}
