# Foundation verification queue

These are the highest-value empirical probes extracted from the six foundation research documents. Keeping them in one place makes it easier to turn research unknowns into small tooling or owned-data PRs.

## P0: unblock architecture contracts

- Build an authoritative content-generation fingerprint and prove whether current live sessions can observe catalog invalidation/rebuild.
- Add exact D2 seed recurrence vectors and state serialization tests.
- Establish a lossless `.d2s` section scanner and byte-exact no-op round trip for a small owned-save corpus.
- Enumerate ItemStatCost op dependencies used by the mounted LoD data and identify unsupported forms.
- Extract all seven `CharStats` class rows into local/non-distributed golden vectors and determine fixed-point scaling for resource gains.

## P1: original-behavior traces

- Trace poison/regeneration/missile/AI timer boundaries and first-trigger/end-tick semantics.
- Produce one-field `.d2s` diffs for progression, hotkeys, difficulty, mercenary, quest, waypoint, and NPC-introduction fields.
- Trace two-hand/offhand/weapon-swap equipment legality and active stat sources.
- Trace durability/quantity/repair/replenish behavior.
- Trace level-up thresholds, multi-level XP grants, stat/skill point awards, and per-class life/mana/stamina rounding.

## P2: multiplayer/durability edges

- Define session content-fingerprint negotiation for simulation-affecting mods.
- Trace held item, corpse, trade, and service escrow behavior across disconnect/save/exit.
- Prove stale-revision rejection for realm/offline durable writes.
- Determine which temporary states/RNG data survive legacy save/exit versus only modern realm checkpointing.

Probe results should be folded back into the primary research document with a confidence upgrade and a reproducible test vector, not left only in issue/PR discussion.
