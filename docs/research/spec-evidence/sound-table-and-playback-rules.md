# Sound table and playback evidence

Spec: [Sound table and playback rules](../../specs/sound-table-and-playback-rules.md)

## Current evidence scope

The dated records below retain their original coverage. The 2026-10-31 lifecycle continuation expands the supplied-input contract; no STP vector has matching gameplay confirmation, and the document remains draft with its named integration/observation gates.

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
output, data and string anchors were pursued. This initial record uses
disassembly rather than a Ghidra reading; only the transition stage's encoded
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

The initial capture attempt on 2026-10-29 at D2Client RVA 0x5d7a0 was invalid
because D2Client was not loaded; it is not evidence of a quiet branch. A
subsequent original-profile single-player capture with an Expansion level-1
Barbarian in town recorded the admission entry. That entry cannot itself
verify the transition's volume writes. No
STP-R02 vector has a matching gameplay observation.

The successful admission capture returned four inspected records, sequences
0–3, with dropped=0 at that read. Each caller return location was D2Client
RVA 0x892d2, corroborating the update-to-admission path. Requested cue 2768
was changed respectively to 2770, 2769, 2771 and left 2768; admission returned
nonzero slot results in all four records. The argument window was complete
(88 requested and returned bytes). These are identity and group-selection
observations, not verified STP-R03 results: their RNG state and all selector
inputs were not captured. The capture reached its 60-record cap; no sequence
beyond the four inspected records is claimed here.

An independent native instruction listing was also read for the complete
group selector 0x5ee70–0x5ef4b, corroborating its encoded instruction runs.
Its unresolved status remains because helper and caller provenance are still
open, rather than because a group weighting rule was assumed from a header.

## Review and publication state

The current document remains draft. Self-review separates narrow data proof,
the supplied-input transition stage and unresolved integration. Fresh blind
reader review had not yet run at this initial revision.

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
An in-game read of the pointer and its target identified unit type 0,
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

### Additional runtime observations

Source profile: reported vanilla lod-113c, new level-1 Barbarian, Normal,
Rogue Encampment; 2026-10-29. The seed context header read matched unit type
0 and current character id 1. Code-profile identities are the manifest
baseline; complete effective archive/override identities remain uncaptured.

| Vector | Source profile | Date | Module + RVA | Records | Result |
|---|---|---|---|---:|---|
| group-selection follow-up, no promoted vector | reported vanilla Expansion 1.13c | 2026-10-29 | D2Client +0x5ee70, caller return +0x5d7dd | 6 inspected, dropped 0 | base 2768 returned 2769,2771,2770,2768,2771,2769; seed/history snapshots at each entry absent |
| buffered start follow-up, partial STP-V18 assertions | same profile | 2026-10-29 | D2Sound +0x6f30, caller return D2Client +0x5d958 | 4 inspected, dropped 0 | valid 32-byte context windows, entering marker 0 changed to 1, wrapper returned 1; exact device HRESULT and inner operation log unobserved |

Configured capture caps were 16 and 4 records respectively.
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

## Continuation: slot displacement, loaded assets and audio clock

2026-10-30. Original Windows x86 Expansion 1.13c, manifest artifacts
`LOD113C-D2CLIENT` and `LOD113C-D2SOUND`. This continuation adds STP-R03d,
STP-R04c, STP-R06 and STP-V20–STP-V27. Earlier reader results remain dated
records for their earlier boundaries; they do not review these additions.
The broad lifecycle gates remain open.

### Slot selection and admission

The complete admission function D2Client RVA 0x5d7a0 and complete slot
selector 0x5ce70 were read, then corroborated against an independent native
instruction listing from the matching original image. The sixteen slots
occupy RVA 0x1175c0 through 0x1177bf with stride 0x20. Their buffer, class
and voice references are at +0/+4/+8. The allocator 0x87d80 was read first
in its byte-matched form: it chooses the first unused entry of a fixed
200-entry ascending-address pool, zeroes its 89 bytes, marks it in use,
publishes it to the active list and assigns the next handle. Thus the
selector's unsigned voice-pointer ordering has a semantic pool-index
equivalent; it is not an unspecified heap-address tie breaker.

The free-slot pass requires all three conditions: no associated voice,
matching class and zero start marker. D2Client import 0xd2be resolves
D2Sound ordinal 10044, RVA 0x6950, read in its byte-matched form. It reads
wrapper +0x1c, with no OS status query. The occupied-slot scan compares
voice priority byte +0x3c, key float +0x18 and start tick dword +0x34.
Native branches at 0x5ceca/0x5cecc select lower priority; x87 comparisons at
0x5cece–0x5cee6 select the greater finite key; 0x5cef0 selects an earlier
unsigned start tick; 0x5cef4–0x5cefe selects the lower unsigned pool entry.
Final admission is a separate comparison at 0x5cf23–0x5cf33: higher incoming
priority wins; equal priority wins only for the greater incoming pool entry.
It does not repeat the distance/start comparisons against the incoming
voice. These distinctions are exercised by STP-V20–STP-V22.

Accepted displacement queries the old buffer position through import
0xd30c and stores its result at old voice +0x41; import 0xd336 then stops
playback. The old voice's +0x30 becomes 2, +0x2c is cleared, and slot +8
is cleared. The slot is returned before any incoming association is
stored. D2Sound ordinal 10006 at RVA 0x6980 and its stream-stop helper
0x6680 were read first in their byte-matched forms. The valid selector
boundary supplies returning backend operations; full position production
and stream/file/device failure integration remain STP-Q02.

Admission's wider reading also establishes its ordering frontier. It
selects the group unless request flag bit 0 suppresses selection; it then
looks up the earliest-start marked voice for the selected cue via the split
0x5cc10/0x5cc24 lookup. A nonretiring match suppresses the request when
unsigned incoming-start minus old-start is below 2 and old unsigned priority
is at least incoming priority. The base history writes follow that test,
before loading or slot selection. After loading, failure state can reject
the request. Stereo chooses class bit 0; otherwise cached backend mode
1 adds bit 1, mode 2 adds bits 1 and 2. The mode getter at D2Sound 0x9b20
and setter 0x9b00 were read in their byte-matched forms. The getter's cached
field is a mode, regardless of an analyst label suggesting an error code.
These additional admission facts are navigation and closure evidence;
the full STP-R03b integration is not promoted by this continuation.

### WAV parsing and loaded record validation

Client RVA 0x5cd10 invokes D2Sound ordinal 10025 through import 0xd2c4.
The export is RVA 0x8980; its only local parsing helper is 0x8790. Both
functions were read in full, and original-image native instructions
independently corroborated the parser. Client 0x5cd10 was read in full.
The parser receives the asset buffer, a mutable length and a format
destination. Its explicit RIFF/WAVE tests, chunk scan, copied format
fields and returned PCM slice support STP-R04c. No format-tag or RIFF-size
comparison is present. The helper skips chunk payloads by their exact
length, not by a rounded alignment size; STP-V25 discriminates this.

The client checks bits/sample at format +0x0e against 16 and sample rate
at +4 against 22050. It writes failure byte +0x81 on rejection. Success
stores PCM reference +0x74, data length +0x78 and assigns record Stereo
byte +0x50 from `channels==2`. It does not clear failure on success and
does not otherwise validate channel count. The loader descriptor string
Stereo at client RVA 0xd55c0 is associated with this same +0x50 field.
Consequently, buffered routing cannot be inferred from the text cell
without accounting for loaded WAV validation. These are loaded-record
rules; they do not establish a text-to-cache row correspondence or the
stream parser's separate acceptance behavior.

Further source-lifecycle reading covered client 0x5ebc0 (initial load),
0x5da80 (async finalization), 0x5ea90 (memory-pressure eviction), 0x5dae0
(access/LRU update), 0x5dd50 (loaded asset disposal), 0x5db30 (resource path)
and 0x5ed90 (pending-load coordination). The synchronous asset allocator
0xd810 and imported asynchronous descriptor/accessor operations were
followed. Their remaining file-I/O/completion dependencies are not replaced
with assumed results in a promoted full loading rule.

The list dependencies were resolved from the actual import directory:
client 0xd276 → D2Sound ordinal 10034, RVA 0x8710 (prepend);
0xd2ac → ordinal 10018, RVA 0x8640 (append);
0xd258 → ordinal 10062, RVA 0x86c0 (remove first matching entry);
0xd25e → ordinal 10063, RVA 0x8920 (clear list);
0xd264 → ordinal 10053, RVA 0x86a0 (contains).
All five bodies were read completely. These are ordinary two-word list
nodes; the routines do not infer a voice policy from their analyst names.
Allocation/free dependencies remain separate from their list ordering.

### Counter identity and selected unit context

The complete client dispatcher RVA 0x5f020 was read and independently
corroborated against original native instructions. With global audio
enable nonzero it passes the counter at global RVA 0x11c268 to 0x4f860
(ambient/environment), 0x45660 (music selection), 0x5ed90 (pending loads)
and 0x88dd0 (voices), then calls marker retirement 0x5d080 and increments
that counter. The disabled branch skips all those operations. Counter
increment is at RVA 0x5f062, after voice updating. This proves enabled-update
units, rather than milliseconds or a universal server-frame cadence.

The main-loop callers at 0x44bcb and 0x44ea0 were independently pursued:
the ordinary latter route follows a world update, while the former is an
early render/pause route without that world-update requirement. The world
counter and audio counter must therefore remain distinct. A disk interval
constant does not justify assigning a universal 25 Hz rate to audio
invocations. The backend buffer worker 0x7800 separately samples its own
millisecond clock for its fades and waits; that reading does not change
client STP-R02/STP-R05 units.

The current sound-selection seed is the +0x20/+0x24 pair of the unit
selected through global RVA 0x11bbfc. Native assignment sites include
0xa6674 and 0xad734; the complete setter 0xa6640, packet handler 0xad6e0
and unit lookup 0xa4e20 were read. The packet handler looks up the supplied
unit type/id, checks the found unit's type through the lookup, and installs
it only when found. The setter accepts an existing unit or null. Removal
and teardown assignment sites also clear the selected context. Thus the
general owner is the selected client unit; the earlier observed player
case does not prove a player-only type restriction for every assignment.
Wider unit creation/reseeding remains open; the numeric foundation's
NRF-Q08 already records the player-load seed producer gap. This continuation
does not supply a guessed deterministic initial audio seed.

### Lifecycle reading and remaining integration

The complete 0x88dd0 coordination body was read, together with 0x880e0
(stop/save position), 0x88150 (detach references), 0x87bd0 (special cue-range
lookup), 0x87c30 (same-family active lookup), 0x87fe0 (compound reuse),
0x88a70 (request producer), 0x5d080 (marker-to-voice bridge) and 0x5ce30
(completed stream cleanup). These readings identify the normal full-chain
frontier; they are not a claim that all the corresponding contract vectors
or external operations have been verified.

The backend path was followed through D2Sound 0x7650 (prepare), 0x7090
(recreate device buffer), 0x69f0 (destroy buffers), 0x7580 (load parsed
data), 0x72d0 (copy/refill), 0x6a40 (silence fill), 0x7230 (lost-buffer
handling), 0x7800 (worker) and 0x66f0 (stream start). Stream start resolves
Storm ordinal 255 at RVA 0x2a650; the complete begin body was read. It
normalizes the requested byte offset modulo WAV data length and records
the initial position. Underlying stream pumping, file I/O, complete
failure/cleanup paths and matching original-game observations still require
closure before STP-R04b can be promoted.

### Additional original-game observations, 2026-10-30

Reported vanilla `lod-113c`, original LoD Expansion 1.13c launch profile,
level-1 Expansion Barbarian, Normal difficulty, Rogue Encampment. Natural
walking was the request source. Patch-state inspection reported archive
override and virtual clock disabled; the audio-focus accommodation was
active. Complete effective archive identities were not captured.

| Boundary | Inspected records | Actual result | Limits |
|---|---:|---|---|
| D2Client +0x5d7a0, caller return +0x892d2 | 20, cap 20, dropped 0 | walking admission captured; initial four requested cue 2816 and changed it to 2819,2817,2818,2819; all twenty returned nonzero slots | complete 89-byte voice windows; subsequent records include base cue 2768; same-entry seed/history, slot pool and full operation chain absent |
| D2Sound +0x6f30, caller return D2Client +0x5d958 | 6, cap 6, dropped 0 | each returned 1 and changed marker from 0 to 1 | each 80-byte wrapper window complete; per-invocation inner operation log absent |
| underlying buffer Play reached from D2Sound return +0x6f6f | 10, cap 10, dropped 0 | all observed reserved arguments 0,0, flags 1 and signed HRESULT 0; measured stack cleanup 16 bytes | separate capture window from the six wrapper records; Windows device-library file identity uncaptured; no inferred one-to-one join |
| D2Sound +0x66f0 | 0, dropped 0 | no stream-start record during the sampled walking/NPC-interaction attempts | successful gameplay movement and other audio requests establish a responsive run; stream eligibility/trigger not established, so absence is inconclusive |

The Play target was located by reading the actual buffer's vtable and
resolving the function before observing it. This establishes the listed
device arguments/results for these calls, without assigning that platform
implementation an original-game artifact identity. Lock entry/exit was
not captured. These observations therefore still do not satisfy every
STP-V18 assertion, nor any complete group/history/transition vector.
No new executable vector is marked runtime-observed.

The first four new voice snapshots contain little-endian cue bytes
`00 0b 00 00`, meaning 2816; later snapshots contain `d0 0a 00 00`,
meaning 2768. An initial transcription incorrectly treated the first
group as 2768. The record above uses the corrected decoding and preserves
the distinction between two observed bases; the correction does not
reconstruct any missing RNG/history input or promote a vector.

### Review status of this continuation

The earlier nineteen-vector reviews are preserved above. The fresh full
reader review below covers the twenty-seven-vector contract. Status remains
draft; complete playback parity and runtime confirmation remain gated.

## Fresh blind-implementer full pass 4 — 2026-10-30

MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 8, COVERAGE 16 reported,
HARDEST 3. The independent reader found no output mismatch, missing
supplied-boundary semantics or failed discriminator in the twenty-seven
vectors. The eight gated observations retain the stated boundaries; they
do not close STP-Q01, STP-Q02 or STP-Q03.

The reader's reported COVERAGE count is retained as 16; the following
groups describe the actual advisory cases, rather than claiming sixteen
individual cases or exhaustive branch coverage.

| Rule | Advisory cases | Disposition |
|---|---|---|
| STP-R02 | Exact start tick; rising nonendpoint interpolation | No change: unsigned branch order and the signed interpolation calculation specify both; existing endpoint, falling and wrap cases discriminate their stated alternatives. |
| STP-R03a | Absent table; Group Size 0; Group Size 3 preliminary draw | No change: the no-draw gate and size-2-or-3 preliminary draw are explicit; existing singleton, size-2 and larger-group vectors do not claim these additional cases were exercised. |
| STP-R03d | No eligible victim; lower incoming priority | No change: failure and the separate incoming-priority comparison are explicit. V20–V22 distinguish compatible-slot eligibility, victim ranking and equal-priority pool ordering; they do not exhaust all rejection branches. |
| STP-R04a | Successful start with entering marker 1; strictly positive device result | No change: every nonnegative result writes marker 1; zero-success and negative-failure vectors discriminate the result-sign boundary. |
| STP-R04c | Short/malformed header; truncated chunk; invalid format length; unsupported bits/sample | No change: parser bounds, signatures, format-length bounds and the client bits/sample check are explicit. V23–V26 cover their stated acceptance, rate rejection, unaligned skip and retained-field distinctions. |
| STP-R05 | Zero handle; missing voice; Fade In minimum; earlier candidate end; sticky retirement flag on a nonzero request; immediate zero target | No change: selection, duration minimum, replacement precedence and retained-state writes are explicit. These branches remain additional coverage opportunities, not observed cases. |
| STP-R06 | Enabled counter increment without wrap | No change: low32 increment specifies it; V27 deliberately distinguishes wrap and disabled behavior. |

HARDEST dispositions: no change for R02's unsigned branch comparisons
combined with wrapped multiplication and signed division; R03a's separate
preliminary/rejected draws and seed advancement; or R04c's unaligned chunk
skipping and retained fields. The reader found these rules explicit, and
the existing vectors distinguish the named alternatives. No advisory case
is converted into a claimed test or runtime observation. The reader stopping
rule is met for the stated contract; the wider lifecycle and observation
gates remain open.

Scoped validation on 2026-10-30: spec checker 0 errors, 9 warnings;
whitespace check passed for this spec, companion and gap file. All local
link targets exist; the twenty-seven vector headings are unique and
contiguous, all twenty-seven runtime flags remain no, and status remains
draft. Public-path/session-identifier screening found no private path or
collection-session identifiers in the three files. Seven warnings concern
existing public method descriptions; the other two concern the `binary32`
format name and wording about current retained state. They were reviewed
and retained without changing behavior. No game/implementation tests were
run for this documentation revision.

## Additional stream-start observation — 2026-10-30

Source profile: reported vanilla LoD Expansion 1.13c, original manifest
D2Sound and Storm artifacts; virtual clock and archive override disabled
for this launch. A bounded capture window covered the main menu, character
selection and entry of a level-1 Expansion Barbarian into Normal single
player. The stream tuple below was already present at the pre-entry
character-selection sample. It is not attributed to the town, a particular
character-selection action or the music-selection contract.

| Boundary | Count and coverage | Observed inputs | Observed output |
|---|---|---|---|
| D2Sound +0x66f0, caller return D2Sound +0x7f3b | 1 record; cap 16; dropped 0; complete 80-byte wrapper and 80-byte path windows | `data\global\music\act5\icecaves.wav`; path length 35 bytes, NUL present at offset 35; wrapper buffer capacity 262144; offset argument 0; loop argument 0; entering start marker 0, stream reference null, end marker -1 | return 1; start marker 1; new nonnull stream reference; end marker remains -1; measured stack cleanup 8 bytes |
| Storm +0x2a650, caller return D2Sound +0x6768 | 1 record; cap 16; dropped 0; complete 352-byte file-state and 16-byte buffer windows | file reference exactly matching the new reference written by the wrapper; buffer reference exactly matching the wrapper's entering buffer; capacity 262144; flags 0; starting byte offset 0; both optional scalar arguments 2147483647 | return 1; measured stack cleanup 28 bytes |

The two records have the same thread identity, exactly matching file/buffer
associations and the direct caller location expected from the original
stream wrapper. The path window contains its complete terminator; trailing
bytes after that terminator are not part of the path. The wrapper's +0x1c
marker changes 0→1 and +0x48 receives the matching file reference; its
+0x30 end marker is unchanged at -1. The original branch writes -1 there
on success, so these bytes alone do not distinguish that write from
retention of the entering -1 value.

This is a complete measured input/field-output tuple for the reached start
case, with a matching downstream begin call. It does not establish the
complete operation log: file-open and lock entry/exit were not captured in
the same invocation. A separate file-open capture reached its 64-record cap
with dropped 0 before supplying a matching D2Sound caller, so it is not
joined to this tuple. Offset 0 cannot distinguish multiplication by four
from no scaling; loop 0 does not exercise the nonzero-loop flag branch.
The data-format result, later pump/play/stop/cleanup sequence, audible
output and complete effective archive identity remain unobserved. No
existing executable vector receives runtime credit, and STP-R04b/STP-Q02
remain open. The full-pass-4 contract and its twenty-seven vectors are
unchanged by this additional observation record.

An independent byte-level audit reconstructed the post-call wrapper from
its before-image and deltas, checked every captured window length and
terminator, and confirmed the matching buffer/file references, capacity,
returns, stack cleanup and thread association. It found no discrepancy in
the tuple above. This targeted source/runtime audit does not extend the
full-pass-4 contract or promote a complete lifecycle rule.


## Additional stream progress and separate stop observations — 2026-10-30

The same original LoD Expansion 1.13c source profile was used for a normal
single-player Save and Exit, return to the main menu, and subsequent game
entry. These observations are menu-stream and backend field observations;
they are not an in-game music-selection case or a complete logical voice
lifecycle case. A preceding 90-second walking/NPC interaction attempt
captured zero starts, callbacks and stream stops, with zero dropped
records; dialogue did not visibly open, so stream eligibility remained
unestablished for that attempt.

| Boundary | Count and captured inputs | Observed output and limits |
|---|---|---|
| D2Sound +0x66f0 → Storm +0x2a650 | One record at each boundary, each cap 16 and dropped 0; complete 80-byte wrapper, 160-byte path and 352-byte file windows; path `data\global\music\introedit.wav`, complete NUL at offset 31; capacity 262144; offset 0; loop 0; optional scalars 2147483647 | Both return 1, stack cleanup 8/28 bytes; wrapper marker 0→1 and matching file reference installed. Same thread, direct downstream caller return D2Sound +0x6768, matching file/buffer associations. End marker remains -1; unchanged bytes do not prove its write. Zero offset/loop retain the earlier discriminator limits. |
| Storm +0x26930, caller return Storm +0x293f2 | 32 records, cap 32, dropped 0; every 80-byte request and 96-byte stream window complete; matching file and buffer references from the preceding start; data start 44, data length 8819504, buffer capacity 262144, actual-read field 32768 | All enter and leave stream state 2; none observes state 1→2. In every record the written-count field +0x30 becomes `low32(request.fileOffset - stream.dataStart + 32768)`: first 950272→983040, last 1966080→1998848. Measured stack cleanup 4 bytes. |
| D2Sound +0x6980 | Nine records, cap 16, dropped 0; all complete 80-byte wrapper windows | Eight buffer-only records have no captured field change. One stream wrapper, called from D2Client return +0x44e47, changes marker 1→0 and stream reference nonnull→null; its nested stop/end/cancel calls are the separate tuple below. |
| D2Sound +0x6680 → Storm +0x281f0 → Storm +0x27ef0 | One record each, each cap 16 and dropped 0; same thread and caller returns D2Sound +0x69a5, D2Sound +0x66a4 and Storm +0x28245; complete 80-byte wrapper and 352-byte file windows | Wrapper marker 1→0 and stream reference→null; end argument matches entering wrapper reference; file +0x130 changes 1→0; queued-cancel arguments are a null first filter key and the matching buffer reference. Stack cleanup 0/4/8 bytes. Storm end returns 1. D2Sound stop is void, so its recorded EAX is not a status. The canceller returns 0, meaning no matching queued request was removed; the stream-end caller ignores that result. |

The callback body was read completely and corroborated against the original
Storm instructions. It uses chunk-size global RVA +0x43114. This capture
does not directly record that global: 32768 is supported here by the
request actual-read fields and resulting counter updates. Device Lock,
Unlock and Play arguments/results, the returned lock lengths and copied
PCM bytes were not captured. These records therefore confirm the stated
request/stream field associations and counter updates, without claiming
that the full copy, padding, initial Play or error paths were observed.

The stop file window names `data\global\music\act2\desert.wav`, whereas
the earlier start names `introedit.wav`. The numeric file handle was reused,
so it cannot join these observations into one asset lifecycle. Complete
matched wrapper bodies and the original Storm end/cancel instructions
corroborate the reached call chain and field-clearing behavior. The capture
does not include the device Stop result, file-close invocation, stream-list
mutation, reference-count/free outcome or audible output. The unchanged
underlying buffer reference does not imply device-buffer release.

Every before-image/delta was reconstructed and checked for the two new
field subsets. All captured window lengths, caller/stack associations and
listed formulas passed that audit. No existing executable vector receives
runtime credit; STP-Q01, STP-Q02 and STP-Q03 remain open, and the contract
and twenty-seven vectors remain unchanged.

The independent targeted audit confirmed every callback and stop field
association above. It also corrected the initial return description for
Storm +0x27ef0: the helper returns whether any queued request matched and
was removed. Its first filter compares request +0x24; the actual file
handle is request +0x0c, so these are not equated. With the observed null
first filter and nonnull buffer filter, return 0 records no matching queued
request, rather than a stream-stop failure. The caller ignores the value.
This correction does not change a contract rule or runtime vector credit.

## Lifecycle source continuation — 2026-10-31

The following original-code evidence supports STP-R03b/R04b and 42 additional synthetic main cases. These cases have not been executed in the original game. Earlier observation records retain their dates and capture limits; STP-Q01–Q03 remain open.

### Identity, dates and methods

Original Windows x86 LoD Expansion1.13c only. The [source manifest](../source-manifest.md) remains the baseline identity record. Fresh on-disk module hashes used in the2026-10-31 source audits were:

| Module | SHA-256 |
|---|---|
| D2Client.dll | dd8bc6025de921216a97c17f97cd1a50fbb85926e838ec60e13451448836d906 |
| D2Sound.dll | b24dfb0159c76867e7f3d14602e73bccc37d59b7c8d2def88213b01084b61a72 |
| Storm.dll | 4b5fcaf87c98676e6be94f980b06b893c50f1505343cbbf038c5b331a2da2d5b |
| Fog.dll | 96c05770ab06b71a7a2d6abe5dc7c2f943b30ac8fe367e44bdbad7e2643bb863 |
| D2Gfx.dll, supplied window getter only | 50284647a28e8839028e8acb021a6ebe48c006e431139969ccd0fe126e6e8811 |

These are original-file identities, not retroactive per-call loaded-module or effective-MPQ identities. Existing2026-10-29/30 observations keep their recorded profile limits; a later hash does not fill missing historical context or join reused handles across different source files.

For each new source reading, the matching-build source registry was checked before choosing byte-matched C or assembly. C was read first when available. Original PE import/export tables and actual jump/call bytes, not analyst names, establish cross-module targets. Complete applicable bodies were read as listed below; unrelated trailing routines in a shared source range were excluded.

Native instruction listings were compared with section-mapped original PE bytes, including continuation bytes where recorded. Every reported compared range had zero missing bytes and zero differences. Representative audit sets include 8752 bytes in the independent formal audit,3504 in the queued-ownership/notification audit,1982 Fog bytes in the descriptor audit,896 Sound bytes in the Prepare check and 1888 Storm bytes in the later queue-policy check. These sets overlap; they are not summed as unique coverage. An earlier asset audit compared instruction-head bytes only and is not silently upgraded into complete continuation-byte coverage. Mechanical JSON/materialization/number/identity checks validate artifacts, not executable behavior.

The2026-10-29 group, transition and buffer-start evidence and2026-10-30 slot, parser, dispatcher and lower lifecycle readings are reused with their original limits. The2026-10-31 audits freshly reread many of those bodies and extend the explicit source/case contract. A retained earlier complete reading is not described as a new independent proof of every helper or external subsystem.

### STP-R03b navigation: producer through logical retirement

| Original module / RVA | Role and root-to-output relation | Method/date and boundary |
|---|---|---|
| D2Client0x5f020 | Enabled dispatcher reaches environment/music, pending loads0x5ed90, coordinator0x88dd0, marker bridge0x5d080, then counter increment. | Complete2026-10-30 and2026-10-31 reread/native corroboration; unrelated controller internals are not inferred. |
| D2Client0x88dd0,0x88d90,0x89640 | Coordinator calls owned periodic producer before sorting; producer applies unsigned counter modulo 3 then Rand1723 and request0x88a70. | Complete applicable bodies reread2026-10-31; actual RNG agrees with NRF-R01/R02; producer purpose and wider unit-seed lifetime not inferred. |
| D2Client0x88a70,0x87d80,0x87fe0 | Request gates, compound reuse, first-free200-entry pool, handle wrap, offset installation and repeated priority-byte boost. | Matched C allocator first; remaining full bodies assembly,2026-10-31. Source defines pool exhaustion and no silent handle-zero repair. |
| D2Client0x88c20,0x88550,0x88150,0x880e0 | Explicit owner removal, retirement scheduling, association cleanup, query/save/stop/detach. | Complete applicable source/native reads2026-10-31, reusing dated scheduler/list evidence. Automatic unit-kind cleanup routing remains STP-Q01. |
| D2Client0x87c30,0x87bd0,0x88dd0 | Ordinary earliest family candidate then special-range replacement, including replacement-to-null and flag-independent special lookup. | Complete bodies and branch ordering checked2026-10-31; finite synthetic positions do not establish generic positional eligibility. |
| D2Client0x5d7a0→0x5ee70→0x89640 | Admission retains entering base/history record, performs actual selection and keeps selected cue/seed effects through later failure. | Group proof from2026-10-29/30 reused; full source/native reread and coupled two-draw case added2026-10-31. |
| D2Client0x5cc10→0x5cc24 | Enabled sixteen-slot marked search chooses smallest unsigned start, first on equality; retirement tested after selection. | Split tail followed completely and native bytes checked2026-10-31; not a20-byte-only reading. |
| D2Client0x5ce70,0x5d7a0,0x5db30 | Slot acceptance/displacement; prepare/load/properties/start; cue-range path construction and fatal length bound. | Complete applicable bodies reread2026-10-31. Constants located in original PE strings; generic read/decode remains separate. |
| D2Client0x5d080→0x5ce30/0x880e0 | Marker/stream predicate bridge and ordinary bookmark query/stop; later coordinator owns deletion/restart. | Full source/native check2026-10-31. Supplied marker 0/attachedstream checkpoint does not prove normal cadence produces it. |
| D2Client0x5d1d0,0x5dc70,0x5dce0 | Acoustic/property update and bound helpers reached between lifecycle eligibility and record access/LRU refresh. | Complete finite path reads/native check2026-10-31; full acoustic-update/cache/base-scaling contract remains supplied under S3D-R07/Q02, not promoted here. |

The selected-unit reference is D2Client global RVA0x11bbfc; priority equality can include null equality, while a reached RNG draw requires a valid seed-bearing selected unit. Prior complete setter/packet/lookup readings0xa6640,0xad6e0,0xa4e20 and assignment sites support selected-client-unit ownership, not a universal player-only rule or guessed initial seed. Wider seeding/removal lifetime retains STP-Q01 and NRF-Q08.

Sorting uses unsigned-byte priority descending, finite distance key ascending, unsigned start descending, then fixed pool index descending. Deferred removals run after all visits. The entering coordinator record supplies pre-admission fade/radius; admission's entering base supplies history; the selected/current record supplies post-attempt Loop/Async retention and playing Duration. Their phase identities are not interchangeable. The audio counter used for record access and the coordinator tick argument are distinct supplied values; the dispatcher normally passes the entering counter without proving a universal frame or millisecond cadence.

### Original Sound import bindings used by the lifecycle

| D2Client thunk where established | D2Sound export / target RVA | Contract role |
|---|---|---|
|0xd348 |#10005→0x7650 |Prepare; reuse/recreate normalization. |
|0xd336 |#10006→0x6980 |General stop wrapper; attached stream adds its inner teardown Stop before the unconditional Stop. |
|0xd30c |#10032→0x6850 |Stream bookmark query, returned byte position divided by 4. |
|0xd2e8 |#10037→0x6f30 |Buffered repeating Play with signed-result marker policy. |
|0xd2be |#10044→0x6950 |Retained marker getter, not an OS status query. |
|0xd2ee |#10003→0x67c0 |Attached-stream predicate. |
|0xd330 |#10050→0x67f0 |Buffered loop-start setter. |
|0xd324 |#10059→0x9b20 |Retained routing-mode getter. |
|0xd270 |#10001→0x9af0 |Stored client master setting. |
|0xd26a |#10011→0x9ac0 |Stored music setting. |

Original Sound exports additionally bind#10038→0x6820 logical Loop setter,#10039→0x66f0 stream wrapper,#10043→0x6680 direct stream-wrapper end,#10069→0x7580 buffered load,#10051→0x6ce0 property-factor setter,#10008→0x6e20 backend amplitude conversion. List bindings retain the earlier public record: ordinals 10034/10018/10062/10063/10053→0x8710/0x8640/0x86c0/0x8920/0x86a0. Node allocation/free is distinct from list ordering and must be included when a case claims an exhaustive heap log.

Routing state read by0x9b20 is Sound RVA0x164a0, distinct from acoustic-conversion mode0x16598. The original operand and PE export resolve that distinction despite misleading generic labels. With no writer, repeated getters expose the same retained value; a cache-hit property visit can still read routing mode before returning. Do not turn a bypassed admission class getter into a bypassed full property-update getter.

### STP-R04b navigation: assets, asynchronous ownership and backend

| Original module / RVAs | Root-to-output chain and dated evidence | Limits retained |
|---|---|---|
| Client0x5ebc0,0x5ea90,0x5ed90 | Load admission, wrapped eviction target, pending throttling/scan; complete source and critical native paths rechecked2026-10-31. | Synthetic files and valid record storage; no original-data-table default inferred. |
| Client0x5da80,0x5dd50,0x5dae0 | Pending finalization, disposal and access/LRU. Ready descriptor finalization decrements pending; disposal does not. | LRU-node frees differ from payload/descriptor frees; retained failure/size/parsed metadata remain explicit. |
| Client0xd810,0xd610,0xd520;0x5cd10 | Synchronous allocation adds 800 unaccounted bytes; exact read required; loaded format validation. Complete adapter/parser bodies read2026-10-31. | Invalid storage/read/size can terminate; diagnostic/CRT foundation not replaced by clean return 0. |
| Sound0x8980,0x8790,0x7580,0x72d0 | Loaded WAV parser→load wrapper→initial fill; prior2026-10-30 parser evidence reused and2026-10-31 returning/fatal/failure branches audited. | Parser-null can terminate; valid negativeLock is the normal-return discriminator. Stream parser has a separate acceptance boundary. |
| Sound0x7650→0x69f0/0x7090;0x6980 | Prepare reuse versus Stop/release/clear/recreate; checked Sound factory HRESULT; repeated inner stop paths. Full rereads/native check2026-10-31. | Spatial interface/property capability paths consume S3D integration; no clean null-output continuation invented. |
| Sound0x7800→0x7230/0x6980,0x6a40,0x72d0 | Ordered50ms-timeout worker, recovery, buffered completion, half clearing, refill and backend millisecond fades. Full applicable bodies read2026-10-31. | Exact wait/status/Restore/Lock/region/clock providers; no generic device success or universal retry bound. |
| Sound0x6ce0,0x6e20,0x8830,0xaccc | Property factor reset, retained requested amplitude and ordinary conversion/clamp route; complete source for named finite cases2026-10-31. | S3D-R06 ordinary arithmetic boundary only; nondefault acoustic/CPU/FP and full update remain S3D-R07/Q02. |
| Sound0x66f0→Storm#255/0x2a650→0x29df0 | File open before Sound lock; stream begin, RIFF/fmt/data scan, capacity/start/ownership publication. Full begin/scanner reread/native check2026-10-31. | No local PCM-tag/rate/channel compatibility gate; generic format/device validity is not inferred. |
| Storm0x29560,0x27b80,0x26930,0x25dc0,0x266c0 | Worker bootstrap, initial/refill/EOF pump, copy/start callback, physical-to-cache query and live lookup. Full source/native corroboration2026-10-31. | Bootstrap/Play/Stop result nonpropagation is engine policy, not successful device/thread proof. |
| Storm0x27a70,0x27250,0x279f0,0x281a0,0x271d0,0x29200 | Request deadline/ref publication; complete selector/affinity/priority merge/throttle and I/O completion policy. Selector body ends0x273ef; following unrelated body excluded. | This is owned queue policy, not an unexplained external callback. Extra competition/nonzero-bandwidth vector coverage and final review remain. |
| Storm0x27ef0,0x28370,0x4290/0x4250/0x4d60/0x4390/0x4b80 | Queued cancellation, retained notification nodes, signal-before-node-free consumer and list effects. Fresh full readings/native check2026-10-31. | Queued cancellation is not current-read cancellation; failed signal is consumed without retry. Heap internals remain supplied. |
| Storm0x4040,0x3fc0,0x26e20,0x3ef0,0x3e20 | Request drops file then stream refs; final stream drops file before conditional Release; file/archive owner teardown. Full applicable source/native coverage2026-10-30/31. | Generic allocator and archive-resource foundation retained; borrowed buffer survives stream destruction. |
| Fog0x1dcc0,0x1d870,0x1dc40,0x1e170,0x1d740,0x1e010,0x1e240,0x1da20 | Descriptor create/event pool/poll/wait/finalize/destroy/getters;2026-10-30 complete readings reused and full applicable bodies/fresh import/native corroboration2026-10-31. | Event pooling, read-Boolean nonpropagation and checked wait versus unchecked destroy wait remain separate; no I/O-success inference from readiness. |

Fog original exports#10091/#10092/#10093/#10094/#10095/#10097 resolve respectively to0x1dcc0/0x1dc40/0x1e170/0x1e240/0x1da20/0x1e010. Finalizer0x1d740 and event allocator0x1d870 are internal helpers. Fog#10102/#10103→0x17e40/0x17e30 forward open/close. Fog's original imports bind Storm#267→0x28da0 open,#265→0x262d0 size,#271→0x26030 seek,#289→0x295b0 read,#282→0x281a0 priority,#283→0x28370 cookie cancellation,#253→0x26e20 owner close. Fog async-enable/event-count/event-base are actual module RVAs0x4af68/0x4af6c/0x4af70; an earlier shorthand using the absolute-address suffix was not an alternative semantic input.

The direct read chain0x295b0→0x26ed0→0x26d70 reaches actual KERNEL32 read/seek/size/error results. Complete resource-routing source was read; the stream-specific independent audit initially covered the direct branch only, while the descriptor continuation reused/read the broader routing body. Archive routing0x28da0/0x28960, block read0x28e80/0x28450 and cache0x28df0 were pursued in the dated resource reading. Lower MPQ decryption/decompression and error/dialog helpers remain separate subsystem boundaries; a full audio body read is not a new codec implementation proof.

Sound worker window sampling resolves through D2Gfx#10048/RVA0x7fb0 and USER32 IsIconic. Only the original getter and supplied window/state are used here; broader window creation is not audited. A separate Sound playlist uses0x7d30, enable setter#10045/RVA0x7d00 and a dedicated wrapper created by0x7cd0→0x7740. It can share the worker list. Shipped enable imports in Launch/Multi/Win and the Win options callback were identified; this does not prove Client can never enable it indirectly. Music selection stays excluded, and disabled-playlist/singleton-list inputs in worker cases remain explicit. Sound's playlist random helper0x11b8 uses its own per-thread CRT state via0x18fe, not the selected-unit audio seed.

### Dated static corrections and why the examples changed

The2026-10-31 corrections below resolve source/contract or supplied-fixture defects. They are not observed gameplay mismatches and do not retroactively change the meaning of the earlier limited observations.

| Static finding | Correction rationale |
|---|---|
| No-candidate duplicate branch could read as incoming retirement | Original0x88dd0 continues admission on null; only a present candidate without Stop Inst retires incoming. |
| Request offset conflated with saved resume bookmark; pool exhaustion omitted |0x88a70 writes the request offset separately;0x87d80 can returnnull when all 200 entries are occupied. Fresh nonzero-offset/exhaustion examples distinguish them. |
| Marked search implied first slot or retirement-prefilter |0x5cc24 chooses oldest unsigned start, then caller checks retirement without fallback. History writes expose the difference even when admission later fails. |
| Ignored parser/fill results treated as guaranteed malformed-input success |0x7580's normal epilogue returns 1 after returning fill failure, but preceding length/null-data diagnostics can terminate. Valid-WAV negativeLock remains the intended discriminator. |
| Stream EOF described as successful physical stop |0x27b80 records finished after requesting Stop regardless HRESULT. One direct stream end, two ordinary stream Stops, and three EOF-plus-worker Stops are distinct source paths. |
| Initial alignment generalized to whole-chunk advance | Special first aligned request advances by its trimmed count with no delay increment; later requests advance whole chunks. |
| Callback copy count assumed for zero-byte worker read | Worker dispatch checks actual count and current destination; zero-count positive-length requests Stop/finished rather than callback padding; cleared destination skips both. |
| Buffer reuse assumed to repair every field | Prepare's reuse branch retains marker/half/end/fade; recreation clears them, and its checked-HRESULT factory differs from Storm's output-pointer gate. |
| Alias/override, event/error and heap-log ambiguities | Explicit independent resets and producer carry; LRU-node free included;0x85100072 retained as hexadecimal; request scratch storage retained while returned bytes may change. |
| Unnamed stream clock array | Replaced by explicit constant 1000 for each reached construction/selection/worker clock in serialized segments; deadlines 1000/3600, unchanged nonclock outcomes. This is a synthetic input repair, not a runtime fit. |
| Routing mode conflated with acoustic mode or full cache proof | Separate stored inputs and exact reached getter calls; full acoustic-update effects remain an explicit S3D-R07/Q02 supplied dependency. |

### Synthetic case index

The41 main source-case labels D1–D41 map to STP-V(27+n), hence STP-V28–STP-V68. Independent variations do not increase the main-case count. Existing STP-V01–STP-V27 remain separately dated. All41 additions have runtime_observed:no; their expected outputs were recomputed from original source and supplied providers, not run in a game or executable harness.

| Source case | Proposed public vector | Partition | Evidence discriminator |
|---|---|---|---|
| D1 | STP-V28 | Deferred replacement | Incoming Play can precede the deferred old Stop; history and LRU order survive. |
| D2 | STP-V29 | Incoming deferral | Candidate without Stop Inst retires incoming; early deferral bypasses grouping even with request flags zero. |
| D3 | STP-V30 | Loop restart | Stopped Loop/Duration0 resolves transition then waits for another coordinator invocation. |
| D4 | STP-V31 | Duration boundary | Unsigned age equal to Duration retains playback; greater age stops; later cleanup is distinct. |
| D5 | STP-V32 | Logical fade and retirement | Zero-target transition completion follows the stop prepass; backend stop occurs on the following update. |
| D6 | STP-V33 | Failed automatic fade | Admission failure retains zeroed volume without installing the transition. |
| D7 | STP-V34 | Pending async retention | Async-only nonloop retention requires the pending asset state, then finalization permits a later admission. |
| D8 | STP-V35 | Negative initial Lock | Valid parsing plus returning Lock failure preserves PCM/counters while the outer load epilogue returns 1. |
| D9 | STP-V36 | Looping stream | Creation, initial request, callback threshold, EOF reset/refill, cached-query failure and borrowed teardown. |
| D10 | STP-V37 | Request creation/reuse | Null emitter equality, repeated unsigned-byte priority boost, compound reuse and retained initialization. |
| D11 | STP-V38 | Owner removal | Remaining references defeat force; last-reference loop retirement is a separate effect. |
| D12 | STP-V39 | Invalid schedule | Zero-target retirement write precedes the invalid delayed-zero-duration fatal path. |
| D13 | STP-V40 | Wrapped asset budget | Wrapped eviction target permits a load beyond the nominal budget; allocation and accounting differ. |
| D14 | STP-V41 | Descriptor disposal | LRU-node, payload and descriptor free order; disposal does not balance the client pending counter. |
| D15 | STP-V42 | Pending loader tick | Tick0 bypasses interval/pending/backoff guards; a nonzero tick can be throttled. |
| D16 | STP-V43 | Sticky asset failure | Invalid loaded format retains accounted storage; disposal retains failure and blocks a reload. |
| D17 | STP-V44 | Negative Play | Loaded payload/history/cache can persist while failed start clears tentative association and removes incoming. |
| D18 | STP-V45 | Nonloop completion | Drain to finished, Sound worker three-Stop sequence and retained-buffer handoff. |
| D19 | STP-V46 | Current-read stream end | Destination clearing, retained request/file/stream references and later destruction order. |
| D20 | STP-V47 | Logical marker bridge | Marker0/nullstream detaches without bookmark query/Stop; following coordinator removes logical ownership. |
| D21 | STP-V48 | Buffered completion | Played equal to source length retains marker; strict greater-than clears it. |
| D22 | STP-V49 | Backend fade milliseconds | Signed interval, before-start endpoint, exact-end interpolation and retained active flag. |
| D23 | STP-V50 | Lost-buffer recovery | Special Restore retry sleeps 10; other negatives retry immediately; recovery stops without Release. |
| D24 | STP-V51 | Half clearing | Toggle despite negative Lock, clear before Stop, and returned first-region length. |
| D25 | STP-V52 | Buffered refill | Threshold, ignored fill return, end-marker padding, played>written and loop-start source wrap. |
| D26 | STP-V53 | Fresh request offset | Playback offset 123 is distinct from the still-zero saved resume bookmark. |
| D27 | STP-V54 | Voice-pool exhaustion | All200 entries in use returns 0 without counter/list/association changes. |
| D28 | STP-V55 | Duplicate selection segment | Special range replacement, replacement-to-null and lookup independent of duplicate flags. |
| D29 | STP-V56 | Ownerless Defer exception | Nonzero position bypasses duplicate actions; allzero position follows incoming retirement. |
| D30 | STP-V57 | Queued cancellation notification | Notification transferred before request destruction, then consumed by the worker before pumping. |
| D31 | STP-V58 | Owned-buffer destruction | File-reference drop precedes conditional Release and stream free; external owner remains. |
| D32 | STP-V59 | Initial aligned scheduling | Special first trimmed request differs from ordinary chunk/delay accumulation. |
| D33 | STP-V60 | Post-read zero count | Zero actual count is Stop/finished rather than callback padding; cleared destination skips both. |
| D34 | STP-V61 | Oldest marked suppression | Minimum unsigned start tick and retirement-after-selection, not first slot or retirement prefilter. |
| D35 | STP-V62 | Stream creation factory | Capacity mask/minimum, format metadata, pointer-only factory decision and owned versus borrowed release. |
| D36 | STP-V63 | Pending in-flight disposal | Current raw read, successful wait, pooled event/file lifetime, and unchanged client pending count. |
| D37 | STP-V64 | Periodic producer | Owned counter-modulo3 draw precedes a zero-Volume request rejection; nondivisible counter consumes no draw. |
| D38 | STP-V65 | Prepare reuse/recreation | Reuse retains marker/half/end/fade; recreation stops/releases/clears and checks factory HRESULT. |
| D39 | STP-V66 | Ordinary streamed stop | Bookmark query precedes two ignored Stops; marker 0/attachedstream counterpart differs from nullstream bridge. |
| D40 | STP-V67 | Coupled group/coordinator | Actual ordered draws, entering history/fade/radius versus selected Loop retention and no rollback on slot failure. |
| D41 | STP-V68 | Sort and LRU | All four comparator keys affect later visit/LRU order; cache-hit acoustic results remain conditionally supplied. |
| D42 | STP-V69 | Queue policy | Four selector classes, predecessor cookie, unsigned affinity, priority merge/metadata retention and nonzero-bandwidth dispatch-start waits. |

The synthetic cases and independent variants are bounded rather than an exhaustive branch proof. D28 is a candidate-selection segment; D32 is a scheduling stage with named request-construction dependency; D33 begins after the read count is already stored. D18's full producer state carries into D20; D39 consumes the same audited post-Q1 producer checkpoint without a common backend reset. D30 joins its cancellation output to notification consumption; D36 supplies the precise current-read/wait interleaving. No smaller stage is promoted into a whole unbounded worker/coordinator execution.

D37 and D40 contain actual source-derived draw sequences under NRF-R01/R02; other examples with no draw retain empty ordered draw logs. Rejected or failed requests do not roll back earlier seed/history writes. Synthetic files/records and supplied API return shapes are not claims about original MPQ contents or normal platform return distributions.

### Remaining boundaries and acceptance

- STP-Q01: wider selected-unit creation/reseeding/removal lifetime, automatic cleanup caller routing and complete original-table/runtime correspondence remain; no guessed initial seed or player-only identity is introduced.
- S3D-R07/Q02: the complete acoustic-update/cache/base-scaling pipeline, tracking/reference eligibility, option/duck mixing, nondefault panning/property routes, exceptional/fractional floating behavior, actual floating-environment initialization and CPU workaround are not closed by finite endpoint/cache-hit STP cases. Verified S3D-R02/R04/R05/R06 subcalculations retain their own exact bounds. A supplied full-acoustic profile is not ownership transfer or published proof of R07; STP must retain this integration dependency.
- STP-Q02: original queue/selector/merge/throttle policy is owned source behavior and is now explicit, not classified as an OS provider. Generic heap, OS/device, file-integrity/error, archive decryption/decompression and enabled-playlist integration retain their separately named foundations/limits. A source-level normal return never proves successful I/O, physical stop, safe failed-wait lifetime or audible playback.
- STP-Q03/STP-O01 and S3D observation gates: no additional complete original-game case was collected on2026-10-31. The earlier menu-stream and separate stop/callback observations remain field subsets; differing source names and reused numeric handles still forbid a manufactured full lifecycle join. Complete loaded-module/effective-data identities and matching input/draw/device/ownership logs remain absent where already recorded.

### Loader-field and assembly checks — 2026-10-31

Complete D2Client RVA0x5dfd0 table-loader source, its native postpass and original PE strings bind record+0x54/+0x58/+0x5c to Sounds columns Block 1/Block 2/Block 3 (original string RVAs0xd5598/0xd5590/0xd5588). For rows 1 through count−1 the postpass counts consecutive values before the first DWORD−1, capped 3, and stores the count at record+0x80. Admission0x5d7a0 tests count==1 and passes low32(Block 1*2) through Client thunk0xd330 to Sound#10050/RVA0x67f0, which stores wrapper+0x44 loopStart. Sound#10038/RVA0x6820 instead stores wrapper+0x40 logical Loop. These two descriptions were corrected after matching-source and original import/consumer checks; the semantic rules did not swap their operations.

The shared semantic all-zero blocks now explicitly derive count 3, while an earlier separately synthesized post-loader count 0 also skipped the equality-to1 branch. Neither is an archive-default claim. STP-V35 adds Block[1,−1,0]/count 1/loopStart2 with valid PCM length 4 and negative Lock: the setter precedes Load, strict 2<4 passes, and no copy-dependent outputs change. This discriminates the derived count and alias without inventing independent loop-start state. The loader/consumer/thunk/setter corroboration compared 3174 original bytes, zero missing/different, overlapping other source sets.

D2Client0x88dd0/0x5dc70/0x5dce0 read Range/Falloff as a full unsigned 32 selector; byte truncation was removed from the draft inventory, preserving S3D-R02's domain. Client global RVA0x11c280 controls marked-slot search and the routing/crossmix subsection inside0x5d1d0; zero skips to cache comparison, not a whole-property return. The coordinator/admission still invoke that dependency at their reached stages. The neutral shared control name states those roles; no false caller guard was added.

Successful automatic fades in0x88dd0 directly write voice transition state: active 1, current/start volume 0, saved endpoint, start tick=coordinator argument and end tick=low32(argument+chosen entering duration). They do not call0x88230 or reapply the selected record's minimum. The successful STP-V67 variation uses entering Fade In5, selected Fade In20 and a compatible slot, giving interval 100..105 with current 0, matching the explicit conditional acoustic profile. Failed admission keeps the earlier zeroing without installing that transition. The device arguments/counts and LRU/access effects are static assertions; neither variation was executed.

The original selector/worker/list/priority/merge bodies were checked against 2208 original bytes with zero missing/different bytes; this overlaps prior coverage. These bodies establish the predecessor-cookie condition, four selection classes, unsigned affinity wrap, dispatch-start throttle and merge gates/copy-retention rules. The earlier completion-metadata reference must be null for merging; a nonnull metadata object with a null callback does not pass. These static checks do not establish runtime execution.

### STP-V69 queue-policy cases — 2026-10-31

Original Storm selector 0x27250 through its own return/alignment boundary 0x273ef, worker 0x29200, constructor 0x27a70, priority wrapper 0x281a0/internal traversal 0x279f0 and merge 0x271d0 were read completely after C-first registry lookup. List helpers 0x4390/0x4b80 identify the immediate predecessor rather than guessing from encoded links. Original native bytes corroborate the merge metadata-reference gate and all copied/retained fields. The four selector classes, signed/unsigned ordering, below-base score, throttle formula and waits are source-derived policy, not generic external providers.

Arithmetic independently recomputed: base 4096; below score 134217733 versus competitor 134215680 and incorrect variant 134213637; throttle 200, defer 100, equality 200, dispatch-start elapsed 550 versus completion-time 50. The static continuation-aware native comparison for selector/priority/merge/worker/construction ranges checked 1888 bytes with zero missing/mismatch. This is not execution of these vectors.

The partitions intentionally end at selector return, priority-update return or a named pre-read/wait-exit checkpoint. Actual source I/O, completion callbacks, device behavior, unbounded thread scheduling, archive codecs and generic heap internals remain separately stated dependencies. Other merge size/zero gates, wrapped deadline domains and nonzero-bandwidth arithmetic extremes remain additional coverage opportunities; they are not silently converted to observed branches. These source cases do not establish gameplay observation.

The source-derived base cases contain 17 selector/merge/throttle partitions:10 selector results,3 priority/merge variants and 4 worker checkpoints. A further null-cookie exported-wrapper variation was checked from original Storm0x281a0: error 87/SetLastError(87), no traversal/lock/wake and no contracted scalar return. Two additional merge-cap cases use sums 1048576 and 1048580, accepting equality and rejecting the above-cap sum under original Storm0x271d0. STP-V69 therefore contains 20 independently reset partitions. Source comparisons overlap prior counts and are not executable results.

The Sound0x6850 matched source and original Storm0x266c0 lookup additionally establish STP-V66's finished-stream variation: failed lookup returns bookmark 0, no physical query/cache update, error0x85100072, followed by the same two-Stop cleanup. Misleading source comments naming a direct device query were not used as behavior. No new runtime credit follows.

## Additional synthetic assertion boundaries — 2026-10-31

STP-V39's terminationCode4294967295 denotes the unsigned code argument supplied to the reached nonreturning termination operation, not an observed process exit status. The neutral terminationRequestControl is supplied as 0 and becomes 1 before the ordinary termination request. Empty callback sets, normally returning diagnostic/synchronization operations and an absent managed-runtime module are explicit case inputs. The original Client0x88230 →0x3664 →0x35a1 →0x34dd chain and original imported termination operations establish this bounded result; generic shutdown bookkeeping and returning termination providers are outside it.

STP-V64's five independently reset supplied-input cases distinguish the unsigned current audio counter from the coordinator tick argument. Counter 3/tick 4 draws; counter 4/tick 3 does not. Counter 2147483649/tick 3 is divisible by 3 in its stated unsigned 32-bit domain and draws. The operation projection and read-only first-master-read seed snapshot assert the producer-before-master-read order already established by Client0x88dd0 and its producer. They change no engine state or RNG behavior and do not establish that unequal counter/tick inputs occur through the normal dispatcher. These are synthetic static assertions, with no new runtime credit.
