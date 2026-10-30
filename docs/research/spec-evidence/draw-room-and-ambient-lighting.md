# Room and ambient lighting: evidence

Contract: [room and ambient lighting](../../specs/draw-room-and-ambient-lighting.md).

## Evidence identity and reuse

Read date: 2026-10-29. Original x86 Expansion LoD 1.13c artifacts are LOD113C-D2CLIENT and LOD113C-D2COMMON in the [source manifest](../source-manifest.md). Fresh whole-file fingerprints agree exactly: D2Client is 1093632 bytes, SHA-256 `dd8bc6025de921216a97c17f97cd1a50fbb85926e838ec60e13451448836d906`; D2Common is 679936 bytes, SHA-256 `59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10`.

The 2026-10-27 light-map record's viewport root and rebuild order are reused for precisely D2Client `0x8B110 -> 0xA9E30`, on the matching original image. The new reading follows `0xA9E30 -> 0xA9C90` before `0xA8880` obstruction fill and `0xA9B50` source traversal. Earlier collision and temporary-light observations do not observe this contract and retain their original scope in the [light-map companion](draw-light-map-and-light-radius.md).

Read source reconstructions checked against original bytes first where available; other relevant routines were read as instruction views and corroborated by fresh disassembly of matching original images. Function descriptions below are analyst descriptions, not original export names. Labels found in analysis databases were not accepted as identities.

## Public function references

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2CLIENT | `0x8B110`, `0xA9E30` | — | reached viewport rebuild and ordering | RAL-R06 | Reused 2026-10-27 dispatch reading; original-image listing correspondence checked 2026-10-29 |
| LOD113C-D2CLIENT | `0xA9C90` | — | initialize map from focus room and ordered neighbors | RAL-R06 | Complete applicable instruction path; original-image listing |
| LOD113C-D2CLIENT | `0xA9A20` | — | resolve area override, level and stored environment light | RAL-R01..RAL-R04 | Complete selector/control/output reading; level-8 changing branch explicitly gated |
| LOD113C-D2CLIENT | `0xA8560` | — | fill inclusive rectangular map region | RAL-R05 | Complete instructions and original-image listing |
| LOD113C-D2CLIENT | `0x524D0` | — | level-8 special tint and intensity | RAL-R03a, RAL-R03b | Fixed/inactive branch to output read; changing intensity math not fully established |
| LOD113C-D2COMMON | `0x3C000` | 10826 | room level identity | RAL-R01, RAL-R02 | Source reconstruction checked against original bytes; null returns 0; successful metadata chain required |
| LOD113C-D2COMMON | `0x3C6A0` | 10636 | room light-field adapter | RAL-R01, RAL-R04 | Complete instructions; delegates to `0x44D90`; original-image listing |
| LOD113C-D2COMMON | `0x44D90` | — | selected level-definition lighting bytes | RAL-R01 | Complete instructions/original listing; valid index selects stride 156 record; invalid index reaches invalid read rather than a successful zero fallback |
| LOD113C-D2COMMON | `0x374E0` | 10731 | read act environment intensity | RAL-R04 | Complete source reconstruction checked against original bytes; absent act asserts; absent environment returns 0 |
| LOD113C-D2COMMON | `0x37520` | 10189 | copy act environment RGB | RAL-R04 | Complete source reconstruction checked against original bytes; absent act asserts; absent environment leaves outputs unchanged |
| LOD113C-D2COMMON | `0x3BA20` | 10383 | current room adjacency/count outputs | RAL-R06 | Complete source reconstruction checked against original bytes; outputs room reference collection and count |
| LOD113C-D2COMMON | `0x3BA60` | 10561 | room rectangle metadata copy | RAL-R06 | Complete source reconstruction checked against original bytes; copies room bounds to the caller's scratch values |
| LOD113C-D2COMMON | `0x2FE10` | 10331 | focus unit room lookup | RAL-R06 | Reused complete reading in 2026-10-27 light-map companion; this contract consumes resolved room, not room generation |
| LOD113C-D2COMMON | `0x6D9B0` | — | level-definition schema and selected load request | RAL-R01 | Complete relevant schema and field publication path; direct byte descriptor names/offsets corroborated in original image; decoder integration remains GDA-R02 |
| LOD113C-D2COMMON | `0x371C0`, `0x376F0` | — | environment colour/intensity producers | RAL-R07 | Read as remaining general-lighting leads; float/runtime dependencies incomplete, no new formula promoted |

Import resolution on the matching original PE directories: Client `0xC1DE -> Common #10826 / 0x3C000`; `0xC59E -> #10636 / 0x3C6A0`; `0xC5A4 -> #10731 / 0x374E0`; `0xC5B6 -> #10189 / 0x37520`; `0xC3DC -> #10383 / 0x3BA20`; `0xC5B0 -> #10561 / 0x3BA60`; the reused focus room thunk `0xC1F0 -> #10331 / 0x2FE10` is recorded by the light-map companion. Common `0x3C6BB` calls `0x44D90` directly. The selected generic table load at Common `0x6E0A6` calls `0x5EF40`, consumed through GDA-R01/GDA-R02/GDA-R03a/GDA-R03b/GDA-R03c rather than assumed decoder execution.

## Selection and output rationale

Client `0xA9A20` takes room in ESI and the area-override condition in EAX; four caller-supplied outputs receive intensity, red, green and blue. Its first three local colour bytes are zeroed. Nonzero condition queries room identity. Signed identity comparisons isolate 107 and 108 (`>106`, `<=108`); 8 uses the special helper. The normal override reads intensity at Client RVA `0x119610`, RGB at `0x119614..0x119616`. The other two special levels test a nonzero condition at `0x11C2E4` and phase -1 at `0xF2130`, selecting 160/(255,64,48), otherwise zero. These are neutral supplied conditions: their producers and colloquial quest/event meanings have not been assigned.

The three RGB zero tests at `0xA9AAE..0xA9AC3` decide fallback independently of intensity. Level fallback calls Common `#10636`; this adapts room metadata through `room+0x10`, then the level metadata chain at `+0x58` and `+0x1D0` selects the ready record. Common `0x44D90` reads bytes `+0x88/+0x89/+0x8A/+0x8B`. If level RGB is all zero, Client `0xA9B11` requests environment RGB, then `0xA9B1D` intensity. Current act lives at Client RVA `0x11C3B8`; Common getters follow act `+4`, then environment intensity `+0x0C` and RGB `+0x18..+0x1A`. These original layouts locate evidence only.

Common `0x6D9B0` defines the level-definition stride 156. Its instructions `0x6DFCF..0x6E038` construct direct byte conversion kind 4 fields: original name pointers `0x8F920` = `Intensity`, `0x8AEE8` = `Red`, `0x8AEE0` = `Green`, `0x8AED8` = `Blue`, destinations 136..139. Names were read from original immutable bytes, not guessed from offsets. The named `leveldefs` request is at `0x6DF9D` with string RVA `0x8F8E8`. The [game-data companion](core-game-data-table-records.md) already records generic text-request substitution from `leveldefs` to `levels`; named schema plus GDA direct conversions support ready-field identity. The archive read of `leveldefs.txt` failed, as expected from the separate request name; this is not evidence that the compiled table is missing.

At level 8, Client `0x524D0` first tests suppression RVA `0x11C2E0`; true returns 0. Otherwise context RVA `0x11C0A0` is required before active byte `0x10A591` is tested. Inactive returns 0 without colour writes. Active writes (255,64,48); phase RVA `0xF212C` equal to -1 returns 80. Other phases execute integer arithmetic followed by Client import `0xBFFE -> Fog #10083 / 0x16730` (a table lookup after adding 128 and masking by 511), a floating multiply and Client runtime conversion helper `0xCACC4`. Only the fixed/inactive successful branches are promoted; the phase table, multiplier, conversion and producers still gate RAL-R03b. A missing nonsuppressed context is outside the successful valid-context domain; no safe-null fallback is asserted.

Client `0xA8560` uses inclusive loop counts `end-start+1`. It zeroes the cell's first dword, then writes intensity/red/green/blue at offsets 4..7. RAL-R05 exposes the known obstruction clear and lighting bytes semantically; unidentified auxiliary cell storage is not a prescribed storage layout. Its caller forms a word using bitwise OR, without first narrowing intensity to a byte. Consequently intensity high bits can set colour bits; RAL-V06 distinguishes this from independent byte assignments. For ordinary 0..255 intensity the four outputs are the supplied tuple exactly.

Client `0xA9C90` initializes all 2304 cells with the focus room's light. It then walks adjacent pointers/count from Common `#10383`, skipping only identity equality with the focus room. It resolves each neighbor's light before requesting bounds through `#10561`. Signed origin-relative starts and `start+extent` ends are clipped to 0..47. Start comparisons are strictly greater than 48; resulting empty ranges write nothing. No `extent-1` occurs. All original byte-run instructions on these claimed paths were checked in the fresh original-image listings, including branches, calls and final byte writes. Source light accumulation follows this initialization, so a final rendered map after local sources is not evidence of the initial base tuple alone.

## Rejected discovery leads and search limits

The coverage candidate's Client `0xA8520` is an obstruction-byte lookup with outside-window sentinel 1, not an ambient intensity selector. Client `0x4F770` has only located callers in the audio-environment operation `0x4F860`; it returns a floating factor based on room metadata and is an audio investigation lead, not evidence of map initialization. Neither name was used to establish lighting behavior. The output/root chain above establishes the actual owner. No claim of complete absence or all-six-rung negative search is made.

The general environment read at Common `0x371C0` interpolates colour through selected clock tables and float conversion. `0x376F0` includes an Act-IV source selector and general trigonometric path in addition to WCK-R05a's two established branches. These were read to locate the remaining producers, with fresh original-image listings, but the complete float/table dependency execution is unverified. RAL-Q03 is a behavior gate on their production, not on the consumer with supplied stored light.

## Table-data records

TD-RAL-LEVELS, 2026-10-29: dedicated vanilla launch key `lod-113c`, reported family LoD/version 1.13c, original Game.exe profile SHA-256 `74fe9c092a521f7710392548c82f81544e531107db2358617e83818874db40a2` from the launch inventory. Effective `data/global/excel/levels.txt` was read through the running Storm loader at offset 0, length 32000 (32000 returned of total 66615), then a preserved sufficient range 0..11999 (12000 returned, same total). The full header and named rows below are complete in the preserved range. Header-based, case-preserving TSV column selection was used; rows are selected by explicit `Id` and `Name`, not physical row number.

| Record | Row identity | Column | Raw cell | Interpretation / support |
|---|---|---|---|---|
| TD-RAL-LEVELS | Id 1, `Act 1 - Town` | Intensity | `0` | Named text-cell value only |
| TD-RAL-LEVELS | Id 1, `Act 1 - Town` | Red | `0` | Named text-cell value only |
| TD-RAL-LEVELS | Id 1, `Act 1 - Town` | Green | `0` | Named text-cell value only |
| TD-RAL-LEVELS | Id 1, `Act 1 - Town` | Blue | `0` | Named text-cell value only |
| TD-RAL-LEVELS | Id 8, `Act 1 - Cave 1` | Intensity | `0` | Named text-cell value only |
| TD-RAL-LEVELS | Id 8, `Act 1 - Cave 1` | Red | `255` | Named text-cell value only |
| TD-RAL-LEVELS | Id 8, `Act 1 - Cave 1` | Green | `255` | Named text-cell value only |
| TD-RAL-LEVELS | Id 8, `Act 1 - Cave 1` | Blue | `255` | Named text-cell value only |

These single named cells are verified in 1.13c data. They do not prove active compiled-record correspondence, special-area eligibility or gameplay execution. Synthetic tuples in RAL-V01..RAL-V08 do not claim these are their loaded game inputs. Effective full-file hash, physical supplying archive and complete override inventory were not captured; the dedicated launch used the vanilla launch key with no requested data overrides.

TD-RAL-LVLTYPES, same profile/date: effective `data/global/excel/lvltypes.txt`, offset 0, length 5000, returned 5000 of total 10719. The complete first header line contains `Name`, `Id`, `File 1` through `File 32`, `Beta`, `Act`, `Expansion`, and no Intensity/Red/Green/Blue column. This is a complete-header observation, not complete-table absence or proof of no indirect lighting influence. No table rows are reproduced or used as ambient rules.

The separate `data/global/excel/leveldefs.txt` read at offset 0/length 128 failed at archive open with error 2. It establishes absence only for that effective requested path in this launch, not the absence of level definitions or their compiled form.

## Runtime attempt

### Runtime confirmation, 2026-10-29

**RT-RAL-LEVEL-READ.** A new owned `lod-113c` process loaded the existing level-1 research Sorceress on Normal in Rogue Encampment, stationary at subtiles (4863,5653). This process is separate from the initial title/menu attempt below. No save edit or gameplay input mutation was used.

The reached D2Common ordinal 10636, RVA `0x3C6A0`, called the already verified selected-record reader `0x44D90`. A bounded 16-record capture reports dropped 0, with complete output and input windows. Three records (sequence 0, 9 and 10) use the same captured focus-room fixture and resolved metadata/table windows. Sequence 9 is RAL-V09's distinguishing replacement case; sequence 0 corroborates the same level-1 tuple from the normal D2Client `0xA9ADF` ambient caller. Sequence 9's caller is D2Client `0x9EE29`; no wider meaning is assigned to that additional caller.

At sequence 9, the room reference resolves its room metadata through `+0x10`; that metadata's `+0x58` selects a level record whose signed identity at `+0x1D0` is 1. The ready collection has 137 entries, checked before and after the bounded collection. The captured ready-table base contains record 1's four bytes at `1*156+136..139`, all zero. Output pointers select intensity at the fourth byte of the captured output window, red at its first, green at its third and blue at its second. Thus output bytes `20 B1 00 00` mean semantic tuple `(0,32,0,177)`, not green 177. The after-image is four zeros; red 32 and blue 177 were actually replaced. Stack cleanup is 20 bytes, consistent with the five stack arguments; this alone does not establish any additional register arguments. Room, selected metadata and prepared-table windows are complete and unchanged. Sequence 0 and 10's initially zero outputs also remain zero. Other records whose derived windows do not match this known room fixture are not assigned its input state.

Read-only derived capture addresses chased the already located metadata-pointer and table-pointer cells. They did not write an input or substitute an internal layout for the semantic contract. The normal-game reader's unconditional output writes and selected-index computation are the existing complete static boundary; only this zero-tuple case is promoted. RAL-R02..RAL-R07, RAL-V01..RAL-V08 and the broader observation procedures do not inherit a match.

Loaded original instructions were compared byte-for-byte before hooking: Common `0x3C6A0`, 48 bytes, SHA-256 `8767cb252f6648aada79ce84228bb837e31fe7a3fea6492806b9306d54b003a3`; Common `0x44D90`, 96 bytes, SHA-256 `c6d4297836cd3dd8029afe86b66baa3548db9cd96b92ddfb30368b230e48ef81`. Both match the manifest's original image. A whole loaded-module hash was not captured; the identified executing reader slices are checked, not an assertion about every loaded instruction.

**Named text correspondence.** A fresh Storm-loader read on this process requested effective `data/global/excel/levels.txt`, offset 0, length 12000, returned 12000 of 66615 bytes. The full header and complete explicit Id 1, Name `Act 1 - Town` row were parsed by their column headers. Named raw cells `Intensity=0`, `Red=0`, `Green=0`, `Blue=0` match this selected ready level-1 tuple. This is a measured four-cell correspondence only. Effective full-file hash, physical supplying archive, complete override inventory and compiled-resource provenance remain uncaptured; GDA-R02/RAL-Q02 remain open beyond this named case. No whole table or row is published.

**Other captures and limits.** Original getters corroborate stored environment RGB and intensity with complete input windows. A small 16-record Client `0xA9A20` capture with dropped 0 also returns `(255,255,255,255)` for a zero level tuple while the captured environment has intensity 255 and RGB white. These records do not by themselves establish complete cross-operation request ordering or every override input; they remain partial corroboration of RAL-R04 without a runtime promotion. A larger four-target attempt reached its 4096-record cap and evicted 3584 records per target; its retained slices are preserved separately as incomplete and never used to claim order or absence. This was corrected by the bounded no-drop captures above rather than treated as a matching full fallback vector.

Raw windows, file bytes, unsuccessful setup attempts and integrity records remain private. RAL-V09 and RAL-R01's label change require a fresh full reader; the earlier eight-vector reviews remain historical for their recorded boundary.

### RAL-O03

Use an owned original Expansion 1.13c process and a level-1 Sorceress on Normal in Rogue Encampment. Remain stationary during normal drawing. Observe Common ordinal 10636 / `0x3C6A0` and its selected reader `0x44D90`; capture the room-to-level selection, loaded collection count, the selected record's four lighting fields, every output pointer and complete before/after output bytes. Check selected level identity 1 and a zero tuple. Compare only the actually measured prior output values; nonzero prior bytes must become zero, with metadata/table unchanged. Record a complete case even when no bytes change, but do not use an unchanged-zero case to claim a nonzero replacement. Compare the reader's loaded code with the identified original image. This procedure confirms the prepared reader, not room-light fallback or the later map/pixels.

2026-10-29: an initial owned vanilla LoD 1.13c attempt reached title/menu but did not enter gameplay. No gameplay vector was observed. Its effective archive reads remain table-data observations only.

### RAL-O01

Use an owned original Expansion 1.13c client with a research character, any class/level, Normal difficulty. Enter Rogue Encampment, then move toward an adjacent room boundary through ordinary movement. Observe one reached Client `0xA9E30 -> 0xA9C90` initialization. Capture focus room, ordered neighbors/count, origin, each room level and rectangle, returned tuples from `0xA9A20`, and the full map immediately after `0xA9C90` returns, before obstruction/local-source operations. Include all entry inputs and output writes, no dropped records, and source-profile correspondence. Recompute RAL-R06 from that snapshot and compare every selected cell. A screenshot or a map after local-source accumulation does not meet this procedure.

### RAL-O02

On the same owned profile, capture Client `0xA9A20` entry/output with room metadata and selected ready level record. Record all override conditions and RGB/intensity globals, current act/environment and all fallback calls. For a zero-RGB level tuple with no winning override, verify environment RGB precedes environment intensity. For a naturally reached nonzero-RGB tuple, verify no environment call and unchanged selected intensity even when zero. Use actual ready fields; do not synthesize game memory or assume a text cell equals its compiled record. Name the exact observed cases and comparison limits.

## Review and checks

### Full blind-implementer pass 3 after runtime capture, 2026-10-29

Fresh independent full review of the nine-vector contract and RAL-R01's narrow observed case: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 15, COVERAGE 0, HARDEST 3. The reader recomputed every case and confirmed that section 3 declares the existing output tuple, R01 replaces all four values, and V09 distinguishes nonzero prior values from the observed zero replacement. V01's nonzero-blue fixture remains static-only; no broader producer or fallback observation is implied. The supplied-stage versus full-integration acceptance boundary remains explicit.

HARDEST R06 ordering/clipping, R02/R03a sentinel selection, and R05 inclusive packed channels need no change: their exact rules and V07/V08, V02..V04, V05/V06 provide the needed operations and assertions. The existing fifteen gated references stay documented; they do not become defaults. No behavioral, input, vector, expected output, gate, ownership or acceptance change followed this full pass. Reader review is complete for this runtime revision; document status remains draft.

### Final full blind-implementer pass 2, 2026-10-29

Fresh independent full review of all eight vector families including V03(c), V04(c), V08(c)/(d): MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 15, COVERAGE 0, HARDEST 3. Every numeric output and ordered request log matched. The reader confirmed that V08(c)'s equivalent empty-fill optimization imposes no internal API requirement. The fifteen GATED references retain the documented behavior, integration, supplied-stage and observation limits; they are not missing input decisions.

HARDEST dispositions, all no change: R06/V07/V08 explicitly gives request order, signed clipping and overwrite precedence; R02/R03a/V02..V04 separates each special-area eligibility and RGB sentinel decision; R05/V05/V06 gives inclusive iteration and the exact bitwise channel effects. All supplied cases are executable without guessing these choices. At that handoff no rule, input, vector, expected output, acceptance, ownership or gate changed after the full review, and its reader stopping rule was met. The later observed RAL-V09 case and RAL-R01 label/input clarification require the fresh full pass recorded separately; the draft status and explicit user approval gate remain.

Final mechanical lint for this spec, companion and gap file on 2026-10-29: 0 errors, 0 warnings. The same result held for each of the other three lighting contracts. No implementation or plan is approved by these documentation checks.

### Full blind-implementer pass 1, 2026-10-29

Fresh independent reader: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1, GATED 15, COVERAGE 3, HARDEST 3. All original eight vector families recomputed correctly. Existing changing-level-8/environment producer, compiled-data and gameplay observation boundaries remain explicit.

DISCRIM V08(c): no change, because clipping to an empty rectangle at column 48 and rejecting that fill after the already logged room-light/rectangle requests produce identical specified map and request effects. The logical fill operation imposes no required internal function call or implementation API. The reader itself identified that equivalence; neither observable behavior nor an assertion was removed.

COVERAGE dispositions: V04(c) now tests level 8 with overrides disabled and missing context, demonstrating bypass without an assertion; V03(c) tests the enabled sentinel override at level 108; V08(d) tests clipping both axes to index 47 and overwriting exactly the final cell. These add executable cases without changing any rule. HARDEST R06 ordering/overwrites, R02 special-level/RGB fallback and R05 inclusive packed-channel writes require no rule edits: their order and outputs are explicit, with V07, V02/V03 and V05/V06 respectively distinguishing the alternatives. New cases require a fresh full review of the resulting contract, recorded separately.

2026-10-29 writer self-review checked rule/input/output inventory, prepared-input domains, distinct behavior/integration/observation gates, all eight vector families, original-image correspondence and public/private evidence separation. Initial mechanical lint after the complete artifact set and input inventory correction: 0 errors, 0 warnings. The final full reader pass above completes this contract's reader gate; no earlier light-map review was substituted for it. Document-level draft status reflects remaining behavior/integration/observation gates and does not imply user approval.
