# Player Vendor Trade Context Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Require a live, matching Akara Trade interaction before player-authority vendor buy or sell changes items or gold.

**Architecture:** Keep the guard in `commands/item_vendor.lua` at application time. Read authoritative player, interaction target, and item placement facts; leave shared interaction range behavior and administrator/bootstrap commands intact.

**Tech Stack:** Lua authority command/ECS harness and Go session acceptance tests.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-613-vendor-trade-context-design.md`

## Global Constraints

- Issue #613 only; parent #360 retains vendor stock, other NPC trade pages, pricing fidelity, gambling, repair, and Cube work.
- Stale or mismatched player commands no-op at application time and leave session, gold, items, and offers usable and unchanged.
- Administrator/bootstrap semantics and the Akara permanent-offer eligibility check remain intact.
- This is a repository authority correction, without a new original-game formula or version claim.
- Do not commit game data, captures, credentials, or private saves.

## Review Focus

- Validation versus application: an unknown player-supplied sell vendor must reach the apply-time no-op guard; administrator terms validation must remain.
- Admitted player and location: the existing context helper accepts missing player/location, so the guard must explicitly require player position and both locations with matching act and level.
- Buy category comes from the item's authoritative vendor placement slot; sell category comes from the validated payload; use exact comma-token membership.
- The valid player path uses synthetic ordinary Akara stock and a real Akara Trade interaction; Charsi page enablement and vendor-specific stock ownership are outside scope.
- Invalid held sales still assert after a valid trade gate; missing ordinary buys still no-op; permanent offers retain their stricter checks.

---

### Task 1: Player trade authorization through the authority command path

**Files:**
- Create: `internal/content/d2legacy/lua/d2legacy/commands/item_vendor_context_test.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/commands/item_vendor.lua`

**Interfaces:**
- Consumes: `contexts.active(owner)`, admitted `d2legacy.player.identity` with position/location, target entity location, item placement slot, existing `owner.resolve` and `terms`.
- Produces: player buy/sell apply-time no-op on stale or mismatched context; no new public interface.

- [ ] **Step 1: Add a real-path failing Lua suite.** Give `alice` 1000 carried gold, a synthetic ordinary vendor item `stock` in slot `misc` with base cost 40, a held `sale` item with base cost 60, Akara buy/sell multipliers 1024, and an Akara interaction target. Use player entry plus `interaction.open`/`interaction.choose` to prove selling `sale` raises gold to 1060 and places it in `vendor/misc`, then buying `stock` lowers gold to 1020 and places it in `held`. Add closed trade page, mismatched and unknown vendor, out-of-range, missing player admission/location, missing target location, changed level, disallowed category, and missing ordinary buy cases. For the missing-fact cases, remove the relevant ECS component after entry before submitting the command. Assert unchanged gold/item placement and a later valid command still succeeds; retain checkpoint parity and the existing permanent-offer suite. The stale/mismatched test must fail on current code.
- [ ] **Step 2: Confirm RED.** Run `DARK_MAGIC_LUA_TEST_TIERS=fast,integration go test -count=1 ./internal/mod/d2legacy -run TestLuaSuites`; expect the new stale/mismatched assertion to fail on item or gold state, not a fixture error.
- [ ] **Step 3: Implement the local guard.** In `item_vendor.lua`, import the existing interaction context; for player commands, require an admitted player with position/location, current target with location and matching act/level, active in-range `trade` context, case-normalized vendor match, and exact category token. Read buy item/placement first to obtain its authoritative slot; guard before offer purchase, arrangement, placement, or gold mutation. Defer player sell `terms` lookup from `validate` until after the guard; preserve administrator validation.
- [ ] **Step 4: Confirm GREEN.** Run the focused Lua suite and verify all new cases and the existing permanent-offer cases pass. Inspect the test for real `interaction.open`/`interaction.choose` execution and literal expected state.
- [ ] **Step 5: Commit.** Commit only the vendor Lua code and its focused tests with #613 in the message.

### Task 2: Preserve item lifecycle bootstrap coverage

**Files:**
- Modify: `internal/mod/d2legacy/item_acceptance_test.go`

**Interfaces:**
- Consumes: Task 1 player-only guard and the existing administrator command authority.
- Produces: the same Charsi fixture sale/lifecycle/restore assertions, explicitly submitted as administrator bootstrap commands.

- [ ] **Step 1: Confirm the regression boundary.** Run `go test -count=1 ./internal/mod/d2legacy -run 'Test(ItemReconnectRecoveryEquipmentVendorAndServiceRestoreIdentically|EquippedAttackRatingSourceIsAddedAndRemoved|EquippedArmorDefenseSourcesAreAddedAndRemoved)'` after Task 1; expect the Charsi player fixture sale to no-op and at least one existing lifecycle assertion to fail.
- [ ] **Step 2: Adapt only fixture authority.** Introduce or use an administrator item-command helper for the three Charsi bootstrap sales; preserve their ticks, payloads, sequences, and outcome/restore assertions. Leave real player Akara trade coverage in Task 1.
- [ ] **Step 3: Confirm GREEN and commit.** Run the same focused Go tests and the focused Lua suite; commit only the Go test adaptation with #613 in the message.

### Task 3: Verify, review, and integrate

**Files:**
- Read: `.github/PULL_REQUEST_TEMPLATE.md`, issue #613, and full `origin/main...HEAD` diff.

**Interfaces:**
- Consumes: Tasks 1–2 code and tests.
- Produces: a reviewed PR that closes #613 and references #360.

- [ ] **Step 1: Verify final commit.** Run `go test ./...`, `make architecture`, `make test-lua-hardening`, `make test-race`, and `git diff --check origin/main...HEAD`; report owned-game-data test status. If race cannot run, state why unverified.
- [ ] **Step 2: Obtain independent code and evidence reviews.** Resolve in-scope blockers, rerun every check claimed in the PR on the new final commit, and inspect the complete effective diff.
- [ ] **Step 3: Submit and integrate.** Fill every PR template section, push and create the PR, attach it to this chat, wait for required CI, then merge after it passes and record the result.
