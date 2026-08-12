-- First-party Diablo II gameplay composition root.
--
-- This file deliberately contains no formulas or mutations. Its only job is to
-- show which small modules make up the mod and register them in readable order.

local components = require("d2legacy.components.fire_bolt")
local fire_bolt_data = require("d2legacy.data.fire_bolt")
local cast_command = require("d2legacy.commands.cast")
local cast_system = require("d2legacy.systems.cast")

local M = {
    id = "d2legacy.authoritative",
}

function M.start()
    components.register()

    -- Record interpretation happens once during composition. Systems receive a
    -- small immutable definition instead of repeatedly parsing legacy strings.
    M.fire_bolt = fire_bolt_data.load()
    cast_command.register()
    cast_system.register(M.fire_bolt)
end

function M.stop()
    M.fire_bolt = nil
end

return M
