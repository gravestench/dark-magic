# Ambient and weather sound evidence

Spec: [Ambient and weather sound](../../specs/sound-ambient-and-weather.md)

## Identity and reading boundary

2026-10-29: original Windows x86 Expansion 1.13c. [Source manifest](../source-manifest.md) provides `LOD113C-D2CLIENT` and `LOD113C-D2COMMON`. D2Client SHA-256 was checked during this audio investigation and matches `dd8bc6025de921216a97c17f97cd1a50fbb85926e838ec60e13451448836d906`. No earlier companion or ambient contract was present. General audio research contains design recommendations and non-1.13c observations; these were used only as leads.

The initial controller had no matched reconstruction entry. Its full disassembly was read; original-image instruction listing corroborates the client audio update, selection, timer and client random helper. Available byte-matched reconstruction was read before assembly for callees listed below. Public descriptions are analyst labels; misleading original analysis names are not evidence of semantics.

## Function map

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2CLIENT | `0x5f020` | none established | Enabled audio update | SAW-R01, SAW-R02 | Original-image listing plus disassembly: ambient, music, other audio updates, counter increment |
| LOD113C-D2CLIENT | `0x4f860` | none established | Environment audio controller | SAW-R01, SAW-R02, SAW-R04b, SAW-R05 | Full disassembly; eligible supplied-operation caller verified, nested request integrations gated |
| LOD113C-D2CLIENT | `0xcacc4` | none established | Finite float-to-integer conversion | SAW-R05 | Full disassembly and independent original-image listing; finite product in signed 32-bit range |
| LOD113C-D2CLIENT | `0x89640` | none established | Local player seed modulo draw | SAW-R02 | Full disassembly, byte runs corroborated by original-image listing |
| LOD113C-D2CLIENT | `0x5dfd0` | none established | Sound/environment loader | SAW-R01 | Descriptor positions corroborated; generic parser integration not verified |
| LOD113C-D2CLIENT | `0x80450` | none established | Environment eligibility predicate | SAW-R04a | Disassembly and original-image corroboration; Rain predicate identified below |
| LOD113C-D2CLIENT | `0x4f770` | none established | Environment intensity getter | SAW-R04a | Matched reconstruction and original-image corroboration; selected intensity identified below |
| LOD113C-D2CLIENT | `0x87ed0`, `0x87f00` | none established | Write/read source amplitude field | SAW-R04b | Matched reconstruction read; scans source list by handle |
| LOD113C-D2CLIENT | `0x88230` | none established | Source transition scheduling | SAW-R04b | Disassembly read; complete playback integration gated |
| LOD113C-D2CLIENT | `0x88810`, `0x88770` | none established | Exclude source groups during transition | SAW-R04b | Disassembly read; downstream stop helper not fully resolved |
| LOD113C-D2CLIENT | `0x88ba0`, `0x88bf0` | none established | Update source position / stop by handle | SAW-R04b | Matched reconstruction read; downstream source operations unresolved |
| LOD113C-D2CLIENT | `0x4f720`, `0x4f750`, `0x5de10`, `0x5de90` | none established | Environment material sound eligibility bookkeeping | SAW-R04b | Matched reconstruction for `0x4f750`, disassembly for others; material binding not established |
| LOD113C-D2CLIENT | `0x5cf90` | none established | Environment acoustic parameter update | SAW-R04b | Disassembly read; backend integration unresolved |
| LOD113C-D2CLIENT | `0x8ad30`, `0x8afd0` | none established | Screen-shake-associated source intensity | SAW-R04b | Disassembly read; candidate label does not establish weather |
| LOD113C-D2CLIENT | `0x88d90` | none established | Developer/random sound lead | SAW-R04b | Disassembly read; broad random sound range, no environment selection |
| LOD113C-D2COMMON | `0x37400` | 10396 | Environment phase accessor | SAW-Q01 | Matched reconstruction: indirect reads, clock semantics not established |
| LOD113C-D2COMMON | `0x6ccc0` | 10014 | Level table record accessor | SAW-Q01 | Matched reconstruction: stride `0x220`; table identity requires loader binding |
| LOD113C-D2COMMON | `0x6cc90`, `0x6cc30` | 10215, 10499 | Two environment eligibility bytes | SAW-Q02 | Matched reconstruction; reads bytes 5 and 6 of same record family |
| LOD113C-D2COMMON | `0x2fe10` | 10331 | Unit room accessor | SAW-Q01 | Disassembly read; unit-type-sensitive dynamic/static path |
| LOD113C-D2COMMON | `0x3c000` | 10826 | Room level identity accessor | SAW-Q01 | Matched reconstruction read |

## Dispatch, selection and writes

At D2Client `0x5f020`, the enabled-audio test gates a call to `0x4f860`, passing audio counter `0x11c268`. The following call is music coordinator `0x45660`; further calls update other audio state before incrementing the counter. These ordering claims are from the original instruction listing, including rel32 targets, not labels. This proves placement within the audio update, not its relationship to every server tick.

The ambient controller tests audio-enabled state and local player existence, obtains the room through import thunk `0xc1f0`/D2Common ordinal 10331 (`0x2fe10`), then level identity through thunk `0xc1de`/ordinal 10826 (`0x3c000`). Its phase access uses thunk `0xc7e4`/ordinal 10396 (`0x37400`), selecting the first pair for signed phase 1..3 and the second otherwise. The supplied boundary starts after those external accesses. A level table lookup via thunk `0xc214`/ordinal 10014 (`0x6ccc0`) supplies byte `0x21c`; it selects a client environment record of stride `0x58`. Missing-record handling must be completed before accepting full integration.

SAW-R01 reads environment offsets `4` and `0xc` on the first branch, `8` and `0x10` on the second. Client loader descriptor locations corroborate authored names: Day Ambience at descriptor destination 4 (`0x5e3cd`), Night Ambience at 8 (`0x5e3f4`), Day Event at `0xc` (`0x5e41f`), Night Event at `0x10` (`0x5e44a`), and Event Delay at `0x14` (`0x5e475`). The descriptor name references point to the exact named strings. Generic parse and text/compiled correspondence remain a producer integration gap. SAW-R01 uses supplied logical fields and does not require that gap to be closed.

The direct timer stage begins at `0x4fab4` after preceding transition/weather calls. Selected event zero branches past timing/playback; unchanged identity loads stored interval/last tick. Changed nonzero identity computes unsigned divide by three using the high product of `0xaaaaaaab` followed by a right shift, makes two direct calls to `0x89640`, and writes interval `0x11c328`, initial last tick `0x11c32c`, and event identity `0x11c324`. No playback call occurs between these two direct draws. Later comparison is unsigned wrapped tick difference against interval; subsequent successful playback and direct placement draws remain SAW-Q03. An initialization-only vector does not claim the entire update has no further draws.

The client helper `0x89640` tests signed positive modulus, reads seed words from local player offsets `0x20/0x24`, multiplies low by `0x6ac690c5`, adds high with carry, stores both words and uses remainder, with a power-of-two mask optimization. The original-image listing confirms the initial `db`-encoded test/load/branch and the later multiply/carry/store/remainder sequence. This consumes NRF-R01/R02's numerical operation, without importing their server seed-ownership claims. No call lies within the helper's verified path.

## T01 named data cells

Reuse date and collection date: 2026-10-29, owned vanilla LoD 1.13c launch in the coordinating audio investigation. Effective `data/global/excel/soundenviron.txt` was read through the game's Storm loader, offset 0 requested length 60000, returned 5741/5741 bytes. The complete encoded body was decoded, its tab-separated header parsed by name, and 50 complete rows identified. The cited row is selected by `Handle = ESOUNDENVIRON_WILDERNESS`, explicit `Index = 2`; no positional indexing assumption is needed.

| Named column | Raw cell | Interpretation / support |
|---|---|---|
| Day Ambience | `70` | Decimal 70, SAW-R03 |
| Night Ambience | `71` | Decimal 71, SAW-R03 |
| Day Event | `192` | Decimal 192, SAW-R03 |
| Night Event | `197` | Decimal 197, SAW-R03 |
| Event Delay | `250` | Decimal 250, SAW-R03 |

The record establishes the effective text cells, not their compiled runtime values or audible execution. The compatible same-collection record in the [music companion](sound-music-selection-and-transitions.md) preserves the full 5741-byte effective-file SHA-256 `d69131977215f8d3ac91326ffa708fad24eeddb61efae3965cf4b4d9d41de90c`; this fingerprint is reused with the original 2026-10-29 collection date, rather than claimed as a new read. Physical supplying archive and complete override inventory remain uncaptured. The original collection's other limits are retained. No gameplay vector is promoted by reuse of this table read.

Same-collection reconciliation corrected the requested length from the
returned size 5741 to the actual request limit 60000. Returned size,
completeness, named cells, date and profile are unchanged; this correction
adds no behavioral verification.

## Unresolved weather and playback boundary

The controller obtains a predicate through `0x80450`, choosing sound identity 64 when it returns true, and derives a target amplitude from `0x4f770` multiplied by an original-image constant. It uses six-unit per-update convergence toward that target, starting or stopping a retained source. The names of the imported record-byte predicates do not establish rain/indoors: existing labels refer to unrelated resistance bonuses, contradicted by their actual record reads. Complete level-field/producer binding is required.

On ambience change, there is a special case when switching between the current environment's two ambience identities; it schedules a 250-unit transition. Other changes issue group-exclusion calls. Event success selects randomized side/distance placement, updates last tick and samples a new delay. Those callers and callees have not all been resolved as complete behavioral contracts, so these remain research observations rather than accepted rules.

Candidate `0x8ad30` is called by `0x8afd0`, whose inputs derive from a three-part screen-shake envelope. It scales a value into 0..255, uses sound 452 and smooths the source field by six. This is not evidence of general rain audio. Candidate `0x88d90` tests counter modulo three and selects a broad random sound-id range; its occurrence within active-source coordination is not proof of environmental sound selection. The actual sound environment controller is `0x4f860`.

## Observation procedure

SAW-O01: use matching original LoD 1.13c, an expansion character in Normal and an area with differing day/night environment entries. Capture local player seed, phase, selected environment fields, audio counter and timer state immediately before the selection/timer stages and after their writes. Compare synthetic supplied cases only when all inputs match; otherwise construct a fully specified vector from established rules before promotion. For timer initialization, account for preceding playback's seed effects by sampling at the timer-stage entry. Record subsequent event playback separately. Phase is now bound to the world-clock cycle, but no phase-transition observation is claimed. No runtime trace has yet been armed for this candidate; authorization for tracing an owned responsive member is recorded. Runtime confirmation remains possible and uncompleted.

## Search and self-review

The investigation followed candidate functions, located audio-update root and scalar output writes, resolved six common-module import targets, read available matched reconstructions first, corroborated timer/random decoding against original-image listing, and reused complete effective table evidence under its existing limitations. Interactive analysis had no running shared instance; read-only original instruction listings supplied a fallback. No shared target was switched. Complete weather and playback callee closure and gameplay observation remain open; the full search ladder is not claimed exhausted.

Self-review preserves the distinction between supplied verified stages and gated actual playback. No column name is used to derive weather semantics. Initial numeric vector calculations were independently recomputed with integer arithmetic. Fresh blind-reader review is pending the coordinator; no approval is inferred.

Mechanical lint: initial run 0 errors / 18 warnings. The glued predicate token was corrected. Remaining history warnings name the scheduler's retained previous-event state, which is current behavior; matched-reconstruction warnings describe verification methods with original public addresses and no private source paths. These warnings were reviewed and retained.

## Blind reader pass 1

2026-10-29, fresh reader: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1, GATED 5, COVERAGE 5, HARDEST 3. No approval is inferred. The following triage changes the vector set and requires a fresh full review.

| Finding | Outcome |
|---|---|
| DISCRIM: phase 2 absent, permitting first-pair selection only for phases 1 and 3 | Added SAW-V01(d), explicitly supplying phase 2 and all four environment identities, asserting 11/33 and rejecting 22/44 |
| COVERAGE: phase 2 | Addressed by the same SAW-V01(d) case |
| COVERAGE: selected event zero | No change: advisory branch coverage; SAW-R02 explicitly retains scheduler and seed unchanged, while vectors target changed and unchanged nonzero selections |
| COVERAGE: zero delay | No change: advisory branch coverage; SAW-R02 defines both draw guards, including one positive-modulus-one step followed by no step for modulus zero |
| COVERAGE: nonpositive signed modulus | No change: advisory branch coverage; SAW-R02 explicitly returns zero without advancing seed and cites the owning numerical primitive |
| COVERAGE: wrapped arithmetic | No change: advisory branch coverage; wrapping and signed reinterpretation are explicit in SAW-R02, with no full-update edge-case acceptance claimed |
| HARDEST: producer integration | No change: SAW-Q01 already gates phase/environment producers; supplied inputs define the available local boundary |
| HARDEST: weather meaning | No change: SAW-Q02 already gates weather predicates and intensity producers; no guessed meaning is added |
| HARDEST: playback/event integration | No change: SAW-Q03 already gates playback success, additional draws and complete event execution |

The five GATED entries describe documented limitations and require no behavioral default. Review of the changed vector is pending a fresh reader.


## Fresh blind-implementer review: full pass 2

2026-10-29: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 5, COVERAGE 4, HARDEST 3. The independent reader recomputed SAW-V01 and every case, both exact random draws and seed states of SAW-V02, and the unchanged scheduler/seed in SAW-V03. All matched. It independently checked that modulus 250 on the second low word would give 59 rather than the required 3 modulo 274, and that omitted backdating would give lastTick 1000 rather than 997.

COVERAGE outcomes: no change for selected event zero, zero delay, nonpositive signed moduli and wrapped arithmetic. Their operations and no-step versus step behavior are explicit; these remain advisory unexercised partitions rather than accepted observations. HARDEST outcomes: no change for producer/correspondence Q01, weather Q02 and full playback/event sequence Q03, all already explicitly gated. The five GATED entries retain Q01 through Q04 and R04. The current supplied-input contract meets the reader stopping rule; weather/audible parity and runtime confirmation remain open.

Final mechanical validation on 2026-10-29: 0 error(s), 17 warning(s). Warnings were reviewed: temporal field names describe current retained state, and source-reading method names describe public verification methods rather than private reconstruction paths. No implementation or user approval is asserted.


## Static phase and weather producer recheck — 2026-10-29

The prior semantic gaps were rechecked against the original artifact and compatible foundation evidence. D2Common ordinal 10396/RVA 0x37400 returns the first dword of the current act environment (act field +4). The world-clock companion identifies this as cycle, with period in the following dword. WCK-R01 through WCK-R03 own its values and progression. Client current act RVA 0x11c3b8 supplies this accessor. SAW-R01 consumes cycle; it does not map the authored Day/Night field names to clock hours.

D2Common loader 0x6e130 descriptors bind Rain to byte 5, Mud to byte 6 and SoundEnv to byte 0x21c of the prepared level record. Rain and Mud names point to original strings at RVAs 0x8ff28/0x8ff24; SoundEnv at 0x8fac8. The SoundEnv descriptor is at 0x6f81c..0x6f839, and loader call 0x6f8e3 publishes the table pointer/count used by getters 0x6ccc0, 0x6cc90 and 0x6cc30. Original initialized-pointer/global linkage corroborates the same table. Descriptor identity/direct conversion reuse GDA-R03a/GDA-R03b; full compilation remains GDA-R02/GDA-R03d. This closes field identity without asserting complete effective text/runtime correspondence.

SAW-R04a: D2Client predicate 0x80450 returns zero when suppression global RVA 0x11bf78 is nonzero or local player RVA 0x11bbfc is absent. Otherwise it obtains unit room, level identity, then Rain via client thunk 0xcab4 → D2Common #10215/RVA 0x6cc90, converts the returned byte to Boolean, calls Mud through 0xcaae → #10499/RVA 0x6cc30 and ignores that return. Matched getters are leaves. The separate matched getter 0x4f770 follows the same suppression/player/Rain condition, also calls and ignores Mud, and returns stored float RVA 0x1138f4 only when eligible. Its fallback constant at RVA 0xd7578 is positive zero in the original image; the reconstruction label suggesting one was not used as evidence. Complete applicable reads and original-image instruction listings support these output paths. The supplied-input boundary does not execute the external room/level producer. No accessor writes state or draws RNG.

The weather-state producer at 0x811f0, updater 0x81340 and initializer 0x81750 were additionally read as unmatched instructions. The updater separately tests Rain and Mud and can clear stored intensity; the state updater ramps a count and computes intensity. Complete timing/RNG/particle callee and conversion closure remains SAW-Q02. SAW-R04a deliberately ends before target-amplitude conversion and playback; no generated intensity range or outdoor/indoor rule is inferred.

New SAW-V04 covers non-one Rain, Mud-only rejection, negative suppression, absent player and unchanged negative supplied intensity. Lint/self-review are renewed and fresh blind review is required for this changed contract. Earlier reader passes retain their original three-vector scope.


## Fresh blind-implementer full pass 3 — 2026-10-29

MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 5, COVERAGE 4, HARDEST 3. The reader recomputed every current case, including SAW-V04 variants, and confirmed the current-cycle dependency binding rather than period.

COVERAGE outcomes: no change for zero selected event, zero delay, nonpositive signed draw moduli and wrapped scheduler arithmetic. Their outcomes remain explicit and the existing vectors distinguish changed versus unchanged timer state and the two direct draws; these are advisory unexercised partitions. HARDEST outcomes: no change for the two-draw timer (107 then 3, interval 274, last tick 997), suppression/player/Rain selection with unchanged stored intensity, and cycle-versus-period binding. Each was recomputed consistently. GATED outcomes: retain SAW-Q01 through SAW-Q04 and SAW-R04b. The current supplied-stage reader stopping rule is met; full weather/audible execution remains open.


## Additional intensity-state reading — 2026-10-29

D2Client 0x811f0 and helpers 0x018d0, 0x81120 and 0x80ca0 were read with reconstruction availability checked first. The interval helper performs the foundation two-word RNG step and adds a supplied offset, returning the offset without a draw for a nonpositive signed range. Transition states 1 and 3 update a retained count with unsigned integer division; the common intensity write converts that signed count and multiplies by the original `binary32` constant 1/256 at RVA 0xd75b0. State changes can consume additional draws and call a separate particle/setup helper; level-special cases also change parameters and RNG. Those producer inputs and all nested state-change effects have not been closed, so these readings narrow SAW-Q02 but do not add a verified intensity-generation rule or a range assumption to SAW-R04a.

## Full eligible caller closure — 2026-10-29

SAW-R05 reads the complete controller from original-image disassembly at
D2Client 0x4f860 through its return, reusing the same artifact identity and
the prior full native listing. The boundary supplies eligibility, valid
prepared environment, accessor results and normally returning audio
operations. It verifies every outgoing request's arguments/order and the
caller-owned writes/draws; it does not relabel unclosed voice/backend
execution as verified. Earlier reader passes retain their four-vector scope.

### Inputs and retained state mapping

The frame root 0x5f020 calls this controller before music coordination.
Controller argument is the current audio counter. Retained state maps to
D2Client RVAs 0x11c31c (ambience identity), 0x11c314 (ambience handle),
0x11c320 (weather identity), 0x11c318 (weather handle), 0x11c324 (previous
event), 0x11c328 (interval), 0x11c32c (last event tick), and 0x11c310
(last level identity). Initial level and re-read environment index occupy
different temporary values; final auxiliary eligibility is keyed by initial
level, not the environment index or due-event handle.

The context/room/level getters and prepared environment retain SAW-Q01
integration gates. Named environment +0x04/+0x08 ambience, +0x0c/+0x10
events, and +0x14 delay retain their prior descriptor evidence. Two final
environment dwords +0x1c/+0x20 are called auxiliary A/B here; their full
authored binding and eligibility effects remain SAW-Q03, without guessing
meaning from names.

### Weather amplitude conversion

After eligibility selects current weather identity 64 or zero, the first
0x4f770 result is compared with original positive zero at RVA 0xd7578.
An equal value, including negative zero, leaves target amplitude zero.
Otherwise a second accessor result is multiplied by the original finite
float 255 at RVA 0xd759c and passed to helper 0xcacc4. That helper had no
matched reconstruction available; its complete disassembly was read and
the original encoded subtraction instruction corroborated in an independent
native instruction listing. It converts via signed 64-bit integer storage,
subtracts the rounded integer from the original value, and corrects a
round-away-from-zero result by one using the original sign and residual.
The caller retains the low signed 32-bit result. For finite product within
signed 32-bit range this establishes truncation toward zero; no claim is
made about nonfinite, exceptional or out-of-range conversion. Multiplying
a supplied `binary32` by exact 255 is representable in the extended arithmetic
used here before conversion. No intensity clamp is present in this caller.

### Ambience and weather request branches

The changed-ambience branch compares new/old identities against both
environment ambience fields to detect a swap. A swap requests scheduler
0x88230 with old handle, target 0, duration 250 and delay 0. Ordinary changes
request 0x88810(new ambience,current weather identity), then
0x88770(selected event,0). The nonzero new ambience request at 0x4f9e8
passes four zero stack values to common request 0x88a70. The returned handle
is retained without a success check. Swap requests then call 0x87ed0 with
that handle/value 0 and 0x88230 with handle/255/250/0, even for handle zero.
Only nonzero new ambience generates a sound request; a zero selection
clears retained ambience/handle.

Weather request at 0x4fa44 loads its cue from **entering retained weather
identity** 0x11c320, not the current selected identity in the controller.
The current identity is stored later. Thus an initially zero retained cue
can cause a zero-cue request, which fails under a supplied zero result but
still records 64 for the next update. A successful new handle receives
target amplitude via 0x87ed0 and then enters the same amplitude-read/smoothing
path as an existing handle: two equal volume writes can occur. Matched
0x87f00 and 0x87ed0 were read; they scan the same ordered handle list and
read/write current amplitude, with zero/missing handles returning zero or
making no write. Signed conditional comparisons around wrapped +6/-6
implement the stated caller step. A zero computed candidate calls stop
0x88bf0 and clears the handle. Ineligible weather calls that stop helper
even when entering handle is zero and clears both retained fields.

### Event and final operation order

At 0x4fab8 the controller requests acoustic update 0x5cf90 with environment
index before any event-timer draws. The prior SAW-R02 stage is reused for
changed/unchanged selection. At 0x4fb20 wrapped elapsed is compared unsigned
with interval; `jb` skips only when below, so equality is due. The event
request at 0x4fb36 uses selected event and four zeros. Zero result branches
past all position/timer draws and writes.

On a nonzero handle, calls to the fully read seed helper 0x89640 occur in
the exact modulus order 301,2,201. Integer arithmetic produces distance
450+first result, sign -1 for second result zero/+1 otherwise, and vertical
third result-100. Position call 0x88ba0 receives handle,x,y,0. Its matched
reconstruction was read: it finds the matching handle, converts x/y/z and
calls the spatial cache helper after adding original 640 to z. SAW-R05
claims only the caller's integer arguments; nested positional/cache effects
consume S3D-R04 or remain gated. After position returns the caller writes
last tick, then draws the new interval with modulus 2*floor(D/3)+1. There
is no new backdating draw. The direct seed states and moduli for SAW-V08
were independently recomputed from the two-word arithmetic.

Finally, if initial level differs from retained level, the caller requests
0x4f750 clear, then 0x4f720 for each nonzero auxiliary identity in A/B order,
and stores initial level. The matched clear helper was previously read;
their full eligibility interpretation remains gated. 0x88bf0's matched
wrapper was also read; it finds a handle and invokes a downstream stop
operation. The wrapper's existence does not close actual retirement.

### Full-caller observation requirements

SAW-O02 requires the full eligible update: initial and retained level,
prepared environment, phase, both intensity readings, controller retained
state, every audio request result, volume reads/writes and all direct plus
dependency-owned seed steps. Compare the request log before attributing
audible effects. Use a level change for ordinary replacement, a phase switch
within one prepared Day/Night pair for swap requests, and naturally due
events for success/failure timing; record actual complete inputs rather
than assuming a requested transition succeeded. Weather preparation and
intensity scheduling still require SAW-Q02, so this is a measurement
requirement, not a claim that an exact weather case was produced.

No new runtime campaign was performed for this closure. All new executable
vectors remain supplied-operation static cases. Self-review explicitly
checks normal-return dependency gates, all callback argument order, the
retained-weather-cue quirk, post-start double write, equality/wrap due test,
four direct success draws and final level bookkeeping. A fresh full reader
pass is required after SAW-R05 and SAW-V05–SAW-V09 were added.

Mechanical validation of this continuation: zero errors, 32 warnings.
All warnings were reviewed and retained: previous/previousEvent identifies
current retained scheduler state, while reconstruction phrases identify
public verification methods. They do not encode private source paths or
claim that prior reader passes cover the new caller contract.


## Fresh blind-implementer full pass 4 — 2026-10-29

MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 5, COVERAGE 9 advisory items, HARDEST 3. The fresh reader independently recomputed all nine vectors and their subcases, including all request logs, volume effects and direct seed states in the expanded full eligible caller. Every named discriminator rejected its stated alternative.

COVERAGE outcomes: no change for isolated zero-event, zero-delay, nonpositive-modulus or other wrapped initialization partitions; R02 explicitly supplies their guard/wrapping rules and existing initialization/retention vectors distinguish the timer operations. No change for intervening callback seed mutation or changed initialization combined with callback effects: R05 explicitly consumes the seed after each supplied operation, while these fixtures intentionally supply no callback seed mutation. No change for unchanged/zero ambience, active-weather zero target or six-unit convergence/candidate-zero stop, or B-only/both/equal auxiliary identities: the ordered caller branches are explicit; current vectors target ordinary replacement, failed swap, weather start and success/failure event execution. Those remaining partitions are advisory coverage limits, not observations or proof of exhaustive behavior.

HARDEST outcomes: no change for the two timer draws/modulus sequence, retained weather identity and post-start double write, or wrapped event due/success/failure split. Each is explicit and recomputed successfully, with four ordered success draws versus zero failure draws. GATED outcomes: retain Q01 through Q04 and R04b. The current nine-vector supplied-stage contract meets the reader stopping rule. Complete audible integration and gameplay observation remain open; no approval is inferred.
