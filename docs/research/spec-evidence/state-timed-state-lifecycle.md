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
The foundations contract's NRF-Q17/NRF-Q20 and status application's integration gaps agree
with this distinction. Compatible lookup/state guards are reused from NRF-R12a rather than
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
| LOD113C-D2COMMON | 0x38520, 0x37F80 | 10871 | State-list lookup | STL-R06a | 2026-10-12 reading; exact guards/range behavior consumed from NRF-R12a |
| LOD113C-D2COMMON | 0x3A7D0 | 10807 | Attach list | STL-R06b | 2026-10-12 decompiler lead; 2026-10-14 disassembly inspected; multiple propagation and extended-list branches remain gated |
| LOD113C-D2COMMON | 0x3A280 | 10261 | Replace list stat | STL-R06c | 2026-10-14 full disassembly: entry search/insert/remove helpers, delta propagation and special send-list handling still need complete contracts |
| LOD113C-D2COMMON | 0x34510, 0x38730 | 10945 | Set active/changed state | STL-R07, STL-R07a | 2026-10-12 bit-write reading; 2026-10-14 original-image decode checks byte-only tail; disguise helper still open |
| LOD113C-D2COMMON | 0x34450 | 11039 | Force changed-state membership | STL-R08 | 2026-10-14 full disassembly, no undecoded behavioral bytes; update mark crosses to NRF-Q17 |
| LOD113C-D2COMMON | 0x665C0 | none | Mark unit for update | STL-R07, STL-R08 | Call identified; room queue effects remain NRF-Q17, not established by bit writes |
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
bit 0 is the neutral expiryDeathCondition; unit+0xC8 bit 3 is the disguise condition. State
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

D2Common+0x34150 was independently decoded from the hash-matched original image for all 384 bytes. Entry guards, player/monster/boss mask selection, changed-before-active writes and all exits were read. Its only behavioral callee is update marking+0x665C0, consumed through NRF-Q17. STL-R13b is now verified for supplied masks; STL-V14 discriminates the three selections and changed-bit preservation. Definition-to-mask production remains STL-R12b, so this adds no archive/compiled-table claim. The disguise+0x38730 check decoded 288/288 bytes but its remaining-mask helper is still unread; only decoding uncertainty closed there.

The old death observation's unconditional expectation of no later state-off message is not an accepted assertion: sync eligibility and clearing remain STL-Q01. Current STL-O02 records only the supported callback/partition boundary and names that integration gate.

### Self-review and checks, 2026-10-14

Self-review checked ownership/dependency boundaries, retained numeric scenarios, no-draw declarations, vector assertions and explicit gates. STL-V15 additionally distinguishes the raw StayDeath mask return from normalized 1. No implementation plan, code or test consumer was found; identifiers were not renumbered. No sibling identifier update is required.

D2Game thunk+0xA2EC resolves through import slot+0xF8BB8 to Fog#10046/+0x1CCC0 (LOD113C-FOG), the pool-release boundary used by timer cancellation. This identifies the external memory-release operation; it does not claim a fresh Fog allocator verification.

Checks run: `check_spec.py state-timed-state-lifecycle` (initial 124 errors/1 warning; final 0 errors/0 warnings); `git diff --check` scoped to the three owned files (exit 0). The linter was rerun after vector and record edits. Original-image decode checks covered D2Common+0x38730 (288/288 bytes) and +0x34150 (384/384 bytes). No gameplay tests, captures or archive-cell reads were run. At that self-review point no fresh reader had run; the first full pass and its outcomes are recorded below. The revised contract requires another full pass.


### Full blind-reader pass 1, 2026-10-14

Reported counts: MISMATCH 1, GUESS 0, GAP 0, DISCRIM 1, COVERAGE 15, HARDEST 3, GATED 47. The report's COVERAGE list enumerates 16 grouped bullets despite its reported count of 15; both are preserved here without inventing an extra finding. GATED is reader data: supplied-input fixtures remain verified at their named boundaries, and only the spec's explicit per-assertion gates apply.

MISMATCH: STL-V07(c) described a cold application with no draw. CSA-R05/CSA-R06 and MMF-R14a require the defender shatter draw when an actual positive-length cold application reaches that path, including refresh. Fixed by making the case an isolated strict-extension comparison with supplied expiry values 1200/1050, exact false comparison and selected expiry 1200. Full cold invocation/draw coverage remains CSA-V08. Skill cases (a)/(b) remain supplied timed-skill applications; no dependency draw was erased from an actual cold invocation.

DISCRIM: STL-V01(a) stored frame 1001 but did not exclude an additional synchronous call at frame 1000. Fixed with an initially empty handler log and exact outputs: one queued event at 1001 and zero handler invocations immediately after scheduling at 1000.

| COVERAGE group as reported | Outcome |
|---|---|
| R01 immediate, negative type, wrapped addition/indexing | No change: immediate entry is the existing routing statement; cross-queue execution order remains gated R02a/Q07. Negative/wrapped indexing is explicitly Q10, so no expected value is invented. V01 covers the accepted ordinary timed-state scheduling domain |
| R02 override, missile handler, other nonmatching events | Added V16(a)/(b) for distinct handler selection. No change for exact-frame mismatch: V02 already discriminates equal-slot nonmatching frame, and the equality test is independent of handler identity |
| R03 item expiry and non-12 monster bypass | Added V16(c) item routing. No change for non-12 bypass: the listed bypass fact does not add those handlers' bodies to timed-state acceptance; type-12 routing is V02/V16 |
| R04 missing unit/base/storage and extended due list | Added V17(a)-(c) guards. No change for retained extended-list nontermination: Q05 leaves its producer outside-boundary; V03/V11 already discriminate mandatory current-head reload, and no finite termination assertion is added |
| R06a null operations, primary/secondary lookup, initialization | Added V17(d) null setters/getter. No change for lookup: NRF-V33(a) discriminates primary before secondary. V05 already asserts all remaining semantic initialization fields zero/empty |
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
| State setting without storage | No change: consumed NRF-R12a and NRF-V33(c) specify the no-storage write/update requests; duplicating the owning primitive is unnecessary |
| Attach, disguise, skill slots, mask producer | No change: existing R06b/R07a/R10b/R12b and Q07 retain their gates; no integrated output is inferred |
| Cold removal callback | No change: consumed MMF-R13c owns its distinct sequence; this pass does not duplicate that dependency's callback implementation |
| Death traversal with state storage | No change: the retained-list/head-restart limitation and producer boundary are explicit, as triaged in pass 1; no complete lifecycle result is claimed by supplied fixtures |
| Message/client/transition/auxiliary lifetime | No change: Q01/Q02/Q03/Q07 retain these integration/behavior limits; no fixture claims those full paths |

HARDEST, all no change: V03/V11 have explicit head-restart visit logs; V14 preserves changed membership before masking active bits; V07(a)/(b) state the differing existing/new prefixes. No ambiguity or numeric discrepancy was reported. This pass completes the applicable blind-reader review; it does not approve the spec, resolve its evidence gates or establish gameplay confirmation. Subsequent edits in this record and the provenance summary change review reporting only.

### Independent review spot-check, 2026-10-15

A group review re-derived every rule an executable vector depends on from the original 1.13c disassembly (and reconstruction checked against original bytes where it exists), independently of the vector traces, and recomputed STL-V01 through STL-V19. Readings: scheduler D2Game+0x8EA80 (signed `jge 15` type guard, frame -1 immediate path, signed `jg` clamp to frame+1, signed remainder by 64, newest event linked at slot head with a set-if-empty oldest pointer) and link +0x8E680 (absent target type 6/id -1); group dispatch +0x8F070 (missile, player, monster, object, item; immediate then ordinary; consistent with the STL-R02a candidate, which stays unresolved because its region is still an unchecked `db` decoding); player dispatcher +0x8ECC0 (exact frame equality, override else type table, free afterward); missile dispatcher +0x8E9A0 (single handler +0x3E3F0 regardless of type); monster gate +0x769A0 with original-image byte table at +0x76A00 = 00 00 00 01 01 00 01 01 01 00 (types 3,4,5,8,12 bypass); original-image type tables player/monster/item/object slot 12 = +0x77820/+0x74870/+0xD1780/null and object slot 6 = +0x6E470; group map 0,1,3,2,4 for unit types 0..4. D2Common +0x3ADD0, +0x38100/+0x38060/+0x37D70/+0x38020/+0x38380/+0x38000, +0x34510/+0x38730, +0x34450, +0x33AB0, +0x34150 (byte offsets of all exits recomputed: null unit exits without update; missing storage or array exits without update; zero words still update), +0x3AD60 (owner type 4 and condition mask 0x181 retained, StayDeath, state-storage restart). D2Game +0x29240, +0x9F7A0 (thunks resolve +0xA59E to #10819 and +0xAEB6 to #10418), +0x37110, +0x78230, +0x75350. All vectors recomputed with no mismatch except STL-V19(b).

Correction: STL-R13a stated that the nonplayer death cleanup passes a false boss condition for unit types other than monster, and STL-V19(b) asserted log[partition(false)] for an object. The original code of D2Game+0x75350 loads the monster-definition pointer only for unit type 1 with monster data; every other unit (and a null unit) continues with a null pointer and reads its flags word at offset 8, which faults before #11004 or #10627 are called. Its recorded callers +0x75460 and +0x76270 are monster death paths. STL-R13a now names monster death and the non-monster fault; STL-V19(b) uses a nonboss monster with the same expected log, and new STL-V19(c) asserts the object fault with no list or partition request. The review also noted that player death +0x78230 calls D2Common#10084/+0x365A0 with (unit, 1) between the partition and the type 8/9 event removals; STL-R13a now states that the ordered logs list only lifecycle requests. The `modeBound` list-condition name was aligned with foundations' `removeOnModeChange`.

These are vector and rule-text changes, so the completed reader review no longer covers the current contract: a fresh full blind-implementer pass (run 3) is required. Not run in this review.
