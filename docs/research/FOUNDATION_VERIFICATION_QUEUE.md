# Foundation verification queue

These are the highest-value empirical probes extracted from the six foundation research documents. Keeping them in one place makes it easier to turn research unknowns into small tooling or owned-data PRs.

## P0: unblock architecture contracts

- [x] Narrow runtime-recipe v2 game-data generation to effective authoritative
  table bytes and winning provenance. The session store copies those bytes;
  invalidation reparses the pinned generation, while source or mount changes
  produce a new generation only for a future authority. Parser schema and all
  compatibility surfaces are pinned.
- [x] Add a normalized cross-table linkage diagnostic fixture that distinguishes
  symbolic and numeric identity, preserves authored row/source/provenance
  context, and reports missing or ambiguous targets without repair.
- Add exact D2 seed recurrence vectors and state serialization tests.
- Enumerate ItemStatCost op dependencies used by the mounted LoD data and identify unsupported forms.
- Extract all seven `CharStats` class rows into local/non-distributed golden vectors and determine fixed-point scaling for resource gains.

## P1: original-behavior traces

- Trace poison/regeneration/missile/AI timer boundaries and first-trigger/end-tick semantics.
- Trace two-hand/offhand/weapon-swap equipment legality and active stat sources.
- Trace durability/quantity/repair/replenish behavior.
- Trace level-up thresholds, multi-level XP grants, stat/skill point awards, and per-class life/mana/stamina rounding.

## P2: multiplayer/durability edges

- Define session content-fingerprint negotiation for simulation-affecting mods.
- Trace held item, corpse, trade, and service escrow behavior across disconnect/save/exit.
- Prove stale-revision rejection for realm/offline durable writes.

Probe results should be folded back into the primary research document with a confidence upgrade and a reproducible test vector, not left only in issue/PR discussion.
