# Screen fades

**Description**

A temporary overlay can darken part of the game view and then diminish with elapsed time. The client prepares the overlay separately from lights in the world. The rules here describe its timing and drawing request; the conditions that start it and its final appearance remain partly unresolved.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md)
Ids: SFD-R, SFD-Q, SFD-V, SFD-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/draw-screen-fades.md)
Open evidence gaps: [gap file](../plans/gaps/draw-screen-fades.md)

## 1. Scope

This contract covers the supplied start, cancel and reached draw stages of the timed screen overlay, through its graphics request. Gameplay eligibility, outer draw scheduling and graphics-backend pixel blending remain integration boundaries. World light maps, room lighting, unit tint and palette selection are separate behaviors. [World clock](drlg-world-clock.md) and [Tainted Sun](quest-tainted-sun.md) own environmental time and eclipse transitions; no direct dependency from this overlay to those contracts is established.

## 2. Trigger and ownership

D2Client owns the timer and overlay request. A client draw stage reaches this operation after other scene operations. A unit-processing path can request its start, but that path's gameplay trigger and eligibility remain SFD-Q01. The contract accepts a reached start, cancel or draw event; it does not promise a server-frame cadence.

## 3. Inputs and data

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| event | supplied reached operation | start, cancel or draw | Operation to execute |
| deadline | retained client overlay state | u32 milliseconds; zero means inactive | Absolute expiration reading |
| current time | client elapsed-time reading | u32 milliseconds | Clock supplied at the operation's sampling point |
| width, height | current view dimensions | i32 | Drawing bounds |
| view selector | client view state | i32 | 1 selects left half; 2 selects right half; all other values select full width |
| graphics request log | supplied graphics boundary | ordered tuples | Logs left, top, right, bottom, color selector and strength; final pixels unresolved |
| time sampling log | supplied clock boundary | ordered time readings | Records whether and when an operation samples the clock; no stored game field is required |

Constants 500 milliseconds, 255, 47 and selectors 1/2 originate in engine code and are fixed by original code. No archive-table cells are consumed. Arithmetic below uses low32 for addition, subtraction and multiplication; `i32` interprets the low32 result as signed. Division truncates toward zero. All operations draw no RNG.

## 4. Rules (execution order)

**SFD-R01. Start and cancel** [verified in 1.13c code (static)]. Start samples current time once and stores `deadline = low32(current time + 500)`, replacing any prior deadline. Cancel stores zero without sampling the clock. Neither event emits a graphics request.

**SFD-R02. Inactivity and expiration** [verified in 1.13c code (static)]. On draw, deadline zero returns without sampling the clock or requesting graphics. Otherwise sample current time. If unsigned current time exceeds unsigned deadline, clear deadline and return without graphics. Equality remains active. Comparison is not a wrap-aware elapsed-duration comparison.

**SFD-R03. Overlay request** [verified in 1.13c code (static)]. For an active draw, initialize left 0, right width, top 0 and bottom `i32(height - 47)`. Selector 1 changes right to `i32(width - trunc(width/2))`; selector 2 changes left to `trunc(width/2)`; other selectors leave full width. Compute `strength = trunc(i32(low32(deadline-current time)*255)/500)`. If signed right is at least left and signed bottom is at least top, emit one graphics request `(left,top,right,bottom,0,strength)`. Otherwise emit none. Retain deadline in either case. No clamp is applied to strength at this boundary.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| SFD-R01 | Start/cancel timer | verified in 1.13c code (static) | none for reached events; SFD-Q01 gates gameplay eligibility | SFD-V01, SFD-V02, SFD-V08, SFD-V09 |
| SFD-R02 | Expiration decision | verified in 1.13c code (static) | none for reached draw | SFD-V03, SFD-V04, SFD-V10, SFD-V11 |
| SFD-R03 | Prepared graphics request | verified in 1.13c code (static) | SFD-Q02 gates final pixels only | SFD-V04, SFD-V05, SFD-V06, SFD-V07, SFD-V12, SFD-V13, SFD-V14, SFD-V15 |

## 5. Outputs and side effects

| Output | Rule |
|---|---|
| Deadline replaced, canceled or cleared on expiration | SFD-R01, SFD-R02 |
| Time sample requested on start and active draw only | SFD-R01, SFD-R02 |
| Ordered graphics request with bounds, color selector and strength | SFD-R03 |
| Clock sample on start and on a draw with nonzero deadline; none on cancel or inactive draw | SFD-R01, SFD-R02 |

No server message, world-clock write or light-map write occurs in these supplied stages.

## 6. Branches and edge cases

Equality with the deadline emits strength zero rather than clearing the timer. A start that wraps to zero makes the overlay inactive. Other wraparound starts can expire immediately because comparison is unsigned absolute time. Odd widths divide toward zero. Invalid rectangle dimensions suppress forwarding. Unit type, difficulty and PvP do not enter these supplied computations; their upstream eligibility remains SFD-Q01.

## 7. Open questions

**SFD-Q01** [unresolved; integration]. Which gameplay conditions start/cancel this overlay, and what outer scheduling guarantees reach its draw stage? Resolve the complete upstream unit/frame paths; this does not gate supplied-event computations.

**SFD-Q02** [unresolved; integration]. What pixels does each original graphics backend produce for the request, including out-of-range strengths? Resolve backend dispatch and blending; only request generation is established here.

**SFD-Q03** [unresolved; observation]. Matching gameplay observations remain pending for all rules. No completed capture confirms the timer or its request.

## 8. Test vectors

### 8.1 Executable vectors

All vectors use initially empty persistent graphics-request and time-sampling logs. Each case starts independently. `log` below means the graphics-request log. Clock reads append the supplied time to the sampling log; cancel and inactive draw leave that log empty, start and nonzero-deadline draw append one reading. Seeds: no draw. Draws: none. For omitted draw dimensions use width 800, height 600 and selector 0. The graphics boundary logs requests only; pixel blending is outside these vectors.

- **SFD-V01** Replace timer (SFD-R01; boundary: supplied-input stage)
  - Inputs: start, time 1000, deadline 90.
  - Seeds: no draw.
  - Draws: none.
  - Trace: 1000 + 500 = 1500.
  - Outputs: deadline 1500; log empty.
  - Discriminates: replacing 90 rejects extending the old deadline to 590.
  - runtime_observed: no
- **SFD-V02** Cancel (SFD-R01; boundary: supplied-input stage)
  - Inputs: cancel, deadline 1500, time 1000.
  - Seeds: no draw.
  - Draws: none.
  - Trace: store zero.
  - Outputs: deadline 0; log empty; time-sampling log empty.
  - Discriminates: rejects retaining the timer until expiration or sampling the clock while canceling.
  - runtime_observed: no
- **SFD-V03** Expired (SFD-R02; boundary: supplied-input stage)
  - Inputs: draw, deadline 1500, time 1501.
  - Seeds: no draw.
  - Draws: none.
  - Trace: unsigned 1501 exceeds 1500.
  - Outputs: deadline 0; log empty.
  - Discriminates: rejects emitting a negative-strength rectangle.
  - runtime_observed: no
- **SFD-V04** Equality (SFD-R02, SFD-R03; boundary: supplied-input stage)
  - Inputs: draw, deadline 1500, time 1500.
  - Seeds: no draw.
  - Draws: none.
  - Trace: equality survives; strength 0; bottom 600 - 47 = 553.
  - Outputs: deadline 1500; log [(0,0,800,553,0,0)].
  - Discriminates: rejects clearing on greater-than-or-equal comparison.
  - runtime_observed: no
- **SFD-V05** Odd left half (SFD-R03; boundary: supplied-input stage)
  - Inputs: draw, deadline 1500, time 1250, width 801, height 600, selector 1.
  - Seeds: no draw.
  - Draws: none.
  - Trace: half 400; right 401; strength trunc(63750/500) = 127.
  - Outputs: deadline 1500; log [(0,0,401,553,0,127)].
  - Discriminates: rejects floor(width/2) as the left-half right bound and rounding strength to 128.
  - runtime_observed: no
- **SFD-V06** Odd right half (SFD-R03; boundary: supplied-input stage)
  - Inputs: draw, deadline 1500, time 1000, width 801, height 600, selector 2.
  - Seeds: no draw.
  - Draws: none.
  - Trace: left trunc(801/2) = 400; strength 127500/500 = 255.
  - Outputs: deadline 1500; log [(400,0,801,553,0,255)].
  - Discriminates: rejects rounding left upward or interpreting selector 2 as left half.
  - runtime_observed: no
- **SFD-V07** Invalid vertical bound (SFD-R03; boundary: supplied-input stage)
  - Inputs: draw, deadline 1500, time 1000, width 800, height 46, selector 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: bottom -1 is less than top 0.
  - Outputs: deadline 1500; log empty.
  - Discriminates: rejects forwarding the invalid rectangle.
  - runtime_observed: no

- **SFD-V08** Start wraps to zero (SFD-R01; boundary: supplied-input stage)
  - Inputs: start, time 4294966796, deadline 90.
  - Seeds: no draw.
  - Draws: none.
  - Trace: sum 4294967296; low32 is 0.
  - Outputs: deadline 0; log empty.
  - Discriminates: rejects retaining a wider nonzero deadline.
  - runtime_observed: no
- **SFD-V09** Start wraps to nonzero (SFD-R01; boundary: supplied-input stage)
  - Inputs: start, time 4294967000, deadline 90.
  - Seeds: no draw.
  - Draws: none.
  - Trace: sum 4294967500; low32 is 204.
  - Outputs: deadline 204; log empty.
  - Discriminates: rejects saturating the deadline at 4294967295.
  - runtime_observed: no
- **SFD-V10** Inactive draw (SFD-R02; boundary: supplied-input stage)
  - Inputs: draw, deadline 0, current-time source supplies 0 if requested.
  - Seeds: no draw.
  - Draws: none.
  - Trace: zero deadline returns before sampling time.
  - Outputs: deadline 0; log empty; no time sampling.
  - Discriminates: rejects drawing a zero-strength rectangle at time zero.
  - runtime_observed: no
- **SFD-V11** Unsigned expiration (SFD-R02; boundary: supplied-input stage)
  - Inputs: draw, deadline 204, time 4294967000.
  - Seeds: no draw.
  - Draws: none.
  - Trace: unsigned 4294967000 exceeds 204; signed -296 would not.
  - Outputs: deadline 0; log empty.
  - Discriminates: rejects signed or wrap-aware elapsed-duration comparison.
  - runtime_observed: no
- **SFD-V12** Invalid horizontal bound (SFD-R03; boundary: supplied-input stage)
  - Inputs: draw, deadline 1500, time 1000, width -1, height 600, selector 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: right -1 is below left 0.
  - Outputs: deadline 1500; log empty.
  - Discriminates: rejects unsigned rectangle-bound comparison.
  - runtime_observed: no
- **SFD-V13** Unclamped supplied strength (SFD-R03; boundary: supplied-input stage)
  - Inputs: draw, deadline 2000, time 1000, width 800, height 600, selector 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: remaining 1000 times 255 gives 255000; divide by 500 gives 510.
  - Outputs: deadline 2000; log [(0,0,800,553,0,510)].
  - Discriminates: rejects clamping at 255; this is supplied timer state rather than a normal start.
  - runtime_observed: no

- **SFD-V14** Inclusive zero-area rectangle (SFD-R03; boundary: supplied-input stage)
  - Inputs: draw, deadline 1500, time 1000, width 0, height 47, selector 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: left and right both zero; top and bottom both zero; both inclusive comparisons pass; strength 255.
  - Outputs: deadline 1500; log [(0,0,0,0,0,255)]; time-sampling log [1000].
  - Discriminates: rejects suppressing equal bounds with strict comparisons.
  - runtime_observed: no
- **SFD-V15** Wrapped strength multiplication (SFD-R03; boundary: supplied-input stage)
  - Inputs: draw, deadline 4294967295, time 0, width 800, height 600, selector 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: unsigned draw remains active; remaining 4294967295 times 255 has low32 4294967041, interpreted as -255; signed division by 500 truncates to zero.
  - Outputs: deadline 4294967295; log [(0,0,800,553,0,0)]; time-sampling log [0].
  - Discriminates: rejects unbounded-width multiplication or unsigned division of the wrapped product.
  - runtime_observed: no

### 8.2 Observation vectors

**SFD-O01.** Observe a naturally triggered overlay at start, halfway through its lifetime and after expiration. Measure timer sampling, rectangle requests and panel geometry, comparing SFD-R01 through SFD-R03. See the [observation procedure](../research/spec-evidence/draw-screen-fades.md#observation-procedure). Exact gameplay eligibility remains SFD-Q01.

## 9. Acceptance and dependencies

The supplied-stage subset is reproduced when SFD-V01 through SFD-V15 pass. Complete gameplay reproduction additionally requires SFD-Q01 and SFD-Q02 resolved and SFD-O01 observed; passing request vectors establishes no final-pixel parity. SFD-Q03 limits runtime confirmation only.

**Depends on:** [numeric foundations](core-numeric-and-rng-foundations.md), integer-width and division conventions. Graphics rendering is a supplied operation with its integration explicitly unresolved.

**Depended on by:** no established spec consumer.

## 10. Related specs

- [Light map and light radius](draw-light-map-and-light-radius.md): independent world illumination.
- [World clock](drlg-world-clock.md): environmental time; no direct overlay call established.
- [Tainted Sun](quest-tainted-sun.md): quest eclipse transitions; no direct overlay call established.

## Appendix A. Provenance

Original Windows x86 Expansion 1.13c D2Client supplies timing and request generation. Static reading on 2026-10-29 used byte-matched reconstruction where available and original instructions otherwise, including raw-image corroboration of the time import. [Source manifest](../research/source-manifest.md) identifies the client image. The [companion](../research/spec-evidence/draw-screen-fades.md) records references and limits. No gameplay assertion is runtime-confirmed.
