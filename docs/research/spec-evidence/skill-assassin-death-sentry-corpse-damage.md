# Evidence: Death Sentry corpse damage

Spec: [Death Sentry corpse damage](../../specs/skill-assassin-death-sentry-corpse-damage.md).
Artifacts: [original source manifest](../source-manifest.md).
Limits: [gap file](../../plans/gaps/skill-assassin-death-sentry-corpse-damage.md).
Sibling: [Death Sentry corpse selection evidence](skill-assassin-death-sentry-corpse-selection.md) holds the think and corpse-selection records.

## Identity, reuse and method

Lord of Destruction Expansion 1.13c, Windows x86. References use manifest artifacts `LOD113C-D2GAME` and `LOD113C-D2COMMON`. Descriptive function names are analyst labels, not established original exported names. Module-relative RVAs and export ordinals identify matching original files; a reader can hash their own files, navigate in a PE-aware disassembler and follow these calls/writes without private reconstructed source.

The retained 2026-10-12 record used original disassembly for the reached think, corpse search, two-unit distance, percentage/roll/scale, explosion callback and area walk; original import/field/RNG readings supplied the initial arithmetic evidence. It also read the random-element table in an analysis view, while leaving its consumer insufficiently resolved. The first pass had no recorded date; it is not assigned one now. Full originals and investigation details were preserved privately before revision. No prior gameplay observation existed. The think/search/distance portion now supports the sibling spec.

On 2026-10-15 the manifest and current NRF/GDF/SFR/SCX/SLD/SCD/MSD/PVP/DMI/DPO/HRC/STL contracts were assessed for compatible artifact, method, input and execution-boundary coverage. Reused records retain their dates and limits. In particular MSD-R10's original base-value producer and SLD-R11/R16/R17's duration/router close those imported consumer questions once the actual explosion caller is traced; PVP/DMI close supplied scaling/mitigation arithmetic, not the real sentry condition or callback producers. The current PVP declaration-source record additionally establishes the same-artifact own-array wrapper used for the initial maximum-life read, as detailed below.

The available source index was checked for the targeted routines. New verification read byte-equivalent C for D2Game 0x9E2E0 and D2Common 0x34580; no listed equivalent C existed for the new main-function readings. Original instruction listings were then read for D2Game 0xA7C20/0x4F4D0/0x4F370/0xA0C70/0x9E810/0x1A40 and for the identified producer slices below. Hash-checked original-image decoding corroborated the complete encoded regions of 0x4F4D0 and 0xA0C70, including target/state order, zero skill branch, HP/length arguments, radius/center/filter branches and callback order. Reading an unresolved producer slice does not make it a complete contract. No gameplay, hooks or other-build analysis was performed.

## Dispatch and functions

SCD's current skill action chain supplies D2Game do dispatch 0xA19A0 -> table 0x1074A8, indexed by Skills.srvdofunc. The original-image slot 55 at 0x107584 contains preferred pointer 0x6FC6F4D0, establishing action 55 ->0x4F4D0. That function writes the zeroed/routed/physical record and calls area 0xA0C70 with callback 0x4F370; callback applies the copied record at 0xDE0C0 then reacts at 0xDCF20. These table/output anchors establish explosion identity independently of the unclosed AI binding (DST-Q01). The think, search and distance rows are in the [sibling companion](skill-assassin-death-sentry-corpse-selection.md#dispatch-and-functions).

| Artifact | Function/table RVA or export | Rules / semantic role | Method, date and actual boundary |
|---|---|---|---|
| LOD113C-D2GAME | 0x1074A8 slot 55 /0x107584 | DSC-R01..DSC-R14, action 55 root | SCD dispatch reuse; original slot read 2026-10-15 |
| LOD113C-D2GAME | 0x4F4D0 | DSC-R01a/R02..R09/R13 | full instructions and complete original-image corroboration 2026-10-15 |
| LOD113C-D2GAME | 0x4F370 | DSC-R09a/R11a/R14a | full instructions 2026-10-15; entire copy, signed distance compare, ordered apply/reaction, callback 1 |
| LOD113C-D2GAME | 0xA0C70 | DSC-R09a/R09b | complete original-image decode 2026-10-15; producer calls identified, not all producer interiors verified |
| LOD113C-D2COMMON | #10769 /0x34580 | DSC-R09a squared distance | equivalent C 2026-10-15; low32 deltas/products/sum |
| LOD113C-D2COMMON | #11089 /0x54A00 | DSC-R02b base HP | MSD-R10 reused 2026-10-12/14; actual HP-only caller checked 2026-10-15 |
| LOD113C-D2COMMON | #10510 /0x4E900 | DSC-R08a/R13 duration | SLD-R11 retained 2026-10-12; formal arguments/caller checked 2026-10-15 |
| LOD113C-D2GAME | 0x9EFE0; constants 0xFB970 | DSC-R13 routing/type 10 | SLD-R16/R17 instructions/table 2026-10-12 and encoded jump corroboration 2026-10-15 reused |
| LOD113C-D2GAME | 0x1180/0x11D0/0x14D0 | DSC-R03..DSC-R08/R13 | NRF dated seed/modulo/mulDiv contracts reused |
| LOD113C-D2GAME | 0xE07A0 | DSC-R01a preparation | targeted 2026-10-15: path target consistency/lookup/clear calls identified; complete effects DSC-Q04 |
| LOD113C-D2COMMON | #11157 /0x4FFD0 | DSC-R01b corpse-valid (also DST-R03b) | full 2026-10-15 instructions: kind 1/mode 12, excluded-mask overlap 0x33C20, class flag 7 through 0x1470; producer chain not complete |
| LOD113C-D2COMMON | 0x33C20 /0x1470 | excluded-state overlap/MonStats2 flag | targeted 2026-10-15 instructions; encoded reads and mask production not fully corroborated; DSC-Q04 |
| LOD113C-D2GAME | 0xA0370/0x9E610 | DSC-R09b target validity/room overlap | exact calls from full area image; complete interiors unresolved, DSC-Q04 |
| LOD113C-D2GAME | 0xDE0C0/0xDC0B0/0xDCF20 | apply/scale/reaction dependencies | reuse DPO/PVP/DMI/HRC dated contracts, including their actual gates; no local relabeling of effects |

### Import identities and caller arguments

Resolved from matching original import/export directories on 2026-10-15, or reused from the named current dependency record. Locations are evidence navigation, not required implementation storage. Thunks used only by the think/search are listed here for completeness; their rule ownership is in the sibling companion.

| D2Game thunk RVA | Original export / target RVA | Caller semantic operation / limit |
|---|---|---|
| 0xAD7E | D2Common#11089 /0x54A00 | base values(class,selector,difficulty,level,mask 1,output); HP-only |
| 0xACE2 | D2Common#10510 /0x4E900 | semantic duration(caster,skill,L); original fourth stack argument 1 is ignored |
| 0xAC9A | D2Common#10786 /0x51BF0 | calculation(caster,reference,skill,L), SLD-R19/SCX-R11 |
| 0xAB86 | D2Common#10522 /0x7FBD0 | two-unit distance(target,corpse) (DST-R04) |
| 0xAE2C | D2Common#11157 /0x4FFD0 | corpse-valid predicate |
| 0xAECE | D2Common#11106 /0x124E0 | unit-find predicate(candidate,descriptor) (DST-R03b) |
| 0xAE26 | D2Common#10769 /0x34580 | squared-distance(x1,y1,x2,y2) |
| 0xA3CA | D2Common#10945 /0x34510 | state request, RLI-R03; actual bookkeeping gates retained |
| 0xA4DE | D2Common#10158 /0x665C0 | update marking request, complete queue effects unresolved |
| 0xA3E2 | D2Common#10587 /0x38C20 ->0x38B00 | initial stat 7 own-array read, including RLI-R02b minimum adjustment |
| 0xA42A | D2Common#10973 /0x38B70 | stat 12 layer=0 |
| 0xA3F4 | D2Common#10630 /0x4F070 ->0x4E480 | skill instance request, SFR/SCX selection limits |
| 0xA400 | D2Common#10306 /0x505C0 | bonus-inclusive level request; actual bonus producer gates |
| 0xA598 | D2Common#10392 /0x348F0 | path current target, not room; NRF/SCD mapping |
| 0xA478 | D2Common#10331 /0x2FE10 | source room |
| 0xA502 | D2Common#10591 /0x12030 | collection setup/request; body not claimed verified here (DST-R03a) |
| 0xA4F0 | D2Common#10052 /0x12720 | unit-find request; actual result/filter/order DST-Q03 |
| 0xA4FC | D2Common#10344 /0x12000 | release request; not an established complete release contract (DST-R03a) |
| 0xA55C | D2Common#10383 /0x3BA20 | adjacent-room request; actual source-array production remains gated |
| 0xA3B8/0xA454 | D2Common#10750 /0x34B80; #10867 /0x34BB0 | dynamic Y/X getters, NRF-R10b |

At 0x4F59F the output pointer is pushed, then literal 1 at 0x4F5A0. The subsequent pushes 0/12/corpse belong to stat getter 0xA42A, which cleans those three arguments. Then level,difficulty,selector,class are pushed for 0xAD7E; its six stack arguments therefore carry mask 1, not mask 7. The caller initializes the whole buffer; failed base lookup leaves bounds 0. Reads of buffer first two dwords and low32 shift 7 establish `(minHP+maxHP)<<7` and raw 8.8 correspondence, closing DSC-R02b/R08b's prior asm-only limitation through MSD-R10.

At 0x4F71A..0x4F71F the caller pushes 1,L,skill,caster for 0xACE2. Callee 0x4E900's stack reads and 16-byte cleanup agree with `(caster,skill,L,condition)`; the returned duration is pushed unchanged to 0x9EFE0 with element damage/caster and two optional outputs=0. No guess based on stack cleanup alone is used. The compatible [Missile Damage Parameters instruction record](missile-damage-parameters.md#instruction-and-data-decisions) establishes that this same Common duration consumer ignores its fourth argument; only elemental endpoint consumers interpret that argument as mastery. The semantic duration request therefore contains (caster,skill,L), with the ignored legacy argument retained here solely as calling-convention evidence. SLD-R11 owns bands/synergy; SLD-R16/R17 owns type 10 choice/default length. DSC-R08a/V03/V09 omit the unused argument from their semantic requirements; no numeric result changes.

### Initial maximum-life array selection — 2026-10-15

The current [PVP declaration-source clarification](combat-pvp-damage-and-kill-credit.md#declaration-level-source-clarification---2026-10-15) already records the complete same-artifact chain D2Game thunk `0xA3E2` -> D2Common export #10587 / `0x38C20` -> `0x38B00`. Its unit/list/id/definition guards and own-array selection apply equally to the stat-7 request at D2Game `0x4F572`; the skill caller passes layer 0. That record identifies the unchanged manifest Common fingerprint and preserves the complete own-array lookup/minimum helper record's 2026-10-14 date. No additional original-source reading was necessary after this applicability assessment.

The general unit getter through thunk `0xA42A` -> #10973 / `0x38B70` -> `0x38A80` is a different operation: it can select `resolvedStats` when `usesResolvedStats` is true. The initial maximum-life request always selects `ownStats`; the later corpse/caster stat-12 requests use that effective getter. An own-array read still applies RLI-R02b to a present entry when all its conditions hold. A missing own entry returns 0 before adjustment, without falling back to a resolved entry. Existing-array consumption is verified; construction, updating and aggregation remain outside the getter boundary.

Root review found that DSC-R02a's initial wording named RLI-R02a/RLI-R02b without specifying this own-array wrapper, allowing an incorrect effective-array interpretation. DSC-R02a, inputs and the vector preamble now name the distinction. DSC-V08's non-monster isolated case supplies immutable own maxhp 25600 and resolved maxhp 51200, preserving the supported output 25600. Cases (d/e) distinguish present-entry minimum adjustment from a missing-entry zero. Monster replacement cases retain 3840 and 0. No array mutation or changing getter result is invented, and player-corpse eligibility is still not asserted. This input-selection correction requires a fresh full blind-reader pass.

### Semantic layout correspondence

| Semantic term | Original representation | Supported rules |
|---|---|---|
| kind/class/identity/seed/path | Unit+0/+4/+0xC/+0x20..24/+0x2C | DSC-R01..DSC-R14 |
| difficulty/selectors | Game+0x6D u8; selector byte+0x6A or dword+0x74 | DSC-R02b |
| calculations/radius/filter/type | Skills+0x138/+0x13C/+0x140; +0x64/+0x50/+0x1DC | DSC-R03/R06/R09/R13 |
| state 104/corpseSel | literal 0x68; MonStats2 flag bit 7 at+4, GDA-R09a | DSC-R01a/R01b |
| corpse excluded-state mask | prepared DataTables pointer+0x150; active membership bit array | DSC-R01b/Q04; mask contents not established |
| base HP bounds | first two i32 output fields at+0/+4 | DSC-R02b/R08b |
| damage copy/physical |0x70 bytes; physical+8 | DSC-R08a/R11a |
| elemental/duration/hit class | SLD companion semantic map; fire+0x10, poison+0x28/length+0x2C, hitClass+0x60 | DSC-R08a/R13 |
| area center/radius/physical threshold | corpse point, outer radius and signed squared full radius in callback context | DSC-R09a |

## Corrections and affected claims

- DSC-R04 (former range question, settled): prior prose equated max<=min with no draw. The already recorded arithmetic used n=max-min; caller 0x4F685 subtracts in 32 bits and NRF-R02 tests the resulting signed range. V09(c) distinguishes overflow; ordinary V01 results remain unchanged. The question is settled for the consumer, not a new gameplay observation.
- DSC-R05/R08a (former divide-first question, settled): prior blanket dmg>0x100000 divide-first omitted NRF-R05's divisor>(dmg>>4) wide-product alternative. Actual operand order is maxLife,calc,100 for endpoints; casterLevel,dmg,corpseLevel for scaling; pct,dmg,100 and 100-pct,dmg,100 for split. V02 preserves 750/1572840/unchanged controls and adds 314 versus 300. The shared primitive is consumed rather than copied.
- DSC-R09a: original-image area code 0xA0CB5..0xA0CD7 replaces both center coordinates when either supplied coordinate is 0. The code at 0xA0CDB..0xA0CE3 replaces filter 0 by 0x583; signed outer<=0 exits. Callback 0x4F370 retains corpse coordinates for physical. Ordinary R=7 thresholds=9/16 and R=6 equal radii remain; V04 adds the distinguishing zero-coordinate case.
- DSC-R01b/DSC-Q04 (with DST-R03a/DST-R03b): corpse search is not merely any dead monster. Equivalent callback 0x9E2E0 short-circuits corpse-valid then general unit-find predicate. Corpse-valid 0x4FFD0 checks monster/mode 12/excluded active-state overlap/corpseSel; complete producer contracts remain gated. Local request/first-result and supplied-result guards remain code-verified; their actual integrated eligibility is not promoted.
- DSC-R01a/V06: state 104/update requests precede the zero-skill check. The skill lookup changes EDI/EDX/ECX but leaves EAX holding the skill identity, so the zero-skill return is 0. Early target rejection also returns 0. Nonzero invalid record is not claimed safe. V06 tests the request order and state persistence before return.

- DSC-R10/R12/Q01/Q03: current PVP/DMI contracts settle supplied factor/resistance/absorb arithmetic. They do not establish actual sentry scaleCondition or credit, nor local target validity/callback execution. V05 becomes executable with explicit 75/100 player-versus-monster conditions instead of a predicate awaiting capture.
- DSC-R11/R14/Q02: argument 1 to 0xDE0C0 is scale-stage enable, not event identity or proof of proc suppression. The operation at 0xDCF20 is reaction, not a second generic damage apply. Existing copy/call order is preserved; V10 now asserts the complete mutable per-victim inputs and post-apply reaction record. It does not force a get-hit veto draw on a path that may not reach it.
- DSC-R13/Q05: actual duration import and argument order plus current SLD router establish the consumer. Type 10's choices{1,2,4,5}, one sentry step and nonpositive-duration 50 default are now code-verified. Expression producer gates remain; local counts do not claim whole-event budgets.
- DSC-Q04's field-name portion consumes GDF but real data/producer integration remains open. The AI-entry question is DST-Q01 in the sibling companion.
- Question on other builds: every prior 1.14d entry was untested. No other-build source or observation was authorized or collected; this remains a profile limit rather than an in-scope behavioral question.

Prior vector correction rationale remains: V02's plain product formula was replaced by exact argument-ordered mulDiv; V03's independent split loss remained unchanged. Settled range/divide-first questions and the outside-profile other-build question are recorded here rather than retained as open questions. Historical aliases were removed because no non-gap plan/implementation/test consumer was found.

## Named table cells

**TD-DSC**,2026-10-15, reuses coordinated read-only effective text reads through the running game's Storm loader. The member reported vanilla launch key lod-113c. Historical vanilla profile/earlier disabled override inventory was reused; refreshing the version registry failed. Complete current override inventory, loaded-module hashes, physical supplying MPQs and compiled table correspondence were not captured. These limits prevent upgrading text cells into compiled gameplay bindings. No gameplay/hook action was performed for this record.

Full Skills file coverage: virtual `data/global/excel/skills.txt`, offset/returned ranges 0/60000,60000/60000,120000/53535; total 173535 bytes. Joined effective-file SHA-256 `465c28e83657540e0d7ae5b0de826d5657be7fba1aa539a85c3ca3aeea34b498`. Full MonStats coverage: virtual `data/global/excel/monstats.txt`, seven 60000-byte ranges at offsets 0..360000, final 420000/12634; total 432634 bytes; SHA-256 `a15e1cb876bd9af1e8f54199c9619226cf4835c45c799aed2d432901487ecb7c`. Every chunk's successful offset/returned length was checked, joined without gaps and parsed by complete TSV header. Complete-file coverage establishes availability, not absence of arbitrary unqueried content. Rows below are selected by named identity; positional row excludes header and counts separator rows. MonStats data row 417 has explicit hcIdx=416, so raw row index is not silently used as class identity.

Each entry is one named text value, labeled verified in 1.13c data; blank is the raw empty cell, not an invented compiled 0. MonStats cells and Skills Param3/Param4 from this read support the think and are listed as TD-DST in the [sibling companion](skill-assassin-death-sentry-corpse-selection.md#named-table-cells).

| File | Row identity | Column | Raw value | Supported claim |
|---|---|---|---|---|
| Skills | `mon death sentry` (zero-based data row 312) | `Id` | `312` | named skill identity |
| Skills | `mon death sentry` (zero-based data row 312) | `srvdofunc` | `55` | text action selector only |
| Skills | `mon death sentry` (zero-based data row 312) | `calc1` | `par1` | endpoint expression text |
| Skills | `mon death sentry` (zero-based data row 312) | `calc2` | `par2` | endpoint expression text |
| Skills | `mon death sentry` (zero-based data row 312) | `calc3` | `50` | split expression text |
| Skills | `mon death sentry` (zero-based data row 312) | `aurarangecalc` | `ln34` | radius expression text |
| Skills | `mon death sentry` (zero-based data row 312) | `aurafilter` | empty string (`""`) | text cell only; code default separate |
| Skills | `mon death sentry` (zero-based data row 312) | `EType` | `fire` | linked text name; numeric compiled correspondence not established |
| Skills | `mon death sentry` (zero-based data row 312) | `ELen` | empty string (`""`) | duration text cell only |

Table cells verify data only. Their loader meanings consume GDF; compiled calculation/link identities still require correspondence. Binary roots independently establish action 55. No data cell sets runtime_observed=yes.

## Observation procedures

All pending; recipes do not constitute captures. Use an owned vanilla LoD 1.13c profile with matching manifest binary identity and separately recorded effective-data identity. Prepare an Assassin with Death Sentry at one recorded rank/character level, record equipment and relevant skill bonuses, and a single ordinary monster class/level. Work in Normal outside town at a recorded corpse/sentry/target position; capture actual input values before interpreting a result. No specific unverified table value is assumed by these recipes.

- **DSC-O01:** Keep class, corpse level and sentry rank fixed, restoring identical inputs across 1000 blasts. Record pre-mitigation endpoints, corpse seed step and raw constructed damage at D2Game 0x4F4D0, plus victim copy entry 0x4F370. Compare each traced roll with actual n and seed; observing no maximum alone cannot prove exclusion. Record enough cases to distinguish endpoint truncation and caster-below-corpse scaling. Unseen branches remain unobserved.
- **DSC-O02:** With matched fire damage use targets whose actual fire resistance is 75 then 100, with known maximum resistance/pierce/absorb and regular-monster classification. Record pre/post scale and D2Game 0xDC0B0 mitigation values, then current life immediately before/after 0xDE0C0. Distinguish player cap from monster immunity and absorb accumulator from subsequent healing; unavailable player-target eligibility keeps that case pending.
- **DSC-O03:** Record character class/level/items and actual registered event callbacks before testing thorns/leech. Compare matched cases with each effect absent/present. At 0x4F370 log the enabled apply then reaction, callback/event IDs inside 0xDE0C0 and final resources/seeds. Expected placement consumes DPO/IEP; a missing registration or missed capture does not prove suppression. Attribute callback draws separately from local corpse/sentry sites.

The two-sentry competing-claim recipe is DST-O01 in the sibling companion.

## Regeneration and review record

2026-10-15: original scope retained; private snapshots preserve original main/gap files and all investigation history. Repository search for filename and DSC prefix found sibling/companion/gap references, with no implementation/test/non-gap-plan consumer. Current numbering remains practical but aliases are not carried. Source/data corrections and their affected assertions are above; unchanged arithmetic/seeds/ordinary radii remain available. Exact vector inputs/outputs now replace symbolic predicate-only cases. The implementation comparison is retained as historical findings in the gap file, without a new implementation inspection. Mechanical lint and self-review are recorded after final edit; fresh independent review is coordinated separately. Document remains draft because actual integration and gameplay observations remain open.


## Independent review record — 2026-10-15

Recorded on the combined source spec; entries for the think rules are in the sibling companion. A separate reviewer recomputed every executable vector from the rules and cited NRF/SLD/PVP/DMI contracts and re-read the supporting original code: D2Game 0x4F4D0, 0x4F370, 0xA0C70, 0xA7C20, 0x9E810, 0x1A40, 0x11D0, 0x1180, 0x14D0, byte-equivalent C for 0x9E2E0/0x1100/0x1130, and D2Common 0x7FBD0, 0x4FFD0 and equivalent C for 0x34580/0x30F80. All vector outputs recomputed; no behavioral discrepancy was found. The action-55 slot 0x107584 was re-read in the original image as 0x6FC6F4D0.

- DSC-R04: the rule named an undefined `RandAlt` draw. The corpse-range helper 0x11D0 differs from the NRF-R02 Rand helper 0x1180 only in which register carries the modulus (signed n<=0 returns 0 without a step; power-of-two mask otherwise unsigned modulo), so the rule now cites NRF-R02 Rand directly. No vector changes.

These are wording and evidence-link edits; vector inputs, outputs and execution boundaries are unchanged.

## Blind-implementer full pass 1 — 2026-10-15

The fresh reader reviewed all 14 executable vector families of the combined source spec and every case using the semantic contract and cited dependency contracts. The report did not give an explicit case count. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 22 (eight unresolved rule entries and fourteen supplied-stage vector entries), COVERAGE 16, HARDEST 3. No behavioral correction was identified. Documented gates remain limits, not unresolved reader guesses. Entries for the think rules (COVERAGE 1-3 and 16) are in the sibling companion.

Coverage findings are advisory. The following dispositions add focused result-changing cases or ordered-effect assertions, or explain the existing/gated/equivalent recipe. No shared machinery was copied into this contract.

| Finding | Disposition |
|---|---|
| COVERAGE 4: DSC-R02a absent corpse/list/definition | Added V08(f/g/h): isolated getter or player-kind consumer returns 0 with unchanged seeds/arrays and no base request. Absent-corpse getter coverage does not assert that explosion entry accepts that corpse. |
| COVERAGE 5: DSC-R03 calculation request order | Added V01's ordered evaluator effect log with caster/skill/level/reference/result: radius 7, `calc1` 40, `calc2` 80, followed by the existing corpse roll. This rejects request reordering through exact log contents, without asserting only a pure-query count or making a new expression-producer claim. |
| COVERAGE 6: DSC-R04 negative wrapped range | Added V09(e): min -2147483648 and max 2147483647 produce wrapped n -1, no draw, unchanged seeds and physical -2147483648 at this construction stage. V09(c) already covers the opposite wrap to positive n 1. |
| COVERAGE 7: DSC-R05 caster above corpse level | Added V02(e): levels 50/40 retain damage 1000 with no draw. This rejects boosting damage when the scaling condition should fail. |
| COVERAGE 8: DSC-R06 exact clamp bounds | No change: DSC-R06 explicitly maps result 0 to pct 0 and result 100 to pct 100. These are the same observable values/duration-router branches exercised by V07(a)'s -5 and default 150 respectively. There is no distinct branch effect to add beyond those exact recipes and the stated clamp. |
| COVERAGE 9: DSC-R07 non-1 area return ignored | Added hand-off-only V04(d): supplied area returns 0 without callbacks, while final action return remains 1. Caller supplies filter 0 unchanged; local area defaulting is tested separately, avoiding a fabricated producer effect. |
| COVERAGE 10: DSC-R08a positive percentage with disabled/non-element selector | No change: V07(b) already uses pct 50/EType 0, retains physical 13151 and skips duration/router. V03(c) uses EType 13, so the router's physical addition is replaced by final physical 50; this is not an assumed fire route or an absent router call. |
| COVERAGE 11: DSC-R09a wrapping square/radius | Added local geometry V04(e): delta 65534 has signed wrapped square -262140 and passes both signed thresholds, retaining physical 1000/fire 500. Added V04(f): R 2147483647 wraps R+1, making outer -1073741824 and preventing enumeration. Supplied spatial eligibility remains distinct from real room production. |
| COVERAGE 12: DSC-R10 factor 25 | No change: owning PVP-V09(b/e) already discriminates factor 25 for two hirelings, including its priority over the boss factor, with 2560 becoming 640. DSC-R10a consumes that exact selector/arithmetic contract; it adds no local factor-25 branch or proof that an actual sentry has hireling identity. |
| COVERAGE 13: DSC-R11 candidate exclusions/downstream gates | No change: DSC-R11a starts at an accepted callback, while source exclusion/local geometry belongs to DSC-R09a and actual target eligibility remains DSC-R09b/Q04. Enabled apply's guards/effects consume PVP-R06/DPO-R13a and their owning vectors, including PVP-V03/V04. DSC-R11b/Q02 retains callback integration. Supplying a missing original producer would remove an explicit gate rather than improve this local copy test. |
| COVERAGE 14: DSC-R13 other indices/negative length | No change: owning SLD-V12(a) selects index 0/fire, SLD-V16 tests type-10 length -1 becoming 50, and SLD-R17 fixes indices 1/2 to types 2/4. SLD-V16 separately discriminates those lightning/cold routes and replaced durations/hit class. DSC-V09 already establishes the caller seed owner, ordering and nonpositive/positive duration behavior; adding the shared router's entire branch matrix here would duplicate its owner. |
| COVERAGE 15: DSC-R14 reached victim veto draws | No change: HRC-V01 directly tests HRC-R06's conditional victim-seed draws, including the 1600 two-draw and 3200 one-draw bands. DSC-R14a establishes placement/owner only if veto is actually reached; V10 deliberately supplies no-draw reaction operations. A forced veto here would claim unresolved integrated apply/reaction effects. |
| HARDEST 1: V08 own/resolved/minimum and monster replacement | No rule change: own 25600 versus resolved 51200, present-entry minimum 256, missing-entry 0 and monster replacement 3840 are separate explicit cases. Added guard cases complete the local supplied-array distinctions without implying array aggregation. |
| HARDEST 2: V04 centers/radii | No rule change: DSC-R09a explicitly separates exceptional outer-center fallback from the callback's corpse center and signed wrapped thresholds. Added overflow and ignored-area-return cases preserve that separation and their supplied boundaries. |
| HARDEST 3: V09 range/seed/duration/router | No rule change: ordered corpse then optional sentry step, supplied endpoint/length inputs and distinct seed results are explicit. Added negative wrapped range, retaining the existing routing/duration numbers and whole-event draw gate. |

The additions change vector inputs/assertions, so a fresh full pass 2 is required. Rule labels, ownership, gates, all supported existing numeric outputs and all 14 vector families remain unchanged. No original-source reading, gameplay or implementation tests were added during this triage. Mechanical lint after these edits returned 0 errors and 0 warnings; the earlier clean reader result does not complete review of the added cases.


## Blind-implementer full pass 2 — 2026-10-15

The reader recomputed all 14 vector families and 63 cases of the combined source spec, counting each primary case and independent lettered case. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 3, GATED 29, COVERAGE reported as 14 but supplied in 13 grouped entries, HARDEST 3. The gated list comprises seven questions, eight unresolved rule entries and fourteen supplied-stage vector families. No rule correction was found. Every supplied finding is disposed below or, for the think rules (COVERAGE group 1 and the think part of group 13), in the sibling companion; a fourteenth coverage item is not invented to reconcile the report's heading.

| Finding | Disposition |
|---|---|
| DISCRIM 1: V03(c) type-13 routing can be skipped without changing final physical 50 | Added an explicit router request and intermediate record assertion before the final overwrite. SLD-R16 executes with type 13/damage 50/length 0 on the zero record, giving intermediate physical 50, then the local final assignment writes 50. A skipped router fails the ordered log/intermediate assertion. This also matches the owning SLD-V16 type-13 additive-write discriminator. |
| DISCRIM 2: V05(c) zero resistance cannot distinguish scaling from mitigation order | Added V05(d) with scaleCondition true, fire 71 and fire resistance 75, all reductions/absorb/pierce zero. Scaling first produces 12, then mitigation produces 3; reversing stages produces 17 then 2. Pre-mitigation fire 12 and final fire 3 reject that reversal. The boundary remains the supplied shared stage, not an actual hit-time flag or callback producer. |
| DISCRIM 3: V09 independent seed outputs do not prove relative ordering | Added one initially empty shared timeline for the primary case, recording the actual corpse range request/draw/return, duration request/return, router request/sentry step/completed record, and final physical assignment. Exact order rejects moving the duration or either seed step even when final independent seeds/damage would coincide. Timeline observation itself draws or writes nothing and does not impose a router return-value contract. |
| COVERAGE group 2: DSC-R01a state already active, full path and invalid record | No change: active/changed state effects consume STL-R07 and its owning STL-V06; the local request order is V06. Full valid entry-to-gameplay effects retain Q04 preparation/update/eligibility integration. The spec explicitly requires a valid record for nonzero-skill damage and establishes no safe invalid-record fallback, so no guessed successful return/effect fixture is added. |
| COVERAGE group 3: DSC-R02b wrapped bounds sum/shift | No change: DSC-R02b fixes one straight-line low32 sum/shift expression, with no extra saturation/overflow branch. V08 tests replacement/source selection and raw 8.8 output; the signed low32 operation applies equally to all supplied bounds. Another large-number example would repeat that fully specified arithmetic recipe rather than resolve an unknown branch or discrimination failure. |
| COVERAGE group 4: DSC-R03 negative/overflow mulDiv | No change: NRF-V09(a/k) already discriminate truncation of negatives and wrapped narrow multiplication, while its other cases cover the operand/divisor branches. DSC-V01/V02 separately pin this caller's operand order and divide-first values. Copying the entire shared primitive matrix would duplicate NRF's owner. |
| COVERAGE group 5: DSC-R04 positive wrapped range and addition overflow | No change: V09(c) already has wrapped positive n 1 despite reversed signed endpoints; V09(e) has wrapped negative n -1 despite ordered endpoints. DSC-R04 explicitly low32-wraps min plus the returned draw; no additional clamp, guard or undefined arithmetic remains for an overflowed sum. The existing caller/primitive recipes remain acceptance, without claiming exhaustive values. |
| COVERAGE group 6: DSC-R07 nonzero area result | No change: V04(d) now proves the final caller result 1 when supplied area returns 0, rather than only checking an earlier callback result. DSC-R07 contains no test on the returned value; every other supplied value follows that same ignored-result recipe. No meaningful original area success integer is inferred. |
| COVERAGE group 7: DSC-R08a intermediate physical route/EType 0 interior | The intermediate type-13 assertion is added for DISCRIM 1. EType 0 with positive pct 50 already appears in V07(b), producing physical 13151 with no duration/router; pct 0 with nonzero EType 10 also skips routing in V09(c). No additional disabled-branch fixture is needed. |
| COVERAGE group 8: DSC-R09a nondefault filter, absent room, invalid target, source exclusion and nonzero centers | No change: V04's ordinary cases already use two nonzero center coordinates, and (b) separates the exceptional centers. DSC-R09a explicitly retains a nonzero filter, exits without callbacks for absent room, skips the sentry before distance/validity, and calls the callback only after nonzero target validity. Real room/filter/target production remains DSC-R09b/Q04; these supplied-result handling recipes are complete without fabricating actual eligibility. |
| COVERAGE group 9: DSC-R10a/DSC-R12a other classes and scale/mitigation | Scale/mitigation order receives V05(d) for DISCRIM 2. PVP-V09 owns other factor/class pairings and DMI owns their type-specific resistance/absorb rules. V05 already distinguishes player cap and regular-monster immunity. Actual sentry condition/credit/full hit effects retain Q01/Q03. |
| COVERAGE group 10: DSC-R11a rejected callback/later victim | No change: DSC-R11a begins only for a candidate accepted by DSC-R09a and uses the valid callback context supplied by DSC-R07. Its valid callback returns 1 after both calls. V10 explicitly executes a second victim after the first apply mutates every asserted field, and checks the second fresh input plus both post-apply reaction records. Gated dependency effects are not filled by an artificial rejected-context path. |
| COVERAGE group 11: DSC-R13 types/indices 1/2/4 | No change: SLD-R16/R17 owns the random selector and routing table; SLD-V12(a) covers index 0/type 1, and SLD-V16 discriminates type 2/lightning and type 4/cold routing. The fixed selector table establishes indices 1/2 as types 2/4. DSC-V09 owns caller context/seed/timing, now including the shared ordered timeline. |
| COVERAGE group 12: DSC-R14a reached veto | No change: HRC-V01 tests exact reached-veto seed/draw outcomes. DSC's actual reachability through apply/reaction remains gated, and DSC-R14a explicitly conditions those draws on reaching HRC-R06. V10's no-draw supplied reaction cannot establish a full-event veto count. |
| COVERAGE group 13 (explosion part): DSC-R02a/DSC-R05/DSC-R06 principal branches | No change: V08 own-array/guards/minimum/replacement; V02 scaled/equal/zero/higher levels; V07 clamp/disabled element. No specific new unresolved output or accepted wrong alternative was identified for this grouped advisory. |
| HARDEST 1: V04 geometry | No rule change: signed wrapping, two centers, two radii and supplied room boundaries remain explicit. V04(d/e/f) separately exercise ignored caller return, squared overflow and radius overflow without promoting spatial producers. |
| HARDEST 2: V08 own-array/minimum/monster replacement | No rule change: distinct immutable own/resolved arrays and present/missing-entry minimum behavior precede monster replacement only where applicable. All guard cases remain isolated consumer fixtures, and eligibility/aggregation are not inferred. |
| HARDEST 3: V09 seed/order | Added the shared timeline for DISCRIM 3. It observes cross-operation order directly while preserving all existing seed/record outputs and the conditional whole-event draw limit. |

The revised contract has the same rule behavior, labels, ownership and gates. The added V05(d) case and V03(c)/V09 assertions require a fresh full pass 3; the clean arithmetic result from pass 2 does not approve those changes. No original-source reading, gameplay or implementation test was performed during triage. Technical-token checks and mechanical lint are rerun on the resulting files.

## Blind-implementer full pass 3 — 2026-10-15

The fresh reader recomputed all fourteen executable families of the combined source spec and every subcase from rules and inputs, ignoring their traces. No numeric case count was reported for this pass. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 29, COVERAGE 15, HARDEST 3. GATED retains seven question entries, eight unresolved rule entries and fourteen supplied-stage families. Every applicable blocking finding is resolved on this contract. This completes independent documentation review; status remains draft and maintainer approval is pending. Entries for the think rules are in the sibling companion.

| Advisory entry | Disposition |
|---|---|
| DSC-R01a full valid construction | No change: Q04 retains actual preparation, selection and state/update integration. Local construction stages have their own supplied-input vectors; joining them would not verify the producer. |
| DSC-R02a usesResolvedStats false | No change: the initial wrapper always uses ownStats. V08 deliberately sets true with conflicting resolved value, rejecting effective-array selection; false cannot reveal an additional wrapper branch. |
| DSC-R03 zero/negative/wrapped percentages | No change: the shared NRF-R05 primitive and its vectors own signed and overflow branches. DSC-V01/V02 pin these callers' argument order and divide-first distinctions; no duplicate primitive matrix is added. |
| DSC-R05 negative corpse level | No change: the signed caster-below-corpse condition is explicit. V02(b/c/e) already distinguish scaling from equal, zero and higher-caster no-scale outcomes; no new producer or level clamp is inferred. |
| DSC-R07 nonzero area return | No change: V04(d) checks ignored area return 0 against final action return 1. All supplied nonzero returns use that same ignored-result recipe; no area integer success contract is invented. |
| DSC-R08b HP sum/shift overflow | No change: this straight-line low32 sum/shift has no separate conditional branch. V08 pins units and replacement order; the stated wrapping operation remains required for all values. |
| DSC-R09a Y-zero and absent room | No change: either-coordinate-zero predicate is explicit and V04(b) distinguishes its exceptional center. The absent-room no-callback recipe is specified; actual room eligibility remains Q04. |
| DSC-R11a rejected candidate | No change: DSC-R11a starts with an accepted candidate. Rejection is owned by DSC-R09a/DSC-R09b and its actual spatial/validity producer remains Q04; no fresh-copy/apply/reaction assertion is claimed for a rejected callback. |
| DSC-R12a positive absorb | No change: DMI owns absorb arithmetic and its discriminating vectors; DSC-V05 owns caller scaling/cap/order, not a duplicate mitigation matrix. |
| DSC-R13 other random indices | No change: SLD-R16/R17 and SLD-V12/V16 own selector mapping and elemental writes. DSC-V09 discriminates this caller' sentry seed, duration/context and cross-operation order. |
| DSC-R14a reached veto | No change: HRC-R06/HRC-V01 own exact conditional victim draws. DSC-V10 supplies no veto-reached assertion; its complete reaction integration/draw total remains gated. |

HARDEST, all no change: V04/DSC-R09a explicitly separates signed wrapping and two centers; V09/DSC-R13 now records corpse draw, duration, sentry draw and final assignment in one ordered timeline; V08/DSC-R02a/b separates own-array minimum from monster-bound replacement. Their exact assertions and gates remain unchanged. No vector, behavior, input, expected output, acceptance assertion or execution boundary changed in this final triage, so no further reader run is required. Mechanical lint and token/preservation checks are repeated before publication. No gameplay, implementation tests or new verification promotion occurred.

## Blind-implementer delta check 4 — 2026-10-15

Run type: delta check, because full passes 1-3 precede the independent review edits. A fresh reader received the combined spec, foundations, game-data-fields and the cited dependency specs, plus the changed material and the finding each change resolves: DSC-R04's draw now cites NRF-R02 Rand instead of the undefined `RandAlt` (vectors DSC-V01, DSC-V09), and the think rules' entry-gate edit recorded in the sibling companion. The rule text, vectors, outputs and execution boundaries did not change, so no rule's behavior or ownership changed after the last full run.

The reader recomputed every case of V01 and V09 (and the think vectors recorded in the sibling companion) from the rules and inputs alone. Both listed findings are resolved. Counts: MISMATCH 0, GUESS 0, GAP 1, DISCRIM 0, GATED 3 (the AI-entry, local-producer and duration-producer questions, now DST-Q01, DSC-Q04/DST-Q03 and DSC-Q05). The reader confirmed that V09(a/e) reject a step on signed n<=0 or an unsigned n test, V09(c) rejects an independent max<=min test and a skipped step for n=1, V01 rejects masking a non-power-of-two n (863 instead of 2911) and a wrong seed owner, and mask versus modulo for a power-of-two n is mathematically identical. The single GAP concerns a think rule and is disposed in the sibling companion.

Outcome: MISMATCH 0; no unresolved GUESS or DISCRIM; the one GAP has a recorded no-change outcome. No spec edit was made in this triage. Status remains draft.

## Restructure (2026-10-15)

The approved review split the combined Death Sentry spec at its two independent triggers: the AI think and the Skills action-55 explosion. This spec keeps the explosion: guards and corpse-valid producer, initial and monster life, endpoints, roll, level scale, percent, hand-off, split and units, geometry and traversal, scaling, copies and events, mitigation, duration/router and draw counts, with ten executable vectors and three observation recipes. The think and corpse-selection rules, their four vectors, the AI-entry and competing-claim questions and the two-sentry observation recipe moved to [Death Sentry corpse selection](../../specs/skill-assassin-death-sentry-corpse-selection.md) (DST). The former combined local-producer question was divided into DSC-Q04 (explosion producers) and DST-Q03 (think producers). The state-order vector stays here because it asserts DSC-R01a's request order; it also cites DST-R06a. No rule behavior, label, value or vector result changed; ids were renumbered under the unconsumed-spec policy. Records above were partitioned by owning rule; their dates and counts are unchanged.

## Publication review — 2026-10-23

A fresh independent blind reader reviewed the complete post-split DSC and DST
three-file contracts. For DSC, the reader recomputed representative corpse-seed rolls,
level scaling, independent split, wrapped blast geometry, PvP/mitigation order,
monster life and elemental routing across DSC-V01–V10. It confirmed that DSC-R01a
requests state 104 after the selection/mode request, while actual scheduler and corpse
eligibility producers remain gated. No blocking rule/vector mismatch was found within
the 1.13c static/supplied-input boundary. DSC-V01–V10 and DSC-Q01–Q05 are the current
IDs; the 2026-10-15 combined-contract review counts above remain dated history.

An independent evidence reader checked both complete three-file sets and the source
manifest for artifact identity, dated reuse, named text-cell limits, public links and
unresolved producers. A separate architecture reader checked DST/DSC ownership and
dependencies. All three approved bounded publication after split-stale SFR and GDA
consumer references were corrected. The historical DST gap comparison was narrowed to
the state request that the current rules support. Issue #498 now names both three-file
handoffs; #420 remains the gameplay owner.

DSC-Q01–Q05, DSC-O01–O03 and the sibling DST gates remain open. No spec behavioral rule or vector output changed. The historical comparison status
wording was narrowed; original-source evidence labels were not changed, and the
implementation comparison was not renewed. No new original-image investigation, protected game-data capture,
gameplay test or 1.14d confirmation was added. Both specs remain draft.
