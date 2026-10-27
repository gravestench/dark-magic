# Issue #649: macOS Race CI Job Time Budget

## Intent

Let the complete macOS race suite report a result on GitHub Actions. PR #647's second CI attempt passed earlier stages and its self-hosted transform test, then GitHub cancelled the job at its 30-minute limit while `d2legacy.test` was still running. The first attempt completed the race command in about 19 minutes 47 seconds after roughly nine minutes of preceding job work, leaving almost no margin.

## Approaches

1. **Chosen: raise only this job's limit from 30 to 45 minutes.** It retains every existing test and gives the current serial race command room to finish. An independent CI evidence reviewer approved this narrow approach.
2. Split race packages across jobs. This could shorten elapsed time but changes CI structure, scheduling, and resource use beyond the observed gate failure.
3. Reduce race coverage or increase only Go's per-package limit. The former weakens the gate; the latter cannot prevent GitHub's job-level cancellation.

## Boundary and verification

Change only `timeout-minutes` on `Go test and race` in `.github/workflows/ci.yml`. Keep `make test-race`, its 20-minute per-package timeout, other job limits, and gameplay code unchanged. Verify the YAML diff locally and require a completed macOS CI run with a race-stage result. An unrelated test assertion failure is a distinct diagnosis, including the intermittent transform timeout tracked in #635. The 45-minute budget is a practical margin based on observed runs, not a guarantee for every future runner.
