# NoDrop Outcome Session Verification Design

## Intent and boundary

Issue #373 asks whether eligible player and party context changes monster loot, not only the count handed to the loot roller. Earlier PRs cover the shared count and death-time NoDrop input. This slice proves that an authoritative monster death stores different rolled outcomes when party eligibility changes that input. It does not change loot behavior unless the test finds a defect inside this boundary.

## Approaches

1. **Chosen: two production sessions with synthetic treasure data.** Hold the game seed, two living players, generated monster, and command schedule fixed. Join the players into a party in one session. Assert both the death-time NoDrop count and the parsed durable `drops` result. This reaches the existing production seam without a second loot implementation.
2. A Lua treasure-class unit test can control a roll more directly, but existing tests already check adjusted weights. It would not prove that death passes the eligible count into the roller or stores its outcome.
3. Replacing Dark Magic's named loot stream with the original dropper seed would target exact seed parity. That is a separate compatibility decision and is outside this issue slice.

The two independent design reviews favored the production-session approach. It adds only test fixtures and an acceptance test if existing code satisfies the case.

## Test scenario

Use two fresh authoritative sessions with seed 314 and otherwise identical synthetic `generatedHostileRecords()` data. Override the test's treasure-class row to `Picks=1`, `NoDrop=3`, `Item1=rin`, `Prob1=1`. Spawn one monster with two living players, so its spawn player count is two in both sessions. Keep tick order and monster death actions the same. In one session leave the players unpartied; in the other invite and accept before Alice kills the monster.

The credited player's party context should yield `no_drop_player_count` 1 or 2. For the supplied treasure row, the recovered MSD-R22a conversion gives weights 3 or 1. The current first named loot-stream value under seed 314 is `8494684240895940149`, whose residues modulo 4 and 2 are both 1. The first case therefore consumes a NoDrop pick; the second selects `rin`. Parse the durable `d2legacy.monster.death.drops` JSON and assert an empty array versus one object with `code: rin`. Avoid assertions about quality, affixes, or the full JSON encoding. Keep one death per session so no earlier treasure draw advances the stream.

## Constraints and evidence

- The test crosses the authoritative session, room population, party, death, treasure-class, and durable death paths. A small test-local fixture may override the treasure row; product code and unrelated tests remain untouched unless the new assertion exposes a scoped defect.
- MSD-R21–R23 in `docs/specs/monster-scaling-by-players-and-difficulty.md` are static 1.13c evidence for the count, weight, and pick stages. This test proves current Dark Magic outcome wiring with synthetic inputs; it is not a 1.14d runtime observation or a probability estimate.
- Original MSD-R23 draws from `dropper.seed`. Current `treasure_class.lua` deliberately uses a named deterministic authority stream, as documented in `docs/research/ITEM_GENERATION_AND_LOOT.md`. Exact seed and draw-order compatibility remains a separate decision.
- The issue remains open for original-game observation, full loot outcomes, and other unverified scaling behavior. No original game data, captures, private saves, or credentials enter the commit.
