# Review of the issue #442 specification handoffs

Review date: 2026-10-15. Branch: `docs/442-spec-review`, base commit `e191ea7b`
("consolidate 31 specification handoffs for review"). Standard: the spec-writing skill
(`SKILL.md`, `template.md`, `checklist.md`, `provenance.md`, `finding-functions.md` and the
`check_spec.py` lint).

## Summary

The 31 specs are strong work. Every spec passed the lint before and after review. Rules
agreed with the 1.13c code in the large majority of re-derivations, and the 12 new specs had
every executable vector recomputed, with no mismatches beyond the few listed below. The
main weaknesses are size, duplication between specs, a few unanchored dispatch chains,
and vectors that do not yet discriminate well enough: fresh blind readers still found
guesses, gaps and discrimination problems in specs whose companions recorded clean passes.

| Grade | Meaning | Specs |
|---|---|---|
| A | Ready for user review | 3 |
| B | Minor fixes made | 24 |
| C | Significant fixes made (a rule the code contradicted) | 4 |
| D | Not fit, needs rework | 0 |

## What was done

1. **Lint** on all 31 specs: 0 errors, 0 warnings before and after.
2. **Checklist review** of every non-lint item in `checklist.md` section 1, per spec, by nine
   independent reviewers.
3. **Evidence spot-checks** against d2-decomp 1.13c (asm, byte-matched C where it exists,
   and the shipped DLL for undecoded regions): every executable vector in the 12 new specs
   recomputed; in the rewritten specs, every surprising or cross-spec rule plus a random 10%
   sample, read first and compared second.
4. **Follow-up re-derivations** for checks that had not finished: the PvP spot-check and the
   Tainted Sun quest-stage rules (13 claims, none contradicted, two incomplete).
5. **Fixes in place** for clear violations and verified errors (below), one commit per area.
6. **Blind-implementer passes** (Sonnet readers, Opus triage, every behavioral edit checked
   against the asm) on the nine specs whose rules or vectors changed. Every run and every
   finding outcome is recorded in the spec's companion.

## Grades

| Spec | Grade | Notes |
|---|---|---|
| core-numeric-and-rng-foundations | B | R17a mis-described; renamed to avoid a clash with MSD-R19 |
| core-game-data-fields | B | Companion only; vectors not in template shape |
| combat-shield-block-and-avoidance | B | V07 traces contradicted its outputs; ~46 rule ids |
| combat-hit-recovery | B | Reaction function had no dispatch root; added |
| combat-knockback | B | Wrong label cited for HRC-R15 |
| combat-damage-pipeline-order | C | Ratio-128 scaling wrong; R02b, R08b, R15 incomplete or inconsistent |
| combat-damage-mitigation | B | Stale wording |
| combat-leech-and-status-effects | B | Field names not mapped to the pipeline spec |
| combat-status-application | B | Freeze callback attributed to the wrong spec |
| state-timed-state-lifecycle | C | Non-monster death path described wrongly |
| state-movement-modifiers-and-freeze | B | 32 ids; repeats status-application rules |
| combat-item-event-triggers-and-procs | B | Cited parent ids that only exist as sub-ids |
| combat-crushing-blow | B | Settled questions left open |
| combat-open-wounds | A | |
| item-durability-loss | B | 1.14d markers contradicted the build policy |
| combat-pvp-damage-and-kill-credit | A | Kill-credit chain now rooted at the frame loop |
| missile-creation-and-initialization | B | ~40 rule ids |
| missile-damage-parameters | B | |
| missile-hit-and-damage-calculation | B | Unlabeled alias stubs to missile-damage-parameters |
| missile-movement-and-collision | B | Conflicting RVA for one table entry |
| player-mana-regeneration | B | |
| player-stamina-drain-while-running | A | |
| skill-leveled-damage-tohit-and-elements | C | Missed a hit-class override write |
| skill-cast-start-and-do-dispatch | C | Partial target resolution described wrongly |
| skill-calc-expressions-and-synergies | B | ~43 ids; duplicates leveled-damage builders |
| skill-assassin-sentry-firing-and-range | B | AI routine had no dispatch root (partly closed) |
| skill-assassin-blade-sentinel-and-fury | B | R05a overclaimed; R01 summon arguments wrong |
| skill-assassin-death-sentry-corpse-damage | B | Undefined `RandAlt`; missing gates |
| monster-scaling-by-players-and-difficulty | B | Stale contradiction with the party XP spec |
| player-party-experience-and-loot-sharing | B | |
| quest-tainted-sun-world-clock | B | Quest-stage behavior incomplete; one vector wrong |

## Behavioral errors found and fixed

Each was confirmed against the 1.13c code before the fix; each companion records the
original discrepancy and the new evidence.

| Spec | Rule | What was wrong |
|---|---|---|
| combat-damage-pipeline-order | R05, R06a, R06b, R21b | Said some terms are unchanged at ratio 128. Only the elemental fields skip the multiply; drain, poison, status lengths and burn always multiply and divide by 128, so values of 2^24 or more wrap. New V22 case discriminates. |
| combat-damage-pipeline-order | R02b, R08b, R15 | Missing unarmed floors and 16-bit sign extension; treated a verified PvP rule as open; cited the wrong leech rules and draw timing. |
| state-timed-state-lifecycle | R13a, V19 | Said non-monster deaths run the partition. The cleanup reads a missing monster record and faults; both callers are monster death paths. |
| skill-cast-start-and-do-dispatch | R07, V13 | A partly resolved target passes the raw coordinates with a (0,0) destination, not four zeros. Later passes also fixed R05a, R02/R09a and R06. |
| skill-leveled-damage-tohit-and-elements | R18 | Missed that a nonzero Skills field replaces the hit class after the elemental fill. |
| skill-assassin-blade-sentinel-and-fury | R05a, R01 | The missile follows its owner reference, not the creature; summon creation receives caster, skill, level and position (0,0) only. |
| skill-assassin-sentry-firing-and-range | R02 | Every reject issues one death request; owner checks have a fixed order; any negative counter initializes. |
| quest-tainted-sun-world-clock | R01e, R02, R07, R10d, V21(i) | Quest-state writer and event handler were incomplete (always records the triggering unit, stage 0 becomes 2, a separate event 0x28 branch); altar path, fatal sender aborts and player resolution added; V21(i) expected output wrong. |
| core-numeric-and-rng-foundations | R17a | The creation test reads the MonStats `enabled` bit against Fog's constant bit table, not an engine-held mask. |
| combat-status-application | sections 1, 3, 9 | The freeze expiry callback belongs to MMF-R13c, not STL-R11. |
| combat-shield-block-and-avoidance | V07 | Hand traces contradicted the (correct) outputs. |

## Other fixes

- Dispatch roots added where a verified rule had none: hit-recovery reaction function, PvP
  kill credit, timed-state R10a, and (partly) the sentry and Death Sentry AI table entries.
- 13 imported functions given their ordinal and RVA; asm labels corrected to RVAs.
- Wrong or stale cross-spec references and dependency lists corrected.
- Gap files updated to current rule numbers and stripped of rename artefacts
  (`item_combat-knockback`, a stray heading).
- Words run into numbers fixed across about 40 files (`CollideFriend0`, `isMissile1`,
  `Rand 10029`).
- 1.14d markers removed from the durability spec to match the 1.13c build policy.

## Blind-implementer passes

Two rounds. Round 1 ran on the nine specs whose rules or vectors changed in review; round 2
finished or continued the seven that were not clean. Every run ended with MISMATCH 0: no
reader found a wrong rule. What remains is vector discrimination, wording and coverage.

| Spec | Runs (both rounds) | Final GUESS / GAP / DISCRIM | State |
|---|---|---|---|
| core-numeric-and-rng-foundations | 1 delta | 0 / 0 / 3 (no change needed) | Clean |
| skill-assassin-death-sentry-corpse-damage | 1 delta | 0 / 1 / 0 (no change needed) | Clean |
| state-timed-state-lifecycle | 3 full, 3 delta | 0 / 1 / 2 (resolved) | Clean |
| skill-leveled-damage-tohit-and-elements | 2 full, 4 delta | 2 / 3 / 0 (resolved by wording edits) | Clean apart from an unchecked final wording edit |
| skill-assassin-blade-sentinel-and-fury | 6 full | 2 DISCRIM open | NextHit 2 case and the R02 timer wrap need vectors, then a full run |
| skill-cast-start-and-do-dispatch | 8 full | 2 entries open | R02/R06 wording overstates GDF-R08a; V13 must assert the R07 latch order |
| quest-tainted-sun-world-clock | 8 full | 4 / 9 / 9 | A further full run; stage return values unspecified |
| skill-assassin-sentry-firing-and-range | 4 delta, 4 full | 4 / 9 / 11 | Recorded entries to apply, then a full run |
| combat-damage-pipeline-order | 8 full | 5 / 4 / 11 | Recorded entries to apply, then a full run |

Round 2 also fixed behavior against the asm: sentry R10 (selector 7 excludes slot A's unit,
a missing path-result branch); cast start R02 (reset force=1 skips the acceptance check);
Blade Sentinel R01 to R04 arguments; Tainted Sun R01g (fatal path when the quest structure
is missing) and R06b (the intensity request carries the derived act).

## Recommendations (not applied)

**1. Split oversized specs** (SKILL.md section 2 bound of about 25 rule ids):

| Spec | Ids | Suggested split |
|---|---|---|
| combat-shield-block-and-avoidance | ~46 | Hit roll; block and avoidance |
| skill-calc-expressions-and-synergies | ~43 | Expression evaluator; synergies. Move damage builders into leveled-damage |
| missile-creation-and-initialization | ~40 | Creation and seeding; initialization and registry |
| core-game-data-fields | 36 | By record family |
| state-movement-modifiers-and-freeze | 32 | Drop the rules owned by status-application first |
| quest-tainted-sun-world-clock | 31 | Quest stages; world-clock arithmetic |
| core-numeric-and-rng-foundations | 31 | RNG and integers; record layouts and ids |
| skill-assassin-sentry-firing-and-range, ...-death-sentry-... | ~30 each | By trigger |
| combat-pvp-damage-and-kill-credit | 25 | Hostility declaration; kill credit |
| player-party-experience-and-loot-sharing | | Gold-pickup sharing |

**2. Remove duplication by consuming, not restating:** MMF-R07..R14a restate
status-application; LSE-R01 and LSE-R16 restate the pipeline; MMC-R12a and V09 restate
MCI-R26a and V12; missile-hit keeps unlabeled alias stubs (update `coverage.md`, which cites
`MHD-R01`, before removing them); IEP-R19 should consume LSE-R14 as crushing-blow does.

**3. Finish the blind-implementer passes** in the table above: cast start and Blade Sentinel
are one edit and one run away; Tainted Sun, sentry firing and pipeline order need their
recorded entries applied and further full runs. Fresh readers keep finding discrimination
gaps, so budget several runs each.

**4. Settle names and anchors:** give `Unit +0xC6` bit 0 one name in game-data-fields (it
has four across specs); trace the sentry and Death Sentry AI index to its MonStats column;
record the drop routine's skip when the treasure total is 0 as a rule in monster scaling
(it is only in the companion).

**5. Open questions worth prioritizing:** TSW-Q11 (the start routine writes cycle 0 but the
resync packet sends cycle 5); SFR-Q11 (monsters and players use opposite hostility
polarity); SLD-Q10; the item-event callback draw counts.

**6. Strengthen the lint:** it missed words run into numbers in forms like `CollideFriend0`,
`isMissile1`, `in(b)` and `|2026`; 1.14d markers that contradict the build policy; and gap
files that use retired rule numbers or rename artefacts.

**7. Clarify the skill:** state whether the reader-run cap is per task or lifetime, and when
a rewritten rule needs a full run instead of a delta check (two reviewers judged this
differently); require a root-anchored chain before any rule is labeled verified, since three
specs had verified rules without one.

**8. Readability:** combat-hit-recovery's rule text is very compressed; knockback's rule ids
are out of execution order (renumbering is allowed while nothing consumes the spec).

## Where things are

- Fixes: 9 review commits, 2 reader-pass commits and the report on `docs/442-spec-review`
  after `e191ea7b`, pushed to origin.
- Per-group review notes and every reader-pass record: each spec's companion under
  `docs/research/spec-evidence/` ("Independent review (2026-10-15)" and the reader runs).
- All specs remain `draft`: none has the runtime confirmation the done bar requires.
