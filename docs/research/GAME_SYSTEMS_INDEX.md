# Diablo II gameplay systems research index

Status: living research backlog. The index tracks research maturity, not implementation completion.

Research program: [GAME_SYSTEMS_RESEARCH_PROGRAM.md](GAME_SYSTEMS_RESEARCH_PROGRAM.md).
Gameplay source inventory: [SYSTEMS_SOURCE_MATRIX.md](SYSTEMS_SOURCE_MATRIX.md).

| # | Workstream | Primary document | Research status | Current-engine context | Dependencies | Open probes |
| --- | --- | --- | --- | --- | --- | --- |
| 28 | Content loading, TXT/BIN compilation, localization, and modability | [GAME_DATA_LOADING_AND_LINKAGE.md](GAME_DATA_LOADING_AND_LINKAGE.md) | baseline in this research PR | existing owner inspected; fidelity gaps documented | content/VFS, localization | 8 |
| 24 | Time, RNG, simulation ticks, and determinism | [TIMING_RNG_AND_DETERMINISM.md](TIMING_RNG_AND_DETERMINISM.md) | baseline in this research PR | existing owner inspected; fidelity gaps documented | session, ECS, replay | 10 |
| 20 | Save games and persistent character state | [CHARACTER_PERSISTENCE_AND_SAVE_FORMAT.md](CHARACTER_PERSISTENCE_AND_SAVE_FORMAT.md) | baseline in this research PR | existing owner inspected; fidelity gaps documented | data linkage, items, quests | 12 |
| 3 | Item stats, affixes, and property evaluation | [ITEM_STATS_AND_AFFIXES.md](ITEM_STATS_AND_AFFIXES.md) | baseline in this research PR | existing owner inspected; fidelity gaps documented | data linkage, items, skills/states | 12 |
| 1 | Base items, equipment, and inventory | [ITEM_SYSTEM.md](ITEM_SYSTEM.md) | baseline in this research PR | existing owner inspected; fidelity gaps documented | data linkage, item stats, persistence | 11 |
| 13 | Character creation, stats, levels, and progression | [CHARACTER_PROGRESSION.md](CHARACTER_PROGRESSION.md) | baseline in this research PR | existing owner inspected; fidelity gaps documented | data linkage, timing, persistence, item stats, quests | 12 |
| 12 | Combat, damage, defense, and death | `COMBAT_DAMAGE_AND_DEATH.md` | queued | dedicated research not yet consolidated | stats, timing, skills, items | TBD |
| 11 | Skills, missiles, states, and combat actions | `SKILLS_STATES_AND_MISSILES.md` | queued | dedicated research not yet consolidated | timing, stats, combat, missiles/states | TBD |
| 8 | Monsters, spawning, AI, and lifecycle | `MONSTERS_SPAWNING_AND_AI.md` | queued | dedicated research not yet consolidated | timing, combat, world/pathfinding | TBD |
| 18 | Hirelings, mercenaries, pets, and owned units | `HIRELINGS_AND_OWNED_UNITS.md` | queued | dedicated research not yet consolidated | monsters, items, skills, persistence | TBD |
| 23 | Movement, collision, pathfinding, and room streaming | `MOVEMENT_COLLISION_AND_PATHFINDING.md` | queued | Map generation and world movement implementation provide partial support | world, timing, monsters | TBD |
| 2 | Item generation, treasure classes, and drops | `ITEM_GENERATION_AND_LOOT.md` | queued | M6 implementation exists; consolidation queued | data linkage, RNG, item system/stats | TBD |
| 4 | Unique, set, rare, crafted, and personalized items | `ITEM_QUALITIES_AND_SETS.md` | queued | dedicated research not yet consolidated | item generation + stats | TBD |
| 5 | Sockets, gems, runes, jewels, and runewords | `SOCKETS_RUNES_GEMS_JEWELS.md` | queued | dedicated research not yet consolidated | item system + stats/qualities | TBD |
| 6 | Charms and inventory-dependent effects | `CHARMS_AND_CONTAINER_EFFECTS.md` | queued | dedicated research not yet consolidated | item system + stats + containers | TBD |
| 19 | Horadric Cube and recipe engine | `HORADRIC_CUBE_AND_RECIPES.md` | queued | dedicated research not yet consolidated | item system + stats + recipes | TBD |
| 14 | Vendors, pricing, repair, gambling, and economy | `VENDORS_PRICING_AND_ECONOMY.md` | queued | dedicated research not yet consolidated | items + NPCs + difficulty | TBD |
| 7 | Quest items and scripted item interactions | `QUEST_ITEMS.md` | queued | Quest runtime docs provide supporting behavior; dedicated item doc queued | items + quests + world objects | TBD |
| 21 | World objects and environmental interactions | `WORLD_OBJECTS_AND_INTERACTIONS.md` | queued | Map/quest docs provide partial support | world/map + quests + items | TBD |
| 22 | Shrines, wells, curses, buffs, and temporary states | `SHRINES_WELLS_AND_TEMPORARY_EFFECTS.md` | queued | dedicated research not yet consolidated | stats/states + world objects | TBD |
| 10 | Warps, portals, waypoints, teleportation, and transitions | `WARPS_PORTALS_WAYPOINTS.md` | queued | Map/quest docs provide partial support | world/map + quests + persistence | TBD |
| 9 | NPCs, dialogue, gossip, services, and town state | `NPCS_DIALOGUE_AND_SERVICES.md` | queued | Quest/dialogue recovered-data docs provide partial support | quests + vendors + dialogue/localization | TBD |
| 15 | Difficulty progression and game modes | `DIFFICULTY_AND_GAME_MODES.md` | queued | dedicated research not yet consolidated | progression + quests + loot + monsters | TBD |
| 27 | Endgame and special events | `ENDGAME_AND_SPECIAL_EVENTS.md` | queued | dedicated research not yet consolidated | quests + world + difficulty + combat | TBD |
| 17 | Multiplayer parties, hostility, trading, and shared progress | `MULTIPLAYER_PARTIES_PVP_AND_TRADE.md` | queued | dedicated research not yet consolidated | session + items + quests + combat | TBD |
| 16 | Battle.net, realm servers, and responsibilities | `REALM_AND_BATTLENET_ARCHITECTURE.md` | queued | dedicated research not yet consolidated | session + persistence + multiplayer | TBD |
| 25 | UI-visible game-state contracts | `GAMEPLAY_UI_CONTRACTS.md` | queued | dedicated research not yet consolidated | all authoritative snapshots | TBD |
| 26 | Audio, music, ambience, and gameplay cues | `GAMEPLAY_AUDIO_BEHAVIOR.md` | queued | dedicated research not yet consolidated | presentation + world/combat events | TBD |

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
