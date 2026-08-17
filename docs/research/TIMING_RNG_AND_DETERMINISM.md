# Timing, RNG, simulation ticks, and determinism

> Architecture note: authoritative D2 rules may execute in the pinned
> `d2legacy` Lua runtime. The Go session remains the owner of clocks, ordering,
> RNG primitives, registered durable state, checksums, replay, and atomic
> restore. The determinism requirements below apply equally to Go and trusted
> Lua handlers.

Status: implementation-oriented research baseline. Dark Magic already owns a deterministic fixed-step session and replay boundary. This document defines how Diablo-style timers and random streams should fit it and records the original 1.10f seed arithmetic that is currently supported by strong reverse-engineered evidence.

## Executive result

Keep three clocks separate:

```text
authoritative simulation ticks
presentation/animation time
wall-clock / entropy time
```

Only the first belongs in gameplay state. Wall time may create a new seed or drive host scheduling, but it must not be read from inside replayable rules. Presentation can interpolate freely and must not alter authoritative timer order.

Dark Magic's current `internal/game/session` already serializes commands, sorts them canonically, advances ECS by a fixed step, captures external state participants, and records replay checkpoints/checksums. That is the correct owner. Diablo timers and RNG should become explicit ECS fields or registered deterministic state participants, not hidden package globals.

The same rule applies to Lua: future-affecting state must live in ECS or an
explicitly registered, versioned engine-owned store. Arbitrary globals,
closures, userdata, and native resources are not checkpoint state. An
authoritative runtime also needs deterministic table/record traversal where
order affects results, controlled numeric conversions, declared ECS access, and
per-tick instruction/time and memory budgets with an atomic failure policy.

## Current Dark Magic baseline

**verified (Dark Magic):**

- `Session.Step` executes exactly one authoritative tick.
- per-tick commands are ordered by actor/player, sequence, then command kind before application;
- `AdvanceWithSource` samples local intent once per fixed tick and bounds catch-up;
- checkpoints include ECS plus registered non-ECS state participants;
- replay records the fixed `StepNanos`, initial state, commands, and checkpoints;
- architecture documents a 25 Hz simulation policy independent of renderer cadence.

The 25 Hz Dark Magic default is therefore an existing engine contract. Whether every legacy Diablo II subsystem should be interpreted as exactly one 25 Hz game-frame step is a compatibility question to verify per system rather than an excuse to add another clock.

## Original 1.10f seed state

D2MOO reconstructs the core seed as two adjacent unsigned 32-bit values:

```text
low
high
```

`SEED_InitLowSeed` sets:

```text
low  = supplied seed
high = 666
```

One roll computes the unsigned 64-bit value:

```text
next64 = high + 0x6AC690C5 * low
```

and stores that full value back over the low/high pair:

```text
next low  = low 32 bits of next64
next high = high 32 bits of next64
```

Limited rolls use:

```text
nMax <= 0              -> 0
nMax power of two      -> low32(roll) & (nMax - 1)
otherwise              -> low32(roll) % nMax
percentage             -> low32(roll) % 100
```

Confidence: **high** for D2MOO's 1.10f reconstruction.

Synthetic vectors from that arithmetic:

| input low | input high | next64 | next low | next high |
| --- | --- | --- | --- | --- |
| `0x00000001` | `0x0000029A` | `0x000000006AC6935F` | `0x6AC6935F` | `0x00000000` |
| `0x00000000` | `0x0000029A` | `0x000000000000029A` | `0x0000029A` | `0x00000000` |
| `0x12345678` | `0x0000029A` | `0x0797CA9403BA0CF2` | `0x03BA0CF2` | `0x0797CA94` |
| `0xFFFFFFFF` | `0xFFFFFFFF` | `0x6AC690C595396F3A` | `0x95396F3A` | `0x6AC690C5` |

These are good unit vectors; they do not prove which subsystem consumes which stream.

## Entropy is not stream evolution

D2MOO also reconstructs a function that creates a random value from `time(NULL)`, a supplied value, and `GetTickCount`. This separates two concerns:

- acquiring a new nondeterministic seed;
- evolving a deterministic seed after it has been chosen.

Dark Magic should allow wall-clock/OS entropy only at a composition boundary that creates a session seed. The chosen seed becomes recorded session input. No authoritative system may call current time, `math/rand` globals, or OS randomness afterward.

## RNG ownership

A single global RNG is call-order fragile. Prefer explicit owners:

```text
RandomStream
  stable stream ID
  algorithm/version
  seed state or root seed + counter
  roll count
```

Examples: map/zone, room/preset, monster spawn group, monster AI where required, drop event, individual item generation, quest/encounter, stochastic skill/missile effect.

Dark Magic already uses named streams for its deterministic map-generation policy. Gameplay should follow the same ownership discipline. When a compatibility algorithm is known to use one legacy seed in a particular call order, preserve that algorithm's internal sequence rather than splitting it and changing results.

## Per-item RNG

D2MOO's item data includes both an item seed object and an initial/start seed. That supports retaining generation provenance so a particular item can be reproduced, serialized where compatible, and debugged without rerolling immutable facts on load.

## Scheduled events

Authoritative scheduling should use integer ticks.

For composite actions backed by `AnimData.d2`, Dark Magic advances the authored
24.8 frame cursor at the format's 25 Hz rate. A marker on zero-based frame `f`
at speed `s` is scheduled after `max(1, ceil(f * 256 / s))` ticks; completion is
the first cursor wrap, `ceil(frames * 256 / s)`. The effective binary and codec
schema are part of the immutable game-data generation. This keeps gameplay
headless and deterministic while presentation consumes the same raw timing
facts independently.

```text
ScheduledEvent
  due tick
  stable owner/domain
  event kind
  stable sequence
  payload
```

Convert source “frames,” milliseconds, or seconds once at the boundary with a documented rounding rule. Same-tick events need deterministic ordering. A safe modern default is due tick, domain/priority, stable owner/entity ID, sequence, kind. If original behavior needs insertion order, record that sequence explicitly.

## Periodic effects

Poison, regeneration, aura pulses, AI think cycles, object timers, quest timers, and missile lifetime can use:

```text
start tick
next trigger tick
end tick or remaining ticks
period
source identity
stack/refresh policy
```

Exact legacy semantics still need per-system traces: first trigger, inclusive end, refresh phase, stacking, fixed-point distribution, and what survives death/save/transition.

## Animation frames versus gameplay ticks

Renderer frame numbers are not authoritative gameplay clocks. Prefer semantic action events such as attack release/missile/skill at authoritative ticks. Where COF frame events drive compatibility timing, derive the action schedule at action start rather than making gameplay poll renderer playback.

## Host cadence and catch-up

A slow host frame may execute several ticks; a fast renderer frame may execute zero. Authoritative state must satisfy:

```text
Advance(5 * step) == five calls to Step()
```

Rendering FPS, video playback, sleep jitter, or Lua UI work must not change RNG call order inside a tick.

## Replay/checkpoint requirements

RNG state belongs in checksum coverage if it affects the future. Store it in ECS/a `StateParticipant`, or make it derivable from immutable seed plus a covered counter. Hidden mutable RNG state is unacceptable.

Checkpoint restore must restore timer queues and RNG streams atomically with ECS/item/quest authority.

## Save/reload

Persist only RNG/timer state that actually survives the selected persistence model. Per-item seed is likely durable item-format data; live poison/AI timers are session state unless evidence says otherwise. A modern realm checkpoint may persist more than legacy offline `.d2s`; keep those formats separate.

## Deterministic debugging

Optional trace records should include tick, stream ID, pre-state/counter, operation, result, post-state/counter, and semantic caller/event ID. This makes original-game comparisons practical.

## Versioning

RNG algorithm and event-order policy are replay-format concerns. Legacy streams should identify the target, e.g. `d2-seed/1.10f`, rather than vague `legacy`.

## Implementation slices

1. authoritative tick/deadline audit;
2. explicit RNG state/participant with trace hooks;
3. independent legacy D2 seed implementation using the vectors above;
4. deterministic scheduler with snapshot/restore;
5. periodic-effect harness proving batched-step equivalence;
6. replay diagnostics including RNG/timer state and content generation;
7. owned-game timing probes before encoding legacy boundary rules.

## Acceptance criteria

- Same initial state, content fingerprint, commands, and seed state produce identical checkpoints.
- The same pinned mod/dependency/configuration identity is required for replay
  and restore; identity drift is rejected unless an explicit migration applies.
- `Advance(N*step)` equals N `Step` calls.
- Renderer FPS does not change gameplay RNG traces.
- RNG state is covered by checkpoint/replay state.
- No authoritative path reads wall clock after session seed admission.
- Legacy seed vectors and limited-roll mask/modulo cases are exact.
- Same-tick scheduled events have documented order.
- Adding a presentation-only random choice cannot perturb authoritative streams.

## Verification backlog

1. Establish original game-frame cadence for combat, poison, regeneration, AI, missiles, and object events independently.
2. Trace first-trigger/end-tick inclusivity for periodic effects.
3. Trace whether refresh preserves or restarts phase.
4. Map original RNG owners: game, level, room, unit, item, quest, skill.
5. Verify which seed fields are serialized in `.d2s` item data and when they change.
6. Compare a known original item roll sequence against D2MOO/libd2legacy.
7. Capture deterministic skill/missile traces including RNG call order.
8. Verify timer behavior across level transition, death, save/exit, reconnect.
9. Verify same-frame command/event ordering for simultaneous rewards/kills/objects.
10. Add blind holdout traces following libd2's retail-capture methodology.

## Sources

- Dark Magic `internal/game/session/session.go`, `internal/game/simulation`, and `internal/game/ecs`.
- [D2MOO `D2Seed.h`](https://github.com/ThePhrozenKeep/D2MOO/blob/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Common/include/D2Seed.h).
- [D2MOO `D2Seed.cpp`](https://github.com/ThePhrozenKeep/D2MOO/blob/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Common/src/D2Seed.cpp).
- [D2MOO `Items.cpp`](https://github.com/ThePhrozenKeep/D2MOO/blob/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Common/src/Items/Items.cpp).
- [MAP_GENERATION.md](../MAP_GENERATION.md).
- [libd2 verification](https://github.com/jaenster/libd2/blob/e6cdc4927c6180be8dd309b0423b470f64f1fc6c/docs/VERIFICATION.md).
