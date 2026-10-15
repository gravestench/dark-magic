# Game Data Fields

## Description

The game gives every stat, state, character class, monster, skill, item type and animation mode a fixed number and name. Its systems use these numbers to refer to the same thing consistently: the state that marks a unit as frozen, for example, is the same number wherever it is checked. This reference lists the original game's names and numbers that other behaviors rely on, and a few fixed choices the game makes among them.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [Numeric and RNG foundations](core-numeric-and-rng-foundations.md)
Ids: GDF-R, GDF-Q, GDF-V, GDF-O; policy in [README conventions](README.md#conventions)
Evidence: [source, table records, and observation history](../research/spec-evidence/core-game-data-fields.md)
Open evidence gaps: [remaining assertions](../plans/gaps/core-game-data-fields.md)

## 1. Scope

This dictionary defines the original profile's identity mappings and names: player classes, stats, states, player and monster modes, item types, monsters and skills. It also defines the mode-table preparation, the code-fixed stat/state choices and the handler-selector comparison stated below, and the shared name for one unit condition. Table-record layouts, getters, text conversion, flag tests and table loading belong to [Game data table records](core-game-data-table-records.md). Gameplay formulas using these identities belong to their consuming specs. Shared stat/state operations belong to RLI-R02 and RLI-R03.

## 2. Trigger and ownership

- D2Common prepares the identity tables during process initialization, before a game is created, through the loading path of GDA-R01. These are initialization operations, not frame updates.
- D2Game and D2Common consumers cite these identities when their gameplay paths need them. The selections stated here are read only.
- Client and server use the same identities. This dictionary defines no message.

## 3. Inputs and data

Every table path below is under data/global/excel/ in the original 1.13c archive set. Columns present in the effective files use their text-header spellings. Descriptor-only columns use the original schema's names; their missing-column behavior is GDA-Q02.

| Input | Semantic source | Width / format | Meaning |
|-------|-----------------|----------------|---------|
| text table | effective named .txt table | 8-bit byte strings, tab-separated cells and CRLF rows; converted by GDA-R03b | source identities and names |
| identity | caller's class, stat, state, mode, monster, skill or item-type identity | i32 unless a field width is stated | selects a named row |
| definition count | prepared ItemStatCost count (GDA-R04a) | i32 | GDF-R04 stat-read guard |
| state presence | unit state collection | named state ids and boolean membership | consumed through RLI-R03 |
| skill record | supplied prepared Skills record (GDA-R11a) | srvdofunc 16-bit selector | GDF-R09 comparison |
| unit alive exclusion | `unit.aliveExclusionCondition` | boolean | Shared name for one engine-defined unit condition that alive and availability tests in consuming specs treat as not alive. Its producers and wider meaning are not established here; each consumer owns its test and evidence label |

| Mapping / constant | Origin | Binding |
|--------------------|--------|---------|
| stat/state names and explicit ID/id cells | ItemStatCost / States data | original-profile identities corroborated with prepared-table indices; consumers may also fix individual ids |
| player/monster mode names, token/code pairs | PlrMode / MonMode data | original-profile row identities; consumers choose code-defined mode numbers |
| monster identities | MonStats Id/hcIdx; linked MonStats2 Id | original-profile prepared-record identities; separator rows excluded |
| item-type identities | ItemTypes ItemType/Code | original-profile row identities; link-building conversion defined by code (GDA-R14) |
| skill identities | Skills skill/Id | original-profile prepared-record indices |
| handler selectors such as srvdofunc/`itemeventfunc1` | named table columns | data selects engine-defined handlers; a selector is not a skill/stat id |
| fixed stat 171, states 54 and 101, selector values 67/76 | engine code | fixed by original code |
| unit alive exclusion | engine code | one engine-defined unit condition; name assigned here, meaning unresolved beyond each consumer's test |
| custom identity remapping | unresolved | no support or behavior is established here |

[Table records](../research/spec-evidence/core-game-data-fields.md#table-verification-2026-10-14) distinguish text cells from compiled correspondence. An ID cell matching a row index does not establish that the compiler uses that cell to assign every linked identity.

## 4. Rules

All operations below draw no RNG.

**GDF-R01. Class identities** [verified in 1.13c data]. Prepared class order is 0 Amazon, 1 Sorceress, 2 Necromancer, 3 Paladin, 4 Barbarian, 5 Druid, 6 Assassin. The text Expansion separator at zero-based data row 5 is not a class.

**GDF-R02. Stat identities** [verified in 1.13c data]. ItemStatCost has 359 data rows; ID equals zero-based row index throughout, and Stat names the row.

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

**GDF-R03. State identities** [verified in 1.13c data]. States has 185 data rows; id equals zero-based row index throughout and state names the row.

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

The code-fixed block-animation state test is specified by GDF-R04; data identity 101 is holyshield.

**GDF-R04. Fixed code roles** [verified in 1.13c code (static)]. The defense reader selects stat 171 skill_armor_percent, not stat 172 alignment. With supplied definition count <= 171 it skips that read and supplies 0; otherwise it selects stat 171, layer 0 through RLI-R02a/RLI-R02b. The combat eligibility check selects state 54 uninterruptable, not stat 54, through RLI-R03. The block-animation rate path selects state 101 holyshield through RLI-R03; this owns its selected state identity, not the complete rate calculation.

**GDF-R05a. Player-mode identities** [verified in 1.13c data]. PlrMode has 20 zero-based mode ids:

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

**GDF-R05b. Player-mode preparation** [verified in 1.13c code (static)]. Preparation reads Name (string, limit 31) and Token (four-byte code), not Code. Player-type and player-mode definitions are combined at initialization; the player-mode subset retains its own order.

**GDF-R06a. Monster-mode identities** [verified in 1.13c data]. MonMode has 16 rows:

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

**GDF-R06b. Monster-mode preparation** [verified in 1.13c code (static)]. Preparation reads name (string, limit 31), token (four-byte code), and DT_Dir/NU_Dir/WL_Dir/GH_Dir/A1_Dir/A2_Dir/BL_Dir/SC_Dir/S1_Dir..S4_Dir/DD_Dir/KB_Dir/SQ_Dir/RN_Dir (u8); code is not prepared. Token reuse does not change mode identity.

**GDF-R07. Item-type identities** [verified in 1.13c data]. ItemTypes data row 4 is Gold, Code gold; row 51 is Any Shield, Code shld. The Expansion separator is text data row 58. Both cited identities precede it.

**GDF-R08. Monster and skill identities** [verified in 1.13c data]. The cited monster class indices below agree with MonStats hcIdx cells and prepared-record correspondence. Id names the text row. The original 734 ordinary text rows follow hcIdx order, excluding the Expansion separator at data row 410. A linked identity stored inside a prepared record is distinct from its collection index; global equality is not established. MonStats2 Id names linked extended records; its count need not equal MonStats count.

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

**GDF-R09. Handler-selector identity** [verified in 1.13c code (static)]. The skill eligibility exception tests prepared srvdofunc for 67 or 76, not skill identity. Either match takes the exception branch; otherwise execution continues to the interrupt test. This owns the selected-field comparison only.

### Id map

| Id | Summary | Status label | Gates | Vectors / table record |
|----|---------|--------------|-------|------------------------|
| GDF-R01 | Class identities | verified in 1.13c data | none | TD-CLASS |
| GDF-R02 | Stat identities | verified in 1.13c data | none | TD-STAT |
| GDF-R03 | State identities | verified in 1.13c data | none | TD-STATE |
| GDF-R04 | Fixed code roles | verified in 1.13c code (static) | none | GDF-V02 |
| GDF-R05a | Player-mode identities | verified in 1.13c data | none | TD-PLAYER-MODE |
| GDF-R05b | Player-mode preparation | verified in 1.13c code (static) | none | GDF-V03, GDA-V02 |
| GDF-R06a | Monster-mode identities | verified in 1.13c data | none | TD-MONSTER-MODE |
| GDF-R06b | Monster-mode preparation | verified in 1.13c code (static) | GDA-Q02: absent direction-column defaults; supplied fields remain available | GDF-V03, GDA-V02 |
| GDF-R07 | Item-type identities | verified in 1.13c data | none | TD-ITEMTYPE |
| GDF-R08 | Monster and skill identities | verified in 1.13c data | none | TD-MONSTER-SKILL |
| GDF-R09 | Handler-selector identity | verified in 1.13c code (static) | none | GDF-V05 |

## 5. Outputs and side effects

| Output / effect | Rule |
|-----------------|------|
| identity maps | GDF-R01, GDF-R02, GDF-R03, GDF-R05a, GDF-R06a, GDF-R07, GDF-R08 |
| prepared mode names, tokens and direction values | GDF-R05b, GDF-R06b |
| selected stat/state identity / exception branch | GDF-R04, GDF-R09 |
| no RNG, gameplay stat/state/event/unit writes, or messages | all local rules |

## 6. Branches and edge cases

- A token can be reused by another mode: player Knock back and Sequence use GH despite differing Code cells.
- Expansion separator rows are not identities: CharStats, ItemTypes and MonStats each have one.
- Text identities and compiled identities require separate corroboration. A matching count alone proves no ordering. The global MonStats compiled-Id correspondence is unresolved (GDA-Q03) even though the cited identities match.
- A handler selector value can equal an unrelated skill identity; GDF-R09 compares the selector only.

## 7. Open questions

| Id | Kind | Blocked assertions / boundary | Missing evidence |
|----|------|-------------------------------|------------------|
| GDF-Q01 | observation | Gameplay/runtime-confirmation claims only | attributable executions of verified behavioral rules (GDF-O01); table/memory reads are correspondence evidence |

GDF-R06b's absent direction-column defaults are gated by GDA-Q02, and global MonStats compiled correspondence for GDF-R08 by GDA-Q03; both are tracked by [Game data table records](core-game-data-table-records.md#7-open-questions). Attempts and what closes each gap are in [Open evidence gaps](../plans/gaps/core-game-data-fields.md). Observation summaries do not gate supplied-input results.

## 8. Test vectors

### 8.1 Executable vectors

Every vector uses no draw and leaves RNG state unchanged. Supplied logical records are fixtures. Data-only checks identify named table cells rather than claim gameplay execution.

**GDF-V01** — data-only GDF-R02 / GDF-R03 check.
Inputs: ItemStatCost ID 99 / 171, States id 54 / 146. Select by named id and name column. Exact outputs: 99 item_fastergethitrate; 171 skill_armor_percent; state 54 uninterruptable; state 146 invis. No gameplay observation; named table records provide evidence.

**GDF-V02** — GDA-R12, GDF-R04; stat-definition selection boundary.
Inputs: requested fixed stat 171, layer 0, definition count 359. Trace: 359 > 171. Exact output: select ItemStatCost identity 171 skill_armor_percent and stat read (171,0); this vector does not execute RLI-R02a/RLI-R02b. Fresh count 171 or less skips read and supplies 0. The two state-call identity cases supply unit U and a state collection (state 54 present, state 101 absent): exact selected calls are state-presence(U,54) and state-presence(U,101); the nonzero predicates of consumed RLI-R03 results are true/false. The original state-test returns are 4194304 for state 54 (1 << 22) and 0 for absent state 101; no boolean return of 1 is implied. U is a player with a supplied extended stat list, state membership support and state count 185, within the stated RLI-R03 eligibility. runtime_observed: no.

**GDF-V03** — GDF-R05a/GDF-R05b and GDF-R06a/GDF-R06b; identity and selected-column check.
Inputs: player modes 19, 18, 3; monster modes 13, 14, 3. Exact outputs as name/token/code: player 19 Knock back/GH/KB, 18 Sequence/GH/SQ, 3 Run/RN/RN; monster 13 knockback/GH/KB, 14 sequence/xx/xx, 3 gethit/GH/GH. Preparation fixture: supply player-mode 19 Name 'Knock back', Token 'GH', Code 'KB' and monster-mode 13 name 'knockback', token 'GH', code 'KB', each direction cell '0'. Exact prepared player fields are Name 'Knock back'/Token 'GH  '; prepared monster fields name 'knockback'/token 'GH  ' and all 16 direction values 0; Code/code is not a prepared field. No gameplay observation; mode-change observation is GDF-O01.

**GDF-V04** — GDF-R07/GDA-R14; original identity and Code-field check.
Inputs: ItemTypes row identities 51 and 4, both before the separator. Exact outputs: 51 Any Shield / Code shld; 4 Gold / Code gold. The selected Code field is a four-byte code: supplied Code 'gold'/'shld' remains 'gold'/'shld'; full link-building is gated separately. No gameplay observation.

**GDF-V05** — GDF-R09; supplied handler selector branch.
Inputs: three records with (skill identity, srvdofunc) = (12,67), (13,76), (67,68). Trace: compare srvdofunc, not skill identity. Exact outputs: exception taken, exception taken, exception not taken / continue to interrupt test. runtime_observed: no.

### 8.2 Observation vectors

- **GDF-O01**: During knockback, observe player mode 19 / monster mode 13. Pending; data mapping alone does not observe mode changes. [Procedure](../research/spec-evidence/core-game-data-fields.md#observation-procedures).

## 9. Acceptance and dependencies

Accept data-only identities for the original profile and their named table records. Accept verified selection/preparation/comparison assertions demonstrated by GDF-V01..GDF-V05 separately from gated producer integration. No assertion requires the original memory layout.

Full reproduction also requires all in-scope gates resolved and attributable observations of verified behavioral rules where possible. Table reading does not satisfy gameplay observation. This draft does not imply user approval.

**Consumed operations:**

| Dependency | Supplied inputs | Result consumed here | Evidence / acceptance boundary |
|---|---|---|---|
| RLI-R03 state test | unit type and list, state count and selected state 54 or 101 | state mask or 0; nonzero is presence | verified in 1.13c code (static); read-only test, no state-set/update effects consumed |
| RLI-R02a / RLI-R02b stat read | selected stat 171, layer 0, definition count/record and existing ordered stat arrays | i32 stat value after the stated conditional minimum | verified in 1.13c code (static); supplied-array behavior, without claiming stat-array production |
| [Game data table records](core-game-data-table-records.md) | named tables, schema and supplied cells | GDA-R01 loading, GDA-R03b direct conversion, GDA-R04a prepared counts, GDA-R11a Skills records | labels as stated there; GDA-Q02 gates absent direction columns |

**Supporting conventions:** [foundations notation](core-numeric-and-rng-foundations.md#1-scope) supplies integer wrapping. RLI-R01a/RLI-R01b consumes this dictionary rather than producing its records. No RNG helper is called by this boundary.

**Reverse consumers:** gameplay specs citing shared identities and the unit alive-exclusion name, including mitigation/pipeline, hit recovery, item events, knockback, leech, PvP, shield/block, status application, timed states, monster scaling, party experience, cast start, Assassin damage/sentries, leveled skill damage, and movement/freeze. Citation relationships do not themselves impose a plan order.

Owners of the cited RLI ids: [core-record-layouts-and-ids](core-record-layouts-and-ids.md) (RLI).

## 10. Related specs

- [Game data table records](core-game-data-table-records.md): table layouts, getters, conversions, flag tests and loading.
- [Numeric and RNG foundations](core-numeric-and-rng-foundations.md): shared numeric/stat/state/record contracts.
- [Damage mitigation](combat-damage-mitigation.md), [Damage pipeline](combat-damage-pipeline-order.md), [Hit recovery](combat-hit-recovery.md), [Knockback](combat-knockback.md), [Party experience sharing](player-party-experience-sharing.md): consumers.

## Appendix A. Provenance

Evidence target: Lord of Destruction Expansion 1.13c, Windows x86. Original `D2Common.dll`
and `D2Game.dll` analysis specimens are identified by exact SHA-256 fingerprints in the
[source manifest](../research/source-manifest.md).

The [public evidence companion](../research/spec-evidence/core-game-data-fields.md) provides
module-relative function locations, export ordinals where recorded, consumer mappings,
rule-linked rationale, named table cells and observation procedures. Static readings were
recorded on 2026-10-12/13 and targeted rereadings on 2026-10-14 using disassembly,
reconstructions checked against original bytes, and independent instruction listings where
recorded. Function descriptions are analyst labels.

The table records dated 2026-10-14 distinguish effective text cells, selected compiled values
and loaded-memory correspondence; effective-file hashes were not captured for those reads.
The [2026-10-15 named-cell corroboration](../research/spec-evidence/core-game-data-fields.md#named-cell-corroboration-2026-10-15) adds effective text-file hashes and stated cell coverage only.
The [shared unit-condition name](../research/spec-evidence/core-game-data-fields.md#shared-unit-condition-name-2026-10-15) record maps the alive-exclusion name to its original location and consumers.
Supplying MPQs, complete override inventories and fresh loaded-DLL/compiled correspondence remain uncaptured; the manifest and companion state the profile limits.
Table/memory correspondence does not establish gameplay execution. Observations remain pending
and the spec remains draft. Public verification uses the identified original game artifacts
and recorded data scope without requiring the private reconstruction project.
