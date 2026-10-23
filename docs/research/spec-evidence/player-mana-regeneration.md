# Player Mana Regeneration: Evidence

Behavioral contract: [Player mana regeneration](../../specs/player-mana-regeneration.md).
Original artifact identities: [Source manifest](../source-manifest.md), LOD113C-D2GAME and LOD113C-D2COMMON.

## Evidence scope and reuse

Historical static readings and the narrow gameplay observation are dated 2026-10-12. Their compatible coverage is reused with original limits, not renewed by relocation. Current analysis specimens were hash-checked against the manifest on 2026-10-14; this does not retroactively prove historical loaded DLL identity. New source readings use the current byte-inventory of reconstructions checked against original bytes where a function is present, otherwise disassembly. Function descriptions below are analyst descriptions, not exported names unless an ordinal is specified. No source text or game tables are reproduced.

The full 528-line prior draft and its private investigation references, fixture layouts, session/save details and historical reader notes were preserved in a private archive before editing. Public references below identify original artifacts/functions without that archive.

## Dispatch and source mapping

All entries target LoD Expansion 1.13c. Module locations are RVAs; export ordinals refer to D2Common unless stated. Read dates below retain actual scope.

| Original module / RVA | Export | Semantic operation / support | Reading and limits |
|---|---|---|---|
| D2Game 0x2E050; caller 0x2E1E0 | none | Per-frame call into event-wheel dispatch | 2026-10-12 disassembly; caller identified but not fully read |
| D2Game 0x8F070 | none | Wheel dispatch; player group calls 0x8ECC0 | 2026-10-12 partial byte-only decoding; ordering remains gated; current STL-R02a retains that limit |
| D2Game 0x8ECC0; table 0xFBB74, slot 3 | none | Player event selects 0x79B10 with two event arguments | 2026-10-12 instructions/table read; historical return site 0x8ED3B corroborates call |
| D2Game 0x79B10 | none | PMR-R01a/R02: repost and gates, life/stamina/mana calls | 2026-10-12 complete disassembly; reread 2026-10-14, no byte-only instruction region in behavioral body |
| D2Game 0x77950 | none | PMR-R03..R07: mana computation and requests | 2026-10-12 complete disassembly; reread 2026-10-14, all input reads/branches/requests traced |
| D2Game 0x77CB0; 0x77A50 | none | Life/stamina calls before mana | 2026-10-12 life return-path reading (only exit returns 1), stamina entry only; complete effects outside mana boundary |
| D2Game 0x14D0 | none | PMR-R04 percent arithmetic | 2026-10-12 caller operand mapping; current NRF-R05 supplies complete shared contract |
| D2Game 0x8EA80 | none | Schedule request; requested frame and arguments | 2026-10-12 partly decoded disassembly; 2026-10-14 entry-to-field-store slice revisited, byte-only regions not independently decoded in this mana pass; consume current STL-R01/R02 where verified |
| D2Game 0x79E10 | none | Immediate dispatch through player table | 2026-10-12 lead for natural monster entry; producer remains PMR-Q05 |
| D2Game 0xDE0C0; 0x336D0 | none | Type 3 scheduling at frame+1 on damage and frame+25 via helper | 2026-10-12 call-site readings; 2026-10-14 reconstruction checked against original bytes confirms 0x336D0 request; its root caller remains unresolved |
| D2Game 0x79210; 0x78DD0 | none | Player creation/join leads | 2026-10-14 limited direct scheduler/helper label search found no matching call in these functions; not complete dispatch-root exclusion |
| D2Common 0x38B70; 0x38A80 | 10973 / internal | Generic existing stat read | Reused RLI-R02a/RLI-R02b contract, including definition minimum; no new claim about producing resolved arrays |
| D2Common 0x390A0 | 10186 | Guarded maximum mana getter | 2026-10-12 disassembly; complete 80-byte function reread 2026-10-14, null/list/count/selected-definition guards and stat 9 call into 0x38A80 |
| D2Common 0x33CD0; 0x37DB0 | 10494 / internal | State presence | Current RLI-R03 handles absent storage/type/range; earlier unread-guard question no longer applies to this consumed boundary |
| D2Common 0x38520; 0x37F80 | 10871 / internal | State-list lookup | 2026-10-12 lookup reading; current RLI-R03 supplies complete selection contract |
| D2Common 0x3A5F0; 0x3A3B0 | 11111 / internal | Detach request | Historical unlink/partial propagation reading; complete effects remain STL-R05/PMR-Q03 |
| D2Common 0x3AD40; 0x3A9F0 | 11108 / internal | Free request | Historical reconstructions checked against original bytes; complete integrated lifecycle remains gated |
| D2Common 0x3A710; 0x3A330 | 10551 / internal | Add stat request and underlying mutation | 2026-10-14 reconstruction checked against original bytes of wrapper confirms nonzero guard and unit base-list forwarding; underlying propagation 0x39F70/0x38550 remains unread here |
| D2Common 0x60A80 | none | ManaRegen preparation | 2026-10-12 descriptor identifies unsigned byte field; current GDA-R06a owns its logical schema |

D2Game's mana caller resolves the shared operations through thunks 0xA42A(10973), 0xA4EA(10186), 0xA68E(10494), 0xA5D4(10871), 0xABB6(11111), 0xABAA(11108), 0xA4D2(10551). D2Common data export 11170 supplies table context; original D2Game data import slot RVA0xF829C. The max getter passes stat identity 9/layer 0 to the same underlying reader as generic unit stats.

Original ABI details supporting the semantic mapping: mana unit is carried in EDI; current mana saved locally is restored after cleanup, while maximum remains captured. The percent caller passes first operand 100+percent and second operand class base to 0x14D0. The handler copies its two stack arguments unchanged into scheduling before its unit test; calling-convention details do not define a required new API. Original layout fixtures remain private.

## Table evidence

Historical 2026-10-12 record: the owned game loader read CharStats, ItemStatCost and States; the draft named Sorceress/Amazon ManaRegen 120 and stat/state identities. Full hashes/ranges and supplying archives were not recorded. Reuse GDA-R06a/GDF-R01, GDF-R02/GDF-R03 and their named-cell records for schema/identity; GDF's 2026-10-14 TD-CLASS record also corroborates Sorceress ManaRegen 120 against selected compiled data. No new compiled-table correspondence is claimed by this mana pass.

Effective-file reread 2026-10-14: reported launch key lod-113c, same historical vanilla source profile; current member inventory excludes the separate PD2 members. The own-loader read of data/global/excel/charstats.txt requested off 0/whole file and returned 3360/3360 bytes, range 0..3360. The complete returned body was decoded and parsed as tab-separated data; column selection used header names class/ManaRegen (positions 0/12), not assumed positions. Target rows were complete.

| Row identity / indexing | Column | Raw cell | Interpretation / support |
|---|---|---|---|
| Amazon, zero-based file data row 0 / prepared class 0 | ManaRegen | 120 | Unsigned byte 120; PMR-V03/V04 input |
| Sorceress, file data row 1 / prepared class 1 | ManaRegen | 120 | Unsigned byte 120; PMR-V01/V02 input |

File row 5 is the Expansion separator; GDF-R01 supplies the filtered class convention. Current patch inventory reports file override disabled, but no complete override/archive inventory or supplying-MPQ attribution was captured. Version-console refresh failed with connection refused; historical vanilla identity and current reported launch key are reused with that limitation. Effective-file hash was not captured. This current read corroborates effective text cells; it is not a fresh claim of pristine installation identity, compiled runtime equivalence or gameplay execution.

## Historical runtime records

Owned game reported lod-113c on 2026-10-12. Character was dead, type 0, class 4, mode 17, life 0/55, mana 10/10. Private identity and session details are retained privately.

| Target / observation | Actual coverage | Result / unobserved assertions |
|---|---|---|
| D2Game 0x79B10, dead handler entry/exit | 200 records held at cap; 3 read, dropped 0; two player units (character/corpse), each type 0, class 4, mode 17; caller 0x8ED3B; return register 17 and stack cleanup 8 bytes | Supports handler identity/dead-mode exit (PMR-R02a/PMR-V11). Repeated entries are compatible with reposting, but exact due frames, event arguments, insertion, count/cadence and balances were not measured |
| D2Game 0x77950, dead-state window | 0 records over 40 seconds | Inconclusive absence; does not confirm arithmetic or promote any rule |

No mismatches recorded. Both observation hooks were disarmed after that pass. Historic broad runtime labels on repost/full eligibility are narrowed: PMR-R01a is static and PMR-R02a alone retains the matching dead-exit coverage. PMR-V10's full scheduling assertions are not marked observed. Stack cleanup corroborates ABI, not every argument or schedule value.

## Gap assessment and targeted investigation

- PMR-Q01: reuse damage/helper posting leads; 2026-10-14 helper checked against original bytes read and limited creation/join call search leave the initial root unresolved.
- PMR-Q02: reuse historical player-table mapping and current STL-R01/R02; queue order still gated in STL-R02a. Mana pass did not claim to finish dispatch-wheel internals.
- PMR-Q03: inspected cleanup caller completely; lifecycle owner still gates free/propagation. Request order and captured-input use remain verified independently.
- PMR-Q04: reread maximum getter and matched add wrapper; existing NRF stat reads close getter uncertainty, not write effects. Underlying add/propagation and sync are separate unanswered paths.
- PMR-Q05: historical immediate-dispatch lead remains; accepted type 1 case is not proof of a natural producer.
- PMR-Q06: preserved actual observations; no gameplay campaign requested or performed. Cadence/arguments/living arithmetic remain pending.
- PMR-Q08: generic state-test storage guard resolved through RLI-R03. Null-target scheduler consequences remain outside the verified scheduling request.
- PMR-Q09: current lifecycle cancellation contract does not establish every unit-teardown producer; chain lifetime remains unverified here.

No rule is claimed not found after a partial search. No source/table age alone caused re-verification.

## Observation procedures

- PMR-O01: with an owned matching 1.13c game, identify player handler, scheduler and mana tick through the locations above. Record a living player and a dead mode 17 player separately, with game frame, target identity and both event arguments. Observe enough uncapped timestamped records to compare one per server frame; expected cadence 25 per second must be measured rather than inferred from three entries. Verify scheduler arguments/due frame before confirming repost. Do not reuse a quiet mana-tick capture as proof.
- PMR-O02: living Sorceress level 1, Normal, town, no recovery equipment, spend mana below maximum. Read actual mana/max/flat/percent/state inputs; do not assume starting energy/max. Capture consecutive tick inputs and additive requests through final write. If max 2560 and named ManaRegen 120, expect+1 raw/tick until clamp. Repeat positive/negative or suppression inputs only on a dedicated authorized fixture; compare all inputs and ordered clamps.
- PMR-O03: with an authorized potion fixture, distinguish starting below maximum from already full. Capture state 106 list presence, preclamp regen, detach/free requests, contributions and active membership before/after a full-mana tick. A missing list alone does not establish every callback/stat/state effect; resolve the lifecycle code path too.

## Migration and review record

2026-10-14: preserved existing boundary, PMR identifiers and ten original numerical scenarios. Binary locations/layouts moved to public source mapping or private audit. Original symbolic event arguments are instantiated as 11/22 in the supplied request vector; this changes no copying rule. New V11 isolates historically observed exit; V12..V15 distinguish request gates/order, getter guard, operand/wrap arithmetic and cleanup timing/captured values. Complete mutation remains gated; supplied fixture balances are not integrated proof.

Corrections: original Stat notation incorrectly excluded RLI-R02b; the max getter uses the same underlying reader, so consume its minimum semantics. GDF now already lists ManaRegen. Removed the unsupported assertion that arbitrary i32 inputs in 8.8 units cannot overflow, and the implementation default of assuming no additional stat effect. Original broad runtime labels are narrowed to the actually observed exit. The other-build question is outside 1.13c scope; no parity inferred.

Historical blind review 2026-10-12 reported 14 GUESS/5 GAP; old draft said they were folded but no rerun was performed. Its detailed original notes remain archived; that review does not certify this changed contract. No implementation tests, gameplay action or new gameplay observation occurred.

### Full blind-reader pass 1, 2026-10-14; corrections 2026-10-15

Reported counts: MISMATCH 2, GUESS 1, GAP 0, DISCRIM 0, COVERAGE 4, HARDEST 3, GATED 26. GATED included supplied-input vectors; those boundaries alone do not invalidate verified local results. Existing assertion-level integration limits remain unchanged.

Both mismatches and the guess had the same cause: a mechanical spacing cleanup converted notation for event type 3 into a negative number in PMR-R01a. The already-read D2Game+0x79B10 pushes positive 3, matching original PMR-V10/V12. Corrected the rule to unambiguous "event type 3". This is a document error supported by the recorded source, not a change in original behavior or evidence for negative-type scheduling.

| COVERAGE finding | Outcome |
|---|---|
| R01a wrapping next frame | Added V16(a), asserting the wrapped request only; negative-frame insertion remains outside that assertion |
| R02b eligible monster | Added V16(b), a supplied handler-entry case; natural producer remains Q05 |
| R03 initial read order | Added V16(c), a top-level getter read log with supplied results |
| R04 absent selected class record | Added V09(c), class 0 with no selected record, distinguishing return before flat recovery |

HARDEST, all no change: R04/V14 explicitly asserts asymmetric mulDiv's intermediate; R07a and V04/V07/V13/V14 state sequential signed clamps; R01a/V10/V12 distinguish scheduling requests from stored effects. No missing decision was reported for these. The rule correction and new vector assertions require a fresh full reader. Documentation lint ran at 0 errors/0 warnings before these fixes; final results will be recorded after rerunning it.

### Full blind-reader pass 2, 2026-10-15

Final adjudicated counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 8, COVERAGE 3,
HARDEST 3. All V01–V16 were independently recomputed. The preliminary report mistakenly
treated 1,000,000 as greater than hexadecimal 0x100000. The reader rechecked the decimal
threshold 1,048,576 and retracted its V14(a) mismatch/discrimination findings: the documented
result 655,000,000 and swapped-operand result 655,370,000 are correct. Both reports and the
adjudication are retained here; no numeric correction was made to V14(a).

| Advisory finding | Outcome |
|---|---|
| COVERAGE R02b monster mode 0 | Added V12(h), which must still schedule first and then skip all resource calls |
| COVERAGE R04 percent-factor overflow | Added V14(c), distinguishing wrapped factor/product arithmetic from a nonwrapping calculation |
| COVERAGE R07a room subtraction/negation wrap | Added V14(d), asserting wrapped intermediate limits, signed request and synthetic balance |
| HARDEST V14 mulDiv thresholds/order | No change to rules or original case: exact thresholds establish the asymmetric intermediate already asserted |
| HARDEST V13 guarded getter/negative room | No change: the zero getter result and subsequent negative request are explicit |
| HARDEST V16 request/read boundaries | No change: supplied handler requests, direct-entry read log and negative-frame insertion exclusion are explicit |

The eight GATED questions retain their existing kinds and assertion-level limits; no new
gate or label is introduced. The gap record's stale negative-event wording was corrected
to event type 3 consistently with the already-read source and corrected main contract.
The added fixture cases require a third fresh full reader pass. Review completion is pending;
no new source reading, gameplay observation or implementation test followed this pass.

### Full blind-reader pass 3, 2026-10-15

Reported counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1, GATED 5 grouped, COVERAGE 2,
HARDEST 3. Every V01–V16 case, including the new wrapped arithmetic, recomputed correctly.
The grouped gates restate existing questions and do not replace their individual boundaries.

| Finding | Outcome |
|---|---|
| DISCRIM V01 first tick permits retained fractional remainder | Made the existing cumulative-refill assertion executable as V01(a): explicit initial current 0, repeated direct calls with the synthetic balance fed back, exact request sequence and balances at calls 3000/3414. The earlier output already named 3414 calls, but the new fixture makes the repeated execution and assertions unambiguous |
| COVERAGE R03 conditional minimum replacement | No local duplicate: RLI-R02b and RLI-V03 own the actual conditional replacement branches; this consumer explicitly supplies getter results and does not reimplement that dependency |
| COVERAGE R07a clamp order under arbitrary i32 values | Added V14(e), supplied maximum -100/current 100 and suppressed zero recovery. Upper-then-lower requests -100; reversing them requests -200 |
| HARDEST V14 wrapped arithmetic/clamps | No rule change: explicit wrapping and signed tests remain; the new order case discriminates an additional already-stated behavior |
| HARDEST V12 eligibility/order | No change: V12's exact logs distinguish each supplied handler gate, including monster mode 0 |
| HARDEST V13(c) guarded getter/negative room | No change: the existing -100 request already rejects returning before the clamp |

Only fixture material changed after the third full run. Every finding has a recorded outcome,
and the run returned MISMATCH 0; the next fresh reader may use the checklist's delta rule for
V01(a) and V14(e). Review completion remains pending that check. No source claim, evidence
label, rule behavior, ownership or gate changed.

### Delta blind-reader pass 4, 2026-10-15

The fresh reader recomputed V01(a)'s repeated direct-call sequence and V14(e)'s ordered
clamps from the rules and cited contracts. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0.
The exact 3413 requests of +3 followed by +1, the balances at calls 3000/3414, and the
negative-maximum clamp's -100 request all matched. Both targeted findings are resolved.
No contract change follows this check; the applicable blind-reader review is complete.
Declared scheduling, cleanup, mutation and observation gaps remain. Synthetic balances
are fixture results, not original integrated stat writes or gameplay confirmation.

### Delta blind-reader pass 5, 2026-10-15 (after dependency restructure)

Scope: cross-spec citations remapped after the foundations split (RLI) and game-data split
(GDA); no PMR rule changed. The reader received the spec, both foundations and both game-data
specs, core-record-layouts-and-ids and state-timed-state-lifecycle. Counts: RESOLVES 15,
MISMATCH-REF 3, MISMATCH 0 (all V01–V16 cases recomputed against current contracts), GUESS 0,
GAP 2, DISCRIM 2, GATED 6, HARDEST 3. Every remapped RLI/GDA/GDF/STL id resolves to the needed
contract; no remap was wrong.

| Finding | Outcome |
|---|---|
| MISMATCH-REF GDF-R02 lacks names for stats 9/26/27 | No change: GDF-R02 (pre-split stat-identity rule, same text) owns the row-index identity rule; the spec binds the explicit ids from the cited itemstatcost.txt rows. Pre-existing, unchanged by the remap |
| MISMATCH-REF GDF-R03 lacks names for states 85/106 | No change: same reason for the states.txt row-index rule GDF-R03 (pre-split text unchanged) |
| MISMATCH-REF PMR-R04 cites NRF-R01 for notation | Spec edit (citation only): PMR-R04 intro and section 9 cite foundations section 1 notation; NRF-R01 is the LCG rule |
| GAP stat/state names not in dependency specs | No change: as the two GDF rows above |
| GAP guarded maximum getter has no dependency analogue | No change: PMR-R03 owns that getter, re-read in full (D2Common 0x390A0) and stated completely |
| DISCRIM V13(b) accepts skipping the cleanup lookup when maximum is 0 | Spec edit: V13(b) trace/outputs assert one state 106 lookup returning absent, no Detach/Free, no write; V13 cites PMR-R06a. Re-checked D2Game 0x77950 on 2026-10-15: the lookup is guarded only by `max <= current` and `regen > 0` (CMP/JG, TEST/JLE before PUSH 0x6A) |
| DISCRIM V14(a)/(b) final clamp hides mulDiv branch | No change: the scaled intermediate 655000000 is already asserted |
| HARDEST V14(c)–(e), V14(a), V06/V15(c) | No change: explicit wrapped, operand-order and captured-value assertions already discriminate them |

Also tightened section 9's bare GDF-R05/GDF-R06 to GDF-R05a/GDF-R06a (death-mode identities).

### Delta blind-reader pass 6, 2026-10-15

The changed V13 expected output required a rerun; no rule behavior, ownership or gate changed,
so a delta check applies. The fresh reader recomputed V13(a)–(c) and confirmed all three
pass-5 edits. Counts: MISMATCH 0, GUESS 0, GAP 1, DISCRIM 0. The GAP (V13(b) does not restate
that a missing list makes HasState 0 and the lookup absent) is no change: the trace now states
it and RLI-R02a/RLI-R03 own those results. The reader also noted signed comparison appears in
NRF-R05 rather than section 1; PMR-R04 states signed comparison itself, so no change. The
applicable blind-reader review is complete for the restructured references.

### Focused publication review of the current contract, 2026-10-23

A fresh blind implementer read the current PMR spec and cited dependencies without this companion or earlier reviews. It recomputed the numeric and request-order examples, including V01, V04, V14(a)–(e), V15 and V16, and found no arithmetic or branch-order mismatch. It found one discrimination defect: V14(a)'s former 25000000 maximum clamped both correct and swapped NRF-R05 operand orders to the same AddStat request, so the final request could not distinguish them. The writer replaced that fixture with supplied ManaRegen 120, maximum 30000000 and current 0. A separate blind delta review recomputed the class base 10000, correct `mulDiv(65537,10000,100)=6550000`, swapped result 6553700, and unclamped final request/balance 6550000; it approved the revision without a blocker. The new values are synthetic supplied inputs, not a new table-cell or gameplay observation.

An independent publication review checked the spec, companion, gap record, dependency IDs and relative links. It found the [coverage index](../../specs/coverage.md) overstated runtime support for the trigger chain; that row now limits runtime confirmation to PMR-R02a's narrow dead-player mode-17 exit. The earlier 2026-10-12/14/15 source and blind-review records keep their dated scope. PMR-Q01–Q06/Q08/Q09 remain open as recorded below; no new original-game observation, binary reading, integrated stat write, client delivery or 1.14d equivalence was established. Draft status remains appropriate.
