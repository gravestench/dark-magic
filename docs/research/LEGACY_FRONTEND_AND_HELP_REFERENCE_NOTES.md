# Legacy front-end and in-game help reference notes

Status: visual/behavioral reference captured from user-supplied Diablo II Lord
of Destruction 1.14d screenshots on 2026-08-13. The screenshots are not stored
in this repository. These notes record observable behavior and layout, not an
instruction to copy proprietary art into Dark Magic.

## Main menu multiplayer hierarchy

The expansion main menu presents distinct controls in this order:

1. `Single Player`
2. `Battle.net`
3. a smaller `Gateway: <selected gateway>` control directly beneath Battle.net
4. `Other Multiplayer`

`Credits` and `Cinematics` share a smaller row near the bottom, followed by the
full-width exit button. Dark Magic currently lacks the separate Battle.net and
gateway affordances. Do not collapse these into the existing generic
multiplayer entry: Battle.net begins the account/realm flow, while Other
Multiplayer represents self-hosted/local multiplayer choices.

The gateway label is both a status display and an actionable control. Its
selected value must be durable client preference, not authoritative character
or game-session state.

## Gateway selection modal

Activating the gateway control opens a centered modal over the still-visible
main menu. It contains:

- a `Select Gateway` heading;
- explanatory copy about choosing a gateway appropriate to the player's
  location;
- a single-selection list with a highlighted row (the reference shows U.S.
  West, U.S. East, Asia, and Europe);
- separate `OK` and `Cancel` buttons.

Dark Magic's service endpoints need not reproduce Blizzard's historical
regions, but the interaction contract remains useful: selection is explicit,
cancelable, and committed only by confirmation. Modern endpoint discovery,
latency, trust, and availability data can back this presentation without
making the UI understand transport details.

## Battle.net connection progress

Choosing Battle.net opens a blocking, centered, cancelable progress modal over
the main menu. The reference exposes a high-level connection phase such as
`Checking Versions`, plus an animated ellipsis and `Cancel`.

Dark Magic should expose stable user-facing phases rather than raw socket logs.
Candidate semantic phases are:

1. resolving gateway;
2. connecting securely;
3. checking client/engine compatibility;
4. negotiating and verifying mod manifests;
5. downloading/verifying required redistributable mods;
6. authenticating the account;
7. loading the account character roster;
8. entering realm browse/create/join.

Each phase must be cancelable, must redact credentials and sensitive endpoint
details, and must map failures to localized structured reasons. Detailed
network diagnostics belong in opt-in developer logs, not the ordinary modal.

## Select Cinematics modal

The reference uses a tall centered `Select Cinematics` modal over the main-menu
background rather than presenting a separate full-screen menu. It contains one
large framed row per unlocked cinematic and one centered `Cancel` button. Rows
use the ordinary front-end button treatment; unavailable entries can remain
disabled or blank according to verified legacy behavior.

Dark Magic already supports selectable cinematics and optional-content
handling. A later fidelity pass should compare its scene/modal composition,
dimensions, spacing, title, row styling, cancellation, and unlock policy with
this reference.

## In-game Help overlay

The Help screen is a nearly full-screen translucent overlay on top of the live
gameworld. It does not replace the world scene. Its observable structure is:

- an ornate border around the usable screen;
- a centered `Diablo II Help` heading;
- a close icon/control in the upper-right;
- a readable bullet list of core keyboard controls across the upper portion;
- the live HUD still visible beneath the overlay;
- labeled callouts with leader lines anchored to actual HUD elements, including
  life/mana orbs, mouse skills, new-stat/new-skill indicators, stamina,
  run/walk, experience, mini-panel, and belt.

The current Dark Magic overlay contains much of the semantic content but is not
yet visually faithful. The follow-up should preserve data-driven/localized
labels while fixing border coverage, translucency, typography, close control,
leader lines, and anchors derived from the active HUD layout. Hard-coded text
coordinates that can drift independently from HUD composition are not an
acceptable final architecture; callouts should reference named HUD anchors and
be tested across supported presentation profiles and resolutions.

## Testing expectations

For these screens, tests should cover semantic state and composition without
checking proprietary pixels into the repository:

- main-menu order and separate Battle.net/gateway/Other Multiplayer actions;
- confirmed versus canceled gateway preference changes;
- every connection phase, cancellation, failure mapping, and secret redaction;
- cinematic row availability, selection, playback return, and cancellation;
- Help overlay blocking/toggling behavior, localized bullets, named HUD-anchor
  resolution, and representative screenshot/composition captures using an
  owned local asset installation.

