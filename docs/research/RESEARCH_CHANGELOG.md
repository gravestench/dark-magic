# Systems research changelog

## Legacy patch-history baseline

Added the complete original Diablo II and Lord of Destruction version line from
retail 1.00 through legacy 1.14d, including the 1.04c PC hotfix, Macintosh
variants, 1.10 beta series, 1.10 Realm content activation, and 1.13 public-test
builds. The baseline separates executable patch, Classic/Expansion mode,
session type, platform, content era, and character origin.

Documented how the cumulative history constrains Expansion 1.14d feature
presence and removal, why patch notes cannot prove exact formulas or runtime
ordering, and how to combine them with mounted 1.14d data, D2MOO 1.10f,
libd2, and owned-runtime probes.

## Realm chat command baseline

Added a scoped command catalog for the realm lobby and in-game chat. It separates
editor shortcuts, realm social commands, active-game chat, local diagnostics,
and authoritative game-rule requests; records alias/version conflicts as probes;
and defines typed intent/event, security, moderation, and test boundaries.

Added Diablo Wiki/Fandom as an approved lead/corroboration resource while
preserving owned-client traces and applicable Blizzard documentation as stronger
evidence.

## Foundation baseline

Initial systems-research import from the maintainer's research handoff, reconciled against current Dark Magic architecture rather than copied as a speculative roadmap.

Added:

- complete 28-workstream index and dependency order;
- gameplay-system source matrix;
- game-data loading/linkage/modability baseline;
- timing/RNG/determinism baseline;
- character persistence / `.d2s` boundary baseline;
- item-stat/affix evaluation baseline;
- base item/equipment/inventory baseline;
- character progression baseline;
- Codex-oriented implementation handoff;
- consolidated empirical verification queue.

The next intended research tranche is combat/skills/monsters/owned-units/movement.
