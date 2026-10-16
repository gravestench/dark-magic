# Specs

Implementation-independent specifications of original Diablo II (Lord of Destruction 1.13c) behavior. Each is written so anyone can implement the behavior from scratch, with no knowledge of any particular codebase. Source: issue
[#442](https://github.com/gravestench/dark-magic/issues/442) (parent
[#362](https://github.com/gravestench/dark-magic/issues/362)). Each spec is named after the
thing it defines and covers one bounded behavior. The issue and tracker columns below map the
specs to the project's tracker and are the only place issue numbers appear.

## Index

### Core (`core-`)

| Spec | Issues | Parent |
|------|--------|--------|
| [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md) | #469; shared by all specs | #442 |
| [core-record-layouts-and-ids](core-record-layouts-and-ids.md) | #469; shared by all specs | #442 |
| [core-game-data-fields](core-game-data-fields.md) | #470; shared by all specs | #442 |
| [core-game-data-table-records](core-game-data-table-records.md) | #470; shared by all specs | #442 |

### Combat (`combat-`)

| Spec | Issues | Parent |
|------|--------|--------|
| [combat-hit-roll](combat-hit-roll.md) | #488 (spec); #381, #383 (gameplay) | #348 |
| [combat-shield-block-and-avoidance](combat-shield-block-and-avoidance.md) | #488 (spec); #381, #383 (gameplay) | #348 |
| [combat-damage-pipeline-order](combat-damage-pipeline-order.md) | #483 (spec); #383 (gameplay) | #348 |
| [combat-damage-mitigation](combat-damage-mitigation.md) | #471 (spec); #383 (gameplay) | #348 |
| [combat-leech-and-drain](combat-leech-and-drain.md) | #473 (spec); #383 (gameplay) | #348 |
| [combat-status-application](combat-status-application.md) | #472 (spec); #383 (gameplay) | #348 |
| [combat-crushing-blow](combat-crushing-blow.md) | #491 (spec); #383 (gameplay) | #348 |
| [combat-open-wounds](combat-open-wounds.md) | #492 (spec); #383 (gameplay) | #348 |
| [combat-knockback](combat-knockback.md) | #489 (spec); #378, #379 (gameplay) | #345 |
| [combat-hit-recovery](combat-hit-recovery.md) | #490 (spec); #383, #378, #375 (gameplay) | #348 |
| [combat-pvp-damage-and-kill-credit](combat-pvp-damage-and-kill-credit.md) | #493 (spec); #383, #420 (gameplay) | #348 |
| [player-hostility-declaration](player-hostility-declaration.md) | #493 (spec); #383, #420 (gameplay) | #348 |
| [combat-item-event-triggers-and-procs](combat-item-event-triggers-and-procs.md) | #482 (spec); #383 (gameplay) | #348 |

### States (`state-`)

| Spec | Issues | Parent |
|------|--------|--------|
| [state-movement-modifiers-and-freeze](state-movement-modifiers-and-freeze.md) | #494 (spec); #375 (gameplay) | #346 |
| [state-timed-state-lifecycle](state-timed-state-lifecycle.md) | #481 (spec) | #442 |

### Skills (`skill-`)

| Spec | Issues | Parent |
|------|--------|--------|
| [skill-assassin-sentry-firing-and-range](skill-assassin-sentry-firing-and-range.md) | #496 (spec); #418 (gameplay) | #408, #387 |
| [skill-assassin-sentry-missile-launch](skill-assassin-sentry-missile-launch.md) | #496 (spec); #418 (gameplay) | #408, #387 |
| [skill-assassin-blade-sentinel-and-fury](skill-assassin-blade-sentinel-and-fury.md) | #497 (spec); #419 (gameplay) | #408, #387 |
| [skill-assassin-death-sentry-corpse-selection](skill-assassin-death-sentry-corpse-selection.md) | #498 (spec); #420 (gameplay) | #408, #387 |
| [skill-assassin-death-sentry-corpse-damage](skill-assassin-death-sentry-corpse-damage.md) | #498 (spec); #420 (gameplay) | #408, #387 |
| [skill-cast-start-and-do-dispatch](skill-cast-start-and-do-dispatch.md) | #486 (spec); #387 (gameplay) | #442 |
| [skill-leveled-damage-tohit-and-elements](skill-leveled-damage-tohit-and-elements.md) | #487 (spec); #387 (gameplay) | #442 |
| [skill-calc-expressions](skill-calc-expressions.md) | #485 (spec); #387 (gameplay) | #442 |

### Missiles (`missile-`)

| Spec | Issues | Parent |
|------|--------|--------|
| [missile-creation-and-seeding](missile-creation-and-seeding.md) | #499 (spec); #387 (gameplay) | #442 |
| [missile-initialization-and-registry](missile-initialization-and-registry.md) | #499 (spec); #387 (gameplay) | #442 |
| [missile-movement-and-collision](missile-movement-and-collision.md) | #500 (spec); #387 (gameplay) | #442 |
| [missile-hit-and-damage-calculation](missile-hit-and-damage-calculation.md) | #501 (spec); #383, #387 (gameplay) | #442 |
| [missile-damage-parameters](missile-damage-parameters.md) | #502 (spec); #387 (gameplay) | #442 |

### Items (`item-`)

| Spec | Issues | Parent |
|------|--------|--------|
| [item-durability-loss](item-durability-loss.md) | #495 (spec); #383 (gameplay) | #348 |

### World generation (`drlg-`)

| Spec | Issues | Parent |
|------|--------|--------|
| [drlg-world-clock](drlg-world-clock.md) | #505 (spec); #415 (gameplay) | #346 |

### Monsters (`monster-`)

| Spec | Issues | Parent |
|------|--------|--------|
| [monster-scaling-by-players-and-difficulty](monster-scaling-by-players-and-difficulty.md) | #474 (spec); #373 (gameplay) | #344 |

### Player (`player-`)

| Spec | Issues | Parent |
|------|--------|--------|
| [player-party-experience-sharing](player-party-experience-sharing.md) | #503 (spec); #373, #374 (gameplay) | #344 |
| [player-party-gold-sharing](player-party-gold-sharing.md) | #503 (spec); #373, #374 (gameplay) | #344 |
| [player-mana-regeneration](player-mana-regeneration.md) | #484 (spec) | #442 |
| [player-stamina-drain-while-running](player-stamina-drain-while-running.md) | #504 (spec); #346, #377 (related gameplay) | #346 |

### Quests (`quest-`)

| Spec | Issues | Parent |
|------|--------|--------|
| [quest-tainted-sun](quest-tainted-sun.md) | #505 (spec); #415 (gameplay) | #346 |

Read [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md) first: it defines the random
number generator, fixed-point formats, integer rules and time units the other specs rely on;
[core-record-layouts-and-ids](core-record-layouts-and-ids.md) defines the record layouts, stat reads and
state/mode operations; [core-game-data-fields](core-game-data-fields.md) defines the names, identities and
bindings they consume; and [core-game-data-table-records](core-game-data-table-records.md) defines the
game-data columns, getters and table loading. Original binary mappings are recorded in the
linked evidence companions. Suggested order after that: combat-damage-pipeline-order, combat-damage-mitigation,
combat-shield-block-and-avoidance, then the rest in any order.

## Candidate handoffs for existing issues

These are proposed research boundaries. Existing issues supply the questions; the
[coverage map](coverage.md) supplies discovery leads. Neither establishes a verified rule,
creates a new spec file, or expands an existing spec's acceptance boundary. The maintainer
reviews new boundaries before drafting. Original 1.13c evidence does not settle a 1.14d check.

| Issue | Existing coverage / proposed additional boundary | Evidence still needed |
|---|---|---|
| #463 | Status application covers poison application; life regeneration/drain needs its own per-frame consumer contract | Player/monster tick roots, negative regeneration, eligibility, rounding, death and town behavior |
| #462, #449 | Mitigation consumes resistance; creation and display of player resistance are separate producers | Original derived-stat and display paths, selected table cells and effective caps |
| #461 | Monster scaling/mitigation do not fully specify initial resistance materialization | MonStats difficulty selection, six resistance writes and their creation consumers |
| #401 | Focus Fire Golem death/replacement explosion separately from the broad summon-family candidate | Both triggers, target selection, radius, damage and ordering |
| #377 | Running stamina drain does not establish spending Vitality or rescaling current stamina | Attribute-spend handler and ordered current/maximum stamina writes |
| #376 | Path collision candidates need route-versus-step category coverage | Player, hireling, summon, townsfolk and corpse conditions at both consumers |
| #374 | Party XP has an existing spec; kill/owned-unit/quest credit and gold require their own boundary assessment | Each consumer's eligibility/distance rules; do not infer universal sharing from XP alone |
| #466 | Inventory monster skill bindings, then review the existing nine proposed behavior families | Include shared player skills used by monsters, AI selection/handoff, server/client handlers, missiles/summons and difficulty |
| #467 | Playback, positional sound, ambience and music selection/transition candidates | D2Client selection roots, D2Sound execution, effective table bindings and exact transition/loop/fade behavior |
| #468 | Light-map/radius, ambient lighting, palette/fade candidates and the existing world-clock consumer | Original rendering consumers and how world/quest state reaches them |

Issue #464 specifies an implementation transition invariant; original-game transition behavior
needs separate evidence before becoming a behavioral spec. Issue #459 is engineering test
cleanup and does not itself require an original-game behavior spec. Broader roadmap issues
aggregate focused handoffs rather than define one engine-wide spec.

## Evidence build policy

- Evidence is collected on **Diablo II: Lord of Destruction Expansion 1.13c**, from the
   original binary instructions and an owned game. Original artifacts are identified in the
   [source manifest](../research/source-manifest.md). No other build is read or observed for
   these specs.
- The project's issues also name 1.14d as a target. These specs make no claim about 1.14d; the
  `confirmed_1.14d` field that older specs still carry on their vectors is retired and means
  nothing. Specs written or revised after 2026-10-12 omit it.
- Static evidence reads reconstructions checked against original bytes when available,
   otherwise disassembly. Public records identify original modules, relative function
   locations, exports, methods and dates. Private working filenames, paths and reconstruction
   priorities remain in the private audit; public provenance does not require that project.
- A rule is **runtime-confirmed (1.13c)** only after one of its vectors is reproduced on an
  owned 1.13c game with a valid matching capture and a record in linked evidence (or legacy
  inline Provenance). Records identify the cases/assertions actually observed; neither a
  mismatch nor missing records promotes a rule. Table cells use **verified in 1.13c data**
  and do not require or establish gameplay observation. Document status follows the
  spec-writing skill's aggregation policy; in-scope gated rules remain explicit.
- Other builds (for example 1.10f, 1.14d), D2MOO and community formulas are leads, never a
  source of constants.

## How a spec is organized

Below the title, a brief **Description** explains the behavior in ordinary game language.
It precedes status (`draft`, `data-verified (1.13c)`, `code-verified (static only: <reason>)`
or `runtime-confirmed (1.13c)`), evidence build, core-spec links, ID policy and the
companion evidence link. Status describes evidence, not implementation or user approval.
Data-only status makes no code-verification claim; a runtime header names its observed coverage
through linked records and does not imply exhaustive branch observation. Then:

1. **Scope:** the behavior defined and its boundary.
2. **Trigger and ownership:** the module that computes the result, behavioral trigger and
   eligibility, when it runs within a frame, and how the client learns the result. Detailed
   dispatch chains with an RVA on every link live in the linked evidence file.
3. **Inputs and data:** every stat, state, mode, unit field, record field and data-table
   column the rules read, with width and format; data values as single named entries with virtual
   path, row and column. Relevant mappings/constants distinguish Origin from Binding, including
   table identities explicitly selected by code. Original bindings do not imply custom remapping.
4. **Rules:** numbered, in execution order, in language-neutral pseudocode with integer
   widths, signedness, rounding, caps and RNG draws, each with a status label, ending in an
   **Id map** (id, summary, label, gating open questions, vectors or table evidence for
   data-only rules).
5. **Outputs and side effects:** every stat write, state change, scheduled event, unit created
   or destroyed, and message sent (name only), with the rule that performs it.
6. **Branches and edge cases:** unit types, difficulty, melee vs missile, PvP, zero and
   negative inputs, overflow, failure paths.
7. **Open questions:** everything unresolved, with what evidence would settle it.
8. **Test vectors:** executable vectors (every input, RNG seed or `no draw`, ordered draws,
   hand trace, exact outputs, `runtime_observed`) and observation summaries linked to
   tool-neutral reproduction procedures for an owned 1.13c game in the evidence file. Mixed
   vectors distinguish verified and gated assertions; partial acceptance is not full reproduction.
9. **Acceptance and dependencies:** the rule and vector ids that constitute the behavior, and
   the behavioral contracts consumed and known consumers. Plans derive delivery order by
   checking repository gaps; supporting references or reverse consumers are not prerequisites.
10. **Related specs.**
- **Appendix A: Provenance:** evidence build, original modules, methods/read dates, runtime
   status and links to the source manifest and `docs/research/spec-evidence/<spec-name>.md`.
   The companion links back and preserves original module/RVA/export references, dispatch
   chains, observation procedures, public search/runtime records and caveats. Private corpus
   details remain in a separate audit. Existing inline records move when their spec is
   revised; relocation adds no verification.

Older specs may retain the earlier seven-section layout and inline provenance until revised.
Requested migrations preserve their behavioral contracts; the new layout alone is not fresh
verification.

Status labels, exactly one per rule or sub-id: *runtime-confirmed (1.13c)*, *verified in 1.13c
code (static)* (function identified from a dispatch root and the output write, arithmetic traced
entry to write, every callee within the stated boundary resolved and read; boundary crossings
consume named dependency contracts with their labels and integration gates), *verified in 1.13c data* (a name, id or value
read from the 1.13c data files through the game's loader; name maps and cited values only),
*asm-only*, *recovered; unverified*,
*unresolved*, *not found in d2-decomp* (all six rungs of the search ladder run and recorded).
Read each spec's header, Id map and evidence records for its current status; the index's presence
and the coverage map's `written` marker do not imply approval, complete verification or implementation.

## Conventions

- File names: `<category>-<thing>.md`, kebab-case, in one flat folder. The category prefix
  follows the original game's own subsystem (1.13c source folder or data table); skill specs
  add the class (`skill-assassin-...`). The prefix list lives in the spec-writing skill
  (`.agents/skills/spec-writing/template.md`). No dates or issue numbers. Rule ids do not
  depend on the file name.
- [coverage](coverage.md) maps every original behavior found by the discovery sweep to its
  spec, or marks it as a candidate still to write.
- Size: one bounded behavior per spec. About 25 rule ids, fewer than about 8, or independent
  triggers prompt scope review; they do not automatically split or merge a spec. The user
  decides structural changes. The accepted combat-damage-pipeline-order exception stays one
  spec because it defines a single ordered path. Length is not a criterion.
- Ids (`SBA-R07`, `SBA-Q03`, `SBA-V03`, `SBA-O01`): until a spec is consumed (an
  implementation plan in `docs/plans/` outside `gaps/`, implementation code or tests cite its
  file or id prefix), its ids may be renumbered and no former numbers are kept. References in
  sibling specs, companions and gap files alone do not make it consumed; renumbering updates
  those references together. Once consumed, ids are permanent, never renumbered or reused;
  an existing historical alias belongs only in the Id map's `Formerly` column. Sub-ids carry
  one label each.
- A spec body states the current contract; revision and migration records live in its
  companion evidence file.
- Specs never mention a particular implementation. `docs/plans/gaps/<spec-name>.md` holds the
  evidence gaps a verification or migration pass could not close and, when requested, the
  comparison between an implementation and the original; it exists only when it has content.
  Implementation plans live in `docs/plans/`, never here.
- Behavioral contracts use semantic field and operation names. Export ordinals, binary
  addresses, assembly paths, registers, stack positions, memory offsets, packed keys and
  original internal layouts belong in the linked evidence file. Keep numeric semantics,
  game-data identities and exact test values in the spec; pseudocode names impose no API.
- A requested format/evidence migration preserves scope, file boundaries, arithmetic, RNG
  order and vector results (and ids for a consumed spec), then attempts to close the spec's
  evidence gaps with targeted static source and table-data reads. Labels change only on new
  recorded evidence; gaps left open go to the gap file. Original source/read/observation
  records stay in the companion. Gameplay campaigns, implementation gap analysis and
  structural changes are separately scoped work.
- Share original descriptions, pseudocode and vectors. A single named value from a game data
  table (`CharStats.ManaRegen (Sorceress) = 120`) may be cited where a rule or vector needs it.
  Do **not** copy rows, columns, proprietary binaries or source, and never quote reconstructed C.
