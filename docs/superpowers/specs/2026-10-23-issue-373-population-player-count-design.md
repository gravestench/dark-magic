# Issue #373: room population uses the shared eligible-player count

## Intent and boundary

Advance [issue #373](https://github.com/gravestench/dark-magic/issues/373) by making monsters created when a room first activates use the same effective player population as monsters spawned through `system.monster.spawn`. A dead player should not raise the spawned monster's life, experience, or pinned NoDrop ceiling simply because room activation used a different count. This is one bounded slice of #373; difficulty consumers and target-version observation remain open.

## Evidence and limits

`docs/specs/monster-scaling-by-players-and-difficulty.md` MSD-R05 and MSD-R06 are verified by static 1.13c code: the effective count excludes specified player states/modes, and spawn life and experience depend on that count. MSD-R21 caps death-time NoDrop with the count pinned at monster creation. PR #551 implemented a provisional `player_count.live_population` mapping from current ECS health and death facts for direct spawns and NoDrop. Applying the 1.13c rule to 1.14d and mapping all original unit conditions to Dark Magic ECS remain inferred; this slice does not upgrade that provenance.

## Approaches

1. **Use `player_count.live_population` in room activation (chosen).** Reuse the checkpoint-compatible policy already called by direct monster spawn and death. Keep `spawn.components` as the one materializer, so it applies override floors and pins the count. This is the smallest consistent change.
2. Change `player_count.live_population` to count all joined identities. This would alter direct spawns and NoDrop, contradicting the currently published 1.13c rule and PR #551 without new evidence.
3. Keep the room-specific count and adjust `spawn.components`. This would create a second interpretation of the same count and would not fix the inconsistent input.

## Behavior and constraints

At first room activation, filter the system's entity snapshot to active player entities before passing it to the existing shared `live_population` policy. Direct spawn already excludes `d2legacy.world.inactive`; room activation must do the same for its count without excluding inactive monster residents from the broader room system query. Declare `d2legacy.player.vitals` and `d2legacy.player.death` in the room system's read set, because `live_population` reads those components and ECS enforces declared access. Pass the resulting count, including zero, to `spawn.components`; its `monster_spawn` policy handles the minimum one and any explicit override. Preserve existing room selection, monster definition, RNG order, and previously spawned monsters' pinned counts. An independent test must drive the authoritative room-activation path with one living and one committed dead player, assert `spawn_player_count == 1` and unboosted life/XP, and show the same result for a direct spawn. Another case must prove two living players still yield the 50% bonus and count 2; also cover a living but inactive identity. No loot formula, difficulty table, party eligibility, or death-state representation changes belong in this slice.

## Anti-goals and remaining work

Do not treat online player membership alone as the effective count, introduce a new count authority, or infer 1.14d equivalence from a green Dark Magic test. Issue #373 retains its difficulty and original-runtime evidence gates after this PR. The user's dead-player recollection is an explicit 1.14d research question; a target-runtime observation could supersede the provisional mapping in a later issue slice.
