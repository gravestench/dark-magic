# Screen fades

**Description**

A temporary overlay can darken part of the game view and then diminish with elapsed time. The client prepares the overlay separately from lights in the world. The rules describe timing, drawing requests and the software-rendered colour changes. The broader start conditions and hardware-rendered appearance remain partly unresolved.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md)
Ids: SFD-R, SFD-Q, SFD-V, SFD-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/draw-screen-fades.md)
Open evidence gaps: [gap file](../plans/gaps/draw-screen-fades.md)

## 1. Scope

This contract covers the supplied start, cancel and reached draw stages of the timed screen overlay, through its graphics request and the selected software indexed-pixel stage. Gameplay eligibility, outer draw scheduling, prepared-renderer integration and hardware pixel blending remain integration boundaries. World light maps, room lighting, unit tint and palette selection are separate behaviors. [World clock](drlg-world-clock.md) and [Tainted Sun](quest-tainted-sun.md) own environmental time and eclipse transitions; no direct dependency from this overlay to those contracts is established.

## 2. Trigger and ownership

D2Client owns the timer and overlay request. A client draw stage reaches this operation after other scene operations. A unit-processing path can request its start, but that path's gameplay trigger and eligibility remain SFD-Q01. The contract accepts a reached start, cancel or draw event; it does not promise a server-frame cadence.

## 3. Inputs and data

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| event | supplied reached operation | start, cancel, draw, or selected software pixel stage | Operation to execute; a draw vector ends at its graphics request unless the software stage is separately supplied |
| deadline | retained client overlay state | u32 milliseconds; zero means inactive | Absolute expiration reading |
| current time | client elapsed-time reading | u32 milliseconds | Clock supplied at the operation's sampling point |
| width, height | current view dimensions | i32 | Drawing bounds |
| view selector | client view state | i32 | 1 selects left half; 2 selects right half; all other values select full width |
| graphics request log | supplied graphics boundary | ordered tuples | Logs left, top, right, bottom, color selector and strength |
| Software backend | supplied initialized renderer | GDI or DirectDraw | Selects SFD-R04a or SFD-R04b; preparation and presentation remain SFD-Q02 |
| Software view and request | supplied reached fade request | positive i32 view width/height; ordered rectangle from SFD-R03, colour selector 0, i32 strength | The request already passed R03 guards |
| Software raster extents | supplied admitted buffer | positive i32 horizontal extent P and vertical extent B | Horizontal clipping/row advance uses current GDI extent or DirectDraw lock-returned pitch; these are independent of the logical view width/height |
| Queried software pixel | supplied indexed buffer cell | integer 0 <= x < P and 0 <= y < B; old index u8 | Defines one independent pixel result; a block operation applies the same rule to every cell |
| Low/middle/high fade translation columns | supplied prepared software palette tables | three maps from old index u8 to resulting index u8 | Needed only for the selected non-solid band; resolved from original profile palette data, not an invented RGB-opacity formula |
| time sampling log | supplied clock boundary | ordered time readings | Records whether and when an operation samples the clock; no stored game field is required |

Constants 500 milliseconds, 255, 47 and selectors 1/2 originate in engine code and are fixed by original code. R01 through R03 consume no archive cells. The software stage consumes prepared fade columns from palette data; their original profile binding is recorded in the companion. Thresholds 64, 128, 192, low-byte selection, colour selector 0 and endpoint-minus-one clipping originate in engine code. Arithmetic below uses low32 for addition, subtraction and multiplication; `i32` interprets the low32 result as signed. Division truncates toward zero. All operations draw no RNG.

## 4. Rules (execution order)

**SFD-R01. Start and cancel** [runtime-confirmed (1.13c), SFD-V16 start from inactive only; cancel, replacement of a nonzero deadline and wrap cases remain static]. Start samples current time once and stores `deadline = low32(current time + 500)`, replacing any prior deadline. Cancel stores zero without sampling the clock. Neither event emits a graphics request.

**SFD-R02. Inactivity and expiration** [runtime-confirmed (1.13c), SFD-V17 inactivity and SFD-V19 expiration only; equality and wrap cases remain static]. On draw, deadline zero returns without sampling the clock or requesting graphics. Otherwise sample current time. If unsigned current time exceeds unsigned deadline, clear deadline and return without graphics. Equality remains active. Comparison is not a wrap-aware elapsed-duration comparison.

**SFD-R03. Overlay request** [runtime-confirmed (1.13c), SFD-V18 full-width and V23/V24 half-view active draws; other geometry and arithmetic cases remain static]. For an active draw, initialize left 0, right width, top 0 and bottom `i32(height - 47)`. Selector 1 changes right to `i32(width - trunc(width/2))`; selector 2 changes left to `trunc(width/2)`; other selectors leave full width. Compute `strength = trunc(i32(low32(deadline-current time)*255)/500)`. If signed right is at least left and signed bottom is at least top, emit one graphics request `(left,top,right,bottom,0,strength)`. Otherwise emit none. Retain deadline in either case. No clamp is applied to strength at this boundary.

**SFD-R04a. GDI software pixel** [runtime-confirmed (1.13c), SFD-V22 tracked solid-band pixel only; other pixel, table and clipping cases remain static]. At the separately supplied software stage, use positive supplied raster extents P/B, valid admitted storage, and an ordered colour-0 fade request from R03. Clip left/top upward to zero and right/bottom downward to P minus 1 / B minus 1. A queried cell participates only when `left <= x < right` and `top <= y < bottom` after clipping. Otherwise retain its old index. For a participating cell, let `a = u8(strength)`. If a is zero, retain the old index without a lookup. For a in 1..63, 64..127 or 128..191, replace the index with the low, middle or high prepared column entry at the old index, respectively. For a in 192..255, replace it with index 0 without a lookup. No time sample, timer change, graphics request or RNG draw occurs in this stage. An index is not an RGB colour; RGB palette preparation and presentation remain Q02.

**SFD-R04b. DirectDraw software pixel** [verified in 1.13c code (static)]. For the same supplied prepared-state boundary, apply the exact clipping, participation, low-byte band selection and index calculation in R04a. This establishes indexed writes only; DirectDraw palette gamma, surface admission, restoration and presentation remain Q02. No matching DirectDraw pixel case was observed.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| SFD-R01 | Start/cancel timer | runtime-confirmed (1.13c), V16 start only; other cases static | none for reached events; SFD-Q01 gates gameplay eligibility | SFD-V01, SFD-V02, SFD-V08, SFD-V09, SFD-V16 |
| SFD-R02 | Expiration decision | runtime-confirmed (1.13c), V17 inactivity and V19 expiration; other cases static | none for reached draw | SFD-V03, SFD-V04, SFD-V10, SFD-V11, SFD-V17, SFD-V19 |
| SFD-R03 | Prepared graphics request | runtime-confirmed (1.13c), V18 full-width and V23/V24 half-view active draws; other cases static | SFD-Q02 gates final pixels only | SFD-V04, SFD-V05, SFD-V06, SFD-V07, SFD-V12, SFD-V13, SFD-V14, SFD-V15, SFD-V18, SFD-V23, SFD-V24 |
| SFD-R04a | Supplied GDI indexed pixel | runtime-confirmed (1.13c), V22 solid-band sample only; other cases static | Q02 gates preparation/presentation, not this supplied indexed-pixel calculation | SFD-V20, SFD-V21, SFD-V22 |
| SFD-R04b | Supplied DirectDraw indexed pixel | verified in 1.13c code (static) | Q02 gates gamma/surface/presentation integration | SFD-V20, SFD-V21 |

## 5. Outputs and side effects

| Output | Rule |
|---|---|
| Deadline replaced, canceled or cleared on expiration | SFD-R01, SFD-R02 |
| Time sample requested on start and active draw only | SFD-R01, SFD-R02 |
| Ordered graphics request with bounds, color selector and strength | SFD-R03 |
| Clock sample on start and on a draw with nonzero deadline; none on cancel or inactive draw | SFD-R01, SFD-R02 |
| Resulting queried software pixel index, or unchanged index outside participation/at zero low-byte strength | SFD-R04a, SFD-R04b |

No server message, world-clock write or light-map write occurs in these supplied stages.

## 6. Branches and edge cases

Equality with the deadline emits strength zero rather than clearing the timer. A start that wraps to zero makes the overlay inactive. Other wraparound starts can expire immediately because comparison is unsigned absolute time. Odd widths divide toward zero. Invalid rectangle dimensions suppress forwarding. Unit type, difficulty and PvP do not enter these supplied computations; their upstream eligibility remains SFD-Q01.

## 7. Open questions

**SFD-Q01** [unresolved; integration]. Which gameplay conditions start/cancel this overlay, and what outer scheduling guarantees reach its draw stage? Resolve the complete upstream unit/frame paths; this does not gate supplied-event computations.

**SFD-Q02** [unresolved; integration]. What pixels does each original graphics backend produce for the request, including out-of-range strengths? Resolve full renderer admission, palette preparation/lifetime, hardware blending and presentation. R04a/R04b establish supplied software indexed pixels; they do not close this integration question.

**SFD-Q03** [unresolved; observation]. Complete matching cases establish SFD-V16 start from inactive, SFD-V17 inactive draw, SFD-V18 full-width active drawing and SFD-V19 expiration, and SFD-V23/V24 half-view drawing, including their sampling/output assertions. Cancel, nonzero replacement, equality, odd-width half-view, wrap and malformed-bound cases remain unobserved. The [companion](../research/spec-evidence/draw-screen-fades.md) distinguishes complete cases from partial observations.

## 8. Test vectors

### 8.1 Executable vectors

All vectors use initially empty persistent graphics-request and time-sampling logs. Each case starts independently. `log` below means the graphics-request log. Clock reads append the supplied time to the sampling log; cancel and inactive draw leave that log empty, start and nonzero-deadline draw append one reading. Seeds: no draw. Draws: none. For omitted draw dimensions use width 800, height 600 and selector 0. V01 through V19 and V23/V24 end at the graphics-request boundary and assert no pixel blending. V20 through V22 separately exercise the supplied software stage and its indexed-pixel outputs; they assert no RGB presentation or hardware pixels. The software stage leaves both logs empty.

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

- **SFD-V16** Natural start from inactive (SFD-R01; boundary: reached start)
  - Inputs: start, current time 1256016886, deadline 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: sample once; 1256016886 + 500 = 1256017386.
  - Outputs: deadline 1256017386; log empty; time-sampling log [1256016886].
  - Discriminates: rejects retaining zero, a different duration, repeated time sampling or emitting a rectangle during start. Replacement of a nonzero deadline is SFD-V01.
  - runtime_observed: yes; 2026-10-30; [complete sampling record](../research/spec-evidence/draw-screen-fades.md#complete-start-and-inactive-cases-2026-10-30).
- **SFD-V17** Natural inactive draw (SFD-R02; boundary: reached draw)
  - Inputs: draw, deadline 0, current-time source 1256327885 if requested, width 800, height 600, selector 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: return on zero deadline before requesting current time.
  - Outputs: deadline 0; log empty; time-sampling log empty.
  - Discriminates: rejects an unconditional time sample, an emitted zero-strength rectangle or a changed timer.
  - runtime_observed: yes; 2026-10-30; [complete sampling record](../research/spec-evidence/draw-screen-fades.md#complete-start-and-inactive-cases-2026-10-30).

- **SFD-V18** Natural full-width active draw (SFD-R02, SFD-R03; boundary: reached draw through graphics request)
  - Inputs: draw, deadline 1359628702, current time 1359628441, width 800, height 600, selector 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: unsigned time is below deadline; remaining 261; product 66555; trunc(66555/500) = 133; bottom 553. Sample once.
  - Outputs: deadline 1359628702; log [(0,0,800,553,0,133)]; time-sampling log [1359628441].
  - Discriminates: rejects changing the deadline, reversing the fade direction, rounding strength upward to 134, omitting the request, or requesting time twice.
  - runtime_observed: yes; 2026-10-31; [active and expired cases](../research/spec-evidence/draw-screen-fades.md#complete-active-and-expired-cases-2026-10-31).
- **SFD-V19** Natural expiration (SFD-R02; boundary: reached draw)
  - Inputs: draw, deadline 1359628702, current time 1359628725, width 800, height 600, selector 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: unsigned current time exceeds deadline by 23; sample once, clear timer and return.
  - Outputs: deadline 0; log empty; time-sampling log [1359628725].
  - Discriminates: rejects retaining the expired timer, emitting a negative-strength request or skipping the time sample. Equality is a separate case.
  - runtime_observed: yes; 2026-10-31; [active and expired cases](../research/spec-evidence/draw-screen-fades.md#complete-active-and-expired-cases-2026-10-31).

- **SFD-V20** Software endpoint clipping (SFD-R04a, SFD-R04b; boundary: separately supplied software stage)
  - Inputs: independently select GDI and DirectDraw; logical view width 4, height 50; request (0,0,4,3,0,192); deadline 13; time source 1000 if requested. Each subcase starts fresh with old index 7 at every queried cell and constant low/middle/high columns 9/10/11. Query every cell in the supplied P-by-B raster. Cases: (a) P=4,B=50; (b) padded P=8,B=50; (c) P=4,B=2; (d) P=1,B=1.
  - Seeds: no draw.
  - Draws: none.
  - Trace: (a) clips right/bottom to 3/3; (b) right/bottom 4/3; (c) 3/1; (d) 0/0. Band 192..255 writes index 0 only inside each half-open region.
  - Outputs: (a) x=0..2,y=0..2 yield 0 (9 cells), other 191 yield 7; (b) x=0..3,y=0..2 yield 0 (12), other 388 yield 7; (c) x=0..2,y=0 yield 0 (3), other 5 yield 7; (d) the sole cell yields 7. Deadline 13 and both logs remain unchanged/empty in every subcase.
  - Discriminates: rejects substituting logical view width for lock pitch, inclusive endpoints, ignoring raster-height clipping or zero spans, and treating strength 192 as a table band.
  - runtime_observed: no
- **SFD-V21** Software bands and low-byte wrap (SFD-R04a, SFD-R04b; boundary: separately supplied software stage)
  - Inputs: independently select GDI and DirectDraw; logical width 4, height 50, raster P=4/B=50; each request (0,0,4,3,0,strength) uses a fresh case and deadline 17, time source 1000 if requested. Query inside (0,0) and excluded last-column (3,0), old index 7 in both. Low column at 7 is 9 and every other entry 50; middle at 7 is 10 and every other entry 51; high at 7 is 11 and every other entry 52.
  - Seeds: no draw.
  - Draws: none.
  - Trace: select only the low byte, then the band; excluded cells bypass lookup/write.
  - Outputs: for strengths [0,1,63,64,127,128,191,192,255,256,257,-1], inside results are [7,9,9,10,10,11,11,0,0,7,9,0]. Every excluded result is 7; deadline remains 17; both logs empty.
  - Discriminates: rejects swapped bands, threshold errors, clamping high bits instead of low-byte wrapping, or using a fixed table index instead of old index 7.
  - runtime_observed: no
- **SFD-V22** Observed GDI solid-band sample (SFD-R04a; boundary: separately supplied GDI indexed-pixel stage)
  - Inputs: GDI, logical width 800, height 600, raster P=800/B=600; request (0,0,800,553,0,252); queried cell (0,0), old index 250; deadline 1359375821; time source 1359375326 if requested. Use the original-profile Act-I columns; at old index 250, low/middle/high entries are 247/18/16. Other entries are not queried by this case; the correct solid band reads no column.
  - Seeds: no draw.
  - Draws: none.
  - Trace: clipped right 799 and bottom 553 include the queried cell; low-byte strength 252 selects solid index 0.
  - Outputs: queried index 0; deadline 1359375821; both logs empty. Other cells are not assertions of this observation.
  - Discriminates: rejects leaving this nonzero pixel unchanged or applying a translation-table band. It establishes no RGB/presentation or other-backend match.
  - runtime_observed: yes; 2026-10-31; [software pixel evidence](../research/spec-evidence/draw-screen-fades.md#software-backend-and-effective-palette-evidence-2026-10-31).

- **SFD-V23** Natural left-half active draw (SFD-R02, SFD-R03; boundary: reached draw through graphics request)
  - Inputs: draw, deadline 1364277377, current time 1364277083, width 800, height 600, selector 1.
  - Seeds: no draw.
  - Draws: none.
  - Trace: sample once; remaining 294, product 74970; trunc(74970/500) = 149; half width 400, bottom 553.
  - Outputs: deadline 1364277377; log [(0,0,400,553,0,149)]; time-sampling log [1364277083].
  - Discriminates: rejects full-width or reversed-half geometry, changing the timer, rounding strength upward, or repeated time sampling. Odd-width behavior remains a separate static case.
  - runtime_observed: yes; 2026-10-31; [complete half-view cases](../research/spec-evidence/draw-screen-fades.md#complete-half-view-cases-2026-10-31).
- **SFD-V24** Natural right-half active draw (SFD-R02, SFD-R03; boundary: reached draw through graphics request)
  - Inputs: draw, deadline 1364277377, current time 1364277242, width 800, height 600, selector 2.
  - Seeds: no draw.
  - Draws: none.
  - Trace: sample once; remaining 135, product 34425; trunc(34425/500) = 68; half width 400, bottom 553.
  - Outputs: deadline 1364277377; log [(400,0,800,553,0,68)]; time-sampling log [1364277242].
  - Discriminates: rejects full-width or reversed-half geometry, changing the timer, rounding strength upward, or repeated time sampling. Odd-width behavior remains a separate static case.
  - runtime_observed: yes; 2026-10-31; [complete half-view cases](../research/spec-evidence/draw-screen-fades.md#complete-half-view-cases-2026-10-31).

### 8.2 Observation vectors

**SFD-O01.** Observe a naturally triggered overlay at start, halfway through its lifetime and after expiration. Measure timer sampling, rectangle requests and panel geometry, comparing SFD-R01 through SFD-R03. See the [observation procedure](../research/spec-evidence/draw-screen-fades.md#observation-procedure). Exact gameplay eligibility remains SFD-Q01.

## 9. Acceptance and dependencies

The supplied-stage subset is reproduced when SFD-V01 through SFD-V24 pass. Complete gameplay reproduction additionally requires SFD-Q01 and SFD-Q02 resolved and SFD-O01 observed; passing request vectors establishes no final-pixel parity. SFD-Q03 limits runtime confirmation only.

**Depends on:** [numeric foundations](core-numeric-and-rng-foundations.md), integer-width and division conventions. Graphics rendering is a supplied operation with its integration explicitly unresolved.

**Depended on by:** no established spec consumer.

## 10. Related specs

- [Light map and light radius](draw-light-map-and-light-radius.md): independent world illumination.
- [World clock](drlg-world-clock.md): environmental time; no direct overlay call established.
- [Tainted Sun](quest-tainted-sun.md): quest eclipse transitions; no direct overlay call established.

## Appendix A. Provenance

Original Windows x86 Expansion 1.13c D2Client supplies timing and request generation. Static reading on 2026-10-29 used byte-matched reconstruction where available and original instructions otherwise, including raw-image corroboration of the time import. [Source manifest](../research/source-manifest.md) identifies the client image. The [companion](../research/spec-evidence/draw-screen-fades.md) records references and limits. Runtime confirmation on 2026-10-30 covers SFD-V16 and SFD-V17; 2026-10-31 adds SFD-V18, SFD-V19 and only the tracked GDI pixel in SFD-V22, plus SFD-V23/V24 half-view draws. SFD-V01 through SFD-V15 remain unobserved as complete cases.
