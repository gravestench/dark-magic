# Skill and missile sounds

**Description**

Skills, projectiles and changing states request sounds at different moments. This contract establishes several client selection stages and the conditions for their sound requests. Complete scheduling and audible voice lifetimes remain unresolved.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-game-data-table-records](core-game-data-table-records.md)
Ids: SMS-R, SMS-Q, SMS-V, SMS-O; policy in [README conventions](README.md#conventions)
Evidence: [companion](../research/spec-evidence/sound-skill-and-missile-sounds.md)
Open evidence gaps: [gaps](../plans/gaps/sound-skill-and-missile-sounds.md)

## 1. Scope

Client skill cue selection, missile TravelSound/HitSound/ProgSound requests, and States onsound/offsound requests. Verified boundaries consume supplied identities, records, membership and admission results to establish type-specific/progressive selection, state transition requests, travel/hit requests and progression counter changes. Complete state activation dispatch, start/do/progressive scheduling, missile contact/removal lifetime and actual audible playback remain gated. Gameplay effects and projectile damage/movement are owned by the related missile and skill contracts.

## 2. Trigger and ownership

D2Client processes active state definitions and dispatches the selected client active handler. One such handler selects skill sounds using the unit and a supplied skill identifier. The state-dispatch-to-skill-identifier integration is SMS-Q01. This contract specifies the resulting request log, consuming sound playback as a separately gated operation. It does not equate a successful handler return with audible playback or define a server tick cadence.

## 3. Inputs and data

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| unit | supplied presentation unit | unit reference | emitter supplied to sound requests |
| unit type | unit classification | unsigned 32-bit | 0 selects first cue; 1 selects second; other types select neither |
| presentation selector | supplied unit animation-frame marker | unsigned 8-bit | only marker value 3 enables this handler's record lookup; complete production/cadence SMS-Q01 |
| skill id, record count | supplied skill lookup | signed 32-bit | accepted ids satisfy 0 <= id < count |
| skill record | lookup result | present or absent | absent fails only when lookup is required |
| dosound a, dosound b | supplied cue values associated with Skills record fields | signed 16-bit integers | positive selected value generates a request |
| request sound | supplied operation | logs emitter, cue and three zero extra values; result opaque | request result is ignored in this selection stage |
| progressive flag, aurastate, `aurastat1` | supplied Skills fields | boolean, signed 16-bit state/stat identities | eligibility for progressive charge adjustment |
| state/stat counts and associated list | prepared definitions and supplied state-list lookup | signed counts; optional list | bounds and presence for progressive stat read |
| progressive stat value | supplied associated-list stat read at layer 0 | signed 32-bit | clamped to 1..3 when eligible |
| prgsound, prgoverlay | supplied Skills fields | signed 16-bit identities | bases for progressive sound and overlay requests |
| overlay request | supplied operation | logs identity, emitter, selector 2 and five zero auxiliary values | opaque result ignored by progressive sound stage |
| prior state membership, onsound, offsound | supplied state membership and States fields | boolean; signed 16-bit cue identities | selects state-transition request |
| notondead and unit mode | supplied States flag and unit mode | boolean; unsigned mode identity | skips on-sound stage for player mode 17 or monster mode 12 when flagged |
| TravelSound and created missile | supplied missile creation result and Missiles field | optional missile; signed 16-bit cue | post-creation travel request |
| HitSound | supplied Missiles field | signed 16-bit cue | request after collision/callback admission |
| ProgSound, previous progress counter, current progress counter | supplied Missiles field and missile/path progress state | signed 16-bit cue; 32-bit counters compared for equality | progression sound decision and counter write |

Origin and Binding: the original descriptors associate the named sound fields with prepared record values; client code fixes the selector tests, cue sign policies and progressive clamp. Actual linked row values remain unresolved under SMS-Q01/SMS-Q02/SMS-Q03. No shipped row values or complete name-link resolution are asserted. [Loader binding evidence](../research/spec-evidence/sound-skill-and-missile-sounds.md#bindings) identifies descriptor associations. GDA-R11a supplies record identity and the progressive flag schema; complete local sound-name linking remains gated. The animation marker's numeric value does not itself establish a semantic action name or cadence.

## 4. Rules (execution order)

**SMS-R01** Eligibility and record lookup [verified in 1.13c code (static)]. If presentation selector differs from 3, return 1 without a lookup or sound request. Otherwise, reject a negative skill id, an id at or above record count, or an absent record, returning 0 without a sound request. With a valid record continue to SMS-R02. No draw occurs.

**SMS-R02** Type-specific cue [verified in 1.13c code (static)]. With selector 3 and valid record, type 0 selects signed `dosound a`; type 1 selects signed `dosound b`. Other types select nothing. If the selected cue is strictly positive, request sound with that cue, this unit as emitter and three zero extra values. Otherwise make no request. Ignore the supplied request operation's result. Return 1 in all cases in this rule. No draw occurs.

**SMS-R03a** Progressive request selection [verified in 1.13c code (static)]. At the supplied active-handler boundary, an invalid skill id or absent record returns 0 without requests. Otherwise initialize adjustment to 0. Only if progressive is true, aurastate is nonnegative and below state count, `aurastat1` is nonnegative and below stat count, and the associated state list is present, read that list's `aurastat1` at layer 0 and set adjustment to clamp(value, 1, 3) - 1. The read and clamp use signed 32-bit values. Sign-extend prgoverlay and prgsound to 32 bits and add adjustment to each. Request the resulting overlay first, with emitter, selector 2 and five zero auxiliary values; then request the resulting sound with emitter and three zero values. Neither zero nor negative identities suppress these local requests. Ignore both opaque operation results and return 1. No local random draw occurs; dependency-internal effects are outside this stage.

**SMS-R03b** Other skill phase integration [unresolved]. Start-sound, ordinary do-sound, progressive scheduling and actual overlay/voice effects remain gated by SMS-Q02. SMS-R03a verifies supplied progressive selection and call order, not how often the active handler executes.

**SMS-R04a** Supplied state-transition cue stage [verified in 1.13c code (static)]. For an existing unit and valid supplied state record, the on stage first checks notondead: when set, player mode 17 or monster mode 12 skips sound selection. Otherwise a previously absent state requests onsound exactly when its signed 16-bit value is nonnegative; a previously present state makes no on request. The off stage requests offsound exactly when the state was previously present and its signed 16-bit value is nonnegative. Every request uses the unit emitter and three zero auxiliary values. Results are ignored. Zero is forwarded by these stages; audible admission remains separate. Membership mutation, attached-list removal and overlay effects surrounding these decisions are not supplied-stage effects.

**SMS-R04b** Supplied post-creation travel cue [verified in 1.13c code (static)]. At the successful client missile creation tail after its optional creation callback, sign-extend TravelSound. If it is nonzero, request it once with the newly created missile as emitter and three zero auxiliary values. Zero makes no request; negative values are forwarded. The request result does not determine whether the new missile is returned. An additional owner-cue lookup/retirement path remains SMS-Q03 and is not an assertion of this request stage.

**SMS-R04c** Supplied progression cue and counter [verified in 1.13c code (static)]. With valid missile record, if signed ProgSound <= 0, make no request and leave the previous counter unchanged. Otherwise read the previous missile counter and current path counter, supplying current 0 when its animation/path context is absent. Request ProgSound once only when the previous counter is nonzero and differs from current. Whether requested or not, then store current into the previous missile counter. The request uses the missile emitter and three zero auxiliary values and its result is ignored. No local random draw occurs. The following missile simulation operation remains outside this sound-decision stage.

**SMS-R04d** Missile and state integration/lifetimes [unresolved]. Full state-list production, complete missile creation, hit requests, contact/expiry stopping, per-projectile versus per-cast instance counts and voice association remain SMS-Q03. The supplied stages above do not establish audible lifetimes.

**SMS-R04e** Supplied admitted missile-hit cue [verified in 1.13c code (static)]. At the hit path after its collision and hit-callback admission gates, request signed HitSound once if it is nonnegative, using the missile emitter and three zero auxiliary values. Forward zero; suppress negative values. Ignore the sound operation's result. Earlier hit eligibility and callback execution, and later missile/voice retirement, remain SMS-R04d; this stage does not assert that every contact is admitted.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| SMS-R01 | supplied handler eligibility | verified in 1.13c code (static) | none locally; SMS-Q01 gates producer integration | SMS-V01, SMS-V02, SMS-V03, SMS-V06, SMS-V07, SMS-V09 |
| SMS-R02 | supplied type-specific request | verified in 1.13c code (static) | SMS-Q04 gates audible integration only | SMS-V03, SMS-V04, SMS-V05, SMS-V08, SMS-V10 |
| SMS-R03a | supplied progressive requests | verified in 1.13c code (static) | none locally; SMS-Q02 gates scheduling and producer integration | SMS-V11, SMS-V12 |
| SMS-R03b | other skill phases | unresolved | SMS-Q02: phase timing and progressive integration | none |
| SMS-R04a | supplied state-transition cues | verified in 1.13c code (static) | none locally; SMS-Q03 gates full state mutation | SMS-V13 |
| SMS-R04b | supplied travel request | verified in 1.13c code (static) | none locally; SMS-Q03 gates full creation and owner retirement | SMS-V14 |
| SMS-R04c | supplied progression request/counter | verified in 1.13c code (static) | none locally; SMS-Q03 gates missile simulation integration | SMS-V15 |
| SMS-R04d | missile and state lifetime | unresolved | SMS-Q03: full producing paths and voice lifetimes | none |
| SMS-R04e | supplied admitted hit request | verified in 1.13c code (static) | none locally; SMS-Q03 gates collision/callback admission and retirement | SMS-V16 |

## 5. Outputs and side effects

| Output | Rule |
|---|---|
| handler result 0 or 1 | SMS-R01, SMS-R02 |
| request log entry containing cue, emitter and three zero extra values | SMS-R02 |
| ordered overlay and sound request log, handler result | SMS-R03a |
| state-transition sound request log | SMS-R04a |
| created-missile travel request log | SMS-R04b |
| progression request log and previous-counter write | SMS-R04c |
| admitted missile-hit request log | SMS-R04e |

The SMS-R01/SMS-R02 type-specific handler writes no gameplay state. SMS-R04c separately writes the retained client progression counter listed above. Actual voice mutations remain SMS-Q04.

## 6. Branches and edge cases

For SMS-R01/SMS-R02, selector rejection precedes invalid-skill rejection. The selected signed-16 cue is sign-extended; zero and negative values generate no request in SMS-R02. A type other than 0 or 1 still returns success after a valid lookup, and playback failure does not turn that valid handler result into failure. SMS-R03a and SMS-R04a through SMS-R04e use their separately stated cue tests and state effects.

## 7. Open questions

**SMS-Q01** [unresolved] Establish complete state-dispatch scheduling, originating-skill stat production and the animation marker's complete frame/sequence progression. Located marker writes establish its animation source, not a gameplay action name or cadence. This gates integration, not supplied-input SMS-R01/SMS-R02.

**SMS-Q02** [unresolved] Trace stsound/dosound/prgsound consumers and progressive adjustment dependencies, including animation/event timing and delayed or success-only paths.

**SMS-Q03** [unresolved] Trace missile TravelSound/HitSound/ProgSound and state onsound/offsound producers, including start, contact, expiry and removal.

**SMS-Q04** [unresolved] Complete request admission and actual audio-voice lifetime through the [playback contract](sound-table-and-playback-rules.md), STP-R03b/STP-R04b. A logged request is not audible output.

**SMS-Q05** [unresolved] Reproduce verified selection vectors in an original running game with measured inputs. Observation only; [attempt record](../plans/gaps/sound-skill-and-missile-sounds.md).

## 8. Test vectors

### 8.1 Executable vectors

Each vector starts with an empty sound request log. The supplied request operation logs its call and has no other stored effects. Each vector specifies its opaque return so the handler's independence can be checked.

- **SMS-V01** Selector bypass (SMS-R01; boundary: supplied-input stage)
  - Inputs: selector=2; unit type=0; skill id=-1; count=0; record absent; request result=0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: selector differs from 3, so lookup is skipped.
  - Outputs: result=1; requests=[]; no lookup.
  - Discriminates: validating skill before selector would return 0.
  - runtime_observed: no
- **SMS-V02** Invalid skill boundary (SMS-R01; boundary: supplied-input stage)
  - Inputs: selector=3; emitter=unit A; unit type=0; skill id=4; count=4; supplied lookup would return a present record if called; dosound a=17; dosound b=29; request result=0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: id equals count and is rejected.
  - Outputs: result=0; requests=[]; no lookup.
  - Discriminates: inclusive upper bound would enter cue selection, request 17 and return 1.
  - runtime_observed: no
- **SMS-V03** First cue ignores playback failure (SMS-R01, SMS-R02; boundary: supplied-input stage)
  - Inputs: selector=3; emitter=unit A; unit type=0; skill id=1; count=4; record present; dosound a=17; dosound b=29; request result=0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: valid lookup; type 0 selects 17; 17>0 requests once; ignore result 0.
  - Outputs: result=1; requests=[(cue=17, emitter=unit A, extras=0,0,0)].
  - Discriminates: propagating request result returns 0; selecting second field requests 29.
  - runtime_observed: no
- **SMS-V04** Second cue rejects negative value (SMS-R02; boundary: supplied-input stage)
  - Inputs: selector=3; emitter=unit B; unit type=1; skill id=1; count=4; record present; dosound a=17; dosound b=-1; request result=1.
  - Seeds: no draw.
  - Draws: none.
  - Trace: type 1 selects signed -1; not strictly positive.
  - Outputs: result=1; requests=[].
  - Discriminates: a signed nonzero cue test incorrectly requests -1; an unsigned positive test incorrectly requests 65535.
  - runtime_observed: no
- **SMS-V05** Unsupported type (SMS-R02; boundary: supplied-input stage)
  - Inputs: selector=3; emitter=unit C; unit type=2; skill id=1; count=4; record present; dosound a=17; dosound b=29; request result=1.
  - Seeds: no draw.
  - Draws: none.
  - Trace: neither type comparison matches.
  - Outputs: result=1; requests=[].
  - Discriminates: treating all nonplayer units as type 1 requests 29.
  - runtime_observed: no

- **SMS-V06** Selector above 3 bypasses lookup (SMS-R01; boundary: supplied-input stage)
  - Inputs: selector=4; emitter=unit A; unit type=0; skill id=-1; count=0; record absent; request result=0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: 4 differs from 3; lookup is skipped despite invalid id.
  - Outputs: result=1; requests=[]; no lookup.
  - Discriminates: accepting selector>=3 incorrectly validates skill and returns 0.
  - runtime_observed: no
- **SMS-V07** In-range absent record (SMS-R01; boundary: supplied-input stage)
  - Inputs: selector=3; emitter=unit A; unit type=0; skill id=1; count=4; record absent; request result=0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: id is valid, but record lookup returns absent.
  - Outputs: result=0; requests=[].
  - Discriminates: treating absent records as successful no-sound selections incorrectly returns 1.
  - runtime_observed: no
- **SMS-V08** Second positive cue (SMS-R02; boundary: supplied-input stage)
  - Inputs: selector=3; emitter=unit B; unit type=1; skill id=1; count=4; record present; dosound a=17; dosound b=29; request result=0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: type 1 selects 29; 29>0 generates one request; ignore its zero result.
  - Outputs: result=1; requests=[(cue=29, emitter=unit B, extras=0,0,0)].
  - Discriminates: selecting first cue requests 17; skipping type 1 leaves an empty log; propagating request failure returns 0.
  - runtime_observed: no
- **SMS-V09** Negative skill rejection (SMS-R01; boundary: supplied-input stage)
  - Inputs: selector=3; emitter=unit A; unit type=0; skill id=-1; count=4; record absent; request result=0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: negative id is rejected before record lookup.
  - Outputs: result=0; requests=[]; no lookup.
  - Discriminates: an upper-bound-only signed check incorrectly performs a lookup for -1.
  - runtime_observed: no
- **SMS-V10** Zero cue rejection (SMS-R02; boundary: supplied-input stage)
  - Inputs: selector=3; emitter=unit A; unit type=0; skill id=1; count=4; record present; dosound a=0; dosound b=29; request result=1.
  - Seeds: no draw.
  - Draws: none.
  - Trace: type 0 selects 0; strictly positive test fails.
  - Outputs: result=1; requests=[].
  - Discriminates: accepting nonnegative cue values incorrectly requests 0.
  - runtime_observed: no

- **SMS-V11** Progressive clamp and ordered requests (SMS-R03a; boundary: supplied-input stage)
  - Inputs: unit A; skill id 1, count 4, record present; progressive true; aurastate 2, state count 4; `aurastat1` 10, stat count 12; associated list present; supplied stat value 5; prgoverlay 100; prgsound -1; overlay result 0; sound result 0.
  - Seeds: no draw.
  - Draws: none locally.
  - Trace: all adjustment gates pass; clamp 5 to 3, subtract 1 = 2; overlay 100+2=102; sound -1+2=1; call overlay before sound and ignore both failures.
  - Outputs: ordered log [overlay(102,A,2,0,0,0,0,0), sound(1,A,0,0,0)]; one stat read (associated list,10,layer 0); result 1.
  - (a) Inputs: stat value -3; all other inputs independently initialized as above. Trace: clamp to 1 gives adjustment 0. Outputs: ordered log [overlay(100,A,2,0,0,0,0,0), sound(-1,A,0,0,0)]; one stat read (associated list,10,layer 0); result 1.
  - Discriminates: no upper/lower clamp, inverted overlay/sound order, filtering nonpositive adjusted sounds, or propagating request failure changes exact outputs.
  - runtime_observed: no

- **SMS-V12** Progressive adjustment gates (SMS-R03a; boundary: supplied-input stage)
  - Inputs: unit A; skill id 1, count 4, record present; progressive false; aurastate 2, state count 4; `aurastat1` 10, stat count 12; associated list present; supplied stat value 2; prgoverlay -5; prgsound 0; overlay result 0; sound result 0.
  - Seeds: no draw.
  - Draws: none locally in every case.
  - Trace: progressive false bypasses adjustment; both unchanged identities are requested, including zero and negative values.
  - Outputs: ordered log [overlay(-5,A,2,0,0,0,0,0), sound(0,A,0,0,0)]; no stat read; result 1.
  - (a) Inputs: progressive true; aurastate 4. Outputs: same ordered log as base case; no stat read; result 1.
  - (b) Inputs: progressive true; `aurastat1` 12. Outputs: same ordered log as base case; no stat read; result 1.
  - (c) Inputs: progressive true; associated list absent. Outputs: same ordered log as base case; no stat read; result 1.
  - (d) Inputs: skill id -1. Outputs: empty request log; no stat read; result 0.
  - Discriminates: treating any adjustment gate as optional, allowing inclusive state/stat upper bounds, skipping zero/negative requests, or accepting an invalid skill changes the log, read or return.
  - runtime_observed: no

- **SMS-V13** State transition edge and signed cue policy (SMS-R04a; boundary: supplied-input stage)
  - Inputs: unit A, player mode 1; valid state record; transition on; previous membership absent; notondead false; onsound 0; offsound 21; request result 0.
  - Seeds: no draw.
  - Draws: none locally in every case.
  - Trace: no corpse exclusion; absent-to-on selects signed zero, which passes nonnegative check.
  - Outputs: requests=[(0,A,0,0,0)]; no membership writes within this sound-selection stage.
  - (a) Inputs: previous membership present. Outputs: requests=[]; no membership writes within stage.
  - (b) Inputs: transition off; previous membership present; offsound 0. Outputs: requests=[(0,A,0,0,0)]; no membership writes within stage.
  - (c) Inputs: transition off; previous membership absent. Outputs: requests=[]; no membership writes within stage.
  - (d) Inputs: onsound -1. Outputs: requests=[]; no membership writes within stage.
  - (e) Inputs: notondead true; player mode 17. Outputs: requests=[]; no membership writes within stage.
  - (f) Inputs: notondead false; player mode 17. Outputs: requests=[(0,A,0,0,0)]; no membership writes within stage.
  - Discriminates: copying the skill strictly-positive cue policy, replaying on-sound for an already present state, emitting off-sound for an absent state, or unconditionally excluding dead modes changes a case's request log.
  - runtime_observed: no

- **SMS-V14** Post-creation travel request (SMS-R04b; boundary: supplied-input stage)
  - Inputs: successful client creation result missile M after creation callback; TravelSound -1; request result 0; no owner cue to retire.
  - Seeds: no draw.
  - Draws: none locally in the travel request stage.
  - Trace: sign-extended -1 is nonzero; request once from M; ignore failure.
  - Outputs: requests=[(-1,M,0,0,0)]; creation result remains M.
  - (a) Inputs: TravelSound 0. Outputs: requests=[]; creation result remains M.
  - Discriminates: requiring a positive travel cue suppresses the base request; propagating request failure removes the creation result.
  - runtime_observed: no

- **SMS-V15** Progression counter transitions (SMS-R04c; boundary: supplied-input stage)
  - Inputs: missile M, valid record; ProgSound 77; previous counter 5; current path context present with current counter 6; request result 0.
  - Seeds: no draw.
  - Draws: none locally in every case.
  - Trace: 77>0 and previous 5 is nonzero and differs from 6; request then store current.
  - Outputs: requests=[(77,M,0,0,0)]; previous counter 6, written after request.
  - (a) Inputs: previous counter 0. Outputs: requests=[]; previous counter 6.
  - (b) Inputs: current counter 5. Outputs: requests=[]; previous counter 5.
  - (c) Inputs: ProgSound 0. Outputs: requests=[]; previous counter 5, no write.
  - (d) Inputs: current path context absent. Outputs: requests=[(77,M,0,0,0)]; previous counter 0, written after request.
  - Discriminates: requesting on the first nonzero counter, requesting on equality, suppressing the transition to missing-context zero, or conditioning writes on request success changes exact outputs.
  - runtime_observed: no

- **SMS-V16** Admitted hit cue policy (SMS-R04e; boundary: supplied-input stage)
  - Inputs: missile M; valid supplied missile record; hit path already admitted by collision and callback gates; HitSound 0; request result 0.
  - Seeds: no draw.
  - Draws: none locally in this request stage.
  - Trace: signed zero passes nonnegative test; request once and ignore return.
  - Outputs: requests=[(0,M,0,0,0)]; no retirement assertion.
  - (a) Inputs: HitSound -1. Outputs: requests=[]; no retirement assertion.
  - (b) Inputs: HitSound 23. Outputs: requests=[(23,M,0,0,0)]; no retirement assertion.
  - Discriminates: using the travel nonzero policy suppresses zero and forwards -1; using the skill strictly-positive policy suppresses zero.
  - runtime_observed: no

### 8.2 Observation vectors

**SMS-O01** Exercise the active handler on a unit with measured selector 3 and known skill record, recording the selected cue, request arguments and handler result. [Reproduction requirements](../research/spec-evidence/sound-skill-and-missile-sounds.md#observation).

## 9. Acceptance and dependencies

SMS-V01–SMS-V16 establish partial acceptance of supplied-input selection and request/counter stages. Full reproduction additionally requires SMS-Q01–SMS-Q04 and matching original-runtime observations for each verified rule. Depends on GDA-R11a for skill-record identity, GDA-R11a's progressive flag schema and GDA-R13 for state field identity; complete name linking remains gated. Actual playback consumes STP-R03b/STP-R04b, both unresolved. No damage or movement dependency is needed for these supplied stages.

## 10. Related specs

[Missile initialization](missile-initialization-and-registry.md) owns server preparation. [Missile movement and collision](missile-movement-and-collision.md) owns server lifetime and contact. [Timed states](state-timed-state-lifecycle.md) owns state lifetime. [Unit sound selection](sound-unit-sound-selection.md) owns other unit request selection.

## Appendix A. Provenance

Original Expansion 1.13c Windows x86, D2Client and D2Common identities in the [source manifest](../research/source-manifest.md). Static reading on 2026-10-29 supports the supplied handler stage. No shipped data cell or runtime observation is claimed. The [companion](../research/spec-evidence/sound-skill-and-missile-sounds.md) records corrected candidate identities and limitations.
