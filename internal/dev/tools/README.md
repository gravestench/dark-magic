# Developer tools

These repository-private commands inspect or package content and are not Dark
Magic product binaries. Run them from the repository root with `go run`:

- `./internal/dev/tools/asset_inspect` inspects and previews one archive asset.
- `./internal/dev/tools/asset_catalog` verifies the curated presentation catalog and
  generates reports and contact sheets. Pass
  `-listfile ./docs/Diablo2UberListfile.txt` to separately audit community-listed
  paths against the selected local MPQ stack.
- `./internal/dev/tools/mpq2file` extracts one asset from a supported content source.
- `./internal/dev/tools/shim_pack` packages the redistributable Dark Magic shim.
