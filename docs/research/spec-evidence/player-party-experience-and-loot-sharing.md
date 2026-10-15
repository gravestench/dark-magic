# Party experience and loot-sharing evidence

This companion supports the [behavioral contract](../../specs/player-party-experience-and-loot-sharing.md). Original artifacts resolve through the [source manifest](../source-manifest.md). Function descriptions are analyst labels; an original export ordinal identifies the exported entry but does not validate an analyst-assigned name. All addresses below are module-relative RVAs unless explicitly stated as VAs. Private source forms, fixture encodings, complete original documents and collection details are preserved separately.

## Identity and evidence reuse

The historical static readings date to 2026-10-11 and 2026-10-12. They covered the death wrapper, credit selection, collection, scaling, award and split callers listed below. The 2026-10-12 independent audit corrected B[7], multiply/divide branching, hireling hook/award gates, member filtering and gold excess routing. Compatible caller arithmetic and request-site evidence is reused with those dates; a new format does not renew a reading or make unread callees verified.

The current original images match manifest artifacts `LOD113C-D2GAME` (SHA-256 `24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53`, 1138688 bytes) and `LOD113C-D2COMMON` (SHA-256 `59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10`, 679936 bytes). Complete narrow original-image checks on 2026-10-15 use those identities. This establishes analysis-artifact correspondence, not historical loaded-DLL or data-profile equivalence.

Compatible records reused: NRF-R05 (arithmetic reread 2026-10-14; wide division support dated 2026-10-13), NRF-R10b dynamic coordinate extraction, NRF-R12b/R12c existing-array reads dated 2026-10-14, GDF field/name/table records dated 2026-10-14/15, MSD-R21 iterator/alive-count record dated 2026-10-14, and PVP-R10a/R10b/R11a/R11b refreshed 2026-10-15. [NRF evidence](core-numeric-and-rng-foundations.md), [GDF evidence](core-game-data-fields.md), [MSD evidence](monster-scaling-by-players-and-difficulty.md) and [PVP evidence](combat-pvp-damage-and-kill-credit.md) retain their own methods and limits. Getter coverage supplies existing values, never additive/replacing writes. [STL-R06c](state-timed-state-lifecycle.md) and NRF-Q20 preserve actual propagation gaps.

## Original function references

| Artifact | Function RVA / original location | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | 0xDEEE0 → 0x777F0 → 0xDEA00 | none | Death, experience wrapper, routing | R01a/R03/R08a/R13 | 2026-10-11/12 caller reading; consume PVP-R10a/R11a shared contracts |
| LOD113C-D2GAME | 0x15D0 | none | Monster flag read | R01a | Historical named killable consumer; GDF-R18 identity |
| LOD113C-D2GAME | 0xDB5E0; 0xE03A0 | none | Credit selector; unit registry lookup | R02a/R02c | PVP-R10a/R10b 2026-10-15; no new registry-production claim |
| LOD113C-D2GAME | 0xDFA60; 0x9BEB0 | none | Player party id | R03/R04a/R09a | Historical consumers; MSD-R21 dated 2026-10-14 |
| LOD113C-D2GAME | 0xDE8F0; 0xDAED0 | none | Party pool; XP collector | R04a/R04b/R13 | Historical caller/filter audit; pool instructions rechecked 2026-10-15, floating environment gated |
| LOD113C-D2GAME | 0x9C070; 0x9BD00; 0x9C000; 0xDFA90 | none | Same-level iterator, party record, alive callback/test | R04a/R09a/R11 | MSD-R21 2026-10-14; complete 240-byte iterator original-image check 2026-10-15 corrects null/empty-record branch |
| LOD113C-D2GAME | 0x1100; 0x1130 | none | Position wrappers | R04a/R04c | Checked source reconstructions and collector original-image route 2026-10-15; dynamic extraction consumes NRF-R10b; static world conversion remains NRF-Q04 |
| LOD113C-D2GAME | 0x2A3C | none | Fatal-error request | R04a | Historical ninth-member branch; not a silent cap |
| LOD113C-D2GAME | 0xF4FE8; 0xF5724 | none | Alternate divide; float conversion | R04b | Alternate divide interior unread; complete 117-byte conversion/272-byte pool listings checked 2026-10-15; full environment remains Q07 |
| LOD113C-D2GAME | 0xDC030; 0xDAA40; 0xDA9E0; 0xDB6B0 | none | Scaling, level factors, ratio, player/hireling hook | R05a | Historical 2026-10-12 arithmetic; operand order and unsigned cap rechecked through callers 2026-10-15 |
| LOD113C-D2GAME | 0xFA01C; 0xFA048 | none | Level-factor arrays A/B | R05a | Historical instruction/data consumer reading, 11 entries each |
| LOD113C-D2GAME | 0x14D0; 0xF44B0 | none | Multiply/divide and wide support | R05a | Consume NRF-R05; no duplicate helper authority |
| LOD113C-D2GAME | 0xDE6D0; 0xEB090; 0x37030 | none | Player award caller; level-up; resource callback | R06a | Complete award original-image instructions 2026-10-15 prove own-array read and ordered requests; level-up/resource interiors Q07 |
| LOD113C-D2GAME | 0xDDCB0; 0x8BB10; 0x48BA0; 0x48950; 0xDFFD0 | none | Hireling award caller and notification/update requests | R08a | Complete 304-byte original-image/entry-to-return reading 2026-10-15; request/scan arithmetic only; downstream updates Q07 |
| LOD113C-D2GAME | 0xD6E30; 0xD6C80; 0x9C270 | none | Pickup dispatch; gold caller; party split | R09a/R13 | Historical dispatch; complete caller/split listings and image correspondence 2026-10-15; amount/request boundary |
| LOD113C-D2GAME | 0x9BC70; 0x1670 | none | Capacity outputs; bounded add request | R09a/R09d | Complete capacity 64-byte image and both instruction listings 2026-10-15; actual propagation not established |
| LOD113C-D2GAME | 0x12B90 | none | Excess-gold creation lead | R09d | Historical selected item-code/chunk/room observations; placement/creation/interior draws unread |
| LOD113C-D2GAME | 0xD2F00 | none | Item validation lead | R10 | Historical partial path; no universal policy absence |
| LOD113C-D2GAME | 0xDBA40; 0xDDAC0; 0x99210 | none | Difficulty/corpse-experience leads | R12a/R12b | Historical candidate reads, not complete penalty/producer proof |
| LOD113C-D2COMMON | 0x5E7E0 | 10066 | Header class maximum | R05b | Complete 48-byte disassembly and original image 2026-10-15; class clamp/read |
| LOD113C-D2COMMON | 0x5E810 | 10399 | Class threshold | R05b | 2026-10-15 reconstruction checked against original machine code, read first; class clamp and row L+1 |
| LOD113C-D2COMMON | 0x5E840 | 10989 | Ratio accessor | R05b | 2026-10-15 reconstruction checked against original machine code, read first; null/index/header branches |
| LOD113C-D2COMMON | 0x608D0 | 10670 | Experience-to-level scan | R06b | Complete 80-byte listing 2026-10-15; unsigned comparisons, row advances, maximum bound |
| LOD113C-D2COMMON | 0x60920 | none | Experience loader | R05c | GDF-R04 dated record; checked reconstruction reread 2026-10-15; field descriptor and publication do not establish each Level-row correspondence |
| LOD113C-D2COMMON | 0x532C0 | 11156 | Hireling record selector | R05d/R08a | Complete 160-byte listing/image 2026-10-15; selector inputs/returns located, prepared record production still Q07 |
| LOD113C-D2COMMON | 0x2C990 | 10448 | Hireling threshold multiplication | R05d | 2026-10-15 reconstruction checked against original machine code, read first; low products L*(L+1)*L*factor |
| LOD113C-D2COMMON | 0x31940 | 10049 | Gold cap | R09b | Complete 96-byte listing and image 2026-10-15, effective getter consumes NRF-R12b/c |
| LOD113C-D2COMMON | 0x38C20; 0x38B70; 0x38A80 | 10587; 10973; none | Own/effective/existing-array stat reads | R05a/R06a/R08a/R09a/R09b | NRF-R12b/c; own/effective wrapper distinction reuses PVP 2026-10-15 |
| LOD113C-D2COMMON | 0x3A740; 0x3A710; 0x3A330; 0x3A280 | 10887; 10551; none; 10261 | Replacing/add forwarding and propagation | R06a/R08a/R09d | Requests resolved 2026-10-15; additive wrapper checked reconstruction read, downstream propagation/list integration gated Q07/STL/NRF |

## Dispatch, imports and semantic mapping

The compatible PVP/DPO root is the server frame-event/skill/basic attack → damage build → queue → apply → death chain. Its root/output anchors retain the dates and full locations in [PVP dispatch evidence](combat-pvp-damage-and-kill-credit.md#dispatch-and-output-anchors). Within this contract death RVA 0xDEEE0 reaches wrapper 0x777F0 and routing 0xDEA00. Routing uses 0xDB5E0, then 0xDC030/0xDDCB0 and solo 0xDE6D0 or party 0xDE8F0. Party invokes 0x9C070 with callback 0xDAED0; the final award requests return through 0xDE6D0. Pickup 0xD6E30 → 0xD6C80 → 0x9C270 selects count callback 0x9C000, capacity 0x9BC70 and add 0x1670; positive returned excess reaches 0x12B90. Complete pickup deletion/notifications outside the amount handoff are not promoted.

PE imports/export correspondence was resolved on 2026-10-15 against the identified original images:

| D2Game thunk | D2Common export / target RVA | Semantic operation |
|---|---|---|
| 0xA3D6 | #10399 / 0x5E810 | Class threshold |
| 0xA48A | #10066 / 0x5E7E0 | Class maximum |
| 0xA784 | #10989 / 0x5E840 | Ratio |
| 0xA4E4 | #10670 / 0x608D0 | Level scan |
| 0xA3BE | #11156 / 0x532C0 | Hireling record lookup |
| 0xA3A6 | #10448 / 0x2C990 | Threshold factor arithmetic |
| 0xA45A | #10049 / 0x31940 | Gold cap |
| 0xA3E2; 0xA42A | #10587 / 0x38C20; #10973 / 0x38B70 | Own versus effective stat reads |
| 0xA472; 0xA4D2 | #10887 / 0x3A740; #10551 / 0x3A710 | Replacing versus additive requests |

The PVP companion's exact four redirect imports and active-state/first-primary-chain contracts replace generic list-existence inference. R02 consumes those contracts, including separate owner-link and supplied-game lookups. No second owner hop is implied.

Original layout evidence maps neutral terms as follows: Unit +0xC4 bit 26 is skipKillExperience; Unit +0xC6 bit 0 is excludedFromPartyCollection; Unit +0xC8 bit 10 with +0x94/+0x98 lookup supplies ownerExceptionCondition. PlayerData +0x44 is the active-entry lead; its entry first-byte bit 0 is the neutral entry-exclusion condition. MonsterData +0x28 is a pointer to the owner record, whose +0x28/+0x2C/+0x30 supply game/id/type. These are evidence encodings, not required engine layouts. PetType count is DataTables +0xBF0. Experience prepared records have eight i32 fields per row, seven classes then ratio; the header class values supply maximums and header ratio the default shift. The ratio positive index uses row index+1. ThresholdFactor is the selected hireling record field +0x20; production of that record remains gated.

## Corrections and affected claims — 2026-10-15

1. **Missing party record (R04a).** The 2026-10-11/12 PXL wording grouped no matching party record with early solo fallback. MSD's full 2026-10-14 reading instead recorded zero callbacks. A complete original-image check of 0x9C070 (240 bytes) distinguishes them: call 0x9BD00 at 0x9C0CE; its null result takes 0x9C0D5 → epilogue 0x9C0C4, not solo callback 0x9C0BB. Empty record member list also branches to the epilogue. Earlier room/player-data/no-party guards select the solo callback separately. V10(d/e/f) now distinguish zero from one callback and resulting awards. MSD's normalization made counts zero/one both eligible 1; it did not validate PXL's old award claim. Record lookup/room/registry producer coverage is not renewed by this branch check.
2. **Redirect meaning (R02c).** Generic stat-list existence and arbitrary matching lists were unsupported. PVP's 2026-10-15 original-image/import evidence establishes active-state mask intersection and first primary-list selection, excluding base/secondary lists; R02 now consumes it. V08 retains prior outcomes and adds explicit base/secondary, eligibility and failed-lookup controls.
3. **Getter versus write (R06a/R08a/R09a).** Prior wording stated mutated stats/full level-up while local caller evidence established requests. Exact XP current-value calls use #10587 own-array reads; member bonus/owner level/gold read effective values. New caller instructions and imports distinguish those. Actual #10887/#10551 effects, level-up, event-12 callback, hireling update and recreation remain integration gates; vectors assert requests or explicitly synthetic stores. No getter proof is reused as mutation proof.
4. **Floating precision (R04b).** Original wording assumed default nearest rounding. Original pool stores binary floating point with 32-bit storage but has an alternate division branch through 0xF4FE8; process control/precision and complete conversion extremes were not established. The actual-production label is unresolved. V04/V10 retain the exact illustrative quotient and numbers under explicitly supplied environment/raw inputs, rather than fitting unknown runtime rounding.
5. **Arithmetic/accessors.** R05 consumes NRF-R05 with original operand order, avoids universal exact-small-input claims, preserves shift-count masking and signed narrow wrapping. Matched original-code accessors plus new max/scan/threshold/cap readings support R05b/R05d/R06b/R09b at supplied-record boundaries; these labels become verified static for those assertions only. V11/V12/V14/V15 add corresponding discriminators. The hireling caller also requests an experience-change notification after stat replacement and before scanning thresholds; its update/progression/notification/resource interiors remain gated. Non-hireling 86/256 remains signed low-product truncation before doubling.
6. **Search/RNG scope.** R02b/R10/R12a had unretained complete six-step absence searches. They are unresolved, rather than newly claimed absent. Historical candidate observations remain below. No-direct-draw is restricted to local supplied-input calculations/requests; unread callback/update/creation interiors and other death effects are excluded.

The 2026-10-12 V02 correction remains 878 →679 because difference 7 selects B[7]=174, not B[6]=225. Original bytes and all earlier aliases/fixture encodings are retained privately. No renumbering was needed: repository consumption search found no qualifying implementation/plan citation; useful current ids remain without historical aliases.

## Data records and profile limits

GDF-R04's 2026-10-14 record reports full effective `data/global/excel/experience.txt` coverage 7116/7116 bytes and compiled `experience.bin` coverage 3236/3236. It verifies the field schema and file coverage, not every class/Level/ExpRatio alignment. The earlier maximum-cell record was only a 4096-of-7116-byte text read with no full-file hash. It cannot by itself establish full correspondence or absence. No original maximum/ratio cell is required by these supplied fixtures, so no cell is promoted to a gameplay or loaded-record assertion. Q01 retains the precise remaining comparison.

GDF's 2026-10-15 complete effective named-stat/mode/type records corroborate identities within its reported vanilla LoD 1.13c profile. The version catalog refresh failed; a complete current override inventory, loaded DLL identities and physical supplying archives were not established. Previously disabled override is not a complete inventory. Effective hashes identify those files only; they do not prove equality to historical compiled data. This contract consumes only the supported names/identities, retaining their dates and limits. No new table read or gameplay operation ran in this pass.

## Gap assessment — 2026-10-15

- Q01: checked GDF's existing full text/compiled coverage, partial legacy maximum record, loader descriptor/publication and complete max/threshold/ratio/scan consumers. This closes accessor-control-flow uncertainty but not named original text/prepared correspondence; no invented cells.
- Q02/Q04/Q08: checked current GDF/PVP identity/selector/writer limits. Known tests remain supplied-input contracts; wider meanings/producers are not guessed.
- Q03: reused NRF getter/world-scale split; raw distance fixture requires no yard conversion.
- Q05: checked PVP's separately established contacts/creator path and death-routing boundary. Complete per-source killer production remains outside this consumer contract.
- Q06: retained the historical death-penalty searches among identified difficulty/leech/creation candidates and pickup validation. Current PVP/DPO/IEP/OWD contracts establish bounded upstream roles, not complete drop/timer/penalty/corpse policies. The unretained search record cannot support absence. Broad quest/recovery/drop campaigns are separate work.
- Q07: newly read complete narrow prepared accessors, record selector, threshold arithmetic, player/hireling callers, gold capacity/add and cap. Actual propagation remains under STL/NRF; local original unit setter wrapper and additive forwarding still reach unresolved list mutation. Level-up/resource/hireling/recreation request locations identify the smallest remaining interiors. Original pool/conversion listings expose the unestablished alternate divide and environment. These are explicit gates, not failures silently accepted as proof.
- Observation: no matching original-game award/split observation was made or reused as a match. No possible observation is declared impossible merely because this migration excludes gameplay.

## Observation procedures

All five procedures are pending. Use owned matching Expansion 1.13c artifacts/data, a recorded character class/level/stats, known party/room inputs, pre/post raw values and attributable reached-stage records. Record capture completeness and measured assertion scope. An archive read is not gameplay confirmation.

- **PXL-O01:** use two then three players on one map, known levels matching V04/V10, no XP bonus and a victim with measured stored XP/level. Record member coordinates/rooms/modes and actual pool/share/raw/award values at Game 0xDAED0/0xDE8F0/0xDE6D0. Repeat just inside/outside raw distance 6400, another map and death mode. Record floating control/environment and branch before comparing exact produced shares; otherwise compare supplied downstream amounts only.
- **PXL-O02:** use a player with summon/golem/trap/hireling and an attributable damage-over-time source against a measured victim. Record killer entering 0x777F0/0xDEA00, one-hop/active-state/list inputs at 0xDB5E0 and selected player. Direct player and victim-list controls distinguish bypass/override. Do not infer unobserved source attribution.
- **PXL-O03:** use a hireling below owner level, measured record/thresholds and own XP; kill once by hireling and once by owner. Read 0xDDCB0 amount/guard inputs and stat requests, actual XP/level/resources and notification effects. Repeat owner-level equality and binding hook cap; expect V05/V14 arithmetic only with matching supplied inputs.
- **PXL-O04:** use three same-level party players, measured level/gold/caps, then pick known gold amount. Record divisor/recipient lists, cap/accepted/excess at 0xD6C80/0x9C270/0x9BC70 and actual stat storage. Repeat far/dead/other-level/cap controls. At excess 851 verify recreation entry 0x12B90 separately from placement/item count/seed changes; no creation outcome is assumed.
- **PXL-O05:** record complete effective Experience header/target text rows through original loader and corresponding prepared values. With those known, choose an ordinary player kill whose level ratio/bonus/cap are controlled; record 0xDC030 accessor results and 0xDE6D0 requests/stores. Expected original cell values must be established before the award comparison.

## Migration and self-review record

The complete original main and historical gap bytes were archived and hashed privately before editing. Consumption search covered spec filename and PXL prefix in implementation source/tests and plans outside gaps; no qualifying citation was found. Existing 13 parent-rule,18 vector, 5 observation and 8 question families remain. Public locators/decision rationale remain here; private reconstruction paths, commits, raw object fixtures and collection history remain in the archive. Practical ids are retained without aliases. Historical implementation findings in the gap companion remain unrefreshed.

Self-review checks the language-independent contract, preserved supported numbers, request/store distinction, all rule labels/gates, exact supplied-stage vector inputs/results and dependency directions. Fresh blind-reader review is coordinated separately; this record does not claim its completion or user approval. Document lint and stable handoff hashes are recorded once the final owned files are ready. No implementation test, code change, gameplay operation, Git publication or scope expansion was performed by this writer.

### Writer checks - 2026-10-15

The spec checker ran for this spec and its companion/gap record: 0 errors, 0 warnings. The writer independently recomputed the retained factor/share/cap/threshold cases and reviewed the exact integer widths, unsigned comparisons, no-callback/solo distinction and request/store gates. V17 names distinct seed-owning units so its seed fixtures cannot assign two values to one object. Original historical implementation findings are retained without a new comparison. These are document checks and writer self-review; no independent reader result or gameplay match is inferred.

## Blind-implementer full pass 1 - 2026-10-15

A fresh reader reported MISMATCH 0, GUESS 0, GAP 0 and DISCRIM 0 for 18 vector families. It reported 66 cases, but its own per-family breakdown (6, 6, 9, 1, 11, 10, 5, 6, 7, 7, 1, 2, 2, with V01-V06 grouped) sums to 73. No verified case count is adopted. It reported GATED 17 while rendering 19 grouped limits, and COVERAGE 14 while grouping its advisory branches under five families. The actual rendered limits/findings are disposed below rather than replaced by invented totals. HARDEST had three entries. No numeric or blocking ambiguity was reported, but the subsequent coordinate clarification and added fixtures require a fresh full pass 2.

### Documented gates

All rendered limits receive no change: they are already stated at their exact boundary. R01b's meaning is Q02; R02b's attribution is Q05; R02c's producers are Q08; R03's stores are Q07; R04b's floating production is Q07; R04c's world conversion is Q03; R05a's unsafe indices and producers are Q07/Q01/Q08; R05b/R05c's original correspondence is Q01; R05d's record preparation is Q07; R06a's stored/level/resource effects are Q07; R06b's original data is Q01; R07 excludes unread interiors; R08a integration and R08b active-entry production are Q07; R09a's effects are Q07; R09c/R09d meanings/storage/recreation are Q04/Q07; R10 drop policies are Q06; R12a/R12b penalty/corpse behavior is Q06; R13 excludes dependency interiors. This is a record of existing gates, not new promotion, a substituted count, or approval. Pending gameplay observations remain separately open.

### Advisory coverage and difficulty triage

| Rendered finding | Outcome |
|---|---|
| R04a collection exclusion | Added V10(k), explicit exclusion before collection |
| R04a non-player/monster callback filtering | Added isolated callback V10(l), rejected type 3 versus admitted monster type 1 |
| R04a monster death branch | V10(l) contrasts monster mode 12 with mode 1; existing V10(c) retains player mode 17 |
| R04a absent victim distance bypass | Added V10(m), absent victim admits an otherwise zero-X member |
| R04a map/registry misses | Added V10(n), ordered record with another-map member and missing registry identity, only P callback |
| R04a nonpositive sum | Added V10(o), admitted count 2 with sum 0 returns before sharing |
| R04a effective-count independence | Added V13(a), party count 3 unchanged for effective count 1/8 |
| R05a negative input | Added V11(j), negative early return before maximum check |
| R05a non-player hook rejection | Added V11(k), type 3 and absent member return 0 after supplied positive scaling stages |
| R05a m>=25 with v<=0 ratio condition | No change: reaching the ratio branch already requires v>m; with m>=25, v<=0 cannot reach it. The lower/equal-victim factor recipe is already explicit and exercised |
| R05a bonus helper faults | No change: this caller fixes divisor 100, while level-factor divisor is 256 and ratio divisor is positive v. NRF-R05/NRF-V04(g) owns general INT_MIN/-1 fault behavior; that divisor is not supplied by the bonus call. Existing V18 tests caller operand order/divide-first arithmetic |
| R05b/R06b below initial threshold | Added isolated V12(f), supplied experience 99 below initial threshold 100 yields level 0 |
| R08a nonpositive post-share amount | Added update-entry V14(h), no stat or notification/update requests |
| R08a missing record | Added update-entry V14(i), no write or notification/update requests |
| R08a maximum-level gate | Added V14(j), level 98/owner 99/max 99 isolates max-1 rejection after other gates pass |
| R09a owner exception success | Added V15(g), a distinct resolved player Q gives picker-only; V15(h) rejects monster-owner exception and shares |
| R09a unusable party fallback | Added V15(i), absent usable party data gives picker-only |
| R09a missing recipient room | Added V15(j), missing-room B neither affects divisor nor receives a request |
| HARDEST R04a coordinates/single-member identity | Coordinate source corrected below; n=1 original-player discriminator V10(g) retained; no duplicated getter primitive added |
| HARDEST R05a scaling order | No change to established rule/order: numbered operations and V11/V18 give exact results; new V11(j) adds a distinct early-return control |
| HARDEST R09a recipient/remainder behavior | No change to established algorithm: V06/V15 preserve ordered requests and amounts; new V15(g-j) cover distinct branches |

### Coordinate dependency and fixture correction - 2026-10-15

Root review identified that the original regenerated coordinate dependency cited NRF-R08a/NRF-R11, which describe formats/movement rather than these getters. The original D2Game wrappers at 0x1100/0x1130 have byte-checked source reconstructions; their dynamic X/Y imports resolve through thunk 0xA454 -> D2Common #10867/0x34BB0 and 0xA3B8 -> #10750/0x34B80. Those target reconstructions zero-extend the 16-bit integer parts, as established by NRF-R10b. Null dynamic paths are intercepted by the Game wrappers and yield 0. Static types 2/4/5 directly read signed 32-bit X/Y. The complete 352-byte collector listing and original-image instructions at 0xDAED0 confirm the same victim dispatch, member wrapper calls, zero-X gates and wrapped unsigned distance comparison. This is a targeted consumer/source clarification; NRF owns extraction and no movement primitive is copied here.

The earlier V10(j) assigned dynamic player X=65636, outside that getter's unsigned 16-bit result domain. It is corrected to an explicitly isolated collection/pool fixture: static victim type 2 at (65636,100), player members at (100,100), and supplied victim XP/level/party values. Each dx=-65536 still gives low32(square)=0; n=2, total 1347 and supplied raw/amount 673 each are unchanged. This does not assert that a static victim reaches monster-death award routing. The earlier dynamic fixture and positive dx reading are retained in the private fixture audit. Main input widths, consumer/dependency references, case boundary and gap references are corrected consistently. World-scale correspondence for every static producer remains Q03/NRF-Q04.

Description now states the ordinary gold-sharing effect without technical evidence metadata. No existing arithmetic value, parent/family identity or evidence gate is promoted. Additions stay inside the existing 18 families; every changed fixture/input boundary is included in the fresh full-pass-2 handoff. No gameplay observation, implementation test/comparison, or Git action ran during this triage.

## Blind-implementer full pass 2 - 2026-10-15

The fresh reader confirmed the expected main-file hash and reported MISMATCH 0, GUESS 0, GAP 1, DISCRIM 0, GATED 11, COVERAGE 8 and HARDEST 3. No executable case total is adopted. The one GAP, every rendered advisory entry and each HARDEST item is disposed below. The rule/source clarification and additional cases require fresh full pass 3; no reader approval or gameplay confirmation is inferred.

| Finding | Outcome |
|---|---|
| GAP: R01a victim types beyond player/monster | Original complete death entry rechecked as below; R01a now rejects absent/other-type victims at that entry and distinguishes the more permissive isolated wrapper/router boundaries. V07(i)/(j) distinguish those boundaries |
| COVERAGE R01a player mode gates | V07(f-h) add player acceptance with killable=false and independent player modes 0/17 rejection |
| COVERAGE R02a/R02c eligible killer without list / active-mask boundaries | No change: V08(g) and PVP-V24(f) already distinguish an eligible killer's absent primary match from selecting its base; V08(h) plus PVP-V24's independent controls/(d) cover nonintersection, absent membership, stateCount 0, nonextended base and second-word intersection. PVP-R10b owns that helper; no primitive is duplicated |
| COVERAGE R04a missing dynamic paths / zero distance | V10(p)/(q) add missing member/victim paths producing zero X; V10(m) contrasts absent victim, which skips the test. No change to zero-distance acceptance: baseline P and V10(l)'s admitted monster are at the victim; zero distance itself is not rejection. Existing V10(h) is the direct zero-X guard |
| COVERAGE R05a maximum/bonus extremes | V11(a) already exercises the input cap and V18 exceeds that cap after bonus, rejecting a second cap. Added V11(l)/(m) for narrow bonus-product wrap and negative post-bonus result; no extra maximum fixture |
| COVERAGE R06b maximum scan stop | Added V12(g), experience beyond all supplied thresholds stops at maximum 3 without requiring a later row |
| COVERAGE R08a each threshold failure | No change: V05/V14(a-b) fail the first next-level threshold with no rise; V14(g) passes two thresholds and fails T(23), returning final level 22. Repeating that same loop recipe at another index adds no distinct behavior |
| COVERAGE R09a negative amount / ordinary capacity overflow | Added isolated allocation V15(k) with negative supplied amount. Existing V15(a) already saturates recipient and picker capacities, returning/recreating request 851; no duplicate overflow fixture |
| COVERAGE R09b negative effective level | Added V15(l), actual supplied existing-array getter inputs with minimum adjustment disabled give cap -10000 |
| HARDEST R04a/R04b collection, wrapping and floating gate | No change to established arithmetic/gates; V10(j)'s isolated static route and V10(p-q)'s missing-path cases make their source boundaries explicit |
| HARDEST R05a scaling order | No change to numbered order; V11(l-m) add only distinct bonus outputs, leaving cap/ratio/helper dependency gates intact |
| HARDEST R09a ordered allocation | No change to ordered recipient/picker algorithm; existing remainder/cap/liveness logs and the isolated negative case carry exact results |

All rendered GATED limits retain no-change dispositions: R01b/Q02, R02b/Q05, R04b/Q07, R04c/Q03, R05c/Q01, R08b/Q07, R09c/Q04, R09d/Q07, R10/Q06 and R12a/R12b/Q06 remain explicit. Supplied consumer arithmetic, request/store distinctions and all other documented integration/producer/observation limits remain unchanged.

### Death-entry type clarification - 2026-10-15

The original Game death function at 0xDEEE0 has a complete 416-byte instruction listing with no unresolved encoded runs. Its entry normalizes absent victim to non-player/non-monster type 6, selects player type 0 or monster type 1, and otherwise branches from 0xDEEFE to return at 0xDF05F. Only player/monster mode/killable checks can reach wrapper call 0xDEF81. A complete original-image decoding corroborates these branches and the request location. No checked source reconstruction is available for this entry; the complete listing was read. This adds the missing R01a rejection boundary rather than guessing a default for other kinds.

The complete 48-byte wrapper at 0x777F0 checks only killer/victim presence and skipKillExperience before forwarding to 0xDEA00. The complete router requires a player/monster killer and own-array victim experience >0; it does not repeat the death entry's victim-type/mode eligibility. V07(i) rejects types 2/3/4/5/absence at the death entry, while V07(j) permits a supplied type-2 wrapper call only as a router-entry request. Neither that case nor V10(j)'s static-victim collection fixture asserts actual death eligibility for a static victim. Unread non-experience death/scoring/notification interiors are not promoted by this local branch check.

### Additive forwarding locator correction - 2026-10-15

Root review identified that the reconstructed additive wrapper's symbolic callee name pointed to 0x39330. An independent original-image check of D2Common #10551/0x3A710 confirms the tail branch at 0x3A727: its next instruction address 0x3A72C plus signed displacement -0x3FC targets 0x3A330. The function table now uses that original target. The earlier symbolic name is preserved as the discrepancy, not treated as original target identity. Q07 still gates actual mutation/propagation; this provenance correction establishes no target-interior behavior.

ItemTypes row 4 / ItemType Gold / Code gold is explicitly distinguished from recreation's item code `gld `. Known prose separator encoding errors were corrected individually; no question marks, technical identifiers or raw data cells were broadly rewritten. The 13 parent rules, 18 vector families, 5 observation families and 8 question identities remain. No implementation test/comparison, gameplay operation or Git action ran in this triage.

### Routed own-array input clarification - 2026-10-15

The earlier regenerated input table used generic stat-read wording for victim XP/level and member level, leaving source selection ambiguous despite the caller instructions. Complete Game router 0xDEA00 instructions and its 416-byte original-image decoding show victim XP at call 0xDEA34, victim level at 0xDEA5C, and credited-player level at 0xDEB43 all call thunk 0xA3E2, D2Common #10587/0x38C20. Hireling initial level uses that same import at 0xDEB0A. These are own-array reads. Party collection's member level instead calls effective getter thunk 0xA42A at 0xDB005; bonus reads likewise use the effective getter. The exact imports and existing-array behavior consume NRF/PVP's compatible records; no aggregation or mutation proof is added.

Section 3 and R01a now explicitly identify the own-array router inputs. V07(k) gives immutable distinct own/resolved XP and levels on victim and credited player, with resolved-stat selection enabled and minimum adjustment disabled. Correct routed inputs are XP 200, victim level 10 and credited level 10, yielding a solo amount/update request 200. A wrong effective read changes the supplied scaling inputs/result. It is a caller-source discriminator rather than another implementation of the getter. Supplied default ratio 0 is explicit, with shift 0 retaining each earlier default result. No actual experience storage, level-up or hireling effect is promoted. This input/source clarification and its fixture belong in fresh full pass 3.

### Writer checks for full pass 3 - 2026-10-15

The spec checker reports 0 errors and 0 warnings for the current spec, companion and gap file. The writer recomputed the new player/type gates, routed own/effective discriminator, negative/wrapped bonus, bounded level scan and signed gold cases from the stated rules. The original documents and their hashes remain verified; the earlier invalid dynamic coordinate fixture is preserved in a separate private fixture-history record. All known prose separator errors were individually repaired. Fresh reader pass 3 remains pending; no case count, gameplay match or approval is inferred.

## Blind-implementer full pass 3 - 2026-10-15

The fresh reader reported MISMATCH 0, GUESS 0, GAP 0 and DISCRIM 0, with GATED 31, COVERAGE 8 and HARDEST 3. Its reported gates include the Q01-Q08 limits, unresolved leaves, supplied-stage vector limits and pending O01-O05 observations; all retain their existing boundaries and labels. This record adopts no verified case total and creates no substitute gate count. Two advisory branches receive new result-distinct cases, requiring fresh full pass 4 before review completion.

| Finding | Outcome |
|---|---|
| COVERAGE R01a absent killer/victim in isolated wrapper | No change: R01a states both presence guards; consumed PVP-R10a/PVP-V27(b) independently tests absent killer and victim at that wrapper. V07(i) separately tests death-entry absence. No duplicate wrapper matrix |
| COVERAGE R02c nonpositive/greater-than-32 credit scan | No change: consumed PVP-R10b/PVP-V24's independent stateCount=0 control and case (d), count 33 with a second-word match, discriminate the helper's result. PXL does not duplicate that primitive |
| COVERAGE R04a no-party id/unusable data fallback | No change: R04a explicitly selects the same single-credited-player callback recipe for roomless, unusable party data and id 65535. V10(f) exercises that callback/result recipe; V09 separately demonstrates actual solo id-65535 routing |
| COVERAGE R04a static victim types 4/5 | No change: the source-checked wrapper dispatch groups types 2/4/5 into the identical supplied i32 X/Y read. V10(j) discriminates that static route at type 2; another type adds no operation/result recipe |
| COVERAGE R05a monster missing owner/record | No change: R05a step 6 explicitly returns 0 on either missing requirement, the same final-hook zero-result recipe exercised by V11(k). Actual owner/record production remains Q07/Q08; V14(i) separately demonstrates the update stage's missing-record rejection |
| COVERAGE R05a shift outside 0..31 | No change: R05a step 4 states the retention guard; the shared shift=0 fixture default already exercises retention. Negative/at-least-32 shifts reach that same unchanged-x recipe, without a new calculation |
| COVERAGE R08a progression reaches maximum | Added V14(k): new XP exceeds both T(98) and T(99), but max=99 allows progression only through 98. This distinguishes maximum termination from next-threshold failure |
| COVERAGE R09a negative amount through shared party | Added V15(m): signed -1000/3 truncates to -333, producing recipient requests -333/-333 and picker -334. This differs from picker-only and floor division |
| HARDEST collection/wrapped pool/floating gate | No change: the rule and V10(j) separate raw wrapped collection from supplied floating production; Q07 remains explicit |
| HARDEST scaling order | No change: numbered R05a and existing cap/ratio/bonus/own-source fixtures retain exact order; no new helper primitive |
| HARDEST ordered gold allocation | No change to algorithm: V15(m) adds a distinct signed allocation case with exact ordered requests; existing ordinary cap/excess/liveness cases remain |

No source reading or gameplay operation was added in this triage. The new hireling case uses the established supplied-record threshold multiplication and caller guards; the negative party case uses the established signed split/subtraction and add guards. Every output is an arithmetic result or request, with actual mutation/progression/recreation still gated. The three current documents and original historical gap findings remain within their existing boundary. Writer self-review recomputed T(97)=92208200, T(98)=95079600, T(99)=98010000, hook cap 44865 and signed allocation -333/-333/-334. Document lint reports 0 errors and 0 warnings. Fresh full pass 4 remains pending; status and approval are unchanged.

## Blind-implementer full pass 4 - 2026-10-15

The fresh reader recomputed 104 explicitly named fresh executions, counting stated baselines and independent runs and excluding observation vectors. It reported MISMATCH 0, GUESS 0, GAP 2, DISCRIM 0, GATED 23 (corrected from an initial 21), COVERAGE 8 and HARDEST 3. The 23 rule-level gate entries retain their existing questions and exact boundaries. The two output-inventory omissions are now corrected without changing any calculation, supplied input or vector. A fresh final pass checks the resulting inventory and contract.

| Finding | Outcome |
|---|---|
| GAP V07 routing-entry request | Edit: section 5 now names the already specified award-routing entry/forwarding request and R01a. No caller behavior, input or vector changes. |
| GAP V16 NoDrop result | Edit: section 5 explicitly names the dependency-owned pre-normalization count used by R11/MSD-R21 as a comparison output. MSD remains the owner; no treasure-roll or producer assertion is added. |
| COVERAGE both killer/victim eligible | No change: V08(d) distinguishes victim selection erasing an earlier killer list; V08(c) exercises a present victim selection. Consumed PVP-V24(c) supplies both eligibility inputs and a victim redirect lookup failure after a killer match, rejecting retention of the earlier match. |
| COVERAGE multiple primary matches | No change: consumed PVP-V24 uses primary [nonredirect, redirect P3, redirect P4]; its ineligible-victim case returns the first matching P3. PVP-R10b owns the helper and discriminator. |
| COVERAGE direct absent member | No change: R04a expressly rejects absence before collection. V10(l) already exercises the same no-admission/no-level-read callback rejection recipe; V10(n) also pins registry absence before invocation. No absent member is manufactured by the iterator. |
| COVERAGE wrapping sumL | No change: R04a specifies low32 i32 accumulation and the signed nonpositive guard. V10(o) exercises count 2 with a nonpositive sum and no awards; V10(j) separately discriminates wrapped integer evaluation. Another sum that wraps to a nonpositive value uses the same no-award recipe; unsafe later scaling remains Q07. |
| COVERAGE out-of-range ratio shift | No change: R05a step 4 explicitly requires 0<shift<32; the shared shift-0 defaults already exercise unchanged-x retention. Negative or at-least-32 shifts select the same recipe, as recorded in FULL3 triage. |
| COVERAGE widened bonus product | No change: this call fixes divisor 100. NRF-R05 widened branches require 100>(large positive operand>>4), but qualifying a-large>65536 gives a shifted value at least 4096, and b-large>1048576 gives at least 65536. Thus the widened branch cannot be selected by this bonus caller. NRF-V04(e) owns that generic helper branch; V18 pins the actual caller argument order and divide-first path. |
| COVERAGE player add guard above cap | No change: with the same supplied gold/cap values for capacity and add, proposed>cap replaces accepted with low32(cap-gold), so the next wrapped total equals cap; otherwise it remains proposed<=cap. The above-cap add guard is not reachable through this capacity composition without changing the supplied read boundary. V15(f) already distinguishes the reachable negative-total zero replacement. Actual intervening writer/read production remains Q07. |
| COVERAGE unavailable level definition | No change: R09b explicitly returns 0 for this missing prerequisite, the same guarded-zero cap recipe exercised by V15(e) for missing unit/list. The getter definition guards are consumed from NRF-R12b/R12c, rather than duplicated. |
| HARDEST collection/pool/floating boundary | No change: ordered R04a, V10 distance/empty-record/fatal/n=1 cases and R04b/Q07 explicitly separate verified consumers from floating production. |
| HARDEST scaling order | No change: numbered R05a and V11/V18 supply exact signed/unsigned/helper operand-order assertions; V07(k) discriminates routed own-array inputs. |
| HARDEST gold allocation | No change: R09a and V15 assert distinct counting/recipient/picker passes, capacity/remainder, liveness and signed allocation, with storage/recreation gates. |

No source reading, new gameplay, implementation test or implementation comparison was added. These output-inventory clarifications do not turn a dependency-owned count into a local producer or imply original stored effects. Status remains draft and maintainer approval is pending.

## Blind-implementer full pass 5 - 2026-10-15

The fresh full reader reported MISMATCH 0, GUESS 0, GAP 0, DISCRIM 2, GATED 11 unresolved rule IDs, COVERAGE 10 reported across nine topic groups and HARDEST 3. It also retained supplied-stage/integration/observation limits. No numeric case tally was reported. Every rendered entry has a disposition below.

| Finding | Outcome |
|---|---|
| DISCRIM monster mode 0 | Added V07(l), a fresh death-entry mode-0 monster with positive own XP and otherwise passing inputs; it returns before wrapper/routing. The resulting case requires a fresh delta review. |
| DISCRIM both killer and victim have a credit list | No change: consumed PVP-R10a/PVP-V24(c) supplies an earlier matching killer list and an eligible victim list whose distinct owner lookup fails. Correct credit is absent, whereas retaining the earlier killer list yields its player; this owning-contract case already rejects that alternative. PXL-V08(d) separately discriminates victim absence erasing earlier selection. |
| COVERAGE death/router guards | V07(l) adds monster mode 0. V07(i) already rejects absent victim at death entry; PVP-V27(b) owns both absent-unit wrapper cases. V07(d) supplies nonpositive XP and no award work; the same positive-X router guard applies to direct invocation. No second unrelated wrapper matrix. |
| COVERAGE multiple primary matches | No change: PVP-V24(b) consumes the ordered primary chain with two matching redirects and selects its first match. The helper remains owned by PVP-R10b. |
| COVERAGE static members/extreme level sums | No change: R04a rejects every member type outside player/monster before coordinate reading; those dynamic types cannot select a static-member branch. V10(l) pins the callback type guard. Low32 sum and signed nonpositive guard already use V10(o); arbitrary unsafe later scaling remains Q07, as recorded in FULL4. |
| COVERAGE monster hook/cap equality | No change: V14(e/f) exercise cap reduction and unsigned negative-x comparison, and V14(i) pins a missing update record. R05a explicitly rejects missing owner/record; the same zero-result recipe is already recorded. At x==cap, retaining x versus assigning cap produces identical output, so a different comparator at equality is behaviorally equivalent there. |
| COVERAGE player unsigned cap equality/overflow | No change: at new==cap both paths request identical new/delta, so > versus >= cannot be distinguished by an output. V12(d) exercises wrapped unsigned cap selection; generic low32 arithmetic belongs to NRF. No numeric saturation is invented. |
| COVERAGE hireling addition wrap/near maximum | No change: R08a explicitly uses low32(old+2*amount); V14(b) pins doubling, NRF owns wrap arithmetic, and V14(k) already starts at 97 and stops at max-1=98 despite the next threshold fitting. No second threshold or arithmetic primitive matrix. |
| COVERAGE gold fallback/death mode/equality | No change: failed record selects the same picker-only amount recipe as V15(i); V15(d) tests the picker death-mode refusal of recipient work. Mode 0 and 17 select that same declared guard. At capacity equality proposed==cap, accepting amount or replacing it with cap-gold produces the same amount/excess. Existing V15 cap/excess cases pin the distinct results. |
| COVERAGE missing definition | No change: R09b and the consumed NRF getter contract give the guarded-zero recipe already exercised for missing unit/list by V15(e), as recorded in FULL4. |
| COVERAGE further NoDrop inputs | No change: MSD-R21 owns the count and its cases; R11/V16 specify only the result-changing contrast with XP distance. Expanding the dependency matrix here would duplicate that owner. |
| HARDEST collection/scaling/allocation | No change: exact ordered predicates, values, helper arguments, unsigned caps, request logs and supplied effects remain explicit; existing gates continue to qualify complete integration. |

Only the new V07(l) scenario changes executable coverage after this full pass. The other discrepancy is resolved by the existing consumed discriminator, not by adding a duplicate selector. A fresh reader checks the resulting delta, the output-inventory clarifications and the cited owner case; unchanged arithmetic/cases retain the full-pass review above. No source reading, implementation test or gameplay observation was added.

## Fresh delta pass 6 and router-entry correction - 2026-10-15

The fresh delta reader checked the section-5 inventory rows, V07(l), V07(d) and shared inputs, and V08 against PVP-R10a/R10b/V24. Counts: MISMATCH 0, GUESS 1, GAP 0, DISCRIM 1, GATED 2, COVERAGE 0, HARDEST 3. It confirmed dependency-owned NoDrop inventory and the existing PVP selection discriminator. Its one guess/discrimination finding identified a real boundary ambiguity: the old V07 logger used “award-routing request” for both forwarding into the router and proceeding after its positive-experience guard.

The compatible original router image supplies the correction, without a new source campaign: entry 0xDEA00 checks killer type at 0xDEA20..0xDEA29, reads own XP at 0xDEA34, rejects nonpositive XP at 0xDEA3F, and only then requests credit selection at 0xDEA48 through 0xDB5E0. The wrapper forwards present, non-skipped inputs before those router guards. Thus X=0 enters the router but does not reach credit selection. R01a and section 5 now name these two operations separately. V07's default logger observes post-guard CreditSelection; case (d) additionally observes RouterEntry and forbids CreditSelection/awards. Case (j) logs only the isolated wrapper's RouterEntry; case (k) keeps its own scaling/update boundary. Mode-0 case (l) rejects before either operation. This corrects the earlier inventory terminology and ambiguous no-request assertion without inventing an experience amount or changing storage gates.

The two GATED entries receive no change: Q02 retains producer meanings and MSD-R21 owns the pre-normalization NoDrop count with no treasure roll. All three HARDEST entries receive no change beyond the logger correction: V07 now distinguishes the death/wrapper/router stages; R11/V16 retain separate XP and NoDrop counts; PVP-V24 owns list replacement/erasure and fallback. The earlier FULL4/FULL5 no-request wording for X=0 is superseded by this explicit distinction. A fresh reader checks the resulting R01a/section-5/V07 delta; other arithmetic and cases retain their applicable full reviews. No gameplay or implementation tests ran.

## Fresh delta pass 7 and final wording - 2026-10-15

The fresh reader checked R01a, section 5 and all V07 cases with their shared logger definitions and consumed PVP contracts. Counts: MISMATCH 0, GUESS 1, GAP 0, DISCRIM 0, GATED 9 grouped entries, COVERAGE 1, HARDEST 3. The remaining guess was the generic sentence “A rejected stage makes no experience request,” which could be read as erasing an earlier RouterEntry even though V07(d) now explicitly retains it. It now says rejection stops later operations while earlier reached requests remain observable. No guarded calculation or control flow changes.

The advisory positive-X/non-player-victim router branch now has V07(m): it reuses V07(k)'s immutable inputs with victim type 2, begins inside the router and asserts CreditSelection, amount 200 and the level-update request, without a death-eligibility claim. The existing complete router instructions at 0xDEA20..0xDEA48 establish killer-type/XP guards before selection and no victim-type guard there; this case uses that already recorded supplied-input boundary, not a new producer or source campaign.

All nine documented gate groups receive no change. All three HARDEST entries receive no change beyond the wording/fixture: V07 separates entry from post-guard work, PVP-V24 owns selection/erasure and PVP-V27 supplies ordered routing, without complete award/scoring effects. The next fresh delta reader checks the clarified rejection sentence and resulting V07 family. Other arithmetic and unchanged cases retain the five full reviews. Status, integration gates and user approval remain unchanged; no gameplay or implementation test ran.

## Fresh delta pass 8 and lookup inventory - 2026-10-15

The next fresh reader checked the rejection/entry/selection wording, section 5 and the entire V07 family, including case (m), against the cited PVP contracts. Counts: MISMATCH 0, GUESS 0, GAP 1, DISCRIM 0, GATED 1 grouped producer/entry limit, COVERAGE 2, HARDEST 3. It independently confirmed all V07 outputs and the type-2 router amount 200. The one remaining gap was section 5 omitting the active-hireling lookup attempt already specified by R01a and consumed by R08a/PVP-R11a. Section 5 now names the conditional lookup attempt and supplied result, preserving R08b/Q07's actual producer gate. No behavior, input or vector changed for this inventory correction.

Both coverage entries receive no change: consumed PVP-V27(b) exercises absent killer/victim at the wrapper, and PVP-V27's count-8 supplied routing case asserts the conditional hireling lookup and result before later updates/party routing. V07(k)/(m)'s count-7 fixtures intentionally skip that branch; PXL-V14 separately exercises lookup-gate and amount/update boundaries. The reviewed producer limits Q02/Q05/Q07/Q08/PVP-Q13 remain explicit. All three HARDEST entries receive no change: V07 now names nested entry/guard boundaries, PVP-V24 owns selection-versus-failed-lookup behavior, and V07(m) isolates positive-X type-2 router behavior without a death eligibility claim.

A fresh final delta reader checks only this last output-inventory addition against the unchanged rules, inputs and cited routing examples. All other executable cases retain their applicable full/delta review. No source, gameplay or implementation tests were added.

## Fresh delta pass 9 and review completion - 2026-10-15

The fresh reader independently checked the sole final section-5 lookup inventory addition against R01a/R08a/R08b, section 3 and PVP-R11a/V27. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 1, COVERAGE 1, HARDEST 3. No changed arithmetic or vector was introduced after delta 8's clean V07 recomputation. The one gate remains actual active-hireling lookup/entry production under R08b/Q07/PVP-Q13; naming an attempt and supplied result does not establish that search.

The advisory absent-player-data lookup skip receives no change: R01a/PVP-R11a explicitly require both present player data and count >7. PVP-V27's count-8 supplied attempt and V07/V14's count-7 skip establish the two existing attempt/no-attempt recipes; another failed prerequisite produces the same no-attempt result, without a new amount or producer. All three HARDEST entries receive no change: R01a/V07 separates nested stages; R08a names supplied lookup results and subsequent guards; PVP-V27 pins request order while retaining search integration limits.

Five full independent readers covered the unchanged contract, followed by fresh delta checks 6 through 9 for each amended part. Every reported blocking finding has a recorded outcome, and the final applicable read has no unresolved mismatch, guess, gap or discrimination finding. Root review read all three documents, additions and dispositions, corroborated the missing-record branch/coordinate/forwarding evidence and hand-recomputed new cap, progression and signed allocation examples. Document review is complete. Status remains draft; exact integration, floating environment, producer and observation gaps and maintainer approval remain. No implementation test, gameplay or new implementation comparison ran. Publication repeats document checks on committed bytes.
