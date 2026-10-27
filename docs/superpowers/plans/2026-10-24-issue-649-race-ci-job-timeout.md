# macOS Race CI Job Time Budget Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let the complete macOS race CI suite finish without the 30-minute GitHub Actions job limit cancelling it.

**Architecture:** Raise only the existing `Go test and race` job's time budget. Keep the same workflow steps and `make test-race` command, so the required gate still runs the full suite.

**Tech Stack:** GitHub Actions YAML, Makefile, `gh` CI checks.

**Spec:** `docs/superpowers/specs/2026-10-24-issue-649-race-ci-job-timeout-design.md`

## Global Constraints

- Issue #649 alone owns this change; #635 owns the intermittent QUIC transform assertion.
- Change only `timeout-minutes` on `Go test and race` in `.github/workflows/ci.yml` from 30 to 45.
- Keep `make test-race`, its 20-minute per-package timeout, all other job limits, and gameplay code unchanged.
- A completed macOS CI run must report a race-stage result; a separate assertion failure is not a timeout success claim.

## Review Focus

- Accidentally changing a different job's limit would leave this race job cancelled; inspect job context in the diff.
- YAML indentation or numeric parsing failure would prevent CI from running; parse the workflow after edit.
- A green unit step with cancelled race stage is not acceptance; inspect the race job's conclusion after push.
- An unrelated QUIC assertion failure may still occur; report it under #635 without weakening the test.
- Do not claim `make test-race` passed merely from a YAML change; use the actual Actions outcome.

---

### Task 1: Increase the Race Job Limit

**Files:**
- Modify: `.github/workflows/ci.yml`

**Interfaces:**
- Consumes: existing `Go test and race` job and unchanged `make test-race` step.
- Produces: job-level `timeout-minutes: 45` for that job alone.

- [ ] **Step 1: Check RED/current value.** Run `rg -n -C 3 'name: Go test and race|timeout-minutes:' .github/workflows/ci.yml`; record that this job has `timeout-minutes: 30`.
- [ ] **Step 2: Edit one value.** Change that job's `timeout-minutes` from 30 to 45; change no other workflow or Makefile line.
- [ ] **Step 3: Verify before commit.** Run `ruby -ryaml -e 'YAML.parse_file(ARGV.fetch(0))' .github/workflows/ci.yml` and expect exit 0. Inspect `git diff -- .github/workflows/ci.yml` and `git diff --check`; confirm only that numeric value changed.
- [ ] **Step 4: Commit** the workflow edit, then repeat the Ruby parser and run `git diff --check HEAD^ HEAD` on the submitted commit. Obtain independent correctness/scope and formatting/readability reviews; run any specific follow-up check they require.

## Delivery

- [ ] Complete every PR-template section with the Actions evidence and exact checks run on the commit, push and attach a PR that closes #649. Wait for the macOS race job to finish with a result; merge after required CI passes.

After #649 merges, resume the separately authorized #612 PR #647 delivery: rebase it on main, rerun CI, and merge when green. That work is outside this issue's completion rule.
