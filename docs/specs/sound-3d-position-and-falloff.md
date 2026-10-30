# Positional sound and falloff

**Description**

Positional sound places an audible source relative to the local player. The game selects distance limits, computes a listener-relative position and changes source amplitude before requesting playback properties. The verified stages below describe those calculations; complete tracking, special placement and device integration retain explicit gates.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [numeric foundations](core-numeric-and-rng-foundations.md)
Ids: S3D-R, S3D-Q, S3D-V, S3D-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/sound-3d-position-and-falloff.md)
Open evidence gaps: [gap file](../plans/gaps/sound-3d-position-and-falloff.md)

## 1. Scope

Client positional source updates, distance attenuation, listener-relative coordinates, source tracking and the amplitude submitted to playback. Sound selection, voice admission and voice retirement belong to the playback contract. No geometric occlusion or propagation requirement is established. Verified supplied-input stages do not establish the complete audible result.

## 2. Trigger and ownership

D2Client's enabled audio update invokes active-source coordination after ambience and music coordination. Active-source coordination reaches relative-position preparation, distance caching and property updates; voice creation also invokes property updating. D2Sound converts requested amplitude into a backend volume request. The local player is the listener used by the examined relative-position path. Complete eligibility, source-reference lifecycle and nondefault backend routing remain S3D-Q01 through S3D-Q03. No server-frame cadence or network message is established.

## 3. Inputs and data

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| listener and emitter presentation position | Local player and source unit, or supplied accessor results | Two i32 projected-position components each | Inputs to ordinary listener-relative placement; these are not raw world subtile coordinates |
| sound identity | Selected sound record | i32 | Identity 2599 takes a separate gated transform |
| source position | Supplied relative source location | Three finite IEEE binary 32-bit floats | Raw position retained by the cache |
| distance selector | Loaded Sounds.Falloff, or supplied value | u32 | Selects inner/outer distance bounds |
| base amplitude | Supplied result after option/duck scaling | i32 | Input to local distance calculation |
| record volume | Loaded sound record or supplied Volume | u8 | Multiplies the distance-stage amplitude |
| backend amplitude scalar | Device amplitude scalar, supplied | i32, 0..255 for verified backend stage | Multiplies a requested source amplitude |
| requested amplitude | Supplied client amplitude | i32, 0..255 for verified backend stage | Stored and converted to backend volume |
| distance factor | Supplied source factor | Finite binary 32-bit float, 0..1 | Amplitude multiplier is 1 minus this value when not handled by backend |
| backend handles factor | Supplied device/source condition | Boolean | Selects whether local factor multiplication is bypassed |
| source validity | Supplied live source with valid underlying buffer | Boolean precondition | Backend scalar stage requires a valid source |
| ordinary floating division selected | Supplied arithmetic execution condition | Boolean, true for S3D-R05/R06 | Alternate CPU workaround is outside the verified numeric stage |
| floating execution environment | Supplied arithmetic execution condition for S3D-R05/R06 | Round-to-nearest, extended significand precision 64 bits | Exact intermediate arithmetic profile; actual initialization remains S3D-Q02 |

| Mapping / constant | Origin | Binding |
|---|---|---|
| Radius map, coordinate clamp 2000, ordinary height 640, amplitude divisor 255 | Engine code | Fixed by original code |
| Distance selector | Table data | Loaded from Falloff; descriptor and one live compiled cell corroborated |
| cursor_pass.Falloff = 1 | Table data | Named text value; same loaded field corroborated for that identity |
| Special coordinate sound identity 2599 | Engine code | Fixed original identity; special transform remains gated |

## 4. Rules (execution order)

**S3D-R01** Named data value [verified in 1.13c data]

The effective text table contains Falloff = 1 for Sound = cursor_pass, explicit Index = 1. [T01](../research/spec-evidence/sound-3d-position-and-falloff.md#t01-effective-text-cell) supports the named cell. This alone does not establish audible playback.

**S3D-R02** Distance bounds [verified in 1.13c code (static)]

For a valid supplied sound record, interpret its Falloff selector as u32 and obtain these exact binary 32-bit values. No random draw occurs.

| Selector | Outer bound | Inner bound |
|---:|---:|---:|
| 0 | 400 | 60 |
| 1 | 700 | 60 |
| 2 | 1000 | 200 |
| 3 | 1500 | 400 |
| 4 | 2000 | 2000 |
| any other u32 | 700 | 60 |

A negative signed interpretation does not change the unsigned default branch. Invalid record access is outside the valid-record stage; no safe fallback record is specified.

**S3D-R03** Ordinary listener-relative placement [verified in 1.13c code (static)]

For sound identity other than 2599, with supplied projected integer positions, compute i32 differences emitterX-listenerX and emitterY-listenerY with low-32-bit wrapping. The verified supplied-input boundary requires both differences to be exactly representable in binary 32-bit floats. Store relativeX as the first difference, relativeY as twice the second difference in binary 32-bit floating arithmetic, and relativeZ as 640. The second difference is doubled in floating arithmetic, not wrapped as an integer multiplication. No random draw occurs. Upstream production of projected positions, rounding for other differences and the special-identity transform remain S3D-Q01.

**S3D-R04** Position cache and bounded distance key [runtime-confirmed (1.13c)]

For finite supplied source position x/y/z with integer-valued x/y and a valid cache, first copy all three raw binary 32-bit components unchanged into retained position. Separately clamp x and y individually to [-2000,2000], square the clamped values and add the squares as distanceKey. These products and the sum are exact integers representable in binary 32-bit storage. Do not include z. Do not replace the retained raw position with the clamped values. No random draw occurs. Matching observations cover the ordinary and positive-y clamp cases named under S3D-V03 and S3D-V04; they do not establish every integer input branch. Fractional x/y precision behavior remains S3D-Q02.

**S3D-R05** Supplied distance-amplitude stage [verified in 1.13c code (static)]

Boundary: finite relative x/y, supplied inner/outer bounds from S3D-R02 and base amplitude after preceding option/duck scaling. Compute d = sqrt(x*x+y*y) in extended floating arithmetic, using raw x/y, not the clamped distanceKey. If d > inner, compute distanceAmplitude by truncating ((outer-d)*baseAmplitude)/(outer-inner) toward zero to a signed integer. Otherwise distanceAmplitude is baseAmplitude unchanged. The verified numeric boundary requires a nonzero denominator whenever the division executes and a representable signed result. Do not clamp a negative distanceAmplitude to zero.

Then compute the low signed 32-bit product of recordVolume and distanceAmplitude, divide by 255 with signed truncation toward zero, and expose that result as volumeWeightedAmplitude. No random draw occurs. This intermediate value does not establish the final request on every playback mode: stereo-panning gain and device routing remain S3D-R07. Equal bounds with d <= inner avoid division; equal bounds with d > inner have no accepted normal numeric result here.

**S3D-R06** Backend amplitude request conversion [verified in 1.13c code (static)]

Boundary: valid source, requestedAmplitude and backendAmplitudeScalar each in 0..255, finite distanceFactor in 0..1. Store requestedAmplitude as retained requested amplitude. Compute v = trunc_toward_zero(backendAmplitudeScalar*requestedAmplitude/255). If backendHandlesFactor is false, replace v with trunc_toward_zero((1-distanceFactor)*v); otherwise retain v.

For these integer v values, submit backend volume -10000 if v=0, 0 if v=255, and trunc_toward_zero(2000*log₁₀(v/255)) otherwise. Numerical logarithm and division follow the ordinary floating path; alternate CPU workaround and exceptional floating inputs remain S3D-Q02. This rule establishes the requested backend scalar, not device success or audible amplitude. No random draw occurs.

**S3D-R07** Complete tracking, routing and update eligibility [unresolved]

Complete source-reference tracking, special-identity coordinate placement, source admission/culling, option/duck scaling, nondefault panning and backend failure behavior remain S3D-Q01 through S3D-Q03. The accepted scalar and cache stages do not establish full positional parity.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| S3D-R01 | Named Falloff text cell | verified in 1.13c data | none | T01 |
| S3D-R02 | Supplied radius map | verified in 1.13c code (static) | none locally; invalid records S3D-Q01 | S3D-V01 |
| S3D-R03 | Ordinary projected relative placement | verified in 1.13c code (static) | none locally; producer and special identity S3D-Q01 | S3D-V02 |
| S3D-R04 | Raw position copy and bounded key | runtime-confirmed (1.13c) | none for finite supplied inputs | S3D-V03, S3D-V04 |
| S3D-R05 | Supplied distance arithmetic | verified in 1.13c code (static) | none for stated numeric boundary; final routing S3D-Q02 | S3D-V05, S3D-V06 |
| S3D-R06 | Ordinary backend scalar request | verified in 1.13c code (static) | none for supplied ordinary path; device outcome S3D-Q02 | S3D-V07, S3D-V08 |
| S3D-R07 | Complete positional integration | unresolved | S3D-Q01, S3D-Q02, S3D-Q03: complete pipeline | none |

## 5. Outputs and side effects

| Output | Rule |
|---|---|
| Outer/inner bounds | S3D-R02 |
| Prepared relative source position | S3D-R03 |
| Retained raw position and distanceKey | S3D-R04 |
| Intermediate distanceAmplitude and volumeWeightedAmplitude | S3D-R05 |
| Retained requested amplitude and submitted backend volume | S3D-R06 |
| Complete backend position, tracking lifecycle and audible result | S3D-R07; gated |

S3D-R01 describes data and performs no gameplay write.

## 6. Branches and edge cases

Unsigned selector defaults include all bit patterns above 4. The cache clamps distanceKey inputs while preserving raw coordinates. Ordinary placement wraps the differences before floating conversion. Equal radius bounds and unrepresentable floating-to-integer results have the precise limited boundary in S3D-R05. Non-finite coordinates, invalid sources, missing records and nondefault hardware/CPU paths do not acquire guessed behavior. No difficulty or PvP exception is established.

## 7. Open questions

- **S3D-Q01** [unresolved] Complete projected-position production and the special sound-2599 transform, source-reference creation/removal and invalid-record behavior. Local player listener identity and the audio-update root are established; supplied ordinary placement, radius and cache stages remain available.
- **S3D-Q02** [unresolved] Complete full option/duck scaling, nondefault stereo panning, backend position/factor routing, actual floating-environment initialization, fractional cache arithmetic, CPU floating-workaround equivalence and actual device-result semantics. The supplied S3D-R05/R06 stages remain available under their specified arithmetic environment.
- **S3D-Q03** [unresolved] Complete Tracking eligibility, nearest-reference selection/ties, reference disappearance and update suppression/culling across all voice states. Located consumers and attempted observations are in the companion.
- **S3D-Q04** [unresolved] Obtain matching complete observations for S3D-R02, S3D-R03, S3D-R05 and S3D-R06. Backend conversion captures are partial because root/source matching and final method invocation were not captured together; they are not a full vector promotion.

## 8. Test vectors

### 8.1 Executable vectors

Cases are independent supplied-input stages, with no other persistent operations and an empty random-draw log. Ordinary floating division is selected in every numeric case; S3D-R05/R06 cases supply round-to-nearest and 64-bit extended significand precision. Backend request operations log their integer arguments but do not simulate device success.

- **S3D-V01** Selector map (S3D-R02; boundary: supplied-input stage)
  - Inputs: valid supplied record, Falloff selector 2.
  - Seeds: no draw.
  - Draws: none.
  - Trace: selector 2 chooses outer 1000, inner 200.
  - Outputs: outer 1000, inner 200.
  - Discriminates: treating Falloff as a literal radius returns 2 rather than 1000/200.
  - runtime_observed: no.
  - (a) Inputs: valid record, selector 0. Draws: none. Outputs: outer 400, inner 60.
  - (b) Inputs: valid record, selector 1. Draws: none. Outputs: outer 700, inner 60.
  - (c) Inputs: valid record, selector 3. Draws: none. Outputs: outer 1500, inner 400.
  - (d) Inputs: valid record, selector 4. Draws: none. Outputs: outer 2000, inner 2000.
  - (e) Inputs: valid record, selector 4294967295. Draws: none. Outputs: outer 700, inner 60.

- **S3D-V02** Ordinary player-relative placement (S3D-R03; boundary: supplied-input stage)
  - Inputs: sound identity 1; listener projected position (100,200), emitter projected position (90,230).
  - Seeds: no draw.
  - Draws: none.
  - Trace: differences -10 and 30; floating double of 30 is 60.
  - Outputs: relative position (-10,60,640).
  - Discriminates: swapping emitter/listener or failing to double y changes exact signs/components.
  - runtime_observed: no.

- **S3D-V03** Cache ignores height for distance (S3D-R04; boundary: supplied-input stage)
  - Inputs: source position (320,80,640); retained position (320,80,640), distanceKey 108800.
  - Seeds: no draw.
  - Draws: none.
  - Trace: neither x nor y clamps; 320*320+80*80=108800; raw components copied unchanged.
  - Outputs: retained position (320,80,640), distanceKey 108800.
  - Discriminates: including z produces 518400 instead of 108800.
  - runtime_observed: yes (1.13c, 2026-10-29, [C02](../research/spec-evidence/sound-3d-position-and-falloff.md#c02-source-property-captures)).
  - (a) Inputs: source position (320,80,640), retained position (0,0,0), distanceKey 0. Draws: none. Outputs: retained position (320,80,640), distanceKey 108800. runtime_observed: no.

- **S3D-V04** Cache clamps key but not position (S3D-R04; boundary: supplied-input stage)
  - Inputs: source position (681,2726,640); retained position (681,2726,640), distanceKey 4463761.
  - Seeds: no draw.
  - Draws: none.
  - Trace: y clamps to 2000 for the key only; 681*681+2000*2000=4463761.
  - Outputs: retained position (681,2726,640), distanceKey 4463761.
  - Discriminates: using unclamped y changes the key; storing clamped y incorrectly retains 2000.
  - runtime_observed: yes (1.13c, 2026-10-29, [C02](../research/spec-evidence/sound-3d-position-and-falloff.md#c02-source-property-captures)).
  - (a) Inputs: source position (-3000,4000,9), retained position (0,0,0), distanceKey 0. Draws: none. Outputs: retained position (-3000,4000,9), distanceKey 8000000. runtime_observed: no.

- **S3D-V05** Supplied amplitude falls between radii (S3D-R05; boundary: supplied-input stage)
  - Inputs: relative position (500,0,640); outer 1000, inner 200, baseAmplitude 200, recordVolume 128.
  - Seeds: no draw.
  - Draws: none.
  - Trace: d=500; distanceAmplitude=trunc((1000-500)*200/800)=125; volumeWeightedAmplitude=trunc(128*125/255)=62.
  - Outputs: distanceAmplitude 125, volumeWeightedAmplitude 62.
  - Discriminates: omitting inner radius or rounding final division to nearest changes 125/62.
  - runtime_observed: no.

- **S3D-V06** No invented outer clamp (S3D-R05; boundary: supplied-input stage)
  - Inputs: relative position (1100,0,0); outer 1000, inner 200, baseAmplitude 200, recordVolume 255.
  - Seeds: no draw.
  - Draws: none.
  - Trace: d=1100; trunc((1000-1100)*200/800)=-25; trunc(255*-25/255)=-25.
  - Outputs: distanceAmplitude -25, volumeWeightedAmplitude -25.
  - Discriminates: clamping the isolated stage to zero changes both outputs; full caller eligibility remains gated.
  - runtime_observed: no.
  - (a) Inputs: relative position (0,0,0), outer 2000, inner 2000, baseAmplitude 200, recordVolume 255. Draws: none. Outputs: distanceAmplitude 200, volumeWeightedAmplitude 200.

- **S3D-V07** Backend scalar logarithm (S3D-R06; boundary: supplied-input stage)
  - Inputs: valid source; requestedAmplitude 99, backendAmplitudeScalar 255, distanceFactor 0, backendHandlesFactor false; retained requested amplitude 255.
  - Seeds: no draw.
  - Draws: none.
  - Trace: v=99; second factor retains 99; 2000*log₁₀(99/255) is approximately -821.8099716728104; signed truncation gives -821.
  - Outputs: retained requested amplitude 99; one submitted backend volume request -821.
  - Discriminates: nearest rounding gives -822; submitting linear amplitude gives 99.
  - runtime_observed: no.

- **S3D-V08** Backend factor and endpoints (S3D-R06; boundary: supplied-input stage)
  - Inputs: valid source; requestedAmplitude 255, backendAmplitudeScalar 255, distanceFactor 0.5, backendHandlesFactor false; retained requested amplitude 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: initial v=255; factor gives trunc(127.5)=127; trunc(2000*log₁₀(127/255))=-605.
  - Outputs: retained requested amplitude 255; one submitted backend volume request -605.
  - Discriminates: rounding factor to 128 changes the logarithmic request.
  - runtime_observed: no.
  - (a) Inputs: same valid source and amplitudes/factor, backendHandlesFactor true, retained requested amplitude 0. Draws: none. Outputs: retained requested amplitude 255; one backend volume request 0.
  - (b) Inputs: valid source, requestedAmplitude 255, backendAmplitudeScalar 0, distanceFactor 0, backendHandlesFactor false; retained requested amplitude 0. Draws: none. Outputs: retained requested amplitude 255; one backend volume request -10000.

### 8.2 Observation vectors

- **S3D-O01** Measure complete relative source state, selected radii, intermediate client amplitude and outgoing backend properties at stationary and moving positions. The [procedure and C02 captures](../research/spec-evidence/sound-3d-position-and-falloff.md#observation-procedure) distinguish established cache observations from partial conversion and producer captures.

## 9. Acceptance and dependencies

Partial acceptance covers S3D-R01/T01 and S3D-R02 through S3D-R06 with S3D-V01 through S3D-V08, only at their stated supplied-input boundaries. S3D-R04 has matching named cases; other verified stages retain observation gaps. Full reproduction additionally requires S3D-R07 and its integration gates. Neither scalar requests nor unchanged cache captures establish audible output or complete moving-emitter parity.

**Depends on:** The arithmetic notation in section 1 of [numeric foundations](core-numeric-and-rng-foundations.md); the local conversion rules here are supported by the client/audio helpers mapped in the companion. All inputs crossing unresolved producer boundaries are supplied explicitly.

**Depended on by:** no established consumer.

## 10. Related specs

[Sound table and playback rules](sound-table-and-playback-rules.md) owns voice admission and lifecycle. Its unresolved STP-R04b device boundary is not established by the positional scalar calculations.

## Appendix A. Provenance

Windows x86 Expansion 1.13c. [Source manifest](../research/source-manifest.md) identifies D2Client, D2Common and D2Sound. On 2026-10-29 the investigation read available matched reconstructions before disassembly, corroborated decoding and constants against original images, checked loaded selector correspondence, and captured source-property stages on an owned vanilla game. The [companion](../research/spec-evidence/sound-3d-position-and-falloff.md) preserves earlier inconclusive attempts and the exact scope of the matching cache cases.
