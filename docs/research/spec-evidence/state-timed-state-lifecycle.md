# Timed state lifecycle evidence

Behavioral contract: [Timed State Lifecycle](../../specs/state-timed-state-lifecycle.md).
Original artifact identities: [source manifest](../source-manifest.md), Windows x86 Lord of
Destruction Expansion 1.13c. Function descriptions below are analyst descriptions, not
original exported names. References identify original binaries, not a required source distribution.

## Identity, reuse and dates

The 2026-10-12 static investigation is retained at its actual scope: original-image table
reads, disassembly, reconstruction checked against original bytes for small primitives, and decompiler-assisted
leads. The available D2Game/D2Common images were fingerprinted on 2026-10-14 and match the
shared manifest; D2Client's current analysis identity is also recorded there. That correspondence
does not retroactively identify every historical live DLL or establish archive identity.
There were no historical gameplay observations. No archive cell or compiled-table value is
newly asserted by this migration.

The boundary review found that complete free/attach/stat propagation was not established by
the earlier caller-level readings. STL-R05, STL-R06b and STL-R06c therefore remain gated in
the current contract; ordered traversal requests do not establish their stored effects.
The foundations contract's RLI-Q02/RLI-Q03 and status application's integration gaps agree
with this distinction. Compatible lookup/state guards are reused from RLI-R03 rather than
reconstructed from an incomplete lifecycle summary.

## Function references

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | 0x8EA80, 0x8E680 | none | Schedule and allocate/link frame event | STL-R01 | 2026-10-12 disassembly and caller/record-field reading; malformed signed values remain open |
| LOD113C-D2GAME | 0x8F070; 0x8E9A0, 0x8ECC0, 0x8EBC0, 0x8EF20, 0x8EDD0 | none | Group dispatch; missile/player/monster/object/item dispatchers | STL-R02, STL-R02a | 2026-10-12 dispatch reading; group ordering was decompiler-only, requiring checked decoding |
| LOD113C-D2GAME | 0x769A0; 0x77820, 0x74870, 0xD1780 | none | Monster gate; player/monster/item state-expiry handlers | STL-R03 | 2026-10-12 disassembly and reconstruction checked against original bytes; handlers reach #11093 |
| LOD113C-D2COMMON | 0x3ADD0 | 11093 | Primary-chain expiry walk | STL-R04 | 2026-10-12 complete traversal/control-flow reading; FreeList effects gated separately |
| LOD113C-D2COMMON | 0x3A9F0, 0x3A3B0 | none | Free and unlink list | STL-R05 | 2026-10-12 reconstruction checked against original bytes/disassembly; 2026-10-14 full unlink listing inspected; propagation callees prevent complete verification |
| LOD113C-D2COMMON | 0x38100 | 11013 | Allocate initialized effect list | STL-R06a | 2026-10-12 reconstruction checked against original bytes |
| LOD113C-D2COMMON | 0x38060, 0x37D70 | 10785, 10911 | State/callback setters | STL-R06a | 2026-10-12 reconstruction checked against original bytes |
| LOD113C-D2COMMON | 0x38020, 0x38380, 0x38000 | 10161, 11043, 10412 | Expiry setters/getter | STL-R06a | 2026-10-12 reconstruction checked against original bytes; positive expiry adds timed, zero/negative does not clear |
| LOD113C-D2COMMON | 0x38520, 0x37F80 | 10871 | State-list lookup | STL-R06a | 2026-10-12 reading; exact guards/range behavior consumed from RLI-R03 |
| LOD113C-D2COMMON | 0x3A7D0 | 10807 | Attach list | STL-R06b | 2026-10-12 decompiler lead; 2026-10-14 disassembly inspected; multiple propagation and extended-list branches remain gated |
| LOD113C-D2COMMON | 0x3A280 | 10261 | Replace list stat | STL-R06c | 2026-10-14 full disassembly: entry search/insert/remove helpers, delta propagation and special send-list handling still need complete contracts |
| LOD113C-D2COMMON | 0x34510, 0x38730 | 10945 | Set active/changed state | STL-R07, STL-R07a | 2026-10-12 bit-write reading; 2026-10-14 original-image decode checks byte-only tail; disguise helper still open |
| LOD113C-D2COMMON | 0x34450 | 11039 | Force changed-state membership | STL-R08 | 2026-10-14 full disassembly, no undecoded behavioral bytes; update mark crosses to RLI-Q02 |
| LOD113C-D2COMMON | 0x665C0 | none | Mark unit for update | STL-R07, STL-R08 | Call identified; room queue effects remain RLI-Q02, not established by bit writes |
| LOD113C-D2GAME | 0xDC780, 0xDDAC0, 0xDCC20, 0xDCAB0, 0xDC940 | none | Cold/freeze/stun/poison/burn appliers | STL-R09 | Compatible 2026-10-12 and cited CSA/MMF policy records; not new applier verification |
| LOD113C-D2GAME | 0x29240 | none | Timed skill-state application | STL-R10a | 2026-10-12 disassembly, ordered calls; supplied dependency boundary |
| LOD113C-D2COMMON | 0x51BF0, 0x52AE0, 0x33110 | 10786, 10418, 10819 | Expression front end, passive refresh, animation refresh | STL-R10a, STL-R11 | SCX-R11 supplies expression result; complete refresh effects remain integration gaps |
| LOD113C-D2GAME | 0x448E0 | none | Six skill stat-slot writes | STL-R10b | 2026-10-14 disassembly: primary write thunk 0xA76C; paired write thunk 0xA6E2; forwarding equivalence is established by the separate wrapper reading below |
| LOD113C-D2COMMON | 0x3A7B0 | 10188 | Paired skill stat-write forwarding wrapper | STL-R10b | 2026-10-14 reconstruction checked against original bytes; null guard then tail-forward unchanged four arguments to +0x3A280 |
| LOD113C-D2GAME | 0x9F7A0; 0x27190, 0x422B0; 0xDB4C0, 0xDBD90 | none | Generic removal callback and cold/freeze variants | STL-R11 | 2026-10-12 disassembly/reconstruction checked against original bytes; callback request order, death gate; refresh callees gated |
| LOD113C-D2GAME | 0x37110 | none | Cancel matching unit timers | STL-R11a | 2026-10-14 full disassembly: busy test uses entry flag value 1, pending-removal flag value 2; nonbusy detaches/releases; successor saved before action |
| LOD113C-D2COMMON | 0x33AB0 | 11061 | Stay-death mask consumer | STL-R12a | 2026-10-12 disassembly; complete 80-byte function reread 2026-10-14: mask intersection is returned directly, not normalized; null/nonmonster selects player mask, monster selects monster mask; no boss branch |
| LOD113C-D2COMMON | 0x5F6E0, 0x5DAD0 | none | States loader and flag-mask builder | STL-R12b | 2026-10-14 loader tail and full bit-helper listing inspected; producer byte checks/definition correspondence remain explicit |
| LOD113C-D2GAME | 0x78230, 0x75350 | none | Player/monster death cleanup | STL-R13a | 2026-10-12 caller reading; ordered dependency requests; 2026-10-15 review reread: +0x75350 faults for non-monster units before list removal; callers +0x79730/+0x79E40 (player) and +0x75460/+0x76270 (monster) |
| LOD113C-D2COMMON | 0x3AD60, 0x34150, 0x1A550 | 11004, 10627, 10064 | Non-staying list traversal, state partition, boss test | STL-R13a, STL-R13b | 2026-10-12 traversal/caller reading; complete 384-byte original-image partition decoded/read 2026-10-14, as recorded below; mask production remains gated |
| LOD113C-D2GAME | 0x7A5E0 | none | State-message sender | STL-R14 | 2026-10-14 full disassembly inspected; exporter, real root and clearing remain open |
| LOD113C-D2COMMON | 0x342D0, 0x383C0 | 10315, 10658 | Changed-state half getter and list-stat exporter | STL-R14 | Getter/export calls identified; exporter result and clearing integration not verified |
| LOD113C-D2CLIENT | 0x72530, 0x720B0 | none | State set/remove callback dispatch | STL-R15 | 2026-10-12 lead; 2026-10-14 full listings inspected; event callees and packet roots incomplete |
| LOD113C-D2GAME | 0x6E470 | none | Auxiliary timed-effect expiry | STL-R17 | 2026-10-14 full disassembly; comparison is unsigned `ja`, auxiliary removal consumer remains separate |

## Dispatch and output anchors

The frame-event group entry is D2Game+0x8F070. Player dispatch D2Game+0x8ECC0 and
monster dispatch D2Game+0x8EBC0 (then +0x769A0) select type 12 handlers +0x77820 and
+0x74870. Item dispatch +0x8EDD0 selects +0xD1780. These handlers import
D2Common#11093, resolved to +0x3ADD0; this walks the primary list, invokes free
+0x3A9F0 and unlink +0x3A3B0. The removal callback is read from that list and receives
the former unit/state/list while the list is still readable. Actual contribution propagation
inside unlink is not a completed free contract.

The type tables read in 2026-10-12 were D2Game RVAs0xFBB74 (player),0xFBC14
(monster),0xFB110 (item),0xFBDB8 (object); group-map RVA0xFA3DC. The object table's
type 6 selects +0x6E470, an auxiliary effect, and its type 12 slot is null. The object
dispatcher uses the imported pointer-check slot RVA0xF8D14; the import/return/failure path
was not resolved. It does not prove that any timed-state caller schedules an object type 12.

The timed skill-state operation D2Game+0x29240 is an application entry in the retained
investigation, but a complete original skill-table dispatch selection for every caller was
not recorded. The supplied-entry request sequence remains its evidence boundary. The 2026-10-15 review added one root anchor: the original-image skill do-function table (D2Game .data RVA 0x1074A8) slot 68 at RVA 0x1075B8 holds +0x29720, which calls +0x29240; a second recorded caller is +0x3E8A0. Which Skills rows select slot 68 was not enumerated. Its duration
call reaches D2Common#10786/+0x51BF0, and its six-stat operation is D2Game+0x448E0.
The latter's ordinary stat call resolves through thunk+0xA76C to D2Common#10261/+0x3A280;
the attackrate-paired call through thunk+0xA6E2 resolves to #10188/+0x3A7B0.

Death roots retained from 2026-10-12 are D2Game+0x78230 and +0x75350. Both call
D2Common#11004/+0x3AD60, then #10627/+0x34150. Player death's preceding stat write
imports #10887; other-unit death additionally consumes the auxiliary effect operation.
The boss test imports #10064/+0x1A550. Complete production death reachability is no broader
than that retained record; isolated traversal/request fixtures do not establish every death caller.

## Semantic mapping and unresolved callees

Original internal encodings are investigation metadata, not required storage layouts:
list flags value 2 means timed, value 4 mode-bound,0x2000 secondary,0x40000000 filtered apply,
and 0x80000000 extended/state-storage. Death exemption tests mask 0x181. Unit byte+0xC6
bit 0 is the neutral aliveExclusionCondition; unit+0xC8 bit 3 is the disguise condition. State
storage's active half precedes the changed half by ceil(stateCount/32) words. The unit's
primary/secondary list heads are extended-list+0x3C/+0x40; auxiliary effect is unit+0xA4.

Unlink +0x3A3B0 calls stat update +0x39F70 and special-stat helper +0x39CE0.
Attach +0x3A7D0 invokes unlink, +0x39F70 and +0x39CE0, and traverses special extended-list
branches. Replacing write +0x3A280 invokes search +0x382B0, insert +0x38640, removal
+0x381A0, propagation +0x39F70 and special-send helper +0x38550. These local callees
are not replaced by a guessed generic delta rule. Actual effects remain STL-Q07 until read
through their owned outputs, while supplied traversal fixtures remain independent.

The disguise tail of D2Common+0x38730 calls state-definition getter+0x1770 and
active-mask helper+0x33C20; the latter consumes the disguise mask at DataTables+0x10C.
The 2026-10-14 original-image decode checked all 288 bytes, including the former byte-only
tail. That clears decoding uncertainty, but not the unread helper/producer boundary.

## Corrections and limits

- The 2026-10-12 record called D2Game+0x988B0 the state synchronization caller. A full
  2026-10-14 listing shows a skill/monster eligibility/chance routine, with expression
  evaluation and an optional modulo 100 draw, and no call to the state-message sender.
  It is not support for STL-R14. Sender+0x7A5E0 still contains the 0xA7/0xA8/0xA9 branches;
  actual root, predicate and changed-state clearing are STL-Q01. This discrepancy does
  not add a random draw to the lifecycle contract.
- The paired stat 69 write in +0x448E0 uses a different imported operation from stat 68.
  The 2026-10-14 byte-checked reconstruction of #10188/+0x3A7B0 resolves the concern: it null-checks its first argument and tail-forwards all four arguments to +0x3A280. Thus both requests use replacing semantics; complete setter effects remain gated. No numeric fixture output changed.
- Cold removal is D2Game+0xDB4C0, not+0xDA4C0. A code-section-relative label had caused
  the wrong address. Cold sets the state before allocating its list; the stat call is
  replacing, not an additive request. These retained corrections agree with CSA/MMF.
- D2Common+0x32E10 is a mode-bound velocity-list creator, not this timed-state allocator.
  The timed allocator is #11013/+0x38100. D2Common+0x38550 handles a send-list, and
  +0x2AB3E belongs to item-stat decoding; neither is an expiry getter.
- States setfunc/remfunc dispatch is client-side. Earlier searches of server state-record
  readers identified srvactivefunc handling, not these client selector columns.
- The auxiliary expiry's unsigned comparison is explicit in the original listing. Its
  lifetime does not become a state merely because its effect has a frame field.

## Search and gap-closure record

2026-10-12 transition search examined callers of D2Common#11004/#10627 through D2Game
thunks+0xB01E/+0xB012 and #11093 through +0xAA2A. Recorded callers were death roots and
the three state-expiry handlers. This does not search all transition paths and establishes
neither absence nor persistence. No rule is labeled not found.

2026-10-14 migration reused those searches, NRF's supplied lookup/state contract, GDF's
state-field schema and CSA/MMF's applier contracts. It targeted the previously decompiler-only
attach, setter, forced-change, timer cancellation, mask builder, partition, stat-slot writer,
sender and client dispatch listings. Forced changed-state writes and timer cancellation now
have direct static contracts and new executable fixtures. Other paths retain exact gates;
reading a top-level listing does not close unread local callees. The complete original and
new investigation details are preserved privately, outside the repository.

## Runtime records

None. All STL vectors have runtime_observed:no. No gameplay campaign was performed.
No new game-loader table reads were available; named state/stat identities use GDF's
recorded original-profile evidence. Supplied stay-death masks in STL-V08/STL-V09 do not
assert unrecorded cells, compiled masks or loader correspondence.

## Observation procedures

### STL-O01 expiry frame

Use an owned vanilla expansion 1.13c character in Normal; any class/level that can survive
a cold hit without dying. Remove cold-duration-reducing gear and record actual gear/stats.
Stand beside a known cold-damage monster, take one cold hit and avoid another. Read the
current frame, state 11 membership, attached list expiry and type 12 firing/removal. Compare
the observed frame to the recorded expiry, accounting for next-frame scheduling clamp.
This can confirm only the captured application and expiry, not every zero-duration branch.

### STL-O02 death

Use a separately prepared Normal character that can acquire the selected timed skill state
and survive until setup is recorded. First establish its named States row stay-death cells
and compiled-mask correspondence. Record class/level, skill level, duration and active state;
die before expiry while capturing list removal, callback requests, death partition and later
expiry. Repeat with a state of opposite stay-death membership. Expected local callback
requests follow STL-V08; exact death partition and later cleanup require their open gates.

### STL-O03 reapplication

Use a Normal character with a timed skill-state applier, recording class/level/skill and its
actual selected state/stat slots. Stand still, apply it, then reapply before expiry. Record
active/changed membership, list expiry, exported stat count, sync eligibility and selected
message. A nonempty export selects state-with-stats, an empty export state-on. Delivery
cannot be inferred from the bit mark alone. Correlate the client callback separately.

### STL-O04 transition

Use a surviving Normal character near an activated waypoint. Record class/level, gear,
cold application, current frame, state 11 and its list/expiry. Travel before expiry, then
immediately read the same fields on arrival and continue through expiry. This is an
investigative observation with no current retention/removal expectation. Repeat only with
controlled comparable inputs if the first capture is incomplete.

## Migration and review record

2026-10-14: preserved the original file privately and verified SHA-256 equality before
rewriting. Repository search found sibling-spec/companion/gap references but no implementation
plan outside gaps, code or test consumer. Existing execution-ordered identifiers are retained
without permanent/alias boilerplate. Binary layouts and detailed history moved here or to
the private audit. The nine numeric fixture scenarios were retained with supplied-input
boundaries and explicit attribution of gated stored effects; new fixtures cover countdown
restart, forced changed membership and busy timer cancellation.

Initial lint:124 errors,1 warning (description, binary details, missing companion/gap shape
and legacy wording). The semantic draft lint had 0 errors and 39 warnings; spacing, method wording and current-contract wording were corrected. The final recorded lint below is clean.
Fresh blind-reader review of the migrated contract is recorded below; an earlier review is not
claimed for changed gates or acceptance. Open evidence is in the [gap file](../../plans/gaps/state-timed-state-lifecycle.md).

### Completed original-image partition check, 2026-10-14

D2Common+0x34150 was independently decoded from the hash-matched original image for all 384 bytes. Entry guards, player/monster/boss mask selection, changed-before-active writes and all exits were read. Its only behavioral callee is update marking+0x665C0, consumed through RLI-Q02. STL-R13b is now verified for supplied masks; STL-V14 discriminates the three selections and changed-bit preservation. Definition-to-mask production remains STL-R12b, so this adds no archive/compiled-table claim. The disguise+0x38730 check decoded 288/288 bytes but its remaining-mask helper is still unread; only decoding uncertainty closed there.

The old death observation's unconditional expectation of no later state-off message is not an accepted assertion: sync eligibility and clearing remain STL-Q01. Current STL-O02 records only the supported callback/partition boundary and names that integration gate.

### Self-review and checks, 2026-10-14

Self-review checked ownership/dependency boundaries, retained numeric scenarios, no-draw declarations, vector assertions and explicit gates. STL-V15 additionally distinguishes the raw StayDeath mask return from normalized 1. No implementation plan, code or test consumer was found; identifiers were not renumbered. No sibling identifier update is required.

D2Game thunk+0xA2EC resolves through import slot+0xF8BB8 to Fog#10046/+0x1CCC0 (LOD113C-FOG), the pool-release boundary used by timer cancellation. This identifies the external memory-release operation; it does not claim a fresh Fog allocator verification.

Checks run: `check_spec.py state-timed-state-lifecycle` (initial 124 errors/1 warning; final 0 errors/0 warnings); `git diff --check` scoped to the three owned files (exit 0). The linter was rerun after vector and record edits. Original-image decode checks covered D2Common+0x38730 (288/288 bytes) and +0x34150 (384/384 bytes). No gameplay tests, captures or archive-cell reads were run. At that self-review point no fresh reader had run; the first full pass and its outcomes are recorded below. The revised contract requires another full pass.


### Full blind-reader pass 1, 2026-10-14

Reported counts: MISMATCH 1, GUESS 0, GAP 0, DISCRIM 1, COVERAGE 15, HARDEST 3, GATED 47. The report's COVERAGE list enumerates 16 grouped bullets despite its reported count of 15; both are preserved here without inventing an extra finding. GATED is reader data: supplied-input fixtures remain verified at their named boundaries, and only the spec's explicit per-assertion gates apply.

MISMATCH: STL-V07(c) described a cold application with no draw. CSA-R05/CSA-R06 and CSA-R06 require the defender shatter draw when an actual positive-length cold application reaches that path, including refresh. Fixed by making the case an isolated strict-extension comparison with supplied expiry values 1200/1050, exact false comparison and selected expiry 1200. Full cold invocation/draw coverage remains CSA-V08. Skill cases (a)/(b) remain supplied timed-skill applications; no dependency draw was erased from an actual cold invocation.

DISCRIM: STL-V01(a) stored frame 1001 but did not exclude an additional synchronous call at frame 1000. Fixed with an initially empty handler log and exact outputs: one queued event at 1001 and zero handler invocations immediately after scheduling at 1000.

| COVERAGE group as reported | Outcome |
|---|---|
| R01 immediate, negative type, wrapped addition/indexing | No change: immediate entry is the existing routing statement; cross-queue execution order remains gated R02a/Q07. Negative/wrapped indexing is explicitly Q10, so no expected value is invented. V01 covers the accepted ordinary timed-state scheduling domain |
| R02 override, missile handler, other nonmatching events | Added V16(a)/(b) for distinct handler selection. No change for exact-frame mismatch: V02 already discriminates equal-slot nonmatching frame, and the equality test is independent of handler identity |
| R03 item expiry and non-12 monster bypass | Added V16(c) item routing. No change for non-12 bypass: the listed bypass fact does not add those handlers' bodies to timed-state acceptance; type-12 routing is V02/V16 |
| R04 missing unit/base/storage and extended due list | Added V17(a)-(c) guards. No change for retained extended-list nontermination: Q05 leaves its producer outside-boundary; V03/V11 already discriminate mandatory current-head reload, and no finite termination assertion is added |
| R06a null operations, primary/secondary lookup, initialization | Added V17(d) null setters/getter. No change for lookup: RLI-V05(a) discriminates primary before secondary. V05 already asserts all remaining semantic initialization fields zero/empty |
| R07 off, negative identity, disguise | Added off and negative-identity steps to V06. No change for disguise: R07a/Q07 remains explicit and no stored disguise value is guessed |
| R08 invalid identity, missing unit/storage | Added V17(e)-(i) refusal cases; V12 retains valid forced-write/update behavior |
| R09 freeze/stun/poison/burn policies | No change: these are consumed applier policies, discriminated by CSA-V11 (freeze), CSA-V05/CSA-V07 (stun) and CSA-V12/CSA-V13/CSA-V14 (poison/burn). Duplicating complete appliers here would duplicate ownership |
| R10a invalid selectors, allocation failure, absent caster, player pet | Added V18(a)-(d), distinguishing refusal, failure prefix, owner sentinel and player-only final request |
| R06b no attach vector | No change: complete attachment remains Q07, with no invented stored effects |
| R07a no disguise vector | No change: full disguise helper/producer remains Q07 |
| R10b skill-stat branches | No change: caller decoding and complete write/refresh path remain Q07; no executable integrated default is added |
| R11 null assertion/additional death condition | Added V08(f)/(g), distinguishing living mode with forced-dead condition and fatal-before-callback order |
| R13a nonplayer/exempt/extended/boss/auxiliary | Added V19 for nonplayer request order, exemption, boss argument and optional auxiliary request. No change for retained extended-list nontermination: Q05 and head-restart contract remain explicit; auxiliary stored effects stay gated |
| R13b arrays/storage, zero count, later words | Added V14(d)-(g), distinguishing skipped update, zero-count update and word-5 writes |
| R12b/R14/R15/R16/R17 gated operations | No change: existing Q01/Q02/Q03/Q07 retain mask producer, message/client, transition and auxiliary gates; no unverified fixture outputs are invented |

HARDEST entries, all no change: V03/V11 head restarts recompute the documented visit/free logs; V14's supplied masks preserve changed bits before active masking; V07(a)/(b) retain distinct existing/new request prefixes. Added guard cases use the same established operations and arithmetic. No ownership, rule evidence label or existing gate is promoted by the review changes. The modified assertions and V07(c) execution boundary require a fresh full reader pass, coordinated by the lead. Review completion is not claimed.

### Full blind-reader pass 2, 2026-10-14

Fresh reader recomputed all V01–V19: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, COVERAGE 8 grouped findings, HARDEST 3. GATED was reported as 22 rules / 10 questions; only the spec's explicit assertion-level gates apply. The reader specifically checked that V07(c) is an isolated supplied-expiry comparison and that the owning CSA-V08(d) covers equality.

| Advisory COVERAGE group | Outcome |
|---|---|
| Immediate frame -1 | No change: the same routing advisory was triaged in pass 1; immediate queue ordering remains R02a/Q07, and V01 covers ordinary timed-state scheduling. No new missing behavioral decision was reported |
| Monster bypass types 3/4/5/8 | No change: R03 states the bypass; these other handlers' bodies are outside timed-state acceptance. V02/V16 cover the owned type-12 routing |
| Secondary exclusion / due state-storage lists | No change: R04 explicitly traverses primary lists and documents retained-list nontermination; pass-1 triage preserves Q05's producer limit and head-restart assertions rather than invent a terminating fixture |
| State setting without storage | No change: consumed RLI-R03 and RLI-V05(c) specify the no-storage write/update requests; duplicating the owning primitive is unnecessary |
| Attach, disguise, skill slots, mask producer | No change: existing R06b/R07a/R10b/R12b and Q07 retain their gates; no integrated output is inferred |
| Cold removal callback | No change: consumed CSA-R13 owns its distinct sequence; this pass does not duplicate that dependency's callback implementation |
| Death traversal with state storage | No change: the retained-list/head-restart limitation and producer boundary are explicit, as triaged in pass 1; no complete lifecycle result is claimed by supplied fixtures |
| Message/client/transition/auxiliary lifetime | No change: Q01/Q02/Q03/Q07 retain these integration/behavior limits; no fixture claims those full paths |

HARDEST, all no change: V03/V11 have explicit head-restart visit logs; V14 preserves changed membership before masking active bits; V07(a)/(b) state the differing existing/new prefixes. No ambiguity or numeric discrepancy was reported. This pass completes the applicable blind-reader review; it does not approve the spec, resolve its evidence gates or establish gameplay confirmation. Subsequent edits in this record and the provenance summary change review reporting only.

### Independent review spot-check, 2026-10-15

A group review re-derived every rule an executable vector depends on from the original 1.13c disassembly (and reconstruction checked against original bytes where it exists), independently of the vector traces, and recomputed STL-V01 through STL-V19. Readings: scheduler D2Game+0x8EA80 (signed `jge 15` type guard, frame -1 immediate path, signed `jg` clamp to frame+1, signed remainder by 64, newest event linked at slot head with a set-if-empty oldest pointer) and link +0x8E680 (absent target type 6/id -1); group dispatch +0x8F070 (missile, player, monster, object, item; immediate then ordinary; consistent with the STL-R02a candidate, which stays unresolved because its region is still an unchecked `db` decoding); player dispatcher +0x8ECC0 (exact frame equality, override else type table, free afterward); missile dispatcher +0x8E9A0 (single handler +0x3E3F0 regardless of type); monster gate +0x769A0 with original-image byte table at +0x76A00 = 00 00 00 01 01 00 01 01 01 00 (types 3,4,5,8,12 bypass); original-image type tables player/monster/item/object slot 12 = +0x77820/+0x74870/+0xD1780/null and object slot 6 = +0x6E470; group map 0,1,3,2,4 for unit types 0..4. D2Common +0x3ADD0, +0x38100/+0x38060/+0x37D70/+0x38020/+0x38380/+0x38000, +0x34510/+0x38730, +0x34450, +0x33AB0, +0x34150 (byte offsets of all exits recomputed: null unit exits without update; missing storage or array exits without update; zero words still update), +0x3AD60 (owner type 4 and condition mask 0x181 retained, StayDeath, state-storage restart). D2Game +0x29240, +0x9F7A0 (thunks resolve +0xA59E to #10819 and +0xAEB6 to #10418), +0x37110, +0x78230, +0x75350. All vectors recomputed with no mismatch except STL-V19(b).

Correction: STL-R13a stated that the nonplayer death cleanup passes a false boss condition for unit types other than monster, and STL-V19(b) asserted log[partition(false)] for an object. The original code of D2Game+0x75350 loads the monster-definition pointer only for unit type 1 with monster data; every other unit (and a null unit) continues with a null pointer and reads its flags word at offset 8, which faults before #11004 or #10627 are called. Its recorded callers +0x75460 and +0x76270 are monster death paths. STL-R13a now names monster death and the non-monster fault; STL-V19(b) uses a nonboss monster with the same expected log, and new STL-V19(c) asserts the object fault with no list or partition request. The review also noted that player death +0x78230 calls D2Common#10084/+0x365A0 with (unit, 1) between the partition and the type 8/9 event removals; STL-R13a now states that the ordered logs list only lifecycle requests. The `modeBound` list-condition name was aligned with foundations' `removeOnModeChange`.

These are vector and rule-text changes, so the completed reader review no longer covers the current contract: a fresh full blind-implementer pass (run 3) is required. Not run in this review.

### Targeted readings for review-pass triage, 2026-10-15

Before any behavioral edit from the reader passes below, the original 1.13c disassembly was read directly (listing of the hash-matched images). D2Game+0x75350 (monster death): a present auxiliary effect (unit+0xA4) is released, cleared and followed by one further call before the type test; only then is the monster-definition pointer loaded for unit type 1 with unit data, else null, and its flags word read. A non-monster with an auxiliary effect therefore requests auxiliary removal and then faults; a null unit skips the auxiliary step and faults. Its boss argument to #10627 is the result of a call taking the unit's class-bounds-checked MonStats record and the unit. D2Common+0x3AD60 (non-staying traversal): null unit, missing base or base without the state-storage sign bit returns; per list, owner type 4, then flags mask 0x181, then #11061(unit, list state) retain and continue at the saved successor; otherwise free unless the list's sign bit is set, then reload the head in both cases. D2Common+0x34150: a nonzero boss selector is tested first and selects the boss mask (DataTables+0x108) for any unit; otherwise type 0 selects +0x100 and every other unit +0x104; word count is (stateCount+31)/32 signed. D2Game+0x8E9A0 (missile) and +0x8EBC0 (monster): an override is invoked when present instead of the no-override handler (missile single handler +0x3E3F0; monster gate +0x769A0); the fired record is released after the call. D2Common+0x38020 and +0x38380: null list ignored; positive expiry ORs timed; any value is stored. D2Common+0x33AB0: returns mask word floor(s/32) AND the bit-table entry for s mod 32; type 1 selects +0x104, every other unit including null +0x100. D2Common+0x3ADD0: signed `jg` expiry comparison; the countdown decrement is stored before its `jg 0` test. D2Game+0x9F7A0: the aliveExclusionCondition byte is tested before the unit type; player dead modes 0/17, monster 0/12, other types dead; pet request only for type 0. D2Game+0x29240: skill identity signed 0..count-1 and record presence, then aurastate signed 0..stateCount-1, then duration evaluation, then lookup; attach result ignored; pet request for a non-null type 0 target regardless of caster. No value was taken from memory or from a reader report.

### Full blind-reader pass 3, 2026-10-15

Counts: MISMATCH 0, GUESS 11, GAP 13, DISCRIM 11, GATED 35, COVERAGE 24, HARDEST 3.

| Entry | Outcome |
|---|---|
| GUESS V06 changed-clear setup | Spec edit: V06 states a direct fixture write with no update request |
| GUESS V07/V18/R10a log omits update marks | Spec edit: 8.1 preamble excludes SetState/ForceChanged internal update marks from ordered logs and names the vectors that count them |
| GUESS R13a hasStateStorage restart | Spec edit after reading +0x3AD60: free skipped, head restart in both cases, loop gated STL-Q05 |
| GUESS R13a non-monster with auxiliary | Spec edit after reading +0x75350: auxiliary removal then fault; V19(d) added |
| GUESS R13a monster without data | Spec edit: monster death completes only for a monster with loaded monster data |
| GUESS R13a boss source | Spec edit: boss condition of the monster's MonStats record (GDA-R08a), supplied by vectors |
| GUESS R13b word count | Spec edit: ceil(stateCount/32) words, state s is bit s mod 32 of word floor(s/32) |
| GUESS R13b player with boss selector | Spec edit after reading +0x34150: boss selector tested first for any unit; V14(h) added |
| GUESS R02 missile override | Spec edit after reading +0x8E9A0: override wins in every group; V16(d)/(e) added |
| GUESS R06a null expiry write | Spec edit after reading +0x38020/+0x38380: expiry setter ignores null; V17(d) sets expiry |
| GUESS R10a SCX inputs | No change: SCX-R11 owns the evaluator front end; vectors supply the result |
| GAP V03/V04/V11 storage | Spec edit: preamble defaults (base state storage, list hasStateStorage false); R04 names base versus list storage |
| GAP V05 lookup unit | Spec edit: valid unit u stated |
| GAP V09 B state/monster mask | Spec edit: V09 rebuilt with explicit states and both masks |
| GAP V09 argument shapes | Spec edit: write(stat 6,value 0,layer 0) and type 8/9 event removals named |
| GAP V15 packed return | Spec edit: R12a defines the return as 2^(s mod 32) or 0; callers test only zero/nonzero |
| GAP V14 packed words | No change: R13b now defines the word grouping in logical terms |
| GAP section 5 outputs | Spec edit: rows for ExpireLists call, StayDeath return, fatal/fault outcomes |
| GAP R09 prose | Spec edit: scheduling follows each applier contract; freeze schedules even when unchanged |
| GAP R13a state 0 | No change: R12a treats state 0 like any valid state |
| GAP R10a stat-slot granularity | Spec edit: preamble logs the six-slot operation once |
| GAP V07 storage | Spec edit: target with state storage |
| GAP R13a path selection | Spec edit: two operations chosen by callers, not a type test |
| GAP V16 monster bypass | No change: the bypass affects only validity for other types; type 12 routing is covered |
| DISCRIM V05 nonpositive sets timed | Spec edit: untimed L2 with expiry 0 |
| DISCRIM V07(a) keep-minimum | Spec edit: V07(d) existing 900 to 1050 |
| DISCRIM V19 boss versus auxiliary | Spec edit: V19(a) auxiliary present/boss false, (b) absent/true |
| DISCRIM V19 monster skips traversal | Spec edit: V19 adds freeable B and monster-mask C |
| DISCRIM V19(c) fault placement | Spec edit: freeable list and V19(d) auxiliary-present object |
| DISCRIM V09 successor versus head | Spec edit: V09 frees a non-head list and logs visits[A,B,A,C] |
| DISCRIM V09 value/layer swap | No change: both are 0 in the contract, so the alternative is equivalent |
| DISCRIM V08 dead-mode sets | Spec edit: V08(h)-(l) |
| DISCRIM V08 dead monster mask false | Spec edit: V08(k) |
| DISCRIM V15 object mask | Spec edit: V15(f) |
| DISCRIM V16 override for monster/missile | Spec edit: V16(d)/(e) |
| HARDEST R13a/V19 | Resolved by the R13a rewrite and V19 rebuild |
| HARDEST R10a log granularity | Resolved by the preamble log rules |
| HARDEST R04 storage meanings | Resolved by R04 wording and preamble defaults |

COVERAGE: no change beyond branches the DISCRIM edits add (V08 modes, V14 player boss, V19 monster traversal); remaining branches are gated or advisory.

### Full blind-reader pass 4, 2026-10-15

Counts: MISMATCH 0, GUESS 7, GAP 7, DISCRIM 15, GATED 33, COVERAGE 22, HARDEST 3.

| Entry | Outcome |
|---|---|
| GUESS R10a invalid identity | Spec edit after reading +0x29240: signed 0..skillCount-1 and record presence; V18(e) added |
| GUESS R10a evaluation order | Spec edit: duration evaluated after all validity checks |
| GUESS R10a attach return | Spec edit: attach result ignored |
| GUESS R12a words at or above 32 | Spec edit: V15(g)/(h) state 53 |
| GUESS R02 fire-then-remove | No change: R02 already says remove afterward; V16(a) now asserts it |
| GUESS R01 immediate path | No change: only frame -1 is immediate (stated); V01(d) now asserts the ordinary queue |
| GUESS R04 countdown store | Spec edit after reading +0x3ADD0: decrement stored whether or not due |
| GAP V03 C expiry | No change in this pass: R04 states untimed lists are never due (made explicit in pass 5) |
| GAP SCX-R11 versus SCX-R24 | No change: SCX-R11 is the front end supplying (unit, skill, level); SCX-R24 is column referencing inside expressions |
| GAP death-condition naming | Spec edit: section 3 states it is CSA's aliveExclusionCondition (the shared name since 2026-10-15) |
| GAP visit logs not in section 5 | Spec edit: visit-order output row |
| GAP R12a states at or above 32 | Spec edit: V15(g)/(h) |
| GAP V04 frame | No change in this pass (edited in pass 5) |
| GAP R09 no-due-list event | No change: V02/V16 already call ExpireLists with no due list |
| DISCRIM V01(d) | Spec edit: ordinary slot asserted, immediate queue empty |
| DISCRIM V02/V16 remove before handler | Spec edit: V16(a) H observes its event still queued |
| DISCRIM V03 unsigned compare | Spec edit: V17(j) expiry -16 is due |
| DISCRIM V05 state-0 lookup | Spec edit: u has a state-0 list and lookup is still null |
| DISCRIM V07 first-list lookup | Spec edit: V07(b) has an unrelated state-11 list |
| DISCRIM V07(c) cold applier | No change: isolated comparison; cold application owned by CSA-V08 |
| DISCRIM V08 monster mode 17 | Spec edit: V08(m) |
| DISCRIM V08(f) monster condition | Spec edit: V08(n) |
| DISCRIM StayDeath stub returning 1 | Spec edit: preamble requires STL-R12a over supplied masks |
| DISCRIM V14 player boss | Spec edit: V14(h) boss selector 2 |
| DISCRIM V15 word 0 only | Spec edit: V15(g)/(h) |
| DISCRIM V18(a) upper identity bound | Spec edit: V18(e) identity 2 and V18(f) aurastate -1 |
| DISCRIM V18(d) pet with absent caster | Spec edit: V18(g) absent caster, monster target |
| DISCRIM V09/V19 hasStateStorage | No change: the true branch never terminates, so no finite assertion exists; gated STL-Q05 |
| DISCRIM V19 item owner | No change: same traversal; V09 discriminates list owner type from dying-unit type |
| HARDEST R13a, R04, R10a | No change beyond the edits above |

COVERAGE: no change; advisory.

### Full blind-reader pass 5, 2026-10-15

Counts: MISMATCH 0, GUESS 6, GAP 8, DISCRIM 4, COVERAGE 29, HARDEST 3; GATED listed without a count.

| Entry | Outcome |
|---|---|
| GUESS V03 C expiry | Spec edit: C untimed with expiry 1000, asserted not freed |
| GUESS V03 visits at 1080 | Spec edit: visit[A,C] |
| GUESS V04 routing/frame | Spec edit: the event calls ExpireLists(unit,1001); L freed at a positive frame |
| GUESS V02/V16 spy versus real | No change: equivalent with empty chains |
| GUESS V08 refresh arguments | No change: preamble logs refresh operations only; R11 lists monster passive refresh |
| GUESS V01(d) | No change: stated in R01 and asserted in V01(d) |
| GAP V03 C expiry | Spec edit (as above) |
| GAP immediate-queue entry | No change: queue/dispatch details gated STL-R02a/STL-Q07 |
| GAP section 5 omissions | Spec edit: rows for handler/override invocations and expiry getter |
| GAP SCX-R11 | No change: as pass 4 |
| GAP skill count/level | No change: section 3 application inputs; missing-record branch advisory |
| GAP gated models | No change: gated STL-R05/STL-R06c |
| GAP V10 | No change: gated |
| GAP visit definition | Spec edit: preamble defines a visit |
| DISCRIM V15 whole mask word | Spec edit: V15(i)/(j) |
| DISCRIM V16 missile type table | Spec edit: V16(f) missile type 12 |
| DISCRIM V18 pet with caster present | Spec edit: V18(h) |
| DISCRIM V05 negative expiry | Spec edit: -1 writes |
| HARDEST R13a, R04, R10a | No change |

Passes 3-5 were full runs on the revised contract; only vectors and fixture wording changed after pass 5, so the next run was a delta check.

### Delta check (reader run 6), 2026-10-15

All eight listed changes confirmed resolved; changed vectors recomputed. Counts: MISMATCH 0, GUESS 1, GAP 1, DISCRIM 2.

| Entry | Outcome |
|---|---|
| GUESS V04 L on the same chain | Spec edit: L is on a different unit's chain |
| GAP visit wording versus V13 timer entries | Spec edit: a visit covers lists and unit-timer entries |
| DISCRIM V15(g) whole word for states at or above 32 | Spec edit: word 1 holds states 53 and 54 |
| DISCRIM V04 ignoring the expiry-0 write | Spec edit: L freed at frame 1, below its former expiry 5 |

The review's cap of four reader runs was reached with these four fixture edits unchecked by a reader. They change no rule; they change V04's example frame and chain placement, V15(g)'s mask and the visit wording. A further delta check covering V04, V15(g) and the preamble visit wording is required before the blind-implementer pass is complete.

### Delta check (reader run 7), 2026-10-15

Sonnet reader, same file set and restrictions as earlier passes. Changed material: the 8.1 visit definition, STL-V04 and STL-V15(g). All three prior findings confirmed resolved; STL-V04, STL-V15(a)-(j) and the visit-dependent vectors STL-V03, STL-V09, STL-V11, STL-V13, STL-V17 and STL-V19 recomputed with no difference. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1; GATED STL-R05/STL-Q07.

| Entry | Outcome |
|---|---|
| DISCRIM V04 synchronous free at frame 1000 accepted (Outputs did not assert it) | Spec edit: V04 asserts the free log is empty immediately after scheduling at frame 1000. Consistent with STL-R01 (scheduling queues an event and makes no synchronous handler call, already asserted in STL-V01(a)); no rule change |
| Note: expiry-0 write appeared only in the V04 Trace | Spec edit: V04 Inputs name the STL-R06a write of 0 to L |

The V04 assertion is an added acceptance assertion, so a further delta check followed.

### Delta check (reader run 8), 2026-10-15

Sonnet reader, same file set and restrictions, changed material STL-V04 only. Both prior findings confirmed resolved; V04 recomputed with no difference. Counts: MISMATCH 0, GUESS 0, GAP 1, DISCRIM 2; GATED STL-Q07, STL-R05.

| Entry | Outcome |
|---|---|
| GAP V04 frame-1001 dispatch not among cited rules | Spec edit (citation only): V04 header cites frame-1001 routing per STL-R02/STL-R03 as in STL-V02. No input, output or assertion changes, so exempt from a rerun |
| DISCRIM V04 unsigned expiry compare accepted (expiry 0 due either way) | No change: the owning discriminating vector is STL-V17(j), expiry -16 freed at 1050, which rejects an unsigned reading. V04's claim is limited to the expiry-0 write being due at a positive frame |
| DISCRIM V04 strict `<` due test accepted | No change: STL-V03 frees B with expiry 1050 at now 1050, which rejects a strict comparison |

Outcome: the final applicable run (8) returned MISMATCH 0 and no unresolved GUESS, GAP or DISCRIM. The blind-implementer pass for the current contract is complete; COVERAGE entries and documented gates remain as recorded.

### Restructure reference check and reader runs 9-12, 2026-10-15

Trigger: the 2026-10-15 spec split remapped this spec's cross-spec references (NRF to RLI, GDF to GDF/GDA, MMF cold/freeze to CSA, skill-calc-expressions-and-synergies to skill-calc-expressions, the neutral death-condition name to the shared `aliveExclusionCondition`). Every remapped id was checked against the per-spec id maps and the owning specs before run 9; all mapped to the contract the rule needs. Citation-only fixes: section 9 Depends on links each id under its owning spec (RLI and GDA separately); the `removeOnModeChange` row names RLI; section 10 links core-record-layouts-and-ids.

Original-code readings for triage (hash-matched 1.13c listings, read directly; no value taken from a reader report or memory):

- D2Game+0x75350: the boss argument is the result of #10064/+0x1A550 on the class-bounds-checked MonStats record and the unit. The CSA companion identifies #10064 as `recordGate` (CSA-R02: 64 when `boss` is set, else 0). STL-R13a and STL-V19 state selector 0/64; player death passes 0.
- D2Common+0x34150: the changed array is addressed from the active array's storage pointer plus the word count, so the two arrays are present or missing together (STL-R13b wording).
- D2Common+0x3AD60: the owner test is an exact comparison with 4; STL-V19(e) added.
- D2Game+0x29240: the duration evaluator receives the caster argument, the `auralencalc` reference, skill identity and level, before the state lookup; a missing caster is passed as null. Allocation passes list flags 2 (timed only). The attach result is discarded. State-on and forced-changed requests pass 1. The pet request follows passive refresh for a non-null type 0 target on both paths. STL-R10a wording, STL-V07(e) and evaluation logs, STL-V18(i) and level 3 added.
- D2Game+0x769A0: event types 3..12 index a byte table whose entries for 3, 4, 5, 8 and 12 go directly to the type handler; other types first call an imported state test (state 1) and a further check that can drop the event. STL-R03 wording and STL-V16(g) added.
- D2Game+0x8EBC0: in the slot loop only an event whose frame equals the game frame fires, and the record is released after the handler returns on both override and no-override paths (STL-V02 event 986, STL-V16(g)).
- D2Game+0x8EA80 and +0x8E680 (requested by the coordinator for CSA-R11): requested frame -1 takes the immediate-path call and returns. Otherwise +0x8E680 allocates the node and links it at the head of the unit's event chain (node next = old head, old head previous = node, unit head = node); a null unit records type 6/id -1. A type 2 request for a type 1 unit holding state 54 makes one extra call first (CSA-Q02). STL-R01 states head-first linking; STL-V01(f) and (g) added.
- D2Common+0x3AE60 (RLI-R04b) tests base-list flag bit 2, the bit the expiry setter sets as `timed`. Section 3 states `timed` is RLI's `expiryFlag` and `hasStateStorage` is RLI's `usesResolvedStats` (wording only).

| Run | Kind | MISMATCH | GUESS | GAP | DISCRIM |
|---|---|---|---|---|---|
| 9 | delta (remapped references and the vectors that consume them) | 2 | 2 | 0 | 5 |
| 10 | full (after STL-R11 cold ownership and STL-R13a selector edits) | 4 | 6 | 8 | 7 |
| 11 | full (after STL-R03, STL-R10a, STL-R13a, STL-R13b edits and new cases) | 0 | 4 | 10 | 6 |
| 12 | full (after the STL-R01 head-first edit) | 1 | 9 | 7 | 6 |

Run 9 outcomes:

| Entry | Outcome |
|---|---|
| MISMATCH NRF-R09a cited for frames | Spec edit: removed from Depends on; no rule consumes the frame-rate helper |
| MISMATCH RLI-Q02 scope | Spec edit: gate text reads changed-state/update-queue effects |
| GUESS cold callback steps | Spec edit: STL-R11 defers the cold callback to CSA-R13 (no timer cancellation, passive or pet request) |
| GUESS boss raw value | Spec edit after the +0x75350 reading: selector 0/64 through CSA-R02 |
| DISCRIM V07(c) strict comparison | No change: owned by CSA-R05 with CSA-V08 |
| DISCRIM cold branch of STL-R11 | No change: owned by CSA-R13 with CSA-V15 |
| DISCRIM lookup order and range | No change: owned by RLI-R03 with RLI-V05 |
| DISCRIM boss column | Resolved by the recordGate citation; CSA-R02's vector reads `boss` |
| DISCRIM V14 selector 1 for a player | No change: V14(h) shows precedence |
| HARDEST (3) | No change |

Run 10 outcomes:

| Entry | Outcome |
|---|---|
| MISMATCH SCX-R11 versus SCX-R24 | No change: STL calls SCX-R11's skill front end; SCX-R24 covers nested column references |
| MISMATCH CSA-R13 statement about STL-R05 | No change: CSA wording about gated STL-R05 (STL-Q07) |
| MISMATCH hasStateStorage versus usesResolvedStats | Spec edit: naming note in section 3 |
| MISMATCH unused RLI-R02a, NRF-R09a | Spec edit: removed from Depends on |
| GUESS evaluation context | Spec edit: STL-R10a names (caster or null, skill, level) |
| GUESS player partition false | Spec edit: selector 0 |
| GUESS hasStateStorage name | Spec edit: as above |
| GUESS STL-R07 update marking | No change: already stated |
| GUESS nonterminating loop handling | No change: rule text and STL-Q05 |
| GUESS secondary-chain hit in STL-R10a | No change: follows RLI-R03 |
| GAP immediate-path details | No change: gated STL-R02a |
| GAP removeOnModeChange unused locally | No change: consumed by RLI-R04b |
| GAP missing skill record input | Spec edit: section 3 application inputs |
| GAP STL-R06c old value | No change: gated |
| GAP word layout | No change: membership words are the contract |
| GAP stat 6 write | Spec edit: replacing unit-stat write, `hitpoints` by GDF-R02 |
| GAP unit type numbering | Spec edit: RLI section 3.3 |
| GAP monster validity test | Spec edit: STL-R03 |
| DISCRIM evaluation after checks | Spec edit: V18 evaluation log |
| DISCRIM attach result ignored | Spec edit: V18(i) |
| DISCRIM monster bypass | Spec edit: V16(g) |
| DISCRIM one array missing | No change: the arrays are one storage (+0x34150 reading) |
| DISCRIM owner type 4 | Spec edit: V19(e) |
| DISCRIM on=5 | No change: RLI-V05(l) |
| DISCRIM removal after group handler | Spec edit: V16(g) |
| HARDEST (3) | No change |

Run 11 outcomes:

| Entry | Outcome |
|---|---|
| GUESS initial list conditions | Spec edit: timed is the only allocation condition |
| GUESS V07 evaluation | Spec edit: V07 evaluation log |
| GUESS handler arguments | No change: not asserted by this contract |
| GUESS player death type check | No change: operation chosen by its caller (stated) |
| GAP sentinel 6/-1 | Spec edit: engine code |
| GAP evaluation log not in section 5 | Spec edit |
| GAP timed versus expiryFlag | Spec edit after run 12 (wording) |
| GAP V05 gated model, V07(c), word view, monster gate condition, queue hook, V04 supplied list, types 13/14 | No change: gated, deliberate fixture capability, or outside this contract |
| DISCRIM frame <= now | Spec edit: V02 event 986 |
| DISCRIM skill/level swap | Spec edit: V18 level 3 |
| DISCRIM pet request on existing-list path | Spec edit: V07(e) |
| DISCRIM frame -1 | Spec edit: V01(f) |
| DISCRIM V09 value/layer | No change: both 0, already stated |
| DISCRIM on=5 | No change: RLI-V05(l) |
| HARDEST (3) | No change |

Run 12 outcomes:

| Entry | Outcome |
|---|---|
| MISMATCH `timed` versus RLI `expiryFlag` | Spec edit, wording only (+0x3AE60 reading); exempt from a rerun |
| GUESS countdown on untimed lists, negative duration, type 8/9 requests, item/missile dead, lookup without base, target without storage, mask length | No change: already stated in STL-R04, STL-R10a, STL-R13a, STL-R11, or owned by RLI-R03 |
| GUESS immediate firing | No change: gated STL-R02a |
| GUESS STL-R10a `on` value | Open: value 1 confirmed at +0x29240; edit not made after the final run |
| GAP monster gate, handler tables, queue hook, target id, missing record, short masks | No change: outside this contract, fixture capability, or advisory |
| DISCRIM V03, V09/V19 secondary chains | Open: add a due timed secondary list that is neither visited nor freed |
| DISCRIM V07/V18 evaluation before lookup | Open: log the lookup and assert evaluation precedes it |
| DISCRIM V13 successor save | Open: let release clear A's link |
| DISCRIM V06 on=5 | No change: RLI-V05(l) |
| DISCRIM V15/V08 | None found |
| HARDEST (3) | No change |

All executable vectors, including the new cases, recomputed without difference in runs 11 and 12. The four-run cap for this task was reached with the open items above, so the blind-implementer pass for the current contract is not complete. These need spec edits and a further reader run.
