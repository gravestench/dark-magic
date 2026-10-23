# Evidence: monster scaling by player count and difficulty

Spec: [Monster scaling by player count and difficulty](../../specs/monster-scaling-by-players-and-difficulty.md).
Artifacts: [original source manifest](../source-manifest.md).
Limits: [evidence gaps](../../plans/gaps/monster-scaling-by-players-and-difficulty.md).

## Identity and verification method

Evidence profile: Lord of Destruction Expansion 1.13c, Windows x86, little-endian. Artifact ids
`LOD113C-D2GAME` and `LOD113C-D2COMMON` resolve to the fingerprints in the source manifest. The
analysis images read on 2026-10-14 were re-hashed that day and match those fingerprints. RVAs
are relative to each module's image base. Export ordinals belong to the identified module. A
dash means no ordinal is recorded here. All function descriptions are analyst labels, not
original symbol names or required APIs.

Methods:

- 2026-10-12: disassembly readings of the creation, scaling and drop functions, with
  reconstructions checked against original bytes where they existed. These readings covered
  the preset build, appearance, property, equipment and post-creation chains.
- 2026-10-14: independent instruction listings decoded directly from the original images for
  the regions named below, read-only image data (constant tables, an initial variable value,
  float constants) and the column descriptors that the original table loaders build in code.
  Those descriptors bind each text column name to a record field; they are code evidence, not
  archive table reads. No archive table cell was read and no gameplay observation was made.
- Reused: the foundations companion's 2026-10-12/13 and 2026-10-14 readings of the life roll
  and life write requests (see [Reused evidence](#reused-evidence)).

No binaries, whole tables, private saves or raw captures are published here. Private
reconstruction paths, tool sessions and fixture encodings are kept in a separate private audit.

## Dispatch and function references

### Creation chain

Generic unit creation `D2Game.dll+0xE1D90` validates the class, calls the unit seeder
`0xE0110` (NRF-R12a) and dispatches unit type 1 through its type table to the monster
initializer `0xB0320`. The initializer calls the preset lookup `0xCA5F0` and passes its result
to stat initialization `0xAFDB0`. After the initializer returns, a set creation condition bit
leads to `0xE0F40` and, for monsters, `0xAFAB0` → `0x75E10` → `0x77240`.

Inside `0xAFDB0`, in order: MonStats2 row lookup `0x1320` (return without effect when absent);
appearance `0xAF730` (instruction `0xAFDFD`); class draw (`0xAFE1C..0xAFE6C`, `Rand` entry
`0x1180`); count setup `0xAF840` (`0xAFE7A`); difficulty clamp and `unitGate` override
(`0xAFE7F..0xAFEBA`); level (`0xAFED3`) and expansion area level (`0xAFEDB..0xAFF10`);
stat-list setup through thunk `0xADAE`; stat 12 and stat 100 requests (`0xAFF22..0xAFF3B`);
resistance, block and fixed stats (outside this spec); `sel` (`0xAFFFB..0xB0012`); base values
through thunk `0xAD7E` (`0xB0031`); life roll `0x11D0` (`0xB0044`), bonus `0x14D0`
(`0xB0058`), clamp (`0xB005F`), stat 7 and stat 6 requests (`0xB0075`, `0xB0080`); stat 31
(`0xB008F`); experience `0x14D0` (`0xB00A5`) and stat 13 (`0xB00B0`); life regeneration
(outside); inventory setup; skill loop `0xB01B0..0xB020D` through `0x9F9A0`; flag writes;
MonProp loop `0xB024B..0xB02F2`; equipment roll `0x92ED0` (`0xB0304`).

### Function table

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | `0xE1D90` | — | Generic unit creation | MSD-R01, MSD-R19 | Disassembly 2026-10-12; listing of seeding and post-init call 2026-10-14 |
| LOD113C-D2GAME | `0xE0110` | — | Unit seeder | MSD-R01 | Reused foundations reading (NRF-R03a, NRF-R12a) |
| LOD113C-D2GAME | `0xB0320` | — | Monster initializer | MSD-R02 | Disassembly 2026-10-12 |
| LOD113C-D2GAME | `0xCA5F0` | — | Preset lookup and registration | MSD-R02 | Disassembly 2026-10-12; class byte table at `0xC96F8` |
| LOD113C-D2GAME | `0xDF1A0` | — | Preset variant build | MSD-R02 | Disassembly 2026-10-12 |
| LOD113C-D2GAME | `0x1320` | — | MonStats2 row of a class | MSD-R02, MSD-R03 | Disassembly 2026-10-12; table binding 2026-10-14 |
| LOD113C-D2GAME | `0xAFDB0` | — | Monster stat initialization | MSD-R03 to MSD-R16 | Disassembly 2026-10-12; full independent listing 2026-10-14 |
| LOD113C-D2GAME | `0xAF730` | — | Appearance selection | MSD-R03 | Disassembly 2026-10-12 |
| LOD113C-D2GAME | `0x1180`, `0x11D0` | — | Positive-modulus draws | MSD-R04, MSD-R11, MSD-R23 | Reused foundations reading (NRF-R02) |
| LOD113C-D2GAME | `0xAF840` | — | Count, bonus and difficulty setup | MSD-R06, MSD-R14 | Independent listing 2026-10-14 |
| LOD113C-D2GAME | `0x379A0` | — | Effective player count | MSD-R05 | Independent listing 2026-10-14 (previously undecoded bytes) |
| LOD113C-D2GAME | `0x37810` | — | Count callback (alive test inline) | MSD-R05 | Disassembly 2026-10-14 |
| LOD113C-D2GAME | `0xDFC10` | — | Player-unit table iterator with state-7 skip | MSD-R05 | Disassembly and listing 2026-10-14 |
| LOD113C-D2GAME | `0x37430` | 10049 | Player-count override setter | MSD-R05 | Listing 2026-10-14; export table read |
| LOD113C-D2GAME | `0xAF5F0`, `0xAF5D0` | — | Life and experience percent readers | MSD-R06 | Listing 2026-10-14 |
| LOD113C-D2GAME | `0x14D0` | — | `mulDiv` | MSD-R11, MSD-R13 | Reused foundations reading (NRF-R05) |
| LOD113C-D2GAME | `0x19F0` | — | MonProp row of a class | MSD-R15 | Disassembly 2026-10-12; table binding 2026-10-14 |
| LOD113C-D2GAME | `0x92ED0` | — | Equipment roll | MSD-R16 | Disassembly 2026-10-12; independent listing 2026-10-14 |
| LOD113C-D2GAME | `0x1C80` | — | MonEquip first row of a class | MSD-R16 | Disassembly 2026-10-12; index preparation 2026-10-14 |
| LOD113C-D2GAME | `0xAFCE0`, `0x11980`, `0x11490` | — | Equipment item request and item creation path | MSD-R17 | Call sites located 2026-10-14 (`0x9301C`, `0xAFD3F`, `0x1151E`); paths not traced |
| LOD113C-D2GAME | `0x9F9A0` | — | Monster skill grant | MSD-R14 | Listing 2026-10-14 to the grant call; grant outside |
| LOD113C-D2GAME | `0x8F5C0` | — | Initialization callee | MSD-R18 | Disassembly 2026-10-12; no step |
| LOD113C-D2GAME | `0xE0F40`, `0xAFAB0`, `0x75E10` | — | Post-creation mode update route | MSD-R19 | Disassembly 2026-10-12 |
| LOD113C-D2GAME | `0x77240` | — | Monster damage setup | MSD-R19, MSD-R20 | Listing 2026-10-14 of `0x772BA..0x774D2`; draws reused from NRF-R04b |
| LOD113C-D2GAME | `0x770A0` | — | Damage player-count term | MSD-R20 | Full listing 2026-10-14 |
| LOD113C-D2GAME | `0x12380` | — | Treasure-class drop | MSD-R21 to MSD-R23 | Disassembly 2026-10-12; listing 2026-10-14 of `0x12380..0x12700` |
| LOD113C-D2GAME | `0x12D60` | — | Drop caller passing the dropping monster | MSD-R23 | Disassembly 2026-10-12 |
| LOD113C-D2GAME | `0x21260` | — | Owner of a non-player unit | MSD-R21 | Disassembly 2026-10-12 |
| LOD113C-D2GAME | `0x9C1F0`, `0x9C070`, `0x9C000` | — | Party count, same-level party iterator, alive-count callback | MSD-R21 | Listing 2026-10-14 |
| LOD113C-D2GAME | `0x9BEB0`, `0x9BD00` | — | Party id of a player; party record lookup | MSD-R21 | Listing 2026-10-14 of the returns used; party identity outside |
| LOD113C-D2GAME | `0xDFA90` | — | Alive test | MSD-R21 | Listing 2026-10-14 |
| LOD113C-D2GAME | `0xF5724` | — | Float-to-integer truncation | MSD-R22a | Listing 2026-10-14 |
| LOD113C-D2GAME | `0x2C2D0` | — | Game creation | MSD-Q01 | Disassembly 2026-10-14 of the selector writes |
| LOD113C-D2COMMON | `0x54A00` | 11089 | Base values from MonStats and MonLvl | MSD-R10 | Independent listing 2026-10-14 |
| LOD113C-D2COMMON | `0x11E0` | — | `mulDiv` copy | MSD-R10 | Listing 2026-10-14; instruction-identical to D2Game `0x14D0` apart from addresses |
| LOD113C-D2COMMON | `0x3C000` | 10826 | Level id of a room | MSD-R02, MSD-R08, MSD-R21 | Reconstruction checked against bytes |
| LOD113C-D2COMMON | `0x6E0C0` | 10894 | Area monster level | MSD-R08 | Full listing 2026-10-14 |
| LOD113C-D2COMMON | `0x2FE10` | 10331 | Room of a unit | MSD-R21 | Disassembly 2026-10-14 |
| LOD113C-D2COMMON | `0x33CD0` | 10494 | State test | MSD-R05 | Reused foundations reading (RLI-R03) |
| LOD113C-D2COMMON | `0x1A2D0` | 11104 | `unitGate` | MSD-R07 | Reused status-gate reading (CSA-R02) |
| LOD113C-D2COMMON | `0x603C0` | 10237 | DifficultyLevels row, index clamped | MSD-R06, MSD-R14 | Reconstruction checked against bytes |
| LOD113C-D2COMMON | `0x3A740` | 10887 | Replacing stat write | MSD-R09, MSD-R12, MSD-R13 | Reused foundations reading (RLI-Q03) |
| LOD113C-D2COMMON | `0x38B70` | 10973 | Stat read | MSD-R20, MSD-R21 | Reused foundations reading (RLI-R02a) |
| LOD113C-D2COMMON | `0x3AF00` | 11166 | Stat-list setup | MSD-R18 | Disassembly 2026-10-12; one indirect call unresolved |
| LOD113C-D2COMMON | `0x3A7B0` | 10188 | Damage-stat write used by the damage setup | MSD-R20 | Reconstruction checked against bytes; forwards to `0x39280`, not read |
| LOD113C-D2COMMON | `0x391F0` | 10830 | Damage-term skip test | MSD-R20 | Disassembly 2026-10-14; meaning outside (MSD-Q05) |
| LOD113C-D2COMMON | `0x1A3C0` | 10182 | Empty routine called after the skill loop | MSD-R14 | Reconstruction checked against bytes; no effect |
| LOD113C-D2COMMON | `0x55580` | — | MonLvl loader | MSD-R10 | Descriptor listing 2026-10-14 |
| LOD113C-D2COMMON | `0x6E130` | — | Levels loader | MSD-R08 | Descriptor listing 2026-10-14 |
| LOD113C-D2COMMON | `0x5A8A0` region | — | MonStats loader | MSD-R08, MSD-R10, MSD-R14, MSD-R15 | Descriptor listing 2026-10-14 |
| LOD113C-D2COMMON | `0x55A90` region | — | MonProp loader | MSD-R15 | Descriptor listing 2026-10-14 |
| LOD113C-D2COMMON | `0x56C10` region | — | MonStats2 loader | MSD-R02 | Descriptor listing 2026-10-14 |
| LOD113C-D2COMMON | `0x54EF0` | — | MonEquip loader and first-row index | MSD-R16 | Listing 2026-10-14 |

Imported calls from D2Game (thunk → module export): `0xA7F6` → D2Common #11104; `0xAD7E` →
#11089; `0xA556` → #10826; `0xADA8` → #10894; `0xA478` → #10331; `0xA68E` → #10494; `0xA7BA`
→ #10237; `0xA472` → #10887; `0xA42A` → #10973; `0xADAE` → #11166; `0xA6E2` → #10188; `0xA514`
→ #10830; `0xAD90` → #10182; `0xA7AE` → #11139 (`0x1E670`, inventory location query).

### Rationale by rule

- **MSD-R01.** `0xE1D90` calls `0xE0110` with the game seed (`Game +0xD0`); the foundations
  record fixes one step and the `{lo, 666}` initialization.
- **MSD-R02, MSD-R03.** The 2026-10-12 reading covers `0xCA5F0` (class byte table `0xC96F8`
  for classes 195..296; level record `[Game+0xF0 + 4*levelId]`; entry count byte `+0x12`, entries
  of 0x34 bytes; limit 13), `0xDF1A0` (request 3, counts at MonStats2 row bytes `+0x15+i`,
  variant-limit byte `+0x25`, `TotalPieces` byte `+0xEC > 2`) and `0xAF730` (preset count byte
  `+0x03`, variants at `preset + 4 + 16*roll`; 16-slot path counts through thunk `0xADA2`,
  D2Common #10305 `0x84C30`). The 2026-10-14 table binding shows the 0x134-byte rows at
  `DataTables +0xA90` are the MonStats2 rows loaded from `monstats2` and indexed by
  `MonStatsEx`; `+0xEC` is the `TotalPieces` column. Bytes `+0x15..+0x25` were not bound to a text column in this
  reading and are treated as prepared values (MSD-Q06).
- **MSD-R04.** `0xAFE1C..0xAFE6C`: class 0x137 (311) and 0x138 (312) take `Rand(4)` through
  `0x1180` on `Unit +0x20`; the result is stored to `MonsterData +0x0E`, and for 312 also
  `+0x0F`. An invalid class compares as -1 and takes no draw.
- **MSD-R05.** `0x379A0` calls the iterator `0xDFC10` with callback `0x37810`. The iterator
  walks the 128 buckets of the game's player-unit table at `Game +0x1120` (chain `+0xE4`) and
  skips a unit when #10494 reports state 7. The callback skips byte `Unit +0xC6` bit 0, type 0
  in mode 0 or 0x11 and type 1 in mode 0 or 0xC, then increments the count. Then `0x379A0`
  reads `Game +0x6A`; when that byte is in 1..3 (unsigned) it returns the larger of the count
  and the D2Game variable at `0x111C1C` (signed). The variable's image value is 0. Its only
  writer is export #10049 `0x37430`, which stores values 0..8 and ignores others; no D2Game call
  or absolute reference to it exists. Among the module files of the same installation, only
  `D2Client.dll` imports it, from the chat command that answers `players set to %d`, after
  clamping the typed number to 1..8. That client code is outside this boundary and is cited
  only as an importer.
- **MSD-R06.** `0xAF840` zeroes its five-dword output, requires a monster with MonStats row,
  and tests `Align` (`+0x4C`). Nonzero `Align` stores `n = 1` without calling the percent
  readers. Otherwise `n = max(1, E)` and `0xAF5F0`/`0xAF5D0` return the life and experience
  percents. Table images: experience percents at D2Game `0xFB5F0` and life percents at `0xFB614`
  both read `0, 0, 50, 100, 150, 200, 250, 300, 350`; for `n >= 9` life is `(n - 2) * 0x32` and
  experience `(5n + 0x82) * 2`. The function then calls #10237 with the game difficulty byte
  `Game +0x6D` (index clamped to the row range), rewrites `Game +0x6D` to 2 when it is 3 or
  more, and stores the row's `MonsterSkillBonus` (`+0x10`, GDA-R05) as output `+0x08`.
- **MSD-R07.** `0xAFE7F..0xAFEBA` clamps `Game +0x6D` into 0..2 and calls #11104 on the
  monster; a nonzero result sets the index to 0.
- **MSD-R08.** Level is the MonStats word `+0xAA + 2d`. With `Game +0x70` nonzero, `d > 0` and
  neither MonStats flag bit 2 (`noRatio`) nor bit 6 (`boss`) of the `+0x0C` flags (tested on
  the low byte with masks 4 and 0x40), it calls #10826 on the creation room argument and then
  #10894 with the level id, `d` and the expansion value. `0x6E0C0` returns 1 for `d >= 3` or a
  level id outside `1..LevelsCount-1`; otherwise it reads the Levels row (0x220 bytes at
  `DataTables +0xC58`) word `+0x16 + 2d` for a nonzero expansion argument and `+0x10 + 2d`
  otherwise. The Levels loader binds `MonLvl1`, `MonLvl2`, `MonLvl3` to `+0x10/+0x12/+0x14` and
  `MonLvl1Ex`, `MonLvl2Ex`, `MonLvl3Ex` to `+0x16/+0x18/+0x1A`, and stores its records at
  `DataTables +0xC58` with the count at `+0xC5C`.
- **MSD-R09.** Requests through #10887: stat 0x0C with the level, then stat 0x64 with
  `max(1, output +0x10)`.
- **MSD-R10.** `sel` is set when byte `Game +0x6A` or dword `Game +0x74` is nonzero. The call
  passes `(classId, sel, d, level, 7, buffer)`. `0x54A00` clamps the row index to
  `rowCount - 1`, rejects a negative index (buffer left zero), and reads MonLvl rows of 0x78
  bytes at `DataTables +0xB70` (count `+0xB74`). Mask bit 0 reads `MinHP`/`MaxHP`, bit 1 `AC`,
  bit 2 `Exp`, each word at `+0xB0/+0xB6/+0xBC/+0xD4 + 2d`; with flag mask 4 (`noRatio`) set the
  word is copied, else `0x11E0(word, cell, 100)`. Column dword index is `12 + 3*sel + d` for
  life, `3*sel + d` for armor class and `24 + 3*sel + d` for experience.
- **MSD-R11, MSD-R12.** `range = buffer[1] - buffer[0] + 1`; `0x11D0` on `Unit +0x20`;
  `hp += 0x14D0(eax = lifePct, edx = hp, ecx = 100)`; `cmp hp, 0x800000` with a signed branch;
  `shl 8`; #10887 stat 7 then stat 6, both layer 0. The foundations reading of the same slice is
  reused.
- **MSD-R13.** Stat 0x1F with `buffer[2]`; `0x14D0(eax = xpPct, edx = buffer[4], 100) +
  buffer[4]` to stat 0x0D.
- **MSD-R14.** Skill words at MonStats `+0x170 + 2i` (`Skill1..8`); level bytes `+0x198 + i`
  (`Sk1lvl..Sk8lvl`), read sign-extended; a negative word or level `<= 0` skips; otherwise the
  level plus output `+0x08` is passed to `0x9F9A0`, which grants only skills inside the Skills
  count. Then #10182 (empty) is called with the bonus.
- **MSD-R15.** `0x19F0` reads MonStats word `+0x1A` (`MonProp`) and returns the 0x138-byte row
  at `DataTables +0xC30`, bound to the `monprop` loader. The loop reads the property link at
  `+0x04 + 0x60d + 0x10i` and chance byte `+0x124 + 6d + i`. A negative link breaks the loop;
  chance 0 (`jbe`) applies; otherwise the inline step `lo * 0x6AC690C5 + hi`, unsigned `div 100`
  and a remainder test `< chance` apply. Application uses D2Common #10665 (outside).
- **MSD-R16.** `0x92ED0` (with the monster, its level and creation arguments `(game, 0, level,
  1)`): `0x1C80` returns the MonEquip row at the class's first-row index (`MonStats +0x2A`); the
  last argument requires the first row's `oninit` byte; `Unit +0x60` must be non-null. Rows of
  28 bytes: `monster` word `+0x00`, `level` `+0x02`, `oninit` `+0x04`, items `+0x08`, locations
  `+0x14`, mods `+0x17`. The skip loop and processing loop, the `k` count with signed `1..0x0A`,
  the inline step with mask or unsigned divide, the #11139 location test and the mod clamp are
  at `0x92F28..0x9302E`. A blank item code takes a further branch only when the second creation
  argument is nonzero; creation passes 0.
- **MSD-R17.** `0x92ED0` calls `0xAFCE0` at `0x9301C`; `0xAFCE0` calls `0x11980` at `0xAFD3F`;
  `0x11490` calls `0xE1D90` at `0x1151E`. The conditions on these paths were not read.
- **MSD-R18.** The 2026-10-12 reading found no LCG step in `0x1320` or `0x8F5C0`; the skill
  grant route reaches a step only at unit death; the D2Common callees reach none except one
  unresolved indirect call each in #11166 (`0x3AF00`) and `0x3A7D0`. The stat-list setup
  receives D2Game callback `0xD9470`, an item-event registration callback inspected on
  2026-10-14 without finding a draw; the indirect call target was not resolved.
- **MSD-R19.** `0xE1D90` calls `0xE0F40` when creation flag bit 0 is set; the 2026-10-12
  reading follows it to `0x77240`, whose element draws are NRF-R04b.
- **MSD-R20.** `0x77240` at `0x772DD..0x77309` calls #10830 on the monster; a nonzero result
  leaves the term 0. Otherwise it reads stat 100 through #10973, takes `max(1, value)` and
  calls `0x770A0`, which returns 0 when `Game +0x6D` is 0 or the count is below 2, the table at
  D2Game `0xFBBF0` (`0, 0, 8, 16, 24, 32, 40, 48, 56`) below 9, else `8n - 16`. At
  `0x77411..0x774B4`, a nonzero `Game +0x70` and nonzero term add `(v * term) / 128` (signed
  truncating idiom) to each of the three values; a zero expansion value takes the classic
  difficulty adjustment instead. Stats 0x15, 0x16, 0x13 are then requested through #10188.
- **MSD-R21.** `0x12380` arguments: game, dropping unit `[ebp+0x0C]`, killer `[ebp+0x10]`,
  treasure record `[ebp+0x14]`, bypass condition `[ebp+0x20]`. The prepared total is record
  `+0x0C` in an expansion game and `+0x08` otherwise; `Picks` `+0x10`; `NoDrop` `+0x14`. At
  `0x124B5..0x1252E`: a null killer, a non-player killer whose owner (`0x21260`) is null or not
  a player, or a party count `<= 1` gives 1; otherwise the count is capped at 8. `E` comes from
  `0x379A0`; `(E - elig)` uses `cdq`/`sub`/`sar 1`, a truncating halving. A type 1 dropper reads
  stat 100 with a lower bound of 1 and caps `eligible`. `0x9C070` gets the killer's room
  (#10331) and level id (#10826); with no room, no party data or party id 0xFFFF it runs the
  callback on the killer alone; a missing party record (`0x9BD00` null) returns without any
  callback; otherwise it looks up each member id in the player table and runs the callback when
  the member's room level id matches. The callback `0x9C000` counts members that pass
  `0xDFA90`.
- **MSD-R22a, MSD-R22b.** `0x12534..0x125FB`: `fild N`, `fild (N + T)`, a division (with the
  original floating-point division workaround switch at `0x119FAC`), a store to and reload from
  a double, repeated `fmul` (eight per unrolled iteration plus a remainder loop), then
  `1.0 - p` against constant `1.0` at `0xFD1A0`, an exact comparison with `0.0` at `0xFD198`,
  then `1.0 - q`, division by `q`, `fimul T` and `0xF5724`. `0xF5724` rounds with the current
  mode and corrects toward zero, returning the low 32 bits of the truncated value. The
  precision control word is not set in this path (MSD-Q02).
- **MSD-R23.** `0x125FF..0x12657`: a nonzero bypass draws `0x1180` with range `T` on
  `[ebp+0x0C] + 0x20` and skips the NoDrop test; otherwise the range is `T + N'`, a roll below
  `N'` decrements the pick counter and leaves item selection; otherwise `roll - N'` selects. A
  negative `Picks` branches to `0x1263F`, which performs no draw for that pick.

### Image data records

| Artifact | Location | Read | Supports |
|---|---|---|---|
| LOD113C-D2GAME | `0xFB5F0`, 9 dwords | `0, 0, 50, 100, 150, 200, 250, 300, 350` | MSD-R06 experience percents |
| LOD113C-D2GAME | `0xFB614`, 9 dwords | `0, 0, 50, 100, 150, 200, 250, 300, 350` | MSD-R06 life percents |
| LOD113C-D2GAME | `0xFBBF0`, 9 dwords | `0, 0, 8, 16, 24, 32, 40, 48, 56` | MSD-R20 |
| LOD113C-D2GAME | `0x111C1C`, 1 dword | 0 | MSD-R05 override at load |
| LOD113C-D2GAME | `0xFD198`, `0xFD1A0`, doubles | 0.0, 1.0 | MSD-R22a |

### Loader column records

Read 2026-10-14 from the original `D2Common.dll` by decoding the column descriptors each
loader builds before calling the shared table compiler. Each record names the column string,
its field offset and record size; they establish name-to-field binding, not the archive cells.

| Table (loader string) | Record size | Columns → field offsets | Supports |
|---|---|---|---|
| MonLvl (`monlvl`) | 0x78 | `AC`, `AC(N)`, `AC(H)` → 0x00/0x04/0x08; `L-AC` set → 0x0C..0x14; `TH` → 0x18..0x20; `L-TH` → 0x24..0x2C; `HP` → 0x30..0x38; `L-HP` → 0x3C..0x44; `DM` → 0x48..0x50; `L-DM` → 0x54..0x5C; `XP` → 0x60..0x68; `L-XP` → 0x6C..0x74 | MSD-R10 |
| Levels (`levels`) | 0x220 | `MonLvl1..3` → 0x10/0x12/0x14; `MonLvl1Ex..3Ex` → 0x16/0x18/0x1A | MSD-R08 |
| MonStats | 0x1A8 | `MonStatsEx` 0x18, `MonProp` 0x1A, flags dword 0x0C (`noRatio` bit 2, `boss` bit 6), `Align` 0x4C, `Level` 0xAA, `ToBlock` 0xA3, `MinHP` 0xB0, `MaxHP` 0xB6, `AC` 0xBC, `Exp` 0xD4, `Skill1` 0x170, `Skill2` 0x172, `Sk1lvl` 0x198 | MSD-R08, MSD-R10, MSD-R14, MSD-R15 |
| MonProp (`monprop`) | 0x138 | `prop1`, `par1`, `min1`, `max1` → 0x04..0x10, repeating by 0x10 to `prop6`; `(N)` block from 0x64 and `(H)` from 0xC4; `chance1..6` → 0x124..0x129, `(N)` 0x12A.., `(H)` 0x130..; stored at `DataTables +0xC30` | MSD-R15 |
| MonStats2 (`monstats2`) | 0x134 | `TotalPieces` → 0xEC; stored at `DataTables +0xA90` | MSD-R02 |
| MonEquip (`monequip`) | 0x1C | `monster` 0x00, `level` 0x02, `oninit` 0x04, `item1..3` 0x08/0x0C/0x10, `loc1..3` 0x14.., `mod1..3` 0x17..; stored at `DataTables +0xBE0` | MSD-R16 |

After loading MonEquip, the loader sets every MonStats word `+0x2A` to -1, then stores the
index of the first MonEquip row naming each class (`0x55143..0x551D9`). The `DataTables`
structure address is the value of D2Common `0x99E1C`; its `+0xC58`, `+0xC30`, `+0xA90` and
`+0xBE0` fields are written by the Levels, MonProp, MonStats2 and MonEquip loaders.

### Search records

- Override producer (2026-10-14): scanned the D2Game image for relative calls, jumps and
  absolute references to `0x37430` (none), read the D2Game export table (ordinal 10049), and
  scanned the import tables of the installation's module files (one importer, `D2Client.dll`).
- Stat-100 writers (2026-10-14): searched the D2Game disassembly for stat-id 100 pushed before
  the replacing-write thunk; the only stat-100 write found is creation's. Regions still held
  as undecoded bytes and other stat-write operations were not covered (MSD-Q03).
- Selector producers (2026-10-14): searched byte writes to `Game +0x6A` and found the game
  creation routine `0x2C2D0`, which stores a creation argument byte there, stores flag bit 21
  of its flags argument to `Game +0x74`, sets `Game +0x70` from flag bit 20 and stores the
  difficulty byte. The callers' meanings were not read (MSD-Q01).
- Table pointers (2026-10-14): searched the D2Common image for the absolute addresses of the
  `DataTables` fields and decoded the loader calls preceding each store.

### Corrections

- The damage player-count helper is D2Game `0x770A0`. An earlier reading gave it as `0x970A0`,
  an address inside another function; the call at `0x77304` targets `0x770A0`.
- The NoDrop conversion recomputes `p` as `1 - q` before dividing by `q`; an earlier
  formulation used `T * p / (1 - p)`, which is equal in real arithmetic but not as an
  operation sequence.
- The effective player count excludes players with state 7 or failing the alive test and
  includes the override; an earlier formulation described it only as the live player count.
- The base values use `mulDiv`; an earlier formulation gave a plain truncating product, equal
  for the original operand ranges but not in general.
- The area level reads Levels `MonLvl2Ex`/`MonLvl3Ex`; an earlier reading left it undecoded.
- Vector correction (2026-10-14, blind pass 1): the first draft of the damage-term vector gave
  term 16 for stat 100 = 4 and outputs 11, 22, 112. The helper `0x770A0` indexes the image table
  at `0xFBBF0` directly by the count (`[count*4 + table]`), and the table reads
  `0, 0, 8, 16, 24, ...`, so index 4 is 24 and the outputs are 11, 23, 118. The rule text was
  correct; the hand computation used index 3.
- The party iterator's behavior for a missing party record differs from the text of PXL-R04a,
  which states that the callback then runs on the credited player alone; the listing of
  `0x9C070` returns without a callback. For MSD-R21 both give `elig = 1`. The discrepancy is
  reported for the owner of PXL-R04a and is not settled here.

## Reused evidence

| Claim | Source | Original date and method | Scope reused |
|---|---|---|---|
| Life roll owner and form `min + Rand(max - min + 1)`, bonus `mulDiv(bonusPct, hp, 100)` | [foundations companion](core-numeric-and-rng-foundations.md#reached-draw-site-ownership), row `0xAFDB0`, NRF-R04a | 2026-10-12/13 disassembly; 2026-10-14 numeric pass | MSD-R11 seed owner and operand order |
| Clamp `>= 0x800000` to `0x7FFFFF`, `hp << 8`, stat 7 then 6 through #10887 | [foundations companion](core-numeric-and-rng-foundations.md#fixed-point-time-movement-and-object-animation), record retained there for MSD-R12 | Same | MSD-R12 arithmetic and order; stored-write integration stays RLI-Q03 |
| `mulDiv` branches and operand order | Foundations, NRF-R05; pipeline arithmetic audit 2026-10-13 | Disassembly | MSD-R10, MSD-R11, MSD-R13 |
| Draw forms, seeder, `HasState`, `unitGate`, stat reads | Foundations NRF-R02, NRF-R03a, RLI-R03, RLI-R02a, NRF-R12a/NRF-R12b; CSA-R02 | 2026-10-12 to 2026-10-14 | MSD-R01, MSD-R04, MSD-R05, MSD-R07, MSD-R20, MSD-R21 |
| Preset build, appearance, property, equipment, other callees, post-init route | This spec's 2026-10-12 readings | Disassembly | MSD-R02, MSD-R03, MSD-R15 to MSD-R19; the 2026-10-14 listings re-read `0xAFDB0` and `0x92ED0` in full |

The 2026-10-14 listing of the life slice in `0xAFDB0` matches the reused record; it adds no
new life behavior.

## Observation procedures

None of these were performed. Use a vanilla 1.13c Lord of Destruction game, a fixed monster
type in a fixed area, and a reader for the monster's stats and seed.

- **MSD-O01.** In each difficulty, with 1, 2 and 8 living players (or the single-player
  player-count command where `game.selectorByte` permits it), create the monster and read stats
  7 and 6 before any damage, together with the supplied `MinHP`/`MaxHP`, MonLvl cells, level and
  seed at creation. Expect MSD-R10 to MSD-R12.
- **MSD-O02.** Read stat 13 of the same monster at counts 1 and 8; expect MSD-R13.
- **MSD-O03.** Kill monsters dropping one fixed treasure class many times at counts 1, 2 and 8,
  with party members beside the kill, far away on the same level, on another level and dead.
  Record NoDrop frequency per pick and compare with MSD-R21 to MSD-R23.
- **MSD-O04.** Kill the monster with a summoned minion and trace the killer passed to the drop
  routine and the resulting `elig`; expect the owner's party.
- **MSD-O05.** Create a monster, then let a player join or leave; read stats 100, 7, 6 and 13
  before and after.
- **MSD-O06.** If a party larger than the effective count can be formed, kill a monster and
  trace `eligible`; expect the negative-term result.

## Runtime records

None. No vector has `runtime_observed: yes`.

## Status ledger and caveats

| Claim | Label |
|---|---|
| Creation seeding, preset build, appearance, class draw | verified in 1.13c code (static) |
| Effective count with state-7 exclusion, alive test and override | verified in 1.13c code (static) |
| Percents, `Align`, difficulty index, level, area level, base values | verified in 1.13c code (static) |
| Life roll, bonus, clamp, write requests; armor class and experience requests | verified in 1.13c code (static); stored-write integration RLI-Q03 |
| Skill-level bonus, property chances, equipment roll | verified in 1.13c code (static) |
| Equipment item seeding; other callees; post-creation draws | asm-only |
| Damage term and expansion application | verified in 1.13c code (static); inputs and skip condition MSD-Q05 |
| Eligible count, conversion sequence, pick draw | verified in 1.13c code (static) |
| Intermediate rounding precision of the conversion | unresolved |

Caveats: many function names attached to these addresses in analysis databases describe
unrelated behavior and are not used. Several 2026-10-12 readings rest on disassembly in which
some instructions were held as undecoded bytes; the 2026-10-14 listings decoded the regions
cited above directly from the original images.

## Migration record (2026-10-14)

Checks run: `tools/check_spec.py` on this spec, companion and gap file (0 errors, 0 warnings
after pass-1 triage); self-review against checklist.md; blind-implementer review complete after full runs 1 to 3 and
delta checks 1 and 2 (records below); the final delta check returned MISMATCH 0, GUESS 0, GAP 0
and DISCRIM 0. Status stays draft because open questions and gameplay observations remain.

Gaps closed with new evidence (2026-10-14 listings and loader descriptors):

- Effective player count: the callback test, the state-7 skip and the override (with its
  producer) are read; the rule is verified.
- Expansion area level: decoded to Levels `MonLvl2Ex`/`MonLvl3Ex`; verified.
- MonLvl column names and the `sel` column choice: bound by the MonLvl loader.
- Table names of the property, preset and equipment records: MonProp, MonStats2 and MonEquip.
- Party "same level": the iterator compares room level ids; verified.
- NoDrop bypass and negative `Picks`: behavior read; verified.
- Damage term: helper corrected and its application read; verified with supplied inputs.
- `MonsterSkillBonus` consumer: monster skill levels; verified.
- Normal-override predicate: consumed from CSA-R02 instead of restated.

Gaps left open: MSD-Q01 to MSD-Q09, listed in the
[gap file](../../plans/gaps/monster-scaling-by-players-and-difficulty.md).

Label changes: effective-count override from asm-only to verified; area level from asm-only to
verified; same-level test from asm-only to verified; damage term from asm-only to verified;
skill bonus from asm-only to verified; NoDrop bypass and negative picks from asm-only to
verified; property table layout merged into the verified property rule; equipment item
seeding from recovered; unverified to asm-only; x87 sequence split into a verified operation
sequence and an unresolved precision question; the gold and magic-find claim, the meaning of
the selector fields, the post-spawn recompute and the damage combination moved from rules to
open questions. The difficulty-independence statement is folded into the percent rules and
branches. The life write moved into this spec from the foundations spec with its reused
evidence; NRF-R08a remains its format contract.

Old ids were renumbered in execution order; references in
[combat-crushing-blow](../../specs/combat-crushing-blow.md) and its companion were updated.
`docs/specs/coverage.md` still cites two ids that no longer exist and needs an index update.

## Blind-implementer pass 1 (2026-10-14)

Fresh reader on a smaller model, given the spec, the foundations and game-data specs and the
cited dependency specs. Counts: MISMATCH 1, GUESS 11, GAP 11, DISCRIM 11, GATED 21, COVERAGE 23
(advisory), HARDEST 3. Outcomes:

- MISMATCH 1 (damage-term vector): spec edit; vector corrected to term 24 and 11, 23, 118 with
  the evidence under Corrections; a case with `damageTermSkip` added.
- GUESS 1 (no inventory): spec edit; `monster.inventory` named in section 3 and the rule.
- GUESS 2 (end of MonEquip): spec edit; the walk has no count bound (established by the
  listing); the contract covers terminated rows and MSD-Q10 records the storage question.
- GUESS 3-4 (conversion domain): spec edit; covered domain `N > 0`, `T > 0`, no overflow;
  zero divisor and out-of-range truncation added to MSD-Q02; vectors stay in the domain.
- GUESS 5 (player unit type): spec edit; type 0 in section 3 and MSD-V05.
- GUESS 6 (rewrite referent): spec edit; `game.difficulty` is set to 2.
- GUESS 7 (damage units): spec edit; units supplied by the damage setup, MSD-Q05.
- GUESS 8 (party enumeration): spec edit; PXL-R04a added to Depends on with the consumed
  enumeration; the missing-record discrepancy stays a reported finding.
- GUESS 9 (chance of 100 or more): spec edit; every nonzero chance draws.
- GUESS 10 (entry lookup): spec edit; entries 0..entryCount-1 in registration order, first
  stored class equal to `classId`.
- GUESS 11 (per pick or per drop): spec edit; once per pick in MSD-R21 and MSD-R23 (the
  eligibility block sits inside the pick loop of `0x12380`).
- GAP 1 (V11(c) percent): spec edit; `lifePct 0`.
- GAP 2 (preset inputs): spec edit; class, room level id, record entries and MonStats2 row given.
- GAP 3 (bonus rows): spec edit; Normal 0, Nightmare 2, Hell 3.
- GAP 4 (rows 2-4): spec edit; `oninit` and item identities given.
- GAP 5 (Picks and N): spec edit; `Picks = 1`, `N = 100`, and the source of each `N'`.
- GAP 6 (intermediates outside section 5): no change; template.md allows named intermediates
  in Outputs and keeps section 5 to writes.
- GAP 7 (creation condition): spec edit; `creationCondition` in section 3.
- GAP 8: same outcome as GUESS 8.
- GAP 9 (override supply): spec edit; one process-wide value read at each evaluation.
- GAP 10 (MonLvlEx mapping): spec edit; `d = 1` → `MonLvl2Ex`, `d = 2` → `MonLvl3Ex`;
  `MonLvl1Ex` removed from section 3.
- GAP 11 (level signedness): spec edit; signed comparison of the signed row level.
- DISCRIM 1-2 (preset build): spec edit; MSD-V02 (d) exhaustion with a kept duplicate, (e)
  duplicate re-rolls, (f) `k = 3` removal by replacement; each case lists its own state.
- DISCRIM 3 (override): spec edit; MSD-V05 (e) and (f).
- DISCRIM 4 (level): spec edit; MSD-V08 (f) to (i).
- DISCRIM 5 (base values): spec edit; MSD-V10 (e) and (f) operand-order case.
- DISCRIM 6 (life bonus order): spec edit; MSD-V11 (f).
- DISCRIM 7 (chance boundary): spec edit; MSD-V15 (b).
- DISCRIM 8 (later rows): spec edit; MSD-V16 row R5.
- DISCRIM 9 (conversion with `N != T`): spec edit; MSD-V19 (b), values checked to truncate the
  same in single and double precision.
- DISCRIM 10 (roll equal to `N'`): spec edit; MSD-V20 (e).
- DISCRIM 11 (difficulty 4): spec edit; MSD-V06 (h) and MSD-V07 (d).
- HARDEST 1-3: resolved by the entries above.
- COVERAGE 23: advisory; branches added where a result changes (listed DISCRIM cases);
  the rest unchanged.
- GATED 21: documented gates, no change.

A fresh reader must re-run because vectors, inputs and expected outputs changed.

## Blind-implementer pass 2 (2026-10-14)

Fresh reader. Counts: MISMATCH 0, GUESS 6, GAP 6, DISCRIM 13, GATED 16, COVERAGE 34 (advisory),
HARDEST 3. Outcomes:

- GAP 1 (`unitGate` owner): already re-pointed to CSA-R02 in section 3, MSD-R07 and Depends on
  by the foundations split; no further change. The class list in MSD-R07 is a citation of
  CSA-R02's values, kept for readability.
- GUESS 1 (exponent): spec edit; `p = r^eligible` stated.
- GUESS 2 (appearance values): spec edit; the class value overwrites appearance byte 10 (and 11
  for class 312) of the 16 bytes, as the stores to `MonsterData +0x0E/+0x0F` inside the
  `+0x04..+0x13` appearance block show; MSD-V04 asserts it.
- GUESS 3 (unregistered slot): spec edit; `entryCount` unchanged and later monsters repeat the
  lookup. The slot's variant count depends on storage initialization not read; new MSD-Q11
  (behavior) gates it; MSD-V02 (g) supplies 0.
- GUESS 4 (writes when the term is not applied): spec edit; the three writes are always
  requested (both branches join before the writes at `0x774B6`); the classic input values are
  defined as post-adjustment values; MSD-V17 outputs say "logged".
- GUESS 5 (killer membership): spec edit; MSD-V18 preamble.
- GUESS 6 (logged arguments): spec edit; the vector preamble lists the asserted arguments.
- GAP 2 (owner): spec edit; `owner(k)` row in section 3.
- GAP 3 (party inputs): spec edit; `k.partyId` and `game.partyRecord` in PXL-R04a's terms; the
  PXL divergence stays reported.
- GAP 4 (rows 0 and 99): spec edit; nonzero cells in MSD-V10.
- GAP 5 (slots 5-6, blocks): spec edit; MSD-V15 lists all six slots and the other blocks.
- GAP 6 (preset record contents): spec edit; each MSD-V02 case states `entryCount`, order and
  stored classes.
- DISCRIM 1: spec edit; MSD-V09 (c), (d).
- DISCRIM 2: spec edit; new composite MSD-V21 with one seed, ordered draws, final seed and write
  log.
- DISCRIM 3: spec edit; MSD-V08 (j).
- DISCRIM 4: spec edit; MSD-V10 rows 0 and 99 nonzero, (d) shows no clamp to row 0.
- DISCRIM 5: spec edit; MSD-V10 (g).
- DISCRIM 6: spec edit; MSD-V10 (h).
- DISCRIM 7: spec edit; post-draw seeds in MSD-V11 (b), (d), (e), (f).
- DISCRIM 8: spec edit; MSD-V14 skills with level -2 and skill id 0.
- DISCRIM 9: spec edit; MSD-V15 (a) slots after the negative link, (c) the `(H)` block.
- DISCRIM 10: spec edit; MSD-V16 (b) with equal level, `k = 3`, mod 8 and an occupied location;
  a negative row level is recorded as not changing a result on these inputs.
- DISCRIM 11: spec edit; MSD-V17 (f).
- DISCRIM 12: spec edit; MSD-V18 cases for party 4 with `E` 1, dead and roomless members, and a
  player-owned summon.
- DISCRIM 13: spec edit; MSD-V05 (g) with type 1 and type 2 units.
- HARDEST 1-3: resolved by the GUESS 1, GAP 3 and GUESS 3 entries.
- COVERAGE 34: advisory; cases added only through the DISCRIM entries above; the rest unchanged.
- GATED 16: documented gates, no change.

A fresh reader must re-run because vectors, inputs and expected outputs changed.

## Blind-implementer pass 3 (2026-10-14, full run)

Fresh reader. Counts: MISMATCH 0, GUESS 3, GAP 7, DISCRIM 13, COVERAGE 42 (advisory).
Outcomes:

- GUESS 1 (invalid class): spec edit; MSD-R01 cites NRF-R12a: a rejected class creates no
  monster; MSD-R04's invalid-class sentence marked unreachable through generic creation.
- GUESS 2 / GAP 3 (composite inputs): spec edit; MSD-V21 supplies the room level id, MonLvl row
  count, DifficultyLevels rows and MonProp row validity.
- GUESS 3 (word signedness): spec edit; section 3 states that the GDA-R08a 16-bit words are read
  sign-extended (the `0x54A00` listing reads each with a signed word load); MSD-V10 (l) shows
  0xFFF6 read as -10.
- GAP 1 (player-unit row): spec edit; the section 3 row now says each entry's type is tested.
- GAP 2 (V02(g)): spec edit; `entryCount = 1`, `v = 1` (not read).
- GAP 4 (section 5 appearance): spec edit; bytes 10 and 11 overwritten by one drawn value.
- GAP 5 (V18 inputs): spec edit; `E` and the party size declared supplied outcomes of MSD-R05
  and PartyCount.
- GAP 6 (PXL filters): spec edit; MSD-R21 states that only PXL-R04a's enumeration stage
  applies. The NoDrop callback `0x9C000` performs only the alive test; the distance, zero
  position and member-limit filters belong to the experience callback `0xDAED0`. PXL-R04a's
  binary-location text is reported for follow-up, not edited here.
- GAP 7 (stat ids): spec edit; MSD-V17 names stats 21, 22, 19.
- DISCRIM 1: spec edit; MSD-V13 (c).
- DISCRIM 2: spec edit; MSD-V10 (i), (j), (k).
- DISCRIM 3: spec edit; MSD-V05 (h).
- DISCRIM 4: spec edit; MSD-V07 (e), (f).
- DISCRIM 5: spec edit; MSD-V08 (k) with Levels row 0 populated.
- DISCRIM 6: spec edit; MSD-V15 (d) with a high-bit low word and chance 200.
- DISCRIM 7: spec edit; MSD-V16 (c).
- DISCRIM 8: spec edit; MSD-V17 (g).
- DISCRIM 9: spec edit; MSD-V18 type 2, null and stat-100-zero droppers.
- DISCRIM 10: spec edit; MSD-V19 (d) (converting would give 59 in double precision).
- DISCRIM 11: spec edit; MSD-V20 killer and game seeds asserted unchanged.
- DISCRIM 12: spec edit; MSD-V14 `Skill7 = -2`.
- DISCRIM 13: no change; MSD-V02 (b) already rejects the pre-step low word.
- COVERAGE 42: advisory; cases added only through the DISCRIM entries above.

No rule behavior, ownership or gate boundary changed in this triage; the next run can be a
delta check of the changed rules, vectors and cases.

## Delta check 1 (2026-10-14)

Delta reader on the changed rules, vectors and cases of pass 3. Counts: MISMATCH 0,
UNRESOLVED 0, GUESS 0, GAP 0, DISCRIM 6. Outcomes:

- DISCRIM 1 (gate 4 override): spec edit; MSD-V07 (g), class 561 at difficulty 2.
- DISCRIM 2 (L- columns equal to the base cells): spec edit; MSD-V10 (j) uses distinct
  `L-HP(H) = 120` and `L-XP(H) = 250`, outputs 12, 8, 39001, 81917.
- DISCRIM 3 (sign extension of every word): spec edit; MSD-V10 (l) sets all four words to
  0xFFF6, outputs -10 each.
- DISCRIM 4 (state 7 in the party count): spec edit; MSD-R21 states that `alive` does not test
  state 7 (the party iterator `0x9C070` and callback `0x9C000` call only the alive test
  `0xDFA90`; only the player iterator `0xDFC10` of MSD-R05 tests state 7); MSD-V18 adds a
  state-7 member case.
- DISCRIM 5 (vacuous cap claim): spec edit; the claim is removed from MSD-V15 (d).
- DISCRIM 6 (rejected class): no new case; NRF-V16 (d) asserts that a monster of class -1 or
  of class MonStatsCount returns before any allocation or seeder request with `game.seed`
  unchanged. MSD-R01 now cites it.

## Delta check 2 (2026-10-14)

Delta reader on the seven changes of delta check 1. Counts: MISMATCH 0, UNRESOLVED 0, GUESS 0,
GAP 0, DISCRIM 0. All seven changes were confirmed, including NRF-V16 (d) as the vector for
MSD-R01's rejected-class sentence. The blind-implementer pass is complete for the current
contract.

## Focused refresh and evidence applicability (2026-10-15)

The existing 2026-10-12 and 2026-10-14 records were checked against the source manifest and
the current numeric-foundations, game-data-fields, status-application, timed-state and party
contracts. Their cited artifact identities, supplied-input boundaries and instruction coverage
remain applicable. This is reuse of those dated readings, not a new source or archive-table
reading; no gameplay observation was made. No evidence label was promoted and all eleven
questions remain open. All 24 leaf rule ids and labels and all 21 executable-vector blocks
were preserved. No implementation plan, code or test consumes these spec ids; stable ids are
retained because this focused refresh needs no renumbering.

The opening description no longer promises exact half-base increases or an exact full-game
multiplier: MSD-R11's operand-sensitive integer arithmetic, MSD-R12's clamp and MSD-R06's
alignment exception require the precise technical rules. The description also says same map
level rather than suggesting a proximity condition; MSD-R21 has no distance test. The
MonLvl/Levels/MonProp/MonEquip column record remains evidence of engine-defined loader
bindings, not a read of archive cell values. Fixture cells remain supplied values.

Input and result inventories now identify existing preset storage, record counts, creation
eligibility and named stage results. NRF-R12a rejects an absent or ineligible monster record
as well as an out-of-range class before allocation/seeding. The successful creation fixture
supplies that eligibility separately from the post-initialization mode-update condition.
The three-row DifficultyLevels fixture follows GDA-R05; its unused bonus cells are explicitly
supplied as zero. The dependency-operation fixture remains a supplied-input boundary: its
persisted stat values do not verify the original replacing-write integration of RLI-Q03.

MSD-O05 previously said to expect no change after a player joins or leaves, although MSD-Q03
leaves recomputation unestablished. It now asks whether values change, with no predicted
result. This corrects the observation assertion without claiming a source result. Acceptance
also explicitly retains MSD-Q10 for actual end-of-storage equipment rows and MSD-Q11 for
the initial variant count of real unregistered preset storage; neither gates the terminated
equipment rows or supplied zero variant count used by the executable fixtures.

| Question | Applicable evidence and remaining boundary |
|---|---|
| MSD-Q01 | Existing selector writes and alive-test decisions are sufficient for supplied values. Neither the current shared contracts nor those producer slices establish the selectors' wider meaning or the unit-condition producer. No local assertion is gated. |
| MSD-Q02 | The existing conversion sequence and truncation helper remain applicable. Numeric foundations defines integer operations, not the hosting process's floating-point control state. No record closes precision-sensitive results or inputs outside the covered domain. |
| MSD-Q03 | Creation writes remain established; the limited historical stat-100 search is not exhaustive. No shared record establishes join/leave recomputation or its absence. MSD-O05 is an open observation. |
| MSD-Q04 | Current NRF-R12a/NRF-R12b verifies direct successful item seeding, and NRF-R04b verifies reached elemental draw sites. Neither reads the equipment request wrappers, resolves both initialization indirect calls, or establishes upstream mode-slot eligibility. Timed-state list contracts do not establish the targets of these creation-path indirect calls. Complete creation replay remains gated. |
| MSD-Q05 | Existing term arithmetic with supplied values remains available. Current damage-pipeline and reaction contracts do not supply this monster damage setup's base values, skip producer or stored effects of its distinct write operation. |
| MSD-Q06 | GDA-R08a establishes the MonStats-to-MonStats2 link; the existing loader record binds TotalPieces. Neither establishes preparation of the per-slot component counts or variant-limit byte. Supplied values remain available. |
| MSD-Q07 | No gameplay observation record exists; source and table identity checks do not replace the six pending procedures. |
| MSD-Q08 | The existing pick slice establishes the bypass's effect, not which callers choose it. A caller survey remains outside this supplied-condition contract. |
| MSD-Q09 | No applicable gold-amount or magic-find producer record was identified. Absence from the scaling slices is not a whole-program result. |
| MSD-Q10 | Existing equipment-loop evidence establishes the lack of a row-count bound. Current prepared-table contracts do not establish a following different-class sentinel for this storage. Supplied terminated rows remain available. |
| MSD-Q11 | Existing preset lookup and build evidence does not initialize the level record's entry storage. No shared creation-seeder contract establishes that allocation's variant count. A supplied zero count remains available. |

No identical source search was repeated solely because its prior attempt is older. The
PXL-R04a missing-party-record discrepancy recorded under Corrections remains a finding for
that contract's owner; for MSD-R21 both candidate party counts still reduce to elig = 1.
This refresh does not settle the party owner's behavior or duplicate its experience filters.

Mechanical check after this edit pass: check_spec.py on this spec, companion and gap file,
0 errors and 0 warnings. A fresh full blind-implementer review is required for the revised
input, acceptance and observation assertions; the historical passes above do not review them.

## Regeneration blind-implementer full pass 1 (2026-10-15)

A fresh smaller reader received the main contract and its cited dependency contracts, without
source code, private audits, runtime tools or previous review records. It recomputed all 21
executable-vector families. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 12,
COVERAGE 12 groups, HARDEST 3. The reader did not approve the spec.

MISMATCH, GUESS, GAP and DISCRIM have no entries to resolve. GATED 12: MSD-Q01 through
MSD-Q11 plus RLI-Q03 are already documented with their boundaries; no change. In particular,
MSD-R17 to MSD-R19 remain asm-only and MSD-R22b remains unresolved. A supplied-input
arithmetic or request assertion is not a complete creation replay or stored-write assertion.

Coverage dispositions:

1. MSD-R02 special classes, absent level record, full entry limit, absent MonStats2 record and
   no eligible component slots: no additional fixed-class or null-return variants are needed;
   MSD-V03 exercises the excluded-class fallback, MSD-V02(g) exercises an unregistered slot
   with supplied zero count, and the initializer's absent-record return is stated in section 6.
   A zero-component build has a distinct result even with a positive request; MSD-V02(h)
   adds a three-variant request that stores only one all-zero variant and takes only the
   appearance draw.
2. MSD-R02 existing variant count making req nonpositive: no change. The builder's explicit
   bound returns with no draws; MSD-V02(c) already tests returning an existing preset without
   rebuilding. Its variants are supplied storage, not a claim that a new registration calls
   the builder with nonzero initial count (MSD-Q11).
3. MSD-R05 state 7 on a monster entry: no change. MSD-V05 rejects a state-7 entry before its
   type-specific alive test and independently distinguishes the type-specific modes; another
   type using the same exclusion has no new count recipe.
4. MSD-R10 unavailable MonLvl row: no change. MSD-V10(d)/(g) establish the zero-base return,
   including its precedence over noRatio; absent storage uses the same result without a new
   calculation or a claim about its loader.
5. MSD-R12 hp exactly 0x7FFFFF: no change. Clamping that value to the same maximum and leaving
   it unchanged both produce raw 0x7FFFFF00. MSD-V12(c) already asserts that output for a value
   exceeding the maximum; an earlier clamp at the maximum is behaviorally equivalent there.
6. MSD-R16 no first row, oninit clear or absent inventory: MSD-V16(d) adds the first-row
   oninit guard with otherwise eligible equipment. No change for the other two predicates:
   they select the same explicit no-draw/no-request guard recipe without another calculation.
7. MSD-R20 p equal to 0 or 1 and the first positive term at 2: the first positive term is being
   added as MSD-V17(h). Values below 2 use the explicit zero-term recipe, whose unconditional
   writes are already asserted by MSD-V17(c)/(e); no additional zero-term cases are needed.
8. MSD-R21 party count above 8: MSD-V18 adds a party-10/nonmonster-dropper case, so
   the spawn-count cap cannot hide an incorrect party cap.
9. MSD-R21 null killer, no room, no party id or missing party record: no change. These branches
   normalize party counts of 0 or 1 to elig = 1, whose eligible result is already asserted in
   MSD-V18. The missing-record discrepancy with PXL-R04a remains explicitly reported rather
   than being treated as settled behavior of that owner.
10. MSD-R22a q equal to zero: no change. The direct branch specifies N' = 0; inputs that reach
    it through rounding remain precision-limited under MSD-Q02. No precision-dependent
    executable result is invented.
11. MSD-R23 repeated picks: no change. Each pick applies the same stage to the current
    dropper seed; MSD-V20 establishes the range, draw owner, outcomes and unaffected other
    seeds. The scope does not define the outer treasure selection or pick-loop bookkeeping.
12. MSD-R02 v above 2: no change. With zero existing count, MSD-V02(d) already demonstrates
    the three-variant request limit; a larger power-of-two bound leaves the same capped req.

Hardest dispositions, all no change:

1. MSD-R02/MSD-V02: the bounded duplicate attempts, slot-list replacement, stored duplicates
   and exact ordered seed states are explicit in MSD-V02(b)/(d)/(e)/(f).
2. MSD-R21/MSD-V18: owner replacement, same-level liveness, signed halving and the dropper cap
   are explicit in the rule and distinguished by the existing cases. The new party-cap case
   addresses the separate coverage advisory.
3. MSD-R22a/MSD-V19: exponent multiplication order, recomputation of 1 - q and truncation are
   explicit. MSD-Q02 retains intermediate precision limits; these cases do not settle them.

Primary review also corrected ownership and evidence wording: unitGate belongs to CSA-R02,
and local stat-write requests do not establish stored integration (RLI-Q03), update
bookkeeping (RLI-Q02) or client delivery. The previous client-update statements exceeded
the verified boundary. No new message or update behavior is asserted. These boundary wording
corrections and added fixtures require a fresh full second reader.

## Regeneration blind-implementer full pass 2 (2026-10-15)

A fresh smaller reader received the resulting main contract and its cited dependency
contracts, without source code, private audits, runtime tools or previous review records.
It recomputed all 21 executable-vector families, reporting 129 cases across the 20 verified
behavioral rules. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 14 groups,
COVERAGE 8 groups, HARDEST 3. This is a reader review, not user approval or gameplay evidence.

MISMATCH, GUESS, GAP and DISCRIM have no entries to resolve. All 14 GATED groups are already
documented, with no change: MSD-Q01 through MSD-Q11; MSD-R22b's precision limit; the
supplied-input or isolated boundaries of MSD-V01 through MSD-V21, which do not establish
all producers or stored-write integration; and complete creation replay under MSD-Q04/full
reproduction under MSD-Q07. These grouped limits overlap; they are not fourteen new questions.

Coverage dispositions, all no change:

1. MSD-R02 entry limit, missing MonStats2 and individual excluded classes: these take the
   explicitly stated return or unregistered-slot recipes. MSD-V02(g) and MSD-V03 already
   distinguish the supplied zero-count fallback and an excluded class. The initializer's
   missing-record return remains separate from an appearance-stage fixture; no complete
   creation assertion is inferred for absent records.
2. MSD-R06 n = 9: MSD-V06(f) exercises the formula branch and MSD-V06(c) the table branch.
   At n = 9 the two formula results are both 350; this introduces no new arithmetic recipe.
3. MSD-R07/MSD-R08 negative difficulty or area id: negative game difficulty is outside its
   u8 input domain. MSD-V08(k) already exercises the level-id <= 0 return of 1; a negative
   id uses the same bound and result. No invalid-difficulty fixture is added.
4. MSD-R15 invalid MonProp link: the initial record-bound guard explicitly ends the stage
   with no slot draw or application. No new numeric or selection recipe is needed; a missing
   record is not treated as a property application or a verified loader result.
5. MSD-R16 null inventory or absent first equipment row: both use the same no-draw/no-request
   entry-guard recipe now demonstrated by MSD-V16(d). Distinct downstream row selection,
   occupied locations and signed location/mod handling remain in MSD-V16(a) to MSD-V16(c).
6. MSD-R21 missing party record, no party id/player data or no room: these normalize party
   counts of 0 or 1 to elig = 1. MSD-V18 already demonstrates that result and its eligible
   calculation. The documented PXL-R04a discrepancy is not resolved by adding equivalent
   final eligible outputs.
7. MSD-R23 N = 0 or zero Picks: MSD-V19(c) establishes N' = 0 without conversion, and
   MSD-V20(c) already establishes the Rand(T) range and unchanged other seed owners. At
   N' = 0 no nonnegative roll can satisfy roll < N'. Zero Picks concerns how many times the
   outer treasure operation invokes this per-pick stage; its bookkeeping is outside scope.
8. MSD-R12 positive hp below the clamp: MSD-V12(a)/(b) already show positive unclamped
   shifted values. A value exactly equal to the maximum stores the same value whether a
   clamp branch is taken or not, as recorded in pass 1.

Hardest dispositions, all no change:

1. MSD-V02's eight cases: the rules and fixtures specify entry registration, preset sharing,
   bounded attempts, duplicate retention, slot replacement and each ordered draw.
2. MSD-V16's four cases: first-row guards, signed level comparison, per-row draws, location
   availability and mod bounds are explicit and yield exact request logs.
3. MSD-V18's seventeen cases: killer ownership, party membership, liveness, signed halving
   and both caps are specified; the final case prevents the monster cap from masking a
   wrong party cap.

Primary review corrected two statements without changing behavior, inputs, expected values,
assertions or gates. Related specs now attributes state primitives to numeric foundations
and unitGate to CSA-R02. MSD-V19(a)'s trace previously called the entire quotient exact;
only p and q are exact binary values for these inputs (eligible 2 gives the inexact quotient
100/3). The listed integer truncations remain stable across the stated single/double cases;
the trace now says so. MSD-Q02 still limits precision-sensitive results. These wording
corrections do not require another reader run. The final applicable full pass has MISMATCH 0
and no unresolved GUESS, GAP or DISCRIM finding. Document status remains draft.

## Specification review (2026-10-15)

An independent review of the contract against the spec-writing checklist re-read, from the
original-image disassembly, the percent readers `0xAF5F0`/`0xAF5D0` and their consumers in
`0xAFDB0` (life uses `0xAF5F0`, experience `0xAF5D0`), the image tables at `0xFB5F0`,
`0xFB614` and `0xFBBF0`, the effective-count routine `0x379A0` with callback `0x37810`, the
damage term `0x770A0` and its application at `0x77411..0x774B4`, the property loop in
`0xAFDB0`, and the drop routine `0x12380` from the killer/owner test through the pick draw.
All agreed with MSD-R05, MSD-R06, MSD-R12, MSD-R15, MSD-R20, MSD-R21, MSD-R22a and MSD-R23.
The override variable at `0x111C1C` lies in the zero-initialized tail of the data section
(beyond its raw size), consistent with the recorded initial value 0.

Cross-contract change: the missing-party-record discrepancy recorded under Corrections is
settled. The party contract's own 2026-10-15 correction now states that a missing record (and
an empty member list) returns without a callback, matching this spec's reading of `0x9C070`.
The Depends-on sentence that described a difference from PXL-R04a now states the agreement.
Section 3 also cross-references the party contract's name for the same unit condition
(`Unit +0xC6` bit 0). No rule, input, vector, label or gate changed, so no reader rerun is
needed for these reference edits.

Observed but not changed (reported as findings): when the prepared total `T` of a treasure
class is 0, `0x12380` leaves that treasure-class level at `0x1248D` before the NoDrop count or
any pick draw; MSD-R23 does not describe that branch.

## Issue #474 publication review (2026-10-23)

A fresh blind reader recomputed the supplied seed, life, experience, party-count and NoDrop
vectors without a mismatch. This review applies to the current V01–V21 text, including the
new V20(f) zero-total branch; it does not turn those supplied-input cases into gameplay
observations. The earlier 2026-10-15 passes remain their historical records.

The recorded original-image trace of D2Game `0x12380` exits this treasure-class level at
`0x1248D` when its prepared total `T` is zero, before `0x379A0` supplies the count to
NoDrop or a pick draws. MSD-R21/R23 and MSD-V20(f) now state this previously omitted
branch. Dependency references were also checked against the named NRF, RLI, CSA, GDA, GDF
and PXL rule owners; the GDA table-record rules were moved to their actual owner.

A separate loot-count audit rechecked the existing 1.13c static chain: `0x379A0` calls
`0xDFC10` with callback `0x37810`, whose iterator and callback exclude state 7 and dead
player modes from the effective count `E`; `0x12380` reads this function again for NoDrop.
The party bonus follows `0x9C070`/`0x9C000` and its own liveness test. These are distinct
inputs. The [pinned 1.10f D2MOO reconstruction](../SYSTEMS_SOURCE_MATRIX.md) corroborates the
living baseline through `PLAYER_GetPlayerCount` in
[`Player.cpp`](https://github.com/ThePhrozenKeep/D2MOO/blob/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Game/src/PLAYER/Player.cpp#L656-L677)
and the separate living same-level party count through
[`Party.cpp`](https://github.com/ThePhrozenKeep/D2MOO/blob/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Game/src/UNIT/Party.cpp#L262-L274).
This older reconstruction is not evidence of 1.14d behavior.
No 1.14d gameplay observation was made. MSD-O03 now specifies a three-player test that
distinguishes living count from connected occupancy; issue #373 tracks that parity check.

The fresh publication review also corrected the historical implementation path for
`systems/monster_death.lua` in the gap file. All eleven existing 1.13c evidence gaps remain
open at their stated boundaries.
