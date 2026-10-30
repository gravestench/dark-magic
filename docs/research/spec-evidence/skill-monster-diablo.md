# Evidence: Diablo monster skills

Contract: [skill-monster-diablo](../../specs/skill-monster-diablo.md).
Identity: [LOD113C-D2GAME](../source-manifest.md), Windows x86 LoD Expansion 1.13c.
Read date: 2026-10-29. No gameplay confirmation is claimed; the named text cells have the table-data record below.

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | `0xb3d60` | none established | Selector 100 action (analyst description) | DMB-R01, DMB-R04..DMB-R07 | Complete local caller; dependency effects bounded below |
| LOD113C-D2GAME | `0xb3f10` | none established | Selector 103 action (analyst description) | DMB-R01, DMB-R02 | Entry listing; full effect path unresolved |
| LOD113C-D2GAME | `0xb4410` | none established | Selector 104 action (analyst description) | DMB-R01, DMB-R02 | Entry listing; full effect path unresolved |
| LOD113C-D2GAME | `0xb6840` | none established | Selector 152 action (analyst description) | DMB-R01, DMB-R02 | Entry listing; full effect path unresolved |

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

The four entries check the signed skill index before calling another function. The combat-shaped entries then prepare effect records and invoke damage helpers. The object-shaped entry branches on an object flag and object class fields before its damage path. The final entry validates a signed missile selector, constructs a missile request, and branches on creation success. These are instruction observations, not established semantic identities.

For each selector 100, 103, 104 and 152, its negative signed-index branch precedes any call, skill-record access or actor-state write and reaches a result-zero return. Original-image decoding on 2026-10-29 independently corroborated the four entry spans: selector 100 at RVA 0xB3D60 has signed-negative branch 0xB3D74 to 0xB3D95 and zero-return sequence 0xB3D95..0xB3D9E; selector 103 at 0xB3F10 branches 0xB3F28 to 0xB3F4B with zero-return sequence 0xB3F4B..0xB3F58; selector 104 at 0xB4410 branches 0xB441D to 0xB4440 with zero-return sequence 0xB4440..0xB4447; selector 152 at 0xB6840 branches 0xB6851 to 0xB6880 with zero-return sequence 0xB6880..0xB6889. These support DMB-R01 for all four selectors; they do not promote successful action effects.

The method was assembly reading after checking matched-C availability. No listed
handler had a matched-C entry. Unchecked raw-byte runs and unread local/imported
callees prevent verification of DMB-R02. Decompiled names are not evidence of
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

DMB-R02: unresolved; raw-byte decoding, local/imported callees, data and integration
remain open. DMB-Q02: observation pending. The initial draft's blind delta pass 4 and the cold-action continuation's fresh full pass 1 are complete with zero MISMATCH, GUESS, GAP and DISCRIM findings; dated pass records below retain their scope. Explicit evidence gates remain open. The companion does not supply additional behavioral rules.

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
It supports DMB-R03 only. No arithmetic or executed effect is established.

The same complete text capture supplies DiabLight.srvstfunc raw decimal 53 and DiabRun.srvstfunc raw decimal 54. Start handlers have not been identified/read in this pass; DMB-Q01 explicitly includes them. These two cells do not establish successful start execution.

## Blind reader pass 1: findings and outcomes

Full smaller-model review: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1, GATED 3, HARDEST 3. DISCRIM negative-index vector mentioned an empty lookup log without asserting it: edited V01 to assert an empty logical record-read log, making the clamping alternative fail. HARDEST early return/table distinctions/vector order: no behavior change; boundaries and data limits are explicit. Existing GATED findings remain open. Fresh review required after this assertion change and added positive stage.

## Positive local stage verification, 2026-10-29

Original-image disassembly from D2Game RVA 0xB3D60 through 0xB3EAF corroborates
the listing, including raw-byte clear and OR instructions at 0xB3DD9/0xB3DE4.
At 0xB3DD0..0xB3E0F the routine clears 0x1C dwords, writes resultFlags from
record source offset 0x12E OR 1, flags from 0x130 and optional hit class from
0x134. At 0xB3E26 it calls thunk 0xACE2 with actor, skill, level, condition 1;
the returned duration is stored at combat-record offset 0x34. The next fill
call is D2Game 0x9F210. Compatible existing SLD evidence resolves 0xACE2 to
D2Common ordinal 10510 / RVA 0x4E900 (SLD-R11), and the fill to SLD-R15.
The fill listing was reread as a lead and agrees with that existing contract;
its historical dates and limits are retained, not renewed here. DMB-R04 stops
at the fill request and uses supplied duration operation results, retaining
SLD's expression integration gates. Target refresh 0xE07A0 and path-target
thunk 0xA598 precede this supplied boundary; damage continuation 0xDE0C0 and
0xDCF20 follows it and remains unverified here. Neither those unread local
callees nor archive column identities are silently credited.

## Handler-specific remaining reading

- DMB-Q03: Current remaining gate is completed apply/reaction/valid-overlay integration. The continuation below establishes target normalization, caller request order, actual post-apply stat read and local guards. Earlier wording called 0xDE0C0 resolution and 0xDCF20 application; their compatible shared contracts identify them as apply and combat-result reaction respectively.
- DMB-Q04: 0xB3F10 reads used-skill and object/path fields; imports through 0xAABA, 0xA79C, 0xA78A, 0xACE8 and combat 0xA11B0, 0x9F1A0, 0x9F210, 0xDE0C0, 0xDCF20 remain unclosed.
- DMB-Q05: 0xB4410 calls 0x9EE50, 0xE07A0, 0xA598, room/state thunks 0xA478/0xA808 and creation 0xEFB80; handler read, callees unclosed.
- DMB-Q06: 0xB6840 calls target 0x9FA30 and creation 0x6F930, then 0xB32E0/0xB3270, 0x8E6F0/0x8EA80 and 0xB6720; raw tail not independently decoded in full.
- DMB-Q07: Text cells identify 53/54 on DiabLight/DiabRun; start dispatch slots and bodies not yet read.

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
| 193 | 53 | 152 |
| 194 | 0 | 100 |
| 198 | 54 | 103 |
| 199 | 0 | 104 |

These observations corroborate DMB-R03's named selector bindings against the
effective text capture. They close only that selector correspondence question.
They do not establish parameter/expression/state/stat-field correspondence,
gameplay execution or whole compiled-table identity. Original specimen memory
locations explain collection only; the contract requires semantic values.
Private source-profile and collection records remain separately preserved.

## Blind reader full pass 2: findings and outcomes

Fresh full smaller-model review: MISMATCH 0, GUESS 0, GAP 1, DISCRIM 0,
COVERAGE 2, HARDEST 3, GATED 8. GAP V02's field set was not enumerably owned:
DMB-R04, section 3 and V02 now explicitly consume the logical DamageRecord
field set in core-record-layouts-and-ids section 3.1, without original padding
or unnamed extension obligations. This is a semantic contract reference, not
a new parallel record definition. COVERAGE hitClass source zero: added V03
for that conditional branch and its cleared-zero output. COVERAGE resultFlags
already containing bit 1: no change, the same OR operation applies, and V02
already distinguishes OR from replacement; no different operation is selected
by that bit. HARDEST record clear/duration/fill handoff, negative guard across
four actions, text/limited-compiled distinction: existing ordered request
assertions, guard fixture per selector and explicit profile limits state each
choice; no additional change. The eight GATED entries (R02 and Q01-Q07) remain
explicit. Fresh full review is required after the record-boundary and vector
changes; this record does not claim that final pass has occurred.

## Blind reader full pass 3: findings and outcomes

Fresh full smaller-model review: MISMATCH 0, GUESS 0, GAP 1, DISCRIM 0,
COVERAGE 0, HARDEST 3, GATED 8. GAP: the companion incorrectly attributed
the four negative-index guards to selector 139. This was a copied evidence
sentence, not a source contradiction or a new scope choice. Rechecked the
actual four original-image entry spans and corrected that sentence to selectors
100/103/104/152, with their exact branch and zero-return locations above.
DMB-R01 and its vector remain unchanged. HARDEST record handoff, four-selector
guard and data binding distinctions: no change; the shared logical record,
per-selector entry evidence and limited-profile data records make those choices
explicit. The eight GATED findings retain their documented R02/Q01-Q07 limits.
The next fresh delta review is limited to the companion correction against
the unchanged four-guard contract; no contract rule, gate or output changed.

## Blind reader delta pass 4: final applicable review

Fresh smaller-model delta review of the companion guard-span correction:
MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0. The corrected source-span mappings
for selectors 100/103/104/152 agree with the unchanged negative-index rule
and fixture. Selector 139 appears only in the correction history identifying
the erroneous sentence; it does not limit the current four-selector evidence.
No contract edit followed. With the prior full reviews and recorded outcomes,
the applicable fresh-reader stopping rule is met. The spec remains draft
because successful effects and observation gates remain; review is not approval.

## Cold-action continuation — 2026-10-29

### Scope, identity and compatible reuse

This continuation owns the local action-100 caller only; it does not reverify unrelated actions or declare their gaps closed. The complete original D2Game 336-byte caller at RVA 0xB3D60..0xB3EAF was reread and compared with the existing original-image instruction decode, including both encoded clear/OR spans. No byte-equivalent C view is available for this caller. Original D2Game SHA-256 is `24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53`; D2Common SHA-256 is `59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10`, matching the shared manifest. No source or game artifact was changed.

Reuse was checked against the same module fingerprints, operation identities, semantic inputs and asserted boundaries. [MRH's continuation record](skill-monster-resurrect-and-heal.md) supplies the fully read target-normalization helper 0xE07A0 -> registry lookup 0xE03A0 and its Common target accessors; only that named substage is consumed, not the heal caller's latch/return. [SLD's record](skill-leveled-damage-tohit-and-elements.md) retains its 2026-10-12/15 duration/fill/router readings and gates. [RLI's record](core-record-layouts-and-ids.md) retains its 2026-10-12/13/14 getter/search/minimum readings. [DPO's record](combat-damage-pipeline-order.md) retains its 2026-10-15 apply readings, and [HRC's record](combat-hit-recovery.md) its retained 2026-10-11/12 and targeted 2026-10-15 reaction readings. Reuse identifies these actual callees and request contracts; it adds no observation or promotion of their open effects. The caller-to-callee edges and local decisions below are newly checked.

| Artifact | RVA / ordinal / caller edge | Semantic role | Supported boundary |
|---|---|---|---|
| LOD113C-D2GAME | 0xB3D60..0xB3EAF, do table 0x1074A8 selector 100 | Complete cold-action caller | DMB-R04 through DMB-R07 local writes, ordering, guards and return |
| LOD113C-D2GAME | 0xB3DAD -> 0xE07A0 -> 0xE03A0 | Target normalization and game-registry lookup | MRH-R06 normalization substage only; actual stale/item-mode clearing and cache preservation |
| LOD113C-D2COMMON | #10392 / 0x348F0, Game thunk 0xA598; #11163 / 0x34910; #10336 / 0x35040 | Selected target, cached identity and null setter | Compatible byte-matched accessor evidence under MRH-R06; caller reads selection at 0xB3DB6 |
| LOD113C-D2COMMON | #10510 / 0x4E900, Game thunk 0xACE2 at 0xB3E26 | First duration | SLD-R11 request, store into damage coldLength at 0xB3E2B |
| LOD113C-D2GAME | 0xB3E40 -> 0x9F210 | Elemental fill | SLD-R15/R16/R17, including its second duration and range/router draws |
| LOD113C-D2GAME | 0xB3E54 -> 0xDE0C0 | Direct application with scale request 1 | DPO-R08c/DPO-R13a and subsequent stages, retaining their gates |
| LOD113C-D2COMMON | #10973 / 0x38B70 -> 0x38A80, Game thunk 0xA42A at 0xB3E5E | Existing target life read, stat 6/layer 0 | Actual caller arguments and RLI-R02a/RLI-R02b selected-array/minimum contract |
| LOD113C-D2GAME | 0xB3E75 -> 0xDCF20 | Combat-result reaction on same record | HRC-R02..HRC-R11 with subrule labels and gates preserved |
| LOD113C-D2COMMON | #10689 / 0x32C10, Game thunk 0xA83E at 0xB3E9D | Overlay request then entry guard | DMB-R07: full plain-instruction entry/return read and original-image decode; valid interior effects gated |
| LOD113C-D2COMMON | 0x61EC0 descriptor construction; original strings 0x8B130/0x8B124/0x8B118/0x90F64 | Local Skills field-name correspondence | DMB-R04/R07 input identities; no named loaded values inferred |

### Actual caller decisions

Record guard 0xB3D74..0xB3D93 rejects negative index, index >= signed Skills count at table root+0xBA0, or resulting record address zero; it returns 0 at 0xB3D95 before effects. Valid record reaches the actor+0xC4 OR 0x40 do-handled latch at 0xB3DA3, before target normalization. The compatible dispatch evidence supplies the latch's role. Target helper requires the normal-entry context stated by DMB-R05. Selection absence/self at 0xB3DBD..0xB3DCD goes directly to return 1 at 0xB3EA4, preserving the earlier latch; this is a distinct caller behavior from healing.

Preparation at 0xB3DD0..0xB3E40 is the earlier DMB-R04 clear/result/flags/class/duration/fill request, now followed through the local caller's tail. The fill receives the same actor, skill, level and mutable record, with no numeric-EType override in this caller. Apply receives game, actor, selected target, that record and scale flag 1. At 0xB3E59..0xB3E63 the subsequent current-life call supplies target/stat 6/layer 0. The test is zero/nonzero, not signed positive/nonpositive; only exactly zero ORs record.resultFlags with 2 at 0xB3E67. Otherwise existing result bits persist. Reaction receives the current same record after apply and this conditional OR. It is requested even if apply returned zero; no dependency return is tested in this tail.

At 0xB3E7A the skill field +0x4E is read as signed 16-bit after reaction. The caller checks >0 and permits equality with signed overlay count at table root+0xBC0; greater is rejected. It requests Common #10689 with target, selector and zero extra. That callee's independent entry at 0x32C10..0x32C24 rejects negative or >= count before unit access or calls and normally returns at 0x32C93. Thus equal-count requests are locally observable but have no overlay effects with stable count. Positive in-range entry proceeds to list handling: calls 0x37F40/0x38100/0x3A7D0, optional list flag OR 0x100, then stat 178 write via 0x3A280 and update 0x665C0. These later operations are located, not newly verified as completed effects; DPO-Q18 remains. The label calling this helper path allocation is not its semantic evidence. No intervening unit write is present in the rejected branch.

### Field mapping and named text cell

Common Skills loader 0x61EC0 has no matched-C entry. Its descriptors were read as instruction listings, with original-image decoding corroborating the encoded name-pointer spans. The shared GDA descriptor convention explains the mapping; this local read establishes these additional names rather than attributing them to GDA's limited field inventory. At RVA 0x64944..0x649C6 descriptor names point to original null-terminated strings `ResultFlags`, `HitFlags`, `HitClass` at 0x8B130,0x8B124,0x8B118 and destination offsets 0x12E,0x130,0x134. Their kind codes are 3,8,8 respectively. At 0x63599..0x635BD the `srvoverlay` string 0x90F64 maps to destination +0x4E (kind 20, overlay link context). Registers supplying descriptor kind were traced from their assignments (kind 20 at 0x62125, kind 3 at 0x6464A); there is no intervening reassignment on this straight-line descriptor construction. Consumer access widths come from the actual handler. This establishes field identities, not arbitrary numeric-text conversion rules, blank-cell defaults or live compiled values.

The previously captured complete effective Skills text, with the identity, coverage and hash recorded above, was reread by `skill` and `EType` headers: row DiabCold, EType raw `mag`. That single named text value is now in DMB-R03. Its text-to-compiled numeric correspondence remains DMB-Q01; the synthetic numeric EType 3 in V05 is explicitly supplied, not inferred from that cell. No gameplay capture was added.

### Vector and gate audit

V04 executes actual normalization with the compatible MRH boundary, including stale identity, self, item exclusions and an allowed item. V05 executes actual SLD duration, endpoints, fill/router and NRF arithmetic plus actual RLI existing-stat reading; only apply/reaction/valid-overlay effects are supplied and named. Its range draw is 1791398751 modulo 256 = 95 and damage 351; magic routing preserves the first duration and negative hitClass. V06 separates caller request acceptance from the actual invalid callee guard. V07 contrasts signed -1 with exact zero, already-present lethal bit and a resolved-stat minimum adjustment without storage mutation. Shared gates are retained in Id map, outputs and acceptance. Root self-review checked mutable-record identity, independent endpoint ordering, registry preconditions, selected versus own array, signed count/selector, returns and local/dependency draws. A fresh smaller full reader is required after this extension; historical passes above remain scoped to their earlier contract.

### Cold-action continuation fresh smaller full pass 1 — 2026-10-29

The independent smaller blind reader reviewed the current full contract, its shared dependencies and evidence links without source code, game, web or implementation inspection. It recomputed all executable cases from the behavioral rules and inputs. Findings: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, COVERAGE 0; six GATED groups and three HARDEST items. Existing gates remain DMB-Q01/Q02/Q03, unresolved actions 103/104/152 and Q04..Q07, and V06's stopped valid-overlay entry. They require no contract edit and do not invalidate local assertions.

HARDEST V05 shared-record identity, two durations, actual fill/RNG, supplied apply and actual stat read before reaction/overlay: no change, the vector names each operation/capture, one mutable record and one seed step, with the independent reader's recomputation matching. HARDEST V07 exact-zero versus resolved-stat minimum: no change, four cases state stored and returned values separately, including no storage mutation by adjustment. HARDEST V04 target normalization/latch/early-return/cache: no change, independent fixture cases and normalization-entry latch observer supply each decision. No behavioral edit followed this review. The draft remains unapproved and completed dependency effects remain gated; review is not gameplay confirmation.
