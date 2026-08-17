# Presentation asset coverage

Dark Magic keeps presentation identity in the layered shim and verifies known
Diablo II assets without checking proprietary bytes into the repository. Run:

```shell
make presentation-coverage
```

The deterministic JSON report joins three sources:

- static asset paths consumed by `presentation.v1.json` and the Lua shim;
- curated hypotheses in `asset-catalog.v1.json`;
- structural fingerprints in `asset-fixture.v1.json`.

The current baseline contains 173 manifest-owned paths: 99 have verified catalog
and fixture coverage and 74 remain explicitly unverified. Fourteen older catalog
paths are not currently selected by the presentation manifest. The Lua scan also
identifies 84 static paths still owned directly by compatibility or overlay code
and eight dynamic record/content prefixes. These are migration inventory, not
proof that an asset is absent from a supplied installation. The game-world
manifest now owns the exact blue/red warp COF and component paths; authoritative
warp state selects only the declared appearance token.

`catalog_fixture_gaps` must always be empty. CI also pins the complete sorted-set
fingerprint, so adding, removing, or moving a static path requires running the
report and deliberately classifying the change. Path comparison is
case-insensitive to match Diablo archive lookup behavior.

The remaining M15 catalog work is therefore measurable:

1. move the 84 code-owned static paths into versioned manifest data;
2. verify the 74 manifest paths against supported archive profiles and add their
   redistributable structural fixtures;
3. retain dynamic prefixes only where record-driven lookup makes enumeration
   inappropriate;
4. expand the inventory when additional screens, sounds, fonts, or TXT/TBL
   dependencies become runtime-consumed.
