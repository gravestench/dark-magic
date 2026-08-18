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
- selected-right aura families whose ECS emitter, target policy, relations,
  and checkpointed pulse schedule compose maintained stats, direct party
  effects, or deterministic corpse operations without manufacturing casts,
  plus a bounded connected presentation relationship that reuses the offline
  aura renderer without exporting gameplay policy;
- the complete exact-ID `summon.targeted-corpse` family for Raise Skeleton,
  Raise Skeletal Mage, and Revive. One decoder joins `Skills.txt`, `PetType.txt`,
  Skeleton Mastery, and Summon Resist; one transaction validates and
  revalidates the corpse around mana payment, consumes it once, applies a
  record-selected pet/source-monster materialization policy, and creates an
  ordinary friendly monster with owned-unit, AI, combat, limit, modifier, and
  lifetime facts;
- the complete admitted exact-ID `summon.golem` family for Clay, Blood, Iron,
  and Fire Golem. One decoder joins all four Skills rows to Golem Mastery,
  Summon Resist, PetType, SkillDesc localized synergy keys, and the granted
  Holy Fire row. One effect transaction creates an ordinary friendly monster,
  enforces the shared PetType limit, distinguishes hard synergy points from
  effective cast level, and revalidates Iron Golem's metal ground item before
  replacement and consumption. Generic ECS reactions cover Clay slowing,
  Blood life exchange, Iron item provenance/properties, thorns, fire absorb,
  and scheduled Holy Fire pulses;
- the shared melee action path.

`MonStats2.revive` becomes the empty `d2legacy.monster.revivable` capability,
separate from general corpse selection and mutable corpse usability. Skeletal
Mage's record-calculated `NecromageMissile` grant is likewise an ECS fact so a
future generic monster-skill executor can consume it. Exact mage elemental
variant/projectile execution and the full revived-monster AI/skill inheritance
surface remain downstream monster-skill work; their absence is not hidden by
the corpse-summon family declaration.

Fire Bolt remains an explicitly configured Expansion 1.14d acceptance fixture,
not a standalone implementation. It is decoded from Skills.txt/Missiles.txt by
generic family code and does not own a command branch, component schema, system,
damage function, or RNG stream. Straight, radial, area-impact, and on-hit-state
records compose shared cast/projectile/contact mechanisms. Opt-in owned-archive
tests boot the authority against target records without placing copyrighted
tables in Git.

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
archives on 2026-08-18 it reports 357 skill rows, 172 distinct signatures, 33
explicitly admitted configurations, and 324 missing configurations. Every
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
`skill('name'.selector)` expression back to an exact skill ID while preserving
the authored selector (including `.blvl`, `.lvl`, `.edns`, and `.edmn`). The
format path is executable against owned 1.14d data after correcting the string-
TBL decoder to the authored version-1 header. TBL wording is documentation
evidence, not a replacement for formula/probe evidence: it identifies intended
relationships and labels, while Skills.txt parameters and owned runtime vectors
decide exact values, integer rounding, and event order.

For the current fixtures, the joined report confirms Fire Bolt receives hard-
level fire-damage bonuses from Fire Ball and Meteor. It separately confirms all
three cold armors name the other two family members as bonus sources. Frozen
Armor uses those hard levels for duration and freeze length; Shiver Armor and
Chilling Armor use them for duration and cold damage. The localized `Sksyn`
heading retains its `%s` replacement token, while Skills formulas remain the
arithmetic authority.
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

The three Sorceress cold armors are selected by exact ID through the same
manifest, not by skill names in runtime policy. One decoder validates their
common server function, exclusive States group, defense/duration formulas,
mana, reaction event/function, damage bands, hard-point formulas, overlays,
sounds, and Chilling Armor missile. Effective cast level drives ordinary level
scaling; only `.blvl` expressions read hard skill points. This prevents bonuses
from equipment from incorrectly multiplying synergy contributions.

Frozen Armor level 1 pays 7 mana, applies `frozenarmor` for 3000 frames, adds a
30% named defense source, gains 300 frames and 5 defense percentage points per
skill level, and gains 250 frames per hard point in Shiver Armor and Chilling
Armor. Shiver Armor and Chilling Armor use the same state transaction with
their own owned vectors and cross-family hard-point damage/duration synergies.
Recast refresh, group replacement, checkpoint restore, expiration, and exact
stat-source removal are executable for the family.

Mana admission is shared behavior, not a property of either fixture. Blizzard's
Expansion skill guide says that lack of mana makes a skill unusable, while its
character guide describes mana consumption when a skill is used. Dark Magic
therefore rejects an underfunded request before creating a cast or effect and
preserves the available mana; executable coverage locks that boundary. Exact
cost formulas, rounding, successful-cast charge timing, and interruption/refund
rules remain probe-gated by behavior family.

The reaction pair is record-driven. `damagedinmelee`/2 requires a successful
damaging hit before Frozen Armor applies its 30 + 3 frames per level freeze and
5% hard-point length bonuses. `attackedinmelee`/3 makes Shiver Armor roll and
mitigate cold damage on every melee attempt, including a miss.
`hitbymissile`/1 makes Chilling Armor launch the owned `chillingarmorbolt`
through the ordinary projectile, collision, damage, chill, audio, and
presentation path.

Cold resistance independently scales state length and a raw resistance of 100
suppresses cold damage and chill. Monster state length uses the immutable
Normal/Nightmare/Hell full/half/quarter rule. Players and monster entities with
the empty `monster.freeze_immune` classification receive `cold` chill rather
than the action-disabling `freeze` state; boss population currently installs
that marker, and later champion/unique generation can reuse it without changing
the armor system. This follows Blizzard's published rule that champions,
uniques, super uniques, and bosses can only be chilled. Exact action-frame and
same-tick cross-system ordering remain target-runtime probes.

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
classification and implementation of the 324 missing configurations,
additional impact/motion families, richer state/stat-source effects, summons,
corpse/item/object actions, and the rest of the behavior-family matrix below.

Fire Golem's authored `MonStats.deathDmg` flag is established, and Blizzard's
Expansion documentation confirms that replacing the golem explodes the old
one. The available recovered 1.10f general death-damage routine conflicts with
that target-facing description about affected unit classes. Exact 1.14d
damage, channels, radius, target filter, and replacement/death ordering remain
probe-gated rather than inheriting older behavior silently.

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
`d2legacy.skill.aura_effect` relationship entities. Each relationship owns a
keyed set of ordinary `d2legacy.stat.source` entities, so leaving the party,
level, range, life, active room, or selection removes every modifier and its
provenance in the same reconciliation pass. Distinct aura state IDs use
distinct target/state keys
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

Connected presentation uses `WorldView/v5`, not the short-lived semantic-event
tail, because an aura relationship persists until authority removes it. The
reliable projection carries at most 512 stable records containing only public
target ID, state ID, and positive record period. `ClientView/v11` rejects
oversized lists, malformed periods, and duplicate target/state pairs. The
client binds each record to an existing disposable unit mirror through
`d2legacy.presentation.state`; Lua's ordinary state snapshot then feeds the
same cycle and Overlay lookup used offline. Distinct state IDs on one target
remain separate records so their gameplay effects can coexist while the visual
selector displays one aura graphic at a time. Source identity, skill ID/level,
stat values, radius, party/filter eligibility, and same-state arbitration have
no network fields and remain authoritative.

Potential flow:

```text
AuraOwner state
  -> authoritative spatial query at defined ticks
  -> add/refresh source-tagged state/stat source on eligible units
  -> source disappears or unit leaves range
  -> source-tagged contribution expires/removes
```

This design also works for monster auras and some environmental effects.

Defiance is the second exact member of the selected party-stat family. Its
Expansion 1.14d rows share Might's server-do 65, selected-right/immediate,
zero-mana, filter 73731, `ln12` radius, and 50-tick period contract. The
distinct recipe is `skill_armor_percent=ln34`, with Param3/4 producing 70% at
level one plus 10 percentage points per additional level. Its state row is an
aura with the same authored stat, the `paladin_aura_defiance` sound key, and
front/back Defiance overlay references. SkillDesc joins `ln34` and `ln12` to
the `StrSkill31` Defense Bonus and `StrSkill18` Radius TBL labels; the long and
short localized text explicitly describe increased defense for the owner and
party while active.

The decoder resolves reviewed authored stat names to the engine vocabulary:
`damagepercent` remains outgoing damage percent, while
`skill_armor_percent` becomes the existing generic `defense` percent source.
The aura reconciler therefore needs no Defiance branch. A two-owner checkpoint
test selects Might and Defiance simultaneously and preserves both target/state
relationships on both party members; the ordinary derived-stat system consumes
Defiance. Real-MPQ tests also pin the Defiance overlay rows and DCC members.
This evidence does not resolve filter membership beyond current living player
party targets, exact leave/refresh ordering, general visual cadence, or sound
lifetime.

Blessed Aim is the third exact selected party-stat aura and the first admitted
row with an inseparable learned passive. Its active fields produce
`item_tohit_percent=75+15*(level-1)` for eligible party targets. Its separate
`passivestate=penetrate`, `passivestat1=item_tohit_percent`, and
`skill('Blessed Aim'.blvl) * par8` with Param8=5 encode a personal 5% attack-
rating bonus per hard point. The joined SkillDesc/TBL evidence labels the
active values Attack and Radius and describes the party effect. Blizzard's
[Expansion Blessed Aim reference](https://classic.battle.net/diablo2exp/skills/paladin-offense.shtml)
states the complementary behavior: 5% attack rating per hard point while the
aura is not active.

The decoder admits a reviewed self-hard-level/parameter formula shape and an
`item_tohit_percent` stat recipe; it does not branch on ID 108 or the English
name. A separate learned-passive ECS reconciler composes the personal source
onto the existing learned-skill entity. When that same skill is selected on
the right, the personal source is removed before the selected-party aura
becomes the active contribution; switching away restores it. Authority tests
pin 10% passive and 90% active values at level two across checkpoint restore.
Owned-archive tests pin both state rows, formula fields, SkillDesc/TBL labels,
sound key, Overlay rows, and DCC members. Item-granted soft levels and exact
external-aura-plus-passive ordering remain target-runtime probes.

Resist Fire is the fourth exact selected party-stat aura and the first admitted
relationship with multiple active stats. Its owned Expansion 1.14d row declares
`fireresist=dm34`, `maxfireresist=skill('Resist Fire'.blvl)`, and inactive
`maxfireresist=skill('Resist Fire'.blvl)/2`. The active diminishing formula uses
Params 3/4 (35/150) and staged integer arithmetic: truncate
`110*level/(level+6)` before interpolating between the parameters. The resulting
level-1..20 vector exactly matches Blizzard's published 52..131 table. The
[official defensive-aura reference](https://classic.battle.net/diablo2exp/skills/paladin-defense.shtml)
also states +1 active maximum resistance per hard point, half that value rounded
down while inactive, no passive ordinary resistance, a 95% cap, and no passive
increase from item-granted soft levels.

One aura state can therefore own more than one modifier. The generic reconciler
now derives a stable set of stat-source entities keyed by target, aura source,
and stat, instead of relying on one modifier co-composed with the relationship.
The relationship remains the lifecycle authority: losing eligibility removes
every stale keyed source in the same reconciliation pass, and duplicate source
keys collapse deterministically. The selected-aura decoder recognizes reviewed
linear, diminishing, and self-hard-level formulas; the learned-passive decoder
uses a generic integer numerator/divisor recipe. Neither path recognizes ID 100,
the English name, or fire as a skill identity.

At level three, checkpoint coverage pins 76 active fire resistance, +3 active
maximum fire resistance, and +1 inactive maximum fire resistance. Shared fire
mitigation already applies the 95% maximum cap. Owned archive tests pin the
active and passive states, SkillDesc/TBL values, sound key, persistent/cast
Overlay rows, and DCC members. The connected view remains unchanged because it
projects only public target/state/period relationships, never modifier facts.
Hireling/summon filter breadth, item-source ordering, and future hard-versus-soft
level model integration remain target-version probes.

Resist Cold is the fifth exact selected party-stat aura. Its own target records
pin `coldresist=dm34`, `maxcoldresist=skill('Resist Cold'.blvl)`, inactive
`maxcoldresist=skill('Resist Cold'.blvl)/2`, states 4/182, cold-specific
SkillDesc/TBL text, `paladin_aura_resistcold`, and persistent/cast Overlay/DCC
members. Blizzard's same [Expansion defensive-aura
reference](https://classic.battle.net/diablo2exp/skills/paladin-defense.shtml)
pins the full 52..131 active resistance vector, +1 active maximum resistance
per hard point, floor-half inactive maximum resistance, 95% cap, and exclusion
of item-granted soft levels. ID 105 remains an explicit manifest decision; the
matching Fire row cannot admit it by resemblance.

This slice also supplies the previously missing gameplay consumer. The shared
defense component now owns base/effective cold resistance and maximum cold
resistance; player entry preserves the adapter's durable cold-resistance fact,
and derived stats resolve both named sources. Elemental mitigation selects its
resistance/maximum fields from a small channel table, then applies the existing
−100/current-maximum/95% clamps and integer percentage stage. A level-three
checkpoint test proves 76 active cold resistance, maximum 78, +1 inactive
maximum resistance, and 1000 raw cold damage reduced to 240. Cold-duration and
freeze/chill rules, absorb, PvP, item-source ordering, and future hard-versus-
soft level integration remain independent target-version probes.

Resist Lightning is the sixth exact selected party-stat aura. Its independent
owned evidence pins skill ID 110, `lightresist=dm34`,
`maxlightresist=skill('Resist Lightning'.blvl)`, inactive
`maxlightresist=skill('Resist Lightning'.blvl)/2`, states 5/183, localized
lightning-protection text, `paladin_aura_resistlightning`, and the
`aura_resistlight`/`cast_resistlight` Overlay and DCC members. Params 3/4 are
35/150, so the reviewed staged integer recipe produces Blizzard's published
52..131 active vector. The same [Expansion defensive-aura
reference](https://classic.battle.net/diablo2exp/skills/paladin-defense.shtml)
pins the +1 active/floor-half inactive hard-point maximum-resistance rule, 95%
cap, and soft-level exclusion. Similar neighboring rows justify decoder reuse,
not exact-ID admission.

The generic consumer now includes base/effective lightning resistance and
maximum lightning resistance in combat defense, preserves the character-entry
value, resolves both ordinary stat-source keys, and maps lightning through the
shared elemental mitigation table. A level-three checkpoint proves 76 active
resistance, maximum 78, +1 inactive maximum, and 1000 raw lightning damage
reduced to 240. Lightning absorb, PvP conversion, item-source ordering, and the
future hard-versus-soft level model remain separate target-version probes.

Salvation is the seventh exact selected party-stat aura and the first admitted
relationship with three same-progression stats. Its independently owned ID 125
row declares fire, cold, and lightning resistance through `dm34` with Params
3/4 of 50/120 and no passive or maximum-resistance fields. State 8 records
`lightresist` as its representative stat, so the generic decoder now validates
that a state stat matches any authored aura stat rather than assuming column
one. Empty and unrelated state stats remain rejected.

The owned SkillDesc and layered TBL text resolve “Resist All” to fire, cold, and
lightning specifically, with no poison effect. Blizzard's [Expansion defensive-
aura reference](https://classic.battle.net/diablo2exp/skills/paladin-defense.shtml)
independently publishes the 60..108 level vector and says distinct Paladin auras
from multiple Paladins stack. At level three, the relationship owns three
ordinary sources worth 75 each; the existing derived-stat and mitigation
consumers reduce 1000 fire, cold, or lightning damage to 250 and remove all
three sources atomically when selection changes. Owned `resistall` state,
`paladin_aura_salvation` sound, front/back/cast Overlay rows, and DCC members
pin presentation independently of the neighboring aura records.

Vigor is the eighth exact selected party-stat aura and the first whose three
ordinary sources are consumed by both locomotion and stamina systems. Its owned
ID 115 row declares `staminarecoverybonus=ln34`,
`skill_staminapercent=ln34`, and `velocitypercent=dm56`. Params 3/4 produce the
linear stamina vector 50,75,...,525; Params 5/6 produce the independently
published movement vector 13,18,22,25,28,30,32,33,35,36,37,38,39,40,40,41,41,
42,42,43. Blizzard's [Expansion defensive-aura
reference](https://classic.battle.net/diablo2exp/skills/paladin-defense.shtml)
describes the same movement, maximum-stamina, and recovery effects.

State 41 represents the effect with `maxstamina`, while Skills.txt authors
`skill_staminapercent`. ItemStatCost.txt supplies the missing semantic edge:
that authored percentage stat's operation targets `maxstamina`. The decoder
uses this metadata to validate the relationship instead of weakening state
validation or recognizing Vigor. It also derives any reviewed `dmXY` recipe's
parameter columns, so `dm56` shares the progression evaluator without being
treated as `dm34`.

At level three, the relationship owns +100% stamina recovery, +100% maximum
stamina, and +22% velocity sources. Existing consumers double the fixture
Paladin's 89 base stamina, apply doubled fixed-point idle recovery, and raise
production walk velocity from 6 to 7.32. Changing selection removes all three
sources atomically, and checkpoint restore preserves the result. Owned locale
text, `paladin_aura_stamina`, the `staminafront`/`staminaback` Overlay rows, and
their DCC members pin intent and presentation independently.

Thorns is the ninth exact selected party-stat aura and the first relationship
whose ordinary stat source is consumed reactively after another entity commits
melee damage. Its owned ID 103 row declares server-do 65, filter 73731, state
36, `thorns_percent=ln34`, Params 3/4 of 250/40, zero mana, and a 50-tick
period. The corresponding state intentionally leaves its representative stat
blank, and Skills.txt intentionally leaves `immediate` blank. The decoder
accepts those absences only through reviewed recipe flags on
`thorns_percent`; unrelated aura rows cannot bypass either validation rule.
SkillDesc and layered TBL text identify returned damage and radius, while the
[official Expansion skill reference](https://classic.battle.net/diablo2exp/skills/paladin-offense.shtml)
publishes the 250,290,...,1010 level vector and restricts the trigger to actual
melee hits against affected party members.

The aura relationship owns an ordinary `thorns_percent` stat source. A generic
ECS reflection consumer observes factual melee attack, combat, and typed damage
components after the original hit resolves, then adds an empty observation
marker before doing any work. Misses, missile hits, and attacks with no applied
physical damage cannot reflect. For supported PvM, the reflected basis is the
lesser of post-defender-mitigation physical damage and actual committed damage;
the percentage result then enters the shared physical-damage policy against the
attacker. At level three, 20 raw physical becomes 10 after defender resistance,
330% produces 33 returned physical, and 50% attacker resistance commits 16.5.
The ordinary damage/death result attributes a lethal reflection to the aura
bearer, and checkpoint, duplicate-step, and aura-removal coverage prove the
same source and observation lifecycles without a Thorns callback.

That ordering is **high-confidence recovered behavior**, not yet target-runtime
verification. Pinned recovered code applies Thorns after successful melee and
defender mitigation, caps the basis by remaining life, and sends the returned
physical damage through the attacker damage pipeline
([damage-result ordering](https://github.com/ThePhrozenKeep/D2MOO/blob/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Game/src/UNIT/SUnitDmg.cpp#L2377),
[Paladin reflection calculation](https://github.com/ThePhrozenKeep/D2MOO/blob/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Game/src/SKILLS/SkillPal.cpp#L1555)).
Those sources principally reconstruct 1.10 behavior. Player/hireling attackers
are therefore excluded until the target 1.14d hostility and one-eighth PvP
rule are verified. The owned front/back aura overlays and `hit_thorns` overlay
are pinned, but the hit-reaction graphic is not emitted yet.

Prayer is the first `aura.selected-party-periodic` configuration. It reuses the
selected-aura emitter, party/radius relationship, state arbitration, and
presentation authority, but its authored `hitpoints=edns` field represents a
checkpointed direct effect rather than a maintained stat source. The exact ID
99 row pins server-do 65, filter 73731, `ln12` radius, Prayer owner/target state,
50-tick period, and 8.8 fixed-point mana progression. Its five-band `EMin`
progression produces the official level 1-20 healing vector. Layered TBL text
identifies life regeneration for the owner and party, labels healing/radius/
mana, and says the effect occurs every two seconds. The owned front/back
Overlay rows independently pin the persistent Prayer presentation assets.

The periodic-aura decoder contributes immutable pulse facts to the generic
emitter. `d2legacy.skill.aura_pulse` keeps the source, cost, period, and next
tick durable on the owner; ordered `aura_pulse_effect` entities carry each
authored operation and evaluated value. A separate ECS consumer gathers current
aura relationships, orders targets by stable player identity, clamps each heal
to maximum life, and makes one all-or-nothing resource transaction. A pulse
with insufficient mana heals nobody and consumes nothing. A funded pulse
spends the full 8.8 cost only if at least one eligible target gains life, so an
all-full-health party also consumes nothing. The schedule advances before
application, which prevents duplicate effects after repeated stepping or
checkpoint restore. Switching the selected aura removes the schedule and its
effect entities through the same source lifecycle.

Blizzard's [Expansion defensive-aura
reference](https://classic.battle.net/diablo2exp/skills/paladin-defense.shtml)
corroborates party healing and publishes the healing and displayed-mana
vectors. Pinned recovered server code corroborates full-cost admission,
change-gated payment, capped direct-stat changes, and the global-period-plus-one
tick schedule
([basic aura execution](https://github.com/ThePhrozenKeep/D2MOO/blob/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Game/src/SKILLS/SkillPal.cpp#L227),
[resource and periodic helpers](https://github.com/ThePhrozenKeep/D2MOO/blob/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Game/src/SKILLS/Skills.cpp#L1157)).
That recovered source principally reconstructs 1.10, so exact 1.14d pulse phase
and payment/regeneration event ordering remain probe-gated.
Hirelings and summons are also excluded until the current owned-unit target
filter exists, despite Blizzard's Prayer page including hirelings.

The evidence report now preserves generic cross-skill selectors, not just hard-
level selectors. It resolves Cleansing and Meditation's
`skill('Prayer'.edns)` Skills formulas and `skill('Prayer'.edmn)` SkillDesc
formulas to exact Prayer ID 99 without admitting either dependent skill. That
evidence defines the dependency edge for their future reusable behavior
families; it does not guess the selector arithmetic.

Cleansing is the second exact periodic-aura configuration and the first to
compose multiple authored operations on one schedule. Its ID 109 row pins zero
mana, state 45, the shared filter/radius/50-tick contract,
`item_poisonlengthresist=100-dm34`, and
`hitpoints=skill('Prayer'.edns)`. Params 3/4 of 30/90 produce the displayed
level 1-20 duration-reduction vector 39..80 and the corresponding current-
remaining-duration multiplier 61..20. The linked selector resolves to exact
Prayer ID 99; its heal uses the owner's current learned Prayer level and that
row's five-band `EMin` progression, independently of Cleansing's own level. No
learned Prayer means zero healing while duration reduction continues.

On every due pulse, each target's active poison and admitted duration state is
rescheduled to `tick + floor((expires-tick)*remainingPercent/100)`. This scales
the current remaining duration again on later pulses instead of installing a
permanent percentage stat. Effects retain Skills.txt column order, targets and
state instances use stable identity/source order, and the zero-cost Cleansing
schedule never enters a mana transaction. A level-three checkpoint scenario
pins 49% remaining duration for owner and party, curable-curse expiry, repeated
poison compounding, Prayer level-three healing of four, zero healing after
Prayer removal, unrelated Battle Cry preservation, and source cleanup on aura
selection change.

Blizzard's same official Expansion page explicitly states that Cleansing
reduces poison, curse, and shrine duration for all party members and grants the
same healing as Prayer without mana cost. Owned States rows identify curable
curse states and the `shrine_*` duration family. The pinned recovered callback
corroborates current-remaining integer scaling for poison/curable curses and the
linked Prayer operation, but principally reconstructs 1.10 and does not expose
the official shrine branch. Dark Magic therefore includes the target-documented
shrine family while leaving unrelated non-curable curse states unchanged;
exact 1.14d shrine callback classification and same-tick expiry-event order
remain explicit probes. Owned `cleansingfront`/`cleansingback` rows and DCC
members pin persistent presentation separately.

Meditation is the third exact periodic-aura configuration and the first to
compose a maintained stat with a direct-effect schedule. Its ID 120 row pins
filter 73729, zero mana, state 48, the shared radius/50-tick contract,
`manarecoverybonus=ln34`, and `hitpoints=skill('Prayer'.edns)`. Params 3/4 of
300/25 produce the official level 1-20 mana-recovery vector 300..775. The
periodic decoder preserves both authored operations on one definition: ordinary
keyed stat sources maintain the recovery bonus on current aura relationships,
while the existing pulse evaluates Prayer healing from the owner's learned
Prayer level. Removing Prayer reduces only the heal to zero; removing
Meditation removes the stat source, pulse, and presentation relationship.

Mana now advances through its own 25 Hz resource consumer. For each player it
computes `base=max(1,floor(maxManaRaw/(25*CharStats.ManaRegen)))`, applies
`manarecoverybonus+100` with integer truncation, adds flat `manarecovery`, and
clamps the 8.8 result to zero/maximum. Narrow fixtures may omit `ManaRegen`, but
mounted target rows pin it; Paladin authors 120. This makes Meditation a named
stat source consumed by ordinary resource simulation rather than a skill-owned
mana mutation.

The same consumer makes Prayer's recovered `STATE_NOMANAREGEN` behavior
meaningful. A useful paid pulse owns a
`d2legacy.resource.mana_regen_suppression` relationship on its caster. Base
regeneration remains suppressed until a later pulse is ineffective or
underfunded, or the aura is deselected; independently authored flat
`manarecovery` still applies. Source ownership permits future non-aura
suppressors without teaching mana regeneration skill identities.

Blizzard's official Expansion page confirms party mana recovery, the 300..775
vector, free Prayer behavior, and that Meditation does not work on hirelings.
That matches filter 73729 omitting the monster-unit bit present in 73731. Owned
SkillDesc/TBL rows independently label the mana recovery and Prayer dependency;
owned State/Overlay/DCC rows pin sound and persistent presentation. Exact 1.14d
stat-regeneration ordering relative to same-tick casts, aura refresh, and
resource spending remains a runtime probe because the arithmetic/order source
principally reconstructs 1.10.

Redemption is the first `aura.selected-corpse-periodic` configuration. Its ID
124 row pins server-do 82, owner state 50, filter 4354, a 50-tick schedule, a
constant `ln12` radius from Params 1/2, `dm34` per-corpse chance from Params
3/4, and equal `ln56` life/mana recovery from Params 5/6. The owned English TBL
records call the operation an attempt to redeem slain enemies for life and
mana, and label the recovery, chance, radius, and percent values. The official
[Expansion defensive-aura reference](https://classic.battle.net/diablo2exp/skills/paladin-defense.shtml)
independently publishes the level 1-20 chance/recovery vectors, says the effect
benefits only the aura owner, and states that a redeemed corpse cannot be
resurrected.

The reusable corpse-aura decoder turns those fields into a target policy,
chance progression, and ordered `restore_owner_life`, `restore_owner_mana`, and
`consume_corpse` effects. Monster construction maps authored
`MonStats2.corpseSel` to an empty immutable ECS capability; death state owns the
separate mutable `corpse_usable` fact. Each due pulse excludes town, inactive,
other-level, out-of-radius, non-capable, and already-consumed entities, sorts
the remaining corpses by stable spawn identity, and rolls the named
checkpointed RNG stream once per candidate. A success clamps both owner
resources before making the corpse unusable, even when both resources were
already full, then emits a semantic pulse-result entity for later presentation.
Selection removal deletes the pulse and effects through the existing aura
source lifecycle. Integration coverage pins eligibility, stable ordering,
failure/success, full-resource consumption, town exclusion, deselection, and
checkpoint continuation without a skill-name or ID branch.

Owned State, Overlay, Missiles, DCC, and TBL records pin the persistent owner
aura and the `redeemed`/success/failure presentation vocabulary, but runtime
presentation does not map the semantic result to those assets yet. Exact 1.14d
radius-unit conversion, corpse-roll sequencing against unrelated RNG work,
same-tick death eligibility, and visual/sound timing remain explicit probes.
Pinned recovered server code corroborates corpse selection, per-corpse chance,
life-then-mana clamping, and post-success corpse invalidation, but principally
reconstructs 1.10 and therefore remains architecture evidence rather than the
target authority.

The current target set is intentionally narrower than `aurafilter=73731`: only
living player party members in the same level are admitted. Meditation's 73729
filter and official no-hireling rule match that player-only breadth. For 73731
auras, hirelings, summons, and other owned units remain explicit gaps. Town/PvP
alignment, line-of-sight implications, the exact
application/leave tick relative to `perdelay`, equal-strength source ownership,
owner-vs-target state distinctions, and sound lifetime remain explicit
Expansion 1.14d probes. The two-second Might graphic handoff follows its owned
50-tick period and corroborating visual evidence; a controlled target-runtime
capture must still determine whether every aura family uses its own periodic
field for visual handoff or a separate client state scheduler.

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
- Blizzard's official [Monster Bonuses](https://classic.battle.net/diablo2exp/monsters/bonus.shtml)
  page supplies the champion/unique hard-freeze exclusion; the corresponding
  Super Unique and Boss pages publish the same chill-only rule.
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
