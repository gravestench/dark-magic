# OpenDiablo2 Codec Maintenance

Dark Magic consumes the OpenDiablo2-derived codecs, but the codecs should
remain independent Go modules and Git repositories. They are useful outside
this engine, have different release cadences, and should not inherit the
engine's rendering or service dependencies.

The sibling checkout at `../od2_codecs` currently contains eleven repositories:
`bitstream`, `cof`, `dc6`, `dcc`, `ds1`, `dt1`, `gpl`, `mpq`, `pl2`,
`tbl_text`, and `wav`.

## Repository strategy

- Keep each codec in its own repository and publish normal tagged Go modules.
- Do not commit absolute-path `replace` directives to `go.mod` files.
- Use a local `go.work` file while changing Dark Magic and codecs together.
- Keep codec APIs headless. Put viewers and converters under `cmd/` so GUI
  dependencies cannot prevent core package tests.
- Keep proprietary assets out of Git. Real-asset tests should use an optional
  asset-root environment variable and skip cleanly when it is absent.

Create a local workspace without changing module metadata:

```sh
go work init .
go work use ../od2_codecs/{bitstream,cof,dc6,dcc,ds1,dt1,gpl,mpq,pl2,tbl_text,wav}
```

`go.work` remains uncommitted because the sibling location is a local checkout
convention, not part of Dark Magic's build contract.

## Milestones

1. Baseline every module with `go test`, `go vet`, `go test -race`, and a
   minimal README API example. Remove machine-specific replacements.
2. Add malformed/truncated-input tests and fuzz targets for every public
   decoder. Invalid input must return an error rather than panic or allocate
   from unchecked file sizes.
3. Add redistributable synthetic fixtures and encode/decode round-trip tests
   wherever an encoder exists.
4. Add opt-in smoke tests against user-supplied MPQs. Assert dimensions,
   counts, and structural hashes rather than proprietary golden images.
5. Standardize headless commands: `inspect` emits JSON and `convert` writes a
   common format with useful exit status and diagnostics.
6. Tag compatible releases and move Dark Magic from pseudo-versions to tags.

## Priority

| Priority | Modules | First useful outcome |
| --- | --- | --- |
| P0 | `bitstream` | Bounds-safe primitives and fuzz coverage |
| P0 | `mpq` | `fs.FS` conformance and real-archive smoke tests |
| P0 | `dc6`, `dcc` | Headless decode and PNG conversion tests |
| P0 | `ds1`, `dt1` | Structural validation and renderer fixtures |
| P1 | `pl2`, `gpl` | Round trips and palette image export |
| P1 | `tbl_text` | Lookup, duplicate, and corruption tests |
| P1 | `cof` | Validation and JSON inspection command |
| P1 | `wav` | Standard-library interoperability tests |

## First audit (2026-08-05)

- `bitstream`, `dt1`, `mpq`, `pl2` core, and `wav` compile locally.
- Other modules could not finish the offline audit because dependency versions
  were not cached; those results are not yet classified as codec failures.
- `pl2/go.mod` contains a developer-specific absolute `replace` directive.
- Coverage is uneven and several modules have no package tests.
- `dc6` and `dcc` carry GUI dependencies in their module graphs; core decoding
  and optional viewers should be independently testable.
