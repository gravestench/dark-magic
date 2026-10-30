# Music selection and transitions

**Description**

The game coordinates background music as the player moves between areas. This contract establishes the area-song request decision and a playback bookmark calculation. Audible track changes, quest interruptions, looping and fades remain subject to the evidence gates below.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md)
Ids: MUS-R, MUS-Q, MUS-V, MUS-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/sound-music-selection-and-transitions.md)
Open evidence gaps: [gap file](../plans/gaps/sound-music-selection-and-transitions.md)

## 1. Scope

Background music selection for levels, acts and towns through the sound environment; pending requests, playback-position bookmarks, quest or special cue interruptions, looping and fades. This is the in-game music boundary. Menu and cinematic music are outside this boundary. The supplied-input request decision and bookmark stages are verified below; complete audible reproduction is gated. Ambient and weather selection and general voice playback belong to the separate sound contracts.

## 2. Trigger and ownership

D2Client performs the music controller update after the ambient/environment update and before pending sound loads and voice processing in the client audio dispatch. The caller supplies its unsigned audio tick. The dispatch has callers in the game loop; its exact cadence and connection to server frames are not established. D2Sound supplies the stored master and music volume settings and the eventual stream position. The verified stages consume these results as inputs; their settings-to-device integration and the downstream audible result remain integration gates.

## 3. Inputs and data

| Input | Semantic source | Width and format | Meaning | Origin | Binding |
|---|---|---|---|---|---|
| current tick, area-entry tick, bookmark tick | client music state | unsigned 32-bit ticks | request delay and bookmark timing | runtime input | supplied |
| current requested song | client music state | signed 32-bit sound id; zero means no request | retained request | runtime input | supplied |
| candidate song | selected environment Song value | signed 32-bit id | prospective request before range filtering | table data | resolved from data; full loader integration gated |
| prepared sound count and ordered environment songs | sound/environment table initialization | count and each Song are signed 32-bit values | inputs to range initialization | table data | supplied prepared collection; full parser correspondence gated |
| song-range lower and upper bounds | MUS-R02a initialization over sound environments | signed 32-bit ids | inclusive accepted song range | engine code | computed by original code from prepared Song fields |
| environment available | environment lookup result | Boolean | candidate exists; false supplies candidate zero | runtime input | supplied |
| backend eligibility A and B | stored master volume and music volume settings, respectively | signed 32-bit integers | music controller permits playback when both are nonzero | configuration input | supplied query results; original setting identities fixed by code |
| playback position | sound-system position query | signed 32-bit backend position units | position compared with bookmarks | runtime input | supplied; unit conversion gated |
| block 1, block 2, block 3 | selected sound record Block 1, Block 2, Block 3 | signed 32-bit backend position units | ordered comparisons used by bookmark stage | table data | resolved from data; position correspondence gated |
| stored bookmark | per-song music state | signed 32-bit backend position units | bookmark retained between music requests | runtime input | supplied |
| bookmark eligible | controller stage eligibility | Boolean | song nonzero, voice found and voice active | runtime input | supplied; producer integration gated |
| existing requested-song voice | supplied result of first-voice lookup | handle, zero means absent | determines whether to submit a new request | runtime input | supplied result; voice lifecycle gated |
| music-range voice present | supplied fixed-range voice query | Boolean | controls the submission flag together with bookmark | runtime input | supplied query result |
| submission tick | client music state | unsigned 32-bit tick | retained time of the last area-song submission attempt | runtime state | supplied |
| supplied voice operations | first-voice lookup, music-range voice query, stop request and playback request | ordered returning operations | return the stated results without other music-controller state writes in stage vectors | external operation | effects beyond the request log remain gated |
| town song cell | effective data/global/excel/soundenviron.txt; Handle ESOUNDENVIRON_TOWN_1, Index 1, Song | decimal text 4673 | named archive cell | table data | unresolved compiled correspondence |
| wilderness song cell | same virtual file; Handle ESOUNDENVIRON_WILDERNESS, Index 2, Song | decimal text 4679 | named archive cell | table data | unresolved compiled correspondence |

Named cells have [T01](../research/spec-evidence/sound-music-selection-and-transitions.md#t01-effective-song-cells) evidence. They do not establish a filename, audible track identity, loop policy, or quest override. The block fields are runtime supplied in the vectors; no archive values for them are asserted.

## 4. Rules (execution order)

**MUS-R01** Named song cells [verified in 1.13c data]. The two named Song cells in section 3 contain their stated decimal values.

**MUS-R02a** Initialize the accepted song range [verified in 1.13c code (static)]. At the prepared-record initialization stage, set lower bound to the prepared sound count and upper bound to zero. Visit environment records in their existing order. Ignore Song values less than or equal to zero. For a positive Song, replace lower bound when Song is smaller, and replace upper bound when Song is larger, using signed 32-bit comparisons. Publish both bounds. No random draw occurs. The original profile supplies positive valid song identities; this stage does not repair an empty or malformed song set, validate asset existence, or establish how text becomes a prepared record.

**MUS-R02b** Filter a candidate song [verified in 1.13c code (static)]. At the supplied-input request stage, start with candidate zero if no environment is available. Otherwise take the environment candidate. Preserve zero. Replace a nonzero candidate with zero if it is below the lower song bound or above the upper song bound, using signed comparisons. Bounds are inclusive. This stage performs no random draw.

**MUS-R03** Commit a changed request [verified in 1.13c code (static)]. Consume the filtered candidate from MUS-R02b. Compute elapsed as unsigned 32-bit current tick minus area-entry tick, wrapping modulo 2^32. A change is permitted if the current request is zero, either eligibility query returns zero, or elapsed is at least 75. If the candidate differs from the current request and change is permitted, write the candidate as the current request and current tick as bookmark tick. Otherwise leave both fields unchanged. This operation makes no playback call and draws no random number.

**MUS-R04** Periodic bookmark calculation [verified in 1.13c code (static)]. This supplied-input stage runs only when bookmark eligible is true and unsigned current tick is strictly greater than unsigned (bookmark tick + 125 modulo 2^32). Otherwise leave stored bookmark and bookmark tick unchanged. When it runs, initialize next bookmark to zero and perform these tests in order, all comparisons signed:

1. If position is nonnegative and less than block 1, set next bookmark to block 1.
2. If position is at least block 1 and less than block 2, set next bookmark to block 2.
3. If position is at least block 2 and less than block 3, set next bookmark to block 3.

Write next bookmark to the current song's stored bookmark and current tick to bookmark tick. These are sequential tests, including when block values are not increasing. No random draw occurs. A supplied position does not establish how the backend reports position or how a future request uses this bookmark.

**MUS-R05a** Ordinary area-song submission attempt [verified in 1.13c code (static)]. Boundary: the ordinary submission branch after interruption handling has allowed area music to resume, with returning supplied voice operations. If current requested song is zero or either master/music setting is zero, make no lookup or request and retain submission tick. Otherwise query the first voice handle for the current requested song. A nonzero handle causes no further operation or submission-tick write.

When that handle is zero, read the stored bookmark and query whether a music-range voice is present. Select request flags 2 exactly when that query is false and bookmark is zero, otherwise flags 0. Request all music voices stop, then submit the current requested song with absent emitter, delay 0, the selected flags and playback value equal to the stored bookmark. Ignore the playback request result and write current tick to submission tick. The stage performs no direct random draw. Playback, seek interpretation, actual voice stopping and future retry eligibility remain dependency integration gates; the request log does not establish their effects.

**MUS-R05b** Submit, stop, loop and fade music [unresolved]. The complete behavior of changing the audible stream, reusing a bookmark, stopping an existing song, restarting after exhaustion and applying fade metadata is gated by MUS-Q01. No default crossfade, replay interval, seek unit or loop behavior is established.

**MUS-R06** Quest and special cue interruptions [unresolved]. The conditions that interrupt area music, the selected cue, its timing and return to area music are gated by MUS-Q02. A quest state or boss presence alone does not establish a track change.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| MUS-R01 | named song cells | verified in 1.13c data | none for text cells | T01 |
| MUS-R02a | prepared-record song range | verified in 1.13c code (static) | none for supplied prepared records; MUS-Q01 gates parser correspondence | MUS-V12, MUS-V13 |
| MUS-R02b | candidate filter | verified in 1.13c code (static) | none for supplied-input stage; MUS-Q01 gates production binding | MUS-V01, MUS-V02, MUS-V03 |
| MUS-R03 | delayed request write | verified in 1.13c code (static) | none for supplied-input stage; MUS-Q01 gates audible result | MUS-V04, MUS-V05, MUS-V06, MUS-V07 |
| MUS-R04 | bookmark calculation | verified in 1.13c code (static) | none for supplied-input stage; MUS-Q01 gates backend position and reuse | MUS-V08, MUS-V09, MUS-V10, MUS-V11 |
| MUS-R05a | ordinary area-song submission log | verified in 1.13c code (static) | none for supplied ordinary stage; MUS-Q01/MUS-Q02 gate voice effects and interruption entry | MUS-V14 through MUS-V18 |
| MUS-R05b | audible stream lifecycle | unresolved | MUS-Q01: full playback integration | none |
| MUS-R06 | quest interruptions | unresolved | MUS-Q02: interruption policy | none |

## 5. Outputs and side effects

| Output | Rule |
|---|---|
| initialized lower and upper song bounds | MUS-R02a |
| current requested song and bookmark tick changed together | MUS-R03 |
| stored bookmark and bookmark tick written together | MUS-R04 |
| voice query/stop/playback request log and submission tick | MUS-R05a |
| audible start/stop, seek, repeat and fade effects | MUS-R05b, gated |
| interruption cue and resumption effects | MUS-R06, gated |

MUS-R02b returns the filtered candidate; MUS-R01 only establishes data cells. Verified stages emit no server message and write no character save.

## 6. Branches and edge cases

Zero is preserved by the candidate filter. Bounds use signed comparisons, while request delay and bookmark cadence use unsigned 32-bit clock operations. Equality at 75 permits a changed request; equality at bookmark tick plus 125 does not update a bookmark. The bookmark tests do not sort block values. Negative backend positions can still satisfy later tests when negative blocks are supplied. Missing environment means candidate zero within the supplied-input stage; invalid runtime table pointers are not assigned fallback behavior. A zero master or music volume setting can permit a changed stored request without establishing an audible change. Difficulty and player class do not alter these supplied-input arithmetic stages; the unresolved producers may select different inputs.

## 7. Open questions

**MUS-Q01** [unresolved] Full level-to-stream integration: establish text/prepared/runtime correspondence for Levels.SoundEnv and environment records, settings-to-device integration, stream-position units and conversion, bookmark reuse, admission, stream end/restart, looping and fade application. This gates MUS-R05b, voice effects requested by MUS-R05a and production integration of MUS-R02b through MUS-R04, not their supplied-input calculations.

**MUS-Q02** [unresolved] Complete special and quest interruption producers, cue choices, eligibility, timing and resumption. This gates MUS-R06 and actual interruption-to-ordinary entry for MUS-R05a.

**MUS-Q03** [unresolved] Gameplay observations remain pending. Valid music-function and backend-query captures provide partial corroboration, but no complete vector capture or audible transition, delay, bookmark or loop assertion is runtime-confirmed. This observation limitation does not invalidate the verified supplied-input stages. See the [gap file](../plans/gaps/sound-music-selection-and-transitions.md).

## 8. Test vectors

### 8.1 Executable vectors

All cases are independent supplied-input stages. They have no random draw and perform only the state writes described by their named rules. Candidate filter vectors return a value; request vectors begin with the stated bookmark tick. Bookmark vectors supply an already evaluated eligibility Boolean and backend position, and persist only bookmark and bookmark tick. Other music stages do not run inside an individual vector.

- **MUS-V01** Inclusive lower bound (MUS-R02b; boundary: supplied-input stage)
  - Inputs: environment available=true; candidate=4673; lower=4673; upper=4710.
  - Seeds: no draw.
  - Draws: none.
  - Trace: 4673 is nonzero and is not below 4673 or above 4710.
  - Outputs: filtered candidate=4673.
  - Discriminates: exclusive lower bound would return zero.
  - runtime_observed: no
- **MUS-V02** Above upper bound (MUS-R02b; boundary: supplied-input stage)
  - Inputs: environment available=true; candidate=4711; lower=4673; upper=4710.
  - Seeds: no draw.
  - Draws: none.
  - Trace: 4711 exceeds 4710.
  - Outputs: filtered candidate=0.
  - Discriminates: accepting all nonzero candidates would retain 4711.
  - runtime_observed: no
- **MUS-V03** Missing environment (MUS-R02b; boundary: supplied-input stage)
  - Inputs: environment available=false; supplied candidate=4679; lower=4673; upper=4710.
  - Seeds: no draw.
  - Draws: none.
  - Trace: availability false selects zero before the range tests.
  - Outputs: filtered candidate=0.
  - Discriminates: retaining the unavailable row's supplied value would return 4679.
  - runtime_observed: no
- **MUS-V04** Wait below delay (MUS-R03; boundary: supplied-input stage)
  - Inputs: filtered candidate=4679; current request=4673; current tick=1074; area-entry tick=1000; bookmark tick=900; eligibility A=1,B=1.
  - Seeds: no draw.
  - Draws: none.
  - Trace: elapsed=74; current request nonzero; both eligible; change is disallowed.
  - Outputs: current request=4673; bookmark tick=900.
  - Discriminates: immediate area-track switching would write 4679.
  - runtime_observed: no
- **MUS-V05** Delay equality (MUS-R03; boundary: supplied-input stage)
  - Inputs: filtered candidate=4679; current request=4673; current tick=1075; area-entry tick=1000; bookmark tick=900; eligibility A=1,B=1.
  - Seeds: no draw.
  - Draws: none.
  - Trace: elapsed=75 permits change; candidates differ.
  - Outputs: current request=4679; bookmark tick=1075.
  - Discriminates: a strict greater-than delay would leave request 4673.
  - runtime_observed: no
- **MUS-V06** Zero master volume permits stored request (MUS-R03; boundary: supplied-input stage)
  - Inputs: filtered candidate=4679; current request=4673; current tick=1001; area-entry tick=1000; bookmark tick=900; eligibility A=0,B=1.
  - Seeds: no draw.
  - Draws: none.
  - Trace: elapsed=1 but eligibility A zero permits change.
  - Outputs: current request=4679; bookmark tick=1001.
  - Discriminates: freezing the stored request when master volume is zero would retain 4673.
  - runtime_observed: no
- **MUS-V07** Same request does not refresh time (MUS-R03; boundary: supplied-input stage)
  - Inputs: filtered candidate=4673; current request=4673; current tick=1200; area-entry tick=1000; bookmark tick=900; eligibility A=1,B=1.
  - Seeds: no draw.
  - Draws: none.
  - Trace: elapsed=200 permits change, but candidate equals current request.
  - Outputs: current request=4673; bookmark tick=900.
  - Discriminates: unconditional time refresh would write 1200.
  - runtime_observed: no
- **MUS-V08** Bookmark cadence equality (MUS-R04; boundary: supplied-input stage)
  - Inputs: bookmark eligible=true; current tick=225; bookmark tick=100; stored bookmark=7; position=150; blocks=100,200,300.
  - Seeds: no draw.
  - Draws: none.
  - Trace: current tick equals 100+125; strict greater comparison fails.
  - Outputs: stored bookmark=7; bookmark tick=100.
  - Discriminates: updating at equality would write 200 and tick 225.
  - runtime_observed: no
- **MUS-V09** Bookmark selects next block (MUS-R04; boundary: supplied-input stage)
  - Inputs: bookmark eligible=true; current tick=226; bookmark tick=100; stored bookmark=7; position=150; blocks=100,200,300.
  - Seeds: no draw.
  - Draws: none.
  - Trace: 226>225; first comparison fails; second selects 200; third fails.
  - Outputs: stored bookmark=200; bookmark tick=226.
  - Discriminates: rounding position down to a prior block would write 100.
  - runtime_observed: no
- **MUS-V10** Beyond all blocks resets bookmark (MUS-R04; boundary: supplied-input stage)
  - Inputs: bookmark eligible=true; current tick=226; bookmark tick=100; stored bookmark=200; position=300; blocks=100,200,300.
  - Seeds: no draw.
  - Draws: none.
  - Trace: initialized zero; all upper bounds exclude equality or position exceeds them.
  - Outputs: stored bookmark=0; bookmark tick=226.
  - Discriminates: clamping to the last block would write 300.
  - runtime_observed: no
- **MUS-V11** Sequential tests with unordered blocks (MUS-R04; boundary: supplied-input stage)
  - Inputs: bookmark eligible=true; current tick=226; bookmark tick=100; stored bookmark=7; position=150; blocks=200,100,300.
  - Seeds: no draw.
  - Draws: none.
  - Trace: first test selects 200; second fails; third selects 300, replacing 200.
  - Outputs: stored bookmark=300; bookmark tick=226.
  - Discriminates: first-match branching would retain 200.
  - runtime_observed: no

- **MUS-V12** Range from positive songs (MUS-R02a; boundary: supplied-input stage)
  - Inputs: prepared sound count=5000; ordered environment Song values=[0,4679,4673,-1,4684,4679].
  - Seeds: no draw.
  - Draws: none.
  - Trace: initialize bounds 5000/0; ignore 0; select 4679/4679; then 4673/4679; ignore -1; then 4673/4684; duplicate 4679 changes neither bound.
  - Outputs: lower bound=4673; upper bound=4684.
  - Discriminates: including zero/negative entries would lower the bound below 4673; selecting first/last song would not yield both 4673 and 4684.
  - runtime_observed: no
- **MUS-V13** No implicit repair for absent positive songs (MUS-R02a; boundary: supplied-input stage)
  - Inputs: prepared sound count=5000; ordered environment Song values=[0,-1,0].
  - Seeds: no draw.
  - Draws: none.
  - Trace: every value is nonpositive; retain the initial bounds.
  - Outputs: lower bound=5000; upper bound=0.
  - Discriminates: inventing an empty-set range 0/0 would change lower bound to 0.
  - runtime_observed: no

- **MUS-V14** Fresh ordinary submission ignores request failure (MUS-R05a; boundary: supplied ordinary stage)
  - Inputs: current request 4673, master/music settings 100/50, stored bookmark 0, current tick 200, submission tick 100; first-voice query returns 0, music-range query returns false, playback request returns 0. Supplied operations have no other controller-state or seed effects.
  - Seeds: no draw.
  - Draws: none.
  - Trace: no requested-song voice; no music-range voice and zero bookmark select flags 2; request stop then playback; ignore failure and store tick.
  - Outputs: ordered log [firstVoice(4673)→0, musicRangePresent→false, stopMusic, playback(4673,absent,0,2,0)→0]; submission tick 200; current request 4673 and bookmark 0 unchanged.
  - Discriminates: flags 0 for a fresh start, playback before stop, or storing the tick only after a successful request changes an assertion.
  - runtime_observed: no.
- **MUS-V15** Nonzero bookmark selects zero flags (MUS-R05a; boundary: supplied ordinary stage)
  - Inputs: current request 4673, master/music settings 100/50, stored bookmark 700, current tick 201, submission tick 100; first-voice query returns 0, music-range query returns false, playback request returns 9. Supplied operations have no other controller-state or seed effects.
  - Seeds: no draw.
  - Draws: none.
  - Trace: nonzero bookmark prevents flags 2 despite no music-range voice.
  - Outputs: log [firstVoice(4673)→0, musicRangePresent→false, stopMusic, playback(4673,absent,0,0,700)→9]; submission tick 201; current request 4673 and bookmark 700 unchanged.
  - Discriminates: using only the music-range query for flags gives 2; clearing the bookmark loses 700.
  - runtime_observed: no.
- **MUS-V16** Existing music-range voice selects zero flags (MUS-R05a; boundary: supplied ordinary stage)
  - Inputs: current request 4673, master/music settings 100/50, stored bookmark 0, current tick 202, submission tick 100; first-voice query returns 0, music-range query returns true, playback request returns 9. Supplied operations have no other controller-state or seed effects.
  - Seeds: no draw.
  - Draws: none.
  - Trace: present music-range voice prevents flags 2 despite zero bookmark.
  - Outputs: log [firstVoice(4673)→0, musicRangePresent→true, stopMusic, playback(4673,absent,0,0,0)→9]; submission tick 202; current request 4673 and bookmark 0 unchanged.
  - Discriminates: using only the bookmark for flags gives 2.
  - runtime_observed: no.
- **MUS-V17** Existing requested-song handle suppresses resubmission (MUS-R05a; boundary: supplied ordinary stage)
  - Inputs: current request 4673, master/music settings 100/50, stored bookmark 700, current tick 203, submission tick 100; first-voice query returns 9. Later operations would return normally but are not called.
  - Seeds: no draw.
  - Draws: none.
  - Trace: nonzero handle returns from this stage before range query, stop or playback.
  - Outputs: log [firstVoice(4673)→9]; submission tick 100, current request 4673 and bookmark 700 unchanged.
  - Discriminates: resubmitting every update or refreshing the attempt tick changes the log or tick.
  - runtime_observed: no.
- **MUS-V18** Disabled ordinary submission (MUS-R05a; boundary: supplied ordinary stage)
  - Inputs: current request 4673, master/music settings 100/0, stored bookmark 700, current tick 204, submission tick 100. Supplied operations would return normally but are not called.
  - Seeds: no draw.
  - Draws: none.
  - Trace: music setting zero skips the whole stage.
  - Outputs: empty log; submission tick 100, current request 4673 and bookmark 700 unchanged.
  - (a) Fresh inputs: current request 0, settings 100/50, bookmark 700, current tick 204, submission tick 100. Outputs: empty log; submission tick 100, request 0 and bookmark 700 unchanged; no draw.
  - (b) Fresh inputs: current request 4673, settings 0/50, bookmark 700, current tick 204, submission tick 100. Outputs: empty log; submission tick 100, request 4673 and bookmark 700 unchanged; no draw.
  - Discriminates: ignoring either zero setting or submitting song zero changes a case's empty log.
  - runtime_observed: no.

### 8.2 Observation vectors

- **MUS-O01** Cross from town to wilderness and compare area-entry tick, filtered request and the stored request write at the 75-tick boundary. [Procedure](../research/spec-evidence/sound-music-selection-and-transitions.md#observation-procedures). Expected arithmetic is MUS-R02b and MUS-R03 for the measured inputs; no elapsed-seconds claim is made.
- **MUS-O02** Observe an active song across a block boundary, recording backend position, block values, bookmark and bookmark tick before and after a qualifying update. Expected supplied-input calculation is MUS-R04. The same procedure must separately establish backend position units before claiming a seek or audible match.

## 9. Acceptance and dependencies

Partial arithmetic acceptance requires MUS-V01 through MUS-V18 and the MUS-R01 table record. This verifies the named text cells and supplied-input range, request, bookmark and ordinary-submission stages. It does not establish the complete in-game music behavior. Full reproduction additionally requires MUS-Q01 and MUS-Q02 closure and attributable observations for the actual integrated assertions; MUS-Q03 currently prevents runtime-confirmed status.

Depends on: [numeric foundations](core-numeric-and-rng-foundations.md), unsigned 32-bit wrapping and signed comparisons. GDA-R03a and GDA-R03b in [game-data table records](core-game-data-table-records.md) supply descriptor identity and direct numeric conversion only; the full compiler remains GDA-R02/GDA-R03d and is not assumed verified. MUS-R05a consumes supplied request boundaries whose actual effects are gated by STP-R03b/STP-R04b; no audible effect is imported from them. [Sound table and playback rules](sound-table-and-playback-rules.md) owns the voice lifecycle needed to close MUS-Q01; its current gated lifecycle is a pending dependency, not verified support.

Depended on by: no established contract consumer.

## 10. Related specs

- [Sound table and playback rules](sound-table-and-playback-rules.md): shared voice and fade behavior.
- [3D position and falloff](sound-3d-position-and-falloff.md): emitter placement belongs to world sound playback.
- [Ambient and weather](sound-ambient-and-weather.md): selects environment layers before the music update.

## Appendix A. Provenance

Original LoD Expansion 1.13c Windows x86 evidence; D2Client, D2Common and D2Sound identities are in the [source manifest](../research/source-manifest.md). Source inspection on 2026-10-29 used byte-matched reconstructions where available, otherwise disassembly corroborated against the original image and an independent Ghidra listing. The complete effective sound-environment text file was read through the running game's archive loader. No music gameplay rule is runtime-confirmed. The [companion](../research/spec-evidence/sound-music-selection-and-transitions.md) records the exact verified stages, original function locations, data identity and unresolved integration.
