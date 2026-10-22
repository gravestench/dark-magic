# Gaps: skill-assassin-sentry-missile-launch

Spec: [Assassin sentry missile launch](../../specs/skill-assassin-sentry-missile-launch.md)

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|---|---|---|---|---|---|---|
| E01 | integration | STT-Q01; STT-R05/STT-R02; STT-O01 | Actual missile footprint/steering; local initialization requests available | Missile collision/path contracts | 2026-10-15 callback byte-equivalent C reconstruction and handler image establish path 10 request only | Read relevant path/collision handlers and authenticated Size cells |
| E02 | integration | STT-Q02; STT-R01 | Compiled/runtime selectors and original `calc1`/Vel inputs; supplied arithmetic and named text membership remain available | Compiled correspondence and exact required Skills/Missiles cells | 2026-10-15 complete Skills loader file closes text do 17 membership; no compiled/runtime correspondence | Bounded relevant cell reads plus prepared/runtime correspondence |
| E03 | integration | STT-Q03; STT-R03/STT-R04 | Actual speed output/creator/flight integration; STT-V04 supplied arithmetic available | Creator output binding and actual movement | 2026-10-15 speed flags/state/list region assessed; original 2026-10-12 limits retained | Read output through write and match travel with authenticated inputs |
| E04 | integration | STT-Q04; STT-R02/STT-R03 | Actual missile selection, target resolution, creation/path effects and the skill-mode-to-action route; supplied request/return stages remain available | Named dependency producer contracts | 2026-10-15 existing SCD/SCX/SLD/NRF limits reused; target/burst import paths assessed | Complete only the relevant owner contracts, preserving their input/output/effect gates |
| E05 | observation | STT-Q05; STT-O01; all behavioral rules | Runtime-confirmation claims only; static supplied assertions remain available | Original-game matching observations for every rule | 2026-10-15 reused pending recipe; no gameplay campaign performed | Controlled original 1.13c observation campaign with exact assertions/limits |

These rows moved from the [sentry firing and range gap file](skill-assassin-sentry-firing-and-range.md) in the 2026-10-15 restructure; their dated attempts are unchanged.

## Implementation gaps

The comparison below is retained from the 2026-10-12 record and moved from the [sentry firing and range gap file](skill-assassin-sentry-firing-and-range.md#implementation-gaps) on 2026-10-15; no implementation comparison was performed by that migration or this restructure. Its references describe that historical state and may require a fresh implementation review. Where the historical "Original behavior" column differs from the spec, the spec governs. The repository context and the test and done-list rows stay in that file.

| # | Rule | Repo file/doc | Current repo behavior (implemented assumption) | Original behavior | Original status |
|---|---|---|---|---|---|
| G1 | STT-R02 spread / burst | trap_skill.lua, data/trap_skills.lua | Fixed 5 degrees per step centered on target bearing; burst count `max(base + per_level*level + synergy/divisor, 1)` | No fixed spread; count = `Eval(calc1)`, charged-bolt path type, seed `pathX+index`, cap 77 frames; bolts steer | verified in 1.13c code |
| G2 | STT-R05/STT-R02 missile lifetime/collision | same | Lifetime `Missiles.Range` (Charged Bolt Sentry helper capped 77 ticks); collision radius `Size/2`; charged bolts fly straight | Cap 77 verified; collision size unresolved; bolts steer (path type 10) | cap verified; collision unresolved |
| G3 | STT-R03 shared missile speed | same | Speed `Vel * 3 / 64` per tick (from Lightning Sentry Vel 30, Charged Bolt Sentry Vel 12 observations; no acceleration/rank velocity/quantization/steering) | `CreateMissileUnit` 6fc8f930: `(VelLev*level)/8 + Vel`, `<<8`, `*0x4b/100`; not reconciled | asm-only, unresolved |
