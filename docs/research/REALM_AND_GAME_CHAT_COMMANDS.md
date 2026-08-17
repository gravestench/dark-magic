# Realm and in-game chat commands

Status: behavioral baseline for the Diablo II realm lobby and in-game chat
command surface. Exact patch behavior remains subject to owned-client probes.

This document separates commands by **execution owner**. A shared text parser may
recognize the command vocabulary, but it must not turn every slash command into a
gameplay command or send every command to the realm service.

## Source policy

The supplied command catalog matches the Diablo Wiki/Fandom
[Game commands](https://diablo.fandom.com/wiki/Game_commands) article closely.
It is useful discovery and corroboration evidence, not an exact compatibility
contract. The broader [Diablo II](https://diablo.fandom.com/wiki/Diablo_II) wiki
and related articles are approved research leads moving forward.

The maintainer confirmed that `%` characters in the pasted catalog were a
copy/paste formatting error and should be read as `/`. Preserve that correction
as provenance for the supplied observation; do not silently use the corrupted
text as a command grammar.

Prefer, in order:

1. repeatable behavior from an owned target-version client;
2. contemporary Blizzard manuals and classic Battle.net documentation;
3. version-pinned implementation evidence;
4. corroboration across independent community references;
5. an uncorroborated wiki, guide, forum post, or player recollection.

The official Battle.net
[chat commands](https://classic.battle.net/info/commands.shtml),
[friends commands](https://classic.battle.net/info/friends.shtml), and
[channel behavior](https://classic.battle.net/info/channels.shtml) corroborate
much of the supplied list. Patch additions must be gated by the selected legacy
rules version. For example, the Diablo Wiki records `/home`, persistent ignores,
and message filters as 1.13d additions rather than universal Diablo II behavior.

Do not copy wiki prose or data into runtime content. Convert verified behavior
into original typed contracts and tests, and retain the evidence link/version in
the implementing test or research note.

## Ownership model

```text
chat input
  -> local editor/history shortcuts
  -> command lexer + alias normalization
       -> realm social command
       -> current-game social command
       -> local presentation/diagnostic command
       -> authoritative single-player game-rule command
       -> ordinary public/party/whisper message
```

The UI owns text editing, selection, clipboard actions, history, and local
presentation toggles. The authenticated realm connection owns realm presence,
channels, friends, whispers, ignore policy, and moderation. The active game
connection owns in-game message routing. The authoritative session owns any
command that changes simulation rules. No chat text is executed as Lua.

## Shared social commands

These are useful from the realm lobby and, where the selected service/version
supports them, from the in-game chat box:

| Canonical command | Known aliases | Semantic request |
| --- | --- | --- |
| `/whisper <recipient> <message>` | `/w`, `/msg`, `/m`, `/whisper` | Send an authenticated private message. Recipient syntax may identify a character, account, or realm-qualified character. |
| `/reply <message>` | `/reply`; additional aliases require a probe | Reply to the most recent eligible whisper sender. |
| `/ignore <user>` | `/ignore`, `/squelch` | Add an identity to the caller's ignore policy. |
| `/unignore <user>` | `/unignore`, `/unsquelch` | Remove an identity from the caller's ignore policy. |
| `/whois <user>` | `/whois`, `/where`, `/whereis` | Request the privacy-filtered presence/profile projection for a user. |
| `/whoami` | `/whoami` | Request the caller's own realm identity projection. |
| `/users` | `/users` | Request aggregate service counts. Exact output fields are version-sensitive. |
| `/time` | `/time` | Display local time and, when connected, realm/server time. |
| `/away [reason]` | `/away` | Set or clear away presence with an optional response reason. |
| `/dnd [reason]` | `/dnd` | Set or clear do-not-disturb policy and its automatic response. |

Recipient resolution is server-side. Never let a client assert another user's
AccountID or CharacterID. Preserve account names, character names, display names,
and internal IDs as distinct types.

## Realm-lobby-only commands

| Command family | Semantic owner and behavior |
| --- | --- |
| `/join <channel>`, `/channel <channel>` | Realm channel membership transition. Diablo II channels are realm-scoped. |
| `/who <channel>` | Privacy-filtered channel membership query. |
| `/rejoin` | Leave/re-enter the current channel through ordinary membership transitions. |
| `/me <text>` | Realm channel action/emote event. |
| `/designate`, `/resign`, `/kick`, `/ban`, `/unban` | Realm moderation commands requiring current-channel operator authority. |
| `/friends ...`, `/f ...` | Realm-owned ordered friend list: help, add, remove, list, message, promote, and demote. |
| `/options ...`, `/o ...` | Account/session social filters for public, private-channel, and whisper traffic. |
| `/d2notify` | Toggle client presentation of channel enter/leave events; persistence and server participation require a probe. |
| `/home [channel]` | Later-patch home-channel behavior; do not expose under an earlier compatibility profile. |

The supplied catalog's `/clan <name>` wording conflicts with Blizzard's documented
`/join Clan <name>`/`/join Op <account>` channel conventions and needs a
target-client probe. `/stats` is a cross-title Battle.net command and is not a
meaningful Diablo II gameplay feature.

Friends are account-owned realm social state, not character-save state. The
classic documentation reports a 25-entry limit. It documents numbered friend
references as `%f<number>`, while the maintainer-confirmed correction to the
supplied observation produces a slash-based form. Treat these as conflicting
evidence and leave numbered friend shorthand unimplemented until an owned-client
probe establishes its exact grammar.

## In-game-only and local commands

| Command | Owner | Required behavior |
| --- | --- | --- |
| `/fps` | local presentation | Toggle frame rate and network-latency diagnostics without entering deterministic state. |
| `/framerate` | local presentation | Toggle the expanded performance/memory diagnostic overlay. Exact fields vary by renderer/platform and compatibility target. |
| `/nopickup` | player preference + authoritative input policy | Toggle automatic ground-item pickup behavior while preserving explicit show-item pickup. Persist only if verified or deliberately adopted as a Dark Magic preference. |
| `/players <1..8>` | host-authorized served game | Force effective player-count difficulty/drop policy without changing the game's admission cap. Raw remote players cannot mutate it directly. |
| `/soundchaosdebug` | local development/presentation | Debug-only sound sweep; it must not be accepted as an untrusted remote gameplay command. |
| `/time` | local/connected presentation | Display local and connected server time as available. |

`/players` is a typed, validated authority request backed by revisioned
`d2legacy.player_count/v1` state rather than a Lua global. In the absence of an
override, gameplay follows present server players and changes as they join or
leave. The override survives checkpoints until another value replaces it or
the host returns the game to population-following behavior. `maximum_players`
remains only an admission cap.

## Keyboard and editor behavior

The supplied catalog and Blizzard documentation identify these lobby-chat
shortcuts:

- `Ctrl+X`, `Ctrl+C`, `Ctrl+V`, and `Ctrl+A` operate on the text editor;
- `Ctrl+N` and `Alt+N` insert the selected channel member's display name;
- `Alt+W` begins a whisper to the selected member;
- `Tab` cycles recent commands in the Diablo II lobby;
- `Alt+V` mirrors the join/leave-notification toggle;
- `Ctrl+M` toggles frontend/Battle.net music.

In game, opening chat and pressing Up/Down cycles that game's sent-input history.
History and selection are per-context presentation state. Passwords, tokens, and
other secret fields must never enter chat history.

## Typed contract sketch

The future realm capability should expose semantic operations and immutable
events rather than a single `execute_command(text)` escape hatch:

```text
RealmChatIntent
  SendChannelMessage(channel_id, text)
  SendWhisper(recipient_reference, text)
  JoinChannel(channel_reference)
  SetIgnore(subject_reference, ignored)
  SetPresence(status, reason)
  MutateFriendList(operation, subject_reference, position)
  ModerateChannel(operation, subject_reference)

RealmChatEvent
  ChannelMessage
  Whisper
  ChannelAction
  MemberJoined / MemberLeft
  PresenceChanged
  SystemNotice
  CommandRejected(code, localized_arguments)
```

The parser maps text/aliases onto those intents. Buttons such as Whisper,
Squelch, Unsquelch, Channel, and Help invoke the same semantic operations without
manufacturing command strings.

In-game chat should use equivalent typed message events routed through the game
connection or a realm social bridge. Chat remains outside deterministic replay
checksums unless a typed gameplay command such as an allowed `/players` change
is produced.

## Security and moderation requirements

- bound message, recipient, channel, reason, and command lengths before routing;
- normalize commands case-insensitively without normalizing message bodies;
- rate-limit chat, whisper, presence, channel changes, and moderation separately;
- perform ignore/filter decisions on authenticated identities, not display text;
- prevent markup/control/color injection into UI rendering and game names;
- return stable rejection codes and localized arguments, never raw internal errors;
- audit moderation and administrative actions without logging private message text
  by default;
- apply capability/role checks on the service even when the UI hides a command;
- preserve realm chat during lobby UI transitions, but begin gameplay networking
  only after game assignment and selected-character admission are ready.

## Verification queue

1. Capture the lobby Help command list and aliases from each supported target
   client version.
2. Verify whether shared whisper/ignore/friends commands work from Diablo II
   in-game chat and whether responses return through the realm or game transport.
3. Probe recipient grammar for character, account (`*account`), realm-qualified
   character, selected-user insertion, quoting, whitespace, and case behavior.
4. Verify `/reply` sender selection, offline behavior, and interaction with DND,
   ignore, and cross-realm whispers.
5. Verify exact numbered friend-reference syntax (the corrected slash-based
   supplied observation versus Blizzard's documented `%f5`), ordering, limits,
   privacy, and notification behavior.
6. Verify public/private/op channel creation, naming, capacity, operator transfer,
   kick/ban lifetime, and rejoin behavior.
7. Verify per-command maximum lengths, history size, repeat behavior, and error
   colors/text for the target patch.
8. Verify `/nopickup`, `/fps`, `/framerate`, `/players`, and `/time` state,
   persistence, availability, and multiplayer rejection behavior.
9. Decide which later-patch commands (`/home`, persistent ignore, filters) belong
   in each Dark Magic compatibility profile.
10. Verify chat flood limits and disconnection/mute behavior without using the
    production service as a load target.
