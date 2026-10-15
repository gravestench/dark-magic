# Game Data Fields

## Description

The game uses shared definitions for monsters, items, skills, and effects. These definitions let its different systems identify the same creature, ability, or effect consistently. This reference describes the original game's shared names and data fields.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [Numeric and RNG foundations](core-numeric-and-rng-foundations.md)
Ids: GDF-R, GDF-Q, GDF-V, GDF-O; policy in [README conventions](README.md#conventions)
Evidence: [source, table records, and observation history](../research/spec-evidence/core-game-data-fields.md)
Open evidence gaps: [remaining assertions](../plans/gaps/core-game-data-fields.md)

## 1. Scope

This dictionary defines the original profile's table columns, identity mappings, field widths, selected record getters, flag tests, and the bounded table-loading and treasure-field conversion behavior stated below. It defines logical records and results. Gameplay formulas using these fields belong to their consuming specs. Shared stat/state operations belong to NRF-R12. Expression evaluation belongs to [Skill calc expressions and synergies](skill-calc-expressions-and-synergies.md).

Full text compilation remains gated where lookup, callback, locale, or missing-column behavior is unresolved. A dictionary or supplied-record getter can be accepted without claiming reproduction of that compiler.

## 2. Trigger and ownership

- D2Common prepares shared game definitions during process initialization, before a game is created. These are initialization operations, not frame updates.
- D2Game and D2Common consumers request records or fields when their gameplay paths need them. The getters stated here are read only.
- Difficulty loading requires three logical rows. Treasure-class loading converts selected source fields to runtime fields after definitions are read.
- Client and server use shared definitions. This dictionary defines no message.

## 3. Inputs and data

Every table path below is under data/global/excel/ in the original 1.13c archive set. Columns present in the effective files use their text-header spellings. Descriptor-only columns use the original schema's names; their missing-column behavior remains gated.

| Input | Semantic source | Width / format | Meaning |
|-------|-----------------|----------------|---------|
| schema | named table, column, conversion kind, string limit, lookup or callback binding | logical descriptors | selects fields prepared for a table |
| text table | effective named .txt table | 8-bit byte strings, tab-separated cells and CRLF rows | source identities and values; string/code limits count bytes |
| prepared table | named ordered record collection and record count | logical records; count i32 | supplied input to getters |
| table index | caller's class, skill, state, monster, or item-type identity | i32 unless a field width is stated | selects a row |
| difficulty | game difficulty | u8; 0 Normal, 1 Nightmare, 2 Hell | selects difficulty columns |
| skill level | caller-supplied level | i32 | GDF-R08b parameter calculation |
| monster mode | caller-supplied original monster-mode identity | i32, domain 0..15 | selects GDF-R06b availability; identity mapping is GDF-R15a |
| state presence | unit state collection | named state ids and boolean membership | consumed through NRF-R12a |
| flags | named fields of supplied logical records | boolean fields | selected flag-test results |
| generation mode | generate compiled data; read compiled data | independent boolean conditions | GDF-R20a selection |
| load/decode result | selected resource availability and decoded records/count | success/failure, logical collection | supplied result at GDF-R20a boundary |
| treasure source and destination | named source fields; destination initialized fields | u16 group/level/quality fields; i32 Picks/NoDrop | GDF-R10 conversion |
| generated treasure group count | item-type treasure groups generated before this conversion | u16 original counter, supplied as i32 | group adjustment |
| reserved treasure words | supplied unnamed source fields | two u16 values | GDF-R10b copies |
| ASCII comparison mode | original ASCII comparison branch | boolean selection | GDF-R01d boundary |
| optional published count | initialization caller choice | boolean request | controls whether a count is published |

| Mapping / constant | Origin | Binding |
|--------------------|--------|---------|
| stat/state names and explicit ID/id cells | ItemStatCost / States data | original-profile identities corroborated with prepared-table indices; consumers may also fix individual ids |
| player/monster mode names, token/code pairs | PlrMode / MonMode data | original-profile row identities; consumers choose code-defined mode numbers |
| monster identities | MonStats Id/hcIdx; linked MonStats2 Id | original-profile prepared-record identities; separator rows excluded |
| item-type identities | ItemTypes ItemType/Code | original-profile row identities; link-building conversion defined by code |
| skill identities | Skills skill/Id | original-profile prepared-record indices |
| handler selectors such as srvdofunc/`itemeventfunc1` | named table columns | data selects engine-defined handlers; a selector is not a skill/stat id |
| conversion kinds, field widths, count 3, bounds, return masks | engine code | fixed by original code |
| custom identity remapping | unresolved | no support or behavior is established here |

[Table records](../research/spec-evidence/core-game-data-fields.md#table-verification-2026-10-14) distinguish text cells from compiled correspondence. An ID cell matching a row index does not establish that the compiler uses that cell to assign every linked identity.

## 4. Rules

All operations below draw no RNG.

**GDF-R01a. Column schema** [verified in 1.13c code (static)]. A table's schema identifies named columns and conversion kinds. A descriptor with conversion kind Stop ends the schema, regardless of its displayed name. A descriptor named end with another conversion kind does not terminate it. Initialization publishes prepared records; it publishes their count only when requested. A schema proves field identity, not gameplay meaning.

**GDF-R01b. Direct text-cell conversions** [verified in 1.13c code (static)]. At an already selected schema-field/cell boundary:
- Text characters and string/code limits here mean original 8-bit bytes, not Unicode code points. Numeric fixtures use ASCII decimal bytes; string/code copying retains the supplied byte order.
- Decimal cells accumulate value = i32(value * 10 + character - '0'); a leading minus negates the completed i32 value. Empty numeric cells yield 0. This contract covers empty or well-formed decimal text.
- Numeric stores retain the low 32, 16, or 8 bits according to the selected field. Signedness of a later read belongs to its consumer.
- A flag is false for numeric zero and true for any nonzero parsed number; other named flags are unchanged.
- A direct four-byte code takes at most four text characters, retains their order, and pads missing characters with spaces. Empty text yields four spaces.
- A bounded string retains at most its descriptor's byte limit.
These conversions do not establish header selection, name lookup, callback execution, or missing-column defaults.

**GDF-R01d. ASCII header comparison** [verified in 1.13c code (static)]. On the original ASCII comparison path, header names compare without distinguishing A..Z from a..z. This permits the descriptor name stat to match the header Stat. Duplicate and missing-header defaults remain GDF-Q12. Locale-dependent non-ASCII comparison is outside this ASCII operation.

**GDF-R01c. Other text conversions** [unresolved]. Name/code links, expressions, callback fields, descriptor-specific callback outputs, and missing schema columns must match the original compiler. GDF-Q06 / GDF-Q12 gate their complete contracts; GDF-R01b supplied-cell conversions remain available.

**GDF-R02. Difficulty fields** [verified in 1.13c code (static)]. Rows Normal, Nightmare, Hell correspond to difficulty 0, 1, 2. The fields below are i32. A prepared count other than 3 aborts this initialization.

| Columns |
|---------|
| ResistPenalty, DeathExpPenalty, UberCodeOddsNormal, UberCodeOddsGood |
| MonsterSkillBonus, MonsterFreezeDivisor, MonsterColdDivisor, AiCurseDivisor |
| UltraCodeOddsNormal, UltraCodeOddsGood, LifeStealDivisor, ManaStealDivisor |
| UniqueDamageBonus, ChampionDamageBonus, HireableBossDamagePercent |
| MonsterCEDamagePercent, StaticFieldMin, GambleRare, GambleSet, GambleUnique, GambleUber, GambleUltra |

HireableBossDamagePercent does not itself prove its consumer's boss predicate.

**GDF-R03a. Player-class fields and getter** [verified in 1.13c code (static)]. Select the supplied CharStats collection by class index. Negative indices and indices at or above count yield no record. BlockFactor is read as u8.

| Columns | Logical format |
|---------|----------------|
| class | string, byte limit 15 |
| ToHitFactor | i32 |
| str, dex, int, vit, stamina, hpadd, ManaRegen, WalkVelocity, RunVelocity, RunDrain | u8 |
| LifePerLevel, StaminaPerLevel, ManaPerLevel, LifePerVitality, StaminaPerVitality, ManaPerMagic, BlockFactor, StatPerLevel | u8 |
| StartSkill | linked skill identity, 16-bit |

**GDF-R03b. Class identities** [verified in 1.13c data]. Prepared class order is 0 Amazon, 1 Sorceress, 2 Necromancer, 3 Paladin, 4 Barbarian, 5 Druid, 6 Assassin. The text Expansion separator at zero-based data row 5 is not a class.

**GDF-R03c. Class display-name preparation** [unresolved]. The localized display name prepared after loading must match the original localization operation. GDF-Q13 gates that result, not the supplied class field or getter.

**GDF-R04. Experience fields** [verified in 1.13c code (static)]. Experience has i32 fields Amazon, Sorceress, Necromancer, Paladin, Barbarian, Druid, Assassin, ExpRatio. Class index 0..6 selects the matching class field independently of header order. Initialization publishes its supplied decoded records without a separately stored count. Its consumer owns row selection and bounds.

**GDF-R05a. Monster fields and getter** [verified in 1.13c code (static)]. Negative indices and indices at or above the supplied MonStats count yield no record. Valid indices yield that monster's record. MonStatsEx is a linked MonStats2 identity read as signed i16 by the mode getter.

| Columns | Logical format |
|---------|----------------|
| Id, BaseId, MonStatsEx, MissA1 | 16-bit identities |
| Velocity, Run, Level, Level(N), Level(H) | 16-bit numeric words |
| `aip1`..`aip8`, each with base/(N)/(H) variants | 16-bit numeric words |
| MinHP, MaxHP, AC, Exp, each with base/(N)/(H) variants | 16-bit numeric words |
| Align, aidel, aidel(N), aidel(H), Drain, Drain(N), Drain(H), ToBlock, ToBlock(N), ToBlock(H), Crit | u8 |
| coldeffect, coldeffect(N), coldeffect(H) | 8-bit numeric words; signed interpretation belongs to the consumer |
| isSpawn, isMelee, noRatio, opendoors, SetBoss, BossXfer, boss, primeevil, npc, interact, inTown, lUndead, hUndead, demon, flying, killable, switchai, nomultishot, neverCount, petIgnore, deathDmg, genericSpawn, zoo, placespawn, inventory, enabled, NoShldBlock, noAura, rangedtype | named boolean flags |

Difficulty selects base/(N)/(H) in domain 0..2. Drain is read as u8 by leech.

**GDF-R05b. Derived monster animation values** [unresolved]. Initialization prepares two additional animation-rate values from base-class animation data and Velocity/Run relationships. These are not text columns. GDF-Q01 gates their complete production. Supplied-value consumers retain their own evidence labels.

**GDF-R06a. Extended monster fields** [verified in 1.13c code (static)]. MonStats2 has HitClass (u8), direction counts dDT/dNU/dWL/dGH/dA1/dA2/dBL/dSC/dS1..dS4/dDD/dKB/dSQ/dRN (u8), corresponding availability flags mDT..mRN, and boolean noGfxHitTest/noMap/noOvly/isSel/alSel/noSel/shiftSel/corpseSel/revive/isAtt/small/large/soft/critter/Shadow/noUniqueShift/compositeDeath/inert/objCol/deadCol/unflatDead. inTown belongs to MonStats.

**GDF-R06b. Mode-availability getter** [verified in 1.13c code (static)]. For monster class index and mode 0..15, return 0 if the MonStats index is invalid. Read MonStatsEx as signed i16; return 0 if negative, at or above MonStats2 count, or absent extended record. Otherwise select that mode's availability flag: false returns 0; true returns 1 << (mode mod 8). A true result is not always 1. Modes outside 0..15 are outside this logical operation.

**GDF-R07. Missile fields** [verified in 1.13c code (static)]. Missiles has dParam1/dParam2 (i32); HitClass/KnockBack/CollideType/Collision/ClientCol/CollideKill/CollideFriend/NextHit/NextDelay/Size/ToHit/AlwaysExplode (u8); Skill (linked 16-bit skill); and boolean LastCollide/Explosion/Pierce/CanSlow/CanDestroy/ClientSend/GetHit/SoftHit/ApplyMastery/ReturnFire/Town/SrcTown/NoMultiShot/NoUniqueMod/Half2HSrc/MissileSkill. Gameplay meanings belong to consumers.

**GDF-R08a. Skill fields and getter** [verified in 1.13c code (static)]. Negative skill indices and indices at or above count yield no supplied Skills record.

| Columns | Logical format |
|---------|----------------|
| srvstfunc, srvdofunc, ItemEffect | 16-bit numeric selector |
| aurafilter, Param1..Param8, ToHit, LevToHit, MinDam, MaxDam, MinLevDam1..5, MaxLevDam1..5, EMin, EMax, EMinLev1..5, EMaxLev1..5, ELen, ELevLen1..3 | i32 |
| HitShift, SrcDam | u8 |
| EType | linked element code, 8-bit |
| aurarangecalc, `calc1`..`calc4`, ToHitCalc, DmgSymPerCalc, EDmgSymPerCalc, ELenSymPerCalc | compiled expression word, 32-bit; representation/evaluation owned by the expression spec |
| decquant, lob, progressive, finishing, passive, aura, periodic, prgstack, InTown, Kick, InGame, repeat, stsuccessonly, stsounddelay, weaponsnd, immediate, noammo, enhanceable, durability, UseAttackRate, TargetableOnly, SearchEnemyXY, SearchEnemyNear, SearchOpenXY, TargetCorpse, TargetPet, TargetAlly, TargetItem, AttackNoMana, ItemTgtDo, leftskill, interrupt, TgtPlaceCheck, ItemCheckStart, ItemCltCheckStart, general, scroll, usemanaondo, warp | boolean |

**GDF-R08b. Parameter calculation** [verified in 1.13c code (static)]. Return 0 for invalid skill index, absent record, or signed level <= 0. Otherwise return i32(Param3 + i32(i32(level - 1) * Param4)). Multiplication/addition wrap as i32, with no clamp.

**GDF-R09. Stat-definition fields** [verified in 1.13c code (static)]. ItemStatCost has Stat (linked 16-bit stat), `itemevent1`/`itemevent2` (linked 16-bit event), `itemeventfunc1`/`itemeventfunc2` (16-bit selector), ValShift/Send Bits/Send Param Bits/op (u8), MinAccr/Divide/Multiply/Add (i32), maxstat/op base/op `stat1`..3 (linked 16-bit stat), and boolean Send Other/Signed/damagerelated/itemspecific/direct/UpdateAnimRate/fMin/fCallback/Saved/CSvSigned. NRF-R12c owns the ValShift/MinAccr consumer behavior.

**GDF-R10a. Treasure source fields** [verified in 1.13c code (static)]. TreasureClassEx has Treasure Class (string, limit 32); group/level/Unique/Set/Rare/Magic (16-bit words); Picks/NoDrop (i32); Item1..10 (strings, limit 64); Prob1..10 (i32). Two reserved words have no descriptor column.

**GDF-R10b. Selected treasure conversion** [verified in 1.13c code (static)]. With supplied source, initialized destination, generated group count, and reserved words:
- Nonzero group writes low16(i32(group + generated group count)); group 0 leaves destination group unchanged.
- Copy level and Magic/Rare/Set/Unique words unchanged.
- Copy Picks as i32, replacing exactly 0 with 1. Negative Picks is unchanged.
- Copy NoDrop and the two reserved words unchanged.
Probability entry construction/sums are outside this supplied-field conversion.

**GDF-R11a. Stat identities** [verified in 1.13c data]. ItemStatCost has 359 data rows; ID equals zero-based row index throughout, and Stat names the row.

| Id (decimal / hex) | Name |
|--------------------|------|
| 0 / 0x00 | strength |
| 2 / 0x02 | dexterity |
| 6 / 0x06 | hitpoints |
| 7 / 0x07 | maxhp |
| 8 / 0x08 | mana |
| 10 / 0x0A | stamina |
| 12 / 0x0C | level |
| 13 / 0x0D | experience |
| 14 / 0x0E | gold |
| 16 / 0x10 | item_armor_percent |
| 19 / 0x13 | tohit |
| 20 / 0x14 | toblock |
| 21 / 0x15 | mindamage |
| 22 / 0x16 | maxdamage |
| 23 / 0x17 | secondary_mindamage |
| 24 / 0x18 | secondary_maxdamage |
| 29 / 0x1D | lastexp |
| 31 / 0x1F | armorclass |
| 32 / 0x20 | armorclass_vs_missile |
| 33 / 0x21 | armorclass_vs_hth |
| 34 / 0x22 | normal_damage_reduction |
| 35 / 0x23 | magic_damage_reduction |
| 36 / 0x24 | damageresist |
| 37 / 0x25 | magicresist |
| 38 / 0x26 | maxmagicresist |
| 39 / 0x27 | fireresist |
| 40 / 0x28 | maxfireresist |
| 41 / 0x29 | lightresist |
| 42 / 0x2A | maxlightresist |
| 43 / 0x2B | coldresist |
| 44 / 0x2C | maxcoldresist |
| 45 / 0x2D | poisonresist |
| 46 / 0x2E | maxpoisonresist |
| 48 / 0x30 | firemindam |
| 49 / 0x31 | firemaxdam |
| 50 / 0x32 | lightmindam |
| 51 / 0x33 | lightmaxdam |
| 52 / 0x34 | magicmindam |
| 53 / 0x35 | magicmaxdam |
| 54 / 0x36 | coldmindam |
| 55 / 0x37 | coldmaxdam |
| 60 / 0x3C | lifedrainmindam |
| 62 / 0x3E | manadrainmindam |
| 64 / 0x40 | stamdrainmindam |
| 67 / 0x43 | velocitypercent |
| 68 / 0x44 | attackrate |
| 69 / 0x45 | other_animrate |
| 74 / 0x4A | hpregen |
| 85 / 0x55 | item_addexperience |
| 93 / 0x5D | item_fasterattackrate |
| 95 / 0x5F | lastblockframe |
| 96 / 0x60 | item_fastermovevelocity |
| 99 / 0x63 | item_fastergethitrate |
| 100 / 0x64 | monster_playercount |
| 102 / 0x66 | item_fasterblockrate |
| 103 / 0x67 | skill_bypass_undead |
| 104 / 0x68 | skill_bypass_demons |
| 105 / 0x69 | item_fastercastrate |
| 106 / 0x6A | skill_bypass_beasts |
| 110 / 0x6E | item_poisonlengthresist |
| 111 / 0x6F | item_normaldamage |
| 115 / 0x73 | item_ignoretargetac |
| 116 / 0x74 | item_fractionaltargetac |
| 118 / 0x76 | item_halffreezeduration |
| 119 / 0x77 | item_tohit_percent |
| 121 / 0x79 | item_demondamage_percent |
| 131 / 0x83 | thorns_percent |
| 141 / 0x8D | item_deadlystrike |
| 142 / 0x8E | item_absorbfire_percent |
| 143 / 0x8F | item_absorbfire |
| 144 / 0x90 | item_absorblight_percent |
| 145 / 0x91 | item_absorblight |
| 146 / 0x92 | item_absorbmagic_percent |
| 147 / 0x93 | item_absorbmagic |
| 148 / 0x94 | item_absorbcold_percent |
| 149 / 0x95 | item_absorbcold |
| 153 / 0x99 | item_cannotbefrozen |
| 164 / 0xA4 | skill_concentration |
| 171 / 0xAB | skill_armor_percent |
| 172 / 0xAC | alignment |
| 179 / 0xB3 | attack_vs_montype |
| 182 / 0xB6 | armor_override_percent |
| 329 / 0x149 | passive_fire_mastery |
| 330 / 0x14A | passive_ltng_mastery |
| 331 / 0x14B | passive_cold_mastery |
| 333 / 0x14D | passive_fire_pierce |
| 334 / 0x14E | passive_ltng_pierce |
| 335 / 0x14F | passive_cold_pierce |
| 336 / 0x150 | passive_pois_pierce |
| 337 / 0x151 | passive_critical_strike |
| 338 / 0x152 | passive_dodge |
| 339 / 0x153 | passive_avoid |
| 340 / 0x154 | passive_evade |
| 348 / 0x15C | passive_weaponblock |
| 352 / 0x160 | last_sent_hp_pct |
| 357 / 0x165 | passive_mag_mastery |

**GDF-R11b. Fixed code roles** [verified in 1.13c code (static)]. The defense reader selects stat 171 skill_armor_percent, not stat 172 alignment. With supplied definition count <= 171 it skips that read and supplies 0; otherwise it selects stat 171, layer 0 through NRF-R12b/c. The combat eligibility check selects state 54 uninterruptable, not stat 54, through NRF-R12a. The block-animation rate path selects state 101 holyshield through NRF-R12a; this owns its selected state identity, not the complete rate calculation.

**GDF-R12. State identities** [verified in 1.13c data]. States has 185 data rows; id equals zero-based row index throughout and state names the row.

| Id (decimal / hex) | Name |
|--------------------|------|
| 1 / 0x01 | freeze |
| 2 / 0x02 | poison |
| 11 / 0x0B | cold |
| 12 / 0x0C | inferno |
| 15 / 0x0F | concentrate |
| 21 / 0x15 | stunned |
| 42 / 0x2A | concentration |
| 47 / 0x2F | sanctuary |
| 54 / 0x36 | uninterruptable |
| 65 / 0x41 | dodge |
| 66 / 0x42 | avoid |
| 68 / 0x44 | evade |
| 86 / 0x56 | justhit |
| 101 / 0x65 | holyshield |
| 104 / 0x68 | corpse_nodraw |
| 105 / 0x69 | alignment |
| 107 / 0x6B | shatter |
| 115 / 0x73 | burning |
| 131 / 0x83 | shrine_resist_fire |
| 133 / 0x85 | shrine_resist_poison |
| 146 / 0x92 | invis |

The code-fixed block-animation state test is specified by GDF-R11b; data identity 101 is holyshield.

**GDF-R13. State-definition fields** [verified in 1.13c code (static)]. State fields: state/`overlay1`..4/castoverlay/removerlay/pgsvoverlay/stat/onsound/offsound/skill/missile (16-bit linked ids); setfunc/remfunc/group/gfxclass/cltevent/clteventfunc/cltactivefunc/srvactivefunc (16-bit); colorpri/colorshift/light-r/light-g/light-b/gfxtype (u8); itemtype (linked 16-bit); itemtrans (linked 8-bit); and boolean nosend/aura/hide/transform/pgsv/active/remhit/damblue/damred/attblue/attred/curse/curable/plrstaydeath/monstaydeath/bossstaydeath/disguise/restrict/blue/armblue/rfblue/rcblue/rlblue/rpblue/stambarblue/armred/rfred/rcred/rlred/rpred/exp/shatter/life/udead/green/nooverlays/noclear/bossinv/meleeonly/notondead. Definition fields are separate from unit state membership.

**GDF-R14a. Player-mode identities** [verified in 1.13c data]. PlrMode has 20 zero-based mode ids:

| Id | Name | Token | Code |
|----|------|-------|------|
| 0 | Death | DT | DT |
| 1 | Neutral | NU | NU |
| 2 | Walk | WL | WL |
| 3 | Run | RN | RN |
| 4 | Get Hit | GH | GH |
| 5 | Town Neutral | TN | TN |
| 6 | Town Walk | TW | TW |
| 7 | Attack1 | A1 | A1 |
| 8 | Attack2 | A2 | A2 |
| 9 | Block | BL | BL |
| 10 | Cast | SC | SC |
| 11 | Throw | TH | TH |
| 12 | Kick | KK | KK |
| 13 | Skill1 | S1 | S1 |
| 14 | Skill2 | S2 | S2 |
| 15 | Skill3 | S3 | S3 |
| 16 | Skill4 | S4 | S4 |
| 17 | Dead | DD | DD |
| 18 | Sequence | GH | SQ |
| 19 | Knock back | GH | KB |

**GDF-R14b. Player-mode preparation** [verified in 1.13c code (static)]. Preparation reads Name (string, limit 31) and Token (four-byte code), not Code. Player-type and player-mode definitions are combined at initialization; the player-mode subset retains its own order.

**GDF-R15a. Monster-mode identities** [verified in 1.13c data]. MonMode has 16 rows:

| Id | Name | Token | Code |
|----|------|-------|------|
| 0 | death | DT | DT |
| 1 | neutral | NU | NU |
| 2 | walk | WL | WL |
| 3 | gethit | GH | GH |
| 4 | `attack1` | A1 | A1 |
| 5 | `attack2` | A2 | A2 |
| 6 | block | BL | BL |
| 7 | cast | SC | SC |
| 8 | `skill1` | S1 | S1 |
| 9 | `skill2` | S2 | S2 |
| 10 | `skill3` | S3 | S3 |
| 11 | `skill4` | S4 | S4 |
| 12 | dead | DD | DD |
| 13 | knockback | GH | KB |
| 14 | sequence | xx | xx |
| 15 | run | RN | RN |

**GDF-R15b. Monster-mode preparation** [verified in 1.13c code (static)]. Preparation reads name (string, limit 31), token (four-byte code), and DT_Dir/NU_Dir/WL_Dir/GH_Dir/A1_Dir/A2_Dir/BL_Dir/SC_Dir/S1_Dir..S4_Dir/DD_Dir/KB_Dir/SQ_Dir/RN_Dir (u8); code is not prepared. Token reuse does not change mode identity.

**GDF-R16a. Item-type identities** [verified in 1.13c data]. ItemTypes data row 4 is Gold, Code gold; row 51 is Any Shield, Code shld. The Expansion separator is text data row 58. Both cited identities precede it.

**GDF-R16b. Item-type fields** [verified in 1.13c code (static)]. Code is a four-byte code also participating in link-building. Equiv1/Equiv2/Shoots/Quiver are linked 16-bit item types; BodyLoc1/BodyLoc2/StaffMods/Class/StorePage linked 8-bit identities; Repair/Body/Throwable/Reload/ReEquip/AutoStack/Magic/Rare/Normal/Charm/Gem/Beltable/MaxSock1/MaxSock25/MaxSock40/TreasureClass/Rarity/CostFormula/VarInvGfx u8; InvGfx1..6 strings with limit 31. Arbitrary custom link-building remains gated by GDF-Q06.

**GDF-R17a. Monster and skill identities** [verified in 1.13c data]. The cited monster class indices below agree with MonStats hcIdx cells and prepared-record correspondence. Id names the text row. The original 734 ordinary text rows follow hcIdx order, excluding the Expansion separator at data row 410. A linked identity stored inside a prepared record is distinct from its collection index; global equality is not established. MonStats2 Id names linked extended records; its count need not equal MonStats count.

| Monster id | Id |
|------------|----|
| 78 | `sandleaper1` |
| 110 | `vulture1` |
| 118 | `willowisp1` |
| 136 | `batdemon1` |
| 230 | firebeast |
| 231 | iceglobe |
| 247 | `frogdemon1` |
| 271 | roguehire |
| 338 | act2hire |
| 359 | act3hire |
| 403 | `trappedsoul1` |
| 560 | act5hire1 |
| 561 | act5hire2 |

Skills has 357 rows, no separator, Id equal to zero-based row index, and skill as name.

**GDF-R17b. Handler-selector identity** [verified in 1.13c code (static)]. The skill eligibility exception tests prepared srvdofunc for 67 or 76, not skill identity. Either match takes the exception branch; otherwise execution continues to the interrupt test. This owns the selected-field comparison only.

**GDF-R18. Named flag-test results** [verified in 1.13c code (static)]. Supplied false flags return 0. True MonStats inTown/NoShldBlock tests return 4; true Missiles GetHit returns 64 and SoftHit 128; true Skills interrupt returns 128. These are original primitive return values, not a requirement for physical record storage. MonStats noRatio/killable are distinct named fields. Consumers own responses to flags.

**GDF-R19a. Prepared table ownership** [verified in 1.13c code (static)]. Initialization publishes named MonStats/MonStats2/Skills/SkillDesc/CharStats/ItemStatCost/Missiles/States/ItemTypes collections and counts. Difficulty/Experience are separate; Experience has no separately published count. Game/unit/path/stat-list/AI/damage records are built by code rather than loaded from these tables. Their shared semantic fields are described in the foundations sections 3.2/3.3; each consuming gameplay contract owns its required fields and operations. NRF-R16a/b consumes this dictionary’s table-field and identity correspondence.

**GDF-R19b. PetType count** [verified in 1.13c code (static)]. The count used by the party experience pet gate belongs to PetType. Its loader rejects count >= 256. Count <= 0 publishes 0; otherwise it publishes the positive count. Why a consumer compares it to 7 belongs to that consumer. Original PetType has 20 text rows; matching PlrMode's count does not change ownership.

**GDF-R20a. Selected load path** [verified in 1.13c code (static)]. With explicitly supplied resource/decode outcomes:
1. If generate-compiled-data is true, request text. Missing text aborts initialization. Prepare zero-initialized records, request publication by the compiled writer, and discard this intermediate collection.
2. Independently select the final form: compiled when read-compiled-data is true, text otherwise.
3. Missing selected final resource aborts; there is no fallback to the other form.
4. Compiled success supplies prepared records/count. Text success prepares zero-initialized records from decoded text rows.
5. Publish final count only when requested.
Decoder/compiler/writer integration is GDF-R20b; supplied-result control flow remains available.

**GDF-R20b. Full preparation integration** [unresolved]. Full resource decoding, name/code linking, locale-sensitive header matching, writer success/failure, expression/callback conversion, and arbitrary compiled correspondence need complete producer contracts. GDF-Q06/Q10/Q12/Q13 gate those subsets, not supplied-result control flow.

### Id map

| Id | Summary | Status label | Gates | Vectors / table record |
|----|---------|--------------|-------|------------------------|
| GDF-R01a | Column schema | verified in 1.13c code (static) | none | GDF-V01, GDF-V10, GDF-V13 |
| GDF-R01b | Direct text-cell conversions | verified in 1.13c code (static) | none | GDF-V01, GDF-V19 |
| GDF-R01d | ASCII header comparison | verified in 1.13c code (static) | none | GDF-V22 |
| GDF-R01c | Other text conversions | unresolved | GDF-Q06 / GDF-Q12: full text conversion | none |
| GDF-R02 | Difficulty fields | verified in 1.13c code (static) | none | GDF-V01, GDF-V10 |
| GDF-R03a | Player-class fields and getter | verified in 1.13c code (static) | none | GDF-V12 |
| GDF-R03b | Class identities | verified in 1.13c data | none | TD-CLASS |
| GDF-R03c | Class display-name preparation | unresolved | GDF-Q13: localized display result | none |
| GDF-R04 | Experience fields | verified in 1.13c code (static) | none | GDF-V13 |
| GDF-R05a | Monster fields and getter | verified in 1.13c code (static) | none | GDF-V02, GDF-V08, GDF-V17 |
| GDF-R05b | Derived monster animation values | unresolved | GDF-Q01: complete rate production | none |
| GDF-R06a | Extended monster fields | verified in 1.13c code (static) | none | GDF-V09 |
| GDF-R06b | Mode-availability getter | verified in 1.13c code (static) | none | GDF-V09 |
| GDF-R07 | Missile fields | verified in 1.13c code (static) | none | GDF-V14 |
| GDF-R08a | Skill fields and getter | verified in 1.13c code (static) | none | GDF-V06, GDF-V18 |
| GDF-R08b | Parameter calculation | verified in 1.13c code (static) | none | GDF-V06 |
| GDF-R09 | Stat-definition fields | verified in 1.13c code (static) | none | GDF-V11 |
| GDF-R10a | Treasure source fields | verified in 1.13c code (static) | none | GDF-V05 |
| GDF-R10b | Selected treasure conversion | verified in 1.13c code (static) | none | GDF-V05 |
| GDF-R11a | Stat identities | verified in 1.13c data | none | TD-STAT |
| GDF-R11b | Fixed code roles | verified in 1.13c code (static) | none | GDF-V11 |
| GDF-R12 | State identities | verified in 1.13c data | none | TD-STATE |
| GDF-R13 | State-definition fields | verified in 1.13c code (static) | none | GDF-V15 |
| GDF-R14a | Player-mode identities | verified in 1.13c data | none | TD-PLAYER-MODE |
| GDF-R14b | Player-mode preparation | verified in 1.13c code (static) | none | GDF-V04, GDF-V10 |
| GDF-R15a | Monster-mode identities | verified in 1.13c data | none | TD-MONSTER-MODE |
| GDF-R15b | Monster-mode preparation | verified in 1.13c code (static) | GDF-Q12: absent direction-column defaults; supplied fields remain available | GDF-V04, GDF-V10 |
| GDF-R16a | Item-type identities | verified in 1.13c data | none | TD-ITEMTYPE |
| GDF-R16b | Item-type fields | verified in 1.13c code (static) | GDF-Q06: custom linked values; field identities remain available | GDF-V16 |
| GDF-R17a | Monster and skill identities | verified in 1.13c data | none | TD-MONSTER-SKILL |
| GDF-R17b | Handler-selector identity | verified in 1.13c code (static) | none | GDF-V07 |
| GDF-R18 | Named flag-test results | verified in 1.13c code (static) | none | GDF-V02, GDF-V14, GDF-V18 |
| GDF-R19a | Prepared table ownership | verified in 1.13c code (static) | none | GDF-V08, GDF-V10, GDF-V17 |
| GDF-R19b | PetType count | verified in 1.13c code (static) | none | GDF-V20 |
| GDF-R20a | Selected load path | verified in 1.13c code (static) | none | GDF-V21 |
| GDF-R20b | Full preparation integration | unresolved | GDF-Q06 / GDF-Q10 / GDF-Q12 / GDF-Q13: full preparation | none |

## 5. Outputs and side effects

| Output / effect | Rule |
|-----------------|------|
| named fields and direct normalized cells | GDF-R01a/b and field rules |
| header equality | GDF-R01d |
| published logical collections / requested counts | GDF-R19a, GDF-R20a |
| Difficulty initialization abort when prepared count is not 3 | GDF-R02 |
| published supplied Experience records, without a separately stored count | GDF-R04 |
| selected logical record or no record; named field values | GDF-R03a, GDF-R05a, GDF-R08a |
| availability mask / 0 | GDF-R06b |
| parameter result / 0 | GDF-R08b |
| converted treasure fields | GDF-R10b |
| identity maps | GDF-R03b, GDF-R11a, GDF-R12, GDF-R14a, GDF-R15a, GDF-R16a, GDF-R17a |
| selected stat/state identity / exception branch / flag result | GDF-R11b, GDF-R17b, GDF-R18 |
| PetType count / initialization abort | GDF-R19b |
| requested load form / generation / initialization abort | GDF-R20a |
| no RNG, gameplay stat/state/event/unit writes, or messages | all local rules |

## 6. Branches and edge cases

- Supported difficulty is 0..2; no substitute is defined outside that domain.
- No-record results are visible to the caller; its response belongs to its gameplay contract.
- A token can be reused by another mode: player Knock back and Sequence use GH despite differing Code cells.
- A prepared field may be derived by code rather than a text column.
- Nonzero flag cells include 2 and negative values, not only 1.
- Zero treasure Picks becomes 1; negative Picks remains negative.
- Text identities and compiled identities require separate corroboration. A matching count alone proves no ordering. The global MonStats compiled-Id correspondence is unresolved even though the cited identities match.

## 7. Open questions

| Id | Kind | Blocked assertions / boundary | Missing evidence |
|----|------|-------------------------------|------------------|
| GDF-Q01 | integration | Complete derived monster animation-rate production; supplied-value consumers remain separate | base animation-data selection contracts used by the identified writer |
| GDF-Q06 | behavior | Full name/code links, expression/callback conversions, descriptor-specific preparation; direct conversions remain available | complete conversion handlers and their consumers |
| GDF-Q10 | integration | Full treasure probability/group generation and global compiled-row equivalence; supplied treasure conversion remains available | probability producers, generated-group lifecycle, explanation of compiled/text mismatch |
| GDF-Q11 | observation | Gameplay/runtime-confirmation claims only | attributable executions of verified behavioral rules; table/memory reads are correspondence evidence |
| GDF-Q12 | behavior | Missing-column defaults and complete name-link/callback cell behavior | final missing-field switch and indirect conversion contracts |
| GDF-Q13 | integration | Localized class display-name result | imported localization/string behavior used by the identified writer |

Attempts and what closes each gap are in [Open evidence gaps](../plans/gaps/core-game-data-fields.md). Observation summaries do not gate supplied-input results.

## 8. Test vectors

### 8.1 Executable vectors

Every vector uses no draw and leaves RNG state unchanged. Supplied logical records are fixtures. Data-only checks identify named table cells rather than claim gameplay execution. Record reads corroborate only named fields; they do not observe execution of a getter or flag-test call.

**GDF-V01** — GDF-R01a/b, GDF-R02; original-profile difficulty field preparation.
Inputs: the 22 GDF-R02 i32 columns; rows Normal/Nightmare/Hell at indices 0/1/2; prepared count 3; cells Hell.ResistPenalty = "-100", Hell.UberCodeOddsNormal = "20", Hell.UltraCodeOddsNormal = "30", Hell.MonsterFreezeDivisor = "4", Nightmare.MonsterColdDivisor = "2", Hell.LifeStealDivisor = "3", Normal.ManaStealDivisor = "1", Hell.HireableBossDamagePercent = "25". Trace: parse each selected cell as i32 and bind by column. Exact outputs: the corresponding fields are -100, 20, 30, 4, 2, 3, 1, 25. All are independent of header order. Fresh supplied reordering fixture: retain count 3 and the named rows; provide all 22 headers, with UberCodeOddsNormal first, ResistPenalty second, then the remaining 20 GDF-R02 headers in their listed order. Every fixture cell is "0" except Hell.UberCodeOddsNormal "20" and Hell.ResistPenalty "-100". Exact Hell outputs remain UberCodeOddsNormal 20 and ResistPenalty -100; its other 20 fields are 0. The other two rows have all fields 0. This fixture tests name binding, not original values of those zero cells. runtime_observed: no. Named compiled correspondence is in the companion; it does not execute this preparation path.

**GDF-V02** — GDF-R05a, GDF-R18; supplied named monster flags.
Inputs: act5hire1 has inTown true, NoShldBlock false; `sandleaper1` has both false. Trace: select named flags and their return masks. Exact outputs: act5hire1 town 4 / shield 0; `sandleaper1` town 0 / shield 0; booleans true/false/false/false respectively. A separate supplied record has inTown false and NoShldBlock true; exact town/shield outputs are 0/4 and boolean predicates false/true. This distinguishes the true NoShldBlock return mask from 1 or a different flag. runtime_observed: no. Named-input correspondence is recorded in the companion; it does not observe the flag-test calls.

**GDF-V03** — data-only GDF-R11a / GDF-R12 check.
Inputs: ItemStatCost ID 99 / 171, States id 54 / 146. Select by named id and name column. Exact outputs: 99 item_fastergethitrate; 171 skill_armor_percent; state 54 uninterruptable; state 146 invis. No gameplay observation; named table records provide evidence.

**GDF-V04** — GDF-R14a/b and GDF-R15a/b; identity and selected-column check.
Inputs: player modes 19, 18, 3; monster modes 13, 14, 3. Exact outputs as name/token/code: player 19 Knock back/GH/KB, 18 Sequence/GH/SQ, 3 Run/RN/RN; monster 13 knockback/GH/KB, 14 sequence/xx/xx, 3 gethit/GH/GH. Preparation fixture: supply player-mode 19 Name 'Knock back', Token 'GH', Code 'KB' and monster-mode 13 name 'knockback', token 'GH', code 'KB', each direction cell '0'. Exact prepared player fields are Name 'Knock back'/Token 'GH  '; prepared monster fields name 'knockback'/token 'GH  ' and all 16 direction values 0; Code/code is not a prepared field. No gameplay observation; mode-change observation is GDF-O01.

**GDF-V05** — GDF-R10a/b; supplied treasure-field conversion.
Inputs: Picks 0, NoDrop 15, group 0, level 7, Magic 1024, Rare/Set/Unique 983 each; generated group count 5; source reserved words 0/0; initialized destination group 77, all other converted fields 0. Trace: zero Picks becomes 1, zero group leaves 77, copy remaining fields. Exact outputs: group 77, level 7, Picks 1, NoDrop 15, Magic/Rare/Set/Unique 1024/983/983/983, reserved 0/0. Fresh variant Picks 3 gives Picks 3; fresh variant Picks -2 gives -2. Fresh nonzero-group cases keep the other inputs unchanged: group 10 plus generated count 5 yields destination group 15; group 65535 plus count 5 yields low16(65540) = 4. Fresh copy discriminator retains the other inputs, supplies source reserved words 17/29 and initialized destination reserved words 91/92: exact destination reserved words become 17/29, not 91/92 or a swapped pair. These are supplied words, not new table-column or cell claims. A separate quality-copy fixture retains the original inputs and initialized destination, supplying source Magic 101, Rare 202, Set 303 and Unique 404. Copy each quality word to its same-named destination. Exact destination Magic/Rare/Set/Unique values are 101/202/303/404; the remaining outputs are group 77, level 7, Picks 1, NoDrop 15 and reserved words 0/0. All four quality values differ, so any swapped pair produces a different result. These are synthetic supplied words, not original-table cell claims. runtime_observed: no.

**GDF-V06** — GDF-R08a/b; supplied skill parameter operation.
Inputs: skill index 271, count 357, supplied record Param3 20 / Param4 5, level 4. Trace: valid index and positive level; 20 + (4 - 1) * 5 = 35. Exact output 35. Fresh level 0 or -1 gives 0; fresh index 357 gives no record and parameter result 0. For the fresh negative-index case, also supply entry 0 with Param3 9 / Param4 2; index -1 still gives no record and 0, rather than using entry 0's result 15 at level 4. Fresh valid-index overflow case: index 271/count 357, level 3, Param3 2147483640, Param4 10; i32(2147483640 + 2 * 10) = -2147483636, with no clamp. Fresh absent-record case supplies index 0/count 1, level 1, and no record at the selected identity: parameter result 0, with no parameter-field read. runtime_observed: no.

**GDF-V07** — GDF-R17b; supplied handler selector branch.
Inputs: three records with (skill identity, srvdofunc) = (12,67), (13,76), (67,68). Trace: compare srvdofunc, not skill identity. Exact outputs: exception taken, exception taken, exception not taken / continue to interrupt test. runtime_observed: no.

**GDF-V08** — GDF-R05a, GDF-R19a; supplied monster record selection.
Inputs: MonStats collection count 734, entry 78 = record R78 whose identity is `sandleaper1`; query 78, then 734, then -1. Other entries are opaque and not accessed. Trace: 78 is in range; 734/-1 are not. Exact outputs: R78, no record, no record. Separate supplied fixture: collection count 3; all entries present. Entry 0 has stored Id 0 and Velocity 0; entry 1 is record A with stored Id 2 and Velocity 11; entry 2 is record B with stored Id 1 and Velocity 22. Query index 1. Trace: 1 is in range; select collection entry 1 without searching its stored Id. Exact output: record A with Id 2 and Velocity 11, not record B with Id 1 and Velocity 22. These stored identities and values are synthetic supplied fields, not original-table cell claims. runtime_observed: no. Named-record identity correspondence is in the evidence companion.

**GDF-V09** — GDF-R06a/b; supplied mode availability.
Inputs: monster index 78, MonStats count 734, that record's MonStatsEx 78; MonStats2 count 609 and supplied linked record availability mKB true / mGH true / mA1 false, every other flag false; modes 13, 3, 4. Trace: both indices valid; true masks are 1 << 5 and 1 << 3. Exact outputs: 32, 8, 0. Fresh asymmetric supplied flags mKB=true/mGH=false give mode 13 result 32 and mode 3 result 0; mKB=false/mGH=true give mode 13 result 0 and mode 3 result 8. All other flags remain false. Fresh invalid MonStatsEx -1 or 609 gives 0. Fresh in-range absent-record case: monster index 0/count 1, MonStatsEx 0, MonStats2 count 1, no extended record at that selected identity; querying mode 13 returns 0 without a flag read. Fresh invalid monster indices -1 and 734, querying mode 13, each return 0 without reading a MonStats record or its link. Separate supplied link-selection fixture: MonStats count 3 with all three records present; query monster index 1 whose record has MonStatsEx 2. The other MonStats records have MonStatsEx 0. MonStats2 count 3 with all three records present; entry 0 has every availability flag false, entry 1 has every flag false including mGH, and entry 2 has mGH true with every other flag false. Query mode 3. Trace: select monster entry 1, then its linked extended entry 2; mGH is true, yielding 1 << 3. Exact output 8. Using the monster index 1 as the extended index would incorrectly return 0. This is supplied-record linkage, not a new compiler correspondence claim. runtime_observed: no.

**GDF-V10** — GDF-R01a, GDF-R02, GDF-R19a; supplied initialization schema.
Inputs: Difficulty schema has the 22 GDF-R02 columns followed by terminator, prepared count 3, publish-count requested; Experience schema has its eight GDF-R04 columns followed by terminator, publish-count not requested. Trace: stop each schema at terminator, publish named collections, condition count publication. Exact outputs: Difficulty 22 fields and published count 3; Experience 8 fields and no separately published count. Fresh Difficulty prepared count 2 aborts initialization. Fresh schema-termination fixture: descriptors in order are (name end, kind i32), (name after, kind i32), (name arbitrary-stop-name, kind Stop), (name must-not-be-read, kind i32). Exact selected field names are end and after, count 2; the final descriptor is not visited. Termination therefore depends on conversion kind, not the name end. runtime_observed: no.

**GDF-V11** — GDF-R09, GDF-R11b; stat-definition selection boundary.
Inputs: requested fixed stat 171, layer 0, definition count 359. Trace: 359 > 171. Exact output: select ItemStatCost identity 171 skill_armor_percent and stat read (171,0); this vector does not execute NRF-R12b/c. Fresh count 171 or less skips read and supplies 0. The two state-call identity cases supply unit U and a state collection (state 54 present, state 101 absent): exact selected calls are state-presence(U,54) and state-presence(U,101); the nonzero predicates of consumed NRF-R12a results are true/false. The original state-test returns are 4194304 for state 54 (1 << 22) and 0 for absent state 101; no boolean return of 1 is implied. U is a player with a supplied extended stat list, state membership support and state count 185, within the stated NRF-R12a eligibility. runtime_observed: no.

**GDF-V12** — GDF-R03a/b; supplied class record getter.
Inputs: class index 1, count 7, supplied row Sorceress with BlockFactor 20, ManaRegen 120, ToHitFactor -15. Trace: 1 in range; BlockFactor unsigned. Exact outputs: Sorceress record; 20, 120, -15. Fresh indices 7 and -1 give no record; no dereference is required by this boundary. Fresh supplied Sorceress fixture keeps the valid index/count and other fields but sets BlockFactor 200; the field read returns 200, not signed -56. This is a fixture, not a new original-table cell claim. runtime_observed: no. Compiled correspondence for the named fields is in the companion.

**GDF-V13** — GDF-R01a, GDF-R04; schema field identity.
Inputs: Experience schema with its eight named i32 columns, class index 3, row 10. Trace: class 3 is Paladin. Exact selected field identities: (Experience,row 10,Paladin) and (Experience,row 10,ExpRatio). The identity-only case asserts no cell value. Fresh supplied header-order fixture: headers ExpRatio, Assassin, Druid, Barbarian, Paladin, Necromancer, Sorceress, Amazon; row 10 cells "101", "202", "303", "404", "505", "606", "707", "808" in that order. Supply rows 0..9 with all eight cells "0". Exact selected row-10 outputs remain Paladin 505 and ExpRatio 101, not the fourth and eighth input positions. These are fixture values, not new original-table claims. runtime_observed: no.

**GDF-V14** — GDF-R07, GDF-R18; supplied missile flags.
Inputs: record A GetHit true, SoftHit true, ClientSend false; record B GetHit false, SoftHit false, ClientSend true; all other flags false. Trace: GetHit/SoftHit are distinct from ClientSend. Exact test results: A 64/128, B 0/0. runtime_observed: no.

**GDF-V15** — GDF-R13; schema field identity.
Inputs: States identity 54 and the GDF-R13 schema. Trace: select the state definition independently from unit state membership. Exact selected identities: (States,54,stat), (States,54,skill), (States,54,curse), (States,54,life). curse and life are independent boolean fields. No cell values or active state membership are asserted. runtime_observed: no.

**GDF-V16** — GDF-R16a/b; original identity and Code-field check.
Inputs: ItemTypes row identities 51 and 4, both before the separator. Exact outputs: 51 Any Shield / Code shld; 4 Gold / Code gold. The selected Code field is a four-byte code: supplied Code 'gold'/'shld' remains 'gold'/'shld'; full link-building is gated separately. No gameplay observation.

**GDF-V17** — GDF-R05a, GDF-R17a, GDF-R19a; supplied original-profile monster identity.
Inputs: MonStats count 734, index 560, supplied entry act5hire1 with identity 560, MonStatsEx identity 560, Velocity 8, Run 14. Trace: 560 in range. Exact output: that act5hire1 record; identity 560, linked identity 560, Velocity 8, Run 14. runtime_observed: no. Named text cells and compiled correspondence corroborate these inputs. The different collection-index/stored-identity discriminator is GDF-V08; this identity-correspondence case does not duplicate that fixture. For independent supplied difficulty-column selection, retain this valid selected record and supply Drain 7, Drain(N) 13, Drain(H) 29. Difficulty 0, 1 and 2 select base, Nightmare and Hell respectively; exact selected unsigned field values are 7, 13 and 29. This asserts field selection, not the leech calculation or new original-table values.

**GDF-V18** — GDF-R08a, GDF-R18; supplied skill flags.
Inputs: record A interrupt true / leftskill false; record B interrupt false / leftskill true; other flags false. Trace: select interrupt only. Exact outputs: A 128 / interrupt set, B 0 / interrupt clear. runtime_observed: no.

**GDF-V19** — GDF-R01b; supplied direct conversion discrimination.
Inputs: selected fields i32 "-100", u16 "65537", u8 "-1", flag "2", flag "-1", flag "", four-code "SL", bounded string "abcdef" with byte limit 3; initial other flag true. Trace: wrapping accumulation, low-bit store, nonzero test, space padding, length truncation. Exact outputs: -100, 1, 255, true, true, false, "SL  ", "abc"; the other flag stays true. Fresh empty i32 gives 0, four-code "" gives four spaces, and four-code "ABCDE" yields "ABCD". Fresh selected i32 cell "2147483648" yields -2147483648; cell "-2147483649" yields 2147483647. These cases distinguish wrapping accumulation from saturation. A separate supplied bounded-string cell contains the two raw bytes [0xC3, 0xA9] and has byte limit 1. Trace: retain the first byte only. Exact retained bytes [0xC3], rather than retaining both bytes as one Unicode character. This defines neither a text encoding nor locale behavior. runtime_observed: no.

**GDF-V20** — GDF-R19b; supplied PetType count.
Inputs: fresh decoded counts 20, 0, -1, 255, 256. Trace: reject >=256; nonpositive becomes 0. Exact outputs: published 20, published 0, published 0, published 255, initialization abort. The original count 20 is corroborated by PetType text and compiled count records. runtime_observed: no.

**GDF-V21** — GDF-R20a; supplied load-path outcomes.
Inputs: final collection F has count 3; intermediate I count 3; I and F are distinct supplied decoded collections; all schema fields present; no conversion side effects; requested count true. Fresh cases:
- generate false / read compiled true; selected compiled success F: requests compiled only; publishes F/count 3.
- generate false / read compiled true; compiled missing, text available: requests compiled only; aborts, no publication.
- generate false / read compiled false; text success F: requests text only; publishes F/count 3.
- generate true / read compiled true; generation text missing, compiled available: requests generation text, then aborts. No decoder, writer, final compiled read or publication is requested.
- generate false / read compiled false; final text missing, compiled available: requests text only, then aborts. No decoder, compiled fallback or publication is requested.
- generate true / read compiled true; text decode succeeds I, writer success, compiled decode succeeds F: exact ordered log is text-decode(I), writer(I), discard(I), compiled-decode(F), publish(F/count 3).
- generate true / read compiled false; first text decode succeeds I, writer success, second text decode succeeds distinct F: exact ordered log is text-decode(I), writer(I), discard(I), second-text-decode(F), publish(F/count 3). It publishes F, not the reused intermediate I.
Trace follows GDF-R20a independently of file availability in the unused form. runtime_observed: no. These supplied decoded outcomes discriminate resource selection and publication order; they do not execute or verify original pre-conversion record zeroing. Full decoding/writer execution is gated separately.

**GDF-V22** — GDF-R01d; supplied ASCII comparison path.
Inputs: header names Stat and Value; descriptor names stat and value; a negative comparison Stat versus state; ASCII comparison mode. Trace: case-fold letters, compare full strings. Exact outputs: equal, equal, not equal. This tests the original comparison path, not duplicate/missing-column defaults or another locale. runtime_observed: no.

### 8.2 Observation vectors

- **GDF-O01**: During knockback, observe player mode 19 / monster mode 13. Pending; data mapping alone does not observe mode changes. [Procedure](../research/spec-evidence/core-game-data-fields.md#observation-procedures).
- **GDF-O02**: Corroborate derived animation values and identified writer for additional classes. Existing values are correspondence only; complete production remains GDF-Q01. [Procedure](../research/spec-evidence/core-game-data-fields.md#observation-procedures).
- **GDF-O03**: Corroborate PetType count production at initialization. Count 20 is a correspondence read, not execution of its producer. [Procedure](../research/spec-evidence/core-game-data-fields.md#observation-procedures).

## 9. Acceptance and dependencies

Accept data-only identities for the original profile and their named table records. Accept verified supplied-record/schema/cell/control-flow assertions demonstrated by GDF-V01..GDF-V22 separately from gated producer integration. No assertion requires the original memory layout.

Full reproduction also requires all in-scope gates resolved and attributable observations of verified behavioral rules where possible. Table reading does not satisfy gameplay observation. This draft does not imply user approval.

**Consumed operations:**

| Dependency | Supplied inputs | Result consumed here | Evidence / acceptance boundary |
|---|---|---|---|
| NRF-R12a state test | unit type and list, state count and selected state 54 or 101 | state mask or 0; nonzero is presence | verified in 1.13c code (static); read-only test, no state-set/update effects consumed |
| NRF-R12b / NRF-R12c stat read | selected stat 171, layer 0, definition count/record and existing ordered stat arrays | i32 stat value after the stated conditional minimum | verified in 1.13c code (static); supplied-array behavior, without claiming stat-array production |

**Supporting conventions:** [foundations notation](core-numeric-and-rng-foundations.md#1-scope) supplies integer wrapping; its sections 3.2/3.3 describe shared semantic records. NRF-R16a/b consumes this dictionary rather than producing its records. No RNG helper is called by this boundary.

**Expression boundary:** GDF-R08a identifies the named 32-bit compiled references. The expression owner, SCX-R01..SCX-R11 and SCX-R13..SCX-R15, defines evaluation/pool selection; SCX-R12 leaves compiler-produced output properties gated. No expression evaluation or compiler result is asserted by GDF-R01b or GDF-V19. GDF-R01c/GDF-R20b retain conversion integration gates.

**Reverse consumers:** gameplay specs citing shared fields and identities, including mitigation/pipeline, hit recovery, item events, knockback, leech, PvP, shield/block, monster scaling, party experience, Assassin damage/sentries, leveled skill damage, and movement/freeze. Citation relationships do not themselves impose a plan order.

## 10. Related specs

- [Numeric and RNG foundations](core-numeric-and-rng-foundations.md): shared numeric/stat/state/record contracts.
- [Skill calc expressions and synergies](skill-calc-expressions-and-synergies.md): expression contract.
- [Damage mitigation](combat-damage-mitigation.md), [Damage pipeline](combat-damage-pipeline-order.md), [Hit recovery](combat-hit-recovery.md), [Knockback](combat-knockback.md), [Party experience and loot sharing](player-party-experience-and-loot-sharing.md): consumers.

## Appendix A. Provenance

Evidence target: Lord of Destruction Expansion 1.13c, Windows x86. Original `D2Common.dll`,
`D2Game.dll`, `Fog.dll` and `Storm.dll` analysis specimens are identified by exact SHA-256
fingerprints in the [source manifest](../research/source-manifest.md).

The [public evidence companion](../research/spec-evidence/core-game-data-fields.md) provides
module-relative function locations, export ordinals where recorded, descriptor/consumer
mappings, rule-linked rationale, named table cells and observation procedures. Static readings
were recorded on 2026-10-12/13 and targeted rereadings on 2026-10-14 using disassembly,
reconstructions checked against original bytes, and independent instruction listings where
recorded. Function descriptions are analyst labels.

The table records dated 2026-10-14 distinguish effective text cells, selected compiled values
and loaded-memory correspondence; effective-file hashes were not captured for those reads.
The [2026-10-15 named-cell corroboration](../research/spec-evidence/core-game-data-fields.md#named-cell-corroboration-2026-10-15) adds effective text-file hashes and stated cell coverage only.
Supplying MPQs, complete override inventories and fresh loaded-DLL/compiled correspondence remain uncaptured; the manifest and companion state the profile limits.
Table/memory correspondence does not establish gameplay execution. Observations remain pending
and the spec remains draft. Public verification uses the identified original game artifacts
and recorded data scope without requiring the private reconstruction project.
