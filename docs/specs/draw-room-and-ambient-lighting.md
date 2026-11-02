# Room and ambient lighting

**Description**

Rooms have background light before nearby flames, creatures and other local sources brighten them. The client chooses a room's light from area overrides, the level's lighting settings, or the act's current environment, then fills its local lighting map. Some changing area colours and day-cycle lighting are still unresolved.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-table-records](core-game-data-table-records.md)
Ids: RAL-R, RAL-Q, RAL-V, RAL-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/draw-room-and-ambient-lighting.md)
Open evidence gaps: [gap file](../plans/gaps/draw-room-and-ambient-lighting.md)

## 1. Scope

This contract covers the client selection of room background intensity and colour, its prepared level-field consumer, and initialization of the 48 by 48 local light map before obstruction sampling and local-light accumulation. It also identifies the environment and special-area results still requiring evidence. [Light map and radius](draw-light-map-and-light-radius.md) owns rebuild eligibility, collision sampling and local sources. [World clock](drlg-world-clock.md) owns clock assignments, progression and its verified eclipse intensity branches; [Tainted Sun](quest-tainted-sun.md) owns quest triggers. Final pixel conversion, weather, audio, complete room generation and screen fades are outside this boundary.

## 2. Trigger and ownership

D2Client's reached viewport-render rebuild stage initializes background lighting, then fills viewport obstruction state, then updates and accumulates local light sources. The original client `DRAW\dLightMap` operations compute the map result; D2Common provides room metadata, ready level fields and stored environment light. LMR-R11 owns whether a reached viewport stage requests a rebuild. This contract defines one such reached initialization, without assigning a server or display-frame frequency. No server message is emitted by these operations.

## 3. Inputs and data

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| Reached initialization; clear-only condition | Supplied client rebuild stage | Event; Boolean | Select ambient initialization or a zero map |
| Local focus player and room | Supplied client focus and resolved current room | Present/absent references | Select initial background and ordered neighboring rooms |
| Room level identity | Ready room's level metadata | Signed 32-bit integer | Chooses the level record and special-area branches |
| Ready level collection | Prepared level-definition records and count | Ordered collection | A successful field read requires a present selected record and index in range |
| Level intensity / RGB | Prepared `Levels.Intensity`, `Red`, `Green`, `Blue` fields | Four unsigned bytes | Background tuple; named-field preparation consumes GDA-R03a/GDA-R03b/GDA-R03c |
| Use area overrides | Reached room-light operation | Boolean | Rebuild supplies true; false bypasses area-override selection |
| Client override intensity / RGB | Supplied current client override state | Signed 32-bit intensity; three unsigned bytes | Normal area override candidate |
| Special-area enable condition / phase | Supplied client state for levels 107 and 108 | Boolean; signed 32-bit integer | Fixed special colour only when enabled and phase equals -1 |
| Level-8 suppression / initialization / active condition / phase | Supplied level-8 client state | Boolean; Boolean; Boolean; signed 32-bit integer | RAL-R03; a nonsuppressed successful invocation requires initialization |
| Fixed phase-lookup argument | Supplied reached level-8 lookup operation | Signed 32-bit integer | RAL-R03c selects the fixed engine constant; phase-to-argument production remains RAL-Q01 |
| Act and environment | Supplied current client act; optional environment | Present act; environment optional | Stored environment fallback; absent environment supplies intensity 0 and leaves colour bytes unchanged |
| Environment light | Stored act environment | Signed 32-bit intensity; three unsigned bytes | Produced through the world-clock presentation boundary; general production remains RAL-Q03 |
| Environment presentation request | World-clock presentation boundary WCK-R04/WCK-R08b | Signed 32-bit source; act index 0..4; present environment | RAL-R07a's act index 3 is Act IV; the source is supplied, not inferred from the current room |
| Map origin | Supplied local map window | Signed 32-bit x/y in subtiles | Relative clipping origin |
| Room rectangle | Supplied nonnull room's subtile bounds | Signed 32-bit x/y/width/height | Adjacent-room fill bounds |
| Ordered adjacent-room sequence | Supplied current room metadata | Ordered nonnull room references | Later entries overwrite earlier filled cells; current-room entries are skipped |
| Map cells | Supplied 48 by 48 light map | Obstruction state and unsigned intensity/RGB bytes | Background initialization outputs |
| Fill bounds | Supplied rectangular fill operation | Signed 32-bit top/bottom/left/right indices, inclusive | Empty ranges write nothing; nonempty ranges must lie in 0..47 |
| Fill intensity and colour | Supplied rectangular fill operation | Signed 32-bit intensity; unsigned RGB bytes | Tuple to replace selected map cells through RAL-R05 |
| Existing level-reader output tuple | Caller-supplied level-reader result state | Four unsigned bytes | RAL-R01 replaces all four values; they do not select fallback |

| Mapping or constant | Origin | Binding |
|---|---|---|
| Map side 48, last index 47 | Engine code | Fixed by original code |
| Levels 8, 107 and 108; phase sentinel -1 | Engine code | Fixed branch identities; this contract does not establish their names or remapping |
| Special RGB (255,64,48), intensity 160 or 80 | Engine code | Fixed by original code |
| Fixed phase table `T[0..511]` | Initialized engine constants | Exact `binary32` values defined below; RAL-R03c consumes this fixed original profile |
| Act-IV source selectors 103..106 and target intensities 128,64,56,48; other-source target 16 | Engine code | Fixed by original code; distinct from prepared level-light fields |
| Prepared level-field names | Table schema and consumer code | Named fields selected from prepared level-definition records; original-profile correspondence is established by the companion's complete compiled/text/ready comparison |

The fixed table has the following exact value definition for integer `j` in 0..511:

```
T[j] = RNE32(sin(j * (13176795 / 1073741824)))
```

The fraction, multiplication and sine are exact real operations; the sine argument is in radians. `RNE32` rounds once to the nearest IEEE 754 `binary32` value, with an exact midpoint choosing the value whose significand integer is even. `T[0]` is positive zero. There is no intermediate rounding of the angle. This expression defines the engine-owned constant values; it requires no particular construction algorithm or runtime sine call. The [exhaustive equivalence record](../research/spec-evidence/draw-room-and-ambient-lighting.md#exact-fixed-lookup-equivalence-2026-10-31) verifies every index against the original fixed table. It does not establish an original table-generation algorithm or the remaining changing-intensity branch.

No claim assigns a lighting column to `LvlTypes`. The [effective-header observation](../research/spec-evidence/draw-room-and-ambient-lighting.md#table-data-records) records its actual columns; downstream influence through tile resources is RAL-Q04.

**RAL-Q02** Original-profile correspondence is established: all 137 prepared level-definition records match the effective compiled resource, and their 548 lighting bytes match the effective text fields selected by explicit level Id and named column. This measured profile supports RAL-R01's named-field interpretation; it does not establish generic decoder correctness for other resources or profiles. The [complete correspondence record](../research/spec-evidence/draw-room-and-ambient-lighting.md#complete-level-definition-correspondence-2026-10-30) records identities and limits.

## 4. Rules (execution order)

All verified operations below draw no RNG. Their integer operations use 32-bit wrap unless byte narrowing is specified.

**RAL-R01** Read prepared level light [runtime-confirmed (1.13c)]. For a supplied room with valid level metadata and an in-range prepared record, return its unsigned `Intensity`, `Red`, `Green`, `Blue` bytes, replacing all four caller output values even when the selected tuple is all zero. The level identity selects the record by collection index, not by searching a stored identifier. Missing room metadata or an invalid selected record has no successful fallback contract here; it is outside the valid prepared-input domain. Matching observation covers only RAL-V09's original-profile level-1 zero tuple and output replacement; other prepared values remain statically verified.

**RAL-R02** Select a non-level-8 override candidate [verified in 1.13c code (static)]. Start candidate RGB at (0,0,0). If use-area-overrides is false, go directly to RAL-R04's level fallback. Otherwise, for levels 107 and 108, use intensity 160 and RGB (255,64,48) only when the special-area enable condition is true and its phase equals -1. If either condition fails, use candidate intensity 0 and RGB (0,0,0); do not substitute the normal client override. For any other level except 8, copy the client override intensity and RGB. Any candidate with at least one nonzero RGB byte is final, even when intensity is zero. An all-zero RGB candidate selects RAL-R04, regardless of its intensity. Level 8 instead uses RAL-R03.

**RAL-R03a** Select the fixed or inactive level-8 candidate [verified in 1.13c code (static)]. With use-area-overrides true and level identity 8, start RGB at (0,0,0). A true suppression condition returns candidate intensity 0 without inspecting initialization or active condition. Otherwise require initialization true. An inactive condition returns intensity 0 and RGB (0,0,0). An active condition with phase -1 returns intensity 80 and RGB (255,64,48). Nonzero RGB finishes selection; zero RGB selects RAL-R04.

**RAL-R03b** Select a changing level-8 candidate [unresolved]. With nonsuppressed, initialized, active level-8 state and phase other than -1, the operation supplies RGB (255,64,48) and a phase-dependent intensity. The exact intensity for this branch requires RAL-Q01. It remains gated; no default intensity is specified.

**RAL-R03c** Read the fixed phase lookup [verified in 1.13c code (static)]. For the supplied signed 32-bit lookup argument `a`, calculate `j=u32(a+128) AND 511` with 32-bit wrapping addition, and return the exact `binary32` value `T[j]` defined in section 3. This operation writes no state and draws no RNG. It establishes only the fixed lookup; phase-to-argument production, intensity arithmetic and its floating-environment preconditions, full eligibility and natural transition remain RAL-Q01.

**RAL-R04** Select level or environment fallback [verified in 1.13c code (static)]. Request RAL-R01. If at least one level RGB byte is nonzero, return that complete level tuple, including intensity zero when supplied. If all three are zero, request the current act's stored environment RGB, then its stored intensity. A present environment replaces all three colour bytes and the intensity. An absent environment leaves the zero colour bytes and returns intensity 0. A zero intensity alone never triggers fallback. This reads stored light; it does not itself advance the clock or compute an environment colour.

**RAL-R05** Fill a supplied map rectangle [verified in 1.13c code (static)]. Inputs are inclusive row bounds `top..bottom`, inclusive column bounds `left..right`, and intensity `I` plus RGB bytes. The valid nonempty domain has indices 0 through 47. If top>bottom or left>right, write nothing. Otherwise visit rows increasing, then columns increasing. In every selected cell clear obstruction to zero and assign:

```
intensity = u8(I)
red       = u8(redInput   OR (u32(I) >>> 8))
green     = u8(greenInput OR (u32(I) >>> 16))
blue      = u8(blueInput  OR (u32(I) >>> 24))
```

For intensities 0 through 255 these are exactly the supplied intensity/RGB. The operation overwrites the cell; it does not accumulate it or retain a greater prior intensity. No clipping occurs inside this operation.

**RAL-R06** Initialize the reached viewport background [verified in 1.13c code (static)]. With clear-only true, absent focus player, or absent current room, fill the entire map with intensity 0 and RGB (0,0,0) through RAL-R05, then finish without a room-light request. Otherwise:

1. Request the current room's light with use-area-overrides true and fill all map indices 0..47 with its resulting tuple.
2. Visit the current room's supplied adjacent-room sequence in order. Skip any entry identical to the current room. For every other entry, first request its light with use-area-overrides true, then read its rectangle. Compute signed `left=roomX-originX`, `top=roomY-originY`, `right=left+width`, `bottom=top+height`.
3. Skip that room's fill when left>48, right<0, top>48 or bottom<0. Otherwise clamp left/top upward to 0 and right/bottom downward to 47, and invoke RAL-R05. The upper bound is computed with the supplied width/height directly, without subtracting one. A start of 48 can therefore reach an empty fill after clipping. These operations do not silently convert negative or overflowed bounds to unsigned comparisons.

The final map is complete before the obstruction and source stages in LMR-R12/LMR-R13/LMR-R05. Room lookup, adjacency generation and light-result production beyond verified RAL-R01..RAL-R04 boundaries remain their named input/dependency gates. The exact ordered request log includes an out-of-window room's light request before its rectangle is rejected.

**RAL-R07a** Update Act-IV environment intensity [runtime-confirmed (1.13c)]. For a reached environment-intensity request with act index 3, choose target intensity 128 for source 103, 64 for 104, 56 for 105, 48 for 106, and 16 for every other signed source value. Starting with the stored signed intensity, first increment it by one when it is below the target. Then compare the resulting value with the target and decrement it by one when it is above the target. Finally, if the resulting intensity is zero, replace it with the target. Store that final intensity. The zero test follows the step: initial zero normally becomes one, while initial -1 steps to zero and then becomes the target. This branch ignores clock cycle, ticks, rate and eclipse flag; it performs no colour write, floating calculation or RNG draw. Matching observation covers RAL-V11's source-104 decrement through a reached presentation wrapper; other selectors and zero-order branches remain statically verified.

**RAL-R07b** Produce other changing environment light [unresolved]. The world-clock's reached intensity and colour requests update stored light used by RAL-R04. RAL-R07a establishes the Act-IV intensity branch, and WCK-R05a establishes its two supplied non-Act-IV intensity branches; WCK-R04/WCK-R08b establish the source-120 colour override and request ordering. General normal-day intensity, colour interpolation including Act IV, their float behavior and complete visual integration require RAL-Q03; clock or quest state alone does not specify these missing outputs.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| RAL-R01 | Prepared level fields | runtime-confirmed (1.13c) | none for the measured original profile; generic decoder correctness remains GDA-R02 | RAL-V01, RAL-V09; only V09 observed |
| RAL-R02 | Non-level-8 candidate | verified in 1.13c code (static) | none at supplied state | RAL-V02, RAL-V03 |
| RAL-R03a | Fixed/inactive level-8 candidate | verified in 1.13c code (static) | none at supplied state | RAL-V04 |
| RAL-R03b | Changing level-8 candidate | unresolved | RAL-Q01: changing intensity | none |
| RAL-R03c | Fixed phase lookup | verified in 1.13c code (static) | none for supplied lookup argument; full changing intensity remains RAL-Q01 | RAL-V12 |
| RAL-R04 | Level/environment fallback | verified in 1.13c code (static) | RAL-Q03: produced environment values; supplied stored values available | RAL-V01, RAL-V02, RAL-V04 |
| RAL-R05 | Inclusive fill | verified in 1.13c code (static) | none | RAL-V05, RAL-V06 |
| RAL-R06 | Initial map and ordered neighbors | verified in 1.13c code (static) | RAL-Q01/RAL-Q03: gated light-result integrations only; supplied tuples available | RAL-V07, RAL-V08 |
| RAL-R07a | Act-IV intensity producer | runtime-confirmed (1.13c) | none for supplied presentation request; general source production remains WCK-Q02 | RAL-V10, RAL-V11; only V11 observed |
| RAL-R07b | Other environment production | unresolved | RAL-Q03: remaining general light production | none |

## 5. Outputs and side effects

| Effect | Rule |
|---|---|
| Returned prepared intensity/RGB | RAL-R01 |
| Selected intensity/RGB; ordered level/environment requests | RAL-R02, RAL-R03a, RAL-R03b, RAL-R04 |
| Returned fixed `binary32` lookup value | RAL-R03c |
| Map-cell obstruction zero and replaced intensity/RGB | RAL-R05, RAL-R06 |
| Ordered room-light and rectangle requests | RAL-R06 |
| Stored Act-IV environment intensity | RAL-R07a |
| Other stored environment intensity/RGB, gated producer output | RAL-R07b |

These operations do not change clock fields, quest state, local-source radii or source colours, and emit no network messages.

## 6. Branches and edge cases

The all-zero sentinel tests RGB, never intensity. A single nonzero blue channel is sufficient. Special levels 107 and 108 bypass the normal client override even when their fixed special colour is ineligible. Clear-only takes precedence over a valid focus room. Current-room references in the neighbor list are skipped; later overlapping rooms overwrite earlier ones. Rectangle ends are inclusive, including the extra boundary sample computed from width/height. Original invalid prepared-record reads have no defined successful recovery in this contract. There is no difficulty, PvP or melee/missile branch in the verified supplied-input stages.

## 7. Open questions

**RAL-Q01** [behavior] What exact executable intensity contract covers every active level-8 phase other than -1, including phase-to-argument production, integer wrap, floating-environment preconditions and complete eligibility? RAL-R03c establishes the fixed lookup profile for a supplied argument; that lookup subassertion is closed. The bounded multiplier/conversion evidence is recorded in the companion. Complete the composed intensity contract and distinguishing vectors, then collect a natural transition. RAL-R03b remains gated; RAL-R03a's fixed/inactive branches do not depend on it.

**RAL-Q03** [behavior] What are the exact general normal-day intensity and environment-colour outputs outside RAL-R07a, WCK-R05a and the source-120 override? Complete the floating calculation contract and record actual original-game values, including Act-IV colour. RAL-R04 with explicitly supplied stored environment values and RAL-R06 with supplied light tuples remain available.

**RAL-Q04** [outside-boundary] Does `LvlTypes` influence ambient appearance indirectly through selected tile resources or room generation? Its observed header contains no explicit intensity/RGB column; complete upstream resource influence belongs to the room/tile owner. No local numeric assertion is gated by this question.

**RAL-Q05** [observation] Gameplay confirmation beyond RAL-V09's prepared level-1 read/output replacement and RAL-V11's source-104 Act-IV intensity decrement remains pending. Room override/fallback ordering, background map fills, other Act-IV intensity cases and remaining environment outputs do not inherit those observations. This limits runtime-confirmation claims, not statically verified supplied-input rules. See the [runtime records](../research/spec-evidence/draw-room-and-ambient-lighting.md#runtime-attempt).

## 8. Test vectors

### 8.1 Executable vectors

Each case is fresh, has Seeds: `no draw`, and Draws: none. Unless overridden, all rooms and metadata exist; selected level records are valid. Intensity/RGB tuples are `(I,R,G,B)`. The current act exists. An absent environment returns intensity 0 without changing colour bytes. Supplied `RoomLight(room,true)` results log their calls and return the listed tuple without other side effects; `Rectangle(room)` logs its call and returns the listed bounds. Fill persists all map writes. These supplied operations test RAL-R06 ordering and placement without asserting a gated producer. Unless a case says otherwise, use-area-overrides is false, level identity is 1, normal client override is (0,0,0,0), level light is (7,0,0,0), and environment light is (90,10,20,30). Vectors isolate supplied-input stages except RAL-V11 and its explicitly named reached wrapper.

- **RAL-V01** Prepared field selection and fallback (RAL-R01, RAL-R04; boundary: supplied-input stage)
  - Inputs: room level index 1; ready records at index 0 have (9,99,0,0), index 1 has (0,0,0,7). Environment (90,10,20,30).
  - Seeds: no draw.
  - Draws: none.
  - Trace: select index 1; blue 7 is nonzero; retain the complete level tuple.
  - Outputs: (0,0,0,7); no environment request.
  - Discriminates: intensity-based fallback would return (90,10,20,30); incorrectly selecting index 0 would return (9,99,0,0).
  - runtime_observed: no

- **RAL-V02** RGB sentinel overrides candidate intensity (RAL-R02, RAL-R04; boundary: supplied-input stage)
  - Inputs: use-area-overrides true, level 1, client override (200,0,0,0); level (33,0,0,0); environment (90,10,20,30).
  - Seeds: no draw.
  - Draws: none.
  - Trace: override RGB all zero -> level; level RGB all zero -> environment RGB then intensity.
  - Outputs: (90,10,20,30); ordered requests level, environment RGB, environment intensity.
  - Discriminates: treating either positive intensity as final yields 200 or 33.
  - (a) Fresh Inputs: client override (0,0,2,0), all other inputs as above. Outputs: (0,0,2,0), no level/environment request.
  - (b) Fresh Inputs: use-area-overrides false, client override (200,7,8,9), level (33,0,0,0), environment absent. Outputs: (0,0,0,0); requests level, environment RGB, environment intensity. Rejects consulting the override with the condition false or retaining level intensity 33.
  - runtime_observed: no

- **RAL-V03** Special area selection (RAL-R02, RAL-R04; boundary: supplied-input stage)
  - Inputs: use-area-overrides true; level 107; enable true; special phase -1; client override (22,1,2,3); level (55,0,0,9).
  - Seeds: no draw.
  - Draws: none.
  - Trace: special enabled/sentinel branch returns fixed candidate, bypassing both normal override and level.
  - Outputs: (160,255,64,48), no level/environment request.
  - Discriminates: using the client override returns (22,1,2,3).
  - (a) Fresh Inputs: level 108, enable false, phase -1, same client override/level. Outputs: (55,0,0,9), one level request. Rejects using normal override on special ineligibility.
  - (b) Fresh Inputs: level 107, enable true, phase 0, same client override/level. Outputs: (55,0,0,9), one level request.
  - (c) Fresh Inputs: use-area-overrides true, level 108, enable true, special phase -1, client override (22,1,2,3), level (55,0,0,9). Trace: the second special level uses the same enabled sentinel branch. Outputs: (160,255,64,48), no level/environment request. Draws: none.
  - runtime_observed: no

- **RAL-V04** Level-8 fixed and suppressed branches (RAL-R03a, RAL-R04; boundary: supplied-input stage)
  - Inputs: use-area-overrides true, level 8; suppression false, initialization true, active true, phase -1; level (23,1,2,3).
  - Seeds: no draw.
  - Draws: none.
  - Trace: active sentinel branch selects the fixed level-8 light.
  - Outputs: (80,255,64,48), no level/environment request.
  - Discriminates: using the other special-area intensity gives 160.
  - (a) Fresh Inputs: suppression true, initialization false, active true, phase -1; same level tuple. Outputs: (23,1,2,3), one level request; no initialization requirement is reached.
  - (b) Fresh Inputs: suppression false, initialization true, active false, phase 5; same level tuple. Outputs: (23,1,2,3), one level request; the changing-intensity branch is not reached.
  - (c) Fresh Inputs: use-area-overrides false, level 8; suppression false, initialization false, active true, phase -1; level (23,1,2,3). Trace: disabling area overrides bypasses the level-8 helper and its initialization requirement. Outputs: (23,1,2,3), one level request, no environment request or initialization assertion. Draws: none.
  - runtime_observed: no

- **RAL-V05** Inclusive replacement (RAL-R05; boundary: supplied-input stage)
  - Inputs: rows 1..2, columns 2..3; tuple (17,4,5,6); every map cell starts obstruction 1 and tuple (200,9,8,7).
  - Seeds: no draw.
  - Draws: none.
  - Trace: visit (2,1),(3,1),(2,2),(3,2), replacing each.
  - Outputs: those four cells obstruction 0, tuple (17,4,5,6); every other cell unchanged.
  - Discriminates: exclusive ends fill only one cell; accumulation retains a greater old intensity.
  - (a) Fresh Inputs: top 2,bottom 1,left 2,right 3; same tuple and initial map. Outputs: every cell unchanged.
  - (b) Fresh Inputs: top 1,bottom 2,left 3,right 2; same tuple and initial map. Outputs: every cell unchanged.
  - runtime_observed: no

- **RAL-V06** Intensity high bits affect stored channels (RAL-R05; boundary: supplied-input stage)
  - Inputs: single cell rows 0..0,columns 0..0; intensity 256, RGB (2,4,8); that cell initially obstruction 1 and tuple (9,9,9,9).
  - Seeds: no draw.
  - Draws: none.
  - Trace: low intensity byte is 0; red is 2 OR 1 = 3; other shifted intensity bytes are zero.
  - Outputs: cell (0,0) obstruction 0, tuple (0,3,4,8); all other cells unchanged.
  - Discriminates: separately narrowing intensity before combining channels yields red 2.
  - (a) Fresh Inputs: same single-cell bounds and initial cell, intensity 50462976 (`0x03020100`), RGB (2,4,8). Seeds: no draw. Draws: none. Trace: low intensity byte 0; shifted low bytes 1,2,3 merge into red, green, blue respectively, giving 2 OR 1 = 3, 4 OR 2 = 6, 8 OR 3 = 11. Outputs: cell (0,0) obstruction 0, tuple (0,3,6,11); all other cells unchanged. Discriminates: omitting the green/blue merges leaves 4/8; swapping their shifted inputs yields 7/10. Both fail.
  - runtime_observed: no

- **RAL-V07** Ordered clipped adjacent rooms (RAL-R05, RAL-R06; boundary: supplied-input stage)
  - Inputs: clear-only false; focus/current room A; origin (100,200); adjacent sequence [A,B,C,D]; supplied RoomLight(A,true)=(10,1,2,3), B=(20,4,5,6), C=(30,7,8,9), D=(40,10,11,12). Rectangles B=(99,200,2,1), C=(101,201,0,0), D=(149,200,0,0). All cells initially obstruction 1 and tuple (99,9,9,9).
  - Seeds: no draw.
  - Draws: none.
  - Trace: initialize all cells from A; skip A in sequence; B gives inclusive columns -1..1 and rows 0..1, clipped to columns 0..1; C selects only (1,1); D left 49 rejects after its light and rectangle requests.
  - Outputs: cell (1,1) tuple (30,7,8,9); (0,0),(1,0),(0,1) tuple (20,4,5,6); all remaining cells tuple (10,1,2,3); every obstruction 0. Exact request log: RoomLight(A,true), RoomLight(B,true), Rectangle(B), RoomLight(C,true), Rectangle(C), RoomLight(D,true), Rectangle(D).
  - Discriminates: width/height minus one loses B's last row/column and C's single cell; reversed neighbor order leaves (1,1) at B's light; clipping before resolving D omits required requests.
  - runtime_observed: no

- **RAL-V08** Clear paths and exact outer edge (RAL-R05, RAL-R06; boundary: supplied-input stage)
  - Inputs: clear-only true; focus/current room A present; supplied A light (90,10,20,30); every cell initially obstruction 1 and tuple (99,9,9,9).
  - Seeds: no draw.
  - Draws: none.
  - Trace: clear-only selects zero fill before any room-light request.
  - Outputs: all 2304 cells obstruction 0 and tuple (0,0,0,0); empty request log.
  - Discriminates: consulting room A would leave its light or a request in the log.
  - (a) Fresh Inputs: clear-only false, focus absent; same initial map. Outputs: same zero map and empty log.
  - (b) Fresh Inputs: clear-only false, focus present, current room absent; same initial map. Outputs: same zero map and empty log.
  - (c) Fresh Inputs: clear-only false, current room A, origin (100,200), adjacent [B], supplied A light (90,10,20,30), B light (7,1,1,1), Rectangle(B)=(148,200,0,0). Outputs: every cell obstruction 0 and tuple (90,10,20,30). Log RoomLight(A,true), RoomLight(B,true), Rectangle(B). Clipped columns 48..47 form an empty fill; no cell 47 write occurs.
  - (d) Fresh Inputs: clear-only false, focus/current room A, origin (100,200), adjacent [B], supplied A light (90,10,20,30), B light (7,1,1,1), Rectangle(B)=(147,247,2,2); every cell initially obstruction 1 and tuple (99,9,9,9). Trace: starts are (47,47), ends (49,49), clipped to inclusive ends (47,47). Outputs: cell (47,47) tuple (7,1,1,1); all other cells tuple (90,10,20,30); every obstruction 0. Exact log RoomLight(A,true), RoomLight(B,true), Rectangle(B). Draws: none. Discriminates: rejecting starts at 47, clipping ends to 46 or using exclusive ends loses the last-cell overwrite; leaving ends at 49 exceeds the supplied map.
  - runtime_observed: no

- **RAL-V09** Original-profile zero level tuple replaces existing outputs (RAL-R01; boundary: supplied-input level reader)
  - Inputs: present room and level metadata with level index 1; prepared level collection count 137; selected record's intensity/RGB `(0,0,0,0)`; caller output tuple initially `(0,32,0,177)`. Other prepared records are unread and arbitrary; no override or environment operation belongs to this reader.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: index 1 is in range; read its four unsigned bytes and replace the four output values with zero.
  - Outputs: returned intensity/RGB `(0,0,0,0)`; caller outputs become `(0,0,0,0)`; room metadata and prepared table unchanged.
  - Discriminates: treating an all-zero tuple as a no-write request retains red 32 and blue 177; substituting a white/environment tuple produces nonzero channels. Both are rejected.
  - runtime_observed: yes (1.13c, 2026-10-29, [RT-RAL-LEVEL-READ](../research/spec-evidence/draw-room-and-ambient-lighting.md#runtime-confirmation-2026-10-29))

- **RAL-V10** Act-IV intensity source selection, step and zero replacement (RAL-R07a; boundary: supplied presentation request)
  - Inputs: each row is a fresh request with act index 3, the listed signed source and prior stored intensity, and stored RGB (1,2,3). Clock cycle 2, ticks 0, rate 128 and flag 0 are present but unused by this branch.
  - Seeds: no draw.
  - Draws: none.
  - Trace: select the target from the source, step the signed intensity toward it, then replace a resulting zero with the target.

  | Source | Prior intensity | Target | After step | Final intensity |
  |---|---|---|---|---|
  | 103 | -1 | 128 | 0 | 128 |
  | 104 | -1 | 64 | 0 | 64 |
  | 105 | -1 | 56 | 0 | 56 |
  | 106 | -1 | 48 | 0 | 48 |
  | 107 | -1 | 16 | 0 | 16 |
  | 102 | -1 | 16 | 0 | 16 |
  | -1 | -1 | 16 | 0 | 16 |
  | -2147483648 | -1 | 16 | 0 | 16 |
  | 103 | 0 | 128 | 1 | 1 |
  | 103 | 128 | 128 | 128 | 128 |
  | 104 | 65 | 64 | 64 | 64 |
  | 106 | 47 | 48 | 48 | 48 |
  | 105 | -2147483648 | 56 | -2147483647 | -2147483647 |
  | 105 | 2147483647 | 56 | 2147483646 | 2147483646 |

  - Outputs: stored intensity is each row's Final intensity; RGB remains (1,2,3); clock fields are unchanged. Fresh cases with source 103/intensity 4 produce intensity 5 both with the default clock above and with cycle 5/ticks 960/rate 4/flag 2; RGB/clock remain unchanged.
  - Discriminates: testing prior zero before the step yields 128 instead of 1; omitting the post-step zero replacement retains 0 instead of each target. A fixed target, shifted source range, unsigned intensity comparison, unconditional decrement or applying the eclipse floor fails a listed case. Assigning the target immediately fails the 0,65,47 and signed-extreme cases. The flag-2 case rejects consulting the non-Act-IV eclipse branch first.
  - runtime_observed: no

- **RAL-V11** Original-profile Act-IV decrement through the presentation wrapper (RAL-R07a; boundary: reached WCK-R04 presentation wrapper)
  - Inputs: present act/environment; source 104 with derived act index 3; stored intensity 128 and RGB (255,194,27); clock cycle 2, period 0, ticks 17104, rate 128 and flag 0. Other clock fields are valid, retained and unused by the asserted operations. The presentation stage includes the owning WCK-R04 clock advance and the subsequent colour request; the latter's RGB result remains RAL-R07b and is not asserted.
  - Seeds: no draw.
  - Draws: none.
  - Trace: WCK-R04 advances this Act-IV clock by 16 without transition, giving ticks 17120 and period 0. RAL-R07a selects target 64 from source 104, decrements intensity 128 to 127, then retains this nonzero result. The subsequent colour operation writes no intensity; its colour result is outside the verified assertions.
  - Outputs: stored intensity 127; clock ticks 17120, cycle 2 and period 0; wrapper period-change result 0. No RGB value or map/pixel result is asserted. Clock writes belong to WCK-R04; the intensity assertion belongs to RAL-R07a.
  - Discriminates: assigning target 64 immediately, retaining prior intensity 128, selecting the source-103 target, or applying a normal-day/eclipsed non-Act-IV branch fails intensity 127. Advancing the clock by one or returning a nonzero period-change result fails the dependency assertions.
  - runtime_observed: yes (1.13c, 2026-10-30, [Act-IV transition record](../research/spec-evidence/draw-room-and-ambient-lighting.md#act-iv-transition-confirmation-2026-10-30)); only the specified intensity/clock assertions, not colour production.

- **RAL-V12** Exact fixed lookup values and wrapped argument selection (RAL-R03c; boundary: supplied lookup operation)
  - Inputs: each row is a fresh lookup with the listed signed 32-bit argument and the fixed table defined in section 3. No phase state, clock, environment or changing-intensity operation is supplied.
  - Seeds: no draw.
  - Draws: none.
  - Trace: add 128 with 32-bit wrap, mask by 511, and select the exact constant. The output column gives the IEEE 754 `binary32` encoding of the returned mathematical value, independent of memory byte order.

  | Argument | Index | Returned `binary32` encoding |
  |---|---|---|
  | -128 | 0 | `0x00000000` |
  | -123 | 5 | `0x3D7B2B74` |
  | -121 | 7 | `0x3DAFB681` |
  | 0 | 128 | `0x3F800000` |
  | 128 | 256 | `0xB3BBBD2E` |
  | 384 | 0 | `0x00000000` |
  | 2147483647 | 127 | `0x3F7FFB11` |
  | -2147483648 | 128 | `0x3F800000` |

  - Outputs: only the listed lookup value; all state unchanged. No intensity, eligibility or gameplay assertion is made.
  - Discriminates: rounding the angle to `binary32` before sine changes index 5; substituting an exact-pi equally spaced sine changes index 7 and the half-turn result. Omitting the argument bias or masking fails the selected indices, including argument 384; nonwrapping signed-overflow rejection fails the maximum signed argument.
  - runtime_observed: no

### 8.2 Observation vectors

**RAL-O01** Observe one original-game rebuild with a complete current-room/neighbor input snapshot, selected light tuples and final pre-obstruction map. Expect the exact RAL-R06 ordering and RAL-R05 cell values for the captured inputs. The [reproduction procedure](../research/spec-evidence/draw-room-and-ambient-lighting.md#ral-o01) distinguishes resolver output from the later local-light result.

**RAL-O02** Observe a level-light tuple with zero RGB switching to stored environment light, and a nonzero RGB tuple bypassing that fallback. Expect RAL-R04 with the actually captured fields. The [procedure](../research/spec-evidence/draw-room-and-ambient-lighting.md#ral-o02) requires compiled ready-record inputs rather than text assumptions.

**RAL-O03** Observe a prepared level-1 read that replaces the caller's output tuple, including any initially nonzero channels. RAL-V09 supplies the matched case; the [procedure](../research/spec-evidence/draw-room-and-ambient-lighting.md#ral-o03) records the selected record and all four output values independently of later fallback.

## 9. Acceptance and dependencies

RAL-V01..RAL-V12 establish the verified prepared-input subset RAL-R01/RAL-R02/RAL-R03a/RAL-R03c/RAL-R04/RAL-R05/RAL-R06/RAL-R07a, with the input-domain and supplied-operation boundaries stated above. The companion establishes original-profile text/compiled/ready correspondence for all selected level-light fields; this is data correspondence, not execution of every room or fallback branch. Full original behavior additionally requires RAL-Q01 and RAL-Q03's gated outputs and the pending gameplay confirmations, including RAL-O01/RAL-O02. A pass with supplied light tuples or a supplied presentation request is partial acceptance, not end-to-end quest-to-pixel reproduction.

**Depends on:** GDA-R03a/GDA-R03b/GDA-R03c for named direct byte fields and ASCII header matching; GDA-R02 retains generic compiled-decoder correctness beyond the measured original-profile correspondence. WCK-R02/WCK-R03 supply the core clock advance and transition used by RAL-V11. WCK-R04/WCK-R05a/WCK-R08b supply the established world-clock intensity branches, source mapping, presentation ordering and source-120 override; WCK-Q01 retains the unverified general environment producer. LMR-R11 supplies reached rebuild eligibility. Tainted Sun's quest start/reset operations reach the world-clock producer; no direct quest-state input or quest progression rule is added here.

**Depended on by:** LMR-R13 and the later local-source accumulation consume initialized map cells and obstruction preparation order; their numeric supplied-input contracts remain independently usable.

## 10. Related specs

- [Light map and radius](draw-light-map-and-light-radius.md): local sources, collision, shadowing and accumulation follow base initialization.
- [World clock](drlg-world-clock.md): owns clock progression and bounded eclipse intensity effects.
- [Tainted Sun](quest-tainted-sun.md): owns quest-stage requests, not the renderer's room selection.
- [Game-data records](core-game-data-table-records.md): owns preparation primitives and compiled correspondence gates.

## Appendix A. Provenance

Original x86 LoD Expansion 1.13c D2Client, D2Common and Fog images correspond to LOD113C-D2CLIENT, LOD113C-D2COMMON and LOD113C-FOG in the [source manifest](../research/source-manifest.md). The 2026-10-29 and 2026-10-30 readings used matched reconstruction where available, otherwise instruction views corroborated against original-image Ghidra listings. Prior 2026-10-27 light-map dispatch evidence is reused with its original date and scope. Effective archive reads and the complete 2026-10-30 comparison establish original-profile text/compiled/ready correspondence in the [companion](../research/spec-evidence/draw-room-and-ambient-lighting.md). RAL-V09 confirms one prepared level-1 read/output replacement on 2026-10-29; RAL-V11 confirms the reached source-104 Act-IV intensity decrement and its specified clock/return assertions on 2026-10-30. Other producer/resolver observations remain partial. Wider gameplay confirmation remains pending. Status remains draft with the exact gates above.
