# Sound table and playback rules

**Description**

Sound playback connects a requested game sound to an audio voice and updates that voice over time. A sound group can exclude recent choices by drawing again. This contract establishes that selection stage and volume transition scheduling/calculation; complete duplicate handling and device playback remain evidence gates.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md)
Ids: STP-R, STP-Q, STP-V, STP-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/sound-table-and-playback-rules.md)
Open evidence gaps: [gap file](../plans/gaps/sound-table-and-playback-rules.md)

## 1. Scope

Sound-record interpretation and the client playback lifecycle: group choice, volume, looping, fading, duplicate deferral, stopping and voice priority. Music selection, world emitter selection and positional attenuation belong to their separate sound contracts. Group choice, the supplied history commit, valid volume scheduling/calculation stages and named archive cells below have verified assertions; the full lifecycle is gated.

## 2. Trigger and ownership

D2Client owns sound request coordination and per-update voice state. Its sound update dispatch reaches the volume-transition stage before voice eligibility and playback/update operations. D2Sound owns the eventual device operations. Full request-to-device integration, wider player-context lifecycle and clock-unit integration remain unresolved under STP-Q01 and STP-Q02. The supplied-input stages make no claim about server frame cadence.

## 3. Inputs and data

| Input | Semantic source | Width and format | Meaning | Origin | Binding |
|---|---|---|---|---|---|
| transition active | voice update state | Boolean | whether to evaluate a pending transition | runtime input | supplied |
| current tick, start tick, end tick | client audio update clock and voice state | unsigned 32-bit tick values | unsigned comparisons select the transition branch | runtime input | supplied |
| start volume, end volume, current volume | voice transition state | signed 32-bit integers | interpolation endpoints and stored result | runtime input | supplied |
| base cue, sound table present | group selection request | signed 32-bit cue; Boolean | base is a valid record index when the table is present | runtime input | supplied |
| Group Size | supplied base Sounds record | unsigned 8-bit | group member count; candidates are contiguous from base cue | table data | descriptor-associated field; supplied decoded value |
| most recent choice, older choice | base record's retained history | two 32-bit cue ids | ordered repeat-exclusion entries | engine state | supplied |
| sound-selection seed | current client player-context seed | two unsigned 32-bit words | advanced by every group-selection draw; see NRF-R01/NRF-R02 | runtime state | supplied seed; wider context lifecycle unresolved |
| selected cue | supplied history-commit input | 32-bit cue id | chosen cue written after early suppression passes | runtime input | supplied |
| ordered voices and requested handle | scheduling request and current voice state | ordered logical records; unsigned 32-bit handle | first matching record is selected; zero handle skips | runtime input | supplied |
| target volume, requested duration, delay | scheduling request | signed 32-bit target; unsigned 32-bit duration/delay | requested endpoint, interval length and start delay | runtime input | supplied |
| emitter association list | selected voice | ordered associations | a second entry suppresses scheduling | runtime state | supplied |
| retirement-request condition | selected voice | Boolean | target zero sets it; later lifecycle meaning unresolved | runtime state | supplied |
| Fade In, Fade Out | supplied selected Sounds record | unsigned 8-bit values | minimum schedule duration for target 255 or 0 | table data | descriptor-associated fields; supplied decoded values |
| valid buffer context and buffer-start marker | supplied device wrapper state | context reference and signed 32-bit marker | marker is written on successful buffered start | runtime input | supplied |
| device critical-section operations and repeating-play operation | supplied external device boundary | ordered operations; signed 32-bit device result | local wrapper logs calls and branches on result sign | external operation | supplied results; real device execution gated |
| cursor_pass Volume | effective data/global/excel/sounds.txt, Sound=cursor_pass, Volume | decimal text, raw 255 | named archive cell only; consumer scale is gated | table data | unresolved consumer |
| cursor_pass Group Size | same row, Group Size | decimal text, raw 0 | named archive cell only | table data | unresolved consumer |
| cursor_pass Stop Inst | same row, Stop Inst | decimal text, raw 1 | named archive cell only | table data | unresolved consumer |

Archive cells are supported by [T01](../research/spec-evidence/sound-table-and-playback-rules.md#t01-effective-sound-cells). They do not establish runtime compiled-table correspondence. Arithmetic vectors use supplied transition state without assigning a table column to an endpoint or duration.

## 4. Rules (execution order)

**STP-R01** Named sound cells [verified in 1.13c data]. The three named cursor_pass cells in section 3 have the stated raw values. No playback meaning follows solely from their names.

**STP-R02** Pending volume transition [verified in 1.13c code (static)]. At the supplied-input transition stage:

1. If transition active is false, leave current volume unchanged.
2. If unsigned current tick is greater than unsigned end tick, set current volume to end volume and clear transition active.
3. Otherwise, if unsigned current tick is less than unsigned start tick, set current volume to start volume and retain transition active.
4. Otherwise set current volume to `start volume + trunc_toward_zero(((end volume - start volume) * (current tick - start tick)) / (end tick - start tick))`, retaining transition active. Subtraction, multiplication and final addition retain their low 32 bits; division interprets the resulting numerator and denominator as signed 32-bit values.

The valid supplied-input interpolation boundary requires a nonzero signed denominator and a representable signed quotient. A zero denominator or signed division overflow has no normal volume output in this contract. Clock wrap is not normalized into a relative interval. No random draw occurs.

**STP-R03** Group selection and playback admission.

- **STP-R03a** Group selection [verified in 1.13c code (static)]. With a supplied valid base sound record and stable table/seed inputs, return the base cue unchanged and make no draw if the sound table is absent or Group Size is 0 or 1. Otherwise let exclusion count be `min(Group Size-1,2)`. For Group Size 2 or 3 only, draw `Rand(sound-selection seed,Group Size+1)` using NRF-R01/NRF-R02; decrement exclusion count if that result is zero. Repeatedly draw `Rand(sound-selection seed,Group Size)`, form `low32(base cue + draw)`, and reject that candidate if it equals any of the first exclusion-count entries of `[most recent choice, older choice]`. Return the first candidate not rejected. Every attempted draw advances the same seed, including rejected candidates. Selection does not modify history. There is no draw-count cap or fallback candidate.
- **STP-R03b** Admission, duplicate and voice priority [unresolved]. A selected sound may be rejected before device playback. Complete duplicate stopping/deferral and voice-priority policy remain STP-Q01. STP-R03a does not establish admission or audible output.
- **STP-R03c** History commit stage [verified in 1.13c code (static)]. At the admission stage after its early recent-voice suppression test has passed, set the original base record's older choice to its entering most recent choice, then most recent choice to the supplied selected cue. No draw occurs. These writes precede asset loading and later slot/device admission; a later admission failure does not roll them back within this path. Eligibility for reaching this supplied stage remains STP-Q01.

**STP-R04** Device playback, loop and fade integration.

- **STP-R04a** Buffered start wrapper [verified in 1.13c code (static)]. With a supplied valid buffer context and returning device operations, enter the audio device critical section, request repeating playback from the underlying buffer, and examine the signed 32-bit device result. A negative result leaves the buffer-start marker unchanged, exits the critical section and returns 0. A nonnegative result sets the buffer-start marker to 1, exits the critical section and returns 1. No RNG draw occurs. Critical-section and device execution are supplied boundary operations; this wrapper does not verify their internal effects. It does not branch on the sound-table Loop field.
- **STP-R04b** Complete device and retirement integration [unresolved]. Audio voices pass from client coordination to device playback. Streamed starts, clock units, sound-table Loop lifecycle, volume scaling and voice retirement remain STP-Q02. Repeating buffered playback in STP-R04a does not establish that a non-looping logical voice continues indefinitely; later client stopping remains gated.

**STP-R05** Schedule a volume transition [verified in 1.13c code (static)]. Inputs are an ordered voice list, requested handle, target volume, requested duration, delay and current tick. All tick/duration arithmetic retains low32; timing comparisons are unsigned. Target/current volumes are signed 32-bit values.

1. If handle is zero or no voice has that handle, return without writes. Choose the first matching voice. If its emitter association list has a second entry, return without writes.
2. Starting from requested duration, if target is 255 raise duration to at least unsigned Fade In; if target is 0 raise it to at least unsigned Fade Out. Other targets have no table duration minimum.
3. If a transition is active and its end volume equals target, compute candidate end=`low32(current tick+delay+duration)`; if candidate end is unsigned greater than or equal to the existing end tick, return without writes. Otherwise continue. If an active transition's end volume is 0 and the requested target is nonzero, return without writes.
4. If target is zero, set the voice's retirement-request condition true. Other targets do not clear this condition.
5. When duration is zero, this supplied-input boundary requires delay zero. Set current volume to target and return; do not clear or overwrite existing transition state. The invalid zero-duration/nonzero-delay pair is outside this valid scheduling boundary under STP-Q02.
6. With positive duration, set transition start tick=`low32(current tick+delay)`, end tick=`low32(start tick+duration)`, start volume to entering current volume, end volume to target and active=true. Current volume is not changed by scheduling.

No draw occurs. The valid supplied boundary requires a valid associated sound record whenever a matched voice reaches the duration policy. This rule schedules state consumed by STP-R02; device volume application and the interpretation of retirement requests remain STP-Q02.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| STP-R01 | Named archive cells | verified in 1.13c data | none | T01 |
| STP-R02 | Supplied transition arithmetic | verified in 1.13c code (static) | none for valid supplied inputs | STP-V01–STP-V07 |
| STP-R03a | Group choice with repeat rejection | verified in 1.13c code (static) | none for supplied valid group/seed/history | STP-V08–STP-V11 |
| STP-R03b | Admission lifecycle | unresolved | STP-Q01: admission and duplicate/priority results | none |
| STP-R03c | Supplied history commit | verified in 1.13c code (static) | STP-Q01 gates reaching stage only | STP-V12 |
| STP-R04a | Supplied buffered start result | verified in 1.13c code (static) | STP-Q02 gates real device and logical voice lifetime | STP-V18, STP-V19 |
| STP-R04b | Complete device/lifecycle integration | unresolved | STP-Q02: stream/device/lifetime outcomes | none |
| STP-R05 | Transition scheduling | verified in 1.13c code (static) | STP-Q02 gates device effects and retirement integration | STP-V13–STP-V17 |

## 5. Outputs and side effects

| Output | Rule |
|---|---|
| current volume assigned on active transition branches | STP-R02 |
| transition active cleared only after current tick exceeds end tick | STP-R02 |
| returned group cue and updated sound-selection seed; history unchanged | STP-R03a |
| most recent and older choice committed in the original base record | STP-R03c |
| scheduled start/end ticks, endpoint volumes, active condition; immediate current-volume write; retirement-request condition | STP-R05 |
| critical-section and repeating-device-play request log; buffer-start marker set on nonnegative result; Boolean wrapper result | STP-R04a |

Actual playback starts/stops and other lifecycle writes remain gated under STP-R03b and STP-R04b; STP-R04a establishes the local wrapper's supplied-operation requests and result handling only.

## 6. Branches and edge cases

End-tick equality retains an active transition; the next greater tick clears it. Interpolation decreases as well as increases volume, truncating negative division toward zero. Inactive transitions do not overwrite volume. Zero-duration interpolation, overflow, missing records, unavailable sound devices and invalid sound identifiers do not acquire guessed fallback behavior.

## 7. Open questions

**STP-Q01** [unresolved] Complete admission, duplicate and voice-priority effects and the wider player-context seed producer/lifetime. This does not gate STP-R03a with supplied stable group/history/seed inputs or STP-R03c after its supplied early-suppression boundary.

**STP-Q02** [unresolved] Establish clock-unit integration, device calls, looping, volume conversion and failure/retirement paths, including invalid zero-duration/nonzero-delay scheduling. This gates full playback reproduction, not STP-R02/STP-R05 with supplied valid inputs.

**STP-Q03** [unresolved] Confirm STP-R02, STP-R03a, STP-R03c and STP-R05 with complete measured inputs and output on the original game. This is an observation gap only. See the [attempts and closure conditions](../plans/gaps/sound-table-and-playback-rules.md).

## 8. Test vectors

### 8.1 Executable vectors

Each case has independent state and no omitted persistent operations. STP-V01–STP-V07 and STP-V12–STP-V19 assert exactly zero random draws and an empty draw log. Group vectors log every draw and preserve the final seed. History is supplied in most-recent/older order. Scheduler vectors start from the explicitly listed voice; fields omitted by a branch remain unchanged. Device-wrapper vectors log every supplied critical-section/device operation; those operations have no other stored effects beyond returning the stated device result.

- **STP-V01** Inactive transition (STP-R02; boundary: supplied-input stage)
  - Inputs: active=false; current tick=15; start tick=10; end tick=20; start volume=0; end volume=255; current volume=73.
  - Seeds: no draw.
  - Draws: none.
  - Trace: inactive branch leaves volume untouched.
  - Outputs: current volume=73; active=false.
  - Discriminates: unconditional endpoint assignment would replace 73.
  - runtime_observed: no
- **STP-V02** Before transition (STP-R02; boundary: supplied-input stage)
  - Inputs: active=true; current tick=9; start tick=10; end tick=20; start volume=10; end volume=110; current volume=73.
  - Seeds: no draw.
  - Draws: none.
  - Trace: 9 is not greater than 20 and is less than 10.
  - Outputs: current volume=10; active=true.
  - Discriminates: extrapolation would produce 0.
  - runtime_observed: no
- **STP-V03** Falling interpolation (STP-R02; boundary: supplied-input stage)
  - Inputs: active=true; current tick=11; start tick=10; end tick=13; start volume=10; end volume=0; current volume=73.
  - Seeds: no draw.
  - Draws: none.
  - Trace: difference=-10; elapsed=1; duration=3; quotient=-3; volume=10-3=7.
  - Outputs: current volume=7; active=true.
  - Discriminates: floor division would produce 6.
  - runtime_observed: no
- **STP-V04** Exact end (STP-R02; boundary: supplied-input stage)
  - Inputs: active=true; current tick=20; start tick=10; end tick=20; start volume=0; end volume=255; current volume=73.
  - Seeds: no draw.
  - Draws: none.
  - Trace: 255*10/10+0=255; strict greater-than test is false.
  - Outputs: current volume=255; active=true.
  - Discriminates: greater-or-equal completion incorrectly clears active.
  - runtime_observed: no
- **STP-V05** After end (STP-R02; boundary: supplied-input stage)
  - Inputs: active=true; current tick=21; start tick=10; end tick=20; start volume=0; end volume=255; current volume=73.
  - Seeds: no draw.
  - Draws: none.
  - Trace: unsigned 21>20 selects completion before interpolation.
  - Outputs: current volume=255; active=false.
  - Discriminates: unbounded interpolation would produce 280.
  - runtime_observed: no

- **STP-V06** High-bit tick uses unsigned completion (STP-R02; boundary: supplied-input stage)
  - Inputs: active=true; current tick=2147483648; start tick=0; end tick=20; start volume=0; end volume=255; current volume=73.
  - Seeds: no draw.
  - Draws: none.
  - Trace: unsigned 2147483648>20 selects completion before interpolation.
  - Outputs: current volume=255; active=false.
  - Discriminates: signed current tick comparison treats the high-bit tick as negative and selects before-start volume 0, retaining active.
  - runtime_observed: no
- **STP-V07** Wrapped product uses signed division (STP-R02; boundary: supplied-input stage)
  - Inputs: active=true; current tick=32768; start tick=0; end tick=65536; start volume=0; end volume=65536; current volume=73.
  - Seeds: no draw.
  - Draws: none.
  - Trace: 65536*32768 retains low32=2147483648, interpreted signed as -2147483648; denominator=65536; signed quotient=-32768; final volume=-32768.
  - Outputs: current volume=-32768; active=true.
  - Discriminates: a widened product or unsigned division produces +32768 instead of -32768.
  - runtime_observed: no

- **STP-V08** Singleton consumes no draw (STP-R03a; boundary: supplied-input stage)
  - Inputs: table present=true; base cue=100; Group Size=1; history=[101,102].
  - Seeds: sound-selection seed={lo=0,hi=1}.
  - Draws: none.
  - Trace: Group Size <=1 returns the base without stepping the seed.
  - Outputs: cue=100; seed={lo=0,hi=1}; history=[101,102].
  - Discriminates: drawing modulus 1 would change the seed to {1,0}.
  - runtime_observed: no
- **STP-V09** Rejected group choices still draw (STP-R03a; boundary: supplied-input stage)
  - Inputs: table present=true; base cue=100; Group Size=2; history=[101,100].
  - Seeds: sound-selection seed={lo=0,hi=1}.
  - Draws: (sound-selection seed,3,1), (sound-selection seed,2,1), (sound-selection seed,2,1), (sound-selection seed,2,0).
  - Trace: exclusion count starts at 1 and stays 1 after preliminary result 1. Candidates 101,101 are rejected; candidate 100 is accepted. Ordered seed states are {1,0}, {1791398085,0}, {1721382809,747178471}, {986833572,717975634}.
  - Outputs: cue=100; seed={lo=986833572,hi=717975634}; history=[101,100].
  - Discriminates: omitting the preliminary draw or returning the first member fails the exact draw log; committing selection inside this stage changes history incorrectly.
  - runtime_observed: no
- **STP-V10** Small-group exclusion reduction (STP-R03a; boundary: supplied-input stage)
  - Inputs: table present=true; base cue=100; Group Size=2; history=[100,101].
  - Seeds: sound-selection seed={lo=0,hi=0}.
  - Draws: (sound-selection seed,3,0), (sound-selection seed,2,0).
  - Trace: preliminary zero reduces exclusion count from 1 to 0. The fixed-point seed remains zero; candidate 100 is accepted despite being most recent.
  - Outputs: cue=100; seed={lo=0,hi=0}; history=[100,101].
  - Discriminates: always excluding the most recent choice cannot return on these inputs; excluding no history without the preliminary draw fails the two-draw log.
  - runtime_observed: no
- **STP-V11** Larger group has no preliminary draw (STP-R03a; boundary: supplied-input stage)
  - Inputs: table present=true; base cue=100; Group Size=4; history=[101,102].
  - Seeds: sound-selection seed={lo=0,hi=1}.
  - Draws: (sound-selection seed,4,1), (sound-selection seed,4,1), (sound-selection seed,4,1), (sound-selection seed,4,0).
  - Trace: exclusion count=2; candidates 101,101,101 are rejected; 100 is accepted. Seed states are {1,0}, {1791398085,0}, {1721382809,747178471}, {986833572,717975634}.
  - Outputs: cue=100; seed={lo=986833572,hi=717975634}; history=[101,102].
  - Discriminates: applying a Group Size+1 preliminary draw adds a modulus-5 entry that the required draw log rejects.
  - runtime_observed: no
- **STP-V12** Commit uses entering recent value (STP-R03c; boundary: supplied-input stage)
  - Inputs: early recent-voice suppression has passed; original base cue=100; history=[101,102]; selected cue=103.
  - Seeds: no draw.
  - Draws: none.
  - Trace: older=entering recent=101; recent=selected=103.
  - Outputs: original base history=[103,101].
  - Discriminates: writing recent before reading it would make older=103; writing the selected record instead leaves base history unchanged.
  - runtime_observed: no
- **STP-V13** Fade-out duration minimum and delay (STP-R05; boundary: supplied-input stage)
  - Inputs: one voice, handle=7, emitter associations=[], current volume=200, active=false, retirement=false, transition fields start tick=0/end tick=0/start volume=0/end volume=0; request handle=7,target=0,duration=2,delay=3,current tick=100; Fade In=9,Fade Out=5.
  - Seeds: no draw.
  - Draws: none.
  - Trace: effective duration=max(2,5)=5; retirement=true; start=103; end=108; start volume=200; end volume=0.
  - Outputs: current volume=200; active=true; retirement=true; transition fields start tick=103/end tick=108/start volume=200/end volume=0; associations unchanged.
  - Discriminates: ignoring Fade Out yields end=105; applying delay after duration changes start; immediate volume assignment incorrectly changes 200 to 0.
  - runtime_observed: no
- **STP-V14** Existing same-target transition is not lengthened (STP-R05; boundary: supplied-input stage)
  - Inputs: one voice, handle=7, associations=[], current volume=40, active=true, retirement=false, transition start tick=80/end tick=120/start volume=0/end volume=255; request handle=7,target=255,duration=30,delay=0,current tick=100; Fade In=10,Fade Out=5.
  - Seeds: no draw.
  - Draws: none.
  - Trace: effective duration=30; candidate end=130 >=120, so skip before writes.
  - Outputs: current volume=40; active=true; retirement=false; transition start tick=80/end tick=120/start volume=0/end volume=255; associations unchanged.
  - Discriminates: replacing every request would produce a transition ending at 130.
  - runtime_observed: no
- **STP-V15** Fade to zero resists nonzero retargeting (STP-R05; boundary: supplied-input stage)
  - Inputs: one voice, handle=7, associations=[], current volume=40, active=true, retirement=true, transition start tick=80/end tick=120/start volume=200/end volume=0; request handle=7,target=255,duration=5,delay=0,current tick=100; Fade In=2,Fade Out=5.
  - Seeds: no draw.
  - Draws: none.
  - Trace: target differs, but entering transition endpoint is zero and requested target is nonzero; skip.
  - Outputs: current volume=40; active=true; retirement=true; transition start tick=80/end tick=120/start volume=200/end volume=0; associations unchanged.
  - Discriminates: unrestricted retargeting would replace endpoint zero with 255.
  - runtime_observed: no
- **STP-V16** Immediate set retains pending transition (STP-R05; boundary: supplied-input stage)
  - Inputs: one voice, handle=7, associations=[], current volume=40, active=true, retirement=false, transition start tick=120/end tick=180/start volume=10/end volume=90; request handle=7,target=80,duration=0,delay=0,current tick=100; Fade In=2,Fade Out=5.
  - Seeds: no draw.
  - Draws: none.
  - Trace: target is neither endpoint with a duration minimum; prior target is nonzero and differs. Immediate set writes 80 only.
  - Outputs: current volume=80; active=true; retirement=false; transition start tick=120/end tick=180/start volume=10/end volume=90; associations unchanged.
  - Discriminates: clearing active or replacing interval fields during an immediate set violates the retained-state assertion.
  - runtime_observed: no
- **STP-V17** Multiple emitter associations suppress scheduling (STP-R05; boundary: supplied-input stage)
  - Inputs: one voice, handle=7, associations=[unit A,unit B], current volume=200, active=false, retirement=false, transition start tick=0/end tick=0/start volume=0/end volume=0; request handle=7,target=0,duration=2,delay=3,current tick=100; Fade In=9,Fade Out=5.
  - Seeds: no draw.
  - Draws: none.
  - Trace: first matching voice has a second association; return before duration adjustment or retirement write.
  - Outputs: current volume=200; active=false; retirement=false; transition start tick=0/end tick=0/start volume=0/end volume=0; associations=[unit A,unit B].
  - Discriminates: treating a second emitter as irrelevant would set retirement and schedule an end at 108.
  - runtime_observed: no

- **STP-V18** Successful supplied device start (STP-R04a; boundary: supplied-input stage)
  - Inputs: valid buffer context; buffer-start marker=0; supplied device result=0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: enter lock; request repeating playback; signed result 0 is nonnegative, so marker=1; exit lock; success.
  - Outputs: result=1; marker=1; operation log=[enter audio device lock, request repeating buffer playback, exit audio device lock].
  - Discriminates: treating only strictly positive device results as success would return 0 and retain marker 0.
  - runtime_observed: no
- **STP-V19** Failed start retains old marker (STP-R04a; boundary: supplied-input stage)
  - Inputs: valid buffer context; buffer-start marker=1; supplied device result=-1.
  - Seeds: no draw.
  - Draws: none.
  - Trace: enter lock; request repeating playback; signed result -1 is negative; exit without marker write.
  - Outputs: result=0; marker=1; operation log=[enter audio device lock, request repeating buffer playback, exit audio device lock].
  - Discriminates: clearing marker on failure would write 0; omitting unlock would fail the ordered log.
  - runtime_observed: no

### 8.2 Observation vectors

**STP-O01** Measure a naturally initiated volume transition at the start, at the end and immediately after it, comparing active state and current volume against STP-R02. [Reproduction procedure and collection limits](../research/spec-evidence/sound-table-and-playback-rules.md#o01-transition-observation).

## 9. Acceptance and dependencies

STP-V01–STP-V19 define partial acceptance for the verified supplied-input stages. T01 supports STP-R01 only. Full reproduction requires resolving STP-Q01 and STP-Q02 and recording matching gameplay observations; passing the verified stages does not establish sound-table playback parity. Depends on NRF-R01/NRF-R02 for exact group RNG steps/draw forms and the numeric foundation's integer arithmetic conventions. No complete downstream consumer is established here.

## 10. Related specs

[Numeric and RNG foundations](core-numeric-and-rng-foundations.md) supplies arithmetic terminology. Positional, unit, ambient, skill/missile and music sound contracts own their respective request sources and spatial policy.

## Appendix A. Provenance

Original Windows x86 Expansion 1.13c; [source manifest](../research/source-manifest.md) identifies D2Client and D2Sound. Static readings and a narrow Storm-loader table read were performed on 2026-10-29. No gameplay vector is runtime-confirmed. The [companion](../research/spec-evidence/sound-table-and-playback-rules.md) records original function locations, byte corroboration, actual collection failures and remaining boundaries.
