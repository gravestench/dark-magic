# Monster melee specials

**Description**

Monster close-range skills can combine attacks with extra effects, jump toward a target, change their movement, or consume a corpse. Swarm movement chooses between two skill calculations, and corpse eating has a separate health-adjustment calculation. The complete attacks, movement and corpse-removal effects remain gated where their shared operations are unresolved.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md), [core-record-layouts-and-ids](core-record-layouts-and-ids.md); names: [core-game-data-table-records](core-game-data-table-records.md), [core-game-data-fields](core-game-data-fields.md)
Ids: MML-R, MML-Q, MML-V, MML-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/skill-monster-melee-specials.md)
Open evidence gaps: [gap file](../plans/gaps/skill-monster-melee-specials.md)

## 1. Scope

This contract covers Fire Hit, Quick Strike, monster Frenzy, Jump, Swarm Move, corpse eating, Overseer Whip and Siege Beast Stomp action/start callbacks. The verified subsets are Swarm Move's calculation/condition stage and corpse eating's supplied-input health arithmetic. AI choice, actual attack resolution, path planning, corpse lifecycle, timed states and area target enumeration retain the explicit dependency gates below. Blood-sucking behavior is owned by [monster resurrect and heal](skill-monster-resurrect-and-heal.md); it is not duplicated here.

## 2. Trigger and ownership

D2Game owns the skill callbacks. Skills `srvstfunc` selects start behavior and `srvdofunc` selects the reached action through SCD-R05a/SCD-R06. Starts and actions run synchronously with their dispatcher calls; action frequency and animation/AI production remain the dispatch contract's own gates. Swarm Move start prepares a path and current-skill conditions; its action consumes the current skill. Shared attack, movement, state and corpse operations own their client-visible consequences; this contract establishes no unconditional client message.

## 3. Inputs and data

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| game, caster, identity, invocation level | skill callback | references; i32 | Selected start/action context |
| prepared skill record and count | loaded Skills, GDA-R11a | record or absent; signed i32 count | Signed identity validation |
| current skill and conditions | caster current-skill operation and condition set | node or absent; u32 condition set | Swarm alternate-calculation condition and other conditions |
| selected calculations | `Skills.calc1`, `Skills.calc2` | compiled references; i32 evaluated result | Swarm argument or corpse health percentage |
| frame-event argument | caster's logical sequence-event argument; SCD-R09a | unsigned 24-bit | Swarm action replaces it with low 24 bits of chosen result |
| corpse maximum health | supplied stat 7 `maxhp` read at the corpse-capacity query | i32 raw 8.8 health | Positive quantity enables the eating health stage |
| current and maximum caster health | supplied stat 6 `hitpoints`, stat 7 `maxhp` at reached reads | i32 raw 8.8 health | Health adjustment and upper limit |
| target, corpse state, path, action phase | target refresh, predicate, current skill and movement operations | references/booleans/logical fields | Gated full callbacks |
| damage/element/state/missile fields | prepared Skills records | GDA schema, values not assumed | Gated attack/whip/stomp integration |

| Mapping / constant | Origin | Binding |
|---|---|---|
| Named selectors in MML-R01 | effective `data/global/excel/skills.txt`, named row/column; T01 | Data selects fixed engine callbacks; compiled correspondence MML-Q01 |
| alternate-calculation condition, 24-bit replacement, clamp 1 and divisor 100 | original engine instructions | Fixed by code; the alternate condition's upstream production remains MML-Q06 |
| health stat names/identities | named shared data contract GDF-R02 | Data identity; caller stat selection is fixed by code |

The raw lower bound 1 is one health fraction in 8.8 units. It is not 256 raw units.

## 4. Rules (execution order)

**MML-R01** Named handler cells [verified in 1.13c data]. T01 establishes Fire Hit start 42/action 83; Quick Strike start 50/action 92; MonFrenzy start 64/action 109; Jump start 47/action 89; Swarm Move start 48/action 90; RegurgitatorEat action 108; Overseer Whip action 131; Siege Beast Stomp action 134. These are individual text-cell associations, not proof of complete AI execution or compiled correspondence.

**MML-R02** Swarm calculation stage [verified in 1.13c code (static)]. For action selector 90, return 0 for signed identity negative or at least skill count, absent prepared record, or absent current skill; do not evaluate a calculation or change the frame-event argument on these paths. With a present current skill, inspect its alternate-calculation condition. If set, clear the entire current-skill condition set and select `calc2`; otherwise retain the conditions and select `calc1`. Evaluate exactly the chosen reference once with `(caster, invocation identity, invocation level)`. Replace the caster frame-event argument with `u32(result) mod 16777216`, preserving the separate low-byte sequence field, then return 1. The logical fields are separate; no storage layout is required. The stage makes no direct random draw; the evaluator accounts for its own draws. Upstream condition production/path preparation is MML-Q06 and does not gate this supplied-condition/result contract.

**MML-R03** Corpse health arithmetic [verified in 1.13c code (static)]. At action 108's reached health stage, take the supplied corpse maximum health `B`, caster current health `H` and maximum health `M`. If signed B is at most 0, request no calculation and no health write, then return 1. Otherwise evaluate skill `calc1` with caster, invocation identity and level to obtain signed percentage `P`. Compute `D=mulDiv(P,B,100)` using NRF-R05 in exactly that operand order. Compute `T=i32(H+D)`, replace it with 1 if signed T is at most 1, then set `W=min(T,M)` by signed comparison. Request a layer-0 stat 6 `hitpoints` write of raw value W and return 1. There is no local draw beyond the calculation's own accounting. Actual corpse eligibility/removal and the stat-read/write integration are MML-Q07; this stage accepts supplied values and a logged write operation without claiming those dependencies executed.

**MML-R04** Fire Hit [unresolved]. Action 83 refreshes target, refuses absent/self target, requests two target/combat operations, then requests start dispatch with the invocation identity/level and force control 0, returning its result unchanged. Start 42 and complete target/combat/ordinary-melee effects are MML-Q02.

**MML-R05** Quick Strike [unresolved]. Action 92 validates skill record, refreshes target and rejects absent/self target, then requests target/combat preparation. It chooses `srvmissilea`; a negative choice on a monster can be replaced by a monster missile operation. Invalid missile choice returns 0; otherwise it requests a missile launch and returns 1 without testing that builder's result. Start 50 requests a zero-initialized damage operation after target validation. Complete attack/damage/monster fallback and launch semantics remain MML-Q03.

**MML-R06** Monster Frenzy [unresolved]. Action 109 requires a current skill and distinct target. After target/combat preparation it sets the skill-action condition. A local animation-event predicate determines whether start dispatch is requested, but subsequent effect application is requested either way. An additional current-skill value can select another follow-up operation. The callback returns 1 after those reached stages; start 64 refreshes and requires a distinct target. Exact event meaning, damage and effect/state writes are MML-Q04.

**MML-R07** Jump [unresolved]. Start 47 and action 89 use current-skill/path phase fields to prepare and advance a jump. The reached action removes selected unit events before examining its phase. Distinct phase paths can change path configuration or request contact combat. A zero or absent target can select stored path coordinates; a distinct target produces a projected destination. Complete phase meanings, destination/room tests, collision and damage remain MML-Q05.

**MML-R08** Swarm path preparation [unresolved]. Start 48 refreshes and requires a distinct target, a path and current skill. It requests an initial path configuration and path build, retries with an alternate configuration after failure, and returns 0 if neither build succeeds. A successful start sets current-skill conditions to its initial movement condition and returns 1. Complete movement/path effects, later alternate-condition production and timing are MML-Q06. No RNG or distance rule is inferred from the path labels.

**MML-R09** Corpse consumption [unresolved]. Action 108 validates the skill and a distinct target, requires monster type, a target predicate to be false and corpse mode 12, then requests corpse identity/lifecycle operations before MML-R03. Consequently a later health-stage no-write result does not mean no corpse effect. Full corpse eligibility, removal, target refresh and health-stat integration remain MML-Q07.

**MML-R10** Overseer Whip [unresolved]. Action 131 validates skill and a target-state selector, refreshes a distinct target and checks a target predicate. A particular target family takes a `calc1`-based chance test with a caster-seed draw modulo 100. That chance result and current target state select between an ordinary attack-like operation and a special target effect. Both reached branches return 1. Target-family identity, precise eligibility, state and follow-up effects remain MML-Q08; the draw's presence does not supply the missing transition contract.

**MML-R11** Siege Beast Stomp [unresolved]. Action 134 validates skill and a strictly positive missile selector below missile count, initializes a damage record, applies configured damage fields, requests two damage computations and a radius calculation, then requests area processing with a fixed filter. It returns 1 after the reached area request; record/missile failure returns 0 first. Damage calculations, radius and area callback/filter semantics remain MML-Q09.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| MML-R01 | Named start/action cells | verified in 1.13c data | none for cells; MML-Q01 compiled binding | T01 in companion |
| MML-R02 | Swarm condition/calculation/argument | verified in 1.13c code (static) | SCX dependencies for actual evaluated result; MML-Q06 upstream path/condition only | MML-V01, MML-V02, MML-V05 |
| MML-R03 | Corpse health arithmetic | verified in 1.13c code (static) | MML-Q07 actual corpse/stat integration; supplied stage available | MML-V03, MML-V04 |
| MML-R04 | Fire Hit | unresolved | MML-Q02 full effects | none |
| MML-R05 | Quick Strike | unresolved | MML-Q03 full effects | none |
| MML-R06 | Frenzy | unresolved | MML-Q04 full effects | none |
| MML-R07 | Jump | unresolved | MML-Q05 movement/contact | none |
| MML-R08 | Swarm preparation | unresolved | MML-Q06 path/condition production | none |
| MML-R09 | Corpse lifecycle | unresolved | MML-Q07 eligibility/removal/stats | none |
| MML-R10 | Whip | unresolved | MML-Q08 chance/state/effects | none |
| MML-R11 | Stomp | unresolved | MML-Q09 damage/radius/area | none |

## 5. Outputs and side effects

| Output | Rule | Boundary |
|---|---|---|
| current-skill conditions cleared or retained; frame-event argument replaced; result 0/1 | MML-R02 | Verified supplied stage, evaluator effects separate |
| layer-0 hitpoints write request or no write; result 1 | MML-R03 | Verified arithmetic/request with supplied values |
| target/combat/launch/effect requests and callback results | MML-R04, MML-R05, MML-R06 | Gated complete attacks |
| path/phase/event/movement requests | MML-R07, MML-R08 | Gated complete movement |
| corpse lifecycle requests | MML-R09 | Gated consumption |
| attack/special-target effect requests | MML-R10 | Gated whip |
| damage/radius/area processing requests | MML-R11 | Gated stomp |

No save write, unconditional message or actual health-stat persistence is asserted by the supplied arithmetic fixture.

## 6. Branches and edge cases

Swarm's alternate branch clears every current-skill condition, including unrelated bits; its ordinary branch preserves them. Negative calculation results narrow to unsigned 24 bits for the frame-event argument. Corpse health arithmetic performs the raw lower bound before its maximum-health cap; a maximum below 1 can therefore produce a value below that lower bound. The addition wraps before clamping. Nonpositive corpse maximum health skips calculation and write. Callback return 0 does not imply no earlier dependency side effect. Local verified stages have no difficulty or player-versus-player branch; the dependency inputs and full gated attacks can differ. Null caster is not assigned a safe success behavior.

## 7. Open questions

**MML-Q01** [unresolved] Match the named effective text cells to compiled records and linked numeric values; synthetic supplied-record vectors establish no such correspondence.

**MML-Q02** [unresolved] Close Fire Hit start, target preparation and ordinary melee helper through actual damage/effect writes.

**MML-Q03** [unresolved] Close Quick Strike start/damage, target operations, monster missile fallback and launch builder.

**MML-Q04** [unresolved] Close Frenzy's animation predicate, effect application and additional current-skill follow-up through state/stat writes.

**MML-Q05** [unresolved] Close Jump start/action phases, path/collision/room operations, target projection and contact combat.

**MML-Q06** [unresolved] Close Swarm Move start and upstream alternate-condition production, including movement and path side effects. This does not gate the supplied-condition MML-R02 stage.

**MML-Q07** [unresolved] Close corpse predicate/lifecycle and actual stat reads/writes. The raw health arithmetic's supplied operands, write request and result remain independently available.

**MML-Q08** [unresolved] Resolve whip target family, exact chance branch, state test and both effect paths with discriminating seed vectors.

**MML-Q09** [unresolved] Close stomp damage, radius calculation, area enumeration/filter and callback effects.

**MML-Q10** [unresolved] Matching gameplay observations remain pending. The companion records inconclusive observation limits and a reproduction procedure; no matching gameplay invocation or effect was captured.

## 8. Test vectors

### 8.1 Executable vectors

Fixtures are independent. The supplied evaluator logs selected column and `(caster, identity, level)`, returns the given result without draws and performs no other mutations. The supplied health writer logs `(stat,layer,value)` and does not mutate stored health. Current-skill conditions and frame-event argument are stored semantic fixture state. Guard fixtures additionally observe record access, current-skill lookup, condition writes and argument writes without changing the supplied values. No gated attack, movement or corpse lifecycle is simulated.

- **MML-V01** Swarm alternate clears all conditions before `calc2` (MML-R02; boundary: supplied-input stage)
  - Inputs: monster C; action 90; identity 166 in skill count 400; record/current skill present; level 4; conditions {alternate-calculation, other}; frame-event argument 77; separate low-byte sequence value 9; `calc1` result 11 if called; `calc2` result -1.
  - Seeds: no draw in supplied evaluation or local stage.
  - Draws: none.
  - Trace: alternate present; clear all conditions; evaluate `calc2`(C,166,4) once; u32(-1) modulo 16777216=16777215.
  - Outputs: result 1; conditions empty; evaluator log [`calc2`(C,166,4)]; frame-event argument 16777215; separate low-byte value 9 unchanged.
  - Discriminates: clearing only the alternate condition leaves other present; selecting `calc1` gives 11; sign extension gives -1 instead of 16777215.
  - runtime_observed: no.

- **MML-V02** Swarm ordinary retains conditions (MML-R02; boundary: supplied-input stage)
  - Inputs: monster C; action 90; identity 166; count 400; record/current skill present; level 2; conditions {other}; frame-event argument 77; low-byte value 9; `calc1` result 11; `calc2` result 19 if called.
  - Seeds: no draw in supplied evaluation or local stage.
  - Draws: none.
  - Trace: alternate absent; retain conditions; evaluate `calc1`(C,166,2); retain low 24 bits 11.
  - Outputs: result 1; conditions {other}; evaluator log [`calc1`(C,166,2)]; frame-event argument 11; low-byte value 9 unchanged.
  - Discriminates: unconditional condition clearing changes {other}; unconditional `calc2` changes result to 19.
  - runtime_observed: no.

- **MML-V03** Eating uses ordered mulDiv arguments (MML-R03; boundary: supplied-input stage)
  - Inputs: action 108 reached health stage; monster C; identity 214; level 1; B=70001 raw; H=256 raw; M=100000000 raw; `calc1` result P=100001.
  - Seeds: no draw in supplied evaluation or local stage.
  - Draws: none.
  - Trace: B positive; evaluate `calc1`(C,214,1). NRF-R05 sees B at most 1048576 and P above 65536; 100 at most (P>>4), so D=trunc(100001/100)*70001=1000*70001=70001000. T=70001256; lower and upper clamps leave it.
  - Outputs: result 1; evaluator log [`calc1`(C,214,1)]; writer log [(hitpoints,0,70001256)]; stored health is not asserted.
  - Discriminates: swapping mulDiv operands yields 70000700 before addition, therefore writer value 70000956 instead of 70001256.
  - runtime_observed: no.

- **MML-V04** Eating lower bound then maximum cap (MML-R03; boundary: supplied-input stage)
  - Inputs: action 108 reached stage; monster C; identity 214; level 1; B=1000 raw; H=200 raw; M=0 raw; `calc1` result -100.
  - Seeds: no draw in supplied evaluation or local stage.
  - Draws: none.
  - Trace: D=trunc(-100*1000/100)=-1000; i32(H+D)=-800; lower bound gives 1; min(1,0)=0.
  - Outputs: result 1; evaluator log [`calc1`(C,214,1)]; writer log [(hitpoints,0,0)].
  - Discriminates: capping before applying the lower bound produces 1 and fails the expected write.
  - runtime_observed: no.

- **MML-V05** Swarm refusal leaves condition and event state unchanged (MML-R02; boundary: supplied-input stage)
  - Inputs: present monster C; action 90; level 2; skill count 400; identity 166; valid prepared record; stored skill node N with conditions {other}; current skill selection N; caster frame-event argument 77; separate low-byte sequence value 9. Each case independently overrides these defaults. Complete prepared storage is used except where explicitly malformed.
  - Seeds: no draw.
  - Draws: none.
  - Trace: signed identity bounds precede prepared-record access; a present record precedes current-skill lookup; an absent current-skill selection returns before conditions/evaluator/argument setter.
  - Outputs: every case returns exactly 0; node N conditions remain {other}; frame-event argument remains 77; low-byte value remains 9; evaluator log empty; no condition or argument write; no direct draw.
  - Discriminates: continuing despite a failed guard would evaluate a calculation or replace the frame-event argument; an unconditional cleanup would clear N's conditions.
  - runtime_observed: no.
  - (a) Inputs: identity -1. Outputs: 0; no record access, no current-skill lookup, no evaluation or writes; all stated fixture values unchanged.
  - (b) Inputs: identity 400. Outputs: 0; no record access, no current-skill lookup, no evaluation or writes; all stated fixture values unchanged.
  - (c) Inputs: identity 0; synthetic malformed prepared storage resolves that record as absent despite count 400. Outputs: 0; no current-skill lookup, no evaluation or writes; all stated fixture values unchanged.
  - (d) Inputs: identity 166; record present; current skill selection absent, while stored unselected node N still exists. Outputs: 0; one current-skill lookup; no condition read/write, evaluation or argument write; N's conditions and the caster event fields remain unchanged.

### 8.2 Observation vectors

**MML-O01** Capture Swarm Move start/action conditions, selected calculation and frame-event argument at one eligible movement transition; compare the captured inputs to MML-R02. **MML-O02** Capture a corpse-eating action's target eligibility, corpse maximum health, reached caster health/max and percentage, then its write request; compare MML-R03 separately from gated corpse effects. Tool-neutral procedures are in the companion.

## 9. Acceptance and dependencies

Partial acceptance comprises the T01 cells and MML-V01 to MML-V05. This establishes supplied swarm calculation state and corpse health arithmetic only. Full reproduction requires exact vectors and compatible observations after MML-Q01 to MML-Q10 close; no attack, path or corpse-removal parity follows from the partial pass.

**Depends on:** SCD-R05a/SCD-R06 for start/action selection; SCD-R09a for the frame-event argument convention; GDA-R11a for supplied prepared records; GDF-R02 for health stat names; NRF-R05 for operand-ordered mulDiv; SCX-R11 for selected calculation context, with that contract's compiler/parameter gates retained. Actual combat/path/corpse/stat integration is explicitly unresolved here.

**Depended on by:** no known plan/code/test consumer.

## 10. Related specs

[Skill dispatch](skill-cast-start-and-do-dispatch.md) owns invocation and outer return handling. [Combat damage pipeline](combat-damage-pipeline-order.md) owns resolved hit order; [combat status application](combat-status-application.md) owns timed effect application. [Monster resurrect and heal](skill-monster-resurrect-and-heal.md) owns blood-sucking callbacks separately.

## Appendix A. Provenance

Original LoD Expansion 1.13c Windows x86 D2Game callback instructions and original start/action tables were read on 2026-10-29 against `LOD113C-D2GAME` in the [source manifest](../research/source-manifest.md). D2Common field operations use byte-equivalent C and named shared contracts within their recorded boundaries. The [companion](../research/spec-evidence/skill-monster-melee-specials.md) records exact roots, imported operations, named archive cells, unresolved local callees and actual setup attempts. No gameplay rule is runtime-confirmed.
