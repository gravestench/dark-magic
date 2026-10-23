# Issue #374: first party experience-sharing slice

## Purpose and acceptance boundary

Issue [#374](https://github.com/gravestench/dark-magic/issues/374) asks the
authoritative game to share party rewards only with eligible members. A player
who kills a monster should be able to share that monster's experience with
living party members in the same level and near the victim. A far, dead, or
other-level member should not receive a party experience award. The player
credited through an owned unit uses the same party route as a direct player
kill. This first slice covers **monster-kill experience only**. Kill and quest
credit, gold sharing, hireling experience, and other reward producers remain
under #374 and their own evidence boundaries. The pull request will use
`Refs #374` and leave it open.

The repo already has authoritative party membership in
`policy/party.lua`, monster death and owner credit in
`systems/monster_death.lua`, and per-player scaling/capping in
`policy/experience_gain.lua`. Monster death currently awards only the credited
player. This design adds a party award route to that existing death path; it
does not add a second death, progression, or party authority. Offline, hosted,
replay, and checkpointed sessions continue to use the same Lua policy.

## Evidence and confidence

The merged [party experience
contract](../../specs/player-party-experience-sharing.md) states that the
1.13c server's PXL-R04a party collector filters members by the credited
player's party and level, alive state, and victim-centered squared coordinate
distance `<= 6400`, rejecting a zero X coordinate. With at least two members,
its 1.13c static instructions construct a pool
`X + trunc((n - 1) * X * 89 / 256)` and distribute shares in proportion to
member level. PXL-R05a then applies the existing level and experience-ratio
scaling to each raw share. A lone admitted member takes the solo route for the
original credited player. PXL-R09 explicitly keeps NoDrop's same-level party
enumeration independent of the victim-distance filter. These are **verified
from 1.13c static code**, as recorded by the
[evidence companion](../../research/spec-evidence/player-party-experience-sharing.md),
not 1.14d runtime observations.

The exact binary floating-point environment and alternate divide path for
PXL-R04b are **unresolved** in the
[gap record](../../plans/gaps/player-party-experience-sharing.md). The
distance-to-game-world unit correspondence and the 1.14d integer behavior
remain unresolved. PXL-R04a **does verify 1.13c signed 32-bit wrapping** in
the pool's products and sum; this implementation keeps that arithmetic. The
floating share calculation uses the contract's ordinary, finite supplied
examples and is labeled a **provisional Dark Magic interpretation**. Applying
1.13c static rules to the project's 1.14d target is **inferred**. The
original-game probe must eventually compare the boundary and share values in
both supported patches before parity is claimed. Neither the design nor its
tests promote synthetic fixtures to original-game evidence.

## Approaches considered

1. **Chosen: a narrow party XP policy called by the existing death owner.**
   It takes the immutable entity snapshot, credited player, victim position,
   and stored victim experience/level. It returns eligible recipients and
   their raw shares; `monster_death.lua` uses the existing gain and cap path to
   award them. Party membership stays in `party.lua`, and one function owns
   the provisional share math. The user approved this first-slice approach.
2. Put XP distance and pool math inside `policy/party.lua`. This would make
   the general membership owner depend on victim experience and progression
   details that NoDrop and party UI never need.
3. Put collection, pool math, and awards directly in
   `systems/monster_death.lua`. This avoids a file but mixes eligibility and
   arithmetic with corpse, loot, and event handling, making the unresolved
   floating rule harder to replace after a probe.

The first approach keeps the new rule testable and keeps NoDrop's party count
unchanged. It introduces no Go gameplay formula or dependency.

## Authoritative flow

1. On one accepted monster `unit_died` event, preserve the current ownership
   resolver's `credited_id`. The experience route first requires a credited
   player with identity and progress and a positive stored monster experience
   value. It reads the monster's current authoritative position and level.
   The existing monster-death observed marker still prevents a second award.
2. A player without a party follows the current solo award path, including
   the existing level factor, item experience bonus, class cap, and stored
   progress update. No victim-distance test applies to solo credit (PXL-R03).
   If the runtime lacks a usable victim or player position, keep that solo
   path rather than silently deleting a currently valid award; record this as
   a local missing-context fallback, not as an original-game rule.
3. For a credited player in a party, start from the party owner's ordered
   membership and the system's active entity snapshot. A member must still
   have a player identity and progress, positive health, no death marker, and
   the credited player's level id. The collector uses the victim and member
   positions, excludes a zero X coordinate, and accepts squared distance
   `<= 6400`. This is **victim-centered**, not distance from the credited
   player. The original dynamic-coordinate representation is not in the ECS:
   `d2legacy.world.position` holds `f64` coordinates. For this slice, use
   these authoritative world coordinates directly and compare their squared
   distance without a new yard/screen conversion. This is an explicit
   inference whose boundary must be checked against the original runtime.
   Missing member location/position fails that member's eligibility. The
   membership order determines recipient order for deterministic writes.
4. Freeze the eligible recipient list, each entry's level, and the level sum
   before writing any experience. If no member qualifies, make no party
   award. If exactly one qualifies, award the original credited player through
   the existing solo route, matching PXL-R04a's unusual one-member branch.
   If at least two qualify, construct the integer pool with the recovered
   `89/256` term and signed 32-bit wrapping at each product and sum. The
   supported game profile has at most eight party members;
   reject a ninth instead of silently dropping one. Nonpositive level sums
   produce no party award. Do not use effective `/players` count in this pool.
5. Divide the pool by the level sum, multiply by each member's frozen level,
   and truncate toward zero to get each raw share. Lua's numeric division is
   used provisionally for ordinary finite values; the function is isolated
   so exact 1.13c floating behavior can replace it when PXL-R04b is settled.
   Apply `experience_gain.gain` separately to each share with that member's
   class, level, resolved `item_addexperience`, and the victim's level. The
   recovered PXL-R05a returns one for a nonpositive raw share before level-cap
   checks; align that existing helper for this reachable party case and prove
   that ordinary positive solo awards do not change. Cap and write each
   member's progress with the existing `capped_total` path. Do not add a new
   progress store or client-side award.
6. The monster's death fact retains its current `xp` field as the monster's
   stored experience, not a sum of awarded shares. The present NoDrop path
   continues to count living same-level party members without this XP
   distance filter. Quest, kill-credit event projection, gold, and hireling
   recipients remain untouched in this slice.

## Validation and limits

Write failing Lua policy cases before implementation for party pool size,
member ordering, boundary distances (at and just past 80 in the current
coordinate system), zero X, absent positions, dead or marked-dead members,
other levels, and the one-member/nobody-admitted branches. Use finite examples
from PXL-V09 and explicitly test the provisional share function; do not claim
the unresolved floating boundary is verified. Exercise the production session
with near and far party members, a member dying, joining or leaving, and an
owned-unit kill. Assert actual per-player experience totals and that a far
same-level member can still affect NoDrop while receiving no party XP. Check a
solo kill and repeated death processing for regressions. Include replay or
checkpoint restoration where it distinguishes stale membership from current
party facts.

Run focused tests, the Lua suite, architecture checks, full Go tests, Lua
hardening, and race CI on the submitted commit; list exact commands and game
data skips in the pull request. A separate read-only review should check the
recipient list, arithmetic boundary, deterministic ordering, original-source
claims, and any change to solo or NoDrop behavior. Update #374 with the
delivered XP slice after merge and leave its other reward criteria open.

The builder may choose small function names and test fixture structure. The
builder must keep the provisional coordinate and floating choices visible in
code and PR evidence, rather than describing them as verified 1.14d behavior.
