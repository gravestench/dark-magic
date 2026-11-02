# Positional sound evidence

Spec: [Positional sound and falloff](../../specs/sound-3d-position-and-falloff.md)

## Identity and methods

Read date: 2026-10-29. Original Windows x86 Expansion 1.13c. [Source manifest](../source-manifest.md) identifies LOD113C-D2CLIENT, LOD113C-D2COMMON and LOD113C-D2SOUND. D2Client original-image SHA-256 was checked during this investigation: dd8bc6025de921216a97c17f97cd1a50fbb85926e838ec60e13451448836d906, matching the manifest. D2Sound references use the matching original bundle identified by the manifest; no different-version code was used.

The initial reading found no matched reconstruction for the three discovery candidates or the radius/coordinate helpers. It read disassembly and retained every gameplay assertion as gated. The continued reading checked the match registry again and read matched reconstruction first for the available D2Sound option getters/position wrapper and D2Common position accessor. Original-image native instruction listings corroborated the earlier byte-encoded portions in the client coordinate/cache/radius/update functions and ordinary D2Sound arithmetic/property routines. This closes the earlier decoding limitation for the now-verified stages; alternate CPU arithmetic paths remain outside that verification boundary. No private reconstruction filenames are required to use the public references below.

Descriptions are analyst descriptions. Existing labels such as GetLastError, GetSoundDistanceFactor and SetSound3DDistance do not establish meaning.

## Current function map

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2CLIENT | 0x5f020 | none established | Enabled audio update | S3D-R02 through S3D-R07 | Original listing and disassembly; reaches active coordination |
| LOD113C-D2CLIENT | 0x88dd0 | none established | Active-source coordination | S3D-R02 through S3D-R07 | Full disassembly read; scalar helper paths and output calls located; lifecycle integration gated |
| LOD113C-D2CLIENT | 0x88a70 | none established | Source request preparation | S3D-R03, S3D-R07 | Disassembly; initial relative placement and references |
| LOD113C-D2CLIENT | 0x5d7a0 | none established | Voice creation | S3D-R07 | Disassembly; invokes source-property updater after initializing cache |
| LOD113C-D2CLIENT | 0x5d1d0 | none established | Source-property updater | S3D-R02, S3D-R05, S3D-R07 | Full disassembly and original listing; distance stage isolated after option scaling |
| LOD113C-D2CLIENT | 0x5dc70 | none established | Outer-radius lookup | S3D-R02 | Complete helper plus original jump table and constants |
| LOD113C-D2CLIENT | 0x5dce0 | none established | Inner-radius lookup | S3D-R02 | Complete helper plus original jump table and constants |
| LOD113C-D2CLIENT | 0x87f30 | none established | Relative position | S3D-R03, S3D-R07 | Complete disassembly plus original listing; ordinary branch verified, special branch gated |
| LOD113C-D2CLIENT | 0x88350 | none established | Raw position and clamped distance key | S3D-R04 | Complete disassembly and original listing; finite path has no callees |
| LOD113C-D2CLIENT | 0x88440 | none established | Source context factor | S3D-R07 | Disassembly read; consumer/factor integration remains gated |
| LOD113C-D2CLIENT | 0xcacc4 | none established | Floating conversion | S3D-R05 | Complete disassembly; representable ordinary truncation, byte-run corroboration |
| LOD113C-D2CLIENT | 0x5dfd0 | none established | Sound table loader | S3D-R02 | Original descriptor names/offset stores checked; one loaded cell corroborated |
| LOD113C-D2COMMON | 0x2ffe0 | 10628 | Presentation-position accessor | S3D-R03 | Matched reconstruction read; ordinary valid unit output boundary |
| LOD113C-D2SOUND | 0x68d0 | 10020 | Position property wrapper | S3D-R07 | Matched reconstruction; conditionally invokes backend position method |
| LOD113C-D2SOUND | 0x6ce0 | 10051 | Source factor property wrapper | S3D-R06, S3D-R07 | Disassembly/original listing; stores factor, conditional backend property route |
| LOD113C-D2SOUND | 0x6e20 | 10008 | Amplitude request conversion | S3D-R06 | Complete disassembly/original listing through backend scalar call |
| LOD113C-D2SOUND | 0x8830 | none established | Logarithmic scalar conversion | S3D-R06 | Complete disassembly/original listing; ordinary division path |
| LOD113C-D2SOUND | 0xaccc | none established | Floating conversion | S3D-R06 | Complete disassembly; ordinary representable truncation |
| LOD113C-D2SOUND | 0x6d80 | 10028 | Pan request conversion | S3D-R07 | Complete disassembly/original listing; integration gated |
| LOD113C-D2SOUND | 0x8b50 | 10047 | Stereo gain/pan calculation | S3D-R07 | Complete wrapper disassembly; power-library and alternate CPU paths remain gated |
| LOD113C-D2SOUND | 0x9cf0 | none established | Power execution selector | S3D-R07 | Disassembly read; selects deeper numerical-library implementations not verified |
| LOD113C-D2SOUND | 0x9ac0, 0x9af0, 0x9a90, 0x9b20 | 10011, 10001, 10067, 10059 | Option scalar getters | S3D-R07 | Matched reconstruction read; globals only, not error semantics |
| LOD113C-D2SOUND | 0x9aa0, 0x9ad0, 0x9a70, 0x9b00 | none established | Option scalar setters | S3D-R07 | Matched reconstruction read; option producer integration not accepted here |

## Dispatch and boundary mapping

D2Client 0x5f020 calls ambient 0x4f860, music 0x45660, other audio coordination, then active-source coordination 0x88dd0 at return site 0x5f057, before incrementing audio counter 0x11c268. The enabled-audio test gates this operation. Within 0x88dd0, source-reference preparation reaches 0x87f30 and 0x88350 at return site 0x89071. Voice-creation 0x5d7a0 and active-source update return site 0x893a2 reach 0x5d1d0. These actual call targets and output writes establish identity beyond analyst names.

Ordinary position preparation calls import thunk 0xc21a, resolving to D2Common ordinal 10628/RVA 0x2ffe0, first for the local player read from client global 0x11bbfc and then for the emitter. The accessor copies two integer presentation coordinates from the unit's appropriate static or dynamic path. Ordinary code subtracts listener components from emitter components, converts to float, doubles vertical difference and writes height 640. Identity 2599 takes an extra floating transform before subtraction; the ordinary supplied-input rule deliberately excludes that path until its complete source-specific meaning is resolved.

0x88350 copies input x/y/z to source offsets 0xc/0x10/0x14. It clamps temporary x/y at original constants -2000 and 2000, computes their squared sum and stores a binary 32-bit float at 0x18. Native opcodes corroborate the earlier D8 C9 / D8 CB / DE C1 byte runs. The output uses floating intermediate products and final binary 32-bit storage. For the verified integer-valued x/y domain every clamped product and sum is exact even at single-precision significand width; fractional cache arithmetic remains gated. Retained raw coordinates remain unclamped. This key is used for ordering/eligibility, while 0x5d1d0 separately computes sqrt of raw x/y.

The updater's radius helper return sites are 0x5d22d and 0x5d23a. Outer/inner constants correspond to original client constants at 0xd7544 through 0xd755c. Jump tables at 0x5dcc4 and 0x5dd30 were read as data from original bytes; a linear disassembler's apparent instructions in those table slots are not executed code. The read maps selectors exactly as S3D-R02. Both helpers dereference a null record on invalid lookup rather than returning a safe radius; the valid-record boundary is explicit.

The distance arithmetic lies at updater 0x5d405 through 0x5d488. It computes sqrt(rawX²+rawY²), conditionally divides (outer-distance)*precomputed amplitude by outer-inner, truncates through 0xcacc4, then applies the sound row's Volume byte via signed divide by 255. Its preceding music/master-option, sound-id exception, ducking and source-mode branches are distinct from the supplied distance-stage arithmetic. Its following software-stereo branch can further multiply by a computed gain; no scalar-stage vector claims that branch was bypassed in every real voice.

## Table binding and radius records

Original loader descriptor stores associate Falloff name reference 0xd55f8 with loaded offset 0x47 at 0x5e151/0x5e16a; record stride is 0x8e. Native name/store decoding also distinguishes Reverb 0x4b, Stereo 0x50, Tracking 0x51, Solo 0x52 and Music Vol 0x53. Reverb must not be mislabeled Stereo. These names locate consumers but do not prove an audible policy.

The original radius constants are exact binary 32-bit values: outer map 400/700/1000/1500/2000 with default 700, inner map 60/60/200/400/2000 with default 60. Unsigned compare sends all selectors above 4, including negative signed bit patterns, to the defaults. No custom remapping is established.

### T01 effective text cell

Initial collection, 2026-10-29: a reported vanilla LoD 1.13c runtime supplied effective data/global/excel/sounds.txt through Storm. Offset 0, requested length 4000, returned 4000 from total size 505304. The encoded response body was decoded; complete header and target row each had 25 columns. Header-name parsing selected Sound=cursor_pass, explicit Index=1; raw Falloff text cell was 1, decimal integer 1. Coverage is sufficient for this cell, not full table absence claims. Effective-file fingerprint, physical supplying archive and complete override inventory were not captured.

A separate collection on a reported vanilla LoD 1.13c runtime read client loaded sound-table base and count from globals 0x11c24c/0x11c250, obtaining count 4699. It read 17 bytes beginning at record index 1's field 0x47; the first four bytes decode to u32 1. The loader descriptor and loaded-field match corroborate the compiled selector for this named identity. This does not establish whole-table correspondence or runtime confirmation of a radius/attenuation vector. Heap addresses and original fixture bytes are preserved privately.

## Backend conversion and remaining device boundary

Client thunk 0xd300 resolves to D2Sound ordinal 10020/RVA 0x68d0; thunk 0xd2dc to ordinal 10051/RVA 0x6ce0; thunk 0xd2e2 to ordinal 10008/RVA 0x6e20. The position wrapper checks source validity, then source capability bit 2 before invoking the backend position method with x/y/z and a final zero. The factor wrapper stores the supplied float at source offset 0x3c, with a conditional property-interface call when device mode equals 2. Its published analyst name does not prove a physical distance parameter.

0x6e20 stores requested amplitude at source offset 0x38; multiplies by backend scalar global 0x15994 and divides by 255; bypasses local (1-factor) multiplication only when source capability bit 4 and device mode global 0x16598 equals 2. Otherwise it multiplies by (1-factor) and truncates. For valid ordinary integer amplitude 0..255, 0x8830 returns -10000 at zero, zero at 255, otherwise 2000*log₁₀(v/255). The original minimum threshold is binary 32-bit 0.0001; because this stage's v is integer, no additional nonzero integer endpoint is hidden. Internal constants are at D2Sound 0x10970 (-10000), 0x10974 (1), 0x10978 (0), 0x1097c (threshold), 0x10980 (binary 64-bit -2000), 0x10988 (binary 64-bit zero) and 0x10990 (binary 64-bit one). The logarithm is expressed using x87 base-2 primitives; sign and ratio yield the stated log₁₀ expression. The ordinary divide branch is selected by global 0x165e8 being zero.

Ordinary conversion 0xaccc first rounds through an integer temporary, then corrects fractional sign to implement truncation toward zero before returning the 64-bit integer, whose low result is used here. The parallel client helper 0xcacc4 was also read. Exceptional/nonrepresentable conversion behavior is not accepted by the supplied numeric boundary.

The scalar is passed immediately to underlying buffer vtable slot 0x3c at 0x6e92. This establishes the requested value; the wrapper ignores the device result for behavioral state purposes. The actual device implementation and waveform are not established. Software stereo 0x8b50 uses distance-dependent logarithm, power and CPU-dispatched floating library routines. The wrapper and power selector 0x9cf0 were read, identifying deeper targets 0xb2d0 and 0x9d4d, plus alternate divide paths; full numerical equivalence remains S3D-Q02. It is not promoted from a pow label or a single device capture.

## Tracking and eligibility attempts

Active coordination 0x88dd0 tests global tracking-enable 0x11c3a0, loaded Tracking byte 0x51 and a nonempty source-reference list before recomputing reference positions. It compares raw squared x/y distances to choose a nearer reference, passes that position through 0x88350 and averages context factors obtained through 0x88440. The factor helper returns 0 or 0.5 based on Falloff selector, a sound group, environment information or an imported predicate. Complete reference lifecycle, tie behavior, specialized eligibility and predicate meaning remain S3D-Q03 rather than guessed named rules.

Existing voices call property updating only after state, start/duration, active-amplitude and radius checks; voice creation invokes it once after clearing cache. An unchanged-property fast path compares effective base amplitude, ordered factor equality and raw position bits before deciding whether to call the backend. Option/duck globals change at the end of coordination. These paths are located and read, but producer integration and every admission/retirement outcome belong to the remaining lifecycle contract. No claim of full frame-to-waveform parity follows.

## Earlier runtime attempt

On 2026-10-29, the original-profile capture used a level-1 Expansion Barbarian in Normal Rogue Encampment. An initial capture was invalid because D2Client was not yet loaded. Entry-only captures at D2Client 0x5dc70 and 0x5d1d0 contain four inspected records each, with zero reported dropped records and configured caps of 30. Calls came from 0x5d23a and 0x893a2 respectively. Required source-register values, floating return values and source-state outputs were not captured. Integer return registers from a floating-return helper were not treated as radii. These records remain inconclusive for numerical vectors.

## C02 source-property captures

Collection date 2026-10-29, reported vanilla LoD 1.13c profile, original Game executable identity matching the manifest, level-1 Expansion Barbarian in Normal Rogue Encampment. Source/module correspondence was recorded. These observations did not use live-state writes or synthetic function calls.

| Vector / boundary | Source profile | Date | Module + RVA | Records | Result |
|---|---|---|---|---:|---|
| S3D-V03, S3D-V04 base cases | Reported vanilla LoD 1.13c profile; table limits in T01 | 2026-10-29 | D2Client 0x88350, caller 0x89071 | 12 read, dropped 0, 12 cap | Matching finite cache inputs and complete 40-byte target window; two cases below |
| Additional cache preparation | Same runtime profile | 2026-10-29 | D2Client 0x88350, caller 0x89071 | 8 read, dropped 0, 8 cap | Matching repeated raw positions after movement, not additional vector promotions |
| Position producer preparation | Same runtime profile | 2026-10-29 | D2Common 0x2ffe0 | 12 read, dropped 0, 12 cap | Complete accessor input/output windows; no complete relative-position vector capture |
| Backend scalar preparation | Same runtime profile | 2026-10-29 | D2Sound 0x6e20 | 12 read initially, then 8; dropped 0 | Retained requested-amplitude writes observed; final backend method not captured |
| Backend conversion preparation | Same runtime profile | 2026-10-29 | D2Sound 0xaccc | 24 read, dropped 0 | Integer conversion outputs observed; partial matching only, no S3D-V07 promotion |

The first cache capture's alternating source cases were:
- Input position (320,80,640), retained raw position already (320,80,640), retained distanceKey already 108800; complete target window showed no changes. This matches S3D-V03's exact retained outputs. It is a repeated-value case, not proof that a previously different cache was overwritten.
- Input position (681,2726,640), retained raw position already (681,2726,640), retained distanceKey 4463761; complete window again showed no changes. Since 681²+2000²=4463761, this matches S3D-V04's raw-position preservation and clamped-key result. An unclamped-y key or a stored y=2000 would disagree. It does not establish the negative-clamp branch.
Later repeated inputs (320,-16,640) retained key 102656, and (662,2648,640) retained key 4438244; they are supplemental observations, not new accepted vectors.

Backend source captures included requested amplitude 99, retained prior amplitude 255, source factor zero and valid underlying buffer. Source offset 0x38 changed to 99. Backend scalar global 0x15994 was read as 255; device mode 0x16598 as 0. The conversion trace includes a call from return site 0x6e71 with input-source identity matching this source and integer result 99, followed by conversion from 0x6e90 yielding -821. This is consistent with 2000*log₁₀(99/255) truncated toward zero. The capture did not observe the final backend method and did not simultaneously capture every ordinary-division environment condition; S3D-R06 remains static and S3D-V07 remains runtime_observed no.

A separate relative-position entry capture was not retained; no position-vector assertion is credited to it. Exact fixture encodings and raw capture windows are retained privately.

## C03 moving-source cache writes

Collection date 2026-10-30, original Windows x86 Expansion 1.13c profile, level-1 Expansion Barbarian in Normal Rogue Encampment. Original module identity matches the source manifest. Normal movement supplied the changing source positions; no synthetic game-function invocation or gameplay-state write supplied the inputs.

D2Client 0x88350 was observed for 500 calls, with a configured record cap of 500, zero reported dropped/evicted records and no calls in flight at completion. Every call captured the complete 12-byte position input and complete 40-byte cache window before execution, plus changed output bytes. Of these, 87 changed the cache; all 87 had complete windows and matched the raw-position copy and bounded-key calculation. Fifty began with nonzero prior position. These are capped observations, not a cadence or completeness claim about the whole game session.

Two exact discriminating cases extend the executable vectors:

| Case | Input x/y/z | Prior retained x/y/z and key | Result retained x/y/z and key | Caller return RVA |
|---|---|---|---|---|
| S3D-V03(b) | 219, -86, 640 | 224, -80, 640; 56576 | 219, -86, 640; 55357 | D2Client 0x89071 |
| S3D-V04(b) | 679, 2716, 640 | 681, 2726, 640; 4463761 | 679, 2716, 640; 4461041 | D2Client 0x89071 |

The first case changes both horizontal coordinates and the key; z stays 640. The second changes both raw coordinates and the key while raw y remains above the positive clamp. Input windows were unchanged by the callee. The full cache windows distinguish actual writes from C02’s repeated-value cases; other unchanged bytes are retained. These observations support overwrite behavior for these exact cases and add no observation claim for fractional inputs, negative clamping, position production, attenuation, backend submission or audible output. Exact raw windows are retained privately.

## Continued route reading — 2026-10-30

The remaining integration gates retain their original scope. The following additional reads locate the still-unaccepted producer and numerical routes; they do not promote S3D-R07 or runtime labels.

D2Common 0x2ffe0 copies path integers at +0x4/+0x8 for unit types 2, 4 and 5, and +0x8/+0xc for other types. A null unit takes the assertion/error path; no substitute position is returned. D2Common 0xda40 supplies one dynamic-path producer: it stores the two incoming fixed-point dwords, logically shifts each right 11, stores the arithmetic-right-shifted wrapped difference divided by 2 at +0x8 and wrapped sum divided by 4 at +0xc. D2Common 0x2fed0 supplies a static-path producer: it stores supplied world x/y, sets presentation x to the wrapped difference shifted left 4 and presentation y to the wrapped sum shifted left 3. Both producer instruction sequences were corroborated against the original image. These located producers do not establish every path or eligibility condition: 0xda40 can reject a position when its room-check flag is set and no room is found. Matched reconstructions at 0x300e0 and 0x30120 also directly set dynamic presentation fields. Full callers, other producers and eligibility remain in S3D-Q01.

The full D2Client 0x5d1d0 property path applies sequential signed integer percentage scalings before distance: an option scalar when Music Vol is nonzero, the other volume option unconditionally, context percentage 0xe1374 except identities 1–15, 52–71 and 4673–4698, then percentage 0xe1378 except when Solo is nonzero. Each multiplication retains the low signed 32-bit product and each division by 100 truncates toward zero. Client import thunks 0xd26a/0xd270 resolve to D2Sound ordinals 10011/10001, RVAs 0x9ac0/0x9af0. Separate mode/position-option branches follow, so this list does not establish the final pre-distance amplitude. After distance and record Volume scaling, positions are multiplied by original binary 32-bit 0.0031250000465661287 and stored as binary 32-bit values. Zero property flags select software gain/pan; property flag bit 2 selects the position wrapper. Factor and amplitude wrappers follow. These located branches still require the complete option, context and device contract in S3D-Q02.

D2Sound 0x8b50 uses fixed binary 64-bit power base 10, with constants at 0x10968 and 0x10998–0x109bc corroborated from the original image. For finite source components its square-root distance is capped at 100; its logarithmic intermediate is stored as binary 32-bit before the power exponent calculation. The power selector 0x9cf0 checks the CPU selector, MXCSR exception masks and x87 exception masks before choosing 0xb2d0/0xb2e9 or 0x9d4d. The latter normal route loads/classifies through 0xc2a5, uses x87 fyl2x, then 0xc260 (frndint/f2xm1/fscale), and returns through control-word restoration or result handling at 0xc2fe/0xc349. The reachable SSE polynomial/table path and exact results across the selector still require closure; the pow label alone is not an accepted equivalence proof.

The alternate floating-division chain 0xa61c → 0xa509 → 0x9f40 was read completely and corroborated by original instructions. The final helper has no callees. It tests the denominator significand against original pattern bytes, can scale both extended operands by 15/16 (17/16 in a near-underflow branch) under temporary round-to-nearest extended precision, restores the control word, then divides. The original scales are at 0x159d0/0x159d4, with the pattern at 0x159c0. Reading these paths does not establish that their rounded results are identical to ordinary division. S3D-Q02 retains that question and actual floating-environment initialization.

D2Client 0x88dd0 chooses the first reference initially and replaces it only for strictly smaller ordered raw squared x/y distance; equal distances retain the first reference. It accumulates binary 32-bit context-factor sums and computes the mean, then moves the retained factor toward that target by at most original binary 32-bit 0.05000000074505806 per coordination, clamping at the target. This is not immediate assignment of the mean. Full eligibility, reference lifecycle and the factor predicate remain S3D-Q03. The factor fallback import at client 0xc9be resolves to D2Common ordinal 10839/RVA 0x314a0, reaching 0x30730 and collision-grid traversal 0x4ad20; no physical propagation model follows from that call chain.

## Observation procedure

Use matching original LoD 1.13c files and a level 1 expansion character in Normal Rogue Encampment. Capture source-position input and the complete retained position/key window around the cache operation while walking near a world source. For S3D-V03/V04 repeat the exact specified values; a repeated-value case is accepted only for its stated retained-output assertions. C03 adds exact different-prior-state cases S3D-V03(b)/V04(b); the zero-state subcases and negative-clamp case remain unobserved.

For ordinary placement, measure local player and emitter projected positions from the game accessor, the sound identity and complete resulting relative vector in the same operation. For attenuation, capture raw coordinates, both radii, pre-distance amplitude, record Volume and intermediate result. For backend scalar conversion, capture source validity, scalar, factor, routing condition, ordinary division selection and the final method's submitted scalar together. Distinguish these observations from hearing a quieter sound, and do not infer cadence from capped records.

## Review and status record

An initial fresh blind reader reviewed the fully gated positional draft before these new behavioral stages: MISMATCH 0, GUESS 0, GAP 4, DISCRIM 0, GATED 2, COVERAGE 8, HARDEST 3. Its GAP/HARDEST entries were existing producer/binding/tracking/observation gates and required no guessed defaults; coverage concerned unestablished behavior and stayed gated. This is historical review of that earlier boundary and does not approve the current expanded contract. A new full blind-reader review is required.

Initial lint returned 0 errors/7 warnings, then 0 errors/1 reviewed method warning after wording fixes. The current expansion closes actual static/observation slices, retains earlier inconclusive records, and changes labels only for the newly supported boundaries. No user approval is inferred.

Continued self-review: original D2Sound and D2Common images were rehashed and match the manifest. Numeric stages explicitly supply their arithmetic environment; its actual initialization is not inferred. Cache verification is limited to integer-valued x/y, whose bounded square/sum is exactly representable. Ordinary relative placement requires exactly representable differences. Final lint after the expansion returned 0 errors / 6 warnings, all retained public source-reading method descriptions; new fresh-reader review remains required.


## Fresh blind-implementer full pass 2 — 2026-10-29

MISMATCH 0, GUESS 0, GAP 1 documented observation limitation; all eight vector discriminators and ten subcases passed (DISCRIM 0 unresolved); GATED 4; no additional missing-branch COVERAGE items were listed; HARDEST 3. The reader independently recomputed all 18 executable cases.

GAP outcome: no change. V03/V04 and C02 already state that the retained values equal the expected outputs at entry; these captures confirm the named retained-output assertions and do not demonstrate overwriting different prior state. The unobserved zero-state subcases test the static write contract separately. No broader observation claim or label is added. HARDEST outcomes: no change for R05 signed truncation/product and equal-bound boundary, R06 separate truncations/logarithm versus actual device success, or R04 raw coordinates versus bounded key; their current text and vectors explicitly distinguish those operations. GATED outcomes: retain Q01 through Q04. Reader stopping rule met for the resulting supplied-stage contract; complete positional parity and user approval remain unclaimed.

## Supplied-input inventory and observation limits — 2026-10-30

The input inventory explicitly names the valid-record and valid-cache preconditions, prior retained position/key and requested amplitude, supplied inner/outer radius bounds and integer-valued x/y cache boundary. These declarations expose the starting inputs already supplied by the vectors; no behavior, vector result or evidence label changes. Fractional cache arithmetic remains S3D-Q02. C03 adds two complete moving-cache overwrite cases; the zero-state subcases and negative-clamp case remain unobserved.

Additional unexercised partitions include low-32-bit relative-difference wrap, positive-x and negative-y cache clamps, signed-product overflow, outer-radius equality, distance factor one and a nontrivial backend scalar. They do not establish additional observed cases. Raw position remains separate from the clamped key; attenuation retains signed truncation, low signed 32-bit product and the equal-bound division precondition; amplitude conversion retains its separate truncations and submitted-request/device boundary. S3D-R07 and S3D-Q01 through S3D-Q04 remain open, and the document remains draft.
