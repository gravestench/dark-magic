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

The M15.5 baseline contains 139 manifest-owned paths: 73 have verified catalog
and fixture coverage and 66 remain explicitly unverified. Seventeen older
catalog paths are not currently selected by the presentation manifest. The Lua
scan also identifies 47 static paths still owned directly by compatibility or
overlay code and one dynamic skill-icon prefix. These are migration inventory,
not proof that an asset is absent from a supplied installation.

`catalog_fixture_gaps` must always be empty. CI also pins the complete sorted-set
fingerprint, so adding, removing, or moving a static path requires running the
report and deliberately classifying the change. Path comparison is
case-insensitive to match Diablo archive lookup behavior.

The remaining M15 catalog work is therefore measurable:

1. move the 47 code-owned static paths into versioned manifest data;
2. verify the 66 manifest paths against supported archive profiles and add their
   redistributable structural fixtures;
3. retain dynamic prefixes only where record-driven lookup makes enumeration
   inappropriate;
4. expand the inventory when additional screens, sounds, fonts, or TXT/TBL
   dependencies become runtime-consumed.
