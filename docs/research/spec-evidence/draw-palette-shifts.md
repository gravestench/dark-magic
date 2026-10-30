# Evidence: monster palette shifts

Spec: [Monster palette shifts](../../specs/draw-palette-shifts.md).

## Identity and reuse

Evidence build is original Lord of Destruction Expansion 1.13c, Windows x86. Artifact `LOD113C-D2CLIENT` in the [source manifest](../source-manifest.md) has SHA-256 `dd8bc6025de921216a97c17f97cd1a50fbb85926e838ec60e13451448836d906`. The 2026-10-29 original-image instruction listing reports that same digest. Relevant D2Common identity is `LOD113C-D2COMMON`.

The existing light-map companion and 2026-10-27 private light-map audit were checked for overlapping claims. Those records concern light-map/radius sources and cannot confirm palette remapping or selector precedence. Their observations and dates remain unchanged. World-clock and Tainted Sun contracts were read as neighboring boundaries: their colour/intensity requests do not establish the indexed replacement tables specified here. This is new targeted evidence, not renewed confirmation of those records.

## Function references

Descriptions below are analyst descriptions, not established original export names.

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2CLIENT | 0x666D0 -> 0x6CC00 -> 0x6C490 | none established | Visibility/type/component rendering chain | PLS-R01..PLS-R04 | Selected call-path reading; surrounding rendering is outside the reproduced stage |
| LOD113C-D2CLIENT | 0x53A30 | none established | Monster table selection | PLS-R01..PLS-R04 | Complete selection instructions; original-image listing and dispatch-table bytes checked 2026-10-29 |
| LOD113C-D2CLIENT | 0xA58E0 | none established | Appearance-selector accessor | PLS-R01..PLS-R04 | Complete instructions, required unit/graphics assertions and selector return |
| LOD113C-D2CLIENT | 0x01230 | none established | Primary monster-record lookup | PLS-R06 | Reconstruction checked against original bytes; label claiming a skill lookup rejected by monster-table use |
| LOD113C-D2CLIENT | 0x01260 | none established | Linked extended monster-record lookup | PLS-R04, PLS-R05 | Complete instruction reading; class/count, signed link/count, record address |
| LOD113C-D2CLIENT | 0x53B20 | none established | Alternate table preparation | PLS-R05 | Complete instructions; original-image listing covers all writes and loop bounds |
| LOD113C-D2CLIENT | 0x53C30 | none established | Monster palette asset loader | PLS-R06 | Partial loader reading: transforms before publication; file-helper behavior remains unresolved |
| LOD113C-D2CLIENT | 0x537A0 | none established | Cached-asset lookup | PLS-R06 | Reconstruction checked against original bytes, id search through linked cache entries; full lifetime not established |
| LOD113C-D2CLIENT | 0x53850, 0x538F0 | none established | GreenBlood and random-transform initialization | PLS-R06 | File strings, requested sizes and publication leads; complete error/lifetime paths remain unresolved |
| LOD113C-D2CLIENT | 0xA7780, 0x539E0 | none established | Creation selector and random-selector helper | PLS-R07 | Partial producer reading; not a verified state/unique precedence contract |

## Dispatch and output rationale

### Selection

The visibility rendering path at `0x666D0` calls type rendering at `0x6CC00`, which calls component rendering at `0x6C490`. The latter at preferred VA `0x6FB1C6F4` calls `0x53A30` when its current replacement-table input is zero and retains that returned selection for rendering. This anchors the table consumer in presentation rather than a skill dispatch. The rest of component composition is outside this contract; a selected reference does not prove the downstream pixels.

At `0x53A30`, null/non-monster returns zero. Class selects an eight-byte cache entry; its palette-asset reference is read before the selector accessor. Missing cache returns zero even with otherwise invalid global selector. The accessor `0xA58E0` asserts on absent unit/graphics and returns graphics field `+0x38`. Consumer eligibility prevents a null unit reaching that accessor.

The unsigned `selector <= 7` dispatch reads bytes at RVA `0x53B14`: `0,0,1,1,1,1,1,1`. The two jump targets at `0x53B0C` are preferred VAs `0x6FB03A78` and `0x6FB03ADC`. Thus selectors 0/1 run the class check, 2..7 the base/alternate branch. The class check admits exactly unsigned identities `0x16B` and `0x16C` (363/364). Base bank addresses are asset `+4+256*selector`.

Outside unsigned 0..7, global availability at preferred VA `0x6FBCC2D4` selects the global path. Subtract 8 and require signed result 0..29, else assert/exit. Return global base `0x6FBC7808+256*(selector-8)`. This branch precedes the alternate-mode test at `0x6FBCC380`. Random selection does not advance a seed.

The alternate branch resolves the linked extended record with `0x01260`, tests byte `+0x11C == 2`, and returns asset `+0x804+256*selector`; otherwise it returns the base bank. `0x01260` checks signed class against primary count, reads signed word `+0x18` of the primary record, checks that link against extended count and resolves the extended record. MonStats and MonStats2 ownership/linkage correspond to the existing game-data table contract; the specific byte-to-column mapping is PLS-Q03 rather than a guessed `Utrans` meaning.

The original consumer does not bound the class-cache access here, nor add a bank clamp on the no-global path. The supplied-input contract's valid domain is explicit; PLS-Q04 retains the unanswered malformed/producer domain.

### Preparation

The loader at `0x53C30` reads base data, calls `0x53B20`, then assigns class identity and publishes the asset to its cache/list. Preparation therefore has a real caller and an output path. The current record deliberately does not verify every loading helper or imply asset availability from a supplied vector.

`0x53B20` inlines class/link resolution, checks extended byte `+0x11C == 2` and otherwise returns without copying. At `0x53B99`, `rep movsd` copies `0x200` dwords (2048 bytes) from asset `+4` to `+0x804`, preserving all eight base banks. It then executes eight outer iterations, 256 bytes each. Each candidate comes from translation[unsigned base byte]; a zero candidate skips its destination write, retaining the initial copy. Nonzero writes go to the alternate bank. The base table is never overwritten, and remapping is not iterated to a fixed point. The original-image instruction listing explicitly confirms all previously byte-encoded instructions in this arithmetic/loop path.

### Rejected labels and unresolved producers

The initial candidate `0x40190` is a forwarding wrapper to an import with literal zero; its label does not identify palette selection and it is excluded from the reproduced path. `0x524D0` and `0xA9A20`, despite fade/unit-colour labels, participate in environment/room tint consumers and belong with ambient lighting, not this indexed monster-table stage.

Original strings identify `Data\Global\Monsters\RandTransforms.dat`, `Data\Global\Monsters\GreenBlood.dat` and monster `\Cof\palshift.dat` path components. These strings establish requested asset names, not the complete file-read/error/cleanup contract or actual effective bytes. No archive table value is copied or labeled data-verified.

The creation path `0xA7780` contains selector writes, a call to `0x539E0`, and later monster/variant overrides. The random helper yields 9..38 after resetting/advancing a unit seed, while the creation slice later resets selectors at least 30 to 2. These partial observations are leads, not new verified RNG rules: omitted local helpers and later producers prevent a complete appearance-precedence claim. In particular they do not establish `States.colorshift`/`colorpri` or `MonStats2.Utrans` semantics. PLS-Q02 records that remaining research rather than importing community interpretations.

A targeted D2Common descriptor probe at `0x56C10` associates original string `localBlood` (RVA `0x91D74`) with destination field `+0x11C`. This narrows PLS-Q03 to a concrete candidate column; complete original instruction/compiler and effective compiled-record correspondence were not verified in this probe, so it does not promote an archive binding or replace the supplied transformation-kind input.

## Search record and limitations

On 2026-10-29 the existing coverage RVAs were checked, caller references were searched in original D2Client code, and the output anchors for returned table references and alternate-bank writes were traced. The selection dispatch table was read directly from the PE image. Label-only leads were separated from computing functions. This targeted path search stopped with identified consumer/preparation functions and incomplete asset/appearance producers; it is not an exhausted six-rung absence claim.

An isolated read-only import of the original D2Client image produced targeted instruction listings corresponding to the manifest fingerprint. Source forms, unsuccessful analysis setup attempts and exact collection commands are separately preserved in the private audit. Original-image disassembly corroborates static reading, not gameplay confirmation.

## Observation procedure

For PLS-O01 use an owned vanilla Expansion 1.13c game, a Normal character and nearby rendered monsters. Record character class/level, location, monster class, current appearance selector, class-cache presence, global/alternate mode values, selected linked record kind and actual returned table. Trace the selection stage and capture complete bank contents used by a measured case; compare to the semantic input/output rules without assuming a selector producer. For eligible alternate preparation, observe the original loader and record all eight base banks, the 256-entry translation and the resulting alternate banks before/after preparation. A naturally encountered subset can corroborate only that subset. Do not modify memory or directly invoke game functions to manufacture confirmation.

| Vector | Source profile | Date | Module + RVA | Records | Result |
|---|---|---|---|---|---|
| PLS-O01 / PLS-V01..PLS-V06 | dedicated vanilla LoD Expansion 1.13c launch; title screen; empty character inventory | 2026-10-29 | D2Client + 0x53A30 requested | 0 | No gameplay capture and no match; the attempt stopped before entry observation. |

The initial title-screen attempt neither confirms nor contradicts the static reading. A separate gameplay campaign below provides the later observations. No raw character, game data or capture is included here.

## Runtime observations

On 2026-10-29 a separate owned vanilla Expansion 1.13c launch created a level-1 Sorceress through the ordinary character-creation interface, entered Normal in the Rogue Encampment and walked into Blood Moor. No gameplay-memory write, direct game-function invocation or forced hook supplied an input. Source profile is the vanilla launch key; complete loaded-module/override identities were not captured. Before recording, the following loaded code ranges exactly matched the corresponding bytes in the manifest's D2Client image:

| D2Client RVA / bytes | SHA-256 of measured range | Purpose |
|---|---|---|
| 0x53A30 / 240 | `b1844b6e01b5f1d1b05645da4a4ee39a82c68b2bc29a19ad60cdf9202f6cd51c` | Complete selection body and dispatch table |
| 0x53B20 / 272 | `1792e0bb91c5681354722c661629080d682bf9e088e988ab1df5ba85183eb89d` | Preparation instructions |
| 0xA58E0 / 64 | `ddbb7e435690c4e2db8336e74322cd50dc473c231145d9f5aa223037c7166471` | Selector accessor |

The loaded first 448 bytes at `0x6CC00` and first 96 bytes at `0x6C490` also matched the original image. These establish the specific saved-argument path used to capture the actual input unit. At `0x6CC00`, ESI holds the unit received in ECX and remains that value when the function calls `0x6C490`, passing the same unit as its first stack argument. The latter saves its incoming ESI. At the later call to `0x53A30` (return RVA `0x6C6F9`), that saved unit is the second stack dword after the return address. Capturing that unit and its graphics pointer therefore measures this invocation's unit, not a global focus unit or an inferred neighboring stream. The measured unit type also agrees with the caller's cached type field. No records were paired across targets by thread/depth or independent sequence numbers.

An initial requested 48-dword capture returned only 16 dwords and no requested deep-stack windows. Those 16 records are incomplete for unit/selector attribution and do not confirm a vector. A new 32-record capture used the supported saved-unit location. It returned both complete requested windows (128 unit bytes and 64 graphics bytes), no lost records, return cleanup 0 and no changed bytes in either window for the present-player cases. Records 5, 14 and 23 each measured player type 0, class identity 1, appearance selector 2 and return value zero. The routine returned normally. These independently match all specified PLS-V07 assertions. The unused monster cache/record/bank arrays are not asserted by that early-return vector.

| Vector | Source profile | Date | Module + RVA | Records | Result |
|---|---|---|---|---|---|
| PLS-V07 | owned vanilla Expansion 1.13c, Normal, level-1 Sorceress in Rogue Encampment | 2026-10-29 | D2Client + 0x53A30; caller return +0x6C6F9 | 3 matching present-player cases within a valid 32-record capture; dropped 0 | Match: no table, normal return, type/class/selector retained; PLS-R01 present-player case only |
| PLS-O01, PLS-R01 monster-cache branch | same launch, town rendering | 2026-10-29 | D2Client + 0x53A30 | same 32-record capture | Partial: classes 149, 150, 152 and 155 with selector 2 returned no table. A later read found absent cache entries; no in-call cache snapshot. These are not credited as the cache-precedence vector PLS-V01(c). |
| PLS-O01, PLS-R05 | same launch after normal walking into Blood Moor | 2026-10-29 | D2Client + 0x53B20; caller return +0x53EAD | 1; dropped 0 | Partial: class 19 measured from its caller's saved class, 512-byte input window unchanged. The linked record's transformation kind was 1 and a later full asset read had eight zero alternate banks. No complete alternate-bank before-image; PLS-V05/PLS-V06 not matched or promoted. |

Preparation attribution uses the loader's saved class at callee entry stack displacement `+0x3C`, within the captured 16-dword range; EDX directly supplies its asset. The class header is assigned only after preparation, so its zero entry value was not mistaken for class 0. The partial input window holds the header and first 508 base bytes, not all eight banks. Separate live reads after return recovered the complete 4104-byte asset in two contiguous ranges, the selected 308-byte extended record and the 256-byte translation. Those read-only results support the reported partial correspondence; they do not manufacture a before-image or match the synthetic bank arrays of PLS-V05/PLS-V06. The full selected/global/alternate-table outputs of PLS-R02 through PLS-R04 were not confirmed.

The recorder's actual 16-dword and 512-byte limits were checked against returned records; requesting a larger count did not expand the capture. Bounded captures used at most 32 records and reported no eviction. These records change only PLS-R01's label and add PLS-V07. All earlier static rules, synthetic vector results, unresolved integrations and unobserved branches retain their status.

## Review and checks

### Fresh full reader after runtime delta, 2026-10-29

Fresh independent full review of all seven vector families and PLS-R01's present-player runtime case: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 15, COVERAGE 2, HARDEST 3. All numerical/table results and PLS-V07's early-return/retained-metadata assertions recomputed correctly. Existing asset, selector-producer, malformed-domain and gameplay observation limits remain explicit.

COVERAGE R03 with out-of-bank selector and random tables disabled: no change, because that domain is explicitly PLS-Q04 rather than an available safe-fallback contract; neither a guessed clamp nor an invented bank is added. COVERAGE R04 invalid class: no change, because the shared linked-record resolution is exercised for invalid classes by V06 and its absent-record selection result by V04(f). HARDEST R03 unsigned split/signed difference, R04 signed class/link eligibility and R05 sparse one-pass all-bank translation require no edits: their rules and V03, V04, V05 state and discriminate those operations. No behavioral, input, vector, expected output, ownership, gate or acceptance change followed this reader. Review is complete for the seven-vector runtime revision; the spec remains draft and user approval is not inferred.

Final mechanical lint after the ambient sibling existed and public method wording was normalized: 0 errors, 0 warnings, 2026-10-29. The original method remains reconstruction checked against original bytes plus matching original-image instructions.

Self-review on 2026-10-29 checked rule/input/output coverage, arithmetic and exact synthetic-bank results, supplied-input boundaries and provenance separation. The initial lint returned two errors for links to the concurrently written ambient spec before it existed, and three public-method warnings. Final lint is recorded separately after that sibling exists. The spec remains draft independently of review because the stated asset/selector integrations and gameplay observations are unresolved.

Historical independent full blind-implementer review before the runtime delta, 2026-10-29: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 14, COVERAGE 1, HARDEST 3. All six then-present vector families recomputed correctly. The GATED list retains R06/R07, Q01..Q05, the six supplied-stage vector boundaries and O01; it does not treat prepared tables as verified asset/selector production.

COVERAGE R04 invalid MonStats class bounds: no change. The linked-record resolver's invalid-class outcome is already discriminated in preparation V06, and selection V04(f) exercises the resolved-absent fallback. These are the same explicitly stated record-resolution contract and no additional output category is uncovered. HARDEST R03 unsigned-domain/signed-difference bounds needs no change: R03 and V03 distinguish the split and fatal upper/lower results. HARDEST R04 class/link/alternate conditions needs no change: R04 and V04 separately state and test those conditions. HARDEST R05 eight-bank sparse translation needs no change: V05 supplies every entry by a total formula and distinguishes zero preservation and one-pass translation. The subsequent PLS-V07 addition changes a vector and partial-acceptance assertion, so a new fresh full blind reader is required; the earlier pass does not cover that delta. Document status remains draft and user approval is not inferred.
