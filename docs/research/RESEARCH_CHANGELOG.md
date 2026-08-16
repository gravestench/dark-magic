# Systems research changelog

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
