# Diablo game-data records

Diablo II's tab-separated tables are the engine's authored game database. They
are not incidental loader code: their rows define items, skills, monsters,
levels, objects, sounds, experience, treasure classes, character defaults, and
most other simulation and presentation facts.

## Current recovery state

The repository still contains 91 model files representing the historical typed
table schemas. The Service Mesh retirement in commit `68e3c9e` removed the old
record-manager implementation, including typed loading and lookup behavior. The
current `internal/recordstore` provides layered, cached generic TSV rows to
native code and Lua. Typed schema binding is owned by the independent
`github.com/gravestench/tsv` codec, while `internal/gamedata` owns catalog
snapshots, indexes, diagnostics, and compatibility handling for the surviving
historical structs. Generic `map[string]string` access does not replace the typed
game-data catalog.

The schemas must be preserved during repository cleanup. The old service
lifecycle and integration wrappers should not be restored wholesale; useful
behavior should be recovered behind explicit internal ownership.

## Target architecture

The game-data subsystem should provide:

- a layered-content TSV decoder shared by typed and generic access;
- typed collections for every supported Blizzard table, retaining unknown
  columns when practical for forward compatibility and mods;
- deterministic indexes by each table's canonical identifiers and codes;
- cross-table validation with actionable source, row, and column errors;
- immutable snapshots for simulation and read-only Lua views for shim/mod logic;
- generation-aware invalidation and atomic rebuilds during content hot reload;
- explicit host ownership, without global registries or one service per table;
- synthetic coverage for every schema and optional real-archive tests proving
  that all available tables decode and their primary indexes are valid.

Gameplay systems should depend on focused catalog interfaces or immutable typed
snapshots, not parse TSV independently. Generic record access remains valuable
for data exploration, mod-defined tables, and columns the engine does not yet
model.

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
values that do not match their apparent documentation. The catalog must preserve
those facts. It must not silently "fix" source data or treat every inconsistency
as a fatal engine error.

Structural failures that make a table uninterpretable remain fatal. Recoverable
conversion and relationship problems should produce source-located diagnostics,
retain the original generic row, and allow policy at the consuming system to
decide whether the affected record matters. Validation distinguishes verified
invariants from observations and heuristics, and records may explicitly be
classified as active, unused, sentinel, malformed-but-tolerated, or unknown.
