# Creature asset, animation, and generation roadmap

Status: planning and architecture only. This document defines the M31-M43
program in `ROADMAP.md`; it does not authorize implementation of a particular
game or mod. Normal CI uses redistributable synthetic assets. Optional legacy
compatibility probes require legally supplied local data.

## Outcome and dependency graph

One logical creature can target live 3D, flattened sprites, independently
composited sprites, legacy-compatible output where representable, and developer
inspection without making Blender, glTF, COF, DC6/DCC, Raylib, or a graphics API
the creature abstraction.

```text
M31 schemas/evidence
  |
  +--> M32 glTF import --> M33 creature lab --> M34 composition --> M35 animation
  |                           |                    |                 |
  |                           +--------------------+-----------------+
  |                                                |
  +--> M36 Blender frontend <----------------------+
  |                                                |
  +--> M37 materials/palettes <--------------------+
                                                   |
                                                   v
                              M38 sprite generation --> M39 modern composite
                                         |                    |
                                         +--> M40 legacy export
                                         |
                              M41 retargeting (after M35)
                                         |
                              M42 build/cache/live reload
                                         |
                              M43 measured advanced runtime
```

M31-M38 form the near-term proof. M39-M43 must not delay that observable slice.

## Architectural ownership

| Concern | Owner |
| --- | --- |
| Creature, skeleton, animation, component, socket, and event semantics | Dark Magic engine/domain packages |
| Import, validation, build graph, rendering, quantization, packaging | Dark Magic asset tooling |
| Artist UI, scene annotations, previews, export orchestration | Replaceable authoring frontends (initially Blender and evaluated Dust3D workflows) |
| DCC/DC6/COF/PL2 binary mechanics | Independent codec repositories |
| Skinning, drawing, shaders, texture storage | Renderer backend adapters |
| Logical IDs, variants, overrides, package layering | Mod/content layer |

Likely package boundaries, to be admitted by M31 architecture tests, are
`internal/creature/asset`, `internal/creature/animation`,
`internal/creature/runtime`, `internal/assets/gltf`,
`internal/assets/build`, `internal/assets/generate`, and
`internal/dev/tools/dm_asset`. Existing `internal/assets/decode` remains the
legacy decoding boundary; `internal/presentation/render` and
`internal/platform/raylib/renderer` are adapters, not semantic owners.

## Authoring and runtime diagrams

```text
          Blender or Dust3D
                    |
                 glTF/GLB + companion manifests
                    |
                    v
           Dark Magic import adapters
                    |
                    v
       Canonical semantic asset representation
             +------+-------+
             |              |
             v              v
        3D runtime      Asset generator
                         +---+---+
                         |       |
                         v       v
                    modern     legacy
                    sprites   DCC/DC6/COF
```

```text
Blender addon
  +-- authoring UI, annotations, validation, preview
  +-- rig/component/socket/event/package orchestration
  |
  v
Shared Dark Magic contracts
  +-- validator + compiler + renderer + quantizer + exporters
```

```text
CreatureDefinition
  +-- SkeletonDefinition + AnimationLibrary + Components
  +-- Materials + Sockets + collision/selection/render metadata
  |
  v
CreatureInstance
  +-- SkeletonPose + AnimationPlayer + ComponentState
  +-- MaterialState + AttachmentState
  |
  v
Renderer adapter
```

```text
Skeleton + Animation + Components
              |
          Pose sampler
              |
   Direction/camera/light profile
              |
       Component renderer
              |
          RGBA frames
              |
      Palette quantizer
          +---+---+
          |       |
          v       v
       modern   legacy encoder
       sprites   DCC/DC6
```

The build graph records models, skeletons, clips, materials, textures, render
profiles, palettes, tool versions, and processor versions as independent inputs.
Runtime 3D and sprite outputs are derived nodes; legacy output derives from the
sprite/quantization nodes rather than directly from Blender scene state.

## M31: Evidence, boundaries, and executable schemas

- **Goal / reason:** turn existing composite/palette research and the historical
  addon into versioned, inspectable semantics before implementations harden
  accidental format or renderer assumptions.
- **Prerequisites:** M3, M12, `docs/COMPOSITE_ANIMATION.md`,
  `docs/PALETTED_RENDERING.md`, and `docs/formats/{COF,DCC,DC6}.md`.
- **Decisions:** logical IDs resolve through the layered VFS; immutable resources
  are separate from instance state; facing is continuous and adapter-quantized;
  clocks are time-based; collision and selection are authored separately.
- **Packages / modules / tools:** introduce semantic packages above and
  `dm-asset inspect|validate --json`; no Blender code yet.
- **Schemas / assets:** draft explicitly versioned `darkmagic.creature/v1`,
  `skeleton/v1`, `animation/v1`, `material/v1`, `composite/v1`, and
  `render-profile/v1`. Create minimal humanoid, equipment, non-humanoid,
  palette-region, and socket fixtures with stable licenses/provenance.
- **Tests / acceptance / visible result:** schema round trips, unsupported-version
  rejection, exact field-path diagnostics, and cross-reference validation. The
  CLI inspects all five fixture shapes as text or stable JSON and distinguishes
  missing resources, invalid semantics, and unsupported versions.
- **Blender / codecs:** document addon comparison only; codecs receive separately
  tracked API gaps. No encoding.
- **Risks / deferred:** extras versus companion manifests remains a prototype;
  compiled binary schemas and runtime rendering are deferred.

## M32: glTF import and golden poses

- **Goal / reason:** prove standard interchange for mesh, skin, hierarchy, clips,
  textures, morph-target preservation, and metadata escape hatches without
  turning glTF into `CreatureDefinition`.
- **Prerequisites:** M31 schemas and fixtures.
- **Decisions:** GLB/glTF is a source/import representation; unsupported or lossy
  features are explicit diagnostics; Dark Magic-only semantics use the M31
  metadata prototype result.
- **Packages / tools / interfaces:** `internal/assets/gltf`; extend `dm-asset
  inspect|validate|build`; define importer and normalized import-result contracts.
- **Assets:** one redistributable humanoid with idle/walk/attack and one
  non-humanoid with a materially different hierarchy. Prototype at least one
  Dust3D-authored GLB so its generated topology, skin weights, bone naming,
  procedural clips, materials, axes, and units are measured by the same importer
  contract rather than accepted by visual inspection alone.
- **Tests / acceptance / visible result:** assert hierarchy, bind pose, weights,
  clip durations, material references, bounds, and selected bone transforms at
  fixed timestamps. Both GLBs import headlessly with deterministic structural
  JSON; malformed weights, cycles, duplicate bones, and missing targets identify
  the exact node/accessor.
- **Blender / codecs:** validate export from supported Blender versions in a
  non-blocking matrix; no codec work.
- **Risks / deferred:** CPU/GPU skinning, advanced materials, retargeting, and
  runtime graphs are deferred.

## M33: Creature laboratory and reference 3D playback

- **Goal / reason:** make animation visually debuggable early rather than hiding
  uncertainty behind a compiler.
- **Prerequisites:** M32.
- **Decisions:** evaluation and CPU reference pose/skinning are backend-neutral;
  Raylib is the first adapter, not an asset dependency.
- **Packages / tools / interfaces:** `internal/creature/runtime`, a
  renderer-neutral pose/debug-draw snapshot, and `creature-lab` developer mode.
- **Assets:** M32 fixtures plus visible axis/bone/bounds markers.
- **Tests / acceptance / visible result:** play, pause, scrub, speed, loop, facing,
  free rotation, clip selection, bone names, skeleton, bounds, and current time.
  Headless tests assert sampled poses and resource teardown; the lab plays all
  clips without proprietary assets or mutable state in shared resources.
- **Blender / codecs:** none.
- **Risks / deferred:** GPU skinning, production world replacement, components,
  palette shaders, IK, and polished UX.

## M34: Components, sockets, variants, and equipment

- **Goal / reason:** prove arbitrary replaceable body/equipment composition on a
  shared skeleton and clock, including non-humanoid cases.
- **Prerequisites:** M33.
- **Decisions:** semantic sockets map to rig bones/authored transforms;
  components are arbitrary named resources, not the COF set; variants are
  overlays, not copied definitions; visual changes preserve normalized time.
- **Packages / interfaces:** component resolver, `CreatureFactory`, checked
  attachments, immutable definitions and mutable component/material/attachment
  state. Extend lab component, equipment, socket, and visibility controls.
- **Assets:** equipment humanoid and socket/non-humanoid fixtures.
- **Tests / acceptance / visible result:** swap weapon, helmet, armor, and a
  non-humanoid component at 42% of attack without recreating the player or
  changing normalized time; socket transforms and shared-pose component meshes
  match golden transforms. Missing/incompatible bindings are actionable.
- **Blender / codecs:** metadata authoring design only; no codec work.
- **Risks / deferred:** per-frame legacy ordering, cloth, physics, and gameplay
  inventory coupling.

## M35: Semantic animation and events

- **Goal / reason:** generalize the useful COF shared-clock/event behavior into a
  time-based animation layer independent of sprite frames and named game states.
- **Prerequisites:** M34.
- **Decisions:** arbitrary typed events; interval-crossing delivery including
  loop wrap and playback-rate changes; data-defined states/transitions; one
  primary clock with optional secondary layers; gameplay movement is
  authoritative and root motion is explicit/opt-in.
- **Packages / schemas / tools:** animation player, event cursor, facing and
  sampling contracts; extend `animation/v1` and lab timeline/event navigation.
- **Assets:** clips with boundary, same-time, loop-wrap, socket, additive, and
  mask examples.
- **Tests / acceptance / visible result:** an event at 0.42 fires exactly once
  across a 0.40-to-0.47 step, under rate changes and loop wrap; timeline displays
  and jumps among typed events. State names are fixture data, not engine enums.
- **Blender / codecs:** map COF event semantics in compatibility tests only.
- **Risks / deferred:** blend trees, network gameplay reactions, IK, advanced
  graphs, and implicit root motion.

## M36: Modern Blender addon

- **Goal / reason:** give artists first-class rig, semantic-bone, socket,
  component, material-region, clip, event, collision/selection, camera, preview,
  validation, and package-export UI without forking pipeline logic.
- **Prerequisites:** M31-M35 stable contracts.
- **Decisions:** publish minimum/tested Blender versions plus addon/API schema
  versions; addon orchestrates shared CLI/API; `.blend` is source, never runtime.
- **Modules / tools:** separately packaged addon and headless Blender integration
  runner; `dm-asset` remains the validator/compiler authority.
- **Assets / tests / acceptance:** export M32/M34 fixtures in a pinned Blender
  matrix, then validate hierarchy, clips, components, sockets, events, bounds,
  and deterministic semantic metadata through engine import. Artist diagnostics
  select the offending Blender object/bone/field.
- **Codec work:** none; export buttons may call later shared exporters.
- **Risks / deferred:** renderer parity and multiple Blender versions are tested,
  not assumed; no embedded Pillow palettes or private DC6 encoder. Dust3D remains
  a separate optional authoring frontend and is not coupled to this addon.

## M37: Materials and palette parameters

- **Goal / reason:** support reusable texture/material assets plus instance-local
  skin, armor, cloth, metal, weapon, variant, status, alpha, and emissive changes.
- **Prerequisites:** M34 and a prototype comparing indexed textures/LUTs with
  grayscale or region-ID masks and existing PL2 behavior.
- **Decisions:** material semantics are independent of Blender nodes and GPU
  handles; transforms never mutate shared textures; backend support is explicit.
- **Packages / schemas / tools:** material resources/instances, palette adapter,
  lab controls, and backend-neutral capture fixtures.
- **Assets / tests / acceptance:** palette-region fixture shows independently
  adjustable regions on two instances sharing resources; headless and Raylib
  captures agree on declared transform semantics and preserve transparency.
- **Blender / codecs:** addon authors regions and previews; `pl2` owns binary
  mechanics and reusable palette helpers only.
- **Risks / deferred:** the prototype chooses the first technique; universal PBR
  parity, indexed-only runtime storage, and premature shader optimization defer.

## M38: Deterministic directional sprite generation

- **Goal / reason:** render any supported creature pose to flattened or separate
  component sprites with reproducible alignment and metadata.
- **Prerequisites:** M33-M37.
- **Decisions:** explicit camera, lighting, sampling, direction, and output
  profiles are authoritative; all components share pose/camera/light/origin;
  tight crops retain reconstructing offsets; output records tool versions.
- **Packages / schemas / tools:** pose sampler and renderer interface under
  `internal/assets/generate`; `dm-asset render-sprites`; lab generation/compare.
- **Assets / tests / acceptance:** render the near-term proof to eight directions.
  Structural tests assert timestamps, dimensions, bounds, offsets, hashes,
  alignment, terminal-frame policy, and event-to-sample mapping. Repeated pinned
  builds are byte-identical or satisfy a documented structural equivalence.
- **Blender / codecs:** prototype Blender versus Dark Magic renderer (and hybrid)
  for artist parity, headless CI, determinism, deployment, and maintenance;
  neither is selected before evidence. No legacy encoding.
- **Risks / deferred:** cross-GPU bit identity may be unrealistic; quantify the
  equivalence contract. Quantization, atlas packing, and parallelism defer.

## M39: Modern sprite-composite runtime

- **Goal / reason:** provide a semantic, human-readable successor to COF runtime
  composition without inheriting fixed components/directions/frame clocks.
- **Prerequisites:** M38.
- **Decisions:** storage strategy is adapter-selected; definition covers arbitrary
  components, timing, directions, common origins, ordering, transforms, palette,
  events, shadows, sockets, and bounds. Binary compilation needs profile evidence.
- **Packages / schemas / tools:** `composite/v1`, sprite-composite renderer, lab
  synchronized 3D/sprite/layers comparison and optional visual diff.
- **Assets / tests / acceptance:** equipment swap at 42% remains synchronized;
  layers reconstruct the flattened reference at every sampled direction/time;
  modern source and sprite timelines report identical semantic events.
- **Blender / codecs:** addon previews through shared output; no codec changes.
- **Risks / deferred:** atlases/arrays/compression and legacy loss mapping defer.

## M40: Quantization and legacy exporters

- **Goal / reason:** make RGBA-to-indexed conversion and representable
  DCC/DC6/COF generation explicit adapters, never canonical creature semantics.
- **Prerequisites:** M38-M39 and required tagged codec encoder APIs.
- **Decisions:** configurable palette, transparency index, reserved indices,
  mapping, dithering, and masks; modern-to-legacy conversion emits a machine-
  readable loss/unsupported-feature report.
- **Packages / tools:** `dm-asset quantize|export-dc6|export-dcc|export-cof`;
  exporter interfaces consume stable intermediate frames/composites.
- **Assets / tests / acceptance:** synthetic round trips verify directions,
  frames, offsets/origins, palette indices, transparency, ordering, and events.
  Optional owned-asset checks assert structure only. Blender invokes the same
  tooling and contains no encoder.
- **Codec work:** deterministic marshal/encode, indexed-frame APIs, COF
  construction/validation, and palette helpers belong in `dc6`, `dcc`, `cof`,
  and `pl2`; update `docs/formats/CODEC_FOLLOWUPS.md` before implementation.
- **Risks / deferred:** DCC encoding complexity and COF lossiness may split this
  milestone into one PR per format. Modern-only features are never silently lost.

## M41: Skeleton mapping and retargeting

- **Goal / reason:** reuse animation across compatible rigs without a universal
  skeleton or humanoid assumption.
- **Prerequisites:** M35 and at least three distinct fixture rigs.
- **Decisions:** semantic roles map to actual bones; explicit profiles cover rest
  pose, coordinate/root conversion, proportions, missing bones, masks, additive
  clips, and root policy.
- **Packages / schemas / tools:** retarget profile and validator; lab source/target
  comparison; `dm-asset validate-retarget` JSON diagnostics.
- **Assets / tests / acceptance:** compatible humanoids retarget with golden bone
  transforms; a non-humanoid profile proves custom mappings; incompatible pairs
  name every missing role/bone and failed constraint rather than merely failing.
- **Blender / codecs:** addon authors/visualizes profiles; no codec work.
- **Risks / deferred:** library choice requires prototype; automatic mapping,
  motion matching, and high-end deformation correction defer.

## M42: Build graph, cache, and live iteration

- **Goal / reason:** rebuild only affected outputs and make lab iteration fast,
  inspectable, and reproducible.
- **Prerequisites:** stable M31-M41 inputs and outputs.
- **Decisions:** content keys include source/config/profile/tool/processor versions;
  generated artifacts are immutable; hot reload is transactional; development
  generation is local and explicit.
- **Packages / tools:** dependency graph, content store, watch/invalidation,
  `dm-asset build --json`, lab regenerate/reload, CI determinism report.
- **Assets / tests / acceptance:** changing `helmet.glb` rebuilds only dependents;
  an animation invalidates relevant samples; palette-only change requantizes
  without rerendering; failed rebuild preserves the live instance. Blender-to-
  engine integration checks the complete pinned path.
- **Blender / codecs:** headless Blender nodes are versioned external processors;
  codec versions participate in keys.
- **Risks / deferred:** cache eviction policy and production cache-miss generation
  defer; file watchers are hints and correctness never depends on event delivery.

## M43: Measured advanced runtime and generation

- **Goal / reason:** admit expensive optimizations and extension points only from
  creature-lab/world captures and explicit budgets.
- **Prerequisites:** M42 instrumentation baseline.
- **Decisions:** animation evaluation remains independent of skinning; resources
  are shareable while instances own pose/material/equipment state; visual LOD and
  collision remain separate.
- **Packages / tools:** metrics for skeleton evaluation, skinning, draw calls,
  memory, caches, render/quantize time; profile-driven backend improvements.
- **Assets / tests / acceptance:** a documented profile chooses or rejects each
  proposed GPU skinning, instancing, alternate representation, atlas, or parallel
  rendering change; accepted work meets budgets without changing golden poses,
  events, composition, deterministic checks, or teardown behavior.
- **Blender / codecs:** none unless evidence identifies an owning-repo bottleneck.
- **Risks / deferred:** procedural animation, IK, foot placement, look-at, cloth,
  automatic LOD, distributed generation, advanced graphs, and production on-
  demand generation remain future work until separately accepted.

## Known architectural decisions

- Semantic definitions are independent of source format, DCC/COF, Blender, and
  render backend; glTF is the preferred first interchange format.
- Versioned schemas never silently reinterpret old versions.
- Immutable resources and mutable instance state are separate.
- Arbitrary components share a primary pose/clock; equipment changes normally
  preserve normalized time. Secondary layers may own explicit clocks.
- Facing is continuous; exporters quantize to any configured direction count.
- Sockets are semantic roles mapped to bones or authored/procedural transforms.
- Animation events are time-based and interval-safe; collision/selection are not
  inferred from render meshes.
- Generated component frames preserve a common origin. Legacy export is lossy,
  explicit, and implemented through independent codecs.
- Headless JSON tooling and redistributable humanoid/non-humanoid fixtures are
  acceptance infrastructure, not optional polish.

## Decisions requiring prototypes

1. glTF extras/custom extension versus companion manifests.
2. CPU reference versus first GPU skinning path and matrix conventions.
3. Indexed textures/LUTs versus mask-based palette-aware 3D materials.
4. Exact Diablo-oriented camera and legacy sample/order conventions.
5. Blender, Dark Magic offline renderer, or both for sprite generation.
6. Determinism contract across Blender versions, GPUs, and render backends.
7. Modern composite atlas/array/indexed storage after semantic profiling.
8. Retargeting library and rest-pose/proportion strategy.

Prototype results must record fixtures, versions, measurements, rejected
alternatives, and the decision before the dependent milestone starts.

## Unresolved compatibility questions

- Exact DCC/DC6 encoder constraints versus original renderer/tool limits,
  direction ordering, terminal samples, frame trailers, offsets, and bounds.
- Complete modern-event-to-COF mapping and which events/order rules are lossy.
- PL2 transform behavior for modern material regions and indexed transparency.
- Camera/projection/lighting parity across historical output and current tools.
- Which legacy component suppression, weapon-class fallback, and per-frame order
  rules belong only in the compatibility adapter.

## Codec repository follow-ups

`gravestench/dc6`, `dcc`, `cof`, and `pl2` own strict validation, deterministic
marshaling/encoding, indexed-frame/palette APIs, COF construction, round trips,
and inspectors. Dark Magic owns creature semantics, build stages, direction
sampling, loss reports, and presentation. The detailed handoff belongs in
`docs/formats/CODEC_FOLLOWUPS.md`; codec work should be separately branched and
tagged before M40 consumes it.

## Blender addon migration findings

The first-party
[`gravestench/blender-addon-dc6-export`](https://gitlab.com/gravestench/blender-addon-dc6-export)
prototype was created in 2020 for Blender 2.80. It demonstrated a useful
artist-visible circular orthographic camera rig, arbitrary direction count,
single-frame/animation batch rendering, palette remapping, and direct DC6 output.

Classify its capabilities as follows:

| Capability | Disposition |
| --- | --- |
| Camera target and evenly spaced directional views | Reuse conceptually as explicit camera profiles |
| Render all directions and timeline frames | Reuse conceptually through an explicit sampler/build graph |
| Palette selection and nearest-color conversion | Redesign as shared headless quantization |
| DC6 writing inside the addon | Redesign in `gravestench/dc6` plus Dark Magic exporter adapter |
| Blender scene mutation, multiview naming, render handlers | Requires validation/redesign for supported Blender versions |
| Destructive rig recreation and deletion by name/collection | Obsolete and unsafe |
| Bundled static palettes, Pillow setup, transparency index 0 | Obsolete as architectural ownership; replace with versioned inputs |
| Manual Cycles/black-world/stereoscopy scene setup | Obsolete as undocumented authority; express in profiles |

The addon README itself labels single-frame export under-tested, warns that rig
property changes destroy/recreate the rig (and can delete misplaced objects),
and hard-codes palette index 0 as transparent. Those are migration lessons, not
contracts. The replacement addon should annotate and preview; shared tooling
should validate, compile, quantize, and encode.

## Dust3D authoring evaluation

[`huxingyi/dust3d`](https://github.com/huxingyi/dust3d) is a promising MIT-licensed
low-poly authoring frontend for Dark Magic's original creature assets. It offers
real-time sketch-to-mesh construction, front/side background-image references,
automatic UV unwrapping, bone assignment and skinning assistance, procedural
animation, and GLB/FBX export. This substantially lowers the cost of producing
stylized humanoid and non-humanoid prototypes for M32-M38.

Dust3D is not treated as an automatic image-to-3D authority. Its documented
reference-image workflow is artist-guided: an artist loads front/side images and
traces structural nodes and parts. That is useful and controllable, but generated
geometry, rigs, skin weights, clips, materials, coordinate conventions, and
licenses still require validation and provenance.

The initial integration is deliberately file-based:

1. Author and retain the native `.ds3` file plus source-image provenance outside
   runtime packages.
2. Export GLB as the preferred interchange artifact; use FBX only as a diagnostic
   fallback.
3. Run the same `dm-asset inspect|validate|build` path used for Blender exports.
4. Record Dust3D version, source hash, export settings, and importer diagnostics
   as build inputs.
5. Compare fixed-time golden poses and deterministic eight-direction sprite
   output against a Blender-authored fixture before admitting Dust3D as a tested
   production frontend.

Do not embed or fork Dust3D into the engine, depend on `.ds3` at runtime, or let
its generated skeleton/animation vocabulary become Dark Magic's creature model.
If GLB export loses required semantic sockets, typed events, collision/selection
metadata, or material regions, store those in the same versioned companion
manifest evaluated by M31-M32 rather than inventing a Dust3D-specific runtime.

## Future possibilities

Preserve extension points for procedural animation, IK, foot placement, aim/look
offsets, recoil, secondary motion, morph/facial layers, GPU instancing, automatic
LOD/impostors, production on-demand generation, remote workers, and additional
renderer/importer/exporter plugins. None belongs in the near-term proof, and no
plugin interface should be generalized beyond demonstrated second implementations.
