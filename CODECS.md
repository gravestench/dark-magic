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
- The TSV codec is restored as Dark Magic's typed tabular format boundary. Its
  original slice-only API remained pending after this first pass and was closed
  by the streaming pass below.

## Completed streaming-I/O pass (2026-08-10)

- `bitstream` now has incremental readers and writers that retain at most the
  current partial byte instead of copying the complete source.
- Sequential formats consume `io.Reader` and emit to `io.Writer` where
  applicable: COF, DS1, GPL, PL2, TSV, and WAV/Huffman. Existing byte-slice
  entry points remain wrappers for compatibility.
- Offset-oriented formats expose lazy `io.ReaderAt` files: DC6 frames, DCC
  directions, DT1 tiles, and TBL tables. Opening reads only bounded metadata;
  callers choose when to materialize payloads.
- MPQ archive metadata and sectors use positional reads. Separate entry streams
  and `MpqDataStream.ReadAt` are safe for concurrent use without a shared seek
  cursor. The optional real-archive race test passes against an owned English
  `d2data.mpq` containing 10,814 listed files.
- Dark Magic consumes the new module revisions directly. Directory and MPQ
  assets take the random-access path; ZIP and minimal test files retain a
  compatibility buffer fallback. COF, DS1, PL2, palette, and localization
  loaders consume streams directly.
- The pass is published as `bitstream` v0.3.0 and v0.2.0 for COF, DC6, DCC,
  DS1, DT1, GPL, MPQ, PL2, TBL, TSV, and WAV. Dark Magic depends on the stable
  tags, not checkout paths or pseudo-versions.

The next performance step is selective residency in the engine caches: retain
lazy DC6/DCC/DT1 file handles long enough to decode only requested frames,
directions, and tiles. The codec boundary no longer requires another redesign
for that work.

## Historical reverse-engineering research

Paul Siramy's Phrozen Keep-era documentation and tools are a priority source for
the next DS1, DT1, DC6, and DCC pass. The maintained source index, evidence
labels, and regression queue live in
[`docs/research/paul-siramy.md`](docs/research/paul-siramy.md). Claims recovered
from archived community material must be verified against legally supplied data
or synthetic fixtures before they become codec or engine contracts.

The current legacy-format research pass turns those findings into format and
runtime specifications under [`docs/formats`](docs/formats) and records the
codec-repository work that should be planned separately in
[`docs/formats/CODEC_FOLLOWUPS.md`](docs/formats/CODEC_FOLLOWUPS.md). Codex should
use that handoff before changing `gravestench/cof`, `dc6`, `dcc`, `ds1`, `dt1`,
or `pl2`; the follow-up list intentionally separates codec-owned probes/tests
from Dark Magic-owned DRLG, composite resolution, and rendering work.
