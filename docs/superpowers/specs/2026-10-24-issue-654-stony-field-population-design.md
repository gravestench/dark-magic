# Issue #654: Ordinary Stony Field population

## Intent and evidence

Characters can walk into Stony Field through the live outdoor seam introduced by
issue #648, but its active rooms contain no ordinary monsters. The existing
authoritative room population plan deliberately marks level-4 rooms activated
without reading candidates or consuming monster random draws. This issue makes
eligible Stony Field rooms produce and retain ordinary monster residents through
that same plan. The gameplay reason is to make the third outdoor area useful for
combat while preserving earlier areas and checkpoint continuation.

Owned Expansion 1.14d `Levels.txt` row `Id=4` has `NumMon=3`; its first three
`mon*` and `nmon*` entries are `skeleton1`, `zombie2`, and `crownest1`, and its
three difficulty density columns are `520`. Slots four and five name
`goatman1` and `cr_archer1`. The five matching `MonStats.txt` rows are enabled,
spawnable, and non-NPC. These are table facts. Selecting the first `NumMon`
entries, density scaling, rarity and group rolls, room timing, and random-call
order are current Dark Magic policy, not proven 1.13c/1.14d retail behavior.
The monster research document must retain this distinction and name what
further original-code or retail observation would settle.

## Goals and boundaries

- Use the existing ordered population plan, candidate interpreter, spawn
  command, room residency, and checkpoint authority for level 4. No second
  population owner or population-plan schema change is introduced.
- Generate ordinary residents only in eligible occupied or neighboring active
  rooms with `populate` policy and collision-safe candidate points. Keep their
  identities and active state stable when leaving, returning, and restoring.
- Give level 4 dedicated population random streams, including monster life, so
  visits to Stony Field do not advance levels 2 or 3. Preserve the current
  level-2 and level-3 stream names and results.
- Select `mon*` for Normal and `nmon*` for Nightmare/Hell, plus the matching
  density column, through the current interpreter. Synthetic fixtures with
  distinct values establish column selection; owned rows alone cannot.
- Keep legacy v1/v2 topology snapshots at their original map and population
  set. Level 4 must not appear during their recovery.

Champion, unique, boss, quest, and named packs such as Rakanishu remain outside
this issue. Exact retail density, family frequency, group size, and RNG parity
remain open research. Seam rendering, warps, other areas, and unrelated
refactoring are outside the boundary.

## Approaches

1. **Extend the current population owner (selected).** Remove the level-4
   first-activation bypass. Use the level-3 room-local point, full-footprint,
   queued-spawn, and returning-resident safeguards for level 4; allocate
   `d2legacy.population.level4.*` streams. The Go geometry adapter publishes
   room-centered points for both levels 3 and 4. This is the smallest change
   that preserves one authority and isolates prior-area outcomes.
2. **Introduce a separate Stony Field spawner.** This would duplicate candidate
   interpretation, random draws, active-room lifecycle, and checkpoint state,
   making both authority and save continuation ambiguous. Reject it.
3. **Generalize every zone's population policy at once.** This could remove
   level checks, but would alter Blood Moor's established placement and random
   behavior without an issue or evidence for doing so. Reject it for this
   slice.

The selected approach was independently reviewed for architecture and owned
data provenance. Level 2 keeps its unqualified stream names; level 3 keeps its
existing `level3` names. Level 4 alone gains the new behavior.

## Production behavior and proof

The host continues to publish only immutable room geometry and candidate
points. Lua remains the decision maker for level rows, family selection,
density, group size, and spawned entity components. It validates that a level-4
zone is the third zone of the four-map topology, activates occupied and linked
rooms, and materializes eligible groups once. Existing room deactivation and
restoration continue to own monster and item residents. Full collider
footprints are reserved before queued or returning spawns, including across
the first return after checkpoint restore.

Synthetic tests must prove all three difficulty columns, candidate slots
1–`NumMon`, ordinary `enabled`/`isSpawn`/non-`npc` eligibility, sparse or empty
candidates, blocked or duplicate points, deterministic level-4 continuation,
and unchanged level-2/3 monster IDs, positions, health, and random streams.
Owned 1.14d production tests must find a seed with eligible Stony Field
residents, traverse from Cold Plains offline and connected, observe their
authoritative and public identities/positions, checkpoint in level 4, and
compare restored leave/return behavior with uninterrupted play. Explicitly
skip owned-data tests when MPQs are absent and report exactly which checks ran.
Legacy v1/v2 topology recovery tests under the new binary must prove no
level-4 population is synthesized. Registering level-4 random streams changes
the strict checkpoint stream set; cross-binary restoration of checkpoints
written before this issue is not promised by this design and needs separate
migration evidence if required. Replace issue #648's explicit no-monster and
no-draw level-4 assertions with level-4 population tests.

## Builder calls and remaining uncertainty

Use the existing local test fixture and its chosen seed where it yields an
eligible room; if it does not, search deterministic seeds in test setup and
pin the first successful one. Do not hard-code a game seed into production.
Update `docs/research/MONSTERS_SPAWNING_AND_AI.md` with exact inspected fields,
test evidence, and the retail rules still unknown. If owned data reveals a
conflict with this design, revise issue #654's evidence before growing scope.
