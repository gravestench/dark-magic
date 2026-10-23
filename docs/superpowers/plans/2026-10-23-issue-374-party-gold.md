# Party Gold Pickup and Sharing Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let a player collect a trusted world gold pile through the authoritative interaction path and divide accepted gold among current party members under the bounded #374 rule.

**Architecture:** A `d2legacy.gold.pile` ECS component rides on an existing selectable interaction target. `interaction.open` dispatches to a gold operation; Lua policy computes the ordered allocation; the operation writes existing carried-gold layouts and updates the pile atomically at one tick. No new command or gold store is introduced.

**Tech Stack:** First-party `d2legacy` Lua, generic ECS/authority commands, Go session tests, existing Superpowers workflow.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-374-party-gold-design.md`

## Global Constraints

- Link issue #374 with `Refs #374`; leave it open for kill/quest credit, retail gold generation, and evidence gaps.
- Apply GPS-R01/R02 from 1.13c static evidence only to the stated supplied-input assertions; 1.14d parity is inferred.
- Production pile amount is integer `1..2147483647`; stored levels are `1..99`; existing balances are `0..2147483647`.
- Gold recipient selection has no XP proximity test. Count living same-level members separately from the ordered recipient pass; the latter does not recheck recipient liveness.
- Production uses `player.progress.level` as a provisional effective-level proxy and clamps accepted gold nonnegative for an already-over-cap balance. Isolated policy tests retain signed original arithmetic.
- Keep positive excess on the same pile and mark fully consumed piles inactive; both are explicit Dark Magic integration choices pending original recreation evidence.
- Do not commit protected game data, captures, saves, or credentials.

## Review Focus

- An unadmitted or forged player/remote target must never receive gold even though generic interaction admission has an administrator/bootstrap fallback; Task 2 tests both.
- A dead member may receive a share when another living member makes the divisor greater than one; Task 1 and Task 2 test the distinct passes.
- A recipient or picker already above the cap must not lose gold or make the pile increase; Task 1 and Task 2 test conservation.
- An authored pile consumed before a later player joins must stay consumed after `ensure_player` and checkpoint restore; Task 2 tests it.
- A malformed trusted pile amount or missing spatial facts must fail at import, before a command can access it; Task 2 tests it.
- If all eligible balances are at cap, repeated pickup must neither increase gold nor increase the pile; Task 2 tests it.

---

### Task 1: Gold split policy

**Files:**
- Create: `internal/content/d2legacy/lua/d2legacy/policy/party_gold.lua`
- Create: `internal/content/d2legacy/lua/d2legacy/policy/party_gold_test.lua`

**Interfaces:**
- Consumes: plain ordered member snapshots `{id, counts_for_divisor, can_receive, gold, level}`; the picker entry is identified by `picker_id`.
- Produces: `party_gold.allocate(amount, picker_id, members, safe_production)` returning `{requests = ordered {id, accepted, kind, amount, next_balance}, excess, count}`. `kind` is `add` or `replace`; `amount` is the requested stat argument (zero for `replace`). `safe_production=false` models signed GPS-R01/R02 request routing; `next_balance` is only a Dark Magic modeled result, not an assertion about original stat propagation. `true` uses the documented nonnegative over-cap clamp, and the production caller writes its `next_balance` only after complete validation.

- [ ] **Step 1: Write failing `party_gold_test.lua` cases.** Assert GPS-V01 ordered add requests of A 333, B 333, picker 334; GPS-V02 ordinary cap add requests 100/901 and picker-cap excess 851; dead B counts false but can receive true; negative -1000/3 -> add(-333), add(-333), add(-334); GPS-V02(f) signed wrap accepts 100 but routes a `replace(gold, 0)` request; supplied level 10 -> cap 100000; safe over-cap recipients accept zero and do not debit. Assert count/order/excess and local request routing only; do not claim an effective-stat getter is integrated or infer original stored balances or draws inside stat/recreation dependencies.
- [ ] **Step 2: Run `GOCACHE=/private/tmp/dark-magic-issue373-live-gocache GOMAXPROCS=2 GOFLAGS=-p=2 go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/policy/party_gold' -count=1`; record the missing-module failure.**
- [ ] **Step 3: Implement `allocate` and signed helpers in `party_gold.lua`.** Use truncation toward zero and explicit low32/i32 at the operations GPS-R01 specifies. Return each original-mode add/replace request separately from its modeled balance. When `safe_production=true`, clamp accepted to `max(0, min(requested, cap-gold))` for positive requested amounts and preserve an over-cap balance. Keep the picker-only branch and stable other-member iteration distinct.
- [ ] **Step 4: Run the focused policy suite until it passes, then commit the two files.**

### Task 2: Authoritative gold-pile pickup

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/components/items.lua` (pile schema)
- Modify: `internal/content/d2legacy/lua/d2legacy/items/bootstrap.lua` (trusted target import)
- Create: `internal/content/d2legacy/lua/d2legacy/systems/gold_pickup.lua` (operation)
- Modify: `internal/content/d2legacy/lua/d2legacy/authoritative.lua` (registration)
- Create: `internal/content/d2legacy/lua/d2legacy/systems/gold_pickup_test.lua`
- Create: `internal/mod/d2legacy/party_gold_test.go`

**Interfaces:**
- Consumes: `party_gold.allocate(..., true)` from Task 1, `party.snapshot()`, `d2legacy.player.identity/progress/vitals/death`, `d2legacy.world.location/position`, `d2legacy.items.layout`, and `d2legacy.interaction.target`.
- Produces: optional trusted `gold_amount` on `d2legacy.interactions.targets`; `d2legacy.gold.pile` `{amount:i64}`; `gold_pickup.register()` registering an operation for this component. The pile's stable target ID and inactive flag are the existing world identity/lifetime mechanism.

- [ ] **Step 1: Write failing Lua and Go production tests.** Import a selectable gold target with act/level/position/radius/amount. Submit `interaction.open` and assert actual carried-gold layouts plus pile amount/inactive state for solo, far same-level 1000/3, other-level, dead recipient, leave/rejoin, 1001 cap/excess, over-cap safety, zero acceptance followed by repeat pickup, duplicate pickup, and distinct checkpoint restores before and after pickup. Assert the selectable exists before pickup, disappears after complete pickup, and balances appear in item projection. Test an unadmitted/forged remote attempt, out-of-range attempt, malformed amount, missing spatial facts, an invalid-domain recipient skipped, and an invalid-domain picker abort with pile and every layout unchanged. Preserve original party XP and NoDrop assertions in their existing suites.
- [ ] **Step 2: Run focused `go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/systems/gold_pickup|TestPartyGold' -count=1` with the task GOCACHE/GOMAXPROCS/GOFLAGS environment; record the expected failure before implementation.**
- [ ] **Step 3: Add the ECS schema, validated import, and operation.** Require `selectable=true`, act/level, finite coordinates and positive radius for a trusted pile; validate these before `create_target` returns a previously imported ID, and preserve the existing pile state across joins/restores. Revalidate one admitted living picker whose identity equals owner, target act/level/range, active positive amount, and valid layout/progress without relying on generic `contexts.in_range` alone. Snapshot current party in stable order; compute the living divisor from same-level players, then build recipient snapshots separately, allowing a dead same-level recipient. Reject missing/invalid picker before any mutation. Apply only safe-production modeled balances after validation, set pile amount to excess, and set `d2legacy.world.inactive` at zero (do not destroy its target ID).
- [ ] **Step 4: Run focused Lua and Go session suites until they pass; commit the production files and tests.**

### Task 3: Verify, review, and integrate the delivered slice

**Files:**
- Review: all changed paths, issue #374, GPS spec/evidence/gap, PR template.
- Update: issue #374 after merge with the delivered gold boundary and remaining work.

- [ ] **Step 1: Run `make test`, `make architecture`, `make test-race`, and `make test-lua-hardening` on the candidate commit with the task GOCACHE/GOMAXPROCS/GOFLAGS environment.** Record exact results and whether `MPQ_DIRECTORY` real-asset tests ran.
- [ ] **Step 2: Have fresh independent subagents review production authority, replay/conservation, actual session assertions, source evidence, scope, and PR claims.** Fix concrete findings and rerun the affected and final checks on the submitted commit.
- [ ] **Step 3: Fetch/rebase onto current `main`, verify the submitted commit, write every PR-template section, push, and open a PR with `Refs #374`.** Attach the PR artifact.
- [ ] **Step 4: Merge only after all GitHub CI checks pass.** Update #374 with the slice, preserve open parity/other-reward criteria, and continue to the next linked issue.
