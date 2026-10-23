# Issue #464: end forced motion on level transitions

## Intent and acceptance boundary

Issue [#464](https://github.com/gravestench/dark-magic/issues/464) closes a
transaction leak: an active `d2legacy.world.forced_motion` survives a seam or
warp, so movement on the next tick resumes toward a point in the previous
level. Every admitted level relocation must record exactly one `interrupted`
`d2legacy.world.forced_motion_event` using the position and accumulated
distance before the relocation, then remove the active transaction. Checkpoint
restore must produce the same outcome. The next tick must not produce a second
outcome or continue that motion.

This is an engine audit invariant for an already admitted forced-motion
transaction. Combat production of knockback (#378), connected-client event
playback, and handling a pending `forced_motion_request` that has not yet been
admitted are outside this issue. An invalid destination or rejected interaction
must leave the active transaction untouched.

## Evidence and limits

The repository's
[movement research](../../research/MOVEMENT_COLLISION_AND_PATHFINDING.md)
describes the knockback transaction and its one-outcome contract. The merged
interruption work (#465) supplies `forced_motion.interrupt(context, entity,
structural)`, which records progress at the current position and removes the
transaction. `transition.apply` currently relocates the player and stops
ordinary motion, while both `world_transition` and `world_warp` call it.

Original 1.13c and 1.14d behavior for knockback across a level transition is
unknown. This change cannot be presented as verified original-game parity; an
owned runtime observation would be needed to make that claim.

## Approaches

1. **Chosen: interrupt in the shared transition transaction.** After
   destination validation and required-component lookup, call the existing
   forced-motion interruption once, then change level, position, bounds, and
   velocity. Seam and warp callers pass the tick and an appropriate structural
   writer. This gives every current relocation path the same ordering and
   outcome without another policy owner.
2. Interrupt independently in the seam and warp callers. This avoids changing
   `transition.apply`'s call signature but duplicates the ordering rule and
   allows later transition callers to forget it.
3. Clean up after relocation in movement. This retains the old-level target
   until another tick and records the destination position as the end point,
   violating the issue's acceptance criteria.

## Design

Change `transition.apply` to receive the transition tick context and a writer
with `create` and `remove`, in addition to the entity and destination. Validate
the destination and fetch the location, position, bounds, and velocity first;
only then call `forced_motion.interrupt`. Relocate immediately afterward. This
ordering prevents malformed transitions from emitting a false interruption and
ensures the event uses the pre-transition position and existing
`applied_distance`. The interruption helper itself remains the sole owner of
event fields and removal.

The seam system already has an ECS system callback. Receive its `context` and
`structural` writer, pass both to `transition.apply`, and declare the
`forced_motion`/`selectable` reads and `forced_motion`/event writes required by
`forced_motion.interrupt`. Its seam matching and destination policy remain
unchanged.

The warp runs inside an admitted interaction command. Thread that command
through `commands/interactions.open` and `interactions/operations.apply` to the
warp handler, using a fourth argument that existing operation handlers may
ignore. The warp passes `command.tick` as context and a small writer whose
`create`/`remove` methods call the existing immediate ECS API. This introduces
no persistent state or second event policy. Range rejection still happens
before the operation handler and therefore before any interruption.

The existing checkpointed forced-motion component and outcome event remain the
authority. No game data, schemas, Go engine behavior, or presentation path
changes are needed.

## Verification

Use test-driven development in the existing seam and warp Lua tests. Each path
must enter a level with a partially progressed active forced-motion
transaction and assert one `interrupted` event with the original level's end
position, stable target identity, original start and requested distance, and
the progress to date. Assert destination state, absent active motion, and no
second event or motion on the next tick. Restore a checkpoint immediately
before each transition and compare the recorded outcome and resulting state.
Keep a rejected, stale warp interaction active to prove it emits no event and
does not remove the transaction. Focused tests and the broad repository checks
must run on the submitted commit; the PR must record exact results and any
unverified original-runtime behavior.
