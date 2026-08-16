# Game data loading, linkage, localization, and modability

Status: implementation-oriented research baseline for expansion 1.14d. Current Dark Magic ownership is established; exact 1.14d TXT-to-BIN behavior and several linkage rules still require probes. Classic and earlier-patch compatibility are out of scope.

## Executive result

Dark Magic should keep its existing split:

```text
layered VFS bytes
  -> raw table document / provenance
  -> generic loss-preserving rows
  -> typed immutable catalog generation
  -> domain-specific indexes/views
  -> session-pinned authoritative data generation
```

The important remaining work is not another record manager. It is to make row identity, column preservation, cross-table linkage, simulation-generation pinning, and localization linkage explicit. Source-faithful rewriting and legacy BIN/tool interoperability are out of scope.

The current `internal/game/data/store` already reads original layered bytes, exposes streams, parses generic TSV, preserves duplicate/unnamed headers, reports source provenance, caches defensive copies, and invalidates per path. `internal/game/data/catalog` already builds a large typed snapshot atomically and derives deterministic indexes. [GAME_DATA_RECORDS.md](../GAME_DATA_RECORDS.md) correctly states that source bytes outrank documentation and that unused/sentinel/dangling data must not be silently repaired.

## Evidence and confidence

- **verified (Dark Magic)**: raw bytes and generic TSV share the layered VFS; typed snapshots rebuild atomically; source provenance is available.
- **verified (Dark Magic)**: current generic parsing keeps the first duplicate header at its authored name and suffixes later duplicates; unnamed headers receive synthetic names.
- **high**: Diablo II game rules are extensively data-driven by TXT/BIN tables and cross-table IDs; the bundled Data File Guide and D2MOO are the primary semantic references for each admitted table.
- **uncertain**: exact 1.14d compiler rules for every TXT -> BIN table, especially which string references become indexes/ordinals and which columns are discarded.
- **uncertain**: exact expansion 1.14d `-direct -txt` precedence and cache/recompile behavior.
- **uncertain**: which tables must be byte/semantic-identical across legacy client and server processes for all multiplayer modes.

## Current Dark Magic baseline

Relevant current owners:

- `internal/content`: deterministic layered directory/MPQ/ZIP/shim VFS.
- `internal/game/data/store`: raw table bytes, streams, generic TSV rows, provenance, cache invalidation.
- `internal/game/data/catalog`: typed generation and indexes.
- `internal/game/data/model`: surviving typed schemas.
- `internal/localization`: TBL-backed localization.
- `internal/game/session`: authoritative session lifetime and replay/checkpoint boundary.

Do **not** introduce a per-table service registry, table-specific global singleton, or gameplay-side ad hoc TXT parser.

## A table has several identities

A major compatibility risk is collapsing all identity to one Go map key. Keep these concepts separate:

1. **content path**: e.g. the logical VFS path that selected the table.
2. **source provenance**: which mounted layer supplied those bytes.
3. **generation identity**: a hash/version of the admitted content generation.
4. **authored row ordinal**: position in the source table, when legacy code or BIN data relies on row order.
5. **symbolic identifier/code**: `code`, `Id`, `skill`, `Treasure Class`, etc.
6. **numeric runtime identifier**: an explicit authored ID where the table provides one.
7. **derived index key**: a Dark Magic lookup convenience, not necessarily legacy identity.

A typed `map[code]record` is an index over a table; it must not erase the ordered record collection. This matters for tables where adjacent rows, numeric ordinals, redirect indexes, or act-local ordering are observable.

## Raw-table representation

The current generic `map[string]string` rows are useful but do not preserve original header order as a first-class structure. Raw bytes remain available, so no source information is irretrievably lost. If Dark Magic needs stronger diagnostics, introduce a renderer/gameplay-independent raw document shape roughly equivalent to:

```text
RawTableDocument
  logical path
  source layer + source path
  source hash
  ordered header cells
  ordered rows of ordered cells
  normalized header aliases used by generic lookup
  parse diagnostics
```

This is **not** required for ordinary typed gameplay reads. It is useful only when diagnostics must expose duplicate headers, unknown columns, column order, or trailing cells; source-faithful rewriting is not a product requirement.

The current generic parser fills missing cells with empty strings. Extra cells beyond the header are not represented in the row map. A raw document should preserve them and report them, even if typed decoding ignores them.

## Typed catalog generations

An authoritative gameplay session should consume a **pinned immutable game-data generation**. Hot reload can safely build a replacement generation for tools, menus, or future sessions, but a running deterministic session should not silently begin reading new Skills, ItemStatCost, TreasureClass, collision, or monster records halfway through a replay.

Recommended rule:

```text
session start
  -> capture GameDataGenerationID
  -> all authoritative domain views derive from that generation
  -> replay/checkpoint records include the generation ID
  -> incompatible live reload is refused/deferred for that session
```

Presentation-only data can have a different reload policy where doing so cannot affect authority. The classification should be explicit rather than inferred from filename.

A generation ID should be derived from the effective layered content, not from process time. At minimum it needs the canonical list of authoritative table paths plus content hashes and relevant parser/schema version.

## Cross-table linkage

Do not assume every apparent integer means the same thing. For every foreign-looking field document:

- source table and column;
- target table;
- whether the value is symbolic text, explicit numeric ID, row ordinal, act-local index, or bit-packed identifier;
- missing/sentinel value;
- whether the relationship is required or advisory;
- patch/expansion gating.

Build typed links as derived immutable indexes. Keep the authored source value available for diagnostics and mod compatibility.

Examples already visible in Dark Magic show why this matters:

- DS1 dynamic preset objects use act-local `MonPreset` ordering.
- sound redirects can require stable ordered `Sounds.txt` records.
- item/body/component codes are symbolic but ultimately drive runtime and save behavior.
- quest/dialogue recovered data uses logical identifiers that localization/audio resolve separately.

## Expansion, difficulty, and version gates

Do not delete rows merely because the current run is expansion-only or a row appears unused. Preserve rows and classify applicability in the consuming view.

A domain view may filter by:

- the fixed expansion ruleset where an authored expansion flag remains observable;
- difficulty;
- Ladder/non-Ladder;
- patch/content version;
- game mode.

The underlying generation remains a faithful view of the mounted data.

## Localization linkage

Localization is a separate content domain, not a field substitution step inside game-data loading.

Preferred boundary:

```text
typed gameplay record stores string key / logical ID
        |
        v
presentation asks localization capability
        |
        v
active TBL/string resources resolve player-visible text
```

Never bake localized display text into authoritative IDs. A save, replay, network command, item identity, or quest record should survive changing language.

Research still needed:

- precedence among `string.tbl`, expansion/patch string tables for each supported patch;
- numeric-string-index behavior where legacy formats store indexes instead of keys;
- duplicate/missing keys and locale fallbacks;
- Resurrected localization differences if supported later.

## TXT versus BIN

Dark Magic consumes the legally supplied expansion data needed by its own typed
runtime. Reproducing the legacy compiler, writing legacy BIN files, and
supporting old community-tool round trips are out of scope. Any future BIN
reader must be justified solely as an internal data-ingestion path and emit the
same canonical records as TXT; it is not an interoperability commitment.

## Mod override semantics

Current VFS precedence should remain the single source-selection mechanism. Domain loaders should not invent a second override system.

For mod-added records, define per table whether:

- new symbolic IDs are legal;
- numeric ID ranges are extendable;
- row order is semantically significant;
- save/network encoding can represent the new ID;
- old saves can load without the mod;
- client/server must share the exact extension.

A content mod that adds a new skill may use Dark Magic-native identifiers without
regard for vanilla save, packet, or community-tool encodings.

## Client/server agreement

For networked authoritative play, exchange or negotiate a **simulation content fingerprint** before character admission. At minimum include the rule/data generation that affects:

- skills/states/missiles;
- items/stats/recipes/drop tables;
- monsters/AI parameters;
- levels/collision/warps;
- quests/difficulty;
- character progression.

A mismatched client may still be able to render with different presentation assets, but it must not predict or submit semantics using an incompatible rule catalog without an explicit compatibility protocol.

## Failure and diagnostics

Structural problems that make a required table uninterpretable are fatal to the generation. Recoverable row/link issues should be source-located and retained.

Diagnostics should include:

```text
logical table path
source layer + source path
row ordinal
authored key/code if known
column
raw value
expected relationship/type
severity
compatibility classification
```

Do not “fix” dangling references, duplicate keys, sentinel rows, contradictory expansion flags, or patch leftovers in the loader.

## Implementation slices

1. **Generation fingerprint — implemented**
   - canonical hash of effective authoritative table bytes, winning provenance,
     and parser schema;
   - pinned through runtime-recipe v2 and every session identity surface.
2. **Raw ordered table document**
   - preserve ordered headers/cells/trailing values when diagnostics require them.
3. **Link registry/report**
   - document and validate per-table foreign-key semantics without making a runtime service locator.
4. **Session pinning — implemented for generic records**
   - copied table bytes back an immutable per-authority store, so invalidation
     cannot swap generations during a live replayable session.
5. **Localization linkage audit**
   - record key/index semantics and table precedence.

## Acceptance criteria

- Same effective mounted bytes produce the same generation fingerprint.
- Reordering an order-significant table changes the fingerprint and preserves row ordinals.
- Duplicate and unnamed headers remain inspectable; trailing cells are not lost in the raw representation.
- Typed indexes never become the only copy of ordered records.
- A live authoritative session cannot silently change simulation table generation.
- Invalid cross-links report source, row, column, and raw value.
- Mod overrides report which layer won.
- Localization can change without changing authoritative entity/item/quest IDs.
- Gameplay does not depend on legacy BIN writing or third-party tool compatibility.

## Verification backlog

1. Trace `-direct -txt` behavior on an owned expansion 1.14d install.
2. Diff representative TXT and generated BIN records to determine row-order/index compilation.
3. Identify tables where references are row ordinals rather than explicit IDs/codes.
4. Verify extra-cell behavior in shipped/modded TXT files and preserve it in raw tooling.
5. Establish exact localization TBL precedence and duplicate-key behavior.
6. Build client/server content-fingerprint experiments for a simulation-affecting one-row mod.
7. Verify expansion 1.14d table selection and treatment of authored expansion flags.

## Sources

- Dark Magic [GAME_DATA_RECORDS.md](../GAME_DATA_RECORDS.md).
- Dark Magic `internal/game/data/store/store.go` and `internal/game/data/catalog/catalog.go`.
- Bundled `docs/diabloiidatafileguide.html.gz` for table semantics, subordinate to mounted bytes.
- [D2MOO 1.10f](https://github.com/ThePhrozenKeep/D2MOO/tree/3b21043b99e987bad41cf0f7b49f1f246db52d5c) as secondary recovered implementation evidence only; it does not define the 1.14d target when behavior differs.
- [libd2 verification methodology](https://github.com/jaenster/libd2/blob/e6cdc4927c6180be8dd309b0423b470f64f1fc6c/docs/VERIFICATION.md) for retail-engine golden/holdout testing.
