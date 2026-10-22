# Gaps: quest-tainted-sun

Spec: [Tainted Sun quest](../../specs/quest-tainted-sun.md)
Evidence: [public companion](../../research/spec-evidence/quest-tainted-sun.md)
Related gaps: [world clock](drlg-world-clock.md)

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|---|---|---|---|---|---|---|
| TSW-E01 | outside-boundary | TSW-Q01; TSW-R01c | None locally for supplied events; actual gameplay trigger production, including Lost City relationship, unresolved | Rooted event/selector producers | 2026-10-12 callback paths reused; 2026-10-15 initialization/event/auxiliary reading confirms bounded comparisons, no complete absence search | Trace actual producer to callback and output write |
| TSW-E02 | outside-boundary | TSW-Q02; TSW-R01b; TSW-R01d; TSW-R01e; TSW-R01i; TSW-R02b; TSW-R02c; TSW-R02d | None locally for supplied bits; gameplay meanings of flags/markers unresolved | Semantic consumers | 2026-10-15 matched bit wrappers/Fog masks establish test/set/clear and correct bit 14; do not establish quest meaning | Anchored readers/writers establishing each named role |
| TSW-E03 | observation | TSW-Q03; TSW-R04a; TSW-R04b | Runtime encoding confirmation of the quest's start/reset/failed-join messages only; World-clock client handling is WCK-E02 | Valid message capture | 2026-10-15 source construction, row lengths and client assignment checked; no packet capture collected | Capture both message forms and eligible client handling with recorded inputs |
| TSW-E04 | outside-boundary | TSW-Q04; TSW-R05b; TSW-R06 | None locally for supplied join flags; saved/repeated-difficulty flag production and universal clock-save absence unresolved | Complete save/load scope and repeat cases | 2026-10-12 join reading reused; 2026-10-15 assessment finds no six-rung save absence proof; no new save campaign | Bounded save-format contract and normal save/rejoin observations |
| TSW-E05 | integration | TSW-Q05; TSW-R02b; TSW-R02c; TSW-R07a; TSW-R07b; TSW-V07 | Complete reachable RNG budget; direct bodies and two supplied quest-owned draw sites remain available | Callback/item/gold/presentation/delivery helper reachability | Dated 2026-10-12 eight-body audit reused; 2026-10-15 inline/helper sites and owners checked; helper interiors excluded | Read all owned reachable draw paths or cite verified bounded dependencies |
| TSW-E06 | observation | WCK-Q03; TSW-O01; TSW-O02; TSW-O03 | Runtime-confirmation claims only for quest start, join, altar, recipient and divergence cases; frame/resync cases are WCK-E05 | Attributable counted observations | 2026-10-12 idle day/table reads preserved narrowly; 2026-10-15 original resync source read, no gameplay campaign | Matching valid captures for named assertions, with inputs and dropped-record limits |
| TSW-E07 | integration | TSW-Q06; TSW-R01a; TSW-R01d; TSW-R01e; TSW-R01f; TSW-R01g; TSW-R01h; TSW-R01i; TSW-R02b; TSW-R02c; TSW-R02d; TSW-R04a; TSW-R04b; TSW-R04c; TSW-R05a | Complete dependency-produced quest/timer/drop/iteration/delivery effects; supplied request order/local writes remain available. Clock-side resolution, notification and presentation effects are WCK-E06 | Independently verified dependency contracts and outer ordering | 2026-10-15 targeted reads expose crossings and preserve partial-failure state; timer insertion is read but elapsed units/delivery and helper interiors incomplete; 2026-10-15 follow-up re-read the animation, event, altar, sender and resync bodies and named every remaining crossing (mode-update, roster-update, iterations, notifications, drops, player lookup, diagnostic switch producer and broadcast delivery) | Establish each needed bounded dependency and applicable caller integration |
| TSW-E08 | observation | TSW-Q07; WCK-R01b; TSW-R04a; WCK-R06b | None locally: server cycle 0 and message cycle 5 are both exact; only the intent and the client-side reconciliation are open | Client clock after the start message and the first later resync | 2026-10-15 follow-up confirmed both values statically (start assignment, fixed sender payload); no capture | Observe TSW-O01 start with the client clock read before and after the first resync, or trace a rooted client consumer |

## Implementation gaps

Historical comparison recorded 2026-10-12, retained without a new implementation inspection. Its status/absence/source-identity statements describe that dated comparison; current behavioral authority and gates are the spec and Open evidence gaps above. Only the private source-project name in one absence cell has been replaced by a public description. The complete original bytes remain privately preserved.

Issue: #415 (Parent #346). This historical section remains input to a separately authorized future plan.

Restructure (2026-10-15): the dated table was divided by subject when the combined contract was split. Rows G4, G7, G8, G10, G11 and G12 (quest trigger, persistence and plan decisions) remain here; rows G1, G2, G3, G5, G6 and G9 (clock model, rates, act derivation, start effect and consumers) are in the [world clock gap file](drlg-world-clock.md#implementation-gaps). Row numbers and wording are unchanged; the "rule 3.x" labels are the dated comparison's own section references.

| # | Rule | Repo item (old wording) | Current repo behavior | Original behavior | Verified/gated |
|---|------|-------------------------|-----------------------|-------------------|----------------|
| G4 | 3.4 start event | `commands/quests.lua`: no eclipse reference; only quest event kind `quest_rewarded`; Tainted Sun is quest id 10 (`a2q3`) in recovered quest data | Nothing starts the eclipse | Start through quest-10 handlers (rule 3.3); trigger event unresolved | verified for call chain; trigger gated |
| G7 | 3.4 trigger | Research claim (`docs/research/QUESTS_AND_WORLD.md`, `docs/research/MOVEMENT_COLLISION_AND_PATHFINDING.md` world-clock section): Tainted Sun triggers on entering the Lost City | Research docs say Lost City entry | No level-id comparison in any a2q3 path read; contradicted | verified (static) |
| G8 | Persistence 3.9 | Flag lives on checkpointed environment component; parity test for normal cycle only; no eclipse tests | Eclipse flag checkpointed; no eclipse tests | Env not saved; rebuilt from quest bits on join | verified for code path; save format gated |
| G10 | Ownership | Owner candidates: `commands/quests.lua`, `systems/environment.lua`, or `d2legacy.quest.event` component as bridge | Open | n/a | gated (decision for plan) |
| G11 | Risk | Do not enable the flag until start trigger (3.1) resolved or a maintainer accepts "quest started" as trigger | n/a | n/a | gated (plan decision) |
| G12 | Process | Done checklist: runtime vectors for both boundaries and a reload case outstanding; #415 should link spec; implementation tests: events change Act II clock once, no repeat on retry, survive checkpoint/restore | n/a | n/a | gated |
