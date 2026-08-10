# Creature asset research backlog

Status: **candidate research program; not scheduled implementation milestones**.

This document is the bridge between the established M31-M43 program in
`docs/CREATURE_ASSET_ROADMAP.md` and the broader brainstorming ledger in
`docs/research/creature-visual-system.md`.

The purpose is to preserve promising ideas in a form Codex can investigate later
without pretending that every brainstormed concept is already approved
architecture.

## Promotion policy

Do not automatically turn these research checkpoints into M44+ work.

Before promotion, review each checkpoint and classify it:

```text
ACCEPT
PROTOTYPE
DEFER
REJECT
NEEDS MORE RESEARCH
```

An `ACCEPT` or successful `PROTOTYPE` can become a numbered roadmap milestone
only after it has:

- a concrete problem statement;
- repository ownership and package boundaries;
- dependencies on completed M31-M43 work;
- alternatives considered;
- synthetic fixtures;
- objective acceptance criteria;
- migration/compatibility risks;
- explicit deferred scope.

Rejected and deferred ideas should stay documented so future work does not
reopen the same question without new evidence.

## Dependency context

The current roadmap already establishes the near-term path:

```text
M31 schemas/evidence
  -> M32 glTF import
  -> M33 Creature Lab
  -> M34 composition/sockets/equipment
  -> M35 semantic animation/events
  -> M36 Blender frontend
  -> M37 materials/palettes
  -> M38 deterministic sprite generation
  -> M39 modern sprite composition
  -> M40 legacy export
  -> M41 retargeting
  -> M42 build/cache/live reload
  -> M43 measured advanced runtime
```

The research tracks below should reuse those capabilities rather than building
parallel implementations.

---

## R1: Generalized visual actor / clip / track model

**Question:** Should Dark Magic have one generalized visual-animation model above
both legacy sprite composition and modern rigged rendering?

### Motivation

A track-based abstraction may let gameplay and creature state remain independent
from whether a visual layer comes from DCC/DC6, generated sprites, rigged 3D,
particles, effects, or future renderers.

Conceptual vocabulary only:

```text
VisualActor
  +-- VisualClip
        +-- RigTrack
        +-- SpriteTrack
        +-- ParticleTrack
        +-- EventTrack
```

### Research tasks

- Map the existing M31-M39 semantic types into this conceptual model.
- Map COF/DCC/DC6 composite behavior into the model without preserving fixed
  legacy component/direction assumptions.
- Model one modern rigged creature and one legacy composite through the same
  test-facing interface.
- Identify whether tracks are the right abstraction or whether components plus
  animation channels already cover the need.
- Define what belongs to semantic state versus renderer-specific track state.
- Verify resource/state separation and shared-clock behavior.

### Prototype

Create a small in-memory/reference prototype only after M34-M35 contracts are
stable. It should expose one synthetic attack clip whose body is rigged, whose
weapon is replaceable, and whose event timeline is semantic. A legacy adapter
prototype should expose equivalent time/direction/component information through
the same inspection API.

### Evidence for acceptance

- Gameplay-facing code can inspect/play the same semantic clip without branching
  on DCC versus glTF origin.
- The abstraction does not require fixed component names or direction counts.
- It does not duplicate the existing creature/component/animation model merely
  for naming convenience.
- Renderer adapters remain replaceable.

### Reasons to reject

Reject or collapse the idea if M31-M39 types already provide the same separation
without another abstraction layer.

---

## R2: Mixed 2D / 3D compositional rendering

**Question:** Is mixed representation useful enough to justify supporting 3D,
sprite, particle, and legacy tracks on one actor?

### Candidate use cases

```text
3D body + 3D weapon + sprite aura + particles
legacy sprite body + generated equipment overlay + procedural light
```

### Research tasks

- Inventory real engine/modding use cases rather than designing from possibility
  alone.
- Determine ordering and depth semantics between 3D geometry and 2D overlays.
- Determine shared-origin/socket/facing contracts.
- Determine how shadows, selection, lighting, and palette transforms interact.
- Measure whether this can remain a presentation-only concern.

### Prototype

In Creature Lab, attach one generated sprite/effect layer to a rigged synthetic
creature and one modern effect to a legacy sprite creature. Do not generalize a
plugin system before these two cases work.

### Acceptance evidence

- Composition semantics are deterministic and inspectable.
- Gameplay collision/selection remain representation-independent.
- The additional renderer complexity enables actual useful workflows.

### Default disposition

`PROTOTYPE` after M39. Do not block M31-M40 on this work.

---

## R3: Unified sprite track/compositor boundary

**Question:** Can decoded legacy frames and newly generated sprite frames share
one modern sprite-track/compositor contract?

### Motivation

Generated sprites should not require a second compositor merely because they
came from Blender or a 3D renderer.

Conceptually:

```text
legacy DCC      -> SpriteTrack
rendered sprite -> SpriteTrack
                       |
                       v
                 modern compositor
```

### Research tasks

- Compare M39 modern composite semantics to current legacy presentation paths.
- Identify the minimum normalized frame metadata: image/index data, actor-origin
  offset, bounds, direction/sample identity, palette/material parameters.
- Keep decoder-specific storage and historical quirks behind adapters.
- Prove per-component and flattened generated output can use the same renderer.

### Acceptance evidence

- A synthetic legacy-style fixture and an M38 generated fixture render through
  the same compositor contract.
- Common-origin reconstruction is exact.
- Palette/indexed data can remain indexed until the renderer chooses a lookup.
- No codec-specific state leaks into the semantic compositor API.

### Relationship to M39

This may become an M39 refinement rather than a new milestone. Promote it only if
current M39 implementation cannot accommodate both sources cleanly.

---

## R4: Indexed textures and palette-space materials

**Question:** Which representation best preserves Diablo-style palette behavior
for both legacy sprites and new 3D-authored assets?

### Candidate approaches

- indexed textures + palette/LUT resources;
- grayscale masks;
- region-ID textures;
- multi-mask materials;
- PL2-compatible lookup tables;
- hybrid indexed sprite / mask-based 3D materials.

### Research tasks

- Compare memory, authoring complexity, renderer cost, fidelity, and modding
  ergonomics.
- Verify transparency semantics separately from color index semantics.
- Prototype instance-local recoloring without mutating shared resources.
- Compare legacy PL2 behavior with modern named regions such as skin, cloth,
  metal, armor primary/secondary, weapon, and monster variant.
- Determine what Blender should preview and how closely that preview can match
  runtime/headless rendering.

### Prototype

Use the M37 palette fixture with at least two creature instances sharing the same
source textures but using different palette/material parameters. Capture a
legacy indexed sprite path and a modern 3D material path.

### Acceptance evidence

- Transform semantics are documented and backend-neutral.
- Shared resources remain immutable.
- The chosen strategy is authorable in Blender and inspectable headlessly.
- It does not force every modern material to be 8-bit indexed if that provides no
  benefit.

### Default disposition

`PROTOTYPE`; this should inform M37 implementation details rather than introduce
an unrelated renderer architecture.

---

## R5: Render-profile and multi-output asset generation

**Question:** Can one source model/rig/animation library reproducibly target live
3D, multiple sprite-direction profiles, portraits/thumbnails, and legacy output?

### Candidate profile inputs

```text
camera
lighting
direction count
sample times / FPS
output resolution
palette
quantization
shadow policy
component policy
renderer/tool versions
```

### Research tasks

- Define explicit profiles instead of relying on Blender scene state.
- Reproduce the useful camera/directional concepts from the historical
  `gravestench/blender-addon-dc6-export` addon while redesigning old Blender API
  assumptions.
- Compare 8-, 16-, and higher-direction output from one source clip.
- Define terminal-frame and event-to-sample policies.
- Determine whether output profiles belong in creature packages, project config,
  or shared tool data.

### Prototype

Render one synthetic attack clip as:

```text
live 3D
8-direction indexed sprites
16-direction modern sprites
portrait/thumbnail
```

The source clip must remain unchanged across outputs.

### Acceptance evidence

- Profiles are versioned, reproducible, and headlessly inspectable.
- Generated output records enough provenance to reproduce the build.
- Direction count is not part of logical actor orientation.

### Relationship to M38/M40

Most of this should sharpen M38 and M40 rather than become separate runtime work.

---

## R6: Development-time on-demand generation

**Question:** Should Creature Lab and authoring tools be able to request missing
or stale generated representations automatically during development?

### Candidate workflow

```text
edit model/animation/material
        |
        v
invalidate affected dependency nodes
        |
        v
Creature Lab requests representation
        |
        v
render/quantize only missing products
        |
        v
transactional hot reload
```

### Research tasks

- Reuse the M42 dependency graph/content cache rather than building a second
  generation cache.
- Define generation requests in terms of actor, clip, component, direction,
  sample, and render profile.
- Ensure failed generation leaves the current live artifact untouched.
- Measure interactive latency for common edits.
- Separate local development convenience from production runtime behavior.

### Prototype

Change a helmet source, one animation clip, and one palette independently and
verify only the required generated nodes rebuild.

### Acceptance evidence

- Helmet changes do not rebuild unrelated body frames.
- Clip changes do not invalidate unrelated clips.
- Palette-only changes can re-quantize without geometry rendering where valid.
- Generation and reload are deterministic and observable in Creature Lab.

### Default disposition

Likely `ACCEPT` as a development goal after M42, but implementation shape should
be prototyped.

---

## R7: Asset generation subsystem and processor/plugin boundaries

**Question:** What reusable generation boundary should serve the CLI, Creature
Lab, Blender addon, CI, and batch tools?

### Candidate subsystem

```text
request(actor, clip, component, direction, sample, profile)
       |
       v
asset generation subsystem
       +-- dependency lookup
       +-- cache lookup
       +-- renderer
       +-- quantizer
       +-- encoder/exporter
       |
       v
artifact
```

### Candidate extension seams

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

### Research tasks

- Prefer library/subsystem boundaries first; do not create a daemon/service
  without a measured isolation or distributed-processing need.
- Require at least two real implementations/consumers before generalizing an
  extension interface.
- Keep DCC/DC6/COF binary mechanics in independent codec repositories.
- Keep Blender as a frontend/orchestrator.
- Define processor versioning for deterministic cache keys.

### Acceptance evidence

- CLI and Creature Lab can use the same generation API.
- Blender can orchestrate the same API without embedding private encoders.
- Adding a second renderer/quantizer does not require gameplay changes.
- The abstraction is smaller than the combined implementation it replaces.

### Default disposition

`PROTOTYPE`; avoid speculative plugin frameworks.

---

## R8: Alternate representations, skinning, instancing, and advanced animation

**Question:** Which advanced runtime capabilities are actually justified after
M43 profiling?

### Candidate directions

- GPU skinning;
- instanced skeletal rendering;
- alternate high/gameplay/low-detail representations;
- generated sprite or impostor LOD;
- IK and foot placement;
- aim/look offsets;
- recoil;
- procedural secondary motion;
- facial/morph layers;
- automatic LOD;
- parallel or distributed generation.

### Research tasks

- Start with profile captures and explicit budgets.
- Keep animation evaluation independent from skinning backend.
- Keep visual LOD independent from gameplay collision/footprint.
- Verify shared resource/per-instance state boundaries under high instance counts.
- Do not build advanced animation graphs until real creature behavior requires
  them.

### Acceptance evidence

Each accepted optimization or feature must meet a measured budget and preserve
existing golden poses, event delivery, composition, teardown, and deterministic
asset checks.

### Default disposition

`DEFER` until M43 produces evidence.

---

## R9: Mod-level creature and visual overrides

**Question:** How should mods replace or extend creature visuals without copying
entire definitions?

### Candidate operations

- replace a model, material, or texture;
- replace/add a clip;
- add/remove a component;
- replace equipment visuals;
- add a variant;
- remap a socket;
- change palette/material parameters;
- alter visual-state resolution.

### Research tasks

- Reuse the layered VFS and stable logical asset IDs.
- Define merge/override semantics for versioned creature manifests.
- Preserve source provenance in diagnostics.
- Detect incompatible overrides early.
- Verify that generated-cache keys include resolved provenance/content.

### Acceptance evidence

A synthetic mod can replace one weapon and one animation without copying the
whole creature package, and removal of the mod cleanly restores the base
resolution.

### Default disposition

Likely future `ACCEPT`, but schema merge semantics need deliberate design.

---

## R10: Authoring frontend interoperability

**Question:** Can Blender, Dust3D, and future tools feed one canonical import and
validation boundary without creating tool-specific runtime models?

### Research tasks

- Continue using glTF/GLB plus versioned semantic metadata as the common boundary
  unless prototypes reject it.
- Compare Blender and Dust3D exports using the same golden-pose, bounds, material,
  socket, and sprite-generation tests.
- Keep source-authoring files (`.blend`, `.ds3`, etc.) outside runtime contracts.
- Evaluate additional authoring/rigging tools only through the candidate-tool
  rubric in `docs/CREATURE_ASSET_ROADMAP.md`.
- Record source/tool versions and asset provenance.

### Historical Blender addon requirement

Always include the first-party historical addon in Blender-planning reviews:

<https://gitlab.com/gravestench/blender-addon-dc6-export>

It targeted Blender 2.80-era APIs and should be mined for useful workflow ideas,
not preserved as-is.

### Acceptance evidence

At least two authoring frontends can produce assets that pass the same canonical
headless validation without requiring tool-specific runtime branches.

---

# Cross-cutting research rules

## Synthetic fixtures first

Use redistributable synthetic assets for normal CI. Maintain humanoid,
equipment-heavy humanoid, structurally distinct non-humanoid, palette-region,
and socket fixtures. Use fixed-time golden pose tests in addition to screenshots.

## Proprietary compatibility is optional

Original Diablo II data remains user-supplied and optional. Never commit Blizzard
assets. Prefer structural assertions and derived diagnostics over redistributing
content.

## Source-of-truth policy

For newly authored creatures, source models, rigs, animations, textures,
materials, semantic manifests, and generation profiles are source assets.
Generated sprite sets, atlases, DC6/DCC, and other encoded products should be
rebuildable artifacts where practical.

## Ownership policy

Future work must name its owner explicitly:

```text
Dark Magic runtime
Dark Magic asset tooling
Dark Magic developer tooling
Blender addon / authoring frontend
independent codec repository
renderer backend
content/mod layer
```

Examples:

```text
animation event semantics       -> Dark Magic
artist event marker UI          -> Blender addon
DCC encoding                    -> gravestench/dcc
directional sprite generation   -> Dark Magic asset tooling
legacy DCC playback adapter     -> Dark Magic presentation/runtime adapter
```

## No premature compatibility constraints

Do not hard-code the legacy D2 component set, 8 directions, frame-number-based
animation semantics, COF as the canonical model, or DCC/DC6 as the source format
for newly authored creatures.

## No premature plugin framework

Only promote importer/renderer/processor interfaces after real second
implementations demonstrate the seam. Prefer narrow contracts over a speculative
general plugin system.

## No silent milestone promotion

When research produces a decision, update this file with the evidence and then
add a concrete numbered milestone to `ROADMAP.md` and/or
`docs/CREATURE_ASSET_ROADMAP.md` in a separate planning review. Do not implement
an R-track simply because it appears here.

# Suggested review order

Once M31-M39 are mature enough to provide real prototypes, review the backlog in
this order:

1. R1 generalized visual model;
2. R3 common sprite compositor boundary;
3. R4 palette/material representation;
4. R5 render-profile/multi-output generation;
5. R6 development-time on-demand generation;
6. R7 shared generation subsystem/plugin seams;
7. R2 mixed 2D/3D composition;
8. R9 mod-level overrides;
9. R10 frontend interoperability;
10. R8 advanced runtime work only after M43 measurement.

This ordering is advisory. New evidence may change it.

# Review outcome template

Use this template when resolving a research checkpoint:

```text
Research checkpoint:
Decision: ACCEPT | PROTOTYPE | DEFER | REJECT | NEEDS MORE RESEARCH
Date/revision:
Problem:
Evidence/fixtures:
Alternatives:
Chosen direction:
Rejected directions:
Owning package/repository:
Dependencies:
Tests/acceptance:
Compatibility impact:
Migration impact:
Follow-up milestone/issue:
```

The goal is to keep good ideas visible while making it impossible for a
brainstorming note to masquerade as an approved implementation contract.
