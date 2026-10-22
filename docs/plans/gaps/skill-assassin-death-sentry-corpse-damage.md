# Gaps: skill-assassin-death-sentry-corpse-damage

Spec: [Death Sentry corpse damage](../../specs/skill-assassin-death-sentry-corpse-damage.md).

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|-----|------|---------|-------------------------------|------------------|-----------------------------------|----------------|
| Hit-time scale/credit | outside-boundary | DSC-Q01, DSC-R10b | none locally; actual sentry condition/credit producer, supplied PVP arithmetic available | actual condition and credit chain | 2026-10-15: current PVP contracts settle consumer, not flag production | verify producer/credit with caller controls |
| Complete downstream effects/draws | integration | DSC-Q02, DSC-R11b, DSC-R14b, DSC-O03 | registered callback effects/complete draw budget; local ordered calls available | event callbacks, registration and resulting writes/draws | 2026-10-15: DPO/IEP/HRC contracts assessed; argument 1 confirmed scale enable, no generic suppression proof | close relevant callback producer contracts and integrate |
| Actual fire application | integration | DSC-Q03, DSC-R12b, DSC-O02 | complete eligible hit/resource/status/reaction outcome; supplied DMI arithmetic available | actual eligibility/conditions/callback results | 2026-10-15: current DMI/PVP and V05 supplied caps resolve arithmetic only | integrate actual producers with DPO/DMI/HRC |
| Local producer integration | integration | DSC-Q04, DSC-R01b, DSC-R09b | actual preparation/skill/level/expression/mask/class/room/validity/update effects; supplied consumers available | complete producer contracts and compiled text correspondence | 2026-10-15: targeted source/import/image reads and named cells identify precise calls; current PVP own-array wrapper record resolves initial maximum-life array selection for supplied arrays; construction/other interiors remain; no six-rung absence claimed | complete the named producer paths and integration cases |
| Duration expression producers | outside-boundary | DSC-Q05, DSC-R13 | none locally; actual unresolved SCX parameter producers only | expression-specific producer contract | 2026-10-15: actual Common duration import/argument order and SLD-R11/R16/R17 close local source/routing | resolve only parameter operations requested by actual expression |
| Gameplay confirmation | observation | DSC-O01, DSC-O02, DSC-O03 | runtime-confirmation claims only, except independently gated behavior above | observations of every behavioral rule | 2026-10-15: existing records checked; no gameplay campaign authorized/performed; recipes retained | controlled matching observations with actual inputs/profile/coverage |

Restructure (2026-10-15): the AI-entry, competing-claim and think-producer rows, and the two-sentry observation, moved to the [corpse selection gap file](skill-assassin-death-sentry-corpse-selection.md) when the spec was split. Attempted records keep their original dates.

## Implementation gaps

The following comparison is retained from the earlier gap file without a new implementation inspection. Its findings and evidence labels describe that comparison date; current original-game corrections and gates are in the spec/companion above. Only obsolete rule references were translated. Corpse-selection rows are in the [corpse selection gap file](skill-assassin-death-sentry-corpse-selection.md#implementation-gaps).

Repo side: family decoder `trap.assassin-family` and lifecycle hardening (PR #409, PR #411); code `lua/d2legacy/systems/trap_skill.lua`, `data/trap_skills.lua`, tests `trap_skill_test.lua`. Issue #420 (parents #408, #387). All repo values were "implemented assumptions", not verified behavior. Original column: all verified in 1.13c code (static) unless noted.

| Rule | Repo item | Current repo behavior | Original behavior | Original status |
|---|---|---|---|---|
| DSC-R03/DSC-R04 | Roll bounds | Integer roll `Param1..Param2` percent of corpse max health | Roll excludes max (`min..max-1`) on corpse's seed | verified |
| DSC-R05 | Level scaling | Absent from assumptions | `dmg*casterLevel/corpseLevel` when caster lower | verified |
| DSC-R06/DSC-R08a | Fire/physical split | `fire = floor(amount*calc3/100)`, physical = remainder | Both parts truncated independently; physical `(100-pct)*dmg/100`, not remainder | verified |
| DSC-R08b | Units | Whole life points assumed | x256 fixed point; monster max life `(maxHP+minHP)<<7` | asm-only |
| DSC-R09a | Blast radius | Same radius as search radius; only hostile monsters hit | Blast from corpse, outer `(R+1)/2`, full `R/2`, physical 0 beyond `R/2` | verified |
| DSC-R10b | PvP | No player targets, no PvP path | Unresolved (modifier stage not read) | unresolved |
| kill credit | Kill credit | Goes to owner | Not determined | unresolved |
| tests | Tests do not pin corpse selection or rounding; fixtures bypass real decoder values | n/a | n/a | n/a |
