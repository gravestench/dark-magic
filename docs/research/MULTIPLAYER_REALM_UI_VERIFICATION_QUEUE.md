# Multiplayer/realm/UI verification queue

This queue consolidates empirical work for parties/PvP/trade, Dark Magic
networking behavior, Realm services, and UI-visible gameplay contracts. Vanilla
protocol and old community-tool interoperability are out of scope.

Current Dark Magic evidence includes a versioned, bounded, authenticated-owner
party projection carried by `ClientView/v5` and consumed by the Lua party
panel. This proves the authority/projection boundary; it does not answer the
queued expansion 1.14d roster-field, event-timing, or action-layout probes.

## P0: architecture-shaping probes

- Bind two synthetic remote principals to one authoritative Dark Magic `Session` and prove local and remote movement commands produce identical replay/checkpoint results.
- Define a canonical `SessionPlayerID`/CharacterID/AccountID separation and spoofing rejection vectors.
- Capture one expansion 1.14d party invite/accept/leave sequence including relationship/roster events.
- Capture one expansion 1.14d hostility declaration including level/town restrictions, cooldown timing, target party propagation and declarer party removal.
- Capture one complete two-player trade from initiation through offer changes, dual acceptance, item/gold exchange and cancellation-on-change behavior.
- Build a character lease/CAS simulation proving two game workers cannot commit the same durable revision concurrently.
- Define the first versioned per-client world/player snapshot and prove private inventory/hidden server facts are filtered.
- Map current Lua HUD/item/quest/interaction reads to explicit semantic view models and remove one direct/raw state dependency if any remains.

## P1: party/social

- party invite/cancel/accept state transitions;
- party size/ID lifecycle;
- leader semantics if any;
- same-level/proximity eligibility for XP/quest/gold/NoDrop;
- party XP formula and member-level weighting;
- gold sharing range/overflow;
- party quest-credit representative cases;
- player location/health visibility to party UI;
- Hardcore corpse-loot permission directionality/persistence;
- shared realm/in-game whisper, ignore/squelch, presence, friends, alias, and
  command-history behavior described in
  [REALM_AND_GAME_CHAT_COMMANDS.md](REALM_AND_GAME_CHAT_COMMANDS.md);
- disconnect/reconnect party state.

Party XP capture tooling is ready at `internal/dev/tools/party_xp_probe` with a
version-locked template under `docs/research/probes/`. Required paired cases:
equal-level neutral/party baseline; two, three, and eight same-area members;
distance brackets around the observed cutoff; unequal member levels that
distinguish Blizzard's documented direct raw share from an inverse final award; player/
monster level differences 5, 6, 9, and 10; odd XP pools that distinguish floor,
nearest, and ceiling; killer-owned summon attribution; dead member; different
named area; join/leave before spawn versus before death. Keep the monster,
difficulty, connected roster, and character levels identical within each
baseline pair.

## P1: hostility and PvP

- expansion 1.14d minimum level/town restrictions;
- 60-second hostility delay semantics and reset;
- unhostile/remove-hostility rules;
- hostility propagation to target party;
- same-party hostility/automatic leave;
- town combat/skill/missile restrictions;
- PvP damage scalar/order;
- PvP life/mana leech;
- crushing blow/Open Wounds/poison/stun/hit-recovery;
- minion/hireling/trap hostility and attribution;
- PvP death gold/XP/corpse/ear rules;
- hostile portal/waypoint/party interactions.

## P1: player trade

- trade initiation location/range/town requirements;
- trade request/accept/cancel state machine;
- offer inventory/grid rules;
- item/gold offer update semantics;
- acceptance reset after offer change and any delay;
- item visibility/tooltip identified state to counterparty;
- gold limits/overflow;
- destination inventory-full behavior;
- quest/nontradeable/socket-parent/corpse/held restrictions;
- durability/charge/personalization preservation;
- death/hostility/zone transition during trade;
- disconnect cancellation/return;
- simultaneous action ordering/race behavior.

## P1: modern network/session

- command tick lead/lag policy under latency/jitter;
- sequence replay/duplicate rejection;
- deterministic same-tick ordering across clients;
- interest-management room/level rules;
- snapshot/delta schema and reconciliation;
- semantic event ack/recovery;
- movement prediction correction thresholds;
- skill/combat prediction policy;
- reconnect grace period;
- game checkpoint persistence/restore;
- stale GameRules/content fingerprint rejection;
- rate limits and malformed message handling;
- admin/system command isolation/audit.

## P1: realm/persistence control plane

- character create/delete/rename/convert rules;
- expansion 1.14d Hardcore/Ladder eligibility compatibility;
- per-account character count/list order;
- game create/list/join metadata and password/private behavior;
- game-worker registration/capacity/allocation;
- character lease acquire/renew/release timeout;
- CAS durable commit and stale revision behavior;
- server crash during active lease;
- disconnect versus explicit game leave;
- game idle/lifetime/destruction;
- durable save extraction from transient checkpoint;
- ladder update from committed revision;
- realm-wide special-event delivery;
- audit/admin/moderation surfaces.

## P1: realm chat and channels

- authenticated realm-wide public channel message routing;
- channel member joins/leaves and full character-composite presence projection;
- realm-scoped public/private/operator channel naming and membership;
- typed whisper, reply, ignore/squelch, away, DND, and friends operations;
- channel operator designate/resign/kick/ban/unban authority and audit;
- privacy-filtered who/whois/users/presence projections;
- command alias normalization and target-patch feature gating;
- independent flood limits for public chat, whispers, presence, and channel churn;
- structured red error/system responses and legacy color mapping;
- continuity across lobby panels and teardown/reconnect behavior;
- separation of realm social, game chat, local diagnostic, and authoritative
  gameplay command dispatch.

## P1: UI-visible gameplay state

- HUD life/mana/stamina/XP rounding;
- character sheet derived stat display;
- skill tooltip calculations and client/server differences;
- item tooltip ordering/localization;
- identified/unidentified information;
- set/runeword/socket property display;
- ground item label visibility/ownership;
- quest log state/objective transitions;
- waypoint list ordering/locks;
- NPC dialogue choice ordering;
- vendor price/stock refresh presentation;
- party roster location/health/hostility fields;
- trade item/acceptance/status fields;
- automap exploration/party/portal/quest markers;
- target monster name/HP/unique/immunity information;
- death/corpse/Hardcore UI;
- game list/create/join fields;
- command rejection/error messages;
- prediction/correction presentation.

## P2: social/realm extensions

These are useful modern features but not baseline Diablo gameplay blockers:

- friends/presence service;
- account-wide stash/profile if a later target/mod wants it;
- matchmaking/regions/worker autoscaling;
- moderation/reporting;
- replay spectator service;
- server browser richer filters;
- modern encryption/session-token transport;
- web/admin dashboards.

Keep them behind semantic realm/game APIs rather than pushing them into simulation.

## Probe artifact standard

Each completed probe should record:

```text
probe ID/title
target client/server/content version
topology/accounts/characters/game setup
exact command/action sequence
raw packet/log/save capture when lawful/needed
normalized state transitions
server tick/time relationships
security/privacy expectations
conflicts with older secondary sources, if relevant
confidence upgrade
primary research doc updated
```

Keep credentials/private account data and proprietary captures outside Git. Commit sanitized schemas, semantic traces, hashes, fixtures and test vectors.
