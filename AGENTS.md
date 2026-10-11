# Instructions for coding agents

Read [CONTRIBUTING.md](CONTRIBUTING.md) before editing. Its "Rules for
AI-assisted work" section is mandatory. In short:

- **Work only on the linked issue.** Before changing code, name the issue and
  its acceptance boundary. If the task has no issue, stop and ask for one.
- **Stay inside that boundary.** Do not refactor, rename, reformat untouched
  files, update dependencies, or fix unrelated problems. Report anything
  outside the issue to the person you are working with instead of changing it.
- **Ask before growing the change.** If the work needs much more than the
  issue describes, stop and ask whether to split it.
- **Write the pull request description before submitting.** Fill in every
  section of [the template](.github/PULL_REQUEST_TEMPLATE.md): what changed,
  scope, evidence, the exact checks run and their results, what remains
  unverified, and how AI was used.
- **Report only what you ran.** Never claim a test or check passed unless it
  ran on the submitted commit. Tests that need game data skip in CI, so say
  whether they ran locally.
- **Never commit Diablo II game data**, captures made from it, credentials, or
  private saves.
