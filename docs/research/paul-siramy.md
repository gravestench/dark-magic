# Paul Siramy and Phrozen Keep research ledger

Paul Siramy's reverse-engineering work and tools are a primary research lead for
Dark Magic's legacy Diablo II world and animation support. Much of the original
site at `http://paul.siramy.free.fr/` is incomplete or only recoverable through
archival captures, mirrors, old forum posts, and redistributed source archives.

This ledger records sources and testable claims. It does not treat community
documentation as infallible: findings should be checked against legally supplied
game data and converted into synthetic regression fixtures whenever possible.

## Evidence labels

- **Documented**: explicitly described by Siramy or a contemporary tool author.
- **Observed**: confirmed against supplied Diablo II data or tool source.
- **Inferred**: a Dark Magic interpretation that still needs verification.

For every adopted rule, preserve the source URL, archived capture date where
available, affected game version, and a non-proprietary regression test.

## Source index

### Original site and DS1 editor

- Original site: <http://paul.siramy.free.fr/>
- DS1 documentation entry point:
  <http://paul.siramy.free.fr/_divers/ds1/doc/index.html>
- Win_DS1Edit distribution path cited by contemporary guides:
  <http://paul.siramy.free.fr/_divers/ds1/win_ds1edit.zip>
- A surviving forum reference links the DS1 documentation and identifies
  `DT1 Tools` as the conversion/editing toolchain:
  <https://www.diablofans.cz/forum/viewtopic.php?p=92197>
- A contemporary map tutorial records the DS1 editor distribution layout and
  its separate tile/object editing modes:
  <https://www.diablofans.cz/forum/viewtopic.php?p=92026>

Archive discovery should enumerate the original URL tree through the Internet
Archive CDX service rather than relying only on captures of the index page.

### Animation research

- Paul Siramy, *Extracting Diablo II Animations* (54-page PDF mirror):
  <https://tristram-archives.github.io/diablo2_infodump/2013/just%20hosting%20these%2C%20Downloaded%20from%20Internet/documentation/extracting_diablo_2_animations.pdf>
- MergeDCC binary and source listing:
  <https://www.allegro.cc/depot/MergeDCC/>
- SixDice release history, including Paul Siramy bug reports and concrete
  DC6/DCC edge cases:
  <https://www.bahj.com/sixdice/>

## Verified research queue

### DC6

- **Documented:** A transparent scanline run longer than 127 pixels must be
  emitted as multiple legal runs. SixDice 0.61 fixed an encoder that emitted
  these incorrectly; Diablo II's decoder could fail to display the result.
- **Current status:** the independent Go DC6 module decodes but does not expose
  a DC6 encoder, so this is an encoder acceptance requirement rather than a
  current encoding regression.
- **Fixture:** synthesize scanlines containing transparent runs of 127, 128,
  254, 255, and a run crossing an opaque segment; encode, decode, and compare
  indexed pixels and scanline termination.
- **Research:** resolve whether frame terminators are three or four bytes in
  each known asset family and distinguish the file header's four-byte
  termination marker from per-frame terminators.

### DCC

- **Documented:** DCC bitstreams can exceed 16 KiB; fixed-size intermediate
  buffers truncate valid assets.
- **Documented:** Some valid missile frames have height five, including cited
  directions of `arrow.dcc`; cell calculations must handle this boundary.
- **Documented:** A historical encoder could leak content from the previous
  frame's bottom line into the next frame.
- **Current status:** the independent Go DCC decoder sizes streams from encoded
  bit lengths and the public encoder returns `not yet implemented`. Decoder
  regressions should be tested now; encoder requirements belong to the future
  implementation specification.
- **Fixtures:** construct synthetic direction/frame cases for heights 1–8,
  streams around 16 KiB, transparent frames, and consecutive frames with
  deliberately distinct bottom rows.

### DS1 and DT1

- Recover the complete DS1 documentation tree, Win_DS1Edit source, DT1 Tools
  source, example INI files, changelogs, diagrams, and forum attachments.
- Record coordinate conversions among tiles, five-by-five subtiles, isometric
  pixels, object positions, paths, and editor coordinates.
- Verify layer ordering, wall orientation values, special tile identifiers,
  object/path encoding, groups, substitutions, unknown fields by version, and
  Act-specific behavior.
- Recover DT1 tile-index construction, orientation semantics, rarity/selection
  rules, subtile collision flags, material/sound fields, animation flags, and
  DS1-to-DT1 matching behavior.
- Connect observed runtime selection to `LvlTypes`, `LvlPrest`, `LvlMaze`,
  `LvlSub`, `LvlWarp`, and `Levels`, distinguishing documented rules from editor
  conveniences.
- Verify pop/poppad and logical/substitution behavior against real maps before
  implementing procedural level assembly.

## Adoption policy

1. Preserve the original and archived source URL.
2. Summarize rather than copying copyrighted documentation wholesale.
3. Confirm structural claims against supplied archives when possible.
4. Add malformed and boundary cases to the independent codec repository.
5. Keep engine rules in Dark Magic only when they describe composition or game
   behavior rather than the binary format itself.
6. Never commit Blizzard-owned decoded images, maps, animation bytes, or audio.
