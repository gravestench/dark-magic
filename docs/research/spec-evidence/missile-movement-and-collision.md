# Missile movement and collision: evidence

Contract: [Missile Movement and Collision](../../specs/missile-movement-and-collision.md).
Open assertions: [gap file](../../plans/gaps/missile-movement-and-collision.md).

## Identity and reuse

Evidence build is original Lord of Destruction Expansion 1.13c Windows x86. Artifact ids
`LOD113C-D2GAME` and `LOD113C-D2COMMON` refer to the exact analysis files in the
[source manifest](../source-manifest.md). Their file SHA-256/size correspondence was checked
before original-image decoding on 2026-10-15. This identifies analysis specimens, not DLLs
loaded by historical gameplay/table sessions. Version-resource, data-profile and override limits
in the manifest remain applicable.

The 2026-10-12 record traced the dispatcher, plain step, search/contact gates, pierce and
cooldown caller. Its actual date and static-only scope are retained. Existing NRF, GDF, PVP,
MCI, IEP, STL, SBA and movement-rate records were checked first for compatible identity and
coverage. They support the named dependency slices, not unread path/attachment/damage interiors.
New reading targeted unresolved byte runs, table dispatch and caller return handling; it does
not renew the earlier date or imply gameplay observation. Descriptions below are analyst labels.

## Original functions and output anchors

RVA is module-relative; ordinals identify exports only for the matching module. All references
in the first column use manifest artifacts. New reconstruction reads first selected a byte-checked
reconstruction where available, otherwise instruction listings; instruction-only callers with
embedded bytes were corroborated against the original image as described below.

| Artifact | Function RVA / ordinal | Analyst description | Rules | Method/date and boundary |
|---|---|---|---|---|
| LOD113C-D2GAME | `0x3E3F0` | Missile dispatcher | MMC-R01 | 2026-10-12 listing; 2026-10-15 full 288-byte original decode, town branches/indirect call/removal-result consumer |
| LOD113C-D2GAME | `0x3F540`, `0x3F320` | Plain wrapper/step | MMC-R02/MMC-R03/MMC-R06/MMC-R07/MMC-R09a | Full instruction listings 2026-10-15; entry/return paths and output writes |
| LOD113C-D2GAME | `0xE1930`, `0xE07A0`, `0xE1610` | Move wrapper,room-link and transition requests | MMC-R05a | Wrapper listing 2026-10-15; nested effects remain supplied/integration gap |
| LOD113C-D2GAME | `0x3EDE0` | Contact/end caller | MMC-R03/MMC-R11..MMC-R16 | Full listing and 832-byte original decode 2026-10-15; type/pierce/roll/event/handler/finish branches traced; delegated interiors excluded |
| LOD113C-D2GAME | `0x3AA20`, `0x3AD70`, `0x3ADD0`, `0x3AD30`, `0x39510` | Shared/type-specific search filters | MMC-R08/MMC-R10 | Full listings 2026-10-15; accept/reject returns, shared-filter bypass and no last-collision write |
| LOD113C-D2GAME | `0x38FB0` | Contact type gate | MMC-R11 | Listing plus complete 64-byte original image/jump-table check 2026-10-15 |
| LOD113C-D2GAME | `0x38F40`, `0x6F660` | Pierce use/creation | MMC-R12 | Full listings 2026-10-15; current owner stats, own-base input, local seed and layer-0 request; source creation call MIR-R15a |
| LOD113C-D2GAME | `0x3E590`, `0x38E60` | Cooldown caller/expiry callback identity | MMC-R16 | Full 224-byte caller listing 2026-10-15; callback body/attachment integration not established |
| LOD113C-D2GAME | `0x8E9A0`, `0x8E680`, `0x6F8A0` | Callback-free walker/creation event | MMC-Q03 | Walker directed read 2026-10-15 plus MIR-R01b; registration association does not establish recurring order |
| LOD113C-D2GAME | `0xE20A0` | Removal request target | MMC-R01/MMC-Q04 | Direct caller target retained from 2026-10-12; complete destruction bookkeeping unread |
| LOD113C-D2COMMON | `0xCEB0` | Velocity ramp/step arithmetic | MMC-R04a | Compatible current NRF-R11a contract; 2026-10-12 read and later NRF evidence retain their dates |
| LOD113C-D2COMMON | `0xE230` / 10342; `0xD8E0` | Shared path step/advance | MMC-R05a/MMC-Q04 | Directed full listings 2026-10-15 expose reset/update/position/room callees; complete integrated path output not established |
| LOD113C-D2COMMON | `0xD580`, `0x3E110`, `0xD420` | Bounded tracer,increment normalization,move test | MMC-R05b | Full listings; original image 416/128-byte byte-run checks 2026-10-15; supplied collision-result consumer only |
| LOD113C-D2COMMON | `0x4D760`, `0x4D6E0`, `0x4CE00` / 10573 | Relocation/query callers | MMC-Q04 | Available byte-checked reconstruction first for `0x4D760`; other listings; `0x4CE00` full 144-byte image/table check. Nested shapes/room map not completed |
| LOD113C-D2COMMON | `0x7F6A0` / 10389; `0x7F490` | Candidate search/room predicate | MMC-R09b | Full listings plus original 624/112-byte images and jump tables 2026-10-15; supplied rooms/sizes/positions and filter callback |
| LOD113C-D2COMMON | `0x3BA20` / 10383; `0x3BA60` | Stored room collection/bounds copy | MMC-R09b | Byte-checked reconstructions read 2026-10-15; stored array/count and bounds, no sort/discovery inferred |
| LOD113C-D2COMMON | `0x30F80` / 10589 | Unit size selection | MMC-R09b/MMC-Q04 | Byte-checked reconstruction read 2026-10-15; player/monster/missile cases, underlying class-record production retains gate |
| LOD113C-D2COMMON | `0x347C0` / 10214 | Recorded-cell accessor | MMC-R05c/MMC-R09a | Byte-checked reconstruction 2026-10-15; recording-condition assertion then stored list/count |
| LOD113C-D2COMMON | `0x36030` / 11003; `0x34F80` | Fresh type-3 allocation/path selector writes | MMC-R05c | Directed allocation listing and full 192-byte selector original-image check 2026-10-15; entry 4 (table base RVA `0x94180`, entry at `0x94190`) is `0x60000`,setting both conditions; later planner effects remain gated |
| LOD113C-D2COMMON | `0x6A270` / 10737; `0x6A6D0` / 10970; `0x6A030` / 10130 | Remaining read/clamped write/activation read | MMC-R06/MMC-R07 | Remaining byte-checked reconstruction and setter listing 2026-10-15; activation compatible 2026-10-12 record |
| LOD113C-D2COMMON | `0x6A130` / 10227; `0x6A1B0` / 10024 | Last-collision compare/write | MMC-R10/MMC-R11 | Full listings 2026-10-15; valid missile/row/flag/metadata guards; direct identity writes |
| LOD113C-D2COMMON | `0x35BA0` / 11009 | Stationary collision cache | MMC-R07 | Full listing 2026-10-15; speed-zero query with mask 32767 and word store; actual query integration open |
| LOD113C-D2COMMON | `0x35640` / 11056; `0x348B0` / 11164; `0x4BE10` / 10232 | Footprint/mask stores and removal request | MMC-R07/MMC-R08/MMC-R15 | Creation caller/current MIR-R05; available reconstruction for first; final caller listing. Nested map shapes remain gated |
| LOD113C-D2COMMON | `0x2FE10` / 10331; `0x3C390` / 10057; `0x2D320` | Room/town classification | MMC-R01 | Compatible current MCI/PVP records, actual lookup/null/town boundaries retained |
| LOD113C-D2COMMON | `0x391F0` / 10830; `0x38B70` / 10973; `0x39430` / 10910; `0x38C20` / 10587; `0x3A740` / 10887 | Tier/stat reads and replacing request | MMC-R08/MMC-R12 | Current PVP-R05a/RLI-R02a/c/MIR-R15a; reader is not writer proof; RLI-Q03 remains |
| LOD113C-D2COMMON | `0x7C8D0` | Missiles loader descriptors | Inputs | Compatible GDA-R10/MCI descriptor records; text cells do not establish all compiled records |

### Dispatch tables and original byte checks

On 2026-10-15 a PE-aware read of the matching Game image checked the 53-entry do table at
RVA `0x10D970` and count at `0x10DA44`: slot 1=`0x3F540`, slot 4 and slots 38..52 absent. The
53 bound is separate from pointer presence. The nine pairs at `0x10D928` map selectors 0..8
to filters `0,0x3AD70,0x3ADD0,0x3AD30,0,0x3ADD0,0,0x39510,0x3AD30` and masks
`0,132,260,388,0,260,4,64,389`. Addresses here are resolved module-relative handler RVAs.
The contact type table at `0x38FDC` targets `0x38FCD,0x38FC9,0x38FD0,0x38FD5`: reject,
player test,monster test,accept. Its unsigned bound also accepts supplied negative selector.

Hit/damage counts and local indirect-call/result consumers are checked in `0x3EDE0`; their
full tables and interiors consume the MHD companion. A null table entry is not an absent-handler
fallback. Original-image decoding distinguishes executable instructions from embedded table
words; raw table words were inspected separately, not interpreted as instructions.

The compatible MHD table record dated 2026-10-15 checks hit table `0x10DA48`,count 71 at
`0x10DB64`,null slots 0,30,34,41,42,46,49,60..70; damage table `0x10DB68`,count 31 at
`0x10DBE4`,null slots 0,15..30. Positive signed-index guards skip slot 0,while eligible null
slots fault. This record is reused for table identity/scope,not per-handler execution.

### Trace and candidate reasoning

The path-step `0xE230` reaches velocity `0xCEB0`, advance `0xD8E0`, then tracer `0xD580`.
The latter calls normalizer `0x3E110` and CheckMove `0xD420`; CheckMove selects collision
relocation/query and accumulates its returned word. New source inspection answers local trace
count,cap,normalization and blocked-endpoint questions. It also exposes remaining callers for
position commit,room transitions,footprint shapes and path reset; supplying a query result does
not establish any of those effects. A query selector `0x3401` is translated to `0x3C01` before
the nested operation; neither a universal collision-map meaning nor full path completion is
inferred from the local trace vector.

At `0x7F6A0`, the adjacent collection is consumed in stored order; unit successors are read
after rejected candidates. The type and nine geometry tables were checked against the original
image. `0x7F490` first copies eight room-bound words through `0x3BA60`. Its actual signed
branches test upper X only after x+margin<left, and upper Y only after y+margin<top. With usual
positive extents those branches accept cells outside a conventional expanded rectangle. The
contract deliberately states the literal predicate, not an inferred spatial interpretation.
Candidate size/type/position consumers are verified with supplied inputs; actual class-data,
room-list and map creation remain integrations.

## Corrections and retained uncertainty

1. The 2026-10-12 invalid-step prose promised a graceful end. Source `0x3F320` does request
   no-target contact, but `0x3EDE0` dereferences its absent row. MMC-R03/V01 now distinguish
   the dispatcher guard from direct-entry fault; no observed crash is claimed.
2. The earlier broad “force-end can survive bit 4” reading omitted the caller at `0x3F51A`:
   after no-candidate barrier contact, the step writes result 2 regardless of contact return.
   Movement failure/expiry instead forward the contact return. MMC-R09a/V06 distinguish final
   dispatcher removal, not merely the hit callback's result.
3. The earlier room description said expanded bounds contain the cell. Full original decode
   `0x7F490` supports the unusual conditional upper-bound comparisons above. MMC-R09b/V07
   replace that unsupported geometric simplification.
4. The earlier stat-setter label implied established stored counts. Current NRF/STL contracts
   expose replacing/propagation gaps: MMC-R12a establishes decrement arithmetic and the layer-0
   request; MMC-R12b/Q06 keep storage gated. Creation count is MIR-R15a's; the zero-seed local
   remainder sequence in MMC-V09(d) remains exactly 66,18,35,30.
5. Earlier cooldown prose omitted allocation failure and implied established expiry. The
   `0x3E590` null-result branch precedes schedule/list-state/callback/attach/state operations;
   success requests those in order. MMC-R16/V11 cover requests; actual attach/expiry remains open.
6. The tracer increments count even without recording-condition; that condition controls only
   cell writes. The ten-cell cap therefore still applies. Last free-centre/full-endpoint results
   are local tracer outputs, not proof of the outer path-step's return/publication.
7. Direction scale now consumes the current verified NRF-R11a/b/MCI supplied-input contracts;
   actual planner production remains gated. Strict greater-than clamp and later zero-speed skip
   retain source behavior. Source own-base 328 is distinct from selected aggregate 328.
8. Numeric mask consumers are established; complete terrain/player/monster map-writer meanings,
   recurring queue order and owner seed producer were not closed by a label or data name.
   Other builds remain outside the evidence profile.
9. The initial recording/restricted flag writer is now located: allocation `0x36030` selects
   type 4 through `0x34F80`,whose original-image table entry 4 at `0x94190` is `0x60000`.
   Its clear/OR writes set bits `0x20000`/`0x40000` on that local branch. This closes the missing
   initial writer in MMC-R05c,without establishing later planner retention or full allocation/
   placement side effects. Current MCI allocation/planning gates remain applicable.

## Named-cell records

Compatible original-profile effective loader reads were reused, with structure/coverage and
headers rechecked on 2026-10-15. ItemStatCost's original read is dated 2026-10-13; States' loader
record is dated 2026-10-15. Each read reports vanilla LoD 1.13c profile, no known overrides;
complete override inventory, physical supplying archive and historical loaded DLL hashes were
not captured. The source-manifest limits apply. These are text-cell claims only, not compiled
record equivalence or gameplay observations.

| Virtual file | Coverage | Effective SHA-256 |
|---|---|---|
| `data/global/excel/itemstatcost.txt` | offset 0,returned 40109,total 40109; complete header/target rows | `ee7fa82b48e0c4f27a566334b2da13cf1c83f5c4c0439552084a231cb5938375` |
| `data/global/excel/states.txt` | offset 0,returned 20920,total 20920; complete header/target rows | `898806c2b69e55e316ecfed48dde95aa2933829b94fdad234bdabc27bdbc23e1` |

Rows below use explicit ID/id and zero-based data-row index excluding header; indices agree.
Each numeric identity is also read from its ID/id cell as the displayed decimal token.

| File | Identity/index | Column | Raw cell | Claim |
|---|---|---|---|---|
| ItemStatCost | ID19/index 19 | Stat | `tohit` | Missile hit bonus name |
| ItemStatCost | ID156/index 156 | Stat | `item_pierce` | Pierce chance component |
| ItemStatCost | ID166/index 166 | Stat | `skill_pierce` | Pierce chance component |
| ItemStatCost | ID172/index 172 | Stat | `alignment` | Matching state-list tier name |
| ItemStatCost | ID328/index 328 | Stat | `pierce_idx` | Source own-base and missile count name |
| States | id 86/index 86 | state | `justhit` | Cooldown state name |
| States | id 105/index 105 | state | `alignment` | Tier-list state name |

Each named cell is **verified in 1.13c data**. Code-fixed numeric selectors are independently
shown by callers; presence of these names in text does not define arithmetic or execution.

## Gap investigations

| Question | Attempt/result on 2026-10-15 |
|---|---|
| MMC-Q01 | Reused MCI allocation/planner records; read `0x36030` type-3 branch,`0x34F80` and original table entry 4=`0x60000`,closing initial two writes. Directed `0x36230`/`0x35780` leaves later planner/retention integration open |
| MMC-Q02 | Reused MCI event association; read `0x8E9A0` walker and no-callback dispatch. Recurring production/requeue/order not established; partial root search does not prove absence |
| MMC-Q03 | Read path-step/advance/normalizer/trace/CheckMove; checked query/size/room accessors and original bytes/tables. Nested query/map/position/reset/destruction paths remain incomplete,so actual effects gated |
| MMC-Q04 | Compared numeric masks and creation footprint request with current GDF/MCI; no complete terrain/player/monster writer chain. Keep numerical descriptions |
| MMC-Q05 | Assessed prior evidence:no matching gameplay records. Migration excludes gameplay; recipes below remain pending |
| MMC-Q06 | Checked NRF stat reader and current STL replacing-write record; `0x3A740` is a request boundary,not complete stored propagation |
| MMC-Q07 | Checked current SBA/IEP/MHD ownership and direct contact calls/results; per-handler interiors not read by this slice |
| MMC-Q08 | Full cooldown caller read; allocation failure/success order settled. STL attachment/removal/expiry remains unresolved |

The 2026-10-15 owner own-base 328 assessment (MCI own-base writer assessment and stat-reader
contract reused; missile decrement is not an owner seed producer; no universal threshold
inferred) now belongs to MIR-Q07; see the restructure record below.

## Observation procedures

### MMC-O01

Use a Normal Amazon able to fire ordinary arrows with Pierce,recording class/character/skill
levels,gear and actual chance contributions. Stand in a safe Blood Moor area; fire separated
shots for chance 60,70 and zero control using natural available setups. At creation record source
own-base 328 and selected 156/166,then local seed/draws/count-write request in Game `0x6F660`;
read actual missile 328 afterward. Compare full seed pairs and count against MIR-V13 and
MMC-V09(d). At later
contacts record current owner stats,missile count,decrement requests and final dispatcher result.
Repeated identical counts do not prove the owner seed producer; a readback discrepancy calls
for write/attachment investigation. No memory writes are part of this procedure.

### MMC-O02

Use a Normal Amazon with ordinary bow/arrow,recording character/skill levels,gear and selected
Missiles row. Stand still in Blood Moor,fire one arrow into open ground. Record creation path
conditions,position/direction and row speed fields,then one moving update's actual entered cells,
collision bits and path-step return at Common `0xE230`/`0xD580`. Compare recording and restricted
blocking conditions with their actual writer before claiming producer closure. Capture before
and after the same update; a path snapshot alone does not establish how it was initialized.

### MMC-O03

Use the same Normal arrow setup,recording exact row/speed/owner and target identity. Place a
living hostile monster immediately before a wall along a shot direction such that both cells
are crossed in one measured update. Record trace queries,cells,move return,contact target/force,
hit-handler return,plain-step result and dispatcher removal at Game `0x3F320`/`0x3EDE0`/`0x3E3F0`.
Control with an equivalent open-ground shot and an unobstructed monster. Distinguish movement
failure,early silent gate and later no-candidate barrier; only the latter discards contact return.
If an aborting handler cannot be naturally selected,report that branch unobserved rather than
manufacturing an expected game row. Compare recorded timing,not visual disappearance alone.

## Migration and review record

2026-10-15: retained the existing boundary,17 parent rule families,11 vector families and three
observation recipes. The contract is unconsumed by implementation/plans/tests in the repository;
ids remain in practical execution order without aliases. Original bytes and detailed private
source forms/search history are preserved separately with SHA-256. All six prior question
families were assessed; the other-build question now states the evidence boundary while pending
1.13c observations and explicit dependency integrations remain visible. No implementation
comparison existed or was requested. No gameplay action or implementation test ran.

The document checker `.agents/skills/spec-writing/tools/check_spec.py`,invoked with argument
`missile-movement-and-collision` using the existing Python 3.13.15 interpreter,returned 0 errors
and 0 warnings on 2026-10-15. Writer self-review
checked inputs/outputs,gates,source ownership and preserved families; hand recomputation checked
all four seed pairs,trace endpoints/cap,ramp/clamp,lifetime and geometry cases. This is document
validation,not an implementation test. Fresh blind-reader review is pending; historical review
is not treated as review of this contract.

Root source review requested clearer mask-versus-bit-index terminology and an explicit
pathFootprintSize input distinct from Missiles.Size. Those clarifications and the delegation
of the already consumed creation-pierce calculation to MIR-R15a/MIR-R15b leave all exact vector
inputs,draws,outputs and gates unchanged.

### Blind-reader FULL1 and triage

The independent full reader examined 11 vector families and reported 64 independent scenarios
(family counts 7,3,4,4,8,2,10,7,7,7,5; retained trajectories counted once). Results:
MISMATCH 0,GUESS 0,GAP 2,DISCRIM 0,GATED 10,COVERAGE 11,HARDEST 3. This is a document reading,
not execution against a game or implementation. The resulting edits require a fresh full pass.

| Finding | Outcome |
|---|---|
| GAP: V03(c) path selector absent from inputs | Edit: section 3 now names the semantic path selector/value 4 and its bounded initialization purpose |
| GAP: V07 selected candidate/filter calls absent from outputs | Edit: section 5 inventories search result and ordered filter invocations |
| COVERAGE: R01 nonpositive do/null room | Edit: V01(h/i) cover do 0/-1 and absent containing-room result; no Town gate for null room |
| COVERAGE: R02 other forwarded result | Edit: V01(i) supplies result 3 and asserts no rewrite/removal |
| COVERAGE: R04b diagonal | No change: dependency-owned NRF-R11b/NRF-V15 covers both cited direction entries; MMC supplies direction and tests caller scale/outer gate |
| COVERAGE: R05a transition true | Edit: V03(d) logs the transition request before lifetime; actual transition remains gated |
| COVERAGE: R05b Y normalization/ordinary blocking/recording off/normal arrival | Edit: V03(e/f) distinguish all four local branches with exact cells,count,bits,endpoints and logs |
| COVERAGE: R06 nonexpiry/positive clamp | No additional case: V04's first two retained updates already assert nonexpiry; valid signed-i16 remaining cannot exceed 32767,so decrement cannot exceed 32766. Generic setter upper clamp is dependency-owned MCI/NRF evidence |
| COVERAGE: R07/R08 other masks,high-only bits,path absent | Edit: V05(f/g) exercise masks 132/388/260/389 and absent path at the collision-stage boundary. No high-only-u16 input is invented; low16 bit word is the semantic input and original upper-register contents are not a behavioral value |
| COVERAGE: R09a/b empty cells,multiple rooms,exhaustion,other geometry | Edit: V07(f/g) add filter rejection/next room/exhaustion and remaining four geometry pairs. Existing V05(d) supplies empty cells; no second empty-list matrix |
| COVERAGE: R10/R11 individual eligibility,other type gates,hostility,metadata | Edit: V08(e/f) separate each false eligibility condition and absent metadata; V09 supplies passing hostility explicitly. MHD-V16 owns complete direct type/creator cases; MHD-V02 covers early contact gates. MMC-V08 already distinguishes search/contact rejection and write order,so no complete contact matrix copied |
| COVERAGE: R12 creation types/chance components,negative count,missing owner/Pierce | Edit: V09(d/e) test negative count,item-stat fallback and missing owner/Pierce; V09(b/c) explicitly supply linked direct owner,present targets and passing hostility. MIR-V13 owns creation types/chance/local-seed cases; MHD-V03 owns other contact pierce decisions. Those duplicate matrices are not expanded locally |
| COVERAGE: R14/R15/R16 other hand-off/cooldown combinations | Edit: V10(g) distinguishes force-end without room; V11(d) distinguishes zero requested delay. MHD-V06/MHD-V13/MHD-V16 own the broader cooldown/callback/creator combinations; their integrations remain gated |
| HARDEST: V03 trace | Edit: Y-normalization/record-off/ordinary-blocking fixtures supplement existing exact endpoint/cap trace |
| HARDEST: V07 geometry/order | Edit: next-room/filter-exhaustion and remaining geometry fixtures make candidate traversal explicit |
| HARDEST: V10 callback/final result | No change: V06(a) already asserts final Remove despite callback abort; V10(e) asserts full action replacement and ignored damage return; supplied dependencies remain explicit |
| GATED: R12c and Q01..Q09 | No change: all ten remain declared with typed boundaries; added fixtures do not promote dependency effects or observations |

Root review additionally requested pathFootprintSize=1 as an explicit vector default and present
owner/target/hostility inputs for the sequential pierce fixture. Both are now explicit. The exact
four local pierce seed pairs and legacy chance/count results remain unchanged. No source replay
was needed for these input/output/fixture corrections. FULL2 is pending.

### Blind-reader FULL2 and final triage

The fresh reader independently recalculated 90 scenarios across all eleven executable families. Retained trajectories count once; separately supplied input values count separately. MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0. Discrimination was checked for all eleven families. GATED: nine question IDs, Q01/Q03/Q04/Q07/Q08/Q09 integration, Q06 observation and Q02/Q05 outside-boundary; unresolved R12c retains Q07. COVERAGE: all seventeen parent rule families map to executable vectors; no new uncovered-branch finding was reported. HARDEST: six calculation clusters.

All six HARDEST entries receive no change: V03 separately states the full endpoint, normalized increments, blocked-cell output and ten-cell cap; V05 orders the collision gates; V07 supplies literal room tests, candidate order and geometry; V08 separates search acceptance, contact refusal and last-collision mutation; V09 supplies all four full seed pairs and strict chance comparison; V10 asserts complete callback-result replacement and the final dispatcher result. These assertions were independently recalculated without a mismatch or missing choice. The nine question IDs retain their existing assertion-level limits; no producer, persistence, delegated effect or observation is promoted.

Root review read all FULL1 revisions, including both inventory additions, footprint-size default, explicit sequential-pierce owner/targets and every added scenario. The final main contract is unchanged by this triage. Independent document review is complete, with maintainer approval pending. Draft status and the exact integration/observation gates remain. This review ran no implementation tests or original-game gameplay.

### Restructure (2026-10-15)

De-duplication approved by the maintainer after the issue 442 specification review. The former
creation-pierce rule restated MIR-R15a/MIR-R15b and is removed; MMC-R12 now consumes the
creation count by reference (MIR-V13 exercises it) and keeps only the contact decrement
(MMC-R12a) and replacing integration (MMC-R12b). The former owner own-base 328 producer question
is removed in favor of MIR-Q07; the remaining questions renumber to MMC-Q01..Q08. MMC-V09 now
covers the decrement/storage boundary. Its creation-stage assertions that MIR-V13 already holds
(chance 0 skip, negative chance one failed draw) were dropped; the assertions MIR-V13 lacks (zero
own-base seed sequence 66,18,35,30 with full seed pairs and the equal-chance stop at 66/67) are
kept as case MMC-V09(d), which cites MIR-R15a. No rule behavior, label, value or vector result
changed. Review records above predate this restructure and keep their original ids; the phase-A
id map relates them to current ids. Lint after the change is recorded in the restructure report.
