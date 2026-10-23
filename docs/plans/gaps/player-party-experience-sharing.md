# Gaps: party experience sharing

Spec: [behavioral contract](../../specs/player-party-experience-sharing.md)
Evidence: [companion](../../research/spec-evidence/player-party-experience-sharing.md)

## Open evidence gaps

Affects is a reference list. Only the named assertions/boundaries are blocked; supplied-input arithmetic and ordered requests remain available for partial acceptance.

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|---|---|---|---|---|---|---|
| Experience correspondence | integration | PXL-Q01; PXL-R05c; PXL-R05b/PXL-R06b integration | Original text/prepared record equality; not supplied-record accessors | Named Level/header/ratio correspondence | 2026-10-15 checked GDF full dated text/compiled coverage and original loader/accessors; old partial maximum cell has no full hash; no named correspondence promoted | Complete named text/prepared comparison and loaded profile identity |
| Gate purpose | outside-boundary | PXL-Q02; PXL-R01b | Original producer meanings only | Skip/count gate writers and purpose | 2026-10-15 reused GDF PetType identity/PVP gate consumers; count identity does not explain >7 | Trace actual producers and qualifying contexts |
| Distance scale | outside-boundary | PXL-Q03; PXL-R04c | Yards/screens conversion only | Raw getter scale/world comparison | 2026-10-15 reused NRF-R10b/NRF-Q03; checked dynamic u16/static i32 caller routes and corrected V09(j) supplied fixture; raw unsigned 6400 predicate preserved | Attributable world-scale correspondence |
| Neutral flags | outside-boundary | PXL-Q04 | Collection-exclusion production | Producer chain | 2026-10-15 retained neutral tests; complete writer meaning not established | Trace writers |
| Upstream killer | outside-boundary | PXL-Q05; PXL-R02b | All-source attribution | Remaining source-specific killer producer | 2026-10-15 reused PVP/IEP/DPO/OWD bounded contacts; no universal resolution inference | Complete source-specific root/output path |
| Penalty/corpse | behavior | PXL-Q06; PXL-R10a/PXL-R10b | Complete death penalty/recovery | Root/output/callee proof | 2026-10-11/12 candidate difficulty/leech/creation searches retained; 2026-10-15 current dependency assessment finds no complete six-step absence evidence | Complete bounded original consumer/producer reading and exact vectors |
| Award integration | integration | PXL-Q07; PXL-R06a/PXL-R08a | Actual stat writes, level/resources and hireling/notifications | Setter/propagation/update interiors | 2026-10-15 exact imports, own/effective reads and complete caller/threshold readings; RLI-Q03/STL-R06c retain actual mutation gate | Complete per-effect contracts and original-profile correspondence |
| Floating production/extremes | behavior | PXL-Q07; PXL-R04b/PXL-R05a | Original floating environment/alternate divide/conversion and unsafe extreme factor indices | Complete selected floating branch and environment; extreme-index behavior | 2026-10-15 complete pool/conversion instruction and image checks expose alternate divide and absent environment proof; supplied finite examples retained | Verify selected branch, precision/rounding and conversion/exception behavior; bound extreme indices |
| Active hireling production | integration | PXL-Q07; PXL-R08b/PXL-R05d | Actual entry/record lookup and preparation | Complete active-entry and prepared-record producer | 2026-10-15 record selector and threshold multiplication read; existing PVP-Q09 lookup limit retained | Producer/table identity and complete lookup contract |
| Ownership producers | outside-boundary | PXL-Q08; PXL-R02a/PXL-R02c | Actual registered owner/mask/list origin | Registration and redirect producers | 2026-10-15 reused PVP complete active-state/primary-list selector; no extra owner hop or producer claim | Complete actual producer path |
| Gameplay confirmation | observation | PXL-O01/PXL-O02/PXL-O03/PXL-O04; all verified rules | Runtime confirmation, not otherwise verified local assertions | Matching attributable original-game vectors | 2026-10-15 no gameplay action performed; every executable vector remains unobserved; original procedures retained | Reproduce each rule's attributable case with exact inputs/outputs/profile and record scope |

## Implementation gaps

The following historical findings are retained verbatim. They were not rechecked against an implementation during this migration; their old section numbers are historical records. The gold-split row G6 moved to the [gold gap file](player-party-gold-sharing.md) in the 2026-10-15 restructure; the other rows are unchanged.

<!-- Gap list: original 1.13c behavior (docs/specs/player-party-experience-sharing.md) vs current Dark Magic code. Input for a future implementation plan; not a plan. -->
# Gaps: player-party-experience-sharing (removed repo-comparison items)

Old section 2 "What Dark Magic already has": `policy/experience_gain.lua` (per-killer formula, from 1.10f D2MOO), `systems/monster_death.lua` (credits one killer, resolved through ownership; no party pool, no hireling share, no distance), `data/progression.lua` (Experience.txt: MaxLvl row supplies cap and shift, other rows supply ExpRatio keyed by the Level column), `policy/party.lua` (same-level living members), `docs/research/MULTIPLAYER_PARTIES_PVP_AND_TRADE.md` (XP and gold sharing marked as needing probes; quotes an official guide: about two screens, a 35% pool increase, shares proportional to level). Gold sharing: not implemented. All values are "implemented assumptions".

| # | Rule | Repo file/function or doc | Current repo behavior | Original behavior | Original status |
|---|------|---------------------------|----------------------|-------------------|-----------------|
| G1 | 3.4 / 3.3 party pool | `systems/monster_death.lua` | Credits one killer (resolved through ownership); no party pool, no distance | Splits among in-range party members using pool `X + trunc((n-1)*X*89/256)` weighted by level | verified (static) |
| G2 | 3.8 hireling share | `systems/monster_death.lua` | No hireling share | Hireling gets extra XP (86/256 if not killer; stat 13 += 2*x_m) in addition to player's award | verified (static); two fields asm-only |
| G3 | 3.4 pool size | `docs/research/MULTIPLAYER_PARTIES_PVP_AND_TRADE.md` | Quotes official guide: 35% pool increase | `(n-1)*89/256` (about 34.8% per extra member, repeats per member) | verified (static) |
| G4 | 3.5 non-positive input | `policy/experience_gain.lua` | Returns 0 for `x <= 0` | Returns 1 (reachable when a party raw share truncates to 0) | verified (static) |
| G5 | 3.4 proximity | research doc | "about two screens" | squared distance from the victim `<= 6400` (0x1900); units unresolved | verified (static); scale unresolved |
| G7 | 3.5 formula order/constants | `policy/experience_gain.lua` | Per-killer formula matches 1.13c order and constants | same | verified (static); no gap |
| G8 | Experience table indexing (old section 7, unresolved) | `data/progression.lua` | MaxLvl row supplies cap and shift; other rows keyed by the Level column | Accessors read row `level+1` (ordinals 10989, 10399, `LoadExperienceTxtTable` rva 0x60920); whether file row `level+1` is Level `level` needs the real file | gated |
| G9 | 3.2 killer credit | `policy/party.lua`, `monster_death.lua` | Same-level living members; ownership-resolved killer | Resolver per 3.2 incl. stat-list flag 0x800 owner override | verified (static) |
| G10 | verification tooling | `party_xp_probe` | Probe accepts only 1.14d; 1.13c capture needs re-targeting | n/a (process item) | n/a |

Old tracker/process content also removed (not gaps): "Issue(s): none assigned (completes the kill-award gap left by #373) · Parent #344", "Done for this spec" checklist (runtime vectors, 1.14d spot-check, implementing issue links spec).

## Restructure (2026-10-15)

Gold-pickup sharing moved to [party gold sharing](../../specs/player-party-gold-sharing.md). This file was renamed with its spec. The gold parts of the neutral-flag, drop-policy, award-integration and gameplay-confirmation rows moved to the [gold gap file](player-party-gold-sharing.md) as rows for GPS-Q01, GPS-Q03, GPS-Q02 and GPS-O01; the rows here keep their experience parts with the original attempt dates. Ids use the renumbered experience contract.

## Publication review - 2026-10-23

Fresh blind and evidence reviews approved the separated 1.13c supplied-input experience contract and corrected its coverage locator. No question or observation gap above was closed. The NoDrop comparison remains dependency-owned and does not establish a drop outcome, dead-player runtime behavior or 1.14d parity.
