# Gameplay audio, environmental sound, ambience, music, and cue behavior research

Status: implementation-oriented research baseline. Dark Magic already has a renderer-thread-safe audio mixer, typed `Sounds.txt` records, archive-backed lazy resolution, mixer buses, scoped/persistent playback, fades, groups, streaming, and Lua exposure. The missing work is primarily **semantic cue routing and the large portion of Diablo sound metadata currently parsed but not applied**.

## Executive conclusion

Audio belongs on the presentation side of the authority boundary, but it should be driven by **semantic gameplay/world events and authoritative world snapshots**, not inferred from pixels or used to decide gameplay timing.

Recommended flow:

```text
authoritative simulation/world event
  attack / missile / death / object operation / quest / transition / state
        |
        v
AudioCue semantic record
  sound/sound-set identity
  emitter entity or world position
  event identity/seed
  category
  optional lifetime/tracking context
        |
        v
presentation audio policy
  resolve Sounds.txt / MonSounds / object / level soundscape
  choose deterministic variation
  pitch/volume/delay/priority/instance policy
  positional attenuation/pan/tracking
        |
        v
internal/audio Mixer
        |
        v
platform backend
```

Audio playback handles, decoded buffers, device clocks, and mixer slots are **not authoritative gameplay state**.

## Current Dark Magic audio architecture

`internal/audio` is already a good engine boundary.

The mixer provides:

- generation-checked native sound handles;
- ordered renderer/audio-thread commands;
- buses: `music`, `ui`, `sfx`, `ambience`, `speech`, `cinematic`;
- per-sound volume and stereo pan;
- looping and streaming;
- playback groups and group stop;
- deterministic fade progression through explicit `Advance(delta)` rather than consulting wall clock internally;
- PCM stream support for cinematics/video;
- diagnostics.

The Raylib/platform backend consumes commands without being visible to gameplay packages.

Keep this separation.

## Current Sounds.txt resolution

`internal/audio/catalog.go` already:

- resolves a named `Sounds.txt` record;
- follows `Redirect`;
- performs deterministic weighted grouped-variant selection from `Group Size` / `Group Weight` using a caller-provided seed;
- resolves localized/music/global SFX asset paths;
- chooses a volume inside `Volume Min` / `Volume Max`;
- routes music/UI/ambient/speech-ish rows to suitable mixer buses;
- carries loop/stream state.

This is a strong foundation.

## Parsed Sounds.txt fields not yet represented in playback

Dark Magic's typed `SoundEntry` already contains significantly more metadata than the current `audio.Definition` / `PlayOptions` use.

Fields currently needing explicit semantics include:

- `Pitch Min`, `Pitch Max`;
- `Fade In`, `Fade Out`;
- `Defer Inst`;
- `Stop Inst`;
- `Duration`;
- `Compound`;
- `Falloff`;
- `LFEMix`;
- `3dSpread`;
- `Priority`;
- `Is2D`;
- `Tracking`;
- `Solo`;
- `Music Vol`;
- `Block 1`, `Block 2`, `Block 3`;
- `Delay`;
- `HDOptOut` for later/new-graphics compatibility contexts.

Do not discard these in the normalized sound definition even if the initial backend cannot implement every one.

Suggested normalized definition:

```text
SoundDefinition
  stable ID
  resolved asset
  bus/category
  volume range
  pitch range
  group/weight
  loop/stream
  fade in/out
  start delay
  duration
  duplicate-instance policy
  compound window
  falloff/spread/2D/tracking
  priority
  solo/ducking policy
  music-volume coupling
  environment blocks
```

Unsupported backend fields should remain inspectable and produce diagnostics, not vanish during parsing.

## Gameplay event sounds

Combat/skill/AI docs should emit semantic events at authoritative effect boundaries.

Examples:

```text
AttackAttempted
HitResolved
BlockResolved
DamageApplied
UnitKilled
SkillCastStarted
SkillEffect
MissileSpawned
MissileImpacted
StateApplied
ItemPickedUp
ItemDropped
ItemEquipped
ObjectOperated
QuestAdvanced
ZoneTransitioned
```

Audio maps these to sound cues where authored data or compatibility rules require them.

The sound **must not cause** the attack, missile, damage, or animation event. It reacts to the already-decided semantic event.

## COF/animation sound keyframes

Legacy composite animation research records a COF key event type for sound.

That remains useful presentation semantics, but Dark Magic should distinguish:

- gameplay effect timing, owned by simulation;
- animation event timing, owned by the logical animation player;
- sound cue emission, which may be attached to the animation event for footsteps/swings/vocals where appropriate.

If simulation skips or advances multiple presentation frames, the animation event system must still emit crossed sound events exactly once according to its playback policy.

Headless simulation must never require the sound event to perform gameplay effects.

## Monster sound sets

`MonStats.txt` carries `MonSound` and `UMonSound` references to monster-sound definitions. Unique monsters can therefore use a distinct sound set from their base family.

A normalized monster presentation definition should carry semantic sound roles rather than hard-code file paths.

Potential roles to recover from `MonSounds.txt` include categories such as:

```text
neutral/idle
walk/run/footstep
attack 1/2
hit/get-hit
skill/cast
death
special vocal
```

Exact columns/selection rules need source/data inspection before a compatibility claim.

Monster AI/combat emits events; the monster audio presenter resolves the appropriate sound role and `Sounds.txt` variation.

## NPC speech and dialogue

NPC speech is not the same as ambient SFX.

Use the `speech` bus and preserve:

- dialogue/quest line identity;
- localized audio resolution;
- speaker entity;
- optional subtitles/string identity;
- interruption/stop-instance behavior;
- conversation scope;
- ducking/solo behavior if authored.

Quest/dialogue runtime decides which speech line occurs. Audio does not decide quest state.

## Skills and missiles

Skills/missiles can have several different sound points:

- cast/start;
- action/effect;
- missile loop/travel;
- impact;
- expiration/explosion;
- state aura loop.

Model these as separate semantic cue roles. Do not use one `Skill.Sound` field for every phase.

Missile travel sounds that track a moving entity need emitter tracking until the missile is removed. The current mixer can manage a sound handle, but the presentation layer needs to update pan/attenuation or use future backend-native 3D tracking.

## Items and UI

Separate UI-only audio from world item audio.

Examples:

- panel open/close/click: `ui` bus;
- inventory move/equip: UI or item cue depending on authored behavior;
- world pickup/drop: positional `sfx` cue;
- potion use: gameplay/item effect cue;
- vendor buy/sell: NPC/UI/item cue based on authored behavior.

UI sounds should not enter deterministic gameplay RNG streams.

## Level sound environments

D2MOO's `Levels.txt` structure exposes:

- `Rain`;
- `Mud`;
- `IsInside`;
- `SoundEnv`.

This is strong evidence that level/zone identity carries an authored sound-environment selection and weather/environment facts.

Dark Magic's generated/preset zone definition should retain these values in authoritative/read-only world metadata so presentation can build a soundscape.

Example:

```text
ZoneAudioContext
  level ID
  sound environment ID
  inside/outside
  rain enabled
  environmental/weather flags
  current room/region tags
```

The renderer/audio layer consumes this context. It does not edit it.

## Sound environment resolution

The exact original runtime meaning of `Levels.SoundEnv` and the `Sounds.txt` `Block 1..3` environment/loop fields still needs dedicated client-side/source/owned-game verification.

Do not prematurely invent a `SoundEnv.txt` schema if the mounted data says otherwise.

Research should determine:

- what table/index `SoundEnv` references;
- whether it selects one ambient loop, several blocks, or a scheduling profile;
- how `Block1/2/3` affect loop boundaries or environment playback;
- whether inside/outside alters environment selection;
- transition/crossfade rules;
- random ambient-event timing.

Preserve raw fields until confirmed.

## Ambient scene versus ambient event

`Sounds.txt` already distinguishes `IsAmbientScene` and `IsAmbientEvent`.

Treat these as different jobs:

### Ambient scene

Long-lived zone/room soundscape layers:

- wind;
- dungeon room tone;
- rain;
- fire/river/hell ambience;
- other looping environment beds.

These should be grouped by zone/soundscape and crossfaded when context changes.

### Ambient event

Intermittent one-shots or short events selected by the current sound environment:

- distant creature calls;
- thunder;
- environmental creaks;
- occasional battle/settlement details.

These may use presentation-only deterministic scheduling so they do not perturb gameplay RNG.

## Environmental emitters

Some ambience should be localized to actual world objects/locations rather than a global bed.

Examples:

- fire/campfire;
- waterfall/river;
- portal;
- forge;
- waypoint;
- machinery/traps;
- torches or magical objects if authored.

World-object definitions can expose an optional sound-emitter role:

```text
WorldAudioEmitter
  entity/object ID
  sound role/record
  world position
  loop/one-shot policy
  activation state
```

Object state changes start/stop/replace the sound. Renderer visibility must not be the authority for object state, but presentation culling may suppress inaudible far-away emitters.

## Rain/weather

Rain should be treated as a zone/environment presentation layer derived from authoritative level/weather context.

Potential layers:

- continuous rain bed;
- randomized thunder events;
- indoor attenuation/suppression;
- surface-dependent footstep/wet cues if verified.

The exact Diablo II rain audio policy needs traces. Do not tie the rain loop directly to whether rain particles are currently rendered.

## Footsteps and terrain/material sounds

The current map pipeline knows semantic tiles/subtiles and DT1 material/collision metadata. A future footstep system can derive a **surface material semantic** from the authoritative current location and emit a footstep cue on animation locomotion events.

Do not sample rendered pixels to decide whether a player is walking on dirt or stone.

Exact legacy surface-sound mapping remains a probe. The modern engine can support richer material mappings without changing gameplay.

## Positional audio

World SFX should accept a world emitter position and listener position.

Minimum modern policy:

```text
distance -> attenuation
screen/world relative horizontal offset -> pan
Sounds.Falloff -> authored distance parameter
Is2D -> bypass positional attenuation/pan
Tracking -> update emitter position from entity
```

The current backend exposes stereo pan; richer 3D/spatial output can be added behind `internal/audio` without changing semantic cue sources.

`3dSpread` and LFE metadata should be preserved even if Raylib cannot initially reproduce them exactly.

## Listener ownership

The listener is presentation state, usually the local player's/camera's world anchor.

It is **not** server simulation state and need not be replay-authoritative for gameplay.

For deterministic visual/audio captures, the capture harness can pin listener/camera state.

## Occlusion and acoustic environment

Exact Diablo II likely uses simpler audio than a modern physically based system. Do not require geometric acoustics for compatibility.

However the new engine can later support renderer-independent optional policies such as:

- wall/door occlusion from semantic collision;
- indoor/outdoor low-pass/reverb presets;
- distance zones;
- room/portal acoustic connectivity.

These must remain presentation enhancements unless they affect game mechanics, which they currently should not.

## Deterministic audio variation without gameplay RNG contamination

Grouped sound variant, volume range, pitch range and ambient-event scheduling can be deterministic for captures/replays without consuming simulation RNG.

Derive an audio seed from immutable semantic facts such as:

```text
session presentation seed
semantic event sequence/tick
emitter entity ID
sound role
occurrence index
```

Do not call monster/item/loot RNG just to choose between three attack grunts.

## Duplicate/compound sound policy

`Defer Inst`, `Stop Inst`, `Compound`, `Priority` and `Solo` indicate that original sound playback has explicit concurrency rules.

Dark Magic should add a voice-management layer above mixer slots:

```text
SoundVoicePolicy
  group/instance key
  defer duplicate?
  stop prior duplicate?
  compound window?
  priority
  voice limit
  solo/duck buses?
```

This prevents combat with many monsters from degenerating into uncontrolled hundreds of identical voices.

Exact legacy interpretation of these fields needs probes.

## Music

Music should use persistent grouped streamed playback on the `music` bus.

Music selection should be driven by semantic scene/zone/cinematic state:

- frontend/menu;
- town;
- outdoor/dungeon level;
- boss/special encounter if authored;
- credits/cinematic.

Zone transition should request a music transition/crossfade policy, not simply start a new file every render frame.

The current `play_persistent` group mechanism is already a useful building block.

## Music and ambience transitions

Use an explicit soundscape controller:

```text
current SoundscapeKey
requested SoundscapeKey
active music layer(s)
active ambience layer(s)
transition/fade state
ambient-event scheduler
```

Scene/zone changes update the key. The controller diffs the old/new soundscape and fades/stops/starts layers.

Avoid Lua scenes manually leaking persistent handles across transitions.

## Save/replay/network behavior

Ordinary audio playback state does not belong in durable character saves.

For replay/network:

- simulation replays reproduce semantic events;
- the local audio presenter may replay corresponding cues;
- currently playing loop phase need not be network authoritative unless strict audiovisual capture synchronization requires it;
- persistent zone soundscape is reconstructed from current world/scene snapshot.

NPC dialogue/cinematic synchronization can use explicit media/session timing where needed.

## Audio debugging tools

Extend diagnostics/dev tooling with:

- active voices by bus/group/priority;
- current music and soundscape key;
- current zone `SoundEnv`, inside/rain context;
- emitter positions/radii;
- resolved Sounds.txt record and selected grouped variant;
- source semantic event ID/tick;
- volume/pitch/falloff result;
- duplicate/compound/priority decision;
- ignored unsupported metadata fields.

A developer hotkey overlay could visualize sound emitters/radii similarly to collision/entity-origin diagnostics.

## Suggested implementation slices

### AUD1 — lossless normalized Sounds definition

Carry every relevant typed `Sounds.txt` field into a renderer-independent normalized definition and inspector. Do not change playback behavior yet.

### AUD2 — semantic AudioCue bus

Add a presentation event adapter from simulation/world semantic events to audio cues. Start with UI-independent player attack/hit/death and item pickup/drop cues using synthetic records.

### AUD3 — positional world SFX

Implement emitter/listener distance attenuation, pan, `Is2D`, falloff and tracking above the existing mixer.

### AUD4 — monster sound roles

Normalize MonSound/UMonSound mapping and drive idle/attack/hit/death cues from monster semantic events.

### AUD5 — zone soundscape

Expose level `SoundEnv`, rain and inside/outside in the world snapshot, then add persistent ambient scene/music layers with crossfade.

### AUD6 — ambient events and voice policy

Implement deterministic presentation-only ambient scheduling plus Defer/Stop/Compound/Priority/Solo behavior.

### AUD7 — object/NPC/skill coverage

Add object loops/events, NPC speech, missile travel/impact, skill/aura loops, quest and transition cues.

## Verification backlog

1. Exact `Levels.SoundEnv` target/index semantics.
2. `Sounds.txt` Block1/2/3 meaning and units.
3. Pitch min/max units and randomization.
4. Fade In/Out units and behavior.
5. Delay and Duration units relative to game/audio ticks.
6. Defer Inst and Stop Inst exact duplicate-key semantics.
7. Compound field semantics/window and volume behavior.
8. Priority and global/per-channel voice limits.
9. Falloff units/curve and `Is2D` interaction.
10. `3dSpread`, LFE mix and tracking behavior.
11. Solo and MusicVol ducking/routing behavior.
12. AmbientScene versus AmbientEvent scheduling.
13. Zone music selection and transition/crossfade rules.
14. Rain/inside/outside soundscape changes.
15. MonSounds columns, variants and event-role mapping.
16. Unique monster `UMonSound` override behavior.
17. Object sound fields and loop/start/operate/destroy rules.
18. Skill/missile sound source fields and action timing.
19. COF sound keyframe mapping and duplicate interaction with skill cues.
20. Player footsteps, weapon swings, hit/block and surface-material cues.
21. NPC speech interruption, subtitles and localized file resolution.
22. Audio behavior when an emitter enters/leaves active room or presentation residency.
23. Multiplayer remote-unit cue replication versus locally derivable cues.
24. Audio replay/capture deterministic seed derivation.

## Primary sources inspected

- Current Dark Magic `internal/audio/audio.go`, `catalog.go`, Lua audio module, Raylib backend architecture, and typed `SoundEntry`.
- Current Dark Magic `MonsterStats.MonSound/UMonSound` and gameplay event architecture.
- D2MOO pinned 1.10f `Levels.txt` structure including `Rain`, `Mud`, `IsInside`, and `SoundEnv`.
- Legacy COF research sound key events and existing Dark Magic map/world semantic state.

D2MOO is primarily server/common reconstruction and therefore cannot by itself answer every client audio behavior. Client-side/open-engine corroboration and owned-game capture are explicitly required for the unresolved sound-environment and voice-management rules.
