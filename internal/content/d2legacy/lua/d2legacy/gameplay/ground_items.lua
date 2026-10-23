-- Retained ground art reads the same public presentation on local and remote worlds.

local ecs = require("engine.ecs/v1")
local render = require("engine.render/v1")
local M = {}

local function destroy(entry)
    if entry.pending_job then
        render.preload_release(entry.pending_job)
    end
    entry.node:destroy()
end

function M.new(parent, world, level_id, palette)
    local width, height = world:canvas()
    local view = { entries = {} }

    function view:update()
        local seen = {}
        for _, entity in
            ipairs(ecs.query({
                all = {
                    "d2legacy.item.presentation",
                    "d2legacy.world.selectable",
                    "d2legacy.world.position",
                    "d2legacy.world.location",
                },
                none = { "d2legacy.world.inactive" },
            }))
        do
            local location = ecs.get(entity, "d2legacy.world.location")
            local presentation = ecs.get(entity, "d2legacy.item.presentation")
            local path = presentation:get("world_dc6")
            if location:get("level_id") == level_id and path ~= "" then
                local key = tostring(entity:id())
                seen[key] = true
                local entry = self.entries[key]
                if not entry then
                    local animated = presentation:get("world_animated")
                    entry = { node = render.create("world", parent), path = path, animated = animated }
                    entry.node:set_visible(false)
                    entry.pending_job = render.preload({
                        {
                            kind = animated and "dc6_animation" or "dc6_frame",
                            path = path,
                            palette = palette,
                            direction = 0,
                            frame = 0,
                            anchor = "offsets",
                        },
                    })
                    self.entries[key] = entry
                end
                if entry.pending_job then
                    local status = render.preload_status(entry.pending_job)
                    if status and status.done then
                        render.preload_release(entry.pending_job)
                        entry.pending_job = nil
                        if status.failed == 0 then
                            if entry.animated then
                                local _, w, h, x, y =
                                    entry.node:set_dc6_animation(entry.path, palette, 0, 25, "once", "offsets")
                                entry.offset_x, entry.offset_y = x + w / 2, y + h / 2
                            else
                                local w, h, x, y = entry.node:set_dc6(entry.path, palette, 0, 0)
                                entry.offset_x, entry.offset_y = x + w / 2, y + h / 2
                            end
                            entry.node:set_visible(true)
                        end
                    end
                end
                local position = ecs.get(entity, "d2legacy.world.position")
                local sx, sy = position:get("x"), position:get("y")
                local x, y = world:subtile_to_pixel(sx, sy)
                entry.node:set_position(x - width / 2 + (entry.offset_x or 0), y - height / 2 + (entry.offset_y or 0))
                entry.node:set_z(world:entity_depth(sx, sy))
            end
        end
        for key, entry in pairs(self.entries) do
            if not seen[key] then
                destroy(entry)
                self.entries[key] = nil
            end
        end
    end

    function view:destroy()
        for _, entry in pairs(self.entries) do
            destroy(entry)
        end
        self.entries = {}
    end

    return view
end

return M
