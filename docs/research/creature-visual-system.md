# Creature visual-system brainstorming ledger

Status: **brainstorming and research input, not approved architecture**.

This document preserves design ideas for long-term creature extensibility,
3D-authored assets, skeletal animation, generated sprites, compositional entity
rendering, and developer tooling. It is intentionally broader and more
speculative than `docs/CREATURE_ASSET_ROADMAP.md`.

The ideas here should be reviewed before they become implementation contracts.
When a future roadmap pass evaluates one of them, classify it as:

```text
ACCEPT
PROTOTYPE
DEFER
REJECT
NEEDS MORE RESEARCH
```

For accepted or prototyped ideas, record the problem being solved, alternatives,
ownership, dependencies, test strategy, compatibility consequences, and migration
cost. Do not let this ledger silently become architecture by citation alone.

## Relationship to the current roadmap

`docs/CREATURE_ASSET_ROADMAP.md` preserves the proposed M31-M43 creature-asset
program. GitHub issues, milestones, and the Roadmap project determine which
slices are current. Many ideas below are already partially represented there: canonical
schemas, glTF import, Creature Lab, arbitrary components and sockets, animation
events, Blender authoring, palette-aware materials, sprite generation, modern
sprite composition, legacy export, retargeting, build caching, and measured
runtime optimization.

This ledger exists for ideas that are either broader than those milestones,
insufficiently proven, or worth keeping visible as future research directions.
The companion `docs/CREATURE_ASSET_RESEARCH_BACKLOG.md` turns these ideas into
candidate research checkpoints without promoting them to scheduled GitHub
issues and milestones yet.

## Historical first-party source

Review the previous first-party Blender addon:

<https://gitlab.com/gravestench/blender-addon-dc6-export>

The addon was built for an older Blender generation and should be treated as a
historical prototype, not as a current implementation baseline. Blender Python,
scene, render, registration, property, and dependency APIs may have changed.
Direct code reuse should be considered independently from conceptual reuse.

Useful areas to inspect include:

- camera-rig construction and directional views;
- orthographic/isometric presentation assumptions;
- frame and animation sampling;
- batch rendering;
- transparency and cropping;
- palette/index conversion;
- DC6 export flow;
- Blender UI decisions;
- scene-state assumptions that should become explicit profiles;
- responsibilities that should move to shared headless tooling or codec repos.

For each observed feature, classify it as `reuse conceptually`, `redesign`,
`obsolete`, or `requires validation`.

## Existing legacy compatibility research

Use the existing DCC, DC6, COF, PL2, composite animation, and paletted-rendering
documentation as compatibility input. Preserve useful semantics such as shared
component clocks, common origins, semantic animation events, equipment-driven
component replacement, palette transforms, direction mapping, and ordering.
Do not preserve fixed historical component sets, fixed direction counts, or
binary-format constraints as requirements for new assets.

## Generalized visual actor abstraction

One of the strongest ideas to review is a generalized visual-animation model
above both modern rigged rendering and legacy sprite composition. Names are not
settled; a possible vocabulary is:

```text
VisualActor
VisualClip
VisualTrack
VisualComponent
VisualEvent
```

Conceptually:

```text
VisualActor
  |
  +-- logical bounds / footprint
  +-- skeleton or pose source
  +-- component slots
  +-- material parameters
  +-- semantic sockets
  |
  +-- VisualClip: attack
        |
        +-- RigTrack      -> body.glb
        +-- RigTrack      -> sword.glb
        +-- SpriteTrack   -> legacy aura.dcc
        +-- ParticleTrack -> fire
        +-- EventTrack    -> attack.hit @ 0.42
```

The important question is whether a track-oriented semantic model can unify:

- rigged meshes;
- generated sprite tracks;
- decoded DCC/DC6 tracks;
- particle/effect tracks;
- sound/event timelines;
- attachments and equipment;
- procedural visual layers.

Do not adopt the names or API shape without a prototype. The architectural goal
is to prevent source formats and renderer backends from becoming gameplay
semantics.

## Legacy adapters into the generalized model

A possible compatibility direction is:

```text
COF + DCC/DC6
      |
      v
legacy adapter
      |
      v
VisualActor / VisualClip / VisualTrack
```

Potential mappings include COF component ordering to track ordering, COF timing
to clip timing, COF keyframes to semantic events, and DCC/DC6 resources to sprite
tracks. Compatibility-specific weapon-class and component suppression rules
should remain in the adapter/resolver rather than defining the general model.

The reverse direction may also be useful where representable:

```text
modern creature
      |
      v
pose + direction sampler
      |
      v
component sprite renderer
      |
      v
palette/index conversion
      |
      +--> modern sprite representation
      +--> DC6
      +--> DCC
      +--> COF-compatible metadata
```

The mapping should emit explicit loss/unsupported-feature diagnostics.

## Mixed 2D and 3D composition

Do not assume an entity must be entirely sprite-based or entirely 3D. A future
compositor might support combinations such as:

```text
3D rigged body
+ 3D equipment
+ legacy sprite aura
+ modern particle effect
+ procedural light
```

or:

```text
legacy sprite body
+ modern particles
+ generated equipment overlay
```

This is speculative. The near-term requirement is only to avoid architecture
that makes mixed representation impossible. A future prototype should establish
whether mixed tracks are actually useful enough to justify the complexity.

## Canonical creature semantics

A Dark Magic-native creature definition should remain independent from Blender,
glTF, COF, DCC/DC6, Raylib, or any future graphics API. A possible shape is:

```text
CreatureDefinition
  +-- SkeletonDefinition
  +-- ModelSet
  +-- AnimationLibrary
  +-- ComponentManifest
  +-- MaterialDefinitions
  +-- SocketDefinitions
  +-- CollisionDefinition
  +-- SelectionDefinition
  +-- VisualDefinition
```

Potential semantic fields include identity, bind/rest pose, models, skinning,
components, variants, material regions, animation clips, events, states, facing,
sockets, equipment mappings, shadow metadata, logical bounds, collision, and
selection geometry.

Version every native schema explicitly and never silently reinterpret an older
schema revision.

## glTF and authoring boundaries

Continue evaluating glTF/GLB as the preferred standard interchange for meshes,
skinning, skeletons, animations, materials, textures, and morph targets. Do not
make glTF the runtime actor model.

Dark Magic-specific metadata may live in glTF extras, a validated extension, or
companion manifests. Prototype this choice. Candidate semantics that may not fit
cleanly in portable glTF include semantic sockets, gameplay animation events,
palette regions, logical animation states, collision/selection, render profiles,
and mod/content identifiers.

## Blender frontend principles

A future Blender addon should be an artist-facing frontend over shared contracts,
not the only implementation of the asset pipeline.

Useful authoring capabilities include:

- rig and skin-weight validation;
- semantic bone/socket mapping;
- component assignment;
- material/palette region authoring;
- animation clip and loop definitions;
- typed event markers;
- collision/selection authoring;
- root-motion policy;
- render/camera profile selection;
- directional sprite preview;
- package export;
- actionable validation diagnostics.

Every concept authored in Blender should also have a tool-independent serialized
representation. `.blend` files are source assets, not runtime contracts.

## Semantic bones and sockets

Separate semantic roles from actual rig bone names. Examples:

```text
body.root
body.center
head
hand.left
hand.right
weapon.left
weapon.right
shield
projectile.origin
spell.origin
foot.left
foot.right
```

One rig may map `weapon.right` to `hand_r`; a non-humanoid may map it to
`claw_primary`. This idea can support attachments, generic gameplay hooks, and
retargeting without requiring a universal skeleton.

## Retargeting and rig diversity

Keep the architecture valid for humanoids, quadrupeds, insects, serpents,
floating creatures, winged creatures, multi-limbed creatures, and creatures
with optional/detachable parts.

Potential retarget research includes reference rigs, semantic mappings,
rest-pose conversion, proportion handling, missing bones, masks, additive
animation, root policy, and diagnostics explaining why a source/target pair is
incompatible.

## Time-based animation semantics

Modern animation clips should probably be time-based rather than defined by
sprite frame numbers. Candidate semantics include duration, normalized time,
loop policy, playback speed, interpolation, events, masks, additive layers, root
motion, and synchronization groups.

A sprite renderer/exporter can sample this timeline into discrete frames.

Animation events should be typed and interval-safe. Examples:

```text
attack.hit
projectile.spawn
sound.play
effect.spawn
footstep
ability.trigger
```

An event at 0.42 must still fire if update time advances directly from 0.40 to
0.47. Event delivery also needs defined loop-wrap and playback-rate behavior.

## Shared primary clock and secondary layers

Composed body, weapon, shield, and armor participating in the same action should
usually derive from one primary clip/time/facing state. Equipment or appearance
changes should preserve normalized animation progress unless an explicit
transition requires otherwise.

Secondary animation layers may eventually own separate clocks for cloth, wings,
tails, auras, facial animation, or procedural effects. Do not let this weaken the
primary synchronization contract.

## Continuous facing with sampled directions

Logical actor orientation should be continuous. Direction counts such as 8, 16,
32, or 64 belong to render/export profiles. One source animation library should
be able to drive live 3D, multiple sprite-direction profiles, portrait renders,
thumbnails, and legacy output.

## Component and appearance resolution

Prefer arbitrary named components over a fixed Diablo component list. Candidate
components include body regions, equipment, wings, horns, tail, cape, back
attachments, and effects.

Appearance should be resolved from data:

```text
CreatureDefinition
+ Variant
+ Equipment
+ Material Parameters
+ Animation / Facing
      |
      v
Visual Resolver
      |
      v
ResolvedVisualState
      |
      v
Renderer or Sprite Baker
```

This resolver can generalize legacy token/mode/weapon-class/component selection
without putting those rules into the renderer.

## Material and palette architecture

Review preserving indexed texture information as a first-class GPU capability:

```text
IndexedTexture
Palette
PaletteTransform
```

A shader can map indices through a palette and optional transform table instead
of eagerly expanding every legacy texture to immutable RGBA.

For 3D-authored assets, investigate palette-aware material regions such as skin,
primary armor, secondary armor, cloth, metal, weapon, and monster variant.
Potential representations include indexed source textures, masks, region-ID
textures, grayscale masks, lookup textures, or PL2-like transforms.

This area requires prototypes before settling the representation.

## Blender preview of game palette semantics

A useful authoring feature would preview the shared Dark Magic interpretation of
palette mapping, recoloring, transparency, lighting ramps, and quantization in
Blender. Avoid a Blender-only approximation that silently diverges from runtime
or headless generation semantics.

## 3D-to-sprite generation

A major long-term pipeline idea is:

```text
model + rig + animation + appearance
+ direction + sample time
+ camera/light/render profile
      |
      v
configured renderer
      |
      v
RGBA frame
      |
      v
palette quantizer
      |
      v
indexed SpriteFrame
```

Support both flattened creature output and independently generated component
layers. Component layers must share exact pose, animation time, camera,
projection, origin, scale, lighting, and render resolution.

Tight cropping is acceptable only if metadata retains the common actor origin and
reconstructing offsets exactly.

## Explicit render, lighting, and sampling profiles

Do not use undocumented Blender scene state as output authority. Reusable profiles
should define camera/projection, actor origin, lighting, direction count, sample
times/FPS, terminal-frame behavior, output resolution, palette, quantization,
shadow policy, and component policy.

One source clip should therefore be able to target profiles such as:

```text
d2_creature_8dir
d2_creature_16dir
modern_sprite_16dir
portrait
thumbnail
```

## Palette quantization as an independent stage

Keep RGBA rendering separate from palette quantization. Quantization can then
have explicit policy for palette selection, transparency index, reserved indices,
nearest-color mapping, dithering, and masks. This makes quantization independently
testable and replaceable.

## Generated assets are build products

Treat models, rigs, animations, materials, textures, manifests, and generation
profiles as valuable source inputs. DCC/DC6, atlases, indexed sprites, and
composite sprite packages should be reproducible build products where practical.

This allows renderer, lighting, crop, quantization, and encoder improvements to
regenerate artifacts without destroying source semantics.

## Common compositor semantics

Generated sprite frames and decoded legacy sprite frames should ideally enter the
same modern sprite-track/compositor abstraction:

```text
legacy DCC      -> SpriteTrack
rendered sprite -> SpriteTrack
```

The storage backend may differ; compositor semantics should not depend on how a
frame was produced.

## Modern sprite-composite model

A modern semantic successor to COF may eventually describe arbitrary components,
directions, timing, common origins, ordering, transforms, palette parameters,
shadows, events, sockets, and bounds. Do not invent a binary format before the
semantic model and profiling justify it.

Potential storage backends include atlases, texture arrays, indexed texture
arrays, per-animation packing, or per-component packing. Storage should remain an
adapter decision.

## Creature Lab / Visual Actor Lab

A dedicated developer environment remains one of the highest-value ideas.
Potential capabilities include:

- creature/variant/clip selection;
- play, pause, scrub, frame step, event step, speed control;
- continuous 3D rotation and discrete sprite direction selection;
- skeleton, bones, bone names, sockets, attachments, collision, selection, and
  bounds overlays;
- component visibility and equipment swapping;
- material and palette parameter editing;
- current time/frame/direction/event inspection;
- resolved logical asset IDs and generated artifact paths;
- 3D vs generated sprite vs legacy sprite comparison;
- per-component vs flattened rendering comparison;
- regeneration and hot reload.

Synchronized comparison and visual-diff modes are especially useful for validating
legacy compatibility and generated output.

## Development-time render-on-demand

A useful development workflow is:

```text
edit model/animation
      |
      v
invalidate dependency node
      |
      v
request current representation
      |
      v
render only affected samples/components
      |
      v
reload in Creature Lab
```

This is a good development target. Production runtime generation is a separate,
later question and should not be implied by implementing development-time
regeneration.

## Asset generation service/subsystem

Another idea worth preserving is a reusable generation subsystem that can power
the CLI, Creature Lab, Blender addon, CI, and batch tools:

```text
request(actor, clip, component, direction, sample, profile)
      |
      v
asset generation subsystem
      +-- dependency lookup
      +-- cache lookup
      +-- render
      +-- quantize
      +-- encode/export
      |
      v
artifact
```

Whether this should literally be a process/service is undecided. Prefer a local
library/subsystem first unless process isolation or distributed work provides a
measured benefit.

## Dependency graph and content-addressed generation

Model dependencies explicitly so changes invalidate only affected products.
Potential key inputs include source model, skeleton, clip, component, material,
texture, render profile, light profile, palette, renderer version, processor
version, and tool version.

Desired behavior examples:

- changing a helmet rebuilds helmet-dependent outputs only;
- changing one clip rebuilds samples for that clip only;
- changing only the palette can re-quantize without rerendering geometry;
- changing lighting rerenders affected frames;
- a failed rebuild does not destroy the currently usable live artifact.

Parallel generation by creature, animation, direction, or component is a later
optimization and must preserve deterministic output.

## Plugin/processor extensibility

Potential extension seams include:

```text
Importer
Validator
Processor
AnimationProcessor
Renderer
Quantizer
Encoder
Exporter
Inspector
```

Possible implementations may include glTF and legacy importers, Blender or engine
sprite renderers, palette quantizers, codec exporters, and modern composite
exporters. Do not generalize an interface merely because only one implementation
exists; prove at least a second real implementation or adapter need first.

## Renderer backend separation

Creature semantics should remain independent from Raylib or any future backend.
Possible consumers include a Raylib 3D adapter, future graphics backends,
software/reference rendering, sprite composition, legacy sprite rendering, and
offline Blender rendering.

The Blender-renderer-vs-engine-renderer decision for sprite generation remains a
prototype question. Compare artist parity, headless CI, determinism, deployment,
maintenance, and performance rather than assuming one answer.

## Resource/state separation

Keep immutable/shared definitions distinct from mutable instance state.

Shared examples:

```text
SkeletonDefinition
AnimationClip
Mesh
Material
CreatureDefinition
VisualClipDefinition
ComponentDefinition
```

Instance examples:

```text
SkeletonPose
AnimationPlayer
CreatureInstance
VisualActorInstance
ComponentState
MaterialInstance
AttachmentState
```

This supports sharing, caching, instancing, and equipment swaps without corrupting
other instances.

## Collision, selection, and visual bounds

Keep separate concepts for frame bounds, actor logical bounds, selection bounds,
and gameplay collision/footprint. Do not derive gameplay collision from animated
render meshes.

## GPU skinning, instancing, LOD, and procedural animation

Preserve extension points for CPU reference skinning, GPU skinning, instanced
skinning, alternate render representations, LOD/impostors, IK, foot placement,
aim/look offsets, recoil, secondary motion, facial/morph layers, and cloth.
These should be driven by measured future needs, not pulled into the first asset
pipeline implementation.

## Mod/content extensibility

Prefer stable logical asset IDs and layered resolution so a mod can replace or
extend individual models, materials, textures, clips, components, sockets,
variants, or palette behavior without copying an entire creature package.

Conceptual IDs might resemble:

```text
creature:demon
skeleton:demon/base
animation:demon/attack
component:demon/horns
material:demon/skin
```

The exact naming grammar remains a content-system decision.

## Synthetic fixtures and tests

Normal CI should use redistributable synthetic creatures, including at least a
simple humanoid, an equipment-heavy humanoid, a structurally distinct
non-humanoid, a palette-region fixture, and an exaggerated socket fixture.

Use golden poses at fixed timestamps in addition to visual snapshots. Sprite
regressions should assert dimensions, bounds, offsets, direction/sample counts,
event mapping, palette indices, component alignment, and structural hashes where
appropriate.

Optional original Diablo II compatibility testing remains user-supplied and must
never commit proprietary assets.

## Codec repository ownership

Binary-format mechanics belong in the independent codec repositories. Candidate
future work includes strict validators, deterministic encoders/marshal APIs,
indexed-frame interfaces, COF construction, palette helpers, round-trip tests,
and inspectors.

Dark Magic should own semantic creature definitions, resolution, sampling,
render/generation orchestration, loss reports, and presentation adapters.

## Ideas most worth preserving for later review

The following ideas currently look especially promising, while still remaining
subject to review:

1. A generalized visual actor/clip/track abstraction above legacy and modern
   rendering.
2. Legacy COF/DCC/DC6 as adapters rather than canonical creature semantics.
3. Shared primary animation clocks with explicit secondary layers.
4. Semantic bones and sockets.
5. Arbitrary components and appearance resolution.
6. Equipment/component replacement that preserves animation progress.
7. Continuous orientation with discrete directions as render/export profiles.
8. Blender as an authoring frontend over shared headless contracts.
9. A dedicated Creature Lab.
10. Palette-aware 3D materials and indexed sprite rendering.
11. Generated sprites and legacy assets as reproducible build products.
12. Common compositor semantics for decoded and generated sprite tracks.
13. Dependency-aware/content-addressed generation and hot reload.
14. Explicit codec ownership boundaries.
15. Development-time render-on-demand.
16. Mixed 2D/3D composition as a future compatibility/extensibility experiment.

## Ideas that require prototypes before adoption

At minimum, prototype or research:

- the exact generalized visual actor/track API;
- mixed 2D/3D track usefulness and renderer complexity;
- glTF extras/extensions versus companion manifests;
- Blender rendering versus Dark Magic/offline rendering;
- exact D2-compatible camera/light/sampling profiles;
- palette-aware 3D material representation;
- indexed GPU texture/LUT strategy;
- modern sprite atlas/storage layout;
- retargeting implementation/library strategy;
- CPU/GPU skinning transition;
- content-addressed cache semantics;
- development-time versus production on-demand generation;
- plugin boundaries and what actually needs extension interfaces.

## Review rule

When future work reaches this area, do not simply start implementing this file.
Review it alongside the current repository and `docs/CREATURE_ASSET_ROADMAP.md`,
then promote only accepted/prototyped items into concrete milestones with testable
acceptance criteria. Preserve rejected and deferred decisions in the research
record so later work understands why they were not adopted.
