# Evidence: unit sound selection

Spec: [Unit sound selection](../../specs/sound-unit-sound-selection.md).

## Identity and method

Read date: 2026-10-29. Original Expansion 1.13c Windows x86. Artifact identities are [LOD113C-D2CLIENT and LOD113C-D2COMMON](../source-manifest.md). Both current images were independently hashed and match those manifest entries. Effective text-data source-profile limits are recorded in T01.

The initial client candidates and additional helpers have no byte-matched C entries; disassembly was read. Imported table helpers and the unit mode/path-field accessor have byte-matched C readings. Descriptions below are analyst labels, never original exported names.

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2CLIENT | 0x22490 | none established | Effective class accessor | USN-R01 | Disassembly; player conversion import unresolved |
| LOD113C-D2CLIENT | 0x22430, 0x223d0 | none established | Effective kind and mode accessors | USN-R01–USN-R04 | Disassembly; player conversion import unresolved |
| LOD113C-D2CLIENT | 0x22a70 | none established | Creature sound-record selector | USN-R01 | Disassembly; eligibility and data mapping incomplete |
| LOD113C-D2CLIENT | 0x22ca0 | none established | Four-selector cue request | USN-R02a, USN-R02b | Disassembly; full dispatch and playback unresolved |
| LOD113C-D2CLIENT | 0x22c10 | none established | Shared-cooldown variation candidate | USN-R04 | Disassembly corroborated against original-image decoding; caller semantics incomplete |
| LOD113C-D2CLIENT | 0x23c00 | none established | Movement phase and secondary request | USN-R03 | Disassembly; phase helper and callee chain incomplete |
| LOD113C-D2CLIENT | 0x895d0, 0x89640 | none established | Wrapped-distance and presentation-random helpers | USN-R03, USN-R04 | Disassembly corroborated against original-image decoding; parent contract incomplete |
| LOD113C-D2CLIENT | 0x22070, 0x220a0 | none established | Unit timing store/accessor | USN-R02a, USN-R02b, USN-R04 | Disassembly; labels misleading |
| LOD113C-D2COMMON | 0x2f1b0 | 10175 | Unit mode/path-field accessor | USN-R01 | Byte-matched reconstruction read; names misleading |
| LOD113C-D2COMMON | 0x501e0 | 10350 | Identity conversion candidate | USN-R01 | Disassembly read; remaining callee unresolved |
| LOD113C-D2COMMON | 0x52ff0 | 10876 | Direct sound-record lookup | USN-R01 | Byte-matched reconstruction read; table name unbound |
| LOD113C-D2COMMON | 0x52fa0 | 10917 | Indirect sound-record lookup | USN-R01 | Byte-matched reconstruction read; table name unbound |

## Search and anchored links

The coverage candidates were leads. Output-side inspection finds D2Client 0x22CA0, 0x22C10 and 0x23C00 calling 0x88A70, the common cue request lead. Incoming references locate 0x23170 and 0x24490 calling 0x22CA0, 0x23F20 calling 0x22C10, and 0x4F4A0 calling 0x23C00. The latter call follows 0x237E0 in the read local tail. A reference from 0x1C150 to 0x23C00 is declared but its executed call site was not established in this pass. These are partial caller links, not complete client frame/action dispatch chains.

The discovery ladder has partial caller/output/data anchors; it is not exhausted. No claim of “not found” is made. Original-image native disassembly corroborated the encoded instruction runs for 0x22C10, 0x22FB0 and 0x895D0–0x8968D. Loader descriptor instruction bytes were also checked against the original D2Common image. Callee and dispatch gates remain. T01 provides a narrow text-cell read; no integration vector was executed, and archive contents do not establish gameplay confirmation.

## Reading rationale and caveats

0x22490 initializes kind and class from the unit, calls import thunk 0xC65E, and invokes thunk 0xC658 only for kind zero. It returns the class local. The label suggesting a sound id is unsupported. The companion accessors return kind and mode using the same conversion machinery. Import resolution gives D2Common ordinal 10175 at 0x2F1B0 and ordinal 10350 at 0x501E0. The latter's entry tests a unit flag and begins a table-driven conversion; its remaining path is unread.

0x22A70 first resolves effective class, checks a source count, and examines monster data flags. A first alternate path consults an imported record and its +0x18 field. A later path tests another monster flag or local helper 0x4C0D0, then reads a signed word at +0x16 from another imported record. Positive values use the direct lookup, otherwise the indirect lookup. The meanings of those flags and record fields are not established by column names.

The direct imported lookup validates table presence, nonnegative index and count, then selects a 0x94-byte record. The indirect lookup validates two tables and the upper class bound, reads a signed word at source-record stride 0x1A8 plus 0x14, then validates the resulting sound-record index. This does not prove the tables' names, nor make original memory strides a behavioral requirement. Import thunks 0xCABA and 0xCF76 resolve respectively to ordinal 10876 / 0x52FF0 and ordinal 10917 / 0x52FA0.

0x22CA0 requires effective kind one and a non-null selected record. Selector 8, 9, 10, 11 chooses dwords at +0x44, +0x48, +0x4C, +0x50 respectively. Other selectors return. Cue 0xA84 requires unsigned current tick minus the unit's +0x7C value to be at least 25; other zero cues return. The request passes three zeros and the unit to 0x88A70 with the selected cue in a register, then stores the current tick to unit +0x7C and shared 0x11C4C4. Actual cue binding and argument semantics remain gated.

0x23C00 includes kind/mode eligibility, special class checks 15 and 110, signed animation field +0x4C, initial factor 2 and default chance 100. Selected record fields +0x58/+0x5C/+0x60/+0x64 replace secondary cue, factor, phase displacement and chance candidates. It computes phase comparisons through 0x895D0 and requests through 0x22FB0 then 0x88A70. After primary request success it sets a result-dependent value through 0x87ED0; a secondary request can consume a random draw via 0x89640. Those local callees and producer mappings are not fully read, so arithmetic and seed order are not verified.

The 0x22C10 original-image disassembly confirms a shared unsigned elapsed-tick threshold of 4, one inline random step modulo 3 and selector addition 6, record field +0x78, then timing writes. The loader binds +0x78 to `Flee`, so the label suggesting random footsteps is misleading. 0x22070 and 0x220A0 store/read unit +0x7C despite their graphics-related labels.

### Loader and additional traversal

The original D2Common image was independently rehashed and matches the manifest's SHA-256. D2Common 0x58120 builds descriptors naming `monsounds` and writes the resulting table to the global consumed by the two imported sound-record accessors. Raw descriptor-name strings and native instruction bytes bind `Skill1`–`Skill4` to record +0x44/+0x48/+0x4C/+0x50, `Footstep` to +0x54, `FootstepLayer` to +0x58, `FsCnt` to +0x5C, `FsOff` to +0x60, `FsPrb` to +0x64, and `Flee` to +0x78. These are code-loader bindings, not archive cell verification or names from which execution semantics were inferred. The generic loader callee remains unread.

The 0x22FB0 cue-selection path uses the player-class table's +0x28 or monster sound record +0x54 as its base. It adds a material-dependent adjustment through 0x226A0 and a further 24 when its flag input is nonzero. Material mapping and player table production remain unread. 0x4C0D0 tests monster-data flag bit 4, and has an adjacent bit-3 accessor; it does not independently prove that either flag means unique-monster status. Full 0x501E0 reading confirms state-driven identity conversion and a remaining callee at 0x4F780; it is not a defense-malus function as its label implies.

Incoming-reference extension locates D2Client 0x2B1E0 calling 0x4F4A0. It also locates calls to 0x23F20 in 0x29990, 0x5BCC0, 0x9A920 and 0xAD740. Their complete dispatch ownership is still unverified.

## Effective-identity follow-up

The full D2Common 0x4F780 leaf was read after checking the matched registry (no matched C), and its encoded instruction runs were corroborated by original-image native disassembly. It has no callees. Its mode-availability byte/mask lookup is the operation already specified by GDA-R09b, using the MonStats/MonStats2 records and GDF-R06a mode identities; this closes the former unread-callee and unnamed extended-mode-field gaps.

Original data at D2Common RVA 0x93088 maps player modes 0..19 respectively to monster modes `[0,1,2,15,3,1,2,4,5,6,7,4,11,8,9,10,11,12,14,13]`. The inverse original table at 0x930D8 maps monster modes 0..15 to player modes `[0,1,2,4,7,8,9,10,13,14,15,16,17,19,18,3]`. These are engine-fixed code data, not mode names inferred from columns. The leaf tests mapped monster mode availability; unavailable modes 5/7 fall back to 4, 6 to 3, 9/10/11 to 8, 15 to 2, and other modes to 1. Mode 1 is the terminal fallback even if unavailable. The table/jump data bytes were parsed as data rather than trusting disassembly that treats jump tables as instructions.

State descriptor identity reuses GDA-R13 and its evidence record: `gfxtype` is the unsigned byte at +0x2D, `gfxclass` the signed word at +0x2E. D2Common 0x501E0 acts on the first active state in a prepared list whose gfxtype equals 1 or 2: 1 selects monster kind and gfxclass, converting an original player mode through the leaf; 2 selects player kind and gfxclass, converting an original monster mode through the inverse table, with monster mode 4 and resulting player class 6 specially selecting player mode 12.

Targeted list-producer reading found D2Common 0x5F6E0 calls 0x5DA30 with flag index 16 for the list/count used at DataTables +0x17C/+0x180. GDA-R13 identifies flag 16 as `disguise`, not `transform`. The 0x5DA30 producer iterates state rows in ascending index and appends indices carrying that bit. Its allocation import and encoded producer instructions have not yet been fully verified against original bytes. Unit membership eligibility and its bit-3 cache producer remain untraced. Thus the unread leaf and descriptor-name gaps are closed, while list preparation/membership and full identity integration remain explicitly open.

The material helper 0x226A0 was read through its return. It derives a default material from local-player level data, tests the emitter tile in the current room and neighboring rooms through 0x21ED0, and uses 0x21E80 when a tile match is found. Both tile helpers were subsequently read after checking their matched registry (no matched C). 0x21ED0 subtracts supplied room origins, divides signed coordinate differences by 5 toward zero and returns the first tile whose local x/y match. Its two room imports resolve to D2Common #10561 / 0x3BA60 and #10544 / 0x3BB50; both matched C helpers were read. The former supplies a room-bound rectangle and the latter a tile collection/count. The material flag selector at 0x21E80 gives first-match priority to masks 0x20, 0x08, 0x10, 0x40, 0x80, 0x400, returning selectors 1, 2, 3, 4, 6, 5 respectively; absent metadata or no bit yields zero. The producer meaning of these flags and level/environment bindings remain unresolved; no material names are assigned from suggestive labels. Room/current-level/neighbor imports additionally resolve to D2Common #10331 / 0x2FE10, #10826 / 0x3C000, #10014 / 0x6CCC0, #10383 / 0x3BA20. All four bodies were subsequently read, matched C except the unit-room accessor at 0x2FE10. They expose room, level identity, bounded level record and room-neighbor collection/count. Their misleading labels do not name these behaviors. The remaining material integration slice is the Levels field at +0x21C, the client 0x58-byte environment record field +0x1C and DT1 material metadata producer identities; numeric traversal bodies are now read.

## Supplied action-stage verification

The level-field identity gap is narrowed by compatible original-artifact evidence in the [music loader recheck](sound-music-selection-and-transitions.md): the level record byte +0x21C is `Levels.SoundEnv`, with descriptor at D2Common 0x6F81C–0x6F839 and its original string at 0x8FAC8; loader 0x6E130 publishes the table used by getter 0x6CCC0. This reuses the recorded 2026-10-29 descriptor and independent image-listing coverage, not a new reading by this writer. Generic parser integration retains GDA-R02/GDA-R03d gates. The remaining material producer identities are the environment field +0x1C and tile material flags.

USN-R02b is deliberately separated from USN-R02a's unresolved producers. The computing stage at D2Client 0x22CA0 is reached from the action dispatch table in 0x24490; a second caller at 0x23170 supplies selector 8 on one mode branch. The complete 0x22CA0 instructions were corroborated against the fingerprinted original image, including the initial encoded push. Effective-kind and record-lookup operations are supplied at this stage; their execution is USN-R01/USN-R02a, not silently credited to the stage.

The stage consumes the loader-mapped Skill1..4 fields, checks zero or the fixed cue 2692 condition, and calls the cue-request boundary at 0x88A70. Original 0x220A0's non-null-unit timing read is fully read. After the request, instructions at 0x22D39 and 0x22D41 independently load current time and write unit +0x7C and shared 0x11C4C4 respectively; no test of the request's return intervenes. This supports the failure-return and distinct-clock-read assertions in USN-V01. Its special-cue comparison uses unsigned below 25 after wrapping subtraction, supporting USN-V02 and USN-V03. Negative cue identities pass the nonzero check and are forwarded; any rejection inside playback is outside this stage.

Self-review on 2026-10-29 recomputed V01 as cue 102/ticks 101 and 102 despite request return zero; V02 wraps the difference to 25; V03 suppresses elapsed 24, zero, invalid selector, wrong kind and missing record, while forwarding -1. No supplied-result vector establishes runtime identity production, audible playback or original asset association of cue 2692. Fresh review is required for this added stage and pending with the coordinator.

## Observation procedure

USN-O01 is pending. Use a fingerprinted original Expansion 1.13c single-player game and a test character, Normal difficulty and an ordinary nearby monster. Record effective kind/class/mode and selected record at entry, then action and movement requests with their arguments and results. First resolve the static selectors and table bindings; only then specify exact expected values for a chosen monster row. Observe idle, movement, attack and hit/death transitions separately. Capture actual tick differences and random inputs where claimed, preserving admission failures separately from selected cues. Do not infer a matching request solely from audible similarity, or infer absence from a zero-record trace without checking trigger and collection validity.

## Review and limitations

USN-R02b is verified statically only for supplied action-stage inputs. Other behavioral rules remain unresolved and no rule is runtime-confirmed. USN-R05 verifies only the named text cell in T01. No executable vector is fabricated for a gated selection boundary. A fresh blind-reader review remains pending with the coordinator. Private audit records preserve paths, source forms and exact commands separately; public records contain only original artifact references and bounded results.

## Table record T01

Date 2026-10-29; source profile: reported vanilla LoD Expansion 1.13c (`lod-113c`). The original Game.exe profile fingerprint matches the source manifest; loaded-module full fingerprints and complete override inventory were not captured. No data override was reported. This is a text-cell source-profile observation, not proof of unmodified compiled tables.

Effective virtual path: `data/global/excel/monsounds.txt`, read through the game's Storm loader. Initial range 0–1000 returned 1000 of 22927 bytes. A second request at offset 0 for 22927 bytes returned all 22927 bytes. The full base-64 body was decoded to 22927 bytes and split as tab-separated text by the complete 41-column header. Row selection uses named `Id=act1hire`, not a numeric row assumption. The raw `FsCnt` cell is `2`, interpreted as decimal integer 2, supporting USN-R05. Other separately inspected cells were `Footstep=medium_walk_dirt_1`, `FsOff=0`, `FsPrb=100`, and blank `Flee`; these do not establish execution semantics.

Full file coverage was obtained; effective-file hash, physical supplying MPQ and compiled correspondence were not captured. No gameplay vector was observed or promoted. The data read did not use live-state writes.

## Fresh blind-implementer review

Historical full pass 1, 2026-10-29, before the supplied action stage was added: MISMATCH 0, GUESS 0, GAP 0 newly uncovered, DISCRIM 0. The reader checked the then-sole accepted data assertion USN-R05 against T01 (act1hire.FsCnt raw text 2); that version had no verified behavioral vectors to recompute. Its gated scope covered the four behavioral candidates. This historical result does not review the subsequently added action stage.

Historical HARDEST 3 outcomes: no change for dispatch/effective identity, alternate/default binding, and complete movement/action callee paths. The reader confirmed that numeric foundations do not by themselves establish client seed ownership or draw order and that base unit identities do not establish effective identity conversion. No behavior was added to satisfy an unsupported interpretation. That pass met its stopping rule only for its then-current data-only subset; full selection parity and runtime confirmation remain open.

Full pass 2, 2026-10-29, supplied action stage: MISMATCH 0, GUESS 0, GAP 0 newly uncovered, DISCRIM 1. The sole new finding was accepted: only selector 9 had distinct cue fields, so incorrect mappings for selectors 8, 10 and 11 could pass the prior vectors. USN-V01 now has independently initialized cases (a)–(c) selecting 8, 10 and 11 from distinct 101/102/103/104 cue fields and requiring exact selected requests 101, 103 and 104 respectively. Self-review recomputed all four mappings and post-request timestamps. These are evidence-backed vector improvements, not changes to the rule or its static label. A repeat fresh-reader pass is pending; the stopping rule is not claimed for this revision.

Mechanical validation records are dated to their actual revision. The earlier data-only revision had 0 errors and 8 warnings; the supplied action revision before the distinct-selector additions had 0 errors and 12 warnings. Warnings were reviewed: temporal field names describe current retained state, and source-reading method names describe public verification methods rather than private reconstruction paths. No implementation or user approval is asserted.


## Fresh blind-implementer full pass 3 — 2026-10-29

MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1, GATED 5, COVERAGE 0 for the supplied action stage, HARDEST 3. All current vectors, including the new distinct-field selector cases 8/10/11, recomputed correctly.

DISCRIM outcome: the V02 phrase signed subtraction was ambiguous. A 32-bit two’s-complement subtraction produces the same low word 25 for this vector and is equivalent here; unwrapped mathematical subtraction produces a negative difference and suppresses incorrectly. The discriminator wording now names the latter explicitly. This clarifies only the incorrect alternative; rule, inputs, assertions and expected outputs are unchanged, so no further full run is required under the wording-only exception. The vector does not claim to distinguish equivalent low-word arithmetic on these inputs.

GATED outcomes: retain USN-Q01 through USN-Q05. HARDEST outcomes: no change for effective identity/record production, fixed-cue asset/playback admission or movement/secondary-cue RNG order; their existing gates already name these limitations. COVERAGE 0 is limited to this supplied stage, not full unit sound scope or runtime coverage. Reader stopping rule met for the current supplied stage; no user approval inferred.
