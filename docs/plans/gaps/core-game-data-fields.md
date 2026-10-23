# Game Data Fields — gaps

Behavioral spec: [Game Data Fields](../../specs/core-game-data-fields.md).
Evidence: [verification and attempts](../../research/spec-evidence/core-game-data-fields.md).

## Open evidence gaps

The 2026-10-15 [assessment](../../research/spec-evidence/core-game-data-fields.md#focused-refresh-and-affected-dependencies-2026-10-15) reuses the dated compatible source records below. Named text-cell/hash corroboration adds profile coverage; it does not close a compiler, producer or execution gap. Affects lists references; only the stated assertions are blocked.

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|-----|------|---------|-------------------------------|------------------|-----------------------------------|----------------|
| Mode-table direction-column defaults | behavior | GDF-R06b; owner question GDA-Q02 | absent MonMode direction-column defaults only; supplied direction fields and mode identities remain available | type-specific missing-field and callback outputs | 2026-10-14: Fog #10208 parser and #10207 second/missing-field switch read; record zeroing, present empty numeric 0 and nonzero flags proven; indirect and linked defaults incomplete; MonMode/PlrMode have only three text headers; 2026-10-15: existing partial reading reused; no identical default/lookup attempt repeated | trace each required default/indirect field with consumers and exact cases |
| Cached/text global monster correspondence | integration | GDF-R08; owner question GDA-Q03 | global equivalence of text cells, linked identity and cached record Id; the 13 cited monster identities remain available | explanation/ownership of 11 MonStats mismatches | 2026-10-14: complete text/compiled files read and ID correspondence checked; all 13 cited monster indices agree, 11 tail record Id values differ; duplicate cr_lancer8 is a lead only; 2026-10-15: all 11 recorded mismatches retained; no new compiled correspondence or origin history | trace preparation identity assignment and cached origin/live post-load fields, or constrain complete acceptance to a proven named profile |
| Attributable execution observations | observation | GDF-Q01; verified behavioral rules; GDF-O01 | runtime-confirmation claims only; no gate on verified supplied-input assertions | matching executions at their stated boundary | 2026-10-14: fresh Storm archive verification and historical memory scope audited; no gameplay driven or traces touched; old correspondence does not observe getter/compiler execution; 2026-10-15: named cells/hashes from coordinated loader reads add no getter/compiler execution; hooks untouched | separately authorized gameplay/initialization campaign with inputs, writes/returns, actual assertion coverage |

No implementation comparison was requested in this migration.

Restructure (2026-10-15): the combined gap file was split with the spec ([record](../../research/spec-evidence/core-game-data-fields.md#restructure-2026-10-15)). The first two rows track this contract's share of gaps owned by [Game Data Table Records](../../specs/core-game-data-table-records.md) questions GDA-Q02 and GDA-Q03, with their dated attempts copied unchanged; the observation row is this contract's own GDF-Q01. Ids were updated to the current numbering.

## Publication review (2026-10-23)

A fresh blind review of the split GDF dictionary found no new behavioral or identity gap. GDF-V03's all-zero monster direction inputs are explicitly synthetic; GDA-Q02 still owns actual absent-column defaults, GDA-Q03 owns global compiled MonStats correspondence, and GDF-Q01 owns pending execution observations. The sibling table-records publication has its own [issue #568](https://github.com/gravestench/dark-magic/issues/568). No 1.14d comparison was performed.
