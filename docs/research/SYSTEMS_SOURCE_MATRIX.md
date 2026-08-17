# Diablo II gameplay-systems research source matrix

Status: living source inventory for `docs/research/` gameplay-system work.

The legacy binary-format source inventory remains in [../formats/SOURCE_MATRIX.md](../formats/SOURCE_MATRIX.md). This matrix adds gameplay, save, stat, item, timing, combat, skill, AI, movement, world-interaction, economy, multiplayer, realm, audio, and verification sources without changing the format research evidence hierarchy.

## Evidence grades

- **A / verified**: lawfully owned original-game bytes or repeatable retail-engine trace.
- **B / high**: pinned reverse-engineered original-runtime behavior, preferably D2MOO, with clear routine/data relationship, or applicable contemporary first-party behavioral documentation.
- **C / medium**: credible independent implementation, historical format documentation, or corroborated community research.
- **D / lead only**: current guide/wiki/forum/player report.

## Inspected sources

| Source | Version / commit | License | Grade / role | Material used in current research docs |
| --- | --- | --- | --- | --- |
| Dark Magic repository | current `main` at each research/consolidation pass | MIT project source | implementation baseline | typed game data, 25 Hz session/replay, stat-source authority, fixed-point combat, monster materialization/AI/melee, skill cast requests, item authority, loot, generated world, transitions, audio backend/catalog, Lua snapshot/intent boundary, server/realm composition roots, roadmap |
| ThePhrozenKeep/D2MOO | `3b21043b99e987bad41cf0f7b49f1f246db52d5c` | MIT | B; reconstructed 1.10f runtime | seed arithmetic, items/stat lists, combat/damage, skills/missiles/states, monsters/population/AI, pets/hirelings, path types, inactive units, item generation/quality, sockets/runewords, charms, Cube, vendors, quest items, objects/shrines/transitions, NPCs, difficulty, party/hostility/trade, campaign/endgame |
| jaenster/libd2 | `e6cdc4927c6180be8dd309b0423b470f64f1fc6c` | MIT for clean-room source; exclude embedded Blizzard-derived data | C corroboration with unusually strong verification methodology | retail-capture/holdout methodology; save byte-exact round trip; item rolls traced against binary; pathfinding checked against generated maps/server movement gates |
| nokka/d2s | `master` inspected 2026-08-10 | MIT | C; independent save parser/document | `.d2s` header/section layout, bit-packed stats/skills/items, quest/waypoint/NPC blocks, corpse/merc/golem sections |
| PvPGN server | `9cd173f4e02ba3d9f8f15a67ca308b5eb78723e4` | GPL-2.0 | C; independent legacy Battle.net/Diablo-II server corroboration only | realm/control-server/game-server/database responsibility split, realm character/game ownership, server selection, character directories, ladder storage, queue/lifetime policy; not evidence of Blizzard's private implementation topology |
| Blizzard classic Battle.net documentation | live archive inspected 2026-08-14; target-patch applicability must still be checked | first-party public documentation | B for documented user-visible Battle.net behavior; below owned-client traces for exact patch behavior | chat commands and aliases, friends/privacy controls, channel types/restrictions/operator behavior, Diablo II realm-scoped channels |
| Diablo Wiki / Fandom | live pages inspected 2026-08-14; individual article revisions vary | CC BY-SA community reference; do not copy content into runtime | D by default; C only after independent corroboration | broad Diablo II research discovery, patch and terminology leads, realm/game chat command catalog; follow citations and corroborate before making compatibility claims |
| bundled Diablo II Data File Guide | repository `docs/diabloiidatafileguide.html.gz` | historical community documentation | C semantic guide | TXT field meanings, formula/linkage leads; mounted game bytes remain stronger |
| Dark Magic recovered Riiablo data | pinned provenance recorded in `riiablo-recovered-data.md` | Apache-2.0 upstream | C declarative corroboration | quest/speech/DS1 object declarative mappings |

GPL sources are behavioral/research references only. Do not mechanically translate GPL implementation into Dark Magic.

## D2MOO paths inspected for foundation research

- `source/D2Common/include/D2Seed.h`
- `source/D2Common/src/D2Seed.cpp`
- `source/D2Common/src/Items/Items.cpp`
- `source/D2Common/src/D2StatList.cpp`
- quest paths already listed in [../formats/SOURCE_MATRIX.md](../formats/SOURCE_MATRIX.md)

### Foundation claims supported

`D2Seed.h` supports the two-word seed state, initialization high word `666`,
`high + 0x6AC690C5 * low` roll recurrence, modulo/mask limited rolls, and
percentage roll. `D2Seed.cpp` supports the distinction between deterministic
stream evolution and wall-clock/tick-count-derived seed acquisition.

`Items.cpp` supports first-class per-item item/start seed, owner/body-location,
quality, affix, flag, item-level, inventory-page/cell, ear, and graphics fields.

`D2StatList.cpp` supports packed stat/layer identity, ordered stat storage,
linked/extended stat sources, ItemStatCost-controlled keep-zero/callback/op
behavior, `ValShift` use, and derived dependencies. Exact formulas and rounding
still require traces; reverse-engineered function names are not a specification.

## D2MOO paths inspected for combat/simulation research

### Combat and damage

- `source/D2Game/include/UNIT/SUnitDmg.h`
- `source/D2Game/src/UNIT/SUnitDmg.cpp`
- `source/D2Common/src/D2StatList.cpp`
- related `D2Items`, `D2States`, `D2Combat`, `DifficultyLevels`, monster and skill table paths reached from the damage pipeline

High-value observations include 8-bit fractional damage/life arithmetic, attacker-seeded damage rolls, target-type bonuses, critical/deadly-strike ordering evidence, typed conversion and mitigation, leech/drain difficulty handling, timed state effects, semantic unit events, and death-result handling. Exact chance-to-hit/block/PvP and several secondary effects remain probes.

### Skills and missiles

- `source/D2Game/src/SKILLS/Skills.cpp`
- class-specific and monster-specific `source/D2Game/src/SKILLS/Skill*.cpp`
- `source/D2Game/src/MISSILES/Missiles.cpp`
- `source/D2Game/src/MISSILES/MissMode.cpp`
- `source/D2Common/src/D2States.cpp`
- `source/D2Common/src/D2StatList.cpp`

`Skills.cpp` supports a data-driven server-start/server-do behavior-family model rather than one implementation per displayed skill name. Missile creation supports owner/source identity, target unit/position, level-scaled velocity/range, activation frame, collision masks, acceleration/max velocity, skill identity/level, damage snapshot stats, deterministic pierce budget, and path configuration.

### Monsters and AI

- `source/D2Game/src/MONSTER/MonsterRegion.cpp`
- `source/D2Game/src/MONSTER/MonsterSpawn.cpp`
- `source/D2Game/src/MONSTER/MonsterAI.cpp`
- `source/D2Game/src/MONSTER/Monster.cpp`
- `source/D2Game/src/MONSTER/MonsterMode.cpp`
- `source/D2Game/src/MONSTER/MonsterUnique.cpp`
- `source/D2Game/src/AI/AiGeneral.cpp`
- `source/D2Game/src/AI/AiUtil.cpp`
- `source/D2Game/src/AI/AiTactics.cpp`

`AiGeneral.cpp` supports persistent owner/minion/parameter/command/map-AI state. `MonsterRegion.cpp` supports deterministic room/population selection, preset pseudo-identities, quest/difficulty substitutions, group/champion placement and collision-aware spawn constraints. `MonsterSpawn.cpp` contains useful surrounding rules but its central normal-spawn reconstruction is still stubbed/commented; do not promote exact initialization order from that incomplete body.

### Hirelings and pets

- `source/D2Game/src/PLAYER/PlayerPets.cpp`
- `source/D2Game/include/PLAYER/PlayerPets.h`
- mercenary paths in `source/D2Game/src/MONSTER/MonsterAI.cpp`
- player-save/NPC/service paths reached from hireling flows

`PlayerPets.cpp` supports per-PetType ownership lists and maxima, special hireable lifecycle, minion ownership and deterministic excess removal machinery. `MonsterAI.cpp` supplies high-value 1.10f hireling level/stat/skill formulas and demonstrates monster-runtime reuse with explicit player ownership.

### Movement and room streaming

- `source/D2Common/include/Path/Path.h`
- `source/D2Common/src/Path/AStar.cpp`
- `source/D2Common/src/Path/IDAStar.cpp`
- `source/D2Common/src/Path/PathWF.cpp`
- `source/D2Common/src/Path/Path.cpp`
- `source/D2Common/src/Path/PathMisc.cpp`
- `source/D2Common/src/Path/Step.cpp`
- `source/D2Common/src/Units/Units.cpp`
- `source/D2Common/src/Drlg/DrlgDrlg.cpp`
- `source/D2Game/src/PLAYER/PlrModes.cpp`
- `source/D2Game/src/UNIT/SUnitInactive.cpp`

`Path.h` exposes 18 semantic path types, separate footprint/move-test collision masks, 16-bit fractional coordinates, velocity/acceleration state, target-unit state, collision patterns, and saved steps. `SUnitInactive.cpp` supports a real inactive-room unit archive with enough information to reconstruct monster identity, quality/modifiers, alignment, AI/minion state, XP, HP and special parameters rather than tying simulation existence to render residency.

`Units.cpp` supports the separate item-FRW diminishing-return conversion and
additive `velocitypercent` path-speed channel. `PlrModes.cpp` supports the 25 Hz
8.8 RunDrain, torso-armor/slower-drain, idle/walk/town recovery, and exhaustion
fallback structure. `DrlgDrlg.cpp` identifies the five town level IDs used by
that exception. These are reconstructed earlier-runtime evidence; mounted
Expansion 1.14d CharStats/ItemStatCost/Properties rows and future owned-runtime
holdouts remain the target authority.

## D2MOO paths inspected for item/economy research

- `source/D2Game/include/ITEMS/Items.h`
- `source/D2Game/src/ITEMS/Items.cpp`
- `source/D2Game/src/ITEMS/ItemMode.cpp`
- `source/D2Common/src/Items/Items.cpp`
- `source/D2Common/src/Items/ItemMods.cpp`
- `source/D2Common/include/DataTbls/HoradricCube.h`
- `source/D2Common/src/DataTbls/HoradricCube.cpp`
- `source/D2Game/include/PLAYER/PlrTrade.h`
- `source/D2Game/src/PLAYER/PlrTrade.cpp`
- `source/D2Game/include/UNIT/SUnitNpc.h`
- `source/D2Game/src/UNIT/SUnitNpc.cpp`

### Item-generation claims supported

The inspected TreasureClass path supports recursive class traversal, player/party-aware NoDrop adjustment, propagation of stronger nested quality modifiers, owner/minion Magic Find and Gold Find participation, explicit Unique -> Set -> Rare -> Magic checks in the TC path, and contextual item construction flags. A lower-level/direct quality helper uses a different Unique -> Rare -> Set -> Magic ordering, so quality policy is generation-context-sensitive rather than one universal helper order.

The item construction path also exposes independent seed/start-seed, quality, quantity, minimum/maximum durability, forced affix IDs, special record identity, socket and ethereal suppression/forcing, and quest-item handling. Quest items bypass ordinary quality behavior in inspected paths and are exempt from ordinary ground cleanup in the inspected 1.10f routine.

### Sockets, runewords, charms, and effects

`D2Common/src/Items/Items.cpp` supports socket children as nested item identities and runeword lookup by full socket occupancy, ordered socketed item IDs, included item types, and excluded item types. `ItemMode.cpp` demonstrates conditional stat-list activation for equipped items and usable charms rather than permanent flattening of every item modifier into player state.

`ItemMods.cpp` supports item property dispatch into stat/state/skill/aura/charged-skill behaviors and reinforces the requirement for parameterized stat/source identity.

### Cube and vendor/economy

`HoradricCube.h/.cpp` and `PlrTrade.cpp` support `CubeMain` as a declarative matcher/transformation language with up to seven inputs and three outputs, input flags for code/type/quality/socket/ethereal/grade/runeword restrictions, stat predicates, and output flags for copy/socket/ethereal/unsocket/remove/grade/repair/recharge behavior.

`D2Common/src/Items/Items.cpp` exposes a transaction-cost routine that depends on more than base price and NPC multipliers, including item quality/affixes, durability/repairability, charged skills, vendor/difficulty/quest context, max-buy caps, and reduced-price stats. The exact arithmetic still requires independent traces before it becomes a compatibility contract.

`SUnitNpc.cpp` supports distinct vendor inventory generation, cache/refresh, buy/sell, repair, gambling, identify, heal, hireling, and related server-side service paths.

## D2MOO paths inspected for world/NPC/difficulty/endgame research

- `source/D2Game/include/OBJECTS/ObjMode.h`
- `source/D2Game/src/OBJECTS/ObjMode.cpp`
- `source/D2Game/src/OBJECTS/Objects.cpp`
- `source/D2Game/include/OBJECTS/ObjRgn.h`
- `source/D2Game/src/OBJECTS/ObjRgn.cpp`
- `source/D2Common/include/DataTbls/ObjectsTbls.h`
- `source/D2Common/src/DataTbls/ObjectsTbls.cpp`
- `source/D2Common/include/DataTbls/LevelsTbls.h`
- `source/D2Common/src/DataTbls/LevelsTbls.cpp`
- `source/D2Game/src/UNIT/SUnitNpc.cpp`
- `source/D2Game/src/QUESTS/ACT5/A5Q6.cpp`
- difficulty/state structures in `source/D2Common/include/D2DataTbls.h`

### Object and transition claims supported

The object runtime exposes data-selected operation families for doors, stairs, teleport pads, chests, traps, shrines, caskets/urns, corpses, racks, barrels, quest objects and other interactables. Object operation can affect state/mode, collision/selectability, scheduled events, loot/traps, quest behavior, transitions and sound cues. Dark Magic therefore normalizes authored behavior IDs into semantic handlers rather than object-ID switch logic.

The Act V Baal quest demonstrates durable per-difficulty progression plus authoritative dynamic portal creation. Later Uber-related Cube operations appear behind explicit version feature conditionals in D2MOO, supporting deliberate content-era gating rather than back-projecting later behavior into a 1.10f baseline.

### Difficulty claims supported

`D2DifficultyLevelsTxt` exposes resistance and death-XP penalties, monster skill/control-duration modifiers, life/mana steal divisors, champion/unique damage bonuses, Static Field minimum, and gambling odds. Difficulty therefore belongs in shared session rules consumed by multiple systems, not scattered local switches.

## D2MOO paths inspected for multiplayer/party/trade research

- `source/D2Game/include/UNIT/Party.h`
- `source/D2Game/src/UNIT/Party.cpp`
- `source/D2Game/src/PLAYER/PartyScreen.cpp`
- `source/D2Game/include/PLAYER/PlrTrade.h`
- `source/D2Game/src/PLAYER/PlrTrade.cpp`
- related player-list/friendly/unit routines reached by party-screen behavior

The party runtime supports explicit party identity/membership, same-level living-party iteration/counting, synchronization, and gold sharing. Party state is therefore gameplay state and participates in loot/reward behavior rather than being a client-only roster grouping.

`PartyScreen.cpp` supports separate Hardcore corpse-loot permission, invitation/alliance relationships, ignore/squelch relationships, and server-side hostility. In the inspected 1.10f path hostility requires the declarer in town and both players at least level 9, propagates against a target party, can force the declarer out of a shared party, and uses a legacy 60-second host-clock cooldown. Dark Magic should preserve supported semantics with authoritative deterministic timing rather than importing `GetTickCount()` into replayable simulation state.

`PlrTrade.cpp` exposes player trade as an explicit server state machine. Dark Magic research therefore models trade as server-owned escrow/offer state with atomic commit rather than two independent client inventory moves.

## Gameplay audio evidence inspected

### Dark Magic current audio baseline

- `internal/audio/audio.go`
- `internal/audio/catalog.go`
- `internal/game/data/model/sound.go`
- `internal/runtime/lua/audio_module.go`

Dark Magic already has checked sound handles, owner-thread backend commands, named buses (`music`, `ui`, `sfx`, `ambience`, `speech`, `cinematic`), groups, fades, streaming and a typed `Sounds.txt` model. The catalog currently consumes redirects/group weighting, volume range, routing, loop and stream behavior.

The typed `Sounds.txt` model also preserves fields not yet carried into playback: pitch range, Fade In/Out, Defer/Stop Instance, Duration, Compound, Falloff, LFE mix, 3D spread, Priority, Is2D, Tracking, Solo, Music Vol, Block1/2/3 and Delay. Their exact legacy semantics remain probes rather than assumed contracts.

### World/audio linkage

D2MOO `D2LevelsTxt` includes rain, inside/outside and `dwSoundEnv`; monster data contains sound-set references. This supports an authoritative world-context-to-soundscape boundary, but exact client-side sound-environment selection, attenuation, timing and mix behavior still need retail/client corroboration.

Audio playback/random variation is presentation state. It must not consume gameplay RNG streams or become authoritative gameplay timing.

## PvPGN realm responsibility corroboration

Inspected at `9cd173f4e02ba3d9f8f15a67ca308b5eb78723e4`:

- `conf/realm.conf.in`
- `conf/d2cs.conf.in`
- `conf/d2dbs.conf.in`
- repository README/license metadata

`realm.conf.in` describes a realm as the home for closed characters and games. `d2cs.conf.in` separates the Diablo-II control server from a list of game servers and from the Battle.net gateway, and includes character/game listing, creation and server-selection/queue policy. `d2dbs.conf.in` separately owns character-save/info directories and ladder persistence.

This independently corroborates that a useful Diablo-II-compatible service architecture separates realm/control-plane responsibilities, game workers and durable character storage. It does **not** prove Blizzard's private server topology or require Dark Magic to copy PvPGN protocols or internal structures.

Dark Magic's modern semantic split remains:

```text
realm/control plane
      -> game-worker allocation/admission
      -> transport-independent authoritative Session
      -> durable character store/leases/revisions
```

Original-client BNCS/MCP/D2GS compatibility, if pursued, belongs in protocol adapters around those semantic services.

## libd2 verification methodology

[libd2 `docs/VERIFICATION.md`](https://github.com/jaenster/libd2/blob/e6cdc4927c6180be8dd309b0423b470f64f1fc6c/docs/VERIFICATION.md)
reports that its DRLG checks compare against retail-engine captures rather than
self-generated expected values, including blind holdout seeds. It also reports:

- save testing against real `.d2s` files with byte-exact round trip;
- item validation against treasure/affix tables plus rolls traced through the binary's routine;
- pathfinding validation against generated maps and server movement gates;
- recovered game/network rules documented with binary routine addresses.

Dark Magic should emulate the **verification strategy**, not copy implementation structure.

## Independent `.d2s` source

[nokka/d2s](https://github.com/nokka/d2s) is MIT licensed and documents a
LoD-era `.d2s` parser. It is useful because it is independent of Dark Magic and
D2MOO. Claims from it remain **medium confidence** until matched against owned
saves or another strong source.

High-value leads include:

- 765-byte fixed header for its supported era;
- file-size/checksum fields;
- 16-byte name storage;
- status/progression/class/level/hotkey/weapon-set/appearance/difficulty/map/mercenary fields;
- quest, waypoint, and NPC-introduction blocks;
- bit-packed stat section terminated by stat ID `0x1ff`;
- 30-byte class skill levels;
- variable-length item lists plus corpse/mercenary/golem sections.

Do not copy its parser as the Dark Magic save codec. Build independent binary fixtures and preserve unsupported bytes.

## Source-conflict policy

When system sources disagree:

1. owned-game trace or bytes;
2. applicable contemporary Blizzard documentation for user-visible behavior;
3. D2MOO behavior for the explicitly labeled 1.10f target;
4. contemporary Data File Guide / Phrozen Keep research;
5. independent implementation with documented retail verification;
6. agreement among several independent sources;
7. one independent parser/engine or corroborated community reference;
8. an uncorroborated wiki/guide/forum lead;
9. inference.

Patch differences can make two sources simultaneously correct. Record the target version before declaring a conflict.

For service architecture, independent server implementations such as PvPGN can corroborate useful responsibility boundaries but rank below direct original-game/client/server observations for compatibility behavior.

Diablo Wiki/Fandom is an approved discovery resource, not a blanket authority.
Record the exact article and inspection date, follow its citations where possible,
and promote a claim only after target-version traces or stronger independent
evidence corroborate it. See [REALM_AND_GAME_CHAT_COMMANDS.md](REALM_AND_GAME_CHAT_COMMANDS.md)
for the first command-specific application of this policy.

## High-value evidence still to add

These are intentionally **not** claimed as validated behavior yet:

- repeatable retail traces for exact hit/block/avoidance/mitigation/PvP/death arithmetic;
- original-game timing captures for poison, regeneration, AI, missiles, auras and skill delay/interruption;
- complete legacy TXT->BIN compiler behavior;
- exact NoDrop/MF/Gold Find rounding and direct-vs-TC quality caller coverage across patches;
- exact runeword/socket/charm/container edge cases;
- full Cube output-operation coverage from owned data plus traces;
- retail vendor pricing/repair/gamble traces;
- exact object/shrine reset/math, waypoint bits and dynamic portal ownership/lifetime behavior;
- original player-death/corpse/XP-loss traces;
- client-side `Sounds.txt` pitch/fade/compound/falloff/tracking/solo/block semantics and `Levels.SoundEnv` behavior;
- party XP/loot sharing and hostility/PvP edge cases beyond inspected 1.10f server logic;
- exact trade UI/state timing and disconnect behavior;
- packet/protocol verification for BNCS/MCP/D2GS and original realm behavior;
- Resurrected-only data/protocol behavior;
- lossless `.d2s` round trips against a representative owned-save corpus.

Add exact versions, licenses, paths, captures, hashes and fixture descriptions here when those probes graduate.
