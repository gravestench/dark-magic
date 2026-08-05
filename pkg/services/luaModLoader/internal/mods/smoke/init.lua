local Smoke = {
    initialized = false,
}

function Smoke:Init()
    self.initialized = true
    self.log("embedded smoke mod initialized")
end

return Smoke
