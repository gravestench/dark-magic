# Diablo II gameplay systems research index

Status: living research backlog. The index tracks research maturity, not implementation completion.

Research program: [GAME_SYSTEMS_RESEARCH_PROGRAM.md](GAME_SYSTEMS_RESEARCH_PROGRAM.md).
Gameplay source inventory: [SYSTEMS_SOURCE_MATRIX.md](SYSTEMS_SOURCE_MATRIX.md).

| # | Workstream | Primary document | Research status | Current-engine context | Dependencies | Open probes |
| --- | --- | --- | --- | --- | --- | --- |
| 28 | Content loading, TXT/BIN compilation, localization, and modability | [GAME_DATA_LOADING_AND_LINKAGE.md](GAME_DATA_LOADING_AND_LINKAGE.md) | baseline | existing owner inspected; fidelity gaps documented | content/VFS, localization | 8 |
| 24 | Time, RNG, simulation ticks, and determinism | [TIMING_RNG_AND_DETERMINISM.md](TIMING_RNG_AND_DETERMINISM.md) | baseline | existing owner inspected; fidelity gaps documented | session, ECS, replay | 10 |
| 20 | Save games and persistent character state | [CHARACTER_PERSISTENCE_AND_SAVE_FORMAT.md](CHARACTER_PERSISTENCE_AND_SAVE_FORMAT.md) | baseline | existing owner inspected; fidelity gaps documented | data linkage, items, quests | 12 |
| 3 | Item stats, affixes, and property evaluation | [ITEM_STATS_AND_AFFIXES.md](ITEM_STATS_AND_AFFIXES.md) | baseline | existing owner inspected; fidelity gaps documented | data linkage, items, skills/states | 12 |
| 1 | Base items, equipment, and inventory | [ITEM_SYSTEM.md](ITEM_SYSTEM.md) | baseline | existing owner inspected; fidelity gaps documented | data linkage, item stats, persistence | 11 |
| 13 | Character creation, stats, levels, and progression | [CHARACTER_PROGRESSION.md](CHARACTER_PROGRESSION.md) | baseline | existing owner inspected; fidelity gaps documented | data linkage, timing, persistence, item stats, quests | 12 |
| 12 | Combat, damage, defense, and death | [COMBAT_DAMAGE_AND_DEATH.md](COMBAT_DAMAGE_AND_DEATH.md) | baseline | session/item/player/targeting owners inspected; combat transaction slices documented | stats, timing, skills, items | 17 |
| 11 | Skills, missiles, states, and combat actions | [SKILLS_STATES_AND_MISSILES.md](SKILLS_STATES_AND_MISSILES.md) | baseline | existing `player.use_skill` admission inspected; execution/state/missile gap documented | timing, stats, combat, missiles/states | 22 |
| 8 | Monsters, spawning, AI, and lifecycle | [MONSTERS_SPAWNING_AND_AI.md](MONSTERS_SPAWNING_AND_AI.md) | baseline | typed monster data/world/targeting inspected; population/AI/lifecycle gaps documented | timing, combat, world/pathfinding | 22 |
| 18 | Hirelings, mercenaries, pets, and owned units | [HIRELINGS_AND_OWNED_UNITS.md](HIRELINGS_AND_OWNED_UNITS.md) | baseline | hireling/pet data and hireling item container inspected; owned-unit/persistence gaps documented | monsters, items, skills, persistence | 24 |
| 23 | Movement, collision, pathfinding, and room streaming | [MOVEMENT_COLLISION_AND_PATHFINDING.md](MOVEMENT_COLLISION_AND_PATHFINDING.md) | baseline | current deterministic A*, movement commands, zone transitions and diagnostics inspected; legacy path-type/streaming gaps documented | world, timing, monsters | 25 |
| 2 | Item generation, treasure classes, and drops | [ITEM_GENERATION_AND_LOOT.md](ITEM_GENERATION_AND_LOOT.md) | baseline | M6 owner reconciled; context-sensitive quality/NoDrop gaps documented | data linkage, RNG, item system/stats | queued probes |
| 4 | Unique, set, rare, crafted, and personalized items | [ITEM_QUALITIES_AND_SETS.md](ITEM_QUALITIES_AND_SETS.md) | baseline | durable recipe and conditional-source slices documented | item generation + stats | queued probes |
| 5 | Sockets, gems, runes, jewels, and runewords | [SOCKETS_RUNES_GEMS_JEWELS.md](SOCKETS_RUNES_GEMS_JEWELS.md) | baseline | nested identity and recognition slices documented | item system + stats/qualities | queued probes |
| 6 | Charms and inventory-dependent effects | [CHARMS_AND_CONTAINER_EFFECTS.md](CHARMS_AND_CONTAINER_EFFECTS.md) | baseline | container-conditioned source activation documented | item system + stats + containers | queued probes |
| 19 | Horadric Cube and recipe engine | [HORADRIC_CUBE_AND_RECIPES.md](HORADRIC_CUBE_AND_RECIPES.md) | baseline | ordered matcher/transaction slices documented | item system + stats + recipes | queued probes |
| 14 | Vendors, pricing, repair, gambling, and economy | [VENDORS_PRICING_AND_ECONOMY.md](VENDORS_PRICING_AND_ECONOMY.md) | baseline | existing commerce owner reconciled; quote/stock/gamble gaps documented | items + NPCs + difficulty | queued probes |
| 7 | Quest items and scripted item interactions | [QUEST_ITEMS.md](QUEST_ITEMS.md) | baseline | item lifecycle/event boundary documented | items + quests + world objects | queued probes |
| 21 | World objects and environmental interactions | [WORLD_OBJECTS_AND_INTERACTIONS.md](WORLD_OBJECTS_AND_INTERACTIONS.md) | baseline | selector/interaction/world owners reconciled; object-instance slices documented | world/map + quests + items | queued probes |
| 22 | Shrines, wells, curses, buffs, and temporary states | [SHRINES_WELLS_AND_TEMPORARY_EFFECTS.md](SHRINES_WELLS_AND_TEMPORARY_EFFECTS.md) | baseline | shared timed stat-source dependency documented | stats/states + world objects | queued probes |
| 10 | Warps, portals, waypoints, teleportation, and transitions | [WARPS_PORTALS_WAYPOINTS.md](WARPS_PORTALS_WAYPOINTS.md) | baseline | transition owner reconciled; endpoint/waypoint/portal slices documented | world/map + quests + persistence | queued probes |
| 9 | NPCs, dialogue, gossip, services, and town state | [NPCS_DIALOGUE_AND_SERVICES.md](NPCS_DIALOGUE_AND_SERVICES.md) | baseline | interaction/item/quest ownership and capability snapshots documented | quests + vendors + dialogue/localization | queued probes |
| 15 | Difficulty progression and game modes | [DIFFICULTY_AND_GAME_MODES.md](DIFFICULTY_AND_GAME_MODES.md) | baseline | immutable GameRules and first scalar consumers documented | progression + quests + loot + monsters | queued probes |
| 27 | Endgame and special events | [ENDGAME_AND_SPECIAL_EVENTS.md](ENDGAME_AND_SPECIAL_EVENTS.md) | baseline | encounter/portal composition and version gates documented | quests + world + difficulty + combat | queued probes |
| 17 | Multiplayer parties, hostility, trading, and shared progress | [MULTIPLAYER_PARTIES_PVP_AND_TRADE.md](MULTIPLAYER_PARTIES_PVP_AND_TRADE.md) | baseline | party/hostility/trade state and authority boundaries documented | session + items + quests + combat | queued probes |
| 16 | Battle.net, realm servers, and responsibilities | [REALM_AND_BATTLENET_ARCHITECTURE.md](REALM_AND_BATTLENET_ARCHITECTURE.md) | baseline | lease/worker/directory/transport slices documented | session + persistence + multiplayer | queued probes |
| 25 | UI-visible game-state contracts | [GAMEPLAY_UI_CONTRACTS.md](GAMEPLAY_UI_CONTRACTS.md) | baseline | revisioned semantic projection boundary documented | all authoritative snapshots | queued probes |
| 26 | Audio, music, ambience, and gameplay cues | [GAMEPLAY_AUDIO_BEHAVIOR.md](GAMEPLAY_AUDIO_BEHAVIOR.md) | baseline | semantic event and non-authoritative soundscape slices documented | presentation + world/combat events | queued probes |

## Current implementation handoffs

- [FOUNDATION_HANDOFF_FOR_CODEX.md](FOUNDATION_HANDOFF_FOR_CODEX.md) turns the first six foundation baselines into bounded implementation checkpoints.
- [COMBAT_SIMULATION_HANDOFF_FOR_CODEX.md](COMBAT_SIMULATION_HANDOFF_FOR_CODEX.md) orders the combat/skill/monster/movement/owned-unit vertical slices.
- [ITEM_ECONOMY_AUDIO_HANDOFF_FOR_CODEX.md](ITEM_ECONOMY_AUDIO_HANDOFF_FOR_CODEX.md) orders generated-item, source-activation, economy, quest-item, and audio slices.
- [WORLD_NPC_DIFFICULTY_HANDOFF_FOR_CODEX.md](WORLD_NPC_DIFFICULTY_HANDOFF_FOR_CODEX.md) orders object, transition, NPC, difficulty, and endgame slices.
- [MULTIPLAYER_REALM_UI_HANDOFF_FOR_CODEX.md](MULTIPLAYER_REALM_UI_HANDOFF_FOR_CODEX.md) orders party, trade, realm, transport, persistence, and UI projection slices.
- [FOUNDATION_VERIFICATION_QUEUE.md](FOUNDATION_VERIFICATION_QUEUE.md) tracks the highest-value foundation probes.
- [COMBAT_SIMULATION_VERIFICATION_QUEUE.md](COMBAT_SIMULATION_VERIFICATION_QUEUE.md) tracks the empirical combat/simulation probes.
- [ITEM_ECONOMY_AUDIO_VERIFICATION_QUEUE.md](ITEM_ECONOMY_AUDIO_VERIFICATION_QUEUE.md), [WORLD_NPC_DIFFICULTY_VERIFICATION_QUEUE.md](WORLD_NPC_DIFFICULTY_VERIFICATION_QUEUE.md), and [MULTIPLAYER_REALM_UI_VERIFICATION_QUEUE.md](MULTIPLAYER_REALM_UI_VERIFICATION_QUEUE.md) track the later tranche probes.

## Existing cross-cutting research

- [QUESTS_AND_WORLD.md](QUESTS_AND_WORLD.md) covers the five-act player route, world gates, required set pieces, and quest/world coupling.
- [QUEST_RUNTIME_MODEL.md](QUEST_RUNTIME_MODEL.md) covers quest bits, live controller state, events, party credit, and authoritative quest ownership.
- [riiablo-recovered-data.md](riiablo-recovered-data.md) records the imported declarative quest/dialogue/object mappings and their limits.
- [../MAP_GENERATION.md](../MAP_GENERATION.md) covers DRLG, map composition, collision, semantic tiles, warps, and deterministic Dark Magic map-generation policy.
- [../GAME_DATA_RECORDS.md](../GAME_DATA_RECORDS.md) is the existing typed/generic table admission policy and remains normative for table-loading changes.

## Status vocabulary

- **queued**: no dedicated implementation-oriented document yet.
- **researching**: sources are being inspected and conflicts/probes are still being assembled.
- **baseline**: ownership, major state/data model, implementation slices, and actionable probes are documented.
- **validated**: the important unresolved probes have been answered by owned-game traces, byte-exact fixtures, or multiple strong independent sources.

Implementation status is deliberately tracked separately in `ROADMAP.md`.
