# libd2 research ledger

Repository: [jaenster/libd2](https://github.com/jaenster/libd2)

Pinned research revision: `e6cdc4927c6180be8dd309b0423b470f64f1fc6c`
(inspected 2026-08-10).

## Why it matters

libd2 is a clean-room Zig reimplementation of the Diablo II 1.14d engine core.
Its DRLG package covers all five acts and exposes rooms, presets, adjacency,
collision, roads, objects, monsters, and warps from a seed. Its verification
documentation reports cell-for-cell comparison against retail-engine captures,
including blind holdout seeds. That makes it unusually valuable for testing
Dark Magic's future deterministic map generator and for identifying behavior
missing from the older 1.10f D2MOO reconstruction.

Useful areas include:

- DRLG seed derivation and RNG call order;
- preset, maze, and outdoor generation rules;
- Act-specific placement and border logic;
- room graphs, level adjacency, warps, and coordinate transforms;
- tile generation, substitutions, and collision materialization;
- preset object and monster population;
- DS1/DT1 parsing and the boundary between formats and generation;
- deterministic verification, trace, and holdout-test design.

## Evidence and licensing boundary

The clean-room source is MIT-licensed. The repository license and README state
that binary blobs and fixtures under `packages/*/src/blobs` and
`packages/*/src/maps` are derived from Blizzard data and are not covered by the
MIT license. Dark Magic must not copy or redistribute those assets.

Treat libd2 as strong independent implementation evidence, not automatically as
the original engine specification. Dark Magic should restate observed behavior
independently and corroborate it with lawful owned-asset probes, original-runtime
captures, D2MOO, historical documentation, or agreement among implementations.
Keep 1.10f/1.14d differences visible rather than merging them into one rule.

## Initial audit queue

1. Compare seed initialization and RNG call order with `docs/MAP_GENERATION.md`
   and D2MOO for representative preset, maze, and outdoor levels.
2. Compare level/room graph output and coordinate conventions for all acts.
3. Catalog Act-specific generator rules that are absent or unresolved in the
   existing Dark Magic research specification.
4. Design non-proprietary structural traces: room bounds, adjacency, selected
   presets, warp endpoints, collision hashes, object counts, and RNG checkpoints.
5. Add blind holdout seeds before implementing each generator family so source
   inspection cannot overfit Dark Magic's acceptance tests.
6. Record any 1.10f versus 1.14d conflict in the source matrix before selecting
   a compatibility policy.
