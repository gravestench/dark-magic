# Dark Magic roadmap

Status: fully refreshed through the G4 player-population/override correction,
the target-locked party-XP probe contract, and the G5 production Warp Lab,
post-warp route invalidation, semantic motion ownership, stat-derived locomotion
playback, pinned class movement/stamina, authoritative drain/recovery/FRW,
armor/shield/cold-source ordering, and progression/source-derived maximum-
stamina plus environment-period source slices, G6 deterministic dynamic
occupancy, a generic checkpointed forced-motion transaction, and target-pinned
monster knockback capability/size profiles plus authored missile knockback
values, deterministic forced-motion replacement/locomotion ownership, and
stable semantic motion-event identities. The client now also has a compile-time
desktop boundary: Raylib remains the production/default backend, while an
experimental `ebitengine` tag drives the same client, Lua scenes, retained
composer, input actions, captures, and profiler diagnostics through Ebitengine.
Both binaries compile in CI, and a matched, audio-muted A/B workflow produces
backend-tagged real-asset profiles, captures, and a Markdown comparison. The
first corrected Blood Moor run measured Raylib/Ebitengine frame p95 at
17.277/16.811 ms and final native rendering at 0.505/0.399 ms with 150/149 draw
submissions; this establishes competitiveness, not a production-default change.
Fresh full-client compilation from separate empty Go build caches took
33.87 seconds for Raylib and 22.68 seconds for Ebitengine; immediate rebuilds
took 0.51 and 0.39 seconds respectively on the same machine.
The authored-button parity defect found during review was an incorrect draw-
mode-4 destination blend factor; its focused UI Lab crop is now pixel-identical
between backends. The
Ebitengine adapter is not release-equivalent yet: native audio is muted, the
developer-console overlay is headless, final display-palette quantization is
rejected, and node-palette quantization is CPU-cached. G7 now separates authoritative world
existence from active simulation with an empty ECS inactive tag: room residents
retain their entity IDs, full component state, and timed-state/stat-source/event
references across deactivate, checkpoint, restore, and reactivate, while
simulation and both local/remote presentation projections exclude them. Room
residency now has a world-owned stable resident ID rather than depending on
monster identity, and the activation record preserves each resident's generic
velocity-mover opt-in. Residency is scoped by canonical level/room IDs, and
production DS1 interaction targets now acquire it from generated zone geometry
without turning map/presentation data into authority. Warp Lab's paired
authoritative endpoints use the same geometry join and the real-asset lab test
pins both endpoint residents. Active moving residents now synchronize their
room membership from authoritative level/position before the next activation
decision, so crossing a generated boundary does not leave them owned by their
spawn room. The first owned-unit residency slice now proves that an ordinary
resident can retain its ECS owner entity reference, category/limit/lifetime
policy, durable identity, and attribution fields across deactivate,
checkpoint, restore, and reactivate. Its lifetime system excludes the same
empty inactive marker; absolute expirations are evaluated on the first active
tick without claiming exact 1.14d inactive timer aging. The first corpse-
residency slice also keeps the dead monster's semantic entity,
death/loot/identity/appearance/spatial state, and room identity across the same
checkpoint path. Death now removes the generic velocity-mover opt-in alongside
AI, collider, and selection, so deactivation cannot record and reintroduce a
stale simulation capability. Exact corpse lifetime and usability policy remain
probe-gated. Straight-missile materialization now asks the installed population
plan for canonical room residency and assigns a deterministic world-owned ID.
Projectile movement and lifetime progression exclude the same empty inactive
marker, so an in-flight entity keeps its authoritative position, projectile
state, and room identity across deactivate, checkpoint, restore, and reactivate
without a projectile archive. The long synthetic lifetime used to exercise
that boundary is test scaffolding; exact Expansion 1.14d missile lifetime and
inactive-room timing remain unresolved. The first ground-item residency slice
now gives an imported `world` placement authoritative position/location and a
generic ECS room-attachment request. Plan admission resolves that request into
the same stable resident contract; inactive items disappear from local/private
item projections, survive checkpoint reconstruction on the original entity,
and reuse ordinary placement commands to remove spatial state on pickup and
reacquire it on re-drop. This proves residency and placement transitions, not
public loot ownership, legal drop-point search, pickup range, or exact 1.14d
ground lifetime. A synthetic data-selected one-shot object family now proves
that interaction can mutate ordinary ECS object mode/used/revision state, and a
separate pending-action entity carries a raw target reference plus its own
stable room residency. Both cross inactive checkpoint/reactivation on their
original entities with checksum parity. This completes G7's type/relationship
mechanism breadth without claiming retail Objects.txt family mappings,
scheduled-event execution, collision transitions, or exact 1.14d inactive
event timing. G8 now has one Lua-owned direct-damage commit boundary shared by
melee and missile contact. Its explicit result records distinguish rolled,
mitigated-and-committed, remaining-health, channel, and lethal facts. A
successful melee result is one ECS entity composed from the generic damage
event and melee-specific reaction event, so death, reactive-state, replay, and
future effect consumers do not need parallel attacks or inferred joins. The
current whole-health player component quantizes applied raw damage at that
boundary; exact Expansion 1.14d fractional player-life storage/rounding remains
probe-gated rather than inherited from older recovered code. The next G8 layer
composes a typed damage-bundle component on that same result entity. Physical,
fire, lightning, cold, magic, and poison rolled/mitigated values stay separate
through channel mitigation; only the immediate channels join at the health-
commit boundary. Poison remains recorded but cannot mutate health until its
rate/duration transaction exists. Current melee and missile families merely
populate their authored single channel. Drain and duration semantics remain
unimplemented and no unsupported channel arithmetic is implied. G9 remains current through
target-locked mounted-data and localized TBL skill evidence, case-stable pinned
MPQ tables, AnimData/effective-attack-rate generic melee action, current-state
melee target revalidation, missile, timed-state, and reactive-state slices as
of 2026-08-17. A matched frontend profile also removed the title-to-main-menu
localization stall by buffering each small MPQ-backed TBL once before decoding;
staged title/menu, secondary-destination, and character-interaction preload
bundles then reduced the settled main-menu heap from 487 MB to 216 MB without
leaving background work pending. G6 research now also has an Expansion-1.14d-
only owned-runtime knockback probe contract that rejects Classic, earlier-
version, community-tool, and unmatched-control observations before any older
recovered chance/distance hypothesis can become gameplay policy.

This file is the implementation-status authority. The documents under
`docs/research/` are the fidelity and evidence authorities. A checked item here
means executable repository evidence satisfies the stated acceptance boundary;
it does not upgrade an inferred Diablo II behavior to verified behavior.

## Direction

Dark Magic now has one authoritative, renderer-independent `d2legacy` Lua
simulation that can run offline, on a listen server, on a dedicated server, or
under the Realm. The next era is therefore gameplay completion, not a second
network gameplay stack.

The sole product target is an increasingly complete expansion-only Diablo II:
Lord of Destruction 1.14d implementation. Classic mode and earlier patch
behavior are out of scope. Lua owns Diablo II policy. Go owns reusable engine,
transport, storage, replay, checkpoint, renderer, and audio mechanisms.

The acceptance loop for this era is:

```text
Realm allocates one pinned game
  -> multiple authenticated players join one Session
  -> immutable content/GameRules plus checkpointed mutable policies govern the game
  -> world activation, locomotion, occupancy, combat, party context, and loot run authoritatively
  -> a real item reaches the ground, inventory, equipment, and gameplay stats
  -> quest/object state advances
  -> checkpoint/reconnect reproduces the state
  -> one revisioned durable character commit succeeds
```

Do not add broad content until the mechanism beneath it is coherent. Do not
create parallel stat, combat, skill, monster, item, quest, transition, session,
party, targeting, or persistence authorities.

## Status vocabulary

- **complete**: the acceptance boundary is covered by executable evidence.
- **partial**: useful production implementation exists, but at least one named
  acceptance condition is absent.
- **foundation complete**: the old milestone's architectural purpose is done;
  remaining breadth has moved into the gameplay queue below.
- **research only**: evidence exists, but production implementation does not.
- **deferred**: intentionally outside the current critical path.

Compatibility labels in code and research remain: **verified**,
**high-confidence recovered behavior**, **inferred**, **synthetic Dark Magic
policy**, and **unresolved**.

## Current implementation baseline

| Area | Status | Repository evidence and remaining boundary |
| --- | --- | --- |
| M0-M14 engine/application foundations | complete | Reproducible core, layered content, Lua runtime, ECS, rendering composition, application host, and service-mesh retirement are established. |
| M15 asset knowledge | partial | Typed/recovered coverage is broad. The owned 1.14d Expansion Skills/Missiles report now inventories 357 skill rows, 172 server behavior signatures, 3 exact-ID implementations, and winning-layer provenance. A second exact-ID report joins Skills/SkillDesc formulas to layered locale TBL text, replacement tokens, and cross-skill references; unresolved records and source-sensitive mappings remain research work. |
| M16 presentation primitives | partial | MPQ-backed render/audio primitives exist. Client assembly now consumes a backend-neutral desktop contract; Raylib is the production default and the `ebitengine` build tag supplies an experimental retained-composition/input/capture adapter. Ebitengine native audio, console drawing, and GPU palette parity remain before it can become release-equivalent. |
| M17 front end | foundation complete | The Lua-authored front end and Realm flow exist. MPQ-backed locale tables now cross one sequential buffering boundary instead of issuing decoder-granularity random archive reads. Startup warms only title/main-menu assets, secondary destinations use visible main-menu think time, and character interaction animations remain scoped to character creation. Remaining work is UI fidelity, not the former multi-second transition stall or whole-frontend eager preload. |
| M18 in-game shell | foundation complete | HUD and major overlay shells exist; the party panel now consumes an owner-scoped semantic projection, while remaining raw/ad hoc reads migrate as their gameplay domains mature. |
| M19 character/item/save fidelity | partial | Canonical profile and Realm character persistence exist; the complete Dark Magic durable semantic character does not. Vanilla save interoperability is out of scope. |
| M20 world fidelity | partial | Deterministic Act I generation, collision, transitions, dynamic occupancy, population, and level-scoped persistent-identity room residents exist. Timed-state/stat-source/event references, an owned-unit graph, corpse, straight projectile, imported ground item, stateful interaction object, and separately resident pending-action relationship survive inactivation without scalar graph copies. The inactive ECS tag removes live capabilities, suspends opted-in systems, and filters projections; generic pre-plan attachment plus pickup/re-drop transitions reuse the same contract. Public loot policy, retail object/event families, exact corpse/projectile/event timing, 1.14d streaming behavior, and campaign breadth remain. |
| M21 Diablo simulation | foundation complete | Lua owns the current player, monster, skill, missile, state, death, loot, quest, item, and owned-unit vertical slices. Melee and missile contact now share one ordered direct-damage commit/result boundary, with ECS component composition preserving both generic and source-specific semantics. Block/avoidance, typed bundle breadth, secondary damage effects, player death, movement, item activation, object, and content breadth remain below. |
| M22 networking | complete | One `Session`, authenticated semantic commands, deterministic ordering, filtered views, reconnect, replay/checkpoint, direct/listen/dedicated/Realm modes, and impairment/soak coverage exist. |
| M23 Realm/persistence | partial | Accounts, characters, leases, CAS commits, allocation, admission, reconnect, checkpoints, PostgreSQL, mail, and process workers exist. Publication/revocation, complete durable character semantics, and production operations remain. |
| M24 packaging/release | partial | Build/release foundations exist; the gameplay acceptance loop and final supported-platform release gate are not complete. |
| M25-M30 performance/UI/architecture | partial | Major residency, profiling, Lua-policy migration, and archetype ECS work landed. The matched title-to-main-menu capture reduced the worst profiled update from 4.134 s to 152 ms and removed the 3.77 s TBL random-read hotspot. Staged frontend bundles then reduced settled main-menu heap from 487 MB to 216 MB, preloader-retained heap from 357 MB to 112 MB, and decoded-cache weight from 339 MB to 59 MB with zero pending preloads. A new compile-time Raylib/Ebitengine experiment keeps simulation and composition identical, compiles both clients in CI, and owns a matched capture/profile/summary command rather than relying on subjective window feel. Measured backend results and Ebitengine feature parity remain open. |
| M31-M43 creature authoring | deferred | Generated creature representation is independent work and must not displace the gameplay critical path. |
| M44 Realm cloud operations | deferred | Local topology-neutral Realm is the prerequisite. Existing deployment groundwork does not make cloud operations a gameplay gate. |

The old milestone numbering is retained only as historical orientation. New work
uses the ordered gameplay gates below. This avoids preserving an obsolete plan
in which networking followed the first gameplay loop.

### Active frontend performance follow-up

- [x] Capture a real-asset, per-scene CPU/heap profile across trademark, title,
  loading, and main-menu navigation rather than attributing the hitch to host
  memory pressure alone.
- [x] Replace decoder-granularity `ReaderAt` access to compressed MPQ TBL files
  with one bounded sequential read per table, and lock the boundary with a
  filesystem that would expose any regression back to random reads.
- [x] Repeat the same instrumented navigation: the title scene's worst update
  fell from 4.134 s to 152 ms (96.3%), while the former 3.77 s TBL decoder
  hotspot disappeared and steady main-menu updates remained sub-millisecond at
  p95.
- [x] Split the former whole-frontend startup bundle into title/main-menu,
  secondary-destination, and character-creation-interaction stages without
  adding another cache/lifetime authority. A settled real-asset main-menu
  capture completed every background request while reducing profiled heap from
  487 MB to 216 MB, the preloader subtree from 357 MB to 112 MB, and decoded-
  cache weight from 339 MB to 59 MB. Main-menu update p95 remained below one
  millisecond and improved from 0.794 ms to 0.344 ms.
- [x] Move client assembly behind one compile-time desktop contract without
  leaking either native API into gameplay, Lua presentation, or the retained
  composer; preserve Raylib as the untagged default and add an explicit
  `ebitengine` build.
- [x] Prove the experimental Ebitengine path with a real-asset `ui_lab`
  lifecycle/capture smoke run and compile both backend binaries in CI.
- [x] Add a matched A/B profiler that builds once, disables native audio in
  both clients, runs identical scene/fixture/settle inputs, preserves raw CPU/
  heap/diagnostic/capture artifacts, and writes a compact Markdown comparison.
- [x] Record the first corrected simple/UI Lab and Blood Moor decision inputs:
  authored-button crops are pixel-identical, world captures align, draw counts
  are 150/149, frame p95 is 17.277/16.811 ms, and final native rendering is
  0.505/0.399 ms for Raylib/Ebitengine on the initial Apple Silicon run.
- [ ] Repeat the matched runs and add representative dense-combat and palette-
  heavy profiles before changing the production default.
- [ ] If Ebitengine remains competitive, replace its muted audio, headless
  console, rejected final-palette transform, and CPU node-palette fallback with
  native adapters and visual/performance parity evidence. Otherwise remove the
  experiment rather than carrying two indefinite renderer products.

## P0: post-networking gameplay foundations

### G1 — Network gameplay-boundary acceptance

Status: **complete**.

- [x] Remote principals share the same authoritative `internal/game/session.Session`.
- [x] Local and remote semantic commands enter the same validators and systems.
- [x] Same-tick command ordering, duplicate handling, replay, and checkpoints are deterministic.
- [x] Account, Character, and SessionPlayer identities are distinct and server-bound.
- [x] A client cannot choose another membership's authoritative identity.
- [x] Reconnect rotates credentials while restoring the same membership/character relationship.
- [x] Allowlisted private/public projections prevent raw or other-player private ECS leakage.
- [x] Runtime package, Lua, configuration, capability, protocol, and mounted-asset identity are negotiated before admission.
- [x] Replay/checkpoint participants pin that runtime identity.
- [x] Realm character leases prevent simultaneous authoritative use.
- [x] Durable commits require the active lease/revision and reject stale or replayed writes.

Evidence is concentrated in `internal/game/session`, `internal/app/gameserver`,
`internal/app/realm`, `internal/app/clientsession`, and their QUIC, recovery,
spoofing, privacy, lease, CAS, and process-worker tests.

### G2 — Pinned authoritative game-data generation

Status: **foundation complete; per-consumer linkage audit ongoing**.

Already true:

- [x] `RuntimeRecipe.AssetSetID` deterministically pins every externally mounted file by path-independent content digest.
- [x] Package, authoritative Lua, gameplay configuration, capability/API, and network identities are immutable for a session and cross admission, reconnect, worker allocation, replay, checkpoint, and durable compatibility boundaries.
- [x] Existing sessions cannot silently adopt another runtime recipe.
- [x] The layered VFS and existing record/store adapters remain the only content-loading authority.

Still required:

- [x] Introduce an explicit `GameDataGenerationID` in runtime-recipe v2 that
  binds mounted bytes to the authoritative record parser/schema contract.
- [x] Narrow its byte input from the presentation-inclusive mounted asset set to
  the effective authoritative `.txt` data paths and preserve their winning
  layer/path provenance.
- [x] Include the effective `AnimData.d2` binary in that same immutable
  generation now that its fixed-point records schedule gameplay action events.
- [x] Pin copied immutable record bytes for one live authority; invalidation can
  only reparse that generation, while source edits or mount changes create a
  different store and generation for a future authority.
- [x] Preserve MPQ case-insensitive table lookup after pinning while retaining
  the authored winning path/case in generation provenance.
- [x] Discover startup-critical `MonPreset.txt` and `SkillDesc.txt` hash-table
  members when a retail MPQ's incomplete `(listfile)` omits them, and pin them
  normally rather than letting labs, character admission, or servers bypass
  the immutable record generation.
- [x] Compose the policy-neutral authoritative data module as a default so the
  interactive client retains its presentation-profile-aware `engine.data/v1`
  override instead of failing lab startup on duplicate registration.
- [x] Carry the explicit generation through the canonical runtime identity and
  therefore session admission, replay, checkpoint, reconnect, worker allocation,
  and durable compatibility identity hashes.
- [ ] Preserve and expose row ordinal, symbolic ID, explicit numeric ID, act-local index, source provenance, and derived index as distinct concepts where consumed.
- [x] Add deterministic byte/provenance/change/presentation-exclusion and
  no-live-swap generation tests.
- [x] Add a policy-neutral cross-table linkage diagnostic fixture reporting row
  ordinal/source line, authored key, column/raw value, target, identity kind,
  severity, and pinned provenance without repairing authored data.
- [ ] Complete the per-consumer audit of symbolic ID, explicit numeric ID,
  row ordinal, act-local index, provenance, and derived lookup identities as
  each `d2legacy` table relationship is admitted.

Known behavior: current Dark Magic identity and pinning are verified. Exact
legacy TXT-to-BIN compilation/link behavior remains unresolved and must not be
claimed by this gate.

### G3 — Immutable session `GameRules`

Status: **partial; immutable authority and first consumers implemented**.

- [x] Add one immutable `d2legacy`-owned session rules value covering difficulty,
  the fixed expansion/1.14d ruleset, Hardcore, Ladder eligibility where 1.14d
  behavior distinguishes it, content generation, and explicit gameplay
  configuration.
- [x] Keep `maximum_players` as an admission-capacity fact only, and move the
  optional `/players X` gameplay override out of immutable `GameRules` into
  separate command-mutated checkpointed state.
- [x] Validate expansion-only rules at game/worker creation and bind them into
  runtime identity plus checkpointed authoritative state.
- [ ] Feed combat, monsters, loot/NoDrop, XP, quests, vendors, hirelings, states,
  death, portals, and endgame eligibility through purpose-specific rule queries.
- [ ] Remove scattered session-wide mode decisions as each consumer migrates.
- [x] Prove copied Lua reads, checkpoint restoration, runtime identity, and
  changed-rule rejection; admission/reconnect inherit the pinned identity.

Implemented consumers: authoritative player entry, entry-world generation,
Blood Moor population, and monster stat/XP/treasure-class interpretation must
agree with the immutable game difficulty.
Dedicated and Realm workers now generate their initial town and wilderness
from the same pinned difficulty later installed as `GameRules`. Remaining
domains migrate in their own evidence-backed slices.

`d2legacy.game_rules/v2` rejects the superseded immutable `player_count`
configuration. `maximum_players` is consulted when admitting a player but is
not a monster, reward, party-projection, or `/players X` scaling input. The
mutable gameplay override lives in separately revisioned and checkpointed
`d2legacy.player_count/v1` authority state.

Per-player durable difficulty/quest facts and initial-data fields already exist;
they are not a substitute for one immutable game-wide semantic context.

### G4 — Multiplayer player-count and party context

Status: **partial; party authority, party-aware NoDrop, and party UI projection
implemented; other reward consumers pending**.

- [x] Represent live present-player count, optional `/players X` override,
  effective gameplay count, nearby eligible count, and party reward eligibility
  as distinct contexts; joining/leaving updates the default live count while an
  explicit command forces the override until changed/cleared.
- [ ] Implement verified monster-life, monster-XP, NoDrop, and difficulty consumers.
- [x] Implement authoritative invite, cancel, accept, leave, membership,
  identity, game-departure cleanup, and reconnect party state.
- [x] Feed the credited player/owned-unit owner and living same-level party
  count into monster-death NoDrop policy.
- [ ] Extend same-level living-member queries with verified proximity, then add
  party XP, kill/owned-unit credit, quest credit, and gold sharing.
- [x] Project party state to UI; do not make the UI roster authoritative.

Monster spawn now pins `spawn_player_count` from the effective gameplay count
and applies the expansion 1.14d 50%-per-additional-player life and base-XP
bonuses. NoDrop distinguishes actual game population, effective gameplay
count, additional nearby party members, and the monster's spawn count; the
latter caps later drop benefits.
Blood Moor population is no longer created eagerly at startup: the authority
checkpoints the generated room plan, activates the room containing a player plus
its immediate graph neighbors, and pins the current all-player count when each
monster is materialized. Monster death resolves the credited player through
owned-unit attribution, counts their living party members in the same level,
and passes that context into NoDrop while retaining the spawn-count cap. Durable
death/event facts record each input and the final eligible count for replay
diagnostics. Broader level population and any narrower target-version proximity
rule remain open, so the combined consumer gate is not yet complete.

The population/override separation is now executable. With no override,
monster and NoDrop consumers count present authoritative player entities, so
entry and departure change subsequent behavior without mutating a setting.
Host-authorized `game.player_count.override` implements `/players X` semantics
from 1 through 8; `game.player_count.follow_population` clears it. The override
may be above or below live population and remains independent of a lower
admission cap. Both commands are deterministic and the separate state survives
checkpoint reconstruction. Integration coverage proves one -> two -> one live
players, an override of eight in a two-slot game, clearing back to live count,
and admission rejection at capacity.

Party relationships now live in one checkpointed `d2legacy.party/v1` state.
Authenticated player commands can invite/cancel/accept/leave without supplying
a party ID; acceptance allocates one stable authority-owned identity, and
explicit game departure removes invitations/membership while transport
reconnect preserves both. Living same-level member queries are available to
reward consumers, with NoDrop as the first integrated path. Checkpoint
continuation and live QUIC reconnect tests cover the state boundary. Exact
1.14d roster-event timing and NoDrop proximity details still require their
owned-runtime probes and are not inferred from older protocol behavior.

The authority now materializes a bounded `d2legacy.player.party_view` for each
player after policy evaluation. `ClientView/v5` selects only the authenticated
owner's versioned roster summary: player/name/class/level plus that owner's
relationship to each present player and only their own party ID. Other party
IDs and membership lists, invitation timestamps, coordinates, vitals, and raw
authority state are not projected. Offline and connected presentation read the
same component shape, and the party panel renders it without becoming a second
membership authority. Exact expansion 1.14d roster-event timing, location/
health visibility, and action-layout fidelity remain explicit UI probes rather
than compatibility claims.

Party XP remains probe-gated. Blizzard's expansion documentation establishes
same-area and roughly two-screen eligibility, a 35% party-pool increase,
raw shares directly weighted by character level, and a subsequent player/
monster-level penalty, but does
not specify the exact expansion 1.14d distance threshold or integer rounding.
`party_xp_probe` now rejects non-1.14d/community captures, validates paired
neutral/party observations with identical rosters and monster context, and
normalizes deltas plus direct/inverse/equal share hypotheses and candidate pool
rounding. No party-XP gameplay formula
lands until sanitized owned-runtime vectors resolve those remaining choices.

### G5 — Locomotion and motion-state foundation

Status: **partial**.

Deterministic pointer-first A*, level collision, prediction-compatible movement,
and monster chase exist. Still required:

- [x] Make direct-start gameplay fixtures activate the ordinary offline Session
  and route gameplay input through wrapper scenes instead of leaving animation-
  only intents stranded while authority remains in frontend mode.
- [x] Replace Warp Lab's private actor, route state, locomotion system, and
  direct portal teleport with production game-world movement, collision,
  interaction admission, shared relocation, animation, camera, and world
  presentation; retain only read-only diagnostics and masking in the lab. Pin
  both arrival anchors against the full player footprint and prove round-trip
  travel followed by fresh locomotion. Cross-level presentation activation now
  invalidates the old world-relative target/path/selection state and snaps
  camera interpolation before accepting pointer coordinates in the new map;
  the acceptance deliberately queues a stale return-side route and proves it
  cannot retain motion ownership in town.
- [x] Gate click-to-operate on authoritative route completion and treat stale
  mutable target/range observations as rejected actions rather than fatal
  session errors; cover the actual point-click ordering in the owned-data lab.

- [x] Replace placeholder walk/run rates with one immutable, case-insensitive
  `CharStats.txt` class catalog shared by authority and client prediction; pin
  all seven Expansion 1.14d classes to the owned-data 6/9 walk/run records.
- [x] Make current/max stamina live authoritative 8.8 state admitted from the
  durable character, project and persist it through the owner-private HUD, and
  share the same exact raw values with connected prediction.
- [x] Implement the pinned `CharStats.txt` RunDrain cadence, wilderness-only
  running drain, torso-armor drain multiplier, slower-drain and recovery stat
  channels, idle/walking/town recovery, zero-stamina walk fallback, and generic
  `item_fastermovevelocity` diminishing returns. The owned Expansion 1.14d
  archive pins every class's starting stamina, RunDrain, per-level/per-Vitality
  terms, ItemStatCost identities, and `move1`/`move2`/`move3` Properties links.
- [x] Recompute authoritative 8.8 maximum stamina from the pinned class starting
  Vitality/stamina facts, quarter-unit per-level and per-Vitality terms, direct
  `maxstamina`, bonus Vitality, active/passive skill-percent, and item-per-level
  ItemStatCost operands. Durable Vitality now survives admission as a live
  stamina progression fact; equipment and generic sources share the same graph.
  Max-source changes preserve positive current stamina proportionally with the
  recovered double/truncate/clamp callback, zero remains zero, and level-up
  explicitly fills the new derived maximum. Owned Expansion 1.14d tests pin the
  relevant ItemStatCost operations and `stam`/`stam/lvl` Properties links.
- [x] Add a checkpointed per-act environment cycle and the Properties func 18
  signed packed-min/max evaluator required by ItemStatCost op 6
  `item_stamina_bytime`. The high-confidence recovered 360-unit cycle preserves
  normal, Act III night, and Act IV cadence, 15-unit rounding, wraparound, and
  linear center/opposite interpolation. Owned Expansion 1.14d records pin stat
  ID 295, op 6, its `maxstamina` dependency, and the `stam/time` property.
  Source changes flow through the same proportional max-resource callback and
  checkpoint/replay boundary as every other maximum-stamina operand.
- [ ] Pin stat-allocation/max-callback ordering before exposing a live base-
  Vitality allocation command. Also connect the Act II Tainted Sun quest to
  the existing eclipse cycle facts only after its target-runtime transition is
  captured. These are explicit holdouts, not permission to trust admitted
  redundant max-resource fields.
- [x] Centralize the high-confidence recovered movement order: item Faster
  Run/Walk receives its 150-point diminishing conversion, then joins skill,
  state, and equipped armor/shield `velocitypercent` sources before the final
  25% floor. Authority and prediction consume the same integer percentage.
  Owned Expansion 1.14d records pin representative zero/five/ten Armor.txt
  penalties across torso armor and shields; equipment tests prove independent
  pieces stack. A generic timed `cold` source proves the recovered player
  `-50` movement effect orders with skill/armor/item sources, checkpoints, and
  expires without introducing skill-specific policy.
- [ ] Capture owned Expansion 1.14d runtime vectors for extreme negative and
  positive movement modifiers, cold/freeze target conversion, resistance,
  Cannot Be Frozen, Half Freeze Duration, difficulty divisors, and the paired
  `attackrate`/`other_animrate` effects before enabling broad cold/freeze
  content. Recovered executable structure is high-confidence evidence, not a
  substitute for those target-runtime boundaries.
- [x] Separate route planning from authoritative motion execution state. The
  client retains only replaceable world-scoped waypoints; admitted locomotion
  and melee approach now claim one checkpointed `d2legacy.player.motion` fact,
  and one ordered authority motion boundary derives player velocity, WL/RN mode,
  class/stat speed, and exhaustion correction. Explicit locomotion
  replaces attack approach, exhaustion downgrades the same fact, and relocation
  clears ownership instead of relying on zero velocity as an implicit signal.
- [x] Keep presentation playback state separate from authoritative distance
  integration while sharing the stat-derived effective velocity percentage.
  Expansion WL/RN runtime bases 213/101 are scaled by the same walk/run
  percentage that drives path velocity; `AnimData.d2` still owns frame count
  and events. Local and revisioned public player projections carry class,
  `velocitypercent`, and item FRW, and retained playback preserves frame phase
  across mid-mode rate changes. Regressions prove raw displacement cannot alter
  cadence, while FRW/chill can, without restarting the animation.

### G6 — Dynamic occupancy and knockback

Status: **partial**.

- [x] Separate unit footprint radius from an explicit `blocks_movement`
  occupancy policy. Players and living monsters opt in; monster death already
  removes the collider, and inactive room residents retain the policy with the
  rest of their checkpointed ECS graph while active-system queries exclude it.
- [x] Resolve same-level multi-unit motion contention in stable ECS order using
  current plus already-committed positions. Axis-separated static collision and
  dynamic circle footprints compose without renderer geometry; simultaneous
  contenders cannot swap or enter the same occupied space, and an explicit
  nonblocking unit remains passable. Admission and warp anchors may begin in a
  temporary overlap, so movement that strictly increases separation is allowed
  while entry or deeper overlap remains blocked. Checkpoint parity pins both
  decisions.
- [x] Pin owned Expansion 1.14d `MonStats2` knockback-mode and small/normal/large
  target facts, including representative capable small/large monsters and a
  mode-incapable normal monster. Spawned and inactive monsters retain
  the resulting generic target profile; the owned `knock` property and
  `item_knockback` melee/missile event hooks are pinned without guessing their
  binary-owned chance arithmetic.
- [x] Preserve the owned Expansion 1.14d `Missiles.KnockBack` byte in generic
  straight-missile definitions and checkpointed projectile instances. Blank,
  `1`, `33`, and `75` representative rows are pinned; combat does not interpret
  the byte or emit forced motion until the target binary's roll/result policy
  is verified.
- [ ] Verify remaining target-runtime category rules for players, hirelings,
  summons, NPCs, and corpses, then decide which categories participate in A*
  planning versus only fixed-tick motion resolution.
- [x] Add a semantic forced-motion request resolved by movement/collision
  authority. Direction is derived away from the supplied source, the request's
  policy-owned distance and speed advance over fixed ticks, active progress is
  checkpointed, and durable semantic outcomes distinguish completed, partial,
  blocked, and invalid transactions. Presentation can observe the event but
  cannot move the target.
- [x] Emit stable selectable `player:`/`monster:` target IDs for invalid,
  replaced, blocked, partial, and completed forced-motion events, retaining an
  `entity:` fallback only for internal non-selectable movers.
- [x] Add a strict `diablo-ii-lod-1.14d-expansion` owned-runtime knockback probe
  contract covering target category/record/size/KB mode, matched controls,
  item/missile mechanisms, open/collision-limited displacement, lethal/block/
  uninterruptible exclusions, reactions, and confidence intervals. Older
  recovered size-weighted, raw-byte-percent, five-unit, mode-fallback, and
  velocity findings remain labeled candidates and do not drive combat.
- [ ] Recover and pin remaining Expansion 1.14d knockback chance, distance,
  speed, player/owned-unit eligibility, interruption, and GH/KB mode rules
  before combat emits the generic request. Older recovered server/client path
  code is structural evidence only, not permission to copy its constants into
  the target policy.
- [ ] Cover remaining competing forced-motion requests and target-runtime
  interaction with active skills, hit recovery, death, and client playback. A
  newer admitted request now deterministically replaces one active transaction,
  emits the displaced transaction's exact `replaced` progress outcome, and owns
  velocity until completion; only fresh subsequent locomotion moves the target.

### G7 — Active-room/inactive-unit vertical slice

Status: **mechanism breadth complete; exact Expansion 1.14d activation/timer
policy remains probe-gated**.

- [x] Separate authoritative world existence from active simulation and
  presentation residency for one ordinary monster. An empty ECS inactive tag
  filters Lua systems and local/remote projections; the engine movement opt-in
  tag is removed only while inactive.
- [x] Preserve one ordinary monster's stable ECS/semantic identity and current
  component-owned stats, combat profile, appearance, AI/action, death, motion,
  location, collision, and selection state without an allowlisted scalar copy.
- [x] Preserve cross-entity timed-state instance, stat-source, and state-event
  target references through deactivate -> checkpoint -> restore -> reactivate.
  The referenced monster entity ID does not change.
- [x] Replace the population-specific room marker with
  `d2legacy.world.room_resident {id, level_id, room_id}`. Plan records use the stable
  resident ID and remember whether the engine velocity-mover opt-in existed;
  a non-monster resident checkpoints/reactivates without acquiring movement.
- [x] Canonicalize generated room/link IDs as strings, scope activation by level
  plus room, and attach production DS1 interaction targets by testing their
  authoritative subtile point against generated room bounds. Same-named rooms
  in another level remain active; missing room geometry does not invent a link.
- [x] Attach Warp Lab's paired authoritative warp entities to generated rooms
  through the same entry-world point resolver. The endpoint stays one ordinary
  interaction/transition entity; residency adds no warp-specific lifecycle.
- [x] Synchronize every active positioned resident to its current generated
  room before activation decisions. A moving non-monster crosses from room A to
  B, then remains active when A leaves the player window and B remains active.
- [x] Preserve an owned resident's authoritative owner entity reference,
  category/limit/lifetime policy, durable identity, and immediate/ultimate
  attribution across deactivate -> checkpoint -> restore -> reactivate. Its
  lifecycle query uses the same empty inactive marker rather than a second
  summon archive.
- [x] Preserve an ordinary corpse's stable entity, death/loot facts, monster
  identity, appearance, position, and room membership through inactive
  checkpoint/reactivation. Death removes AI, collider, selection, and the
  generic velocity-mover opt-in, so reactivation restores no live capability.
- [x] Attach a production-cast straight projectile to the installed room plan
  with a deterministic world-owned resident ID. The common inactive ECS tag
  suspends its movement and lifetime fields, and the same entity/component
  state survives deactivate -> checkpoint -> restore -> reactivate.
- [x] Attach an imported ground-item placement through a generic pre-plan ECS
  room request. Inactivation filters local/private item projections; the item
  survives checkpoint/reactivation on the same entity, pickup removes world
  components, and re-drop resolves residency without an item archive.
- [x] Preserve a stateful interaction object's mode/used/seed/revision facts
  plus a separately resident pending-action entity and its raw object reference
  through inactive checkpoint/reactivation. The admitted one-shot family is
  synthetic mechanism evidence, not a retail Objects.txt mapping.
- [x] Drive initial Blood Moor population activation from a deterministic
  all-player room graph.
- [x] Reproduce first-activation transitions through replay/checkpoint.
- [x] Reproduce deactivate -> checkpoint -> restore -> reactivate continuation
  with the same authoritative checksum.

The checkpointed `d2legacy.population.plan/v5` stores a deterministic active
flag and stable inactive resident records per room. A generated monster carries
the world-owned semantic room-resident component; leaving the occupied-room-
plus-neighbors window adds `d2legacy.world.inactive` and removes the generic
velocity-mover opt-in, when present, without destroying the entity. Simulation
queries, local monster
snapshots, and revisioned remote world views exclude the inactive tag. Re-entry
removes it and restores movement opt-in on the same entity, so ECS checkpointing
retains every component and raw relationship reference without a second archive
schema. The acceptance fixture crosses a three-room graph, proves AI does not
advance while inactive, checkpoints/reconstructs the Lua runtime, preserves a
timed-state/stat-source/event graph and its entity IDs, preserves a second non-
monster/non-moving resident, proves an equal room ID in another level is not
affected, and reaches identical reactivation checksums. The generated monster
also carries a synthetic `d2legacy.owned_unit` relationship whose raw owner
entity, category, limit, lifetime flags, durable ID, and attribution survive
the same inactive checkpoint. `d2legacy.owned_unit.lifecycle` excludes the
empty inactive tag, then evaluates its unchanged absolute expiration on the
first active tick. This is deterministic scaffolding, not a claim about exact
Expansion 1.14d timer aging. A second acceptance path commits ordinary monster
death before leaving the room, preserves the corpse's semantic component set
through inactive checkpoint reconstruction, and proves the room plan records
`velocity_mover=false`. The death transaction now removes the generic mover
opt-in together with AI, collider, and selection; neither direct reactivation
nor restore may invent it. A third production path casts the generic straight-
missile family inside a generated room. Materialization assigns a deterministic
projectile resident ID through the population plan's canonical point resolver;
the ordinary room-sync and activation systems then move it across the same ECS
boundary. While inactive, the projectile movement query does not change its
position or remaining lifetime. Checkpoint reconstruction and reactivation
preserve the original entity and component state with checksum parity, without
introducing a missile-specific dormant record. The test's extended lifetime is
synthetic scaffolding and does not assert exact target timing. Entry-world
preparation joins resolved DS1 interaction points and synthetic paired Warp
Lab endpoints to
the same canonical residency contract; the mounted-asset lab proves both warps
materialize with resident components. Before each activation decision, active
positioned residents resolve against the same authoritative room bounds; an
entity crossing a boundary is no longer inactivated with its spawn room.
An imported item whose authoritative placement is `world` starts with generic
position/location plus `d2legacy.world.room_attach`; population-plan admission
resolves its stable `item:<owner>:<id>` identity through the same room geometry
and removes the transient request. The ordinary inactive marker then filters
both Lua and revisioned private item projections. Checkpoint reconstruction
preserves identity, placement, presentation, spatial, and residency components;
reactivation reaches checksum parity. Existing item movement owns the inverse
transition: pickup removes all world/inactive components, while a re-drop at an
authoritative player level/point resolves residency again. The fixture remains
player-layout-owned and synthetic, so it does not claim the unresolved public
ground-item ownership, loot materialization, legal-position search, pickup
range/path, reservation, allocation, or lifetime policies.
Finally, a synthetic object definition opts into the existing sorted component-
family interaction dispatch. Its one-shot handler commits mode, used, and
revision facts on the object entity; a pending-action entity keeps due-tick,
sequence, kind, active flag, and a raw ECS reference to that object. Both carry
their own stable resident IDs because room activation queries must not depend on
an implicit relationship traversal. Deactivate,
checkpoint reconstruction, reactivation, and repeat interaction preserve the
entire graph and checksum without an object archive. The pending action is not
executed in this slice: exact Expansion 1.14d operation/event function mapping,
mode timing, collision/selectability changes, delayed execution, and inactive
event aging remain unresolved.
This is Dark Magic semantic state, not a vanilla save/protocol compatibility
structure.

Exact expansion 1.14d activation distance/tick ordering, long-inactive healing,
timer aging while inactive, corpse lifetime/usability, projectile lifetime,
retail stateful-object operation/event families and scheduling,
public ground-item generation/ownership/drop/pickup/lifetime, broader
generated-level coverage, and independent visible-
but-not-simulated presentation residency remain open and probe-gated. Older
recovered inactive-unit code is architectural evidence only.

## P1: strengthen and complete the first multiplayer gameplay loop

### G8 — Combat fidelity tranche 1

Status: **partial**. One Lua-owned melee/missile damage path, timed states,
death transaction, fixed-point vocabulary, deterministic vectors, and an
explicit shared direct-damage result record exist.

- [x] Route successful melee and straight-missile contact through one Lua-owned
  health-mutation boundary that reports channel, rolled raw damage, damage
  actually committed after mitigation/storage quantization, remaining raw
  health, and lethality.
- [x] Compose a successful melee result as one ECS entity carrying both
  `d2legacy.combat.event` and `d2legacy.combat.melee_event`. Generic death/event
  consumers and melee reaction consumers therefore observe one authoritative
  fact without source-specific joins or duplicate event entities.
- [x] Keep misses and invalidated melee impacts as melee-resolution facts with
  no generic damage component, while missile contact emits the same generic
  ordered result vocabulary with `source_kind=missile`.
- [x] Quantize committed player damage once at the current whole-health
  component boundary so event output matches durable ECS state. This is an
  internal-consistency rule, not a verified Expansion 1.14d rounding claim.
- [x] Compose a typed ECS damage-bundle fact on each successful result. Preserve
  physical, fire, lightning, cold, magic, and poison rolled/mitigated values
  independently through per-channel mitigation, summing only immediate channels
  at the health commit. Poison is retained but explicitly excluded until its
  periodic rate/duration transaction exists. Existing scalar totals remain a
  convenient generic event summary.
- [ ] Extend the bundle and ordered transaction with verified drain, cold/
  freeze, poison-duration, conversion, and periodic-application facts without
  treating them as immediate health damage.
- [ ] Add target-locked Expansion 1.14d evidence and ordered stages for block,
  avoidance, resistance caps/negative values, pierce, absorb, critical/deadly/
  mastery, Crushing Blow, Open Wounds, leech, hit recovery, poison/periodic
  damage, and durability.
- [ ] Complete ordinary softcore player death, corpse, gold, and XP semantics
  before Hardcore durable death or broad PvP.

Implement one shared ordered transaction for chance-to-hit, block, avoidance,
physical/elemental/magic mitigation, caps/negative resistance, pierce, absorb,
critical/deadly/mastery, Crushing Blow, Open Wounds, leech, hit recovery,
knockback, poison/periodic damage, and durability. Keep unresolved arithmetic
labeled and probe-driven. Then complete ordinary player death/corpse/gold/XP
semantics before Hardcore durable death or broad PvP.

### G9 — Skill/state/missile behavior-family coverage

Status: **partial**. Generic cast lifecycle, timed state, melee, and straight-
missile behavior families plus supporting target/motion primitives exist. Fire
Bolt is the first explicitly supported expansion 1.14d straight-missile
configuration; it no longer owns a standalone component, command branch,
system, damage function, random stream, or private admission list. Exact skill
admission now comes from one target-locked implementation manifest shared by
runtime composition and the coverage report.

- [x] Generate a mounted-data report of server start/do behavior IDs, consuming
  skills, implementation family, missing family, and evidence status.
- [x] Generate exact skill-investigation evidence by joining Skills.txt and
  SkillDesc.txt to base/Expansion/patch locale TBL keys, winning text source,
  replacement tokens, and resolved `.blvl`/`.lvl` cross-skill references.
- [x] Replace the first skill-specific Fire Bolt authority with an explicitly
  configured, definition-driven straight-missile family dispatched by skill ID.
- [x] Replace Attack's skill-ID-zero command branch with an exact-ID,
  definition-driven `action.melee` family routed through the same learned-skill,
  resource, cast, effect, and completion lifecycle as mana-costing skills.
- [x] Replace fallback base melee impact/completion ticks with the definition's
  animation mode plus the actor/weapon composite's pinned AnimData fixed-point
  attack event and cursor-wrap schedule.
- [x] Route `UseAttackRate` timing through one reusable action-rate policy fed by
  authored `attackrate`, `item_fasterattackrate`, weapon-speed, dual-wield, and
  pinned AnimData facts rather than an Attack-specific delay table.
- [x] Centralize current PvE melee target legality and revalidate semantic ID,
  player/hostile alignment, living state, act/level, footprint reach, and the
  authored melee-barrier trace both before Attack animation and at impact.
- [x] Prove the definition decoder handles multiple authored configurations
  without skill-name/ID branches; keep the second configuration synthetic so it
  does not claim incomplete behavior for another retail skill.
- [x] Replace the provisional name-selected self-state placeholder with a
  definition-driven timed self-state/stat-source family: shared cast/mana,
  level and hard-point-synergy formulas, refresh/expiration, checkpoint, and
  exact source removal.
- [x] Make ordinary mana admission a shared cast-lifecycle invariant:
  underfunded requests start no action or effect and preserve the partial mana
  balance; lock the rejection path with deterministic executable coverage.
- [x] Add the generic state-group replacement and successful-melee-hit reaction
  mechanisms used by Frozen Armor, including exact source removal, row-derived
  freeze length/synergies, expansion difficulty divisors, checkpointing, and
  monster action suppression.
- [ ] Complete Frozen Armor's PvP chill conversion, target resistance/immunity
  and monster-class duration modifiers, exact integer/tick ordering, animation
  action timing, and presentation before upgrading it from partial behavior.
- [ ] Implement reusable targeted, point, self, area/nova, buff/debuff/curse/aura,
  summon, corpse, movement, missile, and trap families in dependency order.
- [ ] Use representative skills as fixtures; do not implement seven trees independently.

`skill_behavior_coverage` mounts owned archives, reads the winning Expansion
1.14d Skills.txt and Missiles.txt tables, groups every skill by server start/do
and referenced missile server-do function IDs, and reports every consumer with
its explicit family, missing-family flag, and evidence status. The current
owned-data baseline is 357 skill rows, 172 signatures, 3 explicitly admitted
configurations, and 354 missing configurations. The report fails if a declared
skill or referenced server missile is absent, and its synthetic test proves a
row with the same function signature is not admitted by resemblance. Generated
reports remain local; copyrighted tables are never copied into Git.

`skill_evidence` is the required companion investigation report for skill-tree
synergies and skills that modify other skills. It accepts exact IDs and a locale,
joins each Skills.txt row to SkillDesc.txt, resolves every name/description/
tooltip label through layered `string.tbl`, `expansionstring.tbl`, and
`patchstring.tbl`, records replacement tokens such as `%s`, and parses every
`skill('name'.blvl|lvl)` formula back to the referenced target skill ID. Missing
keys and unknown skill references fail or remain explicit instead of silently
dropping documentation. Fire Bolt now reports Fire Ball/Meteor hard-level fire-
damage synergies; Frozen Armor reports Shiver/Chilling Armor hard-level duration
and freeze-length modifiers in both gameplay and tooltip formulas. TBL wording
establishes intended relationships and player-visible claims; Skills.txt calc/
parameter fields and owned 1.14d runtime probes remain authoritative for exact
arithmetic, rounding, and ordering. The corrected version-1 TBL codec and a
layer-precedence/source test make this evidence path executable.

`manifests/skill-behavior-coverage.v1.json` is locked to
`diablo-ii-lod-1.14d-expansion`. Runtime composition consumes the same exact-ID
declarations as the report. `d2legacy.data.missile_skills` validates an admitted
row pair into an immutable `missile.straight` definition; the earlier Frozen
Armor name lookup is now the generic `state.self-timed-stat` decoder selected
by ID. It validates server function 18, `frozenarmor`,
`skill_armor_percent`, `ln12`, and the `ln34 + hard-level synergies * par7`
duration shape. The owned target row and official Blizzard table produce 7
mana, 30% + 5% per level defense, 3000 + 300 frames per level duration, and
250 frames per Shiver/Chilling Armor hard point. A source-tagged state and
`defense` percentage source apply, refresh, survive checkpoint, expire, and are
removed together. The decoder now also validates `damagedinmelee` event function
2, the Param5/6 freeze-length formula, Param8 hard-point synergies, the target
`freeze` state, and the armor state's owned States.txt group. A successful melee
hit applies that source-tagged freeze to a monster attacker for the row-derived
duration (30 + 3 frames per skill level, then +5% per Shiver/Chilling Armor hard
point) and suppresses its actions until expiration. Normal/Nightmare/Hell use
the official full/half/quarter cold-length relationship. Applying another state
in the same authored group removes the displaced instance and its exact stat
source. Frozen Armor remains partial because PvP must chill rather than freeze;
target cold resistance/immunity, monster-class modifiers, exact rounding/tick
ordering, animation timing, and presentation are not yet implemented.

Ordinary Attack is now the first `action.melee` configuration rather than an
exception outside the skill system. Its exact Expansion 1.14d Skills.txt row
must declare ID 0, server/client start and do functions 1/1, the A1 weapon
action, attack-rate and target/search flags, weapon source damage, and zero mana
before the decoder constructs its immutable definition. The mode is carried
through the generic cast event and combined with the actor token and equipped
weapon class. The session-pinned `AnimData.d2` record then supplies its 24.8
rate, frame count, and typed event bytes through renderer-free
`engine.animdata/v1`; Lua schedules the first attack marker and cursor wrap at
integer simulation ticks. The owned unarmed Amazon `AMA1HTH` record is 13
frames at rate 256 with its attack event on frame 8, producing impact/completion
delays of 8/13 ticks at the unmodified 100% rate. `player.use_skill`
creates the same generic cast request used by Fire Bolt and Frozen Armor; the
shared lifecycle verifies the authoritative learned level and accepts the
literal zero cost, then a family adapter emits the reusable approach, selected-
hand, animation, and impact action. No command, component, or system branches
on Attack's ID or name, and a synthetic second-row decoder test proves family
reuse without claiming another retail melee skill.

The reusable action-rate layer now resolves the owned ItemStatCost names
`attackrate` (signed ID 68, `UpdateAnimRate=1`) and
`item_fasterattackrate` (signed ID 93) from named stat sources. Equipped weapon
base speed contributes with the authored inverse sign, so the owned Expansion
rows Phase Blade `speed=-30` and War Pike `speed=20` become +30 and -20
`attackrate`. Owned Properties.txt maps `swing1`, `swing2`, and `swing3`
directly to `item_fasterattackrate`. One skill-agnostic policy applies integer
effective IAS `120*IAS/(IAS+120)`, primary/secondary weapon averaging for dual
wield, the 15%-175% rate bounds, integer effective AnimData speed, and the same
fixed-point marker/wrap scheduler. Equipment and passive/skill facts enter the
existing provenance-preserving source resolver; equipping, swapping, or
removing the source updates subsequent actions without an Attack-only branch.
An admitted action snapshots its impact and completion ticks, so checkpoint and
replay do not depend on later presentation playback.

The table identities, property mapping, weapon values, and owned
`AMA1HTH` record are verified directly against the pinned Expansion 1.14d
generation. The arithmetic and dual-wield structure are high-confidence
recovered behavior whose exact 1.14d breakpoint, sequence-action, shapeshift,
slow/chill, and mid-swing stat-change boundaries still require owned runtime
vectors. No older-version branch or compatibility mode exists.

The generic melee target service now treats command target IDs as untrusted
requests and re-resolves current ECS facts. Player attacks require a living
`hostile` target; monster attacks require a living `player`; both require the
same act and level. Named Attack rechecks those facts before beginning its
animation, while named and targetless Shift-Attack resolution rechecks them,
the selected hand's reach, and the current level collision at the AnimData
impact tick. Movement and combat share one per-level immutable collision-map
registry. The engine exposes visual `BlockLOS` and the distinct DT1
flying/melee-barrier (`BlockJump`) ray trace separately, avoiding a policy bug
where opaque tiles would automatically become melee walls. This is a
high-confidence structural recovery, not an exact 1.14d completion claim:
current continuous footprint-distance arithmetic, dynamic door collision,
PvP hostility, special unit range exceptions, and path-to-range behavior still
need owned target-version probes. Therefore this does not admit Bash, Jab, or
any other superficially similar row.

The shared lifecycle rejects a mana-costing skill before creating its cast when
the authoritative 8.8 fixed-point balance is below the computed cost. It
consumes the request, emits no effect, deals no damage, and leaves mana
unchanged. This follows Blizzard's expansion documentation that a skill is
unusable for lack of mana and that mana is consumed when a skill is used; exact
cost formula/rounding and interruption/refund edges remain target probes.

Fire Bolt has owned-target record evidence. Ice Bolt and other visually or
structurally similar skills remain missing until their own Expansion 1.14d
launch, motion, impact, state, and ordering semantics are verified.

Next: probe and replace Attack's remaining inferred distance, dynamic-door,
special-unit, and path-to-range edges and confirm its attack-rate breakpoint,
dual-wield, slow, sequence, and mid-action boundaries against owned 1.14d
runtime vectors. In evidence order, finish Frozen Armor's remaining target-sensitive
cold-duration/PvP rules. Then use the report to select one high-leverage missing
target/point/area signature. Evidence upgrades and exact-ID declarations land
together; no declaration is added merely because another skill shares server
function IDs. Synergy and every skill-that-modifies-another-skill investigation
must begin with the joined locale TBL keys/text/replacement-token evidence,
then bind those player-visible relationships to Skills.txt formulas and owned
1.14d runtime vectors before implementation.

### G10 — Item-source lifecycle

Status: **partial**. Provenance-preserving stat sources, authoritative containers,
equipment transactions, generation, sockets, and runeword recognition foundations
exist. Active equipment now projects weapon `attackrate` and authored
`item_fasterattackrate` into generic action timing and removes them on unequip.

Activate ordinary equipment, weapon swap, broken/requirements suppression,
charms by container, socket children, gems/runes/jewels, runewords, set thresholds,
item skills/charges/auras/procs. Moving or removing an item must remove exactly
that item's sources.

### G11 — Kill-to-ground-item acceptance

Status: **partial**. Monster death emits deterministic loot facts through the
existing M6 generator; the complete authoritative world-item/pickup/equip loop is absent.

Complete death -> TreasureClass -> quality/item/affixes/properties -> ground
item -> eligibility/visibility -> pickup -> inventory -> equip/use -> gameplay
change, including player-count NoDrop, MF/Gold Find, owned-unit attribution,
deterministic placement/cleanup, multiplayer privacy, and replay/checkpoint parity.

### G12 — World-object primitive set

Status: **partial foundation; representative behaviors missing**.

Implement one reusable door, chest, shrine, waypoint, and Town Portal authority.
Object operation must commit state/collision/loot/effect/travel first; animation,
audio, and UI observe committed semantic state.

- [x] Add component-dispatched world-object operation, paired warp endpoints,
  and one relocation transaction shared with authored level seams; prove the
  pair through interaction admission, production locomotion, checkpoint restore,
  bidirectional active-world switching, footprint-safe arrival, post-return
  locomotion, old-world route/selection invalidation, camera discontinuity, and
  the ordinary game renderer in Warp Lab.
- [ ] Pin and implement expansion 1.14d Town Portal creation, owner/party access,
  replacement, lifetime, origin/return placement, and teardown behavior before
  treating the development pair as the Town Portal gameplay feature.

### G13 — Monster quality/pack/boss framework

Status: **partial**. Ordinary data-derived hostile materialization, population,
AI, combat, death, corpse facts, and owned-unit attribution exist.

Add groups/packs, champions, uniques/modifiers, minions, SuperUniques, reusable
AI families, corpse resurrection eligibility, special death behavior, and an
encounter-controller primitive before bespoke bosses proliferate.

### G14 — Complete Act I

Status: **not started as a campaign acceptance target**.

Complete Den of Evil, Blood Raven, Cain, Countess, Horadric Malus/Charsi,
Andariel, and the Act II transition. The slice must exercise multiplayer quest
credit, NPC/dialogue state, quest items, objects/portals, bosses, rewards,
services, persistent quest bits, waypoints, and transition without quest-specific
subsystem duplication.

## P2: durability, economy, and breadth

### G15 — Durable character semantic model

Status: **partial**. Identity, revision/lease safety, profile/Realm storage, and
some player/quest/item projection exist. Complete canonical base stats,
allocation, skills, inventory/equipment/swap/stash, corpse, per-difficulty
quest/waypoint/completion, hireling, and other proven durable facts. Keep transient
checkpoint state separate.

### G16 — Legacy interoperability

Status: **out of scope**. Dark Magic will not import, export, or preserve vanilla
`.d2s` files; speak BNCS, MCP, or D2GS; interoperate with vanilla servers; or
preserve compatibility with old community tools. Its canonical content,
network, replay, checkpoint, and durable character formats are independent.

### G17 — Trade

Status: **not started**. Implement request -> open -> offer mutation -> dual
acceptance -> reset on mutation -> final revalidation -> atomic exchange, with
disconnect/cancel restoration and item/socket/gold/stale-state validation.

### G18 — Hirelings and broad owned-unit behavior

Status: **partial foundation**. Generic ownership and attribution exist; full
hire/reward, generated identity, stats/skills/equipment, AI/follow, XP/death/
resurrection, transition, and persistence behavior remain.

### G19 — Cube, vendors, and economy completeness

Status: **partial foundation**. Authoritative containers, vendor placement,
basic buy/sell/services, and item generation exist. Add declarative atomic Cube
matching/transformations and verified stock, quotes, buyback, repair, recharge,
gambling, identify, heal, hire, and resurrect services.

### G20 — Campaign, class, UI, audio, and content breadth

Status: **partial foundations**.

After G1-G19 are stable, complete Acts I-V and all seven expansion class trees.
Make Normal broadly playable before perfecting Nightmare/Hell edges, while every
system consumes shared GameRules from the start. Incrementally add revisioned,
privacy-filtered semantic projections and event-driven audio throughout earlier
gates; do not postpone all UI/audio work to this gate.

## Parallel verification queue

Implementation and empirical verification proceed independently. High-value
probes remain:

- foundation: explicit content generation, live invalidation, cross-table links,
  ItemStatCost operations, and CharStats vectors;
- combat/motion: block, avoidance, mitigation, absorb, critical/deadly/mastery,
  Crushing Blow, Open Wounds, poison, leech, hit recovery, durability, PvP,
  attack-rate breakpoints/dual wield/mid-action changes, cast timing, path
  types, Tainted Sun environment activation, base-Vitality allocation/max-
  callback ordering, owned-runtime cold/freeze boundaries, and inactive rooms;
- items/economy: NoDrop, MF, runewords, charms, sockets, Cube operations, and pricing;
- world: object operations, doors, chests, shrines, warps, waypoints, portals,
  quest dialogue, difficulty consumers, and endgame eligibility;
- multiplayer: execute the version-locked party-XP distance/rounding matrix,
  then party XP/quest credit, hostility, trade, interest management, reconnect,
  and PvP.

Every probe targets expansion 1.14d and records owned-data/runtime setup, action
sequence, normalized observations, timing/RNG context, confidence upgrade, and
safe executable fixtures. Earlier-patch or Classic observations may explain a
source conflict but are not implementation requirements. Proprietary captures
and credentials do not enter Git.

## Explicit deferrals

- Vanilla client/server protocols, vanilla save files, and old community-tool
  interoperability are permanently outside the supported product boundary.
- Exact retail seed/layout reproduction is optional until explicitly targeted.
- Classic-mode and pre-1.14d compatibility branches are out of scope.
- Features not present in expansion 1.14d must not be back-projected into the
  target ruleset.
- Modern shaders, HRTF, upgraded rendering, and optional modern UI do not block
  authoritative Diablo behavior.
- Cloud deployment and generated creature representation do not displace G1-G20.

## Delivery policy

For each gate:

1. inspect current `main` and the relevant existing authority;
2. read the research baseline and verification queue;
3. state verified, recovered, inferred, synthetic, and unresolved behavior;
4. keep one primary behavioral objective per PR where practical;
5. add deterministic vectors plus replay/checkpoint coverage;
6. add multiplayer coverage when player count, privacy, identity, or ordering matters;
7. update this roadmap and affected research status only when acceptance is met.

Prefer a coherent final authority over wrappers around superseded systems.
