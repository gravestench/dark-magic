# Evidence: Monster curse and aura skills

Contract: [skill-monster-curse-and-aura](../../specs/skill-monster-curse-and-aura.md).
Identity: [LOD113C-D2GAME](../source-manifest.md), Windows x86 LoD Expansion 1.13c.
Read dates: 2026-10-29 through 2026-11-02. No gameplay confirmation is claimed; the named text cells and narrow loaded-field observations have dated records below.

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | `0xb39f0` | none established | Selector 111 action (analyst description) | MCA-R01 | Entry listing; full effect path unresolved |
| LOD113C-D2GAME | `0xb5950` | none established | Selector 112 action (analyst description) | MCA-R01 | Entry listing; full effect path unresolved |

## Dispatch and search record

The compatible [dispatch evidence](skill-cast-start-and-do-dispatch.md) identifies
D2Game RVA `0xA19A0` and the do table at RVA `0x1074A8`. On 2026-10-29 the table
slots listed above were independently read from the original image with PE section
mapping and little-endian 32-bit pointers; each equals image base plus its listed
RVA. These establish selector-to-entry links, not archive row bindings or AI callers.
Only root-anchor and entry/output-path reading were completed; no absence claim or
six-rung search completion is made. Existing companion inventory had no contract
for this candidate. Dispatch evidence is reused only for that root.

## Static observations and limits

The initial record in this section is historical. The 2026-10-30 continuation
below supersedes its unread-helper statements only for the explicitly listed
new readings. The earlier preparation reviews do not review the continuation.

The first entry obtains a position through a helper and returns immediately on helper failure; a later path enumerates units and makes an effect call. The second entry advances the actor seed before validating the randomly selected skill record, takes the low result modulo five, and selects a record from a five-element constant table. It builds one of five effect requests and calls a shared effect helper. No reading of those helper bodies or table identities is credited.

The method was assembly reading after checking matched-C availability. No listed
handler had a matched-C entry. Unchecked raw-byte runs and unread local/imported
callees prevent verification of MCA-R01. Decompiled names are not evidence of
function identity. Original effect data, state/stat identities and imports remain
unresolved.

## Observation procedure

Use an identified vanilla Expansion 1.13c game; retain version, loaded module
fingerprints, archive profile and patch state.
Use a suitable character in the appropriate boss encounter (curse/aura actions
require the original monster row to be established first). Record difficulty,
character level, equipment and stats, actor/target identities and coordinates,
server frame, selector, skill level, entry seed and final seed. Trace entry,
dependency requests, output writes and final return for one fully specified action.
Negative-index cases are static executable checks only. Confirm effects through naturally triggered gameplay and attributed traces; never manufacture an invalid-index observation. No matching
gameplay observation has been collected for these records.

## Claim ledger and review

Initial preparation handoff: MCA-R02 was unresolved; raw-byte decoding, local/imported callees, data and integration
remain open. MCA-Q02: observation pending. The earlier selection-stage review remains historical. The expanded successful preparation contract completed three full reader passes and a final clean delta pass 4, recorded below; that final applicable review meets the stopping rule for the current contract. Explicit effect, data integration and observation gates remain open. The companion does not supply additional behavioral rules.

## Selection-stage verification

On 2026-10-29 the original image's instructions at D2Game RVA 0xB5950 through
0xB5986 were independently decoded and corroborated against the instruction
listing. They multiply unsigned seed low by 0x6AC690C5, add old seed high with
carry, store both halves, divide the low half by five and use the remainder as
an index. The original pointer operand identifies RVA 0xFB454; PE section
mapping yields the five signed 32-bit values 66, 72, 82, 87, 91. These facts
establish MCA-R01 and its five vectors without reading effect helpers: the
stage ends before the first record validation and makes no calls. No seed
promotion to runtime-confirmed is claimed. The exact effect guard and request
consumer remain MCA-R02. The prior fully gated entry is superseded by this
verified stage; the effect integration remains unresolved.

## Table-data record

2026-10-29: reused the original LoD Expansion 1.13c effective Storm
read of `data/global/excel/skills.txt`, contiguous requested/returned ranges [0,60000), [60000,120000), [120000,173535), with successful returned counts 60000, 60000 and 53535 respectively; 173535 total bytes. This establishes complete header and row coverage. SHA-256:
`465c28e83657540e0d7ae5b0de826d5657be7fba1aa539a85c3ca3aeea34b498`.
The complete bytes were parsed as tab-separated text by header name, selecting
named `skill` rows, not positional row numbers. The narrow raw decimal cells
in the contract were rechecked from this same capture. The source profile was
reported as vanilla LoD Expansion 1.13c; physical supplying archive, whole archive-set
hash and loaded compiled correspondence were not established. This reuse
retains the original collection identity and does not renew gameplay evidence.
It supports MCA-R03 only. No arithmetic or executed effect is established.

## Blind reader pass 1: findings and outcomes

Full smaller-model review: MISMATCH 0, GUESS 0, GAP 1, DISCRIM 1, GATED 4, COVERAGE 1, HARDEST 3. GAP selected identity absent from side effects: no change, because template explicitly allows intermediate selected-index/identity assertions in vectors; selection is not an externally persisted write. DISCRIM high-half update not discriminated: add MCA-V06 with multiply high and low-half carry. COVERAGE carry branch: addressed by V06. HARDEST LCG, fixed identities and text/compiled distinction: no change to behavior; explicit foundation references, engine binding and text-only gate already state these choices. Existing GATED findings remain explicit. Fresh review required after V06.

## Handler-specific remaining reading

- MCA-Q03: 0xB39F0 handler read; target 0x9FA30 read, enumeration 0x9E570, eligibility 0xE0A60, relationship 0x4ECD0 and effect 0xA09F0 unclosed; raw-byte body not fully corroborated.
- MCA-Q04: 0xB5950 selector stage independently decoded; remaining request preparation listing read, formula imports 0xAC9A/0xACCA and enumeration 0xA0AD0 with callback 0x50250 remain unclosed. MCA-R01 draw/identity selection is independent of these effects.

## Blind reader pass 2: findings and outcomes

Fresh full smaller-model review: MISMATCH 0, GUESS 0, GAP 1, DISCRIM 2,
GATED 4, COVERAGE 2, HARDEST 3. GAP selected identity absent from section 5:
no change, because the template permits intermediate identity assertions in
vectors without treating them as persisted side effects. DISCRIM unsigned
reduction lacks a high-bit-low case: added MCA-V07 (low 3582796170).
DISCRIM validation-before-draw not excluded: MCA-V07 supplies an invalid
selected-record context and explicitly asserts seed update and empty logical
selected-record validation/access log at the pre-validation stage boundary.
Neither assertion invents downstream return behavior. COVERAGE selector 111
and null actor: no change; MCA-R01 is explicitly selector 112 with a non-null
actor, so these are outside its preconditions and remain in the gated action
contract. HARDEST carry, fixed mapping and ordering: the existing V06, fixed
engine identity list and new V07 respectively make those choices explicit;
no further behavior change. The four GATED entries retain their documented
limits. This edit requires another fresh full reader pass.

## Narrow compiled-selector correspondence

2026-10-29: reused read-only loaded-record evidence from a profile reported as
LoD Expansion 1.13c,
original Game.exe identity `74fe9c092a521f7710392548c82f81544e531107db2358617e83818874db40a2`.
Independent loaded-DLL hashes and a full loose-file
override inventory were not captured. This is limited observed-profile
correspondence, not authentication of every loaded game artifact or archive.

The public reproduction chain begins at D2Game RVA `0xF829C`, follows the
data-table global pointer to its root, then reads Skills pointer at root+0xB98
and signed count at root+0xBA0. Observed count was 357. With record stride
0x23C, each requested identity selected a 48-byte live read; stored unsigned
16-bit identity, start selector and action selector occur at offsets 0,
0x2C and 0x2E respectively. Every narrow read succeeded, and every stored
identity equaled its requested identity. Only the listed fields are published.

| Requested/stored identity | Start selector | Action selector |
|---|---|---|
| 177 | 0 | 111 |
| 212 | 0 | 112 |

These observations corroborate MCA-R03's named selector bindings against the
effective text capture. They close only that selector correspondence question.
They do not establish parameter/expression/state/stat-field correspondence,
gameplay execution or whole compiled-table identity. Original specimen memory
locations explain collection only; the contract requires semantic values.
Private source-profile and collection records remain separately preserved.

## Blind reader full pass 3: findings and outcomes

Fresh full smaller-model review: MISMATCH 0, GUESS 0, GAP 1, DISCRIM 0, COVERAGE 0, GATED 12, HARDEST 3. GAP V07 selected-record collection/count lacked section 3 source: added a semantic contextual input row identifying the supplied loaded Skills collection and signed count, explicitly unused until after the selection-stage boundary. This documents an existing V07 input and changes no operation, assertion or expected value. HARDEST V06 carry, V07 unsigned/prevalidation ordering and R03 text/limited-compiled distinction: no change; existing exact traces and source-profile limits state each choice. The twelve GATED entries (R02, Q01-Q04 and V01-V07 stage limits) remain explicit and require no invented effects. No mismatch, guess, discriminator or coverage correction was needed. Next review is a fresh delta check of the added section-3 contextual input and its V07 association only, following three full passes with zero mismatches and recorded outcomes.

## Blind reader delta pass 4: final applicable review

Fresh smaller-model delta review, limited to the section 3 selected-record
collection/count input and its existing V07 association: MISMATCH 0, GUESS 0,
GAP 0, DISCRIM 0. The reader confirmed the runtime-context source, count 0
making selected identity 66 invalid when later validation starts, no context
read before the stage boundary, seed 2/0 becoming 3582796170/0, unsigned
remainder 0 and empty validation/access log. No contract change followed.
Together with the three full runs and their recorded outcomes, this meets
the fresh-reader stopping rule. Explicit effect/integration/observation gates
remain; review completion does not approve the spec or establish full parity.

## Successful action-112 preparation investigation, 2026-10-29

Prior selection evidence and unresolved Q04 were checked first. The smallest
unread slice was the remainder of D2Game RVA 0xB5950 from selected-record
validation through the effect request/handoff. Original-image instruction
decoding corroborates the complete 0xB5950..0xB5BCF listing, including the
jump-table operand, all five destination addresses and raw clear operation.
No handler, enumerator 0xA0AD0 or callback 0x50250 entry is in the matched-source
manifest; their instruction listings were the applicable primary reading view.

### Closed guard and caller boundary

R04: selected identity is checked as signed against count at 0xB5986..0xB59B6,
then signed target-state word at selected record+0x82 is checked against state
count at 0xB59BC..0xB59D7. Rejections reach 0xB5BA9 and return 0 before any
calculation/recipient operation. Success sets actor+0xC4 bit0x40, the do-handled
latch mapped by existing SCD evidence. Calls through thunk 0xAC9A request
selected-record+0x64 then +0x60, using actor, selected identity and caller level.
The retained radius is passed to the recipient processor; duration is stored
in the request. No local radius/duration clamp is present.

Import resolution from original PE import/export tables: D2Game thunk 0xAC9A
via IAT0xF8850 resolves D2Common ordinal 10786 RVA 0x51BF0. Its byte-corresponding
source view was read before further instructions: absent pool or unsigned
reference beyond pool size returns 0; otherwise the front end supplies context
(actor, selected skill, caller level) to the interpreter. Existing SCX-R11
owns that behavior and its interpreter/producer gates; the new local vectors
supply operation results rather than claim new evaluation evidence.

### Selected field names

Original D2Common strings at RVAs0x90F20/0x90F14/0x90F04 name
auratargetstate/auralencalc/aurarangecalc. The original loader's descriptor
construction at 0x6369F..0x6371D pairs these names with offsets 0x82/0x60/0x64.
This corroborates the field identities read by the action; live field values
and expression correspondence remain Q01. GDA-R11a already supplies
Param1..8 at offsets 0x148..0x164, so selected-record+0x160 is Param7.

### Closed five request shapes

The fixed code jump table at RVA 0xB5BB8 contains destinations
0xB5A93/0xB5AA8/0xB5ABD/0xB5AD2/0xB5AF4 for indices 0..4. These agree with
the independently decoded branch bodies and establish R05's exact shape.
The low32 modulo 5 result cannot exceed 4, so the fallback branch at0xB5B3F
(import0xACCA) is unreachable from this entry's established selector. It is
not used to invent a sixth curse variant. That import resolves D2Common
ordinal 10832 RVA 0x4E5E0; its semantics are not needed for the reachable path.

The semantic request map, for reproducible inspection only: game+0, actor+4,
monsterOnly+8, selected skill+0x10, caller level+0x14, duration+0x18,
primary stat/value+0x1C/+0x34, five additional stat/value slots at
+0x20..0x30/+0x38..0x48, target state+0x4C, three event ids+0x50..0x58
and callback selectors+0x5C..0x64. These offsets are evidence maps, not
requirements on an implementation or a copied binary structure. Stat25's
damagepercent name is reused from the damage-pipeline input contract; shared
GDF names cover 36/67/68. The request constructor copies Param7 directly into
both speed and attack-rate slots in index 4; it does not evaluate a reduction
formula or substitute a conventional player-spell interpretation.

The call at0xB5B96 goes to D2Game0xA0AD0 with stack arguments
(filter 3, actor, radius, callback 0x50250, request). Its return is forwarded;
there is no replacement success constant. The actual original call target is RVA 0xA0AD0; proposed lead 0x9FAB0 was rejected by this direct target check.

### Recipient continuation read, still gated

The complete enumerator0xA0AD0 listing was read and original-image decoded.
It rejects absent callback, failed target-coordinate production or missing
resolved room; after setup it enumerates room-unit candidates, checks squared
distance against low32(radius*radius), then eligibility0xA0370 before calling
0x50250 with each candidate and the request. Candidate eligibility/room bounds
and coordinate helper/import contracts remain unclosed, so no recipient set
or processor-produced result is promoted by R05. Enumerating no candidates
after successful setup returns 1; callback returns are ignored by the loop.
This is a located lead, not a new verified recipient rule.

The complete callback 0x50250 listing and original-image bytes were read.
It consumes the request's monster-only gate and state, derives a primary
stat value through0x4E230 when selected, checks relationship0x4ECD0 and calls
curse-list producer0xA00B0. On success it applies the five additional stat
slots through0x4E230 and imported stat-list setter, and handles the event
descriptors through removal0x37110 and registration0x37190, selecting callbacks
from the fixed event-function table. Actual curse-list eligibility/lifetime,
value conversion, stat-list writes and event callback effects remain Q04;
no effect is inferred solely from request construction.

### Reused effective text cells and limits

The preserved complete original-profile Skills capture already identified
above was parsed by named headers again. Selected identities 66/72/82/87/91
have names Amplify Damage/Weaken/Life Tap/Decrepify/Lower Resist, respectively.
Each has auralencalc raw `ln34` and aurarangecalc raw `ln12`; the matching
auratargetstate raw names are amplifydamage/weaken/lifetap/decrepify/lowerresist.
Each Param7 text cell is empty. These are narrow text observations only:
no numeric compiled Param7 value or actual expression output is inferred.
The request mapping follows fixed original code even where it differs from
a reader's expectation for a named player spell. R03's existing text-only and
limited-selector correspondence labels are not upgraded to gameplay evidence.

### Initial expanded-contract review and runtime status

R04/R05 and vectorsV08..V15 are new supplied-input preparation assertions.
They require fresh full blind review, separate from the completed earlier
selection-only review. No gameplay vector was collected in this continuation.
Original-image instruction decoding supplied the corroboration.

## Expanded preparation full review 1: findings and outcomes

Fresh full smaller-model review of the expanded contract: MISMATCH 1,
GUESS 0, GAP 0, DISCRIM 3, COVERAGE 4, GATED 5 groups, HARDEST 3.
The reader independently recomputed V01-V15 arithmetic, request shapes and
operation order correctly. MISMATCH: section 5 and acceptance blanket final-return
disclaimers contradicted R05's supplied-return forwarding. Edited those summary
statements to distinguish actual processor-result production/effects (gated)
from the verified request and forwarding of a fixture-supplied return.

DISCRIM selected identity bounds: negative selected identity is unreachable
from the fixed five-choice R01 mapping (66/72/82/87/91); no changed choice
table or artificial negative selection is introduced. Existing mapping vectors
establish this. Equality at the selected-record bound is reachable with a
supplied count: added V16, skill count 66 and seed 1/0, asserting the retained
draw and empty individual-record-read log. DISCRIM target-state upper bound:
added V17 with state=count 100. DISCRIM signed nonpositive radius/duration:
added V18/V19 with returns(-3,-7) and(0,0), asserting an unclamped recipient
request and the normally returning supplied result. These changes test
existing R04/R05 behavior, not new downstream defaults.

COVERAGE the same bounds and signed-return branches are addressed as above;
selector 111 remains expressly gated under Q03, with no inferred aura behavior.
HARDEST supplied-record versus compiled-profile values: no change, Q01 and
the fixtures state that distinction. HARDEST selected rather than invocation
identity and radius-first order: no change, explicit request logs and contexts
already discriminate them. HARDEST fixture return forwarding versus actual
recipient execution: addressed by the summary correction above. GATED groups
R02/Q01/Q02/Q03/Q04 retain their exact limits. Fresh full review of the resulting
contract remains required; prior selection-only review remains historical.

## Expanded preparation full review 2: findings and outcomes

Fresh full smaller-model review: MISMATCH 0, GUESS 0, GAP 2. Raw DISCRIM
19/19 reports nineteen passing checks, not nineteen defects: normalized
actionable DISCRIM 0. Raw coverage enumerates nineteen covered vectors and
O01; normalized uncovered COVERAGE 0. All V01-V19 numeric values, request
shapes and ordering were independently recomputed correctly.

GAP target-state field mapping absent from shared GDA inventory: added the
explicit existing local loader-and-consumer evidence reference in section 3
and R04. GAP aura-length expression mapping absent from shared GDA inventory:
the same local record explicitly identifies the two expression fields; added
that reference and made the reference/result widths explicit. GDA-R11a is
consumed only for shared record/Param7 schema. No shared foundation edit or
invented broader foundation coverage was made. Q01 now states remaining
value correspondence rather than falsely gating established field names.
These edits resolve authority/reference ambiguity without changing operations,
inputs, outputs, execution boundaries or the actual remaining value gates.

GATED raw two core areas are selector 111 / Q03 and recipient 112 / Q04, with
Q01 data-value correspondence and Q02 observation limits retained separately;
none was treated as a request for invented behavior. HARDEST five raw entries:
V06 carry, V07 unsigned reduction, V16/V17 strict bounds, V12 Param7 copies
and monster-only flag, V18/V19 unclamped forwarding. No behavior changes:
the existing hand traces, exact requests and fixture assertions state each
choice. The corrected references require the next fresh full review; the
earlier completed selection-only review is not reused for the expansion.

## Expanded preparation full review 3: findings and outcomes

Fresh full smaller-model review: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1,
COVERAGE 1, GATED 5 groups, HARDEST reported 3 grouped narrative findings.
All nineteen existing vectors' arithmetic, request payloads and ordering
recomputed correctly. The corrected local field-evidence references passed.
DISCRIM processor return fixtures only 0/1 did not exclude Booleanization:
added MCA-V20 with a normally returning supplied integer 37 and exact final
result 37, explicitly without claiming that the actual processor produces 37.
COVERAGE that same return-value distinction is addressed by V20. No rule,
gate, execution boundary or ownership changed; only the fixture and its
Id-map/acceptance references were added.

HARDEST V06 carry, V07 unsigned/order, V12 both Param7 copies and monster-only,
and V18/V19 unclamped results were narrated in three groups: no changes,
because their existing exact traces and request assertions state each
distinction. GATED Q01/Q02/Q03/Q04 and the actual-producer/effect-result versus
fixture boundary retain their documented limits. With three full preparation
reviews and every finding triaged, the next fresh delta check covers V20
and its Id-map/acceptance references only. No unchanged source was reread.

## Expanded preparation delta review 4: final applicable review

Fresh smaller-model delta check of V20 and its Id-map/acceptance references:
MISMATCH 0, GUESS 0, GAP 0, actionable DISCRIM 0. The reader confirmed
the independent fixed index-0 prepared fixture, supplied processor return 37
and unchanged final result 37, with no claim that the actual processor
produces 37. It verified R05's Id map includes V20 and acceptance names
V08-V20. Each finding had outcome no change: the fixture, exact-forwarding
assertion and boundary statements already match the rule. No rule, vector,
gate or evidence label was changed after this pass.

With the preceding three full preparation reviews and every recorded finding
resolved or explicitly gated, the current expanded contract meets the
fresh-reader stopping rule. Earlier pass records retain their historical
scope and are not individually treated as final approval. The spec remains
draft: selector 111 effects, actual selector 112 recipients/list/stat/event
execution, original-profile non-selector data correspondence and gameplay
confirmation remain open. Reader completion does not establish those effects
or infer user approval.

## 2026-10-30 recipient continuation

This continuation reads the actual callers and consumers beyond the reviewed
preparation stage. Its new contract material requires fresh blind review. No
gameplay or complete state/list integration is established by these readings.

### Original instruction and dependency record

The source manifest's original D2Game and D2Common images were freshly hashed:
D2Game SHA-256 `24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53`;
D2Common SHA-256 `59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10`.
Original-image decoding corroborated D2Game ranges 0xB39F0..0xB3B15,
0x4ECD0..0x4ED5F, 0x1560..0x15C3, 0xDFA90..0xDFADF,
0x4E230..0x4E291, 0x9E570..0x9E5FF, 0x4E760..0x4E9CF,
0x4F090..0x4F36F, 0xA0370..0xA074F and 0xA09F0..0xA0ACF.
The earlier fingerprint-compatible complete 0xA0AD0/0x50250 decoding is reused.
No byte-matched source entry exists for the Game entries just listed. D2Common matched
sources 0x12030, 0x120F0, 0x12250 and 0x3BA20/0x3BA60 were read for collection
initialization, capacity growth and room geometry. These names are analyst
labels; semantics come from the instructions and requests, not those labels.

| Caller / import | Resolved original target | Reading and limit |
|---|---|---|
| Game 0x9E570 -> 0xA502 | Common #10591 / 0x12030 | Allocates initial fifteen-entry collection and stores query inputs |
| Game 0x9E570 -> 0xA4F0 | Common #10052 / 0x12720 | Full collection loop read; room-range/ordering integration remains Q03 |
| Game 0x9E570 -> 0xA478 | Common #10331 / 0x2FE10 | Actor room accessor |
| Game 0x4E230 -> 0xA7F6 | Common #11104 / 0x1A2D0 | Hireling class predicate; compatible CSA-R02/CBL-R04a consumption |
| Game 0x4E230 -> 0xA3E2 | Common #10587 / 0x38C20 | Base list own-array stat read; RLI-R02a/RLI-R02b |
| Game 0x50250 -> 0xA6E2 | Common #10188 / 0x3A7B0 -> 0x3A280 | Matched null-list wrapper read; replacing writer effects remain STL-R06c/STL-Q07 |
| Game 0x50250 -> 0xA514 | Common #10830 / 0x391F0 | Alignment from state 105/stat 172; PVP-R05a/RLI-R02a/RLI-R02b |

The original 0x4E230 selector bytes are `00 00 01 00 01 00 01 00 01 00`,
covering stat identities 36..45. Table entries select the reduction branch
for 36,37,39,41,43,45 and the unchanged-value branch for the other four.
For a nonpositive reduction on a nonplayer, nonhireling target with base-own
stat at least 100, the signed multiply/shift sequence computes truncation
toward zero of value/5. This is not a read of the target's resolved total.

### Narrow loaded Skills correspondence

On 2026-10-30 an independently loaded original-profile game reported LoD
Expansion 1.13c and the LoD 1.13c schema. The version registry's Game.exe
identity was SHA-256 `74fe9c092a521f7710392548c82f81544e531107db2358617e83818874db40a2`; no game-data override was requested.
Patch-state output was retained. Independent hashes of every loaded DLL and
the whole archive set were not collected, so this observation is limited to
the captured profile and is not gameplay confirmation.

The game's effective `data/global/excel/skills.txt` was read in contiguous
requested ranges [0,60000), [60000,120000), [120000,180000). Returned counts
were 60000,60000,53535; total 173535 and SHA-256
`465c28e83657540e0d7ae5b0de826d5657be7fba1aa539a85c3ca3aeea34b498`.
The complete file matches the earlier capture. Header-based `skill`/`Id`
selection, not a row-number assumption, supplies the named cells below.

A read-only loaded-data chain independently followed Game RVA 0xF829C through
the shared data root, then Skills pointer/count +0xB98/+0xBA0. Count was 357;
each requested record returned 572 bytes (stride 0x23C). Stored u16 identity
at +0 matched the requested identity for all seven records. The earlier local
loader/consumer field maps supply signed state +0x82, references +0x60/+0x64,
`calc1` +0x138 and i32 Params +0x148..+0x164. The pool pointer/count at +0x40/+0x44
reported 5891 bytes. Two read-only ranges [1200,2200) and [3550,3770) returned
1000 and 220 bytes, covering each cited complete expression. Runtime addresses
and full records are not part of the contract.

| Named Skills row / Id | Target state i16 | Length/range reference u32 | Param7 i32 |
|---|---|---|---|
| Amplify Damage / 66 | 9 | 1287 /1290 | 0 |
| Weaken /72 | 19 | 1504 /1507 | 0 |
| Life Tap /82 | 58 | 1799 /1802 | 0 |
| Decrepify /87 | 60 | 1940 /1943 | 0 |
| Lower Resist /91 | 61 | 2065 /2068 | 0 |
| FetishAura /177 | -1 | 0xFFFFFFFF /3599 | 0 |
| MonCurseCast /212 | -1 | 3716 /3719 | -50 |

Every listed length expression is complete bytes `04 02 00`; every listed
range expression is `04 00 00`, except FetishAura has only its range expression.
SCX-R01/R04/R11/R20/R21 therefore establish their parameter-2/parameter-0
evaluation without assuming a compiler grammar. The selected five rows'
Param1/2 and Param3/4 values separately matched their named decimal text cells:
66: (3,1)/(200,75);72: (9,1)/(350,60);82: (4,1)/(400,60);
87: (6,0)/(100,15);91: (7,1)/(500,50). Life Tap `calc1` reference 1805 has
complete bytes `04 04 00` and Params5/6=(50,0). Decrepify `calc1` is
0xFFFFFFFF, an absent calculation under SCX-R11. Empty selected Param7 text
cells were not interpreted as zero: zero is the separate loaded observation.
FetishAura Params1..4=(10,50,5,10) matched their four named text cells.

An effective `data/global/excel/states.txt` read requested [0,60000), returned
20920 of 20920 bytes. Header-based explicit id selection found state names
9 amplifydamage,19 weaken,58 lifetap,60 decrepify,61 lowerresist,68 evade.
The first five raw `curse` cells are `1`; each raw `exp`,`aura`,`life` cell
is empty. State68 has empty `curse`. These state metadata cells have no new
compiled-mask corroboration; their raw names do not upgrade list execution.

### Action 111 actual local sequence and limits

Original instructions establish coordinate helper 0x9FA30 first; on success
the actor do latch precedes collection setup. Search extent is wrapping
Skills.Param4 +2*callerLevel, with invoking skill 177 in the named profile.
The query descriptor filter is 0x583. Collection enumeration then examines
each candidate in order. The first relationship call is 0xE0A60 with actor
in ECX,candidate in EAX,game on stack; a nonzero return skips the candidate.
The class test admits unsigned identities 141..145 or 396..400. It then calls
0x4ECD0 with candidate in ESI and stack(game,actor), requiring nonzero.

The latter checks MonStats2.mWL, monster-data exclusion bit, target flags
bits 1/2/3, and availability0xDFA90 before calling 0xE0A60 with the same
actor,candidate,game inputs and normalizing its nonzero return. The intervening
0x1560 and0xDFA90 bodies write only scratch registers/stack; neither mutates
the relationship graph. With unchanged supplied relation/owner/tier inputs,
the two opposite tests exclude the effect call. This stability-qualified
conclusion does not assert an unconditional engine-wide absence, exclude
concurrent mutation, or invent a natural successful aura case.

If the later test succeeds, 0xA09F0 requests state 68/stat 25 with duration
wrapping Param2*L and value Param1+(L-1)*Param3, actor source and candidate
target, invoking skill/level and default callback. It forwards to 0xA00B0.
The whole action releases the collection and returns 1 after normal collection
completion. Coordinate failure forwards its zero result. Valid invoking skill
record and successful allocation remain preconditions; malformed-record faults
are not replaced by a safe default. Actual collection and complete state/list
effects remain MCA-Q03, with the precise shared owners below.

### Action 112 actual consumer and remaining owners

0xA0AD0 uses filter 3. 0xA0370 consequently admits only players outside modes
0/17 and monsters outside 0/12. Other kinds fail; other filter-bit checks are
inactive. Geometry/room iteration precedes this predicate. 0x50250 applies
its optional monster-only/alignment/AI-control predicate, then transforms a
present primary stat through 0x4E230. A transformed zero returns 0 before
relationship/state processing. Next0x4ECD0 eligibility precedes0xA00B0.
Null applier return returns 0; a non-null list permits the additional stat
loop, then event cancellation/registration, and returns 1. Its result is
ignored by the outer collection loop. These return meanings are separate
from the action's forwarding of the outer processor result.

The new loaded Lower Resist Param7=0 means its prepared primary value is 0.
If its earlier monster-only guards pass,0x4E230 leaves zero and the callback
returns 0 before state/list/event operations. This conditional profile result
does not claim which AI guard naturally passes or an observed cast result.

The original callback table at Game RVA0x1077A8 resolves selector 4 to 0x4F090
and selector 5 to 0x4E760. Their complete bodies were read and original bytes
decoded. The earlier description of selector 4 as retaliation was a lead, not a verified role claim. Its exact combat roles are established in the callback continuation below through `calc1`/2/3 and
DE0C0/DCF20 plus a class 290 tail; selector 5 computes a source-skill percentage
of physical damage and a life write/overlay request. A named skill's expected
player-spell meaning cannot substitute for these fixed bindings. Full
callback contracts and their dependency effects remain MCA-Q04.

Generic0xA00B0 is the same consumer covered by OWD-R06a/R06b/R07a/R07b/R08.
For these curse states, unresolved OWD-R06b is relevant rather than skipped.
Replacement/detach/free retain OWD-R07b/STL-R05; attachment and replacing
list-stat effects retain STL-R06b/STL-R06c/STL-Q07; unit-write propagation and
refresh retain RLI-Q03/RLI-Q02. IEP-R01/R05/R07 own explicit event insertion,
removal and traversal, but callback arithmetic is not established by merely
registering its selector. No shared rule is silently upgraded here.

### Read-only monster-control guard closure

For the captured index-4 state 61, 0x50250 supplies control 0 to 0x9F120.
The complete chain0x9F120 ->0xB0B20, plus0x22A80,0x9C9A0,
Common #11085/0x31200 and0x1400, was read. There are no state/list/stat/event
writes or draws in that chain. The Common0x1400/Game0x1560 check is mode 2
availability `MonStats2.mWL` at+0xF0, not `isAtt` at+0x04 bit 9. GDA-R09a
and its original descriptor correspondence distinguish those fields.
Common0x31200 additionally requires !MonStats.boss and MonStats.switchai.
Game0xB0B20 rejects missing/nonmonster units, control above 19, state 54,
failed Common0x31200, the target-flag/availability condition and monster
condition mask 10. Control0 reaches0x9C9A0's immediate true branch after
those checks. Game0x9F120 separately normalizes the class through signed
MonStats word+2; canonical class 492 with state 143 rejects. Other paths
continue through 0xB0B20. The earlier 0x50250 alignment read is read-only under
PVP-R05a/RLI-R02a/R02b. These readings support R12's exhaustive guard-fail /
guard-pass then zero-primary result without supplying a successful guard.

Original Game0x9F120..0x9F19F,0xB0B20..0xB0B9F and0x9C9A0..0x9C9FF
were decoded from the fingerprinted image. Common0x1400's two raw-byte
instructions are the same field reads/masks as the separately decoded
Game0x1560; direct Common-image corroboration is retained with the continuation
audit. None of these new Game/Common entries has a matched source entry.
The monster-data exclusion condition in R10 means present MonsterData byte
+0x16 mask 0x20; missing MonsterData does not reject at that check. It is
distinct from the mode-availability and actor seed conditions.

### Natural encounter attempt and limits

2026-10-30, Normal difficulty, original-profile level 21 Barbarian, life 95,
base strength 30/dexterity 20/vitality 25/energy 10, no assigned level-up points:
ordinary movement reached ActIII level 76 from town 75. A nearby class 301 pack
reduced life to 0 before an identified fetish-shaman encounter. No action 111
or 112 entry, selected request, applied state/stat or final return was
attributed. This is an unsuccessful encounter attempt, not a matching
observation or a negative proof that these actions are absent.

### Continuation self-review and current review gate

Self-review checked field identities against GDA's original schema and
corrected the new mode-availability attribution to `mWL`, before blind review.
New R06..R12 and V21..V26 are outside the previous preparation review.
All prior review records remain historical for their own scope; fresh review
of this continuation is pending. Complete lifecycle integration and gameplay
confirmation remain open, independently of the new local proofs.

## Recipient continuation full review 1: findings and outcomes

Fresh full smaller-model review: MISMATCH 0, GUESS 1, GAP 2, DISCRIM 2,
COVERAGE 6 advisory groups, HARDEST 3. No complete-effect gate was removed.

GUESS availability attribution: R10's abbreviated SFR-R03e reference could
mean its entire searcher/room/town/condition-A recipe. Replaced it with the
exact local candidate-only Unavailable definition from the already read
0xDFA90, and explicitly limited shared consumption to that subcondition.
The local0x4ECD0 does not request actor availability, room/town or condition A.
GAP V24 actor availability therefore requires no new actor requirement:
case(d) now supplies an unavailable actor and asserts the candidate checks
still occur after its first relationship read. This resolves the ambiguity
against actual source rather than importing unrelated SFR checks.

GAP V26 guard production: added the exact fixed control-0 guard recipe to R12
and its semantic loaded-record/state/condition inputs to section 3. This is
the already read0x9F120/0xB0B20/0x22A80/0x9C9A0/Common0x31200/0x1400 chain,
now stated in the contract. No successful guard result is supplied. The
false-flag path specifically requires Unavailable=true; a true target flag
bypasses that read. The canonical 492/state 143 early rejection is separate.
V26's passing case now derives each step from its declared inputs; new cases
for state 54, false flag with available target, and alignment 1 assert rejected
prefixes and absence of the primary transform.

DISCRIM negative modes: V22 adds player/monster mode-1, both accepted, so
rejecting every negative mode fails. It also adds player 12/monster 17 accepted
to distinguish the two kinds' excluded mode sets. DISCRIM reordered second
relationship: V24(d) supplies a false required target flag and asserts exactly
one relationship read, with target predicates examined first. Moving the
second read before required flags fails the count. The target-predicate log
also excludes an invented actor-availability gate.

COVERAGE six advisory groups: negative modes and other-kind-specific modes
are covered by V22's additions. R10 guards are covered by V24(d/e/f/g/h):
false required flag, missing mWL, monster exclusion, unavailable candidate,
and absent monster data, respectively. A null candidate with a present valid
actor returns nonzero at the preceding PVP-R05a relationship read and skips
before the explicit null check; no artificial relationship result was added.
Actor availability is not consumed by this local predicate, as clarified.
R10's second-read success followed by R11 is not manufactured from unchanged
owner/relation inputs: the opposite tests preclude it in that context. R11's
isolated builder vector remains independent, while full integration stays Q03.
R11 negative level now has V25 L=-2, yielding duration-100/value-5 without
clamping. R12 guard rejection has the three new V26 prefixes above. Other
fixed-control guard rejections retain the same zero/no-effect outcome and
the explicit R12 recipe; no speculative successful state effects were added.

HARDEST R10's opposite tests and ordering: clarified and discriminated as
above. HARDEST R12's actual guard then zero primary: made executable from
section 3/R12, without using companion-only behavior. HARDEST R06/R07 data
versus evaluation: no change; the dated narrow profile and complete captured
expression evaluation remain separate, and no gameplay promotion follows.
Q01/Q02/Q03/Q04 retain their full remaining integration/observation limits.
This revised continuation requires a fresh full reader pass; the earlier
preparation stopping rule does not review these changes.

## Recipient continuation full review 2: findings and outcomes

Fresh full smaller-model review: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1,
COVERAGE 2 advisory groups, HARDEST 3. Its raw GATED 42 counts subrule/vector
boundaries, not forty-two distinct open gaps; Q01/Q02/Q03/Q04 remain the
existing full integration/observation gates.

DISCRIM V26 guard order was narrated but not asserted. Added exact ordered
logical guard/field-request logs for every V26 case. The canonical-link read
is distinguished from the later switchRecord/modeRecord eligibility checks:
the canonical MonStats use legitimately precedes state 54, while the switch-AI
definition checks follow it. V26(c) now rejects extra switchRecord/modeRecord
reads before its active-state 54 rejection; V26(e) rejects such reads after
failing alignment. This tests logical consumption, not CPU prefetches or a
required binary storage layout. R12's behavior and all gates are unchanged.

COVERAGE R10 second class range: V24 adds classes 396/400 and adjacent 395/401,
with exact relationship counts 2/2/1/1. Its second-read success-to-R11 remains
unmanufactured for unchanged relation inputs: opposite tests exclude that
transition in this fixture context. V25 separately executes the actual
request-builder contract; integration remains Q03. COVERAGE R12 rejection
prefixes: added V26(f), canonical-base 492 plus active state 143, with its
exact early read log and no state 54/switch-AI checks. The other advisory
invalid-link/mWL/boss/switchai/condition-mask variants retain the explicit
R12 recipe and zero/no-effect result, without claiming exhaustive fixture
coverage of those rejection prefixes. The passing case and state 54/flag/
alignment/canonical rejection cases now assert their complete prefixes.

HARDEST V06 LCG carry: no change; its exact hand trace remains reviewed
selection-stage evidence. HARDEST V24 opposite relationship/order: retained
and extended only for the second class range. HARDEST V26 derived guard then
zero primary: strengthened with the ordered logs above. No fixture-supplied
guard success, stored effect or gameplay confirmation was introduced.
All findings are triaged; the revised V24/V26 require a fresh smaller-reader
check before this continuation reaches its stopping rule.

## Recipient continuation full review 3: final applicable review

Fresh full smaller-model review: MISMATCH 0, GUESS 0, GAP 1, DISCRIM 0,
COVERAGE 5 advisory groups, HARDEST 3. All Q01/Q02/Q03/Q04 gates were retained.
No contract, input, vector, expected output, acceptance, ownership or gate
was changed in response to this pass.

GAP selected skill identity absent from section 5: no change. R01 explicitly
states, "The selected identity is an intermediate result" and ends before
effect application, recipient selection or final return. Its persisted write
is the actor seed update, already listed in section 5. V01..V07/V16 assert
the computed choice/identity to check that algorithm, not an additional unit,
record or game-state write. The vector convention explicitly permits named
intermediate results, including a selected index, while section 5 lists writes
and effects. Adding the local selected identity as a persisted effect would
blur the existing boundary. The review entry is resolved by that existing
contract/template distinction; no behavioral edit or new assertion is needed.

COVERAGE R04 negative/above-count selected identity: no change. A negative
choice is unreachable from R01's fixed five positive identities; no altered
choice table is introduced to manufacture it. The signed upper-bound branch
is explicit and its equality boundary is exercised by V16; a greater identity
uses the same rejection/result with retained prior seed, so another such case
adds no distinct effect. COVERAGE R09 stat 39/41/43/45: no change. These fixed
selectors reach the same adjustment path already exercised for 36/37, while
V23's stat 38 case distinguishes the excluded branch. The full fixed set is
specified; this does not claim a separate fixture for every table entry.

COVERAGE R10 required target bits 1/2 independently false: no change. The
rule requires their conjunction with bit 3; V24(d) exercises false required
flags, checks the short-circuit before Unavailable/second relationship, and
asserts the resulting empty effect log. Per-bit machine-read order is not a
new contract assertion. COVERAGE R12 invalid class/link, mWL/boss/switchai or
condition-mask rejection: no change. R12 explicitly derives those guarded
conditions and the zero/no-effect result. V26 asserts the passing definition
path and several distinct early rejection prefixes; exhaustive prefix fixture
coverage is not claimed. COVERAGE R12 absent monster data: no change. R12's
mask test explicitly supplies zero for missing data, so with V26(b)'s other
inputs its control-0 path still reaches the same zero primary and callback 0.
This follows the already read null guard, without borrowing R10's different
missing-data case as a gameplay observation.

HARDEST R10/V24 opposite relationship and ordering: no change, exact counts,
required-flag log and second-range cases already distinguish the contract.
HARDEST R12/V26 derived guard then zero primary: no change, its exact ordered
logical logs distinguish canonical and switch-AI uses plus rejection prefixes.
HARDEST R06/R07/V21 data versus evaluation: no change, narrow dated loaded
values and captured expression evaluation remain distinct from gameplay.

The three full continuation reviews and the recorded outcome of every entry
meet the fresh-reader stopping rule for this resulting contract. This
evidence-only record changes none of the elements that require a new pass.
Review completion establishes neither user approval nor full reproduction:
the spec remains draft, with its complete in-scope effects and attributable
gameplay observations still subject to Q01/Q02/Q03/Q04.

## Actual outer-processor continuation, 2026-10-30

The completed three-pass recipient review above applies to the contract
through R12/V26. New R13/V27..V29 follow new callee and original-image
evidence and require fresh review. The earlier stopping rule is not applied
to this expansion. No gameplay confirmation or status promotion is made.

### Action-111 complete local purity and argument proof

The original 0xB39F0/0x4ECD0/0xE0A60 call instructions and register/stack use
were followed through both predicate invocations. Let E be 0xB39F0's entry
stack pointer. The saved game is at E-0xC; after the later saved EBP, loop
stack pointer is E-0x1C, so its `[esp+0x10]` reload is the same saved game.
The first call takes candidate in EAX, actor in ECX, and a pushed game copy.
The later 0x4ECD0 call receives candidate in ESI and copies of game/actor on
the stack. After its two pushes, offsets+0xC/+0x10 select that game/actor;
its nested 0xE0A60 receives the same three identities.

0xE0A60 preserves ESI/EDI through its normal returns;0x1560 preserves ESI;
0xDFA90 does not alter those candidate/actor registers;0x4ECD0 restores its
saved EDI. Its `[esp+4]` writes target its own pushed-ECX scratch slot, not
either caller argument.0xE0A60's `[esp+8]` scratch write is likewise local.
Common 0x391F0 overwrites its own copied stack argument with layer 0 before
tailcalling its stat reader. That copied argument is not an address of the
handler's stored game or an actor/candidate object field. No output pointer
to handler storage is passed into this relation-read graph.

The complete normal-return graph was checked, not only the first predicate:

| Original node | Applicable successor / complete read |
|---|---|
| Game 0xE0A60 | Registry owner lookup 0xE03A0; two-player directional relation 0xF02A0; Common #10666/0x84D30 |
| Game 0xE03A0 | Reads game kind/id bucket and unit-next links; no persistent store |
| Game 0xF02A0 | Common #10920/0x2EB70 player-data getter, then reads directional node/flags; no insertion or flag write |
| Common 0x84D30 | Calls 0x391F0 twice; tests existing tiers |
| Common 0x391F0 | State105 list lookup 0x38520 ->0x37F80; selected-array read 0x38BD0 ->0x38A80 ->0x382B0 |
| Common 0x37F80 | Existing primary/secondary list traversal; no write/callback |
| Common 0x38A80 | Existing selected array, binary-search index, optional minimum adjustment; no write |
| Common 0x382B0 | Byte-matched source binary search, local index variables only |

Common 0x2EB70 and 0x382B0 have available byte-matched sources, read before
using their new purity facts; the remaining nodes have no matched source
entry and their full instruction bodies were used. Original-image decoding
corroborated Game 0xE0A60..0xE0B3F,0xE03A0..0xE040F,0xF02A0..0xF02EF and
Common 0x84D30..0x84D9F,0x391F0..0x3923F,0x38520..0x3854F,
0x37F80..0x37FCF,0x38BD0..0x38C1F,0x38A80..0x38AFF and 0x382B0..0x382EF.
Imports were resolved against the fingerprinted original PE export table.
The two-player getter's fatal-assert branches are outside the established
valid-player path; malformed registry kinds can fault and cyclic owner links
can fail to terminate under PVP-R07. No normal-path allocation, callback,
RNG, message pump or persistent relation/tier write was found in this complete
applicable graph. The alignment read uses the list's selected array, which
is its own array for a nonextended list; it does not recalculate attached
contributions or mutate a cached total.

This closes local ABI/alias/purity uncertainty for normally terminating,
well-formed calls. It strengthens R10's existing unchanged-input consequence.
It does not establish an engine-wide single-writer/concurrency invariant;
an unconditional no-effect/unused classification still lacks that evidence.
Actual collection and generic effect execution retain Q03. No rule or status
was changed merely to bypass those remaining assertions.

### Actual action-112 outer producer

The previously read/decoded Game 0xA0AD0 body was connected to every callee
reachable with fixed filter 3. MTP-R09 supplies 0x9FA30 refresh/coordinates;
only MTP-R04's 0x1280 rectangle-search subrecipe is consumed, without its
direct-teleport caller's linked-monster fallback or relocation.

| Game thunk / local body | Resolved original target | Established consumer fact |
|---|---|---|
| 0xA55C | Common #10383/0x3BA20, byte-matched source | Copies stored room-array pointer/count; adds no room |
| 0xA6D0 in 0x9E610 | Common #10561/0x3BA60 room-rectangle getter, byte-matched source | Copies rectangle into local output storage |
| 0xAE26 | Common #10769/0x34580, byte-matched source | Wrapping(dx*dx+dy*dy) |
| 0xA3B8 | Common #10750/0x34B80, byte-matched source | Candidate dynamic Y unsigned high16 |
| 0xA454 | Common #10867/0x34BB0, byte-matched source | Candidate dynamic X unsigned high16 |
| 0xA0370 | Previously decoded local predicate | R08 fixed filter 3 kind/modes |
| 0x50250 | Actual fixed curse recipient | R12 produces 0 for a player in captured index 4, before other helper effects |

Original Game 0x9E610..0x9E67B decoding corroborates the three raw LEA runs
and all branch targets. The first successful X-side comparison skips its
opposite-edge comparison, and a successful Y-side comparison admits the
room. R13 therefore preserves the actual ordered signed recipe rather than
assigning a conventional geometric meaning to the analyst label. With
center(100,100),radius 7, both rectangles(90,90,40,40) and(130,90,40,40)
pass; the second lies outside conventional radius/rectangle overlap. The
following candidate distance check is a separate signed inclusive comparison
of wrapped squared values. Candidate positions are read Y then X, unlike
the center helper's X-then-Y sequence. Next-unit links are read after any
normally returning callback, whose return value is ignored.

The outer producer actually returns 0 after coordinate or point-room failure;
it returns 1 after normal traversal, including an empty array or zero
recipients. V27 executes R01/R04/R05/R06/R07/R08/R12/R13 together using
captured index 4 data and supplied readable world records. Both reached
player callbacks actually return 0 under R12, yet the final action return is 1.
No supplied successful processor/state/list result is used. Its exact
distance values 50,25,64,49,1089 and seed 1791398089/0 were independently
recomputed. V28/V29 separately distinguish coordinate failure, absent initial
room and empty-array completion, with preparation writes retained.

This closes Q04's actual outer-coordinate/room/traversal/produced-return
slice at its declared valid-storage/normal-continuation boundary. Complete
generic curse-list effects, effects on a traversal in progress, event
registration/callback arithmetic, and OWD/STL/RLI propagation remain Q04.
The scope is not reduced to the newly executable index 4 case; all other
selected branches and their complete effects remain required.

The complete alignment-read chain also corrects a consumed-field detail:
Common 0x391F0 does not unconditionally call the own-array getter 0x38B00.
It reaches 0x38BD0/0x38A80, selecting the resolved array for an extended list
and the own array otherwise, with RLI-R02b's minimum adjustment. Section 3
now names that existing selected-array result; V26(b)'s state 105 list is
explicitly nonextended, retaining all earlier results. New V26(g) makes it
extended with own 172=0/resolved 172=1 and rejects at alignment 1, so an
incorrect own-array substitution is discriminated. No shared authority was
edited and no attached-stat production was inferred from this read.

R13, V26(g), V27..V29 and their input/output/acceptance references require
fresh full blind review. Self-review checks original Y-then-X candidate
reads separately from the center helper, signed distance comparison,
the unusual room prefilter, actual callback 0 versus final outer 1, and the
scope of the MTP dependencies. All remaining lifecycle/observation gates
remain explicit; no new runtime case was collected for this expansion.

## Actual traversal full review: final applicable findings and outcomes

Fresh smaller-model full review of the resulting complete contract returned
MISMATCH 0, GUESS 0, GAP 0 and DISCRIM 0. Every entry in those four lists
therefore has outcome no change. The reader independently recomputed the
new actual traversal and selected-array cases without a disagreement.

GATED 7 enumerated existing limits: R02, Q01 through Q04, O01 and the
isolated-stage boundaries of V21/V25, with some grouped in the report.
These are references to the existing limits, not seven new evidence gaps.
Outcome no change: none was removed, narrowed or treated as established by
the supplied-world-record traversal.

COVERAGE 2 advisory groups, both outcome no change. R12's remaining rejection
prefixes lead to the same callback 0 and no later effect request already
distinguished by V26's ordered guard cases; the rule states the other
prefixes, and the vectors do not claim exhaustive prefix coverage. R13's
other room-prefilter branches, adjacent-room search and static or absent-path
coordinates are not all separately exercised. The exact prefilter is stated
as an ordered wrapping recipe supported by the original-image reading;
V27 discriminates its exercised nonconventional admission from replacement
by conventional overlap. Adjacent search consumes only the named MTP-R04
subrecipe, whose behavior and evidence belong to that dependency. Static
coordinate reads and absent dynamic-path zeros remain explicit R13 inputs
and operations, not supplied successful geometry results. The existing
vectors claim their listed dynamic-coordinate and containing-room cases;
this advisory does not identify a contradictory result or require an
exhaustive traversal fixture suite.

HARDEST 3, each outcome no change: V06 already supplies the exact LCG
carry/high-word result; V26 supplies selected-array and ordered guard reads;
V27 distinguishes actual callback 0 from normal outer/action result 1 and
asserts callback-before-next-link order. Their current assertions retain
those distinctions without inferring stored curse effects or gameplay
execution.

This fresh full run returns MISMATCH 0 with every finding resolved or
explicitly gated, meeting the applicable reader stopping rule for the
resulting R13/V26(g)/V27..V29 contract in its full context. This record
changes no rule, input, vector, output, acceptance assertion, ownership or
gate boundary. The spec remains draft: complete in-scope effects and
attributable gameplay observations remain open.

## Callback arithmetic and dependency continuation, 2026-10-31

This expansion adds R14-R16 and V30-V34 after the completed actual-traversal
review. A fresh full reader is required; that prior review does not cover the
new rules or additional captured callback-field assertions. Complete stored
curse effects and gameplay confirmation remain open under the existing Q04/Q02.

### Original callback identity, roles and complete local reads

The root remains action 112 at Game 0xB5950, prepared callback requests in
R05, actual recipient 0x50250, and the event-function table at 0x1077A8.
Fixed selector 4 points to 0x4F090; selector 5 points to 0x4E760. R05 requests
(5,4) for selected 82 and (1,5)/(2,5) for selected 87. Successful generic
state/event installation is still its named integration gate; these bindings
do not prove that a corresponding gameplay event occurred.

The fingerprint-compatible complete original callback decoding from the
preceding continuation is reused, retaining its 2026-10-30 identity and limits.
Both bodies have no byte-matched source entry. New local helpers Game 0x1540,
0x21260 and 0x4DF20 likewise have no matched entry; their complete listings
were read, and fresh original-image decoding corroborated all instructions,
including the raw 0x21260 branches and final registry tail jump. No analyst
label is used as semantic authority. Original D2Game and D2Common SHA-256
still match the source manifest's fingerprints listed above.

IEP-R07 supplies ECX=game, EDX=requested event, stack owner/other/record/
registered skill/level. At 0x4F090 owner is retained in EBP, other in EBX,
and game in the saved local. Its DE0C0 call pushes record, true, other, game
and sets EAX=owner. The DE0C0 prologue resolves EAX as defender and the
stack other as attacker. Its subsequent DCF20 call supplies the same
owner/other/game/current record roles. DPO-R13a/R26 consume the apply/reaction
contracts; no inferred reversal into a retaliation target is made. The old
brief retaliation wording above is superseded by these exact roles.
At 0x4E760, owner is initially ESI and other EBX; all resource reads and
replacing/overlay requests use EBX. Owner is used for target-state/list lookup.
Both callbacks obtain the list's source identifiers before formula evaluation;
source lookup with a valid kind may return null, which is passed to SCX rather than causing
an extra local failure or substitution of owner/other.

| Game thunk / local helper | Resolved original body / export | Consumer and evidence |
|---|---|---|
| 0xA5D4 | Common #10871 / 0x38520 | Existing state-list lookup; compatible complete purity reading reused |
| 0xA68E | Common #10494 / 0x33CD0 | State membership consumes RLI-R03, not new state construction |
| 0xA84A | Common #10473 / 0x38080 | Complete matched source read: list field +0x0C, source id; no write |
| 0xA7FC | Common #11131 / 0x380E0 | Complete matched source read: list field +0x08, source kind; null fallback 6 is unreachable after present-list guard |
| 0xE03A0 | Game local registry lookup | Complete compatible purity/owner lookup reading reused |
| 0xAC9A | Common #10786 / 0x51BF0 skill-calculation front end | Existing SCX-R11/R20/R21 identity/evaluation record reused; contexts and selected references checked at these callers |
| 0x14D0 | Game mulDiv | NRF-R05; exact first/second argument order retained |
| 0xA7F6 | Common #11104 / 0x1A2D0 | Hireling predicate consumed under CSA-R02/CBL-R04a |
| 0xA42A | Common #10973 / 0x38B70 | Unit stat 6 read under RLI-R02a/R02b |
| 0xA3A0 | Common #10913 / 0x390F0 | Effective stat 7 cap under LSE-R10; no separate cap formula |
| 0xA472 | Common #10887 / 0x3A740 | Replacing write requested; integration remains RLI-Q03 |
| 0xA83E | Common #10689 / 0x32C10 | LSE-R13 overlay operation; stored attach/write/update remains IEP-Q09/RLI-Q02 |
| 0xA514 | Common #10830 / 0x391F0 | Existing selected-array alignment read under PVP/RLI |
| 0x1540 | Game local, complete original-image read | Present monster -> monster data -> current MonStats pointer; no allocation/mutation |
| 0x21260 | Game local, complete original-image read | Present monster/data/control/game -> control's kind/id registry lookup; absent chain gives null |
| 0x4DF20 | Game local, complete original-image read | Distinct positive-amount capped-add request and computed amount return, not LSE-R09's gated life-add recipe |

The imported ordinals above were freshly resolved using the original PE import
and export directories; unchanged shared-body evidence is reused with its
existing labels. The matched getters' analyst names were discarded in favor
of their actual loads and the list-source producer/consumer correspondence.

Selector 5 checks positive incoming physical, then level/state/list and source,
then selected `calc1`. It reads other life then cap. It does not test positive
calculated amount or state 92; zero percentage can still request lowering an
above-cap life. Signed wrapping sum is floored at 1 before the maximum cap.
The signed overlay guard is strictly positive and below count. Original field
+0x10C is prgoverlay, not hitoverlay. Selector 4 zeroes 0x70 bytes before
setting physical/result/flags/hitClass; the contract consumes the existing
logical record field set, imposing no binary padding obligation. Result 32
is always ORed into selected ResultFlags. Its regular-monster/hireling and
other-player branches select the exact `calc1`/2/3 reference before SCX.

After normal combat/reaction continuation, selector 4 checks other kind 1/class 290.
The original compare-EAX-with-itself followed by MOV BL,100 leaves the
conditional branch untaken, so this route always obtains the current other
MonStats record and its unsigned difficulty-indexed Drain. No fixed-100
substitute is inferred from that initial BL load. It then requires alignment 0,
resolves owner's optional control-linked unit, checks link/owner availability,
and uses the new record's current physical. The ordered Drain/20/50 percent
mulDiv calls, subtract-actual-return steps and overlay requests are R16.
Its 0x4DF20 operation has no alive/state/minimum-1 gate and uses signed
strict-greater cap logic. Sequential values depend on actual replacing-write
integration; the explicit view-update fixture does not prove that integration.

### Additional narrow callback data correspondence

This reuses the complete effective Skills file and complete 572-byte loaded
records captured earlier on 2026-10-30, with exactly the same source-profile
limits; no renewed capture or gameplay verification is claimed. Text selection
uses the skill/Id headers. Named raw cells: Life Tap/82 `calc1` is `ln56`,
Param5 is `50`, Param6 is `0`, prgoverlay is `steallife`; `calc2`/`calc3`,
ResultFlags/HitFlags/HitClass are blank. Decrepify/87 has blank `calc1`/2/3,
prgoverlay and ResultFlags/HitFlags/HitClass; Param5 is `-50`, Param6 blank.
The loaded records, not blank text, establish ResultFlags/HitFlags/HitClass 0
for both, 82 prgoverlay 151 and 87 prgoverlay-1. The compiled 82 `calc1`
reference 1805 selects complete bytes 04 04 00 already captured, using
Param5/6=50/0. Its other two references and all three 87 references are
0xFFFFFFFF. No full expression-pool identity is claimed.

The local field-name association was checked against Common loader 0x61EC0,
which has no matched source entry in the Common manifest. Original decoding
corroborates descriptor stores at RVAs 0x644A8 onward for prgoverlay and
0x64944 onward for ResultFlags/HitFlags/HitClass. Original read-only column
strings are at Common RVAs 0x90B24,0x8B130,0x8B124,0x8B118 respectively.
The corresponding compiled offsets are +0x10C(signed 16),+0x12E(u16),
+0x130(u32),+0x134(u32). Callback consumers establish the widths and semantic
uses. This is local consumer evidence rather than an assertion that the
shared GDA inventory already enumerates every field.

### Expansion checks and remaining closure

V30 uses the actual selected-87 empty calculation and -1 overlay; it follows
that actual front-end result through a positive-life/cap replacing request,
without supplying a percentage or successful state producer. V31 separates
zero/negative levels, negative calculated value, wrapping addition and
resource rejections. V32 uses captured selected-82 parameter bytecode to
construct the actual local combat request and assert exact event roles;
DPO/HRC interiors are explicitly not supplied as proof of effects. V33
separates regular monster, hireling, player/other-kind and zero-level branches.
V34 follows the class-290 request distribution with an explicit sequential
logical stat-view fixture and separately checks the distinct capped-add.
All are static stage fixtures, not observed gameplay or generic storage proof.

Complete Q04 still requires actual generic curse-policy/replacement and
state-list propagation through OWD/STL/RLI, combat/reaction effects under
DPO/HRC, replacing resource effects and overlays under their owners, plus
attributable gameplay. No existing full gate was split away or dropped.
New rules, inputs, vectors, captured-field claims and acceptance references
require fresh smaller-model full review before another stable checkpoint.

The callback stage's readable-registry context does not turn malformed list
source types into a safe missing match. The compatible complete Game 0xE03A0
reading distinguishes valid kinds 0..5 with no match from invalid kinds that
can dereference null, and from an absent game assertion path. R14/R15 state
that normal-result boundary explicitly; V30 supplies valid monster kind 1
with no matching id. The action's present monster source ordinarily supplies
a valid kind, but no universal producer invariant is inferred.

## Callback full review 1: findings and outcomes, 2026-10-31

Fresh smaller-model full review of R01-R16/V01-V34 returned MISMATCH 0,
GUESS 0, GAP 0, DISCRIM 0, GATED 4, COVERAGE 3 advisory groups and
HARDEST 3. The four gated entries are the existing Q01-Q04, not newly
invented gaps. Every empty defect-list entry has outcome no change; no
conflicting value or missing behavioral input was reported.

COVERAGE R14 rejected entry guards and overlay bounds: no change. V31(e/f)
already distinguishes the callback-0/no-request outcome, and V30's negative
prgoverlay distinguishes successful replacing life without an overlay from
V31's admitted overlay. Zero or out-of-range overlay identities select the
same no-overlay result family under the explicit strict bounds; the other
entry rejections share the specified callback-0 family. No exhaustive guard
or overlay-bound fixture claim is made.

COVERAGE R15 rejected guards/nonpositive incoming physical: added V35.
The prior R15 fixtures all returned 1, so the result-0/no-record/no-handoff
branch needed an executable case. Independent incoming physical 0/-1 cases
also distinguish the positive incoming-damage guard from V33(e)'s zero
calculated damage, which still reaches combat handoffs. Other rejected
R15 guards select the now-exercised return-0/no-handoff family; no change
for duplicating every rejection prefix.

COVERAGE R16 Drain 0/100, nonzero alignment, link/owner availability and
nonpositive remaining: added V36. Its owner-only positive case distinguishes
absent link from mandatory-link behavior and uses Drain 100. Five early
continuation cases distinguish an empty effect log with final callback 1,
which was not asserted by the prior complete continuation fixture.
The ordered checks also distinguish link failure from a later owner check,
and owner rechecking after combat. This does not assert that an extra
mathematically equivalent multiply-by-100 must change a result. Negative
remaining shares the explicit nonpositive early-end branch; no separate
negative fixture is required to duplicate its callback-1/empty-effects result.

HARDEST V27 room admission/distance/traversal/callback-versus-final: no
change, its exact visit and callback/next-link logs already state those
choices. HARDEST V33 owner/hireling/other kind/level normalization: no change,
its selected references, formula contexts and handoff records discriminate
those choices. HARDEST V34 sequential capped adds and remaining distribution:
no change, the fixture explicitly supplies view updates, exact returned
amounts and ordered later reads without proving actual stored-write effects.
All Q01-Q04 remain retained. New V35/V36 and Id-map/acceptance references
require another fresh full reader; the preceding traversal review cannot
stand in for the resulting callback contract.

## Callback full review 2: final applicable findings and outcomes, 2026-10-31

Fresh smaller-model full review of the resulting R01-R16/V01-V36 contract
returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 4, COVERAGE 3
advisory groups and HARDEST 3. Every empty defect list has outcome no change.
GATED 4 names the existing Q01-Q04; these remain visible and retain their
full boundaries.

COVERAGE R12 remaining rejection prefixes: no change. V26 exercises distinct
ordered guard rejections and the guard-pass/zero-primary result. The remaining
invalid record/link, mWL, boss, switchai and condition-mask failures select
the same specified callback-0/no-later-effect result, with their exact
positions already in R12; exhaustive prefix coverage is not claimed.
COVERAGE R13 other prefilter branches, adjacent search and static/absent-path
coordinates: no change, retaining the traversal review's recorded rationale.
R13 explicitly states those operations and consumes only MTP-R04's named
rectangle-search subrecipe. V27-V29 claim their exact dynamic-coordinate,
containing-room and normal-completion cases; they do not claim an exhaustive
geometry or storage fixture suite. No missing behavioral decision was reported.
COVERAGE R14 remaining entry failures and overlay bounds: no change, retaining
full review 1's callback-0 and successful-no-overlay result-family rationale.
These branches add no different result absent from V30/V31's stated cases.

HARDEST V27 traversal and final caller result, V33 kind/reference/level
selection and V34 sequential capped-add returns: each outcome no change.
Their exact visit/handoff logs, calculation contexts and explicit logical
view updates already distinguish the reported choices. No actual stored
effect is inferred from a fixture update or a normally returning dependency.

The final applicable full run returns MISMATCH 0 with every finding resolved
or explicitly gated, meeting the reader stopping rule for this resulting
contract. Two full callback reviews suffice; no extra run is required merely
to reach a count. This evidence-only record changes no behavior, input,
vector, expected output, acceptance assertion, ownership or gate. The spec
remains draft, with Q01-Q04 and attributable gameplay confirmation open.

## Selected curse-policy continuation, 2026-10-31

New R17/V37/V38 and the supplementary loaded-state assertions extend the
completed callback contract. Its final callback full review 2 is historical
for this expansion; a fresh full reader is required. Complete lifecycle,
AI, combat/resource/overlay effects and gameplay observations retain their
existing gates.

### Loaded-state membership correspondence

On 2026-10-31 the separately owned running profile again reported LoD 1.13c
schema, character level 21 and Act III. It was launched under the recorded
original lod-113c version selection with no requested data override. The
registry Game.exe identity remains the earlier recorded SHA-256; independent
hashes of the running loaded DLLs, the whole archive set and a full override
inventory were not captured. This is narrow observed-profile data, not a
renewed authentication of earlier captures or gameplay confirmation.

A read at Game RVA 0xF829C followed the imported data-global pointer to its
runtime root. At root+0xBC the loaded States pointer was read, at +0xC4 the
signed count 185; pre/post header reads agreed. Each requested identity 9,19,
57,58,60,61,68 selected a complete 60-byte record. Stored unsigned 16 identity
at +0 agreed with its requested index. Definition flag words at +0x10/+0x14
were separately corroborated against the selected membership-mask bits.
Three complete 24-byte masks, sufficient for count 185, were read: curse mask
index 11 at root+0xF8, exp index 30 at +0x144 and life index 32 at +0x14C.
The six words in each mask cover every valid index of this captured collection;
no unreturned suffix is used as evidence of absence.

Narrow results: states 9 amplifydamage,19 weaken,58 lifetap,60 decrepify and 61
lowerresist each have curse=true,exp=false,life=false. State68 evade has
all three false. State57 attract has curse=true,exp=true,life=false. Names
and raw effective States curse/exp/life cells reuse the earlier complete
20920-byte effective file with its original collection date and limits;
loaded identities/mask correspondence are this separate observation. No
whole compiled table identity or state-creation behavior follows.

Common #10033/RVA 0x33B40 is the curse membership consumer reached by Game
0xA00B0 through thunk 0xAE5C. Its matched source was checked for signed
state-count guards and the root+0xF8 word/bit test. The underlying data-export
bit table consumes GDA-R16. No generic state producer or custom-data behavior
is inferred from these seven loaded records.

### Actual local request and curse branch

The existing complete Game 0x50250 and 0xA00B0 readings were reused. Their
relevant instructions were checked against original-image decoding, retaining
the manifest image identities. 0x50250 first uses its current candidate for
0x4E230 primary transformation, then 0x4ECD0 eligibility. The latter skips
monster definition checks for a player, checks target flags 2/3/1 and target
availability, then calls 0xE0A60 with actor/candidate/game. It does not inherit
the action-111 caller's class ranges or earlier opposite relationship test.
The compatible complete relationship/purity graph remains applicable.

After eligibility,0x50250 copies a fresh nine-field request into its local
storage for each call to 0xA00B0. Duration is copied from the prepared effect
request rather than referenced through it. Original 0xA00B0 at RVA 0xA0197
onward reads stat 109; signed resistance>=100 rejects before any duration
write. With nonzero resistance below 100, the call 0x14D0 receives resistance
first and duration second with denominator 100. It stores the wrapping
subtraction back into this local request before checking state 57 at 0xA01C1.
A rejecting state 57 therefore does not undo that local duration write.
The next recipient gets its own copy of the unchanged prepared duration.

With no rejection, a curse request calls Game thunk 0xA7B4 -> Common #10195 /
RVA 0x384F0. The complete wrapper requires an extended base, then calls
Common 0x37F40 with condition 32. That complete search chooses only the primary
head because the secondary selector is absent, strips that selector and
returns the first entry matching the condition; it does not compare state
identity. These two functions have no matched source entries and their
whole listings plus original-image decoding were checked. Noncurse follows
existing state lookup through Common #10871/0x38520 instead.

A stack-position check avoids a false alias: the instruction at Game 0xA01E0
stores [esp+0x1C]=32 after two pushes. It updates the caller-local curse flags
at pre-push offset 0x14; it does not overwrite the saved state-record reference
at pre-push offset 0x1C. The original decoding agrees. The following comparison
and refresh/replacement/creation still consume OWD-R07a/R07b/R08 and their
shared stored-effect gates; R17 stops before claiming that selected list
is a successful application.

V37 independently covers negative/zero/high resistance, post-adjustment
state 57 rejection, noncurse bypass, negative duration and an absent base.
V38 follows the actual action 112 choice 66 with captured bytecode and two
players with resistance 50/25 and active 57. Actual primary transformation,
positive relationship, eligibility and generic curse rejection all execute
from specified records; no successful applier result is supplied. The local
durations 100/150, unchanged prepared 200, two recipient 0 returns and final
action 1 distinguish local copying and ignored callback return. This is a
static valid-record fixture, not a naturally reproduced player state or cast.

### Shared lifecycle and monster-only AI limits

The reviewed STL-R05a/R06d/R06e/R06f handoff was read with its exact direct
control, ordinary list, ancestry, prepared capacity/no-resize and callback
preservation boundaries. It does not close complete Q04; allocator/shrink,
indirect calculations/notifications, extended rebuilding and actual release
remain gated. No shared file was edited and no generic storage authority was
duplicated in this consumer.

Further local monster-only code was located and read: Game 0x500A0 chooses
AI selector 10/12/11 for selected target states 23/27/56, otherwise 0; removal
wrapper 0x50130 requests cancellation and selector 0 restoration before generic
0x9F7A0 removal. Their complete listings have no matched entries; original
bytes were decoded. Game 0x9D230/0x9D0F0 and 0x21420 were read to identify the
installer/reset/command boundary. Compatible MTP-R18 supplies its entry
state 54 assertion/guard facts, and MTP-R22/V21b its exact selector/table limits.
These do not supply arbitrary-class initializers, old alternate handlers or
thinker effects. Those exact AI slices remain Q04. Captured selected 91 has
state 61 and Param7=0, so its actual R12 callback returns before those calls;
no original AI-change effect is inferred from their presence.

The section-3 table's seven callback input rows were also completed with
explicit Origin/Binding cells during this expansion. Their semantic sources,
widths, conditions and values were already present; this corrects table shape
and names their existing runtime/data provenance without changing callback
behavior. The next full review must cover these associations as well.

## Selected curse-policy full review: findings and outcomes, 2026-10-31

The fresh smaller-model full review of the resulting R01-R17/V01-V38
contract returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 7,
COVERAGE 7 advisory groups and HARDEST 3. Each empty defect list has outcome
no change. The seven gated entries refer to existing Q01-Q04, R02 and
V37/V38's stated boundaries; they are not seven additional questions.

Each COVERAGE entry has outcome no change:

- R09 resistance stats 39/41/43/45: V23 already exercises selected-set
  membership with stats 36/37, exclusion with 38, the own-versus-resolved
  distinction and signed division. The other four set members use exactly
  the same reduction and return family, without separate transformations.
- R12 remaining control-0 record/link, mWL, boss, switchai and condition-mask
  failures: retain the earlier full-review triage. These end at callback 0
  without later effects, already exercised by V26's ordered rejection cases;
  V26(b) separately reaches the guard-pass/zero-primary outcome. No exhaustive
  rejection-prefix claim is made.
- R13 stale cache, other room-prefilter paths and static/absent-path
  coordinates: stale-cache clearing belongs to the consumed MTP-R09 and its
  MTP-V06, rather than a new local cache rule. V27-V29 cover the local
  traversal, coordinate failure, room failure and empty-success results.
  The exact prefilter and coordinate-source recipe remain explicit in R13;
  the remaining paths select already-covered admit/reject and distance
  outcomes. This retains the earlier traversal triage without claiming
  exhaustive geometry or coordinate-storage fixtures.
- R14 remaining entry guards, nonpositive current/maximum life and positive
  valid overlay: V31(e/f) already exercise both nonpositive resource failures.
  V31(a-d) already request valid overlay 12 after the replacing request,
  while V30 exercises successful replacement without an overlay. Other entry
  rejections share the callback-0/no-request result. The strict overlay bounds
  and their existing no-overlay result-family rationale remain unchanged.
- R15 remaining guards and other-unit kinds: V35 supplies the return-0/
  no-handoff family. V33(a-d) distinguishes regular-monster owner, hireling,
  player other and nonplayer other calculation choices. Every further
  present nonplayer kind uses the same `calc3` branch for a nonregular owner;
  only monster class 290 enters R16. Other kinds add neither a new local
  calculation choice nor a different normal callback result. Dependency
  effects retain their owners.
- R16 Nightmare/Hell Drain, link/owner combinations and linked second add:
  difficulty selects the supplied unsigned Drain byte before the same local
  arithmetic. V34 covers Drain 50, a present link, sequential capped returns
  and the second link request; V36 covers Drain 0/100, no link, unavailable
  link/owner and nonpositive remaining. Other difficulty cells or availability
  combinations repeat these arithmetic/skip result families. Actual stored
  writes remain gated rather than inferred from V34's explicit view fixture.
- R17 an extended primary chain containing no curse list: the search ends
  with absent, the same continuation result already exercised by V37(h).
  V37(a-c) separately distinguish primary order, matching a curse condition
  rather than state identity, and ignoring the secondary chain. No-list
  selection is not rejection or proof of successful creation; its later
  lifecycle outcome remains Q04.

Each HARDEST entry also has outcome no change. R13/V27-V29 already provide
exact room/coordinate/next-link logs and distinguish recipient returns from
the final action result. R15-R16/V32-V36 already specify combat roles,
formula contexts, sequential requested writes and computed capped returns.
R17/V37-V38 already distinguish local duration copies, resistance before
state 57, rejection versus absent-list continuation and ignored callback
returns. No missing decision was reported for these three groups.

The final applicable full run has MISMATCH 0 and no unresolved GUESS, GAP or
DISCRIM, meeting the reader stopping rule for this contract. This record
changes no rule, input, vector, expected output, acceptance, ownership or
gate; another reader is not required for this evidence-only triage. Earlier
review passes retain their historical scope. The document remains draft,
with full Q01-Q04 and attributable gameplay confirmation open.

## Successful existing-curse refresh continuation, 2026-10-31

R18/V39 extend the preceding reviewed contract; that completed review is
historical for this expansion. The new assertion is an actual existing-list
refresh and produced successful recipient continuation, not an allocated-list
fixture or a supplied successful applier return. Complete creation,
replacement, callback effects and gameplay confirmation remain open.

Compatible complete readings of Game 0xA00B0,0x50250 and 0xB5950 were reused
with their recorded original-image identities and dates. Original decoding
of the reached continuation was checked again. Game 0xA01FA onward compares
selected list state, skill and level through thunks 0xAEAA,0xAE62,0xAEBC:
Common #11132/RVA 0x38040, #10555/0x37D50, #10515/0x37D30. The complete matched
source of all three accessors was read before the new image decoding; the
consumer establishes the semantic field names, not their analyst labels.

For equality, Game 0xA021F loads the local request duration. Nonzero duration
adds it to the current game frame with wrapping at 0xA0230 and calls Game
0xA778 -> Common #10161/RVA 0x38020 at 0xA0234. The complete matched setter
was read first and its original instructions corroborated: non-null list
gets the expiry store, positive expiry sets timed, nonpositive does not clear
timed. This consumes STL-R06a, with no allocation, attachment or stat-value
propagation required on this branch. Game 0xA0248 then calls 0x8EA80 with the
same game/target, type 12, computed expiry, zero override and zero parameters.
STL-R01 owns the actual schedule in its valid frame domain; malformed/wrapped
indexing and later dispatch/removal retain their existing gates. Duration 0
skips both calls. Game 0xA024E returns the same existing list. No owner,
magnitude, callback, state/skill/level or chain store intervenes.

The actual recipient stores the returned list at Game 0x50353. For fixed
indices 0/1, original 0xB5950 initializes monster-only and auxiliary controls
to zero, all five additional stat identities to -1 and event-function
descriptors to zero. Thus 0x50363 skips monster-only AI, all five iterations
at 0x50387 skip additional-stat transforms/writes, 0x503D3 skips the auxiliary
call, and 0x503F6 skips item-event cancellation/registration. Game 0x50459
produces integer 1. These are code-produced branches; no successful list
processor or writer result is supplied. R13 then normally completes with 1
and R05 forwards it. No callback body or later expiry is executed by this case.

V39(a) follows actual selected 66 through radius 3, prepared duration 200,
resistance 50 and local duration 100 at frame 1000. The same existing list
gets expiry 1100, shortening 1500; its magnitude -73, prior source owner and
callback remain. V39(b) uses caller/existing level 0: captured expressions
give radius/duration 0, a candidate exactly at the center passes the inclusive
distance predicate, and refresh returns the list without an expiry/event
write. Both cases derive recipient/action 1. The fixed request's -100 does
not replace the existing -73. These supplied valid records do not establish
how the old curse was originally installed or that a natural cast occurred.

This branch does not need the shared prepared direct-stat definition,
ancestry/capacity slices, because no replacing stat write, attach or unlink is
executed. Those conditions are not waived for creation/replacement or the
callback life/combat paths. Full Q01-Q04 remain, and the expanded contract
requires a fresh smaller-model full review.

### Additional natural travel attempt

On 2026-10-31, a natural Act III forest encounter ended in a Jungle Hunter death before an eligible fetish shaman was reached. Bounded entry observations at Game 0xB39F0 and 0xB5950 each allowed 64 records over a 600-second window; each retained zero records, zero drops/evictions and zero in-flight calls. This is an incomplete encounter, not evidence that either handler is unused or that any gameplay vector matched.

## Existing-refresh full review and selection-vector correction, 2026-11-01

Fresh smaller-model FULL review through R18/V39 returned MISMATCH 0,
GUESS 0, GAP 0, DISCRIM 1, GATED 9, COVERAGE 1 and HARDEST 3.
All V01-V39 numeric results were independently recomputed without a mismatch;
the three empty defect lists have outcome no change. The nine gated entries
are the existing Q01-Q04/R02/O01 and V37-V39 boundary limits, not new gaps.

DISCRIM V39's original L-first chain could not reject an incorrect second
state-list lookup after curse selection: changed V39(a/b)'s fixture and
assertions. Its primary chain is now N,L,L105, with N and L both state 9 but
only L marked as a curse list. N's expiry is 1700, magnitude -31, owner
monster/id 8 and callback J; these remain unchanged. RLI-R03 would select N
by state, whereas R17 selects L by condition. These are independently supplied
list conditions/state metadata under RLI section 3 and STL-R06a, not a
claim that this duplicate-state configuration was naturally produced. The
original Game 0xA01E8 curse-search result jumps directly to 0xA01F6 for
state/skill/level comparisons; it does not enter the noncurse state lookup at
0xA01F1. This reuses the complete original-image/source record rather than
renewing its verification date.

Case(a) still writes only L.expiry=1100, leaving N.expiry=1700; the incorrect
second lookup instead writes N.expiry=1100 and leaves L.expiry=1500.
Case(b) still returns L without an expiry or event write; the incorrect
lookup returns N. The exact returned-list and unchanged-N assertions now
reject that alternative in both cases. All original arithmetic, scheduling
and normal recipient/action results remain unchanged.

COVERAGE nonzero duration producing a nonpositive wrapped expiry: no change.
R18 explicitly retains STL-Q10's malformed/wrapped scheduling boundary;
another fixture does not establish that gated schedule or its continuation.
HARDEST R13/V27-V29 traversal/final result, R15-R16/V32-V36 roles and
sequential requests, and R17-R18/V37-V39 selected-list/local-duration/refresh
versus final result: each outcome no change beyond the selection fixture
correction. Their exact logs, roles, returned identities and metadata scope
already state the reported decisions. No shared ownership or gate changed.

Only V39(a/b)'s fixture and discriminating assertions changed. A fresh
focused delta reader must recompute those cases and check the second-lookup
finding; the preceding full review does not complete review of that delta.
Full allocator, creation/replacement, notification/release, callback and
observation limits remain, with document status draft.

## Existing-refresh selection delta: final applicable review, 2026-11-01

A fresh smaller-model focused delta review of V39(a/b), in the full
R17/R18 and cited dependency context, returned MISMATCH 0, GUESS 0,
GAP 0 and DISCRIM 0. Each empty list has outcome no change. Both cases
were independently recomputed: the curse search returns L, case(a) changes
only L.expiry to 1100 and retains N.expiry=1700, and case(b) leaves both
expiries unchanged while returning L. Independent supplied list conditions
and state identities are valid in the cited contracts; no natural producer
is inferred. A second state lookup would select N and fails the returned-L
assertion in(b), and the only-L expiry/unchanged-N assertions in(a).

This delta resolves the preceding full review's one DISCRIM entry. Its
MISMATCH-0 full review and this clean final delta meet the reader stopping
rule for the resulting contract. No further behavior, vector, input,
acceptance, ownership or gate change follows this evidence record. Full
Q01-Q04 and the shared allocator, creation/replacement, notification,
release, callback and gameplay-observation limits remain; status is draft.

## Callback life-write integration through the shared direct domain, 2026-11-01

R19/V40/V41 extend the completed R18/V39 contract. Its final clean delta
remains historical for this new integration; the resulting contract needs
a fresh full reader. No shared file or stat-storage authority was changed.

### Actual caller and dependency correspondence

The complete compatible readings and original-image records of Game 0x4DF20,
0x4E760 and 0x4F090 were reused with their original dates and scope. The
selector-5 callback calls Game 0xA472 at 0x4E8AC with other, stat 6, its computed
value and layer 0. Each class-290 capped-add calls the same thunk at 0x4DF52
with that add's recipient, stat 6, requested value and layer 0. Original
import/export resolution confirms Common ordinal 10887/RVA 0x3A740.

Common 0x3A740 and 0x38550 have no matched-source entries. Their complete
instruction listings were read after that check, then decoded against the
manifest-matched original image. The 64-byte unit wrapper loads the unit's
base list and forwards stat/value/layer to 0x3A280. After a changed list-write,
a player alone reaches extra changed-key collection at 0x38550; its stat-6
selector byte at 0x38630 is zero and selects immediate return 0x38625.
Neither caller uses a numeric return from this unit-facing writer. The
capped-add restores its own computed amount as its result; selector 5
continues to its overlay guard and callback result.

The already verified MRH-R08 explicitly permits this same isolated unit
writer when reached from another owned skill path. R19 consumes that
contract, including its extended root base, independent own/resolved arrays,
nonzero request/result, direct definitions and prepared-storage conditions.
STL-R06e/STL-R06d remain the owners of actual own replacement and resolved
delta propagation. No new generic writer recipe is introduced here.

### Compatible life-definition evidence reused

The dated 2026-10-30 original-profile record in
[monster healing evidence](skill-monster-resurrect-and-heal.md#life-definition-and-direct-propagation-evidence)
is compatible with this consumer. Its effective ItemStatCost loader read
requested offset 0/length 64000 and returned 40109/total 40109 bytes; header
parsing selected `hitpoints`, explicit ID 6. A separate loaded check read
count 359 and 100 bytes of record 6 with stride 0x144, confirming stored ID 6.
Definition bytes +4/+5 were 0x10 and +0x50 through +0x53 were zero; six
bit-mask words were 1,2,4,8,16,32. Under the shared consumer mapping these
select filteredStat=false, notifyResolvedChange=false, collectChangedKey=true,
retainResolvedZero=false and zero dependent expansion/formula/marker.
The same narrow record has fMin=false, MinAccr=0 and ValShift=8.

These retain their original collection date and limits: reported vanilla
LoD 1.13c, no data override requested; effective-file hash, loaded-module
hashes and physical archive inventory were not captured. Original on-disk
Common identity matches the source manifest separately. No new table or
gameplay observation is claimed by reusing this record. The supplied stat-7
read definitions in V40/V41 are explicit vector inputs, not new archive claims.

### Resulting consumer assertions and limits

V40 supplies existing nonzero life entries in both arrays, each count 2 and
capacity 8, with an associated root base and no parent/secondary condition.
All MRH-R08 direct-definition/no-resize conditions are stated. It executes
the actual writer rather than V34's explicit view-update fixture.
The first Q request 1000 changes own 900 to 1000 and resolved 950 to 1050.
The O request 1000 changes own 900 to 1000 and resolved 990 to 1090. The
next Q read therefore sees 1050. Its request 1000 equals own 1000 and makes
no change, while the capped-add computes return -50 from 1000-1050.
The final return is ignored. Earlier returns 50 and 10 leave remaining 40
before that last add. Overlay operations remain separate requests under
their owner, not proof of later stored effects.

V41 follows the actual captured selected-87 empty calculation through the
selector-5 callback. Its percentage and added amount are zero, but current
life 1024 exceeds maximum 768, so the local operation requests replacing
other's life with 768. Own 512 becomes 768; the delta 256 changes resolved
1024 to 1280 through the actual direct writer. No second maximum clamp
belongs to that writer. Negative prgoverlay omits the overlay operation and
the callback produces 1. State/event installation remains an input boundary;
this does not infer a naturally produced initial over-maximum configuration.

Both cases preserve all unrelated entries, backing stores, counts, chain
relations and changed-key collections in their stated domain; neither
invokes stat callbacks or supplies a writer result/storage effect. They
close actual prepared life writes and subsequent reads, not arbitrary
storage, allocation/deletion/failure, dependent formulas, notifications,
release, original timed-state installation, earlier combat or later overlay
effects. Q01-Q04 and document status draft remain.

### Full reader outcome and coverage triage, 2026-11-01

The fresh smaller-model FULL reader independently recomputed all V01-V41
and returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 8,
COVERAGE 6 and HARDEST 3. Each empty defect list has outcome no change.
The eight GATED entries are existing R02, Q01-Q04, O01, V21 and V25
limits; none is a new question or a confirmation of gameplay.

Each COVERAGE advisory has the following outcome:

- R14 early callback guards: added V42. V31(e/f) reject only after
  calculation and both resource reads; they do not cover an early rejection
  with an empty calculation/resource log. V42 independently rejects absent
  or wrong-kind owner/other, unavailable other, absent/nonpositive incoming,
  absent skill, out-of-range/inactive state and missing list before source
  lookup or calculation. Its supplied membership/list configuration makes
  no natural-producer claim.
- R15 guards and other-kind selection: no change. V35 already executes
  physical 0 and -1 with result 0 and no skill access, calculation,
  construction or handoff. The other entry guards have that same result
  and no-handoff outcome; this is an equivalence of the rejected continuation,
  not a claim V35 executes each guard. V33(c/f) executes the nonplayer
  `calc3` branch with a monster other, while V33(b/d) executes player
  `calc2` and V33(a) the overriding regular-monster-owner `calc1` choice.
  Present object/item/missile/tile others take the same `calc3` and
  non-class-290 continuation; those individual kinds are not separately
  exercised. No new result family is missing and combat effects remain Q04.
- R16 absent MonStats: no change. V36(a) explicitly returns 1 before
  alignment/link with only Drain=0 in its named checkpoint log. An absent
  MonStats record takes that same no-alignment/no-link/no-effect frontier,
  without a Drain read; the absent-record guard itself is not claimed as
  executed by V36. The rule reads a present record's difficulty-selected
  unsigned Drain byte, not a separate optional Drain value. V34/V40 execute
  the nonzero-Drain linked path; V36 covers the distinct later skip families.
- R17 transformed primary and eligibility: added V43 for actual fixed-index-0
  flag rejection before application, contrasting V38's application followed
  by policy rejection. Both recipient results are 0 but the operation logs
  distinguish the frontiers. No new transformed-zero case is added: the
  fixed primaries of indices 0/1 are -100/-50, which remain nonzero after
  R09's eligible divide by five, and indices 2/3 have no primary. V23's
  supplied -1 establishes the transformer's zero result in isolation;
  R12/V26/V27 establish the actual fixed-index-4 zero rejection and final
  traversal result. Neither is credited as an execution of R17's zero arm
  for fixed indices 0-3.
- R18 mismatched state/skill/level: no change. Replacement/creation remains
  Q04, not an expected successful refresh or a missing invented result.
  V39 establishes only the exact equal-metadata refresh. Nonpositive wrapped
  scheduling remains the named STL-Q10 domain.
- R19 absent entries: no change. MRH-R08 and STL-R06d/R06e own the exact
  prepared insertion domain, including the owning MRH-V13(a-c)
  discriminators. V40/V41 exercise existing own/resolved entries and do not
  claim absent insertion. Their consumer contract already cites the owner;
  duplicating its insertion machinery here would add no local result.
  Resize, failure, indirect/extended rebuilding and release stay gated.

All three HARDEST entries have outcome no change. V40 states the distinct
local capped-add returns 50,10,-50, actual own/resolved deltas 100,100,0,
and the ignored final negative return. R13/V27 explicitly orders room
prefilter, distance, callback, then next-link read and the final traversal
result. R17-R18/V39 explicitly uses each fresh local duration, first primary
curse selection and the selected list's refresh without a second state
lookup. These assertions and their existing discriminators remain intact.

Only V42/V43 and their Id-map associations extend the reviewed material;
rules, semantic inputs, outputs, ownership and gates are unchanged. A fresh
focused reader must check those additions before the resulting contract
meets the stopping rule. Earlier FULL results are retained with their exact
V01-V41 scope. No observation or status promotion follows this review.

### Focused V42/V43 reader outcome, 2026-11-01

A fresh smaller-model focused DELTA in the complete rule/dependency context
returned MISMATCH 0, GUESS 0, GAP 0 and DISCRIM 0 for V42/V43. Each empty
list has outcome no change. All thirteen V42 guards independently produce
callback 0 with no source lookup, calculation, resource read, replacing write,
overlay or mutation. Case(m)'s active membership and absent matching list
are distinct supplied facts on readable storage, not a natural-producer claim.

V43 independently recomputes seed 1791398085, remainder 0, selected skill 66,
radius 3 then duration 200 and each player's unchanged primary -100. T's
false flag 1 and U's false flag 2 reject before relationship or application;
actual recipient results 0/0 are ignored and final action result is 1.
No candidate-local duration or lifecycle operation occurs.

The preceding MISMATCH-0 FULL review, its complete advisory/HARDEST triage
and this clean focused DELTA meet the reader stopping rule for the resulting
contract. No rule, input, vector, output, acceptance assertion, ownership or
gate change follows this record. Public metadata updates do not renew any
source or gameplay observation. Q01-Q04 and status draft remain unchanged.

## Actual aura collection through prepared storage, 2026-11-01

R20/V44-V46 extend the reviewed R01-R19/V01-V43 contract and require a
fresh full reader. The prior clean FULL and DELTA records retain their
recorded scopes. This continuation establishes preparation and the actual
fixed collector's output writes at a supplied initialized-storage boundary;
it does not substitute successful allocation/release results for their
unclosed callees.

### Reuse, identity and methods

The 2026-10-30 Game 0xB39F0/0x9E570 readings, dispatch identification,
Common 0x12030/0x120F0/0x12250 and 0x3BA20/0x3BA60 readings were checked
for compatible artifact identity and coverage. Their original dates and
limits remain. The remaining question was the previously uncredited exact
fixed-filter room ordering and candidate-array production, rather than
whether a collection request was made.

The original D2Game and D2Common images were freshly hashed on 2026-11-01
and agree with the manifest identities and earlier fingerprints quoted
above. Original-image instruction decoding corroborates full Common ranges
0x12720..0x1288F, 0x121D0..0x1224F, 0x124E0..0x1271F,
0x2FE10..0x2FE6F, 0x3C390..0x3C3AF and 0x2D320..0x2D3BF,
and Game 0x9E570..0x9E5FF. Availability was checked before their new
interpretation: these functions have no C view verified byte-for-byte. C verified byte-for-byte
was read for Common 0x12030, 0x12250, 0x3BA20, 0x3BA60 and cleanup
0x12000. Common 0x120F0's matched record is an instruction transcription,
not a C reconstruction; it was read as such. Original read dates are not
renewed merely by their reuse. Private source details and raw decoding
records remain separately preserved.

### Original references mapped to R20

| Artifact / original function | Role and reached boundary |
|---|---|
| D2Game 0xB39F0, action slot 111 | Calls coordinate helper 0x9FA30; failed coordinates return before latch; success sets latch and computes invoking Param4+2*level with i32 wrap before 0x9E570 |
| D2Game 0x9E570 | Requests context storage first, then constructs fixed descriptor, obtains actor room and initializes/enumerates collection |
| D2Game 0xA478 -> D2Common ordinal 10331 / 0x2FE10 | Actor current-room lookup after context allocation; player/monster/missile path may be absent, static kinds require readable position storage |
| D2Game 0xA502 -> D2Common ordinal 10591 / 0x12030 | Initializes collection inputs and capacity 15; actual backing allocation remains Q03 |
| D2Game 0xA4F0 -> D2Common ordinal 10052 / 0x12720 | Actual ordered room traversal, saved next links, accepted candidate writes, count and growth-request frontier |
| D2Common 0x12250 -> 0x3BA60 | Strict containment selects current room alone, otherwise 0x3BA20 supplies the stored ordered room array/count |
| D2Common 0x121D0 -> 0x3BA60 | Exact signed wrapping prefilter, equivalent to the R13 subrecipe; no conventional circle/rectangle substitution |
| D2Common 0x124E0 | Fixed descriptor's actual X-then-Y geometry, kind/mode, player identity, target flags and candidate-room predicate; increments accepted counter |
| D2Common 0x2FE10 -> 0x3C390 -> 0x2D320 | Candidate room and fixed level exclusion; absent candidate room does not reject |
| D2Common 0x12030 -> 0x09222 -> Fog ordinal 10045 / 0x1CD50 | Initial candidate backing allocation; its supplying result remains Q03 |
| D2Common 0x120F0 -> 0x09252 -> Fog ordinal 10047 / 0x1CDA0 | On fifteenth acceptance, requests storage growth after storing the candidate; continuation through allocator remains Q03 |
| D2Game 0xA4FC -> D2Common ordinal 10344 / 0x12000 | Cleanup requests backing release; clears backing reference only after normal return; actual release remains unclosed |
| D2Common 0x12000 -> 0x09228 -> Fog ordinal 10046 / 0x1CCC0 | Actual backing-release dependency boundary, not a supplied successful release |
| D2Game 0xA2E6 / 0xA2EC -> Fog ordinals 10045 / 10046, RVAs 0x1CD50 / 0x1CCC0 | Initial context allocation and later context release; request boundaries, not established successful storage/release effects |

The original manifest-listed artifacts are the public references. Function
descriptions are analyst labels, not original exported names. The allocator
investigation is compatible with the shared STL-Q07 limit; no new generic
allocator contract is established here.

### Fixed descriptor, room and output correspondence

Game 0x9E570 clears its local descriptor, then installs filter 0x583,
actor, center and extent. It passes no custom collection predicate.
Common 0x12030 initializes the separate header's predicate to null, its
room-level filter to zero, its count to zero and capacity to fifteen.
Thus the collector's optional callback validation and room-level town test
are unreachable on this fixed route. Descriptor 0x40 limit, 0x200 collision
and 0x800 custom-test branches are also inactive.

Common 0x12250 performs four strict signed comparisons with wrapping
subtraction/addition. When all hold, 0x12720 treats the header's current-room
reference as a one-element sequence. Otherwise it copies the current room's
stored room-array reference and unsigned count through 0x3BA20. The loop
does not prepend current room or deduplicate repeated entries. Each admitted
room uses 0x121D0's checked unusual prefilter. At 0x12805 the current unit's
next link is saved before the actual fixed predicate is called at 0x12827.
This differs from R13's callback-before-next-link route.

Common 0x124E0 reads dynamic integer X at path+2, then Y at path+6;
static X/Y are read in that same semantic order. Geometry uses wrapping
squares and a signed inclusive comparison before kind rejection. The fixed
flags enable players and monsters only, reject their respective modes
0/17 and 0/12, and require unit flag bits 2 then 3. Player self-exclusion
is at 0x125AA. The monster route, with descriptor bit 4 clear, jumps
directly from 0x125E4 to 0x1262B without that identity comparison. A monster
actor may therefore enter its own collection at this stage. No local
alive-exclusion-condition, hireling or relationship test is added.

The candidate's current-room check is distinct from the inactive room-level
town filter. Common 0x3C390 returns zero for absent room; a present room
reaches 0x2D320. Its original-image two-way jump table returns 1 exactly
for level identities 1,40,75,103,109 and 0 for every other signed identity.
The table's bound and selectors were checked as data, not decoded as
instructions. These identities are code bindings, not archive-value claims.

On acceptance, 0x124E0 increments descriptor counter before returning 1.
Common 0x12833 stores the candidate at the current accepted index, increments
the local count at 0x12839 and compares to capacity. Equality calls growth
0x120F0 at 0x12840. That helper computes capacity+15 and requests allocation
before writing a new capacity/backing reference. With initial capacity 15,
fourteen accepted units avoid growth; a fifteenth is stored before the
request, even if it is the last unit. The final header-count store at
0x12862 is reached only after normal traversal. V46(b) stops before the
growth operation returns and therefore asserts neither a new capacity nor
a final collector result. No successful allocation fixture supplies either.

### Checks, open boundary and review

V44 independently traces the actual coordinate/latch/extent preparation up
to initial allocation, including zero coordinates, level zero, negative
extent and wrapped multiplication. V45's supplied initialized context
executes actual room/filter decisions and array/count writes. Its strict
inside case visits only current room; the equal-edge case uses the stored
sequence and preserves its duplicate candidate. Distinct cases preserve
monster actor inclusion versus player actor exclusion and require the exact
X-then-Y coordinate log. V46 distinguishes fourteen no-growth writes/return
from the fifteenth write followed by a gated growth request.

These cases establish no natural producer for the initial initialized
context. The actual header/backing allocation requests, growth and both
cleanup releases remain Q03. Shared timed-state and later aura effects,
the applicable engine synchronization invariant, source-profile production
outside the captured cells and attributable gameplay observations remain
unchanged gates. Complete action return/effect parity is not inferred from
a helper's returned count. No gameplay action or observation was performed
for this source continuation. Status remains draft; the new behavior,
inputs and vectors require a fresh FULL reader before review completion.

### Full R20 reader outcome and advisory triage, 2026-11-01

The fresh smaller-model FULL reader independently recomputed V01-V46 and
returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 5 groups,
COVERAGE 3 and HARDEST 3. Each empty defect list has outcome no change.
The five gated groups are existing R02 and Q01-Q04; O01's pending
observation and the vectors' stated stage boundaries remain explicit limits,
not newly discovered behavior or a promotion of the document.

All three COVERAGE advisories are resolved with one added vector V47:

- Absent dynamic candidate path: V47(a) executes the collector's actual
  local zero-coordinate continuation. MTP-R09 owns action target-coordinate
  production; MTP-V06 instead demonstrates stale-cache clearing and stored
  point(10,0), not this collector's absent candidate path. NRF-R10b's direct
  getter is fatal on null and is not invoked on this local null-path route.
  V44(d) already executes action coordinate result 0 before setup, but is
  not credited as candidate-collector coverage.
- Player modes 0/17, monster mode 0, flag 3 and absent candidate room:
  V47(a) exercises each exactly. The first four conditions reject their
  named candidates; **absent candidate room does not reject** and is
  deliberately accepted for N. The reader's grouping of every entry as a
  rejection is not adopted. V45's monster mode 12 and flag-2 failures
  remain their actual cases, not claims that every guard was exercised.
- Negative extent and wrapped geometry: V44(b/c) already demonstrates
  negative/wrapped preparation, but no actual collector geometry. V47(b)
  executes extent -5 with inclusive squared distance 25. V47(a/c) execute
  negative wrapped squared sums; V47(c) additionally executes wrapped
  containment subtraction and candidate-minus-center subtraction. Their
  supplied readable world records do not assert a natural coordinate or
  initial-storage producer, a successful allocator or gameplay confirmation.

All three HARDEST findings have outcome no change to rules. R20/V44
explicitly order coordinates, latch and extent before initial allocation,
then actor-room lookup only after that request. V45 retains its exact
current-room versus stored-array choice, duplicate entries and X-then-Y
read log. V46 retains the fifteenth stored candidate and accepted counter
before growth, with capacity 15/output count 0 at the gated frontier rather
than an inferred successful allocation or final result.

Only V47, its Id-map reference and its acceptance association extend this
FULL-reviewed material. Rules, semantic inputs/outputs, ownership and gates
are unchanged. A fresh focused DELTA of V47 is required before the resulting
contract meets the reader stopping rule. No further new rule growth occurs
before that result is triaged. Q01-Q04 and status draft remain.

### Focused V47 reader outcome, 2026-11-01

A fresh smaller-model focused DELTA independently recomputed V47(a-c) in
the complete rule/dependency context and returned MISMATCH 0, GUESS 0,
GAP 0 and DISCRIM 0. Each empty list has outcome no change. All three
strict-containment choices and R13 room prefilters admit current R0.

Case(a)'s absent candidate path produces X/Y 0 locally, deltas -65535,
wrapped squares -131071 and sum -262142, which passes signed comparison
to 25. N's absent candidate room does not reject; F's false flag 3 and
the player 0/17 and monster 0 modes reject their named candidates. Case(b)
executes extent -5 and inclusive distance 25. Case(c)'s wrapped lower
2147483643, upper -2147483643 and end -2147483608 select current room
despite the empty stored array; candidate delta -2147418113 and wrapped
sum -262142 pass. Actual arrays/counts/returns remain [N]/1, [M]/1, [T]/1.
The reader also checked saved next links before predicates, X-then-Y
coordinate results, exact logs, capacity preservation and no draw/effect.

The MISMATCH-0 FULL review through V46, complete advisory/HARDEST triage
and this clean final focused DELTA meet the reader stopping rule for the
resulting R01-R20/V01-V47 contract. No rule, input, vector, output,
acceptance assertion, ownership or gate change follows this record.
Q01-Q04, pending observation, supplied-storage limits and status draft
remain unchanged; review completion is not user approval or gameplay proof.

## Independent noncurse aura refresh and local return, 2026-11-01

R11's local optional-list return and V48 extend the completed collector
contract. R17 already explicitly admits the independent R11 request boundary
for its timed-state policy. R18's first paragraph now names that existing
noncurse association; its fixed-index-0/1 curse-recipient success suffix does
not apply to V48. The new local return/input association requires a fresh
FULL reader, not a delta or an inference from the previous clean review.

### Compatible source reuse and return correspondence

The complete original Game 0xA09F0 and 0xA00B0 readings were reused with
their recorded 2026-10-30 and 2026-10-31 dates and limits. Original-profile
Skills 177 Param1/2/3 and state 68 noncurse correspondence retain their dated
narrow records; no fresh archive or gameplay observation is claimed.

The unestablished assertion was the local helper's returned result, not its
previously verified arithmetic/request shape. After confirming no verified
C view exists for 0xA09F0, the complete listing was checked and original
image 0xA09F0..0xA0ACF decoded on 2026-11-01. The original D2Game image
continues to match the source-manifest fingerprint. At 0xA0ABE the helper
passes its actual prepared request to Game 0xA00B0. Instructions
0xA0AC3..0xA0AC8 discard local request storage and return without changing
the returned list reference. Its action-111 caller at 0xB3ADA then reads
collection count for the next iteration without consuming that local result.
That caller observation establishes ignoring this local result, not a new
normal action result across unresolved cleanup.

The actual aura caller supplies no custom removal callback. Thus the
optional callback-address validation branch in 0xA09F0 is bypassed. The
new-list branch in 0xA00B0 would select its default callback when the request
has no custom callback. The existing-list refresh does not enter that
creation branch or overwrite its supplied callback. V48 asserts retention
of K, not conversion of the request's default choice into a callback write.

### Actual noncurse refresh path and shared contracts

Game 0xA00B0 consumes OWD-R06a's target/state eligibility; V48's player
target skips its monster-only checks. Its named state 68 curse membership is
false, so it skips resistance and state 57 even though both supplied values
would reject a curse request. At 0xA01EF..0xA01F5 it performs the requested
state-list lookup through the already resolved RLI-R03 dependency, rather
than the first-curse search used by a curse request. The earlier primary
curse list C is not selected; L is the state 68 list.

The compatible metadata getters and refresh sequence recorded for R18 are
reused without enlarging the shared owner contracts. Equal state, skill and
level reach 0xA021F; nonzero duration stores frame+duration through the
expiry setter and schedules type 12 through Game 0x8EA80/STL-R01, then
returns L at 0xA024D..0xA0256. Duration0 skips both operations. A lower
requested signed level takes 0xA0257..0xA0262 to null, without detach,
creation or an expiry write, as OWD-R07a states. These are actual reached
callee outputs, not supplied application success, requested magnitude writes
or allocator results.

V48(a) computes duration 200 and requested value 25, stores expiry 1200 and
one pending event, while retaining L's own magnitude -73, old source owner
and callback. V48(b)'s level 0 computes duration 0/value 5, returns L and keeps
expiry 1500 with empty queues. V48(c)'s level 3 computes duration 150/value 20,
but existing level 4 rejects with null and no write. The helper forwards
these produced results L/L/null without converting them to boolean success.
All preexisting state/list/stat records and seeds otherwise remain unchanged.

This independent request does not assert passing the original action's two
opposite relationship tests, natural recipient production, a curse-recipient
integer result or any complete action result. It establishes no original
state installation, allocation, replacement, later timer firing/removal,
cleanup/release or wider aura effect. Q01-Q04 and status draft remain.
The new contract and all vectors require fresh full blind review.

### Full independent-refresh reader and triage, 2026-11-01

The fresh smaller-model FULL review independently recomputed V01-V48.
Its original report returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1,
GATED 4, COVERAGE 8 and HARDEST 3. The empty MISMATCH/GUESS/GAP lists
have outcome no change. GATED Q01-Q04 retain their exact existing profile,
observation, aura and curse boundaries.

The sole reported DISCRIM concerned whether V48 tests the action caller
ignoring the local helper return. The writer checked V48's explicit helper
boundary and requested classification clarification without editing the
contract. The same reader confirmed that no V48 assertion claims to test
that outer caller; its local result does discriminate optional-reference
forwarding from booleanization. It reclassified the entry as advisory
COVERAGE of R11's separate caller-ignore clause. Both reports are retained;
the corrected lists are DISCRIM 0 and COVERAGE 9, not a second independent
recomputation. No blocking discriminator remains.

All nine advisory outcomes are recorded:

- R11 caller-ignore clause: no change. V48's independent helper boundary
  explicitly excludes the action result and natural R10 eligibility. The
  source return/use relationship is established, but a complete action-111
  caller result across collection storage/cleanup and natural eligibility
  remains Q03. No fabricated successful allocation, helper result or
  eligible natural producer is added to claim that this stage demonstrates
  the outer caller. This coverage remains documented.
- R04 negative selected identity: no change. R01's five code-fixed outputs
  cannot be negative. A negative identity at the isolated supplied-validation
  boundary takes the same result-0/no-record-access frontier as V16's
  at-count rejection; V16 is not claimed as an execution of the negative arm.
- R07 extreme positive levels: no change. The positive-level branch is
  exercised in V21; wrapping linear-parameter arithmetic is consumed from
  SCX-R21 and the numeric foundation, not a second local formula. SCX-V03
  discriminates wrapping multiplication in the expression arithmetic owner.
  It is not claimed to execute every captured linear expression at an
  extreme level. No distinct local branch or missing supplied input is added.
- R09 stat identities 39/41/43/45: no change. They select the same divide-by-five
  operation and result family as 36/37 in V23, with 38 discriminating exclusion.
  Those four individual identities are not separately exercised.
- R13 successful recipient followed by continuation: added V49. Two actual
  equal-metadata refreshes produce recipient 1/1, the first return is followed
  by reading the next link and applying the second, and final action result
  is 1. This distinguishes stopping only on success, which V38's two rejected
  recipients do not. The same advisory's static-coordinate kinds: no change;
  R08 rejects every static kind, so with readable supplied static storage
  they add no accepted-recipient result family. Their coordinate branch is
  explicit but not claimed exercised by the dynamic R13 fixtures.
- R17 fixed indices 1-3: no change. V09-V11 establish their distinct prepared
  shapes, while V32-V33 establish the registered callbacks independently.
  Actual creation/replacement, event registration and their full stored
  effects remain Q04. No successful applier fixture is introduced to bypass
  that gate; the complete flows remain unexercised.
- R18 fixed index 1 success suffix: no change. The verified suffix is shared
  with index 0 after a non-null refresh: monsterOnly is false and there are
  no additional stats, auxiliary operation or event registration for either
  fixed row. V39/V49 exercise that result family on index 0; V09 separately
  establishes row 1's identity/request. No claim is made that index 1's full
  successful path has been observed or separately executed by a fixture.
- R20 allocation/growth completion/cleanup: no change. All remain Q03;
  V44/V46 stop at the declared requests rather than supply storage success.
- R20 additional room-prefilter arms: no change. R20 consumes R13's exact
  prefilter subrecipe, with its established admit/reject family. V45-V47
  exercise this collector's distinct room-source, candidate and wrapping
  outcomes; they do not claim exhaustive combinations of the shared
  prefilter's comparisons.

HARDEST 3 each have outcome no change. R13/V27-V29 retain exact target,
room, distance, callback and next-link order; V49 adds the actual successful
continuation discriminator. R20/V45-V47 retain strict containment, stored
array fallback, X-then-Y filtering and explicit growth frontier. R16/R19/V40
retain the distinct capped-add returns versus own/resolved writes and
subsequent current-life reads. No rule or dependency behavior changed in
this triage.

Only V49 and its Id-map/acceptance references extend the FULL-reviewed
contract. A fresh focused DELTA must recompute that added vector and its
continuation finding before the stopping rule is met. Full Q01-Q04 and
document status draft remain; no gameplay action was performed.

### Focused successful-traversal reader outcome, 2026-11-01

A fresh smaller-model focused DELTA independently recomputed V49 from
its inherited V38 inputs, current rules and named dependency contracts.
It returned MISMATCH 0, GUESS 0 and GAP 0. The initial report labeled
three rejected wrong alternatives as `DISCRIM (3 validated)`. On explicit
classification clarification, the same reader confirmed that none still
passes V49: these are validated discriminators, not blocking findings.
The corrected DISCRIM count is 0. Both reports and their exact scopes are
retained; this correction is not another independent recomputation.

All four empty defect lists have outcome no change. The reader recomputed
selection 66/seed 1791398085, radius 3 before duration 200, squared distances 1/4
within 9, fresh per-candidate durations 100/150, actual selected lists L/D,
expiries 1100/1150 and their type 12 events. Actual recipient results 1/1
continue through the next-link reads to final action 1. No supplied successful
callback, applier or allocator result is used.

The three validated discriminators remain intact: stopping after T's first
success omits U's effects/visit; leaking T's adjusted duration makes U's
duration 75/expiry 1075 instead of 150/1150; overwriting requested magnitude
violates retained own values -73/-55. No edit was needed for these checks.

The MISMATCH-0 FULL review through V48, its corrected classification and
complete nine-advisory/three-HARDEST triage, and this clean final focused
DELTA meet the reader stopping rule for R01-R20/V01-V49. No rule, input,
vector, output, acceptance assertion, ownership or gate change follows this
record. Full Q01-Q04, original installation, allocation/growth/release,
later lifecycle, natural aura-recipient and observation limits remain.
Status remains draft; review completion is not user approval or gameplay proof.

## Existing curse refresh through actual event registration, 2026-11-02

R21/V50/V51 extend the reviewed R01-R20/V01-V49 contract and require a
fresh FULL reader. The remaining question was the actual recipient's event
suffix and the origin of the later registered callback arguments, rather
than another reading of the already verified life arithmetic or generic
event machinery.

### Source reuse and exact caller correspondence

The complete compatible Game 0x50250 reading and original-image evidence
were reused with their recorded dates. The shared IEP-R01/R02/R05/R06/R07
contracts and their 2026-10-12 through 2026-10-15 evidence were checked for
the same manifest-listed D2Game artifact, event fields and logical collection
boundary. In particular the IEP companion explicitly excludes allocator
internals from its semantic entry-creation/removal contract. Consuming that
verified contract does not reverify a heap allocator or supply a successful
registration fixture.

Availability was checked for Game 0x50250, 0x37110 and 0x37190: none has a
C view verified byte-for-byte. Their complete listings were read before
original-image decoding of 0x50250..0x5046F, 0x37110..0x3718F and
0x37190..0x3721F on 2026-11-02. The original image SHA-256 remains
`24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53`,
matching the manifest. Original callback table 0x1077A8 slots 4/5 resolve
to Game 0x4F090/0x4E760, agreeing with the existing R15/R14 bindings.
The entry-creation/removal instructions are compatible with IEP's logical
head insertion and matching-kind/identity removal; no new generic event
authority or shared-file edit is introduced.

| Original reference | Consumer assertion |
|---|---|
| Game 0x50250 -> 0xA00B0, return test before 0x50363 | Actual null applier returns recipient 0 before additional/event work; existing refresh supplies a produced list on equal metadata |
| Game 0x50363..0x503E9 | Fixed indices 2/3 have monsterOnly false, five absent additional stats and auxiliary disabled; no reached AI/stat/auxiliary request |
| Game 0x503EA..0x50407 -> 0x37110 at 0x50403 | Fixed first event/selector guards pass; cancel candidate's skill-kind entries with selected target state as owner identity |
| Game 0x50410..0x50454 -> 0x37190 at 0x5044B | Resolve each fixed selector and create on candidate in prepared order, callback values selected skill then caller level, skill kind and state owner identity |
| Game 0x50459..0x50462 | Actual recipient result 1 after normal event continuation, independent of a numeric create/remove return |
| Game 0x37110 | Matching kind and owner identity; currently executing matches deferred under IEP-R05; logical unlink order consumed, physical release internals not reverified |
| Game 0x37190 | Event/callback/value/kind/owner fields then newest-first linking; logical creation consumed from IEP-R01, allocator internals outside that contract |

The cancellation key is the selected state, not selected skill or invoking
skill. In the creation call, the candidate remains the owner; it is not
replaced by the caster. The prepared selected skill/level become the
entry's two callback values. The action's invocation identity never
substitutes for the selected identity here. Fixed index 3's event 1 insertion
precedes event 2 insertion, so event 2 is the newer head. No callback is
invoked merely by registering these entries.

### Meaningful cases and later callback source

V50(a/b) follows actual selected 87/82 preparation and an existing equal-level
refresh, then uses IEP's verified operations to remove old matching skill
entries and create the exact fixed bindings. Unrelated item-kind or other-state
entries remain in order. V50(c) has a lower requested level and therefore
produces null before refresh/cancellation/registration, leaving all old entries
and expiry unchanged; outer traversal still returns 1. No application result
or allocator effect is supplied to obtain any of these continuations.

V51 starts from V50(a)'s actual registered metadata and issues explicit
event 1 then event 2 requests through IEP-R07. This is an event-consumer boundary,
not a claim that a combat producer naturally issued these events. The entries
supply owner T, other Q, registered skill 87 and level 1 to R14. R14 still
obtains its calculation source from T's existing state 60 list, whose old
monster/id 7 source is absent from the registry; it does not substitute the
new caster A or other Q. The captured empty expression produces zero without
bytecode evaluation.

Despite zero percentage, the first actual R19 write replaces Q's own life
200 with 300 and adds its delta 100 to resolved 300, yielding 400. The next
event's actual current-life read is 400; replacing own 300 with 400 adds 100
to resolved 400, yielding 500. Both callbacks/dispatchers return 1. This
combines actual registered arguments with the already verified prepared
direct writer, rather than a supplied callback success or view-update fixture.
List/event identity, owner metadata and unrelated records remain unchanged.

The original timed-list installation, arbitrary creation/replacement, physical
allocator/release/failure behavior, wider callback combat/overlay effects,
later timer removal and natural event producers remain their existing gates.
The new code-verified consumer slice does not close Q01-Q04 or promote draft.

### Observation status, 2026-11-02

No new attributable action, effect or callback gameplay observation was obtained.
The previous forest route remained inconclusive, and no absence inference or
runtime promotion follows. The existing observation gate remains open.

### Fresh full reader and advisory triage, 2026-11-02

A fresh smaller blind reader recomputed the entire R01-R21/V01-V51 contract
using only the specification and its dependencies. Counts were MISMATCH 0,
GUESS 0, GAP 0, DISCRIM 0, GATED 5, COVERAGE 4 and HARDEST 3. The five gates
are existing R02 and Q01-Q04; all remain. Each empty defect list requires no
change. The exact report is preserved privately.

The R07 positive-level wrapping advisory needs no change: V21 covers its
positive/nonpositive control branch; SCX-R21 and numeric foundations own
wrapping, with SCX-V03(b) discriminating multiplication wrap. No extreme-level
original-profile execution is credited. The R13 negative-radius/two room
fallthrough and absent-dynamic/static-coordinate advisories add V52: two
rejected room branches forbid head reads, both fallthrough admission paths
reach actual traversal, and local absent-path zeros differ from static reads.
R20/V47 is not credited as R13 coverage. Remaining static kinds use the same
R08 rejection family exercised by V22 and are not individually claimed.
The R21 executing-entry advisory needs no change: IEP-R05/R07 own the logical
operation, and IEP-V09/V09(a) discriminate deferred in-call presence/removal
independent of result. V50 supplies candidate/key/kind integration, without
claiming executing-entry coverage or physical release.

All three HARDEST outcomes require no change. V50 asserts refresh before
cancellation, the state key, candidate owner, exact insertion order and null
guard. V51 asserts registered arguments, absent old source, exact callback
roles and sequential actual life writes/reads without a natural combat
producer. V40 distinguishes capped returns from stored deltas and later
reads, including the ignored negative return. V52 changes vectors and its
Id-map association only. No behavior or gate changed.

The fresh focused V52 delta independently recomputed all changed material and
returned MISMATCH 0, GUESS 0, GAP 0 and DISCRIM 0; each empty list needs no
change. It confirmed the four room outcomes, forbidden rejected-room head
reads, local zero/static coordinate sources, signed wrapped distance,
player-only recipient rejection and continued traversal to final processor 1.
The full review through V51 plus this final vector-only delta completes the
reader stopping rule for R01-R21/V01-V52. The reader's derivation and final
report are preserved privately; its references to covered branches are not
new advisory findings or a new full review. No contract change followed.
All Q01-Q04, owner allocation/release/failure and observation gates remain;
no new gameplay input, recorder or observation occurred during this review.
