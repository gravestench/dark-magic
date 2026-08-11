# Item, economy, quest-item, and gameplay-audio research handoff for Codex

Status: implementation handoff derived from the item/economy/audio gameplay research tranche. Read the linked research documents before broad implementation.

## Read first

1. [ITEM_GENERATION_AND_LOOT.md](ITEM_GENERATION_AND_LOOT.md)
2. [ITEM_QUALITIES_AND_SETS.md](ITEM_QUALITIES_AND_SETS.md)
3. [SOCKETS_RUNES_GEMS_JEWELS.md](SOCKETS_RUNES_GEMS_JEWELS.md)
4. [CHARMS_AND_CONTAINER_EFFECTS.md](CHARMS_AND_CONTAINER_EFFECTS.md)
5. [HORADRIC_CUBE_AND_RECIPES.md](HORADRIC_CUBE_AND_RECIPES.md)
6. [VENDORS_PRICING_AND_ECONOMY.md](VENDORS_PRICING_AND_ECONOMY.md)
7. [QUEST_ITEMS.md](QUEST_ITEMS.md)
8. [GAMEPLAY_AUDIO_BEHAVIOR.md](GAMEPLAY_AUDIO_BEHAVIOR.md)
9. [ITEM_SYSTEM.md](ITEM_SYSTEM.md)
10. [ITEM_STATS_AND_AFFIXES.md](ITEM_STATS_AND_AFFIXES.md)

## Current-engine boundaries to preserve

Do not replace these owners:

- `internal/game/loot` remains deterministic item/drop/quality/affix/property generation.
- `internal/game/item` remains authoritative item identity, container placement, vendor transfer, held state, service escrow and gold balances.
- `ContainerCube`, `ContainerVendor`, `ContainerQuest`, `ContainerHeld`, `ContainerHireling` remain semantic authoritative containers.
- `internal/game/session` remains fixed-tick command/replay/checkpoint authority.
- typed/generic game-data catalogs remain immutable content snapshots.
- `internal/audio` remains the mixer/backend command boundary.
- Lua remains presentation/intent and may display quotes/play presentation cues, but it must not own loot/economy/Cube/quest/audio-causing gameplay state.

## Highest-leverage implementation order

### 1. Rich generated-item recipe / instance state

Join M6 generation output to authoritative item identity with enough state for later systems:

```text
base code
generation source/context
item level
quality
concrete unique/set identity
selected affix/name IDs
rolled property sources
generation/item seed metadata
identified
ethereal
durability/max durability
quantity
socket capacity
personalization
```

Do not reroll on load/move/identify.

### 2. World loot materialization

Connect monster/chest/other loot output to stable `ContainerWorld` identity, legal collision-aware position, pickup/drop commands, ground lifetime and replay/checkpoint.

### 3. Conditional item source activation

Build one stat/effect source activation evaluator for:

- equipped items;
- active weapon set;
- broken/requirements suppression;
- inventory-only charms;
- socket child/host dependency;
- set threshold sources;
- recognized runewords.

Start simple and recompute owner sources after item transactions; optimize later only if needed.

### 4. Nested socket items

Represent inserted gem/rune/jewel identities as ordered children of the host. Add atomic insertion and filler property-source application.

### 5. Runeword recognition

Normalize RuneWords data and implement exact ordered rune sequence + host-type recognition. Keep individual rune stats and runeword stats as separate sources.

### 6. Set/unique property materialization

Turn M6-selected concrete special rows into durable rolled item sources. Add set partial/full conditional sources separately.

### 7. NoDrop and loot-context fidelity

Add verified effective-player/party NoDrop policy and generation-context quality semantics. Do not globally replace M6 quality order without distinguishing TC versus direct/vendor/Cube/gamble paths.

### 8. Cube recipe matcher

Add typed/normalized CubeMain preserving authored order. Implement deterministic multiset input matching over `ContainerCube` and one simple consume/create recipe.

### 9. Cube transformations

Add use-item/copy/upgrade/reroll/socket/unsocket/repair/recharge policies incrementally, preserving exact item identity fields according to verified outputs.

### 10. Vendor quotes and richer pricing

Separate quote from commit behind current buy/sell command authority. Retain current simple price formula until richer transaction-cost vectors are verified.

### 11. Vendor stock and gambling

Generate authoritative vendor stock through a dedicated generation context and refresh generation. Add per-player stable gambling offers; UI viewing must not reroll them.

### 12. Quest-item lifecycle events

Emit stable item pickup/drop/consume/transform events from item authority and connect one real quest-item flow to quest runtime/service escrow.

### 13. Lossless normalized sound definitions

Carry all relevant `Sounds.txt` metadata through `audio.Definition`/a new normalized layer, including pitch/fade/instance/falloff/tracking/priority/environment fields. Unsupported backend features stay inspectable.

### 14. Semantic gameplay audio cues

Bridge semantic simulation/world events to presentation audio:

```text
attack/hit/death
skill/missile
item pickup/drop/equip
object operation
quest event
zone transition
```

Do not let cue playback feed back into authority.

### 15. Zone/environment soundscape

Expose/read level `SoundEnv`, rain and inside/outside context, then implement music/ambient-scene crossfades and ambient-event scheduling on the presentation side.

### 16. Positional/monster/object audio

Add world emitter/listener attenuation/pan/tracking, MonSound/UMonSound roles, object loops/events, NPC speech and skill/missile travel/impact cues.

## Critical compatibility distinctions

### Quality order is context-sensitive

Do not "fix" M6 by applying one order everywhere.

Pinned 1.10f evidence shows:

- TreasureClass drop path: Unique -> Set -> Rare -> Magic -> Superior -> Normal;
- generic/direct item quality helper: Unique -> Rare -> Set -> Magic -> Superior -> Normal.

Preserve semantic generation context and verify callers.

### Socket capacity is not socket contents

Keep host capacity, ordered child identities, filler properties, and runeword source distinct.

### Item placement is not effect activation

An item may remain equipped/in inventory but have sources disabled by weapon set, broken state, requirements, or container policy.

### Vendor stock is real authoritative items

Do not render a catalog and mint a new unrelated item only after purchase unless the verified original context specifically behaves that way.

### Audio randomness must not consume gameplay RNG

Grouped grunt/ambient variation should use presentation/audio-specific deterministic seeds if reproducibility is desired.

## Things not to implement as shortcuts

Do not:

- copy rolled item properties into the player permanently;
- reroll items on identify/load/pickup;
- flatten socket fillers into host stats and delete them;
- use a runeword flag without validating child sequence/host type;
- sort Cube recipes and lose authored row order;
- let one Cube item satisfy multiple inputs accidentally via map iteration;
- implement Cube recipes as arbitrary Lua inventory mutation;
- trust client-supplied vendor cost/socket count/gamble outcome;
- use the simple current vendor formula as a compatibility claim;
- refresh vendor/gamble stock because a UI panel redraws;
- infer quest completion only from item possession;
- let ambient sound/music ownership leak across scenes through unmanaged handles;
- make audio playback a prerequisite for gameplay/animation effect timing.

## Cross-system event vocabulary

The item/economy/audio work benefits from stable semantic events such as:

```text
ItemGenerated
ItemMaterialized
ItemPickedUp
ItemDropped
ItemMoved
ItemSourceActivated
ItemSourceDeactivated
SocketInserted
SocketRemoved
RunewordRecognized
RunewordInvalidated
CubeTransmuted
VendorPurchase
VendorSale
ItemRepaired
ItemRecharged
GamblePurchased
QuestItemConsumed
AudioCueRequested
SoundscapeChanged
```

Exact API is not prescribed. Preserve causal IDs/ticks/actor/item/world context.

## Replay/checkpoint requirements

Authoritative future-affecting state includes:

- stable item identities and placements;
- complete rolled item-instance fields;
- nested socket child order;
- per-game unique occurrence state if enabled;
- vendor stock items/refresh generation;
- gamble offers/hidden seeds/context;
- Cube transaction RNG/results once admitted;
- carried/stashed gold;
- quest-service escrow;
- quest item identities/provenance;
- item durability/quantity/charges/replenish timers.

Audio mixer handles/voice phase are not authoritative gameplay state. Reconstruct persistent soundscapes from current scene/world context and semantic events.

## Recommended PR granularity

Prefer vertical PRs such as:

```text
loot result -> world item identity
rich item generation fields
conditional charm/equipment sources
nested sockets
runeword recognition
set conditional bonuses
NoDrop party policy
CubeMain parser/matcher
one Cube recipe
vendor quote API
one generated vendor stock
repair/recharge
one gamble offer flow
quest item lifecycle event
normalized Sounds metadata
semantic combat audio cue
zone ambience/music soundscape
```

## Definition of a successful item/economy vertical slice

A strong demonstration:

```text
monster death
 -> deterministic TC drop
 -> concrete rolled item recipe
 -> stable world item identity
 -> positional/audio drop cue
 -> pickup into inventory
 -> charm/equipment source activates only if placement policy permits
 -> vendor quote computed by authority
 -> sell moves the same item identity into vendor stock
 -> buyback returns the same identity
 -> replay/checkpoint reproduces state
```

A second strong slice:

```text
socketable host + runes in inventory
 -> insert exact ordered child identities
 -> individual rune sources apply
 -> exact runeword recognized
 -> runeword source applies
 -> Cube unsocket transaction removes child/runeword sources atomically
 -> item identity/personalization/ethereal state preserved
```

And for audio:

```text
enter generated Blood Moor
 -> authoritative zone snapshot exposes SoundEnv/rain/inside state
 -> presentation chooses soundscape
 -> streamed music/ambient scene crossfades
 -> intermittent ambient event uses audio-only deterministic seed
 -> nearby monster attack event emits positional MonSound cue
 -> none of those choices consume gameplay RNG or mutate simulation
```

## Verification discipline

Use [ITEM_ECONOMY_AUDIO_VERIFICATION_QUEUE.md](ITEM_ECONOMY_AUDIO_VERIFICATION_QUEUE.md). For each sensitive rule:

1. pin runtime/content version;
2. record owned inputs;
3. capture exact result/bytes/timing where possible;
4. keep one holdout vector;
5. distinguish original behavior from Dark Magic policy;
6. update the primary research document with confidence/probe result.
