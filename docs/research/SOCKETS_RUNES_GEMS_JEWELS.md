# Sockets, socket fillers, gems, runes, jewels, and runewords research

Status: implementation-oriented research baseline. The typed data catalog already loads Gems and RuneWords, and M6 can apply many Properties functions. Dark Magic does not yet have a complete first-class socket-container/runeword lifecycle in authoritative item state.

## Executive conclusion

Sockets are **nested authoritative item containment**, not a flat list of extra stats.

A socketed host item owns ordered inserted item identities:

```text
Host Item
  socket capacity
  ordered socket entries
      rune/gem/jewel item identity
      insertion index
  host ordinary stat sources
  inserted-item stat sources by host category
  optional runeword recognition/stat source
```

This preserves removal/destruction, save/network serialization, item identity, runeword recognition, and correct stat provenance.

Do not flatten inserted fillers into permanent host properties and then delete their identities.

## Socket capacity versus filled sockets

Keep distinct:

- number of sockets the host has;
- number of inserted items;
- exact insertion order;
- remaining empty sockets.

D2Common's runeword lookup explicitly compares the host's socket-count stat against the number of socketed inventory items before testing a rune sequence. A runeword therefore requires the relevant socket complement to be fully occupied in that path.

The legacy save format also stores a bounded socketed-item count and serializes child items separately.

## Host item as nested inventory

D2MOO uses an item inventory to contain socket fillers. Dark Magic does not need to copy that structure, but it should preserve the semantic relationship.

Suggested state:

```text
ItemSocketState
  Capacity int
  Inserted []ItemID   // ordered
```

Possible placement representation:

```text
ContainerSocket
HostItemID
SocketIndex
```

or an equivalent host-owned child collection. The requirements are stable child identity, deterministic order, one canonical owner, and atomic insertion/removal.

## Insertion is atomic

A socket insertion transaction should validate before mutating:

- host exists and is socketable;
- host has an empty socket;
- filler exists and is currently held/eligible;
- filler type is permitted;
- host/filler are owned by the acting authority as required;
- the operation is not blocked by trade/service state;
- stat/runeword recomputation can complete.

Then atomically:

```text
move filler under host
apply filler stat source
re-evaluate runeword
emit item-changed event
```

A replay/checkpoint must never expose "filler consumed but stats not applied".

## Socket filler property sets depend on host category

D2Common item-property code distinguishes property sets for socket fillers corresponding to weapon, helm/armor, and shield-like families.

The same gem/rune definition can therefore contribute different properties depending on the host.

Dark Magic should resolve:

```text
filler definition
+ host semantic category
-> socket property set
-> source-tagged effective stats
```

Do not mutate filler catalog rows.

## Gems

`Gems.txt` should supply the authored progression and three host-category property sets.

Gem upgrade recipes belong to the Cube engine. A gem inserted into a host remains the same item identity, now nested under the host.

## Runes

Runes are socket fillers with:

- ordinary per-host-category properties;
- ordered class/code identity used for runeword recognition;
- Cube upgrade recipes;
- content-version/ladder restrictions where applicable.

Runeword bonuses are **additional** to the individual rune properties. Model them as a separate conditional host stat source.

## Jewels

Jewels carry generated item properties/affixes. Their socket contribution should come from the jewel instance's rolled stat sources.

Preserving the jewel identity allows:

- exact rolls to survive save/load;
- unsocket recipes to destroy/remove the jewel correctly;
- trade/duplication validation to see real child identities;
- host stat diagnostics to attribute every contribution.

## Runeword recognition evidence

D2Common's `ITEMS_GetRunesTxtRecordFromItem` provides high-value 1.10f evidence.

Observed recognition shape:

1. the host is not a quest item in the shown path;
2. enumerate ordered socketed child item class IDs;
3. require socket count to equal inserted item count;
4. iterate complete RuneWords rows;
5. compare the authored rune sequence in order;
6. reject rows whose excluded item types match the host;
7. require at least one allowed item type match;
8. return the matching RuneWords record.

This is strong evidence that **sequence order and host type are authoritative**.

### Exact-length caution

Recognition should require the full authored sequence, not merely a matching prefix. Add explicit tests for:

- correct sequence and socket count;
- right runes, wrong order;
- correct prefix with extra filler;
- correct sequence in wrong host type;
- excluded host type;
- non-rune filler mixed into the sequence.

## Runeword identity and stat source

A recognized runeword should establish:

```text
RunewordState
  RuneWords row ID
  source host item ID
  generated/rolled runeword property instances
```

The host also carries a runeword flag/identity for presentation/save compatibility.

Do not rewrite the host quality into a new quality family. Runeword is orthogonal to Normal/Superior/Ethereal/etc. eligibility.

## Runeword invalidation

Any operation that changes socket contents must remove the prior runeword stat source and re-evaluate recognition.

That includes:

- insertion;
- unsocket recipe;
- host duplication/import repair;
- save load with malformed or changed child state;
- mod/content generation change where the recognized RuneWords row no longer exists.

Runtime recognition should be deterministic against a pinned content generation.

## Runeword base-item eligibility

Classic LoD runeword eligibility depends on more than "socketable".

Research/validation must pin:

- host item type inheritance;
- excluded types;
- quality restrictions;
- exact socket count/fullness;
- quest-item restrictions;
- expansion/ladder/version restrictions;
- whether Superior/Ethereal are allowed in the target patch;
- behavior with magic/rare/set/unique/crafted hosts.

Do not rely on UI tooltip folklore. Use table/runtime evidence and owned-game probes.

## Socket creation contexts are distinct

Sockets can arise from multiple sources:

- base item intrinsic maximum/capability;
- random generation;
- quest service such as Larzuk;
- Cube recipes;
- unique/set/other authored properties;
- direct/scripted item creation.

These contexts have different count rules. Do not centralize them into `AddRandomSockets()`.

Use a transaction that sets/changes host socket capacity with a source/reason.

## Random drop sockets

Pinned 1.10f ordinary generation evidence includes:

- difficulty caps in that path of 3/4/6 for Normal/Nightmare/Hell;
- force/suppress flags;
- a default 33% creation chance;
- newer-format count derived from item start seed modulo max plus one.

This is not evidence for Larzuk or Cube socket counts.

## Larzuk and quest-service socketing

Dark Magic already has an atomic server-owned `ServiceRule` framework. This is the correct authority shape for Larzuk:

```text
quest/service eligibility
+ target item identity
+ verified socket-count policy
-> atomic host mutation
+ quest completion/service consumption
```

The client names the semantic target/service only. It does not send the number of sockets to create.

Exact count by item quality/base/ilvl belongs in a dedicated probe.

## Unsocketing

Cube unsocket recipes can remove inserted socket fillers while preserving the host.

The transaction needs to specify whether fillers are destroyed or output separately according to the recipe. In classic Hel + TP unsocket semantics, the important architectural point is that child item identities can be removed independently from the host and all filler/runeword sources must be removed atomically.

## Save/network semantics

The durable semantic item format should serialize:

```text
host item
socket capacity
ordered child item IDs / nested child records
runeword identity if recognized
```

Legacy `.d2s` encoding can map that to its nested serialized item representation.

Do not make the runeword flag the sole truth. On import, recognition should be validated against socket contents while preserving opaque legacy data if needed for round-trip.

## Suggested implementation slices

### SCK1 — nested socket placement

Add ordered child item identities under a socketable host and atomic insert validation. No runeword behavior required yet.

### SCK2 — filler stat sources

Apply Gem/Rune/Jewel sources based on host category through the shared stat-source model.

### SCK3 — runeword recognizer

Normalize RuneWords data, recognize exact ordered sequences and allowed/excluded host types, and expose diagnostics.

### SCK4 — runeword stat source

Apply/remove runeword properties separately from individual filler properties.

### SCK5 — service/Cube integration

Implement one authoritative socket-creation service and one unsocket recipe through item authority.

## Verification backlog

1. Exact host quality restrictions for runewords in the targeted patch.
2. RuneWords `complete`, included types, excluded types, ladder/version semantics.
3. Exact sequence/full-socket recognition and malformed save behavior.
4. Gem property-set mapping to weapon/armor/helm/shield categories.
5. Jewel property behavior and restrictions when socketed.
6. Runeword property roll timing/seed and persistence.
7. Runeword stat-list flags/state and removal ordering.
8. Random drop socket count versus base max/ilvl/difficulty.
9. Larzuk socket count by base item, ilvl and quality.
10. Cube socket recipes and output count/randomness.
11. Socket capacity reductions/upgrades when an item is transformed.
12. Unsocket child destruction/output behavior.
13. Ethereal/Superior preservation through socketing/runewords.
14. Personalized item preservation through socket operations.
15. Durability/repair interactions with socket fillers/runewords.
16. Save/network nested-child ordering and exact legacy round-trip.

## Primary sources inspected

- D2MOO pinned 1.10f D2Common item/runeword lookup, item save paths, D2Game socket/item creation paths, and `Runes.txt`/`Gems.txt` table models.
- Current Dark Magic typed Gems/RuneWords data, item authority/service framework, and M6 property materialization.
