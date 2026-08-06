# Diablo II MPQ knowledge audit

This audit mines the local OpenDiablo2, AbyssEngine, and Riiablo repositories
for hard-won facts about Blizzard's released Diablo II data: paths, palettes,
frames, coordinates, labels, sounds, animation layers, and screen composition.

It deliberately does **not** propose carrying over their architecture,
functionality, or implementation. Dark Magic should implement its own behavior
through its Go runtime and Lua shim. These repositories are reference notebooks
and cross-checks for what the original assets mean and how they fit together.

## Audited snapshots

| Repository | Revision | Most useful knowledge |
| --- | --- | --- |
| OpenDiablo2 | `75b480a56a6a12ee1b62377bbed0d242e99aca37` | Go constants linking MPQ paths to palettes, frame indices, screen coordinates, localization keys, and screen composition |
| AbyssEngine | `885eea067f0d41a4fb75de2e88b3704215859a58` | Small, readable catalog of paths plus confirmed main-menu layering, tiling, blend modes, music, fonts, and video paths |
| Riiablo | `4eba03fc1c63dd76759a385a642dd484852bde9e` | Broadest asset catalog: front end, panels, inventory, skills, maps, composites, sound handles, and many exact DC6 frame uses |

The local OpenDiablo2 checkout has pre-existing uncommitted changes. This audit
is read-only and does not depend on them.

## What to extract

For every screen or panel, capture a data specification with:

- canonical MPQ path and case-insensitive normalized path;
- file type and any companion palette, PL2 transform, font TBL, or DC6;
- DC6 direction/frame meaning and animation timing;
- logical 640x480 or 800x600 coordinates, anchor, crop, and tiling;
- blend/draw mode and layer order;
- localized string key or numeric string ID;
- sound handle/path and the user interaction that triggers it;
- variants for classic/expansion, language, resolution, act, and class;
- confidence and agreement between reference engines.

The result should live as data consumed by shim Lua, not as Go constants.

## Shared path vocabulary

All three projects agree on the important path families:

| Purpose | MPQ path family |
| --- | --- |
| Front end | `data/global/ui/FrontEnd/` |
| Character list | `data/global/ui/CharSelect/` |
| In-game control panel | `data/global/ui/PANEL/` |
| Menus and quest/waypoint panels | `data/global/ui/MENU/` |
| Skill trees and icons | `data/global/ui/SPELLS/` |
| Cursor art | `data/global/ui/CURSOR/` |
| Localized UI art | `data/local/ui/{language}/` |
| Fonts | `data/local/FONT/{font-language}/` |
| String tables | `data/local/lng/{language}/` |
| Music | `data/global/music/` |
| Global/local SFX | `data/global/sfx/`, `data/local/sfx/` |
| Video | `data/local/video/` and language subdirectories |
| Game records | `data/global/excel/` |
| Palettes | `data/global/Palette/{palette}/pal.dat` |

MPQ lookup is case-insensitive and paths occur with `/` and `\` separators.
Dark Magic should normalize both without treating source spelling as a distinct
asset.

Known palette names are `act1` through `act5`, `endgame`, `endgame2`, `fechar`,
`loading`, `menu0` through `menu4`, `sky`, `static`, `trademark`, and `units`.

## Main menu

### Assets

- `data/global/ui/FrontEnd/gameselectscreenEXP.dc6` is the expansion game-select
  background used by OpenDiablo2 and AbyssEngine with the `sky` palette.
- Riiablo also identifies `data/global/ui/FrontEnd/TitleScreen.dc6`. This may be
  a different menu generation or classic/expansion variant and must be checked
  against the files rather than silently aliased.
- `D2logoBlackLeft.DC6` and `D2logoBlackRight.DC6` are the opaque/dark logo
  layers.
- `D2logoFireLeft.DC6` and `D2logoFireRight.DC6` are animated luminous/additive
  layers. Both halves share an anchor at `(400, 120)` in the 800x600 layout.
- AbyssEngine draws frame 0 of `gameselectscreenEXP.dc6` as a `4 x 3` tile grid
  from `(0, 0)`.
- `trademarkscreenEXP.dc6`, `TCPIPscreen.dc6`, and
  `CinematicsSelectionEXP.dc6` provide sibling front-end modes.
- Button sheets include `WideButtonBlank.dc6`, `3WideButtonBlank.dc6`,
  `MediumButtonBlank.dc6`, `CancelButtonBlank.dc6`, and
  `NarrowButtonBlank.dc6`. Riiablo confirms frame 0 as up and frame 1 as down
  for common wide/medium buttons.
- Popup sheets include `PopUpOkCancel.dc6`, `PopUpOkCancel2.dc6`, `PopUpOk.dc6`,
  `PopUpLarge.dc6`, `PopUpLargest.dc6`, `PopUpWide.dc6`, and
  `PopUp_340x224.dc6`; `fechar` is the recurring popup palette.

### Audio and text

- AbyssEngine and OpenDiablo2 use `data/global/music/introedit.wav` as title
  music. Riiablo's menu uses `data/global/music/Act4/diablo.wav`; this
  discrepancy is a verification target, not a choice to inherit blindly.
- Riiablo identifies `data/global/sfx/cursor/button.wav` and
  `data/global/sfx/cursor/select.wav` for button interaction.
- Front-end labels use Exocet and formal fonts. Common localized string IDs in
  Riiablo are `5106` single player, `5107` multiplayer, `5109` exit Diablo, and
  `5101` exit/back. Cross-check these against the loaded TBL rather than baking
  English text into Lua.

### 800x600 coordinates found in OpenDiablo2

| Element | Position |
| --- | --- |
| Logo halves | `(400, 120)` |
| Single player | `(264, 290)` |
| Multiplayer | `(264, 330)` |
| Credits | `(264, 505)` |
| Cinematics | `(401, 505)` |
| Exit Diablo | `(264, 535)` |
| Copyright lines | `(400, 500)`, `(400, 525)` |
| TCP/IP heading | `(400, 23)` |
| Host / join buttons | `(264, 200)`, `(264, 240)` |
| Server-IP popup | `(270, 175)` |

Some of these positions include OpenDiablo2-specific buttons. Mark every
element as original, inferred, or project-added when turning this into shim
data.

## Character creation/class selection

### Shared assets

- Background: `data/global/ui/FrontEnd/charactercreationscreenEXP.dc6`, `sky`.
- Campfire: `data/global/ui/FrontEnd/fire.DC6` at approximately `(380, 335)`.
- Each class has distinct unselected, hover, selected, forward-walk, and
  back-walk DC6 files below its FrontEnd class directory. Several classes also
  have `*s.DC6` overlay files.

The OpenDiablo2 resource catalog contains the full filename set for Amazon,
Sorceress, Necromancer, Paladin, Barbarian, Assassin, and Druid. Naming is not
uniform: Druid uses the `DZ` token, files vary in case, and not every class has
every overlay. Preserve the discovered filename rather than generating paths
from a supposedly universal pattern.

### 800x600 anchors and hit regions

| Class | Draw anchor | Approximate hit region `(x,y,w,h)` |
| --- | --- | --- |
| Amazon | `(100,339)` | `(70,220,55,200)` |
| Assassin | `(231,365)` | `(175,235,50,180)` |
| Necromancer | `(300,335)` | `(265,220,55,175)` |
| Barbarian | `(400,330)` | `(364,201,90,170)` |
| Paladin | `(521,338)` | `(490,210,65,180)` |
| Sorceress | `(626,352)` | `(580,240,65,160)` |
| Druid | `(720,370)` | `(680,220,70,195)` |

OpenDiablo2 also records approximate stance durations in milliseconds. These
are useful clues but should be verified from frame counts and the intended
front-end animation rate before becoming authoritative data.

Name/edit controls occupy the lower panel: name label `(321,475)`, textbox
`(318,493)`, expansion checkbox `(318,526)`, hardcore checkbox `(318,548)`,
exit `(33,537)`, and OK `(630,537)`.

## Saved-character selection

- Background: `data/global/ui/CharSelect/characterselectscreenEXP.dc6`, `sky`.
- Selection highlight: `charselectbox.dc6`, starting at `(37,86)`.
- Layout: two columns by four rows; each cell is `272 x 92`.
- Character image is offset `(-40,+50)` from the text anchor.
- Left text anchor starts `(115,100)`; right column uses `x=385`; rows advance
  by `95`; label lines advance by `15`.
- Scrollbar is `(586,87)` with height `369`.
- Buttons: new `(33,468)`, convert `(233,468)`, delete `(433,468)`, exit
  `(33,537)`, OK `(625,537)`.
- Delete confirmation uses `PopUpOKCancel.dc6` with `fechar` at `(270,175)`;
  cancel `(282,308)`, yes `(422,308)`, centered label near `(400,185)`.
- A repeated click inside approximately `1.25s` activates the selected entry in
  OpenDiablo2. Treat the threshold as a UX observation, not an MPQ fact.

## Loading screen

- OpenDiablo2 and AbyssEngine identify
  `data/global/ui/Loading/loadingscreen.dc6`.
- Riiablo identifies `data/local/ui/loadingscreen.dc6`. Verify whether both
  exist and whether selection differs by game version/language.
- Use the `loading` palette.
- Riiablo maps aggregate loading progress across the DC6 frames, suggesting the
  frames are progressive visual states rather than a time-based loop.

## In-game HUD

Core sheets:

- `data/global/ui/PANEL/800ctrlpnl7.dc6` — assembled control panel;
- `overlap.DC6` — globe overlap;
- `hlthmana.DC6` — health/mana numeric indicator;
- `level.DC6` — new-stat/new-skill buttons;
- `runbutton.dc6`, `menubutton.DC6`, `minipanel.DC6`, `minipanel_s.dc6`,
  `minipanelbtn.DC6`;
- `Skillicon.DC6` plus class sheets `AmSkillicon`, `BaSkillicon`,
  `DrSkillicon`, `AsSkillicon`, `NeSkillicon`, `PaSkillicon`, and
  `SoSkillicon`.

OpenDiablo2 assigns frames in `800ctrlpnl7.dc6`: health/status `0`, mana/status
`1`, stamina `2`, potions `3`, new-skills selector `4`, and right-globe holder
`5`. `overlap.DC6` frame `1` is the right globe overlap. These assignments
should be confirmed by rendering a contact sheet.

Useful 800x600 positions: run `(255,570)`, stamina `(273,572)` sized about
`102x19`, experience `(256,561)` sized about `120x4`, new stats `(206,561)`,
and new skills `(563,561)`.

`minipanelbtn.DC6` is a sequence of up/down pairs. Riiablo uses frames `0/1`,
`2/3`, ... through at least `16/17`; capture semantic button names by comparing
the panel code and rendered contact sheet.

## Inventory and character panels

- Main inventory: `data/global/ui/PANEL/invchar6.DC6`.
- Alternate weapon tabs: `invchar6Tab.DC6`, frames `0` and `1` for right/left.
- Exit/button sheet: `buysellbtn.DC6`, commonly frames `10/11` for up/down.
- Drop-gold button: `goldcoinbtn.dc6`, frames `0/1`.
- Slot overlays: `inv_armor.DC6`, `inv_belt.DC6`, `inv_boots.DC6`,
  `inv_helm_glove.DC6`, `inv_ring_amulet.DC6`, and `inv_weapons.DC6`.
- Hireling panel: `NpcInv.dc6` plus a subset of the same slot overlays.

The decisive geometry is not the hardcoded widget layout: it comes from
`data/global/excel/Inventory.txt`, including panel/grid bounds and each body
slot rectangle. Dark Magic should load those records and keep only sprite-local
pixel corrections as shim data.

## Other panels with strong asset clues

- Skills: `data/global/ui/SPELLS/skltree_{class}_back.DC6`; class backgrounds
  contain multiple tab frames. Skill icons use the class icon sheets and
  `SkillDesc.txt`'s `IconCel`, with adjacent normal/pressed or available-state
  frames observed in Riiablo.
- Escape menu localized art: `options.dc6`, `exit.dc6`, `returntogame.dc6`,
  `soundoptions.dc6`, `videoOptions.dc6`, `automapOptions.dc6`, and related
  option labels under `data/local/ui/{language}/`.
- Quest log: `questbackground.dc6`, `questdone.dc6`, `expquesttabs.dc6`,
  `questlast.dc6`, `questsockets.dc6`, and `a{act}q{quest}.dc6`. OpenDiablo2
  observes quest-icon frames `24` complete, `25` in progress, `26` not started;
  socket frames `0/1` normal/highlighted. Verify on contact sheets.
- Waypoints: `expwaygatetabs.dc6`, `waygatebackground.dc6`, and
  `waygateicons.dc6`.
- Cursor: `data/global/ui/CURSOR/ohand.DC6`; spinning pentagram:
  `pentspin.DC6`.
- Credits: `data/global/ui/CharSelect/creditsbckgexpand.dc6` and localized
  `data/local/ui/{language}/ExpansionCredits.txt`.

## Music and video path catalog

AbyssEngine's `src/common/ResourcePaths.h` provides a concise catalog of music
paths for every act, including town, wilderness, dungeon, action, and
resolution cues. Rather than copying those into Go, import them into a shim
manifest and cross-reference `Sounds.txt` / `SoundEnviron.txt` for the actual
selection rules.

Confirmed startup/video paths include:

- `data/local/video/New_Bliz640x480.bik`;
- `data/local/video/BlizNorth640x480.bik`;
- language-specific `d2intro640x292.bik`, `act02start640x292.bik`,
  `act03start640x292.bik`, `act04start640x292.bik`, `act04end640x292.bik`,
  `d2x_intro_640x292.bik`, and `d2x_out_640x292.bik`.

## Verification work before use

Reference projects disagree in places and sometimes contain project-specific
choices. Before declaring a fact authoritative:

1. Resolve the path against a user-provided Diablo II/LOD MPQ stack.
2. Decode the file and generate frame/direction contact sheets.
3. Record dimensions, offsets, palette, and transparent index.
4. Compare all three references and label disagreements.
5. Check `Sounds.txt`, `SoundEnviron.txt`, `Inventory.txt`, `SkillDesc.txt`, and
   localization TBL records before retaining hardcoded interpretations.
6. Store verified facts in versioned shim manifests with source/provenance and
   confidence fields.

The next practical deliverable is a read-only asset-catalog command that scans
the configured MPQs and emits this verified metadata. That converts forum and
reference-engine lore into reproducible Dark Magic data without importing any
reference implementation.

## First verification run

The initial catalog was run against the local English Diablo II/LOD MPQ set on
2026-08-05. All 90 curated hypotheses resolved, all 80 DC6 assets decoded with
their expected palettes, and no warnings were reported. Resolution by archive
was: 58 from `d2data.mpq`, 22 from `d2exp.mpq`, four patched TXT records from
`patch_d2.mpq`, three from `d2music.mpq`, two from `d2sfx.mpq`, and one from
`d2xmusic.mpq`.

Hash comparison clarified the notable path disagreements:

- `FrontEnd/TitleScreen.dc6` and `FrontEnd/gameselectscreenEXP.dc6` are
  distinct files despite having the same byte length and frame count.
- `data/local/ui/loadingscreen.dc6` and
  `data/global/ui/Loading/loadingscreen.dc6` are also distinct despite having
  the same byte length and frame count.
- `expquesttabs.dc6` and `expwaygatetabs.dc6` are byte-identical in this MPQ
  set; their separate names still carry different screen semantics.

The generated report and contact sheets are diagnostic build artifacts and are
not committed because they reproduce original game imagery. Regenerate them
from a user-owned installation with `cmd/asset_catalog`.
