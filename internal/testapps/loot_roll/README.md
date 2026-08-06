# Deterministic loot roller

`loot_roll` is a headless test application for treasure-class selection. Its
input can be the game's tab-delimited `TreasureClass.txt` or
`TreasureClassEx.txt`. JSON is also supported so the command can be exercised
without proprietary game data.

```sh
go run ./internal/testapps/loot_roll \
  -file ./internal/testapps/loot_roll/example.json \
  -class "Act 1 Demo" \
  -seed 42
```

Each result includes the terminal item code and the treasure-class path used to
reach it. Reusing a seed and input file produces the same sequence. Positive
`picks` use `noDrop` and entry weights as weighted outcomes. Negative `picks`
use entry weights as fixed counts, matching the legacy OpenDiablo2 behavior.

Pass any combination of `-weapons`, `-armor`, and `-misc` to resolve direct
terminal codes into their base item name key, kind, level, type, and artwork:

```sh
go run ./internal/testapps/loot_roll \
  -file /path/to/TreasureClassEx.txt -class "Act 1 Good" -seed 42 \
  -weapons /path/to/weapons.txt -armor /path/to/armor.txt -misc /path/to/misc.txt \
  -item-types /path/to/ItemTypes.txt
```

With `-item-types`, generic and level-qualified codes are expanded through the
`Equiv1`/`Equiv2` hierarchy. A code such as `armo33` selects an equivalent item
whose base level is 33, 34, or 35. Candidates are sorted before seeded selection
so results do not depend on Go map iteration order.
