# Foundation research decisions

These are the architectural conclusions supported strongly enough by current Dark Magic structure and the inspected research sources to guide future implementation without waiting for every compatibility probe.

## Keep

- one layered VFS and one typed game-data catalog;
- immutable/pinned authoritative data generations;
- fixed-tick session ownership and deterministic replay/checkpoints;
- explicit RNG owners/state rather than hidden globals;
- `.d2s` as an adapter around a canonical durable character, not the live schema;
- layered/parameterized stat sources with provenance;
- one authoritative item location per stable item identity;
- atomic item/stat-source transitions;
- data-driven class/progression definitions;
- Lua as presentation/snapshot/intent, never gameplay authority.

## Prototype before committing

- raw ordered table document and legacy BIN compatibility representation;
- exact scheduler ordering where original same-frame behavior matters;
- legacy save codec section model/version support;
- stat dependency/invalidation cache strategy;
- richer item instance composition around current placement state;
- progression fixed-point representation after owned-data vectors exist.

## Do not claim yet

- complete legacy TXT->BIN equivalence;
- exact original timing for all periodic systems;
- exact `.d2s` write compatibility/checksum until owned round trips pass;
- complete ItemStatCost op/cap/rounding fidelity;
- complete equipment/corpse/trade semantics;
- exact per-class numeric progression formulas until owned-data/original-runtime probes are recorded.
