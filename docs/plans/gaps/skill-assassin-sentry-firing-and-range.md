# Gaps: skill-assassin-sentry-firing-and-range

Spec: [Assassin sentry firing and range](../../specs/skill-assassin-sentry-firing-and-range.md)

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|---|---|---|---|---|---|---|
| E01 | integration | SFR-Q01; SFR-R05/SFR-R06b; SFR-O01/SFR-O02 | First/post-shot due frame; local range/chance requests remain available | AI scheduling and skill-mode completion producers | 2026-10-15 reached-think/budget reread; reused 2026-10-12 summon context identifies requests only | Read scheduling/mode-end through event due writes, then matched observations |
| E02 | integration | SFR-Q02; SFR-R01g | AI101 to reached-think entry; supplied think contract unaffected | Dispatcher/table index consumer | 2026-10-15 image pointer scan finds0x10EDE0. Same-day review: that cell is the think field of AI table entry 101 (selector 0x9D0F0, installer 0x9D230, executor 0x9D3B0), but the MonStats index column, special-table override and executor event root are not anchored | Name the MonStats+0x1E column, resolve the special-table selector, and trace the event that reaches 0x9D3B0 |
| E03 | integration | SFR-Q03; SFR-R02/SFR-R06a | Original trap `calc4`/aip/Skill1 inputs; supplied arithmetic remains available | Exact required MonStats/Skills cells and their compiled correspondence | 2026-10-15 complete Skills loader file read (recorded with the launch spec's TD-SKILLS17); no compiled/runtime or MonStats correspondence | Bounded relevant cell reads plus prepared/runtime correspondence |
| E04 | outside-boundary | SFR-Q04; SFR-R04/SFR-R07b/SFR-R10 | none locally for supplied dynamic positions; other/static producers and extreme static overflow not established | Position producers outside supplied getter domain | 2026-10-15 reused NRF-R10a/b; static/other routes remain NRF limits | Owning position/path producer contract |
| E05 | behavior | SFR-Q05; SFR-R03a/SFR-R03b/SFR-R03c/SFR-R03d/SFR-R03g/SFR-R08/SFR-R09 | Full stored/room/chooser execution; supplied callback choices and reached draw remain available | Stored-mode producers/writes, room lists, class field, path helpers and collision | 2026-10-15 dispatcher/selector 6/7/class/chooser assessed; table cells and chooser image decoded; integration incomplete | Read owned producers through writes and resolve exact class-column identity |
| E06 | outside-boundary | SFR-Q06; SFR-R03e | none locally with supplied eligibilityA; actual producer/meaning remains unknown | Candidate eligibilityA producer | 2026-10-15 reused GDF-R03 closes state 146 invis; checked unavailable helper and neutral condition | Trace condition writers/eligibility owner |
| E07 | integration | SFR-Q07; SFR-R03f/SFR-R03g | Actual reach result; opposite local return polarities and direct candidate draw verified | Weapon/range/separation/collision producer chains | 2026-10-15 import and original-image branch correction [record](../../research/spec-evidence/skill-assassin-sentry-firing-and-range.md#state146-predicate-conflict-sfr-q07); full imported dependencies incomplete | Finish range/separation/collision contracts at original reach boundary |
| E08 | integration | SFR-Q08; SFR-R01b/SFR-R01f/SFR-R02 | Actual skill/owner/room/mode/event effects; supplied request/return stages remain available | Named dependency producer contracts | 2026-10-15 existing SCD/SCX/SLD/NRF limits reused; owner import paths assessed; PVP-R05a/R07 exact final-predicate evidence reused | Complete only the relevant owner contracts, preserving their input/output/effect gates |
| E09 | observation | SFR-Q09; SFR-O01/SFR-O02/SFR-O03; all behavioral rules | Runtime-confirmation claims only; static supplied assertions remain available | Original-game matching observations for every rule | 2026-10-15 reused pending recipes; no gameplay campaign performed | Controlled original 1.13c observation campaign with exact assertions/limits |

The 2026-10-15 restructure moved the launch-side rows (missile footprint, action-17 selector cells, shared speed, burst creation integration and the burst observation) to the [launch gap file](skill-assassin-sentry-missile-launch.md).

## Implementation gaps

The comparison below is retained from the 2026-10-12 record; no implementation comparison was performed by this migration. Its references describe that historical state and may require a fresh implementation review. Current original-behavior authority is the linked spec and evidence gates. Where the historical "Original behavior" column differs from the spec, the spec governs. The room-candidate cap rejects reduced distances above 48 (SFR-R03c). The stored-target shortcut persists across thinks (SFR-R03a, SFR-R03d), so "retarget every think" is not established. The rows on burst spread, missile lifetime/collision and shared speed moved to the [launch gap file](skill-assassin-sentry-missile-launch.md#implementation-gaps) on 2026-10-15.

Repo context (old spec wording): the Assassin trap family (skills 251, 256, 257, 261, 262, 266, 271, 272, 276, 277) is one decoder,
`trap.assassin-family`; PR #409 added it, PR #411 hardened lifecycle. Code: `lua/d2legacy/systems/trap_skill.lua`, `data/trap_skills.lua`;
tests `trap_skill_test.lua`; research `docs/research/SKILLS_STATES_AND_MISSILES.md` (trap section, "partial 1.14d fidelity").
Issue #418 (parents #408, #387). Old spec says all repo values below are "implemented assumptions", not verified behavior.

| # | Rule | Repo file/doc | Current repo behavior (implemented assumption) | Original behavior | Original status |
|---|---|---|---|---|---|
| G1 | SFR-R05 first shot | trap_skill.lua, data/trap_skills.lua | First shot at `deploy tick + fire_interval`; `fire_interval = max(MonStats.aip4, 1)` | No interval timer in think code; first-think scheduling not found (Blade Sentinel do 44 schedules at frame+1) | unresolved (gated) |
| G2 | SFR-R06a aip meaning | same | `aip4` = interval between shots, `aip3` = range | `aip4` = active range; `aip3` = no-target/out-of-range delay; `aip2` = roll-fail delay; `aip1` = percent chance | verified in 1.13c code |
| G3 | SFR-R07a/SFR-R10 range metric | same | Target range `max(aip3,1)` as Euclidean world-unit distance | Octagonal `max(dx,dy)+min(dx,dy)/2` on path coords; fires when `< aip4`; hard candidate cap 49 | verified (metric), asm-only (unit) |
| G4 | SFR-R03/SFR-R08 target selection | same | Hostile, alive, same act and level; nearest first, tie by lowest id; one shot per cycle | Strict-nearest `<`, ties = first encountered; cap 49; stored-target shortcut; RNG skip for some player candidates; retarget every think | asm-only |
| G5 | SFR-R02 shot budget | same | Budget (`Param1` + synergy) spent only when a target found | Budget = lazy eval of Skills `calc4`, decremented immediately before a shot, 0 kills sentry; owner must exist/have room/not in town | verified in 1.13c code |
| G6 | tests | trap_skill_test.lua | Tests do not pin first-shot delay, range boundary, spread, burst count; fixtures bypass real decoder values | n/a | n/a |
| G7 | done-list | issue #418 | Confirmed rules to be applied to shared sentry/projectile paths with boundary tests and checkpoint/replay parity | n/a | n/a |
