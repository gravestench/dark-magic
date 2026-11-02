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

The 2026-10-30 continuation below corrects the original preparation. Its table reads are data evidence, and its unsuccessful gameplay attempt is not a matching observation.

The original 2026-10-29 SMS-O01 recipe required original Expansion 1.13c, single player, and a unit with an independently identified active state selecting handler 16. That preparation is invalid for the original data, as established below; it is retained here as the disputed earlier recipe. Record the
unit type, its presentation selector, supplied skill id, record count and
the two cue values before dispatch. Record every sound request's arguments
and the handler result. Match SMS-V03 only when those inputs coincide;
otherwise compare the rule using actual inputs. A failed device request
can still yield handler success. Exact class/equipment/skill preparation
cannot yet be prescribed because the state/skill producer is SMS-Q01.
This is a blocked observation recipe, not a complete executable gameplay
procedure or performed observation.

At the 2026-10-29 revision, no runtime gameplay observation or named table-cell read had been performed. The eligible state/skill preparation remained unresolved, and runtime collection was unavailable during that pass. The dated continuation below supersedes the preparation and table-data statements. This limits observation
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

## Original data and producer continuation — 2026-10-30

Both original images were rehashed on this date and retained the D2Client and D2Common identities above. Complete local function bodies were read for the entries below, checking matched reconstruction availability first. Disassembly was used for entries without a matched reconstruction. Native independent decoding of the original images corroborated the encoded ordinary/sequence marker scans, missile-do dispatch and the encoded instructions in the start/ordinary/progressive skill dispatchers. Function reading does not credit unread callees or establish execution. These findings remain within SMS-Q01–SMS-Q05; they do not add a smaller verified behavioral stage or close full eligibility, cadence or lifetime questions.

| Artifact | Function RVA | Analyst description | Method and remaining boundary |
|---|---|---|---|
| LOD113C-D2CLIENT | 0xa08e0, 0x14f40, 0x40a70 | active-state dispatcher and shared sound callbacks | complete disassembly; callback selection is separate from an original row selecting that callback |
| LOD113C-D2CLIENT | 0xa1cf0, 0xa22b0, 0xa2370, 0xa30c0, 0xc9030 | skill start dispatcher and local callers | complete disassembly; generic skill activation, path/target and animation preparation dependencies remain |
| LOD113C-D2CLIENT | 0xa1b10, 0xa2950, 0xa2c90, 0xaefb0 | first progressive, charged progressive, ordinary and received progressive dispatch | complete disassembly; overlays, activation and complete message admission remain |
| LOD113C-D2CLIENT | 0x197d0, 0x4f2c0, 0xa3000 | presentation-frame to used-skill dispatch | complete disassembly; full frame root and animation initialization remain |
| LOD113C-D2CLIENT | 0x2b1e0 | per-unit presentation/removal dispatch | complete disassembly and native table-jump decoding; broader frame loop and animation initialization remain |
| LOD113C-D2CLIENT | 0xadd30, 0xadd70, 0xadeb0, 0x72610, 0x73760, 0x73630 | state messages, transmitted stats and on/off paths | complete disassembly; complete command validation, list lifetime and overlay effects remain |
| LOD113C-D2CLIENT | 0x72c80, 0x723d0, 0x74480 | surrounding state presentation and associated-skill readers | disassembly read; these are not demonstrated producers of fixed stats 350/351 |
| LOD113C-D2CLIENT | 0xb30f0, 0xb1420, 0xb1720, 0xb17c0, 0xb1af0, 0xb92b0 | full local missile creation, hit, removal, expiry and do-dispatch paths | complete disassembly; collision/path/registry ownership dependencies and audible retirement remain |
| LOD113C-D2CLIENT | 0xbb470, 0xb8f00, 0xbb5b0 | distinct progression sound consumers | complete disassembly; only 0xbb470 has the row/dispatch correspondence recorded below |
| LOD113C-D2CLIENT | 0x225d0, 0x88bf0, 0x88c20 | associated-voice lookup and retirement requests | 0x88bf0 matched reconstruction; others complete disassembly; full playback admission/device lifetime remain STP-R03b/STP-R04b |
| LOD113C-D2CLIENT | 0x88a70, 0x88350, 0x88550 | request creation, voice position write and retirement flag | complete disassembly; shared playback/position contracts retain ownership; capture interpretation does not establish the complete coordinator or device lifetime |
| LOD113C-D2COMMON | 0x34000, 0x33cd0, 0x37db0 | active-mask and state membership operations | complete disassembly; state/list creation remains separate |
| LOD113C-D2COMMON | 0x38520, 0x37f80, 0x39140, 0x38b00 | associated-list and layer/stat access | complete disassembly; consumes RLI-R03 and RLI-R02a/RLI-R02b for their recorded contracts and limits |
| LOD113C-D2COMMON | 0x2e6e0, 0x2f100, 0x32460 | sequence extraction and animation advancement | complete disassembly and native decoding; animation source/reset and complete client-frame scheduling remain |
| LOD113C-D2COMMON | 0x61ec0, 0x7c8d0 | Skills and Missiles field descriptors | targeted descriptor/name readings, not a complete generic text-link compiler proof |

The earlier 2026-10-29 readings of D2Common 0x2ee10 and 0x69f90/0x69fb0 remain matched-code evidence with their original scope. The state-loader list construction at D2Common 0x5f6e0, 0x5da30 and 0x5dad0 reuses the 2026-10-30 reading recorded by [unit sound selection](sound-unit-sound-selection.md); this does not renew unrelated dependencies. The active flag is bit 5, building the ordered state-index list and membership mask used by 0xa08e0. A reconstruction registered only for an initial short function is not credited for anonymous code following that function in the same disassembly container.

### Effective files and separate compiled correspondence

The table reads used the running original Expansion 1.13c game's Storm loader. The launch profile reported vanilla LoD 1.13c, with no override requested for this run. Physical supplying archives, an exhaustive preexisting override inventory and complete loaded-module fingerprints were not captured. The on-disk binaries were independently identified above. These limits prevent transferring the readings to another effective profile.

| Effective virtual path | Contiguous returned coverage / total bytes | SHA-256 / coverage |
|---|---|---|
| `data/global/excel/states.txt` | 0..20920 / 20920 | `898806c2b69e55e316ecfed48dde95aa2933829b94fdad234bdabc27bdbc23e1`; complete |
| `data/global/excel/skills.txt` | 0..64000, 64000..128000, 128000..173535 / 173535 | `465c28e83657540e0d7ae5b0de826d5657be7fba1aa539a85c3ca3aeea34b498`; complete |
| `data/global/excel/itemstatcost.txt` | 0..40109 / 40109 | `ee7fa82b48e0c4f27a566334b2da13cf1c83f5c4c0439552084a231cb5938375`; complete |
| `data/global/excel/missiles.txt` | 0..64000, 64000..128000, 128000..192000, 192000..192979 / 192979 | `89bee1fdeeb520d6eb3920437d237ba5a70bdf045770ed9808ca1a611fd932f0`; complete |
| `data/global/excel/sounds.txt` | 0..64000 and 192000..256000 / 505304 | complete header and cited named cells only; no full-file hash or absence claim |

Each returned range was checked for success, offset, returned size and total size, joined before parsing, and parsed by the full tab-separated header. IDs below use explicit `Id`/`ID`/`Index` cells rather than assuming a text line number. Separators do not establish identities. The following narrow cells are **verified in 1.13c data**; their code-defined meaning comes from the separate consumer/descriptor readings.

| File / named row | Named cell → raw value | Compiled correspondence separately measured |
|---|---|---|
| Skills / Jab, Id 10 | `cltstfunc` → `12`; `cltdofunc` → `16`; `dosound a` → `weapon_1ht_1`; `dosound b` → `weapon_2ht_1` | record id 10; +0xf2=12, +0xf4=16, +0x102=286, +0x104=292 |
| Skills / Cobra Strike, Id 265 | `progressive` → `1`; `cltprgfunc1`, `cltprgfunc2`, `cltprgfunc3` → `13`; `aurastate` → `progressive_steal`; `aurastat1` → `progressive_steal`; `prgsound` → `assassin_chargeup_cobra_1` | record id 265; progressive bit set; +0xf6/+0xf8/+0xfa=13; +0x80=123; +0x54=319; +0x10e=2282 |
| ItemStatCost / modifierlist_skill | `ID` → `350` | fixed 350 is read by client code; the source name is not a string stored in the live stat list |
| ItemStatCost / modifierlist_level | `ID` → `351` | fixed 351 is read by client code; the source name is not a string stored in the live stat list |
| Sounds / weapon_1ht_1 | `Index` → `286`; `FileName` → `combat\weapon\one hand thrust01.wav`; `Group Size` → `6` | Skills Jab +0x102=286; this does not identify which group variant was played |
| Sounds / weapon_2ht_1 | `Index` → `292`; `FileName` → `combat\weapon\two hand thrust01.wav`; `Group Size` → `6` | Skills Jab +0x104=292; no gameplay selection capture |
| Sounds / assassin_chargeup_cobra_1 | `Index` → `2282`; `FileName` → `skill\assassin\cobra1.wav`; `Tracking` → `1` | Skills Cobra +0x10e=2282; no voice or audible-lifetime capture |
| Sounds / sorceress_firebolt_1 | `Index` → `2410`; `Group Size` → `3`; `Loop` → `1`; `Stop Inst` → `0`; `Tracking` → `1`; `FileName` → `skill\sorceress\firebolt1.wav` | base cue and subsequent selected variant measured separately below; no complete Sounds record/compiler or device-output claim |
| Missiles / firebolt, Id 58 | `pCltDoFunc` → `1`; `pCltHitFunc` → blank; `TravelSound` → `sorceress_firebolt_1`; `HitSound` → `sorceress_firebolt_impact_1`; `ProgSound` → blank; `CollideKill` → `1` | record id 58; +0x08=1, +0x0a=0, +0x12=2410, +0x14=2413, +0x34=0, +0x186=1 |
| Missiles / moltenboulder, Id 452 | `pCltDoFunc` → `47`; `pCltHitFunc` → `52`; `TravelSound` → `druid_boulderloop`; `HitSound` → blank; `ProgSound` → `sorceress_fireball_1`; `CollideKill`, `AlwaysExplode` → `1` | record id 452; +0x08=47, +0x0a=52, +0x12=2331, +0x14=0, +0x34=2416, +0x186/+0x18c=1 |
| States / freeze, Id 1 | `onsound` → `impact_bluemotize_1`; `offsound` → blank | +0x26=365, +0x28=0 |
| States / poison, Id 2 | `onsound` → `impact_poison_1`; `offsound` → blank | +0x26=376, +0x28=0 |
| States / wolf, Id 139; bear, Id 140 | `onsound`, `offsound` → `druid_morph` | each record +0x26/+0x28=2346 |

The compiled States array contained 185 records of 0x3c bytes. All 11100 bytes were read in contiguous ranges of 4096, 4096 and 2908 bytes; a request exceeding the reader's per-read limit was not used as evidence. The Skills array contained 357 records of 0x23c bytes; the complete 572-byte Jab and Cobra records were read independently. The Missiles array contained 684 records of 0x1a4 bytes; complete 420-byte firebolt and moltenboulder records were read independently. Pointers/counts were obtained from the live DataTables fields, not addresses inferred from a preferred image base. Text cells and compiled values agree for the cells explicitly recorded above. This is a finite correspondence check, not proof of every compiled row or the complete link compiler. A blank sound cell must not be presumed to compile to -1: the measured state offsound and missile HitSound examples compile to zero, which their consumers forward.

### Disputed original preparation and corrected shared dispatch identity

The 2026-10-29 recipe assumed an active States row could select handler 16. Complete effective States coverage and the complete compiled array establish that no row selects 13 or 16 through `cltactivefunc`. The active, nonzero selectors in this profile are sanctuary 86, openwounds 85, shiverarmor 87, attached 88, hurricane 90 and vine_beast 89. This contradicts the proposed preparation, not SMS-R01/SMS-R02's supplied arithmetic or a measured gameplay result: no matching observation had occurred.

The original table at D2Client RVA 0xdea48 is shared. Its 130-entry bound is at 0xdec50; slot 13 contains 0x40a70 and slot 16 contains 0x14f40. D2Common 0x61ec0's final descriptor block maps `cltstfunc` (string RVA 0x90c0c) to Skills +0xf2, `cltdofunc` (0x90c00) to +0xf4, and `cltprgfunc1/2/3` (0x90bf4/0x90be8/0x90bdc) to +0xf6/+0xf8/+0xfa. D2Client 0xa2c90 consumes +0xf4 through the shared table. D2Client 0xa1b10 consumes +0xf6; 0xa2950 selects successive +0xf6/+0xf8/+0xfa callbacks. Thus Jab and Cobra Strike provide actual original-data routes to these handlers. Neither the state dispatcher alone nor a slot pointer proves that an active state selects them.

SMS-O01 now requires an Amazon legally able to use Jab. Record class, selected skill id 10, level, used-skill node, unit marker, ordinary callback selector 16 and compiled cue fields. Capture the root-to-callback execution and every resulting request/result. Match SMS-V03 only if its complete supplied inputs coincide; otherwise evaluate SMS-R01/SMS-R02 with the measured inputs, without relabeling a different cue as 17. A Cobra Strike fixture similarly needs the progressive event's actual charge index, associated state/list/stat and skill identity; three table cells equal to 13 do not prove callback count or cadence. Both preparations remain unperformed. The earlier hypothetical state recipe is not retained as an executable alternative.

### Full local producer findings and retained boundaries

Active processing 0xa08e0 tests the active mask, traverses the prepared state-index list, tests each membership and bounds-checks the signed `cltactivefunc` before calling the shared table. Defaults are skill id 0 and level 1; an associated list supplies layer-zero stats 350/351. The list/count are reloaded after each callback. D2Common 0x34000 tests the active-mask overlap; 0x33cd0/0x37db0 read membership from the extended list bitmap. Associated-list lookup 0x38520/0x37f80 traverses the two linked-list chains by state identity; 0x39140/0x38b00 reads the packed layer/stat key with the stat-definition handling in the existing RLI contracts. This resolves the local reads, not all state creation and callback scheduling.

For marker cadence, D2Common 0x32460 clears unit +0x4e before advancement. Ordinary animation scans crossed frame-data bytes in order and retains the last recognized event 1..4; other bytes do not overwrite that marker. Its wrap path scans to the animation end and resumes after the wrap adjustment. Sequence mode calls 0x2f100, which calls 0x2e6e0 with previous/current 8-bit-fraction frame positions. Equal positions use the current six-byte sequence record's event byte; crossed frames scan from previous integer frame plus one through current integer frame, retaining the last nonzero event, with zero initialized when none is found. The result is stored at +0x4e. These are different event scans. Complete animation setup/reset and client update cadence remain SMS-Q01; marker 3 is not assigned an invented action name.

The used-skill wrapper 0xa3000 obtains the unit's current skill id and level and calls ordinary dispatch 0xa2c90. Presentation callers include 0x197d0 and 0x4f2c0, with marker checks and a handled latch at unit +0xc4 bit 6; their special skill branches prevent treating every animation frame as one unconditional callback. The previously read player frame 0x1c150 and monster frame 0x4f4a0 are corresponding presentation roots, not a server tick contract. Per-unit dispatch 0x2b1e0's original jump table at 0x2b340 selects 0x1c150 for kind 0, 0x4f4a0 for kind 1 and 0xb1af0 for kind 3, after its preceding removal/flag and presentation-preparation handling. This locates the local unit-kind root, not the complete outer frame cadence. Received progressive dispatch 0xaefb0 selects 0xa2950 for progressive skills and 0xa1b10 otherwise. Complete packet admission remains distinct.

GDA-R11a supplies the flag names/bit positions below with its existing evidence date and limits. Start dispatch 0xa1cf0 uses the separate table at 0xde928 with count 70, selecting Skills +0xf2. Its callback result, `stsuccessonly` bit 12, `stsounddelay` bit 13, `weaponsnd` bit 14, item-controlled replacement sound, class sound and target overlay paths must all be preserved in full integration. Callback failure does not universally suppress start sounds: the success-only flag changes that path. Ordinary dispatch 0xa2c90 validates the initial +0xf4 selector, may replace it for item control, and admits generic positive `dosound`, target sound and target overlay only after its callback/missile success handling. Progressive 0xa1b10 and 0xa2950 have their own callback-success gates and target/path/seed effects. Charged progressive dispatch may execute earlier charge callbacks when `prgstack` is set. The fixed shared-table identity does not collapse these callers into the supplied arithmetic stage. Complete generic flag/activation, delay helper, overlay and frame ownership remains SMS-Q02.

State message 0xadd70 calls 0x73760 before applying the following transmitted stat payload. Its generic stat decoder and 0x72610 create/find the associated list and write the decoded stat/layer; fixed skill identities 350/351 need not have dedicated hardcoded writers here. The separate 0xadeb0 path also invokes state-on and then state presentation. State-off message 0xadd30 invokes 0x73630 after its preceding local cleanup. State-on's notondead exclusion sets membership and returns without the on cue; existing membership takes refresh/list handling. State-off detaches the matching associated list before its sound request when present, then clears membership and performs later presentation work. These surrounding mutations are real full-path obligations, still excluded from the supplied sound-only stage. Complete list, transformation, overlay and message-validation ownership remains SMS-Q03.

Full creation 0xb30f0 validates/prepares the missile and invokes its optional creation callback before the travel request. Only inside the nonzero TravelSound branch does it optionally search the owner for fixed cue 314 and request retirement through 0x88bf0. The per-missile dispatcher 0xb1af0 reads `pCltDoFunc`, requires it positive and below 91 and jumps through table 0xdd8a0. Slot 47 reaches 0xbb470; the effective moltenboulder row and compiled selector independently agree. The progression handler then continues through 0xb92b0 to the common tick/collision path 0xb17c0, with invalid/special removal paths reaching 0xb1720. Server missile movement contracts do not close these client paths.

Hit path 0xb1420 includes next-hit state 86 and last-collision suppression, target eligibility/friend policy, a returned hit bitmask and an optional hit callback that can return before HitSound. Its no-target/no-explicit-collision path can skip hit effects unless AlwaysExplode is set. A surviving target with CollideKill false can return before travel retirement. Where retirement is reached, nonnegative TravelSound is looked up through 0x225d0 and a found handle is passed to 0x88c20 with force false. Common removal 0xb1720 also requests associated travel-voice retirement. Descriptor readings bind CollideKill +0x186, CollideFriend +0x187, NextHit +0x188, NextDelay +0x189 and AlwaysExplode +0x18c. Complete target, path, hit callback and client-registry removal dependencies remain open.

Associated-voice lookup 0x225d0 normalizes the cue's sound group and searches the unit-associated handles for a selected cue in that group. Retirement request 0x88c20 removes owner/handle associations and conditionally invokes the retirement helper using remaining owners, force, StopInst and buffer/stream fields. It is not an unconditional immediate stop. Specialized progression reader 0xb8f00 uses a different end-counter comparison, force-true retirement request and positive ProgSound; 0xbb5b0 is another separate consumer. These are not collapsed into 0xbb470 or credited as a full voice lifecycle. The playback contract retains admission/device/retirement gates; even a matching wrapper result would not prove an audible outcome.

### Capture validity and remaining observation

The initial missile-creation attempt on this date watched D2Client 0xb30f0 during ordinary town walking. It expired by its time limit with zero records, zero evictions and no in-flight calls. No eligible missile creation was established during the window, so this is inconclusive and matches no vector. Table reads above remain data evidence only. A later normal menu transition temporarily ceased answering fresh state/frame requests; the control server subsequently recovered and a fresh frame showed the title menu. No skill input or gameplay result is inferred from the unresponsive interval.

After normal menu creation of a level-one Expansion Sorceress, the starter equipment supplied Fire Bolt. Measured base skill level was zero and effective dispatched level was one; those are distinct measurements. Normal walking reached the Blood Moor. The first two casts produced ordinary-dispatch, creation and expiry records, but the common-request records did not expose the cue register, so they alone were incomplete evidence of the requested identity. A high-frequency position-record window filled and was replaced before its records were retained; it is not used for a negative or complete-history assertion. A subsequent setup stopped on an already-armed target before any cast. Its incomplete window is also not credited. The coordinated windows below were collected only after all their recording targets successfully armed.

**SMS-O02 — partial Fire Bolt producer and retirement observation, 2026-10-30.** Original Expansion 1.13c, single player, the effective data profile above. Reproduction: create a normal Sorceress, retain the starter staff's Fire Bolt, leave town through normal walking, cast once into open ground, and observe the ordinary skill callback, client missile creation, common request, complete associated voice and later client hit/removal path. Preserve record identities across reused pointers; a pointer alone is insufficient. Do not infer a hit or audible stop from an expiry request.

| Observed boundary | Actual measured assertions | Limits |
|---|---|---|
| ordinary skill dispatch, D2Client 0xa2c90 from return location 0xa3035 | one call, skill 36, level 1, item control 0, emitter the player; result 1; stack cleanup 16 bytes | full earlier input/activation and animation cadence were not recorded |
| creation, 0xb30f0 from 0xa2792 | one complete 92-byte parameter snapshot, missile id 58, owner the same player; result missile M | complete local allocation/path outputs are not a supplied-stage vector match |
| travel request, 0x88a70 from 0xb36e4 | emitter M, three auxiliary values zero, returned logical handle 716; complete emitter snapshot identifies missile 58 with unit id 475 | the cue register was not directly exposed; the immediately initialized voice below corroborates the requested identity |
| voice initialization, 0x88350 from 0x88b3a | all 89 voice bytes captured; +0x04 cue 2410, +0x08 handle 716 | this is before later group selection; it does not prove device output |
| subsequent voice updates, 0x88350 from 0x89071 | same voice and handle 716; cue became 2412, in the separately read 2410 group of size 3 | no RNG input/draw captured; this does not independently confirm the group-selection formula |
| hit/removal entry, 0xb1420 from 0xb1a9f | one call for missile 58/unit id 475; supplied collision flag zero; result zero; snapshot changes unit kind 3 to 6, id to -1, clears its associated voice-list pointer and other removed resources | entry target register was not exposed; caller reading identifies the expiry path; a zero result is not proof no side effects occurred |
| association retirement, 0x88c20 from 0xb16bf | emitter M with matching missile id/unit id, force false; one call | logical handle input register not directly exposed; complete retirement-voice snapshot below links the corresponding handle |
| retirement helper, 0x88550 from 0x88ccf | complete 89-byte voice, cue 2412, handle 716, empty owner-list pointer; retirement field +0x3d changes 0 to 1 | actual device stop, resource release and eventual slot retirement remain STP-R04b |

This eight-second action window retained one ordinary-dispatch, one creation, one hit/removal, one association-retirement and one retirement-helper record, seven common-request records and 497 complete position-operation records. All reported zero dropped/evicted records; all relevant dereference windows were complete. Targets were stopped after collection. The separate associated-voice lookup window stopped at its 256-record cap before the later event and supports no absence or complete-lifetime assertion. An earlier two-second coordinated cast retained one ordinary dispatch, one creation, four common requests and 185 complete position records, including base travel cue 2410 with handle 687. It ended before expiry and cannot assert that expiry or retirement was absent.

The complete voice identities, original caller locations, request emitter, parameter snapshot and creation/removal unit id link the eight-second case across its local producer/request/retirement boundaries. Independent function record sequences and durations do not provide a common timestamp, so no measured wall-clock ordering or exact cadence is claimed. Caller evidence supplies the local nesting/order. The observed retirement flag is consistent with the statically read scheduler/retirement boundary; it does not close admission or audible lifetime. SMS-V14's synthetic negative-cue/failure inputs and SMS-V16's admitted-hit inputs were not exercised, so their `runtime_observed` flags remain no. No new smaller rule is promoted from this partial full-scope observation.

A later attempt to aim Fire Bolt at a nearby monster retained zero skill-dispatch/creation/hit records. The character had died during intervening record analysis, independently confirmed by mode 17, zero life and the fresh death screen. That is invalid preparation for a cast, not evidence that the handler or hit path failed. The normal death-screen action returned the character to town. It does not invalidate the preceding live cast, but leaves actual target-hit admission and HitSound unobserved. Jab, Cobra Strike, moltenboulder progression, state on/off, complete scheduling and audible output also remain unobserved. Full original-game captures of those legal fixtures and the remaining full-path assertions remain SMS-Q01–SMS-Q05.

## Supplied-stage inventory and event boundaries — 2026-10-30

Section 3 explicitly names the supplied on/off state-transition selector for SMS-R04a and current path-context presence for SMS-R04c. Section 5 names lookup occurrence for SMS-R01 and layer-zero stat-read occurrence for SMS-R03a. The rules and vectors already require these decisions; the inventory correction adds no default, width, formula, request result or gate.

SMS-R04c and every SMS-V15 case expose ordered request/counter-write events. Final counter plus a sound-only log cannot distinguish writing before the request. The base case requests then writes 6; first-seen and equality cases write 6 and 5 respectively without a request, including a write that leaves the value unchanged; the nonpositive-cue guard performs no write; absent context requests then writes 0. The request result remains ignored. This observation boundary changes no callback or arithmetic and does not claim that the partial gameplay capture observed progression.

Additional unexercised partitions include invalid/absent SMS-R03a skill records; SMS-R04a monster mode 12 and positive/negative offsound; SMS-R04b positive TravelSound; and SMS-R04c negative ProgSound. All five broad questions and draft status remain. The partial Fire Bolt observation confirms neither these synthetic cases nor audible lifetime.
