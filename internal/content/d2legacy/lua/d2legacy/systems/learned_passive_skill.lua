-- Reconcile record-declared learned-skill passives into stat sources.
--
-- The learned-skill entity owns the source through component composition, so
-- forgetting the skill or removing the player also removes the modifier. A
-- passive declared by an aura record is suppressed while that same aura is
-- selected, leaving the active party relationship as the only source.

local ecs = require("engine.ecs/v1")
local M = {}

local function source_id(owner, skill_id)
    return "skill-passive:" .. owner:id() .. ":" .. skill_id
end

local function desired_source(learned, definitions)
    local definition = definitions[learned:get("skill_id")]
    local passive = definition and definition.learned_passive or nil
    local level = learned:get("level")
    if not passive or level <= 0 then
        return nil
    end
    local owner = learned:get("owner")
    local assignment = ecs.get(owner, "d2legacy.player.skill_assignment")
    if assignment and assignment:get("right") == definition.skill_id then
        return nil
    end
    return {
        target = owner,
        source_id = source_id(owner, definition.skill_id),
        stat = passive.stat,
        operation = passive.operation,
        value = level * passive.value_per_hard_level,
        order = 250,
    }
end

local function update_source(source, desired)
    source:set("target", desired.target)
    source:set("source_id", desired.source_id)
    source:set("stat", desired.stat)
    source:set("operation", desired.operation)
    source:set("value", desired.value)
    source:set("order", desired.order)
end

function M.register(definitions)
    ecs.system({
        id = "d2legacy.skill.learned_passive_stat",
        phase = "pre_simulation",
        query = { all = { "d2legacy.player.learned_skill" } },
        read = { "d2legacy.player.learned_skill", "d2legacy.player.skill_assignment", "d2legacy.stat.source" },
        write = { "d2legacy.stat.source" },
        update = function(_, entities, structural)
            for _, entity in ipairs(entities) do
                local learned = ecs.get(entity, "d2legacy.player.learned_skill")
                local current = ecs.get(entity, "d2legacy.stat.source")
                local desired = desired_source(learned, definitions)
                if desired then
                    if current then
                        assert(
                            current:get("source_id") == desired.source_id,
                            "learned skill entity already owns an unrelated stat source"
                        )
                        update_source(current, desired)
                    else
                        structural:set(entity, "d2legacy.stat.source", desired)
                    end
                elseif
                    current and current:get("source_id") == source_id(learned:get("owner"), learned:get("skill_id"))
                then
                    structural:remove(entity, "d2legacy.stat.source")
                end
            end
        end,
    })
end

return M
