# OpenDiablo2 Codec Maintenance

Dark Magic consumes the OpenDiablo2-derived codecs, but the codecs should
remain independent Go modules and Git repositories. They are useful outside
this engine, have different release cadences, and should not inherit the
engine's rendering or service dependencies.

The codec collection contains twelve repositories:
`bitstream`, `cof`, `dc6`, `dcc`, `ds1`, `dt1`, `gpl`, `mpq`, `pl2`,
`tbl_text`, `tsv`, and `wav`.

## Repository strategy

- Keep each codec in its own repository and publish normal tagged Go modules.
- Do not commit absolute-path `replace` directives to `go.mod` files.
- Use an uncommitted `go.work` file while changing Dark Magic and codecs
  together.
- Keep codec APIs headless. Put viewers and converters under `cmd/` so GUI
  dependencies cannot prevent core package tests.
- Keep proprietary assets out of Git. Real-asset tests should use an optional
  asset-root environment variable and skip cleanly when it is absent.

Create a development workspace without changing module metadata by adding this
repository and the checked-out codec module directories:

```sh
go work init .
go work use /path/to/codec-module
```

`go.work` remains uncommitted because checkout locations are not part of Dark
Magic's build contract.

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
| P1 | `tsv` | Typed record decoding with malformed-row diagnostics |
| P1 | `wav` | Standard-library interoperability tests |

## Completed maintenance pass (2026-08-05)

- The original eleven binary codecs pass `go test ./...` and `go vet ./...` in a shared
  workspace, with race checks on their decoder paths.
- Every public decoder has malformed/truncated-input coverage and a fuzz target;
  format counts, offsets, dimensions, and allocation sizes are bounded before
  use. Synthetic fixtures and encoder round trips cover formats that can be
  produced without proprietary data.
- `mpq` includes an opt-in `OD2_MPQ_TEST_FILE` smoke test. Optional expected
  file-count and list-hash variables make owned archive fixtures reproducible
  without placing Diablo II data in Git.
- Headless JSON inspection and common-format conversion commands live under
  each applicable module's `cmd/` tree. GUI viewers remain optional commands.
- `bitstream` is published as `v0.2.0`; the other ten modules are published as
  `v0.1.0`. Dark Magic consumes these tags and passes `go test ./...` and
  `go vet ./...` with `GOWORK=off`, proving that no filesystem replacement is
  needed.
- The TSV codec is restored as Dark Magic's typed tabular format boundary. It
  remains on its historical pseudo-version pending malformed-input,
  concurrency, diagnostics, and tagged-release maintenance.
