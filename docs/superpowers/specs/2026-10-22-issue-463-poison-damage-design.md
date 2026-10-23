# Issue #463: poison damage over time

## Purpose and acceptance boundary

Issue [#463](https://github.com/gravestench/dark-magic/issues/463) makes a
landed poison hit drain life over its authored duration. A mitigated rate of
`X` raw 1/256-life units lasts `L` simulation frames for players and monsters.
An equal or stronger application replaces the old rate and expiry; a weaker
application leaves both unchanged. The state is visible as a green target tint,
and Cleansing can shorten its remaining duration. Poison death uses the
attacker recorded at application time for ordinary death and reward handling.

This issue supplies the shared mechanism and a basic-melee producer from
existing poison stat sources. Skill-specific poison formulas and durations,
poison-length reduction items, monster modifiers, curses, and pierce remain
with their linked issues. A poison channel without a positive duration remains
an inert staged fact; the mechanism does not invent a duration for it.

## Evidence and chosen interpretation

The owned 1.14d `ItemStatCost.txt` identifies `poisonmindam` (57),
`poisonmaxdam` (58), `poisonlength` (59), and `hpregen` (74). The owned
`States.txt` poison row is state 2 with `green` and `castoverlay=poisonhit`.
The repository's
[status-application specification](../../specs/combat-status-application.md)
verifies the 1.13c poison applier's positive-rate/length gates, negative
`hpregen` state stat, stronger-or-equal replacement, and expiry reset. The
1.14d libd2 reconstruction cited in #463 corroborates a per-frame regeneration
consumer and town suppression, but leaves the negative branch unimplemented.
The exact 1.14d player-life rounding, first/last tick, and town rule are not
verified against the owned runtime. They are implementation interpretations
below, not claims of original-game parity. The 1.13c behavior remains inferred
for 1.14d where no owned runtime trace exists; 1.10f D2MOO is code-shape
corroboration only.

Use the existing typed damage transaction, combat events, timed-state owner,
and presentation projection. A direct poison call in each damage producer would
duplicate application policy; a separate poison lifecycle would duplicate the
timed-state and Cleansing authorities. Neither is needed for this boundary.

## Authoritative data flow

1. Basic melee reads resolved `poisonmindam`, `poisonmaxdam`, and
   `poisonlength`. With positive maximum and length, it rolls the poison rate
   through the existing deterministic damage roll and emits a typed poison
   channel. The ordinary resistance policy from #461/#462 mitigates that rate
   once, before state application. No poison amount enters the immediate-health
   total. The rolled and mitigated rates remain on
   `d2legacy.combat.damage_bundle`; add a separate `poison_length_frames` stage
   field. Other producers may supply the same duration when their own issues
   define it. Zero duration never starts a poison state.
2. An effects-phase observer joins a factual `d2legacy.combat.event` with its
   damage bundle. It accepts only a landed event with positive mitigated poison
   and positive length, and resolves the target from the event's public target
   ID. A zero-immediate-damage poison hit is still valid. It emits one
   `d2legacy.state.request` for `poison`, with negative `hpregen`, one
   `exclusive_group`, `replacement_priority=mitigated_rate`, and
   `reject_lower_priority=true`. The timed-state owner decides same-frame and
   existing-state ranking. The application source ID includes the unique
   combat-event identity so even a repeat hit from the same attacker creates a
   new instance and resets its application tick. Preserve the event's public
   attacker ID in an explicit state attribution field; do not parse it from
   the source ID. The timed-state owner carries that field from request to
   instance. A lethal immediate hit does not leave poison on a dead target.
3. A pre-simulation poison consumer visits only active poison instances and
   their living targets. For an application at tick `t` with length `L`, it
   subtracts the mitigated rate on ticks `t+1` through `t+L` inclusive. The
   timed-state owner removes the instance in effects at `t+L`; no tick follows.
   The poison consumer runs before the existing Cleansing pulse in each frame.
   If Cleansing moves expiry to the current tick, timed-state expiry removes it
   in effects and there is no later poison tick. A target in town takes no
   poison life loss while the state clock continues; this town rule is inferred
   from the 1.14d corroborating implementation. Inactive or dead targets take
   no poison ticks. Each lethal tick emits the ordinary `unit_died` combat
   event with the stored attacker ID, so existing player/monster death owners
   perform attribution and rewards exactly once.

Monster life already uses raw units. Player vitals currently store whole life;
add a bounded raw poison remainder to the existing vitals owner, rather than a
second health authority. Each tick adds `X` to the remainder, subtracts each
complete 256-unit life point from `health`, and retains the subpoint remainder.
Tests inspect both fields to prove an exact `X` raw reduction each frame.
Clamp at zero and clear the remainder on poison death. Other existing
whole-life damage and healing contracts stay in place; combined fractional
damage/healing ordering remains an original-runtime evidence gap.

## Presentation

The local `gameplay/world.lua` state snapshots already expose authoritative
timed instances. The connected world-view adapter currently projects aura
relationships only. Extend that narrow projection to publish the active
poison state for visible public targets, without sending its rate, attacker,
source, or expiry. Keep an explicit aura/non-aura distinction so poison does
not enter the aura-overlay rotation. The state presentation adapter reads the
owned row's `green` flag and resolves a green actor tint; its RGB approximation
is a presentation choice, not an asserted original palette transform. The
world screen applies the tint to the controlled hero, remote players, and
monsters while the state is present, and restores neutral tint on removal,
replacement, world change, or actor reuse. The existing `poisonhit` application
overlay remains record-driven.

## Verification and review

Use test-driven development around the existing systems. Cover positive and
zero length, zero/immune mitigated rate, stronger/equal/weaker replacement,
same-attacker refresh, multiple applications in one frame, exact `L` ticks,
player subpoint accumulation, monster raw life, death credit, town suppression,
Cleansing shortening a real poison instance, and tint apply/reset for all actor
classes. Exercise the connected state projection and reject malformed or
private state data. Include a melee acceptance case with authored poison stat
sources and player/monster resistance. Run focused Lua/Go tests, the relevant
owned-data test, `make test`, `make architecture`, `make test-lua-hardening`,
`make test-race`, and `git diff --check` on the submitted commit. Record any
game-data skips and remaining 1.14d runtime uncertainty in the full PR
template. Complete the issue only if every #463 criterion is exercised through
the production path; otherwise reference it and leave it open.
