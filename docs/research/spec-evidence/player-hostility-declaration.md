# Hostility declaration evidence

This companion supports [the behavioral contract](../../specs/player-hostility-declaration.md). Original binaries are identified through the [source manifest](../source-manifest.md). Function descriptions are analyst labels unless explicitly identified as original exports. RVAs are module-relative; preferred VAs aid navigation and are not runtime-base assumptions. Private source collection details and complete originals are preserved separately. The hit-time consumer of the relation flag is recorded in the [PvP damage and kill-credit companion](combat-pvp-damage-and-kill-credit.md).

## Original function references

| Artifact | Function RVA / original location | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| `LOD113C-D2GAME` | VA 0x6fd11130 (rva 0xf1130); checks 0x6fd1115d, 0x6fd11174, 0x6fd11138-0x6fd1114d, 0x6fd1120d, 0x6fd1122c-0x6fd11253; write sites 0x6fd102f0, 0x6fd105b0 | none | declare-hostility handler | HOS-R02, HOS-R03, HOS-R01 | Historical decompiler lead 2026-10-11/12; actual instruction read 2026-10-15 below |
| `LOD113C-D2GAME` | 0xf10f0 | none | RemovePlayerFromParty | HOS-R02, HOS-R03 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xf02a0 | none | CheckPlayerInteractionFlags (relation mask test) | HOS-R02, HOS-R04; hit-time consumer PVP-R05a | Historical 2026-10-11/12 disassembly/read record; limitations below |

## Import/export correspondence and targeted readings - 2026-10-15

Original D2Game.dll SHA-256 24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53 and D2Common.dll SHA-256 59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10 were rechecked against the available analysis specimens. This establishes current artifact identity; it does not identify DLLs loaded by an earlier gameplay session. PE import slots and matching target export tables establish:

| D2Game thunk | Import ordinal | D2Common target RVA | Role and support |
|---|---|---|---|
| 0xA3E2 | 10587 | 0x38C20 | Unit own-array stat read; initial declaration checks use 0x38B00, not effective-array selection |
| 0xA42A | 10973 | 0x38B70 | Unit effective stat read; repeated target level check consumes RLI-R02a/RLI-R02b |
| 0xA478 | 10331 | 0x2FE10 | Unit room getter; HOS-R02 (also PVP-R06) |
| 0xA808 | 10057 | 0x3C390 | Town predicate; HOS-R02 (also PVP-R06) |
| 0xA4C0 | 10920 | 0x2EB70 | Player data access; relation helpers; rejects absent/non-player unit by assertion |

D2Game import slot RVA 0xF8D10 identifies KERNEL32.GetTickCount. Its supplied u32 samples are an external time input, not a gameplay RNG draw. Original-image instruction decoding corroborates Game 0xF10F0..0xF12BC, 0xF0090/0xF00E0 and 0xF02A0/0xF02F0/0xF0640, including previously encoded instruction runs.

## Dispatch and output anchors

1. Player-relation command handler Game 0xCC8E0 (7-byte command) -> relation-operation dispatcher 0xF0A70 -> table RVA 0xFA3B0, slot 4 at 0xFA3C0 containing 0x6FD11130 -> declaration entry 0xF1130. Alternate 6-byte command handler 0xCC860 reaches the same dispatcher. The dispatcher looks up the target id in the game's player hash and passes the acting player separately. Instruction and table checks establish acting player in the handler's second register argument and target in its first stack argument; these storage details do not constrain a reconstruction's API.
2. Declaration 0xF1130 has room/town/level guards, relation-node requests 0xF02F0 twice, relation read 0xF02A0, timestamp read 0xF00E0 and store 0xF0090, then accepted activation 0xF0FF0 and cleanup 0x386F0. Its rejection write at 0xF12B2 is the initial rejection output anchor. Relation flag mutation is 0xF0640, notifications 0xF05B0. The full activation/party/cleanup/message callees are not completely verified here; HOS-R03/HOS-R01/HOS-Q01 retain that integration.

## Semantic field correspondence

| Semantic term | Original encoded field/condition | Rules |
|---|---|---|
| relation nodes | PlayerData+0x38; node otherId+0,flags+4,timestamp+8,next+0x10; eligibility mask 0x08 | HOS-R02 (also PVP-R05a) |

These encodings explain the evidence and are not implementation requirements. Original binary fixtures and exact earlier text are retained privately with SHA-256.

## Corrections and evidence limits

### Declaration verification

The original record said the handler was absent from the earlier source view and used only a decompiler reading. The handler entry 0xF1130 is inside the listing beginning at 0xF10F0; original bytes corroborate the complete handler instruction range. Command lookup/table dispatch establishes acting/target roles. HOS-R02 retains verified local numeric/room/time decisions and explicit request boundaries. The 60000-ms timestamp addition is in the declaration handler itself, not the preceding remove-from-party entry 0xF10F0. Initial level/room rejection writes 1; time refusal instead sends requests and leaves that rejection output unchanged. Full enable/disable, party removal, cleanup and message effects remain HOS-Q01; HOS-R03 moves from recovered to unresolved integration on located source, and HOS-R01 remains unresolved for complete command behavior.

### Relation raw-result fixture correction

The original PVP-R05a/PVP-V12 already identified eligibility mask `0x08` and reported a boolean hostile result, so no original numeric vector output changes here. Common relation consumption in Game `0xF02A0` returns `node.flags AND 8`: raw 0 or 8, not normalized 1. During this migration's new HOS-V02 draft, its equality case initially supplied an impossible raw result 1/request 1; before stable handoff it was corrected to raw 8/request 8 with the node's eligibility flag set. This affects only that new supplied-input case. Existing old boolean cases remain true/false as written; the declaration comparison must use the raw result. The complete original snapshot and pre-handoff edits are retained privately.

### Declaration level-source clarification - 2026-10-15

Full pass 3 exposed that HOS-R02's earlier phrase "target level checked again" did not distinguish the actual source arrays. Reusing the already complete Game handler instruction record: initial actor check at `0xF1158` and target check at `0xF116F` call thunk `0xA3E2`, import #10587, Common `0x38C20`; the later target check at `0xF11B0` calls thunk `0xA42A`, #10973, Common `0x38B70`. The import identities were already resolved against the exact images. A narrow complete read of Common `0x38C20` (96-byte function; no checked source reconstruction listed) confirms unit/list/id/definition guards followed by `0x38B00`, which always selects the own array. Common `0x38B70` instead selects through `0x38A80`, whose own/resolved behavior is the existing RLI-R02a contract. Compatible full helper/minimum records remain dated 2026-10-14; this does not renew array-construction or aggregation proof. The currently available Common image was rehashed to the manifest's unchanged 679936-byte SHA-256 `59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10`.

HOS-R02 and its input/fixture preamble now name initial own-array and later effective reads separately. Existing declaration fixtures supply their same level in both arrays, retaining every earlier result. HOS-V02(i) supplies immutable target own 9/effective 8 with usesResolvedStats=true and minimum adjustment disabled: initial checks pass, the later check returns early. No setter, changing getter return, concurrent mutation or guessed node-creation effect is invented. This is a semantic input/source clarification and therefore requires another fresh full reader even after full pass 3's zero numeric mismatches.

### RNG

The historical 2026-10-12 RNG audit for the source specification covered the declaration checks as part of its owned local slice; HOS-R02 states that the checks consume no random value, and HOS-V03 asserts it with the HOS-V02 fixtures and explicitly supplied no-draw operations. Tick samples are external time inputs, not draws.

## Compatible data and dependency records

The stat 12 identity consumes GDF-R02 and its table records for effective itemstatcost.txt; the underlying records retain their actual dates, effective-loader profile and limits. Own and effective stat reads consume RLI-R02a/RLI-R02b. Pending HOS-O01 has no measured matching case.

<a id="gap-assessment-and-migration"></a>

## Gap assessment and migration - 2026-10-15

| Question | Assessment and outcome |
|---|---|
| HOS-Q01 | Newly located embedded handler, command/table caller and timestamp helpers establish roles/local checks; complete side-effect callees remain incomplete, so no integrated declaration closure |
| HOS-Q02 | Pending gameplay coverage for HOS-R02; no gameplay actions authorized in this migration. Formerly part of the source specification's general gameplay-confirmation question |

<a id="observation-procedure-hos-o01"></a>

## Observation procedure HOS-O01

Pending; no observation recorded. On original 1.13c, use two player characters A and B in one game, with recorded own and effective levels on both sides of 9 (8 and 9). (a) With A in the field, request hostility toward B and read the rejection result. (b) With A in town and either character at level 8, repeat. (c) With A in town and both at level 9, request hostility and record the refusal chat, notifications and A's resulting directional relation flag toward B; disable it, then re-request within 60 seconds and after 60 seconds, recording whether the refusal chat appears. Expected local readings follow HOS-V01 and HOS-V02; stored enable/disable effects remain HOS-Q01 and are not expected values.

## Review history

The declaration rules were reviewed while they were part of the [PvP damage and kill-credit specification](../../specs/combat-pvp-damage-and-kill-credit.md). The full dated writer review and blind-implementer records remain in that specification's companion with current ids: [writer review](combat-pvp-damage-and-kill-credit.md#writer-review), [full pass 1](combat-pvp-damage-and-kill-credit.md#blind-implementer-full-pass-1---2026-10-15), [full pass 2](combat-pvp-damage-and-kill-credit.md#blind-implementer-full-pass-2---2026-10-15), [full pass 3](combat-pvp-damage-and-kill-credit.md#blind-implementer-full-pass-3---2026-10-15), [full pass 4](combat-pvp-damage-and-kill-credit.md#blind-implementer-full-pass-4-and-root-review---2026-10-15) and [delta pass 5](combat-pvp-damage-and-kill-credit.md#blind-implementer-delta-pass-5---2026-10-15). Findings that concerned this specification's rules:

| Pass (2026-10-15) | Finding | Disposition |
|---|---|---|
| Full 1 | DISCRIM 7: signed timestamp comparison | Added HOS-V02(f) |
| Full 1 | GATED 2: command validation/stored relations/messages/cleanup | No change: HOS-Q01 |
| Full 1 | HARDEST 2: unsigned timestamp/equality/second tick | Added HOS-V02(f); HOS-V02(b) retained |
| Full 2 | DISCRIM 7: timestamp sum overflow | Added HOS-V02(g) |
| Full 2 | GATED 1 (grouped, including HOS-Q01) | No change |
| Full 2 | HARDEST 2: HOS-V02(f), unsigned time | No change; HOS-V02(g) added separately |
| Full 3 | DISCRIM 4: raw 8/request 0 | Added HOS-V02(h) |
| Full 3 | DISCRIM 7: repeated target level below 9 | Own/effective clarification above; HOS-V02(i) added |
| Full 3 | GATED 2: HOS-Q01 | No change |
| Full 3 | HARDEST 3: HOS-R02 unsigned/equality/ticks/wrapping | No change |
| Full 4 | DISCRIM 3: initial actor own/effective distinction | Added HOS-V02(j) |
| Full 4 | DISCRIM 4: disable bypasses repeat target-level guard | Added HOS-V02(k) |
| Full 4 | GATED 2: HOS-Q01 | No change |
| Full 4 | HARDEST 1: HOS-V02(i) | No change; HOS-V02(j)/(k) added |
| Delta 5 | HOS-V02(j), HOS-V02(k) changed cases | MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0 |

## Restructure (2026-10-15)

The approved specification review split the hostility-declaration request out of [PvP damage, hostility and kill credit](../../specs/combat-pvp-damage-and-kill-credit.md) into this specification with prefix HOS. Moved here: the command-role, declaration-check and side-effect rules, the relation-meaning rule, the declaration integration question, the low-level rejection and timing vectors, the declaration rows of section 3 and section 5, the level and delay constants, the declaration dispatch anchors, function and import rows, and the declaration correction records above. New in this file, derived from existing content without changing behavior: HOS-V03 carries the declaration slice of the source no-draw vector; HOS-Q02 and HOS-O01 carry the declaration part of the source general gameplay-confirmation question and observation. No rule behavior, label, value or vector result changed. The id correspondence is kept outside the repository for the follow-up reference update. Neither specification is consumed, so both are numbered cleanly. No source campaign, implementation test or gameplay observation ran for this restructure, and the new specification has not had its own blind-implementer pass.
