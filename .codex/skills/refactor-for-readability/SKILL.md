---
name: refactor-for-readability
description: >-
  Refactor an existing repository for human readability while preserving behavior. Use for repo-wide or
  package-by-package readability passes, oversized or deeply nested functions, mixed-domain files, unclear unexported
  helpers, code cuddling, inadequate rationale comments, line-length cleanup, or requests to make code approachable
  without redesigning its contracts.
---

# Refactor for Readability

Make the code easier for an unfamiliar maintainer to understand without changing observable behavior. Work from the
top down, complete one coherent section at a time, verify it, and pause for review before committing or moving on.

## Establish the Safety Boundary

1. Read repository instructions and inspect the working tree before editing.
2. Preserve unrelated changes and untracked files. Never fold them into the refactor.
3. Confirm the work is on an appropriate refactor branch. Create or switch branches only when authorized.
4. Identify the repository's formatter, linter, tests, CI scope, and line-length policy.
5. Run a proportionate baseline. Record pre-existing failures so the refactor is not blamed for them or used to hide
   them.

Treat public APIs, serialized formats, error behavior, ordering, concurrency, cancellation, ownership, and timing as
part of functional equivalence unless the user explicitly authorizes a behavior change.

## Work Section by Section

Choose the next section from the top-level execution path, then follow its dependencies. A section should be a coherent
domain such as command startup, configuration, logging, connection setup, recovery, package acquisition, or
presentation.

For each section:

1. Map responsibilities, callers, state transitions, invariants, and existing tests.
2. Identify mixed domains, long workflows, deep nesting, cuddled phases, and comments that merely restate syntax.
3. Refactor production code and its tests together.
4. Format, lint, and run targeted tests after meaningful increments.
5. Review the whole section again, including comments added earlier in the branch.
6. Run broader regression checks appropriate to the risk.
7. Summarize the section and pause for user review. Do not commit, push, or proceed to another section unless requested.

## Organize Around Domains

- Give each file one recognizable big idea. Prefer names such as `main.go`, `flags.go`, `logging.go`,
  `client_config.go`, `recovery.go`, or `codec.go`.
- Keep closely related state and helpers together. Do not create a file per function or proliferate tiny files.
- Keep entry points as readable orchestration: validate, construct, execute, and clean up through named helpers.
- Use descriptive unexported helpers to expose the phases and decisions hidden inside large functions.
- Split a function when it changes responsibility, crosses an abstraction boundary, repeats cleanup/error handling, or
  requires comments to explain multiple independent phases.
- Allow a somewhat longer linear constructor, fixture, or channel multiplexer when splitting it would obscure ownership.
  Length is a signal; mixed responsibility is the defect.

## Reduce Cognitive Load

- Enforce a 120-character maximum for code and comments unless a repository-specific rule is stricter.
- Keep nesting to four or five levels at most. Prefer validation, early returns, `continue`, and small helpers.
- Separate logical phases with whitespace. Avoid cuddling unrelated assignments, side effects, control flow, and
  assertions into one visual block.
- Keep related setup together when the statements form one obvious unit; whitespace should reveal structure, not add
  noise.
- Prefer precise domain names over generic helpers such as `handle`, `process`, `doWork`, or `setup`.
- Avoid abstraction solely to reduce line count. A helper should name a concept, isolate an invariant, or simplify
  control flow.
- Preserve local idioms unless they are the readability problem under review.

## Write Comments for Consequences

Add a comment immediately above every function, including unexported functions, test helpers, and methods. Audit
existing comments instead of grandfathering them in.

Function comments should explain the function's responsibility and at least one useful implication when applicable:

- why the boundary exists;
- which invariant it preserves;
- what callers may rely on;
- why ordering, locking, copying, cleanup, or validation matters;
- what failure or security condition it prevents.

Within function bodies, comment non-obvious decisions such as concurrency boundaries, trust checks, retry limits,
deterministic ordering, defensive copying, cancellation, data ownership, and compatibility constraints. Place comments
next to the decision they explain.

Do not add comments that translate syntax into English, repeat names, speculate, or become more difficult to maintain
than the code. If a block needs a long procedural explanation, first try extracting a well-named helper.

## Refactor Tests as First-Class Code

- Break long scenario tests into named phases and fixture helpers.
- Keep the top-level test readable as a behavioral narrative.
- Name helpers for the contract they establish or assert, not the mechanics they perform.
- Explain why sequencing, waits, retries, clocks, or synchronization are necessary.
- Preserve assertions and coverage. Do not make failures less diagnostic merely to shorten the test.
- Keep fixture ownership and cleanup visible so failures cannot leak processes, sockets, goroutines, or files.

## Verify Functional Equivalence

Use the repository's own commands first. For Go repositories, a typical progression is:

```text
gofmt changed files
golangci-lint run on the changed scope
go test on the changed package
go test on affected dependents
go test -race on synchronization-sensitive packages
go test ./...
git diff --check
```

Also verify all changed lines meet the line limit and every function has a preceding comment. A temporary audit command
is acceptable, but do not add repository tooling unless requested.

When full lint contains pre-existing failures, run both the configured CI scope and a changed-code or changed-package
lint check. Report the distinction clearly; never describe a partial lint run as repository-wide success.

Review the final diff for accidental behavior changes, renamed external contracts, altered error wrapping, changed
cleanup order, dropped assertions, and comment claims that the code does not guarantee.

## Hand Off Each Section

Lead with the completed outcome, then report:

- the domain/file layout;
- the important readability and rationale improvements;
- verification performed and any pre-existing failures;
- unrelated working-tree state left untouched;
- whether the section is uncommitted, committed, or pushed.

Keep commits cohesive and scoped to one reviewed section. Stage only intended files. Commit and push only when the user
asks, then identify the next top-down section before continuing.
