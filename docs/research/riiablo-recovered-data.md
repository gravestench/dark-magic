# Riiablo recovered-data audit

This audit separates portable Diablo II knowledge from Riiablo implementation
details. The source project is Apache-2.0 licensed; imported tables retain their
original bytes and provenance under `internal/content/shim/data/recovered`.

## Imported declarative sources

| Riiablo source | Normalized meaning | Dark Magic owner |
| --- | --- | --- |
| `assets/data/quests.txt` | Quest IDs, act/order, prerequisites, icons, title and stage string keys | `dm.quest_catalog/v1` |
| `assets/data/speech.txt` | Logical `Sounds.txt` ID to localization key | `dm.quest_catalog/v1` |
| `assets/data/ds1types.txt` | DS1 definition ID to descriptive name and level type | `dm.map_catalog/v1` |
| `assets/data/obj.txt` | Act-local DS1 object ID to global `Objects.txt` ID | `dm.map_catalog/v1` |

DS1 object type `2` is static and uses `obj.txt`; type `1` is dynamic and uses
the mounted `MonPreset.txt` act-local ordering. Dark Magic applies the recovered
object mapping only to static records and preserves the original type/ID for
both. A session-owned resolver builds act-local `MonPreset.txt` indices from the
ordinary typed game-data snapshot for dynamic records. Both lookups are O(1)
and are created once rather than cloning catalogs for each decoded object.

## Dialogue behavior supported by evidence

`NpcDialogBox` and `DialogScroller` agree on a reusable payload contract:

1. Resolve a logical speech ID through `speech.txt`.
2. Resolve its `soundstr` through the active localization tables.
3. Interpret the first localized line as a timing value divided by 60.
4. Display the remaining lines while playing the logical sound ID.
5. Stop the sound when dialogue completes or is dismissed.

Dark Magic normalizes steps 1–3 through `dm.quest_catalog.dialog`. Lua owns the
presentation lifecycle and calls the existing locale/audio capabilities; this
avoids embedding font metrics or a particular widget toolkit in game data.

## Behavior intentionally not imported

Riiablo's `Npc.createMenu` constructs only Act I introduction and first-gossip
IDs from an NPC's lowercased display name. It is marked by incomplete menu and
inventory work and does not model quest-state selection, class variants, or the
full five acts. Treating that demonstration as recovered original behavior
would manufacture incorrect rules.

Likewise, Riiablo's quest panel always displays the first quest stage. Its UI
layout is useful implementation reference, but it is not evidence for quest
state transitions. The initial cross-source quest and world survey now lives in
[QUESTS_AND_WORLD.md](QUESTS_AND_WORLD.md), with the recovered runtime shape in
[QUEST_RUNTIME_MODEL.md](QUEST_RUNTIME_MODEL.md). Exact transitions and NPC
speech selection still require verified gameplay traces.
