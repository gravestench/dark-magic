# Evidence: Death Sentry corpse selection

Spec: [Death Sentry corpse selection](../../specs/skill-assassin-death-sentry-corpse-selection.md).
Artifacts: [original source manifest](../source-manifest.md).
Limits: [gap file](../../plans/gaps/skill-assassin-death-sentry-corpse-selection.md).
Sibling: [Death Sentry corpse damage evidence](skill-assassin-death-sentry-corpse-damage.md) holds the explosion records.

## Identity, reuse and method

Lord of Destruction Expansion 1.13c, Windows x86. References use manifest artifacts `LOD113C-D2GAME` and `LOD113C-D2COMMON`. Descriptive function names are analyst labels, not established original exported names. Module-relative RVAs and export ordinals identify matching original files; a reader can hash their own files, navigate in a PE-aware disassembler and follow these calls/writes without private reconstructed source.

The retained 2026-10-12 record used original disassembly for the reached think, corpse search and two-unit distance; original import/field/RNG readings supplied the initial arithmetic evidence. The first pass had no recorded date; it is not assigned one now. Full originals and investigation details were preserved privately before revision. No prior gameplay observation existed.

On 2026-10-15 the manifest and current NRF/GDF/SFR/SCD contracts were assessed for compatible artifact, method, input and execution-boundary coverage. Reused records retain their dates and limits.

The available source index was checked for the targeted routines. New verification read byte-equivalent C for D2Game 0x9E2E0; no listed equivalent C existed for the new main-function readings. Original instruction listings were then read for D2Game 0xA7C20/0x9E810/0x1A40 and for the identified producer slices below. Reading an unresolved producer slice does not make it a complete contract. No gameplay, hooks or other-build analysis was performed.

## Dispatch and functions

The reached think 0xA7C20 calls budget 0xA4A10, target wrapper 0xB1BC0, corpse search 0x9E810, range progression 0x1A40 and the two-unit distance import. Its successful corpse branch stores AI parameter=0 and requests Skill1/mode 9 through 0xF2380; fallback requests Skill2/mode 14, while delay goes through 0xF22C0. This does not establish the numeric AI selector or dispatch-table root for DeathSentry. Text AI=DeathSentry is a separate data observation. Search was bounded to these caller/output/table slices; no six-rung absence claim is made. The action-55 explosion root reached by the mode-9 request is recorded in the [sibling companion](skill-assassin-death-sentry-corpse-damage.md#dispatch-and-functions).

| Artifact | Function/table RVA or export | Rules / semantic role | Method, date and actual boundary |
|---|---|---|---|
| LOD113C-D2GAME | 0xA7C20 | DST-R01..DST-R05/DST-R06a, reached think | full instructions 2026-10-12 and 2026-10-15; actual AI root DST-Q01 |
| LOD113C-D2GAME | 0xA4A10 /0xB1BC0 | budget/shared target | dated SFR contracts; complete target producers retain SFR gaps |
| LOD113C-D2GAME | 0x9E810 | DST-R03a, search request/first return/release | complete consumer instructions 2026-10-15; collection producer DST-Q03 |
| LOD113C-D2GAME | 0x9E2E0 | DST-R03b corpse callback | byte-equivalent C 2026-10-15: short-circuit corpse-valid then unit-find predicate |
| LOD113C-D2GAME | 0x1A40 | DST-R04 Param3/Param4 | full instructions 2026-10-15: signed lookup/level guards, multiply/add return |
| LOD113C-D2COMMON | #10522 /0x7FBD0 | DST-R04, two-unit distance | full retained 2026-10-12 and targeted 2026-10-15 instructions; both half sizes, per-axis clamps, octagonal result |
| LOD113C-D2COMMON | #10589 /0x30F80 | DST-R04, size by kind | SFR's 2026-10-12 record/new equivalent C 2026-10-15 reused; no duplicate reread required |
| LOD113C-D2GAME | 0x1180 | DST-R05 | NRF dated seed/modulo contract reused |
| LOD113C-D2COMMON | #11157 /0x4FFD0 | DSC-R01b corpse-valid, consumed by DST-R03b | recorded in the sibling companion; producer chain not complete |
| LOD113C-D2COMMON | #11106 /0x124E0 | DST-R03b unit-find predicate | targeted 2026-10-15; radius/type/mode/flag/room/collision dependencies located, not complete; DST-Q03 |

### Import identities and caller arguments

Resolved from matching original import/export directories on 2026-10-15, or reused from the named current dependency record. Locations are evidence navigation, not required implementation storage. The complete thunk table is in the [sibling companion](skill-assassin-death-sentry-corpse-damage.md#import-identities-and-caller-arguments); rows used by the think and search are repeated here.

| D2Game thunk RVA | Original export / target RVA | Caller semantic operation / limit |
|---|---|---|
| 0xAB86 | D2Common#10522 /0x7FBD0 | two-unit distance(target,corpse) |
| 0xAE2C | D2Common#11157 /0x4FFD0 | corpse-valid predicate |
| 0xAECE | D2Common#11106 /0x124E0 | unit-find predicate(candidate,descriptor) |
| 0xA3F4 | D2Common#10630 /0x4F070 ->0x4E480 | skill instance request, SFR/SCX selection limits |
| 0xA400 | D2Common#10306 /0x505C0 | bonus-inclusive level request; actual bonus producer gates |
| 0xA478 | D2Common#10331 /0x2FE10 | source room |
| 0xA502 | D2Common#10591 /0x12030 | collection setup/request; body not claimed verified here |
| 0xA4F0 | D2Common#10052 /0x12720 | unit-find request; actual result/filter/order DST-Q03 |
| 0xA4FC | D2Common#10344 /0x12000 | release request; not an established complete release contract |

### Semantic layout correspondence

| Semantic term | Original representation | Supported rules |
|---|---|---|
| kind/class/identity/seed/path | Unit+0/+4/+0xC/+0x20..24/+0x2C | DST-R03a..DST-R06a |
| lastCorpse | AI control parameter=0 at+0x14 | DST-R04/DST-R06a |
| difficulty | Game+0x6D u8 | DST-R02/DST-R05 |
| Skill1/Skill2; `aip2`/3/4 | MonStats+0x170/+0x172; +0x5C/+0x62/+0x68 plus 2*difficulty | DST-R01..DST-R05 |
| max-distance parameters | Skills+0x150/+0x154 | DST-R04 |

## Corrections and affected claims

- DST-R03a/DST-R03b/DST-Q03: corpse search is not merely any dead monster. Equivalent callback 0x9E2E0 short-circuits corpse-valid then general unit-find predicate. Corpse-valid 0x4FFD0 (DSC-R01b) checks monster/mode 12/excluded active-state overlap/corpseSel; complete producer contracts remain gated. Local request/first-result and supplied-result guards remain code-verified; their actual integrated eligibility is not promoted.
- DST-Q01: named text AI/skills cells close text-name absence only. Actual compiled selector/AI-table binding is still unresolved, with no unsupported all-six-rungs absence label. DST-Q03's field-name portion consumes GDF but real data/producer integration remains open.
- Question on other builds: every prior 1.14d entry was untested. No other-build source or observation was authorized or collected; this remains a profile limit rather than an in-scope behavioral question.

Prior vector correction rationale remains: DST-V03's earlier distance=10 acceptance was withdrawn because the strict half=10 gate rejects equality. Historical aliases were removed because no non-gap plan/implementation/test consumer was found.

## Named table cells

**TD-DST**, 2026-10-15, is the think-relevant subset of the same coordinated read-only effective text reads recorded as TD-DSC in the [sibling companion](skill-assassin-death-sentry-corpse-damage.md#named-table-cells); the read was performed once and is not repeated. Those reads went through the running game's Storm loader. The member reported vanilla launch key lod-113c. Historical vanilla profile/earlier disabled override inventory was reused; refreshing the version registry failed. Complete current override inventory, loaded-module hashes, physical supplying MPQs and compiled table correspondence were not captured. These limits prevent upgrading text cells into compiled gameplay bindings. No gameplay/hook action was performed for this record.

Full Skills file coverage: virtual `data/global/excel/skills.txt`, offset/returned ranges 0/60000,60000/60000,120000/53535; total 173535 bytes. Joined effective-file SHA-256 `465c28e83657540e0d7ae5b0de826d5657be7fba1aa539a85c3ca3aeea34b498`. Full MonStats coverage: virtual `data/global/excel/monstats.txt`, seven 60000-byte ranges at offsets 0..360000, final 420000/12634; total 432634 bytes; SHA-256 `a15e1cb876bd9af1e8f54199c9619226cf4835c45c799aed2d432901487ecb7c`. Every chunk's successful offset/returned length was checked, joined without gaps and parsed by complete TSV header. Complete-file coverage establishes availability, not absence of arbitrary unqueried content. Rows below are selected by named identity; positional row excludes header and counts separator rows. MonStats data row 417 has explicit hcIdx=416, so raw row index is not silently used as class identity.

Each entry is one named text value, labeled verified in 1.13c data; blank is the raw empty cell, not an invented compiled 0.

| File | Row identity | Column | Raw value | Supported claim |
|---|---|---|---|---|
| Skills | `mon death sentry` (zero-based data row 312) | `Id` | `312` | named skill identity |
| Skills | `mon death sentry` (zero-based data row 312) | `Param3` | `10` | search progression text |
| Skills | `mon death sentry` (zero-based data row 312) | `Param4` | `1` | progression increment text |
| MonStats | `deathsentry` (data row 417) | `hcIdx` | `416` | explicit class identity |
| MonStats | `deathsentry` (data row 417) | `AI` | `DeathSentry` | text AI name, not actual table anchor |
| MonStats | `deathsentry` (data row 417) | `Skill1` | `mon death sentry` | named corpse skill |
| MonStats | `deathsentry` (data row 417) | `Skill2` | `death sentry ltng` | named fallback skill |
| MonStats | `deathsentry` (data row 417) | `aip2` | `0` | Normal delay text |
| MonStats | `deathsentry` (data row 417) | `aip2(N)` | `0` | Nightmare delay text |
| MonStats | `deathsentry` (data row 417) | `aip2(H)` | `0` | Hell delay text |
| MonStats | `deathsentry` (data row 417) | `aip3` | `50` | Normal chance text |
| MonStats | `deathsentry` (data row 417) | `aip3(N)` | `50` | Nightmare chance text |
| MonStats | `deathsentry` (data row 417) | `aip3(H)` | `50` | Hell chance text |
| MonStats | `deathsentry` (data row 417) | `aip4` | `16` | Normal fallback range text |
| MonStats | `deathsentry` (data row 417) | `aip4(N)` | `16` | Nightmare fallback range text |
| MonStats | `deathsentry` (data row 417) | `aip4(H)` | `16` | Hell fallback range text |

Table cells verify data only. Their loader meanings consume GDF; compiled calculation/link/AI identities still require correspondence. No data cell sets runtime_observed=yes.

## Observation procedures

All pending; recipes do not constitute captures. Use an owned vanilla LoD 1.13c profile with matching manifest binary identity and separately recorded effective-data identity. Prepare an Assassin with Death Sentry at one recorded rank/character level, record equipment and relevant skill bonuses, and a single ordinary monster class/level. Work in Normal outside town at a recorded corpse/sentry/target position; capture actual input values before interpreting a result. No specific unverified table value is assumed by these recipes.

- **DST-O01:** Place two owned sentries near one eligible corpse and one target. Record their think 0xA7C20 selections/remembered identities, Skill1/mode 9 dispatch, corpse-valid checks and state 104 request at 0x4F4D0 within the same frame window. Resolve actual event order rather than assuming simultaneous execution. If tooling lacks the required two-unit scheduling setup, report that limitation instead of choosing a race outcome.

## Regeneration and review record

2026-10-15 (recorded on the combined source spec before the restructure below): original scope retained; private snapshots preserve original main/gap files and all investigation history. Repository search for filename and prefix found sibling/companion/gap references, with no implementation/test/non-gap-plan consumer. Current numbering remains practical but aliases are not carried. Source/data corrections and their affected assertions are above; unchanged arithmetic/seeds remain available. Exact vector inputs/outputs now replace symbolic predicate-only cases. The implementation comparison is retained as historical findings in the gap file, without a new implementation inspection. Mechanical lint and self-review are recorded after final edit; fresh independent review is coordinated separately. Document remains draft because actual integration and gameplay observations remain open.

## Independent review record — 2026-10-15

Recorded on the combined source spec; entries for this spec's rules are kept here, the remainder in the sibling companion. A separate reviewer recomputed every executable vector from the rules and cited contracts and re-read the supporting original code, including D2Game 0xA7C20, 0x9E810, 0x1A40, 0x1180, byte-equivalent C for 0x9E2E0, and D2Common 0x7FBD0, 0x4FFD0 and equivalent C for 0x30F80. All vector outputs recomputed; no behavioral discrepancy was found.

- DST-R02/DST-R04/DST-R05: the gap file already listed the AI-entry question (now DST-Q01) as affecting these think rules, but their Id-map Gates column named only the producer question (now DST-Q03). The Id map now names the DST-Q01 entry gate consistently with DST-R01 and the gap file. No assertion, vector or label changed.
- DST-Q01 lead: a byte search of the original D2Game image finds the think pointer 0x6FCC7C20 exactly once, at RVA 0x10EE10. Group-review correction (2026-10-15): this cell is the think field (+8) of AI table entry 104, which starts at 0x10EE08 in the table at base 0x10E788 with 16-byte {flag, init, think, aux} entries. The original-image cells are {0, 0x6FCC39B0, 0x6FCC7C20, 0}. The SFR think occupies the same field of entry 101. The table indexer is D2Game 0x9D0F0: it uses MonStats+0x1E (signed, 0..147) unless a special index selects table 0x10E668. Installer 0x9D230 calls init 0xA39B0 and stores the think. Executor 0x9D3B0 calls the stored think. Init 0xA39B0 writes AI-control+0x18 = -1, the SFR-R02 budget counter. See the [SFR companion](skill-assassin-sentry-firing-and-range.md#review-additions-2026-10-15). Still open: that DeathSentry's compiled MonStats AI value is 104, the special-table selector, and the event root of 0x9D3B0.
- DST-R03a: the search parameters (room, target point, radius 10, callback 0x9E2E0, descriptor with filter 0x1002) are passed to the #10591 call; #10052 and #10344 then take collection addresses. The spec's single semantic "find" request summarizes these supplied D2Common operations, whose bodies remain unread under DST-Q03.

These are wording/gate-column and evidence-link edits; vector inputs, outputs and execution boundaries are unchanged.

## Blind-implementer full pass 1 — 2026-10-15

Recorded on the combined source spec (14 vector families). Counts for that whole contract: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 22, COVERAGE 16, HARDEST 3. No behavioral correction was identified. Dispositions for this spec's rules follow; the others are in the sibling companion.

| Finding | Disposition |
|---|---|
| COVERAGE 1: DST-R03a valid empty collection | Added DST-V03(f): valid find returns count 0, no corpse, and still requests collection release. Invalid-input DST-V03(e) instead makes neither request. |
| COVERAGE 2: DST-R04 remembered equality, wrapped/invalid progression | Equality already has DST-V03(a), which retains lastCorpse 77 and falls back. Added DST-V03(g): low32 progression wraps to -2147483648, half -1073741824, so ordinary distance 9 falls back rather than spending a shot. No additional invalid-progression fixture: DST-R04 specifies guarded maxDist 0 and strict comparison, giving failure for ordinary nonnegative distance; DST-R01 already stops at nonpositive selected level. No invalid skill/level producer is invented to reach an otherwise blocked think. |
| COVERAGE 3: DST-R05 roll equals `aip3` | Added DST-V04(d), sentry roll 51 and `aip3` 51: delay 3, no decrement or mode, seed advanced once. This distinguishes strict chance from accepting equality. |
| COVERAGE 16: DST-R01 negative budget initialization | No change: SFR-R02 owns lazy initialization. SFR-V05/SFR-V16 start counter -1 with supplied `calc4` 2/3; SFR-V24 distinguishes zero/negative initialization and owner failures. DST-V01 consumes lost and verifies stop/continue handling without replacing that owner contract. |

## Blind-implementer full pass 2 — 2026-10-15

Recorded on the combined source spec (14 families, 63 cases). Counts for that whole contract: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 3, GATED 29, COVERAGE 13 grouped entries, HARDEST 3. No rule correction was found; no DISCRIM finding concerned this spec's rules.

| Finding | Disposition |
|---|---|
| COVERAGE group 1: DST-R04 remembered equality, invalid skill/nonpositive L and distance/progression cases | No change: DST-V03(a) already tests remembered identity equality, DST-V03(b/c) distinguish strict distance with rank progression, DST-V03(g) tests wrapping progression. DST-R01/DST-V01 stops nonpositive selected level; DST-R04's isolated invalid lookup recipe gives maxDist 0 and rejects ordinary nonnegative distance. Actual invalid skill production is not invented to bypass DST-R01. |
| COVERAGE group 13 (think part): DST-R01/DST-R02/DST-R03a/DST-R05 principal branches | No change: DST-V01 covers initial stops; DST-V02 absent target/delay; DST-V03 successful/empty/invalid search; DST-V04 chance/range/equality/decrement failure. No specific new unresolved output or accepted wrong alternative was identified for this grouped advisory. |

## Blind-implementer full pass 3 — 2026-10-15

Recorded on the combined source spec. Counts for that whole contract: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 29, COVERAGE 15, HARDEST 3. Every applicable blocking finding is resolved.

| Advisory entry | Disposition |
|---|---|
| DST-R01 negative level | No change: DST-R01 specifies signed nonpositive rejection; DST-V01(a) tests the same rejection branch at zero with no subsequent work. A negative value introduces no other local effect. |
| DST-R02 difficulty 0/2 | No change: supplied indexed signed-word selection is explicit. DST-V02 selects the distinct Nightmare value 7 rather than Normal 2/Hell 9 and asserts ordered delay/no later work; no producer is inferred. |
| DST-R04 negative coordinate/Y separation | No change: the shared distance contract owns axis arithmetic. DST-V03 discriminates strict distance, sizes and progression; the cited owner remains authoritative for coordinate production/arithmetic. |
| DST-R05 negative AI words | No change: signed range/chance comparisons are explicit; DST-V04 tests pass/fail/equality and exact reached draw ownership. Negative nonzero chance follows the same rejection branch, with no new clamp or unsigned interpretation permitted. |

## Blind-implementer delta check 4 — 2026-10-15

Recorded on the combined source spec. The reader recomputed the changed vectors, including DST-V02, DST-V03 and DST-V04, from the rules and inputs alone, after the Id-map Gates column of DST-R02/DST-R04/DST-R05 gained the DST-Q01 entry gate. Counts: MISMATCH 0, GUESS 0, GAP 1, DISCRIM 0, GATED 3.

| Finding | Disposition |
|---|---|
| GAP 1: Id-map row DST-R03a lists only DST-Q03, not the DST-Q01 entry gate, although its search is requested from the same think after DST-R01/DST-R02 | No change: DST-R03a is reached only after DST-R01, whose row and the gap file carry the DST-Q01 entry gate, so the entry limit applies transitively; DST-R03a's own row names its separate collection-producer gate. DST-V03(e/f) are isolated supplied-input search stages, and no assertion depends on the entry. The same transitive reasoning applies to DST-R06a. Listing DST-Q01 on DST-R03a/DST-R06a (and in the gap file's AI-entry row) for presentation consistency is left to the maintainer: it is a Gates-column edit that this standard treats as needing another reader run, and the four-run cap for this review is reached. |

Outcome: MISMATCH 0; no unresolved GUESS or DISCRIM; the one GAP has a recorded no-change outcome. Status remains draft.

## Restructure (2026-10-15)

The approved review split the combined Death Sentry spec at its two independent triggers. This spec now holds the reached think and corpse selection: budget/skill, shared target, corpse search and integrated-search gate, corpse branch, fallback, and the remembered-corpse/competing-claim rules, with their four executable vectors and the two-sentry observation recipe. The explosion action (guards, corpse-valid producer, life/endpoint/roll/scale/split, geometry, victim hand-off, scaling/mitigation, duration/router and draw counts) stayed in [Death Sentry corpse damage](../../specs/skill-assassin-death-sentry-corpse-damage.md). The former combined local-producer question was divided: think-side producers are DST-Q03, explosion-side producers DSC-Q04. The state-order vector that also cites DST-R06a stays with the explosion, which owns the asserted request order. No rule behavior, label, value or vector result changed; ids were renumbered under the unconsumed-spec policy. Records above were partitioned by owning rule; their dates and counts are unchanged.
