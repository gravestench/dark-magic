# Ambient and weather sound

**Description**

The game selects background ambience and occasional environmental sounds from the current area's sound environment. It replaces ambience when the selection changes, adjusts weather sound toward the selected intensity, and requests occasional sounds at varied positions. The request sequence and direct timing/position choices are established with supplied playback results; actual audible playback remains unresolved.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [numeric and RNG foundations](core-numeric-and-rng-foundations.md)
Ids: SAW-R, SAW-Q, SAW-V, SAW-O; policy in [README conventions](README.md#conventions)
Evidence: [companion records](../research/spec-evidence/sound-ambient-and-weather.md)
Open evidence gaps: [gap file](../plans/gaps/sound-ambient-and-weather.md)

## 1. Scope

Environment-based ambience/event selection, event scheduling, weather-related sound eligibility and transitions. Music selection belongs to its separate sound contract. Screen-shake audio and developer random-sound playback are not treated as weather based on analyst labels.

## 2. Trigger and ownership

D2Client invokes the examined ambient controller from the enabled audio update before music coordination. It requires a local player, room and level. The verified stages below start with an eligible update and supplied phase/environment inputs; their producers and resulting audible playback remain SAW-Q01 and SAW-Q03. Server-frame cadence and network-message policy are not established.

## 3. Inputs and data

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| phase | Current act world-clock cycle, WCK-R01 through WCK-R03; supplied | i32 | Values 1 through 3 select the first pair; other values select the second |
| environment | Selected environment record, supplied | Four sound identities, u32 delay | `Day Ambience`, `Night Ambience`, `Day Event`, `Night Event`, `Event Delay` |
| selected event and previous event | Selection output and retained scheduler state | i32 identities | Distinguish a changed event selection |
| tick, last tick, interval | Client audio update counter and scheduler | u32 | Wrapped elapsed-time comparison |
| weather suppression value | Client weather state, supplied | i32 | Nonzero suppresses the weather predicate and intensity |
| local player present | Client player context, supplied | Boolean | Absence suppresses weather |
| level Rain and Mud | Prepared Levels record, supplied | u8 each | Rain alone controls this predicate; Mud does not alter these returns |
| stored weather intensity | Client weather state, supplied | finite 32-bit binary floating point | Returned unchanged when weather is eligible; producer remains gated |
| random state | Local client's player seed, supplied | Pair of u32 words | Direct timer draws; independent initial value required |
| retained ambience identity/handle | Controller state | signed 32-bit identity; unsigned 32-bit handle | Previous ambience and its playback request result |
| retained weather identity/handle | Controller state | signed 32-bit identity; unsigned 32-bit handle | Previous weather cue and request result; distinct from the current selected weather cue |
| first/second intensity readings | SAW-R04a accessor results, supplied | finite `binary32` | First zero suppresses conversion; otherwise convert the second reading |
| current weather amplitude | Supplied volume-read operation | signed 32-bit | Current amplitude associated with weather handle |
| level identity and retained last level | Initial eligible level and controller state | signed 32-bit | Control final auxiliary eligibility requests |
| environment identity | Prepared level SoundEnv, supplied | unsigned byte, promoted to 32-bit | Acoustic-update request input |
| auxiliary eligibility A/B | Supplied environment fields | signed 32-bit identities | Two optional nonzero enable requests; header binding and downstream effects unresolved |
| supplied audio operations | Request, transition, volume, cleanup, stop, position, acoustic and auxiliary operations below | Ordered call log, prescribed returns and seed effects | Define the local caller boundary; real voice/backend execution remains gated |

| Mapping / constant | Origin | Binding |
|---|---|---|
| Phase interval 1..3, division by 3 | Engine code | Fixed by original code |
| Named environment fields | Table data with client loader descriptors | Resolved from supplied record; text/compiled identity remains SAW-Q01 |
| Weather cue 64; transition duration 250; amplitude multiplier 255; amplitude step 6; event distance 450..750 and vertical offset -100..100 | Engine code | Fixed by original code; cue 64 has no table-name assertion here |

No original cell value is required by the synthetic executable vectors. [T01](../research/spec-evidence/sound-ambient-and-weather.md#t01-named-data-cells) records narrow original-profile cells separately.

## 4. Rules (execution order)

**SAW-R01** Environment pair selection [verified in 1.13c code (static)]

With a valid supplied environment, if signed `1 <= phase <= 3`, select its `Day Ambience` and `Day Event`. Otherwise select `Night Ambience` and `Night Event`. Copy identities unchanged. This stage makes no random draw. The phase is the current act world-clock cycle, not its period field. WCK-R01 through WCK-R03 own its progression; no clock-hour mapping is introduced here.

**SAW-R02** Initial event timing [verified in 1.13c code (static)]

Boundary: enter the timer stage after preceding ambient/weather calls. If selected event is zero, leave scheduler state and seed unchanged. If it equals the previous event, retain interval and last tick, with no direct timer draw. Otherwise, let `D` be u32 `Event Delay`, `q = floor(D/3)`, and `m = low32(2*q+1)`. Interpret draw moduli as i32. A draw with nonpositive modulus returns zero without advancing the seed; positive draws consume exactly one NRF-R01 step and return new low-word remainder modulo that positive value (NRF-R02).

Draw `a` with modulus `m`. Store `interval = low32(D + a - q)`. Draw `b` with modulus interpreted from interval. Store `lastTick = low32(tick - b)` and `previousEvent = selectedEvent`. Writes and arithmetic are u32, wrapping at 32 bits. Seed owner is the local client's player seed; this does not establish the seed's initialization or draws made by preceding playback calls. This rule covers initialization only; the same update subsequently evaluates event eligibility under SAW-R04b.

**SAW-R03** Named environment cells [verified in 1.13c data]

For `ESOUNDENVIRON_WILDERNESS`, explicit `Index = 2`, the effective text values are `Day Ambience = 70`, `Night Ambience = 71`, `Day Event = 192`, `Night Event = 197`, and `Event Delay = 250`. These named cells in [T01](../research/spec-evidence/sound-ambient-and-weather.md#t01-named-data-cells) do not establish compiled runtime correspondence or audible behavior.

**SAW-R04a** Weather predicate and intensity selection [verified in 1.13c code (static)]

Boundary: the two weather accessors with supplied suppression, player-presence, prepared Rain/Mud and stored-intensity inputs. A nonzero suppression value or absent local player returns weather eligibility false and intensity positive zero. Otherwise weather eligibility is true exactly when the unsigned Rain byte is nonzero, and intensity is the stored weather intensity in that case, positive zero otherwise. Mud does not change either return. These accessors draw no random number and make no persistent write. This stage does not generate, scale, clamp or convert intensity, or establish audible playback.

**SAW-R04b** Playback, weather and subsequent event execution [unresolved]

Actual ambience/weather/event playback, voice allocation, cleanup/stop lifetime, acoustic and auxiliary effects remain gated by SAW-Q02 and SAW-Q03. SAW-R05 establishes the caller's supplied-operation request log and direct state/draw effects, rather than these nested operations.

**SAW-R05** Eligible ambient-controller orchestration [verified in 1.13c code (static)]

Boundary: one eligible update with a valid supplied environment and player seed, with all supplied operations returning normally. Consume SAW-R01's selected ambience/event and SAW-R04a's weather eligibility/intensity results. Callback names below describe semantic operations, not required APIs. Each operation logs its exact arguments; prescribed operation results and seed mutations are supplied inputs and occur before the caller continues. `RequestSound(cue)` requests that cue with absent emitter, zero delay, zero request flags and zero opaque playback value. It returns the supplied handle. `SetVolume(handle,value)` and `ReadVolume(handle)` operate on the supplied handle-amplitude state; zero/missing handles make no volume write and read as zero. `Schedule(handle,target,duration,delay)` logs a request to STP-R05; its actual voice-state execution requires the supplied playback dependency. Stop, cleanup, position, acoustic and auxiliary operations log their inputs; their nested effects remain supplied/gated.

All integer arithmetic retains low32. Compare identities and amplitudes signed; compare elapsed ticks and intervals unsigned. Direct draws use NRF-R01/NRF-R02 on the same local client player seed, including any supplied intervening callback seed changes. No draw is omitted merely because it returns zero.

1. Compute current weather identity as 64 if weather eligibility is true, otherwise 0. Set target weather amplitude to 0 if the first intensity reading compares equal to positive zero (including negative zero). Otherwise set it to the signed integer truncation toward zero of `second intensity reading * 255`. This verified conversion boundary requires finite readings with the product inside signed 32-bit range; invalid/out-of-range conversion remains SAW-Q02. Do not clamp the amplitude.
2. If selected ambience differs from retained ambience, determine swap=true exactly when the new/old identities are the environment's Day/Night Ambience in either order. If swap, request `Schedule(retained ambience handle,0,250,0)`. Otherwise request `KeepAmbient(selected ambience,current weather identity)`, then `KeepEvent(selected event,0)`, preserving that order. If selected ambience is nonzero, request it and store the returned handle even if zero; if swap, then request `SetVolume(new handle,0)` and `Schedule(new handle,255,250,0)`, even when the returned handle is zero. Store selected ambience as retained ambience. If selected ambience is zero, instead clear retained ambience and ambience handle without a new sound request. Unchanged ambience skips this entire step.
3. If current weather identity is zero, request `Stop(retained weather handle)` even for handle zero, then clear weather handle and retained weather identity. Otherwise: if target amplitude is nonzero and weather handle is zero, request sound using the **retained weather identity**, store its returned handle, and if nonzero request `SetVolume(handle,target amplitude)`. With any nonzero weather handle now present, read its current volume `v`. If target > v, set candidate=`low32(v+6)` and replace candidate with target when signed target < candidate. Otherwise set candidate=`low32(v-6)` and replace candidate with target when signed target > candidate. If candidate is nonzero, request `SetVolume(handle,candidate)`; otherwise request `Stop(handle)` and clear weather handle. Finally store current weather identity as retained weather identity, including after a failed start request. A successful new start can therefore produce two volume-set requests in this step.
4. Request `UpdateAcoustics(environment identity)`. If selected event is zero, skip to step 6. Otherwise execute SAW-R02's changed/unchanged timer stage using the seed as it stands after preceding operations. Compute elapsed=`low32(tick-lastTick)`. If unsigned elapsed < interval, skip to step 6.
5. Request `RequestSound(selected event)`. A zero returned handle skips all following position/timer draws and writes. On a nonzero handle, draw `a=Rand(seed,301)`, then `b=Rand(seed,2)`, then `c=Rand(seed,201)`. Set distance=`450+a`, sign=-1 if b=0 else +1; request `SetPosition(handle,sign*distance,c-100,0)`. After the position operation returns, write lastTick=tick, draw `d=Rand(seed,low32(2*floor(D/3)+1))` with signed modulus interpretation, and write interval=`low32(D+d-floor(D/3))`. Retained previous event is unchanged by this step. No initial backdating draw follows successful due-event playback.
6. If initial level identity differs from retained last level, request `ClearAuxiliaryEligibility()`, then `EnableAuxiliary(A)` if A is nonzero, then `EnableAuxiliary(B)` if B is nonzero, even if A equals B. Store initial level identity as retained last level. Otherwise make no auxiliary request.

Only step 3's target-amplitude comparison uses the computed weather amplitude; it does not replace the authored/retained cue identity used for the weather start request. Callback execution may affect the seed according to its supplied contract; this rule establishes the caller's ordering and direct draws only.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| SAW-R01 | Supplied environment pair | verified in 1.13c code (static) | none locally; producer integration SAW-Q01 | SAW-V01 |
| SAW-R02 | Direct initial timer draws/writes | verified in 1.13c code (static) | none locally; preceding calls and later event execution SAW-Q03 | SAW-V02, SAW-V03 |
| SAW-R03 | Named text cells | verified in 1.13c data | none | T01 |
| SAW-R04a | Supplied weather predicate/intensity | verified in 1.13c code (static) | none locally; prepared-record and intensity-producer integration SAW-Q01/SAW-Q02 | SAW-V04 |
| SAW-R04b | Audible transitions and weather/event execution | unresolved | SAW-Q02, SAW-Q03: full playback/weather integration | none |
| SAW-R05 | Full eligible supplied-operation caller | verified in 1.13c code (static) | SAW-Q01/Q02/Q03 gate real producers and nested playback/position/acoustic/auxiliary effects | SAW-V05–SAW-V09 |

## 5. Outputs and side effects

| Output | Rule |
|---|---|
| Selected ambience and event identities | SAW-R01 |
| Scheduler interval, last tick, previous event; direct seed updates | SAW-R02 |
| Returned weather eligibility and selected intensity | SAW-R04a |
| Audible ambience, weather and event playback | SAW-R04b; gated |
| Ordered audio-operation request log; retained ambience/weather identities and handles; last-level identity; direct due-event seed/interval/last-tick updates | SAW-R05 |

SAW-R03 is data-only and performs no runtime write.

## 6. Branches and edge cases

SAW-R01 covers phases outside 1..3 without guessing day/night clock boundaries. SAW-R02 explicitly handles zero/unchanged events, zero delay and nonpositive signed draw moduli; all intermediate arithmetic wraps. Missing player, room, level or environment and malformed loader records remain SAW-Q01. SAW-R05 verifies unsigned due comparisons, supplied failed-request behavior and direct post-success draws; real request execution remains SAW-Q03. No difficulty or PvP exception is established.

## 7. Open questions

- **SAW-Q01** [unresolved] Establish missing-record behavior, compiled/text correspondence and update cadence. The phase accessor is bound to the current act world-clock cycle and the level byte to Levels.SoundEnv; full data/runtime integration remains open. Supplied-input SAW-R01 and SAW-R02 remain available.
- **SAW-Q02** [unresolved] Complete the stored-intensity producer, its update scheduling, invalid/out-of-range amplitude conversion, the named data binding of fixed cue 64 and outside/inside semantics. SAW-R04a establishes the supplied-input Rain predicate/intensity selection; SAW-R05 establishes valid-range conversion and the retained-identity request rule, without those producer/audible effects.
- **SAW-Q03** [unresolved] Complete nested playback, cleanup, stop, position, acoustic and auxiliary effects, including any dependency-owned RNG. SAW-R05 establishes the full eligible caller's request order, supplied success/failure branches and direct draws; these remain available without claiming audible integration.
- **SAW-Q04** [unresolved] Gameplay observations are pending for SAW-R01, SAW-R02, SAW-R04a and SAW-R05. No attributable complete input/output capture is available for their vectors.

## 8. Test vectors

### 8.1 Executable vectors

All vectors operate at supplied-input stages, with no actual backend playback. The timer stage persists its scheduler and seed writes within each case. Each case starts independently.

Full-caller fixtures SAW-V05–SAW-V09 use these defaults unless overridden: phase=2; environment identity=2, Day Ambience=11, Night Ambience=22, Day Event=0, Night Event=0, Event Delay=250, auxiliary A=0/B=0; current/retained last level=10/10; retained ambience=11 with handle=5; weather eligibility=false, first/second intensity=0/0, retained weather identity=64 with handle=0; previous event=9, interval=274, lastTick=997, tick=1000. Sound requests return zero unless the vector specifies another result. Every callback logs its name/arguments and has no seed mutation in these fixtures. A nonzero returned sound handle has a supplied amplitude entry initialized to 255; SetVolume persists to that entry, and ReadVolume returns the stored entry (zero for a missing/zero handle). Schedule logs only its caller request here; stop/cleanup/position/acoustic/auxiliary callbacks log only. No callback changes controller identities, handles, clock or scheduler fields. All vector assertions concern the listed local state and operation logs; no waveforms or hidden voice mutations are asserted. Each listed `RequestSound` log entry expands to absent emitter plus delay=0, request flags=0 and opaque playback value=0.

- **SAW-V01** Pair selection (SAW-R01; boundary: supplied-input stage)
  - Inputs: environment Day Ambience 11, Night Ambience 22, Day Event 33, Night Event 44; phase 3.
  - Seeds: no draw.
  - Draws: none.
  - Trace: 3 lies in inclusive interval 1..3; copy first pair.
  - Outputs: selected ambience 11, selected event 33.
  - Discriminates: phase 3 excluded from first pair would incorrectly select 22/44.
  - runtime_observed: no.
  - (a) Inputs: same environment, phase 4. Draws: none. Outputs: selected ambience 22, selected event 44.
  - (b) Inputs: same environment, phase 0. Draws: none. Outputs: selected ambience 22, selected event 44.
  - (c) Inputs: same environment, phase 1. Draws: none. Outputs: selected ambience 11, selected event 33.
  - (d) Inputs: environment Day Ambience 11, Night Ambience 22, Day Event 33, Night Event 44; phase 2. Seeds: no draw. Draws: none. Trace: 2 lies in inclusive interval 1..3; copy first pair. Outputs: selected ambience 11, selected event 33. Discriminates: selecting the first pair only for phases 1 and 3 incorrectly produces 22/44 here. runtime_observed: no.

- **SAW-V02** Initialize changed event timer (SAW-R02; boundary: supplied-input stage)
  - Inputs: selected event 33, previous event 44, D 250, tick 1000, old interval 999, old last tick 888.
  - Seeds: local client player {lo: 1, hi: 0}.
  - Draws: (local client player, 167, 107), then (local client player, 274, 3).
  - Trace: q=83; m=167; first step low/high=1791398085/0; interval=250+107-83=274. Second step low/high=1721382809/747178471; lastTick=1000-3=997.
  - Outputs: interval 274, lastTick 997, previousEvent 33, seed {lo: 1721382809, hi: 747178471}.
  - Discriminates: using nominal delay 250 for the second modulus or omitting random initial backdating changes the draw or lastTick.
  - runtime_observed: no.

- **SAW-V03** Retain unchanged timer (SAW-R02; boundary: supplied-input stage)
  - Inputs: selected event 33, previous event 33, D 250, tick 1000, interval 274, lastTick 997.
  - Seeds: local client player {lo: 1, hi: 0}.
  - Draws: none.
  - Trace: identities equal; retain timer without stepping seed.
  - Outputs: interval 274, lastTick 997, previousEvent 33, seed {lo: 1, hi: 0}.
  - Discriminates: restarting the timer on every update changes the seed and lastTick.
  - runtime_observed: no.

- **SAW-V04** Weather eligibility and selected intensity (SAW-R04a; boundary: supplied accessor stages)
  - Inputs: suppression 0, local player present true, Rain 2, Mud 0, stored intensity 0.5.
  - Seeds: no draw.
  - Draws: none.
  - Trace: suppression is zero and player exists; nonzero Rain admits weather and selects stored intensity.
  - Outputs: eligibility true, intensity 0.5; no persistent writes.
  - (a) Fresh inputs: suppression 0, player present true, Rain 0, Mud 1, stored intensity 0.5. Outputs: eligibility false, intensity positive zero; no writes or draws.
  - (b) Fresh inputs: suppression -1, player present true, Rain 2, Mud 1, stored intensity 0.5. Outputs: eligibility false, intensity positive zero; no writes or draws.
  - (c) Fresh inputs: suppression 0, player present false, Rain 2, Mud 1, stored intensity 0.5. Outputs: eligibility false, intensity positive zero; no writes or draws.
  - (d) Fresh inputs: suppression 0, player present true, Rain 2, Mud 1, stored intensity -0.25. Outputs: eligibility true, intensity -0.25; no writes or draws.
  - Discriminates: testing Rain equal to 1 fails the base case; using Mud for eligibility fails (a); ignoring suppression or absence fails (b)/(c); clamping intensity fails (d).
  - runtime_observed: no.

- **SAW-V05** Ordinary ambience replacement and final auxiliary requests (SAW-R01, SAW-R05; boundary: supplied-operation full eligible caller)
  - Inputs: full-caller defaults; retained ambience=33, ambience handle=5; retained last level=8; auxiliary A=101/B=0; RequestSound(11) returns handle 7.
  - Seeds: local client player {lo:1,hi:0}.
  - Draws: none.
  - Trace: selected 11 differs from 33 without a Day/Night swap; cleanup requests precede new request. Disabled weather requests stop(0), then acoustic update. Event zero skips timers; changed level clears auxiliary eligibility then enables 101.
  - Outputs: ambience identity/handle=11/7; weather identity/handle=0/0; previous event=9, interval=274, lastTick=997; retained last level=10; seed={lo:1,hi:0}. Ordered log=[KeepAmbient(11,0),KeepEvent(0,0),RequestSound(11),Stop(0),UpdateAcoustics(2),ClearAuxiliaryEligibility(),EnableAuxiliary(101)].
  - Discriminates: starting ambience before cleanup, omitting zero-handle Stop, or doing auxiliary requests before the acoustic update fails the ordered log; treating every change as a swap emits schedules instead of cleanup.
  - runtime_observed: no.
- **SAW-V06** Failed crossfade request still records selection (SAW-R01, SAW-R05; boundary: supplied-operation full eligible caller)
  - Inputs: full-caller defaults; retained ambience=22, ambience handle=5; RequestSound(11) returns 0.
  - Seeds: local client player {lo:1,hi:0}.
  - Draws: none.
  - Trace: Day/Night swap schedules old handle down; failed new request stores handle zero and still issues zero-handle volume/schedule requests before retaining identity 11.
  - Outputs: ambience identity/handle=11/0; weather identity/handle=0/0; previous event=9, interval=274, lastTick=997; retained last level=10; seed={lo:1,hi:0}. Ordered log=[Schedule(5,0,250,0),RequestSound(11),SetVolume(0,0),Schedule(0,255,250,0),Stop(0),UpdateAcoustics(2)].
  - Discriminates: retrying failed requests by retaining identity 22, omitting the caller's zero-handle requests, or using ordinary cleanup changes state/log.
  - runtime_observed: no.
- **SAW-V07** Weather starts with retained cue and truncates amplitude (SAW-R01, SAW-R05; boundary: supplied-operation full eligible caller)
  - Inputs: full-caller defaults; weather eligibility=true, first/second intensity=0.5/0.5; retained weather identity=0, weather handle=0; RequestSound(0) returns 0.
  - Seeds: local client player {lo:1,hi:0}.
  - Draws: none.
  - Trace: selected weather cue=64; target=trunc(0.5*255)=127. Start requests entering cue 0 and fails; retained weather identity becomes 64 for a later update.
  - Outputs: ambience identity/handle=11/5; weather identity/handle=64/0; previous event=9, interval=274, lastTick=997; retained last level=10; seed={lo:1,hi:0}. Ordered log=[RequestSound(0),UpdateAcoustics(2)].
  - Discriminates: using current cue 64 immediately requests the wrong cue; case (a)'s amplitude output 127 rejects rounding 127.5 to 128.
  - runtime_observed: no.
  - (a) Fresh Inputs: full-caller defaults; weather eligibility=true, first/second intensity=0.5/0.5; retained weather identity=64, weather handle=0; RequestSound(64) returns 8. Seeds: {lo:1,hi:0}. Draws: none. Trace: set new handle to 127, then ReadVolume returns the persisted 127; smoothing selects 127 and writes it a second time. Outputs: ambience identity/handle=11/5; weather identity/handle=64/8; amplitude(8)=127; previous event=9, interval=274, lastTick=997; retained last level=10; seed={lo:1,hi:0}. Ordered log=[RequestSound(64),SetVolume(8,127),ReadVolume(8)->127,SetVolume(8,127),UpdateAcoustics(2)]. Discriminates: skipping the post-start read/smoothing pass removes the second write; an incorrect nonpersisting setter changes the smoothing result. runtime_observed: no.
- **SAW-V08** Due-event success at wrapped equality (SAW-R01, SAW-R02, SAW-R05; boundary: supplied-operation full eligible caller)
  - Inputs: full-caller defaults; Day Event=33; previous event=33; interval=32; lastTick=4294967280; tick=16; RequestSound(33) returns 9.
  - Seeds: local client player {lo:1,hi:0}.
  - Draws: (local client player,301,197), (local client player,2,1), (local client player,201,153), (local client player,167,35).
  - Trace: unchanged event skips initial draws; low32(16-4294967280)=32 is due at equality. Position distance=450+197=647, sign=+1, vertical=153-100=53. After SetPosition returns write lastTick=16; q=83, next interval=250+35-83=202. Ordered seed states={1791398085,0},{1721382809,747178471},{986833572,717975634},{2963478662,411600752}.
  - Outputs: ambience identity/handle=11/5; weather identity/handle=0/0; previous event=33, interval=202, lastTick=16; retained last level=10; seed={lo:2963478662,hi:411600752}. Ordered log=[Stop(0),UpdateAcoustics(2),RequestSound(33),SetPosition(9,647,53,0)].
  - Discriminates: strict-greater due check skips the event; unwrapped subtraction misclassifies due time; reversing sign/vertical draws changes positions; initial backdating after success adds an unwanted fifth direct draw.
  - runtime_observed: no.
  - (a) Fresh Inputs: full-caller defaults; Day Event=33; previous event=33; interval=32; lastTick=4294967280; tick=16; RequestSound(33) returns 9. Seeds: local client player {lo:0,hi:0}. Draws: (local client player,301,0),(local client player,2,0),(local client player,201,0),(local client player,167,0). Trace: fixed-point zero seed remains zero through all four steps; distance=450, sign=-1, vertical=-100; next interval=250-83=167. Outputs: ambience identity/handle=11/5; weather identity/handle=0/0; previous event=33, interval=167, lastTick=16; retained last level=10; seed={lo:0,hi:0}; ordered log=[Stop(0),UpdateAcoustics(2),RequestSound(33),SetPosition(9,-450,-100,0)]. Discriminates: zero sign mapped to positive produces +450 instead of -450. runtime_observed: no.
- **SAW-V09** Failed due-event request preserves scheduler (SAW-R01, SAW-R02, SAW-R05; boundary: supplied-operation full eligible caller)
  - Inputs: full-caller defaults; Day Event=33; previous event=33; interval=32; lastTick=4294967280; tick=16; RequestSound(33) returns 0.
  - Seeds: local client player {lo:1,hi:0}.
  - Draws: none.
  - Trace: wrapped elapsed 32 is due; failed request bypasses position draws, position request, timestamp write and interval redraw.
  - Outputs: ambience identity/handle=11/5; weather identity/handle=0/0; previous event=33, interval=32, lastTick=4294967280; retained last level=10; seed={lo:1,hi:0}. Ordered log=[Stop(0),UpdateAcoustics(2),RequestSound(33)].
  - Discriminates: consuming random position draws before checking the returned handle changes seed; resetting the timer on failure changes lastTick or interval.
  - runtime_observed: no.

### 8.2 Observation vectors

- **SAW-O01** Capture an eligible environment update and changed-event initialization; compare the supplied phase, selected pair and direct timer inputs/outputs against SAW-V01/SAW-V02. [Procedure](../research/spec-evidence/sound-ambient-and-weather.md#observation-procedure). No match is currently recorded.
- **SAW-O02** Capture one complete eligible controller update with all supplied operation results, retained state and seed changes; compare the full caller request sequence and direct updates against SAW-R05. [Requirements](../research/spec-evidence/sound-ambient-and-weather.md#full-caller-observation-requirements). No complete match is recorded.

## 9. Acceptance and dependencies

SAW-V01 through SAW-V09 establish SAW-R01, SAW-R02, SAW-R04a and SAW-R05 at their supplied-input stages; SAW-R03 is T01. Full reproduction additionally requires resolving SAW-R04b, producer integration SAW-Q01, and attributable observations. Partial acceptance does not establish complete weather execution or ambient sound parity.

**Depends on:** NRF-R01 and NRF-R02 in [numeric foundations](core-numeric-and-rng-foundations.md), for the numerical step/modulus operation only. Seed ownership here is client-specific. [World clock](drlg-world-clock.md) WCK-R01 through WCK-R03 own the cycle consumed as phase, preserving their evidence labels and presentation gates. GDA-R03a/GDA-R03b in [game-data table records](core-game-data-table-records.md) support descriptor identity/direct conversion; GDA-R02/GDA-R03d full compilation remains gated. STP-R05 in [playback](sound-table-and-playback-rules.md) owns scheduling requests against supplied voices, while STP-R03b/STP-R04b request/device integration remains unresolved. S3D-R04 in [positional sound](sound-3d-position-and-falloff.md) owns supplied integer-position cache updates; full position request and playback execution remain gated. SAW-R05's request log is available without asserting those nested effects.

**Depended on by:** no established consumer.

## 10. Related specs

[Sound table and playback rules](sound-table-and-playback-rules.md) owns playback execution. [Positional sound](sound-3d-position-and-falloff.md) owns final placement/falloff. Music selection consumes the same environment independently.

## Appendix A. Provenance

Original Expansion 1.13c Windows x86; [manifest](../research/source-manifest.md) identifies D2Client and D2Common. Read date 2026-10-29. Static disassembly, available matched reconstruction and original-image instruction listing support the verified supplied stages. Named data cells reuse the same-date owned vanilla loader record with its original limits. Runtime vectors remain unobserved; the [companion](../research/spec-evidence/sound-ambient-and-weather.md) gives source mapping and pending gates.
