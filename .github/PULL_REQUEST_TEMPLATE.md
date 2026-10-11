## What changes for a player or contributor?

Explain the problem and the visible result in everyday language. Define any specialized term needed to follow the change.

## Which issue does this finish?

Use `Closes #123` when this pull request satisfies the entire issue. Use `Refs #123` when it advances only part of an issue or its parent. Put the reference in this description so GitHub can link it and close the finished issue when the pull request merges into the default branch.

## What is in scope, and what is not?

Say which part of the issue this pull request delivers and what it deliberately leaves out. List any problems you found outside the issue, with links to the new issues that track them. Changes unrelated to the issue belong in a separate pull request.

## How do we know the behavior is right?

Link the evidence: owned game data (table and column), observations from the original game, or a pinned source such as a D2MOO commit. Mark which claims are verified and which are inferred.

## How was it checked?

List the behavior tested, the exact commands or manual steps run, and their results on this commit. CI has no game data, so say whether tests that need it (`*_real_test.go`, `MPQ_DIRECTORY`) were run locally. Where practical, show that a new test fails without the change. List any known result that remains unverified.

## How was AI used?

Say which parts an agent wrote or reviewed, confirm that you reviewed the complete diff, and describe the separate review pass. Write "None" if no AI was used.
