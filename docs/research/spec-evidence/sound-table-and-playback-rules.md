# Sound table and playback evidence

Spec: [Sound table and playback rules](../../specs/sound-table-and-playback-rules.md)

## Identity and methods

Read date: 2026-10-29. Original Windows x86 LoD Expansion 1.13c.
[Source manifest](../source-manifest.md) artifacts `LOD113C-D2CLIENT` and
`LOD113C-D2SOUND`. The inspected D2Client image was rehashed and matched
`dd8bc6025de921216a97c17f97cd1a50fbb85926e838ec60e13451448836d906`.
No private reconstruction is required to locate the public references below.
Analyst descriptions are not original exported names.

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2CLIENT | 0x5f020 | none established | audio update dispatcher | STP-R02 | disassembly; call site 0x5f052 reaches 0x88dd0 |
| LOD113C-D2CLIENT | 0x88dd0 | none established | client voice update | STP-R02 | disassembly; isolated transition block 0x890cb–0x89111; other lifecycle branches remain gated |
| LOD113C-D2CLIENT | 0x5dfd0 | none established | sound table loader | STP-R01, STP-Q01 | disassembly lead; original SoundHdr.cpp assertion and sounds.txt path anchors |
| LOD113C-D2CLIENT | 0x5d7a0 | none established | voice playback admission | STP-R03, STP-R04 | disassembly read; unresolved local callees prevent verification |
| LOD113C-D2CLIENT | 0x5ce70 | none established | voice slot selection | STP-Q01 | disassembly read with unchecked encoded instruction runs; unresolved |
| LOD113C-D2CLIENT | 0x5ee70 | none established | sound-group selection | STP-R03a | complete disassembly and independent native instruction corroboration; supplied valid group/history/seed boundary |
| LOD113C-D2CLIENT | 0x5ebc0 | none established | sound asset loading | STP-Q02 | disassembly read; local loading/eviction paths unresolved |
| LOD113C-D2CLIENT | 0xd2e8 | import thunk | device start thunk | STP-Q02 | import directory resolves D2Sound ordinal 10037 to RVA 0x6f30 |
| LOD113C-D2SOUND | 0x6f30 | 10037 | buffer start | STP-R04a | byte-matched reconstruction; valid supplied buffer/device-operation boundary |
| LOD113C-D2CLIENT | 0x89640 | none established | current-context RNG helper | STP-R03a | complete disassembly and independent native listing; no owned callees |
| LOD113C-D2CLIENT | 0x88230 | none established | transition scheduler | STP-R05 | complete disassembly; valid supplied schedule boundary |
| LOD113C-D2CLIENT | 0x87e90 | none established | second-emitter-association test | STP-R05 | byte-matched reconstruction; complete helper |
| LOD113C-D2CLIENT | 0x87ed0 | none established | direct volume setter by handle | STP-R05 | byte-matched reconstruction; complete helper |

## Transition stage rationale

STP-R02 maps voice state as follows: active dword at +0x45; initial/final
volume dwords +0x49/+0x4d; interval start/end dwords +0x51/+0x55;
current volume dword +0x1c. Current tick is the audio update argument.
These locations describe evidence, not required implementation storage.

The 70 original bytes at D2Client RVA 0x890cb through 0x89110 were compared
with the literal instruction bytes in the disassembly and matched exactly.
Manual opcode corroboration: the end comparison uses unsigned `jbe` to
retain active state at equality; before-start uses unsigned `jae` into
interpolation. The product uses the low signed 32-bit `imul` result,
followed by `cdq; idiv` and endpoint addition. Completion writes final
volume then zero to active. The inactive branch bypasses every write.
No call or RNG operation exists within this stage. It does not verify the
producer of the interval or the later device-volume conversion.

The dispatcher calls the voice update at RVA 0x5f052. A direct relative-call
scan of the matching original image corroborated that target. Within the
voice update, the transition block precedes the playback/admission call at
0x892cd to 0x5d7a0; that function calls slot selection at 0x5d8ac and group
selection at 0x5d7d8. Call scanning supplies locations, not proof of all
execution eligibility.

## T01: effective sound cells

Source profile: fleet-reported vanilla `lod-113c`, LoD family, staged original
Game.exe SHA-256 `74fe9c092a521f7710392548c82f81544e531107db2358617e83818874db40a2`.
Date: 2026-10-29. Effective virtual path `data/global/excel/sounds.txt`.
Requested offset 0, length 60000; success, returned 60000, total 505304 bytes.
Base64 body decoded; complete tab-separated header and complete target row
were parsed. Partial-file coverage is sufficient for these cells only.

| Row identity | Column | Raw cell | Interpretation | Supports |
|---|---|---|---|---|
| Sound=cursor_pass; explicit Index=1 | Volume | 255 | decimal integer 255 | STP-R01 |
| same named row | Group Size | 0 | decimal integer 0 | STP-R01 |
| same named row | Stop Inst | 1 | decimal integer 1 | STP-R01 |

No physical supplying MPQ or complete override inventory was captured.
No effective-file hash was captured because the full file was not retrieved.
No correspondence between text cells and the runtime compiled/cache record
was established. The original header has `Volume`; it does not establish
the later-version minimum/maximum volume metadata described by some leads.
No column name establishes playback meaning. Table reads never constitute
gameplay confirmation.

## Search and remaining boundaries

The existing coverage candidates and audio research document were checked;
the latter's modern design suggestions and 1.14d data are not 1.13c evidence.
Search anchored on the loader string `sounds.txt` and SoundHdr.cpp assertion,
then original-image direct calls into the named playback candidates. Root,
output, data and string anchors were pursued. Ghidra instance discovery on
2026-10-29 reported no running instances, so no Ghidra reading is claimed.
Existing disassembly was used read-only; only the transition stage's encoded
instruction run was independently byte-corroborated in this record.
This is not a completed six-rung absence search.

The initial group-selector reading left its helper and input/seed provenance
unclosed; the continuation below closes the supplied-input algorithm.
Slot selection 0x5ce70 compares request fields and invokes D2Sound operations;
their complete semantics remain unresolved. Update 0x88dd0 also accesses
loop/duration/duplicate fields, calls local eligibility helpers and retires
voices. Reading those branches is not enough to assign field names without
loader/consumer correspondence and complete callee closure. The initial
pass verified only STP-R02; the continuation below records additional stages.

## O01: transition observation

Use an original Expansion 1.13c single-player character, Normal difficulty,
in the Rogue Encampment. No special class, level or equipment is required
for the collection setup. Start a naturally audible voice with an identified
transition producer, record that voice's transition endpoints and ticks,
and sample current volume/active state before start, inside the interval,
at end and after end. Compare with STP-V01–STP-V05 where inputs coincide;
otherwise recompute the same rule using actual measured inputs. Establish
the producer and its duration units first: an audible fade alone does not
identify this stage. No memory writes or synthetic function calls are needed.

Collection setup on 2026-10-29 first attempted D2Client RVA 0x5d7a0 at the
title screen; it failed because D2Client was not loaded. This is an invalid
capture, not a quiet branch. Entering single player with an Expansion level-1
Barbarian in town loaded D2Client and the same trace arm succeeded. This
admission entry cannot itself verify the transition's volume writes. No
STP-R02 vector has a matching gameplay observation.

The successful admission capture returned four inspected records, sequences
0–3, with dropped=0 at that read. Each caller return location was D2Client
RVA 0x892d2, corroborating the update-to-admission path. Requested cue 2768
was changed respectively to 2770, 2769, 2771 and left 2768; admission returned
nonzero slot results in all four records. The argument window was complete
(88 requested and returned bytes). These are identity and group-selection
observations, not verified STP-R03 results: their RNG state and all selector
inputs were not captured. Later the trace reached its 60-record cap and
reported already stopped when disarm was requested. The owned game was then
closed. No sequence beyond the four inspected records is claimed here.

An independent native instruction listing was also read for the complete
group selector 0x5ee70–0x5ef4b, corroborating its encoded instruction runs.
Its unresolved status remains because helper and caller provenance are still
open, rather than because a group weighting rule was assumed from a header.

## Review and publication state

The current document remains draft. Self-review separates narrow data proof,
the supplied-input transition stage and unresolved integration. Fresh blind
reader review is pending coordination; no approval is inferred.

Fresh blind reader full pass 1 (2026-10-29) reported MISMATCH 0, GUESS 0,
GAP 0, DISCRIM 4, COVERAGE 6 and HARDEST 3. Discrimination findings 1–3
identified missing unsigned high-bit tick and low32-product/signed-division
examples; STP-V06 and STP-V07 add these exact checks. Finding 4 concerned
random-seed consumption: the contract already specified no draw; the vector
preamble now explicitly asserts zero draws and an empty draw log. The six
coverage findings are advisory branch-coverage limits, not claims that the
current examples exercise every branch. The three hardest points are
governed by the explicit supplied-input, signed arithmetic and interval
branch definitions; no different arithmetic was inferred. A fresh full
reader rerun is required for the added vectors and remains pending.

Mechanical lint on 2026-10-29 reported zero errors and three warnings: one
public verification-method phrase and two references to the unavailable
analysis service. They document methods and limitations, disclose no private
paths or session identities, and are retained deliberately.

## Fresh blind-implementer review: full pass 2

2026-10-29: MISMATCH 0, GUESS 0, GAP 1, DISCRIM 0, GATED 2, COVERAGE 6, HARDEST 3. The independent reader recomputed all seven executable vectors, including unsigned high-bit completion and wrapped signed interpolation, and found no output difference. GAP outcome: no change; the pending STP-Q03/STP-O01 observation is already an explicit observation gate, not missing supplied-input semantics. GATED outcomes: STP-Q01 and STP-Q02 remain open integration gates.

Advisory COVERAGE outcomes: no change for start equality, fractional increases, nonzero interpolation start volume, negative signed duration, wrapped final addition and intervals crossing the high-bit boundary. Rule text specifies the operations and branch order; existing vectors discriminate retention, endpoints, truncation, unsigned comparison and low32/signed-product behavior. These unobserved partitions remain stated coverage limits; the review is not exhaustive branch testing. HARDEST outcomes: no change for unsigned ordering with signed interpolation, low32 product before division, or end-check precedence/equality; each is explicit and the final vectors distinguish them. The current supplied-input contract meets the reader stopping rule; full playback and runtime confirmation remain gated. No user approval is inferred.

Final mechanical validation on 2026-10-29: 0 error(s), 3 warning(s). Warnings were reviewed: temporal field names describe current retained state, and source-reading method names describe public verification methods rather than private reconstruction paths. No implementation or user approval is asserted.

## Continuation: group, history, scheduler and buffer-start closure

2026-10-29. Prior reviews above apply to the earlier seven-vector contract;
the added rules/vectors require a fresh full review.

### Group and history

D2Client 0x5ee70 is called from admission at 0x5d7d8, reached by the voice
update call at 0x892cd from frame dispatcher 0x5f052. The original Group Size
descriptor string at RVA 0xd564c is associated with byte +0x3d by loader
0x5dfd0; Volume at 0xd5658 maps to +0x3c. Fade In at 0xd563c maps to
+0x3f, Fade Out at 0xd5630 to +0x40, Stop Inst at 0xd5618 to +0x42,
Compound at 0xd5600 to +0x45 and Priority at 0xd55d0 to +0x4e.
These descriptor associations identify consumers; they do not replace T01
or establish full text/runtime correspondence for an arbitrary row.

The complete native group-selector instruction listing was compared with
the disassembly's encoded runs. Its only owned helper, 0x89640, was read
in full and corroborated by an independent native instruction listing.
Both use the seed at +0x20/+0x24 of the context selected through D2Client
global RVA 0x11bbfc. Their step and modulus/mask behavior agrees with
NRF-R01/NRF-R02. The helper is used only for the preliminary small-group
draw. The repeated member draw is inline, including the same carry and
high-word writes. The helper has no local calls. Native instructions
establish the small-group zero test, exclusion count, rejection back-edge
and the lack of any loop cap.

History reads are the base record's dwords +0x64 and +0x68. Selection
returns without writing them. Admission performs the two history writes
at 0x5d82e–0x5d83a after its early suppression branch, before asset/slot
and device requests. Thus STP-R03c is separately available with reaching
that stage supplied; it does not establish why the earlier test passed.

The original loader's later record loop normalizes zero Group Size bytes
to one and creates family links, so a named text cell with raw zero must
not automatically be treated as an unchanged cached zero. This is another
reason T01 remains text-cell evidence only.

The context pointer has several writers, including D2Client 0xa6674. That
routine was read, but its wider unit/context lifetime was not closed.
An owned in-game read of the pointer and its target identified unit type 0,
unit id 1, matching the current single-player character. This corroborates
the player-context seed source for that case; it is not a proof that every
pointer assignment or earlier record uses the same player. Supplied seed
selection is verified; broader context lifecycle remains STP-Q01.

### Scheduler

The scheduler's caller locations include 0x4f9b8 and 0x4fa0d in the ambient
controller, and wrappers at 0x88594, 0x88759 and 0x8891b. A direct relative
call scan of the matching original image corroborated each location.
Within 0x88230, requested duration enters EAX; the stack arguments are
handle, target volume and delay. This is binary navigation information,
not a required API. The scheduler's own normal-return control flow was
read completely; its two normal local callees 0x87e90 and 0x87ed0 were
read in their byte-matched reconstruction forms.

The former finds the first matching handle at voice +0x08 and checks the
association head at +0x28 for a second node. A second node returns true
and suppresses scheduling. The latter finds the first matching handle
and assigns current volume +0x1c; it does not clear transition state.
With stable supplied lists, both resolve to the same first record already
selected by the scheduler. The scheduler stores +0x49/+0x4d endpoint
volumes, +0x51/+0x55 start/end ticks and +0x45 active; zero-target requests
also set +0x3d, whose retirement execution remains unclosed.

Unsigned comparisons at 0x882ae and 0x882d6 establish duration minimum and
same-target end precedence. The old zero endpoint test at 0x882d9 prevents
nonzero retargeting. The zero-duration/nonzero-delay branch invokes the
diagnostic/termination path; those effects are excluded from valid STP-R05
and remain STP-Q02. Clock units and wrap integration are not promoted by
describing the observed low32 arithmetic.

### Buffered device wrapper

D2Client thunk 0xd2e8 resolves D2Sound ordinal 10037 to 0x6f30. The
byte-matched reconstruction was read first. On a valid context it calls
the buffer's Play method with reserved values 0,0 and flags 1, checks
the signed HRESULT, writes context +0x1c only on nonnegative status,
then leaves the lock on both normal branches. The import directory
resolves D2Sound IAT RVAs 0xf11c/0xf120 to KERNEL32 EnterCriticalSection
and LeaveCriticalSection by name. Device/locking internals are supplied
external operations, not a claim about a particular device implementation.
The installed platform interface header identifies flag 1 as
DSBPLAY_LOOPING. Thus STP-R04a requests repeating low-level buffer playback;
the client still owns logical one-shot/loop retirement. The helper never
reads Sounds.Loop. Null-context diagnostics remain outside this valid
supplied-input wrapper boundary.

The admission lookup was also pursued through the split body 0x5cc10 /
0x5cc24 and its D2Sound ordinal 10044 target 0x6950, read as a byte-matched
reconstruction. That target reads the same start marker rather than making
an OS playing-status query. Complete slot eviction, loading, streamed start,
and retirement remain STP-Q01/STP-Q02. The fresh-instance allocator at
0x87d80 was read as byte-matched reconstruction; its list-publication
dependency is not closed here. No complete request admission contract is
claimed from these additional leads.

### Additional owned runtime observations

Source profile: reported vanilla lod-113c, new level-1 Barbarian, Normal,
Rogue Encampment; 2026-10-29. The seed context header read matched unit type
0 and current character id 1. Code-profile identities are the manifest
baseline; complete effective archive/override identities remain uncaptured.

| Vector | Source profile | Date | Module + RVA | Records | Result |
|---|---|---|---|---:|---|
| group-selection follow-up, no promoted vector | reported vanilla Expansion 1.13c | 2026-10-29 | D2Client +0x5ee70, caller return +0x5d7dd | 6 inspected, dropped 0 | base 2768 returned 2769,2771,2770,2768,2771,2769; seed/history snapshots at each entry absent |
| buffered start follow-up, partial STP-V18 assertions | same profile | 2026-10-29 | D2Sound +0x6f30, caller return D2Client +0x5d958 | 4 inspected, dropped 0 | valid 32-byte context windows, entering marker 0 changed to 1, wrapper returned 1; exact device HRESULT and inner operation log unobserved |

Both traces stopped at their configured record caps (16 and 4 respectively)
and disarm requests reported already stopped. The owned game was closed.
No complete STP-V08–STP-V19 vector is claimed runtime-observed: group draws
and entry seed/history, and the device wrapper's exact HRESULT/operation
sequence, were not collected. Later idle reads of the sound cache confirmed
Group Size 4 for base cue 2768 but cannot reconstruct historical entry
seed/history. The available entry trace snapshots expose argument pointers,
not arbitrary global seed/history snapshots; no memory writes or synthetic
function calls were used to manufacture a match.

### Reproduction requirements for added stages

For group selection, walk a fresh character on ordinary town terrain and
observe the group selector entry and every seed step, with the base cached
record and both history entries captured at the same entry. Recompute
STP-R03a using the actual complete inputs; reject a partial record even if
the returned cue belongs to the expected group. Capture the admission
history commit separately to verify STP-R03c and its ordering.

For scheduling, observe the ambient controller entering a level whose
supplied environment requests a transition. Capture the selected voice's
handle, association chain and entering transition state, duration, target,
delay and clock, then its writes. Compare STP-R05 before later frame updates
execute STP-R02. An audible fade alone is insufficient.

For buffered playback, capture the caller, valid buffer context, exact
device HRESULT, critical-section entry/exit and the repeating-play request.
Compare marker and result against STP-R04a; the wrapper return alone does
not establish the exact supplied-HRESULT STP-V18 or the operation-order log.


## Fresh blind-implementer full pass 3 — 2026-10-29

MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 2 integration gates, COVERAGE 8 advisory areas, HARDEST 3. The reader independently recomputed all nineteen vectors and group draws/final seeds. All stated discriminators rejected their named alternatives. Existing STP-Q03/STP-O01 observation limitations were identified separately from the two integration gates; no observation was promoted.

COVERAGE outcomes: no change for absent-table/size-zero group returns and size-three preliminary draw; history persistence through later admission failure; zero/no-match handles, a raising Fade In minimum, immediate zero-duration scheduling and a shorter replacement; or a positive buffered-device result. The rules explicitly define each branch, and existing vectors distinguish no-draw versus group selection, ordered history commit, schedule rejection versus persistence, and signed-negative versus nonnegative device handling. These remain advisory partitions, not exhaustive branch coverage. HARDEST outcomes: no change for unsigned branch order with signed interpolation, low32 product before signed division or end-tick precedence/equality; current V04 through V07 distinguish those operations. Retain STP-Q01/STP-Q02 and observation Q03. Reader stopping rule met for the current supplied-stage contract; full playback parity and user approval are not established.
