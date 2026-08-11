# World/NPC/difficulty/endgame verification queue

This queue consolidates empirical probes for world objects, shrines/wells, transitions/portals/waypoints, NPC dialogue/services, difficulty/game modes and special/endgame events.

## P0: architecture-shaping probes

- Map representative `Objects.txt` operate/event IDs to pinned 1.10f behavior families and produce a normalized behavior coverage report.
- Capture one door from click through mode/collision change and exact authoritative timing.
- Capture one chest from object seed/operation through trap/loot state and ground item creation.
- Pin one shrine definition from `Shrines.txt` through selected behavior, duration and reset timing.
- Resolve `LvlWarp`/DS1 source and destination coordinates for one real stair/level exit and compare original arrival placement.
- One-field-diff a waypoint activation in an owned `.d2s` and map semantic waypoint ID to per-difficulty bits.
- Capture one Town Portal creation/use/return/replace lifecycle including access by a second party member.
- Capture one NPC interaction from availability through a quest-dependent dialogue choice and resulting quest bit/event.
- Build a normalized Normal/Nightmare/Hell DifficultyLevels vector from owned tables and trace one consumer from each of combat, monster state duration, gambling and death/progression.
- Capture campaign final-boss completion through durable difficulty-progression update and final portal creation.
- Pin target-version Cow Portal eligibility/consumption and Cow King repeatability/party-credit behavior.

## P1: world objects

- operate/event function ID mapping and unused/custom rows;
- object mode transition timing;
- collision/selectability changes by mode;
- object interaction range/LOS/path stopping;
- door locks/keys/monster-open behavior;
- chest lock/key, TC, item-level, used/reset state;
- trap probability/type and RNG;
- racks/stands item generation;
- barrels/destructibles/explosions;
- caskets/urns/corpses/bookcases special drops;
- secret/slime/gate conditions;
- scheduled object events across inactive rooms;
- simultaneous multiplayer operations;
- object sound/emitter mapping.

## P1: shrines and wells

- shrine spawn/type selection by level/region/rarity/effect class;
- `Code` -> callback mapping;
- Arg0/Arg1 semantics;
- Duration/reset units;
- EffectClass mutual exclusion;
- buff refresh/replacement/stacking;
- health/mana/refill/well resource arithmetic;
- negative-state cleanup;
- pet/hireling well behavior;
- health/mana exchange;
- Gem Shrine inventory selection/upgrade;
- Portal Shrine destination;
- storm/exploding/poison damage;
- Monster Shrine transformation;
- multiplayer shrine availability/reset.

## P1: warps, waypoints and portals

- LvlWarp select/offset/ExitWalk coordinate semantics;
- DS1 logical warp/link pairing;
- stair/trapdoor/teleport-pad behavior;
- deterministic free arrival placement;
- waypoint object -> save bit/index;
- per-difficulty waypoint unlock persistence;
- waypoint ordering/act/quest restrictions;
- Town Portal origin/town/return points;
- portal owner/party/hostility access;
- portal replacement/closure/lifetime;
- owner death/disconnect behavior;
- red/quest portal one-way/return policies;
- Cow/special portal lifecycle;
- same-level Teleport restrictions/no-teleport levels;
- owned-unit transition behavior;
- loading/client-ready protocol for remote sessions.

## P1: NPC/dialogue/services

- NPC/quest speech table linkage;
- localized text/sound IDs;
- gossip ordering/randomness;
- NPC introduction/heard save bits;
- quest offer/reward side-effect timing;
- class-specific dialogue;
- party/shared quest credit;
- service availability by quest/difficulty;
- vendor/gamble refresh interaction lifecycle;
- heal/identify exact scope and costs;
- hire/resurrect listings/prices;
- quest-item hand-in sequencing;
- NPC relocation/disappearance;
- scripted NPC movement/map-AI;
- concurrent multiplayer interaction;
- speech interruption/subtitle timing;
- act travel option gates.

## P1: difficulty and modes

- difficulty unlock quest requirements Classic vs Expansion;
- resistance penalties and non-expansion differences;
- death XP penalty/recovery;
- monster stat/player-count scaling;
- MonsterSkillBonus;
- freeze/cold/AI curse divisors;
- life/mana steal divisors;
- unique/champion/mercenary/boss damage terms;
- Static Field minimum;
- gamble quality/base-grade odds;
- exceptional/elite odds caller mapping;
- per-difficulty quest/waypoint persistence;
- difficulty-specific quest items/vendors/hirelings/monster lists;
- `/players X` semantics;
- Hardcore creation/join/death persistence;
- Classic/Expansion content/formula differences;
- ladder/season feature eligibility.

## P1: campaign/endgame encounters

- Diablo seal encounter activation/order/reset;
- Ancients activation, level requirement, reset and party credit;
- Baal throne wave composition/timing;
- chamber portal/transition and final quest credit;
- first-kill/quest boss drop/reward policy;
- final portal/epilogue transition;
- next-difficulty progression update.

## P1: Cow Level

- recipe inputs/location/progression requirement;
- per-difficulty portal eligibility;
- portal access/return/lifetime;
- Cow Level map/population differences;
- Cow King kill credit and lockout rules;
- party member eligibility interaction;
- Cow King special loot/quest state.

## P2: later-version special content

These require a pinned later-patch corpus rather than 1.10f assumptions:

- Uber Dungeon/Uber Tristram Cube operations and introduction version;
- key/organ drops and recipes;
- special portal destination selection/nonduplication;
- Uber boss stats/AI/rewards;
- Hellfire Torch/unique charm restrictions;
- realm/global world-event trigger/counter protocol;
- special boss replacement rules;
- realm announcements and reconnect persistence;
- ladder/season restrictions.

## Probe artifact standard

Each probe should record:

```text
probe ID/title
target game/client/server/content version
owned assets/save/game setup
reproduction steps
raw observed values/events/bits
normalized vector/state transition
RNG/tick state where relevant
patch differences
confidence upgrade
primary doc section updated
```

Keep proprietary captures outside Git. Commit safe normalized fixtures, hashes, screenshots metadata, test vectors and inspector output.
