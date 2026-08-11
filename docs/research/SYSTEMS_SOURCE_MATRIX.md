# Diablo II gameplay-systems research source matrix

Status: living source inventory for `docs/research/` gameplay-system work.

The legacy binary-format source inventory remains in [../formats/SOURCE_MATRIX.md](../formats/SOURCE_MATRIX.md). This matrix adds gameplay, save, stat, item, timing, and verification sources without changing the format research evidence hierarchy.

## Evidence grades

- **A / verified**: lawfully owned original-game bytes or repeatable retail-engine trace.
- **B / high**: pinned reverse-engineered original-runtime behavior, preferably D2MOO, with clear routine/data relationship.
- **C / medium**: credible independent implementation, historical format documentation, or community research.
- **D / lead only**: current guide/wiki/forum/player report.

## Inspected sources

| Source | Version / commit | License | Grade / role | Material used in current foundation docs |
| --- | --- | --- | --- | --- |
| Dark Magic repository | current `main` at research branch creation | MIT project source | implementation baseline | existing game-data catalog, 25 Hz session/replay, persistence boundary, item authority, player admission, roadmap |
| ThePhrozenKeep/D2MOO | `3b21043b99e987bad41cf0f7b49f1f246db52d5c` | MIT | B; reconstructed 1.10f runtime | seed arithmetic, per-item seed/state, item fields, StatList/ItemStatCost-driven aggregation, quest runtime |
| jaenster/libd2 | `e6cdc4927c6180be8dd309b0423b470f64f1fc6c` | MIT for clean-room source; exclude embedded Blizzard-derived data | C corroboration with unusually strong verification methodology | retail-capture/holdout methodology; save byte-exact round trip; item rolls traced against binary |
| nokka/d2s | `master` inspected 2026-08-10 | MIT | C; independent save parser/document | `.d2s` header/section layout, bit-packed stats/skills/items, quest/waypoint/NPC blocks, corpse/merc/golem sections |
| bundled Diablo II Data File Guide | repository `docs/diabloiidatafileguide.html.gz` | historical community documentation | C semantic guide | TXT field meanings, formula/linkage leads; mounted game bytes remain stronger |
| Dark Magic recovered Riiablo data | pinned provenance recorded in `riiablo-recovered-data.md` | Apache-2.0 upstream | C declarative corroboration | quest/speech/DS1 object declarative mappings |

## D2MOO paths inspected for foundation research

- `source/D2Common/include/D2Seed.h`
- `source/D2Common/src/D2Seed.cpp`
- `source/D2Common/src/Items/Items.cpp`
- `source/D2Common/src/D2StatList.cpp`
- quest paths already listed in [../formats/SOURCE_MATRIX.md](../formats/SOURCE_MATRIX.md)

### Claims supported

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

## libd2 verification methodology

[libd2 `docs/VERIFICATION.md`](https://github.com/jaenster/libd2/blob/e6cdc4927c6180be8dd309b0423b470f64f1fc6c/docs/VERIFICATION.md)
reports that its DRLG checks compare against retail-engine captures rather than
self-generated expected values, including blind holdout seeds. It also reports:

- save testing against real `.d2s` files with byte-exact round trip;
- item validation against treasure/affix tables plus rolls traced through the
  binary's routine;
- recovered game/network rules documented with binary routine addresses.

Dark Magic should emulate the **verification strategy**, not copy implementation
structure.

## Independent `.d2s` source

[nokka/d2s](https://github.com/nokka/d2s) is MIT licensed and documents a
LoD-era `.d2s` parser. It is useful because it is independent of Dark Magic and
D2MOO. Claims from it remain **medium confidence** until matched against owned
saves or another strong source.

High-value leads include:

- 765-byte fixed header for its supported era;
- file-size/checksum fields;
- 16-byte name storage;
- status/progression/class/level/hotkey/weapon-set/appearance/difficulty/map/
  mercenary fields;
- quest, waypoint, and NPC-introduction blocks;
- bit-packed stat section terminated by stat ID `0x1ff`;
- 30-byte class skill levels;
- variable-length item lists plus corpse/mercenary/golem sections.

Do not copy its parser as the Dark Magic save codec. Build independent binary
fixtures and preserve unsupported bytes.

## Source-conflict policy

When system sources disagree:

1. owned-game trace or bytes;
2. D2MOO behavior for the explicitly labeled 1.10f target;
3. contemporary Data File Guide / Phrozen Keep research;
4. independent implementation with documented retail verification;
5. agreement among several independent sources;
6. one independent parser/engine;
7. inference.

Patch differences can make two sources simultaneously correct. Record the target
version before declaring a conflict.

## Sources to add in later workstreams

The following are intentionally not claimed as inspected foundation evidence yet:

- packet/protocol repositories for BNCS/MCP/D2GS;
- detailed vendor/economy and combat formula sources;
- Resurrected-only data/protocol sources;
- original-game timing traces for poison, regeneration, AI, missiles;
- complete legacy TXT->BIN compiler behavior;
- trade/hostility/party realm behavior.

Add them here only after exact versions/licenses/paths are recorded.
