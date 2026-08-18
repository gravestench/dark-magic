# Skills, states, missiles, and combat actions research

> Architecture note: this document's recovered behavior, data joins, and test
> vectors remain valid. Its proposed Go package layout describes the current
> transitional implementation, not the permanent owner. Under
> [the engine/`d2legacy` boundary](../ARCHITECTURE.md), D2 skill and missile
> policy migrates to authoritative Lua while justified generic scheduling,
> collision, movement, and data-decoding mechanisms may remain in Go.

Status: implementation-oriented research baseline plus a mounted Expansion
1.14d coverage inventory. Older D2MOO 1.10f reconstruction findings remain
version-labeled architecture clues only; they do not define target behavior.
This document describes boundaries and evidence and is not yet a complete
skill-by-skill implementation specification.

This workstream builds directly on:

- [TIMING_RNG_AND_DETERMINISM.md](TIMING_RNG_AND_DETERMINISM.md);
- [ITEM_STATS_AND_AFFIXES.md](ITEM_STATS_AND_AFFIXES.md);
- [COMBAT_DAMAGE_AND_DEATH.md](COMBAT_DAMAGE_AND_DEATH.md);
- the existing Dark Magic `player.use_skill` authoritative command;
- the typed `Skills`, `SkillDesc`, `Missiles`, `States`, and calculation-table catalogs.

## Executive conclusion

Dark Magic already has the correct **input/authority edge** for skills: presentation selects a side and semantic target; the fixed-tick session validates that the selected skill is learned/allowed and records authoritative intent.

The missing layer is the actual deterministic skill transaction:

```text
player.use_skill command
        |
        v
resolve assigned skill + learned level
        |
        v
validate use
  alive / mode / town / target / range / LOS / item / corpse / resource / delay
        |
        v
SkillCast transaction
  start tick
  source snapshot
  target snapshot/semantic target
  cost
  action/mode
  behavior family
        |
        +--> immediate effect
        +--> spawn missile
        +--> apply state/stat-list source
        +--> summon/owned unit
        +--> move/teleport/charge/leap
        +--> operate corpse/object/item
        `--> schedule continuation/events
```

Do not make each skill a bespoke UI callback. Do not require the renderer's current animation frame to decide whether gameplay fires.

## Current Dark Magic state

The renderer-independent `d2legacy` Lua authority now owns:

- `player.assign_skills` and `player.use_skill` admission;
- learned-skill and left/right assignment validation;
- point targets plus optional semantic target IDs;
- definition-driven mana payment and fixed-tick cast lifecycle;
- generic cast-snapshotted hard-point damage modifiers resolved from exact
  `EDmgSymPerCalc` skill references;
- an exact-ID `action.melee` family whose zero-cost Attack configuration uses
  that same lifecycle before reusable approach, hand, animation, and impact;
- a reusable `missile.straight` family covering projectile construction,
  movement, swept contact, channel mitigation, damage, removal, replay, and
  checkpoint continuation;
- a reusable `missile.radial` family that composes one targetless cast into
  ordinary projectile entities sharing a cast identity plus independent ECS
  cast-target contact locks;
- a reusable straight-missile area-impact family that separates the damaging
  radius transaction from a presentation-only ECS aftermath entity;
- a reusable straight-missile on-hit state family that snapshots duration and
  emits ordinary timed-state requests after nonlethal contact;
- a reusable self-state family and timed state-instance lifecycle;
- a selected-right party-aura family whose ECS emitter and target relations
  own ordinary stat sources without manufacturing casts; and
- the shared melee action path.

Fire Bolt is the first explicitly configured expansion 1.14d straight-missile
fixture. It is decoded from Skills.txt/Missiles.txt by generic family code and
does not own a command branch, component schema, system, damage function, or RNG
stream. A second synthetic record-pair test proves decoder reuse without
claiming incomplete retail behavior for Ice Bolt or another named skill. An
opt-in owned-archive test boots the authority against the target expansion
1.14d records so the generic decoder's production contract is checked without
placing copyrighted tables in Git.

When that generic family materializes inside an installed generated-level
plan, it now attaches `d2legacy.world.room_resident` using the common canonical
point-to-room resolver and a deterministic projectile identity. The ordinary
room activation path adds the empty inactive ECS marker; projectile movement
and remaining-lifetime progression query it out. Production-cast checkpoint
coverage proves that the same entity and projectile/location/residency state
resume after room reactivation. The test uses a synthetic extended lifetime to
isolate residency mechanics. Exact Expansion 1.14d missile expiration and
whether any timer ages while its room is inactive remain probe-gated.

That decoder now also preserves the raw target-authored `Missiles.KnockBack`
byte in the immutable definition and checkpointed projectile. Owned 1.14d rows
demonstrate blank, `1`, `33`, and `75`, so the field is neither collapsed to a
boolean nor applied as a guessed percentage. Projectile contact will consume it
only after a target-binary probe pins the roll and damage-result ordering.

The target-locked `skill_behavior_coverage` tool now reads the winning mounted
Skills.txt and Missiles.txt rows and groups every skill by its server start/do
IDs plus referenced missile server-do IDs. Against the owned 1.14d Expansion
archives on 2026-08-18 it reports 357 skill rows, 172 distinct signatures, 12
explicitly admitted configurations, and 345 missing configurations. Every
consumer carries an implementation family or `missing_family: true` plus an
evidence status. Exact declarations live in
`manifests/skill-behavior-coverage.v1.json`, which runtime composition also
consumes. Thus sharing Fire Bolt's function signature cannot silently enable
Ice Bolt or any other row.

The companion `skill_evidence` tool makes localized documentation part of every
exact skill investigation, especially synergy and cross-skill modifier work. It
joins Skills.txt -> SkillDesc.txt -> the requested locale's layered base,
Expansion, and patch TBL records; reports each resolved key, text, winning
virtual source, and printf-style replacement token; and resolves every
`skill('name'.blvl|lvl)` expression back to an exact skill ID. The format path
is executable against owned 1.14d data after correcting the string-TBL decoder
to the authored version-1 header. TBL wording is documentation evidence, not a
replacement for formula/probe evidence: it identifies intended relationships
and labels, while Skills.txt parameters and owned runtime vectors decide exact
values, integer rounding, and event order.

For the current fixtures, the joined report confirms Fire Bolt receives hard-
level fire-damage bonuses from Fire Ball and Meteor. It separately confirms
Frozen Armor receives hard-level modifiers from Shiver Armor and Chilling Armor
for both seconds-per-level duration and freeze-length-per-level, matching the
two owned Skills.txt expressions already decoded by the generic state family.
Ice Blast resolves its Ice Bolt/Blizzard/Frozen Orb cold-damage references,
Glacial Spike freeze-length reference, and localized direct-freeze claim; the
owned parameters remain the arithmetic authority.
Glacial Spike resolves Ice Bolt/Ice Blast/Frozen Orb cold-damage references and
its Blizzard freeze-duration modifier. Its layered TBL records call it a
magical ice comet that freezes or kills nearby enemies and provide radius and
freeze labels; the exact Skills/Missiles formulas remain the arithmetic and
event-shape authority.
Teleport has no cross-skill formula. Its owned `skillld54` TBL key resolves to
the explicit claim that it instantly moves to a destination within line of
sight; Skills.txt and Levels.txt remain authoritative for assignment, signed
mana, dispatch, and per-level permission facts.
Attack resolves to the localized name `Attack` and description `normal attack`
with no cross-skill modifier formula. Nova resolves to the localized name and
descriptions "creates an electrically charged ring" and "to shock nearby
enemies / creates an expanding ring of lightning," with no replacement tokens
or cross-skill modifier formula. This provides target/area intent, not exact
angular, acceleration, collision, or timing arithmetic.

Nova is the first exact Expansion 1.14d `missile.radial` configuration. Its
decoder requires the owned server/client function 22/25 shape, three identical
`nova` missile slots, targetless SC casting, lightning element, and the missile
function/collision/repeat-contact fields. It preserves 12 base rays plus 4 per
level, 15 base mana plus 1 per level, all five elemental damage growth bands,
24 velocity, 13-tick lifetime, presentation fields, and four-tick next-hit
delay in a reusable immutable definition. The generic materializer creates one
ECS projectile per direction with a shared deterministic cast ID. The contact
system represents each cast/target suppression window as its own checkpointed
`d2legacy.missile.contact_lock` entity, so radial repeat policy composes with
the shared direct-damage path rather than becoming a Nova special case. A
synthetic second record fixture proves family reuse, and authority coverage
checks level-one mana, twelve rays, one lightning result during the lock,
checkpoint parity, and lifetime cleanup.

This is partial target behavior. Even angular spacing is a deterministic Dark
Magic policy pending an owned-runtime initial-phase vector. The target missile's
`Accel=-1000` is not yet applied, and the exact action timing and complete
ordering of `LastCollide=1`, `NextHit=1`, and `NextDelay=4` remain probes. No
other nova-like skill is admitted by function or visual resemblance.

The first executable cross-skill damage modifier consumes that joined Fire Bolt
evidence. The fail-closed `d2legacy.data.skill_modifiers` decoder accepts the exact reusable
`(skill('…'.blvl)+…)*par8` shape, resolves every named hard-level reference to a
numeric skill ID, and preserves the owned `Param8=16`. At cast admission, the
shared lifecycle sums the authoritative learned Fire Ball and Meteor hard
levels and snapshots their percentage on the generic cast component. Generic
projectile construction applies the snapshot after the five authored skill-
level damage bands; neither projectile contact nor damage resolution recognizes
Fire Bolt. A two-level-one-synergy fixture therefore snapshots 32% and turns the
owned level-one raw 768-1536 range into the current floor-rounded 1013-2027
range across checkpoint reconstruction.

The formula structure and 16% relationship are owned-record facts and match the
localized tooltip. Floor-after-percentage is high-confidence implementation
policy, not yet a measured 1.14d rounding vector. Ordering against Fire Mastery,
items, PvP conversion, resistance, and other modifier families remains open;
the decoder rejects unreviewed formula shapes rather than treating this as a
general-purpose legacy expression evaluator.

Fire Ball is the first exact `missile.straight-impact-area` consumer. Its owned
Expansion 1.14d rows bind travel function 1 to hit function 1, collision type
3, `sHitPar1=4`, and `ExplosionMissile=explodingarrowexp`; the referenced
explosion row supplies `ExpArrowExplode` and a 16-tick lifetime. Its joined TBL
describes an explosive sphere of fire, while `EDmgSymPerCalc` resolves Fire
Bolt and Meteor hard levels at owned `Param8=14`. Generic swept contact now
computes one impact point, selects same-level targets in semantic-ID order, and
emits one independently rolled shared fire-damage result per in-radius target.
The separately materialized `d2legacy.missile.effect` owns only spatial,
presentation, and lifetime facts. It survives checkpoint reconstruction but
cannot deal damage, so a visual aftermath cannot replay the impact transaction.
A second synthetic record shape proves decoder reuse without admitting another
retail skill.

The area behavior remains partial. Radius 4 is preserved directly from the
owned server-hit parameter, but exact 1.14d conversion against unit footprints,
impact-point rounding, per-target RNG sequencing, and ordering against mastery,
resistance, PvP conversion, and secondary effects require owned-runtime vectors.

Ice Blast is the first exact `missile.straight-freeze` consumer. Its owned skill
row supplies 75 base freeze frames plus 5 per level; `ELenSymPerCalc` resolves
10% per Glacial Spike hard point, while `EDmgSymPerCalc` resolves 8% cold damage
per Ice Bolt, Blizzard, and Frozen Orb hard point. Its missile binds travel
function 1 and damage function 4 to a single-hit collision and references the
16-tick `freezingarrowexp1` presentation explosion. Joined TBL text explicitly
says the projectile freezes its enemy and exposes both cold-damage and freeze-
length synergies.

The generic cast snapshots the resolved effect duration beside the damage
modifier. On nonlethal contact, the projectile emits a `d2legacy.state.request`
with a stable caster/skill source. Existing monster cold-duration divisors,
timed-state instances, action-disable filters, checkpointing, refresh, and
expiration own the result; the missile system contains no Ice Blast branch.
Exact resistance/immunity, monster-class cold effectiveness, cross-source
replacement, PvP chill conversion, and animation/impact ordering remain partial
until owned 1.14d runtime vectors resolve them.

Glacial Spike is the first exact
`missile.straight-impact-area-freeze` consumer. Its owned skill row supplies a
4 + 0-per-level `ln12` radius, a 50 + 3-per-level `ln34` freeze length, 3% per
Blizzard hard point for duration, and 5% cold damage per Ice Bolt, Ice Blast,
and Frozen Orb hard point. The owned missile binds travel function 1 to hit
function 13 with `frze`, HitFlags 2, and one-shot collision. Generic swept
contact therefore reuses the stable area query and shared cold-damage result,
then emits the same ordinary monster-cold state request for each surviving
in-radius target. ECS timed-state/action filters own checkpoint, suppression,
refresh, and expiration; no production branch recognizes Glacial Spike by name
or ID.

The first referenced `freezingarrowexp1` row supplies the current
presentation-only center effect. The second `freezingarrowexp2` ejecta row is
recorded as unresolved rather than silently treated as equivalent. Exact
radius/footprint units, impact and percentage rounding, per-target RNG,
resistance/immunity, monster-class effectiveness, cross-source replacement,
PvP chill conversion, ejecta presentation, and action timing remain target
1.14d probes.

Teleport is the first exact `movement.point-relocate` consumer. Its owned skill
row supplies server-do 27, `warp=1`, SC action, right-skill-only assignment,
`range=none`, and 24 base mana reduced by 1 per level to an authored 1-mana
floor. The generic decoder preserves the signed progression rather than
forcing every skill's level delta nonnegative. It also pins all numeric
Levels.txt Teleport policies; only null level 0 is disabled and Duriel's Lair
ID 73 uses policy 2 instead of the ordinary 1.

At the effect tick, the point-movement system validates same-level bounds,
static footprint collision, and blocking ECS occupancy. Policy 2 additionally
uses the distinct BlockLOS trace. A successful action atomically updates the
existing player position, stops semantic/raw/forced motion, and emits a generic
value-only relocation event; failure emits an explicit outcome without moving.
The localized line-of-sight wording and recovered policy-2 description do not
resolve exact viewport/range, the policy-2 boundary, invalid-target mana
consumption, nearest-free fallback, room-edge order, owned-unit following, or
SC/presentation timing. Those remain owned 1.14d probes.

Ordinary Attack is exact skill ID 0 in the owned Expansion 1.14d Skills.txt,
not a non-skill command. Its row supplies server/client start and do functions
1/1, an A1 weapon action, attack-rate and target/search flags, weapon source
damage, and a literal zero mana cost. The `action.melee` decoder validates that
contract and creates the same definition shape consumed by shared cast
admission. A generic family system then emits the reusable approach, selected-
hand, animation, and impact mechanism. The command no longer contains an ID-0
branch, and the family decoder accepts a second synthetic ID without branching
on name or ID. Its `anim` field now flows through the generic action event. The
actor token, that mode, and equipped weapon class select a record from the
session-pinned AnimData binary; the record's 24.8 speed advances at 25 ticks per
second, its first attack marker schedules impact, and cursor wrap schedules
completion. Owned 1.14d `AMA1HTH` is 13 frames at rate 256 with an attack marker
on frame 8, so the target fixture resolves to +8 impact and +13 completion.
The generic action-rate policy now combines that record with resolved
`attackrate`, `item_fasterattackrate`, and primary/secondary weapon-rate facts.
It uses integer effective IAS `120*IAS/(IAS+120)`, applies the dual-weapon base-
rate average and 15%-175% bounds, then multiplies the AnimData speed and
truncates before scheduling marker/wrap ticks. The owned target generation pins
the table side of that contract: ItemStatCost signed IDs 68/93, the
`UpdateAnimRate` flag on `attackrate`, Properties `swing1/2/3` mapping directly
to `item_fasterattackrate`, and Expansion weapon speeds including Phase Blade
-30 and War Pike +20. Weapon speed enters `attackrate` with the inverse sign.
Equipped affix/socket sources and passive/skill sources share the same named
provenance path; no skill ID owns the formula. Exact 1.14d breakpoint,
dual-wield edge, sequence, shapeshift, slow/chill, and mid-action update vectors
remain target-runtime probes rather than older-version compatibility promises.
Attack now re-resolves current PvE alignment,
life, act/level, footprint reach, and melee-barrier collision before animation
and again at impact. Exact 1.14d distance arithmetic, dynamic-door collision,
special-unit exceptions, PvP hostility, and path-to-range behavior remain
unresolved; other melee skills are not admitted by resemblance.

The timed self-state fixture is selected by exact ID through the same manifest
rather than by the string `Frozen Armor`. The generic decoder validates the
owned row's server function, state/stat names, linear defense formula, and
linear-plus-hard-point-synergy duration shape. Its shared cast path pays 7 mana;
level 1 applies `frozenarmor` for 3000 frames, attaches a 30% named defense
source, adds 300 frames and 5 defense percentage points per skill level, and
adds 250 frames per hard point in each of Shiver Armor and Chilling Armor.
Refresh, checkpoint, expiration, and exact stat-source removal are executable.
These values match Blizzard's official 120-second/30% level-1 table and
10-second duration synergies.

Mana admission is shared behavior, not a property of either fixture. Blizzard's
Expansion skill guide says that lack of mana makes a skill unusable, while its
character guide describes mana consumption when a skill is used. Dark Magic
therefore rejects an underfunded request before creating a cast or effect and
preserves the available mana; executable coverage locks that boundary. Exact
cost formulas, rounding, successful-cast charge timing, and interruption/refund
rules remain probe-gated by behavior family.

The generic decoder now also validates the owned row's `damagedinmelee` event
function 2, Param5/6 freeze formula, Param8 hard-point synergy percentage, the
target `freeze` state, and the armor state's States.txt group. A factual
successful melee hit applies a source-tagged freeze state to a monster attacker;
the row produces 30 + 3 frames per skill level and +5% per hard point in each
named synergy. The immutable Normal/Nightmare/Hell rule applies the official
full/half/quarter cold-length relationship. While active, the generic disabled-
action fact stops the monster AI and motion; expiration restores eligibility.
The same state engine replaces another active state in the owned exclusive group
and removes the displaced stat source exactly.

Evidence remains deliberately partial. PvP must chill instead of freeze, and
target cold resistance/immunity, champion/boss modifiers, exact integer/tick
ordering, presentation, and animation action timing remain absent. These edges
must not be inferred from the older reconstruction.

## Client-function-30 curse presentation probe

Owned Expansion 1.14d Skills/Missiles rows establish the presentation inputs,
but not their client-function-30 attachment roles. Amplify Damage (ID 66)
references `curseamplifydamage` plus `cursecast`; Weaken (ID 72) references
`curseweaken` plus `cursecast`. Their DCCs, frame counts, direction counts,
transparency, and light/color fields are exact data. Those facts alone do not
say whether each instance attaches to the caster, cursor, or affected target,
whether one record translates between anchors, or how instance count changes
with zero, one, or multiple affected targets.

`internal/dev/tools/curse_presentation_probe` makes that missing visual vector
explicit. It accepts only fixed-camera, stationary-actor video/manual frame
logs from an executable-fingerprinted, owned Expansion 1.14d single-player
runtime. Asset identification must use the owned MPQ DCCs. Classic, earlier
patches, vanilla servers, community tools, memory inspection, and imported save
characters are rejected. The analyzer fingerprints the sanitized capture,
requires both record-referenced missile rows to be reported even when absent,
and normalizes first/last frame, anchor-relative start/end offsets, translation,
and follow behavior. It reports—but does not fill—coverage gaps for empty,
single-target, and multi-target cases for both skills.
Use `target_index` only with a `target` anchor (zero-based into the case's
`targets` array); omit it for `caster` and `cursor` anchors. A referenced layer
that is not visible must still be present with `present=false` and an empty
`instances` array.

Start from
`docs/research/probes/curse-presentation-lod-114d-expansion.template.json` and
run:

```shell
go run ./internal/dev/tools/curse_presentation_probe \
  -input /path/to/sanitized-curse-presentation.json
```

No client-function-30 role is promoted into production until the report says
the six-case matrix is complete and the underlying frame log is reviewed.

## Missile travel and impact audio probe

The owned Expansion 1.14d records identify candidate sounds but do not define
their complete playback lifecycle. `Missiles.txt` maps Fire Bolt, Fire Ball,
Ice Blast, and Glacial Spike to distinct travel records and impact records;
Nova has a travel record and no `HitSound`. The joined `Sounds.txt` rows mark
the four straight-missile travel records `Loop=1`, their impact records
`Loop=0`, and Nova's `sorceress_nova` record `Loop=0`. That difference matters:
a looping row may require a retained emitter tied to projectile creation,
movement, collision, expiration, and visibility, while Nova may be one cast
sound or one sound per radial projectile. Record names and comments cannot
establish either policy.

`internal/dev/tools/missile_audio_probe` turns those unknowns into a reviewable
capture contract. It accepts only isolated audio/video frame logs from a
probe-created character in an executable-fingerprinted, owned Expansion 1.14d
single-player runtime. The camera and actors must remain fixed, and sound
identity must be established by comparing the capture to waveforms from the
owned MPQs. Classic, older patches, vanilla servers, community tools, memory
inspection, and imported save characters are outside the schema.

The target-locked matrix compares Fire Bolt expiration with Fire Bolt, Fire
Ball, Ice Blast, and Glacial Spike contact, plus empty and three-target Nova
casts. Every referenced travel/hit record must be reported even when it is not
audible. The analyzer fingerprints the original capture and normalizes
projectile lifetime, contact timing, sound interval timing, and observed
instance count. It preserves the record's `Loop` fact in the report but does
not turn that flag into an inferred start/stop or multiplicity rule. The
permanent real-MPQ test separately pins all referenced `Missiles.txt` and
`Sounds.txt` rows, filenames, group sizes, loop flags, stream flags, impact
effect references, and the immutable record-generation identity.

Start from
`docs/research/probes/missile-audio-lod-114d-expansion.template.json`, replace
the example observation with measured values, add the remaining matrix rows,
and run:

```shell
go run ./internal/dev/tools/missile_audio_probe \
  -input /path/to/sanitized-missile-audio.json
```

The example is a capture-shape illustration, not evidence. No semantic missile
audio event, looping handle policy, or network projection should be implemented
until the report covers all seven cases and its isolated tracks are reviewed.

## Faster-cast-rate and SQ timing probe

The owned Expansion 1.14d records identify the inputs but do not, by
themselves, define the runtime timing formula. `ItemStatCost.txt` names signed
stat ID 105 `item_fastercastrate`; `Properties.txt` maps `cast1`, `cast2`, and
`cast3` to that stat; and the `ModStr4v` key resolves through the owned English
`string.tbl` to the player-visible text `Faster Cast Rate`. Fire Bolt supplies
an ordinary Sorceress `SC` action. Lightning supplies the distinct `SQ` action,
`seqtrans=SC`, sequence 12, and `UseAttackRate=1`. These joins explain which
data participates without assuming how 1.14d turns it into release and
completion frames.

`internal/dev/tools/cast_rate_probe` accepts only a 25 Hz visual frame log from
an executable-fingerprinted, probe-created character in an owned Expansion
1.14d single-player runtime. It rejects Classic, earlier patches, vanilla
servers, imported saves, community-derived stat identification, and memory
inspection. Each case preserves the exact skill action fields, equipped weapon
class, every owned `cast1/2/3` item-property contribution, their summed raw FCR,
the `ModStr4v` key and English localized text, and visually observed cast start/
effect/neutral boundaries. The analyzer rejects a total that cannot be
reproduced from those sources, fingerprints the capture, and normalizes the
absolute frame numbers into effect and completion delays.

The required matrix samples Fire Bolt's `SC` action at raw FCR 0 and paired
8/9, 19/20, 36/37, 62/63, 104/105, and 199/200 values with `HTH`; repeats 0
and 105 with `1HS` and `STF`; and samples Lightning's `SQ`/12 action at 0 and
105 with `HTH`. Pairing both sides makes each observation useful even if 1.14d
disproves a candidate transition. These are required measurements, not embedded
breakpoint results. A complete report still requires human review before a
formula or cap is promoted into the shared action-rate policy. Start from
`docs/research/probes/cast-rate-lod-114d-expansion.template.json` and run:

```shell
go run ./internal/dev/tools/cast_rate_probe \
  -input /path/to/sanitized-cast-rate.json
```

Mid-cast equipment/stat changes, interruption/refund behavior, other classes,
and non-cast sequences remain separate target-runtime probes.

Still open are complete skill-level formulas, target/range/LOS and delay policy,
classification and implementation of the 345 missing configurations,
additional impact/motion families, richer state/stat-source effects, summons,
corpse/item/object actions, and the rest of the behavior-family matrix below.

## Skills.txt dispatch is strongly data-driven

D2MOO's `Skills.cpp` reconstructs large server start/do dispatch tables. `Skills.txt` records identify behavior families by numeric server-function IDs, and those IDs route many named skills through shared implementations.

Examples of shared behavior families visible in the 1.10f tables include:

- ordinary attack and left-hand swing;
- kick/jab/impale and related melee actions;
- projectile firing and throwing;
- curses and buffs;
- auras;
- teleport;
- charge, leap, leap attack, whirlwind;
- corpse explosion and corpse-consuming skills;
- golems, skeletons, revives, druid summons and assassin shadows;
- sentries/traps;
- walls/prisons;
- monster skills and hireling missile behavior;
- stateful charge-up/release skills.

This is important architecture evidence: **skill identity and behavior family are separate concepts**.

Dark Magic should preserve the data-driven relationship while using independent semantic handler names/interfaces. A possible normalized layer is:

```text
SkillDefinition
  stable skill ID
  class / requirements / skill tab
  serverStartBehavior
  serverDoBehavior
  target policy
  cost formula
  delay policy
  calc references
  missile references
  state references
```

The normalized behavior registry must be stable/fingerprinted for replay. Changing the implementation bound to a behavior ID without changing the replay/content generation must not silently produce a different simulation.

## Start versus effect phases

The separate server-start and server-do function families are evidence that a skill can have distinct phases:

1. validate/initiate action;
2. enter animation/action mode;
3. fire an authoritative event at the appropriate action point;
4. perform the effect;
5. possibly schedule further effects;
6. complete/recover.

Dark Magic's future animation-event system should be able to align presentation with these semantic action events, but **gameplay timing must remain authoritative even without presentation assets**.

A headless server must be able to cast every supported skill.

## Target policies

The original runtime contains numerous target checks in shared skills and monster-spawn helpers. The first normalized policy now owns current PvE melee unit targets; broader policies must distinguish at least:

- self;
- unit;
- hostile unit;
- friendly unit;
- point/ground;
- corpse;
- item;
- object/portal;
- direction/vector;
- area around caster;
- area around target;
- summon placement;
- movement destination.

Target validation may additionally need:

- range;
- line of sight;
- collision/footprint clearance;
- town restriction;
- target state/alive/dead restrictions;
- owner/party/hostility;
- item type/equipment state;
- corpse eligibility;
- summon limit;
- skill-specific exceptions.

The optional client `TargetID` is useful presentation context, but the server must re-resolve or validate targets from authoritative world state just as current interaction code already does.

### Current melee target boundary

`d2legacy.gameplay.combat_target` is the single current authority for melee
unit legality. It resolves the semantic ID against the current ECS snapshot,
accepts only player-to-hostile or hostile-to-player alignment, rejects dead or
cross-act/level targets, and evaluates the selected hand's footprint reach plus
the current level's barrier trace. Player Attack applies it before the
AnimData-backed animation; the combat resolver applies it again at the impact
tick, so movement, death, transitions, or collision changes during the swing
cannot preserve a stale hit. Targetless Shift-Attack still swings toward its
point and may choose the nearest currently eligible opponent only at impact.

The collision vocabulary matters. Reconstructed `UNITS_IsInMeleeRange` uses
`COLLIDE_MASK_PLAYER_FLYING`, composed of door and missile-barrier bits, rather
than the visual collision bit. Dark Magic therefore leaves
`map:line_clear(...)` tied to DT1 `BlockLOS` and exposes a separate
`map:barrier_clear(...)` tied to DT1 `BlockJump`. The latter covers immutable
authored terrain now. Dynamic door footprints are not yet authoritative and
remain a blocker for an exact claim.

The current center-distance/continuous-radius reach calculation predates this
slice. Older reconstructed code instead uses an integer unit-size distance
table, range bonus, special tentacle exception, and a footprint-adjusted
barrier trace. That is valuable architecture evidence but cannot establish
Expansion 1.14d identity. Owned 1.14d probes must decide the exact metric,
rounding, unit sizes, exceptions, path-to-range behavior, and dynamic mask
before the remaining range boundary is marked verified.

## Mana and resource costs

Costs should be committed once by the authoritative Lua cast transaction, not
by an individual effect handler or presentation callback.

For ordinary mana skills, insufficient mana prevents the cast from starting and
does not consume the partial balance. This is a shared lifecycle rule rather
than a skill-name branch. The current authoritative Lua cast transaction owns
the check and exact 8.8 fixed-point payment; it consumes a rejected request so
replay cannot turn it into a later cast after regeneration.

Research/implementation needs to distinguish:

```text
base skill cost
level scaling
calculation expression
cost modifiers
minimum/maximum behavior
resource availability
when the cost is charged
whether failed/interrupted actions refund
```

Some skills use non-mana resources or consume items/corpses/charges. A generic `ResourceCost` vocabulary should support those without pretending everything is mana.

Exact formula and rounding behavior remain a probe item.

## Skill delays and cooldowns

Diablo II contains skill delay/cooldown behavior distinct from animation duration. The engine should model it as authoritative future-tick eligibility.

Do not express cooldown as "wait until animation node finishes."

Suggested state:

```text
SkillCooldown
  skill/group identity
  availableTick
  source
```

Whether delay is per-skill, shared by a group, transformed by states, or modified by patch-specific rules must be verified before compatibility claims.

## Calculation expressions and synergies

The typed catalog already includes skill calculation records. Skills and missiles reference formula/calculation slots extensively.

Dark Magic needs one deterministic calculation boundary that can evaluate formulas from explicit inputs such as:

- skill level;
- base stats;
- effective stats;
- character level;
- synergies and prerequisite skill levels;
- missile level;
- difficulty;
- target stats when allowed;
- parameters from the data row.

Do not implement formulas ad hoc in every skill handler. The evaluator should be bounded, deterministic, diagnostic, and testable with source-located errors.

Unknown or unsupported formulas should fail at admission/definition validation where practical rather than halfway through a fight.

## Missile creation evidence

D2MOO's `MISSILES_CreateMissileFromParams` exposes a rich missile state model.

Observed inputs/derived properties include:

- owner and optional origin unit;
- missile class from `Missiles.txt`;
- spawn position and target position/unit;
- velocity from base velocity plus per-level scaling;
- fixed-point velocity representation;
- slow-missile state interaction;
- range plus per-level range and sub-loop adjustments;
- activation frame;
- collision type and separate footprint/movement collision masks;
- acceleration and maximum velocity;
- skill identity and skill level;
- source damage snapshot copied into missile stats;
- owner identity;
- pierce count determined from owner stats/RNG;
- to-hit bonus;
- damage frame rate;
- optional client synchronization facts.

This strongly argues that missiles should be **authoritative entities/state**, not particles.

A future Dark Magic missile entity likely needs:

```text
identity/class
owner/source
skill + skill level
position + fractional motion
path/movement type
velocity/acceleration/max velocity
remaining lifetime/range
activation tick
collision policy
hit memory / pierce budget
damage/stat snapshot or explicit dynamic-source policy
presentation identity
```

## Snapshot versus live-source missile damage

The 1.10f creation path calculates damage data when the missile is created and writes damage stats onto the missile. This is significant evidence that at least substantial portions of missile damage are **snapshotted at creation**.

Dark Magic should not automatically recompute all projectile damage from the caster's current equipment on impact. For each effect family, determine what is snapshotted and what remains dynamically queried.

This is especially important for:

- weapon swaps after projectile launch;
- buffs expiring in flight;
- skill-level changes;
- owner death/despawn;
- minion/trap attribution.

## Missile collision is not generic sprite overlap

Missile creation assigns:

- a footprint collision mask;
- a move-test collision mask based on `CollideType`;
- target/path information;
- optional destructibility;
- specialized path types for some missile families.

Missile collision belongs in authoritative world/collision code and needs semantic contact events.

A useful abstraction:

```text
MissileStep
  -> path movement
  -> terrain/object collision
  -> entity collision candidates
  -> collision policy
       ignore / hit / pierce / bounce / explode / stop / spawn child
  -> CombatResolution or behavior callback
```

Rendering the missile can lag/interpolate independently.

## Pierce

The 1.10f path derives a finite pierce count from owner pierce stats using deterministic RNG and caps the observed index loop at four increments.

Do not implement pierce as an infinite boolean. Preserve an explicit remaining-hit/pierce policy and verify skill-specific exceptions.

## States and stat-list sources

`States.txt` is only the vocabulary/metadata. Runtime state also needs an active source record.

D2MOO's state/stat-list behavior and the combat stun path provide evidence for state instances carrying:

- state ID;
- owner/source identity;
- attached stat modifications;
- expiration game frame;
- removal callback/behavior;
- possible stacking/refresh relationship;
- state mask semantics;
- persistence/death behavior.

Conceptually:

```text
UnitStateInstance
  ID
  Source
  AppliedTick
  ExpiresTick
  StatSourceID
  Parameters
```

A unit may need several sources contributing to the same effective stat while only one visible logical state is toggled. Do not collapse active states into `map[stateID]bool` if source/expiration semantics require more.

## State refresh and stacking

The combat stun path demonstrates refresh behavior: if a stun state already exists, its expiration may be updated/rescheduled instead of adding a duplicate.

Other states may stack, replace, keep the strongest source, keep the longest duration, or allow multiple independent sources. This must be researched by state/skill family rather than applying one universal rule.

The state engine should therefore support a policy layer instead of hard-coding "latest wins" globally.

The first target-backed replacement policy is now executable: States.txt group
1 and Blizzard's statement that only one cold armor can be active cause a newly
applied armor state to replace another member of that group. Replacement emits
an explicit reason and removes the displaced source-tagged stat contribution.
Other groups and same-state/multiple-source policies remain family-specific.

## Auras and continuously applied states

Auras are better modeled as **owned periodic/proximity sources** than as
permanent edits to every nearby entity. Might is the first executable
`aura.selected-party-stat` configuration and establishes that boundary without
a Might-specific command, component, system, or stat consumer.

The owned Expansion 1.14d rows establish the following exact facts for skill ID
98:

- `aura=1`, `immediate=1`, blank `leftskill`, and `range=none` make selection in
  the right skill slot the activation; a click does not start a cast;
- `mana=0` and `lvlmana=0` make that activation free;
- `aurarangecalc=ln12` with Param1/2 gives radius `16 + 2*(level-1)`;
- `aurastat1=damagepercent` and `aurastatcalc1=ln34` with Param3/4 gives
  `40 + 10*(level-1)` percent damage;
- owner and target state are `might`, whose States row is marked `aura=1`,
  references `paladin_aura_might`, and selects the front/back Might overlays;
- `perdelay=50` supplies the record period used by the presentation cycle; and
- the layered English TBL says the aura increases damage done by the owner and
  party, while `StrSkill4`/`StrSkill18` label the displayed damage and radius.

The official Blizzard Expansion [skill
basics](https://classic.battle.net/diablo2exp/skills/basics.shtml) and [Paladin
offensive-aura](https://classic.battle.net/diablo2exp/skills/paladin-offense.shtml)
documentation independently state that Paladin auras operate while readied as
the right-mouse skill, that a Paladin can select only one aura at a time, and
that different auras supplied by party members can coexist. Those control facts
are consistent with the target rows; they do not substitute for unresolved
1.14d runtime timing/filter evidence.

Authority materializes one `d2legacy.skill.aura_emitter` on each living owner
whose right assignment is an admitted aura. It reconciles living same-level
party members inside the record-derived radius into
`d2legacy.skill.aura_effect` relationship entities. Each relationship is
co-composed with its ordinary `d2legacy.stat.source`, so leaving the party,
level, range, life, active room, or selection destroys the modifier and its
provenance atomically. Distinct aura state IDs use distinct target/state keys
and therefore remain simultaneously effective. Duplicate copies of the same
state select the strongest learned level, then the strongest value, with a
stable source-ID tie breaker; this prevents a weaker duplicate from multiplying
the effective stat while keeping replay/checkpoint order deterministic.

Presentation preserves all of those gameplay relationships but chooses only
one aura snapshot per affected unit. It alternates distinct aura graphics at
each selected aura's `perdelay` period converted from the 25 Hz simulation
cadence, while non-aura timed-state graphics remain independent. Thus cycling a
ground graphic never disables a modifier. The Might front/back Overlay rows and
DCC members are pinned by the owned-asset test.

Potential flow:

```text
AuraOwner state
  -> authoritative spatial query at defined ticks
  -> add/refresh source-tagged state/stat source on eligible units
  -> source disappears or unit leaves range
  -> source-tagged contribution expires/removes
```

This design also works for monster auras and some environmental effects.

The current target set is intentionally narrower than `aurafilter=73731`: only
living player party members in the same level are admitted. Hirelings, summons,
other owned units, town/PvP alignment, line-of-sight implications, the exact
application/leave tick relative to `perdelay`, equal-strength source ownership,
owner-vs-target state distinctions, sound lifetime, and connected persistent
overlay projection remain explicit Expansion 1.14d probes. The two-second Might
graphic handoff follows its owned 50-tick period and corroborating visual
evidence; a controlled target-runtime capture must still determine whether
every aura family uses its own periodic field for visual handoff or a separate
client state scheduler.

## Summons and owned units

Skill behavior should not create an unowned generic monster. Summoning must enter the owned-unit system described in [HIRELINGS_AND_OWNED_UNITS.md](HIRELINGS_AND_OWNED_UNITS.md): owner, pet type, limit/replacement policy, inherited/snapshotted stats, AI, and lifecycle are authoritative.

## Movement skills

Charge, Leap, Leap Attack, Whirlwind, Dragon Flight, teleport and similar skills should not bypass the world system by directly setting presentation coordinates.

They need explicit movement/effect policies that can:

- validate destination;
- select collision/path type;
- move over several ticks or atomically teleport;
- apply contact/hit rules;
- handle interruption;
- update authoritative facing/mode;
- emit presentation events.

The path research document records the original path-type vocabulary.

## Proposed Dark Magic boundaries

Possible package ownership:

```text
internal/game/skill
    catalog.go       normalized skill definitions / behavior registry
    cast.go          authoritative cast state machine
    calc.go          deterministic formula adapter
    target.go        target-policy validation

internal/game/state
    state.go         active source-tagged timed states
    policy.go        stack/refresh/removal rules

internal/game/missile
    state.go         authoritative projectile state
    movement.go      path/collision update
    impact.go        behavior/combat dispatch
```

These are candidate package names, not mandatory layout. Repository review should prefer focused existing owners where equivalent abstractions already exist.

## Suggested implementation slices

### S1 — consume `d2legacy.player.skill_intent`

Build a deterministic cast-request stage that resolves the current authoritative assignment/learned level and clears or acknowledges intent exactly once.

No actual damage is required yet.

### S2 — generic cast transaction

Add target-policy validation, mana cost, start/complete tick state, and one basic melee or instant skill behavior.

### S3 — behavior registry

Normalize `SrvStFunc`/`SrvDoFunc`-style data dispatch into explicit trusted behavior handlers with coverage diagnostics. Unsupported behavior IDs must be visible in tooling.

### S4 — state engine

Implement timed source-tagged state application with expiration in authoritative ticks and checkpoint/replay coverage. Start with one refreshable state such as stun/chill.

### S5 — missile vertical slice

Implement one straight projectile end-to-end:

```text
cast -> missile entity -> collision -> combat -> removal
```

Then add pierce, acceleration, child/spawn behavior, and special movement families incrementally.

## Verification backlog

1. Exact mana-cost formulas, integer scale, and charge timing.
2. Exact skill-delay/cooldown semantics and shared delay groups.
3. Cast-speed/action-frame relationship versus server effect timing.
4. Interrupt rules from hit recovery, stun, block, knockback, death, movement.
5. Complete target policy flags in Skills.txt, including exact melee distance,
   dynamic barriers, special units, PvP hostility, and path-to-range behavior.
6. Confirm 1.14d attack-rate breakpoints, dual-wield odd rounding, sequence and
   shapeshift bases, slow/chill inputs, and whether a rate change retimes an
   action already in progress.
7. Formula evaluator opcode/parameter semantics and overflow/rounding.
8. Synergy ordering and soft-point/hard-point distinctions.
9. Weapon contribution and alternate-weapon snapshot behavior.
10. Aura update cadence, range metric, stacking, and source removal.
11. Curse overwrite/priority behavior.
12. Passive skill/stat source lifetime.
13. State persistence through death, level transition, save, shapeshift, and dispel.
14. Missile velocity/range fixed-point stepping and exact lifetime boundaries.
15. Missile collision-type table and collision-mask semantics.
16. Pierce count and hit-memory behavior.
17. Splash/explosion area queries and multi-hit prevention.
18. Homing/guided missile target loss behavior.
19. Missile damage snapshot versus dynamic owner lookup by family.
20. Trap/minion ownership and kill/proc attribution.
21. Corpse skill eligibility and corpse-consumption transaction ordering.
22. Summon limit/replacement policies by PetType.
23. Teleport/charge/leap/whirlwind collision and path-type rules.

## Evidence sources inspected

- User-owned Expansion 1.14d `patch_d2.mpq` Skills.txt and States.txt rows are
  the primary target data for the current self-state slice.
- User-owned Expansion 1.14d ItemStatCost.txt, Properties.txt, Weapons.txt,
  Skills.txt, `AnimData.d2`, and layered TBL records are the primary target data
  for action-rate names, mappings, weapon inputs, action kinds, localized
  meaning, and base animation timing. They identify FCR's inputs but do not
  substitute for runtime timing observations.
- Blizzard's official [Sorceress Cold Spells](https://classic.battle.net/diablo2exp/skills/sorceress-cold.shtml)
  table supplies the published Frozen Armor effect, level vectors, synergies,
  PvP distinction, cold-armor exclusion, and difficulty cold-length warning.
- Blizzard's official [Basic Skill Information](https://classic.battle.net/diablo2exp/skills/basics.shtml)
  says a lack of mana makes an active skill unusable; its
  [Character Information](https://classic.battle.net/diablo2exp/basics/characters.shtml)
  page describes mana as consumed when a skill is used. Together they support
  preserving mana when admission rejects an underfunded attempt.
- D2MOO pinned 1.10f `source/D2Game/src/SKILLS/Skills.cpp`,
  `source/D2Common/src/Units/Units.cpp`, `D2Collision.h`, class/monster skill
  files, `Items.cpp`, `Missiles.cpp`, `MissMode.cpp`, `D2States.cpp`, and `D2StatList.cpp`
  remain older secondary architecture evidence only. For melee they expose
  re-fetch-at-impact, alignment/alive filters, integer unit-distance, the
  door/missile-barrier trace, weapon-speed sign, and effective-rate arithmetic;
  none is treated as a supported older ruleset or a substitute for owned
  Expansion 1.14d vectors.
- Current Dark Magic authority, movement/targeting code, and typed game-data
  catalog define the implementation baseline, not Diablo behavior evidence.
