# Sound table and playback rules

**Description**

Sound playback connects a requested game sound to an audio voice and updates that voice over time. A sound group can exclude recent choices by drawing again. Requests share a limited set of playback slots, while loading, fading and completion determine when those slots can be reused. Spatial sound updates and external device behavior retain explicit evidence limits.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md)
Ids: STP-R, STP-Q, STP-V, STP-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/sound-table-and-playback-rules.md)
Open evidence gaps: [gap file](../plans/gaps/sound-table-and-playback-rules.md)

## 1. Scope

Sound-record interpretation and the client playback lifecycle: group choice, volume, looping, fading, duplicate deferral, stopping and voice priority. Music selection, world emitter selection and positional attenuation belong to their separate sound contracts. The supplied-input request/admission, duplicate/deferral, asset, buffer, stream and retirement lifecycle is defined here. Complete spatial integration, external source/device behavior and matching gameplay observations remain gated.

## 2. Trigger and ownership

D2Client owns sound request coordination and per-update voice state. Its sound update dispatch reaches the volume-transition stage before voice eligibility and playback/update operations. D2Sound owns the eventual device operations. The dispatcher, logical coordinator, asset processor, backend worker and stream I/O worker have separate entry boundaries and explicitly controlled interleavings. Wider player-context lifetime, spatial integrations and external source/device behavior remain STP-Q01/STP-Q02. The supplied-input stages make no claim about server frame cadence.

## 3. Inputs and data

| Input | Semantic source | Width and format | Meaning | Origin | Binding |
|---|---|---|---|---|---|
| transition active | voice update state | Boolean | whether to evaluate a pending transition | runtime input | supplied |
| current tick, start tick, end tick | client audio update counter and voice state | unsigned 32-bit tick values | unsigned comparisons select the transition branch | runtime input | supplied arithmetic stage; enabled-update units bound by STP-R06 |
| start volume, end volume, current volume | voice transition state | signed 32-bit integers | interpolation endpoints and stored result | runtime input | supplied |
| base cue, sound table present | group selection request | signed 32-bit cue; Boolean | base is a valid record index when the table is present | runtime input | supplied |
| Group Size | supplied base Sounds record | unsigned 8-bit | group member count; candidates are contiguous from base cue | table data | descriptor-associated field; supplied decoded value |
| most recent choice, older choice | base record's retained history | two 32-bit cue ids | ordered repeat-exclusion entries | engine state | supplied |
| sound-selection seed | current selected client unit's seed | two unsigned 32-bit words | advanced by every group-selection draw; see NRF-R01/NRF-R02 | runtime state | supplied seed; installation and wider seed lifecycle remain distinct |
| selected cue | supplied history-commit input | 32-bit cue id | chosen cue written after early suppression passes | runtime input | supplied |
| ordered voices and requested handle | scheduling request and current voice state | ordered logical records; unsigned 32-bit handle | first matching record is selected; zero handle skips | runtime input | supplied |
| target volume, requested duration, delay | scheduling request | signed 32-bit target; unsigned 32-bit duration/delay | requested endpoint, interval length and start delay | runtime input | supplied |
| emitter association list | selected voice | ordered associations | a second entry suppresses scheduling | runtime state | supplied |
| retirement-request condition | selected voice | Boolean | target zero sets it; STP-R03b consumes it | runtime state | supplied |
| Fade In, Fade Out | supplied selected Sounds record | unsigned 8-bit values | minimum schedule duration for target 255 or 0 | table data | descriptor-associated fields; supplied decoded values |
| valid buffer context and buffer-start marker | supplied device wrapper state | context reference and signed 32-bit marker | marker is written on successful buffered start | runtime input | supplied |
| device critical-section operations and repeating-play operation | supplied external device boundary | ordered operations; signed 32-bit device result | local wrapper logs calls and branches on result sign | external operation | supplied results; real device execution gated |
| ordered sixteen-slot pool and incoming voice | client admission state | slot class, associated voice and start marker; unsigned-byte priority, finite `binary32` key, unsigned 32-bit start and fixed pool index | free selection and displacement order | runtime state | STP-R03d; positional key production remains with the positional contract |
| loaded asset bytes, stored payload/length, Stereo and failure flag | loaded sound record | byte buffer and decoded WAV fields; Boolean flags | validation writes and preserved state | runtime state | STP-R04c; I/O is gated separately |
| audio updating enabled and audio counter | client audio dispatcher | Boolean and unsigned 32-bit integer | whether an invocation requests updates and advances the counter | runtime state | STP-R06; wall-clock cadence is not supplied by this field |
| cursor_pass Volume | effective data/global/excel/sounds.txt, Sound=cursor_pass, Volume | decimal text, raw 255 | named archive cell only; consumer scale is gated | table data | unresolved consumer |
| cursor_pass Group Size | same row, Group Size | decimal text, raw 0 | named archive cell only | table data | unresolved consumer |
| cursor_pass Stop Inst | same row, Stop Inst | decimal text, raw 1 | named archive cell only | table data | unresolved consumer |

Archive cells are supported by [T01](../research/spec-evidence/sound-table-and-playback-rules.md#t01-effective-sound-cells). They do not establish runtime compiled-table correspondence. Arithmetic vectors use supplied transition state without assigning a table column to an endpoint or duration.

### 3.1 Lifecycle state and external operations

- Request cue, optional emitter, delay, playback offset, variant-selection bypass, automatic-fade bypass and remaining positional flags. Cue ids and signed comparisons remain 32-bit; decoded priority is unsigned 8-bit, offset/delay/counter arithmetic wraps to 32 bits.
- Selected client-unit reference and its two-word selection seed; a valid seed-bearing selected unit is required on every reached random draw, while null equality alone remains valid for the request-priority comparison; ordered 200-voice pool, ordered active list, handle counter and ordered bidirectional emitter/voice associations. Unit installation and wider seed lifetime remain STP-Q01, including the shared seed-lifecycle question NRF-Q08; no player identity is inferred merely from pointer equality.
- Valid prepared sound records and table storage, ordered group history and family identity; Volume, Group Size, Priority, Loop, Fade In/Out, Defer Inst, Stop Inst, Duration, Compound, range, async-only, tracking, interrupt, stream/stereo, cache, asset state/failure, reference count, raw/parsed payload and lengths. Field names are semantic; encoded layouts belong in the companion.
- Sixteen ordered slots, each backend class, voice association, valid wrapper, marker and cached properties; marked-slot search/routing-mix control (this control does not gate the complete property-update invocation); retained routing-mode state, distinct from backend acoustic-conversion mode. Block 1/2/3 are signed 32-bit decoded fields; the loader counts consecutive values before the first -1, capped at 3, into an unsigned-byte block count. The loop-start input aliases Block 1; it is not independent state. Optional acoustic control is a supplied decoded value. Positional contract supplies finite distance key, radius, tracking results, volume/property calculations and their existing gates.
- Audio enable, current enabled-update counter, cached master/music values, designated cue ranges, periodic random-cue producer enable and its exact effects, modal/interrupt gain state and last-gain tick. Worker milliseconds, stream-pump inhibition and request interleaving are separate supplied inputs.
- Ordered queued/current requests and completion notifications; request deadline, signed 32-bit priority, source offset/length, unsigned 32-bit cache/affinity key and cookie, raw/device destinations, stream/file references, completion event/metadata and free-after control. Affinity enable, previous archive/key, archive sector size/base, millisecond clocks, bandwidth, throttle numerator, last dispatch time and exit condition are distinct supplied state. The throttle-numerator producer is outside the supplied consumer boundary; no byte-rate meaning is inferred.
- Fatal-handler state: terminationRequestControl is unsigned 32, with its equality-to1 early-termination condition; ordered termination/finalizer callback sets, optional managed-runtime module presence and returning diagnostic/synchronization operations are supplied inputs. The selected OS termination operation consumes an unsigned 32 code and does not return when so supplied. terminationCallbacks supplies the termination callbacks; initializerFinalizerCallbacks=[] means both cleanup ranges are empty. For STP-R05 the caller fixes signed−1, preserving all 32 bits as 4294967295 at that boundary; Origin engine code, Binding fixed by original code. No process-status observation is inferred.
- Asset budget, accounted bytes, LRU order, scan/failure times, pending counter, descriptor ownership/flags, events and IO results. Prepared stream format/source bytes, direct/archive eligibility, source-block metadata, requested capacity, active chunk configuration, scratch allocation state, device readiness, buffer ownership and every reached creation/scanner return are distinct inputs. External allocators, file/codec/OS/device operations supply exact returns, byte counts, regions and data; normal returns and fatal termination are distinct outcomes.

Decoded Volume/Priority/Group Size/Fade In/Fade Out/Loop/Stream/Stereo/Cache/Defer Inst/Stop Inst/Async Only/Tracking/Interrupt/acoustic selectors are unsigned bytes. Range/Falloff is unsigned 32-bit and consumes the full S3D-R02 selector domain. Duration is unsigned 16-bit; Compound is signed 16-bit. Block 1/2/3 are signed 32-bit. Asset sizes, offsets, references, counters, flags and handles use unsigned 32-bit storage unless a rule explicitly interprets a value as signed. All retained arithmetic uses low32 except the stated byte-priority wrap. Device result signs use signed 32-bit interpretation. Exact operations and exceptions are defined by the numeric foundation.

Fixed engine bindings include the 200-entry logical voice pool, sixteen playback slots, the duplicate/path cue intervals 2934–4656 and 4657–4698 (the audibility music interval remains separately supplied), periodic producer modulus 3/range 1723/base 2934, pending scan 1–2933, resource age 750, scan interval 25, pending cap 15, retry interval 250, event-pool cap 1024, default chunk 16384, Sound-worker wait 50 ms, selector urgency threshold 1000 ms, I/O no-request wait 250 ms, merge length cap 1048576 and affinity wrap increment 134217728. Named error values are invalid parameter 87, invalid data 13 and invalid stream state0x85100072; these are engine-fixed numeric bindings. Their Origin is engine code and Binding fixed by original code; these are independent of supplied table values. Decoded table fields have Origin table data and Binding resolved from data; the block count and family/history state are derived/retained engine state. Synthetic records in vectors do not assert archive cells.

Spatial key/radius/scalar calculations consume [S3D-R02–R06](sound-3d-position-and-falloff.md). Complete tracking/context, option/duck scaling, acoustic cache/no-op and property application consume S3D-R07 and remain under S3D-Q01–Q03. A vector may supply the complete finite acoustic-update profile explicitly; that verifies STP caller ordering and retention conditionally, not the S3D integration. Generic allocators, archive codecs, file/error callbacks, OS event/thread/read operations and device methods are exact returning providers where supplied; their internals and unprovided failures are outside this consumer contract. The enabled Sound playlist consumes the separate music boundary and is explicitly disabled in worker fixtures.

## 4. Rules (execution order)

**STP-R01** Named sound cells [verified in 1.13c data]. The three named cursor_pass cells in section 3 have the stated raw values. No playback meaning follows solely from their names.

**STP-R02** Pending volume transition [verified in 1.13c code (static)]. At the supplied-input transition stage:

1. If transition active is false, leave current volume unchanged.
2. If unsigned current tick is greater than unsigned end tick, set current volume to end volume and clear transition active.
3. Otherwise, if unsigned current tick is less than unsigned start tick, set current volume to start volume and retain transition active.
4. Otherwise set current volume to `start volume + trunc_toward_zero(((end volume - start volume) * (current tick - start tick)) / (end tick - start tick))`, retaining transition active. Subtraction, multiplication and final addition retain their low 32 bits; division interprets the resulting numerator and denominator as signed 32-bit values.

The valid supplied-input interpolation boundary requires a nonzero signed denominator and a representable signed quotient. A zero denominator or signed division overflow has no normal volume output in this contract. Clock wrap is not normalized into a relative interval. No random draw occurs.

**STP-R03** Group selection and playback admission.

- **STP-R03a** Group selection [verified in 1.13c code (static)]. With stable table/seed inputs, return the base cue unchanged and make no draw when the sound table is absent; no record access is required on that branch. When the table is present, require a supplied valid base sound record; Group Size 0 or 1 also returns the base without a draw. Otherwise let exclusion count be `min(Group Size-1,2)`. For Group Size 2 or 3 only, draw `Rand(sound-selection seed,Group Size+1)` using NRF-R01/NRF-R02; decrement exclusion count if that result is zero. Repeatedly draw `Rand(sound-selection seed,Group Size)`, form `low32(base cue + draw)`, and reject that candidate if it equals any of the first exclusion-count entries of `[most recent choice, older choice]`. Return the first candidate not rejected. Every attempted draw advances the same seed, including rejected candidates. Selection does not modify history. There is no draw-count cap or fallback candidate.
**STP-R03b** Request, admission, duplicate policy and logical lifetime [verified in 1.13c code (static)]. The following supplied-input lifecycle consumes the spatial/dependency gates listed in section 3 and the Id map.

### Request creation and associations

Reject a request when audio is disabled, cue is nonpositive or the valid cue record has zero Volume. This is not a missing-record fallback: invalid table/record access has no normal rejected-request result established by this path.

For nonzero Compound, find the first nonretiring voice in the same family. Negative Compound accepts that voice without an age test; positive Compound accepts it only when the wrapping unsigned age is no greater than that value. Reuse bypasses new-voice initialization. Otherwise seek the first free voice in the fixed 200-voice pool. If none exists, return zero without changing the handle counter, active list, voice state or associations. If allocation succeeds, clear its stored state, mark it in use and prepend it to the active list. Increment the handle counter with 32-bit wrap and assign that value; do not skip zero or repair collisions. Install cue, due tick as current counter plus delay, flags and supplied playback offset; leave the separate saved resume bookmark cleared. Set current volume 255 and decoded priority. A nonnull emitter initializes position through S3D; null uses position (0,0,320).

For both reuse and creation, emitter reference equality with the selected client-unit reference adds 80 modulo 256 to priority, including null equality. This is applied on each reuse, not once per voice. For a nonnull emitter, prepend the association to the voice and attach its handle to the emitter. Do not silently deduplicate repeated associations. Return the resulting handle; zero after wrap is not repaired.

### Update ordering and stopping

When client coordination is disabled, return without producer, getter or voice work. An enabled coordinator first runs the optional periodic random-cue producer. If that producer is enabled and the unsigned current audio counter modulo 3 is zero, draw Rand(selected-unit seed,1723) using NRF-R01/NRF-R02, then request cue 2934+draw with null emitter, delay 0, flags 0 and playback offset 0. The draw precedes request rejection, so a zero-Volume record can reject the voice while retaining the seed advance. Nondivisible counters make no draw/request. Its purpose is not inferred from this enable input. This producer is distinct from the ambient/environment controller called by the outer audio dispatcher. After the optional producer, sort active voices by priority descending, distance key ascending, start tick descending and fixed pool index descending. The positional contract owns nonfinite-key behavior; these full cases use finite keys.

Before updating any voice, visit every retirement request. A voice with an active zero-target transition bypasses this stop prepass. Otherwise a pending voice becomes stopped and a playing voice is stopped through the backend/slot bridge. Do not treat fade completion later in this invocation as having already occurred at this prepass.

After the retirement prepass, sample the client master setting once, even for an empty active list, and retain whether its signed value is positive. At the start of each sorted-voice visit, resolve its record and form audibility from that retained condition; if its cue is in the supplied inclusive music-cue interval, read the music setting for this visit and also require it to be nonzero. This per-visit read precedes the stopped-state branch.

For each sorted voice:

1. A stopped voice must have no slot. A nonretiring looping voice with zero Duration resolves an active transition to its endpoint, clears that transition and becomes pending. It does not start again in this invocation. Every other stopped voice is prepended to the deferred-removal list. Continue to the next voice.
2. For other states, tracking is entered only when both the global tracking enable and decoded Tracking are nonzero and the reference list is nonempty. An empty list bypasses it; a null reference value on a traversed node takes the fatal path. Complete reference selection/position/context production consumes S3D-R07 and remains an explicitly gated integration under S3D-Q01–Q03. With the supplied resulting state, continue to STP-R02. Then resolve the radius from the record for the voice cue entering this coordinator visit, and retain that radius for the rest of this visit even if group choice changes the cue.
3. A pending voice whose due tick is still in the future or whose volume is zero remains pending without entering failed-admission cleanup. Otherwise the already formed per-visit audibility value must be true. Pending distance eligibility requires an ordered key no greater than radius squared: greater or unordered fails.
4. Duplicate-family filtering uses an emitter filter only for a sole associated emitter. Start with no candidate. When either Defer Inst or Stop Inst is set, select the earliest-starting other playing family voice, retaining the first on equal ticks. Independently of both flags, for an incoming cue in 2934 through 4656 inclusive with a sole emitter, the first other playing voice in that whole inclusive interval containing the sole emitter replaces that result even when the interval-wide lookup returns no candidate. Preserve that replacement order.
5. When Defer Inst is set, the voice has no emitter references and any position coordinate is nonzero or unordered, bypass duplicate actions and continue admission. Otherwise, if there is no candidate, continue admission. With a candidate and Stop Inst set, mark the old voice for retirement and prepend it to deferred removal, then continue incoming admission. With a candidate and Stop Inst clear, retire and queue the incoming voice, skipping admission. Deferral is not a promise to keep it pending.
6. A saved bookmark is transferred to the playback offset and cleared, selecting an automatic three-update fade. Otherwise Fade In from the record entering this coordinator visit may select an automatic fade when no transition is already active and the caller has not bypassed it. Save entering volume and set current volume to zero before admission. On success, attach the returned slot, set playing and directly write transition active=true, start volume=current volume (zero), end volume=saved volume, start tick=the coordinator tick argument and end tick=low32(tick argument+the chosen duration). These direct automatic-fade writes do not invoke STP-R05 or reapply a minimum from the selected record. On failure, do not restore the zeroed volume or install that transition.
7. After an admission attempt, resolve the record again from the voice's possibly changed cue. This selected/current record supplies Loop, Async Only and asset state for failure retention and Duration for any playing check. A failed admitted attempt, or failed audibility/range test, queues a slotless nonlooping voice for removal unless async-only is set and its asset is pending. Looping voices remain. The exception does not cover an unloaded asset or a failed loaded asset. It does not apply to the earlier future-due/zero-volume bypass.
8. A playing voice must have a slot. Nonzero Duration expires only when unsigned wrapping age is strictly greater than Duration; equality remains playing. Ineligible audibility or the strictly ordered key>radius-squared test stops it; an unordered key does not take this playing-stop branch. Pending eligibility instead fails on an unordered key. Otherwise apply positional/backend properties; for a loaded record store the current global audio counter as its access tick and move its cue to the LRU tail. The coordinator tick argument and that global counter are distinct supplied values; STP-R06 normally passes the entering counter as the argument. Stopping saves the backend query result in the slot-associated voice bookmark, stops playback, sets stopped and clears both slot associations.

After all voices have been visited, process deferred removals in prepend order. Remove the first active-list occurrence, remove each handle from each emitter's list and clear voice associations, stop a still-playing backend, then mark the voice free. Consequently replacement can start the incoming backend before the old backend is stopped. Clear the temporary deferred list after processing, including its empty case. After cleanup, derive the modal condition as mode equal to 0 or 1 and either of the two supplied modal inputs nonzero. Derive the interrupt condition from surviving active voices: at least one record has Interrupt nonzero and either no active transition or a nonzero transition endpoint. Only when unsigned tick>last-gain tick update gains. If modal condition is set, duck gain becomes signed max(low32(gain-5),0); otherwise it becomes signed min(low32(gain+5),100). If interrupt condition is set, interrupt gain becomes signed max(low32(gain-2),70); otherwise it becomes signed min(low32(gain+2),100). Always assign last-gain tick=tick, even when tick did not increase. These writes follow this invocation's property applications.

The subsequent backend-marker bridge belongs after the coordinator in STP-R06 and scans slots in their original order. Empty slots are skipped. A record with Loop nonzero is skipped before reading its marker or stream predicate. For a nonloop record, a nonzero stored marker retains playback. With marker zero and no attached stream, set the logical voice stopped and clear both slot associations without querying a bookmark or stopping the device; retain its existing bookmark, owners and in-use state. With marker zero and an attached stream, first run the ordinary query/save/stop/detach operation instead. Logical deletion or loop restart is handled only by a later coordinator invocation. A marker is retained engine state, not a fresh OS playback query.


### Actual group, suppression, slot and start chain

Admission retains the record associated with the voice cue at this admission entry. This is its entering group/history record, not a forced lookup through Family Base and not necessarily the cue of an earlier user request. Unless caller flag bit 0 bypasses selection, run STP-R03a and replace the voice cue with the returned selected cue. Keep that cue and every seed advance even when suppression, loading, slot selection or starting later fails. When the marked-slot search enable input is zero, treat this search as absent. Otherwise search all sixteen slots whose stored marker is nonzero and whose associated voice has the selected cue; choose the smallest unsigned voice start tick, retaining the earliest slot on equal ticks. Do not filter retiring voices out of this selection. No selected slot means no suppression. A selected retiring voice also bypasses suppression, without a fallback search. If the selected slot exists and its associated voice is not retiring, low32(incomingStart-oldStart)<2 unsigned, and old priority>=incoming priority unsigned-byte, return failure before history commit. Otherwise commit STP-R03c to the entering base record; do not move that write to the selected record.

For a nonstreamed selected record not already loaded, request asset admission with synchronous completion exactly when Async Only is zero, and age gate clear. If it is still not loaded, fail. For both streamed and buffered records, a set asset failure flag rejects. Decode requested slot class from Stereo: Stereo nonzero chooses class 1 without mode queries. Otherwise begin class 0, read the retained routing-mode state and OR2 only if it equals 1, then read that state again and OR6 only if it equals 2. Preserve both read calls and ordering; with no intervening writer both reads expose the same retained value. This routing-mode state is distinct from the backend acoustic-conversion mode. Invoke STP-R03d using this class.

With a returned slot, prepare its backend before association. For buffered data set the logical Loop value; when the decoded block count equals 1, set the backend loop-start offset to low32(Block 1*2). Load the selected record's raw payload and size through the buffered load wrapper. The caller checks the wrapper result and would fail on zero. The original load wrapper's ordinary returning path produces 1 even when its initial fill returned failure; null/invalid parser inputs may instead terminate. Do not invent a normal negative-Lock-to-load-failure return. With streamed data this buffered load stage is bypassed.

Then temporarily attach incoming to the slot. If class contains component 4, submit the decoded optional acoustic control as 1.0 when nonzero, else 0.0, through its existing property dependency. Clear cached XYZ and factor, set cached base volume=-1, then apply positional/backend properties. For buffered data invoke STP-R04a. For streamed data build the selected source path and start with the voice playback offset and decoded Loop. A start failure clears the tentative slot-voice association and returns no slot; it does not roll back history, loading, property/cache or earlier displacement effects. Success returns the slot; the caller installs the reverse association and playing state.

Source-path construction selects DATA\LOCAL for cues 2934..4656 inclusive, otherwise DATA\GLOBAL. Append \MUSIC\ for cues 4657..4698 inclusive, otherwise \SFX\, then the selected record's filename. Require prefix-plus-filename length strictly less than 80 before copying the filename; this path's overlength handling is fatal, not truncation. Filename storage and destination are valid supplied strings/buffer. The constants are original engine code, not inferred from cue names. File content/read/decode effects remain the named source/provider boundary.

### Explicit owner removal

Removing an emitter first removes its first matching handle association, then removes its first reference from the first matching active voice. Remaining references prevent retirement even when force is true. On the last reference, force retires; without force, nonlooping playback may finish, while looping playback retires unless a zero-target fade is already active. A direct handle retirement request sets retirement and schedules Fade Out only for a playing voice with a positive decoded Fade Out. Automatic unit-kind-specific cleanup routing remains an integration question under STP-Q01; the supplied owner/handle/force removal operation is defined here without generalizing its callers.

- **STP-R03c** History commit stage [verified in 1.13c code (static)]. At the admission stage after its early recent-voice suppression test has passed, set the original base record's older choice to its entering most recent choice, then most recent choice to the supplied selected cue. No draw occurs. These writes precede asset loading and later slot/device admission; a later admission failure does not roll them back within this path. STP-R03b defines the admission caller reaching this stage; STP-Q01 retains its wider context/seed-lifetime gate.
- **STP-R03d** Slot selection and displacement [verified in 1.13c code (static)]. The selector receives an incoming allocated voice, a requested backend class and the ordered sixteen-slot pool. Voice priority is unsigned 8-bit; start ticks and pool indices compare unsigned. The distance key is a finite `binary32` value supplied by the positional contract. The normal boundary requires valid buffer contexts and returning backend operations.

  First return the earliest slot with no associated voice, exactly matching backend class and buffer-start marker zero. The marker is STP-R04a's stored state, rather than a fresh operating-system playback query. If no such slot exists, consider only occupied slots with exactly matching class. Select the victim by lowest priority, then greatest distance key, then earliest start tick, then lowest voice-pool index. No eligible victim means failure without writes. Reject an incoming voice with lower priority than the victim; at equal priority accept it only when its pool index is greater than the victim's. Incoming distance and start tick do not participate in this final acceptance test.

  On accepted displacement, when a buffer exists, query its resume position and store the result in the old voice, then stop its backend playback. Set the displaced voice state to stopped, clear its associated slot, and clear the slot's voice association. Return the released slot. The incoming voice is not attached by this selector; admission performs that later. No random draw occurs. Pointer ordering in the original selects the stated order of the fixed, ascending-address voice pool; heap allocation order is not an interchangeable substitute.

**STP-R04** Device playback, loop and fade integration.

- **STP-R04a** Buffered start wrapper [verified in 1.13c code (static)]. With a supplied valid buffer context and returning device operations, enter the audio device critical section, request repeating playback from the underlying buffer, and examine the signed 32-bit device result. A negative result leaves the buffer-start marker unchanged, exits the critical section and returns 0. A nonnegative result sets the buffer-start marker to 1, exits the critical section and returns 1. No RNG draw occurs. Critical-section and device execution are supplied boundary operations; this wrapper does not verify their internal effects. It does not branch on the sound-table Loop field.
**STP-R04b** Asset, buffer, stream and completion lifetime [verified in 1.13c code (static)]. The following returning-provider and controlled-interleaving boundary defines local policy; S3D and external provider integration retain their explicit gates.

### Asset admission and accounted state

A preexisting failure flag or loaded state returns without a new load. Pending state returns unless synchronous completion is requested; other nonzero asset states also return without a new load. Open-for-size failure sets the persistent failure flag. The budget test adds used bytes and file size with 32-bit wrap. If that result exceeds budget unsigned, request free bytes equal to wrapping used minus budget plus file size. Eviction computes its unsigned target as wrapping used minus requested free and immediately succeeds when used is already no greater than target. A request greater than used bytes can wrap the target, so the nominal budget is not a hard cap.

LRU eviction scans from the head. Its first pass skips entries with nonzero reference count or Cache; a synchronous fallback pass may ignore those two protections. Both passes retain the optional age gate low32(currentCounter-lastAccess)>=750 unsigned and skip cues with a marked slot from the enabled marked-slot search. Free at most one qualifying entry per pass before continuing the outer progress loop; return failure when no pass makes progress and used remains above the unsigned target. Disposal of an unloaded asset returns without writes. Otherwise a marked slot for the cue takes the fatal diagnostic path. Subtract the retained size with low32 wrap, remove the first matching LRU entry and clear its access tick, then destroy a nonnull descriptor or release raw storage. Finally clear raw/descriptor references and set asset state unloaded. It retains failure and parsed metadata, and disposal of pending assets does not itself decrement the client pending counter.

Synchronous file reads allocate low32(fileSize+800) bytes but account only fileSize. Require nonnull allocated storage, nonzero fileSize and a successful read Boolean with exact actual count before loaded-state validation; these invalid read/storage paths terminate rather than return an empty valid sound. Validation failure retains allocated/accounted storage while setting failure. Async descriptor creation is different: a null returned descriptor leaves the asset unloaded without new size/accounting or pending-count changes; descriptor existence does not separately prove a successful underlying read. Finalizing a pending descriptor sets loaded, validates and decrements the pending counter even on validation failure. With an absent sound table the pending-loader returns. Otherwise a nonzero tick first requires unsigned low32(tick-lastScan)>=25 and otherwise returns without writes; tick zero bypasses that interval. Scan exactly cue 1..2933 with adequate valid storage, regardless the supplied table-count bound. For an unloaded record with a nonzero reference count or Cache, tick zero requests synchronous loading and bypasses pending/backoff guards. A nonzero tick starts an asynchronous load only when pendingCount<15 and unsigned low32(tick-lastMemoryFailure)>=250. Pass age gate true to either load attempt. After the possible start, poll any pending descriptor and finalize when ready. At loop end store lastScan=tick. Eviction failure records lastMemoryFailure=currentCounter without setting the record failure flag. Descriptor readiness/wait/cancel semantics retain their exact source/provider dependency and are not invented from this client counter.


### Asynchronous descriptor and event ownership

A descriptor is a valid checked object with ready, payload-owned and linked conditions, file owner, pool, payload/size, read priority, event node and optional callback metadata. Creation requires a returning nonnull descriptor allocation; null allocation takes unrecoverable diagnostic handling, without an ordinary open-style return-failure guarantee. It clears allocated state, opens the source and retains the supplied priority/pool. Open failure frees the descriptor and fails. A nonzero requested source offset is sought; exact seek result -1 terminates; other values do not take that failure test. Positive signed requested size is used; otherwise query size. A supplied payload is borrowed; absent payload is allocated and marked owned. Copy optional callback metadata as supplied.

Only request-async and global-async-enabled together select asynchronous reading. Acquire an event node; node acquisition failure closes the file, frees owned payload and descriptor, then fails. Event acquisition first claims the first reusable node. Otherwise it polls existing event handles with zero timeout, finalizes the first signaled node's owner and reclaims that node. If none can be reused and event count is at least 1024, fail acquisition; otherwise append a node and increment the event count before creating a new manual-reset, initially unsignaled event. A null CreateEvent output is stored without an additional local node-acquisition failure test. Store the descriptor as node owner, attach the node and event handle, then request the read with payload/size, overlapped state, priority and optional metadata. Ignore the underlying read Boolean and return the descriptor. The synchronous descriptor path likewise ignores the read Boolean, invokes its optional completion callback and finalizes before returning. A descriptor or ready flag is not proof of a successful read or valid WAV.

Poll on null descriptor returns false. A valid already-ready descriptor returns true without waiting. Otherwise perform Wait(event,0); exactly result 0 finalizes. Every nonzero wait result leaves it not finalized by this branch, then return current ready state. Explicit waiting polls first; if still pending, raise queued requests matching its current payload-address cookie to priority 1 and wait infinitely. A nonzero infinite-wait result takes the fatal path; result 0 finalizes. The payload getter waits if needed, requires ready afterward and returns payload. The size getter validates and returns size without waiting. Priority is distinct from the payload-address cancellation/matching cookie.

Finalization is idempotent for an already-ready descriptor. For an attached event node, request ResetEvent and ignore its result, mark the node reusable with no owner and detach it from the descriptor; retain the pooled event handle. Drop and clear the descriptor's file owner, then mark ready. If callback-produced replacement payload is nonnull, free the old current payload through the pool-aware free operation, install replacement pointer/size and clear the callback metadata. This replacement is not conditioned on payload-owned. It is a supplied metadata effect, without an invented codec producer.

Destruction validates and unlinks a linked descriptor, then polls. If pending, cancel queued requests matching its payload-address cookie, wait infinitely and finalize. Unlike explicit waiting, destruction ignores the infinite-wait result. Queue cancellation visits queued requests only: it does not cancel a currently selected OS read or clear its destination. A current read can retain the payload until completion signals the event, as the joined disposal case supplies. A failed disposal wait does not establish completed I/O or safe memory lifetime.

After readiness handling, an owned payload must be nonnull; free it, clear its pointer and invalidate the descriptor's checked marker while preserving its low ready/owned/linked conditions, before freeing the descriptor. A borrowed payload is not freed and that marker invalidation is skipped; the descriptor is still freed. Finalizing a descriptor closes/releases its file ownership and recycles its event node, but does not decrement the client's pending-asset counter. Client finalization and client disposal keep their separately stated counter policies.


### Backend preparation and buffer recreation

Preparation requires a live wrapper and inner buffer with no attached stream. Enter the sound lock. With buffer reuse enabled, clear source reference, source length, played/written counters, end-data, logical Loop and loop-start fields. Retain the marker, half flag, end marker, backend fade state and existing interfaces. With reuse disabled, save wrapper flags, stop through the wrapper, release an existing property interface, then spatial interface, then inner buffer in that order, clear their references and clear all wrapper semantic state. Recreate using the saved flags.

The ordinary nonspatial Sound factory requests PCM tag 1,22050 samples/second,16 bits/sample; channels are 2 when saved flags mask 1 is set, otherwise 1. Block alignment is channels*2, average byte rate 22050*alignment; capacity is 262144 for two channels,131072 for one. Factory control options begin0xa0, add0x8000 when its separate option input is nonzero, then OR0x14 for spatial mask 2 or0x40 otherwise. This Sound factory checks the signed factory result; it does not use the Storm creator's output-pointer-only decision. Ordinary successful recreation stores capacity/half capacity, saved flags and end marker-1. Spatial interface acquisition and property-capability results consume the explicit S3D integration boundary; no default successful capability is supplied.

Preparation ignores the factory helper's Boolean result and continues. Therefore a negative factory result with an explicitly valid nonnull output can leave capacity/half/flags/end-marker zero from the clearing stage while the following calls still execute. A null or released output has no safe normal continuation guaranteed here; no clean failure return or complete unwind is invented.

After either reuse branch, request physical position 0. For the nonspatial, interface-absent path request pan 0, set property scalar 0, store requested volume 255 and submit its converted backend volume through S3D-R06. Device results are ignored. Existing/new spatial interfaces add their S3D-owned setup/property operations; those integration outputs remain gated. Leave the lock. This operation itself does not run a worker, query status, make a random draw, start Play or sample a clock. Successful preparation does not clear a retained reuse-path marker or fade by an invented invariant repair.

### Buffered bridge and failures

The buffered load wrapper requires a live wrapper/inner buffer, nonnull source bytes and nonzero source length. Enter synchronization and call the WAV parser with the source length as its initial in/out length; do not branch on the returned pointer. Require unsigned loop-start<data length as returned or retained in that in/out value, regardless logical Loop; otherwise take the fatal diagnostic path. Then store the parser pointer and length and invoke initial fill from source offset 0 to destination 0 for the entire buffer capacity. Ignore the fill return. Always set halfFlag1; set endMarker=dataLength exactly when Loop is clear and unsigned dataLength<capacity, otherwise-1. Leave synchronization and return 1 on this normal path.

A null parser result is not guaranteed normal success: the preceding length assertion or the fill's null-data assertion can terminate. A parsed zero-length data chunk with loop-start0 fails the preceding strict comparison. With valid parsed data and a returning negative device Lock, fill returns failure without copying PCM, while the outer normal epilogue still returns 1 and installs its later bookkeeping. Preserve unchanged buffer bytes, counters/end-data, marker and later Play outcome independently; no initial progress reset is invented inside the load wrapper. Unlock results are ignored on the copied path. Buffered repeating Play success sets the start marker; negative Play leaves it unchanged and reports failure through the wrapper in STP-R04a.

Logical Loop is not equivalent to the repeating device flag. The separate worker tracks PCM progress, fills/pads and completes nonlooping data; device-position failure and exact equality conditions require explicit source/case assertions. A successful request/load/Play wrapper is not proof of audible output.


### Exact worker completion, recovery and independent fades

The worker waits on its event for 50 milliseconds and performs an iteration only on timeout; every other wait result exits it. Each iteration enters the sound lock, samples the game-window iconic state, then visits the ordered live-wrapper list. For each wrapper it invokes the buffer-status recovery operation before any changed-iconic volume update, backend fade, stream query or buffered progress operation. A negative status result does not by itself skip the later worker stages: its return is ignored here.

Recovery requires a live inner buffer. A nonnegative status with buffer-lost bit clear returns normally. With that bit set, repeatedly request Restore until a nonnegative result. Result 0x88780096 sleeps 10 milliseconds before retry; other negative results retry immediately. No retry bound is established. Success invokes the stop wrapper, which ends an attached stream when present, requests device Stop and clears the marker. It retains the wrapper and inner buffer; Stop result is ignored.

When backend fade is active, sample its separate unsigned storage as signed 32-bit milliseconds. Within the inclusive signed interval [start,end], compute from + trunc_toward_zero(low32((target-from)*(now-start))/(end-start)), preserving the active flag. Normal interpolation requires nonzero signed denominator and representable signed quotient. A time before start OR after end applies the target immediately and clears the active flag. If target is zero on that outside-interval branch, first end an attached stream if present, then request Stop and clear the marker. Exact end at target zero only sets volume zero: the fade remains active and does not itself stop the buffer. Subsequent stream/buffer stages still execute. Volume conversion is an explicit dependency on the established acoustic/backend conversion contract; a stored scalar is not a device attenuation value.

An attached stream is queried first. A failed stream lookup, including its finished state, ends the stream through the wrapper and then makes an additional unconditional Stop request before clearing the marker. Nonnegative device results are not required for these marker/state writes. With no stream, marker zero skips position work. Otherwise request the physical byte position; a negative result skips progress/refill/completion. For a successful position result, let remainder=oldPlayed & (capacity-1), base=low32(oldPlayed-remainder); if unsigned physical<remainder, add capacity to base; store played=low32(base+physical). This is a bit mask, not an interchangeable general modulo operation.

With end-data set and logical Loop clear, physical<halfCapacity and halfFlag zero requests clearing the second half then sets halfFlag one; physical>=halfCapacity and halfFlag one requests clearing the first half then sets halfFlag zero. Set the half flag even if Lock fails. The clear helper enters the nested sound lock and requests Lock for that half. A negative result makes no copy or Unlock. A successful result zeroes precisely the returned first-region byte count and calls Unlock for that region; it ignores Unlock result and does not request or process a second region. Only after these half tests, unsigned played>dataLength requests Stop and clears the marker. Equality retains the marker. Loop bypasses this end-data completion branch.

With end-data clear, compute gap=low32(written-played). Refill when unsigned gap<halfCapacity OR unsigned played>written; equality gap==halfCapacity with played<=written does not refill. Request halfCapacity bytes from source offset=written into ring destination=written & (capacity-1). Ignore the fill return, then always set endMarker=-1 and halfFlag=(saved destination!=0). A helper that sets end-data does not cause same-iteration completion: that branch is not revisited.

The fill helper requires a valid wrapper, live inner buffer, nonnull source and nonzero requested length. A negative Lock returns failure before payload, written or end-data changes. On nonnegative Lock, copying/padding uses the requested length, while Unlock uses the returned first-region length. This differs from half clearing, which zeroes the returned length. Valid supplied regions must cover the actual writes; a short returned region does not shorten this fill's copy policy.

For each fill segment, if unsigned source offset>=dataLength, remap it to low32(loopStart+((sourceOffset-loopStart) modulo (dataLength-loopStart))); this remapping is independent of logical Loop. Reached division requires nonzero denominator. Choose the copied count using the original signed comparison between remaining requested bytes and low32(dataLength-sourceOffset). Copy source bytes, add the copied count to written, source offset, destination counter and destination pointer, and subtract it from remaining requested count with low32 arithmetic. With Loop and remaining bytes, repeat. Without Loop, pad the remaining region with zeros: endMarker=-1 permits the full remaining count; otherwise a destination counter>=endMarker skips padding, while a lower counter selects low32(endMarker-destinationCounter) bytes without a second clamp to the original remaining count. Padding does not increment written. Unlock the first region, ignore its result, then for nonloop set end-data when unsigned written>=dataLength. Return success. Invalid source metadata, signed copy lengths and insufficient returned storage do not gain a safe fallback.

A separate Sound playlist may run before wrapper iteration when enabled and its dedicated wrapper has no attached stream. Its selection belongs to the excluded music-selection boundary. Its separately allocated wrapper joins the shared worker list; therefore disabling the playlist is an explicit case input, not proof that it can never run during gameplay. Wider UI enable reachability and enabled-playlist integration remain named dependencies, with no source-independent defaults.

### Streams and ownership

A stream start opens its source before entering the sound lock, scales the supplied bookmark by four with 32-bit wrap, and passes the loop flag independently. Stream parsing/preparation, the supplied buffer versus created-buffer ownership policy, chunk configuration and file/stream reference increments are part of this boundary. Query-produced bookmarks lose only the byte position's low two bits on divide/rescale; arbitrary supplied bookmarks can also wrap high bits.

The stream service's inhibition input must be zero for pumping. Initial request, pending start, playing/refill, nonloop drain and finished states remain distinct. Callback startup uses the request's requested byte length, not its destination offset, when testing whether enough data has been requested. Successful Lock and required region length govern callback writes; a negative result or wrong length has its actual no-update path. Preserve copy count, silence padding, request/file/ring cursor arithmetic and ignored Play result separately.

Loop EOF resets source position and progress before refill; nonloop EOF drains and requests Stop before recording the finished state under the actual position/length conditions. The Stop result is ignored, so finished state does not prove physical device success. For each stream belonging to the file being ended, first clear a matching current in-flight request's device destination, then cancel queued requests whose device destination equals this stream's buffer, then unlink and drop the stream's own reference. These matches use buffer identity and are not restricted by the request's file or stream identity. Clear the file stream flag after traversing its streams and return 1 even when the list contained no matching stream. An in-flight request may keep file/stream resources alive after wrapper marker/handle clear. Request destruction drops its file/stream references; final stream destruction drops its file reference and releases the device buffer only when it owns that buffer. A borrowed wrapper buffer survives stream-node destruction. These are ordered ownership effects, not a claim that a stop request synchronously frees every resource. When canceling a queued request with a nonnull completion event, first allocate and enqueue a separate notification node carrying that event, then destroy and free the request. Cancellation does not itself signal the event. The I/O worker drains notification nodes before stream pumping and request selection: request SetEvent, then unlink and free the node regardless that result. A failed event signal is consumed without retry. Final stream destruction drops its file reference before requesting Release for a nonnull owned buffer, then frees stream storage; a nonzero returned device reference count does not prevent those engine writes. Ending a stream through the stream service does not itself request device Stop or close a separately retained external file owner.



### Stream request construction and I/O worker

Use this request/worker contract for the explicitly serialized stream and descriptor partitions. Supply a valid file, valid ordered lists, returning synchronization and allocation/free providers, exact clock/read/event results, and every reached source-routing field. A reached stream callback requires a valid attached stream; invalid request objects gain no safe fallback. The competing-request selector, affinity score, priority-update merge and nonzero-bandwidth throttle below are owned engine policy, not external providers. Existing concrete partitions use an empty queue or one selectable request and bandwidth zero; an initial pump may construct several requests before a separately bounded selection/completion checkpoint. Those examples do not claim branch-exhaustive competing-queue execution.

### Request construction and references

A request requires either a raw read buffer or a device-buffer destination. If both are null, fail with invalid-parameter error before publishing the request or incrementing references. Otherwise enter stream synchronization, allocate a zero-initialized request, append it to the queue tail and sample the supplied millisecond clock. Record the supplied completion event, file, source offset, priority, raw buffer, cancellation identity, device destination and its offset, optional stream, requested source length, free-after-completion condition and optional completion metadata. Deadline is low32(clock+delay); delay is not an unconditional wait instruction. With a direct OS handle, the cache/affinity key equals source offset. With no direct handle, key=low32(supplied archive file-record base+(unsigned sourceOffset>>1)); the file record supplying that base must be valid. This key is not substituted for the logical source offset or claimed to be the exact decoded byte position.

Increment the file reference once and the stream reference once when that stream is nonnull. Do not substitute a device AddRef for these engine references. Leave synchronization. Request SetEvent on the worker wake event only when the supplied wake condition requests it, ignoring that returned Boolean; do not infer a wake from mere queue insertion. Construction does not read source bytes, run a stream callback, start playback or release its references.

### Ordered worker iteration and selection boundary

The I/O worker samples its clock at entry and at the start of an iteration. It derives throttle zero when bandwidth is zero; no throttle division is reached in that case. Enter stream synchronization and first drain queued completion notifications in their stored order. For each, request SetEvent, then unlink/free its notification node regardless the returned Boolean. No failed-signal retry or replacement is introduced. Only afterward run the stream pump and request selector. Pump inhibition, stream state and queue contents remain live supplied inputs; an inhibited pump does not create work.

### Complete request selection policy

Sample the selector clock once and traverse the valid queued requests in original order. If a request has a nonzero cancellation cookie and its immediately preceding queued request exists with the same cookie, skip this request for selection. This is an adjacent-predecessor condition, not a global deduplication or cancellation. A zero cookie never takes this skip. The request remains queued.

For each remaining candidate compute delta=int32(low32(deadline-selectorClock)). Selection has these descending classes:

1. Attached stream and signed delta<1000.
2. Signed priority>=1, when no class 1 candidate exists.
3. Signed delta<1000, when neither higher class exists.
4. Remaining candidates, when none of those classes exists.

Within each of the first three classes choose the smallest signed delta; an exact tie retains the first traversal candidate. Priority magnitude does not further rank class 1, class 2 or class 3. A new higher-class candidate replaces a lower-class selection regardless its relative delta. The selector returns the chosen pointer separately from urgency: urgency is true for classes 1–3 and false for class 4 or no selection. A future deadline may belong to an urgent class when its signed delta is below 1000, and signed wrapped/overdue deltas follow the same comparison.

In class 4, choose signed priority descending, then affinity score unsigned ascending, then signed delta ascending; exact ties retain the first candidate. When the affinity-enable input is zero, every score is 0. Otherwise let A be the request file's archive identity. Start base 0; if A is nonnull and unsigned previousCacheKey>that archive's sectorSize, set base=low32(previousCacheKey-sectorSize). If A differs from the previously dispatched archive identity, score=0xfffffffe. If they match, including both null, let k=request.cacheKey: when unsigned k>=base, score=low32(k-base); otherwise score=low32(k+0x08000000). The below-base branch does NOT subtract base as well. Direct files share null archive identity for this calculation; do not add a separate file-identity comparison. Both archive metadata and caller-supplied affinity controls remain distinct from codec/OS internals.

### Complete worker throttle and wait policy

At worker entry initialize previousArchive=null, previousCacheKey=0 and lastDispatchTime=entryClock. At each loop sample loopClock. If bandwidth=0, throttle=0; otherwise throttle=unsigned floor(low32(throttleNumerator*1000)/bandwidth), where throttleNumerator is the current stored global input. The source body reads that input; this rule does not invent its producer as an audio byte counter. A nonzero divisor is required when reached.

After notification drain, stream pump and selection, an urgent chosen request bypasses throttling. A nonurgent chosen request dispatches only when unsigned low32(loopClock-lastDispatchTime)>=throttle. If below, leave it queued, leave synchronization and wait on the wake event for low32(throttle-loopClock+lastDispatchTime) milliseconds. With no chosen request, leave synchronization and wait250ms. Returned wait values do not themselves select Sound-worker exit behavior; the I/O worker then checks its exit input. On a supplied nonzero exit input, clear it and return; otherwise loop. The entry path also samples its entry clock before testing an already-set exit input, clears that input and returns without an iteration.

Immediately before actual dispatch, retain chosen request cacheKey and file.archive as previousCacheKey/previousArchive and retain loopClock as lastDispatchTime. This is dispatch-start time, not a newly sampled completion time. Nonzero bandwidth and future deadlines therefore require these separate inputs, not an assumed fixed cadence. With bandwidth 0 the throttle test admits any selected class 4 request, including the singleton deadline 3600 at constant clock 1000. Delay is not an unconditional pre-read sleep.

With a selected request, publish it as current and unlink it from the queued list while synchronized, then leave synchronization before source I/O. Being current is distinct from remaining queued. Queue-only cancellation therefore does not abort a selected OS read. Stream end can clear a matching current device destination without changing an already chosen read target or canceling that source read.

The notification-only example explicitly supplies an exit-condition write at its wait return. This is not an inferred eventual-stop or event-delivery guarantee, and the I/O wait policy is distinct from the Sound buffer-worker's50ms timeout-only iteration policy.

### Priority changes and compatible queued-request merging

A priority-update operation requires a nonzero cancellation cookie; a null cookie writes invalid-parameter engine error 87 and requests SetLastError(87), without traversing the queue or requesting the wake event. Under stream synchronization, visit queued requests in order and assign the supplied signed priority to every matching cookie. Remember the most recently visited matching request, even if other nonmatching requests intervene. If the new priority is at least 1, a previous matching request exists, and the current request's OLD signed priority was below 1, try merging that previous match into the current one. No merge attempt is made from this wrapper for an old priority already at least 1. Keep the current request as the remembered match afterward. After traversal leave synchronization; the wrapper requests SetEvent on the worker wake event for new priority>=1, even when no request matched, ignoring that event result.

The merge requires equal cookies; low32(earlier.sourceOffset+earlier.length)==later.sourceOffset; unsigned low32(earlier.length+later.length)<=1048576; and earlier.freeAfter nonzero. Both actual-read counts, both attached-stream references and both device destinations must be zero/null. The earlier completion event and earlier completion-metadata reference must also be null. Do not add checks for later.freeAfter, later completion event/callback, equal file pointers, or raw-pointer contiguity that the source does not perform. Supplied valid buffer storage must cover the resulting requested write; arbitrary invalid combinations gain no safe fallback.

On an accepted merge copy earlier sourceOffset, cacheKey, deadline and rawReadBuffer into the later request, replace later length with the wrapping sum, then destroy/unlink/free the earlier request using the ordinary file-then-stream reference destructor. Other later fields, including its file identity, completion event/metadata, updated priority and freeAfter, remain. A failed gate changes no merge fields and frees neither request; the outer priority write still occurred. This is a restricted source transformation, not a generic queue reorder or a promise to abort current I/O.

### Read target, completion dispatch and cleanup

At read dispatch, a nonnull device destination selects the shared scratch buffer; otherwise use the request's raw buffer. Use the selected file/source offset/requested length with the original file routing and supplied actual I/O results. A null selected read target records actual count zero without manufacturing source bytes. Direct ReadFile and archive routing remain distinct; read Boolean, exact byte count, optional error policy and metadata effects retain their named contracts. A descriptor or queued request does not independently prove successful or complete I/O.

After the read, inspect the request's current device destination again. A nonnull destination and nonzero actual count dispatch the stream callback. A nonnull destination, zero actual count and positive requested length instead request device Stop; when a stream is attached, write its finished state, ignoring the Stop result. Requested length zero skips that Stop/state write. A cleared destination skips both callback and zero-read Stop. This second check is why a read can complete into retained scratch storage after stream end while making no device copy or state write. Optional completion callbacks, when present, consume their explicitly supplied metadata contract; the selected stream and disposal examples supply none.

Save the completion event, then reenter stream synchronization. If free-after-completion is set, destroy the request: first drop its file reference and perform file cleanup/free if that reaches zero; then drop its stream reference and perform stream cleanup/free if that reaches zero; unlink the request and free its storage. Final stream cleanup itself drops its retained file reference before requesting Release for a nonnull owned device buffer. Borrowed buffers receive no such Release. If free-after-completion is clear, skip request destruction/free; the live unqueued request retains its references. Clear current only after this conditional cleanup, then leave synchronization. Signal a saved nonnull completion event afterward, even though the request storage may already be freed; do not reconstruct the freed request to deliver it.

On a supplied exit condition at the worker tail, clear that condition and return. Thread creation, actual scheduling, successful device/event effects and safe memory after a failed external wait are not implied by these state transitions. Preserve the descriptor rule's checked explicit wait versus unchecked destruction wait, and its pooled-event ownership, separately.


### Ordinary stop versus stream-wrapper end

An ordinary client stop first queries the slot's backend resume position when its wrapper and associated voice are present and stores that result in the slot-associated voice's bookmark. An absent wrapper stream returns bookmark 0 without a device-position query. With an attached stream, the stream query/cache policy in this rule applies. A failed stream lookup returns bookmark 0 without a physical-position query; on successful lookup divide the returned unsigned byte position by 4. Then invoke the general buffer-stop operation, set the associated voice stopped and clear both slot associations. Keep owners and in-use state until an owning cleanup stage removes them.

General buffer-stop enters the sound lock and requires a live inner buffer. If a stream is attached, first perform stream-wrapper end: end/unlink the stream service, request device Stop once, close the wrapper's external file owner, then clear wrapper stream and marker. After that conditional operation, general buffer-stop unconditionally requests device Stop again and clears the marker. Thus an ordinary stream stop with an attached stream requests two Stops, while an ordinary buffered stop requests one. Both results are ignored. The wrapper/inner buffer persist; stream-wrapper end is distinct from buffer/interface destruction.

With an attached stream, direct stream-wrapper end used by the Sound worker and the in-flight-stop case makes only its own one Stop; with no wrapper stream it only enters/leaves synchronization, leaving marker and device unchanged. The worker's later unconditional Stop can add another. The nonloop stream service's finished transition can have made an earlier Stop as well. Preserve the calling path and its ordered count: do not collapse general buffer-stop, direct stream end, stream-service end and resource Release into one operation.

### Stream creation, preparation and failure effects

Begin with a valid file owner. Requested capacity zero returns failure before entering the stream lock or ending an old stream; device-not-ready likewise returns failure at this early point. Otherwise enter the lock and end any preexisting stream on this file before checking new-source eligibility. Failure later does not reinstate it. Eligibility requires direct-source permission in either global/input bit 0, or an archive-backed file with no direct OS handle. A failed eligibility check leaves the lock and fails.

A zero configured chunk size becomes 16384; copy the configured value to active chunk size. Allocate scratch only if absent, retaining any existing scratch without resizing or checking its capacity here. The allocation return is stored without a local null check. For an archive with mode 0 or 1, write mode 2. Leave the lock. Normalize requested capacity using r=requestedCapacity & low32(chunkBytes-1); when r is nonzero, add low32(chunkBytes-r). If the resulting capacity is unsigned less than low32(2*chunkBytes), replace it with that two-chunk value. This is a mask rule; equivalence to ordinary rounding needs power-of-two chunk and valid arithmetic inputs.

Clear the two file read/check metadata fields before parsing. Direct files seek to byte 0; that initial seek result is ignored before the next read. Archive files reset/clamp their logical position and clear cached read state. Require a successful twelve-byte header read, RIFF and WAVE tags; do not use the RIFF size as a parser bound. Find fmt by repeated eight-byte chunk-header reads, skipping unknown chunks by exactly their declared length without odd-byte padding. Require fmt length>=16, read its first 16 bytes, skip the rest and require that relative seek not return the failure sentinel. Copy those 16 format bytes and set extra-format size 0. Find data by the same scan, retaining its declared length and body position. There is no local PCM-tag, channel-count, sample-rate, average-rate, block-alignment or bit-depth compatibility gate. Device/codec validity is distinct from passing this parser.

A nonzero supplied byte offset is reduced by unsigned modulo declared data length; zero offset bypasses that division. A nonzero offset with zero length does not acquire a clean failure return: division is invalid. Zero average rate likewise may pass creation but fails a later reached pump division. Chunk scanning uses returning exact-read and seek results; generic archive decoding, optional read-error callbacks and allocator internals remain explicit provider/dependency limits.

A supplied device buffer is borrowed: no factory, AddRef, creation volume or creation pan calls. With no supplied buffer, request a buffer with normalized capacity and parsed format. Descriptor flags begin0x80, become0xa0 when input mask0x20 is set, then OR0x40 when that mask is set. Ignore the factory result sign and branch solely on the output buffer pointer: null fails, nonnull continues even with a negative result. For the created buffer only, submit raw device volume unless it is 0 or 2147483647, then raw pan unless it is 0 or 2147483647; ignore both results. These arguments are independent of the Sound scalar-to-attenuation converter.

Reenter the stream lock and allocate/link a zero-initialized stream; its allocation is not locally null-checked. Set file, average bytes/second, Loop from input mask0x40000, state initial, data start/length, normalized start/cached/progress offset, volume/pan inputs, buffer, normalized capacity and owns-buffer condition. Set silence byte 128 exactly for bits/sample 8, otherwise 0. Initial write cursor and silence tail remain 0; initial stream reference count is 1. Increment the file reference count once. Request physical position 0 and ignore its result; set the file's stream flag, then leave the lock. Invoke worker bootstrap and signal its wake event, ignoring those results. Bootstrap with an existing worker skips creation; without one it stores CreateEvent/CreateThread outputs and requests thread priority 1 without checking those returns. Begin returns success without synchronously running the pump, reading PCM or requesting Play. Worker existence/running, provider success and audible output do not follow from that return.

Parse/scanner/relative-seek failures record invalid-data error and fail before a new stream/file reference is published. Factory null output fails without assigning that parse error. Both retain prior chunk/scratch/archive/file-metadata/seek effects and the open external file owner; no rollback or owner close occurs inside this creator. After a factory yields a nonnull buffer there is no later checked failure branch providing Release-on-failure unwinding. Invalid allocator/provider state may terminate; do not invent a safe cleanup return.

The Sound caller opens its source before entering its own sound lock. Creator failure causes that caller to close the newly opened file owner and return failure, leaving the wrapper marker/stream field unchanged. Creator success stores the file in the wrapper, sets marker 1/end-marker-1 and returns success. This wrapper path normally passes its existing buffer, distinct from the generic creator's optional owned-buffer branch.

### Stream pump and callback progression (valid prepared stream inputs)

Use separate stream states: initial, pending-start, playing, draining and finished. A nonzero pump-inhibition input returns without processing streams. Iterate the original ordered stream list; stream-list membership and positive-handle domain remain valid supplied state. Capacity, chunk bytes and average byte rate must make each reached unsigned division valid. A physical-position query failure returns a negative sentinel and leaves cached position unchanged. On success, let delta=low32((progress modulo capacity)-physicalPosition); if delta interpreted signed is nonpositive, add capacity. Cached position becomes unsigned min(low32(progress-delta),dataLength). The helper returns physical position, separately from this cache write. Stream lookup can therefore succeed and expose a retained cached value even when the underlying physical query failed.

An initial stream sets file logical position=dataStart+startOffset. Let total=min(capacity,low32(dataLength-startOffset)), scheduledDestination=0 and delay=0. In source-block alignment mode, first set A=min(chunkBytes,total). From supplied block metadata compute r=low32(filePosition+A) & (sectorSize-1); if 0<r<A, reduce A by r. Queue a special request at destination 0 for A bytes with delay 0, advance file position by A and set scheduledDestination=A. This special first request does not advance delay. Then, while scheduledDestination<total, queue an ordinary request at that destination for min(chunkBytes,total-scheduledDestination) source bytes and the current delay. Advance source file position by requested bytes; advance scheduledDestination by a whole chunk; add unsigned floor(low32(chunkBytes*1000)/averageBytesPerSecond) to delay. Without alignment mode there is no special first request and the ordinary loop starts from destination 0/delay 0. Finally store ring write cursor=scheduledDestination modulo capacity and record pending-start. Block metadata and reached division validity are explicit inputs; the special first trimmed size is not interchangeable with a whole chunk. Pending-start receives no automatic retry from this pump.

In playing state, first test file logical position against dataStart+dataLength. At EOF, Loop resets source position to dataStart and resets progress and silence tail to zero, then continues playing-refill processing. Without Loop, query physical position. A query failure skips this stream. Cached position>=dataLength requests device Stop and records finished regardless Stop result. Otherwise compute free ring bytes as low32(physical-writeCursor) modulo capacity. When at least one chunk is free, request clearing min(chunkBytes,capacity-writeCursor) bytes at the write cursor. A failed Lock changes no counters or ring cursor; a successful Lock fills the returned first region with the stream silence byte, ignores Unlock result, adds returned bytes to progress and silence-tail counters, and advances write cursor with a single capacity subtraction when needed. Whether insufficient space or Lock failure, the successful-query below-end branch records draining. Draining repeats this same EOF handling, rather than normal source reads.

In ordinary playing refill, a successful physical query and at least one free chunk schedule one source read. Requested byte count is the unsigned minimum of chunkBytes, contiguous capacity-writeCursor and remaining source bytes, with optional source-block adjustment. Delay is unsigned low32((capacity-2*chunkBytes)*1000) divided by averageBytesPerSecond. Advance file position by requested bytes and ring write cursor by a whole chunk, subtracting capacity once if needed. Request admission returns and provider effects are not replaced by assumed device success.

At worker read completion, a nonnull device destination with actualReadBytes>0 dispatches the stream callback. With a nonnull destination, actualReadBytes zero and requested source length positive, the worker instead requests device Stop and records finished when a stream is attached; it does not call Lock or the padding callback, and ignores the Stop result. A cleared destination skips both callback and this zero-read stop path. Subsequent request destruction and retained-reference cleanup still occur. These worker dispatch conditions are separate from the direct callback contract.

The callback enters its stream lock and requests one chunk at the request's destination offset. A negative Lock or returned first-region length unequal to chunkBytes exits with no PCM copy, Unlock, progress/tail or state write. On accepted Lock, copy min(actualReadBytes,chunkBytes) from worker scratch, pad the remainder with the supplied stream silence byte, and Unlock the first region, ignoring that result. Set progress=low32(requestFileOffset-dataStart+returnedFirstRegionLength); if unsigned progress>dataLength, set silence tail=low32(progress-dataLength). Do not reset silence tail on the opposite branch. For pending-start only, compare low32(requestFileOffset-dataStart+requestedSourceLength) against min(dataLength,65536,capacity). Reaching the threshold requests repeating Play and records playing regardless the Play result. Below threshold retains pending-start. Request destruction and reference release still happen after callbacks returning without an update; pump pending-start handling does not requeue them.

- **STP-R04c** Loaded WAV validation [verified in 1.13c code (static)]. For a sound record whose asset state is loaded, parse the supplied asset bytes as follows. Require at least 32 bytes, the `RIFF` signature at byte 0 and `WAVE` at byte 8. Starting at byte 12, scan chunk headers, each containing a four-byte id and unsigned 32-bit payload length. Skip an unmatched chunk by exactly that length, without adding an alignment byte. A missing header, or an unmatched payload extending past the remaining bytes, fails. Find `fmt ` first; require payload length at least 16 and no greater than the remaining bytes. Copy its first sixteen bytes into the format fields and set the extra-format-byte count to zero. Starting after that entire format payload, find `data` by the same scan and require its declared length to fit. Success returns the data payload and its declared length; neither the RIFF length field nor format-tag value is validated on this path.

  The client additionally requires bits per sample 16 and sample rate 22050. A parser failure or either rejected field sets the record's failure flag and returns failure. Success stores the PCM payload reference and data length, assigns Stereo true exactly when the parsed channel count equals 2, and returns success. It does not reject other channel counts here. Success does not clear an entering failure flag. This overwrites the decoded Stereo field: a text-table Stereo cell alone does not establish the buffered routing class. The supplied boundary requires a loaded record. STP-R04b defines local asset/async and streamed-parser policy; external I/O/codec/device-format outcomes retain STP-Q02. No draw occurs.

**STP-R05** Schedule a volume transition [verified in 1.13c code (static)]. Inputs are an ordered voice list, requested handle, target volume, requested duration, delay and current tick. All tick/duration arithmetic retains low32; timing comparisons are unsigned. Target/current volumes are signed 32-bit values.

1. If handle is zero or no voice has that handle, return without writes. Choose the first matching voice. If its emitter association list has a second entry, return without writes.
2. Starting from requested duration, if target is 255 raise duration to at least unsigned Fade In; if target is 0 raise it to at least unsigned Fade Out. Other targets have no table duration minimum.
3. If a transition is active and its end volume equals target, compute candidate end=`low32(current tick+delay+duration)`; if candidate end is unsigned greater than or equal to the existing end tick, return without writes. Otherwise continue. If an active transition's end volume is 0 and the requested target is nonzero, return without writes.
4. If target is zero, set the voice's retirement-request condition true. Other targets do not clear this condition.
5. When duration is zero, a nonzero delay takes fatal diagnostic termination after the preceding target-zero retirement write and before assigning volume or schedule fields. With delay zero, set current volume to target and return; do not clear or overwrite existing transition state.
6. With positive duration, set transition start tick=`low32(current tick+delay)`, end tick=`low32(start tick+duration)`, start volume to entering current volume, end volume to target and active=true. Current volume is not changed by scheduling.

The reached fatal scheduling branch reports signed 32 code−1 to its fatal-handling foundation. With terminationRequestControl not equal to 1, empty termination/finalizer callback sets, returning diagnostic/synchronization operations and the optional managed-runtime module absent, that foundation sets terminationRequestControl=1 and requests ordinary OS process termination with unsigned 32 code 4294967295. The supplied termination operation does not return. terminationCode denotes that received code argument, not a measured OS exit status. If terminationRequestControl equals 1, the handler first requests immediate termination of the current process with the same code; a supplied nonreturning operation ends that path before callbacks. Other callback/module outcomes retain their explicitly supplied foundation boundary and are not assigned defaults here. No request normalizes the invalid delay into an immediate-volume update.

No draw occurs. The valid supplied boundary requires a valid associated sound record whenever a matched voice reaches the duration policy. This rule schedules state consumed by STP-R02; STP-R03b consumes retirement requests; positional volume application retains its named S3D integration gate.

**STP-R06** Audio update counter [verified in 1.13c code (static)]. When client audio updating is enabled, one dispatcher invocation passes the entering unsigned 32-bit audio counter to the ambient/environment controller, music-selection controller, pending-load processor and voice updater in that order. It then processes backend-marker retirement and increments the counter by one with low32 wrap. When updating is disabled this dispatcher neither calls these stages nor increments the counter. The voice updater consumes this counter for STP-R02; request creation adds delay to it, and STP-R05 schedules against it. These durations therefore count enabled audio updates. This rule does not assign a wall-clock rate to dispatcher invocations. The backend worker's separately sampled millisecond clock is not this counter.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| STP-R01 | Named archive cells | verified in 1.13c data | none | T01 |
| STP-R02 | Supplied transition arithmetic | verified in 1.13c code (static) | none for valid supplied inputs | STP-V01–STP-V07, STP-V30, STP-V32 |
| STP-R03a | Group choice with repeat rejection | verified in 1.13c code (static) | stable supplied seed/history; wider seed lifecycle STP-Q01 | STP-V08–STP-V11, STP-V67 |
| STP-R03b | Request, duplicate, admission and logical lifecycle | verified in 1.13c code (static) | STP-Q01 wider context/automatic cleanup callers; STP-Q02 and S3D-R07/Q01–Q03 gate full acoustic/tracking integration; supplied provider policy remains available | STP-V28–STP-V35, STP-V37–STP-V39, STP-V44, STP-V47, STP-V53–STP-V56, STP-V61, STP-V64, STP-V66–STP-V68 |
| STP-R03c | History commit | verified in 1.13c code (static) | supplied stage or R03b caller; wider context STP-Q01 | STP-V12, STP-V44, STP-V61, STP-V67 |
| STP-R03d | Sixteen-slot selection and displacement | verified in 1.13c code (static) | finite supplied key; returning backend operations; S3D key production separate | STP-V20–STP-V22, STP-V28 |
| STP-R04a | Buffered start result | verified in 1.13c code (static) | exact supplied device returns; STP-Q02 real device behavior | STP-V18, STP-V19, STP-V35, STP-V44 |
| STP-R04b | Asset, buffer, worker, stream and ownership lifecycle | verified in 1.13c code (static) | STP-Q02 external file/codec/device/thread providers, enabled playlist and spatial factory/property integration; controlled interleavings only | STP-V28, STP-V34–STP-V36, STP-V40–STP-V60, STP-V62, STP-V63, STP-V65, STP-V66, STP-V69 |
| STP-R04c | Loaded WAV and client format validation | verified in 1.13c code (static) | loaded supplied bytes; R04b provides local loading policy; external source/device STP-Q02 | STP-V23–STP-V26, STP-V43 |
| STP-R05 | Transition scheduling | verified in 1.13c code (static) | valid record; supplied fatal-foundation boundary; external device/termination effects STP-Q02 | STP-V13–STP-V17, STP-V39 |
| STP-R06 | Enabled audio-update counter | verified in 1.13c code (static) | controller internals have separate owners; wall-clock cadence unassigned under STP-Q02 | STP-V27, STP-V47 |

## 5. Outputs and side effects

| Output | Rule |
|---|---|
| current volume assigned on active transition branches | STP-R02 |
| transition active cleared only after current tick exceeds end tick | STP-R02 |
| returned group cue and updated sound-selection seed; history unchanged | STP-R03a |
| most recent and older choice committed in the original base record | STP-R03c |
| scheduled start/end ticks, endpoint volumes, active condition; immediate current-volume write; retirement-request condition, fatal-handler request control and received unsigned 32 termination-code argument under the stated conditional foundation boundary; delayed-zero-duration no-normal-return outcome | STP-R05 |
| critical-section and repeating-device-play request log; buffer-start marker set on nonnegative result; Boolean wrapper result | STP-R04a |
| selected/released slot; displaced voice resume position, stopped state and cleared associations | STP-R03d |
| loaded format acceptance, PCM payload and length, overwritten Stereo or retained/set failure flag | STP-R04c |
| ordered stage requests and one enabled-update counter increment | STP-R06 |

STP-R03b additionally writes request handles, allocation/in-use state, ordered active/deferred/owner associations, cue/seed/history, priority, voice state/slot/bookmark/offset, retirement/automatic transition state, loaded access tick/LRU and modal/interrupt gains. It requests preparation, source loading and backend start/stop in the defined order. Named ordered caller-operation/read logs and read-only call-time snapshots record those observation points without extra engine mutations. Acoustic dependency results are conditional on their supplied profile or owning contract.

STP-R04b writes asset state/failure/raw/parsed references/accounted bytes/LRU/pending and scan/failure counters; descriptor ready/owned/linked state, payload/file/event ownership and pooled-node state; wrapper source/progress/half/end/loop/marker/fade/interface state; source positions, configured/active chunk and scratch state; stream state/caches/cursors/progress/tail; request queues/current/notifications/deadlines, selected request and urgency, priority updates and merged source-offset/cache-key/deadline/raw-buffer/length fields, previous archive/cache key and last-dispatch time, wait/exit state, and ordered file/stream/device reference releases. It writes the stated engine errors and requests external error updates, returns the defined local results or reaches the stated diagnostic termination. Exact source/device bytes and operation logs are asserted only for explicitly supplied providers. Requested Stop, Play, event or factory calls do not establish their physical success.

## 6. Branches and edge cases

End-tick equality retains an active transition; the next greater tick clears it. Interpolation decreases as well as increases volume, truncating negative division toward zero. Inactive transitions do not overwrite volume. Zero-duration interpolation, overflow, missing records, unavailable sound devices and invalid sound identifiers do not acquire guessed fallback behavior. STP-R05 states the fatal delayed zero-duration branch; STP-R03b/R04b distinguish normal rejection, retained state and fatal continuation.

## 7. Open questions

**STP-Q01** [unresolved] Wider selected-unit installation/seed/reseeding lifetime and automatic unit-kind cleanup callers. This is an integration/outside-boundary question: it does not gate the supplied stable-seed request/admission/duplicate policy or explicit owner/handle/force removal operation in STP-R03b. No initial seed or player-only identity is inferred.

**STP-Q02** [unresolved] Complete original-game playback integration. Full tracking/context and acoustic-update cache/option/duck/property results consume S3D-R07 under S3D-Q01–Q03; supplied finite profiles do not verify those dependencies. Generic archive codecs/read-error callbacks and diagnostic/termination foundations, external allocation/event/thread/OS read/device behavior, enabled playlist selection and spatial factory-capability operations require their named provider/owner results. The client enabled-update counter and worker milliseconds are separate; dispatcher wall-clock cadence is not assigned. These limits block full reproduction, while the exact returning-provider policy, controlled lifecycle and fatal argument branches remain available.

**STP-Q03** [unresolved] Matching gameplay observations for the verified behavioral assertions. No STP executable vector is runtime-confirmed. This observation gap does not erase static supplied-input results. See the [attempts and closure conditions](../plans/gaps/sound-table-and-playback-rules.md).

## 8. Test vectors

### 8.1 Executable vectors

Each case has independent state and no omitted persistent operations. STP-V01–STP-V07 and STP-V12–STP-V19 assert exactly zero random draws and an empty draw log. Group vectors log every draw and preserve the final seed. History is supplied in most-recent/older order. Scheduler vectors start from the explicitly listed voice; fields omitted by a branch remain unchanged. Device-wrapper vectors log every supplied critical-section/device operation; those operations have no other stored effects beyond returning the stated device result.

- **STP-V01** Inactive transition (STP-R02; boundary: supplied-input stage)
  - Inputs: active=false; current tick=15; start tick=10; end tick=20; start volume=0; end volume=255; current volume=73.
  - Seeds: no draw.
  - Draws: none.
  - Trace: inactive branch leaves volume untouched.
  - Outputs: current volume=73; active=false.
  - Discriminates: unconditional endpoint assignment would replace 73.
  - runtime_observed: no
- **STP-V02** Before transition (STP-R02; boundary: supplied-input stage)
  - Inputs: active=true; current tick=9; start tick=10; end tick=20; start volume=10; end volume=110; current volume=73.
  - Seeds: no draw.
  - Draws: none.
  - Trace: 9 is not greater than 20 and is less than 10.
  - Outputs: current volume=10; active=true.
  - Discriminates: extrapolation would produce 0.
  - runtime_observed: no
- **STP-V03** Falling interpolation (STP-R02; boundary: supplied-input stage)
  - Inputs: active=true; current tick=11; start tick=10; end tick=13; start volume=10; end volume=0; current volume=73.
  - Seeds: no draw.
  - Draws: none.
  - Trace: difference=-10; elapsed=1; duration=3; quotient=-3; volume=10-3=7.
  - Outputs: current volume=7; active=true.
  - Discriminates: floor division would produce 6.
  - runtime_observed: no
- **STP-V04** Exact end (STP-R02; boundary: supplied-input stage)
  - Inputs: active=true; current tick=20; start tick=10; end tick=20; start volume=0; end volume=255; current volume=73.
  - Seeds: no draw.
  - Draws: none.
  - Trace: 255*10/10+0=255; strict greater-than test is false.
  - Outputs: current volume=255; active=true.
  - Discriminates: greater-or-equal completion incorrectly clears active.
  - runtime_observed: no
- **STP-V05** After end (STP-R02; boundary: supplied-input stage)
  - Inputs: active=true; current tick=21; start tick=10; end tick=20; start volume=0; end volume=255; current volume=73.
  - Seeds: no draw.
  - Draws: none.
  - Trace: unsigned 21>20 selects completion before interpolation.
  - Outputs: current volume=255; active=false.
  - Discriminates: unbounded interpolation would produce 280.
  - runtime_observed: no

- **STP-V06** High-bit tick uses unsigned completion (STP-R02; boundary: supplied-input stage)
  - Inputs: active=true; current tick=2147483648; start tick=0; end tick=20; start volume=0; end volume=255; current volume=73.
  - Seeds: no draw.
  - Draws: none.
  - Trace: unsigned 2147483648>20 selects completion before interpolation.
  - Outputs: current volume=255; active=false.
  - Discriminates: signed current tick comparison treats the high-bit tick as negative and selects before-start volume 0, retaining active.
  - runtime_observed: no
- **STP-V07** Wrapped product uses signed division (STP-R02; boundary: supplied-input stage)
  - Inputs: active=true; current tick=32768; start tick=0; end tick=65536; start volume=0; end volume=65536; current volume=73.
  - Seeds: no draw.
  - Draws: none.
  - Trace: 65536*32768 retains low32=2147483648, interpreted signed as -2147483648; denominator=65536; signed quotient=-32768; final volume=-32768.
  - Outputs: current volume=-32768; active=true.
  - Discriminates: a widened product or unsigned division produces +32768 instead of -32768.
  - runtime_observed: no

- **STP-V08** Singleton consumes no draw (STP-R03a; boundary: supplied-input stage)
  - Inputs: table present=true; base cue=100; Group Size=1; history=[101,102].
  - Seeds: sound-selection seed={lo=0,hi=1}.
  - Draws: none.
  - Trace: Group Size <=1 returns the base without stepping the seed.
  - Outputs: cue=100; seed={lo=0,hi=1}; history=[101,102].
  - Discriminates: drawing modulus 1 would change the seed to {1,0}.
  - runtime_observed: no
  - Independent zero-group variation: reset STP-V08 and set Group Size=0. Return cue 100 with seed{lo=0,hi=1}, history[101,102] unchanged and draw log[]. No generator call occurs; runtime_observed:no.
  - Independent absent-table variation: table present=false, base cue 100, seed{lo=0,hi=1}. Retained history is[101,102]; no record access is reached or required. Return cue 100 with seed/history unchanged and draw log[]. Normal return precedes reading Group Size; runtime_observed:no.

- **STP-V09** Rejected group choices still draw (STP-R03a; boundary: supplied-input stage)
  - Inputs: table present=true; base cue=100; Group Size=2; history=[101,100].
  - Seeds: sound-selection seed={lo=0,hi=1}.
  - Draws: (sound-selection seed,3,1), (sound-selection seed,2,1), (sound-selection seed,2,1), (sound-selection seed,2,0).
  - Trace: exclusion count starts at 1 and stays 1 after preliminary result 1. Candidates 101,101 are rejected; candidate 100 is accepted. Ordered seed states are {1,0}, {1791398085,0}, {1721382809,747178471}, {986833572,717975634}.
  - Outputs: cue=100; seed={lo=986833572,hi=717975634}; history=[101,100].
  - Discriminates: omitting the preliminary draw or returning the first member fails the exact draw log; committing selection inside this stage changes history incorrectly.
  - runtime_observed: no
- **STP-V10** Small-group exclusion reduction (STP-R03a; boundary: supplied-input stage)
  - Inputs: table present=true; base cue=100; Group Size=2; history=[100,101].
  - Seeds: sound-selection seed={lo=0,hi=0}.
  - Draws: (sound-selection seed,3,0), (sound-selection seed,2,0).
  - Trace: preliminary zero reduces exclusion count from 1 to 0. The fixed-point seed remains zero; candidate 100 is accepted despite being most recent.
  - Outputs: cue=100; seed={lo=0,hi=0}; history=[100,101].
  - Discriminates: always excluding the most recent choice cannot return on these inputs; excluding no history without the preliminary draw fails the two-draw log.
  - runtime_observed: no
- **STP-V11** Larger group has no preliminary draw (STP-R03a; boundary: supplied-input stage)
  - Inputs: table present=true; base cue=100; Group Size=4; history=[101,102].
  - Seeds: sound-selection seed={lo=0,hi=1}.
  - Draws: (sound-selection seed,4,1), (sound-selection seed,4,1), (sound-selection seed,4,1), (sound-selection seed,4,0).
  - Trace: exclusion count=2; candidates 101,101,101 are rejected; 100 is accepted. Seed states are {1,0}, {1791398085,0}, {1721382809,747178471}, {986833572,717975634}.
  - Outputs: cue=100; seed={lo=986833572,hi=717975634}; history=[101,102].
  - Discriminates: applying a Group Size+1 preliminary draw adds a modulus-5 entry that the required draw log rejects.
  - runtime_observed: no
- **STP-V12** Commit uses entering recent value (STP-R03c; boundary: supplied-input stage)
  - Inputs: early recent-voice suppression has passed; original base cue=100; history=[101,102]; selected cue=103.
  - Seeds: no draw.
  - Draws: none.
  - Trace: older=entering recent=101; recent=selected=103.
  - Outputs: original base history=[103,101].
  - Discriminates: writing recent before reading it would make older=103; writing the selected record instead leaves base history unchanged.
  - runtime_observed: no
- **STP-V13** Fade-out duration minimum and delay (STP-R05; boundary: supplied-input stage)
  - Inputs: one voice, handle=7, emitter associations=[], current volume=200, active=false, retirement=false, transition fields start tick=0/end tick=0/start volume=0/end volume=0; request handle=7,target=0,duration=2,delay=3,current tick=100; Fade In=9,Fade Out=5.
  - Seeds: no draw.
  - Draws: none.
  - Trace: effective duration=max(2,5)=5; retirement=true; start=103; end=108; start volume=200; end volume=0.
  - Outputs: current volume=200; active=true; retirement=true; transition fields start tick=103/end tick=108/start volume=200/end volume=0; associations unchanged.
  - Discriminates: ignoring Fade Out yields end=105; applying delay after duration changes start; immediate volume assignment incorrectly changes 200 to 0.
  - runtime_observed: no
  - Independent fade-in-minimum variation: reset STP-V13 and request target 255 with duration 2; other inputs unchanged. Effective duration=max(2,9)=9. Exact outputs: current volume 200, active=true, retirement=false, start tick 103/end tick 112/start volume 200/end volume 255, associations[] unchanged. No draw. Omitting the Fade In minimum would end at 105; runtime_observed:no.
  - Independent invalid-handle variations: reset STP-V13 separately for requested handle 0 and requested handle 8. The sole existing voice has handle 7. Each returns without writes: current volume 200, active=false, retirement=false, all transition fields 0, associations[] unchanged, no RNG/device operation. Draw log[]; runtime_observed:no.

- **STP-V14** Existing same-target transition is not lengthened (STP-R05; boundary: supplied-input stage)
  - Inputs: one voice, handle=7, associations=[], current volume=40, active=true, retirement=false, transition start tick=80/end tick=120/start volume=0/end volume=255; request handle=7,target=255,duration=30,delay=0,current tick=100; Fade In=10,Fade Out=5.
  - Seeds: no draw.
  - Draws: none.
  - Trace: effective duration=30; candidate end=130 >=120, so skip before writes.
  - Outputs: current volume=40; active=true; retirement=false; transition start tick=80/end tick=120/start volume=0/end volume=255; associations unchanged.
  - Discriminates: replacing every request would produce a transition ending at 130.
  - runtime_observed: no
  - Independent equal-end variation: reset STP-V14 and set requested duration=20. Effective duration remains 20; candidate end=120 equals the existing end. Return without writes: current volume 40, active=true, retirement=false, start tick 80/end tick 120/start volume 0/end volume 255 and associations[] unchanged. A wrong strict-greater comparison would reschedule start tick 100/start volume 40. Seed unconsumed; draws[]; runtime_observed:no.

- **STP-V15** Fade to zero resists nonzero retargeting (STP-R05; boundary: supplied-input stage)
  - Inputs: one voice, handle=7, associations=[], current volume=40, active=true, retirement=true, transition start tick=80/end tick=120/start volume=200/end volume=0; request handle=7,target=255,duration=5,delay=0,current tick=100; Fade In=2,Fade Out=5.
  - Seeds: no draw.
  - Draws: none.
  - Trace: target differs, but entering transition endpoint is zero and requested target is nonzero; skip.
  - Outputs: current volume=40; active=true; retirement=true; transition start tick=80/end tick=120/start volume=200/end volume=0; associations unchanged.
  - Discriminates: unrestricted retargeting would replace endpoint zero with 255.
  - runtime_observed: no
- **STP-V16** Immediate set retains pending transition (STP-R05; boundary: supplied-input stage)
  - Inputs: one voice, handle=7, associations=[], current volume=40, active=true, retirement=false, transition start tick=120/end tick=180/start volume=10/end volume=90; request handle=7,target=80,duration=0,delay=0,current tick=100; Fade In=2,Fade Out=5.
  - Seeds: no draw.
  - Draws: none.
  - Trace: target is neither endpoint with a duration minimum; prior target is nonzero and differs. Immediate set writes 80 only.
  - Outputs: current volume=80; active=true; retirement=false; transition start tick=120/end tick=180/start volume=10/end volume=90; associations unchanged.
  - Discriminates: clearing active or replacing interval fields during an immediate set violates the retained-state assertion.
  - runtime_observed: no
- **STP-V17** Multiple emitter associations suppress scheduling (STP-R05; boundary: supplied-input stage)
  - Inputs: one voice, handle=7, associations=[unit A,unit B], current volume=200, active=false, retirement=false, transition start tick=0/end tick=0/start volume=0/end volume=0; request handle=7,target=0,duration=2,delay=3,current tick=100; Fade In=9,Fade Out=5.
  - Seeds: no draw.
  - Draws: none.
  - Trace: first matching voice has a second association; return before duration adjustment or retirement write.
  - Outputs: current volume=200; active=false; retirement=false; transition start tick=0/end tick=0/start volume=0/end volume=0; associations=[unit A,unit B].
  - Discriminates: treating a second emitter as irrelevant would set retirement and schedule an end at 108.
  - runtime_observed: no

- **STP-V18** Successful supplied device start (STP-R04a; boundary: supplied-input stage)
  - Inputs: valid buffer context; buffer-start marker=0; supplied device result=0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: enter lock; request repeating playback; signed result 0 is nonnegative, so marker=1; exit lock; success.
  - Outputs: result=1; marker=1; operation log=[enter audio device lock, request repeating buffer playback, exit audio device lock].
  - Discriminates: treating only strictly positive device results as success would return 0 and retain marker 0.
  - runtime_observed: no
- **STP-V19** Failed start retains old marker (STP-R04a; boundary: supplied-input stage)
  - Inputs: valid buffer context; buffer-start marker=1; supplied device result=-1.
  - Seeds: no draw.
  - Draws: none.
  - Trace: enter lock; request repeating playback; signed result -1 is negative; exit without marker write.
  - Outputs: result=0; marker=1; operation log=[enter audio device lock, request repeating buffer playback, exit audio device lock].
  - Discriminates: clearing marker on failure would write 0; omitting unlock would fail the ordered log.
  - runtime_observed: no

- **STP-V20** Earliest compatible free slot (STP-R03d; boundary: slot selector)
  - Inputs: incoming allocated voice; requested class=7. Sixteen ordered slots: slot 1 unassociated/class=8/marker=0; slot 2 unassociated/class=7/marker=1; slots 3 and 4 unassociated/class=7/marker=0; slots 5–16 unassociated/class=8/marker=0. Every buffer context is valid. The supplied marker getter returns each stated marker without other effects.
  - Seeds: no draw.
  - Draws: none.
  - Trace: skip mismatching slot 1; marker excludes slot 2; slot 3 passes before slot 4.
  - Outputs: return slot 3; every voice/slot/marker unchanged; no position or stop request.
  - Discriminates: selecting any empty slot returns slot 1; ignoring the marker returns slot 2; attaching the incoming voice inside selection changes slot 3 incorrectly.
  - runtime_observed: no
- **STP-V21** Victim ordering and displacement (STP-R03d; boundary: slot selector)
  - Inputs: requested class=7; sixteen occupied slots of class 7, with valid buffers. Slots 1–4 contain respectively `(voice pool index,priority,key,start tick)`=(8,10,5,0), (6,10,20,100), (4,10,20,50), (2,10,20,50). Slots 5–16 have pool indices 10–21 respectively, priority=20,key=100,start tick=0. Incoming voice index=100,priority=11,key=999,start tick=999; it has no slot. Every incumbent is playing and points back to its slot, with saved resume position=0. Supplied position query returns 123; supplied stop returns normally and clears only that buffer's start marker.
  - Seeds: no draw.
  - Draws: none.
  - Trace: minimum priority excludes slots 5–16; key excludes slot 1; oldest start excludes slot 2; lower pool index selects slot 4 over slot 3. Incoming priority 11 exceeds 10.
  - Outputs: return slot 4; its old voice index 2 has resume position=123,state=stopped,slot=null; slot 4's voice=null and marker=0. Incoming voice remains unattached. Other listed state unchanged. Operation log=[query position of slot 4, stop slot 4].
  - Discriminates: nearest-key eviction chooses slot 1; newer-start eviction chooses slot 2; larger-index final tie chooses slot 3; comparing incoming distance during acceptance rejects incorrectly.
  - runtime_observed: no
- **STP-V22** Equal-priority acceptance uses pool order (STP-R03d; boundary: slot selector)
  - Inputs: two independent cases. Requested class=7. Slot 1 contains playing voice index=8,priority=10,key=1,start tick=100, saved resume=0, associated with that slot. Slots 2–16 are unassociated/class=8/marker=0. Incoming voice has priority=10,key=999,start tick=0,no slot; its pool index is (a) 9 or (b) 7. Buffers valid; position query returns 7; stop returns normally, clearing the selected buffer marker only.
  - Seeds: no draw.
  - Draws: none.
  - Trace: slot 1 is the sole eligible victim. At equal priority only the incoming pool index determines acceptance.
  - Outputs: (a) return slot 1; old voice resume=7,state=stopped,slot=null; slot voice=null,marker=0; log=[query position of slot 1, stop slot 1]. (b) return failure; all state unchanged; empty operation log. Incoming voice remains unchanged in both cases.
  - Discriminates: always allowing an equal priority fails (b); preferring lower incoming indices reverses both cases; comparing incoming key or start rejects (a).
  - runtime_observed: no
- **STP-V23** Loaded WAV validation overwrites Stereo (STP-R04c; boundary: loaded record and supplied bytes)
  - Inputs: loaded record, failure=false,Stereo=false, old payload=null,old data length=0. A 56-byte little-endian buffer contains `RIFF`, RIFF length=0, `WAVE`, a `fmt ` chunk of length 16 with `(format tag,channels,sample rate,average bytes/sec,block align,bits/sample)`=(99,2,22050,88200,4,16), followed by a `data` chunk of length 12 and twelve zero payload bytes.
  - Seeds: no draw.
  - Draws: none.
  - Trace: format payload begins at byte 20, data payload at byte 44. Neither RIFF length=0 nor format tag=99 is rejected; client rate/bits checks pass.
  - Outputs: success; payload references byte 44; data length=12; Stereo=true; failure=false. Asset state remains loaded; input bytes unchanged.
  - Discriminates: requiring PCM tag 1 or checking the RIFF length rejects this case; preserving entering Stereo leaves false.
  - runtime_observed: no
- **STP-V24** Unsupported sample rate sets failure (STP-R04c; boundary: loaded record and supplied bytes)
  - Inputs: same buffer/record as STP-V23 except sample rate=44100,average bytes/sec=176400,Stereo=true; existing stored payload references a separate old asset and old data length=8.
  - Seeds: no draw.
  - Draws: none.
  - Trace: WAV parser succeeds; client rejects sample rate before payload/Stereo writes.
  - Outputs: failure result; failure=true; Stereo=true; stored old payload and data length=8 unchanged; asset state remains loaded; bytes unchanged.
  - Discriminates: accepting arbitrary WAV rates incorrectly replaces payload/length and returns success.
  - runtime_observed: no
  - Independent unsupported-bit-depth variation: reset STP-V24; set WAV sample rate 22050, average bytes/sec 44100, block alignment 2 and bits/sample 8. Other bytes/record fields retain that structurally valid input. Parsing succeeds, then the client rejects bits 8. Outputs equal the main failure/retained-payload result: failure=true, Stereo=true, old payload and length 8 retained, loaded state and bytes unchanged. Omitting the16-bit check would replace the payload/length and return success. No draw; runtime_observed:no.

- **STP-V25** Odd skipped chunk has no added padding (STP-R04c; boundary: loaded record and supplied bytes)
  - Inputs: loaded record with failure=false,Stereo=true,payload=null,data length=0. A 57-byte buffer contains `RIFF`, length=49, `WAVE`; `JUNK` length=1 and one zero payload byte; immediately at byte 21, `fmt ` length=16 with format=(1,1,22050,44100,2,16); at byte 45, `data` length=4 with four zero payload bytes. No padding byte follows JUNK.
  - Seeds: no draw.
  - Draws: none.
  - Trace: skip eight header bytes and exactly one JUNK payload byte; find the unaligned format header at 21 and PCM at 53.
  - Outputs: success; stored payload references byte 53; data length=4; Stereo=false; failure=false; asset state remains loaded; bytes unchanged.
  - Discriminates: adding RIFF word-alignment padding misreads the format header and rejects.
  - runtime_observed: no
- **STP-V26** Successful parsing retains an old failure flag (STP-R04c; boundary: loaded record and supplied bytes)
  - Inputs: same valid 56-byte buffer as STP-V23, but channels=3,average bytes/sec=132300,block align=6; record enters failure=true,Stereo=true,payload=null,data length=0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: client checks sample rate and bits, but does not restrict channels to 1 or 2 and does not clear failure on success.
  - Outputs: success; payload references byte 44; data length=12; Stereo=false; failure=true; asset state remains loaded; bytes unchanged.
  - Discriminates: rejecting all channel counts except mono/stereo fails; clearing failure on success changes retained state incorrectly.
  - runtime_observed: no
- **STP-V27** Enabled-update counter wraps after stage requests (STP-R06; boundary: audio dispatcher counter)
  - Inputs: two independent cases with audio counter=4294967295; updating (a) enabled or (b) disabled. Supplied controller/load/voice/retirement operations log their requests and have no stored effects.
  - Seeds: no draw at this dispatcher boundary; supplied operations make no draw.
  - Draws: none.
  - Trace: enabled path passes the same entering counter to each clock-consuming stage, then retires markers before incrementing.
  - Outputs: (a) counter=0; operation log=[ambient/environment update(4294967295), music selection update(4294967295), pending loads(4294967295), voice update(4294967295), backend-marker retirement]. (b) counter=4294967295; empty operation log. This boundary asserts no wider controller effects.
  - Discriminates: incrementing before dispatch passes 0 to consumers; incrementing while disabled changes case (b); treating the counter as sampled milliseconds need not produce either result.
  - runtime_observed: no

### 8.1.1 Lifecycle vector interpretation and logical fixture

Inputs below are synthetic case inputs, not engine defaults. Unmentioned semantic fields remain equal to the fully resolved input or preceding step output. Resolve shared record copies and their local differences before applying a case override. Independent variations restart from their entering fixture; sequential steps carry state. Array order is observable where stated. Dotted override keys name actual shared-fixture paths. voicePool.entries contains the entry states and free-index declaration; count, allocationOrder, activeOrder and nextHandleCounter are separate pool metadata. Replacing entries resets only entry states; metadata changes only through its explicit paths. State names are pending/playing/stopped for logical voices and unloaded/pending/loaded for assets. Each voice has flat activeTransition/startVolume/endVolume/startTick/endTick fields. Its distanceKey is the retained bounded result supplied by S3D-R04; positionDependency.rawSquaredXY is a separate producer input. Their numeric values are both 0 in these cases, without asserting general equality. All tick/counter arithmetic stated as low32 uses unsigned 32 arithmetic; priority addition uses unsigned 8 wrap. Boolean controls may be written false/true or 0/1 with identical meaning.

The periodic random-cue producer is disabled in this shared input; it is distinct from the dispatcher ambient/environment controller.

Verified S3D-R02/R05/R06 supply only their scalar/radius subcalculations. They do not define the full acoustic-update cache guard, option/duck base scaling or resulting cache/device write profile. Those complete finite effects below are an explicit SUPPLIED input under unresolved S3D-R07/Q02 integration. These vectors do not derive or prove its published integration and do not transfer its ownership into STP. File/async providers must also supply their declared results, and no worker interleaving is inferred. Generic floating-point behavior, actual sound production and unspecified dependency results remain unestablished.

#### Shared semantic fixture

There are 200 logical voice entries, allocated in ascending index order, and 16 persistent playback slots scanned in ascending index order. O names pool index 0 and I names index 1; indices 2–199 are free. No invocation of the supplied acoustic profile is inferred in a case that stops before that stage; the absence of such a call is a caller-boundary assertion. A free entry has all scalar voice fields 0, no slot/references and no active transition. Pool-index identity persists across reuse: N0 denotes a newly allocated voice at index 0. Slot references identify a pool entry independently of its in-use flag.

The names requestSelection and ownerA both refer to the single owners.A.seed state; they are aliases, not two RNG states. All named semantic fields are specified below. Other declared scalar controls are 0/false and optional references absent; this is explicit fixture initialization. No statement about padding or opaque storage is part of the semantic contract. No producer/dispatcher controller runs implicitly at a direct coordinator, request, asset or selector entry.

##### Clock, controls, owners and seeds

```json
{
  "clock": {
    "tickArgument": 100,
    "audioCounter": 100
  },
  "controls": {
    "clientAudioEnabled": true,
    "periodicRandomCueProducerEnabled": false,
    "periodicProducerEffects": [],
    "periodicProducerDraws": [],
    "selectedEmitter": "A",
    "trackingEnabled": false,
    "tableCount": 2934,
    "validRecords": "All IDs0–2933 have valid record storage. Unlisted records have all semantic scalar fields0 and absent references, including Cache0/asset unloaded.",
    "specialCueRange": [
      2934,
      4656
    ],
    "masterVolume": 100,
    "musicVolume": 100,
    "modalMode": 2,
    "modalOtherInputs": [
      0,
      0
    ],
    "duckGain": 100,
    "interruptGain": 100,
    "lastScalarTick": 100,
    "markedSlotSearchAndRoutingMixEnabled": true
  },
  "owners": {
    "A": {
      "type": 0,
      "seed": [
        1,
        0
      ],
      "handles": [
        501,
        500
      ]
    },
    "B": null,
    "otherOwners": "none"
  },
  "voicePool": {
    "count": 200,
    "allocationOrder": "ascending index",
    "activeOrder": [
      "I",
      "O"
    ],
    "nextHandleCounter": 501,
    "entries": {
      "O": {
        "inUse": true,
        "cue": 101,
        "handle": 500,
        "position": [
          0,
          0,
          320
        ],
        "distanceKey": 0,
        "currentVolume": 255,
        "contextFactor": 0,
        "playbackOffset": 0,
        "references": [
          "A"
        ],
        "slot": "S0",
        "state": "playing",
        "start": 98,
        "requestFlags": 1,
        "priority": 90,
        "retirement": false,
        "bookmark": 0,
        "activeTransition": false,
        "startVolume": 0,
        "endVolume": 0,
        "startTick": 0,
        "endTick": 0
      },
      "I": {
        "inUse": true,
        "cue": 100,
        "handle": 501,
        "position": [
          0,
          0,
          320
        ],
        "distanceKey": 0,
        "currentVolume": 255,
        "contextFactor": 0,
        "playbackOffset": 0,
        "references": [
          "A"
        ],
        "slot": null,
        "state": "pending",
        "start": 100,
        "requestFlags": 1,
        "priority": 100,
        "retirement": false,
        "bookmark": 0,
        "activeTransition": false,
        "startVolume": 0,
        "endVolume": 0,
        "startTick": 0,
        "endTick": 0
      },
      "freeIndices": "2–199"
    }
  },
  "externalConcurrency": {
    "workerInterleaving": [],
    "otherStateWriters": []
  },
  "seedAliases": {
    "requestSelection": "owners.A.seed",
    "ownerA": "owners.A.seed"
  }
}
```

##### Decoded records and assets

Block1/Block2/Block3 are all 0, so decodedBlockCount3 follows the loader count; only count 1 enables the loop-start setter. loopStartInputAlias names Block1 and is not another value. This supplied decoded state has no claim about original archive cells.

Record101 is a resolved copy of record 100 with the listed differences. Case changes to 100 do not propagate to 101. The PCM references below identify the data-chunk payload, not a binary address. The shared synthetic PCM WAV has a44-byte canonical header: RIFF/WAVE, RIFF size 40, one fmt chunk of length 16 with PCM tag 1, two channels,22050 samples/second,88200 bytes/second, block alignment 4,16 bits/sample, followed by a data chunk of length 4 and bytes[1,2,3,4]. No extra chunks or bytes are supplied. Total file length 48. WAV101 is a distinct allocation with the same file bytes.

```json
{
  "records": {
    "100": {
      "Volume": 255,
      "GroupSize": 2,
      "Loop": 0,
      "FadeIn": 0,
      "FadeOut": 0,
      "DeferInst": 0,
      "StopInst": 1,
      "Duration": 0,
      "Compound": 0,
      "Range": 0,
      "Reverb": 0,
      "Cache": 0,
      "AsyncOnly": 0,
      "Priority": 20,
      "Stream": 0,
      "Stereo": 1,
      "Tracking": 0,
      "Interrupt": 0,
      "Channel": 0,
      "Block1": 0,
      "Block2": 0,
      "Block3": 0,
      "FamilyBase": 100,
      "historyRecent": 7,
      "historyOlder": 6,
      "filename": "fixture100.wav",
      "raw": "WAV100",
      "rawSize": 48,
      "PCM": "WAV100.PCM",
      "PCMSize": 4,
      "lastAccess": 90,
      "failure": 0,
      "descriptor": null,
      "state": "loaded",
      "referenceCount": 0,
      "decodedBlockCount": 3,
      "loopStartInputAlias": "Block1"
    },
    "101": {
      "resolvedCopyOf": 100,
      "differences": {
        "GroupSize": 0,
        "Priority": 10,
        "historyRecent": 0,
        "historyOlder": 0,
        "raw": "WAV101",
        "PCM": "WAV101.PCM",
        "filename": "fixture101.wav"
      }
    }
  },
  "assets": {
    "WAV100": {
      "fileLength": 48,
      "PCM": [
        1,
        2,
        3,
        4
      ]
    },
    "WAV101": {
      "fileLength": 48,
      "PCM": [
        1,
        2,
        3,
        4
      ]
    },
    "usedBytes": 96,
    "budget": 1000000,
    "LRU": [
      100,
      101
    ],
    "pendingCount": 0,
    "lastMemoryFailure": 0,
    "lastPendingScan": 100
  }
}
```

##### Slots, wrappers and supplied backend results

Each slot Sn has a distinct valid wrapper Wn and live inner buffer Bn. Every slot class is 1. S0 references O; S1–S15 are unassociated. S0 caches position(0,0,320), base volume 255 and factor 0. Every other slot cache is position(0,0,0), base volume 0 and factor 0. Slot-wide replacement descriptions set only the stated per-slot/wrapper fields; the slot count 16 and distinct wrapper/buffer identities remain supplied. Every wrapper starts with the common values below, then its listed differences.

```json
{
  "wrapperCommon": {
    "bufferCapacity": 262144,
    "halfCapacity": 131072,
    "threeDInterface": null,
    "propertyInterface": null,
    "bufferFlags": 1,
    "requestedVolume": 255,
    "propertyScalar": 0,
    "streamHandle": null,
    "logicalLoop": 0,
    "loopStart": 0,
    "backendFadeActive": false,
    "backendFadeStartVolume": 0,
    "backendFadeEndVolume": 0,
    "backendFadeStartMs": 0,
    "backendFadeEndMs": 0
  },
  "wrappers": {
    "W0": {
      "marker": 1,
      "data": "WAV101.PCM",
      "dataLength": 4,
      "playedBytes": 0,
      "writtenBytes": 4,
      "halfFlag": 1,
      "endDataFlag": 1,
      "endMarker": 4,
      "payload": "[1,2,3,4] followed by262140 zero bytes"
    },
    "W1to15": {
      "marker": 0,
      "data": null,
      "dataLength": 0,
      "playedBytes": 0,
      "writtenBytes": 0,
      "halfFlag": 0,
      "endDataFlag": 0,
      "endMarker": -1,
      "payload": "byte170 repeated262144 times"
    }
  },
  "backendControls": {
    "reuseBuffers": true,
    "masterScalar": 255,
    "acousticConversionMode": 0,
    "wrappersRegistered": "all supplied wrappers; no worker service interleaving",
    "routingMode": 0,
    "routingModeReadProvider": "Every reached routing getter returns the retained routingMode value0; no writer interleaves. This is the original supplied dependency input, not an inferred default."
  },
  "deviceResults": {
    "SetCurrentPosition": 0,
    "SetPan": 0,
    "SetVolume": 0,
    "Lock": {
      "HRESULT": 0,
      "firstRegion": "entire destination Bn payload",
      "firstLength": 262144,
      "secondRegion": null,
      "secondLength": 0
    },
    "Unlock": 0,
    "Play": 0,
    "Stop": 0
  },
  "positionDependency": {
    "recordVolume": 255,
    "rangeIndex": 0,
    "radius": 400,
    "innerRadius": 60,
    "position": [
      0,
      0,
      320
    ],
    "rawSquaredXY": 0,
    "contextFactor": 0,
    "clientMaster": 100,
    "clientMusic": 100,
    "gains": [
      100,
      100
    ],
    "class": 1,
    "threeDInterface": null,
    "propertyInterface": null,
    "applicationEnabled": true,
    "verifiedSubcalculations": [
      "S3D-R02",
      "S3D-R05",
      "S3D-R06"
    ],
    "scope": "Scalar/radius subcalculations and declared finite inputs only; no full acoustic-update or cache-guard derivation is claimed."
  },
  "bufferedBookmarkQuery": {
    "whenStreamAbsent": "returns0 inside synchronization; no device-position query"
  },
  "allocationProvider": {
    "listNodes": "each requested node allocation succeeds with a fresh distinct identity; each free returns normally",
    "assetStorage": "each requested raw-storage allocation succeeds; request size is source-file size plus800, while accounting uses only source-file size"
  },
  "fileProvider": {
    "ordinaryOperations": [],
    "whenExplicitlyInvoked": "Open sizing and read handles as distinct valid handles; size equals declared synthetic file length; full read returns success and exactly requested bytes; closes succeed; no undeclared file effects"
  },
  "floatingPoint": {
    "precision": 64,
    "rounding": "nearest ties-to-even",
    "exceptions": "masked",
    "clientWorkaround": false,
    "backendWorkaround": false
  },
  "suppliedFiniteAcousticUpdateProfile": {
    "name": "SUPPLIED-FINITE-ACOUSTIC-UPDATE",
    "status": "Explicit supplied finite profile; unresolved S3D-R07/Q02 integration. Owned by acoustic-update integration, not an STP cache rule.",
    "returnsNormally": true,
    "preserveUnspecifiedState": true,
    "perCallSamples": {
      "routingMode": 0,
      "masterVolume": 100,
      "musicClassification": "none for cues100/101"
    },
    "readBoundary": "Every reached acoustic-update call samples these supplied values; routing sampling precedes the base-cache comparison, including cache hits. No relative order beyond this declared boundary is invented.",
    "finiteBaseResults": [
      {
        "currentVolume": 255,
        "gains": [
          100,
          100
        ],
        "recordVolume": 255,
        "rangeIndex": 0,
        "factor": 0,
        "base": 255
      },
      {
        "currentVolume": 0,
        "gains": [
          100,
          100
        ],
        "recordVolume": 255,
        "rangeIndex": 0,
        "factor": 0,
        "base": 0
      }
    ],
    "cacheHit": {
      "condition": "cached XYZ equals supplied XYZ, cached base equals profile base, and cached factor equals supplied factor",
      "effects": "No cache or device writes; per-call sampling still occurs; return normally."
    },
    "cacheMissOrderedEffects": [
      "set propertyScalar0",
      "set requested amplitude to the profile base",
      "call SetVolume with0 for base255 or−10000 for base0; supplied return0",
      "cache supplied XYZ, profile base and supplied factor"
    ],
    "uncoveredInputs": "GAP unless an additional profile is explicitly supplied; no extrapolation to other modes, option/duck gains, volumes, factors or floating-point inputs."
  }
}
```

##### Shared ordered buffered-start trace

The routingMode input and acousticConversionMode input are distinct retained states, both explicitly supplied as 0. Every reached routing getter read returns 0 because there are no writers. The supplied full acoustic-update profile includes a routingMode read for crossmix before the base-cache comparison, even when that supplied profile later cache-matches and makes no device call. Stereo1 avoids routing reads for admission class selection only; it does not remove this property-stage read. STP-V28–STP-V35, STP-V44 and the other named playing-property visits consume this same supplied state whenever that stage is reached. No getter-call count is inferred from a device-call suppression.

The following explicit sequence is reused only when a vector invokes it; numbered substitutions change the supplied lock/play/loop results without reordering the other steps. Full acoustic-update effects in this trace are conditional on SUPPLIED-FINITE-ACOUSTIC-UPDATE and unresolved S3D-R07/Q02 integration; verified scalar/radius pieces consume S3D-R02/R05/R06.

1. Prepare Wn under synchronization: Bn.SetCurrentPosition(0)→0; Bn.SetPan(0)→0; propertyScalar0 with no property interface; requestedVolume255 and Bn.SetVolume(0)→0; leave synchronization.
2. Store decoded Loop. Enter load synchronization; parse the complete supplied synthetic WAV, validate loopStart0 against PCM length 4, and store its PCM data reference/dataLength4. Then enter nested fill synchronization; Lock(0,262144,flags 0) returns the declared regions. For successful Lock and Loop0 copy the 4 PCM bytes and zero the remaining 262140 bytes; writtenBytes4/endDataFlag1. For Loop1 repeat the 4 PCM bytes 65536 times; writtenBytes262144/endDataFlag0. Unlock declared regions→0; leave fill synchronization. Set halfFlag1 and endMarker4 for nonloop or−1 for loop; leave load synchronization. Load returns 1. The data reference and dataLength become the supplied PCM and 4.
3. Tentatively associate Sn with the incoming voice. Reset cached position/factor to 0 and cached base volume to−1. Apply propertyScalar0, requestedVolume255 and Bn.SetVolume(0)→0; cache position(0,0,320), base volume 255 and factor 0.
4. Enter buffered-start synchronization; Bn.Play(0,0,flags 1) returns the supplied result. A nonnegative result sets marker 1; a negative result retains its prior marker and returns failure. Leave synchronization.
5. On success the caller stores the reverse voice-slot association and playing state. The immediately following exact cache match produces no device operations. A playing coordinator visit refreshes the record access tick and removes its old LRU node before allocating/appending its new tail node.

An absent-stream stop trace, where invoked by these vectors, is ordered: synchronized bookmark query returns 0 for absent stream and stores bookmark 0; synchronized Bn.Stop→0 and marker 0; then voice stopped, reverse slot absent and slot association absent. Temporary cleanup removes the active node and matching owner handles/references before the stop helper; it clears inUse after that helper and eventually frees temporary deferred nodes.

Request-only limitation: STP-V37/STP-V53/STP-V54 explicitly reset the voice pool and active/owner lists but retain the shared S0 reference to pool index 0. The request boundary does not inspect it. If index 0 is reused as N0, S0 still references that same pool entry while N0 has no reverse slot. These vectors preserve that supplied nonreciprocal state; they do not silently reset S0 or assert normal lifecycle reachability.

Asset-only cases retain any unrelated row fields unless explicitly changed, even where an unloaded record retains a prior raw reference or stale size. No invariant repair is inferred.


### 8.1.2 Stream and worker fixture

#### Shared interpretation

Use unsigned 32-bit arithmetic for byte counters, sizes, offsets, masking, reference counts, and deadlines unless a case explicitly says signed. low32 discards high bits. Device result signs are interpreted as signed 32-bit values. A method call and its supplied return do not prove audible or physical device success. Device buffer bytes, source bytes, pointer identities, handles, and object liveness are distinct observations.

A main vector and each variation start from their declared fixture independently. Sequential checkpoints within one vector carry every object, list, byte, handle, provider-consumption position, and global value forward. Never reset a completed producer to a common fixture before a consumer. Unless an output changes a named field, it retains its exact input value. Freed objects have no remaining readable-field oracle; retain only the ordered trace and liveness result. Device lifetime after Release remains a provider outcome.

All unspecified semantic fields are zero or null; named provider identities, valid list membership, and allocator bookkeeping remain valid. Lists contain exactly the stated identities in semantic traversal order, with no unrelated entries. Memory allocation and free are explicit successful providers where named; no internal allocator state, node encoding, or reuse is asserted. All critical-section operations return normally and are reentrant. No interleaving occurs except where explicitly stated. All methods not listed in a case are absent. No case below calls a game random generator: Seeds/nodraw = no seed input is read or changed; orderedDraws = []. Clock inputs are not random draws.

#### Named semantic fixtures

##### STREAM-CREATION

Source owner F is live with one external reference; direct OS handle H is valid; archive is null; stream flag 0; logical position 777; read-check enable fields initially 99 and 88, other read-check status 7; optional owned read buffers null. Source is a 54-byte RIFF/WAVE file: RIFF size 46, fmt length 16, PCM tag 1, channels 2, sample rate 22050, average rate 88200, block alignment 4, bits/sample 16; data begins 44, length 10, bytes[1,2,3,4,5,6,7,8,9,10]. No extra chunks. Reads return BOOL 1 and exactly the requested bytes; size queries 54; seeks return the exact tracked position; close returns BOOL 1. The owner has valid list membership.

Global device D is present; direct-source permission 1; configured chunk 16384; active chunk 16384; scratch absent; source alignment mode 0; existing worker T; wake event E; worker exit 0; bandwidth 0; pump inhibition 0; queues, stream list, completion notifications empty; current request null; read-error policy 0. Allocating scratch and a zero-initialized stream succeeds. Generic free returns normally. Worker T does not run during creation.

Device creation is absent for the borrowed-buffer cases. SetCurrentPosition, Lock, Unlock, Play, Stop return 0 unless overridden. A successful Lock supplies its specified contiguous first region and length, no second region. Wake returns BOOL 1 unless overridden. Device interfaces are valid. No AddRef is supplied or synthesized.

##### STREAM-WRAPPER

W has inner buffer B; capacity 262144; half capacity 131072; no secondary/property/3D interface; flags 1; stored/requested volume 255; property scalar 0; marker 0; stream null; data pointer null; data length 0; played 0; written 0; halfFlag 0; endData 0; endMarker −1; loop 0; loopStart 0; backend fade inactive with from/target/start/end all 0. B contains 0xaa in every byte. All other semantic fields zero/null. This wrapper survives stream-node destruction.

A created stream S has semantic fields file, write cursor, average rate, loop, state, silence tail, data start/length, start offset, cached position, progress, raw volume/pan inputs, buffer, capacity, ownsBuffer, silence byte, and references. Requests Q have completion event, file, source offset, cache offset, deadline, priority, raw read buffer, cancellation identity, device destination, destination offset, stream, requested length, freeAfter, actual read count, and completion callback. Their unspecified semantic fields are zero/null. Scratch allocation bytes are opaque provider state; exact newly read prefixes are stated, and untouched scratch bytes retain their supplied allocation values.

##### BUFFERED-WORKER

W has inner B; capacity 262144; half capacity 131072; data pointer null; data length 4; marker 1; played 0; written 4; halfFlag 1; endData 1; endMarker −1; loop 0; loopStart 0; stream null; flags 0; stored volume 0; property scalar 0; backend fade inactive and remaining fade fields 0. Other semantic fields zero/null. B=[1,2,3,4] followed by 262140 zero bytes.

Sound worker list is the singleton W; event E; scheduled playlist disabled; cached iconic state 0; supplied window Hwin with IsIconic(Hwin)=0. Wait timeout is 50 ms. Default Wait results[258,258,0] mean two iterations then exit. GetStatus results are two successes with status 0. GetPosition results are success(play 4,write 0), success(play 5,write 0). Stop returns 0. No other external operation occurs. Each iteration orders: Wait timeout, enter Sound lock, window getter, IsIconic, nested status check, unchanged context-volume branch, inactive fade branch, then buffered position/completion, leave lock. Device and source bytes are explicit independent buffers.

##### REQUEST-RETENTION

After STREAM-CREATION plus W start with bookmark 0x40000001 and Loop 1, then the initial pump has selected Q1 for a read: wrapper W marker 1/stream F; stream S state pending-start, start/cached/progress 4, tail 0, write 16384, capacity 262144, data 44/10, average 88200, ownsBuffer 0, buffer B; file F logical position 54/references 3/stream flag 1; stream S references 2. Current request Q1 is unlinked from the request queue; queue empty; Q1 file F, stream S, source 48, destination buffer B at 0, length 6, freeAfter 1, actual 0, no completion event or callback; no other request. B still all 0xaa. Reference owners are external F owner, S, Q1 for F; list owner and Q1 for S. This is a supplied selected-read checkpoint, not an additional thread scheduling claim.


### 8.1.3 Lifecycle executable cases

The stream fixture identities W and B alias logical W1 and B1 when STP-V45 hands its complete output to STP-V47; all producer-only state carries forward. Other standalone or independent cases restart as stated. The normal-stop case STP-V66 declares its own checkpoint. No shared-fixture reset occurs during a declared sequential continuation.

#### STP-V28 — Replace an older duplicate after the incoming Play

##### Inputs

```json
{
  "entry": "audio coordinator",
  "overrides": {},
  "suppliedAcousticUpdateProfile": "SUPPLIED-FINITE-ACOUSTIC-UPDATE",
  "acousticIntegrationGate": "S3D-R07/Q02 unresolved; scalar/radius subcalculations alone use verified R02/R05/R06"
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
[
  "Periodic random-cue producer disabled; order already I then O; retirement prepass has no effect.",
  "I sole owner A selects playing family-matching O; set O retirement=true and allocate/prepend deferred O.",
  "Request selection bypassed; marked matching cue100 absent; commit history100=(recent100,older7).",
  "Select free compatible S1; run the complete shared buffered-start trace with Loop0 and successful Lock/Play.",
  "Install I playing/S1; repeat properties cache-match. Refresh100 access100, free its old LRU node, allocate/append new100 node.",
  "Visit O as still playing; cache-match. Refresh101 access100, free its old LRU node, allocate/append new101 node.",
  "Deferred cleanup frees active O node, removes A handle500 and clears O references; run shared stop trace on W0, then clear O.inUse and free deferred O node.",
  "Gains remain100; lastScalarTick100."
]
```

##### exactOutputs

Caller logical-state/order assertions use the declared normal dependency returns. Acoustic-update getter/base/property/cache assertions, including its SetVolume calls, are conditional on the supplied finite profile and unresolved S3D-R07/Q02 integration.

```json
{
  "draws": [],
  "seedA": [
    1,
    0
  ],
  "sortOrder": [
    "I",
    "O"
  ],
  "duplicate": "O selected; set O.retirement=1 and prepend deferred O before I admission",
  "history100": {
    "recent": 100,
    "older": 7
  },
  "I": {
    "inUse": 1,
    "cue": 100,
    "handle": 501,
    "slot": "S1",
    "state": "playing",
    "currentVolume": 255,
    "activeTransition": 0
  },
  "O": {
    "inUse": 0,
    "retirement": 1,
    "state": "stopped",
    "slot": null,
    "references": [],
    "bookmark": 0
  },
  "ownerAHandles": [
    501
  ],
  "activeVoices": [
    "I"
  ],
  "S0Voice": null,
  "S1Voice": "I",
  "W0Marker": 0,
  "W1": {
    "marker": 1,
    "PCM": "1,2,3,4 then262140 zeros",
    "data": "WAV100.PCM",
    "dataLength": 4,
    "writtenBytes": 4,
    "endDataFlag": 1,
    "endMarker": 4,
    "halfFlag": 1
  },
  "LRU": [
    100,
    101
  ],
  "accessTicks": {
    "100": 100,
    "101": 100
  },
  "usedBytes": 96,
  "pendingCount": 0,
  "lastScalarTick": 100,
  "allUnlistedState": "exactly unchanged from fully specified fixture",
  "S1Cache": {
    "XYZ": [
      0,
      0,
      320
    ],
    "baseVolume": 255,
    "factor": 0
  }
}
```

##### Ordered output assertion

The exact projection of the main path's device log onto playback operations is [B1.Play(repeating) returning 0, B0.Stop returning 0], in that order. Other setup calls retain the complete shared start trace. This projected output excludes a Stop of B0 before B1 Play even when final logical/slot fields agree.

##### Independent null tracked-reference variation

Reset the complete STP-V28 input, then set trackingEnabled=true, record 100.Tracking=1 and I.references=[null], represented by one valid node whose value is null. No normal reachability is claimed for that supplied reference value. Other inputs remain unchanged, including the disabled periodic producer and nonretiring/no-transition prefix. Sort order[I,O]; initial master getter 100 once; cue 100 causes no music getter. I terminates at the null-reference guard before context/position calculation, STP-R02, grouping or admission. Exact pre-termination audio state equals this sorted changed input. Master calls 1/music 0; no RNG/acoustic-profile/cache/LRU/device/file/audio-owned allocation or free. Diagnostic-foundation effects are outside that audio log; no normal return or safe fallback is asserted. Draw log[], runtime_observed:no.

##### Discriminates

Old Stop is after new Play, and old/incoming list mutations follow deferred cleanup

runtime_observed: no


#### STP-V29 — Defer removes the incoming request

##### Inputs

```json
{
  "entry": "audio coordinator",
  "overrides": {
    "records.100.DeferInst": 1,
    "records.100.StopInst": 0,
    "records.101.DeferInst": 1,
    "records.101.StopInst": 0
  },
  "independentVariation": {
    "name": "selection bypass disabled but duplicate deferral still comes first",
    "overrides": {
      "voicePool.entries.I.requestFlags": 0
    }
  },
  "suppliedAcousticUpdateProfile": "SUPPLIED-FINITE-ACOUSTIC-UPDATE",
  "acousticIntegrationGate": "S3D-R07/Q02 unresolved; scalar/radius subcalculations alone use verified R02/R05/R06"
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
{
  "base": [
    "Order I,O; no retirement-prepass effects.",
    "Duplicate lookup selects O; DeferInst1/StopInst0 marks I retirement=true and prepends deferred I. Skip admission, random selection and history.",
    "O properties cache-match; refresh101 access100 and replace its LRU node at the tail.",
    "Remove active I, A handle501 and I reference A. Pending I has no backend stop/query; clear I.inUse and free deferred node.",
    "Gains stay100; lastScalarTick100."
  ],
  "independentFlags0Variation": [
    "Restart complete STP-V29 entering input, changing only I.requestFlags1→0.",
    "Same duplicate lookup selects O. DeferInst1/StopInst0 sets incoming retirement and enqueues I before admission; continue to O without invoking group selection.",
    "Same O cache/access/LRU update and deferred incoming cleanup as base. I cue100 and requestFlags0 remain retained after inUse clears; seed remains(1,0), ordered draws[]."
  ]
}
```

##### exactOutputs

Caller logical-state/order assertions use the declared normal dependency returns. Acoustic-update getter/base/property/cache assertions, including its SetVolume calls, are conditional on the supplied finite profile and unresolved S3D-R07/Q02 integration.

```json
{
  "draws": [],
  "seedA": [
    1,
    0
  ],
  "history100": {
    "recent": 7,
    "older": 6
  },
  "I": {
    "inUse": 0,
    "state": "pending",
    "slot": null,
    "retirement": 1,
    "references": []
  },
  "O": {
    "inUse": 1,
    "state": "playing",
    "slot": "S0",
    "retirement": 0
  },
  "ownerAHandles": [
    500
  ],
  "activeVoices": [
    "O"
  ],
  "deviceOperations": [],
  "LRU": [
    100,
    101
  ],
  "accessTicks": {
    "100": 90,
    "101": 100
  },
  "allUnlistedState": "exactly unchanged",
  "independentFlags0Variation": {
    "draws": [],
    "seedA": [
      1,
      0
    ],
    "history100": {
      "recent": 7,
      "older": 6
    },
    "I": {
      "inUse": 0,
      "state": "pending",
      "slot": null,
      "retirement": 1,
      "references": [],
      "requestFlags": 0,
      "cue": 100
    },
    "O": {
      "inUse": 1,
      "state": "playing",
      "slot": "S0",
      "retirement": 0
    },
    "ownerAHandles": [
      500
    ],
    "activeVoices": [
      "O"
    ],
    "deviceOperations": [],
    "LRU": [
      100,
      101
    ],
    "accessTicks": {
      "100": 90,
      "101": 100
    },
    "allUnlistedState": "exactly unchanged",
    "groupSelectionCalls": 0
  }
}
```

##### Discriminates

Defer removes incoming rather than retaining it pending; no group draw/history commit. With requestFlags0, premature group selection would be observable: actual deferral still skips selection entirely, keeps cue 100/history 7/6 and leaves the single seed(1,0) with no draws

runtime_observed: no


#### STP-V30 — Stopped loop resolves its endpoint before a later admission

##### Inputs

```json
{
  "entry": "two consecutive audio coordinator invocations",
  "overrides": {
    "voicePool.activeOrder": [
      "I"
    ],
    "voicePool.entries.O": "free entry with zero/absent semantic fields",
    "owners.A.handles": [
      501
    ],
    "slots.S0": "nullvoice,marker0,class0 with valid mono wrapper",
    "voicePool.entries.I.state": "stopped",
    "voicePool.entries.I.activeTransition": 1,
    "voicePool.entries.I.currentVolume": 17,
    "voicePool.entries.I.startVolume": 0,
    "voicePool.entries.I.endVolume": 255,
    "voicePool.entries.I.startTick": 50,
    "voicePool.entries.I.endTick": 60,
    "records.100.Loop": 1,
    "records.100.StopInst": 0
  },
  "clocks": [
    {
      "tickArgument": 100,
      "audioCounter": 100
    },
    {
      "tickArgument": 101,
      "audioCounter": 101
    }
  ],
  "suppliedAcousticUpdateProfile": "SUPPLIED-FINITE-ACOUSTIC-UPDATE",
  "acousticIntegrationGate": "S3D-R07/Q02 unresolved; scalar/radius subcalculations alone use verified R02/R05/R06"
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
[
  {
    "tick": 100,
    "actions": [
      "Stopped I has no slot, Loop1, Duration0, no retirement. Copy endpoint255 to currentVolume, clear transition active and set pending. Continue without admission/property/access refresh.",
      "No device/list activity; gains100; lastScalarTick100."
    ]
  },
  {
    "tick": 101,
    "actions": [
      "Pending I now qualifies; no duplicate; bypass random selection; commit cue history100/7.",
      "S0 incompatible; choose S1. Run shared buffered-start trace with Loop1: repeat4 PCM bytes65536 times; written262144, endDataFlag0, endMarker−1.",
      "Install playing I/S1; cache-match; refresh100 access101 and replace its LRU node at tail; gains100,lastScalarTick101."
    ]
  }
]
```

##### exactOutputs

Caller logical-state/order assertions use the declared normal dependency returns. Acoustic-update getter/base/property/cache assertions, including its SetVolume calls, are conditional on the supplied finite profile and unresolved S3D-R07/Q02 integration.

```json
[
  {
    "tick": 100,
    "outputs": {
      "I": {
        "inUse": 1,
        "state": "pending",
        "slot": null,
        "currentVolume": 255,
        "activeTransition": 0
      },
      "history100": {
        "recent": 7,
        "older": 6
      },
      "deviceOperations": [],
      "record100Access": 90,
      "allUnlistedState": "unchanged"
    }
  },
  {
    "tick": 101,
    "outputs": {
      "I": {
        "inUse": 1,
        "state": "playing",
        "slot": "S1",
        "currentVolume": 255,
        "activeTransition": 0
      },
      "history100": {
        "recent": 100,
        "older": 7
      },
      "W1": {
        "marker": 1,
        "logicalLoop": 1,
        "PCM": "bytes1,2,3,4 repeated65536 times",
        "writtenBytes": 262144,
        "endDataFlag": 0,
        "endMarker": -1,
        "data": "WAV100.PCM",
        "dataLength": 4,
        "halfFlag": 1
      },
      "record100Access": 101,
      "LRU": [
        101,
        100
      ],
      "lastScalarTick": 101,
      "allUnlistedState": "unchanged except normal property cache assignments",
      "S1Cache": {
        "XYZ": [
          0,
          0,
          320
        ],
        "baseVolume": 255,
        "factor": 0
      }
    }
  }
]
```

##### Discriminates

Stopped loop resolves fade then postpones admission until another invocation

runtime_observed: no


#### STP-V31 — Duration equality, later stop, then later disposal

##### Inputs

```json
{
  "entry": "three sequential audio coordinator invocations",
  "overrides": {
    "voicePool.activeOrder": [
      "O"
    ],
    "voicePool.entries.I": "free entry with zero/absent semantic fields",
    "owners.A.handles": [
      500
    ],
    "voicePool.entries.O.cue": 100,
    "voicePool.entries.O.start": 100,
    "records.100.Loop": 1,
    "records.100.Duration": 10,
    "records.100.StopInst": 0,
    "wrappers.W0.logicalLoop": 1,
    "wrappers.W0.endDataFlag": 0
  },
  "clocks": [
    {
      "tickArgument": 110,
      "audioCounter": 110
    },
    {
      "tickArgument": 111,
      "audioCounter": 111
    },
    {
      "tickArgument": 112,
      "audioCounter": 112
    }
  ],
  "suppliedAcousticUpdateProfile": "SUPPLIED-FINITE-ACOUSTIC-UPDATE",
  "acousticIntegrationGate": "S3D-R07/Q02 unresolved; scalar/radius subcalculations alone use verified R02/R05/R06"
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
[
  {
    "tick": 110,
    "actions": [
      "Elapsed10 equals Duration10: keep playing. Property cache-match, refresh100 access110 and replace its LRU node at tail. No device operations; lastScalarTick110."
    ]
  },
  {
    "tick": 111,
    "actions": [
      "Elapsed11 exceeds10. Shared stop trace returns bookmark0, stops B0, clears marker and both voice-slot associations; O becomes stopped.",
      "Do not queue cleanup in this invocation: the initial stopped-state branch has already passed. No access refresh; lastScalarTick111."
    ]
  },
  {
    "tick": 112,
    "actions": [
      "Stopped Loop1 with nonzero Duration does not restart; prepend deferred O.",
      "Remove active O/A handle500/O reference A. Stop helper sees stopped/no slot and performs no backend work; clear inUse, free deferred node. lastScalarTick112."
    ]
  }
]
```

##### exactOutputs

Caller logical-state/order assertions use the declared normal dependency returns. Acoustic-update getter/base/property/cache assertions, including its SetVolume calls, are conditional on the supplied finite profile and unresolved S3D-R07/Q02 integration.

```json
[
  {
    "tick": 110,
    "outputs": {
      "O": {
        "inUse": 1,
        "state": "playing",
        "slot": "S0"
      },
      "W0Marker": 1,
      "record100Access": 110,
      "allUnlistedState": "unchanged except LRU access and lastScalarTick110",
      "LRU": [
        101,
        100
      ],
      "lastScalarTick": 110
    }
  },
  {
    "tick": 111,
    "outputs": {
      "O": {
        "inUse": 1,
        "state": "stopped",
        "slot": null,
        "bookmark": 0
      },
      "S0Voice": null,
      "W0Marker": 0,
      "ownerAHandles": [
        500
      ],
      "activeVoices": [
        "O"
      ],
      "allUnlistedState": "unchanged except lastScalarTick111",
      "LRU": [
        101,
        100
      ],
      "record100Access": 110,
      "lastScalarTick": 111
    }
  },
  {
    "tick": 112,
    "outputs": {
      "O": {
        "inUse": 0,
        "state": "stopped",
        "slot": null,
        "references": []
      },
      "ownerAHandles": [],
      "activeVoices": [],
      "allUnlistedState": "unchanged except lastScalarTick112",
      "LRU": [
        101,
        100
      ],
      "record100Access": 110,
      "lastScalarTick": 112
    }
  }
]
```

##### Discriminates

Strict duration boundary; duration prevents automatic loop restart; cleanup is later

runtime_observed: no


#### STP-V32 — Fade-to-zero completes after the retirement prepass

##### Inputs

```json
{
  "entry": "two sequential audio coordinator invocations",
  "overrides": {
    "clock.tickArgument": 111,
    "clock.audioCounter": 111,
    "controls.lastScalarTick": 111,
    "voicePool.activeOrder": [
      "O"
    ],
    "voicePool.entries.I": "free entry with zero/absent semantic fields",
    "owners.A.handles": [
      500
    ],
    "voicePool.entries.O.cue": 100,
    "voicePool.entries.O.currentVolume": 11,
    "voicePool.entries.O.retirement": 1,
    "voicePool.entries.O.activeTransition": 1,
    "voicePool.entries.O.startVolume": 110,
    "voicePool.entries.O.endVolume": 0,
    "voicePool.entries.O.startTick": 100,
    "voicePool.entries.O.endTick": 110,
    "slots.S0.cachedBaseVolume": 11,
    "records.100.StopInst": 0,
    "slots.S1": "all zero cache,nullvoice,normalclass1,W1 marker0"
  },
  "clocks": [
    {
      "tickArgument": 111,
      "audioCounter": 111
    },
    {
      "tickArgument": 112,
      "audioCounter": 112
    }
  ],
  "suppliedAcousticUpdateProfile": "SUPPLIED-FINITE-ACOUSTIC-UPDATE",
  "acousticIntegrationGate": "S3D-R07/Q02 unresolved; scalar/radius subcalculations alone use verified R02/R05/R06"
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
[
  {
    "tick": 111,
    "actions": [
      "Retirement prepass skips active transition whose target is0.",
      "Inline transition sees111>110; store currentVolume0 and clear active. Keep playing.",
      "Consume the supplied finite acoustic-update profile: propertyScalar0, requestedVolume0, B0.SetVolume(−10000)→0; cache baseVolume0.",
      "Refresh100 access111/LRU tail; no Stop/deferred cleanup; marker remains1."
    ]
  },
  {
    "tick": 112,
    "actions": [
      "Retirement prepass now sees inactive transition and runs shared stop trace, then clears associations and sets stopped.",
      "Stopped retired O is deferred; cleanup removes active/owner/reference nodes, clears inUse and frees deferred node. Stop helper makes no second backend call.",
      "Gains100; lastScalarTick112."
    ]
  }
]
```

##### exactOutputs

Caller logical-state/order assertions use the declared normal dependency returns. Acoustic-update getter/base/property/cache assertions, including its SetVolume calls, are conditional on the supplied finite profile and unresolved S3D-R07/Q02 integration.

```json
[
  {
    "tick": 111,
    "outputs": {
      "O": {
        "inUse": 1,
        "state": "playing",
        "slot": "S0",
        "retirement": 1,
        "currentVolume": 0,
        "activeTransition": 0,
        "startVolume": 110,
        "endVolume": 0,
        "startTick": 100,
        "endTick": 110
      },
      "W0Marker": 1,
      "S0Voice": "O",
      "S0CachedBaseVolume": 0,
      "ownerAHandles": [
        500
      ],
      "activeVoices": [
        "O"
      ],
      "record100Access": 111,
      "lastScalarTick": 111,
      "allUnlistedState": "unchanged except LRU row100 moved to tail",
      "W0RequestedVolume": 0,
      "LRU": [
        101,
        100
      ]
    }
  },
  {
    "tick": 112,
    "outputs": {
      "O": {
        "inUse": 0,
        "state": "stopped",
        "slot": null,
        "retirement": 1,
        "currentVolume": 0,
        "activeTransition": 0,
        "references": [],
        "bookmark": 0
      },
      "S0Voice": null,
      "W0Marker": 0,
      "ownerAHandles": [],
      "activeVoices": [],
      "lastScalarTick": 112,
      "allUnlistedState": "unchanged from first-step output"
    }
  }
]
```

##### Discriminates

Fade completion occurs after retirement prepass; actual stop/removal is a later invocation

runtime_observed: no


#### STP-V33 — Failed fade-in retains a zero-volume pending loop

##### Inputs

```json
{
  "entry": "two sequential audio coordinator invocations",
  "overrides": {
    "voicePool.activeOrder": [
      "I"
    ],
    "voicePool.entries.O": "free entry with zero/absent semantic fields",
    "owners.A.handles": [
      501
    ],
    "records.100.Loop": 1,
    "records.100.FadeIn": 5,
    "records.100.StopInst": 0,
    "slots": "All16 slots class0/nullvoice, valid mono wrappers(flags0,capacity131072,half65536,marker0), so no compatible class1 slot or victim"
  },
  "clocks": [
    {
      "tickArgument": 100,
      "audioCounter": 100
    },
    {
      "tickArgument": 101,
      "audioCounter": 101
    }
  ]
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
[
  {
    "tick": 100,
    "actions": [
      "Pending due I has current255 and no duplicate. Capture255 locally, then write currentVolume0 before admission.",
      "Bypass group selection; commit history100/7. Every slot has class0 while Stereo1 requests class1, so no free slot or victim; return no slot.",
      "Do not restore volume or schedule a transition. Loop1 retains pending I; no properties/device/access update."
    ]
  },
  {
    "tick": 101,
    "actions": [
      "CurrentVolume0 skips admission and its failure-removal branch.",
      "No group/history/slot/device/list activity; only lastScalarTick advances101; gains100."
    ]
  }
]
```

##### exactOutputs

```json
[
  {
    "tick": 100,
    "outputs": {
      "I": {
        "inUse": 1,
        "state": "pending",
        "slot": null,
        "currentVolume": 0,
        "activeTransition": 0
      },
      "history100": {
        "recent": 100,
        "older": 7
      },
      "deviceOperations": [],
      "record100Access": 90,
      "allUnlistedState": "unchanged except lastScalarTick100"
    }
  },
  {
    "tick": 101,
    "outputs": {
      "I": {
        "inUse": 1,
        "state": "pending",
        "currentVolume": 0,
        "activeTransition": 0,
        "slot": null
      },
      "history100": {
        "recent": 100,
        "older": 7
      },
      "deviceOperations": [],
      "allUnlistedState": "unchanged except lastScalarTick101"
    }
  }
]
```

##### Discriminates

Failed automatic fade-in can retain zero-volume pending loop; no implicit restore/retry

runtime_observed: no


#### STP-V34 — Pending async asset finalizes before a later successful admission

##### Inputs

```json
{
  "entry": "coordinator100; pending-asset scan125; coordinator125",
  "overrides": {
    "voicePool.activeOrder": [
      "I"
    ],
    "voicePool.entries.O": "free entry with zero/absent semantic fields",
    "owners.A.handles": [
      501
    ],
    "slots.S0": "nullvoice,marker0,class0 valid mono wrapper",
    "records.100.StopInst": 0,
    "records.100.AsyncOnly": 1,
    "records.100.state": "pending",
    "records.100.raw": null,
    "records.100.PCM": null,
    "records.100.PCMSize": 0,
    "records.100.descriptor": "D100",
    "assets.pendingCount": 1,
    "assets.lastPendingScan": 100,
    "asyncDescriptorD100": {
      "valid": true,
      "ready": true,
      "ownsPayload": true,
      "data": "WAV100",
      "size": 48,
      "file": null,
      "eventNode": null
    }
  },
  "clocks": [
    {
      "tickArgument": 100,
      "audioCounter": 100
    },
    {
      "tickArgument": 125,
      "audioCounter": 125
    }
  ],
  "provider": "D100 is already ready when polled; no OS wait or inferred read completion",
  "suppliedAcousticUpdateProfile": "SUPPLIED-FINITE-ACOUSTIC-UPDATE",
  "acousticIntegrationGate": "S3D-R07/Q02 unresolved; scalar/radius subcalculations alone use verified R02/R05/R06"
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
[
  {
    "stage": "coordinator100",
    "actions": [
      "Bypass group selection; commit history100/7. Asset admission with synchronous=false returns while row pending; no descriptor wait.",
      "Admission fails while asset not loaded; nonloop request survives specifically because AsyncOnly1 and asset pending."
    ]
  },
  {
    "stage": "pending scan125",
    "actions": [
      "Elapsed25 permits full scan of valid records1–2933; only100 pending. Poll supplied ready D100→ready without OS wait.",
      "Take supplied WAV100 data/size; mark loaded; validate PCM16/22050/stereo2. pendingCount1→0; descriptor remainsD100; lastPendingScan125."
    ]
  },
  {
    "stage": "coordinator125",
    "actions": [
      "Commit history again: recent100/older100. Loaded asset bypasses loader.",
      "Choose S1; run complete shared buffered-start trace Loop0/Lock0/Play0. Install playing I/S1; cache-match.",
      "Refresh100 access125/LRU tail; lastScalarTick125,gains100."
    ]
  }
]
```

##### exactOutputs

Caller logical-state/order assertions use the declared normal dependency returns. Acoustic-update getter/base/property/cache assertions, including its SetVolume calls, are conditional on the supplied finite profile and unresolved S3D-R07/Q02 integration.

```json
[
  {
    "stage": "coordinator100",
    "outputs": {
      "I": {
        "inUse": 1,
        "state": "pending",
        "slot": null,
        "currentVolume": 255,
        "activeTransition": 0
      },
      "history100": {
        "recent": 100,
        "older": 7
      },
      "pendingCount": 1,
      "deviceOperations": [],
      "allUnlistedState": "unchanged"
    }
  },
  {
    "stage": "pending scan125",
    "outputs": {
      "record100": {
        "raw": "WAV100",
        "rawSize": 48,
        "PCM": "WAV100.PCM",
        "PCMSize": 4,
        "Stereo": 1,
        "state": "loaded",
        "failure": 0,
        "descriptor": "D100"
      },
      "pendingCount": 0,
      "lastPendingScan": 125,
      "allUnlistedState": "unchanged"
    }
  },
  {
    "stage": "coordinator125",
    "outputs": {
      "I": {
        "inUse": 1,
        "state": "playing",
        "slot": "S1",
        "currentVolume": 255
      },
      "history100": {
        "recent": 100,
        "older": 100
      },
      "W1Marker": 1,
      "pendingCount": 0,
      "record100Access": 125,
      "LRU": [
        101,
        100
      ],
      "allUnlistedState": "unchanged except normalcache/lastScalarTick125",
      "W1": {
        "marker": 1,
        "PCM": "1,2,3,4 then262140 zeros",
        "data": "WAV100.PCM",
        "dataLength": 4,
        "writtenBytes": 4,
        "endDataFlag": 1,
        "endMarker": 4,
        "halfFlag": 1
      },
      "S1Cache": {
        "XYZ": [
          0,
          0,
          320
        ],
        "baseVolume": 255,
        "factor": 0
      }
    }
  }
]
```

##### Discriminates

Pending async descriptor is retained/retried and finalizes before successful admission, without inferred read completion

runtime_observed: no


#### STP-V35 — Negative Lock is ignored by the outer buffered load

##### Inputs

```json
{
  "entry": "audio coordinator100",
  "overrides": {
    "voicePool.entries.O": "free entry with zero/absent semantic fields",
    "slots.S0": "nullvoice marker0,class0 (incompatible)",
    "owners.A.handles": [
      501
    ],
    "records.100.StopInst": 0,
    "deviceResults.Lock.HRESULT": -1,
    "voicePool.activeOrder": [
      "I"
    ]
  },
  "suppliedAcousticUpdateProfile": "SUPPLIED-FINITE-ACOUSTIC-UPDATE",
  "acousticIntegrationGate": "S3D-R07/Q02 unresolved; scalar/radius subcalculations alone use verified R02/R05/R06"
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
[
  "Only I active; commit history100/7; S0 incompatible; select S1.",
  "Run shared prepare sequence and store Loop0. Enter load; parse valid WAV100; enter nested fill; B1.Lock(0,262144,flags0)→−1.",
  "Leave fill without copy/Unlock/counter update. Outer load nevertheless stores halfFlag1/endMarker4, leaves synchronization and returns1. Parsed data reference/length are retained.",
  "Tentatively attach I, reset/apply/cache properties as in shared trace. Buffered Play(0,0,flags1)→0; set marker1.",
  "Install I playing/S1; cache-match; refresh100 access100/LRU tail. No worker interleaving."
]
```

##### exactOutputs

Caller logical-state/order assertions use the declared normal dependency returns. Acoustic-update getter/base/property/cache assertions, including its SetVolume calls, are conditional on the supplied finite profile and unresolved S3D-R07/Q02 integration.

```json
{
  "draws": [],
  "history100": {
    "recent": 100,
    "older": 7
  },
  "I": {
    "inUse": 1,
    "state": "playing",
    "slot": "S1",
    "currentVolume": 255
  },
  "W1": {
    "marker": 1,
    "data": "WAV100.PCM",
    "dataLength": 4,
    "writtenBytes": 0,
    "endDataFlag": 0,
    "halfFlag": 1,
    "endMarker": 4,
    "payload": "byte170 repeated262144; no PCM copy"
  },
  "activeVoices": [
    "I"
  ],
  "allUnlistedState": "exactly unchanged",
  "S1Cache": {
    "XYZ": [
      0,
      0,
      320
    ],
    "baseVolume": 255,
    "factor": 0
  },
  "LRU": [
    101,
    100
  ],
  "record100Access": 100,
  "usedBytes": 96,
  "pendingCount": 0,
  "lastScalarTick": 100
}
```

##### Discriminates

Negative fill Lock does not propagate through the outer buffered load wrapper; admission may return success and Play retained bytes

runtime_observed: no


Independent block-count variation of STP-V35: restart its complete input, then set record 100 Block1=1, Block2=-1 and Block3=0. The loader-derived count is 1 and loop-start input aliases Block1. Prepare clears W1.loopStart to 0; after assigning Loop, admission sets it to low32(1*2)=2 before Load. Parsing yields data length 4, so unsigned 2<4 passes. The supplied negative Lock returns before any loop-start-dependent copy. Exact outputs equal STP-V35 relative to this changed input except W1.loopStart=2; the changed record blocks/count remain as supplied. The extra loop-start setter makes no device call and adds no synchronization. Payload remains all0xaa, written 0/endData0, no Unlock, marker 1 and voice playing with S1, and the same conditional acoustic/profile, history/LRU/access/seed/draw outputs. This rejects treating the block count as Block2 or omitting the count 1 setter. runtime_observed:no.

#### STP-V36 — borrowed looping stream

Inputs: STREAM-CREATION and STREAM-WRAPPER. W start receives path for F, bookmark 0x40000001, Loop 1. Source is opened before the Sound lock, returning owner F. Byte offset=low32(bookmark*4)=4; any nonzero caller Loop becomes the creator loop bit. Creator receives capacity 262144, offset 4, raw volume/pan 2147483647, supplied buffer B. Pump inhibition 0 is explicit. Only the listed stages execute. Device positions are success 16384, success 32768, then HRESULT −1. Locks are full 16384 at destinations 0 and 16384. All copies/reads succeed as specified. Request selection, read completion, and destruction are serialized dependency stages with no unrelated requests and bandwidth 0. Every reached clock call in request construction, selection, and these worker segments returns 1000; no unbounded worker invocation or cadence between checkpoints is asserted.

Seeds/nodraw: no seed touched. orderedDraws: [].

Trace: create borrowed S; set device position 0; wake; W stores F/marker 1/endMarker −1. Initial pump queues Q1(source 48,destination 0,length 6,delay 0,deadline 1000), advances logical 54/write 16384, sets pending-start. Select Q1; read bytes 5..10; Lock 16384, copy 6/pad 16378 zero, Unlock; progress 16388/tail 16378; startup requested-source threshold 48−44+6=10 reaches min(10,65536,262144), Play repeating then playing. Destroy Q1, file F references 3→2 then stream S references 2→1, clear current. Loop EOF resets file logical position 44/progress 0/tail 0; position 16384 writes cache 10 but free ring 0, no request. Next pump position 32768 writes cache 10/free 16384; queues Q2(source 44,destination 16384,length 10,delay 2600,deadline 3600), file logical position 54/write 32768. Select/read 10, Lock 16384, copy 10/pad 16374 zero; progress 16384/tail 16374, no new Play; destroy refs back to file F references 2 / stream S references 1. Query with negative device result retains cached 10; stream lookup still succeeds; Sound bookmark=10>>2=2. Sound stop ends S, then device Stop, closes owner F, clears wrapper W stream/marker.

ExactOutputs: successful creation and W start; final wrapper W marker 0/stream null/inner buffer B retained/endMarker −1/loop 0; other wrapper W fields unchanged. S and F freed, current null, queues empty, scratch retained. B first 16384 bytes=[5,6,7,8,9,10]+16378 zeros; next 16384=[1..10]+16374 zeros; remaining 229376 bytes retain 0xaa. Play count 1, Stop count 1, AddRef 0, Release 0, CloseH 1. Query returns 2. Q1 callback state playing/progress 16388/tail 16378; Q2 callback state playing/progress 16384/tail 16374. Start cache 4; cache 10 after successful position queries.

Discriminates: bookmark high-bit wrap, requested source length versus destination, chunk-sized progress/padding, loop reset, relative delay based on capacity −2 chunks, ignored cached query failure, borrowed teardown.


runtime_observed: no.

STP-V36 independent variations: startup Play −1 still records playing and identical Q1 counters; no actual playback success claim. Q1 Lock −1, or success with first length 8192 while required chunk 16384, performs no copy/pad/Unlock/Play/progress/tail/state update; after Q destruction stream S state pending-start/progress 4/tail 0/cache 4/write 16384/file F references 2 / stream S references 1, B unchanged. One further pump ignores pending-start, with no query/requeue. These variations stop at that checkpoint; no hidden Sound worker runs. The mismatched returned region is an artificial supplied shape, not a platform claim.


#### STP-V37 — Fresh request and negative-compound reuse

##### Inputs

```json
{
  "entry": "logical request creation twice in sequence",
  "overrides": {
    "controls.selectedEmitter": null,
    "voicePool.activeOrder": [],
    "voicePool.entries": "All200 entries free, with zero/absent semantic fields",
    "owners.A.handles": [],
    "voicePool.nextHandleCounter": 501,
    "records.100.Priority": 250,
    "records.100.Compound": 0
  },
  "requests": [
    {
      "step": 1,
      "arguments": {
        "cue": 100,
        "emitter": null,
        "delay": 0,
        "requestFlags": 1,
        "playbackOffset": 0
      },
      "callerChanges": {}
    },
    {
      "step": 2,
      "arguments": {
        "cue": 100,
        "emitter": null,
        "delay": 99,
        "requestFlags": 3,
        "playbackOffset": 123
      },
      "callerChanges": {
        "records.100.Compound": -1
      }
    }
  ],
  "slotReferenceLimitation": "Retain S0 reference to pool index0 exactly as stated in shared fixture; request path does not inspect it."
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
[
  {
    "step": 1,
    "actions": [
      "Audio/cue/Volume checks pass; Compound0 bypasses reuse lookup.",
      "Allocate first free pool index0 as N0; zero its semantic fields, mark inUse and prepend active N0. Increment handle counter501→502 and store502.",
      "Initialize cue100,start100,requestFlags1,offset0,currentVolume255 and priority250. Null emitter supplies position(0,0,320), distanceKey0, contextFactor0.",
      "Null emitter equals selected null emitter: unsigned8 priority250+80→74. No owner attachment; return502."
    ]
  },
  {
    "step": 2,
    "actions": [
      "Caller sets Compound−1. Request cue100/null emitter/delay99/flags3/offset123 finds first same-family nonretiring N0, without an age test.",
      "Reuse skips allocation and all field initialization; keep original start100/flags1/offset0/position/handle.",
      "Null equality adds80 again:74→154. Return existing502; pool index1 remains free."
    ]
  }
]
```

##### exactOutputs

```json
[
  {
    "step": 1,
    "outputs": {
      "returnHandle": 502,
      "counter": 502,
      "allocatedPoolIndex": 0,
      "newVoice": {
        "inUse": 1,
        "cue": 100,
        "handle": 502,
        "distanceKey": 0,
        "currentVolume": 255,
        "contextFactor": 0,
        "playbackOffset": 0,
        "references": [],
        "slot": null,
        "state": "pending",
        "start": 100,
        "requestFlags": 1,
        "priority": 74,
        "retirement": 0,
        "bookmark": 0,
        "activeTransition": 0,
        "position": [
          0,
          0,
          320
        ]
      },
      "activeVoices": [
        "N0"
      ],
      "allOtherVoiceFields": "zero/false/null/empty according to the shared voice schema",
      "seedA": [
        1,
        0
      ],
      "deviceOperations": [],
      "draws": []
    }
  },
  {
    "step": 2,
    "outputs": {
      "returnHandle": 502,
      "counter": 502,
      "poolIndex1StillFree": true,
      "N0": {
        "priority": 154,
        "start": 100,
        "requestFlags": 1,
        "playbackOffset": 0,
        "position": [
          0,
          0,
          320
        ],
        "handle": 502
      },
      "allOtherState": "unchanged",
      "deviceOperations": [],
      "draws": []
    }
  }
]
```

##### Independent handle-counter-wrap variation

Reset the complete STP-V37 input, set nextHandleCounter=4294967295 and execute only its first request. The counter increments with low32 wrap to 0; allocation still initializes N0 at pool index 0 and stores handle 0. Exact outputs equal the first-request outputs relative to this changed input except returnHandle0/counter 0/newVoice.handle0. N0 is inUse/pending, active[N0], start 100/flags 1/offset 0/bookmark 0/currentVolume255/priority 74/position(0,0,320)/distanceKey0/factor 0/no transition and refs[]. Owner A handles[] and seed(1,0) remain; all other entries stay free. The supplied S0 pool-index reference is retained as already stated for this request-only boundary. No second reuse or coordinator runs. No draw/device/file operation occurs. This distinguishes allocated handle 0 from the no-write pool-exhaustion result and rejects repairing the handle to 1. runtime_observed:no.

##### Discriminates

Bytewrap, null pointer equality boost, reuse ignores new initialization arguments and adds boost again

runtime_observed: no


#### STP-V38 — Force does not bypass shared owner references

##### Inputs

```json
{
  "entry": "remove owner reference twice in sequence",
  "overrides": {
    "voicePool.activeOrder": [
      "I"
    ],
    "voicePool.entries.O": "free entry with zero/absent semantic fields",
    "voicePool.entries.I.references": [
      "A",
      "B"
    ],
    "owners.A.handles": [
      501
    ],
    "owners.B": {
      "handles": [
        501
      ]
    },
    "records.100.Loop": 1,
    "records.100.FadeOut": 0,
    "records.100.StopInst": 0
  },
  "requests": [
    {
      "handle": 501,
      "owner": "A",
      "force": 1
    },
    {
      "handle": 501,
      "owner": "B",
      "force": 0
    }
  ],
  "B": "Only its handle list and reference identity are used; other B state is not read."
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
[
  {
    "step": 1,
    "actions": [
      "Remove first handle501 from A; find I; remove first A reference.",
      "B reference remains: return even though force=true. No retirement/fade/device work."
    ]
  },
  {
    "step": 2,
    "actions": [
      "Remove501 from B and remove its reference; I has no references left.",
      "Loop1, no active zero-target fade and force=false request retirement. Set retirement=true.",
      "FadeOut0 means no transition/device call; voice remains inUse/pending."
    ]
  }
]
```

##### exactOutputs

```json
[
  {
    "step": 1,
    "outputs": {
      "I": {
        "references": [
          "B"
        ],
        "retirement": 0,
        "state": "pending",
        "inUse": 1
      },
      "ownerAHandles": [],
      "ownerBHandles": [
        501
      ],
      "allOtherState": "unchanged",
      "deviceOperations": []
    }
  },
  {
    "step": 2,
    "outputs": {
      "I": {
        "references": [],
        "retirement": 1,
        "state": "pending",
        "inUse": 1
      },
      "ownerAHandles": [],
      "ownerBHandles": [],
      "allOtherState": "unchanged",
      "deviceOperations": []
    }
  }
]
```

##### Discriminates

Force does not bypass shared references; last loop reference loss requests retirement without immediate free

runtime_observed: no


#### STP-V39 — Invalid delayed zero-duration schedule retires before termination

##### Inputs

```json
{
  "entry": "schedule logical volume transition",
  "overrides": {
    "voicePool.activeOrder": [
      "I"
    ],
    "voicePool.entries.O": "free entry with zero/absent semantic fields",
    "owners.A.handles": [
      501
    ],
    "records.100.FadeOut": 0,
    "terminationRequestControl": 0,
    "managedRuntimeModulePresent": false,
    "terminationCallbacks": [],
    "initializerFinalizerCallbacks": [],
    "terminationProvider": "Receives the unsigned32 code selected by fatal handling and does not return",
    "diagnosticProvider": "assert reporter and diagnostic-formatting helpers return normally; no additional state effects before fatal reporting"
  },
  "arguments": {
    "duration": 0,
    "handle": 501,
    "target": 0,
    "delay": 1
  }
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
[
  "Find I; its single owner permits scheduling. FadeOut0 leaves effective duration0; no active-transition veto.",
  "Target0/effective duration0 sets I.retirement=true before validating delay.",
  "Delay1 with duration0 triggers fatal diagnostics. Diagnostic helpers return without extra state effects; supplied exit/finalizer callback lists are empty.",
  "Fatal handling sets terminationRequestControl1 then passes unsigned32 code4294967295 to the supplied nonreturning OS operation; scheduling has no normal return."
]
```

The empty initializerFinalizerCallbacks input covers both cleanup ranges. preFatalWrites lists scheduler-owned audio writes; the later fatal-foundation flag is separate. terminationCode below is the unsigned 32 argument received at the reached termination operation. The fatal-foundation flag is observed at that same pretermination request, while voiceBeforeFatal records the scheduler prefix. No post-exit memory or measured process-status oracle is asserted.

##### exactOutputs

```json
{
  "preFatalWrites": {
    "I_retirement": 1
  },
  "voiceBeforeFatal": {
    "currentVolume": 255,
    "activeTransition": 0,
    "startVolume": 0,
    "endVolume": 0,
    "startTick": 0,
    "endTick": 0
  },
  "normalReturn": false,
  "terminationCode": 4294967295,
  "fatalFoundationAtTerminationRequest": {
    "terminationRequestControl": 1
  },
  "deviceOperations": [],
  "draws": []
}
```

##### Discriminates

No delay normalization or immediate-volume fallback; retirement write precedes fatal path

runtime_observed: no


#### STP-V40 — Budget target underflow permits an above-budget source

##### Inputs

```json
{
  "entry": "asset admission cue100, synchronous=true, ageGate=false",
  "overrides": {
    "voicePool.activeOrder": [],
    "voicePool.entries": "all200 entries free, with zero/absent semantic fields",
    "owners.A.handles": [],
    "slots": "allvoicesnull/allmarkers0",
    "records.100.state": "unloaded",
    "records.100.raw": null,
    "records.100.rawSize": 0,
    "records.100.PCM": null,
    "records.100.PCMSize": 0,
    "records.101.state": "unloaded",
    "assets.usedBytes": 0,
    "assets.budget": 100,
    "assets.LRU": [],
    "assetFile100": {
      "totalBytes": 200,
      "RIFFSize": 192,
      "fmt": "PCM stereo2,22050Hz,16bit,avg88200,block4",
      "dataLength": 156,
      "PCM": "1,2,3,4 repeated39 times"
    }
  },
  "file": {
    "name": "fixture100.wav",
    "length": 200,
    "format": "Same canonical PCM stereo WAV header as shared fixture except RIFF size192 and data length156",
    "PCM": "[1,2,3,4] repeated39 times"
  },
  "provider": "Explicit synchronous sizing/read handles; successful storage allocation; complete successful reads and closes",
  "allocationProviderArgument": {
    "requestedBytes": 1000,
    "sourceBytes": 200,
    "additionalBytes": 800,
    "returnedStorage": "Asset200"
  }
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
[
  "Build DATA\\GLOBAL\\SFX\\fixture100.wav; open sizing handle, report200 bytes and close.",
  "Used0 plus requested200 exceeds budget100. Request free100; unsigned32 eviction target becomes4294967196. No entries are freed; eviction returns success.",
  "Open distinct read handle; validate size200; request1000 bytes of storage (source200 plus800), receive Asset200, read exactly200 bytes, close.",
  "Record raw/source size, account200 bytes, access100 and append100 LRU; mark loaded. Validate supplied complete PCM16/22050/stereo2 file and identify156 PCM bytes."
]
```

##### exactOutputs

```json
{
  "budgetPath": {
    "sum": 200,
    "requestedFree": 100,
    "evictionTarget": 4294967196,
    "freedEntries": 0,
    "evictionReturn": 1
  },
  "record100": {
    "raw": "Asset200",
    "rawSize": 200,
    "PCM": "Asset200.PCM",
    "PCMSize": 156,
    "state": "loaded",
    "Stereo": 1,
    "failure": 0,
    "lastAccess": 100
  },
  "usedBytes": 200,
  "LRU": [
    100
  ],
  "pendingCount": 0,
  "draws": [],
  "deviceOperations": [],
  "allOtherState": "unchanged",
  "requestedAllocationBytes": 1000,
  "accountedBytesAdded": 200
}
```

##### Discriminates

Unsigned target underflow accepts file above budget; budget is not a strict cap

runtime_observed: no


#### STP-V41 — Disposal leaves pending count and stale size

##### Inputs

```json
{
  "entry": "dispose asset cue100",
  "overrides": {
    "voicePool.activeOrder": [],
    "voicePool.entries": "all200 entries free, with zero/absent semantic fields",
    "owners.A.handles": [],
    "slots": "allvoicesnull/allmarkers0",
    "records.100.state": "pending",
    "records.100.raw": null,
    "records.100.rawSize": 48,
    "records.100.PCM": null,
    "records.100.PCMSize": 0,
    "records.100.descriptor": "D100",
    "records.101.state": "unloaded",
    "assets.usedBytes": 48,
    "assets.pendingCount": 1,
    "assets.LRU": [
      100
    ],
    "asyncDescriptorD100": {
      "valid": true,
      "ready": true,
      "ownsPayload": true,
      "payload": "Raw100",
      "size": 48,
      "file": null,
      "event": null,
      "next": null,
      "callbacks": [],
      "linked": false
    }
  }
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
[
  "No marked associated cue100 slot prevents disposal. Subtract48 from usedBytes; remove/free sole LRU100 node; clear access tick.",
  "Destroy valid ready owned descriptor D100: pollready succeeds, no wait/cancel/file/event effects.",
  "Free Raw100 payload; clear the descriptor payload reference and checked marker, retaining its ready and owned flags until D100 is freed.",
  "Clear record descriptor/raw reference and set unloaded. Do not decrement pendingCount; retain rawSize48 and other stated fields."
]
```

##### exactOutputs

```json
{
  "record100": {
    "state": "unloaded",
    "raw": null,
    "rawSize": 48,
    "descriptor": null,
    "PCM": null,
    "PCMSize": 0,
    "lastAccess": 0,
    "failure": 0
  },
  "usedBytes": 0,
  "LRU": [],
  "pendingCount": 1,
  "deviceOperations": [],
  "draws": [],
  "allOtherState": "unchanged",
  "payloadDescriptorFrees": [
    "Raw100",
    "D100"
  ],
  "orderedHeapFrees": [
    "the sole existing LRU node forcue100",
    "Raw100",
    "D100"
  ]
}
```

##### Discriminates

Pending asset disposal does not balance pendingCount; stale size remains

runtime_observed: no


#### STP-V42 — Tick zero bypasses normal prefetch throttles

##### Inputs

```json
{
  "entry": "pending-asset maintenance, two independent inputs",
  "overrides": {
    "clock.audioCounter": 0,
    "voicePool.activeOrder": [],
    "voicePool.entries": "all200 entries free, with zero/absent semantic fields",
    "owners.A.handles": [],
    "slots": "allvoicesnull/allmarkers0",
    "records.100.state": "unloaded",
    "records.100.raw": null,
    "records.100.rawSize": 0,
    "records.100.PCM": null,
    "records.100.PCMSize": 0,
    "records.100.Cache": 1,
    "records.101.state": "unloaded",
    "assets.usedBytes": 0,
    "assets.LRU": [],
    "assets.pendingCount": 15,
    "assets.lastPendingScan": 999,
    "assets.lastMemoryFailure": 0
  },
  "variations": [
    {
      "name": "tick0",
      "clock": {
        "tickArgument": 0,
        "audioCounter": 0
      }
    },
    {
      "name": "tick1",
      "clock": {
        "tickArgument": 1,
        "audioCounter": 1
      }
    }
  ],
  "file": "Complete48-byte shared WAV100; sizing and read handles return its exact bytes"
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
[
  {
    "variation": "tick0",
    "actions": [
      "Tick0 bypasses normal interval, pendingCount15 and lastMemoryFailure0 backoff gates. Scan valid records1–2933; only100 Cache1/unloaded needs admission.",
      "Request synchronous=true,ageGate=true. Build source path; sizing handle reports48; close. Open distinct read handle, request848 bytes (source48 plus800) and receive Asset48, read exactly48, close.",
      "Account48, access0, append100 LRU, mark loaded and validate shared PCM file. No pending counter increment/decrement. Finish full scan; lastPendingScan0."
    ]
  },
  {
    "variation": "tick1",
    "actions": [
      "Restart original input, not tick0 output. Unsigned32(1−999) permits scan.",
      "PendingCount15 blocks starting cached100; no other eligible row. No file/allocation activity; lastPendingScan1."
    ]
  }
]
```

##### exactOutputs

```json
[
  {
    "variation": "tick0",
    "outputs": {
      "record100": {
        "state": "loaded",
        "raw": "Asset48",
        "rawSize": 48,
        "PCM": "Asset48.PCM",
        "PCMSize": 4,
        "lastAccess": 0,
        "Stereo": 1,
        "failure": 0
      },
      "usedBytes": 48,
      "LRU": [
        100
      ],
      "pendingCount": 15,
      "lastPendingScan": 0,
      "draws": [],
      "deviceOperations": [],
      "allOtherState": "unchanged",
      "requestedAllocationBytes": 848,
      "accountedBytesAdded": 48
    }
  },
  {
    "variation": "tick1",
    "outputs": {
      "record100": {
        "state": "unloaded",
        "raw": null
      },
      "usedBytes": 0,
      "LRU": [],
      "pendingCount": 15,
      "lastPendingScan": 1,
      "fileOperations": [],
      "draws": [],
      "allOtherState": "unchanged"
    }
  }
]
```

##### Discriminates

Special tick 0 forces sync despite prefetch throttles; nonzero tick does not PendingCount15 with no live descriptor is supplied state; no repair is inferred.

runtime_observed: no


#### STP-V43 — Invalid source failure stays sticky after disposal

##### Inputs

```json
{
  "entry": "synchronous asset admission; disposal; synchronous retry",
  "overrides": {
    "voicePool.activeOrder": [],
    "voicePool.entries": "all200 entries free, with zero/absent semantic fields",
    "owners.A.handles": [],
    "slots": "allvoicesnull/allmarkers0",
    "records.100.state": "unloaded",
    "records.100.raw": null,
    "records.100.rawSize": 0,
    "records.100.PCM": null,
    "records.100.PCMSize": 0,
    "records.101.state": "unloaded",
    "assets.usedBytes": 0,
    "assets.LRU": [],
    "assetFile100": "48 bytes equal WAV100 except first4bytes BAD!"
  },
  "requests": [
    {
      "operation": "admit",
      "cue": 100,
      "synchronous": true,
      "ageGate": false
    },
    {
      "operation": "dispose",
      "cue": 100
    },
    {
      "operation": "admit",
      "cue": 100,
      "synchronous": true,
      "ageGate": false
    }
  ],
  "file": "48-byte shared WAV100 except its first four bytes are ASCII BAD!"
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
[
  {
    "step": 1,
    "actions": [
      "Build source path; sizing handle reports48; close. Open read handle, request848 bytes (source48 plus800) and receive Bad48, read exact48 and close.",
      "Account48, access100, append100 LRU and mark loaded before validation.",
      "Parser rejects BAD! signature; set failure=true while retaining loaded state/raw allocation/accounting."
    ]
  },
  {
    "step": 2,
    "actions": [
      "No marked associated cue100 slot; subtract48, remove/free LRU100 node and clear access.",
      "Free Bad48; clear raw reference and set unloaded; leave failure=true and stale rawSize48."
    ]
  },
  {
    "step": 3,
    "actions": [
      "Failure flag rejects retry immediately. No file/allocation effects and no implicit reset."
    ]
  }
]
```

##### exactOutputs

```json
[
  {
    "step": 1,
    "outputs": {
      "record100": {
        "state": "loaded",
        "raw": "Bad48",
        "rawSize": 48,
        "PCM": null,
        "PCMSize": 0,
        "failure": 1,
        "lastAccess": 100
      },
      "usedBytes": 48,
      "LRU": [
        100
      ],
      "allOtherState": "unchanged",
      "requestedAllocationBytes": 848,
      "accountedBytesAdded": 48
    }
  },
  {
    "step": 2,
    "outputs": {
      "record100": {
        "state": "unloaded",
        "raw": null,
        "rawSize": 48,
        "failure": 1,
        "lastAccess": 0,
        "PCM": null,
        "PCMSize": 0
      },
      "usedBytes": 0,
      "LRU": [],
      "allOtherState": "unchanged"
    }
  },
  {
    "step": 3,
    "outputs": {
      "record100": {
        "state": "unloaded",
        "raw": null,
        "failure": 1
      },
      "fileOperations": [],
      "heapOperations": [],
      "allOtherState": "unchanged"
    }
  }
]
```

##### Discriminates

Validation failure keeps loaded allocation; disposal does not clear the sticky failure gate

runtime_observed: no


#### STP-V44 — Negative Play retains copied data and history

##### Inputs

```json
{
  "entry": "audio coordinator100",
  "overrides": {
    "voicePool.activeOrder": [
      "I"
    ],
    "voicePool.entries.O": "free entry with zero/absent semantic fields",
    "slots.S0": "nullvoice marker0,class0 incompatible",
    "owners.A.handles": [
      501
    ],
    "records.100.StopInst": 0,
    "deviceResults.Play": -1
  },
  "suppliedAcousticUpdateProfile": "SUPPLIED-FINITE-ACOUSTIC-UPDATE",
  "acousticIntegrationGate": "S3D-R07/Q02 unresolved; scalar/radius subcalculations alone use verified R02/R05/R06"
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
[
  "Only I active; no retirement/fade effects. Bypass selection; no marked selected cue100; commit history100/7.",
  "Select S1; execute shared prepare and successful nonloop load/fill, copying4 PCM bytes then262140 zeros.",
  "Tentatively attach I; reset/apply/cache position(0,0,320),base255,factor0.",
  "Buffered Play(0,0,flags1)→−1 under synchronization. Keep marker0; return failure.",
  "Clear tentative S1 association; caller retains I pending/no slot. Nonloop loaded AsyncOnly0 queues deferred I; no playing property/access/LRU refresh.",
  "Cleanup removes active I/A handle501/I reference A; pending no-slot stop helper makes no Stop/query; clear inUse and free temporary deferred node.",
  "Do not roll back history, payload, properties or cache. Gains unchanged; lastScalarTick100."
]
```

##### exactOutputs

Caller logical-state/order assertions use the declared normal dependency returns. Acoustic-update getter/base/property/cache assertions, including its SetVolume calls, are conditional on the supplied finite profile and unresolved S3D-R07/Q02 integration.

```json
{
  "draws": [],
  "seedA": [
    1,
    0
  ],
  "history100": {
    "recent": 100,
    "older": 7
  },
  "I": {
    "inUse": 0,
    "state": "pending",
    "slot": null,
    "references": [],
    "currentVolume": 255,
    "activeTransition": 0,
    "retirement": 0
  },
  "ownerAHandles": [],
  "activeVoices": [],
  "S1Voice": null,
  "S1Cache": {
    "XYZ": [
      0,
      0,
      320
    ],
    "baseVolume": 255,
    "factor": 0
  },
  "W1": {
    "marker": 0,
    "data": "WAV100.PCM",
    "dataLength": 4,
    "writtenBytes": 4,
    "endDataFlag": 1,
    "halfFlag": 1,
    "endMarker": 4,
    "payload": "bytes1,2,3,4 then262140zeros"
  },
  "LRU": [
    100,
    101
  ],
  "record100Access": 90,
  "usedBytes": 96,
  "pendingCount": 0,
  "lastScalarTick": 100,
  "deviceStopCalls": 0,
  "allUnlistedState": "unchanged from explicit fixture/overrides"
}
```

##### Discriminates

Valid load/copy may be retained despite negativePlay; failedPlay clears tentative slot association and removes nonloop incoming voice without deviceStop or history rollback; distinct from no-slot failure and ignored negativeLock.

runtime_observed: no


#### STP-V45 — nonloop drain through Sound worker handoff

Inputs: STP-V36 creation, initial pump, successful Q1 callback/destruction only, with Loop 0. No loop-reset or Q2 stage. Start checkpoint: stream S state playing/cache 4/progress 16388/tail 16378/write 16384/file F references 2 / stream S references 1; file F logical position 54; wrapper W marker 1/stream F. Positions success 4 then success 10. Drain Lock at 16384 requests/returns 16384 bytes; Unlock 0. Stop returns[0,0,0]. One Sound worker iteration: singleton W, Wait[258,0] at 50 ms, playlist disabled, iconic 0 unchanged, fade inactive, GetStatus success/status 0. All producer state carries into this worker.

Seeds/nodraw: none. orderedDraws: [].

Trace: first EOF query cache 4<10, free 245764; Lock/zero 16384/Unlock; progress 32772/tail 32762/write 32768; draining. Next query cache 10, Stop#1 ignored, finished. Sound worker status check precedes stream lookup; finished lookup fails and records error 0x85100072. Nested Sound stop ends S, drops stream S references 1→0 then file F references 2→1, no Release, frees S and clears file stream flag; Stop#2; owner close file F references 1→0/CloseH/free file F; wrapper W stream null/marker 0. Worker then unconditionally Stop#3, marker 0; next Wait exits.

ExactOutputs: F/S freed; wrapper W marker 0/stream null/inner buffer B retained/endMarker −1; other wrapper W fields retained; wrapper-list node retained; scratch retained; current null, queues empty. Last live S checkpoint finished/cache 10/progress 32772/tail 32762/write 32768. Three ordered Stops, Release 0, CloseH 1, no further reads/Play. Device payload after Q1 first region as STP-V36; drain zeroes B[16384,32768); remainder retains prior bytes. Final engine error 0x85100072. Handoff to the separate logical fixture carries every producer-only global/list/provider/resource state; only logical associations are initialized there. This vector does not itself detach/free a logical voice.

Additional exact output for the main completed path and Stop-result variations: projected ordered lifecycle log=[B.Stop#1, S.references1→0, F.references2→1, free S, F.streamFlag0, B.Stop#2, F.references1→0, CloseHandle(H), free F, W.stream=null/W.marker0, B.Stop#3]. Substitute the declared provider results for the numbered Stops. Query/Lock-prefix variations stop at their declared earlier checkpoint and assert only the events reached there.

Discriminates: nonloop drain, physical-versus-cache completion, three Stop calls before logical handoff, retained external owner until Sound close. runtime_observed: no.

Independent variations: Stop#1−1 or all three −1 retain the same engine outputs with physical result unspecified. First EOF position HRESULT −1 stops after that pump with playing/cache 4/progress 16388/tail 16378/write 16384, no Lock/Stop. First EOF position 4 then Lock −1 stops after that pump with draining/cache 4 and unchanged counters, no Unlock/Stop. A failure entering draining retains draining/cache/counters and calls no Stop.


#### STP-V46 — stop while read remains current

Inputs: REQUEST-RETENTION. Sound stop runs during the selected read; Stop 0 and owner close returns normally. The read later returns BOOL 1/actual 6/bytes 5..10. No other interleaving.

Seeds/nodraw: none. orderedDraws: [].

Trace: clear Q1 device destination first; cancel queued matches (none); unlink S, drop stream S references 2→1; clear F stream flag; end returns 1. Sound calls Stop, closes owner file F references 3→2, wrapper W stream null/marker 0. Read returns 6; current destination null skips callback and zero-read Stop. Request destructor drops file F references 2→1, then stream S references 1→0; stream destructor drops file F references 1→0, closes/frees F, skips Release for borrowed buffer B, frees S; Q freed; current cleared last.

ExactOutputs: immediate Sound checkpoint file F references 2 / stream S references 1 still live, S unlinked, current request Q1/destination null, file F stream flag 0, wrapper W marker 0/stream null. Final F/S/Q freed, current null, queue empty, B unchanged/all 0xaa, scratch retains allocation with read prefix[5..10], untouched bytes retained. One Stop, one CloseH, no Release/Lock/Unlock/Play. wrapper W inner B remains; other wrapper W fields unchanged.

Additional exact ordered outputs are two sequential projections. Immediate stop projection=[Q1.deviceDestination=null, cancel queued matches for B (none), unlink S, S.references2→1, F.streamFlag0, B.Stop→0, drop external F owner 3→2, W.stream=null/W.marker0]. Later completion projection=[read actual 6 into retained scratch, Q1 drops F2→1, Q1 drops S1→0, S drops F1→0, CloseHandle(H), free F, free S, free Q1, current=null]. There is no device callback or Release between those events.

Discriminates: wrapper clearing precedes retained-resource destruction; destination clear precedes file flag clear; request file drop precedes stream drop. runtime_observed: no.


#### STP-V47 — Completed marker detaches before later owner cleanup

##### Inputs

```json
{
  "entry": "complete marker-to-logical bridge, then following audio coordinator101",
  "overrides": {
    "voicePool.activeOrder": [
      "I"
    ],
    "voicePool.entries.O": "free entry with zero/absent semantic fields",
    "slots.S0": "nullvoice marker0,class0 incompatible",
    "owners.A.handles": [
      501
    ],
    "voicePool.entries.I.state": "playing",
    "voicePool.entries.I.slot": "S1",
    "voicePool.entries.I.bookmark": 77,
    "slots.S1.voice": "I",
    "wrappers.W1.marker": 0,
    "records.100.StopInst": 0,
    "records.100.Loop": 0,
    "records.100.Stream": 1,
    "wrappers.W1.streamHandle": null
  },
  "clientClock": {
    "bridge": "No coordinator dispatch implied",
    "followingCoordinator": {
      "tickArgument": 101,
      "audioCounter": 101
    }
  },
  "producerHandoff": {
    "source": "Complete STP-V45 output after its worker service; carry every producer field and resource identity into this suffix",
    "knownFields": {
      "wrapper": "W1",
      "innerBuffer": "B1 retained",
      "marker": 0,
      "streamHandle": null,
      "endMarker": -1,
      "streamContext": "freed",
      "sourceFile": "freed",
      "sourceHandle": "closed exactly once",
      "wrapperList": [
        "W1"
      ],
      "wrapperListNode": "retained",
      "scratch": "retained",
      "queue": [],
      "currentRequest": null,
      "precedingStopCalls": 3
    },
    "mergeRule": "Producer state takes precedence over common wrapper/backend/worker/provider state. Fill only Client fields or other fields absent from producer output. Do not reset the worker wrapper/list/globals from the common fixture."
  },
  "independentVariation": {
    "name": "Loop1 bridge only",
    "override": {
      "records.100.Loop": 1
    },
    "scope": "Marker bridge only, no following coordinator"
  }
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
[
  {
    "stage": "nonloop marker bridge",
    "actions": [
      "Scan16 slots in order; skip empty S0. S1 has I and decoded Loop0. Read marker→0, then stream-present predicate→false.",
      "Do not query bookmark, stop a buffer or enter backend synchronization. Set I stopped, clear I.slot and S1 association. Retain inUse, bookmark77, active list and owner references."
    ]
  },
  {
    "stage": "following coordinator101",
    "actions": [
      "Only stopped I active; no retirement-prepass action. Loop0 queues deferred I without admission or property/cache/history/access work.",
      "Remove active I, A handle501 and I reference A; stop helper sees stopped/no slot and performs no query/Stop. Clear inUse and free temporary deferred node.",
      "Gains remain100; lastScalarTick101. Every producer-only field remains exactly the carried output."
    ]
  },
  {
    "variation": "independent Loop1 bridge only",
    "actions": [
      "Restart producer/Client entering state, except decoded Loop1. Skip I before either marker getter or stream predicate.",
      "No association, voice, owner, producer or device state changes."
    ]
  }
]
```

##### exactOutputs

```json
{
  "steps": [
    {
      "stage": "bridge",
      "outputs": {
        "I": {
          "inUse": 1,
          "state": "stopped",
          "slot": null,
          "bookmark": 77,
          "references": [
            "A"
          ]
        },
        "ownerAHandles": [
          501
        ],
        "activeVoices": [
          "I"
        ],
        "S1Voice": null,
        "W1Marker": 0,
        "W1Stream": null,
        "deviceOperations": [],
        "listAllocations": [],
        "allOtherState": "unchanged"
      }
    },
    {
      "stage": "following coordinator101",
      "outputs": {
        "I": {
          "inUse": 0,
          "state": "stopped",
          "slot": null,
          "bookmark": 77,
          "references": []
        },
        "ownerAHandles": [],
        "activeVoices": [],
        "S1Voice": null,
        "W1Marker": 0,
        "W1Stream": null,
        "deviceOperations": [],
        "history100": {
          "recent": 7,
          "older": 6
        },
        "LRU": [
          100,
          101
        ],
        "accessTicks": {
          "100": 90,
          "101": 90
        },
        "usedBytes": 96,
        "pendingCount": 0,
        "lastScalarTick": 101,
        "allOtherState": "unchanged"
      }
    }
  ],
  "independentLoop1BridgeOnly": {
    "I": {
      "inUse": 1,
      "state": "playing",
      "slot": "S1",
      "bookmark": 77,
      "references": [
        "A"
      ]
    },
    "ownerAHandles": [
      501
    ],
    "activeVoices": [
      "I"
    ],
    "S1Voice": "I",
    "markerGetterCalls": 0,
    "streamPredicateCalls": 0,
    "deviceOperations": [],
    "allOtherState": "unchanged"
  },
  "producerState": "Every carried producer field remains unchanged throughout; no additional Stop or bookmark/device-position query in either Client stage."
}
```

##### Discriminates

Backendmarker0 does not free voice or owners during marker bridge; bookmark 77 is not zeroed by unrequested query; following coordinator cleanup has no duplicate backend Stop; Loop1 skips both getters. The three Stop calls belong to the producer prefix; this suffix adds none.

runtime_observed: no


#### STP-V48 — buffered equality and strict completion

Inputs: BUFFERED-WORKER, all defaults, carry state between its two iterations.

Seeds/nodraw: none. orderedDraws: [].

Trace: first position 4 unwraps played 4; endData 1/loop 0; physical<half and halfFlag 1 make no clear; played==dataLength retains marker. Second position 5 unwraps played 5; same no-clear branch; played>dataLength requests Stop then marker 0. Next Wait exits.

ExactOutputs: checkpoints(played 4,marker 1,Stop 0), then(played 5,marker 0,Stop 1). All other wrapper W fields and B bytes unchanged; W/B retained; no Lock/Unlock/stream call/Release/volume call. Discriminates: equality versus strict unsigned greater-than. runtime_observed: no.

Independent variations reset BUFFERED-WORKER and use one iteration Wait[258,0]/one status: Stop −1 with position 5 has the same engine result. Position HRESULT −1 retains played 0/marker 1, no clear/Stop. GetStatus HRESULT −1 with successful position 5 returns from status helper without Restore but worker continues; played 5/marker 0/Stop 1.


#### STP-V49 — backend millisecond fade

Inputs: BUFFERED-WORKER reset; one iteration Wait[258,0], one GetStatus success 0; flags 4, acoustic-conversion mode 2, volume scalar 255, stored volume 100; fade active 1/from 100/target 0/start 1000/end 2000. SetVolume −1 and Stop −1. Each subcase independently resets. Clock is signed 32 milliseconds.

Seeds/nodraw: none. orderedDraws: [].

Trace: time 999 (before start) or 2001 (after end) selects endpoint; store volume 0, convert exact zero to attenuation −10000 (flags 4/acoustic-conversion mode 2 bypass property factor), SetVolume −10000 ignored; target 0 Stop ignored, marker 0, fade inactive. Buffered position is skipped. Time 2000 instead takes inclusive interpolation; volume 0, SetVolume −10000, fade remains active; supply successful position 0, which keeps marker 1/played 0.

ExactOutputs: before/after: stored 0/active 0/marker 0/played 0; device log GetStatus,SetVolume(−10000),Stop; GetPosition calls 0. Exact end: stored 0/active 1/marker 1/played 0; device log GetStatus,SetVolume(−10000),GetPosition; Stop 0. Other fields/payload unchanged. Discriminates: signed outside-interval endpoint even before start, versus inclusive end; independent from logical Client fade timing. runtime_observed: no.


#### STP-V50 — lost-buffer recovery

Inputs: BUFFERED-WORKER reset; one iteration Wait[258,0]; GetStatus success/status 2; Restore results[0x88780096,−1,0]; Sleep 10 returns; Stop −1; no GetPosition.

Seeds/nodraw: none. orderedDraws: [].

Trace: first Restore result sleeps 10 then retries; next negative retries immediately; success invokes stop helper with stream null, Stop then marker 0. Context unchanged/fade 0; buffered position skipped. Exit on next Wait.

ExactOutputs: marker 0; all other wrapper W fields/payload unchanged, W/B retained; ordered log GetStatus,Restore,Sleep(10),Restore,Restore,Stop; no GetPosition/Release. Discriminates: special retry sleep, unbounded retry policy conditional on providers, recovery Stop rather than Release. runtime_observed: no.


#### STP-V51 — half-buffer clearing

Inputs: BUFFERED-WORKER reset; one iteration Wait[258,0]/one successful status; B firsthalf 0x7e repeated 131072, secondhalf 0x7f repeated 131072. Each subcase supplies one successful position(play below,write 0). Clearing helper calls Lock(offset,requestedLength,&firstRegion,&returnedLength,null,null,0); success zeros returned first length then Unlock(firstRegion,returnedLength,null,0), ignoring result. Every subcase independent.

Seeds/nodraw: none. orderedDraws: [].

Trace and exactOutputs:

- halfFlag 0, play 3, Lock(secondhalf,131072) success/fullregion, Unlock −1: played 3/halfFlag 1/marker 1; secondhalf zero, firsthalf unchanged; one Unlock, no Stop.
- Same but Lock −1: played 3/halfFlag 1/marker 1; all payload unchanged; no Unlock/Stop.
- halfFlag 1, play 131072, Lock(firsthalf,131072) success/fullregion, Unlock 0, Stop −1: zero firsthalf, retain secondhalf; played 131072/halfFlag 0/marker 0; suffix GetPosition,Lock,Unlock,Stop.
- Same but Lock −1: payload unchanged; played 131072/halfFlag 0/marker 0; suffix GetPosition,Lock,Stop.
- halfFlag 0, play 3, Lock(secondhalf,131072) success but supplied first length 16, Unlock 0: zero only first 16 bytes of secondhalf; remainder 131056 bytes 0x7f, firsthalf unchanged; played 3/halfFlag 1/marker 1; Unlock length 16, no Stop. This is an artificial supplied return shape, not a real platform assertion.

All other W state unchanged; no stream/Release/Play. Discriminates: half selector toggles despite helper failure, clearing precedes completion Stop, returned length controls this clearing helper. runtime_observed: no.


#### STP-V52 — buffered source refill

Inputs: BUFFERED-WORKER prefix with exactly one iteration: Wait[258,0] at 50 ms, one GetStatus success/status 0, singleton W, unchanged iconic 0, scheduled playlist disabled, backend fade inactive. Preserve live identities W.inner=B and the valid device interface. Override ring capacity 8/half 4, B eight 0x7e bytes, sourceP=[1..10], W data pointerP/length 10/played 0/written 8/marker 1/endData 0/halfFlag 1/endMarker −1/loop 0/loopStart 0/stream null. W flags 0/stored volume 0/property scalar 0; fade from/target/start/end 0; no secondary/property/3D interface. Other semantic fields remain zero/null without zeroing the named live buffer or provider identities. Position 6 succeeds; Lock(destination 0,length 4) success/full 4; Unlock −1. The small ring is a valid supplied arithmetic checkpoint, not normal game allocation policy.

Seeds/nodraw: none. orderedDraws: [].

Trace: played 6, unsigned written−played 2<half 4; refill(source 8,destination 0,length 4). Copy 9,10, zero 2; written 10/endData 1. Worker ignores helper result, sets endMarker −1 and halfFlag=(destination!=0)=0. It does not revisit completion this iteration.

ExactOutputs: B=[9,10,0,0,0x7e,0x7e,0x7e,0x7e]; played 6/written 10/endData 1/halfFlag 0/endMarker −1/marker 1; others unchanged, source unchanged, no Stop/Play/stream/Release. Discriminates: requested refill length, end transition delayed to later iteration, bookkeeping independent of helper success. runtime_observed: no.

Independent variations: Lock −1 with initial endMarker 3 retains written 8/endData 0/B but produces played 6/halfFlag 0/endMarker −1/marker 1, no Unlock. Position 4 yields equality gap 4, no refill, played 4 only. Initial endMarker 3 with fullLock copies 2/pads 1, B=[9,10,0,0x7e,0x7e,0x7e,0x7e,0x7e], written 10/endData 1/halfFlag 0/endMarker −1; Unlock still 4. Written 2/position 3 triggers the second condition played>written despite unsigned gap 0xffffffff; Lock destination 2/length 4 copies 3..6, written 6/endData 0/halfFlag 1, B=[0x7e,0x7e,3,4,5,6,0x7e,0x7e]. Loop 1/source length 6/source[1..6]/loopStart 2 with basewritten 8/position 6 wraps source 8→4 then 6→2: output[5,6,3,4,0x7e,0x7e,0x7e,0x7e], written 12/endData 0/halfFlag 0/marker 1. For this refill helper successful copying uses requested length; returned region length is only used for Unlock. No short-region success is supplied here.


#### STP-V53 — Fresh nonzero playback offset is independent of saved bookmark

##### Inputs

```json
{
  "entry": "complete logical request creation only",
  "overrides": {
    "controls.selectedEmitter": null,
    "voicePool.activeOrder": [],
    "voicePool.entries": "All200 entries free, with zero/absent semantic fields",
    "owners.A.handles": [],
    "voicePool.nextHandleCounter": 501,
    "records.100.Priority": 250,
    "records.100.Compound": 0
  },
  "arguments": {
    "cue": 100,
    "emitter": null,
    "delay": 0,
    "requestFlags": 1,
    "playbackOffset": 123
  },
  "slotReferenceLimitation": "Retain S0 reference to pool index0; N0 reuses that entry as in STP-V37."
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
[
  "Checks pass; Compound0 bypasses reuse. Allocate first free index0, reset semantic fields, mark inUse, prepend N0 and increment handle501→502.",
  "Initialize cue100/start100/flags1/currentVolume255/priority250; store supplied playbackOffset123. Saved bookmark remains0.",
  "Null emitter initializes position(0,0,320), distanceKey0 and contextFactor0. Null equals selected null: priority250+80 wraps to74.",
  "Attach no owner; return502. No coordinator/admission/stream seek is invoked."
]
```

##### exactOutputs

```json
{
  "returnHandle": 502,
  "counter": 502,
  "allocatedPoolIndex": 0,
  "newVoice": {
    "inUse": 1,
    "cue": 100,
    "handle": 502,
    "distanceKey": 0,
    "currentVolume": 255,
    "contextFactor": 0,
    "playbackOffset": 123,
    "references": [],
    "slot": null,
    "state": "pending",
    "start": 100,
    "requestFlags": 1,
    "priority": 74,
    "retirement": 0,
    "bookmark": 0,
    "activeTransition": 0,
    "position": [
      0,
      0,
      320
    ]
  },
  "activeVoices": [
    "N0"
  ],
  "allOtherVoiceFields": "zero/false/null/empty according to shared voice schema",
  "seedA": [
    1,
    0
  ],
  "deviceOperations": [],
  "draws": [],
  "listAllocations": [
    "one active node {value:N0,next:null}"
  ],
  "listFrees": [],
  "allOtherState": "unchanged"
}
```

##### Discriminates

A fresh nonzero request offset is initialized as playbackOffset; the saved bookmarkfield stays 0. This does not execute or claim subsequent stream seek behavior.

runtime_observed: no


#### STP-V54 — Pool exhaustion returns before initialization

##### Inputs

```json
{
  "entry": "complete logical request creation only",
  "overrides": {
    "controls.selectedEmitter": null,
    "voicePool.activeOrder": [],
    "voicePool.entries": "All200 entries have inUse=true and all other entry fields zero/absent; no free entry. Active order is supplied separately.",
    "owners.A.handles": [],
    "voicePool.nextHandleCounter": 501,
    "records.100.Priority": 250,
    "records.100.Compound": 0
  },
  "arguments": {
    "cue": 100,
    "emitter": null,
    "delay": 0,
    "requestFlags": 1,
    "playbackOffset": 0
  },
  "suppliedState": "All200 entries inUse=true, other voice fields zero/absent; active list remains empty. This synthetic occupied-but-unlinked pool is not normalized into a lifecycle invariant."
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
[
  "Audio/positive valid cue/Volume255 checks pass; Compound0 bypasses reuse.",
  "Scan all200 inUse flags in ascending pool order; none free. Return no voice.",
  "Return handle0 immediately, before counter increment, list allocation, field initialization, selected-emitter boost or owner attachment."
]
```

##### exactOutputs

```json
{
  "returnHandle": 0,
  "counter": 501,
  "allocatedPoolIndex": null,
  "activeVoices": [],
  "ownerAHandles": [],
  "poolWrites": [],
  "listAllocations": [],
  "listFrees": [],
  "associationWrites": [],
  "deviceOperations": [],
  "draws": [],
  "seedA": [
    1,
    0
  ],
  "allOtherState": "unchanged"
}
```

##### Discriminates

No counter increment, active-node allocation, pool zeroing, field initialization or association attempt when every pool entry is occupied. Supplied occupied-but-unlinked pool is a synthetic allocator input, not an assertion about normal engine lifecycle.

runtime_observed: no


#### STP-V55 — Special-range selection replaces the ordinary candidate, including with none

##### Inputs

```json
{
  "entry": "complete duplicate-selection decision segment only",
  "overrides": {
    "controls.tableCount": 4657,
    "controls.validRecords": "Every row0..4656 has valid record storage; defaultzero except inherited100/101 and explicit2934/4656/2933 below",
    "voicePool.activeOrder": [
      "I",
      "O",
      "J"
    ],
    "voicePool.entries.freeIndices": "3..199",
    "voicePool.entries.I.cue": 2934,
    "voicePool.entries.J": {
      "index": 2,
      "inUse": 1,
      "cue": 4656,
      "handle": 499,
      "distanceKey": 0,
      "currentVolume": 255,
      "contextFactor": 0,
      "playbackOffset": 0,
      "references": [
        "A"
      ],
      "slot": "S2",
      "state": "playing",
      "start": 97,
      "requestFlags": 1,
      "priority": 80,
      "retirement": 0,
      "bookmark": 0,
      "activeTransition": 0,
      "startVolume": 0,
      "endVolume": 0,
      "startTick": 0,
      "endTick": 0,
      "position": [
        0,
        0,
        320
      ]
    },
    "slots.S2.voice": "J",
    "owners.A.handles": [
      501,
      500,
      499
    ],
    "records.2934": {
      "Volume": 255,
      "GroupSize": 2,
      "Loop": 0,
      "FadeIn": 0,
      "FadeOut": 0,
      "DeferInst": 1,
      "StopInst": 0,
      "Duration": 0,
      "Compound": 0,
      "Range": 0,
      "Reverb": 0,
      "Cache": 0,
      "AsyncOnly": 0,
      "Priority": 20,
      "Stream": 0,
      "Stereo": 1,
      "Tracking": 0,
      "Interrupt": 0,
      "Channel": 0,
      "Block1": 0,
      "Block2": 0,
      "FamilyBase": 100,
      "historyRecent": 7,
      "historyOlder": 6,
      "raw": "WAV100",
      "rawSize": 48,
      "PCM": "WAV100.PCM",
      "PCMSize": 4,
      "lastAccess": 90,
      "failure": 0,
      "descriptor": null,
      "state": "loaded",
      "referenceCount": 0,
      "filename": "fixture100.wav"
    },
    "records.4656": {
      "Volume": 255,
      "GroupSize": 2,
      "Loop": 0,
      "FadeIn": 0,
      "FadeOut": 0,
      "DeferInst": 0,
      "StopInst": 0,
      "Duration": 0,
      "Compound": 0,
      "Range": 0,
      "Reverb": 0,
      "Cache": 0,
      "AsyncOnly": 0,
      "Priority": 20,
      "Stream": 0,
      "Stereo": 1,
      "Tracking": 0,
      "Interrupt": 0,
      "Channel": 0,
      "Block1": 0,
      "Block2": 0,
      "FamilyBase": 200,
      "historyRecent": 7,
      "historyOlder": 6,
      "raw": "WAV100",
      "rawSize": 48,
      "PCM": "WAV100.PCM",
      "PCMSize": 4,
      "lastAccess": 90,
      "failure": 0,
      "descriptor": null,
      "state": "loaded",
      "referenceCount": 0,
      "filename": "fixture100.wav"
    },
    "records.2933": {
      "Volume": 255,
      "GroupSize": 2,
      "Loop": 0,
      "FadeIn": 0,
      "FadeOut": 0,
      "DeferInst": 0,
      "StopInst": 0,
      "Duration": 0,
      "Compound": 0,
      "Range": 0,
      "Reverb": 0,
      "Cache": 0,
      "AsyncOnly": 0,
      "Priority": 20,
      "Stream": 0,
      "Stereo": 1,
      "Tracking": 0,
      "Interrupt": 0,
      "Channel": 0,
      "Block1": 0,
      "Block2": 0,
      "FamilyBase": 200,
      "historyRecent": 7,
      "historyOlder": 6,
      "raw": "WAV100",
      "rawSize": 48,
      "PCM": "WAV100.PCM",
      "PCMSize": 4,
      "lastAccess": 90,
      "failure": 0,
      "descriptor": null,
      "state": "loaded",
      "referenceCount": 0,
      "filename": "fixture100.wav"
    }
  },
  "selectorInputs": {
    "incoming": "I",
    "incomingCue": 2934,
    "soleOwner": "A",
    "positionIsNonzero": true,
    "decodedRecord": 2934,
    "deferredList": [],
    "activeScanOrder": [
      "I",
      "O",
      "J"
    ],
    "candidateVoices": {
      "O": {
        "cue": 101,
        "state": "playing",
        "references": [
          "A"
        ],
        "family": 100
      },
      "J": {
        "cue": 4656,
        "state": "playing",
        "references": [
          "A"
        ],
        "family": 200
      }
    },
    "newRecords": "Complete copies of shared record100, materialized first; record2934 has family100/DeferInst1/StopInst0;4656 and2933 have family200/DeferInst0/StopInst0"
  },
  "boundary": "Begin at candidate decision with supplied semantic inputs; stop before enqueue, retirement or admission effects. Earlier coordinator sorting and later cleanup are not invoked.",
  "independentVariations": [
    {
      "name": "special search finds none",
      "override": {
        "voicePool.entries.J.cue": 2933
      }
    },
    {
      "name": "both policy flags zero; special lookup still selectsJ",
      "overrides": {
        "records.2934.DeferInst": 0,
        "records.2934.StopInst": 0
      },
      "initialCandidate": null,
      "unchangedInputs": "IncomingI cue2934 has sole ownerA; playingJ cue4656/refA remains in scan order."
    }
  ]
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
[
  {
    "variation": "special candidate",
    "actions": [
      "DeferInst1 runs family search with family100, sole owner A and exclude I. Skip I; choose O; reject J family200. Ordinary candidate O.",
      "Incoming cue2934 and sole owner A enable special search. Skip I; reject O cue101; accept first matching J cue4656 at inclusive upper boundary.",
      "Replace ordinary O with special J. Incoming references are present, so no no-reference exception. Selected candidate is J; subsequent policy would defer incoming I, outside this segment."
    ]
  },
  {
    "variation": "special search finds none",
    "actions": [
      "Restart entering fixture with J cue2933. Family search still returns O.",
      "Special search rejects O101 and J2933 outside inclusive2934–4656 range; returns none.",
      "Replace ordinary O with none. Decision proceeds toward admission, outside this segment; no candidate retained as fallback."
    ]
  },
  {
    "variation": "both policy flags zero",
    "actions": [
      "Restart entering fixture with decoded DeferInst0/StopInst0 and explicitly supplied initial candidate none.",
      "Skip ordinary family search completely; ordinary candidate remains none.",
      "Sole ownerA and incoming2934 still enable the special lookup independently of both policy flags. Skip I/O101 and select J4656/refA.",
      "Final candidateJ is retained at the segment boundary. Subsequent policy would defer incomingI outside this segment; no memory/list/device effect occurs here."
    ]
  }
]
```

##### exactOutputs

```json
{
  "specialCandidate": {
    "ordinaryCandidate": "O",
    "specialCandidate": "J",
    "finalCandidate": "J",
    "nextBehavior": "would defer incoming I; not executed",
    "memoryChanges": [],
    "allocations": [],
    "frees": [],
    "deviceOperations": [],
    "allOtherState": "unchanged"
  },
  "specialNone": {
    "ordinaryCandidate": "O",
    "specialCandidate": null,
    "finalCandidate": null,
    "nextBehavior": "would continue admission; not executed",
    "memoryChanges": [],
    "allocations": [],
    "frees": [],
    "deviceOperations": [],
    "allOtherState": "unchanged"
  },
  "bothFlagsZero": {
    "ordinarySearchCalls": 0,
    "ordinaryCandidate": null,
    "specialSearchCalls": 1,
    "specialCandidate": "J",
    "finalCandidate": "J",
    "nextBehavior": "would defer incoming I; not executed",
    "memoryChanges": [],
    "allocations": [],
    "frees": [],
    "deviceOperations": [],
    "allOtherState": "unchanged"
  }
}
```

##### Discriminates

Incoming2934 and candidate 4656 demonstrate inclusive special boundaries. A special search returning none erases the ordinary candidate. Selection is first matching special voice in scan order, not same-family selection. Special lookup remains enabled by sole owner and range even when both DeferInst and StopInst are 0; ordinary family search is skipped in that variation.

runtime_observed: no


#### STP-V56 — No-reference positional exception changes defer policy

##### Inputs

```json
{
  "entry": "complete audio coordinator100; independent position variations",
  "overrides": {
    "voicePool.entries.I.references": [],
    "owners.A.handles": [
      500
    ],
    "records.100.Loop": 1,
    "records.100.DeferInst": 1,
    "records.100.StopInst": 0,
    "slots.allClasses": 0,
    "records.101.Loop": 1,
    "records.101.DeferInst": 0,
    "records.101.StopInst": 0
  },
  "variations": [
    {
      "name": "positioned",
      "position": [
        0,
        0,
        320
      ],
      "distanceKey": 0
    },
    {
      "name": "zero position",
      "position": [
        0,
        0,
        0
      ],
      "distanceKey": 0
    }
  ],
  "control": "I references empty; A handles only500 for O. I Loop1/DeferInst1/StopInst0; O record101 explicitly Loop1/DeferInst0/StopInst0. All16 classes0; incoming Stereo1 requires class1.",
  "suppliedAcousticUpdateProfile": "SUPPLIED-FINITE-ACOUSTIC-UPDATE",
  "acousticIntegrationGate": "S3D-R07/Q02 unresolved; scalar/radius subcalculations alone use verified R02/R05/R06"
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
[
  {
    "variation": "positioned",
    "actions": [
      "Order I100 priority then O90; no retirement/fade/tracking effects. Empty I references imply no owner filter.",
      "z320 differs from zero sentinel even though distanceKey0; family search chooses O. Sole-owner special override is absent.",
      "DeferInst1 plus absent reference list plus nonzero position bypasses candidate deferral.",
      "Call admission once. Selected-cue marked search reads all16 markers; W0 marked but its associated cue101 differs100. No recent-cue suppression. Commit history100/7.",
      "All slots incompatible with class1, so no slot/victim; no prepare/load/property/Play. Loop1 keeps I pending/inUse.",
      "O playing properties cache-match; refresh101 access100. Free old LRU101 node then allocate/append its new tail node. No deferred cleanup; gains100,lastScalarTick100."
    ]
  },
  {
    "variation": "zero position",
    "actions": [
      "Restart input except XYZ(0,0,0). No owner filter; family search returns O; position flag is zero.",
      "DeferInst exception does not apply. Candidate O plus StopInst0 sets I retirement=true and allocates/prepends deferred I. Skip admission/history.",
      "O cache-match refreshes101: free old LRU101 node then allocate/append new tail node.",
      "Free active I node, clear empty I references, skip backend work for pending no-slot I, clear inUse and free temporary deferred I node. Gains100,lastScalarTick100."
    ]
  }
]
```

##### exactOutputs

Caller logical-state/order assertions use the declared normal dependency returns. Acoustic-update getter/base/property/cache assertions, including its SetVolume calls, are conditional on the supplied finite profile and unresolved S3D-R07/Q02 integration.

```json
[
  {
    "variation": "positioned no-ref request bypasses DeferInst candidate",
    "outputs": {
      "I": {
        "inUse": 1,
        "state": "pending",
        "slot": null,
        "retirement": 0,
        "currentVolume": 255,
        "bookmark": 0,
        "playbackOffset": 0,
        "references": [],
        "position": [
          0,
          0,
          320
        ],
        "distanceKey": 0
      },
      "O": "all semantic voice fields unchanged; remains inUse1/state1/S0/refA",
      "activeVoices": [
        "I",
        "O"
      ],
      "ownerAHandles": [
        500
      ],
      "S0Voice": "O",
      "S1Voice": null,
      "history100": {
        "recent": 100,
        "older": 7
      },
      "history101": {
        "recent": 0,
        "older": 0
      },
      "accessTicks": {
        "100": 90,
        "101": 100
      },
      "LRU": [
        100,
        101
      ],
      "usedBytes": 96,
      "pendingCount": 0,
      "counter": 501,
      "markerGetterCalls": {
        "admissionRecentCueScan": 16,
        "slotSelection": 0
      },
      "admissionCalls": 1,
      "listAllocations": [
        "one new LRU101 tailnode"
      ],
      "listFrees": [
        "old LRU101 node"
      ],
      "deviceOperations": [],
      "draws": [],
      "lastScalarTick": 100,
      "allOtherState": "unchanged",
      "listHeapOrder": [
        "free old LRU101 node",
        "allocate/append new LRU101 node"
      ]
    }
  },
  {
    "variation": "zero-position no-ref request is deferred despite Loop1",
    "outputs": {
      "I": {
        "inUse": 0,
        "state": "pending",
        "slot": null,
        "retirement": 1,
        "currentVolume": 255,
        "bookmark": 0,
        "playbackOffset": 0,
        "references": [],
        "position": [
          0,
          0,
          0
        ],
        "distanceKey": 0
      },
      "O": "all semantic voice fields unchanged; remains inUse1/state1/S0/refA",
      "activeVoices": [
        "O"
      ],
      "ownerAHandles": [
        500
      ],
      "S0Voice": "O",
      "S1Voice": null,
      "history100": {
        "recent": 7,
        "older": 6
      },
      "history101": {
        "recent": 0,
        "older": 0
      },
      "accessTicks": {
        "100": 90,
        "101": 100
      },
      "LRU": [
        100,
        101
      ],
      "usedBytes": 96,
      "pendingCount": 0,
      "counter": 501,
      "markerGetterCalls": {
        "admissionRecentCueScan": 0,
        "slotSelection": 0
      },
      "admissionCalls": 0,
      "listAllocations": [
        "one temporary deferrednode containingI",
        "one new LRU101 tailnode"
      ],
      "listFrees": [
        "old LRU101 node",
        "activeI node",
        "temporary deferredI node"
      ],
      "deviceOperations": [],
      "draws": [],
      "lastScalarTick": 100,
      "allOtherState": "unchanged",
      "listHeapOrder": [
        "allocate/prepend temporary deferredI node",
        "free old LRU101 node",
        "allocate/append new LRU101 node",
        "free activeI node",
        "free temporary deferredI node"
      ]
    }
  }
]
```

##### Discriminates

The exception needs BOTH no reference-list pointer and at least one nonzero coordinate. z320 alone qualifies although distanceKey remains 0. Positioned Loop1 request survives failed admission after history shift; zeroXYZ gets retirement 1 and isfreed before any admission/history write.

runtime_observed: no


#### STP-V57 — queued cancellation joined to notification delivery

Inputs: one queued Q1, stream list[S], current null, notification list empty. file F references 3 (external owner,S,Q1), direct handle Hfile/archive null/logical 54/streamFlag 1. stream S references 2 (list,Q1), pending-start/file F/borrowed buffer B/progress 4/tail 0/cache 4/write 16384; capacity 262144/data 44/10/average 88200, other semantic fields retained from REQUEST-RETENTION. B all 0xaa. Q1 source 48/dest 0/length 6/freeAfter 1/actual 0/cancellation identity 0/no completion callback; completion event Hnotify initially unsignaled. Cancel filters(identity 0,buffer B). Notification allocation of notification N1 and request free succeed; valid list providers. No worker runs during cancellation.

Seeds/nodraw: none. orderedDraws: [].

Trace: match buffer B; allocate/link N1 carrying Hnotify before request destruction; save successor; drop file F references 3→2 then stream S references 2→1; unlink/free Q1; return 1. No event signal yet. Carry ALL outputs into notification-only I/O worker invocation. Additional inputs: exit 0, bandwidth 0/counter 0, pump inhibition 1, wake event Eworker, named clocks(entry 1000,loop 1000,selector 1000); SetEvent(Hnotify) BOOL 1/signals event. Empty selector means Wait(Eworker,250); provider writes exit 1 immediately before returning 0, the only allowed interleaving. Worker drains N1 before pump/selection: signal, unlink/free N1 regardless signal return; inhibited pump changes no stream; empty queue causes no read; Wait; clear exit 0 and return 0.

ExactOutputs: cancellation checkpoint request Q1 freed/notification N1 live/event unsignaled/file F references 2 / stream S references 1; queue empty, S still listed/pending/cache 4/write 16384/progress 4/tail 0, file F stream flag 1/logical 54, B unchanged; no Stop/Release/CloseH. Joined suffix notification N1 freed/notifications empty/event signaled, one SetEvent, exit 0/inhibition 1/current null; file F references 2 / stream S references 1 and all producer state unchanged; worker return 0. No read/callback/device calls.

Additional exact cancellation projection=[allocate N1 carrying Hnotify, enqueue N1, Q1 drops F3→2, Q1 drops S2→1, unlink/free Q1]. The null-completion-event variation omits the first two events. For the joined main and failed-signal suffixes, the exact external/free projection is [SetEvent(Hnotify) with the declared Boolean, unlink/free N1, Wait(Eworker,250), clear exit 0/return 0]. The observable uninhibited-pump variation below distinguishes pump order; the inhibited main suffix alone makes no such discrimination.

Discriminates: retained notification is separate from request and actual successful signal; consumer drains before pumping. runtime_observed: no.

Independent variations: null Q completion event allocates no N1, frees Q with same refs, no event signal. Signal failure resets original cancellation input then carries its result; suffix SetEvent BOOL 0 leaves event unsignaled but consumes/frees N1, no retry/replacement; all other suffix outputs equal. The worker-not-running case has no eventual delivery guarantee.


##### Independent observable notification-before-pump variation

This independently reset prefix starts at I/O-worker entry and pauses after its first pump return, before selection/read, with outer stream synchronization still held. There is no worker-return assertion. It consumes the creation-only STP-V36 checkpoint and the following inputs; the joined cancellation output above is not reset into it.

###### Inputs

Carry the complete creation-only checkpoint of the borrowed looping-stream example. No initial pump, selection, read or callback has yet occurred. Concrete controlling state:

- Wrapper W has valid borrowed inner buffer B, marker 1, stream owner F, end marker−1; capacity 262144, half 131072, logical loop field 0, no interfaces, fade inactive. Other wrapper fields retain that creation checkpoint. B contains0xaa in every byte and has not received PCM writes.
- Sole listed stream S is initial, loop 1, file F, buffer B, ownsBuffer0, references 1 (stream-list owner), dataStart44/dataLength10, startOffset4/cachedPosition4/progress 4, silenceTail0, writeCursor0, capacity 262144, averageBytesPerSecond88200, silenceByte0. Its raw volume/pan inputs remain 2147483647. No other stream exists.
- File F has direct handle H, archive null, streamFlag1, logicalPosition777, references 2 (external owner plus S). Read/check controls remain cleared from creation; optional owned read buffers null. The source is the same valid RIFF/WAVE file, data bytes[1..10] at 44, source length 54, stereo16-bit22050Hz/average 88200/block 4. OS position remains 44 after creation parsing. No file operation occurs in this prefix.
- Active/configured chunk 16384, source-block alignment mode 0, pump inhibition 0. Scratch allocation from creation remains live; every byte retains its supplied allocation value because no read or callback occurs. Existing worker handle/wake event identities remain valid. Wake event Ewake is supplied signaled at this checkpoint, consistent with successful creation wake and no intervening wait; its state is unchanged in this prefix.
- Request queue empty, current request null. Install exactly one already-live notification node Nsig in the completion-notification list, carrying valid event Efail initially unsignaled. Nsig allocation is part of the input checkpoint, not an operation of this prefix. No other notification exists.
- Worker exit input 0, bandwidth 0, throttle numerator 0. Every reached clock returns 1000: worker entry clock, first loop clock and Q1 request-construction clock. There is no selector clock call inside this stage. Worker prior archive/cache identity initializesnull/0, and last dispatch time initializes 1000; no dispatch occurs.
- Returning reentrant stream synchronization. Freeing Nsig returns normally. The only requested allocation is a valid fresh zero-initialized request Q1; allocation/list insertion returns normally. No request service, external writer or other interleaving occurs. No wait, device method, source read/seek/size, CloseHandle, AddRef or Release provider is reached.
- SetEvent(Efail) returns BOOL0, leaves Efail unsignaled, and records a read-only call-time snapshot of the engine state specified below. It makes no queue/ref/object mutation. No SetEvent(Ewake) is reached: the pump supplies request-constructor wakeCondition0.

Seeds/no draw: no seed is read or changed. orderedDraws:[].

###### Ordered trace

1. Worker entry samples 1000, sees exit 0, initializes its previous source state and dispatch-time record. First loop samples 1000 and computes throttle 0 from bandwidth 0 without a division.
2. Enter the outer stream lock. Visit Nsig. Request SetEvent(Efail); record the supplied snapshot BEFORE returning BOOL0.
3. Unlink/free Nsig despite the failed signal. Completion-notification list becomes empty. No retry or replacement node is produced.
4. Only now invoke the enabled stream pump. For initial S, total=min(262144,10−4)=6; write F logical position 44+4=48. No source-block alignment branch runs.
5. Construct Q1(source 48,destination B at 0,length 6,delay 0,priority 0,stream S,freeAfter1). Enter nested stream synchronization; allocate/link Q1 at queue tail; construction clock 1000 gives deadline 1000; direct-handle cache key 48. Raw buffer, cancellation cookie, completion event and completion-metadata reference are all null. Increment file F references 2→3, stream S references 1→2; leave the nested lock. WakeCondition0 bypasses the constructor's SetEvent branch.
6. Pump advances F logical position 48→54 by requested 6 and scheduled ring destination 0→16384 by one whole chunk, finally storing writeCursor16384 and state pending-start. Progress/cache/tail stay 4/4/0. The pump returns.
7. Pause BEFORE selector invocation. There is no read, callback, request destruction, outer lock leave, wait, exit-tail check or normal worker return in this bounded prefix.

###### Exact outputs

At the sole SetEvent call, the recorded snapshot is:

| Observable at call | Required value |
|---|---|
| Request queue/current | [] / null |
| File F references/logical position/streamFlag |2 /777 /1 |
| Stream S references/state/writeCursor |1 /initial /0 |
| Stream S cachedPosition/progress/silenceTail |4 /4 /0 |
| Notification list/Nsig |[Nsig] /live |
| Efail |unsignaled |

At the post-pump pause:

- Request queue[Q1], current null, completion notifications[], Nsig freed. Stream list[S] remains.
- File F references 3/logicalPosition54/streamFlag1; external owner remains live. Stream S references 2/state pending-start/writeCursor16384; cache 4/progress 4/tail 0/start 4/data 44/10/loop 1/borrowed ownership and every other stream field unchanged.
- Q1 is live and queued: file F, source 48, cache key 48, deadline 1000, priority 0, raw buffer null, cookie 0, device destination B, destination offset 0, stream S, requested length 6, actual count 0, freeAfter1, completion event null, completion metadata reference null. This is one request appended to an initially empty queue.
- W marker 1/stream F/inner B and all wrapper fields remain unchanged. B/source/scratch bytes remain unchanged. F OS position 44 remains unchanged. Efail remains unsignaled; Ewake remains in its carried signaled state.
- Ordered allocation/free log: free Nsig, then allocate Q1. No Nsig allocation occurs within the stage. Sole event operation is SetEvent(Efail) returning 0; no request-constructor wake operation, callback, device call, source I/O, wait, request free or source/stream destruction occurs.
- Worker exit input remains 0. Previous archive/cache remainsnull/0 and last dispatch time 1000; no selector result has been produced and current request remainsnull. No value is asserted for selector-output scratch before that selector runs. Outer stream synchronization is still held once; the nested constructor entered/left once within it. No worker-return value is asserted at an internal pause.

###### Discriminates

A pump-first implementation would have queued Q1 and raised references to file 3/stream 2 before the event call, with file logical 54, state pending-start and writeCursor16384. It cannot produce the required call-time snapshot. Merely checking final notification consumption would miss that ordering error; this snapshot makes notification-before-pump observable even though SetEvent fails. The failed signal is still consumed before Q1 allocation, with no retry.

runtime_observed:no.



#### STP-V58 — final owned versus borrowed stream release

Inputs: queue/current/notifications empty; sole stream S references 1/file F/buffer Bowned/owns 1/playing/cache 4/write 16384/progress 16388/tail 16378; file F references 2 (external owner,S), direct handle Hfile/archive null/logical 54/flag 1. External owner staysalive. Release returns 0; stream free succeeds. No Sound caller runs.

Seeds/nodraw: none. orderedDraws: [].

Trace: end F; no current destination; empty cancellation returns 0; unlink S; stream S references 1→0; final destructor file F references 2→1; owned nonnull buffer Release once; free S; after traversal file F stream flag 0; end returns 1.

ExactOutputs: file F live / references 1/logical 54/flag 0; stream S freed; lists/current empty; one Release, no Stop/CloseH/notification/event call; external owner retained. Release 7 variation still frees S/returns 1/file F references 1/flag 0, physical buffer lifetime unspecified. Borrowed owns 0 variation makes zero Release and retains buffer, otherwise same engine outputs.

Additional exact ordered output=[unlink S, S.references1→0, F.references2→1, Bowned.Release with the supplied result, free S, F.streamFlag0, return 1]. The borrowed variation omits Release and preserves B; the Release7 variation substitutes 7 without changing event order or engine liveness.

Discriminates: file-reference drop precedes Release and stream free; engine ownership differs from physical lifetime; no implicit owner close/Stop. runtime_observed: no.


#### STP-V59 — initial request scheduling stage

Inputs: single initial S/file F/dataStart 44/dataLength 40000/start 0/cache 0/progress 0/tail 0/write 0/capacity 65536/chunk 16384/average 88200/buffer B/owns 0/references 1; file F logical position 123/references 2/directH plus valid archive metadataA with sector size 4096; pump inhibition 0. Source alignment mode is the variation. No worker runs. This stage consumes STP-R04b request construction with supplied fresh valid identities Q1,Q2,Q3 and construction clocks[1000,1000,1000]. The valid queue begins empty; each construction appends at its tail and increments F/S references once. No worker or other queue mutation interleaves. Encoded node links and generic allocator internals remain provider-owned. Priority 0/freeAfter 1/no raw buffer/cancellation identity/event/callback/wake. Cache offset equals source offset because directH exists. Device calls and I/O absent.

Seeds/nodraw: none. orderedDraws: [].

Trace: compute total 40000. Mode 0 runs ordinary requests, adding whole chunks to destination and delay 185 after each. Mode 1 first trims min(chunk,total) by nonzero smaller end-sector remainder 44, queues special 16340 at dest 0/delay 0, then ordinary requests from dest 16340 with delay still 0. Final ring cursor is scheduled destination modulo capacity, state pending-start.

ExactOutputs, requests in emission order (source,destination,length,delay,deadline): mode 0[(44,0,16384,0,1000),(16428,16384,16384,185,1185),(32812,32768,7232,370,1370)], final file F logical position 40044/write 49152. Mode 1[(44,0,16340,0,1000),(16384,16340,16384,0,1000),(32768,32724,7276,185,1185)], final file F logical position 40044/write 49108. Both file F references 5 / stream S references 4/stream S state pending-start; all other S fields unchanged; no reads/destruction/device calls. The dependency's pending request records are live; their semantic traversal order is exactly[Q1,Q2,Q3] in both variants. No encoded node layout is asserted.

Discriminates: special aligned first advances by actual trimmed length without delay increment; later initial requests advance whole chunk. This is the smaller supplied scheduling stage, no whole-worker/allocator credit. runtime_observed: no.


#### STP-V60 — post-read zero-count completion segment

Inputs: REQUEST-RETENTION, but actual read count 0 already stored; entry is immediately before device-destination dispatch, ending after request destruction/current clear/lock leave, before any next worker loop. Q completion callback/event null/freeAfter 1, device destination buffer B, requested 6. Stop −1; no further read occurs inside this segment. Scratch/B unchanged within segment. Valid free/list/lock providers.

Seeds/nodraw: none. orderedDraws: [].

Trace: nonnull destination+actual 0+requested 6 calls Stop ignored, writes stream S state finished; enter lock; drop request-held file F references 3→2 then stream S references 2→1; unlink/free request Q; clear current request last; leave.

ExactOutputs: stream S live / listed / state finished / references 1, file F live / references 2/flag 1, wrapper W marker 1/stream F; request Q freed/current null; one Stop, no callback/Lock/Unlock/Play/CloseH/Release; B and scratch unchanged in this post-read segment. No inference about the preceding OS read BOOL or error handling.

Discriminates: zero actual read is not a padding callback. runtime_observed: no.

Independent destination-cleared variation starts from STP-V46's post-Sound-stop checkpoint, with actual 0 instead of 6: Q destination null, file F references 2 / stream S references 1, S unlinked/still pending, file F stream flag 0, wrapper W marker 0/stream null, borrowed buffer B. Supply CloseH BOOL 1/free success. Skip callback/Stop/state write; drop file F references 2→1 then stream S references 1→0→file F references 1→0/CloseH/free file F, skip Release, free stream S/Q, clear current request. No device call in this segment; prior Sound Stop remains outside it. Requested 0 variation instead uses base file F references 3 / stream S references 2/destination buffer B with actual 0/requested 0: skip Stop/state write; request destruction yields file F references 2 / stream S references 1/stream S state pending-start/wrapper W marker 1/stream F. It is an explicit internal guard checkpoint, not a natural initial-pump zero-request claim.


#### STP-V61 — Suppression chooses the earliest start, then tests retirement

##### Inputs

```json
{
  "entry": "complete admission entry and return only",
  "overrides": {
    "controls.markedSlotSearchAndRoutingMixEnabled": 1,
    "voicePool.activeOrder": [
      "I",
      "O",
      "J"
    ],
    "owners.A.handles": [
      501,
      500,
      499
    ],
    "voicePool.entries.freeIndices": "3..199",
    "voicePool.entries.I.start": 101,
    "voicePool.entries.I.requestFlags": 1,
    "voicePool.entries.I.priority": 100,
    "voicePool.entries.O.cue": 100,
    "voicePool.entries.O.start": 100,
    "voicePool.entries.O.priority": 100,
    "voicePool.entries.O.retirement": 0,
    "voicePool.entries.J": {
      "index": 2,
      "inUse": 1,
      "cue": 100,
      "handle": 499,
      "distanceKey": 0,
      "currentVolume": 255,
      "contextFactor": 0,
      "playbackOffset": 0,
      "references": [
        "A"
      ],
      "slot": "S1",
      "state": "playing",
      "start": 90,
      "requestFlags": 1,
      "priority": 100,
      "retirement": 0,
      "bookmark": 0,
      "activeTransition": 0,
      "startVolume": 0,
      "endVolume": 0,
      "startTick": 0,
      "endTick": 0,
      "position": [
        0,
        0,
        320
      ]
    },
    "slots.allClasses": 0,
    "slots.S1.voice": "J",
    "wrappers.W0.data": "WAV100.PCM",
    "wrappers.W1": {
      "inner": "B1",
      "bufferCapacity": 262144,
      "halfCapacity": 131072,
      "threeDInterface": null,
      "propertyInterface": null,
      "bufferFlags": 1,
      "requestedVolume": 255,
      "propertyScalar": 0,
      "streamHandle": null,
      "logicalLoop": 0,
      "loopStart": 0,
      "backendFadeActive": 0,
      "backendFadeStartVolume": 0,
      "backendFadeEndVolume": 0,
      "backendFadeStartMs": 0,
      "backendFadeEndMs": 0,
      "marker": 1,
      "data": "WAV100.PCM",
      "dataLength": 4,
      "playedBytes": 0,
      "writtenBytes": 4,
      "halfFlag": 1,
      "endDataFlag": 1,
      "endMarker": 4,
      "payload": [
        1,
        2,
        3,
        4,
        "zero repeated 262140 times"
      ]
    },
    "records.100.Stereo": 1,
    "records.100.Stream": 0,
    "records.100.state": "loaded",
    "records.100.failure": 0
  },
  "slotCount": 16,
  "slotInputs": [
    {
      "index": 0,
      "wrapper": "W0",
      "buffer": "B0",
      "class": 0,
      "voice": "O",
      "marker": 1
    },
    {
      "index": 1,
      "wrapper": "W1",
      "buffer": "B1",
      "class": 0,
      "voice": "J",
      "marker": 1
    },
    {
      "index": 2,
      "wrapper": "W2",
      "buffer": "B2",
      "class": 0,
      "voice": null,
      "marker": 0
    },
    {
      "index": 3,
      "wrapper": "W3",
      "buffer": "B3",
      "class": 0,
      "voice": null,
      "marker": 0
    },
    {
      "index": 4,
      "wrapper": "W4",
      "buffer": "B4",
      "class": 0,
      "voice": null,
      "marker": 0
    },
    {
      "index": 5,
      "wrapper": "W5",
      "buffer": "B5",
      "class": 0,
      "voice": null,
      "marker": 0
    },
    {
      "index": 6,
      "wrapper": "W6",
      "buffer": "B6",
      "class": 0,
      "voice": null,
      "marker": 0
    },
    {
      "index": 7,
      "wrapper": "W7",
      "buffer": "B7",
      "class": 0,
      "voice": null,
      "marker": 0
    },
    {
      "index": 8,
      "wrapper": "W8",
      "buffer": "B8",
      "class": 0,
      "voice": null,
      "marker": 0
    },
    {
      "index": 9,
      "wrapper": "W9",
      "buffer": "B9",
      "class": 0,
      "voice": null,
      "marker": 0
    },
    {
      "index": 10,
      "wrapper": "W10",
      "buffer": "B10",
      "class": 0,
      "voice": null,
      "marker": 0
    },
    {
      "index": 11,
      "wrapper": "W11",
      "buffer": "B11",
      "class": 0,
      "voice": null,
      "marker": 0
    },
    {
      "index": 12,
      "wrapper": "W12",
      "buffer": "B12",
      "class": 0,
      "voice": null,
      "marker": 0
    },
    {
      "index": 13,
      "wrapper": "W13",
      "buffer": "B13",
      "class": 0,
      "voice": null,
      "marker": 0
    },
    {
      "index": 14,
      "wrapper": "W14",
      "buffer": "B14",
      "class": 0,
      "voice": null,
      "marker": 0
    },
    {
      "index": 15,
      "wrapper": "W15",
      "buffer": "B15",
      "class": 0,
      "voice": null,
      "marker": 0
    }
  ],
  "control": {
    "incoming": "I",
    "cue": 100,
    "start": 101,
    "priority": 100,
    "selectionBypass": true,
    "markedSlotSearchEnabled": true,
    "asset": "WAV100 loaded,rawSize48/PCMSize4/failure0/Stereo1/Stream0",
    "history": [
      7,
      6
    ]
  },
  "independentVariation": {
    "name": "oldest retiring; no fallback",
    "overrides": {
      "voicePool.entries.O.start": 100,
      "voicePool.entries.O.retirement": 1,
      "voicePool.entries.J.start": 101,
      "voicePool.entries.J.retirement": 0
    }
  }
}
```

##### Seeds

requestSelection and ownerA are aliases for the same owners.A.seed state, entering(1,0) and remaining unchanged. They are not independent seeds. No other seed is consumed. For a fatal case this is the pre-termination seed state.

##### orderedDraws

`[]` — no RNG draw occurs.

##### Trace

```json
[
  {
    "variation": "base",
    "actions": [
      "Retain entering record100; bypass group selection. Read all16 markers in slot order.",
      "S0 matching Ostart100 is first candidate; S1 matching Jstart90 is unsigned smaller and replaces it. Retain first only on equal starts.",
      "Selected J is not retiring. Unsigned32(101−90)=11, so no early suppression. Commit history recent100/older7.",
      "Loaded buffered/Stereo1 requests class1 without routing-mode queries. All classes0: both free-slot and eviction scans return none.",
      "Return no slot before prepare/load/association/cache/property/Play. No other input state changes."
    ]
  },
  {
    "variation": "oldest retiring",
    "actions": [
      "Restart entering state with Ostart100/retirement1 and Jstart101/retirement0. Complete scan chooses O.",
      "Test selected O retirement only now: skip suppression; do not filter first or fall back to J.",
      "Commit recent100/older7 and return no slot after incompatible class search, with no further writes."
    ]
  }
]
```

##### exactOutputs

```json
{
  "base": {
    "selectedOldSlot": "S1",
    "selectedOldVoice": "J",
    "returnSlot": null,
    "history100": {
      "recent": 100,
      "older": 7
    },
    "voicePool.entries": "Every incoming/O/J semantic field unchanged from input",
    "activeOrder": [
      "I",
      "O",
      "J"
    ],
    "ownerAHandles": [
      501,
      500,
      499
    ],
    "slotAssociations": {
      "S0": "O",
      "S1": "J",
      "S2to15": null
    },
    "markers": {
      "W0": 1,
      "W1": 1,
      "W2to15": 0
    },
    "history101": {
      "recent": 0,
      "older": 0
    },
    "LRU": [
      100,
      101
    ],
    "accessTicks": {
      "100": 90,
      "101": 90
    },
    "usedBytes": 96,
    "pendingCount": 0,
    "nextHandleCounter": 501,
    "markerGetterOrder": [
      "W0",
      "W1",
      "W2",
      "W3",
      "W4",
      "W5",
      "W6",
      "W7",
      "W8",
      "W9",
      "W10",
      "W11",
      "W12",
      "W13",
      "W14",
      "W15"
    ],
    "routingModeGetterCalls": 0,
    "prepareCalls": 0,
    "loadCalls": 0,
    "startCalls": 0,
    "deviceOperations": [],
    "listAllocations": [],
    "listFrees": [],
    "allOtherState": "unchanged"
  },
  "oldestRetiring": {
    "selectedOldSlot": "S0",
    "selectedOldVoice": "O",
    "returnSlot": null,
    "history100": {
      "recent": 100,
      "older": 7
    },
    "allOtherState": "Same output assertions as base, relative to variation input; retain O.retirement=true and J.start101."
  }
}
```

##### Discriminates

Both actual cases return no slot but commit history 100/7. Incorrect first-match selection in the base would choose start 100, suppress at elapsed 1 and leave history 7/6. Incorrect prefiltering of retirement in the variation would choose start 101, suppress at elapsed 0 and likewise leave 7/6. The actual search chooses before retirement testing, with no fallback.

runtime_observed: no

#### STP-V62 — creation, factory pointer policy, and owned end

Inputs: STREAM-CREATION, with mono 8-bit PCM/channels 1/rate 11025/average 11025/block 1/bits 8; source and chunk/data lengths otherwise unchanged. F owner 1/logical 777/read-enable 99/88/status 7. Requested capacity 20000, creator flags 0x40060, start 23, raw volume 100/pan −200, supplied buffer null. Configured chunk 0/active 77/scratch absent; valid device; direct permission 1; existing worker T/no interleaving. File reads success/full 12,8,16,8; size 54/tracked seeks; allocations scratch and zeroed S succeed; CreateSoundBuffer returns HRESULT 0/output Bnew; descriptor flags 0xe0/capacity 32768/reserved 0/format mono 8-bit with extra size 0, no outer unknown. SetVolume/SetPan/SetCurrentPosition HRESULT −1; wake BOOL 0; valid interfaces/free/list providers. No event/thread creation because T exists.

Seeds/nodraw: none. orderedDraws: [].

Trace: default chunk 16384/allocateScratch; normalize capacity 32768; clear read-enable fields; seek/read RIFF,fmt,data; nonzero offset 23 modulo 10→3; create buffer; raw SetVolume 100 then SetPan −200; allocate/link stream S/references 1, file F references 1→2; SetCurrentPosition 0 ignored; file F stream flag 1; bootstrap skips creation; wake ignored; return 1.

ExactOutputs: configured/active chunk 16384, scratch retained; OS position 44/file F logical position 777/read-enable 0/0/status 7; file F references 2/flag 1; stream S state initial/file F/start/cache/progress 3/tail 0/write 0/data 44/10/capacity 32768/average 11025/loop 1/raw volume 100/pan −200/Bnew/owns 1/silence 128/references 1. Ordered device log CreateSoundBuffer,SetVolume(100),SetPan(−200),SetCurrentPosition(0); no AddRef/Release/Stop/Play/CloseH. Wake once despite BOOL 0. Return 1 does not prove worker created/running or audible output.

Discriminates: mask capacity normalization, modulo start, silence 8-bit, raw creation controls, ownership versus factory result sign. runtime_observed: no.

Independent variations: factory HRESULT −1/output Bnew has identical engine outputs; factory HRESULT 0/output null returns 0 before S/reference increment/controls/position/wake, file F external-owner references 1/flag 0, OS 44/logical 777/read-enable 0/0, scratch retained; no new explicit parse error assignment. Malformed RIFF tag instead returns 0/error 0x0d after read 12, OS position 12/logical 777/read-enable 0/0/scratch retained; no factory/S/reference increment/ownerclose. A supplied valid borrowed buffer Bborrow with capacity at least 32768 and format tag 85 bypasses factory/creation controls/AddRef, succeeds under providers/owns 0/silence 128/SetCurrentPosition once; no actual codec support claim. Owned-end suffix carries the successful creation state: end drops stream S references 1→0, file F references 2→1, Release Bnew once, free stream S, clear file F stream flag 0, return 1; file F owner retained, no Stop/CloseH, scratch retained. Release return 0 is supplied, physical lifetime external.

#### STP-V63 — Dispose a pending asset while its file read is in flight

**Inputs**

- Enter pending-asset disposal for cue 100. Valid sound records exist for cues 0..2933; records are zero except the stated cue 100 values. Cue100 has pending asset state, no failure, no raw payload reference, retained payload size 48, no parsed PCM reference, parsed PCM length 0, last access 90, descriptor D, reference count 0 and caching disabled. Active voice list is empty. Marked-slot search is enabled 1. All16 valid playback slots have no voice and stored start marker 0. LRU is[100], accounted bytes 48 and client pending count 1.
- Descriptor D is valid, not ready, owns its payload, and is not linked in the optional descriptor list. It owns file F and payloadRaw48 of size 48, uses default resource pool 0 and read priority−2, and has no callback or replacement payload metadata. Its event slot isE. Its read-overlap state is zero except for the event referenceH. The descriptor list is empty.
- Event slot E is live, contains valid manual-reset event H, is not reusable and identifies D as its owner. H is unsignaled. The event pool and handle will remain available after this case; disposal does not request their destruction.
- File F is a valid direct file with OS handle Hfile, path “fixture.bin”, no archive, no stream association and no auxiliary file buffers. Its reference count 2 consists of descriptor D's ownership and current request Q's ownership. Its logical position is 0. The file list contains only F with valid links.
- Q is the current selected request, already removed from the queue with cleared links. It references F, writes to Raw48 from file offset 0 for 48 bytes, uses Raw48 as its cancellation cookie, notifies H, requests destruction after completion and has actual-read count 0. It has no device buffer, stream, destination offset or callback metadata. Request queue, stream list and pending-notification list are empty.
- The IO worker is paused inside its direct-file read helper: seek to offset 0 has already returned 0; ReadFile(Hfile,Raw48,48) is in flight and has not written data. The helper's saved position is−1 and actual-read count 0. Read-error policy is 0. Worker exit flag is 0. There is no other worker progress until the explicitly supplied infinite-wait interleaving.
- Provider sequence: Wait(H,0) returns 258; Wait(H,infinite), represented numerically as0xffffffff, returns 0 only after the specified worker completion. ReadFile then returns Boolean1 and actual count 48, writing byte values 0..47 into still-live Raw48. At that read return, the controlled environment sets worker exit flag 1. SetEvent(H) returns 1 and signals H; ResetEvent(H) returns 1 and leaves H unsignaled; CloseHandle(Hfile) returns 1. All critical-section operations and listed generic heap frees return normally. No other interleaving or external operation occurs.

**Seeds / no draw**

No RNG state is consumed or changed.

**Draws**

None.

**Trace**

1. Client disposal finds no marked slot. Accounted bytes become 0. Remove cue 100 from LRU, free its LRU node and clear its access tick to 0.
2. Descriptor destruction validates D. Its linked flag is clear, so no descriptor-list removal occurs. Poll sees not-ready; Wait(H,0) returns 258, leaving D not-ready.
3. Cancellation searches queued requests for cookieRaw48. The queue is empty, so it matches none. It leaves currentQ and its Raw48 destination intact and creates no pending-notification node.
4. Descriptor destruction blocks in Wait(H,infinite). Raw48,D,F and Q all remain live.
5. Resume the supplied ReadFile. It writes all 48 bytes, reports actual 48 and returns 1. The helper records saved position 48 and actual 48; the worker stores Q's actual-read count 48. Apply the controlled worker-exit flag write 1.
6. Q has no stream or callback metadata, so perform neither callback. Save its notification H and enter the request critical section.
7. Destroy Q: drop its file reference, F2→1; no stream reference exists. Its list links are already clear. Free Q, clear the current-request reference and leave the critical section.
8. Call SetEvent(H), result 1, making H signaled. Before the waiting client resumes, let the worker tail see exit flag 1, clear it to 0 and return 0. It performs no further request selection or stream pumping.
9. The client's infinite wait now returns 0. Descriptor destruction proceeds to finalization. This destructor does not test the wait result; the supplied successful result and ordering establish completion for this case.
10. ResetEvent(H), result 1. Mark E reusable, clear E's descriptor owner and clear D's event-slot reference. Keep E/H live; do not close H.
11. Drop D's file ownership: unlink F, decrement F1→0, close Hfile and free F. Clear D's file reference. No archive or auxiliary-buffer destruction occurs.
12. Set D ready. There is no replacement payload, so retain Raw48 until the next step.
13. Because D owns Raw48, free Raw48, clear its payload reference and remove its valid-signature marker while retaining the ready/owned flags and clear linked flag. Free D.
14. Client disposal clears cue 100's descriptor/raw references and sets asset state to empty. Retain size 48, parsed reference null, parsed length 0 and failure 0. Pending count remains 1.

**Exact outputs**

- Cue100: empty asset state; descriptor/raw/parsed references null; retained size 48; parsed length 0; access tick 0; failure 0. LRU[], accounted bytes 0, pending count 1. Voices and slots remain unchanged.
- Q,F,Raw48 and D are freed. Raw48 receives exactly 48 supplied bytes before it is freed. Ordered heap frees: LRU node 100,Q,F,Raw48,D.
- Immediately before descriptor free: file, payload and event-slot references are null; valid-signature marker removed; ready and owned flags retained; linked flag clear. No freed object's contents are used as subsequent live state.
- E remains live and reusable, contains H and has no descriptor owner. H remains live and unsignaled. Exactly one event signal, one reset, one file-handle close and zero event-handle closes occur.
- Current request null; request/stream/pending-notification lists empty. F's ownership drops 2→1 at request destruction, then 1→0 at descriptor finalization. No stream references are involved.
- Worker return 0; exit flag 0. Zero cancellation matches, notification-node allocations, callbacks, RNG draws and device operations. No further reads occur. All other live state is unchanged.

Additional exact ordered output projection=[free LRU100 node, Wait(H,0)→258, queued-cookie cancel (zero matches), begin Wait(H,infinite), ReadFile writes/returns 48, Q drops F2→1, free Q, current=null, SetEvent(H)→1, worker clears exit 0/returns 0, Wait(H,infinite)→0, ResetEvent(H)→1, E.reusable=true/E.owner=null/D.eventSlot=null, D drops F1→0, CloseHandle(Hfile)→1, free F, D.file=null, D.ready=true, free Raw48, D.payload=null/checked marker invalid, free D, client descriptor/raw=null/asset unloaded]. These events assert request destruction before signaling and finalization after the supplied wait return; the controlled environment retains the stated worker-tail/client-resume order.

**Discriminates**

Queued cancellation does not abort an already selected read. Disposal waits before freeing that read's destination. Request completion drops its file reference before notification; descriptor finalization then releases the last file reference. The event is reset and pooled, rather than closed. Client pending count does not become 0 merely because a pending descriptor was disposed.

**runtime_observed: no**

#### STP-V64 — Periodic cue draw remains consumed after request rejection

**Inputs**

- Enter the full audio coordinator. Audio enabled 1; periodic random-cue producer enabled 1. Supplied audio counter 3 and coordinator tick argument 3 are distinct inputs with the same value here. No dispatcher increment or other worker interleaving occurs.
- Selected client unitA is valid and has selection seed(lo 1,hi 0). No initial seed is inferred from a player identity or creation default.
- Valid decoded table storage contains 4657 records, indexed 0..4656. Every supplied record field is zero, including Volume0. These are synthetic fixture values, not claims about original game table cells.
- All200 logical voices are free with cleared state. Ordered active list is empty; next-handle counter 501. Tracking is disabled. Designated inclusive music-cue interval is 2934..4656.
- Modal mode 2, both other modal inputs 0, modal gain 100, interrupt gain 100, last gain tick 3.
- The backend's actual stored master-getter value is 100 and actual stored music-getter value 100. These stored inputs are read through the original getter behavior; they are not an unread external getter provider. No device, file, allocation or other external operation is supplied or needed.

A read-only observer records the logical operation/read order and seed value at the first master-setting read. It changes no engine state, getter result or RNG behavior.

**Seeds**

A selection seed starts(lo 1,hi 0). The producer consumes NRF-R01/NRF-R02 for its reached positive-modulus draw.

**Draws**

One Rand(A selection seed,1723), in this order:

1. Widened step:1×1791398085+0=1791398085.
2. Store seed(lo 1791398085,hi 0).
3. 1723 is not a power of two; unsigned 1791398085 modulo 1723=154.

No further draw occurs.

**Trace**

1. Audio enable passes. Periodic producer enable passes, so run the owned producer before sorting, retirement or the master getter.
2. Unsigned audio counter 3 modulo 3=0. Draw154 as above and form requested cue 2934+154=3088.
3. Request logical cue 3088 with null emitter, delay 0, flags 0 and playback offset 0.
4. Audio enabled, positive cue and valid record checks pass. Cue3088 has Volume0, so request returns 0 before compound lookup, voice allocation, handle increment, position setup or owner association. The prior producer seed step remains consumed.
5. Active list is empty, so sorting, retirement and per-voice work have no voice effects. Read the stored master value 100 once. No active voice reaches the music getter.
6. Deferred-removal list is empty. Its clear operation returns without allocation or free.
7. Modal mode 2 selects false; there is no interrupting active cue. Coordinator tick 3 is no greater than last gain tick 3, so skip gain arithmetic and store last gain tick 3.

**Exact outputs**

- SeedA(lo 1791398085,hi 0); one draw 154; one logical request for cue 3088 returning 0.
- All200 voices remain free/cleared; active list[]; no associations created; handle counter 501.
- Every table field unchanged. No variant/history commit, asset load or device playback is reached.
- Modal gain 100, interrupt gain 100, last gain tick 3, audio counter 3.
- Master getter called once, returning 100. Music getter called zero times. Empty deferred-list clear called once.
- Allocation/free/device/file operation logs[]; all other live state unchanged.
- Exact ordered projection=[Rand(A,1723) returns 154 after seed update, request cue 3088 returns 0, first retained master-setting read returns 100]. At that read the observed A seed is(lo 1791398085,hi 0). The observer is read-only; no additional engine mutation is part of this output.

**Independent variation — counter 4**

Reset the entire input fixture, including seedA(lo 1,hi 0) and last gain tick 3. Change only audio counter to 4 and coordinator tick argument to 4.

Unsigned4 modulo 3=1, so producer makes no draw or logical request. The empty coordinator still reads master 100 and clears its empty deferred list. Tick4>last 3 runs gain arithmetic: false modal condition adds 5 to 100 and saturates at 100; false interrupt condition adds 2 to 100 and saturates at 100. Store last gain tick 4.

Exact variation outputs: seedA(lo 1,hi 0), draws[], logical requests 0, handle counter 501, all voices free, active list[], both gains 100, last gain tick 4 and audio counter 4. Master getter calls 1; music getter calls 0. Ordered projection=[first retained master-setting read returns 100], with A seed(lo 1,hi 0) at that read and no producer draw/request event. Allocation/free/device/file logs remain[]; all other input state unchanged.

**Independent crossed-clock variations**

These are direct coordinator fixtures with distinct supplied clock inputs; they do not assert that the ordinary dispatcher passes unequal values.

1. Reset the main input, then change only coordinator tick argument to 4. Keep audio counter 3 and entering last gain tick 3. Counter3 modulo 3 is 0, so draw/request/seed and ordered master-read observation exactly equal the main case. Tick4>last 3 runs the two false-condition gain increases, both saturating at 100, then stores last gain tick 4. Exact outputs equal the main outputs except coordinator argument 4/last gain tick 4; audio counter stays 3. First-master-read seed(1791398085,0), ordered projection[Rand→request 0→master 100], free voices/handle 501/unchanged records/empty allocation/device/file logs remain. Reading tick 4 for the modulo gate would incorrectly omit the draw.
2. Reset the counter 4 variation input, then change only coordinator tick argument to 3. Keep audio counter 4 and entering last gain tick 3. Counter4 modulo 3 is 1, so no producer draw/request. The coordinator still reads master 100 with seed(1,0). Tick3 equals last 3, so gains stay 100 without arithmetic and last gain tick stays 3. Exact outputs equal the counter 4 variation except argument 3/last gain tick 3; audio counter stays 4. Ordered projection[master 100] and all other declared no-operation/state outputs remain. Reading tick 3 or accepting either clock as divisible would incorrectly draw.

Both cases have the stated empty or one-entry draw logs, no additional RNG, no worker/dispatcher increment and runtime_observed:no.

**Independent unsigned full-width counter variation**

Reset the main input, then set only audio counter to 2147483649 (0x80000001); coordinator tick argument and entering last gain tick remain 3. Unsigned2147483649=715827883×3, so its remainder is 0. The producer makes the same single Rand1723 result 154, consumes seed(1,0)→(1791398085,0) and requests cue 3088 returning 0 before allocation. Exact outputs equal the main outputs except retained audio counter 2147483649. The ordered projection and first-master-read seed are identical to the main case; gains/last tick remain 100/100/3, handle 501, voices free and all other table/provider outputs unchanged. Signed32 interpretation gives−2147483647, whose remainder is nonzero; byte or16-bit narrowing gives 1. Those wrong gate interpretations omit the required draw and fail. No new clock or external operation is introduced, draw log[154],runtime_observed:no.

**Discriminates**

The periodic producer is an owned call with an unsigned counter-modulo gate. A rejected logical request does not undo its earlier random draw. Counter4 skips producer drawing while still reaching the coordinator's later gain stage; it is not an early return from the entire coordinator.

**runtime_observed: no**

#### STP-V65 — Prepare reuse and recreation

##### Inputs shared by every case

W is valid, with inner B, capacity 262144, half 131072, flags 1, stream null, property/spatial interfaces null; data pointer SourcePCM with source bytes [1..10], dataLength 10, played 11, written 12, endData 1, loop 1, loopStart 2, marker 7, halfFlag 1, endMarker 99. Backend fade active 1/from 100/target 0/start 1ms/end 9ms; propertyScalar 8, stored volume 100. Other semantic fields zero/null without zeroing named live identities. OldB aliases the inner buffer B; its payload is 0xaa repeated 262144. No other thread runs.

Device D and interfaces are valid. Global acoustic-conversion mode 0, volume scalar 255, factory extraFlag 0, spatial property capability 0; reuse is the case input. All lock operations return normally and are reentrant. Stop, all ordinary device controls, and old inner Release return 0; optional old property/spatial Release return 7. Factory returns HRESULT 0/output Bnew unless overridden, with descriptor flags 0xe0/capacity 262144/reserved 0, PCM tag 1/channels 2/sample 22050/average 88200/block 4/bits 16/extra size 0, no outer unknown. NewB aliases the factory output Bnew; its payload of 0xbb repeated 262144 is a supplied device result; Prepare makes no payload initialization claim. Floating point/conversion inputs are valid: property reset 0 and scalar 255 make volume 255 reach exact upper attenuation clamp 0. No window/iconic/status/wait/clock/random query is reached. No stream/file/Lock/Unlock/Play/AddRef/heap free operation occurs. SourcePCM is retained unchanged.

##### Reuse existing buffer

Inputs: shared fixture, reuse 1. Seeds/no draw: none. orderedDraws: [].

Trace: validate W/inner/stream null; enter Sound lock. Clear dataPointer/dataLength/played/written/endData/loop/loopStart. Request B position 0, direct pan 0 (flags 1 has no spatial routing bit). Store propertyScalar 0; null property interface means no external property call. Store volume 255; convert to attenuation 0; B.SetVolume(0). Leave Sound lock.

Exact Outputs: inner B/capacity 262144/half 131072/flags 1 retained. Data pointer null/length 0/played 0/written 0/endData 0/loop 0/loopStart 0. Marker 7/halfFlag 1/endMarker 99/fade active 1/from 100/target 0/start 1/end 9 all retained. propertyScalar 0/stored volume 255. Stream null/interfaces null, all other fields unchanged. Device log [position 0,pan 0,volume 0]; no Stop/Release/Create. B/source payload unchanged. Prepare returns normally; no scalar return is contracted.

Discriminates: reuse resets source/progress while retaining marker, half selector, end marker, and backend fade. Property scalar reset happens before volume conversion. runtime_observed: no.

##### Recreate stereo buffer

Inputs: shared fixture, reuse 0. Seeds/no draw: none. orderedDraws: [].

Trace: validate W/inner/stream null; enter Sound lock; save flags 1. Stop helper enters nested lock, calls B.Stop and clears marker, leaves nested lock. Release old inner B (optional interfaces null); clear all wrapper semantic fields. Factory CreateSoundBuffer returns nonnegative HRESULT/valid Bnew; flags 1 means no special interface queries; write capacity 262144/half 131072/flags 1/endMarker −1. Prepare ignores factory BOOL and continues position 0/pan 0/propertyScalar 0/stored volume 255→attenuation 0, then leaves lock.

Exact Outputs: inner Bnew/capacity 262144/half 131072/flags 1/endMarker −1/stored volume 255; data pointer null/length 0/played 0/written 0/endData 0/loop 0/loopStart 0/marker 0/halfFlag 0/propertyScalar 0/fade active 0/from 0/target 0/start 0/end 0; stream/interfaces null. Other semantic fields zero. Device log [B.Stop,B.Release,Create(desc 0xe0,262144,PCM 2ch22050/16),Bnew.position 0,Bnew.pan 0,Bnew.volume 0]. Old buffer physical lifetime is external; new payload remains provider 0xbb. SourcePCM unchanged. Normal return, scalar return unspecified. No AddRef, source free, or clean failure unwind is implied.

Discriminates: nonreuse destroys retained wrapper state and recreates a checked factory buffer. runtime_observed: no.

##### Optional-interface release ordering variation

Inputs: recreate case independently reset, but valid old property interface Pprop and old spatial interface Tspatial; no new interfaces queried because saved flags 1. This is a conditional supplied checkpoint, not natural reachability proof. Seeds/no draw: none. orderedDraws: [].

Trace: B.Stop; Pprop.Release; Tspatial.Release; B.Release; create Bnew; position 0/pan 0/volume 0. The optional Release returns 7 are ignored. Exact Outputs: recreated W exactly as above, interfaces null, Release calls 3, Stop 1/Create 1; source unchanged, no physical-lifetime claim. Exact ordered device output=[B.Stop→0,Pprop.Release→7,Tspatial.Release→7,B.Release→0,CreateSoundBuffer→Bnew/HRESULT0,Bnew.SetCurrentPosition(0)→0,Bnew.SetPan(0)→0,Bnew.SetVolume(0)→0]. Discriminates: property before spatial before inner. runtime_observed: no.

##### Negative checked-factory result with valid output variation

Inputs: recreate case independently reset; factory writes valid Bnew but HRESULT −1. All following controls return 0. Seeds/no draw: none. orderedDraws: [].

Trace: Stop/Release oldB, clear W; factory checks negative HRESULT and returns 0 before metadata finalization. Prepare ignores BOOL and uses valid Bnew for position 0/pan 0/property reset 0/volume 255→attenuation 0. Exact Outputs: inner Bnew/stored volume 255; capacity 0/half 0/flags 0/endMarker 0; all other semantic fields zero/null, including marker/halfFlag/fade. New payload provider 0xbb unchanged. Device log [Stop B,Release B,Create −1/Bnew,position 0,pan 0,volume 0]. Source unchanged. Normal return is conditional on valid supplied Bnew; future playback validity is not asserted.

Discriminates: this factory checks HRESULT rather than using the Storm creator's nonnull pointer gate; its caller ignores its BOOL. runtime_observed: no.

##### Limits

With null factory output, Prepare's next inner dereference is invalid; there is no invented safe cleanup/failure return. Nonnegative factory HRESULT also has no explicit output-pointer gate. Flags with spatial routing/capability queries require separate 3D/property providers and remain outside this case. No buffered payload load, normal caller path reachability, logical coordinator, real device success, or new verified subrule is established.

#### STP-V66 — normal streamed stop and attached-stream marker bridge

All unspecified semantic fields retain the complete declared producer checkpoint; no common wrapper, source, stream, queue or worker state is reapplied after that checkpoint. Freed objects have no final readable-field oracle. Last-live snapshots below describe trace observations before freeing. Negative device results do not prove physical stopping.

#### Inputs

The producer checkpoint is the successful nonloop prefix used by STP-V45: creation, initial pump, Q1 read/callback and Q1 destruction, stopping before the first EOF pump. Carry every producer-only field and provider-consumption position. The source is the complete54-byte PCM RIFF/WAVE file with44-byte canonical header: RIFF size 46; fmt length 16, PCM tag 1, stereo 2,22050Hz,88200 bytes/second, block alignment 4,16 bits/sample; data length 10 and bytes[1,2,3,4,5,6,7,8,9,10]. No extra chunks. Logical Client associations below are supplied separately; they do not reset this producer.

The following are all controlling values read by this continuation. Other producer values retain their exact prefix values, including opaque scratch bytes and worker registrations not read here. Both lock providers are valid, reentrant and return normally. Generic frees return normally. There are no external side effects other than the listed supplied operations, and no interleaving.

```json
{
  "entry": "normal logical voice stop, voice I",
  "client": {
    "voicePool": {
      "count": 200,
      "allocationOrder": "ascending index",
      "activeOrder": [
        "I"
      ],
      "nextHandleCounter": 501,
      "entries": {
        "I": {
          "index": 1,
          "inUse": true,
          "cue": 100,
          "handle": 501,
          "position": [
            0,
            0,
            320
          ],
          "distanceKey": 0,
          "currentVolume": 255,
          "contextFactor": 0,
          "playbackOffset": 0,
          "references": [
            "A"
          ],
          "slot": "S1",
          "state": "playing",
          "start": 100,
          "requestFlags": 1,
          "priority": 100,
          "retirement": false,
          "bookmark": 77,
          "activeTransition": false,
          "startVolume": 0,
          "endVolume": 0,
          "startTick": 0,
          "endTick": 0
        },
        "freeIndices": "0,2–199; all free-entry semantic fields zero/absent"
      }
    },
    "owners": {
      "A": {
        "type": 0,
        "seed": [
          1,
          0
        ],
        "handles": [
          501
        ]
      },
      "B": null
    },
    "slots": {
      "count": 16,
      "S1": {
        "wrapper": "W1",
        "class": 1,
        "voice": "I",
        "cachedPosition": [
          0,
          0,
          0
        ],
        "cachedBaseVolume": 0,
        "cachedFactor": 0
      },
      "S0": {
        "wrapper": "W0",
        "class": 0,
        "voice": null
      },
      "S2to15": "distinct valid wrappers/buffers; class1; no associated voice; all other semantic fields retained unchanged"
    },
    "record100": {
      "valid": true,
      "Loop": 0,
      "Stream": 1,
      "StopInst": 0,
      "Volume": 255,
      "Stereo": 1,
      "historyRecent": 7,
      "historyOlder": 6,
      "lastAccess": 90,
      "assetState": "loaded",
      "raw": "WAV100",
      "rawSize": 48,
      "PCM": "WAV100.PCM",
      "PCMSize": 4,
      "failure": 0,
      "descriptor": null,
      "remainingDecodedFields": "same complete common logical record100; not read by normal stop"
    },
    "record101": {
      "historyRecent": 0,
      "historyOlder": 0,
      "lastAccess": 90
    },
    "assets": {
      "usedBytes": 96,
      "pendingCount": 0,
      "LRU": [
        100,
        101
      ]
    },
    "clock": {
      "tickArgument": 100,
      "audioCounter": 100,
      "lastScalarTick": 100
    },
    "otherClientState": "exact shared logical fixture after these explicit association/entry changes; no controller is invoked"
  },
  "wrapper": {
    "identity": "W1",
    "inner": "B1",
    "bufferCapacity": 262144,
    "halfCapacity": 131072,
    "marker": 1,
    "streamHandle": "F",
    "endMarker": -1,
    "bufferFlags": 1,
    "requestedVolume": 255,
    "propertyScalar": 0,
    "logicalLoop": 0,
    "loopStart": 0,
    "data": null,
    "dataLength": 0,
    "playedBytes": 0,
    "writtenBytes": 0,
    "halfFlag": 0,
    "endDataFlag": 0,
    "threeDInterface": null,
    "propertyInterface": null,
    "backendFadeActive": 0,
    "backendFadeStartVolume": 0,
    "backendFadeEndVolume": 0,
    "backendFadeStartMs": 0,
    "backendFadeEndMs": 0
  },
  "buffer": {
    "identity": "B1",
    "ownership": "borrowed by S, retained inner buffer of W1",
    "payload": "[5,6,7,8,9,10] then16378 zero bytes, then245760 bytes170"
  },
  "stream": {
    "identity": "S",
    "file": "F",
    "buffer": "B1",
    "ownsBuffer": false,
    "state": "playing",
    "cachedPosition": 4,
    "progress": 16388,
    "silenceTail": 16378,
    "writeCursor": 16384,
    "capacity": 262144,
    "dataStart": 44,
    "dataLength": 10,
    "startOffset": 4,
    "loop": false,
    "averageBytesPerSecond": 88200,
    "rawVolumeInput": 2147483647,
    "rawPanInput": 2147483647,
    "silenceByte": 0,
    "references": 1,
    "referenceOwners": [
      "stream list"
    ]
  },
  "source": {
    "identity": "F",
    "handle": "H",
    "archive": null,
    "logicalPosition": 54,
    "references": 2,
    "streamFlag": true,
    "optionalOwnedReadBuffers": [],
    "otherState": "carried complete nonloop producer checkpoint",
    "referenceOwners": [
      "Sound source owner",
      "S"
    ]
  },
  "lists": {
    "streams": [
      "S"
    ],
    "sources": [
      "F"
    ],
    "requests": [],
    "currentRequest": null,
    "currentWorkerRequestContext": null,
    "completionNotifications": [],
    "scratch": "same live allocation and bytes as producer checkpoint"
  },
  "providers": {
    "GetCurrentPosition": {
      "HRESULT": 0,
      "play": 10,
      "write": "not consumed; any valid supplied output"
    },
    "Stop": [
      -1,
      -1
    ],
    "CloseHandle": {
      "handle": "H",
      "BOOL": 1
    },
    "interleaving": [],
    "otherDeviceOrFileOperations": []
  },
  "otherWrappers": {
    "W0": {
      "marker": 0
    },
    "W2to15": {
      "marker": 0
    },
    "remainingFields": "valid distinct shared logical wrappers; all their other entering fields are retained and not read in this case"
  },
  "streamGlobals": {
    "activeChunk": 16384,
    "pumpInhibition": 0,
    "scheduling": "serialized; no pump or worker invoked by this continuation"
  }
}
```

#### Seeds

The one owners.A.seed enters(1,0) and is unchanged. requestSelection is an alias for that same state, not an independent seed. No random seed is read or modified.

#### orderedDraws

`[]`

#### Trace

1. The logical stop sees I playing with live S1/W1/B1. It does not test the decoded Loop or wrapper marker before stopping.
2. Query the bookmark under Sound synchronization. The attached source selects the one playing stream under stream synchronization. B1.GetCurrentPosition returns success with physical play position 10.
3. Recompute cached position: progress 16388 modulo capacity 262144 is 16388; forward distance to the physical cursor is 16378; candidate progress−distance=10, clamped to data length 10. Cache changes 4→10. The query returns success with position 10 and source length 10. Leave stream synchronization, then Sound synchronization.
4. Convert the returned position by unsigned division by 4:10→2. Save I.bookmark=2 before beginning teardown. No second position query occurs.
5. Enter normal marker-stop Sound synchronization. Because the stream is attached, enter the nested attached-stream teardown. End the stream under stream synchronization. The nested cancellation scan finds the supplied empty request queue/current request, so there is no callback, request allocation or deferred request ownership.
6. Unlink S; references 1→0. Destruction drops the stream-owned F reference 2→1. Borrowed B1 is not released. Free S, clear F.streamFlag, and leave stream synchronization. The last live stream remains playing/cache 10 with its original progress/tail/write counters; no EOF/draining/finished transition is invented.
7. Call B1.Stop#1→−1 and ignore the result.
8. Close the remaining owner F reference under stream synchronization: unlink F, references 1→0; CloseHandle(H)→1; optional buffers/archive are absent; free F. Leave stream synchronization.
9. Clear W1.streamHandle and W1.marker; leave nested Sound synchronization.
10. The normal stop layer unconditionally calls B1.Stop#2→−1, ignores the result, writes marker 0 and leaves outer Sound synchronization.
11. Set I stopped, clear I.slot and S1.voice, and return normally. Retain I.inUse, retirement, active-list membership and A references/handle. No owner cleanup is performed.

#### exactOutputs

```json
{
  "normalReturn": true,
  "I": {
    "inUse": true,
    "cue": 100,
    "handle": 501,
    "state": "stopped",
    "slot": null,
    "bookmark": 2,
    "references": [
      "A"
    ],
    "currentVolume": 255,
    "playbackOffset": 0,
    "retirement": false,
    "activeTransition": false,
    "allOtherVoiceFields": "unchanged"
  },
  "activeOrder": [
    "I"
  ],
  "ownerAHandles": [
    501
  ],
  "S1": {
    "voice": null,
    "cache": "unchanged position(0,0,0)/base0/factor0"
  },
  "W1": {
    "marker": 0,
    "streamHandle": null,
    "inner": "B1 retained",
    "endMarker": -1,
    "allOtherFields": "unchanged from carried checkpoint"
  },
  "resources": {
    "S": "freed",
    "F": "freed",
    "H": "closed once",
    "B1": "retained; no Release; payload unchanged",
    "scratch": "retained"
  },
  "lastLiveStreamSnapshot": {
    "state": "playing",
    "cachedPosition": 10,
    "progress": 16388,
    "silenceTail": 16378,
    "writeCursor": 16384
  },
  "referenceTrace": {
    "S": [
      1,
      0
    ],
    "F": [
      2,
      1,
      0
    ]
  },
  "resourceOrder": [
    "unlink/free S",
    "B1.Stop#1→−1",
    "unlink F; CloseHandle(H)→1; free F",
    "clear wrapper stream/marker",
    "B1.Stop#2→−1"
  ],
  "devicePositionQueryCalls": 1,
  "bookmarkQueryCalls": 1,
  "markerGetterCalls": 0,
  "streamPresencePredicateCalls": 0,
  "StopCalls": 2,
  "ReleaseCalls": 0,
  "CloseHandleCalls": 1,
  "LockCalls": 0,
  "UnlockCalls": 0,
  "PlayCalls": 0,
  "clientListAllocations": [],
  "clientListFrees": [],
  "streamList": [],
  "sourceList": [],
  "requestQueue": [],
  "currentRequest": null,
  "history100": {
    "recent": 7,
    "older": 6
  },
  "LRU": [
    100,
    101
  ],
  "accessTicks": {
    "100": 90,
    "101": 90
  },
  "usedBytes": 96,
  "pendingCount": 0,
  "nextHandleCounter": 501,
  "lastScalarTick": 100,
  "audioCounter": 100,
  "ownersASeed": [
    1,
    0
  ],
  "allOtherState": "exactly unchanged from the entering complete checkpoint; no final field oracle for freed resources",
  "physicalPlayback": "unspecified after ignored negative Stops",
  "AddRefCalls": 0,
  "wrapperRegistry": "complete entering list and nodes retained"
}
```

#### Independent marker-bridge counterpart

Inputs: restart the complete fixture, changing only W1.marker to 0. Record100 Loop remains 0 and W1.streamHandle remains F. This is a supplied checkpoint, not a claim that normal callback cadence produces marker 0 with an attached stream.

Trace: scan 16 slots in order; S0 is empty. I at S1 has decoded Loop0. Read marker 0, then stream-present=true. Invoke the slot stop-and-detach helper. It performs the identical query/cache/save-before-two-Stops/resource-close sequence above. It then stops/detaches I. The marker bridge rereads the cleared S1 association and does no second query, Stop or detachment. Remaining slots are empty.

ExactOutputs: every base output above applies relative to this marker 0 input, except markerGetterCalls=1 and streamPresencePredicateCalls=1. Bookmark77→2; Stops2; Release0; CloseHandle1; I remains inUse with owner/active references. No following coordinator runs. Seeds unchanged(1,0); orderedDraws=[].

#### Independent Loop1 bridge-only variation

Inputs: restart that marker 0/attached-stream bridge input, changing only decoded record 100 Loop to 1. Keep the nonloop producer stream itself unchanged. No normal reachability or Loop1 producer claim is made.

Trace: S1 resolves the valid Loop1 record and is skipped before the marker getter or stream-presence predicate. Other slots are empty.

ExactOutputs: I remains playing/inUse/S1/bookmark 77/references[A]/retirementfalse; S1.voice=I; ownerA handles[501]; active[I]. W1 retains marker 0/stream F/inner B1. S remains playing/cache 4/progress 16388/tail 16378/write 16384/references 1; F remains live/logical 54/references 2/streamFlagtrue/Hopen. Every other fixture value is unchanged. Marker getter 0, stream predicate 0, bookmark query 0, position query 0, Stop0, AddRef0, Release0, CloseHandle0, allocation/free 0; wrapper registry unchanged. Seeds unchanged(1,0); orderedDraws=[].

##### Independent finished-stream lookup variation

Restart STP-V66 main inputs, changing only S.state to finished. Ordinary logical stop reaches lookup, which reports invalid-stream-state error0x85100072 and returns failure without a physical-position query or cache write. The Sound bookmark wrapper returns 0, so I.bookmark becomes 0 before teardown. The same two ignored Stops, stream/file reference drops and closes occur in the same order. Exact outputs equal the main case relative to this changed input except bookmark 0, physical position-query calls 0, and last-live S snapshot state finished/cache 4; progress 16388/tail 16378/write 16384 remain. The engine error is0x85100072. The borrowed buffer and payload remain retained; owner/in-use/active logical references remain as in the main output. No EOF pump or earlier physical Stop is inferred from the supplied finished state. Seeds unchanged(1,0), orderedDraws=[], runtime_observed:no.

#### Discriminates

The bookmark is captured from the live stream before either Stop and before the source is closed. Negative Stops do not undo bookmark 2, resource teardown, marker 0 or logical detachment. Stream attachment, not the wrapper marker alone, determines whether the marker bridge performs teardown. Loop1 skips the entire bridge path. Logical detachment retains inUse and owner references until a separately invoked cleanup.

The layer distinction is observable: the direct attached-stream teardown used by STP-V46 has one Stop; this normal logical stop adds an unconditional outer Stop for two; the STP-V45 EOF-plus-worker path has three Stops across its earlier completion and worker teardown. These counts do not imply identical resource lifetimes or physical success.

runtime_observed: no

#### STP-V67 — Group selection changes the record used for failure retention

**Inputs**

- Enter full audio coordinator with audio enabled 1, periodic producer disabled 0, audio counter 100 and tick argument 100. Selected client unitA is valid with seed(lo 1,hi 0). Tracking disabled 0; marked-slot search enabled 1. No worker or other interleaving.
- Valid decoded records 0..2933 exist. All unlisted fields are zero in this supplied fixture. Record100: Volume255, Group Size2, Loop0, Fade In0, Fade Out0, Defer Inst0, Stop Inst1, Duration0, Compound0, range selector 0, Cache0, Async Only0, Priority20, Stream0, Stereo1, Tracking0, Interrupt0, family 100, history(recent 7,older 6), loaded state 2, failure 0, reference count 0, descriptor null, payloadWAV100 size 48, PCM data length 4, access tick 90. Record101 has the same fields except Group Size0, Loop1, Priority10, history(0,0), payloadWAV101 and its corresponding PCM reference. Both assets have the valid supplied format below.
- WAV100/WAV101 are distinct48-byte synthetic RIFF/WAVE payloads: RIFF size 40, format chunk size 16, PCM format 1, stereo channels 2, sample rate 22050, average byte rate 88200, block alignment 4, bits per sample 16, extra-format count 0, data length 4 and PCM bytes[1,2,3,4]. These payloads are retained inputs; no parser or loader is reached by this case.
- Accounted asset bytes 96, budget 1000000, LRU[100,101], pending count 0, last memory failure 0 and last pending scan 100.
- Only logical voiceI is active, at pool index 1 in the200-entry pool. I: in use 1, cue 100, handle 501, position(0,0,320), finite distance key 0, volume 255, context factor 0, playback offset 0, refs[A], slot null, pending state 0, start/due 100, flags 0, priority 100, retirement 0, saved resume bookmark 0, no active transition and all transition values 0. The other 199 entries are free/cleared. A's handle list is[501]. Exactly one existing node represents each of activeI, A's handle 501 and I's refA; links are valid. Next-handle counter 501.
- All16 ordered slots have backend class 0, no voice, zero property caches and distinct valid wrappers with live inner buffers. Every wrapper has marker 0, capacity 131072, half capacity 65536, flags 0, no stream/3D/property interface and end marker−1; all other stored state zero. Each buffer contains 131072 supplied bytes0xaa. These buffers are valid but class 0 is incompatible with selected Stereo1's requested class 1.
- Actual stored backend master and music getter values are 100 each. Designated music interval 2934..4656; modal mode 2, other modal inputs[0,0], both gains 100, last gain tick 100. Floating inputs use finite stated values,64-bit x87 precision, nearest ties-to-even and masked exceptions.
- Every requested list-node allocation succeeds with a fresh distinct node; every listed free returns normally. Generic allocator metadata is outside the audio result. No device/file result is supplied because no such operation is reached.

**Seeds**

A selection seed(lo 1,hi 0); consume NRF-R01/NRF-R02 at the reached group-selection draw sites.

**Draws**

1. Rand(3): widened 1×1791398085+0=1791398085; seed becomes(lo 1791398085,hi 0); modulo 3 returns 0.
2. Rand(2): widened 1791398085×1791398085+0=3209107098941667225; seed becomes(lo 1721382809,hi 747178471); low-word mask 1 returns 1.

No other draw occurs.

**Trace**

1. Periodic producer is disabled. Only I is sorted; retirement prepass has no effect. Master getter returns 100. Entering cue 100 lies below the music interval, so music getter is not called.
2. Entering record 100 supplies range selector 0, giving radius 400; key 0 is eligible. Due100 and volume 255 permit an attempt. Sole ownerA filters duplicate-family lookup; only I exists and is excluded, so no duplicate candidate. No active transition/bookmark exists and entering Fade In0 selects no automatic fade.
3. Flags0 allow group selection. Group Size2 starts with one history exclusion. First draw 0 reduces exclusions to 0. Second draw 1 selects cue 100+1=101, without a history rejection. Store I.cue101.
4. Search for a marked slot for selected cue 101. Read each of the 16 wrappers' stored markers, all 0; return no marked voice. There is no recent-voice suppression.
5. Commit history in the retained entering base record 100: older becomes 7 and recent becomes 101. Record101's history remains(0,0).
6. Selected record 101 is loaded, not failed and Stereo1. No loading occurs. Request class 1; all slots are class 0 and empty, so neither compatible free slot nor occupied victim exists. Admission returns null without device calls or cache changes.
7. Store I.slot null. The coordinator reloads selected record 101; its Loop1 retains I pending. No playing-property/access/LRU update occurs. The already calculated entering radius remains local for this visit; this case does not vary radius or exercise successful playing Duration.
8. Clear the empty deferred list. Both gains remain 100; store last gain tick 100.

**Exact outputs**

- SeedA(lo 1721382809,hi 747178471), ordered draw results[0,1]. I remains in use 1/pending 0, cue 101, slot null, volume 255, no active transition, retirement 0, refs[A]. Its other stored fields remain unchanged.
- Base100 history(recent 101,older 7); selected 101 history(0,0). All other record fields and asset bytes unchanged.
- Active[I], A handles[501], next-handle counter 501. Access ticks 100:90 and 101:90; LRU[100,101], accounted bytes 96, pending count 0.
- All slots, wrappers, caches and buffer bytes unchanged. Master getter calls 1, music getter calls 0, marker getter calls 16. No list allocation/free, device or file operation.
- Gains[100,100], last gain tick 100; all other live state unchanged.

**Independent variation — inverse Loop fields**

Reset all main inputs, including seed(1,0) and both histories. Set entering record 100 Loop1 and selected record 101 Loop0. Every draw, selected cue, base-history write and admission failure is identical. After reloading selected 101, Loop0/Async Only0 queues I for deferred removal.

Allocate one temporary deferred node. Free nodes in this order: activeI, A handle 501, I refA, temporary deferredI. Between reference cleanup and final temporary-list clear, the pending/null-slot stop helper makes no query or Stop, then I becomes free. Exact outputs: I in use 0, cue 101, pending 0, slot null, volume 255, no transition, retirement 0 and refs[]; active[], A handles[]. Seeds/histories/assets/caches/gains match the main outputs. No device or file call occurs. In particular this failed admission does not itself set retirement.

**Independent variation — entering Fade In before selection**

Reset all main inputs. Change entering 100 Fade In to 5; retain selected 101 Fade In0/Loop1. Before either group draw, save entering volume 255 and set I volume 0. Draw/select/commit history exactly as in the main trace, then fail class 1 slot admission. Selected Loop1 retains I pending at cue 101/volume 0; no automatic fade is installed and volume is not restored. Refs[A], active[I] and A handles[501] remain.

Carry the full result to a following invocation with counter 101 and tick argument 101. Current volume 0 bypasses admission and failed-admission cleanup. Make no new draw, history commit or marker lookup. I stays in use 1/pending 0/cue 101/volume 0/no transition; seed remains(lo 1721382809,hi 747178471), base history(101,7), selected history(0,0). Master getter calls 1; music getter calls 0. Gains saturate at 100 and last gain tick becomes 101. Other live state is unchanged except these supplied clocks.

**Independent variation — successful automatic fade keeps the entering duration**

Reset all STP-V67 main inputs. Set entering 100 Fade In=5 and selected 101 Fade In=20; retain selected Loop1. Set S0 class 1, W0 flags 1/capacity 262144/half 131072 and supply B0 as 262144 bytes0xaa. All other slots remain class 0. Reuse is enabled. Reached SetCurrentPosition/SetPan/SetVolume/Lock/Unlock/Play operations return 0; Lock supplies the entire requested262144-byte region, no second region. The complete SUPPLIED-FINITE-ACOUSTIC-UPDATE profile is supplied, with routing-mode0, master 100, gains 100/factor 0 and base 0 for current volume 0. Backend acoustic-conversion mode 0 and scalar 255 are distinct supplied states. Every reached synchronization operation returns normally; no worker runs.

The two group draws and final seed/history exactly match the main case. Save volume 255 and zero current volume before admission. Sixteen marker reads find no suppression; the earliest compatible free S0 adds one marker read. Prepare W0 (position 0, pan 0, requested 255/SetVolume0), set logical Loop1, parse selected WAV101 and fill the entire buffer with 65536 repetitions of[1,2,3,4]. No block-count1 loop-start setter is reached. Apply the supplied base 0 profile (requestedVolume0/SetVolume−10000/cache XYZ(0,0,320),base 0,factor 0), then Play succeeds. Install I playing/S0 and directly set transition active 1/startVolume0/endVolume255/startTick100/endTick105. The selected Fade In20 does not extend it to 120. The immediately following property call matches the supplied cache and makes no device write. Refresh selected 101 access to 100 and its LRU tail node; clear the empty deferred list.

Exact outputs: I in use 1/playing/cue 101/slotS0/currentVolume0, transition(active 1,from 0,to 255,start 100,end 105), retirement 0/refs[A]/handle 501. A handles[501], active[I], next-handle501. W0 marker 1/stream null/source WAV101.PCM/dataLength4/played 0/written 262144/endData0/halfFlag1/endMarker−1/logicalLoop1/loopStart0/requestedVolume0/propertyScalar0; its buffer is the repeated PCM above. S0 references I and caches XYZ(0,0,320),base 0,factor 0. Base100 history(101,7), selected 101 history(0,0); access 100:90/access 101:100, final LRU[100,101], used 96/pending 0. Seed(lo 1721382809,hi 747178471), ordered draws[0,1]. Master calls 3, routing calls 2, music 0, marker 17. Device log is position 0,pan 0,SetVolume0,Lock0/262144,Unlockfull,SetVolume−10000,Play repeating; no Stop/query/file operation. One old 101 LRU node is freed and one fresh tail node allocated; no other allocation/free. All other live fields retain the resolved changed input. runtime_observed:no.

**Discriminates**

The entering base record controls duplicate and automatic-fade eligibility before selection. Group selection mutates the cue and commits the base history. The selected record then controls failed-admission Loop/Async retention. Treating entering Loop as retained for the whole visit reverses both Loop examples; treating selected Fade In as the earlier policy misses the zero-volume failed loop and its next-tick no-draw bypass.

**runtime_observed: no**

#### STP-V68 — Coordinator sort order determines later LRU order

**Inputs**

Full acoustic-update calls consume the explicit SUPPLIED-FINITE-ACOUSTIC-UPDATE profile declared in the logical fixture, including its matching-cache path, one master/routing sample per call and no music sample for cues 100/101. The getter/cache/base results below are conditional profile assertions under S3D-R07/Q02; sort, visitation, access/LRU and retained logical state are STP-R03b assertions.

Each case starts independently from the same materialized state, then applies only its row in the table below.

- Full audio coordinator, enabled 1; periodic producer 0; audio counter 100 and tick argument 100; tracking 0. Shared property-mix control 1. Selected valid unitA has seed(lo 1,hi 0). No worker interleaving. Next-handle counter 501.
- Only I and O are active. They occupy fixed pool indices 1 and 0 respectively; all other 198 entries in the200-voice pool are free/cleared. I uses cue 100, handle 501 and slotS1; O uses cue 101, handle 500 and slotS0. Both are in use/playing, volume 255, context factor 0, saved bookmark 0, playback offset 0, flags 1, retirement 0 and no transition. Their default position is(0,0,320), finite key 0, start 100 and priority 100. A handles[501,500]; each voice refs[A]. Bidirectional associations and list links are valid. These are supplied playing states, with no claim about how their priority values arose historically.
- Valid decoded records 0..2933 are zero except 100/101. Both have Volume255, range selector 0, Duration0, Loop0, Fade In/Out0, Defer Inst0, Stop Inst1, Compound0, Cache0, Async Only0, Stream0, Stereo1, Tracking0, Interrupt0, option-category0, family 100, loaded state 2, failure 0, reference count 0, no descriptor, raw size 48, parsed PCM length 4 and access 90. Record100 has Group Size2, Priority20 and history(7,6);101 has Group Size0, Priority10 and history(0,0). Neither group field is consumed by these already-playing visits.
- Distinct valid loaded WAV100/WAV101 supply48-byte RIFF/WAVE payloads: RIFF size 40, format length 16, PCM format 1, channels 2, sample rate 22050, average 88200, alignment 4, sample bits 16, data length 4 and PCM[1,2,3,4]. LRU[100,101], used bytes 96, budget 1000000, pending count 0, last memory failure 0 and last pending scan 100.
- All16 slots have class 1 and distinct valid wrappers/inner buffers. S0→O and S1→I; others have no voice. The two occupied slot caches exactly match their associated voice's raw XYZ, base amplitude 255 and factor 0. Other caches zero. W0/W1 markers 1, capacity 262144, half 131072, flags 1, retained requested amplitude 255, no stream/3D/property interfaces, played 0, written 4, end-data1, half flag 1 and end marker 4; each contains its four associated PCM bytes followed by 262140 zeros. Other wrappers have marker 0, no source payload, zero counters and end marker−1; their device bytes are 262144 supplied bytes0xaa. No worker runs.
- Actual stored master getter 100, music getter 100 and backend mode getter 0. The mode getter is distinct from the backend's later volume-conversion mode. Supplied backend scalar 255, conversion mode 0 and factor storage 0 remain unchanged because no backend property operation is reached.
- Designated music interval 2934..4656, modal mode 2, other modal inputs[0,0], modal/interrupt gains 100, last gain tick 100. Ordinary finite floating profile: round-to-nearest ties-to-even,64-bit extended precision, masked exceptions; both arithmetic-workaround selectors 0.
- Existing active/owner/ref/LRU list nodes are valid. Every requested replacement LRU-node allocation succeeds with a fresh distinct node; every free returns normally. Generic heap metadata is outside the audio result. No device/file provider call is supplied or needed.

| Independent case | Initial active order | I priority / O priority | I key / O key | I start / O start | Position/cache changes | Expected visit order | Final LRU |
|---|---|---|---|---|---|---|---|
| A — priority | [I,O] |100 /110 |0 /0 |100 /100 | none | [O,I] | [101,100] |
| B — distance | [I,O] |100 /100 |1 /0 |100 /100 | I position and S1 cached XYZ both(1,0,320) | [O,I] | [101,100] |
| C — start | [I,O] |100 /100 |0 /0 |98 /99 | none | [O,I] | [101,100] |
| D — pool order | [O,I] |100 /100 |0 /0 |100 /100 | none; I index 1/O index 0 | [I,O] | [100,101] |

**Seeds / no draw**

A seed remains(lo 1,hi 0). No RNG is consumed.

**Draws**

None.

**Trace**

1. Enabled coordinator skips the disabled periodic producer and sorts by priority descending, finite distance key ascending, unsigned start descending, then fixed pool index descending. Every table row begins in reverse of the expected order. Retirement prepass has no effect.
2. Read actual stored master 100. Both cues lie below the music interval. Each sorted playing visit has Duration0 and an eligible finite key, with no tracking or transition.
3. For each voice, obtain outer 400/inner 60 from range selector 0, consume the supplied profile base amplitude 255 for current 255/master 100/gains 100 and option-category0, including its supplied mode 0 sample. No nondefault mix branch applies. The associated cache has identical base 255, factor 0 and raw XYZ, so property update returns before attenuation, cache writes or backend operations.
4. After that return, refresh this voice's loaded record access tick to 100. Remove/free its existing LRU node, allocate a fresh node and append its cue to the LRU tail.
5. Cases A/B/C visit O(cue 101) first: LRU[100,101]→[100,101]; then I(cue 100):→[101,100]. Case D visits I first: [100,101]→[101,100]; then O:→[100,101]. Equal cue order after a first tail refresh does not mean the list-node refresh was skipped.
6. Clear empty deferred-removal list. Gains remain 100; last gain tick remains 100.

For B, supplied position(1,0,320) and key 1 are consistent with S3D-R04. S3D-R02/R05 gives distance 1<=inner 60, hence distance amplitude 255 and volume-weighted amplitude 255 for the isolated supplied stage. In this actual visit the matching cache returns before that distance stage. This vector does not claim a newly executed attenuation or generic nondefault-position pipeline.

**Exact outputs**

- Active and final LRU orders are exactly those in the table. Both record access ticks become 100.
- Both voices remain in use/playing with their original cue, handle, slot, refs[A], volume 255, factor 0, position/key, priority/start and transition fields unchanged from each resolved case input. A handles[501,500], next-handle counter 501.
- All slot caches, wrapper state, markers and device bytes unchanged. All table fields other than access ticks unchanged; used bytes 96, pending count 0.
- Per invocation: actual master getter calls 3 (one coordinator plus two properties); music getter calls 0; mode getter calls 2; marker getter calls 0. No draw, device, file, Stop or position-query operation.
- Two original LRU nodes freed and two fresh LRU nodes allocated, in visitation order. No other allocation or free. Deferred list empty; gains[100,100]; last gain tick 100. All other live state unchanged.

**Independent music-audibility variation**

Reset STP-V68 case A. Set I.cue3088. Table count 4657 has valid storage 0..4656; record 3088 replaces record 100 as the loaded record backed by WAV100/PCM4. Copy100's stated fields, then set Group Size0/Loop0/Duration0/Stream0; retain Volume255/Stereo1/range 0/Tracking0/Interrupt0/access 90. Record100 is now all-zero/unloaded with absent payloads and contributes no accounted bytes. Record101 retains its input; used bytes 96/LRU[3088,101] account for the two loaded records. Stored music getter 0, master 100/routing 0; other controls unchanged. I enters playing/S1 with W1 marker 1/no stream; other voice/slot/wrapper state is case A. Stop returns 0. The finite acoustic profile applies to O's cue 101 cache hit only and must not be invoked for I.

Full order[O,I]. Initial master 100; O's nonmusic property cache hit refreshes 101 access 100 and its LRU tail node. I's per-visit music getter 0 causes its playing branch to stop before the acoustic stage. Absent-stream bookmark query returns 0 without physical query; B1.Stop→0 clears marker, then I becomes stopped/no slot and S1.voice=null. I is not queued for removal in this invocation.

Exact outputs: I stays inUse/active/cue 3088/stopped/no slot/bookmark 0/refs[A]/retirementfalse; other voice fields unchanged. W1.marker0/S1.voice=null; caches/other wrapper state unchanged. O stays playing/inUse/S0 with original fields. Active[O,I], A handles[501,500], next handle 501. Access3088:90/101:100, LRU[3088,101], used 96/pending 0, gains 100/100,last gain tick 100. One old 101 LRU node freed and one fresh node allocated. Master2/routing 1/music 1, Stop1/physical-query0; no RNG/admission/load/file/worker or I acoustic-profile call. Seed retained(1,0), draw log[],runtime_observed:no.

**Discriminates**

Skipping or reversing the coordinator sort produces the opposite final LRU order in every case. The cases isolate all four tuple fields without relying on comparator counts or a particular sorting algorithm. LRU order is a functional later-stage output, while backend operations remain empty because each property cache matches.

**runtime_observed: no**


#### STP-V69 — Owned request selection, merge and worker throttle

Seeds / no draw: no random seed is read or changed in any partition. orderedDraws: []. runtime_observed: no for every partition and variation.

##### Shared selector fixture

All named request/file/stream identities and list links are valid. Queue order is exactly the case table order; no other queued request, completion notification or current request exists. Allocation/free and event/device/IO operations are absent during selector entry. The selector clock returns 1000 exactly once. Its prior-dispatch archive identity is null, prior cache key 0 and affinity-enable input 0 unless overridden. All integer fields are 32-bit; priorities and deadline differences are signed, while affinity keys and scores are unsigned.

Each queued request defaults to: file F, source offset 0, cache key 0, deadline 3000, priority 0, cancellation cookie 0, distinct valid raw buffer with capacity 4, device destination null, destination offset 0, stream null, length 4, actualRead 0, freeAfter 1, completion event null, completion metadata reference null. Explicit cache-key changes also set source offset to that same value unless a row says otherwise; no I/O is reached. Default file F has a live external owner, valid direct handle, archive null, stream flag 0 and zero unlisted semantic fields. Its reference count is one external reference plus one per queued request naming F, plus one per listed stream retaining F. No close occurs.

Where a request attaches stream S, the stream list contains exactly S: file F, buffer B valid/borrowed, ownsBuffer 0, playing state, capacity 16384, chunk 16384, data start 0/length 100, average rate 1, start/cache/progress/tail/write 0, loop 0. Its reference count is one list-owner reference plus one per attached queued request. File F stream flag is 1 in those cases. No stream pump/callback runs. Other semantic fields are zero/null, preserving named live identities. Buffer/source bytes are opaque retained supplied values because no read or write reaches them.

Outputs common to every selector partition: selected identity and urgency are the only semantic outputs; no request, queue order, reference, file, stream, buffer, notification, event or global field changes. The result is the pair(selected request, urgency), not a promise to dispatch immediately. Urgency is 1 for selector classes 1–3,0 for class 4. Every case begins with no previously selected pointer; exact ties retain the first eligible request in queue traversal order.

##### A. Selector classes, class-local ties and predecessor cookie

###### A1 — attached-stream near-deadline class outranks other work

Inputs: queue[Q1,Q2,Q3,Q4], with these overrides. Deadline delta is deadline −1000 interpreted signed.

| Request | Stream | Priority | Deadline | Signed delta |
|---|---|---:|---:|---:|
| Q1 | null |9 |900 |−100 |
| Q2 | S |−5 |1999 |999 |
| Q3 | S |100 |1900 |900 |
| Q4 | null |10 |500 |−500 |

File F references 6: external owner, S and four requests. Stream S references 3: list owner and Q2/Q3. Seeds/no draw:none. orderedDraws:[].

Trace: Q1 first supplies high-priority class 2. Q2 enters higher class 1 despite its later deadline; Q3 replaces it on smaller signed delta 900. Q4 cannot replace a selected class 1 candidate. ExactOutputs:(Q3,urgency 1), queue unchanged, file F references 6/stream S references 3 and all other shared state unchanged. Discriminates: stream+delta<1000 is the first class; priority magnitude and more-overdue nonstream work do not defeat it. runtime_observed:no.

Independent tie variation: reset A1; Q2 deadline 1900. Q2 and Q3 now both delta 900; select Q2/urgency 1 despite Q3 priority 100 versus Q2−5. All references/state unchanged. This tests first-in-class tie retention without a generic priority tie-breaker.

###### A2 — positive signed priority class ignores magnitude

Inputs: queue[Q1,Q2,Q3,Q4].

| Request | Stream | Priority | Deadline | Signed delta |
|---|---|---:|---:|---:|
| Q1 | S |0 |2000 |1000 |
| Q2 | null |2 |5000 |4000 |
| Q3 | null |99 |5000 |4000 |
| Q4 | null |0 |900 |−100 |

File F references 6; stream S references 2. Seeds/no draw:none. orderedDraws:[].

Trace: Q1 fails class 1 at exact delta 1000. Q2 enters class 2; Q3's equal delta retains Q2 regardless larger priority 99. Q4 is only class 3 and cannot replace class 2. ExactOutputs:(Q2,urgency 1); all queue/reference/shared state unchanged. Discriminates: signed priority>=1 creates a class, not magnitude ranking within that class; class 2 beats a more-overdue ordinary request. runtime_observed:no.

###### A3 — ordinary near-deadline class and equality boundary

Inputs: queue[Q1,Q2,Q3,Q4].

| Request | Stream | Priority | Deadline | Signed delta |
|---|---|---:|---:|---:|
| Q1 | S |−2 |2000 |1000 |
| Q2 | null |0 |900 |−100 |
| Q3 | null |−9 |800 |−200 |
| Q4 | null |0 |800 |−200 |

File F references 6; stream S references 2. Seeds/no draw:none. orderedDraws:[].

Trace: Q1 is class 4 at exact delta 1000. Q2 enters class 3; Q3 replaces it at −200; equal Q4 retains Q3 even though Q4 priority 0 is greater than −9. ExactOutputs:(Q3,urgency 1); all other state unchanged. Discriminates: class 3 is signed delta<1000 and uses delta, not priority magnitude; exact 1000 is excluded. runtime_observed:no.

###### A4 — fallback priority, unsigned affinity, delta and exact tie

Inputs: no streams; affinity enabled 1, prior archive null/prior key 0; queue[Q1,Q2,Q3,Q4,Q5]. All files are direct F/archive null.

| Request | Priority | Cache key/source offset | Deadline | Signed delta | Affinity score |
|---|---:|---:|---:|---:|---:|
| Q1 |−1 |0 |2000 |1000 |0 |
| Q2 |0 |100 |3000 |2000 |100 |
| Q3 |0 |90 |4000 |3000 |90 |
| Q4 |0 |90 |3500 |2500 |90 |
| Q5 |0 |90 |3500 |2500 |90 |

File F references 6. Seeds/no draw:none. orderedDraws:[].

Trace: all candidates are class 4. Signed priority 0 defeats Q1 despite its best affinity/delta. Score 90 defeats score 100; delta 2500 defeats 3000; exact tie keeps Q4 over Q5. ExactOutputs:(Q4,urgency 0); queue and references unchanged. Discriminates: the fallback tuple is signed priority descending, unsigned score ascending, signed delta ascending, then traversal retention. runtime_observed:no.

Independent affinity-disabled variation: reset A4, affinity input 0. Every score is 0; Q2's delta 2000 wins among priority 0. ExactOutputs:(Q2,urgency 0); all state unchanged. Direct-file null archive identity is included, not treated as a mismatch merely because there is no archive.

###### A5 — adjacent predecessor cookie, not global deduplication

Inputs: no streams; queue[Q1,Q2,Q3,Q4], file F references 5. K is one nonzero cookie identity.

| Request | Cookie | Priority | Deadline | Signed delta |
|---|---|---:|---:|---:|
| Q1 | K |0 |5000 |4000 |
| Q2 | K |10 |0 |−1000 |
| Q3 |0 |0 |4000 |3000 |
| Q4 | K |10 |100 |−900 |

Seeds/no draw:none. orderedDraws:[]. Trace: Q2 is skipped because its immediately preceding queued request Q1 has the same nonzero cookie. Q4 remains eligible because Q3 precedes it and has cookie 0, even though earlier Q1 also has K. Q4's class 2 beats the ordinary candidates. ExactOutputs:(Q4,urgency 1); Q2 is not removed or mutated; full queue/ref state unchanged. Discriminates: adjacent-predecessor filtering versus global same-cookie exclusion or ignoring the filter. runtime_observed:no.

Independent zero-cookie variation: reset A5 and set all four cookies 0. No candidate is skipped; Q2 delta −1000 wins class 2 over Q4−900. ExactOutputs:(Q2,urgency 1); no state mutation.

##### B. Affinity below-base arithmetic

Inputs: selector clock 1000, affinity enabled 1, previous archive identityA, previous cache key 8192. Archive A and different archiveB each have sector size 4096. Files FA/FB have no direct handles and reference the corresponding valid archive and file-record metadata. Archive A references 1 through FA; archive B references 1 through FB; those counts remain unchanged. No default file F or stream exists. FA references 3(external,Q1,Q2), FB references 2(external,Q3). Queue[Q1,Q2,Q3]; cookie 0/priority 0/other request fields as shared fixture. Explicit cache keys are supplied selector-stage state; no prior constructor or source I/O is asserted.

| Request | File/archive | Cache key | Deadline | delta | Score |
|---|---|---:|---:|---:|---:|
| Q1 | FA/A |5 |5000 |4000 |134217733 =0x08000005 |
| Q2 | FA/A |134219776 =0x08000800 |6000 |5000 |134215680 =0x07fff800 |
| Q3 | FB/B |0 |2000 |1000 |4294967294 =0xfffffffe |

Seeds/no draw:none. orderedDraws:[].

Trace: base=8192−4096=4096. Q1 is below base, so score=key+0x08000000=134217733, without subtracting base. Q2 is at/above base, score=134219776−4096=134215680. Q3 archive differs, score 0xfffffffe. All are class 4; Q2 wins by unsigned score despite its later delta. ExactOutputs:(Q2,urgency 0); queues/files/archives/references unchanged. Discriminates: incorrectly subtracting base in the below-base branch would give Q1 score 134213637 and choose Q1. runtime_observed:no.

Independent exact-sector-boundary variation: reset B, previous key 4096. The comparison is strict greater-than; base stays 0. Scores Q1=5,Q2=134219776,Q3=0xfffffffe; select Q1/urgency 0, all state unchanged. No actual archive decoding or file-byte read is reached.

##### C. Priority update, compatible merge and retained fields

Inputs: exported priority-update wrapper with nonzero cookieK/new priority 1; queue[Q1,Q2,Q3], current null, stream list empty, notifications empty. Valid returning lock/free providers. Wake event Ewake initially unsignaled; SetEvent returns BOOL 0 and leaves it unsignaled. No clock, I/O, device or other event operation occurs. All request actual counts 0/device destinationsnull/streams null.

| Field | Earlier matching Q1 | Nonmatching Q2 | Later matching Q3 |
|---|---|---|---|
| file | F1 | F2 | F3 |
| file references |2(external,Q1) |2(external,Q2) |2(external,Q3) |
| cookie | K | X, different nonzero | K |
| old priority |0 |9 |0 |
| source offset |100 |0 |106 |
| requested length |6 |4 |4 |
| cache key |200 |0 |300 |
| deadline |1111 |3000 |2222 |
| raw buffer | Rfirst, valid capacity 10 | Rother, capacity 4 | Rlater, capacity 4 |
| freeAfter |1 |1 |0 |
| completion event | null | null | Enotify, unsignaled |
| completion metadata reference | null | null | M, whose contained function is null |

Each file has a valid direct handle, archive null, stream flag 0 and no auxiliary owned buffer; its external owner remains. No source bytes are read. Raw buffers and metadata/event objects remain valid supplied identities. All other request fields are zero/null. Different earlier/later files, later freeAfter 0 and nonnull later metadata/event are explicit conditional inputs; no historical constructor reachability or future combined read is claimed. The combined destination Rfirst has capacity 10.

Seeds/no draw:none. orderedDraws:[].

Trace: enter lock. Q1 matches, assign priority 1, remember Q1. Q2 does not match and stays priority 9. Q3 matches; save its old priority 0, assign priority 1, attempt merge with previous matching Q1 despite intervening Q2. Cookies match;100+6=106;6+4=10<=1048576; earlier freeAfter 1; all required zero/null fields pass. Copy Q1 source offset 100/cache key 200/deadline 1111/raw buffer Rfirst into Q3; set Q3 length 10. Destroy Q1: file F1 references 2→1, unlink/free Q1. Q3 retains file F3/freeAfter 0/completion event Enotify/metadata M/new priority 1. Leave lock; request SetEvent(Ewake), BOOL 0 ignored.

ExactOutputs: queue[Q2,Q3]; Q1 freed; file F1 references 1, F2 references 2, F3 references 2, all live/open and flags 0. Q2 unchanged. Q3(source 100,length 10,key 200,deadline 1111,raw buffer Rfirst,priority 1,file F3,freeAfter 0,event Enotify,metadata M), actual 0/device destination null/stream null. Ewake/Enotify both unsignaled; current null; notifications empty; one free(Q1), zero allocations/CloseHandle/device calls, one wake SetEvent. No event notification is transferred by ordinary request destruction. Other buffers/metadata unchanged.

Discriminates: merge uses the most recent matching request across nonmatches; copies only the defined earlier fields, keeps later file/event/metadata/freeAfter and does not invent an equal-file gate. runtime_observed:no.

Independent completion-metadata-reference gate: reset C; Q1 completion metadata points to Mempty with contained callback function null. The NONNULL reference blocks merging even though that function is null. Priorities Q1/Q3 still become 1; queue stays[Q1,Q2,Q3], each file references 2, all request fields otherwise unchanged; no free/close/allocation; wake SetEvent still once/BOOL 0/event unsignaled; Enotify unchanged. Discriminates: testing only the contained callback rather than the metadata reference is wrong.

Independent old-priority gate: reset C with Q3 old priority 1. No merge is attempted from the priority-update loop; outputs are the same unmerged queue/ref result, Q1/Q3 priority 1, but Q1 metadata stays null. One ignored wake signal request, no frees. New priority magnitudes and other merge rejection gates remain additional source-defined partitions, not executed here.

Independent null-cookie wrapper variation: reset C with an initially zero engine error and replace the operation cookieK with null, retaining new priority 1. The wrapper writes invalid-parameter engine error 87 and requests SetLastError(87), whose supplied provider stores 87 and returns normally. It returns without entering stream synchronization, traversing or modifying requests, merging/freeing, or requesting the wake event. Exact outputs: original queue[Q1,Q2,Q3], old priorities[0,9,0], file references[2,2,2], buffers/events/metadata/notifications/current state unchanged; engine and supplied external error 87. No scalar return-value assertion is made for this wrapper. No RNG/clock/read/device/allocation/free/Wait/SetEvent occurs; orderedDraws=[], runtime_observed:no.

Independent merge-length-cap variants: reset partition C separately for each following case, retaining all its other fields/providers. In the equality case set Q1 length 1048572, Q3 source offset 1048672 and supply Rfirst with capacity 1048576. Contiguity is 100+1048572=1048672 and combined length 1048576 equals the allowed cap. Merge occurs: Q1 freed/F1 references 1, queue[Q2,Q3], Q3 source 100/cache 200/deadline 1111/rawRfirst/length 1048576/priority 1 and retained fileF3/freeAfter0/eventEnotify/metadataM. Other C successful-merge outputs and order apply. In the above-cap case set Q1 length 1048576, Q3 source offset 1048676 and Rfirst capacity 1048580. Contiguity passes but combined length 1048580>1048576 rejects merging. Exact outputs: queue[Q1,Q2,Q3], file refs all 2, Q1/Q3 priorities 1, all other fields equal this changed input, no free/close; one failed ignored wake leaves events unsignaled. Neither case reads or writes raw/source bytes. Both have no draws and runtime_observed:no. A wrong strict-less cap test rejects the equality case; omitting the cap merges the above-cap case.

##### D. Worker throttle and dispatch-start checkpoints

All D partitions begin under stream synchronization immediately after selection, before the worker's dispatch-versus-wait decision. Their named entry/loop/selector clocks and previous state are carried from the declared prefix; no new clock is sampled inside this decision segment. The stream pump was inhibited 1; stream/notification lists empty; current null; queue[Q1]. Q1 is a nonstream raw read request: file F/direct H/archive null/source 48/cache key 48/length 6/raw R capacity 6/device destination null/cookie 0/actual 0/freeAfter 1/completion metadata null; completion event Enotify unsignaled. File F references 2(external,Q1), flag 0, valid unlisted semantic fields zero/null. Affinity disabled 0. Source bytes and destination R are opaque unchanged supplied arrays. No actual read/complete/destructor/device/notification call is included. Source routing is not executed. Exit input initially 0. Critical leave returns normally.

Throttle numerator 2, bandwidth 10 give unsigned floor(2*1000/10)=200 ms. Initial previous archive null/cache key 0, lastDispatchTime 1000 from entry clock 1000. A nonurgent single Q1 with deadline 3000/priority 0 is selected in class 4. Queue storage and file references do not change merely from a defer decision.

###### D.1 — below-throttle deferral

Inputs: loop clock 1100, selector clock 1100, chosen Q1/urgency 0; elapsed 100. Wake event Ewake is a valid auto-reset event initially signaled. Wait(Ewake,100) returns 0, makes Ewake unsignaled and, as the only permitted interleaving, writes exit input 1 before returning. Seeds/no draw:none. orderedDraws:[].

Trace: elapsed 100<200; leave Q1 queued/current null, leave lock; timeout argument=200−1100+1000=100. Wait; tail sees exit 1, clears exit 0 and returns 0 without another iteration. ExactOutputs: Q1 still queued/unmodified, F references 2; previous archive/cache key/dispatch time remain null/0/1000; Enotify unsignaled, Ewake unsignaled; one Wait 100, no read/free/SetEvent/selection-after-wait. Worker return 0 within this supplied suffix. Discriminates: nonurgent selected pointer is not sufficient to dispatch when nonzero bandwidth throttles. runtime_observed:no.

###### D.2 — equality dispatch

Inputs: reset D with loop clock 1200/selector clock 1200, urgency 0; deadline 3000 gives signed delta 1800; elapsed 200. No Wait/provider write is supplied. Stage ends after publication/unlink and lock leave, before read-target selection. Seeds/no draw:none. orderedDraws:[].

Trace: elapsed 200>=throttle 200 dispatches. Retain Q1 cache key 48/archive null and loop clock 1200 as lastDispatchTime; set current request Q1, unlink it, leave lock. ExactOutputs: queue empty/current request Q1 live; F references 2; Q1 fields unchanged; previous archive null/cache key 48/lastDispatchTime 1200; exit 0, events unchanged. No Wait/read/free/SetEvent. Discriminates: equality admits dispatch, and saved time is the dispatch-start clock. runtime_observed:no.

###### D.3 — urgent bypass

Inputs: reset D.1 clock inputs 1100, but Q1 priority 1 so selector class 2/urgency 1 at deadline 3000. Stage ends at the same pre-read checkpoint as D.2; no Wait/provider exit write. Seeds/no draw:none. orderedDraws:[].

Trace: urgency bypasses elapsed 100<throttle 200. ExactOutputs: queue empty/current request Q1 live, F references 2, previous archive null/key 48/lastDispatchTime 1100; event/exit/request state otherwise unchanged. No Wait/read/free/SetEvent. Discriminates: priority urgency bypasses throttle; deadline 3000 is not an unconditional sleep-until time. runtime_observed:no.

###### D.4 — prior completion does not replace dispatch-start time

Inputs: reset D.2's entry state but lastDispatchTime 1200, loop/selector clock 1750, Q1 deadline 4000, priority 0/urgency 0. The preceding request was dispatched at 1200 and completed at the supplied environment time 1700; that completion did not sample/store a replacement dispatch timestamp. Previous archive null/cache key 900. This is a supplied next-decision checkpoint, not a reconstructed hidden read; no earlier request remains live. Seeds/no draw:none. orderedDraws:[].

Trace: elapsed=1750−1200=550>=200; dispatch Q1, retain its key 48/archive null and lastDispatchTime 1750. ExactOutputs: queue empty/current request Q1 live/F references 2; no Wait/read/free/SetEvent, exit/event/request states unchanged. An incorrect completion-time throttle would compute 1750−1700=50 and defer 150 instead. Discriminates: I/O duration contributes to elapsed time from dispatch start; no universal rate/cadence is inferred. runtime_observed:no.


### 8.2 Observation vectors

**STP-O01** Measure a naturally initiated volume transition at the start, at the end and immediately after it, comparing active state and current volume against STP-R02. [Reproduction procedure and collection limits](../research/spec-evidence/sound-table-and-playback-rules.md#o01-transition-observation).

## 9. Acceptance and dependencies

STP-V01–STP-V69 define partial acceptance for the verified supplied-input boundaries and explicitly conditional profiles. T01 supports STP-R01 only. Evaluate each case at its declared entry/exit; helper, post-read and decision segments do not claim unexecuted caller or whole-worker behavior. Assertions attributed to a supplied acoustic profile test STP ordering/retention conditionally; they do not establish the gated acoustic integration. Passing these vectors does not establish complete sound-table playback parity or matching gameplay observation.

Consumes NRF-R01/NRF-R02 for exact draw steps/forms and numeric foundations for width/rounding; [S3D-R02–R06](sound-3d-position-and-falloff.md) for verified scalar/radius/key stages, with full tracking/acoustic integration explicitly gated by S3D-R07/Q01–Q03. Source/resource, codec, event/thread, allocator and device operations consume the exact providers declared by each boundary. Enabled playlist selection is owned by [music selection](sound-music-selection-and-transitions.md), while the outer dispatcher's ambient/music controllers retain their own contracts. Reverse consumers are unit, ambient, skill/missile and music request sources; their request production is not verified by this playback contract. Full reproduction still requires resolving STP-Q01/STP-Q02 and recording the STP-Q03 observations.

## 10. Related specs

[Numeric and RNG foundations](core-numeric-and-rng-foundations.md) supplies arithmetic terminology. Positional, unit, ambient, skill/missile and music sound contracts own their respective request sources and spatial policy.

## Appendix A. Provenance

Original Windows x86 Expansion 1.13c; [source manifest](../research/source-manifest.md) identifies D2Client and D2Sound. Static source readings and audits span 2026-10-29–2026-10-31, with compatible earlier evidence retaining its original method/date/limits. A narrow Storm-loader table read was performed on 2026-10-29. The source manifest also identifies Storm and Fog for the asset/stream ownership paths. No gameplay vector is runtime-confirmed. The [companion](../research/spec-evidence/sound-table-and-playback-rules.md) records original function locations, byte corroboration, actual collection failures and remaining boundaries.
