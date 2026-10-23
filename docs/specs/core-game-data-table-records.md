# Game Data Table Records

## Description

Before a game starts, the game reads its data tables (monsters, skills, missiles, character classes, difficulty settings, treasure classes and more) and prepares them as records that the rest of the game looks up. This reference describes how those tables are read and converted, which fields each record carries, and how other systems fetch a record or test one of its flags. It has no visible effect on its own; it makes every system see the same values from the data files.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [Numeric and RNG foundations](core-numeric-and-rng-foundations.md); names: [Game data fields](core-game-data-fields.md)
Ids: GDA-R, GDA-Q, GDA-V, GDA-O; policy in [README conventions](README.md#conventions)
Evidence: [source, table records, and observation history](../research/spec-evidence/core-game-data-table-records.md)
Open evidence gaps: [remaining assertions](../plans/gaps/core-game-data-table-records.md)

## 1. Scope

This contract defines the original profile's table columns, field widths, selected record getters, flag tests, and the bounded table-loading and treasure-field conversion behavior stated below. It defines logical records and results. Identity mappings and names (classes, stats, states, modes, item types, monsters, skills) belong to [Game data fields](core-game-data-fields.md). Gameplay formulas using these fields belong to their consuming specs. Shared stat/state operations belong to RLI-R02 and RLI-R03. Expression evaluation belongs to [Skill calc expressions](skill-calc-expressions.md).

Full text compilation remains gated where lookup, callback, locale, or missing-column behavior is unresolved. A supplied-record getter can be accepted without claiming reproduction of that compiler.

## 2. Trigger and ownership

- D2Common prepares shared game definitions during process initialization, before a game is created. These are initialization operations, not frame updates.
- D2Game and D2Common consumers request records or fields when their gameplay paths need them. The getters stated here are read only.
- Difficulty loading requires three logical rows. Treasure-class loading converts selected source fields to runtime fields after definitions are read.
- Client and server use shared definitions. This contract defines no message.

## 3. Inputs and data

Every table path below is under data/global/excel/ in the original 1.13c archive set. Columns present in the effective files use their text-header spellings. Descriptor-only columns use the original schema's names; their missing-column behavior remains gated.

| Input | Semantic source | Width / format | Meaning |
|-------|-----------------|----------------|---------|
| schema | named table, column, conversion kind, string limit, lookup or callback binding | logical descriptors | selects fields prepared for a table |
| text table | effective named .txt table | 8-bit byte strings, tab-separated cells and CRLF rows | source identities and values; string/code limits count bytes |
| prepared table | named ordered record collection and record count | logical records; count i32 | supplied input to getters |
| table index | caller's class, skill, state, monster, or item-type identity (GDF identities) | i32 unless a field width is stated | selects a row |
| difficulty | game difficulty | u8; 0 Normal, 1 Nightmare, 2 Hell | selects difficulty columns |
| skill level | caller-supplied level | i32 | GDA-R11b parameter calculation |
| monster mode | caller-supplied original monster-mode identity | i32, domain 0..15 | selects GDA-R09b availability; identity mapping is GDF-R06a |
| flags | named fields of supplied logical records | boolean fields | selected flag-test results |
| generation mode | generate compiled data; read compiled data | independent boolean conditions | GDA-R01 selection |
| load/decode result | selected resource availability and decoded records/count | success/failure, logical collection | supplied result at GDA-R01 boundary |
| treasure source and destination | named source fields; destination initialized fields | u16 group/level/quality fields; i32 Picks/NoDrop | GDA-R15b conversion |
| generated treasure group count | item-type treasure groups generated before this conversion | u16 original counter, supplied as i32 | group adjustment |
| reserved treasure words | supplied unnamed source fields | two u16 values | GDA-R15b copies |
| ASCII comparison mode | original ASCII comparison branch | boolean selection | GDA-R03c boundary |
| optional published count | initialization caller choice | boolean request | controls whether a count is published |

| Mapping / constant | Origin | Binding |
|--------------------|--------|---------|
| column names and their tables | named table columns | data supplies cells; the schema binding a column name to a field is fixed by original code |
| handler selectors such as srvdofunc/`itemeventfunc1` | named table columns | data selects engine-defined handlers; a selector is not a skill/stat id |
| conversion kinds, field widths, count 3, bounds, return masks | engine code | fixed by original code |
| custom identity remapping | unresolved | no support or behavior is established here |

[Table records](../research/spec-evidence/core-game-data-table-records.md#table-verification-2026-10-14) distinguish text cells from compiled correspondence.

## 4. Rules

All operations below draw no RNG.

**GDA-R01. Selected load path** [verified in 1.13c code (static)]. With explicitly supplied resource/decode outcomes:
1. If generate-compiled-data is true, request text. Missing text aborts initialization. Prepare zero-initialized records, request publication by the compiled writer, and discard this intermediate collection.
2. Independently select the final form: compiled when read-compiled-data is true, text otherwise.
3. Missing selected final resource aborts; there is no fallback to the other form.
4. Compiled success supplies prepared records/count. Text success prepares zero-initialized records from decoded text rows.
5. Publish final count only when requested.
Decoder/compiler/writer integration is GDA-R02; supplied-result control flow remains available.

**GDA-R02. Full preparation integration** [unresolved]. Full resource decoding, name/code linking, locale-sensitive header matching, writer success/failure, expression/callback conversion, and arbitrary compiled correspondence need complete producer contracts. GDA-Q01/GDA-Q03/GDA-Q02/GDA-Q04 gate those subsets, not supplied-result control flow.

**GDA-R03a. Column schema** [verified in 1.13c code (static)]. A table's schema identifies named columns and conversion kinds. Unique, present text headers bind to descriptors by name regardless of column order; missing and duplicate headers remain GDA-Q02. A descriptor with conversion kind Stop ends the schema, regardless of its displayed name. A descriptor named end with another conversion kind does not terminate it. Initialization publishes prepared records; it publishes their count only when requested. A schema proves field identity, not gameplay meaning.

**GDA-R03b. Direct text-cell conversions** [verified in 1.13c code (static)]. At an already selected schema-field/cell boundary:
- Text characters and string/code limits here mean original 8-bit bytes, not Unicode code points. Numeric fixtures use ASCII decimal bytes; string/code copying retains the supplied byte order.
- Decimal cells accumulate value = i32(value * 10 + character - '0'); a leading minus negates the completed i32 value. Empty numeric cells yield 0. This contract covers empty or well-formed decimal text.
- Numeric stores retain the low 32, 16, or 8 bits according to the selected field. Signedness of a later read belongs to its consumer.
- A flag is false for numeric zero and true for any nonzero parsed number; other named flags are unchanged.
- A direct four-byte code takes at most four text characters, retains their order, and pads missing characters with spaces. Empty text yields four spaces.
- A bounded string retains at most its descriptor's byte limit.
These conversions do not establish header selection, name lookup, callback execution, or missing-column defaults.

**GDA-R03c. ASCII header comparison** [verified in 1.13c code (static)]. On the original ASCII comparison path, header names compare without distinguishing A..Z from a..z. This permits the descriptor name stat to match the header Stat. Duplicate and missing-header defaults remain GDA-Q02. Locale-dependent non-ASCII comparison is outside this ASCII operation.

**GDA-R03d. Other text conversions** [unresolved]. Name/code links, expressions, callback fields, descriptor-specific callback outputs, and missing schema columns must match the original compiler. GDA-Q01 / GDA-Q02 gate their complete contracts; GDA-R03b supplied-cell conversions remain available.

**GDA-R04a. Prepared table ownership** [verified in 1.13c code (static)]. Initialization publishes named MonStats/MonStats2/Skills/SkillDesc/CharStats/ItemStatCost/Missiles/States collections and counts. ItemTypes publication is the separately labeled GDA-R14 assertion. Difficulty/Experience are separate; Experience has no separately published count. Game/unit/path/stat-list/AI/damage records are built by code rather than loaded from these tables. Their shared semantic fields are described in the foundations sections 3.2/3.3; each consuming gameplay contract owns its required fields and operations. RLI-R01a/RLI-R01b consumes this dictionary’s table-field and identity correspondence.

**GDA-R04b. PetType count** [verified in 1.13c code (static)]. The count used by the party experience pet gate belongs to PetType. Its loader rejects count >= 256. Count <= 0 publishes 0; otherwise it publishes the positive count. Why a consumer compares it to 7 belongs to that consumer. Original PetType has 20 text rows; matching PlrMode's count does not change ownership.

**GDA-R05. Difficulty fields** [verified in 1.13c code (static)]. Rows Normal, Nightmare, Hell correspond to difficulty 0, 1, 2. The fields below are i32. A prepared count other than 3 aborts this initialization.

| Columns |
|---------|
| ResistPenalty, DeathExpPenalty, UberCodeOddsNormal, UberCodeOddsGood |
| MonsterSkillBonus, MonsterFreezeDivisor, MonsterColdDivisor, AiCurseDivisor |
| UltraCodeOddsNormal, UltraCodeOddsGood, LifeStealDivisor, ManaStealDivisor |
| UniqueDamageBonus, ChampionDamageBonus, HireableBossDamagePercent |
| MonsterCEDamagePercent, StaticFieldMin, GambleRare, GambleSet, GambleUnique, GambleUber, GambleUltra |

HireableBossDamagePercent does not itself prove its consumer's boss predicate.

**GDA-R06a. Player-class fields and getter** [verified in 1.13c code (static)]. Select the supplied CharStats collection by class index. Negative indices and indices at or above count yield no record. BlockFactor is read as u8.

| Columns | Logical format |
|---------|----------------|
| class | string, byte limit 15 |
| ToHitFactor | i32 |
| str, dex, int, vit, stamina, hpadd, ManaRegen, WalkVelocity, RunVelocity, RunDrain | u8 |
| LifePerLevel, StaminaPerLevel, ManaPerLevel, LifePerVitality, StaminaPerVitality, ManaPerMagic, BlockFactor, StatPerLevel | u8 |
| StartSkill | linked skill identity, 16-bit |

**GDA-R06b. Class display-name preparation** [unresolved]. The localized display name prepared after loading must match the original localization operation. GDA-Q04 gates that result, not the supplied class field or getter.

**GDA-R07. Experience fields** [verified in 1.13c code (static)]. Experience has i32 fields Amazon, Sorceress, Necromancer, Paladin, Barbarian, Druid, Assassin, ExpRatio. Class index 0..6 selects the matching class field independently of header order. Initialization publishes its supplied decoded records without a separately stored count. Its consumer owns row selection and bounds.

**GDA-R08a. Monster fields and getter** [verified in 1.13c code (static)]. Negative indices and indices at or above the supplied MonStats count yield no record. Valid indices yield that monster's record. MonStatsEx is a linked MonStats2 identity read as signed i16 by the mode getter.

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

**GDA-R08b. Derived monster animation values** [unresolved]. Initialization prepares two additional animation-rate values from base-class animation data and Velocity/Run relationships. These are not text columns. GDA-Q05 gates their complete production. Supplied-value consumers retain their own evidence labels.

**GDA-R09a. Extended monster fields** [verified in 1.13c code (static)]. MonStats2 has HitClass (u8), direction counts dDT/dNU/dWL/dGH/dA1/dA2/dBL/dSC/dS1..dS4/dDD/dKB/dSQ/dRN (u8), corresponding availability flags mDT..mRN, and boolean noGfxHitTest/noMap/noOvly/isSel/alSel/noSel/shiftSel/corpseSel/revive/isAtt/small/large/soft/critter/Shadow/noUniqueShift/compositeDeath/inert/objCol/deadCol/unflatDead. inTown belongs to MonStats.

**GDA-R09b. Mode-availability getter** [verified in 1.13c code (static)]. For monster class index and mode 0..15, return 0 if the MonStats index is invalid. Read MonStatsEx as signed i16; return 0 if negative, at or above MonStats2 count, or absent extended record. Otherwise select that mode's availability flag: false returns 0; true returns 1 << (mode mod 8). A true result is not always 1. Modes outside 0..15 are outside this logical operation.

**GDA-R10. Missile fields** [verified in 1.13c code (static)]. Missiles has dParam1/dParam2 (i32); HitClass/KnockBack/CollideType/Collision/ClientCol/CollideKill/CollideFriend/NextHit/NextDelay/Size/ToHit/AlwaysExplode (u8); Skill (linked 16-bit skill); and boolean LastCollide/Explosion/Pierce/CanSlow/CanDestroy/ClientSend/GetHit/SoftHit/ApplyMastery/ReturnFire/Town/SrcTown/NoMultiShot/NoUniqueMod/Half2HSrc/MissileSkill. Gameplay meanings belong to consumers.

**GDA-R11a. Skill fields and getter** [verified in 1.13c code (static)]. Negative skill indices and indices at or above count yield no supplied Skills record.

| Columns | Logical format |
|---------|----------------|
| srvstfunc, srvdofunc, ItemEffect | 16-bit numeric selector |
| aurafilter, Param1..Param8, ToHit, LevToHit, MinDam, MaxDam, MinLevDam1..5, MaxLevDam1..5, EMin, EMax, EMinLev1..5, EMaxLev1..5, ELen, ELevLen1..3 | i32 |
| HitShift, SrcDam | u8 |
| EType | linked element code, 8-bit |
| aurarangecalc, `calc1`..`calc4`, ToHitCalc, DmgSymPerCalc, EDmgSymPerCalc, ELenSymPerCalc | compiled expression word, 32-bit; representation/evaluation owned by the expression spec |
| decquant, lob, progressive, finishing, passive, aura, periodic, prgstack, InTown, Kick, InGame, repeat, stsuccessonly, stsounddelay, weaponsnd, immediate, noammo, enhanceable, durability, UseAttackRate, TargetableOnly, SearchEnemyXY, SearchEnemyNear, SearchOpenXY, TargetCorpse, TargetPet, TargetAlly, TargetItem, AttackNoMana, ItemTgtDo, leftskill, interrupt, TgtPlaceCheck, ItemCheckStart, ItemCltCheckStart, general, scroll, usemanaondo, warp | boolean |

**GDA-R11b. Parameter calculation** [verified in 1.13c code (static)]. Return 0 for invalid skill index, absent record, or signed level <= 0. Otherwise return i32(Param3 + i32(i32(level - 1) * Param4)). Multiplication/addition wrap as i32, with no clamp.

**GDA-R12. Stat-definition fields** [verified in 1.13c code (static)]. ItemStatCost has Stat (linked 16-bit stat), `itemevent1`/`itemevent2` (linked 16-bit event), `itemeventfunc1`/`itemeventfunc2` (16-bit selector), ValShift/Send Bits/Send Param Bits/op (u8), MinAccr/Divide/Multiply/Add (i32), maxstat/op base/op `stat1`..3 (linked 16-bit stat), and boolean Send Other/Signed/damagerelated/itemspecific/direct/UpdateAnimRate/fMin/fCallback/Saved/CSvSigned. RLI-R02b owns the ValShift/MinAccr consumer behavior.

**GDA-R13. State-definition fields** [verified in 1.13c code (static)]. State fields: state/`overlay1`..4/castoverlay/removerlay/pgsvoverlay/stat/onsound/offsound/skill/missile (16-bit linked ids); setfunc/remfunc/group/gfxclass/cltevent/clteventfunc/cltactivefunc/srvactivefunc (16-bit); colorpri/colorshift/light-r/light-g/light-b/gfxtype (u8); itemtype (linked 16-bit); itemtrans (linked 8-bit); and boolean nosend/aura/hide/transform/pgsv/active/remhit/damblue/damred/attblue/attred/curse/curable/plrstaydeath/monstaydeath/bossstaydeath/disguise/restrict/blue/armblue/rfblue/rcblue/rlblue/rpblue/stambarblue/armred/rfred/rcred/rlred/rpred/exp/shatter/life/udead/green/nooverlays/noclear/bossinv/meleeonly/notondead. Definition fields are separate from unit state membership.

**GDA-R14. Item-type fields and publication** [asm-only]. The identified ItemTypes loader writes its prepared collection and count; its dispatch-root caller has not been established. Code is a four-byte code also participating in link-building. Equiv1/Equiv2/Shoots/Quiver are linked 16-bit item types; BodyLoc1/BodyLoc2/StaffMods/Class/StorePage linked 8-bit identities; Repair/Body/Throwable/Reload/ReEquip/AutoStack/Magic/Rare/Normal/Charm/Gem/Beltable/MaxSock1/MaxSock25/MaxSock40/TreasureClass/Rarity/CostFormula/VarInvGfx u8; InvGfx1..6 strings with limit 31. Arbitrary custom link-building remains gated by GDA-Q01.

**GDA-R15a. Treasure source fields** [verified in 1.13c code (static)]. TreasureClassEx has Treasure Class (string, limit 32); group/level/Unique/Set/Rare/Magic (16-bit words); Picks/NoDrop (i32); Item1..10 (strings, limit 64); Prob1..10 (i32). Two reserved words have no descriptor column.

**GDA-R15b. Selected treasure conversion** [verified in 1.13c code (static)]. With supplied source, initialized destination, generated group count, and reserved words:
- Nonzero group writes low16(i32(group + generated group count)); group 0 leaves destination group unchanged.
- Copy level and Magic/Rare/Set/Unique words unchanged.
- Copy Picks as i32, replacing exactly 0 with 1. Negative Picks is unchanged.
- Copy NoDrop and the two reserved words unchanged.
Probability entry construction/sums are outside this supplied-field conversion.

**GDA-R16. Named flag-test results** [verified in 1.13c code (static)]. Supplied false flags return 0. True MonStats inTown/NoShldBlock tests return 4; true Missiles GetHit returns 64 and SoftHit 128; true Skills interrupt returns 128. These are original primitive return values, not a requirement for physical record storage. MonStats noRatio/killable are distinct named fields. Consumers own responses to flags.

### Id map

| Id | Summary | Status label | Gates | Vectors / table record |
|----|---------|--------------|-------|------------------------|
| GDA-R01 | Selected load path | verified in 1.13c code (static) | none | GDA-V01 |
| GDA-R02 | Full preparation integration | unresolved | GDA-Q01 / GDA-Q03 / GDA-Q02 / GDA-Q04: full preparation | none |
| GDA-R03a | Column schema | verified in 1.13c code (static) | none | GDA-V06, GDA-V02, GDA-V08 |
| GDA-R03b | Direct text-cell conversions | verified in 1.13c code (static) | none | GDA-V06, GDA-V03 |
| GDA-R03c | ASCII header comparison | verified in 1.13c code (static) | none | GDA-V04 |
| GDA-R03d | Other text conversions | unresolved | GDA-Q01 / GDA-Q02: full text conversion | none |
| GDA-R04a | Prepared table ownership | verified in 1.13c code (static) | none | GDA-V09, GDA-V02, GDA-V10 |
| GDA-R04b | PetType count | verified in 1.13c code (static) | none | GDA-V05 |
| GDA-R05 | Difficulty fields | verified in 1.13c code (static) | none | GDA-V06, GDA-V02 |
| GDA-R06a | Player-class fields and getter | verified in 1.13c code (static) | none | GDA-V07 |
| GDA-R06b | Class display-name preparation | unresolved | GDA-Q04: localized display result | none |
| GDA-R07 | Experience fields | verified in 1.13c code (static) | none | GDA-V08 |
| GDA-R08a | Monster fields and getter | verified in 1.13c code (static) | none | GDA-V11, GDA-V09, GDA-V10 |
| GDA-R08b | Derived monster animation values | unresolved | GDA-Q05: complete rate production | none |
| GDA-R09a | Extended monster fields | verified in 1.13c code (static) | none | GDA-V12 |
| GDA-R09b | Mode-availability getter | verified in 1.13c code (static) | none | GDA-V12 |
| GDA-R10 | Missile fields | verified in 1.13c code (static) | none | GDA-V13 |
| GDA-R11a | Skill fields and getter | verified in 1.13c code (static) | none | GDA-V14, GDA-V15 |
| GDA-R11b | Parameter calculation | verified in 1.13c code (static) | none | GDA-V14 |
| GDA-R12 | Stat-definition fields | verified in 1.13c code (static) | none | GDF-V02 |
| GDA-R13 | State-definition fields | verified in 1.13c code (static) | none | GDA-V16 |
| GDA-R14 | Item-type fields and publication | asm-only | GDA-Q01: custom linked values; root path untraced; field identities remain available | GDF-V04 |
| GDA-R15a | Treasure source fields | verified in 1.13c code (static) | none | GDA-V17 |
| GDA-R15b | Selected treasure conversion | verified in 1.13c code (static) | none | GDA-V17 |
| GDA-R16 | Named flag-test results | verified in 1.13c code (static) | none | GDA-V11, GDA-V13, GDA-V15 |

## 5. Outputs and side effects

| Output / effect | Rule |
|-----------------|------|
| requested load form / generation / initialization abort | GDA-R01 |
| named fields and direct normalized cells | GDA-R03a, GDA-R03b and field rules |
| header equality | GDA-R03c |
| published logical collections / requested counts | GDA-R04a, GDA-R01; ItemTypes only at GDA-R14 asm-only boundary |
| PetType count / initialization abort | GDA-R04b |
| Difficulty initialization abort when prepared count is not 3 | GDA-R05 |
| published supplied Experience records, without a separately stored count | GDA-R07 |
| selected logical record or no record; named field values | GDA-R06a, GDA-R08a, GDA-R11a |
| availability mask / 0 | GDA-R09b |
| parameter result / 0 | GDA-R11b |
| converted treasure fields | GDA-R15b |
| flag result | GDA-R16 |
| no RNG, gameplay stat/state/event/unit writes, or messages | all local rules |

## 6. Branches and edge cases

- Supported difficulty is 0..2; no substitute is defined outside that domain.
- No-record results are visible to the caller; its response belongs to its gameplay contract.
- A prepared field may be derived by code rather than a text column.
- Nonzero flag cells include 2 and negative values, not only 1.
- Zero treasure Picks becomes 1; negative Picks remains negative.
- A collection index selects a prepared record; it is not a search by the record's stored identity, and a linked identity stored in a record is distinct from its collection index.

## 7. Open questions

| Id | Kind | Blocked assertions / boundary | Missing evidence |
|----|------|-------------------------------|------------------|
| GDA-Q01 | behavior | Full name/code links, expression/callback conversions, descriptor-specific preparation; direct conversions remain available | complete conversion handlers and their consumers |
| GDA-Q02 | behavior | Missing-column defaults and complete name-link/callback cell behavior | final missing-field switch and indirect conversion contracts |
| GDA-Q03 | integration | Full treasure probability/group generation and global compiled-row equivalence; supplied treasure conversion remains available | probability producers, generated-group lifecycle, explanation of compiled/text mismatch |
| GDA-Q04 | integration | Localized class display-name result | imported localization/string behavior used by the identified writer |
| GDA-Q05 | integration | Complete derived monster animation-rate production; supplied-value consumers remain separate | base animation-data selection contracts used by the identified writer |
| GDA-Q06 | observation | Gameplay/runtime-confirmation claims only | attributable executions of verified behavioral rules; table/memory reads are correspondence evidence |

GDA-Q02 also gates GDF-R06b's absent direction-column defaults, and GDA-Q03 the global MonStats compiled correspondence behind GDF-R08. Attempts and what closes each gap are in [Open evidence gaps](../plans/gaps/core-game-data-table-records.md). Observation summaries do not gate supplied-input results.

## 8. Test vectors

### 8.1 Executable vectors

Every vector uses no draw and leaves RNG state unchanged. Supplied logical records are fixtures. Record reads corroborate only named fields; they do not observe execution of a getter or flag-test call.

**GDA-V01** — GDA-R01; supplied load-path outcomes.
Inputs: final collection F has count 3; intermediate I count 3; I and F are distinct supplied decoded collections; all schema fields present; no conversion side effects; requested count true. Fresh cases:
- generate false / read compiled true; selected compiled success F: requests compiled only; publishes F/count 3.
- generate false / read compiled true; compiled missing, text available: requests compiled only; aborts, no publication.
- generate false / read compiled false; text success F: requests text only; publishes F/count 3.
- generate true / read compiled true; generation text missing, compiled available: requests generation text, then aborts. No decoder, writer, final compiled read or publication is requested.
- generate false / read compiled false; final text missing, compiled available: requests text only, then aborts. No decoder, compiled fallback or publication is requested.
- generate true / read compiled true; text decode succeeds I, writer success, compiled decode succeeds F: exact ordered log is text-decode(I), writer(I), discard(I), compiled-decode(F), publish(F/count 3).
- generate true / read compiled false; first text decode succeeds I, writer success, second text decode succeeds distinct F: exact ordered log is text-decode(I), writer(I), discard(I), second-text-decode(F), publish(F/count 3). It publishes F, not the reused intermediate I.
Trace follows GDA-R01 independently of file availability in the unused form. runtime_observed: no. These supplied decoded outcomes discriminate resource selection and publication order; they do not execute or verify original pre-conversion record zeroing. Full decoding/writer execution is gated separately.

**GDA-V02** — GDA-R03a, GDA-R05, GDA-R04a; supplied initialization schema.
Inputs: Difficulty schema has the 22 GDA-R05 columns followed by terminator, prepared count 3, publish-count requested; Experience schema has its eight GDA-R07 columns followed by terminator, publish-count not requested. Trace: stop each schema at terminator, publish named collections, condition count publication. Exact outputs: Difficulty 22 fields and published count 3; Experience 8 fields and no separately published count. Fresh Difficulty prepared count 2 aborts initialization. Fresh schema-termination fixture: descriptors in order are (name end, kind i32), (name after, kind i32), (name arbitrary-stop-name, kind Stop), (name must-not-be-read, kind i32). Exact selected field names are end and after, count 2; the final descriptor is not visited. Termination therefore depends on conversion kind, not the name end. runtime_observed: no.

**GDA-V03** — GDA-R03b; supplied direct conversion discrimination.
Inputs: selected fields i32 "-100", u16 "65537", u8 "-1", flag "2", flag "-1", flag "", four-code "SL", bounded string "abcdef" with byte limit 3; initial other flag true. Trace: wrapping accumulation, low-bit store, nonzero test, space padding, length truncation. Exact outputs: -100, 1, 255, true, true, false, "SL  ", "abc"; the other flag stays true. Fresh empty i32 gives 0, four-code "" gives four spaces, and four-code "ABCDE" yields "ABCD". Fresh selected i32 cell "2147483648" yields -2147483648; cell "-2147483649" yields 2147483647. These cases distinguish wrapping accumulation from saturation. A separate supplied bounded-string cell contains the two raw bytes [0xC3, 0xA9] and has byte limit 1. Trace: retain the first byte only. Exact retained bytes [0xC3], rather than retaining both bytes as one Unicode character. This defines neither a text encoding nor locale behavior. runtime_observed: no.

**GDA-V04** — GDA-R03c; supplied ASCII comparison path.
Inputs: header names Stat and Value; descriptor names stat and value; a negative comparison Stat versus state; ASCII comparison mode. Trace: case-fold letters, compare full strings. Exact outputs: equal, equal, not equal. This tests the original comparison path, not duplicate/missing-column defaults or another locale. runtime_observed: no.

**GDA-V05** — GDA-R04b; supplied PetType count.
Inputs: fresh decoded counts 20, 0, -1, 255, 256. Trace: reject >=256; nonpositive becomes 0. Exact outputs: published 20, published 0, published 0, published 255, initialization abort. The original count 20 is corroborated by PetType text and compiled count records. runtime_observed: no.

**GDA-V06** — GDA-R03a/GDA-R03b, GDA-R05; original-profile difficulty field preparation.
Inputs: the 22 GDA-R05 i32 columns; rows Normal/Nightmare/Hell at indices 0/1/2; prepared count 3; cells Hell.ResistPenalty = "-100", Hell.UberCodeOddsNormal = "20", Hell.UltraCodeOddsNormal = "30", Hell.MonsterFreezeDivisor = "4", Nightmare.MonsterColdDivisor = "2", Hell.LifeStealDivisor = "3", Normal.ManaStealDivisor = "1", Hell.HireableBossDamagePercent = "25". Trace: parse each selected cell as i32 and bind by column. Exact outputs: the corresponding fields are -100, 20, 30, 4, 2, 3, 1, 25. All are independent of header order. Fresh supplied reordering fixture: retain count 3 and the named rows; provide all 22 headers, with UberCodeOddsNormal first, ResistPenalty second, then the remaining 20 GDA-R05 headers in their listed order. Every fixture cell is "0" except Hell.UberCodeOddsNormal "20" and Hell.ResistPenalty "-100". Exact Hell outputs remain UberCodeOddsNormal 20 and ResistPenalty -100; its other 20 fields are 0. The other two rows have all fields 0. This fixture tests name binding, not original values of those zero cells. runtime_observed: no. Named compiled correspondence is in the companion; it does not execute this preparation path.

**GDA-V07** — GDA-R06a/GDF-R01; supplied class record getter.
Inputs: class index 1, count 7, supplied row Sorceress with BlockFactor 20, ManaRegen 120, ToHitFactor -15. Trace: 1 in range; BlockFactor unsigned. Exact outputs: Sorceress record; 20, 120, -15. Fresh indices 7 and -1 give no record; no dereference is required by this boundary. Fresh supplied Sorceress fixture keeps the valid index/count and other fields but sets BlockFactor 200; the field read returns 200, not signed -56. This is a fixture, not a new original-table cell claim. runtime_observed: no. Compiled correspondence for the named fields is in the companion.

**GDA-V08** — GDA-R03a, GDA-R07; schema field identity.
Inputs: Experience schema with its eight named i32 columns, class index 3, row 10. Trace: class 3 is Paladin. Exact selected field identities: (Experience,row 10,Paladin) and (Experience,row 10,ExpRatio). The identity-only case asserts no cell value. Fresh supplied header-order fixture: headers ExpRatio, Assassin, Druid, Barbarian, Paladin, Necromancer, Sorceress, Amazon; row 10 cells "101", "202", "303", "404", "505", "606", "707", "808" in that order. Supply rows 0..9 with all eight cells "0". Exact selected row-10 outputs remain Paladin 505 and ExpRatio 101, not the fourth and eighth input positions. These are fixture values, not new original-table claims. runtime_observed: no.

**GDA-V09** — GDA-R08a, GDA-R04a; supplied monster record selection.
Inputs: MonStats collection count 734, entry 78 = record R78 whose identity is `sandleaper1`; query 78, then 734, then -1. Other entries are opaque and not accessed. Trace: 78 is in range; 734/-1 are not. Exact outputs: R78, no record, no record. Separate supplied fixture: collection count 3; all entries present. Entry 0 has stored Id 0 and Velocity 0; entry 1 is record A with stored Id 2 and Velocity 11; entry 2 is record B with stored Id 1 and Velocity 22. Query index 1. Trace: 1 is in range; select collection entry 1 without searching its stored Id. Exact output: record A with Id 2 and Velocity 11, not record B with Id 1 and Velocity 22. These stored identities and values are synthetic supplied fields, not original-table cell claims. runtime_observed: no. Named-record identity correspondence is in the evidence companion.

**GDA-V10** — GDA-R08a, GDF-R08, GDA-R04a; supplied original-profile monster identity.
Inputs: MonStats count 734, index 560, supplied entry act5hire1 with identity 560, MonStatsEx identity 560, Velocity 8, Run 14. Trace: 560 in range. Exact output: that act5hire1 record; identity 560, linked identity 560, Velocity 8, Run 14. runtime_observed: no. Named text cells and compiled correspondence corroborate these inputs. The different collection-index/stored-identity discriminator is GDA-V09; this identity-correspondence case does not duplicate that fixture. For independent supplied difficulty-column selection, retain this valid selected record and supply Drain 7, Drain(N) 13, Drain(H) 29. Difficulty 0, 1 and 2 select base, Nightmare and Hell respectively; exact selected unsigned field values are 7, 13 and 29. This asserts field selection, not the leech calculation or new original-table values.

**GDA-V11** — GDA-R08a, GDA-R16; supplied named monster flags.
Inputs: act5hire1 has inTown true, NoShldBlock false; `sandleaper1` has both false. Trace: select named flags and their return masks. Exact outputs: act5hire1 town 4 / shield 0; `sandleaper1` town 0 / shield 0; booleans true/false/false/false respectively. A separate supplied record has inTown false and NoShldBlock true; exact town/shield outputs are 0/4 and boolean predicates false/true. This distinguishes the true NoShldBlock return mask from 1 or a different flag. runtime_observed: no. Named-input correspondence is recorded in the companion; it does not observe the flag-test calls.

**GDA-V12** — GDA-R09a/GDA-R09b; supplied mode availability.
Inputs: monster index 78, MonStats count 734, that record's MonStatsEx 78; MonStats2 count 609 and supplied linked record availability mKB true / mGH true / mA1 false, every other flag false; modes 13, 3, 4. Trace: both indices valid; true masks are 1 << 5 and 1 << 3. Exact outputs: 32, 8, 0. Fresh asymmetric supplied flags mKB=true/mGH=false give mode 13 result 32 and mode 3 result 0; mKB=false/mGH=true give mode 13 result 0 and mode 3 result 8. All other flags remain false. Fresh invalid MonStatsEx -1 or 609 gives 0. Fresh in-range absent-record case: monster index 0/count 1, MonStatsEx 0, MonStats2 count 1, no extended record at that selected identity; querying mode 13 returns 0 without a flag read. Fresh invalid monster indices -1 and 734, querying mode 13, each return 0 without reading a MonStats record or its link. Separate supplied link-selection fixture: MonStats count 3 with all three records present; query monster index 1 whose record has MonStatsEx 2. The other MonStats records have MonStatsEx 0. MonStats2 count 3 with all three records present; entry 0 has every availability flag false, entry 1 has every flag false including mGH, and entry 2 has mGH true with every other flag false. Query mode 3. Trace: select monster entry 1, then its linked extended entry 2; mGH is true, yielding 1 << 3. Exact output 8. Using the monster index 1 as the extended index would incorrectly return 0. This is supplied-record linkage, not a new compiler correspondence claim. runtime_observed: no.

**GDA-V13** — GDA-R10, GDA-R16; supplied missile flags.
Inputs: record A GetHit true, SoftHit true, ClientSend false; record B GetHit false, SoftHit false, ClientSend true; all other flags false. Trace: GetHit/SoftHit are distinct from ClientSend. Exact test results: A 64/128, B 0/0. runtime_observed: no.

**GDA-V14** — GDA-R11a/GDA-R11b; supplied skill parameter operation.
Inputs: skill index 271, count 357, supplied record Param3 20 / Param4 5, level 4. Trace: valid index and positive level; 20 + (4 - 1) * 5 = 35. Exact output 35. Fresh level 0 or -1 gives 0; fresh index 357 gives no record and parameter result 0. For the fresh negative-index case, also supply entry 0 with Param3 9 / Param4 2; index -1 still gives no record and 0, rather than using entry 0's result 15 at level 4. Fresh valid-index overflow case: index 271/count 357, level 3, Param3 2147483640, Param4 10; i32(2147483640 + 2 * 10) = -2147483636, with no clamp. Fresh absent-record case supplies index 0/count 1, level 1, and no record at the selected identity: parameter result 0, with no parameter-field read. runtime_observed: no.

**GDA-V15** — GDA-R11a, GDA-R16; supplied skill flags.
Inputs: record A interrupt true / leftskill false; record B interrupt false / leftskill true; other flags false. Trace: select interrupt only. Exact outputs: A 128 / interrupt set, B 0 / interrupt clear. runtime_observed: no.

**GDA-V16** — GDA-R13; schema field identity.
Inputs: States identity 54 and the GDA-R13 schema. Trace: select the state definition independently from unit state membership. Exact selected identities: (States,54,stat), (States,54,skill), (States,54,curse), (States,54,life). curse and life are independent boolean fields. No cell values or active state membership are asserted. runtime_observed: no.

**GDA-V17** — GDA-R15a/GDA-R15b; supplied treasure-field conversion.
Inputs: Picks 0, NoDrop 15, group 0, level 7, Magic 1024, Rare/Set/Unique 983 each; generated group count 5; source reserved words 0/0; initialized destination group 77, all other converted fields 0. Trace: zero Picks becomes 1, zero group leaves 77, copy remaining fields. Exact outputs: group 77, level 7, Picks 1, NoDrop 15, Magic/Rare/Set/Unique 1024/983/983/983, reserved 0/0. Fresh variant Picks 3 gives Picks 3; fresh variant Picks -2 gives -2. Fresh nonzero-group cases keep the other inputs unchanged: group 10 plus generated count 5 yields destination group 15; group 65535 plus count 5 yields low16(65540) = 4. Fresh copy discriminator retains the other inputs, supplies source reserved words 17/29 and initialized destination reserved words 91/92: exact destination reserved words become 17/29, not 91/92 or a swapped pair. These are supplied words, not new table-column or cell claims. A separate quality-copy fixture retains the original inputs and initialized destination, supplying source Magic 101, Rare 202, Set 303 and Unique 404. Copy each quality word to its same-named destination. Exact destination Magic/Rare/Set/Unique values are 101/202/303/404; the remaining outputs are group 77, level 7, Picks 1, NoDrop 15 and reserved words 0/0. All four quality values differ, so any swapped pair produces a different result. These are synthetic supplied words, not original-table cell claims. runtime_observed: no.

GDA-R12 (stat-definition fields) is exercised by GDF-V02, and GDA-R14 (item-type fields) by GDF-V04, in [Game data fields](core-game-data-fields.md#81-executable-vectors); those vectors assert the identities those fields select.

### 8.2 Observation vectors

- **GDA-O01**: Corroborate derived animation values and identified writer for additional classes. Existing values are correspondence only; complete production remains GDA-Q05. [Procedure](../research/spec-evidence/core-game-data-table-records.md#observation-procedures).
- **GDA-O02**: Corroborate PetType count production at initialization. Count 20 is a correspondence read, not execution of its producer. [Procedure](../research/spec-evidence/core-game-data-table-records.md#observation-procedures).

## 9. Acceptance and dependencies

Accept verified supplied-record/schema/cell/control-flow assertions demonstrated by GDA-V01..GDA-V17 separately from gated producer integration. GDA-R14 item-type fields and publication retain their asm-only label pending the untraced dispatch root; GDF-V04 supports only the selected identities it asserts. No assertion requires the original memory layout.

Full reproduction also requires all in-scope gates resolved and attributable observations of verified behavioral rules where possible. Table reading does not satisfy gameplay observation. This draft does not imply user approval.

**Consumed contracts:** [Game data fields](core-game-data-fields.md) supplies the class, stat, state, mode, item-type, monster and skill identities that select rows and name linked fields; GDF-V02 and GDF-V04 carry the vectors for GDA-R12 and GDA-R14. Owners of the cited RLI ids: [core-record-layouts-and-ids](core-record-layouts-and-ids.md) (RLI).

**Supporting conventions:** [foundations notation](core-numeric-and-rng-foundations.md#1-scope) supplies integer wrapping; its sections 3.2/3.3 describe shared semantic records. RLI-R01a/RLI-R01b consumes these records rather than producing them. No RNG helper is called by this boundary.

**Expression boundary:** GDA-R11a identifies the named 32-bit compiled references. The expression owner, SCX-R01..SCX-R11 and SCX-R13..SCX-R15, defines evaluation/pool selection; SCX-R12 leaves compiler-produced output properties gated. No expression evaluation or compiler result is asserted by GDA-R03b or GDA-V03. GDA-R03d/GDA-R02 retain conversion integration gates.

**Reverse consumers:** gameplay specs citing shared fields, getters and flag tests, including mitigation/pipeline, hit recovery, item events, knockback, leech, PvP, shield/block, monster scaling, party experience, Assassin damage/sentries, leveled skill damage, and movement/freeze. Citation relationships do not themselves impose a plan order.

## 10. Related specs

- [Game data fields](core-game-data-fields.md): identities and names used by these records.
- [Numeric and RNG foundations](core-numeric-and-rng-foundations.md): shared numeric/stat/state/record contracts.
- [Skill calc expressions](skill-calc-expressions.md): expression contract.
- [Damage mitigation](combat-damage-mitigation.md), [Damage pipeline](combat-damage-pipeline-order.md), [Hit recovery](combat-hit-recovery.md), [Knockback](combat-knockback.md), [Party experience sharing](player-party-experience-sharing.md): consumers.

## Appendix A. Provenance

Evidence target: Lord of Destruction Expansion 1.13c, Windows x86. Original `D2Common.dll`,
`D2Game.dll`, `Fog.dll` and `Storm.dll` analysis specimens are identified by exact SHA-256
fingerprints in the [source manifest](../research/source-manifest.md).

The [public evidence companion](../research/spec-evidence/core-game-data-table-records.md) provides
module-relative function locations, export ordinals where recorded, descriptor/consumer
mappings, rule-linked rationale, named table cells and observation procedures. Static readings
were recorded on 2026-10-12/13 and targeted rereadings on 2026-10-14 using disassembly,
reconstructions checked against original bytes, and independent instruction listings where
recorded. Function descriptions are analyst labels.

The table records dated 2026-10-14 distinguish effective text cells, selected compiled values
and loaded-memory correspondence; effective-file hashes were not captured for those reads.
Supplying MPQs, complete override inventories and fresh loaded-DLL/compiled correspondence remain uncaptured; the manifest and companion state the profile limits.
Table/memory correspondence does not establish gameplay execution. Observations remain pending
and the spec remains draft. Public verification uses the identified original game artifacts
and recorded data scope without requiring the private reconstruction project.
