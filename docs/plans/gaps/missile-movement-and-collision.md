# Gaps: missile-movement-and-collision

Spec: [Missile Movement and Collision](../../specs/missile-movement-and-collision.md)

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|-----|------|---------|-------------------------------|------------------|-----------------------------------|----------------|
| Final path conditions | integration | MMC-Q01,MMC-R05c | Later planner retention/final path conditions; initial local writes and supplied consumers available | Complete later planner/retention chain | 2026-10-15: Common allocation/type-selector listing and original entry 4 close initial recording/restricted writes; later planner effects remain incomplete | Trace later planner writes through final path output |
| Recurring queue/order | integration | MMC-Q02,MMC-R01 | Actual recurring dispatch and relative frame order; supplied update available | Full registration/requeue/consumer | 2026-10-15: MCI creation association and callback-free event walker read; order incomplete | Complete queue lifecycle and actual dispatch chain |
| Path/collision/removal | integration | MMC-Q03,MMC-R05a,MMC-R05b,MMC-R07,MMC-R09a,MMC-R09b,MMC-R15 | Actual path/room/cell/query/list/size/map/destruction and out-of-table behavior; supplied values/request logs remain available | Nested query shapes,map/room/reset/destruction callees | 2026-10-15: full local trace/search consumers and original-image/table checks; nested effects remain incomplete | Complete original nested call paths with exact output writes |
| Collision writer names | outside-boundary | MMC-Q04,MMC-R08 | none locally; complete map-writer meaning unknown | Terrain/player/monster writer chain | 2026-10-15: masks and MCI footprint requests compared; no full writer proof | Read each relevant collision-map producer |
| Gameplay confirmation | observation | MMC-Q05,MMC-O01,MMC-O02,MMC-O03 | Runtime-confirmation claims only; verified supplied subset remains eligible | Matching controlled 1.13c records | 2026-10-15: existing evidence assessed; no matching observations,gameplay outside migration | Run linked recipes plus remaining verified-rule observations |
| Pierce persistence | integration | MMC-Q06,MMC-R12b,MMC-V09 | Actual replacing-stat storage/propagation; count arithmetic/write requests available | RLI-Q03/STL-R06c completion | 2026-10-15: current shared writer/reader limits checked; no promotion from supplied persistence | Verify entry management/propagation/callback effects |
| Hand-off execution | integration | MMC-Q07,MMC-R14 | Actual to-hit/proc/handler/damage effects/draws; local order/result consumers available | Applicable SBA/IEP/MHD/individual handler evidence | 2026-10-15: direct contact calls and dependency contracts checked; interiors retain their own gaps | Close exact dependency assertions being consumed |
| Cooldown integration | integration | MMC-Q08,MMC-R16,MMC-V11 | Actual attachment/state storage/expiry/callback; allocation/request order available | STL attach/removal/expiry chain | 2026-10-15: full cooldown caller and null allocation branch read; shared lifecycle unresolved | Verify successful attach and timed removal/callback path |

## Restructure (2026-10-15)

De-duplication: the owner pierce-seed gap (owner own-base stat 328 producer) left this file
because the creation pierce calculation is now consumed by reference to MIR-R15a/MIR-R15b and
its producer gap is MIR-Q07. Former MMC-Q03..Q09 are now MMC-Q02..Q08 and the replacing
integration rule is MMC-R12b; no gap's evidence, kind or boundary changed.
