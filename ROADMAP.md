# Dark Magic roadmap

Status: fully refreshed through the G7 inactive-monster archive tranche on
2026-08-16.

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
  -> immutable content and GameRules govern the game
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
| M15 asset knowledge | partial | Typed/recovered coverage is broad; unresolved records and source-sensitive mappings remain research work. |
| M16 presentation primitives | partial | MPQ-backed render/audio primitives exist; remaining breadth is presentation fidelity, not a gameplay blocker. |
| M17 front end | foundation complete | The Lua-authored front end and Realm flow exist. Remaining polish belongs to UI fidelity. |
| M18 in-game shell | foundation complete | HUD and major overlay shells exist; the party panel now consumes an owner-scoped semantic projection, while remaining raw/ad hoc reads migrate as their gameplay domains mature. |
| M19 character/item/save fidelity | partial | Canonical profile and Realm character persistence exist; the complete Dark Magic durable semantic character does not. Vanilla save interoperability is out of scope. |
| M20 world fidelity | partial | Deterministic Act I generation, collision, transitions, population, and the first inactive-monster archive/restore path exist; dynamic occupancy, complete inactive entity graphs, object authority, and campaign breadth remain. |
| M21 Diablo simulation | foundation complete | Lua owns the current player, monster, skill, missile, state, death, loot, quest, item, and owned-unit vertical slices. Combat, movement, item activation, object, and content breadth remain below. |
| M22 networking | complete | One `Session`, authenticated semantic commands, deterministic ordering, filtered views, reconnect, replay/checkpoint, direct/listen/dedicated/Realm modes, and impairment/soak coverage exist. |
| M23 Realm/persistence | partial | Accounts, characters, leases, CAS commits, allocation, admission, reconnect, checkpoints, PostgreSQL, mail, and process workers exist. Publication/revocation, complete durable character semantics, and production operations remain. |
| M24 packaging/release | partial | Build/release foundations exist; the gameplay acceptance loop and final supported-platform release gate are not complete. |
| M25-M30 performance/UI/architecture | partial | Major residency, profiling, Lua-policy migration, and archetype ECS work landed. Remaining tasks are folded into projections, presentation, cleanup, and gameplay consumers below. |
| M31-M43 creature authoring | deferred | Generated creature representation is independent work and must not displace the gameplay critical path. |
| M44 Realm cloud operations | deferred | Local topology-neutral Realm is the prerequisite. Existing deployment groundwork does not make cloud operations a gameplay gate. |

The old milestone numbering is retained only as historical orientation. New work
uses the ordered gameplay gates below. This avoids preserving an obsolete plan
in which networking followed the first gameplay loop.

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
- [x] Pin copied immutable record bytes for one live authority; invalidation can
  only reparse that generation, while source edits or mount changes create a
  different store and generation for a future authority.
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
  behavior distinguishes it, player-count context, content generation, and
  explicit gameplay configuration.
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

Per-player durable difficulty/quest facts and initial-data fields already exist;
they are not a substitute for one immutable game-wide semantic context.

### G4 — Multiplayer player-count and party context

Status: **partial; party authority, party-aware NoDrop, and party UI projection
implemented; other reward consumers pending**.

- [ ] Represent game player count, effective `/players X` count, nearby eligible
  count, and party reward eligibility as distinct contexts.
- [ ] Implement verified monster-life, monster-XP, NoDrop, and difficulty consumers.
- [x] Implement authoritative invite, cancel, accept, leave, membership,
  identity, game-departure cleanup, and reconnect party state.
- [x] Feed the credited player/owned-unit owner and living same-level party
  count into monster-death NoDrop policy.
- [ ] Extend same-level living-member queries with verified proximity, then add
  party XP, kill/owned-unit credit, quest credit, and gold sharing.
- [x] Project party state to UI; do not make the UI roster authoritative.

Monster spawn now pins the effective player count and applies the expansion
1.14d 50%-per-additional-player life and base-XP bonuses. NoDrop distinguishes
actual game population, effective configured count, additional nearby party
members, and the monster's spawn count; the latter caps later drop benefits.
Blood Moor population is no longer created eagerly at startup: the authority
checkpoints the generated room plan, activates the room containing a player plus
its immediate graph neighbors, and pins the current all-player count when each
monster is materialized. Monster death resolves the credited player through
owned-unit attribution, counts their living party members in the same level,
and passes that context into NoDrop while retaining the spawn-count cap. Durable
death/event facts record each input and the final eligible count for replay
diagnostics. Broader level population and any narrower target-version proximity
rule remain open, so the combined consumer gate is not yet complete.

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

### G5 — Locomotion and motion-state foundation

Status: **partial**.

Deterministic pointer-first A*, level collision, prediction-compatible movement,
and monster chase exist. Still required:

- [ ] Replace placeholder walk/run rates with class/data-derived policy.
- [ ] Add Faster Run/Walk, stamina amount/max, drain, recovery, cannot-run, and
  verified chill/slow interactions.
- [ ] Separate route planning from authoritative motion execution state.
- [ ] Keep presentation animation rate separate from authoritative distance.

### G6 — Dynamic occupancy and knockback

Status: **not started**.

- [ ] Separate unit footprint from movement collision policy.
- [ ] Resolve multi-unit space contention deterministically without scheduler-race winners.
- [ ] Add a semantic knockback request resolved by movement/collision authority.
- [ ] Cover blocked, partial, competing, replay, and checkpoint cases.

### G7 — Active-room/inactive-unit vertical slice

Status: **partial; deterministic room activation and first ordinary-monster
archive/restore implemented**.

- [ ] Separate world existence, active simulation, inactive archive, and presentation residency.
- [x] Archive and restore one ordinary monster with stable semantic identity and
  its current component-owned stats, combat profile, appearance, AI/action,
  death, motion, location, collision, and selection state.
- [ ] Extend the archive to cross-entity timed states/events, owned-unit graphs,
  corpses, items, objects, and target references that cannot remain raw entity handles.
- [x] Drive initial Blood Moor population activation from a deterministic
  all-player room graph.
- [x] Reproduce first-activation transitions through replay/checkpoint.
- [x] Reproduce deactivate -> checkpoint -> restore -> reactivate continuation
  with the same authoritative checksum.

The checkpointed `d2legacy.population.plan/v2` stores a deterministic active
flag and inactive archive per room. A generated monster carries a stable
room-resident marker; leaving the occupied-room-plus-neighbors window removes
its live ECS entity and archives an allowlisted scalar/semantic component map,
while re-entry creates a new live entity with the same spawn/selectable IDs and
behavioral state. The first acceptance fixture crosses a three-room graph,
checkpoints while the monster is inactive, reconstructs a new Lua runtime, and
proves identical reactivation continuation. This is Dark Magic semantic state,
not a vanilla save/protocol compatibility structure.

Exact expansion 1.14d activation distance/tick ordering, long-inactive healing,
corpse lifetime, external target/state graph restoration, broader generated-
level coverage, and presentation residency remain open and probe-gated. Older
recovered inactive-unit code is architectural evidence only.

## P1: strengthen and complete the first multiplayer gameplay loop

### G8 — Combat fidelity tranche 1

Status: **partial**. One Lua-owned melee/missile damage path, timed states,
death transaction, fixed-point vocabulary, and deterministic vectors exist.

Implement one shared ordered transaction for chance-to-hit, block, avoidance,
physical/elemental/magic mitigation, caps/negative resistance, pierce, absorb,
critical/deadly/mastery, Crushing Blow, Open Wounds, leech, hit recovery,
knockback, poison/periodic damage, and durability. Keep unresolved arithmetic
labeled and probe-driven. Then complete ordinary player death/corpse/gold/XP
semantics before Hardcore durable death or broad PvP.

### G9 — Skill/state/missile behavior-family coverage

Status: **partial**. Cast, timed-state, melee, straight missile, Fire Bolt, and
supporting target/motion primitives exist.

- [ ] Generate a mounted-data report of server start/do behavior IDs, consuming
  skills, implementation family, missing family, and evidence status.
- [ ] Implement reusable targeted, point, self, area/nova, buff/debuff/curse/aura,
  summon, corpse, movement, missile, and trap families in dependency order.
- [ ] Use representative skills as fixtures; do not implement seven trees independently.

### G10 — Item-source lifecycle

Status: **partial**. Provenance-preserving stat sources, authoritative containers,
equipment transactions, generation, sockets, and runeword recognition foundations exist.

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
  cast timing, path types, stamina, and inactive rooms;
- items/economy: NoDrop, MF, runewords, charms, sockets, Cube operations, and pricing;
- world: object operations, doors, chests, shrines, warps, waypoints, portals,
  quest dialogue, difficulty consumers, and endgame eligibility;
- multiplayer: party XP/quest credit, hostility, trade, interest management,
  reconnect, and PvP.

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
