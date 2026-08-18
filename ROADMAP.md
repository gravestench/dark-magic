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
unimplemented and no unsupported channel arithmetic is implied. The first
generic consumer now reads only `d2legacy.combat.event`, never the
co-composed melee detail by fallback. An empty ECS `death_observed` component
marks each result after death attribution sees it, filtering it from later
death passes without destroying facts that independent proc/reaction consumers
may still require. Melee-only reaction fixtures remain invisible to generic
damage/death policy. Basic melee now also composes a generic `attack_result`
component on every
resolved impact. Its explicit `hit`, `miss`, or `invalidated` outcome separates
attack resolution from damage commitment: hits may additionally carry damage,
while misses and targets invalidated at impact cannot fabricate a combat-damage
event. Attack rating, defense, and chance remain inspectable ECS facts; block
and avoidance vocabulary/arithmetic stay probe-gated. Combat Lab now
coalesces those co-composed ECS facts by entity instead of
showing duplicate rows, and formats current raw damage/channel/remaining-health
fields rather than the retired scalar names that could fault after a hit.
The block/avoidance gate now has a strict analyzer/template contract accepting
only sanitized visual observations from an owned Expansion 1.14d single-player
runtime with a recorded executable SHA-256. It rejects Classic, earlier patches,
server/community-tool sources, mismatched controls, and outcome/health/reaction
contradictions. No block or avoidance arithmetic is promoted until that matrix
is populated. The first player-death foundation now consumes the same explicit
`unit_died` result through its own empty `player_death_observed` marker. It
composes checkpointed death state onto the existing durable player entity,
preserves immediate and ultimate-owned-unit attribution, snapshots the immutable
Hardcore rule without committing permanence, stops semantic/raw motion, removes
in-flight action components, filters dead actors from action systems, and emits
one value-only semantic event. Corpse/equipment transfer, gold and XP penalties,
recovery/respawn, exact DT/DD timing, multiple-corpse rules, save boundaries, and
Hardcore persistence remain explicitly unresolved Expansion 1.14d probes. A
strict player-death capture/analyzer now accepts only visual observations from a
probe-created softcore character in an owned, executable-fingerprinted Expansion
1.14d single-player runtime. It normalizes death/respawn frame intervals and
equipment/inventory/gold/XP/corpse transitions, rejects imported saves and every
unsupported runtime/tool source, and promotes no behavior until its matrix is
populated. G9 remains current through
target-locked mounted-data and localized TBL skill evidence, case-stable pinned
MPQ tables, AnimData/effective-rate melee and spell actions, current-state melee
target revalidation, straight-missile, timed-state, reactive-state, shared cast
cues/overlays/sounds, exact standalone-DCC ground anchoring and missile
direction/blend presentation, reconnect-safe connected semantic cast/state/
monster-death projection through a bounded `ClientView/v11`/`EventView/v3`
tail and disposable ECS mirrors, including record-authored monster/player
overlay-height attachment, and a bounded `WorldView/v5` living-monster, corpse,
and live-missile projection whose 25 Hz position, facing, and effective monster-
mode updates feed the same presentation-only ECS recipes used offline without
exposing AI policy. The reliable view also carries bounded persistent target/
state/record-period relationships and reconstructs them as disposable
presentation-only ECS entities. Connected aura graphics therefore use the
offline overlay/cycle path without exposing source identity, skill level,
stats, radius, party/filter decisions, or arbitration, alongside definition-
driven radial-missile slices as of
2026-08-18. Nova is now the first
exact-ID `missile.radial` configuration: one targetless cast creates a shared-
identity ring of ordinary ECS projectile entities, level-scaled count/mana/
five-band lightning damage, and separate cast-target contact-lock entities.
The family is reusable and contains no Nova ID/name branch. Exact 1.14d radial
angular phase, negative acceleration, and complete
`LastCollide`/`NextHit` ordering remain explicitly partial. The first reusable
cross-skill damage-modifier slice now resolves Fire Bolt's exact
`EDmgSymPerCalc` hard-point references by skill name to Fire Ball/Meteor IDs,
snapshots their combined level-derived percentage on the generic cast, and
applies it to the generic level-scaled damage range. Joined localized TBL text
confirms both player-visible relationships and the `%s` heading token. Exact
1.14d percentage rounding and modifier ordering remain probe-gated. The next
family now admits Fire Ball as an exact-ID
`missile.straight-impact-area` configuration. Swept first contact produces a
stable impact point, one deterministic radius query, independently rolled
generic fire results for ordered targets, and a separate short-lived
presentation-only missile-effect entity. The latter has no damage component
and can cross checkpoints without reapplying policy. Exact 1.14d footprint/
radius units, per-target RNG stream behavior, impact-point rounding, and
mastery/resistance ordering remain partial. The direct control-effect tranche
now admits Ice Blast as exact-ID `missile.straight-freeze`. The generic cast
snapshots record-authored freeze length plus hard-level duration synergies, the
generic projectile emits an ordinary ECS state request after a nonlethal hit,
and the existing timed-state/action filter owns checkpointing, suppression,
refresh, and expiration. A separate presentation-only entity uses the
referenced freeze explosion. Exact 1.14d resistance/immunity, monster-class
cold effectiveness, cross-source state replacement, PvP chill conversion, and
impact/action timing remain probe-gated. Glacial Spike now composes those same
mechanisms as exact-ID `missile.straight-impact-area-freeze`: its owned radius
and freeze-length formulas, Blizzard duration modifier, three cold-damage
synergies, and localized area-freeze intent produce stable ordered area damage
plus one ordinary freeze request per surviving target. The projectile and
state systems contain no Glacial Spike branch. Its referenced center explosion
is presentation-only; the second ejecta sub-missile, exact radius/footprint and
rounding, resistance/immunity and monster-class effectiveness, cross-source
replacement, PvP chill conversion, and action timing remain explicit probes.
Teleport now opens the first exact-ID point-movement family. Its owned 1.14d
Skills row supplies server-do 27, right-skill-only assignment, warp intent, SC
action, signed 24-minus-1-per-level mana with an authored 1-mana floor, and no
cross-skill modifier. Owned Levels rows supply policy 0/1/2, with only Duriel's
Lair carrying 2; layered TBL text says the action instantly moves to a
destination within line of sight. The generic effect validates bounds, static
footprint, dynamic ECS occupancy, and the conservative limited-level line
trace, then atomically relocates the existing player entity, stops semantic and
raw motion, cancels forced motion, and emits one generic ECS relocation fact.
Exact viewport/visibility range, policy-2 meaning, invalid-target mana behavior,
nearest-free fallback, room-edge/effect ordering, owned-unit following, and
remaining presentation semantics remain explicit 1.14d probes. The first friendly-target timed multi-stat
family now admits exact-ID Enchant. Its owned Skills/States/SkillDesc/skillcalc
and layered TBL evidence drives target-or-self resolution, one checkpointed
state owning three independent ECS stat sources, level-band fire damage, Warmth
hard-point synergy, duration, and attack-rating percentage. Shared melee
consumes the resulting fire and to-hit sources without recognizing Enchant.
Ranged-weapon one-third fire damage, party/PvP target distinctions, target
range/LOS, replacement across casters, action-rate/modifier timing, and exact
modifier/rounding order remain explicit Expansion 1.14d probes. The first
point-centered area-curse family now admits exact-ID Amplify Damage.
Its owned row drives stable hostile-area selection, timed curse-state/stat
sources, level-ranked one-curse replacement, and the one-fifth resistance
reduction when attempting to break a monster's physical immunity. Shared
physical mitigation consumes the resulting source without recognizing the
skill. Exact LineOfSight-4/radius units, curse resistance, target-class
eligibility, PvP, equal-level cross-caster ownership, client-only curse missile presentation,
and ordering against other resistance sources remain explicit 1.14d probes.
The same family now admits exact-ID Weaken from its distinct owned record/TBL
shape. It emits a generic negative outgoing-physical-damage percentage source;
ordinary melee consumes that source without recognizing Weaken, while shared
curse exclusivity and ranked replacement remain owned by the timed-state
mechanism. Exact outgoing modifier ordering, non-weapon monster attacks,
hirelings/summons, target eligibility, PvP, and presentation remain explicit
1.14d probes. Might now opens the selected-right party-aura family. Its owned
Skills/States/SkillDesc/TBL records and Blizzard's Expansion control
documentation make the right assignment itself authoritative: left assignment
is rejected, clicking the selected aura creates no cast, and zero mana is
preserved. One ECS emitter reconciles living same-level party members inside
the level-scaled radius into relationship entities co-composed with ordinary
`damagepercent` stat sources. Distinct aura states coexist, while duplicate
Might sources select the strongest level and a deterministic equal-strength
source. Leaving range, party, level, life, room activity, or selection destroys
the relationship and modifier together. Offline presentation follows the
pinned Might States/Overlay records and cycles one active aura graphic per
affected unit at the record's 50-tick period without disabling the other aura
effects. Exact `aurafilter=73731` owned-unit coverage, application/leave tick
ordering, equal-source ownership, cross-family visual cadence, and sound
lifetime remain explicit 1.14d probes. Connected persistent overlays now cross
a bounded semantic relationship rather than copying aura gameplay facts.
Defiance is the second exact configuration in that family. Its owned records
share Might's selected-right, zero-mana, filter, radius, and 50-tick shape but
select `skill_armor_percent=70+10*(level-1)`. The decoder maps that authored
stat vocabulary to the existing generic defense-percent source. Real TBL text
states that the active aura increases party defense, and the pinned Defiance
state/Overlay rows select its front/back DCCs. A two-Paladin checkpoint test
proves Defiance and Might remain independent relationships on both targets;
generic derived defense consumes Defiance without recognizing skill ID 104.

Spell Lab now wraps the production Blood Moor scene instead of maintaining a
parallel spell simulator. Its ephemeral level-30 Sorceress fixture
learns all 13 exact-ID configurations at level 20 through the owned
Skills/SkillDesc records, begins with Fire Bolt and Amplify Damage assigned, and
uses the ordinary HUD, command admission, mana, cast, projectile, state, damage,
monster, and renderer paths. A real-MPQ acceptance casts Fire Bolt and proves
its record-derived three-mana payment, Sorceress SC action/release/completion,
and projectile damage. This makes current
presentation omissions directly observable without giving the lab or renderer
skill-specific policy. The same acceptance exposed `MonLvl.txt` as another
retail hash-table member absent from incomplete MPQ listfiles; it is now pinned
inside the immutable game-data generation, and production population fails with
record-level context instead of silently admitting an empty encounter. A
generic state-overlay presentation adapter now follows live ECS state IDs into
the pinned States.txt `overlay1`/`overlay2`/`castoverlay`/`removerlay` keys and
then into Overlay.txt. It renders record-selected DCC paths, unit palette,
direction count, animation rate, offsets, front/back precedence, and legacy
blend through ordinary retained world nodes. Apply/refresh/remove events drive
one-shot effects while persistent overlays follow the current target position;
the renderer contains no skill/state-name branch. Owned Expansion records and
DCC members are pinned for Frozen Armor, Enchant, Amplify Damage, and Weaken.
Monster definitions now retain their exact MonStats2 `OverlayHeight` category;
the shared renderer selects the corresponding Overlay.txt `Height1..4` offset,
while player targets select Height2. `ClientView/v11` carries only that bounded
attachment category and reconstructs it as a disposable ECS presentation
anchor, keeping offline and connected state effects on the same path.
The shared cast adapter now also joins every admitted Skills.txt row to its
`anim`, `seqtrans`, start/effect sound, cast-overlay, and client-missile fields.
Non-melee SC actions stop competing motion through an empty ECS action filter,
face the target, use token/mode/weapon-class AnimData for release/completion,
emit semantic start/effect cues, and return to neutral. Offline presentation
resolves those cues into record-selected sounds and the same generic overlay
renderer. Standalone DCC rendering preserves the codec's authored ground
origin, fixing both Fire Ball launch height and `FireCast2` attachment; missile
velocity uses `math.atan2` plus exact owned 1/4/8/16/32-way DCC order instead of
an eight-way collapse. Missiles.txt `Trans=1` and Overlay.txt `Trans=3` select
their table-specific luminous blend paths. A held-pointer gate now emits one
request per authoritative ECS action, so a multi-frame click cannot queue a
second cast while a deliberate hold still repeats after completion. Connected
clients now receive nearby live projectile/effect visuals through bounded
`WorldView/v5` records and reconstruct only a `d2legacy.presentation.missile`
component. The reliable stream owns lifecycle and record-derived DCC/palette/
direction/timing/blend/offset facts; the 25 Hz disposable transform stream
updates positions for already admitted identities. Damage, collision, target,
contact-lock, ownership-policy, and lifetime fields remain server-only. SQ
sequences, faster-cast-rate/equipment timing, overlay light/1-of-N, character,
and multi-direction details, client-only curse missiles, interruption/refund, and
remaining semantic event families remain explicit presentation work. A
strict client-function-30 visual probe now gates the curse missiles on a
fingerprinted owned Expansion 1.14d empty/single/multi-target matrix rather
than recovered/community behavior. A separate strict cast-rate analyzer now
joins owned ItemStatCost/Properties/Skills/TBL identities to visual SC, SQ,
weapon-class, and raw-FCR observations. It reports a missing target matrix and
promotes no breakpoint formula, preventing older-version tables from becoming
1.14d behavior by assumption. Missile audio now has the same evidence boundary:
owned `Missiles.txt`/`Sounds.txt` joins pin the referenced wave records and
their loop/group/stream facts, while a strict seven-case isolated-audio/video
probe measures start, stop, contact/expiration, and radial multiplicity. The
existing unproduced missile-event consumer is deliberately not wired until the
complete Expansion 1.14d matrix establishes lifecycle semantics. A
matched frontend profile also
removed the title-to-main-menu
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
| M15 asset knowledge | partial | Typed/recovered coverage is broad. The owned 1.14d Expansion Skills/Missiles report now inventories 357 skill rows, 172 server behavior signatures, 13 exact-ID implementations, 344 missing skills, and winning-layer provenance. A second exact-ID report joins Skills/SkillDesc formulas to layered locale TBL text, replacement tokens, and cross-skill references. Might and Defiance Skills, States, SkillDesc, locale TBL, sound keys, front/back Overlay rows, and DCC members are pinned alongside the existing missile/cast evidence. Retail `MonPreset.txt`, `MonStats2.txt`, `MonLvl.txt`, and `SkillDesc.txt` members omitted from incomplete listfiles remain explicitly discovered in the immutable generation. Runtime aura filter/timing/sound semantics remain evidence work rather than record inference. |
| M16 presentation primitives | partial | MPQ-backed render/audio primitives exist. Missile entities select record-authored travel/impact DCCs, sounds, exact 1/4/8/16/32-way direction order, authored ground origins, and table-specific luminous blend; semantic timed states and aura relationships resolve States/Overlay records into shared world overlays without skill branches. Distinct aura modifiers stay active while presentation rotates one aura graphic per affected unit using the record period. Connected authority projects only bounded target/state/period relationships and the disposable client ECS binds them to existing unit mirrors, so the identical Lua cycle/overlay path works without exposing source identity, skill level, stats, radius, filter/party policy, or arbitration. MonStats2 `OverlayHeight` selects Overlay.txt `Height1..4` attachment offsets for live monsters, players use Height2, and connected cues retain that category through an ECS presentation anchor. Admitted Skills rows drive SC actor action timing, semantic start/effect cues, cast sounds, and cast overlays through the same world renderer. Connected clients reconstruct bounded living-monster composites, retain the same mirror as a nonselectable/noncolliding DT corpse, and consume a typed death-sound cue. Authority also collapses private AI/velocity facts into the same offline `DT > A1 > WL > authored` presentation precedence; the existing 25 Hz transform channel carries only the resulting mode and facing. The network projection omits AI state/targets, loot, XP, kill attribution, player-count policy, corpse usability, aura gameplay facts, and every other authority field. The same reliable view carries bounded projectile/effect visuals. These presentation-only ECS components keep offline and connected play on the same Lua renderer. Strict owned-runtime probes gate aura sound and cross-family cadence, client-function-30 curse attachment/motion, SC/SQ/FCR/weapon-class timing, and missile travel/impact audio lifecycle/multiplicity on complete target matrices; none promotes inferred behavior. Client assembly consumes a backend-neutral desktop contract; Raylib is the production default and the `ebitengine` tag supplies an experimental retained-composition/input/capture adapter. Populated probe vectors, exact monster animation phase/start timing, overlay light/variant/character/multi-direction semantics, record-referenced client-only curse layers, missile semantic audio production/projection, player-death and remaining semantic event families, Ebitengine native audio, console drawing, and GPU palette parity remain. |
| M17 front end | foundation complete | The Lua-authored front end and Realm flow exist. MPQ-backed locale tables now cross one sequential buffering boundary instead of issuing decoder-granularity random archive reads. Startup warms only title/main-menu assets, secondary destinations use visible main-menu think time, and character interaction animations remain scoped to character creation. Remaining work is UI fidelity, not the former multi-second transition stall or whole-frontend eager preload. |
| M18 in-game shell | foundation complete | HUD and major overlay shells exist; the party panel now consumes an owner-scoped semantic projection, while remaining raw/ad hoc reads migrate as their gameplay domains mature. |
| M19 character/item/save fidelity | partial | Canonical profile and Realm character persistence exist; the complete Dark Magic durable semantic character does not. Vanilla save interoperability is out of scope. |
| M20 world fidelity | partial | Deterministic Act I generation, collision, transitions, dynamic occupancy, population, and level-scoped persistent-identity room residents exist. Timed-state/stat-source/event references, an owned-unit graph, corpse, straight projectile, imported ground item, stateful interaction object, and separately resident pending-action relationship survive inactivation without scalar graph copies. The inactive ECS tag removes live capabilities, suspends opted-in systems, and filters projections; generic pre-plan attachment plus pickup/re-drop transitions reuse the same contract. Public loot policy, retail object/event families, exact corpse/projectile/event timing, 1.14d streaming behavior, and campaign breadth remain. |
| M21 Diablo simulation | foundation complete | Lua owns the current player, monster, skill, missile, state, death, loot, quest, item, and owned-unit vertical slices. Melee and missile contact share one ordered direct-damage commit/result boundary, and lethal player results now compose death/action-filter state onto the same character entity with independent consumer markers. Block/avoidance, typed bundle breadth, secondary damage effects, player-death consequences, movement, item activation, object, and content breadth remain below. |
| M22 networking | complete | One `Session`, authenticated semantic commands, deterministic ordering, filtered views, reconnect, replay/checkpoint, direct/listen/dedicated/Realm modes, and impairment/soak coverage exist. `ClientView/v11` embeds a strongly typed `EventView/v3` and `WorldView/v5`: the former is a proximity-filtered reliable 64-tick/256-entry semantic cast/state/monster-death tail with explicit truncation/gap detection and reconnect epochs; the latter adds bounded nearby living-monster composites, retained corpse presentation, at most 512 presentation-only missile records, and at most 512 persistent target/state/period relationships. Untrusted-input validation bounds each collection, rejects duplicate relationship keys, and preserves stable ordering. Actor/missile transforms share the fixed 25 Hz datagram budget by distance; actor records carry position, facing, and effective two-byte animation mode, never private monster AI state. Aura source identity, skill/stat facts, radius, eligibility, and arbitration never enter the view. Corpse identity and visual lifecycle remain reliable. |
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
- [x] Discover startup-critical `MonPreset.txt`, `MonStats2.txt`, `MonLvl.txt`,
  and `SkillDesc.txt` hash-table
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
monster-death transaction, first player-death state transition, fixed-point
vocabulary, deterministic vectors, and an explicit shared direct-damage result
record exist.

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
- [x] Make monster death attribution consume only the generic damage result,
  not a source-specific melee fallback. Compose an empty `death_observed` ECS
  marker after the pass so checkpoint/replay and later ticks cannot reconsume
  the result, while independent effect consumers retain the event entity.
- [x] Compose a generic attack-result ECS fact with every basic-melee impact.
  Represent current `hit`, `miss`, and target `invalidated` outcomes separately
  from damage commitment, retaining attack rating, defense, and hit chance for
  later consumers and diagnostics. Only a hit may carry a damage result.
- [x] Update Combat Lab to coalesce attack, damage, typed-bundle, melee-detail,
  and animation components by ECS entity. Render explicit outcome/channel/raw
  health fields and lock the composed snapshot boundary with a module test.
- [ ] Extend the generic outcome vocabulary with target-locked Expansion 1.14d
  block and avoidance families; do not infer their order or eligibility from
  the current hit boolean.
- [x] Add a strict Expansion 1.14d single-player owned-runtime defense-outcome
  analyzer and capture template. Fingerprint the executable and sanitized
  capture; require matched controls and normalize miss/damage/block/avoid/lethal
  rates without accepting Classic, older patches, servers, saves, community
  tools, or reconstructed-runtime observations.
- [x] Commit the first ordinary player-death state transition from an explicit
  generic `unit_died` result. Preserve the durable ECS player entity and raw/
  ultimate-owner attribution; compose checkpointed death state plus one semantic
  event; use an independent empty consumer marker; stop motion and remove/filter
  live action capabilities. Record unresolved consequences as pending rather
  than inventing corpse, gold, XP, respawn, animation-timing, save, or Hardcore
  permanence policy.
- [x] Add a strict softcore player-death capture/analyzer for a probe-created
  character in an owned Expansion 1.14d single-player runtime. Fingerprint the
  executable and sanitized visual timeline; normalize DT/DD, respawn-control,
  corpse/equipment/inventory, carried/stashed/ground-gold, XP-loss/recovery,
  multiple-death, and visual save/rejoin observations. Reject Classic, older
  patches, servers, community/memory tools, and imported save data.
- [ ] Populate standing/moving/attacking melee/missile shield-block and passive-
  avoidance matrices from that owned runtime, then promote only the ordering,
  eligibility, cap, and movement facts the observations resolve.
- [ ] Define the complete independent consumer-marker/retirement contract as
  attacker, defender, proc, quest, audio, and presentation event families land;
  do not centralize their policy or retain completed event entities forever.
- [ ] Extend the bundle and ordered transaction with verified drain, cold/
  freeze, poison-duration, conversion, and periodic-application facts without
  treating them as immediate health damage.
- [ ] Add target-locked Expansion 1.14d evidence and ordered stages for block,
  avoidance, resistance caps/negative values, pierce, absorb, critical/deadly/
  mastery, Crushing Blow, Open Wounds, leech, hit recovery, poison/periodic
  damage, and durability.
- [ ] Complete remaining ordinary softcore corpse/equipment, gold, XP,
  recovery/respawn, multiple-corpse, exact animation-timing, and save semantics
  from populated owned-runtime observations before Hardcore durable death or
  broad PvP.

Implement one shared ordered transaction for chance-to-hit, block, avoidance,
physical/elemental/magic mitigation, caps/negative resistance, pierce, absorb,
critical/deadly/mastery, Crushing Blow, Open Wounds, leech, hit recovery,
knockback, poison/periodic damage, and durability. Keep unresolved arithmetic
labeled and probe-driven. Then complete ordinary player death/corpse/gold/XP
semantics before Hardcore durable death or broad PvP.

### G9 — Skill/state/missile behavior-family coverage

Status: **partial**. Generic cast lifecycle, timed state, melee, straight-
missile, and radial-missile behavior families plus supporting target/motion
primitives exist. Fire Bolt is the first explicitly supported expansion 1.14d straight-missile
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
- [x] Add a definition-driven targetless radial-missile family and admit exact
  Expansion 1.14d Nova by ID. Decode its Skills/Missiles count, mana, five
  level-damage bands, lightning channel, motion/lifetime, presentation, and
  repeat-contact records; materialize every ray as an ordinary ECS entity with
  one shared cast ID and an independent cast-target contact-lock entity. Prove
  reuse with a second synthetic record configuration and checkpoint the live
  ring without a Nova-specific branch.
- [x] Add a reusable hard-point elemental-damage modifier shape for exact
  `EDmgSymPerCalc=(skill('…'.blvl)+…)*par8` records. Resolve referenced skill
  names to exact IDs, snapshot the authoritative learned levels as a generic
  cast percentage, and apply it after the five authored damage bands. Fire
  Bolt now receives 16% per Fire Ball and Meteor hard point from its owned row;
  checkpoint coverage proves the resulting damage bounds without a skill-ID
  branch. Keep exact percentage rounding and broader modifier ordering partial
  pending owned 1.14d vectors.
- [x] Add a reusable straight-missile area-impact family and admit exact
  Expansion 1.14d Fire Ball by ID. Validate missile/hit functions, authored
  radius, explosion-missile presentation, level bands, mana, and the generic
  hard-point modifier; resolve a swept impact once, damage stable ordered
  in-radius targets through the shared damage boundary, and materialize a
  separate non-damaging ECS effect with record-authored lifetime. Prove family
  reuse with a synthetic configuration and checkpoint the live aftermath.
- [x] Add a reusable straight-missile on-hit state family and admit exact
  Expansion 1.14d Ice Blast by ID. Validate missile damage function 4, cold
  damage and level bands, mana, direct freeze-explosion presentation,
  hard-point damage synergies, base/per-level freeze frames, and Glacial Spike
  hard-point duration synergy. Snapshot duration on the cast, emit a generic
  monster-cold state request only after nonlethal contact, and prove existing
  ECS action filtering, checkpointing, refresh, and expiration own the result.
- [x] Compose the reusable straight-missile area-impact and on-hit-state
  mechanisms and admit exact Expansion 1.14d Glacial Spike by ID. Validate its
  area hit function, freeze result flags, radius and freeze formulas, Blizzard
  duration modifier, three hard-point damage synergies, and center explosion;
  apply independently rolled shared cold results in stable target order and
  emit an ordinary freeze request for each nonlethal in-radius result without a
  Glacial-Spike-specific command, component, or system branch.
- [x] Add the first reusable point-movement family and admit exact Expansion
  1.14d Teleport by ID. Decode signed mana progression and authored Levels
  policy, validate static/dynamic destination occupancy, atomically relocate the
  existing ECS player while stopping competing motion, and emit a generic
  relocation-result entity. Keep viewport/range, limited-level meaning,
  invalid-target payment/fallback, owned-unit, effect ordering, and remaining
  presentation semantics partial; the shared SC action/cue path is present.
- [x] Add the first reusable friendly-target timed multi-stat family and admit
  exact Expansion 1.14d Enchant by ID. Decode its target flags, duration,
  level-band fire range, `toht` attack bonus, and Warmth hard-point modifier;
  let one timed state own three provenance-preserving ECS stat sources and make
  shared melee consume the resulting to-hit and fire facts without a
  skill-specific branch. Keep ranged one-third fire damage and remaining
  targeting/ordering/presentation edges partial.
- [x] Add the first reusable point-centered area-curse family and admit exact
  Expansion 1.14d Amplify Damage by ID. Decode the radius, duration, target
  state/filter, physical-resistance source, mana, and target-version one-fifth
  immune-breaking rule; apply in stable semantic order and add ranked
  exclusive-state replacement so weaker curses cannot displace stronger ones.
- [x] Reuse the point-centered area-curse family for exact Expansion 1.14d
  Weaken by decoding its distinct radius, duration, target state, and negative
  `damagepercent` stat recipe. Make ordinary melee consume that generic
  outgoing-physical-damage source without a skill ID/name branch, while the
  shared timed-state mechanism owns curse exclusivity and replacement.
- [x] Add the first selected-right party-aura family and admit exact Expansion
  1.14d Might by ID. Decode its right-only, zero-mana activation, level-scaled
  radius/damage, aura states, filter, and period; reconcile one ECS emitter into
  self/party target relationships co-composed with ordinary stat sources.
  Preserve different aura states concurrently, suppress weaker duplicate
  states deterministically, remove effects atomically on eligibility changes,
  and rotate the offline target's record-authored aura graphic without turning
  gameplay modifiers off. Keep the complete target filter, owned units,
  refresh/leave ordering, sound, and equal-source ownership partial pending
  owned 1.14d evidence.
- [x] Carry persistent aura presentation across the reliable connected view as
  a versioned, bounded target/state/record-period list. Preserve multiple
  distinct states on one target, reject duplicate keys and invalid periods,
  bind each relationship to an already admitted disposable unit mirror, and
  feed the shared Lua aura-cycle/Overlay renderer. Keep emitter identity,
  skill/level, stats, radius, filter/party policy, and same-state arbitration
  server-only.
- [x] Reuse `aura.selected-party-stat` for exact Expansion 1.14d Defiance.
  Validate its Skills/States/SkillDesc and layered TBL defense/radius intent,
  map authored `skill_armor_percent` to the generic defense-percent source,
  pin its front/back overlay rows and DCCs, and prove checkpointed Might plus
  Defiance relationships coexist on every eligible party target. Keep the
  family decoder selected by record shape and exact manifest admission, never
  by skill name or ID.
- [x] Add a production-backed Spell Lab scene that grants only the exact-ID
  manifest through owned Skills/SkillDesc records, delegates world/HUD/input/
  authority/presentation to `game_world`, and proves a real-MPQ Fire Bolt cast
  through record-derived mana payment and projectile damage. Keep its overlay
  diagnostic-only and renderer-neutral.
- [x] Add a generic timed-state presentation path: snapshot active ECS
  state-to-target relationships, resolve States.txt overlay references through
  Overlay.txt, render active front/back layers and apply/remove one-shots, and
  audit the dynamic `data/global/overlays/` asset family. Pin owned target rows
  and DCC members without recognizing a skill/state name in presentation.
- [x] Join every admitted exact skill ID to its shared Skills.txt action fields,
  run non-melee SC actions from token/mode/weapon-class AnimData, filter
  locomotion with an empty ECS cast-action component, emit semantic start and
  effect cues, and resolve record-selected cast sounds/overlays without a
  skill-name or ID renderer branch. Pin the owned Sorceress SC +7 release/+14
  completion schedule through the production Spell Lab acceptance.
- [x] Preserve standalone DCC authored ground origins and exact owned
  1/4/8/16/32-way missile direction interleave, use Lua 5.1 `math.atan2` for
  continuous velocity, and map Missiles/Overlay transparency through their
  table-specific luminous modes. Pin Fire Ball's 16 directions, `Trans=1`,
  missile canvas, and `FireCast2` canvas from owned Expansion 1.14d assets.
- [x] Edge-gate held left-skill input against the authoritative ECS action
  lifecycle so a click spanning multiple render frames submits one cast while
  a deliberate hold repeats only after the prior action completes.
- [x] Project nearby cast/state semantic facts through a bounded reliable
  `ClientView/v11`/`EventView/v3` tail, baseline durable history on join/reconnect, reject
  gaps/truncation instead of silently losing presentation, and materialize only
  new facts as short-lived spatial ECS entities in the disposable connected
  client world so the offline Lua cue/overlay/sound path remains the sole
  presentation consumer.
- [x] Add `WorldView/v5` live projectile/effect presentation for connected
  clients. Project only bounded record-derived visual and spatial fields,
  reconstruct a presentation-only ECS component, keep lifecycle reliable, and
  merge actor/missile positions by distance into the existing 25 Hz datagram
  budget. Do not expose damage, collision, targets, contact locks, ownership
  policy, or lifetime state to the disposable client world.
- [x] Extend connected presentation to ordinary monsters and their corpse
  transition. Project bounded MonStats/MonStats2-derived composite facts and
  the stable spawn identity, retain one disposable ECS mirror while changing
  its mode to `DT`, and remove selectable/collider components so the corpse is
  visual but cannot enter interaction or locomotion queries. Carry only the
  typed `monster_death_presented` cue needed for record-authored death audio;
  keep loot, XP, credit, player-count policy, and corpse-use facts server-only.
- [x] Publish the effective living-monster animation outcome through the
  existing `WorldView/v5` actor transform. Match offline precedence by keeping
  death terminal, selecting `A1` for authority attack state, `WL` for nonzero
  motion, and the authored mode otherwise; carry only mode/facing at 25 Hz and
  prove that AI state/target policy never enters the JSON view or client ECS.
- [x] Add a strict owned Expansion 1.14d visual capture/analyzer for client
  function 30 that fingerprints provenance, rejects Classic/old/server/
  community/imported-save evidence, normalizes both record-referenced missile
  layers, and reports the missing empty/single/multi-target matrix without
  promoting an attachment or motion role.
- [x] Add a strict owned Expansion 1.14d cast-rate capture/analyzer. Pin
  ItemStatCost ID 105, Properties `cast1/2/3`, `ModStr4v` TBL text, Fire Bolt's
  SC action, and Lightning's SQ/12 transition; normalize visual release and
  completion timing across a raw-FCR/weapon-class matrix without embedding an
  older breakpoint formula or admitting Lightning's unfinished behavior.
- [x] Add a strict owned Expansion 1.14d missile-audio capture/analyzer. Pin
  Fire Bolt, Fire Ball, Nova, Ice Blast, and Glacial Spike `Missiles.txt` sound
  references plus joined `Sounds.txt` filename/group/loop/stream facts; measure
  effect/contact/removal frames and audible instance intervals across the exact
  seven-case matrix without treating `Loop=1` or radial count as runtime policy.
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
  and monster-class duration modifiers, exact integer/tick ordering, and
  remaining state-effect presentation semantics before upgrading it from
  partial behavior. Its generic SC action, cast sound/overlay, and persistent
  state overlay now use the shared paths.
- [ ] Extend the now-present targeted, point, self, area/nova, buff/curse,
  selected-party-aura, movement, and missile primitives into missing debuff,
  aura-stat/pulse, summon, corpse, trap, and ranged-weapon families in
  dependency order.
- [ ] Use representative skills as fixtures; do not implement seven trees independently.

`skill_behavior_coverage` mounts owned archives, reads the winning Expansion
1.14d Skills.txt and Missiles.txt tables, groups every skill by server start/do
and referenced missile server-do function IDs, and reports every consumer with
its explicit family, missing-family flag, and evidence status. The current
owned-data baseline is 357 skill rows, 172 signatures, 13 explicitly admitted
configurations, and 344 missing configurations. The report fails if a declared
skill or referenced server missile is absent, and its synthetic test proves a
row with the same function signature is not admitted by resemblance. Generated
reports remain local; copyrighted tables are never copied into Git.

Spell Lab is an acceptance instrument, not a second implementation surface.
The scene delegates production `game_world` creation, update, and destruction;
its only authored presentation is a read-only legend reporting the admitted
exact-ID count, current left/right assignments, and mana. Immutable initial
data enables a development-only learned-skill fixture whose IDs are derived
from the target-locked implementation manifest, resolved through the normal
owned Skills/SkillDesc records, and rejected if unknown, passive, or
unassignable. It currently grants the 13 manifest-backed Expansion
1.14d configurations at level 20, assigns Fire Bolt left and Amplify Damage
right, places three production hostiles in Blood Moor, and supplies a deep mana
pool for repeated inspection. Ordinary HUD selection and world clicks still
create the same semantic commands used by offline, listen, dedicated, and Realm
sessions. The owned-asset acceptance asserts the learned set and assignments,
then observes Fire Bolt's Sorceress SC mode, AnimData +7 effect and +14
completion, return to neutral, three-mana payment, and damage after shared
projectile contact. No lab, client, renderer, or presentation system branches
on Fire Bolt or any other skill ID.

That test also closed an asset-generation blind spot: retail `MonLvl.txt` can
be addressable by hash while absent from an MPQ `(listfile)`, just like the
already pinned `MonPreset.txt`, `MonStats2.txt`, and `SkillDesc.txt` members.
The immutable generation now requests it explicitly and proves case-insensitive
record loading from the pinned store. Monster data loading preserves the
underlying record error, and a nonempty Levels candidate set that resolves to
zero valid monsters is now fatal rather than silently producing a lab or served
game with no hostiles.

Timed-state presentation now consumes the same ECS graph instead of asking
skill systems to draw. `gameplay.world.state_snapshots` copies each live
instance's state ID, state-entity/target identities, current position/location,
and facing; inactive targets are excluded by the existing empty ECS marker.
The shared adapter follows `overlay1`, `overlay2`, `castoverlay`, and
`removerlay` from States.txt into Overlay.txt, whose `Filename`, `Frames`,
`AnimRate`, `NumDirections`, `PreDraw`, `Trans`, X/Y offsets, `Height1..4`, and
light metadata create retained DCC recipes under `data/global/overlays/`.
MonStats2 `OverlayHeight=1..4` selects the matching authored height offset;
players select Height2. State apply/refresh/remove events
carry copied target positions into one-shot presentation; active front/back
recipes follow moving targets and disappear with the state. The presentation-
coverage fingerprint now explicitly audits Overlay.txt plus the dynamic overlay
directory, while owned 1.14d coverage proves the Frozen Armor, Enchant,
Amplify Damage, Weaken, and shared curse-hit rows and DCCs exist. Owned data
pins all four height categories and both curse offset vectors. Light radius/
color is retained but has no world-light renderer yet; exact `1ofN`, character
restriction, and multi-direction mapping remain open.

Connected state/cast/monster-death presentation crosses an explicit semantic
boundary. The stateless authority projector selects `d2legacy.state.event`,
`d2legacy.skill.cast_cue`, and only the `monster_death_presented` member of
`d2legacy.monster.death_event`. It copies value fields plus the referenced
actor's current position/location/facing, filters to the authenticated player's
nearby act/level, and sorts stable tick/entity identities. The death cue is a
typed two-field payload: presentation kind and stable monster spawn ID. Loot,
XP, killer/credit identity, player-count inputs, treasure class, and drops are
not serialized. `EventView/v3` retains a 64-tick tail capped at 256 entries;
its `from_tick` and `truncated` facts make a correction gap detectable.
`ClientView/v11` validates type unions, identities, strings, ticks,
coordinates, ordering, and allocation bounds as untrusted network input. It
carries only the bounded 0..4 overlay-height category—not arbitrary appearance
data—and the client mirrors it into an ECS presentation anchor. Join,
reconnect, and authority reassignment increment a local event epoch and
establish a non-presenting high-water baseline, so checkpoint-durable history
never replays. Ordinary reliable corrections create only newer cast, state,
and monster-death cues in the disposable client ECS. Transform datagrams cannot
duplicate or prematurely replace cues, and the next reliable correction
retires the prior short-lived mirrors. Player death, missile audio, combat, and
forced-motion semantic events still need equally typed consumers before they
join this network view; no arbitrary component payload is serialized.

Connected monster state now uses the same presentation-only ECS boundary.
`WorldView/v5` projects a living monster's stable spawn/definition identity,
localized name key, MonStats2 composite token/mode/weapon class/components,
death-sound record key, bounded overlay-height category, facing, position, and
public health. For live animation, the projector applies the same precedence as
the offline snapshot adapter: authored `DT` remains terminal, authority attack
state becomes `A1`, nonzero velocity becomes `WL`, and otherwise the authored
mode survives. The existing 25 Hz actor transform already carries that
two-byte outcome and facing, so the disposable client needs no monster AI or
velocity component to drive the shared composite renderer. Tests pin each
precedence case, the datagram/session update, the client mirror, and absence of
raw AI state/target strings from the JSON projection. Exact authoritative
animation start/phase timing remains unresolved rather than being inferred
from packet arrival. When authority removes selection and collision on death,
the durable `d2legacy.monster.death` marker keeps that same entity in the view as a
`DT` corpse under the unchanged `monster:<spawn_id>` key. The client updates
the existing mirror and removes its selectable and collider components. This
uses ECS query membership to keep the corpse renderable while excluding it
from interaction and locomotion, rather than adding corpse-specific branches
to those systems. Corpse usability, lifetime, loot, XP, attribution, AI, and
occupancy policy remain authoritative and absent from the client view.

Live server-missile presentation now crosses a separate state boundary because
a moving projectile is not a replayable one-shot event. `WorldView/v5` joins
nearby `d2legacy.missile.projectile` and `d2legacy.missile.effect` entities to
their authoritative positions/locations, filters out invisible helper
missiles, and copies at most 512 visual records. Each record contains only the
stable view ID, projectile/effect kind, missile/DCC/palette identity, velocity
or pinned logical direction, direction count, frame rate, loop flag,
transparency mode, and authored offsets. The disposable client ECS reconstructs
these as `d2legacy.presentation.missile`; `gameplay.world.missile_snapshots`
then feeds the existing generic missile adapter and renderer exactly as it does
offline. The reliable correction owns creation, removal, and visual metadata.
The existing 25 Hz transform datagram may update only the position of an ID
already admitted by that reliable view, and merges actor/missile candidates by
distance within its fixed byte budget. Damage ranges, collision/contact rules,
targets, remaining ticks, and other authority never enter the client schema.
This closes connected presentation for admitted server projectiles and impact
effects; Skills.txt client-only missile functions such as curse function 30
remain gated by their separate owned-runtime probe.

Missile audio now has a separate strict evidence contract because the records
do not establish the lifecycle needed by a semantic event producer. Owned
`Missiles.txt` rows bind Fire Bolt, Fire Ball, Ice Blast, and Glacial Spike to
distinct travel/impact sounds and bind Nova to a travel sound without a hit
sound. Their joined `Sounds.txt` rows show that all four straight-projectile
travel records loop while the impact rows and Nova do not. A target-pinned
real-MPQ test protects those filenames, group sizes, loop/stream flags, impact
effect references, and immutable generation identity. The
`missile_audio_probe` then requires fixed-camera, stationary, isolated audio/
video observations and owned-MPQ waveform identification across Fire Bolt
expiration/contact, the other three straight-projectile contacts, and empty/
three-target Nova casts. Its report fingerprints the capture and normalizes
sound intervals against effect, contact, and removal frames. It does not create
`d2legacy.missile.event`, decide retained-handle ownership, or expose those
events through `EventView/v3`; those changes require a complete reviewed
seven-case Expansion 1.14d matrix first.

Cast presentation now follows a second generic record join. `cast_actions`
copies Skills.txt `anim`, `seqtrans`, `seqnum`, `UseAttackRate`, start/effect
sounds, `castoverlay`, and client-missile references for all 12 admitted IDs.
The cast lifecycle combines the actor token and current weapon class with the
SC mode, schedules release at the first non-sound AnimData event, completes at
the AnimData cursor wrap, and owns an empty `d2legacy.skill.cast_action` ECS
filter which excludes locomotion. Start/effect semantic cue entities carry only
skill/actor/target/tick facts; the offline world adapter re-resolves Skills and
Overlay records to play sounds and one-shot overlays. The owned Sorceress
`SOSCHTH`, `SOSC1HS`, and `SOSCSTF` records all pin 14-frame actions with release
at frame 7. Owned ItemStatCost ID 105 and Properties `cast1/2/3` identify
`item_fastercastrate`; `ModStr4v` resolves through owned `string.tbl` to
`Faster Cast Rate`. A strict target-runtime analyzer now requires 25 Hz SC and
SQ/12 visual timing across HTH/1HS/STF and discriminating raw-FCR values before
the shared scheduler changes. The probe reports evidence coverage rather than
declaring a breakpoint formula. Populating that matrix, implementing SQ/FCR,
mid-action weapon/equipment changes, interrupt/refund rules, and client-only
curse missile layers are still open.

Standalone DCC presentation now retains the stable decoded direction canvas and
sets the authored zero point as the render-node origin, matching the ground-
origin treatment already used by COF character composites. Owned Fire Ball
evidence pins `Fireball.dcc` direction 4 to `(-17,-81)-(17,-26)`, placing the
whole missile 26–81 pixels above its semantic world anchor, and `FireCast2.dcc`
to `(-74,-89)-(71,44)`. No guessed character-height offset is stored in the
skill. Velocity direction selection uses Lua 5.1 `math.atan2`, quantizes in
world space, and applies the exact DCC interleave for every direction count in
the owned Missiles table; 16/32-way art is no longer collapsed to actor facings.
Missiles.txt mode 1 and Overlay.txt mode 3 independently map to the renderer's
luminous screen blend. MonStats2 `OverlayHeight=1..4` now selects the matching
Overlay.txt `Height1..4` value and adds it to Yoffset for active and one-shot
state layers; player units select Height2. Remaining light/variant/character/
directional overlay fields stay explicit follow-up rather than being inferred.

World input now distinguishes a physical press from its subsequent rendered
down frames. One request is latched until an authoritative cast-action,
melee-approach, or melee-animation ECS component is observed and later clears.
Releasing rearms the next click; continuing to hold repeats only after that
action boundary. A fast module test covers both cases, closing the gap that let
a single ordinary click queue two back-to-back casts despite simulation tests
correctly validating each admitted command in isolation.

`skill_evidence` is the required companion investigation report for skill-tree
synergies and skills that modify other skills. It accepts exact IDs and a locale,
joins each Skills.txt row to SkillDesc.txt, resolves every name/description/
tooltip label through layered `string.tbl`, `expansionstring.tbl`, and
`patchstring.tbl`, records replacement tokens such as `%s`, and parses every
`skill('name'.blvl|lvl)` formula back to the referenced target skill ID. Missing
keys and unknown skill references fail or remain explicit instead of silently
dropping documentation. Fire Bolt now reports Fire Ball/Meteor hard-level fire-
damage synergies; Frozen Armor reports Shiver/Chilling Armor hard-level duration
and freeze-length modifiers in both gameplay and tooltip formulas; Ice Blast
reports its Ice Bolt/Blizzard/Frozen Orb damage and Glacial Spike duration
relationships. Nova's
locale records name it and describe an expanding electrically charged ring that
shocks nearby enemies; its joined formulas contain no cross-skill references.
Glacial Spike's localized records describe a magical ice comet that freezes or
kills nearby enemies and expose radius/freeze labels; its formulas resolve Ice
Bolt, Ice Blast, and Frozen Orb hard levels for cold damage plus Blizzard hard
levels for freeze duration. The decoder binds those player-visible
relationships to the exact owned Skills/Missiles rows rather than inferring
behavior from text alone.
Teleport's layered `skillld54` record says it instantly moves to a destination
within line of sight, while the exact owned skill row contains no cross-skill
formula. That intent is joined to the server-do, signed mana, assignment, and
Levels policy facts; it does not by itself invent a numeric visibility range or
invalid-destination sequence.
Enchant's layered `skillld52` record states that it enchants the equipped
weapon of a targeted character or minion, adds fire damage to melee and ranged
weapons, and gives ranged weapons one-third fire damage. Its SkillDesc joins
the Attack Bonus and Duration labels plus the `%s` synergy heading from
`patchstring.tbl`; the Skills formula resolves Warmth ID 37 as a 9%-per-hard-
point fire modifier. The owned `skillcalc.txt` row at special-parameter index
20 maps `toht` to to-hit, while the owned Enchant row supplies `ToHit=20`,
`LevToHit=9`, and no `ToHitCalc` override.
Amplify Damage's layered `skillld66` record says it curses a group of enemies
and increases the non-magic damage they receive. Its SkillDesc joins Damage
Taken, Duration, Radius, and percent labels to `par5`, `ln34`, and `ln12`; the
owned row supplies server-do 30, `damageresist=-par5`, and no cross-skill
modifier. The evidence therefore supports a physical-resistance curse rather
than a second damage transaction or an all-channel multiplier.
Weaken's layered `skillld72` record says it curses a group of enemies and
reduces the amount of damage they inflict. Its SkillDesc joins Target's Damage,
Duration, Radius, and percent labels to `-par5`, `ln34`, and `ln12`; the owned
row supplies server-do 30, `damagepercent=-par5`, and no cross-skill modifier.
That evidence supports a source on each cursed attacker's outgoing physical
weapon damage, not a target-defense or incoming-damage multiplier.
TBL wording establishes intended relationships and player-visible claims;
Skills.txt calc/
parameter fields and owned 1.14d runtime probes remain authoritative for exact
arithmetic, rounding, and ordering. The corrected version-1 TBL codec and a
layer-precedence/source test make this evidence path executable.

The reusable `d2legacy.data.skill_modifiers` decoder now exposes the same
evidence-backed hard-point modifier shape rather than merely reporting it; the
straight-missile family is its first consumer. Fire Bolt's owned row resolves
Fire Ball ID 47 and Meteor ID 56, validates `Param8=16`, and stores those facts
in its immutable family definition. The shared cast lifecycle sums the current
authoritative hard levels and snapshots the resulting percentage on
`d2legacy.skill.cast`; generic projectile construction applies that snapshot to
the level-band damage range. This keeps later equipment/source and modifier
families composable without making the missile or damage systems recognize Fire
Bolt. Integer floor-after-percentage is current high-confidence policy; exact
1.14d rounding and ordering against mastery, resistance, PvP, and other source
families remain an owned-runtime probe.

Fire Ball is the first `missile.straight-impact-area` configuration. The exact
owned missile row requires server travel function 1, server hit function 1,
collision type 3, one-shot collision, `sHitPar1=4`, and
`ExplosionMissile=explodingarrowexp`; the referenced row must be an authored
explosion with a 16-tick presentation lifetime. On first swept contact, the
generic system computes one impact point, selects same-level targets in stable
semantic-ID order, and routes each result through the shared typed fire-damage
boundary. It separately creates `d2legacy.missile.effect`, which carries only
position/location/presentation/lifetime facts and therefore cannot apply damage
again. The localized TBL text calls Fire Ball an explosive sphere of fire and
the owned skill row reuses the generic Fire Bolt/Meteor hard-level modifier at
14% each. Exact 1.14d radius/footprint units, impact rounding, per-target RNG,
and ordering against mastery, resistance, PvP, and other sources remain probes.

Ice Blast is the first `missile.straight-freeze` configuration. Its exact owned
skill row supplies 75 base freeze frames, 5 per level, 10% duration per Glacial
Spike hard point, and 8% cold damage per Ice Bolt, Blizzard, and Frozen Orb hard
point. Missile damage function 4 and localized TBL text establish a direct
freezing hit; `ExplosionMissile=freezingarrowexp1` supplies a distinct 16-tick
presentation aftermath. The cast snapshots both modifiers and the resolved
duration. A nonlethal contact then emits an ordinary
`d2legacy.state.request`; normal/nightmare/hell monster duration divisors reuse
the existing cold policy, while timed-state instances and generic action
filters own checkpointing, refresh, suppression, and expiration. Exact target
resistance/immunity, monster-class cold effectiveness, cross-source state
replacement, PvP freeze-to-chill conversion, and impact/state ordering remain
owned-runtime probes.

Glacial Spike is the first `missile.straight-impact-area-freeze`
configuration. Its exact skill row supplies radius `ln12` from Param1/2 (4 + 0
per level), freeze length `ln34` from Param3/4 (50 + 3 frames per level), and a
3% Blizzard-hard-point multiplier through Param7. Its cold damage receives 5%
per Ice Bolt, Ice Blast, and Frozen Orb hard point. The exact missile row binds
server hit function 13, `frze`, HitFlags 2, one-shot collision, and the
`freezingarrowexp1` center explosion. On first swept contact, the generic area
query emits ordered shared cold-damage results and one ordinary monster-cold
state request for every nonlethal in-radius result; checkpointed timed-state
instances and action filters own the rest. The record also references
`freezingarrowexp2` ejecta, which is intentionally not claimed by the current
single-effect presentation recipe. Exact radius/footprint units, impact and
percentage rounding, per-target RNG, resistance/immunity, monster-class cold
effectiveness, cross-source replacement, PvP chill conversion, secondary
ejecta presentation, and impact/state ordering remain owned-runtime probes.

Teleport is the first `movement.point-relocate` configuration. Its exact row
requires server-do function 27, `warp=1`, SC action, right-skill-only assignment,
`range=none`, interruptibility, 24 base mana, -1 mana per level in authored 8.8
units, and a 1-mana floor. The decoder also pins every owned Levels.txt
`Teleport` value: null level 0 is disabled, ordinary target levels are 1, and
Duriel's Lair ID 73 is the sole policy-2 exception. On the cast effect tick, a
generic system validates the same-level target against world bounds, static
footprint collision, and blocking ECS occupancies; policy 2 additionally uses
the distinct BlockLOS trace. Success mutates the existing position atomically,
stops semantic/raw/forced motion, and composes a checkpointed value-only
`d2legacy.world.relocation_event`. Failure leaves position unchanged and records
an explicit outcome. The current 2-to-line-trace interpretation is conservative
secondary evidence, not a completed target-runtime claim. Exact viewport/
visibility range, type-2 meaning, payment on invalid destinations, nearest-free
fallback, player room-edge sequencing, owned-unit following, and action/
effect/relocation ordering plus remaining presentation semantics remain owned
1.14d probes; the generic SC action and teleport cast overlay now resolve
through the shared action/cue path.

Enchant is the first `state.targeted-timed` configuration. Its exact row binds
server-do function 25, friendly/pet targeting, SC action, 25 base mana plus one
per level, the `enchant` state, duration `3600 + (level-1)*600` frames, and
three aura stats. A generic friendly-target resolver accepts living same-level
players/friendly units and conservatively falls back to the caster for missing
or invalid requests. One source-tagged timed state owns independently keyed
`firemindam`, `firemaxdam`, and `item_tohit_percent` ECS sources; refresh,
replacement, expiration, and checkpoint reconstruction remove or restore the
whole owned set. Fire damage uses the same five-band 8.8 progression and
hard-level modifier machinery as other skills, while `toht` resolves to 20% at
level 1 plus 9 percentage points per additional level. Generic derived stats
consume the attack bonus and generic weapon-melee damage consumes the fire
range, so neither boundary knows Enchant's ID or name. Ranged weapon attacks do
not yet exist, so the localized one-third ranged-fire rule is recorded but not
claimed. Exact party/PvP ally policy, target range/LOS, state replacement among
different casters, action-rate/modifier timing, and ordering against mastery,
equipment fire damage, and other percentage sources remain owned 1.14d probes.

Amplify Damage is the first `state.point-area-curse` configuration and the
first admitted member of the ten-row server-do-30 signature. Its exact row
binds aura filter 3, LineOfSight 4, point-centered radius
`3 + (level-1)`, duration `200 + (level-1)*75` frames, four mana, the
`amplifydamage` curse state, and `damageresist=-100`. Stable same-level hostile
selection includes target footprints and emits one ordinary timed-state plus
physical-resistance source per target. The shared curse group stores skill
level as replacement priority: a lower-level request is rejected, while an
equal/higher request may refresh or replace the current slot. Monsters whose
base physical resistance is at least 100 receive one-fifth of the negative
resistance value, matching the target-version recovered generic curse boundary;
ordinary monsters receive the full value. Shared derived defense and physical
mitigation consume those sources, and checkpoint parity preserves the entire
state/source graph. Exact LineOfSight-4 meaning and radius units, curse
resistance/duration reduction, monster class/mode and boss eligibility, town
and PvP targets, Attract precedence, equal-level same-skill cross-caster source
ownership, presentation timing, and ordering with other resistance sources
remain owned 1.14d probes. The generic SC actor action and start sound are
present; the record-referenced client-only curse missile layers are not.

Weaken is the second `state.point-area-curse` configuration. Its exact row
binds the same server-do/filter/LOS shape to point-centered radius
`9 + (level-1)`, duration `350 + (level-1)*60` frames, four mana, the `weaken`
curse state, and `damagepercent=-33`. The record-shape decoder selects a generic
percentage stat recipe rather than identifying Weaken, and each selected
hostile receives an ordinary timed state owning that stat source. Shared weapon
melee resolves the source against both ends of its physical damage range before
rolling; the deterministic unarmed fixture proves the `-33` source changes its
256-512 raw range to 171-343 before the roll. The existing curse slot makes
Amplify Damage and Weaken mutually exclusive and uses learned level as
replacement priority. Exact
ordering with equipment/skill enhanced damage and strength, monster non-weapon
attacks, hireling/summon attacks, curse resistance, eligibility, PvP,
client-only curse missile layers, and exact overlay semantics remain owned
1.14d probes. The generic SC actor action and start sound are present.

Might is the first `aura.selected-party-stat` configuration. Its exact skill
row binds server-do 65, `aura=1`, `immediate=1`, blank `leftskill`, zero mana,
filter 73731, radius `16 + 2*(level-1)`, `damagepercent` value
`40 + 10*(level-1)`, the `might` owner/target states, and `perdelay=50`.
SkillDesc joins those formulas to `StrSkill4` Damage and `StrSkill18` Radius;
the layered TBL long text says the effect applies to the owner and party.
Blizzard's Expansion skill basics and offensive-aura documentation corroborate
the right-mouse activation, one selected aura per Paladin, and coexistence of
different party auras. The learned-skill record makes left assignment
impossible, and using the right button while selected returns before cast or
mana state exists.

The generic authority owns one emitter component on the selected living player
and one relationship entity per eligible target/state pair. The relationship
co-owns its normal stat source, so range, party, level, death, room-inactive, or
selection changes cannot orphan a modifier. Target/state identity lets
different aura states coexist; same-state candidates select learned level,
then value, then a stable source-ID tie breaker. Authority tests cover solo and
party application, range loss, assignment loss, checkpoint restoration,
right-button no-cast/no-mana behavior, deterministic equal-level suppression,
and stronger-source replacement. Generic melee already consumes the resulting
`damagepercent` source.

Offline state snapshots mark aura relationships without collapsing them into
timed states. A presentation-only cycle retains every gameplay effect but
selects one aura state per target, advancing at the selected row's `perdelay`
converted from the 25 Hz simulation cadence. The shared States/Overlay adapter
then renders Might's pinned back/front DCC pair and normal attachment rules.
`WorldView/v5` applies the same semantic boundary to connected sessions: its
reliable list contains only public target ID, active state ID, and positive
record period, is capped at 512 stable entries, and rejects duplicate target/
state keys. It deliberately has no field capable of carrying emitter identity,
skill ID or level, stat values, radius, party/filter decisions, or same-state
arbitration. `ClientView/v11` validation treats the list as untrusted input.
The disposable client ECS binds each admitted relationship to an existing
local/peer/monster mirror as `d2legacy.presentation.state`; missing mirrors are
retried on later reconciliation, and removal destroys the presentation entity.
Lua consumes that component through the same aura snapshot and cycling path as
offline authority. Multiple distinct state IDs on one target survive the wire
and remain independently active even though only one aura graphic is selected
for a presentation interval.

The exact meaning of all filter bits, non-player owned-unit membership,
application/removal tick order, equal-strength source ownership, whether every
aura's visual cadence follows `perdelay`, and `onsound` lifetime remain
explicit target-runtime work. No Classic, older patch, vanilla server/save, or
community-tool compatibility behavior is introduced.

Defiance is the second `aura.selected-party-stat` configuration. Exact skill
ID 104 has the same server-do 65, right-only, immediate, zero-mana,
`aurafilter=73731`, `ln12` radius, and 50-tick period shape as Might. Its
distinct record recipe is `skill_armor_percent=ln34`, with a 70% level-one
bonus plus 10 percentage points per additional level. The Defiance state row
is an aura, declares the same authored stat, carries
`paladin_aura_defiance`, and selects `aura_defiance_front/back`; SkillDesc and
the layered English TBL label the values as Defense Bonus and Radius and state
that the active aura increases the defense rating of the owner and party.

The family decoder now chooses an explicit reviewed stat recipe: authored
`damagepercent` remains the generic outgoing damage-percent source, while
`skill_armor_percent` becomes the generic defense-percent source already
consumed by derived combat stats. The aura system itself remains unaware of
Defiance. A checkpointed two-Paladin authority test selects Might on one owner
and Defiance on the other, retains four target/state relationships, and proves
both stat sources affect both party members. The fixture's canonical base
defense is 99, so the shared integer percentage resolver produces 178 at
Defiance level one. Real-MPQ tests pin the exact skill, state, SkillDesc, TBL,
Overlay, sound-key, and DCC contracts. This does not broaden the unresolved
target filter, owned-unit, refresh/leave, cross-family cadence, or sound
lifetime claims.

`manifests/skill-behavior-coverage.v1.json` is locked to
`diablo-ii-lod-1.14d-expansion`. Runtime composition consumes the same exact-ID
declarations as the report. The targeted-state decoder independently validates
Enchant's function, flags, state, formulas, damage bands, and Warmth reference;
the manifest alone cannot make another server-do-25 row executable.
The area-curse decoder independently validates each admitted row's function,
filter, state curse flag, area formulas, and supported stat recipe. Amplify
Damage selects additive physical resistance with the recovered one-fifth
immune-breaking boundary; Weaken selects outgoing damage percentage with no
immunity transform. The other server-do-30 rows remain missing until their own
stat/event/AI behavior is implemented.
`d2legacy.data.missile_skills` validates admitted
row graphs into immutable straight-trajectory, area-impact, on-hit-state, or
composed area-impact/on-hit-state
definitions; the earlier Frozen
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
ordering, and remaining state-effect presentation semantics are not yet
implemented. Its SC action/cast overlay and persistent armor overlay use the
shared presentation paths.

`d2legacy.data.radial_missile_skills` validates Nova's exact Expansion 1.14d
server/client function 22/25 shape, three matching `nova` server-missile slots,
targetless SC cast policy, lightning channel, and missile function/collision
contract. It decodes 12 base rays plus 4 per skill level, 15 base mana plus 1
per level in authored 8.8 units, and the five authored elemental-damage growth
bands. A cast creates one projectile entity per evenly spaced direction; all
rays share a deterministic cast ID while retaining unique resident identities.
An independent `d2legacy.missile.contact_lock` ECS entity applies the authored
four-tick cast/target repeat delay without making the projectile or damage
system know Nova. Headless coverage proves one target is damaged once during
that lock, the ring survives checkpoint reconstruction, and all projectiles and
locks expire. The current evenly spaced phase is deterministic Dark Magic
policy: exact initial phase, the missile row's `Accel=-1000`, faster-cast-rate timing,
and the complete meaning/order of `LastCollide=1`, `NextHit=1`, and
`NextDelay=4` still require owned 1.14d vectors, so Nova remains partial.

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

Fire Bolt and Nova have owned-target record evidence. Fire Bolt's hard-point
synergy structure and localized relationships are implemented, while exact
percentage rounding/ordering remains partial. Ice Bolt and other visually or
structurally similar skills remain missing until their own Expansion 1.14d
launch, motion, impact, state, and ordering semantics are verified.

The client-function-30 presentation gap now has a strict executable capture
contract. `curse_presentation_probe` rejects Classic, earlier patches, servers,
community tools, memory inspection, and imported saves; requires fixed-camera
stationary visual logs plus owned-MPQ DCC identification; fingerprints the
capture; normalizes anchor-relative timing/motion; and reports missing empty,
single, and multi-target cases for both Amplify Damage and Weaken. It promotes
no role until the six-case owned Expansion 1.14d matrix is populated.

Next: investigate Blessed Aim as the first selected aura whose exact record
combines an active party stat with a learned-skill passive modifier. Its owned
row exposes active `item_tohit_percent=75+15*(level-1)` and a separate
`skill('Blessed Aim'.blvl)*Param8` passive formula, while layered TBL text
documents the active party attack-rating effect. Verify the latest Expansion
behavior and passive intent before designing a reusable learned passive-source
family; do not silently admit only the visible aura half. If that evidence is
insufficient, select another complete record/TBL shape rather than adding a
Blessed-Aim-specific exception. Do not infer pulse damage/healing, target-filter
breadth, or sound lifetime from record names.
In parallel, capture owned Expansion 1.14d player/hireling/summon entry/leave
observations for `aurafilter=73731`, 50-tick application/removal ordering,
equal-strength same-aura ownership, and `onsound` lifetime. Promote those
results before broadening Might beyond living same-level player party members
or treating `perdelay` as a universal client cadence.

Also populate and review the owned Expansion 1.14d missile-audio matrix, the
client-function-30 visual matrix, and the SC/SQ/FCR/weapon-class timing matrix.
Use the completed audio report to define one generic lifecycle/multiplicity
contract, then produce typed semantic missile cues and extend `EventView/v3`
without shipping collision, damage, lifetime, or target authority to clients.
Add the record-
referenced client-only curse missile layers without guessing the `cltmissilea`/
`cltmissilec` attachment and motion roles. Pin Overlay light,
variant, character restriction, and multi-direction behavior without skill-
specific renderer branches. Live monster/corpse and server projectile/effect
presentation already use the typed `WorldView/v5` boundary, and monster death
audio uses the typed `EventView/v3` boundary; extend the semantic event view
only as those remaining semantic-event consumers become presentation-ready;
never expose arbitrary ECS payloads. Use the completed timing report to
implement SQ sequences, faster-cast-rate and equipped-weapon-class timing, plus
interruption/refund behavior against owned Expansion 1.14d vectors. Then probe
and replace Attack's remaining inferred distance, dynamic-door,
special-unit, and path-to-range edges and confirm its attack-rate breakpoint,
dual-wield, slow, sequence, and mid-action boundaries against owned 1.14d
runtime vectors. In evidence order, finish Frozen Armor's remaining target-sensitive
cold-duration/PvP rules and Nova's radial phase/acceleration/repeat-contact
ordering. Populate Teleport's viewport/range, limited-level, invalid-target,
fallback, owned-unit, and timing vectors. Populate Enchant's ally/PvP target,
range/LOS, multi-caster replacement, animation, overlay, modifier-order, and
ranged one-third-fire vectors; the ranged half first requires a reusable
weapon-projectile attack family. Populate Amplify Damage and Weaken radius/LOS,
curse resistance, eligibility, PvP, replacement-ownership, presentation, and
modifier-order vectors. Continue the same server-do-30 family only where its
complete stat/event consumers exist; Decrepify is the next high-leverage
composition candidate after its velocity, outgoing damage, physical-resistance,
and immunity-breaking interactions are pinned together.
Evidence upgrades and exact-ID declarations land
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
