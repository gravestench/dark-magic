# Light map and light radius

**Description**

Light sources contribute brightness and color to the area around them. Their reach can change immediately or grow and shrink over successive lighting updates. Nearby obstructions can cast shadows, and overlapping lights combine their brightness and colors. The rules below specify those computations with prepared source and world inputs; complete source eligibility and the final displayed scene require the separately named producers and graphics consumers.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md), [core-record-layouts-and-ids](core-record-layouts-and-ids.md), [core-game-data-table-records](core-game-data-table-records.md)
Ids: LMR-R, LMR-Q, LMR-V, LMR-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/draw-light-map-and-light-radius.md)
Open evidence gaps: [gap file](../plans/gaps/draw-light-map-and-light-radius.md)

## 1. Scope

This contract covers client light-source creation, radius and color changes, and the light-map update that consumes those sources in Expansion mode. The established supplied-input operations below do not establish every producer's eligibility, record binding, or final rendered pixel. Room and level ambient-light selection, palette shifts, screen fades, and quest transitions are outside this boundary; [world clock](drlg-world-clock.md) and [Tainted Sun](quest-tainted-sun.md) own the cited environmental inputs.

## 2. Trigger and ownership

D2Client owns these computations. The viewport-render path rebuilds the light map only when its unsigned lighting-update counter exceeds its last-processed counter, recording the new counter before processing sources. The source traversal follows the current source sequence; source creation inserts at its front. Each completed client-world tick advances the current counter under LMR-R11. This is a client lighting update, not a promise of one radius step per server frame or per displayed frame; wider scheduler eligibility remains LMR-Q01.

Object initialization and mode changes feed object-light requests; LMR-R04 specifies the selected prepared object's per-mode light and color fields. A registered player stat-change callback feeds radius and color updates. Complete source eligibility, other authored producer bindings and upstream stat aggregation remain LMR-Q02. Supplied-input radius operations are independently usable under their stated preconditions.

A reached skill-action light stage can attach a temporary light to the acting unit under LMR-R17. A reached projectile-creation light stage instead manages that projectile's primary light under LMR-R18. LMR-R19 specifies explicit temporary-light clearing and the light portion of unit cleanup. These are separate associations; choosing when gameplay reaches those stages remains a producer responsibility.

## 3. Inputs and data

Arithmetic uses [numeric foundations](core-numeric-and-rng-foundations.md). Unless specified otherwise, an integer operation has 32-bit width, addition wraps, comparisons are signed, and a right shift is arithmetic. `u16(value)` takes the low 16 bits as unsigned; `u8(value)` takes the low eight bits. Source state in this document is semantic state, not a required storage layout.

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| Global source sequence | Existing client lighting state | Ordered sequence of distinct live sources | Creation inserts at its front; updates traverse it in order; removal unlinks the selected entry |
| Allocation outcome | Supplied successful source/cache allocation | Fresh source identity and sufficient required storage | Successful-allocation precondition for the specified operations; generic failure is LMR-Q03 |
| Game-world initialization and viewport-creation events | Supplied reached initialization stages | Ordered events | Reset the respective counters under LMR-R11 |
| Requested radius | Light creation or radius-change request | Signed 32-bit integer | Radius before conversion to lighting coordinates |
| Object light request | Object initialization or mode-change request | Unsigned 16-bit integer | Twice the requested source radius; LMR-R04 |
| Object/source association | Existing object light state | Associated source or absent | Determines reuse, creation, or removal under LMR-R04 |
| Prepared object light fields and selected object mode | Ready `Objects` record's `Lit0` through `Lit7`, `Red`, `Green`, `Blue`; supplied selected mode 0 through 7 | Unsigned byte fields; integer mode | Produces the wrapper's request and initial RGB; selecting mode eligibility and proving an arbitrary compiled record's authored source remain separate |
| Source kind | Original client source selector | Signed 32-bit integer | Kind 0 can use viewport shadows, kind 2 uses its cached shadows, other cases use the uncached path |
| Current radius and target radius | Existing source state | Signed 32-bit, eight lighting-coordinate units per requested-radius unit | Reach used now and requested future reach |
| Source center | Existing source state | Signed 32-bit x and y in lighting-coordinate units | Center consumed by the light-map stage |
| Unit position | Supplied unit's current position | 32-bit fixed-point x and y, 16 fractional bits per subtile | Converted into lighting coordinates during allocation/update |
| Source intensity and RGB | Creation/color-change request | Four unsigned bytes | Intensity byte and red, green, blue channel bytes |
| Source removal condition | Existing source state | Boolean | Selects radius-expiry removal and suppresses unit-position refresh |
| Unit identity and lookup result | Supplied source owner and current unit registry | Unit category, identity, Boolean alternate-registry condition; present/absent result | Determines whether a moving source's center is refreshed; category 6 is the code-fixed ownerless sentinel |
| Lighting quality selector | Supplied selection for this update | Signed 32-bit integer | Selects the kind-0 rendering path |
| Radius modifier | Stat 89 `item_lightradius`, player or monster stat-change callback's new value | Signed 32-bit integer | Added to the player's fixed base or the monster's recomputed base |
| Packed color value | Stat 90 `item_lightcolor`, player/monster stat-change callback's new value | 32-bit integer | Zero selects white; otherwise three low bytes supply channels |
| Update counter and last-processed counter | Viewport lighting state | Unsigned 32-bit integers | Rebuild eligibility |
| Selected sample position | Supplied candidate sample for the uncached radial stage | Signed 32-bit x/y in lighting-coordinate units | The sample at which to compute a contribution |
| Light-map origin | Supplied local map window | Signed 32-bit x/y in subtiles | Selects a 48 by 48 sample window |
| Existing map cell | Supplied local map state | Unsigned intensity/red/green/blue bytes | Accumulated local lighting before this contribution |
| Colored-light condition | Current client rendering condition | Boolean | Enables weighted RGB or forces zero RGB |
| Local-player identity | Current client focus player, or absent | Identity | Low-quality uncached player-radius comparison; absent identity compares as -1 |
| Completed client-tick event and reached viewport-render event | Supplied completed simulation stage and eligible viewport stage | Ordered events | Increment current update counter or test rebuild eligibility; wider scheduler eligibility is outside this contract |
| Room collision context | Supplied current room, its ordered adjacent-room sequence and each room's subtile bounds | Logical room references, signed 32-bit origins/extents | Selects the room containing a queried subtile |
| Collision grids | Supplied room collision storage or absent | Signed 32-bit origins/width; row-major unsigned 16-bit cells | Supplies already materialized collision bits; grid bounds are consistent with their room's eligible query coordinates |
| Viewport obstruction flags | Supplied initialized light-map cells | Unsigned byte, valid values 0 or 1 | Records whether that viewport sample blocks the light query |
| Shadow scratch state | Two persistent logical grids, blocker B and propagated P | Signed 32-bit values at logical offsets -32 through 31 on each axis | Shared working values; successful shadow stages require P in 0 through 16 and produce B values 0 or 16 |
| Source shadow cache | Existing source-local cache | Present/absent storage; valid flag; row-major signed 32-bit samples | A valid cache supplies one sample per source raster position |
| Owner integer position and current room | Resolved source owner | Unsigned 16-bit dynamic integer coordinates or signed 32-bit static coordinates; room or absent | Centers collision sampling independently of the source's lighting-coordinate center |
| Prepared missile light fields | Selected ready `Missiles` record's `Light`, `Red`, `Green`, `Blue` | Unsigned bytes | Light request and initial RGB for LMR-R17/R18; the selected record must be valid |
| Primary and temporary light associations | Supplied unit's light state | Each absent or a distinct live source; temporary-association storage present/absent | Primary lighting and temporary skill lighting are independently managed |
| Reached projectile light conditions | Supplied creation-stage state | Missile-light enable condition, suppression Boolean, force byte | Controls LMR-R18; a nonzero force byte selects the light operation without replacing the record's radius |
| Reached light operation | Supplied skill/creation/cleanup stage | Apply temporary light, apply primary projectile light, clear temporary light, or clean up both lights | Wider gameplay and settings eligibility are not inferred from the reached operation |
| Prepared monster lighting inputs | Selected MonStats/MonStats2 collections and counts, component-code collection, unique item-code lookup and ready item records | Signed class/link identities; unsigned light/RGB/count/variant bytes; four-byte codes | Supplies LMR-R20/R21/R22; valid selected variants fit their prepared component lists and any reached fixed override list |
| Prepared MonStats base and alignment | Selected ready MonStats record | Signed 16-bit `BaseId`; unsigned 8-bit `Align` | Base identity selects R20 override lists; alignment gates R21 small-light replacement |
| Monster visual-component state | Supplied monster state or absent | Sixteen unsigned selected-variant bytes | Selects the component code contributing an item light radius |
| Monster room and light conditions | Supplied current room or absent, signed level identity, client-context presence, level-8 small-light condition, small-light exclusion flag | Integer/Boolean fields | Selects component overrides or the reached initializer's special radius; no difficulty or condition-production default is inferred |
| Stored monster base radius | Existing unit light bookkeeping | Integer radius request | Written by LMR-R21; LMR-R22 recomputes its base rather than reading this stored value |
| Prepared overlay light fields | Selected ready `Overlay` record's `InitRadius`, `Radius`, `Red`, `Green`, `Blue` | Signed 32-bit initial/target requests; unsigned RGB bytes | Reached overlay light stage LMR-R23; the target request also enables the entire stage |
| New overlay light association | Supplied newly initialized overlay state | Initially absent source | Receives LMR-R23's creation result; primary and temporary unit associations are independent |

| Mapping or constant | Origin | Binding |
|---|---|---|
| Radius cap 18, stored scale 8, update step 8, stored-radius ceiling correction 248 | Engine code | Fixed by original code |
| Player base radius 13 | Engine code | Fixed by original code |
| Stat identities 89 `item_lightradius` and 90 `item_lightcolor` | Table data for names/ids; engine code for callback selection | [Named data records](../research/spec-evidence/draw-light-map-and-light-radius.md#table-data-records) verify names/ids; original dispatch fixes each operation |
| Distance weights 983 and 407, reciprocal scale 65536, map side 48 and channel ceiling 255 | Engine code | Fixed by original code |
| Shadow collision mask 34, missing-data sentinel 39, opaque value 16, interpolation scale 256, working-grid side 64 and clear count 1024 | Engine code | Fixed by original code |
| Prepared `Objects.Lit0` through `Lit7`, `Red`, `Green`, `Blue` | Table data interpreted by original loader descriptors and consumer code | Resolved byte fields of the selected ready record; supplied mode selects `Lit[mode]` |
| Prepared `Missiles.Light`, `Red`, `Green`, `Blue` | Table data interpreted by original loader descriptors and consumer code | Resolved byte fields of the selected ready record; LMR-R17/R18 consume these fields |
| State light producers and wider monster/skill/projectile eligibility | Unresolved | Remaining producer bindings and eligibility require LMR-Q02 |
| MonStats2 `light`, `light-r`, `light-g`, `light-b`; item `lightradius`; CompCode `code` | Table data interpreted by original schema and ready-field consumers | Prepared unsigned bytes/four-byte codes under LMR-R20/R21/R22; arbitrary compiled correspondence remains GDA-R02 |
| Component skip codes and level/base-identity override lists in LMR-R20 | Engine code and embedded constants | Fixed by original code, independent of difficulty |
| Prepared `Overlay.InitRadius`, `Radius`, `Red`, `Green`, `Blue` | Table data interpreted by original loader descriptors and consumer code | Resolved signed 32-bit radius requests and unsigned color bytes; arbitrary compiled correspondence remains GDA-R02 |

## 4. Rules (execution order)

**LMR-R01** Create a source [verified in 1.13c code (static)]. At the supplied-input creation boundary, a requested radius `r <= 0` returns no source and adds nothing to the source sequence. Otherwise use `q = min(r,18)`, create a source with current radius and target radius both `q*8`, the supplied kind/intensity/RGB, and the supplied owner identity. With a valid owner position, its initial center is `(logicalShiftRight(positionX,13)+4, logicalShiftRight(positionY,13)+4)` with 32-bit addition. Insert the new source at the front of the source sequence. A kind-2 source starts with an invalid shadow cache and reserves a square cache of `(2*q+1)^2` integer samples. A positive request with an absent owner reaches the coordinate getter's fatal assertion, rather than successfully creating an ownerless source. A player/monster/missile-like owner requiring a dynamic path also fatally asserts if that path is absent; object/item/tile-like owners require valid static position storage and have no successful missing-storage contract. No RNG draw occurs in the light-source computation. Generic allocator failure remains LMR-Q03 and is not a successful-creation case.

**LMR-R02** Change radius immediately [verified in 1.13c code (static)]. For an existing source and signed request `r <= 0`, leave both radii unchanged. For `r > 0`, assign `min(r,18)*8` to both current and target radius. For kind 2, invalidate and discard its cached shadow samples even if the numeric radius is unchanged. No RNG draw occurs. A positive request requires an existing source; absence takes the original fatal-error path, not a successful no-op.

**LMR-R03** Change only the target radius [verified in 1.13c code (static)]. For an existing source and signed request `r > 0`, set target radius to `min(r,18)*8` without changing current radius. A request `r <= 0` leaves the source unchanged. A positive request with no source takes the fatal-error path. No RNG draw occurs.

**LMR-R04** Apply an object-light request [verified in 1.13c code (static)]. Read `w = u16(request)`. If `w == 0`, remove an existing object source immediately and clear the object's source association; absent association is a no-op. Otherwise set `q = floor(w/2)`. If there is no associated source, attempt LMR-R01 with radius `q`, kind 2, intensity 255 and supplied RGB, and associate its result; a failed/nonpositive creation leaves no source. With an associated source, apply LMR-R03 with `q`. Existing-source RGB is not updated by this operation. In particular, request 1 does not remove an existing source, and cannot create an absent source. No RNG draw occurs.

At the prepared-object-record boundary, read the unsigned byte `Objects.Lit[mode]` for the supplied selected mode 0 through 7, and supply that as the request above. Read the record's unsigned `Red`, `Green`, `Blue` bytes as RGB for that request. The original text-loader descriptors bind these named fields through GDA-R03a/GDA-R03b/GDA-R03c; GDA-R02 still owns correspondence to an arbitrary compiled record. Full mode-selection and spawn eligibility remain LMR-Q02.

**LMR-R05** Refresh position and approach the target [verified in 1.13c code (static)]. During each source update, first refresh its center only if removal condition is false, its owner category is not the special ownerless category, and its registry lookup finds the owner. The refreshed center is `(arithmeticShiftRight(positionX,13)+4, arithmeticShiftRight(positionY,13)+4)`. A missing lookup leaves the center unchanged. Then, if current and target differ, invalidate/discard the kind-2 shadow cache and change current radius by exactly +8 when target is greater, or -8 otherwise. Do not replace this step with direct assignment or a generic overshoot clamp. After the step, a current radius at least 256 becomes 248. If removal condition is true and current radius is now nonpositive, remove the source and stop; otherwise dispatch rendering through LMR-R08. No RNG draw occurs.

**LMR-R06** Apply a packed color change [verified in 1.13c code (static)]. At this callback boundary, an absent unit, a unit other than player or monster, or a unit without a light source produces no change. With a source, packed value zero sets RGB to `(255,255,255)`. A nonzero value sets RGB to `(u8(value >> 16),u8(value >> 8),u8(value))`. Source intensity and both radii remain unchanged. No RNG draw occurs.

**LMR-R07** Apply a player radius-modifier change [verified in 1.13c code (static)]. When the registered client stat callback receives stat 89 `item_lightradius` for a player with an existing source, compute the signed 32-bit wrapped sum `13 + modifier`, replace a nonpositive sum with zero, then apply immediate-radius operation LMR-R02. Therefore a nonpositive sum leaves the player's existing source radius unchanged. An absent source is not created by this callback. Original equipment/stat aggregation that produces the supplied callback value remains LMR-Q02; LMR-R20/R22 specify prepared monster base-radius selection and the monster callback. The callback dispatch sends stat 90 `item_lightcolor` to LMR-R06. No RNG draw occurs in the player callback.

**LMR-R08** Dispatch light-map rendering [verified in 1.13c code (static)]. After LMR-R05's surviving-source update, kind 0 with lighting quality greater than 1 selects the viewport-shadow path in LMR-R16. Kind 2 with an invalid cache first reserves missing cache storage for `(2*(currentRadius >> 3)+1)^2` integer samples, then requests LMR-R15, and finally requests the cached path in LMR-R16. Kind 2 with a valid cache skips rebuilding and requests its cached path directly. All other cases request the uncached path. Radius changes invalidate kind-2 caches through LMR-R02/LMR-R05; a position refresh alone does not invalidate a valid cache. Successful full shadow/raster execution uses positive current radii divisible by eight and below 256, sufficient cache storage, and the supplied room/map/scratch inputs below. Allocation-failure behavior remains LMR-Q03. No RNG draw occurs.

**LMR-R09** Compute an uncached radial sample [verified in 1.13c code (static)]. At the supplied candidate-sample boundary, a source with current radius outside signed range 1 through 255 contributes nothing. Otherwise choose effective radius `r` initially equal to current radius. For lighting quality zero, a monster-owned source uses `r=8`, a missile-owned source contributes nothing, and a player-owned source whose identity differs from the local player uses `r=min(r,16)`; other cases retain `r`. These effective-radius changes do not modify source current or target radius.

For a selected sample `(x,y)` in lighting coordinates, compute signed wrapped differences `centerX-x` and `centerY-y`, and their 32-bit absolute values. Let `hi` be the greater absolute value under signed comparison and `lo` the other. Then:

```
distance = low32(hi*983 + lo*407) >>> 10
scale = truncateTowardZero((intensity*65536) / r)
contribution = low32((r-distance)*scale) >> 16
```

If signed `contribution <= 0`, emit no contribution. Otherwise request LMR-R10 for that sample with this intensity and the source RGB. No RNG draw occurs. The verified sample boundary supplies differences whose absolute values are representable as nonnegative signed 32-bit values; any positive contribution must be no greater than 255. Invalid-radius and nonpositive-contribution cases remain within this boundary and emit no request. LMR-R16 owns full enumeration, coarse viewport rejection and the shadowed paths.

**LMR-R10** Accumulate one light-map contribution [verified in 1.13c code (static)]. Given a positive contribution intensity `n` in 1 through 255 and source RGB bytes, compute `column=(x >> 3)-originX` and `row=(y >> 3)-originY` with 32-bit subtraction. If either signed result is outside 0 through 47, do not change the map. Otherwise select that map cell. Let its existing intensity be `p` and set `total=min(p+n,255)`. If the colored-light condition is false, write intensity `total` and RGB `(0,0,0)`. Otherwise, for each channel independently, use the existing byte `old` and supplied byte `new`:

```
reciprocal = floor(65536 / total)
channel = min(255, low32((old*p + new*n)*reciprocal) >>> 16)
```

Write intensity `total` and the three computed RGB bytes. Use the saturated intensity in the reciprocal; do not use the unsaturated sum or round to nearest. The other fields of the cell are unchanged. Source current/target radius and source RGB are not changed. No RNG draw occurs.

**LMR-R11** Schedule one reached light-map rebuild [verified in 1.13c code (static)]. Game-world initialization sets the current update counter to zero. Viewport creation sets the last-processed counter to zero. Each completed client-world tick increments the unsigned current counter by one modulo `2^32`. A reached eligible viewport-render stage compares the counters unsigned: if current is less than or equal to last-processed, skip rebuilding. Otherwise assign last-processed=current before requesting one light-map rebuild and one source-sequence traversal. Do not issue a catch-up rebuild for each skipped counter value. A wrapped current value remains skipped while it is less than or equal to last-processed; viewport creation can reset last-processed independently. This specifies reached client stages, not the wider game-loop eligibility or a server/display frame rate. No RNG draw occurs in this scheduling stage.

**LMR-R12** Query supplied room collision for lighting [runtime-confirmed (1.13c)]. For a subtile `(x,y)`, select the supplied current room if its signed bounds contain the point, using inclusive lower and exclusive upper bounds. Otherwise select the first nonnull room in its ordered adjacent-room sequence whose bounds contain the point. An absent current room, no containing room, absent selected collision storage, or absent collision cells returns 39. This missing-data result is not masked. With a selected grid, return the unsigned 16-bit cell at `(x-gridOriginX,y-gridOriginY)`, bitwise AND 34. A lighting blocker is any result whose low 16 bits are nonzero. This reads supplied collision storage; generating those room relationships and bits belongs outside this lighting-consumer boundary. No RNG draw occurs. Matching gameplay observation covers LMR-V21's current-room masking case only; adjacent-room and missing-data branches remain statically verified without matching observations.

**LMR-R13** Prepare shadow working state [verified in 1.13c code (static)]. Use logical working-grid offsets `(u,v)` from -32 through 31, with u increasing across columns and v across rows. `ClearShadow` sets P to zero only for rows v=-32 through -17, all 64 columns: exactly 1024 values. It leaves every other P value unchanged. Do not replace it with clearing the complete grid.

For a valid positive source radius divisible by eight, let `q=radius/8`, `cx=centerX >> 3`, `cy=centerY >> 3`.

- `FillViewportBlockers`: for v=-q through q, then u=-q through q, inspect light-map cell `(cx+u-originX,cy+v-originY)`. If either index is outside 0 through 47, use obstruction flag 1; otherwise use that cell's supplied obstruction byte. Store B[u,v]=flag*16. B outside this square remains unchanged.
- `FillOwnerBlockers`: with a resolved owner and nonnull current room, query LMR-R12 at `(ownerIntegerX+u,ownerIntegerY+v)` for every v=-32 through 31 and u=-32 through 31, in that row-major order. Store B[u,v]=16 for a nonzero low-16-bit result, otherwise 0. This fills all 4096 B values. Its sampling center is the owner's integer position, not a replacement derived from source center.

The caller chooses one fill operation and invokes `ClearShadow` before it. Neither operation changes map intensity/RGB or the source radius. No RNG draw occurs.

**LMR-R14** Propagate a shadow sample and sweep outward [verified in 1.13c code (static)]. For a working-grid target `(u,v)`, its direction sample in lighting coordinates is `(8*(centerX >> 3)+4+8*u, 8*(centerY >> 3)+4+8*v)`. Compute signed differences from this direction sample toward the actual source center. If both are zero, leave P[u,v] unchanged. Otherwise let sx and sy each be -1 for a negative difference and +1 otherwise; let dx and dy be the nonnegative signed magnitudes. All inputs in this successful stage keep these differences representable.

Define `Fetch(a,b)` as B[a,b] when B[a,b] is nonzero, otherwise P[a,b]. Thus an obstruction overrides an already propagated shadow at that neighboring sample.

```
if dx == 0: P[u,v] = Fetch(u,v+sy)
else if dy == 0: P[u,v] = Fetch(u+sx,v)
else if dx >= dy:
    w = truncateTowardZero((dy*256)/dx)
    a = Fetch(u+sx,v)
    b = Fetch(u+sx,v+sy)
    P[u,v] = low32((256-w)*a + w*b) >> 8
else:
    w = truncateTowardZero((dx*256)/dy)
    a = Fetch(u,v+sy)
    b = Fetch(u+sx,v+sy)
    P[u,v] = low32((256-w)*a + w*b) >> 8
```

For `SweepShadows(q)`, do nothing when q<2. Otherwise visit rings m=2 through q. For each ring, visit i=0 through m; at each i propagate targets in exactly this order:

```
(-i,-m), (i,-m), (-i,m), (i,m),
(m,-i), (m,i), (-m,-i), (-m,i)
```

Repeated targets at i=0 or i=m are processed repeatedly. The central three-by-three P square is never assigned by this sweep. B is unchanged. No RNG draw occurs.

**LMR-R15** Build a source-local shadow cache [verified in 1.13c code (static)]. For current radius outside signed range 1 through 255, return without changing scratch or cache. For a valid radius divisible by eight, invoke `ClearShadow` from LMR-R13, then resolve the source's owner from its stored identity. If the owner is absent, return with the partial scratch clear performed and cache validity unchanged. A resolved owner without a current room takes the fatal-error path. With a room, fill all owner-centered blockers through LMR-R13 and perform `SweepShadows(q)` through LMR-R14. This successful path requires cache validity false and cache storage present; either violated invariant takes the fatal-error path after propagation. Copy P[u,v] for v=-q through q and u=-q through q into the cache in row-major order, then set cache validity true. No RNG draw occurs. Collision-storage production remains outside the supplied-input boundary; an absent collision grid is nevertheless explicitly handled by LMR-R12's return 39.

**LMR-R16** Enumerate a complete source raster [verified in 1.13c code (static)]. Select the viewport-shadow, cached, or uncached path under LMR-R08. A current radius outside signed range 1 through 255 produces no contribution. Full successful shadowed rasters require the divisible-by-eight radius and scratch/cache constraints in LMR-R08; the uncached path also accepts the effective-radius selection in LMR-R09, including its quality-zero missile rejection. Shadowed paths use the current radius directly, without those low-quality owner reductions. Let the resulting effective radius be r and `N=(2*r) >> 3`. Set `x0=centerX-(centerX & 7)-r` and `y0=centerY-(centerY & 7)-r`.

Before cache validation or shadow preparation, reject the raster if any signed condition holds:

```
(x0 >> 3) > originX+48
(x0 >> 3)+N+1 < originX
(y0 >> 3) > originY+48
(y0 >> 3)+N+1 < originY
```

For the viewport-shadow path that survives rejection, invoke `ClearShadow`, `FillViewportBlockers`, then `SweepShadows(r/8)` from LMR-R13/LMR-R14. For the cached path that survives rejection, absent cache storage or a false cache-valid flag takes the fatal-error path; a valid cache is reused as supplied. The uncached path performs neither scratch preparation nor cache validation.

Visit j=0 through N, then i=0 through N, with sample `(x0+8*i,y0+8*j)`. Compute the unshadowed radial intensity using the distance, scale and contribution arithmetic of LMR-R09, with the already selected effective radius. For an uncached sample, emit a positive result through LMR-R10. For a shadowed sample, read `s=P[i-r/8,j-r/8]` on the viewport-shadow path or cache sample `(i,j)` on the cached path. Skip s>=16; otherwise replace radial intensity with `low32(radialIntensity*(8-(s >> 1))) >> 3` and emit only a signed positive result. At valid supplied shadow values 0 through 16, this preserves the per-contribution bound of 255. Map writes follow LMR-R10's clipping and channel arithmetic in traversal order. No RNG draw occurs. Preparing base environmental lighting and obstruction flags, and consuming the resulting map in a graphics driver, are outside this local raster boundary.

The following rules specify reached producer and cleanup stages feeding the source operations above; they do not reorder the source-update traversal.

**LMR-R17** Apply a temporary skill light [runtime-confirmed (1.13c)]. Supply a valid acting unit and position, a selected valid ready missile record, and successful allocation. If the record's unsigned `Light` byte is zero, leave both light associations and the source sequence unchanged. Otherwise create a kind-1 source through LMR-R01 with requested radius 1, intensity 255 and the record's unsigned `Red`, `Green`, `Blue`. Set its target radius through LMR-R03 using `Light`, leaving current radius 8. If temporary-association storage exists, replace that association with the new source and then remove its displaced source, if any. If that storage is absent, attach nothing and remove nothing: the created source remains in the global source sequence. The unit's primary light association is unchanged. No RNG draw occurs in this light stage. Matching observation covers only LMR-V22(c): positive Light, present temporary storage with no displaced source, new front insertion and attachment, final source state, and unchanged primary light. Zero Light, displaced-source removal and absent-storage branches remain statically verified without matching vector observations.

**LMR-R18** Apply a projectile's primary light [verified in 1.13c code (static)]. At the reached projectile-creation light stage, supply the projectile with valid position and a selected valid ready missile record. If the missile-light enable condition is false or suppression is true, remove its existing primary source if present and clear that association. Otherwise, if both the record's unsigned `Light` and the force byte are zero, leave its primary source unchanged. In every other case, first remove and clear its existing primary source, then request LMR-R01 with radius equal to `Light`, kind 1, intensity 255 and the record's unsigned RGB, and associate the result. Thus a nonzero force byte with `Light=0` removes an old primary light and creates none; it does not supply a fallback radius. A positive `Light` sets current and target equally through capped creation, without the radius-1 growth used by LMR-R17. Temporary-light state is unchanged. Successful allocation is supplied for a positive request; no RNG draw occurs in this light stage.

The original text-loader descriptors bind the prepared `Missiles.Light`, `Red`, `Green`, `Blue` fields through GDA-R03a/GDA-R03b/GDA-R03c. GDA-R02 owns correspondence to an arbitrary compiled record. These two rules require a valid selected record: they do not define missing-record recovery. Wider skill-action, projectile-spawn and configuration eligibility remain LMR-Q02.

**LMR-R19** Clear temporary lighting or clean up a unit's lights [verified in 1.13c code (static)]. `ClearTemporary` first replaces a present temporary association with absence, then removes its displaced source if present; it leaves the primary association unchanged. An absent unit or absent temporary-association storage makes `ClearTemporary` a no-op. `CleanupLights`, for a valid unit, first removes its primary source if present and clears that association, then performs `ClearTemporary`. Present source references must name distinct live entries in the source sequence. This specifies the reached light cleanup stage, not which gameplay transitions select it or the remaining unit teardown. No RNG draw occurs.

**LMR-R20** Select a monster's component light radius [verified in 1.13c code (static)]. An absent unit, a non-monster unit, or absent visual-component state returns radius zero. Otherwise inspect component indices 0 through 15 in order. Resolve the monster's class index into MonStats, then its signed `MonStatsEx` into MonStats2; an invalid class/link yields no component code. For each component, compare its selected unsigned variant against that component's unsigned variant count: a variant at or above the count yields code zero. Otherwise read that variant's prepared component-code index; an index at or above the CompCode count gives default code zero, otherwise use that CompCode entry's four-byte code.

After that valid-variant selection, a present room whose level identity is 40 through 74 inclusive enables the following overrides. Read the selected monster record's signed `BaseId` and validate it against the MonStats count. Base identity 0 or 170 selects its fixed lists below; another or invalid base identity retains the default code. A listed component replaces the default code with its selected-variant entry, even if that default was zero; an unlisted component retains its default. An absent room or any other level identity retains the default. This condition does not select by difficulty.

| Base identity | Component indices | Fixed code list in variant order |
|---|---|---|
| 0 | 0 | `lit `, `lit `, `des `, `des `, `hvy `, `hvy `, `hvy ` |
| 0 or 170 | 1 | `lit `, `med `, `hvy ` |
| 0 | 2, 3, 4 | `lit `, `des `, `hvy ` |
| 0 | 5 | `axe `, `axe `, `fla `, `fla `, `hax `, `hax `, `mac `, `mac `, `scm `, `scm ` |
| 0 | 7 | zero code, `buc `, `lrg `, `kit `, `sml ` |
| 0 | 8, 9 | twelve zero-code entries |
| 170 | 0, 2, 3, 4 | `lit `, `des `, `hvy ` |
| 170 | 6 | `sbw ` |
| 170 | 8, 9 | eleven zero-code entries |

The successful prepared-record boundary requires every reached override index to fit the corresponding list; out-of-list malformed combinations have no specified recovery result. Code zero and the exact four-byte codes `lit `, `med `, `hvy ` contribute zero without item lookup. For another code, find its unique ready item record by code; an absent item contributes zero, otherwise use its unsigned `lightradius` byte. Return the unsigned maximum over the sixteen contributions, initially zero. No RNG draw occurs. Component-list preparation and arbitrary compiled correspondence remain producer integration under LMR-Q02/GDA-R02.

**LMR-R21** Initialize a reached monster primary light [verified in 1.13c code (static)]. An absent/non-monster unit or an invalid/absent selected MonStats or linked MonStats2 record produces no light change. Otherwise let `r=max(LMR-R20 result, MonStats2.light)`, comparing unsigned bytes. Obtain the current room's level identity; absence gives zero. At level identity 8, an absent client context takes the fatal-assertion path. With that context present, if the supplied level-8 small-light condition is true, the unit's small-light exclusion flag is false, and `MonStats.Align` is neither 1 nor 2, replace r with exactly 3, including when the computed maximum was zero or greater than 3. In all other cases keep r, and if it is zero leave existing lighting unchanged.

For the resulting positive r, create a kind-0 source through LMR-R01 with radius r, intensity 255 and the linked MonStats2 unsigned `light-r`, `light-g`, `light-b`. Then remove and clear an existing primary source, associate the new source, and store the uncapped r as the unit's stored base radius. The source itself still uses LMR-R01's radius cap. Temporary-light state is unchanged. Successful allocation and valid owner position are supplied; no RNG draw occurs in this lighting stage. Wider initializer eligibility and production of the small-light condition remain LMR-Q02.

**LMR-R22** Apply a monster radius-modifier callback [verified in 1.13c code (static)]. The registered stat-89 callback for a monster with an existing primary source resolves its selected linked MonStats2 record. Missing/invalid selection produces no change. Otherwise recompute `base=max(LMR-R20 result, MonStats2.light)` from unsigned bytes, compute the signed wrapped sum `base+modifier`, replace a nonpositive sum with zero, and apply LMR-R02. Thus zero/nonpositive results preserve existing radii and cache, while positive results assign capped current/target and invalidate a kind-2 cache. An absent source is not created. The callback does not apply LMR-R21's level-8 radius-3 replacement and does not read the stored base radius. RGB and intensity are unchanged; no RNG draw occurs. Upstream stat aggregation remains LMR-Q02.

**LMR-R23** Apply a reached overlay-light stage [verified in 1.13c code (static)]. Given a newly initialized overlay with no light association and valid prepared fields, if `Overlay.Radius` is zero, perform no light operation, including when `InitRadius` is positive. Otherwise create a kind-1 source through LMR-R01 with requested radius `InitRadius`, intensity 255 and the prepared RGB, using the supplied valid owner position and successful allocation when creation requires it. Store the creation result as the overlay's light association. If `Radius` differs from `InitRadius`, apply LMR-R03 to that result with request `Radius`; if equal, do not call the setter. Thus positive initial and target requests can produce unequal current and target radii, a negative target leaves a successfully created source unchanged, and a nonpositive initial request with a positive target reaches the target setter's fatal absent-source assertion after storing the absent result. This stage does not change the unit's primary or temporary light association and consumes no RNG. Complete overlay eligibility, animation and lifetime remain LMR-Q02.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| LMR-R01 | Create capped source | verified in 1.13c code (static) | none for supplied valid owner and successful allocation; LMR-Q03 names generic allocator-failure integration | LMR-V01, LMR-V02 |
| LMR-R02 | Immediate radius | verified in 1.13c code (static) | none | LMR-V03 |
| LMR-R03 | Target radius | verified in 1.13c code (static) | none | LMR-V04 |
| LMR-R04 | Object request conversion/removal | verified in 1.13c code (static) | LMR-Q02 blocks upstream mode/spawn eligibility and arbitrary compiled-record correspondence, not prepared field selection or supplied requests | LMR-V05, LMR-V06 |
| LMR-R05 | Position refresh and radius step | verified in 1.13c code (static) | LMR-Q01 blocks assigning server-frame cadence | LMR-V07, LMR-V08 |
| LMR-R06 | Packed RGB | verified in 1.13c code (static) | LMR-Q02 blocks original stat producer, not supplied callback operation | LMR-V09 |
| LMR-R07 | Player modifier callback | verified in 1.13c code (static) | LMR-Q02 blocks equipment/stat aggregation, not supplied player/monster callbacks | LMR-V09, LMR-V10 |
| LMR-R08 | Rendering dispatch | verified in 1.13c code (static) | LMR-Q03 blocks allocation-failure recovery; prepared successful execution available | LMR-V19, LMR-V20 |
| LMR-R09 | Supplied uncached radial sample | verified in 1.13c code (static) | none | LMR-V11 |
| LMR-R10 | One map-cell contribution | verified in 1.13c code (static) | none for supplied origin/cell/contribution | LMR-V12, LMR-V13 |
| LMR-R11 | Reached tick and viewport scheduling | verified in 1.13c code (static) | LMR-Q01 blocks wider scheduler eligibility, not supplied completed/reached stages | LMR-V14 |
| LMR-R12 | Lighting collision query | runtime-confirmed (1.13c) | none locally; LMR-Q04 names outside-boundary collision producers | LMR-V15, LMR-V21; only V21 observed |
| LMR-R13 | Partial scratch clear and blocker sampling | verified in 1.13c code (static) | none for supplied initialized map/room grids and scratch state | LMR-V16, LMR-V18 |
| LMR-R14 | Directional interpolation and ring sweep | verified in 1.13c code (static) | none | LMR-V17, LMR-V18 |
| LMR-R15 | Source-local cache build | verified in 1.13c code (static) | LMR-Q03 blocks failed storage reservation; supplied valid storage available | LMR-V18, LMR-V20 |
| LMR-R16 | Complete prepared source raster | verified in 1.13c code (static) | none locally; LMR-Q04 names outside-boundary map preparation and graphics consumption | LMR-V19, LMR-V20 |
| LMR-R17 | Temporary skill-light replacement | runtime-confirmed (1.13c) | LMR-Q02 limits wider skill/record selection, not valid prepared-record light application | LMR-V22; only case (c) observed |
| LMR-R18 | Reached projectile primary-light operation | verified in 1.13c code (static) | LMR-Q02 limits spawn/configuration eligibility and arbitrary record correspondence | LMR-V23 |
| LMR-R19 | Reached temporary and dual light cleanup | verified in 1.13c code (static) | none locally for distinct live sources; wider transition eligibility is LMR-Q02 | LMR-V24 |
| LMR-R20 | Prepared monster component light selection | verified in 1.13c code (static) | LMR-Q02/GDA-R02 limit component/data production, not valid prepared-field selection | LMR-V25 |
| LMR-R21 | Reached monster primary-light initialization | verified in 1.13c code (static) | LMR-Q02 limits upstream eligibility and condition production | LMR-V26 |
| LMR-R22 | Monster modifier callback | verified in 1.13c code (static) | LMR-Q02 limits stat aggregation | LMR-V27 |
| LMR-R23 | Prepared overlay initial/target light requests | verified in 1.13c code (static) | LMR-Q02 limits wider overlay eligibility, lifetime and arbitrary compiled correspondence | LMR-V28 |

## 5. Outputs and side effects

| Output | Rule |
|---|---|
| New source inserted at the front, initialized radii/center/intensity/RGB and kind-2 cache allocation | LMR-R01 |
| Immediate current/target radius replacement and kind-2 cache invalidation | LMR-R02 |
| Target-radius replacement | LMR-R03 |
| Object/source association or immediate removal/cleared association | LMR-R04 |
| Refreshed center, stepped current radius, invalidated cache, expiry removal | LMR-R05 |
| Source RGB replacement | LMR-R06 |
| Radius-change request from player modifier | LMR-R07 |
| Rendering-path selection, cache-build request and source-raster request | LMR-R08 |
| Positive uncached sample contribution request | LMR-R09 |
| Selected map-cell intensity and RGB, or clipped no-op | LMR-R10 |
| Current/last-processed counters and reached rebuild request | LMR-R11 |
| Lighting collision result; no collision-storage write | LMR-R12 |
| Partially cleared P and sampled B working values | LMR-R13 |
| Ordered target evaluations and interpolated P values; evaluation traces need not be persistent game state | LMR-R14 |
| Source-local cache samples and valid flag | LMR-R15 |
| Ordered map-cell contribution requests and resulting map writes | LMR-R16 |
| New temporary light, target growth request, temporary replacement and old-source removal | LMR-R17 |
| Primary projectile light retention, removal or replacement | LMR-R18 |
| Temporary association clearing or ordered primary-then-temporary light cleanup | LMR-R19 |
| Maximum component radius, monster primary source and stored uncapped base radius | LMR-R20, LMR-R21 |
| Recomputed monster modifier radius and immediate setter effects | LMR-R22 |
| Overlay light association, new kind-1 source and optional target request | LMR-R23 |

These operations do not establish a server stat write, network message, or durable character-save effect.

## 6. Branches and edge cases

Nonpositive radius requests in LMR-R02/R03 leave an existing source intact; object request zero instead removes it under LMR-R04. Object request 1 is a distinct no-create/no-remove case. Allocation centers use a logical shift, whereas subsequent position refresh uses an arithmetic shift; synthetic negative coordinates must not silently equate those expressions. Radius stepping has no generic overshoot correction. RGB replacement does not change intensity. The specified local light computations consume no RNG. Component overrides select by level identity rather than difficulty; source kind and owner category still select the explicit branches above. Wider difficulty, attack and PvP source-creation eligibility is not established. Classic mode is outside the evidence profile. Specified invalid-record guard exits remain part of their rules; wider missing-data recovery and source eligibility remain LMR-Q02.

## 7. Open questions

| Question | Status | Evidence that would settle it |
|---|---|---|
| **LMR-Q01** Wider client scheduler eligibility | unresolved | Establish which game-loop modes and timing events reach completed client-world ticks and eligible viewport stages; LMR-R11 already specifies the supplied-stage counters, resets and wrap behavior |
| **LMR-Q02** Wider authored producer and callback integration | unresolved | Verify equipment/stat aggregation, state producer dependencies, component/data preparation, complete monster/object-mode/skill-action/projectile-spawn/configuration eligibility and arbitrary compiled-record correspondence. Prepared object selection and the reached missile-record light application/replacement/cleanup stages are specified by LMR-R04/R17/R18/R19; prepared monster selection, initialization and callback are specified by LMR-R20/R21/R22; reached prepared overlay lighting is specified by LMR-R23, while full overlay eligibility and lifetime remain unestablished; specified record guards remain usable, while wider missing-data recovery is unestablished |
| **LMR-Q03** Generic allocation-failure integration | unresolved | Establish allocator/diagnostic failure semantics in the infrastructure owner; no local successful-allocation assertion is blocked. Positive absent-owner and required dynamic-path assertions are specified in LMR-R01 |
| **LMR-Q04** World/base-map producers and final graphics consumption | unresolved | Verify collision-grid materialization, environmental/base obstruction initialization and graphics-driver map consumption in their owning contracts; no local supplied-input shadow/raster assertion is blocked |
| **LMR-Q05** Gameplay confirmation | unresolved | Reproduce the executable cases on an identified owned 1.13c client, distinguishing source-state changes from the rendered image |

Attempted checks and remaining execution boundaries are recorded in the linked [gap file](../plans/gaps/draw-light-map-and-light-radius.md).

## 8. Test vectors

### 8.1 Executable vectors

All vectors below use the supplied-input stage boundary, Expansion mode, no RNG seed, and no RNG draws. `Remove` removes the named source from the source sequence; `Invalidate` discards that source's kind-2 cache. Successful creation supplies valid owner identity and position and successful allocation. Cases are independent unless their own Inputs explicitly specify prior state. Source-state outputs do not assert a final rendered pixel. Shadow vectors give all initial working-grid/cache/map values they consume; those stores persist across the operations within that vector. Collision queries, cache-build requests, rebuild requests and emitted map contributions are logged in order when a vector asserts them. Map intensity/RGB writes persist within a vector under LMR-R10. A room's grid cells below are already materialized; these vectors do not test a producer of world geometry.

- **LMR-V01** Creation cap and center (LMR-R01; boundary: supplied-input stage)
  - Inputs: radius 60; kind 2; intensity 255; RGB `(10,20,30)`; owner position `(655360,1310720)`; source sequence `[A]`.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: `q=min(60,18)=18`; radii `18*8=144`; centers `(655360 >>> 13)+4=84` and `(1310720 >>> 13)+4=164`; cache side `2*18+1=37`.
  - Outputs: new source B; sequence `[B,A]`; B current/target `144/144`, center `(84,164)`, intensity 255, RGB `(10,20,30)`, kind 2, invalid cache with `37*37=1369` samples reserved.
  - Discriminates: uncapped allocation produces radius 480 and 14641 samples, rejected by 144 and 1369.
  - runtime_observed: no
  - (a) Inputs: radius 2; kind 1; intensity 255; RGB `(10,20,30)`; valid owner fixed-point position `(-65536,-131072)`; source sequence `[A]`. Trace: radius `2*8=16`; 32-bit logical shifts yield `524280+4=524284`, `524272+4=524276`. Outputs: new B at front `[B,A]`, current/target `16/16`, center `(524284,524276)`, kind 1, intensity 255, RGB `(10,20,30)`; no kind-2 cache reservation. Draws: none. Discriminates: arithmetic shifts give `(-4,-12)`, rejected by the positive center values.

- **LMR-V02** Nonpositive creation (LMR-R01; boundary: supplied-input stage)
  - Inputs: radius 0; kind 1; intensity 255; RGB `(255,255,255)`; source sequence `[A]`; owner lookup is not needed.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: `0<=0`, so return before source creation.
  - Outputs: no source returned; sequence `[A]`; no owner-position read and no source allocation.
  - Discriminates: forcing a minimum radius of 1 would create a source, rejected by the absent result.
  - runtime_observed: no
  - (a) Inputs: radius 1, kind 1, intensity 255, RGB `(255,255,255)`; owner absent; successful source-record allocation available. Trace: positive radius reaches coordinate read with an absent owner and fatally asserts. Outputs: fatal assertion, no successful source-creation result. Draws: none.

- **LMR-V03** Immediate no-op and equal-radius invalidation (LMR-R02; boundary: supplied-input stage)
  - Inputs: kind-2 source current/target `104/120`, valid cache; request 0.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: nonpositive request skips assignment and cache invalidation.
  - Outputs: current/target `104/120`, cache still valid.
  - Discriminates: extinguishing on zero would change current radius, rejected by 104.
  - runtime_observed: no
  - (a) Inputs: kind-2 source current/target `104/104`, valid cache; request 13. Trace: `13*8=104`; assign both and invalidate despite equal existing values. Outputs: current/target `104/104`, cache invalid and discarded. Draws: none.
  - (b) Inputs: kind-1 source current/target `8/16`; request 19. Trace: `min(19,18)*8=144`. Outputs: current/target `144/144`. Draws: none.

- **LMR-V04** Target-only update (LMR-R03; boundary: supplied-input stage)
  - Inputs: kind-2 source current/target `8/8`, valid cache; request 99.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: `min(99,18)*8=144`; assign target only.
  - Outputs: current/target `8/144`, valid cache unchanged by this request.
  - Discriminates: immediate resizing changes current to 144; eager invalidation changes cache validity, both rejected.
  - runtime_observed: no
  - (a) Inputs: kind-2 source current/target `8/144`, valid cache; request 0. Trace: nonpositive request skips writes. Outputs: current/target `8/144`, cache valid unchanged. Draws: none.

- **LMR-V05** Object zero removes, one does not (LMR-R04; boundary: supplied-input stage)
  - Inputs: object associated with A; A current/target `40/40`, RGB `(1,2,3)`; request 0; supplied RGB `(9,8,7)`.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: `u16(0)=0`, remove A and clear association.
  - Outputs: A removed; object has no source.
  - Discriminates: treating zero like the target setter leaves A present, rejected by removal.
  - runtime_observed: no
  - (a) Inputs: object associated with A; A current/target `40/40`, RGB `(1,2,3)`; request 1; supplied RGB `(9,8,7)`. Trace: `floor(1/2)=0`; target setter is a no-op. Outputs: A retained at `40/40`, RGB `(1,2,3)`. Draws: none.
  - (b) Inputs: object has no source; request 1; supplied RGB `(9,8,7)`. Trace: half-request zero makes creation return absent. Outputs: no source and no association. Draws: none.

- **LMR-V06** Object half-radius and unchanged existing color (LMR-R04; boundary: supplied-input stage)
  - Inputs: object associated with A; A kind 2, current/target `40/24`, intensity 255, RGB `(1,2,3)`; request 11; supplied RGB `(9,8,7)`.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: `floor(11/2)=5`; target `5*8=40`; existing source is reused.
  - Outputs: A current/target `40/40`, intensity 255, RGB `(1,2,3)`; no new source.
  - Discriminates: ignoring a positive request leaves target 24; rounding half upward produces target 48; refreshing RGB produces `(9,8,7)`; all are rejected.
  - runtime_observed: no
  - (a) Inputs: object has no associated source; request 11; supplied RGB `(9,8,7)`; valid owner position `(655360,1310720)`; empty source sequence; successful allocation. Trace: `floor(11/2)=5`; create kind 2 at radius 40 and center `(84,164)`; cache side 11. Outputs: new A associated with object and source sequence `[A]`, current/target `40/40`, center `(84,164)`, kind 2, intensity 255, RGB `(9,8,7)`, invalid cache with 121 integer samples reserved. Draws: none.
  - (b) Inputs: selected ready object record has `Lit0..Lit7=[0,11,0,0,0,0,0,0]`, `Red=9`, `Green=8`, `Blue=7`; supplied selected mode 1; object associated with kind-2 A current/target `40/24`, intensity 255, RGB `(1,2,3)`. Trace: mode 1 selects request 11, half-request 5 updates target 40. Outputs: A current/target `40/40`, intensity 255, RGB `(1,2,3)`; no new source. Draws: none. Discriminates: using `Lit0` would remove A; using an adjacent `Lit2` would also remove A.
  - (c) Inputs: selected ready object record has `Lit0..Lit7=[0,11,0,0,0,0,0,0]`, `Red=9`, `Green=8`, `Blue=7`; supplied selected mode 2; object associated with kind-2 A current/target `40/24`, intensity 255, RGB `(1,2,3)`. Trace: mode 2 selects request 0. Outputs: A removed and association cleared. Draws: none.

- **LMR-V07** One step, position refresh, and cache invalidation (LMR-R05; boundary: supplied-input stage)
  - Inputs: source A kind 2, removal false, ordinary owner found, current/target `8/32`, center `(0,0)`, valid cache; owner fixed-point position `(655360,1310720)`.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: arithmetic shifts produce center `(84,164)`; unequal radii invalidate cache; target greater gives `8+8=16`.
  - Outputs: A retained, center `(84,164)`, current/target `16/32`, cache invalid and discarded before renderer dispatch. No rendered sample asserted.
  - Discriminates: direct target assignment gives 32; retaining cache leaves it valid, both rejected.
  - runtime_observed: no
  - (a) Inputs: kind-1 source current/target `16/0`, removal true, center `(7,9)`; owner found at fixed-point `(655360,1310720)`. Trace: removal suppresses center refresh; step gives 8. Outputs: retained at current/target `8/0`, center `(7,9)`. Draws: none.
  - (b) Inputs: kind-1 source current/target `8/0`, removal true, center `(7,9)`. Trace: step gives 0; removal condition holds. Outputs: source removed; no renderer dispatch. Draws: none.
  - (c) Inputs: kind-1 source current/target `16/16`, removal false, ordinary owner found, center `(0,0)`; owner fixed-point position `(-65536,-131072)`. Trace: arithmetic shifts yield `-8+4=-4`, `-16+4=-12`; equal radii skip stepping. Outputs: retained at current/target `16/16`, center `(-4,-12)` before renderer dispatch. Draws: none. Discriminates: logical shifts give `(524284,524276)`, rejected by the negative center values.

- **LMR-V08** Stored-radius ceiling correction (LMR-R05; boundary: supplied-input stage)
  - Inputs: source current/target `256/256`, kind 1, removal false; owner absent; center `(7,9)`.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: equal radii skip step; `256>=256` sets current to 248.
  - Outputs: current/target `248/256`, center `(7,9)`; source retained. No rendered sample asserted.
  - Discriminates: placing the ceiling correction only inside the unequal-radius branch leaves 256, rejected by 248.
  - runtime_observed: no

- **LMR-V09** Packed color and zero reset (LMR-R06, LMR-R07; boundary: supplied-input stage)
  - Inputs: player with source current/target `104/104`, intensity 200, RGB `(9,8,7)`; packed value 1122867 (`0x112233`).
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: high/middle/low selected bytes are 17,34,51.
  - Outputs: RGB `(17,34,51)`, intensity 200, current/target `104/104`.
  - Discriminates: reversing channel extraction produces `(51,34,17)`, rejected by `(17,34,51)`.
  - runtime_observed: no
  - (a) Inputs: monster with source current/target `40/80`, intensity 200, RGB `(9,8,7)`; packed value 0. Trace: zero selects white. Outputs: RGB `(255,255,255)`, intensity 200, current/target `40/80`. Draws: none.
  - (b) Inputs: registered player callback stat 90 `item_lightcolor`, new value 1122867; player source current/target `104/104`, intensity 200, RGB `(9,8,7)`. Trace: stat 90 dispatches packed color operation; selected bytes 17,34,51. Outputs: RGB `(17,34,51)`, current/target `104/104`, intensity 200. Draws: none.
  - (c) Inputs: object with source current/target `40/80`, intensity 200, RGB `(9,8,7)`; supplied packed-color operation value 1122867. Trace: unsupported category returns before source writes. Outputs: current/target `40/80`, intensity 200, RGB `(9,8,7)` unchanged. Draws: none.

- **LMR-V10** Player modifier is absolute relative to base (LMR-R07; boundary: supplied-input stage)
  - Inputs: player with kind-0 source current/target `104/104`, RGB `(255,255,255)`, intensity 255; callback stat 89 `item_lightradius`, new modifier 2.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: `13+2=15`; immediate radius `15*8=120`.
  - Outputs: current/target `120/120`; source RGB/intensity unchanged.
  - Discriminates: treating the supplied modifier as the complete radius gives 16, rejected by 120.
  - runtime_observed: no
  - (a) Inputs: player with source current/target `120/120`; callback stat 89 `item_lightradius`, modifier -13. Trace: `13-13=0`; immediate setter ignores zero. Outputs: current/target `120/120`, source retained. Draws: none.
  - (b) Inputs: player with source current/target `120/120`; callback stat 89 `item_lightradius`, modifier 0. Trace: `13+0=13`, immediate radius 104. Outputs: current/target `104/104`. Draws: none.
  - (c) Inputs: player has no source; callback stat 89 `item_lightradius`, modifier 2. Trace: absent source skips radius setter. Outputs: player still has no source, no source created. Draws: none.

- **LMR-V11** Integer radial falloff and low-quality eligibility (LMR-R09; boundary: supplied-input stage)
  - Inputs: kind-1 source current radius 16, center `(0,0)`, intensity 255, RGB `(10,20,30)`; lighting quality 1; owner category missile; selected sample `(8,0)`.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: effective radius 16; distance `(8*983+0*407) >>> 10=7`; scale `16711680/16=1044480`; contribution `(9*1044480)>>16=143`.
  - Outputs: one request at `(8,0)`, intensity 143, RGB `(10,20,30)`; source radius remains 16.
  - Discriminates: exact Euclidean distance gives 8 and intensity 127; rounding final contribution upward gives 144; both rejected by 143.
  - runtime_observed: no
  - (a) Inputs: same source values explicitly current 16, center `(0,0)`, intensity 255, RGB `(10,20,30)`; quality 1; missile owner; sample `(8,8)`. Trace: distance `(8*983+8*407)>>>10=10`; contribution `(6*1044480)>>16=95`. Outputs: request at `(8,8)` with intensity 95 and RGB `(10,20,30)`, source radius 16. Draws: none.
  - (b) Inputs: current 16, center `(0,0)`, intensity 255, RGB `(10,20,30)`; quality 0; missile owner; sample `(0,0)`. Trace: low quality excludes missile-owned source before falloff. Outputs: no contribution request; source radius 16. Draws: none.
  - (c) Inputs: current 104, center `(0,0)`, intensity 255, RGB `(10,20,30)`; quality 0; monster owner; sample `(8,0)`. Trace: effective radius 8; distance 7; scale 2088960; contribution 31. Outputs: request intensity 31 at `(8,0)` with RGB `(10,20,30)`; source radius 104 unchanged. Draws: none.
  - (d) Inputs: current 104, center `(0,0)`, intensity 255, RGB `(10,20,30)`; quality 0; player owner identity 2, local player identity 1; sample `(8,0)`. Trace: effective radius 16, distance 7, scale 1044480, contribution 143. Outputs: request intensity 143 at `(8,0)`, RGB `(10,20,30)`; source radius 104 unchanged. Draws: none.
  - (e) Inputs: current 256, center `(0,0)`, intensity 255, RGB `(10,20,30)`; quality 1; missile owner; sample `(0,0)`. Trace: invalid current radius exits before falloff. Outputs: no contribution; source radius 256 unchanged. Draws: none.
  - (f) Inputs: current 16, center `(0,0)`, intensity 255, RGB `(10,20,30)`; quality 1; missile owner; sample `(24,0)`. Trace: distance `(24*983)>>>10=23`; contribution `(-7*1044480)>>16=-112`, not positive. Outputs: no contribution; source radius 16 unchanged. Draws: none.

- **LMR-V12** Saturated intensity is the color denominator (LMR-R10; boundary: supplied-input stage)
  - Inputs: sample `(0,0)`, map origin `(0,0)`, colored-light true; cell intensity 200 and RGB `(10,0,0)`; new intensity 100 and RGB `(200,0,255)`.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: `total=min(300,255)=255`, reciprocal 257; red `(22000*257)>>>16=86`, green 0, blue `(25500*257)>>>16=99`.
  - Outputs: cell intensity 255 and RGB `(86,0,99)`; other cells and source unchanged.
  - Discriminates: dividing by unsaturated 300 gives red 73 and blue 85; both rejected.
  - runtime_observed: no
  - (a) Inputs: sample `(0,0)`, origin `(0,0)`, colored-light false; cell intensity 200, RGB `(10,20,30)`; new intensity 100, RGB `(200,0,255)`. Trace: total 255; disabled color forces zeros. Outputs: cell intensity 255, RGB `(0,0,0)`. Draws: none.
  - (b) Inputs: sample `(0,0)`, origin `(0,0)`, colored-light true; cell intensity 255, RGB `(255,255,255)`; new intensity 255, RGB `(255,255,255)`. Trace: total 255, reciprocal 257; each raw weighted channel `(130050*257)>>>16=509`, clamped 255. Outputs: cell intensity 255, RGB `(255,255,255)`. Draws: none.

- **LMR-V13** Map clipping and reciprocal truncation (LMR-R10; boundary: supplied-input stage)
  - Inputs: sample `(8,8)`, origin `(0,0)`, colored-light true; selected cell intensity 2 and RGB `(10,10,10)`; new intensity 3 and RGB `(30,30,30)`.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: column 1,row 1; total 5; reciprocal `floor(65536/5)=13107`; each channel `(110*13107)>>>16=21`.
  - Outputs: cell `(column 1,row 1)` intensity 5 and RGB `(21,21,21)`; other cells unchanged.
  - Discriminates: direct exact division gives 22, rejected by 21.
  - runtime_observed: no
  - (a) Inputs: sample `(-1,8)`, origin `(0,0)`, colored-light true; map cells each intensity 2, RGB `(10,10,10)`; new intensity 3, RGB `(30,30,30)`. Trace: `-1>>3=-1`, negative column clips. Outputs: map unchanged. Draws: none.
  - (b) Inputs: sample `(384,8)`, origin `(0,0)`, colored-light true; map cells each intensity 2, RGB `(10,10,10)`; new intensity 3, RGB `(30,30,30)`. Trace: `384>>3=48`, column 48 clips. Outputs: map unchanged. Draws: none.
  - (c) Inputs: sample `(8,384)`, origin `(0,0)`, colored-light true; map cells each intensity 2, RGB `(10,10,10)`; new intensity 3, RGB `(30,30,30)`. Trace: row 48 clips. Outputs: map unchanged. Draws: none.

- **LMR-V14** Reached tick/render scheduling (LMR-R11; boundary: supplied-input stage)
  - Inputs: current counter 77 and last-processed 66; ordered events game-world initialization, viewport creation, eligible viewport render, completed client tick, eligible viewport render, eligible viewport render; empty rebuild log.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: resets produce 0/0; first render skips; tick produces 1/0; next render sets last 1 then requests one rebuild; repeated render skips.
  - Outputs: counters 1/1; rebuild log has exactly one request with counters already 1/1 at request time.
  - Discriminates: rebuilding on equality adds two requests; assigning last after the rebuild request changes logged counter state.
  - runtime_observed: no
  - (a) Inputs: counters 10/10; events three completed ticks then one eligible render; empty log. Trace: current 11,12,13; render sets last 13 once. Outputs: counters 13/13, exactly one rebuild request, not three. Draws: none.
  - (b) Inputs: counters 4294967295/4294967295; events completed tick, eligible render, completed tick, eligible render, viewport creation, eligible render; empty log. Trace: wrap 0/4294967295 skips, then 1/4294967295 skips; viewport resets last 0, render accepts 1>0. Outputs: counters 1/1 and one rebuild request. Draws: none.

- **LMR-V15** Room collision selection and missing-data sentinel (LMR-R12; boundary: supplied-input stage)
  - Inputs: current room A bounds `[0,2)` by `[0,2)` with grid origin `(0,0)`, width 2, all unsigned 16 cells 65535; adjacent list `[absent,B]`, B bounds `[2,4)` by `[0,2)` with grid origin `(2,0)`, width 2, all cells 8; queried subtile `(2,1)`.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: x2 is outside A's exclusive upper bound; skip absent adjacent, select B; cell 8 AND34=0.
  - Outputs: result 0, not a blocker; no storage writes.
  - Discriminates: inclusive upper bounds select A and return 34, rejected by 0.
  - runtime_observed: no
  - (a) Inputs: current A bounds `[0,2)` by `[0,2)` and absent collision storage, no adjacent rooms; query `(1,1)`. Trace: select A; missing storage returns 39 without masking. Outputs: result 39 and blocker true. Draws: none.
  - (b) Inputs: current A bounds `[0,2)` by `[0,2)`, grid origin `(0,0)`, width 2, all cells 2; no adjacent rooms; query `(1,1)`. Trace: cell 2 AND34=2. Outputs: result 2 and blocker true. Draws: none.

- **LMR-V16** Partial scratch clear and viewport sampling (LMR-R13; boundary: supplied-input stage)
  - Inputs: radius 8, center `(4,4)`, origin `(0,0)`; viewport obstruction flags all 0 except cell `(1,0)` is 1; P all 7; B all 9; operations ClearShadow, FillViewportBlockers.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: clear 16*64=1024 P values at rows-32..-17; q=1 fills B at offsets-1..1, using outside-window flag 1.
  - Outputs: P rows-32..-17 are 0; all other P remain 7. B square, rows v=-1,0,1 and columns u=-1,0,1, is `[[16,16,16],[16,0,16],[16,0,0]]`; B outside square remains 9; map/source unchanged.
  - Discriminates: clearing all P changes its central 7; treating outside viewport as clear changes the negative-edge 16 values.
  - runtime_observed: no

- **LMR-V17** Neighbor choice, obstruction precedence and interpolation (LMR-R14; boundary: isolated rule)
  - Inputs: source center `(4,4)`; target offset `(2,1)`; B all 0 except B[1,1]=16; P all 0 except P[1,1]=3 and P[1,0]=4.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: direction sample `(20,12)` gives differences(-16,-8), sx=sy=-1, x-major weight 128; a=16 overrides propagated 3; b=4; `(128*16+128*4)>>8=10`.
  - Outputs: P[2,1]=10; every other P/B value unchanged.
  - Discriminates: preferring propagated 3 gives 3, rejected by 10; omitting the diagonal gives 16, rejected by 10.
  - runtime_observed: no
  - (a) Inputs: center `(4,4)`, target `(1,2)`; B all 0 except B[1,1]=16; P all 0 except P[1,1]=3, P[0,1]=4. Trace: direction(-8,-16), y-major weight 128; output `(128*16+128*4)>>8=10`. Outputs: only P[1,2] becomes 10. Draws: none.
  - (b) Inputs: center `(4,4)`, target `(2,0)`; B all 0; P all 0 except P[1,0]=5. Trace: horizontal direction selects Fetch(1,0). Outputs: only P[2,0] becomes 5. Draws: none.
  - (c) Inputs: center `(4,4)`, target `(0,-2)`; B all 0; P all 0 except P[0,-1]=6. Trace: vertical direction selects Fetch(0,-1). Outputs: only P[0,-2] becomes 6. Draws: none.
  - (d) Inputs: center `(4,4)`, target `(0,0)`; B all 0; P all 0 except P[0,0]=7. Trace: coincident direction returns without assignment. Outputs: P unchanged, including center 7. Draws: none.

- **LMR-V18** Ring sweep and owner-centered cache (LMR-R12, LMR-R13, LMR-R14, LMR-R15; boundary: supplied-input stage)
  - Inputs: source center (4,4), q=2; B all 0 except B[-1,0]=16; P all 0; empty logical target-evaluation log; invoke SweepShadows(2) only.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: ring 2 visits 3*8=24 targets including repeats. The left-axis outer cell selects blocker 16; the two adjacent left-edge cells interpolate half 16 and half 0.
  - Outputs: P central five-by-five square, rows v=-2..2 and columns u=-2..2, is `[[0,0,0,0,0],[8,0,0,0,0],[16,0,0,0,0],[8,0,0,0,0],[0,0,0,0,0]]`; all other P stay 0 and B unchanged. Target-evaluation log is exactly `[(0,-2),(0,-2),(0,2),(0,2),(2,0),(2,0),(-2,0),(-2,0),(-1,-2),(1,-2),(-1,2),(1,2),(2,-1),(2,1),(-2,-1),(-2,1),(-2,-2),(2,-2),(-2,2),(2,2),(2,-2),(2,2),(-2,-2),(-2,2)]`. This is a logical evaluation trace, not required persistent game state.
  - Discriminates: reading only P instead of the nonzero blocker produces all-zero shadows; deduplicating targets violates the exact 24-entry ordered evaluation trace.
  - runtime_observed: no
  - (a) Inputs: source radius 8, center `(4,4)`, owner resolves with integer position `(0,0)` and room bounds `[-32,32)` by `[-32,32)`, grid origin `(-32,-32)`, width 64, all 4096 cells 0; no adjacent rooms; P all 0 except P[0,0]=4; B all 16; cache invalid with storage for 9 samples; invoke cache build. Trace: partial clear leaves center 4; query 4096 positions row-major from(-32,-32) through(31,31), all 0; q=1 skips sweep; copy three-by-three. Outputs: B all 0, P center 4 and all other values 0; cache rows `[[0,0,0],[0,4,0],[0,0,0]]`, valid true; query log 4096 requests in the stated order; source radius/center unchanged. Draws: none.

- **LMR-V19** Complete cached, viewport-shadow and uncached rasters (LMR-R08, LMR-R13, LMR-R14, LMR-R16, LMR-R10; boundary: supplied-input stage)
  - Inputs: source kind 2, radius 8, center `(4,4)`, intensity 255, RGB `(10,20,30)`, valid cache `[[0,0,0],[0,4,0],[0,0,0]]`; quality 1; origin `(-1,-1)`; map all cells intensity 0/RGB 0/obstruction 0; colored-light false; empty contribution/cache-build log.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: N=2, starts(-8,-8), nine samples in row-major order. Only samples(0,0),(8,0),(0,8),(8,8) have positive unshadowed 95. Cache center 4 reduces the first to `(95*(8-2))>>3=71`; others remain 95.
  - Outputs: contribution log exactly `[(0,0,71),(8,0,95),(0,8,95),(8,8,95)]`, all with RGB `(10,20,30)`; selected map cells get those intensities and RGB 0; other cells unchanged; no cache-build request; source/cache unchanged.
  - Discriminates: ignoring shadow changes 71 to 95; rebuilding a valid cache adds an unwanted request; swapping traversal axes changes log order.
  - runtime_observed: no
  - (a) Inputs: source kind 0, radius 8, center `(4,4)`, intensity 255, RGB `(10,20,30)`; quality 2; origin `(-1,-1)`; map all intensity 0/RGB 0/obstruction 0; colored-light false; P all 0 except P[0,0]=4; B all 16; empty logs. Trace: viewport shadow fill writes central B square 0; partial P clear retains center 4; q=1 skips sweep; same four samples and attenuation. Outputs: contribution log exactly `[(0,0,71),(8,0,95),(0,8,95),(8,8,95)]`, all with RGB `(10,20,30)`; corresponding map intensities 71/95/95/95 and RGB 0; other cells unchanged; P center 4 retained; B central three-by-three 0, other B16. Draws: none.
  - (b) Inputs: source kind 1, radius 8, center `(4,4)`, intensity 255, RGB `(10,20,30)`; quality 1; origin `(-1,-1)`; map all intensity 0/RGB 0/obstruction 0; colored-light false; P all 0 except P[0,0]=4, B all 16; empty logs. Trace: uncached path uses no shadow and touches no scratch. Outputs: log exactly `[(0,0,95),(8,0,95),(0,8,95),(8,8,95)]`, all with RGB `(10,20,30)`; corresponding map intensities 95 and RGB 0; other cells and P/B unchanged. Draws: none.

- **LMR-V20** Failed owner resolution and coarse rejection order (LMR-R08, LMR-R15, LMR-R16; boundary: supplied-input stage)
  - Inputs: source kind 2, radius 8, center `(10000,10000)`, intensity 255, RGB `(10,20,30)`; quality 1; cache invalid with 9 samples of existing storage; owner lookup absent; P all 7; B all 0; origin `(0,0)`, map all intensity 0/RGB 0/obstruction 0; empty logs.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: dispatch requests cache build; ClearShadow clears top 16 rows; absent owner leaves cache invalid. Application coarse reject sees x-start 1249>48, returns before cache-valid assertion.
  - Outputs: one cache-build request, no map contribution, cache still invalid, map/B unchanged; P top 16 rows 0 and other rows 7; no fatal assertion.
  - Discriminates: checking cache-valid before coarse rejection would assert; skipping the failed builder would omit the partial scratch clear.
  - runtime_observed: no
  - (a) Inputs: same explicitly kind 2/radius 8/intensity 255/RGB `(10,20,30)`, invalid stored cache 9, absent owner, P all 7, B all 0, origin `(0,0)` and map all 0, but center `(4,4)`; quality 1. Trace: builder clears partial P then returns; raster intersects viewport and detects invalid cache. Outputs: fatal assertion and no successful raster result; no map-contribution request. Draws: none.

- **LMR-V21** Current-room collision masking in a reached lighting query (LMR-R12; boundary: supplied-input stage)
  - Inputs: current room A bounds `[4840,4880)` by `[5640,5680)`; present collision storage with grid origin `(4840,5640)`, width 40, selected unsigned 16 cell at linear index 51 equal 16; query `(4851,5641)` with mask 34. Other cells and adjacent-room list are unread and arbitrary; this current-room case does not consume them.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: query is inside current room; selected index `(5641-5640)*40+(4851-4840)=51`; cell 16 AND34=0.
  - Outputs: result 0 and blocker false; no collision-storage write. Captured relevant room/grid input bytes are unchanged across the call.
  - Discriminates: treating every nonzero raw collision cell as a blocker would return blocker true for 16; exact masked result 0 rejects it.
  - runtime_observed: yes (1.13c, 2026-10-27, [RT-LMR-COLLISION-MASK](../research/spec-evidence/draw-light-map-and-light-radius.md#runtime-collection-2026-10-27))

- **LMR-V22** Temporary skill-light replacement and association availability (LMR-R17, LMR-R01, LMR-R03; boundary: reached light stage)
  - Inputs: valid player owner identity 5, fixed-point position `(0,0)`; selected ready missile fields `Light=7`, RGB `(9,20,30)`; primary association P, present temporary storage associated with T; distinct live kind-1 sources P and T, each current/target 16, center `(4,4)`, intensity 100, RGB `(1,2,3)`, no cache; source sequence `[T,P]`; successful allocation of N; empty operation log. Unread other source fields are arbitrary and unchanged.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: create N at radius 1 -> current/target 8; set target to `7*8=56`; replace temporary T with N; remove T.
  - Outputs: operation log `[Create(N),SetTarget(N,56),AttachTemporary(N),Remove(T)]`; sequence `[N,P]`; N kind 1, owner player 5, center `(4,4)`, intensity 255, RGB `(9,20,30)`, current 8, target 56, no shadow cache. Primary P and all its fields unchanged; temporary N. This rejects assigning the new source to the primary slot or initializing current radius to 56.
  - runtime_observed: no
  - (a) Inputs: identical owner, sources, associations and sequence, but `Light=0`; successful allocation is available but unrequested. Trace: zero byte exits before creation. Outputs: empty operation log, `[T,P]`, both associations and source fields unchanged. Draws: none.
  - (b) Inputs: same valid owner/position, `Light=255`, RGB `(9,20,30)`, primary P with the fields above; temporary storage absent and no T; sequence `[P]`; successful N allocation. Trace: initial radius 1 ->8, target capped to `18*8=144`; missing storage returns no displaced association. Outputs: log `[Create(N),SetTarget(N,144)]`, sequence `[N,P]`; N current 8/target 144, other initialized fields as above; primary P unchanged; temporary storage remains absent, N is unattached. Draws: none.
  - (c) Inputs: valid player owner identity 1, fixed-point position `(313950208,372211712)`; selected ready missile record 58 with `Light=7`, RGB `(255,178,64)`; primary association P, present temporary storage with absent association; existing source sequence `[H]+S`, where S is an opaque existing suffix. H is kind 1, owner player 1, center `(38328,45440)`, current 8, target 72, intensity 255, RGB `(255,178,64)`, removal condition false and no shadow cache. P is kind 0, owner player 1, same center, current/target 104, intensity 255, RGB `(255,255,255)`, removal condition false, no shadow cache, and no successor. Other existing fields and the contents of S are arbitrary and unread by this operation. Supply successful fresh source N allocation, Expansion mode, no RNG seed or draws.
    - Trace: logical position conversion gives `(313950208 >>> 13)+4=38328`, `(372211712 >>> 13)+4=45440`; requested initial radius 1 leaves current 8; `7*8=56` is the final target; insert N before H and attach it to the empty temporary association. No displaced source is returned for removal. The transient initial target before its update is not a runtime assertion of this case.
    - Outputs: sequence prefix `[N,H]+S`, with H's successor still S; temporary association N and primary P. N has kind 1, owner player 1, center `(38328,45440)`, current 8, target 56, intensity 255, RGB `(255,178,64)`, removal condition false and no shadow cache. H, P and owner position are unchanged. No tail traversal or final pixel is asserted; S remains opaque.
    - Discriminates: appending N leaves H as head; omitting the old-head link loses H; assigning primary instead of temporary changes P; setting current directly to 56 disagrees with observed current 8.
    - runtime_observed: yes (1.13c, 2026-10-27, [RT-LMR-CAST-LISTHEAD](../research/spec-evidence/draw-light-map-and-light-radius.md#runtime-temporary-light-list-head-confirmation))

- **LMR-V23** Primary projectile light selection and zero-radius force (LMR-R18, LMR-R01; boundary: reached light stage)
  - Inputs: valid missile owner identity 50, fixed-point position `(0,0)`; selected ready missile fields `Light=7`, RGB `(9,20,30)`; enable condition true, suppression false, force byte 0; primary P and present temporary T, distinct live kind-1 sources with current/target 16, center `(4,4)`, intensity 100, RGB `(1,2,3)`, no cache; source sequence `[P,T]`; successful allocation N; empty operation log. Unread other fields are arbitrary and unchanged.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: remove P and clear primary; create N with requested radius 7, yielding both radii 56; attach primary N.
  - Outputs: log `[Remove(P),ClearPrimary,Create(N),AttachPrimary(N)]`; sequence `[N,T]`; primary N, temporary T unchanged; N kind 1, owner missile 50, center `(4,4)`, intensity 255, RGB `(9,20,30)`, current/target 56, no cache. This rejects temporary-light growth or attachment for the projectile.
  - runtime_observed: no
  - (a) Inputs: same initial state, enable true/suppression false, but `Light=0`, force 0. Trace: both zero skip operation. Outputs: empty log, sequence `[P,T]`, both associations and source fields unchanged. Draws: none.
  - (b) Inputs: same initial state, enable true/suppression false, `Light=0`, force 1. Trace: remove P, clear primary, request radius 0 -> no source. Outputs: log `[Remove(P),ClearPrimary,CreateRequest(0),AttachPrimary(absent)]`; sequence `[T]`; primary absent, T and its association unchanged; no allocated N. Draws: none.
  - (c) Inputs: same initial state and `Light=7`, force 0, but enable false, suppression false. Trace: disabled branch removes P. Outputs: log `[Remove(P),ClearPrimary]`; sequence `[T]`, primary absent, T unchanged, no creation request. Draws: none.
  - (d) Inputs: same initial state and `Light=7`, force 0, enable true, suppression true. Outputs: exactly the log, sequence and associations of (c). Draws: none.

- **LMR-V24** Separate temporary clearing and dual cleanup (LMR-R19; boundary: reached light cleanup stage)
  - Inputs: valid unit with primary P and present temporary storage associated with T; distinct live kind-1 sources P and T, current/target 16, center `(4,4)`, intensity 100, RGB `(1,2,3)`, no cache; source sequence `[T,P]`; operation `ClearTemporary`; empty operation log. Other fields are unread, arbitrary and unchanged.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: detach T first, then remove T.
  - Outputs: log `[ClearTemporary,Remove(T)]`; source sequence `[P]`, primary P and its fields unchanged, temporary absent.
  - runtime_observed: no
  - (a) Inputs: same initial sources and associations, operation `CleanupLights`. Trace: remove P, clear primary; clear temporary then remove T. Outputs: log `[Remove(P),ClearPrimary,ClearTemporary,Remove(T)]`, empty source sequence, both associations absent. Draws: none.
  - (b) Inputs: valid unit, primary P with the fields above, temporary storage absent; source sequence `[P]`, operation `ClearTemporary`. Outputs: empty log, sequence `[P]`, primary unchanged, storage still absent. Draws: none.

- **LMR-V25** Prepared component maximum and level overrides (LMR-R20; boundary: prepared monster light selection)
  - Inputs: monster class 10; MonStats count 200, selected record has BaseId 5 and MonStatsEx 1; MonStats2 count 2 with selected record present. Its component variant counts are 2 for component 5, 1 for components 6, 7, 8, and 0 for every other component. Component-code indices are component 5 variants `[1,2]`, component 6 `[3]`, component 7 `[4]`, component 8 `[5]`; other list bytes are unread/arbitrary. Present visual-component state selects variant 1 for component 5 and variant 0 for all others. CompCode count 6, entries `[zero code,"axe ","swd ","hvy ","bad ","fla "]`. Unique ready item records by code: `axe ` light 9, `swd ` light 13, `hvy ` light 255, `fla ` light 4, `sbw ` light 21; every other code absent. Current room present, level identity 1. Other record fields/entries are unread and arbitrary; no light association or source state is changed by this operation.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: contributions in component order are zero except component 5 selects `swd ` ->13 and component 8 selects `fla ` ->4; component 6 `hvy ` is skipped despite its item light 255; component 7 `bad ` has no item.
  - Outputs: radius 13; no writes. This rejects summing component radii or treating the skipped armor code as an item contribution.
  - runtime_observed: no
  - (a) Inputs: identical fixture except selected BaseId 0 and level identity 40. Trace: component 5 variant 1 overrides `swd ` with `axe ` ->9; component 6 is unlisted and still skipped; components 7/8 override to zero. Outputs: radius 9, no writes. Draws: none.
  - (b) Inputs: identical fixture except BaseId 170 and level identity 74. Trace: component 5 remains `swd ` ->13, component 6 overrides `hvy ` with `sbw ` ->21, component 8 overrides to zero. Outputs: radius 21, no writes. Draws: none.
  - (c) Inputs: fixture (a) with level identity 39, then independently 75. Trace: neither level enables overrides. Outputs: radius 13 in each case, no writes. Draws: none.
  - (d) Inputs: initial fixture with visual-component state absent. Outputs: radius 0; no component or item lookup and no writes. Draws: none.
  - (e) Inputs: initial fixture but ready `swd ` item light 200. Trace: unsigned maximum of 200 and 4. Outputs: radius 200, no writes. Draws: none.

- **LMR-V26** Monster light initialization, special replacement and source cap (LMR-R20, LMR-R21, LMR-R01; boundary: reached monster light initializer)
  - Inputs: complete initial prepared-record/component/room fixture of LMR-V25, plus selected MonStats Align 0; selected MonStats2 `light=7`, RGB `(9,20,30)`; valid monster identity 50, fixed-point position `(0,0)`; present client context, level-8 small-light condition true, small-light exclusion flag false. Primary P and present temporary T are distinct live kind-1 sources, current/target 16, center `(4,4)`, intensity 100, RGB `(1,2,3)`, no cache; source sequence `[T,P]`, stored base radius 6; successful allocation N; empty operation log. Other fields are unread/arbitrary and unchanged.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: component radius 13 exceeds template 7; level 1 retains r=13; create N, remove P, clear primary, attach N, store 13.
  - Outputs: log `[Create(N),Remove(P),ClearPrimary,AttachPrimary(N),StoreBaseRadius(13)]`; sequence `[N,T]`; N kind 0, owner monster 50, center `(4,4)`, intensity 255, RGB `(9,20,30)`, current/target 104, no cache; primary N, stored base radius 13; temporary T unchanged.
  - runtime_observed: no
  - (a) Inputs: same initial state but level identity 8. Trace: maximum 13 is replaced by 3. Outputs: same ordered log with stored radius 3; N current/target 24, other initialized values unchanged; sequence `[N,T]`, primary N, temporary T unchanged. Draws: none.
  - (b) Inputs: same initial state at level 8 but Align 1. Trace: alignment excludes the radius-3 replacement. Outputs: identical to initial case, stored radius 13 and current/target 104. Draws: none.
  - (c) Inputs: initial fixture at level 1 but visual-component state absent and template light 0. Trace: maximum 0 selects no operation. Outputs: empty log; sequence `[T,P]`, primary P, temporary T and stored base radius 6 unchanged; no allocation. Draws: none. The same fixture at level 8 with the supplied context/condition/exclusion/Align 0 instead creates kind 0 N with radii 24 and stored radius 3, with the ordered replacement effects of (a).
  - (d) Inputs: initial fixture with ready `swd ` item light 200. Trace: r=200; creation caps source radius to 18, while bookkeeping retains 200. Outputs: ordered log with `StoreBaseRadius(200)`; N current/target 144, other initialized values unchanged; stored base radius 200. Draws: none.
  - (e) Inputs: fixture (a), but client context absent. Trace: reached level-8 context check fatally asserts before source creation. Outputs: fatal assertion, no successful initialization result, empty source-operation log; stored base radius remains 6, source sequence remains `[T,P]`, and primary P and temporary T retain their associations and fields. Draws: none.

- **LMR-V27** Monster callback recomputes its base (LMR-R20, LMR-R22, LMR-R02; boundary: reached stat-89 callback)
  - Inputs: complete initial prepared/component fixture of LMR-V25, but room level identity 8; selected MonStats2 light 7, RGB `(9,20,30)`, MonStats Align 0. Valid monster identity 50 has primary kind-2 source P, current radius 16, target 24, center `(4,4)`, intensity 100, RGB `(1,2,3)`, valid shadow cache with 25 zero samples; source sequence `[P]`. Stored base radius 3, context present, level-8 small-light condition true, exclusion flag false; callback modifier 2. No temporary association. Other fields are unread/arbitrary and unchanged.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: component 13 exceeds template 7; recomputed base is 13 despite stored 3 and initializer condition; wrapped sum 15 is positive; immediate setter assigns `15*8=120`.
  - Outputs: P current/target 120; cache invalid and discarded; sequence, center, intensity, RGB and stored base radius unchanged. This rejects using stored radius 3 or repeating the initializer's special radius-3 selection.
  - runtime_observed: no
  - (a) Inputs: identical initial source/cache fixture but modifier -20. Trace: base 13 plus -20 is -7 ->0; zero request is an immediate-setter no-op. Outputs: P current 16/target 24, cache valid with the same 25 zero samples; all other state unchanged. Draws: none.

- **LMR-V28** Overlay uses distinct initial and target requests (LMR-R23, LMR-R01, LMR-R03; boundary: reached overlay light stage)
  - Inputs: newly initialized overlay with no light association; prepared `InitRadius=3`, `Radius=7`, RGB `(9,8,7)`; valid player owner identity 1 with alternate-registry condition false, fixed-point position `(655360,1310720)`; source sequence `[P,T]`, unit primary P and temporary T; existing distinct P/T fields are arbitrary and unchanged. Successful allocation is supplied. Each case uses this independent initial state except its stated field overrides. The light-operation log starts empty and logs creation and target requests in order.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: target 7 enables creation at radius 3; `3*8=24`; position shifts give `(84,164)`; unequal raw requests call target setter with 7, giving `7*8=56`.
  - Outputs: overlay associated with new N; sequence `[N,P,T]`; N kind 1, owner identity 1 with alternate-registry condition false, center `(84,164)`, current/target `24/56`, intensity 255, RGB `(9,8,7)`, removal false and invalid/absent shadow cache; unit primary P and temporary T and their fields unchanged. Ordered log `[Create(3),SetTarget(7)]`.
  - Discriminates: creating directly at target 7 gives current 56; ignoring the target gives target 24; attaching to the unit's temporary association changes T. All are rejected.
  - runtime_observed: no
  - (a) Inputs: `InitRadius=3`, `Radius=0`. Trace: zero target skips the whole stage. Outputs: overlay association absent, sequence `[P,T]`, primary P and temporary T unchanged; no allocation; log empty. Draws: none.
  - (b) Inputs: `InitRadius=7`, `Radius=7`. Trace: enabled creation gives 56; equal raw requests skip target setter. Outputs: associated N with current/target `56/56`, other new-source fields as above, sequence `[N,P,T]`; primary P and temporary T unchanged; log `[Create(7)]`. Draws: none.
  - (c) Inputs: `InitRadius=0`, `Radius=7`. Trace: creation returns absent without allocation; store absent association; unequal positive target reaches the fatal absent-source assertion. Outputs: fatal assertion, no successful stage result; overlay association remains absent, sequence `[P,T]`, primary P and temporary T unchanged; log `[Create(0),SetTarget(7)]`. Draws: none.
  - (d) Inputs: `InitRadius=3`, `Radius=-1`. Trace: nonzero negative target enables creation; unequal target request is a nonpositive setter no-op. Outputs: associated N current/target `24/24`, other new-source fields as above, sequence `[N,P,T]`; primary P and temporary T unchanged; log `[Create(3),SetTarget(-1)]`. Draws: none.
  - (e) Inputs: `InitRadius=0`, `Radius=-1`. Trace: creation returns absent; unequal nonpositive target setter returns before checking source. Outputs: overlay association absent, sequence `[P,T]`, primary P and temporary T unchanged; no allocation or fatal assertion; log `[Create(0),SetTarget(-1)]`. Draws: none.
  - (f) Inputs: `InitRadius=-1`, `Radius=-1`. Trace: enabled nonpositive creation returns absent; equal raw requests skip setter. Outputs: overlay association absent, sequence `[P,T]`, primary P and temporary T unchanged; no allocation or fatal assertion; log `[Create(-1)]`. Draws: none.
  - (g) Inputs: `InitRadius=20`, `Radius=21`. Trace: creation caps to 18, giving 144; unequal raw requests still call the setter, which also caps to 18. Outputs: associated N current/target `144/144`, other new-source fields as above, sequence `[N,P,T]`; primary P and temporary T unchanged; log `[Create(20),SetTarget(21)]`. Draws: none.

### 8.2 Observation vectors

- **LMR-O01** Observe creation and target-radius growth on an owned 1.13c source. Compare source-state values to LMR-V01/V04/V07 and record only reached cases; [procedure](../research/spec-evidence/draw-light-map-and-light-radius.md#observation-procedures).
- **LMR-O02** Observe a player radius modifier and color change, recording callback selector, old/new input, source state and resulting map separately; [procedure](../research/spec-evidence/draw-light-map-and-light-radius.md#observation-procedures).

## 9. Acceptance and dependencies

LMR-V01 through LMR-V28 define acceptance for the supplied successful-allocation/source/record, room/grid, scratch/cache and initialized-map operations LMR-R01 through LMR-R23, including reached temporary/projectile/monster/overlay light application and cleanup, complete source-raster enumeration and ordered map writes. These vectors establish the prepared-input lighting stages, not a complete original-game producer-to-displayed-scene claim. LMR-Q01 limits wider scheduler eligibility, LMR-Q02 limits source/equipment/record producer integration, LMR-Q03 names generic failure infrastructure, and LMR-Q04 names world/base-map producers and graphics consumption. Passing the supplied-input vectors does not settle those owners. LMR-V21 supplies matching gameplay observation for the current-room collision-mask case, and LMR-V22(c) for positive temporary-light creation/attachment with no displaced source; pending LMR-Q05 limits all remaining gameplay-confirmation claims, not use of the verified supplied-input operations. A table-cell or idle-memory read alone cannot establish gameplay observation acceptance.

**Depends on:** [numeric foundations section 1](core-numeric-and-rng-foundations.md#1-scope), for 32-bit wrapping, shifts, truncation and fatal-assertion semantics; [unit foundations section 3](core-record-layouts-and-ids.md#3-inputs-and-data), for semantic unit categories and identity; GDA-R03a/GDA-R03b/GDA-R03c in [game-data records](core-game-data-table-records.md), for binding selected ready object, missile and monster light fields through numeric byte descriptors, overlay initial/target requests through numeric 32-bit descriptors, and CompCode codes through direct four-byte conversion. GDA-R02 owns arbitrary compiled-record correspondence. Supplied collision maps, base environmental map state and scratch/cache contents are explicit inputs; producer integration is not supplied by a guessed default.

**Depended on by:** no current behavioral spec is established as consuming this new contract.

## 10. Related specs

- [Other drawing boundaries](coverage.md#draw): room and ambient lighting, palette shifts, and screen fades are separate from this local light-source contract.
- [World clock](drlg-world-clock.md): environmental timing and intensity inputs, outside local source-radius computation.
- [Tainted Sun](quest-tainted-sun.md): quest-owned environmental changes.
- [Game-data records](core-game-data-table-records.md): shared table-loader contracts; prepared object fields use the cited schema/conversion rules, while arbitrary compiled correspondence and other source producers remain LMR-Q02.

## Appendix A. Provenance

Evidence build is original Windows x86 Lord of Destruction Expansion 1.13c. Static readings dated 2026-10-27 and the prepared-overlay continuation dated 2026-10-29 use original D2Client and D2Common instructions, with reconstructions checked against original bytes where available. Original artifact identities are [LOD113C-D2CLIENT, LOD113C-D2COMMON and LOD113C-FOG](../research/source-manifest.md). LMR-V21 reproduces one current-room collision-mask case, and LMR-V22(c) reproduces one temporary-light front-insertion/attachment case on an owned 1.13c game on 2026-10-27; relevant loaded instruction identity and exact observation coverage are recorded in the companion. The [companion](../research/spec-evidence/draw-light-map-and-light-radius.md) gives the function chains, reading boundaries, exact observation coverage and remaining observations. No final rendered scene or cross-patch parity is established by these vectors.
