# Diablo II gameplay-systems research source matrix

Status: living source inventory for `docs/research/` gameplay-system work.

The legacy binary-format source inventory remains in [../formats/SOURCE_MATRIX.md](../formats/SOURCE_MATRIX.md). This matrix adds gameplay, save, stat, item, timing, combat, skill, AI, movement, and verification sources without changing the format research evidence hierarchy.

## Evidence grades

- **A / verified**: lawfully owned original-game bytes or repeatable retail-engine trace.
- **B / high**: pinned reverse-engineered original-runtime behavior, preferably D2MOO, with clear routine/data relationship.
- **C / medium**: credible independent implementation, historical format documentation, or community research.
- **D / lead only**: current guide/wiki/forum/player report.

## Inspected sources

| Source | Version / commit | License | Grade / role | Material used in current research docs |
| --- | --- | --- | --- | --- |
| Dark Magic repository | current `main` at each research-branch creation | MIT project source | implementation baseline | game-data catalog, 25 Hz session/replay, persistence, item authority, player/skills/movement/targeting, A*, generated world, transitions, roadmap |
| ThePhrozenKeep/D2MOO | `3b21043b99e987bad41cf0f7b49f1f246db52d5c` | MIT | B; reconstructed 1.10f runtime | seed arithmetic, items/stat lists, combat/damage, skill behavior dispatch, missiles, states, monsters/population, AI control, pet/hireling ownership, path types, inactive-room units, quests |
| jaenster/libd2 | `e6cdc4927c6180be8dd309b0423b470f64f1fc6c` | MIT for clean-room source; exclude embedded Blizzard-derived data | C corroboration with unusually strong verification methodology | retail-capture/holdout methodology; save byte-exact round trip; item rolls traced against binary; pathfinding checked against generated maps/server movement gates |
| nokka/d2s | `master` inspected 2026-08-10 | MIT | C; independent save parser/document | `.d2s` header/section layout, bit-packed stats/skills/items, quest/waypoint/NPC blocks, corpse/merc/golem sections |
| bundled Diablo II Data File Guide | repository `docs/diabloiidatafileguide.html.gz` | historical community documentation | C semantic guide | TXT field meanings, formula/linkage leads; mounted game bytes remain stronger |
| Dark Magic recovered Riiablo data | pinned provenance recorded in `riiablo-recovered-data.md` | Apache-2.0 upstream | C declarative corroboration | quest/speech/DS1 object declarative mappings |

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
- `source/D2Game/src/UNIT/SUnitInactive.cpp`

`Path.h` exposes 18 semantic path types, separate footprint/move-test collision masks, 16-bit fractional coordinates, velocity/acceleration state, target-unit state, collision patterns, and saved steps. `SUnitInactive.cpp` supports a real inactive-room unit archive with enough information to reconstruct monster identity, quality/modifiers, alignment, AI/minion state, XP, HP and special parameters rather than tying simulation existence to render residency.

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
2. D2MOO behavior for the explicitly labeled 1.10f target;
3. contemporary Data File Guide / Phrozen Keep research;
4. independent implementation with documented retail verification;
5. agreement among several independent sources;
6. one independent parser/engine;
7. inference.

Patch differences can make two sources simultaneously correct. Record the target version before declaring a conflict.

## Sources to add in later workstreams

The following are intentionally not claimed as inspected evidence yet:

- packet/protocol repositories for BNCS/MCP/D2GS;
- detailed vendor/economy traces;
- Resurrected-only data/protocol sources;
- original-game timing captures for poison, regeneration, AI, missiles and aura pulses;
- complete legacy TXT->BIN compiler behavior;
- trade/hostility/party realm behavior;
- exact original player-death/corpse/XP-loss traces.

Add them here only after exact versions/licenses/paths are recorded.
