# Gaps: combat-knockback

Spec: [Knockback](../../specs/combat-knockback.md)
Evidence: [public companion](../../research/spec-evidence/combat-knockback.md)

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|---|---|---|---|---|---|---|
| KNB-Q01 | integration | KNB-R01, KNB-R10c | Automatic compiled stat/selector binding; supplied callback remains executable | Matching compiled item stat record and loader correspondence | 2026-10-15 complete effective ItemStatCost text identifies item_knockback/ID 81 and selectors 7; loader descriptor reuse identifies fields, not compiled contents | Verify matching compiled record and code mapping |
| KNB-Q02 | integration | KNB-R10c, KNB-R10d | Item timing relative to block/crushing blow/deadly strike; no gate on supplied callback or conditional GH stage | Event-name numeric conversion and registration chain | 2026-10-15 text cells domeleedamage/domissiledamage checked against IEP-R09a/DPO-R13 numeric sites; numeric binding/compiled registration not established | Read name conversion, registration and numeric sites for this row |
| KNB-Q04 | observation | KNB-R07a, KNB-R08, KNB-O03 | Completed-displacement measurement only; supplied arithmetic remains verified | Attributable plan/actual-position capture | 2026-10-15 reused NRF-R10 scale and inspected generator widths; no gameplay action in migration | Matching KNB-O03 capture separating plan from later movement |
| KNB-Q06 | integration | KNB-R06d, KNB-R10a, KNB-V13, KNB-V21 | Complete outer mode-dispatch refusal outcome, entry preparation and full mode-change RNG total; local start refusal and ignored setter return remain verified | Complete outer dispatcher/refusal and dependent entry/state/elemental effects | 2026-10-15 full D2Game 0x75160 establishes calls/return; V13c consumes RLI-R04a's equal-mode setter return 0 while handler returns 1, with update effects still RLI-Q02; 0xDA750 entry processing is not proven to process timers; HRC-R15/HRC-R14 retain elemental/speed contracts | Finish outer caller return handling/dependent operations in their owners, preserving local supplied-stage assertions |
| KNB-Q07 | behavior | KNB-R09b, KNB-V13, KNB-V30, KNB-O04 | Actual collision-selected endpoint; supplied endpoint writes remain verified | Complete collision interior/byte-run decoding/callees | 2026-10-15 full D2Common 0x67390 establishes writes/count retention; partial 0x66C20 inspection does not establish clip algorithm | Complete original-image-checked collision path/callees |
| KNB-Q07 | integration | KNB-R06d, KNB-R09b, KNB-O03, KNB-O04 | Player displacement, count initialization, mode progression and later actual motion; supplied stages remain usable | Full player mode-19 preparation/planner/tick contracts | 2026-10-15 read D2Game 0x78760/0x78610 and D2Common 0x67390; player helper has type-8 setup, full jump/return/path-success/motion chain unverified | Resolve player/planner/tick chain and count initialization, then compare attributed capture |
| KNB-Q08 | outside-boundary | KNB-R03a, KNB-O06 | None locally; complete subtype producer/record/mode equivalence | Hireling/summon/NPC owner contracts | 2026-10-15 reaction reading/HRC-R02a confirms local kind routing, not uniform records/producers | Verify subtype producers/records in owning contracts |
| KNB-Q10 | outside-boundary | KNB-R09b | None locally; movement-conflict/network/client result | Movement conflicts and client correction contracts | 2026-10-15 reviewed earlier partial search; client not completely searched, no absence label retained | Movement/network-owner investigation and dependency contract |
| KNB-Q11 | integration | KNB-R05a, KNB-V12, KNB-O05 | Only disguise bookkeeping and update-queue effects; state-on routing/membership remains available through its owners | STL-R07a/STL-Q07 disguise completion and RLI-Q02 room-update-queue effects | Reused compatible RLI-R03/STL-R07 state operation records dated 2026-10-12/14; 2026-10-15 resolved #10945/RVA 0x34510 and read it plus 0x38730/0x665C0, establishing SetState(92,1), not ordinary death | Owning state/update contracts verify only their remaining integration assertions; local request fixture stays independent |
| KNB-Q12 | observation | All verified behavioral rules; KNB-O01–KNB-O06 | Runtime-confirmation claims only; independent evidence gates remain | Matching owned 1.13c gameplay captures | 2026-10-15 static/table work only; no matching prior gameplay observation; companion procedures written | Valid attributed observations naming actual cases/assertions for every in-scope rule |

## Implementation gaps

The following comparison is preserved from the earlier extraction. It is historical,
not a current implementation audit. Its wording/findings remain unchanged, including
claims superseded by current evidence. Current original-game conclusions are the spec
and companion. The historical rule subjects do not revalidate implementation findings.

Sources named in old spec: `docs/research/MOVEMENT_COLLISION_AND_PATHFINDING.md`, `internal/dev/tools/knockback_probe/` (schema `d2legacy.knockback_probe/v1`).

| # | Spec rule | Repo item (old wording) | Current repo behavior | Original (1.13c) behavior | Status of original |
|---|-----------|-------------------------|-----------------------|---------------------------|--------------------|
| G1 | KNB-R09a/KNB-R09b (displacement) | `world.forced_motion_request` (checkpointed; kind, source point, distance, speed, request tick) | Forced motion runs through the same collision, occupancy and bounds as locomotion; outcomes completed, partial, blocked, invalid, replaced. Combat emission DISABLED until rules pinned. | Single waypoint, endpoint clipped by D2Common collision trace (0x66C20); no speed/velocity constant; no occupancy handling found. | Displacement verified (asm); speed not found |
| G2 | KNB-R02a (missile chance) | `Missiles.KnockBack` raw byte preserved (examples blank, 1, 33, 75); old caution "do not treat as percent" | Byte preserved raw, not interpreted as percent | Percent chance out of 100 via Rand(missile seed,100) < byte | verified in 1.13c code (static) |
| G3 | KNB-R01 (item chance) | `Properties.knock -> item_knockback` | Property mapped to `item_knockback` stat | Roll 0..127 vs 128/64/32 by small/normal/large; requires stat > 0 | verified (asm); which ItemStatCost row has itemeventfunc 7 unresolved |
| G4 | KNB-R06a/KNB-R06d (GH vs KB fallback) | `MonStats2` `mKB`, `dKB`, `small`/`large` "feed a monster knockback profile" | `dKB` (direction count) included in knockback profile as a decision input | Only the mode-13 (mKB) bit decides; `dKB` not consulted by the decision logic | verified (asm); contradicts repo profile |
| G5 | KNB-R07a (distance) | "Path distance five with sand-leaper exception"; "path velocity 4096" hypotheses | Repo assumed distance 5 (sand-leaper exception) and velocity 4096 | Byte 5 or 10 (class 0x4E), effective step (byte>>1)+1 = 3 or 6 path units; velocity 4096 not found | step verified (asm); class name unresolved; 4096 not found |
| G6 | KNB-R04 / KNB-R06a | Older-source hypotheses "death beats knockback", "KB-incapable monsters get hit reaction" | Treated as hypotheses | Both hold | verified (asm) |
| G7 | Section 8 | Probe `knockback_probe` accepts only `1.14d`; no explicit GH-vs-KB mode field (inferred from `reaction` plus `mode_supported`); template has empty trials | Tool cannot capture 1.13c; mode inferred | n/a | tooling note |
| G8 | Section 8 | Probe schema `mechanism`: control, item_knockback, missile_knockback; per-trial fields hit, combat_blocked, lethal, uninterruptible, displacement (subtiles), reaction | Capture format | n/a | tooling note |
| G9 | KNB-R06b | "Combat will request the movement"; combat emission disabled | Combat does not emit knockback yet | Server applies reaction in ApplyCombatResultToUnit after damage setup | verified (asm) |
| G10 | Done checklist | #378 issue linkage and checklist items | Tracker state | n/a | process only |
