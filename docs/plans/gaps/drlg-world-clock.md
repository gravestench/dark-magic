# Gaps: drlg-world-clock

Spec: [World clock](../../specs/drlg-world-clock.md)
Evidence: [public companion](../../research/spec-evidence/drlg-world-clock.md)
Related gaps: [Tainted Sun quest](quest-tainted-sun.md)

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|---|---|---|---|---|---|---|
| WCK-E01 | integration | WCK-Q01; WCK-R01b; WCK-R01c; WCK-R05a; WCK-R05b | Complete darkness/colour result; supplied two-branch intensity arithmetic remains available | Presentation dependencies and applicable rendered observation | 2026-10-12 appearance pending; 2026-10-15 intensity/colour slices read, float/Act-IV/general-lighting paths excluded | Establish bounded presentation dependency and observe claimed cases without expanding this contract to general lighting |
| WCK-E02 | observation | WCK-Q03; WCK-R08b | Runtime confirmation of the client's World-clock handling only; quest message encodings are TSW-E03 | Valid message capture | 2026-10-15 source construction, row lengths and client assignment checked; no packet capture collected | Capture both message forms and eligible client handling with recorded inputs |
| WCK-E03 | outside-boundary | WCK-Q04; WCK-R09 | None locally; complete other consumers/stamina relationship unresolved | Rooted consumer search | 2026-10-12 getter leads reused; 2026-10-15 exact getter read supplies three outputs, no complete six-rung consumer search | Identify requested consumer through root and output, or complete a bounded search record |
| WCK-E04 | integration | WCK-Q02; WCK-R04; WCK-R08a | Client source production/outer cadence only; supplied mapping/period return remain available | Complete scheduling root and source semantic identity | 2026-10-15 original wrapper/caller/getter decoding closes supplied arithmetic; outer callers and source producers remain incomplete | Trace actual scheduling and level/source producers to wrapper |
| WCK-E05 | observation | WCK-Q03; WCK-V11; WCK-V12; WCK-V13; WCK-O01 | Runtime-confirmation claims only for exact frame, boundary, transition, wrap, recipient and divergence cases; quest observations are TSW-E06 | Attributable counted observations | 2026-10-12 idle day/table reads preserved narrowly; 2026-10-15 original resync source read, no gameplay campaign | Matching valid captures for named assertions, with inputs and dropped-record limits |
| WCK-E06 | integration | WCK-Q05; WCK-R01a; WCK-R07; WCK-R08b; WCK-R06b | Complete dependency-produced allocation/resolution/notification/presentation/delivery effects; supplied request order/local writes remain available. Quest-side callback, timer and drop effects are TSW-E07 | Independently verified dependency contracts and outer ordering | 2026-10-15 targeted reads expose crossings and preserve partial-failure state; 2026-10-15 follow-up re-read the resync body and named its remaining crossings (player lookup, notifications and delivery) | Establish each needed bounded dependency and applicable caller integration |

## Implementation gaps

Historical comparison recorded 2026-10-12, retained without a new implementation inspection. Its status/absence/source-identity statements describe that dated comparison; current behavioral authority and gates are the spec and Open evidence gaps above. Only the private source-project name in one absence cell has been replaced by a public description. The complete original bytes remain privately preserved.

Issue: #415 (Parent #346). This historical section remains input to a separately authorized future plan.

Restructure (2026-10-15): the dated table was divided by subject when the combined contract was split. Rows G1, G2, G3, G5, G6 and G9 (clock model, rates, act derivation, start effect and consumers) are here; rows G4, G7, G8, G10, G11 and G12 (quest trigger, persistence and plan decisions) are in the [Tainted Sun quest gap file](quest-tainted-sun.md#implementation-gaps). Row numbers and wording are unchanged; the "rule 3.x" labels are the dated comparison's own section references.

| # | Rule | Repo item (old wording) | Current repo behavior | Original behavior | Verified/gated |
|---|------|-------------------------|-----------------------|-------------------|----------------|
| G1 | Clock model | `lua/d2legacy/systems/environment.lua` (+ `environment_test.lua`), component `d2legacy.world.environment` (`act`, `cycle_index`, `period_of_day`, `ticks`, `time_rate`, `eclipse`), one per act, advanced in `pre_simulation` | Initial cycle 2, period 0, ticks 0, rate 128, eclipse=false. Three tables (normal, Act IV, eclipse); eclipse table used when `eclipse` true | Matches initial state and three tables (rules 3.1, 3.5) | verified |
| G2 | 3.7 advance rates | Same system: normal advance 1 tick; Act III night 11; Act IV 16 | Act III night +11; no +2 night rate for Acts I, II, V | Act III night +10 total; Acts I/II/V night +2 total; Act IV +16 (rule 3.6) | verified (code) |
| G3 | 3.7 act derivation | Environment component has `act` per component | Act is the env owner's act | Act derived from the caller unit's level id via thresholds 40/75/103/109 (table `0x6fde4668`), not from the env owner | verified |
| G5 | 3.1/3.4 eclipse flag setter | `eclipse` flag read by clock | Nothing in the repo ever sets it true | Start sets rate index 1 (rate 4), cycle 0, ticks 0, period 3, flag 1 | verified |
| G6 | 3.4 start effect | Dark Magic eclipse should set rate to 4 and reset cycle/ticks, not just flip the flag | Flipping flag only (if it were set) | Start jumps to fixed state | verified |
| G9 | Consumers 3.10 | `systems/maximum_stamina.lua` (time-based maximum stamina); `components/shared.lua` | Stamina reads the clock | Other consumers not found in the bounded original-code search | gated |
