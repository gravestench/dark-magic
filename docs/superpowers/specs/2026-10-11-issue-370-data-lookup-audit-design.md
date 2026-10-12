# Issue #370: game-data lookup identity audit

## Purpose and boundary

Issue [#370](https://github.com/gravestench/dark-magic/issues/370) is the
remaining child of the G2 pinned-data milestone. A table join can appear to
work when a row's ordinal, authored number, and symbolic key happen to agree,
then select a different row when they diverge. The outcome is a reviewable
inventory of the table relationships used by current `d2legacy` consumers,
with executable cases that distinguish the relevant identities and corrections
only where a current consumer demonstrably uses the wrong one.

This audit establishes Dark Magic's interpretation of its pinned data. It does
not claim to reproduce the original Expansion 1.14d TXT-to-BIN compiler or to
verify 1.13c table bytes. The existing generation pinning, generic record
store, and DS1 object identity work from #369 remain the foundation.

## Ownership and approach

Keep join semantics in the consuming first-party `d2legacy` Lua module or its
existing narrow Go adapter. The generic `internal/game/data/store` may report a
caller-declared link but must not infer Diablo relationships from column names.
Do not add a global catalog, link registry, or second source-selection path.

Add a compact ledger to
`docs/research/GAME_DATA_LOADING_AND_LINKAGE.md`. Inventory production
`records.load` consumers under `internal/content/d2legacy/lua/d2legacy` and
the recovered DS1 object adapter, separating actual cross-table or positional
lookups from single-table scans. Group repeated uses of the same relationship
only when their identity and sentinel rules are identical. For each admitted
relationship, record:

- consumer and source table/column;
- target table/column, or positional target if no authored column exists;
- identity kind: symbolic, explicit numeric, row ordinal, act-local, or derived;
- empty, zero, negative, or other sentinel behavior actually used;
- version and evidence grade, with a pointer to the supporting owned data,
  repository test, or pinned research; and
- a test that distinguishes plausible identity choices, or an explicit
  evidence gap where the rule is not yet defensible.

Source provenance and the pinned generation accompany the relationship rather
than becoming another lookup key. The ledger must identify unknown historical
compilation rules without converting them into implementation facts.

## Executable checks and corrections

Use the existing Lua suite and synthetic rows to place a source reference at
one identity while a different target row occupies the same ordinal or numeric
value. Cover at least the current positional `LvlTypes` lookup, a numeric
`Skills.Id` lookup, and a symbolic cross-table lookup. Retain the #369 DS1
act-local divergence test as the act-local evidence, and verify it still runs.
Tests should reach the production consumer and assert the selected record or
observable result, not merely inspect an index helper.

Audit the remaining consumers against the ledger. Where a wrong identity is
confirmed, first add a failing consumer-level test, then make the smallest
correction in that owner. If the evidence does not distinguish two historical
interpretations, leave the current behavior unchanged, mark the ledger entry
uncertain, and name the needed owned-data or retail-runtime probe. Do not
expand this issue into new gameplay behavior or a general table loader redesign.

## Validation and review

Run focused Lua and Go tests for changed consumers, then `make test`,
`make architecture`, and `make test-race` on the final commit. Run applicable
real-asset checks locally when legally obtained game data is available; report
any skip as unverified. Review the finished diff against every #370 criterion,
the production call paths, dependency direction, deterministic pinning,
provenance, and scope. Complete every section of the pull request template
before submission, with exact commands and results from the submitted commit.

Issue #370 is complete only when the ledger covers current admitted
relationships, divergent-identity tests exercise the selected consumers, and
any corrected lookup has a failing-before/passing-after test. The audit may
leave explicitly labeled unknown legacy TXT-to-BIN semantics for later research.
