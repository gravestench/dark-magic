# Unit sound selection

**Description**

Unit sound selection connects a creature's actions with its sound cues. The exact rules for choosing creature and player cues remain partly unresolved, including special creature overrides and the timing of movement sounds.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md), [core-record-layouts-and-ids](core-record-layouts-and-ids.md)
Ids: USN-R, USN-Q, USN-V, USN-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/sound-unit-sound-selection.md)
Open evidence gaps: [gap file](../plans/gaps/sound-unit-sound-selection.md)

## 1. Scope

Selection of sound records and cue requests for players and monsters, including action and movement timing. Playback admission, asset variation and positional attenuation are separate sound boundaries. This contract gates the unverified selection paths rather than assigning meanings to table columns from their names.

## 2. Trigger and ownership

D2Client owns the candidate selection operations. Calls from monster presentation updates into movement timing are located; the complete frame dispatch and action-event roots remain USN-Q01. The precise cadence and relationship to server frames are unresolved. No server message contract is established here.

## 3. Inputs and data

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| Unit | Client unit state | Logical optional unit | Emitter whose sound is requested |
| Effective unit kind, class and mode | Unit identity conversion | Integer; exact conversion unresolved | Selects player/monster branches |
| Creature sound record | Candidate sound-record lookup | Optional logical record | Source of action, movement and secondary cue selections; field meanings gated |
| Current tick and previous cue tick | Client presentation timing | Unsigned 32-bit values | Timing comparisons; tick duration unresolved |
| Presentation seed | Client-local random state | Two unsigned 32-bit words | Candidate variation and chance draws; ownership confirmation gated |
| Action selector | Supplied action-event selector | Signed 32-bit integer | 8..11 select the four creature skill cue fields |
| Four skill cues | Supplied `MonSounds.Skill1` through `Skill4` | Signed 32-bit identities | Exact cue forwarded by USN-R02b |
| Cue request operation and return | Supplied playback boundary | Logged operation and integer return | Return ignored by USN-R02b; audible admission is separate |
| Clock read sequence | Client presentation clock | Unsigned 32-bit ticks | One admission read and two ordered post-request reads for an admitted action |
| Unit cue tick and shared cue tick | Client cue timing state | Unsigned 32-bit ticks | Unit tick supplies fixed-cue cooldown; both receive ordered post-request writes |

The sound-record loader binds the observed fields to `MonSounds` names as recorded in the companion; the complete `MonStats` joins and override meanings remain unresolved. One effective text-file value is established: `MonSounds.FsCnt (act1hire) = 2`, from `data/global/excel/monsounds.txt`, named row `act1hire`, column `FsCnt` ([table record T01](../research/spec-evidence/sound-unit-sound-selection.md#table-record-t01)). Origin: table data. Binding: resolved from data for the text cell; correspondence to the compiled runtime field is unresolved.

## 4. Rules (execution order)

**USN-R01** Effective sound identity and creature record [unresolved]

The candidate path obtains an effective identity and selects either a default creature sound record or an alternate record. Player identity conversion, alternate-record eligibility and the exact table bindings are gated by USN-Q01 and USN-Q02. This rule establishes no executable record-selection result until those contracts are resolved.

**USN-R02a** Action dispatch integration [unresolved]

The action path consumes the effective-kind and selected creature-record results from USN-R01. Its full client trigger and those producers remain USN-Q01/USN-Q02. Playback admission is a separate boundary: a request does not establish an audible cue.

**USN-R02b** Supplied action selection and timing writes [verified in 1.13c code (static)]

At the supplied-result stage, return without requests or timing writes unless effective kind is monster and the selected record exists. Selector 8 chooses `Skill1`, 9 chooses `Skill2`, 10 chooses `Skill3`, and 11 chooses `Skill4`; every other selector returns without requests or timing writes. Each supplied cue identity is a signed 32-bit value. For cue identity 2692, compute unsigned 32-bit subtraction of current admission tick minus the unit's earlier cue tick, wrapping modulo 2^32; values below 25 suppress the request. For other identities, exactly zero suppresses the request; negative values are forwarded unchanged.

Otherwise issue one request carrying the selected cue, unit emitter and three zero auxiliary selectors. Ignore its return. After the request returns, read the clock and write that value as the unit's cue tick; read the clock again and write that value as the shared cue tick. The reads are ordered and can differ. No random draw occurs locally in this stage; the request operation's internal effects belong to its separate owner. Origin of selector mapping, fixed identity 2692 and threshold 25: engine code. Binding: fixed by original code; these constants do not establish what a custom table associates with cue 2692.

**USN-R03** Movement cue timing and secondary cue [unresolved]

The candidate movement path tests effective kind and mode, compares animation progress against repeating phase positions, requests a primary cue and can request a second cue using a random threshold. The complete phase calculation, modes, data bindings and request admission are gated by USN-Q01 and USN-Q04. Reaching a request does not establish audible playback.

**USN-R04** Shared movement variation and cooldown [unresolved]

The candidate variation path has a shared elapsed-tick gate and consumes a presentation random draw before requesting a cue variant. Its dispatch, random-state ownership, timing units and cue argument meanings are gated by USN-Q01 and USN-Q04.

**USN-R05** Named movement-count data cell [verified in 1.13c data]

The effective original-profile `MonSounds` text table contains decimal `2` in column `FsCnt` of named row `act1hire`. This establishes the named cell only; execution frequency, meaning and compiled-runtime correspondence remain USN-Q04.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| USN-R01 | Identity and record selection | unresolved | USN-Q01, USN-Q02: full selection result | none |
| USN-R02a | Action dispatch integration | unresolved | USN-Q01, USN-Q02, USN-Q03: full entry, identity producers and audible result | none |
| USN-R02b | Supplied action selection and timing | verified in 1.13c code (static) | none locally; USN-R02a gates producer integration | USN-V01, USN-V02, USN-V03 |
| USN-R03 | Movement phase requests | unresolved | USN-Q01, USN-Q04: movement timing and secondary cue | none |
| USN-R04 | Variation cooldown | unresolved | USN-Q01, USN-Q04: variant, random draw and admission | none |
| USN-R05 | Named movement-count cell | verified in 1.13c data | none for named text cell; USN-Q04 gates execution meaning | T01 |

## 5. Outputs and side effects

| Output | Rule | Evidence boundary |
|---|---|---|
| Selected creature sound record | USN-R01 | Gated identity and table mapping |
| Cue request with unit emitter and selector arguments | USN-R02b, USN-R03, USN-R04 | USN-R02b verifies supplied action request; other request paths gated; playback is separate |
| Per-unit and shared cue timing updates | USN-R02b, USN-R03, USN-R04 | USN-R02b verifies ordered post-request writes; other timing paths gated |
| Presentation seed advancement | USN-R03, USN-R04 | Gated complete draw order |

## 6. Branches and edge cases

Missing unit, missing records, alternate creature forms, zero cue, special fixed cue, phase wrap, tick wrap, and secondary-cue chance boundaries remain explicitly within USN-Q01–USN-Q04. The supplied action stage makes no difficulty or PvP distinction; selection producer differences remain gated.

## 7. Open questions

- **USN-Q01** [unresolved] Establish the complete client dispatch roots and identity conversion, including player transformations. Read all local and imported callees and corroborate encoded instruction runs against the matching image.
- **USN-Q02** [unresolved] Bind default and alternate record selection to original data-loader fields and effective archive cells; distinguish unique-monster and other override conditions.
- **USN-Q03** [unresolved] Establish the full action-event entry, original asset association of fixed cue 2692, clock tick duration and playback admission. None gates the supplied-result selection and timing arithmetic in USN-R02b.
- **USN-Q04** [unresolved] Complete movement phase arithmetic, primary/secondary cues, random draw order, shared cooldown and variation semantics.
- **USN-Q05** [unresolved] Confirm established selection and request assertions on an owned original runtime after static gates close. No gameplay observation has been collected.

## 8. Test vectors

### 8.1 Executable vectors

These vectors test USN-R02b with explicitly supplied identity and record results, not their producers. The request operation logs its arguments, returns the supplied value and makes no other changes. Clock results are supplied in read order; local seeds are `no draw`. Initial request log is empty in every case. The unit is non-null emitter A. Initial shared cue tick is 7. USN-R05 instead uses table record T01.

- **USN-V01** Exact selector mapping and rejected request return (USN-R02b; boundary: supplied-input stage)
  - Inputs: effective kind monster; selected record present; `Skill1=101`, `Skill2=102`, `Skill3=103`, `Skill4=104`; selector 9; unit cue tick 90; clock reads 100, 101, 102; request return 0.
  - Seeds: no draw.
  - Draws: none locally.
  - Trace: selector 9 chooses 102; cue is neither zero nor 2692; log request; ignore return 0; write next clock values in order.
  - Outputs: request log [(102, A, 0, 0, 0)]; unit cue tick 101; shared cue tick 102; writes ordered after request.
  - (a) Inputs: selector 8; other V01 inputs unchanged and independently initialized. Trace: choose `Skill1=101`, request once and ignore return 0, then write clock values 101 and 102. Outputs: request log [(101, A, 0, 0, 0)]; unit cue tick 101; shared cue tick 102; writes after request.
  - (b) Inputs: selector 10; other V01 inputs unchanged and independently initialized. Trace: choose `Skill3=103`, request once and ignore return 0, then write clock values 101 and 102. Outputs: request log [(103, A, 0, 0, 0)]; unit cue tick 101; shared cue tick 102; writes after request.
  - (c) Inputs: selector 11; other V01 inputs unchanged and independently initialized. Trace: choose `Skill4=104`, request once and ignore return 0, then write clock values 101 and 102. Outputs: request log [(104, A, 0, 0, 0)]; unit cue tick 101; shared cue tick 102; writes after request.
  - Discriminates: selecting the adjacent field, updating before the request, reusing one clock read, or suppressing timing writes after a failed request changes exact assertions.
  - runtime_observed: no

- **USN-V02** Fixed-cue equality and tick wrap (USN-R02b; boundary: supplied-input stage)
  - Inputs: effective kind monster; record present; all four skill cues 2692; selector 11; unit cue tick 4294967290; clock reads 19, 20, 21; request return 1.
  - Seeds: no draw.
  - Draws: none locally.
  - Trace: unsigned subtraction 19 - 4294967290 wraps to 25; threshold admits equality; request then two timing writes.
  - Outputs: request log [(2692, A, 0, 0, 0)]; unit cue tick 20; shared cue tick 21.
  - Discriminates: unwrapped mathematical subtraction, a strict greater-than threshold, or shared timing used for the cooldown changes the request decision or writes.
  - runtime_observed: no

- **USN-V03** Suppression and negative forwarding (USN-R02b; boundary: supplied-input stage)
  - Inputs: effective kind monster; record present; all four skill cues 2692; selector 8; unit cue tick 90; clock reads 114, 115, 116; request return 0.
  - Seeds: no draw.
  - Draws: none locally in every case.
  - Trace: run each independent case from the stated initial state; suppressed cases preserve timing values.
  - Outputs: base case has elapsed 24, empty request log, unit cue tick 90, shared cue tick 7.
  - (a) Inputs: all four skill cues 0. Outputs: empty log, unit cue tick 90, shared cue tick 7.
  - (b) Inputs: all four skill cues -1. Outputs: request log [(-1, A, 0, 0, 0)], unit cue tick 115, shared cue tick 116.
  - (c) Inputs: selector 12. Outputs: empty log, unit cue tick 90, shared cue tick 7.
  - (d) Inputs: effective kind player. Outputs: empty log, unit cue tick 90, shared cue tick 7.
  - (e) Inputs: selected record absent. Outputs: empty log, unit cue tick 90, shared cue tick 7.
  - Discriminates: unconditional fixed-cue admission, treating every nonpositive identity as zero, or writing timestamps on suppressed requests changes a case's exact outputs.
  - runtime_observed: no

### 8.2 Observation vectors

- **USN-O01** Observe an ordinary monster changing action modes and moving, recording selected record, selector arguments and request result. Expected exact values await USN-Q01–USN-Q04; [procedure](../research/spec-evidence/sound-unit-sound-selection.md#observation-procedure).

## 9. Acceptance and dependencies

Full reproduction requires resolving USN-Q01–USN-Q04, providing exact discriminating executable vectors for USN-R01–USN-R04 and completing USN-O01 for the established assertions. USN-R02b is available through USN-V01–USN-V03 independently of its gated producers and audible playback. USN-R05 provides only a named text-cell assertion. USN-Q05 blocks runtime-confirmation claims independently of static verification.

Depends on: unit identity and numeric/RNG contracts only where specific rule references are established after the open mappings close. Depended on by: no established consumer.

## 10. Related specs

[Numeric and RNG foundations](core-numeric-and-rng-foundations.md) supplies arithmetic terminology. [Record layouts and ids](core-record-layouts-and-ids.md) supplies unit identity context. Neither reference resolves the sound-specific gates.

## Appendix A. Provenance

Evidence build: original Expansion 1.13c Windows x86. The [source manifest](../research/source-manifest.md) identifies the original D2Client and D2Common artifacts. The [companion](../research/spec-evidence/sound-unit-sound-selection.md) records 2026-10-29 disassembly readings, byte-matched imported helper readings, a narrow effective text-table cell read, misleading labels and remaining dispatch/callee limits. No gameplay vector is runtime-confirmed.
