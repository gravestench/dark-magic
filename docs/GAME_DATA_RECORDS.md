# Diablo game-data records

Diablo II's tab-separated tables are the engine's authored game database. They
are not incidental loader code: their rows define items, skills, monsters,
levels, objects, sounds, experience, treasure classes, character defaults, and
most other simulation and presentation facts.

## Current recovery state

The repository preserves historical row schemas under
`internal/game/data/model`. The generic `internal/game/data/store` provides
layered, cached TSV rows to native code and Lua. The small
`internal/game/data/typed` mechanism binds one explicitly requested table to a
Go schema using `github.com/gravestench/tsv` and can build a caller-selected
deterministic index.

There is deliberately no global typed game-data snapshot. The retired snapshot
hardcoded every known Diablo table into the host, eagerly made unrelated tables
mandatory, and encoded d2legacy composition policy in generic Go. The first-party
mod now chooses the records it needs through `engine.records/v1`; Lua owns joins,
required/optional decisions, and gameplay meaning.

The schemas must be preserved during repository cleanup. The old service
lifecycle and integration wrappers should not be restored wholesale; useful
behavior should be recovered behind explicit internal ownership.

## Target architecture

The game-data subsystem should provide:

- a layered-content TSV decoder shared by typed and generic access;
- generic immutable rows that retain unknown columns for forward compatibility
  and mods;
- optional caller-selected typed binding for native codec/adapter boundaries;
- deterministic indexes only where a narrow caller explicitly requests one;
- d2legacy-owned cross-table validation with actionable table/row/column errors;
- authoritative Lua state/checkpoint identity that pins interpreted rule state;
- generation-aware per-table invalidation during development hot reload;
- explicit host ownership, without global registries or one service per table;
- synthetic coverage for retained schemas and focused real-archive tests proving
  only the tables required by the tested path.

Authoritative gameplay systems load immutable generic rows through the versioned
records capability and interpret them inside d2legacy. Format-level binary facts
that affect simulation belong to the same pinned generation: the renderer-free
`engine.animdata/v1` capability exposes typed fixed-point timing/event records,
while d2legacy chooses what an attack marker means. Native code may depend on a
focused adapter value, never an application-wide catalog or snapshot.

## Authentic-data policy

### Research sources and verification order

The repository includes the community-authored Diablo II Data File Guide at
`docs/diabloiidatafileguide.html.gz`. It is the primary semantic reference when
implementing or reviewing a Blizzard TSV table: consult its table overview,
field descriptions, value domains, formulas, and cross-table links before
deciding what a column means. Keep the document compressed; it can be searched
without producing a multi-megabyte working-tree artifact:

```sh
gzip -cd docs/diabloiidatafileguide.html.gz | rg -n -i 'treasureclass\.txt|NoDrop'
```

The guide documents accumulated reverse-engineering knowledge, not guaranteed
invariants for every patch or mod. Use this order of authority:

1. The layered game data supplies the bytes the running engine must preserve.
2. The Data File Guide supplies intended semantics and known relationships.
3. Existing model comments and historical code are implementation leads only.

For every newly admitted typed table, review the relevant guide section, test
representative documented fields, and run the optional real-archive test. When
the guide and shipped data disagree, retain the shipped value and record the
disagreement as a diagnostic or test comment; do not silently normalize it.
The uber list file is useful for asset/table discovery, but does not define TSV
semantics.

The shipped tables are known to contain unused rows and columns, sentinel
values, dangling references, contradictory fields, patch-era leftovers, and
values that do not match their apparent documentation. The record layer must preserve
those facts. It must not silently "fix" source data or treat every inconsistency
as a fatal engine error.

Structural failures that make a table uninterpretable remain fatal. Recoverable
conversion and relationship problems should produce source-located diagnostics,
retain the original generic row, and allow policy at the consuming system to
decide whether the affected record matters. Validation distinguishes verified
invariants from observations and heuristics, and records may explicitly be
classified as active, unused, sentinel, malformed-but-tolerated, or unknown.
