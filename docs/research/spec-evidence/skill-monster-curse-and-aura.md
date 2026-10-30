# Evidence: Monster curse and aura skills

Contract: [skill-monster-curse-and-aura](../../specs/skill-monster-curse-and-aura.md).
Identity: [LOD113C-D2GAME](../source-manifest.md), Windows x86 LoD Expansion 1.13c.
Read date: 2026-10-29. No gameplay confirmation is claimed; the named text cells have the table-data record below.

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

MCA-R02: unresolved; raw-byte decoding, local/imported callees, data and integration
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
