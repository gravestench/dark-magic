# Contributing

Dark Magic welcomes both conventional and AI-assisted contributions. Most of
the project has been developed with repository-aware LLM coding agents under
active maintainer direction, and contributors are explicitly encouraged to
point their own coding agents at the repository and work in the same way.

Using an LLM is neither a waiver from engineering responsibility nor a reason to
distrust a contribution by itself. The submitter owns the scope, source
provenance, design, validation, security, performance, and final diff. Reviews
should judge repository evidence, architecture, behavior, tests, and
maintainability rather than whether every line was typed by hand. Human-only
contributions are equally welcome.

AI-assisted code is not second-class code; unverified code is. Large generated
diffs are also not evidence of progress. The goal is a small, coherent,
reviewable slice that advances a real acceptance boundary.

## Point the agent at the repository first

Do not begin with a blank prompt and general Diablo II knowledge. Give the agent
access to the current checkout and have it inspect the project before editing.
At minimum, it should read:

- [README.md](README.md) for the product and repository orientation;
- [ROADMAP.md](ROADMAP.md) for the stable product boundary and technical
  direction;
- the [Roadmap project](https://github.com/users/gravestench/projects/1),
  [issues](https://github.com/gravestench/dark-magic/issues), and
  [milestones](https://github.com/gravestench/dark-magic/milestones) for current
  status, ordering, and acceptance boundaries;
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for dependency and ownership
  rules;
- the relevant documents under [`docs/research`](docs/research), including the
  applicable source matrix or evidence ledger;
- the neighboring implementation, tests, package documentation, and recent
  history for the area being changed; and
- the issue, milestone, or pull request that defines the requested work.

The linked GitHub issue and milestone are the implementation-status authority;
the Roadmap project is their live scheduling and organization view. Research
documents are the behavioral evidence authorities. Neither an LLM's
recollection nor a plausible implementation upgrades an inferred behavior to
verified behavior.

A useful starter prompt is:

```text
Inspect this repository before editing. Read README.md, CONTRIBUTING.md,
ROADMAP.md, the linked GitHub issue and milestone, docs/ARCHITECTURE.md, and the
relevant research documents, source matrices, packages, and tests. Identify the
exact acceptance boundary, the current owner of each mechanism and policy, the
existing interfaces, and the repository evidence supporting the requested
behavior.

Propose a small coherent plan, then implement it without creating a parallel
authority. Add executable acceptance coverage, run the relevant test suites,
and perform a fresh review for correctness, architecture, determinism, source
provenance, security, dead code, overstated claims, and egregious performance
problems. Report the commands run and anything that remains unverified.
```

Adapt that prompt to the task. Ask the agent to cite concrete paths, symbols,
tests, and documents for its claims. Models can be confidently wrong; repository
evidence is the tie-breaker.

## Work in acceptance-sized slices

Start by naming the milestone, issue, or acceptance condition being advanced.
Inspect what already exists before proposing new abstractions. Do not add broad
content before the mechanism beneath it is coherent, and do not create parallel
stat, combat, skill, monster, item, quest, transition, session, party,
targeting, networking, or persistence authorities.

A strong contribution normally has one understandable through-line:

1. establish the missing behavior or mechanism;
2. connect it through the production path that owns it;
3. prove the named boundary with executable evidence;
4. update the linked issue and any durable documentation without overstating
   what passed; and
5. remove temporary scaffolding that is no longer needed.

Comments, unused helper functions, empty test files, test names, or assertions
that stop before the claimed outcome do not constitute acceptance coverage.

## Preserve architecture and ownership

The full contract is in [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md). In
particular:

- `cmd/client`, `cmd/server`, `cmd/realm`, tools, and test applications are
  composition roots. Keep product behavior in the internal packages that own
  it.
- Go owns reusable engine mechanisms. The first-party `d2legacy` Lua package
  owns Diablo II policy and data interpretation. Do not move policy into Go
  merely because it is easier for an agent to generate there.
- New long-lived native components use explicit construction, implement
  `Start(context.Context)` and `Stop(context.Context)` where lifecycle applies,
  and are registered through `internal/app/host`. Required dependencies are
  constructor arguments expressed as narrow, consumer-owned interfaces.
- Lua integrations use explicit, versioned capabilities under
  `internal/runtime/lua`. Never expose the host, runtime manager, renderer
  backend, or mutable Go objects as a service locator.
- Native resources use checked handles and remain attached to the active script
  scope so disable, reload, and shutdown release them.
- Short-lived scenes and overlays belong to
  `internal/presentation/navigation`, not the application component graph.
- Renderer and audio mutations cross their thread-safe command boundaries and
  are drained by the native owner thread.
- Gameplay state that can affect a future tick must be deterministic,
  serializable, checkpointed, and owned by an explicit engine store. Do not hide
  authority in Lua globals, closures, presentation state, or native handles.

When an agent proposes a new registry, cache, abstraction, manager, or copy of
state, make it explain why the existing owner cannot satisfy the requirement.
Prefer deleting duplication over documenting two authorities.

## Validate the actual claim

Use focused tests while iterating, then run the broad suites appropriate to the
change. Code changes should normally pass:

```sh
make test
make architecture
make test-race
```

Run the specialized suites that match the touched boundary, such as the Lua
hardening, network hardening/soak, Realm integration, capture, or profiling
targets in the Makefile. Changes that depend on legally obtained game data
should also exercise the relevant real-asset lab or capture path locally; do not
commit the resulting protected imagery or assets.

Documentation-only changes should still verify referenced paths, commands,
status language, and links against the current branch. State clearly in the PR
when a suite could not be run and why.

Tests must execute the promised behavior. After they pass, ask an agent—ideally
in a fresh context—to compare the PR's title and claims against the assertions
and production call path. A green suite does not prove code that the suite never
calls.

## Use review passes, not one-shot generation

LLMs are effective at broad repository analysis and repeated review, but the
first implementation pass should be treated as a draft. Perform at least one
separate review pass after the tests run. A fresh agent or context is often more
useful than asking the implementation session to approve itself.

Have that pass look specifically for:

- acceptance claims that the tests do not actually reach or assert;
- unused helpers, dead code, speculative scaffolding, and duplicated APIs;
- dependency-direction or Go/Lua ownership violations;
- nondeterministic state, incomplete checkpoint/replay behavior, or hidden
  lifetime ownership;
- authentication, admission, private-state, path, asset, and other trust-boundary
  mistakes;
- source-license or provenance problems;
- stale documentation and roadmap status that overstates implementation; and
- obvious performance pathologies in startup, owner-thread, per-tick, network,
  allocation, or asset-loading paths.

Weak generated output should be rewritten or removed, not defended because it
was expensive to produce. The contributor should understand the final design
well enough to explain why the ownership is correct and how the tests prove the
claim.

## Optimize at the right time

Correctness and coherent foundations come first, but that is not permission to
ignore pathological performance. Fix egregious problems when they are found,
including unbounded work, repeated archive or asset reads, accidental quadratic
loops, runaway allocation, blocking work on an owner thread, unbounded queues,
or per-tick operations that obviously scale with unrelated global state.

Broader optimization passes normally follow a coherent, tested foundation or a
completed milestone/acceptance slice. At that point the behavior is stable
enough to profile and the optimized code is likely to remain. This avoids
prematurely tuning disposable scaffolding while still making performance an
explicit development phase.

Measure rather than guess. For relevant client work, use the Makefile's profile
and capture targets and the checked profile budgets. For server, networking, and
Realm work, use representative deterministic tests, impairment/soak coverage,
and targeted profiling or benchmarks. Do not introduce another cache, lifetime
owner, packed representation, or concurrency model solely to make a number look
better without preserving the architecture and proving the measured need.

## Protect provenance, licenses, assets, and private data

Community projects and reverse-engineering references are evidence, not
code-generation templates. An LLM cannot launder incompatible source code or
erase attribution and license obligations. Record the source, target version,
confidence, conflicts, and holdout evidence required by the applicable research
ledger or source matrix.

Do not upload or commit Blizzard game assets, real-asset captures, credentials,
private keys, account data, private save files, proprietary material, or
unredacted sensitive logs. Be equally careful before sending such material to a
hosted model provider. Use synthetic fixtures, hashes, schemas, narrow excerpts,
and local tooling wherever possible.

## Pull request expectations

Keep the PR focused on one coherent slice. Its description should explain:

- what boundary was missing and why the chosen owner is correct;
- what changed in the production path;
- which tests, captures, profiles, or research evidence validate it;
- any performance effect or deferred optimization work;
- what remains incomplete; and
- any material LLM-assisted workflow used to produce and review the change.

A note such as the following is useful context, not an apology:

```text
Development: LLM-assisted. I reviewed the complete diff, ran make test and the
relevant focused suites, and used a separate review pass to check architecture,
acceptance coverage, dead code, and performance risks.
```

There is no need to publish every prompt, token count, or model name unless it
is relevant to reproducing a result. There is also no benefit in concealing
material tool use. The project is intentionally AI-forward and quality-driven;
transparency makes the workflow easier to review and improve.

Reviewers should identify concrete architectural, behavioral, evidentiary,
security, performance, or maintenance problems rather than rejecting work solely
because an LLM participated. Conversely, AI assistance does not lower the merge
bar. The same repository contracts and acceptance evidence apply to everyone.
