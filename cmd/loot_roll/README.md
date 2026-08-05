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
