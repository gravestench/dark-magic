# Gaps: missile-creation-and-seeding

Spec: [Missile Creation and Seeding](../../specs/missile-creation-and-seeding.md)

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|---|---|---|---|---|---|---|
| Adjacent collection producer | outside-boundary | MCI-Q01, MCI-R03, MCI-V03 | None locally; production/order of supplied room adjacency collection only | Room-generation ownership contract | 2026-10-15: matched original getter Common 3BA20 and Game 1280 selector read; stored order consumed unchanged, no sort/producer inferred | Owning room contract establishing generated collection/order |
| Neutral field meanings | outside-boundary | MCI-Q02, MCI-R09, MCI-R11, MIR-R01a, MIR-R06 | None locally; wider consumers of conditions/auxiliary/selector 11 remain unknown | Established semantic consumers | 2026-10-15: compatible direct writes reviewed; current sibling contracts do not supply wider meanings; neutral names retain supported values | Identify/read relevant consumers without guessing names |
| Client synchronization | integration | MCI-Q03, MCI-R12 | Actual recipients, cadence, full eligibility and delivery; creator update-mark request only | Full synchronization root and sender/network dependency contracts | 2026-10-15: Game 6F740 local sender and original 7B0A0 call at 7B348 checked; encoded call confirmed, full caller/delivery not established | Trace/read actual synchronization root through selected sender and delivery boundary |
| Gameplay confirmation | observation | MCI-Q04, MCI-O01, MCI-O02, MCI-O03 | Runtime-confirmation claims only; static supplied-input assertions remain available | Matching captures with actual inputs/assertions and build profile | 2026-10-12: no reachable member; 2026-10-15: historical observation families assessed and tool-neutral procedures retained, no gameplay campaign performed | Authorized original 1.13c gameplay captures matching named assertions |

Restructure (2026-10-15): the combined gap file was split by owning rule when the spec was divided. Rows for metadata, event, placement, path, collision, planner, damage, attribution, pierce, stat-lifecycle and monster-mod gaps moved to the [initialization and registry gap file](missile-initialization-and-registry.md); the gameplay-confirmation row was split between both files. Attempt records are unchanged.
