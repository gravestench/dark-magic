# Hirelings, mercenaries, pets, and owned-unit research

Status: implementation-oriented research baseline. Runtime findings are primarily D2MOO 1.10f and must remain version-labeled until original-game/owned-save probes promote them.

Hirelings and summons are not separate mini-games. They reuse monster movement, AI, skills, combat, item equipment, states, death, and presentation, while adding **owner relationship, population limits, persistence, and attribution**.

## Executive conclusion

Dark Magic should build one general owned-unit relationship and then layer hireling-specific persistence/progression on top.

Conceptually:

```text
Owner
  |
  +-- OwnedUnit relation
       type/category
       runtime entity
       durable identity (optional)
       limit/replacement policy
       owner-attribution policy
       transition/leash policy
       death policy
       persistence policy

Owned runtime unit
  -> ordinary monster-like stats / AI / skills / equipment / combat
```

A Rogue mercenary, Valkyrie, Skeleton, Revive, Shadow Master, Druid wolf, trap/sentry, and temporary quest follower can share some ownership infrastructure without pretending they have identical lifecycle rules.

## Current Dark Magic foundation

The typed game-data catalog already includes:

- `Hireling` records;
- `HirelingDescription` records;
- `PetType` records;
- monster, skill, item, body-location, state, difficulty and presentation data.

The engine also already has:

- session-owned ECS entity identity;
- player ownership identity;
- authoritative item/equipment locations including a distinct hireling container;
- deterministic movement/targeting foundations;
- pending monster/AI/combat work described by adjacent research docs.

The hireling system should extend those owners rather than invent a second inventory, second combat calculator, or separate renderer-driven actor model.

## PetType is an ownership/lifetime category

D2MOO's `PlayerPets.cpp` allocates a pet list indexed by `PetType.txt` records and initializes per-type maximum counts from the table's base maximum. This is strong evidence that pet type is not merely a presentation category.

The original player-owned-unit bookkeeping tracks, per pet type:

- current count;
- maximum count;
- linked runtime unit identities;
- owner relationship;
- flags and some pet-specific metadata.

Changing a pet type's max can remove excess pets for non-hireable categories.

Dark Magic should normalize this into an explicit owned-unit policy rather than hard-coding limits inside individual skill handlers.

Example:

```text
OwnedUnitCategory
  id
  baseMax
  replacementPolicy
  survivesOwnerDeath
  persistsAcrossZones
  durable
  equipmentPolicy
```

Skill/stat sources may modify the effective max.

## Owner relationships are authoritative

D2MOO AI control stores owner type/GUID, maintains minion lists, and can retrieve the owner from a minion. Pet bookkeeping separately links player to pet GUIDs.

Dark Magic should use stable ECS/entity relationships, for example:

```text
d2legacy.owned_unit
  owner_entity
  category
  durable_id
  slot/index
```

Do not infer ownership from:

- visual color;
- proximity;
- monster class;
- who most recently targeted the unit.

Ownership participates in combat attribution, AI allegiance, skill effects, transitions, XP, quests, and networking.

## Hirelings are durable owned units

Hirelings differ from ordinary summons because they have durable identity and progression.

Independent `.d2s` evidence records mercenary identity fields in the character header, including:

- dead/alive status;
- mercenary ID;
- name ID;
- mercenary type;
- experience;
- separate mercenary item section for expansion saves.

D2MOO save/player-pet code additionally preserves seed/name/hireling identity used to reconstruct the runtime mercenary.

Dark Magic's durable character model should therefore eventually contain a semantic hireling record rather than storing the live ECS entity directly.

Suggested durable shape:

```text
HirelingRecord
  stable durable identity
  hireling definition/type
  name identity
  seed
  level / experience
  dead state
  equipment identities
  version/source metadata
  opaque legacy extension data
```

The `.d2s` codec maps legacy fields to/from this model and preserves unknown bits/bytes separately.

## One active hireling and replacement semantics

The observed D2MOO hireling setup path checks for an existing `PETTYPE_HIREABLE`, detaches/removes the previous runtime pet, then associates the new one.

This is strong evidence for an explicit **single active hireling slot** in the classic expansion behavior.

Hiring/reward/replacement should therefore be an authoritative transaction:

```text
validate unlock/vendor/reward
validate price/eligibility
resolve new hireling definition/seed/name
if old hireling exists:
    apply replacement policy
    preserve/drop/delete equipment according to verified rule
create/update durable hireling record
materialize runtime owned unit if appropriate
```

Do not let vendor UI directly spawn a monster and call it a mercenary.

Exact old-equipment behavior during replacement remains a verification item.

## Hireling level/stat scaling evidence

`MONSTERAI_UpdateMercStatsAndSkills` provides unusually concrete 1.10f evidence.

The observed runtime:

- sets mercenary level;
- finds the appropriate `Hireling.txt` row for ID and level;
- calculates `levelUps = currentLevel - authoredHirelingLevel`;
- calculates next/current XP thresholds through hireling experience helpers;
- raises current XP to at least the threshold appropriate for the level;
- derives strength and dexterity using per-level coefficients divided by 8, with minimums;
- derives max/current HP from base HP and per-level HP in 8-bit fixed-point, with a minimum;
- derives defense linearly from per-level defense;
- derives secondary min/max damage with per-level coefficients divided by 8;
- derives attack rate;
- applies one shared resistance value to fire/lightning/cold/poison, with per-level coefficient divided by 4;
- derives HP regeneration from max HP;
- evaluates up to six authored skills;
- gates a skill by the skill's required level;
- derives skill level from base skill level plus a per-level term shifted by five, clamped to 0..32.

These formulas are strong **1.10f implementation evidence**, but Dark Magic should still test them against owned data/characters before freezing the exact integer boundaries as a cross-version contract.

## Hireling skill set is data-driven

The runtime does not need a separate C/Go class for "Act II merc aura" or "Rogue arrow" just because those are familiar player-facing archetypes.

`Hireling.txt` provides skill IDs, skill modes and level-scaling fields. Skills then execute through the ordinary skill machinery.

Dark Magic should model:

```text
HirelingDefinition
  archetype / act / difficulty/source
  base level/stat coefficients
  XP scaling
  skill slots
      skill ID
      action/mode
      base level
      level scaling
  equipment policy
  presentation identity
```

AI decides when to use those learned/effective skills.

## Hirelings reuse monster AI with ownership context

D2MOO marks merc units with monster-unit machinery and updates summon AI with the player as owner. Hireling behavior belongs in the AI system with an owner-aware behavior profile.

Important behavior to research/implement:

- follow/leash distance;
- teleport-to-owner when too far/stuck;
- combat target acquisition based on owner allegiance;
- retreat/reposition behavior;
- ranged versus melee spacing;
- aura behavior;
- door/path interaction;
- town restrictions;
- owner level/zone transitions;
- owner death;
- mercenary death/revival.

Do not implement follow behavior in the presentation scene.

## Mercenary death differs from ordinary summon death

D2MOO pet handling treats hireables specially when killing/removing player pets. A dead hireling can remain represented in pet metadata and the client is sent a resurrection cost.

This is evidence for a state transition:

```text
active hireling runtime entity
  -> lethal damage
  -> dead hireling durable/runtime state
  -> no ordinary summon replacement/removal
  -> revival service becomes available
  -> authoritative revival transaction
  -> rematerialize/reactivate runtime unit
```

The durable record must survive absence of a live ECS entity.

## Resurrection cost

The original runtime computes hireling resurrection cost from the hireling/monster state and sends it to the player.

Dark Magic should treat the displayed price as a server-owned quote/result, consistent with the current vendor/service command philosophy. Lua may display the quote but should not calculate or trust the price.

Exact formula/caps need a probe.

## Equipment

Dark Magic already has `ContainerHireling` and named body-slot validation. This is the correct place for mercenary equipment authority.

Future work needs to connect equipped items to:

- effective stat sources;
- skill/item procs;
- resistances/damage;
- presentation components;
- durability;
- persistence in the mercenary save section.

Equipment should remain attached to stable item identity. Do not copy item stats into a one-off mercenary struct and lose source provenance.

## Summons and temporary pets

Summons should use the same owned-unit relation but different policies.

Examples of policy questions:

```text
Skeleton
  limit from skill/stat
  oldest/newest replacement?
  survives level transition?
  corpse required to create?

Valkyrie
  single active
  recast replaces old
  generated equipment/stats

Revive
  temporary duration
  source corpse definition
  special AI/stat inheritance

Sentry/trap
  owned combat attribution
  perhaps immobile path/collider
  separate target/skill policy
```

Do not assume every pet is a persistent follower with inventory.

## Limit/replacement transactions

When a summon would exceed its category maximum, replacement must be deterministic.

The observed pet-list code removes excess non-hireable pets when maximum count decreases. Exact ordering/which pet is removed must be verified by category.

Dark Magic should make replacement policy explicit and stable:

```text
Reject
ReplaceOldest
ReplaceNewest
ReplaceSpecificSlot
KillExcessInStableOrder
```

Never depend on map iteration order.

## Owned-unit combat attribution

Every damage/kill source should be able to resolve:

```text
runtime source entity
 -> immediate owner
 -> controlling player/party if applicable
```

This affects:

- XP credit;
- quest kill credit;
- loot ownership/policy;
- PvP hostility;
- on-kill/on-striking effects;
- score/telemetry;
- replay diagnostics.

Preserve both the immediate source and ultimate owner. A trap killing a monster is not identical to the player directly hitting it even if both credit the player.

## Owner transitions between levels/zones

Current Dark Magic now has trusted zone-transition/admission boundaries. Owned units need a parallel policy.

Potential categories:

- follow owner to new level and materialize near legal entry coordinate;
- remain inactive in previous level;
- be destroyed on transition;
- be serialized and restored later;
- be forbidden in town and hidden/inactivated;
- teleport/leash to owner after transition.

Do not carry an old world coordinate into a new zone.

Transition policy belongs to owner/category semantics, not the renderer.

## Multiplayer and disconnects

Future network behavior needs authoritative ownership transfer/lifecycle decisions for:

- player disconnect while hireling/pets exist;
- reconnect to the same session;
- player leaving game permanently;
- owner death;
- party membership changes;
- charm/conversion temporary allegiance;
- host migration, if ever supported.

The server/session must decide whether owned units remain, become inactive, or are removed. Client disappearance alone cannot delete authority.

## Recommended architecture

Possible responsibilities:

```text
internal/game/ownedunit
    relation.go
    category.go
    limits.go
    attribution.go
    transitions.go

internal/game/hireling
    definition.go
    progression.go
    persistence.go
    services.go
```

AI/skill/combat/item packages should consume the relation through narrow interfaces rather than import persistence directly.

## Suggested implementation slices

### H1 — generic owned-unit relation

Add owner/category relationship, deterministic category counts/limits, ultimate-owner attribution, and checkpoint/replay support.

Use a synthetic pet first.

### H2 — one summon vertical slice

Implement one simple skill-created owned monster with a fixed limit and deterministic replacement policy.

This proves skill -> spawn -> AI -> combat attribution -> death/removal.

### H3 — hireling durable record

Add semantic hireling persistence independent of `.d2s`, with legacy importer/exporter mapping planned but not necessarily complete.

### H4 — hireling materialization/progression

Materialize one synthetic/owned-data hireling using typed `Hireling.txt`, verified 1.10f stat formulas, monster AI, skills and hireling equipment container.

### H5 — death/revival/transition

Implement dead durable state, authoritative resurrection quote/transaction, and zone-follow policy.

## Verification backlog

1. Hire/reward unlock rules for Acts I, II, III and V.
2. Differences between Normal/Nightmare/Hell hireling offerings/archetypes.
3. Exact initial hireling level relative to player and vendor context.
4. Name/seed generation and persistence.
5. Exact 1.10f/target-patch stat rounding for all Hireling.txt coefficients.
6. XP gain, level cap, next-level formula, and player-level relationship.
7. Aura selection/skill variants for Act II hirelings by difficulty/version.
8. Act III elemental skill selection and scaling.
9. Equipment body-slot and weapon restrictions by hireling archetype.
10. Equipment handling when replacing a hireling.
11. Resurrection cost formula/caps and gold source.
12. Death persistence across save/exit/rejoin.
13. Town behavior and whether/when hirelings are hidden/inactive.
14. Transition/leash/teleport thresholds and placement search.
15. AI target selection and owner protection behavior.
16. Potion/healing interactions.
17. Owner/player XP and mercenary XP sharing on kills.
18. PetType maximum modification by skills/stats and exact excess-removal order.
19. Per-summon replacement semantics: skeletons, golems, wolves, ravens, shadows, traps.
20. Revive source-monster stat/AI/skill inheritance and duration.
21. Corpse requirement/consumption ordering for summon skills.
22. Minion/trap PvP and quest-credit attribution.
23. Network replication fields for owned units versus derivable client presentation.
24. `.d2s` mercenary item/death fields and byte-exact round-trip behavior.

## Primary sources inspected

- D2MOO pinned 1.10f `source/D2Game/src/PLAYER/PlayerPets.cpp`.
- D2MOO `source/D2Game/src/MONSTER/MonsterAI.cpp`, especially mercenary materialization/stat/skill updates.
- D2MOO player save and NPC/service code reachable from hireling flows.
- Independent `.d2s` layout research used in [CHARACTER_PERSISTENCE_AND_SAVE_FORMAT.md](CHARACTER_PERSISTENCE_AND_SAVE_FORMAT.md).
- Current Dark Magic typed `Hireling`/`PetType` data, `ContainerHireling`, player/session/ECS authority, and transition/navigation systems.
