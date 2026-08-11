# Codex handoff: foundation research conclusions

This checkpoint translates the foundation research into a small set of near-term engineering questions. It is not an instruction to implement all of them at once.

## Highest-leverage next code questions

1. **Pin authoritative game-data generations to sessions/replays.** The catalog is already atomic; the missing contract is preventing a live deterministic session from silently observing a replacement simulation-data generation.
2. **Make RNG/timer state explicit replay state.** The session already supports non-ECS `StateParticipant`s, which is the natural place for deterministic streams/schedulers that are not ECS components.
3. **Keep the save codec lossless and separate from `persistence.Character`.** The current persistence comments already establish this boundary. Start with byte-exact read/round-trip against owned saves before adding semantic writes.
4. **Introduce stat provenance before combat expands.** Current item property generation is strong, but combat needs layered/parameterized stat sources and invalidation rather than a flattened total map.
5. **Compose richer item instances around the existing item-placement authority.** Do not replace `internal/game/item`; add generation/condition/socket/corpse/trade semantics while preserving its atomic location model.
6. **Expand player progression from durable base facts.** Add base attributes, unspent points, purchased skill levels, and permanent reward sources before making character-sheet derived values authoritative.

## Things not to do

- do not add a second record loader beside `internal/game/data/store` + `catalog`;
- do not read wall clock/global randomness from authoritative gameplay;
- do not make `.d2s` structs the live ECS schema;
- do not flatten stat provenance into one mutable total map;
- do not move item legality into Lua/UI;
- do not hard-code seven classes' numeric progression when `CharStats`/`Experience` already exist;
- do not claim original compatibility for formulas/timing still listed in verification backlogs.

## Suggested next research PR

After this foundation PR is reviewed, research the tightly coupled combat tranche together but keep separate documents:

- `COMBAT_DAMAGE_AND_DEATH.md`
- `SKILLS_STATES_AND_MISSILES.md`
- `MONSTERS_SPAWNING_AND_AI.md`
- `HIRELINGS_AND_OWNED_UNITS.md`
- `MOVEMENT_COLLISION_AND_PATHFINDING.md`

That tranche should reuse the timing/stat/item/progression terminology established here.
