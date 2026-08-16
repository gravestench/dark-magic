-- Character-roster authority adapters.
--
-- Character selection and creation have one presentation. The only thing that
-- differs between offline/direct play and Realm play is who owns the roster and
-- what happens after a character is selected. Keeping those differences in
-- this small module prevents the two frontend screens from drifting apart.

local scenes = require("engine.scene/v1")
local saves = require("d2legacy.save/v1")
local network_flow = require("d2legacy.network.flow")

local roster = {}

local function local_characters()
    return saves.characters()
end

local local_authority = {
    create_scene = "character_create",
    select_scene = "character_select",
    can_delete = true,
}

function local_authority.characters()
    return local_characters()
end

function local_authority.activate(_, character_id)
    assert(saves.select(character_id))
    if network_flow.start_selected() then
        scenes.replace("game_loading")
    end
end

function local_authority.create(_, name, class, expansion, hardcore)
    local id, err = saves.create_named(name, class, expansion, hardcore)
    if not id then
        return false, err
    end
    if network_flow.start_selected() then
        scenes.replace("game_loading")
    end
    return true
end

function local_authority.delete(_, character_id)
    return saves.delete(character_id), nil, true
end

function local_authority.leave_selection()
    scenes.replace(network_flow.cancel_destination("main_menu"))
end

function local_authority.leave_creation()
    if #local_characters() == 0 then
        local_authority.leave_selection()
    else
        scenes.replace(local_authority.select_scene)
    end
end

function local_authority.update()
    return false
end

local function realm_authority()
    -- Realm is intentionally loaded only for this adapter. Offline character
    -- screens therefore remain usable in runtimes that do not install Realm.
    local realm = require("engine.realm/v1")
    local realm_common = require("d2legacy.realm.common")
    local authority = {
        create_scene = "realm_create",
        select_scene = "realm_characters",
        can_delete = true,
    }

    function authority.characters()
        local result = {}
        for _, summary in ipairs(realm.status().characters or {}) do
            if summary.character then
                result[#result + 1] = summary.character
            end
        end
        return result
    end

    function authority.activate(scene, character_id)
        scene.roster_selection_requested = realm.select_character(character_id)
        if not scene.roster_selection_requested then
            scene.roster_error = "REALM IS BUSY"
        end
    end

    function authority.create(scene, name, class, expansion, hardcore)
        scene.roster_creation_requested = realm.create_character(name, class, expansion, hardcore)
        if not scene.roster_creation_requested then
            return false, "REALM IS BUSY"
        end
        return true
    end

    function authority.delete(scene, character_id)
        scene.roster_delete_requested = realm.delete_character(character_id)
        if not scene.roster_delete_requested then
            return false, "REALM IS BUSY", false
        end
        return true, nil, false
    end

    function authority.leave_selection()
        realm.cancel()
        scenes.replace("main_menu")
    end

    function authority.leave_creation()
        if #authority.characters() == 0 then
            authority.leave_selection()
        else
            scenes.replace(authority.select_scene)
        end
    end

    function authority.update(scene)
        local state = realm.status()
        if state.phase == "error" then
            scene.roster_error = realm_common.error(state)
            return false
        end

        if scene.roster_delete_requested and state.phase == "characters" then
            scene.roster_delete_requested = nil
            return true
        end

        local requested = scene.roster_selection_requested or scene.roster_creation_requested
        local selected = state.selected and state.selected.character
        if
            requested
            and selected
            and selected.id
            and (state.phase == "character_selected" or state.phase == "characters")
            and not scene.roster_join_requested
        then
            scene.roster_join_requested = realm.join_channel("Diablo II")
        elseif state.phase == "lobby" then
            scenes.replace("realm_lobby")
        end
        return false
    end

    return authority
end

function roster.for_mode(mode)
    if mode == "local" then
        return local_authority
    end
    if mode == "realm" then
        return realm_authority()
    end
    error("unknown character roster mode: " .. tostring(mode))
end

return roster
