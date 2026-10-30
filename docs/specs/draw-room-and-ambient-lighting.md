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
| Level-8 suppression / context / active condition / phase | Supplied level-8 client state | Boolean; present context; Boolean; signed 32-bit integer | RAL-R03; a nonsuppressed successful invocation requires context |
| Act and environment | Supplied current client act; optional environment | Present act; environment optional | Stored environment fallback; absent environment supplies intensity 0 and leaves colour bytes unchanged |
| Environment light | Stored act environment | Signed 32-bit intensity; three unsigned bytes | Produced through the world-clock presentation boundary; general production remains RAL-Q03 |
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
| Prepared level-field names | Table schema and consumer code | Named fields selected from prepared level-definition records; arbitrary compiled correspondence remains GDA-R02 |

No claim assigns a lighting column to `LvlTypes`. The [effective-header observation](../research/spec-evidence/draw-room-and-ambient-lighting.md#table-data-records) records its actual columns; downstream influence through tile resources is RAL-Q04.

## 4. Rules (execution order)

All verified operations below draw no RNG. Their integer operations use 32-bit wrap unless byte narrowing is specified.

**RAL-R01** Read prepared level light [runtime-confirmed (1.13c)]. For a supplied room with valid level metadata and an in-range prepared record, return its unsigned `Intensity`, `Red`, `Green`, `Blue` bytes, replacing all four caller output values even when the selected tuple is all zero. The level identity selects the record by collection index, not by searching a stored identifier. Missing room metadata or an invalid selected record has no successful fallback contract here; it is outside the valid prepared-input domain. Matching observation covers only RAL-V09's original-profile level-1 zero tuple and output replacement; other prepared values remain statically verified.

**RAL-R02** Select a non-level-8 override candidate [verified in 1.13c code (static)]. Start candidate RGB at (0,0,0). If use-area-overrides is false, go directly to RAL-R04's level fallback. Otherwise, for levels 107 and 108, use intensity 160 and RGB (255,64,48) only when the special-area enable condition is true and its phase equals -1. If either condition fails, use candidate intensity 0 and RGB (0,0,0); do not substitute the normal client override. For any other level except 8, copy the client override intensity and RGB. Any candidate with at least one nonzero RGB byte is final, even when intensity is zero. An all-zero RGB candidate selects RAL-R04, regardless of its intensity. Level 8 instead uses RAL-R03.

**RAL-R03a** Select the fixed or inactive level-8 candidate [verified in 1.13c code (static)]. With use-area-overrides true and level identity 8, start RGB at (0,0,0). A true suppression condition returns candidate intensity 0 without inspecting context or active condition. Otherwise require a present context. An inactive condition returns intensity 0 and RGB (0,0,0). An active condition with phase -1 returns intensity 80 and RGB (255,64,48). Nonzero RGB finishes selection; zero RGB selects RAL-R04.

**RAL-R03b** Select a changing level-8 candidate [unresolved]. With nonsuppressed, valid-context, active level-8 state and phase other than -1, the operation supplies RGB (255,64,48) and a phase-dependent intensity. The exact intensity for this branch requires RAL-Q01. It remains gated; no default intensity is specified.

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

**RAL-R07** Produce changing environment light [unresolved]. The world-clock's reached intensity and colour requests update stored light used by RAL-R04. WCK-R05a establishes its two supplied non-Act-IV intensity branches; WCK-R04/WCK-R08b establish the source-120 colour override and request ordering. General normal-day, Act-IV and colour-interpolation values, their float behavior and complete visual integration require RAL-Q03; clock or quest state alone does not specify these missing outputs.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| RAL-R01 | Prepared level fields | runtime-confirmed (1.13c) | RAL-Q02: arbitrary compiled correspondence beyond the named observation only | RAL-V01, RAL-V09; only V09 observed |
| RAL-R02 | Non-level-8 candidate | verified in 1.13c code (static) | none at supplied state | RAL-V02, RAL-V03 |
| RAL-R03a | Fixed/inactive level-8 candidate | verified in 1.13c code (static) | none at supplied state | RAL-V04 |
| RAL-R03b | Changing level-8 candidate | unresolved | RAL-Q01: changing intensity | none |
| RAL-R04 | Level/environment fallback | verified in 1.13c code (static) | RAL-Q03: produced environment values; supplied stored values available | RAL-V01, RAL-V02, RAL-V04 |
| RAL-R05 | Inclusive fill | verified in 1.13c code (static) | none | RAL-V05, RAL-V06 |
| RAL-R06 | Initial map and ordered neighbors | verified in 1.13c code (static) | RAL-Q01/RAL-Q03: gated light-result integrations only; supplied tuples available | RAL-V07, RAL-V08 |
| RAL-R07 | Environment producer | unresolved | RAL-Q03: general light production | none |

## 5. Outputs and side effects

| Effect | Rule |
|---|---|
| Returned prepared intensity/RGB | RAL-R01 |
| Selected intensity/RGB; ordered level/environment requests | RAL-R02, RAL-R03a, RAL-R03b, RAL-R04 |
| Map-cell obstruction zero and replaced intensity/RGB | RAL-R05, RAL-R06 |
| Ordered room-light and rectangle requests | RAL-R06 |
| Stored environment intensity/RGB, gated producer output | RAL-R07 |

These operations do not change clock fields, quest state, local-source radii or source colours, and emit no network messages.

## 6. Branches and edge cases

The all-zero sentinel tests RGB, never intensity. A single nonzero blue channel is sufficient. Special levels 107 and 108 bypass the normal client override even when their fixed special colour is ineligible. Clear-only takes precedence over a valid focus room. Current-room references in the neighbor list are skipped; later overlapping rooms overwrite earlier ones. Rectangle ends are inclusive, including the extra boundary sample computed from width/height. Original invalid prepared-record reads have no defined successful recovery in this contract. There is no difficulty, PvP or melee/missile branch in the verified supplied-input stages.

## 7. Open questions

**RAL-Q01** [behavior] What exact intensity does active level-8 phase other than -1 produce, including integer wrap, cosine-table selection, floating conversion and phase eligibility? Complete the helper and shared math dependency reading, then collect a natural transition. RAL-R03b remains gated; RAL-R03a's fixed/inactive branches do not depend on it.

**RAL-Q02** [integration] Which effective compiled level-definition values beyond RAL-V09's four named level-1 cells correspond to the text and ready fields in the selected original profile? Loader descriptors and the named observation are established; arbitrary compiled-resource correspondence remains GDA-R02. Capture other ready records and effective compiled resources without treating text equality as execution evidence.

**RAL-Q03** [behavior] What are the exact general normal-day, Act-IV and environment-colour outputs outside WCK-R05a and the source-120 override? Complete the float-helper/table path and record actual original-game values. RAL-R04 with explicitly supplied stored environment values and RAL-R06 with supplied light tuples remain available.

**RAL-Q04** [outside-boundary] Does `LvlTypes` influence ambient appearance indirectly through selected tile resources or room generation? Its observed header contains no explicit intensity/RGB column; complete upstream resource influence belongs to the room/tile owner. No local numeric assertion is gated by this question.

**RAL-Q05** [observation] Gameplay confirmation beyond RAL-V09's prepared level-1 read and output replacement remains pending. Room override/fallback ordering, background map fills and changing environment outputs do not inherit that observation. This limits runtime-confirmation claims, not statically verified supplied-input rules. See the [runtime records](../research/spec-evidence/draw-room-and-ambient-lighting.md#runtime-attempt).

## 8. Test vectors

### 8.1 Executable vectors

Each case is fresh, has Seeds: `no draw`, and Draws: none. Unless overridden, all rooms and metadata exist; selected level records are valid. Intensity/RGB tuples are `(I,R,G,B)`. The current act exists. An absent environment returns intensity 0 without changing colour bytes. Supplied `RoomLight(room,true)` results log their calls and return the listed tuple without other side effects; `Rectangle(room)` logs its call and returns the listed bounds. Fill persists all map writes. These supplied operations test RAL-R06 ordering and placement without asserting a gated producer. Unless a case says otherwise, use-area-overrides is false, level identity is 1, normal client override is (0,0,0,0), level light is (7,0,0,0), and environment light is (90,10,20,30). All vectors are supplied-input stages.

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
  - Inputs: use-area-overrides true, level 8; suppression false, context present, active true, phase -1; level (23,1,2,3).
  - Seeds: no draw.
  - Draws: none.
  - Trace: active sentinel branch selects the fixed level-8 light.
  - Outputs: (80,255,64,48), no level/environment request.
  - Discriminates: using the other special-area intensity gives 160.
  - (a) Fresh Inputs: suppression true, context absent, active true, phase -1; same level tuple. Outputs: (23,1,2,3), one level request; no context requirement is reached.
  - (b) Fresh Inputs: suppression false, context present, active false, phase 5; same level tuple. Outputs: (23,1,2,3), one level request; the changing-intensity branch is not reached.
  - (c) Fresh Inputs: use-area-overrides false, level 8; suppression false, context absent, active true, phase -1; level (23,1,2,3). Trace: disabling area overrides bypasses the level-8 helper and its context requirement. Outputs: (23,1,2,3), one level request, no environment request or context assertion. Draws: none.
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

### 8.2 Observation vectors

**RAL-O01** Observe one original-game rebuild with a complete current-room/neighbor input snapshot, selected light tuples and final pre-obstruction map. Expect the exact RAL-R06 ordering and RAL-R05 cell values for the captured inputs. The [reproduction procedure](../research/spec-evidence/draw-room-and-ambient-lighting.md#ral-o01) distinguishes resolver output from the later local-light result.

**RAL-O02** Observe a level-light tuple with zero RGB switching to stored environment light, and a nonzero RGB tuple bypassing that fallback. Expect RAL-R04 with the actually captured fields. The [procedure](../research/spec-evidence/draw-room-and-ambient-lighting.md#ral-o02) requires compiled ready-record inputs rather than text assumptions.

**RAL-O03** Observe a prepared level-1 read that replaces the caller's output tuple, including any initially nonzero channels. RAL-V09 supplies the matched case; the [procedure](../research/spec-evidence/draw-room-and-ambient-lighting.md#ral-o03) records the selected record and all four output values independently of later fallback.

## 9. Acceptance and dependencies

RAL-V01..RAL-V09 establish the verified prepared-input subset RAL-R01/RAL-R02/RAL-R03a/RAL-R04/RAL-R05/RAL-R06, with the input-domain and supplied-operation boundaries stated above. Full original behavior additionally requires RAL-Q01 and RAL-Q03's gated outputs, RAL-Q02's effective compiled correspondence and RAL-O01/RAL-O02 gameplay confirmation. A pass with supplied light tuples is partial acceptance, not end-to-end quest-to-pixel reproduction.

**Depends on:** GDA-R03a/GDA-R03b/GDA-R03c for named direct byte fields and ASCII header matching; GDA-R02 retains arbitrary compiled correspondence. WCK-R04/WCK-R05a/WCK-R08b supply the established world-clock intensity branches, source mapping, presentation ordering and source-120 override; WCK-Q01 retains the unverified general environment producer. LMR-R11 supplies reached rebuild eligibility. Tainted Sun's quest start/reset operations reach the world-clock producer; no direct quest-state input or quest progression rule is added here.

**Depended on by:** LMR-R13 and the later local-source accumulation consume initialized map cells and obstruction preparation order; their numeric supplied-input contracts remain independently usable.

## 10. Related specs

- [Light map and radius](draw-light-map-and-light-radius.md): local sources, collision, shadowing and accumulation follow base initialization.
- [World clock](drlg-world-clock.md): owns clock progression and bounded eclipse intensity effects.
- [Tainted Sun](quest-tainted-sun.md): owns quest-stage requests, not the renderer's room selection.
- [Game-data records](core-game-data-table-records.md): owns preparation primitives and compiled correspondence gates.

## Appendix A. Provenance

Original x86 LoD Expansion 1.13c D2Client and D2Common images correspond to LOD113C-D2CLIENT and LOD113C-D2COMMON in the [source manifest](../research/source-manifest.md). The 2026-10-29 reading used matched reconstruction where available, otherwise instruction views corroborated against original-image Ghidra listings. Prior 2026-10-27 light-map dispatch evidence is reused with its original date and scope. Effective archive reads establish only the named text cells/header facts recorded in the [companion](../research/spec-evidence/draw-room-and-ambient-lighting.md); RAL-V09 confirms one prepared level-1 read and output replacement on 2026-10-29; all wider gameplay observations remain pending. Status remains draft with the exact gates above and user review pending.
