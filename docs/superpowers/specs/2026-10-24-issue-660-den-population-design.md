# Issue #660: ordinary Den of Evil population

## Intent and evidence

Entering the generated Den currently shows an empty cave. The player should
encounter ordinary hostile residents selected from the Den's owned 1.14d area
rows. This supplies the combat population needed before a later kill-all quest;
it does not implement that quest. The Den remains a click-activated cave warp.
Rogue Encampment, Blood Moor, Cold Plains, and Stony Field retain their visible,
seamless outdoor joins.

Owned Expansion 1.14d `Levels.txt` level 8 has `NumMon=3`, all three difficulty
`MonDen` fields equal to `600`, and `mon1`–`mon3` / `nmon1`–`nmon3` equal to
`zombie1`, `brute1`, and `fallenshaman1`. The matching `MonStats.txt` rows are
enabled ordinary spawn candidates. The previous-level Den chamber definitions
83–86 in `LvlPrest.txt` have `Populate=1`. Tests will assert these fields from
owned data. The original game's density arithmetic, group construction, and
random call order have not been verified; existing Dark Magic spawning policy
remains explicitly provisional.

## Boundary

The existing `Prepared.PopulationCommand` remains the sole admission command,
and `d2legacy.bootstrap.population` remains the sole owner of family selection,
density rolls, placement decisions, activation, and persisted resident state.
This change adds only ordinary level-8 residents and the geometry needed for
safe placement. It does not add quest flags, rewards, champions, unique packs,
new cave generation, cave audio, or a separate spawn authority.

## Approaches

1. **Extend the existing plan (chosen).** Publish a level-8 zone after the
   ordered outdoor zones when a validated Den cave pair exists. The same Lua
   plan activates its one room on entry and retains its residents across exits
   and recovery. This gives Den the established identity and checkpoint path.
2. **Create a cave-specific command and lifecycle.** This would isolate the
   cave payload but duplicate candidate selection, room activation, and
   recovery state. Two authorities would make later quest kill counts harder to
   trust, so this is rejected.
3. **Spawn from the scene or warp handler.** That would couple durable combat
   entities to a presentation transition and cause different offline/connected
   behavior. The current architecture rules this out.

## Data and admission

`Prepared.PopulationData` keeps seam order for levels 2–4. If `CaveWarp` is
present, it requires the keyed level-8 zone and materialized map to agree with
the cave pair's level and dimensions, then appends exactly one level-8 zone.
The zone contains the actual room bounds, stamp `Populate` decision, links,
and deterministic open candidate points reachable on player-sized collision
from the validated cave arrival. The level-8 search continues past open but
disconnected center pockets instead of filling its four candidates from them.
The cave pair supplies two protected
locations in level 8: its arrival and return-warp source. Their coordinates
and clearance radii are published as cave-only exclusion geometry. Missing or
mismatched cave facts fail before the command is admitted; old prepared
topologies without a cave keep their historical outdoor-only payload.

The Lua command validator accepts the legacy ordered outdoor prefixes `[2]`,
`[2,3]`, and `[2,3,4]`, or the exact cave-bearing order `[2,3,4,8]`.
Level 8 requires a finite cave exclusion envelope with exactly the two
trusted endpoint locations and positive clearances. The arrival clearance
covers the player footprint; the return-source clearance covers the authored
interactive warp radius (currently 3.5 subtiles). Existing flat
single-zone commands and old outdoor plans remain readable. This is an
additive level-specific payload under the current plan schema, so the schema
remains `d2legacy.population.plan/v6`. The Den gets a dedicated random stream
prefix, preserving prior level-2/3/4 draws and IDs.

## Placement and lifecycle

The first population phase with a player in the Den's room uses the level-8
`Levels.txt` row and the existing `MonStats.txt` interpreter. It performs the
existing provisional density, rarity, and group policy, with the current
first-populated-room fallback so a safe eligible room is useful. Before each
spawn, the population owner checks the entire collider footprint against
decoded terrain, room bounds, existing level-8 colliders, and both cave
endpoint clearances. Unsafe points are skipped with an inspectable reason; if
no point remains, no unsafe resident is created. Identifiers use the existing
`level:8:room:<id>:monster:<member>` form.

The existing active-room lifecycle deactivates Den residents after the player
warps out and restores the same entities when a player returns. Death and
combat components remain on those entities. Player-scoped projection keeps
Den residents out of outdoor views. A current-binary checkpoint containing a
level-8 plan restores it exactly. A current-binary restore of a legacy v6 plan
without level 8 does not synthesize missing residents. Cross-binary migration
is outside this issue.

## Checks

- Synthetic Go and Lua tests reject missing/mismatched keyed cave geometry,
  malformed new level-8 envelope fields, unsafe points, duplicate IDs, and
  incorrect zone ordering; old v1–v3 outdoor payloads retain their shape.
- Owned 1.14d tests pin the exact `Levels`, `MonStats`, and `LvlPrest` fields,
  then cover Normal/Nightmare/Hell seeds, all available Den chamber variants
  including seed 42's disconnected center pocket, deterministic zone
  publication, eligible families, arrival-reachable room points, endpoint
  clearance, and at least one Den hostile after entry. Owned cases must run
  locally and explicitly skip when the MPQ is unavailable.
- Offline and connected pointer journeys enter the cave, see authoritative
  Den residents, leave and return with stable IDs and combat/death state,
  reconnect, and restore a current-binary checkpoint. The existing outdoor
  seamless-join and population tests remain green.
- Run the repository's full Go/Lua/architecture/lint/build gates and race CI
  on the submitted commit. Record exact commands, owned-data availability,
  and any unverified native-pixel behavior in the pull request.
