# Diablo II gameplay systems research index

Status: living research index. All 28 gameplay workstreams now have implementation-oriented baselines; this table tracks research maturity, not implementation completion.

Target: expansion-only Diablo II: Lord of Destruction 1.14d. Classic and
earlier-patch compatibility are out of scope; older recovered sources are
secondary evidence where 1.14d behavior has not yet been measured.

Research program: [GAME_SYSTEMS_RESEARCH_PROGRAM.md](GAME_SYSTEMS_RESEARCH_PROGRAM.md).
Gameplay source inventory: [SYSTEMS_SOURCE_MATRIX.md](SYSTEMS_SOURCE_MATRIX.md).
Current implementation sequence: [NEXT_GAMEPLAY_IMPLEMENTATION_SEQUENCE.md](NEXT_GAMEPLAY_IMPLEMENTATION_SEQUENCE.md).

| # | Workstream | Primary document | Research status | Current-engine context | Dependencies | Open probes |
| --- | --- | --- | --- | --- | --- | --- |
| 28 | Content loading, TXT/BIN compilation, localization, and modability | [GAME_DATA_LOADING_AND_LINKAGE.md](GAME_DATA_LOADING_AND_LINKAGE.md) | baseline | existing owner inspected; fidelity gaps documented | content/VFS, localization | 8 |
| 24 | Time, RNG, simulation ticks, and determinism | [TIMING_RNG_AND_DETERMINISM.md](TIMING_RNG_AND_DETERMINISM.md) | baseline | fixed-tick session/replay/checkpoint owner exists; legacy timer/RNG details remain probes | session, ECS, replay | 10 |
| 20 | Persistent character state | [CHARACTER_PERSISTENCE_AND_SAVE_FORMAT.md](CHARACTER_PERSISTENCE_AND_SAVE_FORMAT.md) | baseline | trusted Dark Magic profile/Realm persistence exists; complete durable semantic character state remains open; vanilla `.d2s` interoperability is out of scope | data linkage, items, quests | Dark Magic durability only |
| 3 | Item stats, affixes, and property evaluation | [ITEM_STATS_AND_AFFIXES.md](ITEM_STATS_AND_AFFIXES.md) | baseline | M21.1 `engine.stats/v1` provenance/source authority implemented; ItemStatCost-derived operations remain gated by probes | data linkage, items, skills/states | 12 |
| 1 | Base items, equipment, and inventory | [ITEM_SYSTEM.md](ITEM_SYSTEM.md) | baseline | authoritative containers exist; an imported world placement now resolves generic room residency, survives inactive checkpoint/reactivation, and removes/reacquires spatial state on pickup/re-drop, while public loot ownership and richer item behavior remain | data linkage, item stats, persistence | 11 |
| 13 | Character creation, stats, levels, and progression | [CHARACTER_PROGRESSION.md](CHARACTER_PROGRESSION.md) | baseline | authoritative player entry/progression/vitals/skill assignment exist; full progression/difficulty/save synchronization remains | data linkage, timing, persistence, item stats, quests | 12 |
| 12 | Combat, damage, defense, and death | [COMBAT_DAMAGE_AND_DEATH.md](COMBAT_DAMAGE_AND_DEATH.md) | baseline | G8 separates attack outcomes from typed damage and gives monster/player death independent ECS markers; the first same-character player-death/action-filter transition is implemented, while strict owned-1.14d analyzers now gate both defense outcomes and softcore timing/corpse/equipment/gold/XP/save observations | stats, timing, skills, items | 17 |
| 11 | Skills, missiles, states, and combat actions | [SKILLS_STATES_AND_MISSILES.md](SKILLS_STATES_AND_MISSILES.md) | baseline | M21.6 immutable cast requests and M21.7 checkpointed cast lifecycle are implemented; straight/radial/area-impact missile families use ordinary ECS composition, Nova owns shared-cast locks, Fire Bolt snapshots hard-point percentages, and Fire Ball separates ordered radius damage from presentation-only aftermath while exact modifier/radius/RNG ordering remains probe-gated | timing, stats, combat, missiles/states | 22 |
| 8 | Monsters, spawning, AI, and lifecycle | [MONSTERS_SPAWNING_AND_AI.md](MONSTERS_SPAWNING_AND_AI.md) | baseline | ordinary-hostile materialization, scheduled acquire/chase/attack AI, population, death, and world-owned persistent room identity now preserve live state graphs, owned-unit relationships, ordinary corpses, and straight projectiles without parallel inactive archives; quality and broader activation graphs remain | timing, combat, world/pathfinding | 22 |
| 18 | Hirelings, mercenaries, pets, and owned units | [HIRELINGS_AND_OWNED_UNITS.md](HIRELINGS_AND_OWNED_UNITS.md) | baseline | generic ownership, deterministic limits/replacement, kill attribution, expiration, and owned-resident checkpoint/reactivation continuity exist; summon skills and broad pet/hireling behavior remain | monsters, items, skills, persistence | 24 |
| 23 | Movement, collision, pathfinding, and room streaming | [MOVEMENT_COLLISION_AND_PATHFINDING.md](MOVEMENT_COLLISION_AND_PATHFINDING.md) | baseline | pointer-first movement, deterministic A*, DT1 collision, and type-independent room activation now cover owned-unit graphs, corpses, projectiles, imported ground items, stateful objects, and separately resident pending-action relationships; exact 1.14d streaming/timer policy remains | world, timing, monsters | 25 |
| 2 | Item generation, treasure classes, and drops | [ITEM_GENERATION_AND_LOOT.md](ITEM_GENERATION_AND_LOOT.md) | baseline | death feeds effective party context into NoDrop, and an imported world-item fixture now proves room residency/pickup/re-drop mechanisms; public loot materialization, legal placement, allocation, pickup, lifetime, and target 1.14d breadth remain | data linkage, RNG, item system/stats | queued probes |
| 4 | Unique, set, rare, crafted, and personalized items | [ITEM_QUALITIES_AND_SETS.md](ITEM_QUALITIES_AND_SETS.md) | baseline | M6 special-item selection exists; durable recipes and conditional source activation remain | item generation + stats | queued probes |
| 5 | Sockets, gems, runes, jewels, and runewords | [SOCKETS_RUNES_GEMS_JEWELS.md](SOCKETS_RUNES_GEMS_JEWELS.md) | baseline | nested identity and runeword recognition/source-lifecycle slices documented; authoritative container foundation exists | item system + stats/qualities | queued probes |
| 6 | Charms and inventory-dependent effects | [CHARMS_AND_CONTAINER_EFFECTS.md](CHARMS_AND_CONTAINER_EFFECTS.md) | baseline | M21.1 now supplies the source model required for container-conditioned activation | item system + stats + containers | queued probes |
| 19 | Horadric Cube and recipe engine | [HORADRIC_CUBE_AND_RECIPES.md](HORADRIC_CUBE_AND_RECIPES.md) | baseline | authoritative Cube container exists; ordered matcher and atomic transformation engine remain | item system + stats + recipes | queued probes |
| 14 | Vendors, pricing, repair, gambling, and economy | [VENDORS_PRICING_AND_ECONOMY.md](VENDORS_PRICING_AND_ECONOMY.md) | baseline | server-owned vendor placement/buy/sell/service authority exists; full quote/stock/repair/gamble semantics remain | items + NPCs + difficulty | queued probes |
| 7 | Quest items and scripted item interactions | [QUEST_ITEMS.md](QUEST_ITEMS.md) | baseline | quest-service container/transactions exist; complete generation/pickup/drop/consume/difficulty lifecycle remains | items + quests + world objects | queued probes |
| 21 | World objects and environmental interactions | [WORLD_OBJECTS_AND_INTERACTIONS.md](WORLD_OBJECTS_AND_INTERACTIONS.md) | baseline | production DS1 targets attach to room residency; a synthetic data-selected object now proves ECS mode/used/seed/revision authority plus separately resident pending-action relationship continuity, while retail family mapping/event execution remains | world/map + quests + items | queued probes |
| 22 | Shrines, wells, curses, buffs, and temporary states | [SHRINES_WELLS_AND_TEMPORARY_EFFECTS.md](SHRINES_WELLS_AND_TEMPORARY_EFFECTS.md) | baseline | typed shrine data exists; M21.8 timed-state engine is the prerequisite for reusable shrine/buff behavior | stats/states + world objects | queued probes |
| 10 | Warps, portals, waypoints, teleportation, and transitions | [WARPS_PORTALS_WAYPOINTS.md](WARPS_PORTALS_WAYPOINTS.md) | baseline | trusted Act I transition owner/seam and paired Warp Lab interaction endpoints exist; generated room residency is attached, while authored endpoint breadth, waypoint state and dynamic portals remain | world/map + quests + persistence | queued probes |
| 9 | NPCs, dialogue, gossip, services, and town state | [NPCS_DIALOGUE_AND_SERVICES.md](NPCS_DIALOGUE_AND_SERVICES.md) | baseline | server-owned interaction context, vendor categories/services and range revalidation exist; dialogue/quest capability snapshots remain | quests + vendors + dialogue/localization | queued probes |
| 15 | Difficulty progression and game modes | [DIFFICULTY_AND_GAME_MODES.md](DIFFICULTY_AND_GAME_MODES.md) | baseline | immutable expansion-1.14d `d2legacy.game_rules/v2` is checkpointed and identity-pinned; separate `d2legacy.player_count/v1` follows live population or a checkpointed host override; remaining combat/loot/quest/economy consumers remain | progression + quests + loot + monsters | queued probes |
| 27 | Endgame and special events | [ENDGAME_AND_SPECIAL_EVENTS.md](ENDGAME_AND_SPECIAL_EVENTS.md) | baseline | encounter/portal composition and version gates documented; campaign/special-event controllers remain | quests + world + difficulty + combat | queued probes |
| 17 | Multiplayer parties, hostility, trading, and shared progress | [MULTIPLAYER_PARTIES_PVP_AND_TRADE.md](MULTIPLAYER_PARTIES_PVP_AND_TRADE.md) | baseline | checkpointed party authority, same-level queries, cleanup/reconnect, party-aware NoDrop, and a privacy-filtered owner party projection exist; XP/quest/gold, hostility, trade, and exact 1.14d UI behavior remain | session + items + quests + combat | queued probes |
| 16 | Battle.net, realm servers, and responsibilities | [REALM_AND_BATTLENET_ARCHITECTURE.md](REALM_AND_BATTLENET_ARCHITECTURE.md) | baseline | `server` owns the transport-independent Session and `realm` exists as a control-plane shell; leases/directory/storage/transports remain | session + persistence + multiplayer | queued probes |
| 25 | UI-visible game-state contracts | [GAMEPLAY_UI_CONTRACTS.md](GAMEPLAY_UI_CONTRACTS.md) | baseline | Lua follows snapshot -> intent -> authority -> next snapshot; `ClientView/v5` adds a bounded owner-scoped party contract consumed by the party panel, while other gameplay domains still need formal projections | all authoritative snapshots | queued probes |
| 26 | Audio, music, ambience, and gameplay cues | [GAMEPLAY_AUDIO_BEHAVIOR.md](GAMEPLAY_AUDIO_BEHAVIOR.md) | baseline | mixer/buses/groups/fades/streaming and typed Sounds resolver exist; rich Sounds semantics, semantic cue bridge, soundscapes and positional tracking remain | presentation + world/combat events | queued probes |

## Implementation cursor

The research program now has a baseline for every indexed workstream. Implementation has consumed the first seven combat/simulation handoff checkpoints:

- M21.1 shared stat/effect sources: implemented.
- M21.2 fixed-point combat vocabulary: implemented.
- M21.3 ordinary hostile materialization: implemented.
- M21.4 scheduled basic monster AI: implemented.
- M21.5 basic melee transaction: implemented.
- M21.6 assigned skill-intent consumption into immutable cast requests: implemented.
- M21.7 generic skill cast lifecycle: implemented.
- M21.8 timed state engine: implemented.
- G9 radial-missile tranche: active; exact-ID Nova now exercises generic
  level-scaled count/mana/five-band damage, shared-cast projectile composition,
  ECS contact locking, checkpoint parity, and cleanup. Exact owned-1.14d phase,
  acceleration, action timing, and repeat-contact ordering remain probes.
- G9 cross-skill modifier tranche: active; Fire Bolt's owned
  `EDmgSymPerCalc` resolves Fire Ball/Meteor hard levels and `Param8=16` into a
  checkpointed generic cast percentage applied after authored damage bands.
  Exact 1.14d rounding and ordering against mastery/source/PvP families remain
  probes.
- G9 impact-area tranche: active; exact-ID Fire Ball resolves swept contact
  into stable ordered radius damage and a separate checkpointed, non-damaging
  ECS explosion effect. Exact owned-1.14d footprint units, impact rounding,
  per-target RNG, and mastery/resistance/PvP ordering remain probes.
- G8 combat fidelity tranche 1: active; shared direct-damage commit/result and
  typed six-channel ECS stage facts implemented; death attribution now uses an
  independent ECS consumer marker, and basic melee exposes generic hit/miss/
  invalidated attack outcomes. Combat Lab coalesces that component composition
  into one diagnostic record. The block/avoidance owned-runtime analyzer and
  template are ready. A common lethal result now composes player-death state,
  action filters, attribution, and a semantic event onto the same character ECS
  entity. A separate strict analyzer/template is ready for probe-created
  softcore 1.14d visual death/corpse/equipment/gold/XP/save observations and
  rejects imported save data. Both observation matrices, ordered combat-family
  breadth, remaining consumers/retirement, drain/durations, and exact respawn/
  Hardcore death consequences remain.

Use [NEXT_GAMEPLAY_IMPLEMENTATION_SEQUENCE.md](NEXT_GAMEPLAY_IMPLEMENTATION_SEQUENCE.md) for the reconciled dependency queue and post-first-loop ordering across combat fidelity, items, world interactions, economy, audio, UI, networking, and realm persistence. `ROADMAP.md` remains authoritative for checkbox completion and the current G8 cursor.

## Current implementation handoffs

- [FOUNDATION_HANDOFF_FOR_CODEX.md](FOUNDATION_HANDOFF_FOR_CODEX.md) records the foundation slices; several have now graduated into M21 implementation.
- [COMBAT_SIMULATION_HANDOFF_FOR_CODEX.md](COMBAT_SIMULATION_HANDOFF_FOR_CODEX.md) remains the source-sensitive contract for M21.8-M21.12 and later combat fidelity.
- [ITEM_ECONOMY_AUDIO_HANDOFF_FOR_CODEX.md](ITEM_ECONOMY_AUDIO_HANDOFF_FOR_CODEX.md) orders generated-item, source-activation, economy, quest-item, and audio slices.
- [WORLD_NPC_DIFFICULTY_HANDOFF_FOR_CODEX.md](WORLD_NPC_DIFFICULTY_HANDOFF_FOR_CODEX.md) orders object, transition, NPC, difficulty, and endgame slices.
- [MULTIPLAYER_REALM_UI_HANDOFF_FOR_CODEX.md](MULTIPLAYER_REALM_UI_HANDOFF_FOR_CODEX.md) orders party, trade, realm, transport, persistence, and UI projection slices.
- [REALM_AND_GAME_CHAT_COMMANDS.md](REALM_AND_GAME_CHAT_COMMANDS.md) separates realm social, in-game chat, local diagnostics, and authoritative gameplay commands and records the legacy command verification queue.
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
