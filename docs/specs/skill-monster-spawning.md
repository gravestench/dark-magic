# Monster skill spawning and calls for help

**Description**

Some monster skills create new monsters, prepare a later spawn through a target, or call nearby creatures for help. Different skills use different preparation and placement paths. This contract identifies those paths and the minion spawner's ordered fallback requests; the complete effects of creation, infestation and help responses still require investigation.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [numeric and RNG foundations](core-numeric-and-rng-foundations.md), [record layouts and ids](core-record-layouts-and-ids.md); names: [game-data fields](core-game-data-fields.md), [game-data table records](core-game-data-table-records.md)
Ids: MSP-R, MSP-Q, MSP-V, MSP-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/skill-monster-spawning.md)
Open evidence gaps: [gap file](../plans/gaps/skill-monster-spawning.md)

## 1. Scope

The server start/action callbacks behind Nest, EvilHutSpawner, MinionSpawner, Impregnate and Cry Help in the original Expansion profile. This includes their local skill decisions and requested effects; generic monster creation, AI deciding when to invoke them, general spatial iteration and event/storage machinery are dependency boundaries. The supplied preparation/request stages do not establish complete monster initialization or target-effect execution.

## 2. Trigger and ownership

D2Game owns these callbacks, selected by `Skills.srvstfunc` and `Skills.srvdofunc`. MSP-R01 supplies the observed original-profile selector cells. The [skill-dispatch contract](skill-cast-start-and-do-dispatch.md) owns SCD-R03–SCD-R09c's start/action and frame routing with its existing gates. Exact AI choice/cadence, event latency and client notification remain MSP-Q05. The fallback stage begins after current-skill fields and a usable room have been supplied and ends when a creator result is selected; it does not include later created-unit initialization.

## 3. Inputs and data

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| skill id / loaded count | callback argument / Skills loader | signed i32 | Guard input |
| caster / current skill | selected unit and active skill state | logical values | Preparation and identity; full integration gated |
| level | callback level | signed i32 | Calculation/preparation input |
| creation class, creation mode | supplied current-skill prepared values | signed i32 | Same values forwarded on both fallback requests |
| position x/y | supplied prepared position | signed i32 | Same position forwarded on both requests |
| room and game context | supplied loaded context | logical identities | Same context forwarded on both requests |
| creator results | supplied ordered operation results | created-unit identity or empty | Selects whether a fallback is requested |
| request log | supplied creator-operation log | ordered requests | Records class/mode/context/position/placement selector/final option |
| target, help state, delayed effect | selected target, frame, search/filter and event inputs | logical values | Full Impregnate/Cry Help integrations are gated |

| Mapping / constant | Origin | Binding | Evidence |
|---|---|---|---|
| MSP-R01 selector cells | table data | Data selects engine-defined handlers | MSP-D01 |
| Initial placement selector -1, fallback selector 4, final option 0 | engine code | Fixed by original minion callback | MSP-F03 |
| Skill-prepared class/mode/position values | runtime state | Produced by start/preparation dependencies | MSP-Q01/MSP-Q04 |

Additional supplied-stage inputs are:

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| spawn class / offsets / mode | Selected MonStats.spawn / spawnx / spawny / spawnmode | signed 16-bit class; signed 8-bit offsets/mode | Primary monster-profile preparation; [local loader/consumer mapping](../research/spec-evidence/skill-monster-spawning.md#successful-preparation-followup) |
| summon class / mode | Selected Skills.summon / summode | signed 16-bit class; signed 8-bit mode | Alternative skill-record preparation; local loader/consumer mapping |
| actor position | Supplied actor position | signed 32-bit coordinates | Position before adding spawn offsets; dynamic-path original reads zero-extend 16-bit components |
| loaded class count | MonStats collection | signed 32-bit | Primary and fallback class validation |
| AI fallback class | Supplied actor AI context | signed 32-bit | Used only when primary class is invalid; producer remains Q04 |
| cached spawn tuple | Current skill's prepared class/x/y/mode fields | four signed 32-bit values | Written by starts and later read by actions |
| target context | Supplied selected distinct target | unit identity, type and id | Entry after target normalization/eligibility, which retains its dependency gates |
| expression result, game frame | Selected Skills.`calc1` operation / game clock | signed 32-bit result, unsigned 32-bit frame | Cry Help request deadline; SCX-R11 evaluation supplied |
| overlay selector / count | Selected skill overlay field / Overlay collection | signed 16-bit selector, signed 32-bit count | Controls optional target/created-unit overlay request; field-name binding remains Q01 |
| supplied operations | Class preparation, collision, creator, effect producer, help propagation and overlay operations | ordered logged operations | Vector operations log calls and return supplied values, with effects/draws only when explicitly stated |

## 4. Rules (execution order)

**MSP-R01** **Named original text selectors** [verified in 1.13c data]. Exact cells of effective `data/global/excel/skills.txt` are Nest.`srvstfunc` = 49 and `.srvdofunc` = 91; EvilHutSpawner.`srvstfunc` = 49 and `.srvdofunc` = 91; MinionSpawner.`srvstfunc` = 62 and `.srvdofunc` = 135; Impregnate.`srvdofunc` = 133; Cry Help.`srvdofunc` = 128. These are text cells, not claims of actual action execution; see MSP-D01 and its narrow compiled-selector corroboration.

**MSP-R02** **Invalid action id** [verified in 1.13c code (static)]. At the nest, minion-spawn, impregnation or cry-for-help action entry, if signed skill id < 0 or skill id >= loaded Skills count, return signed 0 before helper calls, RNG or gameplay writes. Loaded table storage is assumed valid for the supplied count. Successful and malformed-storage cases are not inferred from this guard.

**MSP-R03** **Minion creation fallback with supplied operation results** [verified in 1.13c code (static)]. Given prepared creation class/mode/position, game context and a nonempty usable room, first request creation with placement selector -1 and final option 0. If that request returns a nonempty unit identity, select it and issue no fallback. If it returns empty, request creation once more with placement selector 4 and final option 0. Select the second result, including empty. Both requests carry identical class, mode, game, room and position. There is no loop or third attempt. The stage draws no RNG locally; the supplied creator logs each request and returns its supplied identity/empty value without other effects or draws. The selectors are opaque placement arguments; this rule does not assign a search geometry to them. Actual creator effects and subsequent selected-unit owner-link/stat/overlay effects remain MSP-Q01, and the stage result is the selected unit identity, not the outer callback's final result.

**MSP-R04** **Spawner preparation and finalization** [unresolved]. Nest/EvilHutSpawner and MinionSpawner use start/preparation values from the current skill and request unit creation. MSP-R07/R08 establish supplied-record class/mode/position preparation and cache writes; complete producer integration, event effects, created-unit conditions, owner-link/stat/overlay operations and callback success retain MSP-Q01/MSP-Q04. MSP-R03 does not establish the effects of either real creator call.

**MSP-R05** **Impregnation** [unresolved]. The action obtains a selected target, tests eligibility and issues the state request of MSP-R10 to the shared effect producer, not a spatial iterator. The exact trigger for later creation, target lifetime and replacement/repeat behavior are gated by MSP-Q02. A successful action result is not asserted to create a monster immediately.

**MSP-R06** **Calling for help** [unresolved]. The action obtains a selected target and uses the calculation/request order of MSP-R11 before optional overlay handling. Target eligibility, response timing, which creatures respond and all shared callback effects are gated by MSP-Q03. The action's name alone does not establish a radius, guaranteed summons or a direct spawn.

**MSP-R07** **Primary spawn tuple and class fallback** [verified in 1.13c code (static)]. At the shared preparation boundary with a supplied valid selected actor MonStats record, actor position and AI fallback context, when monster-spawn preference is nonzero and actor is a monster, take signed `MonStats.spawn` as the primary class. Compute x/y as low32(actorX + signExtend(spawnx)) and low32(actorY + signExtend(spawny)). Sign-extend `spawnmode`; retain 0 through 15 inclusive, otherwise replace it with 1. The alternative boundary, used when preference is zero or actor is not a monster, takes signed Skills.summon and summode from a supplied valid skill record; it rejects an out-of-range summon class before writing the tuple. With a valid summon class, x/y are the actor position unchanged and mode uses the same 0..15 clamp.

The wrapper accepts a primary class >=0 and <loaded class count. Otherwise, with a supplied usable actor AI fallback record, accept its fallback class only when strictly >0 and <count; else return -1. This fallback does not roll back primary x/y/mode writes already performed. Error paths where the lower producer never wrote output fields require the original output context and remain Q04; this contract does not invent initial values for them. No RNG is drawn locally. The caller's skill level is forwarded to preparation but neither of these established record-selection branches computes with it.

**MSP-R08** **Cache start-prepared creation values** [verified in 1.13c code (static)]. With a present current skill and a supplied preparation result (class,x,y,mode), start 62 writes those four values into that current skill in order class,x,y,mode and returns 1. It does not reject a negative supplied class. Start 49 rejects a negative supplied class with result 0 before any of these four writes; a nonnegative class causes the same four ordered writes. After those writes start 49 continues into room/collision/preparation/event handling under Q04; its final result and those effects are not established by this cache stage. The four setters persist the values; they are not merely logged requests. This stage draws no RNG locally and excludes draws owned by preparation.

**MSP-R09** **Nest creation request after preparation** [verified in 1.13c code (static)]. With a supplied present coordinate room and nonzero prepared x/y, request collision processing at that room and position with controls (1,256), then request creation once with the prepared class/mode, same game/room/position, placement selector 1 and final option 0. Do not choose the minion fallback selectors -1/4 for this nest action. The supplied collision operation's return is ignored. The stage selects the creator's returned unit or empty result; it does not establish created-unit flags, overlay or outer callback success. If room is absent, or either coordinate equals zero, reject this local stage with selected result empty before collision/creation. This stage begins after the nest's event/preparation operations and excludes their effects and RNG. Real collision/creator effects remain Q01/Q04; a supplied operation can log the exact request without producing a monster.

**MSP-R10** **Impregnate effect request** [verified in 1.13c code (static)]. After valid skill lookup set the actor's do-handled latch, then consume the target normalization/eligibility results supplied at this boundary; with an eligible distinct target, prepare an effect request for actor, target, skill identity and caller level. Request state identity 110, duration 0, no primary stat, primary value 0, and the infestation removal operation; no additional request field selects a timed expiry. Invoke the shared effect producer once and return 1 regardless of its normally returned identity/empty result. The producer's eligibility, stored state/list, removal timing and callback effects remain Q02; duration 0 here is a prepared request value, not a claim that an actual timed event occurs immediately or that a monster is created by this action. This local request stage draws no RNG; it makes no radius iteration request.

**MSP-R11** **Cry Help calculation and propagation request** [verified in 1.13c code (static)]. With a supplied selected distinct target and monster actor, request the invoking skill's `calc1` under SCX-R11 with actor, skill and caller level. Let d be its signed 32-bit return, replaced by 1 when d<1. Compute deadline=low32(current game frame + d). Request help propagation with kind 1, target type/id and deadline; unestablished remaining propagation fields and actual responders remain Q03. Ignore the propagation operation's normally returned value. Then, if the selected signed 16-bit overlay selector is strictly >0 and <loaded Overlay count, request overlay application to the target with trailing control 0. Selector0, negatives and equality with count skip this optional operation. Return1 after the normally returning operations. This stage does not infer responder cadence or future execution from a computed deadline. It draws nothing locally; expression, propagation and overlay effects retain their owning integration gates.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| MSP-R01 | Named selector cells | verified in 1.13c data | none for cells; AI/frame entry MSP-Q05 | MSP-D01 in companion |
| MSP-R02 | Invalid-id exit | verified in 1.13c code (static) | none locally | MSP-V01 |
| MSP-R03 | Minion fallback request order | verified in 1.13c code (static) | MSP-Q01: actual creation/finalization only; supplied-result selection available | MSP-V02 |
| MSP-R04 | Spawner preparation/finalization | unresolved | MSP-Q01/MSP-Q04: successful creation, fields and final effects | none |
| MSP-R05 | Impregnation effect | unresolved | MSP-Q02: target-associated and delayed effects | none |
| MSP-R06 | Help effect | unresolved | MSP-Q03: actual help/response effects | none |
| MSP-R07 | Spawn tuple and primary/fallback class selection | verified in 1.13c code (static) | Q04 producer/profile integration; supplied records/context available | MSP-V03 |
| MSP-R08 | Persist class/x/y/mode in current skill | verified in 1.13c code (static) | Q04 start 49 continuation only; supplied preparation/cache stage available | MSP-V04 |
| MSP-R09 | Nest collision then single creation request | verified in 1.13c code (static) | Q01/Q04 actual effects; supplied requests available | MSP-V05 |
| MSP-R10 | Impregnate state request and ignore producer result | verified in 1.13c code (static) | Q02 actual state/removal/spawn effects | MSP-V06 |
| MSP-R11 | Help deadline/request and optional overlay order | verified in 1.13c code (static) | Q03 actual responders/unestablished fields; supplied request subset available | MSP-V07 |

## 5. Outputs and side effects

| Output / effect | Rule |
|---|---|
| Callback result 0, no helper/RNG/gameplay writes on invalid-id branch | MSP-R02 |
| Ordered creation requests, with initial/fallback selection result | MSP-R03 |
| Created units and full spawning owner/stat/overlay/finalization effects, gated | MSP-R04 |
| Actual infestation lifecycle/removal/creation effect, gated; prepared request and local result are R10 | MSP-R05 |
| Actual help/responder effect, gated; prepared request subset is R11 | MSP-R06 |
| Prepared spawn tuple and chosen class | MSP-R07 |
| Four persisted current-skill preparation fields | MSP-R08 |
| Ordered nest collision/creator requests and selected result | MSP-R09 |
| Actor latch, target-associated state 110 request and result 1 | MSP-R10 |
| Calc request, computed help deadline/request subset, optional overlay request and result 1 | MSP-R11 |

No stored creation effect or named client message follows from supplied requests alone. R08 establishes only current-skill preparation persistence.

## 6. Branches and edge cases

Equality with the loaded count fails the signed-id guard. The minion fallback is conditional on an empty first result, not on the amount of space visibly available; the first successful result suppresses the fallback. If both requests fail, the stage's selected result is empty and later callback handling remains MSP-Q01. The stage performs no unit-type or difficulty check beyond its explicit prepared-input boundary; actual restrictions belong to the gated preparation. It does not equate -1/4 with modes, nor presume that the two creation helpers have no RNG effects in the real game. Other builds and Classic execution are outside the evidence profile.

## 7. Open questions

| Question | Status | Evidence that settles it |
|---|---|---|
| **MSP-Q01** | integration unresolved | Complete creator, selected-unit condition/initialization, optional-overlay and callback-result effects for nest/minion creation |
| **MSP-Q02** | unresolved | Actual target eligibility, shared state/list production and removal callback effects; R10 supplied request available for Impregnate |
| **MSP-Q03** | unresolved | Actual owner-linked propagation, remaining request fields, timing and responder behavior; R11 supplied request subset available for Cry Help |
| **MSP-Q04** | unresolved | Remaining actual record/AI fallback production, malformed-output paths, start 49 continuation and nest event/room preparation; R07/R08 supplied tuples/cache writes available |
| **MSP-Q05** | integration unresolved | Actual AI/frame entry; text and narrow loaded selectors agree but do not prove execution |
| **MSP-Q06** | observation pending | Naturally reached creation/fallback, infestation and help captures; invalid-id entry may be excluded by dispatcher and remains static |

The [gap file](../plans/gaps/skill-monster-spawning.md) records attempted slices and exact blocked assertions.

## 8. Test vectors

### 8.1 Executable vectors

Each case is independent. Callback guard vectors start with empty helper/write logs. Fallback cases begin inside MSP-R03 after preparation, with an empty request log. The supplied creator logs calls, returns the listed result and has no RNG or stored effects. No later initialization, state or outer-callback result is asserted.

- **MSP-V01** Invalid-id entry (MSP-R02; boundary: isolated rule)
  - Inputs: loaded Skills count = 357, caster absent, level 1. For each nest, minion-spawn, impregnation and cry-for-help action run (a) skill id -1; (b) skill id 357.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: signed lower or exclusive upper comparison fails before any helper call.
  - Outputs: each independent case result = 0; helper log = []; write log = []; seeds unchanged.
  - Discriminates: rejects inclusive upper bounds or helper/unit work before id validation.
  - runtime_observed: no

- **MSP-V02** Exact minion fallback order (MSP-R03; boundary: supplied-input stage)
  - Inputs: creation class = 7 (synthetic supplied id, not a named species assertion), mode = 3, game = G, room = R, position = (100,200); request log empty; cases below supply creator results.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: initial selector -1; only an empty result permits one selector-4 fallback.
  - Cases:
    - (a) supplied results [unit A]. Outputs: log = [(class 7,mode 3,G,R,100,200,selector-1,option 0)]; stage selected result = unit A.
    - (b) supplied results [empty, unit B]. Outputs: log = [(class 7,mode 3,G,R,100,200,selector-1,option 0),(class 7,mode 3,G,R,100,200,selector 4,option 0)]; selected result = unit B.
    - (c) supplied results [empty, empty]. Outputs: log = [(class 7,mode 3,G,R,100,200,selector-1,option 0),(class 7,mode 3,G,R,100,200,selector 4,option 0)]; selected result = empty; no third request.
  - Outputs: only the case-specific request log and selected result; no unit state changes; all seeds unchanged.
  - Discriminates: (a) rejects unconditional fallback; (b) rejects reversed selectors, changed coordinates/context, or ignoring the second result; (c) rejects an unbounded retry loop or treating failure as success. All logged requests reject swapping distinct class 7 and mode 3.
  - runtime_observed: no

- **MSP-V03** Spawn tuple and class selection (MSP-R07; boundary: supplied-input stage)
  - Inputs: actor monster, valid selected actor MonStats record, position (100,200), loaded class count 400, monster preference 1; spawnx=-3, spawny=5; AI fallback context present. Cases override class/mode/fallback or choose the valid skill-record alternative. Output tuple fields are observed separately from the returned class.
  - Seeds: no draw.
  - Draws: none.
  - Trace: choose monster or skill record, prepare coordinates/mode, validate primary class, then apply strictly-positive fallback only when primary invalid.
  - Cases:
    - (a) Inputs: spawn=7, spawnmode=16, fallback=9. Outputs: tuple(x=97,y=205,mode=1), returned class 7; fallback unused.
    - (b) Inputs: spawn=0, spawnmode=0, fallback=9. Outputs: tuple(x=97,y=205,mode=0), returned class 0; fallback unused.
    - (c) Inputs: spawn=400, spawnmode=3, fallback=9. Outputs: tuple(x=97,y=205,mode=3), returned class 9; prepared coordinates/mode not rolled back.
    - (d) Inputs: spawn=-1, spawnmode=-1, fallback=0. Outputs: tuple(x=97,y=205,mode=1), returned class-1; fallback 0 is rejected.
    - (e) Inputs: preference 0; valid Skills record with summon=7,summode=-1; actor still monster, position(100,200), class count 400, fallback 9. Outputs: tuple(x=100,y=200,mode=1), returned class 7; monster offsets not applied.
  - Outputs: each case only the exact tuple/returned class above, with no local seed update or stored monster creation.
  - Discriminates: rejects unsigned spawn offsets, clamping mode 0, accepting mode 16, applying monster offsets to skill-record mode, inclusive primary upper bound, or accepting fallback 0 as valid.
  - runtime_observed: no

- **MSP-V04** Start cache writes (MSP-R08; boundary: supplied-input stage)
  - Inputs: current skill present with cached tuple(class 9,x 10,y 11,mode 2); supplied preparation tuple(class 7,x 97,y 205,mode 1); each case starts independently and observes the four setter operations and persisted tuple.
  - Seeds: no draw; preparation's earlier operations are outside this vector.
  - Draws: none.
  - Trace: apply start-specific class check, then write class,x,y,mode in order.
  - Cases:
    - (a) Inputs: start 62. Outputs: write log[class 7,x 97,y 205,mode 1], persisted tuple(7,97,205,1), final result 1.
    - (b) Inputs: start 62; supplied tuple(-1,97,205,1). Outputs: write log[class-1,x 97,y 205,mode 1], persisted tuple(-1,97,205,1), final result 1.
    - (c) Inputs: start 49; supplied tuple(-1,97,205,1). Outputs: write log[], cached tuple remains(9,10,11,2), final result 0.
    - (d) Inputs: start 49. Outputs: write log[class 7,x 97,y 205,mode 1], persisted tuple(7,97,205,1); assertion boundary stops after the fourth write, before subsequent room/collision/event operations; no outer result is asserted.
  - Outputs: only case-specific persisted values/logs and explicitly named results; no local seed change.
  - Discriminates: start 62 case(b) rejects copying start 49's negative-class rejection; case(c) rejects partial writes before rejection; distinct x/y/mode reject setter swaps; persisted values reject a log-only implementation.
  - runtime_observed: no

- **MSP-V05** Nest single creator request (MSP-R09; boundary: supplied-input stage)
  - Inputs: class 7,mode 3,gameG,coordinate roomR,position(100,200); prior nest event/preparation phase already completed; collision operation logs arguments and returns 0 without draws/effects; creator logs arguments, draws nothing and returns supplied results below.
  - Seeds: no draw in this post-preparation stage.
  - Draws: none.
  - Trace: check room/nonzero coordinates, issue collision controls(1,256), ignore its return, issue exactly one creator request with selector 1/option 0.
  - Cases:
    - (a) Inputs: creator returns unitA. Outputs: ordered log[collision(R,100,200,1,256),create(class 7,mode 3,G,R,100,200,selector 1,option 0)], stage selected result unitA.
    - (b) Inputs: creator returns empty. Outputs: ordered log[collision(R,100,200,1,256),create(class 7,mode 3,G,R,100,200,selector 1,option 0)], stage selected result empty; no fallback.
    - (c) Inputs: coordinate x=0,y 200,roomR. Outputs: log[], stage selected result empty.
    - (d) Inputs: coordinate x 100,y=0,roomR. Outputs: log[], stage selected result empty.
    - (e) Inputs: room empty,position(100,200). Outputs: log[], stage selected result empty.
  - Outputs: case-specific logs/selected result, no stored monster changes or outer callback-result assertion, no local seed update.
  - Discriminates: collision result 0 does not suppress creation; selector 1 differs from minion selectors-1/4; creation failure does not cause a second attempt; each coordinate guard prevents operations independently.
  - runtime_observed: no

- **MSP-V06** Impregnate request independent of producer result (MSP-R10; boundary: supplied-input stage)
  - Inputs: valid skill 300,level 2,gameG,actorA with latchfalse; supplied normalized targetT distinct fromA and eligible; target eligibility/normalization effects outside the assertion boundary; effect producer logs semantic request and captures the actor latch at its entry, draws nothing and has no stored effects; independent cases return empty or listL.
  - Seeds: no draw locally.
  - Draws: none.
  - Trace: set latch, consume eligible target context, prepare state request, call producer once and ignore its normal return.
  - Outputs: each independent producer-return case has final latchtrue and captured producer-entry latch=true; log[effect(actorA,targetT,skill 300,level 2,duration 0,noPrimaryStat,primaryValue0,state 110,infestationRemovalOperation)]; final result 1; no immediate creation request or local seed update asserted.
  - Discriminates: replacing state 110/duration 0, calling a radius iterator, or forwarding an empty/list result instead of returning 1 fails the request log/final result. Moving the latch write after producer entry fails its captured entry latch=true even if the final latch is true.
  - runtime_observed: no

- **MSP-V07** Help request deadline and overlay order (MSP-R11; boundary: supplied-input stage)
  - Inputs: actorA monster, supplied distinct targetT(type 1,id 700), skill 290,level 2; `calc1` referenceC; expression logs(actorA,C,290,2) and returns case result without draws; help propagation logs only the established kind/target/deadline subset, leaves unestablished remaining request fields unasserted, has no effects/draws and normally returns 37; overlay operation logs target/selector/control and has no effects/draws. Each case has fresh logs.
  - Seeds: no draw locally or in the supplied operations.
  - Draws: none.
  - Trace: request `calc1`, clamp signed result below 1 to 1, add frame with low32 wrapping, propagate known request subset, then optionally request overlay; ignore supplied propagation result 37 and return 1.
  - Cases:
    - (a) Inputs: frame 10,calc result-5,overlay 5,count 10. Outputs: ordered log[calc(actorA,C,290,2),help(actorA,kind 1,targetType 1,targetId 700,deadline 11),overlay(targetT,5,control 0)], final result 1.
    - (b) Inputs: frame 4294967294,calc result 3,overlay 0,count 10. Outputs: ordered log[calc(actorA,C,290,2),help(actorA,kind 1,targetType 1,targetId 700,deadline 1)], final result 1; no overlay.
    - (c) Inputs: frame 10,calc result 1,overlay 10,count 10. Outputs: ordered log[calc(actorA,C,290,2),help(actorA,kind 1,targetType 1,targetId 700,deadline 11)], final result 1; no overlay.
  - Outputs: only the case-specific request subset/log and callback result, no actual responder or target-state write, no seed update.
  - Discriminates: negative deadline offsets, saturating frame addition, accepting overlay 0/equality, applying overlay before propagation, or forwarding producer 37 each fail an explicit assertion.
  - runtime_observed: no

### 8.2 Observation vectors

**MSP-O01** Observe Nest/EvilHutSpawner and MinionSpawner in separate naturally eligible encounters, capturing start-prepared inputs, actual first/fallback creation calls, returned identities and later initialization/state. [Procedure](../research/spec-evidence/skill-monster-spawning.md#observation-procedure) also identifies the pending Impregnate/Cry Help observations. Compare exact request order to MSP-V02; full effects remain gated until their dependencies are established.

## 9. Acceptance and dependencies

Partial acceptance is MSP-R01's named cells, MSP-V01 for MSP-R02, MSP-V02 for supplied-result MSP-R03 and MSP-V03–MSP-V07 for their explicitly supplied preparation/request stages MSP-R07–MSP-R11. Full reproduction requires resolving MSP-Q01–MSP-Q05 and executable integration vectors for MSP-R04–MSP-R06. A passed fallback schedule establishes no complete monster initialization, infestation or help response.

**Depends on:** SCD-R03–SCD-R09c for shared callback entry with existing gates; NRF-R01/NRF-R02 for no-draw seed semantics. Generic creation, state, event and target-iteration contracts required for full behavior are not replaced by the supplied operation in MSP-R03.

**Depended on by:** no known consumed plan, implementation or test contract.

[Skill calculations](skill-calc-expressions.md) SCX-R11 supplies Cry Help's expression front end and retains expression-production/draw gates. SCD-R01/SCD-R07/SCD-R09 name the do-handled latch. The shared field setters used by R08 are closed local dependencies, not a new skill-state authority.

## 10. Related specs

- [Skill cast start and do dispatch](skill-cast-start-and-do-dispatch.md): shared callback dispatch and frame boundary.
- [Monster scaling](monster-scaling-by-players-and-difficulty.md): downstream creation/scaling relationships, retaining its own gates.
- [Maggots and burrowing](skill-monster-maggots-and-burrowing.md): another consumer of general monster placement; distinct hatch-attempt schedule.

## Appendix A. Provenance

Windows x 86 Expansion 1.13c original `LOD113C-D2GAME` in the [source manifest](../research/source-manifest.md). Original dispatch slots and instructions were read on 2026-10-29; original-image decoding corroborated encoded minion/nest branches. Selected imported current-skill field readers were read through byte-equivalent reconstruction. Effective Skills text and narrow loaded selectors agreed on that date. No gameplay vector is runtime-confirmed. The [companion](../research/spec-evidence/skill-monster-spawning.md) records exact boundaries and remaining dependencies.
