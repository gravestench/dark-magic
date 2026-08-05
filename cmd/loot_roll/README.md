# Deterministic loot roller

`loot_roll` is a headless test application for treasure-class selection. Its
input can be the game's tab-delimited `TreasureClass.txt` or
`TreasureClassEx.txt`. JSON is also supported so the command can be exercised
without proprietary game data.

```sh
go run ./cmd/loot_roll \
  -file ./cmd/loot_roll/example.json \
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
go run ./cmd/loot_roll \
  -file /path/to/TreasureClassEx.txt -class "Act 1 Good" -seed 42 \
  -weapons /path/to/weapons.txt -armor /path/to/armor.txt -misc /path/to/misc.txt
```

Dynamic type codes remain marked unresolved until the item-type expansion stage.
