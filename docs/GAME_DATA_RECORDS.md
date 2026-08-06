# Diablo game-data records

Diablo II's tab-separated tables are the engine's authored game database. They
are not incidental loader code: their rows define items, skills, monsters,
levels, objects, sounds, experience, treasure classes, character defaults, and
most other simulation and presentation facts.

## Current recovery state

The repository still contains 91 model files representing the historical typed
table schemas. The Service Mesh retirement in commit `68e3c9e` removed the old
record-manager implementation, including typed loading and lookup behavior. The
current `internal/recordstore` correctly provides layered, cached generic TSV
rows to native code and Lua, but `map[string]string` access does not replace the
typed game-data catalog.

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
