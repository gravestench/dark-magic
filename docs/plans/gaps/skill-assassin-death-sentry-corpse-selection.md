# Gaps: skill-assassin-death-sentry-corpse-selection

Spec: [Death Sentry corpse selection](../../specs/skill-assassin-death-sentry-corpse-selection.md).

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|-----|------|---------|-------------------------------|------------------|-----------------------------------|----------------|
| AI entry | integration | DST-Q01, DST-R01, DST-R02, DST-R04, DST-R05 | DeathSentry AI entry to reached think; supplied think stages available | text/compiled AI selector/table chain | 2026-10-15: action 55 image anchor and named MonStats AI cell established; structural think match insufficient. Review byte search found the think pointer once, at D2Game RVA 0x10EE10. This is the think field of AI table entry 104 (indexer 0x9D0F0, installer 0x9D230, executor 0x9D3B0; see companion). The compiled AI index, special-table override and executor event root are unresolved | show that DeathSentry's compiled MonStats AI value is 104, resolve the special-table selector, and trace the event root of 0x9D3B0 |
| Competing corpse claim | behavior | DST-Q02, DST-R06b, DST-O01 | actual two-sentry claim outcome | think/mode/state/eligibility scheduling chain | 2026-10-15: source separates remembered id and later state request; no atomic outcome established | source order plus controlled two-sentry observation |
| Think producer integration | integration | DST-Q03, DST-R03b | actual target/skill/level/search/mode/delay effects; supplied consumers available | complete producer contracts and compiled text correspondence | 2026-10-15: targeted source/import/image reads and named cells identify precise calls; search callback reads corpse-valid (DSC-R01b, gated by DSC-Q04) then the unit-find predicate, whose dependencies are located but not complete; no six-rung absence claimed | complete the named producer paths and integration cases |
| Gameplay confirmation | observation | DST-O01 | runtime-confirmation claims only, except independently gated behavior above | observations of every behavioral rule | 2026-10-15: existing records checked; no gameplay campaign authorized/performed; recipes retained | controlled matching observations with actual inputs/profile/coverage |

Restructure (2026-10-15): these rows moved from the combined Death Sentry gap file when the spec was split; the think part of the former combined local-producer row is DST-Q03 above, the explosion part stays in the [corpse damage gap file](skill-assassin-death-sentry-corpse-damage.md). Attempted records keep their original dates.

## Implementation gaps

The following comparison is retained from the earlier gap file without a new implementation inspection. Its findings and evidence labels describe that comparison date; current original-game corrections and gates are in the spec/companion above. Only obsolete rule references were translated.

Repo side: family decoder `trap.assassin-family` and lifecycle hardening (PR #409, PR #411); code `lua/d2legacy/systems/trap_skill.lua`, `data/trap_skills.lua`, tests `trap_skill_test.lua`. Issue #420 (parents #408, #387). All repo values were "implemented assumptions", not verified behavior. Original column: all verified in 1.13c code (static) unless noted.

| Rule | Repo item | Current repo behavior | Original behavior | Original status |
|---|---|---|---|---|
| DST-R03a/DST-R04/DSC-R09a | Corpse selection | Corpse usable, in range (`Param3/Param4` radius, halved), nearest first, tie by id | Radius-10 search centered on nearest hostile target, first match, last-corpse id excluded, `maxDist/2` gate | verified |
| DST-R04 | Corpse path gating | Corpse costs one shot; gated like other shots | No chance roll, no `aip3/aip4` check on corpse branch | verified |
| DST-R04/DST-R06a | Consumption timing | Corpse consumed before damage resolved; no corpse falls back to projectile | Consumed by state 0x68 when explosion runs; selection not atomic | verified / race unresolved |
| tests | Tests do not pin corpse selection or rounding; fixtures bypass real decoder values | n/a | n/a | n/a |
