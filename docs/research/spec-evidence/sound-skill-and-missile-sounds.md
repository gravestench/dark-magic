# Skill and missile sound evidence

Spec: [Skill and missile sounds](../../specs/sound-skill-and-missile-sounds.md)

## Sources and identity

Evidence date 2026-10-29, original Expansion 1.13c Windows x86.
[Manifest](../source-manifest.md) artifacts `LOD113C-D2CLIENT` and
`LOD113C-D2COMMON`. Both original inspected images were rehashed and matched
the manifest: D2Client SHA-256
`dd8bc6025de921216a97c17f97cd1a50fbb85926e838ec60e13451448836d906`;
D2Common SHA-256
`59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10`.
Candidate names are analyst labels, not original exported names.

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2CLIENT | 0xa08e0 | none established | active state dispatcher | SMS-Q01 | disassembly, independently corroborated native listing; integration dependencies open |
| LOD113C-D2CLIENT | 0x14f40 | none established | type-specific skill cue handler | SMS-R01, SMS-R02 | complete disassembly; no encoded instruction runs; only callee is gated sound-request boundary |
| LOD113C-D2CLIENT | 0x40a70 | none established | progressive sound and overlay handler | SMS-Q02 | disassembly read; state/stat and overlay dependencies unclosed |
| LOD113C-D2CLIENT | 0x88a70 | none established | common sound request | SMS-Q04 | shared playback evidence; admission and audible integration unresolved |
| LOD113C-D2CLIENT | 0x14fd0 | none established | chain target update | SMS-Q03 | disassembly read as candidate lead; it calls lookup at 0x2be0 and unit sound selection, not a proved missile travel/hit lifecycle |
| LOD113C-D2COMMON | 0x61ec0 | none established | skill record descriptors | SMS-R02 | descriptor reading; complete generic name linking not established |
| LOD113C-D2COMMON | 0x7c8d0 | none established | missile record descriptors | SMS-Q03 | targeted descriptor search; consumers not closed |

## Dispatch identity

The original D2Client data table at RVA 0xdea48 has handler slot 13 at
0xdea7c containing 0x40a70, and slot 16 at 0xdea88 containing 0x14f40.
Absolute-pointer scanning of the matching image found these entries;
neither candidate had a direct relative-call reference.

D2Client 0xa08e0 iterates an active-state list, tests membership, obtains
the state record (stride 0x3c), reads its signed word +0x34 and bounds-checks
that handler index against the count at RVA 0xdec50. State +0x34 is
`cltactivefunc` under [GDA-R13](../../specs/core-game-data-table-records.md).
The call at 0xa09bb dispatches through table 0xdea48. It supplies the unit,
an argument obtained from stat 350 when an associated list is present
(otherwise 0), and a second argument obtained from stat 351 (otherwise 1).
The complete state/list producer and scheduling path remains SMS-Q01.
This is a table dispatch root for the handlers, not proof that every skill
uses either handler.

## Bindings

The two candidates both read the Skills array at DataTables +0xb98 and
count +0xba0, stride 0x23c. They are not identified as missile operations
merely because an analyst label calls one a missile-state sound routine.

D2Common 0x61ec0 builds linked-field descriptors naming original strings:
`dosound a` at RVA 0x90b84 maps to Skills +0x102; `dosound b` at 0x90b78
maps to +0x104; `prgsound` at 0x90b30 maps to +0x10e; `prgoverlay` at
0x90b24 maps to +0x10c. These are descriptor mappings to consumers, not
archive cell evidence or a claim that name-link resolution is complete.
The descriptors' sound-link context is separate from the handler's
signed-word/positive test, which is read directly in client code.

At D2Client 0x14f40, the input unit is ECX and skill id EDX. Unit +0x4e is
tested against 3 before any skill lookup. The field's producer remains
unresolved, hence the contract's neutral presentation selector. Negative
or out-of-bounds skill id and absent record return zero. Unit dword +0
selects first cue only for 0, second cue only for 1. Both signed words are
tested with `jle` before requests. At 0x14f92 and 0x14fb2 the handler passes
unit plus three zero stack values with EBX holding the selected cue to
0x88a70, then discards its result. The final return is one.

D2Client 0x40a70 reads prgsound/prgoverlay, conditionally obtains a state
list/stat result, clamps that result into 1..3 and offsets both ids by the
clamped result minus one before overlay/sound requests. Because the flags,
lookup, stat and overlay execution dependencies remain unclosed, no
progressive behavioral rule is promoted from this reading.

D2Common missile loader 0x7c8d0 contains original descriptor strings
TravelSound (RVA 0x8b258), HitSound (0x8b24c), ProgSound (0x8b240).
Targeted searches found their descriptor construction, but this record
does not claim complete consumer or lifecycle tracing. State onsound and
offsound fields are established by GDA-R13; their request consumers remain
SMS-Q03. Column existence does not establish when playback starts or stops.

## Search limits and reuse

Existing missile creation, initialization and movement contracts were read
as server boundary references. Existing game-data descriptor evidence was
checked for the Skills/States contracts; its actual recorded labels and
limits are retained. No observation of server missile lifetime is reused
as evidence of client audio lifetime. The audio research document's 1.14d
travel-loop observations and proposed event routing are leads only.

Matched reconstruction availability was checked before reading candidate
disassembly; no matched reconstruction was available for these two client
handlers. A candidate sound-index lookup at 0x2be0 was available only as
unverified pseudocode; it was not used as behavioral proof. The full
0x14fd0 reading showed why its sound-index call is not alone an anchor for
the requested missile audio lifetime. Static searches and original-image
table-slot checks anchored the active-handler path. This is an incomplete
search for the remaining stsound/dosound, missile and state producers,
not a six-rung absence claim.

## Producer and request-stage continuation

New readings dated 2026-10-29 retain the same manifest identities. The matched registry was checked before the new client assembly readings: no matched C for 0x73630, 0x73760, 0xBB470 or 0xBB5B0; complete assembly bodies were read. D2Common 0x2EE10 has matched C and was read first. Client 0xB30F0 was read through its final creation/request tail; full preceding creation dependencies are not credited to its supplied tail. The new boundaries end at logged sound/overlay requests, not device output.

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2CLIENT | 0x40a70 | none established | progressive handler | SMS-R03a | complete assembly; supplied record/list/stat and overlay/request operations |
| LOD113C-D2CLIENT | 0x73760, 0x73630 | none established | state on/off stages | SMS-R04a | complete assembly; supplied membership/definition, sound-decision portion |
| LOD113C-D2CLIENT | 0xadd70, 0xadd30, 0xadeb0 | none established | state-message entry leads | SMS-Q03 | first two wrappers read; third incoming reference only |
| LOD113C-D2CLIENT | 0xb30f0 | none established | successful missile creation tail | SMS-R04b | local request tail; preceding creation and later owner retirement gated |
| LOD113C-D2CLIENT | 0xbb470 | none established | missile progression cue | SMS-R04c | complete assembly; counter accessors read, following missile simulation boundary excluded |
| LOD113C-D2CLIENT | 0xb1420 | none established | admitted missile-hit cue | SMS-R04e | hit-handler admission and request tail read; collision/callback producer integration gated |
| LOD113C-D2COMMON | 0x69f90, 0x69fb0 | 11018, 10637 | missile previous-progress accessor/setter | SMS-R04c | complete matched C; optional type-3 data guards and counter read/write |
| LOD113C-D2COMMON | 0x2ee10 | none established | frame marker writer | SMS-Q01 | complete matched C; frame byte 1..4 copied to unit marker |
| LOD113C-D2COMMON | 0x32460, 0x2f100 | none established | frame/sequence marker production | SMS-Q01 | assembly read; native decoding corroborates ordinary frame-marker writes, full sequence dependencies open |

### Progressive selection

GDA-R11a's schema identifies Skills +0x04 bit 2 as progressive. The tested mask byte in 0x40A70 selects that bit; source descriptor reading independently binds Skills +0x80 to aurastate and +0x54 to `aurastat1`. The handler validates both signed ids against States/ItemStatCost counts before associated-list lookup and a layer-zero stat read. Client imports resolve to D2Common #10871 / 0x38520 for state-list lookup and #10680 / 0x39140 for list-stat reading; supplied results are explicit in SMS-R03a, not evidence that their producers execute. Existing RLI-R03 and RLI-R02a/RLI-R02b describe those operation families; this continuation does not renew their verification dates.

The clamp uses signed comparisons: <=1 becomes 1, >=3 becomes 3. Sign-extended prgoverlay/prgsound each receive clamped-minus-one. At 0x40B2E, overlay operation 0x6B300 precedes sound operation 0x88A70. There is no positivity check on either adjusted identity in this handler, and neither return is tested. The overlay operation and actual voice mutations remain outside the supplied selection stage. The original handler dispatch table slot 13 is the recorded root anchor for this stage.

### State sound edges

0xADD30 decodes a unit identity and state byte from its supplied message, finds the unit, calls 0x720B0, then 0x73630. 0xADD70 finds the unit and calls 0x73760 before applying following transmitted stat data. These are concrete incoming roots; message command identity and complete network validation remain unresolved. 0x73760 tests the States notondead flag (GDA-R13) against monster mode 12 / player mode 17 before its membership check. Existing membership takes the list/refresh path without an on-sound request. Absent membership reaches signed onsound +0x26: only negative values skip 0x88A70. It then requests membership set. 0x73630 tests existing membership first; when present, it may detach an associated list and then reads signed offsound +0x28; only negative values skip the request. Its clear-membership request occurs afterward. Membership/list/overlay mutations surrounding the sound stage remain full-integration questions, not omitted verified effects.

### Missile sound requests and retirement leads

Missile descriptor reading binds TravelSound +0x12, HitSound +0x14 and ProgSound +0x34. Creation 0xB30F0's successful tail follows its optional callback, sign-extends TravelSound and requests any nonzero identity from the new missile. It then optionally searches the owner for fixed cue 314 through 0x225D0 and invokes 0x88BF0 on a found handle. Both helpers were read, with 0x88BF0 available as matched C; their sound-record/group and device effects remain playback integration. This is not evidence that travel sound ends at contact.

The progression handler 0xBB470 uses ProgSound only when signed-positive. Original-image pointer scanning finds it at RVA 0xDD95C, slot 47 of the client missile-do table at 0xDD8A0. The tail of 0xB1AF0 reads the signed selector at missile record +0x08, requires it positive and below the count at 0xDDA0C, and jumps through that table. This anchors the supplied progression stage to a real missile dispatch root; effective rows selecting 47 and complete frame scheduling remain integration gates. Import thunk 0xC322 resolves to D2Common ordinal 11018 / 0x69F90, which reads previous counter at missile data +0x28, returning zero for missing unit, non-missile type or missing inner data. Current counter comes from unit +0x54 → context +0x30 → field +0x40, with zero for absent context. A nonzero previous value different from current requests ProgSound. The handler then calls thunk 0xC448 → ordinal 10637 / 0x69FB0 to store current, regardless of the request return; that setter has the same type/data guards. Both imported helpers are matched C and were read. The tail jumps to 0xB92B0; its complete missile simulation remains outside the sound-decision stage.

At the missile-hit path 0xB1420, the admitted request branch at 0xB15E1 reads signed HitSound +0x14 and skips only negative values. Its preceding collision gates and optional hit callback can return before this branch; supplied admission in SMS-R04e does not claim all contacts reach it. A later path reads TravelSound, searches 0x225D0 for an associated voice and requests 0x88C20 when found, subject to additional target/record conditions. The later path's exact stop semantics and missile disposal are still SMS-Q03. Another specialized progression reader at 0xB8F00 has an end-of-life counter comparison, 0x88C20 request and positive ProgSound branch; it is a separate lead, not collapsed into SMS-R04c.

### Animation marker provenance

Matched D2Common 0x2EE10 copies a frame-data byte 1, 2, 3 or 4 into unit +0x4E; other frame bytes do not overwrite it. Its frame source is unit +0x50 plus frame index +0x10. Ordinary animation update 0x32460 clears the marker and scans crossed frame bytes, writing 1..4; the relevant original instruction stream was independently decoded natively. Sequence update 0x2F100 calls 0x2E6E0 and writes that operation's returned event byte into the same marker. Thus the handler's presentation selector is an animation marker. Full ordinary/sequence progression, complete frame dispatch and timing remain SMS-Q01; no action meaning is inferred from marker value 3.

### New-stage self-review and limits

SMS-V11–SMS-V16 were recomputed from the new supplied-stage rules: charge 5 clamps to adjustment 2; charge -3 to 0; invalid adjustment gates keep base identities; state and hit stages forward zero but reject negative; travel forwards negative but rejects zero; progression writes current even when first-seen or unchanged, but skips writes when cue is nonpositive. None is a gameplay observation. All new rule labels are restricted to supplied inputs and logged effects. Fresh blind review is required for this expanded contract.

## Observation

SMS-O01 requires original Expansion 1.13c, single player, and a unit with
an independently identified active state selecting handler 16. Record the
unit type, its presentation selector, supplied skill id, record count and
the two cue values before dispatch. Record every sound request's arguments
and the handler result. Match SMS-V03 only when those inputs coincide;
otherwise compare the rule using actual inputs. A failed device request
can still yield handler success. Exact class/equipment/skill preparation
cannot yet be prescribed because the state/skill producer is SMS-Q01.
This is a blocked observation recipe, not a complete executable gameplay
procedure or performed observation.

No runtime gameplay observation or named table-cell read was performed for
this spec. The eligible state/skill preparation remains unresolved, and
runtime collection was unavailable during this pass. This limits observation
and does not establish a contradiction in static selection arithmetic.

## Review state

Self-review preserves the supplied-input boundary and gated producer/device
integration. Fresh blind reader review had not yet run at this initial revision.

Fresh blind reader full pass 1 reported no behavioral MISMATCH (0), with
no ungated behavioral GUESS. Exact other list counts were not supplied to
the writer and are not invented here. The field-qualification finding is
resolved by describing supplied signed-16 cue values and explicitly gating
actual linked-row binding under SMS-Q01/SMS-Q02; GDA-R11a is consumed for
record identity only. The SMS-V04 discriminator now distinguishes a signed
nonzero request of -1 from an unsigned positive request of 65535.
Advisory partition/hardest findings are addressed with SMS-V06 selector 4
bypass, SMS-V07 in-range absent record, SMS-V08 positive second cue,
SMS-V09 negative id and SMS-V10 zero cue. Each includes complete inputs,
empty draw specification and exact result/request-log assertions. A fresh
full reader pass is required after these contract/vector changes.

## Historical fresh blind-implementer review: full pass 2

2026-10-29: MISMATCH 0, GUESS 0, GAP 0 blocking, DISCRIM 0 unresolved (all 10 claimed discriminators distinguished their stated alternatives), GATED 5 questions, COVERAGE 1 grouped advisory item, HARDEST 3. The fresh reader recomputed every one of the ten vectors and found all outputs consistent with the supplied-input contract. It confirmed that live row linkage is not inferred from descriptor names and that GDA-R11a is used only for record identity/bounds.

The reader noted that no-lookup vector assertions require observing whether lookup was queried. Outcome no change: those vectors already explicitly assert no lookup, which is an acceptance observation of the supplied lookup operation, not inference from the sound request log. COVERAGE outcome no change for negative first cue and zero second cue: the shared strictly-positive signed test is explicit and zero/negative supplied values already have discriminating examples across the two fields. These remain advisory partitions, not runtime coverage. HARDEST outcomes: selector-before-validation, signed positive cue selection, and missing-record failure versus ignored playback failure are explicit and distinguished by V01/V06, V04/V10 and V07/V03/V08. All surrounding Q01 through Q05 gates remain open. The current supplied-input contract meets the blind-reader stopping rule; full skill/missile/state audio parity is not established.

That earlier revision's mechanical validation on 2026-10-29 had 0 errors and 2 warnings. It does not review the new SMS-R03a/SMS-R04a–SMS-R04e stages or SMS-V11–SMS-V16. Fresh full review of this continuation remains pending. Temporal field names describe current retained state, and source-reading method names describe public verification methods rather than private reconstruction paths. No implementation or user approval is asserted.


## Fresh blind-implementer full pass 3 — 2026-10-29

MISMATCH 0 across sixteen base cases and eighteen subcases (34 expected outcomes), GUESS 0, GAP 2 prose findings, DISCRIM 0 failed claims, GATED 5, COVERAGE 3 advisory clusters, HARDEST 3. All exact outputs and stated discriminators recomputed successfully.

GAP outcomes: section 6 now explicitly scopes the positive-only selected-cue test to SMS-R02, preserving the differing progressive/state/travel/hit policies already defined in their rules and vectors. Section 5 now explicitly scopes the no-gameplay-write statement to the type-specific handler and names R04c’s distinct retained client-counter write. Neither correction changes any rule, input, vector result, operation order, ownership or gate; the repeated generic prose had overreached its intended handler scope. A fresh delta reader will check these clarifications against V13 through V16 and their owning rules.

COVERAGE outcomes: no change for negative aurastate/`aurastat1`, flagged monster mode 12/off-negative state cases, or negative ProgSound. Signed bounds, state membership/nonnegative cue policy and nonpositive progression guard are explicit; current vectors distinguish the relevant request/no-request and store/no-store outcomes. These remain advisory unobserved partitions. HARDEST outcomes: no change for selector-before-validation, signed-positive cue selection, or first-seen/equality/missing-context progression with request-failure-independent store; V01/V06, V04/V10 and V15 explicitly distinguish those cases. Retain Q01 through Q05. No approval or full audio parity is inferred.


## Fresh blind delta pass 4 — 2026-10-29

Changed material: the two prose-scope clarifications identified in full pass 3; rules, vector inputs/results and gates unchanged. A fresh reader checked section 5/6 against the differing cue policies and recomputed every V13 through V16 case. MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0 on the changed material. Both prior gaps are resolved.

The checked set is seventeen cases: V13 seven, V14 two, V15 five and V16 three (four bases plus thirteen lettered subcases). The reader initially wrote nineteen, then confirmed the addition error and all seventeen recomputations without changing any behavioral finding. The current supplied-stage contract meets the reader stopping rule. Q01 through Q05 remain explicit; no runtime or complete audio parity claim is added.
