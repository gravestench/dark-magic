-- First-party Diablo II gameplay composition root.
--
-- This file deliberately contains no formulas or mutations. Its only job is to
-- show which small modules make up the mod and register them in readable order.

local components = require("d2legacy.components.fire_bolt")
local fire_bolt_data = require("d2legacy.data.fire_bolt")
local cast_command = require("d2legacy.commands.cast")
local cast_system = require("d2legacy.systems.cast")
local fire_bolt_system = require("d2legacy.systems.fire_bolt")
local projectile_system = require("d2legacy.systems.projectile")
local melee_components = require("d2legacy.components.melee")
local melee_system = require("d2legacy.systems.melee")

local M = {
    id = "d2legacy.authoritative",
}

function M.start()
    components.register()
    melee_components.register()

    -- Record interpretation happens once during composition. Systems receive a
    -- small immutable definition instead of repeatedly parsing legacy strings.
    M.fire_bolt = fire_bolt_data.load()
    cast_command.register()
    cast_system.register(M.fire_bolt)
    fire_bolt_system.register(M.fire_bolt)
    projectile_system.register()
    melee_system.register()
end

function M.stop()
    M.fire_bolt = nil
end

return M
