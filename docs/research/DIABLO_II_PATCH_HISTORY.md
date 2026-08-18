# Diablo II legacy patch history and 1.14d inference guide

Status: historical research baseline for the original Diablo II executable
family. This document covers retail Diablo II, Diablo II: Lord of Destruction,
and their shared patch line through 1.14d. Diablo II: Resurrected is a separate
product and is not included.

Dark Magic still targets **Lord of Destruction, Expansion mode, patch 1.14d**.
The older versions below are evidence for how that target accumulated its
behavior. They are not additional compatibility targets.

## Executive conclusions

1. **Patch and game mode are separate axes.** Versions 1.00-1.06b predate Lord
   of Destruction and are necessarily Classic-only. From 1.07 onward, the same
   patched installation can run non-expansion (Classic/Standard) or Expansion
   games. Later notes frequently distinguish those modes.
2. **There is no complete standalone 1.14d gameplay specification.** The 1.14d
   note contains one crash-reporting change. Final behavior is cumulative:
   original retail behavior, the LoD 1.07 baseline, all applicable later
   patches, mounted final data, executable behavior, and Realm policy.
3. **1.10 is the largest architectural discontinuity.** It made skills and
   monsters substantially data-driven, added synergies, revised difficulty and
   experience, added Ladder and the World Event, and changed many item and
   combat rules. D2MOO's reconstructed 1.10f runtime is therefore highly useful
   structural evidence, but it cannot prove later 1.11-1.14d behavior.
4. **1.13c is the last documented broad gameplay/balance release.** The 1.13d
   release is mostly chat, exploit, aura, and crash corrections. The 1.14
   series is predominantly operating-system, installer, renderer, crash, and
   diagnostics work, with a few real gameplay fixes.
5. **Patch notes describe deltas, not exhaustive state.** They omit unchanged
   inherited behavior, undocumented fixes, data-only changes, server-only
   changes, implementation ordering, integer rounding, and some platform
   differences. Notes are discovery evidence, not executable truth.
6. **Old item and save state can be version-sensitive.** Several notes say
   changes are not retroactive or describe migration/cleanup. A 1.14d trace
   made with a character or item imported from another patch may not represent
   a freshly generated 1.14d state.

## Scope and terminology

- **Classic**, **Standard**, and **non-expansion** mean a Diablo II game or
  character that does not use the Lord of Destruction content mode.
- **Expansion** and **LoD** mean the Lord of Destruction mode.
- **Patch version** identifies the client/data generation. It must not be
  inferred solely from the character's Classic/Expansion flag.
- **1.10f** is commonly an executable/DLL build label used by
  reverse-engineering work. The public patch family is normally presented as
  **1.10**.
- **1.12a** appears in updater filenames and technical tooling, while the
  official notes normally call the public release **1.12**.
- Minor Realm hotfixes and content activation could occur without a new client
  version. Those need a separate content-era or Realm-policy identity.

The minimum compatibility identity for research artifacts should therefore be:

```text
product = diablo-ii-legacy
patch = 1.14d
mode = expansion | classic
session = single-player | open | realm
content-era = explicit when Realm/Ladder content matters
platform = windows | mac
character-origin = created-in-target | converted | imported
```

Dark Magic's current product target is
`diablo-ii-legacy / 1.14d / expansion`. Classic remains out of implementation
scope even though it is necessary historical context.

## Version and changelog inventory

Dates below are release/deployment dates found in contemporary reporting,
archived first-party pages, and the consolidated Basin Wiki history. Early
dates and 1.14c differ by a day in some sources because of street-date,
announcement, deployment, platform, or time-zone conventions. Unknown or
weakly established dates are left broad rather than falsely precise.

### Original Diablo II: Classic-only patches

| Version | Date | Material changes and research significance | Full notes |
| --- | --- | --- | --- |
| 1.00 | 28/29 Jun 2000 | Original retail baseline. There is no preceding patch delta. Establishing exact 1.00 behavior requires owned retail bytes or contemporary captures, not a changelog. | [Contemporary release report](https://www.gamespot.com/articles/diablo-ii-patched/1100-2595553/) |
| 1.01 | 27-29 Jun 2000 | Day-one crash, quest, skill, item, save-character, multiplayer, and Battle.net corrections. Some retail users may effectively have begun at 1.01 even though discs contain 1.00. | [1.01](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.01) |
| 1.02 | Jul 2000 | Additional stability, networking, interface, quest, monster, item, and skill corrections. | [1.02](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.02) |
| 1.03 | 4 Aug 2000 | Large early balance pass: class skills, vendor inventory behavior, hostility/multiplayer policy, monsters, items, crashes, and exploits. Contemporary reports note that some skill changes were absent from the packaged readme. | [1.03](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.03) |
| 1.04 | 20 Dec 2000 | Major corrective release: Cube-result persistence, previously unavailable base-item drops, bow enhanced damage and Guided Arrow, Duriel preload, death/corpse defects, Realm disconnections, skills, items, account/password features, and invalid-item cleanup. It also introduced regressions later repaired in 1.05. | [1.04](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.04) |
| 1.04b | 23 Dec 2000 | PC hotfix for Play CD access and Windows NT/2000 Video Tester crashes. | [1.04b](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.04b) |
| 1.04c | 24 Dec 2000 | Optional PC CD/DVD-access hotfix for machines still unable to start after 1.04b. This is a real distributed build but not a new cross-platform gameplay era. | [Contemporary 1.04c notes](https://diablo2.judgehype.com/patch104/) |
| 1.05 | 31 Jan 2001 | Repaired important 1.04 regressions, including reduced spell ranges and socket/gem value changes; also fixed character deletion, dual-potion Barbarian games, skills, quests, items, UI, and Realm behavior. | [1.05](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.05) |
| 1.05b | 2 Feb 2001 | Video/chat crash, copy-protection, stale channel-avatar, and character-select visual fixes. | [1.05b](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.05b) |
| 1.06 | 19 Apr 2001 | Network optimization, accessible Battle.net terms, and expanded CD/DVD/CD-R compatibility. | [1.06](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.06) |
| 1.06b | 17 May 2001 | Chat-spam crash, Open Battle.net character-save, Samsung-drive copy-protection, and terms-of-service changes. Last pre-expansion client patch. | [1.06b](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.06b) |

### Shared Classic and Lord of Destruction patch line

From this point forward, a version number alone is insufficient. The patched
installation may still run Classic characters/games, and some changelog entries
explicitly select Standard or Expansion behavior.

| Version | Date | Material changes and research significance | Full notes |
| --- | --- | --- | --- |
| 1.07 | 29 Jun 2001 | LoD retail baseline and first shared Classic/Expansion generation. Added expansion systems and made extensive skill, item, hireling, hostility, and balance changes. The notes explicitly preserve the old Blessed Hammer/Concentration interaction in non-expansion games while changing it in Expansion, proving that mode must be an independent rule input. | [1.07](https://theamazonbasin.com/wiki/index.php/Patch#Version_1.07) |
| 1.08 | 29 Jun 2001 | Immediate LoD update with extensive class, hireling, item, monster, quest, Realm, and balance changes. LoD 1.07 disc behavior and live 1.08 behavior must not be treated as identical. | [1.08](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.08) |
| 1.08b Mac | Jul 2001 | Archived Macintosh updater/build. Track as a platform variant unless executable comparison proves a distinct gameplay rule set. | [Archived patch listing](https://www.ausgamers.com/files/browsegame/html/3/9/downloads) |
| 1.09 | 20 Aug 2001 | Major Standard-and-LoD balance/fix release covering skills, hirelings, items and item generation, sets/uniques, runes, monsters, quests, experience, multiplayer, and exploits. It warns that some item changes do not apply retroactively. | [1.09](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.09) |
| 1.09b | 5 Sep 2001 | Focused bug, exploit, Realm, and item corrections. | [1.09b](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.09b) |
| 1.09c | 14 Nov 2001 | Additional server, crash, exploit, item, and gameplay fixes. | [1.09c](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.09c) |
| 1.09d | 21 Nov 2001 Mac; 5 Dec 2001 PC | Final 1.09-family crash, networking, and gameplay corrections. Platform release order differed. | [1.09d](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.09d) |
| 1.10 beta | 3 Jul-4 Aug 2003 | Public LoD-only test series. It deliberately excluded Classic, Battle.net Realm play, Ladder, and Open Battle.net. Beta values must not be used as final 1.10 facts. | [Official readme](https://classic.battle.net/diablo2exp/beta/readme.shtml), [official beta changes](https://classic.battle.net/diablo2exp/beta/patchchanges.shtml) |
| 1.10 final | 28 Oct 2003 | Largest legacy redesign: data-driven skill/monster systems, synergies, stat/inventory and collision optimization, harder Nightmare/Hell, experience and anti-power-leveling changes, new items/recipes/runewords, Ladder, World Event/Uber Diablo, broad class rebalance, and many fixes. Notes cover both Classic and Expansion where applicable. | [1.10](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.10) |
| 1.10 Realm update | 8 Jul 2004 | Server-side Ladder season 2 activation introduced Ladder-only runewords without a separately numbered client patch. Demonstrates that client version does not fully identify Realm content. | [1.10 history](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.10) |
| 1.11 | 1 Aug 2005 | Added Pandemonium Event/Uber Tristram, ten runewords, new special-boss uniques, hireling improvements, and many crash, exploit, gameplay, localization, and account fixes. | [1.11](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.11) |
| 1.11b | 13 Sep 2005 | Fixed a crash when returning to Battle.net chat and a crash when a hireling equipped the Peace runeword. | [1.11b](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.11b) |
| 1.12 / 1.12a | 17 Jun 2008 | Allowed no-CD operation when all required MPQs are installed and fixed Rosetta/OpenGL incompatibility on Intel Macs. `1.12a` occurs in updater/build naming; official prose generally says `1.12`. | [1.12](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.12) |
| 1.13a PTR | 10 Dec 2009 | First public-test iteration of the 1.13 work: respecs, balance experiments, skills, fixes, and new-content testing. Some secondary tables misprint this as 2010, which would place it after 1.13b and 1.13c. | [1.13a](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.13a), [combined 1.13 history](https://diablo2.diablowiki.net/Diablo_2_Patch_Notes_1.13) |
| 1.13b PTR | 23 Feb 2010 | Revised PTR balance and bug fixes; not a production Realm target. | [1.13b](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.13b) |
| 1.13c | 23 Mar 2010 | Production 1.13: respec Tokens/Essences and associated boss drops, orange important-item labels, flat gold-bank cap, removal of the Hardcore prerequisite, lower Fire Enchanted explosion damage, removal of Oblivion Knight Iron Maiden, skill/balance changes, and major dupe, TPPK, state, and aura-stack fixes. This is the last documented broad gameplay/balance release. | [1.13c](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.13c) |
| 1.13d | 27 Oct 2011 | Persistent ignores, message filters, home-channel commands, a dupe fix, unintended aura-stack fixes, correct multiple mercenary auras, and game-name/cinematic/windowed-mode fixes. | [1.13d](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.13d) |
| 1.14a | 10 Mar 2016 | Modern Windows compatibility, new Mac installer/support, and first-run save migration. It also represents substantial executable repackaging/refactoring even though documented gameplay changes are small. | [1.14a](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.14a) |
| 1.14b | 7 Apr 2016 | Restored Glide-wrapper loading, fixed mercenaries becoming `An Evil Force`, fixed a Mac Save & Exit crash, capped frame rate at 200, and supplied corrected German installers. | [1.14b](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.14b) |
| 1.14c | 17/18 May 2016 | Fixed three more Mac Save & Exit crashes and restored nGlide loading on PC. Sources differ by one day, likely announcement/deployment or time-zone convention. | [1.14c](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.14c) |
| 1.14d | 7 Jun 2016 | Added Blizzard Error and System Survey reporting for crashes and assertions. This is the final legacy patch, but its one-line delta must not be mistaken for the complete 1.14d ruleset. | [1.14d](https://theamazonbasin.com/wiki/index.php/Patch#Patch_1.14d) |

## Historical phases that matter to implementation

### 1.00-1.03: early Classic stabilization

The retail baseline changed quickly. Crashes, skills, vendors, quests, items,
hostility, and networking all changed within weeks. “Original Diablo II
behavior” is therefore ambiguous unless a patch is named.

### 1.04-1.06b: mature pre-expansion Classic

Patch 1.04 was broad and also introduced regressions repaired by 1.05. Patches
1.05b-1.06b concentrated increasingly on stability, networking, Battle.net,
and media compatibility. Patch 1.06b is the last purely pre-expansion target,
but it is not the only possible Classic target: post-1.07 clients retain
non-expansion mode.

### 1.07-1.09d: early LoD and dual-mode rules

The LoD launch created a shared executable with explicit Classic/Expansion
branches. Item changes could be non-retroactive, and disc 1.07 was followed
immediately by 1.08. Any research using “LoD launch behavior” must say whether
it means retail-disc 1.07 or deployed 1.08.

### 1.10: architectural and balance discontinuity

Patch 1.10 is the best explanation for why a reverse-engineered 1.10f runtime
remains structurally relevant to 1.14d. It introduced the data-driven systems
and many content structures that later patches inherited. It also means
pre-1.10 mechanics cannot be casually extrapolated forward.

### 1.11-1.13d: added endgame content and final broad balance state

Patch 1.11 added the Pandemonium Event and related content. Patch 1.12 mainly
changed distribution/CD requirements. Patch 1.13c supplied the last broad
balance and gameplay pass; 1.13d added communication features and important
targeted fixes.

### 1.14a-1.14d: platform modernization

The 1.14 series primarily modernized packaging, operating-system support,
render-wrapper loading, crash handling, and diagnostics. It did include at
least one visible gameplay fix (`An Evil Force` mercenaries), so it cannot be
modeled as a byte-identical wrapper around 1.13d.

## What the history lets us infer about Expansion 1.14d

### High-confidence inherited feature presence

Unless removed by later evidence, an Expansion 1.14d implementation should
expect the cumulative presence of at least:

- LoD classes, Act V, expansion items, hirelings, runes, jewels, charms,
  runewords, elite items, and Expansion difficulty/content branches from 1.07;
- the post-1.08 and post-1.09 skill, item, monster, hireling, quest, and
  experience corrections;
- 1.10's data-driven skill and monster structure, synergies, revised
  high-difficulty/experience rules, Ladder identity, and World Event support;
- 1.11's Pandemonium Event, new runewords, special-boss uniques, and hireling
  lifecycle changes;
- 1.12 no-CD/full-MPQ installation behavior where platform integration matters;
- 1.13c respecs, Essences/Token of Absolution, orange important-item labels,
  flat gold storage limit, current broad class balance, and the documented
  monster/skill/exploit corrections;
- 1.13d chat persistence/filter/home-channel surface and its aura/dupe fixes;
- 1.14 platform behavior and the corrected mercenary naming state.

This is a feature-presence map, not proof of exact arithmetic or sequencing.

### Facts that should be explicitly absent or changed

The cumulative notes also warn against accidentally reproducing older behavior
in the 1.14d target. Examples include:

- pre-1.10 skill progression without synergies;
- pre-1.10 experience/power-leveling rules;
- Oblivion Knights casting Iron Maiden after its documented 1.13c removal;
- the earlier Fire Enchanted explosion magnitude after 1.13c reduction;
- pre-fix aura stacking, TPPK, dupe, and active-state disconnect exploits;
- pre-1.13 Hardcore-character creation prerequisites;
- the pre-1.13 level-bound personal gold-bank cap;
- the pre-1.11 hireling purchase/experience policy;
- mercenaries entering the `An Evil Force` state fixed in 1.14b.

Each negative claim should still be checked against the applicable mode and
session type. A Realm exploit fix may not describe offline control flow, while
a data or common-runtime fix often does.

### What cannot be inferred safely from patch notes

Patch notes do **not** establish:

- exact integer/fixed-point arithmetic, overflow, rounding, or random-call
  order;
- frame timing, animation event frames, AI cadence, or collision order;
- the final contents of every TXT/BIN/MPQ record;
- complete save layouts or migration behavior;
- every undocumented bug fix or regression;
- whether a vaguely worded fix is client, common runtime, game server, Realm,
  or presentation behavior;
- final server-side Ladder eligibility, runeword activation, World Event
  counters, or season policy;
- whether an old item/save was rewritten or merely interpreted differently;
- the exact relationship between Windows and Mac binaries.

Those require mounted 1.14d data, owned-runtime probes, executable analysis, or
independent corroboration.

## Evidence policy for using the history

Use patch history in this order:

1. **Discover the relevant introduction/removal window.** For example, a
   respec implementation belongs to 1.13c-era evidence, not 1.10f evidence.
2. **Identify mode and session applicability.** Classic versus Expansion and
   single-player/Open/Realm must remain explicit.
3. **Inspect mounted Expansion 1.14d records.** These prove final authored data,
   not necessarily the consuming algorithm.
4. **Use recovered runtime source for structure.** D2MOO 1.10f is valuable when
   the subsystem was already present, but later deltas must be applied and
   conflicts must remain visible.
5. **Confirm exact behavior with an owned 1.14d runtime.** Record executable
   hash, platform, mode, session, difficulty, character origin, and inputs.
6. **Use at least one independent corroboration for consequential claims** when
   a retail probe cannot observe internal ordering.

A historical note can support statements such as “the feature exists by
1.14d” or “this old behavior was intentionally removed.” It should not alone
be promoted to an exact Dark Magic formula.

## Consequences for current Dark Magic research

### D2MOO 1.10f

D2MOO is strongest for architecture introduced or stabilized by 1.10: data
tables, behavior dispatch, stats, damage, skills, missiles, monsters, items,
quests, and game-server structure. Before adopting a 1.10f observation for
1.14d:

- check 1.11-1.14d notes for a relevant delta;
- compare the mounted 1.14d record shape and values;
- reject paths guarded for later content if the 1.10f build did not activate
  them;
- probe exact ordering and boundary values in the owned 1.14d runtime.

### D2MOO-to-1.14d delta-risk map

The changelogs provide a concrete audit queue for code or data recovered from
D2MOO. “Risk” here does not mean the recovered 1.10f implementation is wrong
for its own target. It means a later documented delta prevents us from assuming
it is also correct for Expansion 1.14d.

| D2MOO subsystem or rule family | Post-1.10 evidence creating risk | 1.14d research treatment |
| --- | --- | --- |
| Endgame portals, special bosses, keys/organs, and special unique drops | 1.11 introduced the Pandemonium Event and its rewards. | D2MOO's ordinary quest/Cube/portal machinery may be reusable architecture, but 1.11 encounter activation, recipes, drops, monster definitions, and Realm gates require later data or probes. Compile-time later-version hooks are not proof that a 1.10f runtime activated them. |
| Runeword catalog and availability | 1.11 added ten runewords; 1.10 Ladder season 2 also activated Ladder-only runewords server-side. | Use mounted 1.14d `Runes.txt` for definitions and maintain a separate content-era/Realm eligibility policy. Do not infer final availability from D2MOO's compile target or table parser alone. |
| Hireling purchase, leveling, experience, and equipment edge cases | 1.11 changed purchasable hireling level and experience gain; 1.11b fixed Peace-on-hireling crashes; 1.14b fixed the `An Evil Force` state. | Reuse generic ownership/AI structures cautiously, but treat hireling progression, equipment-triggered states, display identity, and lifecycle edge cases as later-version work. |
| Item transmutation and ethereal behavior | 1.11 fixed transmuted ethereal items losing their ethereal stat bonuses. | D2MOO Cube matching is structural evidence. Final output-copy/stat-regeneration ordering needs 1.14d data and a target probe. |
| Charged skills and synergy contribution | 1.11 says item skill charges no longer grant synergy bonuses to characters that do not own those skills. | Any D2MOO stat/skill dependency path that counts item-granted charges must be compared with this later restriction before adoption. |
| Set bonus source lifecycle | 1.11 fixed multi-piece bonuses persisting after death and removal. | Validate equipment/source teardown on death, corpse creation, recovery, and checkpoint restore against 1.14d; do not inherit 1.10f ordering without a probe. |
| Hostility and town-active offensive effects | 1.11 fixed waypoint timing around the hostility delay and Blessed Hammer remaining active in town; 1.13c fixed TPPK. | Party/hostility state shape may be reusable, but activation timers, town filtering, missiles already in flight, portals, waypoints, and hostile transitions are high-risk. |
| Diablo/Uber event locality | 1.11 fixed Uber Diablo dying when Shenk or Blood Raven died nearby. | Do not generalize quest-boss or nearby-death dispatch from 1.10f to special-event bosses without exact identity and event filters. |
| Respecs and associated Cube/drop content | 1.13c introduced the production respec system, Essences, and Token of Absolution after PTR revisions. | This functionality is absent from a faithful 1.10f baseline. Implement only from final 1.13c/1.14d records and observed transaction semantics; PTR values are not authoritative. |
| Skill formulas and class balance | 1.13a/b PTR and 1.13c document changes across multiple classes and skills. | D2MOO often remains useful for behavior-family dispatch, but every affected skill's parameters and sometimes its logic must come from mounted 1.14d data and exact probes. Never promote a 1.10f numeric constant because the function shape survived. |
| Monster curses and special-enchantment damage | 1.13c removed Oblivion Knight Iron Maiden and greatly reduced Fire Enchanted explosion damage. | AI spell selection and enchantment damage derived from 1.10f are specifically unsafe. Confirm both final records and consumer code paths. |
| Uber minion experience and Hellfire Torch proc behavior | 1.13c removed experience from Uber Mephisto/Baal minions and reduced the Torch Firestorm proc rate. | Later event monsters and item procs need 1.13c-or-newer evidence even if generic monster XP and item-event dispatch are structurally shared. |
| Gold storage and Hardcore creation gates | 1.13c made the gold-bank cap flat rather than level-bound and removed the prerequisite for creating Hardcore characters. | D2MOO progression/account rules in these areas are historical evidence only, not final policy. |
| Aura ownership, stacking, and mercenary multi-aura behavior | 1.13c and 1.13d fixed multiple unintended aura stacks; 1.13d also fixed mercenaries failing to have multiple auras when they should. | D2MOO aura/stat-list architecture is useful, but arbitration, source identity, replacement, coexistence, equipment switching, death, and reactivation require 1.14d-specific vectors. |
| Important-item presentation | 1.13c made runes and Pandemonium artifacts orange. | D2MOO server/item identity cannot by itself establish final client label/color rules. Keep semantic item importance separate from presentation. |
| Chat ignores, message filtering, and home channels | 1.13d added persistent ignores, content filters, and home-channel commands. | These are absent from a 1.10f feature surface and belong to Realm/client-social policy, not the gameplay simulation core. |
| Save locations, migration, platform startup, renderer loading, and diagnostics | 1.14a-d changed save migration, supported operating systems, Mac packaging, Glide/nGlide loading, crash handling, and reporting. | D2MOO is not evidence for final 1.14 host/platform integration. These changes normally should not leak into deterministic gameplay policy, but save discovery/migration must remain outside the simulation profile format. |

The same method works in the other direction. Where no later note mentions a
subsystem, D2MOO becomes a stronger candidate for structural continuity—but
silence is still not proof. Undocumented fixes and data changes remain
possible, and the recovered source can itself be incomplete. The graduation
rule remains: recovered 1.10f structure + mounted 1.14d authored data + owned
1.14d boundary observations.

### Mounted 1.14d MPQs and TXT/BIN records

Mounted final data is the strongest authored-content source. It can establish
record identity, parameters, references, strings, assets, and formulas stored
in data. It cannot by itself establish how the executable interprets every
field, when a consumer runs, or whether Realm policy overrides it.

### libd2 1.14d

The clean-room 1.14d implementation is valuable independent corroboration and
has a strong verification methodology. It remains implementation evidence, not
original-runtime authority. Conflicts should produce a targeted retail probe,
not an automatic choice of either implementation.

### Save and fixture provenance

Every compatibility fixture should say whether the character and relevant
items were:

- freshly created/generated in Expansion 1.14d;
- converted from Classic to Expansion;
- migrated from an older patch;
- imported from Open Battle.net or an editor.

Fresh target-created fixtures are the default for exact rules. Historical or
converted fixtures are separate migration/interoperability evidence.

## Recommended version-aware research fields

Future evidence records should use structured fields rather than embedding all
provenance in prose:

```json
{
  "product": "diablo-ii-legacy",
  "patch": "1.14d",
  "mode": "expansion",
  "session": "single-player",
  "platform": "windows",
  "content_era": "offline-1.14d",
  "character_origin": "created-in-target",
  "executable_sha256": "...",
  "data_fingerprint": "..."
}
```

For historical comparisons, change one dimension at a time. A useful minimum
set of boundaries is 1.06b Classic, 1.07 Expansion, 1.09d Expansion, 1.10
Classic and Expansion, 1.11b Expansion, 1.13c Expansion, and 1.14d Classic and
Expansion. Dark Magic does not need to implement those targets to use them as
diagnostic controls.

## Source assessment

### Strongest surviving public evidence

- [Blizzard's archived 1.10 beta readme](https://classic.battle.net/diablo2exp/beta/readme.shtml)
  and [beta changes](https://classic.battle.net/diablo2exp/beta/patchchanges.shtml)
  are first-party and clearly distinguish beta scope from final scope.
- The installed game's cumulative `Patch.txt`, when obtained from a lawful
  installation, is the preferred first-party text artifact for released patch
  notes. No such file is currently present in the inspected workspace roots.
- The [Basin Wiki consolidated patch history](https://theamazonbasin.com/wiki/index.php/Patch)
  is the most complete accessible index found. It reproduces the version
  history and links many contemporary announcements and Wayback captures.

### Corroborating chronology and edge cases

- [Codex Gamicus changelog index](https://gamicus.fandom.com/wiki/Diablo_II/Changelog)
  supplies separate pages for 1.01-1.12.
- [MobyGames Diablo II patches](https://www.mobygames.com/game/1878/diablo-ii/patches/)
  and [LoD patches](https://www.mobygames.com/game/4451/diablo-ii-lord-of-destruction/patches/)
  corroborate public release dates and platform packages.
- [JudgeHype's contemporary 1.04 series page](https://diablo2.judgehype.com/patch104/)
  preserves the easily missed 1.04c PC hotfix.
- [Diablo Wiki's combined 1.13 history](https://diablo2.diablowiki.net/Diablo_2_Patch_Notes_1.13)
  corroborates the 1.13 PTR-to-production sequence.
- [D2VersionChanger](https://github.com/ChaosMarc/D2VersionChanger) independently
  catalogs playable archival coverage from Classic 1.00 through LoD 1.14d. It
  is useful for version inventory, not behavioral authority.

### Repository/source inspection result

The four inspected workspace roots (`dark-magic`, `od2_codecs`, `akara`, and
`d2_english_mpq`) contain no complete historical `Patch.txt` or equivalent
release-note corpus. Dark Magic's existing documents consistently identify
Expansion 1.14d as the product target, D2MOO 1.10f as older reconstructed
runtime evidence, and libd2 as clean-room 1.14d corroboration. That evidence
model is consistent with the patch history, but it previously lacked one
central document explaining the cumulative-version relationship.

## Verification backlog

1. Extract and hash `Patch.txt` from the owned 1.14d installation, if present,
   and compare its headings and text to the consolidated public history.
2. Fingerprint owned Windows 1.14d executables and mounted MPQs in every probe.
3. Create fresh 1.14d Classic and Expansion characters to isolate mode branches
   without save migration.
4. Capture Classic-to-Expansion conversion before implementing conversion.
5. Build a data-table diff for legally obtained 1.10, 1.11b, 1.13c, and 1.14d
   records where the changelog indicates behavior changes.
6. Record documented feature introduction/removal version beside every
   patch-sensitive gameplay claim.
7. Separate Realm/Ladder content-era facts from offline client-version facts.
8. Treat 1.04c, 1.08b Mac, 1.10 betas, and 1.13 PTR builds as platform/test
   variants, not default production gameplay targets.
9. Resolve exact 1.14c deployment date only if build chronology becomes
   materially relevant.
10. Do not expand Dark Magic's compatibility scope until the Expansion 1.14d
    target is complete; use older versions as differential evidence only.
