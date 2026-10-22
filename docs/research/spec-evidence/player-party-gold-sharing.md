# Party gold sharing evidence

This companion supports the [behavioral contract](../../specs/player-party-gold-sharing.md). Original artifacts resolve through the [source manifest](../source-manifest.md). Function descriptions are analyst labels; an original export ordinal identifies the exported entry but does not validate an analyst-assigned name. All addresses below are module-relative RVAs unless explicitly stated as VAs. Private source forms, fixture encodings, complete original documents and collection details are preserved separately.

The records below were made while gold-pickup sharing was part of the combined party experience contract and were divided on 2026-10-15 (see the restructure record). The [party experience companion](player-party-experience-sharing.md) keeps the full combined records, including each reader pass's counts.

## Identity and evidence reuse

The historical static readings date to 2026-10-11 and 2026-10-12. They covered the split caller among the award and split callers of the combined contract. The 2026-10-12 independent audit corrected gold excess routing among other combined-contract items. Compatible caller arithmetic and request-site evidence is reused with those dates; a new format does not renew a reading or make unread callees verified.

The current original images match manifest artifacts `LOD113C-D2GAME` (SHA-256 `24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53`, 1138688 bytes) and `LOD113C-D2COMMON` (SHA-256 `59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10`, 679936 bytes). Complete narrow original-image checks on 2026-10-15 use those identities. This establishes analysis-artifact correspondence, not historical loaded-DLL or data-profile equivalence.

Compatible records reused: RLI-R02a/RLI-R02b existing-array reads dated 2026-10-14, GDF field/name/table records dated 2026-10-14/15, MSD-R21 iterator/alive-count record dated 2026-10-14 and the PXL-R04a iterator branch check dated 2026-10-15. [NRF evidence](core-numeric-and-rng-foundations.md), [GDF evidence](core-game-data-fields.md), [MSD evidence](monster-scaling-by-players-and-difficulty.md) and [PXL evidence](player-party-experience-sharing.md) retain their own methods and limits. Getter coverage supplies existing values, never additive/replacing writes. [STL-R06c](state-timed-state-lifecycle.md) and RLI-Q03 preserve actual propagation gaps.

## Original function references

| Artifact | Function RVA / original location | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | 0xD6E30; 0xD6C80; 0x9C270 | none | Pickup dispatch; gold caller; party split | GPS-R01/GPS-R06 | Historical dispatch; complete caller/split listings and image correspondence 2026-10-15; amount/request boundary |
| LOD113C-D2GAME | 0xDFA60; 0x9BEB0 | none | Player party id | GPS-R01 | Historical consumers; MSD-R21 dated 2026-10-14 |
| LOD113C-D2GAME | 0x9C070; 0x9BD00; 0x9C000; 0xDFA90 | none | Same-level iterator, party record, alive callback/test | GPS-R01 | MSD-R21 2026-10-14; complete 240-byte iterator original-image check 2026-10-15 corrects null/empty-record branch (PXL-R04a) |
| LOD113C-D2GAME | 0x9BC70; 0x1670 | none | Capacity outputs; bounded add request | GPS-R01/GPS-R04 | Complete capacity 64-byte image and both instruction listings 2026-10-15; actual propagation not established |
| LOD113C-D2GAME | 0x12B90 | none | Excess-gold creation lead | GPS-R04 | Historical selected item-code/chunk/room observations; placement/creation/interior draws unread |
| LOD113C-D2GAME | 0xD2F00 | none | Item validation lead | GPS-R05 | Historical partial path; no universal policy absence |
| LOD113C-D2COMMON | 0x31940 | 10049 | Gold cap | GPS-R02 | Complete 96-byte listing and image 2026-10-15, effective getter consumes RLI-R02a/RLI-R02b |
| LOD113C-D2COMMON | 0x38C20; 0x38B70; 0x38A80 | 10587; 10973; none | Own/effective/existing-array stat reads | GPS-R01/GPS-R02 | RLI-R02a/RLI-R02b; own/effective wrapper distinction reuses PVP 2026-10-15 |
| LOD113C-D2COMMON | 0x3A740; 0x3A710; 0x3A330; 0x3A280 | 10887; 10551; none; 10261 | Replacing/add forwarding and propagation | GPS-R01/GPS-R04 | Requests resolved 2026-10-15; additive wrapper checked reconstruction read, downstream propagation/list integration gated GPS-Q02/STL/NRF |

## Dispatch, imports and semantic mapping

Pickup 0xD6E30 → 0xD6C80 → 0x9C270 selects count callback 0x9C000, capacity 0x9BC70 and add 0x1670; positive returned excess reaches 0x12B90. Complete pickup deletion/notifications outside the amount handoff are not promoted. The count uses the same-level iterator 0x9C070 recorded under PXL-R04a.

PE imports/export correspondence was resolved on 2026-10-15 against the identified original images:

| D2Game thunk | D2Common export / target RVA | Semantic operation |
|---|---|---|
| 0xA45A | #10049 / 0x31940 | Gold cap |
| 0xA3E2; 0xA42A | #10587 / 0x38C20; #10973 / 0x38B70 | Own versus effective stat reads |
| 0xA472; 0xA4D2 | #10887 / 0x3A740; #10551 / 0x3A710 | Replacing versus additive requests |

Original layout evidence maps neutral terms as follows: Unit +0xC6 bit 0 is aliveExclusionCondition; Unit +0xC8 bit 10 with +0x94/+0x98 lookup supplies ownerExceptionCondition. These are evidence encodings, not required engine layouts.

## Corrections and affected claims — 2026-10-15

These are the gold parts of the combined contract's correction list; the full list is in the [party experience companion](player-party-experience-sharing.md#corrections-and-affected-claims--2026-10-15).

1. **Getter versus write (GPS-R01).** Prior wording stated mutated stats while local caller evidence established requests. Gold reads effective values. New caller instructions and imports distinguish those. Actual #10887/#10551 effects and recreation remain integration gates; vectors assert requests. No getter proof is reused as mutation proof.
2. **Gold cap (GPS-R02).** The new complete cap reading supports GPS-R02 at the supplied-getter boundary; that label became verified static for those assertions only. GPS-V02 adds the corresponding discriminators.
3. **Search/RNG scope (GPS-R05, GPS-R06).** The drop ownership/timer rule had an unretained complete six-step absence search. It is unresolved, rather than newly claimed absent. No-direct-draw is restricted to local supplied-input calculations/requests; unread update/creation interiors are excluded.

### Additive forwarding locator correction - 2026-10-15

Root review identified that the reconstructed additive wrapper's symbolic callee name pointed to 0x39330. An independent original-image check of D2Common #10551/0x3A710 confirms the tail branch at 0x3A727: its next instruction address 0x3A72C plus signed displacement -0x3FC targets 0x3A330. The function table now uses that original target. The earlier symbolic name is preserved as the discrepancy, not treated as original target identity. GPS-Q02 still gates actual mutation/propagation; this provenance correction establishes no target-interior behavior.

ItemTypes row 4 / ItemType Gold / Code gold is explicitly distinguished from recreation's item code `gld `.

## Data records and profile limits

GDF's 2026-10-15 complete effective named-stat/mode/type records corroborate identities within its reported vanilla LoD 1.13c profile, including stat 14 `gold`, stat 12 `level`, the player death/dead modes and ItemTypes row 4 (GDF-R07). The version catalog refresh failed; a complete current override inventory, loaded DLL identities and physical supplying archives were not established. This contract consumes only the supported names/identities, retaining their dates and limits. No gold table cell is required by the supplied fixtures.

## Gap assessment — 2026-10-15

These are the gold parts of the combined assessment recorded in the party experience companion.

- Q01 (owner condition): checked current GDF/PVP identity/writer limits. The known test remains a supplied-input contract; the wider meaning/producer is not guessed.
- Q02 (storage/recreation): newly read complete gold capacity/add and cap. Actual propagation remains under STL/NRF; additive forwarding still reaches unresolved list mutation. The recreation request location identifies the smallest remaining interior. These are explicit gates, not failures silently accepted as proof.
- Q03 (drop policy): retained the historical pickup-validation candidate. Current contracts establish bounded upstream roles, not complete drop/timer policies. The unretained search record cannot support absence. Broad drop campaigns are separate work.
- Observation: no matching original-game split observation was made or reused as a match.

## Observation procedures

The procedure is pending. Use owned matching Expansion 1.13c artifacts/data, a recorded character class/level/stats, known party/room inputs, pre/post raw values and attributable reached-stage records. Record capture completeness and measured assertion scope. An archive read is not gameplay confirmation.

- **GPS-O01:** use three same-level party players, measured level/gold/caps, then pick known gold amount. Record divisor/recipient lists, cap/accepted/excess at 0xD6C80/0x9C270/0x9BC70 and actual stat storage. Repeat far/dead/other-level/cap controls. At excess 851 verify recreation entry 0x12B90 separately from placement/item count/seed changes; no creation outcome is assumed.

## Blind-implementer review records

Five full independent reader passes and delta passes 6 to 9 ran on the combined contract on 2026-10-15. Their counts, gate dispositions and non-gold findings are in the [party experience companion](player-party-experience-sharing.md#blind-implementer-full-pass-1---2026-10-15); delta passes 6 to 9 reported no gold finding. The gold findings and their outcomes are below, by pass.

### Full pass 1 - 2026-10-15

The gold rules' documented gates received no change: GPS-R01's effects are GPS-Q02; GPS-R03/GPS-R04 meanings/storage/recreation are GPS-Q01/GPS-Q02; GPS-R05 drop policies are GPS-Q03.

| Rendered finding | Outcome |
|---|---|
| GPS-R01 owner exception success | Added GPS-V02(g), a distinct resolved player Q gives picker-only; GPS-V02(h) rejects monster-owner exception and shares |
| GPS-R01 unusable party fallback | Added GPS-V02(i), absent usable party data gives picker-only |
| GPS-R01 missing recipient room | Added GPS-V02(j), missing-room B neither affects divisor nor receives a request |
| HARDEST GPS-R01 recipient/remainder behavior | No change to established algorithm: GPS-V01/GPS-V02 preserve ordered requests and amounts; new GPS-V02(g-j) cover distinct branches |

### Full pass 2 - 2026-10-15

| Finding | Outcome |
|---|---|
| COVERAGE GPS-R01 negative amount / ordinary capacity overflow | Added isolated allocation GPS-V02(k) with negative supplied amount. Existing GPS-V02(a) already saturates recipient and picker capacities, returning/recreating request 851; no duplicate overflow fixture |
| COVERAGE GPS-R02 negative effective level | Added GPS-V02(l), actual supplied existing-array getter inputs with minimum adjustment disabled give cap -10000 |
| HARDEST GPS-R01 ordered allocation | No change to ordered recipient/picker algorithm; existing remainder/cap/liveness logs and the isolated negative case carry exact results |

GATED limits GPS-R03/GPS-Q01, GPS-R04/GPS-Q02 and GPS-R05/GPS-Q03 retained no-change dispositions.

### Full pass 3 - 2026-10-15

| Finding | Outcome |
|---|---|
| COVERAGE GPS-R01 negative amount through shared party | Added GPS-V02(m): signed -1000/3 truncates to -333, producing recipient requests -333/-333 and picker -334. This differs from picker-only and floor division |
| HARDEST ordered gold allocation | No change to algorithm: GPS-V02(m) adds a distinct signed allocation case with exact ordered requests; existing ordinary cap/excess/liveness cases remain |

The negative party case uses the established signed split/subtraction and add guards. Writer self-review recomputed signed allocation -333/-333/-334.

### Full pass 4 - 2026-10-15

| Finding | Outcome |
|---|---|
| COVERAGE player add guard above cap | No change: with the same supplied gold/cap values for capacity and add, proposed>cap replaces accepted with low32(cap-gold), so the next wrapped total equals cap; otherwise it remains proposed<=cap. The above-cap add guard is not reachable through this capacity composition without changing the supplied read boundary. GPS-V02(f) already distinguishes the reachable negative-total zero replacement. Actual intervening writer/read production remains GPS-Q02. |
| COVERAGE unavailable level definition | No change: GPS-R02 explicitly returns 0 for this missing prerequisite, the same guarded-zero cap recipe exercised by GPS-V02(e) for missing unit/list. The getter definition guards are consumed from RLI-R02a/RLI-R02b, rather than duplicated. |
| HARDEST gold allocation | No change: GPS-R01 and GPS-V02 assert distinct counting/recipient/picker passes, capacity/remainder, liveness and signed allocation, with storage/recreation gates. |

### Full pass 5 - 2026-10-15

| Finding | Outcome |
|---|---|
| COVERAGE gold fallback/death mode/equality | No change: failed record selects the same picker-only amount recipe as GPS-V02(i); GPS-V02(d) tests the picker death-mode refusal of recipient work. Mode 0 and 17 select that same declared guard. At capacity equality proposed==cap, accepting amount or replacing it with cap-gold produces the same amount/excess. Existing GPS-V02 cap/excess cases pin the distinct results. |
| COVERAGE missing definition | No change: GPS-R02 and the consumed NRF getter contract give the guarded-zero recipe already exercised for missing unit/list by GPS-V02(e), as recorded in FULL4. |

## Specification review - 2026-10-15

The independent checklist review of the combined contract re-read the gold cap D2Common `0x31940` from original-image disassembly (`level * 10000`, zero when the stat-definition count is 12 or less). It agreed with GPS-R02. The full review record is in the party experience companion.

## Restructure (2026-10-15)

The review recommendation to split oversized specs was approved by the user. This spec, companion and [gap file](../../plans/gaps/player-party-gold-sharing.md) were created from the gold-pickup part of the combined party experience and loot-sharing contract, now [party experience sharing](../../specs/player-party-experience-sharing.md). The gold split/capacity rule, gold cap, owner-condition meaning, gold storage and excess recreation, the drop ownership/timer question, the gold half of the direct-draw boundary, the two gold vectors, the gold half of the local-seed vector (now GPS-V03, keeping the same seed objects and assertions) and the gold observation moved here. The owner-condition producer, gold storage/recreation integration and drop-policy parts of the combined questions became GPS-Q01, GPS-Q02 and GPS-Q03. The gold function rows, pickup dispatch, imports, layout mapping, observation procedure and gold reader findings were copied or moved here as recorded above.

No rule behavior, label, value, gate or vector result changed. The same-level party iterator stays with the experience contract as PXL-R04a and is consumed by GPS-R01. With six rule ids this spec is below the usual size signal; it remains one bounded behavior because the gold split has its own trigger, inputs and outputs. No source reading, gameplay operation or reader run was performed for this restructure; a fresh blind-implementer pass on the separated contract has not been run.
