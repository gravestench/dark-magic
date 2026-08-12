# Diablo II quests and world progression research

Status: implementation-oriented research baseline for Lord of Destruction-era
Diablo II. The player-visible route is broadly stable across legacy patches and
Resurrected, but the runtime evidence cited here is D2MOO's reconstruction of
1.10f. Patch-sensitive details and unverified edge cases are called out rather
than silently generalized.

This document answers four separate questions:

1. what the player is asked to do;
2. which world areas and authored set pieces make that possible;
3. which events change quest or world state;
4. which facts still require a trace against the original game.

It complements [MAP_GENERATION.md](../MAP_GENERATION.md), which describes how
rooms and tiles are generated, and
[riiablo-recovered-data.md](riiablo-recovered-data.md), which describes the
quest-log and dialogue metadata already exposed to Lua.

## Evidence policy

| Grade | Evidence | Appropriate use |
| --- | --- | --- |
| A | owned-asset observation or repeatable original-game trace | compatibility contract |
| B | D2MOO 1.10f quest/runtime source, preferably corroborated | implementation blueprint, with patch label |
| C | original manual, contemporary modding research, data tables | intended behavior and data meaning |
| D | current guide/wiki/forum report | player route, discovery hints, edge-case lead |

Forum and guide pages are useful indexes, not final authority. D2MOO itself
states that it targets 1.10f and aims to stay close to the original game. Its
quest sources are therefore the strongest currently accessible behavioral
reference, but uncertain decompiler names and TODOs remain hypotheses.

## Canonical player route

Arrows show ordinary physical traversal. Parenthesized branches are optional
for act completion. Town portals and waypoints are transport shortcuts and do
not replace the underlying world links.

### Act I — Rogue Encampment

```text
Rogue Encampment
  -> Blood Moor -> Cold Plains
       |             |-> Burial Grounds (Blood Raven; Crypt/Mausoleum branches)
       |             `-> Stony Field -> Underground Passage 1 -> Dark Wood
       |                                      |                    |
       `-> Den of Evil                        |                    `-> Tree of Inifuss
                                              `-> Underground Passage 2 (dead-end branch)

Dark Wood -> Black Marsh -> Tamoe Highland -> Monastery Gate
                 |              |
                 |              `-> Pit 1 -> Pit 2 (optional)
                 `-> Forgotten Tower -> Cellar 1..5 (Countess)

Monastery Gate -> Outer Cloister -> Barracks -> Jail 1..3
  -> Inner Cloister -> Cathedral -> Catacombs 1..4 (Andariel)

Stony Field --Cairn Stones portal--> Tristram (Cain)
```

| # | Quest | Start/discovery | Required world interaction | Resolution/reward | Act gate? |
| --- | --- | --- | --- | --- | --- |
| 1 | Den of Evil | Akara; also progresses through leaving town/entering the den | Enter the Den of Evil and reduce its monster population to zero; the last-kill path is an explicit quest event | Return to Akara for one skill point; the den receives its cleared visual state | No |
| 2 | Sisters' Burial Grounds | Kashya, or discovering Burial Grounds | Kill Blood Raven in the Burial Grounds | Return to Kashya; unlock free Rogue mercenary/recruitment service | No |
| 3 | The Search for Cain | Akara / Tree of Inifuss discovery | Use the Tree, obtain/interpret the scroll, activate the five Cairn Stones in the decoded order, enter Tristram, operate Cain's gibbet | Cain returns to town and identifies items for free; Akara gives the ring reward | No |
| 4 | The Forgotten Tower | Read the Moldy Tome in Stony Field, or discover the tower | Descend five cellar levels and kill the Countess | Quest completion plus Countess chest/drop behavior | No |
| 5 | Tools of the Trade | Charsi, or monastery progress | Find the Horadric Malus in the Barracks, pick it up, and return it | Charsi's one-time imbue service becomes available | No |
| 6 | Sisters to the Slaughter | Cain; monastery progression can expose it | Reach Catacombs 4 and kill Andariel | Speak to Warriv and travel east to Lut Gholein | **Yes: Andariel/act travel** |

Important topology rules: the Burial Grounds and Forgotten Tower are branches,
not links on the critical path. Tristram is reached by an activated quest portal,
not by an ordinary level warp. The Malus is a mandatory preset/object placement
inside a randomly composed Barracks. Andariel's chamber is a required Catacombs
set piece.

### Act II — Lut Gholein

```text
Lut Gholein
  |-> Sewers 1..3 (Radament)
  `-> Rocky Waste -> Dry Hills -> Far Oasis -> Lost City -> Valley of Snakes
          |             |            |            |             |
          |             |            |            |             `-> Claw Viper Temple 1..2
          |             |            |            `-> Ancient Tunnels (optional)
          |             |            `-> Maggot Lair 1..3
          |             `-> Halls of the Dead 1..3
          `-> Stony Tomb 1..2 (optional)

Palace Cellar 1..3 --portal--> Arcane Sanctuary
  --one of four arms--> Summoner platform --journal portal--> Canyon of the Magi
  -> seven tomb exteriors --correct tomb/orifice--> Tal Rasha's Chamber (Duriel)
```

| # | Quest | Start/discovery | Required world interaction | Resolution/reward | Act gate? |
| --- | --- | --- | --- | --- | --- |
| 1 | Radament's Lair | Atma | Descend to Sewers 3 and kill Radament; obtain the Book of Skill | Book grants one skill point; Atma causes vendor-price favor | No |
| 2 | The Horadric Staff | Cain / component discovery | Cube: Halls of the Dead 3; Staff of Kings: Maggot Lair 3; Viper Amulet: Claw Viper Temple 2; transmute staff + amulet | Insert completed staff into the correct tomb's orifice to open Tal Rasha's Chamber | **Part of tomb gate** |
| 3 | Tainted Sun | Triggered on entering Lost City when darkness begins | Destroy the Tainted Sun Altar in Claw Viper Temple 2 and collect the amulet | Sunlight returns; amulet doubles as staff component | **Feeds staff gate** |
| 4 | Arcane Sanctuary | Jerhyn/Drognan palace access sequence | Traverse Palace Cellars and enter the Arcane Sanctuary | Find the Summoner's platform; tightly coupled to quests 5 and 6 | **Access chain** |
| 5 | The Summoner | Discover/kill the Summoner | Kill him and read Horazon's Journal | Journal reveals the true tomb symbol and creates the Canyon portal | **Reveals tomb** |
| 6 | The Seven Tombs | Jerhyn; progresses with Journal/staff | Enter the marked tomb, use the orifice, kill Duriel, pass into Tyrael's chamber | Speak with Tyrael, Jerhyn, then Meshif; sail to Kurast | **Yes** |

Act II is a dependency graph, not six independent quests. The three staff
components may be discovered out of conversational order. Palace access is a
world-state gate. The Arcane Sanctuary is a fixed four-arm structure whose
Summoner arm varies. The Canyon contains seven tomb entrances; the journal's
symbol identifies the quest tomb, while the other six remain valid dungeons.

### Act III — Kurast Docks

```text
Kurast Docks -> Spider Forest -> Great Marsh? -> Flayer Jungle
                   |                              |
                   |-> Spider Cavern (Eye)        |-> Flayer Dungeon 1..3 (Brain)
                   |-> Arachnid Lair (optional)   `-> Swampy Pit 1..3 (optional)
                   `-> Jade Figurine world drop

Flayer Jungle -> Lower Kurast -> Kurast Bazaar -> Upper Kurast -> Kurast Causeway -> Travincal
                                    |                |                |
                                    |                |                `-> High Council; Compelling Orb
                                    |                `-> temples
                                    |-> Ruined Temple (Lam Esen's Tome)
                                    `-> Sewers 1 -> Sewers 2 (Heart)

Travincal --orb/stairs--> Durance of Hate 1..3 (Mephisto) --red portal--> Pandemonium Fortress
```

The Great Marsh can be required by a particular generated route or can be a
side branch when Spider Forest links directly to Flayer Jungle. Code and data
must represent connectivity, not assume one walkthrough's ordering.

| # | Quest | Start/discovery | Required world interaction | Resolution/reward | Act gate? |
| --- | --- | --- | --- | --- | --- |
| 1 | The Golden Bird | A qualifying early unique monster drops Jade Figurine | Bring figurine to Meshif, receive Golden Bird, give it to Alkor | Potion of Life permanently adds 20 life | No |
| 2 | Blade of the Old Religion | Hratli / Gidbinn discovery | Find and activate the Gidbinn altar in Flayer Jungle, kill its spawned unique group, recover blade | Return through Ormus/Asheara flow; rare ring and Iron Wolf availability | No |
| 3 | Khalim's Will | Cain | Eye in Spider Cavern; Brain in Flayer Dungeon 3; Heart in Kurast Sewers 2; Flail from Travincal Council; transmute all | Use Khalim's Will to destroy Compelling Orb and open Durance access | **Yes: Durance gate** |
| 4 | Lam Esen's Tome | Alkor | Retrieve tome from Ruined Temple in Kurast Bazaar | Return to Alkor for five stat points | No |
| 5 | The Blackened Temple | Ormus/Cain or Travincal entry | Kill the High Council in Travincal | Flail component and prerequisite world state for orb/Durance | **Part of Durance gate** |
| 6 | The Guardian | Ormus / Durance progress | Reach Durance 3 and kill Mephisto | Take the red portal to Act IV | **Yes** |

Act III contains two important opportunistic/item-driven quests. Golden Bird
begins from a drop and has a multi-NPC exchange; Gidbinn begins either from NPC
dialogue or world discovery. These are good tests that `accepted` cannot be a
universal prerequisite for item and object events.

### Act IV — Pandemonium Fortress

```text
Pandemonium Fortress -> Outer Steppes -> Plains of Despair (Izual)
  -> City of the Damned -> River of Flame
       |                       |-> Hellforge (Hephasto; Soulstone destruction)
       |                       `-> Chaos Sanctuary
       |                              -> five seals / three seal boss groups
       |                              -> Diablo
       `-> waypoint
```

| # | Quest | Start/discovery | Required world interaction | Resolution/reward | Act gate? |
| --- | --- | --- | --- | --- | --- |
| 1 | The Fallen Angel | Tyrael / Izual encounter | Kill Izual and speak to his spirit | Return to Tyrael for two skill points | No |
| 2 | Hell's Forge | Cain; requires Mephisto's Soulstone to execute | Kill Hephasto, take Hellforge Hammer, equip it, place Soulstone on forge, strike forge | Forge gem/rune reward | No |
| 3 | Terror's End | Tyrael / Chaos Sanctuary entry | Activate five seals and kill the three seal-triggered unique groups; Diablo then spawns | Kill Diablo; in expansion, Tyrael opens travel to Harrogath | **Yes** |

The Chaos Sanctuary gate is event-driven: operating seals can spawn groups and
the Diablo condition depends on the seal encounter state, not merely arrival in
the center. The Hellforge similarly requires an item-on-object interaction and
an equipped quest tool, not a generic object click.

### Act V — Harrogath

```text
Harrogath -> Bloody Foothills (Shenk) -> Frigid Highlands (15 captives)
  -> Arreat Plateau -> Crystalline Passage
       |                  |-> Frozen River (Anya)
       |                  `-> Glacial Trail -> Frozen Tundra -> Ancients' Way
       |                                           |              |-> Icy Cellar (optional)
       |                                           |              `-> Arreat Summit (Ancients)
       |                                           `-> Infernal Pit (optional)
       `-> Pit of Acheron (optional)

Anya's portal -> Nihlathak's Temple -> Halls of Anguish -> Halls of Pain -> Halls of Vaught

Arreat Summit -> Worldstone Keep 1..3 -> Throne of Destruction
  -> five waves -> Worldstone Chamber (Baal)
```

| # | Quest | Start/discovery | Required world interaction | Resolution/reward | Act gate? |
| --- | --- | --- | --- | --- | --- |
| 1 | Siege on Harrogath | Larzuk | Kill Shenk at the end of Bloody Foothills | Larzuk sockets one eligible item | No |
| 2 | Rescue on Mount Arreat | Qual-Kehk | Break five prison doors and rescue three barbarians at each in Frigid Highlands | Ral/Ort/Tal runes; barbarian mercenary service | No |
| 3 | Prison of Ice | Malah / Frozen River discovery | Kill Frozenstein, obtain Malah's potion, use it on frozen Anya | Anya returns; class-specific rare item and permanent resistance scroll | No, but opens quest 4 portal |
| 4 | Betrayal of Harrogath | Anya | Enter her portal, traverse Nihlathak's halls, kill Nihlathak | Anya personalizes one item | No |
| 5 | Rite of Passage | Qual-Kehk / Summit altar | Meet level requirement, activate altar, defeat all three Ancients in one valid encounter | Large experience award; passage to Worldstone Keep opens | **Yes: level and encounter gate** |
| 6 | Eve of Destruction | Ancients completion / Keep progress | Reach throne, defeat five scripted waves, follow Baal into Worldstone Chamber, kill him | Difficulty/game completion and ending sequence | **Yes** |

The Ancients encounter has reset/failure semantics (for example, leaving the
Summit during the fight) and a per-difficulty level requirement. The Throne is
a sequencer: five distinct waves must be spawned and cleared before Baal moves
to the chamber portal. These should be modeled as encounter controllers rather
than ordinary monster-count objectives.

## What “mandatory” means

The quest log's numerical order is presentation order, not a strict dependency
chain. A robust implementation should distinguish:

- **story mandatory**: defeating the act boss and executing the travel handoff;
- **world gate mandatory**: staff/orifice, palace access, Compelling Orb,
  Ancients' passage, seals/waves;
- **reward optional**: objective can be done or reward claimed later;
- **discoverable**: world/item event can begin progress before an NPC offer;
- **party completable**: eligible nearby/in-act party members may receive goal
  or reward-pending credit even if they did not deliver the final hit.

Player guides often compress these categories into “required” or “skippable.”
That is useful for routing but insufficient for runtime behavior.

## Map-generation implications

Quest logic consumes a world graph produced by DRLG, but also constrains that
graph. Each generated game must make all required actions reachable.

1. **Required preset placement.** Trees, altars, chests, stairs, orifices,
   forges, seals, prisons, boss arenas, and portals are authored DS1/object
   content placed into otherwise procedural levels.
2. **Branch identity matters.** Spider Cavern is not Arachnid Lair; the correct
   Tal Rasha tomb is one of seven; Ruined Temple is one of several Kurast
   temples. Display-name similarity cannot drive logic.
3. **Object state changes topology.** Cairn Stones create a portal, the Journal
   creates a portal, the staff/orifice opens a wall, the Compelling Orb opens
   Durance access, the Ancients unlock the Keep, and Baal creates/uses the final
   chamber transition.
4. **Generation and quest state interact.** Preset unit suppression, special
   rooms, spawn initialization, and object modes can depend on quest/game state.
5. **World coordinates are not quest identity.** A quest should target stable
   level/object/monster identifiers. Random room coordinates are resolved after
   generation.

## Sources

### Runtime and declarative data

- [D2MOO repository and version statement](https://github.com/ThePhrozenKeep/D2MOO/tree/3b21043b99e987bad41cf0f7b49f1f246db52d5c)
- [D2MOO quest IDs, quest-state slots, and link to Necrolis' original Phrozen Keep research](https://github.com/ThePhrozenKeep/D2MOO/blob/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Game/include/QUESTS/Quests.h)
- [D2MOO generic quest event dispatcher and state machinery](https://github.com/ThePhrozenKeep/D2MOO/blob/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Game/src/QUESTS/Quests.cpp)
- [D2MOO Act I quest implementations](https://github.com/ThePhrozenKeep/D2MOO/tree/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Game/src/QUESTS/ACT1)
- [D2MOO Act II quest implementations](https://github.com/ThePhrozenKeep/D2MOO/tree/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Game/src/QUESTS/ACT2)
- [D2MOO Act III quest implementations](https://github.com/ThePhrozenKeep/D2MOO/tree/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Game/src/QUESTS/ACT3)
- [D2MOO Act IV quest implementations](https://github.com/ThePhrozenKeep/D2MOO/tree/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Game/src/QUESTS/ACT4)
- [D2MOO Act V quest implementations](https://github.com/ThePhrozenKeep/D2MOO/tree/3b21043b99e987bad41cf0f7b49f1f246db52d5c/source/D2Game/src/QUESTS/ACT5)
- [1.13 `Levels.txt` field and row reference](https://github.com/fabd/diablo2/blob/master/code/d2_113_data/Levels.txt)
- Local recovered quest hierarchy: `internal/content/d2legacy/data/recovered/riiablo/quests.txt`.

### Contemporary/player-facing corroboration

- [Original Diablo II manual](https://ftp.blizzard.com/pub/misc/Diablo%20II%20Manual.pdf) — quest log, acts, towns, waypoints, and intended player-facing systems.
- [The Phrozen Keep](https://d2mods.info/forum/) — historical modding corpus; use individual claims only with corroboration.
- [Phrozen Keep quest-structure research cited by D2MOO](https://d2mods.info/forum/viewtopic.php?p=412899#p412899).
- [PureDiablo quest index](https://www.purediablo.com/diablo-2/quests) — player-facing sequence and rewards.
- [Diablo2.io quest index](https://diablo2.io/quests/) — independent per-quest route corroboration.
- [nokka/d2s quest save-layout notes](https://github.com/nokka/d2s#quests) — independent evidence that quest records repeat per difficulty and include act travel/NPC introductions.

## Known uncertainties and required probes

- Record exact per-quest bit transitions for every entry path: NPC offer,
  direct area discovery, item pickup, party completion, join-in-progress, and
  reward claim.
- Establish party eligibility radius/area rules and late-join behavior for each
  boss, unique, object, and wave encounter.
- Verify patch differences between 1.10f, 1.14d, and current Resurrected,
  especially act gates and classic-versus-expansion behavior.
- Trace exact object-mode sequences for Cairn Stones/gibbet, Horadric orifice,
  Tainted Sun altar, Compelling Orb, Hellforge, seals, prisons, Summit altar,
  and Worldstone transitions.
- Tie every logical target above to `Levels.txt`, `Objects.txt`, `MonStats.txt`,
  item codes, `LvlPrest.txt`, and DS1 preset IDs from owned data.
- Capture generated area graphs for representative seeds, including both Act
  III Great Marsh layouts and all Arcane/tomb configurations.
- Extract message IDs and NPC dialogue selection tables without treating text
  presentation as authoritative state logic.

