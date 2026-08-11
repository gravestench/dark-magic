# Shrines, wells, curses, buffs, and temporary environmental effects research

Status: implementation-oriented research baseline. Dark Magic already loads `Shrines.txt`, has item/stat foundations, fixed-tick scheduling, semantic world objects, and the combat/state system research needed to own timed effects. The missing work is the authoritative shrine/well operation and reusable temporary-effect layer.

## Executive conclusion

Shrines/wells are world-object operations that produce **player/world/item/monster state transactions**. Timed shrine buffs belong in the shared state/stat-source system, not on the shrine object and not in UI timers.

Conceptual flow:

```text
Shrine/Well ObjectInstance
  current shrine type / used-reset state
        |
        v
player operate request
        |
        v
validate object/range/availability
        |
        v
Shrine behavior family
  heal / mana / refill / exchange / portal / gem
  storm / monster / exploding / poison
  combat / defense / skill / stamina buff
        |
        +--> resource mutation
        +--> item generation/transformation
        +--> missile/combat/state request
        +--> monster/world/portal request
        +--> timed state/stat source
        `--> object cooldown/used state
```

Presentation shows the shrine animation, overhead phrase, overlay and audio after the authoritative result.

## Shrines.txt is already typed in Dark Magic

The current model exposes:

- shrine type/name/effect text;
- behavior `Code`;
- `Arg0`, `Arg1`;
- duration in frames;
- reset time;
- rarity;
- view name and activation phrase;
- effect class;
- minimum level.

This is sufficient to build a normalized shrine definition without hard-coding each specific shrine row by display name.

## D2MOO behavior-family evidence

Pinned 1.10f `ObjMode` has explicit shrine callbacks including:

- Health;
- Fill Mana;
- Refill;
- Exchange Health;
- Exchange Mana;
- Enirhs/reverse shrine behavior;
- Portal;
- Gem;
- Storm;
- Monster;
- Exploding;
- Poison;
- Combat Boost;
- Stamina;
- Defensive Boost;
- Skill Boost.

A generic shrine operate function dispatches through a shrine callback table.

This strongly supports normalizing `Shrines.Code` -> semantic behavior handler while keeping `Arg0/Arg1/Duration` data-driven.

## Shrine spawn/type selection

The shrine object and the shrine effect definition are separate concepts. The same shrine-like world object can receive a selected shrine type according to level/region rarity constraints.

The normalized runtime should preserve:

```text
ShrineObjectState
  ObjectID
  ShrineDefinitionID
  Seed/selection context
  Available/Used
  ResetTick or one-shot policy
```

Do not infer shrine type from the rendered sprite variant alone.

Exact regional shrine selection/rarity/effect-class rules need source/probe work.

## Timed buff state

Combat/defense/skill/stamina shrines create temporary effects with an expiration.

Use the shared timed state/source model:

```text
StateInstance
  State/Shrine source identity
  AppliedTick
  ExpiresTick
  stat modifiers
  source object/entity
```

The shrine object does not need to remain active for the buff to persist.

State expiration belongs to authoritative simulation ticks and checkpoints.

## Refresh/overwrite behavior

Using a new shrine while another temporary shrine effect is active can interact through state/source rules.

Potential behaviors to verify:

- replace another shrine buff of the same effect class;
- replace any shrine buff;
- stack unrelated shrine effects;
- refresh duration;
- preserve stronger/longer effect;
- clear curses or negative states in well/refill paths.

`EffectClass` is likely relevant to mutual exclusion but exact semantics require runtime/data verification.

Do not choose a universal stacking policy until tested.

## Resource shrines/wells

Health/mana/refill/well effects should mutate authoritative resources through player/stat APIs.

Possible affected resources/states:

- life;
- mana;
- stamina;
- poison;
- curses;
- chill/freeze;
- other negative states.

Wells may also affect minions/hirelings or apply partial restoration rules. This requires targeted probes.

Do not set UI orb values directly.

## Exchange shrines

Exchange Health/Mana families imply nontrivial transfers/percentage calculations. These should use explicit integer/fixed-point resource vectors.

Record exact preconditions, caps and rounding before compatibility implementation.

## Gem shrine

Gem Shrine behavior interacts with actual item identity and inventory state.

It should operate through item authority:

```text
inspect eligible gem(s) / player inventory according to rule
 -> choose/upgrade or generate result
 -> atomically transform/create item
 -> place result legally
```

The client must not name the upgraded gem grade if the original behavior chooses it authoritatively.

Exact gem selection/upgrade/fallback behavior needs owned-data traces.

## Portal shrine

Portal shrine should emit a semantic portal/transition creation request to the portal/transition system. It should not spawn a visual-only portal.

Eligibility and destination are world/game state.

## Storm/exploding/poison shrines

Negative or offensive shrine effects should reuse combat/missile/state systems.

Examples:

- storm damage/missiles;
- explosions;
- poison state/damage;
- monster-affecting effects.

Shrine code owns selection/parameters, not duplicate damage implementations.

## Monster shrine

Monster shrine behavior interacts with nearby monster quality/state. It should call the monster/unique-modifier system through a semantic transformation request.

Research needs to establish:

- candidate search range;
- eligible monster types/qualities;
- normal -> champion/unique behavior;
- selected modifiers;
- failure/no-target behavior;
- XP/loot effects of transformed monsters.

## Shrine object reset

`Shrines.txt` exposes reset timing. The object instance should store an authoritative availability state and future reset tick.

Possible semantics:

```text
ResetTime == one-shot
or
Used -> ResetTick -> Available
```

Exact conversion from table reset units to 25 Hz frames needs verification because historical table revisions/comments differ.

Do not use wall-clock minutes.

## Wells

Wells are object behavior rather than a `Shrines.txt` row in some original implementations. Architecturally they can still share:

- object operate validation;
- resource/status cleanup primitives;
- reset/cooldown state;
- audio/presentation events.

Keep a general `TemporaryEffect/RestoreBehavior` library below object behavior, not a shrine-only monolith.

## Curses and temporary environmental states

Some object/shrine effects can apply negative timed states. The shared state engine should support source/category metadata so UI/audio can explain where an effect came from.

Suggested source kinds:

```text
Skill
Item
Shrine
Well
ObjectTrap
MonsterModifier
Quest/Environment
```

This is useful for replacement/removal policies.

## Presentation and audio

Shrine interaction can emit:

- operate sound;
- activation phrase/text;
- overlay/light/color effects;
- looping buff/state effects;
- portal/object visuals.

These consume semantic events/state. See [GAMEPLAY_AUDIO_BEHAVIOR.md](GAMEPLAY_AUDIO_BEHAVIOR.md).

## Suggested implementation slices

### SH1 — normalized shrine definitions

Map typed `Shrines.Code/Arg/Duration/Reset/EffectClass` into immutable semantic definitions and inspection output.

### SH2 — one resource shrine

Implement Health or Mana shrine through generic world object operation and authoritative resource mutation.

### SH3 — one timed shrine buff

Implement stamina/defense/combat buff as a source-tagged timed state with resettable shrine object.

### SH4 — well

Implement a well/refill object using shared resource/state cleanup and object reset.

### SH5 — item/world/combat shrine families

Add Gem, Portal and one hostile/offensive shrine through item/transition/combat boundaries.

## Verification backlog

1. Shrine type selection by region/level/rarity/effect class.
2. Exact `Shrines.Code` callback mapping by target version.
3. Arg0/Arg1 meaning for every shrine family.
4. Duration frame semantics and reset-time units.
5. Shrine buff mutual exclusion/EffectClass behavior.
6. Refresh/replace/stack ordering.
7. Health/mana/refill/well resource percentages and rounding.
8. Negative state/curse/poison removal from wells/refill.
9. Hireling/pet effects from wells/shrines.
10. Health/Mana exchange arithmetic and minimum/cap behavior.
11. Gem Shrine selection/upgrade/create rules.
12. Portal Shrine destination and eligibility.
13. Storm/explosion/poison damage/state behavior.
14. Monster Shrine candidate/quality transformation.
15. Shrine object reset after room inactivation/return.
16. Multiplayer simultaneous use and per-player/global availability.
17. Shrine phrase/overlay/audio mapping and duration.

## Primary sources inspected

- Current Dark Magic typed `Shrines.txt`, object/world/stat/item/audio foundations.
- D2MOO pinned 1.10f shrine/object callbacks in `OBJECTS/ObjMode.h/.cpp` and related player/item/monster/state paths.
