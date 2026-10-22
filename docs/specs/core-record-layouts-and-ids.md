# Record Layouts and Ids

## Description

Characters, monsters, missiles and items each carry a record of their values, active effects and current
action. These shared rules define how a stored value is read, how an active effect is looked up and switched on
or off, how a change of action clears effects tied to the old action, and how the names of values, effects and
actions match the game's data. The damage record that combat hands from one stage of a hit to the next is also
defined here, so every combat system reads the same fields.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: RLI-R, RLI-Q, RLI-V; policy in [README conventions](README.md#conventions)
Evidence: [source, table and review records](../research/spec-evidence/core-record-layouts-and-ids.md)
Open evidence gaps: [remaining boundaries and observations](../plans/gaps/core-record-layouts-and-ids.md)

## 1. Scope

This spec defines the shared semantic records and identities consumed across gameplay specs: the logical record
structures, the original-profile unit type codes, the shared damage record, and the correspondence of stat,
state and mode names to identities; reads of existing stat arrays with their conditional minimum adjustment;
state-list lookup, state test and state-set routing; and unit mode change with its mode-bound list processing.
Random draws, seeding, integer arithmetic, fixed-point and time units belong to
[core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); the object animation draw reached from
the mode setter's animation reset is NRF-R13. Name tables and per-table column semantics belong to
[core-game-data-fields](core-game-data-fields.md). Timed status application belongs to
[combat-status-application](combat-status-application.md). Building resolved stat arrays, attached-list
aggregation, complete timer/list internals, disguise and update-queue effects, animation selection and
networking retain their dependency boundaries.

Notation: integer widths, wrapping, signed division and fault conventions are those of
[core-numeric-and-rng-foundations section 1](core-numeric-and-rng-foundations.md#1-scope). Illustrative
operation names impose no language, API or storage design.

## 2. Trigger and ownership

- D2Common owns the stat getters, the state helpers and the mode setter. Calls run when the owning gameplay
  operation reaches them; these operations do not execute independently.
- Status application ([combat-status-application](combat-status-application.md)), damage mitigation and the
  damage pipeline consume the stat reads, the state routing and the damage record; their placement within the hit
  is owned there.
- Mode changes call list removal before animation reset.
- Reads emit no client message. State, mode and list changes reach clients through their owning lifecycle and
  networking systems.
- Full dispatch/output anchors and the source mapping of each semantic operation are in the linked companion.

## 3. Inputs and data

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| entity identity | `unit.type`, `unit.classId`, `unit.id`, `unit.mode` | i32 type/class/mode; u32 id | Original-profile types: 0 player, 1 monster, 2 object, 3 missile, 4 item, 5 tile; modes named by GDF-R05/GDF-R06 |
| stat and state containers | `unit.statList`, lists and associated owner | logical record or null | Fields and arrays below; null getter inputs return zero according to RLI-R02a |
| state eligibility | `stateCount`, `list.hasStateStorage`, active and changed state sets | i32 count; boolean; bits by state id | The original profile has 185 states; GDF-R03 table evidence |
| list chains | primary and secondary attached-list sequences; `list.state` | ordered lists; i32 state | Lookup tests primary before secondary; mode-bound removal uses primary only |
| list conditions | `expiryFlag`, `removeOnModeChange`, `usesResolvedStats`, `hasStateStorage` | boolean conditions | The original extended-list condition is shared by hasStateStorage and usesResolvedStats; these names denote the same original-profile condition. Other original flags are fixed by code; no bit layout is required |
| mode condition | `unit.modeChanged` | boolean | Written by RLI-R04a; a neutral name whose wider role is RLI-Q01 |
| combat record | semantic fields in section 3.1 | i32 quantities unless stated | Shared record consumed by the damage pipeline and status application; no rule here computes with it |
| dependency operations | state-bit set/clear, update mark, list free, animation reset | supplied operation contract | Ordered requests (RLI-R03, RLI-R04a, RLI-R04b) are distinct from dependency-produced stored effects |
| state and mode arguments | state identity `s`, state-write value `on`, requested mode | i32 | Distinct from current membership/mode; any nonzero `on` requests active membership |

| Mapping / constant | Origin | Binding |
|---|---|---|
| Stat/state/mode/class names and definition counts | Original-profile tables, with exact files and identity conventions in GDA-R12/GDF-R02 to GDF-R09 | Data names correspond to the original compiled identities; number-to-operation uses are fixed separately in code; correspondence does not prove configurable remapping |

### 3.1 Damage record

| Field | Width / units | Consumer |
|---|---|---|
| `flags`, `resultFlags` | u32, u16 | Pipeline owns masks; `resultFlags.lethal` is used by CSA-R01 |
| `physical`, `fire`, `lightning`, `magic`, `cold`, `poison`, `burnMagnitude` | i32 8.8 | Damage and status consumers |
| `burnLength`, `poisonLength`, `coldLength`, `freezeLength`, `stunLength` | i32 frames | Corresponding status applier ([combat-status-application](combat-status-application.md)) |
| `lifeLeech`, `manaBurn`, `staminaBurn` | i32 raw quantity | Drain consumers |
| `absorbHeal`, `total` | i32 8.8 | Mitigation and pipeline |
| `reductionScale` | i32, 0 means none; 1024 means 1.0 | Mitigation |
| `hitClass`, `reservedCondition`, `conversionEnabled`, `conversionPercent` | i32 hit-class value; u8 conditions; i32 percent | Pipeline owns meanings and conversion |

Cold/freeze divisors are DifficultyLevels inputs of the status appliers, never fields of this combat record.
Duration fields are frames rather than damage units. This is a semantic contract, not an internal binary layout.

### 3.2 Record structures

Logical records used by related specs are `Unit`, `PlayerData`, `MonsterData`, `ItemData`, `Game`, `PathRecord`,
`StatList`, `AiControlBlock`, `DataTables`, `DamageRecord`, and the named table records
(`DifficultyLevelsRecord`, `CharStatsRecord`, `ExperienceRecord`, `MonStatsRecord`, `MonStats2Record`,
`MissilesRecord`, `SkillsRecord`, `ItemStatCostRecord`, `TreasureClassRecord`). Their relevant semantic fields
are the inputs above or the owning spec's fields. Binary fixture mappings and original layout evidence live in
the companion; unestablished roles remain RLI-Q01.

### 3.3 Unit types

The original code's type cases are 0 player, 1 monster, 2 object, 3 missile, 4 item and 5 tile. The stat/state
test accepts types 0, 1 and 3; item seed helpers require 4 (NRF-R12b); the mode setter refuses 5. This classification does
not constrain a new implementation's entity architecture.

### 3.4 Unit modes

Names and original identities are GDF-R05/GDF-R06. This spec directly tests monster mode 1 for the unchanged-mode
exception. Other mode roles belong to the gameplay consumers.

### 3.5 Stats

Stat identities and names are GDF-R02. `Stat(unit,id)` means layer 0; `GetStat(list,id,layer)` reads the list's
own array. The 8.8 format of life, mana and stamina and the diminishing sources belong to
[core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md) (NRF-R08a, NRF-R07).

### 3.6 States

Names and identities are GDF-R03. RLI-V01 cites 54 `uninterruptable` and 146 `invis` as name examples; RLI-V04
uses states 11, 21, 54 and 184 only as identities. The status states 1, 2, 11, 21, 107 and 115
are applied under [combat-status-application](combat-status-application.md). This spec uses the original numeric
identities without requiring a storage layout or claiming custom remapping.

### 3.7 Stat-reader inputs (RLI-R02a, RLI-R02b)

Operation names are illustrative; no original memory layout or function signature is required.

| Input | Width / source | Meaning |
|-------|----------------|---------|
| `unit.statList`, or the supplied `list` | logical list or null | Current list to read |
| `id`, `layer` | i32 arguments | Requested stat and layer; each identity component is reduced modulo 65536 after the id bounds check |
| `statDefinitionCount`, definitions | i32 count; loaded ItemStatCost data | Valid id range and definition for each valid id |
| `list.usesResolvedStats` | boolean, engine-defined list property | Selects resolved values for the unit read and enables the minimum adjustment's list condition |
| `list.ownStats`, `list.resolvedStats` | arrays of `(statId, layer, value)`; value i32 | Existing values; identities are unique and ordered first by signed 16-bit stat id, then unsigned 16-bit layer |
| selected array's active count | signed 16-bit | Nonpositive means no entries; positive counts name valid entries in the ordered array |
| `list.owner` | associated unit or null; type code i32 | Owner tested by the minimum adjustment; not inferred from the unit argument |
| `definition.fMin`, `MinAccr`, `ValShift` | boolean, i32, u8; ItemStatCost data | Minimum enabled, comparison/replacement amount, replacement shift |

The arrays and owner association are inputs to the getter, not produced by it. The boundary
requires valid ordered entries and a loaded definition table; it makes no claim about malformed
memory or how attached-list contributions populate `resolvedStats`.
The current read contract does not establish the complete stat-update lifecycle.

| Mapping / value | Origin | Binding |
|-----------------|--------|---------|
| Stat identities and definition count | `data/global/excel/ItemStatCost.txt`, `Stat` and explicit `ID`; GDA-R12/GDF-R02 | Definition lookup uses the supplied original-profile id; low 16-bit stat and layer identity is fixed by code |
| `fMin`, `MinAccr`, `ValShift` | Same table, columns `fMin`, `MinAccr`, `ValShift` | Loaded definition values; engine code fixes their conditions and comparison/shift order |
| Original-profile stats 34, 35, 143, 145, 147, 149 | Same table, named reduction/flat-absorb rows | `fMin` is clear and compiled `ValShift` is 0; stored values pass through unchanged. Their whole-point use before conversion to 8.8 belongs to DMI-R02/DMI-R08 |

Text and compiled cells used here are recorded in the [companion](../research/spec-evidence/core-record-layouts-and-ids.md#stat-table-verification-2026-10-14).

## 4. Rules (execution order)

These independent definitions and operations are listed in stable id order. Ordering within an operation is
explicit.

**RLI-R01** **Shared names.**

- **RLI-R01a** [verified in 1.13c code (static)]. Loaded table fields have the named column semantics supplied
  by GDA-R03 and the per-table rules. A gameplay caller consumes named fields; original packing is evidence
  rather than a required new internal layout.
- **RLI-R01b** [verified in 1.13c data]. Names/identities follow GDA-R12 and GDF-R02 to GDF-R09's verified
  original-profile records: ItemStatCost Stat/ID correspondence, States state/id, mode row/code, ItemTypes
  compiler identity and individually checked MonStats class correspondences. It does not assert universal
  MonStats text-to-compiled identity. This rule does not infer which editable text cell controls linker
  assignment or that custom archives can freely renumber code-fixed bindings.

**RLI-R02** **Existing stat reads.**

- **RLI-R02a** [verified in 1.13c code (static)]. **Read an existing stat value.**
  `Stat(unit, id, layer)` returns 0 for a null unit or null unit stat list. Both this read and
  `GetStat(list, id, layer)` return 0 for a null list, signed `id < 0`,
  `id >= statDefinitionCount`, or an unavailable stat-definition record. For valid inputs,
  the effective identity is `(id modulo 65536, layer modulo 65536)`, where "modulo 65536" keeps the low 16 bits
  as an unsigned value (layer -65535 becomes 1).
  The unit read selects `resolvedStats` when `list.usesResolvedStats` is true, otherwise
  `ownStats`. The list read always selects `ownStats`, even when `usesResolvedStats` is true.
  Look up that identity in the selected array; a missing entry returns 0 immediately,
  without RLI-R02b's minimum adjustment. A present entry supplies its signed i32 stored
  value to RLI-R02b. No contributions are added during either read, and there is no fallback
  to another array or attached list. The arrays are supplied in the ordering defined in
  section 3.7; constructing and updating them is outside this getter boundary.
  Reads change no stat, list, state, or seed and make **no draw**.
- **RLI-R02b** [verified in 1.13c code (static)]. **Conditional minimum adjustment.**
  For a present entry, let `v` be its stored i32 value. Return `v` unchanged unless all of
  these conditions hold: a definition is supplied, its `fMin` is set,
  `list.usesResolvedStats` is true, the list has an associated owner, that owner's type
  is player (0) or monster (1), and signed `v < definition.MinAccr`.
  If all hold, return the low 32 bits of
  `definition.MinAccr << (definition.ValShift modulo 32)`, interpreted as i32.
  The comparison is with **unshifted** `MinAccr`; `ValShift` applies only to the replacement.
  There is no general shift of the stored value and no `maxstat` lookup or maximum clamp.
  This adjustment is shared by the unit and list reads; selecting `ownStats` in the list
  read does not suppress it when the other conditions hold.

**RLI-R03** **State-list lookup, state test and state set** [verified in 1.13c code (static)].
FindStateList(unit,s) requires a non-null unit (null faults);
a null list, missing state storage or s=0 returns null. There is no range test: any other s, including s<0 or
s>=stateCount, returns the first list whose state equals s on the primary chain, then the secondary chain, or
null. HasState returns 0 for null unit, a type other than 0, 1 or 3, s<0 or s>=stateCount,
null list or missing state storage; otherwise returns `1<<(s mod 32)` if active, or 0. SetState makes no
request when signed s<0 or s>=stateCount, and it applies no unit-type test (HasState's types 0, 1 and 3 filter
does not apply). For a valid s it requests the active bit write carrying the `on` argument unchanged (a nonzero
value, for example 5, is passed as given and sets the bit; zero clears it), which stores only when the unit has
a base list with state storage, then requests MarkForUpdate unconditionally, even when no base list
existed. A null unit can therefore fault in update marking. Actual
changed-state/disguise/update-queue bookkeeping is consumed through STL-R07/STL-R07a and remains RLI-Q02 where
that contract is gated. Find/test make no draw; state-set call routing makes none directly.

**RLI-R04** **Unit mode set.**

- **RLI-R04a** [verified in 1.13c code (static)]. Mode is not range tested. The local mode setter follows:

  ```
  if unit == null or unit.type == 5: return 0
  if requestedMode == unit.mode:
      if not (unit.type == 1 and unit.mode == 1):
          request MarkForUpdate(unit)
          unit.modeChanged = true
      return 0
  request MarkForUpdate(unit)
  unit.modeChanged = true
  unit.mode = requestedMode
  call mode-bound list processing (RLI-R04b)
  request animation reset
  return 1
  ```

  Each request is a call made at its position in this sequence and observes the unit's fields as written up to
  that point: at the changing-mode update request the unit still has its old mode and old modeChanged value.
  Every equal-mode case returns before list processing and animation reset. A monster already in mode 1
  skips both update marking and the modeChanged write; every other equal-mode case performs those two actions.
  Complete update side effects remain RLI-Q02 and animation reset effects NRF-Q07; list dependency
  integration is RLI-Q03.
  The setter itself makes no draw; object reset's local draw is NRF-R13 and server monster-mode callers can add
  NRF-R04b draws.
- **RLI-R04b** [verified in 1.13c code (static)]. Require non-null unit. If no base list or expiryFlag clear,
  return. Walk primary chain. Save successor before testing removeOnModeChange. An unflagged list continues to
  saved successor. For a flagged list: if 0<state<stateCount, clear that state's active bit on the unit (the bit write behind
  RLI-R03's SetState) and request update marking for the unit; then, independently of that state test, if
  usesResolvedStats=false request FreeList(list) through STL-R05. After handling any flagged list restart from
  the CURRENT primary head, not the saved successor. Finally clear base.expiryFlag; every request made during the walk observes it still set. A retained flagged
  extended
  list can cause repeated head traversal; finite supplied fixtures ensure handled lists are removed or no longer
  selected. Free/changed-state side effects remain their dependency integration gates; this rule establishes the
  exact requests/traversal.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| RLI-R01a | Semantic table fields through GDF | verified in 1.13c code (static) | none for named descriptor correspondence | RLI-V01 |
| RLI-R01b | Original-profile data name correspondence | verified in 1.13c data | none for recorded cells; no custom-binding claim | GDF table records; RLI-V01 examples |
| RLI-R02a | Stat guards/identity/selected array | verified in 1.13c code (static) | none at supplied-existing-array boundary | RLI-V02, RLI-V03 |
| RLI-R02b | Conditional minimum adjustment | verified in 1.13c code (static) | none at supplied-definition boundary | RLI-V03 |
| RLI-R03 | State lookup/test and set routing | verified in 1.13c code (static) | RLI-Q02: gated disguise/update bookkeeping only | RLI-V04, RLI-V05 |
| RLI-R04a | Mode refusal/change/request ordering | verified in 1.13c code (static) | RLI-Q02/RLI-Q03 and NRF-Q07: update/animation/list integration | RLI-V06 |
| RLI-R04b | Mode-bound traversal with head reload | verified in 1.13c code (static) | RLI-Q02/RLI-Q03: bookkeeping/free integration | RLI-V06, RLI-V07 |

## 5. Outputs and side effects

| Output / effect | Owning rule |
|---|---|
| Returned stat value only; arrays/lists/states/seeds unchanged | RLI-R02a/RLI-R02b |
| First matching list or state mask; ordered state-bit/update requests | RLI-R03 |
| Mode return 0/1, mode/modeChanged write; update/list/reset requests | RLI-R04a |
| Mode-bound list state/update/free requests, head restart, expiryFlag clear | RLI-R04b |

Requested dependency effects become actual stored/queued effects only under that dependency's acceptance. No read
sends messages, creates units or persists saves.

## 6. Branches and edge cases

- Stat getters reduce identity components only after signed id validation; a missing entry bypasses conditional
  minimum replacement. Existing-array getters do not combine attached lists.
- State lookup/test and mode refusal are explicitly per rule. No mode range check is imposed by the setter; the
  prepared object draw's upper-bound assertion is NRF-R13's.
- Unknown condition names convey only the tested branch, not inferred gameplay semantics. Custom remapping and
  complete new-format support are outside original-profile bindings.
- Mode-bound list traversal restarts at the primary head after a flagged list, and can repeat indefinitely if
  that list remains flagged and attached. Finite vectors supply a removal operation and ordinary nonextended
  mode-bound lists.
- Producer routes and unresolved dependency effects do not acquire verification from supplied-return fixtures.

## 7. Open questions

| Question | Kind / status | Evidence that would settle it |
|---|---|---|
| **RLI-Q01** | outside-boundary / open | Establish wider roles of neutral runtime conditions and record fields (such as `unit.modeChanged`, `itemData.sentinel` and the section 3.2 structures) in their owning specs; no binary field name is required to use the known branch. Status-path conditions are CSA-Q03 |
| **RLI-Q02** | integration / open | Changed-state, disguise and update-queue side effects (RLI-R03, RLI-R04a, RLI-R04b). Supplied inputs and ordered requests remain eligible. Animation preparation and the effects of the animation reset requested by RLI-R04a are NRF-Q07; status-path refresh and timer-node effects are CSA-Q02 |
| **RLI-Q03** | integration / open | Upgrade actual list-free, attach and replacing stat-write integration contracts where STL-R05, STL-R06b/STL-R06c or producer/indirect effects remain gated; consumers that request replacing stat writes cite this gate. Supplied requests/arithmetic do not settle them |
| **RLI-Q04** | observation / pending | Matching gameplay/helper captures for the verified behavioral rules; no observation vector is defined and no gameplay observation has been made |

Attempts and exact blocked assertion boundaries are in the [gap file](../plans/gaps/core-record-layouts-and-ids.md). Getters, state helpers and the mode setter have no independently visible effect; caller captures can observe their inputs/results. Archive cell reads and recorded layout reads are not completed executable gameplay observations.

## 8. Test vectors

### 8.1 Executable vectors

Defaults for every vector unless its Inputs state otherwise:

- Fixtures: stat definition count 359 with available named entries; `stateCount` 185; valid non-null
  associations; no unmentioned lists, states or conditions. Every case starts fresh from its own Inputs and the
  vector's Inputs. The illustrative state-write argument `on` is 1 unless a case gives another value.
- Local execution: execute each cited verified RLI operation and its in-boundary RLI callees. In particular, the
  mode setter executes RLI-R04b's traversal. List visits are observed local actions, not supplied replacement
  operations.
- Operations: supplied dependency operations (state-bit set/clear, update mark, list free and animation reset) log their semantic arguments in call order and apply exactly the
  stored effect the vector states; none draws. A log entry names the operation and its arguments; where a vector
  states an observed field (`mode`, `modeChanged`, `expiryFlag`), the entry records that field's value at call
  time. Only the observed fields a vector states are asserted; a field it does not state at a request is not
  asserted there. Logs (request order, visit order) are asserted outputs. This makes local arithmetic and ordering executable without
  accepting dependency integration. RLI-V02, RLI-V03, RLI-V04 and RLI-V05 use the getters of
  RLI-R02a/RLI-R02b themselves, with no dependency mock.
- Seeds: every vector is `no draw`; no seed is read or stepped.
- Faults: an Outputs entry "fault" means the operation does not return normally and no defined state follows;
  a test checks only that the operation does not complete with a result.
- Assertions: numeric assertions belong to the rule ids a vector names; actual dependency effects remain the Id
  map's named gates.
- Every vector has `runtime_observed: no`.

- **RLI-V01** Named data correspondence (RLI-R01a, RLI-R01b; boundary: isolated rule)
  - Inputs: original-profile tables through GDF's linked table and loader records
  - Seeds: `no draw`
  - Draws: none
  - Trace: look up each identity in its named table.
  - Outputs: stat 99 is `item_fastergethitrate`; state 0x36 is `uninterruptable`; player mode 0x13 is
    `Knock back`; monster mode 13 is `knockback` with code `KB` and token `GH`; state 0x92 is `invis`.
  - Discriminates: each assertion fixes one identity-to-name pair; reading the hexadecimal identities 0x36, 0x13
    and 0x92 as decimal selects different rows. This vector makes no gameplay-execution claim.
  - runtime_observed: no

- **RLI-V02** Stat guards, identity and array selection (RLI-R02a; boundary: isolated rule)
  - Inputs: definition count 359; every available definition has `fMin` false, `MinAccr` 0, `ValShift` 0. A
    non-null unit has a list with `usesResolvedStats` true, associated owner type 0,
    `ownStats = [(34,0,11), (34,1,13), (36,0,25)]`, `resolvedStats = [(34,0,37), (34,1,41), (36,0,50)]`; each
    array has active count 3. Tuples are `(statId, layer, value)`.
  - Seeds: `no draw`
  - Draws: none
  - Trace: RLI-R02a guards, identity reduction and array selection.
  - Outputs: per case; no input changes
  - (a) `Stat(unit, 34, 0)` = 37; `GetStat(list, 34, 0)` = 11.
  - (b) `Stat(unit, 34, 65537)` = 41; `Stat(unit, 34, -65535)` = 41; `Stat(unit, 34, 2)` = 0.
  - (c) Inputs: `usesResolvedStats` false. `Stat(unit, 34, 0)` = 11.
  - (d) Each of: unit null; unit list null; list argument null; id -1; id 359; id 65570; definition 34
    unavailable. Outputs: 0 in the applicable read.
  - (e) Inputs: resolved active count 0, or -1. `Stat(unit, 34, 0)` = 0 without inspecting entries.
  - (f) Inputs: `resolvedStats = [(34,1,41), (36,0,50)]`, count 2. `Stat(unit, 34, 0)` = 0.
  - (g) Inputs: definition count 359 with a loaded record present at index 359; `usesResolvedStats` false;
    `ownStats = [(359,0,5)]`; `Stat(unit, 359, 0)`. Outputs: 0.
  - Discriminates: (a) rejects summing or cross-reading the arrays; (b)'s layer 2 and (f) reject fallback to
    another layer or array; (d)'s id 65570 rejects reduction before the bounds check (its low 16 bits are 34); (g) rejects an
    inclusive upper bound (it would return 5).
  - runtime_observed: no

- **RLI-V03** Minimum condition, comparison and replacement (RLI-R02a, RLI-R02b; boundary: isolated rule)
  - Inputs: definition count 359; stat 7 available with `fMin` true, `MinAccr` 1, `ValShift` 8; unit type 4;
    its list has `usesResolvedStats` true, associated owner type 0, `ownStats = [(7,0,-1)]`,
    `resolvedStats = [(7,0,v)]`, active counts 1.
  - Seeds: `no draw`
  - Draws: none
  - Trace: RLI-R02b's six conditions, comparison with unshifted `MinAccr`, then `MinAccr << (ValShift mod 32)`.
  - Outputs: per case; no input changes
  - (a) `Stat(unit, 7, 0)` with `v` -1, 0, 1, 255, 256 respectively. Outputs: 256, 256, 1, 255, 256.
  - (b) Inputs: `v` -1 and, separately, `fMin` cleared, `usesResolvedStats` false, no associated owner, owner
    type 2, owner type 3, or owner type 4. Outputs: -1 for each. Inputs: `v` -1, owner type 1. Outputs: 256.
  - (c) Inputs: `v` 1024; `GetStat(list, 7, 0)`. Outputs: 256 (own value -1 replaced).
  - (d) Inputs: `resolvedStats` empty, active count 0. Outputs: 0.
  - (e) Inputs: `v` 2, `MinAccr` 3, `ValShift` 36. Outputs: `2 < 3`; shift `36 mod 32 = 4`; 48.
  - (f) Inputs: `v` -1, `MinAccr` 1073741824, `ValShift` 2. Outputs: signed comparison holds; shifted low 32 bits
    0; 0.
  - Discriminates: (a)'s `v` 1 rejects a conventional clamp to 256 and the unit type 4 rejects using the read
    unit's type instead of the owner's; (c) shows the list read applies the minimum; (d) rejects replacing a
    missing entry.
  - runtime_observed: no

- **RLI-V04** Null list and state test (RLI-R03, RLI-R02a; boundary: isolated rule)
  - Inputs: per case
  - Seeds: `no draw`
  - Draws: none
  - Trace: RLI-R03/RLI-R02a guards, then the state bit `1 << (s mod 32)`.
  - Outputs: per case
  - (a) Inputs: non-null unit with a null stat list; `Stat(unit, 6, 0)`. Outputs: 0.
  - (b) Inputs: type 2 unit whose list has state storage and state 0x15 active; HasState for s 0x15, 0 and 184.
    Outputs: 0 for each.
  - (c) Inputs: type 1 unit with `hasStateStorage` true, state 0x15 active, state 0x0B inactive. Outputs:
    `HasState(0x15)` = 0x00200000; `HasState(0x0B)` = 0; `HasState(185)` = 0; `HasState(-1)` = 0.
  - (d) Inputs: type 1 unit with state storage and state 0x15 active; `FindStateList(unit, 0)`. Outputs: null.
  - (e) Inputs: type 0 unit with state storage and state 54 active; `HasState(54)`. Outputs: 0x00400000
    (`1 << 22`).
  - (f) Inputs: type 3 unit with state storage and state 184 active; `HasState(184)`. Outputs: 0x01000000
    (`1 << 24`).
  - (g) Inputs: type 1 unit with state storage whose bit set also has bit 185 set; `HasState(185)`. Outputs: 0.
  - (h) Inputs: type 1 unit with state storage and state 31 active; `HasState(-1)`. Outputs: 0.
  - (i) Inputs: separately, a type 4 and a type 5 unit, each with state storage and state 0x15 active;
    `HasState(0x15)`. Outputs: 0 for each.
  - (j) Inputs: type 1 unit with state storage and state 0 active; `HasState(0)`. Outputs: 1 (`1 << 0`).
  - Discriminates: (c)'s `HasState(185)` rejects an inclusive upper bound and `HasState(-1)` a missing lower
    bound; (b) rejects testing every unit type; (e) and (f) reject accepting only type 1 and a bit index taken
    without `mod 32`; (g) rejects an absent upper bound (it would return `1 << 25`) and (h) an absent lower bound
    (it would return 0x80000000); (i) rejects accepting types above 3; (j) rejects `s <= 0` as the lower bound.
  - runtime_observed: no

- **RLI-V05** State lookup and set routing (RLI-R03; boundary: supplied-input stage)
  - Inputs: per case; non-null unit of type 1; supplied state-bit and update operations log their requests and
    take no draw
  - Seeds: `no draw`
  - Draws: none
  - Trace: RLI-R03 lookup order and set routing.
  - Outputs: per case
  - (a) Inputs: base list with state storage; primary chain `[(state 2, P2), (state 11, P11)]`, secondary chain
    `[(state 11, S11)]`; `FindStateList(unit, 11)`. Outputs: P11.
  - (b) Inputs: base list with state storage; primary chain `[(state 2, P2), (state 11, P11)]`;
    `FindStateList(unit, 0)`. Outputs: null.
  - (c) Inputs: unit without a base list; `SetState(unit, 11, on)`. Outputs: log: bit-set request for state 11
    (no stored effect), update request.
  - (d) Inputs: `SetState(unit, 185, on)`. Outputs: no requests.
  - (e) Inputs: base list with state storage, state 11 off; `SetState(unit, 11, on)` with a supplied bit-only
    operation. Outputs: log: bit-set request, update request; state 11 on.
  - (f) Inputs: base list with state storage, state 11 on; `SetState(unit, 11, on)`. Outputs: log: bit-set
    request, update request; state 11 still on.
  - (g) Inputs: base list with state storage; primary chain `[(state 2, P2)]`, secondary chain
    `[(state 11, S11)]`; `FindStateList(unit, 11)`. Outputs: S11.
  - (h) Inputs: base list with state storage; primary chain `[(state 200, P200)]`, secondary chain empty;
    `FindStateList(unit, 200)`. Outputs: P200.
  - (i) Inputs: base list with state storage; primary chain empty, secondary chain `[(state -1, Sm)]`;
    `FindStateList(unit, -1)`. Outputs: Sm.
  - (j) Inputs: base list with state storage, state 11 on; `SetState(unit, 11, 0)`. Outputs: log: bit write
    carrying 0, update request; state 11 off.
  - (k) Inputs: base list with state storage; `SetState(unit, -1, 1)`. Outputs: no requests.
  - (l) Inputs: base list with state storage, state 11 off; `SetState(unit, 11, 5)`. Outputs: log: bit write
    carrying 5, update request; state 11 on.
  - Discriminates: (a) rejects testing the secondary chain first and (g) skipping it; (c) and (f) reject skipping the update request
    when nothing changes; (d) rejects an inclusive upper bound; (h) and (i) reject a range test in FindStateList (either would return
    null); (j) rejects ignoring a zero argument; (k) rejects a missing lower bound; (l) rejects normalizing the
    argument to 1.
  - runtime_observed: no

- **RLI-V06** Mode setter (RLI-R04a, RLI-R04b; boundary: supplied-input stage)
  - Inputs: per case; `modeChanged` false; base list `expiryFlag` false with empty chains unless stated
  - Seeds: `no draw`
  - Draws: none
  - Trace: RLI-R04a's pseudocode; mode-bound processing by RLI-R04b.
  - Outputs: per case, as ordered logs
  - (a) Inputs: monster, mode 1, requested 1. Outputs: return 0; mode 1; `modeChanged` false; no requests.
  - (b) Inputs: player, mode 1, requested 1. Outputs: update request; `modeChanged` true; return 0; mode 1; no
    list processing or reset request.
  - (c) Inputs: player, mode 1, requested 2. Outputs: update request, observing mode 1 and `modeChanged` false;
    `modeChanged` = true (mode still 1); mode = 2; mode-bound processing (no list effect); reset request, observing
    mode 2; return 1.
  - (d) Inputs: type 5 unit, mode 1, requested 2. Outputs: return 0; no write or request.
  - (e) Inputs: null unit, requested 2. Outputs: return 0; no request.
  - (f) Inputs: player, mode 1, requested 2; base `expiryFlag` true; primary chain `[L]` with L
    `removeOnModeChange` true, `usesResolvedStats` false, state 3 active; supplied FreeList removes L from the
    chain. Outputs: update request (observing mode 1, `modeChanged` false); `modeChanged` = true; mode = 2;
    state 3 bit clear; update request; `FreeList(L)` (each of these three observes `expiryFlag` true and mode 2); restart at
    the empty primary head; `expiryFlag` = false; reset request; return 1.
  - (g) Inputs: player, mode 1, requested -1. Outputs: update request; `modeChanged` = true; mode = -1;
    mode-bound processing (no list effect); reset request; return 1.
  - (h) Inputs: type 5 unit, mode 1, requested 1. Outputs: return 0; no write or request.
  - (i) Inputs: monster, mode 2, requested 2. Outputs: update request; `modeChanged` true; return 0; mode 2; no
    list processing or reset request.
  - Discriminates: (a) against (b) fixes the monster-mode-1 exception; (c)'s order rejects storing the mode
    before marking `modeChanged`; (f) rejects clearing `expiryFlag` before the walk; (g) rejects a mode range
    guard; (h) rejects testing the equal mode before the type 5 refusal; (i) limits the skip to monster mode 1.
    (g) does not establish that the real animation dependency accepts mode -1 (NRF-Q07).
  - runtime_observed: no

- **RLI-V07** Mode-bound head restart (RLI-R04b; boundary: supplied-input stage)
  - Inputs: per case; non-null unit of type 1 and its base list; `stateCount` 185; secondary chain empty
  - Seeds: `no draw`
  - Draws: none
  - Trace: RLI-R04b's walk.
  - Outputs: per case
  - (a) Inputs: `expiryFlag` true; primary chain N, M, K; N and K have `removeOnModeChange` false, states 3 and
    5; M has `removeOnModeChange` true, `usesResolvedStats` false, state 11 active; supplied FreeList removes M;
    state/update operations log and apply the stated state change. Outputs: visit log [N, M, N, K]; requests
    state 11 bit clear, update, `FreeList(M)`, each observing `expiryFlag` true; final chain N, K; state 11 off;
    `expiryFlag` false.
  - (b) Inputs: `expiryFlag` false; primary chain N, M, K with N and K unflagged and M flagged. Outputs: no visits,
    requests or changes.
  - (c) Inputs: `expiryFlag` true; primary chain `[Z]`; Z has `removeOnModeChange` true, `usesResolvedStats` false,
    state 0; supplied FreeList removes Z. Outputs: visit log [Z]; requests `FreeList(Z)` only, with no bit clear or
    update request; final chain empty; `expiryFlag` false.
  - (d) Inputs: `expiryFlag` true; primary chain `[Y]`; Y has `removeOnModeChange` true, `usesResolvedStats` false,
    state 200; supplied FreeList removes Y. Outputs: visit log [Y]; requests `FreeList(Y)` only; final chain empty;
    `expiryFlag` false.
  - Discriminates: (a)'s repeated N rejects saved-successor traversal after a flagged list, and the observed
    `expiryFlag` rejects clearing it before the walk; (b) rejects walking
    without the flag; (c) and (d) reject nesting the free under the state-range test. State/update integration is RLI-Q02; real lifecycle effects are consumed through STL-R05.
  - runtime_observed: no

### 8.2 Observation vectors

No observation vector is defined. The getters, state helpers and mode setter have no independently visible
effect; their results are observable only through caller captures, which remain pending under RLI-Q04. Read-only
layout and table reads recorded in the companion are not gameplay observations.

## 9. Acceptance and dependencies

Partial acceptance comprises each verified assertion in RLI-V01 to RLI-V07 within its stated supplied-input
boundary, excluding the Id map's named integration effects. Original-profile named cells require the linked table
records, not gameplay execution. Full reproduction additionally requires resolving every in-scope gate, accepting
actual dependency-backed stored/queued effects, and matching observations for the behavioral rules. Passing
mocked requests never proves full state, mode or animation parity. All vectors remain runtime_observed:no.

**Depends on:** [core-game-data-fields](core-game-data-fields.md), GDA-R03a and the cited per-table/name rules for fields and original-profile identities; [state-timed-state-lifecycle](state-timed-state-lifecycle.md), STL-R05 list free in mode-bound traversal and STL-R07/STL-R07a state bookkeeping. STL-R07 verifies active/changed membership; STL-R07a disguise and RLI-Q02 update-queue effects remain gated. [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md) supplies notation only.

**Depended on by:** gameplay specs consuming stat reads, state routing, mode changes, unit types or the damage record, including [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md) (NRF-R07 reads through RLI-R02a/RLI-R02b), [combat-status-application](combat-status-application.md) (state routing, stat reads, damage record), [damage mitigation](combat-damage-mitigation.md) (stat reads), [damage pipeline](combat-damage-pipeline-order.md) (damage record), timed states and the related combat specs below. A reverse consumer relationship is not an automatic delivery prerequisite.

## 10. Related specs

- [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md): notation, draw forms and the diminishing read through RLI-R02a; the object animation draw NRF-R13 runs inside the animation reset requested by RLI-R04a.
- [core-game-data-fields](core-game-data-fields.md): the names of ids cited here (GDF-R01 to GDF-R09), with table fields in [core-game-data-table-records](core-game-data-table-records.md) (GDA-R01 to GDA-R16); this spec names the structures (RLI-R01).
- [state-timed-state-lifecycle](state-timed-state-lifecycle.md): state bookkeeping behind RLI-R03 and the list free requested by RLI-R04b.
- [combat-status-application](combat-status-application.md): the status appliers consume the damage record's duration fields, the state routing of RLI-R03 and the stat reads of RLI-R02a.
- [combat-damage-pipeline-order](combat-damage-pipeline-order.md): the damage record (section 3.1).
- [combat-damage-mitigation](combat-damage-mitigation.md): the stat reads of RLI-R02a and the minimum metadata of section 3.7.

## Appendix A. Provenance

Evidence target: Lord of Destruction Expansion 1.13c, Windows x86. The original
`D2Common.dll` and `Fog.dll` specimens are identified by exact SHA-256 fingerprints in the
[source manifest](../research/source-manifest.md). The manifest also records the loader
binary and the session executable, with their correspondence limits.

The [public evidence companion](../research/spec-evidence/core-record-layouts-and-ids.md)
provides module-relative function locations, export ordinals where present, rule-linked
rationale and named table cells. Static readings were recorded on 2026-10-12/13, 2026-10-14
(including the targeted stat-getter reading) and 2026-10-15 using original instruction
listings and reconstructions checked against those instructions where recorded. Function
descriptions are analyst labels.

Recorded table reads establish only their stated scope; effective data-file hashes and
supplying MPQs were not captured. Gameplay observations remain pending, and the spec
remains draft. Public verification requires matching original game files, not access to
the private reconstruction project.
