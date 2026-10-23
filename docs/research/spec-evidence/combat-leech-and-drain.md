# Evidence: leech and drain

Spec: [Leech and Drain](../../specs/combat-leech-and-drain.md).
Artifacts: [original source manifest](../source-manifest.md).
Limits: [evidence gaps](../../plans/gaps/combat-leech-and-drain.md).

## Identity and verification method

Evidence profile: Lord of Destruction Expansion 1.13c, Windows x86, little-endian. Artifact ids
`LOD113C-D2GAME` and `LOD113C-D2COMMON` resolve to the fingerprints in the source manifest. RVAs are
module-relative; ordinals belong to the identified `D2Common.dll`. Function descriptions are analyst
labels, not original symbol names or required APIs.

Readings dated 2026-10-12 and 2026-10-13 are reused with their dates where noted. On 2026-10-14 the
leech operation was reread from entry to every return, and the following were read for the first
time from original instruction listings: the cap getters, the unit stat getter's guard sequence, the
overlay helper with its list search and update mark, the additive unit write wrapper and the
apparent-form helper. Byte runs without an instruction listing were decoded by hand from the
original bytes (opcode and operand decode); the decodes are named in the rationale. Table cells were
read on 2026-10-14 through the running original game's own archive loader (section
[Table records](#table-records)). No gameplay observation was made.

## Owning category

No function in this spec reaches a source-file assert. The leech operation `0xDBA40`, the capped adds
`0xDAE70` and `0xDB570`, the alive test `0xDFA90` and the mana drain `0xDA970` are called only from the
hit-application function D2Game `0xDE0C0` (the capped life add also from the same function for the
absorb heal), which the pipeline places in `combat` (`UNIT\SUnitDmg`). Category: `combat`.

<a id="dispatch-chain"></a>
## Dispatch chain

Reused from the 2026-10-13 reading and
[combat-damage-pipeline-order](../../specs/combat-damage-pipeline-order.md) DPO-R08d; the leech and
drain call sites in `0xDE0C0` were reread by the status-application work on 2026-10-14
([its chain](combat-status-application.md#dispatch-chain)).

1. Skill server do-function table (`SkillsRecord.srvdofunc`) at D2Game VA `0x6FD274A8`, index 1 ->
   D2Game `0xA2C40` -> build/enqueue `0xDDDE0` and dequeue/apply `0xDED30` (also through `0xA20C0`).
2. Missile `SrvHitFunc` table at VA `0x6FD2DA48` -> missile contact `0x3AE10` and callback route
   `0xA10B0`; both call `0xDE0C0`.
3. `0xDED30` -> hit-application `0xDE0C0`.
4. In `0xDE0C0`: life clamp of record `+0x08` against stat 6; leech `0xDBA40` (record in `eax`,
   defender in `ecx`, game in `edx`, attacker pushed); absorb heal through `0xDB570` (defender in
   `edi`, record `+0x48` pushed); total-damage life subtraction; mana drain `0xDA970` (defender in
   `ebx`, record `+0x3C` in `edi`); stamina drain `0xDA930`; then the status appliers (CSA-R01).
5. Inside `0xDBA40`: precondition, effectiveness, shift, routing; divisors through thunk `0xA7BA`;
   hireling test through thunk `0xA7F6`; apparent form through thunk `0xA7CC`; mana add `0xDAE70`
   and life add `0xDB570`; draw `0x1180`; overlay through thunk `0xA83E`; additive write through
   thunk `0xA4D2`; cap getter through thunk `0xA3A0`; stat reads through thunk `0xA42A`.

## Function references

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | `0xDBA40` | — | Leech operation | LSE-R01 to LSE-R08, LSE-R12, LSE-R14 | Disassembly 2026-10-13; full reread 2026-10-14 |
| LOD113C-D2GAME | `0xDAE70` | — | Capped mana add | LSE-R09 | Byte run decoded and checked against an independent listing 2026-10-13; decode rechecked 2026-10-14 |
| LOD113C-D2GAME | `0xDB570` | — | Capped life add with gates | LSE-R09, LSE-R11 | Disassembly with a decoded byte-run tail (checked 2026-10-13; rechecked 2026-10-14) |
| LOD113C-D2GAME | `0xDFA90` | — | Not-alive test | LSE-R11 | Disassembly 2026-10-13; reread 2026-10-14 |
| LOD113C-D2GAME | `0xDA970` | — | Defender mana drain | LSE-R15 | Disassembly 2026-10-13; reread 2026-10-14 |
| LOD113C-D2GAME | `0xDE0C0` | — | Hit application: leech placement | none here (placement owned by DPO-R14 to DPO-R18, CSA-R01; context for LSE-V13) | Disassembly 2026-10-13; call slice reread 2026-10-14 (status work) |
| LOD113C-D2GAME | `0xDC0B0` | — | Scale/mitigation entry writing the total | none here (owner DPO-R12a; context for LSE-V01) | Disassembly 2026-10-13 |
| LOD113C-D2GAME | `0x1180`, `0x14D0` | — | Draw, mulDiv | LSE-R08, LSE-R12 | Owned by NRF-R02, NRF-R05 |
| LOD113C-D2COMMON | `0x603C0` | 10237 | Difficulty record, clamped index | LSE-R05 | Reconstruction checked against bytes |
| LOD113C-D2COMMON | `0x1A2D0` | 11104 | Class-group selector | LSE-R04 | Owned by CSA-R02 |
| LOD113C-D2COMMON | `0x390A0` | 10186 | Mana cap: stat 9 read | LSE-R09, LSE-R10 | Full disassembly 2026-10-14 |
| LOD113C-D2COMMON | `0x390F0` | 10913 | Life cap: stat 7 read | LSE-R09, LSE-R10, LSE-R14 | Full disassembly 2026-10-14 |
| LOD113C-D2COMMON | `0x38B70` | 10973 | Unit stat read | LSE-R09, LSE-R14, LSE-R15 | Guard sequence read 2026-10-14; owned by RLI-R02a |
| LOD113C-D2COMMON | `0x38A80` | — | Selected-array getter | LSE-R10 | Owned by RLI-R02a, RLI-R02b |
| LOD113C-D2COMMON | `0x3A740` | 10887 | Unit replacing stat write | LSE-R09, LSE-R15 | Requested operation only; stored effect LSE-Q01; NRF-R08a owns the numeric format, not this write |
| LOD113C-D2COMMON | `0x3A710` | 10551 | Unit additive stat write | LSE-R14 | Reconstruction checked against bytes 2026-10-14; core `0x3A330` not read (LSE-Q01) |
| LOD113C-D2COMMON | `0x33CD0` | 10494 | State test | LSE-R11 | Owned by RLI-R03 |
| LOD113C-D2COMMON | `0x32C10` | 10689 | Overlay write | LSE-R13 | Full disassembly 2026-10-14 |
| LOD113C-D2COMMON | `0x37F40` | — | First list on the primary chain with given flags | LSE-R13 | Full disassembly 2026-10-14 |
| LOD113C-D2COMMON | `0x38100` | 11013 (called directly here) | New list | LSE-R13 | Owned by STL-R06a |
| LOD113C-D2COMMON | `0x3A7D0` | 10807 | Attach list | LSE-R13 | Dependency STL-R06b (gated) |
| LOD113C-D2COMMON | `0x3A280` | 10261 | List stat write | LSE-R13 | Dependency STL-R06c (gated) |
| LOD113C-D2COMMON | `0x665C0` | 10158 | Unit update mark | LSE-R13 | Full disassembly 2026-10-14; that attempt did not read room getter `0x2FE10` or establish the queue effect; current compatible room-getter evidence does not close LSE-Q01/RLI-Q02 |
| LOD113C-D2COMMON | `0x501E0` | 10350 | Apparent form (type, class, mode) | LSE-R14 | Full disassembly 2026-10-14 |
| LOD113C-D2COMMON | `0x4F780` | — | Monster mode fallback lookup | LSE-R14 (unused output) | Byte run hand-decoded 2026-10-14 |

Imported calls from D2Game go through thunks: difficulty record `0xA7BA` -> #10237, class group
`0xA7F6` -> #11104, apparent form `0xA7CC` -> #10350, life cap `0xA3A0` -> #10913, mana cap `0xA4EA`
-> #10186, unit stat read `0xA42A` -> #10973, unit replacing write `0xA472` -> #10887, unit additive
write `0xA4D2` -> #10551, state test `0xA68E` -> #10494, overlay `0xA83E` -> #10689. Thunk-to-ordinal
resolution reused from 2026-10-13.

## Semantic-to-original mapping

| Semantic term | Original mapping |
|---|---|
| `R.physical`, `R.fire`, `R.lightning`, `R.magic`, `R.cold`, `R.poison` | DamageRecord `+0x08`, `+0x10`, `+0x1C`, `+0x20`, `+0x24`, `+0x28` |
| `R.lifeLeech`, `R.manaBurn`, `R.staminaBurn` | `+0x38`, `+0x3C`, `+0x40` |
| `R.absorbHeal`, `R.total` | `+0x48`, `+0x4C` |
| unit type, class, id, mode | Unit `+0x00`, `+0x04`, `+0x0C`, `+0x10` |
| `unit.pool` | Unit `+0x08` |
| monster data and record | Unit `+0x14` -> MonsterData; MonsterData `+0x00` -> MonStats record |
| `Drain` columns | MonStats byte `+0xA0 + difficulty`, read with `movzx` |
| seed | Unit `+0x20`, `+0x24` |
| hit game `G` | `edx` argument of `0xDBA40`; difficulty byte Game `+0x6D` |
| `A.game` | Unit `+0x80` |
| `aliveExclusionCondition` | Unit byte `+0xC6` bit 0 |
| `disguiseCondition` | Unit `+0xC8` bit 3 |
| `formStateList` | DataTables word list at `+0x17C`, signed word count `+0x180` |
| `States.gfxtype` | States record byte `+0x2D` (stride `0x3C`, records DataTables `+0xBC`, count `+0xC4`; GDF field table) |
| state membership | StatList state bitfield `+0x58` through the bit-mask table; extended-list condition is the sign bit of StatList `+0x10` |
| primary attached-list chain, `overlayListCondition` | Extended StatList `+0x3C` head, successor `+0x2C`; list flags `+0x10` bit `0x80` |
| `overlayPendingCondition` | Base StatList `+0x10` bit `0x100` |
| `overlayCount` | DataTables `+0xBC0` |
| `NewList(..., conditions {overlayListCondition}, ...)` | flags argument `0x80` to `0x38100` (no expiry flag `2`) |
| difficulty outside 0..2 | #10237 clamps the index into `[0, count - 1]` (reconstruction checked against bytes); the drain byte read at `+0xA0 + difficulty` has no guard. Recorded as evidence only; the contract covers 0..2 (CSA-Q06) |
| DifficultyLevels divisors | record `+0x28` LifeStealDivisor, `+0x2C` ManaStealDivisor; stride `0x58` |

Native argument placement, which the named roles replace: leech `(eax = record, ecx = defender,
edx = game, stack = attacker)`, callee cleanup 4 bytes; capped adds `(edi = unit, stack = amount)`;
mana drain `(ebx = defender, edi = amount)`; draw `(ecx = seed address, eax = modulus)`; mulDiv
`(eax = a, edx = b, ecx = c)`; apparent form `(unit, &type, &class, &mode, file, line)` with 24 bytes
of callee cleanup; overlay `(unit, id, 0)`; additive write `(unit, stat, value, layer)`.

## Rule rationale

- **LSE-R01.** Entry loads `+0x38`; when zero, loads `+0x3C`; both zero jump to the common epilogue.
- **LSE-R02.** A null defender or a type other than 1 stores 100 in the slot later read as `P`. For a
  monster: null MonsterData or null record exits; otherwise `movzx` of the record byte at
  `+0xA0 + game difficulty` (game from `edx`); `test; jle` exits on 0.
- **LSE-R03.** `shl 6` of both fields, stored, before the attacker test. `0xDBA40` contains no
  access to `+0x4C` (`R.total`) on any path (2026-10-14 reread). The total itself is written by
  `0xDC0B0` (read 2026-10-13), owned by DPO-R12a.
- **LSE-R04.** Null attacker sets local type 6 and joins the apparent-form call with class -1 and mode
  0. Type 0 takes the divisor block. Type 1 calls #11104; nonzero jumps to the base block. Every other
  case joins the apparent-form call.
- **LSE-R05.** #10237 with the attacker game's difficulty byte; `+0x28` nonzero: `cdq; idiv` of `+0x38`,
  stored; `+0x2C` nonzero: same for `+0x3C`.
- **LSE-R07.** Base block loads `+0x08`, clears the mana flag register, and `jle` exits.
- **LSE-R08.** `+0x3C` nonzero: mulDiv(`+0x3C`, base, 100), then mulDiv(P, m, 100) when P is not 100;
  `cdq; and edx, 3Fh; add; sar 6` (signed division by 64); pushed to `0xDAE70`; flag set to 1. Then
  `+0x38` likewise into `0xDB570`.
- **LSE-R09.** `0xDAE70`: `jg` on the amount, else return 0; cap through #10186; current through
  #10973 stat 8 layer 0; `lea` sum; signed `jl` keeps the sum, else the cap; #10887 write stat 8 layer
  0; equal-to-cap returns `cap - cur`, else the amount. `0xDB570` is the same with #10913 and stat 6
  after its gates. Decodes rechecked 2026-10-14: `8D 04 2B` lea, `3B C6` cmp, `7C 02` jl, `8B DE` mov,
  the write call to thunk `0xA472`, `3B DE` / `74 0A` for the return choice.
- **LSE-R10.** `0x390A0`: null unit returns 0; null stat list returns 0; ItemStatCost count `<= 9`
  returns 0; record = records + `9 * 0x144` (null returns 0); then the selected-array getter
  `0x38A80` with key `0x90000` (stat 9, layer 0) and the unit's stat list. `0x390F0` is the same with
  `7 * 0x144` and key `0x70000`. `0x38B70` performs the same guards for a supplied id and layer before
  the same getter, so both caps equal `Stat(u, 9, 0)` and `Stat(u, 7, 0)` of RLI-R02a/RLI-R02b.
- **LSE-R11.** `0xDB570` after the amount test calls `0xDFA90` (unit in `eax`), whose nonzero result
  returns 0; then #10494 with state `0x5C`, nonzero returns 0. `0xDFA90` returns 1 for null, byte
  `+0xC6` bit 0, type 0 with mode 0 or 17, type 1 with mode 0 or 12, or any other type; else 0.
- **LSE-R12.** After the life add: with the flag set and `+0x38` nonzero, draw `0x1180` with modulus 2
  on the attacker seed; zero pushes `0x97`, nonzero `0x98`. Flag clear: `0x97`. `+0x38` zero: flag set
  pushes `0x98`, else the function returns. Each push is followed by #10689 `(attacker, id, 0)`.
- **LSE-R13.** `0x32C10`: `id < 0` or `id >= ` DataTables `+0xBC0` returns. With a stat list whose
  flags have the sign bit set, `0x37F40` (flags `0x80`) walks from the extended list's `+0x3C` head along
  `+0x2C` returning the first list with any requested flag bit; it makes no write. No list: `0x38100`
  with `(unit +0x08, 0x80, 0, unit type, unit id)`, then `0x3A7D0 (unit, list, 1)`. Then the unit's
  stat list (reloaded) gets `or [list+0x10], 0x100` when present; `0x3A280 (list, 0xB2, movzx id, 0)`;
  `0x665C0 (unit)`. `0x665C0` asserts on a null unit and otherwise queues the unit on its room's update
  list once, using `0x2FE10`.
- **LSE-R14.** After #10350 the local type is reloaded; 0 joins the base block. Otherwise #10973 reads
  defender stats 8 and 10 (layer 0); both fields are divided by 64 (`cdq; and 3Fh; add; sar 6`) and
  stored; signed `jl` selects `min(+0x38, +0x08)` and `min(+0x3C, defender mana)`; signed `jge` selects
  `min(+0x40, defender stamina)` with `+0x40` undivided; the sum's zero flag exits. mulDiv(P, total,
  100) when P is not 100. #10913 on the attacker minus #10973 stat 6; signed `jl` keeps the smaller;
  `jle` exits; #10551 `(attacker, 6, total, 0)`; #10689 `(attacker, 0x97, 0)`. #10551 (`0x3A710`) does
  nothing for a zero value and otherwise adds to the unit's stat list through `0x3A330`.
  `0x501E0`: null unit or Unit `+0xC8` bit 3 clear returns unchanged; a nonpositive list count returns
  unchanged; for each list entry in order, the entry is skipped unless the unit type is 0, 1 or 3, the
  state id is in `[0, state count)`, the stat list exists with its sign-bit condition and the state bit
  is set; the States record byte `+0x2D` equal to 1 writes type 1, class = word `+0x2E` and, for a
  player, mode through `0x4F780`; equal to 2 writes type 0, class = word `+0x2E` and, for a monster,
  mode through a constant table (class 6 in mode 4 gives 12); any other value continues. `0x4F780`
  (hand decode 2026-10-14) reads MonStats, MonStats2 and the bit-mask table, follows a mode fallback
  table until an available mode or mode 1, and writes only its return register and its own argument
  slot; it calls nothing and makes no draw. Leech reads only the type local after the call.
- **LSE-R15.** `0xDA970`: `jg` on the amount else return 0; #10973 stat 8; subtract; signed compare with
  `0x100`, `jge` keeps; #10887 write; `neg; sbb; and` returns the amount when the new value is
  nonzero, else 0.

## Corrections

Recorded 2026-10-14. Earlier inline evidence converted several analysis labels to module locations
by adding the label offset in the wrong direction, so these callee locations were 0x1000 too low.
The listings' labels are RVA minus 0x1000; the corrected RVAs are confirmed by the call targets of the
functions read on 2026-10-14:

| Stated earlier | Correct | Role |
|---|---|---|
| `0x37A80` | `0x38A80` | Selected-array getter reached by the cap getters |
| `0x36F40` | `0x37F40` | Overlay list search |
| `0x37100` | `0x38100` | New list in the overlay helper |
| `0x397D0` | `0x3A7D0` | Attach in the overlay helper |
| `0x39280` | `0x3A280` | List stat write in the overlay helper and the replacing write |
| `0x655C0` | `0x665C0` | Update mark |
| `0x39330` | `0x3A330` | Additive write core |
| `0x4E780` | `0x4F780` | Mode fallback lookup |
| `0x37550` | `0x38550` | Player call inside the replacing write ([crushing-blow evidence](combat-crushing-blow.md)) |

Effect on claims: the earlier question whether the cap reader sums attached lists rested on the
wrong location. The cap getters reach `0x38A80`, the getter of RLI-R02a/RLI-R02b, so the caps are
the unit reads (LSE-R11). No vector value changed.

Label changes from new evidence (2026-10-14): the cap sources (LSE-R11), the overlay write requests
(LSE-R14) and the monster-attacker branch (LSE-R15) moved from asm-only to verified in 1.13c code
(static): every local callee on their paths is now read, the stat reads are owned by RLI-R02a, and
the remaining stored effects are explicit integration gates (LSE-Q01). The branch's earlier
note that defender stat reads with a null defender were unread is settled by RLI-R02a (a null unit
reads 0). The divisor values (LSE-R07) are backed by a new table read; their earlier support was the
2026-10-12 live record of GDA-R05, kept in that companion.

<a id="table-records"></a>
## Table records

Read 2026-10-14 through the running vanilla LoD 1.13c member's Storm loader (member identity in the
private audit; build reported as the 1.13c LoD launch entry). Physical supplying archives, effective
file hashes and overrides were not captured. Headers were parsed by name; data-row indexing excludes
the header and begins at 0.

| Effective virtual path (`data/global/excel/`) | Range read / total bytes | Coverage | Row | Column | Raw cell | Supports |
|---|---|---|---|---|---|---|
| `difficultylevels.txt` | 0+586 / 586 | full file | `Normal` | `LifeStealDivisor` | `1` | LSE-R06 |
| `difficultylevels.txt` | same | full file | `Nightmare` | `LifeStealDivisor` | `2` | LSE-R06 |
| `difficultylevels.txt` | same | full file | `Hell` | `LifeStealDivisor` | `3` | LSE-R06 |
| `difficultylevels.txt` | same | full file | `Normal` | `ManaStealDivisor` | `1` | LSE-R06 |
| `difficultylevels.txt` | same | full file | `Nightmare` | `ManaStealDivisor` | `2` | LSE-R06 |
| `difficultylevels.txt` | same | full file | `Hell` | `ManaStealDivisor` | `3` | LSE-R06 |
| `itemstatcost.txt` | 0+2600 / 40109 | header and rows 0 to 19 | `Stat = maxmana` | `ID` | `9` | stat 9 name (LSE-R10) |
| `states.txt` | 9000+3000 / 20920 | rows 70 to 99 within the range; header from the full read recorded in [game-data evidence](core-game-data-fields.md) | `state = death_delay` | `id` | `92` | state 92 name (LSE-R11) |

Interpretation: decimal text to i32. Stat 178 `unit_dooverlay` reuses the
[crushing-blow table record](combat-crushing-blow.md). Stats 6, 7, 8, 10 and state names other than
92 reuse GDF-R02 and GDF-R03. These reads are archive contents only; the code consumers are cited in
the rationale. Compiled-table correspondence for the divisor rows is recorded by the game-data
companion's 2026-10-13 live record of all three difficulty records.

## Search records

No rule is labeled `not found in d2-decomp`. Rungs (2026-10-13, reused): root anchor through the
do-function and missile tables (DPO-R08d) and the call lines of `0xDE0C0`; output anchor on stat ids
6, 8 and `0xB2` pushed before the stat and overlay helpers; data anchor on MonStats `+0xA0` and
DifficultyLevels `+0x28`, `+0x2C`; decompile view only to check byte runs; runtime not armed.
2026-10-14 targeted readings: the callees named above. For `overlayCount` and `formStateList`, the
D2Common image's instruction listings were searched for stores to DataTables `+0xBBC`, `+0xBC0`,
`+0x17C` and `+0x180`; only readers were found (the overlay helper, a family of record getters, the
apparent-form helper and a unit-linked cleanup routine), so their loaders remain unidentified
(LSE-Q02, LSE-Q03).

<a id="observation-procedures"></a>
## Observation procedures

Both are pending. Use an owned matching original installation and a read-capable debugger or
instrumentation tool. Record module fingerprints, the effective data profile, date, inputs, frames,
seed pairs and outputs. Add each RVA to the loaded module base. Offsets locate evidence; they are not
required layouts. Keep captures private and publish narrow measurements.

### LSE-O01: leech draw count

Any class, level 1 or higher, Normal, single player. Prepare a weapon with life and mana leech
(stats 60 `lifedrainmindam` and 62 `manadrainmindam`) and a control weapon identical except for the
mana leech. Stand next to one zombie in the Blood Moor and attack once with each, from the same saved
game seed and the same actions. Capture entry and exit of `D2Game.dll+0xDBA40` and read the attacker's
seed at Unit `+0x20`/`+0x24` through the attacker argument. Expected: the first hit's leech steps the
seed exactly once; the control's leech does not step it.

### LSE-O02: defender mana drain

Two characters in one game, or a player hitting a monster with mana; Normal (divisor 1); attacker with
`manadrainmindam` 5. Capture entry of `D2Game.dll+0xDA970` (defender in `ebx`, amount in `edi`) and the
defender's stat 8 before and after. Expected: amount `5 << 6` = 320 raw units; new mana = old - 320,
or 0 when that is below 256.

## Runtime records

| Vector | Source profile | Date | Module + RVA | Records | Result |
|---|---|---|---|---|---|
| none | vanilla LoD 1.13c | 2026-10-13 | none armed | 0 | A running member held a dead level-1 character with no hit in progress; no hit could be produced and no vector was observed. |

## Migration record (2026-10-14)

- Scope: the timed status application rules, their vectors, observation and questions moved to
  [combat-status-application](../../specs/combat-status-application.md); its companion's
  [identifier mapping](combat-status-application.md#identifier-mapping) records the move. This spec keeps
  leech and drain and was retitled "Leech and Drain"; the file name is unchanged.
- The spec is unconsumed (no plan outside the gap folder, code or test cites it), so ids were
  renumbered in execution order without aliases and sibling references updated.
- Checks run: the spec lint on this spec, companion and gap file (result reported with this change).
- Gaps closed: cap sources and the cap-reader question (LSE-R11); overlay helper callees (LSE-R14
  requests); the monster-attacker branch's apparent form, additive write and null-unit reads (LSE-R15);
  divisor table cells (LSE-R07). Each closure's evidence is above.
- Gaps left open: [gap file](../../plans/gaps/combat-leech-and-drain.md) (stored-effect
  integration, disguise and form-list producers, overlay count identity, client messages, pending
  gameplay observation).
- Label changes: LSE-R11, LSE-R14 and LSE-R15 asm-only -> verified in 1.13c code (static). The rule
  stating why the poison magnitude is in the total, which was unresolved, became part of LSE-Q05,
  because it asks for confirmation, not for a behavior. The later-stage pointer rule was dropped; its
  pointers are in the Scope and Related specs.
- Blind-implementer pass: complete for this contract (full runs 1 to 3 and delta check 1, all
  2026-10-14; records below). The final applicable run returned MISMATCH 0 with no unresolved GUESS,
  GAP or DISCRIM. Two passes on 2026-10-13 against the earlier contract
  (run 1: 8 guesses, 7 gaps; run 2: 10 guesses, 4 gaps; all folded then) do not review this contract.
- Status: draft.

## Contradiction records

### Life-add gates on the monster-attacker branch (LSE-R15)

- Affected ids: LSE-R12, LSE-R15, section 6; vector LSE-V12 (m).
- Conflict: section 6 of the 2026-10-14 draft (and the earlier inline text's edge-case list) said a
  dead, dying or state-92 attacker gains no life, without limiting this to `LifeAdd`; LSE-R15 adds
  life through the additive write with no gate.
- Settled 2026-10-14 by rereading `0xDBA40`: the monster branch calls #10551 `(attacker, 6, total,
  0)` directly; it does not call `0xDB570`, `0xDFA90` or #10494, and #10551 (`0x3A710`) tests only for
  a zero value. The gate applies to `LifeAdd` only. Section 6 was narrowed; LSE-V12 (m) asserts the
  ungated add.

## Blind-implementer pass 1 (2026-10-14, full run)

Fresh Sonnet reader given the spec and its cited dependency specs only. Counts: MISMATCH 0, GUESS 4,
GAP 6, DISCRIM 9, GATED 15, COVERAGE 24, HARDEST 3. Outcomes, spec edits unless stated:

| Entry | Outcome |
|---|---|
| GUESS 1 (defender stats when a case swaps in a monster) | LSE-V12 (a) and LSE-V14 (c) state `D` stats 8 and 10 |
| GUESS 2 (section 6 versus ungated LSE-R15 add) | Settled from the code (contradiction record above); section 6 and LSE-R15 text fixed; LSE-V12 (m) |
| GUESS 3 (form-state list count and bound) | LSE-R15 iterates indices 0 to count - 1; a nonpositive signed count leaves the type unchanged |
| GUESS 4 (difficulty sources) | LSE-R03 and LSE-R06 state `G.difficulty` and `A.game.difficulty` as separate inputs |
| GAP 1 (overlay list flag and pending condition names) | Neutral names `overlayListCondition`, `overlayPendingCondition` with LSE-Q03; raw values stay in this companion; STL-R06a and STL-R04 cited |
| GAP 2 (missing dependencies) | Section 9 adds DPO-R16, DPO-R17, DPO-R18 and CSA-R01 |
| GAP 3 (state 92 identity) | No change in the spec: GDF-R03 does not list state 92 (nor GDF-R02 stats 9 and 178), so the spec cites this companion's table record; reported as a game-data follow-up |
| GAP 4 (state storage for form states) | LSE-V12 (g), (h), (i), (l) state `hasStateStorage` true |
| GAP 5 (type-2 unit's exclusion condition) | LSE-V08 (f) states it false and mode 1 |
| GAP 6 (clamp state in LSE-V13) | LSE-V13 states defender life 0x20000 and the already-clamped, unchanged physical value |
| DISCRIM 1 (mulDiv operand order) | LSE-V04 (b): 838860 versus swapped 838400 |
| DISCRIM 2 (effectiveness on the mana branch) | LSE-V03 (e): mana 640 |
| DISCRIM 3 (base test before divisors) | LSE-V06 (b): record 320 |
| DISCRIM 4 (drain column versus divisor difficulty) | LSE-V03 (f): `G` Nightmare, `A.game` Normal |
| DISCRIM 5 (mana cap source) | LSE-V07 (f): stat 8 = 0x1000, return 0x100 |
| DISCRIM 6 (effectiveness per term) | LSE-V12 (j): 135 versus 134 |
| DISCRIM 7 (room cap before effectiveness) | LSE-V12 (k): 128 versus 64 |
| DISCRIM 8 (ignoring the disguise condition) | LSE-V12 (l) |
| DISCRIM 9 (negative truncation) | Verified `shl 6` then `cdq; idiv` (truncating); LSE-V05 (c): -21 versus floor -22 |
| COVERAGE (advisory) | Added LSE-V05 (d) for the zero life divisor skip; the other entries are branches whose result an existing case or a documented fault already shows: no change |
| HARDEST (monster branch, overlay helper, stage order) | Covered by GAP 1, GUESS 1 to 3 and DISCRIM 6 to 8 |

New expected values follow from the rules by hand computation: `mulDiv(640, 0x20000, 100)` =
838860 (divide-last branch), swapped 838400 (divide-first branch); `mulDiv(50, 271, 100)` = 135;
`mulDiv(-21, 25600, 100)` = -5376. Rules, vectors and cases changed, so a fresh full reader run is
required (run 2 of the full runs).

## Blind-implementer pass 2 (2026-10-14, full run)

Fresh Sonnet reader, spec and cited dependency specs only. Counts: MISMATCH 0, GUESS 5, GAP 7,
DISCRIM 6, COVERAGE 24. Outcomes, spec edits unless stated:

| Entry | Outcome |
|---|---|
| GUESS 1 / GAP 3 (difficulty clamp versus GDA-R05 rows 0..2) | Section 3 rows for `G.difficulty` and `A.game.difficulty`, LSE-R06 and section 6 restrict the contract to 0..2 and cite CSA-Q06; the clamp and the unguarded drain read are kept in the mapping table above as evidence |
| GUESS 2 (when the stat list is read in LSE-R14) | Settled from `0x32C10`: the unit's stat list field is reloaded after the attach call before the `0x100` test; LSE-R14 says it is re-read after `Attach` |
| GUESS 3 / GAP 1 (raw flag in LSE-V11) | LSE-V11 writes `conditions {overlayListCondition}`; `0x80` mapping added above |
| GUESS 4 (pending condition in the log) | Preamble logs setting `overlayPendingCondition` as an ordered event in LSE-V11 |
| GUESS 5 (what LSE-V13 logs) | Preamble: the attacker draw, `AbsorbHeal` and `LifeWrite` calls (whatever their amounts), the real `ManaDrain` with its argument and the following CSA-R01 requests are logged |
| GAP 2 (overlay condition meanings) | No change: already LSE-Q03 |
| GAP 4 (`R.total` in LSE-V01) | No change: a named intermediate; LSE-R01 only states what mitigation writes (DPO-R12a owns the write) |
| GAP 5 (neighbor log entries in LSE-V13) | Preamble names them consumed-contract intermediates; no section 5 row |
| GAP 6 (Draws notation) | Preamble defines (seed owner, modulus, result) |
| GAP 7 (asm-only lifecycle dependencies) | No change: LSE-R14 asserts requests only and its Gates column already cites LSE-Q01 for stored effects |
| DISCRIM 1 (old low word) | LSE-V09 (d): seed `{1, 1}` gives lo 0x6AC690C6, overlay 151 |
| DISCRIM 2 (base test on the monster path) | LSE-V12 (n): physical 0, total 261; LSE-V12 Discriminates corrected |
| DISCRIM 3 (`< 0` versus `<= 0`) | LSE-V07 (g) |
| DISCRIM 4 (draw before the base return) | LSE-V06 (c) |
| DISCRIM 5 (mana gates) | LSE-V08 (h) |
| DISCRIM 6 (total on the monster path) | LSE-V01 (d) |
| COVERAGE (advisory) | No change: each listed branch either does not change a rule's result or is already shown by a case or a documented fault |

No rule's behavior, ownership or gate boundary changed; vectors, cases and preamble text changed, so
the next reader run is full run 3.

## Blind-implementer pass 3 (2026-10-14, full run)

Fresh Sonnet reader, spec and cited dependency specs only. Counts: MISMATCH 0, GUESS 4, GAP 5,
DISCRIM 6 (five labeled, one unlabeled), COVERAGE 16. Outcomes, spec edits unless stated:

| Entry | Outcome |
|---|---|
| GUESS 1 (comparison signedness) | Settled from the listings: every comparison on these paths is a signed branch (`jg`, `jle`, `jl`, `jge`) or a zero test; section 4's notation line states signed i32 |
| GUESS 2 (guarded calls in LSE-R16 and LSE-V13) | `0xDE0C0` calls `0xDB570` for the absorb heal unconditionally (DPO-R16's guard is inside), and DPO-R17's positive-total test is part of the life-write step; section 3 row and the LSE-V13 preamble say the calls are logged whatever their amounts |
| GUESS 3 (base list as a search candidate) | Settled from `0x37F40`: with the extended-list condition set it loads the chain head at `+0x3C` before testing, so the base list is never tested; LSE-R14 states it; LSE-V11 (h) |
| GUESS 4 (scaled total of 0 or below) | LSE-R15 comment: such a total reaches the room test and then returns at `total <= 0` |
| GAP 1 (rest of the CSA-R01 log) | LSE-V13 asserts the prefix through `StaminaDrain(D, 0)` and cites CSA-V01 for the rest |
| GAP 2 (`LifeWrite` argument) | Section 3 row for `AbsorbHeal(D, R.absorbHeal)` and `LifeWrite(D, R.total)` (DPO-R16, DPO-R17) |
| GAP 3 (owner notation) | Preamble defines `owner t/i` |
| GAP 4 (difficulties in LSE-V05) | Inputs and every case state `G` and `A.game` difficulty |
| GAP 5 (stats in LSE-V07 and LSE-V08) | Inputs state the defaults and that cases give the stats they read |
| DISCRIM 1 (effectiveness after division) | LSE-V03 (g): 151 versus 150 |
| DISCRIM 2 (mana operand order) | LSE-V04 (c): 6553 versus 6550 |
| DISCRIM 3 (effectiveness operand order on the monster path) | LSE-V12 (o): 50049 versus 50000 |
| DISCRIM 4 (drain uses the rewritten value) | LSE-V14 (g): wrapped shift, `v` negative, no drain versus a drain to 0 |
| DISCRIM 5 (stat list read again after attach) | LSE-V11 (g); preamble lets a case's `Attach` create a stat list |
| DISCRIM 6 (unlabeled: skip of effectiveness 100) | LSE-V13 now uses physical 0x30000: 19660 versus 19659 |
| COVERAGE (advisory) | No change: each listed branch either does not change a rule's result or is already shown |

Each new rejecting input was checked by hand against both alternatives: `mulDiv(150, 6451, 100)` =
9676 versus 150; `mulDiv(320, 0x20000, 100)` = 419430 versus divide-first 419200;
`mulDiv(50, 100099, 100)` = 50049 versus divide-first 50000; `0x02000000 << 6` wraps to
-2147483648 and divides to -33554432; `mulDiv(640, 0x30000, 100)` = 1258291 versus
`mulDiv(100, 1258291, 100)` = 1258200. No rule's behavior, ownership or gate boundary changed, so
the next reader run is a delta check.

## Blind-implementer delta check 1 (2026-10-14)

Fresh reader given the same files plus the list of rules, vectors and cases changed after full run 3
and the findings each resolves. It recomputed every changed or added case and checked each
Discriminates clause. Counts: MISMATCH 0, UNRESOLVED 0, GUESS 0, GAP 0, DISCRIM 0. No spec change.
The blind-implementer pass is complete; Status stays draft because LSE-Q01 to LSE-Q05 and the
gameplay observations remain open.

## Contract refresh (2026-10-15)

This refresh reuses the dated 2026-10-12 to 2026-10-14 source and table records above for their
established original-artifact/profile boundaries. It adds no source reading, archive read,
compiled-table correspondence or gameplay observation, and does not renew their verification
dates. The initial focused refresh preserved all 17 rule bodies and labels, 14 executable-vector
blocks, two observation summaries and the historical implementation comparison; subsequent
reader-driven assertion additions are recorded below. No implementation plan, code or test
cites this spec or its prefix outside gap comparisons; the current ids are retained for this
campaign's stable cross-spec references.

### Current-contract corrections and ownership

- The opening description previously said without qualification that a dead or dying attacker
  regains no life. That overstated LSE-R12: the already-verified LSE-R15 additive path has no
  alive/death-delay gate, as the contradiction record above establishes. The description now
  limits refusal to the ordinary life-add path. It describes an effect request rather than
  claiming confirmed stored overlay state or client display. No eligibility rule changed.
- The client-view and output summaries likewise state update requests and retain the open
  sender/body/delivery boundary under LSE-Q04/RLI-Q02; unresolved STL-R14 selection does not
  establish it. The original source-folder assert used for category attribution remains in
  this companion's owning-category record rather than the language-neutral trigger summary.
- Section 3, section 9 and LSE-Q01 previously cited NRF-R08a as the replacing-stat-write owner.
  Its current contract owns fixed-point formats only. The unit replacing and additive writes
  remain supplied operations whose original stored effects are LSE-Q01; their binary references
  remain in the function table. This removes a false dependency claim without accepting a write
  implementation. Section 3 inventories current stat values, stat definitions and supplied
  operation effects; section 5 states capped helper returns exactly, including the existing
  negative above-cap result of LSE-V07 (e).
- Executable dependencies now identify exact GDF leaf ids. The surrounding total/clamp/apply,
  absorb-heal/life-write and status-handoff contracts are separately identified as placement
  and supporting contracts. DPO-R16 calls local `LifeAdd`; CSA-R01 calls local `ManaDrain`.
  Their reverse consumption does not require a local helper fixture to execute the whole hit.
- The update-mark reference now includes original D2Common ordinal 10158, as recorded by the
  [numeric foundations](core-numeric-and-rng-foundations.md#states-mode-and-creation) and
  [hit-recovery](combat-hit-recovery.md) function tables. This is a shared original-artifact
  locator, not a new behavioral reading. Update-queue effects remain RLI-Q02/LSE-Q01.

### Evidence reuse and remaining gaps

| Gap | Compatible evidence considered | Current outcome and smallest remaining check |
|---|---|---|
| LSE-Q01 stored write integration | Original local requests and cap/return reasoning above; current NRF-R08a/RLI-R02a/RLI-R02b and RLI-Q02; STL-R06a/R06b/R06c; HRC-R02b and KNB-R03b room lookup | Numeric/stat retrieval and successful list creation remain usable. Attach and list writes remain unresolved, and no complete replacing/additive write contract is supplied. The 2026-10-15 HRC/KNB record identifies and reads D2Common `0x2FE10` -> `0x3C390` -> `0x2D320` for room/level/town selection; it does not verify update queue mutation through `0x665C0`. Complete the stored-write and update-queue output paths with integration vectors; no gate closes |
| LSE-Q02 disguise and form list | Local `0x501E0` reading and LSE-V12 (g), (h), (i), (l); current STL-R07a and RLI-Q02 | Supplied-form selection remains verified, but disguise bookkeeping and list production remain unresolved. Trace the disguise writer and list/count producer; the historical readers-only search is not absence proof |
| LSE-Q03 overlay identity | Local `0x32C10` reading, LSE-V11, GDA-R13 definition fields, compatible item-event overlay references and the crushing-blow question | Supplied count/conditions and ordered requests remain supported. Neither definition-field correspondence nor another consumer identifies the counted collection or its original rows. Trace publication of the collection/count and verify named overlay cells before accepting profile-specific ids |
| LSE-Q04 client messages | Local update request, current RLI-Q02 and unresolved STL-R14/STL-Q01 | No compatible sender/body/delivery contract establishes this message boundary. Trace the actual server update sender under a net owner; a queue request alone proves no client message |
| LSE-Q05 observations | Original 2026-10-13 failed-hit record, 2026-10-14 static/table work and current DPO-Q02 | No valid matching gameplay capture exists. A new observation campaign is outside this migration; retain both pending procedures and all runtime flags as no |

The historical LSE attempt's unread-room-getter statement is retained as a statement of that
attempt's coverage, rather than a current claim that no shared room-getter evidence exists.
Newer GDF effective-file fingerprints do not establish identical historical loaded tables or
override sets for the LSE reads; the original table records retain their uncaptured identity
limits. No rule label, evidence gap, expected value or observation status is promoted.

Self-review completed for the revised dependency references, input/output inventories, public
original-artifact references, retained source dates and gates, and the complete focused diff.
Mechanical documentation checks returned 0 errors and 0 warnings. Before/after comparison
confirmed exact preservation of all 17 rule bodies and their Id-map rows, all 14 vector blocks,
both observation summaries, Appendix A's evidence dates and the historical implementation
comparison. The historical three full runs and delta check alone do not review this refresh;
the fresh reviews and their resulting assertion additions follow.

## Regeneration blind-implementer full run 1 (2026-10-15)

A fresh smaller-model reader used the main contract and cited dependency contracts only, with
no original-instruction access, runtime, repository implementation, web or earlier review reports. It
independently recomputed all 14 vector families. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 2,
GATED 20, COVERAGE 10, HARDEST 3. No mismatched value or missing behavioral input was reported.

| Entry | Outcome |
|---|---|
| DISCRIM 1: LSE-V09 could pass with local life add before mana add because only final stats were asserted | Add an observer on the actual local helpers' supplied unit-write requests. The main case now asserts `(A, 8, 0x500, 0)` before `(A, 6, 0x1A00, 0)`. Neither `ManaAdd` nor `LifeAdd` is replaced by a logging mock; their local guards and arithmetic execute |
| DISCRIM 2: LSE-V07's zero-amount case could read or rewrite the unchanged value and still pass | Add read/write observers around actual capped-add execution. Cases (c) and (g) assert empty unit-field/stat/state-read and unit-stat-write logs, retaining all existing final values and returns |
| GATED 20 | No change: LSE-Q01 to LSE-Q05; integration-sensitive LSE-R10/R14/R15/R17; supplied-input-stage LSE-V02/V03/V04/V05/V06/V09/V10/V11/V12/V13/V14 retain their documented limits. LSE-V01/V07/V08 are isolated fixtures, not claims to execute their surrounding hit path |
| COVERAGE 1: LSE-R03 null defender and present monster data with absent MonStats record | No change: the null defender shares the supplied nonmonster `P = 100` path; LSE-V03 (c) demonstrates the unchanged-record missing-data return. An absent nested record uses the same early-return result, with both guards explicit in LSE-R03 |
| COVERAGE 2: LSE-R05 other attacker type | No change: other types join the already exercised LSE-R15 path with apparent type unchanged absent disguise. LSE-V12 main and (f) show nonplayer routing and record rewrites; another fixed type adds no arithmetic or request recipe |
| COVERAGE 3: LSE-R06 negative nonzero divisor | No change: LSE-V05 (c) shows truncating signed negative quotient and the nonpositive-add result; (a)/(d) distinguish nonzero/zero divisors. A negative divisor uses that same signed-quotient/nonpositive-amount recipe, with no additional sign branch |
| COVERAGE 4: LSE-R10 addition overflow and null unit with nonpositive amount | Add LSE-V07 (h): positive amount 2147483647 plus current 256 wraps to -2147483393, which is below the signed cap and is requested unchanged. This distinguishes wrapped addition from widened/saturating arithmetic. No extra null-unit case: the (c)/(g) empty read/write logs now require the amount guard before any unit access, and LSE-R10 states it independently of unit identity |
| COVERAGE 5: LSE-R11 null list or missing cap definition | No change: RLI-V02 (d) directly checks null-list/unavailable-definition reads returning 0. LSE-V07 (a)/(e)/(f) check using the returned cap, including an initially above-cap current value. A supplied zero cap follows that same signed comparison/`cap - cur` recipe; no new getter contract belongs here |
| COVERAGE 6: LSE-R12 null unit or monster mode 0 | No change: LSE-R12 explicitly rejects null and monster modes 0/12. LSE-V08 (a)/(b)/(f) show refused life addition, and (e) distinguishes the player/monster mode sets. Extra refused identities use the same no-write/zero-return result |
| COVERAGE 7: LSE-R14 list without state storage and zero overlay count | No change: LSE-V11 (b) already supplies a present stat list without state storage and requires allocation. LSE-V11 (d) rejects id equal to the count; count 0/id 0 is the same bound recipe rather than another overlay behavior |
| COVERAGE 8: LSE-R15 nonpositive form count, no active usable form or invalid gfxtype | No change: the default empty form list exercises count 0; a negative signed count also executes no iteration. LSE-V12 (i) skips an active kind-0 state and an inactive kind-2 state, then retains type; kinds other than 1/2 share the explicit no-selection branch. Cases (g)/(h)/(l) distinguish selection, order and disguise eligibility |
| COVERAGE 9: LSE-R16 remainder after logged prefix and later statuses | No change: LSE-V13 explicitly asserts only the prefix through stamina drain and delegates subsequent requests to CSA-R01/CSA-V01. Executing those status rules is outside this fixture's assertion boundary |
| COVERAGE 10: LSE-R17 negative drain return and subtraction wrap | Add explicit return 0 to existing LSE-V14 (g), whose negative input already asserts no write; add isolated LSE-V14 (h): current -2147483648 minus positive drain 1 wraps to 2147483647 and is retained, returning 1. Widened/saturating subtraction would instead floor to 0 and return 0 |
| HARDEST 1: LSE-V12 form selection, three minimums, room and null attacker | No change: all named state/count/gfxtype and resource inputs are supplied, with separate exact cases (a) to (o). The form producer and stored additive-write effects remain LSE-Q02/LSE-Q01; no default for either is invented |
| HARDEST 2: LSE-V13 rewritten record, adds, hand-offs and drain | No change: the fixture states which local operations execute, which surrounding requests are supplied/logged and where the asserted prefix stops. CSA-V01 owns the rest; the reader independently recomputed the local values without a guess |
| HARDEST 3: LSE-V03 to LSE-V05 hit-game versus attacker-game difficulty and mulDiv | No change: LSE-V03 (f) discriminates the two difficulty sources, LSE-V04 (b)/(c) operand order and LSE-V05 truncating/zero-divisor behavior. Inputs and exact dependency operations are explicit |

A primary-reader citation correction also removed `DifficultyRecord` from the list of operations
owned by GDA-R05. LSE-R06 locally selects the prepared row in the covered 0..2 domain, supported
by the existing D2Common #10237 / `0x603C0` record; GDA-R05 owns row preparation/fields only.
No out-of-domain clamp assertion or new source-reading date is added.

The two new arithmetic cases follow the already stated i32 operations:
`i32(low32(256 + 2147483647)) = -2147483393` and
`i32(low32(-2147483648 - 1)) = 2147483647`. All earlier numeric results, 17 rule bodies, Id-map
labels/gates, observation flags and the historical implementation comparison remain unchanged.
The added observers assert actual local execution; supplied persistence does not accept original
stored-write integration. A fresh full run 2 is required after these assertion/case additions.

## Regeneration blind-implementer full run 2 (2026-10-15)

A fresh smaller-model reader independently recomputed all 14 vector families from the main
contract and its cited dependencies. Counts: MISMATCH 0, GUESS 0, GAP 1, DISCRIM 4, GATED 6,
COVERAGE 7, HARDEST 3. The one GAP is an access limitation on a data-only record, not a
contradictory cell or an unexecutable behavioral vector.

| Entry | Outcome |
|---|---|
| GAP 1: LSE-R07's linked table evidence was outside the reader's allowed source set | No spec edit or new data claim. Primary review checked the existing Table records section: the six DifficultyLevels cells give 1, 2, 3 for both divisors, from the complete 586-byte file read on 2026-10-14. The recorded vanilla-profile, uncaptured hash/physical archive/override and compiled-correspondence limits remain. Full run 3 receives only this companion's identity and table-record sections in addition to the behavioral contracts, without earlier reviews |
| DISCRIM 1: LSE-V07 did not distinguish a zero-mana guard from only the zero-life guard | Add LSE-V07 (i), `ManaAdd(A, 0)` with current mana 0x1100 above cap 0x1000: unchanged mana, return 0 and an empty actual read/write log. A `< 0` guard would write 0x1000 and return -0x100 |
| DISCRIM 2: LSE-V11 did not reject an overlay lower bound of `id <= 0` | Add LSE-V11 (i), valid id 0 with positive count and an existing overlay list: pending condition, list-stat write request with value 0, then update request. No stored zero-entry behavior is asserted |
| DISCRIM 3: LSE-V12 (g) used difficulty 0/divisor 1 and could pass if apparent-player forms wrongly took player divisors | Preserve (g); add independent LSE-V12 (p) with a monster in apparent player form and attacker-game Hell divisors 3/3. The record stays 640/320, mana is 102 and life 0x10CC; dividing would give 213/106, mana 33 and life 0x1044 |
| DISCRIM 4: only monster form selection was shown, leaving the allowed input-type set undistinguished | Add LSE-V12 (q) for actual type 3 with an active kind-2 form state: apparent-player path, mana 102, life refused by actual-type eligibility, one draw and overlay 152. This rejects an over-restrictive form-type filter. No synthetic true state result or pure-read invocation requirement is introduced for type 2: omitting that local type rejection while consuming RLI-R03's state predicate is behaviorally equivalent, because that predicate already rejects type 2 |
| GATED 6 | No change: LSE-Q01 to LSE-Q05 and the difficulty-domain limit of CSA-Q06 remain explicit. Supplied values/requests do not establish producers, stored integration, messages or gameplay confirmation |
| COVERAGE 1: LSE-R03 null or nonmonster defender gives 100; absent MonStats is separate from absent monster data | No change: supplied player defenders throughout exercise `P = 100`, and LSE-R03 explicitly shares it with null/nonmonster inputs. LSE-V03 (c) shows the unchanged-record missing-data return; the absent nested record has that same output recipe and does not need a second fixed-null fixture |
| COVERAGE 2: other attacker types take the monster route; form checks admit only 0/1/3 | LSE-V12 (q) now covers the other-type route and allowed type-3 form selection. For excluded type 2, the owning RLI-R03 predicate already returns 0; see DISCRIM 4's equivalence disposition. No additional observable result is left unrepresented by a missing local calculation recipe |
| COVERAGE 3: zero mana add independently returns before stat reads | Added LSE-V07 (i), as for DISCRIM 1 |
| COVERAGE 4: overlay id 0 is accepted with a positive count | Added LSE-V11 (i), as for DISCRIM 2 |
| COVERAGE 5: negative life room caps a positive total below zero before returning | Add LSE-V12 (r): total 271 and room -256 become total -256, then no additive write or overlay; current life 0x2100 remains. Existing positive/zero-room cases (b)/(c) did not show an initially above-cap attacker |
| COVERAGE 6: apparent player form with a non-1 attacker-game divisor still skips divisors | Added LSE-V12 (p), as for DISCRIM 3 |
| COVERAGE 7: nonmonster actual type with matching form state | Added LSE-V12 (q), as for DISCRIM 4; actual-type life-add eligibility remains independently visible after apparent-form selection |
| HARDEST 1: LSE-V12 routing, signed division, per-term minimums, effectiveness and room | No additional change: existing cases distinguish the arithmetic and ordered decisions; (p)/(q)/(r) add the newly identified divisor/type/negative-room boundaries. All inputs are explicit; producer and stored-write integration gates remain |
| HARDEST 2: LSE-V13 leech, absorb heal, life write, mana drain and stamina drain | No change: the fixture names the actual local leech/drain operations and supplied surrounding requests, asserts their ordered prefix and cites CSA-V01 for the remaining status sequence |
| HARDEST 3: LSE-V11 list search, allocation, attachment, stat-list reread and write order | No additional change beyond valid-zero-id case (i): cases (a)/(b)/(g)/(h) already distinguish selection, storage guard, post-attach reread and base-list exclusion. Stored list effects remain LSE-Q01 |

The new form case computations reuse `mulDiv(320, 2048, 100) / 64 = 102` and
`mulDiv(640, 2048, 100) / 64 = 204`. Applying forbidden player divisors in case (p) instead gives
`mulDiv(106, 2048, 100) / 64 = 33` and `mulDiv(213, 2048, 100) / 64 = 68`.
The actual type-3 life add in (q) returns 0; its apparent player form does not change LSE-R12's
actual-unit type test. Case (r) has `8192 - 8448 = -256`, so signed capping precedes the
nonpositive return. Earlier values and all 17 rule bodies, labels and gates remain unchanged.
Only fixtures and assertions were added; full run 3 is required for the resulting contract.

## Regeneration blind-implementer full run 3 (2026-10-15)

A fresh smaller-model reader independently outlined the 16 verified code rules and checked
data-only LSE-R07. It recomputed all 14 vector families and 92 explicit cases from the main
contract and its named dependencies. It read only the companion's identity and Table records
sections for the data claim, without previous reviews, implementation, original instructions,
runtime or web access.

Final finding counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 5, COVERAGE 0, HARDEST 3.
The reader initially placed its review-coverage totals under the DISCRIM/COVERAGE headings
and counted a nonexistent unlettered LSE-V07 execution. It explicitly corrected those reporting
units: zero findings in both lists, nine named V07 cases and 92 total explicit cases.
No rule, fixture or result was changed to fit that bookkeeping correction.

The five GATED entries retain LSE-Q01 to LSE-Q05: original stored write/attach/update effects,
form producers, overlay identities/condition meanings, client delivery and gameplay observation.
The existing 0..2 difficulty domain and its CSA-Q06 limit also remain; a different reporting
grouping closes no question. Supplied fixture persistence is not accepted original integration.

The R07 check resolved the previous reader's access limitation: the allowed record identifies
six named LifeStealDivisor/ManaStealDivisor cells, raw 1/2/3 values, the dated full-file read,
virtual path, header selection and row convention. Its uncaptured effective-file hashes,
physical archive and override limits remain. This is consistency review of the recorded table
evidence, not a new archive read or gameplay observation.

| HARDEST entry | Disposition |
|---|---|
| LSE-V12 apparent-form routing, monster arithmetic, caps and signed boundaries | No change: the reader recomputed all 19 cases, including the Hell apparent-player, actual type-3 life refusal and negative-room cases, without a mismatch or missing decision. Producer and stored-effect gates remain. |
| LSE-V11 overlay search and post-attach ordering | No change: all ten cases, including valid overlay id 0, matched the supplied count/conditions and ordered requests. Original stored attachment, list writes and update effects remain LSE-Q01. |
| LSE-V07 capped-add boundaries, wrapping and returns | No change: its nine named cases matched the local guards, signed cap arithmetic, empty early-return logs and exact returns. The observer logs actual local execution and supplies only the dependency effects already declared. |

This final applicable full run has no mismatch or unresolved required finding. Every earlier
finding has a recorded resolution; no behavior, input, fixture, expected output or acceptance
boundary changed after this clean run. Reader review is complete for this revision. Draft
status, evidence gaps and pending observations remain; neither approval nor gameplay parity
is inferred.

## Independent review spot-check, 2026-10-15

A separate reviewer re-read original D2Game disassembly (original 1.13c instruction listings, read-only), deriving each reading before comparing it with the rule text: 0xDBA40 (LSE-R02 to LSE-R06, LSE-R08, LSE-R09, LSE-R13, LSE-R15), 0xDB570 and 0xDAE70 (LSE-R10), 0xDFA90 (LSE-R12), 0xDA970 (LSE-R17), 0xDE0C0 (LSE-R16 order) and D2Common 0x390F0 (LSE-R11 life cap: stat 7, layer 0, through the selected-array getter). All matched, including the hit-game drain column versus the attacker-game divisor row, the zero-divisor skips, the signed division by 64, the draw only after both rewritten values are tested, and the monster-branch room cap without life-add gates. Seed steps for `{1,0}`, `{0,2}` and `{1,1}` were recomputed. Section 3 now names the pipeline's aliases for the drain and poison fields; no rule, vector or label changed, so the final full reader run above still applies.

## Restructure (2026-10-15)

De-duplication pass after the issue 442 spec review, which found that two rules restated the
damage pipeline. Records above this one are dated and keep the rule numbers in force when they
were written; this record maps them to the current ids.

- File rename: `combat-leech-and-status-effects` became `combat-leech-and-drain` (spec, this
  companion and the gap file). The timed status appliers moved to
  [combat-status-application](../../specs/combat-status-application.md) on 2026-10-14, so the old
  name no longer described the scope; the title "Leech and Drain" and prefix LSE are unchanged.
- Old R01 (leech-related terms of the immediate total) restated DPO-R12a and was removed. Its
  one local consequence, that the leech operation neither reads nor writes `R.total`, moved into
  the scale-and-store rule (now LSE-R03), with its rationale (no access to `+0x4C` in `0xDBA40`).
  LSE-V01 keeps its poison-term total (no pipeline vector has a nonzero poison term) citing
  DPO-R12a, and its two leech-leaves-the-total cases, relettered (a) and (b). Its regular-monster
  and hireling total cases were dropped because DPO-V30 (a) to (c) assert the same terms.
- Old R16 (stage order) restated DPO-R14 to DPO-R18 and CSA-R01 and was removed. Section 2's
  placement bullet now cites those owners and keeps the local consequence that the leech overlay
  draw is the first draw of the stage. The mana-drain rule (now LSE-R15) names its placement by
  reference. LSE-V13 is unchanged in inputs and results and now cites LSE-R15 and LSE-R08 with the
  placement owners.
- Remaining rules renumbered in execution order: old R02 to R15 became R01 to R14, old R17 became
  R15. Questions, vectors and observations keep their ids. The current-state sections of this
  companion (function references, rule rationale, table records) and the gap file use the new
  ids; dated records do not.
- No rule's behavior, label, value or vector result changed. Sibling references to the old file
  name and rule ids are updated separately from the restructure id map.

## Issue #473 publication review (2026-10-23)

Issue [#473](https://github.com/gravestench/dark-magic/issues/473) owns the current leech-and-drain documentation handoff; its issue-body paths were corrected from the pre-rename `combat-leech-and-status-effects` filenames to the current files. A fresh blind reader used only the main contract and its named dependency specs to review LSE-R01..R15 and LSE-V01..V14. The reader recomputed the hit-game Drain versus attacker-game divisor selection, truncating arithmetic, caps, negative amounts, one-step seed results, apparent-form routing, the monster room cap and signed defender mana drain without a material contradiction. This was supplied-input document review, not execution or original-game corroboration.

A separate publication reviewer checked the current spec, public evidence, gap file, source manifest, index and dependency owners. The current spec still cited several former NRF section-3 record locations after the RLI split, and its section 9 mixed GDA field ids into a GDF link with an interrupted NRF list. Sections 3 and 9 now name RLI sections 3.1, 3.2, 3.3 and 3.7 for the relevant logical records and reads, and separate NRF numeric/RNG, GDF identity and GDA table-field owners. Section 10 adds direct owner links. These are citation/ownership repairs; no LSE rule, vector input or expected result changed.

The gap file now identifies its older implementation comparison as historical, with no recorded exact code snapshot. Its old `asm-only` assessment of the monster-attacker branch does not supersede the current LSE-R14 static label or the LSE-Q01 stored-write integration gate. The spec acceptance paragraph explicitly limits this 1.13c supplied-input handoff: it establishes neither 1.14d parity nor Dark Magic implementation. The original-profile named-cell limits, five open LSE-Q questions, pending observations and draft status remain. This audit did not read original bytes or game files, execute fixtures, capture gameplay, or refresh the current implementation comparison.
