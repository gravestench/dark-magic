# Character persistence and legacy save-format research

Status: implementation-oriented baseline. The current in-memory persistence store is intentionally not a `.d2s` schema. A loss-preserving legacy codec and a durable Dark Magic character model still need to be built.

## Executive result

Keep three layers separate:

```text
legacy .d2s bytes
   <-> loss-preserving LegacyCharacterFile codec
   <-> canonical DurableCharacter model
   <-> session admission / live ECS authority
```

A realm/offline storage policy sits beside the codec:

```text
DurableCharacter
   <-> atomic offline repository
   or
   <-> transactional realm repository
```

Do not make the legacy file layout the live ECS schema, and do not make modern realm transaction policy a side effect of a `.d2s` encoder.

Dark Magic already points in this direction: `internal/persistence.Character` explicitly says it is not a Diablo II save-file schema and that future importers should preserve unsupported bytes separately.

## Current Dark Magic baseline

The current store owns opaque character ID, name/class, level, expansion/hardcore flags, immutable appearance, a character-sheet `Stats` snapshot, and roster selection. `internal/game/player` copies an admitted subset into authoritative ECS through a system-only command. Once admitted, ECS is live gameplay authority.

## Independent legacy layout evidence

The MIT-licensed `nokka/d2s` parser documents a Lord of Destruction-era fixed header of 765 bytes followed by variable sections. Its independent layout includes:

- signature/version/filesize/checksum;
- active weapon set;
- 16-byte character-name field;
- status and progression;
- class and level;
- timestamp;
- 16 assigned/hotkey skill IDs and left/right skills for two weapon sets;
- menu appearance;
- difficulty bytes and a map-related field;
- mercenary metadata;
- per-difficulty quest blocks;
- waypoint blocks;
- NPC-introduction blocks;
- bit-packed stats;
- a 30-byte class-skill section;
- variable-length item lists;
- corpse items;
- expansion mercenary items;
- Necromancer iron-golem item.

Confidence: **medium** as an independent parser/document, not original-runtime authority. The exact layout must be verified against owned saves for every supported version.

The same parser documents attributes as 9-bit stat IDs followed by stat-specific bit widths, terminated by `0x1ff`, and item property lists with a similar `0x1ff` terminator pattern. This aligns conceptually with ItemStatCost-driven save metadata but must be validated field-by-field before Dark Magic writes saves.

## Loss-preserving codec model

A codec should preserve bytes it does not understand:

```text
LegacyCharacterFile
  version
  raw fixed header + parsed known fields
  raw/parsed quest blocks by difficulty
  raw/parsed waypoint blocks by difficulty
  raw/parsed NPC-introduction blocks
  raw/parsed stat bitstream
  raw/parsed skill block
  ordered item sections
  corpse section
  mercenary section
  golem section
  unknown/trailing sections
  original checksum/file-size fields
```

Every parsed field should retain source byte/bit range for diagnostics. An unmodified parse->serialize must be byte-exact for supported saves. A semantic edit should change only required fields plus derived size/checksum bytes.

## Canonical durable character model

The canonical model should use semantic fields and stable IDs rather than binary layout:

```text
DurableCharacter
  stable character ID
  name / class / edition flags
  level / experience
  base attributes
  unspent stat points
  unspent skill points
  learned skill base levels
  skill hotkeys / weapon-set assignments
  active weapon set
  durable appearance inputs
  difficulty unlock/progression/title state
  last durable location policy
  quest records per difficulty
  waypoints per difficulty
  NPC introduction/gossip flags per difficulty
  permanent quest rewards
  hireling identity/type/name/experience/death state
  corpse recovery state where compatible
  carried + stashed gold
  item archive covering all durable containers
  preserved legacy extension data
  content/schema compatibility metadata
```

Do not duplicate derived combat stats as durable truth unless the legacy format requires the value. Store canonical base/permanent state and reconstruct derived session stats from the pinned data generation.

Legacy saves may contain redundant fields, such as header level plus stat-section level. Import should validate/report disagreement while preserving both raw values; the canonical model chooses one semantic authority by documented rule.

## Per-difficulty domains

Quest, waypoint, NPC-introduction, and act/difficulty progression are not one global enum:

```text
DifficultyProgress[Normal|Nightmare|Hell]
  quest records
  waypoint bits
  NPC intro/gossip bits
  act travel/unlock
```

Existing quest research already shows why durable quest bits and live controller state must remain separate.

## Items and corpses

The item codec should be independent from the header codec but compose into it. A legacy item's bit length depends on flags/quality/socket/property data, so fixed-size item structs are not a valid canonical assumption.

Canonical item identity should survive equipment/inventory movement, weapon swap, death/corpse transfer, mercenary transfer, trade/vendor/service transactions, and save/reload. A legacy save adapter maps that identity/location to bitstream representation. Anti-duplication belongs to repository/session transaction policy, not the decoder.

## Checksum and corruption

Independent save documents agree that the fixed header carries file size/checksum near the beginning and that invalid checksum can invalidate a save. Exact checksum arithmetic must be verified against owned files before write compatibility is claimed.

Codec rules:

- validate signature/version/size before deep parse;
- compute checksum independently and report expected/actual;
- never repair a corrupt file merely by reading it;
- explicit write/repair recomputes derived fields after semantic edits;
- preserve a previous-good backup before first write of a user-owned legacy save.

## Offline repository policy

This is Dark Magic reliability policy, not a claim about original Diablo II write sequencing:

1. serialize to a temporary file in the same directory;
2. flush it;
3. optionally parse/validate the temporary result;
4. atomically replace/rename the destination;
5. retain a bounded previous-good backup;
6. where supported, flush directory metadata.

Use a per-character writer lock and revision/hash. A stale writer must not silently overwrite a newer revision.

## Realm/server persistence

Realm storage is revisioned semantic persistence:

```text
character ID
revision
schema/content compatibility
DurableCharacter
item identities
transaction/audit metadata
```

A server should serialize mutation of one durable character. Join loads revision R, validates content/schema, admits through system authority, and records a lease/session identity. A checkpoint builds a durable snapshot from authoritative owners and commits only if revision R is still current.

Disconnect/crash recovery must be idempotent. Item/trade/service escrow has an explicit recovery owner. A retry cannot duplicate an item or reward.

## Save triggers

Exact original save cadence is **unresolved** and mode-dependent. Dark Magic can define safe durability triggers such as explicit save/exit, successful act/difficulty transition, periodic realm checkpoint, clean disconnect, and shutdown. The semantic snapshot is independent of the trigger.

## Versioning and migration

Use separate versions for legacy `.d2s`, Dark Magic durable schema, item archive, quest record, realm repository, and simulation content fingerprint. Unknown legacy data remains attached so newer versions can reinterpret it later.

Classic -> Expansion conversion is a semantic operation, not “flip one status bit.” Research exact transformation before implementing it.

## Name and class validation

Current Dark Magic uses a 2-15-character ASCII policy with limited internal punctuation. Independent legacy documentation corroborates 16-byte storage and 2-15 visible length but disagrees on punctuation and does not establish all realm rules.

Separate binary encodability, legacy offline validation, legacy realm validation, and Dark Magic realm naming policy.

## Ownership boundary

| State | Owner |
| --- | --- |
| raw `.d2s` bytes and lossless metadata | save codec |
| selectable roster / durable semantic character | persistence repository |
| live vitals/movement/actions | authoritative session/ECS |
| item movement/escrow in session | item authority |
| quest live counters/timers | quest controllers |
| durable quest/waypoint/reward bits | durable character model |
| resolved visual appearance | derived presentation snapshot |
| realm revision/lock/audit | realm repository |

Lua receives defensive snapshots and opaque IDs, not file handles.

## Failure/recovery invariants

- Unsupported/corrupt input never partially overwrites the user's file.
- Unknown bits/sections survive no-op round trip.
- Failed semantic validation does not modify destination.
- Failed write leaves old valid file or complete new file.
- Stale realm revision fails instead of overwriting newer state.
- Repeated checkpoint transaction is idempotent.
- Exactly one durable location owns each item.
- Disconnect during trade/service has deterministic escrow recovery.
- A character cannot race into two mutually exclusive realm sessions.

## Implementation slices

1. lossless header/section scanner;
2. stat/skill bitstream codec driven by exact save metadata;
3. legacy item codec adapter;
4. canonical `DurableCharacter` expanding current roster metadata;
5. offline atomic repository with revision/backup;
6. explicit session export/import from ECS/item/quest owners;
7. realm repository contract with revisioned transactions/locks;
8. migration and Classic/Expansion lab after byte-exact round trips are green.

## Acceptance criteria

- Owned supported `.d2s` corpus round-trips byte-identically without edits.
- Unknown/reserved bytes and unmodeled item properties survive.
- Size/checksum verification catches corrupt fixtures.
- Canonical import->session->export preserves durable semantics.
- Live ECS does not mutate roster snapshots behind Lua.
- Offline write fault injection leaves valid prior/new file.
- Concurrent writers cannot both commit from the same revision.
- Per-difficulty quest/waypoint/NPC state remains distinct.
- Legacy and Dark Magic schema versions evolve independently.

## Verification backlog

1. Collect owned Classic/LoD saves from target patches.
2. Confirm fixed-header offsets/version ranges against bytes.
3. Verify checksum arithmetic with multiple unmodified saves.
4. Build one-field save diffs for class/level/flags/weapon/hotkeys/difficulty/map/mercenary.
5. Diff stat IDs/bit widths/ValShift/Add behavior using ItemStatCost.
6. Trace 30 class-skill bytes/global offsets for all classes.
7. Identify all item/corpse/merc/golem section markers and malformed behavior.
8. Verify quest/waypoint/NPC blocks one flag at a time.
9. Determine which temporary states survive save/exit.
10. Trace original offline save triggers/backup behavior.
11. Trace lawful legacy realm save cadence/locks/disconnect recovery.
12. Verify Classic -> Expansion conversion effects.

## Sources

- Dark Magic `internal/persistence/store.go` and `internal/game/player/player.go`.
- [QUEST_RUNTIME_MODEL.md](QUEST_RUNTIME_MODEL.md).
- [nokka/d2s README](https://github.com/nokka/d2s/blob/master/README.md), MIT-licensed independent parser/documentation.
- [nokka/d2s license](https://github.com/nokka/d2s/blob/master/LICENSE).
- [libd2 verification](https://github.com/jaenster/libd2/blob/e6cdc4927c6180be8dd309b0423b470f64f1fc6c/docs/VERIFICATION.md), which reports byte-exact `.d2s` round-trip testing.
