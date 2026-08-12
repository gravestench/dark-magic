# Developer tools

These repository-private commands inspect or package content and are not Dark
Magic product binaries. Run them from the repository root with `go run`:

- `./internal/dev/tools/asset_inspect` inspects and previews one archive asset.
- `./internal/dev/tools/asset_catalog` verifies the curated presentation catalog and
  generates reports and contact sheets. Pass
  `-listfile ./docs/Diablo2UberListfile.txt` to separately audit community-listed
  paths against the selected local MPQ stack.
- `./internal/dev/tools/mpq2file` extracts one asset from a supported content source.
- `./internal/dev/tools/dt1_catalog` prints DT1 identities and, with `-stamps`,
  the exact mounted DT1 libraries declared by each DS1 stamp.
- `./internal/dev/tools/d2legacy_pack` packages the redistributable first-party `d2legacy` mod.
