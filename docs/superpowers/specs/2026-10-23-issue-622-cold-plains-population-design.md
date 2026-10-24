# Issue #622: Ordinary Cold Plains Population Design

## Intent and evidence

Issue #622 makes the live Cold Plains from #620 a populated wilderness area using the already authoritative
Blood Moor room-population lifecycle. Players should meet ordinary level-3 monsters in offline and connected
sessions, leave them inactive when their rooms leave the active window, and recover the same population state
after a checkpoint. This is a child of the Act I journey #354 and relates to monster roadmap #356.

The owned 1.14d `patch_d2.mpq` `data/global/excel/Levels.txt` row `Id=3` records `NumMon=3`, `MonDen`,
`MonDen(N)`, and `MonDen(H)` of 520, and the three selected ordinary candidate codes `brute1`, `corruptrogue1`,
and `fallenshaman1` in the current interpreter. A populated fourth slot is ignored because the current code
stops at `NumMon`; that is interpreter behavior, not a retail claim. The selected `MonStats.txt` rows are
enabled, spawnable ordinary monsters. Cold Plains' current `LvlPrest.txt` stone-fill definitions mark
population-eligible rooms. The existing 520/100000 roll, first-group fallback while `created==0` (which still
yields no group when a room has no usable points), rarity weights, group sizes, and room eligibility are Dark
Magic policy; original 1.13c/1.14d density and RNG call order are not established. Tests and docs must keep that
distinction.

## Approaches

- **Chosen: one versioned, level-keyed plan in the existing population owner.** The entry-world adapter supplies
  ordered zone facts for the admitted wilderness levels, and one system-owned bootstrap command installs them
  atomically. Lua retains one room-lifecycle system and one checkpointed plan state. This fits the current
  command, residency, replay, and authority boundaries. Independent architecture and evidence subagents reviewed
  the approach on 2026-10-23.
- Separate per-level plan states and commands under the same module. This adds admission ordering and
  partial-state cases, and does not preserve old full checkpoints because the state registration count and
  runtime identity both change.
- Eagerly create all level-3 hostiles at world install. This bypasses the current current-room-plus-neighbors
  activation policy and changes early random draws, work, and resident lifecycle.

## Prepared population facts and admission

`entryworld.Prepared` builds population facts from the actual admitted, materialized level-2 and level-3 zones
in canonical order. A restored version-1 two-zone entry topology emits level 2 alone; no level-3 plan is
synthesized. A missing map or zone, duplicate level, zone request ID inconsistent with its keyed level,
malformed externally supplied point, or invalid link fails before command admission. A materialized map has no
intrinsic level ID; the prepared world's keyed association is trusted after the zone request is checked. Level 2
retains its current room, anchor, and point order, including the development fixture's near-player behavior. Its
existing `OpenPointNearSubtile` projection checks a radius-zero center cell. Level 3 computes its own
room-local, distinct center points from its own map and rooms; it must not reuse the level-2 player spawn as a
nearby anchor. A generated room whose anchors are all blocked is valid with zero points and creates no group.
After choosing a monster definition, Lua filters the level-3 candidate points through the existing collision
registry's `destination_clear(level, x, y, definition.collider_radius)` for full-radius terrain safety and room
bounds. Before each materialization it also rejects overlap with current same-level active colliders and with
level-3 spawns already queued in this activation tick, using deterministic candidate order and the sum of
collider radii. It uses up to the available safe distinct points; zero safe points produces zero spawns. Level-2
placement remains unchanged. The envelope is `{ zones: [zone2, zone3] }` (or `[zone2]` for v1 topology), with
each zone carrying the existing `seed`, `act`, `level_id`, `difficulty`, `rooms`, and `links` fields.

`PopulationCommand` stays the single privileged tick-1 command, as used by offline and listen composition. Lua
validates every zone, uniqueness, canonical ordering, difficulty, room and link facts, and duplicate
installation before one state replacement. The existing flat single-zone command remains valid for current
synthetic callers, normalized internally into the same envelope. No presentation or host-side code may make
monster-family or density decisions.

## Authoritative lifecycle and determinism

The existing `d2legacy.population.plan` state becomes an explicit version-6 envelope with ordered `zones`; each
zone retains independent `created`, room activation, active status, and inactive-resident records. The
room-lifecycle system walks zones in canonical order, qualifies room and resident matches by both level ID and
room ID, and activates only occupied rooms and their neighbors in that level. Crossing 2→3→2 retains already
created monsters and does not duplicate them. Stable spawn IDs remain level qualified. Existing level-2 random
stream names and room order remain unchanged; register distinct level-3 density, family, group, seed, and
monster-life streams so visiting Cold Plains cannot alter future Blood Moor draws. `spawn.components` takes an
optional life-stream name for this population path while retaining the current default for level 2 and every
existing caller. Connected clients observe authority snapshots and do not spawn their own hostiles.

Changing the registered plan from v5 to v6 changes runtime identity and the Go state-store schema. A version-1
**entry-world topology restored under the current runtime** can bootstrap its level-2 plan and then
checkpoint/restore normally. A pre-v6 **full session checkpoint** cannot be silently accepted by the current
exact-identity recovery gate; cross-version migration is outside the current issue #622 scope and must not
weaken that gate. The #622 tests must exercise the actual current-runtime recovery entry point, not only decode
a Lua table.

## Acceptance and verification

- Synthetic admission: both levels in fixed order; all three difficulties; unique candidate/room/level facts;
  malformed, missing, and duplicate facts fail atomically. Blocked generated level-3 rooms retain zero points
  and zero spawns. Synthetic activation checks terrain footprints, occupied positions, and same-tick group
  spacing. Version-1 topology admits only level 2. Existing Blood Moor command and activation fixtures retain
  stable IDs and random draw order.
- Owned assets: generate and materialize Cold Plains from owned MPQ data, check level-3 candidate
  selection/provenance, distinct room-local points, terrain-safe and nonoverlapping full-radius spawned
  footprints, ordinary monster definitions, and reachable room activation. Game data and captures remain outside
  the repository; dependent tests skip in CI and run locally.
- Production path: walk into Cold Plains through the real seam; observe authoritative level-3 monster identity,
  location, and visibility offline and over QUIC; return to Blood Moor, reenter Cold Plains, and verify
  residents are not recreated. Current/neighbor rooms activate; distant rooms deactivate with their resident
  state.
- Recovery: checkpoint in Cold Plains, restore under the same runtime and topology, repeat 3→2→3 and compare
  authoritative state and semantic events to uninterrupted continuation. Reject a different plan/topology before
  publication or restore. Check the restored version-1 two-zone topology retains level-2-only population.
- Documentation: update monster population research with verified owned table fields, Dark Magic policy, and
  original-game unknowns. Report exact final-commit local and CI checks in the PR.

## Scope

This issue touches only entry-world population facts, the d2legacy Lua population owner, the existing spawn
materializer's optional stream argument, registered random streams, composition/bootstrap integration, tests,
and directly relevant research docs. It does not add champions, bosses, quest monsters, exact retail density or
generator RNG parity, NPCs, waypoint activation, other exits, or cross-version full-checkpoint migration.

## Sources

- [Issue #622](https://github.com/gravestench/dark-magic/issues/622), parent
  [#354](https://github.com/gravestench/dark-magic/issues/354), monster roadmap
  [#356](https://github.com/gravestench/dark-magic/issues/356), and live crossing
  [#620](https://github.com/gravestench/dark-magic/issues/620).
- `internal/mod/d2legacy/adapter/entryworld/population.go`, `bootstrap.go`, and
  `internal/content/d2legacy/lua/d2legacy/bootstrap/population.lua` own the current Blood Moor path.
- `docs/research/MONSTERS_SPAWNING_AND_AI.md` and `docs/research/NEXT_GAMEPLAY_IMPLEMENTATION_SEQUENCE.md`
  record evidence and policy limits.
