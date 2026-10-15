# Original-engine specifications

Language-independent research handoffs for Lord of Destruction Expansion 1.13c.
These documents record verified assertions and explicit unknowns; publication does not
establish implementation, user approval, gameplay confirmation or 1.14d parity.

## Index

| Specification | Specification issue | Gameplay issue |
|---|---|---|
| [Numeric and RNG foundations](core-numeric-and-rng-foundations.md) | [#469](https://github.com/gravestench/dark-magic/issues/469) | Shared foundation |
| [Game-data fields](core-game-data-fields.md) | [#470](https://github.com/gravestench/dark-magic/issues/470) | Shared foundation |
| [Damage mitigation](combat-damage-mitigation.md) | [#471](https://github.com/gravestench/dark-magic/issues/471) | [#383](https://github.com/gravestench/dark-magic/issues/383) |
| [Combat status application](combat-status-application.md) | [#472](https://github.com/gravestench/dark-magic/issues/472) | [#383](https://github.com/gravestench/dark-magic/issues/383) |
| [Leech and drain](combat-leech-and-status-effects.md) | [#473](https://github.com/gravestench/dark-magic/issues/473) | [#383](https://github.com/gravestench/dark-magic/issues/383) |
| [Monster scaling by players and difficulty](monster-scaling-by-players-and-difficulty.md) | [#474](https://github.com/gravestench/dark-magic/issues/474) | [#373](https://github.com/gravestench/dark-magic/issues/373) |

The handoffs are coordinated in [#442](https://github.com/gravestench/dark-magic/issues/442).
Each draft publication PR identifies sibling documents that still need to be published.
An index entry is a navigation reference, not a claim that its dependency is merged.

## Evidence build policy

Evidence in these handoffs targets original Expansion 1.13c. The project also targets
1.14d, but findings do not transfer between patches without corresponding evidence.
Original binary fingerprints and data-profile limitations are recorded in the
[source manifest](../research/source-manifest.md). Public companions identify original
module locations, methods, dates and the assertions supported without requiring a
private reconstruction project. No proprietary binaries, tables, captures or saves are
included. Pending gameplay observations remain unconfirmed.

## Conventions

- Each spec defines one bounded behavior through semantic inputs, rules, outputs,
  exceptions, exact examples and acceptance boundaries. Operation names impose no API.
- IDs identify a spec's rules (`R`), questions (`Q`), executable vectors (`V`) and
  observation procedures (`O`). Sub-IDs identify separately specified parts of a rule.
- IDs become permanent when an implementation plan, implementation code or tests consume
  the spec. Sibling-spec, companion and gap-file references alone do not make it consumed.
  Until then, a renumbering updates those references together. Explicit retained-number
  decisions are recorded in the companion; removed numbers are not reused as aliases.
- A rule's evidence label and integration gates determine what is supported. The Id map
  separates verified supplied-input assertions from unresolved dependency effects.
- Document status describes evidence, not implementation or maintainer approval. A draft
  can contain usable verified assertions alongside explicit gaps; it is not full parity.
- Binary addresses, ordinals and layouts belong in the public evidence companion;
  private research history stays private. Numeric widths, units, rounding, game-data
  identities and exact expected values remain in the behavioral contract.
- Each companion links back to its spec. Evidence gaps live in
  `docs/plans/gaps/<spec-name>.md`; implementation comparison appears there only when requested.
- Publication leaves the gameplay implementation issues open. An issue devoted to
  reviewing and publishing a spec can close when its approved documentation PR merges.
