# Coverage map

Every original Diablo II (Lord of Destruction 1.13c) behavior found by the discovery sweep,
grouped by category prefix, with its spec or its status as a candidate. A candidate row is a
proposed spec, not a commitment: the maintainer picks which rows to write. Rows are never deleted
or renamed; a row that merges into another spec links it, a row dropped is marked `out of scope`
with a reason.

Status values: `written` (linked), `candidate`, `out of scope`.

All entry points are **leads from a discovery sweep, not verified rules**. Names after an RVA are
analyst labels, which may be wrong; the original module and RVA are the reference. Data-table
columns are discovery leads unless linked rule/evidence records establish otherwise. A written
spec can still contain gated rules and unobserved behavior; presence does not establish approval
or complete verification.

## Sweep log

| Date | Original artifacts / identity | Method and coverage | Limits |
|---|---|---|---|
| 2026-10-12 | D2Game.dll, D2Common.dll, D2Client.dll, target LoD 1.13c; historical analysis specimens | Assert-string source identities and referencing functions; instruction/data-table enumeration of missile, item-event, timer, packet, quest, object, level, path, skill, AI and item-stat dispatch regions | Historical coverage matched citations to entry points only; it did not establish complete rule/branch coverage. Reconstruction details are preserved privately. Current artifact fingerprints do not retroactively establish each historical specimen's identity. |
| 2026-10-14 | [Source manifest](../research/source-manifest.md): LOD113C-D2CLIENT and LOD113C-D2SOUND | Read-only original-image scan for audio/lighting assertion strings and named export leads; D2Client dLightMap, SoundHdr, UnitSnd, PalShift and View identities; D2Sound utility/playback assertion identities | Discovery leads only; no selection, transition, playback, lighting rule or dispatch chain verified. Music selection was absent from this map and is added with unresolved location. This is a targeted inventory update, not a new complete engine sweep. |

Known limits of this sweep: the corpus `asm` covers about 80% of D2Client functions, so
assert-string counts are a lower bound; States.txt set/remove callback tables, Objects.txt
InitFn/PopulateFn dispatch, client missile functions (CltDoFunc/CltHitFunc), and D2Sound/D2Win
were not enumerated.

## core

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md) | D2Game, D2Common, Fog | named inputs in spec | see spec | written | RNG, integers, fixed point and time units. |
| [core-record-layouts-and-ids](core-record-layouts-and-ids.md) | D2Common, D2Game | named inputs in spec | see spec | written | Split from core-numeric-and-rng-foundations (2026-10-15): records, stat reads, states and modes. |
| [core-game-data-fields](core-game-data-fields.md) | D2Common, D2Game, Fog, Storm | all cited | see spec | written | Identities and names. |
| [core-game-data-table-records](core-game-data-table-records.md) | D2Common, Fog, Storm | all cited | see spec | written | Split from core-game-data-fields (2026-10-15): table layouts, getters, conversions and loading. |
| core-data-table-loading | D2Common `DATATBLS\DataTbls`, `MissileTbls`, `ItemTbls`, `MonsterTbls`, `LvlTbls`, `FieldTbls`, `TokTbls` | all tables | D2Common `0x7C8D0` (Missiles loader, label `LoadMissilesTxtTable`) | candidate | Binary table compilation, linking of string keys to row ids, expansion vs classic rows. |

## combat

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| [combat-damage-pipeline-order](combat-damage-pipeline-order.md) | D2Game `UNIT\SUnitDmg` | many | `0xDE0C0`, `0xDD450`, `0xDCF20` | written | |
| [combat-damage-mitigation](combat-damage-mitigation.md) | `UNIT\SUnitDmg` | resist/absorb stats, DifficultyLevels | `0xDB3C0`, `0xDBD00`, `0xDC0B0` | written | |
| [combat-hit-roll](combat-hit-roll.md) | `UNIT\SUnitDmg` | to-hit and defense stats | see spec | written | Split from combat-shield-block-and-avoidance (2026-10-15): the attack hit roll. |
| [combat-shield-block-and-avoidance](combat-shield-block-and-avoidance.md) | `UNIT\SUnitDmg` | block/avoid stats | `0xDB790`, `0x3EDE0` | written | Block and avoidance; the hit roll is combat-hit-roll. |
| [combat-leech-and-drain](combat-leech-and-drain.md) | `UNIT\SUnitDmg` | leech and drain fields, DifficultyLevels leech divisors | `0xDBA40`, `0xDB570`, `0xDA970` | written | Leech and drain only; status application is combat-status-application. Renamed 2026-10-15. |
| [combat-status-application](combat-status-application.md) | `UNIT\SUnitDmg` (assert file string empty; ownership follows the caller) | stun, cold, freeze, poison, burn fields; MonStats ColdEffect/Velocity, DifficultyLevels cold/freeze divisors | `0xDE0C0` (hand-off), `0xDCC20`, `0xDC780`, `0xDDAC0`, `0xDCAB0`, `0xDC940`, `0xDB4C0` | written | |
| [combat-crushing-blow](combat-crushing-blow.md) | `UNIT\SUnitEvent` | ItemStatCost | `0xACC30`, `0x37030` | written | |
| [combat-open-wounds](combat-open-wounds.md) | `UNIT\SUnitEvent` | ItemStatCost | `0xAD410`, `0x8E6F0` | written | |
| [combat-knockback](combat-knockback.md) | `UNIT\SUnitEvent`, `SUnitDmg` | MonStats2 | `0xACE60`, `0x75160` | written | |
| [combat-hit-recovery](combat-hit-recovery.md) | `UNIT\SUnitDmg` | MonStats2 | `0xDCF20`, `0x75210` | written | |
| [combat-pvp-damage-and-kill-credit](combat-pvp-damage-and-kill-credit.md) | `UNIT\SUnitDmg` | hostility, party | `0xDEEE0`, `0xDEA00`, `0x6F930` | written | Hostility declaration is player-hostility-declaration. |
| [combat-item-event-triggers-and-procs](combat-item-event-triggers-and-procs.md) | `UNIT\SUnitEvent` | ItemStatCost itemevent1/2, itemeventfunc1/2 | dispatcher `0x37030`; callback table VA `0x6FD277A8` slots 1-31 (e.g. `0xAD910`, `0xAF2E0`, `0xAE0E0`); trigger callers `0xDE0C0` (hit), `0x76740` (kill), `0xDE6D0` (level-up) | written | Dispatcher and trigger points; IEP-R14–R18 specify retaliation, damage-to-mana, after-kill resource, and Rest in Peace callbacks. Knockback, Open Wounds and Crushing Blow have separate specs. Unread selector effects and complete original item bindings remain gated by IEP-Q06/IEP-Q08; chance-to-cast, slow/freeze target and reanimation spawning retain separate ownership. |
| combat-area-damage-sweep | `UNIT\SUnitDmg` | Missiles/Skills radius columns | `0xDE380` (callers `0x76270`, `0x265E0`, `0x44300`, `0x457D0`) | candidate | Target filtering and order for area hits. |

## state

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| [state-movement-modifiers-and-freeze](state-movement-modifiers-and-freeze.md) | D2Common `Stats` | speed stats | `0x323E0`, `0xDC780` | written | |
| [state-timed-state-lifecycle](state-timed-state-lifecycle.md) | D2Common `Stats\StatsEx`; D2Game timer events | States (setfunc, remfunc, flags); statlist expire frame | timer event-type table VA `0x6FD1BDB8`, type 6 `0x6E470`; D2Common `0x32E10`, `0x3A9F0`, `0x38550` | written | Apply, refresh, expire, remove callback. Set/remove callback table not located. |
| state-curses | D2Game `GAME\Targets`; D2Common States | States curse flag; Skills curse columns | `0x50B70`, caller `0x51A90`; `0xEFDB0` | candidate | Application, override rule, duration, caster-side tracking list. |
| state-aura-pulse-and-propagation | `SKILLS` (aura), state side | Skills aurafilter, aurarangecalc, aurastate, auraevent | `0x42570` (skill table VA `0x6FD2750C`); `0x9ABC0` | candidate | Shared pulse, range, refresh. Per-aura effects go to `skill-paladin-`. |
| state-life-mana-stamina-regeneration | D2Game regen tick | regen stats, stamina | `0x6CEE0` (table VA `0x6FD27BBC` slot 21); `0x8DA80` | out of scope | Moved to player-life-mana-stamina-regeneration (regeneration is `player` per template.md, maintainer decision 2026-10-12). |

## missile

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| [missile-creation-and-seeding](missile-creation-and-seeding.md) | D2Game `MISSILES\Missiles`; D2Common `UNITS\Missile` | Missiles Vel, MaxVel, Accel, Skill | `0x6F930` (84 callers); D2Common `0x6A390`, `0x6A350` | written | Creation request, allocation and seeding (renamed and split 2026-10-15). |
| [missile-initialization-and-registry](missile-initialization-and-registry.md) | D2Game `MISSILES\Missiles`; D2Common `UNITS\Missile` | Missiles Range, LevRange, HitShift, Light | see spec | written | Split from missile-creation-and-seeding (2026-10-15): initialization, attribution and registry. |
| [missile-movement-and-collision](missile-movement-and-collision.md) | `MISSILES`; D2Common `UNITS\Missile`, `PATH` | Missiles CollideType, Collision, CollideKill, CollideFriend, LastCollide, Pierce, NextHit, NextDelay | `0x3F320` → `0x3EDE0`; `0xAE9A0` → `0x3E3F0` | written | |
| [missile-hit-and-damage-calculation](missile-hit-and-damage-calculation.md) | `MISSILES`; D2Common `UNITS\Missile` | Missiles SrvDmgFunc, MinDamage..EMax, LevDam*, DmgCalc1, ElemType, SrcDamage, SrcMissDmg | `0x3E510`, `0x3EDE0`; dmg table VA `0x6FD2DB68` (15 slots); D2Common `0x6A3E0`, `0x6B2B0` | written | Hit path partly cited by pvp, shield, knockback specs; damage arithmetic not written. |
| [missile-damage-parameters](missile-damage-parameters.md) | D2Game `MISSILES`; D2Common `UNITS\Missile` | Missiles MinDamage..HitShift, SrcDamage, SrcMissDmg | `0x6F930` → `0x6F620`; D2Common `0x6BAA0`, `0x6B4A0`, `0x6A580`, `0x6A4B0`, `0x6B360`, `0x6A3E0` | written | Split from missile-hit-and-damage-calculation; owns MDP-R01..MDP-R12, which missile-hit-and-damage-calculation consumes. Columns are read at creation, not at hit. |
| missile-server-do-functions | `MISSILES` | Missiles SrvDoFunc, Param1-5, SrvCalc1, SubMissile1-3 | dispatch `0x3E3F0`; table VA `0x6FD2D970` (37 slots) | candidate | Split by family: periodic spawners, chains, homing, teleport. |
| missile-server-hit-functions | `MISSILES` | Missiles SrvHitFunc, Explosion, sHitPar1-3, HitSubMissile1-4 | table VA `0x6FD2DA48` (59 slots) | candidate | Split by family: explosion, sub-missiles, chain/bounce, monster spawn. |
| missile-stream | D2Common `UNITS\MisStream` | Missiles path length | D2Common `0x1D760`, `0x1D7C0`, `0x1D710` | candidate | Small; may fold into movement. |
| missile-client-functions | D2Client `UNIT\Missile` | Missiles CltDoFunc, CltHitFunc | D2Client `0xB30F0`, `0xB1FC0` | candidate | Client prediction and visuals; tables not enumerated. |

## skill

Skill function tables: D2Game srvstfunc RVA `0x107338`, srvdofunc RVA `0x1074A8`. Skill names in
this section come from a reference copy of the Skills table and are leads.

Class files without an assert string (amazon, assassin) were assigned by address range. Skill
index to skill name is not decoded; family names are leads to confirm against the Skills
srvdofunc values. Event callback table VA `0x6FD277A8` (31 entries, dispatcher `0x9E290`) is the
same table as in combat-item-event-triggers-and-procs; the two rows may merge. Client tables:
D2Client cltst VA `0x6FB8E928` (53), cltdo VA `0x6FB8EA48` (96).

### Shared machinery

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| [skill-cast-start-and-do-dispatch](skill-cast-start-and-do-dispatch.md) | D2Game `SKILLS\Skills` | Skills srvstfunc, srvdofunc | `0xA19A0`, `0x9FD00`, `0x9E420` | written | |
| skill-charge-up-progression | `SKILLS\Skills` | srvprgfunc1-3, prgcalc1-3, prgdam (lead) | `0x91B70`, `0x9E8E0` | candidate | May merge with the assassin martial-arts row. |
| skill-mana-cost-and-deduction | `SKILLS\Skills` (D2Common, D2Game) | mana, lvlmana, minmana, manashift | D2Common `0x16F0`, `0x4FDC0`; D2Game `0x9F850` | candidate | |
| skill-delay-and-cooldown | D2Game `SKILLS\Skills` | Skills delay; skill-delay state | `0x9F1A0` | candidate | Hireling cooldown `0xA4A10` is in the sentry spec. |
| skill-levels-bonuses-and-charged-skills | D2Common `SKILLS\Skills` | all-skills, class and tab bonuses, charged skills | D2Common `0x4FCB0`, `0x52CA0`, `0x4EE00` | candidate | |
| [skill-calc-expressions](skill-calc-expressions.md) | D2Common `SKILLS\Skills`, `DATATBLS` | calc columns | D2Common `0x37AF0`, `0x3AC50`, `0x51B80` | written | Renamed 2026-10-15; synergy damage builders are owned by skill-leveled-damage-tohit-and-elements. |
| [skill-leveled-damage-tohit-and-elements](skill-leveled-damage-tohit-and-elements.md) | D2Common `SKILLS\Skills` | MinDam, MaxDam, MinLevDam1-5, EMin, EMax, EMinLev*, ToHit, LevToHit, HitShift | D2Common `0x1670`, `0x16B0`, `0x1730`, `0x50360`, `0x50460` | written | |
| skill-requirements-and-point-allocation | `SKILLS\Skills` (D2Common, D2Game) | reqlevel, reqskill1-3 | D2Common `0x50610`, `0x52210`; D2Game `0x9FA00` | candidate | |
| skill-selection-left-right | D2Game `GAME\CCmd`, `SKILLS\Skills`; D2Client `UI` | left/right skill, hotkeys | `0xA0830`; D2Client `0x3CFA0`, `0x3D080` | candidate | Server select handler not located. |
| skill-event-callbacks | `SKILLS\Skills`, `SkillItem` | aura/target event function columns (lead), ItemStatCost itemevent | `0x9E290`, `0x50250`; table VA `0x6FD277A8` | candidate | See note above. |
| skill-basic-attacks-and-throw | `SKILLS\Skills` | srvdo 1-5, srvst 1-3 and 65 | `0xA2C40`, `0xA2A40`, `0xA1630` | candidate | `0xA2C40`, `0xA1630` cited by written specs, not as skill rules. |
| skill-client-cast-dispatch | D2Client `SKILLS\Skills` | cltstfunc, cltdofunc | D2Client `0xA1CF0`, `0xA1B10`, `0x72610`, `0x728B0` | candidate | Client prediction and visuals; in scope (maintainer decision 2026-10-12). |

### Amazon (D2Game RVA `0x4C000`-`0x4E000`, no assert string)

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| skill-amazon-bow-and-javelin-volleys | inferred `SKILLS\SkillAma` | srvdo 8, 10-12; srvst 4, 8 | `0x4D420`, `0x4D270`, `0x4C840` | candidate | |
| skill-amazon-spear-and-lightning-skills | inferred `SKILLS\SkillAma` | srvdo 6, 7, 13, 14; srvst 5-7 | `0x4C2A0`, `0x4DAC0` | candidate | |
| skill-amazon-inner-sight-and-valkyrie | inferred `SKILLS\SkillAma` | srvdo 15, 16; srvst 9-10 | `0x4D8C0`, `0x4D710` | candidate | |

### Sorceress (`SKILLS\SkillSor`, RVA `0x41000`-`0x44400`)

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| skill-sorceress-armors-and-energy-shield | `SKILLS\SkillSor` | srvdo 25; event table 1-3, 24 | `0x42570`, `0x42C70`, `0x41FB0` | candidate | |
| skill-sorceress-teleport | `SKILLS\SkillSor` | srvdo 27; srvst 14 | `0x43450`, `0x42AD0` | candidate | |
| skill-sorceress-novas-static-and-area-casts | `SKILLS\SkillSor` | srvdo 20, 21, 29 | `0x430D0`, `0x431C0`, `0x43FA0` | candidate | |
| skill-sorceress-placed-and-falling-missiles | `SKILLS\SkillSor` | srvdo 17-19, 22, 24, 26, 28; srvst 11-13 | `0x43B60`, `0x428F0`, `0x44390` | candidate | |
| skill-sorceress-enchant-and-thunderstorm | `SKILLS\SkillSor` | srvdo 23 | `0x42730` | candidate | |

### Necromancer (`SKILLS\SkillNec`, RVA `0x4E000`-`0x52000`)

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| skill-necromancer-curses | `SKILLS\SkillNec` | event table 4, 5, 22, 23, 26, 27 | `0x50B70`, `0x4E760`, `0x4EF70` | candidate | Casting only; shared curse rules in state-curses. |
| skill-necromancer-corpse-explosion | `SKILLS\SkillNec` | srvdo 55 | `0x4F4D0`, `0x4F370` | candidate | Shared damage code written in skill-assassin-death-sentry-corpse-damage; player-cast corpse selection is not. |
| skill-necromancer-raise-skeletons-and-golems | `SKILLS\SkillNec` | srvdo 30, 31, 58; srvst 15-21 | `0x505D0`, `0x518B0`, `0x51340` | candidate | |
| skill-necromancer-bone-and-poison-skills | `SKILLS\SkillNec` | srvdo 56, 57, 59-63 | `0x51750`, `0x51540`, `0x50190` | candidate | |

### Paladin (`SKILLS\SkilPal`, RVA `0x98000`-`0x9B000`)

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| skill-paladin-auras | `SKILLS\SkilPal` | srvdo 81; aurastate, aurastat columns | `0x9ABC0`, `0x98390` | candidate | Largest paladin row; may split. |
| skill-paladin-combat-skills-and-charge | `SKILLS\SkilPal` | srvdo 64, 66, 67, 73, 79, 80, 150; srvst 29, 35 | `0x98E90`, `0x98BA0`, `0x9AF50` | candidate | |
| skill-paladin-hammer-bolt-and-fist | `SKILLS\SkilPal` | srvdo 65, 82; srvst 31 | `0x9A400`, `0x9A8D0`, `0x99620` | candidate | `0x99620` cited by state-movement-modifiers-and-freeze. |

### Barbarian (`SKILLS\SkillBar`, RVA `0x26000`-`0x2A000`)

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| skill-barbarian-warcries | `SKILLS\SkillBar` | srvdo 76-78 | `0x293B0`, `0x28BE0`, `0x294A0` | candidate | |
| skill-barbarian-leap-and-whirlwind | `SKILLS\SkillBar` | srvdo 69, 70, 72, 74; srvst 38-41 | `0x27600`, `0x276D0`, `0x28290` | candidate | |
| skill-barbarian-melee-combat-skills | `SKILLS\SkillBar` | srvdo 9, 71, 75; srvst 32-34 | `0x29600`, `0x29000`, `0x27C90` | candidate | |
| skill-barbarian-find-item-and-grim-ward | `SKILLS\SkillBar` | srvdo 70 (lead) | `0x27B90`, `0x26A80` | candidate | May fold into warcries. |

### Druid (`SKILLS\SkillDruid`, RVA `0x44000`-`0x47000`)

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| skill-druid-summons | `SKILLS\SkillDruid` | srvdo 114, 115, 119 | `0x468A0`, `0x46720`, `0x46560` | candidate | |
| skill-druid-shapeshifting | `SKILLS\SkillDruid` | srvdo 124; srvst 56-58 | `0x45CE0`, `0x44D70`, `0x44C20` | candidate | |
| skill-druid-elemental-skills | `SKILLS\SkillDruid` | srvdo 116-118, 122, 123, 145, 146; event table 25 | `0x46210`, `0x456B0`, `0x45AF0` | candidate | |

### Assassin (D2Game RVA `0x91B00`-`0x96400`, no assert string)

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| [skill-assassin-sentry-firing-and-range](skill-assassin-sentry-firing-and-range.md) | inferred `SKILLS\SkillAss`, AI | AI think | `0xA7DC0`, `0xA4A10`, `0xB1BC0` | written | Think, range and firing decision. |
| [skill-assassin-sentry-missile-launch](skill-assassin-sentry-missile-launch.md) | inferred `SKILLS\SkillAss` | Skills calc1, Missiles Vel/VelLev/Size | see spec | written | Split from skill-assassin-sentry-firing-and-range (2026-10-15): missile launch. |
| [skill-assassin-blade-sentinel-and-fury](skill-assassin-blade-sentinel-and-fury.md) | inferred `SKILLS\SkillAss` | srvdo 44, 48 | `0x96310`, `0x95560` | written | |
| [skill-assassin-death-sentry-corpse-selection](skill-assassin-death-sentry-corpse-selection.md) | AI | Skills Param3/Param4, MonStats | `0xA7C20` | written | Split from skill-assassin-death-sentry-corpse-damage (2026-10-15): think and corpse selection. |
| [skill-assassin-death-sentry-corpse-damage](skill-assassin-death-sentry-corpse-damage.md) | `SKILLS\SkillNec` do 55 | srvdo 55 | `0x4F4D0` | written | Corpse explosion. |
| skill-assassin-martial-arts-charges-and-finishers | inferred `SKILLS\SkillAss` | srvdo 38, 42, 50, 52; srvprgfunc | `0x93340`, `0x95FE0`, `0x91B70` | candidate | |
| skill-assassin-trap-placement | inferred `SKILLS\SkillAss` | srvdo 39-41, 43, 47 | `0x94970`, `0x92230`, `0x94B10` | candidate | |
| skill-assassin-shadows-and-buffs | inferred `SKILLS\SkillAss` | srvdo 33, 45, 49; srvst 22-28 | `0x96280`, `0x95980`, `0x92770` | candidate | |
| skill-assassin-mind-and-movement-skills | inferred `SKILLS\SkillAss` | srvdo 34, 36, 37, 46, 51, 53, 54 | `0x95690`, `0x95040`, `0x94F50` | candidate | |

### Item-granted skills (`SKILLS\SkillItem`)

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| skill-item-granted-skills-and-charges | D2Game `SKILLS\SkillItem` | ItemStatCost single-skill and charged-skill stats, Skills | `0xADAC0`, `0xACF10`, `0xAFCE0`, `0xD2380`, `0xDA350`; srvdo 113; event table 6-21, 28-31 | candidate | |

### Monster skills (`SKILLS\SkillMonst`, RVA `0xB2E00`-`0xB6C80`, and an unnamed file at RVA `0xC000`-`0xE000`)

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| skill-monster-maggots-and-burrowing | D2Game `SKILLS\SkillMonst` | Skills srvst/srvdo, aux params | st 43 `0xB31B0`, st 51 `0xB2FA0`, do 87 `0xB42D0` | candidate | Egg laying, submerge/emerge. |
| skill-monster-spawning | `SKILLS\SkillMonst` and an unnamed file | Skills, MonStats | do 91 `0xB5E50`; do 135 `0xCDC0`; do 133 `0xC5D0` | candidate | Also cry for help (do 128). |
| skill-monster-resurrect-and-heal | `SKILLS\SkillMonst` | Skills | do 97 `0xB5BD0`; do 96 `0xB3770`; st 61 `0xC1B0` | candidate | Also blood suck. |
| skill-monster-melee-specials | `SKILLS\SkillMonst` | Skills | do 83 `0xB4660`; do 92 `0xB5600`; do 109 `0xB6790` | candidate | Also leap, swarm move, eat, whip/stomp (RVA `0xC000`-`0xE000`). |
| skill-monster-missiles-and-novas | `SKILLS\SkillMonst` | Skills, Missiles | do 88 `0xB56C0`; do 99 `0xB52A0`; do 95 `0xB6A60` | candidate | Also turrets 105/106, do 110, 148/149. |
| skill-monster-teleport | `SKILLS\SkillMonst` | Skills | do 98 `0xB46B0`; do 129 `0xDBD0` | candidate | |
| skill-monster-diablo | `SKILLS\SkillMonst` | Skills | do 103 `0xB3F10`; do 100 `0xB3D60`; do 104 `0xB4410`; do 152 `0xB6840` | candidate | |
| skill-monster-curse-and-aura | `SKILLS\SkillMonst` | Skills | do 112 `0xB5950`; do 111 `0xB39F0` | candidate | |
| skill-monster-baal | unnamed file | Skills | do 140 `0xCB30`; do 141 `0xCA20`; do 139 `0xD290` | candidate | |

## monster

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| [monster-scaling-by-players-and-difficulty](monster-scaling-by-players-and-difficulty.md) | D2Game `MONSTER\Monster` | MonStats, DifficultyLevels | `0xAFDB0`, `0xAF840`, `0x12380` | written | Life, XP, NoDrop player count. |
| monster-region-population | `MONSTER\MonsterRegion` | Levels mon*/nmon*/umon*, NumMon, MonDen, MonUMin/Max; MonStats Rarity | `0xCA3B0`, `0xCAB30`, `0xE1430`, `0x47A20`, `0x2EEA0` | candidate | |
| monster-creation-and-placement | `MONSTER\Monster` | MonStats, MonStats2, MonPlace, MonPreset | `0x32DF0`, `0x36820`, `0xAFAB0` | candidate | Init seed draws partly in MSD-R01 to MSD-R19. |
| monster-level-by-area-and-rank | D2Common `Monsters`, `MonsterTbls` | Levels MonLvl*, MonLvl table | D2Common `0x55580`, `0x1B1D0`; D2Game `0x98410` | candidate | |
| monster-champion-and-unique-selection | `MONSTER\Monster` | Levels MonUMin/MonUMax; MonStats flags | `0x91610`, `0x91660`, `0x9F120` | candidate | |
| monster-unique-mods | `MONSTER`; D2Common `MonsterTbls` | MonUMod | umod table `.data` RVA `0x10E550`; users `0x246D0`, `0x248A0`, `0x91660`; `0x21830`; D2Common `0x59CB0` | candidate | Second table `0x10E5E0` unresolved. May split into assignment and effects. |
| monster-minions-and-packs | `MONSTER`, `Ai\AiGeneral` | MonStats MinGrp/MaxGrp, PartyMin/Max, minion1/2; SuperUniques | `0x23CA0`, `0x21490`, `0x48410` | candidate | |
| monster-superuniques | `MONSTER`; D2Common `MonsterTbls` | SuperUniques | D2Common `0x59870`, `0x53C80`; D2Game `0xBAFF0` | candidate | |
| monster-modes-and-mode-table | `MONSTER\MonsterMode` | MonMode, MonSeq, MonStats2 mode flags | mode table `.rdata` RVA about `0xFA490`-`0xFA5F0`; `0x748C0`; D2Common `0x55250`, `0x19330` | candidate | Entry layout unresolved. |
| monster-death-and-corpse | `MONSTER\MonsterMode` | MonStats2 corpse columns | `0x76270`, `0x75460`, `0x75900`; D2Common `0x2EF30` | candidate | |
| monster-corpse-revivability-and-resurrection | `MONSTER` | MonStats revive flags | D2Common `0x50020`, `0x500E0`; D2Game `0x48FC0`, `0xAE200` | candidate | Skill side in skill-monster-resurrect-and-heal. |
| monster-equipment-and-components | D2Common `Monsters` | MonEquip, MonStats2 components | D2Common `0x54EF0`, `0x1930`; D2Game `0xAF730` | candidate | Appearance draws in MSD-R03. |
| monster-properties-monprop | D2Common `MonsterTbls` | MonProp | D2Common `0x55A90` | candidate | D2Game consumer not found. |
| monster-client-unit-init | D2Client `UNIT\Monster`, `UNIT\MonUnique` | MonStats2 | D2Client `0x4D610`, `0x4EF50`, `0x1FC40` | candidate | |
| (Diablo clone) | none | none | `0xE8780` | out of scope | Realm-triggered. |

## ai

AI table: D2Game `.data` RVA `0x10E788`, 148 entries {flag, init, think, aux}, indexed by MonStats
AI (lookup `0x9D0F0` clamps to 0..0x93). AI names come from a reference copy of MonAi and are
leads.

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| ai-dispatch-and-think-scheduling | D2Game `MONSTER\MonsterAI`, `Ai\AiGeneral` | MonStats AI, aip1-8 | `0x9D0F0`; special table RVA `0x10E668` (18 entries); `0x487A0`; `0x20B30` | candidate | Special-table selector unresolved. |
| ai-target-selection-helpers | `Ai\AiGeneral` | none | `0x20CF0`, `0x246D0`, `0xB1BC0` | candidate | May fold into dispatch. |
| ai-melee-approach-family | `Ai\AiThink*` | aip* | `0x83AC0`, `0x838B0`, `0x83CE0` | candidate | Skeleton, zombie, brute and about 13 similar. |
| ai-pack-and-flee | `Ai\AiThink*` | aip* | `0x83130`, `0x826B0`, `0x800C0` | candidate | Fallen, quill rat, fetish. |
| ai-ranged-and-caster | `Ai\AiThink*` | aip* | `0x874A0`, `0x864A0`, `0x851E0`, `0x1AF80`, `0x1A9F0`, `0x179A0` | candidate | May split. |
| ai-shamans-and-healers | `Ai\AiThink*` | aip* | `0x890A0`, `0x87BB0`, `0x87E60`, `0x88380`, `0x19080` | candidate | |
| ai-spawners-and-nests | `Ai\AiThink*` | aip*; MonStats spawn columns | `0x87010`, `0x7CFD0`, `0x15E00`, `0x6E4E0` | candidate | About 13 rows. |
| ai-flying-and-swarm | `Ai\AiThink*` | aip* | `0x83550`, `0x89C30`, `0x80450` | candidate | |
| ai-burrow-leap-and-special-movement | `Ai\AiThink*` | aip* | `0x88D80`, `0x88810`, `0x7CB10` | candidate | |
| ai-traps-and-turrets | `Ai\AiThink*` | aip* | `0x7C580`, `0x16270`, `0xA57E0`, `0x16410`, `0x7C910` | candidate | |
| ai-act-bosses-andariel-duriel-mephisto | `Ai\AiThink*` | aip* | `0x7FE60`, `0x89430`, `0x85B60`; also `0x7EA10`, `0x85780`, `0x19E70` | candidate | |
| ai-diablo | `Ai\AiThink4` | aip* | `0xA9610` (aux `0xA59B0`) | candidate | |
| ai-act5-monsters | `Ai\AiThink4`, `AiThink*` | aip* | `0x18730`, `0x17CB0`, `0x1A550`, `0xA6530`, `0xA7290` | candidate | |
| ai-nihlathak-and-ancients | `Ai\AiThink4` | aip* | `0xA7580`, `0xA7530`, `0xA4F40` | candidate | |
| ai-baal | `Ai\AiBaal` | aip* | `0xB8610` (aux `0xB78A0`), `0xB8460`; `0xA8430`, `0xA9D40`, `0xA4D10`, `0xA5D70` | candidate | |
| ai-uber-bosses | `Ai\AiThink*`, `AiBaal` | aip* | `0xB6C80`, `0x7B4E0`, `0xA39D0` | candidate | In scope (maintainer decision 2026-10-12). |
| ai-town-npcs | `Ai\AiGeneral` | aip* | `0x1DB40`, `0x1BC60`, `0xA58D0`, `0x9CB00` | candidate | About 19 rows. |
| ai-hireling | `Ai\AiGeneral` | Hireling | `0x1DD40` (aux `0x16060`) | candidate | |
| ai-necromancer-and-druid-summons | `Ai\AiThink4`, `AiGeneral` | aip* | `0x1E900`, `0xAC8E0`, `0xAB360` | candidate | |
| ai-shadow-warrior-and-master | `Ai\AiThink4` | aip* | `0xAC410`, `0xAB580`, `0xA6B70`, `0xA5350` | candidate | |
| (sentry and death sentry AI) | `Ai\AiThink4` | none | `0xA7DC0`, `0xA7C20` | written | In [skill-assassin-sentry-firing-and-range](skill-assassin-sentry-firing-and-range.md) and [skill-assassin-death-sentry-corpse-selection](skill-assassin-death-sentry-corpse-selection.md). |

## game

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| game-creation-and-difficulty-setup | D2Game `GAME\Game` | DifficultyLevels | `0x2C2D0`, `0xEE90`, `0xAF840` | candidate | Seeds the game RNG and initializes quests (`0x348A0`). |
| game-frame-loop-and-tick-rate | `GAME\Game` | none | `0x2E050`, `0x29800`, `0x2DF10` | candidate | |
| game-event-queue-and-timers | `GAME\Event`, `Task` | none | `0x8E630`, `0x8EA80`, `0x529F0`, `0x522D0`, `0x8E6F0` | candidate | |
| game-client-join-leave-and-act-change | `GAME\Clients` | Levels (Act) | `0x2A740`, `0x2BB10`, `0x90C60`, `0x2DEC0` | candidate | |
| game-room-activation-and-inactive-units | D2Game `UNIT\SUnitInactive`; D2Common `DUNGEON`, `DRLG\DrlgRoom` | none | D2Common `0x3B9F0`, `0x44DF0`, `0x2DF80` | candidate | Which rooms and units are simulated; SUnitInactive has no labeled entry point yet. |
| game-arena-scoring | `GAME\Arena` | none | `0xB24D0`, `0xB2820`, `0xB2940` | out of scope | Excluded by maintainer decision 2026-10-12. |
| game-client-recording | D2Client `GAME\Record` | none | D2Client `0xAB8D0` | candidate | In scope (maintainer decision 2026-10-12). |
| (server admin and game-server status packets 0xFA-0xFD) | `GAME` | none | `0x29CB0` | out of scope | Realm and game-server management. |

## net

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| net-client-command-table-and-validation | D2Game `GAME\CCmd` | none | CCmd table `.rdata` RVA `0xFA080`, {handler, flag}, ids 0x01-0x66 | candidate | Null ids: 0x2B, 0x4A, 0x4E, 0x55-0x57, 0x5A-0x5C, 0x64, 0x65. |
| net-client-movement-commands | CCmd 0x01-0x04 | none | `0xCE0D0`, `0xCE090`, `0xCE000` | candidate | |
| net-client-skill-commands | CCmd 0x05-0x11 | Skills | `0xCE5A0`, `0xCE9D0`, `0xCEA50`; validator `0xCD630` | candidate | |
| net-client-item-commands | CCmd 0x16-0x2A, 0x50 | Items | `0xD08B0`, `0xCEE60`, `0xCF270`, `0xCEF40` | candidate | Pickup, drop, cursor, body, belt, socket, cube, gold drop. |
| net-client-interact-npc-trade-commands | CCmd 0x13, 0x2F-0x38, 0x4F | none | `0xD0910`, `0xCF220`, `0xD1470` | candidate | |
| net-client-stat-skill-point-commands | CCmd 0x3A-0x3C | none | `0xCC960`, `0xCDA20`, `0xCF480` | candidate | |
| net-client-party-merc-misc-commands | CCmd 0x14/0x15, 0x46/0x47, 0x5D-0x63 | none | `0xCC8E0`, `0xCD3D0`, `0xD0BC0` | candidate | Waypoint 0x49 → object; quest 0x40, 0x58 → quest. |
| net-server-unit-messages | D2Game `UNIT\SUnitMsg`, `PLAYER\PlrMsg`, `MONSTER\MonsterMsg` | none | `0x8B0F0`, `0x8B470`, `0xF34C0` | candidate | Senders; no server-side dispatch table. |
| net-server-command-table-client-dispatch | D2Client `GAME\SCmd` | none | D2Client `.data` RVA `0xDDE5C`, {unit handler, handler, size}, ids 0x00-0xAE | candidate | |
| net-server-session-and-map-commands | SCmd 0x00-0x0B, 0x59, 0x5B/0x5C | none | D2Client `0xADB50`, `0xADF20`, `0xAD8F0` | candidate | |
| net-server-unit-action-and-stat-commands | SCmd 0x0C-0x23, 0x4C/0x4D, 0x67-0x6D, 0x95, 0xAB | none | D2Client `0xAF240`, `0xAEDF0`, `0xAD0B0` | candidate | |
| net-server-item-skill-missile-state-commands | SCmd 0x3E-0x48, 0x73, 0x94-0xAC | none | D2Client `0xAEEE0`, `0xAE080`, `0xAE420` | candidate | Includes client item creation from packets (`UNIT\Item`, `0x833F0`, `0x85990`). |
| net-server-chat-npc-quest-party-commands | SCmd 0x26-0x2A, 0x52, 0x5A, 0x75-0x90 | none | D2Client `0xAF9F0`, `0xAFA70`, `0xAD520` | candidate | Includes roster (`GAME\Roster`, `RosterPets`). |
| (logon/realm, Warden packets) | D2Client `GAME\Msg` | none | outside the CCmd table | out of scope | Battle.net and Warden. |

## item

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| [item-durability-loss](item-durability-loss.md) | D2Game `ITEMS\Items`, `PLAYER\PlrModes` | ItemStatCost, ItemTypes, BodyLocs | `0x12EE0`, `0x78A90`, `0xDA480` | written | |
| item-creation-and-quality-roll | D2Game `ITEMS\Items` | Weapons/Armor/Misc level and rarity, ItemRatio, ItemTypes | `0x11490`, `0x11070`, `0x10A20` (fans out to `0xEE90`, `0xF370`, `0xEBB0`, `0xEC90`) | candidate | Owns quality roll and magic find (D2Common `0x36650`, `0x496F0`). |
| item-magic-and-rare-affix-selection | `ITEMS\Items` | MagicPrefix, MagicSuffix, RarePrefix, RareSuffix, AutoMagic | `0x14F20`, `0x15240`, `0x10730`, `0x14BF0`; rare name leads `0x13A30`, `0x13C20` | candidate | |
| item-unique-and-set-selection | `ITEMS\Items`; D2Common `DATATBLS\ItemTbls` | UniqueItems, SetItems, Sets | `0x103C0` → `0x144D0`, `0x148A0`; `0x14B30` → `0x14210`, `0x14370` | candidate | Includes fallback to rare (unresolved). |
| item-superior-and-low-quality | `ITEMS\Items` | QualityItems, LowQualityItems | `0x101E0` → `0x140B0`; `0xBEE30` | candidate | May fold into item-creation. |
| item-type-specific-init | `ITEMS\Items` | Misc quantity, ItemTypes | `0x10600` (switch 0..0x21), `0x10550`, `0x107E0` | candidate | Quantity, gold, keys, books. |
| item-sockets-and-ethereal-roll | `ITEMS\Items`; D2Common `ITEMS` | ItemTypes MaxSock*, gemsockets, nodurability | D2Common `0x46830`, `0x2D9B0`; D2Game `0xF370` | candidate | Ethereal roll site unresolved. |
| item-socket-insertion-and-runewords | D2Game `ITEMS\ItemMode`; D2Common `DATATBLS\ItemTbls` | Gems, Runes | `0xD8020`; D2Common `0x71C20`, `0x266E0` | candidate | |
| item-property-functions | D2Common `ITEMS` | Properties func1-7, stat1-7, set1-7, val1-7 | D2Common fn table VA `0x6FDEB0D8` (21 entries), stat-pair VA `0x6FDEA928`, per-stat VA `0x6FDEAD50`; dispatcher `0x45AA0` | candidate | Includes per-level stats 214-250. |
| item-stat-change-and-item-events | D2Game `ITEMS\ItemMode` | ItemStatCost itemevent1/2, itemeventfunc1/2 | `0xD9470`, `0xAD0F0` (table VA `0x6FD27238`) | candidate | Registration side; trigger side in combat-item-event-triggers-and-procs. May merge. |
| item-use-consumables | `ITEMS\ItemMode` | Misc pSpell, spelldesc, stat1/calc1, useable | `0xD8BA0`, `0xD9BF0`, `0xD8440`, `0xD4140` | candidate | Potions, scrolls, tomes, keys. |
| item-cube-recipes | `ITEMS`, `PLAYER\PlrTrade`; D2Common `DATATBLS` | CubeMain | `0x70240`, `0x72220`, `0x70CF0`; D2Common `0x1BA60`, `0x1C020` | candidate | May split into match and output. `0x72220` also listed under npc-gambling: identity unresolved. |
| item-set-bonus-activation | D2Common `ITEMS\Items` | SetItems aprop/add func, Sets PCode/FCode | D2Common `0x23F90`; D2Game `0xD5590` | candidate | |
| item-requirements-and-item-level | D2Common `ITEMS\Items` | reqstr, reqdex, levelreq, affix levelreq | D2Common `0x29F30`, `0x26190`, `0x26DA0`, `0x287D0` | candidate | Shared with inventory equip rules. |
| item-timed-events | `ITEMS\ItemMode` | ItemStatCost stats 252/253 | event table VA `0x6FD1B110` ([3] `0xDA480`, [4] `0xD1790`, [12] `0xD1780`); `0xEA10`, `0xEA70` | candidate | Replenish and recharge; partly cited by item-durability-loss. |
| item-vendor-value | D2Common `ITEMS\Items` | cost columns | D2Common `0x28E80`, `0x250B0` | out of scope | Merged into npc-buy-sell-and-pricing. |

## inventory

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| inventory-grid-placement | D2Common `INVENTORY\Inventory` | Inventory grids, invwidth/invheight | D2Common `0x1E1B0`, `0x1E360`, `0x1E6D0`, `0x20A30` | candidate | Inventory, stash, cube, trade pages. |
| inventory-equip-slot-rules | D2Game `ITEMS\ItemMode`; D2Common `INVENTORY` | BodyLocs, ItemTypes BodyLoc1/2, equiv, class, two-handed | `0xD2610`, `0xD1BC0`, `0xD51A0`, `0xCB1F0` | candidate | |
| inventory-belt | D2Common `INVENTORY`; D2Game `ITEMS\ItemMode` | Belts, ItemTypes Beltable | D2Common `0x200D0`, `0x4E7A0`; D2Game `0xD5C30`, `0xD7E30` | candidate | |
| inventory-pickup-drop-and-cursor | `ITEMS\ItemMode`, `PLAYER\Player` | none | `0xD6E30`, `0xD62E0`, `0xD8310`, `0x383F0` | candidate | Includes auto-pickup of gold and potions to belt. |
| inventory-stacking-and-quantity | `ITEMS\ItemMode` | stackable, maxstack, quantity stat | `0xD4A70`, `0xD7300`, `0xD8600`; D2Common `0x260F0` | candidate | |
| inventory-weapon-swap | `ITEMS\ItemMode` | none | `0xD46B0` | candidate | May fold into equip-slot rules. |
| inventory-load-fixups | D2Game `INVENTORY\InvMode` | ItemTypes, BodyLocs | `0x9B2E0`, `0x9B3B0`, `0x9B630` | candidate | Overlaps save-load-and-validation. |

## player

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| [player-party-experience-sharing](player-party-experience-sharing.md) | D2Game `UNIT\Party`, `PLAYER` | Experience, DifficultyLevels | `0xDEA00`, `0xDE8F0`, collector `0xDAED0` | written | Renamed 2026-10-15; gold pickup sharing is player-party-gold-sharing. |
| [player-party-gold-sharing](player-party-gold-sharing.md) | D2Game `UNIT\Party`, `PLAYER` | ItemTypes gold | split caller `0x9C270`; see spec | written | Split from player-party-experience-sharing (2026-10-15). |
| [player-hostility-declaration](player-hostility-declaration.md) | D2Game `UNIT\Party` | none | see spec | written | Split from combat-pvp-damage-and-kill-credit (2026-10-15). |
| player-life-mana-stamina-regeneration | D2Game regen tick | regen stats, stamina | `0x6CEE0` (table VA `0x6FD27BBC` slot 21); `0x8DA80` | candidate | Includes per-frame poison drain; table `0x6FD27BBC` (37 handlers) unidentified. Monster regeneration is in scope too. Mana part written as player-mana-regeneration: the per-frame player tick is `0x79B10` (type-3 unit event, slot 3 of table RVA `0xFBB74`) calling life `0x77CB0`, stamina `0x77A50`, mana `0x77950`; `0x6CEE0`/`0x6A560` is the object or monster refill tick, not player regeneration. Life and stamina regeneration remain candidates. |
| [player-mana-regeneration](player-mana-regeneration.md) | D2Game `PLAYER\PlrModes` | CharStats.ManaRegen; ItemStatCost rows 8, 9, 26, 27; States 85, 106 | `0x79B10` -> `0x77950`; D2Common ordinals 10973, 10186, 10494, 10871, 11111, 11108, 10551 | written | Draft: handler requests and arithmetic verified statically; only the dead-player mode-17 exit is runtime-confirmed. |
| [player-stamina-drain-while-running](player-stamina-drain-while-running.md) | D2Game `PLAYER\PlrModes` | CharStats.RunDrain; Armor.speed; ItemStatCost row 154; BodyLocs row 3 | `0x79A10` -> mode table `.data` RVA `0x1079D0` slot 3 -> `0x79B90` -> `0x77BB0` | written | Draft: static only; split out of player-modes-and-transitions. |
| player-level-up-and-attribute-points | D2Game `PLAYER\Player` | CharStats per-level columns, Experience | `0xEB090`, `0xEAEE0`, `0xE7940` | candidate | Effect only; commands in `net`. |
| player-modes-and-transitions | `PLAYER\PlrModes` | PlrMode, CharStats | mode table VA `0x6FD1A3F8`; `0x78430`, `0x78610`, `0x79A10`; walk/run `0x78710`, `0x78760` | candidate | Stamina drain on run probably here. Stamina drain written as player-stamina-drain-while-running; the rest remains a candidate. |
| player-death-corpse-and-penalty | `PLAYER\PlrModes`, `Player`; D2Common | DifficultyLevels, gold bank | `0x79730`, `0xD5740`; D2Common `0x2C7C0` | candidate | Overlaps item-durability-loss DUR-R09. |
| player-creation-and-starting-items | `PLAYER\PlrModes` | CharStats item1-10, StartSkill | `0x79210`, `0x78DD0` | candidate | |
| player-pets-and-summon-limits | `PLAYER\PlayerPets` | PetType | `0x96FA0`, `0x97A20`, `0x964B0`, `0x21FB0`; D2Common `0x5F310` | candidate | `GAME\Targets` summon lists (`0xEFE40`) belong here too. |
| player-hireling-lifecycle | `PLAYER\PlayerPets`, `UNIT\SUNITNPC` | Hireling, MonStats | `0xC1250`, `0xC1030`, `0x96EA0`, `0x98120`, `0xC0B90` | candidate | Hiring cost in npc; XP share written. |
| player-trade | `PLAYER\PlrTrade` | none | `0x70FB0`, `0x71E10`, `0x73AA0`, `0xE95E0` | candidate | |
| player-party-membership-and-hostility | `UNIT\Party` | none | `0x9BE40`, `0x9BF70`, `0x9C840`, `0xE9A10` | candidate | Check against combat-pvp-damage-and-kill-credit; hostility declaration is written in player-hostility-declaration. |
| player-npc-introductions | `PLAYER\PlrIntro` | none | `0x692D0`, `0x69360` | candidate | Small; may fold into npc or quest. |

## save

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| save-character-file-format | D2Game `PLAYER\PlrSave`, `PlrSave2` | none | `0x53D10`, `0x54000`, `0x53F10`, `0xEC7F0`, `0xEC0E0` | candidate | |
| save-item-bit-encoding | D2Common `ITEMS`, `Stats\StatsEx` | ItemStatCost Save Bits, Save Add, Save Param Bits | D2Common `0x27180`, `0x2A600`, `0x29D90`, `0x2D1D0` | candidate | |
| save-load-and-validation | `PLAYER\PlrSave` | CharStats, Experience | `0x56250`, `0x55E10`, `0x54BC0`, `0x54720` | candidate | |
| save-mercenary-and-corpse-records | `PLAYER\PlrSave`, `PlrSave2` | Hireling | `0x55AD0`, `0xEC470` | candidate | May fold into file format. |

## treasure

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| treasure-class-picks-and-nodrop | TreasureClassEx | Picks, NoDrop, Prob1-10, Item1-10, group, level | D2Game `0x12380`, `0x12D60` | candidate | NoDrop player-count part written in monster-scaling. |
| treasure-class-loading-and-item-type-tcs | D2Common `DATATBLS` | TreasureClassEx; auto weapon/armor TCs | D2Common `0x5A030`, `0x547D0`, `0x54430` | candidate | |
| treasure-monster-drop-selection | TreasureClassEx | MonStats TreasureClass1-4; SuperUniques TC; TC upgrade | D2Game `0x11CA0`, `0x11A70` | candidate | |
| treasure-item-quality-and-magic-find | TreasureClassEx; `ITEMS` | TC Unique/Set/Rare/Magic; ItemRatio; MF stat | D2Common `0x36650`, `0x496F0` | out of scope | Merged into item-creation-and-quality-roll (item owns quality and magic find). |
| treasure-gold-drops | TreasureClassEx | gld; gold-find stat | D2Game `0x120E0`, `0x12B90` | candidate | |

## npc

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| npc-vendor-inventory-and-refresh | D2Game `UNIT\SUNITNPC` | Items vendor min/max/magic columns; npc | `0xBE2A0`, `0xBF2E0`, `0x90320` | candidate | |
| npc-buy-sell-and-pricing | `UNIT\SUNITNPC`; `ITEMS` | npc buy/sell multipliers, max buy; item cost | `0xC00C0`, `0xD34B0`; D2Client `0x2C4F0`, `0x81D70`, `0x9F460` | candidate | Includes client price display. |
| npc-gambling | `UNIT\SUNITNPC` | Gamble | `0x72220`; D2Common `0x72C50`; D2Client `0x46C00`, `0x47860` | candidate | |
| npc-repair | `UNIT\SUNITNPC` | npc repair multiplier; durability | `0xCFBC0`; D2Common `0x3C020`; D2Client `0x45B40` | candidate | |
| npc-hiring-roster-and-resurrection | `UNIT\SUNITNPC`; D2Client `UI\Hireables` | Hireling | `0xC1250`, `0xC1030`, `0xC0B90`; D2Common `0x58B60`; D2Client `0x5B260`, `0x5B6C0`, `0x5BCC0`, `0x9AE80` | candidate | Includes the hireling panel. Hireling leveling is in `player`. |
| npc-identify-heal-and-services | `UNIT\SUNITNPC` | none | `0xEA5F0`, `0x69C00` | candidate | Identify path not found. |
| npc-menu-and-interaction-flow | D2Client `UI\npcmenu`, `UI\dialog` | npc menu definitions, speech | D2Client `0x46230`, `0x45BD0`, `0x4A270`, `0x4B420`, `0x46E00`; D2Game `0xBF810` | candidate | |

## quest

Quest init table: D2Game RVA `0x10DBF0`, 37 rows {init fn, act, flag, flag, quest id, QuestRecord
slot}; act-level rows at `0x10DF6C` (ids 37-40).

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| quest-system-core-callbacks-and-timers | D2Game `QUESTS\QUESTS` | none | `0x348A0`, `0x35090`, `0x33B70` | candidate | Includes quest packets. |
| quest-rewards | `QUESTS` | Items, ItemTypes | `0x38090`, `0xC23C0`, `0xC4900` | candidate | Skill/stat points, resistances, imbue, socket, personalize, respec. |
| quest-record-and-save-bits | D2Common `QuestRecord` | none | D2Common `0x30970`, `0x30880`, `0x2D020` | candidate | Bit layout per slot (0-41). |
| quest-act1-state-machines | `Quests\a1q0`-`a1q6` | Levels, SuperUniques, Objects | `0x62D90`, `0x65990`, `0x69200`; Cairn stones op 9 `0x664F0` | candidate | |
| quest-act2-state-machines | `Quests\a2q0`-`a2q8` | same | `0xC5690`, `0xC7B40`, `0xC9F60` | candidate | a2q3 partly covered by the world-clock spec. |
| [quest-tainted-sun](quest-tainted-sun.md) | a2q3 | none | `0xC7480`, op 24 `0xC76B0` | written | Quest stages; the world clock is drlg-world-clock (split 2026-10-15). |
| quest-act3-state-machines | `Quests\a3q0`-`a3q7` | same | `0xE6660`, `0xE93C0`, `0xEA080` | candidate | |
| quest-act4-state-machines | `Quests\a4q1`-`a4q3` | same | `0xB9D80`, `0xBAE40`, `0xBC5F0`, `0xBC220` | candidate | Seals and Diablo. |
| quest-act5-state-machines | `Quests\a5q1`-`a5q6` | same | `0x5BEA0`, `0x5FC20`, `0x60080` | candidate | |
| quest-secret-and-act-transition-entries | `QUESTS` | Levels | ids 25-30: `0x62FA0`, `0xC4880`, `0xE5920`, `0xB97D0`; ids 37-40: `0x63CC0`; leads `0xB9F30`, `0xB96A0` | candidate | Purpose unknown; leads: Cow Level, act travel, Uber/Pandemonium. |

## object

Operate table: D2Game `.data` RVA `0x107BB8`, indexed by Objects `OperateFn` (0-73; 35-38 and
60 null).

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| object-operate-dispatch-and-modes | D2Game `OBJECTS\Objects` | Objects OperateFn, Mode*, Flags | `0x6A720`, `0x6A330`, `0x69DA0` | candidate | |
| object-chests-containers-and-traps | `Objects` | Objects, TreasureClassEx | op 4 `0x6F040`; ops 39-41 `0xCA050`, `0xCA170`; `0x3C160` | candidate | |
| object-shrines-and-wells | `Objects` | Shrines | op 2 `0x6DA00`; `0xB800`; `0x6E930` | candidate | Shrine op number is a lead. |
| object-doors-and-misc-toggles | `Objects` | Objects | op 13 `0x69AD0`; op 11 `0x69A80` | candidate | |
| object-town-portals-and-red-portals | `Objects`, `QUESTS` | none | op 15 `0x6D0F0`; op 34 `0x6D620`; `0x50800`; `0x660D0` | candidate | |
| object-waypoints | `Objects`; D2Common `WayPoint` | Levels Waypoint | op 23 `0x6CF90`; `0x34D00`; CCmd 0x49 `0xD0D30` | candidate | Client panel in `ui-waypoint-panel`. |
| object-region-population | D2Game `OBJECTS\objrgn` | Objects, ObjGroup | `0xB9A0`, `0x32CE0`, `0x7B510` | candidate | |
| object-quest-operate-functions | `Objects`, `Quests` | Objects | ops 21, 25, 33, 49, 52-59, 62-73 (e.g. `0x65D40`, `0xC6240`, `0xBB8D0`) | candidate | May fold into the per-act quest rows. |

## drlg

Level-type dispatch: D2Common `0x2E360` (ordinal 10322): 1 maze `0x11BF0`, 2 preset `0xCD10`, 3
outdoor `0x43580`.

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| drlg-level-type-dispatch-and-seeding | D2Common `DRLG\Drlg` | Levels, LvlTypes | `0x2E360`, `0x3C800` | candidate | |
| drlg-maze-generation | `DRLG\Maze` | LvlMaze | `0x11BF0`, `0x10160`, `0x10AF0`; special cases `0x10CA0`, `0x114C0`, `0x116B0` | candidate | |
| drlg-preset-rooms | `DRLG\Preset` | LvlPrest | `0xB560`, `0xC690`, `0xB320` | candidate | |
| drlg-outdoor-generation | `DRLG\Outdoors`, `OutRoom` | Levels, LvlPrest | `0x43580`, `0x43190`, `0x17AE0` | candidate | May need one spec per act. |
| drlg-outdoor-preset-placement | `DRLG\OutPlace`, `OutJung` | LvlPrest | `0x15A80`, `0x14D00`, `0x85D20`; act 4/5 `0x84DA0`, `0x854C0`, `0x856C0` | candidate | |
| drlg-room-tiles-and-warps | `DRLG\RoomTile` | LvlWarp | `0x7F3A0`, `0x19FA0`, `0x6C5D0` | candidate | |
| drlg-vertices-grid-logic | `DRLG\DrlgVer`, `DrlgGrid`, `DrlgLogic` | none | `0x1B6E0`, `0x367F0`, `0x81020` | candidate | Includes wall visibility. |
| drlg-level-object-placement | `DRLG\DrlgRoom` | Objects, ObjGroup, Shrines | `0x41730`, `0x41070`, `0x41CA0` | candidate | |
| drlg-animated-tiles | `DRLG\DrlgAnim` | none | not found in d2-decomp | candidate | Entry point not located. |
| drlg-environment-day-cycle | D2Common `ENVIRONMENT\Env` | none | `0x37570`, `0x37BA0`, `0x37920` | candidate | Eclipse part is written ([drlg-world-clock](drlg-world-clock.md)). |
| [drlg-world-clock](drlg-world-clock.md) | D2Common `ENVIRONMENT` | none | see spec | written | Split from quest-tainted-sun (2026-10-15). |

## path

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| path-types-and-search | D2Common `PATH\Path` | MonStats2 / Missiles path type | path-type table `.rdata` RVA `0x942E8` (14 entries); `0x7B630`; `0x708F0` | candidate | |
| path-movement-step-and-speed | `PATH\Path` | none | `0xD8E0`, `0xD240`, `0x37310` | candidate | |
| path-wander-and-flee-paths | `PATH\Path` | none | `0x66AA0`, `0x67AD0`, `0x18830` | candidate | |
| path-collision-grid-and-masks | `COLLISN\Collisn` | Objects collision templates | `0x4D010`, `0x4B780`, `0x4CA90` | candidate | |
| path-line-of-sight | `PATH\Path`, `COLLISN\Collisn` | none | `0x66FB0`, `0x3E6E0` | candidate | |
| path-unit-placement-and-teleport | `PATH\Path`, `COLLISN\Collisn` | none | `0x1A130`, `0x4D5C0`, `0xDDC0` | candidate | |

## ui

All D2Client.

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| ui-item-description-text | `UI\inv`, `UNIT\Item` | ItemStatCost descpriority, descfunc, descval, descstr*, dgrp*; Sets, SetItems, UniqueItems | `0x914F0`, `0x928C0`, `0x55B20`; sets `0x55F10`, sockets `0x55FB0`, ethereal `0x8E180` | candidate | Likely splits: stat-line formatting, tooltip assembly. |
| ui-item-name-and-color | `UI\inv`, `UNIT\Item` | quality, affix names | `0x8DFB0`, `0x20620` | candidate | May fold into the row above. |
| ui-skill-description-text | `UI\SkillDesc` | SkillDesc page, row, column, str*, desc*, dsc2*, dsc3*, ddam calc | `0x2DD50`, `0x2E770`, `0x307B0`; damage calc `0x2E410`, `0x2C2D0` | candidate | Large; split into damage calc and line formatting. |
| ui-monster-and-mercenary-stat-text | `UI\SkillDesc`, `UI\Hireables` | Hireling, MonStats | `0x33B50`, `0x54E10`, `0x2FB50` | candidate | |
| ui-skill-tree-and-skill-select | `UI\spellsel`, `SKILLS\SkillsBar` | SkillDesc page, row, column, ListRow, IconCel | `0x3CFA0`, `0x359B0`, `0x3D080`, `0x3A910` | candidate | Includes hotkey slots. |
| ui-inventory-panels-and-item-placement | `UI\inv`, `UI\useitem` | Inventory, ItemTypes, BodyLocs | `0x8C2B0`, `0x958A0`, `0x9B530` | candidate | Client display and click mapping only; rules in `inventory-`. |
| ui-belt-and-cube-panels | `UI\inv` | Belts, ItemTypes | `0x7F520`, `0x8CA80`, `0x4AEE0` | candidate | |
| ui-character-stat-panel | `UI\panel` | CharStats, ItemStatCost | `0x8EA20`, `0xC30E0`, `0x28DC0` | candidate | |
| ui-orbs-and-control-panel | `UI\panel`, `UI\ui` | none | `0x27BD0`, `0x27590`, `0x8CFB0` | candidate | |
| ui-automap | `UI\automap` | AutoMap, Levels | `0x62760`, `0x61490`, `0x619E0`; fade `0x60950`; map file `0x5F290`, `0x61E70` | candidate | |
| ui-minimap-and-party-markers | `UI\automap`, `UI\Party` | none | `0x266C0`, `0x39DC0`, `0x7C270` | candidate | |
| ui-chat-input-and-history | `UI\chat`, `UI\editbox` | none | `0x6F200`, `0x6EE40`, `0x37F80` | candidate | |
| ui-chat-commands | `UI\chat` | none | `0x6EC30`, `0x39C80`, `0x143E0`, `0x14630` | candidate | Includes chat throttle. |
| ui-key-bindings | `UI\CmdTbl` | keyboard map | `0x39910`, `0x65E70`, `0x125A0` | candidate | |
| ui-overhead-and-scrolling-text | `UI\text` | none | `0x7CA40`, `0x7C690`, `0x7D850` | candidate | |
| ui-ground-item-labels | `UI\showitems` | none | `0x58FB0`, `0x58DE0` | candidate | Label overlap layout. |
| ui-unit-name-and-health-tooltip | `UI\ui` | MonStats, SuperUniques | `0xBFEC0`, `0x3D8A0`, `0x68CF0` | candidate | |
| ui-party-panel | `UI\Party`, `GAME\Roster` | none | `0x760F0`, `0x9E3D0`, `0x218C0` | candidate | |
| ui-quest-log | `UI\QuestLog` | quest record bits | `0x1DA90`, `0x1E160`, `0x1D740` | candidate | |
| ui-waypoint-panel | `UI` panel family | Levels Waypoint | `0xAAEC0`, `0xAA4E0`, `0xAAA20` | candidate | |
| ui-unit-interaction-and-click-actions | `UNIT\CUnit`, `UNIT\Player`, `UNIT\UnitMode` | none | `0xA5B90`, `0xA68E0`, `0x1B7F0`, `0x1BC40`, `0x29F30` | candidate | Which click does what on which unit. |
| ui-help-and-dialog-windows | `UI\dialog`, `UI\ui` | none | `0xA3E40`, `0x37C80` | candidate | Low gameplay value. |

## draw

All D2Client.

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| [draw-light-map-and-light-radius](draw-light-map-and-light-radius.md) | `DRAW\dLightMap` | Radius/color, prepared object/missile/monster fields, component light selection, separate light associations, collision inputs, shadow scratch/cache and map-cell writes | `0xA8C20`, `0xA9B50`, `0xA9100`, `0xA8A40` | written | Draft: prepared-input shadow/raster, temporary/projectile/monster/overlay light and cleanup stages verified; collision masking and one temporary-light creation/attachment case runtime-confirmed. Source/world producers and remaining gameplay observations stay explicit. `0x1BC0` was rejected as a sprite-composition lead; see companion. |
| draw-room-and-ambient-lighting | `DRAW\dLightMap` | Levels, LvlTypes lighting | `0x4F770`, `0xA8520` | candidate | May merge with the row above. |
| draw-palette-shifts | `GAME\PalShift` | States colorshift, colorpri; MonStats2 Utrans | `0x538F0`, `0x53B20`, `0x539E0`, `0x40190` | candidate | |
| draw-unit-color-transform-and-glow | `ENGINE\GfxUtil` | Items InvTrans; MonStats2 | `0x6DEE0`, `0x6DFE0`, `0x52200` | candidate | |
| draw-unit-sprite-composition | `ENGINE\Gfx` | COF layers; MonStats2 components | `0x6C490`, `0x6CC00`, `0x6BF30` | candidate | Large. |
| draw-overlays | `ENGINE\Gfx` | Overlay; States overlay columns | `0x6B300`, `0x2A590`, `0x40450` | candidate | |
| draw-shadows | `ENGINE\Gfx`, `GAME\Wall2` | MonStats2 Shadow | `0x6BA90`, `0x7A180` | candidate | |
| draw-wall-fading-and-draw-order | `GAME\Wall2` | tile orientation | `0x7A340`, `0x7A5D0`, `0x87A90` | candidate | |
| draw-particles-and-weather | `ENGINE\Particle` | Levels weather | `0x80700`, `0x804C0`, `0xC7FB0` | candidate | |
| draw-view-perspective-and-screen-shake | `GAME\View` | none | `0x8A090`, `0x8AFD0`, `0xC4830` | candidate | |
| draw-unit-selection-and-hover | `GAME\Select` | MonStats2 selectable, bounds | `0x51650`, `0x16E50`, `0x6A550` | candidate | |
| draw-screen-fades | `ENGINE\Gfx`, `UI\dialog` | none | `0x10EC0`, `0x524D0` | candidate | Small; may fold into the view row. |

## sound

D2Client selection/coordination leads; D2Sound playback ownership is also a discovery target.

| Candidate spec | Original source | Data tables and columns | 1.13c entry points | Status | Notes |
|---|---|---|---|---|---|
| sound-table-and-playback-rules | `Sound\SoundHdr` | Sounds Volume, Group Size, Loop, Fade In/Out, Defer Inst, Stop Inst, Priority, Falloff, Redirect, Stereo | `0x5DFD0`, `0x5D7A0`, `0x5CE70` | candidate | |
| sound-3d-position-and-falloff | `Sound\SoundHdr` | Sounds Falloff | `0x5D1D0`, `0x5DCE0`, `0x88440` | candidate | May merge with the row above. |
| sound-unit-sound-selection | `UNIT\UnitSnd` | MonSounds; PlrType sounds | `0x22490`, `0x22CA0`, `0x22C10`, `0x23C00` | candidate | |
| sound-ambient-and-weather | `UNIT\UnitSnd` | Levels, LvlTypes sound environment | `0x88D90`, `0x4F860`, `0x8AD30` | candidate | |
| sound-skill-and-missile-sounds | `SKILLS`, `UNIT\Missile` | Skills stsound, dosound; Missiles TravelSound, HitSound, ProgSound; States onsound, offsound | `0x40A70`, `0x14F40`, `0x2BE0` | candidate | |
| sound-music-selection-and-transitions | D2Client / D2Sound, exact selection owner unresolved | Music/environment/level bindings require discovery; no column meaning established | Unresolved: trace selection and transition requests into original audio operations | candidate | Track selection, looping and fades; issue suggestions are leads. Establish trigger/ownership and table consumers before writing rules. |

## Out of scope

| Area | Original source | Reason |
|---|---|---|
| Warden anti-cheat | D2Client `WARDEN\WardenClient` | Not gameplay. |
| Archive I/O and process shell | D2Client `CORE\ARCHIVE`, `CORE\WINMAIN`; D2Hell `Archive` | Platform layer. |
| Battle.net, realm, video drivers | various | Not single-player or open-game behavior. |
