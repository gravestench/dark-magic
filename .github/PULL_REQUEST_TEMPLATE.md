## What changes for a player or contributor?

Explain the problem and the visible result in everyday language. Give a small example if it helps. Define any specialized term needed to follow the change.

## Which issue checks does this finish?

Link the issue and name the completion criteria this pull request meets. For each, say what changed and point to the test or demonstration that shows it. Use `Closes #123` only when the entire issue is finished; use `Refs #123` for partial work or a parent. For a typo, broken build, or fix to a recently merged pull request that has no issue, link the earlier pull request and explain the defect. Put the reference in this description so GitHub can link it and close a finished issue when the pull request merges into the default branch.

## What is in scope, and what is not?

Say which part of the issue this pull request delivers and what it deliberately leaves out. List any problems you found outside the issue, with links to the new issues that track them. Changes unrelated to the issue belong in a separate pull request.

## Where should a reviewer look?

Link a few key files, functions, tests, or documents. Say what each location owns or proves and why the change belongs there. A changed-file list alone is not enough.

## How do we know the behavior is right?

For each important behavior claim, link the evidence: owned game data (table and column), observations from the original game, research, or a pinned source such as a D2MOO commit. Name the target version. Mark each claim as verified, inferred, or still unknown, and say what would settle an unknown. Make sure each source supports the claim beside it. For a non-game change, name the governing issue or document and explain why game evidence does not apply.

## How was it checked?

Match the issue checks above to tests, captures, or repeatable manual steps, including important failure and edge cases. List exact commands and results on this commit. CI has no game data, so say whether tests that need it (`*_real_test.go`, `MPQ_DIRECTORY`) ran locally. Where practical, show that a new test fails without the change. Identify skipped or unrun checks instead of calling them passed.

## What remains uncertain or unfinished?

List unverified behavior, known risks, and follow-up issues. If nothing remains, say so and explain the limit of what your checks prove.

## How was AI used?

Say which parts an agent wrote or reviewed, confirm that you reviewed the complete diff, and describe what the separate review pass checked and found. Write "None" if no AI was used.
