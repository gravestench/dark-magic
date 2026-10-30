# Evidence: Baal monster skills

Contract: [skill-monster-baal](../../specs/skill-monster-baal.md).
Identity: [LOD113C-D2GAME](../source-manifest.md), Windows x86 LoD Expansion 1.13c.
Read date: 2026-10-29. No gameplay confirmation is claimed; the named text cells have the table-data record below.

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | `0x0d290` | none established | Selector 139 action (analyst description) | BMS-R01, BMS-R02 | Entry listing; full effect path unresolved |
| LOD113C-D2GAME | `0x0cb30` | none established | Selector 140 action (analyst description) | BMS-R02 | Entry listing; full effect path unresolved |
| LOD113C-D2GAME | `0x0ca20` | none established | Selector 141 action (analyst description) | BMS-R04 through BMS-R07 | Request preparation and normal-return continuation; complete original-image decode |
| LOD113C-D2GAME | `0x1B10` | none | Default-radius calculation | BMS-R05 | Complete local instruction reading; no calls |
| LOD113C-D2GAME | `0x9EEB0` | none | Query coordinator and ordered callback loop | BMS-R05, BMS-R07 | Complete image-corresponding listing; real room/collection production remains gated |
| LOD113C-D2COMMON | `0x124E0` | 11106 | Fixed-filter candidate predicate | BMS-R06 | Complete original-image decode; every path reachable with filter 0x3002 traced |
| LOD113C-D2GAME | `0xC8D0` | none | Select recipient and invoke shared explosion | BMS-R07 | Byte-equivalent reconstructed source read before adopting callback effects |
| LOD113C-D2COMMON | `0x30E80` | 11129 | Selected-target setter | BMS-R07 | Byte-equivalent source; complete valid-monster/path branch has no calls |
| LOD113C-D2GAME | `0x4F4D0` | none | Shared corpse-explosion action | DSC-R01a through DSC-R14b | Reused 2026-10-15 evidence; same original artifact and action entry, distinct supplied skill 285 |

## Dispatch and search record

The compatible [dispatch evidence](skill-cast-start-and-do-dispatch.md) identifies
D2Game RVA `0xA19A0` and the do table at RVA `0x1074A8`. On 2026-10-29 the table
slots listed above were independently read from the original image with PE section
mapping and little-endian 32-bit pointers; each equals image base plus its listed
RVA. These establish selector-to-entry links, not archive row bindings or AI callers.
Only root-anchor and entry/output-path reading were completed; no absence claim or
six-rung search completion is made. Existing companion inventory had no contract
for this candidate. Dispatch evidence is reused only for that root.

## Initial static observations and limits (before the continuation)

Entry 139 rejects nonpositive missile selectors, obtains target coordinates, requests a missile, then supplies two direction-related values to imported helpers. Entry 140 advances the actor seed, takes the low result modulo three and adds two and the game difficulty byte to obtain a loop count; each iteration consumes two further draws modulo eighteen and offsets each coordinate by minus nine before a creation attempt. Entry 141 prepares an enumeration request and computes a signed 32-bit quantity from two skill parameters as base plus (level minus one) times increment when the index is valid and level positive, otherwise zero. Its callback and enumeration helper remain unread.

For selector 139 only, a negative signed-index check precedes any call or write. The early return clears the result register. This supports only BMS-R01; record-valid and successful paths are not promoted.

The method was assembly reading after checking matched-C availability. No listed
handler had a matched-C entry. Unchecked raw-byte runs and unread local/imported
callees prevent verification of BMS-R02. Decompiled names are not evidence of
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

BMS-R02 retains the cold/tentacle and actual world-query/shared-effect limits.
BMS-Q02 remains observation pending. The first draft's applicable full pass 4
was complete; the successful corpse-query continuation below adds BMS-R05
through BMS-R07 and has completed its applicable fresh full pass 3 with zero
MISMATCH, GUESS, GAP and DISCRIM findings. Earlier pass records keep their
historical scope. The companion does not supply additional behavioral rules.

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
It supports BMS-R03 only. No arithmetic or executed effect is established.

## Blind reader pass 1: findings and outcomes

Full smaller-model review: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1, GATED 3, HARDEST 1. DISCRIM negative-index vector mentioned an empty lookup log without asserting it: edited V01 to assert an empty logical record-read log, making the clamping alternative fail. HARDEST early return/table distinctions/vector order: no behavior change; boundaries and data limits are explicit. Existing GATED findings remain open. Fresh review required after this assertion change and added positive stage.

## Positive local stage verification, 2026-10-29

Original-image disassembly from the true selector-141 entry D2Game RVA 0xCA20
through 0xCB2F corroborates the entire instruction listing, including embedded
callback pointer at 0xCACC. An earlier diagnostic decoding begun in that pointer
operand was discarded; it is not evidence of instructions. The local arithmetic
at 0xCAF8..0xCB15 supplies zero on invalid lookup/nonpositive level, otherwise
loads signed operands at record offsets 0x158 (base) and 0x15C (increment).
The subsequent call at 0xCB19 targets 0x9EEB0; its return is overwritten by 1 at
0xCB1F. BMS-R04 ends at the logged request and the normal-return continuation.
No callee is omitted inside its arithmetic boundary. Base/ increment column
identities are not inferred from their likely Param names. The callback at
0xC8D0 was read: it invokes import thunk 0xACD0 and local 0x4F4D0 with fixed
skill identity 285. These downstream operations and recipient enumeration remain
unread; their target/effect interpretation is not promoted. Upstream actor
coordinates are outside this supplied calculation stage, not arithmetic gates.

## Handler-specific remaining reading

- BMS-Q03: 0xD290 independently decoded; target 0x9FA30 and coordinate getters 0x1100/0x1130 read; creation 0xA28F0 and imports through 0xA50E/0xAF5E unclosed.
- BMS-Q04: 0xCB30 read: initial seed draw modulo 3 plus difficulty+2 attempts; two modulo 18 draws per attempt offset positions by -9; room lookup 0x1280, creation 0xEEF10, events 0x8EA80 and import 0xAC76 remain unclosed.
- BMS-Q05: 0xCA20 full original decode; 0x9EEB0 enumeration unread; callback 0xC8D0 read and calls import 0xACD0 and local 0x4F4D0 with fixed skill 285; these downstream operations remain unread. BMS-R04 quantity arithmetic and normal-return continuation are independent of their effects.

Schema reconciliation: compatible GDA-R11a evidence maps Param1..8 to +0x148..+0x164 in steps of four, establishing +0x158=Param5 and +0x15C=Param6. Therefore BMS-R04 uses those names. Rechecking the same preserved complete effective Skills capture by named headers yields Baal Corpse Explode.Param5 raw 40, Param6 raw empty string; no numeric value is inferred for the empty cell or live compiled field. The supplied base 7 / increment -3 vector is synthetic, not an original record-value assertion.

Further boundary clarification: the computed quantity is now named enumeration search parameter, not recipient count. The read entry of 0x9EEB0 loads request+0x14 and substitutes helper 0x1B10 result when zero, so a stronger range/count interpretation remains unsupported. Callback 0xC8D0 byte-corresponding source view was read on this followup and agrees with the two-call sequence; the prior assembly-first reading of this callback did not follow source order and contributes no independent promotion. Matching method coverage is limited to the supplied local search-parameter stage; callback effect semantics remain gated.

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
| 285 | 0 | 141 |
| 315 | 0 | 140 |
| 318 | 0 | 139 |

These observations corroborate BMS-R03's named selector bindings against the
effective text capture. They close only that selector correspondence question.
They do not establish parameter/expression/state/stat-field correspondence,
gameplay execution or whole compiled-table identity. Original specimen memory
locations explain collection only; the contract requires semantic values.
Private source-profile and collection records remain separately preserved.

## Blind reader full pass 2: findings and outcomes

Fresh full smaller-model review: MISMATCH 0, GUESS 0, GAP 0; raw DISCRIM
heading count 2, with both entries explicitly reporting passing discrimination
(V01 rejects clamping; V02 rejects wrong level, negative clamp and forwarded
return). Normalized actionable DISCRIM deficiencies: 0. The raw report is not
treated as two failures or silently replaced. COVERAGE invalid lookup and
nonpositive level: added independent V03/V04 asserting parameter zero while
still issuing the request and returning 1 after normal continuation. Runtime
and downstream effect coverage remain explicitly gated, with no change.
HARDEST search parameter versus recipient count: no change; the rule names
only a supplied request parameter and gates recipient effects. GATED 5
(Q01-Q05) retain their stated boundaries. The added vectors require a fresh
full review; its completion is not claimed in this record.

## Blind reader full pass 3: findings and outcomes

Fresh full smaller-model review: all numeric outputs recomputed correctly,
with MISMATCH 1 identifying a prose scope contradiction, GUESS 0, GAP 0,
DISCRIM 0, GATED 5, COVERAGE 1, HARDEST 2. MISMATCH: section 6's blanket
missing-record/level-zero gate contradicted the already verified supplied-input
BMS-R04/V03/V04 behavior. Section 6 now explicitly qualifies the remaining
limits as full integration and preserves the established supplied-stage zero
parameter and normal-return result. No existing rule, gate or execution boundary
changes. COVERAGE overflow: added BMS-V05 (valid lookup, level 2, base
2147483647, increment 1) asserting signed wrap to -2147483648 and final result
1, discriminating saturation and unbounded arithmetic. HARDEST supplied-stage
versus real data binding and normal-return continuation: no further change;
the revised section 6 and existing supplied-operation precondition state these
distinctions. GATED Q01-Q05 remain explicit. The next fresh delta review covers
section 6's qualification and new V05, with corresponding Id map and acceptance
vector references; no other behavior changed.

## Blind reader full pass 4: final applicable review

A fresh full review was used, rather than the delta proposed after pass 3,
because that pass had recorded a scope MISMATCH. Result: MISMATCH 0, GUESS 0,
GAP 0, DISCRIM 0, COVERAGE 0, GATED 6 grouped findings (R02 and Q01-Q05),
HARDEST 2. The reader independently recomputed V01 through V05 as rejection
result 0 and enumeration search parameters -2, 0, 0 and -2147483648; each
normally returning enumeration continuation ignores the supplied return 0 and
returns 1. HARDEST signed arithmetic/normal-return continuation and full
integration versus supplied-stage distinction: no change, the rule and section
6 already state both. No changes recommended. Explicit gates remain. The
final applicable review meets the stopping rule and does not approve the spec
or establish gameplay confirmation.

## Corpse-action successful-path continuation, 2026-10-29

### Reuse and exact boundary

The existing [shared explosion evidence](skill-assassin-death-sentry-corpse-damage.md#identity-reuse-and-method)
was checked for original artifact, entry and argument compatibility. Its original
2026-10-15 complete reading of D2Game 0x4F4D0 applies to the same
`LOD113C-D2GAME` fingerprint. The action's skill identity/level and caster are
inputs, rather than a hard-coded trap class. This callback supplies actor and
skill 285 at that same entry; it does not reuse the trap's AI or corpse-selection
contract. DSC's target, expression, state, spatial, scale and event producer
limits remain in force. The new readings establish the caller and target write,
not new verification of every line of the shared action.

Original D2Common SHA-256 was rechecked against `LOD113C-D2COMMON`:
`59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10`.
The current original D2Game identity is the existing `LOD113C-D2GAME` record.
Source availability was checked before new instruction reading. Small
byte-equivalent source views were read for D2Game 0xC8D0 and D2Common 0x30E80,
0x34BB0, 0x34B80, 0x12030 and 0x12000. Other credited local paths use instruction
readings and original-image decoding. Query construction/allocation and world
collection are not promoted merely because their entry bodies were read.

### Request, fallback and candidate chain

1. Action table 0x1074A8 slot 141 selects 0xCA20, which prepares fixed default
   identity 285, filter 0x3002, callback 0xC8D0, actor, invocation level and raw
   radius override. Its call to 0x9EEB0 is at 0xCB19 and the normal-return
   overwrite to 1 is at 0xCB1F. For a monster, the guarded coordinate calls are
   D2Game 0xA454 -> D2Common #10867/0x34BB0 (X) and 0xA3B8 ->
   #10750/0x34B80 (Y). Their valid-path reads zero-extend the integer words;
   absent dynamic path in the caller supplies zero without invoking the fatal
   null-path getter. This is compatible with NRF-R10b.
2. Coordinator 0x9EEB0 reads the radius override. Only zero calls 0x1B10 with
   the request's fixed identity 285 and invocation level. That helper validates
   the Skills index/record and positive level, then uses record +0x148/+0x14C
   (GDA-R11a Param1/Param2) in wrapping i32 arithmetic. Nonzero overrides make
   no such call. This establishes BMS-R05; original compiled values remain Q01.
3. The coordinator obtains the actor room through D2Common #10331/0x2FE10,
   searches from it through D2Game 0x1280, then constructs the query through
   #10591/0x12030 and collects through #10052/0x12720. These are resolved
   original entries, not evidence of complete real room eligibility/allocation.
   The query supplies no custom filter callback, so collection reaches the
   internal default predicate 0x124E0 with the descriptor carrying filter 0x3002.
4. Original D2Common 0x124E0..0x126FF was decoded independently. Its kind table
   at 0x12704 has targets 0x12582, 0x125B9, 0x125F0, 0x125F8 and 0x12623
   for kinds 0..4. The table was read as little-endian pointers, not interpreted
   as instructions. For filter 0x3002 only the monster branch can accept, and
   its dead-only condition requires mode 12. The signed wrapped squared-distance
   comparison precedes the branch. The count-limit, special monster, selectable,
   targetable, town, collision and custom-filter conditions are all disabled by
   this fixed filter. No helper call is reached on its admitted monster branch.
   Dynamic positions are zero-extended integer words or zero on absent path.
   The helper's accepted-query tally is internal query bookkeeping; no unit
   write follows. BMS-R06 specifies the candidate result, not a new storage API.

The query also enables a separate room-exclusion condition in 0x12720. Its
room chains/overlap checks, 0x2D320 condition, backing-array growth and actual
allocation effects are not closed by the predicate reading. A geometric
predicate vector does not prove the collector visits that candidate.

### Ordered callbacks and selected-target effect

After collection, 0x9EEB0 at 0x9EF9A..0x9EFC3 walks the returned array in
index order. Each callback receives game, actor, recipient, invocation level
and extra value 0. It does not branch on the callback return. At 0x9EFCA it
requests D2Common #10344/0x12000 collection release, including when the
returned count is zero. The release helper's allocator-produced effects stay
outside the supplied collection-stage assertion.

Callback 0xC8D0 first calls D2Game thunk 0xACD0, resolved through IAT slot
0x6FD18874 to D2Common #11129/0x30E80. With a present monster actor and readable
path, that setter has a complete no-call branch: it replaces path target and
cached target kind/identity with the non-null recipient. Its actor-null,
unsupported-kind assertion paths and malformed-path behavior are not the
admitted callback fixture. The callback then directly calls 0x4F4D0 with fixed
skill 285 and the invocation level. It makes no surrounding skill-dispatch
call and does not restore a previous target. The shared action may make its
own target/state changes; BMS does not erase them or claim that they always
leave the selected target unchanged. This establishes BMS-R07's actual target
write plus ordered hand-off, preserving all DSC integration gates.

The target-field interpretation was also cross-checked against the small
byte-equivalent getters D2Common #10392/0x348F0 and #11163/0x34910: they read the same target
reference and its cached two identity fields. Their descriptive labels suggest
rooms/world coordinates, but those labels were not adopted. The compatible
target-normalization reading in the healing companion confirms the kind/id
use; no normalization operation is added to this callback.

### Remaining producer and observation limits

The bounded immediate search for skill constant 285 found a request in D2Game
0xA8430 and the callback/request entries above. The separate 0xB6D30 hit uses
285 as an assertion line, not a skill request, and was discarded. The 0xA8430
actor binding and complete think eligibility are not established. A subsequent
original table read at 0x10E788, row 134 (16-byte rows), gives flag 2,
initializer 0xA3880, think entry 0xA8430 and null auxiliary entry. The ordinary
branch of row getter 0x9D0F0 reads a signed monster-record selector, accepts
0..147 and otherwise uses row 0. These are narrowed original dispatch leads;
the upstream think caller and original named monster-record binding are not
closed. This remains a lead for a naturally reachable observation, not a
verified complete AI policy.
No absence claim or complete six-rung AI search is made.

No new gameplay invocation, memory write or controlled-call observation was
performed. BMS-V06 through BMS-V08 use explicitly supplied prepared records,
candidate units or a stable collected list. Actual original compiled operands,
room production, corpse eligibility and explosion effects remain the named
Q01/Q05 and DSC gates. The continuation requires a new full blind reader after
mechanical lint and self-review; no earlier review approves these additions.

## Corpse-action continuation blind full pass 1

Fresh smaller-model full review returned MISMATCH 0, GUESS 0, GAP 0,
DISCRIM 0, GATED 5 (Q01-Q05), COVERAGE 3 and HARDEST 0. Its initial case
and gate totals were corrected by the reader to 21 cases and five questions;
those corrected totals are the report used here. No missing behavioral choice
or arithmetic mismatch was found.

Each advisory coverage item was addressed: V06(h) now explicitly tests a
synthetic malformed in-range absent default record, without a normal-loader
reachability claim; V07(h) adds wrapped subtraction with an explicitly
synthetic supplied center; V09 supplies action-produced target changes and
asserts both the next replacement and the final un-restored target. These
additions change executable assertions and require a fresh full review.
The reader supplied no HARDEST entry; no difficulty finding is invented on
its behalf. Actual room/record/shared-effect gates remain unchanged.

## Corpse-action continuation blind full pass 2

Fresh full smaller-model review returned MISMATCH 0, GUESS 0, GAP 0,
DISCRIM 0, GATED 10 referenced questions (BMS-Q01-Q05 and DSC-Q01-Q05),
COVERAGE 2 and HARDEST 3. All current arithmetic, callback arguments and
target-state effects recomputed correctly.

COVERAGE multiplication widening: no change. In BMS-R04 the product has no
separate comparison or output before the final wrapping sum. Algebraically,
low32(base + low32(a*b)) equals low32(base + a*b). Widening that product and
performing the final required low32 is therefore an equivalent implementation;
no discriminator should reject it. Unbounded final arithmetic and saturation
are already distinguished by V05. COVERAGE radius-square wrap: added V07(i),
where zero distance is rejected because radius 50000 squares to signed
-1794967296; wide/saturating radius square incorrectly accepts.

HARDEST wrapping/normal continuation, fixed predicate geometry versus actual
candidate production, and ordered callbacks/target effects versus DSC effects:
no rule change; their inputs, exact arithmetic and gates already state these
boundaries. The added executable case requires another fresh full review.

The original export directory additionally identifies the fixed-filter
predicate at D2Common 0x124E0 as export 11106. This corrects its initial
internal-only locator; no function semantics or rule boundary changes.

## Corpse-action continuation blind full pass 3: final applicable review

Fresh full smaller-model review returned MISMATCH 0, GUESS 0, GAP 0 and
DISCRIM 0. Its raw GATED count was four groups of documented limitations;
the local Q01-Q05 and DSC producer/effect gates remain as explicitly listed,
not removed by a group count. Its coverage comments concern actual room/
candidate collection and DSC effects (already gated), and not combining the
absent actor-path case with every independent radius arithmetic case. No
uncovered result-distinct local branch or additional fix was identified:
the position fallback and radius calculation have no cross-condition branch.

HARDEST actual collection/release plus shared-action integration, original
compiled default-operand correspondence, and owned-game observation: no change,
each is already an exact named integration or observation gate. The reader
recomputed V01 through V09, including the signed radius-square overflow and
supplied action's target mutations. The applicable review meets the stopping
rule. No contract/vector/gate change follows this final review record; the
spec remains draft and no gameplay confirmation or approval is inferred.
