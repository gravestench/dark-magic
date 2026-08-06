# Deterministic item-quality roller

`quality_roll` selects the matching row from a supplied `itemratio.txt`, applies
level difference, magic find, minimum denominators, and TreasureClass modifiers,
then performs the ordered quality checks with a stable seed.

```sh
go run ./internal/testapps/quality_roll \
  -file /path/to/itemratio.txt -version 100 \
  -monster-level 35 -item-level 28 -magic-find 100 \
  -tc-unique 512 -seed 42
```

The JSON output includes every final denominator. Diablo II quality rolls use
128 as the successful range, so smaller denominators represent better odds.
