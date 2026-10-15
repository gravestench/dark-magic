# Evidence: combat status application

Spec: [Combat Status Application](../../specs/combat-status-application.md).
Artifacts: [original source manifest](../source-manifest.md).
Limits: [evidence gaps](../../plans/gaps/combat-status-application.md).

## Identity and verification method

Evidence profile: Lord of Destruction Expansion 1.13c, Windows x86, little-endian. Artifact ids
`LOD113C-D2GAME` and `LOD113C-D2COMMON` resolve to the fingerprints in the source manifest. RVAs are
module-relative; ordinals belong to the identified `D2Common.dll`. Function descriptions are analyst
labels, not original symbol names or required APIs.

This spec was created on 2026-10-14 by moving the timed status application rules out of
[core-numeric-and-rng-foundations](../../specs/core-numeric-and-rng-foundations.md) (NRF) and
[combat-leech-and-status-effects](../../specs/combat-leech-and-status-effects.md) (LSE) into one owner.
The rules were built from the NRF text and reconciled rule by rule against LSE. Existing static
readings are reused with their dates (2026-10-12/13 for LSE's applier readings, 2026-10-12/13/14 for
NRF's). Where the two disagreed, or where a vector needed an assertion neither stated, the original
instructions were reread on 2026-10-14 (see [Contradiction records](#contradiction-records)). Readings
used original instruction listings; for the cold expiry callback, a reconstruction checked against the
original bytes. No game data was read and no gameplay observation was made in this pass.

## Owning category

The template assigns the category from the source-file assert string of the function that computes
the result. Every applier here asserts through one file-name argument at D2Game `.rdata` VA
`0x6FD19FF9`, which holds an empty string in the 1.13c image (the byte after the terminator of
another file name), with line arguments `0xA2A` (null attacker), `0x250` (zero freeze divisor) and
`0x272` (freeze reaching a non-monster). LSE recorded this on 2026-10-13. No folder is recoverable,
so the category follows the only caller: the hit-application function D2Game `0xDE0C0`, which the
pipeline spec places in `combat` (`UNIT\SUnitDmg`). The cold applier is also called from the freeze
applier, which has the same caller. Category: `combat`.

<a id="dispatch-chain"></a>
## Dispatch chain

Reused from LSE section 2 (2026-10-13) and
[combat-damage-pipeline-order](../../specs/combat-damage-pipeline-order.md) DPO-R08d; not re-walked
in this pass except for the hand-off slice.

1. Skill server do-function table (`SkillsRecord.srvdofunc`) at D2Game VA `0x6FD274A8`, index 1 ->
   D2Game `0xA2C40` -> build/enqueue `0xDDDE0` and dequeue/apply `0xDED30` (also through `0xA20C0`).
2. Missile `SrvHitFunc` table at VA `0x6FD2DA48` -> missile contact `0x3AE10`, and the callback route
   `0xA10B0`; both call `0xDE0C0`.
3. `0xDED30` -> hit-application `0xDE0C0`.
4. In `0xDE0C0` (2026-10-14 reread of the slice from the life clamp to the lethal decision): life clamp,
   leech `0xDBA40`, absorb heal through `0xDB570`, total-damage life subtraction and write, then mana
   drain `0xDA970` (record `+0x3C`), stamina drain `0xDA930` (`+0x40`), stun `0xDCC20` (`+0x44`),
   cold `0xDC780` (`+0x30`), freeze `0xDDAC0` (`+0x34`), poison `0xDCAB0` (`+0x28`, `+0x2C`), burn
   `0xDC940` (`+0x14`, `+0x18`), then the unit stat read of stat 6 layer 0 through D2Common #10973;
   a positive signed result clears bit `0x2` of the record byte at `+0x04` (CSA-R01).
5. Freeze fallback: `0xDDAC0` jumps to a block that calls `0xDC780` with the unchanged length and
   attacker (CSA-R07).
6. Expiry: the cold applier stores callback `0xDB4C0` on its list; the lifecycle calls it from list
   removal (STL-R05) after event kind 12 expiry (STL-R03).

## Function references

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | `0xDE0C0` | — | Hit application: status hand-off and lethal decision | CSA-R01 | Disassembly 2026-10-13 (LSE, NRF); hand-off slice reread 2026-10-14 |
| LOD113C-D2GAME | `0xDCC20` | — | Stun applier | CSA-R03, CSA-R04 | Disassembly 2026-10-13; independent listing 2026-10-14 (NRF); full reread 2026-10-14 |
| LOD113C-D2GAME | `0xDC780` | — | Cold applier with shatter draw | CSA-R05, CSA-R06 | Disassembly 2026-10-13; full reread 2026-10-14 |
| LOD113C-D2GAME | `0xDDAC0` | — | Freeze applier, fallback and tail | CSA-R07, CSA-R08, CSA-R09 | Disassembly; tail bytes decoded and checked against an independent listing 2026-10-13 and 2026-10-14; reread 2026-10-14 |
| LOD113C-D2GAME | `0xDCAB0` | — | Poison applier | CSA-R10, CSA-R11, CSA-R12 | Disassembly 2026-10-13; full reread 2026-10-14 |
| LOD113C-D2GAME | `0xDC940` | — | Burn applier | CSA-R10, CSA-R11, CSA-R12 | Disassembly 2026-10-13; call/argument skim 2026-10-14 confirming state 115 and the poison shape |
| LOD113C-D2GAME | `0xDB4C0` | — | Cold expiry callback | CSA-R13 | Reconstruction checked against bytes; reread 2026-10-14 |
| LOD113C-D2GAME | `0xDAE20` | — | Cold effect value | CSA-R05, CSA-R08 | Byte run decoded and checked against an independent listing 2026-10-13; NRF correction 2026-10-14 (invalid monsters give -50) |
| LOD113C-D2GAME | `0x1220` | — | MonStats record by class | CSA-R03, CSA-R07 | Reconstruction checked against bytes; returns null out of range |
| LOD113C-D2GAME | `0x13B0` | — | MonStats2 flag test through MonStatsEx | CSA-R06 | Disassembly and independent listing (NRF 2026-10-14) |
| LOD113C-D2GAME | `0x22A80` | — | Monster condition-word intersection | CSA-R02, CSA-R07 | Full disassembly 2026-10-14 (NRF); reread 2026-10-14 |
| LOD113C-D2COMMON | `0x1A550` | 10064 | Boss condition (recordGate) | CSA-R02, CSA-R03, CSA-R07 | Full disassembly 2026-10-14 (NRF) |
| LOD113C-D2COMMON | `0x1A2D0` | 11104 | Class-group selector (unitGate) | CSA-R02, CSA-R03, CSA-R07 | Full disassembly 2026-10-14 (NRF) |
| LOD113C-D2COMMON | `0x603C0` | 10237 | Difficulty record, clamped index | CSA-R05, CSA-R08 | Reconstruction checked against bytes |
| LOD113C-D2COMMON | `0x38520` | 10871 | Find state list | dependency NRF-R12a | Disassembly; null-unit entry reread 2026-10-14 |
| LOD113C-D2COMMON | `0x33CD0` | 10494 | State test | dependency NRF-R12a; CSA-R07 | Disassembly |
| LOD113C-D2COMMON | `0x34510` | 10945 | State set request | dependency NRF-R12a, STL-R07 | Disassembly |
| LOD113C-D2COMMON | `0x38100`, `0x38060`, `0x37D70` | 11013 / 10785 / 10911 | New list, list state, list callback | dependency STL-R06a | Reconstructions checked against bytes |
| LOD113C-D2COMMON | `0x38000`, `0x38020`, `0x38380` | 10412 / 10161 / 11043 | Expiry getter and setters | dependency STL-R06a | Reconstructions checked against bytes |
| LOD113C-D2COMMON | `0x3A7D0` | 10807 | Attach list | dependency STL-R06b (gated, CSA-Q01) | Entry/callee inspection only |
| LOD113C-D2COMMON | `0x3A280`, `0x3A7B0` | 10261 / 10188 | Replacing list stat write (new / existing list) | dependency STL-R06c (gated, CSA-Q01) | Entry/callee inspection only |
| LOD113C-D2COMMON | `0x39140`, `0x38B70` | 10680 / 10973 | List and unit stat reads | dependency NRF-R12b, NRF-R12c | Entry-to-return reading 2026-10-14 (NRF) |
| LOD113C-D2COMMON | `0x33110` | 10819 | Animation-rate refresh | dependency (CSA-Q02) | Quotient slice only (NRF) |
| LOD113C-D2COMMON | `0x33AB0` | 11061 | Stay-death test | dependency STL-R12a | Disassembly 2026-10-14 (NRF) |
| LOD113C-D2GAME | `0x8EA80`, `0x8E6F0` | — | Timer scheduling, node removal | dependency STL-R01; CSA-Q02 | Independent listings 2026-10-14 (NRF); scheduler head bytes checked 2026-10-13 (LSE) |
| LOD113C-D2GAME | `0xDA970`, `0xDA930` | — | Mana and stamina drains | dependency LSE-R17, DPO-R18 | Owned elsewhere |
| LOD113C-D2GAME | `0x9F7A0`, `0xDBD90` | — | Generic and freeze expiry callbacks | identity only; STL-R11 | Owned by the lifecycle |

Imported calls from D2Game go through thunks: state set `0xA3CA` -> #10945, state test `0xA68E` ->
#10494, find list `0xA5D4` -> #10871, new list `0xA6A0` -> #11013, list state `0xA5B6` -> #10785,
callback `0xA75A` -> #10911, attach `0xA60A` -> #10807, new-list stat write `0xA76C` -> #10261,
existing-list stat write `0xA6E2` -> #10188, list stat read `0xA712` -> #10680, expiry read `0xA790`
-> #10412, expiry writes `0xA778` -> #10161 (cold, freeze, stun) and `0xA7C6` -> #11043 (poison,
burn), difficulty record `0xA7BA` -> #10237, boss `0xA580` -> #10064, class group `0xA7F6` -> #11104,
refresh `0xA59E` -> #10819, stay-death `0xA7E4` -> #11061, unit stat read `0xA42A` -> #10973.

## Semantic-to-original mapping

| Semantic term | Original mapping |
|---|---|
| `R.manaBurn`, `R.staminaBurn` | DamageRecord `+0x3C`, `+0x40` |
| `R.stunLength`, `R.coldLength`, `R.freezeLength` | `+0x44`, `+0x30`, `+0x34` |
| `R.poison`, `R.poisonLength` | `+0x28`, `+0x2C` |
| `R.burnMagnitude`, `R.burnLength` | `+0x14`, `+0x18` (CSA-R10; LSE recorded the stack positions 2026-10-13) |
| `R.resultFlags.lethal` | bit `0x2` of the record byte at `+0x04` |
| unit type, class, id, mode | Unit `+0x00`, `+0x04`, `+0x0C`, `+0x10` |
| monster data, condition word | Unit `+0x14` -> MonsterData; word `+0x16` |
| seed `{lo,hi}` | Unit `+0x20`, `+0x24` |
| attacker game | Unit `+0x80`; Game `+0x1C` pool, `+0x6D` difficulty byte, `+0xA8` frame |
| `aliveExclusionCondition` | Unit byte `+0xC6` bit 0 |
| event chain | Unit `+0xDC`; node kind byte `+0x00`, successor `+0x24` |
| DifficultyLevels divisors | record `+0x18` MonsterColdDivisor, `+0x14` MonsterFreezeDivisor; stride `0x58`, clamped index |
| MonStats | `boss` bit 6 of `+0x0C`; `Velocity` i16 `+0x32`; `MonStatsEx` i16 `+0x18`; `ColdEffect` i8 `+0x168 + difficulty`; count at DataTables `+0xA80` |
| MonStats2 `deadCol` | bit 19 of record `+0x04`; count at DataTables `+0xA98` |
| state list / stat array | StatList state `+0x14`, expiry `+0x18` (STL owner) |

Native argument placement, which the named roles replace: stun `eax` attacker, `edi` defender, stack
length; cold `eax` attacker, `esi` defender, stack length; freeze `eax` defender, `ecx` attacker, stack
length; poison and burn stack `(attacker, defender, magnitude, length)` with callee cleanup of 16 bytes;
cold callback `ecx` unit, `edx` state, stack list.

## Rule rationale

- **CSA-R01.** `0xDE0C0` loads `+0x3C` into `edi` with the defender in `ebx` and calls `0xDA970`, then
  `+0x40` and `0xDA930`; then stun with `eax` = attacker, `edi` = defender, pushing `+0x44`; cold with
  `eax` = attacker, `esi` = defender, pushing `+0x30`; freeze with `eax` = defender, `ecx` = attacker,
  pushing `+0x34`; poison pushing `+0x2C`, `+0x28`, defender, attacker; burn pushing `+0x18`, `+0x14`,
  defender, attacker. It then reads stat 6 layer 0 and, on `jle` (signed nonpositive), skips the
  `and byte [record+4], 0FDh`. No step or draw helper appears in this slice.
- **CSA-R02.** `0x1A550` returns 0 for a non-type-1 unit, uses the supplied record or looks up the class
  (out-of-range returns 0), and returns the bit-6 mask (64). `0x22A80` returns 0 for null, nonmonster or
  absent MonsterData; otherwise ANDs word `+0x16` with the low word of its stack argument and converts
  nonzero to 1. `0x1A2D0` returns 0 unless type 1 and selects 1/2/3/4 for classes 271/338/359/560-561.
- **CSA-R03, CSA-R04.** `0xDCC20` tests `len <= 0` (signed) first. A null or non-type-1 defender jumps
  directly to the 250 cap. For a monster: if MonsterData is non-null and byte `+0x16` has bit `0x8`,
  one inline step of the attacker's `+0x20/+0x24` and unsigned `div` by 100; remainder `< 90` returns.
  Then #10064 `(0, defender)`, the class record from `0x1220` with an unguarded word read of `+0x32`,
  and #11104: nonzero with `len >= 13` stores 13 and skips the 250 test. Then #10871 for state 21,
  the null-attacker assertion, `expiry = attacker game frame + len`; no list: #11013 with flags 2 and
  attacker type/id, schedule kind 12, #10785 21, #10911 `0x9F7A0`, #10807 mode 1, #10945 21 on; list:
  #10161 then schedule kind 12.
- **CSA-R05, CSA-R06.** `0xDC780`: `len <= 0` returns; #10871 state 11 on the defender before the
  attacker assertion; `0xDAE20(defender, attacker game)`; zero returns. Monster with negative E: #10237
  on the attacker game's difficulty, divide by `+0x18` when nonzero (`cdq; idiv`). `len <= 1` stores 1.
  No list: #10945 11 on, #11013, schedule kind 12, #10785 11, #10911 `0xDB4C0`, #10807 mode 1, #10261
  for stats 67, 68, 69 with E at layer 0, then #10819. List: #10412, signed `jge` skips unless old is
  less, then #10161 and schedule. Both paths then step the defender seed inline, unsigned `div` 100;
  remainder `< 20`, non-null type-1 defender and `0x13B0(class, 19)` zero call #10945 `(defender,
  107, 1)`; every other outcome calls #10945 `(defender, 107, 0)`.
- **CSA-R07, CSA-R08, CSA-R09.** `0xDDAC0`: `len <= 0` returns; type 6 for null; type 0 jumps to the
  cold call; type 1: #10494 state 54 nonzero returns; then #10064 with the `0x1220` record, `0x22A80`
  with mask 8, #11104; any nonzero jumps to the cold call. Other paths: #10871 state 1, null-attacker
  assertion, `0xDAE20(defender, attacker game)`; `jge` returns for E >= 0; monster: #10237 and divisor
  `+0x14`, zero takes the assertion with line `0x250`, else `idiv` with no minimum. No list: #10945 1
  on, #11013, #10785 1, #10911 `0xDBD90`, #10807 mode 1. List: #10412; if not less, the old value
  replaces the candidate, else #10161. Common: schedule kind 12 at the resulting expiry. Then null or
  non-type-1 takes the line-`0x272` assertion. Monster: walk Unit `+0xDC`, loading the successor before
  calling `0x8E6F0` for kind-2 nodes, then schedule kind 2 at `frame + len + 1`.
- **CSA-R10, CSA-R11, CSA-R12.** `0xDCAB0` tests the length argument then the magnitude argument
  (`jle` each), calls #10871 state 2, asserts a non-null attacker, and for a non-null type-1 defender
  walks Unit `+0xDC` removing kind-3 nodes through `0x8E6F0` (successor loaded first) and schedules
  kind 3 at `frame + 1`. No list: #10945 2 on, #11013, schedule kind 12, #10261 stat 74 = negated
  magnitude, #10785 2, #10911 `0x9F7A0`, #10807 mode 1. List: #10680 stat 74 layer 0, negate, signed
  `jg` against the magnitude returns; else #11043, schedule kind 12, #10188 stat 74. `0xDC940` has the
  same sequence with state 115 and the same stat, callback and thunks.
- **CSA-R13.** `0xDB4C0` computes `alive` (non-null, byte `+0xC6` bit 0 clear, type 0 mode not 0/17 or
  type 1 mode not 0/12; other types are not alive). Not alive: #11061 `(unit, state)` nonzero returns.
  Then #10945 `(unit, state, 0)`; `alive` re-evaluated: #10945 `(unit, 107, 0)`; then #10819. No list
  removal and no draw.

<a id="contradiction-records"></a>
## Contradiction records

### Shatter request on the non-shatter path (CSA-R06)

- Affected ids: CSA-R06; vectors CSA-V02, CSA-V03, CSA-V08, CSA-V09, CSA-V10.
- Conflict: NRF (rows of NRF-R04a and NRF-R13a, read 2026-10-12/13/14) stated only that the cold draw
  sets state 107 when the shatter condition holds; its vectors NRF-V08, NRF-V18, NRF-V21 and NRF-V30
  recorded no state 107 request for a failed condition. LSE-R19 (read 2026-10-13) stated that every
  other outcome calls the state setter with state 107 off.
- Same inputs and boundary: both readings cover the same function and the same branch; no profile or
  capture question applies.
- Settled 2026-10-14 by rereading the tail of D2Game `0xDC780`: the failed-condition target pushes
  `(0, 0x6B, defender)` and calls thunk `0xA3CA` (#10945). LSE is correct. CSA-R06 states both requests.
  The NRF vectors' existing values are unchanged; their CSA conversions add the state 107 off request
  as an established output. LSE-V12, LSE-V13, LSE-V17 and LSE-V20 already carried it.

### Null defender at the cold applier (CSA-R05)

- Affected ids: CSA-R05 and the fault row of section 5.
- Conflict: LSE-R11a (2026-10-13) said a null defender would be treated as a nonmonster with E = -50 and
  fault at the shatter step. NRF-R12a (2026-10-14) states that finding a state list requires a non-null
  unit, and NRF-R13d places a null stun defender's fault at the list lookup.
- Settled 2026-10-14: `0xDC780` calls #10871 for the defender before the attacker test and the effect
  getter; #10871 (`0x38520`) dereferences the unit's stat-list field with no null test. The fault occurs
  at the list lookup. Both readings agree the path is not a successful fixture; only the fault point
  changes. The same lookup precedes every applier's later work, so CSA states it for all five.

### One monster condition, two neutral names (CSA-R02, CSA-R03, CSA-R07)

- NRF named the stun-cancel condition `stunCancelCondition` and the freeze test `monsterGate(unit, 8)`
  separately; LSE-R12b said they were the same flag. Not a disagreement about behavior, but NRF-V18
  supplied `stunCancelCondition = true` together with `monsterGate = 0`.
- Settled 2026-10-14: `0xDCC20` tests byte MonsterData `+0x16` with `0x8` after a MonsterData null test;
  `0x22A80` with mask 8 tests word `+0x16` AND 8 after the same type and null tests. Both select the
  same bit. CSA names it once (`monsterFlag8`). NRF-V18's combination never reached the freeze gates
  (freeze length 0), so its values are unchanged; CSA-V02 states `monsterFlag8` true, and its freeze
  control case sets it false explicitly.

### Agreement checks without conflict

The following were compared rule by rule and agreed: stun gate order, caps and request order
(NRF-R13d, LSE-R10a/R10b/R18); cold divisor only for monsters with negative E, minimum 1 after
division, strict extension and draw on every application (NRF-R13a, LSE-R11a/R19); freeze routing,
state-54 rejection, gate order, no minimum, zero divisor fatal, timer 12 on a retained expiry and the
kind-2 tail (NRF-R04d/R13a/R13e, LSE-R12a/R12b/R20); poison and burn length-then-magnitude gates,
pre-step, magnitude comparison and unconditional expiry replacement (NRF-R13c/R15b/R15c,
LSE-R13a/R14a/R14b/R21); hand-off order (NRF-R15a, LSE-R16); ColdEffect -50 for invalid records
(NRF-R13a correction 2026-10-14, LSE-R11b). LSE-R12c's claim that the scheduler calls a further helper
for a kind-2 request on a state-54 monster concerns the scheduler, which the lifecycle owns; CSA-R07
returns for state 54 before its kind-2 request, so that edge is reachable only if a dependency sets
state 54 in between. It stays in CSA-Q02.

## Reused evidence and table records

- LSE's applier readings (2026-10-13), including its byte-run checks for `0xDAE20`, the freeze tail and
  the scheduler head, are reused for CSA-R03 to CSA-R12 with their dates. NRF's 2026-10-14 independent
  listings and corrections (stun monster-only branch, invalid-monster cold effect, gate helpers) are
  reused for CSA-R02 to CSA-R09. NRF's cold-callback reconstruction reading is reused for CSA-R13 and
  was reread on 2026-10-14.
- Divisor values used by vectors reuse the
  [DifficultyLevels named cell records](core-game-data-fields.md#named-cell-records) (`TD-DIFFICULTY`:
  MonsterColdDivisor 1/2/4 and MonsterFreezeDivisor 1/2/4 for Normal/Nightmare/Hell, compiled words
  agree) and NRF's [difficulty cells](core-numeric-and-rng-foundations.md#difficulty-divisors-and-identities).
  Their identity limits are stated in those records. Per-class `ColdEffect`, `Velocity`, `boss` and
  `deadCol` are supplied vector inputs; no MonStats cell is claimed.
- Class names 271 `roguehire`, 338 `act2hire`, 359 `act3hire`, 560 `act5hire1`, 561 `act5hire2` reuse
  GDF-R17's records.

## Search records

No rule is labeled `not found in d2-decomp`. Rungs used (LSE 2026-10-13, reused): root anchor through
the do-function and missile tables (DPO-R08d) and the call lines of `0xDE0C0`; output anchor on the
state, stat and timer-kind immediates pushed before the state setter, list-stat writers and scheduler;
data anchor on MonStats `+0x0C`, `+0x32`, `+0x168`, `+0x18` and DifficultyLevels `+0x14`, `+0x18`;
string anchor on the empty assert file string and its line numbers; decompile view only to check byte
runs; runtime not armed. The 2026-10-14 rereads are targeted readings, not new searches.

## Identifier mapping

This mapping lets the reference updates in NRF, LSE and their companions and gap files point to the
new owner. "Stays" means the item remains in its spec.

### From core-numeric-and-rng-foundations

| Old id | New owner |
|---|---|
| NRF-R04a, stun-cancel and cold-shatter rows | Stay in NRF-R04a as references to CSA-R03 and CSA-R06; the shatter-condition sentence below the table moves to CSA-R06 |
| NRF-R04d | CSA-R07 |
| NRF-R13 (preamble: boundary, timer policy, fixtures) | CSA section 4 preamble, section 8.1 defaults, CSA-Q01 |
| NRF-R13a, cold part | CSA-R05 (draw in CSA-R06) |
| NRF-R13a, freeze part | CSA-R08 (type assertion in CSA-R09) |
| NRF-R13b | CSA-R13 |
| NRF-R13c | CSA-R10, CSA-R12 (pre-step CSA-R11) |
| NRF-R13d | CSA-R03, CSA-R04 |
| NRF-R13e | CSA-R09 |
| NRF-R13f | CSA-R02 |
| NRF-R15a | CSA-R01 (drain arithmetic stays LSE-R17 and DPO-R18) |
| NRF-R15b | CSA-R10, CSA-R12 |
| NRF-R15c | CSA-R11 |
| NRF-V08 | CSA-V08 (a) to (d) |
| NRF-V09 | CSA-V14 (a) to (c), (e), (f) |
| NRF-V12 | CSA-V05 (a) to (e), CSA-V06 (a) to (c) |
| NRF-V18 | CSA-V02 |
| NRF-V19 | CSA-V13 (a) to (c), (e) |
| NRF-V21 | CSA-V07 (a) to (c); cold contrast CSA-V08 (e), (f) |
| NRF-V23 | CSA-V15 (a) to (f) |
| NRF-V30 (c) | CSA-V10 (a) to (c); the remaining NRF-V30 cases stay for NRF-R04b, relabelled (a) to (d) |
| NRF-V33 (b) | CSA-V01; the remaining NRF-V33 cases stay for NRF-R12a, relabelled (a) to (f) |
| NRF-V34 (a), (c), (d) | CSA-V11 (a) to (c) |
| NRF-V34 (b) | CSA-V12 |
| NRF-V35 | CSA-V04 |
| NRF-O04 | CSA-O01 |
| NRF-Q09 (status conditions) | CSA-Q03 |
| NRF-Q17 (animation refresh, timer-node and handler parts for status) | CSA-Q02 |
| NRF-Q20 (status attach/write parts) | CSA-Q01 |
| NRF-Q21 (status rules and NRF-O04) | CSA-Q05 |
| NRF section 3 rows: status gates, status dependency operations, cold/freeze divisors, monster ColdEffect, status states/stats/timer kinds | CSA section 3 |
| NRF gap rows "Animation and update bookkeeping", "Timer-node removal and downstream handlers", "Actual attach and replacing-write integration", "Wider runtime-condition meanings", "Gameplay confirmation" (status parts) | CSA gap rows of the same kinds |

### From combat-leech-and-status-effects

| Old id | New owner |
|---|---|
| LSE-R04b (hireling predicate) | Stays for leech; its stun and freeze uses are CSA-R02 `unitGate` |
| LSE-R10a | CSA-R03, CSA-R04 |
| LSE-R10b | CSA-R02 (boss and class-group helpers), CSA-R03 (cap 13) |
| LSE-R11a | CSA-R05 (null-defender fault point settled above) |
| LSE-R11b | CSA-R05 (cold effect), CSA-R06 (`deadCol` test) |
| LSE-R12a | CSA-R07, CSA-R08, CSA-R09 |
| LSE-R12b | CSA-R07, CSA-R02 |
| LSE-R12c | Scheduler edge: STL owner, CSA-Q02; frame floor: CSA-R08 note |
| LSE-R13a | CSA-R10, CSA-R11, CSA-R12 |
| LSE-R13b, LSE-R13c | Stay in LSE (immediate total) |
| LSE-R14a | CSA-R10, CSA-R11, CSA-R12 |
| LSE-R14b | CSA-R10 |
| LSE-R16 | Leech, absorb heal and life subtraction order stay in LSE; the status part from the mana drain onward is CSA-R01 |
| LSE-R18 | CSA-R03 |
| LSE-R19 | CSA-R06 |
| LSE-R20 | CSA-R07 |
| LSE-R21 | CSA-R01 (draw statement), CSA-R12 |
| LSE-V11 | CSA-V05 (e) to (h) |
| LSE-V12 | CSA-V09 |
| LSE-V13 | CSA-V10 (c), (f); CSA-V11 (a), (d) |
| LSE-V14 | CSA-V13 (d), (f); monster case CSA-V12 |
| LSE-V16 | CSA-V14 (a), (b), (d) |
| LSE-V17 | Leech portion stays; status portion CSA-V03 |
| LSE-V19 | CSA-V06 (d), CSA-V07 (d) to (f) |
| LSE-V20 | CSA-V10 (d), (e) |
| LSE-O03 | CSA-O02 |
| LSE-Q08, LSE-Q09 (resolved) | Answers folded into CSA-R12 and CSA-R06 |
| LSE-Q10 (status part) | CSA-Q04 |
| LSE-Q12 | CSA-Q02 |
| LSE-Q13 | CSA-Q03 |
| LSE section 5 rows for stun, cold, freeze, poison, burn and their fatal paths | CSA section 5 |

## Vector conversion record

Vectors take the template shape. Values are the source vectors' values; the following assertions were
added or restated, each established by the cited rule and evidence:

- State 107 off requests in cold cases whose shatter condition fails (contradiction record above).
- NRF-V12 and NRF-V30 supplied gate results (`recordGate` 64, `recordGate` 1) are expressed through the
  class record's `boss` flag, which gives 64; routing depends only on nonzero.
- NRF-V18's freeze-only control states its full request list (it pointed to NRF-V34 for them):
  40 / 2 = 20, expiry 1020, timer 2 at 1021.
- LSE-V12's `len = 1` case and existing-list case state their shatter outcome and candidate expiry
  (1006), computed from CSA-R05/CSA-R06 with the vector's seed; LSE gave only the expiry.
- LSE-V17's status portion starts from the attacker seed after the leech draw, `{0x6AC690C5, 0}`, and
  adds the final lethal decision with a supplied positive life.
- NRF-V34 (b) asserted the pre-step alone; CSA-V12 adds the following no-list requests and a player
  contrast. CSA-V15 (g) adds an alive unit with a nonzero stay-death result. LSE-V19 and LSE-V20 class
  ids `0x152` and `0x167` are written 338 and 359. MonStats counts are stated wherever a vector uses a
  class other than 0.
- Hand computations rechecked on 2026-10-14: `{1,0x29A}` -> `{0x6AC6935F,0}` (51), then
  `{0x2F2ED81B,0x2C890AFD}` (31); `{5,0x29A}` -> `{0x15E0D673,2}` (99); `{0,95}` -> `{95,0}`;
  `{1,0}` -> `{0x6AC690C5,0}` (85); `{0x6AC690C5,0}` -> `{0x669A3799,747178471}` (9); `{0,5}` ->
  `{5,0}`; `{5,0}` -> `{0x15E0D3D9,2}` (33); `{0,50}` -> `{50,0}`.

<a id="observation-procedures"></a>
## Observation procedures

Both are pending. Use an owned matching original installation and a read-capable debugger or
instrumentation tool. Record module fingerprints, the effective data profile, date, inputs, frames,
seed pairs and outputs. Add each RVA to the loaded module base. Offsets locate evidence; they are not
required layouts. Keep captures private and publish narrow measurements.

### CSA-O01: cold extends only

Use a Normal monster that survives three positive cold hits with known adjusted lengths and nonzero
ColdEffect. Capture entry to `D2Game.dll+0xDC780`, the defender seed at Unit `+0x20/+0x24`, the
current frame (attacker game `+0xA8`), the length argument and the state 11 list expiry (Unit `+0x5C`
-> primary chain -> list with state `+0x14` equal to 11, expiry `+0x18`).

1. First hit: expiry becomes `frame1 + adjustedLength1`; defender seed steps once.
2. Second hit with candidate `frame2 + adjustedLength2` not later than the existing expiry: expiry
   unchanged; defender seed still steps once.
3. Third hit with a strictly later candidate: expiry raised; defender seed steps once.

Choose sources and timing from measured inputs; exclude kills, freezes and unrelated seed draws.

### CSA-O02: poison is replaced, not extended

Normal difficulty, Blood Moor. Attack one zombie with a long-duration poison source, then, within that
duration, with a stronger poison source of shorter duration. Capture entry to `D2Game.dll+0xDCAB0`
with its four arguments and read the defender's state 2 list (state `+0x14` equal to 2, expiry `+0x18`)
and its stat 74 value. Expected: after the first hit, stat 74 = `-magnitude1` and expiry
`frame1 + length1`; after the second, stat 74 = `-magnitude2` and the shorter `frame2 + length2`.

## Runtime records

| Vector | Source profile | Date | Module + RVA | Records | Result |
|---|---|---|---|---|---|
| none | — | — | — | 0 | No observation was made for this spec. LSE's 2026-10-13 runtime note recorded a member with a dead character and no hit; it observed no vector. |

## Creation record (2026-10-14)

- Checks run: the spec lint on this spec, companion and gap file (result reported with this change).
- Gaps closed by this pass: the shatter-off request and the null-defender fault point (contradiction
  records); the single monster condition name.
- Gaps left open: [gap file](../../plans/gaps/combat-status-application.md) (integration of stored
  dependency effects, timer-node and handler execution, outside-boundary condition producers and client
  messages, pending gameplay observation).
- Labels: every rule carries `verified in 1.13c code (static)`, matching its NRF and LSE source labels;
  no promotion.
- Blind-implementer pass: complete on 2026-10-14. Full runs 1 to 5 and delta checks 1 and 2, with
  results and triage below; delta check 2 returned MISMATCH 0 and no unresolved GUESS, GAP or DISCRIM.
- Status: draft.

## Blind-implementer pass 1 (2026-10-14)

Fresh Sonnet reader given the spec and its cited dependency specs only. Counts: MISMATCH 0, GUESS 5,
GAP 4, DISCRIM 11 (plus one note on CSA-V03), GATED 14, COVERAGE 36, HARDEST 3. Outcomes, all spec
edits unless stated:

| Entry | Outcome |
|---|---|
| GUESS 1-4 (persistence across V03's second cold, V08 (b)-(d), V08 (e) start state, V02 (b) inheritance) | Section 8.1 preamble now states case structure ("after (x)", "Fresh (x), then") and that supplied list operations persist their stored effects within a vector; V02 (b), V03 trace, V08 (b) and (e) state their starting state |
| GUESS 5 (stat list for the life read; lethal mask) | Preamble names `D.statList` for `Stat(D, 6, 0)`; V01, V02, V03 put own stat 6 in `D.statList`; section 3 names `lethal` as the 0x0002 bit |
| GAP 1 (lethal bit only in V01) | Section 3 row |
| GAP 2 (helper results not in section 5) | Section 5 row for CSA-R02 |
| GAP 3 (which operations log) | Preamble lists logged and unlogged operations; `FindStateList` logs; request lists omit the one lookup per applier |
| GAP 4 (removed-node successor) | Preamble: removal unlinks and clears the removed node's successor (a fixture choice; real removal is CSA-Q02) |
| DISCRIM 1 (old low word) | V05 (f) named as the rejecting case (old low word 0 would cancel) |
| DISCRIM 2 (stun threshold) | V05 (i) roll 89 cancels, (j) roll 90 stuns |
| DISCRIM 3 (shatter threshold) | V09 (h) roll 20 off, (i) roll 19 on |
| DISCRIM 4 (cold/freeze divisor swap) | V09 (k) and V11 (e): supplied unequal row, cold 3 and freeze 5, marked as fixture input |
| DISCRIM 5 (state 54 before player routing) | V10 (g): player with state 54 still gets cold |
| DISCRIM 6 (`E == 0` vs `E >= 0` in freeze) | V11 (f): E +10 gives no freeze |
| DISCRIM 7 (pre-step after acceptance) | V12 (b): rejected stronger list still gets the kind-3 removal and timer |
| DISCRIM 8 (lookup before gates) | Lookups now log; V13 (e), V14 (a), (b) assert no logged lookup |
| DISCRIM 9 (boss/Velocity/cap order) | No change to rules: order is unobservable in the fixture (unlogged reads, effect-free returns); V06 Discriminates corrected to claim only lookup-after-checks |
| DISCRIM 10 (0x10008 claim false for a u16 word) | V04 Discriminates corrected; the case rejects no alternative |
| DISCRIM 11 (shared not-alive set) | V15 (h) monster mode 17 alive, (i) player mode 12 alive |
| V03 swapped applier order | V02 and V03 Discriminates cite CSA-V01 as the order check |
| COVERAGE (advisory) | Added V07 (g)-(i) for the 250/251 and 13 boundaries and V09 (j) for positive E; the other entries are branches whose result is already shown by an existing case or a documented fault path, so no change |
| HARDEST (V03, V02, V08) | Addressed by the preamble persistence and case-structure fix |

All added expected values follow from CSA-R01 to CSA-R13 by hand computation:
`{0,89}` -> `{89,0}` (89), `{0,90}` -> `{90,0}` (90), `{0,20}` -> `{20,0}` (20), `{0,19}` ->
`{19,0}` (19). Vectors, inputs and assertions changed, so a fresh reader pass is required.

Follow-up conversion (2026-10-14, template section 8.1 update): the preamble's case-chaining rule was
removed; every case that relied on another case's state or outputs (V02 (a), V05 (c), (d), (g),
V08 (b) to (f), V09 (b), (c), (h), (i), V10 (b), (e), (g), V12 (a), (b), V14 (d)) now lists its
starting state among its own inputs and states its own outputs. V08's second 1200-list case became
V08 (f). Expected values are unchanged.

## Blind-implementer pass 2 (2026-10-14)

Fresh Sonnet reader, same inputs. Counts: MISMATCH 0, GUESS 2, GAP 4, DISCRIM 7, GATED 14,
COVERAGE 23 (advisory), HARDEST 3.

| Entry | Outcome |
|---|---|
| GUESS 1 (fixture `Schedule` did not link nodes, unlike STL-R01) | Spec edit and correction, below: the fixture now links each scheduled node at the chain head; CSA-R11 states the in-hit consequence; V02, V03, V11 (a), V12 gain the resulting removal request or chain assertions |
| GUESS 2 / GAP 2 (V10 seeds) | Every V10 case states both seeds and their final values |
| GAP 1 (defender id) | Preamble gives `D` id 9, so owner 0/7 is distinguishable from the defender |
| GAP 3 (difficulty of the supplied row) | V09 (k) and V11 (e) state that `A.game.difficulty` 1 selects the supplied row |
| GAP 4 (`GetStat` inputs) | Preamble: stated lists have `usesResolvedStats` false and own-array stats; definition 74 `fMin` false |
| DISCRIM 1 (lethal toggle/assignment) | V01 (c) lethal clear with life 256 and (d) lethal clear with life 0 both stay 0x0008 |
| DISCRIM 2 (gates before the cancel draw) | V06 (e) flagged boss monster: A steps, cancelled; (f): A steps, not cancelled, boss returns |
| DISCRIM 3 (state 54 after gates) | V10 (h) boss monster with state 54: no request, no draw |
| DISCRIM 4 (class groups through appliers) | V07 (j) class 560 stun capped 13; V10 (i) class 271 freeze falls back to cold |
| DISCRIM 5 (only first kind-3 removed) | V12 (c) chain with two kind-3 nodes; both removed |
| DISCRIM 6 (fixed difficulty column) | V09 (l) Normal and (m) Hell with distinct columns -40/-50/-60 |
| DISCRIM 7 (attacker vs defender game) | V09 (n): the defender is given a different game (fixture input); outputs follow the attacker's game. This is a case, not a "one game per hit" claim, which was not established here |
| COVERAGE (advisory) | No further vectors: the remaining entries repeat branches whose result an existing case already shows, or are documented fault paths |
| HARDEST | Covered by the fixes above |

### Correction: nodes scheduled earlier in the same hit (CSA-R11)

- Affected ids: CSA-R11; CSA-V02, CSA-V03, CSA-V11 (a), CSA-V12.
- Original discrepancy: NRF-V18 and NRF-V34 used a supplied scheduler that logged requests without
  adding nodes to the traversed chain, so NRF-V18 listed two kind-3 requests and no removal, and
  NRF-V34 (a) gave a remaining chain of R only. CSA's first draft copied that fixture.
- Evidence (2026-10-14): STL-R01 states that each scheduled record is linked into the unit's chain,
  newest first. Rereading confirmed it: the scheduler D2Game `0x8EA80` allocates through `0x8E680`,
  which stores the unit's chain head (Unit `+0xDC`) as the new node's successor (`+0x24`) and makes
  the new node the head; the kind byte is then written. Poison `0xDCAB0` schedules kind 3 on the
  defender, and burn `0xDC940` walks the same defender chain before its own request, so burn's
  pre-step requests removal of poison's kind-3 node in the same hit. Kind-12 nodes are linked but
  never match the kind-2 or kind-3 walks.
- Corrected assertions: V02 and V03 add one removal request (burn removes poison's kind-3 node);
  V02, V11 (a) and V12 state final chains including the linked `K12`/`K3`/`K2` nodes. Every other
  expected value is unchanged. Whether a removal of an active node defers rather than unlinks stays
  in CSA-Q02.

## Blind-implementer pass 3 (2026-10-14)

Fresh Sonnet reader. Counts: MISMATCH 1, GUESS 2, GAP 4, DISCRIM 5, GATED 14, COVERAGE 17
(advisory), HARDEST 3.

| Entry | Outcome |
|---|---|
| MISMATCH (V06 "Draws: none" against cases (e), (f)) | Draws stated per case |
| GUESS 1 (cold callback with a null unit) | CSA-R13 now states that the lifecycle calls a callback only for a list with a unit (STL-R05), and what a direct null call would do: consult stay-death, then a null state request outside the successful boundary (NRF-R12a). Basis: the 2026-10-14 reread of `0xDB4C0` (null takes the not-alive branch, skips the shatter clear) |
| GUESS 2 (direct call or hand-off) | Preamble "Entry" line: V01 to V03 use the hand-off; V04 to V15 call directly |
| GAP 1 (definition 6 only in V01) | Added to the shared defaults |
| GAP 2 (V05 (e) seed) | Seed and unchanged final value stated |
| GAP 3 (supplied difficulty row) | Preamble line declares such rows as fixture input selected by `A.game.difficulty` |
| GAP 4 (condition producers) | No change: CSA-Q03 |
| DISCRIM 1 (player E via record) | V08 (g): player with class 0 ColdEffect -30 gets stats -50 |
| DISCRIM 2 (fallback divisor) | V10 (j): boss monster, supplied row cold 3 / freeze 5, length 30, expiry 1010 |
| DISCRIM 3 (defender game beyond cold) | Preamble statement that every applier reads only `A.game`, plus V05 (k) for stun beside V09 (n) for cold |
| DISCRIM 4 (`len >= 13` vs `> 13`) | No change: both give 13 at 13; equivalent |
| DISCRIM 5 (life read before appliers) | No change: CSA-V01's log order already rejects it |
| COVERAGE | No change: the remaining entries repeat branches an existing case shows, or are documented fault paths |
| HARDEST | Covered by the entries above |

Lint decision (2026-10-14): after phase 2 removed the moved NRF ids, the lint reports 45 id-ref
WARNs, all in this companion's records (identifier mapping, contradiction records, vector conversion
and correction records) naming the old NRF ids those records describe. They are kept as records;
the gap file's one live reference to a removed id was rewritten.

## Blind-implementer pass 4 (2026-10-14)

Fresh Sonnet reader. Counts: MISMATCH 0, GUESS 3, GAP 7, DISCRIM 5, GATED 14, COVERAGE 14
(advisory), HARDEST 3.

### Contradiction: lethal bit for nonpositive final life (CSA-R01)

- Affected ids: CSA-R01, CSA-V01; related DPO-R19b.
- Conflict: CSA-R01 (built from NRF-R15a, "otherwise retains it") said the bit is unchanged when the
  final life is not positive. DPO-R19b says a nonpositive final life sets result bit `0x2`. NRF-V33's
  cases did not distinguish the two, because the bit was already set in every nonpositive case.
- Settled 2026-10-14 by rereading D2Game `0xDE0C0` after the burn call: the stat 6 read is followed by
  `jle` to `or byte ptr [record+4], 2`; the positive path executes `and byte ptr [record+4], 0FDh`.
  The hand-off itself sets the bit for nonpositive life and clears it for positive life; DPO-R19b is
  correct. CSA-R01 now states both halves. No owner fix is needed in DPO; the NRF wording that CSA
  inherited was the error.
- Corrected expected value: CSA-V01 (d), added in pass 2 as "lethal clear, life 0, stays 0x0008",
  now expects 0x000A. That case was derived from the wrong CSA-R01 wording, not from a source vector;
  no NRF or LSE vector value changes.

### Contradiction: difficulty-record clamp citation (section 3)

- Affected ids: CSA section 3 divisor row, CSA-R05, CSA-R08; new CSA-Q06.
- Conflict: section 3 said the divisor row comes from "the clamping difficulty-record lookup
  (GDF-R02)"; GDF-R02 defines rows 0..2 and no clamp, and GDF section 6 says no substitute exists
  outside 0..2.
- Evidence: the getter D2Common `0x603C0` (#10237), checked reconstruction reread 2026-10-14, clamps
  its index into `[0, count - 1]`. But for a monster defender the cold-effect read indexes the
  per-difficulty column directly with no guard, and divisors are used only for monsters, so the clamp
  cannot produce a successful result on its own outside 0..2.
- Resolution: CSA now covers difficulty 0..2 only and cites GDF-R02 for the row identities; the range
  of the difficulty producer is CSA-Q06 (outside-boundary). The clamp remains recorded here as
  evidence.

### Outcomes

| Entry | Outcome |
|---|---|
| GUESS 1 (V07 Velocity) | Inputs say the Velocity applies to the class record each case names |
| GUESS 2 (chain order) | Shared defaults: event chains are listed head first |
| GUESS 3 (V05 per-case flag) | V05 (f), (g), (i), (j) restate `monsterFlag8` true |
| GAP 3 (flags 2, mode 1, 64) | Section 3 row gives each meaning: timed list (STL-R04, STL-R06a), apply-all-stats attach mode (STL-R06b, CSA-Q01), boss bit value |
| GAP 4 / DISCRIM 5 (V03 stat 74) | Added to V03 Outputs |
| GAP 5 (final chains) | No change: chain assertions use named nodes; section 5 already lists the removal and timer requests |
| GAP 6 (deadCol lookup inputs) | Defaults give `MonStatsEx` 0 and MonStats2 count 1 |
| GAP 7 (binary details in STL; section 2 evidence sentence) | STL details: no change here (coordinator follow-up). The empty-assert-string sentence moved out of section 2; the basis stays in "Owning category" above |
| Other GAPs | Folded into the entries above |
| DISCRIM 1 (whole-point life test) | V01 (f): own stat 6 = 128 clears the bit; V01 (d), (e) show the set half |
| DISCRIM 2 (no-list cap) | V08 (h): no-list cold with length 300 expires at 1300 |
| DISCRIM 3 (pre-step before gates) | V13 (g) and V14 (g): monster with a kind-3 node and a failing gate gets no lookup, removal or kind-3 request |
| DISCRIM 4 (other appliers' game) | V11 (g) freeze and V13 (h) poison with a defender in a different game |
| COVERAGE | No change: remaining entries repeat branches an existing case shows, or are documented fault paths |
| HARDEST | No change beyond the above |

Lint after this pass: 0 errors. The remaining warnings are id-ref warnings in this companion's
records, which name NRF and LSE ids that phase 2 removed or renumbered (mapping, contradiction and
conversion records); they are kept as records. The gap file's live references were rewritten to
current ids.

## Blind-implementer pass 5 (2026-10-14, full run)

Fresh Sonnet reader, full review. Counts: MISMATCH 0, GUESS 0, GAP 4, DISCRIM 2, COVERAGE 20
(advisory), HARDEST 3. This is the last full run; under checklist section 2 the next reader run is
a delta check over the changes listed here.

| Entry | Outcome |
|---|---|
| GAP 1 (V01's `Stat(D, 6, 0)` log entry not in section 5) | No change: a read, logged as a named intermediate of the hand-off fixture; it writes nothing |
| GAP 2 (final-chain assertions in V02, V03, V11, V12) | No change: named intermediates, defined by the preamble's `Schedule` (links each new node at the head) and `RemoveTimerNode` (unlinks and clears the successor) fixtures |
| GAP 3 (V02 (b) lethal output) | V02 (b) states `Stat(D, 6, 0)` = 0, so `lethal` stays set |
| GAP 4 (vector-local count and Velocity overrides) | No change: per-case and per-vector Inputs are where overrides belong |
| DISCRIM 1 (`StayDeath` arguments) | V15: the supplied `StayDeath` logs its arguments; (b), (c), (e), (f) assert `StayDeath(D, 11)`; (a), (d), (g), (h), (i) assert that no call is logged (the unit is alive there, so CSA-R13 does not consult it) |
| DISCRIM 2 (class 561) | V07 (k): class 561, length 100, capped 13 |
| COVERAGE (flagged monster that passes the cancel draw, then returns on Velocity 0) | Added V06 (g): V06 (f) showed the same outcome only for the boss return; the other entries: no change, they repeat branches an existing case shows or are documented fault paths |
| HARDEST | No change |

Changes since this full run, for the delta reader: V02 (b) lethal output (GAP 3); V06 Draws line
and new case (g) (COVERAGE); V07 Inputs (class 561 record) and new case (k) (DISCRIM 2); V07
Discriminates; V15 Inputs (`StayDeath` logs) and cases (a) to (i) `StayDeath` assertions (DISCRIM 1);
V06 and V15 Discriminates. No rule text changed.

## Delta check 1 (2026-10-14)

Fresh reader over the pass-5 change list. Result: MISMATCH 0, UNRESOLVED 0, GUESS 0, GAP 0,
DISCRIM 1; all four pass-5 findings confirmed resolved. Remaining DISCRIM: class 359 (group value 3)
was never reached through the stun applier. Outcome: V07 Inputs add a class-359 record (MonStats
count 562 already exceeds 359); new case V07 (l) (class 359, length 100, capped 13, new-list requests
at 1013); V07 Discriminates now list (c), (d), (l), (j) and (k) as covering group values 1, 2, 3 and 4.
The next reader run is delta check 2 over these changes.

## Delta check 2 (2026-10-14)

Fresh reader over the delta check 1 changes (V07 Inputs, case (l), Discriminates). Result:
MISMATCH 0, UNRESOLVED 0, GUESS 0, GAP 0, DISCRIM 0; case (l) recomputed with the expected request
list. The blind-implementer pass is complete: full runs 1 to 5, delta checks 1 and 2. Status stays
draft because CSA-Q01 to CSA-Q06 and the pending observations remain.
