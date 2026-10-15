# Item-event evidence

Behavioral contract: [combat-item-event-triggers-and-procs](../../specs/combat-item-event-triggers-and-procs.md).
Original artifacts: [source manifest](../source-manifest.md), `LOD113C-D2GAME`,
`LOD113C-D2COMMON`, `LOD113C-FOG`, and `LOD113C-STORM`, Windows x86 LoD Expansion 1.13c.
Function descriptions below are analyst descriptions, not original exported names.

## Evidence reuse and identities

The 2026-10-12 investigation read server unit-event registration, dispatch sites, specified
generic callbacks, shared data descriptors and imports through original-code disassembly.
It read the 200-byte callback table and descriptor fields from the original analysis image.
The 2026-10-13 correction reread dispatcher/removal disassembly and event-11 placement.
These compatible readings are reused with their dates and stated limits. The manifest's
2026-10-14 fingerprint correspondence identifies the analyzed original D2Game/D2Common
specimens; fingerprinting does not renew behavior verification or establish data equivalence.

Historical data reads used an original 1.13c installation's archive loader and recorded
names/indices for events, listed item stat bindings and state 172. Complete range/header/raw
cell records for the entire item catalog were not retained. No new data-loader access was
available to this worker. Existing narrow records in the game-data-fields and Crushing Blow
companions can support their own cells, but do not establish the complete item-event catalog.
IEP-R12b therefore retains the catalog as a lead and IEP-Q08 gates its complete binding claim.
Effective file hashes, physical archive attribution and complete overrides remain uncaptured.

## Original function references

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | 0x37190 | none | Create event entry | IEP-R01 | 2026-10-12 disassembly; allocation followed by head insertion; allocator internals outside semantic entry contract |
| LOD113C-D2GAME | 0x9E290; table 0x1077A8 | none | Resolve selector/create entry; shared callback table | IEP-R02, IEP-R06, IEP-R12a | 2026-10-12 disassembly and original-image table read, 50 words; bounds and null-slot branch to no creation |
| LOD113C-D2GAME | 0xD9470 | none | Stat-change item registration | IEP-R03 | 2026-10-12 registration block; 2026-10-14 targeted disassembly reread through 0xD955D, before unrelated stat switch |
| LOD113C-D2GAME | 0x370E0 | none | Find matching item entry | IEP-R03 | 2026-10-12 disassembly; kind, owner key and callback identity match |
| LOD113C-D2GAME | 0x37110 | none | Remove matching registrations | IEP-R05 | 2026-10-12 and 2026-10-13 disassembly; immediate versus executing-entry deferred removal |
| LOD113C-D2GAME | 0x38620 | none | Player setup supplies stat-change callback | IEP-R04, IEP-Q01, IEP-Q02 | 2026-10-14 disassembly and original-image listing, 208 bytes; setup request established, entire notification producer boundary unresolved |
| LOD113C-D2COMMON | 0x3AF00 | 11166 | Create base stat list with supplied callback/context | IEP-R04 | 2026-10-12 and targeted 2026-10-14 disassembly; writes callback/context; notification traversal not read |
| LOD113C-D2GAME | 0x28290, 0x428F0, 0x4F970 | none | Skill/item registration paths | IEP-R06 | 2026-10-12 compatible disassembly; common selector table and kind/owner identity, not skill effects |
| LOD113C-D2COMMON | 0x63940 through 0x63D3A | none | Skill event field descriptors | IEP-R06 | 2026-10-12 original-image descriptor reads; event linking and selector fields, not exact skill row bindings |
| LOD113C-D2GAME | 0x37030 | none | Dispatch registered unit callbacks | IEP-R07, IEP-R13a | 2026-10-12 and 2026-10-13 disassembly; complete traversal/result and successor reload |
| LOD113C-D2GAME | 0x3EDE0 | none | Missile contact roots, two dispatch sites | IEP-R09a | 2026-10-12 disassembly; event 0 precedes missile hit handler |
| LOD113C-D2GAME | 0xDD450 | none | Damage build root | IEP-R09a | 2026-10-12 event-3 call; consumes damage-pipeline contract |
| LOD113C-D2GAME | 0xDC0B0 | none | Mitigation event root | IEP-R09a | 2026-10-12 and 2026-10-13 event-11 call before per-type loop |
| LOD113C-D2GAME | 0xDE0C0 | none | Apply and death roots | IEP-R09a, IEP-R11 | 2026-10-12 events 5/1 or 6/2 and 10/9; result suppression guard; consumes damage-pipeline contract |
| LOD113C-D2GAME | 0x76740 | none | Regeneration death root | IEP-R09a | 2026-10-12 events 10/9; consumes Open Wounds regeneration-death contract |
| LOD113C-D2GAME | 0xDED30 | none | Melee completion root | IEP-R09a | 2026-10-12 events 7 then 3 with copied record after durability |
| LOD113C-D2GAME | 0xDE6D0, 0xDDCB0 | none | Two level-change roots | IEP-R09b, IEP-Q05 | 2026-10-12 disassembly; 2026-10-14 original-image listings of 128/304 bytes corroborate event-12 requests; full eligibility/callee integration unresolved |
| LOD113C-D2GAME | 0x76040 | none | Monster dead-mode event | IEP-R09a, IEP-R13 | 2026-10-12 event 13 after dead mode |
| LOD113C-D2GAME | 0xAD910, 0xAD860 | none | Physical/lightning retaliation | IEP-R08, IEP-R14 | 2026-10-12 disassembly; guards, generated record, apply then reaction |
| LOD113C-D2GAME | 0xACDA0 | none | Damage taken to mana | IEP-R08, IEP-R15 | 2026-10-12 disassembly; positive/player/resource guards, signed percentage, clamp, write and overlay requests |
| LOD113C-D2GAME | 0xACBA0 | none | Mana after kill | IEP-R08, IEP-R16 | 2026-10-12 disassembly; player/nonzero guard, raw conversion/clamp and ordered requests |
| LOD113C-D2GAME | 0xACB10, 0xACA70 | none | Life after kill; demon-only variant | IEP-R08, IEP-R17 | 2026-10-12 disassembly; demon request in second callback, nonzero guard and resource requests |
| LOD113C-D2GAME | 0xAC910 | none | Rest-in-Peace state request | IEP-R18 | 2026-10-12 disassembly; non-null guards and fixed state request; no stat read |
| LOD113C-D2COMMON | 0x34510 | 10945 | State-enable routing | IEP-R18 | Consumed NRF-R12a and timed-state bookkeeping contracts retain their integration gates |
| LOD113C-D2GAME | 0xAF210 | none | Reanimation registration slice | IEP-R13b | 2026-10-12 reading; 2026-10-14 full 208-byte disassembly; positive value guard before direct draw, player chain, entry creation |
| LOD113C-D2GAME | 0x21260, 0xE03A0 | none | Owner lookup and unit lookup boundary | IEP-R13b | 2026-10-14 disassembly and original-image corroboration, 48/112 bytes; valid game/type/id lookup; table 0xFB014, 20 bytes |
| LOD113C-D2COMMON | 0x32C10 | 10689 | Overlay request execution | IEP-R19 | 2026-10-14 full 144-byte disassembly; five local helpers unresolved; no promotion |
| LOD113C-D2GAME | 0x14D0 | none | Signed mulDiv entry | IEP-R15 | Consumed NRF-R05; callback operand order preserved |
| LOD113C-FOG | 0x1CCC0 | 10046 | Storage release forwarder | IEP-R05, IEP-R07 | 2026-10-13 disassembly; downstream allocator internals outside list behavior |

### Import resolution

Original D2Common export-table resolution on 2026-10-14 confirms the following targets.
Export identity proves navigation, not complete getter/write/classification effects.

| Semantic operation | D2Game thunk RVA | D2Common ordinal | Target RVA | Contract / limit |
|---|---|---|---|---|
| Read existing stat | 0xA802 | 10910 | 0x39430 | NRF-R12b/NRF-R12c |
| Read existing unit stat | 0xA42A | 10973 | 0x38B70 | NRF-R12b/NRF-R12c; this getter does not establish replacing-write effects |
| Resource stat operation | 0xA472 | 10887 | 0x3A740 | Same integration limit; callback callsite retained |
| Effective maximum life | 0xA3A0 | 10913 | 0x390F0 | Targeted 2026-10-14 read; selects stat 7, consumes NRF-R12b/NRF-R12c at 0x38A80 |
| Effective maximum mana | 0xA4EA | 10186 | 0x390A0 | Targeted 2026-10-14 read; selects stat 9, consumes NRF-R12b/NRF-R12c at 0x38A80 |
| Demon eligibility | 0xA814 | 10255 | 0x1A610 | Targeted 2026-10-14 read; consumes compatible SBA-R02g predicate contract |
| Overlay request | 0xA83E | 10689 | 0x32C10 | IEP-R19/IEP-Q09 |
| State enable | 0xA3CA | 10945 | 0x34510 | NRF-R12a; update integration retained |
| Base list creation | 0xADAE | 11166 | 0x3AF00 | Callback/context attachment; notification integration IEP-Q01/IEP-Q03 |

### Callback address catalog

The original callback table is D2Game RVA 0x1077A8, 200 bytes. The historical table read
established these selector/address pairs. These locations identify the callbacks; unread
callbacks' roles are leads, not static effect verification.

| Selector | Callback RVA | Selector | Callback RVA | Selector | Callback RVA |
|---|---|---|---|---|---|
| 1 | 0x434E0 | 12 | 0xAD6C0 | 23 | 0x4E5F0 |
| 2 | 0x42C70 | 13 | 0xACDA0 | 24 | 0x41FB0 |
| 3 | 0x42B40 | 14 | 0xAD580 | 25 | 0x44600 |
| 4 | 0x4F090 | 15 | 0xAD410 | 26 | 0x4E500 |
| 5 | 0x4E760 | 16 | 0xACC30 | 27 | 0x4EF70 |
| 6 | 0xAD910 | 17 | 0xACBA0 | 28 | 0xACB10 |
| 7 | 0xACE60 | 18 | 0xACA70 | 29 | 0xAC910 |
| 8 | 0xAF2E0 | 19 | 0xAD2A0 | 30 | 0xADBF0 |
| 9 | 0xAE0E0 | 20 | 0xADDC0 | 31 | 0xAF210 |
| 10 | 0xAD860 | 21 | 0xADCE0 | 0, 32 through 49 | zero |
| 11 | 0xAD7B0 | 22 | 0x4DFE0 | none | none |

Historical role leads for selectors 8/9/14/20/21/30 came from decompilation, not completed
instruction/callee analysis. Selectors 11/12/19 had partial instruction readings. Skill slots
1 through 5 and 22 through 27 were not read. A role label does not establish chance, duration,
eligibility, recursion behavior, or draw totals. Reanimation spawning callback is RVA 0xAE200;
only its registration is in this contract.

## Dispatch chains and semantic mapping

Root chains: the original damage build 0xDD450, mitigation 0xDC0B0, apply 0xDE0C0 and melee
completion 0xDED30 each call dispatcher 0x37030, which invokes entries selected from the
table above. Missile contact 0x3EDE0, regeneration death 0x76740, level roots 0xDE6D0/0xDDCB0,
and dead-mode root 0x76040 reach that same dispatcher. The historical cross-reference
enumeration found 16 direct call sites; it found no event-4 or event-8 dispatch.
Registration path 0xD9470 reaches lookup 0x370E0, selector/create 0x9E290 to 0x37190, or
removal 0x37110. Skill paths 0x28290/0x428F0 also reach 0x9E290; 0x4F970 reads the table inline.

| Semantic term | Original encoding | Supported rule |
|---|---|---|
| Owner entries | Unit +0x90 | IEP-R01/IEP-R05/IEP-R07 |
| Event, executing, remove-after-call | Entry byte +0; word +2 masks 1,2 | IEP-R05/IEP-R07 |
| Kind, owner identity | Entry +4, +8; item 2, skill 1, one-shot 0 | IEP-R01/IEP-R03/IEP-R06/IEP-R13a |
| Callback values, callback | Entry +0x0C/+0x10, +0x14 | IEP-R07 |
| Previous/next | Entry +0x18/+0x1C | IEP-R01/IEP-R05/IEP-R07 |
| Stat identity | Packed `(statId << 16) \| layer` | IEP-R03/IEP-R08; packing is evidence only |
| Stat event fields | ItemStatCost +0x48/+0x4A events, +0x4C/+0x4E selectors; stride 0x144; callback flag bit 11 of +4 | IEP-R03/IEP-R04; GDF-R09 owns logical fields |
| Skill event fields | +0x84/+0x86/+0x88 events; +0x8A/+0x8C/+0x8E selectors; passive +0xB8/+0xBA; owner identity +0x80 | IEP-R06 |
| Retaliation eligibility | Other Unit +0xC4 mask 4 | IEP-R14; producer meaning IEP-Q02 |
| Suppress item events | Damage result mask 0x20 | IEP-R11/IEP-R14 |
| Retaliation hit/reaction flags | Damage result 0x4021, including hit 1 and neutral reaction 0x4000 | IEP-R14; reaction integration IEP-Q04 |
| Physical/lightning components | Damage +0x08/+0x1C | IEP-R14 |
| Physical/lightning type treatment | Damage +0x60 values 0x8D/0x4D | IEP-R14 |
| Damage total | Damage +0x4C | IEP-R15 |
| Reanimation exclusion | MonsterData +0x16 mask 0x0C; absent MonsterData does not set exclusion | IEP-R13b |

These encodings permit independent instruction checking; they impose no storage design on
an implementation of the semantic contract.

## Dispatcher correction

Targeted reconciliation on 2026-10-13 established the seven callback inputs: game in ecx,
event in edx, then owner, other, record, first value, second value on the stack. The original
dispatcher loads event into edx before invocation. It stores each callback result and does
not branch on a nonzero result. Null owner, empty list and no matching event return zero.

Before testing an entry it saves the successor. For a matching entry it marks execution,
calls, clears execution and reloads the successor. Removal then tests deferred-removal or
one-shot kind. Removal of a nonexecuting successor updates the current entry's next link;
removal of the executing entry marks deferred removal. IEP-V08 distinguishes removing the
successor, argument order and result handling. IEP-V09 additionally covers deferred self-removal.

The superseded reading advanced from the pre-call successor after every call. That was
valid only for nonmatches. Earlier damage-pipeline callback prose also omitted the event
argument. The correction changes traversal/argument semantics, preserves prior arithmetic
results, and is not gameplay confirmation. The historical contract and investigation details
are preserved separately; the current public rationale here retains the discrepancy and fix.

## Targeted gap investigations, 2026-10-14

New instruction readings first checked for a reconstruction verified against the original
machine code; none was available for targeted functions, so disassembly was used. Original-image listings
were generated from the manifest-matching D2Game bytes for 0x38620 (208 bytes), 0xDE6D0
(128 bytes) and 0xDDCB0 (304 bytes). All requested bytes decoded, and the relevant calls,
branches and writes agreed with the disassembly. This corroborates previously undecoded
instruction runs; it does not read omitted callees or establish every producer.

**Installation identity correction.** The original setup is RVA 0x38620, not 0x37620.
The earlier record confused an internal relative label with the original module RVA.
At 0x3865F it supplies callback VA 0x6FCF9470, then calls import thunk RVA 0xADAE at
0x38667. That resolves to Common ordinal 11166 / RVA 0x3AF00, which stores callback/context
in base-list +0x5C/+0x60. Setup sets Unit +0xC4 masks 2/4/8. Its later type check requires
a player; therefore this setup path alone cannot establish monster/hireling registration.
Call-site metadata identified caller regions 0x550A0, 0x55840, 0x79210, 0xE1D90, 0xECF30,
0xED540; those complete callers were not read. A callback-address search additionally found
monster path 0xAFDB0; its callback installation region was inspected, without claiming the
whole monster setup path. IEP-Q01/IEP-Q03 remain integration gates.

**Level roots.** Original-image corroboration locates event-12 requests at 0xDE743 and
0xDDDD0. The former adjusts level/experience inputs and compares a supplied value against
a level-derived helper result before its notification. The latter loops over level thresholds
and reaches its notification after an increase. Their helper/caller boundaries remain unread;
the recorded difference does not establish a complete level-change contract. IEP-R09b remains
unresolved for that eligibility rather than borrowing the combat-root label.

**Reanimation guard correction.** Full 0xAF210 disassembly tests the owner stat strictly
positive before calling random helper 0x11D0. The earlier prose omitted that guard. Nonpositive
values consume no draw; after a positive value, chance draw precedes owner-chain lookup and
is consumed even when no player is found. IEP-V12(b) distinguishes this correction. The
existing one-shot registration behavior and older numeric vectors are unchanged. The owner
lookup calls 0x21260, which uses 0xE03A0 to resolve a monster owner identity. Original-image
listings of both functions corroborate all relevant branches and the tail transfer; the
20-byte unit-bucket table at 0xFB014 was also read. For valid game and unit identities, lookup
walks the corresponding bucket and returns the matching unit or null. Null/non-monster and
absent owner metadata end the owner lookup. This closes the in-boundary owner-lookup reading;
owner-relationship production and the spawn callback remain outside the registration slice.

**Overlay integration.** Common 0x32C10 bounds the requested overlay against its loaded count,
finds or creates/attaches a flag-0x80 list, marks the base list, stores stat 178 and requests
update. Calls target 0x37F40, 0x38100, 0x3A7D0, 0x3A280 and 0x665C0. Compatible Crushing Blow
records also leave these helpers open. Repeating the outer helper adds an exact remaining
callee boundary, not a verified stored overlay effect. IEP-R19 becomes unresolved with IEP-Q09.

**Resource/classification integration.** New export resolution and reads locate maximum
mana/life at Common 0x390A0/0x390F0 (stat 9/7) through 0x38A80. Demon classifier 0x1A610
checks monster and class bounds, then a MonStats masked flag. These readings identify consumers
and local scope. Compatible NRF-R12b/NRF-R12c evidence owns 0x38A80 and SBA-R02g owns the
same demon predicate/name correspondence. Those contracts close maximum-getter and demon
classification reading gaps; IEP-V14 exercises distinct maximum stat selection. Replacing
stat-write production remains IEP-Q10/NRF-Q20. No supplied-input arithmetic result changed.

## Observation procedures

No gameplay records exist; all observation vectors remain pending. These procedures specify
measurements, not permission to run a gameplay campaign.

**IEP-O01.** Use an original-1.13c player with two items contributing the same proc stat/layer.
Record class, level, difficulty, exact contributions, and the original-profile named cells.
In a safe town position, read registrations before equip, after each equip, after one removal,
and after both removals. Expect one registration per positive event for the summed identity,
retained after the first removal and absent at zero. For ordering, separately use one item
with two distinct proc stats and capture notification calls and resulting entry order; compare
supplied notification order with IEP-R01/IEP-R03. Observe hireling/monster producers separately.

**IEP-O02.** Use a player/monster melee pair with known proc entries at events 3/5/1/9/10/7.
Record class, levels, difficulty, current life, damage and definitions. Capture dispatcher
event/owner/other/record inputs for one killing queued melee hit. Compare only reached stages
with IEP-V11; missed captures, absent registrations and suppression make absence inconclusive.

**IEP-O03.** Use a player wielding a missile weapon and a known event-8 entry; record the item
definition and layer. At a safe isolated target, capture missile attack/contact/apply roots
and dispatcher requests across repeated attacks. Confirm root coverage before interpreting no
event 8; do not infer no casts from an unregistered or ineligible callback.

**IEP-O04.** Use a melee monster attacker with documented physical/lightning resistance and
reduction and a holder with a measured retaliation value. Record levels, difficulty, resource
values and eligibility condition. Capture generated record, apply result, resource delta and
following dispatches for a nonlethal and a lethal retaliation. Compare mitigated loss and
suppression independently; do not promote unobserved reaction/overlay effects.

## Review and migration record

Historical targeted reader, 2026-10-13: IEP-R01/IEP-R05/IEP-R07 and IEP-V08 were reviewed
alongside damage-pipeline and mitigation vectors. MISMATCH 0, GUESS 0, GAP 0, required
COVERAGE 0, GATED 2, HARDEST 3. The documented gates belonged to stat integration and
callback-specific draw totals. The mutation/argument vector was among the hardest cases;
no undocumented choice remained in that targeted set. Deferred self-removal coverage was
advisory; IEP-V09 now covers it. That review did not cover the full regenerated contract.

Migration 2026-10-14: private originals of both edited files were hash-compared before
replacement and matched. Repository searches found no consuming implementation plan/code/test
citation, so obsolete permanence boilerplate was removed; already ordered identifiers were
retained and mixed evidence boundaries split into sub-ids. Binary layouts moved here; public
history retains dates, correction rationale and limits. Original numeric resource outcomes
0xA00 damage, 0x1A00 mana, 0x6900 life and capped 0x7D00 life were retained as semantic
outputs. Additional executable vectors cover missing verified rules and discriminating branches.

Label/gate changes: notification integration and overlay execution are unresolved with exact
callee/producer gaps; combat/death roots remain static while full level-root eligibility is
separately unresolved. The code selector catalog is static; complete data/role binding is
recovered and gated. Neutral eligibility/reaction names preserve known tests without guessing
their wider meaning. Callback arithmetic/ordered requests are static at supplied-input
boundaries, with stored-write/overlay integration explicitly gated. Maximum getters consume
the established stat-reader contract and demon classification consumes SBA-R02g.

Checks before rewriting: spec linter reported 18 errors and 15 warnings, chiefly private
provenance and earlier layout. A revised lint returned zero errors and one method-wording
warning, which was resolved. Self-review checked semantic inputs/outputs, ordered callbacks,
raw resource results, source dates, local callee limits and supplied-input gates. A fresh full
blind-implementer pass is required: new semantic fixtures,
guard correction, acceptance subsets and execution/gate boundaries supersede the historical
targeted review. No implementation tests, gameplay observation, commits or publishing occurred.

### Full reader pass 1, 2026-10-14

The fresh reader used the regenerated semantic contract and cited dependency contracts.
Reported counts: MISMATCH 0, GUESS 0, GAP 1, DISCRIM 3, GATED 8, COVERAGE 17, HARDEST 3.
All recomputed arithmetic matched. COVERAGE was supplied as grouped rule categories; the
reported count is retained rather than inventing additional findings to reconcile grouping.

| Finding | Outcome |
|---|---|
| GAP: IEP-V14 asserted no resource write/overlay without invoking a resource callback | Spec edit: removed that assertion. V14 asserts only the requested effective maximum results; it supplies no callback execution and claims none |
| DISCRIM: IEP-V03 did not supply multiple item contributors | Spec edit: named distinct A/B contributions of 5 and the exact add/remove sequence and summed notifications; total 10 must still have one entry |
| DISCRIM: IEP-V06 had no differing physical component | Spec edit: total 0x3200, physical 0x1000, fire 0x2200 and current life 0x1000; physical-only or life-clamped alternatives produce 0x1333 rather than 0x1A00 |
| DISCRIM: IEP-V10 cannot reject two separate but identical tables | No behavioral change: identical mappings are behaviorally equivalent. R06 and V10 now explicitly require identical selector resolution and preserved callback values, without requiring a storage architecture; the original single table remains source evidence |
| COVERAGE: IEP-R01 absent owner during creation | Spec edit: IEP-V01(a) creates no entry for absent owner |
| COVERAGE: IEP-R02 selector bounds/populated range | Existing IEP-V02(b) already covers -1, 0, 32, 49 and 50; added V02(d) for every populated selector 1 through 31 |
| COVERAGE: IEP-R03 distinct layers and both-event removal | Spec edit: IEP-V03(a) registers two events at each of two layers, then removes exactly both entries at one layer |
| COVERAGE: IEP-R05 multiple matches | Spec edit: IEP-V03(a)'s zero notification removes two matching entries while retaining the other layer |
| COVERAGE: IEP-R07 new head and shared record mutation | Spec edit: IEP-V08(c) creates head H during X, does not visit H, and requires Z to see X's physical-field mutation |
| COVERAGE: IEP-R09a other dispatch sites | Spec edit: IEP-V11(b) through (f) cover missile contact/mitigation/apply, missile attacker-event suppression, regeneration death with/without killer, and dead mode; the base covers melee build/scale |
| COVERAGE: IEP-R10 complete root enumeration | No change to evidence boundary: the dated complete call-site enumeration establishes the absence claim. Reached-site fixtures cannot prove enumeration completeness; V11's additional roots cover their positive request behavior without claiming an executable absence proof |
| COVERAGE: IEP-R12a populated selectors beyond 6 | Spec edit: IEP-V02(d) checks all populated selectors create entries; unread callback effects remain outside the catalog contract |
| COVERAGE: IEP-R13b exclusion, type, absent owner and chance failure | Spec edit: IEP-V12(c) through (f) distinguish no-draw guards from a consumed failing draw; (g) also covers a direct player owner |
| COVERAGE: IEP-R14 lightning and missing/nonpositive guards | Spec edit: IEP-V05(b) requires lightning-only record/ordered requests; (c) covers absent units and values 0/-1 |
| COVERAGE: IEP-R15 absent record, nonpositive total/value, overkill and clamps | Spec edit: V06 base distinguishes overkill/physical-only alternatives; (c) covers guards, (d) cap, (e) signed-wrap floor |
| COVERAGE: IEP-R16 positive/full/zero mana-after-kill | Spec edit: IEP-V07(e)/(f) cover positive return/write and full/zero refusal |
| COVERAGE: IEP-R17 positive demon, negative life bonus and zero | Spec edit: IEP-V07(g) through (i) cover those branches; (j) supplies missing-unit guards for the kill-resource selectors |
| COVERAGE: IEP-R18 absent owner | Spec edit: IEP-V13(b) requests no state and returns zero |
| HARDEST: IEP-R07 successor reload/deferred self-removal | No change: R05/R07 explicitly distinguish immediate and deferred removal; V08/V09 already discriminate both cases; V08(c) adds record/head behavior |
| HARDEST: IEP-R13b draw/chance/player lookup | No change to rule: R13b orders guards, draw and lookup explicitly; expanded V12 adds result-distinct controls |
| HARDEST: IEP-R15 percentage/resource clamps | No change to rule: R15 states operand order, wrapping, floor then cap; expanded V06 exercises both clamps |

The eight reported GATED findings restate existing documented limits; no new gate was
introduced and the gap file's evidence questions are unchanged. These edits change fixtures,
assertions and R06's storage-neutral expression, so another fresh full reader pass is required.
This is pass 1; no delta-review stopping rule is claimed. No new original-code reading or
gameplay observation was needed for these already established branches.

### Full reader pass 2, 2026-10-14

Reported counts: MISMATCH 0, GUESS 0, GAP 2, DISCRIM 0, GATED 28, COVERAGE 3,
HARDEST 3. The 28 GATED entries include the vectors' supplied-input boundaries and repeated
existing dependency limits. They are reader data, not 28 new evidence questions or blanket
gates on verified assertions. All arithmetic matched and no ambiguous decision was reported.

| Finding | Outcome |
|---|---|
| GAP: IEP-V11 lacks attacker eligibility and clear result input for build event 3 | Spec edit: V11 names player attacker A, monster D and the shared record's suppression condition false at the build/apply/death sites, with DPO-R23's exact reached-site eligibility. R09a's table now states the player/hireling-attacker qualification explicitly |
| GAP: IEP-V13 omits NRF-R12a's unconditional update request | Spec edit: V13 supplies state count 173 and target state storage, executes NRF-R12a routing, and logs ordered active-bit then MarkForUpdate requests. It distinguishes the callback's state request, dependency routing, and unexecuted stored/update effects. Section 3 names the consumed environment and section 5 includes the requests |
| COVERAGE: IEP-R03 negative new total | Spec edit: V02(e) supplies -5 and requires registration; this rejects a positive-only registration gate |
| COVERAGE: IEP-R05/IEP-R06 kind-scoped skill removal | Spec edit: V10(a) gives skill/item entries the same supplied owner identity; removing skill kind retains and dispatches only the item entry |
| COVERAGE: IEP-R14 minimum positive value | Spec edit: V05(d) supplies value 1, producing physical 256 and the ordered requests; this separates the positive boundary from 0/-1 guard cases and the existing value-10 sample |
| HARDEST: IEP-V08 mutation and final return handling | No change: R07 and V08 already reload the successor after callback mutation, continue after nonzero results, and assert final result 7. DPO-R23 separately owns ignoring that final result at its caller |
| HARDEST: IEP-V12 draw/chance/player lookup | No change: R13b and V12 already distinguish no-draw guards, consumed failure, equality failure, absent player and direct player ownership |
| HARDEST: IEP-V06 percentage/clamps | No change: R15 and V06 already fix mulDiv operand order, raw units, wrapping, signed floor and maximum cap with distinct cases |

No new source reading, evidence label or gap-file row was needed. The dependency contracts
were read to resolve both gaps. These changed inputs/assertions and the explicit R09a
qualification require another fresh full reader pass. Final lint is run after this edit;
the third full pass and any subsequent delta stopping rule remain unclaimed here.

### Full reader pass 3, triaged 2026-10-15

Reported counts: MISMATCH 0, GUESS 0, GAP 2, DISCRIM 3, GATED 11 grouped, COVERAGE 6,
HARDEST 3. The grouped GATED list restates existing dependency and supplied-stage limits;
its grouping/count does not introduce new questions or blanket gates. The changes below
affect fixture inputs/assertions and evidence references only. Rule behavior, ownership,
evidence labels, execution boundaries and gates remain unchanged.

| Finding | Outcome |
|---|---|
| GAP: IEP-V03 omits the second event/selector | Spec edit: base definition explicitly supplies second event 0/selector 0, unless the layer case overrides them; the base still produces counts 1,1,1,0 |
| GAP: IEP-V11 death report does not itself establish DPO-R23's clear-result death gate | Spec edit: base explicitly supplies the allowed death-event gate for the same result-suppression-clear record, separately from the death report; no inference from lethality alone is required |
| DISCRIM: IEP-V04 does not reject one-shot removal only on nonzero callback result | Spec edit: V04(a) returns zero from the callback, requires exactly one call, all-zero dispatch returns and final empty list |
| DISCRIM: IEP-V09 does not reject deferred removal only on nonzero callback result | Spec edit: V09(a) returns zero from X after requesting self-removal, then visits Z; final list must contain only Z and final return remains 8 |
| DISCRIM: IEP-V11 lacks a contrasting blocked death-result gate | Spec edit: V11(g) supplies a death report but blocked result gate at the isolated death tail and requires no events 10/9 |
| COVERAGE: zero callback results during one-shot/deferred removal | Covered by the two discriminating cases V04(a) and V09(a); no additional duplicate case |
| COVERAGE: build event 3 ineligible | Spec edit: V11(h) separately supplies an ordinary non-hireling monster and a player with suppressed result; both produce no build event 3 |
| COVERAGE: death result suppression set | Covered by V11(g), which isolates the death tail and cannot pass merely by skipping earlier apply dispatches |
| COVERAGE: event 6 item suppression set while missile-attacker suppression is clear | Spec edit: V11(i) supplies that combination and requires neither event 6 nor 2. DPO-V31 independently owns the other flagged-apply combinations, but its current cases keep result suppression clear and therefore do not replace this control |
| COVERAGE: state count invalid or base storage absent | No new local case: NRF-R12a owns the shared routing; NRF-V33(c) covers absent base list, (d)/(k) invalid identity. V13 now explicitly cites those controls. R18's callback selection/non-null request and unconditional return 1 do not introduce a new routing branch |
| COVERAGE: regeneration tick without death | Spec edit: V11(j) supplies an alive regeneration outcome with positive life and requires no death notifications; existing V11(d)/(e) retain present/missing-killer death cases |
| HARDEST: IEP-V08 traversal | No change: successor reload, shared-record mutation, insertion and final dispatcher result are already specified and discriminated |
| HARDEST: IEP-V11 stage ownership | No rule/boundary change: fixtures explicitly supply reached dependency outcomes and execute only the named decisions. Added controls distinguish build, apply and death decisions without claiming full upstream execution |
| HARDEST: IEP-V06 arithmetic | No change: the existing vector fixes total versus component/life, operand order, wrapping, floor and cap; no ambiguity or arithmetic mismatch was reported |

All three full reader runs returned MISMATCH 0, and every reported finding has a recorded
outcome. Under the checklist's stopping rule the next reader may be a fresh delta check of
V03, V04(a), V09(a), V11 base/(g)/(h)/(i)/(j), and V13's dependency-control reference.
That next check is pending; no review completion is claimed. No new original-code reading,
gameplay observation, evidence label or gap-file row was needed. Final lint follows these
fixture edits.

### Delta reader pass 4, 2026-10-15

The fresh reader checked all changed material since full pass 3: V03 base, V04(a),
V09(a), V11 base and (g)–(j), and V13's dependency-control reference. Reported counts:
MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0. It independently recomputed the changed cases and
confirmed the explicit second-event inputs, zero-return one-shot/deferred removal, separate
death-report/result gates, build eligibility, missile suppression and alive-regeneration
controls. The two GAP and three DISCRIM findings from pass 3 are resolved; the reader's
summary informally called them three gaps, but introduced no additional finding.

The NRF controls cover missing base storage and invalid state identities without duplicating
their behavior locally. No contract change follows this review. It completes the applicable
blind-reader review of the supplied-input assertions; declared integration, original-profile
binding and observation gaps remain. This does not approve the spec or establish runtime parity.

Publication self-review also corrected the import table's generic description of Common
#10973/+0x38B70 to identify the existing-stat getter, consuming the foundations' established
function record. This reference clarification changes no behavioral rule, fixture or gate.
